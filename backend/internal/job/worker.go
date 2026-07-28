package job

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"wealthos-backend/internal/config"
	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/integration/sepay"
	"wealthos-backend/internal/metrics"
	"wealthos-backend/internal/service"
	"wealthos-backend/internal/storage"
)

type Worker struct {
	cfg     *config.Config
	store   storage.Store
	service *service.WealthService
}

func NewWorker(cfg *config.Config, store storage.Store, service *service.WealthService) *Worker {
	return &Worker{cfg: cfg, store: store, service: service}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.processIncomingBankFeedEvents()
				w.processForecastScenarios()
				w.reconcile()
			case <-ctx.Done():
				log.Printf("stopping background jobs, env=%s", w.cfg.Env)
				return
			}
		}
	}()

	// Webhooks are the normal path. A slower reconciliation pass repairs any
	// missed delivery without turning the UI into a polling client.
	reconcileTicker := time.NewTicker(15 * time.Minute)
	go func() {
		defer reconcileTicker.Stop()
		w.reconcileBankHub(ctx)
		for {
			select {
			case <-reconcileTicker.C:
				w.reconcileBankHub(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (w *Worker) processForecastScenarios() {
	if w.service == nil {
		return
	}
	for _, scenario := range w.store.ListForecastScenariosByStatus("running") {
		if scenario.ID == "" || scenario.Status != "running" {
			continue
		}
		result := `{"status":"done","simulatedAt":"` + time.Now().Format(time.RFC3339) + `","assumptions":` + quoteJSONString(scenario.Assumptions) + `}`
		if _, err := w.store.FinalizeForecastScenario(scenario.ID, "done", result); err != nil {
			log.Printf("forecast scenario finalization failed: id=%s err=%v", scenario.ID, err)
		}
	}
}

func quoteJSONString(value string) string {
	return strconv.Quote(value)
}

func (w *Worker) reconcile() {
	// Kept for existing reconciliation extensions. Bank Hub recovery runs on
	// its own lower-frequency schedule in reconcileBankHub.
}

func (w *Worker) reconcileBankHub(ctx context.Context) {
	if w.cfg == nil || w.service == nil || w.store == nil {
		return
	}
	client := sepay.NewBankHubClient(w.cfg.SePayBankHubBaseURL, w.cfg.SePayBankHubClientID, w.cfg.SePayBankHubSecret)
	if !client.Configured() {
		return
	}
	today := time.Now().UTC()
	for _, conn := range w.store.ListAllBankConnections() {
		if conn.Provider != "sepay" || conn.Status != "connected" || conn.ExternalID == "" {
			continue
		}
		// A Bank Hub company is owned by the Finora user profile. Do not use
		// the deployment default for an existing connection: that would allow
		// reconciliation to cross user/company boundaries after per-user
		// companies are enabled.
		providerAccount, found := w.store.GetSePayBankAccountByXID(conn.ExternalID)
		if !found || providerAccount.UserID == "" {
			log.Printf("Bank Hub reconciliation skipped: no user-owned account for connection=%s", conn.ID)
			continue
		}
		profile, found := w.store.GetSePayUserProfile(providerAccount.UserID)
		if !found || strings.TrimSpace(profile.CompanyXID) == "" {
			log.Printf("Bank Hub reconciliation skipped: no company profile for connection=%s", conn.ID)
			continue
		}
		start := conn.LastSyncedAt
		if start.IsZero() {
			start = today.AddDate(0, 0, -7)
		} else {
			// Small overlap is intentional: provider IDs make re-import safe and
			// absorb clock skew or a transaction posted around the prior cursor.
			start = start.Add(-5 * time.Minute)
		}
		transactions, err := client.ListTransactions(ctx, profile.CompanyXID, conn.ExternalID, start.Format("2006-01-02"), today.Format("2006-01-02"))
		if err != nil {
			metrics.Inc("sepay_reconciliation_failures_total")
			log.Printf("Bank Hub reconciliation failed: connection=%s err=%v", conn.ID, err)
			continue
		}
		for _, transaction := range transactions {
			direction := map[string]string{"credit": "in", "debit": "out"}[strings.ToLower(strings.TrimSpace(transaction.TransferType))]
			amount := formatBankHubAmount(transaction.Amount)
			if direction == "" || amount == "" || transaction.TransactionID == "" {
				continue
			}
			_, err := w.service.EnqueueSePayIncoming(service.SePayWebhookEvent{
				ConnectionID: string(conn.ID), ProviderAccountID: conn.ExternalID, Direction: direction, Amount: amount, Currency: "VND",
				Description: transaction.Content, Reference: transaction.ReferenceNumber,
				ExternalID: transaction.TransactionID, OccurredAt: transaction.TransactionDate,
			})
			if err != nil {
				log.Printf("Bank Hub reconciliation enqueue failed: connection=%s transaction=%s err=%v", conn.ID, transaction.TransactionID, err)
			}
		}
	}
}

func formatBankHubAmount(value any) string {
	switch amount := value.(type) {
	case float64:
		return strconv.FormatFloat(amount, 'f', -1, 64)
	case string:
		return amount
	case json.Number:
		return amount.String()
	default:
		return ""
	}
}

func (w *Worker) processIncomingBankFeedEvents() {
	if w.service == nil {
		return
	}
	for _, event := range w.store.ListBankFeedEvents(domain.ID(""), domain.BankFeedEventStateQueued) {
		if event.State == "" || event.State != domain.BankFeedEventStateQueued {
			continue
		}
		if age := time.Since(event.CreatedAt); age > 0 {
			metrics.Add("sepay_queue_lag_seconds_total", int64(age.Seconds()))
		}
		if err := w.service.ProcessBankFeedEvent(event.ID); err != nil {
			log.Printf("bank feed event process failed: id=%s err=%v", event.ID, err)
		}
	}
}

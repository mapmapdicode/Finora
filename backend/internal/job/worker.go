package job

import (
	"context"
	"log"
	"strconv"
	"time"

	"wealthos-backend/internal/config"
	"wealthos-backend/internal/domain"
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
				w.processQueuedBankFeed()
				w.processForecastScenarios()
				w.reconcile()
			case <-ctx.Done():
				log.Printf("stopping background jobs, env=%s", w.cfg.Env)
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
	log.Printf("wallet/job: reconciliation tick")
}

func (w *Worker) processQueuedBankFeed() {
	if w.service == nil {
		return
	}
	for _, feed := range w.store.ListBankFeedByState(domain.ID(""), domain.PostingStateAutoReady) {
		if feed.ID == "" || feed.PostingState != domain.PostingStateAutoReady {
			continue
		}
		if _, err := w.service.ProcessQueuedBankFeed(feed.ID); err != nil {
			log.Printf("bank feed process failed: id=%s err=%v", feed.ID, err)
		}
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
		if err := w.service.ProcessBankFeedEvent(event.ID); err != nil {
			log.Printf("bank feed event process failed: id=%s err=%v", event.ID, err)
		}
	}
}

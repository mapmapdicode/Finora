package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthos-backend/internal/db"
	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/storage"
)

// Run with POSTGRES_INTEGRATION_URL pointing to an isolated disposable
// database. The test proves the durable ingress invariant at the database
// boundary rather than only against the in-memory store.
func TestPostgresSePayIngressIsIdempotentAndClaimedOnce(t *testing.T) {
	dsn := os.Getenv("POSTGRES_INTEGRATION_URL")
	if dsn == "" {
		t.Skip("set POSTGRES_INTEGRATION_URL to run Postgres integration tests")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer pool.Close()
	migrationDir := filepath.Clean(filepath.Join("..", "db", "migrations"))
	if err := db.NewMigrationRunner(pool, migrationDir).Run(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	store := storage.NewPostgresStore(pool)
	userID := store.SeedDemoUser("sepay-it-"+uuid.NewString()+"@example.test", "SePay IT", "pass")
	if userID == "" {
		t.Fatal("seed user")
	}
	user, err := store.EnsureUserPortfolio("SePay integration "+uuid.NewString(), "VND", userID)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	portfolio, ok := store.FirstPortfolio(user.ID)
	if !ok {
		t.Fatal("portfolio")
	}
	account, err := store.CreateAccount(domain.Account{UserID: user.ID, PortfolioID: portfolio.ID, Name: "Bank", Type: "bank", Currency: "VND"})
	if err != nil {
		t.Fatalf("account: %v", err)
	}
	connection, err := store.CreateBankConnection(domain.BankConnection{UserID: user.ID, Provider: "sepay", ExternalID: "provider-account-" + uuid.NewString()})
	if err != nil {
		t.Fatalf("connection: %v", err)
	}

	const retries = 100
	transactionID := "transaction-" + uuid.NewString()
	eventKey := "sepay::provider-account::" + transactionID
	var enqueueFailures atomic.Int32
	var wg sync.WaitGroup
	for range retries {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := store.EnqueueBankFeedEvent(domain.BankFeedEvent{UserID: user.ID, ConnectionID: connection.ID, Provider: "sepay", EventKey: eventKey, ExternalID: transactionID, Payload: `{"transaction_id":"` + transactionID + `"}`}); err != nil {
				enqueueFailures.Add(1)
			}
		}()
	}
	wg.Wait()
	if enqueueFailures.Load() != 0 {
		t.Fatalf("enqueue failures: %d", enqueueFailures.Load())
	}
	events := store.ListBankFeedEvents(user.ID, "")
	if len(events) != 1 {
		t.Fatalf("expected one durable event, got %d", len(events))
	}

	var claims atomic.Int32
	for range retries {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := store.ClaimBankFeedEvent(events[0].ID); ok {
				claims.Add(1)
			}
		}()
	}
	wg.Wait()
	if claims.Load() != 1 {
		t.Fatalf("expected one worker claim, got %d", claims.Load())
	}

	for range retries {
		_, err = store.IngestBankFeed(domain.BankFeedTransaction{UserID: user.ID, ConnectionID: connection.ID, AccountID: account.ID, Amount: "125000", Currency: "VND", Direction: "out", Description: "retry", OccurredAt: time.Now().UTC(), ExternalID: transactionID, PostingState: domain.PostingStateReview, ClassificationStatus: "needs_review"})
		if err != nil {
			t.Fatalf("ingest: %v", err)
		}
	}
	if feeds := store.ListBankFeed(user.ID); len(feeds) != 1 {
		t.Fatalf("expected one source feed, got %d", len(feeds))
	}
}

package job

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"wealthos-backend/internal/config"
	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/service"
	"wealthos-backend/internal/storage"
)

func TestReconcileBankHubUsesAccountOwnersCompany(t *testing.T) {
	var requestedCompany string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/token":
			_, _ = w.Write([]byte(`{"access_token":"test-token"}`))
		case "/v1/transaction":
			requestedCompany = r.URL.Query().Get("company_xid")
			_, _ = w.Write([]byte(`{"data":[],"meta":{"total_pages":1}}`))
		default:
			t.Fatalf("unexpected request %s", r.URL.Path)
		}
	}))
	defer server.Close()

	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("reconcile-owner@example.test", "Owner", "hash")
	ws, err := store.EnsureUserPortfolio("Owner", "VND", userID)
	if err != nil {
		t.Fatal(err)
	}
	account, err := store.CreateAccount(domain.Account{UserID: ws.ID, Name: "Bank", Type: "bank", Currency: "VND"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.UpsertSePayUserProfile(domain.SePayUserProfile{UserID: userID, CompanyXID: "user-company", Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	providerAccount, err := store.UpsertSePayBankAccount(domain.SePayBankAccount{UserID: userID, BankAccountXID: "provider-account", Status: "linked", SupportsIn: true, SupportsOut: true})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.UpsertBankAccountMapping(domain.BankAccountMapping{SePayBankAccountID: providerAccount.ID, UserID: userID, AccountID: account.ID, Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.CreateBankConnection(domain.BankConnection{UserID: ws.ID, Provider: "sepay", ExternalID: "provider-account", Status: "connected"}); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{SePayBankHubBaseURL: server.URL, SePayBankHubClientID: "id", SePayBankHubSecret: "secret", SePayBankHubCompanyID: "deployment-default"}
	NewWorker(cfg, store, service.NewWealthService(store, nil)).reconcileBankHub(context.Background())
	if requestedCompany != "user-company" {
		t.Fatalf("reconciliation company = %q, want user-owned company", requestedCompany)
	}
}

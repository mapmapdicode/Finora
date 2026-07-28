package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"wealthos-backend/internal/config"
	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/integration/sepay"
	"wealthos-backend/internal/service"
	"wealthos-backend/internal/storage"
)

type fakeBankHub struct {
	accounts []sepay.BankHubAccount
	company  string
	query    string
}

func TestParseStandardSePayWebhookPayload(t *testing.T) {
	event, err := parseSePayWebhookPayload([]byte(`{"id":92704,"gateway":"Vietcombank","transactionDate":"2026-07-27 11:08:33","accountNumber":"1017588888","content":"chuyen tien","transferType":"in","description":"NGUYEN VAN A","transferAmount":5000000,"referenceCode":"FT24012345678"}`))
	if err != nil {
		t.Fatalf("parse standard payload: %v", err)
	}
	if event.AccountID != "1017588888" || event.Direction != "in" || event.Amount != "5000000" || event.ExternalID != "92704" || event.Reference != "FT24012345678" || event.OccurredAt != "2026-07-27 11:08:33" {
		t.Fatalf("unexpected standard event: %+v", event)
	}
}

func TestVerifySePayWebhookAcceptsDocumentedSHA256Prefix(t *testing.T) {
	body := []byte(`{"id":1}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	h := &WealthHandler{secret: "test-hmac-secret"}
	req := httptest.NewRequest(http.MethodPost, "/hooks/sepay-webhook", strings.NewReader(string(body)))
	req.Header.Set("X-SePay-Timestamp", timestamp)
	req.Header.Set("X-SePay-Signature", "sha256="+hex.EncodeToString(signSePayPayload(h.secret, timestamp, body)))
	if err := h.verifySePayWebhook(body, req); err != nil {
		t.Fatalf("verify documented signature: %v", err)
	}
}

func TestBotPublicAPIRequiresAccountSecretAndScopesHistory(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("bot-api@example.test", "Bot API", "hash")
	ws, err := store.EnsureUserPortfolio("Bot API", "VND", userID)
	if err != nil {
		t.Fatal(err)
	}
	portfolio, err := store.CreatePortfolio(domain.Portfolio{UserID: ws.ID, Name: "Main", BaseCurrency: "VND"})
	if err != nil {
		t.Fatal(err)
	}
	account, err := store.CreateAccount(domain.Account{UserID: ws.ID, PortfolioID: portfolio.ID, Name: "Bot cash", Type: "cash", Currency: "VND"})
	if err != nil {
		t.Fatal(err)
	}
	secret := "finora_bot_test_secret"
	digest := sha256.Sum256([]byte(secret))
	if _, err := store.UpsertBotAccountKey(domain.BotAccountKey{AccountID: account.ID, SecretHash: hex.EncodeToString(digest[:]), Prefix: "finora_bot_test"}); err != nil {
		t.Fatal(err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/public/v1/accounts/"+string(account.ID)+"/transactions", strings.NewReader(`{"type":"expense","amount":"123000","name":"Ăn trưa","occurredAt":"2026-07-20"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Finora-Account-Key", secret)
	c.Params = gin.Params{{Key: "id", Value: string(account.ID)}}
	h.BotCreateTransaction(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", w.Code, w.Body.String())
	}

	historyW := httptest.NewRecorder()
	historyC, _ := gin.CreateTestContext(historyW)
	historyC.Request = httptest.NewRequest(http.MethodGet, "/public/v1/accounts/"+string(account.ID)+"/transactions/history?from=2026-07-01&to=2026-07-31", nil)
	historyC.Request.Header.Set("X-Finora-Account-Key", secret)
	historyC.Params = gin.Params{{Key: "id", Value: string(account.ID)}}
	h.BotListTransactions(historyC)
	if historyW.Code != http.StatusOK || !strings.Contains(historyW.Body.String(), "Ăn trưa") {
		t.Fatalf("history status=%d body=%s", historyW.Code, historyW.Body.String())
	}

	deniedW := httptest.NewRecorder()
	deniedC, _ := gin.CreateTestContext(deniedW)
	deniedC.Request = httptest.NewRequest(http.MethodGet, "/public/v1/accounts/"+string(account.ID)+"/transactions/history?from=2026-07-01&to=2026-07-31", nil)
	deniedC.Request.Header.Set("X-Finora-Account-Key", "wrong")
	deniedC.Params = gin.Params{{Key: "id", Value: string(account.ID)}}
	h.BotListTransactions(deniedC)
	if deniedW.Code != http.StatusUnauthorized {
		t.Fatalf("wrong secret status=%d", deniedW.Code)
	}
}

func (f *fakeBankHub) Configured() bool { return true }
func (f *fakeBankHub) CreateLink(context.Context, string, string) (sepay.LinkSession, error) {
	return sepay.LinkSession{}, nil
}
func (f *fakeBankHub) ListBankAccounts(_ context.Context, company, query string) ([]sepay.BankHubAccount, error) {
	f.company, f.query = company, query
	return f.accounts, nil
}

func TestRequireEditorRoleBlocksViewer(t *testing.T) {
	h := &WealthHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("user_id", "ws-1")
	c.Set("user_role", "viewer")

	if h.requireEditorRole(c) {
		t.Fatalf("expected viewer role to be rejected")
	}
	if w.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Result().StatusCode)
	}
}

func TestRequireEditorRoleAllowsEditorAndOwner(t *testing.T) {
	t.Run("editor", func(t *testing.T) {
		h := &WealthHandler{}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("user_id", "ws-1")
		c.Set("user_role", "editor")
		if !h.requireEditorRole(c) {
			t.Fatalf("expected editor role to be allowed")
		}
	})

	t.Run("owner", func(t *testing.T) {
		h := &WealthHandler{}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("user_id", "ws-1")
		c.Set("user_role", "owner")
		if !h.requireEditorRole(c) {
			t.Fatalf("expected owner role to be allowed")
		}
	})
}

func TestSyncMySePayBankAccountsFetchesProviderDataAndScopesItToCurrentUser(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := domain.ID("sepay-sync-user")
	if _, err := store.UpsertSePayUserProfile(domain.SePayUserProfile{UserID: userID, CompanyXID: "company-user-1", Status: "link_pending"}); err != nil {
		t.Fatal(err)
	}
	fake := &fakeBankHub{accounts: []sepay.BankHubAccount{
		{XID: "provider-account-1", BankCode: "MBB", BrandName: "MBBank", AccountNumber: "001 234 567", BankAPIConnected: true, Active: true},
		{XID: "provider-account-other", BankCode: "MBB", BrandName: "MBBank", AccountNumber: "999999", BankAPIConnected: true, Active: true},
	}}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
	h.bankHub, h.pilotBanks = fake, map[string]struct{}{"MBB": {}}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/me/sepay/bank-accounts/sync", strings.NewReader(`{"accountNumber":"001234567"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", string(userID))
	h.SyncMySePayBankAccounts(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if fake.company != "company-user-1" || fake.query != "001234567" {
		t.Fatalf("provider lookup = %q / %q", fake.company, fake.query)
	}
	accounts := store.ListSePayBankAccounts(userID)
	if len(accounts) != 1 || accounts[0].BankAccountXID != "provider-account-1" || !accounts[0].SupportsIn || !accounts[0].SupportsOut {
		t.Fatalf("unexpected persisted accounts: %+v", accounts)
	}
}

func TestRequireEditorRoleRejectsMissingRole(t *testing.T) {
	h := &WealthHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("user_id", "ws-1")

	if h.requireEditorRole(c) {
		t.Fatalf("expected missing role to be rejected")
	}
	if w.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Result().StatusCode)
	}
}

func TestRequireOwnerRoleOnlyAllowsOwner(t *testing.T) {
	t.Run("owner", func(t *testing.T) {
		h := &WealthHandler{}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("user_id", "ws-1")
		c.Set("user_role", "owner")
		if !h.requireOwnerRole(c) {
			t.Fatalf("expected owner role to be allowed")
		}
	})

	t.Run("editor", func(t *testing.T) {
		h := &WealthHandler{}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("user_id", "ws-1")
		c.Set("user_role", "editor")
		if h.requireOwnerRole(c) {
			t.Fatalf("expected editor role to be rejected")
		}
		if w.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", w.Result().StatusCode)
		}
	})

	t.Run("viewer", func(t *testing.T) {
		h := &WealthHandler{}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("user_id", "ws-1")
		c.Set("user_role", "viewer")
		if h.requireOwnerRole(c) {
			t.Fatalf("expected viewer role to be rejected")
		}
		if w.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", w.Result().StatusCode)
		}
	})
}

func TestListPortfolioSnapshotsRejectsInvalidLimit(t *testing.T) {
	store := storage.NewInMemoryStore()
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", uid)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	p, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		t.Fatalf("missing portfolio")
	}

	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
	r := gin.New()
	r.GET("/portfolios/:id/snapshots", h.ListPortfolioSnapshots)

	req := httptest.NewRequest(http.MethodGet, "/portfolios/"+string(p.ID)+"/snapshots?limit=bad", nil)
	req.Header.Set("x-user-id", string(ws.ID))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if got, want := resp.Result().StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func TestListPortfolioSnapshotsSupportsPagination(t *testing.T) {
	store := storage.NewInMemoryStore()
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", uid)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	p, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		t.Fatalf("missing portfolio")
	}
	acc, err := store.CreateAccount(domain.Account{
		UserID:      ws.ID,
		PortfolioID: p.ID,
		Name:        "Main",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	svc := service.NewWealthService(store, nil)
	for i := 0; i < 4; i++ {
		_, err := store.CreateTransaction(domain.Transaction{
			UserID:     ws.ID,
			AccountID:  acc.ID,
			Type:       domain.TransactionTypeIncome,
			Amount:     "10.00",
			Currency:   "VND",
			OccurredAt: time.Now().UTC(),
			Status:     domain.TransactionStatusPosted,
		})
		if err != nil {
			t.Fatalf("create tx %d: %v", i, err)
		}
		if _, err := svc.GetPortfolioNetWorth(string(p.ID)); err != nil {
			t.Fatalf("compute net worth %d: %v", i, err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	h := NewWealthHandler(store, svc, nil)
	r := gin.New()
	r.GET("/portfolios/:id/snapshots", h.ListPortfolioSnapshots)

	req1 := httptest.NewRequest(http.MethodGet, "/portfolios/"+string(p.ID)+"/snapshots?limit=3", nil)
	req1.Header.Set("x-user-id", string(ws.ID))
	resp1 := httptest.NewRecorder()
	r.ServeHTTP(resp1, req1)

	if resp1.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 first page, got %d", resp1.Result().StatusCode)
	}
	var page1 struct {
		Items      []service.NetWorthResult `json:"items"`
		NextCursor string                   `json:"nextCursor"`
	}
	if err := json.NewDecoder(resp1.Result().Body).Decode(&page1); err != nil {
		t.Fatalf("decode first page: %v", err)
	}
	if len(page1.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(page1.Items))
	}
	if page1.NextCursor == "" {
		t.Fatalf("expected next cursor")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/portfolios/"+string(p.ID)+"/snapshots?limit=3&cursor="+url.QueryEscape(page1.NextCursor), nil)
	req2.Header.Set("x-user-id", string(ws.ID))
	resp2 := httptest.NewRecorder()
	r.ServeHTTP(resp2, req2)

	if resp2.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 second page, got %d", resp2.Result().StatusCode)
	}
	var page2 struct {
		Items      []service.NetWorthResult `json:"items"`
		NextCursor string                   `json:"nextCursor"`
	}
	if err := json.NewDecoder(resp2.Result().Body).Decode(&page2); err != nil {
		t.Fatalf("decode second page: %v", err)
	}
	if len(page2.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(page2.Items))
	}
	if page2.NextCursor != "" {
		t.Fatalf("expected no next cursor on final page")
	}
}

func TestListTransactionsSupportsPaginationAndFilters(t *testing.T) {
	store := storage.NewInMemoryStore()
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", uid)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	p, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		t.Fatalf("missing portfolio")
	}
	acc, err := store.CreateAccount(domain.Account{
		UserID:      ws.ID,
		PortfolioID: p.ID,
		Name:        "Main",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	now := time.Now().UTC()
	_, err = store.CreateTransaction(domain.Transaction{
		UserID:     ws.ID,
		AccountID:  acc.ID,
		CategoryID: "salary",
		Type:       domain.TransactionTypeExpense,
		Amount:     "10.00",
		Currency:   "VND",
		OccurredAt: now.Add(-3 * time.Hour),
		Status:     domain.TransactionStatusPosted,
		Note:       "coffee",
	})
	if err != nil {
		t.Fatalf("create tx1: %v", err)
	}
	_, err = store.CreateTransaction(domain.Transaction{
		UserID:     ws.ID,
		AccountID:  acc.ID,
		CategoryID: "bonus",
		Type:       domain.TransactionTypeIncome,
		Amount:     "20.00",
		Currency:   "VND",
		OccurredAt: now.Add(-2 * time.Hour),
		Status:     domain.TransactionStatusPosted,
		Note:       "monthly salary",
	})
	if err != nil {
		t.Fatalf("create tx2: %v", err)
	}
	_, err = store.CreateTransaction(domain.Transaction{
		UserID:     ws.ID,
		AccountID:  acc.ID,
		CategoryID: "salary",
		Type:       domain.TransactionTypeExpense,
		Amount:     "30.00",
		Currency:   "VND",
		OccurredAt: now.Add(-1 * time.Hour),
		Status:     domain.TransactionStatusPending,
		Note:       "bonus",
	})
	if err != nil {
		t.Fatalf("create tx3: %v", err)
	}

	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
	r := gin.New()
	r.GET("/transactions", h.ListTransactions)

	type pageResp struct {
		Items      []domain.Transaction `json:"items"`
		NextCursor string               `json:"nextCursor"`
	}

	req1 := httptest.NewRequest(http.MethodGet, "/transactions?limit=2&accountId="+string(acc.ID), nil)
	req1.Header.Set("x-user-id", string(ws.ID))
	resp1 := httptest.NewRecorder()
	r.ServeHTTP(resp1, req1)
	if got, want := resp1.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
	var p1 pageResp
	if err := json.NewDecoder(resp1.Result().Body).Decode(&p1); err != nil {
		t.Fatalf("decode page1: %v", err)
	}
	if len(p1.Items) != 2 {
		t.Fatalf("expected 2 items on first page, got %d", len(p1.Items))
	}
	if p1.NextCursor == "" {
		t.Fatalf("expected next cursor")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/transactions?limit=2&accountId="+string(acc.ID)+"&cursor="+url.QueryEscape(p1.NextCursor), nil)
	req2.Header.Set("x-user-id", string(ws.ID))
	resp2 := httptest.NewRecorder()
	r.ServeHTTP(resp2, req2)
	if got, want := resp2.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
	var p2 pageResp
	if err := json.NewDecoder(resp2.Result().Body).Decode(&p2); err != nil {
		t.Fatalf("decode page2: %v", err)
	}
	if len(p2.Items) != 1 {
		t.Fatalf("expected 1 item on final page, got %d", len(p2.Items))
	}
	if p2.NextCursor != "" {
		t.Fatalf("expected no next cursor on final page")
	}

	req3 := httptest.NewRequest(http.MethodGet, "/transactions?type=income&accountId="+string(acc.ID)+"&limit=10", nil)
	req3.Header.Set("x-user-id", string(ws.ID))
	resp3 := httptest.NewRecorder()
	r.ServeHTTP(resp3, req3)
	var p3 pageResp
	if got, want := resp3.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
	if err := json.NewDecoder(resp3.Result().Body).Decode(&p3); err != nil {
		t.Fatalf("decode page3: %v", err)
	}
	if len(p3.Items) != 1 {
		t.Fatalf("expected only income transaction, got %d", len(p3.Items))
	}
	if p3.Items[0].Type != domain.TransactionTypeIncome {
		t.Fatalf("expected income, got %s", p3.Items[0].Type)
	}

	req4 := httptest.NewRequest(http.MethodGet, "/transactions?search=salary&accountId="+string(acc.ID)+"&limit=10", nil)
	req4.Header.Set("x-user-id", string(ws.ID))
	resp4 := httptest.NewRecorder()
	r.ServeHTTP(resp4, req4)
	var p4 pageResp
	if got, want := resp4.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
	if err := json.NewDecoder(resp4.Result().Body).Decode(&p4); err != nil {
		t.Fatalf("decode page4: %v", err)
	}
	if len(p4.Items) != 3 {
		t.Fatalf("expected 3 transactions with salary in note or category, got %d", len(p4.Items))
	}
}

func TestListTransactionsRejectsInvalidCursor(t *testing.T) {
	store := storage.NewInMemoryStore()
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", uid)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
	r := gin.New()
	r.GET("/transactions", h.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/transactions?cursor=bad-cursor", nil)
	req.Header.Set("x-user-id", string(ws.ID))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Result().StatusCode)
	}
}

func TestGetPortfolioNetWorthSupportsAsOfQuery(t *testing.T) {
	store := storage.NewInMemoryStore()
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", uid)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	p, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		t.Fatalf("missing default portfolio")
	}
	acc, err := store.CreateAccount(domain.Account{
		UserID:      ws.ID,
		PortfolioID: p.ID,
		Name:        "Main",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	prop, err := store.CreateProperty(domain.Property{
		UserID:      ws.ID,
		PortfolioID: p.ID,
		Name:        "Villa",
		Address:     "HCM",
	})
	if err != nil {
		t.Fatalf("create property: %v", err)
	}
	base := time.Date(2026, time.July, 15, 8, 0, 0, 0, time.UTC)
	_, err = store.AddPropertyValuation(domain.PropertyValuation{
		PropertyID:  prop.ID,
		Amount:      "1000.00",
		Currency:    "VND",
		Source:      "manual",
		EffectiveAt: base,
	})
	if err != nil {
		t.Fatalf("add property valuation early: %v", err)
	}
	_, err = store.AddPropertyValuation(domain.PropertyValuation{
		PropertyID:  prop.ID,
		Amount:      "2000.00",
		Currency:    "VND",
		Source:      "manual",
		EffectiveAt: base.AddDate(0, 0, 4),
	})
	if err != nil {
		t.Fatalf("add property valuation late: %v", err)
	}

	_, err = store.CreateTransaction(domain.Transaction{
		UserID:      ws.ID,
		AccountID:   acc.ID,
		PortfolioID: p.ID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "500.00",
		Currency:    "VND",
		OccurredAt:  base.AddDate(0, 0, 2),
		Status:      domain.TransactionStatusPosted,
	})
	if err != nil {
		t.Fatalf("create first income tx: %v", err)
	}
	_, err = store.CreateTransaction(domain.Transaction{
		UserID:      ws.ID,
		AccountID:   acc.ID,
		PortfolioID: p.ID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "250.00",
		Currency:    "VND",
		OccurredAt:  base.AddDate(0, 0, 6),
		Status:      domain.TransactionStatusPosted,
	})
	if err != nil {
		t.Fatalf("create second income tx: %v", err)
	}

	svc := service.NewWealthService(store, nil)
	h := NewWealthHandler(store, svc, nil)
	r := gin.New()
	r.GET("/portfolios/:id/net-worth", h.GetPortfolioNetWorth)

	req := httptest.NewRequest(http.MethodGet, "/portfolios/"+string(p.ID)+"/net-worth?asOf="+base.AddDate(0, 0, 3).Format(time.RFC3339), nil)
	req.Header.Set("x-user-id", string(ws.ID))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if got, want := resp.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
	var out service.NetWorthResult
	if err := json.NewDecoder(resp.Result().Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.NetWorth != "1500.00" {
		t.Fatalf("expected as-of net worth 1500.00, got %s", out.NetWorth)
	}

	reqInvalid := httptest.NewRequest(http.MethodGet, "/portfolios/"+string(p.ID)+"/net-worth?asOf=not-a-date", nil)
	reqInvalid.Header.Set("x-user-id", string(ws.ID))
	respInvalid := httptest.NewRecorder()
	r.ServeHTTP(respInvalid, reqInvalid)
	if got, want := respInvalid.Result().StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("expected bad request for invalid asOf, got %d", got)
	}
}

func TestParseAsOfSupportsFlexibleInputFormats(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "rfc3339 seconds",
			input:    "2026-07-22T10:00:00Z",
			expected: "2026-07-22 10:00:00 +0000 UTC",
		},
		{
			name:     "rfc3339 without seconds",
			input:    "2026-07-22T10:00Z",
			expected: "2026-07-22 10:00:00 +0000 UTC",
		},
		{
			name:     "datetime with local separator",
			input:    "2026-07-22 10:00:00",
			expected: "2026-07-22 10:00:00 +0000 UTC",
		},
		{
			name:     "date only",
			input:    "2026-07-22",
			expected: "2026-07-22 00:00:00 +0000 UTC",
		},
		{
			name:     "rfc3339 with timezone offset",
			input:    "2026-07-22T14:00:00+07:00",
			expected: "2026-07-22 07:00:00 +0000 UTC",
		},
		{
			name:     "unix timestamp",
			input:    "1690080000",
			expected: "2023-07-23 02:40:00 +0000 UTC",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAsOf(tc.input)
			if err != nil {
				t.Fatalf("parseAsOf %q error: %v", tc.input, err)
			}
			if got.Format("2006-01-02 15:04:05 -0700 MST") != tc.expected {
				t.Fatalf("for %q expected %s, got %s", tc.input, tc.expected, got.Format("2006-01-02 15:04:05 -0700 MST"))
			}
		})
	}
}

func TestParseDateFilterRejectsInvalidInput(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		got := parseDateFilter("2026-07-22 10:00")
		if got.IsZero() {
			t.Fatal("expected valid date to parse")
		}
	})
	t.Run("invalid", func(t *testing.T) {
		got := parseDateFilter("invalid-date")
		if !got.IsZero() {
			t.Fatalf("expected zero for invalid date, got %s", got.Format(time.RFC3339))
		}
	})
}

func TestListAssistantCommandsFiltersByUser(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws1, err := store.EnsureUserPortfolio("", "VND", userID)
	if err != nil {
		t.Fatalf("create user 1: %v", err)
	}
	otherID := store.SeedDemoUser("other@x.com", "Another", "pass")
	ws2, err := store.EnsureUserPortfolio("", "VND", otherID)
	if err != nil {
		t.Fatalf("create user 2: %v", err)
	}

	_, err = store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:  ws1.ID,
		Command: "cmd ws1",
	})
	if err != nil {
		t.Fatalf("create ws1 command: %v", err)
	}
	_, err = store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:  ws2.ID,
		Command: "cmd ws2",
	})
	if err != nil {
		t.Fatalf("create ws2 command: %v", err)
	}

	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
	r := gin.New()
	r.GET("/assistant/commands", h.ListAssistantCommands)

	req := httptest.NewRequest(http.MethodGet, "/assistant/commands", nil)
	req.Header.Set("x-user-id", string(ws1.ID))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Result().StatusCode)
	}
	var commands []domain.AssistantCommand
	if err := json.NewDecoder(resp.Result().Body).Decode(&commands); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("expected only commands for selected user, got %d", len(commands))
	}
	if string(commands[0].UserID) != string(ws1.ID) {
		t.Fatalf("expected user %s, got %s", ws1.ID, commands[0].UserID)
	}
}

func TestTelegramWebhookRejectsInvalidSecret(t *testing.T) {
	store := storage.NewInMemoryStore()
	h := NewWealthHandler(store, service.NewWealthService(store, nil), &config.Config{
		TelegramWebhookSecret: "top-secret",
	})
	r := gin.New()
	r.POST("/assistant/telegram/webhook", h.TelegramWebhook)

	body := `{"message":{"text":"open chrome"}}`
	req := httptest.NewRequest(http.MethodPost, "/assistant/telegram/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Result().StatusCode)
	}
}

func TestTelegramWebhookCreatesAssistantCommandForLinkedUser(t *testing.T) {
	store := storage.NewInMemoryStore()
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", uid)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	h := NewWealthHandler(store, service.NewWealthService(store, nil), &config.Config{
		TelegramWebhookSecret: "top-secret",
	})
	r := gin.New()
	r.POST("/assistant/telegram/webhook", h.TelegramWebhook)

	body := `{"message":{"text":"tÃ£o mÃ´i lÃ©nh"},"userId":"` + string(ws.ID) + `","update_id":1}`
	req := httptest.NewRequest(http.MethodPost, "/assistant/telegram/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "top-secret")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Result().StatusCode)
	}

	var out struct {
		Status    string `json:"status"`
		CommandID string `json:"commandId"`
		UserID    string `json:"userId"`
	}
	if err := json.NewDecoder(resp.Result().Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.Status != "received" {
		t.Fatalf("expected received status, got %q", out.Status)
	}
	if out.UserID != string(ws.ID) {
		t.Fatalf("expected user %s, got %s", ws.ID, out.UserID)
	}
	if out.CommandID == "" {
		t.Fatalf("expected command id")
	}

	commands := store.ListAssistantCommands(ws.ID)
	if len(commands) != 1 {
		t.Fatalf("expected command to be persisted, got %d", len(commands))
	}
}

func TestTelegramWebhookRequiresUserLink(t *testing.T) {
	store := storage.NewInMemoryStore()
	h := NewWealthHandler(store, service.NewWealthService(store, nil), &config.Config{
		TelegramWebhookSecret: "top-secret",
	})
	r := gin.New()
	r.POST("/assistant/telegram/webhook", h.TelegramWebhook)

	body := `{"message":{"text":"open chrome"}}`
	req := httptest.NewRequest(http.MethodPost, "/assistant/telegram/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "top-secret")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.Result().StatusCode)
	}
}

func TestCreateAssistantCommandSetsIntentAndStatus(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ownerID := domain.ID(userID)
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	cases := []struct {
		name       string
		command    string
		plan       string
		wantStatus string
		wantPlan   string
	}{
		{name: "read intent", command: "show my balance", plan: "", wantStatus: assistantStatusPlanned, wantPlan: assistantIntentRead},
		{name: "write intent", command: "add new transaction", plan: "", wantStatus: assistantStatusAwaitingApproval, wantPlan: assistantIntentWrite},
		{name: "external intent", command: "open chrome", plan: "", wantStatus: assistantStatusAwaitingApproval, wantPlan: assistantIntentExternalAction},
		{name: "draft intent", command: "prepare a draft scenario", plan: "", wantStatus: assistantStatusPlanned, wantPlan: assistantIntentDraft},
		{name: "blocked intent", command: "delete all data", plan: "", wantStatus: assistantStatusRejected, wantPlan: assistantIntentBlocked},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"command":"` + tc.command + `","plan":"` + tc.plan + `"}`
			req := httptest.NewRequest(http.MethodPost, "/assistant/commands", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(resp)
			c.Request = req
			c.Set("user_id", string(ws.ID))
			c.Set("user_role", "owner")
			c.Set("user_id", string(ownerID))
			h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
			h.CreateAssistantCommand(c)

			if resp.Result().StatusCode != http.StatusCreated {
				t.Fatalf("expected 201, got %d", resp.Result().StatusCode)
			}
			var out domain.AssistantCommand
			if err := json.NewDecoder(resp.Result().Body).Decode(&out); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if out.Status != tc.wantStatus {
				t.Fatalf("expected status %s, got %s", tc.wantStatus, out.Status)
			}
			if out.Plan != tc.wantPlan {
				t.Fatalf("expected plan %s, got %s", tc.wantPlan, out.Plan)
			}
		})
	}
}

func TestApproveCommandTransitions(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ownerID := domain.ID(userID)
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	cmd, err := store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:     ws.ID,
		Command:    "open chrome",
		Status:     assistantStatusAwaitingApproval,
		Plan:       assistantIntentExternalAction,
		ApprovalID: "manual-appr-id",
	})
	if err != nil {
		t.Fatalf("create assistant command: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/assistant/commands/"+string(cmd.ID)+"/approve?approvalId="+string(cmd.ApprovalID), nil)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: string(cmd.ID)}}
	c.Set("user_id", string(ws.ID))
	c.Set("user_role", "owner")
	c.Set("user_id", string(ownerID))
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
	h.ApproveCommand(c)

	var out domain.AssistantCommand
	if err := json.NewDecoder(resp.Result().Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Result().StatusCode)
	}
	if out.Status != assistantStatusDispatched {
		t.Fatalf("expected status %s, got %s", assistantStatusDispatched, out.Status)
	}
	if out.ApprovalID != "" {
		t.Fatalf("expected approval id to be consumed/cleared, got %q", out.ApprovalID)
	}
}

func TestApproveCommandRejectsInvalidTransition(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ownerID := domain.ID(userID)
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	cmd, err := store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:  ws.ID,
		Command: "list recent expenses",
		Status:  assistantStatusPlanned,
		Plan:    assistantIntentRead,
	})
	if err != nil {
		t.Fatalf("create assistant command: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/assistant/commands/"+string(cmd.ID)+"/approve", nil)
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: string(cmd.ID)}}
	c.Set("user_id", string(ws.ID))
	c.Set("user_role", "owner")
	c.Set("user_id", string(ownerID))
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
	h.ApproveCommand(c)

	if resp.Result().StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.Result().StatusCode)
	}
}

func TestCancelCommandTransitionsAndTerminalRules(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ownerID := domain.ID(userID)
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	allow, err := store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:  ws.ID,
		Command: "list recent expenses",
		Status:  assistantStatusPlanned,
		Plan:    assistantIntentRead,
	})
	if err != nil {
		t.Fatalf("create assistant command: %v", err)
	}
	completed, err := store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:  ws.ID,
		Command: "list recent expenses",
		Status:  assistantStatusCompleted,
		Plan:    assistantIntentRead,
	})
	if err != nil {
		t.Fatalf("create assistant command: %v", err)
	}

	t.Run("cancel allowed state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/assistant/commands/"+string(allow.ID)+"/cancel", nil)
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: string(allow.ID)}}
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "owner")
		c.Set("user_id", string(ownerID))
		h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
		h.CancelCommand(c)

		if resp.Result().StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.Result().StatusCode)
		}
		var out domain.AssistantCommand
		if err := json.NewDecoder(resp.Result().Body).Decode(&out); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if out.Status != assistantStatusCancelled {
			t.Fatalf("expected status %s, got %s", assistantStatusCancelled, out.Status)
		}
	})

	t.Run("cancel blocked state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/assistant/commands/"+string(completed.ID)+"/cancel", nil)
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: string(completed.ID)}}
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "owner")
		c.Set("user_id", string(ownerID))
		h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
		h.CancelCommand(c)

		if resp.Result().StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.Result().StatusCode)
		}
	})
}

func TestApproveCommandRejectsReplayToken(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ownerID := domain.ID(userID)
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	cmd, err := store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:  ws.ID,
		Command: "open chrome",
		Status:  assistantStatusAwaitingApproval,
		Plan:    assistantIntentExternalAction,
	})
	if err != nil {
		t.Fatalf("create assistant command: %v", err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)

	approveReq := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/assistant/commands/"+string(cmd.ID)+"/approve?approvalId="+string(cmd.ApprovalID), nil)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: string(cmd.ID)}}
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "owner")
		c.Set("user_id", string(ownerID))
		h.ApproveCommand(c)
		return resp
	}

	resp1 := approveReq()
	if resp1.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected first approval 200, got %d", resp1.Result().StatusCode)
	}

	resp2 := approveReq()
	if resp2.Result().StatusCode != http.StatusConflict {
		t.Fatalf("expected second approval to be rejected, got %d", resp2.Result().StatusCode)
	}
}

func TestTelegramApprovalCallbackSingleUse(t *testing.T) {
	store := storage.NewInMemoryStore()
	ownerID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	cmd, err := store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:  ws.ID,
		Command: "open chrome",
		Status:  assistantStatusAwaitingApproval,
		Plan:    assistantIntentExternalAction,
	})
	if err != nil {
		t.Fatalf("create assistant command: %v", err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), &config.Config{
		TelegramWebhookSecret: "top-secret",
	})
	r := gin.New()
	r.POST("/assistant/telegram/webhook", h.TelegramWebhook)

	payload := `{"callback_query":{"data":"approve:` + string(cmd.ID) + `:` + cmd.ApprovalID + `","from":{"id":9999999},"message":{"chat":{"id":11},"from":{"id":9999999}},"id":"cb-1"},"userId":"` + string(ws.ID) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/assistant/telegram/webhook", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "top-secret")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 first callback approval, got %d", resp.Result().StatusCode)
	}

	resp2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/assistant/telegram/webhook", strings.NewReader(payload))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Telegram-Bot-Api-Secret-Token", "top-secret")
	r.ServeHTTP(resp2, req2)
	if resp2.Result().StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 on replayed callback approval, got %d", resp2.Result().StatusCode)
	}
}

func TestExecutorEventsChecksSecretAndTransitionMapping(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", userID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	cmd, err := store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:  ws.ID,
		Command: "open chrome",
		Status:  assistantStatusPlanned,
		Plan:    assistantIntentExternalAction,
	})
	if err != nil {
		t.Fatalf("create assistant command: %v", err)
	}
	completedCmd, err := store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:  ws.ID,
		Command: "open chrome",
		Status:  assistantStatusCompleted,
		Plan:    assistantIntentExternalAction,
	})
	if err != nil {
		t.Fatalf("create assistant command: %v", err)
	}

	h := NewWealthHandler(store, service.NewWealthService(store, nil), &config.Config{
		HermesExecutorSecret: "hermes-secret",
	})
	cases := []struct {
		name          string
		requestID     string
		commandID     domain.ID
		payloadStatus string
		wantStatus    string
		wantCode      int
	}{
		{name: "invalid secret", requestID: "exec-1", commandID: cmd.ID, payloadStatus: "completed", wantCode: http.StatusUnauthorized},
		{name: "accepted maps to dispatched", requestID: "exec-1", commandID: cmd.ID, payloadStatus: "accepted", wantStatus: assistantStatusDispatched, wantCode: http.StatusOK},
		{name: "running maps from started", requestID: "exec-1", commandID: cmd.ID, payloadStatus: "started", wantStatus: assistantStatusRunning, wantCode: http.StatusOK},
		{name: "terminal idempotency", requestID: "exec-2", commandID: completedCmd.ID, payloadStatus: "completed", wantStatus: assistantStatusCompleted, wantCode: http.StatusOK},
		{name: "invalid event status", requestID: "exec-3", commandID: cmd.ID, payloadStatus: "this-is-not-a-status", wantCode: http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := `{"commandId":"` + string(tc.commandID) + `","status":"` + tc.payloadStatus + `"}`
			req := httptest.NewRequest(http.MethodPost, "/assistant/executors/"+tc.requestID+"/events", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			if tc.name != "invalid secret" {
				req.Header.Set("X-Hermes-Secret", "hermes-secret")
			}
			resp := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(resp)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tc.requestID}}
			h.ExecutorEvents(c)

			if resp.Result().StatusCode != tc.wantCode {
				t.Fatalf("expected %d, got %d", tc.wantCode, resp.Result().StatusCode)
			}
			if tc.wantCode != http.StatusOK {
				return
			}

			var out struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(resp.Result().Body).Decode(&out); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if out.Status != tc.wantStatus {
				t.Fatalf("expected status %s, got %s", tc.wantStatus, out.Status)
			}
		})
	}
}

func TestCreateSePayConnectionReturnsConnectUrlAndPersistedState(t *testing.T) {
	store := storage.NewInMemoryStore()
	ownerID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/sepay/connect", strings.NewReader(`{"provider":"sepay","scope":"read_transactions"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Host = "app.local:8443"
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = req
	c.Set("user_id", string(ws.ID))
	c.Set("user_role", "owner")

	h.CreateSePayConnection(c)

	if resp.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Result().StatusCode)
	}

	var out struct {
		ConnectionID  string `json:"connectionId"`
		Provider      string `json:"provider"`
		Scope         string `json:"scope"`
		ExternalID    string `json:"externalId"`
		CallbackState string `json:"callbackState"`
		ConnectURL    string `json:"connectUrl"`
	}
	if err := json.NewDecoder(resp.Result().Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.ConnectionID == "" {
		t.Fatalf("expected connectionId")
	}
	if out.Provider != sepayDefaultProvider {
		t.Fatalf("expected provider %q, got %q", sepayDefaultProvider, out.Provider)
	}
	if out.Scope != sepayDefaultReadScope {
		t.Fatalf("expected scope %q, got %q", sepayDefaultReadScope, out.Scope)
	}
	if out.ExternalID == "" {
		t.Fatalf("expected externalId")
	}
	if out.CallbackState == "" {
		t.Fatalf("expected callbackState")
	}
	if out.ConnectURL == "" {
		t.Fatalf("expected connectUrl")
	}

	u, err := url.Parse(out.ConnectURL)
	if err != nil {
		t.Fatalf("invalid connectUrl: %v", err)
	}
	if u.Path != sepayCallbackPath {
		t.Fatalf("expected callback path %q, got %q", sepayCallbackPath, u.Path)
	}
	if u.Query().Get("state") != out.CallbackState {
		t.Fatalf("expected callback state in connectUrl")
	}

	conn, ok := store.GetBankConnection(domain.ID(out.ConnectionID))
	if !ok {
		t.Fatalf("expected persisted bank connection")
	}
	if conn.CallbackState != out.CallbackState {
		t.Fatalf("expected persisted callback state, got %q", conn.CallbackState)
	}
	if conn.ExternalID != out.ExternalID {
		t.Fatalf("expected persisted externalId %q, got %q", out.ExternalID, conn.ExternalID)
	}
}

func TestBankHubEventLinksAccountAndIPNQueuesIt(t *testing.T) {
	store := storage.NewInMemoryStore()
	ownerID := store.SeedDemoUser("bankhub@wealthos.vn", "Bank Hub", "pass")
	ws, err := store.EnsureUserPortfolio("Bank Hub", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := store.CreateAccount(domain.Account{UserID: ws.ID, Name: "Bank", Type: "bank", Currency: "VND"}); err != nil {
		t.Fatalf("create account: %v", err)
	}
	conn, err := store.CreateBankConnection(domain.BankConnection{
		UserID:     ws.ID,
		Provider:   "sepay",
		ExternalID: "link-token-xid",
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), &config.Config{SePayBankHubAPIKey: "ipn-secret"})

	eventResp := httptest.NewRecorder()
	eventCtx, _ := gin.CreateTestContext(eventResp)
	eventCtx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/sepay/bankhub/events", strings.NewReader(`{"event":"BANK_ACCOUNT_LINKED","metadata":{"link_token_xid":"link-token-xid","bank_account_xid":"bank-account-xid","brand_name":"MBBank"}}`))
	eventCtx.Request.Header.Set("Authorization", "Apikey ipn-secret")
	h.BankHubEvent(eventCtx)
	if eventResp.Code != http.StatusOK {
		t.Fatalf("expected linked event 200, got %d: %s", eventResp.Code, eventResp.Body.String())
	}
	linked, ok := store.GetBankConnection(conn.ID)
	if !ok || linked.ExternalID != "bank-account-xid" || linked.BankCode != "MBBank" {
		t.Fatalf("expected Bank Hub account mapping, got %+v", linked)
	}

	ipnResp := httptest.NewRecorder()
	ipnCtx, _ := gin.CreateTestContext(ipnResp)
	ipnCtx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/sepay/bankhub/ipn", strings.NewReader(`{"transaction_date":"2026-07-27 10:00:00","bank_account_xid":"bank-account-xid","transfer_type":"debit","amount":125000,"content":"Cafe","reference_code":"FT1","transaction_id":"tx-1"}`))
	ipnCtx.Request.Header.Set("Authorization", "Apikey ipn-secret")
	h.BankHubIPN(ipnCtx)
	if ipnResp.Code != http.StatusOK {
		t.Fatalf("expected IPN 200, got %d: %s", ipnResp.Code, ipnResp.Body.String())
	}
	events := store.ListBankFeedEvents(ws.ID, domain.BankFeedEventStateQueued)
	if len(events) != 1 || events[0].ExternalID != "tx-1" {
		t.Fatalf("expected one queued IPN event, got %+v", events)
	}
}

func TestSePayCallbackStrictStateValidation(t *testing.T) {
	store := storage.NewInMemoryStore()
	ownerID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	conn, err := store.CreateBankConnection(domain.BankConnection{
		UserID:        ws.ID,
		Provider:      sepayDefaultProvider,
		Scope:         sepayDefaultReadScope,
		ExternalID:    "ext",
		CallbackState: "known-state",
		SyncStatus:    "idle",
	})
	if err != nil {
		t.Fatalf("create bank connection: %v", err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)

	t.Run("missing_state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/sepay/callback?connectionId="+string(conn.ID), nil)
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		h.SePayCallback(c)
		if resp.Result().StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 for missing state, got %d", resp.Result().StatusCode)
		}
	})

	t.Run("unknown_state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/sepay/callback?state=unknown", nil)
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		h.SePayCallback(c)
		if resp.Result().StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 for unknown state, got %d", resp.Result().StatusCode)
		}
	})

	t.Run("valid_state_updates_connection", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/sepay/callback?state="+conn.CallbackState+"&connectionId="+string(conn.ID), nil)
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		h.SePayCallback(c)
		if resp.Result().StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for valid state, got %d", resp.Result().StatusCode)
		}
		refreshed, ok := store.GetBankConnection(conn.ID)
		if !ok {
			t.Fatalf("expected bank connection after callback")
		}
		if refreshed.SyncStatus != "callback" {
			t.Fatalf("expected sync status callback, got %q", refreshed.SyncStatus)
		}
		if refreshed.LastSyncedAt.IsZero() {
			t.Fatalf("expected lastSyncedAt to be set")
		}
		if refreshed.CallbackState != conn.CallbackState {
			t.Fatalf("expected callbackState %q, got %q", conn.CallbackState, refreshed.CallbackState)
		}
	})
}

func TestSyncBankConnectionRateLimitsAndCooldown(t *testing.T) {
	store := storage.NewInMemoryStore()
	ownerID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	conn, err := store.CreateBankConnection(domain.BankConnection{
		UserID:     ws.ID,
		Provider:   sepayDefaultProvider,
		Scope:      sepayDefaultReadScope,
		ExternalID: "ext",
		SyncStatus: "idle",
	})
	if err != nil {
		t.Fatalf("create bank connection: %v", err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)

	request := func() int {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/bank-connections/"+string(conn.ID)+"/sync", nil)
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "owner")
		c.Params = gin.Params{{Key: "id", Value: string(conn.ID)}}
		h.SyncBankConnection(c)
		return resp.Result().StatusCode
	}

	if status := request(); status != http.StatusAccepted {
		t.Fatalf("first sync expected 202, got %d", status)
	}
	if status := request(); status != http.StatusTooManyRequests {
		t.Fatalf("second sync expected 429, got %d", status)
	}

	_ = store.UpdateBankConnection(conn.ID, func(item *domain.BankConnection) {
		item.LastSyncRequestedAt = time.Now().UTC().Add(-(sepayMinSyncCooldown + 2*time.Second))
	})

	if status := request(); status != http.StatusAccepted {
		t.Fatalf("sync after cooldown expected 202, got %d", status)
	}

	refreshed, _ := store.GetBankConnection(conn.ID)
	if refreshed == nil {
		t.Fatalf("expected persisted connection")
	}
	if time.Since(refreshed.LastSyncRequestedAt) > 5*time.Second {
		t.Fatalf("expected lastSyncRequestedAt updated after cooldown check")
	}
}

func TestCreateSePayConnectionRequiresOwnerRole(t *testing.T) {
	store := storage.NewInMemoryStore()
	ownerID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)

	t.Run("viewer_rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/sepay/connect", strings.NewReader(`{"provider":"sepay","scope":"read_transactions"}`))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "viewer")
		c.Set("user_id", string(ownerID))
		h.CreateSePayConnection(c)
		if resp.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.Result().StatusCode)
		}
	})

	t.Run("editor_rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/sepay/connect", strings.NewReader(`{"provider":"sepay","scope":"read_transactions"}`))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "editor")
		c.Set("user_id", string(ownerID))
		h.CreateSePayConnection(c)
		if resp.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.Result().StatusCode)
		}
	})

	t.Run("owner_allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/sepay/connect", strings.NewReader(`{"provider":"sepay","scope":"read_transactions"}`))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "owner")
		c.Set("user_id", string(ownerID))
		h.CreateSePayConnection(c)
		if resp.Result().StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.Result().StatusCode)
		}
	})
}

func TestRevokeBankConnectionRequiresOwnerRole(t *testing.T) {
	store := storage.NewInMemoryStore()
	ownerID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	conn, err := store.CreateBankConnection(domain.BankConnection{
		UserID:     ws.ID,
		Provider:   sepayDefaultProvider,
		Scope:      sepayDefaultReadScope,
		ExternalID: "ext",
		SyncStatus: "idle",
	})
	if err != nil {
		t.Fatalf("create bank connection: %v", err)
	}

	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)

	t.Run("viewer_rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/bank-connections/"+string(conn.ID)+"/revoke", nil)
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "viewer")
		c.Set("user_id", string(ownerID))
		c.Params = gin.Params{{Key: "id", Value: string(conn.ID)}}
		h.RevokeBankConnection(c)
		if resp.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.Result().StatusCode)
		}
	})

	t.Run("editor_rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/bank-connections/"+string(conn.ID)+"/revoke", nil)
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "editor")
		c.Set("user_id", string(ownerID))
		c.Params = gin.Params{{Key: "id", Value: string(conn.ID)}}
		h.RevokeBankConnection(c)
		if resp.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.Result().StatusCode)
		}
	})

	t.Run("owner_allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/bank-connections/"+string(conn.ID)+"/revoke", nil)
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "owner")
		c.Set("user_id", string(ownerID))
		c.Params = gin.Params{{Key: "id", Value: string(conn.ID)}}
		h.RevokeBankConnection(c)
		if resp.Result().StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.Result().StatusCode)
		}
	})
}

func TestSyncBankConnectionRejectsViewerRole(t *testing.T) {
	store := storage.NewInMemoryStore()
	ownerID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	conn, err := store.CreateBankConnection(domain.BankConnection{
		UserID:     ws.ID,
		Provider:   sepayDefaultProvider,
		Scope:      sepayDefaultReadScope,
		ExternalID: "ext",
		SyncStatus: "idle",
	})
	if err != nil {
		t.Fatalf("create bank connection: %v", err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)

	t.Run("viewer_rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/bank-connections/"+string(conn.ID)+"/sync", nil)
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "viewer")
		c.Set("user_id", string(ownerID))
		c.Params = gin.Params{{Key: "id", Value: string(conn.ID)}}
		h.SyncBankConnection(c)
		if resp.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.Result().StatusCode)
		}
	})
}

func TestRunForecastScenarioTransitionsToRunning(t *testing.T) {
	store := storage.NewInMemoryStore()
	ownerID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", ownerID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	scenario, err := store.CreateForecastScenario(domain.ForecastScenario{
		UserID:      ws.ID,
		Name:        "Scenario test",
		Assumptions: `{"growthRate":0.1}`,
	})
	if err != nil {
		t.Fatalf("create forecast scenario: %v", err)
	}

	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
	body := strings.NewReader(`{"foo":"bar"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast-scenarios/"+string(scenario.ID)+"/run", body)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = req
	c.Set("user_id", string(ws.ID))
	c.Set("user_role", "owner")
	c.Set("user_id", string(ownerID))
	c.Params = gin.Params{{Key: "id", Value: string(scenario.ID)}}

	h.RunForecastScenario(c)

	if resp.Result().StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.Result().StatusCode)
	}

	var out domain.ForecastScenario
	if err := json.NewDecoder(resp.Result().Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.Status != "running" {
		t.Fatalf("expected status running, got %q", out.Status)
	}
	if out.ID != scenario.ID {
		t.Fatalf("expected scenario %s, got %s", scenario.ID, out.ID)
	}
	if out.Assumptions == "" {
		t.Fatalf("expected assumptions in response")
	}

	var refreshed *domain.ForecastScenario
	for _, item := range store.ListForecastScenarios(ws.ID) {
		if item.ID == scenario.ID {
			refreshed = &item
			break
		}
	}
	if refreshed == nil {
		t.Fatalf("scenario not found after run")
	}
	if refreshed.Status != "running" {
		t.Fatalf("stored status expected running, got %q", refreshed.Status)
	}
}

func TestBankAutomationRulesRequireEditorRole(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", userID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	acc, err := store.CreateAccount(domain.Account{
		UserID:      ws.ID,
		PortfolioID: "",
		Name:        "Main",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	body := `{"accountId":"` + string(acc.ID) + `","name":"Inbound rule","predicate":"in","type":"in","actionType":"income","direction":"in","priority":1,"enabled":true}`
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)

	t.Run("create_blocked_for_viewer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/bank-automation-rules", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "viewer")
		c.Set("user_id", string(userID))
		h.CreateAutomationRule(c)
		if resp.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.Result().StatusCode)
		}
	})

	t.Run("modify_blocked_for_viewer", func(t *testing.T) {
		rule, err := store.CreateAutomationRule(domain.AutomationRule{
			UserID:    ws.ID,
			AccountID: acc.ID,
			Name:      "Existing",
			Priority:  1,
			Enabled:   true,
		})
		if err != nil {
			t.Fatalf("create rule: %v", err)
		}
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/bank-automation-rules/"+string(rule.ID), strings.NewReader(`{"name":"updated"}`))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = req
		c.Set("user_id", string(ws.ID))
		c.Set("user_role", "viewer")
		c.Set("user_id", string(userID))
		c.Params = gin.Params{{Key: "id", Value: string(rule.ID)}}
		h.ModifyAutomationRule(c)
		if resp.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.Result().StatusCode)
		}

		reqDelete := httptest.NewRequest(http.MethodDelete, "/api/v1/bank-automation-rules/"+string(rule.ID), nil)
		reqDelete.Header.Set("Content-Type", "application/json")
		respDelete := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(respDelete)
		c2.Request = reqDelete
		c2.Set("user_id", string(ws.ID))
		c2.Set("user_role", "viewer")
		c2.Set("user_id", string(userID))
		c2.Params = gin.Params{{Key: "id", Value: string(rule.ID)}}
		h.ModifyAutomationRule(c2)
		if respDelete.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", respDelete.Result().StatusCode)
		}
	})
}

func TestModifyAutomationRulePatchOnlyNameKeepsEnabled(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("owner@wealthos.vn", "Owner", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", userID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	acc, err := store.CreateAccount(domain.Account{
		UserID:      ws.ID,
		PortfolioID: "",
		Name:        "Main",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	rule, err := store.CreateAutomationRule(domain.AutomationRule{
		UserID:    ws.ID,
		AccountID: acc.ID,
		Name:      "Initial",
		Priority:  10,
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}

	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bank-automation-rules/"+string(rule.ID), strings.NewReader(`{"name":"Updated name"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = req
	c.Set("user_id", string(ws.ID))
	c.Set("user_role", "owner")
	c.Set("user_id", string(userID))
	c.Params = gin.Params{{Key: "id", Value: string(rule.ID)}}
	h.ModifyAutomationRule(c)

	if resp.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Result().StatusCode)
	}
	got, ok := store.GetAutomationRule(rule.ID)
	if !ok {
		t.Fatal("rule missing after patch")
	}
	if got.Name != "Updated name" {
		t.Fatalf("expected name updated to 'Updated name', got %q", got.Name)
	}
	if !got.Enabled {
		t.Fatal("expected enabled to remain true after partial patch without enabled field")
	}
}

func TestCreateAccountRecordsAuditLog(t *testing.T) {
	store := storage.NewInMemoryStore()
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo", "VND", uid)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	p, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		t.Fatal("missing default portfolio")
	}

	h := NewWealthHandler(store, service.NewWealthService(store, nil), &config.Config{})
	payload := `{"portfolioId":"` + string(p.ID) + `","name":"Cash","type":"cash","currency":"VND"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-test-1")
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = req
	c.Set("user_id", string(ws.ID))
	c.Set("user_role", "owner")
	c.Set("user_id", string(uid))

	h.CreateAccount(c)
	if resp.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Result().StatusCode)
	}

	var created domain.Account
	if err := json.NewDecoder(resp.Result().Body).Decode(&created); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	logs := store.ListAuditLogs(ws.ID)
	if len(logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(logs))
	}
	log := logs[0]
	if log.Action != "create_account" {
		t.Fatalf("expected action create_account, got %q", log.Action)
	}
	if log.TargetType != "account" {
		t.Fatalf("expected target type account, got %q", log.TargetType)
	}
	if log.TargetID != created.ID {
		t.Fatalf("expected target id %s, got %s", created.ID, log.TargetID)
	}
	if log.CorrelationID != "corr-test-1" {
		t.Fatalf("expected correlation id corr-test-1, got %q", log.CorrelationID)
	}
}

func TestListAuditLogsRequiresUser(t *testing.T) {
	store := storage.NewInMemoryStore()
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws1, err := store.EnsureUserPortfolio("User 1", "VND", uid)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	_, err = store.EnsureUserPortfolio("User 2", "VND", uid)
	if err != nil {
		t.Fatalf("create second user: %v", err)
	}
	p, ok := store.FirstPortfolio(ws1.ID)
	if !ok {
		t.Fatal("missing default portfolio")
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), &config.Config{})
	payload := `{"portfolioId":"` + string(p.ID) + `","name":"Cash","type":"cash","currency":"VND"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = req
	c.Set("user_id", string(ws1.ID))
	c.Set("user_role", "owner")
	c.Set("user_id", string(uid))
	h.CreateAccount(c)
	if resp.Result().StatusCode != http.StatusCreated {
		t.Fatalf("seed account: expected 201, got %d", resp.Result().StatusCode)
	}

	reqList := httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs", nil)
	respList := httptest.NewRecorder()
	cList, _ := gin.CreateTestContext(respList)
	cList.Request = reqList
	cList.Set("user_id", string(ws1.ID))
	cList.Set("user_role", "owner")
	cList.Set("user_id", string(uid))
	h.ListAuditLogs(cList)

	if respList.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", respList.Result().StatusCode)
	}
	var logs []domain.AuditLog
	if err := json.NewDecoder(respList.Result().Body).Decode(&logs); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log for user 1, got %d", len(logs))
	}
}

func TestMapMySePayBankAccountEnforcesUserOwnershipAndUserRole(t *testing.T) {
	store := storage.NewInMemoryStore()
	ownerID := store.SeedDemoUser("owner-map@example.test", "Owner", "pass")
	otherID := store.SeedDemoUser("other-map@example.test", "Other", "pass")
	ws, err := store.EnsureUserPortfolio("Map user", "VND", ownerID)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	portfolio, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		t.Fatal("portfolio")
	}
	account, err := store.CreateAccount(domain.Account{UserID: ws.ID, PortfolioID: portfolio.ID, Name: "Bank", Type: "bank", Currency: "VND"})
	if err != nil {
		t.Fatalf("account: %v", err)
	}
	bankAccount, err := store.UpsertSePayBankAccount(domain.SePayBankAccount{UserID: ownerID, BankAccountXID: "provider-map-1", BankCode: "MBB", BankName: "MBBank", AccountNumberMasked: "•••• 1234", SupportsIn: true, SupportsOut: true, Status: "linked"})
	if err != nil {
		t.Fatalf("sepay account: %v", err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), &config.Config{})

	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/me/sepay/bank-accounts/"+string(bankAccount.ID)+"/map", strings.NewReader(`{"accountId":"`+string(account.ID)+`"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: string(bankAccount.ID)}}
	c.Set("user_id", string(ownerID))
	h.MapMySePayBankAccount(c)
	if resp.Code != http.StatusOK {
		t.Fatalf("map expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	mapping, ok := store.GetBankAccountMapping(bankAccount.ID)
	if !ok || mapping.AccountID != account.ID {
		t.Fatalf("mapping missing: %+v", mapping)
	}

	denied := httptest.NewRecorder()
	cDenied, _ := gin.CreateTestContext(denied)
	cDenied.Request = httptest.NewRequest(http.MethodPost, "/api/v1/me/sepay/bank-accounts/"+string(bankAccount.ID)+"/map", strings.NewReader(`{"accountId":"`+string(account.ID)+`"}`))
	cDenied.Request.Header.Set("Content-Type", "application/json")
	cDenied.Params = gin.Params{{Key: "id", Value: string(bankAccount.ID)}}
	cDenied.Set("user_id", string(otherID))
	h.MapMySePayBankAccount(cDenied)
	if denied.Code != http.StatusNotFound {
		t.Fatalf("foreign account must be hidden, got %d", denied.Code)
	}
}

func TestBankHubIPNReplayReturnsSuccessAndQueuesOneEvent(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("ipn-replay@example.test", "IPN", "pass")
	ws, err := store.EnsureUserPortfolio("IPN", "VND", userID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateAccount(domain.Account{UserID: ws.ID, Name: "Bank", Type: "bank", Currency: "VND"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateBankConnection(domain.BankConnection{UserID: ws.ID, Provider: "sepay", ExternalID: "provider-ipn-1"}); err != nil {
		t.Fatal(err)
	}
	h := NewWealthHandler(store, service.NewWealthService(store, nil), &config.Config{SePayBankHubAPIKey: "key"})
	payload := `{"transaction_date":"2026-07-27 10:00:00","bank_account_xid":"provider-ipn-1","transfer_type":"debit","amount":120000,"content":"coffee","transaction_id":"provider-tx-1"}`
	for i := 0; i < 2; i++ {
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/sepay/bankhub/ipn", strings.NewReader(payload))
		c.Request.Header.Set("Authorization", "Apikey key")
		h.BankHubIPN(c)
		if resp.Code != http.StatusOK {
			t.Fatalf("replay %d got %d: %s", i, resp.Code, resp.Body.String())
		}
	}
	events := store.ListBankFeedEvents(ws.ID, domain.BankFeedEventStateQueued)
	if len(events) != 1 {
		t.Fatalf("expected one queued source event, got %d", len(events))
	}
}

func TestBankHubIPNUnknownAccountIsDurablyQuarantinedAndAcknowledged(t *testing.T) {
	store := storage.NewInMemoryStore()
	h := NewWealthHandler(store, service.NewWealthService(store, nil), &config.Config{SePayBankHubAPIKey: "key"})
	payload := `{"transaction_date":"2026-07-27 10:00:00","bank_account_xid":"not-linked","transfer_type":"credit","amount":20000,"content":"salary","transaction_id":"unknown-tx"}`
	for range 2 {
		resp := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(resp)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/sepay/bankhub/ipn", strings.NewReader(payload))
		c.Request.Header.Set("Authorization", "Apikey key")
		h.BankHubIPN(c)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected durable quarantine ACK, got %d: %s", resp.Code, resp.Body.String())
		}
	}
	first, err := store.QuarantineSePayEvent(domain.SePayUnmappedEvent{Provider: "sepay", BankAccountXID: "not-linked", TransactionID: "unknown-tx", Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.QuarantineSePayEvent(domain.SePayUnmappedEvent{Provider: "sepay", BankAccountXID: "not-linked", TransactionID: "unknown-tx", Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("quarantine retry was not idempotent: %s != %s", first.ID, second.ID)
	}
}

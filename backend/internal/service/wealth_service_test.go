package service

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/storage"
)

func clearSnapshotState() {
	snapshotMu.Lock()
	defer snapshotMu.Unlock()
	for k := range historyByWorkspace {
		delete(historyByWorkspace, k)
	}
	for k := range snapshotVersion {
		delete(snapshotVersion, k)
	}
}

func mustParseMoney(t *testing.T, value string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		t.Fatalf("parse money %q: %v", value, err)
	}
	return v
}

func setupDemoWorkspace(store *storage.InMemoryStore) (*domain.Workspace, domain.ID, error) {
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.CreateWorkspace("Demo Workspace", "VND", uid)
	if err != nil {
		return nil, "", err
	}
	p, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		return nil, "", err
	}
	_, err = store.CreateAccount(domain.Account{
		WorkspaceID: ws.ID,
		PortfolioID: p.ID,
		Name:        "Main",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		return nil, "", err
	}
	return ws, p.ID, nil
}

func TestProcessQueuedBankFeed_OutboundAutoExpense(t *testing.T) {
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}

	conn, err := store.CreateBankConnection(domain.BankConnection{
		WorkspaceID: ws.ID,
		Provider:    "sepay",
		ExternalID:  "conn-1",
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}

	svc := NewWealthService(store, nil)
	feed, err := svc.ProcessSePayIncoming(SePayWebhookEvent{
		ConnectionID: string(conn.ID),
		AccountID:    string(conn.ID),
		Direction:    "out",
		Amount:       "100000",
		Currency:     "VND",
		Description:  "Purchase",
		Reference:    "REF001",
		ExternalID:   "tx-out-1",
		OccurredAt:   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("incoming webhook: %v", err)
	}
	if feed.PostingState != domain.PostingStateAutoReady {
		t.Fatalf("expected auto ready, got %s", feed.PostingState)
	}

	out, err := svc.ProcessQueuedBankFeed(feed.ID)
	if err != nil {
		t.Fatalf("process queued: %v", err)
	}
	if out.PostingState != domain.PostingStatePosted {
		t.Fatalf("expected posted, got %s", out.PostingState)
	}
	if out.PostedTxnID == "" {
		t.Fatal("expected posted transaction id")
	}
	txn, ok := store.GetTransaction(out.PostedTxnID)
	if !ok {
		t.Fatal("posted transaction missing")
	}
	if txn.Type != domain.TransactionTypeExpense {
		t.Fatalf("expected expense, got %s", txn.Type)
	}
}

func TestEnqueueAndProcessSePayEvent(t *testing.T) {
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}
	conn, err := store.CreateBankConnection(domain.BankConnection{
		WorkspaceID: ws.ID,
		Provider:    "sepay",
		ExternalID:  "conn-queue",
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}

	svc := NewWealthService(store, nil)
	event, err := svc.EnqueueSePayIncoming(SePayWebhookEvent{
		ConnectionID: string(conn.ID),
		Direction:    "out",
		Amount:       "120000",
		Currency:     "VND",
		Description:  "Queue test",
		ExternalID:   "evt-1",
		OccurredAt:   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("enqueue webhook: %v", err)
	}
	if event.State != domain.BankFeedEventStateQueued {
		t.Fatalf("expected queued event, got %s", event.State)
	}

	if err := svc.ProcessBankFeedEvent(event.ID); err != nil {
		t.Fatalf("process event: %v", err)
	}

	updated, ok := store.GetBankFeedEvent(event.ID)
	if !ok {
		t.Fatal("event missing after processing")
	}
	if updated.State != domain.BankFeedEventStateDone {
		t.Fatalf("expected event done, got %s", updated.State)
	}

	feeds := store.ListBankFeed(domain.ID(ws.ID))
	if len(feeds) != 1 {
		t.Fatalf("expected 1 feed, got %d", len(feeds))
	}
	out, err := svc.ProcessQueuedBankFeed(feeds[0].ID)
	if err != nil {
		t.Fatalf("process queued feed: %v", err)
	}
	if out.PostingState != domain.PostingStatePosted {
		t.Fatalf("expected feed posted, got %s", out.PostingState)
	}
}

func TestEnqueueSePayEventDeduplicatesByExternalID(t *testing.T) {
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}
	conn, err := store.CreateBankConnection(domain.BankConnection{
		WorkspaceID: ws.ID,
		Provider:    "sepay",
		ExternalID:  "conn-queue-dedupe",
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}

	svc := NewWealthService(store, nil)
	payload := SePayWebhookEvent{
		ConnectionID: string(conn.ID),
		Direction:    "in",
		Amount:       "50000",
		Currency:     "VND",
		Description:  "Dedup test",
		ExternalID:   "evt-dupe",
		OccurredAt:   time.Now().UTC().Format(time.RFC3339),
	}
	first, err := svc.EnqueueSePayIncoming(payload)
	if err != nil {
		t.Fatalf("first enqueue: %v", err)
	}
	second, err := svc.EnqueueSePayIncoming(payload)
	if err != nil {
		t.Fatalf("second enqueue: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected deduped events, got %s and %s", first.ID, second.ID)
	}
}

func TestProcessQueuedBankFeed_InboundAutoIncome(t *testing.T) {
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}

	conn, err := store.CreateBankConnection(domain.BankConnection{
		WorkspaceID: ws.ID,
		Provider:    "sepay",
		ExternalID:  "conn-2",
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}

	svc := NewWealthService(store, nil)
	feed, err := svc.ProcessSePayIncoming(SePayWebhookEvent{
		ConnectionID: string(conn.ID),
		Direction:    "in",
		Amount:       "250000",
		Currency:     "VND",
		Description:  "Luong",
		ExternalID:   "tx-in-1",
		OccurredAt:   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("incoming webhook: %v", err)
	}

	out, err := svc.ProcessQueuedBankFeed(feed.ID)
	if err != nil {
		t.Fatalf("process queued: %v", err)
	}
	if out.PostingState != domain.PostingStatePosted {
		t.Fatalf("expected posted, got %s", out.PostingState)
	}
	if out.Confidence < 70 {
		t.Fatalf("expected auto income confidence, got %.2f", out.Confidence)
	}
	if out.PostedTxnID == "" {
		t.Fatal("expected posted transaction id")
	}
	txn, ok := store.GetTransaction(out.PostedTxnID)
	if !ok {
		t.Fatal("posted transaction missing")
	}
	if txn.Type != domain.TransactionTypeIncome {
		t.Fatalf("expected income, got %s", txn.Type)
	}
}

func TestProcessQueuedBankFeed_InboundTransferNeedsReview(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}

	conn, err := store.CreateBankConnection(domain.BankConnection{
		WorkspaceID: ws.ID,
		Provider:    "sepay",
		ExternalID:  "conn-3",
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}

	svc := NewWealthService(store, nil)
	feed, err := svc.ProcessSePayIncoming(SePayWebhookEvent{
		ConnectionID: string(conn.ID),
		Direction:    "in",
		Amount:       "250000",
		Currency:     "VND",
		Description:  "chuyen tien cho ban",
		ExternalID:   "tx-in-2",
		OccurredAt:   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("incoming webhook: %v", err)
	}

	if _, err := svc.ProcessQueuedBankFeed(feed.ID); err != nil {
		t.Fatalf("process queued: %v", err)
	}

	got, ok := store.GetBankFeed(feed.ID)
	if !ok {
		t.Fatal("feed not found")
	}
	if got.PostingState != domain.PostingStateReview {
		t.Fatalf("expected review for transfer keyword, got %s", got.PostingState)
	}
	if got.PostedTxnID != "" {
		t.Fatal("did not expect posted transaction for transfer-like income")
	}
}

func TestReclassifyBankFeedClearsRuleIDAndPosts(t *testing.T) {
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}

	conn, err := store.CreateBankConnection(domain.BankConnection{
		WorkspaceID: ws.ID,
		Provider:    "sepay",
		ExternalID:  "conn-reclass",
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}

	svc := NewWealthService(store, nil)
	feed, err := svc.ProcessSePayIncoming(SePayWebhookEvent{
		ConnectionID: string(conn.ID),
		Direction:    "in",
		Amount:       "250000",
		Currency:     "VND",
		Description:  "chuyen tien cho ban",
		ExternalID:   "tx-in-reclass",
		OccurredAt:   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("incoming webhook: %v", err)
	}
	got, ok := store.GetBankFeed(feed.ID)
	if !ok {
		t.Fatal("feed not found")
	}
	wasPostedInitially := got.PostingState == domain.PostingStatePosted

	tx, err := svc.ReclassifyBankFeed(feed.ID, "", domain.TransactionTypeExpense, "manual-food", "manual review")
	if err != nil {
		t.Fatalf("reclassify bank feed: %v", err)
	}

	updated, ok := store.GetBankFeed(feed.ID)
	if !ok {
		t.Fatal("feed not found after reclassify")
	}
	if updated.RuleID != "" {
		t.Fatalf("expected manual reclassify to clear rule id, got %s", updated.RuleID)
	}
	if updated.PostingState != domain.PostingStatePosted {
		t.Fatalf("expected posted after reclassify, got %s", updated.PostingState)
	}
	if string(updated.PostedTxnID) == "" {
		t.Fatal("expected posted transaction id after reclassify")
	}

	if tx.ID != updated.PostedTxnID {
		t.Fatalf("mismatch posted transaction id: service=%s feed=%s", tx.ID, updated.PostedTxnID)
	}
	if !wasPostedInitially {
		if tx.Type != domain.TransactionTypeExpense {
			t.Fatalf("expected reclassified transaction type expense, got %s", tx.Type)
		}
		if tx.CategoryID != "manual-food" {
			t.Fatalf("expected reclassified transaction category manual-food, got %s", tx.CategoryID)
		}
	}
}

func TestComputeNetWorthAttributionAndVersioning(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}

	prop, err := store.CreateProperty(domain.Property{
		WorkspaceID: ws.ID,
		PortfolioID: "p-none",
		Name:        "Villa",
		Address:     "HCM",
	})
	if err != nil {
		t.Fatalf("create property: %v", err)
	}
	_, err = store.AddPropertyValuation(domain.PropertyValuation{
		PropertyID:  prop.ID,
		Amount:      "10000.00",
		Currency:    "VND",
		Source:      "manual",
		EffectiveAt: time.Now().UTC().AddDate(0, 0, -60),
	})
	if err != nil {
		t.Fatalf("add property valuation: %v", err)
	}

	asset, err := store.CreateAsset(domain.Asset{
		WorkspaceID: ws.ID,
		PortfolioID: "p-none",
		Name:        "Safe",
		Type:        "gold",
	})
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}
	_, err = store.AddAssetValuation(domain.AssetValuation{
		AssetID:     asset.ID,
		Amount:      "5000.00",
		Currency:    "VND",
		Source:      "manual",
		EffectiveAt: time.Now().UTC().AddDate(0, 0, -60),
	})
	if err != nil {
		t.Fatalf("add asset valuation: %v", err)
	}

	svc := NewWealthService(store, nil)

	first, err := svc.ComputeNetWorth(ws.ID)
	if err != nil {
		t.Fatalf("compute first net worth: %v", err)
	}
	if first.SnapshotVersion != 1 {
		t.Fatalf("expected first snapshot version 1, got %d", first.SnapshotVersion)
	}
	if v := first.NetWorth; v != "15000.00" {
		t.Fatalf("unexpected first net worth: %s", v)
	}
	if first.Attribution.ExternalCashFlow != "0.00" {
		t.Fatalf("unexpected first external cash flow: %s", first.Attribution.ExternalCashFlow)
	}
	if first.NetWorthChange != "0.00" {
		t.Fatalf("unexpected first networth change: %s", first.NetWorthChange)
	}
	if first.DataQuality.StaleValuations != 2 {
		t.Fatalf("expected stale valuations 2, got %d", first.DataQuality.StaleValuations)
	}

	_, err = store.CreateTransaction(domain.Transaction{
		WorkspaceID: ws.ID,
		AccountID:   store.ListAccounts(ws.ID)[0].ID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "3000.00",
		Currency:    "VND",
		OccurredAt:  time.Now().UTC(),
		Status:      domain.TransactionStatusPosted,
	})
	if err != nil {
		t.Fatalf("create income: %v", err)
	}
	_, err = store.AddPropertyValuation(domain.PropertyValuation{
		PropertyID:  prop.ID,
		Amount:      "12000.00",
		Currency:    "VND",
		Source:      "manual",
		EffectiveAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("update property valuation: %v", err)
	}

	second, err := svc.ComputeNetWorth(ws.ID)
	if err != nil {
		t.Fatalf("compute second net worth: %v", err)
	}
	if second.SnapshotVersion != 2 {
		t.Fatalf("expected second snapshot version 2, got %d", second.SnapshotVersion)
	}
	if second.NetWorth != "20000.00" {
		t.Fatalf("unexpected second net worth: %s", second.NetWorth)
	}
	if second.NetWorthChange != "5000.00" {
		t.Fatalf("unexpected second networth change: %s", second.NetWorthChange)
	}
	if second.Attribution.ExternalCashFlow != "3000.00" {
		t.Fatalf("unexpected external cash flow: %s", second.Attribution.ExternalCashFlow)
	}
	if second.Attribution.ValuationChange != "2000.00" {
		t.Fatalf("unexpected valuation change: %s", second.Attribution.ValuationChange)
	}

	firstNet := mustParseMoney(t, first.NetWorth)
	secondNet := mustParseMoney(t, second.NetWorth)
	if fmt.Sprintf("%.2f", secondNet-firstNet) != second.NetWorthChange {
		t.Fatalf("net worth change mismatch: %.2f vs %s", secondNet-firstNet, second.NetWorthChange)
	}
	if secondNet-firstNet != 5000.00 {
		t.Fatalf("expected net worth delta 5000.00, got %.2f", secondNet-firstNet)
	}
	if second.DataQuality.StaleValuations != 1 {
		t.Fatalf("expected stale valuations to update to 1, got %d", second.DataQuality.StaleValuations)
	}
}

func TestComputeNetWorthAtHistoricalCutsByAsOf(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, pID, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}

	base := time.Date(2026, time.July, 15, 8, 0, 0, 0, time.UTC)
	prop, err := store.CreateProperty(domain.Property{
		WorkspaceID: ws.ID,
		PortfolioID: pID,
		Name:        "Historical Land",
		Address:     "HCM",
	})
	if err != nil {
		t.Fatalf("create property: %v", err)
	}
	_, err = store.AddPropertyValuation(domain.PropertyValuation{
		PropertyID:  prop.ID,
		Amount:      "1000.00",
		Currency:    "VND",
		Source:      "manual",
		EffectiveAt: base,
	})
	if err != nil {
		t.Fatalf("add early valuation: %v", err)
	}
	_, err = store.AddPropertyValuation(domain.PropertyValuation{
		PropertyID:  prop.ID,
		Amount:      "2000.00",
		Currency:    "VND",
		Source:      "manual",
		EffectiveAt: base.AddDate(0, 0, 4),
	})
	if err != nil {
		t.Fatalf("add late valuation: %v", err)
	}

	accounts := store.ListAccounts(ws.ID)
	if len(accounts) == 0 {
		t.Fatal("missing account")
	}
	acc := accounts[0]

	_, err = store.CreateTransaction(domain.Transaction{
		WorkspaceID: ws.ID,
		AccountID:   acc.ID,
		PortfolioID: pID,
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
		WorkspaceID: ws.ID,
		AccountID:   acc.ID,
		PortfolioID: pID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "250.00",
		Currency:    "VND",
		OccurredAt:  base.AddDate(0, 0, 6),
		Status:      domain.TransactionStatusPosted,
	})
	if err != nil {
		t.Fatalf("create second income tx: %v", err)
	}

	svc := NewWealthService(store, nil)

	mid, err := svc.GetPortfolioNetWorthAt(string(pID), base.AddDate(0, 0, 3))
	if err != nil {
		t.Fatalf("compute mid as-of: %v", err)
	}
	if mid.NetWorth != "1500.00" {
		t.Fatalf("unexpected mid as-of net worth: %s", mid.NetWorth)
	}

	late, err := svc.GetPortfolioNetWorthAt(string(pID), base.AddDate(0, 0, 7))
	if err != nil {
		t.Fatalf("compute late as-of: %v", err)
	}
	if late.NetWorth != "2750.00" {
		t.Fatalf("unexpected late as-of net worth: %s", late.NetWorth)
	}

	snapshotMu.RLock()
	workspaceHistory := historyByWorkspace[ws.ID]
	snapshotMu.RUnlock()
	if len(workspaceHistory) != 0 {
		t.Fatalf("expected no snapshots written for as-of queries, got %d workspace history entries", len(workspaceHistory))
	}

	current, err := svc.GetPortfolioNetWorth(string(pID))
	if err != nil {
		t.Fatalf("compute current net worth: %v", err)
	}
	if current.NetWorth != "2750.00" {
		t.Fatalf("expected current net worth to match latest as-of, got %s", current.NetWorth)
	}
	snapshotMu.RLock()
	workspaceHistory = historyByWorkspace[ws.ID]
	historyLen := len(workspaceHistory[pID])
	snapshotMu.RUnlock()
	if historyLen != 1 {
		t.Fatalf("expected current compute to persist one snapshot, got %d", historyLen)
	}
}

func TestGetPortfolioNetWorthUsesPortfolioScope(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.CreateWorkspace("Demo Workspace", "VND", uid)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	p1, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		t.Fatal("missing default portfolio")
	}
	p2, err := store.CreatePortfolio(domain.Portfolio{
		WorkspaceID:  ws.ID,
		Name:         "Secondary",
		BaseCurrency: "VND",
	})
	if err != nil {
		t.Fatalf("create second portfolio: %v", err)
	}

	acc1, err := store.CreateAccount(domain.Account{
		WorkspaceID: ws.ID,
		PortfolioID: p1.ID,
		Name:        "Main",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	acc2, err := store.CreateAccount(domain.Account{
		WorkspaceID: ws.ID,
		PortfolioID: p2.ID,
		Name:        "Backup",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	if _, err := store.CreateTransaction(domain.Transaction{
		WorkspaceID: ws.ID,
		AccountID:   acc1.ID,
		PortfolioID: p1.ID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "100.00",
		Currency:    "VND",
		OccurredAt:  time.Now().UTC(),
		Status:      domain.TransactionStatusPosted,
	}); err != nil {
		t.Fatalf("create p1 transaction: %v", err)
	}
	if _, err := store.CreateTransaction(domain.Transaction{
		WorkspaceID: ws.ID,
		AccountID:   acc2.ID,
		PortfolioID: p2.ID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "50.00",
		Currency:    "VND",
		OccurredAt:  time.Now().UTC(),
		Status:      domain.TransactionStatusPosted,
	}); err != nil {
		t.Fatalf("create p2 transaction: %v", err)
	}

	prop1, err := store.CreateProperty(domain.Property{
		WorkspaceID: ws.ID,
		PortfolioID: p1.ID,
		Name:        "Villa",
		Address:     "HCM",
	})
	if err != nil {
		t.Fatalf("create p1 property: %v", err)
	}
	if _, err := store.AddPropertyValuation(domain.PropertyValuation{
		PropertyID:  prop1.ID,
		Amount:      "1000.00",
		Currency:    "VND",
		Source:      "manual",
		EffectiveAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("add p1 valuation: %v", err)
	}

	prop2, err := store.CreateProperty(domain.Property{
		WorkspaceID: ws.ID,
		PortfolioID: p2.ID,
		Name:        "House",
		Address:     "HCM",
	})
	if err != nil {
		t.Fatalf("create p2 property: %v", err)
	}
	if _, err := store.AddPropertyValuation(domain.PropertyValuation{
		PropertyID:  prop2.ID,
		Amount:      "500.00",
		Currency:    "VND",
		Source:      "manual",
		EffectiveAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("add p2 valuation: %v", err)
	}

	svc := NewWealthService(store, nil)
	p1Net, err := svc.GetPortfolioNetWorth(string(p1.ID))
	if err != nil {
		t.Fatalf("compute p1 net worth: %v", err)
	}
	p2Net, err := svc.GetPortfolioNetWorth(string(p2.ID))
	if err != nil {
		t.Fatalf("compute p2 net worth: %v", err)
	}

	if p1Net.NetWorth != "1100.00" {
		t.Fatalf("expected p1 net worth 1100.00, got %s", p1Net.NetWorth)
	}
	if p2Net.NetWorth != "550.00" {
		t.Fatalf("expected p2 net worth 550.00, got %s", p2Net.NetWorth)
	}
	if p1Net.SnapshotVersion != 1 {
		t.Fatalf("expected p1 snapshot version 1, got %d", p1Net.SnapshotVersion)
	}
	if p2Net.SnapshotVersion != 1 {
		t.Fatalf("expected p2 snapshot version 1, got %d", p2Net.SnapshotVersion)
	}

	wsNet, err := svc.ComputeNetWorth(ws.ID)
	if err != nil {
		t.Fatalf("compute workspace net worth: %v", err)
	}
	if wsNet.NetWorth != "1650.00" {
		t.Fatalf("expected workspace net worth 1650.00, got %s", wsNet.NetWorth)
	}
	if wsNet.SnapshotVersion != 1 {
		t.Fatalf("expected workspace snapshot version 1, got %d", wsNet.SnapshotVersion)
	}
}

func TestGetPortfolioSnapshotsPagination(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}
	acc := store.ListAccounts(ws.ID)
	if len(acc) == 0 {
		t.Fatalf("missing account")
	}

	svc := NewWealthService(store, nil)
	for i := 0; i < 5; i++ {
		_, err := store.CreateTransaction(domain.Transaction{
			WorkspaceID: ws.ID,
			AccountID:   acc[0].ID,
			Type:        domain.TransactionTypeIncome,
			Amount:      "10.00",
			Currency:    "VND",
			OccurredAt:  time.Now().UTC(),
			Status:      domain.TransactionStatusPosted,
		})
		if err != nil {
			t.Fatalf("create transaction %d: %v", i, err)
		}
		if _, err := svc.ComputeNetWorth(ws.ID); err != nil {
			t.Fatalf("compute net worth %d: %v", i, err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	page1 := svc.GetPortfolioSnapshots(ws.ID, 3, "")
	if len(page1.Items) != 3 {
		t.Fatalf("expected first page 3 items, got %d", len(page1.Items))
	}
	if page1.NextCursor == "" {
		t.Fatalf("expected nextCursor for first page")
	}
	if page1.Items[0].SnapshotVersion <= page1.Items[1].SnapshotVersion {
		t.Fatalf("expected descending snapshot versions")
	}
	if page1.Items[1].SnapshotVersion <= page1.Items[2].SnapshotVersion {
		t.Fatalf("expected descending snapshot versions")
	}

	page2 := svc.GetPortfolioSnapshots(ws.ID, 3, page1.NextCursor)
	if len(page2.Items) != 2 {
		t.Fatalf("expected second page 2 items, got %d", len(page2.Items))
	}
	if page2.NextCursor != "" {
		t.Fatalf("expected empty nextCursor on last page")
	}
	if page2.Items[0].SnapshotVersion >= page1.Items[2].SnapshotVersion {
		t.Fatalf("expected older versions on next page")
	}
}

func TestGetPortfolioSnapshotsForPortfolio(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.CreateWorkspace("Demo", "VND", uid)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	p1, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		t.Fatalf("missing portfolio")
	}
	p2, err := store.CreatePortfolio(domain.Portfolio{
		WorkspaceID:  ws.ID,
		Name:         "Backup",
		BaseCurrency: "VND",
	})
	if err != nil {
		t.Fatalf("create portfolio 2: %v", err)
	}

	acc1, err := store.CreateAccount(domain.Account{
		WorkspaceID: ws.ID,
		PortfolioID: p1.ID,
		Name:        "Main",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account p1: %v", err)
	}
	acc2, err := store.CreateAccount(domain.Account{
		WorkspaceID: ws.ID,
		PortfolioID: p2.ID,
		Name:        "Backup",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account p2: %v", err)
	}

	svc := NewWealthService(store, nil)
	if _, err := store.CreateTransaction(domain.Transaction{
		WorkspaceID: ws.ID,
		AccountID:   acc1.ID,
		PortfolioID: p1.ID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "10.00",
		Currency:    "VND",
		OccurredAt:  time.Now().UTC(),
		Status:      domain.TransactionStatusPosted,
	}); err != nil {
		t.Fatalf("create tx p1 first: %v", err)
	}
	if _, err := svc.GetPortfolioNetWorth(string(p1.ID)); err != nil {
		t.Fatalf("compute p1 snapshot 1: %v", err)
	}

	if _, err := store.CreateTransaction(domain.Transaction{
		WorkspaceID: ws.ID,
		AccountID:   acc1.ID,
		PortfolioID: p1.ID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "20.00",
		Currency:    "VND",
		OccurredAt:  time.Now().UTC(),
		Status:      domain.TransactionStatusPosted,
	}); err != nil {
		t.Fatalf("create tx p1 second: %v", err)
	}
	if _, err := svc.GetPortfolioNetWorth(string(p1.ID)); err != nil {
		t.Fatalf("compute p1 snapshot 2: %v", err)
	}

	if _, err := store.CreateTransaction(domain.Transaction{
		WorkspaceID: ws.ID,
		AccountID:   acc2.ID,
		PortfolioID: p2.ID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "30.00",
		Currency:    "VND",
		OccurredAt:  time.Now().UTC(),
		Status:      domain.TransactionStatusPosted,
	}); err != nil {
		t.Fatalf("create tx p2: %v", err)
	}
	if _, err := svc.GetPortfolioNetWorth(string(p2.ID)); err != nil {
		t.Fatalf("compute p2 snapshot 1: %v", err)
	}

	p1Snapshots := svc.GetPortfolioSnapshotsForPortfolio(ws.ID, p1.ID, 10, "")
	if len(p1Snapshots.Items) != 2 {
		t.Fatalf("expected 2 p1 snapshots, got %d", len(p1Snapshots.Items))
	}
	if p1Snapshots.Items[0].SnapshotVersion != 2 {
		t.Fatalf("expected latest p1 version 2, got %d", p1Snapshots.Items[0].SnapshotVersion)
	}
	if p1Snapshots.Items[1].SnapshotVersion != 1 {
		t.Fatalf("expected older p1 version 1, got %d", p1Snapshots.Items[1].SnapshotVersion)
	}

	p2Snapshots := svc.GetPortfolioSnapshotsForPortfolio(ws.ID, p2.ID, 10, "")
	if len(p2Snapshots.Items) != 1 {
		t.Fatalf("expected 1 p2 snapshot, got %d", len(p2Snapshots.Items))
	}
	if p2Snapshots.Items[0].SnapshotVersion != 1 {
		t.Fatalf("expected p2 version 1, got %d", p2Snapshots.Items[0].SnapshotVersion)
	}
}

func TestCreateTransactionRejectsInvalidStatus(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}
	accs := store.ListAccounts(ws.ID)
	if len(accs) == 0 {
		t.Fatal("missing account")
	}
	svc := NewWealthService(store, nil)

	_, err = svc.CreateTransaction(domain.Transaction{
		WorkspaceID: ws.ID,
		AccountID:   accs[0].ID,
		Type:        domain.TransactionTypeExpense,
		Amount:      "12000",
		Currency:    "VND",
		Status:      domain.TransactionStatus("unknown"),
		OccurredAt:  time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestCreateTransactionRejectsCrossWorkspaceAccount(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws1, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace 1: %v", err)
	}
	ws2, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace 2: %v", err)
	}
	accounts1 := store.ListAccounts(ws1.ID)
	accounts2 := store.ListAccounts(ws2.ID)
	if len(accounts1) == 0 || len(accounts2) == 0 {
		t.Fatal("missing accounts in workspace setup")
	}
	svc := NewWealthService(store, nil)

	_, err = svc.CreateTransaction(domain.Transaction{
		WorkspaceID: ws1.ID,
		AccountID:   accounts2[0].ID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "50000",
		Currency:    "VND",
		OccurredAt:  time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected cross-workspace account error")
	}
}

func TestPendingTransactionDoesNotAffectNetWorth(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}
	accs := store.ListAccounts(ws.ID)
	if len(accs) == 0 {
		t.Fatal("missing account")
	}
	svc := NewWealthService(store, nil)

	posted, err := svc.CreateTransaction(domain.Transaction{
		WorkspaceID: ws.ID,
		AccountID:   accs[0].ID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "150",
		Currency:    "VND",
		Status:      domain.TransactionStatusPosted,
		OccurredAt:  time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create posted tx: %v", err)
	}
	_ = posted

	_, err = svc.CreateTransaction(domain.Transaction{
		WorkspaceID: ws.ID,
		AccountID:   accs[0].ID,
		Type:        domain.TransactionTypeIncome,
		Amount:      "150",
		Currency:    "VND",
		Status:      domain.TransactionStatusPending,
		OccurredAt:  time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create pending tx: %v", err)
	}

	res, err := svc.ComputeNetWorth(ws.ID)
	if err != nil {
		t.Fatalf("compute net worth: %v", err)
	}
	if res.Cash != "150.00" {
		t.Fatalf("expected cash 150.00, got %s", res.Cash)
	}
	if res.NetWorth != "150.00" {
		t.Fatalf("expected net worth 150.00, got %s", res.NetWorth)
	}
}

func TestCreateTransactionRejectsNonPositiveAmount(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoWorkspace(store)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}
	accs := store.ListAccounts(ws.ID)
	if len(accs) == 0 {
		t.Fatal("missing account")
	}
	svc := NewWealthService(store, nil)

	_, err = svc.CreateTransaction(domain.Transaction{
		WorkspaceID: ws.ID,
		AccountID:   accs[0].ID,
		Type:        domain.TransactionTypeExpense,
		Amount:      "0",
		Currency:    "VND",
		OccurredAt:  time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected non-positive amount error")
	}
}

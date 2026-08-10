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
	for k := range historyByUser {
		delete(historyByUser, k)
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

func setupDemoUser(store *storage.InMemoryStore) (*domain.User, domain.ID, error) {
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo User", "VND", uid)
	if err != nil {
		return nil, "", err
	}
	p, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		return nil, "", err
	}
	_, err = store.CreateAccount(domain.Account{
		UserID:      ws.ID,
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

func TestCreateTransactionWithoutCategoryLeavesCategoryEmpty(t *testing.T) {
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}
	accounts := store.ListAccounts(ws.ID)
	if len(accounts) == 0 {
		t.Fatal("expected a test account")
	}

	transaction, err := NewWealthService(store, nil).CreateTransaction(domain.Transaction{
		UserID:     ws.ID,
		AccountID:  accounts[0].ID,
		Type:       domain.TransactionTypeExpense,
		Amount:     "1",
		Currency:   "VND",
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create transaction without category: %v", err)
	}
	if transaction.CategoryID != "" {
		t.Fatalf("expected an empty category ID, got %q", transaction.CategoryID)
	}
}

func TestProcessSePayIncoming_OutboundRequiresReview(t *testing.T) {
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}

	conn, err := store.CreateBankConnection(domain.BankConnection{
		UserID:     ws.ID,
		Provider:   "sepay",
		ExternalID: "conn-1",
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
	if feed.PostingState != domain.PostingStateReview {
		t.Fatalf("expected review, got %s", feed.PostingState)
	}
	if feed.PostedTxnID != "" {
		t.Fatal("webhook must not create a ledger transaction")
	}
}

func TestDailyRateLoanAccrualAndMonthEndSchedule(t *testing.T) {
	store := storage.NewInMemoryStore()
	ws, portfolioID, err := setupDemoUser(store)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	loan, err := store.CreateLoan(domain.Loan{
		UserID: ws.ID, PortfolioID: portfolioID, Counterparty: "Anh Minh",
		Direction: domain.LoanDirectionReceivable, PrincipalInitial: "100000000", PrincipalBalance: "100000000",
		AnnualRate: "0", DailyRatePerMillion: "3000", StartAt: start, DueAt: start.AddDate(0, 1, 0), Status: domain.LoanStatusActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	svc := NewWealthService(store, nil)
	rows, total := svc.loanAccrualsByLoan(loan, start.AddDate(0, 0, 10))
	if len(rows) != 1 || total.TotalAccrued != 3000000 {
		t.Fatalf("expected 3,000,000 accrued after 10 days, got %#v / %v", rows, total.TotalAccrued)
	}
	next := nextMonthlyPaymentDate(start, start)
	if want := time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC); !next.Equal(want) {
		t.Fatalf("expected month-end payment %s, got %s", want, next)
	}
	if got := loanDailyInterest(loan); got != 300000 {
		t.Fatalf("expected 300,000 daily interest, got %v", got)
	}
}

func TestCreateTransferWritesBothLedgerEntries(t *testing.T) {
	store := storage.NewInMemoryStore()
	user, portfolioID, err := setupDemoUser(store)
	if err != nil {
		t.Fatal(err)
	}
	accounts := store.ListAccounts(user.ID)
	if len(accounts) == 0 {
		t.Fatal("missing source account")
	}
	from := accounts[0]
	to, err := store.CreateAccount(domain.Account{UserID: user.ID, PortfolioID: portfolioID, Name: "Savings", Type: "bank", Currency: "VND"})
	if err != nil {
		t.Fatal(err)
	}

	transfer, err := NewWealthService(store, nil).CreateTransfer(domain.Transfer{
		UserID: user.ID, FromAccountID: from.ID, ToAccountID: to.ID, Amount: "250000", Currency: "VND", OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}
	if transfer.ID == "" {
		t.Fatal("expected persisted transfer")
	}
	if got := len(store.ListTransactions(user.ID, from.ID)); got != 1 {
		t.Fatalf("expected one outgoing ledger transaction, got %d", got)
	}
	if got := len(store.ListTransactions(user.ID, to.ID)); got != 1 {
		t.Fatalf("expected one incoming ledger transaction, got %d", got)
	}
}

func TestCreateLoanPaymentAtomicallyWritesLedgerAndBalance(t *testing.T) {
	store := storage.NewInMemoryStore()
	user, portfolioID, err := setupDemoUser(store)
	if err != nil {
		t.Fatal(err)
	}
	account := store.ListAccounts(user.ID)[0]
	loan, err := store.CreateLoan(domain.Loan{
		UserID: user.ID, PortfolioID: portfolioID, Counterparty: "Borrower", Direction: domain.LoanDirectionReceivable,
		PrincipalInitial: "100000", PrincipalBalance: "100000", AnnualRate: "0", StartAt: time.Now().UTC(), DueAt: time.Now().UTC().AddDate(0, 1, 0), Status: domain.LoanStatusActive,
	})
	if err != nil {
		t.Fatal(err)
	}

	payment, err := NewWealthService(store, nil).CreateLoanPayment(string(loan.ID), domain.LoanPayment{
		AccountID: account.ID, Principal: "40000", Interest: "1000", Fee: "0", Waived: "0", OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create loan payment: %v", err)
	}
	if payment.TransactionID == "" {
		t.Fatal("expected a linked ledger transaction")
	}
	updated, ok := store.GetLoan(loan.ID)
	if !ok || updated.PrincipalBalance != "60000.00" {
		t.Fatalf("expected principal balance 60000.00, got %#v", updated)
	}
	if got := len(store.ListLoanPayments(user.ID, loan.ID)); got != 1 {
		t.Fatalf("expected one payment, got %d", got)
	}
	if got := len(store.ListTransactions(user.ID, account.ID)); got != 1 {
		t.Fatalf("expected one ledger transaction, got %d", got)
	}
}

func TestInterestReceiptStartsTheNextAccrualPeriodAndKeepsItsHistory(t *testing.T) {
	store := storage.NewInMemoryStore()
	user, portfolioID, err := setupDemoUser(store)
	if err != nil {
		t.Fatal(err)
	}
	account := store.ListAccounts(user.ID)[0]
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	loan, err := store.CreateLoan(domain.Loan{
		UserID: user.ID, PortfolioID: portfolioID, Counterparty: "Borrower", Direction: domain.LoanDirectionReceivable,
		PrincipalInitial: "50000000", PrincipalBalance: "50000000", AnnualRate: "0", DailyRatePerMillion: "1000",
		StartAt: start, DueAt: start.AddDate(0, 1, 0), Status: domain.LoanStatusActive,
	})
	if err != nil {
		t.Fatal(err)
	}

	svc := NewWealthService(store, nil)
	receivedAt := start.AddDate(0, 0, 3)
	payment, err := svc.CreateLoanPayment(string(loan.ID), domain.LoanPayment{
		AccountID: account.ID, Principal: "0", Interest: "150000", InterestDays: 3, Fee: "0", Waived: "0", OccurredAt: receivedAt,
	})
	if err != nil {
		t.Fatalf("record interest receipt: %v", err)
	}
	if payment.InterestDays != 3 {
		t.Fatalf("expected history to preserve 3 interest days, got %d", payment.InterestDays)
	}

	rows, summary := svc.loanAccrualsByLoan(loan, start.AddDate(0, 0, 5))
	if len(rows) != 2 || rows[0].Days != 3 || rows[1].Days != 2 {
		t.Fatalf("expected accrual periods 3 days then 2 days, got %#v", rows)
	}
	if summary.TotalPaid != 150000 || summary.TotalAccrued != 250000 {
		t.Fatalf("expected 150,000 paid and 250,000 accrued, got %#v", summary)
	}
	if last := lastInterestPaymentDate(store.ListLoanPayments(user.ID, loan.ID)); !last.Equal(receivedAt) {
		t.Fatalf("expected last interest receipt on %s, got %s", receivedAt, last)
	}
}

func TestInterestPaidInAdvanceContinuesAccruingFromThePaymentDate(t *testing.T) {
	store := storage.NewInMemoryStore()
	user, portfolioID, err := setupDemoUser(store)
	if err != nil {
		t.Fatal(err)
	}
	account := store.ListAccounts(user.ID)[0]
	start := time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC)
	loan, err := store.CreateLoan(domain.Loan{
		UserID: user.ID, PortfolioID: portfolioID, Counterparty: "Borrower", Direction: domain.LoanDirectionReceivable,
		PrincipalInitial: "200000000", PrincipalBalance: "200000000", AnnualRate: "0", DailyRatePerMillion: "3000",
		DayCountBasis: "ACT_365", StartAt: start, DueAt: start.AddDate(0, 3, 0), Status: domain.LoanStatusActive,
	})
	if err != nil {
		t.Fatal(err)
	}

	svc := NewWealthService(store, nil)
	paidAt := time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)
	if _, err := svc.CreateLoanPayment(string(loan.ID), domain.LoanPayment{
		AccountID: account.ID, Principal: "0", Interest: "6000000", InterestDays: 10, Fee: "0", Waived: "0", OccurredAt: paidAt,
	}); err != nil {
		t.Fatalf("record advance interest: %v", err)
	}

	rows, summary := svc.loanAccrualsByLoan(loan, time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC))
	if len(rows) != 2 || rows[0].Days != 10 || rows[1].Days != 5 {
		t.Fatalf("expected 10/08–20/08 then 20/08–25/08, got %#v", rows)
	}
	if summary.TotalAccrued != 9000000 || summary.TotalPaid != 6000000 || rows[1].UnpaidInterest != "3000000.00" {
		t.Fatalf("expected 9m accrued, 6m paid and 3m still unpaid, got summary=%#v rows=%#v", summary, rows)
	}
}

func TestEnqueueAndProcessSePayEvent(t *testing.T) {
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}
	conn, err := store.CreateBankConnection(domain.BankConnection{
		UserID:     ws.ID,
		Provider:   "sepay",
		ExternalID: "conn-queue",
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
	if feeds[0].PostingState != domain.PostingStateReview {
		t.Fatalf("expected feed pending review, got %s", feeds[0].PostingState)
	}
}

func TestEnqueueSePayEventDeduplicatesByExternalID(t *testing.T) {
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}
	conn, err := store.CreateBankConnection(domain.BankConnection{
		UserID:     ws.ID,
		Provider:   "sepay",
		ExternalID: "conn-queue-dedupe",
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

func TestProcessSePayIncoming_InboundRequiresReview(t *testing.T) {
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}

	conn, err := store.CreateBankConnection(domain.BankConnection{
		UserID:     ws.ID,
		Provider:   "sepay",
		ExternalID: "conn-2",
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

	if feed.PostingState != domain.PostingStateReview {
		t.Fatalf("expected feed pending review, got %s", feed.PostingState)
	}
	if feed.PostedTxnID != "" {
		t.Fatal("webhook must not create income")
	}
}

func TestSuggestionUsesRememberedFeedbackBeforeOtherMatchersAndKeepsFeedImmutable(t *testing.T) {
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("classifier@example.test", "Classifier", "pass")
	ws, err := store.EnsureUserPortfolio("Classifier", "VND", userID)
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
	connection, err := store.CreateBankConnection(domain.BankConnection{UserID: ws.ID, Provider: "sepay", ExternalID: "provider-a"})
	if err != nil {
		t.Fatalf("connection: %v", err)
	}
	prior, err := store.IngestBankFeed(domain.BankFeedTransaction{UserID: userID, ConnectionID: connection.ID, AccountID: account.ID, Amount: "50000", Currency: "VND", Direction: "out", Description: "CA PHE HIGHLANDS", Reference: "ABC", OccurredAt: time.Now().Add(-time.Hour), ExternalID: "prior", PostingState: domain.PostingStatePosted, ClassificationStatus: "confirmed"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateClassificationFeedback(domain.ClassificationFeedback{BankFeedTransactionID: prior.ID, UserID: userID, Action: "corrected", Name: "Cà phê", CategoryID: "food", AccountID: account.ID, TransactionType: "expense", RememberChoice: true})
	if err != nil {
		t.Fatal(err)
	}
	current, err := store.IngestBankFeed(domain.BankFeedTransaction{UserID: userID, ConnectionID: connection.ID, AccountID: account.ID, Amount: "72000", Currency: "VND", Direction: "out", Description: "ca phe highlands", Reference: "abc", OccurredAt: time.Now(), ExternalID: "current", PostingState: domain.PostingStateReview, ClassificationStatus: "needs_review", RawProviderData: `{"amount":"72000"}`})
	if err != nil {
		t.Fatal(err)
	}
	svc := NewWealthService(store, nil)
	svc.createSuggestionIfMissing(current)
	suggestions := store.ListTransactionSuggestions(current.ID)
	if len(suggestions) != 1 {
		t.Fatalf("suggestions=%d", len(suggestions))
	}
	if suggestions[0].Source != "rule" || suggestions[0].Confidence != 100 || suggestions[0].SuggestedName != "Cà phê" {
		t.Fatalf("unexpected suggestion: %+v", suggestions[0])
	}
	stored, ok := store.GetBankFeed(current.ID)
	if !ok || stored.Amount != "72000" || stored.RawProviderData != `{"amount":"72000"}` {
		t.Fatalf("classifier mutated immutable feed: %+v", stored)
	}
}

func TestComparableHistoryRequiresAccountAmountAndCadence(t *testing.T) {
	now := time.Now().UTC()
	feed := domain.BankFeedTransaction{Timestamped: domain.Timestamped{ID: "new"}, AccountID: "account-1", Direction: "out", Amount: "100000", Description: "Netflix monthly", OccurredAt: now, PostingState: domain.PostingStateReview}
	prior := domain.BankFeedTransaction{Timestamped: domain.Timestamped{ID: "old"}, AccountID: "account-1", Direction: "out", Amount: "103000", Description: "NETFLIX MONTHLY", OccurredAt: now.AddDate(0, 0, -30), PostingState: domain.PostingStatePosted, PostedTxnID: "txn"}
	needle := normalizeSuggestionText(feed.Description + " " + feed.Reference)
	if !isComparableHistoryFeed(feed, prior, needle) {
		t.Fatal("expected monthly same-account history to match")
	}
	prior.AccountID = "another-account"
	if isComparableHistoryFeed(feed, prior, needle) {
		t.Fatal("different account must not match")
	}
	prior.AccountID, prior.Amount = "account-1", "140000"
	if isComparableHistoryFeed(feed, prior, needle) {
		t.Fatal("amount outside range must not match")
	}
	prior.Amount, prior.OccurredAt = "103000", now.AddDate(0, 0, -10)
	if isComparableHistoryFeed(feed, prior, needle) {
		t.Fatal("unrelated cadence must not match")
	}
}

func TestTokenJaccardSupportsConservativeSemanticMatching(t *testing.T) {
	if score := tokenJaccard(normalizeSuggestionText("THANH TOAN Grab"), normalizeSuggestionText("grab thanh toan don hang")); score < 0.60 {
		t.Fatalf("expected merchant token overlap, got %v", score)
	}
	if score := tokenJaccard("luong cong ty", "thanh toan dien"); score != 0 {
		t.Fatalf("unrelated content must not match, got %v", score)
	}
}

func TestProcessQueuedBankFeed_InboundTransferNeedsReview(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}

	conn, err := store.CreateBankConnection(domain.BankConnection{
		UserID:     ws.ID,
		Provider:   "sepay",
		ExternalID: "conn-3",
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
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}

	conn, err := store.CreateBankConnection(domain.BankConnection{
		UserID:     ws.ID,
		Provider:   "sepay",
		ExternalID: "conn-reclass",
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

	tx, err := svc.ReclassifyBankFeed(feed.ID, "", domain.TransactionTypeExpense, "", "manual review")
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
		if tx.CategoryID != "" {
			t.Fatalf("expected reclassified transaction without a category, got %s", tx.CategoryID)
		}
	}
}

func TestComputeNetWorthAttributionAndVersioning(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}

	prop, err := store.CreateProperty(domain.Property{
		UserID:      ws.ID,
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
		UserID:      ws.ID,
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
		UserID:     ws.ID,
		AccountID:  store.ListAccounts(ws.ID)[0].ID,
		Type:       domain.TransactionTypeIncome,
		Amount:     "3000.00",
		Currency:   "VND",
		OccurredAt: time.Now().UTC(),
		Status:     domain.TransactionStatusPosted,
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
	ws, pID, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}

	base := time.Date(2026, time.July, 15, 8, 0, 0, 0, time.UTC)
	prop, err := store.CreateProperty(domain.Property{
		UserID:      ws.ID,
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
		UserID:      ws.ID,
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
		UserID:      ws.ID,
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
	userHistory := historyByUser[ws.ID]
	snapshotMu.RUnlock()
	if len(userHistory) != 0 {
		t.Fatalf("expected no snapshots written for as-of queries, got %d user history entries", len(userHistory))
	}

	current, err := svc.GetPortfolioNetWorth(string(pID))
	if err != nil {
		t.Fatalf("compute current net worth: %v", err)
	}
	if current.NetWorth != "2750.00" {
		t.Fatalf("expected current net worth to match latest as-of, got %s", current.NetWorth)
	}
	snapshotMu.RLock()
	userHistory = historyByUser[ws.ID]
	historyLen := len(userHistory[pID])
	snapshotMu.RUnlock()
	if historyLen != 1 {
		t.Fatalf("expected current compute to persist one snapshot, got %d", historyLen)
	}
}

func TestGetPortfolioNetWorthUsesPortfolioScope(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	uid := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "pass")
	ws, err := store.EnsureUserPortfolio("Demo User", "VND", uid)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	p1, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		t.Fatal("missing default portfolio")
	}
	p2, err := store.CreatePortfolio(domain.Portfolio{
		UserID:       ws.ID,
		Name:         "Secondary",
		BaseCurrency: "VND",
	})
	if err != nil {
		t.Fatalf("create second portfolio: %v", err)
	}

	acc1, err := store.CreateAccount(domain.Account{
		UserID:      ws.ID,
		PortfolioID: p1.ID,
		Name:        "Main",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	acc2, err := store.CreateAccount(domain.Account{
		UserID:      ws.ID,
		PortfolioID: p2.ID,
		Name:        "Backup",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	if _, err := store.CreateTransaction(domain.Transaction{
		UserID:      ws.ID,
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
		UserID:      ws.ID,
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
		UserID:      ws.ID,
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
		UserID:      ws.ID,
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
		t.Fatalf("compute user net worth: %v", err)
	}
	if wsNet.NetWorth != "1650.00" {
		t.Fatalf("expected user net worth 1650.00, got %s", wsNet.NetWorth)
	}
	if wsNet.SnapshotVersion != 1 {
		t.Fatalf("expected user snapshot version 1, got %d", wsNet.SnapshotVersion)
	}
}

func TestGetPortfolioSnapshotsPagination(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}
	acc := store.ListAccounts(ws.ID)
	if len(acc) == 0 {
		t.Fatalf("missing account")
	}

	svc := NewWealthService(store, nil)
	for i := 0; i < 5; i++ {
		_, err := store.CreateTransaction(domain.Transaction{
			UserID:     ws.ID,
			AccountID:  acc[0].ID,
			Type:       domain.TransactionTypeIncome,
			Amount:     "10.00",
			Currency:   "VND",
			OccurredAt: time.Now().UTC(),
			Status:     domain.TransactionStatusPosted,
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
	ws, err := store.EnsureUserPortfolio("Demo", "VND", uid)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	p1, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		t.Fatalf("missing portfolio")
	}
	p2, err := store.CreatePortfolio(domain.Portfolio{
		UserID:       ws.ID,
		Name:         "Backup",
		BaseCurrency: "VND",
	})
	if err != nil {
		t.Fatalf("create portfolio 2: %v", err)
	}

	acc1, err := store.CreateAccount(domain.Account{
		UserID:      ws.ID,
		PortfolioID: p1.ID,
		Name:        "Main",
		Type:        "cash",
		Currency:    "VND",
	})
	if err != nil {
		t.Fatalf("create account p1: %v", err)
	}
	acc2, err := store.CreateAccount(domain.Account{
		UserID:      ws.ID,
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
		UserID:      ws.ID,
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
		UserID:      ws.ID,
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
		UserID:      ws.ID,
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
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}
	accs := store.ListAccounts(ws.ID)
	if len(accs) == 0 {
		t.Fatal("missing account")
	}
	svc := NewWealthService(store, nil)

	_, err = svc.CreateTransaction(domain.Transaction{
		UserID:     ws.ID,
		AccountID:  accs[0].ID,
		Type:       domain.TransactionTypeExpense,
		Amount:     "12000",
		Currency:   "VND",
		Status:     domain.TransactionStatus("unknown"),
		OccurredAt: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestCreateTransactionRejectsCrossUserAccount(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws1, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user 1: %v", err)
	}
	otherID := store.SeedDemoUser("other@example.com", "Other User", "pass")
	ws2, err := store.EnsureUserPortfolio("", "VND", otherID)
	if err != nil {
		t.Fatalf("prepare user 2: %v", err)
	}
	account2, err := store.CreateAccount(domain.Account{UserID: ws2.ID, Name: "Other bank", Type: "bank", Currency: "VND"})
	if err != nil {
		t.Fatalf("create user 2 account: %v", err)
	}
	accounts1 := store.ListAccounts(ws1.ID)
	if len(accounts1) == 0 {
		t.Fatal("missing accounts in user setup")
	}
	svc := NewWealthService(store, nil)

	_, err = svc.CreateTransaction(domain.Transaction{
		UserID:     ws1.ID,
		AccountID:  account2.ID,
		Type:       domain.TransactionTypeIncome,
		Amount:     "50000",
		Currency:   "VND",
		OccurredAt: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected cross-user account error")
	}
}

func TestPendingTransactionDoesNotAffectNetWorth(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}
	accs := store.ListAccounts(ws.ID)
	if len(accs) == 0 {
		t.Fatal("missing account")
	}
	svc := NewWealthService(store, nil)

	posted, err := svc.CreateTransaction(domain.Transaction{
		UserID:     ws.ID,
		AccountID:  accs[0].ID,
		Type:       domain.TransactionTypeIncome,
		Amount:     "150",
		Currency:   "VND",
		Status:     domain.TransactionStatusPosted,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create posted tx: %v", err)
	}
	_ = posted

	_, err = svc.CreateTransaction(domain.Transaction{
		UserID:     ws.ID,
		AccountID:  accs[0].ID,
		Type:       domain.TransactionTypeIncome,
		Amount:     "150",
		Currency:   "VND",
		Status:     domain.TransactionStatusPending,
		OccurredAt: time.Now().UTC(),
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

func TestAccountBalanceOverrideReplacesLedgerCashWithoutTransaction(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	userID := store.SeedDemoUser("override@wealthos.vn", "Override User", "pass")
	ws, err := store.EnsureUserPortfolio("Override", "VND", userID)
	if err != nil {
		t.Fatal(err)
	}
	portfolio, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		t.Fatal("missing portfolio")
	}
	account, err := store.CreateAccount(domain.Account{
		UserID:            ws.ID,
		PortfolioID:       portfolio.ID,
		Name:              "Cash",
		Type:              "cash",
		Currency:          "VND",
		BalanceOverride:   "42",
		BalanceOverrideAt: time.Now().UTC().Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	svc := NewWealthService(store, nil)
	if _, err := svc.CreateTransaction(domain.Transaction{UserID: ws.ID, AccountID: account.ID, PortfolioID: portfolio.ID, Type: domain.TransactionTypeIncome, Amount: "100", Currency: "VND", Status: domain.TransactionStatusPosted, OccurredAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	result, err := svc.ComputeNetWorth(ws.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Cash != "42.00" || result.NetWorth != "42.00" {
		t.Fatalf("expected balance override to replace ledger cash, got %+v", result)
	}
}

func TestCreateTransactionRejectsNonPositiveAmount(t *testing.T) {
	clearSnapshotState()
	store := storage.NewInMemoryStore()
	ws, _, err := setupDemoUser(store)
	if err != nil {
		t.Fatalf("prepare user: %v", err)
	}
	accs := store.ListAccounts(ws.ID)
	if len(accs) == 0 {
		t.Fatal("missing account")
	}
	svc := NewWealthService(store, nil)

	_, err = svc.CreateTransaction(domain.Transaction{
		UserID:     ws.ID,
		AccountID:  accs[0].ID,
		Type:       domain.TransactionTypeExpense,
		Amount:     "0",
		Currency:   "VND",
		OccurredAt: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected non-positive amount error")
	}
}

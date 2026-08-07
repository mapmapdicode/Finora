package storage

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"wealthos-backend/internal/domain"
)

type InMemoryStore struct {
	mu                      sync.RWMutex
	users                   map[domain.ID]*domain.User
	emailVerificationTokens map[string]emailVerificationToken
	portfolios              map[domain.ID]*domain.Portfolio
	accounts                map[domain.ID]*domain.Account
	botAccountKeys          map[domain.ID]*domain.BotAccountKey
	transactions            map[domain.ID]*domain.Transaction
	transfers               map[domain.ID]*domain.Transfer
	customers               map[domain.ID]*domain.Customer
	loans                   map[domain.ID]*domain.Loan
	loanPayments            map[domain.ID]*domain.LoanPayment
	properties              map[domain.ID]*domain.Property
	propertyValues          map[domain.ID]*domain.PropertyValuation
	assets                  map[domain.ID]*domain.Asset
	assetValues             map[domain.ID]*domain.AssetValuation
	budgets                 map[domain.ID]*domain.Budget
	budgetAllocs            map[domain.ID]*domain.BudgetAllocation
	forecast                map[domain.ID]*domain.ForecastScenario
	bankConnections         map[domain.ID]*domain.BankConnection
	bankFeed                map[domain.ID]*domain.BankFeedTransaction
	bankFeedKeys            map[string]domain.ID
	bankFeedEvents          map[domain.ID]*domain.BankFeedEvent
	bankFeedEventKeys       map[string]domain.ID
	bankRecon               map[domain.ID]*domain.BankReconciliation
	automationRules         map[domain.ID]*domain.AutomationRule
	bankPaymentReqs         map[domain.ID]*domain.BankPaymentRequest
	assistantCmds           map[domain.ID]*domain.AssistantCommand
	auditLogs               map[domain.ID]*domain.AuditLog
	userSettings            map[domain.ID]*domain.UserSettings
	sepayProfiles           map[domain.ID]*domain.SePayUserProfile
	sepayAccounts           map[domain.ID]*domain.SePayBankAccount
	sepayAccountXIDs        map[string]domain.ID
	bankAccountMaps         map[domain.ID]*domain.BankAccountMapping
	sepayLinkSessions       map[string]sepayLinkSession
	unmappedSePay           map[string]*domain.SePayUnmappedEvent
	suggestions             map[domain.ID]*domain.TransactionSuggestion
	feedback                map[domain.ID]*domain.ClassificationFeedback
	// idempotency keeps track of processed idempotency keys in-memory.
	idempotencyKeys map[string]time.Time
}

type emailVerificationToken struct {
	UserID    domain.ID
	ExpiresAt time.Time
	UsedAt    *time.Time
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		users:                   map[domain.ID]*domain.User{},
		emailVerificationTokens: map[string]emailVerificationToken{},
		userSettings:            map[domain.ID]*domain.UserSettings{},
		portfolios:              map[domain.ID]*domain.Portfolio{},
		accounts:                map[domain.ID]*domain.Account{},
		botAccountKeys:          map[domain.ID]*domain.BotAccountKey{},
		transactions:            map[domain.ID]*domain.Transaction{},
		transfers:               map[domain.ID]*domain.Transfer{},
		customers:               map[domain.ID]*domain.Customer{},
		loans:                   map[domain.ID]*domain.Loan{},
		loanPayments:            map[domain.ID]*domain.LoanPayment{},
		properties:              map[domain.ID]*domain.Property{},
		propertyValues:          map[domain.ID]*domain.PropertyValuation{},
		assets:                  map[domain.ID]*domain.Asset{},
		assetValues:             map[domain.ID]*domain.AssetValuation{},
		budgets:                 map[domain.ID]*domain.Budget{},
		budgetAllocs:            map[domain.ID]*domain.BudgetAllocation{},
		forecast:                map[domain.ID]*domain.ForecastScenario{},
		bankConnections:         map[domain.ID]*domain.BankConnection{},
		bankFeed:                map[domain.ID]*domain.BankFeedTransaction{},
		bankFeedKeys:            map[string]domain.ID{},
		bankFeedEvents:          map[domain.ID]*domain.BankFeedEvent{},
		bankFeedEventKeys:       map[string]domain.ID{},
		bankRecon:               map[domain.ID]*domain.BankReconciliation{},
		automationRules:         map[domain.ID]*domain.AutomationRule{},
		bankPaymentReqs:         map[domain.ID]*domain.BankPaymentRequest{},
		assistantCmds:           map[domain.ID]*domain.AssistantCommand{},
		auditLogs:               map[domain.ID]*domain.AuditLog{},
		idempotencyKeys:         map[string]time.Time{},
		sepayProfiles:           map[domain.ID]*domain.SePayUserProfile{},
		sepayAccounts:           map[domain.ID]*domain.SePayBankAccount{},
		sepayAccountXIDs:        map[string]domain.ID{},
		bankAccountMaps:         map[domain.ID]*domain.BankAccountMapping{},
		sepayLinkSessions:       map[string]sepayLinkSession{},
		unmappedSePay:           map[string]*domain.SePayUnmappedEvent{},
		suggestions:             map[domain.ID]*domain.TransactionSuggestion{},
		feedback:                map[domain.ID]*domain.ClassificationFeedback{},
	}
}

type sepayLinkSession struct {
	userID    domain.ID
	expiresAt time.Time
	completed bool
}

func newID() domain.ID {
	return domain.ID(uuid.NewString())
}

func now() time.Time {
	return time.Now().UTC()
}

func normalizeCustomerName(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

func isValidTransactionStatus(status domain.TransactionStatus) bool {
	switch status {
	case "", domain.TransactionStatusDraft, domain.TransactionStatusPending, domain.TransactionStatusPosted, domain.TransactionStatusReconciled, domain.TransactionStatusVoided:
		return true
	default:
		return false
	}
}

func parseAmountString(value string) (float64, error) {
	var f float64
	var err error
	f, err = strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, errors.New("invalid amount")
	}
	return f, nil
}

func (s *InMemoryStore) SeedDemoUser(email, name string, password string) domain.ID {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := domain.ID("demo-user")
	if email != "demo@wealthos.vn" && email != "" {
		id = newID()
	}
	verifiedAt := now()
	s.users[id] = &domain.User{
		Timestamped: domain.Timestamped{
			ID:        id,
			CreatedAt: now(),
			UpdatedAt: now(),
		},
		Email:           email,
		Name:            name,
		Password:        password,
		EmailVerifiedAt: &verifiedAt,
	}
	return id
}

func (s *InMemoryStore) CreateUser(input domain.User) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" || strings.TrimSpace(input.Name) == "" || input.Password == "" {
		return domain.User{}, errors.New("email, password, name are required")
	}
	for _, existing := range s.users {
		if strings.EqualFold(strings.TrimSpace(existing.Email), email) {
			return domain.User{}, errors.New("email already exists")
		}
	}
	createdAt := now()
	input.Timestamped = domain.Timestamped{ID: newID(), CreatedAt: createdAt, UpdatedAt: createdAt}
	input.Email = email
	input.EmailVerifiedAt = nil
	s.users[input.ID] = &input
	return input, nil
}

func (s *InMemoryStore) CreateAuditLog(input domain.AuditLog) (domain.AuditLog, error) {
	if strings.TrimSpace(string(input.UserID)) == "" {
		return domain.AuditLog{}, errors.New("userId is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if input.ID == "" {
		input.ID = newID()
	}
	input.CreatedAt = now()
	input.UpdatedAt = input.CreatedAt
	if strings.TrimSpace(input.Action) == "" {
		input.Action = "unknown"
	}
	s.auditLogs[input.ID] = &input
	return input, nil
}

func (s *InMemoryStore) ListAuditLogs(userID domain.ID) []domain.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.AuditLog, 0)
	for _, a := range s.auditLogs {
		if a.UserID != userID {
			continue
		}
		out = append(out, *a)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *InMemoryStore) GetUser(id domain.ID) (*domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, false
	}
	cp := *u
	return &cp, true
}

func (s *InMemoryStore) GetUserSettings(userID domain.ID) (*domain.UserSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.userSettings[userID]
	if !ok {
		return &domain.UserSettings{
			UserID:            userID,
			AmountDisplayMode: domain.AmountDisplayModeFull,
			UpdatedAt:         time.Now(),
		}, nil
	}
	cp := *st
	return &cp, nil
}

func (s *InMemoryStore) UpsertUserSettings(input domain.UserSettings) (*domain.UserSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if input.AmountDisplayMode != domain.AmountDisplayModeCompact && input.AmountDisplayMode != domain.AmountDisplayModeFull {
		input.AmountDisplayMode = domain.AmountDisplayModeFull
	}
	input.UpdatedAt = time.Now()
	s.userSettings[input.UserID] = &input
	cp := input
	return &cp, nil
}

func (s *InMemoryStore) GetUserByEmail(email string) (*domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	if cleanEmail == "" {
		return nil, false
	}
	prefix := cleanEmail
	if idx := strings.Index(cleanEmail, "@"); idx != -1 {
		prefix = cleanEmail[:idx]
	}

	for _, u := range s.users {
		uEmail := strings.ToLower(strings.TrimSpace(u.Email))
		uName := strings.ToLower(strings.TrimSpace(u.Name))
		uPrefix := uEmail
		if idx := strings.Index(uEmail, "@"); idx != -1 {
			uPrefix = uEmail[:idx]
		}

		if uEmail == cleanEmail || uName == cleanEmail || (prefix != "" && prefix == uPrefix) {
			cp := *u
			return &cp, true
		}
	}
	return nil, false
}

func (s *InMemoryStore) CreateEmailVerificationToken(userID domain.ID, tokenHash string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[userID]; !ok {
		return errors.New("user not found")
	}
	for hash, token := range s.emailVerificationTokens {
		if token.UserID == userID && token.UsedAt == nil {
			delete(s.emailVerificationTokens, hash)
		}
	}
	s.emailVerificationTokens[tokenHash] = emailVerificationToken{UserID: userID, ExpiresAt: expiresAt}
	return nil
}

func (s *InMemoryStore) VerifyEmail(email, tokenHash string, at time.Time) (*domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.emailVerificationTokens[tokenHash]
	if !ok || token.UsedAt != nil || !at.Before(token.ExpiresAt) {
		return nil, errors.New("verification code is invalid or expired")
	}
	user, ok := s.users[token.UserID]
	if !ok || !strings.EqualFold(strings.TrimSpace(user.Email), strings.TrimSpace(email)) {
		return nil, errors.New("verification code is invalid or expired")
	}
	verifiedAt := at.UTC()
	user.EmailVerifiedAt = &verifiedAt
	user.UpdatedAt = verifiedAt
	token.UsedAt = &verifiedAt
	s.emailVerificationTokens[tokenHash] = token
	copy := *user
	return &copy, nil
}

func (s *InMemoryStore) GetPortfolio(id domain.ID) (*domain.Portfolio, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.portfolios[id]
	if !ok {
		return nil, false
	}
	cp := *p
	return &cp, true
}

func (s *InMemoryStore) GetAccount(id domain.ID) (*domain.Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.accounts[id]
	if !ok {
		return nil, false
	}
	cp := *a
	return &cp, true
}

func (s *InMemoryStore) GetLoan(id domain.ID) (*domain.Loan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.loans[id]
	if !ok {
		return nil, false
	}
	cp := *l
	return &cp, true
}

func (s *InMemoryStore) UpdateLoan(id domain.ID, mutate func(*domain.Loan)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.loans[id]
	if !ok {
		return false
	}
	mutate(l)
	l.UpdatedAt = now()
	return true
}

func (s *InMemoryStore) GetBankConnection(id domain.ID) (*domain.BankConnection, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.bankConnections[id]
	if !ok {
		return nil, false
	}
	cp := *c
	return &cp, true
}

func (s *InMemoryStore) GetBankConnectionByCallbackState(callbackState string) (*domain.BankConnection, bool) {
	state := strings.TrimSpace(callbackState)
	if state == "" {
		return nil, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.bankConnections {
		if strings.TrimSpace(c.CallbackState) == state {
			cp := *c
			return &cp, true
		}
	}
	return nil, false
}

func (s *InMemoryStore) UpdateBankConnection(id domain.ID, mutate func(*domain.BankConnection)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.bankConnections[id]
	if !ok {
		return false
	}
	mutate(c)
	c.UpdatedAt = now()
	return true
}

func (s *InMemoryStore) GetBankFeed(id domain.ID) (*domain.BankFeedTransaction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.bankFeed[id]
	if !ok {
		return nil, false
	}
	cp := *item
	return &cp, true
}

func (s *InMemoryStore) CreateTransactionStrict(input domain.Transaction) (domain.Transaction, error) {
	if input.UserID == "" || input.AccountID == "" || input.Currency == "" {
		return domain.Transaction{}, errors.New("userId, accountId and currency are required")
	}
	if input.Amount == "" {
		return domain.Transaction{}, errors.New("amount is required")
	}
	amt, err := parseAmountString(string(input.Amount))
	if err != nil {
		return domain.Transaction{}, errors.New("invalid amount")
	}
	if amt <= 0 {
		return domain.Transaction{}, errors.New("amount must be greater than 0")
	}
	if input.Type == "" {
		return domain.Transaction{}, errors.New("type is required")
	}
	acc, ok := s.accounts[input.AccountID]
	if !ok {
		return domain.Transaction{}, errors.New("accountId does not exist")
	}
	if acc.UserID != input.UserID {
		return domain.Transaction{}, errors.New("account does not belong to user")
	}
	if !isValidTransactionStatus(input.Status) {
		return domain.Transaction{}, errors.New("invalid transaction status")
	}
	return s.CreateTransaction(input)
}

func (s *InMemoryStore) RecordIdempotency(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.idempotencyKeys[key]; exists {
		return false
	}
	s.idempotencyKeys[key] = now()
	return true
}

func (s *InMemoryStore) ClearIdempotencyOlderThan(cutoff time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for k, ts := range s.idempotencyKeys {
		if ts.Before(cutoff) {
			delete(s.idempotencyKeys, k)
			removed++
		}
	}
	return removed
}

func (s *InMemoryStore) GetUserByID(id domain.ID) (*domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, false
	}
	cp := *u
	return &cp, true
}

func (s *InMemoryStore) EnsureUserPortfolio(_ string, baseCurrency string, userID domain.ID) (*domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	for _, portfolio := range s.portfolios {
		if portfolio.UserID == userID {
			copy := *user
			return &copy, nil
		}
	}

	// Auto-seed one default portfolio for onboarding workflow.
	pID := newID()
	s.portfolios[pID] = &domain.Portfolio{
		Timestamped: domain.Timestamped{
			ID:        pID,
			CreatedAt: now(),
			UpdatedAt: now(),
		},
		UserID:       userID,
		Name:         "Default",
		BaseCurrency: baseCurrency,
	}
	copy := *user
	return &copy, nil
}

func (s *InMemoryStore) CreateAccount(input domain.Account) (domain.Account, error) {
	if input.UserID == "" || input.Name == "" {
		return domain.Account{}, errors.New("userId and name are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	input.ID = id
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.accounts[id] = &input
	return input, nil
}

func (s *InMemoryStore) UpsertBotAccountKey(input domain.BotAccountKey) (domain.BotAccountKey, error) {
	if input.AccountID == "" || strings.TrimSpace(input.SecretHash) == "" {
		return domain.BotAccountKey{}, errors.New("accountId and secret hash are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.botAccountKeys {
		if existing.AccountID == input.AccountID {
			input.ID, input.CreatedAt = existing.ID, existing.CreatedAt
		}
	}
	if input.ID == "" {
		input.ID, input.CreatedAt = newID(), now()
	}
	input.UpdatedAt = now()
	copy := input
	s.botAccountKeys[input.ID] = &copy
	return input, nil
}

func (s *InMemoryStore) GetActiveBotAccountKey(accountID domain.ID) (*domain.BotAccountKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, key := range s.botAccountKeys {
		if key.AccountID == accountID && key.RevokedAt.IsZero() {
			copy := *key
			return &copy, true
		}
	}
	return nil, false
}

func (s *InMemoryStore) CreatePortfolio(input domain.Portfolio) (domain.Portfolio, error) {
	if input.UserID == "" || input.Name == "" {
		return domain.Portfolio{}, errors.New("userId and name are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	input.ID = newID()
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.portfolios[input.ID] = &input
	return input, nil
}

func (s *InMemoryStore) ListPortfolios(userID domain.ID) []domain.Portfolio {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Portfolio{}
	for _, p := range s.portfolios {
		if p.UserID == userID {
			out = append(out, *p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *InMemoryStore) FirstPortfolio(userID domain.ID) (domain.Portfolio, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.portfolios {
		if p.UserID == userID {
			cp := *p
			return cp, true
		}
	}
	return domain.Portfolio{}, false
}

func (s *InMemoryStore) ListAccounts(userID domain.ID) []domain.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Account{}
	for _, a := range s.accounts {
		if a.UserID == userID {
			out = append(out, *a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *InMemoryStore) DeleteAccount(userID domain.ID, id domain.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.accounts[id]
	if !ok || a.UserID != userID {
		return errors.New("account not found")
	}
	for _, tx := range s.transactions {
		if tx.AccountID == id {
			return errors.New("cannot delete account with transaction history")
		}
	}
	delete(s.accounts, id)
	return nil
}

func (s *InMemoryStore) DeletePortfolio(userID domain.ID, id domain.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.portfolios[id]
	if !ok || p.UserID != userID {
		return errors.New("portfolio not found")
	}
	for _, account := range s.accounts {
		if account.PortfolioID == id {
			return errors.New("cannot delete portfolio with financial history")
		}
	}
	for _, loan := range s.loans {
		if loan.PortfolioID == id {
			return errors.New("cannot delete portfolio with financial history")
		}
	}
	for _, property := range s.properties {
		if property.PortfolioID == id {
			return errors.New("cannot delete portfolio with financial history")
		}
	}
	for _, asset := range s.assets {
		if asset.PortfolioID == id {
			return errors.New("cannot delete portfolio with financial history")
		}
	}
	for _, tx := range s.transactions {
		if tx.PortfolioID == id {
			return errors.New("cannot delete portfolio with financial history")
		}
	}
	delete(s.portfolios, id)
	return nil
}

func (s *InMemoryStore) DeleteLoan(userID domain.ID, id domain.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.loans[id]
	if !ok || l.UserID != userID {
		return errors.New("loan not found")
	}
	for _, payment := range s.loanPayments {
		if payment.LoanID == id {
			return errors.New("cannot delete loan with payment history")
		}
	}
	for transactionID, transaction := range s.transactions {
		if transaction.UserID == userID && transaction.Type == domain.TransactionTypeLoanDisbursement && transaction.Source == "loan_disbursement" && transaction.Note == "loan principal: "+string(id) {
			delete(s.transactions, transactionID)
		}
	}
	delete(s.loans, id)
	return nil
}

func (s *InMemoryStore) DeleteProperty(userID domain.ID, id domain.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	pr, ok := s.properties[id]
	if !ok || pr.UserID != userID {
		return errors.New("property not found")
	}
	for _, value := range s.propertyValues {
		if value.PropertyID == id {
			return errors.New("cannot delete property with valuation history")
		}
	}
	delete(s.properties, id)
	return nil
}

func (s *InMemoryStore) DeleteAsset(userID domain.ID, id domain.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ast, ok := s.assets[id]
	if !ok || ast.UserID != userID {
		return errors.New("asset not found")
	}
	for _, value := range s.assetValues {
		if value.AssetID == id {
			return errors.New("cannot delete asset with valuation history")
		}
	}
	delete(s.assets, id)
	return nil
}

func (s *InMemoryStore) CreateTransaction(input domain.Transaction) (domain.Transaction, error) {
	if input.UserID == "" || input.AccountID == "" || input.Amount == "" {
		return domain.Transaction{}, errors.New("userId, accountId and amount are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if input.Status == "" {
		input.Status = domain.TransactionStatusPosted
	}
	id := newID()
	input.ID = id
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.transactions[id] = &input
	return input, nil
}

func (s *InMemoryStore) GetTransaction(id domain.ID) (*domain.Transaction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.transactions[id]
	if !ok {
		return nil, false
	}
	cp := *t
	return &cp, true
}

func (s *InMemoryStore) ListTransactions(userID domain.ID, accountID domain.ID) []domain.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Transaction{}
	for _, t := range s.transactions {
		if t.UserID != userID {
			continue
		}
		if accountID != "" && t.AccountID != accountID {
			continue
		}
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].OccurredAt.Equal(out[j].OccurredAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].OccurredAt.After(out[j].OccurredAt)
	})
	return out
}

func (s *InMemoryStore) CreateTransfer(input domain.Transfer) (domain.Transfer, error) {
	if input.UserID == "" || input.FromAccountID == "" || input.ToAccountID == "" || input.Amount == "" {
		return domain.Transfer{}, errors.New("missing required transfer fields")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	input.ID = id
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.transfers[id] = &input
	for _, entry := range []domain.Transaction{
		{UserID: input.UserID, AccountID: input.FromAccountID, Type: domain.TransactionTypeTransfer, Amount: input.Amount, Currency: input.Currency, Note: "internal transfer - out: " + input.Note, OccurredAt: input.OccurredAt, Status: domain.TransactionStatusPosted, Source: "transfer"},
		{UserID: input.UserID, AccountID: input.ToAccountID, Type: domain.TransactionTypeTransfer, Amount: input.Amount, Currency: input.Currency, Note: "internal transfer - in: " + input.Note, OccurredAt: input.OccurredAt, Status: domain.TransactionStatusPosted, Source: "transfer"},
	} {
		entry.ID, entry.CreatedAt, entry.UpdatedAt = newID(), now(), now()
		s.transactions[entry.ID] = &entry
	}
	return input, nil
}

func (s *InMemoryStore) CreateCustomer(input domain.Customer) (domain.Customer, error) {
	if input.UserID == "" || strings.TrimSpace(input.Name) == "" {
		return domain.Customer{}, errors.New("customer name is required")
	}
	input.Name = strings.TrimSpace(input.Name)
	input.NormalizedName = normalizeCustomerName(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.customers {
		if existing.UserID == input.UserID && existing.NormalizedName == input.NormalizedName {
			if input.Phone != "" {
				existing.Phone = input.Phone
				existing.UpdatedAt = now()
			}
			return *existing, nil
		}
	}
	input.ID, input.CreatedAt, input.UpdatedAt = newID(), now(), now()
	s.customers[input.ID] = &input
	return input, nil
}

func (s *InMemoryStore) ListCustomers(userID domain.ID, query string, limit int) []domain.Customer {
	needle := normalizeCustomerName(query)
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Customer, 0)
	for _, customer := range s.customers {
		if customer.UserID != userID {
			continue
		}
		if needle != "" && !strings.Contains(customer.NormalizedName, needle) && !strings.Contains(customer.Phone, strings.TrimSpace(query)) {
			continue
		}
		out = append(out, *customer)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (s *InMemoryStore) GetCustomer(id domain.ID) (*domain.Customer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	customer, ok := s.customers[id]
	if !ok {
		return nil, false
	}
	copy := *customer
	return &copy, true
}

func (s *InMemoryStore) CreateLoan(input domain.Loan) (domain.Loan, error) {
	if input.UserID == "" || input.PrincipalInitial == "" || input.AnnualRate == "" {
		return domain.Loan{}, errors.New("missing required loan fields")
	}
	if input.Status == "" {
		input.Status = domain.LoanStatusDraft
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	input.ID = id
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.loans[id] = &input
	return input, nil
}

func (s *InMemoryStore) ListLoans(userID domain.ID) []domain.Loan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Loan{}
	for _, l := range s.loans {
		if l.UserID == userID {
			out = append(out, *l)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *InMemoryStore) ListLoanPayments(userID domain.ID, loanID domain.ID) []domain.LoanPayment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.LoanPayment{}
	for _, p := range s.loanPayments {
		if p.UserID != userID {
			continue
		}
		if loanID != "" && p.LoanID != loanID {
			continue
		}
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OccurredAt.After(out[j].OccurredAt) })
	return out
}

func (s *InMemoryStore) CreateLoanPayment(input domain.LoanPayment) (domain.LoanPayment, error) {
	if input.UserID == "" || input.LoanID == "" {
		return domain.LoanPayment{}, errors.New("missing required loan payment fields")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	input.ID = id
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.loanPayments[id] = &input
	return input, nil
}

func (s *InMemoryStore) SettleLoanPayment(loanID domain.ID, expectedPrincipalBalance, nextPrincipalBalance string, nextStatus domain.LoanStatus, payment domain.LoanPayment, ledger domain.Transaction) (domain.LoanPayment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	loan, ok := s.loans[loanID]
	if !ok || loan.PrincipalBalance != expectedPrincipalBalance {
		return domain.LoanPayment{}, errors.New("loan balance changed; retry payment")
	}
	if payment.UserID == "" || payment.AccountID == "" || ledger.UserID == "" || ledger.AccountID == "" {
		return domain.LoanPayment{}, errors.New("missing settlement data")
	}
	ledger.ID, ledger.CreatedAt, ledger.UpdatedAt = newID(), now(), now()
	s.transactions[ledger.ID] = &ledger
	payment.ID, payment.TransactionID = newID(), ledger.ID
	payment.CreatedAt, payment.UpdatedAt = now(), now()
	s.loanPayments[payment.ID] = &payment
	loan.PrincipalBalance, loan.Status, loan.UpdatedAt = nextPrincipalBalance, nextStatus, now()
	return payment, nil
}

func (s *InMemoryStore) CreateProperty(input domain.Property) (domain.Property, error) {
	if input.UserID == "" || input.Name == "" {
		return domain.Property{}, errors.New("missing required property fields")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	input.ID = id
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.properties[id] = &input
	return input, nil
}

func (s *InMemoryStore) ListProperties(userID domain.ID) []domain.Property {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Property{}
	for _, p := range s.properties {
		if p.UserID == userID {
			out = append(out, *p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *InMemoryStore) GetProperty(id domain.ID) (*domain.Property, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.properties[id]
	if !ok {
		return nil, false
	}
	cp := *p
	return &cp, true
}

func (s *InMemoryStore) AddPropertyValuation(v domain.PropertyValuation) (domain.PropertyValuation, error) {
	if v.PropertyID == "" || v.Amount == "" {
		return domain.PropertyValuation{}, errors.New("propertyId and valuationAmount are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	v.ID = id
	v.CreatedAt = now()
	v.UpdatedAt = now()
	s.propertyValues[id] = &v
	return v, nil
}

func (s *InMemoryStore) ListPropertyValues(userID domain.ID) []domain.PropertyValuation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.PropertyValuation{}
	for _, value := range s.propertyValues {
		p, ok := s.properties[value.PropertyID]
		if !ok || p.UserID != userID {
			continue
		}
		out = append(out, *value)
	}
	return out
}

func (s *InMemoryStore) CreateAsset(input domain.Asset) (domain.Asset, error) {
	if input.UserID == "" || input.Name == "" {
		return domain.Asset{}, errors.New("missing required asset fields")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	input.ID = id
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.assets[id] = &input
	return input, nil
}

func (s *InMemoryStore) ListAssets(userID domain.ID) []domain.Asset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Asset{}
	for _, a := range s.assets {
		if a.UserID == userID {
			out = append(out, *a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *InMemoryStore) AddAssetValuation(v domain.AssetValuation) (domain.AssetValuation, error) {
	if v.AssetID == "" || v.Amount == "" {
		return domain.AssetValuation{}, errors.New("assetId and valuationAmount are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	v.ID = id
	v.CreatedAt = now()
	v.UpdatedAt = now()
	s.assetValues[id] = &v
	return v, nil
}

func (s *InMemoryStore) ListAssetValues(userID domain.ID) []domain.AssetValuation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AssetValuation{}
	for _, value := range s.assetValues {
		a, ok := s.assets[value.AssetID]
		if !ok || a.UserID != userID {
			continue
		}
		out = append(out, *value)
	}
	return out
}

func (s *InMemoryStore) GetAsset(id domain.ID) (*domain.Asset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.assets[id]
	if !ok {
		return nil, false
	}
	cp := *a
	return &cp, true
}

func (s *InMemoryStore) CreateBudget(input domain.Budget) (domain.Budget, error) {
	if input.UserID == "" || input.Period == "" {
		return domain.Budget{}, errors.New("missing required budget fields")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	input.ID = id
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.budgets[id] = &input
	return input, nil
}

func (s *InMemoryStore) UpsertBudget(input domain.Budget) (domain.Budget, error) {
	if input.UserID == "" || input.Period == "" {
		return domain.Budget{}, errors.New("missing required budget fields")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, b := range s.budgets {
		if b.UserID != input.UserID {
			continue
		}
		if b.Period != input.Period {
			continue
		}
		if b.CategoryID != input.CategoryID {
			continue
		}
		b.Limit = firstNonEmptyBudgetLimit(input.Limit, b.Limit)
		b.UpdatedAt = now()
		return *b, nil
	}
	id := newID()
	input.ID = id
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.budgets[id] = &input
	return input, nil
}

func firstNonEmptyBudgetLimit(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return "0"
}

func (s *InMemoryStore) ListBudgets(userID domain.ID, period string) []domain.Budget {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Budget{}
	for _, b := range s.budgets {
		if b.UserID == userID && b.Period == period {
			out = append(out, *b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *InMemoryStore) UpsertBudgetAllocs(input domain.BudgetAllocation) (domain.BudgetAllocation, error) {
	if input.BudgetID == "" {
		return domain.BudgetAllocation{}, errors.New("budgetId is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	amountSpent := input.AmountSpent
	if amountSpent == "" {
		amountSpent = "0"
	}
	if !strings.EqualFold(strings.TrimSpace(input.Currency), "VND") && strings.TrimSpace(input.Currency) != "" {
		input.Currency = strings.TrimSpace(input.Currency)
	} else {
		input.Currency = "VND"
	}
	for _, alloc := range s.budgetAllocs {
		if alloc.BudgetID != input.BudgetID {
			continue
		}
		alloc.AmountSpent = amountSpent
		alloc.Currency = input.Currency
		alloc.UpdatedAt = now()
		cp := *alloc
		return cp, nil
	}
	input.ID = newID()
	input.CreatedAt = now()
	input.UpdatedAt = now()
	input.AmountSpent = amountSpent
	s.budgetAllocs[input.ID] = &input
	return input, nil
}

func (s *InMemoryStore) CreateForecastScenario(input domain.ForecastScenario) (domain.ForecastScenario, error) {
	if input.UserID == "" || input.Name == "" {
		return domain.ForecastScenario{}, errors.New("missing required fields")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	input.ID = id
	input.CreatedAt = now()
	input.UpdatedAt = now()
	input.Status = "draft"
	s.forecast[id] = &input
	return input, nil
}

func (s *InMemoryStore) ListForecastScenarios(userID domain.ID) []domain.ForecastScenario {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.ForecastScenario{}
	for _, f := range s.forecast {
		if f.UserID == userID {
			out = append(out, *f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *InMemoryStore) ListForecastScenariosByStatus(status string) []domain.ForecastScenario {
	s.mu.RLock()
	defer s.mu.RUnlock()
	normalizedStatus := strings.TrimSpace(strings.ToLower(status))
	out := []domain.ForecastScenario{}
	for _, f := range s.forecast {
		if normalizedStatus == "" || strings.EqualFold(string(f.Status), normalizedStatus) {
			out = append(out, *f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func (s *InMemoryStore) RunForecastScenario(id domain.ID, assumptions string) (domain.ForecastScenario, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.forecast[id]
	if !ok {
		return domain.ForecastScenario{}, fmt.Errorf("scenario not found")
	}
	if strings.TrimSpace(assumptions) == "" {
		assumptions = "{}"
	}
	f.Status = "running"
	f.Assumptions = assumptions
	f.UpdatedAt = now()
	return *f, nil
}

func (s *InMemoryStore) FinalizeForecastScenario(id domain.ID, status string, result string) (domain.ForecastScenario, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.forecast[id]
	if !ok {
		return domain.ForecastScenario{}, fmt.Errorf("scenario not found")
	}
	if strings.TrimSpace(status) == "" {
		status = "done"
	}
	f.Status = status
	f.Result = strings.TrimSpace(result)
	f.UpdatedAt = now()
	return *f, nil
}

func (s *InMemoryStore) CreateBankConnection(input domain.BankConnection) (domain.BankConnection, error) {
	if input.UserID == "" {
		return domain.BankConnection{}, errors.New("userId is required")
	}
	callbackState := strings.TrimSpace(input.CallbackState)
	if callbackState == "" {
		callbackState = "not_called"
	}
	syncStatus := strings.TrimSpace(input.SyncStatus)
	if syncStatus == "" {
		syncStatus = "idle"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	input.ID = newID()
	input.CreatedAt = now()
	input.UpdatedAt = now()
	input.Status = "connected"
	input.SyncStatus = syncStatus
	input.CallbackState = callbackState
	s.bankConnections[input.ID] = &input
	return input, nil
}

func (s *InMemoryStore) ListBankConnections(userID domain.ID) []domain.BankConnection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankConnection{}
	for _, c := range s.bankConnections {
		if c.UserID == userID {
			out = append(out, *c)
		}
	}
	return out
}

func (s *InMemoryStore) UpsertSePayUserProfile(input domain.SePayUserProfile) (domain.SePayUserProfile, error) {
	if input.UserID == "" || input.CompanyXID == "" {
		return domain.SePayUserProfile{}, errors.New("userId and companyXid are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.sepayProfiles[input.UserID]
	if ok {
		input.CreatedAt = current.CreatedAt
	} else {
		input.CreatedAt = now()
	}
	input.UpdatedAt = now()
	copy := input
	s.sepayProfiles[input.UserID] = &copy
	return copy, nil
}

func (s *InMemoryStore) GetSePayUserProfile(userID domain.ID) (*domain.SePayUserProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.sepayProfiles[userID]
	if !ok {
		return nil, false
	}
	copy := *item
	return &copy, true
}

func (s *InMemoryStore) UpsertSePayBankAccount(input domain.SePayBankAccount) (domain.SePayBankAccount, error) {
	if input.UserID == "" || input.BankAccountXID == "" {
		return domain.SePayBankAccount{}, errors.New("userId and bankAccountXid are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if existingID, ok := s.sepayAccountXIDs[input.BankAccountXID]; ok {
		existing := s.sepayAccounts[existingID]
		if existing.UserID != input.UserID {
			return domain.SePayBankAccount{}, errors.New("bank account belongs to another user")
		}
		input.ID, input.CreatedAt = existing.ID, existing.CreatedAt
	} else {
		input.ID, input.CreatedAt = newID(), now()
	}
	input.UpdatedAt = now()
	if input.Status == "" {
		input.Status = "linked"
	}
	copy := input
	s.sepayAccounts[input.ID] = &copy
	s.sepayAccountXIDs[input.BankAccountXID] = input.ID
	return copy, nil
}

func (s *InMemoryStore) ListSePayBankAccounts(userID domain.ID) []domain.SePayBankAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.SePayBankAccount{}
	for _, item := range s.sepayAccounts {
		if item.UserID == userID {
			out = append(out, *item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (s *InMemoryStore) GetSePayBankAccount(id domain.ID) (*domain.SePayBankAccount, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.sepayAccounts[id]
	if !ok {
		return nil, false
	}
	copy := *item
	return &copy, true
}

func (s *InMemoryStore) GetSePayBankAccountByXID(xid string) (*domain.SePayBankAccount, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.sepayAccountXIDs[xid]
	if !ok {
		return nil, false
	}
	item := s.sepayAccounts[id]
	copy := *item
	return &copy, true
}

func (s *InMemoryStore) SetSePayBankAccountStatus(id domain.ID, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.sepayAccounts[id]
	if !ok {
		return false
	}
	item.Status = status
	item.UpdatedAt = now()
	return true
}

func (s *InMemoryStore) UpsertBankAccountMapping(input domain.BankAccountMapping) (domain.BankAccountMapping, error) {
	if input.SePayBankAccountID == "" || input.UserID == "" || input.AccountID == "" {
		return domain.BankAccountMapping{}, errors.New("bank account mapping fields are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.bankAccountMaps[input.SePayBankAccountID]; ok {
		input.ID, input.CreatedAt = old.ID, old.CreatedAt
	} else {
		input.ID, input.CreatedAt = newID(), now()
	}
	if input.Status == "" {
		input.Status = "active"
	}
	input.UpdatedAt = now()
	copy := input
	s.bankAccountMaps[input.SePayBankAccountID] = &copy
	return copy, nil
}

func (s *InMemoryStore) GetBankAccountMapping(accountID domain.ID) (*domain.BankAccountMapping, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.bankAccountMaps[accountID]
	if !ok {
		return nil, false
	}
	copy := *item
	return &copy, true
}

func (s *InMemoryStore) DeactivateBankAccountMapping(accountID domain.ID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.bankAccountMaps[accountID]
	if !ok {
		return false
	}
	item.Status = "inactive"
	item.UpdatedAt = now()
	return true
}

func (s *InMemoryStore) CreateSePayLinkSession(xid string, userID domain.ID, expiresAt time.Time) error {
	if xid == "" || userID == "" {
		return errors.New("xid and userId are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sepayLinkSessions[xid] = sepayLinkSession{userID: userID, expiresAt: expiresAt}
	return nil
}

func (s *InMemoryStore) GetSePayLinkSessionUser(xid string) (domain.ID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sepayLinkSessions[xid]
	if !ok || (!session.expiresAt.IsZero() && session.expiresAt.Before(now())) {
		return "", false
	}
	return session.userID, true
}

func (s *InMemoryStore) CompleteSePayLinkSession(xid string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sepayLinkSessions[xid]
	if !ok {
		return false
	}
	session.completed = true
	s.sepayLinkSessions[xid] = session
	return true
}

func (s *InMemoryStore) QuarantineSePayEvent(input domain.SePayUnmappedEvent) (domain.SePayUnmappedEvent, error) {
	if input.Provider == "" || input.BankAccountXID == "" || input.TransactionID == "" {
		return domain.SePayUnmappedEvent{}, errors.New("provider, bankAccountXid and transactionId are required")
	}
	key := input.Provider + "::" + input.BankAccountXID + "::" + input.TransactionID
	s.mu.Lock()
	defer s.mu.Unlock()
	if current, ok := s.unmappedSePay[key]; ok {
		copy := *current
		return copy, nil
	}
	input.ID, input.CreatedAt, input.UpdatedAt = newID(), now(), now()
	if input.Status == "" {
		input.Status = "quarantined"
	}
	copy := input
	s.unmappedSePay[key] = &copy
	return copy, nil
}

func (s *InMemoryStore) CreateTransactionSuggestion(input domain.TransactionSuggestion) (domain.TransactionSuggestion, error) {
	if input.BankFeedTransactionID == "" || input.Source == "" {
		return domain.TransactionSuggestion{}, errors.New("suggestion feed and source are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	input.ID, input.CreatedAt, input.UpdatedAt = newID(), now(), now()
	if input.Version == "" {
		input.Version = "v1"
	}
	copy := input
	s.suggestions[input.ID] = &copy
	return copy, nil
}

func (s *InMemoryStore) ListTransactionSuggestions(feedID domain.ID) []domain.TransactionSuggestion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.TransactionSuggestion{}
	for _, item := range s.suggestions {
		if item.BankFeedTransactionID == feedID {
			out = append(out, *item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (s *InMemoryStore) CreateClassificationFeedback(input domain.ClassificationFeedback) (domain.ClassificationFeedback, error) {
	if input.BankFeedTransactionID == "" || input.UserID == "" || input.Action == "" {
		return domain.ClassificationFeedback{}, errors.New("feedback feed, user and action are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	input.ID, input.CreatedAt, input.UpdatedAt = newID(), now(), now()
	copy := input
	s.feedback[input.ID] = &copy
	return copy, nil
}

func (s *InMemoryStore) ListClassificationFeedback(userID domain.ID) []domain.ClassificationFeedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.ClassificationFeedback{}
	for _, item := range s.feedback {
		if item.UserID == userID {
			out = append(out, *item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *InMemoryStore) EnqueueBankFeedEvent(input domain.BankFeedEvent) (domain.BankFeedEvent, error) {
	if input.UserID == "" || input.ConnectionID == "" {
		return domain.BankFeedEvent{}, errors.New("userId and connectionId are required")
	}
	if input.Provider == "" {
		return domain.BankFeedEvent{}, errors.New("provider is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if input.EventKey != "" {
		if existingID, ok := s.bankFeedEventKeys[input.EventKey]; ok {
			if existing, ok2 := s.bankFeedEvents[existingID]; ok2 {
				cp := *existing
				return cp, nil
			}
		}
	}
	input.ID = newID()
	input.CreatedAt = now()
	input.UpdatedAt = now()
	if input.State == "" {
		input.State = domain.BankFeedEventStateQueued
	}
	s.bankFeedEvents[input.ID] = &input
	if input.EventKey != "" {
		s.bankFeedEventKeys[input.EventKey] = input.ID
	}
	return input, nil
}

func (s *InMemoryStore) ListBankFeedEvents(userID domain.ID, state string) []domain.BankFeedEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankFeedEvent{}
	for _, e := range s.bankFeedEvents {
		if userID != "" && e.UserID != userID {
			continue
		}
		if state != "" && e.State != state {
			continue
		}
		out = append(out, *e)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *InMemoryStore) GetBankFeedEvent(id domain.ID) (*domain.BankFeedEvent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.bankFeedEvents[id]
	if !ok {
		return nil, false
	}
	cp := *e
	return &cp, true
}

func (s *InMemoryStore) ClaimBankFeedEvent(id domain.ID) (*domain.BankFeedEvent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	event, ok := s.bankFeedEvents[id]
	if !ok || event.State != domain.BankFeedEventStateQueued {
		return nil, false
	}
	event.State = domain.BankFeedEventStateRunning
	event.Attempts++
	event.LastError = ""
	event.UpdatedAt = now()
	copy := *event
	return &copy, true
}

func (s *InMemoryStore) UpdateBankFeedEvent(id domain.ID, mutate func(*domain.BankFeedEvent)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.bankFeedEvents[id]
	if !ok {
		return false
	}
	mutate(e)
	e.UpdatedAt = now()
	return true
}

func (s *InMemoryStore) IngestBankFeed(input domain.BankFeedTransaction) (domain.BankFeedTransaction, error) {
	if input.UserID == "" || input.ConnectionID == "" {
		return domain.BankFeedTransaction{}, errors.New("userId and connectionId are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if input.ExternalID != "" {
		key := string(input.ConnectionID) + "::" + input.ExternalID
		if existingID, ok := s.bankFeedKeys[key]; ok {
			if existing, ok2 := s.bankFeed[existingID]; ok2 {
				cp := *existing
				return cp, nil
			}
		}
	}
	input.ID = newID()
	input.CreatedAt = now()
	input.UpdatedAt = now()
	if input.PostingState == "" {
		input.PostingState = domain.PostingStateReview
	}
	s.bankFeed[input.ID] = &input
	if input.ExternalID != "" {
		s.bankFeedKeys[string(input.ConnectionID)+"::"+input.ExternalID] = input.ID
	}
	return input, nil
}

func (s *InMemoryStore) ListBankFeed(userID domain.ID) []domain.BankFeedTransaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankFeedTransaction{}
	for _, t := range s.bankFeed {
		if t.UserID == userID {
			out = append(out, *t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OccurredAt.After(out[j].OccurredAt) })
	return out
}

func (s *InMemoryStore) ListBankFeedByState(userID domain.ID, state domain.TransactionPostingState) []domain.BankFeedTransaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankFeedTransaction{}
	for _, t := range s.bankFeed {
		if t.PostingState != state {
			continue
		}
		if userID != "" && t.UserID != userID {
			continue
		}
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OccurredAt.After(out[j].OccurredAt) })
	return out
}

func (s *InMemoryStore) UpdateFeedState(id domain.ID, state domain.TransactionPostingState, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.bankFeed[id]
	if !ok {
		return errors.New("bank feed transaction not found")
	}
	t.PostingState = state
	if state == domain.PostingStateIgnored {
		t.ClassificationStatus = "ignored"
	}
	t.Evidence = reason
	t.UpdatedAt = now()
	return nil
}

func (s *InMemoryStore) UpdateFeed(id domain.ID, mutate func(*domain.BankFeedTransaction)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.bankFeed[id]
	if !ok {
		return false
	}
	mutate(t)
	t.UpdatedAt = now()
	return true
}

func (s *InMemoryStore) LinkBankFeedPosting(feedID domain.ID, txnID domain.ID) bool {
	return s.UpdateFeed(feedID, func(f *domain.BankFeedTransaction) {
		f.PostedTxnID = txnID
		f.PostingState = domain.PostingStatePosted
		f.ClassificationStatus = "confirmed"
	})
}

func (s *InMemoryStore) CreateAutomationRule(input domain.AutomationRule) (domain.AutomationRule, error) {
	if input.UserID == "" {
		return domain.AutomationRule{}, errors.New("userId is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	input.ID = newID()
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.automationRules[input.ID] = &input
	return input, nil
}

func (s *InMemoryStore) ListUserAutomationRules(userID domain.ID) []domain.AutomationRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AutomationRule{}
	for _, r := range s.automationRules {
		if r.UserID == userID {
			out = append(out, *r)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Priority == out[j].Priority {
			return out[i].CreatedAt.After(out[j].CreatedAt)
		}
		return out[i].Priority < out[j].Priority
	})
	return out
}

func (s *InMemoryStore) ListUserRules(userID domain.ID) []domain.AutomationRule {
	return s.ListUserAutomationRules(userID)
}

func (s *InMemoryStore) GetUserRules(userID domain.ID, accountID domain.ID, direction string) []domain.AutomationRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AutomationRule{}
	for _, r := range s.automationRules {
		if r.UserID != userID || !r.Enabled {
			continue
		}
		if direction != "" && r.Direction != "" && r.Direction != direction {
			continue
		}
		if accountID != "" && r.AccountID != "" && r.AccountID != accountID {
			continue
		}
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool {
		iIsAccountScoped := out[i].AccountID != "" && out[i].AccountID == accountID
		jIsAccountScoped := out[j].AccountID != "" && out[j].AccountID == accountID
		if iIsAccountScoped != jIsAccountScoped {
			return iIsAccountScoped
		}
		if out[i].Priority == out[j].Priority {
			if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
				return out[i].CreatedAt.After(out[j].CreatedAt)
			}
			return out[i].ID > out[j].ID
		}
		return out[i].Priority < out[j].Priority
	})
	return out
}

func (s *InMemoryStore) GetAutomationRule(id domain.ID) (*domain.AutomationRule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.automationRules[id]
	if !ok {
		return nil, false
	}
	cp := *r
	return &cp, true
}

func (s *InMemoryStore) UpdateAutomationRule(id domain.ID, mutate func(*domain.AutomationRule)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.automationRules[id]
	if !ok {
		return false
	}
	mutate(r)
	r.UpdatedAt = now()
	return true
}

func (s *InMemoryStore) DeleteAutomationRule(id domain.ID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.automationRules[id]; !ok {
		return false
	}
	delete(s.automationRules, id)
	return true
}

func (s *InMemoryStore) CreateBankPaymentRequest(input domain.BankPaymentRequest) (domain.BankPaymentRequest, error) {
	if input.UserID == "" || input.LoanID == "" || input.Code == "" {
		return domain.BankPaymentRequest{}, errors.New("userId, loanId and paymentCode are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	input.ID = newID()
	input.CreatedAt = now()
	input.UpdatedAt = now()
	if input.Status == "" {
		input.Status = "open"
	}
	s.bankPaymentReqs[input.ID] = &input
	return input, nil
}

func (s *InMemoryStore) GetBankPaymentRequestByCode(userID domain.ID, code string) (*domain.BankPaymentRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, req := range s.bankPaymentReqs {
		if req.UserID != userID || req.Code == "" {
			continue
		}
		if req.Code == code {
			cp := *req
			return &cp, true
		}
	}
	return nil, false
}

func (s *InMemoryStore) ListBankPaymentRequests(userID domain.ID) []domain.BankPaymentRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankPaymentRequest{}
	for _, req := range s.bankPaymentReqs {
		if req.UserID == userID {
			out = append(out, *req)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *InMemoryStore) RevokeBankConnection(id domain.ID) (*domain.BankConnection, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.bankConnections[id]
	if !ok {
		return nil, false
	}
	c.Status = "revoked"
	c.UpdatedAt = now()
	cp := *c
	return &cp, true
}

func (s *InMemoryStore) ListAutomationRules(userID domain.ID) []domain.AutomationRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AutomationRule{}
	for _, r := range s.automationRules {
		if r.UserID == userID {
			out = append(out, *r)
		}
	}
	return out
}

func (s *InMemoryStore) CreateAssistantCommand(input domain.AssistantCommand) (domain.AssistantCommand, error) {
	if input.UserID == "" || input.Command == "" {
		return domain.AssistantCommand{}, errors.New("missing assistant command fields")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	input.ID = newID()
	if input.Status == "" {
		input.Status = "pending"
	}
	if input.Status == "awaiting_approval" && input.ApprovalID == "" {
		input.ApprovalID = "appr_" + uuid.NewString()
		input.ApprovalExpiresAt = now().Add(10 * time.Minute)
	}
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.assistantCmds[input.ID] = &input
	return input, nil
}

func (s *InMemoryStore) GetAssistantCommand(id domain.ID) (*domain.AssistantCommand, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.assistantCmds[id]
	if !ok {
		return nil, false
	}
	cp := *c
	return &cp, true
}

func (s *InMemoryStore) ListAssistantCommands(userID domain.ID) []domain.AssistantCommand {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AssistantCommand{}
	for _, c := range s.assistantCmds {
		if c.UserID == userID {
			out = append(out, *c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *InMemoryStore) UpdateAssistantCommand(id domain.ID, mutate func(*domain.AssistantCommand)) (*domain.AssistantCommand, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.assistantCmds[id]
	if !ok {
		return nil, errors.New("assistant command not found")
	}
	if mutate == nil {
		return nil, errors.New("missing update mutation")
	}
	mutated := *c
	mutate(&mutated)
	mutated.UpdatedAt = now()
	cp := mutated
	*c = cp
	return &cp, nil
}

func (s *InMemoryStore) CreateBankReconciliation(input domain.BankReconciliation) (domain.BankReconciliation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if input.ID == "" {
		input.ID = newID()
	}
	if input.CreatedAt.IsZero() {
		input.CreatedAt = now()
	}
	if input.UpdatedAt.IsZero() {
		input.UpdatedAt = now()
	}
	cp := input
	s.bankRecon[input.ID] = &cp
	return cp, nil
}

func (s *InMemoryStore) ListBankReconciliations(userID domain.ID, connectionID domain.ID) []domain.BankReconciliation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankReconciliation{}
	for _, r := range s.bankRecon {
		if userID != "" && r.UserID != userID {
			continue
		}
		if connectionID != "" && r.ConnectionID != connectionID {
			continue
		}
		out = append(out, *r)
	}
	return out
}

func (s *InMemoryStore) ListAllBankConnections() []domain.BankConnection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankConnection{}
	for _, c := range s.bankConnections {
		out = append(out, *c)
	}
	return out
}

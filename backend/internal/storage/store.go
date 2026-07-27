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
	mu                sync.RWMutex
	users             map[domain.ID]*domain.User
	workspaces        map[domain.ID]*domain.Workspace
	memberships       map[domain.ID]*domain.WorkspaceMember
	portfolios        map[domain.ID]*domain.Portfolio
	accounts          map[domain.ID]*domain.Account
	transactions      map[domain.ID]*domain.Transaction
	transfers         map[domain.ID]*domain.Transfer
	loans             map[domain.ID]*domain.Loan
	loanPayments      map[domain.ID]*domain.LoanPayment
	properties        map[domain.ID]*domain.Property
	propertyValues    map[domain.ID]*domain.PropertyValuation
	assets            map[domain.ID]*domain.Asset
	assetValues       map[domain.ID]*domain.AssetValuation
	budgets           map[domain.ID]*domain.Budget
	budgetAllocs      map[domain.ID]*domain.BudgetAllocation
	forecast          map[domain.ID]*domain.ForecastScenario
	bankConnections   map[domain.ID]*domain.BankConnection
	bankFeed          map[domain.ID]*domain.BankFeedTransaction
	bankFeedKeys      map[string]domain.ID
	bankFeedEvents    map[domain.ID]*domain.BankFeedEvent
	bankFeedEventKeys map[string]domain.ID
	bankRecon         map[domain.ID]*domain.BankReconciliation
	automationRules   map[domain.ID]*domain.AutomationRule
	bankPaymentReqs   map[domain.ID]*domain.BankPaymentRequest
	assistantCmds     map[domain.ID]*domain.AssistantCommand
	auditLogs         map[domain.ID]*domain.AuditLog
	userSettings      map[domain.ID]*domain.UserSettings
	// idempotency keeps track of processed idempotency keys in-memory.
	idempotencyKeys map[string]time.Time
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		users:             map[domain.ID]*domain.User{},
		userSettings:      map[domain.ID]*domain.UserSettings{},
		workspaces:        map[domain.ID]*domain.Workspace{},
		memberships:       map[domain.ID]*domain.WorkspaceMember{},
		portfolios:        map[domain.ID]*domain.Portfolio{},
		accounts:          map[domain.ID]*domain.Account{},
		transactions:      map[domain.ID]*domain.Transaction{},
		transfers:         map[domain.ID]*domain.Transfer{},
		loans:             map[domain.ID]*domain.Loan{},
		loanPayments:      map[domain.ID]*domain.LoanPayment{},
		properties:        map[domain.ID]*domain.Property{},
		propertyValues:    map[domain.ID]*domain.PropertyValuation{},
		assets:            map[domain.ID]*domain.Asset{},
		assetValues:       map[domain.ID]*domain.AssetValuation{},
		budgets:           map[domain.ID]*domain.Budget{},
		budgetAllocs:      map[domain.ID]*domain.BudgetAllocation{},
		forecast:          map[domain.ID]*domain.ForecastScenario{},
		bankConnections:   map[domain.ID]*domain.BankConnection{},
		bankFeed:          map[domain.ID]*domain.BankFeedTransaction{},
		bankFeedKeys:      map[string]domain.ID{},
		bankFeedEvents:    map[domain.ID]*domain.BankFeedEvent{},
		bankFeedEventKeys: map[string]domain.ID{},
		bankRecon:         map[domain.ID]*domain.BankReconciliation{},
		automationRules:   map[domain.ID]*domain.AutomationRule{},
		bankPaymentReqs:   map[domain.ID]*domain.BankPaymentRequest{},
		assistantCmds:     map[domain.ID]*domain.AssistantCommand{},
		auditLogs:         map[domain.ID]*domain.AuditLog{},
		idempotencyKeys:   map[string]time.Time{},
	}
}

func newID() domain.ID {
	return domain.ID(uuid.NewString())
}

func now() time.Time {
	return time.Now().UTC()
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
	s.users[id] = &domain.User{
		Timestamped: domain.Timestamped{
			ID:        id,
			CreatedAt: now(),
			UpdatedAt: now(),
		},
		Email:    email,
		Name:     name,
		Password: password,
	}
	return id
}

func (s *InMemoryStore) CreateAuditLog(input domain.AuditLog) (domain.AuditLog, error) {
	if strings.TrimSpace(string(input.WorkspaceID)) == "" {
		return domain.AuditLog{}, errors.New("workspaceId is required")
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

func (s *InMemoryStore) ListAuditLogs(workspaceID domain.ID) []domain.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.AuditLog, 0)
	for _, a := range s.auditLogs {
		if a.WorkspaceID != workspaceID {
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

func (s *InMemoryStore) GetWorkspace(id domain.ID) (*domain.Workspace, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ws, ok := s.workspaces[id]
	if !ok {
		return nil, false
	}
	cp := *ws
	return &cp, true
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
	if input.WorkspaceID == "" || input.AccountID == "" || input.Currency == "" {
		return domain.Transaction{}, errors.New("workspaceId, accountId and currency are required")
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
	if acc.WorkspaceID != input.WorkspaceID {
		return domain.Transaction{}, errors.New("account does not belong to workspace")
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

func (s *InMemoryStore) CreateWorkspace(name, baseCurrency string, ownerID domain.ID) (*domain.Workspace, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	ws := &domain.Workspace{
		Timestamped: domain.Timestamped{
			ID:        id,
			CreatedAt: now(),
			UpdatedAt: now(),
		},
		Name:         name,
		BaseCurrency: baseCurrency,
	}
	s.workspaces[id] = ws
	s.memberships[newID()] = &domain.WorkspaceMember{
		Timestamped: domain.Timestamped{
			ID:        newID(),
			CreatedAt: now(),
			UpdatedAt: now(),
		},
		WorkspaceID: id,
		UserID:      ownerID,
		Role:        domain.RoleOwner,
	}

	// Auto-seed one default portfolio for onboarding workflow.
	pID := newID()
	s.portfolios[pID] = &domain.Portfolio{
		Timestamped: domain.Timestamped{
			ID:        pID,
			CreatedAt: now(),
			UpdatedAt: now(),
		},
		WorkspaceID:  id,
		Name:         "Default",
		BaseCurrency: baseCurrency,
	}
	return ws, nil
}

func (s *InMemoryStore) ListWorkspaces(userID domain.ID) []domain.Workspace {
	s.mu.RLock()
	defer s.mu.RUnlock()
	workspaceIDs := map[domain.ID]struct{}{}
	for _, m := range s.memberships {
		if m.UserID == userID {
			workspaceIDs[m.WorkspaceID] = struct{}{}
		}
	}
	out := make([]domain.Workspace, 0, len(workspaceIDs))
	for wid := range workspaceIDs {
		if ws, ok := s.workspaces[wid]; ok {
			cp := *ws
			out = append(out, cp)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *InMemoryStore) GetWorkspaceMemberRole(userID, workspaceID domain.ID) (domain.Role, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.memberships {
		if m.UserID == userID && m.WorkspaceID == workspaceID {
			return m.Role, true
		}
	}
	return "", false
}

func (s *InMemoryStore) CreateAccount(input domain.Account) (domain.Account, error) {
	if input.WorkspaceID == "" || input.Name == "" {
		return domain.Account{}, errors.New("workspaceId and name are required")
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

func (s *InMemoryStore) CreatePortfolio(input domain.Portfolio) (domain.Portfolio, error) {
	if input.WorkspaceID == "" || input.Name == "" {
		return domain.Portfolio{}, errors.New("workspaceId and name are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	input.ID = newID()
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.portfolios[input.ID] = &input
	return input, nil
}

func (s *InMemoryStore) ListPortfolios(workspaceID domain.ID) []domain.Portfolio {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Portfolio{}
	for _, p := range s.portfolios {
		if p.WorkspaceID == workspaceID {
			out = append(out, *p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *InMemoryStore) FirstPortfolio(workspaceID domain.ID) (domain.Portfolio, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.portfolios {
		if p.WorkspaceID == workspaceID {
			cp := *p
			return cp, true
		}
	}
	return domain.Portfolio{}, false
}

func (s *InMemoryStore) ListAccounts(workspaceID domain.ID) []domain.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Account{}
	for _, a := range s.accounts {
		if a.WorkspaceID == workspaceID {
			out = append(out, *a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *InMemoryStore) DeleteAccount(workspaceID domain.ID, id domain.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.accounts[id]
	if !ok || a.WorkspaceID != workspaceID {
		return errors.New("account not found")
	}
	delete(s.accounts, id)
	return nil
}

func (s *InMemoryStore) DeletePortfolio(workspaceID domain.ID, id domain.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.portfolios[id]
	if !ok || p.WorkspaceID != workspaceID {
		return errors.New("portfolio not found")
	}
	delete(s.portfolios, id)
	return nil
}

func (s *InMemoryStore) DeleteLoan(workspaceID domain.ID, id domain.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.loans[id]
	if !ok || l.WorkspaceID != workspaceID {
		return errors.New("loan not found")
	}
	delete(s.loans, id)
	return nil
}

func (s *InMemoryStore) DeleteProperty(workspaceID domain.ID, id domain.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	pr, ok := s.properties[id]
	if !ok || pr.WorkspaceID != workspaceID {
		return errors.New("property not found")
	}
	delete(s.properties, id)
	return nil
}

func (s *InMemoryStore) DeleteAsset(workspaceID domain.ID, id domain.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ast, ok := s.assets[id]
	if !ok || ast.WorkspaceID != workspaceID {
		return errors.New("asset not found")
	}
	delete(s.assets, id)
	return nil
}

func (s *InMemoryStore) CreateTransaction(input domain.Transaction) (domain.Transaction, error) {
	if input.WorkspaceID == "" || input.AccountID == "" || input.Amount == "" {
		return domain.Transaction{}, errors.New("workspaceId, accountId and amount are required")
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

func (s *InMemoryStore) ListTransactions(workspaceID domain.ID, accountID domain.ID) []domain.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Transaction{}
	for _, t := range s.transactions {
		if t.WorkspaceID != workspaceID {
			continue
		}
		if accountID != "" && t.AccountID != accountID {
			continue
		}
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OccurredAt.After(out[j].OccurredAt) })
	return out
}

func (s *InMemoryStore) CreateTransfer(input domain.Transfer) (domain.Transfer, error) {
	if input.WorkspaceID == "" || input.FromAccountID == "" || input.ToAccountID == "" || input.Amount == "" {
		return domain.Transfer{}, errors.New("missing required transfer fields")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := newID()
	input.ID = id
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.transfers[id] = &input
	return input, nil
}

func (s *InMemoryStore) CreateLoan(input domain.Loan) (domain.Loan, error) {
	if input.WorkspaceID == "" || input.PrincipalInitial == "" || input.AnnualRate == "" {
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

func (s *InMemoryStore) ListLoans(workspaceID domain.ID) []domain.Loan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Loan{}
	for _, l := range s.loans {
		if l.WorkspaceID == workspaceID {
			out = append(out, *l)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *InMemoryStore) ListLoanPayments(workspaceID domain.ID, loanID domain.ID) []domain.LoanPayment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.LoanPayment{}
	for _, p := range s.loanPayments {
		if p.WorkspaceID != workspaceID {
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
	if input.WorkspaceID == "" || input.LoanID == "" {
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

func (s *InMemoryStore) CreateProperty(input domain.Property) (domain.Property, error) {
	if input.WorkspaceID == "" || input.Name == "" {
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

func (s *InMemoryStore) ListProperties(workspaceID domain.ID) []domain.Property {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Property{}
	for _, p := range s.properties {
		if p.WorkspaceID == workspaceID {
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

func (s *InMemoryStore) ListPropertyValues(workspaceID domain.ID) []domain.PropertyValuation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.PropertyValuation{}
	for _, value := range s.propertyValues {
		p, ok := s.properties[value.PropertyID]
		if !ok || p.WorkspaceID != workspaceID {
			continue
		}
		out = append(out, *value)
	}
	return out
}

func (s *InMemoryStore) CreateAsset(input domain.Asset) (domain.Asset, error) {
	if input.WorkspaceID == "" || input.Name == "" {
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

func (s *InMemoryStore) ListAssets(workspaceID domain.ID) []domain.Asset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Asset{}
	for _, a := range s.assets {
		if a.WorkspaceID == workspaceID {
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

func (s *InMemoryStore) ListAssetValues(workspaceID domain.ID) []domain.AssetValuation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AssetValuation{}
	for _, value := range s.assetValues {
		a, ok := s.assets[value.AssetID]
		if !ok || a.WorkspaceID != workspaceID {
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
	if input.WorkspaceID == "" || input.Period == "" {
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
	if input.WorkspaceID == "" || input.Period == "" {
		return domain.Budget{}, errors.New("missing required budget fields")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, b := range s.budgets {
		if b.WorkspaceID != input.WorkspaceID {
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

func (s *InMemoryStore) ListBudgets(workspaceID domain.ID, period string) []domain.Budget {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Budget{}
	for _, b := range s.budgets {
		if b.WorkspaceID == workspaceID && b.Period == period {
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
	if input.WorkspaceID == "" || input.Name == "" {
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

func (s *InMemoryStore) ListForecastScenarios(workspaceID domain.ID) []domain.ForecastScenario {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.ForecastScenario{}
	for _, f := range s.forecast {
		if f.WorkspaceID == workspaceID {
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
	if input.WorkspaceID == "" {
		return domain.BankConnection{}, errors.New("workspaceId is required")
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

func (s *InMemoryStore) ListBankConnections(workspaceID domain.ID) []domain.BankConnection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankConnection{}
	for _, c := range s.bankConnections {
		if c.WorkspaceID == workspaceID {
			out = append(out, *c)
		}
	}
	return out
}

func (s *InMemoryStore) EnqueueBankFeedEvent(input domain.BankFeedEvent) (domain.BankFeedEvent, error) {
	if input.WorkspaceID == "" || input.ConnectionID == "" {
		return domain.BankFeedEvent{}, errors.New("workspaceId and connectionId are required")
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

func (s *InMemoryStore) ListBankFeedEvents(workspaceID domain.ID, state string) []domain.BankFeedEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankFeedEvent{}
	for _, e := range s.bankFeedEvents {
		if workspaceID != "" && e.WorkspaceID != workspaceID {
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
	if input.WorkspaceID == "" || input.ConnectionID == "" {
		return domain.BankFeedTransaction{}, errors.New("workspaceId and connectionId are required")
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

func (s *InMemoryStore) ListBankFeed(workspaceID domain.ID) []domain.BankFeedTransaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankFeedTransaction{}
	for _, t := range s.bankFeed {
		if t.WorkspaceID == workspaceID {
			out = append(out, *t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OccurredAt.After(out[j].OccurredAt) })
	return out
}

func (s *InMemoryStore) ListBankFeedByState(workspaceID domain.ID, state domain.TransactionPostingState) []domain.BankFeedTransaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankFeedTransaction{}
	for _, t := range s.bankFeed {
		if t.PostingState != state {
			continue
		}
		if workspaceID != "" && t.WorkspaceID != workspaceID {
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
	})
}

func (s *InMemoryStore) CreateAutomationRule(input domain.AutomationRule) (domain.AutomationRule, error) {
	if input.WorkspaceID == "" {
		return domain.AutomationRule{}, errors.New("workspaceId is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	input.ID = newID()
	input.CreatedAt = now()
	input.UpdatedAt = now()
	s.automationRules[input.ID] = &input
	return input, nil
}

func (s *InMemoryStore) ListWorkspaceAutomationRules(workspaceID domain.ID) []domain.AutomationRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AutomationRule{}
	for _, r := range s.automationRules {
		if r.WorkspaceID == workspaceID {
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

func (s *InMemoryStore) ListWorkspaceRules(workspaceID domain.ID) []domain.AutomationRule {
	return s.ListWorkspaceAutomationRules(workspaceID)
}

func (s *InMemoryStore) GetWorkspaceRules(workspaceID domain.ID, accountID domain.ID, direction string) []domain.AutomationRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AutomationRule{}
	for _, r := range s.automationRules {
		if r.WorkspaceID != workspaceID || !r.Enabled {
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
	if input.WorkspaceID == "" || input.LoanID == "" || input.Code == "" {
		return domain.BankPaymentRequest{}, errors.New("workspaceId, loanId and paymentCode are required")
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

func (s *InMemoryStore) GetBankPaymentRequestByCode(workspaceID domain.ID, code string) (*domain.BankPaymentRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, req := range s.bankPaymentReqs {
		if req.WorkspaceID != workspaceID || req.Code == "" {
			continue
		}
		if req.Code == code {
			cp := *req
			return &cp, true
		}
	}
	return nil, false
}

func (s *InMemoryStore) ListBankPaymentRequests(workspaceID domain.ID) []domain.BankPaymentRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankPaymentRequest{}
	for _, req := range s.bankPaymentReqs {
		if req.WorkspaceID == workspaceID {
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

func (s *InMemoryStore) ListAutomationRules(workspaceID domain.ID) []domain.AutomationRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AutomationRule{}
	for _, r := range s.automationRules {
		if r.WorkspaceID == workspaceID {
			out = append(out, *r)
		}
	}
	return out
}

func (s *InMemoryStore) CreateAssistantCommand(input domain.AssistantCommand) (domain.AssistantCommand, error) {
	if input.WorkspaceID == "" || input.UserID == "" || input.Command == "" {
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

func (s *InMemoryStore) ListAssistantCommands(workspaceID domain.ID) []domain.AssistantCommand {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AssistantCommand{}
	for _, c := range s.assistantCmds {
		if c.WorkspaceID == workspaceID {
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

func (s *InMemoryStore) ListBankReconciliations(workspaceID domain.ID, connectionID domain.ID) []domain.BankReconciliation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.BankReconciliation{}
	for _, r := range s.bankRecon {
		if workspaceID != "" && r.WorkspaceID != workspaceID {
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


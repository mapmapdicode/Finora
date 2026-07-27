package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"wealthos-backend/internal/cache"
	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/storage"
)

type WealthService struct {
	store storage.Store
	cache cache.CacheService
}


type LoginResult struct {
	User      domain.User       `json:"user"`
	Token     string            `json:"token"`
	Workspace *domain.Workspace `json:"workspace,omitempty"`
}

type NetWorthResult struct {
	AsOfAt          time.Time           `json:"asOfAt"`
	BaseCurrency    string              `json:"baseCurrency"`
	NetWorth        string              `json:"netWorth"`
	NetWorthChange  string              `json:"netWorthChange"`
	Cash            string              `json:"cash"`
	Liabilities     string              `json:"liabilities"`
	SnapshotVersion int                 `json:"snapshotVersion"`
	Assets          NetWorthAssets      `json:"assets"`
	DataQuality     NetWorthDataQuality `json:"dataQuality"`
	Attribution     NetWorthAttribution `json:"attribution"`
}

type SePayWebhookEvent struct {
	ConnectionID string `json:"connectionId"`
	AccountID    string `json:"accountId"`
	Direction    string `json:"direction"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	Counterparty string `json:"counterparty"`
	Description  string `json:"description"`
	Reference    string `json:"reference"`
	Content      string `json:"content"`
	ExternalID   string `json:"externalTransactionId"`
	OccurredAt   string `json:"occurredAt"`
}

type bankFeedClassification struct {
	txType      domain.TransactionType
	categoryID  domain.ID
	accountID   domain.ID
	confidence  float64
	reason      string
	autoClassed bool
	loanID      domain.ID
}

const sepayAutoPostThreshold = 70.0
const bankDirectionIn = "in"
const bankDirectionOut = "out"
const valuationStaleAfterDays = 45
const maxPortfolioSnapshots = 36
const defaultPortfolioSnapshotLimit = 20
const maxBankFeedEventAttempts = 3

var (
	inboundTransferKeywords = []string{
		"chuyen tien", "chuyen khoan", "chuyen", "transfer", "bank transfer", "dientu",
		"chuyen khoan no bo", "chuyen khoan trong noi", "chuyen khoan noi bo",
	}
	inboundIncomeKeywords = []string{
		"luong", "luong thuong", "thu nhap", "salary", "hoa hong", "nhan tien",
		"refund", "bonus", "freelance", "phi", "thay lo", "chi tieu khach", "luong phi",
	}
)

var (
	snapshotMu sync.RWMutex
	// historyByWorkspace stores bounded net-worth snapshots per workspace and optional portfolio scope.
	historyByWorkspace = map[domain.ID]map[domain.ID][]netWorthSnapshot{}
	snapshotVersion    = map[domain.ID]map[domain.ID]int{}
)

type netWorthSnapshot struct {
	AsOfAt               time.Time
	SnapshotVersion      int
	BaseCurrency         string
	NetWorth             float64
	NetWorthChange       float64
	Cash                 float64
	Liabilities          float64
	Receivables          float64
	Property             float64
	OtherAssets          float64
	AccruedInterest      float64
	StaleValuations      int
	ReconciledAccounts   int
	ExternalCashFlow     float64
	AccruedInterestDelta float64
	ValuationChange      float64
	AccruedFee           float64
}

type NetWorthAssets struct {
	Cash            string `json:"cash"`
	Receivables     string `json:"receivables"`
	Property        string `json:"property"`
	OtherAssets     string `json:"otherAssets"`
	AccruedInterest string `json:"accruedInterest"`
}

type NetWorthDataQuality struct {
	ReconciledAccounts int    `json:"reconciledAccounts"`
	StaleValuations    int    `json:"staleValuations"`
	AsOfSource         string `json:"asOfSource"`
}

type NetWorthAttribution struct {
	ExternalCashFlow string `json:"externalCashFlow"`
	AccruedInterest  string `json:"accruedInterest"`
	ValuationChange  string `json:"valuationChange"`
	AccruedFee       string `json:"accruedFee"`
}

type PortfolioSnapshotPage struct {
	Items      []NetWorthResult `json:"items"`
	NextCursor string           `json:"nextCursor"`
}

type LoanAccrualRow struct {
	PeriodStart        time.Time `json:"periodStart"`
	PeriodEnd          time.Time `json:"periodEnd"`
	Principal          string    `json:"principal"`
	AccruedInterest    string    `json:"accruedInterest"`
	PaidInterest       string    `json:"paidInterest"`
	UnpaidInterest     string    `json:"unpaidInterest"`
	RemainingPrincipal string    `json:"remainingPrincipal"`
	Days               int       `json:"days"`
}

type LoanAccrualResponse struct {
	LoanID               string           `json:"loanId"`
	WorkspaceID          string           `json:"workspaceId"`
	AsOfAt               time.Time        `json:"asOfAt"`
	Status               string           `json:"status"`
	Currency             string           `json:"currency"`
	PrincipalInitial     string           `json:"principalInitial"`
	PrincipalBalance     string           `json:"principalBalance"`
	AnnualRate           string           `json:"annualRate"`
	DayCountBasis        string           `json:"dayCountBasis"`
	TotalAccruedInterest string           `json:"totalAccruedInterest"`
	TotalPaidInterest    string           `json:"totalPaidInterest"`
	UnpaidInterest       string           `json:"unpaidInterest"`
	Rows                 []LoanAccrualRow `json:"accruals"`
}

type LoanAccrualSummary struct {
	TotalDays    int
	TotalAccrued float64
	TotalPaid    float64
}

type BudgetRow struct {
	CategoryID string `json:"categoryId"`
	Limit      string `json:"limit"`
	Spent      string `json:"spent"`
	Currency   string `json:"currency"`
}

type BudgetPeriod struct {
	Period      string      `json:"period"`
	WorkspaceID string      `json:"workspaceId"`
	AsOfAt      time.Time   `json:"asOfAt"`
	Rows        []BudgetRow `json:"rows"`
}

type PaymentRequestCreate struct {
	Amount    string
	Currency  string
	ExpiresAt time.Time
	Note      string
}

func NewWealthService(store storage.Store, cacheService cache.CacheService) *WealthService {
	return &WealthService{store: store, cache: cacheService}
}

func (s *WealthService) GetUserSettings(ctx context.Context, userID domain.ID) (*domain.UserSettings, error) {
	if userID == "" {
		return &domain.UserSettings{AmountDisplayMode: domain.AmountDisplayModeFull}, nil
	}

	// 1. Check Redis cache first
	if s.cache != nil {
		cached, err := s.cache.GetUserSettings(ctx, string(userID))
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	// 2. Fallback to PostgreSQL DB / Store
	settings, err := s.store.GetUserSettings(userID)
	if err != nil {
		return &domain.UserSettings{
			UserID:            userID,
			AmountDisplayMode: domain.AmountDisplayModeFull,
			UpdatedAt:         time.Now(),
		}, nil
	}

	// 3. Populate Redis Cache asynchronously or synchronously
	if s.cache != nil && settings != nil {
		_ = s.cache.SetUserSettings(ctx, string(userID), settings)
	}

	return settings, nil
}

func (s *WealthService) UpdateUserSettings(ctx context.Context, userID domain.ID, mode domain.AmountDisplayMode) (*domain.UserSettings, error) {
	if userID == "" {
		return nil, errors.New("unauthorized: missing user ID")
	}

	if mode != domain.AmountDisplayModeCompact && mode != domain.AmountDisplayModeFull {
		mode = domain.AmountDisplayModeFull
	}

	updated, err := s.store.UpsertUserSettings(domain.UserSettings{
		UserID:            userID,
		AmountDisplayMode: mode,
	})
	if err != nil {
		return nil, err
	}

	// Update Redis cache
	if s.cache != nil && updated != nil {
		_ = s.cache.SetUserSettings(ctx, string(userID), updated)
	}

	return updated, nil
}


func (s *WealthService) RegisterUser(email, password, name string) (domain.User, error) {
	if email == "" || password == "" || name == "" {
		return domain.User{}, errors.New("email, password, name are required")
	}
	if _, exists := s.store.GetUserByEmail(email); exists {
		return domain.User{}, errors.New("email already exists")
	}
	id := s.store.SeedDemoUser(email, name, password)
	user, ok := s.store.GetUser(id)
	if !ok {
		return domain.User{}, errors.New("cannot create user")
	}
	return *user, nil
}

func (s *WealthService) Authenticate(email, password string) (LoginResult, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	u, ok := s.store.GetUserByEmail(email)
	if !ok || strings.TrimSpace(u.Password) != password {
		return LoginResult{}, errors.New("invalid credentials")
	}
	wsList := s.store.ListWorkspaces(u.ID)
	var ws *domain.Workspace
	if len(wsList) > 0 {
		cp := wsList[0]
		ws = &cp
	}
	token := "token-" + string(u.ID)
	return LoginResult{User: *u, Token: token, Workspace: ws}, nil
}

func (s *WealthService) CurrentNetWorth(workspaceID string) (NetWorthResult, error) {
	return s.ComputeNetWorth(domain.ID(workspaceID))
}

func (s *WealthService) ComputeNetWorth(workspaceID domain.ID) (NetWorthResult, error) {
	return s.ComputeNetWorthAt(workspaceID, time.Time{})
}

func (s *WealthService) ComputeNetWorthAt(workspaceID domain.ID, asOf time.Time) (NetWorthResult, error) {
	return s.computeNetWorthForPortfolioAt(workspaceID, "", asOf, true)
}

func (s *WealthService) computeNetWorthForPortfolio(workspaceID, portfolioID domain.ID) (NetWorthResult, error) {
	return s.computeNetWorthForPortfolioAt(workspaceID, portfolioID, time.Time{}, true)
}

func (s *WealthService) computeNetWorthForPortfolioAt(workspaceID, portfolioID domain.ID, asOf time.Time, persist bool) (NetWorthResult, error) {
	ws, ok := s.store.GetWorkspace(workspaceID)
	if !ok {
		return NetWorthResult{}, errors.New("workspace not found")
	}

	if asOf.IsZero() {
		asOf = nowUTC()
	}

	cash := s.computeCash(workspaceID, portfolioID, asOf)
	liabilities, accruedPayableInterest := s.computeLoanLiabilityExposure(workspaceID, portfolioID, asOf)
	receivables, accruedReceivableInterest := s.computeLoanReceivableExposure(workspaceID, portfolioID, asOf)
	property, otherAssets, staleValuations := s.computeAssetValuationWorkspace(workspaceID, portfolioID, asOf)
	valuationAssets := property + otherAssets
	accruedInterestNet := accruedReceivableInterest - accruedPayableInterest
	reconciledAccounts := len(s.store.ListAccounts(workspaceID))
	if portfolioID != "" {
		reconciledAccounts = len(s.portfolioAccounts(workspaceID, portfolioID))
	}

	net := cash + valuationAssets + receivables + accruedReceivableInterest - liabilities - accruedPayableInterest
	current := netWorthSnapshot{
		AsOfAt:             asOf,
		BaseCurrency:       ws.BaseCurrency,
		NetWorth:           net,
		Cash:               cash,
		Liabilities:        liabilities,
		Receivables:        receivables,
		Property:           property,
		OtherAssets:        otherAssets,
		AccruedInterest:    accruedInterestNet,
		StaleValuations:    staleValuations,
		ReconciledAccounts: reconciledAccounts,
		AccruedFee:         0,
	}

	prev, hasPrev := s.snapshotAtOrBefore(workspaceID, portfolioID, asOf)
	if hasPrev {
		current.NetWorthChange = current.NetWorth - prev.NetWorth
		current.ExternalCashFlow = s.computeExternalCashFlow(current, prev)
		current.ValuationChange = (current.Property + current.OtherAssets) - (prev.Property + prev.OtherAssets)
		current.AccruedInterestDelta = current.AccruedInterest - prev.AccruedInterest
	} else {
		current.NetWorthChange = 0
	}

	if persist {
		recorded := s.addNetWorthSnapshot(workspaceID, portfolioID, current)
		return s.formatNetWorthResult(recorded), nil
	}

	return s.formatNetWorthResult(current), nil
}

func (s *WealthService) portfolioAccounts(workspaceID, portfolioID domain.ID) []domain.Account {
	accounts := s.store.ListAccounts(workspaceID)
	if portfolioID == "" {
		return accounts
	}
	result := make([]domain.Account, 0)
	for _, account := range accounts {
		if account.PortfolioID == portfolioID || account.PortfolioID == "" {
			result = append(result, account)
		}
	}
	return result
}

func (s *WealthService) computeCash(workspaceID, portfolioID domain.ID, asOf time.Time) float64 {
	txs := s.store.ListTransactions(workspaceID, "")
	accountByID := map[domain.ID]domain.Account{}
	if portfolioID != "" {
		for _, account := range s.portfolioAccounts(workspaceID, portfolioID) {
			accountByID[account.ID] = account
		}
	}
	cash := 0.0
	for _, t := range txs {
		if !s.matchesPortfolio(portfolioID, t.PortfolioID, t.AccountID, accountByID) {
			continue
		}
		if !t.OccurredAt.IsZero() && t.OccurredAt.After(asOf) {
			continue
		}
		if t.Status != "" && t.Status != domain.TransactionStatusPosted && t.Status != domain.TransactionStatusReconciled {
			continue
		}
		amt, err := parseAmount(t.Amount)
		if err != nil {
			continue
		}
		switch t.Type {
		case domain.TransactionTypeIncome:
			cash += amt
		case domain.TransactionTypeExpense, domain.TransactionTypeInvestment, domain.TransactionTypeLoanDisbursement:
			cash -= amt
		case domain.TransactionTypeLoanPayment:
			// Keep loan payment sign aligned with loan direction when possible.
			sign := 1.0
			if paymentSign, ok := s.loanPaymentSign(workspaceID, portfolioID, t); ok {
				sign = paymentSign
			}
			cash += sign * amt
		case domain.TransactionTypeTransfer, domain.TransactionTypeValuationAdj:
			// ignored in cash rollup
		default:
			// ignore unknown types
		}
	}
	return cash
}

func (s *WealthService) matchesPortfolio(portfolioID, txPortfolioID, txAccountID domain.ID, accountByID map[domain.ID]domain.Account) bool {
	if portfolioID == "" {
		return true
	}
	if txPortfolioID == portfolioID {
		return true
	}
	if txPortfolioID != "" && txPortfolioID != portfolioID {
		return false
	}
	if txAccountID != "" {
		if account, ok := accountByID[txAccountID]; ok {
			return account.PortfolioID == portfolioID || account.PortfolioID == ""
		}
		if acc, ok2 := s.store.GetAccount(txAccountID); ok2 {
			return acc.PortfolioID == portfolioID || acc.PortfolioID == ""
		}
	}
	return true
}

func (s *WealthService) snapshotAtOrBefore(workspaceID, portfolioID domain.ID, asOf time.Time) (netWorthSnapshot, bool) {
	snapshotMu.RLock()
	defer snapshotMu.RUnlock()
	scopeHistory, ok := historyByWorkspace[workspaceID]
	if !ok {
		return netWorthSnapshot{}, false
	}
	history, ok := scopeHistory[portfolioID]
	if !ok || len(history) == 0 {
		return netWorthSnapshot{}, false
	}

	var chosen netWorthSnapshot
	for i := len(history) - 1; i >= 0; i-- {
		if !history[i].AsOfAt.After(asOf) {
			chosen = history[i]
			return chosen, true
		}
	}
	return netWorthSnapshot{}, false
}

func (s *WealthService) addNetWorthSnapshot(workspaceID, portfolioID domain.ID, snapshot netWorthSnapshot) netWorthSnapshot {
	snapshotMu.Lock()
	defer snapshotMu.Unlock()

	scopeVersions, ok := snapshotVersion[workspaceID]
	if !ok {
		scopeVersions = map[domain.ID]int{}
		snapshotVersion[workspaceID] = scopeVersions
	}

	nextVersion := scopeVersions[portfolioID] + 1
	scopeVersions[portfolioID] = nextVersion
	snapshot.SnapshotVersion = nextVersion

	scopeHistory, ok := historyByWorkspace[workspaceID]
	if !ok {
		scopeHistory = map[domain.ID][]netWorthSnapshot{}
		historyByWorkspace[workspaceID] = scopeHistory
	}

	list := scopeHistory[portfolioID]
	list = append(list, snapshot)
	if len(list) > maxPortfolioSnapshots {
		list = list[len(list)-maxPortfolioSnapshots:]
	}
	scopeHistory[portfolioID] = append([]netWorthSnapshot{}, list...)
	return snapshot
}

func (s *WealthService) formatNetWorthResult(snapshot netWorthSnapshot) NetWorthResult {
	return NetWorthResult{
		AsOfAt:          snapshot.AsOfAt,
		BaseCurrency:    snapshot.BaseCurrency,
		Cash:            formatMoney(snapshot.Cash),
		Liabilities:     formatMoney(snapshot.Liabilities),
		NetWorth:        formatMoney(snapshot.NetWorth),
		SnapshotVersion: snapshot.SnapshotVersion,
		Assets: NetWorthAssets{
			Cash:            formatMoney(snapshot.Cash),
			Receivables:     formatMoney(snapshot.Receivables),
			Property:        formatMoney(snapshot.Property),
			OtherAssets:     formatMoney(snapshot.OtherAssets),
			AccruedInterest: formatMoney(snapshot.AccruedInterest),
		},
		DataQuality: NetWorthDataQuality{
			ReconciledAccounts: snapshot.ReconciledAccounts,
			StaleValuations:    snapshot.StaleValuations,
			AsOfSource:         "ledger",
		},
		Attribution: NetWorthAttribution{
			ExternalCashFlow: formatMoney(snapshot.ExternalCashFlow),
			AccruedInterest:  formatMoney(snapshot.AccruedInterestDelta),
			ValuationChange:  formatMoney(snapshot.ValuationChange),
			AccruedFee:       formatMoney(snapshot.AccruedFee),
		},
		NetWorthChange: formatMoney(snapshot.NetWorthChange),
	}
}

func (s *WealthService) loanPaymentSign(workspaceID, portfolioID domain.ID, t domain.Transaction) (float64, bool) {
	loans := s.store.ListLoans(workspaceID)
	for _, ln := range loans {
		if !s.matchesEntityToPortfolio(portfolioID, ln.PortfolioID) {
			continue
		}
		pp := s.store.ListLoanPayments(workspaceID, ln.ID)
		for _, p := range pp {
			if p.TransactionID != "" && p.TransactionID == t.ID {
				if ln.Direction == domain.LoanDirectionReceivable {
					return 1, true
				}
				return -1, true
			}
		}
	}
	return 1, false
}

func (s *WealthService) computeLoanLiabilityExposure(workspaceID, portfolioID domain.ID, asOf time.Time) (float64, float64) {
	liability := 0.0
	accrued := 0.0
	loans := s.store.ListLoans(workspaceID)
	for _, loan := range loans {
		if !s.matchesEntityToPortfolio(portfolioID, loan.PortfolioID) {
			continue
		}
		if loan.Direction != domain.LoanDirectionPayable || loan.Status == domain.LoanStatusClosed || loan.Status == domain.LoanStatusCancelled || loan.Status == domain.LoanStatusWrittenOff {
			continue
		}
		principal, err := parseAmount(loan.PrincipalBalance)
		if err == nil {
			liability += principal
		}
		_, summary := s.loanAccrualsByLoan(loan, asOf)
		if summary.TotalAccrued > summary.TotalPaid {
			accrued += summary.TotalAccrued - summary.TotalPaid
		}
	}
	return liability, accrued
}

func (s *WealthService) computeLoanReceivableExposure(workspaceID, portfolioID domain.ID, asOf time.Time) (float64, float64) {
	receivable := 0.0
	accrued := 0.0
	loans := s.store.ListLoans(workspaceID)
	for _, loan := range loans {
		if !s.matchesEntityToPortfolio(portfolioID, loan.PortfolioID) {
			continue
		}
		if loan.Direction != domain.LoanDirectionReceivable || loan.Status == domain.LoanStatusClosed || loan.Status == domain.LoanStatusCancelled || loan.Status == domain.LoanStatusWrittenOff {
			continue
		}
		principal, err := parseAmount(loan.PrincipalBalance)
		if err == nil {
			receivable += principal
		}
		_, summary := s.loanAccrualsByLoan(loan, asOf)
		if summary.TotalAccrued > summary.TotalPaid {
			accrued += summary.TotalAccrued - summary.TotalPaid
		}
	}
	return receivable, accrued
}

func (s *WealthService) matchesEntityToPortfolio(scopePortfolioID, entityPortfolioID domain.ID) bool {
	if scopePortfolioID == "" {
		return true
	}
	return scopePortfolioID == entityPortfolioID || entityPortfolioID == ""
}

func (s *WealthService) loanAccrualsByLoan(loan domain.Loan, asOf time.Time) ([]LoanAccrualRow, LoanAccrualSummary) {
	periodStart := loan.StartAt
	if periodStart.IsZero() {
		periodStart = nowUTC()
	}
	if asOf.IsZero() {
		asOf = nowUTC()
	}
	if asOf.Before(periodStart) {
		asOf = periodStart
	}

	annualRate := parseAmountWithFallback(loan.AnnualRate)
	if annualRate > 1.0 {
		annualRate = annualRate / 100.0
	}
	daysPerYear := parseDayCountBasis(loan.DayCountBasis)
	if daysPerYear <= 0 {
		daysPerYear = 365
	}

	payments := s.store.ListLoanPayments(loan.WorkspaceID, loan.ID)
	sort.Slice(payments, func(i, j int) bool {
		return payments[i].OccurredAt.Before(payments[j].OccurredAt)
	})

	principal, err := parseAmount(loan.PrincipalInitial)
	if err != nil {
		principal = 0
	}

	rows := make([]LoanAccrualRow, 0)
	summary := LoanAccrualSummary{}
	cursor := periodStart
	for _, payment := range payments {
		if payment.OccurredAt.IsZero() || payment.OccurredAt.After(asOf) {
			continue
		}
		if payment.OccurredAt.Before(cursor) {
			continue
		}
		days := elapsedDays(cursor, payment.OccurredAt)
		accrued := 0.0
		paidInterest := parseAmountWithFallback(payment.Interest)
		if principal > 0 && days > 0 && annualRate > 0 {
			accrued = principal * annualRate * float64(days) / float64(daysPerYear)
		}
		unpaid := accrued - paidInterest
		if unpaid < 0 {
			unpaid = 0
		}
		principal -= parseAmountWithFallback(payment.Principal)
		if principal < 0 {
			principal = 0
		}

		rows = append(rows, LoanAccrualRow{
			PeriodStart:        cursor,
			PeriodEnd:          payment.OccurredAt,
			Principal:          formatMoney(principal + parseAmountWithFallback(payment.Principal)),
			AccruedInterest:    formatMoney(accrued),
			PaidInterest:       formatMoney(paidInterest),
			UnpaidInterest:     formatMoney(unpaid),
			RemainingPrincipal: formatMoney(principal),
			Days:               days,
		})
		summary.TotalDays += days
		summary.TotalAccrued += accrued
		summary.TotalPaid += paidInterest
		cursor = payment.OccurredAt
	}

	if asOf.After(cursor) {
		days := elapsedDays(cursor, asOf)
		accrued := 0.0
		if principal > 0 && days > 0 && annualRate > 0 {
			accrued = principal * annualRate * float64(days) / float64(daysPerYear)
		}
		if days > 0 {
			rows = append(rows, LoanAccrualRow{
				PeriodStart:        cursor,
				PeriodEnd:          asOf,
				Principal:          formatMoney(principal),
				AccruedInterest:    formatMoney(accrued),
				PaidInterest:       "0.00",
				UnpaidInterest:     formatMoney(accrued),
				RemainingPrincipal: formatMoney(principal),
				Days:               days,
			})
			summary.TotalDays += days
			summary.TotalAccrued += accrued
		}
	}

	return rows, summary
}

func (s *WealthService) computeLoanAccrualSummary(rows []LoanAccrualRow) LoanAccrualSummary {
	summary := LoanAccrualSummary{}
	for _, row := range rows {
		summary.TotalDays += row.Days
		summary.TotalAccrued += parseAmountWithFallback(row.AccruedInterest)
		summary.TotalPaid += parseAmountWithFallback(row.PaidInterest)
	}
	return summary
}

func (s *WealthService) GetLoanAccruals(loanID string) (LoanAccrualResponse, error) {
	loan, ok := s.store.GetLoan(domain.ID(loanID))
	if !ok {
		return LoanAccrualResponse{}, errors.New("loan not found")
	}

	asOf := nowUTC()
	rows, summary := s.loanAccrualsByLoan(*loan, asOf)
	summary = s.computeLoanAccrualSummary(rows)
	unpaid := summary.TotalAccrued - summary.TotalPaid
	if unpaid < 0 {
		unpaid = 0
	}
	return LoanAccrualResponse{
		LoanID:               string(loan.ID),
		WorkspaceID:          string(loan.WorkspaceID),
		AsOfAt:               asOf,
		Status:               string(loan.Status),
		Currency:             "VND",
		PrincipalInitial:     loan.PrincipalInitial,
		PrincipalBalance:     loan.PrincipalBalance,
		AnnualRate:           loan.AnnualRate,
		DayCountBasis:        firstNonEmpty(loan.DayCountBasis, "365"),
		TotalAccruedInterest: formatMoney(summary.TotalAccrued),
		TotalPaidInterest:    formatMoney(summary.TotalPaid),
		UnpaidInterest:       formatMoney(unpaid),
		Rows:                 rows,
	}, nil
}

func elapsedDays(start, end time.Time) int {
	if start.IsZero() || end.IsZero() || !end.After(start) {
		return 0
	}
	delta := end.UTC().Sub(start.UTC())
	if delta < 0 {
		return 0
	}
	return int(math.Ceil(delta.Hours() / 24))
}

func parseDayCountBasis(value string) int {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "actual/365", "365", "act/365", "365f", "365/365":
		return 365
	case "360", "act/360", "30/360", "30_360", "360/360":
		return 360
	case "366":
		return 366
	default:
		return 365
	}
}

func (s *WealthService) GetBudget(workspaceID domain.ID, period string) (BudgetPeriod, error) {
	if workspaceID == "" {
		return BudgetPeriod{}, errors.New("workspaceId is required")
	}
	period = strings.TrimSpace(period)
	if period == "" {
		return BudgetPeriod{}, errors.New("period is required")
	}
	periodStart, periodEnd, err := parseBudgetPeriod(period)
	if err != nil {
		return BudgetPeriod{}, err
	}

	budgets := s.store.ListBudgets(workspaceID, period)
	rows := make([]BudgetRow, 0, len(budgets))
	for _, budget := range budgets {
		spent, currency := s.budgetSpentForPeriod(workspaceID, periodStart, periodEnd, budget.CategoryID)
		rows = append(rows, BudgetRow{
			CategoryID: string(budget.CategoryID),
			Limit:      budget.Limit,
			Spent:      formatMoney(spent),
			Currency:   currency,
		})
	}
	return BudgetPeriod{
		Period:      period,
		WorkspaceID: string(workspaceID),
		AsOfAt:      nowUTC(),
		Rows:        rows,
	}, nil
}

func (s *WealthService) UpsertBudget(workspaceID domain.ID, period string, categoryID string, limit string) (domain.Budget, error) {
	period = strings.TrimSpace(period)
	if workspaceID == "" {
		return domain.Budget{}, errors.New("workspaceId is required")
	}
	if period == "" {
		return domain.Budget{}, errors.New("period is required")
	}
	if _, err := parseAmount(limit); err != nil {
		return domain.Budget{}, errors.New("invalid limit")
	}
	categoryID = strings.TrimSpace(categoryID)
	if strings.EqualFold(categoryID, "uncategorized") {
		categoryID = ""
	}
	return s.store.UpsertBudget(domain.Budget{
		WorkspaceID: workspaceID,
		Period:      period,
		CategoryID:  domain.ID(categoryID),
		Limit:       limit,
	})
}

func (s *WealthService) budgetSpentForPeriod(workspaceID domain.ID, start, end time.Time, categoryID domain.ID) (float64, string) {
	transactions := s.store.ListTransactions(workspaceID, "")
	total := 0.0
	currency := "VND"
	for _, tx := range transactions {
		if tx.Type != domain.TransactionTypeExpense {
			continue
		}
		if tx.Status != "" && tx.Status != domain.TransactionStatusPosted && tx.Status != domain.TransactionStatusReconciled {
			continue
		}
		if tx.OccurredAt.Before(start) || !tx.OccurredAt.Before(end) {
			continue
		}
		if categoryID != "" && tx.CategoryID != categoryID {
			continue
		}
		amt := parseAmountWithFallback(tx.Amount)
		total += amt
		if tx.Currency != "" {
			currency = tx.Currency
		}
	}
	return total, currency
}

func parseBudgetPeriod(period string) (time.Time, time.Time, error) {
	normalized := strings.TrimSpace(period)
	if normalized == "" {
		return time.Time{}, time.Time{}, errors.New("period is required")
	}
	t, err := time.Parse("2006-01", normalized)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid period format, expected YYYY-MM")
	}
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return start, end, nil
}

func (s *WealthService) computeAssetValuationWorkspace(workspaceID, portfolioID domain.ID, asOf time.Time) (float64, float64, int) {
	propertyValues := s.store.ListPropertyValues(workspaceID)
	assetValues := s.store.ListAssetValues(workspaceID)
	portfolioPropertyIDs := map[domain.ID]struct{}{}
	portfolioAssetIDs := map[domain.ID]struct{}{}
	if portfolioID != "" {
		for _, p := range s.store.ListProperties(workspaceID) {
			if p.PortfolioID == portfolioID {
				portfolioPropertyIDs[p.ID] = struct{}{}
			}
		}
		for _, a := range s.store.ListAssets(workspaceID) {
			if a.PortfolioID == portfolioID {
				portfolioAssetIDs[a.ID] = struct{}{}
			}
		}
	}

	byProperty := map[domain.ID]domain.PropertyValuation{}
	for _, pv := range propertyValues {
		if portfolioID != "" {
			if _, ok := portfolioPropertyIDs[pv.PropertyID]; !ok {
				continue
			}
		}
		if !pv.EffectiveAt.IsZero() && pv.EffectiveAt.After(asOf) {
			continue
		}
		cur := byProperty[pv.PropertyID]
		if pv.EffectiveAt.After(cur.EffectiveAt) || cur.EffectiveAt.IsZero() {
			byProperty[pv.PropertyID] = pv
		}
	}
	property := 0.0
	staleValuations := 0
	staleCutoff := asOf.AddDate(0, 0, -valuationStaleAfterDays)
	for _, pv := range byProperty {
		if v, err := parseAmount(pv.Amount); err == nil {
			property += v
		}
		if isValuationStale(pv.EffectiveAt, staleCutoff) || pv.IsStale {
			staleValuations++
		}
	}

	byAsset := map[domain.ID]domain.AssetValuation{}
	for _, av := range assetValues {
		if portfolioID != "" {
			if _, ok := portfolioAssetIDs[av.AssetID]; !ok {
				continue
			}
		}
		if !av.EffectiveAt.IsZero() && av.EffectiveAt.After(asOf) {
			continue
		}
		cur := byAsset[av.AssetID]
		if av.EffectiveAt.After(cur.EffectiveAt) || cur.EffectiveAt.IsZero() {
			byAsset[av.AssetID] = av
		}
	}
	other := 0.0
	for _, av := range byAsset {
		if v, err := parseAmount(av.Amount); err == nil {
			other += v
		}
		if isValuationStale(av.EffectiveAt, staleCutoff) {
			staleValuations++
		}
	}
	return property, other, staleValuations
}

func (s *WealthService) computeExternalCashFlow(current, previous netWorthSnapshot) float64 {
	return (current.Cash + current.Receivables - current.Liabilities) -
		(previous.Cash + previous.Receivables - previous.Liabilities)
}

func isValuationStale(effectiveAt time.Time, cutoff time.Time) bool {
	if effectiveAt.IsZero() {
		return false
	}
	return effectiveAt.Before(cutoff)
}

func (s *WealthService) CreateTransaction(input domain.Transaction) (domain.Transaction, error) {
	if input.Type == "" {
		return domain.Transaction{}, errors.New("type is required")
	}
	if _, ok := s.validateTransactionType(input.Type); !ok {
		return domain.Transaction{}, fmt.Errorf("transaction type is invalid")
	}
	if _, ok := s.validateTransactionStatus(input.Status); !ok {
		return domain.Transaction{}, fmt.Errorf("transaction status is invalid")
	}

	if input.AccountID == "" {
		return domain.Transaction{}, errors.New("accountId is required")
	}

	amount, err := parseAmount(input.Amount)
	if err != nil {
		return domain.Transaction{}, errors.New("invalid amount")
	}
	if amount <= 0 {
		return domain.Transaction{}, errors.New("amount must be greater than 0")
	}
	if input.Status == "" {
		input.Status = domain.TransactionStatusPosted
	}
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now().UTC()
	}

	acc, ok := s.store.GetAccount(input.AccountID)
	if !ok {
		return domain.Transaction{}, errors.New("account not found")
	}
	if acc.WorkspaceID != input.WorkspaceID {
		return domain.Transaction{}, errors.New("account does not belong to workspace")
	}
	if input.Currency == "" {
		input.Currency = acc.Currency
	}

	// For income / expense, keep compatibility with UI by fallback to uncategorized.
	if input.Type == domain.TransactionTypeIncome || input.Type == domain.TransactionTypeExpense {
		if input.CategoryID == "" {
			input.CategoryID = "uncategorized"
		}
	}

	return s.store.CreateTransactionStrict(input)
}

func (s *WealthService) CreateTransfer(input domain.Transfer) (domain.Transfer, error) {
	if input.Amount == "" {
		return domain.Transfer{}, errors.New("amount is required")
	}
	amount, err := parseAmount(input.Amount)
	if err != nil {
		return domain.Transfer{}, errors.New("invalid amount")
	}
	if amount <= 0 {
		return domain.Transfer{}, errors.New("amount must be greater than 0")
	}

	if input.FromAccountID == input.ToAccountID {
		return domain.Transfer{}, errors.New("fromAccountId and toAccountId must be different")
	}

	from, ok := s.store.GetAccount(input.FromAccountID)
	if !ok {
		return domain.Transfer{}, errors.New("source account not found")
	}
	to, ok := s.store.GetAccount(input.ToAccountID)
	if !ok {
		return domain.Transfer{}, errors.New("target account not found")
	}
	if from.WorkspaceID != input.WorkspaceID || to.WorkspaceID != input.WorkspaceID {
		return domain.Transfer{}, errors.New("accounts must belong to workspace")
	}
	if input.Currency == "" {
		input.Currency = from.Currency
	}
	if from.Currency != to.Currency || from.Currency != input.Currency {
		return domain.Transfer{}, errors.New("currency mismatch in transfer")
	}

	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now().UTC()
	}

	transfer, err := s.store.CreateTransfer(input)
	if err != nil {
		return domain.Transfer{}, err
	}

	// Create two opposite transactions for transfer impact. Net-worth remains unchanged.
	_, _ = s.CreateTransaction(domain.Transaction{
		WorkspaceID: input.WorkspaceID,
		AccountID:   from.ID,
		Type:        domain.TransactionTypeTransfer,
		Amount:      input.Amount,
		Currency:    input.Currency,
		Note:        "internal transfer - out: " + input.Note,
		OccurredAt:  input.OccurredAt,
		PortfolioID: from.PortfolioID,
		Status:      domain.TransactionStatusPosted,
		Source:      "transfer",
	})
	_, _ = s.CreateTransaction(domain.Transaction{
		WorkspaceID: input.WorkspaceID,
		AccountID:   to.ID,
		Type:        domain.TransactionTypeTransfer,
		Amount:      input.Amount,
		Currency:    input.Currency,
		Note:        "internal transfer - in: " + input.Note,
		OccurredAt:  input.OccurredAt,
		PortfolioID: to.PortfolioID,
		Status:      domain.TransactionStatusPosted,
		Source:      "transfer",
	})

	return transfer, nil
}

func (s *WealthService) CreateLoanPayment(loanID string, payment domain.LoanPayment) (domain.LoanPayment, error) {
	loan, ok := s.store.GetLoan(domain.ID(loanID))
	if !ok {
		return domain.LoanPayment{}, errors.New("loan not found")
	}

	principal, err := parseAmount(payment.Principal)
	if err != nil {
		return domain.LoanPayment{}, errors.New("invalid principal amount")
	}
	interest := parseAmountWithFallback(payment.Interest)
	fee := parseAmountWithFallback(payment.Fee)
	waived := parseAmountWithFallback(payment.Waived)

	if principal < 0 || interest < 0 || fee < 0 || waived < 0 {
		return domain.LoanPayment{}, errors.New("payment amounts cannot be negative")
	}

	balance, err := parseAmount(loan.PrincipalBalance)
	if err != nil {
		return domain.LoanPayment{}, errors.New("invalid current loan balance")
	}

	if principal > balance {
		return domain.LoanPayment{}, errors.New("principal payment cannot exceed remaining principal")
	}

	payment.LoanID = domain.ID(loanID)
	payment.WorkspaceID = loan.WorkspaceID
	if payment.OccurredAt.IsZero() {
		payment.OccurredAt = time.Now().UTC()
	}

	totalAmount := principal + interest + fee
	if totalAmount <= 0 {
		return domain.LoanPayment{}, errors.New("payment amount is empty")
	}

	nextBalance := balance - principal
	if nextBalance < 0 {
		nextBalance = 0
	}

	// Auto-post a loan payment ledger line when an account is provided or a default account exists.
	postingAccountID := payment.AccountID
	if postingAccountID == "" {
		if accID := s.findFirstAccountID(loan.WorkspaceID, ""); accID != "" {
			postingAccountID = accID
		}
	}

	if postingAccountID != "" {
		if acc, ok := s.store.GetAccount(postingAccountID); ok {
			t, err := s.store.CreateTransaction(domain.Transaction{
				WorkspaceID: loan.WorkspaceID,
				AccountID:   acc.ID,
				PortfolioID: acc.PortfolioID,
				Type:        domain.TransactionTypeLoanPayment,
				Amount:      formatMoney(totalAmount),
				Currency:    currencyOrDefault(acc.Currency, loan.Direction),
				Note:        fmt.Sprintf("loan payment for %s", loanID),
				OccurredAt:  payment.OccurredAt,
				CategoryID:  "",
				Status:      domain.TransactionStatusPosted,
				Source:      "loan_payment",
			})
			if err == nil {
				payment.TransactionID = t.ID
			}
		}
	}

	s.store.UpdateLoan(loan.ID, func(l *domain.Loan) {
		l.PrincipalBalance = formatMoney(nextBalance)
		if nextBalance <= 0 && l.Status != domain.LoanStatusCancelled {
			l.Status = domain.LoanStatusClosed
		}
		l.UpdatedAt = nowUTC()
	})

	return s.store.CreateLoanPayment(payment)
}

func (s *WealthService) CreateLoanPaymentRequest(workspaceID domain.ID, loanID domain.ID, req PaymentRequestCreate) (domain.BankPaymentRequest, error) {
	if loanID == "" {
		return domain.BankPaymentRequest{}, errors.New("loanId is required")
	}
	loan, ok := s.store.GetLoan(loanID)
	if !ok {
		return domain.BankPaymentRequest{}, errors.New("loan not found")
	}
	if loan.WorkspaceID != workspaceID {
		return domain.BankPaymentRequest{}, errors.New("loan does not belong to workspace")
	}
	if loan.Direction != domain.LoanDirectionReceivable {
		return domain.BankPaymentRequest{}, errors.New("loan payment request only allowed for receivable loans")
	}

	amount := strings.TrimSpace(req.Amount)
	if amount == "" {
		amount = loan.PrincipalBalance
	}
	amountValue, err := parseAmount(amount)
	if err != nil || amountValue < 0 {
		return domain.BankPaymentRequest{}, errors.New("invalid payment amount")
	}
	if req.Currency == "" {
		req.Currency = "VND"
	}

	expiresAt := req.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().UTC().Add(7 * 24 * time.Hour)
	}

	reqCode := "WOS-" + strings.ToUpper(uuidString())
	code := reqCode

	created, err := s.store.CreateBankPaymentRequest(domain.BankPaymentRequest{
		WorkspaceID: workspaceID,
		LoanID:      loanID,
		Code:        code,
		Amount:      formatMoney(amountValue),
		Currency:    req.Currency,
		ExpiresAt:   expiresAt,
		Status:      "open",
		Note:        req.Note,
		Source:      "sepay_loan_request",
	})
	return created, err
}

func (s *WealthService) GetPortfolioNetWorth(portfolioID string) (NetWorthResult, error) {
	pID := domain.ID(portfolioID)
	portfolio, ok := s.store.GetPortfolio(pID)
	if !ok {
		return NetWorthResult{}, errors.New("portfolio not found")
	}
	return s.computeNetWorthForPortfolio(portfolio.WorkspaceID, pID)
}

func (s *WealthService) GetPortfolioNetWorthAt(portfolioID string, asOf time.Time) (NetWorthResult, error) {
	pID := domain.ID(portfolioID)
	portfolio, ok := s.store.GetPortfolio(pID)
	if !ok {
		return NetWorthResult{}, errors.New("portfolio not found")
	}
	return s.computeNetWorthForPortfolioAt(portfolio.WorkspaceID, pID, asOf, false)
}

func (s *WealthService) GetPortfolioSnapshots(workspaceID domain.ID, limit int, cursor string) PortfolioSnapshotPage {
	return s.getPortfolioSnapshots(workspaceID, "", limit, cursor)
}

func (s *WealthService) GetPortfolioSnapshotsForPortfolio(workspaceID, portfolioID domain.ID, limit int, cursor string) PortfolioSnapshotPage {
	return s.getPortfolioSnapshots(workspaceID, portfolioID, limit, cursor)
}

func (s *WealthService) getPortfolioSnapshots(workspaceID, portfolioID domain.ID, limit int, cursor string) PortfolioSnapshotPage {
	if limit <= 0 {
		limit = defaultPortfolioSnapshotLimit
	}
	if limit > maxPortfolioSnapshots {
		limit = maxPortfolioSnapshots
	}

	snapshotMu.RLock()
	workspaceHistory := historyByWorkspace[workspaceID]
	history := append([]netWorthSnapshot{}, workspaceHistory[portfolioID]...)
	snapshotMu.RUnlock()
	if len(history) == 0 {
		if _, err := s.computeNetWorthForPortfolio(workspaceID, portfolioID); err != nil {
			return PortfolioSnapshotPage{}
		}
		snapshotMu.RLock()
		workspaceHistory = historyByWorkspace[workspaceID]
		history = append([]netWorthSnapshot{}, workspaceHistory[portfolioID]...)
		snapshotMu.RUnlock()
	}

	start := len(history) - 1
	if cursor != "" {
		cursorAt, err := time.Parse(time.RFC3339Nano, cursor)
		if err != nil {
			return PortfolioSnapshotPage{}
		}
		start = -1
		for i := len(history) - 1; i >= 0; i-- {
			if history[i].AsOfAt.Before(cursorAt) {
				start = i
				break
			}
		}
	}
	if start < 0 {
		return PortfolioSnapshotPage{Items: []NetWorthResult{}, NextCursor: ""}
	}

	items := make([]NetWorthResult, 0, limit)
	for i := start; i >= 0 && len(items) < limit; i-- {
		items = append(items, s.formatNetWorthResult(history[i]))
	}

	if len(items) == 0 {
		return PortfolioSnapshotPage{Items: []NetWorthResult{}}
	}

	nextCursor := ""
	oldestIndex := start - (len(items) - 1)
	if len(items) == limit && oldestIndex > 0 {
		nextCursor = items[len(items)-1].AsOfAt.Format(time.RFC3339Nano)
	}

	return PortfolioSnapshotPage{
		Items:      items,
		NextCursor: nextCursor,
	}
}

func (s *WealthService) ProcessSePayIncoming(raw SePayWebhookEvent) (domain.BankFeedTransaction, error) {
	event := strings.TrimSpace(raw.Direction)
	if event == "" {
		return domain.BankFeedTransaction{}, errors.New("direction is required")
	}
	direction := strings.ToLower(event)
	if direction != bankDirectionIn && direction != bankDirectionOut {
		return domain.BankFeedTransaction{}, errors.New("direction must be in/out")
	}

	amount, err := parseAmount(raw.Amount)
	if err != nil || amount <= 0 {
		return domain.BankFeedTransaction{}, errors.New("amount must be positive")
	}

	workspaceID, err := s.resolveWorkspaceForFeed(raw.ConnectionID)
	if err != nil {
		return domain.BankFeedTransaction{}, err
	}

	occurredAt, err := parseTime(raw.OccurredAt)
	if err != nil {
		return domain.BankFeedTransaction{}, err
	}

	feed, err := s.store.IngestBankFeed(domain.BankFeedTransaction{
		WorkspaceID:  workspaceID,
		ConnectionID: domain.ID(raw.ConnectionID),
		AccountID:    domain.ID(raw.AccountID),
		Amount:       raw.Amount,
		Currency:     firstNonEmpty(raw.Currency, "VND"),
		Direction:    direction,
		CounterParty: raw.Counterparty,
		Description:  strings.TrimSpace(raw.Description),
		Reference:    strings.TrimSpace(raw.Reference),
		OccurredAt:   occurredAt,
		ExternalID:   raw.ExternalID,
		PostedTxnID:  "",
		PostingState: domain.PostingStateAutoReady,
		Confidence:   0,
		Evidence:     "received from sepay webhook",
	})
	if err != nil {
		return domain.BankFeedTransaction{}, err
	}

	return feed, nil
}

func (s *WealthService) EnqueueSePayIncoming(raw SePayWebhookEvent) (domain.BankFeedEvent, error) {
	workspaceID, err := s.resolveWorkspaceForFeed(raw.ConnectionID)
	if err != nil {
		return domain.BankFeedEvent{}, err
	}
	if strings.TrimSpace(raw.Direction) == "" {
		return domain.BankFeedEvent{}, errors.New("direction is required")
	}
	if amount, err := parseAmount(raw.Amount); err != nil || amount <= 0 {
		return domain.BankFeedEvent{}, errors.New("amount must be positive")
	}
	if _, err := parseTime(raw.OccurredAt); err != nil {
		return domain.BankFeedEvent{}, err
	}

	payload, err := json.Marshal(raw)
	if err != nil {
		return domain.BankFeedEvent{}, err
	}

	eventKey := strings.TrimSpace(raw.ExternalID)
	if eventKey == "" {
		eventKey = strings.TrimSpace(raw.ConnectionID + "::" + raw.Direction + "::" + raw.OccurredAt + "::" + raw.Amount + "::" + raw.Reference)
	}

	return s.store.EnqueueBankFeedEvent(domain.BankFeedEvent{
		WorkspaceID:  workspaceID,
		ConnectionID: domain.ID(raw.ConnectionID),
		Provider:     "sepay",
		EventKey:     eventKey,
		ExternalID:   raw.ExternalID,
		State:        domain.BankFeedEventStateQueued,
		Attempts:     0,
		Payload:      string(payload),
	})
}

func (s *WealthService) ProcessBankFeedEvent(eventID domain.ID) error {
	ev, ok := s.store.GetBankFeedEvent(eventID)
	if !ok {
		return errors.New("bank feed event not found")
	}
	if ev.State != "" && ev.State != domain.BankFeedEventStateQueued && ev.State != domain.BankFeedEventStateRunning {
		return nil
	}
	if ev.State == "" {
		ev.State = domain.BankFeedEventStateQueued
	}

	nextAttempt := ev.Attempts + 1
	if ok := s.store.UpdateBankFeedEvent(ev.ID, func(event *domain.BankFeedEvent) {
		event.State = domain.BankFeedEventStateRunning
		event.Attempts = nextAttempt
		event.LastError = ""
	}); !ok {
		return errors.New("failed to mark event running")
	}

	var payload SePayWebhookEvent
	if err := json.Unmarshal([]byte(ev.Payload), &payload); err != nil {
		_ = s.store.UpdateBankFeedEvent(eventID, func(event *domain.BankFeedEvent) {
			event.State = nextFailureState(nextAttempt)
			event.LastError = "invalid event payload"
		})
		return errors.New("invalid event payload")
	}

	_, err := s.ProcessSePayIncoming(payload)
	if err != nil {
		next := nextFailureState(nextAttempt)
		_ = s.store.UpdateBankFeedEvent(eventID, func(event *domain.BankFeedEvent) {
			event.State = next
			event.LastError = err.Error()
		})
		_ = s.store.UpdateBankConnection(ev.ConnectionID, func(conn *domain.BankConnection) {
			conn.SyncStatus = "failed"
		})
		return err
	}

	_ = s.store.UpdateBankFeedEvent(eventID, func(event *domain.BankFeedEvent) {
		event.State = domain.BankFeedEventStateDone
		event.LastError = ""
	})
	_ = s.store.UpdateBankConnection(ev.ConnectionID, func(conn *domain.BankConnection) {
		conn.SyncStatus = "synced"
		conn.LastSyncedAt = nowUTC()
	})
	return nil
}

func nextFailureState(attempt int) string {
	if attempt < maxBankFeedEventAttempts {
		return domain.BankFeedEventStateQueued
	}
	return domain.BankFeedEventStateFailed
}

func (s *WealthService) ProcessQueuedBankFeed(feedID domain.ID) (domain.BankFeedTransaction, error) {
	feed, ok := s.store.GetBankFeed(feedID)
	if !ok {
		return domain.BankFeedTransaction{}, errors.New("bank feed transaction not found")
	}
	if feed.PostingState != domain.PostingStateAutoReady {
		return *feed, nil
	}

	out, err := s.applyBankFeedPolicy(*feed, feed.WorkspaceID, strings.ToLower(feed.Direction))
	if err != nil {
		return out, nil
	}
	return out, nil
}

func (s *WealthService) applyBankFeedPolicy(feed domain.BankFeedTransaction, workspaceID domain.ID, direction string) (domain.BankFeedTransaction, error) {
	if direction == bankDirectionOut {
		return s.applyOutboundRule(feed, workspaceID)
	}
	return s.applyInboundRule(feed, workspaceID)
}

func (s *WealthService) applyOutboundRule(feed domain.BankFeedTransaction, workspaceID domain.ID) (domain.BankFeedTransaction, error) {
	accountID := s.resolveFeedAccount(feed, workspaceID, feed.AccountID)
	if accountID == "" {
		s.updateFeedReason(feed.ID, domain.PostingStateReview, "no account available for outbound bank feed")
		if updated, ok := s.store.GetBankFeed(feed.ID); ok {
			feed = *updated
		}
		return feed, nil
	}

	rule, confidence, reason := s.matchBestRule(workspaceID, bankDirectionOut, feed)
	if rule != nil {
		if strings.EqualFold(rule.ActionType, "transfer") {
			s.updateFeedReason(feed.ID, domain.PostingStateReview, "outbound transfer policy -> review + keep as transfer")
			if updated, ok := s.store.GetBankFeed(feed.ID); ok {
				feed = *updated
			}
			return feed, nil
		}

		txType := domain.TransactionType(rule.Type)
		if _, ok := s.validateTransactionType(txType); !ok {
			txType = domain.TransactionTypeExpense
		}
		if txType != domain.TransactionTypeIncome && txType != domain.TransactionTypeExpense && txType != domain.TransactionTypeValuationAdj {
			txType = domain.TransactionTypeExpense
		}
		reviewReason := reason
		if txType != domain.TransactionTypeExpense {
			reviewReason = "rule actionType is not expense -> fallback to expense for safe handling"
			txType = domain.TransactionTypeExpense
		}
		_, err := s.postBankFeedTransaction(feed, accountID, txType, rule.CategoryID, confidence, reviewReason, true)
		if err == nil {
			s.updateFeedState(feed.ID, domain.PostingStatePosted)
		} else {
			s.updateFeedReason(feed.ID, domain.PostingStateReview, err.Error())
		}
		if updated, ok := s.store.GetBankFeed(feed.ID); ok {
			feed = *updated
		}
		return feed, err
	}

	_, err := s.postBankFeedTransaction(feed, accountID, domain.TransactionTypeExpense, "", 0, "auto expense from outbound transfer", true)
	if err == nil {
		s.updateFeedState(feed.ID, domain.PostingStatePosted)
	} else {
		s.updateFeedReason(feed.ID, domain.PostingStateReview, err.Error())
	}
	if updated, ok := s.store.GetBankFeed(feed.ID); ok {
		feed = *updated
	}
	return feed, err
}

func (s *WealthService) applyInboundRule(feed domain.BankFeedTransaction, workspaceID domain.ID) (domain.BankFeedTransaction, error) {
	// 1) Match explicit VietQR / loan payment request code first.
	if paymentReq := s.findMatchingPaymentRequest(workspaceID, feed); paymentReq != nil {
		paymentAmount, _ := parseAmount(feed.Amount)
		payment, err := s.CreateLoanPayment(string(paymentReq.LoanID), domain.LoanPayment{
			Principal:  formatMoney(paymentAmount),
			Interest:   "0",
			Fee:        "0",
			Waived:     "0",
			OccurredAt: feed.OccurredAt,
		})
		if err == nil {
			_ = payment
			if payment.TransactionID != "" {
				s.store.LinkBankFeedPosting(feed.ID, payment.TransactionID)
			}
			s.updateFeedState(feed.ID, domain.PostingStatePosted)
			s.store.UpdateFeed(feed.ID, func(f *domain.BankFeedTransaction) {
				f.RuleID = paymentReq.ID
				f.Evidence = "matched loan payment request"
				f.Confidence = 100
				f.PostingState = domain.PostingStatePosted
				f.UpdatedAt = nowUTC()
			})
			if updated, ok := s.store.GetBankFeed(feed.ID); ok {
				return *updated, nil
			}
			return feed, nil
		}
	}

	// 2) Apply user rules.
	rule, confidence, reason := s.matchBestRule(workspaceID, bankDirectionIn, feed)
	if rule != nil {
		ruleType := firstNonEmpty(rule.ActionType, rule.Type)
		txType := domain.TransactionType(ruleType)
		if _, ok := s.validateTransactionType(txType); !ok {
			txType = ""
		}
		if txType == "" {
			s.updateFeedReason(feed.ID, domain.PostingStateReview, "unsupported rule action type")
			if updated, ok := s.store.GetBankFeed(feed.ID); ok {
				feed = *updated
			}
			return feed, nil
		}
		if txType == domain.TransactionTypeTransfer || strings.EqualFold(rule.ActionType, "transfer") {
			s.updateFeedReason(feed.ID, domain.PostingStateReview, "rule detected possible inbound transfer; manual confirm required")
			if updated, ok := s.store.GetBankFeed(feed.ID); ok {
				feed = *updated
			}
			return feed, nil
		}
		if txType == domain.TransactionTypeLoanPayment {
			s.updateFeedReason(feed.ID, domain.PostingStateReview, "rule suggests loan payment; confirm allocation split before posting")
			if updated, ok := s.store.GetBankFeed(feed.ID); ok {
				feed = *updated
			}
			return feed, nil
		}
		if confidence < sepayAutoPostThreshold {
			s.updateFeedReason(feed.ID, domain.PostingStateReview, reason)
			if updated, ok := s.store.GetBankFeed(feed.ID); ok {
				feed = *updated
			}
			return feed, nil
		}

		if confidence >= sepayAutoPostThreshold {
			// inbound auto-post with confidence-based policy
			accountID := s.resolveFeedAccount(feed, workspaceID, feed.AccountID)
			if accountID != "" {
				_, err := s.postBankFeedTransaction(feed, accountID, txType, rule.CategoryID, confidence, reason, true)
				if err == nil {
					s.updateFeedState(feed.ID, domain.PostingStatePosted)
					if updated, ok := s.store.GetBankFeed(feed.ID); ok {
						feed = *updated
					}
					return feed, nil
				}
			}
			s.updateFeedReason(feed.ID, domain.PostingStateReview, "no account found for inbound bank feed")
			if updated, ok := s.store.GetBankFeed(feed.ID); ok {
				feed = *updated
			}
			return feed, nil
		}
		s.updateFeedReason(feed.ID, domain.PostingStateReview, reason)
		if updated, ok := s.store.GetBankFeed(feed.ID); ok {
			feed = *updated
		}
		return feed, nil
	}

	// 3) Heuristic confidence by content.
	classification := s.classifyInboundHeuristic(feed)
	switch classification.txType {
	case domain.TransactionTypeTransfer:
		s.updateFeedReason(feed.ID, domain.PostingStateReview, classification.reason)
		if updated, ok := s.store.GetBankFeed(feed.ID); ok {
			feed = *updated
		}
		return feed, nil
	case domain.TransactionTypeIncome:
		accountID := s.resolveFeedAccount(feed, workspaceID, feed.AccountID)
		if accountID == "" {
			s.updateFeedReason(feed.ID, domain.PostingStateReview, "no account found for inbound bank feed")
			if updated, ok := s.store.GetBankFeed(feed.ID); ok {
				feed = *updated
			}
			return feed, nil
		}
		if classification.confidence >= sepayAutoPostThreshold {
			_, err := s.postBankFeedTransaction(feed, accountID, classification.txType, classification.categoryID, classification.confidence, classification.reason, classification.autoClassed)
			if err == nil {
				s.updateFeedState(feed.ID, domain.PostingStatePosted)
				if updated, ok := s.store.GetBankFeed(feed.ID); ok {
					feed = *updated
				}
				return feed, nil
			}
		}
		s.updateFeedReason(feed.ID, domain.PostingStateReview, classification.reason)
		if updated, ok := s.store.GetBankFeed(feed.ID); ok {
			feed = *updated
		}
		return feed, nil
	default:
		s.updateFeedReason(feed.ID, domain.PostingStateReview, classification.reason)
		if updated, ok := s.store.GetBankFeed(feed.ID); ok {
			feed = *updated
		}
		return feed, nil
	}
}

func (s *WealthService) postBankFeedTransaction(feed domain.BankFeedTransaction, accountID domain.ID, txType domain.TransactionType, categoryID domain.ID, confidence float64, reason string, auto bool) (domain.Transaction, error) {
	if accountID == "" {
		return domain.Transaction{}, errors.New("missing posting account")
	}
	acc, ok := s.store.GetAccount(accountID)
	if !ok {
		return domain.Transaction{}, errors.New("posting account not found")
	}
	tx, err := s.store.CreateTransaction(domain.Transaction{
		WorkspaceID: acc.WorkspaceID,
		AccountID:   accountID,
		CategoryID:  categoryID,
		Type:        txType,
		Amount:      feed.Amount,
		Currency:    feed.Currency,
		Note:        feed.Description,
		OccurredAt:  feed.OccurredAt,
		Status:      domain.TransactionStatusPosted,
		Source:      "bank_feed",
	})
	if err != nil {
		return domain.Transaction{}, err
	}
	_ = s.store.LinkBankFeedPosting(feed.ID, tx.ID)
	_ = confidence
	_ = auto
	s.store.UpdateFeed(feed.ID, func(f *domain.BankFeedTransaction) {
		f.Confidence = confidence
		f.Evidence = reason
		f.AutoClassified = auto
		f.PostedTxnID = tx.ID
		f.PostingState = domain.PostingStatePosted
	})
	return tx, nil
}

func (s *WealthService) matchBestRule(workspaceID domain.ID, direction string, feed domain.BankFeedTransaction) (*domain.AutomationRule, float64, string) {
	rules := s.store.GetWorkspaceRules(workspaceID, feed.AccountID, direction)
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		if r.Direction != "" && r.Direction != direction {
			continue
		}
		if feed.Currency != "" && r.MinAmount != "" {
			min, err := parseAmount(r.MinAmount)
			if err == nil {
				got, _ := parseAmount(feed.Amount)
				if got < min {
					continue
				}
			}
		}
		if feed.Currency != "" && r.MaxAmount != "" {
			max, err := parseAmount(r.MaxAmount)
			if err == nil {
				got, _ := parseAmount(feed.Amount)
				if got > max {
					continue
				}
			}
		}
		content := strings.ToLower(feed.Description + " " + feed.Reference + " " + feed.CounterParty + " " + feed.Evidence)
		if r.ContentPattern != "" {
			rx, err := regexp.Compile("(?i)" + regexp.QuoteMeta(r.ContentPattern))
			if err == nil && !rx.MatchString(content) {
				continue
			}
		}
		if r.ReferencePattern != "" {
			rx, err := regexp.Compile("(?i)" + regexp.QuoteMeta(r.ReferencePattern))
			if err == nil && !rx.MatchString(content) {
				continue
			}
		}
		conf := 80.0
		if r.Predicate != "" {
			conf += 10
		}
		return &r, conf, "rule matched: " + r.Name
	}
	return nil, 0, ""
}

func (s *WealthService) resolveWorkspaceForFeed(connectionID string) (domain.ID, error) {
	if connectionID == "" {
		return "", errors.New("connectionId is required")
	}
	conn, ok := s.store.GetBankConnection(domain.ID(connectionID))
	if !ok {
		return "", errors.New("connection not found")
	}
	return conn.WorkspaceID, nil
}

func (s *WealthService) resolveFeedAccount(feed domain.BankFeedTransaction, workspaceID domain.ID, fallback domain.ID) domain.ID {
	if fallback != "" {
		acc, ok := s.store.GetAccount(fallback)
		if ok && acc.WorkspaceID == workspaceID {
			return acc.ID
		}
	}
	return s.findFirstAccountID(workspaceID, feed.Currency)
}

func (s *WealthService) classifyInboundHeuristic(feed domain.BankFeedTransaction) bankFeedClassification {
	text := normalizeSePayText(feed.Description + " " + feed.Reference + " " + feed.CounterParty)
	if text == "" {
		return bankFeedClassification{
			txType:      "",
			confidence:  0,
			reason:      "empty note/reference",
			autoClassed: false,
		}
	}

	if containsAny(text, inboundIncomeKeywords) {
		return bankFeedClassification{
			txType:      domain.TransactionTypeIncome,
			confidence:  90,
			reason:      "matched income keyword",
			autoClassed: true,
		}
	}
	if containsAny(text, inboundTransferKeywords) {
		return bankFeedClassification{
			txType:      domain.TransactionTypeTransfer,
			confidence:  85,
			reason:      "matched transfer keyword; requires manual review",
			autoClassed: false,
		}
	}
	return bankFeedClassification{
		txType:      "",
		confidence:  0,
		reason:      "cannot infer inbound transaction type safely",
		autoClassed: false,
	}
}

func containsAny(text string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(text, normalizeSePayText(n)) {
			return true
		}
	}
	return false
}

func normalizeSePayText(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func (s *WealthService) updateFeedState(feedID domain.ID, state domain.TransactionPostingState) {
	_ = s.store.UpdateFeed(feedID, func(f *domain.BankFeedTransaction) {
		f.PostingState = state
		f.UpdatedAt = nowUTC()
	})
}

func (s *WealthService) updateFeedReason(feedID domain.ID, state domain.TransactionPostingState, reason string) {
	_ = s.store.UpdateFeed(feedID, func(f *domain.BankFeedTransaction) {
		f.PostingState = state
		f.Evidence = reason
		f.UpdatedAt = nowUTC()
	})
}

func (s *WealthService) findMatchingPaymentRequest(workspaceID domain.ID, feed domain.BankFeedTransaction) *domain.BankPaymentRequest {
	text := strings.ToLower(feed.Description + " " + feed.Reference)
	pattern := regexp.MustCompile(`(?i)WOS-[A-Za-z0-9-]+`)
	matches := pattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}
	for _, code := range matches {
		if req, ok := s.store.GetBankPaymentRequestByCode(workspaceID, code); ok {
			if req.Status != "open" {
				continue
			}
			if req.ExpiresAt.Before(nowUTC()) {
				continue
			}
			return req
		}
	}
	return nil
}

func (s *WealthService) RulePreview(workspaceID domain.ID, sample []domain.BankFeedTransaction) map[string]any {
	estimated := map[string]int{
		"matched":      0,
		"auto_posted":  0,
		"needs_review": 0,
	}
	samples := []domain.BankFeedTransaction{}
	for i, item := range sample {
		if i > 9 {
			break
		}
		class := s.classifyInboundHeuristic(item)
		if class.txType != "" {
			estimated["matched"]++
			if class.autoClassed {
				estimated["auto_posted"]++
			} else {
				estimated["needs_review"]++
			}
			samples = append(samples, item)
			continue
		}
		rule, _, _ := s.matchBestRule(workspaceID, item.Direction, item)
		if rule != nil {
			estimated["matched"]++
			samples = append(samples, item)
		}
	}
	return map[string]any{
		"estimatedAffected": len(samples),
		"samples":           samples,
		"counts":            estimated,
	}
}

func (s *WealthService) ApproveBankFeed(id domain.ID, feed domain.BankFeedTransaction) (domain.Transaction, error) {
	if id == "" {
		return domain.Transaction{}, errors.New("feed id is required")
	}

	if existing, ok := s.store.GetTransaction(feed.PostedTxnID); ok && existing.ID != "" {
		return *existing, nil
	}

	accountID := feed.AccountID
	if accountID == "" {
		accountID = s.resolveFeedAccount(feed, feed.WorkspaceID, "")
	}
	if accountID == "" {
		return domain.Transaction{}, errors.New("missing account for posting bank feed")
	}

	txType := domain.TransactionTypeExpense
	if strings.ToLower(feed.Direction) == bankDirectionIn {
		txType = domain.TransactionTypeIncome
	}
	if feed.PostingState == domain.PostingStatePosted && feed.PostedTxnID != "" {
		// Idempotent behavior: return already created ledger entry.
		if txn, ok := s.store.GetTransaction(feed.PostedTxnID); ok {
			return *txn, nil
		}
	}

	confidence := feed.Confidence
	evidence := firstNonEmpty(feed.Evidence, "manual approval")
	categoryID := domain.ID("uncategorized")

	tx, err := s.postBankFeedTransaction(feed, accountID, txType, categoryID, confidence, evidence, false)
	if err != nil {
		return domain.Transaction{}, err
	}
	_ = s.store.UpdateFeed(feed.ID, func(item *domain.BankFeedTransaction) {
		item.PostingState = domain.PostingStatePosted
		item.Evidence = evidence
		item.AutoClassified = false
		item.Confidence = confidence
		item.PostedTxnID = tx.ID
		item.RuleID = ""
	})
	return tx, nil
}

func (s *WealthService) ReclassifyBankFeed(id domain.ID, accountID domain.ID, txType domain.TransactionType, categoryID domain.ID, reason string) (domain.Transaction, error) {
	if id == "" {
		return domain.Transaction{}, errors.New("feed id is required")
	}
	feed, ok := s.store.GetBankFeed(id)
	if !ok {
		return domain.Transaction{}, errors.New("bank feed transaction not found")
	}
	if _, ok := s.validateTransactionType(txType); !ok {
		return domain.Transaction{}, errors.New("invalid transaction type")
	}
	if accountID == "" {
		accountID = s.resolveFeedAccount(*feed, feed.WorkspaceID, feed.AccountID)
	}
	if accountID == "" {
		return domain.Transaction{}, errors.New("missing account for posting")
	}
	if categoryID == "" {
		categoryID = "uncategorized"
	}

	tx, err := s.postBankFeedTransaction(*feed, accountID, txType, categoryID, feed.Confidence, reason, false)
	if err != nil {
		return domain.Transaction{}, err
	}

	_ = s.store.UpdateFeed(feed.ID, func(item *domain.BankFeedTransaction) {
		item.PostingState = domain.PostingStatePosted
		item.RuleID = ""
		item.Evidence = firstNonEmpty(reason, item.Evidence)
		item.AutoClassified = false
	})
	return tx, nil
}

func (s *WealthService) SeedDemoData(seedUserID domain.ID, workspaceID domain.ID) {
	ws, ok := s.store.GetWorkspace(workspaceID)
	if !ok || ws == nil {
		return
	}

	p, ok := s.store.FirstPortfolio(workspaceID)
	if !ok {
		return
	}

	now := nowUTC()
	accounts := s.store.ListAccounts(workspaceID)

	mainAccountID := domain.ID("")
	savingsAccountID := domain.ID("")
	for _, account := range accounts {
		switch account.Name {
		case "Ví tiền":
			mainAccountID = account.ID
		case "Tài khoản tiết kiệm":
			savingsAccountID = account.ID
		}
	}

	if mainAccountID == "" {
		acc, err := s.store.CreateAccount(domain.Account{
			WorkspaceID: workspaceID,
			PortfolioID: p.ID,
			Name:        "Ví tiền",
			Type:        "cash",
			Currency:    ws.BaseCurrency,
		})
		if err == nil {
			mainAccountID = acc.ID
		}
	}
	if savingsAccountID == "" {
		acc, err := s.store.CreateAccount(domain.Account{
			WorkspaceID: workspaceID,
			PortfolioID: p.ID,
			Name:        "Tài khoản tiết kiệm",
			Type:        "savings",
			Currency:    ws.BaseCurrency,
		})
		if err == nil {
			savingsAccountID = acc.ID
		}
	}
	if mainAccountID == "" {
		return
	}

	if len(s.store.ListTransactions(workspaceID, "")) == 0 {
		_, _ = s.CreateTransaction(domain.Transaction{
			WorkspaceID: workspaceID,
			AccountID:   mainAccountID,
			PortfolioID: p.ID,
			Type:        domain.TransactionTypeIncome,
			Amount:      "12000000",
			Currency:    ws.BaseCurrency,
			Note:        "Lương tháng",
			Status:      domain.TransactionStatusPosted,
			OccurredAt:  now.AddDate(0, 0, -18),
		})
		_, _ = s.CreateTransaction(domain.Transaction{
			WorkspaceID: workspaceID,
			AccountID:   mainAccountID,
			PortfolioID: p.ID,
			Type:        domain.TransactionTypeExpense,
			Amount:      "3000000",
			Currency:    ws.BaseCurrency,
			Note:        "Tiêu vặt",
			Status:      domain.TransactionStatusPosted,
			OccurredAt:  now.AddDate(0, 0, -15),
		})
		_, _ = s.CreateTransaction(domain.Transaction{
			WorkspaceID: workspaceID,
			AccountID:   mainAccountID,
			PortfolioID: p.ID,
			Type:        domain.TransactionTypeExpense,
			Amount:      "5000000",
			Currency:    ws.BaseCurrency,
			Note:        "Trả tiền nhà",
			Status:      domain.TransactionStatusPosted,
			OccurredAt:  now.AddDate(0, 0, -7),
		})
	}

	if len(s.store.ListTransactions(workspaceID, "")) > 0 && savingsAccountID != "" {
		_ = mainAccountID
		// Keep transfer explicit to show double-entry transfer behavior on dashboard.
		_, _ = s.CreateTransfer(domain.Transfer{
			WorkspaceID:   workspaceID,
			FromAccountID: mainAccountID,
			ToAccountID:   savingsAccountID,
			Amount:        "2500000",
			Currency:      ws.BaseCurrency,
			Note:          "Chuyển quỹ dự phòng",
			OccurredAt:    now.AddDate(0, 0, -2),
		})
	}

	if len(s.store.ListProperties(workspaceID)) == 0 {
		prop, err := s.store.CreateProperty(domain.Property{
			WorkspaceID: workspaceID,
			PortfolioID: p.ID,
			Name:        "Căn hộ FPT",
			Address:     "TP.HCM",
			AreaM2:      "85",
			PurchaseAt:  now.AddDate(-2, 0, 0),
		})
		if err == nil {
			_, _ = s.store.AddPropertyValuation(domain.PropertyValuation{
				PropertyID:  prop.ID,
				Amount:      "2500000000",
				Currency:    ws.BaseCurrency,
				Source:      "initial appraisal",
				EffectiveAt: now,
			})
		}
	}

	if len(s.store.ListAssets(workspaceID)) == 0 {
		asset, err := s.store.CreateAsset(domain.Asset{
			WorkspaceID: workspaceID,
			PortfolioID: p.ID,
			Name:        "Quỹ ETF Nhiều tài sản",
			Type:        "investment_fund",
		})
		if err == nil {
			_, _ = s.store.AddAssetValuation(domain.AssetValuation{
				AssetID:     asset.ID,
				Amount:      "60000000",
				Currency:    ws.BaseCurrency,
				Source:      "initial valuation",
				EffectiveAt: now.AddDate(0, -1, 0),
			})
		}
	}

	if len(s.store.ListLoans(workspaceID)) == 0 {
		payable, err := s.store.CreateLoan(domain.Loan{
			WorkspaceID:      workspaceID,
			PortfolioID:      p.ID,
			Counterparty:     "Ngân hàng ABC",
			Direction:        domain.LoanDirectionPayable,
			PrincipalInitial: "50000000",
			PrincipalBalance: "50000000",
			AnnualRate:       "9",
			DayCountBasis:    "365",
			StartAt:          now.AddDate(0, -4, 0),
			DueAt:            now.AddDate(1, 0, 0),
			Status:           domain.LoanStatusActive,
			InterestCompound: true,
		})
		if err == nil {
			_, _ = s.CreateLoanPayment(string(payable.ID), domain.LoanPayment{
				Principal:  "5000000",
				Interest:   "300000",
				Fee:        "0",
				Waived:     "0",
				OccurredAt: now.AddDate(0, 0, -1),
			})
		}

		_, _ = s.store.CreateLoan(domain.Loan{
			WorkspaceID:      workspaceID,
			PortfolioID:      p.ID,
			Counterparty:     "Anh Minh",
			Direction:        domain.LoanDirectionReceivable,
			PrincipalInitial: "15000000",
			PrincipalBalance: "15000000",
			AnnualRate:       "8",
			DayCountBasis:    "365",
			StartAt:          now.AddDate(0, -1, 0),
			DueAt:            now.AddDate(0, 3, 0),
			Status:           domain.LoanStatusActive,
			InterestCompound: false,
		})
	}

	period := now.Format("2006-01")
	if period == "" {
		period = now.Format("2006-01")
	}
	_, _ = s.UpsertBudget(workspaceID, period, "uncategorized", "4000000")

	if len(s.store.ListForecastScenarios(workspaceID)) == 0 {
		_, _ = s.store.CreateForecastScenario(domain.ForecastScenario{
			WorkspaceID: workspaceID,
			Name:        "Kịch bản tiết kiệm 12 tháng",
			Assumptions: "{\"inflation\": 0.05, \"targetGrowth\": 0.08, \"timeHorizonMonths\": 12}",
			Status:      "draft",
		})
	}

	if len(s.store.ListBankConnections(workspaceID)) == 0 {
		conn, err := s.store.CreateBankConnection(domain.BankConnection{
			WorkspaceID: workspaceID,
			Provider:    "sepay",
			Status:      "connected",
			Scope:       "read",
			BankCode:    "VCB",
		})
		if err == nil {
			_, _ = s.ProcessSePayIncoming(SePayWebhookEvent{
				ConnectionID: string(conn.ID),
				Direction:    "in",
				Amount:       "1200000",
				Currency:     ws.BaseCurrency,
				Description:  "Luong",
				Reference:    "DEMO-01",
				ExternalID:   "seed-in-001",
				OccurredAt:   now.AddDate(0, 0, -3).Format(time.RFC3339),
			})
			_, _ = s.ProcessSePayIncoming(SePayWebhookEvent{
				ConnectionID: string(conn.ID),
				Direction:    "in",
				Amount:       "250000",
				Currency:     ws.BaseCurrency,
				Description:  "chuyen tien cho ban",
				Reference:    "DEMO-02",
				ExternalID:   "seed-in-review-002",
				OccurredAt:   now.AddDate(0, 0, -1).Format(time.RFC3339),
			})
		}
	}

	if len(s.store.ListWorkspaceRules(workspaceID)) == 0 {
		_, _ = s.store.CreateAutomationRule(domain.AutomationRule{
			WorkspaceID: workspaceID,
			AccountID:   mainAccountID,
			Name:        "Auto-tag Lương",
			Priority:    10,
			Predicate:   "contains(description,luong)",
			Direction:   "in",
			ActionType:  "classify",
			Type:        "income",
			CategoryID:  "income",
			Enabled:     true,
		})
	}

	if len(s.store.ListAssistantCommands(workspaceID)) == 0 {
		_, _ = s.store.CreateAssistantCommand(domain.AssistantCommand{
			WorkspaceID: workspaceID,
			UserID:      seedUserID,
			Command:     "Khởi tạo mục tiêu tài chính: tiết kiệm cho quỹ dự phòng",
			Status:      "ready",
			Plan:        "Kiểm tra lại danh mục đầu tư, điều chỉnh kế hoạch chi tiêu theo tháng.",
		})
	}
}

func (s *WealthService) DashboardKpis(workspaceID domain.ID, top int) ([]domain.Transaction, error) {
	txs := s.store.ListTransactions(workspaceID, "")
	sort.Slice(txs, func(i, j int) bool { return txs[i].OccurredAt.After(txs[j].OccurredAt) })
	if top <= 0 || len(txs) <= top {
		return txs, nil
	}
	return txs[:top], nil
}

func (s *WealthService) validateTransactionType(t domain.TransactionType) (string, bool) {
	switch t {
	case domain.TransactionTypeIncome, domain.TransactionTypeExpense, domain.TransactionTypeTransfer, domain.TransactionTypeInvestment,
		domain.TransactionTypeLoanDisbursement, domain.TransactionTypeLoanPayment, domain.TransactionTypeValuationAdj:
		return string(t), true
	default:
		return string(t), false
	}
}

func (s *WealthService) validateTransactionStatus(status domain.TransactionStatus) (string, bool) {
	switch status {
	case "", domain.TransactionStatusDraft, domain.TransactionStatusPending, domain.TransactionStatusPosted, domain.TransactionStatusReconciled, domain.TransactionStatusVoided:
		return string(status), true
	default:
		return string(status), false
	}
}

func (s *WealthService) findFirstAccountID(workspaceID domain.ID, currency string) domain.ID {
	accounts := s.store.ListAccounts(workspaceID)
	for _, acc := range accounts {
		if currency == "" || acc.Currency == currency {
			return acc.ID
		}
	}
	if len(accounts) > 0 {
		return accounts[0].ID
	}
	return ""
}

func parseAmount(s string) (float64, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, errors.New("invalid amount")
	}
	return f, nil
}

func parseAmountWithFallback(s string) float64 {
	f, err := parseAmount(s)
	if err != nil {
		return 0
	}
	return f
}

func parseTime(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, errors.New("occurredAt is required")
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
		return t.UTC(), nil
	}
	if unix, err := strconv.ParseInt(v, 10, 64); err == nil {
		return time.Unix(unix, 0).UTC(), nil
	}
	return time.Time{}, errors.New("invalid occurredAt format")
}

func firstNonEmpty(values ...string) string {
	for _, item := range values {
		item = strings.TrimSpace(item)
		if item != "" {
			return item
		}
	}
	return ""
}

func uuidString() string {
	return uuid.NewString()
}

func formatMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

func currencyOrDefault(currency string, direction domain.LoanDirection) string {
	if currency != "" {
		return currency
	}
	_ = direction
	return "VND"
}

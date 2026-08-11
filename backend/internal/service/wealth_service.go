package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/mail"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"wealthos-backend/internal/cache"
	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/storage"
)

type WealthService struct {
	store storage.Store
	cache cache.CacheService
}

type LoginResult struct {
	User  domain.User `json:"user"`
	Token string      `json:"token"`
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
	ConnectionID      string `json:"connectionId"`
	ProviderAccountID string `json:"providerAccountId"`
	AccountID         string `json:"accountId"`
	Direction         string `json:"direction"`
	Amount            string `json:"amount"`
	Currency          string `json:"currency"`
	Counterparty      string `json:"counterparty"`
	Description       string `json:"description"`
	Reference         string `json:"reference"`
	Content           string `json:"content"`
	ExternalID        string `json:"externalTransactionId"`
	OccurredAt        string `json:"occurredAt"`
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
	// historyByUser stores bounded net-worth snapshots per user and optional portfolio scope.
	historyByUser   = map[domain.ID]map[domain.ID][]netWorthSnapshot{}
	snapshotVersion = map[domain.ID]map[domain.ID]int{}
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
	UserID               string           `json:"userId"`
	AsOfAt               time.Time        `json:"asOfAt"`
	Status               string           `json:"status"`
	Currency             string           `json:"currency"`
	PrincipalInitial     string           `json:"principalInitial"`
	PrincipalBalance     string           `json:"principalBalance"`
	AnnualRate           string           `json:"annualRate"`
	DailyRatePerMillion  string           `json:"dailyRatePerMillion"`
	DailyInterest        string           `json:"dailyInterest"`
	LastInterestPaidDate time.Time        `json:"lastInterestPaidDate,omitempty"`
	NextPaymentDate      time.Time        `json:"nextPaymentDate"`
	DayCountBasis        string           `json:"dayCountBasis"`
	TotalAccruedInterest string           `json:"totalAccruedInterest"`
	TotalPaidInterest    string           `json:"totalPaidInterest"`
	UnpaidInterest       string           `json:"unpaidInterest"`
	Rows                 []LoanAccrualRow `json:"accruals"`
}

type LoanPortfolioSummary struct {
	ActivePrincipal string `json:"activePrincipal"`
	DailyInterest   string `json:"dailyInterest"`
	AccruedInterest string `json:"accruedInterest"`
	PaidInterest    string `json:"paidInterest"`
}

type LoanScheduleItem struct {
	LoanID           string    `json:"loanId"`
	Borrower         string    `json:"borrower"`
	PaymentDate      time.Time `json:"paymentDate"`
	CycleDays        int       `json:"cycleDays"`
	ExpectedInterest string    `json:"expectedInterest"`
	Status           string    `json:"status"`
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
	Period string      `json:"period"`
	UserID string      `json:"userId"`
	AsOfAt time.Time   `json:"asOfAt"`
	Rows   []BudgetRow `json:"rows"`
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

func (s *WealthService) RegisterUser(email, password, confirmation, name string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	if email == "" || password == "" || name == "" {
		return domain.User{}, errors.New("email, password, name are required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return domain.User{}, errors.New("email is invalid")
	}
	if len(password) < 8 {
		return domain.User{}, errors.New("password must contain at least 8 characters")
	}
	if password != confirmation {
		return domain.User{}, errors.New("password confirmation does not match")
	}
	if _, exists := s.store.GetUserByEmail(email); exists {
		return domain.User{}, errors.New("email already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, errors.New("cannot securely store password")
	}
	return s.store.CreateUser(domain.User{Email: email, Name: name, Password: string(hash)})
}

func (s *WealthService) Authenticate(email, password string) (LoginResult, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	u, ok := s.store.GetUserByEmail(email)
	if !ok || !matchesPassword(u.Password, password) {
		return LoginResult{}, errors.New("invalid credentials")
	}
	if !u.IsEmailVerified() {
		return LoginResult{}, errors.New("email verification required")
	}
	token := "token-" + string(u.ID)
	return LoginResult{User: *u, Token: token}, nil
}

func matchesPassword(stored, provided string) bool {
	if strings.HasPrefix(stored, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(provided)) == nil
	}
	// Existing demo and legacy records predate password hashing. New accounts
	// are always stored as bcrypt hashes.
	return stored == provided
}

func (s *WealthService) CurrentNetWorth(userID string) (NetWorthResult, error) {
	return s.ComputeNetWorth(domain.ID(userID))
}

func (s *WealthService) ComputeNetWorth(userID domain.ID) (NetWorthResult, error) {
	return s.ComputeNetWorthAt(userID, time.Time{})
}

func (s *WealthService) ComputeNetWorthAt(userID domain.ID, asOf time.Time) (NetWorthResult, error) {
	return s.computeNetWorthForPortfolioAt(userID, "", asOf, true)
}

func (s *WealthService) computeNetWorthForPortfolio(userID, portfolioID domain.ID) (NetWorthResult, error) {
	return s.computeNetWorthForPortfolioAt(userID, portfolioID, time.Time{}, true)
}

func (s *WealthService) computeNetWorthForPortfolioAt(userID, portfolioID domain.ID, asOf time.Time, persist bool) (NetWorthResult, error) {
	ws, ok := s.store.GetUser(userID)
	if !ok {
		return NetWorthResult{}, errors.New("user not found")
	}

	if asOf.IsZero() {
		asOf = nowUTC()
	}

	cash := s.computeCash(userID, portfolioID, asOf)
	liabilities, accruedPayableInterest := s.computeLoanLiabilityExposure(userID, portfolioID, asOf)
	receivables, accruedReceivableInterest := s.computeLoanReceivableExposure(userID, portfolioID, asOf)
	property, otherAssets, staleValuations := s.computeAssetValuationUser(userID, portfolioID, asOf)
	valuationAssets := property + otherAssets
	accruedInterestNet := accruedReceivableInterest - accruedPayableInterest
	reconciledAccounts := len(s.store.ListAccounts(userID))
	if portfolioID != "" {
		reconciledAccounts = len(s.portfolioAccounts(userID, portfolioID))
	}

	// Unpaid interest is shown separately as accrued interest. It is not cash
	// or settled principal yet, so it must not inflate the headline net worth.
	net := cash + valuationAssets + receivables - liabilities
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

	prev, hasPrev := s.snapshotAtOrBefore(userID, portfolioID, asOf)
	if hasPrev {
		current.NetWorthChange = current.NetWorth - prev.NetWorth
		current.ExternalCashFlow = s.computeExternalCashFlow(current, prev)
		current.ValuationChange = (current.Property + current.OtherAssets) - (prev.Property + prev.OtherAssets)
		current.AccruedInterestDelta = current.AccruedInterest - prev.AccruedInterest
	} else {
		current.NetWorthChange = 0
	}

	if persist {
		recorded := s.addNetWorthSnapshot(userID, portfolioID, current)
		return s.formatNetWorthResult(recorded), nil
	}

	return s.formatNetWorthResult(current), nil
}

func (s *WealthService) portfolioAccounts(userID, portfolioID domain.ID) []domain.Account {
	accounts := s.store.ListAccounts(userID)
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

func (s *WealthService) computeCash(userID, portfolioID domain.ID, asOf time.Time) float64 {
	txs := s.store.ListTransactions(userID, "")
	accounts := s.portfolioAccounts(userID, portfolioID)
	accountByID := map[domain.ID]domain.Account{}
	for _, account := range accounts {
		accountByID[account.ID] = account
	}
	cashByAccount := map[domain.ID]float64{}
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
			cashByAccount[t.AccountID] += amt
		case domain.TransactionTypeExpense, domain.TransactionTypeInvestment, domain.TransactionTypeLoanDisbursement:
			cashByAccount[t.AccountID] -= amt
		case domain.TransactionTypeLoanPayment:
			// Keep loan payment sign aligned with loan direction when possible.
			sign := 1.0
			if paymentSign, ok := s.loanPaymentSign(userID, portfolioID, t); ok {
				sign = paymentSign
			}
			cashByAccount[t.AccountID] += sign * amt
		case domain.TransactionTypeTransfer, domain.TransactionTypeValuationAdj:
			// ignored in cash rollup
		default:
			// ignore unknown types
		}
	}
	cash := 0.0
	for _, account := range accounts {
		if account.BalanceOverride != "" && (account.BalanceOverrideAt.IsZero() || !asOf.Before(account.BalanceOverrideAt)) {
			cash += parseAmountWithFallback(account.BalanceOverride)
			continue
		}
		cash += cashByAccount[account.ID]
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

func (s *WealthService) snapshotAtOrBefore(userID, portfolioID domain.ID, asOf time.Time) (netWorthSnapshot, bool) {
	snapshotMu.RLock()
	defer snapshotMu.RUnlock()
	scopeHistory, ok := historyByUser[userID]
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

func (s *WealthService) addNetWorthSnapshot(userID, portfolioID domain.ID, snapshot netWorthSnapshot) netWorthSnapshot {
	snapshotMu.Lock()
	defer snapshotMu.Unlock()

	scopeVersions, ok := snapshotVersion[userID]
	if !ok {
		scopeVersions = map[domain.ID]int{}
		snapshotVersion[userID] = scopeVersions
	}

	nextVersion := scopeVersions[portfolioID] + 1
	scopeVersions[portfolioID] = nextVersion
	snapshot.SnapshotVersion = nextVersion

	scopeHistory, ok := historyByUser[userID]
	if !ok {
		scopeHistory = map[domain.ID][]netWorthSnapshot{}
		historyByUser[userID] = scopeHistory
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

func (s *WealthService) loanPaymentSign(userID, portfolioID domain.ID, t domain.Transaction) (float64, bool) {
	loans := s.store.ListLoans(userID)
	for _, ln := range loans {
		if !s.matchesEntityToPortfolio(portfolioID, ln.PortfolioID) {
			continue
		}
		pp := s.store.ListLoanPayments(userID, ln.ID)
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

func (s *WealthService) computeLoanLiabilityExposure(userID, portfolioID domain.ID, asOf time.Time) (float64, float64) {
	liability := 0.0
	accrued := 0.0
	loans := s.store.ListLoans(userID)
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

func (s *WealthService) computeLoanReceivableExposure(userID, portfolioID domain.ID, asOf time.Time) (float64, float64) {
	receivable := 0.0
	accrued := 0.0
	loans := s.store.ListLoans(userID)
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
	dailyRatePerMillion := parseAmountWithFallback(loan.DailyRatePerMillion)
	if annualRate > 1.0 {
		annualRate = annualRate / 100.0
	}
	daysPerYear := parseDayCountBasis(loan.DayCountBasis)
	if daysPerYear <= 0 {
		daysPerYear = 365
	}

	payments := s.store.ListLoanPayments(loan.UserID, loan.ID)
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
		if principal > 0 && days > 0 && dailyRatePerMillion > 0 {
			accrued = principal / 1000000 * dailyRatePerMillion * float64(days)
		} else if principal > 0 && days > 0 && annualRate > 0 {
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
		if principal > 0 && days > 0 && dailyRatePerMillion > 0 {
			accrued = principal / 1000000 * dailyRatePerMillion * float64(days)
		} else if principal > 0 && days > 0 && annualRate > 0 {
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
		UserID:               string(loan.UserID),
		AsOfAt:               asOf,
		Status:               string(loan.Status),
		Currency:             "VND",
		PrincipalInitial:     loan.PrincipalInitial,
		PrincipalBalance:     loan.PrincipalBalance,
		AnnualRate:           loan.AnnualRate,
		DailyRatePerMillion:  loan.DailyRatePerMillion,
		DailyInterest:        formatMoney(loanDailyInterest(*loan)),
		LastInterestPaidDate: lastInterestPaymentDate(s.store.ListLoanPayments(loan.UserID, loan.ID)),
		NextPaymentDate:      nextMonthlyPaymentDate(loan.StartAt, asOf),
		DayCountBasis:        firstNonEmpty(loan.DayCountBasis, "365"),
		TotalAccruedInterest: formatMoney(summary.TotalAccrued),
		TotalPaidInterest:    formatMoney(summary.TotalPaid),
		UnpaidInterest:       formatMoney(unpaid),
		Rows:                 rows,
	}, nil
}

func loanDailyInterest(loan domain.Loan) float64 {
	principal := parseAmountWithFallback(loan.PrincipalBalance)
	if rate := parseAmountWithFallback(loan.DailyRatePerMillion); rate > 0 {
		return principal / 1000000 * rate
	}
	rate := parseAmountWithFallback(loan.AnnualRate)
	if rate > 1 {
		rate /= 100
	}
	return principal * rate / float64(parseDayCountBasis(loan.DayCountBasis))
}

func lastInterestPaymentDate(payments []domain.LoanPayment) time.Time {
	for _, p := range payments {
		if parseAmountWithFallback(p.Interest) > 0 {
			return p.OccurredAt
		}
	}
	return time.Time{}
}

// nextMonthlyPaymentDate keeps the original day-of-month; when absent it uses
// the final day in the relevant month (for example 31 Jan -> 28/29 Feb).
func nextMonthlyPaymentDate(start, after time.Time) time.Time {
	if start.IsZero() {
		return time.Time{}
	}
	if after.Before(start) {
		return start
	}
	loc := start.Location()
	if loc == nil {
		loc = time.UTC
	}
	for n := 1; n < 1200; n++ {
		month := time.Date(start.Year(), start.Month()+time.Month(n), 1, start.Hour(), start.Minute(), start.Second(), 0, loc)
		last := time.Date(month.Year(), month.Month()+1, 0, start.Hour(), start.Minute(), start.Second(), 0, loc).Day()
		day := start.Day()
		if day > last {
			day = last
		}
		candidate := time.Date(month.Year(), month.Month(), day, start.Hour(), start.Minute(), start.Second(), 0, loc)
		if candidate.After(after) {
			return candidate
		}
	}
	return time.Time{}
}

func previousMonthlyPaymentDate(start, due time.Time) time.Time {
	previous := start
	for candidate := nextMonthlyPaymentDate(start, previous); !candidate.IsZero() && !candidate.After(due); candidate = nextMonthlyPaymentDate(start, candidate) {
		previous = candidate
	}
	return previous
}

func (s *WealthService) LoanPortfolioSummary(userID domain.ID) LoanPortfolioSummary {
	var principal, daily, accrued, paid float64
	for _, loan := range s.store.ListLoans(userID) {
		if loan.Direction != domain.LoanDirectionReceivable || loan.Status == domain.LoanStatusClosed || loan.Status == domain.LoanStatusCancelled {
			continue
		}
		principal += parseAmountWithFallback(loan.PrincipalBalance)
		daily += loanDailyInterest(loan)
		_, totals := s.loanAccrualsByLoan(loan, nowUTC())
		accrued += math.Max(0, totals.TotalAccrued-totals.TotalPaid)
		paid += totals.TotalPaid
	}
	return LoanPortfolioSummary{formatMoney(principal), formatMoney(daily), formatMoney(accrued), formatMoney(paid)}
}

func (s *WealthService) LoanSchedule(userID domain.ID, months int) []LoanScheduleItem {
	if months <= 0 {
		months = 3
	}
	if months > 24 {
		months = 24
	}
	now := nowUTC()
	end := now.AddDate(0, months, 0)
	out := []LoanScheduleItem{}
	for _, loan := range s.store.ListLoans(userID) {
		if loan.Direction != domain.LoanDirectionReceivable || loan.Status == domain.LoanStatusClosed || loan.Status == domain.LoanStatusCancelled {
			continue
		}
		for due := nextMonthlyPaymentDate(loan.StartAt, now.Add(-time.Nanosecond)); !due.IsZero() && !due.After(end); due = nextMonthlyPaymentDate(loan.StartAt, due) {
			prev := previousMonthlyPaymentDate(loan.StartAt, due)
			days := elapsedDays(prev, due)
			status := "upcoming"
			if due.Before(now) {
				status = "overdue"
			}
			out = append(out, LoanScheduleItem{string(loan.ID), loan.Counterparty, due, days, formatMoney(loanDailyInterest(loan) * float64(days)), status})
		}
	}
	return out
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

func (s *WealthService) GetBudget(userID domain.ID, period string) (BudgetPeriod, error) {
	if userID == "" {
		return BudgetPeriod{}, errors.New("userId is required")
	}
	period = strings.TrimSpace(period)
	if period == "" {
		return BudgetPeriod{}, errors.New("period is required")
	}
	periodStart, periodEnd, err := parseBudgetPeriod(period)
	if err != nil {
		return BudgetPeriod{}, err
	}

	budgets := s.store.ListBudgets(userID, period)
	rows := make([]BudgetRow, 0, len(budgets))
	for _, budget := range budgets {
		spent, currency := s.budgetSpentForPeriod(userID, periodStart, periodEnd, budget.CategoryID)
		rows = append(rows, BudgetRow{
			CategoryID: string(budget.CategoryID),
			Limit:      budget.Limit,
			Spent:      formatMoney(spent),
			Currency:   currency,
		})
	}
	return BudgetPeriod{
		Period: period,
		UserID: string(userID),
		AsOfAt: nowUTC(),
		Rows:   rows,
	}, nil
}

func (s *WealthService) UpsertBudget(userID domain.ID, period string, categoryID string, limit string) (domain.Budget, error) {
	period = strings.TrimSpace(period)
	if userID == "" {
		return domain.Budget{}, errors.New("userId is required")
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
		UserID:     userID,
		Period:     period,
		CategoryID: domain.ID(categoryID),
		Limit:      limit,
	})
}

func (s *WealthService) budgetSpentForPeriod(userID domain.ID, start, end time.Time, categoryID domain.ID) (float64, string) {
	transactions := s.store.ListTransactions(userID, "")
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

func (s *WealthService) computeAssetValuationUser(userID, portfolioID domain.ID, asOf time.Time) (float64, float64, int) {
	propertyValues := s.store.ListPropertyValues(userID)
	assetValues := s.store.ListAssetValues(userID)
	portfolioPropertyIDs := map[domain.ID]struct{}{}
	portfolioAssetIDs := map[domain.ID]struct{}{}
	if portfolioID != "" {
		for _, p := range s.store.ListProperties(userID) {
			if p.PortfolioID == portfolioID {
				portfolioPropertyIDs[p.ID] = struct{}{}
			}
		}
		for _, a := range s.store.ListAssets(userID) {
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
	if acc.UserID != input.UserID {
		return domain.Transaction{}, errors.New("account does not belong to user")
	}
	if input.Currency == "" {
		input.Currency = acc.Currency
	}

	// category_id is a nullable UUID in PostgreSQL. A transaction created from
	// the quick Thu/Chi form may not have a persisted category yet, so leave it
	// empty and let the storage layer insert NULL instead of a sentinel string.

	return s.store.CreateTransactionStrict(input)
}

func (s *WealthService) UpdateTransaction(input domain.Transaction) (domain.Transaction, error) {
	current, found := s.store.GetTransaction(input.ID)
	if !found || current.UserID != input.UserID {
		return domain.Transaction{}, errors.New("transaction not found")
	}
	if _, ok := s.validateTransactionType(input.Type); !ok {
		return domain.Transaction{}, errors.New("transaction type is invalid")
	}
	if _, ok := s.validateTransactionStatus(input.Status); !ok {
		return domain.Transaction{}, errors.New("transaction status is invalid")
	}
	amount, err := parseAmount(input.Amount)
	if err != nil || amount <= 0 {
		return domain.Transaction{}, errors.New("amount must be greater than 0")
	}
	account, found := s.store.GetAccount(input.AccountID)
	if !found || account.UserID != input.UserID {
		return domain.Transaction{}, errors.New("account does not belong to user")
	}
	input.PortfolioID = account.PortfolioID
	input.Currency = account.Currency
	if input.OccurredAt.IsZero() {
		return domain.Transaction{}, errors.New("occurredAt is required")
	}
	return s.store.UpdateTransaction(input)
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
	if from.UserID != input.UserID || to.UserID != input.UserID {
		return domain.Transfer{}, errors.New("accounts must belong to user")
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
	if payment.InterestDays < 0 {
		return domain.LoanPayment{}, errors.New("interest days cannot be negative")
	}

	balance, err := parseAmount(loan.PrincipalBalance)
	if err != nil {
		return domain.LoanPayment{}, errors.New("invalid current loan balance")
	}

	if principal > balance {
		return domain.LoanPayment{}, errors.New("principal payment cannot exceed remaining principal")
	}

	payment.LoanID = domain.ID(loanID)
	payment.UserID = loan.UserID
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
	if postingAccountID == "" && loan.SettlementAccountID != "" {
		postingAccountID = loan.SettlementAccountID
	}
	if postingAccountID == "" {
		if accID := s.findFirstAccountID(loan.UserID, ""); accID != "" {
			postingAccountID = accID
		}
	}

	if postingAccountID == "" {
		return domain.LoanPayment{}, errors.New("a settlement account is required for loan payment")
	}
	acc, ok := s.store.GetAccount(postingAccountID)
	if !ok || acc.UserID != loan.UserID {
		return domain.LoanPayment{}, errors.New("settlement account does not belong to loan owner")
	}
	payment.AccountID = acc.ID
	nextStatus := loan.Status
	if nextBalance <= 0 && nextStatus != domain.LoanStatusCancelled {
		nextStatus = domain.LoanStatusClosed
	}
	ledgerType := domain.TransactionTypeLoanPayment
	ledgerName := ""
	// A receivable payment made up solely of interest is income, not a loan
	// principal collection. This lets the income journal and monthly cash-flow
	// totals reflect money earned from lending without counting returned principal.
	if loan.Direction == domain.LoanDirectionReceivable && principal == 0 && interest > 0 && fee == 0 {
		ledgerType = domain.TransactionTypeIncome
		ledgerName = "Thu lãi khoản vay"
	}
	ledger := domain.Transaction{
		UserID: loan.UserID, AccountID: acc.ID, PortfolioID: acc.PortfolioID,
		Name: ledgerName, Type: ledgerType, Amount: formatMoney(totalAmount),
		Currency: currencyOrDefault(acc.Currency, loan.Direction),
		Note:     fmt.Sprintf("loan payment for %s", loanID), OccurredAt: payment.OccurredAt,
		Status: domain.TransactionStatusPosted, Source: "loan_payment",
	}
	return s.store.SettleLoanPayment(loan.ID, loan.PrincipalBalance, formatMoney(nextBalance), nextStatus, payment, ledger)
}

func (s *WealthService) CreateLoanPaymentRequest(userID domain.ID, loanID domain.ID, req PaymentRequestCreate) (domain.BankPaymentRequest, error) {
	if loanID == "" {
		return domain.BankPaymentRequest{}, errors.New("loanId is required")
	}
	loan, ok := s.store.GetLoan(loanID)
	if !ok {
		return domain.BankPaymentRequest{}, errors.New("loan not found")
	}
	if loan.UserID != userID {
		return domain.BankPaymentRequest{}, errors.New("loan does not belong to user")
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
		UserID:    userID,
		LoanID:    loanID,
		Code:      code,
		Amount:    formatMoney(amountValue),
		Currency:  req.Currency,
		ExpiresAt: expiresAt,
		Status:    "open",
		Note:      req.Note,
		Source:    "sepay_loan_request",
	})
	return created, err
}

func (s *WealthService) GetPortfolioNetWorth(portfolioID string) (NetWorthResult, error) {
	pID := domain.ID(portfolioID)
	portfolio, ok := s.store.GetPortfolio(pID)
	if !ok {
		return NetWorthResult{}, errors.New("portfolio not found")
	}
	return s.computeNetWorthForPortfolio(portfolio.UserID, pID)
}

func (s *WealthService) GetPortfolioNetWorthAt(portfolioID string, asOf time.Time) (NetWorthResult, error) {
	pID := domain.ID(portfolioID)
	portfolio, ok := s.store.GetPortfolio(pID)
	if !ok {
		return NetWorthResult{}, errors.New("portfolio not found")
	}
	return s.computeNetWorthForPortfolioAt(portfolio.UserID, pID, asOf, false)
}

func (s *WealthService) GetPortfolioSnapshots(userID domain.ID, limit int, cursor string) PortfolioSnapshotPage {
	return s.getPortfolioSnapshots(userID, "", limit, cursor)
}

func (s *WealthService) GetPortfolioSnapshotsForPortfolio(userID, portfolioID domain.ID, limit int, cursor string) PortfolioSnapshotPage {
	return s.getPortfolioSnapshots(userID, portfolioID, limit, cursor)
}

func (s *WealthService) getPortfolioSnapshots(userID, portfolioID domain.ID, limit int, cursor string) PortfolioSnapshotPage {
	if limit <= 0 {
		limit = defaultPortfolioSnapshotLimit
	}
	if limit > maxPortfolioSnapshots {
		limit = maxPortfolioSnapshots
	}

	snapshotMu.RLock()
	userHistory := historyByUser[userID]
	history := append([]netWorthSnapshot{}, userHistory[portfolioID]...)
	snapshotMu.RUnlock()
	if len(history) == 0 {
		if _, err := s.computeNetWorthForPortfolio(userID, portfolioID); err != nil {
			return PortfolioSnapshotPage{}
		}
		snapshotMu.RLock()
		userHistory = historyByUser[userID]
		history = append([]netWorthSnapshot{}, userHistory[portfolioID]...)
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

	userID, err := s.resolveUserForFeed(raw.ConnectionID)
	if err != nil {
		return domain.BankFeedTransaction{}, err
	}

	occurredAt, err := parseTime(raw.OccurredAt)
	if err != nil {
		return domain.BankFeedTransaction{}, err
	}

	accountID := domain.ID(raw.AccountID)
	var sepayBankAccountID domain.ID
	if raw.ProviderAccountID != "" {
		if bankAccount, ok := s.store.GetSePayBankAccountByXID(raw.ProviderAccountID); ok {
			userID, sepayBankAccountID = bankAccount.UserID, bankAccount.ID
			if mapping, mapped := s.store.GetBankAccountMapping(bankAccount.ID); mapped && mapping.Status == "active" {
				if mapping.UserID != userID {
					return domain.BankFeedTransaction{}, errors.New("bank account mapping does not match connection user")
				}
				accountID = mapping.AccountID
			}
		}
	}
	if accountID == "" {
		// Bank Hub IPN identifies the provider account, not the internal ledger
		// account. Phase 1 uses the user's matching/default account until
		// explicit per-bank-account mappings are introduced in Phase 2.
		accountID = s.findFirstAccountID(userID, firstNonEmpty(raw.Currency, "VND"))
	}
	if accountID == "" {
		return domain.BankFeedTransaction{}, errors.New("no Finora account is available for bank feed")
	}

	rawProviderData, _ := json.Marshal(raw)
	feed, err := s.store.IngestBankFeed(domain.BankFeedTransaction{
		UserID:       userID,
		ConnectionID: domain.ID(raw.ConnectionID),
		AccountID:    accountID,
		Amount:       raw.Amount,
		Currency:     firstNonEmpty(raw.Currency, "VND"),
		Direction:    direction,
		CounterParty: raw.Counterparty,
		Description:  strings.TrimSpace(raw.Description),
		Reference:    strings.TrimSpace(raw.Reference),
		OccurredAt:   occurredAt,
		ExternalID:   raw.ExternalID,
		PostedTxnID:  "",
		// Bank feeds are immutable review items in this MVP. Classification may
		// attach a suggestion later, but only the user can create the ledger
		// transaction through confirm/correct.
		PostingState:         domain.PostingStateReview,
		Confidence:           0,
		Evidence:             "received from sepay webhook; awaiting user review",
		SePayBankAccountID:   sepayBankAccountID,
		RawProviderData:      string(rawProviderData),
		ClassificationStatus: "needs_review",
	})
	if err != nil {
		return domain.BankFeedTransaction{}, err
	}

	return feed, nil
}

func (s *WealthService) EnqueueSePayIncoming(raw SePayWebhookEvent) (domain.BankFeedEvent, error) {
	userID, err := s.resolveUserForFeed(raw.ConnectionID)
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

	// Provider transaction IDs are only guaranteed unique within the provider
	// account. Keep that scope in the durable idempotency key.
	eventKey := strings.TrimSpace("sepay::" + raw.ProviderAccountID + "::" + raw.ExternalID)
	if raw.ProviderAccountID == "" {
		eventKey = strings.TrimSpace("sepay::connection::" + raw.ConnectionID + "::" + raw.ExternalID)
	}
	if eventKey == "" {
		eventKey = strings.TrimSpace(raw.ConnectionID + "::" + raw.Direction + "::" + raw.OccurredAt + "::" + raw.Amount + "::" + raw.Reference)
	}

	return s.store.EnqueueBankFeedEvent(domain.BankFeedEvent{
		UserID:       userID,
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
	ev, claimed := s.store.ClaimBankFeedEvent(eventID)
	if !claimed {
		return nil
	}
	nextAttempt := ev.Attempts

	var payload SePayWebhookEvent
	if err := json.Unmarshal([]byte(ev.Payload), &payload); err != nil {
		_ = s.store.UpdateBankFeedEvent(eventID, func(event *domain.BankFeedEvent) {
			event.State = nextFailureState(nextAttempt)
			event.LastError = "invalid event payload"
		})
		return errors.New("invalid event payload")
	}

	feed, err := s.ProcessSePayIncoming(payload)
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
	// Classification is advisory only. The source event and bank-feed fields
	// remain immutable; a separate suggestion record is what mobile renders.
	s.createSuggestionIfMissing(feed)

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

func (s *WealthService) createSuggestionIfMissing(feed domain.BankFeedTransaction) {
	if len(s.store.ListTransactionSuggestions(feed.ID)) > 0 {
		return
	}
	// Exact user feedback is the highest-priority classifier. It only applies
	// when a user explicitly asked Finora to remember that normalized provider
	// content, and it can always point back to the original confirmation.
	if feed.UserID != "" {
		needle := normalizeSuggestionText(feed.Description + " " + feed.Reference)
		for _, feedback := range s.store.ListClassificationFeedback(feed.UserID) {
			if !feedback.RememberChoice || feedback.Action == "ignored" || feedback.Name == "" {
				continue
			}
			prior, ok := s.store.GetBankFeed(feedback.BankFeedTransactionID)
			if !ok || strings.ToLower(prior.Direction) != strings.ToLower(feed.Direction) || normalizeSuggestionText(prior.Description+" "+prior.Reference) != needle {
				continue
			}
			_, _ = s.store.CreateTransactionSuggestion(domain.TransactionSuggestion{BankFeedTransactionID: feed.ID, SuggestedName: feedback.Name, SuggestedCategoryID: feedback.CategoryID, Source: "rule", Confidence: 100, Reason: "exact rule from user-confirmed choice", Version: "v1"})
			return
		}
	}
	if rule, confidence, reason := s.matchBestRule(feed.UserID, strings.ToLower(feed.Direction), feed); rule != nil {
		_, _ = s.store.CreateTransactionSuggestion(domain.TransactionSuggestion{
			BankFeedTransactionID: feed.ID, SuggestedName: suggestionName(feed.Description), SuggestedCategoryID: rule.CategoryID,
			Source: "rule", Confidence: confidence, Reason: reason, Version: "v1",
		})
		return
	}
	// History requires the same direction, Finora account, normalized merchant
	// text, a narrow amount range and a plausible recurring cadence. This keeps
	// an old but superficially similar payment from becoming a category hint.
	needle := normalizeSuggestionText(feed.Description + " " + feed.Reference)
	if needle != "" {
		for _, prior := range s.store.ListBankFeed(feed.UserID) {
			if !isComparableHistoryFeed(feed, prior, needle) {
				continue
			}
			if tx, ok := s.store.GetTransaction(prior.PostedTxnID); ok {
				_, _ = s.store.CreateTransactionSuggestion(domain.TransactionSuggestion{BankFeedTransactionID: feed.ID, SuggestedName: tx.Name, SuggestedCategoryID: tx.CategoryID, Source: "history", Confidence: historyConfidence(feed, prior), Reason: "khớp lịch sử: nội dung, chiều tiền, tài khoản, khoảng tiền và chu kỳ", Version: "v1"})
				return
			}
		}
	}
	// Semantic fallback runs only after exact rules/history. It compares
	// normalized merchant tokens with user-confirmed history, preserves all
	// provider fields, and remains deliberately below auto-confirm confidence.
	if suggestion, ok := s.semanticSuggestion(feed); ok {
		_, _ = s.store.CreateTransactionSuggestion(suggestion)
		return
	}
	if name := suggestionName(feed.Description); name != "" {
		_, _ = s.store.CreateTransactionSuggestion(domain.TransactionSuggestion{BankFeedTransactionID: feed.ID, SuggestedName: name, Source: "ai", Confidence: 35, Reason: "gợi ý semantic không đủ mạnh; cần xác nhận", Version: "v1"})
	}
}

func (s *WealthService) semanticSuggestion(feed domain.BankFeedTransaction) (domain.TransactionSuggestion, bool) {
	needle := normalizeSuggestionText(feed.Description + " " + feed.Reference)
	if needle == "" {
		return domain.TransactionSuggestion{}, false
	}
	bestScore := 0.0
	var best domain.BankFeedTransaction
	for _, prior := range s.store.ListBankFeed(feed.UserID) {
		if prior.ID == feed.ID || prior.PostingState != domain.PostingStatePosted || prior.PostedTxnID == "" ||
			strings.ToLower(prior.Direction) != strings.ToLower(feed.Direction) ||
			(feed.AccountID != "" && prior.AccountID != feed.AccountID) {
			continue
		}
		score := tokenJaccard(needle, normalizeSuggestionText(prior.Description+" "+prior.Reference))
		if score > bestScore {
			bestScore, best = score, prior
		}
	}
	if bestScore < 0.60 {
		return domain.TransactionSuggestion{}, false
	}
	tx, ok := s.store.GetTransaction(best.PostedTxnID)
	if !ok {
		return domain.TransactionSuggestion{}, false
	}
	confidence := math.Min(70, math.Round((40+bestScore*35)*10)/10)
	return domain.TransactionSuggestion{BankFeedTransactionID: feed.ID, SuggestedName: tx.Name, SuggestedCategoryID: tx.CategoryID, Source: "ai", Confidence: confidence, Reason: "semantic match với lịch sử đã xác nhận; cần xác nhận", Version: "v1"}, true
}

func tokenJaccard(left, right string) float64 {
	a, b := map[string]struct{}{}, map[string]struct{}{}
	for _, token := range strings.Fields(left) {
		if len([]rune(token)) > 1 {
			a[token] = struct{}{}
		}
	}
	for _, token := range strings.Fields(right) {
		if len([]rune(token)) > 1 {
			b[token] = struct{}{}
		}
	}
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	intersection := 0
	for token := range a {
		if _, ok := b[token]; ok {
			intersection++
		}
	}
	return float64(intersection) / float64(len(a)+len(b)-intersection)
}

func isComparableHistoryFeed(feed, prior domain.BankFeedTransaction, needle string) bool {
	if prior.ID == feed.ID || prior.PostingState != domain.PostingStatePosted || prior.PostedTxnID == "" ||
		strings.ToLower(prior.Direction) != strings.ToLower(feed.Direction) ||
		normalizeSuggestionText(prior.Description+" "+prior.Reference) != needle {
		return false
	}
	if feed.AccountID != "" && prior.AccountID != feed.AccountID {
		return false
	}
	amount, amountErr := parseAmount(feed.Amount)
	priorAmount, priorAmountErr := parseAmount(prior.Amount)
	if amountErr != nil || priorAmountErr != nil || amount <= 0 || priorAmount <= 0 {
		return false
	}
	// Five percent permits small bill variations while remaining conservative.
	if math.Abs(amount-priorAmount)/math.Max(amount, priorAmount) > 0.05 {
		return false
	}
	days := math.Abs(feed.OccurredAt.Sub(prior.OccurredAt).Hours() / 24)
	// A direct repeat (same day) and monthly-ish recurrence are meaningful;
	// unrelated old history is not.
	return days <= 1 || (days >= 20 && days <= 40)
}

func historyConfidence(feed, prior domain.BankFeedTransaction) float64 {
	days := math.Abs(feed.OccurredAt.Sub(prior.OccurredAt).Hours() / 24)
	if days >= 20 && days <= 40 {
		return 88
	}
	return 82
}

func normalizeSuggestionText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9à-ỹ]+`).ReplaceAllString(value, " ")
	return strings.Join(strings.Fields(value), " ")
}

func suggestionName(description string) string {
	description = strings.TrimSpace(description)
	if description == "" {
		return ""
	}
	if len([]rune(description)) > 80 {
		return string([]rune(description)[:80])
	}
	return description
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

	out, err := s.applyBankFeedPolicy(*feed, feed.UserID, strings.ToLower(feed.Direction))
	if err != nil {
		return out, nil
	}
	return out, nil
}

func (s *WealthService) applyBankFeedPolicy(feed domain.BankFeedTransaction, userID domain.ID, direction string) (domain.BankFeedTransaction, error) {
	if direction == bankDirectionOut {
		return s.applyOutboundRule(feed, userID)
	}
	return s.applyInboundRule(feed, userID)
}

func (s *WealthService) applyOutboundRule(feed domain.BankFeedTransaction, userID domain.ID) (domain.BankFeedTransaction, error) {
	accountID := s.resolveFeedAccount(feed, userID, feed.AccountID)
	if accountID == "" {
		s.updateFeedReason(feed.ID, domain.PostingStateReview, "no account available for outbound bank feed")
		if updated, ok := s.store.GetBankFeed(feed.ID); ok {
			feed = *updated
		}
		return feed, nil
	}

	rule, confidence, reason := s.matchBestRule(userID, bankDirectionOut, feed)
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

func (s *WealthService) applyInboundRule(feed domain.BankFeedTransaction, userID domain.ID) (domain.BankFeedTransaction, error) {
	// 1) Match explicit VietQR / loan payment request code first.
	if paymentReq := s.findMatchingPaymentRequest(userID, feed); paymentReq != nil {
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
	rule, confidence, reason := s.matchBestRule(userID, bankDirectionIn, feed)
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
			accountID := s.resolveFeedAccount(feed, userID, feed.AccountID)
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
		accountID := s.resolveFeedAccount(feed, userID, feed.AccountID)
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
		UserID:     acc.UserID,
		AccountID:  accountID,
		CategoryID: categoryID,
		Type:       txType,
		Amount:     feed.Amount,
		Currency:   feed.Currency,
		Note:       feed.Description,
		OccurredAt: feed.OccurredAt,
		Status:     domain.TransactionStatusPosted,
		Source:     "bank_feed",
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

func (s *WealthService) matchBestRule(userID domain.ID, direction string, feed domain.BankFeedTransaction) (*domain.AutomationRule, float64, string) {
	rules := s.store.GetUserRules(userID, feed.AccountID, direction)
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

func (s *WealthService) resolveUserForFeed(connectionID string) (domain.ID, error) {
	if connectionID == "" {
		return "", errors.New("connectionId is required")
	}
	conn, ok := s.store.GetBankConnection(domain.ID(connectionID))
	if !ok {
		return "", errors.New("connection not found")
	}
	return conn.UserID, nil
}

func (s *WealthService) resolveFeedAccount(feed domain.BankFeedTransaction, userID domain.ID, fallback domain.ID) domain.ID {
	if fallback != "" {
		acc, ok := s.store.GetAccount(fallback)
		if ok && acc.UserID == userID {
			return acc.ID
		}
	}
	return s.findFirstAccountID(userID, feed.Currency)
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

func (s *WealthService) findMatchingPaymentRequest(userID domain.ID, feed domain.BankFeedTransaction) *domain.BankPaymentRequest {
	text := strings.ToLower(feed.Description + " " + feed.Reference)
	pattern := regexp.MustCompile(`(?i)WOS-[A-Za-z0-9-]+`)
	matches := pattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}
	for _, code := range matches {
		if req, ok := s.store.GetBankPaymentRequestByCode(userID, code); ok {
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

func (s *WealthService) RulePreview(userID domain.ID, sample []domain.BankFeedTransaction) map[string]any {
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
		rule, _, _ := s.matchBestRule(userID, item.Direction, item)
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
		accountID = s.resolveFeedAccount(feed, feed.UserID, "")
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
	// A manually approved feed can be posted before a category is assigned.
	// Keep category_id NULL rather than using a non-UUID sentinel.
	categoryID := domain.ID("")

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
		accountID = s.resolveFeedAccount(*feed, feed.UserID, feed.AccountID)
	}
	if accountID == "" {
		return domain.Transaction{}, errors.New("missing account for posting")
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

func (s *WealthService) SeedDemoData(seedUserID domain.ID, userID domain.ID) {
	ws, ok := s.store.GetUser(userID)
	if !ok || ws == nil {
		return
	}

	p, ok := s.store.FirstPortfolio(userID)
	if !ok {
		return
	}

	now := nowUTC()
	accounts := s.store.ListAccounts(userID)

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
			UserID:      userID,
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
			UserID:      userID,
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

	if len(s.store.ListTransactions(userID, "")) == 0 {
		_, _ = s.CreateTransaction(domain.Transaction{
			UserID:      userID,
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
			UserID:      userID,
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
			UserID:      userID,
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

	if len(s.store.ListTransactions(userID, "")) > 0 && savingsAccountID != "" {
		_ = mainAccountID
		// Keep transfer explicit to show double-entry transfer behavior on dashboard.
		_, _ = s.CreateTransfer(domain.Transfer{
			UserID:        userID,
			FromAccountID: mainAccountID,
			ToAccountID:   savingsAccountID,
			Amount:        "2500000",
			Currency:      ws.BaseCurrency,
			Note:          "Chuyển quỹ dự phòng",
			OccurredAt:    now.AddDate(0, 0, -2),
		})
	}

	if len(s.store.ListProperties(userID)) == 0 {
		prop, err := s.store.CreateProperty(domain.Property{
			UserID:      userID,
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

	if len(s.store.ListAssets(userID)) == 0 {
		asset, err := s.store.CreateAsset(domain.Asset{
			UserID:      userID,
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

	if len(s.store.ListLoans(userID)) == 0 {
		payable, err := s.store.CreateLoan(domain.Loan{
			UserID:           userID,
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
			UserID:           userID,
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
	_, _ = s.UpsertBudget(userID, period, "uncategorized", "4000000")

	if len(s.store.ListForecastScenarios(userID)) == 0 {
		_, _ = s.store.CreateForecastScenario(domain.ForecastScenario{
			UserID:      userID,
			Name:        "Kịch bản tiết kiệm 12 tháng",
			Assumptions: "{\"inflation\": 0.05, \"targetGrowth\": 0.08, \"timeHorizonMonths\": 12}",
			Status:      "draft",
		})
	}

	if len(s.store.ListBankConnections(userID)) == 0 {
		conn, err := s.store.CreateBankConnection(domain.BankConnection{
			UserID:   userID,
			Provider: "sepay",
			Status:   "connected",
			Scope:    "read",
			BankCode: "VCB",
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

	if len(s.store.ListUserRules(userID)) == 0 {
		_, _ = s.store.CreateAutomationRule(domain.AutomationRule{
			UserID:     userID,
			AccountID:  mainAccountID,
			Name:       "Auto-tag Lương",
			Priority:   10,
			Predicate:  "contains(description,luong)",
			Direction:  "in",
			ActionType: "classify",
			Type:       "income",
			CategoryID: "income",
			Enabled:    true,
		})
	}

	if len(s.store.ListAssistantCommands(userID)) == 0 {
		_, _ = s.store.CreateAssistantCommand(domain.AssistantCommand{
			UserID:  userID,
			Command: "Khởi tạo mục tiêu tài chính: tiết kiệm cho quỹ dự phòng",
			Status:  "ready",
			Plan:    "Kiểm tra lại danh mục đầu tư, điều chỉnh kế hoạch chi tiêu theo tháng.",
		})
	}
}

func (s *WealthService) DashboardKpis(userID domain.ID, top int) ([]domain.Transaction, error) {
	txs := s.store.ListTransactions(userID, "")
	sort.Slice(txs, func(i, j int) bool {
		if txs[i].OccurredAt.Equal(txs[j].OccurredAt) {
			return txs[i].ID > txs[j].ID
		}
		return txs[i].OccurredAt.After(txs[j].OccurredAt)
	})
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

func (s *WealthService) findFirstAccountID(userID domain.ID, currency string) domain.ID {
	accounts := s.store.ListAccounts(userID)
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

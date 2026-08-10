package storage

import (
	"time"

	"wealthos-backend/internal/domain"
)

type Store interface {
	SeedDemoUser(email, name, password string) domain.ID
	CreateUser(input domain.User) (domain.User, error)
	CreateAuditLog(input domain.AuditLog) (domain.AuditLog, error)
	ListAuditLogs(userID domain.ID) []domain.AuditLog
	GetUser(id domain.ID) (*domain.User, bool)
	GetUserByID(id domain.ID) (*domain.User, bool)
	GetUserByEmail(email string) (*domain.User, bool)
	CreateEmailVerificationToken(userID domain.ID, tokenHash string, expiresAt time.Time) error
	VerifyEmail(email, tokenHash string, now time.Time) (*domain.User, error)
	GetUserSettings(userID domain.ID) (*domain.UserSettings, error)
	UpsertUserSettings(settings domain.UserSettings) (*domain.UserSettings, error)
	EnsureUserPortfolio(name, baseCurrency string, userID domain.ID) (*domain.User, error)

	GetPortfolio(id domain.ID) (*domain.Portfolio, bool)
	CreatePortfolio(input domain.Portfolio) (domain.Portfolio, error)
	ListPortfolios(userID domain.ID) []domain.Portfolio
	FirstPortfolio(userID domain.ID) (domain.Portfolio, bool)

	GetAccount(id domain.ID) (*domain.Account, bool)
	UpsertBotAccountKey(input domain.BotAccountKey) (domain.BotAccountKey, error)
	GetActiveBotAccountKey(accountID domain.ID) (*domain.BotAccountKey, bool)
	CreateAccount(input domain.Account) (domain.Account, error)
	ListAccounts(userID domain.ID) []domain.Account
	DeleteAccount(userID domain.ID, id domain.ID) error
	DeletePortfolio(userID domain.ID, id domain.ID) error
	DeleteLoan(userID domain.ID, id domain.ID) error
	DeleteProperty(userID domain.ID, id domain.ID) error
	DeleteAsset(userID domain.ID, id domain.ID) error

	CreateTransactionStrict(input domain.Transaction) (domain.Transaction, error)
	CreateTransaction(input domain.Transaction) (domain.Transaction, error)
	UpdateTransaction(input domain.Transaction) (domain.Transaction, error)
	GetTransaction(id domain.ID) (*domain.Transaction, bool)
	ListTransactions(userID domain.ID, accountID domain.ID) []domain.Transaction

	CreateTransfer(input domain.Transfer) (domain.Transfer, error)

	CreateCustomer(input domain.Customer) (domain.Customer, error)
	ListCustomers(userID domain.ID, query string, limit int) []domain.Customer
	GetCustomer(id domain.ID) (*domain.Customer, bool)

	GetLoan(id domain.ID) (*domain.Loan, bool)
	UpdateLoan(id domain.ID, mutate func(*domain.Loan)) bool
	CreateLoan(input domain.Loan) (domain.Loan, error)
	ListLoans(userID domain.ID) []domain.Loan
	ListLoanPayments(userID domain.ID, loanID domain.ID) []domain.LoanPayment
	UpsertImportReference(input domain.ImportReference) (domain.ImportReference, error)
	GetImportReference(userID domain.ID, entityType, externalCode string) (*domain.ImportReference, bool)
	GetImportReferenceByEntity(userID domain.ID, entityType string, entityID domain.ID) (*domain.ImportReference, bool)

	CreateLoanPayment(input domain.LoanPayment) (domain.LoanPayment, error)
	// SettleLoanPayment persists the ledger entry, loan balance change and payment
	// record as one unit. Implementations must leave all three unchanged on error.
	SettleLoanPayment(loanID domain.ID, expectedPrincipalBalance, nextPrincipalBalance string, nextStatus domain.LoanStatus, payment domain.LoanPayment, ledger domain.Transaction) (domain.LoanPayment, error)

	GetProperty(id domain.ID) (*domain.Property, bool)
	CreateProperty(input domain.Property) (domain.Property, error)
	ListProperties(userID domain.ID) []domain.Property
	AddPropertyValuation(v domain.PropertyValuation) (domain.PropertyValuation, error)
	ListPropertyValues(userID domain.ID) []domain.PropertyValuation

	GetAsset(id domain.ID) (*domain.Asset, bool)
	CreateAsset(input domain.Asset) (domain.Asset, error)
	ListAssets(userID domain.ID) []domain.Asset
	AddAssetValuation(v domain.AssetValuation) (domain.AssetValuation, error)
	ListAssetValues(userID domain.ID) []domain.AssetValuation

	CreateBudget(input domain.Budget) (domain.Budget, error)
	UpsertBudget(input domain.Budget) (domain.Budget, error)
	ListBudgets(userID domain.ID, period string) []domain.Budget
	UpsertBudgetAllocs(input domain.BudgetAllocation) (domain.BudgetAllocation, error)

	CreateForecastScenario(input domain.ForecastScenario) (domain.ForecastScenario, error)
	ListForecastScenarios(userID domain.ID) []domain.ForecastScenario
	ListForecastScenariosByStatus(status string) []domain.ForecastScenario
	RunForecastScenario(id domain.ID, assumptions string) (domain.ForecastScenario, error)
	FinalizeForecastScenario(id domain.ID, status string, result string) (domain.ForecastScenario, error)

	CreateBankConnection(input domain.BankConnection) (domain.BankConnection, error)
	ListBankConnections(userID domain.ID) []domain.BankConnection
	ListAllBankConnections() []domain.BankConnection
	GetBankConnection(id domain.ID) (*domain.BankConnection, bool)
	GetBankConnectionByCallbackState(callbackState string) (*domain.BankConnection, bool)
	UpdateBankConnection(id domain.ID, mutate func(*domain.BankConnection)) bool
	RevokeBankConnection(id domain.ID) (*domain.BankConnection, bool)

	UpsertSePayUserProfile(input domain.SePayUserProfile) (domain.SePayUserProfile, error)
	GetSePayUserProfile(userID domain.ID) (*domain.SePayUserProfile, bool)
	UpsertSePayBankAccount(input domain.SePayBankAccount) (domain.SePayBankAccount, error)
	ListSePayBankAccounts(userID domain.ID) []domain.SePayBankAccount
	GetSePayBankAccount(id domain.ID) (*domain.SePayBankAccount, bool)
	GetSePayBankAccountByXID(xid string) (*domain.SePayBankAccount, bool)
	SetSePayBankAccountStatus(id domain.ID, status string) bool
	UpsertBankAccountMapping(input domain.BankAccountMapping) (domain.BankAccountMapping, error)
	GetBankAccountMapping(sepayBankAccountID domain.ID) (*domain.BankAccountMapping, bool)
	DeactivateBankAccountMapping(sepayBankAccountID domain.ID) bool
	CreateSePayLinkSession(xid string, userID domain.ID, expiresAt time.Time) error
	GetSePayLinkSessionUser(xid string) (domain.ID, bool)
	CompleteSePayLinkSession(xid string) bool
	QuarantineSePayEvent(input domain.SePayUnmappedEvent) (domain.SePayUnmappedEvent, error)
	CreateTransactionSuggestion(input domain.TransactionSuggestion) (domain.TransactionSuggestion, error)
	ListTransactionSuggestions(feedID domain.ID) []domain.TransactionSuggestion
	CreateClassificationFeedback(input domain.ClassificationFeedback) (domain.ClassificationFeedback, error)
	ListClassificationFeedback(userID domain.ID) []domain.ClassificationFeedback

	CreateBankReconciliation(input domain.BankReconciliation) (domain.BankReconciliation, error)
	ListBankReconciliations(userID domain.ID, connectionID domain.ID) []domain.BankReconciliation

	EnqueueBankFeedEvent(input domain.BankFeedEvent) (domain.BankFeedEvent, error)
	ListBankFeedEvents(userID domain.ID, state string) []domain.BankFeedEvent
	GetBankFeedEvent(id domain.ID) (*domain.BankFeedEvent, bool)
	// ClaimBankFeedEvent atomically transitions a queued event to running. It
	// returns false when another worker already owns (or completed) the event.
	ClaimBankFeedEvent(id domain.ID) (*domain.BankFeedEvent, bool)
	UpdateBankFeedEvent(id domain.ID, mutate func(*domain.BankFeedEvent)) bool

	IngestBankFeed(input domain.BankFeedTransaction) (domain.BankFeedTransaction, error)
	ListBankFeed(userID domain.ID) []domain.BankFeedTransaction
	ListBankFeedByState(userID domain.ID, state domain.TransactionPostingState) []domain.BankFeedTransaction
	GetBankFeed(id domain.ID) (*domain.BankFeedTransaction, bool)
	UpdateFeedState(id domain.ID, state domain.TransactionPostingState, reason string) error
	UpdateFeed(id domain.ID, mutate func(*domain.BankFeedTransaction)) bool
	LinkBankFeedPosting(feedID domain.ID, txnID domain.ID) bool
	GetUserRules(userID domain.ID, accountID domain.ID, direction string) []domain.AutomationRule
	ListUserRules(userID domain.ID) []domain.AutomationRule
	CreateAutomationRule(input domain.AutomationRule) (domain.AutomationRule, error)
	GetAutomationRule(id domain.ID) (*domain.AutomationRule, bool)
	UpdateAutomationRule(id domain.ID, mutate func(*domain.AutomationRule)) bool
	DeleteAutomationRule(id domain.ID) bool
	ListAutomationRules(userID domain.ID) []domain.AutomationRule

	CreateBankPaymentRequest(input domain.BankPaymentRequest) (domain.BankPaymentRequest, error)
	GetBankPaymentRequestByCode(userID domain.ID, code string) (*domain.BankPaymentRequest, bool)
	ListBankPaymentRequests(userID domain.ID) []domain.BankPaymentRequest

	CreateAssistantCommand(input domain.AssistantCommand) (domain.AssistantCommand, error)
	GetAssistantCommand(id domain.ID) (*domain.AssistantCommand, bool)
	ListAssistantCommands(userID domain.ID) []domain.AssistantCommand
	UpdateAssistantCommand(id domain.ID, mutate func(*domain.AssistantCommand)) (*domain.AssistantCommand, error)

	RecordIdempotency(key string) bool
	ClearIdempotencyOlderThan(cutoff time.Time) int
}

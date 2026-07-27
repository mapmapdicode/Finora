package storage

import (
	"time"

	"wealthos-backend/internal/domain"
)

type Store interface {
	SeedDemoUser(email, name, password string) domain.ID
	CreateAuditLog(input domain.AuditLog) (domain.AuditLog, error)
	ListAuditLogs(workspaceID domain.ID) []domain.AuditLog
	GetUser(id domain.ID) (*domain.User, bool)
	GetUserByID(id domain.ID) (*domain.User, bool)
	GetUserByEmail(email string) (*domain.User, bool)
	GetUserSettings(userID domain.ID) (*domain.UserSettings, error)
	UpsertUserSettings(settings domain.UserSettings) (*domain.UserSettings, error)


	GetWorkspace(id domain.ID) (*domain.Workspace, bool)
	CreateWorkspace(name, baseCurrency string, ownerID domain.ID) (*domain.Workspace, error)
	ListWorkspaces(userID domain.ID) []domain.Workspace
	GetWorkspaceMemberRole(userID, workspaceID domain.ID) (domain.Role, bool)

	GetPortfolio(id domain.ID) (*domain.Portfolio, bool)
	CreatePortfolio(input domain.Portfolio) (domain.Portfolio, error)
	ListPortfolios(workspaceID domain.ID) []domain.Portfolio
	FirstPortfolio(workspaceID domain.ID) (domain.Portfolio, bool)

	GetAccount(id domain.ID) (*domain.Account, bool)
	CreateAccount(input domain.Account) (domain.Account, error)
	ListAccounts(workspaceID domain.ID) []domain.Account
	DeleteAccount(workspaceID domain.ID, id domain.ID) error
	DeletePortfolio(workspaceID domain.ID, id domain.ID) error
	DeleteLoan(workspaceID domain.ID, id domain.ID) error
	DeleteProperty(workspaceID domain.ID, id domain.ID) error
	DeleteAsset(workspaceID domain.ID, id domain.ID) error

	CreateTransactionStrict(input domain.Transaction) (domain.Transaction, error)
	CreateTransaction(input domain.Transaction) (domain.Transaction, error)
	GetTransaction(id domain.ID) (*domain.Transaction, bool)
	ListTransactions(workspaceID domain.ID, accountID domain.ID) []domain.Transaction

	CreateTransfer(input domain.Transfer) (domain.Transfer, error)

	GetLoan(id domain.ID) (*domain.Loan, bool)
	UpdateLoan(id domain.ID, mutate func(*domain.Loan)) bool
	CreateLoan(input domain.Loan) (domain.Loan, error)
	ListLoans(workspaceID domain.ID) []domain.Loan
	ListLoanPayments(workspaceID domain.ID, loanID domain.ID) []domain.LoanPayment

	CreateLoanPayment(input domain.LoanPayment) (domain.LoanPayment, error)

	GetProperty(id domain.ID) (*domain.Property, bool)
	CreateProperty(input domain.Property) (domain.Property, error)
	ListProperties(workspaceID domain.ID) []domain.Property
	AddPropertyValuation(v domain.PropertyValuation) (domain.PropertyValuation, error)
	ListPropertyValues(workspaceID domain.ID) []domain.PropertyValuation

	GetAsset(id domain.ID) (*domain.Asset, bool)
	CreateAsset(input domain.Asset) (domain.Asset, error)
	ListAssets(workspaceID domain.ID) []domain.Asset
	AddAssetValuation(v domain.AssetValuation) (domain.AssetValuation, error)
	ListAssetValues(workspaceID domain.ID) []domain.AssetValuation

	CreateBudget(input domain.Budget) (domain.Budget, error)
	UpsertBudget(input domain.Budget) (domain.Budget, error)
	ListBudgets(workspaceID domain.ID, period string) []domain.Budget
	UpsertBudgetAllocs(input domain.BudgetAllocation) (domain.BudgetAllocation, error)

	CreateForecastScenario(input domain.ForecastScenario) (domain.ForecastScenario, error)
	ListForecastScenarios(workspaceID domain.ID) []domain.ForecastScenario
	ListForecastScenariosByStatus(status string) []domain.ForecastScenario
	RunForecastScenario(id domain.ID, assumptions string) (domain.ForecastScenario, error)
	FinalizeForecastScenario(id domain.ID, status string, result string) (domain.ForecastScenario, error)

	CreateBankConnection(input domain.BankConnection) (domain.BankConnection, error)
	ListBankConnections(workspaceID domain.ID) []domain.BankConnection
	ListAllBankConnections() []domain.BankConnection
	GetBankConnection(id domain.ID) (*domain.BankConnection, bool)
	GetBankConnectionByCallbackState(callbackState string) (*domain.BankConnection, bool)
	UpdateBankConnection(id domain.ID, mutate func(*domain.BankConnection)) bool
	RevokeBankConnection(id domain.ID) (*domain.BankConnection, bool)

	CreateBankReconciliation(input domain.BankReconciliation) (domain.BankReconciliation, error)
	ListBankReconciliations(workspaceID domain.ID, connectionID domain.ID) []domain.BankReconciliation

	EnqueueBankFeedEvent(input domain.BankFeedEvent) (domain.BankFeedEvent, error)
	ListBankFeedEvents(workspaceID domain.ID, state string) []domain.BankFeedEvent
	GetBankFeedEvent(id domain.ID) (*domain.BankFeedEvent, bool)
	UpdateBankFeedEvent(id domain.ID, mutate func(*domain.BankFeedEvent)) bool

	IngestBankFeed(input domain.BankFeedTransaction) (domain.BankFeedTransaction, error)
	ListBankFeed(workspaceID domain.ID) []domain.BankFeedTransaction
	ListBankFeedByState(workspaceID domain.ID, state domain.TransactionPostingState) []domain.BankFeedTransaction
	GetBankFeed(id domain.ID) (*domain.BankFeedTransaction, bool)
	UpdateFeedState(id domain.ID, state domain.TransactionPostingState, reason string) error
	UpdateFeed(id domain.ID, mutate func(*domain.BankFeedTransaction)) bool
	LinkBankFeedPosting(feedID domain.ID, txnID domain.ID) bool
	GetWorkspaceRules(workspaceID domain.ID, accountID domain.ID, direction string) []domain.AutomationRule
	ListWorkspaceRules(workspaceID domain.ID) []domain.AutomationRule
	CreateAutomationRule(input domain.AutomationRule) (domain.AutomationRule, error)
	GetAutomationRule(id domain.ID) (*domain.AutomationRule, bool)
	UpdateAutomationRule(id domain.ID, mutate func(*domain.AutomationRule)) bool
	DeleteAutomationRule(id domain.ID) bool
	ListAutomationRules(workspaceID domain.ID) []domain.AutomationRule

	CreateBankPaymentRequest(input domain.BankPaymentRequest) (domain.BankPaymentRequest, error)
	GetBankPaymentRequestByCode(workspaceID domain.ID, code string) (*domain.BankPaymentRequest, bool)
	ListBankPaymentRequests(workspaceID domain.ID) []domain.BankPaymentRequest

	CreateAssistantCommand(input domain.AssistantCommand) (domain.AssistantCommand, error)
	GetAssistantCommand(id domain.ID) (*domain.AssistantCommand, bool)
	ListAssistantCommands(workspaceID domain.ID) []domain.AssistantCommand
	UpdateAssistantCommand(id domain.ID, mutate func(*domain.AssistantCommand)) (*domain.AssistantCommand, error)

	RecordIdempotency(key string) bool
	ClearIdempotencyOlderThan(cutoff time.Time) int
}

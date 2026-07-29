package domain

import "time"

type ID string

type Role string

const (
	RoleOwner  Role = "owner"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

type TransactionType string
type LoanDirection string
type LoanStatus string
type TransactionStatus string
type TransactionPostingState string

const (
	TransactionTypeIncome           TransactionType = "income"
	TransactionTypeExpense          TransactionType = "expense"
	TransactionTypeTransfer         TransactionType = "transfer"
	TransactionTypeInvestment       TransactionType = "investment_funding"
	TransactionTypeLoanDisbursement TransactionType = "loan_disbursement"
	TransactionTypeLoanPayment      TransactionType = "loan_payment"
	TransactionTypeValuationAdj     TransactionType = "valuation_adjustment"

	LoanDirectionReceivable LoanDirection = "receivable"
	LoanDirectionPayable    LoanDirection = "payable"

	LoanStatusDraft        LoanStatus = "draft"
	LoanStatusActive       LoanStatus = "active"
	LoanStatusOverdue      LoanStatus = "overdue"
	LoanStatusClosed       LoanStatus = "closed"
	LoanStatusRestructured LoanStatus = "restructured"
	LoanStatusWrittenOff   LoanStatus = "written_off"
	LoanStatusCancelled    LoanStatus = "cancelled"

	TransactionStatusDraft      TransactionStatus = "draft"
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusPosted     TransactionStatus = "posted"
	TransactionStatusReconciled TransactionStatus = "reconciled"
	TransactionStatusVoided     TransactionStatus = "voided"

	PostingStateReview    TransactionPostingState = "pending_review"
	PostingStateAutoReady TransactionPostingState = "auto_ready"
	PostingStatePosted    TransactionPostingState = "posted"
	PostingStateIgnored   TransactionPostingState = "ignored"

	BankFeedEventStateQueued  = "queued"
	BankFeedEventStateRunning = "running"
	BankFeedEventStateDone    = "done"
	BankFeedEventStateFailed  = "failed"
)

type Timestamped struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AuditLog struct {
	Timestamped
	UserID         ID     `json:"userId"`
	ActorID        ID     `json:"actorId"`
	ActorRole      string `json:"actorRole"`
	Action         string `json:"action"`
	TargetType     string `json:"targetType"`
	TargetID       ID     `json:"targetId"`
	RequestID      string `json:"requestId"`
	Path           string `json:"path"`
	Method         string `json:"method"`
	PolicyDecision string `json:"policyDecision"`
	Result         string `json:"result"`
	Reason         string `json:"reason"`
	CorrelationID  string `json:"correlationId"`
	BeforeJSON     string `json:"beforeJson"`
	AfterJSON      string `json:"afterJson"`
}

type User struct {
	Timestamped
	Email        string `json:"email"`
	Name         string `json:"name"`
	Password     string `json:"-"`
	BaseCurrency string `json:"baseCurrency,omitempty"`
}

type AmountDisplayMode string

const (
	AmountDisplayModeFull    AmountDisplayMode = "full"
	AmountDisplayModeCompact AmountDisplayMode = "compact"
)

type UserSettings struct {
	UserID            ID                `json:"userId"`
	AmountDisplayMode AmountDisplayMode `json:"amountDisplayMode"`
	UpdatedAt         time.Time         `json:"updatedAt"`
}

type Portfolio struct {
	Timestamped
	UserID       ID     `json:"userId"`
	Name         string `json:"name"`
	BaseCurrency string `json:"baseCurrency"`
}

type Account struct {
	Timestamped
	UserID      ID     `json:"userId"`
	PortfolioID ID     `json:"portfolioId"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Currency    string `json:"currency"`
}

// BotAccountKey authorizes an external bot for exactly one Finora account.
// SecretHash is the SHA-256 digest of the one-time secret, never the secret.
type BotAccountKey struct {
	Timestamped
	AccountID  ID        `json:"accountId"`
	SecretHash string    `json:"-"`
	Prefix     string    `json:"prefix"`
	RevokedAt  time.Time `json:"revokedAt,omitempty"`
}

type Category struct {
	Timestamped
	UserID ID     `json:"userId"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
}

type Transaction struct {
	Timestamped
	UserID      ID                `json:"userId"`
	AccountID   ID                `json:"accountId"`
	CategoryID  ID                `json:"categoryId"`
	PortfolioID ID                `json:"portfolioId"`
	Name        string            `json:"name"`
	Type        TransactionType   `json:"type"`
	Amount      string            `json:"amount"`
	Currency    string            `json:"currency"`
	Note        string            `json:"note"`
	OccurredAt  time.Time         `json:"occurredAt"`
	Status      TransactionStatus `json:"status"`
	Source      string            `json:"source"`
}

type Transfer struct {
	Timestamped
	UserID        ID        `json:"userId"`
	FromAccountID ID        `json:"fromAccountId"`
	ToAccountID   ID        `json:"toAccountId"`
	Amount        string    `json:"amount"`
	Currency      string    `json:"currency"`
	Note          string    `json:"note"`
	OccurredAt    time.Time `json:"occurredAt"`
}

type Loan struct {
	Timestamped
	UserID           ID            `json:"userId"`
	PortfolioID      ID            `json:"portfolioId"`
	Counterparty     string        `json:"counterparty"`
	Direction        LoanDirection `json:"direction"`
	PrincipalInitial string        `json:"principalInitial"`
	PrincipalBalance string        `json:"principalBalance"`
	AnnualRate       string        `json:"annualRate"`
	DayCountBasis    string        `json:"dayCountBasis"`
	StartAt          time.Time     `json:"startAt"`
	DueAt            time.Time     `json:"dueAt"`
	Status           LoanStatus    `json:"status"`
	InterestCompound bool          `json:"interestCompounding"`
	// DailyRatePerMillion is VND interest earned for every VND 1,000,000 of
	// outstanding principal per calendar day.  When present it takes precedence
	// over AnnualRate, preserving compatibility with legacy annual-rate loans.
	DailyRatePerMillion string `json:"dailyRatePerMillion,omitempty"`
	SettlementAccountID ID     `json:"settlementAccountId,omitempty"`
}

type LoanPayment struct {
	Timestamped
	UserID        ID        `json:"userId"`
	LoanID        ID        `json:"loanId"`
	AccountID     ID        `json:"accountId"`
	TransactionID ID        `json:"transactionId"`
	Principal     string    `json:"principalAmount"`
	Interest      string    `json:"interestAmount"`
	Fee           string    `json:"feeAmount"`
	Waived        string    `json:"waivedAmount"`
	OccurredAt    time.Time `json:"occurredAt"`
}

type Property struct {
	Timestamped
	UserID      ID        `json:"userId"`
	PortfolioID ID        `json:"portfolioId"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	AreaM2      string    `json:"areaM2"`
	PurchaseAt  time.Time `json:"purchaseAt"`
}

type PropertyValuation struct {
	Timestamped
	PropertyID  ID        `json:"propertyId"`
	Amount      string    `json:"valuationAmount"`
	Currency    string    `json:"currency"`
	Source      string    `json:"source"`
	EffectiveAt time.Time `json:"effectiveAt"`
	IsStale     bool      `json:"isStale"`
}

type Asset struct {
	Timestamped
	UserID      ID     `json:"userId"`
	PortfolioID ID     `json:"portfolioId"`
	Name        string `json:"name"`
	Type        string `json:"assetType"`
}

type AssetValuation struct {
	Timestamped
	AssetID     ID        `json:"assetId"`
	Amount      string    `json:"valuationAmount"`
	Currency    string    `json:"currency"`
	Source      string    `json:"source"`
	EffectiveAt time.Time `json:"effectiveAt"`
}

type Budget struct {
	Timestamped
	UserID     ID     `json:"userId"`
	Period     string `json:"period"`
	CategoryID ID     `json:"categoryId"`
	Limit      string `json:"limit"`
}

type BudgetAllocation struct {
	Timestamped
	BudgetID    ID     `json:"budgetId"`
	AmountSpent string `json:"amountSpent"`
	Currency    string `json:"currency"`
}

type ForecastScenario struct {
	Timestamped
	UserID      ID     `json:"userId"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Assumptions string `json:"assumptions"`
	Result      string `json:"result"`
}

type BankConnection struct {
	Timestamped
	UserID              ID        `json:"userId"`
	Provider            string    `json:"provider"`
	ExternalID          string    `json:"externalId"`
	Status              string    `json:"status"`
	Scope               string    `json:"scope"`
	BankCode            string    `json:"bankCode"`
	CallbackState       string    `json:"callbackState"`
	SyncStatus          string    `json:"syncStatus"`
	LastSyncedAt        time.Time `json:"lastSyncedAt"`
	LastSyncRequestedAt time.Time `json:"lastSyncRequestedAt"`
}

// SePayUserProfile and the related account/mapping records are intentionally
// user-owned. A mapping may expose feeds in a shared user, but cannot
// grant another member control over the user's provider connection.
type SePayUserProfile struct {
	UserID       ID        `json:"userId"`
	CompanyXID   string    `json:"companyXid"`
	Status       string    `json:"status"`
	LinkedAt     time.Time `json:"linkedAt"`
	LastSyncedAt time.Time `json:"lastSyncedAt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type SePayBankAccount struct {
	Timestamped
	UserID              ID     `json:"userId"`
	BankAccountXID      string `json:"bankAccountXid"`
	BankCode            string `json:"bankCode"`
	BankName            string `json:"bankName"`
	AccountNumberMasked string `json:"accountNumberMasked"`
	SupportsIn          bool   `json:"supportsIn"`
	SupportsOut         bool   `json:"supportsOut"`
	Status              string `json:"status"`
}

type SePayUnmappedEvent struct {
	Timestamped
	Provider       string `json:"provider"`
	BankAccountXID string `json:"bankAccountXid"`
	TransactionID  string `json:"transactionId"`
	Payload        string `json:"payload"`
	Status         string `json:"status"`
}

type BankAccountMapping struct {
	Timestamped
	SePayBankAccountID ID     `json:"sepayBankAccountId"`
	UserID             ID     `json:"userId"`
	AccountID          ID     `json:"accountId"`
	Status             string `json:"status"`
}

type BankFeedEvent struct {
	Timestamped
	UserID       ID     `json:"userId"`
	ConnectionID ID     `json:"connectionId"`
	Provider     string `json:"provider"`
	EventKey     string `json:"eventKey"`
	ExternalID   string `json:"externalTransactionId"`
	State        string `json:"state"`
	Attempts     int    `json:"attempts"`
	LastError    string `json:"lastError"`
	Payload      string `json:"payload"`
}

type BankFeedTransaction struct {
	Timestamped
	UserID               ID                      `json:"userId"`
	ConnectionID         ID                      `json:"connectionId"`
	AccountID            ID                      `json:"accountId"`
	Amount               string                  `json:"amount"`
	Currency             string                  `json:"currency"`
	Direction            string                  `json:"direction"`
	CounterParty         string                  `json:"counterparty"`
	Description          string                  `json:"description"`
	OccurredAt           time.Time               `json:"occurredAt"`
	ExternalID           string                  `json:"externalTransactionId"`
	Reference            string                  `json:"reference"`
	Confidence           float64                 `json:"classificationConfidence"`
	Evidence             string                  `json:"classificationEvidence"`
	PostingState         TransactionPostingState `json:"postingState"`
	PostedTxnID          ID                      `json:"postedTransactionId"`
	AutoClassified       bool                    `json:"autoClassified"`
	RuleID               ID                      `json:"ruleId"`
	Source               string                  `json:"source"`
	SePayBankAccountID   ID                      `json:"sepayBankAccountId"`
	RawProviderData      string                  `json:"rawProviderData"`
	ClassificationStatus string                  `json:"classificationStatus"`
}

type TransactionSuggestion struct {
	Timestamped
	BankFeedTransactionID ID      `json:"bankFeedTransactionId"`
	SuggestedName         string  `json:"suggestedName"`
	SuggestedCategoryID   ID      `json:"suggestedCategoryId"`
	Source                string  `json:"source"`
	Confidence            float64 `json:"confidence"`
	Reason                string  `json:"reason"`
	Version               string  `json:"version"`
}

type ClassificationFeedback struct {
	Timestamped
	BankFeedTransactionID ID     `json:"bankFeedTransactionId"`
	UserID                ID     `json:"userId"`
	Action                string `json:"action"`
	Name                  string `json:"name"`
	CategoryID            ID     `json:"categoryId"`
	AccountID             ID     `json:"accountId"`
	TransactionType       string `json:"transactionType"`
	Note                  string `json:"note"`
	RememberChoice        bool   `json:"rememberChoice"`
}

type BankReconciliation struct {
	Timestamped
	UserID         ID     `json:"userId"`
	ConnectionID   ID     `json:"connectionId"`
	AsOfAt         string `json:"asOfAt"`
	Status         string `json:"status"`
	Policy         string `json:"policy"`
	Source         string `json:"source"`
	Difference     string `json:"difference"`
	Notes          string `json:"notes"`
	PendingCount   int    `json:"pendingCount"`
	PostedCount    int    `json:"postedCount"`
	IgnoredCount   int    `json:"ignoredCount"`
	AutoReadyCount int    `json:"autoReadyCount"`
}

type AutomationRule struct {
	Timestamped
	UserID           ID     `json:"userId"`
	AccountID        ID     `json:"accountId"`
	Name             string `json:"name"`
	Priority         int    `json:"priority"`
	Predicate        string `json:"predicate"`
	Direction        string `json:"direction"`
	ActionType       string `json:"actionType"`
	Type             string `json:"type"`
	CategoryID       ID     `json:"categoryId"`
	Enabled          bool   `json:"enabled"`
	ContentPattern   string `json:"contentPattern"`
	ReferencePattern string `json:"referencePattern"`
	MinAmount        string `json:"minAmount"`
	MaxAmount        string `json:"maxAmount"`
}

type BankPaymentRequest struct {
	Timestamped
	UserID    ID        `json:"userId"`
	LoanID    ID        `json:"loanId"`
	Code      string    `json:"paymentCode"`
	Amount    string    `json:"amount"`
	Currency  string    `json:"currency"`
	ExpiresAt time.Time `json:"expiresAt"`
	Status    string    `json:"status"`
	Note      string    `json:"note"`
	Source    string    `json:"source"`
}

type AssistantCommand struct {
	Timestamped
	UserID     ID     `json:"userId"`
	Command    string `json:"command"`
	Status     string `json:"status"`
	Plan       string `json:"plan"`
	ApprovalID string `json:"approvalId"`
	// ApprovalExpiresAt is set for write/external_action commands.
	ApprovalExpiresAt time.Time  `json:"approvalExpiresAt"`
	ApprovalUsedAt    *time.Time `json:"approvalUsedAt,omitempty"`
}

type NetWorthSnapshot struct {
	AsOfAt       time.Time `json:"asOfAt"`
	BaseCurrency string    `json:"baseCurrency"`
	NetWorth     string    `json:"netWorth"`
	Cash         string    `json:"cash"`
	Liabilities  string    `json:"liabilities"`
}

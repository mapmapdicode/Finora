package dto

import (
	"encoding/base64"
	"fmt"
	"strings"

	"time"
	"wealthos-backend/internal/domain"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	Name            string `json:"name"`
	UserName        string `json:"userName,omitempty"`
}

type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ResendVerificationEmailRequest struct {
	Email string `json:"email"`
}

type UserCreateRequest struct {
	Name         string `json:"name"`
	BaseCurrency string `json:"baseCurrency"`
}

type PortfolioCreateRequest struct {
	Name         string `json:"name"`
	BaseCurrency string `json:"baseCurrency"`
}

type UserSettingsRequest struct {
	AmountDisplayMode string `json:"amountDisplayMode"`
}

type UserSettingsResponse struct {
	UserID            string `json:"userId"`
	AmountDisplayMode string `json:"amountDisplayMode"`
	UpdatedAt         string `json:"updatedAt"`
}

type PortfolioNetWorthResponse struct {
	AsOfAt            time.Time `json:"asOfAt"`
	BaseCurrency      string    `json:"baseCurrency"`
	NetWorth          string    `json:"netWorth"`
	Cash              string    `json:"cash"`
	Liabilities       string    `json:"liabilities"`
	AmountDisplayMode string    `json:"amountDisplayMode"`
}

type PortfolioSnapshotResponse struct {
	AsOfAt            time.Time `json:"asOfAt"`
	NetWorth          string    `json:"netWorth"`
	Attribution       string    `json:"attribution"`
	AmountDisplayMode string    `json:"amountDisplayMode"`
}

type AccountCreateRequest struct {
	PortfolioID    string `json:"portfolioId"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Currency       string `json:"currency"`
	InitialBalance string `json:"initialBalance"`
	// Balance is retained for older mobile clients that used this field.
	Balance string `json:"balance"`
}

type BotTransactionCreateRequest struct {
	Type       string `json:"type"`
	Amount     string `json:"amount"`
	Name       string `json:"name"`
	CategoryID string `json:"categoryId"`
	Note       string `json:"note"`
	OccurredAt string `json:"occurredAt"`
}

type FlexibleTime struct {
	time.Time
}

func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" || s == "null" {
		ft.Time = time.Time{}
		return nil
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			ft.Time = t.UTC()
			return nil
		}
	}
	return fmt.Errorf("cannot parse time: %s", s)
}

func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	if ft.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%q", ft.Time.Format(time.RFC3339Nano))), nil
}

type TransactionCreateRequest struct {
	AccountID   string       `json:"accountId"`
	CategoryID  string       `json:"categoryId"`
	PortfolioID string       `json:"portfolioId"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Amount      string       `json:"amount"`
	Currency    string       `json:"currency"`
	Note        string       `json:"note"`
	Status      string       `json:"status"`
	OccurredAt  FlexibleTime `json:"occurredAt"`
}

type TransactionListQuery struct {
	AccountID  string `json:"accountId"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	CategoryID string `json:"categoryId"`
	Search     string `json:"search"`
	From       string `json:"from"`
	To         string `json:"to"`
	Limit      int    `json:"limit"`
	Cursor     string `json:"cursor"`
}

type TransactionListResponse struct {
	Items             []domain.Transaction `json:"items"`
	NextCursor        string               `json:"nextCursor"`
	AmountDisplayMode string               `json:"amountDisplayMode"`
}

type PaginatedCursor struct {
	OccurredAt time.Time
	ID         string
}

func DecodeTransactionCursor(raw string) (PaginatedCursor, error) {
	c := strings.TrimSpace(raw)
	if c == "" {
		return PaginatedCursor{}, nil
	}
	rawBytes, err := base64.RawURLEncoding.DecodeString(c)
	if err != nil {
		return PaginatedCursor{}, fmt.Errorf("invalid cursor encoding")
	}
	parts := strings.SplitN(string(rawBytes), "|", 2)
	if len(parts) != 2 {
		return PaginatedCursor{}, fmt.Errorf("invalid cursor format")
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return PaginatedCursor{}, fmt.Errorf("invalid cursor timestamp")
	}
	return PaginatedCursor{OccurredAt: ts, ID: parts[1]}, nil
}

func EncodeTransactionCursor(t time.Time, id string) string {
	payload := fmt.Sprintf("%s|%s", t.Format(time.RFC3339Nano), id)
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

type TransferCreateRequest struct {
	FromAccountID string       `json:"fromAccountId"`
	ToAccountID   string       `json:"toAccountId"`
	Amount        string       `json:"amount"`
	Currency      string       `json:"currency"`
	Note          string       `json:"note"`
	OccurredAt    FlexibleTime `json:"occurredAt"`
}

type LoanCreateRequest struct {
	PortfolioID         string       `json:"portfolioId"`
	CustomerID          string       `json:"customerId"`
	Counterparty        string       `json:"counterparty"`
	Direction           string       `json:"direction"`
	PrincipalInitial    string       `json:"principalInitial"`
	AnnualRate          string       `json:"annualRate"`
	DayCountBasis       string       `json:"dayCountBasis"`
	StartAt             FlexibleTime `json:"startAt"`
	DueAt               FlexibleTime `json:"dueAt"`
	InterestCompounding bool         `json:"interestCompounding"`
	DailyRatePerMillion string       `json:"dailyRatePerMillion"`
	SettlementAccountID string       `json:"settlementAccountId"`
}

type CustomerCreateRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type LoanPaymentRequest struct {
	LoanID       string       `json:"loanId"`
	Principal    string       `json:"principalAmount"`
	Interest     string       `json:"interestAmount"`
	InterestDays int          `json:"interestDays"`
	Fee          string       `json:"feeAmount"`
	Waived       string       `json:"waivedAmount"`
	OccurredAt   FlexibleTime `json:"occurredAt"`
	AccountID    string       `json:"accountId"`
}

type PropertyCreateRequest struct {
	PortfolioID string `json:"portfolioId"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	AreaM2      string `json:"areaM2"`
}

type PropertyValuationRequest struct {
	ValuationAmount string       `json:"valuationAmount"`
	Currency        string       `json:"currency"`
	Source          string       `json:"source"`
	EffectiveAt     FlexibleTime `json:"effectiveAt"`
}

type AssetCreateRequest struct {
	PortfolioID string `json:"portfolioId"`
	Name        string `json:"name"`
	AssetType   string `json:"assetType"`
}

type AssetValuationRequest struct {
	AssetID         string       `json:"assetId"`
	ValuationAmount string       `json:"valuationAmount"`
	Currency        string       `json:"currency"`
	Source          string       `json:"source"`
	EffectiveAt     FlexibleTime `json:"effectiveAt"`
}

type BudgetRequest struct {
	Period     string `json:"period"`
	CategoryID string `json:"categoryId"`
	Limit      string `json:"limit"`
}

type ForecastScenarioRequest struct {
	Name        string `json:"name"`
	Assumptions string `json:"assumptions"`
}

type AssistantCommandRequest struct {
	Command string `json:"command"`
	Plan    string `json:"plan"`
}

type BankConnectionRequest struct {
	Provider string `json:"provider"`
	Scope    string `json:"scope"`
}

type BankFeedActionRequest struct {
	Reason     string `json:"reason"`
	AccountID  string `json:"accountId"`
	CategoryID string `json:"categoryId"`
	Type       string `json:"type"`
}

type SePayBankAccountMapRequest struct {
	AccountID string `json:"accountId"`
}

type SePayBankAccountSyncRequest struct {
	// AccountNumber is only a Hosted Link completion hint. The backend obtains
	// all persisted account metadata from Bank Hub itself.
	AccountNumber string `json:"accountNumber"`
}

type BankFeedCorrectRequest struct {
	AccountID      string `json:"accountId"`
	CategoryID     string `json:"categoryId"`
	Type           string `json:"type"`
	Name           string `json:"name"`
	Note           string `json:"note"`
	RememberChoice bool   `json:"rememberChoice"`
}

type AutomationRuleRequest struct {
	AccountID        string `json:"accountId"`
	Name             string `json:"name"`
	Predicate        string `json:"predicate"`
	ActionType       string `json:"actionType"`
	Direction        string `json:"direction"`
	Type             string `json:"type"`
	CategoryID       string `json:"categoryId"`
	Priority         int    `json:"priority"`
	Enabled          bool   `json:"enabled"`
	ContentPattern   string `json:"contentPattern"`
	ReferencePattern string `json:"referencePattern"`
	MinAmount        string `json:"minAmount"`
	MaxAmount        string `json:"maxAmount"`
}

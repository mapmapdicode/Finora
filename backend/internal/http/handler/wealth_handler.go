package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"wealthos-backend/internal/auth"
	"wealthos-backend/internal/config"
	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/email"
	"wealthos-backend/internal/http/dto"
	"wealthos-backend/internal/integration/sepay"
	"wealthos-backend/internal/metrics"
	"wealthos-backend/internal/service"
	"wealthos-backend/internal/storage"
)

type bankFeedPreviewRequest struct {
	Sample []struct {
		Direction    string `json:"direction"`
		Amount       string `json:"amount"`
		Currency     string `json:"currency"`
		Description  string `json:"description"`
		Reference    string `json:"reference"`
		CounterParty string `json:"counterparty"`
		AccountID    string `json:"accountId"`
	} `json:"sample"`
	Limit int `json:"limit"`
}

type loanPaymentRequestPayload struct {
	Amount    string    `json:"amount"`
	Currency  string    `json:"currency"`
	Note      string    `json:"note"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type botAccrualReportLoan struct {
	LoanID           string `json:"loanId"`
	Code             string `json:"code"`
	Counterparty     string `json:"counterparty"`
	PrincipalBalance string `json:"principalBalance"`
	DailyInterest    string `json:"dailyInterest"`
	AccruedInterest  string `json:"accruedInterest"`
	Days             int    `json:"days"`
}

type botAccrualReportTotals struct {
	Principal       string `json:"principal"`
	DailyInterest   string `json:"dailyInterest"`
	AccruedInterest string `json:"accruedInterest"`
}

type markdownImportPreviewRequest struct {
	Markdown  string `json:"markdown"`
	Month     string `json:"month"`
	Overwrite bool   `json:"overwrite"`
}

type assistantCommandApprovalRequest struct {
	ApprovalID string `json:"approvalId"`
}

type telegramWebhookPayload struct {
	UpdateID      int64                    `json:"update_id"`
	UserID        string                   `json:"userId"`
	Message       *telegramWebhookMessage  `json:"message"`
	EditedMessage *telegramWebhookMessage  `json:"edited_message"`
	ChannelPost   *telegramWebhookMessage  `json:"channel_post"`
	CallbackQuery *telegramWebhookCallback `json:"callback_query"`
}

type telegramWebhookMessage struct {
	MessageID int64                `json:"message_id"`
	Text      string               `json:"text"`
	Chat      *telegramWebhookChat `json:"chat"`
	From      *telegramWebhookUser `json:"from"`
}

type telegramWebhookChat struct {
	ID int64 `json:"id"`
}

type telegramWebhookUser struct {
	ID int64 `json:"id"`
}

type telegramWebhookCallback struct {
	ID      string                  `json:"id"`
	Data    string                  `json:"data"`
	From    *telegramWebhookUser    `json:"from"`
	Message *telegramWebhookMessage `json:"message"`
}

const (
	assistantStatusReceived         = "received"
	assistantStatusPlanned          = "planned"
	assistantStatusAwaitingApproval = "awaiting_approval"
	assistantStatusApproved         = "approved"
	assistantStatusDispatched       = "dispatched"
	assistantStatusRunning          = "running"
	assistantStatusCompleted        = "completed"
	assistantStatusFailed           = "failed"
	assistantStatusTimedOut         = "timed_out"
	assistantStatusRejected         = "rejected"
	assistantStatusCancelled        = "cancelled"
	assistantStatusPending          = "pending"

	assistantIntentRead           = "read"
	assistantIntentDraft          = "draft"
	assistantIntentWrite          = "write"
	assistantIntentExternalAction = "external_action"
	assistantIntentBlocked        = "blocked"

	sepayCallbackPath     = "/api/v1/integrations/sepay/callback"
	sepayDefaultReadScope = "read_transactions"
	sepayDefaultProvider  = "sepay"
	sepayMinSyncCooldown  = 30 * time.Second
)

const assistantApprovalTTL = 10 * time.Minute

type hermesExecutorEvent struct {
	CommandID string `json:"commandId"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

var (
	assistantReadKeywords = []string{
		"show",
		"xem",
		"tra cuu",
		"so du",
		"lich su",
		"lich",
		"history",
		"tong",
		"how much",
		"what is",
		"balance",
		"net worth",
		"list",
		"query",
	}
	assistantDraftKeywords = []string{
		"draft",
		"phac thao",
		"du kien",
		"kich ban",
		"scenario",
		"planning",
		"plan",
		"preview",
	}
	assistantWriteKeywords = []string{
		"ghi",
		"them",
		"tao",
		"chi",
		"update",
		"create",
		"doi",
		"sua",
		"thay doi",
		"save",
		"add",
		"insert",
	}
	assistantExternalKeywords = []string{
		"mo",
		"open",
		"truy cap",
		"vao",
		"run",
		"launch",
		"chrome",
		"url",
		"open app",
	}
	assistantBlockedKeywords = []string{
		"xoa",
		"delete",
		"remove",
		"erase",
		"clear",
		"truncate",
		"format",
		"drop",
	}
)

type WealthHandler struct {
	service            *service.WealthService
	store              storage.Store
	secret             string
	telegramSecret     string
	hermesSecret       string
	jwtSecret          string
	jwtTTL             time.Duration
	verificationSecret string
	verificationSender email.VerificationSender
	bankHub            sepay.BankHubLinkClient
	bankHubCompany     string
	bankHubAPIKey      string
	pilotBanks         map[string]struct{}
}

func NewWealthHandler(store storage.Store, svc *service.WealthService, cfg *config.Config) *WealthHandler {
	svcRef := svc
	if svcRef == nil {
		svcRef = service.NewWealthService(store, nil)
	}
	secret := ""
	telegramSecret := ""
	hermesSecret := ""
	jwtSecret := ""
	jwtTTL := 24 * time.Hour
	if cfg != nil {
		secret = cfg.SePayWebhookSecret
		telegramSecret = cfg.TelegramWebhookSecret
		hermesSecret = cfg.HermesExecutorSecret
		jwtSecret = cfg.JWTSecret
		if cfg.JWTTTL > 0 {
			jwtTTL = cfg.JWTTTL
		}
	}
	verificationSecret := jwtSecret
	if verificationSecret == "" {
		verificationSecret = "local-dev-verification-secret"
	}
	var bankHub sepay.BankHubLinkClient
	bankHubCompany := ""
	bankHubAPIKey := ""
	pilotBanks := map[string]struct{}{}
	if cfg != nil {
		candidate := sepay.NewBankHubClient(cfg.SePayBankHubBaseURL, cfg.SePayBankHubClientID, cfg.SePayBankHubSecret)
		if candidate.Configured() {
			bankHub = candidate
		}
		bankHubCompany = cfg.SePayBankHubCompanyID
		bankHubAPIKey = cfg.SePayBankHubAPIKey
		for _, code := range cfg.SePayBankHubPilotBanks {
			pilotBanks[strings.ToUpper(strings.TrimSpace(code))] = struct{}{}
		}
	}
	return &WealthHandler{
		service:            svcRef,
		store:              store,
		secret:             secret,
		telegramSecret:     telegramSecret,
		hermesSecret:       hermesSecret,
		jwtSecret:          jwtSecret,
		jwtTTL:             jwtTTL,
		verificationSecret: verificationSecret,
		verificationSender: email.NewVerificationSender(cfg),
		bankHub:            bankHub,
		bankHubCompany:     strings.TrimSpace(bankHubCompany),
		bankHubAPIKey:      strings.TrimSpace(bankHubAPIKey),
		pilotBanks:         pilotBanks,
	}
}

func (h *WealthHandler) issueAuthToken(userID string) string {
	token, err := auth.IssueToken(h.jwtSecret, userID, h.jwtTTL)
	if err != nil {
		return "token-" + userID
	}
	return token
}

func (h *WealthHandler) Register(c *gin.Context) {
	var body dto.RegisterRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	user, err := h.service.RegisterUser(body.Email, body.Password, body.ConfirmPassword, body.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "REGISTER_FAIL", "message": err.Error()})
		return
	}
	if _, err := h.store.EnsureUserPortfolio("", "VND", user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "PORTFOLIO_CREATE_FAIL", "message": "unable to create default portfolio"})
		return
	}
	if err := h.sendVerificationCode(c, user); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "EMAIL_DELIVERY_FAILED", "message": "unable to send verification email"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"user":                      user,
		"emailVerificationRequired": true,
	})
}

func (h *WealthHandler) Login(c *gin.Context) {
	var body dto.LoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	result, err := h.service.Authenticate(body.Email, body.Password)
	if err != nil {
		if err.Error() == "email verification required" {
			c.JSON(http.StatusForbidden, gin.H{"code": "EMAIL_NOT_VERIFIED", "message": "email verification required", "email": strings.TrimSpace(body.Email)})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"code": "INVALID_CREDENTIALS", "message": err.Error()})
		return
	}
	result.Token = h.issueAuthToken(string(result.User.ID))
	c.JSON(http.StatusOK, result)
}

func (h *WealthHandler) VerifyEmail(c *gin.Context) {
	var body dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	code := strings.TrimSpace(body.Code)
	if len(code) != 6 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_VERIFICATION_CODE", "message": "verification code must contain 6 digits"})
		return
	}
	user, err := h.store.VerifyEmail(strings.TrimSpace(body.Email), h.verificationCodeHash(code), time.Now().UTC())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_VERIFICATION_CODE", "message": "verification code is invalid or expired"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "token": h.issueAuthToken(string(user.ID))})
}

func (h *WealthHandler) ResendVerificationEmail(c *gin.Context) {
	var body dto.ResendVerificationEmailRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	user, ok := h.store.GetUserByEmail(strings.TrimSpace(body.Email))
	if !ok {
		// Do not disclose whether an email address is registered.
		c.JSON(http.StatusAccepted, gin.H{"message": "if the account needs verification, a code has been sent"})
		return
	}
	if user.IsEmailVerified() {
		c.JSON(http.StatusOK, gin.H{"message": "email is already verified"})
		return
	}
	if err := h.sendVerificationCode(c, *user); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "EMAIL_DELIVERY_FAILED", "message": "unable to send verification email"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "verification email sent"})
}

func (h *WealthHandler) sendVerificationCode(c *gin.Context, user domain.User) error {
	raw := make([]byte, 4)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", (uint32(raw[0])<<24|uint32(raw[1])<<16|uint32(raw[2])<<8|uint32(raw[3]))%1000000)
	if err := h.store.CreateEmailVerificationToken(user.ID, h.verificationCodeHash(code), time.Now().UTC().Add(15*time.Minute)); err != nil {
		return err
	}
	return h.verificationSender.SendVerificationCode(c.Request.Context(), user.Email, code)
}

func (h *WealthHandler) verificationCodeHash(code string) string {
	sum := hmac.New(sha256.New, []byte(h.verificationSecret))
	_, _ = sum.Write([]byte(code))
	return hex.EncodeToString(sum.Sum(nil))
}

func (h *WealthHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().UTC()})
}

func (h *WealthHandler) Readyz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready", "ts": time.Now().UTC()})
}

func (h *WealthHandler) ListPortfolios(c *gin.Context) {
	wsID := h.requireUserID(c)
	c.JSON(http.StatusOK, h.store.ListPortfolios(domain.ID(wsID)))
}

func (h *WealthHandler) CreatePortfolio(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.PortfolioCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	wsID := h.requireUserID(c)
	item := domain.Portfolio{
		UserID:       domain.ID(wsID),
		Name:         body.Name,
		BaseCurrency: body.BaseCurrency,
	}
	created, err := h.store.CreatePortfolio(item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "portfolio", created.ID, nil, created, "success", "")
	c.JSON(http.StatusCreated, created)
}

func (h *WealthHandler) GetPortfolioNetWorth(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "portfolioId is required"})
		return
	}
	portfolio, ok := h.store.GetPortfolio(domain.ID(id))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "portfolio not found"})
		return
	}
	if !h.requireUserMatch(c, portfolio.UserID) {
		return
	}

	rawAsOf := strings.TrimSpace(c.Query("asOfAt"))
	if rawAsOf == "" {
		rawAsOf = strings.TrimSpace(c.Query("asOf"))
	}
	var (
		nw  service.NetWorthResult
		err error
	)
	if rawAsOf == "" {
		nw, err = h.service.GetPortfolioNetWorth(id)
	} else {
		asOf, err := parseAsOf(rawAsOf)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_QUERY", "message": err.Error()})
			return
		}
		nw, err = h.service.GetPortfolioNetWorthAt(id, asOf)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, portfolioNetWorthResponse(nw, h.getAmountDisplayMode(c)))
}

// GetCurrentNetWorth returns the user's consolidated position across every
// portfolio. The dashboard must use this endpoint rather than picking an
// arbitrary portfolio, otherwise newly added accounts can be omitted.
func (h *WealthHandler) GetCurrentNetWorth(c *gin.Context) {
	wsID, ok := h.requireUserOrReject(c)
	if !ok {
		return
	}
	nw, err := h.service.CurrentNetWorth(wsID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, portfolioNetWorthResponse(nw, h.getAmountDisplayMode(c)))
}

func portfolioNetWorthResponse(nw service.NetWorthResult, amountDisplayMode string) dto.PortfolioNetWorthResponse {
	return dto.PortfolioNetWorthResponse{
		AsOfAt:          nw.AsOfAt,
		BaseCurrency:    nw.BaseCurrency,
		NetWorth:        nw.NetWorth,
		NetWorthChange:  nw.NetWorthChange,
		Cash:            nw.Cash,
		Liabilities:     nw.Liabilities,
		SnapshotVersion: nw.SnapshotVersion,
		Assets: dto.PortfolioNetWorthAssets{
			Cash:            nw.Assets.Cash,
			Receivables:     nw.Assets.Receivables,
			Property:        nw.Assets.Property,
			OtherAssets:     nw.Assets.OtherAssets,
			AccruedInterest: nw.Assets.AccruedInterest,
		},
		DataQuality: dto.PortfolioNetWorthDataQuality{
			ReconciledAccounts: nw.DataQuality.ReconciledAccounts,
			StaleValuations:    nw.DataQuality.StaleValuations,
			AsOfSource:         nw.DataQuality.AsOfSource,
		},
		Attribution: dto.PortfolioNetWorthAttribution{
			ExternalCashFlow: nw.Attribution.ExternalCashFlow,
			AccruedInterest:  nw.Attribution.AccruedInterest,
			ValuationChange:  nw.Attribution.ValuationChange,
			AccruedFee:       nw.Attribution.AccruedFee,
		},
		AmountDisplayMode: amountDisplayMode,
	}
}

func (h *WealthHandler) requireEditorRole(c *gin.Context) bool {
	_, ok := h.requireUserOrReject(c)
	if !ok {
		return false
	}
	role := strings.TrimSpace(func() string {
		if v, ok := c.Get("user_role"); ok {
			if s, ok2 := v.(string); ok2 {
				return s
			}
		}
		return ""
	}())
	if role == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing user role"})
		return false
	}
	if role == string(domain.RoleViewer) {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "viewer role is read-only"})
		return false
	}
	return true
}

func (h *WealthHandler) requireOwnerRole(c *gin.Context) bool {
	_, ok := h.requireUserOrReject(c)
	if !ok {
		return false
	}
	role := strings.TrimSpace(func() string {
		if v, ok := c.Get("user_role"); ok {
			if s, ok2 := v.(string); ok2 {
				return s
			}
		}
		return ""
	}())
	if role == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing user role"})
		return false
	}
	if role != string(domain.RoleOwner) {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "only owner role is allowed"})
		return false
	}
	return true
}

func (h *WealthHandler) ListPortfolioSnapshots(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "portfolioId is required"})
		return
	}
	p, ok := h.store.GetPortfolio(domain.ID(id))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "portfolio not found"})
		return
	}
	if !h.requireUserMatch(c, p.UserID) {
		return
	}

	limit := 0
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if parsed, err := strconv.Atoi(v); err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_QUERY", "message": "limit must be a positive integer"})
			return
		} else {
			limit = parsed
		}
	}
	cursor := strings.TrimSpace(c.Query("cursor"))
	out := h.service.GetPortfolioSnapshotsForPortfolio(p.UserID, p.ID, limit, cursor)
	c.JSON(http.StatusOK, out)
}

func (h *WealthHandler) ListAccounts(c *gin.Context) {
	wsID := h.requireUserID(c)
	list := h.store.ListAccounts(domain.ID(wsID))
	c.JSON(http.StatusOK, list)
}

func (h *WealthHandler) getOrCreatePrimaryPortfolio(wsID domain.ID) domain.Portfolio {
	if p, ok := h.store.FirstPortfolio(wsID); ok {
		return p
	}
	p, err := h.store.CreatePortfolio(domain.Portfolio{
		UserID:       wsID,
		Name:         "Cá nhân",
		BaseCurrency: "VND",
	})
	if err == nil {
		return p
	}
	return domain.Portfolio{}
}

func (h *WealthHandler) CreateAccount(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.AccountCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	wsID := domain.ID(h.requireUserID(c))
	pID := domain.ID(body.PortfolioID)
	if pID == "" {
		pID = h.getOrCreatePrimaryPortfolio(wsID).ID
	}
	openingBalance := strings.TrimSpace(body.InitialBalance)
	if openingBalance == "" {
		openingBalance = strings.TrimSpace(body.Balance)
	}
	if openingBalance != "" {
		amount, parseErr := strconv.ParseFloat(openingBalance, 64)
		if parseErr != nil || !(amount > 0) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_INITIAL_BALANCE", "message": "initial balance must be greater than 0"})
			return
		}
	}
	account, err := h.store.CreateAccount(domain.Account{
		UserID:      wsID,
		PortfolioID: pID,
		Name:        body.Name,
		Type:        body.Type,
		Currency:    body.Currency,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	if openingBalance != "" {
		_, err = h.service.CreateTransaction(domain.Transaction{
			UserID:      wsID,
			AccountID:   account.ID,
			PortfolioID: pID,
			Name:        "Số dư đầu kỳ",
			Type:        domain.TransactionTypeIncome,
			Amount:      openingBalance,
			Currency:    account.Currency,
			Note:        "Số dư được ghi nhận khi tạo tài khoản",
			Status:      domain.TransactionStatusPosted,
			OccurredAt:  time.Now().UTC(),
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_INITIAL_BALANCE", "message": err.Error()})
			return
		}
	}
	h.recordAudit(c, "", "account", account.ID, nil, account, "success", "")
	c.JSON(http.StatusCreated, account)
}

// CreateBotAccountKey rotates the public-bot secret for one account. The
// plaintext is returned once and is never stored or logged.
func (h *WealthHandler) CreateBotAccountKey(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	accountID := domain.ID(c.Param("id"))
	account, ok := h.store.GetAccount(accountID)
	if !ok || account.UserID != domain.ID(h.requireUserID(c)) {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "account not found"})
		return
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "KEY_GENERATION_FAILED"})
		return
	}
	secret := "finora_bot_" + base64.RawURLEncoding.EncodeToString(raw)
	digest := sha256.Sum256([]byte(secret))
	key, err := h.store.UpsertBotAccountKey(domain.BotAccountKey{AccountID: accountID, SecretHash: hex.EncodeToString(digest[:]), Prefix: secret[:18]})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "KEY_CREATE_FAILED", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "bot_account_key", key.ID, nil, map[string]any{"accountId": accountID, "prefix": key.Prefix}, "success", "")
	c.JSON(http.StatusCreated, gin.H{"accountId": accountID, "secret": secret, "prefix": key.Prefix, "warning": "Lưu secret ngay; Finora sẽ không hiển thị lại."})
}

func (h *WealthHandler) authenticateBotAccount(c *gin.Context) (*domain.Account, bool) {
	accountID := domain.ID(c.Param("id"))
	account, ok := h.store.GetAccount(accountID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		return nil, false
	}
	key, ok := h.store.GetActiveBotAccountKey(accountID)
	secret := strings.TrimSpace(c.GetHeader("X-Finora-Account-Key"))
	if !ok || secret == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return nil, false
	}
	digest := sha256.Sum256([]byte(secret))
	expected, decodeErr := hex.DecodeString(key.SecretHash)
	if decodeErr != nil || subtle.ConstantTimeCompare(expected, digest[:]) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return nil, false
	}
	// Preserve the account-key principal for idempotency and audit records. The
	// public credential remains scoped by the account ID in the URL.
	c.Set("user_id", string(account.UserID))
	return account, true
}

func (h *WealthHandler) BotCreateTransaction(c *gin.Context) {
	account, ok := h.authenticateBotAccount(c)
	if !ok {
		return
	}
	var body dto.BotTransactionCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	typeValue := domain.TransactionType(strings.ToLower(strings.TrimSpace(body.Type)))
	if typeValue != domain.TransactionTypeIncome && typeValue != domain.TransactionTypeExpense {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_TYPE", "message": "type must be income or expense"})
		return
	}
	if _, err := strconv.ParseFloat(strings.TrimSpace(body.Amount), 64); err != nil || strings.TrimSpace(body.Amount) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_AMOUNT"})
		return
	}
	occurredAt := time.Now().UTC()
	if strings.TrimSpace(body.OccurredAt) != "" {
		var err error
		occurredAt, err = parseDateTime(body.OccurredAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DATE"})
			return
		}
	}
	transaction, err := h.service.CreateTransaction(domain.Transaction{UserID: account.UserID, AccountID: account.ID, PortfolioID: account.PortfolioID, CategoryID: domain.ID(body.CategoryID), Name: strings.TrimSpace(body.Name), Type: typeValue, Amount: strings.TrimSpace(body.Amount), Currency: account.Currency, Note: strings.TrimSpace(body.Note), OccurredAt: occurredAt, Status: domain.TransactionStatusPosted, Source: "bot_public_api"})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"transaction": transaction})
}

// BotCreateLoan creates a loan settled through the account addressed by the
// public API path.  This prevents an account-scoped key from selecting another
// account or portfolio during loan creation.
func (h *WealthHandler) BotCreateLoan(c *gin.Context) {
	account, ok := h.authenticateBotAccount(c)
	if !ok {
		return
	}
	var body dto.BotLoanCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}

	counterparty := strings.TrimSpace(body.Counterparty)
	if counterparty == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "COUNTERPARTY_REQUIRED", "message": "counterparty is required"})
		return
	}
	direction := domain.LoanDirection(strings.ToLower(strings.TrimSpace(body.Direction)))
	if direction != domain.LoanDirectionReceivable && direction != domain.LoanDirectionPayable {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DIRECTION", "message": "direction must be receivable or payable"})
		return
	}
	principal := strings.TrimSpace(body.PrincipalInitial)
	principalValue, err := strconv.ParseFloat(principal, 64)
	if err != nil || principal == "" || math.IsNaN(principalValue) || math.IsInf(principalValue, 0) || principalValue <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PRINCIPAL", "message": "principalInitial must be a positive number"})
		return
	}
	annualRate := strings.TrimSpace(body.AnnualRate)
	if annualRate == "" {
		annualRate = "0"
	}
	annualRateValue, err := strconv.ParseFloat(annualRate, 64)
	if err != nil || math.IsNaN(annualRateValue) || math.IsInf(annualRateValue, 0) || annualRateValue < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ANNUAL_RATE", "message": "annualRate must be a non-negative number"})
		return
	}
	dailyRate := strings.TrimSpace(body.DailyRatePerMillion)
	if dailyRate == "" {
		dailyRate = "0"
	}
	dailyRateValue, err := strconv.ParseFloat(dailyRate, 64)
	if err != nil || math.IsNaN(dailyRateValue) || math.IsInf(dailyRateValue, 0) || dailyRateValue < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DAILY_RATE", "message": "dailyRatePerMillion must be a non-negative number"})
		return
	}

	startAt := time.Now().UTC()
	if raw := strings.TrimSpace(body.StartAt); raw != "" {
		startAt, err = parseDateTime(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_START_DATE"})
			return
		}
	}
	dueAt := startAt.AddDate(0, 1, 0)
	if raw := strings.TrimSpace(body.DueAt); raw != "" {
		dueAt, err = parseDateTime(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DUE_DATE"})
			return
		}
	}
	if !dueAt.After(startAt) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_TERM", "message": "dueAt must be after startAt"})
		return
	}

	customer, err := h.store.CreateCustomer(domain.Customer{UserID: account.UserID, Name: counterparty})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	loan, err := h.store.CreateLoan(domain.Loan{
		UserID: account.UserID, PortfolioID: account.PortfolioID, CustomerID: customer.ID,
		Counterparty: counterparty, Direction: direction, PrincipalInitial: principal,
		PrincipalBalance: principal, AnnualRate: annualRate, DailyRatePerMillion: dailyRate,
		DayCountBasis: firstNonEmpty(strings.TrimSpace(body.DayCountBasis), "365"),
		StartAt:       startAt, DueAt: dueAt, Status: domain.LoanStatusActive,
		InterestCompound: body.InterestCompounding, SettlementAccountID: account.ID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	if direction == domain.LoanDirectionReceivable {
		transaction, err := h.service.CreateTransaction(domain.Transaction{
			UserID: account.UserID, AccountID: account.ID, PortfolioID: account.PortfolioID,
			Type: domain.TransactionTypeLoanDisbursement, Amount: principal, Currency: account.Currency,
			Name: "Loan disbursement", Note: "loan principal: " + string(loan.ID),
			OccurredAt: startAt, Status: domain.TransactionStatusPosted, Source: "bot_public_api",
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
			return
		}
		h.recordAudit(c, "", "loan", loan.ID, nil, loan, "success", "public account key")
		c.JSON(http.StatusCreated, gin.H{"loan": loan, "transaction": transaction})
		return
	}
	h.recordAudit(c, "", "loan", loan.ID, nil, loan, "success", "public account key")
	c.JSON(http.StatusCreated, gin.H{"loan": loan})
}

func (h *WealthHandler) BotListTransactions(c *gin.Context) {
	account, ok := h.authenticateBotAccount(c)
	if !ok {
		return
	}
	from, to := parseDateFilter(strings.TrimSpace(c.Query("from"))), parseDateFilter(strings.TrimSpace(c.Query("to")))
	if c.Query("from") == "" || c.Query("to") == "" || from.IsZero() || to.IsZero() || from.After(to) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DATE_RANGE", "message": "from and to are required dates (YYYY-MM-DD or RFC3339), with from <= to"})
		return
	}
	// A date-only `to` means the entire requested calendar day.
	if len(strings.TrimSpace(c.Query("to"))) == len("2006-01-02") {
		to = to.Add(24*time.Hour - time.Nanosecond)
	}
	items := []domain.Transaction{}
	for _, item := range h.store.ListTransactions(account.UserID, account.ID) {
		if !item.OccurredAt.Before(from) && !item.OccurredAt.After(to) {
			items = append(items, item)
		}
	}
	c.JSON(http.StatusOK, gin.H{"accountId": account.ID, "from": from.Format(time.RFC3339), "to": to.Format(time.RFC3339), "items": items})
}

// botUserFromPath identifies the Finora workspace for the simple bot API.
// This endpoint is intended for a user's own trusted automation; it therefore
// does not require a JWT or account key. The UUID is still checked before any
// read or write is allowed, and is attached to the request for audit records.
func (h *WealthHandler) botUserFromPath(c *gin.Context) (domain.ID, bool) {
	userID := domain.ID(strings.TrimSpace(c.Param("id")))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "USER_ID_REQUIRED", "message": "userId is required"})
		return "", false
	}
	if _, found := h.store.GetUserByID(userID); !found {
		c.JSON(http.StatusNotFound, gin.H{"code": "USER_NOT_FOUND", "message": "user not found"})
		return "", false
	}
	c.Set("user_id", string(userID))
	c.Set("user_role", "bot")
	return userID, true
}

func (h *WealthHandler) botAccountForUser(userID, accountID domain.ID) (*domain.Account, error) {
	if accountID != "" {
		account, found := h.store.GetAccount(accountID)
		if !found || account.UserID != userID {
			return nil, errors.New("account not found for user")
		}
		return account, nil
	}
	accounts := h.store.ListAccounts(userID)
	if len(accounts) == 0 {
		return nil, errors.New("user has no account; create an account first")
	}
	return &accounts[0], nil
}

// BotGetUserContext gives a bot the account and open-loan IDs it needs for
// later calls. It intentionally returns only the current user's own data.
func (h *WealthHandler) BotGetUserContext(c *gin.Context) {
	userID, ok := h.botUserFromPath(c)
	if !ok {
		return
	}
	openLoans := make([]domain.Loan, 0)
	for _, loan := range h.store.ListLoans(userID) {
		if loan.Status != domain.LoanStatusClosed && loan.Status != domain.LoanStatusCancelled {
			openLoans = append(openLoans, loan)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"userId":    userID,
		"accounts":  h.store.ListAccounts(userID),
		"openLoans": openLoans,
	})
}

// BotLoanAccrualReport calculates the open receivables server-side and returns
// a ready-to-send Markdown message so a scheduled bot never needs to duplicate
// Finora's interest/day/payment logic.
func (h *WealthHandler) BotLoanAccrualReport(c *gin.Context) {
	userID, ok := h.botUserFromPath(c)
	if !ok {
		return
	}
	items := make([]botAccrualReportLoan, 0)
	principalTotal, dailyTotal, accruedTotal := 0.0, 0.0, 0.0
	for _, loan := range h.store.ListLoans(userID) {
		if loan.Direction != domain.LoanDirectionReceivable || loan.Status == domain.LoanStatusClosed || loan.Status == domain.LoanStatusCancelled {
			continue
		}
		accruals, err := h.service.GetLoanAccruals(string(loan.ID))
		if err != nil {
			continue
		}
		code := ""
		if ref, found := h.store.GetImportReferenceByEntity(userID, "loan", loan.ID); found {
			code = ref.ExternalCode
		}
		if code == "" {
			code = loan.Counterparty
		}
		days := 0
		if count := len(accruals.Rows); count > 0 {
			days = accruals.Rows[count-1].Days
		}
		item := botAccrualReportLoan{
			LoanID:           string(loan.ID),
			Code:             code,
			Counterparty:     loan.Counterparty,
			PrincipalBalance: loan.PrincipalBalance,
			DailyInterest:    accruals.DailyInterest,
			AccruedInterest:  accruals.UnpaidInterest,
			Days:             days,
		}
		items = append(items, item)
		principalTotal += parseReportAmount(item.PrincipalBalance)
		dailyTotal += parseReportAmount(item.DailyInterest)
		accruedTotal += parseReportAmount(item.AccruedInterest)
	}

	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		loc = time.UTC
	}
	asOf := time.Now().In(loc)
	lines := []string{fmt.Sprintf("📊 Lãi cộng dồn — %s", asOf.Format("02/01/2006")), ""}
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%s (%s) — %d ngày — %s", item.Code, compactVND(item.PrincipalBalance), item.Days, compactVND(item.AccruedInterest)))
	}
	totals := botAccrualReportTotals{
		Principal:       formatReportMoney(principalTotal),
		DailyInterest:   formatReportMoney(dailyTotal),
		AccruedInterest: formatReportMoney(accruedTotal),
	}
	lines = append(lines, "", fmt.Sprintf("Tổng gốc: %s", compactVND(totals.Principal)), fmt.Sprintf("Lãi/ngày: %s", compactVND(totals.DailyInterest)), fmt.Sprintf("Tổng lãi cộng dồn: %s", compactVND(totals.AccruedInterest)))
	c.JSON(http.StatusOK, gin.H{
		"userId":   userID,
		"asOfAt":   asOf,
		"currency": "VND",
		"loans":    items,
		"totals":   totals,
		"markdown": strings.Join(lines, "\n"),
	})
}

func parseReportAmount(value string) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0
	}
	return parsed
}

func formatReportMoney(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func compactVND(value string) string {
	amount := math.Round(parseReportAmount(value))
	if amount >= 1000000 && math.Mod(amount, 1000000) == 0 {
		return formatGroupedInteger(int64(amount/1000000)) + "M"
	}
	if amount >= 1000 && math.Mod(amount, 1000) == 0 {
		return formatGroupedInteger(int64(amount/1000)) + "k"
	}
	return formatGroupedInteger(int64(amount)) + "đ"
}

func formatGroupedInteger(value int64) string {
	raw := strconv.FormatInt(value, 10)
	start := 0
	if strings.HasPrefix(raw, "-") {
		start = 1
	}
	for index := len(raw) - 3; index > start; index -= 3 {
		raw = raw[:index] + "," + raw[index:]
	}
	return raw
}

// BotCreateUserTransaction records a single income or expense. accountId is
// optional so a one-account user can send only their user ID and transaction.
func (h *WealthHandler) BotCreateUserTransaction(c *gin.Context) {
	userID, ok := h.botUserFromPath(c)
	if !ok {
		return
	}
	var body dto.BotUserTransactionCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	typeValue := domain.TransactionType(strings.ToLower(strings.TrimSpace(body.Type)))
	if typeValue != domain.TransactionTypeIncome && typeValue != domain.TransactionTypeExpense {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_TYPE", "message": "type must be income or expense"})
		return
	}
	account, err := h.botAccountForUser(userID, domain.ID(strings.TrimSpace(body.AccountID)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "ACCOUNT_NOT_FOUND", "message": err.Error()})
		return
	}
	occurredAt := time.Now().UTC()
	if raw := strings.TrimSpace(body.OccurredAt); raw != "" {
		occurredAt, err = parseDateTime(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DATE", "message": "occurredAt must be YYYY-MM-DD or RFC3339"})
			return
		}
	}
	transaction, err := h.service.CreateTransaction(domain.Transaction{
		UserID: userID, AccountID: account.ID, PortfolioID: account.PortfolioID,
		CategoryID: domain.ID(strings.TrimSpace(body.CategoryID)), Name: strings.TrimSpace(body.Name),
		Type: typeValue, Amount: strings.TrimSpace(body.Amount), Currency: account.Currency,
		Note: strings.TrimSpace(body.Note), OccurredAt: occurredAt,
		Status: domain.TransactionStatusPosted, Source: "bot_user_id_api",
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "transaction", transaction.ID, nil, transaction, "success", "trusted bot user-id API")
	c.JSON(http.StatusCreated, gin.H{"transaction": transaction, "accountId": account.ID})
}

// BotUpdateUserTransaction corrects a transaction previously created by the
// simple bot API. It deliberately cannot edit loan, transfer, or bank-origin
// ledger entries because those records have related financial state.
func (h *WealthHandler) BotUpdateUserTransaction(c *gin.Context) {
	userID, ok := h.botUserFromPath(c)
	if !ok {
		return
	}
	transactionID := domain.ID(strings.TrimSpace(c.Param("transactionId")))
	transaction, found := h.store.GetTransaction(transactionID)
	if !found || transaction.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"code": "TRANSACTION_NOT_FOUND", "message": "transaction not found for user"})
		return
	}
	if transaction.Source != "bot_user_id_api" {
		c.JSON(http.StatusConflict, gin.H{"code": "TRANSACTION_NOT_EDITABLE", "message": "only transactions created by this bot API can be edited"})
		return
	}
	var body dto.BotUserTransactionUpdateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	updated := *transaction
	if body.AccountID != nil {
		account, err := h.botAccountForUser(userID, domain.ID(strings.TrimSpace(*body.AccountID)))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "ACCOUNT_NOT_FOUND", "message": err.Error()})
			return
		}
		updated.AccountID = account.ID
		updated.PortfolioID = account.PortfolioID
		updated.Currency = account.Currency
	}
	if body.Type != nil {
		typeValue := domain.TransactionType(strings.ToLower(strings.TrimSpace(*body.Type)))
		if typeValue != domain.TransactionTypeIncome && typeValue != domain.TransactionTypeExpense {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_TYPE", "message": "type must be income or expense"})
			return
		}
		updated.Type = typeValue
	}
	if body.Amount != nil {
		updated.Amount = strings.TrimSpace(*body.Amount)
	}
	if body.Name != nil {
		updated.Name = strings.TrimSpace(*body.Name)
	}
	if body.CategoryID != nil {
		updated.CategoryID = domain.ID(strings.TrimSpace(*body.CategoryID))
	}
	if body.Note != nil {
		updated.Note = strings.TrimSpace(*body.Note)
	}
	if body.OccurredAt != nil {
		occurredAt, err := parseDateTime(strings.TrimSpace(*body.OccurredAt))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DATE", "message": "occurredAt must be YYYY-MM-DD or RFC3339"})
			return
		}
		updated.OccurredAt = occurredAt
	}
	result, err := h.service.UpdateTransaction(updated)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "transaction", result.ID, transaction, result, "success", "trusted bot user-id API correction")
	c.JSON(http.StatusOK, gin.H{"transaction": result})
}

// BotCreateUserLoanPayment accepts interest-only, principal-only, and mixed
// payments. The loan must belong to the user addressed in the route.
func (h *WealthHandler) BotCreateUserLoanPayment(c *gin.Context) {
	userID, ok := h.botUserFromPath(c)
	if !ok {
		return
	}
	var body dto.BotUserLoanPaymentCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	loanID := domain.ID(strings.TrimSpace(body.LoanID))
	if loanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "LOAN_ID_REQUIRED", "message": "loanId is required"})
		return
	}
	loan, found := h.store.GetLoan(loanID)
	if !found || loan.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"code": "LOAN_NOT_FOUND", "message": "loan not found for user"})
		return
	}
	if loan.Status == domain.LoanStatusClosed || loan.Status == domain.LoanStatusCancelled {
		c.JSON(http.StatusConflict, gin.H{"code": "LOAN_CLOSED", "message": "cannot record a payment for a closed loan"})
		return
	}
	occurredAt := time.Now().UTC()
	var err error
	if raw := strings.TrimSpace(body.OccurredAt); raw != "" {
		occurredAt, err = parseDateTime(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DATE", "message": "occurredAt must be YYYY-MM-DD or RFC3339"})
			return
		}
	}
	payment, err := h.service.CreateLoanPayment(string(loanID), domain.LoanPayment{
		UserID: userID, LoanID: loanID, AccountID: domain.ID(strings.TrimSpace(body.AccountID)),
		Principal:    firstNonEmpty(strings.TrimSpace(body.Principal), "0"),
		Interest:     firstNonEmpty(strings.TrimSpace(body.Interest), "0"),
		InterestDays: body.InterestDays, Fee: firstNonEmpty(strings.TrimSpace(body.Fee), "0"),
		Waived: firstNonEmpty(strings.TrimSpace(body.Waived), "0"), OccurredAt: occurredAt,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "loan_payment", payment.ID, nil, payment, "success", "trusted bot user-id API")
	c.JSON(http.StatusCreated, gin.H{"payment": payment, "loanId": loanID})
}

func (h *WealthHandler) DeleteAccount(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	accountID := domain.ID(c.Param("id"))
	wsID := domain.ID(h.requireUserID(c))

	account, ok := h.store.GetAccount(accountID)
	targetWsID := wsID
	if ok && account.UserID != "" {
		targetWsID = account.UserID
	}

	txs := h.store.ListTransactions(targetWsID, accountID)
	if len(txs) > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"code":    "ACCOUNT_HAS_TRANSACTIONS",
			"message": "Không thể xóa tài khoản này vì đang có giao dịch hoặc dòng tiền liên kết.",
		})
		return
	}

	err := h.store.DeleteAccount(targetWsID, accountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "account", accountID, nil, nil, "success", "deleted account")
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *WealthHandler) DeletePortfolio(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	id := domain.ID(c.Param("id"))
	wsID := domain.ID(h.requireUserID(c))
	if err := h.store.DeletePortfolio(wsID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *WealthHandler) DeleteLoan(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	id := domain.ID(c.Param("id"))
	wsID := domain.ID(h.requireUserID(c))
	loan, found := h.store.GetLoan(id)
	if !found || loan.UserID != wsID {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "loan not found"})
		return
	}
	if err := h.store.DeleteLoan(wsID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "loan", id, loan, nil, "success", "deleted loan")
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *WealthHandler) DeleteProperty(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	id := domain.ID(c.Param("id"))
	wsID := domain.ID(h.requireUserID(c))
	if err := h.store.DeleteProperty(wsID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *WealthHandler) DeleteAsset(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	id := domain.ID(c.Param("id"))
	wsID := domain.ID(h.requireUserID(c))
	if err := h.store.DeleteAsset(wsID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *WealthHandler) ListTransactions(c *gin.Context) {
	query := dto.TransactionListQuery{
		AccountID:  strings.TrimSpace(c.Query("accountId")),
		Type:       strings.TrimSpace(c.Query("type")),
		Status:     strings.TrimSpace(c.Query("status")),
		CategoryID: strings.TrimSpace(c.Query("categoryId")),
		Search:     strings.TrimSpace(c.Query("search")),
		From:       strings.TrimSpace(c.Query("from")),
		To:         strings.TrimSpace(c.Query("to")),
		Cursor:     strings.TrimSpace(c.Query("cursor")),
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_QUERY", "message": "limit must be a positive integer"})
			return
		}
		limit = parsed
	}
	query.Limit = limit

	from := parseDateFilter(query.From)
	to := parseDateFilter(query.To)
	if query.From != "" && from.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_QUERY", "message": "from must be a valid date"})
		return
	}
	if query.To != "" && to.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_QUERY", "message": "to must be a valid date"})
		return
	}
	if (!from.IsZero() && !to.IsZero()) && from.After(to) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_QUERY", "message": "from must be <= to"})
		return
	}
	// A date-only end boundary is the whole calendar day. This is especially
	// important for the weekly/monthly reports, which pass dates from a picker.
	if len(query.To) == len("2006-01-02") && !to.IsZero() {
		to = to.Add(24*time.Hour - time.Nanosecond)
	}

	var cursor dto.PaginatedCursor
	if query.Cursor != "" {
		var err error
		cursor, err = dto.DecodeTransactionCursor(query.Cursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_QUERY", "message": err.Error()})
			return
		}
	}

	wsID := h.requireUserID(c)
	raw := h.store.ListTransactions(domain.ID(wsID), domain.ID(query.AccountID))
	filtered := make([]domain.Transaction, 0, len(raw))
	qType := strings.ToLower(query.Type)
	qStatus := strings.ToLower(query.Status)
	qCategory := query.CategoryID
	qSearch := strings.ToLower(query.Search)
	for _, item := range raw {
		if qType != "" && string(item.Type) != qType {
			continue
		}
		if qStatus != "" && strings.ToLower(string(item.Status)) != qStatus {
			continue
		}
		if qCategory != "" && string(item.CategoryID) != qCategory {
			continue
		}
		if !from.IsZero() && item.OccurredAt.Before(from) {
			continue
		}
		if !to.IsZero() && item.OccurredAt.After(to) {
			continue
		}
		if qSearch != "" {
			text := strings.ToLower(strings.Join([]string{
				string(item.ID),
				item.Note,
				string(item.AccountID),
				string(item.CategoryID),
				item.Note,
				string(item.Type),
				string(item.Status),
				item.Currency,
			}, " "))
			if !strings.Contains(strings.ToLower(text), qSearch) {
				continue
			}
		}
		filtered = append(filtered, item)
	}

	filtered = sortTransactionsForCursor(filtered)
	start := 0
	if query.Cursor != "" {
		for start < len(filtered) {
			item := filtered[start]
			if item.OccurredAt.After(cursor.OccurredAt) {
				start++
				continue
			}
			if item.OccurredAt.Before(cursor.OccurredAt) {
				break
			}
			if item.OccurredAt.Equal(cursor.OccurredAt) && string(item.ID) >= cursor.ID {
				start++
				continue
			}
			break
		}
	}

	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	out := filtered[start:end]
	nextCursor := ""
	if end < len(filtered) && end > 0 {
		last := out[len(out)-1]
		nextCursor = dto.EncodeTransactionCursor(last.OccurredAt, string(last.ID))
	}

	mode := h.getAmountDisplayMode(c)
	c.JSON(http.StatusOK, dto.TransactionListResponse{
		Items:             out,
		NextCursor:        nextCursor,
		AmountDisplayMode: mode,
	})
}

func (h *WealthHandler) CreateTransaction(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.TransactionCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	happened := body.OccurredAt.Time
	if happened.IsZero() {
		happened = time.Now().UTC()
	}
	wsID := domain.ID(h.requireUserID(c))
	pID := domain.ID(body.PortfolioID)
	accID := domain.ID(body.AccountID)
	if pID == "" && accID != "" {
		if acc, ok := h.store.GetAccount(accID); ok && acc.PortfolioID != "" {
			pID = acc.PortfolioID
		}
	}
	if pID == "" {
		pID = h.getOrCreatePrimaryPortfolio(wsID).ID
	}
	item, err := h.service.CreateTransaction(domain.Transaction{
		UserID:      wsID,
		AccountID:   accID,
		CategoryID:  domain.ID(body.CategoryID),
		PortfolioID: pID,
		Name:        body.Name,
		Type:        domain.TransactionType(body.Type),
		Amount:      body.Amount,
		Currency:    body.Currency,
		Note:        body.Note,
		Status:      domain.TransactionStatus(body.Status),
		OccurredAt:  happened,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "transaction", item.ID, nil, item, "success", "")
	c.JSON(http.StatusCreated, item)
}

func (h *WealthHandler) CreateTransfer(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.TransferCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	happened := body.OccurredAt.Time
	if happened.IsZero() {
		happened = time.Now().UTC()
	}
	transfer, err := h.service.CreateTransfer(domain.Transfer{
		UserID:        domain.ID(h.requireUserID(c)),
		FromAccountID: domain.ID(body.FromAccountID),
		ToAccountID:   domain.ID(body.ToAccountID),
		Amount:        body.Amount,
		Currency:      body.Currency,
		Note:          body.Note,
		OccurredAt:    happened,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "transfer", transfer.ID, nil, transfer, "success", "")
	c.JSON(http.StatusCreated, transfer)
}

func (h *WealthHandler) ListLoans(c *gin.Context) {
	wsID := h.requireUserID(c)
	c.JSON(http.StatusOK, h.store.ListLoans(domain.ID(wsID)))
}

func (h *WealthHandler) ListCustomers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	c.JSON(http.StatusOK, h.store.ListCustomers(
		domain.ID(h.requireUserID(c)), c.Query("q"), limit,
	))
}

func (h *WealthHandler) CreateCustomer(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.CustomerCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	customer, err := h.store.CreateCustomer(domain.Customer{
		UserID: domain.ID(h.requireUserID(c)),
		Name:   body.Name,
		Phone:  body.Phone,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "customer", customer.ID, nil, customer, "success", "")
	c.JSON(http.StatusCreated, customer)
}

func (h *WealthHandler) GetLoanPortfolioSummary(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.LoanPortfolioSummary(domain.ID(h.requireUserID(c))))
}

func (h *WealthHandler) GetLoanSchedule(c *gin.Context) {
	months, _ := strconv.Atoi(c.DefaultQuery("months", "3"))
	c.JSON(http.StatusOK, h.service.LoanSchedule(domain.ID(h.requireUserID(c)), months))
}

func (h *WealthHandler) CreateLoan(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.LoanCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	startAt := body.StartAt.Time
	if startAt.IsZero() {
		startAt = time.Now().UTC()
	}
	dueAt := body.DueAt.Time
	if dueAt.IsZero() {
		dueAt = startAt.AddDate(0, 1, 0)
	}
	if body.SettlementAccountID != "" {
		account, found := h.store.GetAccount(domain.ID(body.SettlementAccountID))
		if !found || account.UserID != domain.ID(h.requireUserID(c)) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "settlement account not found"})
			return
		}
	}
	userID := domain.ID(h.requireUserID(c))
	customerID := domain.ID(body.CustomerID)
	counterparty := strings.TrimSpace(body.Counterparty)
	if customerID != "" {
		customer, found := h.store.GetCustomer(customerID)
		if !found || customer.UserID != userID {
			c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "customer not found"})
			return
		}
		counterparty = customer.Name
	} else if counterparty != "" {
		customer, err := h.store.CreateCustomer(domain.Customer{UserID: userID, Name: counterparty})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
			return
		}
		customerID = customer.ID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "customer is required"})
		return
	}
	lo, err := h.store.CreateLoan(domain.Loan{
		UserID:              userID,
		PortfolioID:         domain.ID(body.PortfolioID),
		CustomerID:          customerID,
		Counterparty:        counterparty,
		Direction:           domain.LoanDirection(body.Direction),
		PrincipalInitial:    body.PrincipalInitial,
		PrincipalBalance:    body.PrincipalInitial,
		AnnualRate:          body.AnnualRate,
		DailyRatePerMillion: firstNonEmpty(body.DailyRatePerMillion, "0"),
		SettlementAccountID: domain.ID(body.SettlementAccountID),
		DayCountBasis:       body.DayCountBasis,
		StartAt:             startAt,
		DueAt:               dueAt,
		Status:              domain.LoanStatusActive,
		InterestCompound:    body.InterestCompounding,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	// Disbursement is a movement from cash to a receivable, not an expense.
	// Keep this ledger entry separate from normal spending so net-worth remains
	// unchanged when money is lent out.
	if lo.Direction == domain.LoanDirectionReceivable && lo.SettlementAccountID != "" {
		account, _ := h.store.GetAccount(lo.SettlementAccountID)
		if _, err := h.service.CreateTransaction(domain.Transaction{
			UserID: lo.UserID, AccountID: account.ID, PortfolioID: account.PortfolioID,
			Type: domain.TransactionTypeLoanDisbursement, Amount: lo.PrincipalInitial,
			Currency: account.Currency, Name: "Loan disbursement", Note: "loan principal: " + string(lo.ID),
			OccurredAt: lo.StartAt, Status: domain.TransactionStatusPosted, Source: "loan_disbursement",
		}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
			return
		}
	}
	h.recordAudit(c, "", "loan", lo.ID, nil, lo, "success", "")
	c.JSON(http.StatusCreated, lo)
}

func (h *WealthHandler) GetLoanAccruals(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "loanId is required"})
		return
	}
	loan, ok := h.store.GetLoan(domain.ID(id))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "loan not found"})
		return
	}
	if !h.requireUserMatch(c, loan.UserID) {
		return
	}
	out, err := h.service.GetLoanAccruals(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *WealthHandler) CreateLoanPayment(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.LoanPaymentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	loanID := strings.TrimSpace(c.Param("id"))
	if loanID == "" {
		loanID = strings.TrimSpace(body.LoanID)
	}
	if loanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "loanId is required"})
		return
	}
	loan, ok := h.store.GetLoan(domain.ID(loanID))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "loan not found"})
		return
	}
	if !h.requireUserMatch(c, loan.UserID) {
		return
	}
	payment, err := h.service.CreateLoanPayment(loanID, domain.LoanPayment{
		UserID:       loan.UserID,
		Principal:    body.Principal,
		Interest:     body.Interest,
		InterestDays: body.InterestDays,
		Fee:          body.Fee,
		Waived:       body.Waived,
		AccountID:    domain.ID(firstNonEmpty(body.AccountID, string(loan.SettlementAccountID))),
		OccurredAt:   nowOrUTC(body.OccurredAt.Time),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "loan_payment", payment.ID, nil, payment, "success", "")
	c.JSON(http.StatusCreated, payment)
}

func (h *WealthHandler) ListLoanPayments(c *gin.Context) {
	loanID := domain.ID(strings.TrimSpace(c.Param("id")))
	if loanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "loanId is required"})
		return
	}
	loan, found := h.store.GetLoan(loanID)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "loan not found"})
		return
	}
	if !h.requireUserMatch(c, loan.UserID) {
		return
	}
	c.JSON(http.StatusOK, h.store.ListLoanPayments(loan.UserID, loanID))
}

func (h *WealthHandler) ListProperties(c *gin.Context) {
	wsID := h.requireUserID(c)
	c.JSON(http.StatusOK, h.store.ListProperties(domain.ID(wsID)))
}

func (h *WealthHandler) CreateProperty(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.PropertyCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	item, err := h.store.CreateProperty(domain.Property{
		UserID:      domain.ID(h.requireUserID(c)),
		PortfolioID: domain.ID(body.PortfolioID),
		Name:        body.Name,
		Address:     body.Address,
		AreaM2:      body.AreaM2,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "property", item.ID, nil, item, "success", "")
	c.JSON(http.StatusCreated, item)
}

func (h *WealthHandler) AddPropertyValuation(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "propertyId is required"})
		return
	}
	prop, ok := h.store.GetProperty(domain.ID(id))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "property not found"})
		return
	}
	if !h.requireUserMatch(c, prop.UserID) {
		return
	}
	var body dto.PropertyValuationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	v, err := h.store.AddPropertyValuation(domain.PropertyValuation{
		PropertyID:  domain.ID(id),
		Amount:      body.ValuationAmount,
		Currency:    body.Currency,
		Source:      body.Source,
		EffectiveAt: nowOrUTC(body.EffectiveAt.Time),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "property_valuation", v.ID, nil, v, "success", "")
	c.JSON(http.StatusCreated, v)
}

func (h *WealthHandler) ListAssets(c *gin.Context) {
	wsID := h.requireUserID(c)
	c.JSON(http.StatusOK, h.store.ListAssets(domain.ID(wsID)))
}

func (h *WealthHandler) CreateAsset(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.AssetCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	item, err := h.store.CreateAsset(domain.Asset{
		UserID:      domain.ID(h.requireUserID(c)),
		PortfolioID: domain.ID(body.PortfolioID),
		Name:        body.Name,
		Type:        body.AssetType,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "asset", item.ID, nil, item, "success", "")
	c.JSON(http.StatusCreated, item)
}

func (h *WealthHandler) AddAssetValuation(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	assetID := strings.TrimSpace(c.Param("id"))
	var body dto.AssetValuationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	if assetID == "" {
		assetID = strings.TrimSpace(body.AssetID)
	}
	if assetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "assetId is required"})
		return
	}
	asset, ok := h.store.GetAsset(domain.ID(assetID))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "asset not found"})
		return
	}
	if !h.requireUserMatch(c, asset.UserID) {
		return
	}
	v, err := h.store.AddAssetValuation(domain.AssetValuation{
		AssetID:     domain.ID(assetID),
		Amount:      body.ValuationAmount,
		Currency:    body.Currency,
		Source:      body.Source,
		EffectiveAt: nowOrUTC(body.EffectiveAt.Time),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "asset_valuation", v.ID, nil, v, "success", "")
	c.JSON(http.StatusCreated, v)
}

func (h *WealthHandler) GetBudget(c *gin.Context) {
	wsID := h.requireUserID(c)
	if wsID == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing user"})
		return
	}
	period := strings.TrimSpace(c.Param("period"))
	out, err := h.service.GetBudget(domain.ID(wsID), period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *WealthHandler) UpsertBudget(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	wsID := h.requireUserID(c)
	if wsID == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing user"})
		return
	}

	var body dto.BudgetRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}

	period := strings.TrimSpace(c.Param("period"))
	if period == "" {
		period = strings.TrimSpace(body.Period)
	}
	categoryID := strings.TrimSpace(body.CategoryID)
	limit := strings.TrimSpace(body.Limit)

	item, err := h.service.UpsertBudget(domain.ID(wsID), period, categoryID, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "budget", item.ID, nil, item, "success", "")
	c.JSON(http.StatusOK, item)
}

func (h *WealthHandler) ListForecastScenarios(c *gin.Context) {
	wsID := h.requireUserID(c)
	c.JSON(http.StatusOK, h.store.ListForecastScenarios(domain.ID(wsID)))
}

func (h *WealthHandler) CreateForecastScenario(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.ForecastScenarioRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	scenario, err := h.store.CreateForecastScenario(domain.ForecastScenario{
		UserID:      domain.ID(h.requireUserID(c)),
		Name:        body.Name,
		Assumptions: body.Assumptions,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "forecast_scenario", scenario.ID, nil, scenario, "success", "")
	c.JSON(http.StatusCreated, scenario)
}

func (h *WealthHandler) RunForecastScenario(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	wsID, ok := h.requireUserOrReject(c)
	if !ok {
		return
	}
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "scenarioId is required"})
		return
	}
	found := false
	for _, item := range h.store.ListForecastScenarios(domain.ID(wsID)) {
		if string(item.ID) == id {
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "scenario not found"})
		return
	}
	out, err := h.store.RunForecastScenario(domain.ID(id), c.Query("assumptions"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "forecast_scenario", out.ID, nil, out, "success", "")
	c.JSON(http.StatusAccepted, out)
}

func (h *WealthHandler) CreateSePayConnection(c *gin.Context) {
	if !h.requireOwnerRole(c) {
		return
	}
	var body dto.BankConnectionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	provider := strings.ToLower(strings.TrimSpace(body.Provider))
	if provider == "" {
		provider = sepayDefaultProvider
	}
	if provider != sepayDefaultProvider {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PROVIDER", "message": "provider must be sepay"})
		return
	}
	scope := strings.ToLower(strings.TrimSpace(body.Scope))
	if scope == "" {
		scope = sepayDefaultReadScope
	}
	if scope != sepayDefaultReadScope {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_SCOPE", "message": "only read_transactions scope is supported"})
		return
	}
	callbackState := uuid.NewString()
	conn, err := h.store.CreateBankConnection(domain.BankConnection{
		UserID:        domain.ID(h.requireUserID(c)),
		Provider:      provider,
		Scope:         scope,
		ExternalID:    fmt.Sprintf("conn_%s", uuid.NewString()),
		CallbackState: callbackState,
		SyncStatus:    "idle",
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "bank_connection", conn.ID, nil, conn, "success", "")

	connectURL := h.buildSePayConnectURL(c.Request, callbackState)
	linkSessionXID := ""
	linkExpiresAt := ""
	if h.bankHub != nil && h.bankHub.Configured() {
		link, err := h.bankHub.CreateLink(c.Request.Context(), h.bankHubCompany, connectURL)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"code": "SEPAY_BANKHUB_UNAVAILABLE", "message": err.Error()})
			return
		}
		linkSessionXID = link.XID
		linkExpiresAt = link.ExpiresAt
		connectURL = link.HostedLinkURL
		_ = h.store.UpdateBankConnection(conn.ID, func(item *domain.BankConnection) {
			// Until Bank Hub emits BANK_ACCOUNT_LINKED, ExternalID identifies this
			// short-lived Hosted Link. The event later replaces it with bank_account_xid.
			item.ExternalID = link.XID
			item.Status = "pending"
			item.SyncStatus = "link_pending"
		})
		if refreshed, ok := h.store.GetBankConnection(conn.ID); ok {
			conn = *refreshed
		}
	}
	c.JSON(http.StatusCreated, gin.H{
		"connectionId":     conn.ID,
		"provider":         conn.Provider,
		"scope":            conn.Scope,
		"externalId":       conn.ExternalID,
		"callbackState":    conn.CallbackState,
		"connectUrl":       connectURL,
		"linkSessionXid":   linkSessionXID,
		"hosted_link_url":  connectURL,
		"link_session_xid": linkSessionXID,
		"expires_at":       linkExpiresAt,
	})
}

func (h *WealthHandler) SePayCallback(c *gin.Context) {
	callbackState := strings.TrimSpace(c.Query("state"))
	if callbackState == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_CALLBACK_STATE", "message": "state is required"})
		return
	}

	connection, ok := h.store.GetBankConnectionByCallbackState(callbackState)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "connection not found for callback state"})
		return
	}

	connectionID := strings.TrimSpace(firstNonEmpty(c.Query("connectionId"), c.Query("connection_id")))
	if connectionID != "" && connectionID != string(connection.ID) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_CALLBACK", "message": "connectionId does not match callback state"})
		return
	}
	now := time.Now().UTC()
	_ = h.store.UpdateBankConnection(connection.ID, func(item *domain.BankConnection) {
		item.SyncStatus = "callback"
		item.CallbackState = callbackState
		item.LastSyncedAt = now
	})

	c.JSON(http.StatusOK, gin.H{"status": "callback_received", "connectionId": connection.ID, "callbackState": callbackState})
}

func (h *WealthHandler) GetBankConnections(c *gin.Context) {
	wsID := h.requireUserID(c)
	c.JSON(http.StatusOK, h.store.ListBankConnections(domain.ID(wsID)))
}

func (h *WealthHandler) SyncBankConnection(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	id := domain.ID(c.Param("id"))
	conn, ok := h.store.GetBankConnection(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "connection not found"})
		return
	}
	if !h.requireUserMatch(c, conn.UserID) {
		return
	}
	now := time.Now().UTC()
	if !conn.LastSyncRequestedAt.IsZero() && now.Sub(conn.LastSyncRequestedAt) < sepayMinSyncCooldown {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":    "SYNC_RATE_LIMIT",
			"message": "sync requests must be at least 30s apart",
		})
		return
	}
	if !h.store.UpdateBankConnection(id, func(item *domain.BankConnection) {
		item.SyncStatus = "queued"
		item.LastSyncRequestedAt = now
	}) {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "connection not found"})
		return
	}
	updatedConn, _ := h.store.GetBankConnection(id)
	h.recordAudit(c, "", "bank_connection", id, conn, updatedConn, "success", "")
	c.JSON(http.StatusAccepted, gin.H{"status": "sync_queued", "connectionId": id})
}

func (h *WealthHandler) buildSePayConnectURL(req *http.Request, callbackState string) string {
	if req == nil {
		return sepayCallbackPath
	}
	scheme := strings.TrimSpace(req.Header.Get("X-Forwarded-Proto"))
	if idx := strings.Index(scheme, ","); idx >= 0 {
		scheme = strings.TrimSpace(scheme[:idx])
	}
	if scheme == "" {
		if req.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := strings.TrimSpace(firstNonEmpty(req.Header.Get("X-Forwarded-Host"), req.Host))
	if host == "" {
		return sepayCallbackPath
	}
	u := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   sepayCallbackPath,
	}
	q := url.Values{}
	q.Set("state", callbackState)
	u.RawQuery = q.Encode()
	return u.String()
}

func (h *WealthHandler) RevokeBankConnection(c *gin.Context) {
	if !h.requireOwnerRole(c) {
		return
	}
	id := domain.ID(c.Param("id"))
	conn, ok := h.store.GetBankConnection(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "connection not found"})
		return
	}
	if !h.requireUserMatch(c, conn.UserID) {
		return
	}
	revoked, ok := h.store.RevokeBankConnection(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "connection not found"})
		return
	}
	h.recordAudit(c, "", "bank_connection", id, conn, revoked, "success", "")
	c.JSON(http.StatusOK, gin.H{"status": "revoked", "connectionId": id, "connection": revoked})
}

func (h *WealthHandler) WebhookSePay(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PAYLOAD", "message": "empty webhook payload"})
		return
	}

	if err := h.verifySePayWebhook(body, c.Request); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "SEPAY_WEBHOOK_UNAUTHORIZED", "message": err.Error()})
		return
	}

	payload, err := parseSePayWebhookPayload(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	// Dashboard Webhooks identify the account by accountNumber rather than a
	// Finora connection ID. Resolve only an already configured connection; do
	// not let an incoming payload choose a user or account.
	if payload.ConnectionID == "" && payload.AccountID != "" {
		if connection, ok := h.bankConnectionByExternalID(payload.AccountID); ok {
			payload.ConnectionID = string(connection.ID)
		}
	}
	if payload.ConnectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "SEPAY_WEBHOOK_FAIL", "message": "no mapped Finora connection for this SePay account"})
		return
	}

	event, err := h.service.EnqueueSePayIncoming(payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "SEPAY_WEBHOOK_FAIL", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"status":       "accepted",
		"eventId":      event.ID,
		"eventState":   event.State,
		"connectionId": event.ConnectionID,
		"externalId":   event.ExternalID,
	})
}

// BankHubIPN receives real-time debit/credit events from SePay Bank Hub. The
// provider identifies the bank account, never a Finora user, so we first
// resolve the account against the connection established by a Bank Hub event.
func (h *WealthHandler) BankHubIPN(c *gin.Context) {
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20))
	if err != nil || len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PAYLOAD", "message": "empty webhook payload"})
		return
	}
	if err := h.verifyBankHubAPIKey(c.Request); err != nil {
		metrics.Inc("sepay_webhook_failures_total")
		c.JSON(http.StatusUnauthorized, gin.H{"code": "SEPAY_BANKHUB_UNAUTHORIZED", "message": err.Error()})
		return
	}

	payload, bankAccountXID, err := h.parseBankHubIPN(body)
	if err != nil {
		metrics.Inc("sepay_webhook_failures_total")
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PAYLOAD", "message": err.Error()})
		return
	}
	conn, ok := h.bankConnectionByExternalID(bankAccountXID)
	if !ok {
		metrics.Inc("sepay_unknown_bank_account_total")
		quarantined, quarantineErr := h.store.QuarantineSePayEvent(domain.SePayUnmappedEvent{Provider: "sepay", BankAccountXID: bankAccountXID, TransactionID: payload.ExternalID, Payload: string(body), Status: "quarantined"})
		if quarantineErr != nil {
			metrics.Inc("sepay_webhook_failures_total")
			c.JSON(http.StatusInternalServerError, gin.H{"code": "SEPAY_BANKHUB_IPN_FAIL", "message": "could not persist unmapped event"})
			return
		}
		// Ack a durable, valid provider event even before the user finishes
		// mapping it. This avoids provider retry storms; the record remains
		// visible to operations for reconciliation.
		c.JSON(http.StatusOK, gin.H{"success": true, "status": "quarantined", "eventId": quarantined.ID})
		return
	}
	payload.ConnectionID = string(conn.ID)

	event, err := h.service.EnqueueSePayIncoming(payload)
	if err != nil {
		metrics.Inc("sepay_webhook_failures_total")
		c.JSON(http.StatusBadRequest, gin.H{"code": "SEPAY_BANKHUB_IPN_FAIL", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "eventId": event.ID, "eventState": event.State})
}

// BankHubEvent associates a completed Hosted Link with the resulting bank
// account XID. It is deliberately separate from IPN: SePay documents these as
// different delivery streams with different payloads.
func (h *WealthHandler) BankHubEvent(c *gin.Context) {
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20))
	if err != nil || len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PAYLOAD", "message": "empty webhook payload"})
		return
	}
	if err := h.verifyBankHubAPIKey(c.Request); err != nil {
		metrics.Inc("sepay_webhook_failures_total")
		c.JSON(http.StatusUnauthorized, gin.H{"code": "SEPAY_BANKHUB_UNAUTHORIZED", "message": err.Error()})
		return
	}

	var event struct {
		Event    string `json:"event"`
		Metadata struct {
			BankAccountXID string `json:"bank_account_xid"`
			LinkTokenXID   string `json:"link_token_xid"`
			BrandName      string `json:"brand_name"`
			BankCode       string `json:"bank_code"`
			AccountNumber  string `json:"account_number"`
			SupportsIn     bool   `json:"supports_in"`
			SupportsOut    bool   `json:"supports_out"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	if event.Event == "BANK_ACCOUNT_LINKED" && event.Metadata.BankAccountXID != "" && event.Metadata.LinkTokenXID != "" {
		if userID, found := h.store.GetSePayLinkSessionUser(event.Metadata.LinkTokenXID); found {
			masked := maskBankAccountNumber(event.Metadata.AccountNumber)
			bankCode := strings.ToUpper(strings.TrimSpace(event.Metadata.BankCode))
			_, pilotAllowed := h.pilotBanks[bankCode]
			status := "linked"
			if len(h.pilotBanks) == 0 || !pilotAllowed || !event.Metadata.SupportsIn || !event.Metadata.SupportsOut {
				status = "unsupported"
			}
			_, err := h.store.UpsertSePayBankAccount(domain.SePayBankAccount{UserID: userID, BankAccountXID: event.Metadata.BankAccountXID, BankCode: bankCode, BankName: event.Metadata.BrandName, AccountNumberMasked: masked, SupportsIn: event.Metadata.SupportsIn, SupportsOut: event.Metadata.SupportsOut, Status: status})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": "BANK_ACCOUNT_LINK_FAILED", "message": err.Error()})
				return
			}
			_ = h.store.CompleteSePayLinkSession(event.Metadata.LinkTokenXID)
			if profile, ok := h.store.GetSePayUserProfile(userID); ok {
				profile.Status = "active"
				profile.LinkedAt = time.Now().UTC()
				_, _ = h.store.UpsertSePayUserProfile(*profile)
			}
		} else if conn, ok := h.bankConnectionByExternalID(event.Metadata.LinkTokenXID); ok {
			_ = h.store.UpdateBankConnection(conn.ID, func(item *domain.BankConnection) {
				item.ExternalID = event.Metadata.BankAccountXID
				item.BankCode = event.Metadata.BrandName
				item.Status = "connected"
				item.SyncStatus = "idle"
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"code": "UNKNOWN_LINK_TOKEN", "message": "Bank Hub link token is not known"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func maskBankAccountNumber(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "••••"
	}
	runes := []rune(value)
	if len(runes) <= 4 {
		return "••••" + value
	}
	return "•••• " + string(runes[len(runes)-4:])
}

func (h *WealthHandler) verifyBankHubAPIKey(req *http.Request) error {
	if h.bankHubAPIKey == "" {
		return fmt.Errorf("Bank Hub IPN API key is not configured")
	}
	got := strings.TrimSpace(req.Header.Get("Authorization"))
	const prefix = "Apikey "
	if !strings.HasPrefix(got, prefix) || subtle.ConstantTimeCompare([]byte(strings.TrimPrefix(got, prefix)), []byte(h.bankHubAPIKey)) != 1 {
		return fmt.Errorf("invalid Bank Hub API key")
	}
	return nil
}

func (h *WealthHandler) bankConnectionByExternalID(externalID string) (*domain.BankConnection, bool) {
	key := strings.TrimSpace(externalID)
	if key == "" {
		return nil, false
	}
	for _, connection := range h.store.ListAllBankConnections() {
		if connection.ExternalID == key {
			copy := connection
			return &copy, true
		}
	}
	return nil, false
}

func (h *WealthHandler) parseBankHubIPN(body []byte) (service.SePayWebhookEvent, string, error) {
	var payload struct {
		TransactionDate string `json:"transaction_date"`
		BankAccountXID  string `json:"bank_account_xid"`
		Content         string `json:"content"`
		TransferType    string `json:"transfer_type"`
		Amount          any    `json:"amount"`
		ReferenceCode   string `json:"reference_code"`
		TransactionID   string `json:"transaction_id"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return service.SePayWebhookEvent{}, "", fmt.Errorf("invalid webhook json: %w", err)
	}
	direction := map[string]string{"credit": "in", "debit": "out"}[strings.ToLower(strings.TrimSpace(payload.TransferType))]
	if direction == "" || payload.BankAccountXID == "" || payload.TransactionID == "" {
		return service.SePayWebhookEvent{}, "", fmt.Errorf("bank_account_xid, transaction_id and credit/debit transfer_type are required")
	}
	amount := formatAmountLikeString(payload.Amount)
	if amount == "" {
		return service.SePayWebhookEvent{}, "", fmt.Errorf("amount is required")
	}
	return service.SePayWebhookEvent{
		ProviderAccountID: payload.BankAccountXID,
		Direction:         direction,
		Amount:            amount,
		Currency:          "VND",
		Description:       payload.Content,
		Content:           payload.Content,
		Reference:         payload.ReferenceCode,
		ExternalID:        payload.TransactionID,
		OccurredAt:        payload.TransactionDate,
	}, payload.BankAccountXID, nil
}

func (h *WealthHandler) verifySePayWebhook(body []byte, req *http.Request) error {
	secret := strings.TrimSpace(h.secret)
	if secret == "" {
		return nil
	}

	timestamp := strings.TrimSpace(req.Header.Get("X-SePay-Timestamp"))
	signatureHeader := strings.TrimSpace(req.Header.Get("X-SePay-Signature"))

	if timestamp == "" || signatureHeader == "" {
		return fmt.Errorf("missing required signature headers")
	}

	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid signature timestamp")
	}
	now := time.Now().Unix()
	if now-timestampInt > 300 || timestampInt-now > 300 {
		return fmt.Errorf("signature timestamp outside allowed window")
	}

	expected := signSePayPayload(secret, timestamp, body)

	var signatureRaw string
	for _, p := range strings.Split(signatureHeader, ",") {
		part := strings.TrimSpace(p)
		if strings.HasPrefix(strings.ToLower(part), "sha256=") {
			signatureRaw = part[len("sha256="):]
			break
		}
		if strings.HasPrefix(strings.ToLower(part), "v1=") {
			signatureRaw = strings.TrimPrefix(part, "v1=")
			break
		}
		if strings.HasPrefix(strings.ToLower(part), "sig=") {
			signatureRaw = strings.TrimPrefix(part, "sig=")
			break
		}
	}
	if signatureRaw == "" {
		signatureRaw = signatureHeader
	}

	got, err := hex.DecodeString(signatureRaw)
	if err != nil {
		return fmt.Errorf("invalid signature format")
	}
	if subtle.ConstantTimeCompare(got, expected) != 1 {
		return fmt.Errorf("invalid webhook signature")
	}

	return nil
}

func signSePayPayload(secret, timestamp string, body []byte) []byte {
	hm := hmac.New(sha256.New, []byte(secret))
	_, _ = hm.Write([]byte(timestamp))
	_, _ = hm.Write([]byte("."))
	_, _ = hm.Write(body)
	return hm.Sum(nil)
}

func parseSePayWebhookPayload(body []byte) (service.SePayWebhookEvent, error) {
	var payload struct {
		ConnectionID    string `json:"connectionId"`
		AccountID       string `json:"accountId"`
		Direction       string `json:"direction"`
		TransferType    string `json:"transferType"`
		Amount          any    `json:"amount"`
		TransferAmount  any    `json:"transferAmount"`
		Currency        string `json:"currency"`
		Counterparty    string `json:"counterparty"`
		Description     string `json:"description"`
		Reference       string `json:"reference"`
		Content         string `json:"content"`
		ExternalID      string `json:"externalTransactionId"`
		OccurredAt      string `json:"occurredAt"`
		TransactionDate string `json:"transactionDate"`
		AccountNumber   string `json:"accountNumber"`
		ReferenceCode   string `json:"referenceCode"`
		EventID         string `json:"eventId"`
		ID              any    `json:"id"`
		Code            string `json:"code"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && (payload.Direction != "" || payload.TransferType != "" || payload.ConnectionID != "") {
		ev := service.SePayWebhookEvent{
			ConnectionID: payload.ConnectionID,
			AccountID:    firstNonEmpty(payload.AccountID, payload.AccountNumber),
			Direction:    firstNonEmpty(payload.Direction, payload.TransferType),
			Currency:     firstNonEmpty(payload.Currency, "VND"),
			Counterparty: payload.Counterparty,
			Description:  firstNonEmpty(payload.Description, payload.Content),
			Reference:    firstNonEmpty(payload.Reference, payload.ReferenceCode),
			Content:      payload.Content,
			ExternalID:   firstNonEmpty(payload.ExternalID, payload.EventID, formatAmountLikeString(payload.ID), payload.Code),
			OccurredAt:   firstNonEmpty(payload.OccurredAt, payload.TransactionDate),
		}

		amount := formatAmountLikeString(payload.Amount)
		if amount == "" {
			amount = formatAmountLikeString(payload.TransferAmount)
		}
		if amount != "" {
			ev.Amount = amount
		}
		if ev.Amount == "" {
			return ev, fmt.Errorf("amount is required")
		}
		return ev, nil
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return service.SePayWebhookEvent{}, fmt.Errorf("invalid webhook json: %w", err)
	}

	getStr := func(keys ...string) string {
		for _, key := range keys {
			if v, ok := raw[key]; ok {
				switch vv := v.(type) {
				case string:
					if strings.TrimSpace(vv) != "" {
						return strings.TrimSpace(vv)
					}
				case float64:
					return strconv.FormatFloat(vv, 'f', -1, 64)
				}
			}
		}
		return ""
	}

	amount := getStr("amount", "value", "transactionValue")
	if amount == "" {
		amount = "0"
	}
	ev := service.SePayWebhookEvent{
		ConnectionID: getStr("connectionId", "connection_id", "providerConnectionId"),
		AccountID:    getStr("accountId", "account_id", "accountNumber", "account_number"),
		Direction:    firstNonEmpty(getStr("direction", "transferType"), "out"),
		Amount:       amount,
		Currency:     firstNonEmpty(getStr("currency", "currency_code", "ccy"), "VND"),
		Counterparty: firstNonEmpty(getStr("counterparty", "from", "to"), ""),
		Description:  firstNonEmpty(getStr("description", "content", "note"), ""),
		Reference:    firstNonEmpty(getStr("reference", "ref", "transactionRef"), ""),
		Content:      getStr("content", "description"),
		ExternalID:   firstNonEmpty(getStr("externalTransactionId", "external_id", "transactionId", "txn_id", "id"), ""),
		OccurredAt:   firstNonEmpty(getStr("occurredAt", "occurred_at", "transactionDate", "createdAt"), time.Now().Format(time.RFC3339)),
	}
	if _, err := strconv.ParseFloat(ev.Amount, 64); err != nil {
		return service.SePayWebhookEvent{}, fmt.Errorf("amount must be numeric")
	}
	if ev.ConnectionID == "" {
		return service.SePayWebhookEvent{}, fmt.Errorf("connectionId is required")
	}
	return ev, nil
}

func formatAmountLikeString(v any) string {
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case int:
		return strconv.Itoa(s)
	case int64:
		return strconv.FormatInt(s, 10)
	case json.Number:
		return s.String()
	default:
		return ""
	}
}

func (h *WealthHandler) ListBankFeedTransactions(c *gin.Context) {
	wsID := h.requireUserID(c)
	state := c.Query("state")
	if state == "" {
		state = c.Query("postingState")
	}
	accountID := strings.TrimSpace(c.Query("accountId"))

	var items []domain.BankFeedTransaction

	switch state {
	case "", "all":
		items = h.store.ListBankFeed(domain.ID(wsID))
	case string(domain.PostingStateReview), string(domain.PostingStateAutoReady), string(domain.PostingStatePosted), string(domain.PostingStateIgnored):
		items = h.store.ListBankFeedByState(domain.ID(wsID), domain.TransactionPostingState(state))
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "BAD_REQUEST",
			"message": "invalid state query; must be pending_review, auto_ready, posted, ignored",
		})
		return
	}

	if accountID != "" {
		filtered := make([]domain.BankFeedTransaction, 0, len(items))
		for _, item := range items {
			if string(item.AccountID) == accountID {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	c.JSON(http.StatusOK, items)
}

func (h *WealthHandler) ApproveBankFeed(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	id := c.Param("id")
	feed, ok := h.store.GetBankFeed(domain.ID(id))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "bank feed transaction not found"})
		return
	}
	if !h.requireUserMatch(c, feed.UserID) {
		return
	}
	posted, err := h.service.ApproveBankFeed(domain.ID(id), *feed)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "bank_feed_transaction", domain.ID(id), feed, map[string]any{
		"status":              string("posted"),
		"postedTransactionID": string(posted.ID),
	}, "success", "")
	c.JSON(http.StatusOK, posted)
}

func (h *WealthHandler) ReclassifyBankFeed(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	id := c.Param("id")
	feed, ok := h.store.GetBankFeed(domain.ID(id))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "bank feed transaction not found"})
		return
	}
	if !h.requireUserMatch(c, feed.UserID) {
		return
	}
	var body dto.BankFeedActionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	txType := strings.TrimSpace(strings.ToLower(body.Type))
	if txType != "income" && txType != "expense" && txType != "investment_funding" && txType != "loan_payment" && txType != "loan_disbursement" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "invalid type"})
		return
	}
	accountID := strings.TrimSpace(body.AccountID)
	if accountID == "" {
		accountID = string(feed.AccountID)
	}
	updated, err := h.service.ReclassifyBankFeed(domain.ID(id), domain.ID(accountID), domain.TransactionType(txType), domain.ID(body.CategoryID), strings.TrimSpace(body.Reason))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "bank_feed_transaction", domain.ID(id), feed, updated, "success", "")
	c.JSON(http.StatusOK, updated)
}

func (h *WealthHandler) IgnoreBankFeed(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	id := c.Param("id")
	feed, ok := h.store.GetBankFeed(domain.ID(id))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "bank feed transaction not found"})
		return
	}
	if !h.requireUserMatch(c, feed.UserID) {
		return
	}
	if err := h.store.UpdateFeedState(domain.ID(id), domain.PostingStateIgnored, "ignored_by_user"); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": err.Error()})
		return
	}
	afterFeed, _ := h.store.GetBankFeed(domain.ID(id))
	h.recordAudit(c, "", "bank_feed_transaction", domain.ID(id), feed, afterFeed, "success", "")
	c.JSON(http.StatusOK, gin.H{"status": "ignored", "id": id})
}

func (h *WealthHandler) ListAutomationRules(c *gin.Context) {
	wsID := h.requireUserID(c)
	c.JSON(http.StatusOK, h.store.ListAutomationRules(domain.ID(wsID)))
}

func (h *WealthHandler) CreateAutomationRule(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.AutomationRuleRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	rule, err := h.store.CreateAutomationRule(domain.AutomationRule{
		UserID:           domain.ID(h.requireUserID(c)),
		AccountID:        domain.ID(body.AccountID),
		Name:             body.Name,
		Predicate:        body.Predicate,
		ActionType:       body.ActionType,
		Type:             body.Type,
		CategoryID:       domain.ID(body.CategoryID),
		Priority:         body.Priority,
		Enabled:          body.Enabled,
		Direction:        body.Direction,
		ContentPattern:   body.ContentPattern,
		ReferencePattern: body.ReferencePattern,
		MinAmount:        body.MinAmount,
		MaxAmount:        body.MaxAmount,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "bank_automation_rule", rule.ID, nil, rule, "success", "")
	c.JSON(http.StatusCreated, rule)
}

type automationRulePatch struct {
	Name             *string `json:"name"`
	AccountID        *string `json:"accountId"`
	Predicate        *string `json:"predicate"`
	ActionType       *string `json:"actionType"`
	Direction        *string `json:"direction"`
	Type             *string `json:"type"`
	CategoryID       *string `json:"categoryId"`
	Priority         *int    `json:"priority"`
	Enabled          *bool   `json:"enabled"`
	ContentPattern   *string `json:"contentPattern"`
	ReferencePattern *string `json:"referencePattern"`
	MinAmount        *string `json:"minAmount"`
	MaxAmount        *string `json:"maxAmount"`
}

func (h *WealthHandler) ModifyAutomationRule(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	ruleID := domain.ID(c.Param("id"))
	existing, ok := h.store.GetAutomationRule(ruleID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "rule not found"})
		return
	}
	if !h.requireUserMatch(c, existing.UserID) {
		return
	}
	if c.Request.Method == http.MethodDelete {
		if ok := h.store.DeleteAutomationRule(ruleID); !ok {
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "rule not found"})
			return
		}
		h.recordAudit(c, "", "bank_automation_rule", ruleID, existing, map[string]string{"status": "deleted"}, "success", "")
		c.JSON(http.StatusOK, gin.H{"id": ruleID, "status": "deleted"})
		return
	}
	var body automationRulePatch
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	_ = h.store.UpdateAutomationRule(ruleID, func(rule *domain.AutomationRule) {
		if body.Name != nil {
			rule.Name = strings.TrimSpace(*body.Name)
		}
		if body.AccountID != nil {
			rule.AccountID = domain.ID(strings.TrimSpace(*body.AccountID))
		}
		if body.Direction != nil {
			rule.Direction = strings.TrimSpace(*body.Direction)
		}
		if body.Predicate != nil {
			rule.Predicate = strings.TrimSpace(*body.Predicate)
		}
		if body.ActionType != nil {
			rule.ActionType = strings.TrimSpace(*body.ActionType)
		}
		if body.Type != nil {
			rule.Type = strings.TrimSpace(*body.Type)
		}
		if body.CategoryID != nil {
			rule.CategoryID = domain.ID(strings.TrimSpace(*body.CategoryID))
		}
		if body.ContentPattern != nil {
			rule.ContentPattern = strings.TrimSpace(*body.ContentPattern)
		}
		if body.ReferencePattern != nil {
			rule.ReferencePattern = strings.TrimSpace(*body.ReferencePattern)
		}
		if body.MinAmount != nil {
			rule.MinAmount = strings.TrimSpace(*body.MinAmount)
		}
		if body.MaxAmount != nil {
			rule.MaxAmount = strings.TrimSpace(*body.MaxAmount)
		}
		if body.Priority != nil {
			rule.Priority = *body.Priority
		}
		if body.Enabled != nil {
			rule.Enabled = *body.Enabled
		}
	})
	refreshed, _ := h.store.GetAutomationRule(ruleID)
	h.recordAudit(c, "", "bank_automation_rule", ruleID, existing, refreshed, "success", "")
	c.JSON(http.StatusOK, refreshed)
}

func (h *WealthHandler) PreviewAutomationRule(c *gin.Context) {
	wsID := h.requireUserID(c)
	var payload bankFeedPreviewRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		sample := h.store.ListBankFeed(domain.ID(wsID))
		payload.Sample = sampleToPreviewInput(sample)
	}
	limit := payload.Limit
	if limit <= 0 {
		limit = 10
	}
	samples := make([]domain.BankFeedTransaction, 0, limit)
	for _, row := range payload.Sample {
		samples = append(samples, domain.BankFeedTransaction{
			Direction:    row.Direction,
			Amount:       row.Amount,
			Currency:     row.Currency,
			Description:  row.Description,
			Reference:    row.Reference,
			CounterParty: row.CounterParty,
			AccountID:    rowAccountID(row.AccountID),
			UserID:       domain.ID(wsID),
		})
		if len(samples) >= limit {
			break
		}
	}
	c.JSON(http.StatusOK, h.service.RulePreview(domain.ID(wsID), samples))
}

func (h *WealthHandler) getFeedAccountID(accountID string) domain.ID {
	if strings.TrimSpace(accountID) == "" {
		return ""
	}
	return domain.ID(strings.TrimSpace(accountID))
}

func sampleToPreviewInput(source []domain.BankFeedTransaction) []struct {
	Direction    string `json:"direction"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	Description  string `json:"description"`
	Reference    string `json:"reference"`
	CounterParty string `json:"counterparty"`
	AccountID    string `json:"accountId"`
} {
	out := make([]struct {
		Direction    string `json:"direction"`
		Amount       string `json:"amount"`
		Currency     string `json:"currency"`
		Description  string `json:"description"`
		Reference    string `json:"reference"`
		CounterParty string `json:"counterparty"`
		AccountID    string `json:"accountId"`
	}, 0, len(source))
	for _, row := range source {
		out = append(out, struct {
			Direction    string `json:"direction"`
			Amount       string `json:"amount"`
			Currency     string `json:"currency"`
			Description  string `json:"description"`
			Reference    string `json:"reference"`
			CounterParty string `json:"counterparty"`
			AccountID    string `json:"accountId"`
		}{
			Direction:    row.Direction,
			Amount:       row.Amount,
			Currency:     row.Currency,
			Description:  row.Description,
			Reference:    row.Reference,
			CounterParty: row.CounterParty,
			AccountID:    string(row.AccountID),
		})
	}
	return out
}

func rowAccountID(accountID string) domain.ID {
	if strings.TrimSpace(accountID) == "" {
		return ""
	}
	return domain.ID(strings.TrimSpace(accountID))
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

func (h *WealthHandler) CreatePaymentRequest(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	loanID := strings.TrimSpace(c.Param("id"))
	if loanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "loanId is required"})
		return
	}
	loan, ok := h.store.GetLoan(domain.ID(loanID))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "loan not found"})
		return
	}
	if !h.requireUserMatch(c, loan.UserID) {
		return
	}
	var body loanPaymentRequestPayload
	_ = c.ShouldBindJSON(&body)
	created, err := h.service.CreateLoanPaymentRequest(loan.UserID, domain.ID(loanID), service.PaymentRequestCreate{
		Amount:    body.Amount,
		Currency:  body.Currency,
		ExpiresAt: body.ExpiresAt,
		Note:      body.Note,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "bank_payment_request", created.ID, nil, created, "success", "")
	qr := fmt.Sprintf("WOS://VNPAY/%s?amount=%s", created.Code, created.Amount)
	response := map[string]any{
		"loanId":      loanID,
		"paymentCode": created.Code,
		"amount":      created.Amount,
		"currency":    created.Currency,
		"expiresAt":   created.ExpiresAt,
		"status":      created.Status,
		"qr":          qr,
		"id":          created.ID,
	}
	c.JSON(http.StatusCreated, response)
}

func (h *WealthHandler) TelegramWebhook(c *gin.Context) {
	if !h.validateTelegramWebhookSecret(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "invalid telegram webhook secret"})
		return
	}

	var payload telegramWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}

	action, commandID, approvalID, text, userID, _ := h.extractTelegramCommandAndIdentity(c, payload)
	if action != "" && commandID != "" {
		if commandID == "" || approvalID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "invalid callback payload"})
			return
		}
		cmd, ok := h.store.GetAssistantCommand(domain.ID(commandID))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "command not found"})
			return
		}
		if userID == "" || strings.TrimSpace(string(cmd.UserID)) != userID {
			c.JSON(http.StatusForbidden, gin.H{"code": "USER_MISMATCH", "message": "command does not belong to user"})
			return
		}
		if userID != "" && string(cmd.UserID) != "" && userID != string(cmd.UserID) {
			c.JSON(http.StatusForbidden, gin.H{"code": "USER_MISMATCH", "message": "callback actor does not match command owner"})
			return
		}
		switch action {
		case "approve":
			_, err := h.approveAssistantCommand(cmd, approvalID, userID)
			if err != nil {
				switch err.(type) {
				case approvalConflictError:
					c.JSON(http.StatusConflict, gin.H{"code": "APPROVAL_REJECTED", "message": err.Error()})
				case approvalNotFoundError:
					c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": err.Error()})
				case approvalExpiredError:
					c.JSON(http.StatusConflict, gin.H{"code": "APPROVAL_EXPIRED", "message": err.Error()})
				case approvalInvalidError:
					c.JSON(http.StatusUnauthorized, gin.H{"code": "INVALID_APPROVAL", "message": err.Error()})
				default:
					c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
				}
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"status":    assistantStatusDispatched,
				"commandId": commandID,
			})
		case "reject":
			now := time.Now().UTC()
			updated, err := h.store.UpdateAssistantCommand(domain.ID(commandID), func(current *domain.AssistantCommand) {
				current.Status = assistantStatusCancelled
				current.ApprovalUsedAt = &now
				current.ApprovalID = ""
			})
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": err.Error()})
				return
			}
			c.JSON(http.StatusOK, updated)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "unknown callback action"})
			return
		}
		return
	}

	if strings.TrimSpace(text) == "" {
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "no command text found"})
		return
	}

	if userID == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "USER_LINK_REQUIRED", "message": "telegram chat is not linked to a user"})
		return
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		uid = "telegram-unbound"
	}
	intent := h.classifyAssistantIntent(text, "")
	status := h.initialStatusForAssistantIntent(intent)

	command, err := h.store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:  domain.ID(userID),
		Command: text,
		Plan:    intent,
		Status:  status,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}

	out := gin.H{
		"status":    "received",
		"commandId": command.ID,
		"userId":    userID,
	}
	if command.ApprovalID != "" {
		out["approvalId"] = command.ApprovalID
		out["approvalExpiresAt"] = command.ApprovalExpiresAt
	}
	c.JSON(http.StatusCreated, out)
}

func (h *WealthHandler) CreateAssistantCommand(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body dto.AssistantCommandRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	wsID := domain.ID(h.requireUserID(c))
	intent := h.classifyAssistantIntent(body.Command, body.Plan)
	status := h.initialStatusForAssistantIntent(intent)
	command, err := h.store.CreateAssistantCommand(domain.AssistantCommand{
		UserID:  wsID,
		Command: body.Command,
		Plan:    intent,
		Status:  status,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "assistant_command", command.ID, nil, command, "success", "")
	c.JSON(http.StatusCreated, command)
}

func (h *WealthHandler) ListAssistantCommands(c *gin.Context) {
	wsID := h.requireUserID(c)
	commands := h.store.ListAssistantCommands(domain.ID(wsID))
	c.JSON(http.StatusOK, commands)
}

func (h *WealthHandler) ListAuditLogs(c *gin.Context) {
	wsID := h.requireUserID(c)
	if wsID == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing user"})
		return
	}
	logs := h.store.ListAuditLogs(domain.ID(wsID))
	c.JSON(http.StatusOK, logs)
}

func (h *WealthHandler) GetAssistantCommand(c *gin.Context) {
	cmd, ok := h.store.GetAssistantCommand(domain.ID(c.Param("id")))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "command not found"})
		return
	}
	if !h.requireUserMatch(c, cmd.UserID) {
		return
	}
	c.JSON(http.StatusOK, cmd)
}

func (h *WealthHandler) ApproveCommand(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body assistantCommandApprovalRequest
	_ = c.ShouldBindJSON(&body)
	approvalID := strings.TrimSpace(body.ApprovalID)
	if approvalID == "" {
		approvalID = strings.TrimSpace(c.Query("approvalId"))
	}
	if approvalID == "" {
		approvalID = strings.TrimSpace(c.GetHeader("x-approval-id"))
	}
	cmdID := domain.ID(c.Param("id"))
	cmd, ok := h.store.GetAssistantCommand(cmdID)
	if ok {
		if !h.requireUserMatch(c, cmd.UserID) {
			return
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "command not found"})
		return
	}
	if cmd.Status != assistantStatusAwaitingApproval {
		c.JSON(http.StatusConflict, gin.H{"code": "INVALID_TRANSITION", "message": "command cannot be approved from current status"})
		return
	}
	updated, err := h.approveAssistantCommand(cmd, approvalID, currentUser(c))
	if err != nil {
		switch err.(type) {
		case approvalConflictError:
			c.JSON(http.StatusConflict, gin.H{"code": "INVALID_TRANSITION", "message": err.Error()})
		case approvalNotFoundError:
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": err.Error()})
		case approvalExpiredError:
			c.JSON(http.StatusConflict, gin.H{"code": "APPROVAL_EXPIRED", "message": err.Error()})
		case approvalInvalidError:
			c.JSON(http.StatusUnauthorized, gin.H{"code": "INVALID_APPROVAL", "message": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		}
		return
	}
	h.recordAudit(c, "", "assistant_command", cmdID, cmd, updated, "success", "")
	c.JSON(http.StatusOK, updated)
}

func (h *WealthHandler) CancelCommand(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	cmdID := domain.ID(c.Param("id"))
	cmd, ok := h.store.GetAssistantCommand(cmdID)
	if ok {
		if !h.requireUserMatch(c, cmd.UserID) {
			return
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "command not found"})
		return
	}
	next, ok := h.nextStatusAfterCancel(cmd.Status)
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"code": "INVALID_TRANSITION", "message": "command cannot be cancelled from current status"})
		return
	}
	cmd, err := h.store.UpdateAssistantCommand(cmdID, func(current *domain.AssistantCommand) {
		current.Status = next
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "assistant_command", cmdID, cmd, cmd, "success", "")
	c.JSON(http.StatusOK, cmd)
}

func (h *WealthHandler) ExecutorEvents(c *gin.Context) {
	if !h.validateHermesExecutorSecret(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "invalid hermes executor secret"})
		return
	}

	var payload hermesExecutorEvent
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}

	commandID := strings.TrimSpace(payload.CommandID)
	if commandID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_EVENT", "message": "commandId is required"})
		return
	}

	target := h.mapHermesEventStatus(payload.Status)
	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_EVENT", "message": "unsupported status"})
		return
	}

	cmd, ok := h.store.GetAssistantCommand(domain.ID(commandID))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "command not found"})
		return
	}

	next, ok := h.nextStatusFromExecutorEvent(cmd.Status, target)
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"code": "INVALID_TRANSITION", "message": "executor event cannot be applied in current status"})
		return
	}

	updated, err := h.store.UpdateAssistantCommand(cmd.ID, func(current *domain.AssistantCommand) {
		current.Status = next
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"executorId": c.Param("id"),
		"commandId":  updated.ID,
		"status":     updated.Status,
	})
}

func (h *WealthHandler) validateHermesExecutorSecret(c *gin.Context) bool {
	if strings.TrimSpace(h.hermesSecret) == "" {
		return true
	}
	headerSecret := strings.TrimSpace(c.GetHeader("X-Hermes-Secret"))
	if headerSecret == "" {
		headerSecret = strings.TrimSpace(c.GetHeader("X-Hermes-WebHook-Secret"))
	}
	if headerSecret == "" {
		headerSecret = strings.TrimSpace(c.GetHeader("X-Webhook-Secret"))
	}
	return subtle.ConstantTimeCompare([]byte(headerSecret), []byte(h.hermesSecret)) == 1
}

func (h *WealthHandler) mapHermesEventStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "accepted", "event_accepted", "acknowledged":
		return assistantStatusDispatched
	case "started", "in_progress", "running", "progress":
		return assistantStatusRunning
	case "completed":
		return assistantStatusCompleted
	case "failed":
		return assistantStatusFailed
	case "timed_out", "timeout", "timedout", "expired":
		return assistantStatusTimedOut
	case "cancelled":
		return assistantStatusCancelled
	case "rejected":
		return assistantStatusRejected
	default:
		return ""
	}
}

func (h *WealthHandler) classifyAssistantIntent(command string, plan string) string {
	normalizedPlan := strings.TrimSpace(strings.ToLower(plan))
	switch normalizedPlan {
	case assistantIntentRead, assistantIntentDraft, assistantIntentWrite, assistantIntentExternalAction, assistantIntentBlocked:
		return normalizedPlan
	}

	text := strings.ToLower(strings.TrimSpace(command))
	if text == "" {
		return assistantIntentRead
	}

	if containsAny(text, assistantBlockedKeywords) {
		return assistantIntentBlocked
	}
	if containsAny(text, assistantExternalKeywords) {
		return assistantIntentExternalAction
	}
	if containsAny(text, assistantWriteKeywords) {
		return assistantIntentWrite
	}
	if containsAny(text, assistantDraftKeywords) {
		return assistantIntentDraft
	}
	if containsAny(text, assistantReadKeywords) {
		return assistantIntentRead
	}
	return assistantIntentRead
}

func (h *WealthHandler) initialStatusForAssistantIntent(intent string) string {
	switch intent {
	case assistantIntentWrite, assistantIntentExternalAction:
		return assistantStatusAwaitingApproval
	case assistantIntentBlocked:
		return assistantStatusRejected
	default:
		return assistantStatusPlanned
	}
}

func containsAny(text string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}

func (h *WealthHandler) isTerminalAssistantStatus(status string) bool {
	switch status {
	case assistantStatusCompleted, assistantStatusFailed, assistantStatusTimedOut, assistantStatusCancelled, assistantStatusRejected:
		return true
	default:
		return false
	}
}

type approvalInvalidError string

func (e approvalInvalidError) Error() string { return string(e) }

type approvalExpiredError string

func (e approvalExpiredError) Error() string { return string(e) }

type approvalConflictError string

func (e approvalConflictError) Error() string { return string(e) }

type approvalNotFoundError string

func (e approvalNotFoundError) Error() string { return string(e) }

func (h *WealthHandler) approveAssistantCommand(cmd *domain.AssistantCommand, approvalID, actor string) (*domain.AssistantCommand, error) {
	if cmd == nil {
		return nil, approvalNotFoundError("command not found")
	}
	if cmd.Status != assistantStatusAwaitingApproval {
		return nil, approvalConflictError("command cannot be approved from current status")
	}
	if strings.TrimSpace(approvalID) == "" {
		return nil, approvalInvalidError("approval id is required")
	}
	if cmd.ApprovalID == "" {
		return nil, approvalConflictError("command already approved or does not require approval")
	}
	if strings.TrimSpace(cmd.ApprovalID) != strings.TrimSpace(approvalID) {
		return nil, approvalInvalidError("invalid approval id")
	}
	if cmd.ApprovalUsedAt != nil {
		return nil, approvalConflictError("approval token already used")
	}
	if !cmd.ApprovalExpiresAt.IsZero() && nowOrUTC(cmd.ApprovalExpiresAt).Before(nowOrUTC(time.Now())) {
		return nil, approvalExpiredError("approval token expired")
	}
	now := nowOrUTC(time.Now())
	_ = actor
	updated, err := h.store.UpdateAssistantCommand(cmd.ID, func(current *domain.AssistantCommand) {
		current.Status = assistantStatusDispatched
		current.ApprovalUsedAt = &now
		current.ApprovalID = ""
	})
	return updated, err
}

func (h *WealthHandler) nextStatusAfterCancel(current string) (string, bool) {
	switch current {
	case assistantStatusCancelled:
		return assistantStatusCancelled, true
	case assistantStatusCompleted, assistantStatusFailed, assistantStatusTimedOut, assistantStatusRejected:
		return current, false
	default:
		return assistantStatusCancelled, true
	}
}

func (h *WealthHandler) nextStatusFromExecutorEvent(current, target string) (string, bool) {
	if h.isTerminalAssistantStatus(current) {
		if current == target {
			return current, true
		}
		return "", false
	}

	switch current {
	case assistantStatusPending, assistantStatusReceived, assistantStatusPlanned, assistantStatusAwaitingApproval, assistantStatusApproved, assistantStatusDispatched:
		switch target {
		case assistantStatusDispatched, assistantStatusRunning, assistantStatusCompleted, assistantStatusFailed, assistantStatusTimedOut, assistantStatusCancelled, assistantStatusRejected:
			return target, true
		default:
			return "", false
		}
	case assistantStatusRunning:
		switch target {
		case assistantStatusRunning, assistantStatusCompleted, assistantStatusFailed, assistantStatusTimedOut, assistantStatusCancelled, assistantStatusRejected:
			return target, true
		default:
			return "", false
		}
	default:
		return "", false
	}
}

func (h *WealthHandler) validateTelegramWebhookSecret(c *gin.Context) bool {
	if strings.TrimSpace(h.telegramSecret) == "" {
		return true
	}
	headerSecret := c.GetHeader("X-Telegram-Bot-Api-Secret-Token")
	if headerSecret == "" {
		headerSecret = c.GetHeader("X-Webhook-Secret")
	}
	return headerSecret == h.telegramSecret
}

func (h *WealthHandler) extractTelegramCommandAndIdentity(c *gin.Context, payload telegramWebhookPayload) (string, string, string, string, string, string) {
	userID := strings.TrimSpace(c.Query("userId"))
	if userID == "" {
		userID = strings.TrimSpace(c.GetHeader("x-user-id"))
	}
	if userID == "" {
		userID = strings.TrimSpace(payload.UserID)
	}

	var (
		action         string
		cmdID          string
		apprID         string
		text           string
		telegramUserID string
		sourceMsg      *telegramWebhookMessage
	)

	switch {
	case payload.CallbackQuery != nil:
		payloadText := strings.TrimSpace(payload.CallbackQuery.Data)
		action, cmdID, apprID = parseTelegramCallbackApproval(payloadText)
		if action == "" {
			text = payloadText
		}
		if payload.CallbackQuery.From != nil {
			telegramUserID = strconv.FormatInt(payload.CallbackQuery.From.ID, 10)
		}
		sourceMsg = payload.CallbackQuery.Message
	case payload.Message != nil:
		sourceMsg = payload.Message
		text = strings.TrimSpace(payload.Message.Text)
	case payload.EditedMessage != nil:
		sourceMsg = payload.EditedMessage
		text = strings.TrimSpace(payload.EditedMessage.Text)
	case payload.ChannelPost != nil:
		sourceMsg = payload.ChannelPost
		text = strings.TrimSpace(payload.ChannelPost.Text)
	}

	if text == "" {
		if action != "" {
			return action, cmdID, apprID, "", userID, telegramUserID
		}
		return "", "", "", "", userID, telegramUserID
	}

	if telegramUserID == "" && sourceMsg != nil && sourceMsg.From != nil {
		telegramUserID = strconv.FormatInt(sourceMsg.From.ID, 10)
	}
	return action, cmdID, apprID, text, userID, telegramUserID
}

func parseTelegramCallbackApproval(data string) (string, string, string) {
	parts := strings.Split(data, ":")
	if len(parts) != 3 {
		return "", "", ""
	}
	action := strings.ToLower(strings.TrimSpace(parts[0]))
	if action != "approve" && action != "reject" {
		return "", "", ""
	}
	return action, strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])
}

func (h *WealthHandler) NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "route not found"})
}

func (h *WealthHandler) requireUserID(c *gin.Context) string {
	if v, ok := c.Get("user_id"); ok {
		if ws, ok2 := v.(string); ok2 && strings.TrimSpace(ws) != "" {
			return strings.TrimSpace(ws)
		}
	}

	headerWs := strings.TrimSpace(c.GetHeader("x-user-id"))
	if headerWs != "" {
		return headerWs
	}
	userID := c.Query("userId")
	if userID != "" {
		return userID
	}
	uid := currentUser(c)
	if uid == "" {
		return ""
	}
	return uid
}

func (h *WealthHandler) requireUserOrReject(c *gin.Context) (string, bool) {
	ws := h.requireUserID(c)
	if ws == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing user context"})
		return "", false
	}
	return ws, true
}

func (h *WealthHandler) requireUserMatch(c *gin.Context, resourceUserID domain.ID) bool {
	wsID, ok := h.requireUserOrReject(c)
	if !ok {
		return false
	}
	if strings.TrimSpace(string(resourceUserID)) != wsID {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "user mismatch"})
		return false
	}
	return true
}

func currentUser(c *gin.Context) string {
	if val, ok := c.Get("user_id"); ok {
		if uid, ok2 := val.(string); ok2 {
			return uid
		}
	}
	return ""
}

func defaultUserName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "Tài khoản cá nhân"
	}
	return name
}

func nowOrUTC(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t.UTC()
}

func parseAsOf(raw string) (time.Time, error) {
	t, err := parseDateTime(raw)
	if err != nil {
		return time.Time{}, errors.New("invalid asOf format")
	}
	return t, nil
}

func parseDateFilter(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	t, err := parseDateTime(raw)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseDateTime(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, errors.New("empty date")
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, errors.New("empty date")
	}

	zoneAwareLayouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02T15:04Z07:00",
		"2006-01-02 15:04Z07:00",
		"2006-01-02T15:04-07:00",
		"2006-01-02 15:04-07:00",
	}
	for _, layout := range zoneAwareLayouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC(), nil
		}
	}

	naiveLayouts := []string{
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"2006-01-02T15:04:05Z07:00:00",
	}
	for _, layout := range naiveLayouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC(), nil
		}
	}

	if unix, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return time.Unix(unix, 0).UTC(), nil
	}

	return time.Time{}, errors.New("invalid date format")
}

func sortTransactionsForCursor(items []domain.Transaction) []domain.Transaction {
	sort.Slice(items, func(i, j int) bool {
		if items[i].OccurredAt.Equal(items[j].OccurredAt) {
			return items[i].ID > items[j].ID
		}
		return items[i].OccurredAt.After(items[j].OccurredAt)
	})
	return items
}

func (h *WealthHandler) recordAudit(c *gin.Context, action string, targetType string, targetID domain.ID, before any, after any, result string, reason string) {
	if h == nil || h.store == nil || c == nil {
		return
	}

	wsID := h.resolveAuditUserID(c, targetID)
	if wsID == "" {
		return
	}

	actorID := strings.TrimSpace(currentUser(c))
	if actorID != "" {
		if _, err := uuid.Parse(actorID); err != nil {
			actorID = ""
		}
	}

	requestID := strings.TrimSpace(h.requestID(c))
	if requestID == "" {
		requestID = uuid.NewString()
	}

	method := "UNKNOWN"
	path := ""
	if c.Request != nil {
		method = strings.ToUpper(c.Request.Method)
		path = c.Request.URL.Path
	}
	if path == "" {
		path = c.FullPath()
	}
	if method == "" || strings.EqualFold(method, "UNKNOWN") {
		method = "POST"
	}

	action = h.normalizeAuditAction(method, path, targetType, action)
	policyDecision := "allowed"
	if strings.EqualFold(strings.TrimSpace(result), "failure") || strings.EqualFold(strings.TrimSpace(result), "error") {
		policyDecision = "denied"
	}

	entry := domain.AuditLog{
		UserID:         wsID,
		ActorID:        domain.ID(actorID),
		ActorRole:      h.userRole(c),
		Action:         strings.TrimSpace(action),
		TargetType:     strings.TrimSpace(targetType),
		TargetID:       targetID,
		RequestID:      requestID,
		Path:           path,
		Method:         strings.ToUpper(method),
		PolicyDecision: policyDecision,
		Result:         strings.TrimSpace(result),
		Reason:         strings.TrimSpace(reason),
		CorrelationID:  h.correlationID(c),
		BeforeJSON:     marshalAuditPayload(before),
		AfterJSON:      marshalAuditPayload(after),
	}
	if strings.TrimSpace(entry.Result) == "" {
		entry.Result = "success"
	}
	// Audit must never hold up the request path. The complete, immutable audit
	// payload is captured above, then persisted by an independent worker goroutine.
	store := h.store
	go func(item domain.AuditLog) {
		if _, err := store.CreateAuditLog(item); err != nil {
			log.Printf("audit persistence failed: action=%s target=%s err=%v", item.Action, item.TargetID, err)
		}
	}(entry)
}

func (h *WealthHandler) resolveAuditUserID(c *gin.Context, targetID domain.ID) domain.ID {
	if explicit, ok := c.Get("user_id"); ok {
		if raw, ok := explicit.(string); ok && strings.TrimSpace(raw) != "" {
			return domain.ID(strings.TrimSpace(raw))
		}
	}
	if explicit := strings.TrimSpace(c.GetHeader("x-user-id")); explicit != "" {
		return domain.ID(explicit)
	}
	if explicit := strings.TrimSpace(c.Query("userId")); explicit != "" {
		return domain.ID(explicit)
	}
	if targetID != "" {
		// Keep best effort resilience if handlers log object IDs that are guaranteed to belong to the user.
		// Lookup is best effort and currently unsupported by interface for all target types.
	}
	if uid := strings.TrimSpace(currentUser(c)); uid != "" && h != nil && h.store != nil {
		return domain.ID(uid)
	}
	return ""
}

func (h *WealthHandler) normalizeAuditAction(method, path, targetType, explicit string) string {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" {
		return explicit
	}
	pathLower := strings.ToLower(path)
	methodUpper := strings.ToUpper(method)
	switch methodUpper {
	case http.MethodPost:
		switch {
		case targetType == "bank_connection" && strings.Contains(pathLower, "/revoke"):
			return "revoke_bank_connection"
		case targetType == "bank_connection" && strings.Contains(pathLower, "/sync"):
			return "sync_bank_connection"
		case targetType == "bank_connection" && strings.Contains(pathLower, "/connect"):
			return "create_bank_connection"
		case targetType == "forecast_scenario" && strings.Contains(pathLower, "/run"):
			return "run_forecast_scenario"
		case targetType == "assistant_command" && strings.Contains(pathLower, "/approve"):
			return "approve_assistant_command"
		case targetType == "assistant_command" && strings.Contains(pathLower, "/cancel"):
			return "cancel_assistant_command"
		case targetType == "bank_feed_transaction" && strings.Contains(pathLower, "/approve"):
			return "approve_bank_feed_transaction"
		case targetType == "bank_feed_transaction" && strings.Contains(pathLower, "/reclassify"):
			return "reclassify_bank_feed_transaction"
		case targetType == "bank_feed_transaction" && strings.Contains(pathLower, "/ignore"):
			return "ignore_bank_feed_transaction"
		case targetType == "bank_payment_request" && strings.Contains(pathLower, "/payment-requests"):
			return "create_bank_payment_request"
		default:
			return "create_" + targetType
		}
	case http.MethodPut:
		return "upsert_" + targetType
	case http.MethodPatch:
		return "update_" + targetType
	case http.MethodDelete:
		return "delete_" + targetType
	case http.MethodGet:
		return "read_" + targetType
	default:
		return strings.ToLower(strings.TrimSpace(methodUpper)) + "_" + targetType
	}
}

func (h *WealthHandler) userRole(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if raw, ok := c.Get("user_role"); ok {
		if role, ok := raw.(string); ok {
			return strings.TrimSpace(role)
		}
	}
	return ""
}

func (h *WealthHandler) requestID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if value, ok := c.Get("request_id"); ok {
		if requestID, ok := value.(string); ok {
			return strings.TrimSpace(requestID)
		}
	}
	if c.Request != nil {
		if headerValue := strings.TrimSpace(c.GetHeader("X-Request-ID")); headerValue != "" {
			return headerValue
		}
		if headerValue := strings.TrimSpace(c.GetHeader("X-Request-Id")); headerValue != "" {
			return headerValue
		}
	}
	return ""
}

func (h *WealthHandler) correlationID(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if value := strings.TrimSpace(c.GetHeader("X-Correlation-ID")); value != "" {
		return value
	}
	if value := strings.TrimSpace(c.GetHeader("X-Correlation-Id")); value != "" {
		return value
	}
	return ""
}

func marshalAuditPayload(v any) string {
	if v == nil {
		return ""
	}
	switch typed := v.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(raw)
	}
}

func (h *WealthHandler) getAmountDisplayMode(c *gin.Context) string {
	if header := strings.TrimSpace(c.GetHeader("X-Amount-Display-Mode")); header != "" {
		if header == string(domain.AmountDisplayModeCompact) || header == string(domain.AmountDisplayModeFull) {
			return header
		}
	}
	userID := c.GetString("user_id")
	if userID != "" {
		st, err := h.service.GetUserSettings(c.Request.Context(), domain.ID(userID))
		if err == nil && st != nil && st.AmountDisplayMode != "" {
			return string(st.AmountDisplayMode)
		}
	}
	return string(domain.AmountDisplayModeFull)
}

func (h *WealthHandler) GetUserSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "user context missing"})
		return
	}
	st, err := h.service.GetUserSettings(c.Request.Context(), domain.ID(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.UserSettingsResponse{
		UserID:            string(st.UserID),
		AmountDisplayMode: string(st.AmountDisplayMode),
		UpdatedAt:         st.UpdatedAt.Format(time.RFC3339),
	})
}

func (h *WealthHandler) UpdateUserSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "user context missing"})
		return
	}
	var body dto.UserSettingsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	st, err := h.service.UpdateUserSettings(c.Request.Context(), domain.ID(userID), domain.AmountDisplayMode(body.AmountDisplayMode))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.UserSettingsResponse{
		UserID:            string(st.UserID),
		AmountDisplayMode: string(st.AmountDisplayMode),
		UpdatedAt:         st.UpdatedAt.Format(time.RFC3339),
	})
}

// GetMySePay exposes only the authenticated user's provider state. Shared
// user mappings are included as references, never as another user's
// bank connection data.
func (h *WealthHandler) GetMySePay(c *gin.Context) {
	userID := domain.ID(currentUser(c))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	profile, _ := h.store.GetSePayUserProfile(userID)
	accounts := h.store.ListSePayBankAccounts(userID)
	type accountView struct {
		domain.SePayBankAccount
		Mapping *domain.BankAccountMapping `json:"mapping,omitempty"`
	}
	items := make([]accountView, 0, len(accounts))
	for _, account := range accounts {
		mapping, _ := h.store.GetBankAccountMapping(account.ID)
		items = append(items, accountView{SePayBankAccount: account, Mapping: mapping})
	}
	c.JSON(http.StatusOK, gin.H{"profile": profile, "bankAccounts": items, "sharingWarning": "Giao dịch được chia sẻ với thành viên có quyền trong user đã map."})
}

func (h *WealthHandler) CreateMySePayLinkSession(c *gin.Context) {
	userID := domain.ID(currentUser(c))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	if h.bankHub == nil || !h.bankHub.Configured() || h.bankHubCompany == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "SEPAY_BANKHUB_UNAVAILABLE", "message": "Bank Hub is not configured"})
		return
	}
	companyXID := h.bankHubCompany
	if profile, ok := h.store.GetSePayUserProfile(userID); ok && strings.TrimSpace(profile.CompanyXID) != "" {
		companyXID = profile.CompanyXID
	}
	link, err := h.bankHub.CreateLink(c.Request.Context(), companyXID, "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": "SEPAY_BANKHUB_UNAVAILABLE", "message": err.Error()})
		return
	}
	expiresAt, _ := parseDateTime(link.ExpiresAt)
	if err := h.store.CreateSePayLinkSession(link.XID, userID, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "SEPAY_LINK_SESSION_FAILED", "message": err.Error()})
		return
	}
	_, _ = h.store.UpsertSePayUserProfile(domain.SePayUserProfile{UserID: userID, CompanyXID: companyXID, Status: "link_pending"})
	c.JSON(http.StatusCreated, gin.H{"hosted_link_url": link.HostedLinkURL, "link_session_xid": link.XID, "expires_at": link.ExpiresAt})
}

func (h *WealthHandler) ListMySePayBankAccounts(c *gin.Context) { h.GetMySePay(c) }

// SyncMySePayBankAccounts is invoked only after the Hosted Link web view
// emits its completion event. It never trusts a client-provided provider XID;
// it resolves the account from Bank Hub with the server bearer token and saves
// only exact account-number matches for this user's company profile.
func (h *WealthHandler) SyncMySePayBankAccounts(c *gin.Context) {
	userID := domain.ID(currentUser(c))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	if h.bankHub == nil || !h.bankHub.Configured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "SEPAY_BANKHUB_UNAVAILABLE"})
		return
	}
	var body dto.SePayBankAccountSyncRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.AccountNumber) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": "accountNumber is required"})
		return
	}
	profile, ok := h.store.GetSePayUserProfile(userID)
	if !ok || strings.TrimSpace(profile.CompanyXID) == "" {
		c.JSON(http.StatusConflict, gin.H{"code": "SEPAY_PROFILE_NOT_FOUND"})
		return
	}
	accounts, err := h.bankHub.ListBankAccounts(c.Request.Context(), profile.CompanyXID, body.AccountNumber)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": "SEPAY_BANKHUB_UNAVAILABLE", "message": err.Error()})
		return
	}
	matched := make([]domain.SePayBankAccount, 0, len(accounts))
	want := normalizeBankAccountNumber(body.AccountNumber)
	for _, providerAccount := range accounts {
		if providerAccount.XID == "" || normalizeBankAccountNumber(providerAccount.AccountNumber) != want {
			continue
		}
		bankCode := strings.ToUpper(strings.TrimSpace(providerAccount.BankCode))
		_, pilotAllowed := h.pilotBanks[bankCode]
		connected, active := providerFlag(providerAccount.BankAPIConnected), providerFlag(providerAccount.Active)
		supports := connected && active && len(h.pilotBanks) > 0 && pilotAllowed
		status := "unsupported"
		if supports {
			status = "linked"
		}
		item, upsertErr := h.store.UpsertSePayBankAccount(domain.SePayBankAccount{UserID: userID, BankAccountXID: providerAccount.XID, BankCode: bankCode, BankName: providerAccount.BrandName, AccountNumberMasked: maskBankAccountNumber(providerAccount.AccountNumber), SupportsIn: supports, SupportsOut: supports, Status: status})
		if upsertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "BANK_ACCOUNT_SYNC_FAILED", "message": upsertErr.Error()})
			return
		}
		matched = append(matched, item)
	}
	if len(matched) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": "BANK_ACCOUNT_NOT_FOUND", "message": "SePay has not finished linking this account yet"})
		return
	}
	profile.Status, profile.LinkedAt, profile.LastSyncedAt = "active", time.Now().UTC(), time.Now().UTC()
	_, _ = h.store.UpsertSePayUserProfile(*profile)
	c.JSON(http.StatusOK, gin.H{"bankAccounts": matched})
}

func normalizeBankAccountNumber(value string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, value)
}

func providerFlag(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(strings.TrimSpace(v), "true") || strings.EqualFold(strings.TrimSpace(v), "active") || strings.TrimSpace(v) == "1"
	case float64:
		return v != 0
	default:
		return false
	}
}

func (h *WealthHandler) MapMySePayBankAccount(c *gin.Context) {
	userID := domain.ID(currentUser(c))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	bankAccount, ok := h.store.GetSePayBankAccount(domain.ID(c.Param("id")))
	if !ok || bankAccount.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "bank account not found"})
		return
	}
	if bankAccount.Status != "linked" || !bankAccount.SupportsIn || !bankAccount.SupportsOut {
		c.JSON(http.StatusConflict, gin.H{"code": "BANK_ACCOUNT_UNSUPPORTED", "message": "bank must be enabled for the two-way pilot before mapping"})
		return
	}
	var body dto.SePayBankAccountMapRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.AccountID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": "accountId is required"})
		return
	}
	account, ok := h.store.GetAccount(domain.ID(body.AccountID))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Finora account not found"})
		return
	}
	if account.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "editor or owner role is required for this user"})
		return
	}
	mapping, err := h.store.UpsertBankAccountMapping(domain.BankAccountMapping{SePayBankAccountID: bankAccount.ID, UserID: account.UserID, AccountID: account.ID, Status: "active"})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	// Compatibility bridge for the durable ingress worker. It is never exposed
	// as a user connection; the new mapping remains the ownership authority.
	found := false
	for _, conn := range h.store.ListBankConnections(account.UserID) {
		if conn.Provider == "sepay" && conn.ExternalID == bankAccount.BankAccountXID {
			found = true
			break
		}
	}
	if !found {
		_, _ = h.store.CreateBankConnection(domain.BankConnection{UserID: account.UserID, Provider: "sepay", ExternalID: bankAccount.BankAccountXID, Status: "connected", Scope: "read_transactions", SyncStatus: "idle"})
	}
	h.recordAudit(c, "", "sepay_bank_account_mapping", mapping.ID, nil, mapping, "success", "")
	c.JSON(http.StatusOK, mapping)
}

func (h *WealthHandler) UnlinkMySePayBankAccount(c *gin.Context) {
	userID := domain.ID(currentUser(c))
	account, ok := h.store.GetSePayBankAccount(domain.ID(c.Param("id")))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	if !ok || account.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		return
	}
	if mapping, ok := h.store.GetBankAccountMapping(account.ID); ok {
		_ = h.store.DeactivateBankAccountMapping(account.ID)
		h.recordAudit(c, "", "sepay_bank_account_mapping", mapping.ID, mapping, map[string]string{"status": "inactive"}, "success", "")
	}
	_ = h.store.SetSePayBankAccountStatus(account.ID, "unlinked")
	c.JSON(http.StatusOK, gin.H{"status": "unlinked", "bankAccountId": account.ID})
}

func (h *WealthHandler) myMappedFeed(userID domain.ID) map[domain.ID]domain.BankAccountMapping {
	allowed := map[domain.ID]domain.BankAccountMapping{}
	for _, bankAccount := range h.store.ListSePayBankAccounts(userID) {
		mapping, ok := h.store.GetBankAccountMapping(bankAccount.ID)
		if !ok || mapping.UserID != userID || mapping.Status != "active" {
			continue
		}
		for _, conn := range h.store.ListBankConnections(mapping.UserID) {
			if conn.Provider != "sepay" || conn.ExternalID != bankAccount.BankAccountXID {
				continue
			}
			for _, feed := range h.store.ListBankFeed(mapping.UserID) {
				if feed.ConnectionID == conn.ID {
					allowed[feed.ID] = *mapping
				}
			}
		}
	}
	return allowed
}

func (h *WealthHandler) ListMyBankFeed(c *gin.Context) {
	userID := domain.ID(currentUser(c))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	state := strings.TrimSpace(c.DefaultQuery("state", "needs_review"))
	stateMap := map[string]domain.TransactionPostingState{"needs_review": domain.PostingStateReview, "confirmed": domain.PostingStatePosted, "ignored": domain.PostingStateIgnored, "ai_tagged": domain.PostingStateReview}
	if _, ok := stateMap[state]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "invalid state"})
		return
	}
	allowed := h.myMappedFeed(userID)
	type feedView struct {
		domain.BankFeedTransaction
		Suggestions []domain.TransactionSuggestion `json:"suggestions"`
		Mapping     domain.BankAccountMapping      `json:"mapping"`
	}
	items := []feedView{}
	for feedID, mapping := range allowed {
		feed, ok := h.store.GetBankFeed(feedID)
		if !ok || feed.PostingState != stateMap[state] {
			continue
		}
		suggestions := h.store.ListTransactionSuggestions(feed.ID)
		if state == "ai_tagged" && len(suggestions) == 0 {
			continue
		}
		items = append(items, feedView{BankFeedTransaction: *feed, Suggestions: suggestions, Mapping: mapping})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].OccurredAt.After(items[j].OccurredAt) })
	c.JSON(http.StatusOK, gin.H{"items": items, "state": state})
}

func (h *WealthHandler) getMyFeed(c *gin.Context) (*domain.BankFeedTransaction, domain.BankAccountMapping, domain.ID, bool) {
	userID := domain.ID(currentUser(c))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return nil, domain.BankAccountMapping{}, "", false
	}
	feed, ok := h.store.GetBankFeed(domain.ID(c.Param("id")))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "bank feed not found"})
		return nil, domain.BankAccountMapping{}, "", false
	}
	mapping, ok := h.myMappedFeed(userID)[feed.ID]
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN"})
		return nil, domain.BankAccountMapping{}, "", false
	}
	if mapping.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "user editor or owner role is required"})
		return nil, domain.BankAccountMapping{}, "", false
	}
	return feed, mapping, userID, true
}

func (h *WealthHandler) ConfirmMyBankFeed(c *gin.Context) { h.postMyBankFeed(c, false) }
func (h *WealthHandler) CorrectMyBankFeed(c *gin.Context) { h.postMyBankFeed(c, true) }

func (h *WealthHandler) postMyBankFeed(c *gin.Context, correct bool) {
	feed, mapping, userID, ok := h.getMyFeed(c)
	if !ok {
		return
	}
	if feed.PostingState != domain.PostingStateReview {
		c.JSON(http.StatusConflict, gin.H{"code": "INVALID_STATE", "message": "bank feed is not awaiting review"})
		return
	}
	request := dto.BankFeedCorrectRequest{}
	if correct {
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
			return
		}
	}
	accountID := mapping.AccountID
	if correct && strings.TrimSpace(request.AccountID) != "" {
		accountID = domain.ID(request.AccountID)
	}
	account, exists := h.store.GetAccount(accountID)
	if !exists || account.UserID != mapping.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ACCOUNT", "message": "account must belong to mapped user"})
		return
	}
	typeValue := domain.TransactionTypeExpense
	if strings.EqualFold(feed.Direction, "in") {
		typeValue = domain.TransactionTypeIncome
	}
	if correct && request.Type != "" {
		typeValue = domain.TransactionType(request.Type)
	}
	if typeValue != domain.TransactionTypeIncome && typeValue != domain.TransactionTypeExpense {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_TYPE", "message": "only income or expense is supported"})
		return
	}
	name := feed.Description
	if correct && strings.TrimSpace(request.Name) != "" {
		name = request.Name
	}
	categoryID := domain.ID("")
	if correct {
		categoryID = domain.ID(request.CategoryID)
	}
	transaction, err := h.service.CreateTransaction(domain.Transaction{UserID: mapping.UserID, AccountID: accountID, PortfolioID: account.PortfolioID, CategoryID: categoryID, Name: name, Type: typeValue, Amount: feed.Amount, Currency: feed.Currency, Note: request.Note, OccurredAt: feed.OccurredAt, Status: domain.TransactionStatusPosted, Source: "sepay_bank_feed"})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.store.LinkBankFeedPosting(feed.ID, transaction.ID)
	action := "confirmed"
	if correct {
		action = "corrected"
		metrics.Inc("sepay_ai_suggestions_corrected_total")
	}
	if request.RememberChoice {
		metrics.Inc("sepay_user_confirmed_rule_total")
	}
	_, _ = h.store.CreateClassificationFeedback(domain.ClassificationFeedback{BankFeedTransactionID: feed.ID, UserID: userID, Action: action, Name: name, CategoryID: categoryID, AccountID: accountID, TransactionType: string(typeValue), Note: request.Note, RememberChoice: request.RememberChoice})
	h.recordAudit(c, "", "bank_feed_transaction", feed.ID, feed, map[string]any{"postedTransactionId": transaction.ID, "action": action}, "success", "")
	c.JSON(http.StatusOK, gin.H{"transaction": transaction, "feedId": feed.ID, "status": "confirmed"})
}

func (h *WealthHandler) IgnoreMyBankFeed(c *gin.Context) {
	feed, _, userID, ok := h.getMyFeed(c)
	if !ok {
		return
	}
	if feed.PostingState != domain.PostingStateReview {
		c.JSON(http.StatusConflict, gin.H{"code": "INVALID_STATE"})
		return
	}
	if err := h.store.UpdateFeedState(feed.ID, domain.PostingStateIgnored, "ignored by user"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}
	_, _ = h.store.CreateClassificationFeedback(domain.ClassificationFeedback{BankFeedTransactionID: feed.ID, UserID: userID, Action: "ignored"})
	metrics.Inc("sepay_bank_feed_ignored_total")
	h.recordAudit(c, "", "bank_feed_transaction", feed.ID, feed, map[string]string{"status": "ignored"}, "success", "")
	c.JSON(http.StatusOK, gin.H{"feedId": feed.ID, "status": "ignored"})
}

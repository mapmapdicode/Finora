package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"wealthos-backend/internal/http/dto"
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

type assistantCommandApprovalRequest struct {
	ApprovalID string `json:"approvalId"`
}

type telegramWebhookPayload struct {
	UpdateID      int64                    `json:"update_id"`
	WorkspaceID   string                   `json:"workspaceId"`
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
	service        *service.WealthService
	store          storage.Store
	secret         string
	telegramSecret string
	hermesSecret   string
	jwtSecret      string
	jwtTTL         time.Duration
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
	return &WealthHandler{
		service:        svcRef,
		store:          store,
		secret:         secret,
		telegramSecret: telegramSecret,
		hermesSecret:   hermesSecret,
		jwtSecret:      jwtSecret,
		jwtTTL:         jwtTTL,
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
	user, err := h.service.RegisterUser(body.Email, body.Password, body.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "REGISTER_FAIL", "message": err.Error()})
		return
	}
	workspace, _ := h.store.CreateWorkspace(defaultWorkspaceName(body.WorkspaceName), "VND", user.ID)
	if workspace == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "WORKSPACE_CREATE_FAIL", "message": "unable to create default workspace"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user":      user,
		"workspace": workspace,
		"token":     h.issueAuthToken(string(user.ID)),
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
		c.JSON(http.StatusUnauthorized, gin.H{"code": "INVALID_CREDENTIALS", "message": err.Error()})
		return
	}
	result.Token = h.issueAuthToken(string(result.User.ID))
	c.JSON(http.StatusOK, result)
}

func (h *WealthHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().UTC()})
}

func (h *WealthHandler) Readyz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready", "ts": time.Now().UTC()})
}

func (h *WealthHandler) ListWorkspaces(c *gin.Context) {
	uid := currentUser(c)
	list := h.store.ListWorkspaces(domain.ID(uid))
	if list == nil {
		list = []domain.Workspace{}
	}
	out := make([]gin.H, 0, len(list))
	for _, item := range list {
		role, hasRole := h.store.GetWorkspaceMemberRole(domain.ID(uid), item.ID)
		record := gin.H{
			"id":           item.ID,
			"name":         item.Name,
			"baseCurrency": item.BaseCurrency,
		}
		if item.FiscalYearEnd != "" {
			record["fiscalYearEnd"] = item.FiscalYearEnd
		}
		if hasRole {
			record["role"] = string(role)
		}
		out = append(out, record)
	}
	c.JSON(http.StatusOK, out)
}

func (h *WealthHandler) ListPortfolios(c *gin.Context) {
	wsID := h.requireWorkspaceID(c)
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
	wsID := h.requireWorkspaceID(c)
	item := domain.Portfolio{
		WorkspaceID:  domain.ID(wsID),
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
	if !h.requireWorkspaceMatch(c, portfolio.WorkspaceID) {
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
	mode := h.getAmountDisplayMode(c)
	resp := dto.PortfolioNetWorthResponse{
		AsOfAt:            nw.AsOfAt,
		BaseCurrency:      nw.BaseCurrency,
		NetWorth:          nw.NetWorth,
		Cash:              nw.Cash,
		Liabilities:       nw.Liabilities,
		AmountDisplayMode: mode,
	}
	c.JSON(http.StatusOK, resp)
}


func (h *WealthHandler) requireEditorRole(c *gin.Context) bool {
	_, ok := h.requireWorkspaceOrReject(c)
	if !ok {
		return false
	}
	role := strings.TrimSpace(func() string {
		if v, ok := c.Get("workspace_role"); ok {
			if s, ok2 := v.(string); ok2 {
				return s
			}
		}
		return ""
	}())
	if role == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing workspace role"})
		return false
	}
	if role == string(domain.RoleViewer) {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "viewer role is read-only"})
		return false
	}
	return true
}

func (h *WealthHandler) requireOwnerRole(c *gin.Context) bool {
	_, ok := h.requireWorkspaceOrReject(c)
	if !ok {
		return false
	}
	role := strings.TrimSpace(func() string {
		if v, ok := c.Get("workspace_role"); ok {
			if s, ok2 := v.(string); ok2 {
				return s
			}
		}
		return ""
	}())
	if role == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing workspace role"})
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
	if !h.requireWorkspaceMatch(c, p.WorkspaceID) {
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
	out := h.service.GetPortfolioSnapshotsForPortfolio(p.WorkspaceID, p.ID, limit, cursor)
	c.JSON(http.StatusOK, out)
}

func (h *WealthHandler) ListAccounts(c *gin.Context) {
	wsID := h.requireWorkspaceID(c)
	list := h.store.ListAccounts(domain.ID(wsID))
	c.JSON(http.StatusOK, list)
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
	account, err := h.store.CreateAccount(domain.Account{
		WorkspaceID: domain.ID(h.requireWorkspaceID(c)),
		PortfolioID: domain.ID(body.PortfolioID),
		Name:        body.Name,
		Type:        body.Type,
		Currency:    body.Currency,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "account", account.ID, nil, account, "success", "")
	c.JSON(http.StatusCreated, account)
}

func (h *WealthHandler) DeleteAccount(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	accountID := domain.ID(c.Param("id"))
	wsID := domain.ID(h.requireWorkspaceID(c))

	account, ok := h.store.GetAccount(accountID)
	targetWsID := wsID
	if ok && account.WorkspaceID != "" {
		targetWsID = account.WorkspaceID
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
	wsID := domain.ID(h.requireWorkspaceID(c))
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
	wsID := domain.ID(h.requireWorkspaceID(c))
	if err := h.store.DeleteLoan(wsID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *WealthHandler) DeleteProperty(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	id := domain.ID(c.Param("id"))
	wsID := domain.ID(h.requireWorkspaceID(c))
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
	wsID := domain.ID(h.requireWorkspaceID(c))
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

	var cursor dto.PaginatedCursor
	if query.Cursor != "" {
		var err error
		cursor, err = dto.DecodeTransactionCursor(query.Cursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_QUERY", "message": err.Error()})
			return
		}
	}

	wsID := h.requireWorkspaceID(c)
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
	happened := body.OccurredAt
	if happened.IsZero() {
		happened = time.Now().UTC()
	}
	item, err := h.service.CreateTransaction(domain.Transaction{
		WorkspaceID: domain.ID(h.requireWorkspaceID(c)),
		AccountID:   domain.ID(body.AccountID),
		CategoryID:  domain.ID(body.CategoryID),
		PortfolioID: domain.ID(body.PortfolioID),
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
	happened := body.OccurredAt
	if happened.IsZero() {
		happened = time.Now().UTC()
	}
	transfer, err := h.service.CreateTransfer(domain.Transfer{
		WorkspaceID:   domain.ID(h.requireWorkspaceID(c)),
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
	wsID := h.requireWorkspaceID(c)
	c.JSON(http.StatusOK, h.store.ListLoans(domain.ID(wsID)))
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
	startAt := body.StartAt
	if startAt.IsZero() {
		startAt = time.Now().UTC()
	}
	dueAt := body.DueAt
	if dueAt.IsZero() {
		dueAt = startAt.AddDate(0, 1, 0)
	}
	lo, err := h.store.CreateLoan(domain.Loan{
		WorkspaceID:      domain.ID(h.requireWorkspaceID(c)),
		PortfolioID:      domain.ID(body.PortfolioID),
		Counterparty:     body.Counterparty,
		Direction:        domain.LoanDirection(body.Direction),
		PrincipalInitial: body.PrincipalInitial,
		PrincipalBalance: body.PrincipalInitial,
		AnnualRate:       body.AnnualRate,
		DayCountBasis:    body.DayCountBasis,
		StartAt:          startAt,
		DueAt:            dueAt,
		InterestCompound: body.InterestCompounding,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
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
	if !h.requireWorkspaceMatch(c, loan.WorkspaceID) {
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
	if !h.requireWorkspaceMatch(c, loan.WorkspaceID) {
		return
	}
	payment, err := h.service.CreateLoanPayment(loanID, domain.LoanPayment{
		WorkspaceID: loan.WorkspaceID,
		Principal:   body.Principal,
		Interest:    body.Interest,
		Fee:         body.Fee,
		Waived:      body.Waived,
		OccurredAt:  nowOrUTC(body.OccurredAt),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "loan_payment", payment.ID, nil, payment, "success", "")
	c.JSON(http.StatusCreated, payment)
}

func (h *WealthHandler) ListProperties(c *gin.Context) {
	wsID := h.requireWorkspaceID(c)
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
		WorkspaceID: domain.ID(h.requireWorkspaceID(c)),
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
	if !h.requireWorkspaceMatch(c, prop.WorkspaceID) {
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
		EffectiveAt: nowOrUTC(body.EffectiveAt),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "property_valuation", v.ID, nil, v, "success", "")
	c.JSON(http.StatusCreated, v)
}

func (h *WealthHandler) ListAssets(c *gin.Context) {
	wsID := h.requireWorkspaceID(c)
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
		WorkspaceID: domain.ID(h.requireWorkspaceID(c)),
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
	if !h.requireWorkspaceMatch(c, asset.WorkspaceID) {
		return
	}
	v, err := h.store.AddAssetValuation(domain.AssetValuation{
		AssetID:     domain.ID(assetID),
		Amount:      body.ValuationAmount,
		Currency:    body.Currency,
		Source:      body.Source,
		EffectiveAt: nowOrUTC(body.EffectiveAt),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "asset_valuation", v.ID, nil, v, "success", "")
	c.JSON(http.StatusCreated, v)
}

func (h *WealthHandler) GetBudget(c *gin.Context) {
	wsID := h.requireWorkspaceID(c)
	if wsID == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing workspace"})
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
	wsID := h.requireWorkspaceID(c)
	if wsID == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing workspace"})
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
	wsID := h.requireWorkspaceID(c)
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
		WorkspaceID: domain.ID(h.requireWorkspaceID(c)),
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
	wsID, ok := h.requireWorkspaceOrReject(c)
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
		WorkspaceID:   domain.ID(h.requireWorkspaceID(c)),
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
	c.JSON(http.StatusCreated, gin.H{
		"connectionId":  conn.ID,
		"provider":      conn.Provider,
		"scope":         conn.Scope,
		"externalId":    conn.ExternalID,
		"callbackState": conn.CallbackState,
		"connectUrl":    connectURL,
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
	wsID := h.requireWorkspaceID(c)
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
	if !h.requireWorkspaceMatch(c, conn.WorkspaceID) {
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
	if !h.requireWorkspaceMatch(c, conn.WorkspaceID) {
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
	expectedHex := make([]byte, hex.DecodedLen(len(expected)))
	_ = hex.Encode(expectedHex, expected)

	var signatureRaw string
	for _, p := range strings.Split(signatureHeader, ",") {
		part := strings.TrimSpace(p)
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
		ConnectionID string `json:"connectionId"`
		AccountID    string `json:"accountId"`
		Direction    string `json:"direction"`
		TransferType string `json:"transferType"`
		Amount       any    `json:"amount"`
		Currency     string `json:"currency"`
		Counterparty string `json:"counterparty"`
		Description  string `json:"description"`
		Reference    string `json:"reference"`
		Content      string `json:"content"`
		ExternalID   string `json:"externalTransactionId"`
		OccurredAt   string `json:"occurredAt"`
		EventID      string `json:"eventId"`
		ID           string `json:"id"`
		Code         string `json:"code"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && (payload.Direction != "" || payload.TransferType != "" || payload.ConnectionID != "") {
		ev := service.SePayWebhookEvent{
			ConnectionID: payload.ConnectionID,
			AccountID:    payload.AccountID,
			Direction:    firstNonEmpty(payload.Direction, payload.TransferType),
			Currency:     firstNonEmpty(payload.Currency, "VND"),
			Counterparty: payload.Counterparty,
			Description:  firstNonEmpty(payload.Description, payload.Content),
			Reference:    payload.Reference,
			Content:      payload.Content,
			ExternalID:   firstNonEmpty(payload.ExternalID, payload.EventID, payload.ID, payload.Code),
			OccurredAt:   payload.OccurredAt,
		}

		amount := formatAmountLikeString(payload.Amount)
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
	wsID := h.requireWorkspaceID(c)
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
	if !h.requireWorkspaceMatch(c, feed.WorkspaceID) {
		return
	}
	posted, err := h.service.ApproveBankFeed(domain.ID(id), *feed)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "bank_feed_transaction", domain.ID(id), feed, map[string]any{
		"status":             string("posted"),
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
	if !h.requireWorkspaceMatch(c, feed.WorkspaceID) {
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
	if !h.requireWorkspaceMatch(c, feed.WorkspaceID) {
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
	wsID := h.requireWorkspaceID(c)
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
		WorkspaceID:      domain.ID(h.requireWorkspaceID(c)),
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
	if !h.requireWorkspaceMatch(c, existing.WorkspaceID) {
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
	wsID := h.requireWorkspaceID(c)
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
			WorkspaceID:  domain.ID(wsID),
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
	if !h.requireWorkspaceMatch(c, loan.WorkspaceID) {
		return
	}
	var body loanPaymentRequestPayload
	_ = c.ShouldBindJSON(&body)
	created, err := h.service.CreateLoanPaymentRequest(loan.WorkspaceID, domain.ID(loanID), service.PaymentRequestCreate{
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

	action, commandID, approvalID, text, userID, workspaceID := h.extractTelegramCommandAndIdentity(c, payload)
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
		if workspaceID == "" || strings.TrimSpace(string(cmd.WorkspaceID)) != workspaceID {
			c.JSON(http.StatusForbidden, gin.H{"code": "WORKSPACE_MISMATCH", "message": "command does not belong to workspace"})
			return
		}
		if userID != "" && string(cmd.UserID) != "" && userID != string(cmd.UserID) {
			c.JSON(http.StatusForbidden, gin.H{"code": "WORKSPACE_MISMATCH", "message": "callback actor does not match command owner"})
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

	if workspaceID == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "WORKSPACE_LINK_REQUIRED", "message": "telegram chat is not linked to a workspace"})
		return
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		uid = "telegram-unbound"
	}
	intent := h.classifyAssistantIntent(text, "")
	status := h.initialStatusForAssistantIntent(intent)

	command, err := h.store.CreateAssistantCommand(domain.AssistantCommand{
		WorkspaceID: domain.ID(workspaceID),
		UserID:      domain.ID(uid),
		Command:     text,
		Plan:        intent,
		Status:      status,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}

	out := gin.H{
		"status":      "received",
		"commandId":   command.ID,
		"workspaceId": workspaceID,
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
	uid := domain.ID(currentUser(c))
	wsID := domain.ID(h.requireWorkspaceID(c))
	intent := h.classifyAssistantIntent(body.Command, body.Plan)
	status := h.initialStatusForAssistantIntent(intent)
	command, err := h.store.CreateAssistantCommand(domain.AssistantCommand{
		WorkspaceID: wsID,
		UserID:      uid,
		Command:     body.Command,
		Plan:        intent,
		Status:      status,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": err.Error()})
		return
	}
	h.recordAudit(c, "", "assistant_command", command.ID, nil, command, "success", "")
	c.JSON(http.StatusCreated, command)
}

func (h *WealthHandler) ListAssistantCommands(c *gin.Context) {
	wsID := h.requireWorkspaceID(c)
	commands := h.store.ListAssistantCommands(domain.ID(wsID))
	c.JSON(http.StatusOK, commands)
}

func (h *WealthHandler) ListAuditLogs(c *gin.Context) {
	wsID := h.requireWorkspaceID(c)
	if wsID == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing workspace"})
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
	if !h.requireWorkspaceMatch(c, cmd.WorkspaceID) {
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
		if !h.requireWorkspaceMatch(c, cmd.WorkspaceID) {
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
		if !h.requireWorkspaceMatch(c, cmd.WorkspaceID) {
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
	workspaceID := strings.TrimSpace(c.Query("workspaceId"))
	if workspaceID == "" {
		workspaceID = strings.TrimSpace(c.GetHeader("x-workspace-id"))
	}
	if workspaceID == "" {
		workspaceID = strings.TrimSpace(payload.WorkspaceID)
	}

	var (
		action    string
		cmdID     string
		apprID    string
		text      string
		userID    string
		sourceMsg *telegramWebhookMessage
	)

	switch {
	case payload.CallbackQuery != nil:
		payloadText := strings.TrimSpace(payload.CallbackQuery.Data)
		action, cmdID, apprID = parseTelegramCallbackApproval(payloadText)
		if action == "" {
			text = payloadText
		}
		if payload.CallbackQuery.From != nil {
			userID = strconv.FormatInt(payload.CallbackQuery.From.ID, 10)
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
			return action, cmdID, apprID, "", userID, workspaceID
		}
		return "", "", "", "", userID, workspaceID
	}

	if userID == "" && sourceMsg != nil && sourceMsg.From != nil {
		userID = strconv.FormatInt(sourceMsg.From.ID, 10)
	}
	return action, cmdID, apprID, text, userID, workspaceID
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

func (h *WealthHandler) requireWorkspaceID(c *gin.Context) string {
	if v, ok := c.Get("workspace_id"); ok {
		if ws, ok2 := v.(string); ok2 && strings.TrimSpace(ws) != "" {
			return strings.TrimSpace(ws)
		}
	}

	headerWs := strings.TrimSpace(c.GetHeader("x-workspace-id"))
	if headerWs != "" {
		return headerWs
	}
	workspaceID := c.Query("workspaceId")
	if workspaceID != "" {
		return workspaceID
	}
	uid := currentUser(c)
	if uid == "" {
		return ""
	}
	workspaces := h.store.ListWorkspaces(domain.ID(uid))
	if len(workspaces) > 0 {
		return string(workspaces[0].ID)
	}
	return ""
}

func (h *WealthHandler) requireWorkspaceOrReject(c *gin.Context) (string, bool) {
	ws := h.requireWorkspaceID(c)
	if ws == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "missing workspace context"})
		return "", false
	}
	return ws, true
}

func (h *WealthHandler) requireWorkspaceMatch(c *gin.Context, resourceWorkspaceID domain.ID) bool {
	wsID, ok := h.requireWorkspaceOrReject(c)
	if !ok {
		return false
	}
	if strings.TrimSpace(string(resourceWorkspaceID)) != wsID {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "workspace mismatch"})
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

func defaultWorkspaceName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "Default Workspace"
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

	wsID := h.resolveAuditWorkspaceID(c, targetID)
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
		WorkspaceID:    wsID,
		ActorID:        domain.ID(actorID),
		ActorRole:      h.workspaceRole(c),
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
	_, _ = h.store.CreateAuditLog(entry)
}

func (h *WealthHandler) resolveAuditWorkspaceID(c *gin.Context, targetID domain.ID) domain.ID {
	if explicit, ok := c.Get("workspace_id"); ok {
		if raw, ok := explicit.(string); ok && strings.TrimSpace(raw) != "" {
			return domain.ID(strings.TrimSpace(raw))
		}
	}
	if explicit := strings.TrimSpace(c.GetHeader("x-workspace-id")); explicit != "" {
		return domain.ID(explicit)
	}
	if explicit := strings.TrimSpace(c.Query("workspaceId")); explicit != "" {
		return domain.ID(explicit)
	}
	if targetID != "" {
		// Keep best effort resilience if handlers log object IDs that are guaranteed to belong to the workspace.
		// Lookup is best effort and currently unsupported by interface for all target types.
	}
	if uid := strings.TrimSpace(currentUser(c)); uid != "" && h != nil && h.store != nil {
		list := h.store.ListWorkspaces(domain.ID(uid))
		if len(list) > 0 {
			return list[0].ID
		}
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

func (h *WealthHandler) workspaceRole(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if raw, ok := c.Get("workspace_role"); ok {
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


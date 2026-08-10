package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"wealthos-backend/internal/config"
	"wealthos-backend/internal/http/handler"
	"wealthos-backend/internal/http/middleware"
	"wealthos-backend/internal/metrics"
	"wealthos-backend/internal/service"
	"wealthos-backend/internal/storage"
)

type Server interface {
	Listen(string) error
	Close() error
}

type ginWrapper struct {
	engine *gin.Engine
}

func (g *ginWrapper) Listen(addr string) error {
	return g.engine.Run(addr)
}

func (g *ginWrapper) Close() error {
	return nil
}

func NewServer(cfg *config.Config, store storage.Store, svc *service.WealthService) (Server, error) {
	h := handler.NewWealthHandler(store, svc, cfg)

	r := gin.New()
	r.Use(middleware.CORS(cfg.CorsOrigins))
	r.Use(gin.Logger())
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		traceID := middleware.RequestIDFromContext(c)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_SERVER_ERROR",
			"message": "Internal server panic: " + strings.TrimSpace(fmt.Sprintf("%v", recovered)),
			"traceId": traceID,
		})
	}))
	r.Use(middleware.RequestID(cfg.RequestIDHeader))
	r.Use(middleware.ErrorEnvelope())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "route not found"})
	})

	r.GET("/healthz", h.Healthz)
	r.GET("/readyz", h.Readyz)
	r.GET("/metrics", gin.WrapF(metrics.Handler))
	// Standard SePay Webhooks (the dashboard product) use camelCase payloads
	// and the configured HMAC secret. Bank Hub IPN is a distinct API/product.
	r.POST("/hooks/sepay-webhook", h.WebhookSePay)
	// Public write endpoints are deliberately account-scoped.  Their secret is
	// a credential for that account only, rather than a user-wide API key.
	r.POST("/public/v1/accounts/:id/transactions", middleware.IdempotencyGuard(store), h.BotCreateTransaction)
	r.POST("/public/v1/accounts/:id/loans", middleware.IdempotencyGuard(store), h.BotCreateLoan)
	r.GET("/public/v1/accounts/:id/transactions/history", h.BotListTransactions)
	// Simple ingestion routes for a user's own trusted bot. They deliberately
	// use the user UUID in the path so an integration can operate without a JWT
	// or an account-specific key. See docs/development/22-bot-simple-ingest-api.md.
	r.GET("/public/v1/users/:id/context", h.BotGetUserContext)
	r.GET("/public/v1/users/:id/loans/accrual-report", h.BotLoanAccrualReport)
	r.POST("/public/v1/users/:id/transactions", middleware.IdempotencyGuard(store), h.BotCreateUserTransaction)
	r.PATCH("/public/v1/users/:id/transactions/:transactionId", middleware.IdempotencyGuard(store), h.BotUpdateUserTransaction)
	r.POST("/public/v1/users/:id/loan-payments", middleware.IdempotencyGuard(store), h.BotCreateUserLoanPayment)
	// Stable provider-facing path. Keep this outside /api/v1 so a deployed
	// webhook URL remains short and does not change with API versioning.
	r.POST("/hooks/sepay-bankhub-ipn", h.BankHubIPN)

	api := r.Group("/api/v1")
	{
		api.POST("/auth/register", h.Register)
		api.POST("/auth/login", middleware.LoginRateLimit(12), h.Login)
		api.POST("/auth/verify-email", middleware.LoginRateLimit(12), h.VerifyEmail)
		api.POST("/auth/resend-verification-email", middleware.LoginRateLimit(4), h.ResendVerificationEmail)
		api.GET("/integrations/sepay/callback", h.SePayCallback)
		api.POST("/webhooks/sepay", h.WebhookSePay)
		api.POST("/webhooks/sepay/bankhub/ipn", h.BankHubIPN)
		api.POST("/webhooks/sepay/bankhub/events", h.BankHubEvent)
		api.POST("/assistant/telegram/webhook", h.TelegramWebhook)
		api.POST("/assistant/executors/:id/events", h.ExecutorEvents)

		protected := api.Group("")
		protected.Use(middleware.UserContextMiddleware(cfg.StaticToken, cfg.JWTSecret))
		protected.GET("/user/settings", h.GetUserSettings)
		protected.PUT("/user/settings", h.UpdateUserSettings)
		protected.GET("/me/sepay", h.GetMySePay)
		protected.POST("/me/sepay/link-session", h.CreateMySePayLinkSession)
		protected.GET("/me/sepay/bank-accounts", h.ListMySePayBankAccounts)
		protected.POST("/me/sepay/bank-accounts/sync", h.SyncMySePayBankAccounts)
		protected.POST("/me/sepay/bank-accounts/:id/map", h.MapMySePayBankAccount)
		protected.POST("/me/sepay/bank-accounts/:id/unlink", h.UnlinkMySePayBankAccount)
		protected.GET("/me/bank-feed", h.ListMyBankFeed)
		protected.POST("/bank-feed/:id/confirm", h.ConfirmMyBankFeed)
		protected.POST("/bank-feed/:id/correct", h.CorrectMyBankFeed)
		protected.POST("/bank-feed/:id/ignore", h.IgnoreMyBankFeed)

		userRequired := protected.Group("")
		userRequired.GET("/portfolios", h.ListPortfolios)
		userRequired.POST("/portfolios", middleware.IdempotencyGuard(store), h.CreatePortfolio)
		userRequired.DELETE("/portfolios/:id", h.DeletePortfolio)
		userRequired.GET("/portfolios/:id/net-worth", h.GetPortfolioNetWorth)
		userRequired.GET("/portfolios/:id/snapshots", h.ListPortfolioSnapshots)
		userRequired.GET("/net-worth", h.GetCurrentNetWorth)

		userRequired.GET("/accounts", h.ListAccounts)
		userRequired.POST("/accounts", middleware.IdempotencyGuard(store), h.CreateAccount)
		userRequired.POST("/accounts/:id/bot-api-key", h.CreateBotAccountKey)
		userRequired.DELETE("/accounts/:id", h.DeleteAccount)

		userRequired.GET("/transactions", h.ListTransactions)
		userRequired.POST("/transactions", middleware.IdempotencyGuard(store), h.CreateTransaction)
		userRequired.POST("/imports/markdown/preview", h.PreviewMarkdownImport)
		userRequired.POST("/imports/markdown/commit", h.CommitMarkdownImport)

		userRequired.POST("/transfers", middleware.IdempotencyGuard(store), h.CreateTransfer)

		userRequired.GET("/customers", h.ListCustomers)
		userRequired.POST("/customers", middleware.IdempotencyGuard(store), h.CreateCustomer)

		userRequired.GET("/loans", h.ListLoans)
		userRequired.GET("/loans/summary", h.GetLoanPortfolioSummary)
		userRequired.GET("/loans/schedule", h.GetLoanSchedule)
		userRequired.POST("/loans", middleware.IdempotencyGuard(store), h.CreateLoan)
		userRequired.DELETE("/loans/:id", h.DeleteLoan)
		userRequired.GET("/loans/:id/accruals", h.GetLoanAccruals)
		userRequired.GET("/loans/:id/payments", h.ListLoanPayments)
		userRequired.POST("/loans/:id/payments", middleware.IdempotencyGuard(store), h.CreateLoanPayment)

		userRequired.GET("/properties", h.ListProperties)
		userRequired.POST("/properties", middleware.IdempotencyGuard(store), h.CreateProperty)
		userRequired.DELETE("/properties/:id", h.DeleteProperty)
		userRequired.POST("/properties/:id/valuations", middleware.IdempotencyGuard(store), h.AddPropertyValuation)

		userRequired.GET("/assets", h.ListAssets)
		userRequired.POST("/assets", middleware.IdempotencyGuard(store), h.CreateAsset)
		userRequired.DELETE("/assets/:id", h.DeleteAsset)
		userRequired.POST("/assets/:id/valuations", middleware.IdempotencyGuard(store), h.AddAssetValuation)

		userRequired.GET("/budgets/:period", h.GetBudget)
		userRequired.PUT("/budgets/:period", middleware.IdempotencyGuard(store), h.UpsertBudget)
		userRequired.GET("/forecast-scenarios", h.ListForecastScenarios)
		userRequired.POST("/forecast-scenarios", middleware.IdempotencyGuard(store), h.CreateForecastScenario)
		userRequired.POST("/forecast-scenarios/:id/run", middleware.IdempotencyGuard(store), h.RunForecastScenario)

		userRequired.POST("/integrations/sepay/connect", middleware.IdempotencyGuard(store), h.CreateSePayConnection)
		userRequired.GET("/bank-connections", h.GetBankConnections)
		userRequired.POST("/bank-connections/:id/sync", middleware.IdempotencyGuard(store), h.SyncBankConnection)
		userRequired.POST("/bank-connections/:id/revoke", h.RevokeBankConnection)

		userRequired.GET("/bank-feed-transactions", h.ListBankFeedTransactions)
		userRequired.POST("/bank-feed-transactions/:id/approve", middleware.IdempotencyGuard(store), h.ApproveBankFeed)
		userRequired.POST("/bank-feed-transactions/:id/reclassify", middleware.IdempotencyGuard(store), h.ReclassifyBankFeed)
		userRequired.POST("/bank-feed-transactions/:id/ignore", middleware.IdempotencyGuard(store), h.IgnoreBankFeed)

		userRequired.GET("/bank-automation-rules", h.ListAutomationRules)
		userRequired.POST("/bank-automation-rules", middleware.IdempotencyGuard(store), h.CreateAutomationRule)
		userRequired.PATCH("/bank-automation-rules/:id", h.ModifyAutomationRule)
		userRequired.DELETE("/bank-automation-rules/:id", h.ModifyAutomationRule)
		userRequired.POST("/bank-automation-rules/preview", h.PreviewAutomationRule)

		userRequired.POST("/assistant/commands", h.CreateAssistantCommand)
		userRequired.GET("/assistant/commands", h.ListAssistantCommands)
		userRequired.GET("/assistant/commands/:id", h.GetAssistantCommand)
		userRequired.POST("/assistant/commands/:id/approve", h.ApproveCommand)
		userRequired.POST("/assistant/commands/:id/cancel", h.CancelCommand)
		userRequired.POST("/loans/:id/payment-requests", middleware.IdempotencyGuard(store), h.CreatePaymentRequest)
		userRequired.GET("/audit-logs", h.ListAuditLogs)
	}

	_ = api.POST

	return &ginWrapper{engine: r}, nil
}

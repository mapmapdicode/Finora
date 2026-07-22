package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"wealthos-backend/internal/config"
	"wealthos-backend/internal/http/handler"
	"wealthos-backend/internal/http/middleware"
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
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID(cfg.RequestIDHeader))
	r.Use(middleware.ErrorEnvelope())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "route not found"})
	})

	r.GET("/healthz", h.Healthz)
	r.GET("/readyz", h.Readyz)

	api := r.Group("/api/v1")
	{
		api.POST("/auth/register", h.Register)
		api.POST("/auth/login", middleware.LoginRateLimit(12), h.Login)
		api.GET("/integrations/sepay/callback", h.SePayCallback)
		api.POST("/webhooks/sepay", h.WebhookSePay)
		api.POST("/assistant/telegram/webhook", h.TelegramWebhook)
		api.POST("/assistant/executors/:id/events", h.ExecutorEvents)

		protected := api.Group("")
		protected.Use(middleware.UserContextMiddleware(cfg.StaticToken, cfg.JWTSecret))
		protected.GET("/workspaces", h.ListWorkspaces)

		workspaceRequired := protected.Group("")
		workspaceRequired.Use(middleware.WorkspaceMembershipMiddleware(store, false))
		workspaceRequired.GET("/portfolios", h.ListPortfolios)
		workspaceRequired.POST("/portfolios", middleware.IdempotencyGuard(store), h.CreatePortfolio)
		workspaceRequired.GET("/portfolios/:id/net-worth", h.GetPortfolioNetWorth)
		workspaceRequired.GET("/portfolios/:id/snapshots", h.ListPortfolioSnapshots)

		workspaceRequired.GET("/accounts", h.ListAccounts)
		workspaceRequired.POST("/accounts", middleware.IdempotencyGuard(store), h.CreateAccount)

		workspaceRequired.GET("/transactions", h.ListTransactions)
		workspaceRequired.POST("/transactions", middleware.IdempotencyGuard(store), h.CreateTransaction)

		workspaceRequired.POST("/transfers", middleware.IdempotencyGuard(store), h.CreateTransfer)

		workspaceRequired.GET("/loans", h.ListLoans)
		workspaceRequired.POST("/loans", middleware.IdempotencyGuard(store), h.CreateLoan)
		workspaceRequired.GET("/loans/:id/accruals", h.GetLoanAccruals)
		workspaceRequired.POST("/loans/:id/payments", middleware.IdempotencyGuard(store), h.CreateLoanPayment)

		workspaceRequired.GET("/properties", h.ListProperties)
		workspaceRequired.POST("/properties", middleware.IdempotencyGuard(store), h.CreateProperty)
		workspaceRequired.POST("/properties/:id/valuations", middleware.IdempotencyGuard(store), h.AddPropertyValuation)

		workspaceRequired.GET("/assets", h.ListAssets)
		workspaceRequired.POST("/assets", middleware.IdempotencyGuard(store), h.CreateAsset)
		workspaceRequired.POST("/assets/:id/valuations", middleware.IdempotencyGuard(store), h.AddAssetValuation)

		workspaceRequired.GET("/budgets/:period", h.GetBudget)
		workspaceRequired.PUT("/budgets/:period", middleware.IdempotencyGuard(store), h.UpsertBudget)
		workspaceRequired.GET("/forecast-scenarios", h.ListForecastScenarios)
		workspaceRequired.POST("/forecast-scenarios", middleware.IdempotencyGuard(store), h.CreateForecastScenario)
		workspaceRequired.POST("/forecast-scenarios/:id/run", middleware.IdempotencyGuard(store), h.RunForecastScenario)

		workspaceRequired.POST("/integrations/sepay/connect", middleware.IdempotencyGuard(store), h.CreateSePayConnection)
		workspaceRequired.GET("/bank-connections", h.GetBankConnections)
		workspaceRequired.POST("/bank-connections/:id/sync", middleware.IdempotencyGuard(store), h.SyncBankConnection)
		workspaceRequired.POST("/bank-connections/:id/revoke", h.RevokeBankConnection)

		workspaceRequired.GET("/bank-feed-transactions", h.ListBankFeedTransactions)
		workspaceRequired.POST("/bank-feed-transactions/:id/approve", middleware.IdempotencyGuard(store), h.ApproveBankFeed)
		workspaceRequired.POST("/bank-feed-transactions/:id/reclassify", middleware.IdempotencyGuard(store), h.ReclassifyBankFeed)
		workspaceRequired.POST("/bank-feed-transactions/:id/ignore", middleware.IdempotencyGuard(store), h.IgnoreBankFeed)

		workspaceRequired.GET("/bank-automation-rules", h.ListAutomationRules)
		workspaceRequired.POST("/bank-automation-rules", middleware.IdempotencyGuard(store), h.CreateAutomationRule)
		workspaceRequired.PATCH("/bank-automation-rules/:id", h.ModifyAutomationRule)
		workspaceRequired.DELETE("/bank-automation-rules/:id", h.ModifyAutomationRule)
		workspaceRequired.POST("/bank-automation-rules/preview", h.PreviewAutomationRule)

		workspaceRequired.POST("/assistant/commands", h.CreateAssistantCommand)
		workspaceRequired.GET("/assistant/commands", h.ListAssistantCommands)
		workspaceRequired.GET("/assistant/commands/:id", h.GetAssistantCommand)
		workspaceRequired.POST("/assistant/commands/:id/approve", h.ApproveCommand)
		workspaceRequired.POST("/assistant/commands/:id/cancel", h.CancelCommand)
		workspaceRequired.POST("/loans/:id/payment-requests", middleware.IdempotencyGuard(store), h.CreatePaymentRequest)
		workspaceRequired.GET("/audit-logs", h.ListAuditLogs)
	}

	_ = api.POST

	return &ginWrapper{engine: r}, nil
}

package app

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"wealthos-backend/internal/cache"
	"wealthos-backend/internal/config"
	"wealthos-backend/internal/db"
	"wealthos-backend/internal/domain"
	httpapi "wealthos-backend/internal/http"
	"wealthos-backend/internal/job"
	"wealthos-backend/internal/service"
	"wealthos-backend/internal/storage"
)

func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	redisCache := cache.NewRedisCache(cfg.RedisURL)

	var store storage.Store
	var cleanup func()
	cleanup = func() {}

	if strings.TrimSpace(cfg.DatabaseURL) != "" {
		dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
		if err != nil {
			return err
		}
		pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
		if err != nil {
			return err
		}

		migrator := db.NewMigrationRunner(pool, "./internal/db/migrations")
		if err := migrator.Run(ctx); err != nil {
			pool.Close()
			return err
		}

		store = storage.NewPostgresStore(pool)
		cleanup = pool.Close
	} else {
		store = storage.NewInMemoryStore()
	}

	seedUserID := store.SeedDemoUser("thanhoangz", "Than Hoang Z", "HoangThanZ6^")
	_ = store.SeedDemoUser("demo@wealthos.vn", "Demo User", "demo-pass")
	existingWorkspaces := store.ListWorkspaces(seedUserID)
	var ws *domain.Workspace
	if len(existingWorkspaces) > 0 {
		ws = &existingWorkspaces[0]
	} else {
		created, err := store.CreateWorkspace("Workspace Demo", "VND", seedUserID)
		if err != nil {
			return err
		}
		ws = created
	}

	appService := service.NewWealthService(store, redisCache)

	if ws != nil {
		appService.SeedDemoData(seedUserID, domain.ID(ws.ID))
	}

	// Auto-seed thanhoangz account with 180M Bank account
	seedUserThanHoangZ(store, appService)

	srv, err := httpapi.NewServer(cfg, store, appService)
	if err != nil {
		return err
	}

	j := job.NewWorker(cfg, store, appService)
	j.Start(ctx)

	addr := ":" + cfg.Port
	go func() {
		<-ctx.Done()
		_ = srv.Close()
		cleanup()
	}()

	log.Printf("server started on %s", addr)
	return srv.Listen(addr)
}

func seedUserThanHoangZ(store storage.Store, appService *service.WealthService) {
	username := "thanhoangz"
	password := "HoangThanZ6^"

	u, ok := store.GetUserByEmail(username)
	var userID domain.ID
	if !ok || u == nil {
		userID = store.SeedDemoUser(username, username, password)
	} else {
		userID = u.ID
	}

	if userID == "" {
		return
	}

	existingWorkspaces := store.ListWorkspaces(userID)
	var ws *domain.Workspace
	if len(existingWorkspaces) > 0 {
		ws = &existingWorkspaces[0]
	} else {
		created, err := store.CreateWorkspace(username, "VND", userID)
		if err != nil {
			return
		}
		ws = created
	}

	p, ok := store.FirstPortfolio(ws.ID)
	if !ok {
		return
	}

	accounts := store.ListAccounts(ws.ID)
	var bankAccID domain.ID
	for _, acc := range accounts {
		if strings.EqualFold(acc.Name, "Bank") {
			bankAccID = acc.ID
			break
		}
	}

	if bankAccID == "" {
		acc, err := store.CreateAccount(domain.Account{
			WorkspaceID: ws.ID,
			PortfolioID: p.ID,
			Name:        "Bank",
			Type:        "bank",
			Currency:    "VND",
		})
		if err == nil {
			bankAccID = acc.ID
		}
	}

	if bankAccID == "" {
		return
	}

	txs := store.ListTransactions(ws.ID, bankAccID)
	hasInitialTx := false
	for _, tx := range txs {
		if tx.Amount == "180000000" || tx.Amount == "180000000.0000" {
			hasInitialTx = true
			break
		}
	}

	if !hasInitialTx {
		_, _ = appService.CreateTransaction(domain.Transaction{
			WorkspaceID: ws.ID,
			AccountID:   bankAccID,
			PortfolioID: p.ID,
			Type:        domain.TransactionTypeIncome,
			Amount:      "180000000",
			Currency:    "VND",
			Note:        "Khoản nạp số dư ban đầu cho tài khoản ngân hàng Bank",
			OccurredAt:  time.Now().UTC(),
			Status:      domain.TransactionStatusPosted,
			Source:      "initial_seed",
		})
	}
}

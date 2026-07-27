package app

import (
	"context"
	"log"
	"strings"

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

	seedUserID := store.SeedDemoUser("demo@wealthos.vn", "Demo User", "demo-pass")
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

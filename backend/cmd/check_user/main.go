package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"wealthos-backend/internal/app"
	"wealthos-backend/internal/service"
	"wealthos-backend/internal/storage"
)

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env.production")

	// 1. Test In-Memory Store
	fmt.Println("=== TESTING IN-MEMORY STORE ===")
	memStore := storage.NewInMemoryStore()
	memSvc := service.NewWealthService(memStore, nil)
	
	// Seed demo users in memory
	uID := memStore.SeedDemoUser("thanhoangz", "thanhoangz", "HoangThanZ6^")
	fmt.Printf("InMemory Seed User ID: %s\n", uID)
	
	uMem, okMem := memStore.GetUserByEmail("thanhoangz")
	if okMem {
		fmt.Printf("GetUserByEmail('thanhoangz') found: ID=%s, Email=%s, Password=%s\n", uMem.ID, uMem.Email, uMem.Password)
	} else {
		fmt.Println("GetUserByEmail('thanhoangz') NOT FOUND in memory!")
	}

	resMem, errMem := memSvc.Authenticate("thanhoangz", "HoangThanZ6^")
	if errMem != nil {
		fmt.Printf("memSvc.Authenticate('thanhoangz', 'HoangThanZ6^') FAILED: %v\n", errMem)
	} else {
		fmt.Printf("memSvc.Authenticate SUCCESS! User ID=%s, Workspace=%v, Token=%s\n", resMem.User.ID, resMem.Workspace, resMem.Token)
	}

	// 2. Test PostgreSQL if reachable
	dbURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(dbURL) != "" {
		fmt.Println("\n=== TESTING POSTGRES STORE ===")
		ctx := context.Background()
		pool, err := pgxpool.New(ctx, dbURL)
		if err == nil && pool.Ping(ctx) == nil {
			pgStore := storage.NewPostgresStore(pool)
			pgSvc := service.NewWealthService(pgStore, nil)

			pgUID := pgStore.SeedDemoUser("thanhoangz", "thanhoangz", "HoangThanZ6^")
			fmt.Printf("Postgres Seed User ID: %s\n", pgUID)

			uPG, okPG := pgStore.GetUserByEmail("thanhoangz")
			if okPG {
				fmt.Printf("GetUserByEmail('thanhoangz') found in Postgres: ID=%s, Email=%s, Password=%s\n", uPG.ID, uPG.Email, uPG.Password)
			} else {
				fmt.Println("GetUserByEmail('thanhoangz') NOT FOUND in Postgres!")
			}

			resPG, errPG := pgSvc.Authenticate("thanhoangz", "HoangThanZ6^")
			if errPG != nil {
				fmt.Printf("pgSvc.Authenticate('thanhoangz', 'HoangThanZ6^') FAILED: %v\n", errPG)
			} else {
				fmt.Printf("pgSvc.Authenticate SUCCESS! User ID=%s\n", resPG.User.ID)
			}
			pool.Close()
		}
	}
	_ = app.Run
}

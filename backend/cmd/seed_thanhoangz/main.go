package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"wealthos-backend/internal/db"
	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/storage"
)

type SeedResult struct {
	User struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	} `json:"user"`
	Portfolio struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"portfolio"`
	Account struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		Currency string `json:"currency"`
	} `json:"account"`
	Category struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Kind string `json:"kind"`
	} `json:"category"`
	Transaction struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Amount     string    `json:"amount"`
		Currency   string    `json:"currency"`
		Note       string    `json:"note"`
		OccurredAt time.Time `json:"occurredAt"`
		Status     string    `json:"status"`
	} `json:"transaction"`
	DatabaseURL string `json:"databaseUrl"`
}

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env.production")

	urlsToTry := []string{}
	if envURL := os.Getenv("DATABASE_URL"); envURL != "" {
		urlsToTry = append(urlsToTry, envURL)
	}
	urlsToTry = append(urlsToTry,
		"postgres://postgres:postgres@localhost:5432/wealthos?sslmode=disable",
		"postgres://postgres:postgres@localhost:5432/wealthos_prod?sslmode=disable",
		"postgres://postgres:123456@localhost:5432/wealthos?sslmode=disable",
		"postgres://postgres:root@localhost:5432/wealthos?sslmode=disable",
		"postgres://postgres:@localhost:5432/wealthos?sslmode=disable",
		"postgres://wealthos_user:W3althOS_P4ss_2026_Secure!@110.172.29.117:5432/wealthos_prod?sslmode=disable",
		"postgres://postgres:postgres@110.172.29.117:5432/wealthos_prod?sslmode=disable",
		"postgres://postgres:postgres@110.172.29.117:5432/wealthos?sslmode=disable",
	)

	ctx := context.Background()
	var pool *pgxpool.Pool
	var dbURL string
	var lastErr error

	for _, u := range urlsToTry {
		log.Printf("Attempting connection to: %s", u)
		dbConfig, errParse := pgxpool.ParseConfig(u)
		if errParse != nil {
			log.Printf("Parse error for %s: %v", u, errParse)
			continue
		}
		dbConfig.ConnConfig.ConnectTimeout = 3 * time.Second
		p, errConn := pgxpool.NewWithConfig(ctx, dbConfig)
		if errConn != nil {
			log.Printf("Conn error for %s: %v", u, errConn)
			continue
		}
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		errPing := p.Ping(pingCtx)
		cancel()
		if errPing == nil {
			pool = p
			dbURL = u
			break
		}
		p.Close()
		log.Printf("Ping failed for %s: %v", u, errPing)
		lastErr = errPing
	}

	if pool == nil {
		log.Fatalf("Could not connect to any database: %v", lastErr)
	}
	defer pool.Close()

	log.Printf("Successfully connected to database at: %s", dbURL)

	// Run migrations
	migrator := db.NewMigrationRunner(pool, "./internal/db/migrations")
	if err := migrator.Run(ctx); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	store := storage.NewPostgresStore(pool)

	username := "thanhoangz"
	password := "HoangThanZ6^"
	amountStr := "180000000" // 180 triệu

	// 1. Check or Create User
	var userID domain.ID
	user, exists := store.GetUserByEmail(username)
	if exists {
		userID = user.ID
		log.Printf("User %s already exists with ID: %s", username, userID)
	} else {
		userID = store.SeedDemoUser(username, username, password)
		if userID == "" {
			log.Fatalf("Failed to seed user %s", username)
		}
		log.Printf("Created user %s with ID: %s", username, userID)
	}

	u, _ := store.GetUser(userID)

	// 2. User Settings
	_, _ = pool.Exec(ctx, `
		INSERT INTO user_settings (user_id, amount_display_mode, updated_at)
		VALUES ($1, 'full', now())
		ON CONFLICT (user_id) DO UPDATE SET amount_display_mode = 'full', updated_at = now()
	`, userID)

	// 3. User
	if _, err := store.EnsureUserPortfolio("", "VND", userID); err != nil {
		log.Fatalf("Failed to create default portfolio: %v", err)
	}

	// 4. Portfolio
	p, ok := store.FirstPortfolio(userID)
	if !ok {
		log.Fatalf("Portfolio not found for user %s", userID)
	}

	// 5. Account: Bank
	accounts := store.ListAccounts(userID)
	var bankAccount *domain.Account
	for _, acc := range accounts {
		if strings.EqualFold(acc.Name, "Bank") {
			bankAccount = &acc
			break
		}
	}

	if bankAccount == nil {
		acc, err := store.CreateAccount(domain.Account{
			UserID:      userID,
			PortfolioID: p.ID,
			Name:        "Bank",
			Type:        "bank",
			Currency:    "VND",
		})
		if err != nil {
			log.Fatalf("Failed to create account Bank: %v", err)
		}
		bankAccount = &acc
	}

	// 6. Category: Initial Balance / Số dư ban đầu
	var catID domain.ID
	errCat := pool.QueryRow(ctx, `
		SELECT id FROM categories WHERE user_id=$1 AND name=$2 AND kind=$3
	`, userID, "Số dư ban đầu", "income").Scan(&catID)
	if errCat != nil || catID == "" {
		_ = pool.QueryRow(ctx, `
			INSERT INTO categories (user_id, name, kind)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, name, kind) DO UPDATE SET updated_at = now()
			RETURNING id
		`, userID, "Số dư ban đầu", "income").Scan(&catID)
	}

	// 7. Transaction: 180,000,000 VND
	txList := store.ListTransactions(userID, bankAccount.ID)
	var initTx *domain.Transaction
	if len(txList) > 0 {
		for _, tx := range txList {
			if tx.Amount == amountStr && tx.Type == domain.TransactionTypeIncome {
				initTx = &tx
				break
			}
		}
	}

	if initTx == nil {
		tx, err := store.CreateTransaction(domain.Transaction{
			UserID:      userID,
			AccountID:   bankAccount.ID,
			CategoryID:  catID,
			PortfolioID: p.ID,
			Type:        domain.TransactionTypeIncome,
			Amount:      amountStr,
			Currency:    "VND",
			Note:        "Số dư ban đầu tài khoản Bank",
			OccurredAt:  time.Now().UTC(),
			Status:      domain.TransactionStatusPosted,
			Source:      "initial_seed",
		})
		if err != nil {
			log.Fatalf("Failed to create initial transaction: %v", err)
		}
		initTx = &tx
	}

	result := SeedResult{
		DatabaseURL: dbURL,
	}
	result.User.ID = string(u.ID)
	result.User.Email = u.Email
	result.User.Name = u.Name
	result.User.Password = password

	result.Portfolio.ID = string(p.ID)
	result.Portfolio.Name = p.Name

	result.Account.ID = string(bankAccount.ID)
	result.Account.Name = bankAccount.Name
	result.Account.Type = bankAccount.Type
	result.Account.Currency = bankAccount.Currency

	result.Category.ID = string(catID)
	result.Category.Name = "Số dư ban đầu"
	result.Category.Kind = "income"

	result.Transaction.ID = string(initTx.ID)
	result.Transaction.Type = string(initTx.Type)
	result.Transaction.Amount = initTx.Amount
	result.Transaction.Currency = initTx.Currency
	result.Transaction.Note = initTx.Note
	result.Transaction.OccurredAt = initTx.OccurredAt
	result.Transaction.Status = string(initTx.Status)

	outJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println("=== SEED COMPLETE ===")
	fmt.Println(string(outJSON))
}

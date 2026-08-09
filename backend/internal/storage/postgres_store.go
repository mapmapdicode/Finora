package storage

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"wealthos-backend/internal/domain"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) SeedDemoUser(email, name string, password string) domain.ID {
	ctx := context.Background()
	var id domain.ID
	_ = s.pool.QueryRow(ctx, `
		SELECT id FROM users WHERE email=$1
	`, email).Scan(&id)
	if id != "" {
		return id
	}

	if email == "" || name == "" || password == "" {
		return ""
	}

	err := s.pool.QueryRow(ctx, `
		INSERT INTO users(email, name, password_hash, email_verified_at)
		VALUES($1, $2, $3, now())
		RETURNING id
	`, email, name, password).Scan(&id)
	if err != nil {
		return ""
	}
	return id
}

func (s *PostgresStore) CreateUser(input domain.User) (domain.User, error) {
	var out domain.User
	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO users(email, name, password_hash)
		VALUES (LOWER(TRIM($1)), $2, $3)
		RETURNING id, email, name, password_hash, email_verified_at, created_at, updated_at
	`, input.Email, input.Name, input.Password).Scan(
		&out.ID, &out.Email, &out.Name, &out.Password, &out.EmailVerifiedAt, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "users_email_key") {
			return domain.User{}, errors.New("email already exists")
		}
		return domain.User{}, err
	}
	return out, nil
}

func (s *PostgresStore) CreateAuditLog(input domain.AuditLog) (domain.AuditLog, error) {
	if strings.TrimSpace(string(input.UserID)) == "" {
		return domain.AuditLog{}, errors.New("userId is required")
	}
	stmt := `
		INSERT INTO audit_logs (
			user_id, actor_id, actor_role, action, target_type, target_id, request_id,
			path, method, policy_decision, result, reason, correlation_id, before_json, after_json
		)
		VALUES (
			$1,
			NULLIF($2, '')::UUID,
			$3, $4, $5,
			NULLIF($6, '')::UUID,
			$7, $8, $9, $10, $11, $12,
			$13, NULLIF($14, '')::jsonb, NULLIF($15, '')::jsonb
		)
		RETURNING id, user_id, actor_id, actor_role, action, target_type, COALESCE(target_id::text, ''),
			request_id, path, method, policy_decision, result, reason, correlation_id,
			COALESCE(before_json::text, ''), COALESCE(after_json::text, ''), created_at, updated_at`
	out := domain.AuditLog{}
	if strings.TrimSpace(input.Action) == "" {
		input.Action = "unknown"
	}
	err := s.pool.QueryRow(context.Background(), stmt,
		input.UserID, input.ActorID, input.ActorRole, input.Action, input.TargetType,
		input.TargetID, input.RequestID, input.Path, input.Method, input.PolicyDecision, input.Result,
		input.Reason, input.CorrelationID, input.BeforeJSON, input.AfterJSON).Scan(
		&out.ID, &out.UserID, &out.ActorID, &out.ActorRole, &out.Action, &out.TargetType,
		&out.TargetID, &out.RequestID, &out.Path, &out.Method, &out.PolicyDecision, &out.Result,
		&out.Reason, &out.CorrelationID, &out.BeforeJSON, &out.AfterJSON, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.AuditLog{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListAuditLogs(userID domain.ID) []domain.AuditLog {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, actor_id, actor_role, action, target_type, COALESCE(target_id::text, ''),
		       request_id, path, method, policy_decision, result, reason, correlation_id,
		       COALESCE(before_json::text, ''), COALESCE(after_json::text, ''), created_at, updated_at
		FROM audit_logs WHERE user_id=$1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return []domain.AuditLog{}
	}
	defer rows.Close()
	out := make([]domain.AuditLog, 0)
	for rows.Next() {
		var item domain.AuditLog
		if err := rows.Scan(&item.ID, &item.UserID, &item.ActorID, &item.ActorRole, &item.Action, &item.TargetType,
			&item.TargetID, &item.RequestID, &item.Path, &item.Method, &item.PolicyDecision, &item.Result,
			&item.Reason, &item.CorrelationID, &item.BeforeJSON, &item.AfterJSON, &item.CreatedAt, &item.UpdatedAt); err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (s *PostgresStore) GetUser(id domain.ID) (*domain.User, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, email, name, password_hash, email_verified_at, created_at, updated_at
		FROM users WHERE id=$1`, id)
	var u domain.User
	var pass string
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &pass, &u.EmailVerifiedAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, false
	}
	u.Password = pass
	return &u, true
}

func (s *PostgresStore) GetUserByID(id domain.ID) (*domain.User, bool) {
	return s.GetUser(id)
}

func (s *PostgresStore) GetUserByEmail(email string) (*domain.User, bool) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	if cleanEmail == "" {
		return nil, false
	}
	prefix := cleanEmail
	if idx := strings.Index(cleanEmail, "@"); idx != -1 {
		prefix = cleanEmail[:idx]
	}

	row := s.pool.QueryRow(context.Background(), `
		SELECT id, email, name, password_hash, email_verified_at, created_at, updated_at
		FROM users
		WHERE LOWER(TRIM(email)) = $1
		   OR LOWER(TRIM(name)) = $1
		   OR LOWER(TRIM(SPLIT_PART(email, '@', 1))) = $2
		LIMIT 1`, cleanEmail, prefix)
	var u domain.User
	var pass string
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &pass, &u.EmailVerifiedAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, false
	}
	u.Password = pass
	return &u, true
}

func (s *PostgresStore) CreateEmailVerificationToken(userID domain.ID, tokenHash string, expiresAt time.Time) error {
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(context.Background(), `DELETE FROM email_verification_tokens WHERE user_id=$1 AND used_at IS NULL`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(context.Background(), `INSERT INTO email_verification_tokens(user_id, token_hash, expires_at) VALUES($1, $2, $3)`, userID, tokenHash, expiresAt); err != nil {
		return err
	}
	return tx.Commit(context.Background())
}

func (s *PostgresStore) VerifyEmail(email, tokenHash string, at time.Time) (*domain.User, error) {
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var user domain.User
	err = tx.QueryRow(context.Background(), `
		UPDATE users u SET email_verified_at=$3, updated_at=$3
		FROM email_verification_tokens t
		WHERE t.user_id=u.id AND LOWER(TRIM(u.email))=LOWER(TRIM($1))
		  AND t.token_hash=$2 AND t.used_at IS NULL AND t.expires_at>$3
		RETURNING u.id, u.email, u.name, u.password_hash, u.email_verified_at, u.created_at, u.updated_at
	`, email, tokenHash, at.UTC()).Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errors.New("verification code is invalid or expired")
	}
	if _, err := tx.Exec(context.Background(), `UPDATE email_verification_tokens SET used_at=$2 WHERE token_hash=$1`, tokenHash, at.UTC()); err != nil {
		return nil, err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *PostgresStore) GetUserSettings(userID domain.ID) (*domain.UserSettings, error) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT user_id, amount_display_mode, updated_at
		FROM user_settings WHERE user_id=$1`, userID)
	var st domain.UserSettings
	var modeStr string
	if err := row.Scan(&st.UserID, &modeStr, &st.UpdatedAt); err != nil {
		// Default if not found
		return &domain.UserSettings{
			UserID:            userID,
			AmountDisplayMode: domain.AmountDisplayModeFull,
			UpdatedAt:         time.Now(),
		}, nil
	}
	st.AmountDisplayMode = domain.AmountDisplayMode(modeStr)
	return &st, nil
}

func (s *PostgresStore) UpsertUserSettings(input domain.UserSettings) (*domain.UserSettings, error) {
	if input.AmountDisplayMode != domain.AmountDisplayModeCompact && input.AmountDisplayMode != domain.AmountDisplayModeFull {
		input.AmountDisplayMode = domain.AmountDisplayModeFull
	}
	now := time.Now()
	row := s.pool.QueryRow(context.Background(), `
		INSERT INTO user_settings (user_id, amount_display_mode, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			amount_display_mode = EXCLUDED.amount_display_mode,
			updated_at = EXCLUDED.updated_at
		RETURNING user_id, amount_display_mode, updated_at
	`, input.UserID, string(input.AmountDisplayMode), now)

	var res domain.UserSettings
	var modeStr string
	if err := row.Scan(&res.UserID, &modeStr, &res.UpdatedAt); err != nil {
		return nil, err
	}
	res.AmountDisplayMode = domain.AmountDisplayMode(modeStr)
	return &res, nil
}

func (s *PostgresStore) EnsureUserPortfolio(_ string, baseCurrency string, userID domain.ID) (*domain.User, error) {
	ctx := context.Background()
	user, ok := s.GetUserByID(userID)
	if !ok {
		return nil, errors.New("user not found")
	}
	baseCurrency = defaultBaseCurrency(baseCurrency)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO portfolios(user_id, name, base_currency)
		SELECT $1, 'Default', $2
		WHERE NOT EXISTS (SELECT 1 FROM portfolios WHERE user_id=$1)
	`, userID, baseCurrency)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *PostgresStore) GetPortfolio(id domain.ID) (*domain.Portfolio, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, name, base_currency, created_at, updated_at
		FROM portfolios WHERE id=$1`, id)
	var p domain.Portfolio
	if err := row.Scan(&p.ID, &p.UserID, &p.Name, &p.BaseCurrency, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, false
	}
	return &p, true
}

func (s *PostgresStore) CreatePortfolio(input domain.Portfolio) (domain.Portfolio, error) {
	ctx := context.Background()
	var out domain.Portfolio
	err := s.pool.QueryRow(ctx, `
		INSERT INTO portfolios(user_id, name, base_currency)
		VALUES($1, $2, $3)
		RETURNING id, user_id, name, base_currency, created_at, updated_at
	`, input.UserID, input.Name, defaultBaseCurrency(input.BaseCurrency)).Scan(&out.ID, &out.UserID, &out.Name, &out.BaseCurrency, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.Portfolio{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListPortfolios(userID domain.ID) []domain.Portfolio {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, name, base_currency, created_at, updated_at
		FROM portfolios
		WHERE user_id=$1
		ORDER BY name ASC`, userID)
	if err != nil {
		return []domain.Portfolio{}
	}
	defer rows.Close()
	out := make([]domain.Portfolio, 0)
	for rows.Next() {
		var p domain.Portfolio
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.BaseCurrency, &p.CreatedAt, &p.UpdatedAt); err == nil {
			out = append(out, p)
		}
	}
	return out
}

func (s *PostgresStore) FirstPortfolio(userID domain.ID) (domain.Portfolio, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, name, base_currency, created_at, updated_at
		FROM portfolios
		WHERE user_id=$1
		ORDER BY created_at ASC
		LIMIT 1
	`, userID)
	var p domain.Portfolio
	if err := row.Scan(&p.ID, &p.UserID, &p.Name, &p.BaseCurrency, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return domain.Portfolio{}, false
	}
	return p, true
}

func (s *PostgresStore) GetAccount(id domain.ID) (*domain.Account, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, portfolio_id, name, type, currency, created_at, updated_at
		FROM accounts WHERE id=$1`, id)
	var a domain.Account
	if err := row.Scan(&a.ID, &a.UserID, &a.PortfolioID, &a.Name, &a.Type, &a.Currency, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, false
	}
	return &a, true
}

func (s *PostgresStore) CreateAccount(input domain.Account) (domain.Account, error) {
	ctx := context.Background()
	var out domain.Account
	err := s.pool.QueryRow(ctx, `
		INSERT INTO accounts(user_id, portfolio_id, name, type, currency)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, user_id, portfolio_id, name, type, currency, created_at, updated_at
	`, input.UserID, nilUUID(input.PortfolioID), input.Name, input.Type, defaultCurrency(input.Currency)).Scan(
		&out.ID, &out.UserID, &out.PortfolioID, &out.Name, &out.Type, &out.Currency, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.Account{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListAccounts(userID domain.ID) []domain.Account {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, portfolio_id, name, type, currency, created_at, updated_at
		FROM accounts WHERE user_id=$1
		ORDER BY name ASC
	`, userID)
	if err != nil {
		return []domain.Account{}
	}
	defer rows.Close()
	out := make([]domain.Account, 0)
	for rows.Next() {
		var a domain.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.PortfolioID, &a.Name, &a.Type, &a.Currency, &a.CreatedAt, &a.UpdatedAt); err == nil {
			out = append(out, a)
		}
	}
	return out
}

func (s *PostgresStore) DeleteAccount(userID domain.ID, id domain.ID) error {
	_, err := s.pool.Exec(context.Background(), `DELETE FROM accounts WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (s *PostgresStore) DeletePortfolio(userID domain.ID, id domain.ID) error {
	_, err := s.pool.Exec(context.Background(), `DELETE FROM portfolios WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (s *PostgresStore) DeleteLoan(userID domain.ID, id domain.ID) error {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, `DELETE FROM loans WHERE id=$1 AND user_id=$2`, id, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `
		DELETE FROM transactions
		WHERE user_id=$1 AND type='loan_disbursement' AND source='loan_disbursement' AND note=$2
	`, userID, "loan principal: "+string(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) DeleteProperty(userID domain.ID, id domain.ID) error {
	_, err := s.pool.Exec(context.Background(), `DELETE FROM properties WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (s *PostgresStore) DeleteAsset(userID domain.ID, id domain.ID) error {
	_, err := s.pool.Exec(context.Background(), `DELETE FROM assets WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (s *PostgresStore) CreateTransactionStrict(input domain.Transaction) (domain.Transaction, error) {
	if input.UserID == "" || input.AccountID == "" || input.Currency == "" {
		return domain.Transaction{}, errors.New("userId, accountId and currency are required")
	}
	if _, err := strconv.ParseFloat(input.Amount, 64); err != nil {
		return domain.Transaction{}, errors.New("invalid amount")
	}
	amount, _ := strconv.ParseFloat(input.Amount, 64)
	if amount <= 0 {
		return domain.Transaction{}, errors.New("amount must be greater than 0")
	}
	if input.AccountID == "" {
		return domain.Transaction{}, errors.New("accountId is required")
	}
	var accountUser domain.ID
	if err := s.pool.QueryRow(context.Background(), `SELECT user_id FROM accounts WHERE id=$1`, input.AccountID).Scan(&accountUser); err != nil {
		return domain.Transaction{}, errors.New("accountId does not exist")
	}
	if accountUser != input.UserID {
		return domain.Transaction{}, errors.New("account does not belong to user")
	}
	if input.Status != "" && !isValidTransactionStatus(input.Status) {
		return domain.Transaction{}, errors.New("invalid transaction status")
	}
	return s.CreateTransaction(input)
}

func (s *PostgresStore) CreateTransaction(input domain.Transaction) (domain.Transaction, error) {
	ctx := context.Background()
	var out domain.Transaction
	err := s.pool.QueryRow(ctx, `
		INSERT INTO transactions(user_id, account_id, category_id, portfolio_id, name, type, amount, currency, note, occurred_at, status, source)
		VALUES($1, $2, $3, $4, $5, $6, CAST($7 as NUMERIC), $8, $9, $10, $11, $12)
		RETURNING id, user_id, account_id, COALESCE(category_id::text, ''), COALESCE(portfolio_id::text, ''),
		          COALESCE(name, ''), type, amount::text, currency, note, occurred_at, status, source, created_at, updated_at
	`, input.UserID, input.AccountID, nilUUID(input.CategoryID), nilUUID(input.PortfolioID), input.Name, input.Type,
		input.Amount, input.Currency, input.Note, input.OccurredAt.UTC(), defaultStatus(input.Status), input.Source).Scan(
		&out.ID, &out.UserID, &out.AccountID, &out.CategoryID, &out.PortfolioID, &out.Name,
		&out.Type, &out.Amount, &out.Currency, &out.Note, &out.OccurredAt, &out.Status, &out.Source, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.Transaction{}, err
	}
	return out, nil
}

func (s *PostgresStore) UpdateTransaction(input domain.Transaction) (domain.Transaction, error) {
	var out domain.Transaction
	err := s.pool.QueryRow(context.Background(), `
		UPDATE transactions
		SET account_id=$3,
			category_id=NULLIF($4, '')::UUID,
			portfolio_id=NULLIF($5, '')::UUID,
			name=$6,
			type=$7,
			amount=CAST($8 AS NUMERIC),
			currency=$9,
			note=$10,
			occurred_at=$11,
			status=$12,
			updated_at=now()
		WHERE id=$1 AND user_id=$2
		RETURNING id, user_id, account_id, COALESCE(category_id::text, ''), COALESCE(portfolio_id::text, ''),
		          COALESCE(name, ''), type, amount::text, currency, note, occurred_at, status, source, created_at, updated_at
	`, input.ID, input.UserID, input.AccountID, nilUUID(input.CategoryID), nilUUID(input.PortfolioID), input.Name,
		input.Type, input.Amount, input.Currency, input.Note, input.OccurredAt.UTC(), input.Status).Scan(
		&out.ID, &out.UserID, &out.AccountID, &out.CategoryID, &out.PortfolioID, &out.Name,
		&out.Type, &out.Amount, &out.Currency, &out.Note, &out.OccurredAt, &out.Status, &out.Source, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Transaction{}, errors.New("transaction not found")
		}
		return domain.Transaction{}, err
	}
	return out, nil
}

func (s *PostgresStore) GetTransaction(id domain.ID) (*domain.Transaction, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, account_id, COALESCE(category_id::text, ''), COALESCE(portfolio_id::text, ''),
		       COALESCE(name, ''), type, amount::text, currency, note, occurred_at, status, source, created_at, updated_at
		FROM transactions WHERE id=$1
	`, id)
	var t domain.Transaction
	if err := row.Scan(&t.ID, &t.UserID, &t.AccountID, &t.CategoryID, &t.PortfolioID, &t.Name,
		&t.Type, &t.Amount, &t.Currency, &t.Note, &t.OccurredAt, &t.Status, &t.Source, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, false
	}
	return &t, true
}

func (s *PostgresStore) UpsertBotAccountKey(input domain.BotAccountKey) (domain.BotAccountKey, error) {
	var out domain.BotAccountKey
	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO bot_account_api_keys(account_id, secret_hash, prefix, revoked_at)
		VALUES($1,$2,$3,NULL)
		ON CONFLICT(account_id) DO UPDATE SET secret_hash=EXCLUDED.secret_hash,prefix=EXCLUDED.prefix,revoked_at=NULL,updated_at=now()
		RETURNING id,account_id,secret_hash,prefix,COALESCE(revoked_at,'epoch'),created_at,updated_at`,
		input.AccountID, input.SecretHash, input.Prefix).Scan(&out.ID, &out.AccountID, &out.SecretHash, &out.Prefix, &out.RevokedAt, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.BotAccountKey{}, err
	}
	if out.RevokedAt.Equal(time.Unix(0, 0).UTC()) {
		out.RevokedAt = time.Time{}
	}
	return out, nil
}

func (s *PostgresStore) GetActiveBotAccountKey(accountID domain.ID) (*domain.BotAccountKey, bool) {
	var out domain.BotAccountKey
	err := s.pool.QueryRow(context.Background(), `SELECT id,account_id,secret_hash,prefix,COALESCE(revoked_at,'epoch'),created_at,updated_at FROM bot_account_api_keys WHERE account_id=$1 AND revoked_at IS NULL`, accountID).Scan(&out.ID, &out.AccountID, &out.SecretHash, &out.Prefix, &out.RevokedAt, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, false
	}
	out.RevokedAt = time.Time{}
	return &out, true
}

func (s *PostgresStore) ListTransactions(userID domain.ID, accountID domain.ID) []domain.Transaction {
	query := `
		SELECT id, user_id, account_id, COALESCE(category_id::text, ''), COALESCE(portfolio_id::text, ''),
		       COALESCE(name, ''), type, amount::text, currency, note, occurred_at, status, source, created_at, updated_at
		FROM transactions
		WHERE user_id=$1`
	args := []any{userID}
	if accountID != "" {
		query += " AND account_id=$2"
		args = append(args, accountID)
	}
	query += " ORDER BY occurred_at DESC, id DESC"
	rows, err := s.pool.Query(context.Background(), query, args...)
	if err != nil {
		return []domain.Transaction{}
	}
	defer rows.Close()
	out := make([]domain.Transaction, 0)
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.AccountID, &t.CategoryID, &t.PortfolioID, &t.Name,
			&t.Type, &t.Amount, &t.Currency, &t.Note, &t.OccurredAt, &t.Status, &t.Source, &t.CreatedAt, &t.UpdatedAt); err == nil {
			out = append(out, t)
		}
	}
	return out
}

func (s *PostgresStore) CreateTransfer(input domain.Transfer) (domain.Transfer, error) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Transfer{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var out domain.Transfer
	err = tx.QueryRow(ctx, `
		INSERT INTO transfers(user_id, from_account_id, to_account_id, amount, currency, note, occurred_at)
		VALUES($1, $2, $3, CAST($4 as NUMERIC), $5, $6, $7)
		RETURNING id, user_id, from_account_id, to_account_id, amount::text, currency, note, occurred_at, created_at, updated_at
	`, input.UserID, input.FromAccountID, input.ToAccountID, input.Amount, defaultCurrency(input.Currency), input.Note, input.OccurredAt.UTC()).Scan(
		&out.ID, &out.UserID, &out.FromAccountID, &out.ToAccountID, &out.Amount, &out.Currency, &out.Note, &out.OccurredAt, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.Transfer{}, err
	}
	for _, entry := range []domain.Transaction{
		{UserID: input.UserID, AccountID: input.FromAccountID, Type: domain.TransactionTypeTransfer, Amount: input.Amount, Currency: defaultCurrency(input.Currency), Note: "internal transfer - out: " + input.Note, OccurredAt: input.OccurredAt, Status: domain.TransactionStatusPosted, Source: "transfer"},
		{UserID: input.UserID, AccountID: input.ToAccountID, Type: domain.TransactionTypeTransfer, Amount: input.Amount, Currency: defaultCurrency(input.Currency), Note: "internal transfer - in: " + input.Note, OccurredAt: input.OccurredAt, Status: domain.TransactionStatusPosted, Source: "transfer"},
	} {
		if _, err := tx.Exec(ctx, `INSERT INTO transactions(user_id, account_id, name, type, amount, currency, note, occurred_at, status, source) VALUES($1,$2,'',$3,CAST($4 AS NUMERIC),$5,$6,$7,$8,$9)`, entry.UserID, entry.AccountID, entry.Type, entry.Amount, entry.Currency, entry.Note, entry.OccurredAt.UTC(), entry.Status, entry.Source); err != nil {
			return domain.Transfer{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Transfer{}, err
	}
	return out, nil
}

func (s *PostgresStore) CreateCustomer(input domain.Customer) (domain.Customer, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.UserID == "" || input.Name == "" {
		return domain.Customer{}, errors.New("customer name is required")
	}
	input.NormalizedName = normalizeCustomerName(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	var out domain.Customer
	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO customers(user_id, name, normalized_name, phone)
		VALUES($1, $2, $3, NULLIF($4, ''))
		ON CONFLICT (user_id, normalized_name) DO UPDATE
		SET phone = COALESCE(NULLIF(EXCLUDED.phone, ''), customers.phone), updated_at = now()
		RETURNING id, user_id, name, normalized_name, COALESCE(phone, ''), created_at, updated_at
	`, input.UserID, input.Name, input.NormalizedName, input.Phone).Scan(
		&out.ID, &out.UserID, &out.Name, &out.NormalizedName, &out.Phone, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.Customer{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListCustomers(userID domain.ID, query string, limit int) []domain.Customer {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	needle := normalizeCustomerName(query)
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, name, normalized_name, COALESCE(phone, ''), created_at, updated_at
		FROM customers
		WHERE user_id=$1 AND ($2='' OR normalized_name LIKE '%' || $2 || '%' OR phone LIKE '%' || $3 || '%')
		ORDER BY updated_at DESC LIMIT $4
	`, userID, needle, strings.TrimSpace(query), limit)
	if err != nil {
		return []domain.Customer{}
	}
	defer rows.Close()
	out := make([]domain.Customer, 0)
	for rows.Next() {
		var customer domain.Customer
		if rows.Scan(&customer.ID, &customer.UserID, &customer.Name, &customer.NormalizedName, &customer.Phone, &customer.CreatedAt, &customer.UpdatedAt) == nil {
			out = append(out, customer)
		}
	}
	return out
}

func (s *PostgresStore) GetCustomer(id domain.ID) (*domain.Customer, bool) {
	var customer domain.Customer
	err := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, name, normalized_name, COALESCE(phone, ''), created_at, updated_at
		FROM customers WHERE id=$1
	`, id).Scan(&customer.ID, &customer.UserID, &customer.Name, &customer.NormalizedName, &customer.Phone, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		return nil, false
	}
	return &customer, true
}

func (s *PostgresStore) GetLoan(id domain.ID) (*domain.Loan, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, COALESCE(portfolio_id::text, ''), COALESCE(customer_id::text, ''), counterparty, direction, principal_initial::text, principal_balance::text, annual_rate::text,
		       COALESCE(day_count_basis, ''), start_at, due_at, status, interest_compounding, daily_rate_per_million::text, COALESCE(settlement_account_id::text, ''), created_at, updated_at
		FROM loans WHERE id=$1`, id)
	var l domain.Loan
	if err := row.Scan(&l.ID, &l.UserID, &l.PortfolioID, &l.CustomerID, &l.Counterparty, &l.Direction,
		&l.PrincipalInitial, &l.PrincipalBalance, &l.AnnualRate, &l.DayCountBasis, &l.StartAt, &l.DueAt,
		&l.Status, &l.InterestCompound, &l.DailyRatePerMillion, &l.SettlementAccountID, &l.CreatedAt, &l.UpdatedAt); err != nil {
		return nil, false
	}
	return &l, true
}

func (s *PostgresStore) UpdateLoan(id domain.ID, mutate func(*domain.Loan)) bool {
	current, ok := s.GetLoan(id)
	if !ok {
		return false
	}
	mutated := *current
	mutate(&mutated)
	ctx := context.Background()
	_, err := s.pool.Exec(ctx, `
		UPDATE loans
		SET principal_balance=CAST($2 AS NUMERIC),
		    principal_initial=CAST($3 AS NUMERIC),
		    annual_rate=CAST($4 AS NUMERIC),
		    day_count_basis=$5,
		    direction=$6,
		    customer_id=NULLIF($7, '')::UUID,
		    counterparty=$8,
		    status=$9,
		    interest_compounding=$10,
		    start_at=$11,
		    due_at=$12,
		    daily_rate_per_million=CAST($13 AS NUMERIC),
		    settlement_account_id=NULLIF($14, '')::UUID,
		    updated_at=now()
		WHERE id=$1
	`, id, mutated.PrincipalBalance, mutated.PrincipalInitial, mutated.AnnualRate, mutated.DayCountBasis, mutated.Direction,
		mutated.CustomerID, mutated.Counterparty, mutated.Status, mutated.InterestCompound, mutated.StartAt, mutated.DueAt, mutated.DailyRatePerMillion, mutated.SettlementAccountID)
	return err == nil
}

func (s *PostgresStore) CreateLoan(input domain.Loan) (domain.Loan, error) {
	ctx := context.Background()
	if strings.TrimSpace(input.DailyRatePerMillion) == "" {
		input.DailyRatePerMillion = "0"
	}
	var out domain.Loan
	err := s.pool.QueryRow(ctx, `
		INSERT INTO loans(user_id, portfolio_id, customer_id, counterparty, direction, principal_initial, principal_balance, annual_rate, day_count_basis, start_at, due_at, status, interest_compounding, daily_rate_per_million, settlement_account_id)
		VALUES($1, $2, NULLIF($3, '')::UUID, $4, $5, CAST($6 AS NUMERIC), CAST($7 AS NUMERIC), CAST($8 AS NUMERIC), $9, $10, $11, $12, $13, CAST($14 AS NUMERIC), NULLIF($15, '')::UUID)
		RETURNING id, user_id, COALESCE(portfolio_id::text, ''), COALESCE(customer_id::text, ''), counterparty, direction, principal_initial::text, principal_balance::text, annual_rate::text, COALESCE(day_count_basis, ''), start_at, due_at, status, interest_compounding, daily_rate_per_million::text, COALESCE(settlement_account_id::text, ''), created_at, updated_at
	`, input.UserID, nilUUID(input.PortfolioID), input.CustomerID, input.Counterparty, input.Direction, input.PrincipalInitial, input.PrincipalBalance,
		input.AnnualRate, nullString(input.DayCountBasis), input.StartAt, input.DueAt, defaultLoanStatus(input.Status), input.InterestCompound, input.DailyRatePerMillion, input.SettlementAccountID).Scan(
		&out.ID, &out.UserID, &out.PortfolioID, &out.CustomerID, &out.Counterparty, &out.Direction,
		&out.PrincipalInitial, &out.PrincipalBalance, &out.AnnualRate, &out.DayCountBasis,
		&out.StartAt, &out.DueAt, &out.Status, &out.InterestCompound, &out.DailyRatePerMillion, &out.SettlementAccountID, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.Loan{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListLoans(userID domain.ID) []domain.Loan {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, COALESCE(portfolio_id::text, ''), COALESCE(customer_id::text, ''), counterparty, direction, principal_initial::text, principal_balance::text,
		       annual_rate::text, COALESCE(day_count_basis,''), start_at, due_at, status, interest_compounding, daily_rate_per_million::text, COALESCE(settlement_account_id::text, ''), created_at, updated_at
		FROM loans WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return []domain.Loan{}
	}
	defer rows.Close()
	out := make([]domain.Loan, 0)
	for rows.Next() {
		var l domain.Loan
		if err := rows.Scan(&l.ID, &l.UserID, &l.PortfolioID, &l.CustomerID, &l.Counterparty, &l.Direction, &l.PrincipalInitial, &l.PrincipalBalance,
			&l.AnnualRate, &l.DayCountBasis, &l.StartAt, &l.DueAt, &l.Status, &l.InterestCompound, &l.DailyRatePerMillion, &l.SettlementAccountID, &l.CreatedAt, &l.UpdatedAt); err == nil {
			out = append(out, l)
		}
	}
	return out
}

func (s *PostgresStore) UpsertImportReference(input domain.ImportReference) (domain.ImportReference, error) {
	var out domain.ImportReference
	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO markdown_import_references(user_id, external_code, entity_type, entity_id, import_month)
		VALUES($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, entity_type, external_code) DO UPDATE
		SET entity_id=EXCLUDED.entity_id, import_month=EXCLUDED.import_month, updated_at=now()
		RETURNING id, user_id, external_code, entity_type, entity_id, import_month, created_at, updated_at
	`, input.UserID, strings.TrimSpace(input.ExternalCode), strings.TrimSpace(input.EntityType), input.EntityID, input.ImportMonth).Scan(
		&out.ID, &out.UserID, &out.ExternalCode, &out.EntityType, &out.EntityID, &out.ImportMonth, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.ImportReference{}, err
	}
	return out, nil
}

func (s *PostgresStore) GetImportReference(userID domain.ID, entityType, externalCode string) (*domain.ImportReference, bool) {
	var out domain.ImportReference
	err := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, external_code, entity_type, entity_id, import_month, created_at, updated_at
		FROM markdown_import_references
		WHERE user_id=$1 AND entity_type=$2 AND external_code=$3
	`, userID, strings.TrimSpace(entityType), strings.TrimSpace(externalCode)).Scan(
		&out.ID, &out.UserID, &out.ExternalCode, &out.EntityType, &out.EntityID, &out.ImportMonth, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return nil, false
	}
	return &out, true
}

func (s *PostgresStore) CreateLoanPayment(input domain.LoanPayment) (domain.LoanPayment, error) {
	ctx := context.Background()
	var out domain.LoanPayment
	err := s.pool.QueryRow(ctx, `
		INSERT INTO loan_payments(user_id, loan_id, account_id, transaction_id, principal_amount, interest_amount, interest_days, fee_amount, waived_amount, occurred_at)
		VALUES($1, $2, $3, NULLIF($4, ''), CAST($5 AS NUMERIC), CAST($6 AS NUMERIC), $7, CAST($8 AS NUMERIC), CAST($9 AS NUMERIC), $10)
		RETURNING id, user_id, loan_id, account_id, COALESCE(transaction_id::text, ''), principal_amount::text, interest_amount::text, interest_days, fee_amount::text, waived_amount::text, occurred_at, created_at, updated_at
	`, input.UserID, input.LoanID, nilUUID(input.AccountID), input.TransactionID, input.Principal, input.Interest, input.InterestDays, input.Fee, input.Waived, input.OccurredAt.UTC()).Scan(
		&out.ID, &out.UserID, &out.LoanID, &out.AccountID, &out.TransactionID, &out.Principal, &out.Interest, &out.InterestDays, &out.Fee, &out.Waived,
		&out.OccurredAt, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.LoanPayment{}, err
	}
	return out, nil
}

func (s *PostgresStore) SettleLoanPayment(loanID domain.ID, expectedPrincipalBalance, nextPrincipalBalance string, nextStatus domain.LoanStatus, payment domain.LoanPayment, ledger domain.Transaction) (domain.LoanPayment, error) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.LoanPayment{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var transactionID domain.ID
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions(user_id, account_id, portfolio_id, name, type, amount, currency, note, occurred_at, status, source)
		VALUES($1,$2,NULLIF($3,'')::UUID,'',$4,CAST($5 AS NUMERIC),$6,$7,$8,$9,$10)
		RETURNING id`, ledger.UserID, ledger.AccountID, ledger.PortfolioID, ledger.Type, ledger.Amount, ledger.Currency, ledger.Note, ledger.OccurredAt.UTC(), ledger.Status, ledger.Source).Scan(&transactionID)
	if err != nil {
		return domain.LoanPayment{}, err
	}
	result, err := tx.Exec(ctx, `
		UPDATE loans SET principal_balance=CAST($3 AS NUMERIC), status=$4, updated_at=now()
		WHERE id=$1 AND principal_balance=CAST($2 AS NUMERIC)`, loanID, expectedPrincipalBalance, nextPrincipalBalance, nextStatus)
	if err != nil {
		return domain.LoanPayment{}, err
	}
	if result.RowsAffected() != 1 {
		return domain.LoanPayment{}, errors.New("loan balance changed; retry payment")
	}
	var out domain.LoanPayment
	err = tx.QueryRow(ctx, `
		INSERT INTO loan_payments(user_id, loan_id, account_id, transaction_id, principal_amount, interest_amount, interest_days, fee_amount, waived_amount, occurred_at)
		VALUES($1,$2,$3,$4,CAST($5 AS NUMERIC),CAST($6 AS NUMERIC),$7,CAST($8 AS NUMERIC),CAST($9 AS NUMERIC),$10)
		RETURNING id,user_id,loan_id,account_id,transaction_id::text,principal_amount::text,interest_amount::text,interest_days,fee_amount::text,waived_amount::text,occurred_at,created_at,updated_at`,
		payment.UserID, loanID, payment.AccountID, transactionID, payment.Principal, payment.Interest, payment.InterestDays, payment.Fee, payment.Waived, payment.OccurredAt.UTC()).Scan(&out.ID, &out.UserID, &out.LoanID, &out.AccountID, &out.TransactionID, &out.Principal, &out.Interest, &out.InterestDays, &out.Fee, &out.Waived, &out.OccurredAt, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.LoanPayment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.LoanPayment{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListLoanPayments(userID domain.ID, loanID domain.ID) []domain.LoanPayment {
	query := `
		SELECT id, user_id, loan_id, account_id, COALESCE(transaction_id::text, ''), principal_amount::text, interest_amount::text, interest_days, fee_amount::text, waived_amount::text, occurred_at, created_at, updated_at
		FROM loan_payments
		WHERE user_id=$1`
	args := []any{userID}
	if loanID != "" {
		query += " AND loan_id=$2"
		args = append(args, loanID)
	}
	query += " ORDER BY occurred_at DESC"
	rows, err := s.pool.Query(context.Background(), query, args...)
	if err != nil {
		return []domain.LoanPayment{}
	}
	defer rows.Close()
	out := make([]domain.LoanPayment, 0)
	for rows.Next() {
		var p domain.LoanPayment
		if err := rows.Scan(&p.ID, &p.UserID, &p.LoanID, &p.AccountID, &p.TransactionID, &p.Principal, &p.Interest, &p.InterestDays, &p.Fee, &p.Waived, &p.OccurredAt, &p.CreatedAt, &p.UpdatedAt); err == nil {
			out = append(out, p)
		}
	}
	return out
}

func (s *PostgresStore) GetProperty(id domain.ID) (*domain.Property, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, portfolio_id, name, address, area_m2, purchase_at, created_at, updated_at
		FROM properties WHERE id=$1`, id)
	var p domain.Property
	if err := row.Scan(&p.ID, &p.UserID, &p.PortfolioID, &p.Name, &p.Address, &p.AreaM2, &p.PurchaseAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, false
	}
	return &p, true
}

func (s *PostgresStore) CreateProperty(input domain.Property) (domain.Property, error) {
	ctx := context.Background()
	var out domain.Property
	err := s.pool.QueryRow(ctx, `
		INSERT INTO properties(user_id, portfolio_id, name, address, area_m2, purchase_at)
		VALUES($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, portfolio_id, name, address, area_m2, purchase_at, created_at, updated_at
	`, input.UserID, nilUUID(input.PortfolioID), input.Name, input.Address, input.AreaM2, input.PurchaseAt).Scan(
		&out.ID, &out.UserID, &out.PortfolioID, &out.Name, &out.Address, &out.AreaM2, &out.PurchaseAt, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.Property{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListProperties(userID domain.ID) []domain.Property {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, portfolio_id, name, address, area_m2, purchase_at, created_at, updated_at
		FROM properties WHERE user_id=$1 ORDER BY name ASC`, userID)
	if err != nil {
		return []domain.Property{}
	}
	defer rows.Close()
	out := make([]domain.Property, 0)
	for rows.Next() {
		var p domain.Property
		if err := rows.Scan(&p.ID, &p.UserID, &p.PortfolioID, &p.Name, &p.Address, &p.AreaM2, &p.PurchaseAt, &p.CreatedAt, &p.UpdatedAt); err == nil {
			out = append(out, p)
		}
	}
	return out
}

func (s *PostgresStore) AddPropertyValuation(v domain.PropertyValuation) (domain.PropertyValuation, error) {
	ctx := context.Background()
	var out domain.PropertyValuation
	err := s.pool.QueryRow(ctx, `
		INSERT INTO property_valuations(property_id, amount, currency, source, effective_at, is_stale)
		VALUES($1, CAST($2 AS NUMERIC), $3, $4, $5, $6)
		RETURNING id, property_id, amount::text, currency, source, effective_at, is_stale, created_at, updated_at
	`, v.PropertyID, v.Amount, defaultCurrency(v.Currency), v.Source, v.EffectiveAt.UTC(), v.IsStale).Scan(
		&out.ID, &out.PropertyID, &out.Amount, &out.Currency, &out.Source, &out.EffectiveAt, &out.IsStale, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) ListPropertyValues(userID domain.ID) []domain.PropertyValuation {
	rows, err := s.pool.Query(context.Background(), `
		SELECT pv.id, pv.property_id, pv.amount::text, pv.currency, pv.source, pv.effective_at, pv.is_stale, pv.created_at, pv.updated_at
		FROM property_valuations pv
		JOIN properties p ON p.id = pv.property_id
		WHERE p.user_id=$1
		ORDER BY pv.effective_at DESC`, userID)
	if err != nil {
		return []domain.PropertyValuation{}
	}
	defer rows.Close()
	out := make([]domain.PropertyValuation, 0)
	for rows.Next() {
		var v domain.PropertyValuation
		if err := rows.Scan(&v.ID, &v.PropertyID, &v.Amount, &v.Currency, &v.Source, &v.EffectiveAt, &v.IsStale, &v.CreatedAt, &v.UpdatedAt); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func (s *PostgresStore) GetAsset(id domain.ID) (*domain.Asset, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, portfolio_id, name, asset_type, created_at, updated_at
		FROM assets WHERE id=$1`, id)
	var a domain.Asset
	if err := row.Scan(&a.ID, &a.UserID, &a.PortfolioID, &a.Name, &a.Type, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, false
	}
	return &a, true
}

func (s *PostgresStore) CreateAsset(input domain.Asset) (domain.Asset, error) {
	ctx := context.Background()
	var out domain.Asset
	err := s.pool.QueryRow(ctx, `
		INSERT INTO assets(user_id, portfolio_id, name, asset_type)
		VALUES($1, $2, $3, $4)
		RETURNING id, user_id, portfolio_id, name, asset_type, created_at, updated_at
	`, input.UserID, nilUUID(input.PortfolioID), input.Name, input.Type).Scan(&out.ID, &out.UserID, &out.PortfolioID, &out.Name, &out.Type, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func (s *PostgresStore) ListAssets(userID domain.ID) []domain.Asset {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, portfolio_id, name, asset_type, created_at, updated_at
		FROM assets WHERE user_id=$1 ORDER BY name ASC`, userID)
	if err != nil {
		return []domain.Asset{}
	}
	defer rows.Close()
	out := make([]domain.Asset, 0)
	for rows.Next() {
		var a domain.Asset
		if err := rows.Scan(&a.ID, &a.UserID, &a.PortfolioID, &a.Name, &a.Type, &a.CreatedAt, &a.UpdatedAt); err == nil {
			out = append(out, a)
		}
	}
	return out
}

func (s *PostgresStore) AddAssetValuation(v domain.AssetValuation) (domain.AssetValuation, error) {
	ctx := context.Background()
	var out domain.AssetValuation
	err := s.pool.QueryRow(ctx, `
		INSERT INTO asset_valuations(asset_id, amount, currency, source, effective_at)
		VALUES($1, CAST($2 AS NUMERIC), $3, $4, $5)
		RETURNING id, asset_id, amount::text, currency, source, effective_at, created_at, updated_at
	`, v.AssetID, v.Amount, defaultCurrency(v.Currency), v.Source, v.EffectiveAt.UTC()).Scan(
		&out.ID, &out.AssetID, &out.Amount, &out.Currency, &out.Source, &out.EffectiveAt, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) ListAssetValues(userID domain.ID) []domain.AssetValuation {
	rows, err := s.pool.Query(context.Background(), `
		SELECT av.id, av.asset_id, av.amount::text, av.currency, av.source, av.effective_at, av.created_at, av.updated_at
		FROM asset_valuations av
		JOIN assets a ON a.id = av.asset_id
		WHERE a.user_id=$1
		ORDER BY av.effective_at DESC`, userID)
	if err != nil {
		return []domain.AssetValuation{}
	}
	defer rows.Close()
	out := make([]domain.AssetValuation, 0)
	for rows.Next() {
		var v domain.AssetValuation
		if err := rows.Scan(&v.ID, &v.AssetID, &v.Amount, &v.Currency, &v.Source, &v.EffectiveAt, &v.CreatedAt, &v.UpdatedAt); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func (s *PostgresStore) CreateBudget(input domain.Budget) (domain.Budget, error) {
	ctx := context.Background()
	var out domain.Budget
	err := s.pool.QueryRow(ctx, `
		INSERT INTO budgets(user_id, period, category_id, limit_amount)
		VALUES($1, $2, NULLIF($3, ''), CAST($4 AS NUMERIC))
		RETURNING id, user_id, period, COALESCE(category_id::text, ''), limit_amount::text, created_at, updated_at
	`, input.UserID, input.Period, input.CategoryID, input.Limit).Scan(&out.ID, &out.UserID, &out.Period, &out.CategoryID, &out.Limit, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.Budget{}, err
	}
	return out, nil
}

func (s *PostgresStore) UpsertBudget(input domain.Budget) (domain.Budget, error) {
	ctx := context.Background()
	var out domain.Budget
	category := strings.TrimSpace(string(input.CategoryID))
	_, hasCategory := func(v string) (domain.ID, bool) {
		if strings.TrimSpace(v) == "" {
			return "", false
		}
		return domain.ID(v), true
	}(category)
	if hasCategory {
		err := s.pool.QueryRow(ctx, `
			INSERT INTO budgets(user_id, period, category_id, limit_amount)
			VALUES($1, $2, $3, CAST($4 AS NUMERIC))
			ON CONFLICT (user_id, period, category_id) DO UPDATE
			SET limit_amount = CAST($4 AS NUMERIC), updated_at = now()
			RETURNING id, user_id, period, COALESCE(category_id::text, ''), limit_amount::text, created_at, updated_at
		`, input.UserID, input.Period, input.CategoryID, input.Limit).Scan(
			&out.ID, &out.UserID, &out.Period, &out.CategoryID, &out.Limit, &out.CreatedAt, &out.UpdatedAt,
		)
		if err == nil {
			return out, nil
		}
	}

	// category_id can be NULL; unique (user_id, period, category_id) treats NULL as distinct row,
	// so manual upsert is safer for null-category budgets.
	var existing domain.Budget
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, period, COALESCE(category_id::text, ''), limit_amount::text, created_at, updated_at
		FROM budgets
		WHERE user_id=$1 AND period=$2 AND category_id IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, input.UserID, input.Period).Scan(&existing.ID, &existing.UserID, &existing.Period, &existing.CategoryID, &existing.Limit, &existing.CreatedAt, &existing.UpdatedAt)
	if err == nil {
		_, _ = s.pool.Exec(ctx, `
			UPDATE budgets
			SET limit_amount = CAST($1 AS NUMERIC), updated_at = now()
			WHERE id=$2
		`, input.Limit, existing.ID)
		existing.Limit = input.Limit
		existing.UpdatedAt = nowUTC()
		return existing, nil
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO budgets(user_id, period, category_id, limit_amount)
		VALUES($1, $2, NULLIF($3, ''), CAST($4 AS NUMERIC))
		RETURNING id, user_id, period, COALESCE(category_id::text, ''), limit_amount::text, created_at, updated_at
	`, input.UserID, input.Period, input.CategoryID, input.Limit).Scan(&out.ID, &out.UserID, &out.Period, &out.CategoryID, &out.Limit, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.Budget{}, err
	}
	return out, nil
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

func (s *PostgresStore) ListBudgets(userID domain.ID, period string) []domain.Budget {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, period, COALESCE(category_id::text, ''), limit_amount::text, created_at, updated_at
		FROM budgets
		WHERE user_id=$1 AND period=$2
		ORDER BY COALESCE(category_id::text, '')
	`, userID, period)
	if err != nil {
		return []domain.Budget{}
	}
	defer rows.Close()
	out := make([]domain.Budget, 0)
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.ID, &b.UserID, &b.Period, &b.CategoryID, &b.Limit, &b.CreatedAt, &b.UpdatedAt); err == nil {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CategoryID < out[j].CategoryID
	})
	return out
}

func (s *PostgresStore) UpsertBudgetAllocs(input domain.BudgetAllocation) (domain.BudgetAllocation, error) {
	// idempotent update by budget id
	var out domain.BudgetAllocation
	ctx := context.Background()
	existing := domain.BudgetAllocation{}
	err := s.pool.QueryRow(ctx, `
		UPDATE budget_allocations
		SET amount_spent = CAST($2 AS NUMERIC), currency = $3, updated_at = now()
		WHERE id = (
			SELECT id FROM budget_allocations
			WHERE budget_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		RETURNING id, budget_id, amount_spent::text, currency, created_at, updated_at
	`, input.BudgetID, input.AmountSpent, defaultCurrency(input.Currency)).Scan(
		&existing.ID, &existing.BudgetID, &existing.AmountSpent, &existing.Currency, &existing.CreatedAt, &existing.UpdatedAt,
	)
	if err == nil && existing.ID != "" {
		return existing, nil
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO budget_allocations(budget_id, amount_spent, currency)
		VALUES($1, CAST($2 AS NUMERIC), $3)
		RETURNING id, budget_id, amount_spent::text, currency, created_at, updated_at
	`, input.BudgetID, input.AmountSpent, defaultCurrency(input.Currency)).Scan(
		&out.ID, &out.BudgetID, &out.AmountSpent, &out.Currency, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) CreateForecastScenario(input domain.ForecastScenario) (domain.ForecastScenario, error) {
	ctx := context.Background()
	var out domain.ForecastScenario
	err := s.pool.QueryRow(ctx, `
		INSERT INTO forecast_scenarios(user_id, name, assumptions, status)
		VALUES($1, $2, $3, 'draft')
		RETURNING id, user_id, name, status, assumptions, result, created_at, updated_at
	`, input.UserID, input.Name, input.Assumptions).Scan(
		&out.ID, &out.UserID, &out.Name, &out.Status, &out.Assumptions, &out.Result, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) ListForecastScenarios(userID domain.ID) []domain.ForecastScenario {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, name, status, assumptions, result, created_at, updated_at
		FROM forecast_scenarios WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return []domain.ForecastScenario{}
	}
	defer rows.Close()
	out := make([]domain.ForecastScenario, 0)
	for rows.Next() {
		var f domain.ForecastScenario
		if err := rows.Scan(&f.ID, &f.UserID, &f.Name, &f.Status, &f.Assumptions, &f.Result, &f.CreatedAt, &f.UpdatedAt); err == nil {
			out = append(out, f)
		}
	}
	return out
}

func (s *PostgresStore) ListForecastScenariosByStatus(status string) []domain.ForecastScenario {
	status = strings.TrimSpace(status)
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, name, status, assumptions, result, created_at, updated_at
		FROM forecast_scenarios WHERE status=$1 ORDER BY updated_at DESC`, status)
	if err != nil {
		return []domain.ForecastScenario{}
	}
	defer rows.Close()
	out := make([]domain.ForecastScenario, 0)
	for rows.Next() {
		var f domain.ForecastScenario
		if err := rows.Scan(&f.ID, &f.UserID, &f.Name, &f.Status, &f.Assumptions, &f.Result, &f.CreatedAt, &f.UpdatedAt); err == nil {
			out = append(out, f)
		}
	}
	return out
}

func (s *PostgresStore) RunForecastScenario(id domain.ID, assumptions string) (domain.ForecastScenario, error) {
	ctx := context.Background()
	var f domain.ForecastScenario
	if assumptions == "" {
		assumptions = "{}"
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE forecast_scenarios
		SET status='running', assumptions=$2, updated_at=now()
		WHERE id=$1`, id, assumptions)
	if err != nil {
		return domain.ForecastScenario{}, err
	}
	err = s.pool.QueryRow(ctx, `
		SELECT id, user_id, name, status, assumptions, result, created_at, updated_at
		FROM forecast_scenarios WHERE id=$1`, id).Scan(&f.ID, &f.UserID, &f.Name, &f.Status, &f.Assumptions, &f.Result, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

func (s *PostgresStore) FinalizeForecastScenario(id domain.ID, status string, result string) (domain.ForecastScenario, error) {
	ctx := context.Background()
	status = strings.TrimSpace(status)
	if status == "" {
		status = "done"
	}
	var f domain.ForecastScenario
	err := s.pool.QueryRow(ctx, `
		UPDATE forecast_scenarios
		SET status=$2, result=$3, updated_at=now()
		WHERE id=$1
		RETURNING id, user_id, name, status, assumptions, result, created_at, updated_at
	`, id, status, strings.TrimSpace(result)).Scan(&f.ID, &f.UserID, &f.Name, &f.Status, &f.Assumptions, &f.Result, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return domain.ForecastScenario{}, err
	}
	return f, nil
}

func (s *PostgresStore) CreateBankConnection(input domain.BankConnection) (domain.BankConnection, error) {
	ctx := context.Background()
	var out domain.BankConnection
	err := s.pool.QueryRow(ctx, `
		INSERT INTO bank_connections(
			user_id, provider, external_id, scope, status, bank_code,
			callback_state, sync_status, last_synced_at, last_sync_requested_at
		)
		VALUES(
			$1, $2, $3, $4,
			'connected',
			NULLIF($5, ''),
			COALESCE(NULLIF($6, ''), 'not_called'),
			COALESCE(NULLIF($7, ''), 'idle'),
			NULL, NULL
		)
		RETURNING id, user_id, provider, external_id, status, scope, COALESCE(bank_code, ''), callback_state, sync_status, COALESCE(last_synced_at, '0001-01-01 00:00:00+00'::timestamptz), COALESCE(last_sync_requested_at, '0001-01-01 00:00:00+00'::timestamptz), created_at, updated_at
	`, input.UserID, input.Provider, input.ExternalID, input.Scope, input.BankCode, input.CallbackState, input.SyncStatus).Scan(
		&out.ID, &out.UserID, &out.Provider, &out.ExternalID, &out.Status, &out.Scope, &out.BankCode,
		&out.CallbackState, &out.SyncStatus, &out.LastSyncedAt, &out.LastSyncRequestedAt, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) ListBankConnections(userID domain.ID) []domain.BankConnection {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, provider, external_id, status, scope, COALESCE(bank_code, ''), callback_state,
		       sync_status, COALESCE(last_synced_at, '0001-01-01 00:00:00+00'::timestamptz), COALESCE(last_sync_requested_at, '0001-01-01 00:00:00+00'::timestamptz), created_at, updated_at
		FROM bank_connections WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return []domain.BankConnection{}
	}
	defer rows.Close()
	out := make([]domain.BankConnection, 0)
	for rows.Next() {
		var c domain.BankConnection
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.Provider, &c.ExternalID, &c.Status, &c.Scope, &c.BankCode, &c.CallbackState, &c.SyncStatus,
			&c.LastSyncedAt, &c.LastSyncRequestedAt, &c.CreatedAt, &c.UpdatedAt,
		); err == nil {
			out = append(out, c)
		}
	}
	return out
}

func (s *PostgresStore) GetBankConnection(id domain.ID) (*domain.BankConnection, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, provider, external_id, status, scope, COALESCE(bank_code, ''), callback_state,
		       sync_status, COALESCE(last_synced_at, '0001-01-01 00:00:00+00'::timestamptz), COALESCE(last_sync_requested_at, '0001-01-01 00:00:00+00'::timestamptz), created_at, updated_at
		FROM bank_connections WHERE id=$1`, id)
	var c domain.BankConnection
	if err := row.Scan(&c.ID, &c.UserID, &c.Provider, &c.ExternalID, &c.Status, &c.Scope, &c.BankCode,
		&c.CallbackState, &c.SyncStatus, &c.LastSyncedAt, &c.LastSyncRequestedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, false
	}
	return &c, true
}

func (s *PostgresStore) GetBankConnectionByCallbackState(callbackState string) (*domain.BankConnection, bool) {
	state := strings.TrimSpace(callbackState)
	if state == "" {
		return nil, false
	}
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, provider, external_id, status, scope, COALESCE(bank_code, ''), callback_state,
		       sync_status, COALESCE(last_synced_at, '0001-01-01 00:00:00+00'::timestamptz), COALESCE(last_sync_requested_at, '0001-01-01 00:00:00+00'::timestamptz), created_at, updated_at
		FROM bank_connections WHERE callback_state=$1`, state)
	var c domain.BankConnection
	if err := row.Scan(&c.ID, &c.UserID, &c.Provider, &c.ExternalID, &c.Status, &c.Scope, &c.BankCode, &c.CallbackState,
		&c.SyncStatus, &c.LastSyncedAt, &c.LastSyncRequestedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, false
	}
	return &c, true
}

func (s *PostgresStore) UpdateBankConnection(id domain.ID, mutate func(*domain.BankConnection)) bool {
	current, ok := s.GetBankConnection(id)
	if !ok {
		return false
	}
	updated := *current
	mutate(&updated)
	_, err := s.pool.Exec(context.Background(), `
		UPDATE bank_connections
		SET provider=$2, external_id=$3, status=$4, scope=$5, bank_code=COALESCE(NULLIF($6, ''), bank_code),
		    callback_state=$7, sync_status=$8, last_synced_at=$9, last_sync_requested_at=$10, updated_at=now()
		WHERE id=$1
	`, updated.ID, updated.Provider, updated.ExternalID, updated.Status, updated.Scope, updated.BankCode,
		updated.CallbackState, updated.SyncStatus, nilTime(updated.LastSyncedAt), nilTime(updated.LastSyncRequestedAt))
	return err == nil
}

func nilTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func (s *PostgresStore) EnqueueBankFeedEvent(input domain.BankFeedEvent) (domain.BankFeedEvent, error) {
	if input.UserID == "" || input.ConnectionID == "" {
		return domain.BankFeedEvent{}, errors.New("userId and connectionId are required")
	}
	if input.Provider == "" {
		return domain.BankFeedEvent{}, errors.New("provider is required")
	}
	if input.State == "" {
		input.State = domain.BankFeedEventStateQueued
	}
	ctx := context.Background()
	var out domain.BankFeedEvent
	err := s.pool.QueryRow(ctx, `
		INSERT INTO bank_feed_events(
			user_id, connection_id, provider, event_key, external_transaction_id, state, payload, attempts, last_error
		)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT(event_key) DO UPDATE
			SET updated_at=EXCLUDED.updated_at
		RETURNING id, user_id, connection_id, provider, event_key, external_transaction_id, state, payload::text,
			attempts, COALESCE(last_error, ''), created_at, updated_at
	`, input.UserID, input.ConnectionID, input.Provider, input.EventKey, input.ExternalID, input.State, []byte(input.Payload),
		input.Attempts, input.LastError).
		Scan(&out.ID, &out.UserID, &out.ConnectionID, &out.Provider, &out.EventKey, &out.ExternalID, &out.State,
			&out.Payload, &out.Attempts, &out.LastError, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.BankFeedEvent{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListBankFeedEvents(userID domain.ID, state string) []domain.BankFeedEvent {
	query := "SELECT id, user_id, connection_id, provider, event_key, external_transaction_id, state, payload::text, attempts, COALESCE(last_error, ''), created_at, updated_at FROM bank_feed_events"
	args := make([]any, 0, 2)
	predicates := make([]string, 0, 2)
	if userID != "" {
		args = append(args, userID)
		predicates = append(predicates, "user_id=$"+strconv.Itoa(len(args)))
	}
	if state != "" {
		args = append(args, state)
		predicates = append(predicates, "state=$"+strconv.Itoa(len(args)))
	}
	if len(predicates) > 0 {
		query += " WHERE " + strings.Join(predicates, " AND ")
	}
	query += " ORDER BY created_at ASC"

	rows, err := s.pool.Query(context.Background(), query, args...)
	if err != nil {
		return []domain.BankFeedEvent{}
	}
	defer rows.Close()
	out := make([]domain.BankFeedEvent, 0)
	for rows.Next() {
		var e domain.BankFeedEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.ConnectionID, &e.Provider, &e.EventKey, &e.ExternalID, &e.State,
			&e.Payload, &e.Attempts, &e.LastError, &e.CreatedAt, &e.UpdatedAt); err == nil {
			out = append(out, e)
		}
	}
	return out
}

func (s *PostgresStore) GetBankFeedEvent(id domain.ID) (*domain.BankFeedEvent, bool) {
	var out domain.BankFeedEvent
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, connection_id, provider, event_key, external_transaction_id, state, payload::text, attempts, COALESCE(last_error, ''), created_at, updated_at
		FROM bank_feed_events WHERE id=$1`, id)
	if err := row.Scan(&out.ID, &out.UserID, &out.ConnectionID, &out.Provider, &out.EventKey, &out.ExternalID, &out.State,
		&out.Payload, &out.Attempts, &out.LastError, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, false
	}
	return &out, true
}

func (s *PostgresStore) ClaimBankFeedEvent(id domain.ID) (*domain.BankFeedEvent, bool) {
	var out domain.BankFeedEvent
	err := s.pool.QueryRow(context.Background(), `
		UPDATE bank_feed_events
		SET state=$2, attempts=attempts+1, last_error='', updated_at=now()
		WHERE id=$1 AND state=$3
		RETURNING id, user_id, connection_id, provider, event_key, external_transaction_id, state, payload::text,
			attempts, COALESCE(last_error, ''), created_at, updated_at
	`, id, domain.BankFeedEventStateRunning, domain.BankFeedEventStateQueued).
		Scan(&out.ID, &out.UserID, &out.ConnectionID, &out.Provider, &out.EventKey, &out.ExternalID, &out.State,
			&out.Payload, &out.Attempts, &out.LastError, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, false
	}
	return &out, true
}

func (s *PostgresStore) UpdateBankFeedEvent(id domain.ID, mutate func(*domain.BankFeedEvent)) bool {
	current, ok := s.GetBankFeedEvent(id)
	if !ok {
		return false
	}
	mutated := *current
	mutate(&mutated)
	payload := []byte(mutated.Payload)
	_, err := s.pool.Exec(context.Background(), `
		UPDATE bank_feed_events
		SET state=$2, attempts=$3, last_error=$4, payload=$5, updated_at=now()
		WHERE id=$1
	`, id, mutated.State, mutated.Attempts, mutated.LastError, payload)
	return err == nil
}

func (s *PostgresStore) IngestBankFeed(input domain.BankFeedTransaction) (domain.BankFeedTransaction, error) {
	ctx := context.Background()
	var out domain.BankFeedTransaction
	err := s.pool.QueryRow(ctx, `
		INSERT INTO bank_feed_transactions(
			user_id, connection_id, account_id, amount, currency, direction, counterparty, description, occurred_at,
			external_transaction_id, reference, classification_confidence, classification_evidence, posting_state,
			posted_transaction_id, auto_classified, rule_id, source, user_id, sepay_bank_account_id, raw_provider_data, classification_status)
		VALUES($1,$2,$3,CAST($4 AS NUMERIC),$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,NULLIF($19,'')::uuid,NULLIF($20,'')::uuid,NULLIF($21,'')::jsonb,$22)
		ON CONFLICT(user_id, connection_id, external_transaction_id) DO UPDATE
			SET updated_at=EXCLUDED.updated_at
			RETURNING id, user_id, connection_id, account_id,
				amount::text, currency, direction, counterparty, description, occurred_at,
				external_transaction_id, reference, classification_confidence, classification_evidence, posting_state,
			COALESCE(posted_transaction_id::text,''), auto_classified, COALESCE(rule_id::text,''),
			source, COALESCE(user_id::text,''), COALESCE(sepay_bank_account_id::text,''), COALESCE(raw_provider_data::text,''), classification_status, created_at, updated_at
	`, input.UserID, input.ConnectionID, input.AccountID, input.Amount, input.Currency, input.Direction, input.CounterParty,
		input.Description, input.OccurredAt.UTC(), input.ExternalID, input.Reference, input.Confidence, input.Evidence,
		input.PostingState, nilUUID(input.PostedTxnID), input.AutoClassified, nilUUID(input.RuleID), input.Source, input.UserID, input.SePayBankAccountID, input.RawProviderData, input.ClassificationStatus).Scan(
		&out.ID, &out.UserID, &out.ConnectionID, &out.AccountID, &out.Amount, &out.Currency, &out.Direction, &out.CounterParty,
		&out.Description, &out.OccurredAt, &out.ExternalID, &out.Reference, &out.Confidence, &out.Evidence, &out.PostingState,
		&out.PostedTxnID, &out.AutoClassified, &out.RuleID, &out.Source, &out.UserID, &out.SePayBankAccountID, &out.RawProviderData, &out.ClassificationStatus, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) ListBankFeed(userID domain.ID) []domain.BankFeedTransaction {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, connection_id, account_id, amount::text, currency, direction, counterparty, description,
		       occurred_at, external_transaction_id, reference, classification_confidence, classification_evidence, posting_state,
		       COALESCE(posted_transaction_id::text,''), auto_classified, COALESCE(rule_id::text,''), source, COALESCE(user_id::text,''), COALESCE(sepay_bank_account_id::text,''), COALESCE(raw_provider_data::text,''), classification_status, created_at, updated_at
		FROM bank_feed_transactions
		WHERE user_id=$1
		ORDER BY occurred_at DESC`, userID)
	if err != nil {
		return []domain.BankFeedTransaction{}
	}
	defer rows.Close()
	out := make([]domain.BankFeedTransaction, 0)
	for rows.Next() {
		var f domain.BankFeedTransaction
		if err := rows.Scan(&f.ID, &f.UserID, &f.ConnectionID, &f.AccountID, &f.Amount, &f.Currency, &f.Direction,
			&f.CounterParty, &f.Description, &f.OccurredAt, &f.ExternalID, &f.Reference, &f.Confidence, &f.Evidence, &f.PostingState,
			&f.PostedTxnID, &f.AutoClassified, &f.RuleID, &f.Source, &f.UserID, &f.SePayBankAccountID, &f.RawProviderData, &f.ClassificationStatus, &f.CreatedAt, &f.UpdatedAt); err == nil {
			out = append(out, f)
		}
	}
	return out
}

func (s *PostgresStore) ListBankFeedByState(userID domain.ID, state domain.TransactionPostingState) []domain.BankFeedTransaction {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, connection_id, account_id, amount::text, currency, direction, counterparty, description,
		       occurred_at, external_transaction_id, reference, classification_confidence, classification_evidence, posting_state,
		       COALESCE(posted_transaction_id::text,''), auto_classified, COALESCE(rule_id::text,''), source, COALESCE(user_id::text,''), COALESCE(sepay_bank_account_id::text,''), COALESCE(raw_provider_data::text,''), classification_status, created_at, updated_at
		FROM bank_feed_transactions
		WHERE user_id=$1 AND posting_state=$2
		ORDER BY occurred_at DESC`, userID, state)
	if err != nil {
		return []domain.BankFeedTransaction{}
	}
	defer rows.Close()
	out := make([]domain.BankFeedTransaction, 0)
	for rows.Next() {
		var f domain.BankFeedTransaction
		if err := rows.Scan(&f.ID, &f.UserID, &f.ConnectionID, &f.AccountID, &f.Amount, &f.Currency, &f.Direction,
			&f.CounterParty, &f.Description, &f.OccurredAt, &f.ExternalID, &f.Reference, &f.Confidence, &f.Evidence, &f.PostingState,
			&f.PostedTxnID, &f.AutoClassified, &f.RuleID, &f.Source, &f.UserID, &f.SePayBankAccountID, &f.RawProviderData, &f.ClassificationStatus, &f.CreatedAt, &f.UpdatedAt); err == nil {
			out = append(out, f)
		}
	}
	return out
}

func (s *PostgresStore) GetBankFeed(id domain.ID) (*domain.BankFeedTransaction, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, connection_id, account_id, amount::text, currency, direction, counterparty, description,
		       occurred_at, external_transaction_id, reference, classification_confidence, classification_evidence, posting_state,
		       COALESCE(posted_transaction_id::text,''), auto_classified, COALESCE(rule_id::text,''), source, COALESCE(user_id::text,''), COALESCE(sepay_bank_account_id::text,''), COALESCE(raw_provider_data::text,''), classification_status, created_at, updated_at
		FROM bank_feed_transactions WHERE id=$1`, id)
	var f domain.BankFeedTransaction
	if err := row.Scan(&f.ID, &f.UserID, &f.ConnectionID, &f.AccountID, &f.Amount, &f.Currency, &f.Direction, &f.CounterParty, &f.Description,
		&f.OccurredAt, &f.ExternalID, &f.Reference, &f.Confidence, &f.Evidence, &f.PostingState, &f.PostedTxnID, &f.AutoClassified, &f.RuleID, &f.Source, &f.UserID, &f.SePayBankAccountID, &f.RawProviderData, &f.ClassificationStatus, &f.CreatedAt, &f.UpdatedAt); err != nil {
		return nil, false
	}
	return &f, true
}

func (s *PostgresStore) UpdateFeedState(id domain.ID, state domain.TransactionPostingState, reason string) error {
	_, err := s.pool.Exec(context.Background(), `
		UPDATE bank_feed_transactions
		SET posting_state=$2, classification_status=CASE WHEN $2='ignored' THEN 'ignored' ELSE classification_status END, classification_evidence=$3, updated_at=now()
		WHERE id=$1
	`, id, state, reason)
	return err
}

func (s *PostgresStore) UpdateFeed(id domain.ID, mutate func(*domain.BankFeedTransaction)) bool {
	current, ok := s.GetBankFeed(id)
	if !ok {
		return false
	}
	mutated := *current
	mutate(&mutated)
	_, err := s.pool.Exec(context.Background(), `
		UPDATE bank_feed_transactions
		SET classification_confidence=$2, classification_evidence=$3,
		    posting_state=$4, posted_transaction_id=CAST(NULLIF($5, '') AS UUID), auto_classified=$6, rule_id=CAST(NULLIF($7, '') AS UUID), source=$8, classification_status=$9, updated_at=now()
		WHERE id=$1
	`, id, mutated.Confidence, mutated.Evidence, mutated.PostingState, string(mutated.PostedTxnID),
		mutated.AutoClassified, string(mutated.RuleID), mutated.Source, mutated.ClassificationStatus)
	return err == nil
}

func (s *PostgresStore) LinkBankFeedPosting(feedID domain.ID, txnID domain.ID) bool {
	_, err := s.pool.Exec(context.Background(), `
		UPDATE bank_feed_transactions
		SET posted_transaction_id=$2, posting_state='posted', classification_status='confirmed', updated_at=now()
		WHERE id=$1
	`, feedID, txnID)
	return err == nil
}

func (s *PostgresStore) CreateAutomationRule(input domain.AutomationRule) (domain.AutomationRule, error) {
	ctx := context.Background()
	var out domain.AutomationRule
	err := s.pool.QueryRow(ctx, `
		INSERT INTO bank_automation_rules(user_id, account_id, name, priority, predicate, direction, action_type, type, category_id, enabled, content_pattern, reference_pattern, min_amount, max_amount)
		VALUES($1, NULLIF($2,''), $3, $4, $5, $6, $7, $8, NULLIF($9,''), COALESCE($10, TRUE), $11, $12, NULLIF($13,''), NULLIF($14,''))
		RETURNING id, user_id, COALESCE(account_id::text, ''), name, priority, predicate, direction, action_type, type, COALESCE(category_id::text, ''),
		          enabled, content_pattern, reference_pattern, min_amount::text, max_amount::text, created_at, updated_at
	`, input.UserID, nilUUID(input.AccountID), input.Name, input.Priority, input.Predicate, input.Direction, input.ActionType, input.Type,
		nilUUID(input.CategoryID), input.Enabled, input.ContentPattern, input.ReferencePattern, input.MinAmount, input.MaxAmount).Scan(
		&out.ID, &out.UserID, &out.AccountID, &out.Name, &out.Priority, &out.Predicate, &out.Direction, &out.ActionType, &out.Type,
		&out.CategoryID, &out.Enabled, &out.ContentPattern, &out.ReferencePattern, &out.MinAmount, &out.MaxAmount, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) ListUserAutomationRules(userID domain.ID) []domain.AutomationRule {
	return s.ListAutomationRules(userID)
}

func (s *PostgresStore) ListUserRules(userID domain.ID) []domain.AutomationRule {
	return s.ListAutomationRules(userID)
}

func (s *PostgresStore) ListAutomationRules(userID domain.ID) []domain.AutomationRule {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, COALESCE(account_id::text, ''), name, priority, predicate, direction, action_type, type,
		       COALESCE(category_id::text, ''), enabled, content_pattern, reference_pattern, min_amount::text, max_amount::text, created_at, updated_at
		FROM bank_automation_rules
		WHERE user_id=$1
		ORDER BY priority ASC, created_at ASC`, userID)
	if err != nil {
		return []domain.AutomationRule{}
	}
	defer rows.Close()
	out := make([]domain.AutomationRule, 0)
	for rows.Next() {
		var r domain.AutomationRule
		if err := rows.Scan(&r.ID, &r.UserID, &r.AccountID, &r.Name, &r.Priority, &r.Predicate, &r.Direction,
			&r.ActionType, &r.Type, &r.CategoryID, &r.Enabled, &r.ContentPattern, &r.ReferencePattern, &r.MinAmount, &r.MaxAmount, &r.CreatedAt, &r.UpdatedAt); err == nil {
			out = append(out, r)
		}
	}
	return out
}

func (s *PostgresStore) GetUserRules(userID domain.ID, accountID domain.ID, direction string) []domain.AutomationRule {
	query := `
		SELECT id, user_id, COALESCE(account_id::text, ''), name, priority, predicate, direction, action_type, type,
		       COALESCE(category_id::text, ''), enabled, content_pattern, reference_pattern, min_amount::text, max_amount::text, created_at, updated_at
		FROM bank_automation_rules
		WHERE user_id=$1 AND enabled = true`
	args := []any{userID}
	if direction != "" {
		query += " AND (direction = '' OR direction = $2)"
		args = append(args, direction)
	}
	query += " ORDER BY CASE WHEN account_id::text = $3 THEN 0 ELSE 1 END, priority ASC, created_at ASC"
	args = append(args, string(accountID))
	rows, err := s.pool.Query(context.Background(), query, args...)
	if err != nil {
		return []domain.AutomationRule{}
	}
	defer rows.Close()
	out := make([]domain.AutomationRule, 0)
	for rows.Next() {
		var r domain.AutomationRule
		if err := rows.Scan(&r.ID, &r.UserID, &r.AccountID, &r.Name, &r.Priority, &r.Predicate, &r.Direction,
			&r.ActionType, &r.Type, &r.CategoryID, &r.Enabled, &r.ContentPattern, &r.ReferencePattern, &r.MinAmount, &r.MaxAmount, &r.CreatedAt, &r.UpdatedAt); err == nil {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		wa := out[i].AccountID == accountID
		wb := out[j].AccountID == accountID
		if wa != wb {
			return wa && !wb
		}
		return out[i].Priority < out[j].Priority
	})
	return out
}

func (s *PostgresStore) GetAutomationRule(id domain.ID) (*domain.AutomationRule, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, COALESCE(account_id::text, ''), name, priority, predicate, direction, action_type, type,
		       COALESCE(category_id::text, ''), enabled, content_pattern, reference_pattern, min_amount::text, max_amount::text, created_at, updated_at
		FROM bank_automation_rules WHERE id=$1`, id)
	var r domain.AutomationRule
	if err := row.Scan(&r.ID, &r.UserID, &r.AccountID, &r.Name, &r.Priority, &r.Predicate, &r.Direction,
		&r.ActionType, &r.Type, &r.CategoryID, &r.Enabled, &r.ContentPattern, &r.ReferencePattern, &r.MinAmount, &r.MaxAmount, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return nil, false
	}
	return &r, true
}

func (s *PostgresStore) UpdateAutomationRule(id domain.ID, mutate func(*domain.AutomationRule)) bool {
	current, ok := s.GetAutomationRule(id)
	if !ok {
		return false
	}
	mutated := *current
	mutate(&mutated)
	_, err := s.pool.Exec(context.Background(), `
		UPDATE bank_automation_rules
		SET name=$2, predicate=$3, priority=$4, direction=$5, action_type=$6, type=$7, account_id=$8, category_id=$9, enabled=$10,
		    content_pattern=$11, reference_pattern=$12, min_amount=NULLIF($13, ''), max_amount=NULLIF($14, ''), updated_at=now()
		WHERE id=$1
	`, id, mutated.Name, mutated.Predicate, mutated.Priority, mutated.Direction, mutated.ActionType, mutated.Type, nilUUID(mutated.AccountID), nilUUID(mutated.CategoryID), mutated.Enabled,
		mutated.ContentPattern, mutated.ReferencePattern, mutated.MinAmount, mutated.MaxAmount)
	return err == nil
}

func (s *PostgresStore) DeleteAutomationRule(id domain.ID) bool {
	_, err := s.pool.Exec(context.Background(), `DELETE FROM bank_automation_rules WHERE id=$1`, id)
	return err == nil
}

func (s *PostgresStore) CreateBankPaymentRequest(input domain.BankPaymentRequest) (domain.BankPaymentRequest, error) {
	ctx := context.Background()
	var out domain.BankPaymentRequest
	err := s.pool.QueryRow(ctx, `
		INSERT INTO bank_payment_requests(user_id, loan_id, payment_code, amount, currency, expires_at, status, note, source)
		VALUES($1,$2,$3,CAST($4 AS NUMERIC),$5,$6,$7,$8,$9)
		RETURNING id, user_id, loan_id, payment_code, amount::text, currency, expires_at, status, note, source, created_at, updated_at
	`, input.UserID, input.LoanID, input.Code, input.Amount, input.Currency, input.ExpiresAt, input.Status, input.Note, input.Source).Scan(
		&out.ID, &out.UserID, &out.LoanID, &out.Code, &out.Amount, &out.Currency, &out.ExpiresAt, &out.Status, &out.Note, &out.Source, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) GetBankPaymentRequestByCode(userID domain.ID, code string) (*domain.BankPaymentRequest, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, loan_id, payment_code, amount::text, currency, expires_at, status, note, source, created_at, updated_at
		FROM bank_payment_requests
		WHERE user_id=$1 AND payment_code=$2
	`, userID, code)
	var r domain.BankPaymentRequest
	if err := row.Scan(&r.ID, &r.UserID, &r.LoanID, &r.Code, &r.Amount, &r.Currency, &r.ExpiresAt, &r.Status, &r.Note, &r.Source, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return nil, false
	}
	return &r, true
}

func (s *PostgresStore) ListBankPaymentRequests(userID domain.ID) []domain.BankPaymentRequest {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, loan_id, payment_code, amount::text, currency, expires_at, status, note, source, created_at, updated_at
		FROM bank_payment_requests WHERE user_id=$1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return []domain.BankPaymentRequest{}
	}
	defer rows.Close()
	out := make([]domain.BankPaymentRequest, 0)
	for rows.Next() {
		var r domain.BankPaymentRequest
		if err := rows.Scan(&r.ID, &r.UserID, &r.LoanID, &r.Code, &r.Amount, &r.Currency, &r.ExpiresAt, &r.Status, &r.Note, &r.Source, &r.CreatedAt, &r.UpdatedAt); err == nil {
			out = append(out, r)
		}
	}
	return out
}

func (s *PostgresStore) RevokeBankConnection(id domain.ID) (*domain.BankConnection, bool) {
	var out domain.BankConnection
	res, err := s.pool.Exec(context.Background(), `
		UPDATE bank_connections
		SET status='revoked', updated_at=now()
		WHERE id=$1
		`, id)
	if err != nil || res.RowsAffected() == 0 {
		return nil, false
	}
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, provider, external_id, status, scope, COALESCE(bank_code, ''), callback_state,
		       sync_status, COALESCE(last_synced_at, '0001-01-01 00:00:00+00'::timestamptz), COALESCE(last_sync_requested_at, '0001-01-01 00:00:00+00'::timestamptz), created_at, updated_at
		FROM bank_connections WHERE id=$1`, id)
	if err := row.Scan(&out.ID, &out.UserID, &out.Provider, &out.ExternalID, &out.Status, &out.Scope, &out.BankCode,
		&out.CallbackState, &out.SyncStatus, &out.LastSyncedAt, &out.LastSyncRequestedAt, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, false
	}
	return &out, true
}

func (s *PostgresStore) CreateAssistantCommand(input domain.AssistantCommand) (domain.AssistantCommand, error) {
	ctx := context.Background()
	var out domain.AssistantCommand
	if input.Status == "" {
		input.Status = "pending"
	}
	if input.Status == "awaiting_approval" && input.ApprovalID == "" {
		input.ApprovalID = "appr_" + uuid.NewString()
		input.ApprovalExpiresAt = time.Now().UTC().Add(10 * time.Minute)
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO assistant_commands(user_id, command, status, plan, approval_id, approval_expires_at)
		VALUES($1,$2,$3,$4,$5,$6)
		RETURNING id, user_id, command, status, plan, approval_id, approval_expires_at, approval_used_at, created_at, updated_at
	`, input.UserID, input.Command, input.Status, input.Plan, input.ApprovalID, nullTime(input.ApprovalExpiresAt)).Scan(
		&out.ID, &out.UserID, &out.Command, &out.Status, &out.Plan,
		&out.ApprovalID, &out.ApprovalExpiresAt, &out.ApprovalUsedAt, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) GetAssistantCommand(id domain.ID) (*domain.AssistantCommand, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, user_id, command, status, plan, approval_id, approval_expires_at, approval_used_at, created_at, updated_at
		FROM assistant_commands WHERE id=$1`, id)
	var c domain.AssistantCommand
	if err := row.Scan(&c.ID, &c.UserID, &c.Command, &c.Status, &c.Plan,
		&c.ApprovalID, &c.ApprovalExpiresAt, &c.ApprovalUsedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, false
	}
	return &c, true
}

func (s *PostgresStore) ListAssistantCommands(userID domain.ID) []domain.AssistantCommand {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, user_id, command, status, plan, approval_id, approval_expires_at, approval_used_at, created_at, updated_at
		FROM assistant_commands WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return []domain.AssistantCommand{}
	}
	defer rows.Close()
	out := make([]domain.AssistantCommand, 0)
	for rows.Next() {
		var c domain.AssistantCommand
		if err := rows.Scan(&c.ID, &c.UserID, &c.Command, &c.Status, &c.Plan, &c.ApprovalID, &c.ApprovalExpiresAt, &c.ApprovalUsedAt, &c.CreatedAt, &c.UpdatedAt); err == nil {
			out = append(out, c)
		}
	}
	return out
}

func (s *PostgresStore) UpdateAssistantCommand(id domain.ID, mutate func(*domain.AssistantCommand)) (*domain.AssistantCommand, error) {
	if mutate == nil {
		return nil, errors.New("missing update mutation")
	}
	ctx := context.Background()
	current, ok := s.GetAssistantCommand(id)
	if !ok {
		return nil, errors.New("assistant command not found")
	}
	cp := *current
	mutate(&cp)
	_, err := s.pool.Exec(ctx, `
		UPDATE assistant_commands
		SET user_id=$2, command=$3, status=$4, plan=$5,
			approval_id=$6, approval_expires_at=$7, approval_used_at=$8, updated_at=now()
		WHERE id=$1
	`, id, cp.UserID, cp.Command, cp.Status, cp.Plan,
		cp.ApprovalID, nullTime(cp.ApprovalExpiresAt), nullTimePtr(cp.ApprovalUsedAt))
	if err != nil {
		return nil, err
	}
	command, ok := s.GetAssistantCommand(id)
	if !ok {
		return nil, errors.New("assistant command not found")
	}
	return command, nil
}

func (s *PostgresStore) RecordIdempotency(key string) bool {
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO idempotency_keys(key, created_at)
		VALUES($1, now())
	`, key)
	return err == nil
}

func (s *PostgresStore) ClearIdempotencyOlderThan(cutoff time.Time) int {
	res, err := s.pool.Exec(context.Background(), `
		DELETE FROM idempotency_keys
		WHERE created_at < $1
	`, cutoff)
	if err != nil {
		return 0
	}
	return int(res.RowsAffected())
}

func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}

func nilUUID(id domain.ID) any {
	normalized := strings.TrimSpace(string(id))
	if normalized == "" {
		return nil
	}
	if _, err := uuid.Parse(normalized); err != nil {
		return nil
	}
	return domain.ID(normalized)
}

func defaultBaseCurrency(v string) string {
	if strings.TrimSpace(v) == "" {
		return "VND"
	}
	return v
}

func defaultCurrency(v string) string {
	if strings.TrimSpace(v) == "" {
		return "VND"
	}
	return v
}

func defaultStatus(v domain.TransactionStatus) domain.TransactionStatus {
	if v == "" {
		return domain.TransactionStatusPosted
	}
	return v
}

func defaultLoanStatus(v domain.LoanStatus) domain.LoanStatus {
	if strings.TrimSpace(string(v)) == "" {
		return domain.LoanStatusDraft
	}
	return v
}

func nullString(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func nullTimePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

func (s *PostgresStore) UpsertSePayUserProfile(input domain.SePayUserProfile) (domain.SePayUserProfile, error) {
	var out domain.SePayUserProfile
	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO sepay_user_profiles(user_id, company_xid, status, linked_at, last_synced_at)
		VALUES($1,$2,$3,NULLIF($4,'')::timestamptz,NULLIF($5,'')::timestamptz)
		ON CONFLICT(user_id) DO UPDATE SET company_xid=EXCLUDED.company_xid, status=EXCLUDED.status, linked_at=EXCLUDED.linked_at, last_synced_at=EXCLUDED.last_synced_at, updated_at=now()
		RETURNING user_id, company_xid, status, COALESCE(linked_at,'epoch'), COALESCE(last_synced_at,'epoch'), created_at, updated_at`,
		input.UserID, input.CompanyXID, input.Status, nullableTimeString(input.LinkedAt), nullableTimeString(input.LastSyncedAt)).Scan(&out.UserID, &out.CompanyXID, &out.Status, &out.LinkedAt, &out.LastSyncedAt, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func nullableTimeString(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func (s *PostgresStore) GetSePayUserProfile(userID domain.ID) (*domain.SePayUserProfile, bool) {
	var out domain.SePayUserProfile
	err := s.pool.QueryRow(context.Background(), `SELECT user_id, company_xid,status,COALESCE(linked_at,'epoch'),COALESCE(last_synced_at,'epoch'),created_at,updated_at FROM sepay_user_profiles WHERE user_id=$1`, userID).Scan(&out.UserID, &out.CompanyXID, &out.Status, &out.LinkedAt, &out.LastSyncedAt, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, false
	}
	return &out, true
}

func scanSePayBankAccount(row interface{ Scan(...any) error }) (*domain.SePayBankAccount, error) {
	var out domain.SePayBankAccount
	err := row.Scan(&out.ID, &out.UserID, &out.BankAccountXID, &out.BankCode, &out.BankName, &out.AccountNumberMasked, &out.SupportsIn, &out.SupportsOut, &out.Status, &out.CreatedAt, &out.UpdatedAt)
	return &out, err
}

func (s *PostgresStore) UpsertSePayBankAccount(input domain.SePayBankAccount) (domain.SePayBankAccount, error) {
	row := s.pool.QueryRow(context.Background(), `INSERT INTO sepay_bank_accounts(user_id,bank_account_xid,bank_code,bank_name,account_number_masked,supports_in,supports_out,status) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(bank_account_xid) DO UPDATE SET bank_code=EXCLUDED.bank_code,bank_name=EXCLUDED.bank_name,account_number_masked=EXCLUDED.account_number_masked,supports_in=EXCLUDED.supports_in,supports_out=EXCLUDED.supports_out,status=EXCLUDED.status,updated_at=now() WHERE sepay_bank_accounts.user_id=EXCLUDED.user_id RETURNING id,user_id,bank_account_xid,COALESCE(bank_code,''),COALESCE(bank_name,''),account_number_masked,supports_in,supports_out,status,created_at,updated_at`, input.UserID, input.BankAccountXID, input.BankCode, input.BankName, input.AccountNumberMasked, input.SupportsIn, input.SupportsOut, input.Status)
	out, err := scanSePayBankAccount(row)
	if err != nil {
		return domain.SePayBankAccount{}, err
	}
	return *out, nil
}

func (s *PostgresStore) ListSePayBankAccounts(userID domain.ID) []domain.SePayBankAccount {
	rows, err := s.pool.Query(context.Background(), `SELECT id,user_id,bank_account_xid,COALESCE(bank_code,''),COALESCE(bank_name,''),account_number_masked,supports_in,supports_out,status,created_at,updated_at FROM sepay_bank_accounts WHERE user_id=$1 ORDER BY created_at`, userID)
	if err != nil {
		return []domain.SePayBankAccount{}
	}
	defer rows.Close()
	out := []domain.SePayBankAccount{}
	for rows.Next() {
		item, err := scanSePayBankAccount(rows)
		if err == nil {
			out = append(out, *item)
		}
	}
	return out
}

func (s *PostgresStore) GetSePayBankAccount(id domain.ID) (*domain.SePayBankAccount, bool) {
	item, err := scanSePayBankAccount(s.pool.QueryRow(context.Background(), `SELECT id,user_id,bank_account_xid,COALESCE(bank_code,''),COALESCE(bank_name,''),account_number_masked,supports_in,supports_out,status,created_at,updated_at FROM sepay_bank_accounts WHERE id=$1`, id))
	if err != nil {
		return nil, false
	}
	return item, true
}
func (s *PostgresStore) GetSePayBankAccountByXID(xid string) (*domain.SePayBankAccount, bool) {
	item, err := scanSePayBankAccount(s.pool.QueryRow(context.Background(), `SELECT id,user_id,bank_account_xid,COALESCE(bank_code,''),COALESCE(bank_name,''),account_number_masked,supports_in,supports_out,status,created_at,updated_at FROM sepay_bank_accounts WHERE bank_account_xid=$1`, xid))
	if err != nil {
		return nil, false
	}
	return item, true
}
func (s *PostgresStore) SetSePayBankAccountStatus(id domain.ID, status string) bool {
	tag, err := s.pool.Exec(context.Background(), `UPDATE sepay_bank_accounts SET status=$2,updated_at=now() WHERE id=$1`, id, status)
	return err == nil && tag.RowsAffected() == 1
}

func scanBankAccountMapping(row interface{ Scan(...any) error }) (*domain.BankAccountMapping, error) {
	var out domain.BankAccountMapping
	err := row.Scan(&out.ID, &out.SePayBankAccountID, &out.UserID, &out.UserID, &out.AccountID, &out.Status, &out.CreatedAt, &out.UpdatedAt)
	return &out, err
}
func (s *PostgresStore) UpsertBankAccountMapping(input domain.BankAccountMapping) (domain.BankAccountMapping, error) {
	row := s.pool.QueryRow(context.Background(), `INSERT INTO bank_account_mappings(sepay_bank_account_id,user_id,user_id,account_id,status)VALUES($1,$2,$3,$4,$5)ON CONFLICT(sepay_bank_account_id)DO UPDATE SET user_id=EXCLUDED.user_id,account_id=EXCLUDED.account_id,status=EXCLUDED.status,updated_at=now() WHERE bank_account_mappings.user_id=EXCLUDED.user_id RETURNING id,sepay_bank_account_id,user_id,user_id,account_id,status,created_at,updated_at`, input.SePayBankAccountID, input.UserID, input.UserID, input.AccountID, input.Status)
	out, err := scanBankAccountMapping(row)
	if err != nil {
		return domain.BankAccountMapping{}, err
	}
	return *out, nil
}
func (s *PostgresStore) GetBankAccountMapping(id domain.ID) (*domain.BankAccountMapping, bool) {
	out, err := scanBankAccountMapping(s.pool.QueryRow(context.Background(), `SELECT id,sepay_bank_account_id,user_id,user_id,account_id,status,created_at,updated_at FROM bank_account_mappings WHERE sepay_bank_account_id=$1`, id))
	if err != nil {
		return nil, false
	}
	return out, true
}
func (s *PostgresStore) DeactivateBankAccountMapping(id domain.ID) bool {
	tag, err := s.pool.Exec(context.Background(), `UPDATE bank_account_mappings SET status='inactive',updated_at=now() WHERE sepay_bank_account_id=$1`, id)
	return err == nil && tag.RowsAffected() == 1
}

func (s *PostgresStore) CreateSePayLinkSession(xid string, userID domain.ID, expiresAt time.Time) error {
	_, err := s.pool.Exec(context.Background(), `INSERT INTO sepay_link_sessions(xid,user_id,expires_at)VALUES($1,$2,NULLIF($3,'')::timestamptz)ON CONFLICT(xid)DO UPDATE SET user_id=EXCLUDED.user_id,expires_at=EXCLUDED.expires_at,status='pending',updated_at=now()`, xid, userID, nullableTimeString(expiresAt))
	return err
}
func (s *PostgresStore) GetSePayLinkSessionUser(xid string) (domain.ID, bool) {
	var userID domain.ID
	err := s.pool.QueryRow(context.Background(), `SELECT user_id FROM sepay_link_sessions WHERE xid=$1 AND (expires_at IS NULL OR expires_at>=now())`, xid).Scan(&userID)
	return userID, err == nil
}
func (s *PostgresStore) CompleteSePayLinkSession(xid string) bool {
	tag, err := s.pool.Exec(context.Background(), `UPDATE sepay_link_sessions SET status='completed',updated_at=now() WHERE xid=$1`, xid)
	return err == nil && tag.RowsAffected() == 1
}

func (s *PostgresStore) QuarantineSePayEvent(input domain.SePayUnmappedEvent) (domain.SePayUnmappedEvent, error) {
	var out domain.SePayUnmappedEvent
	err := s.pool.QueryRow(context.Background(), `INSERT INTO sepay_unmapped_events(provider,bank_account_xid,transaction_id,payload,status)VALUES($1,$2,$3,$4,$5)ON CONFLICT(provider,bank_account_xid,transaction_id)DO UPDATE SET updated_at=now() RETURNING id,provider,bank_account_xid,transaction_id,payload::text,status,created_at,updated_at`, input.Provider, input.BankAccountXID, input.TransactionID, []byte(input.Payload), input.Status).Scan(&out.ID, &out.Provider, &out.BankAccountXID, &out.TransactionID, &out.Payload, &out.Status, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func (s *PostgresStore) CreateTransactionSuggestion(input domain.TransactionSuggestion) (domain.TransactionSuggestion, error) {
	var out domain.TransactionSuggestion
	err := s.pool.QueryRow(context.Background(), `INSERT INTO transaction_suggestions(bank_feed_transaction_id,suggested_name,suggested_category_id,source,confidence,reason,version)VALUES($1,$2,NULLIF($3,'')::uuid,$4,$5,$6,$7)RETURNING id,bank_feed_transaction_id,COALESCE(suggested_name,''),COALESCE(suggested_category_id::text,''),source,confidence,reason,version,created_at,updated_at`, input.BankFeedTransactionID, input.SuggestedName, input.SuggestedCategoryID, input.Source, input.Confidence, input.Reason, input.Version).Scan(&out.ID, &out.BankFeedTransactionID, &out.SuggestedName, &out.SuggestedCategoryID, &out.Source, &out.Confidence, &out.Reason, &out.Version, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}
func (s *PostgresStore) ListTransactionSuggestions(feedID domain.ID) []domain.TransactionSuggestion {
	rows, err := s.pool.Query(context.Background(), `SELECT id,bank_feed_transaction_id,COALESCE(suggested_name,''),COALESCE(suggested_category_id::text,''),source,confidence,reason,version,created_at,updated_at FROM transaction_suggestions WHERE bank_feed_transaction_id=$1 ORDER BY created_at`, feedID)
	if err != nil {
		return []domain.TransactionSuggestion{}
	}
	defer rows.Close()
	out := []domain.TransactionSuggestion{}
	for rows.Next() {
		var item domain.TransactionSuggestion
		if rows.Scan(&item.ID, &item.BankFeedTransactionID, &item.SuggestedName, &item.SuggestedCategoryID, &item.Source, &item.Confidence, &item.Reason, &item.Version, &item.CreatedAt, &item.UpdatedAt) == nil {
			out = append(out, item)
		}
	}
	return out
}
func (s *PostgresStore) CreateClassificationFeedback(input domain.ClassificationFeedback) (domain.ClassificationFeedback, error) {
	var out domain.ClassificationFeedback
	err := s.pool.QueryRow(context.Background(), `INSERT INTO classification_feedback(bank_feed_transaction_id,user_id,action,name,category_id,account_id,transaction_type,note,remember_choice)VALUES($1,$2,$3,$4,NULLIF($5,'')::uuid,NULLIF($6,'')::uuid,$7,$8,$9)RETURNING id,bank_feed_transaction_id,user_id,action,COALESCE(name,''),COALESCE(category_id::text,''),COALESCE(account_id::text,''),COALESCE(transaction_type,''),COALESCE(note,''),remember_choice,created_at,updated_at`, input.BankFeedTransactionID, input.UserID, input.Action, input.Name, input.CategoryID, input.AccountID, input.TransactionType, input.Note, input.RememberChoice).Scan(&out.ID, &out.BankFeedTransactionID, &out.UserID, &out.Action, &out.Name, &out.CategoryID, &out.AccountID, &out.TransactionType, &out.Note, &out.RememberChoice, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}
func (s *PostgresStore) ListClassificationFeedback(userID domain.ID) []domain.ClassificationFeedback {
	rows, err := s.pool.Query(context.Background(), `SELECT id,bank_feed_transaction_id,user_id,action,COALESCE(name,''),COALESCE(category_id::text,''),COALESCE(account_id::text,''),COALESCE(transaction_type,''),COALESCE(note,''),remember_choice,created_at,updated_at FROM classification_feedback WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return []domain.ClassificationFeedback{}
	}
	defer rows.Close()
	out := []domain.ClassificationFeedback{}
	for rows.Next() {
		var item domain.ClassificationFeedback
		if rows.Scan(&item.ID, &item.BankFeedTransactionID, &item.UserID, &item.Action, &item.Name, &item.CategoryID, &item.AccountID, &item.TransactionType, &item.Note, &item.RememberChoice, &item.CreatedAt, &item.UpdatedAt) == nil {
			out = append(out, item)
		}
	}
	return out
}

func (s *PostgresStore) CreateBankReconciliation(input domain.BankReconciliation) (domain.BankReconciliation, error) {
	if input.ID == "" {
		input.ID = domain.ID(uuid.New().String())
	}
	if input.CreatedAt.IsZero() {
		input.CreatedAt = time.Now()
	}
	if input.UpdatedAt.IsZero() {
		input.UpdatedAt = time.Now()
	}
	return input, nil
}

func (s *PostgresStore) ListBankReconciliations(userID domain.ID, connectionID domain.ID) []domain.BankReconciliation {
	return []domain.BankReconciliation{}
}

func (s *PostgresStore) ListAllBankConnections() []domain.BankConnection {
	return s.ListBankConnections("")
}

func init() {
	// keep compatibility for tests that may call NewPostgresStore without explicit init in this project
	_ = uuid.NewString
}

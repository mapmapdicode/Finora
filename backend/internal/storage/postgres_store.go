package storage

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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
		INSERT INTO users(email, name, password_hash)
		VALUES($1, $2, $3)
		RETURNING id
	`, email, name, password).Scan(&id)
	if err != nil {
		return ""
	}
	return id
}

func (s *PostgresStore) CreateAuditLog(input domain.AuditLog) (domain.AuditLog, error) {
	if strings.TrimSpace(string(input.WorkspaceID)) == "" {
		return domain.AuditLog{}, errors.New("workspaceId is required")
	}
	stmt := `
		INSERT INTO audit_logs (
			workspace_id, actor_id, actor_role, action, target_type, target_id, request_id,
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
		RETURNING id, workspace_id, actor_id, actor_role, action, target_type, COALESCE(target_id::text, ''),
			request_id, path, method, policy_decision, result, reason, correlation_id,
			COALESCE(before_json::text, ''), COALESCE(after_json::text, ''), created_at, updated_at`
	out := domain.AuditLog{}
	if strings.TrimSpace(input.Action) == "" {
		input.Action = "unknown"
	}
	err := s.pool.QueryRow(context.Background(), stmt,
		input.WorkspaceID, input.ActorID, input.ActorRole, input.Action, input.TargetType,
		input.TargetID, input.RequestID, input.Path, input.Method, input.PolicyDecision, input.Result,
		input.Reason, input.CorrelationID, input.BeforeJSON, input.AfterJSON).Scan(
		&out.ID, &out.WorkspaceID, &out.ActorID, &out.ActorRole, &out.Action, &out.TargetType,
		&out.TargetID, &out.RequestID, &out.Path, &out.Method, &out.PolicyDecision, &out.Result,
		&out.Reason, &out.CorrelationID, &out.BeforeJSON, &out.AfterJSON, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.AuditLog{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListAuditLogs(workspaceID domain.ID) []domain.AuditLog {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, actor_id, actor_role, action, target_type, COALESCE(target_id::text, ''),
		       request_id, path, method, policy_decision, result, reason, correlation_id,
		       COALESCE(before_json::text, ''), COALESCE(after_json::text, ''), created_at, updated_at
		FROM audit_logs WHERE workspace_id=$1 ORDER BY created_at DESC
	`, workspaceID)
	if err != nil {
		return []domain.AuditLog{}
	}
	defer rows.Close()
	out := make([]domain.AuditLog, 0)
	for rows.Next() {
		var item domain.AuditLog
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.ActorID, &item.ActorRole, &item.Action, &item.TargetType,
			&item.TargetID, &item.RequestID, &item.Path, &item.Method, &item.PolicyDecision, &item.Result,
			&item.Reason, &item.CorrelationID, &item.BeforeJSON, &item.AfterJSON, &item.CreatedAt, &item.UpdatedAt); err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (s *PostgresStore) GetUser(id domain.ID) (*domain.User, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users WHERE id=$1`, id)
	var u domain.User
	var pass string
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &pass, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, false
	}
	u.Password = pass
	return &u, true
}

func (s *PostgresStore) GetUserByID(id domain.ID) (*domain.User, bool) {
	return s.GetUser(id)
}

func (s *PostgresStore) GetUserByEmail(email string) (*domain.User, bool) {
	if email == "" {
		return nil, false
	}
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users WHERE email=$1`, email)
	var u domain.User
	var pass string
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &pass, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, false
	}
	u.Password = pass
	return &u, true
}

func (s *PostgresStore) GetWorkspace(id domain.ID) (*domain.Workspace, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, name, base_currency, COALESCE(fiscal_year_end::text, ''),
		       created_at, updated_at
		FROM workspaces WHERE id=$1`, id)
	var ws domain.Workspace
	if err := row.Scan(&ws.ID, &ws.Name, &ws.BaseCurrency, &ws.FiscalYearEnd, &ws.CreatedAt, &ws.UpdatedAt); err != nil {
		return nil, false
	}
	return &ws, true
}

func (s *PostgresStore) CreateWorkspace(name, baseCurrency string, ownerID domain.ID) (*domain.Workspace, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var ws domain.Workspace
	err = tx.QueryRow(ctx, `
		INSERT INTO workspaces(name, base_currency)
		VALUES($1, $2)
		RETURNING id, name, base_currency, COALESCE(fiscal_year_end::text, ''), created_at, updated_at
	`, name, baseCurrency).Scan(&ws.ID, &ws.Name, &ws.BaseCurrency, &ws.FiscalYearEnd, &ws.CreatedAt, &ws.UpdatedAt)
	if err != nil {
		return nil, err
	}
	ws.BaseCurrency = defaultBaseCurrency(ws.BaseCurrency)

	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_members(workspace_id, user_id, role)
		VALUES($1, $2, 'owner')
	`, ws.ID, ownerID)
	if err != nil {
		return nil, err
	}

	// Keep one default portfolio for onboarding flows.
	var portfolioID domain.ID
	err = tx.QueryRow(ctx, `
		INSERT INTO portfolios(workspace_id, name, base_currency)
		VALUES($1, 'Default', $2)
		RETURNING id
	`, ws.ID, ws.BaseCurrency).Scan(&portfolioID)
	if err != nil {
		return nil, err
	}
	_ = portfolioID

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &ws, nil
}

func (s *PostgresStore) ListWorkspaces(userID domain.ID) []domain.Workspace {
	rows, err := s.pool.Query(context.Background(), `
		SELECT w.id, w.name, w.base_currency, COALESCE(w.fiscal_year_end::text, ''), w.created_at, w.updated_at
		FROM workspaces w
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE m.user_id=$1
		ORDER BY w.created_at DESC
	`, userID)
	if err != nil {
		return []domain.Workspace{}
	}
	defer rows.Close()
	out := make([]domain.Workspace, 0)
	for rows.Next() {
		var ws domain.Workspace
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.BaseCurrency, &ws.FiscalYearEnd, &ws.CreatedAt, &ws.UpdatedAt); err != nil {
			continue
		}
		out = append(out, ws)
	}
	return out
}

func (s *PostgresStore) GetWorkspaceMemberRole(userID, workspaceID domain.ID) (domain.Role, bool) {
	if userID == "" || workspaceID == "" {
		return "", false
	}
	var role domain.Role
	err := s.pool.QueryRow(context.Background(), `
		SELECT role
		FROM workspace_members
		WHERE workspace_id=$1 AND user_id=$2
	`, workspaceID, userID).Scan(&role)
	if err != nil {
		return "", false
	}
	return role, true
}

func (s *PostgresStore) GetPortfolio(id domain.ID) (*domain.Portfolio, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, workspace_id, name, base_currency, created_at, updated_at
		FROM portfolios WHERE id=$1`, id)
	var p domain.Portfolio
	if err := row.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.BaseCurrency, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, false
	}
	return &p, true
}

func (s *PostgresStore) CreatePortfolio(input domain.Portfolio) (domain.Portfolio, error) {
	ctx := context.Background()
	var out domain.Portfolio
	err := s.pool.QueryRow(ctx, `
		INSERT INTO portfolios(workspace_id, name, base_currency)
		VALUES($1, $2, $3)
		RETURNING id, workspace_id, name, base_currency, created_at, updated_at
	`, input.WorkspaceID, input.Name, defaultBaseCurrency(input.BaseCurrency)).Scan(&out.ID, &out.WorkspaceID, &out.Name, &out.BaseCurrency, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.Portfolio{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListPortfolios(workspaceID domain.ID) []domain.Portfolio {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, name, base_currency, created_at, updated_at
		FROM portfolios
		WHERE workspace_id=$1
		ORDER BY name ASC`, workspaceID)
	if err != nil {
		return []domain.Portfolio{}
	}
	defer rows.Close()
	out := make([]domain.Portfolio, 0)
	for rows.Next() {
		var p domain.Portfolio
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.BaseCurrency, &p.CreatedAt, &p.UpdatedAt); err == nil {
			out = append(out, p)
		}
	}
	return out
}

func (s *PostgresStore) FirstPortfolio(workspaceID domain.ID) (domain.Portfolio, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, workspace_id, name, base_currency, created_at, updated_at
		FROM portfolios
		WHERE workspace_id=$1
		ORDER BY created_at ASC
		LIMIT 1
	`, workspaceID)
	var p domain.Portfolio
	if err := row.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.BaseCurrency, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return domain.Portfolio{}, false
	}
	return p, true
}

func (s *PostgresStore) GetAccount(id domain.ID) (*domain.Account, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, workspace_id, portfolio_id, name, type, currency, created_at, updated_at
		FROM accounts WHERE id=$1`, id)
	var a domain.Account
	if err := row.Scan(&a.ID, &a.WorkspaceID, &a.PortfolioID, &a.Name, &a.Type, &a.Currency, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, false
	}
	return &a, true
}

func (s *PostgresStore) CreateAccount(input domain.Account) (domain.Account, error) {
	ctx := context.Background()
	var out domain.Account
	err := s.pool.QueryRow(ctx, `
		INSERT INTO accounts(workspace_id, portfolio_id, name, type, currency)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, workspace_id, portfolio_id, name, type, currency, created_at, updated_at
	`, input.WorkspaceID, nilUUID(input.PortfolioID), input.Name, input.Type, defaultCurrency(input.Currency)).Scan(
		&out.ID, &out.WorkspaceID, &out.PortfolioID, &out.Name, &out.Type, &out.Currency, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.Account{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListAccounts(workspaceID domain.ID) []domain.Account {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, portfolio_id, name, type, currency, created_at, updated_at
		FROM accounts WHERE workspace_id=$1
		ORDER BY name ASC
	`, workspaceID)
	if err != nil {
		return []domain.Account{}
	}
	defer rows.Close()
	out := make([]domain.Account, 0)
	for rows.Next() {
		var a domain.Account
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.PortfolioID, &a.Name, &a.Type, &a.Currency, &a.CreatedAt, &a.UpdatedAt); err == nil {
			out = append(out, a)
		}
	}
	return out
}

func (s *PostgresStore) CreateTransactionStrict(input domain.Transaction) (domain.Transaction, error) {
	if input.WorkspaceID == "" || input.AccountID == "" || input.Currency == "" {
		return domain.Transaction{}, errors.New("workspaceId, accountId and currency are required")
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
	var accountWorkspace domain.ID
	if err := s.pool.QueryRow(context.Background(), `SELECT workspace_id FROM accounts WHERE id=$1`, input.AccountID).Scan(&accountWorkspace); err != nil {
		return domain.Transaction{}, errors.New("accountId does not exist")
	}
	if accountWorkspace != input.WorkspaceID {
		return domain.Transaction{}, errors.New("account does not belong to workspace")
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
		INSERT INTO transactions(workspace_id, account_id, category_id, portfolio_id, type, amount, currency, note, occurred_at, status, source)
		VALUES($1, $2, $3, $4, $5, CAST($6 as NUMERIC), $7, $8, $9, $10)
		RETURNING id, workspace_id, account_id, COALESCE(category_id::text, ''), COALESCE(portfolio_id::text, ''),
		          type, amount::text, currency, note, occurred_at, status, source, created_at, updated_at
	`, input.WorkspaceID, input.AccountID, nilUUID(input.CategoryID), nilUUID(input.PortfolioID), input.Type,
		input.Amount, input.Currency, input.Note, input.OccurredAt.UTC(), defaultStatus(input.Status), input.Source).Scan(
		&out.ID, &out.WorkspaceID, &out.AccountID, &out.CategoryID, &out.PortfolioID,
		&out.Type, &out.Amount, &out.Currency, &out.Note, &out.OccurredAt, &out.Status, &out.Source, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.Transaction{}, err
	}
	return out, nil
}

func (s *PostgresStore) GetTransaction(id domain.ID) (*domain.Transaction, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, workspace_id, account_id, COALESCE(category_id::text, ''), COALESCE(portfolio_id::text, ''),
		       type, amount::text, currency, note, occurred_at, status, source, created_at, updated_at
		FROM transactions WHERE id=$1
	`, id)
	var t domain.Transaction
	if err := row.Scan(&t.ID, &t.WorkspaceID, &t.AccountID, &t.CategoryID, &t.PortfolioID,
		&t.Type, &t.Amount, &t.Currency, &t.Note, &t.OccurredAt, &t.Status, &t.Source, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, false
	}
	return &t, true
}

func (s *PostgresStore) ListTransactions(workspaceID domain.ID, accountID domain.ID) []domain.Transaction {
	query := `
		SELECT id, workspace_id, account_id, COALESCE(category_id::text, ''), COALESCE(portfolio_id::text, ''),
		       type, amount::text, currency, note, occurred_at, status, source, created_at, updated_at
		FROM transactions
		WHERE workspace_id=$1`
	args := []any{workspaceID}
	if accountID != "" {
		query += " AND account_id=$2"
		args = append(args, accountID)
	}
	query += " ORDER BY occurred_at DESC"
	rows, err := s.pool.Query(context.Background(), query, args...)
	if err != nil {
		return []domain.Transaction{}
	}
	defer rows.Close()
	out := make([]domain.Transaction, 0)
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.WorkspaceID, &t.AccountID, &t.CategoryID, &t.PortfolioID,
			&t.Type, &t.Amount, &t.Currency, &t.Note, &t.OccurredAt, &t.Status, &t.Source, &t.CreatedAt, &t.UpdatedAt); err == nil {
			out = append(out, t)
		}
	}
	return out
}

func (s *PostgresStore) CreateTransfer(input domain.Transfer) (domain.Transfer, error) {
	ctx := context.Background()
	var out domain.Transfer
	err := s.pool.QueryRow(ctx, `
		INSERT INTO transfers(workspace_id, from_account_id, to_account_id, amount, currency, note, occurred_at)
		VALUES($1, $2, $3, CAST($4 as NUMERIC), $5, $6, $7)
		RETURNING id, workspace_id, from_account_id, to_account_id, amount::text, currency, note, occurred_at, created_at, updated_at
	`, input.WorkspaceID, input.FromAccountID, input.ToAccountID, input.Amount, defaultCurrency(input.Currency), input.Note, input.OccurredAt.UTC()).Scan(
		&out.ID, &out.WorkspaceID, &out.FromAccountID, &out.ToAccountID, &out.Amount, &out.Currency, &out.Note, &out.OccurredAt, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.Transfer{}, err
	}
	return out, nil
}

func (s *PostgresStore) GetLoan(id domain.ID) (*domain.Loan, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, workspace_id, portfolio_id, counterparty, direction, principal_initial::text, principal_balance::text, annual_rate::text,
		       COALESCE(day_count_basis, ''), start_at, due_at, status, interest_compounding, created_at, updated_at
		FROM loans WHERE id=$1`, id)
	var l domain.Loan
	if err := row.Scan(&l.ID, &l.WorkspaceID, &l.PortfolioID, &l.Counterparty, &l.Direction,
		&l.PrincipalInitial, &l.PrincipalBalance, &l.AnnualRate, &l.DayCountBasis, &l.StartAt, &l.DueAt,
		&l.Status, &l.InterestCompound, &l.CreatedAt, &l.UpdatedAt); err != nil {
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
		    counterparty=$7,
		    status=$8,
		    interest_compounding=$9,
		    start_at=$10,
		    due_at=$11,
		    updated_at=now()
		WHERE id=$1
	`, id, mutated.PrincipalBalance, mutated.PrincipalInitial, mutated.AnnualRate, mutated.DayCountBasis, mutated.Direction,
		mutated.Counterparty, mutated.Status, mutated.InterestCompound, mutated.StartAt, mutated.DueAt)
	return err == nil
}

func (s *PostgresStore) CreateLoan(input domain.Loan) (domain.Loan, error) {
	ctx := context.Background()
	var out domain.Loan
	err := s.pool.QueryRow(ctx, `
		INSERT INTO loans(workspace_id, portfolio_id, counterparty, direction, principal_initial, principal_balance, annual_rate, day_count_basis, start_at, due_at, status, interest_compounding)
		VALUES($1, $2, $3, $4, CAST($5 AS NUMERIC), CAST($6 AS NUMERIC), CAST($7 AS NUMERIC), $8, $9, $10, $11, $12)
		RETURNING id, workspace_id, portfolio_id, counterparty, direction, principal_initial::text, principal_balance::text, annual_rate::text, COALESCE(day_count_basis, ''), start_at, due_at, status, interest_compounding, created_at, updated_at
	`, input.WorkspaceID, nilUUID(input.PortfolioID), input.Counterparty, input.Direction, input.PrincipalInitial, input.PrincipalBalance,
		input.AnnualRate, nullString(input.DayCountBasis), input.StartAt, input.DueAt, defaultLoanStatus(input.Status), input.InterestCompound).Scan(
		&out.ID, &out.WorkspaceID, &out.PortfolioID, &out.Counterparty, &out.Direction,
		&out.PrincipalInitial, &out.PrincipalBalance, &out.AnnualRate, &out.DayCountBasis,
		&out.StartAt, &out.DueAt, &out.Status, &out.InterestCompound, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.Loan{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListLoans(workspaceID domain.ID) []domain.Loan {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, portfolio_id, counterparty, direction, principal_initial::text, principal_balance::text,
		       annual_rate::text, COALESCE(day_count_basis,''), start_at, due_at, status, interest_compounding, created_at, updated_at
		FROM loans WHERE workspace_id=$1 ORDER BY created_at DESC`, workspaceID)
	if err != nil {
		return []domain.Loan{}
	}
	defer rows.Close()
	out := make([]domain.Loan, 0)
	for rows.Next() {
		var l domain.Loan
		if err := rows.Scan(&l.ID, &l.WorkspaceID, &l.PortfolioID, &l.Counterparty, &l.Direction, &l.PrincipalInitial, &l.PrincipalBalance,
			&l.AnnualRate, &l.DayCountBasis, &l.StartAt, &l.DueAt, &l.Status, &l.InterestCompound, &l.CreatedAt, &l.UpdatedAt); err == nil {
			out = append(out, l)
		}
	}
	return out
}

func (s *PostgresStore) CreateLoanPayment(input domain.LoanPayment) (domain.LoanPayment, error) {
	ctx := context.Background()
	var out domain.LoanPayment
	err := s.pool.QueryRow(ctx, `
		INSERT INTO loan_payments(workspace_id, loan_id, account_id, transaction_id, principal_amount, interest_amount, fee_amount, waived_amount, occurred_at)
		VALUES($1, $2, $3, NULLIF($4, ''), CAST($5 AS NUMERIC), CAST($6 AS NUMERIC), CAST($7 AS NUMERIC), CAST($8 AS NUMERIC), $9)
		RETURNING id, workspace_id, loan_id, account_id, COALESCE(transaction_id::text, ''), principal_amount::text, interest_amount::text, fee_amount::text, waived_amount::text, occurred_at, created_at, updated_at
	`, input.WorkspaceID, input.LoanID, nilUUID(input.AccountID), input.TransactionID, input.Principal, input.Interest, input.Fee, input.Waived, input.OccurredAt.UTC()).Scan(
		&out.ID, &out.WorkspaceID, &out.LoanID, &out.AccountID, &out.TransactionID, &out.Principal, &out.Interest, &out.Fee, &out.Waived,
		&out.OccurredAt, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.LoanPayment{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListLoanPayments(workspaceID domain.ID, loanID domain.ID) []domain.LoanPayment {
	query := `
		SELECT id, workspace_id, loan_id, account_id, COALESCE(transaction_id::text, ''), principal_amount::text, interest_amount::text, fee_amount::text, waived_amount::text, occurred_at, created_at, updated_at
		FROM loan_payments
		WHERE workspace_id=$1`
	args := []any{workspaceID}
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
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.LoanID, &p.AccountID, &p.TransactionID, &p.Principal, &p.Interest, &p.Fee, &p.Waived, &p.OccurredAt, &p.CreatedAt, &p.UpdatedAt); err == nil {
			out = append(out, p)
		}
	}
	return out
}

func (s *PostgresStore) GetProperty(id domain.ID) (*domain.Property, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, workspace_id, portfolio_id, name, address, area_m2, purchase_at, created_at, updated_at
		FROM properties WHERE id=$1`, id)
	var p domain.Property
	if err := row.Scan(&p.ID, &p.WorkspaceID, &p.PortfolioID, &p.Name, &p.Address, &p.AreaM2, &p.PurchaseAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, false
	}
	return &p, true
}

func (s *PostgresStore) CreateProperty(input domain.Property) (domain.Property, error) {
	ctx := context.Background()
	var out domain.Property
	err := s.pool.QueryRow(ctx, `
		INSERT INTO properties(workspace_id, portfolio_id, name, address, area_m2, purchase_at)
		VALUES($1, $2, $3, $4, $5, $6)
		RETURNING id, workspace_id, portfolio_id, name, address, area_m2, purchase_at, created_at, updated_at
	`, input.WorkspaceID, nilUUID(input.PortfolioID), input.Name, input.Address, input.AreaM2, input.PurchaseAt).Scan(
		&out.ID, &out.WorkspaceID, &out.PortfolioID, &out.Name, &out.Address, &out.AreaM2, &out.PurchaseAt, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.Property{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListProperties(workspaceID domain.ID) []domain.Property {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, portfolio_id, name, address, area_m2, purchase_at, created_at, updated_at
		FROM properties WHERE workspace_id=$1 ORDER BY name ASC`, workspaceID)
	if err != nil {
		return []domain.Property{}
	}
	defer rows.Close()
	out := make([]domain.Property, 0)
	for rows.Next() {
		var p domain.Property
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.PortfolioID, &p.Name, &p.Address, &p.AreaM2, &p.PurchaseAt, &p.CreatedAt, &p.UpdatedAt); err == nil {
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

func (s *PostgresStore) ListPropertyValues(workspaceID domain.ID) []domain.PropertyValuation {
	rows, err := s.pool.Query(context.Background(), `
		SELECT pv.id, pv.property_id, pv.amount::text, pv.currency, pv.source, pv.effective_at, pv.is_stale, pv.created_at, pv.updated_at
		FROM property_valuations pv
		JOIN properties p ON p.id = pv.property_id
		WHERE p.workspace_id=$1
		ORDER BY pv.effective_at DESC`, workspaceID)
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
		SELECT id, workspace_id, portfolio_id, name, asset_type, created_at, updated_at
		FROM assets WHERE id=$1`, id)
	var a domain.Asset
	if err := row.Scan(&a.ID, &a.WorkspaceID, &a.PortfolioID, &a.Name, &a.Type, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, false
	}
	return &a, true
}

func (s *PostgresStore) CreateAsset(input domain.Asset) (domain.Asset, error) {
	ctx := context.Background()
	var out domain.Asset
	err := s.pool.QueryRow(ctx, `
		INSERT INTO assets(workspace_id, portfolio_id, name, asset_type)
		VALUES($1, $2, $3, $4)
		RETURNING id, workspace_id, portfolio_id, name, asset_type, created_at, updated_at
	`, input.WorkspaceID, nilUUID(input.PortfolioID), input.Name, input.Type).Scan(&out.ID, &out.WorkspaceID, &out.PortfolioID, &out.Name, &out.Type, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func (s *PostgresStore) ListAssets(workspaceID domain.ID) []domain.Asset {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, portfolio_id, name, asset_type, created_at, updated_at
		FROM assets WHERE workspace_id=$1 ORDER BY name ASC`, workspaceID)
	if err != nil {
		return []domain.Asset{}
	}
	defer rows.Close()
	out := make([]domain.Asset, 0)
	for rows.Next() {
		var a domain.Asset
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.PortfolioID, &a.Name, &a.Type, &a.CreatedAt, &a.UpdatedAt); err == nil {
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

func (s *PostgresStore) ListAssetValues(workspaceID domain.ID) []domain.AssetValuation {
	rows, err := s.pool.Query(context.Background(), `
		SELECT av.id, av.asset_id, av.amount::text, av.currency, av.source, av.effective_at, av.created_at, av.updated_at
		FROM asset_valuations av
		JOIN assets a ON a.id = av.asset_id
		WHERE a.workspace_id=$1
		ORDER BY av.effective_at DESC`, workspaceID)
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
		INSERT INTO budgets(workspace_id, period, category_id, limit_amount)
		VALUES($1, $2, NULLIF($3, ''), CAST($4 AS NUMERIC))
		RETURNING id, workspace_id, period, COALESCE(category_id::text, ''), limit_amount::text, created_at, updated_at
	`, input.WorkspaceID, input.Period, input.CategoryID, input.Limit).Scan(&out.ID, &out.WorkspaceID, &out.Period, &out.CategoryID, &out.Limit, &out.CreatedAt, &out.UpdatedAt)
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
			INSERT INTO budgets(workspace_id, period, category_id, limit_amount)
			VALUES($1, $2, $3, CAST($4 AS NUMERIC))
			ON CONFLICT (workspace_id, period, category_id) DO UPDATE
			SET limit_amount = CAST($4 AS NUMERIC), updated_at = now()
			RETURNING id, workspace_id, period, COALESCE(category_id::text, ''), limit_amount::text, created_at, updated_at
		`, input.WorkspaceID, input.Period, input.CategoryID, input.Limit).Scan(
			&out.ID, &out.WorkspaceID, &out.Period, &out.CategoryID, &out.Limit, &out.CreatedAt, &out.UpdatedAt,
		)
		if err == nil {
			return out, nil
		}
	}

	// category_id can be NULL; unique (workspace_id, period, category_id) treats NULL as distinct row,
	// so manual upsert is safer for null-category budgets.
	var existing domain.Budget
	err := s.pool.QueryRow(ctx, `
		SELECT id, workspace_id, period, COALESCE(category_id::text, ''), limit_amount::text, created_at, updated_at
		FROM budgets
		WHERE workspace_id=$1 AND period=$2 AND category_id IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, input.WorkspaceID, input.Period).Scan(&existing.ID, &existing.WorkspaceID, &existing.Period, &existing.CategoryID, &existing.Limit, &existing.CreatedAt, &existing.UpdatedAt)
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
		INSERT INTO budgets(workspace_id, period, category_id, limit_amount)
		VALUES($1, $2, NULLIF($3, ''), CAST($4 AS NUMERIC))
		RETURNING id, workspace_id, period, COALESCE(category_id::text, ''), limit_amount::text, created_at, updated_at
	`, input.WorkspaceID, input.Period, input.CategoryID, input.Limit).Scan(&out.ID, &out.WorkspaceID, &out.Period, &out.CategoryID, &out.Limit, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.Budget{}, err
	}
	return out, nil
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

func (s *PostgresStore) ListBudgets(workspaceID domain.ID, period string) []domain.Budget {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, period, COALESCE(category_id::text, ''), limit_amount::text, created_at, updated_at
		FROM budgets
		WHERE workspace_id=$1 AND period=$2
		ORDER BY COALESCE(category_id::text, '')
	`, workspaceID, period)
	if err != nil {
		return []domain.Budget{}
	}
	defer rows.Close()
	out := make([]domain.Budget, 0)
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.ID, &b.WorkspaceID, &b.Period, &b.CategoryID, &b.Limit, &b.CreatedAt, &b.UpdatedAt); err == nil {
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
		INSERT INTO forecast_scenarios(workspace_id, name, assumptions, status)
		VALUES($1, $2, $3, 'draft')
		RETURNING id, workspace_id, name, status, assumptions, result, created_at, updated_at
	`, input.WorkspaceID, input.Name, input.Assumptions).Scan(
		&out.ID, &out.WorkspaceID, &out.Name, &out.Status, &out.Assumptions, &out.Result, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) ListForecastScenarios(workspaceID domain.ID) []domain.ForecastScenario {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, name, status, assumptions, result, created_at, updated_at
		FROM forecast_scenarios WHERE workspace_id=$1 ORDER BY created_at DESC`, workspaceID)
	if err != nil {
		return []domain.ForecastScenario{}
	}
	defer rows.Close()
	out := make([]domain.ForecastScenario, 0)
	for rows.Next() {
		var f domain.ForecastScenario
		if err := rows.Scan(&f.ID, &f.WorkspaceID, &f.Name, &f.Status, &f.Assumptions, &f.Result, &f.CreatedAt, &f.UpdatedAt); err == nil {
			out = append(out, f)
		}
	}
	return out
}

func (s *PostgresStore) ListForecastScenariosByStatus(status string) []domain.ForecastScenario {
	status = strings.TrimSpace(status)
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, name, status, assumptions, result, created_at, updated_at
		FROM forecast_scenarios WHERE status=$1 ORDER BY updated_at DESC`, status)
	if err != nil {
		return []domain.ForecastScenario{}
	}
	defer rows.Close()
	out := make([]domain.ForecastScenario, 0)
	for rows.Next() {
		var f domain.ForecastScenario
		if err := rows.Scan(&f.ID, &f.WorkspaceID, &f.Name, &f.Status, &f.Assumptions, &f.Result, &f.CreatedAt, &f.UpdatedAt); err == nil {
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
		SELECT id, workspace_id, name, status, assumptions, result, created_at, updated_at
		FROM forecast_scenarios WHERE id=$1`, id).Scan(&f.ID, &f.WorkspaceID, &f.Name, &f.Status, &f.Assumptions, &f.Result, &f.CreatedAt, &f.UpdatedAt)
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
		RETURNING id, workspace_id, name, status, assumptions, result, created_at, updated_at
	`, id, status, strings.TrimSpace(result)).Scan(&f.ID, &f.WorkspaceID, &f.Name, &f.Status, &f.Assumptions, &f.Result, &f.CreatedAt, &f.UpdatedAt)
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
			workspace_id, provider, external_id, scope, status, bank_code,
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
		RETURNING id, workspace_id, provider, external_id, status, scope, COALESCE(bank_code, ''), callback_state, sync_status, last_synced_at, last_sync_requested_at, created_at, updated_at
	`, input.WorkspaceID, input.Provider, input.ExternalID, input.Scope, input.BankCode, input.CallbackState, input.SyncStatus).Scan(
		&out.ID, &out.WorkspaceID, &out.Provider, &out.ExternalID, &out.Status, &out.Scope, &out.BankCode,
		&out.CallbackState, &out.SyncStatus, &out.LastSyncedAt, &out.LastSyncRequestedAt, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) ListBankConnections(workspaceID domain.ID) []domain.BankConnection {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, provider, external_id, status, scope, COALESCE(bank_code, ''), callback_state,
		       sync_status, last_synced_at, last_sync_requested_at, created_at, updated_at
		FROM bank_connections WHERE workspace_id=$1 ORDER BY created_at DESC`, workspaceID)
	if err != nil {
		return []domain.BankConnection{}
	}
	defer rows.Close()
	out := make([]domain.BankConnection, 0)
	for rows.Next() {
		var c domain.BankConnection
		if err := rows.Scan(
			&c.ID, &c.WorkspaceID, &c.Provider, &c.ExternalID, &c.Status, &c.Scope, &c.BankCode, &c.CallbackState, &c.SyncStatus,
			&c.LastSyncedAt, &c.LastSyncRequestedAt, &c.CreatedAt, &c.UpdatedAt,
		); err == nil {
			out = append(out, c)
		}
	}
	return out
}

func (s *PostgresStore) GetBankConnection(id domain.ID) (*domain.BankConnection, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, workspace_id, provider, external_id, status, scope, COALESCE(bank_code, ''), callback_state,
		       sync_status, last_synced_at, last_sync_requested_at, created_at, updated_at
		FROM bank_connections WHERE id=$1`, id)
	var c domain.BankConnection
	if err := row.Scan(&c.ID, &c.WorkspaceID, &c.Provider, &c.ExternalID, &c.Status, &c.Scope, &c.BankCode,
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
		SELECT id, workspace_id, provider, external_id, status, scope, COALESCE(bank_code, ''), callback_state,
		       sync_status, last_synced_at, last_sync_requested_at, created_at, updated_at
		FROM bank_connections WHERE callback_state=$1`, state)
	var c domain.BankConnection
	if err := row.Scan(&c.ID, &c.WorkspaceID, &c.Provider, &c.ExternalID, &c.Status, &c.Scope, &c.BankCode, &c.CallbackState,
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
	if input.WorkspaceID == "" || input.ConnectionID == "" {
		return domain.BankFeedEvent{}, errors.New("workspaceId and connectionId are required")
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
			workspace_id, connection_id, provider, event_key, external_transaction_id, state, payload, attempts, last_error
		)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT(event_key) DO UPDATE
			SET updated_at=EXCLUDED.updated_at
		RETURNING id, workspace_id, connection_id, provider, event_key, external_transaction_id, state, payload::text,
			attempts, COALESCE(last_error, ''), created_at, updated_at
	`, input.WorkspaceID, input.ConnectionID, input.Provider, input.EventKey, input.ExternalID, input.State, []byte(input.Payload),
		input.Attempts, input.LastError).
		Scan(&out.ID, &out.WorkspaceID, &out.ConnectionID, &out.Provider, &out.EventKey, &out.ExternalID, &out.State,
			&out.Payload, &out.Attempts, &out.LastError, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.BankFeedEvent{}, err
	}
	return out, nil
}

func (s *PostgresStore) ListBankFeedEvents(workspaceID domain.ID, state string) []domain.BankFeedEvent {
	query := "SELECT id, workspace_id, connection_id, provider, event_key, external_transaction_id, state, payload::text, attempts, COALESCE(last_error, ''), created_at, updated_at FROM bank_feed_events"
	args := make([]any, 0, 2)
	predicates := make([]string, 0, 2)
	if workspaceID != "" {
		args = append(args, workspaceID)
		predicates = append(predicates, "workspace_id=$"+strconv.Itoa(len(args)))
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
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.ConnectionID, &e.Provider, &e.EventKey, &e.ExternalID, &e.State,
			&e.Payload, &e.Attempts, &e.LastError, &e.CreatedAt, &e.UpdatedAt); err == nil {
			out = append(out, e)
		}
	}
	return out
}

func (s *PostgresStore) GetBankFeedEvent(id domain.ID) (*domain.BankFeedEvent, bool) {
	var out domain.BankFeedEvent
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, workspace_id, connection_id, provider, event_key, external_transaction_id, state, payload::text, attempts, COALESCE(last_error, ''), created_at, updated_at
		FROM bank_feed_events WHERE id=$1`, id)
	if err := row.Scan(&out.ID, &out.WorkspaceID, &out.ConnectionID, &out.Provider, &out.EventKey, &out.ExternalID, &out.State,
		&out.Payload, &out.Attempts, &out.LastError, &out.CreatedAt, &out.UpdatedAt); err != nil {
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
			workspace_id, connection_id, account_id, amount, currency, direction, counterparty, description, occurred_at,
			external_transaction_id, reference, classification_confidence, classification_evidence, posting_state,
			posted_transaction_id, auto_classified, rule_id, source)
		VALUES($1,$2,$3,CAST($4 AS NUMERIC),$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		ON CONFLICT(workspace_id, connection_id, external_transaction_id) DO UPDATE
			SET updated_at=EXCLUDED.updated_at
			RETURNING id, workspace_id, connection_id, account_id,
				amount::text, currency, direction, counterparty, description, occurred_at,
				external_transaction_id, reference, classification_confidence, classification_evidence, posting_state,
				COALESCE(posted_transaction_id::text,''), auto_classified, COALESCE(rule_id::text,''),
				source, created_at, updated_at
	`, input.WorkspaceID, input.ConnectionID, input.AccountID, input.Amount, input.Currency, input.Direction, input.CounterParty,
		input.Description, input.OccurredAt.UTC(), input.ExternalID, input.Reference, input.Confidence, input.Evidence,
		input.PostingState, nilUUID(input.PostedTxnID), input.AutoClassified, nilUUID(input.RuleID), input.Source).Scan(
		&out.ID, &out.WorkspaceID, &out.ConnectionID, &out.AccountID, &out.Amount, &out.Currency, &out.Direction, &out.CounterParty,
		&out.Description, &out.OccurredAt, &out.ExternalID, &out.Reference, &out.Confidence, &out.Evidence, &out.PostingState,
		&out.PostedTxnID, &out.AutoClassified, &out.RuleID, &out.Source, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) ListBankFeed(workspaceID domain.ID) []domain.BankFeedTransaction {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, connection_id, account_id, amount::text, currency, direction, counterparty, description,
		       occurred_at, external_transaction_id, reference, classification_confidence, classification_evidence, posting_state,
		       COALESCE(posted_transaction_id::text,''), auto_classified, COALESCE(rule_id::text,''), source, created_at, updated_at
		FROM bank_feed_transactions
		WHERE workspace_id=$1
		ORDER BY occurred_at DESC`, workspaceID)
	if err != nil {
		return []domain.BankFeedTransaction{}
	}
	defer rows.Close()
	out := make([]domain.BankFeedTransaction, 0)
	for rows.Next() {
		var f domain.BankFeedTransaction
		if err := rows.Scan(&f.ID, &f.WorkspaceID, &f.ConnectionID, &f.AccountID, &f.Amount, &f.Currency, &f.Direction,
			&f.CounterParty, &f.Description, &f.OccurredAt, &f.ExternalID, &f.Reference, &f.Confidence, &f.Evidence, &f.PostingState,
			&f.PostedTxnID, &f.AutoClassified, &f.RuleID, &f.Source, &f.CreatedAt, &f.UpdatedAt); err == nil {
			out = append(out, f)
		}
	}
	return out
}

func (s *PostgresStore) ListBankFeedByState(workspaceID domain.ID, state domain.TransactionPostingState) []domain.BankFeedTransaction {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, connection_id, account_id, amount::text, currency, direction, counterparty, description,
		       occurred_at, external_transaction_id, reference, classification_confidence, classification_evidence, posting_state,
		       COALESCE(posted_transaction_id::text,''), auto_classified, COALESCE(rule_id::text,''), source, created_at, updated_at
		FROM bank_feed_transactions
		WHERE workspace_id=$1 AND posting_state=$2
		ORDER BY occurred_at DESC`, workspaceID, state)
	if err != nil {
		return []domain.BankFeedTransaction{}
	}
	defer rows.Close()
	out := make([]domain.BankFeedTransaction, 0)
	for rows.Next() {
		var f domain.BankFeedTransaction
		if err := rows.Scan(&f.ID, &f.WorkspaceID, &f.ConnectionID, &f.AccountID, &f.Amount, &f.Currency, &f.Direction,
			&f.CounterParty, &f.Description, &f.OccurredAt, &f.ExternalID, &f.Reference, &f.Confidence, &f.Evidence, &f.PostingState,
			&f.PostedTxnID, &f.AutoClassified, &f.RuleID, &f.Source, &f.CreatedAt, &f.UpdatedAt); err == nil {
			out = append(out, f)
		}
	}
	return out
}

func (s *PostgresStore) GetBankFeed(id domain.ID) (*domain.BankFeedTransaction, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, workspace_id, connection_id, account_id, amount::text, currency, direction, counterparty, description,
		       occurred_at, external_transaction_id, reference, classification_confidence, classification_evidence, posting_state,
		       COALESCE(posted_transaction_id::text,''), auto_classified, COALESCE(rule_id::text,''), source, created_at, updated_at
		FROM bank_feed_transactions WHERE id=$1`, id)
	var f domain.BankFeedTransaction
	if err := row.Scan(&f.ID, &f.WorkspaceID, &f.ConnectionID, &f.AccountID, &f.Amount, &f.Currency, &f.Direction, &f.CounterParty, &f.Description,
		&f.OccurredAt, &f.ExternalID, &f.Reference, &f.Confidence, &f.Evidence, &f.PostingState, &f.PostedTxnID, &f.AutoClassified, &f.RuleID, &f.Source, &f.CreatedAt, &f.UpdatedAt); err != nil {
		return nil, false
	}
	return &f, true
}

func (s *PostgresStore) UpdateFeedState(id domain.ID, state domain.TransactionPostingState, reason string) error {
	_, err := s.pool.Exec(context.Background(), `
		UPDATE bank_feed_transactions
		SET posting_state=$2, classification_evidence=$3, updated_at=now()
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
		SET amount=CAST($2 AS NUMERIC), currency=$3, direction=$4, counterparty=$5, description=$6, occurred_at=$7,
		    external_transaction_id=$8, reference=$9, classification_confidence=$10, classification_evidence=$11,
		    posting_state=$12, posted_transaction_id=CAST(NULLIF($13, '') AS UUID), auto_classified=$14, rule_id=CAST(NULLIF($15, '') AS UUID), source=$16, updated_at=now()
		WHERE id=$1
	`, id, mutated.Amount, mutated.Currency, mutated.Direction, mutated.CounterParty, mutated.Description, mutated.OccurredAt,
		mutated.ExternalID, mutated.Reference, mutated.Confidence, mutated.Evidence, mutated.PostingState, string(mutated.PostedTxnID),
		mutated.AutoClassified, string(mutated.RuleID), mutated.Source)
	return err == nil
}

func (s *PostgresStore) LinkBankFeedPosting(feedID domain.ID, txnID domain.ID) bool {
	_, err := s.pool.Exec(context.Background(), `
		UPDATE bank_feed_transactions
		SET posted_transaction_id=$2, updated_at=now()
		WHERE id=$1
	`, feedID, txnID)
	return err == nil
}

func (s *PostgresStore) CreateAutomationRule(input domain.AutomationRule) (domain.AutomationRule, error) {
	ctx := context.Background()
	var out domain.AutomationRule
	err := s.pool.QueryRow(ctx, `
		INSERT INTO bank_automation_rules(workspace_id, account_id, name, priority, predicate, direction, action_type, type, category_id, enabled, content_pattern, reference_pattern, min_amount, max_amount)
		VALUES($1, NULLIF($2,''), $3, $4, $5, $6, $7, $8, NULLIF($9,''), COALESCE($10, TRUE), $11, $12, NULLIF($13,''), NULLIF($14,''))
		RETURNING id, workspace_id, COALESCE(account_id::text, ''), name, priority, predicate, direction, action_type, type, COALESCE(category_id::text, ''),
		          enabled, content_pattern, reference_pattern, min_amount::text, max_amount::text, created_at, updated_at
	`, input.WorkspaceID, input.AccountID, input.Name, input.Priority, input.Predicate, input.Direction, input.ActionType, input.Type,
		input.CategoryID, input.Enabled, input.ContentPattern, input.ReferencePattern, input.MinAmount, input.MaxAmount).Scan(
		&out.ID, &out.WorkspaceID, &out.AccountID, &out.Name, &out.Priority, &out.Predicate, &out.Direction, &out.ActionType, &out.Type,
		&out.CategoryID, &out.Enabled, &out.ContentPattern, &out.ReferencePattern, &out.MinAmount, &out.MaxAmount, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) ListWorkspaceAutomationRules(workspaceID domain.ID) []domain.AutomationRule {
	return s.ListAutomationRules(workspaceID)
}

func (s *PostgresStore) ListWorkspaceRules(workspaceID domain.ID) []domain.AutomationRule {
	return s.ListAutomationRules(workspaceID)
}

func (s *PostgresStore) ListAutomationRules(workspaceID domain.ID) []domain.AutomationRule {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, COALESCE(account_id::text, ''), name, priority, predicate, direction, action_type, type,
		       COALESCE(category_id::text, ''), enabled, content_pattern, reference_pattern, min_amount::text, max_amount::text, created_at, updated_at
		FROM bank_automation_rules
		WHERE workspace_id=$1
		ORDER BY priority ASC, created_at ASC`, workspaceID)
	if err != nil {
		return []domain.AutomationRule{}
	}
	defer rows.Close()
	out := make([]domain.AutomationRule, 0)
	for rows.Next() {
		var r domain.AutomationRule
		if err := rows.Scan(&r.ID, &r.WorkspaceID, &r.AccountID, &r.Name, &r.Priority, &r.Predicate, &r.Direction,
			&r.ActionType, &r.Type, &r.CategoryID, &r.Enabled, &r.ContentPattern, &r.ReferencePattern, &r.MinAmount, &r.MaxAmount, &r.CreatedAt, &r.UpdatedAt); err == nil {
			out = append(out, r)
		}
	}
	return out
}

func (s *PostgresStore) GetWorkspaceRules(workspaceID domain.ID, accountID domain.ID, direction string) []domain.AutomationRule {
	query := `
		SELECT id, workspace_id, COALESCE(account_id::text, ''), name, priority, predicate, direction, action_type, type,
		       COALESCE(category_id::text, ''), enabled, content_pattern, reference_pattern, min_amount::text, max_amount::text, created_at, updated_at
		FROM bank_automation_rules
		WHERE workspace_id=$1 AND enabled = true`
	args := []any{workspaceID}
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
		if err := rows.Scan(&r.ID, &r.WorkspaceID, &r.AccountID, &r.Name, &r.Priority, &r.Predicate, &r.Direction,
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
		SELECT id, workspace_id, COALESCE(account_id::text, ''), name, priority, predicate, direction, action_type, type,
		       COALESCE(category_id::text, ''), enabled, content_pattern, reference_pattern, min_amount::text, max_amount::text, created_at, updated_at
		FROM bank_automation_rules WHERE id=$1`, id)
	var r domain.AutomationRule
	if err := row.Scan(&r.ID, &r.WorkspaceID, &r.AccountID, &r.Name, &r.Priority, &r.Predicate, &r.Direction,
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
		SET name=$2, predicate=$3, priority=$4, direction=$5, action_type=$6, type=$7, category_id=NULLIF($8, ''), enabled=$9,
		    content_pattern=$10, reference_pattern=$11, min_amount=NULLIF($12, ''), max_amount=NULLIF($13, ''), updated_at=now()
		WHERE id=$1
	`, id, mutated.Name, mutated.Predicate, mutated.Priority, mutated.Direction, mutated.ActionType, mutated.Type, mutated.CategoryID, mutated.Enabled,
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
		INSERT INTO bank_payment_requests(workspace_id, loan_id, payment_code, amount, currency, expires_at, status, note, source)
		VALUES($1,$2,$3,CAST($4 AS NUMERIC),$5,$6,$7,$8,$9)
		RETURNING id, workspace_id, loan_id, payment_code, amount::text, currency, expires_at, status, note, source, created_at, updated_at
	`, input.WorkspaceID, input.LoanID, input.Code, input.Amount, input.Currency, input.ExpiresAt, input.Status, input.Note, input.Source).Scan(
		&out.ID, &out.WorkspaceID, &out.LoanID, &out.Code, &out.Amount, &out.Currency, &out.ExpiresAt, &out.Status, &out.Note, &out.Source, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) GetBankPaymentRequestByCode(workspaceID domain.ID, code string) (*domain.BankPaymentRequest, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, workspace_id, loan_id, payment_code, amount::text, currency, expires_at, status, note, source, created_at, updated_at
		FROM bank_payment_requests
		WHERE workspace_id=$1 AND payment_code=$2
	`, workspaceID, code)
	var r domain.BankPaymentRequest
	if err := row.Scan(&r.ID, &r.WorkspaceID, &r.LoanID, &r.Code, &r.Amount, &r.Currency, &r.ExpiresAt, &r.Status, &r.Note, &r.Source, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return nil, false
	}
	return &r, true
}

func (s *PostgresStore) ListBankPaymentRequests(workspaceID domain.ID) []domain.BankPaymentRequest {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, loan_id, payment_code, amount::text, currency, expires_at, status, note, source, created_at, updated_at
		FROM bank_payment_requests WHERE workspace_id=$1
		ORDER BY created_at DESC`, workspaceID)
	if err != nil {
		return []domain.BankPaymentRequest{}
	}
	defer rows.Close()
	out := make([]domain.BankPaymentRequest, 0)
	for rows.Next() {
		var r domain.BankPaymentRequest
		if err := rows.Scan(&r.ID, &r.WorkspaceID, &r.LoanID, &r.Code, &r.Amount, &r.Currency, &r.ExpiresAt, &r.Status, &r.Note, &r.Source, &r.CreatedAt, &r.UpdatedAt); err == nil {
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
		SELECT id, workspace_id, provider, external_id, status, scope, COALESCE(bank_code, ''), callback_state,
		       sync_status, last_synced_at, last_sync_requested_at, created_at, updated_at
		FROM bank_connections WHERE id=$1`, id)
	if err := row.Scan(&out.ID, &out.WorkspaceID, &out.Provider, &out.ExternalID, &out.Status, &out.Scope, &out.BankCode,
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
		INSERT INTO assistant_commands(workspace_id, user_id, command, status, plan, approval_id, approval_expires_at)
		VALUES($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, workspace_id, user_id, command, status, plan, approval_id, approval_expires_at, approval_used_at, created_at, updated_at
	`, input.WorkspaceID, input.UserID, input.Command, input.Status, input.Plan, input.ApprovalID, nullTime(input.ApprovalExpiresAt)).Scan(
		&out.ID, &out.WorkspaceID, &out.UserID, &out.Command, &out.Status, &out.Plan,
		&out.ApprovalID, &out.ApprovalExpiresAt, &out.ApprovalUsedAt, &out.CreatedAt, &out.UpdatedAt,
	)
	return out, err
}

func (s *PostgresStore) GetAssistantCommand(id domain.ID) (*domain.AssistantCommand, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, workspace_id, user_id, command, status, plan, approval_id, approval_expires_at, approval_used_at, created_at, updated_at
		FROM assistant_commands WHERE id=$1`, id)
	var c domain.AssistantCommand
	if err := row.Scan(&c.ID, &c.WorkspaceID, &c.UserID, &c.Command, &c.Status, &c.Plan,
		&c.ApprovalID, &c.ApprovalExpiresAt, &c.ApprovalUsedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, false
	}
	return &c, true
}

func (s *PostgresStore) ListAssistantCommands(workspaceID domain.ID) []domain.AssistantCommand {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, workspace_id, user_id, command, status, plan, approval_id, approval_expires_at, approval_used_at, created_at, updated_at
		FROM assistant_commands WHERE workspace_id=$1 ORDER BY created_at DESC`, workspaceID)
	if err != nil {
		return []domain.AssistantCommand{}
	}
	defer rows.Close()
	out := make([]domain.AssistantCommand, 0)
	for rows.Next() {
		var c domain.AssistantCommand
		if err := rows.Scan(&c.ID, &c.WorkspaceID, &c.UserID, &c.Command, &c.Status, &c.Plan, &c.ApprovalID, &c.ApprovalExpiresAt, &c.ApprovalUsedAt, &c.CreatedAt, &c.UpdatedAt); err == nil {
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
		SET workspace_id=$2, user_id=$3, command=$4, status=$5, plan=$6,
			approval_id=$7, approval_expires_at=$8, approval_used_at=$9, updated_at=now()
		WHERE id=$1
	`, id, cp.WorkspaceID, cp.UserID, cp.Command, cp.Status, cp.Plan,
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
	if id == "" {
		return nil
	}
	return id
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

func init() {
	// keep compatibility for tests that may call NewPostgresStore without explicit init in this project
	_ = uuid.NewString
}

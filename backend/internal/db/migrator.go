package db

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MigrationRunner struct {
	Dir string
	DB  *pgxpool.Pool
}

func NewMigrationRunner(db *pgxpool.Pool, dir string) *MigrationRunner {
	return &MigrationRunner{
		Dir: dir,
		DB:  db,
	}
}

func (r *MigrationRunner) Run(ctx context.Context) error {
	entries, err := os.ReadDir(r.Dir)
	if err != nil {
		return fmt.Errorf("read migration dir: %w", err)
	}

	files := make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	if len(files) == 0 {
		return nil
	}

	if _, err := r.DB.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, executed_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return fmt.Errorf("ensure migration table: %w", err)
	}

	rows, err := r.DB.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("load migration history: %w", err)
	}
	defer rows.Close()
	applied := map[string]struct{}{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan migration history: %w", err)
	}

	for _, f := range files {
		version := strings.TrimSuffix(f.Name(), ".sql")
		if _, ok := applied[version]; ok {
			continue
		}

		contentPath := filepath.Join(r.Dir, f.Name())
		raw, err := os.ReadFile(contentPath)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f.Name(), err)
		}
		// SQL editors on Windows commonly prepend a UTF-8 BOM. PostgreSQL does
		// not treat it as whitespace, so remove it before executing migrations.
		stmt := strings.TrimSpace(strings.TrimPrefix(string(raw), "\uFEFF"))
		if stmt == "" {
			continue
		}

		tx, err := r.DB.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, stmt); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES($1)", version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
	}
	return nil
}

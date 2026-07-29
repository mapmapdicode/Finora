ALTER TABLE loans
    ADD COLUMN IF NOT EXISTS daily_rate_per_million NUMERIC(24,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS settlement_account_id UUID REFERENCES accounts(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_loans_settlement_account ON loans (settlement_account_id);

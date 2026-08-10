-- A reconciled real-world balance can differ from imported ledger history.
-- This is deliberately account metadata, not a transaction, so correcting a
-- current bank/cash balance does not manufacture a historical cash-flow row.
ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS balance_override NUMERIC(24,4),
    ADD COLUMN IF NOT EXISTS balance_override_at TIMESTAMPTZ;

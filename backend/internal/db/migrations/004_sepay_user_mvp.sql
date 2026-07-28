-- User-first Bank Hub ownership. Existing user-scoped bank tables remain
-- available for legacy clients while mobile uses these tables exclusively.
CREATE TABLE IF NOT EXISTS sepay_user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    company_xid TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    linked_at TIMESTAMPTZ,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sepay_bank_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bank_account_xid TEXT NOT NULL UNIQUE,
    bank_code TEXT,
    bank_name TEXT,
    account_number_masked TEXT NOT NULL DEFAULT '',
    supports_in BOOLEAN NOT NULL DEFAULT FALSE,
    supports_out BOOLEAN NOT NULL DEFAULT FALSE,
    status TEXT NOT NULL DEFAULT 'linked',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, bank_account_xid)
);

CREATE TABLE IF NOT EXISTS sepay_link_sessions (
    xid TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- IPNs can arrive before the corresponding BANK_ACCOUNT_LINKED event. Keep
-- them durable and idempotent for operator reconciliation instead of forcing
-- the provider into an endless retry loop.
CREATE TABLE IF NOT EXISTS sepay_unmapped_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL DEFAULT 'sepay',
    bank_account_xid TEXT NOT NULL,
    transaction_id TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'quarantined',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(provider, bank_account_xid, transaction_id)
);

CREATE TABLE IF NOT EXISTS bank_account_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sepay_bank_account_id UUID NOT NULL UNIQUE REFERENCES sepay_bank_accounts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(sepay_bank_account_id, account_id)
);

ALTER TABLE bank_feed_transactions
    ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS sepay_bank_account_id UUID REFERENCES sepay_bank_accounts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS raw_provider_data JSONB,
    ADD COLUMN IF NOT EXISTS classification_status TEXT NOT NULL DEFAULT 'needs_review';

CREATE INDEX IF NOT EXISTS idx_sepay_bank_accounts_user ON sepay_bank_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_bank_account_mappings_user ON bank_account_mappings(user_id, user_id);
CREATE INDEX IF NOT EXISTS idx_bank_feed_user_review ON bank_feed_transactions(user_id, classification_status, occurred_at DESC);

CREATE TABLE IF NOT EXISTS transaction_suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_feed_transaction_id UUID NOT NULL REFERENCES bank_feed_transactions(id) ON DELETE CASCADE,
    suggested_name TEXT,
    suggested_category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    source TEXT NOT NULL CHECK(source IN ('rule', 'history', 'ai')),
    confidence NUMERIC(5,2) NOT NULL DEFAULT 0,
    reason TEXT NOT NULL DEFAULT '',
    version TEXT NOT NULL DEFAULT 'v1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS classification_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_feed_transaction_id UUID NOT NULL REFERENCES bank_feed_transactions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action TEXT NOT NULL CHECK(action IN ('confirmed', 'corrected', 'ignored')),
    name TEXT,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    transaction_type TEXT,
    note TEXT,
    remember_choice BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_classification_feedback_user ON classification_feedback(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS markdown_import_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    external_code TEXT NOT NULL,
    entity_type TEXT NOT NULL CHECK (entity_type IN ('account', 'transaction', 'loan', 'payment')),
    entity_id UUID NOT NULL,
    import_month TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT markdown_import_references_user_type_code_unique UNIQUE (user_id, entity_type, external_code)
);

CREATE INDEX IF NOT EXISTS idx_markdown_import_references_user_month
    ON markdown_import_references (user_id, import_month);

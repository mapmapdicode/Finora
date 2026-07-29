CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    normalized_name TEXT NOT NULL,
    phone TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT customers_user_normalized_name_unique UNIQUE (user_id, normalized_name)
);

CREATE INDEX IF NOT EXISTS idx_customers_user_updated
    ON customers (user_id, updated_at DESC);

ALTER TABLE loans
    ADD COLUMN IF NOT EXISTS customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_loans_customer ON loans (customer_id);

-- Keep historical loans intact while giving each old counterparty a reusable
-- customer record. Counterparty remains as a backwards-compatible snapshot.
INSERT INTO customers (user_id, name, normalized_name)
SELECT DISTINCT user_id, btrim(counterparty), lower(regexp_replace(btrim(counterparty), '\s+', ' ', 'g'))
FROM loans
WHERE customer_id IS NULL AND counterparty IS NOT NULL AND btrim(counterparty) <> ''
ON CONFLICT (user_id, normalized_name) DO NOTHING;

UPDATE loans AS loan
SET customer_id = customer.id
FROM customers AS customer
WHERE loan.customer_id IS NULL
  AND loan.user_id = customer.user_id
  AND lower(regexp_replace(btrim(COALESCE(loan.counterparty, '')), '\s+', ' ', 'g')) = customer.normalized_name;

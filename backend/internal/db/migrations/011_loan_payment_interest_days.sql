-- Preserve the selected interest period with each receipt so history remains
-- auditable even after future accrual periods are generated.
ALTER TABLE loan_payments
    ADD COLUMN IF NOT EXISTS interest_days INTEGER NOT NULL DEFAULT 0;

ALTER TABLE loan_payments
    ADD CONSTRAINT loan_payments_interest_days_nonnegative
    CHECK (interest_days >= 0) NOT VALID;

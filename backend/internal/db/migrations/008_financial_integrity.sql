-- Financial facts are append-only. These guards complement application-level
-- validation so direct SQL or a future write path cannot silently corrupt the ledger.

ALTER TABLE accounts
    ADD CONSTRAINT accounts_id_user_unique UNIQUE (id, user_id);
ALTER TABLE categories
    ADD CONSTRAINT categories_id_user_unique UNIQUE (id, user_id);
ALTER TABLE portfolios
    ADD CONSTRAINT portfolios_id_user_unique UNIQUE (id, user_id);

ALTER TABLE transactions
    ADD CONSTRAINT transactions_amount_positive CHECK (amount > 0) NOT VALID,
    ADD CONSTRAINT transactions_type_valid CHECK (type IN ('income', 'expense', 'transfer', 'investment_funding', 'loan_disbursement', 'loan_payment', 'valuation_adjustment')) NOT VALID,
    ADD CONSTRAINT transactions_status_valid CHECK (status IN ('draft', 'pending', 'posted', 'reconciled', 'voided')) NOT VALID,
    ADD CONSTRAINT transactions_account_owner_fk FOREIGN KEY (account_id, user_id) REFERENCES accounts (id, user_id) NOT VALID,
    ADD CONSTRAINT transactions_category_owner_fk FOREIGN KEY (category_id, user_id) REFERENCES categories (id, user_id) NOT VALID,
    ADD CONSTRAINT transactions_portfolio_owner_fk FOREIGN KEY (portfolio_id, user_id) REFERENCES portfolios (id, user_id) NOT VALID;

ALTER TABLE transfers
    ADD CONSTRAINT transfers_amount_positive CHECK (amount > 0) NOT VALID,
    ADD CONSTRAINT transfers_distinct_accounts CHECK (from_account_id <> to_account_id) NOT VALID,
    ADD CONSTRAINT transfers_from_account_owner_fk FOREIGN KEY (from_account_id, user_id) REFERENCES accounts (id, user_id) NOT VALID,
    ADD CONSTRAINT transfers_to_account_owner_fk FOREIGN KEY (to_account_id, user_id) REFERENCES accounts (id, user_id) NOT VALID;

ALTER TABLE loans
    ADD CONSTRAINT loans_principal_positive CHECK (principal_initial > 0 AND principal_balance >= 0) NOT VALID,
    ADD CONSTRAINT loans_annual_rate_nonnegative CHECK (annual_rate >= 0) NOT VALID,
    ADD CONSTRAINT loans_direction_valid CHECK (direction IN ('receivable', 'payable')) NOT VALID,
    ADD CONSTRAINT loans_status_valid CHECK (status IN ('draft', 'active', 'overdue', 'closed', 'restructured', 'written_off', 'cancelled')) NOT VALID,
    ADD CONSTRAINT loans_portfolio_owner_fk FOREIGN KEY (portfolio_id, user_id) REFERENCES portfolios (id, user_id) NOT VALID;

ALTER TABLE loan_payments
    ADD CONSTRAINT loan_payments_amounts_nonnegative CHECK (principal_amount >= 0 AND interest_amount >= 0 AND fee_amount >= 0 AND waived_amount >= 0) NOT VALID,
    ADD CONSTRAINT loan_payments_account_owner_fk FOREIGN KEY (account_id, user_id) REFERENCES accounts (id, user_id) NOT VALID;

CREATE INDEX IF NOT EXISTS idx_transactions_user_occurred_at ON transactions (user_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_account_occurred_at ON transactions (account_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_loan_payments_loan_occurred_at ON loan_payments (loan_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_property_valuations_property_effective_at ON property_valuations (property_id, effective_at DESC);
CREATE INDEX IF NOT EXISTS idx_asset_valuations_asset_effective_at ON asset_valuations (asset_id, effective_at DESC);

CREATE OR REPLACE FUNCTION prevent_financial_history_delete() RETURNS trigger AS $$
BEGIN
    IF TG_TABLE_NAME = 'accounts' AND EXISTS (SELECT 1 FROM transactions WHERE account_id = OLD.id) THEN
        RAISE EXCEPTION 'cannot delete account with transaction history' USING ERRCODE = '23503';
    ELSIF TG_TABLE_NAME = 'loans' AND EXISTS (SELECT 1 FROM loan_payments WHERE loan_id = OLD.id) THEN
        RAISE EXCEPTION 'cannot delete loan with payment history' USING ERRCODE = '23503';
    ELSIF TG_TABLE_NAME = 'properties' AND EXISTS (SELECT 1 FROM property_valuations WHERE property_id = OLD.id) THEN
        RAISE EXCEPTION 'cannot delete property with valuation history' USING ERRCODE = '23503';
    ELSIF TG_TABLE_NAME = 'assets' AND EXISTS (SELECT 1 FROM asset_valuations WHERE asset_id = OLD.id) THEN
        RAISE EXCEPTION 'cannot delete asset with valuation history' USING ERRCODE = '23503';
    ELSIF TG_TABLE_NAME = 'portfolios' AND EXISTS (
        SELECT 1 FROM accounts WHERE portfolio_id = OLD.id
        UNION ALL SELECT 1 FROM loans WHERE portfolio_id = OLD.id
        UNION ALL SELECT 1 FROM properties WHERE portfolio_id = OLD.id
        UNION ALL SELECT 1 FROM assets WHERE portfolio_id = OLD.id
        UNION ALL SELECT 1 FROM transactions WHERE portfolio_id = OLD.id
    ) THEN
        RAISE EXCEPTION 'cannot delete portfolio with financial history' USING ERRCODE = '23503';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_account_financial_history_delete BEFORE DELETE ON accounts FOR EACH ROW EXECUTE FUNCTION prevent_financial_history_delete();
CREATE TRIGGER prevent_loan_payment_history_delete BEFORE DELETE ON loans FOR EACH ROW EXECUTE FUNCTION prevent_financial_history_delete();
CREATE TRIGGER prevent_property_valuation_history_delete BEFORE DELETE ON properties FOR EACH ROW EXECUTE FUNCTION prevent_financial_history_delete();
CREATE TRIGGER prevent_asset_valuation_history_delete BEFORE DELETE ON assets FOR EACH ROW EXECUTE FUNCTION prevent_financial_history_delete();
CREATE TRIGGER prevent_portfolio_financial_history_delete BEFORE DELETE ON portfolios FOR EACH ROW EXECUTE FUNCTION prevent_financial_history_delete();

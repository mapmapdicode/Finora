-- ============================================================================
-- FINORA / WEALTHOS DATABASE SEED SCRIPT
-- User: thanhoangz
-- Password: HoangThanZ6^
-- Workspace: thanhoangz
-- Account: Bank (Initial Balance: 180,000,000 VND)
-- ============================================================================

BEGIN;

-- 1. Insert User: thanhoangz
INSERT INTO users (id, email, name, password_hash, created_at, updated_at)
VALUES (
    'b47a19e2-8f12-4c99-b1d5-91a76ef4821c',
    'thanhoangz',
    'thanhoangz',
    'HoangThanZ6^',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO UPDATE 
SET password_hash = EXCLUDED.password_hash, updated_at = NOW();

-- 2. Insert User Settings
INSERT INTO user_settings (user_id, amount_display_mode, created_at, updated_at)
VALUES (
    'b47a19e2-8f12-4c99-b1d5-91a76ef4821c',
    'full',
    NOW(),
    NOW()
)
ON CONFLICT (user_id) DO UPDATE 
SET amount_display_mode = 'full', updated_at = NOW();

-- 3. Insert Workspace: thanhoangz
INSERT INTO workspaces (id, name, base_currency, fiscal_year_end, created_at, updated_at)
VALUES (
    'e8190c42-7a1b-4f9e-a832-159c3d421890',
    'thanhoangz',
    'VND',
    '2026-12-31',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE 
SET name = EXCLUDED.name, updated_at = NOW();

-- 4. Insert Workspace Member (Owner)
INSERT INTO workspace_members (id, workspace_id, user_id, role, created_at, updated_at)
VALUES (
    'f92a10b3-8c2d-4e11-9a43-260b4e532901',
    'e8190c42-7a1b-4f9e-a832-159c3d421890',
    'b47a19e2-8f12-4c99-b1d5-91a76ef4821c',
    'owner',
    NOW(),
    NOW()
)
ON CONFLICT (workspace_id, user_id) DO NOTHING;

-- 5. Insert Portfolio: Default
INSERT INTO portfolios (id, workspace_id, name, base_currency, created_at, updated_at)
VALUES (
    'c38210a4-9b3e-4f12-8d54-371c5e643012',
    'e8190c42-7a1b-4f9e-a832-159c3d421890',
    'Default',
    'VND',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- 6. Insert Account: Bank
INSERT INTO accounts (id, workspace_id, portfolio_id, name, type, currency, created_at, updated_at)
VALUES (
    'a1928374-b567-4c89-9d01-23456789abcd',
    'e8190c42-7a1b-4f9e-a832-159c3d421890',
    'c38210a4-9b3e-4f12-8d54-371c5e643012',
    'Bank',
    'bank',
    'VND',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- 7. Insert Category: Số dư ban đầu
INSERT INTO categories (id, workspace_id, name, kind, created_at, updated_at)
VALUES (
    'd4719283-a1b2-4c3d-8e5f-678901234567',
    'e8190c42-7a1b-4f9e-a832-159c3d421890',
    'Số dư ban đầu',
    'income',
    NOW(),
    NOW()
)
ON CONFLICT (workspace_id, name, kind) DO NOTHING;

-- 8. Insert Transaction: Initial Deposit of 180,000,000 VND into Bank Account
INSERT INTO transactions (
    id, workspace_id, account_id, category_id, portfolio_id,
    type, amount, currency, note, occurred_at, status, source, created_at, updated_at
)
VALUES (
    't9876543-21ab-4ccd-9e8f-1234567890ab',
    'e8190c42-7a1b-4f9e-a832-159c3d421890',
    'a1928374-b567-4c89-9d01-23456789abcd',
    'd4719283-a1b2-4c3d-8e5f-678901234567',
    'c38210a4-9b3e-4f12-8d54-371c5e643012',
    'income',
    180000000.0000,
    'VND',
    'Khoản nạp số dư ban đầu cho tài khoản ngân hàng Bank',
    NOW(),
    'posted',
    'initial_seed',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

COMMIT;

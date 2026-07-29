-- Earlier development databases used a workspace-owned schema while the
-- current application and migrations are user-owned. Move existing data to
-- the owning workspace member before later migrations add user-scoped rules.
-- Fresh databases already have user_id and this migration becomes a no-op.
DO $$
DECLARE
    target_table TEXT;
    target_tables TEXT[] := ARRAY[
        'accounts', 'assets', 'assistant_commands', 'audit_logs',
        'bank_account_mappings', 'bank_automation_rules', 'bank_connections',
        'bank_feed_events', 'bank_feed_transactions', 'bank_payment_requests',
        'budgets', 'categories', 'forecast_scenarios', 'loan_payments', 'loans',
        'portfolios', 'properties', 'transactions', 'transfers'
    ];
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns c
        WHERE c.table_schema = 'public' AND c.table_name = 'accounts' AND c.column_name = 'workspace_id'
    ) THEN
        RETURN;
    END IF;

    CREATE TEMP TABLE workspace_owner_map ON COMMIT DROP AS
    SELECT workspace_id, COALESCE(
        MIN(user_id::text) FILTER (WHERE role = 'owner'),
        MIN(user_id::text)
    )::uuid AS user_id
    FROM workspace_members
    GROUP BY workspace_id;

    IF EXISTS (
        SELECT 1 FROM accounts a
        LEFT JOIN workspace_owner_map m ON m.workspace_id = a.workspace_id
        WHERE m.user_id IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot migrate workspace-owned data: a workspace has no member';
    END IF;

    FOREACH target_table IN ARRAY target_tables LOOP
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns c
            WHERE c.table_schema = 'public' AND c.table_name = target_table AND c.column_name = 'workspace_id'
        ) THEN
            CONTINUE;
        END IF;

        EXECUTE format('ALTER TABLE %I DROP CONSTRAINT IF EXISTS %I', target_table, target_table || '_workspace_id_fkey');
        EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS user_id UUID', target_table);
        EXECUTE format(
            'UPDATE %I t SET user_id = m.user_id FROM workspace_owner_map m WHERE t.workspace_id = m.workspace_id AND t.user_id IS NULL',
            target_table
        );
        EXECUTE format('ALTER TABLE %I ALTER COLUMN user_id SET NOT NULL', target_table);
        EXECUTE format('ALTER TABLE %I DROP COLUMN workspace_id', target_table);
        IF NOT EXISTS (
            SELECT 1 FROM pg_constraint
            WHERE conrelid = format('public.%I', target_table)::regclass
              AND conname = target_table || '_user_id_fkey'
        ) THEN
            EXECUTE format(
                'ALTER TABLE %I ADD CONSTRAINT %I FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE',
                target_table, target_table || '_user_id_fkey'
            );
        END IF;
    END LOOP;
END;
$$;

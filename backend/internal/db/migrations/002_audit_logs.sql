CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_role TEXT,
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id UUID,
    request_id TEXT,
    path TEXT NOT NULL,
    method TEXT NOT NULL,
    policy_decision TEXT NOT NULL DEFAULT 'allowed',
    result TEXT NOT NULL,
    reason TEXT,
    correlation_id TEXT,
    before_json JSONB,
    after_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_workspace_created
    ON audit_logs (workspace_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor
    ON audit_logs (actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_target
    ON audit_logs (target_type, target_id);

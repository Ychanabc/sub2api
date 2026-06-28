CREATE TABLE IF NOT EXISTS conversation_audit_access_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL DEFAULT 0,
    email TEXT NOT NULL DEFAULT '',
    ip_address TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    fingerprint TEXT NOT NULL DEFAULT '',
    success BOOLEAN NOT NULL DEFAULT FALSE,
    failure_reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversation_audit_access_logs_created_at ON conversation_audit_access_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_audit_access_logs_user_created_at ON conversation_audit_access_logs(user_id, created_at DESC);

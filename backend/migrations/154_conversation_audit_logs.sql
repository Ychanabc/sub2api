CREATE TABLE IF NOT EXISTS conversation_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    user_id BIGINT NOT NULL DEFAULT 0,
    api_key_id BIGINT NOT NULL DEFAULT 0,
    account_id BIGINT NOT NULL DEFAULT 0,
    group_id BIGINT,
    model TEXT NOT NULL DEFAULT '',
    endpoint TEXT NOT NULL DEFAULT '',
    request_type SMALLINT NOT NULL DEFAULT 0,
    request_excerpt TEXT NOT NULL DEFAULT '',
    response_excerpt TEXT NOT NULL DEFAULT '',
    request_hash TEXT NOT NULL DEFAULT '',
    response_hash TEXT NOT NULL DEFAULT '',
    status_code INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversation_audit_logs_created_at ON conversation_audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_audit_logs_user_created_at ON conversation_audit_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_audit_logs_api_key_created_at ON conversation_audit_logs(api_key_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_audit_logs_request_id ON conversation_audit_logs(request_id);

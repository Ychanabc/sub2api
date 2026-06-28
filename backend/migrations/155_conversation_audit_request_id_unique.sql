DELETE FROM conversation_audit_logs a
USING conversation_audit_logs b
WHERE a.request_id = b.request_id
  AND a.id > b.id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_conversation_audit_logs_request_id_uniq
ON conversation_audit_logs(request_id);

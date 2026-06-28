package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type conversationAuditRepository struct {
	db *sql.DB
}

func NewConversationAuditRepository(db *sql.DB) service.ConversationAuditRepository {
	return &conversationAuditRepository{db: db}
}

func (r *conversationAuditRepository) Create(ctx context.Context, log *service.ConversationAuditLog) error {
	if r == nil || r.db == nil || log == nil {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO conversation_audit_logs (
			request_id, user_id, api_key_id, account_id, group_id, model, endpoint,
			request_type, request_excerpt, response_excerpt, request_hash, response_hash,
			status_code, duration_ms, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (request_id) DO UPDATE SET
			user_id = CASE WHEN EXCLUDED.user_id <> 0 THEN EXCLUDED.user_id ELSE conversation_audit_logs.user_id END,
			api_key_id = CASE WHEN EXCLUDED.api_key_id <> 0 THEN EXCLUDED.api_key_id ELSE conversation_audit_logs.api_key_id END,
			account_id = CASE WHEN EXCLUDED.account_id <> 0 THEN EXCLUDED.account_id ELSE conversation_audit_logs.account_id END,
			group_id = COALESCE(EXCLUDED.group_id, conversation_audit_logs.group_id),
			model = CASE WHEN EXCLUDED.model <> '' THEN EXCLUDED.model ELSE conversation_audit_logs.model END,
			endpoint = CASE WHEN EXCLUDED.endpoint <> '' THEN EXCLUDED.endpoint ELSE conversation_audit_logs.endpoint END,
			request_type = CASE WHEN EXCLUDED.request_type <> 0 THEN EXCLUDED.request_type ELSE conversation_audit_logs.request_type END,
			request_excerpt = CASE WHEN EXCLUDED.request_excerpt <> '' THEN EXCLUDED.request_excerpt ELSE conversation_audit_logs.request_excerpt END,
			response_excerpt = CASE WHEN EXCLUDED.response_excerpt <> '' THEN EXCLUDED.response_excerpt ELSE conversation_audit_logs.response_excerpt END,
			request_hash = CASE WHEN EXCLUDED.request_hash <> '' THEN EXCLUDED.request_hash ELSE conversation_audit_logs.request_hash END,
			response_hash = CASE WHEN EXCLUDED.response_hash <> '' THEN EXCLUDED.response_hash ELSE conversation_audit_logs.response_hash END,
			status_code = CASE WHEN EXCLUDED.status_code <> 0 THEN EXCLUDED.status_code ELSE conversation_audit_logs.status_code END,
			duration_ms = CASE WHEN EXCLUDED.duration_ms <> 0 THEN EXCLUDED.duration_ms ELSE conversation_audit_logs.duration_ms END
	`, log.RequestID, log.UserID, log.APIKeyID, log.AccountID, log.GroupID, log.Model, log.Endpoint,
		log.RequestType, log.RequestExcerpt, log.ResponseExcerpt, log.RequestHash, log.ResponseHash,
		log.StatusCode, log.DurationMs, log.CreatedAt)
	if err != nil {
		slog.Warn("conversation audit insert failed", "error", err)
	}
	return err
}

func (r *conversationAuditRepository) DeleteAll(ctx context.Context) (int64, error) {
	if r == nil || r.db == nil {
		return 0, nil
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM conversation_audit_logs`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *conversationAuditRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	if r == nil || r.db == nil {
		return 0, nil
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM conversation_audit_logs WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *conversationAuditRepository) List(ctx context.Context, filter service.ConversationAuditFilter) ([]*service.ConversationAuditLog, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, nil
	}
	where, args := buildConversationAuditWhere(filter)
	countSQL := "SELECT COUNT(*) FROM conversation_audit_logs" + where
	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit := filter.PageSize
	offset := (filter.Page - 1) * filter.PageSize
	query := `
		SELECT id, request_id, user_id, api_key_id, account_id, group_id, model, endpoint,
		       request_type, request_excerpt, response_excerpt, request_hash, response_hash,
		       status_code, duration_ms, created_at
		FROM conversation_audit_logs` + where + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]*service.ConversationAuditLog, 0, limit)
	for rows.Next() {
		var item service.ConversationAuditLog
		if err := rows.Scan(
			&item.ID, &item.RequestID, &item.UserID, &item.APIKeyID, &item.AccountID, &item.GroupID,
			&item.Model, &item.Endpoint, &item.RequestType, &item.RequestExcerpt, &item.ResponseExcerpt,
			&item.RequestHash, &item.ResponseHash, &item.StatusCode, &item.DurationMs, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, &item)
	}
	return items, total, rows.Err()
}

func buildConversationAuditWhere(filter service.ConversationAuditFilter) (string, []any) {
	var conditions []string
	var args []any
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if filter.UserID > 0 {
		add("user_id = $%d", filter.UserID)
	}
	if filter.APIKeyID > 0 {
		add("api_key_id = $%d", filter.APIKeyID)
	}
	if filter.AccountID > 0 {
		add("account_id = $%d", filter.AccountID)
	}
	if strings.TrimSpace(filter.Model) != "" {
		add("model ILIKE $%d", "%"+strings.TrimSpace(filter.Model)+"%")
	}
	if strings.TrimSpace(filter.Search) != "" {
		args = append(args, "%"+strings.TrimSpace(filter.Search)+"%")
		idx := len(args)
		conditions = append(conditions, fmt.Sprintf("(request_id ILIKE $%d OR request_excerpt ILIKE $%d OR response_excerpt ILIKE $%d)", idx, idx, idx))
	}
	if filter.StartTime != nil {
		add("created_at >= $%d", *filter.StartTime)
	}
	if filter.EndTime != nil {
		add("created_at <= $%d", *filter.EndTime)
	}
	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type conversationAuditAccessRepository struct {
	db *sql.DB
}

func NewConversationAuditAccessRepository(db *sql.DB) service.ConversationAuditAccessLogRepository {
	return &conversationAuditAccessRepository{db: db}
}

func (r *conversationAuditAccessRepository) Create(ctx context.Context, log *service.ConversationAuditAccessLog) error {
	if r == nil || r.db == nil || log == nil {
		return nil
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO conversation_audit_access_logs (
			user_id, email, ip_address, user_agent, fingerprint, success, failure_reason, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, log.UserID, log.Email, log.IPAddress, log.UserAgent, log.Fingerprint, log.Success, log.FailureReason, log.CreatedAt)
	return err
}

func (r *conversationAuditAccessRepository) List(ctx context.Context, page, pageSize int) ([]*service.ConversationAuditAccessLog, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, nil
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM conversation_audit_access_logs`).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit := pageSize
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, email, ip_address, user_agent, fingerprint, success, failure_reason, created_at
		FROM conversation_audit_access_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]*service.ConversationAuditAccessLog, 0, limit)
	for rows.Next() {
		var item service.ConversationAuditAccessLog
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.Email, &item.IPAddress, &item.UserAgent,
			&item.Fingerprint, &item.Success, &item.FailureReason, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, &item)
	}
	return items, total, rows.Err()
}

func (r *conversationAuditAccessRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	if r == nil || r.db == nil {
		return 0, nil
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM conversation_audit_access_logs WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

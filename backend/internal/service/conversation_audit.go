package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type ConversationAuditLog struct {
	ID              int64     `json:"id"`
	RequestID       string    `json:"request_id"`
	UserID          int64     `json:"user_id"`
	APIKeyID        int64     `json:"api_key_id"`
	AccountID       int64     `json:"account_id"`
	GroupID         *int64    `json:"group_id,omitempty"`
	Model           string    `json:"model"`
	Endpoint        string    `json:"endpoint"`
	RequestType     int16     `json:"request_type"`
	RequestExcerpt  string    `json:"request_excerpt"`
	ResponseExcerpt string    `json:"response_excerpt"`
	RequestHash     string    `json:"request_hash"`
	ResponseHash    string    `json:"response_hash"`
	StatusCode      int       `json:"status_code"`
	DurationMs      int       `json:"duration_ms"`
	CreatedAt       time.Time `json:"created_at"`
}

type ConversationAuditFilter struct {
	Page      int
	PageSize  int
	UserID    int64
	APIKeyID  int64
	AccountID int64
	Model     string
	Search    string
	StartTime *time.Time
	EndTime   *time.Time
}

type ConversationAuditList struct {
	Items      []*ConversationAuditLog      `json:"items"`
	Pagination *pagination.PaginationResult `json:"pagination"`
}

type ConversationAuditRepository interface {
	Create(ctx context.Context, log *ConversationAuditLog) error
	List(ctx context.Context, filter ConversationAuditFilter) ([]*ConversationAuditLog, int64, error)
	DeleteAll(ctx context.Context) (int64, error)
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

// ConversationAuditAccessLog 记录每一次访问对话审计的尝试（用于安全审计）。
type ConversationAuditAccessLog struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	Email         string    `json:"email"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	Fingerprint   string    `json:"fingerprint"`
	Success       bool      `json:"success"`
	FailureReason string    `json:"failure_reason"`
	CreatedAt     time.Time `json:"created_at"`
}

type ConversationAuditAccessLogList struct {
	Items      []*ConversationAuditAccessLog `json:"items"`
	Pagination *pagination.PaginationResult  `json:"pagination"`
}

type ConversationAuditAccessLogRepository interface {
	Create(ctx context.Context, log *ConversationAuditAccessLog) error
	List(ctx context.Context, page, pageSize int) ([]*ConversationAuditAccessLog, int64, error)
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

type ConversationAuditService struct {
	repo            ConversationAuditRepository
	accessRepo      ConversationAuditAccessLogRepository
	settingService  *SettingService
	lastCleanupUnix int64
}

func NewConversationAuditService(repo ConversationAuditRepository, accessRepo ConversationAuditAccessLogRepository, settingService *SettingService) *ConversationAuditService {
	return &ConversationAuditService{
		repo:           repo,
		accessRepo:     accessRepo,
		settingService: settingService,
	}
}

func (s *ConversationAuditService) Record(ctx context.Context, log *ConversationAuditLog) {
	if s == nil || s.repo == nil || log == nil {
		return
	}
	log.RequestExcerpt = TruncateConversationAuditText(log.RequestExcerpt)
	log.ResponseExcerpt = TruncateConversationAuditText(log.ResponseExcerpt)
	if log.RequestHash == "" {
		log.RequestHash = HashConversationAuditText(log.RequestExcerpt)
	}
	if log.ResponseHash == "" {
		log.ResponseHash = HashConversationAuditText(log.ResponseExcerpt)
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	_ = s.repo.Create(ctx, log)
}

func (s *ConversationAuditService) List(ctx context.Context, filter ConversationAuditFilter) (*ConversationAuditList, error) {
	s.RunConfiguredCleanupIfDue(ctx)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ConversationAuditList{
		Items: items,
		Pagination: &pagination.PaginationResult{
			Page:     filter.Page,
			PageSize: filter.PageSize,
			Total:    total,
			Pages:    int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize)),
		},
	}, nil
}

// DeleteAll 清空全部对话审计记录，返回删除行数。
func (s *ConversationAuditService) DeleteAll(ctx context.Context) (int64, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}
	return s.repo.DeleteAll(ctx)
}

// DeleteOlderThan 删除早于 cutoff 的对话审计记录，返回删除行数。
func (s *ConversationAuditService) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}
	return s.repo.DeleteOlderThan(ctx, cutoff)
}

func (s *ConversationAuditService) RecordAccess(ctx context.Context, log *ConversationAuditAccessLog) {
	if s == nil || s.accessRepo == nil || log == nil {
		return
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	if err := s.accessRepo.Create(ctx, log); err != nil {
		slog.Warn("conversation audit access log insert failed", "user_id", log.UserID, "success", log.Success, "error", err)
	}
}

func (s *ConversationAuditService) ListAccessLogs(ctx context.Context, page, pageSize int) (*ConversationAuditAccessLogList, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	if s == nil || s.accessRepo == nil {
		return &ConversationAuditAccessLogList{
			Items: []*ConversationAuditAccessLog{},
			Pagination: &pagination.PaginationResult{
				Page:     page,
				PageSize: pageSize,
				Total:    0,
				Pages:    0,
			},
		}, nil
	}
	items, total, err := s.accessRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &ConversationAuditAccessLogList{
		Items: items,
		Pagination: &pagination.PaginationResult{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
			Pages:    int((total + int64(pageSize) - 1) / int64(pageSize)),
		},
	}, nil
}

func (s *ConversationAuditService) DeleteAccessLogsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	if s == nil || s.accessRepo == nil {
		return 0, nil
	}
	return s.accessRepo.DeleteOlderThan(ctx, cutoff)
}

func (s *ConversationAuditService) RunConfiguredCleanupIfDue(ctx context.Context) {
	if s == nil || s.settingService == nil || s.repo == nil {
		return
	}
	now := time.Now()
	last := atomic.LoadInt64(&s.lastCleanupUnix)
	if last > 0 && now.Sub(time.Unix(last, 0)) < time.Hour {
		return
	}
	if !atomic.CompareAndSwapInt64(&s.lastCleanupUnix, last, now.Unix()) {
		return
	}
	settings, err := s.settingService.GetAllSettings(ctx)
	if err != nil {
		slog.Warn("conversation audit cleanup skipped: load settings failed", "error", err)
		return
	}
	if settings == nil || !settings.ConversationAuditCleanupEnabled {
		return
	}
	retentionDays := settings.ConversationAuditRetentionDays
	if retentionDays <= 0 {
		retentionDays = 90
	}
	if retentionDays > 3650 {
		retentionDays = 3650
	}
	cutoff := now.AddDate(0, 0, -retentionDays)
	deleted, err := s.repo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		slog.Warn("conversation audit cleanup failed", "error", err)
		return
	}
	var accessDeleted int64
	if s.accessRepo != nil {
		accessDeleted, err = s.accessRepo.DeleteOlderThan(ctx, cutoff)
		if err != nil {
			slog.Warn("conversation audit access cleanup failed", "error", err)
			return
		}
	}
	if deleted > 0 || accessDeleted > 0 {
		slog.Info("conversation audit cleanup completed", "deleted", deleted, "access_deleted", accessDeleted, "retention_days", retentionDays)
	}
}

func HashConversationAuditText(text string) string {
	if text == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func writeConversationAuditStart(ctx context.Context, repo ConversationAuditRepository, input *ConversationAuditStartInput, source string) {
	if repo == nil || input == nil {
		return
	}
	// Resolve the request id with the same precedence used by the usage/billing
	// finish path (resolveUsageBillingRequestID). The start record and the finish
	// record must share an identical request_id so the ON CONFLICT (request_id)
	// upsert merges them into a single audit row (request + response).
	requestID := resolveUsageBillingRequestID(ctx, strings.TrimSpace(input.RequestID))
	if strings.TrimSpace(requestID) == "" {
		return
	}
	user := input.User
	if user == nil && input.APIKey != nil {
		user = input.APIKey.User
	}
	userID := int64(0)
	if user != nil {
		userID = user.ID
	}
	apiKeyID := int64(0)
	var groupID *int64
	if input.APIKey != nil {
		apiKeyID = input.APIKey.ID
		groupID = input.APIKey.GroupID
	}
	statusCode := input.StatusCode
	if statusCode == 0 {
		statusCode = 102
	}
	auditCtx, cancel := detachedBillingContext(ctx)
	defer cancel()
	if err := repo.Create(auditCtx, &ConversationAuditLog{
		RequestID:      requestID,
		UserID:         userID,
		APIKeyID:       apiKeyID,
		GroupID:        groupID,
		Model:          strings.TrimSpace(input.Model),
		Endpoint:       strings.TrimSpace(input.InboundEndpoint),
		RequestType:    input.RequestType,
		RequestExcerpt: TruncateConversationAuditText(input.RequestExcerpt),
		RequestHash:    firstNonEmptyString(input.RequestPayloadHash, HashConversationAuditText(input.RequestExcerpt)),
		StatusCode:     statusCode,
		CreatedAt:      time.Now(),
	}); err != nil {
		slog.Warn("conversation audit start failed", "source", source, "request_id", requestID, "error", err)
	}
}

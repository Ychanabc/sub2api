package admin

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const (
	conversationAuditUnlockTTL   = 15 * time.Minute
	conversationAuditTokenHeader = "X-Conversation-Audit-Token"
)

type ConversationAuditHandler struct {
	service        *service.ConversationAuditService
	settingService *service.SettingService
	userService    *service.UserService
	totpService    *service.TotpService
	tokenSecret    string
}

type conversationAuditUnlockRequest struct {
	SecondaryPassword string `json:"secondary_password"`
	TotpCode          string `json:"totp_code"`
	Fingerprint       string `json:"fingerprint"`
}

type conversationAuditUnlockTokenPayload struct {
	UserID      int64  `json:"uid"`
	ExpiresAt   int64  `json:"exp"`
	Nonce       string `json:"nonce"`
	Fingerprint string `json:"fp,omitempty"`
}

func NewConversationAuditHandler(
	service *service.ConversationAuditService,
	settingService *service.SettingService,
	userService *service.UserService,
	totpService *service.TotpService,
	cfg *config.Config,
) *ConversationAuditHandler {
	secret := ""
	if cfg != nil {
		secret = strings.TrimSpace(cfg.JWT.Secret)
	}
	return &ConversationAuditHandler{
		service:        service,
		settingService: settingService,
		userService:    userService,
		totpService:    totpService,
		tokenSecret:    secret,
	}
}

func (h *ConversationAuditHandler) Status(c *gin.Context) {
	settings, err := h.loadAuditSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	subject, _ := middleware.GetAuthSubjectFromContext(c)
	totpEnabled := false
	if h.totpService != nil && subject.UserID > 0 {
		totpEnabled, _ = h.totpService.IsTotpEnabledForUser(c.Request.Context(), subject.UserID)
	}
	response.Success(c, gin.H{
		"secondary_password_configured": settings.ConversationAuditSecondaryPasswordConfigured,
		"totp_enabled":                  totpEnabled,
		"unlock_ttl_seconds":            int(conversationAuditUnlockTTL.Seconds()),
		"cleanup_enabled":               settings.ConversationAuditCleanupEnabled,
		"retention_days":                settings.ConversationAuditRetentionDays,
	})
}

func (h *ConversationAuditHandler) Unlock(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "Conversation audit service not available")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Admin login required")
		return
	}
	var req conversationAuditUnlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordAccessAttempt(c, subject.UserID, "", "", false, "invalid_request")
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	req.Fingerprint = strings.TrimSpace(req.Fingerprint)
	userEmail := h.userEmail(c.Request.Context(), subject.UserID)
	settings, err := h.loadAuditSettings(c.Request.Context())
	if err != nil {
		h.recordAccessAttempt(c, subject.UserID, userEmail, req.Fingerprint, false, "settings_unavailable")
		response.ErrorFrom(c, err)
		return
	}
	if strings.TrimSpace(settings.ConversationAuditSecondaryPasswordHash) == "" {
		h.recordAccessAttempt(c, subject.UserID, userEmail, req.Fingerprint, false, "secondary_password_not_configured")
		response.ErrorFrom(c, infraerrors.Forbidden("CONVERSATION_AUDIT_LOCKED", "configure a secondary password before accessing conversation audit"))
		return
	}
	if strings.TrimSpace(req.SecondaryPassword) == "" {
		h.recordAccessAttempt(c, subject.UserID, userEmail, req.Fingerprint, false, "secondary_password_required")
		response.BadRequest(c, "Secondary password is required")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(settings.ConversationAuditSecondaryPasswordHash), []byte(req.SecondaryPassword)); err != nil {
		h.recordAccessAttempt(c, subject.UserID, userEmail, req.Fingerprint, false, "secondary_password_invalid")
		response.ErrorFrom(c, infraerrors.Forbidden("CONVERSATION_AUDIT_SECONDARY_PASSWORD_INVALID", "secondary password is invalid"))
		return
	}
	if h.totpService == nil {
		h.recordAccessAttempt(c, subject.UserID, userEmail, req.Fingerprint, false, "totp_service_unavailable")
		response.Error(c, http.StatusServiceUnavailable, "TOTP service not available")
		return
	}
	totpEnabled, err := h.totpService.IsTotpEnabledForUser(c.Request.Context(), subject.UserID)
	if err != nil {
		h.recordAccessAttempt(c, subject.UserID, userEmail, req.Fingerprint, false, "totp_status_error")
		response.ErrorFrom(c, err)
		return
	}
	if !totpEnabled {
		h.recordAccessAttempt(c, subject.UserID, userEmail, req.Fingerprint, false, "totp_not_enabled")
		response.ErrorFrom(c, infraerrors.Forbidden("CONVERSATION_AUDIT_TOTP_REQUIRED", "enable Google Authenticator 2FA for the admin account before accessing conversation audit"))
		return
	}
	if strings.TrimSpace(req.TotpCode) == "" {
		h.recordAccessAttempt(c, subject.UserID, userEmail, req.Fingerprint, false, "totp_required")
		response.BadRequest(c, "TOTP code is required")
		return
	}
	if err := h.totpService.VerifyCode(c.Request.Context(), subject.UserID, strings.TrimSpace(req.TotpCode)); err != nil {
		h.recordAccessAttempt(c, subject.UserID, userEmail, req.Fingerprint, false, "totp_invalid")
		response.ErrorFrom(c, err)
		return
	}
	token, expiresAt, err := h.issueUnlockToken(subject.UserID, req.Fingerprint)
	if err != nil {
		h.recordAccessAttempt(c, subject.UserID, userEmail, req.Fingerprint, false, "token_issue_failed")
		response.ErrorFrom(c, err)
		return
	}
	h.recordAccessAttempt(c, subject.UserID, userEmail, req.Fingerprint, true, "")
	h.service.RunConfiguredCleanupIfDue(c.Request.Context())
	response.Success(c, gin.H{
		"unlock_token": token,
		"expires_at":   expiresAt.Format(time.RFC3339),
		"expires_in":   int(time.Until(expiresAt).Seconds()),
	})
}

func (h *ConversationAuditHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "Conversation audit service not available")
		return
	}
	if !h.requireUnlockToken(c) {
		return
	}
	page, pageSize := response.ParsePagination(c)
	filter := service.ConversationAuditFilter{
		Page:     page,
		PageSize: pageSize,
		Model:    strings.TrimSpace(c.Query("model")),
		Search:   strings.TrimSpace(c.Query("q")),
	}
	filter.UserID = parsePositiveInt64(c.Query("user_id"))
	filter.APIKeyID = parsePositiveInt64(c.Query("api_key_id"))
	filter.AccountID = parsePositiveInt64(c.Query("account_id"))
	if start := parseRFC3339Query(c.Query("start_time")); start != nil {
		filter.StartTime = start
	}
	if end := parseRFC3339Query(c.Query("end_time")); end != nil {
		filter.EndTime = end
	}
	out, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, out)
}

func (h *ConversationAuditHandler) ClearAll(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "Conversation audit service not available")
		return
	}
	if !h.requireUnlockToken(c) {
		return
	}
	deleted, err := h.service.DeleteAll(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": deleted})
}

func (h *ConversationAuditHandler) AccessLogs(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "Conversation audit service not available")
		return
	}
	if !h.requireUnlockToken(c) {
		return
	}
	page, pageSize := response.ParsePagination(c)
	out, err := h.service.ListAccessLogs(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, out)
}

func (h *ConversationAuditHandler) requireUnlockToken(c *gin.Context) bool {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Admin login required")
		return false
	}
	token := strings.TrimSpace(c.GetHeader(conversationAuditTokenHeader))
	if token == "" {
		token = strings.TrimSpace(c.Query("unlock_token"))
	}
	payload, err := h.verifyUnlockToken(token)
	if err != nil {
		response.ErrorFrom(c, err)
		return false
	}
	if payload.UserID != subject.UserID {
		response.ErrorFrom(c, infraerrors.Forbidden("CONVERSATION_AUDIT_TOKEN_USER_MISMATCH", "unlock token does not match current admin"))
		return false
	}
	return true
}

func (h *ConversationAuditHandler) loadAuditSettings(ctx context.Context) (*service.SystemSettings, error) {
	if h == nil || h.settingService == nil {
		return nil, infraerrors.InternalServer("SETTINGS_SERVICE_UNAVAILABLE", "settings service not available")
	}
	return h.settingService.GetAllSettings(ctx)
}

func (h *ConversationAuditHandler) recordAccessAttempt(c *gin.Context, userID int64, email string, fingerprint string, success bool, failureReason string) {
	if h == nil || h.service == nil {
		return
	}
	if c != nil {
		headerFingerprint := strings.TrimSpace(c.GetHeader("X-Client-Fingerprint"))
		if fingerprint == "" {
			fingerprint = headerFingerprint
		}
	}
	userAgent := ""
	ip := ""
	if c != nil {
		userAgent = c.GetHeader("User-Agent")
		ip = c.ClientIP()
	}
	ctx := context.Background()
	if c != nil && c.Request != nil {
		ctx = c.Request.Context()
	}
	h.service.RecordAccess(ctx, &service.ConversationAuditAccessLog{
		UserID:        userID,
		Email:         email,
		IPAddress:     ip,
		UserAgent:     userAgent,
		Fingerprint:   fingerprint,
		Success:       success,
		FailureReason: failureReason,
		CreatedAt:     time.Now(),
	})
}

func (h *ConversationAuditHandler) userEmail(ctx context.Context, userID int64) string {
	if h == nil || h.userService == nil || userID <= 0 {
		return ""
	}
	user, err := h.userService.GetByID(ctx, userID)
	if err != nil || user == nil {
		return ""
	}
	return user.Email
}

func (h *ConversationAuditHandler) issueUnlockToken(userID int64, fingerprint string) (string, time.Time, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(conversationAuditUnlockTTL)
	payload := conversationAuditUnlockTokenPayload{
		UserID:      userID,
		ExpiresAt:   expiresAt.Unix(),
		Nonce:       hex.EncodeToString(nonce),
		Fingerprint: strings.TrimSpace(fingerprint),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(raw)
	signature := h.signUnlockPayload(encodedPayload)
	return encodedPayload + "." + signature, expiresAt, nil
}

func (h *ConversationAuditHandler) verifyUnlockToken(token string) (*conversationAuditUnlockTokenPayload, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, infraerrors.Unauthorized("CONVERSATION_AUDIT_UNLOCK_REQUIRED", "conversation audit unlock is required")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, infraerrors.Unauthorized("CONVERSATION_AUDIT_TOKEN_INVALID", "conversation audit unlock token is invalid")
	}
	expected := h.signUnlockPayload(parts[0])
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[1])) != 1 {
		return nil, infraerrors.Unauthorized("CONVERSATION_AUDIT_TOKEN_INVALID", "conversation audit unlock token is invalid")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, infraerrors.Unauthorized("CONVERSATION_AUDIT_TOKEN_INVALID", "conversation audit unlock token is invalid")
	}
	var payload conversationAuditUnlockTokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, infraerrors.Unauthorized("CONVERSATION_AUDIT_TOKEN_INVALID", "conversation audit unlock token is invalid")
	}
	if payload.UserID <= 0 || payload.ExpiresAt <= time.Now().Unix() {
		return nil, infraerrors.Unauthorized("CONVERSATION_AUDIT_TOKEN_EXPIRED", "conversation audit unlock token has expired")
	}
	return &payload, nil
}

func (h *ConversationAuditHandler) signUnlockPayload(payload string) string {
	secret := strings.TrimSpace(h.tokenSecret)
	if secret == "" {
		secret = "conversation-audit-local-secret"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("conversation-audit-unlock:v1:"))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func parsePositiveInt64(raw string) int64 {
	n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func parseRFC3339Query(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return &t
}

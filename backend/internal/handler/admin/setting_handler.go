package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// semverPattern 婵犵數濮烽弫鍛婃叏閻戝鈧倿鎸婃竟鈺嬬秮瀹曘劑寮堕幋鐙呯幢闂備礁鎲℃笟妤呭矗鎼淬劍鍋勯柛鈩兠肩换鍡樸亜閺嶃劍鐨戞慨锝囧仱閺屾稓鈧綆鍋呯亸顓熴亜椤愶絿绠為柟顔瑰墲閹棃濮€椤厼顥氭繝鐢靛Х椤ｄ粙鍩€椤掆偓閸熻法鐥閺屾盯寮埀顒勬偡閳哄懎违濞撴埃鍋撶€殿噮鍣ｅ畷鐓庘攽閸℃埃鍋撻崹顔规斀闁宠棄妫楅悘銉︾箾閸滃啰绉柕?semver 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犖ч柛銉㈡櫇閸樼娀姊绘担绋款棌闁稿鎳庣叅闁哄稁鍘肩壕濂告煏婵炑冨閺嬫垿妫呴銏″闁瑰嘲缍婇弫鎰緞婵犲嫮鏉搁梻浣虹帛椤ㄥ懘鎮ф繝鍌楁瀺闁靛繈鍊栭埛鎴犵磽娴ｅ箍鈧偤骞嬮敐鍐ф睏闂佸憡鍔﹂崰鏍嚋瑜版帗鐓忓┑鐐靛亾濞呭棝鏌嶉柨瀣伌闁诡喖缍婇獮渚€骞掗幋婵愮€抽梻浣告惈椤戝棝寮婚妸鈺佺厴闁硅揪瀵岄弫濠囨煟閹惧啿顒㈡禍娑㈡⒒娴ｉ涓茬紒鎻掓健瀹曟顫滈埀顒€顕ｉ锕€绠涢柡澶婄仢閼板灝鈹戦埥鍡楃仩闁圭⒈鍋呯€靛吋鎯旈姀銏㈢槇闂佹眹鍨藉褍鐡梻浣筋嚙缁绘垵鐣濋幖浣哄祦闁归偊鍘归崑鍛存煕閹扳晛濡挎い锔诲亝缁绘稓鈧數顭堢敮鍫曟煟鎺抽崝鎴﹀箖?
var semverPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// menuItemIDPattern validates custom menu item IDs: alphanumeric, hyphens, underscores only.
var menuItemIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// generateMenuItemID generates a short random hex ID for a custom menu item.
func generateMenuItemID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate menu item ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func scopesContainOpenID(scopes string) bool {
	for _, scope := range strings.Fields(strings.ToLower(strings.TrimSpace(scopes))) {
		if scope == "openid" {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// SettingHandler handles system settings.
type SettingHandler struct {
	settingService           *service.SettingService
	emailService             *service.EmailService
	turnstileService         *service.TurnstileService
	opsService               *service.OpsService
	paymentConfigService     *service.PaymentConfigService
	paymentService           *service.PaymentService
	userAttributeService     *service.UserAttributeService
	notificationEmailService *service.NotificationEmailService
}

// NewSettingHandler creates a settings handler.
func NewSettingHandler(settingService *service.SettingService, emailService *service.EmailService, turnstileService *service.TurnstileService, opsService *service.OpsService, paymentConfigService *service.PaymentConfigService, paymentService *service.PaymentService, userAttributeService *service.UserAttributeService) *SettingHandler {
	return &SettingHandler{
		settingService:       settingService,
		emailService:         emailService,
		turnstileService:     turnstileService,
		opsService:           opsService,
		paymentConfigService: paymentConfigService,
		paymentService:       paymentService,
		userAttributeService: userAttributeService,
	}
}

// SetNotificationEmailService attaches the notification template service without changing
// the constructor signature used by existing unit tests.
func (h *SettingHandler) SetNotificationEmailService(notificationEmailService *service.NotificationEmailService) {
	h.notificationEmailService = notificationEmailService
}

// GetSettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕閻庤娲忛崕鎶藉焵椤掑﹦绉甸柛鐘崇墱婢规洟宕稿Δ浣哄幍闂佽鍨虫晶妤吽夋径鎰闁哄鍩婇煬顒勬煛鐏炶鈧繈骞婂┑瀣妞ゆ棁鍋愭晶顖氣攽閻愯尙鎽犵紒顔肩灱缁辩偞绻濋崶褑鎽曞┑鐐村灟閸ㄧ懓鏁俊鐐€栧濠氬储瑜旈敐鐐哄煛閸愵亞锛濇繛杈剧到閹碱偉鈪烽梻浣侯焾濞寸兘寮拠宸殨閻犲洤妯婇崥瀣煕椤愵偄浜濇い搴℃喘濮婄粯鎷呴崨濠傛殘闂佽鎮傜粻鏍х暦閻楀牊鍎熸い顓熷灦閺咁亪姊洪幐搴ｇ畵妞わ富鍨跺畷褰掑磼濞戞牔绨婚梺瑙勫閺呮盯鎮橀鍓х＜闁归偊鍠栨禒閬嶆煙椤旂瓔娈滅€规洘顨嗗鍕節娴ｅ壊妫滈梻鍌氬€风粈渚€骞夐垾瓒佹椽鏁冮崒姘憋紱闂佸憡娲﹂崢浠嬪汲濠婂牊鐓ユ繝闈涙瀹告繄鐥崣銉х煓闁哄本绋撴禒锕傚箲閹邦剦妫熼梻渚€鈧偛鑻晶浼存煕韫囨棑鑰挎鐐诧工铻栭柛娑卞弮閸炲爼姊洪崫鍕窛闁稿鍠栭幃浼村Ψ閳哄倵鎷洪梺鍛婄箓鐎氼參宕掗妸鈺傜厱闁宠桨绀侀顓犫偓娈垮櫘閸嬪懐鎹㈠┑瀣倞闁靛ě鍐ㄧ闂傚倷鐒﹂幃鍫曞磿椤曗偓瀵彃鈽夊Ο缁樼槑缂?// GET /api/v1/admin/settings
func (h *SettingHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingService.GetAllSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	authSourceDefaults, err := h.settingService.GetAuthSourceDefaultSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Check if ops monitoring is enabled (respects config.ops.enabled)
	opsEnabled := h.opsService != nil && h.opsService.IsMonitoringEnabled(c.Request.Context())
	defaultSubscriptions := make([]dto.DefaultSubscriptionSetting, 0, len(settings.DefaultSubscriptions))
	for _, sub := range settings.DefaultSubscriptions {
		defaultSubscriptions = append(defaultSubscriptions, dto.DefaultSubscriptionSetting{
			GroupID:      sub.GroupID,
			ValidityDays: sub.ValidityDays,
		})
	}

	// Load payment config
	var paymentCfg *service.PaymentConfig
	if h.paymentConfigService != nil {
		paymentCfg, _ = h.paymentConfigService.GetPaymentConfig(c.Request.Context())
	}
	if paymentCfg == nil {
		paymentCfg = &service.PaymentConfig{}
	}

	payload := dto.SystemSettings{
		RegistrationEnabled:                          settings.RegistrationEnabled,
		EmailVerifyEnabled:                           settings.EmailVerifyEnabled,
		RegistrationEmailSuffixWhitelist:             settings.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                             settings.PromoCodeEnabled,
		PasswordResetEnabled:                         settings.PasswordResetEnabled,
		FrontendURL:                                  settings.FrontendURL,
		InvitationCodeEnabled:                        settings.InvitationCodeEnabled,
		TotpEnabled:                                  settings.TotpEnabled,
		TotpEncryptionKeyConfigured:                  h.settingService.IsTotpEncryptionKeyConfigured(),
		LoginAgreementEnabled:                        settings.LoginAgreementEnabled,
		ConversationAuditSecondaryPasswordConfigured: settings.ConversationAuditSecondaryPasswordConfigured,
		ConversationAuditCleanupEnabled:              settings.ConversationAuditCleanupEnabled,
		ConversationAuditRetentionDays:               settings.ConversationAuditRetentionDays,
		LoginAgreementMode:                           settings.LoginAgreementMode,
		LoginAgreementUpdatedAt:                      settings.LoginAgreementUpdatedAt,
		LoginAgreementDocuments:                      loginAgreementDocumentsToDTO(settings.LoginAgreementDocuments),
		SMTPHost:                                     settings.SMTPHost,
		SMTPPort:                                     settings.SMTPPort,
		SMTPUsername:                                 settings.SMTPUsername,
		SMTPPasswordConfigured:                       settings.SMTPPasswordConfigured,
		SMTPFrom:                                     settings.SMTPFrom,
		SMTPFromName:                                 settings.SMTPFromName,
		SMTPUseTLS:                                   settings.SMTPUseTLS,
		TurnstileEnabled:                             settings.TurnstileEnabled,
		TurnstileSiteKey:                             settings.TurnstileSiteKey,
		TurnstileSecretKeyConfigured:                 settings.TurnstileSecretKeyConfigured,
		APIKeyACLTrustForwardedIP:                    settings.APIKeyACLTrustForwardedIP,
		LinuxDoConnectEnabled:                        settings.LinuxDoConnectEnabled,
		LinuxDoConnectClientID:                       settings.LinuxDoConnectClientID,
		LinuxDoConnectClientSecretConfigured:         settings.LinuxDoConnectClientSecretConfigured,
		LinuxDoConnectRedirectURL:                    settings.LinuxDoConnectRedirectURL,
		DingTalkConnectEnabled:                       settings.DingTalkConnectEnabled,
		DingTalkConnectClientID:                      settings.DingTalkConnectClientID,
		DingTalkConnectClientSecretConfigured:        settings.DingTalkConnectClientSecretConfigured,
		DingTalkConnectRedirectURL:                   settings.DingTalkConnectRedirectURL,
		DingTalkConnectCorpRestrictionPolicy:         settings.DingTalkConnectCorpRestrictionPolicy,
		DingTalkConnectInternalCorpID:                settings.DingTalkConnectInternalCorpID,
		DingTalkConnectBypassRegistration:            settings.DingTalkConnectBypassRegistration,
		DingTalkConnectSyncCorpEmail:                 settings.DingTalkConnectSyncCorpEmail,
		DingTalkConnectSyncDisplayName:               settings.DingTalkConnectSyncDisplayName,
		DingTalkConnectSyncDept:                      settings.DingTalkConnectSyncDept,
		DingTalkConnectSyncCorpEmailAttrKey:          settings.DingTalkConnectSyncCorpEmailAttrKey,
		DingTalkConnectSyncDisplayNameAttrKey:        settings.DingTalkConnectSyncDisplayNameAttrKey,
		DingTalkConnectSyncDeptAttrKey:               settings.DingTalkConnectSyncDeptAttrKey,
		DingTalkConnectSyncCorpEmailAttrName:         settings.DingTalkConnectSyncCorpEmailAttrName,
		DingTalkConnectSyncDisplayNameAttrName:       settings.DingTalkConnectSyncDisplayNameAttrName,
		DingTalkConnectSyncDeptAttrName:              settings.DingTalkConnectSyncDeptAttrName,
		WeChatConnectEnabled:                         settings.WeChatConnectEnabled,
		WeChatConnectAppID:                           settings.WeChatConnectAppID,
		WeChatConnectAppSecretConfigured:             settings.WeChatConnectAppSecretConfigured,
		WeChatConnectOpenAppID:                       settings.WeChatConnectOpenAppID,
		WeChatConnectOpenAppSecretConfigured:         settings.WeChatConnectOpenAppSecretConfigured,
		WeChatConnectMPAppID:                         settings.WeChatConnectMPAppID,
		WeChatConnectMPAppSecretConfigured:           settings.WeChatConnectMPAppSecretConfigured,
		WeChatConnectMobileAppID:                     settings.WeChatConnectMobileAppID,
		WeChatConnectMobileAppSecretConfigured:       settings.WeChatConnectMobileAppSecretConfigured,
		WeChatConnectOpenEnabled:                     settings.WeChatConnectOpenEnabled,
		WeChatConnectMPEnabled:                       settings.WeChatConnectMPEnabled,
		WeChatConnectMobileEnabled:                   settings.WeChatConnectMobileEnabled,
		WeChatConnectMode:                            settings.WeChatConnectMode,
		WeChatConnectScopes:                          settings.WeChatConnectScopes,
		WeChatConnectRedirectURL:                     settings.WeChatConnectRedirectURL,
		WeChatConnectFrontendRedirectURL:             settings.WeChatConnectFrontendRedirectURL,
		OIDCConnectEnabled:                           settings.OIDCConnectEnabled,
		OIDCConnectProviderName:                      settings.OIDCConnectProviderName,
		OIDCConnectClientID:                          settings.OIDCConnectClientID,
		OIDCConnectClientSecretConfigured:            settings.OIDCConnectClientSecretConfigured,
		OIDCConnectIssuerURL:                         settings.OIDCConnectIssuerURL,
		OIDCConnectDiscoveryURL:                      settings.OIDCConnectDiscoveryURL,
		OIDCConnectAuthorizeURL:                      settings.OIDCConnectAuthorizeURL,
		OIDCConnectTokenURL:                          settings.OIDCConnectTokenURL,
		OIDCConnectUserInfoURL:                       settings.OIDCConnectUserInfoURL,
		OIDCConnectJWKSURL:                           settings.OIDCConnectJWKSURL,
		OIDCConnectScopes:                            settings.OIDCConnectScopes,
		OIDCConnectRedirectURL:                       settings.OIDCConnectRedirectURL,
		OIDCConnectFrontendRedirectURL:               settings.OIDCConnectFrontendRedirectURL,
		OIDCConnectTokenAuthMethod:                   settings.OIDCConnectTokenAuthMethod,
		OIDCConnectUsePKCE:                           settings.OIDCConnectUsePKCE,
		OIDCConnectValidateIDToken:                   settings.OIDCConnectValidateIDToken,
		OIDCConnectAllowedSigningAlgs:                settings.OIDCConnectAllowedSigningAlgs,
		OIDCConnectClockSkewSeconds:                  settings.OIDCConnectClockSkewSeconds,
		OIDCConnectRequireEmailVerified:              settings.OIDCConnectRequireEmailVerified,
		OIDCConnectUserInfoEmailPath:                 settings.OIDCConnectUserInfoEmailPath,
		OIDCConnectUserInfoIDPath:                    settings.OIDCConnectUserInfoIDPath,
		OIDCConnectUserInfoUsernamePath:              settings.OIDCConnectUserInfoUsernamePath,
		GitHubOAuthEnabled:                           settings.GitHubOAuthEnabled,
		GitHubOAuthClientID:                          settings.GitHubOAuthClientID,
		GitHubOAuthClientSecretConfigured:            settings.GitHubOAuthClientSecretConfigured,
		GitHubOAuthRedirectURL:                       settings.GitHubOAuthRedirectURL,
		GitHubOAuthFrontendRedirectURL:               settings.GitHubOAuthFrontendRedirectURL,
		GoogleOAuthEnabled:                           settings.GoogleOAuthEnabled,
		GoogleOAuthClientID:                          settings.GoogleOAuthClientID,
		GoogleOAuthClientSecretConfigured:            settings.GoogleOAuthClientSecretConfigured,
		GoogleOAuthRedirectURL:                       settings.GoogleOAuthRedirectURL,
		GoogleOAuthFrontendRedirectURL:               settings.GoogleOAuthFrontendRedirectURL,
		SiteName:                                     settings.SiteName,
		SiteLogo:                                     settings.SiteLogo,
		SiteSubtitle:                                 settings.SiteSubtitle,
		APIBaseURL:                                   settings.APIBaseURL,
		ContactInfo:                                  settings.ContactInfo,
		DocURL:                                       settings.DocURL,
		HomeContent:                                  settings.HomeContent,
		HideCcsImportButton:                          settings.HideCcsImportButton,
		PurchaseSubscriptionEnabled:                  settings.PurchaseSubscriptionEnabled,
		PurchaseSubscriptionURL:                      settings.PurchaseSubscriptionURL,
		TableDefaultPageSize:                         settings.TableDefaultPageSize,
		TablePageSizeOptions:                         settings.TablePageSizeOptions,
		CustomMenuItems:                              dto.ParseCustomMenuItems(settings.CustomMenuItems),
		CustomEndpoints:                              dto.ParseCustomEndpoints(settings.CustomEndpoints),
		DefaultConcurrency:                           settings.DefaultConcurrency,
		DefaultBalance:                               settings.DefaultBalance,
		RiskControlEnabled:                           settings.RiskControlEnabled,
		CyberSessionBlockEnabled:                     settings.CyberSessionBlockEnabled,
		CyberSessionBlockTTLSeconds:                  settings.CyberSessionBlockTTLSeconds,
		AffiliateRebateRate:                          settings.AffiliateRebateRate,
		AffiliateRebateFreezeHours:                   settings.AffiliateRebateFreezeHours,
		AffiliateRebateDurationDays:                  settings.AffiliateRebateDurationDays,
		AffiliateRebatePerInviteeCap:                 settings.AffiliateRebatePerInviteeCap,
		DefaultUserRPMLimit:                          settings.DefaultUserRPMLimit,
		DefaultSubscriptions:                         defaultSubscriptions,
		EnableModelFallback:                          settings.EnableModelFallback,
		FallbackModelAnthropic:                       settings.FallbackModelAnthropic,
		FallbackModelOpenAI:                          settings.FallbackModelOpenAI,
		FallbackModelGemini:                          settings.FallbackModelGemini,
		FallbackModelAntigravity:                     settings.FallbackModelAntigravity,
		EnableIdentityPatch:                          settings.EnableIdentityPatch,
		IdentityPatchPrompt:                          settings.IdentityPatchPrompt,
		OpsMonitoringEnabled:                         opsEnabled && settings.OpsMonitoringEnabled,
		OpsRealtimeMonitoringEnabled:                 settings.OpsRealtimeMonitoringEnabled,
		OpsQueryModeDefault:                          settings.OpsQueryModeDefault,
		OpsMetricsIntervalSeconds:                    settings.OpsMetricsIntervalSeconds,
		MinClaudeCodeVersion:                         settings.MinClaudeCodeVersion,
		MaxClaudeCodeVersion:                         settings.MaxClaudeCodeVersion,
		AllowUngroupedKeyScheduling:                  settings.AllowUngroupedKeyScheduling,
		BackendModeEnabled:                           settings.BackendModeEnabled,
		EnableFingerprintUnification:                 settings.EnableFingerprintUnification,
		EnableMetadataPassthrough:                    settings.EnableMetadataPassthrough,
		EnableCCHSigning:                             settings.EnableCCHSigning,
		EnableClaudeOAuthSystemPromptInjection:       settings.EnableClaudeOAuthSystemPromptInjection,
		ClaudeOAuthSystemPrompt:                      settings.ClaudeOAuthSystemPrompt,
		ClaudeOAuthSystemPromptBlocks:                settings.ClaudeOAuthSystemPromptBlocks,
		EnableAnthropicCacheTTL1hInjection:           settings.EnableAnthropicCacheTTL1hInjection,
		RewriteMessageCacheControl:                   settings.RewriteMessageCacheControl,
		AntigravityUserAgentVersion:                  settings.AntigravityUserAgentVersion,
		OpenAICodexUserAgent:                         settings.OpenAICodexUserAgent,
		OpenAIAllowClaudeCodeCodexPlugin:             settings.OpenAIAllowClaudeCodeCodexPlugin,
		WebSearchEmulationEnabled:                    settings.WebSearchEmulationEnabled,
		PaymentVisibleMethodAlipaySource:             settings.PaymentVisibleMethodAlipaySource,
		PaymentVisibleMethodWxpaySource:              settings.PaymentVisibleMethodWxpaySource,
		PaymentVisibleMethodAlipayEnabled:            settings.PaymentVisibleMethodAlipayEnabled,
		PaymentVisibleMethodWxpayEnabled:             settings.PaymentVisibleMethodWxpayEnabled,
		OpenAIAdvancedSchedulerEnabled:               settings.OpenAIAdvancedSchedulerEnabled,
		BalanceLowNotifyEnabled:                      settings.BalanceLowNotifyEnabled,
		BalanceLowNotifyThreshold:                    settings.BalanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:                  settings.BalanceLowNotifyRechargeURL,
		SubscriptionExpiryNotifyEnabled:              settings.SubscriptionExpiryNotifyEnabled,
		AccountQuotaNotifyEnabled:                    settings.AccountQuotaNotifyEnabled,
		AccountQuotaNotifyEmails:                     dto.NotifyEmailEntriesFromService(settings.AccountQuotaNotifyEmails),
		PaymentEnabled:                               paymentCfg.Enabled,
		PaymentMinAmount:                             paymentCfg.MinAmount,
		PaymentMaxAmount:                             paymentCfg.MaxAmount,
		PaymentDailyLimit:                            paymentCfg.DailyLimit,
		PaymentOrderTimeoutMin:                       paymentCfg.OrderTimeoutMin,
		PaymentMaxPendingOrders:                      paymentCfg.MaxPendingOrders,
		PaymentEnabledTypes:                          paymentCfg.EnabledTypes,
		PaymentBalanceDisabled:                       paymentCfg.BalanceDisabled,
		PaymentBalanceRechargeMultiplier:             paymentCfg.BalanceRechargeMultiplier,
		PaymentRechargeFeeRate:                       paymentCfg.RechargeFeeRate,
		PaymentLoadBalanceStrat:                      paymentCfg.LoadBalanceStrategy,
		PaymentProductNamePrefix:                     paymentCfg.ProductNamePrefix,
		PaymentProductNameSuffix:                     paymentCfg.ProductNameSuffix,
		PaymentHelpImageURL:                          paymentCfg.HelpImageURL,
		PaymentHelpText:                              paymentCfg.HelpText,
		PaymentCancelRateLimitEnabled:                paymentCfg.CancelRateLimitEnabled,
		PaymentCancelRateLimitMax:                    paymentCfg.CancelRateLimitMax,
		PaymentCancelRateLimitWindow:                 paymentCfg.CancelRateLimitWindow,
		PaymentCancelRateLimitUnit:                   paymentCfg.CancelRateLimitUnit,
		PaymentCancelRateLimitMode:                   paymentCfg.CancelRateLimitMode,
		PaymentAlipayForceQRCode:                     paymentCfg.AlipayForceQRCode,

		ChannelMonitorEnabled:                settings.ChannelMonitorEnabled,
		ChannelMonitorDefaultIntervalSeconds: settings.ChannelMonitorDefaultIntervalSeconds,

		AvailableChannelsEnabled: settings.AvailableChannelsEnabled,

		AffiliateEnabled: settings.AffiliateEnabled,

		AllowUserViewErrorRequests: settings.AllowUserViewErrorRequests,
	}

	// OpenAI fast policy (stored under a dedicated setting key)
	if fastPolicy, err := h.settingService.GetOpenAIFastPolicySettings(c.Request.Context()); err != nil {
		slog.Error("openai_fast_policy_settings_get_failed", "error", err)
	} else if fastPolicy != nil {
		payload.OpenAIFastPolicySettings = openaiFastPolicySettingsToDTO(fastPolicy)
	}

	// Default platform quotas JSON map.
	if platformQuotas, err := h.settingService.GetDefaultPlatformQuotas(c.Request.Context()); err != nil {
		slog.Error("default_platform_quotas_get_failed", "error", err)
	} else {
		payload.DefaultPlatformQuotas = platformQuotas
	}

	response.Success(c, systemSettingsResponseData(payload, authSourceDefaults))
}

// openaiFastPolicySettingsToDTO converts service -> dto for OpenAI fast policy.
func openaiFastPolicySettingsToDTO(s *service.OpenAIFastPolicySettings) *dto.OpenAIFastPolicySettings {
	if s == nil {
		return nil
	}
	rules := make([]dto.OpenAIFastPolicyRule, len(s.Rules))
	for i, r := range s.Rules {
		rules[i] = dto.OpenAIFastPolicyRule(r)
	}
	return &dto.OpenAIFastPolicySettings{Rules: rules}
}

// openaiFastPolicySettingsFromDTO converts dto -> service for OpenAI fast policy.
//
// 闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊瑜忛弳锕傛煟閵忋埄鐒剧紒鎰殜閺岀喖骞嶉纰辨毉闂佺顑戠换婵嬪蓟閺囩喓鐝舵い鏍殔娴滈箖姊虹粙娆惧剱闁瑰憡鎮傞敐鐐测攽鐎ｎ偄浜楅柟鑹版彧缁插ジ宕犻弽顓熲拻濞撴埃鍋撻柍褜鍓涢崑娑㈡嚐椤栨稒娅犳い鏇楀亾闁哄矉缍€缁犳盯寮崹顔芥嚈闂備浇顕栭崯顐﹀炊閵娿儰姹楅梻浣藉亹閳峰牓宕滃顑芥灁?ServiceTier闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍏煎€绘慨妤€妫欓悾鐑芥⒑閹肩偛鈧洟鎯岄崒鐐茶摕闁绘梻鈷堥弫宥夋煕閳╁喚娈橀崯鎼佹⒑?DTO 闂傚倸鍊搁崐鎼佸磹妞嬪海鐭嗗〒姘ｅ亾妤犵偞鐗犻、鏇氱秴闁搞儺鍓﹂弫鍐煥閺囨浜鹃梺姹囧€楅崑鎾舵崲濠靛洨绡€闁稿本绋戝▍褔姊哄ú璇插箺闁荤啿鏅犲濠氭偄閸涘﹦绉堕梺鍛婃寙閸涱喗顔忛梻?service 闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊椤掑鏅悷婊冪箻楠炴垿濮€閵堝懐顔婂┑掳鍊愰崑鎾剁棯閸撗呭笡缂佺粯鐩獮瀣枎韫囨洑鎮ｇ紓浣瑰劤婢т粙骞婇幘鐑┾偓鏃堝礃椤忎礁浜鹃柨婵嗙凹缁ㄥジ鏌熼惂鍝ュ埌闁宠棄顦甸獮妯虹暦閸ュ柌鍥ㄧ厸閻忕偠顕ф俊濂告煃鐟欏嫬鐏寸€规洟浜堕崺锟犲磼閸岋箑顩紒杈ㄦ尰閹峰懘骞撻幒宥咁棜闂傚倷绀佹竟濠囧磻閳ь剟鏌熼崘鍙夊枠鐎规洘鍨块獮姗€鎳￠妶鍛偊闂備礁澹婇悡鍫ュ窗閺嶃劍鍙忛幖绮规閺€浠嬫煟閹邦垰鐨烘繝鈧幘顔界厱闁哄啠鍋撻柛銊ф暬閹儳鐣￠幊濠冩そ椤㈡棃宕ㄩ婵囩潖濠电姷鏁搁崑鐐差焽濞嗘搩鏁勯柛鈩冨喕缂嶆牜鈧箍鍎遍ˇ浼村煕閹达附鐓曢柨鏃囶嚙瀵劍淇婇悙顒佸€愰柡灞剧洴婵℃瓕顦叉い锝堝亹缁辨帡宕掑姣欙綁鏌曢崼顒傜М鐎规洘锕㈤崺鐐烘倷椤掆偓椤忓湱绱撻崒娆掑厡闁惧繐閰ｅ畷銏＄附缁嬭法鐤囬棅顐㈡处濞叉﹢宕楀鍏炬棃鏁愰崨顓熸闂佺粯鎸堕崕鐢稿蓟閿熺姴鐐婇柡鍫㈡暩閺佹牠姊烘导娆戞偧闁轰礁顭峰璇测槈閵忕姷鍘搁梺鍛婂姂閸斿危閹扮増鍊垫繛鍫濈仢閺嬫瑧绱掗鐣屾噰闁靛棔绶氬鎾閳ュ厖绨婚梻浣哄亾閼归箖宕愰幖浣€澶婎潩椤撶姷鐦堥梺姹囧灲濞佳勭閿曞倹鐓曢柕濞垮劤閸╋綁鏌熼鎯у幋妞ゃ垺绋戦～婵嬫晲閸℃瑤绨村┑鐘殿暯濡插懘宕归柆宥嗗剳閻熸瑥瀚々鍙夌節婵犲倻澧涢柣鎾寸懇閹鈽夊▎妯煎姺闂佹椿鍘奸惌鍌炲蓟濞戙垹妫橀悹鎭掑壉閵堝洨纾奸弶鍫涘妼缁椦囨煃瑜滈崜銊х礊閸℃顩查悹杞扮秿閻旂厧鐒垫い鎺嗗亾閾绘牠鏌ｅ鈧褎绂掗敃鍌涚厵闁荤喓澧楅崰姗€鏌ｅ☉鍗炴珝妤犵偞甯掕灃闁逞屽墰閻氭儳顓兼径瀣幈闁诲繒鍋犲Λ鍕焵椤掆偓閻栫厧鐣峰┑鍡╁悑闁搞儯鍔屾惔濠傗攽閻愭潙鐏熼柛銊ユ贡缁鏁愭径瀣幗闂佸綊鍋婇崢鐣岀礊閹达附鐓熼柕濠忕畱閻忓瓨鎱ㄦ繝鍛仩缂侇喗鐟ラ埢搴ㄥ箚瑜嶆竟澶愭⒒娴ｇ儤鍤€闁硅绻濋獮鍐磼閻愬弶妲?// service.OpenAIFastTierAny ("all")闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍛婄秶濡わ絽鍟宥夋⒑缁嬪灝顒㈡い銊ワ躬瀵鈽夊顐ｅ媰闂佸憡鎸嗛埀顒€危閸繍娓婚柕鍫濇缁€瀣瑰鍐煟鐎殿喛顕ч鍏煎緞婵犱胶鐐婇梻浣告啞濞诧箓宕滃▎蹇婃瀺闁挎繂娲ㄧ壕钘壝归敐鍡楃祷濞存粓绠栧铏圭矙閹稿孩鎷辨繝銏ｎ潐濞茬喎鐣烽幋锕€绠婚柤濮愬€曠粊锕傛⒑閸撹尙鍘涢柛鐘崇墬閹便劍鎯旈埦鈧弨浠嬫煟閹邦剛鎽犵紓宥嗗灴閺屾稒绻濋崒婊€铏庨梺浼欑悼閸忔ɑ淇婇悜钘夌厸濞达絽鎼慨锔戒繆閻愵亜鈧牜鏁幒鏂哄亾濮樼厧寮柛鈺傜洴楠炲鏁傞悾灞藉箻濠电姵顔栭崰妤呭礉閺囥垹鐒垫い鎺嶇劍缁€鍐磼閸屾稑绗掓い顓滃姂瀹曠喖顢旈崱娆戞毎濠碉紕鍋戦崐鏍礉瑜忕槐鐐哄幢濡炵粯鐏佸┑鐘诧工鐎氬嘲鈻撴禒瀣厽闁归偊鍓氶埢鏇熶繆椤栨浜鹃梻鍌欑濠€閬嶅储瑜忕槐鐐寸節閸ャ儮鍋撴笟鈧鎾閻樻爠鍥ㄧ厱閻忕偛澧介埊鏇㈡煛閸℃劕鈧洟鍩為幋锔藉亹闁割煈鍋呭В鍕節閻㈤潧浜归柛瀣尭閳规垿鎮欓弶鍨殶闂佺绻掗崑鎾剁矙閹达箑鐓″鑸靛姇椤懘鏌ｅΟ娲诲晱闁哥喎鐗撳缁樻媴閾忕懓绗″銈庡幖濞层劑宕氶幒妤婃晬闁绘劖娼欐禍妤€鈹戦悙鏉戠仧闁搞劍妞介幃锟犲即閵忥紕鍘甸梺閫涚祷濞呮洖鈻嶉崨顓涙斀闁炽儱鍟块幃鎴︽煏閸パ冾伃闁轰焦鍔欏畷鍗烆渻缂佹浜為梻鍌欑閹碱偊鎳熼婊呯煋闁割偅娲栫粻鐔兼煙缂併垹鏋涚紒鈧€ｎ偁浜滈柟鎹愭硾閸撹鲸绻涙總鍓叉缂佽鲸鎹囧畷鎺戔枎閹搭厽袦婵＄偑鍊栭崹鐢稿箠濡寧顥ら梻浣圭湽閸ㄥ綊骞夐敓鐘茬；闁稿本绋忔禍婊堟煙閹峰本纭鹃柣锔界矒閺屸剝鎷呴崨濠傛灎闂佸搫鐭夌徊鍊熺亽闂佹儳绻橀埀顒佺〒妞规娊姊?"all" 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柛鎰级閻ゅ嫬鈹戞幊閸娧呭緤娴犲鐤い鎰╁€愰崑鎾愁潩椤掑效闂侀潧娲ょ€氫即鐛幒妤€骞㈡俊鐐村劤椤ユ艾鈹?
// 闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊瑜忛弳锕傛煟閵忋埄鐒剧痪鎯ь煼閺岋綁骞囬鑺ユ瘎闂佹椿鍘介悷鈺呮偂椤愶箑鐐婇柕濞р偓濡插牏绱撴担鎻掍壕婵犮垼鍩栭崝妤呭矗閹剧粯鐓曢柕澶涚到婵″潡鏌嶉挊澶樻Ш缂?闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柛鎰紦缂冩洟姊绘担鍛婃儓婵炲眰鍔嶉幈銊︻槹鎼达絿鐒兼繝鐢靛Т閸婄敻寮ㄦ禒瀣厽婵☆垰鎼痪褔鏌熼崗鐓庡闁哄本绋撴禒锔炬嫚閹绘帩娼庨梻浣告惈閺堫剙煤閻旈鏆﹂柣鎾崇岸閺€浠嬫煕閵夈劌鐓愰柨鐔村劜娣囧﹪鎮欓鍕ㄥ亾閹版澘鐤鹃柣妯哄殺婢舵劕绠婚柤鍛婎問濞肩喖姊虹憴鍕姢闁宦板妽閸掑﹪寮堕幋鏃€鏂€濡炪倖妫佸Λ鍕倿閹灐鐟邦煥閸曨厾鐣肩紓?tier"闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亜顒㈡い鎰矙閺屻劑鎮㈤崫鍕戙垻鐥幆褜鐓奸柡灞剧洴閸╁嫰宕楅埡鍌氱伌鐎规洩缍€缁犳稑鈽夊▎鎴濆箺闂備線娼х换鍫ュ春閸曨垱鍊垮Δ锝呭暞閻撴洟鏌曟繛鍨姎闁逞屽墮缂嶅﹪鐛崼銉ノ╅柕澶婃捣閸犳牠鐛幇顓熷劅闁挎繂鍟犻崑鎾诲箛閻楀牃鎷洪梺鍛婄箓鐎氱兘宕曟惔锝囩＜闁兼悂娼ч崫铏光偓娈垮枛椤兘寮幇鏉垮窛闁稿本绋掗ˉ鍫ユ煕閳规儳浜炬俊鐐€栫敮鎺斺偓姘煎弮瀹曟垹鈧綆鍠楅悡鏇㈡煃閳轰礁鏆熼柟鍐叉嚇閺岋綁顢橀悙闈涒叺闂佸搫琚崝鎴濐嚕閹绢喗鍊锋繛鏉戭儏娴滈箖鏌涘┑鍕姢濞戞挸绉堕幉鎼佹偋閸繄鐟查梺绋匡工閻忔氨鎹㈠☉銏犵闁绘垵妫旂槐妤呮⒑閻熼偊鍤熷┑顔藉▕閹偞绻濋崘锔跨盎闂佽宕樺▔娑欑墡闂備礁鎽滄慨鐢稿礉濞嗘挸钃熼柍鈺佸暙缁剁偤鎮楅敐搴濇喚闁搞倕顑夊娲传閵夈儛锝嗐亜椤撶偞鍠橀柛鈹惧亾濡炪倖甯掗崰姘焽閹邦厾绠鹃柛娆忣檧閼拌法鈧娲樼换鍌炲煝鎼淬劌绠荤€规洖娲ㄩ弸鍐⒒娴ｅ憡鎯堢紒瀣╃窔瀹曘垺绺介崨濠備簵濠电偞鍨崹娲偂閺囥垺鍊堕柣鎰絻閳锋棃鏌ｉ鐑嗗剱缂佺粯鐩畷濂稿閳哄啫濮奸梻浣哥枃椤宕归崸妤€绠栨繛鍡樻尰閸ゆ垶銇勯幒宥囧妽闁哥姵鐗滅槐鎾诲磼濞嗘垼绐楅梺鍝ュУ閻楃姴顕ｉ悽鍓叉晢闁稿本绮庢径鍕箾鐎电孝妞ゆ垵鎳愰悮鎯ь吋婢跺鍘遍柣蹇曞仜婢т粙鎯岄崱娆戠＜婵°倐鍋撴い锕傛涧椤繘鎼归崷顓狅紲濠碘槅鍨抽崕鐢稿箯婵犳碍鈷戦柛娑橆煬閻掑ジ鏌涘☉鍗炵仧缂佹顦靛娲偡闁箑娈舵繝娈垮枤閺佽鐣锋导鏉戠疀闁绘鐗嗛埀顒傛暬閺屻劌鈹戦崱娑扁偓妤呮煛鐎ｎ偆娲撮柡宀嬬秮楠炴帡骞嬮悩杈╅┏闂備線娼уú銈団偓姘嵆閻涱噣骞掑Δ鈧獮銏′繆閻愭潙鍔ゆい銉﹀哺濮婂宕掑▎鎺戝帯濡炪値鍘奸悧鎾诲极閸愵喖唯闁宠桨鑳堕悾娲⒑缁嬫寧婀版慨妯稿姂瀵娊鏁冮崒娑氬弳濠电娀娼уΛ顓炍ｉ崫銉х＜闁逞屽墴瀹曞ジ濮€閳ヨ櫕鐎惧┑鐘灱閸╂牠宕濋弴鐘差棜濠电姵纰嶉悡娆撴煕閹炬鎳庣粭锟犳⒑缁嬫鍎滅紒缁樼箞瀵鍩勯崘銊х獮闁诲函缍嗘禍鍫曞吹鐎ｎ亖鏀介柣鎰皺婢э絾銇勯弴鍡楁搐閻撴﹢鏌熸潏楣冩闁稿﹪顥撻惀顏堟偋閸╄櫕鍨归埀顒佺缁嬫帞鎹㈠☉姘ｅ亾濞戞瑯鐒介柣顓炵焸閺屾稓鈧綆鍓欓弸娑氣偓瑙勬礃瀹€鎼佺嵁閹烘绠ｆい鎾跺枎閸忓﹪姊洪懡銈呪枅缂傚倹鑹鹃々濂稿Ω瑜庨～鏇㈡煕椤愶絾绀冮柍閿嬪灴閺屾稑鈽夊鍫熸暰缂佸墽鍋撻幃鍌炲蓟濞戙垹妫橀柟绋垮閹疯京绱撴担绋库偓鍝ョ矓閻㈢围闁挎繂顦粈鍐煃閸濆嫬鏆欐鐐搭殜濮婄粯鎷呴崨濠呯闂佺顑嗛幐鑽ゆ崲濞戞瑧绡€闁搞儴鍩栭弲娑樷攽鎺抽崐鏇㈠箠韫囨稒鍋傛繛鎴欏灪閻撴洟鎮橀悙鑸殿棄闁哄棙鐟︾换娑氭嫚瑜忛悾鐢告煙椤斿厜鍋撻弬銉︻潔濠殿喗锕徊鑺ョ閻愵剦娈介柣鎰皺娴犮垽鏌涢弮鈧喊宥囨崲濞戞矮娌柛灞惧焹閸嬫挸鈹戦崱娆愭闂侀潧绻堥崐鏇犵不缂佹ǜ浜滈柡鍐ㄥ€瑰▍鏇㈡煕濡粯宕岄柟顔煎槻椤劑宕熼鐘靛帨闁诲氦顫夊ú妯煎垝瀹€鍕厺閹兼番鍔岀粻锝夋煟閹邦厼绲荤紒銊ｅ劦濮婅櫣鎷犻幓鎺戞瘣缂傚倸绉村Λ婵嗙暦濠婂喚娼╅悹浣告贡缁嬪繘姊洪崫鍕偍闁搞劍妞介幃锟犲即閻旂繝绨婚梺瑙勬緲婢у海绮欓懡銈囩＜闁逞屽墴瀹曟﹢顢欓悾灞藉箰闂備礁鎲＄划鍫㈢矆娴ｈ娅犲ù鐓庣摠閻撳啰鎲稿鍫濈婵炲棙鎸搁崒銊︾節婵犲倻澧曠痪鎯ь煼閺岀喖宕滆鐢盯鏌ｉ幘鍐叉殶闁硅尙顭堥…銊╁醇濠靛牜妲堕梻浣虹帛閺屻劑銆冩惔銊ュ惞?
// openaiFastPolicySettingsFromDTO converts dto -> service for OpenAI fast policy.
func openaiFastPolicySettingsFromDTO(s *dto.OpenAIFastPolicySettings) *service.OpenAIFastPolicySettings {
	if s == nil {
		return nil
	}
	rules := make([]service.OpenAIFastPolicyRule, len(s.Rules))
	for i, r := range s.Rules {
		rules[i] = service.OpenAIFastPolicyRule(r)
		tier := strings.ToLower(strings.TrimSpace(rules[i].ServiceTier))
		if tier == "" {
			tier = service.OpenAIFastTierAny
		}
		rules[i].ServiceTier = tier
	}
	return &service.OpenAIFastPolicySettings{Rules: rules}
}

func loginAgreementDocumentsToDTO(items []service.LoginAgreementDocument) []dto.LoginAgreementDocument {
	result := make([]dto.LoginAgreementDocument, 0, len(items))
	for _, item := range items {
		result = append(result, dto.LoginAgreementDocument{
			ID:        item.ID,
			Title:     item.Title,
			ContentMD: item.ContentMD,
		})
	}
	return result
}

func loginAgreementDocumentsToService(items []dto.LoginAgreementDocument) []service.LoginAgreementDocument {
	result := make([]service.LoginAgreementDocument, 0, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		content := strings.TrimSpace(item.ContentMD)
		if title == "" && content == "" {
			continue
		}
		result = append(result, service.LoginAgreementDocument{
			ID:        strings.TrimSpace(item.ID),
			Title:     title,
			ContentMD: content,
		})
	}
	return result
}

// UpdateSettingsRequest 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犵厱婵﹩鍘介妵婵嗏攽闄囬崺鏍ь嚗閸曨厸鍋撻敐搴濈胺婵″弶鍔欏缁樼瑹閳ь剙顭囪婢ф繈姊洪崫鍕櫤闁烩晩鍨堕獮蹇涘箣閿旇棄浜滈柣蹇撶箣閻掞箓寮埀顒勬⒒娴ｈ櫣甯涢拑杈╂喐閺夊灝鏆ｆい銏℃椤㈡棃宕熼妸锔芥澑婵＄偑鍊栭悧妤呭传鎼淬劌纾婚柕蹇嬪€栭悡娑㈡倶閻愭彃鈷旀繛鎻掔摠椤ㄣ儵鎮欓悿娈夸海閻忓啴姊洪柅鐐茶嫰婢у瓨顨ラ悙璇ч練缂佺姵绋戦埥澶娾枎韫囧﹤浜鹃柛顭戝亖娴滄粓鏌熼幆褍鑸归柣蹇ｄ邯閺岋綁骞樼捄鐑樼亪闂佸搫鐭夌徊鍊熺亽闂佸憡绻傜€氼喗顨ラ崶顒佲拺闁告縿鍎辨牎闂佹寧娲忛崹钘夘嚕婵犳碍鏅搁柣妯垮皺閸旓箑顪冮妶鍡楃瑨閻庢凹鍙冮幃鐐烘嚃閳规儳浜炬鐐茬仢閸旀碍淇婇锝庢疁鐎?
type UpdateSettingsRequest struct {
	// 濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴ｆ閺嬩線鏌熼梻瀵割槮缁惧墽绮换娑㈠箣閺冣偓閸ゅ秹鏌涢妷顔煎⒒闁轰礁娲弻鏇＄疀閺囩倫銉︺亜閿旇娅嶉柟顔筋殜瀹曟寰勬繝浣割棜闂傚倷绀侀幉鈥趁洪敃鍌氱；濠㈣埖鍔曢弰銉╂煟閹邦喖鍔嬮柍閿嬪灥闇夐柛蹇氬亹閹冲懘鏌涢幒宥呭祮闁哄矉绻濆畷濂割敃閵忕媭娼介梻浣告啞鐢帡鏁冮妷褏绱﹀ù鐘差儏瀹告繂鈹戦悩鎻掝仼妤犵偛鐗撳缁樻媴閸涘﹨纭€闂佺顑嗛幑鍥Υ閸愵喖宸濋柡澶嬪灩閻ゅ洭妫呴銏″缂佸甯″畷?
	RegistrationEnabled                bool                         `json:"registration_enabled"`
	EmailVerifyEnabled                 bool                         `json:"email_verify_enabled"`
	RegistrationEmailSuffixWhitelist   []string                     `json:"registration_email_suffix_whitelist"`
	PromoCodeEnabled                   bool                         `json:"promo_code_enabled"`
	PasswordResetEnabled               bool                         `json:"password_reset_enabled"`
	FrontendURL                        string                       `json:"frontend_url"`
	InvitationCodeEnabled              bool                         `json:"invitation_code_enabled"`
	TotpEnabled                        bool                         `json:"totp_enabled"`
	LoginAgreementEnabled              bool                         `json:"login_agreement_enabled"`
	ConversationAuditSecondaryPassword string                       `json:"conversation_audit_secondary_password"`
	ConversationAuditCleanupEnabled    *bool                        `json:"conversation_audit_cleanup_enabled"`
	ConversationAuditRetentionDays     *int                         `json:"conversation_audit_retention_days"`
	LoginAgreementMode                 string                       `json:"login_agreement_mode"`
	LoginAgreementUpdatedAt            string                       `json:"login_agreement_updated_at"`
	LoginAgreementDocuments            []dto.LoginAgreementDocument `json:"login_agreement_documents"`

	// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亜顒㈡い鎰矙閺屻劑鎮㈤崫鍕戙垽鏌ｉ鐕佹疁妤犵偞鐗曡彁妞ゆ巻鍋撳┑陇鍋愮槐鎺楀箛椤撗勭杹闂佸搫鐭夌换婵嗙暦閸洖鐓涘ù锝夋敱閻繘姊绘担鍛婃儓妞ゆ垵鎳橀弻濠囨晲婢跺﹨鎽曢梺鍝勬祫缁辨洟鎮块埀顒勬煟鎼搭垳绉靛ù婊勭墵瀹曟垿骞樼拠鎻掑祮闂佺粯锕╅崑鍕閻愵剚鍙忔慨妤€妫楁晶缁樹繆閸欏銇濋柡灞界Х椤т線鏌涢幘瀵告噮缂侇喛顕ч鍏煎緞婵犲嫸绱甸梻鍌欑贰閸撴瑧绮旂€靛摜涓嶉柣妯款嚙缁犲綊鏌熺喊鍗炲箻闁告ɑ绮庣槐鎾愁吋閸℃瑥鈷岄梺鍝勬湰濞茬喎鐣烽悡搴樻斀闁归偊鍘滆濮婅櫣绮欏▎鎯у妧缂備浇寮撶划娆忥耿娓氣偓濮婃椽骞愭惔锝囩暤濠电偞娼欓崐鍧楀箖閻愭番鍋呴柛鎰ㄦ櫇閸樻捇姊绘担鍝ヤ虎妞ゆ垵妫濊棢闁割偁鍎查悡?
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	SMTPFrom     string `json:"smtp_from_email"`
	SMTPFromName string `json:"smtp_from_name"`
	SMTPUseTLS   bool   `json:"smtp_use_tls"`

	// Cloudflare Turnstile 闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊瑜忛弳锕傛煕椤垵浜濋柛娆忕箳閳ь剝顫夊ú鏍洪妸鈺傚剹闁糕剝顦鸿ぐ鎺撴櫜闁割偒鍋呯紞鍫濃攽閻愬弶鍣藉┑顔芥尦濠€渚€姊虹紒妯忣亜螣婵犲洤纾块柟鎵閻?
	TurnstileEnabled   bool   `json:"turnstile_enabled"`
	TurnstileSiteKey   string `json:"turnstile_site_key"`
	TurnstileSecretKey string `json:"turnstile_secret_key"`

	// API Key IP 闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊瑜忛弳锕傛煕椤垵浜濋柛娆忕箳閳ь剝顫夊ú鏍洪妸鈺傚剹闁糕剝顦鸿ぐ鎺撴櫜闁告侗鍙庡Λ宀勬⒑缂佹ê绗傜紒顔界懇瀵顓兼径濞劑鏌ㄩ弬鍨挃闁绘侗鍠栭埞鎴︽倷閺夋垹浠ч梺鎼炲妼缂嶅﹪寮荤€ｎ喖鐐婃い鎺嶈兌閸樻悂姊洪幖鐐插姉闁哄懏绋掔粋鎺曨槼闁靛洤瀚板顒勫礌闄囬崥顐︽⒑鐠団€虫灈闁搞垺鐓￠垾锕傚Ω閳轰胶顦ㄥ銈呭瘨閸ㄥ酣宕ㄩ鍕闁荤喐鐟ョ€氼厾绮堥埀顒勬⒑閸︻厸鎷￠柛瀣樀閹﹢宕橀瑙ｆ嫼闂佸憡绋戦敃銈嗘叏閳ь剟姊洪崫鍕櫤闁烩晩鍨堕妴渚€骞橀弬銉︻潔闂侀潧绻嗗褔骞忓ú顏呪拺闁告稑锕﹂埥澶愭煥閺囨ê鍔ら柍褜鍓欓悘姘舵偡閳哄懎钃熸繛鎴炃氬Σ鍫ユ煕濡ゅ啫浠﹂柣蹇旀崌濮?
	APIKeyACLTrustForwardedIP *bool `json:"api_key_acl_trust_forwarded_ip"`

	// LinuxDo Connect OAuth 闂傚倸鍊搁崐鎼佸磹閹间礁纾圭€瑰嫭鍣磋ぐ鎺戠倞妞ゆ帊绀侀崜顓烆渻閵堝棗濮х紒鐘冲灴閻涱噣濮€閵堝棛鍘撻柡澶屽仦婢瑰棝宕濆鍡愪簻闁哄倸鐏濋顐ょ磼?
	LinuxDoConnectEnabled      bool   `json:"linuxdo_connect_enabled"`
	LinuxDoConnectClientID     string `json:"linuxdo_connect_client_id"`
	LinuxDoConnectClientSecret string `json:"linuxdo_connect_client_secret"`
	LinuxDoConnectRedirectURL  string `json:"linuxdo_connect_redirect_url"`

	// DingTalk Connect OAuth 闂傚倸鍊搁崐鎼佸磹閹间礁纾圭€瑰嫭鍣磋ぐ鎺戠倞妞ゆ帊绀侀崜顓烆渻閵堝棗濮х紒鐘冲灴閻涱噣濮€閵堝棛鍘撻柡澶屽仦婢瑰棝宕濆鍡愪簻闁哄倸鐏濋顐ょ磼?
	DingTalkConnectEnabled                 bool   `json:"dingtalk_connect_enabled"`
	DingTalkConnectClientID                string `json:"dingtalk_connect_client_id"`
	DingTalkConnectClientSecret            string `json:"dingtalk_connect_client_secret"`
	DingTalkConnectRedirectURL             string `json:"dingtalk_connect_redirect_url"`
	DingTalkConnectCorpRestrictionPolicy   string `json:"dingtalk_connect_corp_restriction_policy"`
	DingTalkConnectInternalCorpID          string `json:"dingtalk_connect_internal_corp_id"`
	DingTalkConnectBypassRegistration      bool   `json:"dingtalk_connect_bypass_registration"`
	DingTalkConnectSyncCorpEmail           bool   `json:"dingtalk_connect_sync_corp_email"`
	DingTalkConnectSyncDisplayName         bool   `json:"dingtalk_connect_sync_display_name"`
	DingTalkConnectSyncDept                bool   `json:"dingtalk_connect_sync_dept"`
	DingTalkConnectSyncCorpEmailAttrKey    string `json:"dingtalk_connect_sync_corp_email_attr_key"`
	DingTalkConnectSyncDisplayNameAttrKey  string `json:"dingtalk_connect_sync_display_name_attr_key"`
	DingTalkConnectSyncDeptAttrKey         string `json:"dingtalk_connect_sync_dept_attr_key"`
	DingTalkConnectSyncCorpEmailAttrName   string `json:"dingtalk_connect_sync_corp_email_attr_name"`
	DingTalkConnectSyncDisplayNameAttrName string `json:"dingtalk_connect_sync_display_name_attr_name"`
	DingTalkConnectSyncDeptAttrName        string `json:"dingtalk_connect_sync_dept_attr_name"`

	// WeChat Connect OAuth 闂傚倸鍊搁崐鎼佸磹閹间礁纾圭€瑰嫭鍣磋ぐ鎺戠倞妞ゆ帊绀侀崜顓烆渻閵堝棗濮х紒鐘冲灴閻涱噣濮€閵堝棛鍘撻柡澶屽仦婢瑰棝宕濆鍡愪簻闁哄倸鐏濋顐ょ磼?
	WeChatConnectEnabled             bool   `json:"wechat_connect_enabled"`
	WeChatConnectAppID               string `json:"wechat_connect_app_id"`
	WeChatConnectAppSecret           string `json:"wechat_connect_app_secret"`
	WeChatConnectOpenAppID           string `json:"wechat_connect_open_app_id"`
	WeChatConnectOpenAppSecret       string `json:"wechat_connect_open_app_secret"`
	WeChatConnectMPAppID             string `json:"wechat_connect_mp_app_id"`
	WeChatConnectMPAppSecret         string `json:"wechat_connect_mp_app_secret"`
	WeChatConnectMobileAppID         string `json:"wechat_connect_mobile_app_id"`
	WeChatConnectMobileAppSecret     string `json:"wechat_connect_mobile_app_secret"`
	WeChatConnectOpenEnabled         bool   `json:"wechat_connect_open_enabled"`
	WeChatConnectMPEnabled           bool   `json:"wechat_connect_mp_enabled"`
	WeChatConnectMobileEnabled       bool   `json:"wechat_connect_mobile_enabled"`
	WeChatConnectMode                string `json:"wechat_connect_mode"`
	WeChatConnectScopes              string `json:"wechat_connect_scopes"`
	WeChatConnectRedirectURL         string `json:"wechat_connect_redirect_url"`
	WeChatConnectFrontendRedirectURL string `json:"wechat_connect_frontend_redirect_url"`

	// Generic OIDC OAuth 闂傚倸鍊搁崐鎼佸磹閹间礁纾圭€瑰嫭鍣磋ぐ鎺戠倞妞ゆ帊绀侀崜顓烆渻閵堝棗濮х紒鐘冲灴閻涱噣濮€閵堝棛鍘撻柡澶屽仦婢瑰棝宕濆鍡愪簻闁哄倸鐏濋顐ょ磼?
	OIDCConnectEnabled              bool   `json:"oidc_connect_enabled"`
	OIDCConnectProviderName         string `json:"oidc_connect_provider_name"`
	OIDCConnectClientID             string `json:"oidc_connect_client_id"`
	OIDCConnectClientSecret         string `json:"oidc_connect_client_secret"`
	OIDCConnectIssuerURL            string `json:"oidc_connect_issuer_url"`
	OIDCConnectDiscoveryURL         string `json:"oidc_connect_discovery_url"`
	OIDCConnectAuthorizeURL         string `json:"oidc_connect_authorize_url"`
	OIDCConnectTokenURL             string `json:"oidc_connect_token_url"`
	OIDCConnectUserInfoURL          string `json:"oidc_connect_userinfo_url"`
	OIDCConnectJWKSURL              string `json:"oidc_connect_jwks_url"`
	OIDCConnectScopes               string `json:"oidc_connect_scopes"`
	OIDCConnectRedirectURL          string `json:"oidc_connect_redirect_url"`
	OIDCConnectFrontendRedirectURL  string `json:"oidc_connect_frontend_redirect_url"`
	OIDCConnectTokenAuthMethod      string `json:"oidc_connect_token_auth_method"`
	OIDCConnectUsePKCE              *bool  `json:"oidc_connect_use_pkce"`
	OIDCConnectValidateIDToken      *bool  `json:"oidc_connect_validate_id_token"`
	OIDCConnectAllowedSigningAlgs   string `json:"oidc_connect_allowed_signing_algs"`
	OIDCConnectClockSkewSeconds     int    `json:"oidc_connect_clock_skew_seconds"`
	OIDCConnectRequireEmailVerified bool   `json:"oidc_connect_require_email_verified"`
	OIDCConnectUserInfoEmailPath    string `json:"oidc_connect_userinfo_email_path"`
	OIDCConnectUserInfoIDPath       string `json:"oidc_connect_userinfo_id_path"`
	OIDCConnectUserInfoUsernamePath string `json:"oidc_connect_userinfo_username_path"`

	GitHubOAuthEnabled             bool   `json:"github_oauth_enabled"`
	GitHubOAuthClientID            string `json:"github_oauth_client_id"`
	GitHubOAuthClientSecret        string `json:"github_oauth_client_secret"`
	GitHubOAuthRedirectURL         string `json:"github_oauth_redirect_url"`
	GitHubOAuthFrontendRedirectURL string `json:"github_oauth_frontend_redirect_url"`
	GoogleOAuthEnabled             bool   `json:"google_oauth_enabled"`
	GoogleOAuthClientID            string `json:"google_oauth_client_id"`
	GoogleOAuthClientSecret        string `json:"google_oauth_client_secret"`
	GoogleOAuthRedirectURL         string `json:"google_oauth_redirect_url"`
	GoogleOAuthFrontendRedirectURL string `json:"google_oauth_frontend_redirect_url"`

	// OEM闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊瑜忛弳锕傛煕椤垵浜濋柛娆忕箳閳ь剝顫夊ú鏍洪妸鈺傚剹闁糕剝顦鸿ぐ鎺撴櫜闁割偒鍋呯紞鍫濃攽閻愬弶鍣藉┑顔芥尦濠€渚€姊虹紒妯忣亜螣婵犲洤纾块柟鎵閻?
	SiteName                    string                `json:"site_name"`
	SiteLogo                    string                `json:"site_logo"`
	SiteSubtitle                string                `json:"site_subtitle"`
	APIBaseURL                  string                `json:"api_base_url"`
	ContactInfo                 string                `json:"contact_info"`
	DocURL                      string                `json:"doc_url"`
	HomeContent                 string                `json:"home_content"`
	HideCcsImportButton         bool                  `json:"hide_ccs_import_button"`
	PurchaseSubscriptionEnabled *bool                 `json:"purchase_subscription_enabled"`
	PurchaseSubscriptionURL     *string               `json:"purchase_subscription_url"`
	TableDefaultPageSize        int                   `json:"table_default_page_size"`
	TablePageSizeOptions        []int                 `json:"table_page_size_options"`
	CustomMenuItems             *[]dto.CustomMenuItem `json:"custom_menu_items"`
	CustomEndpoints             *[]dto.CustomEndpoint `json:"custom_endpoints"`

	// 婵犵數濮烽弫鍛婃叏閻㈠壊鏁婇柡宥庡幖缁愭淇婇妶鍛殲闁哄棙绮嶆穱濠囧Χ閸涱厽娈堕梺娲诲幗閻熲晠寮婚悢鍏煎€绘俊顖炴櫜缁爼姊洪柅鐐茶嫰婢у墽绱掗悩铏碍闁伙綁鏀辩缓鐣岀矙鐠囦勘鍔嶉妵鍕籍閸ヮ灝鎾寸箾閸涱厾孝闁宠鍨块幃鈺呭垂椤愶絾鐦庡┑鐘愁問閸犳岸寮幖浣哥獥濠电姴鍋嗛崥瀣熆鐠轰警鍎忓ù婊勵殜濮婃椽鏌呴悙鑼跺濠⒀勬尦閺岀喖顢欑粵瀣杹闂佺粯渚楅崳锝呯暦瑜版帩鏁婇柣鎾冲瘨濞兼稑鈹戦悩鍨毄闁稿鐩幆鍥ㄥ閺夋垹顦梺鐟扮摠缁诲秹宕?
	DefaultConcurrency                        int                               `json:"default_concurrency"`
	DefaultBalance                            float64                           `json:"default_balance"`
	AffiliateRebateRate                       *float64                          `json:"affiliate_rebate_rate"`
	AffiliateRebateFreezeHours                *int                              `json:"affiliate_rebate_freeze_hours"`
	AffiliateRebateDurationDays               *int                              `json:"affiliate_rebate_duration_days"`
	AffiliateRebatePerInviteeCap              *float64                          `json:"affiliate_rebate_per_invitee_cap"`
	DefaultUserRPMLimit                       int                               `json:"default_user_rpm_limit"`
	DefaultSubscriptions                      []dto.DefaultSubscriptionSetting  `json:"default_subscriptions"`
	AuthSourceDefaultEmailBalance             *float64                          `json:"auth_source_default_email_balance"`
	AuthSourceDefaultEmailConcurrency         *int                              `json:"auth_source_default_email_concurrency"`
	AuthSourceDefaultEmailSubscriptions       *[]dto.DefaultSubscriptionSetting `json:"auth_source_default_email_subscriptions"`
	AuthSourceDefaultEmailGrantOnSignup       *bool                             `json:"auth_source_default_email_grant_on_signup"`
	AuthSourceDefaultEmailGrantOnFirstBind    *bool                             `json:"auth_source_default_email_grant_on_first_bind"`
	AuthSourceDefaultLinuxDoBalance           *float64                          `json:"auth_source_default_linuxdo_balance"`
	AuthSourceDefaultLinuxDoConcurrency       *int                              `json:"auth_source_default_linuxdo_concurrency"`
	AuthSourceDefaultLinuxDoSubscriptions     *[]dto.DefaultSubscriptionSetting `json:"auth_source_default_linuxdo_subscriptions"`
	AuthSourceDefaultLinuxDoGrantOnSignup     *bool                             `json:"auth_source_default_linuxdo_grant_on_signup"`
	AuthSourceDefaultLinuxDoGrantOnFirstBind  *bool                             `json:"auth_source_default_linuxdo_grant_on_first_bind"`
	AuthSourceDefaultOIDCBalance              *float64                          `json:"auth_source_default_oidc_balance"`
	AuthSourceDefaultOIDCConcurrency          *int                              `json:"auth_source_default_oidc_concurrency"`
	AuthSourceDefaultOIDCSubscriptions        *[]dto.DefaultSubscriptionSetting `json:"auth_source_default_oidc_subscriptions"`
	AuthSourceDefaultOIDCGrantOnSignup        *bool                             `json:"auth_source_default_oidc_grant_on_signup"`
	AuthSourceDefaultOIDCGrantOnFirstBind     *bool                             `json:"auth_source_default_oidc_grant_on_first_bind"`
	AuthSourceDefaultWeChatBalance            *float64                          `json:"auth_source_default_wechat_balance"`
	AuthSourceDefaultWeChatConcurrency        *int                              `json:"auth_source_default_wechat_concurrency"`
	AuthSourceDefaultWeChatSubscriptions      *[]dto.DefaultSubscriptionSetting `json:"auth_source_default_wechat_subscriptions"`
	AuthSourceDefaultWeChatGrantOnSignup      *bool                             `json:"auth_source_default_wechat_grant_on_signup"`
	AuthSourceDefaultWeChatGrantOnFirstBind   *bool                             `json:"auth_source_default_wechat_grant_on_first_bind"`
	AuthSourceDefaultGitHubBalance            *float64                          `json:"auth_source_default_github_balance"`
	AuthSourceDefaultGitHubConcurrency        *int                              `json:"auth_source_default_github_concurrency"`
	AuthSourceDefaultGitHubSubscriptions      *[]dto.DefaultSubscriptionSetting `json:"auth_source_default_github_subscriptions"`
	AuthSourceDefaultGitHubGrantOnSignup      *bool                             `json:"auth_source_default_github_grant_on_signup"`
	AuthSourceDefaultGitHubGrantOnFirstBind   *bool                             `json:"auth_source_default_github_grant_on_first_bind"`
	AuthSourceDefaultGoogleBalance            *float64                          `json:"auth_source_default_google_balance"`
	AuthSourceDefaultGoogleConcurrency        *int                              `json:"auth_source_default_google_concurrency"`
	AuthSourceDefaultGoogleSubscriptions      *[]dto.DefaultSubscriptionSetting `json:"auth_source_default_google_subscriptions"`
	AuthSourceDefaultGoogleGrantOnSignup      *bool                             `json:"auth_source_default_google_grant_on_signup"`
	AuthSourceDefaultGoogleGrantOnFirstBind   *bool                             `json:"auth_source_default_google_grant_on_first_bind"`
	AuthSourceDefaultDingTalkBalance          *float64                          `json:"auth_source_default_dingtalk_balance"`
	AuthSourceDefaultDingTalkConcurrency      *int                              `json:"auth_source_default_dingtalk_concurrency"`
	AuthSourceDefaultDingTalkSubscriptions    *[]dto.DefaultSubscriptionSetting `json:"auth_source_default_dingtalk_subscriptions"`
	AuthSourceDefaultDingTalkGrantOnSignup    *bool                             `json:"auth_source_default_dingtalk_grant_on_signup"`
	AuthSourceDefaultDingTalkGrantOnFirstBind *bool                             `json:"auth_source_default_dingtalk_grant_on_first_bind"`
	ForceEmailOnThirdPartySignup              *bool                             `json:"force_email_on_third_party_signup"`

	// Model fallback configuration
	EnableModelFallback      bool   `json:"enable_model_fallback"`
	FallbackModelAnthropic   string `json:"fallback_model_anthropic"`
	FallbackModelOpenAI      string `json:"fallback_model_openai"`
	FallbackModelGemini      string `json:"fallback_model_gemini"`
	FallbackModelAntigravity string `json:"fallback_model_antigravity"`

	// Identity patch configuration (Claude -> Gemini)
	EnableIdentityPatch bool   `json:"enable_identity_patch"`
	IdentityPatchPrompt string `json:"identity_patch_prompt"`

	// Ops monitoring (vNext)
	OpsMonitoringEnabled         *bool   `json:"ops_monitoring_enabled"`
	OpsRealtimeMonitoringEnabled *bool   `json:"ops_realtime_monitoring_enabled"`
	OpsQueryModeDefault          *string `json:"ops_query_mode_default"`
	OpsMetricsIntervalSeconds    *int    `json:"ops_metrics_interval_seconds"`

	MinClaudeCodeVersion string `json:"min_claude_code_version"`
	MaxClaudeCodeVersion string `json:"max_claude_code_version"`

	// 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极瀹ュ绀嬫い鎺嶇劍椤斿洭姊绘担鍛婅础闁稿簺鍊濆畷鐢告晝閳ь剟鍩ユ径濞㈢喖鏌ㄧ€ｅ灚缍屽┑鐘愁問閸犳鏁冮埡鍛婵炲棙鎸婚崐璺衡攽閻樺弶澶勯柛濠傜仛椤ㄣ儵鎮欓懠顑胯檸闂佸憡姊圭喊宥夊Φ閸曨垱鏅滈柛顭戝枛缁侇噣姊虹化鏇熸珕闁烩晩鍨伴锝夊箻椤旂⒈娼婇梺鎸庣☉鐎氼剟鐛幇鐗堚拻濞达絽鎲￠崯鐐烘煕閺冣偓閸ㄥ灝鐣峰┑鍫滄勃闁告挆灞鹃敜闂備胶绮崝妤呭磿閵堝鍋?
	AllowUngroupedKeyScheduling bool `json:"allow_ungrouped_key_scheduling"`

	// Backend Mode
	BackendModeEnabled bool `json:"backend_mode_enabled"`

	// Gateway forwarding behavior
	EnableFingerprintUnification           *bool   `json:"enable_fingerprint_unification"`
	EnableMetadataPassthrough              *bool   `json:"enable_metadata_passthrough"`
	EnableCCHSigning                       *bool   `json:"enable_cch_signing"`
	EnableClaudeOAuthSystemPromptInjection *bool   `json:"enable_claude_oauth_system_prompt_injection"`
	ClaudeOAuthSystemPrompt                *string `json:"claude_oauth_system_prompt"`
	ClaudeOAuthSystemPromptBlocks          *string `json:"claude_oauth_system_prompt_blocks"`
	EnableAnthropicCacheTTL1hInjection     *bool   `json:"enable_anthropic_cache_ttl_1h_injection"`
	RewriteMessageCacheControl             *bool   `json:"rewrite_message_cache_control"`
	AntigravityUserAgentVersion            *string `json:"antigravity_user_agent_version"`
	OpenAICodexUserAgent                   *string `json:"openai_codex_user_agent"`
	OpenAIAllowClaudeCodeCodexPlugin       *bool   `json:"openai_allow_claude_code_codex_plugin"`

	// Payment visible method routing
	PaymentVisibleMethodAlipaySource  *string `json:"payment_visible_method_alipay_source"`
	PaymentVisibleMethodWxpaySource   *string `json:"payment_visible_method_wxpay_source"`
	PaymentVisibleMethodAlipayEnabled *bool   `json:"payment_visible_method_alipay_enabled"`
	PaymentVisibleMethodWxpayEnabled  *bool   `json:"payment_visible_method_wxpay_enabled"`

	// OpenAI account scheduling
	OpenAIAdvancedSchedulerEnabled *bool `json:"openai_advanced_scheduler_enabled"`

	// 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｅΟ娆惧殭缂佺姴鐏氶妵鍕疀閹炬惌妫炵紓浣界堪閸婃洝鐏冮梺鎸庣箓閹冲酣寮抽悙鐑樼厱閻庯絽澧庣弧鈧梺鍝勮嫰缁夌懓鐣烽锕€绀嬫い鎺嗗亾濞寸姵妞藉铏圭磼濡櫣浼囧┑鈽嗗亜鐎氼喗绔熼弴銏犵闁兼祴鏅濋鏇㈡⒑閻熼偊鍤熼柛瀣洴閹偠銇愰幒鎾跺幐闁诲函缍嗛崑鍡樻櫠閻㈠憡鐓欐い鏃傜摂濞堟粓鏌ｅ☉鍗炴珝鐎规洖缍婇、娆撴偂鎼搭喗缍撻梻鍌氬€烽懗鍓佸垝椤栫偑鈧啴宕ㄩ鍥ㄧ☉閳规垿宕卞▎鎰啎濠电偞娼欓崥瀣焽濞嗘挸纭€闁规儼濮ら悡蹇涚叓閸ャ儱鍔ょ痪鎯ф健閺屾稑螣閸忓吋姣勭紓浣介哺閹瑰洤鐣烽幒鎴旀瀻闁瑰瓨绻傞‖澶愭⒒娴ｇ懓顕滈柤褰掔畺閸╃偤骞嬮敂缁樻櫓闂佸吋绁撮弫鈺勵樄闁?
	BalanceLowNotifyEnabled         *bool                   `json:"balance_low_notify_enabled"`
	BalanceLowNotifyThreshold       *float64                `json:"balance_low_notify_threshold"`
	BalanceLowNotifyRechargeURL     *string                 `json:"balance_low_notify_recharge_url"`
	SubscriptionExpiryNotifyEnabled *bool                   `json:"subscription_expiry_notify_enabled"`
	AccountQuotaNotifyEnabled       *bool                   `json:"account_quota_notify_enabled"`
	AccountQuotaNotifyEmails        *[]dto.NotifyEmailEntry `json:"account_quota_notify_emails"`

	// Payment configuration (integrated into settings, full replace)
	PaymentEnabled                   *bool    `json:"payment_enabled"`
	PaymentMinAmount                 *float64 `json:"payment_min_amount"`
	PaymentMaxAmount                 *float64 `json:"payment_max_amount"`
	PaymentDailyLimit                *float64 `json:"payment_daily_limit"`
	PaymentOrderTimeoutMin           *int     `json:"payment_order_timeout_minutes"`
	PaymentMaxPendingOrders          *int     `json:"payment_max_pending_orders"`
	PaymentEnabledTypes              []string `json:"payment_enabled_types"`
	PaymentBalanceDisabled           *bool    `json:"payment_balance_disabled"`
	PaymentBalanceRechargeMultiplier *float64 `json:"payment_balance_recharge_multiplier"`
	PaymentRechargeFeeRate           *float64 `json:"payment_recharge_fee_rate"`
	PaymentLoadBalanceStrat          *string  `json:"payment_load_balance_strategy"`
	PaymentProductNamePrefix         *string  `json:"payment_product_name_prefix"`
	PaymentProductNameSuffix         *string  `json:"payment_product_name_suffix"`
	PaymentHelpImageURL              *string  `json:"payment_help_image_url"`
	PaymentHelpText                  *string  `json:"payment_help_text"`

	// Cancel rate limit
	PaymentCancelRateLimitEnabled *bool   `json:"payment_cancel_rate_limit_enabled"`
	PaymentCancelRateLimitMax     *int    `json:"payment_cancel_rate_limit_max"`
	PaymentCancelRateLimitWindow  *int    `json:"payment_cancel_rate_limit_window"`
	PaymentCancelRateLimitUnit    *string `json:"payment_cancel_rate_limit_unit"`
	PaymentCancelRateLimitMode    *string `json:"payment_cancel_rate_limit_window_mode"`

	// Force Alipay mobile clients to use QR code payment instead of mobile redirect
	PaymentAlipayForceQRCode *bool `json:"payment_alipay_force_qrcode"`

	// Channel Monitor feature switch
	ChannelMonitorEnabled                *bool `json:"channel_monitor_enabled"`
	ChannelMonitorDefaultIntervalSeconds *int  `json:"channel_monitor_default_interval_seconds"`

	// Available Channels feature switch (user-facing)
	AvailableChannelsEnabled *bool `json:"available_channels_enabled"`

	// Affiliate (闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亜顒㈡い鎰矙閺屻劑鎮㈤崫鍕戙垽鏌涚仦璇插闁哄瞼鍠栧鑽も偓闈涘濡差喚绱掗悙顒€鍔ゆい顓犲厴瀵鎮㈤悡搴ｎ槶閻熸粌绻掗弫顔尖槈閵忥紕鍘遍柣搴到婢у海绮旈鈧弻锛勪沪閸撗勫垱婵犵绱曢崗妯绘叏閳ь剟鏌嶉埡浣告殶鐟滄澘妫濆缁樻媴閾忕懓绗￠柣銏╁灣閸犲酣鍩㈤弮鍫濆嵆闁绘ɑ鍓氬鐔兼⒑閸︻厼鍔嬫慨濠呭吹婢规洟宕楅梻瀵哥畾濡炪倖鐗楃换鍐敂閻樿褰掓偂鎼达絿鍔梺? feature switch
	AffiliateEnabled *bool `json:"affiliate_enabled"`

	// 婵犵數濮烽弫鍛婃叏閻戝鈧倿顢欓悙顒夋綗闂佸搫娲㈤崹鍦缂佹绠鹃柟瀛樼懃閻掓椽鏌℃担绋款伃闁哄本绋戦埥澶愬础閻愬吀绮撮梻浣侯焾椤戝棝骞愰幖浣哥叀濠㈣埖鍔曠粻鎶芥煙閹屽殶鐟滄澘娲缁樻媴閼恒儯鈧啰绱掔拋鍦瘈鐎规洘濞婇弫鎰緞婵犲嫮鏋冩繝鐢靛Т閿曘倝鎮ф繝鍥х；闁靛ě鍛紡闂佺粯顨呴悡鍫曞箣閿旇棄鈧爼鏌涢幇銊︽珒缂佽妫濋弻锝夊箛閸忓摜鐩庨梺閫炲苯澧柣妤冨█閻涱噣骞囬婊冧簼闂佸憡鍔戦崝搴ㄦ晬濞戙垺鈷戠紓浣股戦ˉ鍫ユ煛閸涙澘鐨洪柟骞垮灲瀹曞崬鈽夊▎鎴濆箞闂備線娼чˇ浠嬪窗閺嶃劍娅犳い鏍仦閻撶喖鏌熼幆褏鎽犵紒鈧€ｎ喗鐓涚€光偓鐎ｎ剛鐦堥悗瑙勬礃閿曘垺淇婇幖浣肝ㄩ柕蹇曞С婢规洟姊鸿ぐ鎺擄紵缂佲偓娓氣偓瀹曟瑩鎮╁畷鍥╊啎闂佺硶鍓濊摫閻忓繋鍗抽弻锝夊箻鐎涙顦伴梺缁樻惄閸嬪﹤鐣烽崼鏇炍╅柨鏃堟暜閸嬫捇顢橀悢缈犵盎闂婎偄娲﹂幐鐐櫠闁秵鐓涘ù锝呭閸庢棃鏌＄仦璇插闁宠鍨垮畷閬嶅煛閸屽偊绠撳铏圭磼濡櫣鐟ㄩ梺璇茬箲瀹€鍛婁繆閻㈢绀嬫い鏍ㄦ皑椤斿﹪姊洪悷鎵憼缂佽绉电粋鎺楁嚃閳哄啰锛濇繛杈剧稻瑜板啯绂?
	RiskControlEnabled *bool `json:"risk_control_enabled"`

	// cyber 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｉ幇顒佲枙闁绘帟濮ょ换娑㈠幢濡纰嶉柣鐔哥懕缁犳捇骞冨Δ鍛櫜閹肩补鈧尙鏁栭梻浣哥－閹虫挾绮旈悽鍨床婵炴垶锕╅崯鍛亜閺冨洤鍚归柛鎴濈秺濮婅櫣绮欓崸妤娾偓妤呮煥閺囥劋閭€殿喖顭烽弫鎰緞鐎ｎ偅鐝抽梻浣告啞娓氭宕ｆ惔鈾€鏋旀い鎾跺亹閺€浠嬫煟閹邦垰鐨洪柣鎺撳劤闇夋繝濠傜凹闁垶鏌熼鍡欑瘈妞ゃ垺娲熼弫鍐焵椤掑嫭鍊峰┑鐘叉处閻撳繐鈹戦悙鎴濆暞濠€鐗堢箾閸繄绉烘慨濠勭帛閹峰懘宕ㄦ繝鍐ㄥ壍婵＄偑鍊х€靛矂宕归崼鏇犲祦婵°倐鍋撴い顓滃姂瀹曠厧鈹戦崼婵喰曞┑锛勫亼閸婃牜鏁繝鍕偨闁跨喓濮撮悿顕€鐓崶銊р姇闁绘挻娲熼弻锝呂熼崫鍕瘣濠电偞鎯岄崰姘跺焵椤掑喚娼愭繛鍙夛耿瀹曞綊宕稿Δ鍐ㄧウ濠碘槅鍨甸崑鎰閸忛棿绻嗘い鏍ㄧ矊鐢埖顨?+ TTL
	CyberSessionBlockEnabled    *bool `json:"cyber_session_block_enabled"`
	CyberSessionBlockTTLSeconds *int  `json:"cyber_session_block_ttl_seconds"`

	// OpenAI fast/flex policy (optional, only updated when provided)
	OpenAIFastPolicySettings *dto.OpenAIFastPolicySettings `json:"openai_fast_policy_settings,omitempty"`

	// 缂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亜顒㈡い鎰矙閺屻劑鎮㈤崫鍕戙垽鏌涢妸銉モ偓鍨潖婵犳艾閱囬柣鏃€浜介埀顒佸浮閺岋繝宕遍鐘垫殼闂佸搫鐭夌紞浣规叏閳ь剟鏌曢崼婵囶棡闂佹鍙冨娲箹閻愭彃顬夌紓浣筋嚙閸熶即骞戦姀鐘婵妫楅弲鐘差渻閵堝棙顥嗙€规洜鏁婚幆鍕償閿濆洨锛濇繛杈剧稻瑜板啯绂嶆ィ鍐┾拺缂佸娉曢悘閬嶆煕鐎ｎ剙浠遍柟顔欍倗鐤€闁圭虎鍨遍弬鈧梻浣虹帛閸旀浜稿▎鎰浄闁靛繈鍊栭悡鏇㈡煛閸屾繍鍤欓柍褜鍓氶悧婊堝箲閵忕姭鏀介悗锝庝簽閸?platform quota 婵犵數濮烽弫鍛婃叏閻㈠壊鏁婇柡宥庡幖缁愭淇婇妶鍛殲闁哄棙绮嶆穱濠囧Χ閸涱厽娈堕梺娲诲幗閻熲晠寮婚悢鍏煎€绘俊顖炴櫜缁爼姊洪柅鐐茶嫰婢у墽绱掗悩铏碍闁伙綁鏀辩缓鐣岀矙鐠囦勘鍔嶉妵鍕籍閸ヮ灝鎾寸箾閸涱厾孝闁宠鍨块幃鈺呭垂椤愶絾鐦庡┑鐘愁問閸犳岸寮繝姘疇婵犻潧顑呯粈鍫㈡喐瀹ュ＆澶婎潩閼哥數鍘遍梺鎸庢椤曆囩嵁濡绻嗘い鎰靛亜閻忥繝鏌曢崶褍顏鐐达耿瀹曪繝鎮欓棃娑樞ゅ┑鐘殿暯濡插懘宕戦崨鏉戝瀭闁革富鍘介～鏇㈡煙閻戞﹩娈旈柣鎺戠仛閵囧嫰骞掗幋婵愪患缂備讲鍋撻悗锝庡亖娴滄粓鏌″鍐ㄥ闁靛棙甯￠弻娑樜熼悜妯烘殘缂備胶绮粙鎺戭焽韫囨稑绀堢憸蹇氣叢濠电姷鏁搁崑娑㈠箯閹寸姴绶ら柛顭戝櫘閸ゆ洜鎲搁弮鍫濇槬闁逞屽墯閵囧嫰骞掗幋婵冨亾閼姐倕顥氬ù鐘差儐閻撴洟鎮橀悙闈涗壕闁汇劍鍨归埀顒佺⊕缁诲倿鈥旈崘顔嘉ч柛鈩冾殘娴犳潙顪冮妶蹇撶槣闁搞劌鐖奸妴浣割潩閼稿灚娅滄繝銏ｆ硾閿曪箓顢欓幒鎴富闁靛牆妫欓ˉ鍡涙煕鐎ｎ偄濮嶆い銏＄懇瀹曞爼顢楁担鍝勫箥闂備礁鍚嬬粊鎾疾閻愯埖锛傞梻鍌欑窔閳ь剛鍋涢懟顖涙櫠椤斿墽妫紓浣靛灩瀵喚鈧娲﹂崑鍛村箚閺冨牆惟闁挎棁顫夌€氬ジ姊绘担鍝勫付妞ゎ偅娲熷畷鎰板即閻愬灚鎳冩繝鐢靛Х閺佹悂宕戦悙宸劷婵炲棙鎸哥壕褰掓煕椤垵鏋ら柡鍡畵閺屾洝绠涚€ｎ亖鍋撻弽顓炵畾闁绘劗鍎ら悡鏇㈡煥閺冨浂鍤欐鐐村笧缁辨帒螖閳ь剟藝闂堟侗娼栭柧蹇撴贡绾惧吋淇婇婵嗗惞闁告瑥妫濆娲川婵犲懎顥濆銈嗗灥椤︻垶顢氶敐澶婅摕闁靛鍠楅弲銏＄箾鏉堝墽鎮奸柟铏尭椤曪綁宕归瑙勬杸闂佺粯顭囩划顖氣槈瑜旈弻锝呂旈埀顒勬晝椤忓牆鏋侀柟鎯ь吔?= 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｉ幇顒佹儓闁搞劌鍊块弻娑㈩敃閿濆棛顦ョ紓浣哄С閸楁娊寮诲☉妯锋斀闁告洦鍋勬慨銏ゆ⒑濞茶骞楅柟鐟版喘瀵鎮㈤悡搴ｇ暰闂佺粯顨呴悧婊兾涢崟顓犵＝濞达絽鎼暩闂佺顑冮崐婵嗩嚕婵犳艾鍗抽柣鏃囨椤旀洟姊虹紒妯哄Е闁告挻宀搁幃鐢稿籍閸屾粎锛濋梺绋挎湰閻熴劑顢欐径鎰厸闁告侗鍨伴埢鍫⑩偓娈垮枛椤兘寮幇鏉块唶妞ゆ劧鍙婇埡鍛拻濞达絼璀﹂弨鏉库攽椤斿搫鈧宕氶幒鎾村劅闁挎繂鎳庨悘濠囨⒑閸撴彃浜為柛鐘虫尵缁鈻庨幋鐘碉紲婵犮垼娉涢鍡涘磹?nil = 闂傚倸鍊搁崐鎼佸磹閹间礁纾圭€瑰嫭鍣磋ぐ鎺戠倞鐟滄粌霉閺嶎厽鐓忓┑鐐靛亾濞呭棝鏌涢妶鍛伃闁哄本鐩獮姗€鎳犻鈧俊浠嬫倵鐟欏嫭绀€鐎殿喖澧庨幑銏犫攽鐎ｎ偒妫冨┑鐐村灦閼归箖鍩呮潏鈺冪＝濞达絾褰冩禍鐐節閻㈤潧孝婵炶绠撳畷鎰版偨閸涘﹦鍙嗗┑鐘绘涧濡繈顢撳Δ鈧…鑳檨闁搞劌鐖煎璇测槈閵忊€充簻闂佸憡绋戦…鍫熺閵忋倖鈷戦柛婵嗗濠€浼存煟閳哄﹤鐏︾€规洘妞藉畷鐔碱敍濮橀硸妲版俊鐐€曠换鎰板箠鎼粹檧鏋嶆俊銈呭暟绾捐棄霉閿濆洦鍤€闁告柣鍊楅幉鎼佸箥椤旈棿妲愰悗瑙勬穿缁绘繈鐛惔銊﹀癄濠㈣泛瀛╅幉浼存⒒娓氣偓濞佳団€﹂崼銉ョ？妞ゆ柨妲堥敐澶婇唶闁绘梻顭堝鍨攽鎺抽崐鎰板磻閹剧繝绻嗘い鎰剁悼閹冲棝鏌涙繝鍐╁唉婵?	DefaultPlatformQuotas map[string]*service.DefaultPlatformQuotaSetting `json:"default_platform_quotas"`

	AuthSourceEmailPlatformQuotas    map[string]*service.DefaultPlatformQuotaSetting `json:"auth_source_default_email_platform_quotas"`
	AuthSourceLinuxDoPlatformQuotas  map[string]*service.DefaultPlatformQuotaSetting `json:"auth_source_default_linuxdo_platform_quotas"`
	AuthSourceOIDCPlatformQuotas     map[string]*service.DefaultPlatformQuotaSetting `json:"auth_source_default_oidc_platform_quotas"`
	AuthSourceWeChatPlatformQuotas   map[string]*service.DefaultPlatformQuotaSetting `json:"auth_source_default_wechat_platform_quotas"`
	AuthSourceGitHubPlatformQuotas   map[string]*service.DefaultPlatformQuotaSetting `json:"auth_source_default_github_platform_quotas"`
	AuthSourceGooglePlatformQuotas   map[string]*service.DefaultPlatformQuotaSetting `json:"auth_source_default_google_platform_quotas"`
	AuthSourceDingTalkPlatformQuotas map[string]*service.DefaultPlatformQuotaSetting `json:"auth_source_default_dingtalk_platform_quotas"`

	AllowUserViewErrorRequests *bool `json:"allow_user_view_error_requests"`
}

// UpdateSettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犵厱婵﹩鍘介妵婵嗏攽闄囬崺鏍ь嚗閸曨厸鍋撻敐搴濈胺婵″弶鍔欏缁樼瑹閳ь剙顭囪婢ф繈姊洪崫鍕櫤闁烩晩鍨堕獮蹇涘箣閿旇棄浜滈柣蹇撶箣閻掞箓寮埀顒勬⒒娴ｈ櫣甯涢拑杈╂喐閺夊灝鏆ｆい銏℃閸╋繝宕ㄩ灏栧亾閻㈠憡鐓曟い顓熷灥娴滅偤鏌涘▎灞戒壕闂傚倷绀侀幖顐ゆ偖椤愶箑纾归柡宥庡幖缁犵偤鏌曟繛鐐珔缂佺姵绋掗妵鍕箣閿濆棭妫勬繛瀛樼矌閸嬫挻绌辨繝鍥ч柛娑卞枛濞咃綁姊洪崫鍕仴闁稿海鏁诲濠氭晲閸涘倻鍠栭幊鏍煛娴ｄ警鍋ч梻鍌欒兌椤牓顢栭幋鐘茬筏濞寸姴顑嗙粻鎺楁⒒娴ｇ懓顕滅紒璇插€搁湁闁稿瞼鍋為崐鍨旈敐鍛殲闁绘挸鍟村娲垂椤曞懎鍓卞銈冨劚閻楁捇寮?
// PUT /api/v1/admin/settings
func (h *SettingHandler) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	previousSettings, err := h.settingService.GetAllSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	previousAuthSourceDefaults, err := h.settingService.GetAuthSourceDefaultSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 婵犵數濮烽弫鍛婃叏閹绢喗鍎夊鑸靛姇缁狙囧箹鐎涙ɑ灏ù婊呭亾娣囧﹪濡堕崟顓炲闂佸憡鐟ョ换姗€寮婚悢铏圭＜闁靛繒濮甸悘宥夋⒑缁嬪灝顒㈡い銊ユ嚇婵℃挳骞掗幋顓熷兊闂佹寧绻傞幊宥嗙珶閺囥垺鈷掑ù锝囩摂閸ゅ啴鏌涢…鎴濈仸闁诡喚鍋ら弫鍐焵椤掑嫷鏁嬮柕澶嗘櫅缁€瀣亜閺嶃劍鐨戦柣銈傚亾濠碉紕鍋戦崐鏍暜閹烘纾瑰┑鐘宠壘閸屻劍銇勯幒鎴濐仾闁抽攱甯掗湁闁挎繂瀚鐔兼煟閹烘鐣洪柡宀嬬節瀹曞ジ顢曢姀鐙€娼剧紓鍌欑贰閸犳鎮烽埡鍛祦闁圭儤顨呯粻锝夋煛閸愶絽浜鹃梺?
	if req.DefaultConcurrency < 1 {
		req.DefaultConcurrency = 1
	}
	if req.DefaultBalance < 0 {
		req.DefaultBalance = 0
	}
	affiliateRebateRate := previousSettings.AffiliateRebateRate
	if req.AffiliateRebateRate != nil {
		affiliateRebateRate = *req.AffiliateRebateRate
	}
	if affiliateRebateRate < service.AffiliateRebateRateMin {
		affiliateRebateRate = service.AffiliateRebateRateMin
	}
	if affiliateRebateRate > service.AffiliateRebateRateMax {
		affiliateRebateRate = service.AffiliateRebateRateMax
	}
	affiliateRebateFreezeHours := previousSettings.AffiliateRebateFreezeHours
	if req.AffiliateRebateFreezeHours != nil {
		affiliateRebateFreezeHours = *req.AffiliateRebateFreezeHours
	}
	if affiliateRebateFreezeHours < 0 {
		affiliateRebateFreezeHours = service.AffiliateRebateFreezeHoursDefault
	}
	if affiliateRebateFreezeHours > service.AffiliateRebateFreezeHoursMax {
		affiliateRebateFreezeHours = service.AffiliateRebateFreezeHoursMax
	}
	affiliateRebateDurationDays := previousSettings.AffiliateRebateDurationDays
	if req.AffiliateRebateDurationDays != nil {
		affiliateRebateDurationDays = *req.AffiliateRebateDurationDays
	}
	if affiliateRebateDurationDays < 0 {
		affiliateRebateDurationDays = service.AffiliateRebateDurationDaysDefault
	}
	if affiliateRebateDurationDays > service.AffiliateRebateDurationDaysMax {
		affiliateRebateDurationDays = service.AffiliateRebateDurationDaysMax
	}
	affiliateRebatePerInviteeCap := previousSettings.AffiliateRebatePerInviteeCap
	if req.AffiliateRebatePerInviteeCap != nil {
		affiliateRebatePerInviteeCap = *req.AffiliateRebatePerInviteeCap
	}
	if affiliateRebatePerInviteeCap < 0 {
		affiliateRebatePerInviteeCap = service.AffiliateRebatePerInviteeCapDefault
	}
	// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亜顒㈡い鎰矙閺屻劑鎮㈤崫鍕戙垻鐥幆褜鐓奸柡灞剧☉閳藉宕￠悙鎻掝劀婵犵數鍋涢惃鐑藉疾閻樿钃熼柡鍥╁枔缁犻箖鏌ｉ幇闈涘闁绘繃妫冨娲传閸曢潧鍓紓浣藉煐瀹€绋款嚕婵犳艾鍗抽柕蹇曞█閸炶泛鈹戦悩缁樻锭婵炲眰鍊濊棟闁冲搫鎳忛埛鎴犵磼鐎ｎ偒鍎ラ柛搴㈠姍閺屻劑寮村Ο铏逛紙閻庢鍠栭…宄邦嚕閹绢喗鍋勯柧蹇撴贡濡插洭姊绘繝搴′簻婵炶绠撻獮鏍煛娴艰鲸妞介、姗€鎮㈤搹鍦闂備線娼ф蹇曠礊閸℃鐒介柟鎵閻撴洟鏌￠崘锝呬壕闂佽崵鍠嗛崕鐢告晲閻愬墎鐤€闁瑰彞鐒﹀浠嬨€侀弮鍫濈妞ゆ挶鍨诲畷婊堟⒒娴ｈ棄鍚瑰┑顔肩仛缁傚秵绂掔€ｎ亞顦梻渚囧墮缁夋潙娲垮┑鐘灱濞夋盯顢栭崶顒€鍌ㄥù鐘差儐閳锋垿姊婚崼鐔恒€掑褎娲熼弻鐔煎礃閼碱剛顔掗柦妯煎枛閺屾洝绠涚€ｎ亖鍋撻弽顓熷亗婵炴垶鈼よぐ鎺撴櫜闁告侗鍠楅崰鎰版倵濞戞瑧绠撴い顏勫暣婵¤埖鎯旈垾鑼嚬缂傚倷娴囬褔鎮ч幘璇参ュù锝堝€介弮鍫濆窛妞ゆ挾濮存慨锔戒繆閻愵亜鈧牜鏁幒鏂哄亾濮樼厧澧扮紒顔碱煼瀵粙顢橀悢鍝勫箥闂佸搫顦悧鎾剁不閺嵮岀€舵い鏇楀亾闁哄本绋掔换婵嬪礋椤掍焦鐦庨梻浣圭湽閸庣儤绂嶉鍫涒偓浣割潨閳ь剟骞冨▎鎾崇骇闁规惌鍘搁崑鎾斥槈濡攱鏂€闂佺粯鍔栬ぐ鍐箖閹达附鐓曢柡鍐ｅ亾闁绘顨堢划姘綇閵娧呯槇闂佹悶鍎撮崺鏍疾濠靛棭娓婚柕鍫濇婵附绻濋埗鈺佷壕缂傚倷闄嶉崝瀣垝濞嗗浚娼栨繛宸簼閻掑鏌ｉ幇顖氳敿閻庢碍婢橀…鑳檨闁哥姵姘ㄥΣ鎰板箳濡や礁浜归梺鑲┾拡閸庣柉顦撮柍褜鍓濋～澶娒哄鍫濈疇闁规崘娉涘鍙変繆閻愵亜鈧洜鎹㈤幇顔瑰亾濮樼厧骞栭崡杈ㄣ亜閹哄棗浜鹃梺瀹狀潐閸ㄥ潡骞冨▎鎾崇骇闁瑰濮冲鎾寸節濞堝灝鏋熼懣褔鏌涢弮鈧崹鐢告偩閻戣棄鐏抽柟棰佺劍閸嶉潧顪冮妶鍡樺鞍缂侇喖鐗忛懞杈ㄧ節濮橆厸鎷洪梺鍛婄箓鐎氼喛鈪归梻浣告啞閺屻劑鏌婇敐澶婃槬濠电姴娲ら崡鎶芥煟濮楀棗浜滃ù婊呭亾缁绘盯宕煎┑鍫滆檸闂佸搫顑囬崰鏍蓟閻旂厧宸濆┑鐘插暙閳ь剚鍔栫换婵嬪焵椤掍胶鐟归柍褜鍓欓～蹇撁洪鍕槯闂佸吋绁撮弬鍌炲磻閻愮儤鈷戦柛婵嗗閸ｈ櫣绱掔€ｎ偄鐏撮柟顔芥そ椤㈡﹢濮€閳锯偓閹风粯绻涙潏鍓у埌闁圭⒈鍋婂畷婵嬫晝閳ь剟鍩為幋锔绘晩闁兼祴鏁╄閳ь剝顫夊ú姗€鏁冮姀銈呮瀬闁圭増婢橀獮銏＄箾閸℃鎹ｉ柟绋款樀濮婂宕掑顑藉亾妞嬪海鐭嗗ù锝呮贡閻濊泛鈹戦悩鍙夊闁稿鍊块弻娑㈩敃閻樻彃濮庣紒鐐劤椤兘寮诲☉妯锋婵炲棙锚婵′即姊洪幖鐐插姶濞存粏娉涘嵄濠电姵纰嶉悡鐘绘煙缁嬪潡顎楀┑顔肩Ч閺岋紕浠﹂崜褎鍒涙繝纰夌磿閸忔﹢寮幇鏉垮窛妞ゆ挆鍛亾妞嬪海纾介柛灞剧懆閸忓苯鈹戦鐐毈闁轰礁鍟存慨鈧柕鍫濇嚀閹芥洟姊虹捄銊ユ灁濠殿喗鎸抽悰顕€濮€閵堝洤褰勯梺鎼炲劘閸斿秶浜搁銏╃唵閻熸瑥瀚粈瀣叏婵犲啯銇濋柟绛圭節婵″爼宕ㄩ閿亾閸撗呯＝濞撴艾娲ら弸娑㈡倵濞戞帗娅婄€殿喛灏欓幑鍕媴閺囩喐顥堢€规洏鍔戦、姗€鎮埀顒勫船閸洘鈷掑ù锝呮啞閸熺偟绱掔€ｎ偄鐏撮柡浣稿暣婵＄兘濡烽姀锛勨偓娲⒑缂佹﹩鐒界紒顕呭灦閹繝鎮㈤梹鎰畾濡炪倖鐗楁笟妤呭磿閵夆晜鐓曢柕鍫濇閹冲懘鏌ｉ妷顔婚偗妤犵偛妫滈ˇ鎶芥偣閹板墎纾跨紒杈ㄥ笚濞煎繘濡搁妷褏鎳嗛梺鍓х帛閻楃娀寮诲☉銏犲嵆婵°倓鐒﹀鎺楁⒑缁嬫鍎愰柟鐟版喘瀹曟椽鍩€椤掍降浜滈柟鐑樺灥椤忣亪鏌?
	if req.TableDefaultPageSize <= 0 {
		req.TableDefaultPageSize = previousSettings.TableDefaultPageSize
	}
	if req.TablePageSizeOptions == nil {
		req.TablePageSizeOptions = previousSettings.TablePageSizeOptions
	}
	req.SMTPHost = strings.TrimSpace(req.SMTPHost)
	req.SMTPUsername = strings.TrimSpace(req.SMTPUsername)
	req.SMTPPassword = strings.TrimSpace(req.SMTPPassword)
	req.SMTPFrom = strings.TrimSpace(req.SMTPFrom)
	req.SMTPFromName = strings.TrimSpace(req.SMTPFromName)
	if req.SMTPPort <= 0 {
		req.SMTPPort = 587
	}
	req.DefaultSubscriptions = normalizeDefaultSubscriptions(req.DefaultSubscriptions)
	req.AuthSourceDefaultEmailSubscriptions = normalizeOptionalDefaultSubscriptions(req.AuthSourceDefaultEmailSubscriptions)
	req.AuthSourceDefaultLinuxDoSubscriptions = normalizeOptionalDefaultSubscriptions(req.AuthSourceDefaultLinuxDoSubscriptions)
	req.AuthSourceDefaultOIDCSubscriptions = normalizeOptionalDefaultSubscriptions(req.AuthSourceDefaultOIDCSubscriptions)
	req.AuthSourceDefaultWeChatSubscriptions = normalizeOptionalDefaultSubscriptions(req.AuthSourceDefaultWeChatSubscriptions)
	req.AuthSourceDefaultDingTalkSubscriptions = normalizeOptionalDefaultSubscriptions(req.AuthSourceDefaultDingTalkSubscriptions)

	// SMTP 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮村Δ鈧禍鎯ь渻閵堝骸骞栭柣蹇旂箚閻忔帡姊洪崗鑲┿偞闁哄懏绻堥幃锛勨偓锝庡枟閳锋垿鏌涢幘鏉戠祷濞存粍绻冪换娑㈠矗婢舵稖鈧法鈧娲橀崹鍧椼€侀弮鍫濋唶闁绘柨鎼獮鍫ユ⒒娴ｈ鐏遍柡鍛洴瀹曟澘顫濋懜闈涒偓鍨归悩宸剱闁绘挸鍟撮弻锝夋偄缁嬫妫嗛柡宥忕磿缁辨挻鎷呴搹鐟扮缂備浇顕ч悧鍡氱亱闂佸憡鍔戦崝澶娢ｉ崼鐔虹闁糕剝锚婵牓鏌涙繝鍕幋闁哄矉缍侀幃娆戔偓闈涘閺嗭繝姊洪幐搴㈢８闁搞劌缍婇敐鐐剁疀閹句焦妞介、鏃堝礋椤撗冩暪闂備胶顢婃竟鍫ュ箵椤忓棗绶ら柣鎴炆戦崣蹇涙煕閵夘喖澧柍閿嬪灴閺岋綁骞橀搹顐ｅ闯缂備礁顦冲▍锝囨閹烘鍋愮€规洖娲ら埛灞剧節绾版ê澧查柟绋垮暱閻ｇ兘濡搁埡鍌滃姺闂佹寧妫佽闁哄鐭傚铏规嫚閹绘帩鍔夊銈嗘⒐閻楃姴鐣烽弶娆炬僵闁搞劍绋堥崑鎾绘晝閸屾岸鍞堕梺闈涱槴閹冲洭宕戦幘缁樻櫇闁稿本顕遍埡鍛厓閺夌偞澹嗛ˇ锕傛煠閺夋寧鍋ユ慨濠呮閳ь剙婀辨慨鐢稿Υ閸愵喗鐓熼柟铏瑰仧閻ｅ灚銇勯姀锛勬噭鐎垫澘瀚埀顒婄秵閸撴盯鎯侀崼婵冩斀闁绘劘灏欐晶鏇㈡煟韫囨梻绠炵€殿喗鎮傛俊鑸靛緞鐎ｎ剙骞嶉梻浣瑰劤濞存岸宕戦崨顓犳殾鐎光偓閳ь剟鍩€椤掑喚娼愭繛鍙夛耿瀹曟繂鈻庨幘宕囩暫濠电姴锕ら悧濠囧吹瀹ュ鐓忓璺虹墕閸旀粓鏌?smtp_host 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｉ幇顒佹儓闁搞劌鍊块弻娑㈩敃閿濆棛顦ョ紓浣哄У瑜板啴婀侀梺鎸庣箓濞层劑濡存繝鍥ㄧ厸閻庯綆浜炴晥闂佸搫琚崝宀勫煘閹达箑骞㈤柍鍝勫€愰敃鍌涒拺缂備焦蓱鐏忎即鏌ㄩ弴妯虹伈鐎殿喖顭烽崹楣冨箛娴ｅ憡鍊梻浣告啞閸旀垿宕濆畝鍕ㄢ偓鏍ㄥ緞鐎ｎ剛鐦堝┑鐐茬墕閻忔繈鎮橀敓鐘崇厵闁告稑锕ら埢鏇犫偓娈垮枦椤曆囧煡婢舵劕顫呴柣姗€娼ф慨鍫曟⒒娴ｄ警鐒剧紒璇叉瀵彃顭ㄩ崼婵嗙€梺姹囧灮椤牓鎮″┑瀣厵闁硅鍔﹂崵娆撴⒑閸楃偞鍠橀柡宀嬬節瀹曞爼鍩℃担鍦簴闂備礁鎲￠…鍥极婵犳艾钃熼柨婵嗩槹閸嬫劙鏌涘▎蹇ｆШ妞わ箓浜跺娲焻閻愯尪瀚板褍寮剁换娑㈠川椤旂晫顦伴梺璇″枟閿曘垹鐣锋總绋款潊闁靛浚婢佺槐鍐测攽閻愯埖褰х紓宥勭椤洩顦抽柟渚垮姂閸┾偓妞ゆ帊妞掔换鍡涙煟閹邦剛鎽犵紓宥嗗灴閺岀喖鎳為妷褏鐓傞梺鍛婂笚鐢帡鍩㈡惔銊ョ疀妞ゆ帒鍊婚悰顕€姊绘担绛嬫綈婵犮垺锕㈤幃妯侯潩鐠鸿櫣鏌ч梺鎸庣箓椤︿即鍩涢幋鐘电＝濞达絽顫栭鍛弿闁割偁鍨洪崰鎰版煛婢跺﹦浠㈤柡鍡樼懇閺屽秶绱掑Ο璇茬３婵犵绱曢崗姗€宕洪敓鐘茬＜婵°倓绀侀褰掓⒒閸屾瑧顦﹂柟璇х磿閹广垽宕掑┃鎯т壕婵鍘ч弸鐔兼煛閸涱厾鍩ｇ€殿喗鎸抽幃銏㈢礄閻樼數娉块梻鍌欑劍鐎笛冾潖婵犳艾纾婚柟鎹愵嚙濮瑰弶銇勯幒鎴濃偓鐢稿磻閹捐埖鍠嗛柛鏇ㄥ墰閿涙﹢姊洪幖鐐插濠殿喓鍊栫粚杈ㄧ節閸パ囧敹闂佸搫娲ㄩ崑鐔煎磹閻愮儤鈷戦柛鎰级閹牓鏌涢悩宕囥€掓俊鍙夊姉閳ь剚绋掔湁缂佽妫濋弻锝夊箛閸忓摜鐩庨梺閫炲苯澧柣妤佹礈閸欏懘姊洪崫鍕枆闁诲繑绋戦埥澶娢熼柨瀣垫綌婵犳鍠楅…鍫熺椤掑嫭鍊块柛蹇氬亹缁♀偓闂侀潧楠忕徊浠嬫偂閹扮増鐓曢柡鍐ｅ亾闁绘濞€楠炲啴鏁撻悩鎻掑祮闂侀潧楠忕槐鏇㈠储閻㈠憡鈷戠紓浣姑悘锕傛嫅闁秵鍊堕煫鍥ㄦ礃閺嗩剟鏌″畝鈧崰鏍х暦濡ゅ懎宸濋柡澶婄仢鐢垳绱撻崒娆愮グ濡炴潙鎽滈弫顕€鎮欓崹顐綗闂佸湱鍎ら崵锕傛偄閻撳簼绱堕梺鍛婃处閸撴岸宕甸崼銉︹拻闁稿本鐟︾粊鐗堛亜閺囩喓澧电€规洘婢樿灃闁告侗鍘鹃敍鐔兼⒑鐟欏嫬鍔舵俊顐㈠閹€斥攽鐎ｎ亞顔愬┑鐑囩秵閸撴瑩鍩€椤掍緡娈滈柟顔兼健閸┾偓妞ゆ帒瀚埛鎴︽偡濞嗗繐顏╃紒鈧崘鈺冪闁告侗鍠楃粈瀣亜閵忊埄鎴犵紦娴犲宸濆┑鐐靛亾鐎?SMTP 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮村Δ鈧禍鎯ь渻閵堝骸骞栭柣蹇旂箚閻忔帡姊洪崗鑲┿偞闁哄懏绻堥幃锛勨偓锝庡枟閳锋垿鏌涢幘鏉戠祷濞存粍绻冪换娑㈠矗婢舵稖鈧法鈧?
	// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌熼梻瀵割槮缁炬儳缍婇弻鐔兼⒒鐎靛壊妲梺姹囧€ら崰妤呭Φ閸曨垰绠涢柍杞拌兌閸旀挳姊洪幖鐐测偓鏍洪悢鐓庤摕婵炴垯鍩勯弫鍐煥濠靛棙顥滄い锔规櫊濮婄儤娼幍顕呮М濠碘槅鍋呴悷褔骞戦姀鐘闁靛繒濮烽鍝勨攽閻愬弶顥滅紒缁樺笚缁傛帡鎳栭埡鍐紳婵炶揪绲捐ぐ鍐╃妤ｅ啯鈷戦柛鎾村絻娴滄劙鏌熼崘鍙夊枠鐎规洘鍨块獮姗€鎳￠妶鍛偊闂備礁澹婇悡鍫ュ窗閺嶃劍鍙忛幖绮规閺€浠嬫煟閹邦剛鎽犵紓宥嗗灴閺屾稒绻濋崒娑滅闂侀€炲苯澧存繛浣冲洤围缂佸娉曢弳锕傛煙閻戞﹩娈曢柛濠勫厴閹鎮藉▓璺ㄥ姼濡ょ姷鍋涢悧鎾愁潖缂佹ɑ濯撮柛娑橈攻閸庢捇姊洪崫銉ユ珡闁稿锕ら悾鐑藉箛閺夎法顔愭繛杈剧秬閵嗏偓闁稿鎹囬弻銊р偓锝庡墰閻﹀牓姊哄Ч鍥х伈婵炰匠鍕浄闁挎梻鏅粻楣冩煕韫囨艾浜归柟鍐插暣閹繝濡舵径瀣帾婵犵數鍋熼崑鎾斥枍閸涱垳纾奸柣娆屽亾闁搞劌鐖奸獮鍐ㄎ旈崨顔芥珳闁硅偐琛ラ崜婵嬫倶閸垻纾藉ù锝嗗絻娴滈箖姊洪崜鎻掍簴闁搞劍妞藉畷浼村箛閹殿喖褰勯梺鎼炲劘閸斿酣鍩ユ径鎰厱闁瑰濮撮埢鍫ユ煛鐏炲墽銆掑ù鐙呯畵楠炲棜顦叉慨锝呯墦濮婅櫣绱掑Ο鑽ゅ弳濠碘槅鍋呯换鍫濐嚕婵犳碍鍋勭痪鎷岄哺閺咁剙鈹戦悙鏉戠仸闁挎洍鏅涜灋闁糕剝鐟х壕钘壝归敐澶嬫锭濠碘€冲悑缁绘稓娑垫搴ｇ槇濡ょ姷鍋涢崯顐︽偩閿熺姴绠ラ柧蹇ｅ亞閳ь剝顕ч埞鎴︽倷閸欏妫￠梺鎼炲妼缂嶅﹤顕ｉ幎钘壩у璺侯儏閳ь剙鐏氱换娑㈠醇濠靛牅铏庡┑鐐叉噺閿曘垽寮诲☉銏℃櫜闊洦娲栭崺宀勬⒑娴兼瑧鍒伴柛銏＄叀閳ワ箓濡搁埡浣侯槹濡炪倖宸婚崑鎾淬亜鎼淬垺宕屾慨濠冩そ瀹曘劍绻濋崟顒€娅戞俊鐐€ч梽鍕熆濮椻偓閸┿垹顓兼径濠囧敹闂侀潧顧€閼靛綊骞忓ú顏呪拺闁绘挸瀵掑鐔兼煕婵犲啯绀冪紒鍌涘浮閺佸倿宕滆閿涙粓鏌ｆ惔顖滅У濞存粎鍋熺划濠氭嚒閵堝倸浜鹃悷娆忓缁€鍐磼鐠囨彃顏紒鍌涘浮閺佸啴宕掑☉妯规偅闂佽绻掗崑鐔煎磻閹炬椿鏁婇柡鍥ュ灪閳锋帒霉閿濆牊顏犻悽顖涚⊕閹便劍绻濋崒銈囧悑閻庤娲樺ú鐔镐繆閼搁潧绶炲┑鐘插閻掔偓绻濋悽闈涗粶婵☆垰锕ョ粋宥呪攽鐎ｎ偄浜楅梺鍝勬储閸ㄦ椽鎮￠崘顔界厓閺夌偞澹嗛ˇ锕傛煛閸℃瑥鏋涢柡宀€鍠栧畷姗€骞撻幒婵呯磻闂備焦鐪归崕鐑樼椤忓牆绠栨繛鍡樻惄閺佸棝鏌嶈閸撶喕妫㈤柣搴秵閸犳鎮￠崘顏呭枑婵犲﹤鐗嗙粈鍫熺節闂堟侗鍎滅紓宥嗙墪椤潡鎳滈棃娑橆潔闂佺顑冮崐婵嬪蓟閿熺姴鐐婇柍杞扮婢瑰姊洪幖鐐测偓鏍洪悢鐓庤摕闁绘梻鍘х粻鏌ユ煙闁箑澧伴柛姗嗗亰閹鎲撮崟顒傤槬閻庢鍠栭悥濂哥嵁閸儱惟闁靛娴烽崰鎾诲箯閻樺樊鍟呮い鏃囧Г椤旀垿姊婚崒娆掑厡妞ゎ厼鐗撻、鏍幢濞戞顔囬梺褰掓？缁€渚€鎮￠垾鎰佺唵閻犲搫褰块崼銉ュ嚑閹兼惌婢€缁诲棙銇勯弽銊х煀閻㈩垵娉涢埞鎴︻敊閹冨缂?SMTP 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮村Δ鈧禍鎯ь渻閵堝骸骞栭柣蹇旂箚閻忔帡姊洪崗鑲┿偞闁哄懏绻堥幃锛勨偓锝庡枟閳锋垿鏌涢幘鏉戠祷濞存粍绻冪换娑㈠矗婢舵稖鈧法鈧?
	if req.SMTPHost == "" && previousSettings.SMTPHost != "" {
		req.SMTPHost = previousSettings.SMTPHost
		req.SMTPPort = previousSettings.SMTPPort
		req.SMTPUsername = previousSettings.SMTPUsername
		req.SMTPFrom = previousSettings.SMTPFrom
		req.SMTPFromName = previousSettings.SMTPFromName
		req.SMTPUseTLS = previousSettings.SMTPUseTLS
	}

	// Turnstile 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柛鎰紦閻㈠姊绘笟鈧褔藝椤愶箑鐤炬繛鎴炶壘椤ユ岸鏌涢敂璇插箺闁哥姵鍔欓弻锝呂旈埀顒勬偋閸℃瑧绠旈柟鐑樻尰閸欏繘鎮峰▎蹇擃伀闁告瑢鍋撻梻浣告惈閻绱炴担鍓插殨妞ゆ帒瀚崹鍌涖亜閺囩偞鍣烘い銉﹀灦缁绘繈鎮介棃娑楃捕闂佺粯顨呯换妯侯嚕閺屻儺鏁冮柨鏇楀亾闁绘帒鐏氶妵鍕箳閸℃ぞ澹曢梻浣告啞钃辩紒顔芥尭閻ｇ兘骞嬮敃鈧粻鑽ょ磽娴ｈ鐒介柛姗€浜跺铏规喆閸曨剛鍑￠梺鍛婂焹閸嬫挾绱撴担闈涘闁告艾顑夋俊?
	if req.TurnstileEnabled {
		// 濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴濐潟閳ь剙鍊块幐濠冪珶閳哄绉€规洏鍔戝鍫曞箣閻欏懐骞㈤梻鍌欑窔閳ь剛鍋涢懟顖涙櫠閹绢喗鐓涢悘鐐插⒔閵嗘帡鏌嶈閸撱劎绱為崱娑樼；闁告侗鍘鹃弳锔锯偓鍏夊亾闁告洦鍓涢崢鎼佹倵閸忓浜鹃柣搴秵閸撴盯鏁嶉悢鍝ョ閻庣數顭堥鎾斥攽閳ヨ櫕鍠樻鐐茬箻閹晝鎷犻懠顒夊斀闂備礁婀遍崕銈夊春閸繍鐒芥繛鎴炴皑绾句粙鏌涚仦鎹愬闁逞屽墰閸忔﹢骞婂Δ鍛唶闁哄洨鍋熼敍鐔兼⒑濮瑰洤鐏い顓炵墢婢规洘绻濆顓犲幍闂佽鍨庣仦鑺ヮ啀缂傚倷璁插褔宕戦幘缁樷拻闁稿本鑹鹃埀顒佹倐閹勭節閸愵亞鎳濆┑掳鍊愰崑鎾绘懚閿濆鐓曟い鎰Т閸旀粓鏌ｉ幘瀛樼闁诡喗顨婇弫鎰償濠靛牏娉挎俊鐐€曞ù姘跺矗閸愵喖钃熼柣鏃傗拡閺佸鏌嶈閸撴稓妲愰幘鎰佹僵缂佹棏鍨板ú顓烆嚕婵犳艾唯闁挎梹鍎崇敮?
		if req.TurnstileSiteKey == "" {
			response.BadRequest(c, "Turnstile Site Key is required when enabled")
			return
		}
		// 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柛娑橈攻閸欏繘鏌ｉ幋婵愭綗闁逞屽墮閸婂湱绮嬮幒鏂哄亾閿濆簼绨介柛鏃撶畱椤啴濡堕崱妤€娼戦梺绋款儐閹瑰洭寮诲☉銏″亜闂佸灝顑呮禒鎾⒑缁洘鏉归柛瀣尭椤啴濡堕崱妤€娼戦梺绋款儐閹稿墽妲愰幘鎰佸悑闁告粌鍟抽崥顐⑽旈悩闈涗粶闁哥噥鍋夐悘鎺楁煟閻樺弶绌块悘蹇旂懅缁綁鎮欓悜妯锋嫼閻熸粎澧楃敮鎺撶娴煎瓨鐓曢柟鎯ь嚟閹冲洭鏌熼鈧褑鐏掓繛鎾村嚬閸ㄨ京鈧潧鐭傚娲濞戞艾顣洪梺纭呮珪閸旀鍒掔紒妯侯嚤閻庢稒顭囬崢鐢告⒑閸涘﹤濮囩€殿喖鐖奸獮鎴︽晲閸ワ絽浜鹃柛顭戝亝缁舵煡鏌ㄩ弴銊ら偗妤犵偛鍟撮幃婊堟嚍閵夈儲鐤傛俊鐐€栭崹鐓幬涢崟顒傤洸?secret key闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢琛″亾閻㈡鐒惧ù鐘欏洦鈷掗柛鏇ㄥ亜椤忣參鏌″畝瀣瘈鐎规洘锕㈤弫鎰板川椤掆偓椤ユ氨绱撻崒姘偓鐑芥嚄閸撲礁鍨濇い鏍仦閸庡孩銇勯弽銊︾殤闁哄棴绠撻弻娑㈩敃閿濆棗顦╅梺杞扮濡婀侀梺鎸庣箓椤︽澘鈻嶅鍛＝闁稿本绋掑畷灞炬叏婵犲嫮甯涚紒妤冨枛閸┾偓妞ゆ巻鍋撴い顓炴穿椤﹀綊鏌嶉妷顖滅暤鐎规洖銈告俊鐑藉Ψ瑜濈槐鐢告⒒閸屾瑧璐伴柛鎾寸懅缁棃骞橀鑹靶曢悷婊呭鐢鎮″▎鎾寸厽闁瑰浼濋鍕ㄦ灁闁哄被鍎辩粻瑙勭箾閿濆骸澧┑鈥炽偢閺岋綁骞囬濠呭惈闂佸搫鏈ú妯兼崲濞戞粠妲婚梺纭呮珪閸旀瑥顕ｉ幎钘壩у璺侯儌閹风粯绻涙潏鍓у埌濠㈢懓锕畷銉╁籍閸喓鍘告繝銏ｆ硾閿曪附绂掗姀掳浜滄い鎰╁灪閸犳ɑ鎱ㄦ繝鍛仩闁归濞€閸ㄦ儳鐣烽崶锕€鎽嬮梻鍌欑閹芥粓宕伴幇顒夌劷婵炲棙鎸婚崑鈺呮煟閹达絽袚闁稿瀚惀顏堝箯瀹€鍕懙闂佸搫鎳忕换鍫濐潖閾忓湱纾兼俊顖涙た濮婂灝鈹戦悙瀛樺磩妞ゎ偄顦辩划鈺呮偄閻撳骸宓嗛梺缁橆焽閺佹悂鏁嶅鍛＝闁稿本鐟╁鐑芥煕閵娧勫殌妞ゆ洩缍侀、娆愭叏閹邦亞鐩庨梻渚€娼ч¨鈧梻鍕娴滄悂鎮介悽鐢碉紲濡炪倖妫侀崑鎰櫠閿旈敮鍋撶憴鍕闁哥姵鐗犻妴浣糕槈濡攱顫嶅┑顔斤供閸撴瑧绮诲ú顏呪拻闁稿本鐟ч崝宥夋倵缁楁稑鎳忓畷鏌ユ煕鐏炵虎鍤ゆ繛鎴炃氬Σ鍫ユ煏韫囨洖啸闁汇倕娲娲川婵犲嫧妲堥梺瀹︽澘濮傞柟顕嗙節瀹曟﹢顢欓挊澶夌敾婵犵數鍋涘Λ妤冩崲閹伴偊鏁傛い鎰╁焺閻斿棛鎲稿鍫濈婵炴垶纰嶉～鏇㈡煙閻戞ɑ鐓ｆ繛宀婁邯閺岋箑螣娓氼垱歇闂佺硶鏅濋崑銈咁潖閾忓厜鍋撻崷顓烆€屾繛鍏煎姍閺屾盯濡搁妷褌铏庨梺浼欑秮閺€杈╃紦娴犲宸濆┑鐘插€风花濠氭⒒娴ｅ憡鍟炵紒鍝勬健瀹曟洟骞撻幒鍡樼亖濡炪倖娲嶉崑鎾存叏婵犲啯銇濋柟绛圭節婵″爼宕ㄩ閿亾閸撗呯＝濞撴艾娲ら弸娑㈡倵濞戞帗娅婄€殿喛灏欓幑鍕媴閺囩喐顥堢€规洏鍔戦、姗€鎮埀顒勫船閸洘鈷掑ù锝呮啞閸熺偟绱掔€ｎ偄鐏撮柡浣稿暣婵＄兘濡烽姀锛勨偓娲⒑缂佹﹩鐒界紒顕呭灦閹繝鎮㈤梹鎰畾濡炪倖鐗楁笟妤呭磿閵夆晜鐓曢柕鍫濇閹冲懘鏌ｉ妷顔婚偗妤犵偛妫滈ˇ鎶芥偣閹板墎绡€闁哄矉缍佹俊鐑筋敊婵犳艾浠愭俊鐐€戦崹娲晝閵忊剝鍙忛柍褜鍓熼弻宥夊传閸曨偂绨藉┑鈽嗗亝閻楁粎妲愰幘璇茬＜婵炲棙甯╅崬褰掓⒑?
		if req.TurnstileSecretKey == "" {
			if previousSettings.TurnstileSecretKey == "" {
				response.BadRequest(c, "Turnstile Secret Key is required when enabled")
				return
			}
			req.TurnstileSecretKey = previousSettings.TurnstileSecretKey
		}

		siteKeyChanged := previousSettings.TurnstileSiteKey != req.TurnstileSiteKey
		secretKeyChanged := previousSettings.TurnstileSecretKey != req.TurnstileSecretKey
		if siteKeyChanged || secretKeyChanged {
			if err := h.turnstileService.ValidateSecretKey(c.Request.Context(), req.TurnstileSecretKey); err != nil {
				response.ErrorFrom(c, err)
				return
			}
		}
	}

	// TOTP 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柛鎰紦閻㈠姊绘笟鈧褔藝椤愶箑鐤炬繝濠傜墕閸氬綊鏌熼柇锕€骞戝ù婊勭矒閺屸€愁吋鎼达絽甯ラ梺鎼炲€栧Λ鍐蓟濞戞ǚ鏀介柛鈩冾殢娴煎矂姊虹紒妯诲蔼闁稿海鏁诲璇测槈濡攱顫嶉梺鍛婎殘閸嬬偤鎮甸灏栨斀闁绘劘娉涙禍褰掓煕閻樺磭澧い顐㈢箻閹煎綊宕烽鐙呯床婵犵妲呴崹鎵崲濡ゅ懎纾婚柟鐗堟緲缁狀垶鏌涢幇闈涙灍闁哄懏绻堥弻娑氫沪閸欍儰瀛╅悶姘箞濮婄粯鎷呯憴鍕哗闂佺瀛╁钘夌暦濠婂喚娼╅悹楦挎閿涙盯姊洪幐搴ｇ畵妞わ缚鍗抽幃锟犲即閵忥紕鍘繝銏ｅ煐閿氶柣蹇曞У閵囧嫰鍩℃繝鍌氼潽闂侀潧娲ょ€氫即寮崒鐐村亗閹兼惌鍠楅崰姗€姊绘笟鈧埀顒傚仜閼活垱鏅堕娑栦簻闁哄啠鍋撻柣妤冨Т閻ｇ兘骞庣粵瀣櫔闂侀€炲苯澧柣锝呭槻椤劑宕奸悢铚傜盎濠电娀娼ч崐鎼佀囬姘ｆ灁闁靛鍎哄〒濠氭煏閸繃顥滃┑顔ㄥ懐纾奸棅顐幘閻瑦銇勯姀鈥冲摵闁哄苯妫楅濂稿幢濡炶浜鹃柣銏犳啞閻撶喖鏌曡箛濠冾潑闁哥喎绻橀弻锝夊箳濡ゅ啰鏆梺鍝勭焿缂嶄線銆佸鈧幃銈嗘媴閸︻厽婢掗梻?	// 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柛鎰紦閻㈢粯淇婇悙顏勨偓鏍偋濡ゅ啰鐭欓柟鐐湽閳ь剙鎳樺畷锝嗗緞瀹€鈧惁鍫ユ⒒閸屾氨澧涚紒瀣浮閺佸秹寮崼鐔哄幘闂佸搫顦冲▔鏇熺閵忋倖鐓忛柛鈩兠粭鎺撱亜椤愶絿绠為柡浣瑰姍瀹曘劑顢楅埀顒勊夊顑芥斀閹烘娊宕愬Δ浣瑰弿闁绘垼妫勭壕缁樼箾閹寸偞鐨戦柣顓熸崌閺屾盯顢曢敐鍡欘槬闂佺粯甯掗悘姘跺Φ閸曨垰绠抽柟鎼灠婵稓绱撴担鍓叉█婵炰匠鍥ㄧ畳婵犵數濮撮敃銈団偓姘煎櫍閹本鎯旈～顑跨盎闂侀潧鐗嗛幊搴敂椤撶喐鍙忓┑鐘插亞閻撹偐鈧娲樼敮鎺楀煝鎼淬劌绠ｆ繝闈涙－閻庢潙鈹戦悩鍨毄濠殿喕鍗冲畷鐟懊洪浣插亾閿曞倸鐐婇柍杞扮閻忓﹪姊洪崫鍕殭闁绘妫濋弻瀣炊椤掍胶鍘搁梺鎼炲劗閺呮盯寮告惔鈭跺綊鎮╅懡銈囨毇濠殿喖锕ㄥ▍锝夊礌閺嶎厼鍗抽柣鎰ゴ閹枫倖淇婇悙顏勨偓銈夊储閻撳篃娲敇閵忕姈銉╂煕濞戞瑦缍戦柦鍐枛閺屾洘寰勫☉姗嗘喘闂佸憡锕㈡禍璺侯潖濞差亜妫橀柕澶涢檮閻濇棃姊洪崫銉バｉ柤褰掔畺閸┿垽骞樼拠鑼潉闂侀€炲苯澧村┑锛勬暬瀹曠喖顢涢敐鍡樻珖婵＄偑鍊栫敮鎺楀磹閺勫繆鍙洪梻鍌氬€峰ù鍥敋閺嶎厼绐楁俊銈呮噷閳ь剙鍟换婵嬪磻閼恒儳娲寸€规洘锕㈡俊鎼佸閳藉棙缍屽┑鐘垫暩閸嬫稑螞濞嗘帗鎳屾俊鐐€戦崕鑽ょ礊娴ｅ壊娼栫紓浣股戞刊鏉戭渻鐎ｎ亞鍑归悷鏇炴濮婂宕掑顒変患闁诲孩纰嶅姗€鎮鹃悿顖樹汗闁圭儤鍨块弫婊冣攽鎺抽崐鎰板磻閹剧粯鐓㈤柛鎰典簻閺嬫盯鏌＄仦璇插鐎殿喗鎸抽幃娆徝圭€ｎ亙澹曠紓浣割儐鐎笛囧汲濠婂牊鈷掗柛灞剧懅閸斿秹鎮楃粭娑樺幘閸濆嫷鍚嬪璺猴功閿涙盯姊洪崷顓炰壕闁伙綆浜濈粋宥咁煥閸喓鍘梺鍓插亝缁诲啴宕冲ú顏呯厱闁靛牆鎳忛崰姗€鏌″畝瀣？濞寸媴绠撳畷婊嗩檨闁诲繗浜槐鎾存媴閸濆嫅锝囩磼鐎ｎ偄娴鐐叉瀵噣宕煎┑鎰秱婵＄偑鍊ら崑鎺楀礈濞嗘劗顩烽柟鐑樺灍閺€浠嬫煥濞戞ê顏╁ù婊冦偢閺屾稒绻濋崘顏勨拰閻庢鍠栭…閿嬩繆濮濆矈妲奸悗瑙勬礀瀵埖绌辨繝鍥舵晬婵犻潧娴傛禒鈺呮⒑閸濆嫭锛旂紒鐘虫崌瀵寮撮悢椋庣獮闂佺硶鍓濊摫闁绘繐绠撳鐑樻姜閹殿噮妲紓浣割槺閺佹悂骞戦姀鐘斀闁搞儮鏅濋惁鍫ユ⒑缁嬫寧婀伴柛鎴ｎ潐缁傛帡骞嗚閺€?TOTP 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿垂妤ｅ啫绠涘ù锝呮贡缁嬩礁鈹戦悩鍨毄濠殿喚鏁婚幊婵嬪礈瑜忛悳缁樹繆閵堝懏鍣洪柣鎾寸洴閺屾盯濡烽敐鍛闂佽绻嗛弲婵堟閹烘鐒?
	if req.TotpEnabled && !previousSettings.TotpEnabled {
		// 闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊椤掑鏅悷婊冪箻楠炴垿濮€閵堝懐鐤€濡炪倖妫佸Λ鍕償婵犲洦鈷戦柣鐔煎亰閸ょ喎鈹戦鈧褔鎮鹃悜钘夎摕闁靛濡囬崢閬嶆⒑閺傘儲娅呴柛鐔跺嵆閸╁﹪寮撮姀锛勫幈闁硅偐琛ュΣ鍕绩婵犳碍鐓忛柛銉戝喚浼冮悗娈垮枙缁瑩銆侀弴顫稏妞ゆ挾濮靛畷鐔兼⒒娴ｈ棄鍚归柛鐘冲姍閹兘鏌嗗鍡忔寖闂佸憡娲﹂崹鐗堫攰闂備礁鎲″ú锕傚垂娴兼潙绀冮柍褜鍓熷娲濞淬倖鐩畷銊╊敍濮橆剛鐟ㄩ梻?TOTP闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍛婄秶闁告挆鍛闂佽瀛╅懝楣兯囨导鏉懳﹂柛鏇ㄥ灠缁犳娊鏌熺€涙绠ラ柍褜鍓氶懝楣冣€﹂崸妤佸殝闂傚牊绋戦～宀勬倵閻熺増鍟炵紒璇茬墦瀵鍩勯崘銊х獮闁诲函缍嗘禍鐐村閸ヮ剚鈷戠痪顓炴噹琚氬┑鐐跺皺閸犲酣锝炶箛鎾佹椽顢旈崪浣诡棃婵犵數鍋為崹鍏笺仈濮濆苯鈧挳姊婚崒娆掑厡妞ゎ厼鐗撻弻濠囨晲閸℃瑯娲搁梺鍓插亝濞叉﹢寮查弻銉︾厽闁归偊鍠栭崝瀣煛閸涱喖绲绘い顓℃硶閹瑰嫭绗熼姘闂備線鈧偛鑻晶浼存煛娴ｇ瓔鍤欐い顐㈢箲缁绘繂顫濋鍕暪闂備胶绮Λ渚€濡撮埀顒€鈹戦鍏兼悙妞ゎ亜鍟存俊鍫曞礃閵娿儱顫氱紓鍌欑贰閸犳骞戦崶褜鍤曞┑鐘崇閺呮彃顭跨捄渚剱婵炲懎妫濆娲嚍閵夊喚浜棟閺夊牄鍔岄杈┾偓鍏夊亾闁告洦鍏橀幏娲⒑閸涘﹦绠撻悗姘煎櫍閹偟鎹勯妸褏锛滈梺缁橆焾濞呮洜浜搁弶澹冲酣宕惰闊剟鏌熼鐣岀煉闁瑰磭鍋ゆ俊鐑芥晜閻ｅ苯绲鹃梻鍌氬€风粈渚€骞夐敓鐘冲仭妞ゆ牜鍋涚壕濠氭煙閹冾暢濞戞挸绉归弻鐔煎箲閹伴潧娈梺钘夊暟閸犳劗鎹㈠☉銏犵闁绘垵娲ｇ欢闈涒攽閻愯尙姣為柡鍛█瀵鈽夐姀鐘殿啋濠德板€愰崑鎾绘煙閸愯尙绠伴柍缁樻崌楠炴牗鎷呴崗澶嬪濠电偠鎻徊浠嬪箟閿熺姴鐤柣鎰劋閻撴洟鏌嶆潪鎵瓘闁告梻鍠栭弻鐔肩嵁閸喚浠奸梺瀹狀潐閸ㄥ綊鍩€椤掑﹦绉甸柛鎾寸懅缁﹪鏁冮崒娑樷偓鍨殽閻愯尙浠㈤柛鏃€纰嶇换娑㈡嚑椤掆偓閳诲牏鈧娲橀崹鍧楃嵁濡偐纾兼俊顖滅帛閻濇娊姊洪崷顓炲付闁宦板妿閹广垽宕熼娑樺壆闁瑰吋鐣崝宥夊磻閻樿櫕鍙忔俊顖氭惈閼稿綊鏌＄€ｎ偆鈯曢柟渚垮妽缁绘繈宕熼鐐殿偧闂備胶鎳撻崲鏌ュ箠閹版澘绠熼柟缁㈠枛缁€瀣亜閹捐泛浠滄繛鍫涘劤缁辨捇宕掑顑藉亾妞嬪海鐭嗗〒姘ｅ亾閽樻繃銇勯弽顐沪闁稿顦埞鎴﹀磼濠婂海鍔哥紓?
		if !h.settingService.IsTotpEncryptionKeyConfigured() {
			response.BadRequest(c, "Cannot enable TOTP: TOTP_ENCRYPTION_KEY environment variable must be configured first. Generate a key with 'openssl rand -hex 32' and set it in your environment.")
			return
		}
	}
	conversationAuditPasswordHash := previousSettings.ConversationAuditSecondaryPasswordHash
	if password := strings.TrimSpace(req.ConversationAuditSecondaryPassword); password != "" {
		if len([]rune(password)) < 6 {
			response.BadRequest(c, "Conversation audit secondary password must be at least 6 characters")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		conversationAuditPasswordHash = string(hash)
	}
	conversationAuditCleanupEnabled := previousSettings.ConversationAuditCleanupEnabled
	if req.ConversationAuditCleanupEnabled != nil {
		conversationAuditCleanupEnabled = *req.ConversationAuditCleanupEnabled
	}
	conversationAuditRetentionDays := previousSettings.ConversationAuditRetentionDays
	if conversationAuditRetentionDays <= 0 {
		conversationAuditRetentionDays = 90
	}
	if req.ConversationAuditRetentionDays != nil {
		conversationAuditRetentionDays = *req.ConversationAuditRetentionDays
	}
	if conversationAuditRetentionDays <= 0 {
		conversationAuditRetentionDays = 90
	}
	if conversationAuditRetentionDays > 3650 {
		conversationAuditRetentionDays = 3650
	}

	loginAgreementMode := strings.ToLower(strings.TrimSpace(req.LoginAgreementMode))
	if loginAgreementMode == "" {
		loginAgreementMode = strings.ToLower(strings.TrimSpace(previousSettings.LoginAgreementMode))
	}
	switch loginAgreementMode {
	case "", "modal":
		loginAgreementMode = "modal"
	case "checkbox":
	default:
		response.BadRequest(c, "Login agreement mode must be modal or checkbox")
		return
	}
	loginAgreementUpdatedAt := strings.TrimSpace(req.LoginAgreementUpdatedAt)
	if loginAgreementUpdatedAt == "" {
		loginAgreementUpdatedAt = strings.TrimSpace(previousSettings.LoginAgreementUpdatedAt)
	}
	loginAgreementDocuments := loginAgreementDocumentsToService(req.LoginAgreementDocuments)
	if len(loginAgreementDocuments) == 0 {
		loginAgreementDocuments = previousSettings.LoginAgreementDocuments
	}
	for _, doc := range loginAgreementDocuments {
		if strings.TrimSpace(doc.Title) == "" {
			response.BadRequest(c, "Login agreement document title is required")
			return
		}
		if len(doc.Title) > 80 {
			response.BadRequest(c, "Login agreement document title is too long (max 80 characters)")
			return
		}
		if len(doc.ContentMD) > 200*1024 {
			response.BadRequest(c, "Login agreement document content is too large (max 200KB)")
			return
		}
	}
	if req.LoginAgreementEnabled && len(loginAgreementDocuments) == 0 {
		response.BadRequest(c, "Login agreement documents are required when enabled")
		return
	}

	// LinuxDo Connect 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柛鎰紦閻㈠姊绘笟鈧褔藝椤愶箑鐤炬繛鎴炶壘椤ユ岸鏌涢敂璇插箺闁哥姵鍔欓弻锝呂旈埀顒勬偋閸℃瑧绠旈柟鐑樻尰閸欏繘鎮峰▎蹇擃伀闁告瑢鍋撻梻浣告惈閻绱炴担鍓插殨妞ゆ帒瀚崹鍌涖亜閺囩偞鍣烘い銉﹀灦缁绘繈鎮介棃娑楃捕闂佺粯顨呯换妯侯嚕閺屻儺鏁冮柨鏇楀亾闁绘帒鐏氶妵鍕箳閸℃ぞ澹曢梻浣告啞钃辩紒顔芥尭閻ｇ兘骞嬮敃鈧粻鑽ょ磽娴ｈ鐒介柛姗€浜跺铏规喆閸曨剛鍑￠梺鍛婂焹閸嬫挾绱撴担闈涘闁告艾顑夋俊?
	if req.LinuxDoConnectEnabled {
		req.LinuxDoConnectClientID = strings.TrimSpace(req.LinuxDoConnectClientID)
		req.LinuxDoConnectClientSecret = strings.TrimSpace(req.LinuxDoConnectClientSecret)
		req.LinuxDoConnectRedirectURL = strings.TrimSpace(req.LinuxDoConnectRedirectURL)

		if req.LinuxDoConnectClientID == "" {
			response.BadRequest(c, "LinuxDo Client ID is required when enabled")
			return
		}
		if req.LinuxDoConnectRedirectURL == "" {
			response.BadRequest(c, "LinuxDo Redirect URL is required when enabled")
			return
		}
		if err := config.ValidateAbsoluteHTTPURL(req.LinuxDoConnectRedirectURL); err != nil {
			response.BadRequest(c, "LinuxDo Redirect URL must be an absolute http(s) URL")
			return
		}

		// 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柛娑橈攻閸欏繘鏌ｉ幋婵愭綗闁逞屽墮閸婂湱绮嬮幒鏂哄亾閿濆簼绨介柛鏃撶畱椤啴濡堕崱妤€娼戦梺绋款儐閹瑰洭寮诲☉銏″亜闂佸灝顑呮禒鎾⒑缁洘鏉归柛瀣尭椤啴濡堕崱妤€娼戦梺绋款儐閹稿墽妲愰幘鎰佸悑闁告粌鍟抽崥顐⑽旈悩闈涗粶闁哥噥鍋夐悘鎺楁煟閻樺弶绌块悘蹇旂懅缁綁鎮欓悜妯锋嫼閻熸粎澧楃敮鎺撶娴煎瓨鐓曢柟鎯ь嚟閹冲洭鏌熼鈧褑鐏掓繛鎾村嚬閸ㄨ京鈧潧鐭傚娲濞戞艾顣洪梺纭呮珪閸旀鍒掔紒妯侯嚤閻庢稒顭囬崢鐢告⒑閸涘﹤濮囩€殿喖鐖奸獮鎴︽晲閸ワ絽浜鹃柛顭戝亝缁舵煡鏌ㄩ弴銊ら偗妤犵偛鍟撮幃婊堟嚍閵夈儲鐤傛俊鐐€栭崹鐓幬涢崟顒傤洸?client_secret闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍛婄秶濡わ絽鍟宥夋⒑閹肩偛鈧牠宕濋弽顓炍﹂柛鏇ㄥ灠閸愨偓濡炪倖鍔﹀鈧紒顔肩埣濮婃椽骞栭悙鎻掝瀳闂佺锕ょ紞濠冧繆閻㈢绀嬫い鏍ㄨ壘閸炪劑姊洪棃娴ゆ稒鎷呴幓鎺嶅闂佸湱鍎ら〃鍡涙偂閸愵喗鐓熼柣鏃傤焾椤ュ寮介敍鍕＝濞达綀娅ｇ敮娑氱磼鐠囪尙澧曢摶鐐烘煕閹扳晛濡锋俊鎻掔墛閹便劌顫滈崱妤€鈷掗梺鍝勵槸閹诧紕鎹㈠┑瀣仺闂傚牊绋愮划鍫曟⒑缂佹﹩娈旀俊顐ｇ箞楠炲棗鐣濋崟顐わ紲闂佺粯鍔樼亸顏堝箺閺囥垺鈷戦柟绋挎捣缁犳捇鏌熷畡閭﹀剶闁诡噯绻濇俊鐑藉煛閸屾粌骞堟繝鐢靛仦閸ㄩ潧鐣烽鈧锝嗐偅閸愨晝鍘搁梺绯曟閺呮盯鐛弽顓熺厓閻熸瑥瀚悘鎾煙椤旇娅婃い銏℃礋閿濈偤顢橀悜鍡橆棥婵犵數濮烽弫鎼佸磻閻愬搫绠板┑鐘宠壘閺勩儵鏌涢幇闈涙灈缁炬儳顭烽弻鐔煎箚瑜忔禍顏堟煕鐎ｎ偅灏甸柟鍙夋尦瀹曠喖顢楅崒銈喰為梻鍌欑閹测€愁潖瑜版帇鈧啴宕ㄩ弶鎴犲弨婵犮垼鍩栭崝鏇㈠垂閸屾稏浜滈柟浼存涧娴滄繈寮崼銉︹拺缁绢厼鎳庨ˉ宥夋煙濞茶绨界紒杈╁仱瀹曞崬顪冮～顔剧М鐎规洩绻濋幃娆撳煛閸屾稒婢戦梻鍌欒兌缁垵鎽銈嗘⒐閻楃姴顕ｉ幎绛嬫晬闁绘劕顕崢鎾绘偡濠婂嫮鐭掔€规洘绮岄埢搴ㄥ箻瀹曞洦鐒鹃梻浣侯潒閸曞灚鐣堕梺缁樺笒閻忔岸濡甸崟顖氱闁瑰瓨绺鹃崑鎾广亹閹烘垵鎯為梺閫炲苯澧存慨濠冩そ楠炴劖鎯旈姀顫喘闂備焦鎮堕崝蹇撯枍閿濆洤鍨濇い鎾卞灪閸嬪嫰鏌ｉ幘铏崳闁告棑绠戦—鍐Χ閸℃鐟ㄩ柣搴㈠嚬閸撶喎顕ｉ崨濠冨閻炴稈鈧厖澹曞Δ鐘靛仜閻忔繈宕濆顓濈箚妞ゆ劧缍嗗▓妯荤箾閸℃劕鐏查柟顔界懇閹粌螣閻撳骸绠炲┑锛勫亼閸婃牠宕濊缁辩偤鍩€椤掆偓椤法鎲撮崝绛嬩邯濠€?
		if req.LinuxDoConnectClientSecret == "" {
			if previousSettings.LinuxDoConnectClientSecret == "" {
				response.BadRequest(c, "LinuxDo Client Secret is required when enabled")
				return
			}
			req.LinuxDoConnectClientSecret = previousSettings.LinuxDoConnectClientSecret
		}
	}

	// DingTalk Connect 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柛鎰紦閻㈠姊绘笟鈧褔藝椤愶箑鐤炬繛鎴炶壘椤ユ岸鏌涢敂璇插箺闁哥姵鍔欓弻锝呂旈埀顒勬偋閸℃瑧绠旈柟鐑樻尰閸欏繘鎮峰▎蹇擃伀闁告瑢鍋撻梻浣告惈閻绱炴担鍓插殨妞ゆ帒瀚崹鍌涖亜閺囩偞鍣烘い銉﹀灦缁绘繈鎮介棃娑楃捕闂佺粯顨呯换妯侯嚕閺屻儺鏁冮柨鏇楀亾闁绘帒鐏氶妵鍕箳閸℃ぞ澹曢梻浣告啞钃辩紒顔芥尭閻ｇ兘骞嬮敃鈧粻鑽ょ磽娴ｈ鐒介柛姗€浜跺铏规喆閸曨剛鍑￠梺鍛婂焹閸嬫挾绱撴担闈涘闁告艾顑夋俊?
	// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌熼梻瀵割槮缁炬儳缍婇弻鐔兼⒒鐎靛壊妲梺姹囧€ら崰妤呭Φ閸曨垰鍐€闁靛ě鍛獥濠电偛鐡ㄧ划宀€绱炴繝鍥ц摕婵炴垯鍨圭粻娑㈡煕韫囨挻鎲搁柤铏そ濮婅櫣绱掑Ο铏圭懆闂佹寧娲忛崐婵嗭耿娓氣偓濮婅櫣绱掑Ο蹇ｄ簻铻ｅ┑鐘叉搐绾惧潡鐓崶銊︾缁炬儳銈搁弻锝呂熼崫鍕瘣闂佽绻嗛弲鐘诲蓟閻斿吋鍤戞い鎺戭槺閸旀悂姊虹€圭媭娼愰柛銊ョ仢閻ｉ攱绺介崨濠備簽婵炶揪绲藉﹢杈╃矙閹扮増鈷掗柛灞剧懆閸忓瞼绱掗鍛仸闁靛棗鍟村畷銊р偓娑櫭禍閬嶆椤愩垺澶勭紒瀣箻閸┾偓妞ゆ巻鍋撻柟铏耿楠炲啫顭ㄩ崼鐔锋疅闂侀潧顦崕浼村礌閺嵮€鏀介柣鎰煐瑜把呯磼闊厾鐭欐鐐搭殔椤劑宕奸悢鍛婄彨闁诲骸绠嶉崕鍗灻洪妸褍顥氶柣鎾冲瘨閻斿棝鎮归搹鐟扮殤闁衡偓婵犳碍鐓曢悗锝冨妼閸旓箓鏌″畝瀣？濞寸媴绠撳畷婊嗩槼闁告帗绋戣灃闁绘﹢娼ф禒婊勭箾瀹割喖寮柕鍡曠椤粓鍩€椤掑嫬绠栭柍鍝勬噹缁€鍐煕濞嗗浚妲虹憸鐗堢洴濮婂宕掑顑藉亾妞嬪海鐭嗗〒姘ｅ亾妤犵偛顦甸弫鎾绘偐閸愯弓鐢婚梻渚€娼чˇ顓㈠磿濞嗗繆妲堟俊顖炴敱椤秹姊洪棃娑氱濠殿喚鍏橀、娆愬緞閹邦厸鎷洪梺鍛婄☉閿曘儲寰勯崟顖涚厱闁靛鍎崑銏⑩偓娈垮櫘閸嬪嫰顢樻總绋垮耿婵☆垰鎼导搴㈢節绾版ɑ顫婇柛銊ゅ嵆婵￠潧顫滈埀顒€鐣烽妷鈺婃晝闁挎梻鏅崢钘夆攽閻愭潙鐏ョ€规洦鍓熷鎼佹偄閾忓湱锛滅紓鍌欑劍椤洤煤鐎涙﹩娈介柣鎰▕閸庢棃鏌熼鐣屾噰鐎殿喖鐖奸獮瀣偑閳ь剟寮抽敓鐘斥拻闁稿本鐟ч崝宥夋倵缁楁稑鍘炬ウ璺ㄧ杸婵炴垶顭囬敍鐔兼⒑閸涘﹦鎳冩い锕備憾閸╃偛顓奸崨顏呮杸闂佺粯蓱瑜板啴顢旈锔界厽妞ゆ挾鍠撻幊鍕磼缂佹娲寸€规洖宕灒闁告繂瀚闂傚倷绀侀幖顐﹀箠閹邦厾绠鹃柍褜鍓氶幈銊︾節閸愨斂浠㈤悗瑙勬礈閸忔﹢銆佸鈧幃鈺呮嚑椤掑浠归梻鍌欑劍閻綊宕洪崟顖氬瀭闁规儼妫勭壕濠氭煏閸繍妲哥紒鐘崇墵閺屾洝绠涚€ｎ亖鍋撻妶澶婃瀬闁糕剝绋掗悡鏇㈡倶閻愪絻妾告繛鍫熸煥闇夋繝濠傛噹娴滃墽绱掔紒妯兼创鐎规洜鍘ч埞鎴﹀醇濠婂棙顎楅梻鍌欑閹碱偊顢栭崱娑欏剮妞ゆ牜鍋涢拑鐔哥箾閹存瑥鐏柛瀣閺屾稑鈽夊鍫濆闂佸憡鍩婄换婵嗩潖?corp_restriction_policy=whitelist 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸ゅ嫰鏌涢锝嗙５闁逞屽墾缁犳挸鐣锋總绋款潊闁炽儱鍟跨花銉╂⒒娴ｇ顥忛柛瀣╃窔瀹曟洘娼忛埡渚囨闂佺鍕垫畷闁绘挻娲樻穱濠囧Χ閸涱厽娈梺鍛婃崌娴滃爼寮?coerce 婵?none闂?	// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亜顒㈡い鎰矙閺屻劑鎮㈤崫鍕戙垻绱掗埀顒勫醇閻旇櫣顔曢梺绯曞墲钃遍悘蹇庡嵆閺屾稒绻濋崨顓犲帿缂備胶绮换鍫澪涢崘銊㈡闁告鍋涙竟鍫ユ⒒娴ｄ警鐒鹃悶姘煎亰瀹曟繈骞嬮敃鈧拑鐔兼煛閸ャ儱鐏╅梺鍗炴处缁绘繈妫冨☉娆戙偐濠碘槅鍋侀崐鏇⑩€旈崘顔嘉ч柛灞剧⊕閻濇牠姊虹粙璺ㄧ闁挎洏鍨介悰顕€宕橀鍛瀭闂佸憡娲﹂崜娆擃敁閹剧粯鈷戦柛娑橈功缂傛岸鏌涙惔銈呭惞婵″弶鍔欓幃娆撴偨閻㈢绱冲┑鐐舵彧缁茶姤绔熸繝鍋斤綁宕ㄦ繝鍕啎缂佺虎鍙冮ˉ鎾跺姬閳ь剟鎮楀▓鍨灈妞ゎ厾鍏樺顐﹀箛椤撶偟绐炴繝鐢靛Т鐎氼剟鐛幇鐗堚拻濞达絽鎲￠崯鐐烘煕閺冣偓椤ㄥ﹤鐣烽幋锕€绠婚柛鎴炴緲椤︾敻鐛鈧畷婊勬媴閻氬闂?admin API 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弴鐐测偓褰掑磿閹寸姵鍠愰柣妤€鐗嗙粭鎺旂磼閳ь剚寰勭仦绋夸壕闁稿繐顦禍楣冩⒑闁偛鑻晶鎾煕閳规儳浜炬俊鐐€栫敮濠勭矆娓氣偓瀹曠敻顢楅崟顒傚幈闂佽宕樺▔娑㈠几濞戙垺鐓忛柛鈩冩礈椤︼箓鎽堕敐澶嬪仭濞达綁娼ч埀顒佹尭鏁堟俊銈呮噺閳锋垿鏌涢敂璇插箹闁告柨顑夐弻娑㈠Ω閵壯呅ㄥ銈冨灪閿曘垽骞冨鍏剧喖宕烽鐘垫毎闂傚倷鑳剁划顖炲礉閺囩倣鐔哥節閸パ咃紵闂婎偄娲︾粙鎺楀磹閻㈠憡鐓曟い顓熷灥娴滄粎绱掑Δ浣稿摵闁诡喗顨婂Λ鍐ㄢ槈閺嵮傚垝闁诲氦顫夊ú妯兼暜閿熺姴绠栨繛鍡樺灍閸嬫捇鎮藉▓璺ㄥ姼濡ょ姷鍋涚粔鐟邦潖濞差亝鍤掗柕鍫濇啗椤掍胶绠鹃柛婊冨暟閹厧霉濠婂嫭鍊愬┑鈩冩倐閸┾剝鎷呴崫鍕啟缂傚倸鍊搁崐鐑芥倿閿曚礁缍樺┑鐐茬摠缁秶鍒掗幘璇茶摕闁绘柨鍚嬮崐缁樹繆椤栨繍鍤欓梻澶婄Ч閺岋綀绠涢幘鑽ゅ椽婵犵鈧櫕鍠橀柛鈹惧亾濡炪倖甯掗崰姘焽閹邦厾绠鹃柛娆忣檧閼拌法鈧娲樼换鍌炲煝鎼淬劌绠荤€规洖娲ㄩ弳顐︽⒒娴ｈ鐏遍柡鍛洴瀹曨垶顢涢悙鑼紵闂佺粯鏌ㄩ崥瀣偂閸愵喗鐓冮弶鐐村椤︼箓鏌￠崱娆忎沪闁靛洤瀚粻娑㈠即閻愬浜繝?DB闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍛婄秶濡わ絽鍟宥夋⒑缁嬫鍎忔い鎴濐樀瀵鈽夐姀鐘靛姶闂佺绻嗛弲婊堝垂閻㈢鐓濈€广儱顦伴崑鎰磽娴ｈ鐒介柛妯挎閳规垿鍩ラ崱妤冧哗闂佸憡鑹鹃澶婎嚕閹惰棄鐓涢柛娑卞枤閸?UI 闂傚倸鍊搁崐宄懊归崶顒夋晪鐟滃酣銆冮妷褏鐭欓柛鏌倐鍋撻崸妤佲拺妞ゆ巻鍋撶紒澶樺櫍閸┾偓妞ゆ帒锕﹂悾鐢碘偓瑙勬礈閸樠囧煘閹达箑绠涙い鎺嶇贰閸氭瑩姊绘担钘夊惞濠殿喗鍎抽々濂稿Ω瑜忕粈濠勨偓骞垮劚閹冲危閸儲鐓忛煫鍥ㄦ礀椤秶鎲搁弮鍫濊摕婵炴垶锕╅悡銉╂煃瑜滈崜鐔煎箖閻愮儤鏅濋柍褜鍓熼妴鍛附缁嬭鈺呮煃閸濆嫬鈧敻骞忓ú顏呪拻闁稿本姘ㄦ晶娑氱磼鐎ｎ偅灏伴柡鍛埣瀵挳鎮㈢紙鐘电泿闂備焦瀵уΛ渚€顢氳缁傚秹骞嶉鐟颁壕閻熸瑥瀚粈鍐煕閵娿儲鍋ラ柣娑卞枛椤粓鍩€椤掆偓椤曪綁宕奸弴鐐殿吅闂佺粯顭囬弫绋跨暤娓氣偓濮婂宕掑▎鎰偘濡炪値鍋勯ˇ鎵博閻旇　鍋撻敐搴℃灈婵☆偅锚閵嗘帒顫濋敐鍛闁诲氦顫夊ú姗€宕归崸妤冨祦婵せ鍋撴鐐叉处閹峰懘宕ㄦ繝鍌楀亾椤栨埃鏀介柣姗嗗枛閻忚鲸绻涙径瀣创妞ゃ垺鐗犲畷鍗烆渻缂佹鈧椽姊洪崫鍕犻柛鏂挎捣瀵囧焵椤掑嫭鈷戦梻鍫熺洴閻涙粎绱掗幓鎺戔挃婵炴垵鐏氶妶锝夊礃閳哄啫骞楁繝寰锋澘鈧劙宕戦幘缁樼厱闁绘柨鎲＄紞鎴︽煟?	req.DingTalkConnectCorpRestrictionPolicy = service.CoerceDingTalkCorpPolicyForWrite(req.DingTalkConnectCorpRestrictionPolicy)

	if req.DingTalkConnectEnabled {
		req.DingTalkConnectClientID = strings.TrimSpace(req.DingTalkConnectClientID)
		req.DingTalkConnectClientSecret = strings.TrimSpace(req.DingTalkConnectClientSecret)
		req.DingTalkConnectRedirectURL = strings.TrimSpace(req.DingTalkConnectRedirectURL)
		req.DingTalkConnectCorpRestrictionPolicy = strings.TrimSpace(req.DingTalkConnectCorpRestrictionPolicy)
		req.DingTalkConnectInternalCorpID = strings.TrimSpace(req.DingTalkConnectInternalCorpID)

		if req.DingTalkConnectClientID == "" {
			response.BadRequest(c, "DingTalk Client ID is required when enabled")
			return
		}
		if req.DingTalkConnectRedirectURL == "" {
			response.BadRequest(c, "DingTalk Redirect URL is required when enabled")
			return
		}
		if err := config.ValidateAbsoluteHTTPURL(req.DingTalkConnectRedirectURL); err != nil {
			response.BadRequest(c, "DingTalk Redirect URL must be an absolute http(s) URL")
			return
		}

		// 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柛娑橈攻閸欏繘鏌ｉ幋婵愭綗闁逞屽墮閸婂湱绮嬮幒鏂哄亾閿濆簼绨介柛鏃撶畱椤啴濡堕崱妤€娼戦梺绋款儐閹瑰洭寮诲☉銏″亜闂佸灝顑呮禒鎾⒑缁洘鏉归柛瀣尭椤啴濡堕崱妤€娼戦梺绋款儐閹稿墽妲愰幘鎰佸悑闁告粌鍟抽崥顐⑽旈悩闈涗粶闁哥噥鍋夐悘鎺楁煟閻樺弶绌块悘蹇旂懅缁綁鎮欓悜妯锋嫼閻熸粎澧楃敮鎺撶娴煎瓨鐓曢柟鎯ь嚟閹冲洭鏌熼鈧褑鐏掓繛鎾村嚬閸ㄨ京鈧潧鐭傚娲濞戞艾顣洪梺纭呮珪閸旀鍒掔紒妯侯嚤閻庢稒顭囬崢鐢告⒑閸涘﹤濮囩€殿喖鐖奸獮鎴︽晲閸ワ絽浜鹃柛顭戝亝缁舵煡鏌ㄩ弴銊ら偗妤犵偛鍟撮幃婊堟嚍閵夈儲鐤傛俊鐐€栭崹鐓幬涢崟顒傤洸?client_secret闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍛婄秶濡わ絽鍟宥夋⒑閹肩偛鈧牠宕濋弽顓炍﹂柛鏇ㄥ灠閸愨偓濡炪倖鍔﹀鈧紒顔肩埣濮婃椽骞栭悙鎻掝瀳闂佺锕ょ紞濠冧繆閻㈢绀嬫い鏍ㄨ壘閸炪劑姊洪棃娴ゆ稒鎷呴幓鎺嶅闂佸湱鍎ら〃鍡涙偂閸愵喗鐓熼柣鏃傤焾椤ュ寮介敍鍕＝濞达綀娅ｇ敮娑氱磼鐠囪尙澧曢摶鐐烘煕閹扳晛濡锋俊鎻掔墛閹便劌顫滈崱妤€鈷掗梺鍝勵槸閹诧紕鎹㈠┑瀣仺闂傚牊绋愮划鍫曟⒑缂佹﹩娈旀俊顐ｇ箞楠炲棗鐣濋崟顐わ紲闂佺粯鍔樼亸顏堝箺閺囥垺鈷戦柟绋挎捣缁犳捇鏌熷畡閭﹀剶闁诡噯绻濇俊鐑藉煛閸屾粌骞堟繝鐢靛仦閸ㄩ潧鐣烽鈧锝嗐偅閸愨晝鍘搁梺绯曟閺呮盯鐛弽顓熺厓閻熸瑥瀚悘鎾煙椤旇娅婃い銏℃礋閿濈偤顢橀悜鍡橆棥婵犵數濮烽弫鎼佸磻閻愬搫绠板┑鐘宠壘閺勩儵鏌涢幇闈涙灈缁炬儳顭烽弻鐔煎箚瑜忔禍顏堟煕鐎ｎ偅灏甸柟鍙夋尦瀹曠喖顢楅崒銈喰為梻鍌欑閹测€愁潖瑜版帇鈧啴宕ㄩ弶鎴犲弨婵犮垼鍩栭崝鏇㈠垂閸屾稏浜滈柟浼存涧娴滄繈寮崼銉︹拺缁绢厼鎳庨ˉ宥夋煙濞茶绨界紒杈╁仱瀹曞崬顪冮～顔剧М鐎规洩绻濋幃娆撳煛閸屾稒婢戦梻鍌欒兌缁垵鎽銈嗘⒐閻楃姴顕ｉ幎绛嬫晬闁绘劕顕崢鎾绘偡濠婂嫮鐭掔€规洘绮岄埢搴ㄥ箻瀹曞洦鐒鹃梻浣侯潒閸曞灚鐣堕梺缁樺笒閻忔岸濡甸崟顖氱闁瑰瓨绺鹃崑鎾广亹閹烘垵鎯為梺閫炲苯澧存慨濠冩そ楠炴劖鎯旈姀顫喘闂備焦鎮堕崝蹇撯枍閿濆洤鍨濇い鎾卞灪閸嬪嫰鏌ｉ幘铏崳闁告棑绠戦—鍐Χ閸℃鐟ㄩ柣搴㈠嚬閸撶喎顕ｉ崨濠冨閻炴稈鈧厖澹曞Δ鐘靛仜閻忔繈宕濆顓濈箚妞ゆ劧缍嗗▓妯荤箾閸℃劕鐏查柟顔界懇閹粌螣閻撳骸绠炲┑锛勫亼閸婃牠宕濊缁辩偤鍩€椤掆偓椤法鎲撮崝绛嬩邯濠€?
		if req.DingTalkConnectClientSecret == "" {
			if previousSettings.DingTalkConnectClientSecret == "" {
				response.BadRequest(c, "DingTalk Client Secret is required when enabled")
				return
			}
			req.DingTalkConnectClientSecret = previousSettings.DingTalkConnectClientSecret
		}

		// Validate DingTalk Corp restriction settings.
		dingTalkCfg := config.DingTalkConnectConfig{
			Enabled:               true,
			DingTalkAppKind:       "internal_app", // 缂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸ゅ嫰鏌涢锝嗙闁诡垳鍋ら獮鏍庨鈧俊濂告煟椤撶噥娈滄鐐寸墪鑿愭い鎺嗗亾濠德ゅ亹缁辨帡鍩€椤掑嫬閱囬柣鏃囨椤旀洟姊洪悷閭﹀殶闁稿鍠栭妴鍌炲矗婢跺牅绨诲銈嗗姧缁插墽绮堥崘顔界厪闁搞儜鍐句紓缂備胶濮甸惄顖炵嵁濡吋宕夊〒姘煎灡濠㈡垿姊婚崒娆掑厡妞ゎ厼鐗忛埀顒佺▓閺呯娀寮€ｎ喗鈷戦梻鍫熺⊕椤ョ偤鎮介娑樻诞闁挎繄鍋涢埞鎴犫偓锝呯仛閺呫垺绻濋悽闈浶㈤柛濠冩倐閿濈偤鎮╅悽鐢碉紳闂佺鏈悷褔宕濆澶嬬厱闁哄啠鍋撻柣妤冨█瀹曟椽宕ㄩ姘兼闂佺艌閸嬫悤ings 闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊椤掑鏅悷婊冪箻楠炴垿濮€閵堝懐顔婂┑掳鍊愰崑鎾剁棯閸撗呭笡缂佺粯鐩獮瀣枎韫囨洑鎮ｇ紓浣瑰劤婢т粙骞婇幘鐑┾偓鏃堝礃椤忎礁浜鹃柨婵嗙凹闁垱銇勬惔銏╂疁闁哄本鐩俊鎼佹晜閸撗呮澖婵°倗濮烽崑鐐烘偋閺団€崇倒婵＄偑鍊栧濠氬磻閹炬枼鏀介梽鍥磻閹邦喗顫曢柟鎯х摠婵挳鏌涘▎蹇ｆЧ闁绘繃娲熷娲濞戞瑦鎮欓柣搴㈢濠㈡﹢顢氶敐鍡欘浄閻庯綆鈧厸鏅濋幉鍛婂緞婵烆澁缍佸畷濂稿Ψ閿旇瀚?internal_app
			AppType:               "internal",
			CorpRestrictionPolicy: req.DingTalkConnectCorpRestrictionPolicy,
			InternalCorpID:        req.DingTalkConnectInternalCorpID,
		}
		// 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕閻庤娲忛崕鎶藉焵椤掑﹦绉甸柛鐘愁殜瀹曟洟骞嬮悩鍐叉瀾闂佺粯顨呴悧鍡欑箔閹烘梻妫柟顖嗗嫬浠撮梺鍝勭灱閸犳挾妲愰幒鎳崇喎煤椤忓嫭鐤傞梻鍌欐祰閸嬫劖鏅跺Δ鍐ｅ亾缁楁稑娲ゆ闂佸憡娲﹂崹鎵不婵犳碍鍋ｉ柧蹇曟嚀閸斿鏌ｈ箛搴ゅ厡缂?corp_restriction_policy闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢琛″亾閻㈡鐒惧ù鐘欏洦鈷掗柛鏇ㄥ亜椤忣參鏌″畝瀣瘈鐎规洘锕㈡俊鎼佸Ψ閵忕姳澹曢梺鑽ゅ枑婢瑰棝寮抽敃鍌涚厱婵炲棗娴氬Σ褰掓煕濞嗗繒绠插ǎ鍥э躬椤㈡稑顫濋崣妯挎闂備胶绮幐濠氬礉濞嗘挸钃熸繛鎴炲焹閸嬫捇鏁嶉崡鐐差仾闁绘繃娲滅槐鎾存媴閾忕懓绗￠梺鐑╂櫓閸ㄤ即鎮鹃悜钘夌闁绘劕绉靛Λ鍐ㄧ暦濮椻偓椤㈡柨顓奸崨顔芥瘑闂傚倸鍊搁崐鐑芥倿閿曞倸绠栭柛顐ｆ礀绾惧綊鏌″搴′簼闁哄棙绮撻弻銊╂偄閸濆嫅銏ゆ煢閸愵亜鏋涢柡灞诲姂閹垽宕崟鎴欏灲濮婃椽宕￠悙鏉戭槱缂備胶绮换鍫濈暦瑜版帩鏁冮柕蹇曞濞兼棃姊绘担鍝ユ瀮妞ゆ泦鍥ㄥ亱闁圭偓鍓氬鏍ㄧ箾瀹割喕绨荤€瑰憡绻傞埞鎴︽倷閹碱厽鐣风紓渚囧枟閹告悂鎮鹃悜钘夌疀妞ゆ垟鏅滃Λ鍐ㄧ暦濮椻偓閸┾剝绻濋崘顏嗏棩?
		if dingTalkCfg.CorpRestrictionPolicy == "" {
			dingTalkCfg.CorpRestrictionPolicy = previousSettings.DingTalkConnectCorpRestrictionPolicy
		}
		// 闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊椤掑鏅悷婊冪Ч濠€渚€姊虹紒妯虹伇婵☆偄瀚板鍛婃媴缁洘鏂€闂佺粯锚閻ゅ洦绔熷鈧弻锝夊箳閹寸姳绮甸梺闈涙搐鐎氫即鐛幒鎴悑闁割偅绻傜敮顖炴⒒娓氣偓濞佳兾涘▎鎴炴殰闁跨喓濮撮拑?internal_only 缂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮崹顔规寖闂佹椿鍘介悷鈺呭蓟閻斿吋鍊绘慨妤€妫欓悾鐑芥⒑缁嬫鍎忔い鎴濇閹广垹鈹戦崶鈺冪槇闂佺鏈划宀勩€傞崫鍕ㄦ斀妞ゆ梹顑欏鎰版煟閹垮嫮绡€鐎殿喖顭烽幃銏ゆ偂鎼达綆鍚嬫俊鐐€栧濠氬煕閸儱姹叉い鎾卞灪閳锋垿鏌涘┑鍡楊仾濠殿垰銈搁弻娑㈠箻鐠虹儤鐏堝Δ鐘靛仜閸燁偉鐏冮梺鍛婁緱閸犳牗绂掗銏♀拺閻熸瑥瀚烽崯蹇涙煕閻樺疇澹橀柣锝囧厴閵嗗倿寮电粻鍎乸e 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂磋閳ь剨绠撻、妤呭礋椤愩倧绱遍梻浣告啞濞诧箓宕愮€ｎ€㈡椽顢旈崟骞喚鐔嗛悹铏瑰皑閸旂喖鏌ｉ妶鍛悙闁宠鍨块、娆愭叏閹邦亞鎹曢梻浣侯焾椤戝棝骞愰幖浣哥厴闁硅揪闄勯弲顒勬煕閺囩偟浠涙い銉︽尵缁辨挻鎷呴崫鍕戯綁鏌ｉ埡濠傜仩妞ゆ洩缍侀、姘跺焵椤掆偓閻ｇ兘骞掗幊铏⒐閹峰懏娼幍顔垮厭?internal闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢琛″亾濞戞瑯鐒界紒鐘卞嵆閺? 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犖ч柛銉㈡櫇閸橆垶姊绘担鍛婂暈婵炶绠撳畷銏ゆ寠婢跺本娈鹃梺鍛婄懃椤﹁京寮ч埀顒勬⒒閸屾氨澧涘〒姘殜瀹曟洟骞囬悧鍫㈠幈闁诲函缍嗛崐鏍箣閻樺啿搴婂┑鐐村灟閸ㄥ綊鐛姀鈥茬箚妞ゆ牗绻嶉崵娆撴⒒婢跺﹦效婵?
		if dingTalkCfg.CorpRestrictionPolicy == "internal_only" {
			dingTalkCfg.AppType = "internal"
		} else {
			dingTalkCfg.AppType = "public"
		}
		if err := config.ValidateDingTalkConfig(dingTalkCfg); err != nil {
			response.ErrorWithDetails(c, http.StatusBadRequest, err.Error(), mapDingTalkValidateError(err), nil)
			return
		}

		// bypass_registration 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾剧粯绻涢幋娆忕労闁轰礁顑嗛妵鍕箻鐠虹儤鐎鹃梺鍛婄懃缁绘﹢寮婚悢铏圭＜闁靛繒濮甸悘鍫㈢磽娴ｅ搫啸濠电偐鍋撳Δ鐘靛仦閻楁粓宕氶幒鏃€鍠嗛柛鏇炵仛缂嶅牓姊?internal_only 濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴濐潟閳ь剙鍊块幐濠冪珶閳哄绉€规洏鍔戝鍫曞箣濠靛牃鍋撻鐑嗘富闁靛牆鎳愮粻浼存煟濡も偓濡繈骞冮悙鍝勫瀭妞ゆ劗濮崇花濠氭⒑閻熺増鎯堟俊顐ｎ殕缁傚秹宕滆绾惧ジ鏌涢幘妤€妫欓妤呮⒑閸涘鎴﹀磿閹惰棄鐓″鑸靛姈閸嬨劎绱掔€ｎ厽纭堕柛鏃撶畱椤啴濡堕崱妤冪憪闂佺厧鍟挎晶搴ｅ垝缂佹ǜ鍋呴柛鎰ㄦ櫇閸樼敻鏌ｆ惔锝嗘毄妞ゎ厼鐗婄粋鎺楁晝閸屾稑鈧爼鐓崶椋庡埌妤犵偞顨婇弻鏇㈠炊瑜嶉顓燁殽閻愬樊鍎忛柍瑙勫灴瀹曞爼鎳滃▓鎸庯紘闂傚倸鍊烽悞锔锯偓绗涘厾楦跨疀濞戞ê鐎梺鎯х箰濠€杈╃不閺冨牊鐓熼柡鍐ㄥ€哥敮鍫曟煟閹烘垹浠涢柕鍥у楠炴帡骞嬪┑鎰磻闁诲氦顫夐幐椋庢濮樿泛钃熸繛鎴欏灩鍞銈嗘瀹曠敻寮弽顓熲拺婵懓娲ら埀顒佸姍瀹曟垿骞樼紒妯锋嫼缂佺虎鍘奸幊搴ㄋ夊澶嬬厵婵炶尪顔婄花鐣屸偓鍨緲閿曨亜鐣风粙璇炬棃鍩€椤掑嫬绀勯柣妯肩帛閻撳啰鎲稿鍫濈闁绘柨顨庡鏍磽娴ｈ偂鎴炲垔閹绢喗鐓熼柣鏃傚帶娴滀即鏌涢妶鍜佸剳缂佽鲸鎸婚幏鍛村礈閹绘帒澹夐梺姹囧焺閸ㄦ娊宕戦悢鐓庣闁圭儤鍨熼弸搴ㄦ煙閹咃紞妞ゆ挻妞藉娲箰鎼淬垻锛曢梺绋款儐閹瑰洭寮婚敐澶樻晣闁绘洑鐒﹂悿渚€姊洪崫鍕拱闁烩晩鍨堕悰顔嘉熸笟顖涘瘜闁荤姴娲╁鎾寸珶閺囥垺鈷掑ù锝勮閻掔偓銇勯幋婵囶棦妤犵偞鍨垮畷鍫曨敆閳ь剛绮婚弽顓熺厽闁硅揪绲鹃ˉ澶岀磼閻欐瑥娲﹂悡娆戠磼鐎ｎ亞浠㈡い鎺嬪灪閵囧嫰寮捄銊愌勬叏婵犲嫮甯涢柟宄版嚇瀹曟粓骞撻幒鎾充簻闂傚倷绀侀幖顐λ囬鐐村€舵繝闈涙处閸欏繘鏌ㄩ弴鐐测偓褰掑煕閹达附鍋ｉ柛銉岛閸嬫捇鎼归銈呯秵缂傚倸鍊烽懗鑸垫叏閻㈠憡鍋嬫俊銈勮兌閳瑰秴鈹戦悩鍙夊闁稿﹤顭烽弻锕€螣閻氬绀嗛悷婊勬瀵鈽夐姀鐘栥劑鏌曡箛濠傚⒉闁绘繃娲熷娲传閵夈儲鐎鹃梺鐟版啞婵炲﹤顕ｆ繝姘櫜闁糕剝锚閸斿懘姊洪棃娑氱畾闁哄懏鐩鏌ユ焼瀹ュ棌鎷洪梺鍛婄缚閸庤鲸鐗庨梻浣虹帛椤ㄥ牊绻涢埀顒勬煟濞戝崬娅嶇€殿喕绮欓、妯款槼闁哄懏绻堝娲箰鎼淬垻锛曢梺绋款儐閹瑰洭寮婚敓鐘茬闁宠桨绀佹慨鏇熺箾閿濆懏鎼愰柨鏇ㄤ簼娣囧﹪宕奸弴鐐殿槹濡炪倖姊婚崢褏鏁鐐粹拻?false闂?		// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌熼梻瀵割槮缁炬儳缍婇弻鐔兼⒒鐎靛壊妲梺姹囧€ら崰妤呭Φ閸曨垰绠涢柍杞拌兌閸旀挳姊洪幖鐐测偓鏍洪悢鐓庤摕婵炴垯鍩勯弫鍐煥濠靛棙顥滄い锔规櫊濮?admin 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣椤愯姤鎱ㄥ鍡楀⒒闁绘帊绮欓弻銈嗘叏閹邦兘鍋撻弽顐熷亾濮橆剦鐓奸柡宀嬬秮瀵噣宕掑顑跨帛缂傚倷璁查崑鎾愁熆鐠哄搫顦柛瀣崌瀹曟寰勬繝浣割棜闂傚倷绀侀幉鈥趁洪敃鍌氱；闁告洦鍨遍崐鍨亜閹惧崬鐏い銉ワ攻閵囧嫰骞囬埡浣轰痪婵犮垼娉涚€氼喚妲愰幒鏂哄亾閿濆骸浜滈柣蹇婃櫇缁?policy 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏℃櫇闁逞屽墴閹潡顢氶埀顒勫蓟閻旂厧绠氶柣妤€鐗滃Λ鍕⒑閸濆嫬顏ラ柛搴ｆ暬瀵鍨鹃幇浣告倯闂佸憡鍔戦崝宀勨€栭崱娆戠＝?bypass 濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴ｆ閺嬩線鏌熼梻瀵割槮缁惧墽绮换娑㈠箣濞嗗繒浠鹃梺绋款儏椤戝寮诲☉銏犵労闁告剬鍛畳闂備浇妗ㄧ粈浣衡偓姘嵆瀵鈽夐姀鐘栥劍銇勯弮鍌氬付妞ゎ偀鏅涢埞鎴︻敊濞嗙偓缍堥梺缁樻惈缁绘繂顕ｇ拠宸悑闁割偒鍋呴鍥⒒娴ｅ憡鍟為柟鎼佺畺瀹曚即寮借閸?DB 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｉ幇顒佹儓闁搞劌鍊块弻娑㈩敃閿濆棛顦ョ紓浣哄Т缂嶅﹪寮诲澶婁紶闁告洦鍓欏▍锝夋⒑濞茶骞楁い銊ワ躬閹繝顢曢敃鈧悙濠囨煏婵犲繐顩い锔哄妼椤啴濡舵惔鈥茬凹濠电偛寮堕…鍥箲閵忕姭妲堥柕蹇曞Т閼板灝鈹戦埥鍡楃仴妞ゆ泦鍛瀳鐎广儱娲ㄧ壕钘壝归敐鍡楃祷濞存粓绠栧娲礈閹绘帊绨梺鐟板暱濞差厼鐣烽幋锕€绠婚柤濮愬€曠粊锕傛⒑閸撹尙鍘涢柛鐘崇墬閹便劍鎯旈埦鈧弨浠嬫煟閹邦剛鎽犵紓宥嗗灴閺屾稒绻濋崒娑滅闂?UI 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｉ幇顒佹儓闁搞劌鍊块幃瑙勬姜閹峰矈鍔呭┑鐐插悑閻楁洟鍩為幋锔藉亹閻庡湱濮撮ˉ婵堢磽娴ｇ懓濮堟い銊ワ躬閻涱噣寮介‖銉ラ叄椤㈡鍩€椤掍椒绻嗛柤娴嬫櫇绾惧ジ鏌嶈閸撶喎鐣烽悢纰辨晬婵炴垶鑹惧鎶芥⒒娴ｈ櫣甯涢柡灞诲姂楠炴牠顢曢敐鍥р叞闂傚倸鍊烽悞锔界箾婵犲洤缁╅柛妤冧紳濞差亝鍋勯柣鎾虫唉閹芥洟姊虹紒妯荤叆闁告艾顑夐幃鐐寸鐎ｎ偆鍙嗛梺鍝勬处椤ㄥ懏绂嶆ィ鍐┾拺閻犲洩灏欑粻姘舵煕閹惧绠炲┑锛勬暬瀹曠喖顢涘В绗哄姂閺屻劑寮村Δ鈧禍楣冩⒑鐠囪尙绠查柟鍛婂▕瀵鎮㈢喊杈ㄦ櫓闂佸吋绁撮弲婵囶殭闂佽姘﹂～澶娒洪弽顓熷亯闁稿繘妫跨换鍡涙煟閹达絾顥夐崬顖炴⒑闂堟侗妲堕柛搴℃惈椤洤鈽夊杈╋紳闂佺鏈懝楣冨焵椤掆偓閸㈡煡婀侀梺鎼炲劘閸斿﹥绂掗崼銉︾厵闂傚倸顕ˇ锕傛煕閵娿儱鈧潡寮诲☉鈶┾偓锕傚箣濠靛洨浜鹃梻浣告惈濡挳姊介崟顒傗攳濠电姴娲ゅ洿闂佸憡渚楅崢钘夆枔閼稿吀绻嗛柣鎰邦杺濞兼劕鈹戦悙璇ц含鐎殿喛顕ч埥澶愬閳╁啯鐝曢梺鑽ゅТ濞诧箒銇愰崘顏嗕笉閻犲洤妯婂〒濠氭煏閸繄绠伴柣锔界矒閺屾稓鈧綆浜峰銉︺亜?
		if dingTalkCfg.CorpRestrictionPolicy != "internal_only" {
			req.DingTalkConnectBypassRegistration = false
			// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌熼梻纾嬵唹闁逞屽墾缁犳捇銆佸Δ鍛妞ゆ劑鍊ゅΣ鍫曟⒒娴ｇ懓顕滄繛鍙夌墵瀹曟劙宕妷銉婵犵數濮电喊宥夋偂閺囩姭鍋撻崗澶婁壕闂侀€炲苯澧寸€规洘鍨甸埥澶愬閳ヨ櫕鐤傞柣鐔哥矌婢ф鏁Δ鍛亗闊洦绋掗悡鏇熴亜閹板墎鎮肩紒鐘劜娣囧﹪宕ｆ径瀣偓鎰版煙椤旂厧妲婚柍璇查閳诲酣骞嬮悩铏啌闂傚倷娴囬鏍窗濡ゅ懏鏅濋柕蹇嬪€曠粻鏍ㄧ箾閸℃ê濮堝┑顖涙尦閺屾盯骞囬妸锔界彟闂侀€炲苯澧い銊ワ工椤繐煤椤忓嫬绐涙繝鐢靛Т閸熶即宕ｉ崱娑欑厽閹兼番鍨归崵顒勬煕閹捐泛鏋涚€规洘妞芥俊鐑芥晝閳ь剛娆㈤悙鐑樼厵闂侇叏绠戞晶鐗堛亜閺冣偓鐢€愁潖妤﹁￥浜归柟鐑樼箖濞堛儵姊虹紒妯洪嚋濠碘€虫喘閹柉顦寸紒杈ㄦ崌瀹曟帒顫濋钘変壕鐎瑰嫭鍣磋ぐ鎺戠倞闁靛绲肩划鎾绘⒑瑜版帗锛熺紒鈧笟鈧畷褰掑磼閻愬鍘卞銈庡幗閸ㄧ敻寮搁妶澶嬬厸闁糕槅鍙冨顔记庨崶褝韬┑鈥崇埣瀹曘劑顢欓崗纰变哗缂傚倸鍊风拋鏌ュ磻閹剧粯鐓曢柟浼存涧閺嬬喖鏌ｉ幘瀛樼缂佺粯鐩畷鍗炍旀担渚炊婵犵鈧啿绾ч柟顔煎€垮璇差吋閸偅顎囬梻浣告啞閹搁箖宕伴幇顔剧焿鐎广儱顦柋鍥煛閸モ晛浠╅柟鑺ユ礋濮婃椽骞愭惔锝囩暤濡炪倧缂氶崡鍐茬暦閹扮増鍊婚柤鎭掑劤閸樺崬鈹戦悙鍙夘棞婵炲瓨鑹鹃‖濠咁槾缂佽鲸甯楃粭鐔煎垂椤旂⒈鐎风紓鍌欐祰妞存悂骞愭繝姘闁告稒娼欑粻锝夋煙閻戞ê鐏﹂柛搴㈡尭閳规垿鎮╅锝咁€忛梺鍛婃礀閻忔岸鎮鹃懖鈺冪＝?internal_only 濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴濐潟閳ь剙鍊块幐濠冪珶閳哄绉€规洏鍔戝鍫曞箣濠靛牃鍋撻鐑嗘富闁靛牆鎳愮粻浼存煟濡も偓濡繈骞冮悙鍝勫瀭妞ゆ劗濮崇花濠氭⒑閻熺増鎯堟俊顐ｎ殕缁傚秹宕滆绾惧ジ鏌涢幘妤€妫欓妤呮⒑閸涘鎴﹀磿閹惰棄鐓″鑸靛姈閸嬨劎绱掔€ｎ厽纭堕柛鏃撶畱椤啴濡堕崱妤冪憪闂佺厧鍟挎晶搴ｅ垝缂佹ǜ鍋呴柛鎰ㄦ櫇閸樼敻鏌ｆ惔锝嗘毄妞ゎ厼鐗婄粋鎺楁晝閸屾稑鈧爼鐓崶椋庡埌妤犵偞顨婇弻鏇㈠炊瑜嶉顓燁殽閻愬樊鍎忛柍瑙勫灴瀹曞爼鎳滃▓鎸庯紘闂傚倸鍊烽悞锔锯偓绗涘厾楦跨疀濞戞ê鐎梺鎯х箰濠€杈╃不閺冨牊鐓熼柡鍐ㄥ€哥敮鍫曟煟閹烘垹浠涢柕鍥у楠炴帡骞嬪┑鎰磻闁诲氦顫夐幐椋庢濮樿泛钃熸繛鎴欏灩鍞銈嗘瀹曠敻寮弽顓熲拺婵懓娲ら埀顒佸姍瀹曟垿骞樼紒妯锋嫼缂佺虎鍘奸幊搴ㄋ夊澶嬬厵婵炶尪顔婄花鐣屸偓鍨緲閿曨亜鐣风粙璇炬棃鍩€椤掑嫬绀勯柣妯肩帛閻撳啰鎲稿鍫濈闁绘柨顨庡鏍磽娴ｈ偂鎴炲垔閹绢喗鐓熼柣鏃傚帶娴滀即鏌涢妶鍜佸剳缂佽鲸鎸婚幏鍛村礈閹绘帒澹夐梻浣规偠閸斿苯顭囬敓鐘靛祦濠电姴娲﹂崑鍕棯閹峰矂鍝洪柡鍜佸墴濮婅櫣绱掑Ο鍝勵潙闁诲繐绻戦悷鈺呭箖閿熺姴绠ｉ柨鏃囆掗幏娲⒑閸涘﹦鈽夐柨鏇樺劤閳ь剙鐏氶悷鈺呭蓟瀹ュ牜妾ㄩ梺鍛婃尵閸犲酣鎮鹃悜钘夌闁挎洍鍋撶紒鐘哄吹缁辨挻鎷呮慨鎴簼缁傛帡顢橀姀鈾€鎷洪梺鍛婄箓鐎氬嘲危瑜版帗鍊电紒妤佺☉閸熺娀寮搁弮鍫熺厱妞ゆ劧绲剧粈鍐煃闁垮鐏撮柟顔筋殜閹倿骞栨担璇♀偓宥呪攽閻橆喖鐏柨鏇樺灩閻ｅ嘲鈹戠€ｎ偅娅囬梺绋挎湰缁嬫捇宕㈤棃娑辨富闁靛牆妫欑壕鐢告煕鐎ｎ偅灏电紒杈ㄥ浮閹晛鐣烽崶褜娼峰┑鐑囩到濞层倝鏁冮鍫濈畺婵犲﹤鐗婄€电姴顭块崗澶嬫澓闁告梹鍨垮濠氬即閻旇櫣褰惧銈嗙墬绾板秹宕虫禒瀣厸闁糕剝鍔曢悘鍙夋叏?false闂?			req.DingTalkConnectSyncCorpEmail = false
			req.DingTalkConnectSyncDisplayName = false
			req.DingTalkConnectSyncDept = false
		}
		// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌熼梻纾嬵唹闁逞屽墾缁犳捇銆佸Δ鍛妞ゆ劑鍊ゅΣ鍫曟⒒娴ｇ懓顕滄繛鍙夌墵瀹曟劙宕妷銉婵犵數濮电喊宥夋偂閺囩姭鍋撻崗澶婁壕闂侀€炲苯澧寸€规洘鍨甸埥澶愬閳ヨ櫕鐤傞柣鐔哥矌婢ф鏁Δ鍛亗闊洦绋掗悡鏇熴亜閹板墎鎮肩紒鐘劜娣囧﹪宕ｆ径瀣偓鎰版煙椤旂厧妲婚柍璇查閳诲酣骞嬮悩铏啌闂傚倷娴囬鏍窗濡ゅ懏鏅濋柕蹇嬪€曠粻鏍ㄧ箾閸℃ê濮堝┑顖涙尦閺屾盯骞囬妸锔界彟闂侀€炲苯澧い銊ワ工椤繐煤椤忓嫬绐涙繝鐢靛Т閸熶即宕ｉ崱娑欑厽閹兼番鍨婚悡顖炴煕濡も偓閸熷潡顢氶敐鍡欑瘈婵﹩鍘藉▍婊勭節閵忥絾纭鹃柨鏇畵瀹曘垽宕ㄦ繝鍕槇闂佹眹鍨藉褍鐡梻浣侯焾閿曘倗绱炴繝鍥х畺婵☆垯璀﹂崥瀣熆鐠轰警鍎岄柟?attr key闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍏煎€绘慨妤€妫欓褰掓⒑鏉炴壆顦︽俊顐ｇ箞瀵鏁愭径瀣珳闂佸憡渚楅崰妤呭汲瀵槩pace + 缂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亝鎹ｉ柣顓炴閵嗘帒顫濋敐鍛闂備胶鎳撶壕顓㈠窗閺嶎厹鈧礁鈽夐姀鈥斥偓鐑芥煠绾板崬澧婚柛鐔锋处娣囧﹪鎮欓鍕ㄥ亾閺嶎厼绀夌憸蹇曞垝婵犳艾绠ｉ柨鏃囨閳?fallback 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极瀹ュ绀嬫い鎺嶇劍椤斿洤鈹戦悩鍨毄濠殿喗娼欑叅闁靛牆娲﹂崗婊堟煃瑜滈崜鐔煎箖濡ゅ啯鍠嗛柛鏇ㄥ墰椤︺儱鈹戦悙鑼勾闁稿﹥绻傞悾鐑芥偡閹佃櫕鏂€闂佹悶鍎崕鎵姳婵犳碍鈷戦柛婵嗗琚梺鍛婃煥闁帮綁骞嗗畝鈧埀顒婄秵娴滄牠寮ㄩ懞銉ｄ簻闁哄秲鍔岄幖鎼佹煕鐎ｃ劌鐏查柡宀嬬磿娴狅箓鎮欓鍌ゆЧ闁诲氦顫夊ú婊堝储瑜旈崺鐐哄箣閿曗偓閻愬﹦鎲告惔銊﹀仧闁圭虎鍠楅埛?		req.DingTalkConnectSyncCorpEmailAttrKey = strings.TrimSpace(req.DingTalkConnectSyncCorpEmailAttrKey)
		if req.DingTalkConnectSyncCorpEmailAttrKey == "" {
			req.DingTalkConnectSyncCorpEmailAttrKey = "dingtalk_email"
		}
		req.DingTalkConnectSyncDisplayNameAttrKey = strings.TrimSpace(req.DingTalkConnectSyncDisplayNameAttrKey)
		if req.DingTalkConnectSyncDisplayNameAttrKey == "" {
			req.DingTalkConnectSyncDisplayNameAttrKey = "dingtalk_name"
		}
		req.DingTalkConnectSyncDeptAttrKey = strings.TrimSpace(req.DingTalkConnectSyncDeptAttrKey)
		if req.DingTalkConnectSyncDeptAttrKey == "" {
			req.DingTalkConnectSyncDeptAttrKey = "dingtalk_department"
		}
		// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌熼梻纾嬵唹闁逞屽墾缁犳捇銆佸Δ鍛妞ゆ劑鍊ゅΣ鍫曟⒒娴ｇ懓顕滄繛鍙夌墵瀹曟劙宕妷銉婵犵數濮电喊宥夋偂閺囩姭鍋撻崗澶婁壕闂侀€炲苯澧寸€规洘鍨甸埥澶愬閳ヨ櫕鐤傞柣鐔哥矌婢ф鏁Δ鍛亗闊洦绋掗悡鏇熴亜閹板墎鎮肩紒鐘劜娣囧﹪宕ｆ径瀣偓鎰版煙椤旂厧妲婚柍璇查閳诲酣骞嬮悩铏啌闂傚倷娴囬鏍窗濡ゅ懏鏅濋柕蹇嬪€曠粻鏍ㄧ箾閸℃ê濮堝┑顖涙尦閺屾盯骞囬妸锔界彟闂侀€炲苯澧い銊ワ工椤繐煤椤忓嫬绐涙繝鐢靛Т閸熶即宕ｉ崱娑欑厽閹兼番鍨婚悡顖炴煕濡も偓閸熷潡顢氶敐鍡欑瘈婵﹩鍘藉▍婊勭節閵忥絾纭鹃柨鏇畵瀹曘垽宕ㄦ繝鍕槇闂佹眹鍨藉褍鐡梻浣侯焾閿曘倗绱炴繝鍥х畺婵☆垯璀﹂崥瀣熆鐠轰警鍎岄柟?attr 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犖ч柛灞剧煯婢规洖鈹戦绛嬬劷闁告鍕珷妞ゆ牜鍋為悡娑㈡煕濞戞瑦绶查柣鎾村姉缁辨帡宕掑☉妯碱儌闂侀€炲苯澧剧紓宥呮閸┾偓妞ゆ帒顦鍧楁⒒閸曨偄顏鐐茬墦婵℃悂濡锋惔锝呮灈妤犵偞鍔栭幆鏃堟晲閸愶絽浜鹃柟鎵閻撶喖鏌ｉ弬鎸庢喐闁硅櫕娲栭埞鎴﹀灳閾忣偄鏋犻梺绯曟杹閸嬫挸顪冮妶鍡楀潑闁稿鎸剧槐鎺楁偐閼碱儷褏鈧娲樺ú鐔风暦閿熺姵鍊烽柍鍝勫€绘禍娆撴⒒娴ｇ懓顕滄繛鍙夌墵瀹曟劘銇愰幒鎴狅紮闂佺粯鍨兼慨銈夋偂閻斿吋鐓欓悗鐢殿焾閳ь剚鐗犻獮瀣攽閸愨晝鈧椽姊洪崫鍕檨闁逞屽墴瀵劑鎳為妷锝勭盎闂佸搫鍟崐鐟扳枍閺囩姷纾奸柣妯诲灇閹达箑桅闁告洦鍨版儫闂佹寧娲嶉崑鎾绘煙椤曞懏娅?+ 缂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亝鎹ｉ柣顓炴閵嗘帒顫濋敐鍛闂備胶鎳撶壕顓㈠窗閺嶎厹鈧礁鈽夐姀鈥斥偓鐑芥煠绾板崬澧婚柛鐔锋处娣囧﹪鎮欓鍕ㄥ亾閺嶎厼绀夌憸蹇曞垝婵犳艾绠ｉ柨鏃囨閳?fallback 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极瀹ュ绀嬫い鎺嶇劍椤斿洤鈹戦悩鍨毄濠殿喗娼欑叅闁靛牆娲﹂崗婊堟煃瑜滈崜鐔煎箖濡ゅ啯鍠嗛柛鏇ㄥ墰椤︺儱鈹戦悙鑼勾闁稿﹥绻傞悾鐑芥偡閹佃櫕鏂€闂佹悶鍎崕鎵姳婵犳碍鈷戦柛婵嗗琚梺鍛婃煥闁帮綁骞嗗畝鈧埀顒婄秵娴滄牠寮ㄩ懞銉ｄ簻闁哄秲鍔岄幖鎼佹煕鐎ｃ劌鐏查柡宀嬬磿娴狅箓鎮欓鍌ゆЧ闁诲氦顫夊ú婊堝储瑜旈崺鐐哄箣閿曗偓楠炪垽鏌嶇悰鈥充壕缂備浇椴搁幑鍥ь潖缂佹ɑ濯撮柛娑橈工閺嗗牓姊洪幐搴㈢５闁哄懐濮撮悾宄邦煥閸喎浜滈梺缁樻尭濮橈箓骞楅弴銏♀拺閻犳亽鍔岄弸鏂库攽椤旂⒈鍤熷瑙勬礋瀹曟宕楅悡搴′紟婵犳鍠楅〃鍛涘Δ鍛祦婵°倕鎳忛悡娆愩亜閺冨倹娅曢柟鍐叉处閹?
		req.DingTalkConnectSyncCorpEmailAttrName = "DingTalk email"
		if req.DingTalkConnectSyncCorpEmailAttrName == "" {
			req.DingTalkConnectSyncCorpEmailAttrName = "DingTalk email"
		}
		req.DingTalkConnectSyncDisplayNameAttrName = "DingTalk display name"
		if req.DingTalkConnectSyncDisplayNameAttrName == "" {
			req.DingTalkConnectSyncDisplayNameAttrName = "DingTalk display name"
		}
		req.DingTalkConnectSyncDeptAttrName = "DingTalk department"
		if req.DingTalkConnectSyncDeptAttrName == "" {
			req.DingTalkConnectSyncDeptAttrName = "DingTalk department"
		}
	}

	if req.WeChatConnectEnabled {
		req.WeChatConnectAppID = strings.TrimSpace(req.WeChatConnectAppID)
		req.WeChatConnectAppSecret = strings.TrimSpace(req.WeChatConnectAppSecret)
		req.WeChatConnectOpenAppID = strings.TrimSpace(req.WeChatConnectOpenAppID)
		req.WeChatConnectOpenAppSecret = strings.TrimSpace(req.WeChatConnectOpenAppSecret)
		req.WeChatConnectMPAppID = strings.TrimSpace(req.WeChatConnectMPAppID)
		req.WeChatConnectMPAppSecret = strings.TrimSpace(req.WeChatConnectMPAppSecret)
		req.WeChatConnectMobileAppID = strings.TrimSpace(req.WeChatConnectMobileAppID)
		req.WeChatConnectMobileAppSecret = strings.TrimSpace(req.WeChatConnectMobileAppSecret)
		req.WeChatConnectMode = strings.ToLower(strings.TrimSpace(req.WeChatConnectMode))
		req.WeChatConnectScopes = strings.TrimSpace(req.WeChatConnectScopes)
		req.WeChatConnectRedirectURL = strings.TrimSpace(req.WeChatConnectRedirectURL)
		req.WeChatConnectFrontendRedirectURL = strings.TrimSpace(req.WeChatConnectFrontendRedirectURL)
		req.WeChatConnectAppID = strings.TrimSpace(firstNonEmpty(req.WeChatConnectAppID, previousSettings.WeChatConnectAppID))
		req.WeChatConnectRedirectURL = strings.TrimSpace(firstNonEmpty(req.WeChatConnectRedirectURL, previousSettings.WeChatConnectRedirectURL))
		req.WeChatConnectFrontendRedirectURL = strings.TrimSpace(firstNonEmpty(req.WeChatConnectFrontendRedirectURL, previousSettings.WeChatConnectFrontendRedirectURL))
		if req.WeChatConnectMode == "" {
			req.WeChatConnectMode = strings.ToLower(strings.TrimSpace(previousSettings.WeChatConnectMode))
		}
		if req.WeChatConnectScopes == "" {
			req.WeChatConnectScopes = strings.TrimSpace(previousSettings.WeChatConnectScopes)
		}

		if req.WeChatConnectMPEnabled && req.WeChatConnectMobileEnabled {
			response.BadRequest(c, "WeChat Official Account and Mobile App cannot be enabled at the same time")
			return
		}
		if req.WeChatConnectMode != "" {
			switch req.WeChatConnectMode {
			case "open", "mp", "mobile":
			default:
				response.BadRequest(c, "WeChat mode must be open, mp, or mobile")
				return
			}
		}
		if !req.WeChatConnectOpenEnabled && !req.WeChatConnectMPEnabled && !req.WeChatConnectMobileEnabled {
			switch req.WeChatConnectMode {
			case "mp":
				req.WeChatConnectMPEnabled = true
			case "mobile":
				req.WeChatConnectMobileEnabled = true
			default:
				req.WeChatConnectOpenEnabled = true
			}
		}
		if req.WeChatConnectMode == "" {
			if req.WeChatConnectMPEnabled {
				req.WeChatConnectMode = "mp"
			} else if req.WeChatConnectMobileEnabled {
				req.WeChatConnectMode = "mobile"
			} else {
				req.WeChatConnectMode = "open"
			}
		}

		req.WeChatConnectOpenAppID = strings.TrimSpace(firstNonEmpty(req.WeChatConnectOpenAppID, req.WeChatConnectAppID, previousSettings.WeChatConnectOpenAppID, previousSettings.WeChatConnectAppID))
		req.WeChatConnectMPAppID = strings.TrimSpace(firstNonEmpty(req.WeChatConnectMPAppID, req.WeChatConnectAppID, previousSettings.WeChatConnectMPAppID, previousSettings.WeChatConnectAppID))
		req.WeChatConnectMobileAppID = strings.TrimSpace(firstNonEmpty(req.WeChatConnectMobileAppID, req.WeChatConnectAppID, previousSettings.WeChatConnectMobileAppID, previousSettings.WeChatConnectAppID))

		if req.WeChatConnectOpenAppSecret == "" {
			req.WeChatConnectOpenAppSecret = strings.TrimSpace(firstNonEmpty(previousSettings.WeChatConnectOpenAppSecret, previousSettings.WeChatConnectAppSecret, req.WeChatConnectAppSecret))
		}
		if req.WeChatConnectMPAppSecret == "" {
			req.WeChatConnectMPAppSecret = strings.TrimSpace(firstNonEmpty(previousSettings.WeChatConnectMPAppSecret, previousSettings.WeChatConnectAppSecret, req.WeChatConnectAppSecret))
		}
		if req.WeChatConnectMobileAppSecret == "" {
			req.WeChatConnectMobileAppSecret = strings.TrimSpace(firstNonEmpty(previousSettings.WeChatConnectMobileAppSecret, previousSettings.WeChatConnectAppSecret, req.WeChatConnectAppSecret))
		}
		if req.WeChatConnectAppSecret == "" {
			req.WeChatConnectAppSecret = strings.TrimSpace(firstNonEmpty(req.WeChatConnectOpenAppSecret, req.WeChatConnectMPAppSecret, req.WeChatConnectMobileAppSecret, previousSettings.WeChatConnectAppSecret))
		}

		if req.WeChatConnectOpenEnabled {
			if req.WeChatConnectOpenAppID == "" {
				response.BadRequest(c, "WeChat PC App ID is required when enabled")
				return
			}
			if req.WeChatConnectOpenAppSecret == "" {
				response.BadRequest(c, "WeChat PC App Secret is required when enabled")
				return
			}
		}
		if req.WeChatConnectMPEnabled {
			if req.WeChatConnectMPAppID == "" {
				response.BadRequest(c, "WeChat Official Account App ID is required when enabled")
				return
			}
			if req.WeChatConnectMPAppSecret == "" {
				response.BadRequest(c, "WeChat Official Account App Secret is required when enabled")
				return
			}
		}
		if req.WeChatConnectMobileEnabled {
			if req.WeChatConnectMobileAppID == "" {
				response.BadRequest(c, "WeChat Mobile App ID is required when enabled")
				return
			}
			if req.WeChatConnectMobileAppSecret == "" {
				response.BadRequest(c, "WeChat Mobile App Secret is required when enabled")
				return
			}
		}

		if req.WeChatConnectScopes == "" {
			if req.WeChatConnectMPEnabled {
				req.WeChatConnectScopes = service.DefaultWeChatConnectScopesForMode("mp")
			} else {
				req.WeChatConnectScopes = service.DefaultWeChatConnectScopesForMode(req.WeChatConnectMode)
			}
		}
		if req.WeChatConnectOpenEnabled || req.WeChatConnectMPEnabled {
			if req.WeChatConnectRedirectURL == "" {
				response.BadRequest(c, "WeChat Redirect URL is required when web oauth is enabled")
				return
			}
			if err := config.ValidateAbsoluteHTTPURL(req.WeChatConnectRedirectURL); err != nil {
				response.BadRequest(c, "WeChat Redirect URL must be an absolute http(s) URL")
				return
			}
			if req.WeChatConnectFrontendRedirectURL == "" {
				req.WeChatConnectFrontendRedirectURL = "/auth/wechat/callback"
			}
			if err := config.ValidateFrontendRedirectURL(req.WeChatConnectFrontendRedirectURL); err != nil {
				response.BadRequest(c, "WeChat Frontend Redirect URL is invalid")
				return
			}
		}
	}

	// Generic OIDC 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柛鎰紦閻㈠姊绘笟鈧褔藝椤愶箑鐤炬繛鎴炶壘椤ユ岸鏌涢敂璇插箺闁哥姵鍔欓弻锝呂旈埀顒勬偋閸℃瑧绠旈柟鐑樻尰閸欏繘鎮峰▎蹇擃伀闁告瑢鍋撻梻浣告惈閻绱炴担鍓插殨妞ゆ帒瀚崹鍌涖亜閺囩偞鍣烘い銉﹀灦缁绘繈鎮介棃娑楃捕闂佺粯顨呯换妯侯嚕閺屻儺鏁冮柨鏇楀亾闁绘帒鐏氶妵鍕箳閸℃ぞ澹曢梻浣告啞钃辩紒顔芥尭閻ｇ兘骞嬮敃鈧粻鑽ょ磽娴ｈ鐒介柛姗€浜跺铏规喆閸曨剛鍑￠梺鍛婂焹閸嬫挾绱撴担闈涘闁告艾顑夋俊?
	oidcUsePKCE, oidcValidateIDToken, err := h.settingService.OIDCSecurityWriteDefaults(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if req.OIDCConnectEnabled {
		req.OIDCConnectProviderName = strings.TrimSpace(req.OIDCConnectProviderName)
		req.OIDCConnectClientID = strings.TrimSpace(req.OIDCConnectClientID)
		req.OIDCConnectClientSecret = strings.TrimSpace(req.OIDCConnectClientSecret)
		req.OIDCConnectIssuerURL = strings.TrimSpace(req.OIDCConnectIssuerURL)
		req.OIDCConnectDiscoveryURL = strings.TrimSpace(req.OIDCConnectDiscoveryURL)
		req.OIDCConnectAuthorizeURL = strings.TrimSpace(req.OIDCConnectAuthorizeURL)
		req.OIDCConnectTokenURL = strings.TrimSpace(req.OIDCConnectTokenURL)
		req.OIDCConnectUserInfoURL = strings.TrimSpace(req.OIDCConnectUserInfoURL)
		req.OIDCConnectJWKSURL = strings.TrimSpace(req.OIDCConnectJWKSURL)
		req.OIDCConnectScopes = strings.TrimSpace(req.OIDCConnectScopes)
		req.OIDCConnectRedirectURL = strings.TrimSpace(req.OIDCConnectRedirectURL)
		req.OIDCConnectFrontendRedirectURL = strings.TrimSpace(req.OIDCConnectFrontendRedirectURL)
		req.OIDCConnectTokenAuthMethod = strings.ToLower(strings.TrimSpace(req.OIDCConnectTokenAuthMethod))
		req.OIDCConnectAllowedSigningAlgs = strings.TrimSpace(req.OIDCConnectAllowedSigningAlgs)
		req.OIDCConnectUserInfoEmailPath = strings.TrimSpace(req.OIDCConnectUserInfoEmailPath)
		req.OIDCConnectUserInfoIDPath = strings.TrimSpace(req.OIDCConnectUserInfoIDPath)
		req.OIDCConnectUserInfoUsernamePath = strings.TrimSpace(req.OIDCConnectUserInfoUsernamePath)
		req.OIDCConnectProviderName = strings.TrimSpace(firstNonEmpty(req.OIDCConnectProviderName, previousSettings.OIDCConnectProviderName, "OIDC"))
		req.OIDCConnectClientID = strings.TrimSpace(firstNonEmpty(req.OIDCConnectClientID, previousSettings.OIDCConnectClientID))
		req.OIDCConnectIssuerURL = strings.TrimSpace(firstNonEmpty(req.OIDCConnectIssuerURL, previousSettings.OIDCConnectIssuerURL))
		req.OIDCConnectDiscoveryURL = strings.TrimSpace(firstNonEmpty(req.OIDCConnectDiscoveryURL, previousSettings.OIDCConnectDiscoveryURL))
		req.OIDCConnectAuthorizeURL = strings.TrimSpace(firstNonEmpty(req.OIDCConnectAuthorizeURL, previousSettings.OIDCConnectAuthorizeURL))
		req.OIDCConnectTokenURL = strings.TrimSpace(firstNonEmpty(req.OIDCConnectTokenURL, previousSettings.OIDCConnectTokenURL))
		req.OIDCConnectUserInfoURL = strings.TrimSpace(firstNonEmpty(req.OIDCConnectUserInfoURL, previousSettings.OIDCConnectUserInfoURL))
		req.OIDCConnectJWKSURL = strings.TrimSpace(firstNonEmpty(req.OIDCConnectJWKSURL, previousSettings.OIDCConnectJWKSURL))
		req.OIDCConnectScopes = strings.TrimSpace(firstNonEmpty(req.OIDCConnectScopes, previousSettings.OIDCConnectScopes, "openid email profile"))
		req.OIDCConnectRedirectURL = strings.TrimSpace(firstNonEmpty(req.OIDCConnectRedirectURL, previousSettings.OIDCConnectRedirectURL))
		req.OIDCConnectFrontendRedirectURL = strings.TrimSpace(firstNonEmpty(req.OIDCConnectFrontendRedirectURL, previousSettings.OIDCConnectFrontendRedirectURL, "/auth/oidc/callback"))
		req.OIDCConnectTokenAuthMethod = strings.ToLower(strings.TrimSpace(firstNonEmpty(req.OIDCConnectTokenAuthMethod, previousSettings.OIDCConnectTokenAuthMethod, "client_secret_post")))
		req.OIDCConnectAllowedSigningAlgs = strings.TrimSpace(firstNonEmpty(req.OIDCConnectAllowedSigningAlgs, previousSettings.OIDCConnectAllowedSigningAlgs, "RS256,ES256,PS256"))
		req.OIDCConnectUserInfoEmailPath = strings.TrimSpace(firstNonEmpty(req.OIDCConnectUserInfoEmailPath, previousSettings.OIDCConnectUserInfoEmailPath))
		req.OIDCConnectUserInfoIDPath = strings.TrimSpace(firstNonEmpty(req.OIDCConnectUserInfoIDPath, previousSettings.OIDCConnectUserInfoIDPath))
		req.OIDCConnectUserInfoUsernamePath = strings.TrimSpace(firstNonEmpty(req.OIDCConnectUserInfoUsernamePath, previousSettings.OIDCConnectUserInfoUsernamePath))
		if req.OIDCConnectUsePKCE != nil {
			oidcUsePKCE = *req.OIDCConnectUsePKCE
		}
		if req.OIDCConnectValidateIDToken != nil {
			oidcValidateIDToken = *req.OIDCConnectValidateIDToken
		}
		if req.OIDCConnectClockSkewSeconds == 0 {
			req.OIDCConnectClockSkewSeconds = previousSettings.OIDCConnectClockSkewSeconds
			if req.OIDCConnectClockSkewSeconds == 0 {
				req.OIDCConnectClockSkewSeconds = 120
			}
		}

		if req.OIDCConnectClientID == "" {
			response.BadRequest(c, "OIDC Client ID is required when enabled")
			return
		}
		if req.OIDCConnectIssuerURL == "" {
			response.BadRequest(c, "OIDC Issuer URL is required when enabled")
			return
		}
		if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectIssuerURL); err != nil {
			response.BadRequest(c, "OIDC Issuer URL must be an absolute http(s) URL")
			return
		}
		if req.OIDCConnectDiscoveryURL != "" {
			if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectDiscoveryURL); err != nil {
				response.BadRequest(c, "OIDC Discovery URL must be an absolute http(s) URL")
				return
			}
		}
		if req.OIDCConnectAuthorizeURL != "" {
			if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectAuthorizeURL); err != nil {
				response.BadRequest(c, "OIDC Authorize URL must be an absolute http(s) URL")
				return
			}
		}
		if req.OIDCConnectTokenURL != "" {
			if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectTokenURL); err != nil {
				response.BadRequest(c, "OIDC Token URL must be an absolute http(s) URL")
				return
			}
		}
		if req.OIDCConnectUserInfoURL != "" {
			if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectUserInfoURL); err != nil {
				response.BadRequest(c, "OIDC UserInfo URL must be an absolute http(s) URL")
				return
			}
		}
		if req.OIDCConnectRedirectURL == "" {
			response.BadRequest(c, "OIDC Redirect URL is required when enabled")
			return
		}
		if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectRedirectURL); err != nil {
			response.BadRequest(c, "OIDC Redirect URL must be an absolute http(s) URL")
			return
		}
		if req.OIDCConnectFrontendRedirectURL == "" {
			response.BadRequest(c, "OIDC Frontend Redirect URL is required when enabled")
			return
		}
		if err := config.ValidateFrontendRedirectURL(req.OIDCConnectFrontendRedirectURL); err != nil {
			response.BadRequest(c, "OIDC Frontend Redirect URL is invalid")
			return
		}
		if !scopesContainOpenID(req.OIDCConnectScopes) {
			response.BadRequest(c, "OIDC scopes must contain openid")
			return
		}
		switch req.OIDCConnectTokenAuthMethod {
		case "", "client_secret_post", "client_secret_basic", "none":
		default:
			response.BadRequest(c, "OIDC Token Auth Method must be one of client_secret_post/client_secret_basic/none")
			return
		}
		if req.OIDCConnectClockSkewSeconds < 0 || req.OIDCConnectClockSkewSeconds > 600 {
			response.BadRequest(c, "OIDC clock skew seconds must be between 0 and 600")
			return
		}
		if oidcValidateIDToken && req.OIDCConnectAllowedSigningAlgs == "" {
			response.BadRequest(c, "OIDC Allowed Signing Algs is required when validate_id_token=true")
			return
		}
		if req.OIDCConnectJWKSURL != "" {
			if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectJWKSURL); err != nil {
				response.BadRequest(c, "OIDC JWKS URL must be an absolute http(s) URL")
				return
			}
		}
		if req.OIDCConnectTokenAuthMethod == "" || req.OIDCConnectTokenAuthMethod == "client_secret_post" || req.OIDCConnectTokenAuthMethod == "client_secret_basic" {
			if req.OIDCConnectClientSecret == "" {
				if previousSettings.OIDCConnectClientSecret == "" {
					response.BadRequest(c, "OIDC Client Secret is required when enabled")
					return
				}
				req.OIDCConnectClientSecret = previousSettings.OIDCConnectClientSecret
			}
		}
	}

	purchaseEnabled := previousSettings.PurchaseSubscriptionEnabled
	if req.PurchaseSubscriptionEnabled != nil {
		purchaseEnabled = *req.PurchaseSubscriptionEnabled
	}
	purchaseURL := previousSettings.PurchaseSubscriptionURL
	if req.PurchaseSubscriptionURL != nil {
		purchaseURL = strings.TrimSpace(*req.PurchaseSubscriptionURL)
	}

	// - 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柤纰卞墻濡查亶鏌ｉ悢鍝ョ煂濠⒀勵殘閺侇噣骞掑Δ鈧崹鍌炴煕瀹€鈧崑鐐烘偂濞嗘挻鐓欐い鏍ㄧ矊椤ｅ吋銇勯妷銉█闁哄矉绱曢埀顒婄岛閺呮繄绮ｉ弮鍫熺厸閻忕偟鍋撶粈瀣偓瑙勬礈閸樠囧煘閹达箑绠涙い鎾愁檧婵″洨妲愰幘瀵哥懝闁搞儜鍌滅泿缂傚倷绀侀ˇ顖滅礊婵犲偆鍤曞┑鐘崇閸嬪嫮鐥幏宀勫摵闁哄拋鍓熷铏圭磼濡搫顫戦悗娈垮枟閹瑰洤鐣烽悧鍫㈢瘈闁稿本顨嗛弬鈧梻浣虹帛閸旀瑩路閸岀偛鍚归悗锝庡墰绾惧ジ鏌＄仦璇插姢闁伙絿鍎ら〃?URL 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柛鎰级閻ゅ嫬鈹戞幊閸娧呭緤娴犲鐤い鏍仜绾惧鎮楅敐搴濈按闁衡偓娴犲鐓熸俊顖濐嚙缁茬粯銇勮箛锝勯偗闁哄本绋栫粻娑㈠籍閸屾粎鍘滈梻浣告惈鐞氼偊宕愬┑鍡欐殾妞ゆ劧绠戝敮闂佹寧妫佹慨銈夋儊鎼淬劍鈷掑ù锝呮憸閿涘秶绱掗鍛仩閾荤偤鏌涢幇闈涙灈闁藉啰鍠栭弻銊╂偄閸濆嫅銏㈢磼閳ь剟宕橀鐣屽帾婵犵數鍊崘銊︾亾闂佸搫顦遍…鍫ュ煘閹达附鍋愰柟棰佺閺呴亶姊洪崫銉バｆい銊ョ墦瀹曟岸骞掗幘鏉戝妳闂佹寧绻傚Λ顓烆焽?	// - 缂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亝鎹ｉ柣顓炴閵嗘帒顫濋敐鍛闁诲氦顫夊ú姗€宕濆▎鎾崇畺婵犲﹤鐗婇崵宥夋煏婢跺牆鍔滈柣锝変憾濮婄粯鎷呯粙娆炬闂佺顑呴幊搴ｅ弲闂佸搫绋侀崢濂稿礄閻樺眰鈧帒顫濋敐鍛闂備礁鎼張顒€煤閻旈鏆﹂柣鎾崇岸閺€浠嬫煙闁箑甯ㄩ柨鏂垮⒔绾捐棄銆掑顒佹悙濞存粍绮庣槐鎺撳緞婵犲嫮楔濡ょ姷鍋涢崯鎶藉Φ閹版澘绠抽柟瀛樼矌閻愬﹪姊绘担鍛婂暈婵炶绠撳畷婊冾潩鐠鸿櫣锛熼梺鐟板⒔缁垶鎮￠悢鍏肩厪闊洤锕ュ▍鍛存煟韫囥儳绡€闁诡喕绮欓、娑㈠Χ閸モ晝妲囨俊銈囧Х閸嬫盯宕导鏉戠疅闁圭虎鍠栫粈瀣亜閹捐泛校闁搞倕鐭傚缁樼瑹閳ь剟鍩€椤掑倸浠滈柤娲诲灡閺呭墎鈧數纭堕崑鎾斥枔閸喗鐏嶉梺缁樻惈缁绘繈鐛繝鍥ㄥ亹婵炶尙绮弲銏＄箾鏉堝墽鍒板鐟帮攻鐎靛ジ寮撮姀鈾€鎷绘繛杈剧秬濞咃絿鏁☉娆庣箚妞ゆ劧绱曢ˇ锕傛煃鐠囪尙效闁轰焦鍔栧鍕偓锝庝簻婢规帗绻濆▓鍨灍妞ゎ厼鐗撳畷娲冀椤撶喎浜楅梺缁樻煥閹诧繝宕ｈ箛鏃€鍙忔俊銈傚亾婵☆偅顨嗛弲鑸电節濮橆厾鍘遍梺闈涚墕濞层倝寮搁幋鐐簻闁靛绲介崝锕傛煙椤旂晫鎳囨い銏℃瀹曠喖顢樺┑鍫㈩槱闂傚倸鍊风粈渚€骞栭銈嗗仏妞ゆ劧绠戠壕褰掓煛閸モ晛鏋戞繛鍛箻濮婄粯鎷呯憴鍕哗闂佺瀛╃划鎾崇暦濮椻偓閳ワ箓骞嬮悙瀵哥懁闂傚倸鍊峰ù鍥х暦閻㈢绐楅柟閭﹀枛閸ㄦ繈鏌熼悙顒傛菇闁逞屽墮閸婂潡寮崘顔肩＜婵炴垶菤閸嬫捇宕稿Δ浣哄弳闂佸搫鍊搁悘婵嬪煕閺冨倻纾奸柣妯挎珪椤ュ牓鏌″畝瀣М鐎殿噮鍓涢幑鍕Ω閿旇鏁诲┑鐘殿暯閸撴繆銇愰崘顔藉亱濠电姴鍟伴埞宥呪攽閻樺弶鎼愰崶鎾⒑閸涘﹣绶遍柛妯挎閳绘捇顢橀姀鈾€鎷?URL 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｉ幇顒佹儓闁搞劌鍊块幃瑙勬姜閹峰矈鍔呭┑鐐插悑閻楁洟鍩為幋锔藉亹閻庡湱濮撮ˉ婵堢磽娴ｇ懓濮堟い銊ワ躬閻涱噣寮介‖銉ラ叄椤㈡鍩€椤掑嫭鍊舵い鏂款潟娴滄粓鏌嶉崫鍕櫤闁轰浇浜槐鎺旂磼濡吋鍒涘Δ鐘靛仜椤戝懘鍩㈡惔銈囩杸闁哄洦纰嶉崑鍛存⒒閸屾瑧顦﹂柟纰卞亞閳ь剚鍝庨崝鎴濈暦閺囥垺鍋ㄧ紒瀣劵閹芥洟姊洪幐搴ｇ畵妞わ富鍨崇划濠氭晲婢跺鍘介梺褰掑亰閸撴岸藟閻愮儤鐓曢柕濞垮劜閸嬨儵鏌ｅ☉鍗炴灕缂佺姵绋戦埥澶娾枎閹存繂笑闂佽楠哥粻宥夊磿閸楃倣娑樷槈閵忕姴鍋嶉梺褰掓？閻掞箓宕愰崹顐闁绘劘灏欐禒銏ゆ煕閺冣偓濮樸劎妲愰幒鏃傜＜婵☆垵娅ｉ鍌炴倵濞堝灝鏆欓柛銉戝懎濮︽俊鐐€栫敮鎺斺偓姘煎弮閸╂盯骞嬮悩鐢碉紳婵炶揪缍€椤曟牠鎮炴禒瀣厓妞ゆ牗绋掔粈瀣煛瀹€瀣М闁诡喓鍨藉畷锝嗗緞鐎ｎ剙缍冮梺璇插椤旀牠宕抽鈧畷鎴﹀礋椤栵絾鏅ｆ繝闈涘€婚…鍫濐啅濠靛洢浜滈柡宥冨妿閳洟鏌ｉ敐鍥у幋婵﹦绮幏鍛村川婵犲倹娈樻繝鐢靛仩椤曟粓姊介崟顓犵焿鐎广儱顦伴崐鐑芥煕椤垵浜芥繛鑲╁枎閳规垿顢欑粵瀣姼闂佺硶鏅滈悧鐘诲春濞戙垹绠ｉ柣妯兼暩閿涙粓鏌ｆ惔顖滅У闁告鏅☉鐢稿礈娴ｈ櫣锛滈梺缁樏崯鍧椝夌€ｎ剛纾奸弶鍫涘妼缁楁氨鈧灚婢樼€氼喗绂掗敃鍌涘癄濠㈣泛锕﹂鍝勨攽閿涘嫬浜奸柛濠冨灴瀹曞綊宕崟顓熸畷闂佸憡绋戦敃锝囨崲閸℃稒鐓犻柟闂寸劍濞懷呯磼閸撲礁浠遍柟顔筋殜閺佹劖鎯旈垾鎰佹交闂佹眹鍩勯崹浼存煀閿濆钃熼柨婵嗘閸庣喖鏌ㄥ☉妯侯伀闁哄棙鐩铏圭矙閸ф鈧鈧娲﹂崜娑㈠矗閸涘瓨鈷戦梻鍫熷崟閸儱鐤炬繛鎴欏灩缁€?
	if purchaseEnabled {
		if purchaseURL == "" {
			response.BadRequest(c, "Purchase Subscription URL is required when enabled")
			return
		}
		if err := config.ValidateAbsoluteHTTPURL(purchaseURL); err != nil {
			response.BadRequest(c, "Purchase Subscription URL must be an absolute http(s) URL")
			return
		}
	} else if purchaseURL != "" {
		if err := config.ValidateAbsoluteHTTPURL(purchaseURL); err != nil {
			response.BadRequest(c, "Purchase Subscription URL must be an absolute http(s) URL")
			return
		}
	}

	// Frontend URL 婵犵數濮烽弫鍛婃叏閹绢喗鍎夊鑸靛姇缁狙囧箹鐎涙ɑ灏ù婊呭亾娣囧﹪濡堕崟顓炲闂佸憡鐟ョ换姗€寮婚悢铏圭＜闁靛繒濮甸悘宥夋⒑缁嬪灝顒㈡い銊ユ嚇婵℃挳骞掗幋顓熷兊闂佹寧绻傞幊宥嗙珶閺囥垺鈷?
	req.FrontendURL = strings.TrimSpace(req.FrontendURL)
	if req.FrontendURL != "" {
		if err := config.ValidateAbsoluteHTTPURL(req.FrontendURL); err != nil {
			response.BadRequest(c, "Frontend URL must be an absolute http(s) URL")
			return
		}
	}

	// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸ゅ嫰鏌涢锝嗙闁稿被鍔庨幉鎼佸籍閸惊銉╂煕閹般劍娅嗛柛搴ｅ枛閺屾洝绠涚€ｎ亞鍔村┑鐐跺皺椤牓鍩為幋锔藉亹閻犲泧鍐х矗闂備胶绮〃鍛存晝椤忓嫮鏆﹂柟杈剧畱閻撴盯鏌涘☉鍗炴灓闁告ü绮欏娲焻濞戞ê绁┑鐐板尃閸曨剚鐝锋繛瀵稿帶閻°劑鎮￠崘顏嗙＝濞达綀鍋傞幋鐘电焼闁逞屽墮椤啴濡舵惔鈥崇哗濠电偛顦板ú鐔肩嵁閹达箑绀嬫い鎰ㄥ亾婵炲樊浜堕弫鍌溾偓骞垮劙缁€浣圭閻愵兛绻嗛柕鍫濇噹閺嗙喓鐥崜褎娅曢柍褜鍓欑粻宥夊磿闁秴绠犻煫鍥ㄧ☉閸氬綊鏌涢妷顔煎闁绘挻鐩幃姗€鎮欓幓鎺嗘寖闂佸疇妫勯ˇ鐢稿箺閸洖鍐€妞ゎ兘鈧磭绉洪柡浣瑰姍瀹曘劑顢欑捄顭戞敤濠碉紕鍋戦崐褏绱撳璺虹闁告劕妯婂鏍ㄧ箾瀹割喕绨兼い銉ョ墛缁绘盯骞嬮悙鍨櫧濠碘槅鍋嗛崑鐔烘閹惧瓨濯撮柛鎾冲级妤旈梻浣芥〃缁€渚€鈥﹂柨瀣╃箚闁圭虎鍠栫粈鍐┿亜閺傛寧顫嶉柕濞炬櫆閻撳繐顭跨拠鈥崇仩閻庢凹鍙冮幃楣冩倻閼恒儮鎷绘繛杈剧到閹诧繝宕悙鐑樼厱闁哄啯鎹囧顔剧磼閸屾氨孝闁宠鍨归埀顒婄到婢у海绮?
	const (
		maxCustomMenuItems    = 20
		maxMenuItemLabelLen   = 50
		maxMenuItemURLLen     = 2048
		maxMenuItemIconSVGLen = 10 * 1024 // 10KB
		maxMenuItemIDLen      = 32
	)

	customMenuJSON := previousSettings.CustomMenuItems
	if req.CustomMenuItems != nil {
		items := *req.CustomMenuItems
		if len(items) > maxCustomMenuItems {
			response.BadRequest(c, "Too many custom menu items (max 20)")
			return
		}
		for i, item := range items {
			if strings.TrimSpace(item.Label) == "" {
				response.BadRequest(c, "Custom menu item label is required")
				return
			}
			if len(item.Label) > maxMenuItemLabelLen {
				response.BadRequest(c, "Custom menu item label is too long (max 50 characters)")
				return
			}
			urlTrimmed := strings.TrimSpace(item.URL)
			if strings.HasPrefix(urlTrimmed, "md:") {
				// Markdown page mode: URL = "md:<slug>"
				slug := strings.TrimPrefix(urlTrimmed, "md:")
				if slug == "" {
					response.BadRequest(c, "Custom menu item markdown slug cannot be empty (use md:slug format)")
					return
				}
			} else {
				if urlTrimmed == "" {
					response.BadRequest(c, "Custom menu item URL is required (use md:slug for markdown pages)")
					return
				}
				if len(item.URL) > maxMenuItemURLLen {
					response.BadRequest(c, "Custom menu item URL is too long (max 2048 characters)")
					return
				}
				if err := config.ValidateAbsoluteHTTPURL(urlTrimmed); err != nil {
					response.BadRequest(c, "Custom menu item URL must be an absolute http(s) URL or md:<slug>")
					return
				}
			}
			if item.Visibility != "user" && item.Visibility != "admin" {
				response.BadRequest(c, "Custom menu item visibility must be 'user' or 'admin'")
				return
			}
			if len(item.IconSVG) > maxMenuItemIconSVGLen {
				response.BadRequest(c, "Custom menu item icon SVG is too large (max 10KB)")
				return
			}
			// Auto-generate ID if missing
			if strings.TrimSpace(item.ID) == "" {
				id, err := generateMenuItemID()
				if err != nil {
					response.Error(c, http.StatusInternalServerError, "Failed to generate menu item ID")
					return
				}
				items[i].ID = id
			} else if len(item.ID) > maxMenuItemIDLen {
				response.BadRequest(c, "Custom menu item ID is too long (max 32 characters)")
				return
			} else if !menuItemIDPattern.MatchString(item.ID) {
				response.BadRequest(c, "Custom menu item ID contains invalid characters (only a-z, A-Z, 0-9, - and _ are allowed)")
				return
			}
		}
		// ID uniqueness check
		seen := make(map[string]struct{}, len(items))
		for _, item := range items {
			if _, exists := seen[item.ID]; exists {
				response.BadRequest(c, "Duplicate custom menu item ID: "+item.ID)
				return
			}
			seen[item.ID] = struct{}{}
		}
		menuBytes, err := json.Marshal(items)
		if err != nil {
			response.BadRequest(c, "Failed to serialize custom menu items")
			return
		}
		customMenuJSON = string(menuBytes)
	}

	// Validate custom endpoints.
	const (
		maxCustomEndpoints        = 10
		maxEndpointNameLen        = 50
		maxEndpointURLLen         = 2048
		maxEndpointDescriptionLen = 200
	)

	customEndpointsJSON := previousSettings.CustomEndpoints
	if req.CustomEndpoints != nil {
		endpoints := *req.CustomEndpoints
		if len(endpoints) > maxCustomEndpoints {
			response.BadRequest(c, "Too many custom endpoints (max 10)")
			return
		}
		for _, ep := range endpoints {
			if strings.TrimSpace(ep.Name) == "" {
				response.BadRequest(c, "Custom endpoint name is required")
				return
			}
			if len(ep.Name) > maxEndpointNameLen {
				response.BadRequest(c, "Custom endpoint name is too long (max 50 characters)")
				return
			}
			if strings.TrimSpace(ep.Endpoint) == "" {
				response.BadRequest(c, "Custom endpoint URL is required")
				return
			}
			if len(ep.Endpoint) > maxEndpointURLLen {
				response.BadRequest(c, "Custom endpoint URL is too long (max 2048 characters)")
				return
			}
			if err := config.ValidateAbsoluteHTTPURL(strings.TrimSpace(ep.Endpoint)); err != nil {
				response.BadRequest(c, "Custom endpoint URL must be an absolute http(s) URL")
				return
			}
			if len(ep.Description) > maxEndpointDescriptionLen {
				response.BadRequest(c, "Custom endpoint description is too long (max 200 characters)")
				return
			}
		}
		endpointBytes, err := json.Marshal(endpoints)
		if err != nil {
			response.BadRequest(c, "Failed to serialize custom endpoints")
			return
		}
		customEndpointsJSON = string(endpointBytes)
	}

	// Ops metrics collector interval validation (seconds).
	if req.OpsMetricsIntervalSeconds != nil {
		v := *req.OpsMetricsIntervalSeconds
		if v < 60 {
			v = 60
		}
		if v > 3600 {
			v = 3600
		}
		req.OpsMetricsIntervalSeconds = &v
	}
	defaultSubscriptions := make([]service.DefaultSubscriptionSetting, 0, len(req.DefaultSubscriptions))
	for _, sub := range req.DefaultSubscriptions {
		defaultSubscriptions = append(defaultSubscriptions, service.DefaultSubscriptionSetting{
			GroupID:      sub.GroupID,
			ValidityDays: sub.ValidityDays,
		})
	}

	// 婵犵數濮烽弫鍛婃叏閹绢喗鍎夊鑸靛姇缁狙囧箹鐎涙ɑ灏ù婊呭亾娣囧﹪濡堕崟顓炲闂佸憡鐟ョ换姗€寮婚悢铏圭＜闁靛繒濮甸悘宥夋⒑缁嬪灝顒㈡い銊ユ嚇婵℃挳骞掗幋顓熷兊闂佹寧绻傞幊宥嗙珶閺囥垺鈷掑ù锝囩摂閸ゅ啴鏌涢…鎴濈仸闁诡喚鍋ら弫鍐焵椤掑嫷鏁嬮柕澶嗘櫅缁€瀣亜閺嶃劍鐨戦柣銈傚亾濠碉紕鍋戦崐鏍ь啅婵犳艾纾婚柟鍓х帛閻撶喐绻濋棃娑欑カ闁肩増瀵ч幈銊︾節閸愨斂浠㈤梺鍝勮嫰閹虫﹢骞冨▎鎾村殤閻犺桨璀︽导鍐ㄢ攽閻橆偅濯伴悘鐐垫櫕娴犳悂姊洪崫鍕伇闁哥姴閰ｉ幃楣冩晸閻樺啿浜遍梺鍓插亝缁诲啴顢旈敓鐘斥拻濞达絿鐡旈崵鍐煕閵娿儱鑸归摶鐐烘煕閹伴潧鏋涢柡瀣╃窔閺岀喖鎮ч崼鐔哄嚒濠碉紕鍋犲Λ鍕箒闂佹寧绻傞悧婊冾焽閹扮増鐓熼柨鏃囧Г椤ュ牊鎱ㄦ繝鍌ょ吋鐎规洘甯掗埢搴ㄥ箣椤撶啘婊堟⒒娴ｅ憡璐￠柍宄扮墦瀹曟垶绻濋崶褏鐣洪悷婊冪Ч閳ワ箓濡搁埡渚€鍞跺┑鐘绘涧閻楁粌危閼哥數绡€闁汇垽娼ф禒婊勪繆椤愶絿鎳呴柡渚囧櫍瀹曞爼濡歌濞叉悂鎮峰鍛暭閻㈩垱顨婇幃锟犲即閻樺啿鏋戦柟鍏肩暘閸斿﹪鍩€椤掆偓閸熸潙鐣烽妸鈺佺骇闁瑰濯Σ浼存⒒娓氣偓閳ь剛鍋涢懟顖涙櫠鐎涙ɑ鍙忓┑鐘叉噺椤忕姷绱掓潏銊ョ瑨閾伙綁鎮归崶顏勭毢妞わ絽寮剁换婵嬫偨闂堟稐娌梺鐟版啞婵炲﹨妫㈤梺鍦亾閺嬬厧危閸儲鐓熼柕蹇嬪焺閻掗箖鏌ｉ幒鎴犱粵闁靛洤瀚伴獮鎺楀箣濠垫劒鐥梻浣瑰▕閺€閬嶅垂閸ф钃熼柣鏃囨閻瑩鏌熺粙鍨劉鐎规洘濞婂铏光偓鍦У椤ュ銇勯敂鐐毈鐎殿喖顭烽弫鎰板川閸屾稒顥堥柛鈹惧亾濡炪倖甯掗崐鐟扮暦閸欏鍙忔俊鐐额嚙娴滈箖鎮楃憴鍕婵＄偘绮欓獮鍐ㄢ枎閹惧厖绱堕梺鍛婃处閸嬪嫰寮妶鍡曠箚闁绘劦浜滈埀顒佺墵瀹曟繂螖閸涱厾鐛ラ柟鍏肩暘閸ㄥ銆呴弻銉︾厽闁逛即娼ф晶鎵磼閻樿崵鐣洪柡宀€鍠撻埀顒傛暩椤牊绂掕閺岀喖鎼归銏狀潔缂備胶绮惄顖氱暦婵傚憡鍋勯柟鐑樻尭閻忔潙鈹戦敍鍕毈鐎规洜鍠栭、娑橆潩妲屾牕鎮堝┑鐘垫暩婵挳鏁冮妶澶婄疇閹兼惌婢€缁诲棝鏌ｉ敐鍛伇缁?缂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亝鎹ｉ柣顓炴閵嗘帒顫濋敐鍛闁诲氦顫夊ú姗€宕濆▎鎾崇畺婵犲﹤鐗婇崵宥夋煏婢跺牆鍔滈柣锝変憾濮婄粯鎷呯粙娆炬闂佺顑呴幊搴ｅ弲闂佸搫绋侀崢濂稿礄閻樺眰鈧帒顫濋敐鍛闂備礁鎼張顒€煤閻旈鏆﹂柣鎾崇岸閺€浠嬫煙闁箑甯ㄧ憸鏂款潖濞差亜浼犻柛鏇ㄥ墮椤庢盯姊洪崨濠冨暗闁哥姵鐗犻悰顕€宕橀…鎴炲缓闂侀€炲苯澧存鐐插暙閳诲酣骞樺畷鍥崜闂備胶鎳撻悺銊ф崲閸屾繍鏆梻鍌氬€搁崐宄懊归崶顒夋晪鐟滃繘骞戦姀銈呯闁挎棁銆€閸嬫捇宕橀埡鍐槇闂佹悶鍎滈崱妯诲暫濠电姷鏁搁崑鐐哄垂閸洘鏅濋柍杞扮贰閻掍粙鏌嶉崫鍕偓鑸电濠婂牊鐓欓柣鎴灻悘銉╂煃瑜滈崜姘跺箖閸屾繄鐔呴梻渚€鈧偛鑻晶鎵磼?semver闂?
	if req.MinClaudeCodeVersion != "" {
		if !semverPattern.MatchString(req.MinClaudeCodeVersion) {
			response.Error(c, http.StatusBadRequest, "min_claude_code_version must be empty or a valid semver (e.g. 2.1.63)")
			return
		}
	}

	// 婵犵數濮烽弫鍛婃叏閹绢喗鍎夊鑸靛姇缁狙囧箹鐎涙ɑ灏ù婊呭亾娣囧﹪濡堕崟顓炲闂佸憡鐟ョ换姗€寮婚悢铏圭＜闁靛繒濮甸悘宥夋⒑缁嬪灝顒㈡い銊ユ嚇婵℃挳骞掗幋顓熷兊闂佹寧绻傞幊宥嗙珶閺囥垺鈷掑ù锝囩摂閸ゅ啴鏌涢…鎴濈仸闁诡喚鍋ら弫鍐焵椤掑嫷鏁嬮柕澶嗘櫅缁€瀣亜閺嶃劍鐨戦柣銈傚亾濠碉紕鍋戦崐鏍ь啅婵犳艾纾婚柟鍓х帛閻撶喐绻濋棃娑欑カ闁肩増瀵ч幈銊︾節閸愨斂浠㈤梺鍝勮嫰閹虫﹢骞冨▎鎾村殤閻犺桨璀︽导鍐ㄢ攽閻橆偅濯伴悘鐐垫櫕娴狀參姊洪棃娑欐悙閻庢氨澧楁穱濠囧箹娴ｈ倽銊╂煥閺冨洤袚闁稿﹦鍏樺娲矗閵壯傝缂傚倸绉崇欢姘嚕婵犳碍鏅插璺猴功椤旀帡姊洪崨濠傚閻忓繑鐟ュ嵄濞寸厧鐡ㄩ埛鎴︽煕濠靛棗顏柨娑樼Ч閺屾稑鈻庤箛鎾存濡炪値鍋勭换鎰弲濡炪倕绻愮€氼亞妲愰崼鏇熲拺闁告稑锕ユ径鍕煕濡亽鍋㈤柛鈺傜洴瀹曞ジ濡烽敂瑙勫闂備礁鎲＄粙鎴︽晝閵壯傜剨濞寸厧鐡ㄩ悡鐔兼煟閺冣偓濞兼瑦鎱ㄩ崒姘ｆ斀闁挎稑瀚弳顒侇殽閻愬弶鍠樼€殿喖鐖煎畷绋课旀担瑙勭暥婵犵數濮撮惀澶愬级鎼存挸浜鹃柟鐗堟緲绾惧鏌熼崜褏甯涢柣鎾存礋閺岀喖鎳滈鈧俊鐓幟瑰鍫㈢暫婵﹤顭峰畷鎺戔枎閹烘垵甯紓鍌欑贰閸ｎ噣宕归崼鏇犲祦闁归偊鍠氶惌娆愪繆椤愶紕绁锋い顐㈩樀钘濋柣妤€鐗婇崕鐔兼煏韫囨洖孝妤犵偞鍨垮缁樼瑹閳ь剟鍩€椤掑倸浠滈柤娲诲灡閺呭墎鈧稒蓱閸欏繐鈹戦悩鎻掓殲闁靛洦绻冮〃銉╂倷閺夋垵顫嶉梺璇″灡濡啴寮幇鏉跨＜婵炴垼椴哥紞宀勬⒒閸屾艾鈧兘鎳楅崜浣稿灊妞ゆ牜鍋戦埀顒€鍟村畷銊╊敍濞戞ê绨ユ繝娈垮枟閵囨盯宕戦幘鍨涘亾鐟欏嫭绀冩俊鐐扮矙楠炲啫鈻庨幘鍏呯炊闂佸憡娲﹂崑鍕极閵堝棔绻嗛柣鎰典簻閳ь剚鐗犲畷婵單旈崨顓犵崶闁瑰吋鐣崹濠氥€呴弻銉︾厽闁逛即娼ф晶鎵磼閻樿崵鐣洪柡宀€鍠撻埀顒傛暩椤牊绂掕閺岀喖鎼归銏狀潔缂備胶绮惄顖氱暦婵傚憡鍋勯柟鐑樻尭閻忔潙鈹戦敍鍕毈鐎规洜鍠栭、娑橆潩妲屾牕鎮堝┑鐘垫暩婵挳鏁冮妶澶婄疇閹兼惌婢€缁诲棝鏌ｉ敐鍛伇缁?缂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亝鎹ｉ柣顓炴閵嗘帒顫濋敐鍛闁诲氦顫夊ú姗€宕濆▎鎾崇畺婵犲﹤鐗婇崵宥夋煏婢跺牆鍔滈柣锝変憾濮婄粯鎷呯粙娆炬闂佺顑呴幊搴ｅ弲闂佸搫绋侀崢濂稿礄閻樺眰鈧帒顫濋敐鍛闂備礁鎼張顒€煤閻旈鏆﹂柣鎾崇岸閺€浠嬫煙闁箑甯ㄧ憸鏂款潖濞差亜浼犻柛鏇ㄥ墮椤庢盯姊洪崨濠冨暗闁哥姵鐗犻悰顕€宕橀…鎴炲缓闂侀€炲苯澧存鐐插暙閳诲酣骞樺畷鍥崜闂備胶鎳撻悺銊ф崲閸屾繍鏆梻鍌氬€搁崐宄懊归崶顒夋晪鐟滃繘骞戦姀銈呯闁挎棁銆€閸嬫捇宕橀埡鍐槇闂佹悶鍎滈崱妯诲暫濠电姷鏁搁崑鐐哄垂閸洘鏅濋柍杞扮贰閻掍粙鏌嶉崫鍕偓鑸电濠婂牊鐓欓柣鎴灻悘銉╂煃瑜滈崜姘跺箖閸屾繄鐔呴梻渚€鈧偛鑻晶鎵磼?semver闂?
	if req.MaxClaudeCodeVersion != "" {
		if !semverPattern.MatchString(req.MaxClaudeCodeVersion) {
			response.Error(c, http.StatusBadRequest, "max_claude_code_version must be empty or a valid semver (e.g. 3.0.0)")
			return
		}
	}
	if req.AntigravityUserAgentVersion != nil {
		normalized := strings.TrimSpace(*req.AntigravityUserAgentVersion)
		req.AntigravityUserAgentVersion = &normalized
		if normalized != "" && !semverPattern.MatchString(normalized) {
			response.Error(c, http.StatusBadRequest, "antigravity_user_agent_version must be empty or a valid semver (e.g. 1.23.2)")
			return
		}
	}
	if req.OpenAICodexUserAgent != nil {
		normalized := strings.TrimSpace(*req.OpenAICodexUserAgent)
		req.OpenAICodexUserAgent = &normalized
		// 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾剧粯绻涢幋娆忕労闁轰礁顑嗛妵鍕箻鐠虹儤鐎鹃梺鍛婄懃缁绘﹢寮婚悢铏圭＜闁靛繒濮甸悘鍫㈢磽娴ｅ搫啸濠电偐鍋撳Δ鐘靛仦閻楁洝褰佸銈嗗坊閸嬫捇鏌ｈ箛锝呮灍闁靛洤瀚伴崺锟犲川椤旇姤娈哥紓鍌欒兌缁垶鎯勯鐐靛祦閻庯綆浜栭弸搴ㄧ叓閸ャ劍纾婚柣銉邯濮婄粯鎷呴悷鎵虫灆闂佽　鍋撻梺顒€绉寸粈澶嬬箾閸℃ɑ灏柦鍐枑缁绘盯骞嬮弮鈧崳浼存煃瑜滈崜銊х不閹捐崵宓侀悗锝庝簴閺€浠嬫煕閵夈垺娅冮梺顓у灠閳规垿鏁嶉崟顐℃澀闂佺锕ラ悧鐘差嚕婵犳艾惟闁冲搫鍊瑰▍鍥⒑闂堟稓澧曟い锔诲灦閹繝鍨鹃幇浣哄數闂佸吋鎮傚褎鎱ㄩ崶鈺冪＜闁绘ê鎼崥鍦磼鏉堛劌绗掗柍钘夘槸椤繈宕楅崗鍏兼闂傚倷绀侀幉锟犲礄瑜版帒纾婚柛鏇ㄥ灠閻撴﹢鏌熸潏鍓х暠闁诲繑濞婇弻娑㈠箛椤撶姰鍋為梺鍦焾閸熸潙顫忛搹鍦煓閻犳亽鍔嶅Σ鈧梻浣侯焾閿曘儵銆冮崼銉ョ闁圭儤鍩堝鈺呮偣妤︽寧顏犳い銏犳嚇濮婅櫣绱掑Ο鍝勵潊闂佸搫鎳忕划鎾愁嚕椤掑嫬鐒垫い鎺戝閳锋帒霉閿濆牜娼愰柛瀣█閺屾稒鎯旈姀掳浠㈤梺璇″枟閿氶柍瑙勫灩閳ь剨缍嗘禍鐐侯敇濞差亝鈷戦柟绋垮閳锋帡鏌￠崪浣镐喊鐎规洩缍佸畷鍗烆渻缂佹ɑ鏉搁梻浣虹帛閸旀牕顭囧▎鎾村€堕柨鏇炲€归悡銉︽叏濮楀棗骞戝ù婊勭矒閺屾洘绻濊箛鎿冩喘闂佺懓鍟块幊鎰閹烘挻缍囬柕濞垮劤閻熴劑鏌х紒妯煎⒌闁哄苯绉烽¨渚€鏌涢幘璺烘灆婵☆偆鍠栧濠氬磼濮樺崬顤€婵炴挻纰嶉〃鍛粹€﹂崶顒€鍐€妞ゎ兘鈧磭绉洪柟顔规櫅椤斿繘顢欓悡搴☆棈缂傚倸鍊烽懗鑸垫叏閻㈢數鐭欓柟鐑橆殔妗呴梺鍛婃处閸ㄦ壆绮婚幎鑺ョ厱闁斥晛鍟ㄦ禒锕€顭跨憴鍕婵﹦绮幏鍛存惞閻熸壆顐奸梻浣告贡椤牓鈥﹀畡鎵殾鐟滅増甯掗柨銈嗕繆閵堝倸浜鹃梺钘夊暟閸犳劗鎹㈠☉銏犵闁绘劗鏁搁悡鍌炴⒑濮瑰洤鈧劙宕戦幘瀵哥瘈婵炲牆鐏濋弸鐔兼煥閺囨娅婄€规洩缍侀崺鈧い鎺戝閻撴洘鎱ㄥ鍡楀闁逞屽墲瀹曢潧危閹版澘绠婚悹鍥皺閿涙粌鈹戦缂存垿鎯侀悜钘夘潊閹鸿櫕绂嶅鍫熺厵閻庢稒顭囩粻銉︿繆閹绘帗鍟為柕鍥у瀵挳顢旈崱娅虫垹绱撴担绋库偓鍦暜閻愬搫鐒垫い鎺戯功缁夌敻鏌涚€ｎ亜顏紒鍌氱Т铻栭柛娑卞枤閸樼敻姊洪悡搴㈣础濠⒀勵殙閵囨劘顦寸紒杈ㄥ浮楠炴捇骞掗幘鍏呮樊闂備線娼уΛ鏃傛濮橆剛鏆﹂柕濞р偓閸嬫挸鈽夊▍杈ㄥ哺楠炲繐煤椤忓應鎷洪梺闈╁瘜閸欏酣鎮為悙顒傜閻庣數顭堥崢鎾煕閳瑰灝鍔滅€垫澘瀚伴獮鍥敇閻樻彃绠哄┑鐘茬棄閺夊簱鍋撻幘鍓佺焼濞达綀顫夐崣蹇擃熆閼搁潧濮堥柣鎾寸懄椤ㄣ儵鎮欓懠顑胯檸闂佽绻嗛弲婊堝Φ閸曨垰绠婚柣鎰娴狀噣鎮楃憴鍕闁绘牕銈稿畷娲晸閻樿尙顦ㄩ柣鐘充航閸斿瞼鑺遍妷鈺傗拻闁稿本鑹鹃埀顒勵棑缁牊鎷呴棃鈺勨偓鍧楁⒑椤掆偓缁夌敻骞嗛悙鍝勭婵烇綆鍓欐俊鑲╃磼閹邦収娈曞ǎ鍥э躬婵″爼宕熼鐓庡腐婵＄偑鍊愰弲鐘诲绩闁秴绠為柕濞垮労濞撳鎮归崶顏勭处濠㈣娲熷缁樻媴缁涘娈愰梺鍝ュУ瀹€绋跨暦濡も偓铻ｅ〒姘煎灠濞堛劑姊洪崫鍕垫Ч闁糕晛鐗嗛…鍥煛娴ｅ嫸绲挎禒锕傚礈瑜嶉～褔鎮跺鍓х暤婵﹥妞介幊锟犲Χ閸涱喚鈧墽绱撴笟鍥ф灈閻庢稈鏅犲鑼崉娴ｆ洘妫冨畷銊╊敊闂傚鏆楅梻鍌欐祰濞夋洟宕伴幘瀛樺弿閻庨潧鎽滄稉宥夋煟濡櫣锛嶇紒鐘荤畺閺岀喓鈧稒顭囬幊鍐煟閹烘挻銇濋柡灞稿墲閹峰懐绮欐惔鎾充壕闁秆勵殔閽冪喐绻涢幋娆忕仼闁绘帗妞介弻娑滅疀婵犲啯鐝曢梺鎼炲妿閺佸宕洪姀銈呯閻犲洩灏欓鎰箾鏉堝墽鍒版繝鈧柆宥呭偍濞寸姴顑嗛埛鎴犵磽娴ｅ厜妫ㄦい蹇撳珋閻斿摜绡€闁搞儜鍜冪吹闂傚鍋勫ú锔剧矙閹烘鍋傞柡鍥ュ灪閻撶喖鏌￠崶銊︾殤妞わ讣绠撻弻锝堢疀閵夈儺娲梺?codex 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸ゅ嫰鏌涢幘鑼槮闁搞劍绻冮妵鍕冀椤愵澀鏉梺閫炲苯澧柛鐔告綑閻ｇ兘濡歌閸嬫挸鈽夊▍顓т簼缁傚秵娼忛妸褏鐦堥梺姹囧灲濞佳冪摥闂備胶顭堥鍛存晝椤忓牆绠犻柟鎹愵嚙閸ㄥ倹銇勯弮鈧懝鍓х礊婵犲倻鏆︽俊銈呮噹缁€鍌炴煠濞村娅囬柍璇茬箻濮婄粯鎷呯憴鍕╀户闂佸憡眉缁瑩濡撮崘鈺冪瘈闁搞儜鍡樻啺婵犵數鍋為崹顖炲垂閸︻厾涓嶉柨婵嗘缁犻箖鏌涢埄鍐炬畼缂佺姵宀搁弻娑氣偓锝庡亝鐏忣參鎽堕弽顓熺厱闁规澘鍚€缁ㄤ粙鏌?
		if len(normalized) > 512 {
			response.Error(c, http.StatusBadRequest, "openai_codex_user_agent must be at most 512 characters")
			return
		}
	}

	// 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾剧粯绻涢幋娆忕仾闁稿鍊濋弻鏇熺箾閻愵剚鐝旈梺鍛婅壘缂嶅﹪鐛弽顬ュ酣顢楅埀顒佷繆閼测晝妫柟顖嗗懐楔闂佸搫鐭夌换婵嗙暦閸洖鐓涘ù锝囶焾椤ュ酣姊绘担瑙勫仩闁稿﹥鐗曠叅闁绘梻鍘ч拑鐔兼煏閸繍妲哥€瑰憡绻傞埞鎴︽偐閺屻儺鈧鏌￠崨顓炶埞闁宠鍨块幃娆撴嚑椤掆偓閳亶姊洪崫鍕闁稿锕畷娲焵椤掍降浜滈柟鍝勬娴滈箖姊洪崨濞氭垹鍒掗幘宕囨殾闁硅揪绠戠粻鑽ょ磽娴ｈ鐒介柛姗€浜跺铏规喆閸曨剛鍑￠梺鍛婂焹閸嬫挾绱撴担闈涘闁告艾顑夋俊鐢稿礋椤栵絾鏅濋梺闈涚箚閹筹綀顦抽柟渚垮妽缁绘繈宕橀埞澶歌檸婵°倗濮烽崑鐐哄礉閺嶎厼绠氶柡鍐ㄧ墛閺呮煡鏌涚仦鎯у毈婵炲吋鎹囧濠氬磼濞嗘帒鍘￠梺鍛婎焼閸涱垳顦繝鐢靛Т鐎氼喗绋夊澶嬬厸鐎广儱楠搁獮鏍ㄦ交濠靛鈷戦柛娑橈功濞叉挳鏌涚€ｎ亷宸ラ摶锝夋煕閺囥劌鐏￠柛濠傜仛缁绘盯宕煎┑鍫滆檸闂佸搫顑冮崐鏍Φ閸曨垼鏁囬柣鎰版涧閳敻姊洪悙钘夊姷缂佺姵鎸绘穱濠囧箹娴ｅ摜鍘告繛杈剧导鐠€锕€危娴煎瓨鈷掑ù锝呮憸閺嬪啯淇婂鐓庡闁瑰嘲顑夊濠氬Ψ閵壯嶇幢闂備胶鎳撴晶鐣屽垝椤栫偞鍊块柤娴嬫杹閸嬫捇鐛崹顔煎濠碘槅鍋呴惄顖氱暦閵忋倕绠瑰ù锝呭帨閹锋椽姊洪崨濠勨槈闁挎洏鍊栭幈銊╁焵椤掑嫭鈷戦柛婵嗗閸ｆ椽鏌ｉ埡濠傜仩闁伙絽鍢查埞鎴犫偓锝庡亜濞堟繈姊虹紒妯哄闁逞屽墯缁嬫捇鎯勬惔銊︹拻濞达絽鎲￠幆鍫ユ煕閻斿搫鈻堢€规洘鍨块獮妯尖偓闈涙憸椤旀洟姊洪崨濠勬噧妞わ箓浜堕幆宀€浠﹂崣銉х畾闂佸憡鐟ラˇ顖涙叏瀹ュ洠鍋撳▓鍨珮闁革綇绲介悾閿嬬附閸撳弶鏅濋梺鎸庣箓濡瑩鎮甸敃鍌涒拻闁稿本鐟ㄩ崗宀勬煕濡姴娴勭紞鏍煛閸ャ儱鐏╃痪鎯х秺閺岋繝宕堕妷銉т患闂佹眹鍊濈粻鏍蓟瀹ュ浼犻柛鏇ㄥ墮濞呫倝姊洪幎鑺ユ暠闁搞劌娼″璇测槈濡攱鏂€闂佺硶鍓濋〃蹇斿閳ь剙鈹戦悩顐ｅ闁告洦鍋勯～鎺楁⒑閸濆嫭婀扮紒瀣崌閸┾偓妞ゆ帒锕︾粔鐢告煕鐎ｎ亝鍣藉ù婊勬倐椤㈡﹢濮€閳藉棙鈷栧┑鐘灱閸╂牠宕濋弽顓熷亗闁靛濡囩弧鈧梻鍌氱墛缁嬫帡藟閻愮儤鐓曢柕鍫濇噺婢跺嫰鏌曢崶褍顏柟顖涙婵偓闁冲搫鍊婚弶褰掓⒒娴ｅ憡鎯堥柣顒€銈稿畷浼村冀椤撶偟顔戦柟鍏肩暘閸斿瞼澹曢崗鍏煎弿婵妫楁晶浼存煏閸偄娅嶉柡宀嬬稻閹棃濮€閳轰焦娅涢梻浣告憸婵潧鐣濈粙娆惧殨妞ゆ劧绲跨弧鈧┑顔斤供閸撴艾危椤掑嫭鈷戦梺顐ゅ仜閼活垱鏅堕鐐寸厪闁搞儜鍐句純濡ょ姷鍋炵敮鎺楊敇婵傜鐐婃い蹇撴瀹曡泛鈹戦悩鍨毄闁稿鐩幃褎绻濋崶銉㈠亾閸愵喖骞㈡俊顖氱毞閺€铏節閻㈤潧孝婵炴潙鍊垮顐㈩吋閸℃瑧鐦堟繝鐢靛Т閸婃悂寮冲▎寰濆綊鎮╃搾渚囦邯閳ユ棃宕橀鍢壯囩叓閸ャ劍绀€闁糕晛鐭傚娲濞戞艾顤€闂佽鍠栭崐鎼侇敋閿濆棛绡€婵﹩鍘藉▍銏ゆ⒑缂佹〞鎴︽偂閿熺姴鑸归悹杞拌濞撳鏌曢崼婵囶棞濠殿喖鍊块弻娑㈠Ω閵壯呅ㄩ梺鎸庣箘閸嬨倝宕洪敓鐘插窛妞ゆ棃妫块悽缁樼節閻㈤潧孝闁挎洏鍊楅埀顒佸嚬閸樻儳鈻庨姀銈呯煑濠㈣泛鐬奸鏇犵磼缂併垹寮い銉︽尵瀵囧焵椤掑嫭鍊垫繛鍫濈仢濞呮﹢鏌涚€ｎ亜顏い銏∩戠缓鐣岀矙鐠恒劌鈧偤姊洪幐搴㈩梿婵☆偄瀚崚濠冨鐎涙ǚ鎷绘繛杈剧秬濞咃絿鏁☉銏＄厱闁哄啯鎸鹃悾鐢碘偓瑙勬礃椤ㄥ懘鍩ユ径濠庢僵妞ゆ巻鍋撻柣蹇擄工椤啴濡堕崱娆忣潷缂備礁顑嗛崝妤呭箲閵忋倕骞㈡繛鎴炵懅閸橆亪姊洪崜鎻掍簼缂佽鍟村畷瀹犮亹閹烘挾鍘撻悷婊勭矒瀹曟粌鈻庨幋鐘殿槸婵犵數濮村ú銈夋倶閹惰姤鐓ラ柡鍥╁仜閳ь剚鎮傞幃娆愮節閸ャ劎鍘繝鐢靛Т缁绘ê顬婇鈧弻锝呪攽閸℃些闂佸疇顫夐崹鍫曠嵁婵犲洦鐓曞┑鐘插暞缁€瀣殽閻愯鏀婚柟顖涙閺佹劙宕掑顐㈩棜婵犳鍠楅敃鈺呭礈閿曗偓鍗辩憸鐗堝笚閻撳繘鏌涢妷鎴濆枤娴煎啫螖閻橀潧浠﹂悽顖椻偓宕囨殾闁告鍎愬〒濠氭煙椤栧棗鏈ˉ锟犳⒒閸屾瑦绁扮€规洜鏁诲畷浼村箛椤旂厧鐏婇悗骞垮劚濡宕归弬妫靛綊鎮╁顔煎壈缂佺偓鍎冲锟犲蓟閵堝棙鍙忛柟閭﹀厴閸嬫挸螖娴ｈ　鏋栭悗骞垮劚濡稓绮绘ィ鍐╃厵閻庢稒顭囬幊鍐煟韫囧﹥娅囬柍褜鍓涢幊鎾诲箟閿涘嫭宕查柛鎰典簼瀹曞弶绻涢幋娆忕仾闁哄懏鐓￠獮鏍垝閸忓浜鹃柛鎰皺椤撴椽姊?>= 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏℃櫆闁芥ê顦純鏇㈡⒒娴ｅ憡鍟為柛鏃撶畵瀹曚即寮介銏╂婵犵數濮电喊宥夋偂濞嗘挻鐓欓悷娆忓婵鎮敃鍌涘仭婵犲﹤鎳庨。濂告偨椤栨侗娈滅€殿喗鐓″濠氬Ψ閿旀儳骞堟繝鐢靛仦閸ㄩ潧鐣烽鍕嚑闁绘棁銆€閸嬫挾鎲撮崟顒傤槷濠电偛鎳忓ú婊堝箲閵忕姭鏀介悗锝庝簽閸橀亶姊洪柅鐐茶嫰婢у鈧娲滈崰鏍х暦瑜版帩鏁冮柕鍫濇缁额剟姊绘担绛嬪殭閻庢稈鏅犻、娆撳冀椤撶偟鐛ュ┑顔筋焾濞夋盯寮告担绯曟斀闁绘ê鍟块幖鎼佹煃瑜滈崗娑氱矆娴ｇ晫浜欓梻浣告啞娓氭宕㈡ィ鍐ㄧ闁挎洍鍋撴い顏勫暣婵″爼宕卞Δ鍐噯闂備礁鎽滈崰鍡涘礉瀹ュ洨鐭?
	if req.MinClaudeCodeVersion != "" && req.MaxClaudeCodeVersion != "" {
		if service.CompareVersions(req.MaxClaudeCodeVersion, req.MinClaudeCodeVersion) < 0 {
			response.Error(c, http.StatusBadRequest, "max_claude_code_version must be greater than or equal to min_claude_code_version")
			return
		}
	}

	// cyber 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｉ幇顒佲枙闁绘帟濮ょ换娑㈠幢濡纰嶉柣鐔哥懕缁犳捇骞冨Δ鍛櫜閹肩补鈧尙鏁栭梻浣哥－閹虫挾绮旈悽鍨床婵炴垶锕╅崯鍛亜閺冨洤鍚归柛鎴濈秺濮婅櫣绮欓崸妤娾偓妤呮煥閺囥劋閭€殿喖顭烽弫鎰緞鐎ｎ偅鐝抽梻浣告啞娓氭宕ｆ惔鈾€鏋旀い鎾跺亹閺€浠嬫煟閹邦垰鐨洪柣鎺撳劤闇夋繝濠傜凹闁垶鏌熼鍡欑瘈妞ゃ垺娲熼弫鍐焵椤掑嫭鍊峰┑鐘叉处閻撳繐鈹戦悙鎴濆暞濠€鐗堢箾閸繄绉烘慨?TTL 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犖ч柛銉㈡櫇閸橆垶姊绘担鍛婂暈婵炶绠撳畷銏ゆ寠婢跺本娈鹃梺鍛婄懃椤﹁京寮ч埀顒勬⒒閸屾氨澧涘〒姘殜瀹曟洟骞囬悧鍫㈠幈闁诲函缍嗛崐鏍箣閻樺啿搴婂┑鐐村灟閸ㄥ綊鐛姀鈥茬箚妞ゆ牗绻嶉崵娆撴⒒婢跺﹦效婵﹥妞藉畷顐﹀礋椤掆偓椤庢盯姊洪崨濠冨鞍鐎光偓閹间胶宓侀柛顐犲劚鎯熼梺闈涱樈閸犳顕欏ú顏呪拺缂侇垱娲栨晶鍙夈亜閵娿儲鍣界紒顔碱煼瀹曠兘顢橀悩纰夌闯濠电偠鎻紞鈧俊顐㈠閹啫煤椤忓懐鍘甸梺鎯ф禋閸嬪懎鐣风仦瑙ｆ斀闁斥晛鍟亸锔筋殽閻愭潙娴鐐差儔椤㈡﹢鎮㈠畡鏉垮釜婵犵绱曢崑鎴﹀磹閺嶎厼绠伴柣鎰靛墯閸欏繑銇勯幘璺烘瀾闁告瑥绻橀弻娑㈠Ψ閹存柨浜鹃梺琛″亾濞寸姴顑嗛悡銉︾箾閹寸伝顏堫敂閳轰急褰掓偐瀹曞洨鐓夐梺鍝勬湰缁嬫帞鎹㈠┑瀣妞ゆ帊鑳惰ぐ褏绱撻崒娆戣窗闁哥姵鐩、姘愁樄闁糕斂鍎插鍕箛椤掑缍傞梻浣虹帛椤牏浜稿▎鎾村仒闁冲搫鍊荤弧鈧梺姹囧灲濞佳囧礈瀹曞洨纾界€广儱鎳忛ˉ銏⑩偓瑙勬礃椤ㄥ懘锝炲鍫濈劦妞ゆ巻鍋撻柣?> 0
	if req.CyberSessionBlockTTLSeconds != nil && *req.CyberSessionBlockTTLSeconds <= 0 {
		response.BadRequest(c, "cyber_session_block_ttl_seconds must be > 0")
		return
	}

	settings := &service.SystemSettings{
		// 缂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾剧懓顪冪€ｎ亜顒㈡い鎰矙閺屻劑鎮㈤崫鍕戙垽鏌涢妸銉モ偓鍨潖婵犳艾閱囬柣鏃€浜介埀顒佸浮閺岋繝宕遍鐘垫殼闂佸搫鐭夌紞浣规叏閳ь剟鏌曢崼婵囶棡闂佹鍙冨娲箹閻愭彃顬夌紓浣筋嚙閸熶即骞戦姀鐘婵妫楅弲鐘差渻閵堝棙顥嗙€规洜鏁婚幆鍕償閿濆洨锛濇繛杈剧稻瑜板啯绂嶆ィ鍐┾拺缂佸娉曢悘閬嶆煕鐎ｎ剙浠遍柟顔欍倗鐤€闁圭虎鍨遍弬鈧梻浣虹帛閸旀浜稿▎鎰浄闁靛繈鍊栭悡鏇㈡煛閸屾繍鍤欓柍褜鍓氶悧婊堝箲閵忕姭鏀介悗锝庝簽閸?platform quota 婵犵數濮烽弫鍛婃叏閻㈠壊鏁婇柡宥庡幖缁愭淇婇妶鍛殲闁哄棙绮嶆穱濠囧Χ閸涱厽娈堕梺娲诲幗閻熲晠寮婚悢鍏煎€绘俊顖炴櫜缁爼姊洪柅鐐茶嫰婢у墽绱掗悩铏碍闁伙綁鏀辩缓鐣岀矙鐠囦勘鍔嶉妵鍕籍閸ヮ灝鎾寸箾閸涱厾孝闁宠鍨块幃鈺呭垂椤愶絾鐦庡┑鐘愁問閸犳岸寮繝姘疇婵犻潧顑呯粈鍫㈡喐瀹ュ＆澶婎潩閼哥數鍘遍梺鎸庢椤曆囩嵁濡绻嗘い鎰靛亜閻忥繝鏌曢崶褍顏鐐达耿瀹曪繝鎮欓棃娑樞ゅ┑鐘殿暯濡插懘宕戦崨鏉戝瀭闁革富鍘介～鏇㈡煙閻戞﹩娈旈柣鎺戠仛閵囧嫰骞掗幋婵愪患缂備讲鍋撻悗锝庡亖娴滄粓鏌″鍐ㄥ闁靛棙甯￠弻娑樜熼悜妯烘殘缂備胶绮粙鎺戭焽韫囨稑绀堢憸蹇氣叢濠电姷鏁搁崑娑㈠箯閹寸姴绶ら柛顭戝櫘閸ゆ洜鎲搁弮鍫濇槬闁逞屽墯閵囧嫰骞掗幋婵冨亾閼姐倕顥氬ù鐘差儐閻撴洟鎮橀悙闈涗壕闁汇劍鍨归埀顒佺⊕缁诲倿鈥旈崘顔嘉ч柛鈩冾殘娴犳潙顪冮妶蹇撶槣闁搞劌鐖奸妴浣割潩閼稿灚娅滄繝銏ｆ硾閿曪箓顢欓幒鎴富闁靛牆妫欓ˉ鍡涙煕鐎ｎ偄濮嶆い銏＄懇瀹曞爼顢楁担鍝勫箥闂備礁鍚嬬粊鎾疾閻愯埖锛傞梻鍌欑窔閳ь剛鍋涢懟顖涙櫠椤斿墽妫紓浣靛灩瀵喚鈧娲﹂崑鍛村箚閺冨牆惟闁挎棁顫夌€氬ジ姊绘担鍝勫付妞ゎ偅娲熷畷鎰板即閻愬灚鎳冩繝鐢靛Х閺佹悂宕戦悙宸劷婵炲棙鎸哥壕褰掓煕椤垵鏋ら柡鍡畵閺屾洝绠涚€ｎ亖鍋撻弽顓炵畾闁绘劗鍎ら悡鏇㈡煥閺冨浂鍤欐鐐村笧缁辨帒螖閳ь剟藝闂堟侗娼栭柧蹇撴贡绾惧吋淇婇婵嗗惞闁告瑥妫濆娲川婵犲懎顥濆銈嗗灥椤︻垶顢?		DefaultPlatformQuotas: req.DefaultPlatformQuotas,

		RegistrationEnabled:              req.RegistrationEnabled,
		EmailVerifyEnabled:               req.EmailVerifyEnabled,
		RegistrationEmailSuffixWhitelist: req.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                 req.PromoCodeEnabled,
		PasswordResetEnabled:             req.PasswordResetEnabled,
		FrontendURL:                      req.FrontendURL,
		InvitationCodeEnabled:            req.InvitationCodeEnabled,
		TotpEnabled:                      req.TotpEnabled,
		LoginAgreementEnabled:            req.LoginAgreementEnabled,
		LoginAgreementMode:               loginAgreementMode,
		LoginAgreementUpdatedAt:          loginAgreementUpdatedAt,
		LoginAgreementDocuments:          loginAgreementDocuments,
		SMTPHost:                         req.SMTPHost,
		SMTPPort:                         req.SMTPPort,
		SMTPUsername:                     req.SMTPUsername,
		SMTPPassword:                     req.SMTPPassword,
		SMTPFrom:                         req.SMTPFrom,
		SMTPFromName:                     req.SMTPFromName,
		SMTPUseTLS:                       req.SMTPUseTLS,
		TurnstileEnabled:                 req.TurnstileEnabled,
		TurnstileSiteKey:                 req.TurnstileSiteKey,
		TurnstileSecretKey:               req.TurnstileSecretKey,
		APIKeyACLTrustForwardedIP: func() bool {
			if req.APIKeyACLTrustForwardedIP != nil {
				return *req.APIKeyACLTrustForwardedIP
			}
			return previousSettings.APIKeyACLTrustForwardedIP
		}(),
		LinuxDoConnectEnabled:                  req.LinuxDoConnectEnabled,
		LinuxDoConnectClientID:                 req.LinuxDoConnectClientID,
		LinuxDoConnectClientSecret:             req.LinuxDoConnectClientSecret,
		LinuxDoConnectRedirectURL:              req.LinuxDoConnectRedirectURL,
		DingTalkConnectEnabled:                 req.DingTalkConnectEnabled,
		DingTalkConnectClientID:                req.DingTalkConnectClientID,
		DingTalkConnectClientSecret:            req.DingTalkConnectClientSecret,
		DingTalkConnectRedirectURL:             req.DingTalkConnectRedirectURL,
		DingTalkConnectCorpRestrictionPolicy:   req.DingTalkConnectCorpRestrictionPolicy,
		DingTalkConnectInternalCorpID:          req.DingTalkConnectInternalCorpID,
		DingTalkConnectBypassRegistration:      req.DingTalkConnectBypassRegistration,
		DingTalkConnectSyncCorpEmail:           req.DingTalkConnectSyncCorpEmail,
		DingTalkConnectSyncDisplayName:         req.DingTalkConnectSyncDisplayName,
		DingTalkConnectSyncDept:                req.DingTalkConnectSyncDept,
		DingTalkConnectSyncCorpEmailAttrKey:    req.DingTalkConnectSyncCorpEmailAttrKey,
		DingTalkConnectSyncDisplayNameAttrKey:  req.DingTalkConnectSyncDisplayNameAttrKey,
		DingTalkConnectSyncDeptAttrKey:         req.DingTalkConnectSyncDeptAttrKey,
		DingTalkConnectSyncCorpEmailAttrName:   req.DingTalkConnectSyncCorpEmailAttrName,
		DingTalkConnectSyncDisplayNameAttrName: req.DingTalkConnectSyncDisplayNameAttrName,
		DingTalkConnectSyncDeptAttrName:        req.DingTalkConnectSyncDeptAttrName,
		WeChatConnectEnabled:                   req.WeChatConnectEnabled,
		WeChatConnectAppID:                     req.WeChatConnectAppID,
		WeChatConnectAppSecret:                 req.WeChatConnectAppSecret,
		WeChatConnectOpenAppID:                 req.WeChatConnectOpenAppID,
		WeChatConnectOpenAppSecret:             req.WeChatConnectOpenAppSecret,
		WeChatConnectMPAppID:                   req.WeChatConnectMPAppID,
		WeChatConnectMPAppSecret:               req.WeChatConnectMPAppSecret,
		WeChatConnectMobileAppID:               req.WeChatConnectMobileAppID,
		WeChatConnectMobileAppSecret:           req.WeChatConnectMobileAppSecret,
		WeChatConnectOpenEnabled:               req.WeChatConnectOpenEnabled,
		WeChatConnectMPEnabled:                 req.WeChatConnectMPEnabled,
		WeChatConnectMobileEnabled:             req.WeChatConnectMobileEnabled,
		WeChatConnectMode:                      req.WeChatConnectMode,
		WeChatConnectScopes:                    req.WeChatConnectScopes,
		WeChatConnectRedirectURL:               req.WeChatConnectRedirectURL,
		WeChatConnectFrontendRedirectURL:       req.WeChatConnectFrontendRedirectURL,
		OIDCConnectEnabled:                     req.OIDCConnectEnabled,
		OIDCConnectProviderName:                req.OIDCConnectProviderName,
		OIDCConnectClientID:                    req.OIDCConnectClientID,
		OIDCConnectClientSecret:                req.OIDCConnectClientSecret,
		OIDCConnectIssuerURL:                   req.OIDCConnectIssuerURL,
		OIDCConnectDiscoveryURL:                req.OIDCConnectDiscoveryURL,
		OIDCConnectAuthorizeURL:                req.OIDCConnectAuthorizeURL,
		OIDCConnectTokenURL:                    req.OIDCConnectTokenURL,
		OIDCConnectUserInfoURL:                 req.OIDCConnectUserInfoURL,
		OIDCConnectJWKSURL:                     req.OIDCConnectJWKSURL,
		OIDCConnectScopes:                      req.OIDCConnectScopes,
		OIDCConnectRedirectURL:                 req.OIDCConnectRedirectURL,
		OIDCConnectFrontendRedirectURL:         req.OIDCConnectFrontendRedirectURL,
		OIDCConnectTokenAuthMethod:             req.OIDCConnectTokenAuthMethod,
		OIDCConnectUsePKCE:                     oidcUsePKCE,
		OIDCConnectValidateIDToken:             oidcValidateIDToken,
		OIDCConnectAllowedSigningAlgs:          req.OIDCConnectAllowedSigningAlgs,
		OIDCConnectClockSkewSeconds:            req.OIDCConnectClockSkewSeconds,
		OIDCConnectRequireEmailVerified:        req.OIDCConnectRequireEmailVerified,
		OIDCConnectUserInfoEmailPath:           req.OIDCConnectUserInfoEmailPath,
		OIDCConnectUserInfoIDPath:              req.OIDCConnectUserInfoIDPath,
		OIDCConnectUserInfoUsernamePath:        req.OIDCConnectUserInfoUsernamePath,
		GitHubOAuthEnabled:                     req.GitHubOAuthEnabled,
		GitHubOAuthClientID:                    req.GitHubOAuthClientID,
		GitHubOAuthClientSecret:                req.GitHubOAuthClientSecret,
		GitHubOAuthRedirectURL:                 req.GitHubOAuthRedirectURL,
		GitHubOAuthFrontendRedirectURL:         req.GitHubOAuthFrontendRedirectURL,
		GoogleOAuthEnabled:                     req.GoogleOAuthEnabled,
		GoogleOAuthClientID:                    req.GoogleOAuthClientID,
		GoogleOAuthClientSecret:                req.GoogleOAuthClientSecret,
		GoogleOAuthRedirectURL:                 req.GoogleOAuthRedirectURL,
		GoogleOAuthFrontendRedirectURL:         req.GoogleOAuthFrontendRedirectURL,
		SiteName:                               req.SiteName,
		SiteLogo:                               req.SiteLogo,
		SiteSubtitle:                           req.SiteSubtitle,
		APIBaseURL:                             req.APIBaseURL,
		ContactInfo:                            req.ContactInfo,
		DocURL:                                 req.DocURL,
		HomeContent:                            req.HomeContent,
		HideCcsImportButton:                    req.HideCcsImportButton,
		PurchaseSubscriptionEnabled:            purchaseEnabled,
		PurchaseSubscriptionURL:                purchaseURL,
		TableDefaultPageSize:                   req.TableDefaultPageSize,
		TablePageSizeOptions:                   req.TablePageSizeOptions,
		CustomMenuItems:                        customMenuJSON,
		CustomEndpoints:                        customEndpointsJSON,
		DefaultConcurrency:                     req.DefaultConcurrency,
		DefaultBalance:                         req.DefaultBalance,
		AffiliateRebateRate:                    affiliateRebateRate,
		AffiliateRebateFreezeHours:             affiliateRebateFreezeHours,
		AffiliateRebateDurationDays:            affiliateRebateDurationDays,
		AffiliateRebatePerInviteeCap:           affiliateRebatePerInviteeCap,
		DefaultUserRPMLimit:                    req.DefaultUserRPMLimit,
		DefaultSubscriptions:                   defaultSubscriptions,
		EnableModelFallback:                    req.EnableModelFallback,
		FallbackModelAnthropic:                 req.FallbackModelAnthropic,
		FallbackModelOpenAI:                    req.FallbackModelOpenAI,
		FallbackModelGemini:                    req.FallbackModelGemini,
		FallbackModelAntigravity:               req.FallbackModelAntigravity,
		EnableIdentityPatch:                    req.EnableIdentityPatch,
		IdentityPatchPrompt:                    req.IdentityPatchPrompt,
		MinClaudeCodeVersion:                   req.MinClaudeCodeVersion,
		MaxClaudeCodeVersion:                   req.MaxClaudeCodeVersion,
		AllowUngroupedKeyScheduling:            req.AllowUngroupedKeyScheduling,
		BackendModeEnabled:                     req.BackendModeEnabled,
		ConversationAuditSecondaryPasswordHash: conversationAuditPasswordHash,
		ConversationAuditCleanupEnabled:        conversationAuditCleanupEnabled,
		ConversationAuditRetentionDays:         conversationAuditRetentionDays,
		AllowUserViewErrorRequests: func() bool {
			if req.AllowUserViewErrorRequests != nil {
				return *req.AllowUserViewErrorRequests
			}
			return previousSettings.AllowUserViewErrorRequests
		}(),
		OpsMonitoringEnabled: func() bool {
			if req.OpsMonitoringEnabled != nil {
				return *req.OpsMonitoringEnabled
			}
			return previousSettings.OpsMonitoringEnabled
		}(),
		OpsRealtimeMonitoringEnabled: func() bool {
			if req.OpsRealtimeMonitoringEnabled != nil {
				return *req.OpsRealtimeMonitoringEnabled
			}
			return previousSettings.OpsRealtimeMonitoringEnabled
		}(),
		OpsQueryModeDefault: func() string {
			if req.OpsQueryModeDefault != nil {
				return *req.OpsQueryModeDefault
			}
			return previousSettings.OpsQueryModeDefault
		}(),
		OpsMetricsIntervalSeconds: func() int {
			if req.OpsMetricsIntervalSeconds != nil {
				return *req.OpsMetricsIntervalSeconds
			}
			return previousSettings.OpsMetricsIntervalSeconds
		}(),
		EnableFingerprintUnification: func() bool {
			if req.EnableFingerprintUnification != nil {
				return *req.EnableFingerprintUnification
			}
			return previousSettings.EnableFingerprintUnification
		}(),
		EnableMetadataPassthrough: func() bool {
			if req.EnableMetadataPassthrough != nil {
				return *req.EnableMetadataPassthrough
			}
			return previousSettings.EnableMetadataPassthrough
		}(),
		EnableCCHSigning: func() bool {
			if req.EnableCCHSigning != nil {
				return *req.EnableCCHSigning
			}
			return previousSettings.EnableCCHSigning
		}(),
		EnableClaudeOAuthSystemPromptInjection: func() bool {
			if req.EnableClaudeOAuthSystemPromptInjection != nil {
				return *req.EnableClaudeOAuthSystemPromptInjection
			}
			return previousSettings.EnableClaudeOAuthSystemPromptInjection
		}(),
		ClaudeOAuthSystemPrompt: func() string {
			if req.ClaudeOAuthSystemPrompt != nil {
				return *req.ClaudeOAuthSystemPrompt
			}
			return previousSettings.ClaudeOAuthSystemPrompt
		}(),
		ClaudeOAuthSystemPromptBlocks: func() string {
			if req.ClaudeOAuthSystemPromptBlocks != nil {
				return *req.ClaudeOAuthSystemPromptBlocks
			}
			return previousSettings.ClaudeOAuthSystemPromptBlocks
		}(),
		EnableAnthropicCacheTTL1hInjection: func() bool {
			if req.EnableAnthropicCacheTTL1hInjection != nil {
				return *req.EnableAnthropicCacheTTL1hInjection
			}
			return previousSettings.EnableAnthropicCacheTTL1hInjection
		}(),
		RewriteMessageCacheControl: func() bool {
			if req.RewriteMessageCacheControl != nil {
				return *req.RewriteMessageCacheControl
			}
			return previousSettings.RewriteMessageCacheControl
		}(),
		AntigravityUserAgentVersion: func() string {
			if req.AntigravityUserAgentVersion != nil {
				return *req.AntigravityUserAgentVersion
			}
			return previousSettings.AntigravityUserAgentVersion
		}(),
		OpenAICodexUserAgent: func() string {
			if req.OpenAICodexUserAgent != nil {
				return *req.OpenAICodexUserAgent
			}
			return previousSettings.OpenAICodexUserAgent
		}(),
		OpenAIAllowClaudeCodeCodexPlugin: func() bool {
			if req.OpenAIAllowClaudeCodeCodexPlugin != nil {
				return *req.OpenAIAllowClaudeCodeCodexPlugin
			}
			return previousSettings.OpenAIAllowClaudeCodeCodexPlugin
		}(),
		PaymentVisibleMethodAlipaySource: func() string {
			if req.PaymentVisibleMethodAlipaySource != nil {
				return strings.TrimSpace(*req.PaymentVisibleMethodAlipaySource)
			}
			return previousSettings.PaymentVisibleMethodAlipaySource
		}(),
		PaymentVisibleMethodWxpaySource: func() string {
			if req.PaymentVisibleMethodWxpaySource != nil {
				return strings.TrimSpace(*req.PaymentVisibleMethodWxpaySource)
			}
			return previousSettings.PaymentVisibleMethodWxpaySource
		}(),
		PaymentVisibleMethodAlipayEnabled: func() bool {
			if req.PaymentVisibleMethodAlipayEnabled != nil {
				return *req.PaymentVisibleMethodAlipayEnabled
			}
			return previousSettings.PaymentVisibleMethodAlipayEnabled
		}(),
		PaymentVisibleMethodWxpayEnabled: func() bool {
			if req.PaymentVisibleMethodWxpayEnabled != nil {
				return *req.PaymentVisibleMethodWxpayEnabled
			}
			return previousSettings.PaymentVisibleMethodWxpayEnabled
		}(),
		OpenAIAdvancedSchedulerEnabled: func() bool {
			if req.OpenAIAdvancedSchedulerEnabled != nil {
				return *req.OpenAIAdvancedSchedulerEnabled
			}
			return previousSettings.OpenAIAdvancedSchedulerEnabled
		}(),
		BalanceLowNotifyEnabled: func() bool {
			if req.BalanceLowNotifyEnabled != nil {
				return *req.BalanceLowNotifyEnabled
			}
			return previousSettings.BalanceLowNotifyEnabled
		}(),
		BalanceLowNotifyThreshold: func() float64 {
			if req.BalanceLowNotifyThreshold != nil {
				return *req.BalanceLowNotifyThreshold
			}
			return previousSettings.BalanceLowNotifyThreshold
		}(),
		BalanceLowNotifyRechargeURL: func() string {
			if req.BalanceLowNotifyRechargeURL != nil {
				return *req.BalanceLowNotifyRechargeURL
			}
			return previousSettings.BalanceLowNotifyRechargeURL
		}(),
		SubscriptionExpiryNotifyEnabled: func() bool {
			if req.SubscriptionExpiryNotifyEnabled != nil {
				return *req.SubscriptionExpiryNotifyEnabled
			}
			return previousSettings.SubscriptionExpiryNotifyEnabled
		}(),
		AccountQuotaNotifyEnabled: func() bool {
			if req.AccountQuotaNotifyEnabled != nil {
				return *req.AccountQuotaNotifyEnabled
			}
			return previousSettings.AccountQuotaNotifyEnabled
		}(),
		AccountQuotaNotifyEmails: func() []service.NotifyEmailEntry {
			if req.AccountQuotaNotifyEmails != nil {
				return dto.NotifyEmailEntriesToService(*req.AccountQuotaNotifyEmails)
			}
			return previousSettings.AccountQuotaNotifyEmails
		}(),
		ChannelMonitorEnabled: func() bool {
			if req.ChannelMonitorEnabled != nil {
				return *req.ChannelMonitorEnabled
			}
			return previousSettings.ChannelMonitorEnabled
		}(),
		ChannelMonitorDefaultIntervalSeconds: func() int {
			if req.ChannelMonitorDefaultIntervalSeconds != nil {
				return *req.ChannelMonitorDefaultIntervalSeconds
			}
			return previousSettings.ChannelMonitorDefaultIntervalSeconds
		}(),
		AvailableChannelsEnabled: func() bool {
			if req.AvailableChannelsEnabled != nil {
				return *req.AvailableChannelsEnabled
			}
			return previousSettings.AvailableChannelsEnabled
		}(),
		AffiliateEnabled: func() bool {
			if req.AffiliateEnabled != nil {
				return *req.AffiliateEnabled
			}
			return previousSettings.AffiliateEnabled
		}(),
		RiskControlEnabled: func() bool {
			if req.RiskControlEnabled != nil {
				return *req.RiskControlEnabled
			}
			return previousSettings.RiskControlEnabled
		}(),
		CyberSessionBlockEnabled: func() bool {
			if req.CyberSessionBlockEnabled != nil {
				return *req.CyberSessionBlockEnabled
			}
			return previousSettings.CyberSessionBlockEnabled
		}(),
		CyberSessionBlockTTLSeconds: func() int {
			if req.CyberSessionBlockTTLSeconds != nil {
				return *req.CyberSessionBlockTTLSeconds
			}
			return previousSettings.CyberSessionBlockTTLSeconds
		}(),
	}
	// Preserve existing auth-source defaults when omitted.
	authSourceDefaults := &service.AuthSourceDefaultSettings{
		Email: service.ProviderDefaultGrantSettings{
			Balance:          float64ValueOrDefault(req.AuthSourceDefaultEmailBalance, previousAuthSourceDefaults.Email.Balance),
			Concurrency:      intValueOrDefault(req.AuthSourceDefaultEmailConcurrency, previousAuthSourceDefaults.Email.Concurrency),
			Subscriptions:    defaultSubscriptionsValueOrDefault(req.AuthSourceDefaultEmailSubscriptions, previousAuthSourceDefaults.Email.Subscriptions),
			GrantOnSignup:    boolValueOrDefault(req.AuthSourceDefaultEmailGrantOnSignup, previousAuthSourceDefaults.Email.GrantOnSignup),
			GrantOnFirstBind: boolValueOrDefault(req.AuthSourceDefaultEmailGrantOnFirstBind, previousAuthSourceDefaults.Email.GrantOnFirstBind),
			PlatformQuotas:   platformQuotasValueOrDefault(req.AuthSourceEmailPlatformQuotas, previousAuthSourceDefaults.Email.PlatformQuotas),
		},
		LinuxDo: service.ProviderDefaultGrantSettings{
			Balance:          float64ValueOrDefault(req.AuthSourceDefaultLinuxDoBalance, previousAuthSourceDefaults.LinuxDo.Balance),
			Concurrency:      intValueOrDefault(req.AuthSourceDefaultLinuxDoConcurrency, previousAuthSourceDefaults.LinuxDo.Concurrency),
			Subscriptions:    defaultSubscriptionsValueOrDefault(req.AuthSourceDefaultLinuxDoSubscriptions, previousAuthSourceDefaults.LinuxDo.Subscriptions),
			GrantOnSignup:    boolValueOrDefault(req.AuthSourceDefaultLinuxDoGrantOnSignup, previousAuthSourceDefaults.LinuxDo.GrantOnSignup),
			GrantOnFirstBind: boolValueOrDefault(req.AuthSourceDefaultLinuxDoGrantOnFirstBind, previousAuthSourceDefaults.LinuxDo.GrantOnFirstBind),
			PlatformQuotas:   platformQuotasValueOrDefault(req.AuthSourceLinuxDoPlatformQuotas, previousAuthSourceDefaults.LinuxDo.PlatformQuotas),
		},
		OIDC: service.ProviderDefaultGrantSettings{
			Balance:          float64ValueOrDefault(req.AuthSourceDefaultOIDCBalance, previousAuthSourceDefaults.OIDC.Balance),
			Concurrency:      intValueOrDefault(req.AuthSourceDefaultOIDCConcurrency, previousAuthSourceDefaults.OIDC.Concurrency),
			Subscriptions:    defaultSubscriptionsValueOrDefault(req.AuthSourceDefaultOIDCSubscriptions, previousAuthSourceDefaults.OIDC.Subscriptions),
			GrantOnSignup:    boolValueOrDefault(req.AuthSourceDefaultOIDCGrantOnSignup, previousAuthSourceDefaults.OIDC.GrantOnSignup),
			GrantOnFirstBind: boolValueOrDefault(req.AuthSourceDefaultOIDCGrantOnFirstBind, previousAuthSourceDefaults.OIDC.GrantOnFirstBind),
			PlatformQuotas:   platformQuotasValueOrDefault(req.AuthSourceOIDCPlatformQuotas, previousAuthSourceDefaults.OIDC.PlatformQuotas),
		},
		WeChat: service.ProviderDefaultGrantSettings{
			Balance:          float64ValueOrDefault(req.AuthSourceDefaultWeChatBalance, previousAuthSourceDefaults.WeChat.Balance),
			Concurrency:      intValueOrDefault(req.AuthSourceDefaultWeChatConcurrency, previousAuthSourceDefaults.WeChat.Concurrency),
			Subscriptions:    defaultSubscriptionsValueOrDefault(req.AuthSourceDefaultWeChatSubscriptions, previousAuthSourceDefaults.WeChat.Subscriptions),
			GrantOnSignup:    boolValueOrDefault(req.AuthSourceDefaultWeChatGrantOnSignup, previousAuthSourceDefaults.WeChat.GrantOnSignup),
			GrantOnFirstBind: boolValueOrDefault(req.AuthSourceDefaultWeChatGrantOnFirstBind, previousAuthSourceDefaults.WeChat.GrantOnFirstBind),
			PlatformQuotas:   platformQuotasValueOrDefault(req.AuthSourceWeChatPlatformQuotas, previousAuthSourceDefaults.WeChat.PlatformQuotas),
		},
		GitHub: service.ProviderDefaultGrantSettings{
			Balance:          float64ValueOrDefault(req.AuthSourceDefaultGitHubBalance, previousAuthSourceDefaults.GitHub.Balance),
			Concurrency:      intValueOrDefault(req.AuthSourceDefaultGitHubConcurrency, previousAuthSourceDefaults.GitHub.Concurrency),
			Subscriptions:    defaultSubscriptionsValueOrDefault(req.AuthSourceDefaultGitHubSubscriptions, previousAuthSourceDefaults.GitHub.Subscriptions),
			GrantOnSignup:    boolValueOrDefault(req.AuthSourceDefaultGitHubGrantOnSignup, previousAuthSourceDefaults.GitHub.GrantOnSignup),
			GrantOnFirstBind: boolValueOrDefault(req.AuthSourceDefaultGitHubGrantOnFirstBind, previousAuthSourceDefaults.GitHub.GrantOnFirstBind),
			PlatformQuotas:   platformQuotasValueOrDefault(req.AuthSourceGitHubPlatformQuotas, previousAuthSourceDefaults.GitHub.PlatformQuotas),
		},
		Google: service.ProviderDefaultGrantSettings{
			Balance:          float64ValueOrDefault(req.AuthSourceDefaultGoogleBalance, previousAuthSourceDefaults.Google.Balance),
			Concurrency:      intValueOrDefault(req.AuthSourceDefaultGoogleConcurrency, previousAuthSourceDefaults.Google.Concurrency),
			Subscriptions:    defaultSubscriptionsValueOrDefault(req.AuthSourceDefaultGoogleSubscriptions, previousAuthSourceDefaults.Google.Subscriptions),
			GrantOnSignup:    boolValueOrDefault(req.AuthSourceDefaultGoogleGrantOnSignup, previousAuthSourceDefaults.Google.GrantOnSignup),
			GrantOnFirstBind: boolValueOrDefault(req.AuthSourceDefaultGoogleGrantOnFirstBind, previousAuthSourceDefaults.Google.GrantOnFirstBind),
			PlatformQuotas:   platformQuotasValueOrDefault(req.AuthSourceGooglePlatformQuotas, previousAuthSourceDefaults.Google.PlatformQuotas),
		},
		DingTalk: service.ProviderDefaultGrantSettings{
			Balance:          float64ValueOrDefault(req.AuthSourceDefaultDingTalkBalance, previousAuthSourceDefaults.DingTalk.Balance),
			Concurrency:      intValueOrDefault(req.AuthSourceDefaultDingTalkConcurrency, previousAuthSourceDefaults.DingTalk.Concurrency),
			Subscriptions:    defaultSubscriptionsValueOrDefault(req.AuthSourceDefaultDingTalkSubscriptions, previousAuthSourceDefaults.DingTalk.Subscriptions),
			GrantOnSignup:    boolValueOrDefault(req.AuthSourceDefaultDingTalkGrantOnSignup, previousAuthSourceDefaults.DingTalk.GrantOnSignup),
			GrantOnFirstBind: boolValueOrDefault(req.AuthSourceDefaultDingTalkGrantOnFirstBind, previousAuthSourceDefaults.DingTalk.GrantOnFirstBind),
			PlatformQuotas:   platformQuotasValueOrDefault(req.AuthSourceDingTalkPlatformQuotas, previousAuthSourceDefaults.DingTalk.PlatformQuotas),
		},
		ForceEmailOnThirdPartySignup: boolValueOrDefault(req.ForceEmailOnThirdPartySignup, previousAuthSourceDefaults.ForceEmailOnThirdPartySignup),
	}
	if err := h.settingService.UpdateSettingsWithAuthSourceDefaults(c.Request.Context(), settings, authSourceDefaults); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Update OpenAI fast policy (stored under dedicated key, only when provided).
	if req.OpenAIFastPolicySettings != nil {
		if err := h.settingService.SetOpenAIFastPolicySettings(c.Request.Context(), openaiFastPolicySettingsFromDTO(req.OpenAIFastPolicySettings)); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}

	// Update payment configuration (integrated into system settings).
	// Skip if no payment fields were provided (prevents accidental wipe).
	if h.paymentConfigService != nil && hasPaymentFields(req) {
		paymentReq := service.UpdatePaymentConfigRequest{
			Enabled:                   req.PaymentEnabled,
			MinAmount:                 req.PaymentMinAmount,
			MaxAmount:                 req.PaymentMaxAmount,
			DailyLimit:                req.PaymentDailyLimit,
			OrderTimeoutMin:           req.PaymentOrderTimeoutMin,
			MaxPendingOrders:          req.PaymentMaxPendingOrders,
			EnabledTypes:              req.PaymentEnabledTypes,
			BalanceDisabled:           req.PaymentBalanceDisabled,
			BalanceRechargeMultiplier: req.PaymentBalanceRechargeMultiplier,
			RechargeFeeRate:           req.PaymentRechargeFeeRate,
			LoadBalanceStrategy:       req.PaymentLoadBalanceStrat,
			ProductNamePrefix:         req.PaymentProductNamePrefix,
			ProductNameSuffix:         req.PaymentProductNameSuffix,
			HelpImageURL:              req.PaymentHelpImageURL,
			HelpText:                  req.PaymentHelpText,
			CancelRateLimitEnabled:    req.PaymentCancelRateLimitEnabled,
			CancelRateLimitMax:        req.PaymentCancelRateLimitMax,
			CancelRateLimitWindow:     req.PaymentCancelRateLimitWindow,
			CancelRateLimitUnit:       req.PaymentCancelRateLimitUnit,
			CancelRateLimitMode:       req.PaymentCancelRateLimitMode,
			AlipayForceQRCode:         req.PaymentAlipayForceQRCode,
		}
		if err := h.paymentConfigService.UpdatePaymentConfig(c.Request.Context(), paymentReq); err != nil {
			response.ErrorFrom(c, err)
			return
		}
		// Refresh in-memory provider registry so config changes take effect immediately
		if h.paymentService != nil {
			h.paymentService.RefreshProviders(c.Request.Context())
		}
	}

	h.auditSettingsUpdate(c, previousSettings, settings, previousAuthSourceDefaults, authSourceDefaults, req)

	// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮村鑸殿€栨繛瀛樼矊缂嶅﹪寮诲☉銏犵疀闁稿繐鎽滈弫鏍⒑濞茶骞楅柟鐟版喘瀵鏁愭径濠庢綂闂侀潧绻嗛弲婵嬪礉閹间焦鈷戦柦妯侯槸閺嗙喖鏌涢悩鏌ュ弰闁糕晝鍋ら獮瀣晜閽樺姹楅梻浣告啞閻熴儵藝椤栫倛鍥煛閸涱喒鎷洪梺纭呭亹閸嬫盯鍩€椤掍胶澧い顐㈢箲缁绘繂顫濋鍌氬Е婵＄偑鍊栫敮鎺楀磹瑜版帒鍚归柍褜鍓熼弻锝嗘償閵忕姴姣堥梺鍛娒妶鎼佸灳閿曞倸鍨傛い鏂诲劤閸犳牠骞冮埄鍐╁劅闁挎繂娲ㄩ幗宀勬⒒閸屾瑨鍏屾い顐㈩儔瀹曠喖宕归銈嗘闂傚倷鑳堕…鍫ヮ敄閹寸姴绶ゅù鐘差儐缁犳帡姊绘担鐟邦嚋缂佽鍊搁湁闁稿瞼鍋為崐鍨旈敐鍛殲闁绘挸鍟村娲垂椤曞懎鍓卞銈冨劚閻楁捇寮婚垾宕囨殕閻庯綆鍓涜ⅵ濠电姷顣介埀顒€纾崺锝団偓瑙勬磸閸旀垿銆佸璺哄窛妞ゆ洖妫涚壕鍏肩節閻㈤潧啸闁轰焦鎮傚畷鎴︽偄閻撳骸鍋嶉悷婊勬閸ㄩ箖鏁冮埀顒勬偩閿熺姵鐒介柨鏇楀亾婵炲牄鍔岄—鍐Χ閸℃鐟ㄩ柣搴㈠嚬閸撴稓鍒掔紒妯侯嚤闁哄鍨归鎰版煛婢跺﹦澧曞褏鏅划鏃堝垂椤愶紕绠?
	updatedSettings, err := h.settingService.GetAllSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	h.ensureDingTalkSyncAttributes(c.Request.Context(), updatedSettings)
	updatedAuthSourceDefaults, err := h.settingService.GetAuthSourceDefaultSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	updatedDefaultSubscriptions := make([]dto.DefaultSubscriptionSetting, 0, len(updatedSettings.DefaultSubscriptions))
	for _, sub := range updatedSettings.DefaultSubscriptions {
		updatedDefaultSubscriptions = append(updatedDefaultSubscriptions, dto.DefaultSubscriptionSetting{
			GroupID:      sub.GroupID,
			ValidityDays: sub.ValidityDays,
		})
	}

	// Reload payment config for response
	var updatedPaymentCfg *service.PaymentConfig
	if h.paymentConfigService != nil {
		updatedPaymentCfg, _ = h.paymentConfigService.GetPaymentConfig(c.Request.Context())
	}
	if updatedPaymentCfg == nil {
		updatedPaymentCfg = &service.PaymentConfig{}
	}

	payload := dto.SystemSettings{
		RegistrationEnabled:                          updatedSettings.RegistrationEnabled,
		EmailVerifyEnabled:                           updatedSettings.EmailVerifyEnabled,
		RegistrationEmailSuffixWhitelist:             updatedSettings.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                             updatedSettings.PromoCodeEnabled,
		PasswordResetEnabled:                         updatedSettings.PasswordResetEnabled,
		FrontendURL:                                  updatedSettings.FrontendURL,
		InvitationCodeEnabled:                        updatedSettings.InvitationCodeEnabled,
		TotpEnabled:                                  updatedSettings.TotpEnabled,
		TotpEncryptionKeyConfigured:                  h.settingService.IsTotpEncryptionKeyConfigured(),
		LoginAgreementEnabled:                        updatedSettings.LoginAgreementEnabled,
		ConversationAuditSecondaryPasswordConfigured: updatedSettings.ConversationAuditSecondaryPasswordConfigured,
		ConversationAuditCleanupEnabled:              updatedSettings.ConversationAuditCleanupEnabled,
		ConversationAuditRetentionDays:               updatedSettings.ConversationAuditRetentionDays,
		LoginAgreementMode:                           updatedSettings.LoginAgreementMode,
		LoginAgreementUpdatedAt:                      updatedSettings.LoginAgreementUpdatedAt,
		LoginAgreementDocuments:                      loginAgreementDocumentsToDTO(updatedSettings.LoginAgreementDocuments),
		SMTPHost:                                     updatedSettings.SMTPHost,
		SMTPPort:                                     updatedSettings.SMTPPort,
		SMTPUsername:                                 updatedSettings.SMTPUsername,
		SMTPPasswordConfigured:                       updatedSettings.SMTPPasswordConfigured,
		SMTPFrom:                                     updatedSettings.SMTPFrom,
		SMTPFromName:                                 updatedSettings.SMTPFromName,
		SMTPUseTLS:                                   updatedSettings.SMTPUseTLS,
		TurnstileEnabled:                             updatedSettings.TurnstileEnabled,
		TurnstileSiteKey:                             updatedSettings.TurnstileSiteKey,
		TurnstileSecretKeyConfigured:                 updatedSettings.TurnstileSecretKeyConfigured,
		APIKeyACLTrustForwardedIP:                    updatedSettings.APIKeyACLTrustForwardedIP,
		LinuxDoConnectEnabled:                        updatedSettings.LinuxDoConnectEnabled,
		LinuxDoConnectClientID:                       updatedSettings.LinuxDoConnectClientID,
		LinuxDoConnectClientSecretConfigured:         updatedSettings.LinuxDoConnectClientSecretConfigured,
		LinuxDoConnectRedirectURL:                    updatedSettings.LinuxDoConnectRedirectURL,
		DingTalkConnectEnabled:                       updatedSettings.DingTalkConnectEnabled,
		DingTalkConnectClientID:                      updatedSettings.DingTalkConnectClientID,
		DingTalkConnectClientSecretConfigured:        updatedSettings.DingTalkConnectClientSecretConfigured,
		DingTalkConnectRedirectURL:                   updatedSettings.DingTalkConnectRedirectURL,
		DingTalkConnectCorpRestrictionPolicy:         updatedSettings.DingTalkConnectCorpRestrictionPolicy,
		DingTalkConnectInternalCorpID:                updatedSettings.DingTalkConnectInternalCorpID,
		DingTalkConnectBypassRegistration:            updatedSettings.DingTalkConnectBypassRegistration,
		DingTalkConnectSyncCorpEmail:                 updatedSettings.DingTalkConnectSyncCorpEmail,
		DingTalkConnectSyncDisplayName:               updatedSettings.DingTalkConnectSyncDisplayName,
		DingTalkConnectSyncDept:                      updatedSettings.DingTalkConnectSyncDept,
		DingTalkConnectSyncCorpEmailAttrKey:          updatedSettings.DingTalkConnectSyncCorpEmailAttrKey,
		DingTalkConnectSyncDisplayNameAttrKey:        updatedSettings.DingTalkConnectSyncDisplayNameAttrKey,
		DingTalkConnectSyncDeptAttrKey:               updatedSettings.DingTalkConnectSyncDeptAttrKey,
		DingTalkConnectSyncCorpEmailAttrName:         updatedSettings.DingTalkConnectSyncCorpEmailAttrName,
		DingTalkConnectSyncDisplayNameAttrName:       updatedSettings.DingTalkConnectSyncDisplayNameAttrName,
		DingTalkConnectSyncDeptAttrName:              updatedSettings.DingTalkConnectSyncDeptAttrName,
		WeChatConnectEnabled:                         updatedSettings.WeChatConnectEnabled,
		WeChatConnectAppID:                           updatedSettings.WeChatConnectAppID,
		WeChatConnectAppSecretConfigured:             updatedSettings.WeChatConnectAppSecretConfigured,
		WeChatConnectOpenAppID:                       updatedSettings.WeChatConnectOpenAppID,
		WeChatConnectOpenAppSecretConfigured:         updatedSettings.WeChatConnectOpenAppSecretConfigured,
		WeChatConnectMPAppID:                         updatedSettings.WeChatConnectMPAppID,
		WeChatConnectMPAppSecretConfigured:           updatedSettings.WeChatConnectMPAppSecretConfigured,
		WeChatConnectMobileAppID:                     updatedSettings.WeChatConnectMobileAppID,
		WeChatConnectMobileAppSecretConfigured:       updatedSettings.WeChatConnectMobileAppSecretConfigured,
		WeChatConnectOpenEnabled:                     updatedSettings.WeChatConnectOpenEnabled,
		WeChatConnectMPEnabled:                       updatedSettings.WeChatConnectMPEnabled,
		WeChatConnectMobileEnabled:                   updatedSettings.WeChatConnectMobileEnabled,
		WeChatConnectMode:                            updatedSettings.WeChatConnectMode,
		WeChatConnectScopes:                          updatedSettings.WeChatConnectScopes,
		WeChatConnectRedirectURL:                     updatedSettings.WeChatConnectRedirectURL,
		WeChatConnectFrontendRedirectURL:             updatedSettings.WeChatConnectFrontendRedirectURL,
		OIDCConnectEnabled:                           updatedSettings.OIDCConnectEnabled,
		OIDCConnectProviderName:                      updatedSettings.OIDCConnectProviderName,
		OIDCConnectClientID:                          updatedSettings.OIDCConnectClientID,
		OIDCConnectClientSecretConfigured:            updatedSettings.OIDCConnectClientSecretConfigured,
		OIDCConnectIssuerURL:                         updatedSettings.OIDCConnectIssuerURL,
		OIDCConnectDiscoveryURL:                      updatedSettings.OIDCConnectDiscoveryURL,
		OIDCConnectAuthorizeURL:                      updatedSettings.OIDCConnectAuthorizeURL,
		OIDCConnectTokenURL:                          updatedSettings.OIDCConnectTokenURL,
		OIDCConnectUserInfoURL:                       updatedSettings.OIDCConnectUserInfoURL,
		OIDCConnectJWKSURL:                           updatedSettings.OIDCConnectJWKSURL,
		OIDCConnectScopes:                            updatedSettings.OIDCConnectScopes,
		OIDCConnectRedirectURL:                       updatedSettings.OIDCConnectRedirectURL,
		OIDCConnectFrontendRedirectURL:               updatedSettings.OIDCConnectFrontendRedirectURL,
		OIDCConnectTokenAuthMethod:                   updatedSettings.OIDCConnectTokenAuthMethod,
		OIDCConnectUsePKCE:                           updatedSettings.OIDCConnectUsePKCE,
		OIDCConnectValidateIDToken:                   updatedSettings.OIDCConnectValidateIDToken,
		OIDCConnectAllowedSigningAlgs:                updatedSettings.OIDCConnectAllowedSigningAlgs,
		OIDCConnectClockSkewSeconds:                  updatedSettings.OIDCConnectClockSkewSeconds,
		OIDCConnectRequireEmailVerified:              updatedSettings.OIDCConnectRequireEmailVerified,
		OIDCConnectUserInfoEmailPath:                 updatedSettings.OIDCConnectUserInfoEmailPath,
		OIDCConnectUserInfoIDPath:                    updatedSettings.OIDCConnectUserInfoIDPath,
		OIDCConnectUserInfoUsernamePath:              updatedSettings.OIDCConnectUserInfoUsernamePath,
		GitHubOAuthEnabled:                           updatedSettings.GitHubOAuthEnabled,
		GitHubOAuthClientID:                          updatedSettings.GitHubOAuthClientID,
		GitHubOAuthClientSecretConfigured:            updatedSettings.GitHubOAuthClientSecretConfigured,
		GitHubOAuthRedirectURL:                       updatedSettings.GitHubOAuthRedirectURL,
		GitHubOAuthFrontendRedirectURL:               updatedSettings.GitHubOAuthFrontendRedirectURL,
		GoogleOAuthEnabled:                           updatedSettings.GoogleOAuthEnabled,
		GoogleOAuthClientID:                          updatedSettings.GoogleOAuthClientID,
		GoogleOAuthClientSecretConfigured:            updatedSettings.GoogleOAuthClientSecretConfigured,
		GoogleOAuthRedirectURL:                       updatedSettings.GoogleOAuthRedirectURL,
		GoogleOAuthFrontendRedirectURL:               updatedSettings.GoogleOAuthFrontendRedirectURL,
		SiteName:                                     updatedSettings.SiteName,
		SiteLogo:                                     updatedSettings.SiteLogo,
		SiteSubtitle:                                 updatedSettings.SiteSubtitle,
		APIBaseURL:                                   updatedSettings.APIBaseURL,
		ContactInfo:                                  updatedSettings.ContactInfo,
		DocURL:                                       updatedSettings.DocURL,
		HomeContent:                                  updatedSettings.HomeContent,
		HideCcsImportButton:                          updatedSettings.HideCcsImportButton,
		PurchaseSubscriptionEnabled:                  updatedSettings.PurchaseSubscriptionEnabled,
		PurchaseSubscriptionURL:                      updatedSettings.PurchaseSubscriptionURL,
		TableDefaultPageSize:                         updatedSettings.TableDefaultPageSize,
		TablePageSizeOptions:                         updatedSettings.TablePageSizeOptions,
		CustomMenuItems:                              dto.ParseCustomMenuItems(updatedSettings.CustomMenuItems),
		CustomEndpoints:                              dto.ParseCustomEndpoints(updatedSettings.CustomEndpoints),
		DefaultConcurrency:                           updatedSettings.DefaultConcurrency,
		DefaultBalance:                               updatedSettings.DefaultBalance,
		AffiliateRebateRate:                          updatedSettings.AffiliateRebateRate,
		AffiliateRebateFreezeHours:                   updatedSettings.AffiliateRebateFreezeHours,
		AffiliateRebateDurationDays:                  updatedSettings.AffiliateRebateDurationDays,
		AffiliateRebatePerInviteeCap:                 updatedSettings.AffiliateRebatePerInviteeCap,
		DefaultUserRPMLimit:                          updatedSettings.DefaultUserRPMLimit,
		DefaultSubscriptions:                         updatedDefaultSubscriptions,
		EnableModelFallback:                          updatedSettings.EnableModelFallback,
		FallbackModelAnthropic:                       updatedSettings.FallbackModelAnthropic,
		FallbackModelOpenAI:                          updatedSettings.FallbackModelOpenAI,
		FallbackModelGemini:                          updatedSettings.FallbackModelGemini,
		FallbackModelAntigravity:                     updatedSettings.FallbackModelAntigravity,
		EnableIdentityPatch:                          updatedSettings.EnableIdentityPatch,
		IdentityPatchPrompt:                          updatedSettings.IdentityPatchPrompt,
		OpsMonitoringEnabled:                         updatedSettings.OpsMonitoringEnabled,
		OpsRealtimeMonitoringEnabled:                 updatedSettings.OpsRealtimeMonitoringEnabled,
		OpsQueryModeDefault:                          updatedSettings.OpsQueryModeDefault,
		OpsMetricsIntervalSeconds:                    updatedSettings.OpsMetricsIntervalSeconds,
		MinClaudeCodeVersion:                         updatedSettings.MinClaudeCodeVersion,
		MaxClaudeCodeVersion:                         updatedSettings.MaxClaudeCodeVersion,
		AllowUngroupedKeyScheduling:                  updatedSettings.AllowUngroupedKeyScheduling,
		BackendModeEnabled:                           updatedSettings.BackendModeEnabled,
		EnableFingerprintUnification:                 updatedSettings.EnableFingerprintUnification,
		EnableMetadataPassthrough:                    updatedSettings.EnableMetadataPassthrough,
		EnableCCHSigning:                             updatedSettings.EnableCCHSigning,
		EnableClaudeOAuthSystemPromptInjection:       updatedSettings.EnableClaudeOAuthSystemPromptInjection,
		ClaudeOAuthSystemPrompt:                      updatedSettings.ClaudeOAuthSystemPrompt,
		ClaudeOAuthSystemPromptBlocks:                updatedSettings.ClaudeOAuthSystemPromptBlocks,
		EnableAnthropicCacheTTL1hInjection:           updatedSettings.EnableAnthropicCacheTTL1hInjection,
		RewriteMessageCacheControl:                   updatedSettings.RewriteMessageCacheControl,
		AntigravityUserAgentVersion:                  updatedSettings.AntigravityUserAgentVersion,
		OpenAICodexUserAgent:                         updatedSettings.OpenAICodexUserAgent,
		OpenAIAllowClaudeCodeCodexPlugin:             updatedSettings.OpenAIAllowClaudeCodeCodexPlugin,
		PaymentVisibleMethodAlipaySource:             updatedSettings.PaymentVisibleMethodAlipaySource,
		PaymentVisibleMethodWxpaySource:              updatedSettings.PaymentVisibleMethodWxpaySource,
		PaymentVisibleMethodAlipayEnabled:            updatedSettings.PaymentVisibleMethodAlipayEnabled,
		PaymentVisibleMethodWxpayEnabled:             updatedSettings.PaymentVisibleMethodWxpayEnabled,
		OpenAIAdvancedSchedulerEnabled:               updatedSettings.OpenAIAdvancedSchedulerEnabled,
		BalanceLowNotifyEnabled:                      updatedSettings.BalanceLowNotifyEnabled,
		BalanceLowNotifyThreshold:                    updatedSettings.BalanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:                  updatedSettings.BalanceLowNotifyRechargeURL,
		SubscriptionExpiryNotifyEnabled:              updatedSettings.SubscriptionExpiryNotifyEnabled,
		AccountQuotaNotifyEnabled:                    updatedSettings.AccountQuotaNotifyEnabled,
		AccountQuotaNotifyEmails:                     dto.NotifyEmailEntriesFromService(updatedSettings.AccountQuotaNotifyEmails),
		PaymentEnabled:                               updatedPaymentCfg.Enabled,
		PaymentMinAmount:                             updatedPaymentCfg.MinAmount,
		PaymentMaxAmount:                             updatedPaymentCfg.MaxAmount,
		PaymentDailyLimit:                            updatedPaymentCfg.DailyLimit,
		PaymentOrderTimeoutMin:                       updatedPaymentCfg.OrderTimeoutMin,
		PaymentMaxPendingOrders:                      updatedPaymentCfg.MaxPendingOrders,
		PaymentEnabledTypes:                          updatedPaymentCfg.EnabledTypes,
		PaymentBalanceDisabled:                       updatedPaymentCfg.BalanceDisabled,
		PaymentBalanceRechargeMultiplier:             updatedPaymentCfg.BalanceRechargeMultiplier,
		PaymentRechargeFeeRate:                       updatedPaymentCfg.RechargeFeeRate,
		PaymentLoadBalanceStrat:                      updatedPaymentCfg.LoadBalanceStrategy,
		PaymentProductNamePrefix:                     updatedPaymentCfg.ProductNamePrefix,
		PaymentProductNameSuffix:                     updatedPaymentCfg.ProductNameSuffix,
		PaymentHelpImageURL:                          updatedPaymentCfg.HelpImageURL,
		PaymentHelpText:                              updatedPaymentCfg.HelpText,
		PaymentCancelRateLimitEnabled:                updatedPaymentCfg.CancelRateLimitEnabled,
		PaymentCancelRateLimitMax:                    updatedPaymentCfg.CancelRateLimitMax,
		PaymentCancelRateLimitWindow:                 updatedPaymentCfg.CancelRateLimitWindow,
		PaymentCancelRateLimitUnit:                   updatedPaymentCfg.CancelRateLimitUnit,
		PaymentCancelRateLimitMode:                   updatedPaymentCfg.CancelRateLimitMode,
		PaymentAlipayForceQRCode:                     updatedPaymentCfg.AlipayForceQRCode,

		ChannelMonitorEnabled:                updatedSettings.ChannelMonitorEnabled,
		ChannelMonitorDefaultIntervalSeconds: updatedSettings.ChannelMonitorDefaultIntervalSeconds,

		AvailableChannelsEnabled: updatedSettings.AvailableChannelsEnabled,

		AffiliateEnabled: updatedSettings.AffiliateEnabled,

		RiskControlEnabled:          updatedSettings.RiskControlEnabled,
		CyberSessionBlockEnabled:    updatedSettings.CyberSessionBlockEnabled,
		CyberSessionBlockTTLSeconds: updatedSettings.CyberSessionBlockTTLSeconds,
		AllowUserViewErrorRequests:  updatedSettings.AllowUserViewErrorRequests,
	}
	if fastPolicy, err := h.settingService.GetOpenAIFastPolicySettings(c.Request.Context()); err != nil {
		slog.Error("openai_fast_policy_settings_get_failed", "error", err)
	} else if fastPolicy != nil {
		payload.OpenAIFastPolicySettings = openaiFastPolicySettingsToDTO(fastPolicy)
	}

	// Default platform quotas闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢琛″亾濞戞瑯鐒界紒鐘卞嵆閺屻倝鎮ч崼婵愬殝缂備胶绮换鍫ュ箖娴犲鍋ㄩ柛顭戝亞缁夌瘶 map闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢灏佹瀻闂勫洭顢氳鐓ゆい鎿冩娇閳ь兛绶氬鎾閳╁啯鐝栭梻渚€鈧偛鑻晶鎾煙椤旀儳浠辩€规洖缍婇、鏇㈡偐鏉堚晝娉块梻鍌欒兌閹虫捇骞栭銈囩煋闁割偅娲栫粈澶屸偓骞垮劚椤︿即鎮￠悢鍏肩厵闁诡垎灞芥濠电偛寮堕幐鎶藉蓟?婵?GetSettings 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｉ幇顒佹儓闁搞劌鍊块弻娑㈩敃閿濆棛顦ョ紓浣哄С閸楁娊寮婚悢铏圭＜闁靛繒濮甸悘鍫㈢磼閻愵剙鍔ゆい顓犲厴瀵鎮㈤悡搴ｎ槶閻熸粌绻掗弫顔尖槈閵忥紕鍘介梺瑙勫礃濞夋盯鍩㈤崼銉︾厸鐎光偓鐎ｎ剛袦婵犲痉銈呯厫闁诡垱妫冮弫宥夊礋椤忓棙鏆伴梻浣藉吹閸犳劗鍒掑鍡曠箚闁归棿绀佸婵囥亜閺嶃劎鐜荤紒杈ㄧ叀濮婃椽宕崟鍨﹂梺缁橆殔鐎氭澘鐣峰┑鍥ㄥ劅闁挎繂娲ｇ粭澶愭⒑缂佹ê濮夐柛搴涘€濋幃鈥斥槈閵忊€斥偓鍫曟煟閹邦垱纭剧悮姘渻閵堝啫鍔橀柛銊ヮ煼閸┾偓妞ゆ帒鍠氬鎰箾閸欏澧辩紒杈╁仦缁绘繈宕堕妷銏犱壕濞达絿纭堕弸搴ㄦ煙閻愵剚缍戝ù婊勵殜閺岀喖宕楅懖鈺傛闂佺锕ラ〃濠冧繆閸撲讲鍋撳☉娆樼劷缂佲檧鍋撻梻浣圭湽閸ㄨ棄顭囪瀹曟粌鐣烽崶鈺冿紲闂佽壈澹堥崺鏍夋径鎰亗闁靛牆顦伴悡銉╂煛閸愩劌鈧憡绂嶉崜褉鍋撶憴鍕闁绘搫绻濆璇测槈濡吋娈曢梺閫炲苯澧寸€规洘鍨块幃娆撳传閸曨叏绱遍梻渚€娼ч悧鍡浰囨导瀛樺亗闊洦鎸撮弨浠嬫煟閹扮増娑уù婊呭亾缁绘盯宕ㄩ鐔锋畻闂佽鍠栭崲鏌ュ煝鎼粹檧鏋庨柟瀵稿У濞堝吋淇婇妶鍥ラ柛瀣仱閹冾煥閸繄顦梺鍦劋椤ㄥ棝宕戦敓鐘崇厸濠㈣泛锕﹀銊╂煙閺嬵偄鈧繂顫忕紒妯诲缂佹稑顑呭▓鎰版⒑閹肩偛濮傚ù婊冪埣楠炲啴鎮滈挊澶屽幐闂佸憡渚楅崹鎵矈閿曞倹鈷戦柛婵嗗閳诲鏌涚€Ｑ冧壕闂備胶顭堥鍡涘箰閹间礁鐓″璺号堥弸宥夋煣韫囷絽浜滈柣蹇旀尦閺屾盯鍩￠崒婊勫垱濡ょ姷鍋炵敮鈥愁嚕椤掑嫬浼犻柛鏇ㄤ簻椤ユ岸姊洪懡銈呮瀾闁荤喆鍎抽埀顒佸嚬閸ｏ綁鐛Δ鍛妞ゎ厽鍨剁€靛矂鏌ｆ惔銊︽锭濠㈢懓锕幃鐐烘倷椤戝彞绨婚梺鍝勬处閿氶柛鏃撳閳ь剝顫夊ú婊堝窗閺嶎厹鈧礁螖娴ｇ懓顎撻梺鎯х箰婢э綁濡堕崱鏇犵畾闂佺粯鍔︽禍婊堝焵椤掍胶澧电€规洘绻堥弫鍐磼濮樻唻绱遍梻浣告啞濞诧箓宕戦埀顒勬煛鐎ｎ偅鈷愮紒缁樼洴瀹曞崬螣鐠囨煡鐎洪梻浣虹帛閹搁箖宕伴弽顓犲祦闁哄稁鐏旀惔顭戞晢闁逞屽墴閹﹢顢氶埀顒勫蓟瀹ュ鏁嶆繛鎴炵懅椤︺劑姊洪棃娑欐悙閻庢氨澧楁穱濠囧箹娴ｈ倽銊╂煥閺冨倻甯涢柨娑樻娣?
	if platformQuotas, err := h.settingService.GetDefaultPlatformQuotas(c.Request.Context()); err != nil {
		slog.Error("default_platform_quotas_get_failed", "error", err)
	} else {
		payload.DefaultPlatformQuotas = platformQuotas
	}
	response.Success(c, systemSettingsResponseData(payload, updatedAuthSourceDefaults))
}

// hasPaymentFields returns true if any payment-related field was explicitly provided.
// mapDingTalkValidateError maps ValidateDingTalkConfig errors to machine-readable reason codes.
func mapDingTalkValidateError(err error) string {
	switch {
	case errors.Is(err, config.ErrDingTalkV1AppTypeMismatch):
		return "dingtalk_apptype_mismatch"
	case errors.Is(err, config.ErrDingTalkV4InvalidAppKind):
		return "dingtalk_app_kind_invalid"
	default:
		return "dingtalk_corp_config_invalid"
	}
}

func hasPaymentFields(req UpdateSettingsRequest) bool {
	return req.PaymentEnabled != nil || req.PaymentMinAmount != nil ||
		req.PaymentMaxAmount != nil || req.PaymentDailyLimit != nil ||
		req.PaymentOrderTimeoutMin != nil || req.PaymentMaxPendingOrders != nil ||
		req.PaymentEnabledTypes != nil || req.PaymentBalanceDisabled != nil ||
		req.PaymentBalanceRechargeMultiplier != nil || req.PaymentRechargeFeeRate != nil ||
		req.PaymentLoadBalanceStrat != nil || req.PaymentProductNamePrefix != nil ||
		req.PaymentProductNameSuffix != nil || req.PaymentHelpImageURL != nil ||
		req.PaymentHelpText != nil || req.PaymentCancelRateLimitEnabled != nil ||
		req.PaymentCancelRateLimitMax != nil || req.PaymentCancelRateLimitWindow != nil ||
		req.PaymentCancelRateLimitUnit != nil || req.PaymentCancelRateLimitMode != nil ||
		req.PaymentAlipayForceQRCode != nil
}

func (h *SettingHandler) auditSettingsUpdate(c *gin.Context, before *service.SystemSettings, after *service.SystemSettings, beforeAuthSourceDefaults *service.AuthSourceDefaultSettings, afterAuthSourceDefaults *service.AuthSourceDefaultSettings, req UpdateSettingsRequest) {
	if before == nil || after == nil {
		return
	}

	changed := diffSettings(before, after, beforeAuthSourceDefaults, afterAuthSourceDefaults, req)
	if len(changed) == 0 {
		return
	}

	subject, _ := middleware.GetAuthSubjectFromContext(c)
	role, _ := middleware.GetUserRoleFromContext(c)
	slog.Info("settings updated",
		"audit", true,
		"user_id", subject.UserID,
		"role", role,
		"changed", changed,
	)
}

func diffSettings(before *service.SystemSettings, after *service.SystemSettings, beforeAuthSourceDefaults *service.AuthSourceDefaultSettings, afterAuthSourceDefaults *service.AuthSourceDefaultSettings, req UpdateSettingsRequest) []string {
	changed := make([]string, 0, 20)
	if before.RegistrationEnabled != after.RegistrationEnabled {
		changed = append(changed, "registration_enabled")
	}
	if before.EmailVerifyEnabled != after.EmailVerifyEnabled {
		changed = append(changed, "email_verify_enabled")
	}
	if !equalStringSlice(before.RegistrationEmailSuffixWhitelist, after.RegistrationEmailSuffixWhitelist) {
		changed = append(changed, "registration_email_suffix_whitelist")
	}
	if before.PromoCodeEnabled != after.PromoCodeEnabled {
		changed = append(changed, "promo_code_enabled")
	}
	if before.InvitationCodeEnabled != after.InvitationCodeEnabled {
		changed = append(changed, "invitation_code_enabled")
	}
	if before.PasswordResetEnabled != after.PasswordResetEnabled {
		changed = append(changed, "password_reset_enabled")
	}
	if before.FrontendURL != after.FrontendURL {
		changed = append(changed, "frontend_url")
	}
	if before.TotpEnabled != after.TotpEnabled {
		changed = append(changed, "totp_enabled")
	}
	if before.LoginAgreementEnabled != after.LoginAgreementEnabled {
		changed = append(changed, "login_agreement_enabled")
	}
	if before.LoginAgreementMode != after.LoginAgreementMode {
		changed = append(changed, "login_agreement_mode")
	}
	if before.LoginAgreementUpdatedAt != after.LoginAgreementUpdatedAt {
		changed = append(changed, "login_agreement_updated_at")
	}
	if !equalLoginAgreementDocuments(before.LoginAgreementDocuments, after.LoginAgreementDocuments) {
		changed = append(changed, "login_agreement_documents")
	}
	if before.SMTPHost != after.SMTPHost {
		changed = append(changed, "smtp_host")
	}
	if before.SMTPPort != after.SMTPPort {
		changed = append(changed, "smtp_port")
	}
	if before.SMTPUsername != after.SMTPUsername {
		changed = append(changed, "smtp_username")
	}
	if req.SMTPPassword != "" {
		changed = append(changed, "smtp_password")
	}
	if before.SMTPFrom != after.SMTPFrom {
		changed = append(changed, "smtp_from_email")
	}
	if before.SMTPFromName != after.SMTPFromName {
		changed = append(changed, "smtp_from_name")
	}
	if before.SMTPUseTLS != after.SMTPUseTLS {
		changed = append(changed, "smtp_use_tls")
	}
	if before.TurnstileEnabled != after.TurnstileEnabled {
		changed = append(changed, "turnstile_enabled")
	}
	if before.TurnstileSiteKey != after.TurnstileSiteKey {
		changed = append(changed, "turnstile_site_key")
	}
	if req.TurnstileSecretKey != "" {
		changed = append(changed, "turnstile_secret_key")
	}
	if before.APIKeyACLTrustForwardedIP != after.APIKeyACLTrustForwardedIP {
		changed = append(changed, "api_key_acl_trust_forwarded_ip")
	}
	if before.LinuxDoConnectEnabled != after.LinuxDoConnectEnabled {
		changed = append(changed, "linuxdo_connect_enabled")
	}
	if before.LinuxDoConnectClientID != after.LinuxDoConnectClientID {
		changed = append(changed, "linuxdo_connect_client_id")
	}
	if req.LinuxDoConnectClientSecret != "" {
		changed = append(changed, "linuxdo_connect_client_secret")
	}
	if before.LinuxDoConnectRedirectURL != after.LinuxDoConnectRedirectURL {
		changed = append(changed, "linuxdo_connect_redirect_url")
	}
	if before.DingTalkConnectEnabled != after.DingTalkConnectEnabled {
		changed = append(changed, "dingtalk_connect_enabled")
	}
	if before.DingTalkConnectClientID != after.DingTalkConnectClientID {
		changed = append(changed, "dingtalk_connect_client_id")
	}
	if req.DingTalkConnectClientSecret != "" {
		changed = append(changed, "dingtalk_connect_client_secret")
	}
	if before.DingTalkConnectRedirectURL != after.DingTalkConnectRedirectURL {
		changed = append(changed, "dingtalk_connect_redirect_url")
	}
	if before.DingTalkConnectCorpRestrictionPolicy != after.DingTalkConnectCorpRestrictionPolicy {
		changed = append(changed, "dingtalk_connect_corp_restriction_policy")
	}
	if before.DingTalkConnectInternalCorpID != after.DingTalkConnectInternalCorpID {
		changed = append(changed, "dingtalk_connect_internal_corp_id")
	}
	if before.DingTalkConnectBypassRegistration != after.DingTalkConnectBypassRegistration {
		changed = append(changed, "dingtalk_connect_bypass_registration")
	}
	if before.DingTalkConnectSyncCorpEmail != after.DingTalkConnectSyncCorpEmail {
		changed = append(changed, "dingtalk_connect_sync_corp_email")
	}
	if before.DingTalkConnectSyncDisplayName != after.DingTalkConnectSyncDisplayName {
		changed = append(changed, "dingtalk_connect_sync_display_name")
	}
	if before.DingTalkConnectSyncDept != after.DingTalkConnectSyncDept {
		changed = append(changed, "dingtalk_connect_sync_dept")
	}
	if before.DingTalkConnectSyncCorpEmailAttrKey != after.DingTalkConnectSyncCorpEmailAttrKey {
		changed = append(changed, "dingtalk_connect_sync_corp_email_attr_key")
	}
	if before.DingTalkConnectSyncDisplayNameAttrKey != after.DingTalkConnectSyncDisplayNameAttrKey {
		changed = append(changed, "dingtalk_connect_sync_display_name_attr_key")
	}
	if before.DingTalkConnectSyncDeptAttrKey != after.DingTalkConnectSyncDeptAttrKey {
		changed = append(changed, "dingtalk_connect_sync_dept_attr_key")
	}
	if before.WeChatConnectEnabled != after.WeChatConnectEnabled {
		changed = append(changed, "wechat_connect_enabled")
	}
	if before.WeChatConnectAppID != after.WeChatConnectAppID {
		changed = append(changed, "wechat_connect_app_id")
	}
	if req.WeChatConnectAppSecret != "" {
		changed = append(changed, "wechat_connect_app_secret")
	}
	if before.WeChatConnectOpenAppID != after.WeChatConnectOpenAppID {
		changed = append(changed, "wechat_connect_open_app_id")
	}
	if req.WeChatConnectOpenAppSecret != "" {
		changed = append(changed, "wechat_connect_open_app_secret")
	}
	if before.WeChatConnectMPAppID != after.WeChatConnectMPAppID {
		changed = append(changed, "wechat_connect_mp_app_id")
	}
	if req.WeChatConnectMPAppSecret != "" {
		changed = append(changed, "wechat_connect_mp_app_secret")
	}
	if before.WeChatConnectMobileAppID != after.WeChatConnectMobileAppID {
		changed = append(changed, "wechat_connect_mobile_app_id")
	}
	if req.WeChatConnectMobileAppSecret != "" {
		changed = append(changed, "wechat_connect_mobile_app_secret")
	}
	if before.WeChatConnectOpenEnabled != after.WeChatConnectOpenEnabled {
		changed = append(changed, "wechat_connect_open_enabled")
	}
	if before.WeChatConnectMPEnabled != after.WeChatConnectMPEnabled {
		changed = append(changed, "wechat_connect_mp_enabled")
	}
	if before.WeChatConnectMobileEnabled != after.WeChatConnectMobileEnabled {
		changed = append(changed, "wechat_connect_mobile_enabled")
	}
	if before.WeChatConnectMode != after.WeChatConnectMode {
		changed = append(changed, "wechat_connect_mode")
	}
	if before.WeChatConnectScopes != after.WeChatConnectScopes {
		changed = append(changed, "wechat_connect_scopes")
	}
	if before.WeChatConnectRedirectURL != after.WeChatConnectRedirectURL {
		changed = append(changed, "wechat_connect_redirect_url")
	}
	if before.WeChatConnectFrontendRedirectURL != after.WeChatConnectFrontendRedirectURL {
		changed = append(changed, "wechat_connect_frontend_redirect_url")
	}
	if before.OIDCConnectEnabled != after.OIDCConnectEnabled {
		changed = append(changed, "oidc_connect_enabled")
	}
	if before.OIDCConnectProviderName != after.OIDCConnectProviderName {
		changed = append(changed, "oidc_connect_provider_name")
	}
	if before.OIDCConnectClientID != after.OIDCConnectClientID {
		changed = append(changed, "oidc_connect_client_id")
	}
	if req.OIDCConnectClientSecret != "" {
		changed = append(changed, "oidc_connect_client_secret")
	}
	if before.OIDCConnectIssuerURL != after.OIDCConnectIssuerURL {
		changed = append(changed, "oidc_connect_issuer_url")
	}
	if before.OIDCConnectDiscoveryURL != after.OIDCConnectDiscoveryURL {
		changed = append(changed, "oidc_connect_discovery_url")
	}
	if before.OIDCConnectAuthorizeURL != after.OIDCConnectAuthorizeURL {
		changed = append(changed, "oidc_connect_authorize_url")
	}
	if before.OIDCConnectTokenURL != after.OIDCConnectTokenURL {
		changed = append(changed, "oidc_connect_token_url")
	}
	if before.OIDCConnectUserInfoURL != after.OIDCConnectUserInfoURL {
		changed = append(changed, "oidc_connect_userinfo_url")
	}
	if before.OIDCConnectJWKSURL != after.OIDCConnectJWKSURL {
		changed = append(changed, "oidc_connect_jwks_url")
	}
	if before.OIDCConnectScopes != after.OIDCConnectScopes {
		changed = append(changed, "oidc_connect_scopes")
	}
	if before.OIDCConnectRedirectURL != after.OIDCConnectRedirectURL {
		changed = append(changed, "oidc_connect_redirect_url")
	}
	if before.OIDCConnectFrontendRedirectURL != after.OIDCConnectFrontendRedirectURL {
		changed = append(changed, "oidc_connect_frontend_redirect_url")
	}
	if before.OIDCConnectTokenAuthMethod != after.OIDCConnectTokenAuthMethod {
		changed = append(changed, "oidc_connect_token_auth_method")
	}
	if before.OIDCConnectUsePKCE != after.OIDCConnectUsePKCE {
		changed = append(changed, "oidc_connect_use_pkce")
	}
	if before.OIDCConnectValidateIDToken != after.OIDCConnectValidateIDToken {
		changed = append(changed, "oidc_connect_validate_id_token")
	}
	if before.OIDCConnectAllowedSigningAlgs != after.OIDCConnectAllowedSigningAlgs {
		changed = append(changed, "oidc_connect_allowed_signing_algs")
	}
	if before.OIDCConnectClockSkewSeconds != after.OIDCConnectClockSkewSeconds {
		changed = append(changed, "oidc_connect_clock_skew_seconds")
	}
	if before.OIDCConnectRequireEmailVerified != after.OIDCConnectRequireEmailVerified {
		changed = append(changed, "oidc_connect_require_email_verified")
	}
	if before.OIDCConnectUserInfoEmailPath != after.OIDCConnectUserInfoEmailPath {
		changed = append(changed, "oidc_connect_userinfo_email_path")
	}
	if before.OIDCConnectUserInfoIDPath != after.OIDCConnectUserInfoIDPath {
		changed = append(changed, "oidc_connect_userinfo_id_path")
	}
	if before.OIDCConnectUserInfoUsernamePath != after.OIDCConnectUserInfoUsernamePath {
		changed = append(changed, "oidc_connect_userinfo_username_path")
	}
	if before.SiteName != after.SiteName {
		changed = append(changed, "site_name")
	}
	if before.SiteLogo != after.SiteLogo {
		changed = append(changed, "site_logo")
	}
	if before.SiteSubtitle != after.SiteSubtitle {
		changed = append(changed, "site_subtitle")
	}
	if before.APIBaseURL != after.APIBaseURL {
		changed = append(changed, "api_base_url")
	}
	if before.ContactInfo != after.ContactInfo {
		changed = append(changed, "contact_info")
	}
	if before.DocURL != after.DocURL {
		changed = append(changed, "doc_url")
	}
	if before.HomeContent != after.HomeContent {
		changed = append(changed, "home_content")
	}
	if before.HideCcsImportButton != after.HideCcsImportButton {
		changed = append(changed, "hide_ccs_import_button")
	}
	if before.DefaultConcurrency != after.DefaultConcurrency {
		changed = append(changed, "default_concurrency")
	}
	if before.DefaultBalance != after.DefaultBalance {
		changed = append(changed, "default_balance")
	}
	if before.AffiliateRebateRate != after.AffiliateRebateRate {
		changed = append(changed, "affiliate_rebate_rate")
	}
	if before.AffiliateRebateFreezeHours != after.AffiliateRebateFreezeHours {
		changed = append(changed, "affiliate_rebate_freeze_hours")
	}
	if before.AffiliateRebateDurationDays != after.AffiliateRebateDurationDays {
		changed = append(changed, "affiliate_rebate_duration_days")
	}
	if before.AffiliateRebatePerInviteeCap != after.AffiliateRebatePerInviteeCap {
		changed = append(changed, "affiliate_rebate_per_invitee_cap")
	}
	if !equalDefaultSubscriptions(before.DefaultSubscriptions, after.DefaultSubscriptions) {
		changed = append(changed, "default_subscriptions")
	}
	if before.EnableModelFallback != after.EnableModelFallback {
		changed = append(changed, "enable_model_fallback")
	}
	if before.FallbackModelAnthropic != after.FallbackModelAnthropic {
		changed = append(changed, "fallback_model_anthropic")
	}
	if before.FallbackModelOpenAI != after.FallbackModelOpenAI {
		changed = append(changed, "fallback_model_openai")
	}
	if before.FallbackModelGemini != after.FallbackModelGemini {
		changed = append(changed, "fallback_model_gemini")
	}
	if before.FallbackModelAntigravity != after.FallbackModelAntigravity {
		changed = append(changed, "fallback_model_antigravity")
	}
	if before.EnableIdentityPatch != after.EnableIdentityPatch {
		changed = append(changed, "enable_identity_patch")
	}
	if before.IdentityPatchPrompt != after.IdentityPatchPrompt {
		changed = append(changed, "identity_patch_prompt")
	}
	if before.OpsMonitoringEnabled != after.OpsMonitoringEnabled {
		changed = append(changed, "ops_monitoring_enabled")
	}
	if before.OpsRealtimeMonitoringEnabled != after.OpsRealtimeMonitoringEnabled {
		changed = append(changed, "ops_realtime_monitoring_enabled")
	}
	if before.OpsQueryModeDefault != after.OpsQueryModeDefault {
		changed = append(changed, "ops_query_mode_default")
	}
	if before.OpsMetricsIntervalSeconds != after.OpsMetricsIntervalSeconds {
		changed = append(changed, "ops_metrics_interval_seconds")
	}
	if before.MinClaudeCodeVersion != after.MinClaudeCodeVersion {
		changed = append(changed, "min_claude_code_version")
	}
	if before.MaxClaudeCodeVersion != after.MaxClaudeCodeVersion {
		changed = append(changed, "max_claude_code_version")
	}
	if before.AllowUngroupedKeyScheduling != after.AllowUngroupedKeyScheduling {
		changed = append(changed, "allow_ungrouped_key_scheduling")
	}
	if before.BackendModeEnabled != after.BackendModeEnabled {
		changed = append(changed, "backend_mode_enabled")
	}
	if before.PurchaseSubscriptionEnabled != after.PurchaseSubscriptionEnabled {
		changed = append(changed, "purchase_subscription_enabled")
	}
	if before.PurchaseSubscriptionURL != after.PurchaseSubscriptionURL {
		changed = append(changed, "purchase_subscription_url")
	}
	if before.TableDefaultPageSize != after.TableDefaultPageSize {
		changed = append(changed, "table_default_page_size")
	}
	if !equalIntSlice(before.TablePageSizeOptions, after.TablePageSizeOptions) {
		changed = append(changed, "table_page_size_options")
	}
	if before.CustomMenuItems != after.CustomMenuItems {
		changed = append(changed, "custom_menu_items")
	}
	if before.CustomEndpoints != after.CustomEndpoints {
		changed = append(changed, "custom_endpoints")
	}
	if before.EnableFingerprintUnification != after.EnableFingerprintUnification {
		changed = append(changed, "enable_fingerprint_unification")
	}
	if before.EnableMetadataPassthrough != after.EnableMetadataPassthrough {
		changed = append(changed, "enable_metadata_passthrough")
	}
	if before.EnableCCHSigning != after.EnableCCHSigning {
		changed = append(changed, "enable_cch_signing")
	}
	if before.EnableClaudeOAuthSystemPromptInjection != after.EnableClaudeOAuthSystemPromptInjection {
		changed = append(changed, "enable_claude_oauth_system_prompt_injection")
	}
	if before.ClaudeOAuthSystemPrompt != after.ClaudeOAuthSystemPrompt {
		changed = append(changed, "claude_oauth_system_prompt")
	}
	if before.ClaudeOAuthSystemPromptBlocks != after.ClaudeOAuthSystemPromptBlocks {
		changed = append(changed, "claude_oauth_system_prompt_blocks")
	}
	if before.EnableAnthropicCacheTTL1hInjection != after.EnableAnthropicCacheTTL1hInjection {
		changed = append(changed, "enable_anthropic_cache_ttl_1h_injection")
	}
	if before.RewriteMessageCacheControl != after.RewriteMessageCacheControl {
		changed = append(changed, "rewrite_message_cache_control")
	}
	if before.AntigravityUserAgentVersion != after.AntigravityUserAgentVersion {
		changed = append(changed, "antigravity_user_agent_version")
	}
	if before.OpenAICodexUserAgent != after.OpenAICodexUserAgent {
		changed = append(changed, "openai_codex_user_agent")
	}
	if before.OpenAIAllowClaudeCodeCodexPlugin != after.OpenAIAllowClaudeCodeCodexPlugin {
		changed = append(changed, "openai_allow_claude_code_codex_plugin")
	}
	if before.PaymentVisibleMethodAlipaySource != after.PaymentVisibleMethodAlipaySource {
		changed = append(changed, "payment_visible_method_alipay_source")
	}
	if before.PaymentVisibleMethodWxpaySource != after.PaymentVisibleMethodWxpaySource {
		changed = append(changed, "payment_visible_method_wxpay_source")
	}
	if before.PaymentVisibleMethodAlipayEnabled != after.PaymentVisibleMethodAlipayEnabled {
		changed = append(changed, "payment_visible_method_alipay_enabled")
	}
	if before.PaymentVisibleMethodWxpayEnabled != after.PaymentVisibleMethodWxpayEnabled {
		changed = append(changed, "payment_visible_method_wxpay_enabled")
	}
	if before.OpenAIAdvancedSchedulerEnabled != after.OpenAIAdvancedSchedulerEnabled {
		changed = append(changed, "openai_advanced_scheduler_enabled")
	}
	// 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｅΟ娆惧殭缂佺姴鐏氶妵鍕疀閹炬惌妫炵紓浣界堪閸婃洝鐏冮梺鎸庣箓閹冲酣寮抽悙鐑樼厱閻庯絽澧庣弧鈧梺鍝勮嫰缁夌懓鐣烽锕€绀嬫い鎺嗗亾濞寸姵妞藉铏圭磼濡櫣浼囧┑鈽嗗亜鐎氼喗绔熼弴鐔洪檮闁告稑锕﹂崢浠嬫⒑瑜版帒浜伴柛鎾寸懃铻炴慨妞诲亾闁哄本娲樺鍕醇濠靛棗顥夐柣鐔哥矌婢ф鏁Δ鍛柧闁哄被鍎查悡鏇㈡煏婢跺牆鐏繛鍛閺岀喖宕ｆ径娑溾偓鎸庢叏婵犲啯銇濈€规洏鍔嶇换婵嬪川椤撗勵棨闂傚倷鑳堕…鍫ユ晝閵堝鍊舵繝闈涱儜缂嶆牠鐓崶銊︹拻缂佲檧鍋撴繝娈垮枟閿曗晠宕滃☉鈶哄洭濡搁埡鍌楁嫼闂傚倸鐗婃笟妤呮偂椤撶姷纾奸柣妯哄暱閻忋儵鏌嶇拠鑼ч柡浣瑰姈瀵板嫮鈧急鍕伖缂傚倸鍊风欢锟犲垂閸楃伝鍝勎熺悰鈩冪€洪梺鍛婄☉閻°劑鎮￠悢鎼炰簻闁规崘娉涙禒婊堟煃瑜滈崜姘跺箖閸岀偑鈧礁顫濋懜鍨珳婵犮垼娉涢鍛濞差亝鈷戦柛娑橈功閳藉鏌ㄩ弴妯衡偓鏍矉瀹ュ拋鐓ラ柛顐ゅ暱閹锋椽姊洪崨濠勨槈闁挎洏鍎虫禍鎼侇敃閿旂晫鍘鹃梺鍛婂姌鐏忔瑩宕洪敐鍡╂闁绘劕妯婂Ο鈧悗瑙勬礈椤牐鐏冩繛杈剧到閹碱偊鐛Δ鍛拻濞达綀娅ｇ敮娑㈡煥濮樿埖鐓曢柣鏃傤焾椤曟粓鏌熸笟鍨妞わ箑顕槐鎺撴綇閵娿儲璇為梺璇″枓閺呯姴鐣峰鈧幊鐘活敄鐠恒劎顦ㄩ梻鍌氬€风欢姘焽瑜嶇叅闁挎洖鍊哥粈澶嬬箾閸℃绠抽柡鍡缁辨帞鈧綆鍙庨崵锕傛煛閸愶絽浜鹃梺鐟板槻閹虫ê鐣烽敐鍡楃窞濠电姴瀚闂傚倸鍊搁崐鎼佸磹妞嬪孩顐芥慨妯挎硾閻掑灚銇勯幒宥堝厡闁宠棄顦遍埀顒冾潐閹哥螞濠靛钃熸繛鎴炲焹閸嬫捇鏁愭惔婵囧€ｉ梻濠庡墻閸撴盯鍩€椤掍緡鍟忛柛鐘崇☉閳绘柨鈽夊顒€鐤惧┑锛勫亼閸婃洜鎹㈤幇鏉跨疇濠㈣埖鍔曠粻顖炴煕濠靛棗顏存繛鎾愁煼閺屾洟宕煎┑鍥舵婵犳鍟崨顖滐紲闂佸搫琚崕鎶芥偩濞差亝鐓欐い鏍ㄨ壘閺嗭絾顨ラ悙宸剶闁轰礁鍊块幃鍓т沪閽樺顔?
	if before.BalanceLowNotifyEnabled != after.BalanceLowNotifyEnabled {
		changed = append(changed, "balance_low_notify_enabled")
	}
	if before.BalanceLowNotifyThreshold != after.BalanceLowNotifyThreshold {
		changed = append(changed, "balance_low_notify_threshold")
	}
	if before.BalanceLowNotifyRechargeURL != after.BalanceLowNotifyRechargeURL {
		changed = append(changed, "balance_low_notify_recharge_url")
	}
	if before.SubscriptionExpiryNotifyEnabled != after.SubscriptionExpiryNotifyEnabled {
		changed = append(changed, "subscription_expiry_notify_enabled")
	}
	if before.AccountQuotaNotifyEnabled != after.AccountQuotaNotifyEnabled {
		changed = append(changed, "account_quota_notify_enabled")
	}
	if !equalNotifyEmailEntries(before.AccountQuotaNotifyEmails, after.AccountQuotaNotifyEmails) {
		changed = append(changed, "account_quota_notify_emails")
	}
	if before.ChannelMonitorEnabled != after.ChannelMonitorEnabled {
		changed = append(changed, "channel_monitor_enabled")
	}
	if before.ChannelMonitorDefaultIntervalSeconds != after.ChannelMonitorDefaultIntervalSeconds {
		changed = append(changed, "channel_monitor_default_interval_seconds")
	}
	if before.AvailableChannelsEnabled != after.AvailableChannelsEnabled {
		changed = append(changed, "available_channels_enabled")
	}
	if before.AffiliateEnabled != after.AffiliateEnabled {
		changed = append(changed, "affiliate_enabled")
	}
	if before.RiskControlEnabled != after.RiskControlEnabled {
		changed = append(changed, "risk_control_enabled")
	}
	if before.CyberSessionBlockEnabled != after.CyberSessionBlockEnabled {
		changed = append(changed, "cyber_session_block_enabled")
	}
	if before.CyberSessionBlockTTLSeconds != after.CyberSessionBlockTTLSeconds {
		changed = append(changed, "cyber_session_block_ttl_seconds")
	}
	// Default platform quotas闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢琛″亾濞戞瑯鐒界紒鐘卞嵆閺屻倝鎮ч崼婵愬殝缂備胶绮换鍫ュ箖娴犲鍋ㄩ柛顭戝亞缁夌瘶 map闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍛婄秶闁告挆鍛闂佽瀛╅懝楣兯囨导鏉懳﹂柛鏇ㄥ灠缁犳娊鏌熺€涙绠ュù鐘靛帶閳规垿顢欏▎鐐秷闂佺粯鎼换婵嗩嚕婵犳艾惟闁靛鍨洪弬鈧梻浣虹《閸撴繈鏁冮埡浣勬盯鍩勯崘顏嗙槇濠电偛鐗嗛悘婵嬫倶閿熺姵鐓欓柛娑橈工閳绘洜鈧鍠栭…鐑藉箖閵忋倖鍋傞幖杈剧悼閺嗕即鏌ｉ悢鍝ョ煁缂侇喗鎸搁悾鐑芥偨閸涘﹤浜滈梻鍌楀亾闁归偊鍠氶悾楣冩⒒娴ｈ棄浜归柍宄扮墦瀹曟粌鈻庨幘鍐插亶闁诲海鏁哥涵鍫曞磻閹炬枼鏋旈柛顭戝枟閻忓牓姊虹紒妯煎⒈闁告鍥ㄥ仼闁绘垼濮ら崑鍌炲箹鐎涙〞鎴﹀棘閳ь剟姊绘担铏广€婇柛鎾寸箞閳ワ箓宕堕妸褜娴勯悗鍏夊亾闁逞屽墰濡?
	if !equalPlatformQuotaSettings(before.DefaultPlatformQuotas, after.DefaultPlatformQuotas) {
		changed = append(changed, service.SettingKeyDefaultPlatformQuotas)
	}
	changed = appendAuthSourceDefaultChanges(changed, beforeAuthSourceDefaults, afterAuthSourceDefaults)
	return changed
}

func appendAuthSourceDefaultChanges(changed []string, before *service.AuthSourceDefaultSettings, after *service.AuthSourceDefaultSettings) []string {
	if before == nil {
		before = &service.AuthSourceDefaultSettings{}
	}
	if after == nil {
		after = &service.AuthSourceDefaultSettings{}
	}

	type providerDefaultGrantField struct {
		name   string
		before service.ProviderDefaultGrantSettings
		after  service.ProviderDefaultGrantSettings
	}

	fields := []providerDefaultGrantField{
		{name: "email", before: before.Email, after: after.Email},
		{name: "linuxdo", before: before.LinuxDo, after: after.LinuxDo},
		{name: "oidc", before: before.OIDC, after: after.OIDC},
		{name: "wechat", before: before.WeChat, after: after.WeChat},
		{name: "github", before: before.GitHub, after: after.GitHub},
		{name: "google", before: before.Google, after: after.Google},
		{name: "dingtalk", before: before.DingTalk, after: after.DingTalk},
	}
	for _, field := range fields {
		if field.before.Balance != field.after.Balance {
			changed = append(changed, "auth_source_default_"+field.name+"_balance")
		}
		if field.before.Concurrency != field.after.Concurrency {
			changed = append(changed, "auth_source_default_"+field.name+"_concurrency")
		}
		if !equalDefaultSubscriptions(field.before.Subscriptions, field.after.Subscriptions) {
			changed = append(changed, "auth_source_default_"+field.name+"_subscriptions")
		}
		if field.before.GrantOnSignup != field.after.GrantOnSignup {
			changed = append(changed, "auth_source_default_"+field.name+"_grant_on_signup")
		}
		if field.before.GrantOnFirstBind != field.after.GrantOnFirstBind {
			changed = append(changed, "auth_source_default_"+field.name+"_grant_on_first_bind")
		}
		// Platform quotas diff闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍏煎€绘俊顖濆吹閸欏棝姊洪崫鍕靛剰闁绘绻橀崺鈧い鎺嗗亾缂佺姴绉瑰畷鏇㈡焼瀹撱儱娲、娑樞掗崹锕€娲ら拑鐔兼煏婢舵稑顩柛娆忔閳规垿鎮╃紒妯婚敪濠碘槅鍋呴悷鈺勬＂闂佺粯鍔栫粊鎾绩閼恒儯浜滈柡宥冨妿閳洟鎮樿箛銉х暤闁哄备鈧磭鏆嗛柍褜鍓熷畷浼村冀椤撶姴绁﹂梺绯曞墲閸戠懓危妤ｅ啯鈷戦柛锔诲幗濞呮洖鈹戦鈧ˉ鎾澄ｉ幇鏉跨婵°倐鍋撶痪鎯у悑閹便劌螣婢剁鎯堥梺鍛婃閸ㄨ泛顫忓ú顏勪紶闁告洦鍘炬导鍥⒑閸濄儱校闁绘濞€楠炲啯瀵奸幖顓熸櫔闂侀€炲苯澧柣锝囨焿閵囨劙骞掑┑鍥ㄦ珦闂備礁鎲￠悷锝夊磿瀹曞洨鐭嗛柛鎰╁妷閺€浠嬫煥濞戞ê顏╁ù婊冦偢閺屾稒绻濋崘顏勨吂闁绘挶鍊濋悡顐﹀炊閵娧€妲堥梺鎼炲妼閸婂潡寮昏缁犳盯骞樼壕瀣攨闂備焦妞块崢浠嬫偡閵夆晛鐓橀柟杈鹃檮閸嬫劙鏌熺紒妯虹瑲婵炲牜鍘剧槐鎾存媴缁嬫鏆㈤梺绋款儑閸嬬喖骞戦姀鐘闁靛繒濮烽娲⒑缂佹ê濮囨い鏇ㄥ幖閳绘捇濡堕崪浣瑰瘜闂侀潧鐗嗙换鎰版儊濠婂牊鐓曢柨婵嗙箳閸掔増銇勯銏㈢閻撱倖銇勮箛鎾愁仼缂佹劖绋掔换婵嬫偨闂堟刀銏ゆ倵濮樼厧寮€规洘濞婇、姗€濮€閳锯偓閹锋椽姊虹涵鍛汗闁稿绋掗幈銊╁磼濞戞氨顔曟繛杈剧到閸熷灝煤閿曞倸纾婚柍鍝勬噺閻撶喖鏌熸潏鍓у埌鐞氭岸姊虹粙娆惧剰閻庢矮鍗冲璇测槈濡攱鏂€闂佸憡娲﹂崑鍕叏婢舵劖鈷?JSON key闂?
		if !equalPlatformQuotaSettings(field.before.PlatformQuotas, field.after.PlatformQuotas) {
			changed = append(changed, service.SettingKeyAuthSourcePlatformQuotas(field.name))
		}
	}
	if before.ForceEmailOnThirdPartySignup != after.ForceEmailOnThirdPartySignup {
		changed = append(changed, "force_email_on_third_party_signup")
	}
	return changed
}

func normalizeDefaultSubscriptions(input []dto.DefaultSubscriptionSetting) []dto.DefaultSubscriptionSetting {
	if len(input) == 0 {
		return nil
	}
	normalized := make([]dto.DefaultSubscriptionSetting, 0, len(input))
	for _, item := range input {
		if item.GroupID <= 0 || item.ValidityDays <= 0 {
			continue
		}
		if item.ValidityDays > service.MaxValidityDays {
			item.ValidityDays = service.MaxValidityDays
		}
		normalized = append(normalized, item)
	}
	return normalized
}

func normalizeOptionalDefaultSubscriptions(input *[]dto.DefaultSubscriptionSetting) *[]dto.DefaultSubscriptionSetting {
	if input == nil {
		return nil
	}
	normalized := normalizeDefaultSubscriptions(*input)
	return &normalized
}

func float64ValueOrDefault(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}

func intValueOrDefault(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func boolValueOrDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func defaultSubscriptionsValueOrDefault(input *[]dto.DefaultSubscriptionSetting, fallback []service.DefaultSubscriptionSetting) []service.DefaultSubscriptionSetting {
	if input == nil {
		return fallback
	}
	result := make([]service.DefaultSubscriptionSetting, 0, len(*input))
	for _, item := range *input {
		result = append(result, service.DefaultSubscriptionSetting{
			GroupID:      item.GroupID,
			ValidityDays: item.ValidityDays,
		})
	}
	return result
}

// platformQuotasValueOrDefault 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柛娑橈攻閸欏繐霉閸忓吋缍戦柛銊ュ€块弻锝夊箻瀹曞洤鍝洪梺鍝勵儐閻楁鎹㈠☉銏犵闁绘劘灏欓崝浼存⒑缁嬫鍎愰柟鍛婃倐閿濈偛鈹戠€ｎ偄浜楅柟鍏肩暘閸ㄦ槒銇愭惔顫箚?auth-source platform quota 闂?nil 闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊瑜忛弳锕傛煕椤垵浜濋柛娆忕箻閺屸剝寰勭€ｎ亝顔呴悷婊呭鐢帞澹曟總鍛婄厽闁归偊鍘肩徊濠氭煕鐎ｎ亞效婵﹥妞介獮鍡氼槾缂佺姷澧楃换婵嬪焵椤掍焦缍囬柕濞р偓閺嬫牠鎮楅獮鍨姎妞わ缚鍗抽幃锟犳偄閸忚偐鍘搁梺绋挎湰濮樸劍绂掗姀銈嗙厵妞ゆ棁妫勯悘鎾煛?// nil = 闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊瑜忛弳锕傛煕椤垵浜濋柛娆忕箻閺屸剝寰勭€ｎ亝顔呭┑鐐叉▕娴滃爼寮崶顒佺厪闊洦娲栧暩闂佸憡绻冮〃濠傤潖缂佹鐟归柛銉戝喚娼荤紓鍌欑椤﹂亶寮拠鑼殾闁哄洢鍨圭粻锝嗙節閵忥紕绠撶紓宥咃工閻ｅ嘲螖閸涱厾顦ч梺鍏肩ゴ閺呮繈鍩㈣箛娑欌拺閻犲洦褰冮銏ゆ煠瑜版帞鐣洪挊婵喢归崗鍏肩稇濡楀懘姊洪崨濠冨闁哥姵宀搁幃鍧楀焵椤掑嫭鈷戦柛锔诲幖娴滈箖鏌涢幘瀵割暡缂佸倹甯楃缓浠嬪川婵炵偓瀚煎┑鐐存綑閸氬鎮疯閹箖鏌嗗鍡欏幗闂佸搫璇炴担闀愬垝闁诲孩顔栭崰娑㈩敋瑜旈崺銉﹀緞婵犲孩寤洪梺绯曞墲閿氶柛蹇擄躬濮婄粯鎷呴搹鐟扮闂佸湱顭堥幗婊堟嚍鏉堛劎闄勭紒瀣硶椤︹晠姊洪懞銉冾亪藝閽樺鈧帗绻濆顓犲帾闂佸壊鍋呯换鍐夐悙鐢电＜闁逞屽墴瀹曟﹢鍩￠崒姘紟婵犵妲呴崹鎶解€﹂崼鐔稿弿闁稿瞼鍋為悡鏇㈡煃鐟欏嫬鍔ゅù婊呭亾娣囧﹪鎮欓鍕ㄥ亾閺嶎偅鏆滃┑鐘叉处閸ゅ嫰鏌涢锝嗙闁绘挴鍋撻梻浣告惈濞层劑宕伴崱妯碱洸婵犲﹤鐗婇悡蹇撯攽閻愯尙浠㈤柛鏃€绮嶉妵鍕Ψ閿濆懐浠鹃梺闈涙搐鐎氱増鎱ㄩ埀顒勬煥濞戞ê顏柛锛卞洦鈷戝ù鍏肩懅缁嬭崵绱掔拠鑼闁伙絿鍏橀幃銏㈠枈鏉堛劍娅屽┑鐐舵彧缁茶棄锕㈡潏顐犱汗闁靛ň鏅滈埛鎺楁煕鐏炲墽鎳呮い锔肩畱椤潡鎮风敮顔垮惈閻庤娲橀崹鍨暦閵娾晩鏁嶆慨姗嗗墰娴滀即姊虹涵鍛棈闁规椿浜炲濠囧箰鎼达絺鏋栭悗骞垮劚閹儻銇愰幒鎾充汗閻庣懓澹婇崰鏍礈闁秵鈷?fallback闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢灏佹瀻闁炽儱纾﹢鎾煕閺冣偓閸ㄥ灝鐣峰ú顏勭劦妞ゆ帊闄嶆禍婊堟煙閻戞ê鐏ラ柍褜鍓濆畷鍨珶閺囩喓闄勭紒瀣硶妤犲洭姊洪崜鎻掍簼缂佸鍨舵穱濠囨嚍閵夛絼绨婚梺闈涱檧婵″洭鎯堥崚?nil闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍛婄秶濡わ絽鍟宥夋⒑缁嬫鍎忔い鎴濐樀瀵鈽夐姀鐘靛姶闂佸憡鍔︽禍鏍ｉ崼銏㈢＝?empty map闂? 闂傚倸鍊搁崐鎼佸磹閹间礁纾圭€瑰嫭鍣磋ぐ鎺戠倞鐟滄粌霉閺嶎厽鐓忓┑鐐靛亾濞呭棝鏌涢妶鍛伃闁哄本鐩獮姗€鎳犻鈧俊浠嬫倵鐟欏嫭绀€鐎殿喖澧庨幑銏犫攽鐎ｎ偒妫冨┑鐐村灦閼归箖鍩呮潏鈺冪＝濞达絾褰冩禍鐐節閻㈤潧孝婵炶绠撳畷鎰版偨閸涘﹦鍙嗗┑鐘绘涧濡繈顢撳Δ鈧…鑳檨闁搞劌鐖煎璇测槈閵忊€充簻闂佸憡绋戦…鍫熺閵忋倖鈷戦柛婵嗗濠€浼存煟閳哄﹤鐏︾€规洘妞藉畷鐔碱敍濮橀硸妲版俊鐐€曠换鎰板箠鎼粹檧鏋嶆俊銈呭暟绾捐棄霉閿濆洦鍤€闁告柣鍊楅幉鎼佸箥椤旈棿妲愰悗瑙勬穿缁绘繈鐛惔銊﹀癄濠㈣泛瀛╅幉?// 濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴ｆ閺嬩線鏌熼梻瀵割槮缁惧墽绮换娑㈠箣閺冣偓閸ゅ秹鏌涢妷顔煎⒒闁轰礁娲弻鏇＄疀閺囩倫銉︺亜閿旇娅嶉柟顔筋殜閻涱噣宕归鐓庮潛婵＄偑鍊х€靛矂宕瑰畷鍥у灊闁割偁鍎遍柋鍥ㄧ節閵忥紕绠撶紓宥咃躬楠炲啴鍩勯崘鈺佸妳闂佹寧绻傚Λ娆撳窗閺嶎厽鈷掑ù锝呮啞閹牓鏌熷畡閭﹀剶鐎规洘绮岄～婵囨綇閵娿儱绨ユ繝鐢靛仦閸垶宕硅ぐ鎺斿祦闁靛骏绱曠粻楣冩煕閳╁厾顏堟倶閻斿吋鐓曢悗锝庡亞绾剧硵 null 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｉ幇顒佹儓闁搞劌鍊块弻娑㈩敃閿濆棛顦ョ紓浣哄С閸楁娊寮诲☉妯锋婵妫涢崣鏇炩攽閻愯尙澧旂紒顔界懇瀵鎮㈤崗灏栨嫽闁诲海鏁搁…鍫熶繆娴犲鈷戠紒瀣儥閸庡秹鏌涙繝鍐炬疁妤犵偛顦辩划娆忊枎閹勫€┑鐘灱濞夋盯鏁冮敃鍌氬偍婵炲樊浜濋埛鎴︽煟閻斿憡绶查柍閿嬫⒒缁辨帡顢氶崨顓犱桓閻庤娲橀崹鎸庝繆閹间礁鐓涢柛灞剧矊楠炲牊淇婇悙顏勨偓銈夊储娴犲鍨傞柛顭戝亞椤╂煡鏌ｅΔ鈧悧鍛崲閸℃稒鍋ｉ柛銉ユ搐閹虫劘鈪靛┑鐘愁問閸犳牠鏁冮敃鍌氱？闁靛牆娲犻崑鎾绘偨闂堟稑浠橀梺瀹狀嚙闁帮綁鐛€ｎ喗鍋愭い鏃堟？閻㈠姊婚崒姘偓鐑芥嚄閸洍鈧箓宕奸埗鈺佷壕婵炴垶甯楀▍鍡欑磼缂佹鈯曠紒缁樼箞瀹曟儼顦撮柣锝呮惈閳规垿鎮欓崣澶嗘灆婵炲瓨绮嶇划搴ｅ垝婵犳碍鏅插璺侯儑閸樹粙姊鸿ぐ鎺戜喊闁告瑥閰ｅ畷顖濈疀閺冨倻顔曢梺鍝勵槹閸ㄧ敻顢旈銏＄厸鐎光偓鐎ｎ剛袣缂備胶濮甸惄顖炵嵁濮椻偓瀹曟粍绗熼崶褎鍊涢梻鍌欐祰瀹曞灚鎱ㄩ幎鑺ュ€块柨鏇炲€归崑鍌炴煃閵夈儳锛嶉柡鍡畵閺岀喐娼忛崜褏鏆犵紓浣哄Т椤兘骞冭ぐ鎺戠倞鐟滃秶鐥閵囧嫰濡烽敂鍓х杽闂佸搫鐭夌紞渚€骞冨▎鎴斿亾閻㈢數銆婇柡瀣濮婃椽鏌呴悙鑼跺濠⒀屽櫍閺屾稑螣閼姐倕顫х紓浣规⒒閸犳牕顕ｉ幘顔碱潊闁抽敮鍋撻柟閿嬫そ濮婃椽宕ㄦ繝鍕暤闁诲孩鍑归崜鐔煎春閳ь剚銇勯幒鎴濃偓褰掑吹閳ь剟姊烘导娆戝埌闁搞垺鐓￠敐鐐差煥閸繄鍔﹀銈嗗笒鐎氼剟寮伴妷锔剧闁瑰瓨鐟ラ悘鈺呮⒑閸楃偞鍠橀柡灞炬礃瀵板嫬鈽夊鍡樺枠闂備礁鎲￠敃銏㈢不閺嵮屾綎婵炲樊浜滄导鐘绘煕閺囩偟浠涢弫鍫ユ⒑鐠囧弶鎹ｉ柣鐔稿▕瀹曟儼顦规俊顐ゅ厴閺岋綁鎮㈤崫銉﹀櫑闁诲孩鍑归崰姘卞垝缂佹ǜ鍋呴柛鎰ㄦ櫇閸樹粙姊洪崷顓炰壕濠电偛锕畷锝夊礋椤栨稓鍘卞┑鐐叉缁绘劙顢旈锔界厓鐟滄粓宕滃┑瀣剁稏濠㈣泛鈯曞ú顏勭厸闁告劑鍔庣粵蹇旂箾鐎电孝妞ゆ垶鍔欏顐﹀炊瑜夐弨浠嬫煕鐏炲墽鐭ら柣鎺戝⒔閳ь剛鎳撻幉锛勬崲閸愵喖桅闁告洦鍨伴崘鈧梺闈浤涢崨顖氬箻闂傚倷绀侀幖顐⑽涚€靛憡宕叉繝闈涙－濞兼牜绱撴担鑲℃垶鍒婇幘顔界厱婵炴垶锕銉╂煛閸℃澧﹂柡宀嬬秮婵偓闁宠桨鑳舵禒顓烆渻閵堝啫濡奸柨鏇ㄤ簼娣囧﹪骞橀鑲╊槰濡炪倖姊婚崢褏绮ｅ☉姗嗘富闁靛牆妫涙晶顒併亜閵娿儲鍣虹紒鍌氱У閵堬綁宕橀埞鐐闂備礁澹婇崑鍡涘窗閹捐鐓″璺虹灱绾惧ジ鏌ｅ▎鎰噧闁硅櫕鍔楁竟?nil map闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍛婄秶濡わ絽鍟宥夋⒑閹肩偛鈧牠宕濋弽顓炍﹂柛鏇ㄥ灠閸愨偓濡炪値鍘介崹闈涒枔瑜斿娲川婵犲嫮鐣甸柣搴㈡皑閸忔﹢骞冮敓鐘参ㄩ柍鍝勫€婚崢閬嶆⒑瑜版帒浜伴柛鐘崇墬缁傚秹鎮欓璺ㄧ畾濠碉紕鍋熼崢褎鎱ㄩ崒娑欏弿濠电姴鍟妵婵囦繆椤愩垹鏆欓柍钘夘槸椤繈顢楁径瀣槖婵犵绱曢崑鎴﹀磹瑜忛埀顒勬涧閻倸鐣烽崷顓熷磯闁惧繗顫夊▓楣冩⒑闂堟稓绠為柛濠冩礈閻ヮ亣顦归柡宀€鍠栭獮鍥ㄦ媴缁嬫娼婚梻鍌欑鎼存粎寰婃禒瀣仼闁绘垼妫勯柋鍥煛閸モ晛鈧綁鍩￠崘鈺佹瀾闂婎偄娲︾粙鎺楀磻閳轰讲鏀介柛灞剧閸熺偤鏌嶉柨瀣伌闁哄本绋戦埞鎴﹀幢濡ゅ﹣鐥俊鐐€戦崕鏌ュ垂閸ф钃熸繛鎴炲焹閸嬫捇鏁愭惔鈥茬凹閻庤娲栭張顒勫箞閵婏妇绡€闁告劏鏂傛禒銏ゆ倵鐟欏嫭澶勯柛瀣工閻ｇ兘鎮㈢喊杈ㄦ櫇闂佹寧绻傚Λ娆擃敊閹达附鈷掗柛灞剧懅椤︼箓鏌熺拠褏绡€鐎规洘绻嗛ˇ瀵糕偓鍨緲閿曨亪骞婇悙鍝勎ㄧ憸婊兠洪幖浣光拺闁告稑锕ョ€垫瑩鏌涘☉鍗炴珯缂佹墎鍓濇穱濠囨倷椤忓嫧鍋撻弽顓炵闁割偅娲栭弰銉╂煕閺囥劌鈧姴鈽夐姀鐘殿吅闂佺粯鍨靛Λ娆擃敊?
// platformQuotasValueOrDefault returns fallback when value is nil.
func platformQuotasValueOrDefault(value, fallback map[string]*service.DefaultPlatformQuotaSetting) map[string]*service.DefaultPlatformQuotaSetting {
	if value == nil {
		return fallback
	}
	return value
}

func systemSettingsResponseData(settings dto.SystemSettings, authSourceDefaults *service.AuthSourceDefaultSettings) map[string]any {
	data := make(map[string]any)
	raw, err := json.Marshal(settings)
	if err == nil {
		_ = json.Unmarshal(raw, &data)
	}
	if authSourceDefaults == nil {
		authSourceDefaults = &service.AuthSourceDefaultSettings{}
	}

	data["auth_source_default_email_balance"] = authSourceDefaults.Email.Balance
	data["auth_source_default_email_concurrency"] = authSourceDefaults.Email.Concurrency
	data["auth_source_default_email_subscriptions"] = authSourceDefaults.Email.Subscriptions
	data["auth_source_default_email_grant_on_signup"] = authSourceDefaults.Email.GrantOnSignup
	data["auth_source_default_email_grant_on_first_bind"] = authSourceDefaults.Email.GrantOnFirstBind
	data["auth_source_default_linuxdo_balance"] = authSourceDefaults.LinuxDo.Balance
	data["auth_source_default_linuxdo_concurrency"] = authSourceDefaults.LinuxDo.Concurrency
	data["auth_source_default_linuxdo_subscriptions"] = authSourceDefaults.LinuxDo.Subscriptions
	data["auth_source_default_linuxdo_grant_on_signup"] = authSourceDefaults.LinuxDo.GrantOnSignup
	data["auth_source_default_linuxdo_grant_on_first_bind"] = authSourceDefaults.LinuxDo.GrantOnFirstBind
	data["auth_source_default_dingtalk_balance"] = authSourceDefaults.DingTalk.Balance
	data["auth_source_default_dingtalk_concurrency"] = authSourceDefaults.DingTalk.Concurrency
	data["auth_source_default_dingtalk_subscriptions"] = authSourceDefaults.DingTalk.Subscriptions
	data["auth_source_default_dingtalk_grant_on_signup"] = authSourceDefaults.DingTalk.GrantOnSignup
	data["auth_source_default_dingtalk_grant_on_first_bind"] = authSourceDefaults.DingTalk.GrantOnFirstBind
	data["auth_source_default_oidc_balance"] = authSourceDefaults.OIDC.Balance
	data["auth_source_default_oidc_concurrency"] = authSourceDefaults.OIDC.Concurrency
	data["auth_source_default_oidc_subscriptions"] = authSourceDefaults.OIDC.Subscriptions
	data["auth_source_default_oidc_grant_on_signup"] = authSourceDefaults.OIDC.GrantOnSignup
	data["auth_source_default_oidc_grant_on_first_bind"] = authSourceDefaults.OIDC.GrantOnFirstBind
	data["auth_source_default_wechat_balance"] = authSourceDefaults.WeChat.Balance
	data["auth_source_default_wechat_concurrency"] = authSourceDefaults.WeChat.Concurrency
	data["auth_source_default_wechat_subscriptions"] = authSourceDefaults.WeChat.Subscriptions
	data["auth_source_default_wechat_grant_on_signup"] = authSourceDefaults.WeChat.GrantOnSignup
	data["auth_source_default_wechat_grant_on_first_bind"] = authSourceDefaults.WeChat.GrantOnFirstBind
	data["auth_source_default_github_balance"] = authSourceDefaults.GitHub.Balance
	data["auth_source_default_github_concurrency"] = authSourceDefaults.GitHub.Concurrency
	data["auth_source_default_github_subscriptions"] = authSourceDefaults.GitHub.Subscriptions
	data["auth_source_default_github_grant_on_signup"] = authSourceDefaults.GitHub.GrantOnSignup
	data["auth_source_default_github_grant_on_first_bind"] = authSourceDefaults.GitHub.GrantOnFirstBind
	data["auth_source_default_google_balance"] = authSourceDefaults.Google.Balance
	data["auth_source_default_google_concurrency"] = authSourceDefaults.Google.Concurrency
	data["auth_source_default_google_subscriptions"] = authSourceDefaults.Google.Subscriptions
	data["auth_source_default_google_grant_on_signup"] = authSourceDefaults.Google.GrantOnSignup
	data["auth_source_default_google_grant_on_first_bind"] = authSourceDefaults.Google.GrantOnFirstBind
	data["auth_source_default_email_platform_quotas"] = authSourceDefaults.Email.PlatformQuotas
	data["auth_source_default_linuxdo_platform_quotas"] = authSourceDefaults.LinuxDo.PlatformQuotas
	data["auth_source_default_oidc_platform_quotas"] = authSourceDefaults.OIDC.PlatformQuotas
	data["auth_source_default_wechat_platform_quotas"] = authSourceDefaults.WeChat.PlatformQuotas
	data["auth_source_default_github_platform_quotas"] = authSourceDefaults.GitHub.PlatformQuotas
	data["auth_source_default_google_platform_quotas"] = authSourceDefaults.Google.PlatformQuotas
	data["auth_source_default_dingtalk_platform_quotas"] = authSourceDefaults.DingTalk.PlatformQuotas
	data["force_email_on_third_party_signup"] = authSourceDefaults.ForceEmailOnThirdPartySignup

	return data
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalDefaultSubscriptions(a, b []service.DefaultSubscriptionSetting) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].GroupID != b[i].GroupID || a[i].ValidityDays != b[i].ValidityDays {
			return false
		}
	}
	return true
}

func equalLoginAgreementDocuments(a, b []service.LoginAgreementDocument) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID || a[i].Title != b[i].Title || a[i].ContentMD != b[i].ContentMD {
			return false
		}
	}
	return true
}

func equalIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalNotifyEmailEntries(a, b []service.NotifyEmailEntry) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Email != b[i].Email || a[i].Verified != b[i].Verified || a[i].Disabled != b[i].Disabled {
			return false
		}
	}
	return true
}

// TestSMTPRequest 濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴ｆ閺嬩線鏌熼梻瀵割槮缁炬儳顭烽弻锝夊箛椤掍讲鏋欓梺缁樺笩婵倗鎹㈠☉銏犲耿婵☆垳绮崕搴ㄦ⒑鏉炴壆顦﹂悗姘嵆瀵鈽夐姀鈺傛櫇闂佺粯蓱瑜板啯鎱ㄩ弴銏♀拺缂佸娼￠妤呮煥閺囥劋閭€殿喖顭烽弫鎰緞婵炩拃鍥ㄧ厱闁靛绲芥俊鍏笺亜韫囨毣鏈犻梻鍌氬€搁崐鎼佸磹妞嬪海鐭嗗〒姘ｅ亾妤犵偞鐗犻、鏇氱秴闁搞儺鍓﹂弫鍐煥閺囨浜鹃梺姹囧€楅崑鎾舵崲濠靛顥堟繛鎴濆船閸撴壆绱掗悙顒€鍔ゆい顓犲厴瀵鏁愭径濠勭杸濡炪倖甯掗崐椋庡垝閻㈠憡鈷戦柛婵嗗椤ョ偟绱掗鑺ュ磳妤犵偛鍟灃闁告侗鍣Λ鍐ㄢ攽閻愭潙鐏﹂柣鐔村劦瀵娊宕奸妷锔惧幗闁瑰吋鐣崹濠氥€傞懠顑藉亾閸忓浜剧紓浣割儐椤戞瑥顭囬弽銊х鐎瑰壊鍠曠花濠氭煕閵娿儱鈧綊濡甸崟顖氬嵆闁绘劖娼欏▍锝夋⒑閸涘﹥灏柛鐔锋健閸╃偤骞嬮敂钘変汗濡炪倖鍔﹀鈧?
type TestSMTPRequest struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	SMTPUseTLS   bool   `json:"smtp_use_tls"`
}

// TestSMTPConnection 濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴ｆ閺嬩線鏌熼梻瀵割槮缁炬儳顭烽弻锝夊箛椤掍讲鏋欓梺缁樺笩婵倗鎹㈠☉銏犲耿婵☆垳绮崕搴ㄦ⒑鏉炴壆顦﹂悗姘嵆瀵鈽夐姀鈺傛櫇闂佺粯蓱瑜板啯鎱ㄩ弴銏♀拺缂佸娼￠妤呮煥閺囥劋閭€殿喖顭烽弫鎰緞婵炩拃鍥ㄧ厱闁靛绲芥俊鍏笺亜韫囨毣鏈犻梻鍌氬€搁崐鎼佸磹妞嬪海鐭嗗〒姘ｅ亾妤犵偞鐗犻、鏇氱秴闁搞儺鍓﹂弫鍐煥閺囨浜鹃梺姹囧€楅崑鎾舵崲濠靛顥堟繛鎴濆船閸撴壆绱掗悙顒€鍔ゆい顓犲厴瀵鏁愭径濠勭杸濡炪倖甯掗崐椋庡垝閻㈠憡鈷?
// POST /api/v1/admin/settings/test-smtp
func (h *SettingHandler) TestSMTPConnection(c *gin.Context) {
	var req TestSMTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	req.SMTPHost = strings.TrimSpace(req.SMTPHost)
	req.SMTPUsername = strings.TrimSpace(req.SMTPUsername)

	var savedConfig *service.SMTPConfig
	if cfg, err := h.emailService.GetSMTPConfig(c.Request.Context()); err == nil && cfg != nil {
		savedConfig = cfg
	}

	if req.SMTPHost == "" && savedConfig != nil {
		req.SMTPHost = savedConfig.Host
	}
	if req.SMTPPort <= 0 {
		if savedConfig != nil && savedConfig.Port > 0 {
			req.SMTPPort = savedConfig.Port
		} else {
			req.SMTPPort = 587
		}
	}
	if req.SMTPUsername == "" && savedConfig != nil {
		req.SMTPUsername = savedConfig.Username
	}
	password := strings.TrimSpace(req.SMTPPassword)
	if password == "" && savedConfig != nil {
		password = savedConfig.Password
	}
	if req.SMTPHost == "" {
		response.BadRequest(c, "SMTP host is required")
		return
	}

	config := &service.SMTPConfig{
		Host:     req.SMTPHost,
		Port:     req.SMTPPort,
		Username: req.SMTPUsername,
		Password: password,
		UseTLS:   req.SMTPUseTLS,
	}

	err := h.emailService.TestSMTPConnectionWithConfig(config)
	if err != nil {
		response.BadRequest(c, "SMTP connection test failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "SMTP connection successful"})
}

// SendTestEmailRequest is the request payload for SMTP test email.
type SendTestEmailRequest struct {
	Email        string `json:"email" binding:"required,email"`
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	SMTPFrom     string `json:"smtp_from_email"`
	SMTPFromName string `json:"smtp_from_name"`
	SMTPUseTLS   bool   `json:"smtp_use_tls"`
}

// SendTestEmail 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柛鎰紦閻㈠姊绘担鐟邦嚋婵炲弶鐗犲畷鎰板箹娴ｇ鐎梺绋跨灱閸嬬偤鎮￠弴鐘冲枑閹兼番鍔婇埀顒€鍟村畷銊р偓娑櫭埀顒€娼￠弻娑⑩€﹂幋婵呯按婵炲瓨绮嶇划鎾诲蓟閻斿憡缍囬柛鎾楀惙鎴犵磼缂併垹骞栭柣鏍с偢瀵鈽夐姀鐘愁棟闁荤姵浜介崜閬嶅Χ閸モ晝锛滈柡澶婄墑閸斿瞼绮鑸电厵鐎瑰嫮澧楅崵鍥┾偓瑙勬礃缁捇鐛幘璇茬闁瑰灝鍟弲娆撴⒒閸屾艾鈧绮堟担铏圭濠电姴娉氭径鎰閻庨潧鎽滈鏇㈡⒑闁偛鑻晶浼存煃瑜滈崜姘额敊閺嶎厼绐楁俊銈呮噹缁犱即鏌ц箛锝呬簼闁告瑥绻橀弻宥堫檨闁告挾鍠栧濠氭晲閸涘倻鍠愬鍕矙閹稿骸鍓甸梻鍌欑劍閻綊宕曢柆宥呯疇閹兼番鍔岄悡?// POST /api/v1/admin/settings/send-test-email
func (h *SettingHandler) SendTestEmail(c *gin.Context) {
	var req SendTestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	req.SMTPHost = strings.TrimSpace(req.SMTPHost)
	req.SMTPUsername = strings.TrimSpace(req.SMTPUsername)
	req.SMTPFrom = strings.TrimSpace(req.SMTPFrom)
	req.SMTPFromName = strings.TrimSpace(req.SMTPFromName)

	var savedConfig *service.SMTPConfig
	if cfg, err := h.emailService.GetSMTPConfig(c.Request.Context()); err == nil && cfg != nil {
		savedConfig = cfg
	}

	if req.SMTPHost == "" && savedConfig != nil {
		req.SMTPHost = savedConfig.Host
	}
	if req.SMTPPort <= 0 {
		if savedConfig != nil && savedConfig.Port > 0 {
			req.SMTPPort = savedConfig.Port
		} else {
			req.SMTPPort = 587
		}
	}
	if req.SMTPUsername == "" && savedConfig != nil {
		req.SMTPUsername = savedConfig.Username
	}
	password := strings.TrimSpace(req.SMTPPassword)
	if password == "" && savedConfig != nil {
		password = savedConfig.Password
	}
	if req.SMTPFrom == "" && savedConfig != nil {
		req.SMTPFrom = savedConfig.From
	}
	if req.SMTPFromName == "" && savedConfig != nil {
		req.SMTPFromName = savedConfig.FromName
	}
	if req.SMTPHost == "" {
		response.BadRequest(c, "SMTP host is required")
		return
	}

	config := &service.SMTPConfig{
		Host:     req.SMTPHost,
		Port:     req.SMTPPort,
		Username: req.SMTPUsername,
		Password: password,
		From:     req.SMTPFrom,
		FromName: req.SMTPFromName,
		UseTLS:   req.SMTPUseTLS,
	}

	siteName := h.settingService.GetSiteName(c.Request.Context())
	subject := "[" + siteName + "] Test Email"
	body := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f5f5f5; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; }
        .content { padding: 40px 30px; text-align: center; }
        .success { color: #10b981; font-size: 48px; margin-bottom: 20px; }
        .footer { background-color: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>` + siteName + `</h1>
        </div>
        <div class="content">
            <div class="success">闂?/div>
            <h2>Email Configuration Successful!</h2>
            <p>This is a test email to verify your SMTP settings are working correctly.</p>
        </div>
        <div class="footer">
            <p>This is an automated test message.</p>
        </div>
    </div>
</body>
</html>
`

	if err := h.emailService.SendEmailWithConfig(config, req.Email, subject, body); err != nil {
		response.BadRequest(c, "Failed to send test email: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Test email sent successfully"})
}

// GetAdminAPIKey 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕閻庤娲忛崕鎶藉焵椤掑﹦绉甸柛鐘崇墱婢规洟宕稿Δ浣哄幍闂佽鍨虫晶妤吽夋径鎰闁哄鍩婇煬顒勬煛鐏炶鈧繈骞婂┑瀣妞ゆ棁鍋愭晶顖氣攽閻愯尙鎽犵紒顔肩灱缁辩偞绻濋崶褑鎽曞┑鐐村灟閸ㄧ懓鏁梻渚€娼х换鍡涘礈濠靛鍋熼柡宥冨妿缁犻箖鏌熺€电浠﹂悘蹇ｅ幘缁辨帗寰勬繝鍕ㄩ悗娈垮枛椤攱淇婇幖浣肝ㄩ柕蹇婃濞兼梹绻濈喊妯活潑闁割煈鍨抽幏鍐晜閻ｅ矈娴勬俊銈忕到閸燁垶鍩涢幋鐘电＜閻庯綆浜滈惃锟犳煕閺冨倸鏋涢柡灞剧〒閳ь剨绲婚崝宀勫焵椤掍胶绠撴い?API Key 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸ゅ嫰鏌涢幘鑼槮闁搞劍绻冮妵鍕冀椤愵澀姹楅梺閫炲苯澧剧紒鐘虫尭閻ｉ攱绺界粙璇俱劑鏌曟径娑㈡鐞氾綁姊婚崒娆戭槮闁圭⒈鍋嗙划娆愮瑹閳ь剙鐣烽姀锛勯檮闁告稑锕ら埀?// GET /api/v1/admin/settings/admin-api-key
func (h *SettingHandler) GetAdminAPIKey(c *gin.Context) {
	maskedKey, exists, err := h.settingService.GetAdminAPIKeyStatus(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"exists":     exists,
		"masked_key": maskedKey,
	})
}

// RegenerateAdminAPIKey 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弴鐐测偓褰掑磿閹寸姵鍠愰柣妤€鐗嗙粭鎺楁煕閵娿儱鈧悂鍩為幋锔藉亹閻庡湱濮撮ˉ婵堢磽娴ｇ懓濮堟い銊ワ躬瀵鎮㈤崗鐓庝罕闂佸壊鍋嗛崰鎾诲礄閿熺姵鈷?闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮村鑸殿€栨繛瀛樼矊缂嶅﹪寮诲☉銏犵疀闁稿繐鎽滈弫鏍⒑濞茶骞楅柟鐟版喘瀵鏁愭径濠庢綂闂侀潧绻嗛弲婵嬪礉閹间焦鈷戦柦妯侯槸閺嗙喖鏌涢悩鏌ュ弰闁糕晝鍋ら獮瀣晜閽樺姹楅梻浣告啞閻熴儵藝椤栫倛鍥煛閸涱喒鎷虹紓浣割儐椤戞瑩宕曞澶嬬厱闁哄啠鍋撻柛銊ョ仢椤曪絾绻濆顓熸珳婵犮垼娉涢敃锕傛儓閸曨垱鈷戠紒瀣劵椤箓鏌涙繝鍐炬疁鐎规洘绻堥獮鏍ㄦ媴閸忓瀚奸梺鑽ゅТ濞测晝浜稿▎鎴犱笉闁绘鐗忕粻楣冩倶閻愭彃鈧憡淇婇幐搴濈箚闁告瑥顦慨宥嗩殽閻愭潙绗掗摶鏍归敐澶嬫珳闁哄鍙冨缁樻媴閸涘﹤鏆堟繛鎾寸椤ㄥ﹤鐣疯ぐ鎺戠闁兼亽鍎插▍鏍⒑缂佹ɑ鐓ラ柛姘儔閸╂盯骞嬮敂钘変化闂佹悶鍎荤徊娲磻閹剧粯鎯炴い鎰╁€撻柇顖炴⒒閸屾瑧鍔嶉悗绗涘厾娲晜閻ｅ矈娲稿┑鐘诧工閻楀﹪宕?API Key
// POST /api/v1/admin/settings/admin-api-key/regenerate
func (h *SettingHandler) RegenerateAdminAPIKey(c *gin.Context) {
	key, err := h.settingService.GenerateAdminAPIKey(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"key": key,
	})
}

// DeleteAdminAPIKey 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极瀹ュ绀嬫い鎺嶇劍椤斿洭姊绘担铏瑰笡闁告梹娲熼、姘额敇閻樺吀绗夋俊銈忕到閸燁垶鎮￠崘顏呭枑婵犲﹤鐗嗙粈鍫熸叏濡潡鍝虹€规洖寮剁换娑㈠箣濞嗗繒浠奸悗鐟版啞缁诲啴濡甸崟顖氱閻犺櫣鍎ら悗楣冩倵鐟欏嫭绀€缁剧虎鍘惧Σ鎰板箻鐎靛摜鎳濋梺鎼炲劀閸曨厺閭┑锛勫亼閸娿倝宕㈡ィ鍐ㄧ婵☆垯璀﹂崵鏇熴亜閹扳晛鐒洪柛姘儏椤法鎹勯悮鏉戝濡炪倧闄勫姗€鈥旈崘顔嘉ч柛鈩冾殔濞兼垿姊虹粙娆惧剱闁圭懓娲璇差吋閸偅顎囬梻浣告啞閹稿鎯勯姘辨殾闁硅揪绠戝洿婵犮垼娉涢敃銊╁箺?API Key
// DELETE /api/v1/admin/settings/admin-api-key
func (h *SettingHandler) DeleteAdminAPIKey(c *gin.Context) {
	if err := h.settingService.DeleteAdminAPIKey(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Admin API key deleted"})
}

// GetOverloadCooldownSettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕閻庤娲忛崕鎶藉焵椤掑﹦绉甸柛鐘崇墱婢规洟宕稿Δ浣哄幍闂佽鍨虫晶妤吽夋径鎰闁哄鍩婇煬顒勬煛鐏炶鈧繈骞婂┑瀣妞ゆ棁鍋愭晶顖氣攽?29闂傚倸鍊搁崐鎼佸磹妞嬪海鐭嗗〒姘ｅ亾妤犵偞鐗犻、鏇氱秴闁搞儺鍓﹂弫鍐煥閺囨浜鹃梺姹囧€楅崑鎾舵崲濞戙垹绠ｉ柣鎰ㄦ櫆閿涘牆鈹戦悙鍙夆枙濞存粍绮撻幃鈥斥槈濡繐缍婇幃鈩冩償闄囬崺鐐电磽閸屾氨孝闁挎洏鍊濋幃楣冩倻閽樺鍞堕梺鍝勬储閸斿秹寮查姀銈嗏拺闂侇偆鍋涢懟顖涙櫠閹殿喚纾兼い鏃囧亹婢э妇鈧娲橀敃銏ゃ€侀弮鈧幏鍛存⒐閹邦剦妫滈梻鍌欐祰瀹曠敻宕伴幇顒夌唵婵せ鍋撴い銏℃椤㈡棃宕ㄩ婊呮闂備線娼х换鍡椢ｉ崨瀛樺€块柣锝呯灱绾惧吋銇勯弮鍌樷偓瀣叕椤掑嫭鐓涢悘鐐跺Г椤ユ粍銇勯幘鐐藉仮鐎规洝绮剧粻娑㈠棘濞嗗彞绱梻鍌氬€搁崐椋庣矆娓氣偓楠炴牠顢曚綅閸ヮ剙绀冩繛鏉戭儐閻忓啴姊洪崫鍕闁逞屽墯閺嬪ジ宕戦幘璇茬妞ゆ棁袙閹?
// GET /api/v1/admin/settings/overload-cooldown
func (h *SettingHandler) GetOverloadCooldownSettings(c *gin.Context) {
	settings, err := h.settingService.GetOverloadCooldownSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.OverloadCooldownSettings{
		Enabled:         settings.Enabled,
		CooldownMinutes: settings.CooldownMinutes,
	})
}

// UpdateOverloadCooldownSettingsRequest 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犵厱婵﹩鍘介妵婵嗏攽闄囬崺鏍ь嚗閸曨厸鍋撻敐搴濈胺婵″弶鍔欏缁樼瑹閳ь剙顭囪婢ф繈姊洪崫鍕櫤闁烩晩鍨堕獮?29闂傚倸鍊搁崐鎼佸磹妞嬪海鐭嗗〒姘ｅ亾妤犵偞鐗犻、鏇氱秴闁搞儺鍓﹂弫鍐煥閺囨浜鹃梺姹囧€楅崑鎾舵崲濞戙垹绠ｉ柣鎰ㄦ櫆閿涘牆鈹戦悙鍙夆枙濞存粍绮撻幃鈥斥槈濡繐缍婇幃鈩冩償闄囬崺鐐电磽閸屾氨孝闁挎洏鍊濋幃楣冩倻閽樺鍞堕梺鍝勬储閸斿秹寮查姀銈嗏拺闂侇偆鍋涢懟顖涙櫠閹殿喚纾兼い鏃囧亹婢э妇鈧娲橀敃銏ゃ€侀弮鈧幏鍛存⒐閹邦剦妫滈梻鍌欐祰瀹曠敻宕伴幇顒夌唵婵せ鍋撴い銏℃椤㈡棃宕ㄩ婊呮闂備線娼х换鍡椢ｉ崨瀛樺€块柣锝呯灱绾惧吋銇勯弮鍌樷偓瀣叕椤掑嫭鐓涢悘鐐跺Г椤ユ粍銇勯幘鐐藉仮鐎规洝绮剧粻娑㈠棘濞嗗彞绱梻鍌氬€搁崐椋庣矆娓氣偓楠炴牠顢曚綅閸ヮ剙绀冩繛鏉戭儐閻忓啴姊洪崫鍕闁逞屽墯閺嬪ジ宕戦幘璇茬妞ゆ棁袙閹风粯绻涙潏鍓у埌婵犫偓閸楃儑鑰垮ù鐘差儐閻撴洖鈹戦悩鎻掓殲缂佲偓閳ь剟鏌ч懡銈呬沪缂佺粯鐩獮瀣枎韫囨洑鎮ｇ紓鍌欒兌婵參宕归崼鏇炶摕闁挎繂鎲橀悢鍏煎殐闁宠桨鑳堕濂告⒒娴ｈ櫣甯涢柟绋跨埣瀹曟洟鎮界粙璺ㄧ暫濠德板€曢幊蹇涘疾閺屻儱绠圭紒顔煎帨閸嬫捇骞囨担瑙勭グ闂?
type UpdateOverloadCooldownSettingsRequest struct {
	Enabled         bool `json:"enabled"`
	CooldownMinutes int  `json:"cooldown_minutes"`
}

// UpdateOverloadCooldownSettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犵厱婵﹩鍘介妵婵嗏攽闄囬崺鏍ь嚗閸曨厸鍋撻敐搴濈胺婵″弶鍔欏缁樼瑹閳ь剙顭囪婢ф繈姊洪崫鍕櫤闁烩晩鍨堕獮?29闂傚倸鍊搁崐鎼佸磹妞嬪海鐭嗗〒姘ｅ亾妤犵偞鐗犻、鏇氱秴闁搞儺鍓﹂弫鍐煥閺囨浜鹃梺姹囧€楅崑鎾舵崲濞戙垹绠ｉ柣鎰ㄦ櫆閿涘牆鈹戦悙鍙夆枙濞存粍绮撻幃鈥斥槈濡繐缍婇幃鈩冩償闄囬崺鐐电磽閸屾氨孝闁挎洏鍊濋幃楣冩倻閽樺鍞堕梺鍝勬储閸斿秹寮查姀銈嗏拺闂侇偆鍋涢懟顖涙櫠閹殿喚纾兼い鏃囧亹婢э妇鈧娲橀敃銏ゃ€侀弮鈧幏鍛存⒐閹邦剦妫滈梻鍌欐祰瀹曠敻宕伴幇顒夌唵婵せ鍋撴い銏℃椤㈡棃宕ㄩ婊呮闂備線娼х换鍡椢ｉ崨瀛樺€块柣锝呯灱绾惧吋銇勯弮鍌樷偓瀣叕椤掑嫭鐓涢悘鐐跺Г椤ユ粍銇勯幘鐐藉仮鐎规洝绮剧粻娑㈠棘濞嗗彞绱梻鍌氬€搁崐椋庣矆娓氣偓楠炴牠顢曚綅閸ヮ剙绀冩繛鏉戭儐閻忓啴姊洪崫鍕闁逞屽墯閺嬪ジ宕戦幘璇茬妞ゆ棁袙閹?
// PUT /api/v1/admin/settings/overload-cooldown
func (h *SettingHandler) UpdateOverloadCooldownSettings(c *gin.Context) {
	var req UpdateOverloadCooldownSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	settings := &service.OverloadCooldownSettings{
		Enabled:         req.Enabled,
		CooldownMinutes: req.CooldownMinutes,
	}

	if err := h.settingService.SetOverloadCooldownSettings(c.Request.Context(), settings); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updatedSettings, err := h.settingService.GetOverloadCooldownSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.OverloadCooldownSettings{
		Enabled:         updatedSettings.Enabled,
		CooldownMinutes: updatedSettings.CooldownMinutes,
	})
}

// GetRateLimit429CooldownSettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕閻庤娲忛崕鎶藉焵椤掑﹦绉甸柛鐘崇墱婢规洟宕稿Δ浣哄幍闂佽鍨虫晶妤吽夋径鎰闁哄鍩婇煬顒勬煛鐏炶鈧繈骞婂┑瀣妞ゆ棁鍋愭晶顖氣攽?29婵犵數濮烽弫鍛婃叏閻㈠壊鏁婇柡宥庡幖缁愭淇婇妶鍛殲闁哄棙绮嶆穱濠囧Χ閸涱厽娈堕梺娲诲幗閻熲晠寮婚悢鍏煎€绘俊顖炴櫜缁爼姊洪柅鐐茶嫰婢у墽绱掗悩铏碍闁伙綁鏀辩缓鐣岀矙鐠囦勘鍔嶉妵鍕籍閸ヮ灝鎾寸箾閸涱厾孝闁宠鍨块幃鈺呭垂椤愶絾鐦庡┑鐘愁問閸犳岸寮繝姘疇婵犻潧娲㈤崑鍛存煕閹般劍娅呴柍褜鍓欓崲鏌ュ煘閹达附鍋愰柛娆忣槸椤︹晠姊虹紒妯诲暗闁哥姵鐗犲濠氭偄閸忚偐鍔烽梺鎸庢磵閸嬫捇鏌ｈ箛銉х瘈闁哄本娲熷畷鍗炍旈埀顒勫汲閻旇櫣纾奸弶鍫涘妽瀹曞瞼鈧娲樼敮鎺楋綖濠靛绀傞悘蹇旂墬閻庢澘鈹戦悩鎰佸晱闁哥姵鐗犻弫鍐Ψ閵夘喗瀵岄梺鑺ッˇ浠嬪吹閺囥垺鐓曢柟鎹愬皺閸斿秶鈧娲栭ˇ鐢稿蓟閺囩喓绡€闁告劑鍔岄～鍥⒑缁嬪尅鏀荤紒璇茬墕椤?
// GET /api/v1/admin/settings/rate-limit-429-cooldown
func (h *SettingHandler) GetRateLimit429CooldownSettings(c *gin.Context) {
	settings, err := h.settingService.GetRateLimit429CooldownSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.RateLimit429CooldownSettings{
		Enabled:         settings.Enabled,
		CooldownSeconds: settings.CooldownSeconds,
	})
}

// UpdateRateLimit429CooldownSettingsRequest 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犵厱婵﹩鍘介妵婵嗏攽闄囬崺鏍ь嚗閸曨厸鍋撻敐搴濈胺婵″弶鍔欏缁樼瑹閳ь剙顭囪婢ф繈姊洪崫鍕櫤闁烩晩鍨堕獮?29婵犵數濮烽弫鍛婃叏閻㈠壊鏁婇柡宥庡幖缁愭淇婇妶鍛殲闁哄棙绮嶆穱濠囧Χ閸涱厽娈堕梺娲诲幗閻熲晠寮婚悢鍏煎€绘俊顖炴櫜缁爼姊洪柅鐐茶嫰婢у墽绱掗悩铏碍闁伙綁鏀辩缓鐣岀矙鐠囦勘鍔嶉妵鍕籍閸ヮ灝鎾寸箾閸涱厾孝闁宠鍨块幃鈺呭垂椤愶絾鐦庡┑鐘愁問閸犳岸寮繝姘疇婵犻潧娲㈤崑鍛存煕閹般劍娅呴柍褜鍓欓崲鏌ュ煘閹达附鍋愰柛娆忣槸椤︹晠姊虹紒妯诲暗闁哥姵鐗犲濠氭偄閸忚偐鍔烽梺鎸庢磵閸嬫捇鏌ｈ箛銉х瘈闁哄本娲熷畷鍗炍旈埀顒勫汲閻旇櫣纾奸弶鍫涘妽瀹曞瞼鈧娲樼敮鎺楋綖濠靛绀傞悘蹇旂墬閻庢澘鈹戦悩鎰佸晱闁哥姵鐗犻弫鍐Ψ閵夘喗瀵岄梺鑺ッˇ浠嬪吹閺囥垺鐓曢柟鎹愬皺閸斿秶鈧娲栭ˇ鐢稿蓟閺囩喓绡€闁告劑鍔岄～鍥⒑缁嬪尅鏀荤紒璇茬墕椤曪絾绻濆顓熸珳婵犮垼娉涜癌鐟滄棃寮诲☉妯锋闁告鍋為悵顔尖攽閳藉棗浜栭柛濠冪箓椤繑銈︾憗銈勬睏闂佸湱鍎ょ换鍐夐弽銊х瘈闁汇垽娼ф禍褰掓煕鐎ｎ偅宕屾慨濠冩そ瀹曨偊宕熼鍛晧闂備礁鎲￠弻銊╂煀閿濆鏄ラ柍褜鍓氶妵鍕箳閹搭垰濮涚紓浣割槺閺佸寮诲☉銏犵闁哄鍨甸幗鐢告倵?
type UpdateRateLimit429CooldownSettingsRequest struct {
	Enabled         bool `json:"enabled"`
	CooldownSeconds int  `json:"cooldown_seconds"`
}

// UpdateRateLimit429CooldownSettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犵厱婵﹩鍘介妵婵嗏攽闄囬崺鏍ь嚗閸曨厸鍋撻敐搴濈胺婵″弶鍔欏缁樼瑹閳ь剙顭囪婢ф繈姊洪崫鍕櫤闁烩晩鍨堕獮?29婵犵數濮烽弫鍛婃叏閻㈠壊鏁婇柡宥庡幖缁愭淇婇妶鍛殲闁哄棙绮嶆穱濠囧Χ閸涱厽娈堕梺娲诲幗閻熲晠寮婚悢鍏煎€绘俊顖炴櫜缁爼姊洪柅鐐茶嫰婢у墽绱掗悩铏碍闁伙綁鏀辩缓鐣岀矙鐠囦勘鍔嶉妵鍕籍閸ヮ灝鎾寸箾閸涱厾孝闁宠鍨块幃鈺呭垂椤愶絾鐦庡┑鐘愁問閸犳岸寮繝姘疇婵犻潧娲㈤崑鍛存煕閹般劍娅呴柍褜鍓欓崲鏌ュ煘閹达附鍋愰柛娆忣槸椤︹晠姊虹紒妯诲暗闁哥姵鐗犲濠氭偄閸忚偐鍔烽梺鎸庢磵閸嬫捇鏌ｈ箛銉х瘈闁哄本娲熷畷鍗炍旈埀顒勫汲閻旇櫣纾奸弶鍫涘妽瀹曞瞼鈧娲樼敮鎺楋綖濠靛绀傞悘蹇旂墬閻庢澘鈹戦悩鎰佸晱闁哥姵鐗犻弫鍐Ψ閵夘喗瀵岄梺鑺ッˇ浠嬪吹閺囥垺鐓曢柟鎹愬皺閸斿秶鈧娲栭ˇ鐢稿蓟閺囩喓绡€闁告劑鍔岄～鍥⒑缁嬪尅鏀荤紒璇茬墕椤?
// PUT /api/v1/admin/settings/rate-limit-429-cooldown
func (h *SettingHandler) UpdateRateLimit429CooldownSettings(c *gin.Context) {
	var req UpdateRateLimit429CooldownSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	settings := &service.RateLimit429CooldownSettings{
		Enabled:         req.Enabled,
		CooldownSeconds: req.CooldownSeconds,
	}

	if err := h.settingService.SetRateLimit429CooldownSettings(c.Request.Context(), settings); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updatedSettings, err := h.settingService.GetRateLimit429CooldownSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.RateLimit429CooldownSettings{
		Enabled:         updatedSettings.Enabled,
		CooldownSeconds: updatedSettings.CooldownSeconds,
	})
}

// GetStreamTimeoutSettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕閻庤娲忛崕鎶藉焵椤掑﹦绉甸柛鐘崇墱婢规洟宕稿Δ浣哄幍闂佽鍨虫晶妤吽夋径鎰闁哄鍩婇煬顒勬煛鐏炶鈧繈骞婂┑瀣妞ゆ棁鍋愭晶顖氣攽閻愯尙鎽犵紒顔肩灱缁辩偞绻濋崶褑鎽曞┑鐐村灟閸ㄧ懓鏁俊鐐€栭弻銊ㄣ亹閻愭畫褍螖閸涱喚鍘介梺缁樻煥閹芥粓骞婇崘顔藉€垫慨妯煎帶楠炴鏌涢幒鎴含妤犵偞锕㈤、娆撴偩鐏炶棄绠洪梻鍌欑劍鐎笛呮崲閸岀偛绠犻柟閭﹀墻閻掍粙鏌ㄩ悢鍝勑ｉ柣鎾崇箻閻擃偊宕惰閸庡繘鏌嶈閸撴岸鎮ч悩璇茬畺闁炽儲鏋煎Σ鍫ユ煏韫囥儳纾块柛姗€浜跺娲焻濞戞ê绁┑鐐板尃閸モ晛绁﹀┑掳鍊曢崯鏉课涢娑栦簻闁规崘娉涘瓭婵炲瓨绮撶粻鏍蓟閺囥垹鐐婄憸宥夘敂椤撱垺鐓涘ù锝囶焾閳ь剙顭烽獮澶愬箻椤旇偐顦板銈嗗笒閸婂摜绱為崘顔解拻闁稿本鐟︾粊鐗堛亜閺囩喓澧电€规洑鍗冲浠嬵敇閻旇渹鍑介梻浣虹帛閹哥霉閻戣棄姹查柨鏇炲€归悡鏇熶繆閵堝懎鏆欏ù婊嗩潐娣囧﹪骞撻幒鏂款杸婵烇絽娲ら敃顏堛€佸☉銏″€烽柣銏㈡暩閻涖儳绱撻崒娆愵樂缂佲偓娓氣偓椤㈡岸顢橀悩鎰佹綗?// GET /api/v1/admin/settings/stream-timeout
func (h *SettingHandler) GetStreamTimeoutSettings(c *gin.Context) {
	settings, err := h.settingService.GetStreamTimeoutSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.StreamTimeoutSettings{
		Enabled:                settings.Enabled,
		Action:                 settings.Action,
		TempUnschedMinutes:     settings.TempUnschedMinutes,
		ThresholdCount:         settings.ThresholdCount,
		ThresholdWindowMinutes: settings.ThresholdWindowMinutes,
	})
}

// GetRectifierSettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕閻庤娲忛崕鎶藉焵椤掑﹦绉甸柛鐘崇墱婢规洟宕稿Δ浣哄幍闂佽鍨虫晶妤吽夋径鎰闁哄鍩婇煬顒勬煛鐏炶鈧繈骞婂┑瀣妞ゆ棁鍋愭晶顖氣攽閻愯尙鎽犵紒顔肩灱缁辩偞绻濋崶褑鎽曞┑鐐村灟閸ㄧ懓鏁梻渚€娼х换鍡涘礈濠靛鍋熼柡鍥ュ灪閻撶喖骞栧ǎ顒€鐏い锝堝皺閳ь剙鍘滈崑鎾剁磼鐎ｎ偒鍎ユ繛鍏肩墬缁绘稑顔忛鑽ょ泿闂佹悶鍔岄崐褰掑Φ閸曨垰鍗抽柣鎰綑濞咃綁姊洪崨濠冨碍闁哥喎娼￠崺鐐哄箣閿旇棄浜瑰銈嗗姦濠⑩偓濠㈣娲熷缁樻媴閸涘﹨纭€闂佹儳绻愰柊锝呯暦閺夎鏃堝川椤旈棿绨垫繝鐢靛仦閸ㄥ爼鍩ｉ妶鍫涗汗闁圭儤鎸诲▍婊堟⒑閸涘﹣绶遍柛鐘愁殜瀹曚即骞囬悧鍫氭嫼闂侀潻瀵岄崢钘夆枍閺囩偐鏀芥い鏇楀亾妞ゃ儲鎸稿嵄闁圭増婢樼粻鎶芥煙閹冾暢闁硅櫕绻冪换婵嬫偨闂堟刀銏ゆ倵濮樼厧娅嶇€殿喗濞婃俊鑸靛緞鐎ｎ亖鍋撻悽鍛婂仭婵炲棗绻愰顏嗙磼閳ь剟宕橀妸銏℃杸闂佺粯鍔橀婊堝绩鐠囩潿搴ㄥ炊瑜濋煬顒侇殽閻愬澧柟宄版嚇瀹曨偊濡烽妷銉ь唶濠电姷鏁搁崑鐘诲箵椤忓棗绶ゅΔ锝呭暞閸嬨倝鏌￠崶鈺佇ラ柣顓炵墦閹妫冨☉娆愬枑缂?// GET /api/v1/admin/settings/rectifier
func (h *SettingHandler) GetRectifierSettings(c *gin.Context) {
	settings, err := h.settingService.GetRectifierSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	patterns := settings.APIKeySignaturePatterns
	if patterns == nil {
		patterns = []string{}
	}
	response.Success(c, dto.RectifierSettings{
		Enabled:                  settings.Enabled,
		ThinkingSignatureEnabled: settings.ThinkingSignatureEnabled,
		ThinkingBudgetEnabled:    settings.ThinkingBudgetEnabled,
		APIKeySignatureEnabled:   settings.APIKeySignatureEnabled,
		APIKeySignaturePatterns:  patterns,
	})
}

// UpdateRectifierSettingsRequest updates rectifier settings.
type UpdateRectifierSettingsRequest struct {
	Enabled                  bool     `json:"enabled"`
	ThinkingSignatureEnabled bool     `json:"thinking_signature_enabled"`
	ThinkingBudgetEnabled    bool     `json:"thinking_budget_enabled"`
	APIKeySignatureEnabled   bool     `json:"apikey_signature_enabled"`
	APIKeySignaturePatterns  []string `json:"apikey_signature_patterns"`
}

// UpdateRectifierSettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犵厱婵﹩鍘介妵婵嗏攽闄囬崺鏍ь嚗閸曨厸鍋撻敐搴濈胺婵″弶鍔欏缁樼瑹閳ь剙顭囪婢ф繈姊洪崫鍕櫤闁烩晩鍨堕獮蹇涘箣閿旇棄浜滈柣蹇撶箣閻掞箓寮埀顒勬⒒娴ｈ櫣甯涢拑杈╂喐閺夊灝鏆ｆい銏℃椤㈡棃宕奸悢灏佸亾閸偅鍙忔慨妤€妫楅獮鏍煕濠靛牆鍔嬮柟渚垮妽缁绘繈宕橀埞澶歌檸闁诲氦顫夊ú鏍礊婵犲洢鈧礁顫濈捄鍝勭獩濡炪倖鎸鹃崑娑⑺夐幇顑芥斀闁绘ê鐏氶弳鈺佲攽椤曞懎寮€规洘鍨块獮妯肩磼濡厧骞嶇紓鍌氬€烽梽宥夊垂瑜版帒鍑犻柣鏂垮悑閻撶喖鏌熼幆褜鍤熼柍钘夘樀濡焦寰勯幇顓犲弳濠电娀娼уΛ娆撳闯濞差亝鐓欓柛娆忣槹閸婃劙鏌″畝鈧崰鏍箖閸撗傛勃闁告瑯鍋侀崕宕囨閹烘挻缍囬柕濞垮劤閻熻尙绱撴担璇℃畼闁哥姵鍔楅幑銏犫攽鐎ｎ亞顦板銈嗘尵閸犳劙顢欓幘缁樷拺閻犲洤寮堕崬澶嬨亜椤愩埄妲搁悡銈嗕繆椤栨繂鍚圭紒鐘冲劤閳规垿鎮╅幓鎺撴缂備浇顕уΛ婵嬪蓟閻斿吋鈷掗悗闈涘濡差噣姊虹紒妯诲鞍缂佸鍨垮﹢渚€姊洪幐搴ｇ畵婵☆偅鐩獮妤呭礃椤忓棛锛滄繛杈剧到婢瑰﹪宕甸悢铏圭＜?// PUT /api/v1/admin/settings/rectifier
func (h *SettingHandler) UpdateRectifierSettings(c *gin.Context) {
	var req UpdateRectifierSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	const maxPatterns = 50
	const maxPatternLen = 500
	if len(req.APIKeySignaturePatterns) > maxPatterns {
		response.BadRequest(c, "Too many signature patterns (max 50)")
		return
	}
	var cleanedPatterns []string
	for _, p := range req.APIKeySignaturePatterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if len(p) > maxPatternLen {
			response.BadRequest(c, "Signature pattern too long (max 500 characters)")
			return
		}
		cleanedPatterns = append(cleanedPatterns, p)
	}

	settings := &service.RectifierSettings{
		Enabled:                  req.Enabled,
		ThinkingSignatureEnabled: req.ThinkingSignatureEnabled,
		ThinkingBudgetEnabled:    req.ThinkingBudgetEnabled,
		APIKeySignatureEnabled:   req.APIKeySignatureEnabled,
		APIKeySignaturePatterns:  cleanedPatterns,
	}

	if err := h.settingService.SetRectifierSettings(c.Request.Context(), settings); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮村鑸殿€栨繛瀛樼矊缂嶅﹪寮诲☉銏犵疀闁稿繐鎽滈弫鏍⒑濞茶骞楅柟鐟版喘瀵鏁愭径濠庢綂闂侀潧绻嗛弲婵嬪礉閹间焦鈷戦柦妯侯槸閺嗙喖鏌涢悩鏌ュ弰闁糕晝鍋ら獮瀣晜閽樺姹楅梻浣告啞閻熴儵藝椤栫倛鍥煛閸涱喒鎷洪梺纭呭亹閸嬫盯鍩€椤掍胶澧い顐㈢箲缁绘繂顫濋鍌氬Е婵＄偑鍊栫敮鎺楀磹瑜版帒鍚归柍褜鍓熼弻锝嗘償閵忕姴姣堥梺鍛娒妶鎼佸灳閿曞倸鍨傛い鏂诲劤閸犳牠骞冮埄鍐╁劅闁挎繂娲ㄩ幗宀勬⒒閸屾瑨鍏屾い顐㈩儔瀹曠喖宕归銈嗘闂傚倷鑳堕…鍫ヮ敄閹寸姴绶ゅù鐘差儐缁犳帡姊绘担鐟邦嚋缂佽鍊搁湁闁稿瞼鍋為崐鍨旈敐鍛殲闁绘挸鍟村娲垂椤曞懎鍓卞銈冨劚閻楁捇寮婚垾宕囨殕閻庯綆鍓涜ⅵ濠电姷顣介埀顒€纾崺锝団偓瑙勬磸閸旀垿銆佸璺哄窛妞ゆ洖妫涚壕鍏肩節閻㈤潧啸闁轰焦鎮傚畷鎴︽偄閻撳骸鍋嶉悷婊勬閸ㄩ箖鏁冮埀顒勬偩閿熺姵鐒介柨鏇楀亾婵炲牄鍔岄—鍐Χ閸℃鐟ㄩ柣搴㈠嚬閸撴稓鍒掔紒妯侯嚤闁哄鍨归鎰版煛婢跺﹦澧曞褏鏅划鏃堝垂椤愶紕绠?
	updatedSettings, err := h.settingService.GetRectifierSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	updatedPatterns := updatedSettings.APIKeySignaturePatterns
	if updatedPatterns == nil {
		updatedPatterns = []string{}
	}
	response.Success(c, dto.RectifierSettings{
		Enabled:                  updatedSettings.Enabled,
		ThinkingSignatureEnabled: updatedSettings.ThinkingSignatureEnabled,
		ThinkingBudgetEnabled:    updatedSettings.ThinkingBudgetEnabled,
		APIKeySignatureEnabled:   updatedSettings.APIKeySignatureEnabled,
		APIKeySignaturePatterns:  updatedPatterns,
	})
}

// GetBetaPolicySettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕閻庤娲忛崕鎶藉焵椤掑﹦绉甸柛鐘崇墱婢规洟宕稿Δ浣哄幍闂佽鍨虫晶妤吽夋径鎰闁哄鍩婇煬顒勬煛鐏炶鈧繈骞婂┑瀣妞ゆ棁鍋愭晶顖氣攽?Beta 缂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮崹顔规寖闂佹椿鍘介悷鈺呭蓟閻斿吋鍊绘慨妤€妫欓悾鐑芥⒑缁嬫鍎忔い鎴濇閹广垹鈹戦崶鈺冪槇闂佺鏈划宀勩€傞崫鍕ㄦ斀妞ゆ梹顑欏鎰版煟閹垮嫮绡€鐎殿喖顭烽幃銏ゆ偂鎼达綆鍚嬫俊鐐€栧濠氬煕閸儱姹叉い鎾卞灪閳锋垹绱撴担鑲℃垹浜搁妸鈺傜厱閻庯綆鍋呭畷宀勬煛瀹€瀣ɑ闁诡垱妫冮弫宥夊礋椤撶喐顔嗛梻鍌欒兌椤牊顨ョ粙璺ㄦ殾妞ゆ帒鍊稿鏌ユ⒒娓氣偓濞佳嗗闂佸搫鎳忛惄顖氱暦?
// GET /api/v1/admin/settings/beta-policy
func (h *SettingHandler) GetBetaPolicySettings(c *gin.Context) {
	settings, err := h.settingService.GetBetaPolicySettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	rules := make([]dto.BetaPolicyRule, len(settings.Rules))
	for i, r := range settings.Rules {
		rules[i] = dto.BetaPolicyRule(r)
	}
	response.Success(c, dto.BetaPolicySettings{Rules: rules})
}

// UpdateBetaPolicySettingsRequest 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犵厱婵﹩鍘介妵婵嗏攽闄囬崺鏍ь嚗閸曨厸鍋撻敐搴濈胺婵″弶鍔欏缁樼瑹閳ь剙顭囪婢ф繈姊洪崫鍕櫤闁烩晩鍨堕獮?Beta 缂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮崹顔规寖闂佹椿鍘介悷鈺呭蓟閻斿吋鍊绘慨妤€妫欓悾鐑芥⒑缁嬫鍎忔い鎴濇閹广垹鈹戦崶鈺冪槇闂佺鏈划宀勩€傞崫鍕ㄦ斀妞ゆ梹顑欏鎰版煟閹垮嫮绡€鐎殿喖顭烽幃銏ゆ偂鎼达綆鍚嬫俊鐐€栧濠氬煕閸儱姹叉い鎾卞灪閳锋垹绱撴担鑲℃垹浜搁妸鈺傜厱閻庯綆鍋呭畷宀勬煛瀹€瀣ɑ闁诡垱妫冮弫宥夊礋椤撶喐顔嗛梻鍌欒兌椤牊顨ョ粙璺ㄦ殾妞ゆ帒鍊稿鏌ユ⒒娓氣偓濞佳嗗闂佸搫鎳忛惄顖氱暦濞嗘挻鍋愮紓浣诡焽閸橀亶鎮楅崗澶婁壕闂侀€炲苯澧寸€规洘鍨块幃娆撳传閸曨叏绱遍梻浣瑰缁诲倿藝娴煎瓨鍋傞柡鍥ュ灪閻撳繐顭块懜寰楊亪寮稿☉姗嗙唵鐟滄粓宕归崼鏇炶摕闁斥晛鍟伴悷褰掓煃瑜滈崜姘辩矉瀹ュ閱囬柡鍥╁仩閹?
type UpdateBetaPolicySettingsRequest struct {
	Rules []dto.BetaPolicyRule `json:"rules"`
}

// UpdateBetaPolicySettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犵厱婵﹩鍘介妵婵嗏攽闄囬崺鏍ь嚗閸曨厸鍋撻敐搴濈胺婵″弶鍔欏缁樼瑹閳ь剙顭囪婢ф繈姊洪崫鍕櫤闁烩晩鍨堕獮?Beta 缂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮崹顔规寖闂佹椿鍘介悷鈺呭蓟閻斿吋鍊绘慨妤€妫欓悾鐑芥⒑缁嬫鍎忔い鎴濇閹广垹鈹戦崶鈺冪槇闂佺鏈划宀勩€傞崫鍕ㄦ斀妞ゆ梹顑欏鎰版煟閹垮嫮绡€鐎殿喖顭烽幃銏ゆ偂鎼达綆鍚嬫俊鐐€栧濠氬煕閸儱姹叉い鎾卞灪閳锋垹绱撴担鑲℃垹浜搁妸鈺傜厱閻庯綆鍋呭畷宀勬煛瀹€瀣ɑ闁诡垱妫冮弫宥夊礋椤撶喐顔嗛梻鍌欒兌椤牊顨ョ粙璺ㄦ殾妞ゆ帒鍊稿鏌ユ⒒娓氣偓濞佳嗗闂佸搫鎳忛惄顖氱暦?
// PUT /api/v1/admin/settings/beta-policy
func (h *SettingHandler) UpdateBetaPolicySettings(c *gin.Context) {
	var req UpdateBetaPolicySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	rules := make([]service.BetaPolicyRule, len(req.Rules))
	for i, r := range req.Rules {
		rules[i] = service.BetaPolicyRule(r)
	}

	settings := &service.BetaPolicySettings{Rules: rules}
	if err := h.settingService.SetBetaPolicySettings(c.Request.Context(), settings); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Re-fetch to return updated settings
	updated, err := h.settingService.GetBetaPolicySettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	outRules := make([]dto.BetaPolicyRule, len(updated.Rules))
	for i, r := range updated.Rules {
		outRules[i] = dto.BetaPolicyRule(r)
	}
	response.Success(c, dto.BetaPolicySettings{Rules: outRules})
}

// UpdateStreamTimeoutSettingsRequest updates stream timeout settings.
type UpdateStreamTimeoutSettingsRequest struct {
	Enabled                bool   `json:"enabled"`
	Action                 string `json:"action"`
	TempUnschedMinutes     int    `json:"temp_unsched_minutes"`
	ThresholdCount         int    `json:"threshold_count"`
	ThresholdWindowMinutes int    `json:"threshold_window_minutes"`
}

// UpdateStreamTimeoutSettings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犵厱婵﹩鍘介妵婵嗏攽闄囬崺鏍ь嚗閸曨厸鍋撻敐搴濈胺婵″弶鍔欏缁樼瑹閳ь剙顭囪婢ф繈姊洪崫鍕櫤闁烩晩鍨堕獮蹇涘箣閿旇棄浜滈柣蹇撶箣閻掞箓寮埀顒勬⒒娴ｈ櫣甯涙い顓炵墕椤曪綁骞橀懜娈挎綗闂佸搫绋侀崢浠嬪磻閳╁啰绠鹃柛鈩冾殘缁犵増銇勮箛锝勭盎闁宠鍨块、娑樷槈濞嗗繐鏀梻浣告惈閻绱炴笟鈧顐﹀箛閺夊灝绐涘銈嗙墬缁嬫垶绂掗幘顔解拻濞撴埃鍋撴繛浣冲懏宕查柛顐犲劚绾惧綊鏌″搴″箺闁稿鍊圭换娑㈠箣濞嗗繒浠鹃梺缁樻尰濞茬喖寮婚悢绋款嚤閻庢稒蓱闁款厼螖閻橀潧浠︽い銊ョ墦閹偓妞ゅ繐鐗滈弫鍥煟濡灝绱﹀瑙勬礋濮婃椽妫冨☉娆愭倷闁诲孩鐭崡鍐差嚕椤愶箑纾兼繝濠傚閺嬪倿姊洪崨濠冨闁告挻绋栭妵鎰偓锝庡枟閳锋帒霉閿濆洨鎽傞柛銈呭暣閺屾盯鎮╁畷鍥р拰闂佽鍠撻崕鐢稿箖閳╁啯鍎熼柨婵嗘椤旀洟姊绘担鍛婂暈妞ゃ劌鐗撳畷鎴﹀箛椤戔晪缍佸畷濂告偄閸撲胶鐣炬俊鐐€栭悧婊堝磻閻旂儤顫曢柟娈垮枤绾剧厧霉閿濆懎顥忛柛銈囧枔缁?// PUT /api/v1/admin/settings/stream-timeout
func (h *SettingHandler) UpdateStreamTimeoutSettings(c *gin.Context) {
	var req UpdateStreamTimeoutSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	settings := &service.StreamTimeoutSettings{
		Enabled:                req.Enabled,
		Action:                 req.Action,
		TempUnschedMinutes:     req.TempUnschedMinutes,
		ThresholdCount:         req.ThresholdCount,
		ThresholdWindowMinutes: req.ThresholdWindowMinutes,
	}

	if err := h.settingService.SetStreamTimeoutSettings(c.Request.Context(), settings); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮村鑸殿€栨繛瀛樼矊缂嶅﹪寮诲☉銏犵疀闁稿繐鎽滈弫鏍⒑濞茶骞楅柟鐟版喘瀵鏁愭径濠庢綂闂侀潧绻嗛弲婵嬪礉閹间焦鈷戦柦妯侯槸閺嗙喖鏌涢悩鏌ュ弰闁糕晝鍋ら獮瀣晜閽樺姹楅梻浣告啞閻熴儵藝椤栫倛鍥煛閸涱喒鎷洪梺纭呭亹閸嬫盯鍩€椤掍胶澧い顐㈢箲缁绘繂顫濋鍌氬Е婵＄偑鍊栫敮鎺楀磹瑜版帒鍚归柍褜鍓熼弻锝嗘償閵忕姴姣堥梺鍛娒妶鎼佸灳閿曞倸鍨傛い鏂诲劤閸犳牠骞冮埄鍐╁劅闁挎繂娲ㄩ幗宀勬⒒閸屾瑨鍏屾い顐㈩儔瀹曠喖宕归銈嗘闂傚倷鑳堕…鍫ヮ敄閹寸姴绶ゅù鐘差儐缁犳帡姊绘担鐟邦嚋缂佽鍊搁湁闁稿瞼鍋為崐鍨旈敐鍛殲闁绘挸鍟村娲垂椤曞懎鍓卞銈冨劚閻楁捇寮婚垾宕囨殕閻庯綆鍓涜ⅵ濠电姷顣介埀顒€纾崺锝団偓瑙勬磸閸旀垿銆佸璺哄窛妞ゆ洖妫涚壕鍏肩節閻㈤潧啸闁轰焦鎮傚畷鎴︽偄閻撳骸鍋嶉悷婊勬閸ㄩ箖鏁冮埀顒勬偩閿熺姵鐒介柨鏇楀亾婵炲牄鍔岄—鍐Χ閸℃鐟ㄩ柣搴㈠嚬閸撴稓鍒掔紒妯侯嚤闁哄鍨归鎰版煛婢跺﹦澧曞褏鏅划鏃堝垂椤愶紕绠?
	updatedSettings, err := h.settingService.GetStreamTimeoutSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.StreamTimeoutSettings{
		Enabled:                updatedSettings.Enabled,
		Action:                 updatedSettings.Action,
		TempUnschedMinutes:     updatedSettings.TempUnschedMinutes,
		ThresholdCount:         updatedSettings.ThresholdCount,
		ThresholdWindowMinutes: updatedSettings.ThresholdWindowMinutes,
	})
}

// GetWebSearchEmulationConfig 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕閻庤娲忛崕鎶藉焵椤掑﹦绉甸柛鐘崇墱婢规洟宕稿Δ浣哄幍闂佽鍨虫晶妤吽夋径鎰闁哄鍩婇煬顒勬煛鐏炶鈧繈骞婂┑瀣妞ゆ棁鍋愭晶顖氣攽?Web Search 濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴濐潟閳ь剙鍊块幐濠冪珶閳哄绉€规洏鍔戝鍫曞箣濠靛牃鍋撻鐑嗘富闁靛牆鎳愮粻浼存煟濡も偓濡稓鍒掗崼銉ラ唶闁绘棁娅ｉ惁鍫ユ⒒閸屾氨澧涚紒瀣尰閺呭爼骞囬悧鍫㈠幈闁诲函缍嗛崑鍛焊椤撶喆浜滄い鎰剁悼缁犵偞銇勯姀鈽呰€块柟顔规櫊瀹曟﹢宕ｆ径灞介叡闂傚倸鍊搁崐椋庣矆娓氣偓楠炴牠顢曢敂钘夋濡炪倖鐗滈崑鐐哄疾濠靛鐓忛柛顐ｇ箥濡牓鏌℃担鍛婃悙闁宠鍨垮畷鎺戭煥鎼达絽濮奸梻浣虹帛閸旀洟鏁冮鍫濊摕鐎广儱顦导鐘绘煕閺囥劌浜愰柛瀣尰缁绘繂顫濋鍌氬Е?
// GET /api/v1/admin/settings/web-search-emulation
func (h *SettingHandler) GetWebSearchEmulationConfig(c *gin.Context) {
	cfg, err := h.settingService.GetWebSearchEmulationConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, service.PopulateWebSearchUsage(c.Request.Context(), cfg))
}

// UpdateWebSearchEmulationConfig 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋為悧鐘汇€侀弴銏犵厱婵﹩鍘介妵婵嗏攽闄囬崺鏍ь嚗閸曨厸鍋撻敐搴濈胺婵″弶鍔欏缁樼瑹閳ь剙顭囪婢ф繈姊洪崫鍕櫤闁烩晩鍨堕獮?Web Search 濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴濐潟閳ь剙鍊块幐濠冪珶閳哄绉€规洏鍔戝鍫曞箣濠靛牃鍋撻鐑嗘富闁靛牆鎳愮粻浼存煟濡も偓濡稓鍒掗崼銉ラ唶闁绘棁娅ｉ惁鍫ユ⒒閸屾氨澧涚紒瀣尰閺呭爼骞囬悧鍫㈠幈闁诲函缍嗛崑鍛焊椤撶喆浜滄い鎰剁悼缁犵偞銇勯姀鈽呰€块柟顔规櫊瀹曟﹢宕ｆ径灞介叡闂傚倸鍊搁崐椋庣矆娓氣偓楠炴牠顢曢敂钘夋濡炪倖鐗滈崑鐐哄疾濠靛鐓忛柛顐ｇ箥濡牓鏌℃担鍛婃悙闁宠鍨垮畷鎺戭煥鎼达絽濮奸梻浣虹帛閸旀洟鏁冮鍫濊摕鐎广儱顦导鐘绘煕閺囥劌浜愰柛瀣尰缁绘繂顫濋鍌氬Е?
// PUT /api/v1/admin/settings/web-search-emulation
func (h *SettingHandler) UpdateWebSearchEmulationConfig(c *gin.Context) {
	var cfg service.WebSearchEmulationConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.settingService.SaveWebSearchEmulationConfig(c.Request.Context(), &cfg); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Re-read (with sanitized api keys) to return current state
	updated, err := h.settingService.GetWebSearchEmulationConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, service.PopulateWebSearchUsage(c.Request.Context(), updated))
}

// ResetWebSearchUsage 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮村鑸殿€栨繛瀛樼矊缂嶅﹪寮诲☉銏犵疀闁稿繐鎽滈弫鏍⒑濞茶骞楅柟鐟版喘瀵鈽夐姀鈺傛櫇闂佹寧绻傚Λ娑⑺囬妸鈺傗拺缂佸顑欓崕娑樏瑰搴濋偗鐎殿喛顕ч埥澶愬閻橀潧濮堕梻浣告啞閸旓附绂嶉弽顬綁宕奸姀銏紳婵炶揪绲介幖顐㈢毈闂備礁鎲￠〃鍡樼箾婵犲偆鍤曢悹鍥ㄧゴ濡插牓鏌曡箛鏇烆潔闁靛ň鏅滈悡娆撴倵濞戞瑯鐒界紒鐘崇墵閺屾盯鏁愰崶褍濡洪梺?provider 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弴鐐测偓褰掑磿閹寸姵鍠愰柣妤€鐗嗙粭鎺旂磼閳ь剚寰勭仦绋夸壕闁稿繐顦禍楣冩⒑闁偛鑻晶鎾煕閳规儳浜炬俊鐐€栫敮濠囨嚄閸洘鍋傛い鏍仦閻撴洘淇婇妶鍌氫壕闂佸摜鍠庡锟犲箖妤ｅ啯鍊婚柦妯侯樈濞煎﹪姊洪棃娑氬闁瑰弶顭堥ˇ褰掓煛鐏炲墽娲撮柛鈹垮劦瀹曞崬螖閳ь剛绮绘繝姘拺婵炶尪顕ч獮妤併亜閵娿儲鍤囬柛鈺冨仱楠炲鏁傞挊澶夋睏缂傚倸鍊烽悞锕佹懌闂備焦顑欓崣鍐潖閾忓湱纾兼慨妤€妫欓悾鍫曟⒑閹稿孩澶勫ù婊勭矒椤㈡岸鏁愰崶锝呬壕闁革富鍘煎暩闂佹眹鍊愰崑鎾寸節閻㈤潧浠﹂柛銊ョ埣閹兘濮€閵堝懐顔?// POST /api/v1/admin/settings/web-search-emulation/reset-usage
func (h *SettingHandler) ResetWebSearchUsage(c *gin.Context) {
	var req struct {
		ProviderType string `json:"provider_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.ProviderType == "" {
		response.BadRequest(c, "provider_type is required")
		return
	}
	if err := service.ResetWebSearchUsage(c.Request.Context(), req.ProviderType); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, nil)
}

// TestWebSearchEmulation 濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴ｆ閺嬩線鏌熼梻瀵割槮缁炬儳顭烽弻锝夊箛椤掍讲鏋欓梺缁樺笩婵倗鎹㈠☉銏犲耿婵☆垳绮崕搴ㄦ⒑鏉炴壆顦﹂悗姘嵆瀵鈽夐姀鈺傛櫇闂佺粯蓱瑜板啯鎱ㄩ弴銏♀拺?Web Search 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸ゅ嫰鏌涢锝嗙缂佺姷濞€閺岀喖骞戦幇闈涙闁荤喐鐟辩粻鎾诲箖濡ゅ懏鏅查幖绮光偓鎰佹交闂備焦鎮堕崝宥囨崲閸儳宓侀柡宥庣仈鎼搭煈鏁嗛柍褜鍓氭穱濠囨嚃閳哄啯锛?
// POST /api/v1/admin/settings/web-search-emulation/test
func (h *SettingHandler) TestWebSearchEmulation(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Query) == "" {
		req.Query = "AI gateway status"
	}

	result, err := service.TestWebSearch(c.Request.Context(), req.Query)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// ensureDingTalkSyncAttributes 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣椤愯姤鎱ㄥ鍡楀⒒闁绘帊绮欓弻銈嗘叏閹邦兘鍋撻弽顐熷亾濮橆剦鐓奸柡宀嬬秮瀵噣宕掑顑跨帛缂傚倷璁查崑鎾愁熆閼搁潧濮堥柛濠勬暬閺屻劌鈹戦崱娑扁偓妤€顭胯閸ㄦ娊鍩€椤掑喚娼愭繛鍙夌墪椤曪綁宕奸弴鐐电杽闂侀潧顭堥崕娆撴偄閻戞ê鐝伴悗鐟板閸犳鈻撳ú顏呪拻?settings 闂傚倸鍊搁崐鎼佸磹閹间礁纾瑰瀣捣閻棗銆掑锝呬壕濡ょ姷鍋涢ˇ鐢稿极閹剧粯鍋愰柛鎰级閻ゅ嫬鈹戞幊閸娧呭緤娴犲鐤い鎰╁€楅惌鍡涙煕閺囥劌鐏￠柣鎾跺枛閺屻劌鈹戦崱妯烘濠电姭鍋撻柣妤€鐗勬禍婊勩亜閹板墎绋荤紒鈧崘顏嗙＜缂備焦顭囩粻鐐翠繆椤愩垹鏆欓柍钘夘樀瀹曪繝鎮欓幓鎺濇缂?admin 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柟闂寸绾惧綊鏌ｉ幋锝呅撻柛濠傛健閺屻劑寮村Δ鈧禍鎯ь渻閵堝骸骞栭柣蹇旂箚閻忔帡姊洪崗鑲┿偞闁哄懏绻堥幃锛勨偓锝庡枟閳锋垿鏌涢幘鏉戠祷濞存粍绻冪换娑㈠矗婢舵稖鈧法鈧娲橀崹鍧椼€侀弮鍫濋唶闁绘柨鎼獮鍫ユ⒒娴ｈ鐏遍柡鍛洴瀹曨垱瀵奸弶鎴犲帎闂佸搫娲㈤崹娲偂?(attr key, attr name)
// 闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸ゅ嫰鏌涢锝嗙５闁逞屽墾缁犳挸鐣锋總绋款潊闁炽儱鍟跨花銉╂⒒娴ｇ儤鍤€妞ゆ洦鍘介幈銊╁箻椤旂厧鐎┑鐘绘涧閻楀啴宕戦幘璇茬濠㈣泛锕ｆ竟鏇㈡⒒娓氣偓閳ь剛鍋涢懟顖涙櫠閹绢喗鐓?upsert 闂傚倸鍊搁崐鎼佸磹瀹勬噴褰掑炊椤掑鏅悷婊冪Ч濠€渚€姊虹紒妯虹伇婵☆偄瀚板鍛婄瑹閳ь剟寮婚悢鍏尖拻閻庨潧澹婂Σ顔剧磽娴ｅ搫啸濠电偐鍋撻梺鍝勭灱閸犳牠骞冨▎鎿冩晢闁逞屽墴瀹曟繄鎹勭悰鈩冩杸?user attribute definition闂傚倸鍊搁崐鎼佸磹閹间礁纾归柣鎴ｅГ閸婂潡鏌ㄩ弮鍫熸殰闁稿鎸剧划顓炩槈濡娅ч梺娲诲幗閻熲晠寮婚悢鍏煎€绘慨妤€妫欓悾鐑芥⒑缁嬪灝顒㈡い銊ユ婵＄敻宕熼姘棟闂佸壊鐓堥崰鎺楀箰閸愵亞纾藉ù锝呮惈鏍￠梺鍝勮閸旀垿鍨鹃敂鐐磯闁靛绠戠壕顖炴⒑閸涘﹦绠撻悗姘槻鍗遍柣銏犳啞閳锋垹绱掔€ｎ偒鍎ラ柛搴㈠姍閺岀喖宕ㄦ繝鍐ㄢ偓鎰版煕閳瑰灝鍔滅€垫澘瀚换娑㈡倷椤掑倵鍋撻崫鍕垫富闁靛牆妫欑€垫瑩鏌涚仦鍓х煂婵炲娼″缁樻媴閻熼偊鍤嬪┑鐐村絻缁绘ê鐣风憴鍕浄閻庯絻鍔夐崑鎾寸瑹閳ь剙顕ｉ鈧畷鐓庘攽鐎ｎ亝鏆┑鐘垫暩婵炩偓婵炰匠鍥ㄥ亱婵犲﹤鍚橀敐鍥ｅ亾濞戞瑯鐒界紒鈾€鍋撻梻浣圭湽閸ㄨ棄顭囪缁傛帒顭ㄩ崟顏嗙畾濡炪倖鍔х紞鍡椻枔濮椻偓閺岀喖鐛崹顔句紙濡ょ姷鍋涘ú顓㈠箖瑜斿畷鐓庘攽閸愩劋澹曢梺鍓插亝濞叉﹢鎮￠妷锔剧闁瑰浼濋鍫晜闁靛牆顦伴悡鏇㈡倵閿濆骸浜濈€规洖鐭傞弻鐔碱敊閻ｅ本鍣伴梺璇″枤閸忔﹢鐛澶婄闂婎偒鍘奸ˉ姘節瀵伴攱婢橀埀顒佹礋瀹曨垶鍩￠崨顔尖偓鍨攽閻樺弶澶勯柣鎾崇箰閳规垿鎮欓幋婵嗘殭妞ゅ繐鐡ㄧ换娑㈡嚃閳轰焦鐏堥梺鍝勫閸撴繂顕ラ崟顒傜瘈闁告洦鍋掗弳銈夋⒒娴ｅ憡鎯堥柛濠囶棑濞嗐垹顫濋鍌涙闂佺鎻粻鎴犵不婵犳碍鐓涢柛灞久崝婊勩亜椤愩垻孝闁宠鍨块幃娆撳矗婢跺鈧盯姊洪崨濠勬噧閻庢稈鏅濈划瀣吋閸滀胶鍙嗛梺鍓插亞閸犳劕鈻嶉崶顒佲拺闂傚牊绋撶粻鐐烘煕婵炲灝鈧繂顕ｉ锕€纾奸柣鎰嚟閸樺崬鈹戦悙鏉戠仴鐎规洦鍓熷绋库槈濞嗗秳绨?name 婵犵數濮烽弫鍛婃叏閻戣棄鏋侀柟闂寸绾惧鏌ｉ幇顒佹儓闁搞劌鍊块弻娑㈩敃閿濆棛顦ョ紓浣哄С閸楁娊寮诲☉妯锋斀闁告洦鍋勬慨銏ゆ⒑濞茶骞楅柟鐟版喘瀵鎮㈤搹鍦紲闂侀潧绻掓慨鐢告倶閸垻纾藉ù锝呮惈鏍￠梺缁橆殘婵炩偓鐎殿喖顭烽幃銏ゆ偂鎼达綆鍞归梻渚€鈧稑宓嗘繛浣冲啠鏋旀い鎾跺Х绾捐棄霉閿濆棗绲诲ù婊堢畺濮婅櫣鍖栭弴鐐测拤濡炪値鍘煎ú銊у垝缂佹ǜ鍋呴柛鎰ㄦ櫇閸橀亶姊洪弬銉︽珔闁告﹢绠栭、鏃堝川椤旂厧浼庨梻浣稿閸嬩線宕归懡銈咁棜闁兼祴鏅濈壕钘壝归敐澶嬫锭闁诲繆鏅犻弻锝堢疀閵壯呮殼闂?name
// ensureDingTalkSyncAttributes creates required DingTalk sync attributes.
func (h *SettingHandler) ensureDingTalkSyncAttributes(ctx context.Context, settings *service.SystemSettings) {
	if h.userAttributeService == nil || settings == nil {
		return
	}
	if settings.DingTalkConnectCorpRestrictionPolicy != "internal_only" {
		return
	}
	if settings.DingTalkConnectSyncDisplayName {
		h.ensureUserAttributeDefinition(ctx, settings.DingTalkConnectSyncDisplayNameAttrKey, settings.DingTalkConnectSyncDisplayNameAttrName, "DingTalk synced display name", service.AttributeTypeText)
	}
	if settings.DingTalkConnectSyncCorpEmail {
		h.ensureUserAttributeDefinition(ctx, settings.DingTalkConnectSyncCorpEmailAttrKey, settings.DingTalkConnectSyncCorpEmailAttrName, "DingTalk synced corporate email", service.AttributeTypeEmail)
	}
	if settings.DingTalkConnectSyncDept {
		h.ensureUserAttributeDefinition(ctx, settings.DingTalkConnectSyncDeptAttrKey, settings.DingTalkConnectSyncDeptAttrName, "DingTalk synced department", service.AttributeTypeText)
	}
}

func (h *SettingHandler) ensureUserAttributeDefinition(ctx context.Context, key, name, description string, attrType service.UserAttributeType) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	existing, err := h.userAttributeService.GetDefinitionByKey(ctx, key)
	if err == nil && existing != nil {
		if strings.TrimSpace(name) != "" && existing.Name != name {
			if _, err := h.userAttributeService.UpdateDefinition(ctx, existing.ID, service.UpdateAttributeDefinitionInput{
				Name: &name,
			}); err != nil {
				slog.Warn("dingtalk: update user attribute definition name failed", "key", key, "err", err.Error())
				return
			}
			slog.Info("dingtalk: updated user attribute definition name", "key", key, "name", name)
		}
		return
	}
	if _, err := h.userAttributeService.CreateDefinition(ctx, service.CreateAttributeDefinitionInput{
		Key:         key,
		Name:        name,
		Description: description,
		Type:        attrType,
		Enabled:     true,
	}); err != nil {
		slog.Warn("dingtalk: ensure user attribute definition failed", "key", key, "err", err.Error())
		return
	}
	slog.Info("dingtalk: created user attribute definition", "key", key, "name", name, "type", attrType)
}

// ListEmailTemplates returns all editable notification email templates.
// GET /api/v1/admin/settings/email-templates
func (h *SettingHandler) ListEmailTemplates(c *gin.Context) {
	if h.notificationEmailService == nil {
		response.InternalError(c, "notification email service is not configured")
		return
	}
	events := h.notificationEmailService.ListEventInfos()
	templates, err := h.notificationEmailService.ListTemplates(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.EmailTemplateListResponse{
		Events:       emailTemplateEventOptionsToDTO(events),
		Locales:      h.notificationEmailService.SupportedLocales(),
		Templates:    emailTemplateSummariesToDTO(templates),
		Placeholders: emailTemplatePlaceholderUnion(events),
	})
}

// GetEmailTemplate returns one editable notification email template.
// GET /api/v1/admin/settings/email-templates/:event/:locale
func (h *SettingHandler) GetEmailTemplate(c *gin.Context) {
	if h.notificationEmailService == nil {
		response.InternalError(c, "notification email service is not configured")
		return
	}
	tmpl, err := h.notificationEmailService.GetTemplate(c.Request.Context(), c.Param("event"), c.Param("locale"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, emailTemplateDetailToDTO(tmpl))
}

// UpdateEmailTemplate saves an override for one event/locale template.
// PUT /api/v1/admin/settings/email-templates/:event/:locale
func (h *SettingHandler) UpdateEmailTemplate(c *gin.Context) {
	if h.notificationEmailService == nil {
		response.InternalError(c, "notification email service is not configured")
		return
	}
	var req dto.UpdateEmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	tmpl, err := h.notificationEmailService.UpdateTemplate(c.Request.Context(), c.Param("event"), c.Param("locale"), req.Subject, req.HTML)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, emailTemplateDetailToDTO(tmpl))
}

// RestoreOfficialEmailTemplate removes an override and returns the built-in template.
// POST /api/v1/admin/settings/email-templates/:event/:locale/restore-official
func (h *SettingHandler) RestoreOfficialEmailTemplate(c *gin.Context) {
	if h.notificationEmailService == nil {
		response.InternalError(c, "notification email service is not configured")
		return
	}
	tmpl, err := h.notificationEmailService.RestoreOfficialTemplate(c.Request.Context(), c.Param("event"), c.Param("locale"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, emailTemplateDetailToDTO(tmpl))
}

// PreviewEmailTemplate renders a template with safe sample variables without saving it.
// POST /api/v1/admin/settings/email-templates/preview
func (h *SettingHandler) PreviewEmailTemplate(c *gin.Context) {
	if h.notificationEmailService == nil {
		response.InternalError(c, "notification email service is not configured")
		return
	}
	var req dto.PreviewEmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	preview, err := h.notificationEmailService.PreviewTemplate(c.Request.Context(), service.NotificationEmailPreviewInput{
		Event:     req.Event,
		Locale:    req.Locale,
		Subject:   req.Subject,
		HTML:      req.HTML,
		Variables: req.Variables,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, dto.EmailTemplatePreviewResponse{Subject: preview.Subject, HTML: preview.HTML})
}

func emailTemplateEventOptionsToDTO(events []service.NotificationEmailEventInfo) []dto.EmailTemplateEventOption {
	items := make([]dto.EmailTemplateEventOption, 0, len(events))
	for _, event := range events {
		items = append(items, dto.EmailTemplateEventOption{
			Value:       event.Event,
			Label:       event.Label,
			Description: event.Description,
			Category:    event.Category,
			Optional:    event.Optional,
		})
	}
	return items
}

func emailTemplateSummariesToDTO(templates []service.NotificationEmailTemplate) []dto.EmailTemplateSummary {
	items := make([]dto.EmailTemplateSummary, 0, len(templates))
	for _, tmpl := range templates {
		items = append(items, dto.EmailTemplateSummary{
			Event:     tmpl.Event,
			Locale:    tmpl.Locale,
			Subject:   tmpl.Subject,
			IsCustom:  tmpl.IsCustom,
			UpdatedAt: emailTemplateUpdatedAt(tmpl),
		})
	}
	return items
}

func emailTemplateDetailToDTO(tmpl service.NotificationEmailTemplate) dto.EmailTemplateDetail {
	return dto.EmailTemplateDetail{
		Event:        tmpl.Event,
		Locale:       tmpl.Locale,
		Subject:      tmpl.Subject,
		HTML:         tmpl.HTML,
		IsCustom:     tmpl.IsCustom,
		UpdatedAt:    emailTemplateUpdatedAt(tmpl),
		Placeholders: tmpl.Placeholders,
	}
}

func emailTemplateUpdatedAt(tmpl service.NotificationEmailTemplate) string {
	if tmpl.UpdatedAt == nil {
		return ""
	}
	return tmpl.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
}

func emailTemplatePlaceholderUnion(events []service.NotificationEmailEventInfo) []string {
	seen := make(map[string]struct{})
	placeholders := make([]string, 0)
	for _, event := range events {
		for _, placeholder := range event.Placeholders {
			if _, ok := seen[placeholder]; ok {
				continue
			}
			seen[placeholder] = struct{}{}
			placeholders = append(placeholders, placeholder)
		}
	}
	return placeholders
}

// equalNullableFloat compares two *float64 values treating nil as a distinct case.
func equalNullableFloat(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// slotOf returns the *float64 for the given window from a DefaultPlatformQuotaSetting.
func slotOf(s *service.DefaultPlatformQuotaSetting, win string) *float64 {
	if s == nil {
		return nil
	}
	switch win {
	case "daily":
		return s.DailyLimitUSD
	case "weekly":
		return s.WeeklyLimitUSD
	case "monthly":
		return s.MonthlyLimitUSD
	}
	return nil
}

// equalPlatformQuotaSettings reports whether two platform-quota maps are identical across all 12 slots.
func equalPlatformQuotaSettings(before, after map[string]*service.DefaultPlatformQuotaSetting) bool {
	for _, platform := range service.AllowedQuotaPlatforms {
		b := before[platform]
		a := after[platform]
		if !equalNullableFloat(slotOf(b, "daily"), slotOf(a, "daily")) {
			return false
		}
		if !equalNullableFloat(slotOf(b, "weekly"), slotOf(a, "weekly")) {
			return false
		}
		if !equalNullableFloat(slotOf(b, "monthly"), slotOf(a, "monthly")) {
			return false
		}
	}
	return true
}

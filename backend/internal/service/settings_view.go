package service

import "strings"

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

type SystemSettings struct {
	RegistrationEnabled                          bool
	EmailVerifyEnabled                           bool
	RegistrationEmailSuffixWhitelist             []string
	PromoCodeEnabled                             bool
	PasswordResetEnabled                         bool
	FrontendURL                                  string
	InvitationCodeEnabled                        bool
	TotpEnabled                                  bool
	LoginAgreementEnabled                        bool
	ConversationAuditSecondaryPasswordHash       string
	ConversationAuditSecondaryPasswordConfigured bool
	ConversationAuditCleanupEnabled              bool
	ConversationAuditRetentionDays               int
	LoginAgreementMode                           string
	LoginAgreementUpdatedAt                      string
	LoginAgreementDocuments                      []LoginAgreementDocument

	SMTPHost               string
	SMTPPort               int
	SMTPUsername           string
	SMTPPassword           string
	SMTPPasswordConfigured bool
	SMTPFrom               string
	SMTPFromName           string
	SMTPUseTLS             bool

	TurnstileEnabled             bool
	TurnstileSiteKey             string
	TurnstileSecretKey           string
	TurnstileSecretKeyConfigured bool
	APIKeyACLTrustForwardedIP    bool

	// LinuxDo Connect OAuth 闂傚倸鍊峰ù鍥儍椤愶箑骞㈤柍杞扮劍椤斿嫮绱?
	LinuxDoConnectEnabled                bool
	LinuxDoConnectClientID               string
	LinuxDoConnectClientSecret           string
	LinuxDoConnectClientSecretConfigured bool
	LinuxDoConnectRedirectURL            string

	// DingTalk Connect OAuth 闂傚倸鍊峰ù鍥儍椤愶箑骞㈤柍杞扮劍椤斿嫮绱?
	DingTalkConnectEnabled                 bool
	DingTalkConnectClientID                string
	DingTalkConnectClientSecret            string
	DingTalkConnectClientSecretConfigured  bool
	DingTalkConnectRedirectURL             string
	DingTalkConnectCorpRestrictionPolicy   string
	DingTalkConnectInternalCorpID          string
	DingTalkConnectBypassRegistration      bool
	DingTalkConnectSyncCorpEmail           bool
	DingTalkConnectSyncDisplayName         bool
	DingTalkConnectSyncDept                bool
	DingTalkConnectSyncCorpEmailAttrKey    string
	DingTalkConnectSyncDisplayNameAttrKey  string
	DingTalkConnectSyncDeptAttrKey         string
	DingTalkConnectSyncCorpEmailAttrName   string
	DingTalkConnectSyncDisplayNameAttrName string
	DingTalkConnectSyncDeptAttrName        string

	// WeChat Connect OAuth 闂傚倸鍊峰ù鍥儍椤愶箑骞㈤柍杞扮劍椤斿嫮绱?
	WeChatConnectEnabled                   bool
	WeChatConnectAppID                     string
	WeChatConnectAppSecret                 string
	WeChatConnectAppSecretConfigured       bool
	WeChatConnectOpenAppID                 string
	WeChatConnectOpenAppSecret             string
	WeChatConnectOpenAppSecretConfigured   bool
	WeChatConnectMPAppID                   string
	WeChatConnectMPAppSecret               string
	WeChatConnectMPAppSecretConfigured     bool
	WeChatConnectMobileAppID               string
	WeChatConnectMobileAppSecret           string
	WeChatConnectMobileAppSecretConfigured bool
	WeChatConnectOpenEnabled               bool
	WeChatConnectMPEnabled                 bool
	WeChatConnectMobileEnabled             bool
	WeChatConnectMode                      string
	WeChatConnectScopes                    string
	WeChatConnectRedirectURL               string
	WeChatConnectFrontendRedirectURL       string

	// Generic OIDC OAuth 闂傚倸鍊峰ù鍥儍椤愶箑骞㈤柍杞扮劍椤斿嫮绱?
	OIDCConnectEnabled                bool
	OIDCConnectProviderName           string
	OIDCConnectClientID               string
	OIDCConnectClientSecret           string
	OIDCConnectClientSecretConfigured bool
	OIDCConnectIssuerURL              string
	OIDCConnectDiscoveryURL           string
	OIDCConnectAuthorizeURL           string
	OIDCConnectTokenURL               string
	OIDCConnectUserInfoURL            string
	OIDCConnectJWKSURL                string
	OIDCConnectScopes                 string
	OIDCConnectRedirectURL            string
	OIDCConnectFrontendRedirectURL    string
	OIDCConnectTokenAuthMethod        string
	OIDCConnectUsePKCE                bool
	OIDCConnectValidateIDToken        bool
	OIDCConnectAllowedSigningAlgs     string
	OIDCConnectClockSkewSeconds       int
	OIDCConnectRequireEmailVerified   bool
	OIDCConnectUserInfoEmailPath      string
	OIDCConnectUserInfoIDPath         string
	OIDCConnectUserInfoUsernamePath   string

	// GitHub / Google 闂傚倸鍊搁崐椋庢閿熺姴鐭楅幖娣妼缁愭鎱ㄥ鈧·鍌炲极婵犲洦鐓曟い鎰靛墰閹ジ鏌ｉ妶鍌氫壕濠碉紕鍋戦崐鏍礉瑜忓濠囨偩瀹€鈧惌鍫㈡喐閻楀牆绗氶柣鎾跺枔閹叉悂骞嬮敂缁樻櫔閻熸粌绻掑Σ?
	GitHubOAuthEnabled                bool
	GitHubOAuthClientID               string
	GitHubOAuthClientSecret           string
	GitHubOAuthClientSecretConfigured bool
	GitHubOAuthRedirectURL            string
	GitHubOAuthFrontendRedirectURL    string
	GoogleOAuthEnabled                bool
	GoogleOAuthClientID               string
	GoogleOAuthClientSecret           string
	GoogleOAuthClientSecretConfigured bool
	GoogleOAuthRedirectURL            string
	GoogleOAuthFrontendRedirectURL    string

	SiteName                    string
	SiteLogo                    string
	SiteSubtitle                string
	APIBaseURL                  string
	ContactInfo                 string
	DocURL                      string
	HomeContent                 string
	HideCcsImportButton         bool
	PurchaseSubscriptionEnabled bool
	PurchaseSubscriptionURL     string
	TableDefaultPageSize        int
	TablePageSizeOptions        []int
	CustomMenuItems             string // JSON array of custom menu items
	CustomEndpoints             string // JSON array of custom endpoints

	DefaultConcurrency           int
	DefaultBalance               float64
	RiskControlEnabled           bool
	CyberSessionBlockEnabled     bool
	CyberSessionBlockTTLSeconds  int
	AffiliateEnabled             bool
	AffiliateRebateRate          float64
	AffiliateRebateFreezeHours   int
	AffiliateRebateDurationDays  int
	AffiliateRebatePerInviteeCap float64
	DefaultUserRPMLimit          int
	DefaultSubscriptions         []DefaultSubscriptionSetting

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
	OpsMonitoringEnabled         bool
	OpsRealtimeMonitoringEnabled bool
	OpsQueryModeDefault          string
	OpsMetricsIntervalSeconds    int

	// Channel Monitor feature
	ChannelMonitorEnabled                bool `json:"channel_monitor_enabled"`
	ChannelMonitorDefaultIntervalSeconds int  `json:"channel_monitor_default_interval_seconds"`

	// Available Channels feature (user-facing aggregate view)
	AvailableChannelsEnabled bool `json:"available_channels_enabled"`

	// Claude Code version check
	MinClaudeCodeVersion string
	MaxClaudeCodeVersion string

	AllowUngroupedKeyScheduling bool
	BackendModeEnabled          bool

	// Gateway forwarding behavior
	EnableFingerprintUnification           bool
	EnableMetadataPassthrough              bool
	EnableCCHSigning                       bool
	EnableClaudeOAuthSystemPromptInjection bool
	ClaudeOAuthSystemPrompt                string
	ClaudeOAuthSystemPromptBlocks          string
	EnableAnthropicCacheTTL1hInjection     bool
	RewriteMessageCacheControl             bool
	AntigravityUserAgentVersion            string
	OpenAICodexUserAgent                   string
	MinCodexVersion                        string
	MaxCodexVersion                        string
	CodexCLIOnlyBlacklist                  string
	CodexCLIOnlyWhitelist                  string
	CodexCLIOnlyAllowAppServerClients      bool
	CodexCLIOnlyEngineFingerprintSignals   string
	OpenAIAllowClaudeCodeCodexPlugin       bool

	// Web Search Emulation
	WebSearchEmulationEnabled bool
	// Payment visible method routing
	PaymentVisibleMethodAlipaySource  string
	PaymentVisibleMethodWxpaySource   string
	PaymentVisibleMethodAlipayEnabled bool
	PaymentVisibleMethodWxpayEnabled  bool

	// OpenAI 闂傚倷娴囧畷鍨叏閻㈢绀夌憸蹇曞垝婵犳艾绠ｉ柨婵嗗暕濮规鏌ｉ悩鍏呰埅闁告柨绉堕埀顒佸嚬閸撶喖寮诲☉銏犵疀闁宠桨绀侀‖鍫濐渻?
	OpenAIAdvancedSchedulerEnabled bool

	// 濠电姷鏁搁崑鐘诲箵椤忓棗绶ゅΔ锝呭暞閸嬶繝鏌嶉崫鍕仺闁绘垼妫勭壕鍏肩箾閹寸偛鐒归柡瀣墵閹鐛崹顔煎闂佺粯顨呭Λ婵嗙暦娴兼潙鍗抽柕蹇ョ磿閸樼數绱撻崒娆戝妽闁挎艾鈹戦鑲┬ч柡?
	BalanceLowNotifyEnabled     bool
	BalanceLowNotifyThreshold   float64
	BalanceLowNotifyRechargeURL string

	// 闂傚倷娴囧畷鍨叏閹惰姤鈷旂€广儱顦崹鍌炴煢濡尨绱氶柨婵嗩槸缁€瀣亜閺嶃劍鐨戞い锔诲亰濮婅櫣绱掑鍡樼暥闂佺粯顨呴敃顏勭暦濞差亜纾奸柣鎰嚟閸樼數绱撻崒娆戝妽闁挎艾鈹戦鑲┬ч柡?
	SubscriptionExpiryNotifyEnabled bool

	// 闂傚倷娴囧畷鍨叏閻㈢绀夌憸蹇曞垝婵犳艾绠ｉ柨婵嗗暕濮规姊虹粔鍡楀濞堟洟鏌嶉柨瀣棃闁哄瞼鍠愮€靛ジ寮堕幋婊冩瀳闂備礁鎼ˇ鏉款渻閽樺娼栨繛宸憾閺佸啴鏌ㄥ┑鍡橆棤婵炲牄鍊濆?
	AccountQuotaNotifyEnabled bool
	AccountQuotaNotifyEmails  []NotifyEmailEntry

	// 缂傚倸鍊搁崐椋庢閿熺姴鍨傞梻鍫熺〒閺嗭箓鏌ｉ姀銈嗘锭闁搞劍绻冪换娑橆啅椤旇崵鐩庣紓浣插亾闁稿瞼鍋為悡蹇撯攽閻愯尙浠㈤柛鏂诲€栫换婵嬪焵椤掑嫬绾ф繛鍡欏亾閸嬨儱鈹戦悙鏉戠仸闁荤噥鍨堕幃鐐烘倷椤戝彞绨婚梺鍦亾閸撴岸鎮℃總鍛婄厽婵炴垵宕▍宥団偓瑙勬礃閿曘垽鐛幘璇茬婵°倐鍋撶憸鐗堢懇濮婂宕掑▎鎴濆濠电姰鍨洪…鍥╁垝閺冨洢浜归柟瑙勭墱閸忔ê鐣烽幒妤佸€烽悗娑櫳戦悵顐︽⒒娴ｄ警娼掗柛鏇炵仛缁傚瘔 = platform闂傚倸鍊烽悞锔锯偓绗涘懐鐭欓柟杈惧瘜閺佸﹪鏌涢悙鎻掑瀭/缂傚倸鍊搁崐鎼佸磹閹间礁纾归柦妯侯槴閺嬪秹鏌ㄥ┑鍡樺櫡闁?= 濠电姷鏁搁崑鐐哄垂閸洖绠伴柛婵勫劤閻捇鏌ｉ悢鐓庝喊闁诲氦鍩栫换娑橆啅椤旇崵鐩庣紓浣插亾闁糕剝绋掗崑锝吤归敐鍛暈閻犳劧绱曠槐鎺楀焵?
	DefaultPlatformQuotas map[string]*DefaultPlatformQuotaSetting `json:"default_platform_quotas"`

	AllowUserViewErrorRequests bool
}

type DefaultSubscriptionSetting struct {
	GroupID      int64 `json:"group_id"`
	ValidityDays int   `json:"validity_days"`
}

type PublicSettings struct {
	RegistrationEnabled              bool
	EmailVerifyEnabled               bool
	ForceEmailOnThirdPartySignup     bool
	RegistrationEmailSuffixWhitelist []string
	PromoCodeEnabled                 bool
	PasswordResetEnabled             bool
	InvitationCodeEnabled            bool
	TotpEnabled                      bool
	LoginAgreementEnabled            bool
	LoginAgreementMode               string
	LoginAgreementUpdatedAt          string
	LoginAgreementRevision           string
	LoginAgreementDocuments          []LoginAgreementDocument
	TurnstileEnabled                 bool
	TurnstileSiteKey                 string
	SiteName                         string
	SiteLogo                         string
	SiteSubtitle                     string
	APIBaseURL                       string
	ContactInfo                      string
	DocURL                           string
	HomeContent                      string
	HideCcsImportButton              bool

	PurchaseSubscriptionEnabled bool
	PurchaseSubscriptionURL     string
	TableDefaultPageSize        int
	TablePageSizeOptions        []int
	CustomMenuItems             string // JSON array of custom menu items
	CustomEndpoints             string // JSON array of custom endpoints

	LinuxDoOAuthEnabled      bool
	DingTalkOAuthEnabled     bool
	WeChatOAuthEnabled       bool
	WeChatOAuthOpenEnabled   bool
	WeChatOAuthMPEnabled     bool
	WeChatOAuthMobileEnabled bool
	BackendModeEnabled       bool
	PaymentEnabled           bool
	OIDCOAuthEnabled         bool
	OIDCOAuthProviderName    string
	GitHubOAuthEnabled       bool
	GoogleOAuthEnabled       bool
	Version                  string

	BalanceLowNotifyEnabled     bool
	AccountQuotaNotifyEnabled   bool
	BalanceLowNotifyThreshold   float64
	BalanceLowNotifyRechargeURL string

	// Channel Monitor feature
	ChannelMonitorEnabled                bool `json:"channel_monitor_enabled"`
	ChannelMonitorDefaultIntervalSeconds int  `json:"channel_monitor_default_interval_seconds"`

	// Available Channels feature (user-facing aggregate view)
	AvailableChannelsEnabled bool `json:"available_channels_enabled"`

	// Affiliate (闂傚倸鍊搁崐椋庢閿熺姴鍌ㄩ柟闂寸绾惧鏌熼崜褏甯涢柛瀣剁秮閺屾盯濡烽姀鈩冪彆闂佺粯甯熷▔鏇犳閹烘梻纾兼俊顖濇娴煎矂鎮? feature toggle
	AffiliateEnabled bool `json:"affiliate_enabled"`

	RiskControlEnabled bool `json:"risk_control_enabled"`

	AllowUserViewErrorRequests bool `json:"allow_user_view_error_requests"`
}

type LoginAgreementDocument struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	ContentMD string `json:"content_md"`
}

type WeChatConnectOAuthConfig struct {
	Enabled             bool
	LegacyAppID         string
	LegacyAppSecret     string
	OpenAppID           string
	OpenAppSecret       string
	MPAppID             string
	MPAppSecret         string
	MobileAppID         string
	MobileAppSecret     string
	OpenEnabled         bool
	MPEnabled           bool
	MobileEnabled       bool
	Mode                string
	Scopes              string
	RedirectURL         string
	FrontendRedirectURL string
}

func (cfg WeChatConnectOAuthConfig) SupportsMode(mode string) bool {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		return cfg.MPEnabled
	case "mobile":
		return cfg.MobileEnabled
	default:
		return cfg.OpenEnabled
	}
}

func (cfg WeChatConnectOAuthConfig) ScopeForMode(mode string) string {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		return normalizeWeChatConnectScopeSetting(cfg.Scopes, "mp")
	case "mobile":
		return ""
	}
	return defaultWeChatConnectScopeForMode("open")
}

func (cfg WeChatConnectOAuthConfig) AppIDForMode(mode string) string {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		return strings.TrimSpace(firstNonEmpty(cfg.MPAppID, cfg.LegacyAppID))
	case "mobile":
		return strings.TrimSpace(firstNonEmpty(cfg.MobileAppID, cfg.LegacyAppID))
	}
	return strings.TrimSpace(firstNonEmpty(cfg.OpenAppID, cfg.LegacyAppID))
}

func (cfg WeChatConnectOAuthConfig) AppSecretForMode(mode string) string {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		return strings.TrimSpace(firstNonEmpty(cfg.MPAppSecret, cfg.LegacyAppSecret))
	case "mobile":
		return strings.TrimSpace(firstNonEmpty(cfg.MobileAppSecret, cfg.LegacyAppSecret))
	}
	return strings.TrimSpace(firstNonEmpty(cfg.OpenAppSecret, cfg.LegacyAppSecret))
}

// StreamTimeoutSettings 婵犵數濮烽弫鎼佸磻閻旂儤宕叉繝闈涙灩瑜版帒鐐婇柍鍝勫暟閻撳姊洪崷顓℃闁哥姵鐗滈悮鎯ь吋婢跺浠柡澶屽仦婵粙顢楅悢闀愮箚闁哄被鍎伴幉楣冩煛瀹€瀣М闁诡喗鐟╁畷锝嗗緞濡紮绱楅梻鍌欑閹猜ゅ綘闂佺锕ラ〃鍫澪ｉ幇鏉跨睄闁割偒鍋呴弲顏堟⒑缁洖澧查柣鐔濆泚鍥煛閸屾粎顔曢柣搴ｆ暩椤牆鐡梻浣烘嚀閸熻法鎹㈤幒妤€鏋佹い鏇楀亾鐎规洖銈稿鎾倷濞堟寧孝闂傚倷娴囬鏍垂鏉堛劎鐝堕柛鈩冾殘閹姐儵姊婚崒娆戭槮闁圭⒈鍋婂畷顖炲礃濞村鐏冨┑鐐村灟閸ㄥ湱绮婚弽銊ょ箚闁靛牆鎳忛崳娲煕鎼达紕效闁哄本鐩鎾Ω閵夈倛鍩呮繝纰樺墲閻撯剝绂嶉崼鏇炶摕闁挎繂顦伴崐鐑芥煕濞嗗浚妯堟俊顐節濮婃椽宕崟顓犱紘闂佸摜濮撮柊锝夈€佸鑸垫櫜闁割偁鍨婚弶鎼佹⒑闂堟冻绱￠柛婊€鐒﹀暩闂傚倷鑳堕崕鐢稿礈濠靛牊鏆滈柟鐑橆殔缁犵娀鐓崶銊︽儎婵炴挸顭烽弻娑樼暆閳ь剟宕戝☉姘辩焼闁糕剝绋掗悡蹇涙煕椤愶絿绠栭柛锝勭矙閺岋繝宕遍弴鐐茬ギ闂佸搫鐭夌紞浣规叏閳ь剟鏌ㄥ┑鍡楊仼闁诡垰鐗忕槐鎾诲磼濮樻瘷銏°亜閿曞倹娑ч柣锝呭槻铻栭柛娑卞幘椤斿懘姊洪懞銉冾亪藝椤栨粍灏庨悗锝庡墰绾句粙鏌涚仦鎹愬闁哄鍊濋弻娑㈡偐閹颁焦鐣奸梺杞扮贰閸ｏ絽鐣锋總绋垮嵆闁绘劖顔栧Σ鎾⒑閼姐倕小缂佲偓娴ｉ潻鑰块弶鍫氭櫇閻?
type StreamTimeoutSettings struct {
	Enabled                bool   `json:"enabled"`
	Action                 string `json:"action"`
	TempUnschedMinutes     int    `json:"temp_unsched_minutes"`
	ThresholdCount         int    `json:"threshold_count"`
	ThresholdWindowMinutes int    `json:"threshold_window_minutes"`
}

// StreamTimeoutAction 婵犵數濮烽弫鎼佸磻閻旂儤宕叉繝闈涙灩瑜版帒鐐婇柍鍝勫暟閻撳姊洪崷顓℃闁哥姵鐗滈悮鎯ь吋婢跺浠柡澶屽仦婵粙顢楅悢闀愮箚闁哄被鍎伴幉楣冩煛瀹€瀣М闁诡喗鐟╁畷锝嗗緞濡紮绱掗梻鍌欑閹碱偊鎯夋總绋跨獥閹兼惌鐓堥弫瀣煥濠靛棛澧㈤柣銈傚亾闂備浇顫夊畷妯衡枖濞戙垺鍋傞柟杈鹃檮閳?const (
const (
	StreamTimeoutActionTempUnsched = "temp_unsched" // 濠电姷鏁搁崑鐐哄垂閸洖绠伴柟闂寸劍閸嬨倝鏌曟繛鍨姶婵炴挸顭烽弻娑氫沪閸撗佲偓鎺楁煃缂佹ɑ鈷掗柍褜鍓欑粻宥夊磿闁秴绠犻幖杈剧稻婵ジ鏌＄仦璇插姕闁稿缍侀幃妤呮晲鎼存繄鐩庨梺鍝勬閻熲晠骞?
	StreamTimeoutActionError       = "error"        // 闂傚倸鍊风粈渚€骞栭銈囩煋闁哄鍤氬ú顏勭厸闁告粈鐒﹂弲鈺呮⒑閹肩偛鍔橀柛鏂块叄閸┿垽寮埀顒傛崲濞戞﹩鍟呮い鏃囧吹閸戝綊姊洪崫鍕櫝闁哄懐濞€楠炲啳銇愰幒鎴犲€為柣鐘荤細濞咃絾绂掗娑氱瘈婵炲牆鐏濋弸鐔兼煙椤旂厧鈧潡鎮伴鈧崺鈧?	StreamTimeoutActionNone        = "none"         // 濠电姷鏁搁崑鐐哄垂閸洖绠伴柛婵勫劤閻捇鏌熺紒銏犳殙闁搞儺鍓欓悡娑㈡煕濞戝崬鏋涘ù?)
	StreamTimeoutActionNone        = "none"
)

// DefaultStreamTimeoutSettings 闂傚倷绀侀幖顐λ囬锕€鐤炬繝濠傜墕閽冪喖鏌曟繛鍨壄婵炲樊浜滈崘鈧銈庡幗閸ㄨ埖绂掑Ο琛℃斀闁宠棄妫楅悘鈩冦亜閹寸偟鎳囩€规洏鍎靛畷鍫曨敆娴ｅ搫骞楅梻渚€娼ч…鍫ュ磿闁稁鏁傛い蹇撴绾剧厧銆掑顒婂伐妞も晩鍓熼弻鐔哥附閸濄儳鏆悗瑙勬磸閸旀垿銆佸▎鎴犻┏閻庯綆鍊ｅ┑瀣拻闁稿本鐟ч崝宥嗙節閵忊槅鐒界紒顔芥瀹曞ジ寮撮悢铚傜棯?
func DefaultStreamTimeoutSettings() *StreamTimeoutSettings {
	return &StreamTimeoutSettings{
		Enabled:                false,
		Action:                 StreamTimeoutActionTempUnsched,
		TempUnschedMinutes:     5,
		ThresholdCount:         3,
		ThresholdWindowMinutes: 10,
	}
}

// RectifierSettings controls request rectification.
type RectifierSettings struct {
	Enabled                  bool     `json:"enabled"`
	ThinkingSignatureEnabled bool     `json:"thinking_signature_enabled"`
	ThinkingBudgetEnabled    bool     `json:"thinking_budget_enabled"`
	APIKeySignatureEnabled   bool     `json:"apikey_signature_enabled"`
	APIKeySignaturePatterns  []string `json:"apikey_signature_patterns"`
}

// DefaultRectifierSettings 闂傚倷绀侀幖顐λ囬锕€鐤炬繝濠傜墕閽冪喖鏌曟繛鍨壄婵炲樊浜滈崘鈧銈庡幗閸ㄨ埖绂掑Ο琛℃斀闁宠棄妫楅悘鈩冦亜閹寸偟鎳囩€规洏鍎靛畷鍫曨敆娴ｅ搫骞楅梻渚€娼ч…鍫ュ磿闁稁鏁傛い蹇撶墛閻撴盯鏌涢埥鍡楀箻闁告帊鍗冲Λ浣瑰緞婵炴帒缍婇弫鎰板川椤斿吋娈橀梻浣虹帛缁哄潡宕愬┑鍡╂綎婵炲樊浜滅痪褔鎮归幁鎺戝闁靛牆顦伴悡鐔兼煏閸繃鍣烘い銉ｅ灲閺屸剝鎷呴崫銉愌呪偓瑙勬礈閸樠囧煘閹达箑骞㈡俊鐐插⒔缁€瀣⒒閸屾艾鈧娆㈤敓鐘茬獥闁规崘顕х粈澶屸偓鍏夊亾闁告洦鍋嗛悾楣冩煙閸忚偐鏆橀柛鏂跨Ч瀹曟垿濡搁敂杞扮盎濡炪倖鎸鹃崰搴ｇ箔瑜忕槐鎺楀焵?
func DefaultRectifierSettings() *RectifierSettings {
	return &RectifierSettings{
		Enabled:                  true,
		ThinkingSignatureEnabled: true,
		ThinkingBudgetEnabled:    true,
	}
}

// Beta Policy 缂傚倸鍊搁崐鐑芥倿閿斿墽鐭欓柟鐑橆殕閸嬪绻濇繝鍌滃婵鐓￠弻鐔兼焽閿曗偓閺嬬喖鏌￠崱妯肩煉闁哄瞼鍠栧鑽も偓闈涘濡差喖鈹?
const (
	BetaPolicyActionPass   = "pass" // 闂傚倸鍊搁崐椋庢閿熺姴纾婚柛鏇ㄥ幗瀹曟煡鎮楅敐搴″妞ゎ偅娲熼弻娑㈩敃閿濆棛顦ㄩ梺鍝勬媼閸撶喖骞冨鈧幃娆戞崉鏉炵増鐫忛梻浣藉吹閸犳劗鎹㈤崼銉ヨ摕婵炴垶鍩冮崑鎾绘晲鎼粹€茬敖闂佺粯鎼╅崹鍫曞蓟閻旇偐宓侀柛顭戝枤娴犻箖姊洪悷鏉挎Щ闁硅櫕锚椤曪綁骞橀钘夆偓濠氭煕閳╁喚鐒洪柣鏂垮悑閳?	BetaPolicyActionFilter = "filter" // 闂傚倷绀侀幖顐λ囬锕€鐤炬繛鎴炩棨濞差亝鏅濋柛灞炬皑椤斿懘姊洪棃娑氬妞わ缚鍗冲鏌ヮ敆閸曨剙鈧爼鏌ｉ幇顖涚【濞存粌缍婇弻?beta header 濠电姷鏁搁崑鐐哄垂閸洖绠归柍鍝勬噹閸屻劑鏌涘▎宥呭姢闁哄嫬鍊垮濠氬磼濮橆兘鍋撻崫銉㈠亾濮樸儱濮傜€规洘鍔曢埥澶愬閳哄倹娅?token
	BetaPolicyActionFilter = "filter"
	BetaPolicyActionBlock  = "block" // 闂傚倸鍊烽懗鍫曞箠閹捐鍚归柡宥庡幗閳锋棃鏌涢弴銊ョ仩闁哄嫨鍎甸弻娑樷槈濡吋鎲奸梺鍝勬媼閸撴盯鍩€椤掆偓閸樻粓宕戦幘缁樼厓鐟滄粓宕滃▎鎴濆疾闂備胶顫嬮崟鍨暥缂備胶濮垫繛濠囧蓟閺囩喎绶為柛顐ｇ箓椤绱撴担鍝勑ｉ柣鎿勭節瀵鈽夐姀鐘栥劎鎲稿鍡欘浄闂侇剙绉甸悡娑樏归敐澶嬩氦闁告梻鍠栭弻?
	BetaPolicyScopeAll     = "all"   // 闂傚倸鍊风粈浣革耿闁秲鈧倹绂掔€ｎ亞锛涢梺鐟板⒔缁垶鎮″☉銏＄厱妞ゆ劧绲跨粻銉︿繆閼碱剙甯舵い顓℃硶閹瑰嫰宕崟鍏哥磾闁诲氦顫夊ú鈺冨緤妤ｅ叝鍥偋閸喎鍔呴梺闈涚墕鐎涒晠鎷烘径鎰拻?	BetaPolicyScopeOAuth   = "oauth"   // 濠?OAuth 闂傚倷娴囧畷鍨叏閻㈢绀夌憸蹇曞垝婵犳艾绠ｉ柨婵嗗暕濮?
	BetaPolicyScopeOAuth   = "oauth"
	BetaPolicyScopeAPIKey  = "apikey"  // 濠?API Key 闂傚倷娴囧畷鍨叏閻㈢绀夌憸蹇曞垝婵犳艾绠ｉ柨婵嗗暕濮?
	BetaPolicyScopeBedrock = "bedrock" // 濠?AWS Bedrock 闂傚倷娴囧畷鍨叏閻㈢绀夌憸蹇曞垝婵犳艾绠ｉ柨婵嗗暕濮?
)

// BetaPolicyRule 闂傚倸鍊风粈渚€骞夐敓鐘偓鍐幢濡炴洖鎼灒濞撴凹鍨抽崝?Beta 缂傚倸鍊搁崐鐑芥倿閿斿墽鐭欓柟鐑橆殕閸嬪绻濇繝鍌滃婵鐓￠弻鐔兼焽閿曗偓閺嬬喖鏌涢悩鍙夘棦闁哄本鐩鎾Ω閵夈倗鐩庨梻?
type BetaPolicyRule struct {
	BetaToken            string   `json:"beta_token"` // beta token 闂?	Action               string   `json:"action"`                           // "pass" | "filter" | "block"
	Action               string   `json:"action"`
	Scope                string   `json:"scope"`                            // "all" | "oauth" | "apikey" | "bedrock"
	ErrorMessage         string   `json:"error_message,omitempty"`          // 闂傚倸鍊烽懗鍫曞储瑜旈妴鍐╂償閵忋埄娲稿┑鐘诧工閻楀﹪宕戦埡鍛厽闁逛即娼ф晶浼存煃缂佹ɑ绀€妞ゎ厼娼￠幊婊堟濞戞﹩娼旈梻浣告惈閹冲寮查悩璇茬畺鐟滅増甯掗悙濠囨偣妤︽寧顏犲ù鐙€鍘剧槐鎾存媴閸撳弶楔闂佺绻戦敃銏ゆ偘?(action=block 闂傚倸鍊风粈渚€骞栭锕€鐤い鎰剁稻濞呯娀骞栧ǎ顒€濡奸柡瀣╃窔閺岀喖骞戦幇闈涙闂?
	ModelWhitelist       []string `json:"model_whitelist,omitempty"`        // 婵犵數濮烽。钘壩ｉ崨鏉戝瀭妞ゅ繐鐗嗛悞鍨亜閹哄棗浜剧紒鍓ц檸閸欏啴宕洪埀顒併亜閹烘垵鈧悂宕㈢€电硶鍋撶憴鍕濠殿噣绠栨俊鐢稿箛閺夎法顔婇梺鐟扮摠缁诲棛绮婄€涙绡€缁剧増蓱椤﹪鏌涢…鎴滈偗闁糕斁鍋撳銈嗗坊閸嬫挾鐥紒銏犲箻婵炴垹鏁婚崺鈧い鎺戝閳锋垿鏌涘☉姗堝伐缂佹纰嶉妵鍕Ω閵夘喗鈻堟繝纰樺墲閹稿啿鐣峰鈧、娆撴嚍閵夛妇褰ㄩ梻鍌欑婢瑰﹪宕戦崨顖氬灊闊洦娲滈惌鍡涙倵閿濆骸澧扮痪?闂傚倷娴囬褏鑺遍懖鈺佺筏濠电姵鐔紞鏍ь熆鐠轰警鍎愭繛鍛Х閳ь剙绠嶉崕閬嵥囬婧惧亾濮橆剦妲告い顓℃硶閹瑰嫰鎼圭憴鍕毆濠电偞鎸婚崺鍐磻閹惧灈鍋撶憴鍕婵犫偓闁秴鐒垫い鎺戯功缁夌敻鏌涢幘鏉戝摵鐎殿喗鐓￠獮鏍ㄦ媴閸︻厼骞愬┑鐐舵彧缁蹭粙骞楀鍕珷鐎规洖娲ㄧ壕?
	FallbackAction       string   `json:"fallback_action,omitempty"`        // 闂傚倸鍊风粈渚€骞栭锔藉亱婵犲﹤瀚々鍙夌節婵犲倹鍣搁柣鎺戯攻閵囧嫰寮村Δ鈧禍楣冩⒑閸濆嫯瀚伴柨鏇樺€濋、妯荤附缁嬭法鍊炲銈庡幗閸ㄥ啿危閹扮増鈷掑ù锝呮啞閸熺偤鏌ら悷鏉库挃缂侇喗妫冨鍫曞箣閻樿櫕顔曢梻浣虹帛閺屻劑宕ョ€ｎ喖纾跨€广儱顦伴悡鏇㈡煛閸ャ儱濡奸柛鏃€鏌ㄩ妴鎺戭潩閿濆懍澹曢柣搴ゎ潐濞叉牕锕㈤柆宥呯劦妞ゆ帒锕︾粔鐢告煕閹炬潙鍝洪柟顕€绠栧畷锟犳倻閸℃瑥鏁搁梺鑽ゅТ閹碱偊骞栭锔绘晜妞ゆ劧闄勯悡娆愩亜閺冨倶鈧鐓鍌楀亾濞堝灝娅橀柛鎾跺枛楠炲啴鎮滈挊澶岊吋濡炪倖姊婚搹搴ㄥ绩?
	FallbackErrorMessage string   `json:"fallback_error_message,omitempty"` // 闂傚倸鍊风粈渚€骞栭锔藉亱婵犲﹤瀚々鍙夌節婵犲倹鍣搁柣鎺戯攻閵囧嫰寮村Δ鈧禍楣冩⒑閸濆嫯瀚伴柨鏇樺€濋、妯荤附缁嬭法鍊炲銈庡幗閸ㄥ啿危閹扮増鈷掑ù锝呮啞閸熺偤鏌ら悷鏉库挃缂侇喗妫冨鍫曞箣閻樿櫕顔曢梻浣虹帛濮婂宕㈣閻氭儳顓兼径瀣幈闂佽宕樼亸娆撳礉瀹ュ鐓熼柨婵嗩槷閹查箖鏌＄仦璇测偓鏍紦閼恒儱绶炲璺虹灱濞夊潡姊绘担椋庝覆缂佹彃娼″畷鎴﹀箻閸撲胶鐓撴繝鐢靛Т閸燁偆娆㈤悙鐑樼厵闂侇叏绠戦獮妤佺箾閸涱喗宕岄柡宀嬬秮閹晠宕ｆ径灞诲亹闂備焦鎮堕崝蹇旂椤掑倸鍨濋柡鍐ㄧ墛閸婂鏌﹀Ο渚Ш闁?(fallback_action=block 闂傚倸鍊风粈渚€骞栭锕€鐤い鎰剁稻濞呯娀骞栧ǎ顒€濡奸柡瀣╃窔閺岀喖骞戦幇闈涙闂?
}

// BetaPolicySettings Beta 缂傚倸鍊搁崐鐑芥倿閿斿墽鐭欓柟鐑橆殕閸嬪绻濇繝鍌滃婵鐓￠弻鐔煎箥椤旂⒈鏆梺缁樻尨閸嬫捇鏌ｆ惔锛勭暛闁稿骸宕叅闁靛ň鏅涢崙?
type BetaPolicySettings struct {
	Rules []BetaPolicyRule `json:"rules"`
}

// OverloadCooldownSettings 529闂傚倷绀侀幖顐λ囬锕€鐤炬繛鎴炩棨濞差亜鐓涘ù鐓庡船缂嶅﹤鐣烽崼鏇ㄦ晢闁逞屽墰缁牓宕橀浣镐壕闁荤喐婢橀顏呯節閵忊槅鐒界紒顔硷躬閺屽棗顓奸崱蹇斿闂備礁鎼ˇ鍐测枖閺団€斥偓鎾⒒?
type OverloadCooldownSettings struct {
	Enabled bool `json:"enabled"`
	// CooldownMinutes 闂傚倸鍊风粈渚€骞夐敓鐘茬闁哄洢鍨规惔濠囨煙缂佹ê绗ч柡鍡愬€濋弻娑㈩敃閻樻彃濮庣紓渚囧亜缁夊綊寮诲☉銏╂晝闁挎繂妫涢ˇ銊モ攽椤曞棛绋婚柛鏃€鐟╁濠氬Ω閵夈垺顫嶅┑鈽嗗灣閸樠囥€傚ú顏呪拺闁硅偐鍋涙俊濂告⒑鐢喚绉€殿喖顭烽弫鍐磼濞戞ü绨介梻浣侯焾閺堫剛鍒掓惔銊?
	CooldownMinutes int `json:"cooldown_minutes"`
}

// RateLimit429CooldownSettings 429濠电姵顔栭崰妤冩暜濡ゅ啰鐭欓柟鐑樸仜閳ь剨绠撳畷濂稿Ψ椤旇姤娅嶅┑鐘垫暩婵敻鎳濇ィ鍐╁€峰┑鐘叉处閻撶喖鏌熼柇锕€鐏￠柕鍡楀暟缁辨捇宕掑鍏尖枅濠殿喖锕ュ钘夘嚕閸撲焦宕夐柕濠冨姂閸婃繈寮?
type RateLimit429CooldownSettings struct {
	// Enabled 闂傚倸鍊风粈渚€骞栭銈傚亾濮樺崬鍘寸€规洝顫夌€靛ジ寮堕幋鐘垫毎濠电偞鎸婚崺鍐磻閹惧灈鍋撶憴鍕闁稿繑蓱娣囧﹪骞栨担瑙勬珕闂佸搫鍊堕崐婵嬫偘濠婂嫮绡€闁汇垽娼ф禒婊勩亜閿旇姤绶查悡銈嗙節闂堟侗鐓柡鍡曞嵆濮婄粯鎷呴崨濠冨創濡炪倖鍨靛Λ婵嗙暦濠靛绠抽柡鍐ㄥ€婚悾鍫曟⒑缁嬫寧婀伴柛婵嗛叄閸┾偓妞ゆ帊绀佸ù顔筋殽閻愬澧柟宄版嚇瀹曘劍绻濇惔婵囧煕缂傚倸鍊搁崐鎼佸磹閹间礁绐楁慨妯挎硾缁愭淇婇妶鍌氫壕濡炪値浜滈崯鏉戠暦閹烘鍊风紒顔款潐鐎氫粙姊虹拠鎻掝劉缂佸甯熼幗顐㈩渻閵囶垯绀佸ú銈囩不瑜版帗鍊垫繛鎴烆伆閹达箑鐓曢柟鐑橆殕閻撴稑霉閿濆娑ч柍褜鍓氶悧鐘诲灳閺冨牆鐒垫い鎺嶆缁诲棙銇勯弽銊ょ繁闁稿簺鍎甸弻?29闂傚倸鍊烽悞锕傚箖閸洖纾块柟鎯版绾惧鏌曢崼婵囶棞妞?
	Enabled         bool `json:"enabled"`
	CooldownSeconds int  `json:"cooldown_seconds"`
}

// DefaultOverloadCooldownSettings returns default overload cooldown settings.
func DefaultOverloadCooldownSettings() *OverloadCooldownSettings {
	return &OverloadCooldownSettings{
		Enabled:         true,
		CooldownMinutes: 10,
	}
}

// DefaultRateLimit429CooldownSettings 闂傚倷绀侀幖顐λ囬锕€鐤炬繝濠傜墕閽冪喖鏌曟繛鍨壄婵炲樊浜滈崘鈧銈庡幗閸ㄨ埖绂掑Ο琛℃斀闁宠棄妫楅悘鈩冦亜閹寸偟鎳囩€规洏鍎靛畷鍫曨敆娴ｅ搫骞?29闂傚倸鍊烽悞锕傚箖閸洖纾块柟鎯版绾惧鏌曢崼婵囶棞妞ゆ洟浜堕弻宥夊传閸曨剙娅ら梺缁樻尨閸嬫捇鏌ｆ惔锛勭暛闁稿骸宕叅闁靛ň鏅涢崙鐘绘煕瀹€鈧崑鐐烘偂閵夛妇绠鹃柟瀵稿剱閻掕棄顭胯濞茬喖寮婚悢纰辨晬婵娉涘浼存⒑鐠団€虫灆缂侇喗鐟╅崹楣冩晝閸屾氨顔婇梺鍝勫暙閸嬪棗危?缂傚倸鍊搁崐椋庣矆娓氣偓钘濋柟娈垮枟閺嗘粍銇勮箛鎾搭棡妞?
func DefaultRateLimit429CooldownSettings() *RateLimit429CooldownSettings {
	return &RateLimit429CooldownSettings{
		Enabled:         true,
		CooldownSeconds: 5,
	}
}

// DefaultBetaPolicySettings 闂傚倷绀侀幖顐λ囬锕€鐤炬繝濠傜墕閽冪喖鏌曟繛鍨壄婵炲樊浜滈崘鈧銈庡幗閸ㄨ埖绂掑Ο琛℃斀闁宠棄妫楅悘鈩冦亜閹寸偟鎳囩€规洏鍎靛畷鍫曨敆娴ｅ搫骞?Beta 缂傚倸鍊搁崐鐑芥倿閿斿墽鐭欓柟鐑橆殕閸嬪绻濇繝鍌滃婵鐓￠弻鐔煎箥椤旂⒈鏆梺缁樻尨閸嬫捇鏌ｆ惔锛勭暛闁稿骸宕叅闁靛ň鏅涢崙?
func DefaultBetaPolicySettings() *BetaPolicySettings {
	return &BetaPolicySettings{
		Rules: []BetaPolicyRule{
			{
				BetaToken: "fast-mode-2026-02-01",
				Action:    BetaPolicyActionFilter,
				Scope:     BetaPolicyScopeAll,
			},
			{
				BetaToken: "context-1m-2025-08-07",
				Action:    BetaPolicyActionFilter,
				Scope:     BetaPolicyScopeAll,
			},
		},
	}
}

// OpenAI Fast Policy 缂傚倸鍊搁崐鐑芥倿閿斿墽鐭欓柟鐑橆殕閸嬪绻濇繝鍌滃婵鐓￠弻鐔兼焽閿曗偓閺嬬喖鏌￠崱妯肩煉闁哄瞼鍠栧鑽も偓闈涘濡差喖鈹?
// OpenAI 闂?"fast 婵犵數濮烽。钘壩ｉ崨鏉戝瀭妞ゅ繐鐗嗛悞鍨亜閹哄棗浜剧紒鍓ц檸閸樻儳鈽夐悽绋跨劦? 闂傚倸鍊搁崐椋庢閿熺姴纾婚柛娑卞枤閳瑰秹鏌ц箛姘兼綈鐎规洘鐓￠弻娑㈠箛閸忓摜鍑归柣搴㈢瀹€绋款潖濞差亜鍨傛い鏇炴噹閸撳啿鈹戦悩顐壕闂佸搫顦伴崵姘炽亹閹烘挻娅滈梺绯曞墲閿氶柣顓у櫍濮婃椽鎮℃惔锝勭驳闂佹悶鍔屽锟犵嵁?service_tier 闂傚倷娴囬褏鈧稈鏅濈划娆撳箳濡炲皷鍋撻崘顔煎耿婵炴垼椴搁弲鈺呮倵閸忓浜鹃梺閫炲苯澧存鐐茬箻閹晝鎷犻幓鎺曗偓鍨攽鎺抽崐鎰板磻閹剧粯鐓熼柟鐑樺灟閸嬨垽鏌?//   - "priority"闂傚倸鍊烽悞锔锯偓绗涘懐鐭欓柟杈鹃檮閸嬪鏌涢埄鍐х繁闁轰礁娲ㄩ幉鎼佹偋閸繄鐟ㄧ紓浣哄缁蹭粙鈥旈崘顏佸亾閿濆簼绨婚柣锔哄妼闇夐柣鎾抽閳ь剙娼″璇测槈濡攱鐎诲┑鈽嗗灥濞咃絾绂掕缁?"fast"闂傚倸鍊烽悞锔锯偓绗涘懐鐭欓柟杈鹃檮閸ゆ劖銇勯弽銊х煂缂佲偓婵犲倵鏀介柣妯诲絻閺嗙喖鏌嶇紒妯活棃闁哄苯绉烽¨渚€鏌涢幘璺烘瀻闁伙絿鍏橀幃鈺冩嫚閹绘帒鏁ら梻浣瑰濞叉垿鎳楅崼鏇€?"priority"闂傚倸鍊烽悞锔锯偓绗涘懐鐭欓柟娆″眰鍔戦崺鈧い鎺戝€荤壕濂稿级閸碍娅嗛柣婵堛亼st 婵犵數濮烽。钘壩ｉ崨鏉戝瀭妞ゅ繐鐗嗛悞鍨亜閹哄棗浜剧紒鍓ц檸閸樻儳鈽夐悽绋跨劦?
//   - "flex"闂傚倸鍊烽悞锔锯偓绗涘懐鐭欓柟鐑橆殕閸庢銇勯弬鍨挃缂佲偓婵犲洦鐓曢柍鈺佸幘椤忓牜鏁嬮弶鍫涘妺缁诲棙銇勯弽銊ь暡闁诡垰鐗撻幃浠嬵敍濡搫濮㈢紓浣介哺鐢繝骞冮鍕ㄦ瀻闊洦鎼╅埀顒€鑻湁闁挎繂鎳庨ˉ蹇涙煕?//   - 闂傚倸鍊烽悞锕€顪冮幐搴ｎ洸婵炲棙鎸婚崵鍕煕椤愶絾绀€婵鐓￠弻鐔煎箥椤旂⒈鏆梺鍝勬媼閸撴岸骞堥妸銉㈡斀闁割偆鍣︾槐鐚簃al 濠电姵顔栭崰妤冩暜濡ゅ啰鐭欓柟鐑樸仜閳ь剨绠撳畷濂稿Ψ椤旇姤娅?
//
// 闂傚倸鍊风粈渚€骞栭锔藉亱婵犲﹤鐗嗙粈鍫熺箾閹存瑥鐏柛瀣ㄥ€濋弻鏇熷緞閸℃ɑ鐝曢梺绋款儌閺呯娀寮婚弴鐔风窞闁割偅绻傛慨璺衡攽閳╁啫绲绘い顓犲厴瀵?BetaPolicyAction*/BetaPolicyScope* 闂傚倷鐒﹂惇褰掑春閸曨垰鍨傞梺顒€绉寸粻鏍煕瀹€鈧崑娑㈡⒒椤栫偞鐓ｉ煫鍥ㄥ嚬濞兼劙鎮楀顓狀暡闁靛洤瀚伴獮妯尖偓闈涙憸閻ｅジ姊哄Ч鍥р偓鏇㈠箠濮椻偓瀵濡搁妷銏☆潔濠碘槅鍨伴悘婵嬵敂椤撱垺鍋℃繝濠傚鎯熼梺鎼炲劀閸愩劍灏嬪┑锛勫亼閸婃牠鎮уΔ鍛殞濡わ絽鍟崑澶愭⒑椤掆偓缁夌敻鍩涢幋锔界厸濠㈣泛瀛╃涵鍫曞极閸儲鈷戦弶鐐村鐠愪即鏌涘Ο鍏兼珪缂?// anthropic-beta header 闂傚倸鍊烽懗鍫曞箠閹惧墎涓嶇€广儱顦崹鍌涚箾瀹割喕绨婚柛?body 闂?service_tier 闂傚倷娴囬褏鈧稈鏅濈划娆撳箳濡炲皷鍋撻崘顔煎耿婵炴垼椴搁弲鈺呮倵閸忓浜鹃梺鍛婃处閸?const (
const (
	OpenAIFastTierAny      = "all"      // 闂傚倸鍊风粈渚€骞夐敓鐘冲亱闁哄洨濮风粈濠傗攽閻樺弶鎼愰柛灞诲姂閺屾洟宕煎┑鍥舵￥濡炪倐鏅犻ˉ鎾舵閹惧瓨濯村┑顔藉焾娴滄繄绮╅悢鐓庨唶闁冲灈鏅涙禍楣冩偡濞嗗繐顏柟钘夊暣閹绠涢敐鍕仐闂佽鍨欢姘剁嵁鐎ｎ喗鏅濋柍褜鍓涚划鍫ュ幢濡晲绨婚梺瑙勫閺呮盯鍩€椤掍胶澧甸柟?service_tier
	OpenAIFastTierPriority = "priority" // 濠电姷鏁搁崑娑㈩敋椤撶喐鍙忛柟缁㈠枛缁犵娀骞栨潏鍓у闁绘帒锕ラ妵鍕即濡も偓娴滈箖姊?fast闂傚倸鍊烽悞锔锯偓绗涘懐鐭欓柟瀛樼箥閻掔晫绱掔仦鍓х憮ority闂?	OpenAIFastTierFlex     = "flex"     // 濠电姷鏁搁崑娑㈩敋椤撶喐鍙忛柟缁㈠枛缁犵娀骞栨潏鍓у闁绘帒锕ラ妵鍕即濡も偓娴滈箖姊?flex
	OpenAIFastTierFlex     = "flex"
)

// OpenAIFastPolicyRule 闂傚倸鍊风粈渚€骞夐敓鐘偓鍐幢濡炴洖鎼灒濞撴凹鍨抽崝?OpenAI fast/flex 缂傚倸鍊搁崐鐑芥倿閿斿墽鐭欓柟鐑橆殕閸嬪绻濇繝鍌滃婵鐓￠弻鐔兼焽閿曗偓閺嬬喖鏌涢悩鍙夘棦闁哄本鐩鎾Ω閵夈倗鐩庨梻?
type OpenAIFastPolicyRule struct {
	ServiceTier          string   `json:"service_tier"`                     // "priority" | "flex" | "auto" | "default" | "scale" | "all"
	Action               string   `json:"action"`                           // "pass" | "filter" | "block"
	Scope                string   `json:"scope"`                            // "all" | "oauth" | "apikey" | "bedrock"
	ErrorMessage         string   `json:"error_message,omitempty"`          // 闂傚倸鍊烽懗鍫曞储瑜旈妴鍐╂償閵忋埄娲稿┑鐘诧工閻楀﹪宕戦埡鍛厽闁逛即娼ф晶浼存煃缂佹ɑ绀€妞ゎ厼娼￠幊婊堟濞戞﹩娼旈梻浣告惈閹冲寮查悩璇茬畺鐟滅増甯掗悙濠囨偣妤︽寧顏犲ù鐙€鍘剧槐鎾存媴閸撳弶楔闂佺绻戦敃銏ゆ偘?(action=block 闂傚倸鍊风粈渚€骞栭锕€鐤い鎰剁稻濞呯娀骞栧ǎ顒€濡奸柡瀣╃窔閺岀喖骞戦幇闈涙闂?
	ModelWhitelist       []string `json:"model_whitelist,omitempty"`        // 婵犵數濮烽。钘壩ｉ崨鏉戝瀭妞ゅ繐鐗嗛悞鍨亜閹哄棗浜剧紒鍓ц檸閸欏啴宕洪埀顒併亜閹烘垵鈧悂宕㈢€电硶鍋撶憴鍕濠殿噣绠栨俊鐢稿箛閺夎法顔婇梺鐟扮摠缁诲棛绮婄€涙绡€缁剧増蓱椤﹪鏌涢…鎴滈偗闁糕斁鍋撳銈嗗坊閸嬫挾鐥紒銏犲箻婵炴垹鏁婚崺鈧い鎺戝閳锋垿鏌涘☉姗堝伐缂佹纰嶉妵鍕Ω閵夘喗鈻堟繝纰樺墲閹稿啿鐣峰鈧、娆撴嚍閵夛妇褰ㄩ梻鍌欑婢瑰﹪宕戦崨顖氬灊闊洦娲滈惌鍡涙倵閿濆骸澧扮痪?闂傚倷娴囬褏鑺遍懖鈺佺筏濠电姵鐔紞鏍ь熆鐠轰警鍎愭繛鍛Х閳ь剙绠嶉崕閬嵥囬婧惧亾濮橆剦妲告い顓℃硶閹瑰嫰鎼圭憴鍕毆濠电偞鎸婚崺鍐磻閹惧灈鍋撶憴鍕婵犫偓闁秴鐒垫い鎺戯功缁夌敻鏌涢幘鏉戝摵鐎殿喗鐓￠獮鏍ㄦ媴閸︻厼骞愬┑鐐舵彧缁蹭粙骞楀鍕珷鐎规洖娲ㄧ壕?
	FallbackAction       string   `json:"fallback_action,omitempty"`        // 闂傚倸鍊风粈渚€骞栭锔藉亱婵犲﹤瀚々鍙夌節婵犲倹鍣搁柣鎺戯攻閵囧嫰寮村Δ鈧禍楣冩⒑閸濆嫯瀚伴柨鏇樺€濋、妯荤附缁嬭法鍊炲銈庡幗閸ㄥ啿危閹扮増鈷掑ù锝呮啞閸熺偤鏌ら悷鏉库挃缂侇喗妫冨鍫曞箣閻樿櫕顔曢梻浣虹帛閺屻劑宕ョ€ｎ喖纾跨€广儱顦伴悡鏇㈡煛閸ャ儱濡奸柛鏃€鏌ㄩ妴鎺戭潩閿濆懍澹曢柣搴ゎ潐濞叉牕锕㈤柆宥呯劦妞ゆ帒锕︾粔鐢告煕閹炬潙鍝洪柟顕€绠栧畷锟犳倻閸℃瑥鏁搁梺鑽ゅТ閹碱偊骞栭锔绘晜妞ゆ劧闄勯悡娆愩亜閺冨倶鈧鐓鍌楀亾濞堝灝娅橀柛鎾跺枛楠炲啴鎮滈挊澶岊吋濡炪倖姊婚搹搴ㄥ绩?
	FallbackErrorMessage string   `json:"fallback_error_message,omitempty"` // 闂傚倸鍊风粈渚€骞栭锔藉亱婵犲﹤瀚々鍙夌節婵犲倹鍣搁柣鎺戯攻閵囧嫰寮村Δ鈧禍楣冩⒑閸濆嫯瀚伴柨鏇樺€濋、妯荤附缁嬭法鍊炲銈庡幗閸ㄥ啿危閹扮増鈷掑ù锝呮啞閸熺偤鏌ら悷鏉库挃缂侇喗妫冨鍫曞箣閻樿櫕顔曢梻浣虹帛濮婂宕㈣閻氭儳顓兼径瀣幈闂佽宕樼亸娆撳礉瀹ュ鐓熼柨婵嗩槷閹查箖鏌＄仦璇测偓鏍紦閼恒儱绶炲璺虹灱濞夊潡姊绘担椋庝覆缂佹彃娼″畷鎴﹀箻閸撲胶鐓撴繝鐢靛Т閸燁偆娆㈤悙鐑樼厵闂侇叏绠戦獮妤佺箾閸涱喗宕岄柡宀嬬秮閹晠宕ｆ径灞诲亹闂備焦鎮堕崝蹇旂椤掑倸鍨濋柡鍐ㄧ墛閸婂鏌﹀Ο渚Ш闁?(fallback_action=block 闂傚倸鍊风粈渚€骞栭锕€鐤い鎰剁稻濞呯娀骞栧ǎ顒€濡奸柡瀣╃窔閺岀喖骞戦幇闈涙闂?
}

// OpenAIFastPolicySettings OpenAI fast 缂傚倸鍊搁崐鐑芥倿閿斿墽鐭欓柟鐑橆殕閸嬪绻濇繝鍌滃婵鐓￠弻鐔煎箥椤旂⒈鏆梺缁樻尨閸嬫捇鏌ｆ惔锛勭暛闁稿骸宕叅闁靛ň鏅涢崙?
type OpenAIFastPolicySettings struct {
	Rules []OpenAIFastPolicyRule `json:"rules"`
}

// DefaultOpenAIFastPolicySettings 闂傚倷绀侀幖顐λ囬锕€鐤炬繝濠傜墕閽冪喖鏌曟繛鍨壄婵炲樊浜滈崘鈧銈庡幗閸ㄨ埖绂掑Ο琛℃斀闁宠棄妫楅悘鈩冦亜閹寸偟鎳囩€规洏鍎靛畷鍫曨敆娴ｅ搫骞?OpenAI fast 缂傚倸鍊搁崐鐑芥倿閿斿墽鐭欓柟鐑橆殕閸嬪绻濇繝鍌滃婵鐓￠弻鐔煎箥椤旂⒈鏆梺缁樻尨閸嬫捇鏌ｆ惔锛勭暛闁稿骸宕叅闁靛ň鏅涢崙鐘绘煕瀹€鈧崑鐐烘偂?// 濠电姵顔栭崰妤冩暜濡ゅ啰鐭欓柟鐑樸仜閳ь剨绠撳畷濂稿Ψ椤旇姤娅嶇紓鍌氬€烽悞锕傗€﹂崶顒€鐓濋柡鍐ㄥ€甸崑鎾荤嵁閸喖濮庡銈忓瘜閸ㄦ娊寮鑲╂殾闁搞儮鏅濋敍婵嬫⒑閸涘﹤濮堥柛搴㈠▕瀵鈻庨幘瀵稿幍濡ょ姷鍋涢悘婵嬫倶椤忓牊鐓欑€瑰嫮澧楅崵鍥┾偓娈垮枟閹告娊宕洪妷鈺佸耿婵☆垵鍋愬暩闂傚倸鍊风粈渚€骞夐敍鍕殰婵°倕鎳忛崑锟犲级閸碍娅嗘い顐ｆ礃缁绘繈妫冨☉娆忕獩闂佸磭鎳撶粔鐢垫崲濠靛顥堟繛鎴炶壘椤も偓婵?OpenAI 濠电姷鏁搁崑鐐哄垂閸洖绠伴柟闂寸劍閺呮繈鏌曟径鍡樻珕闁?service_tier 闂傚倷娴囧畷鍨叏閺夋嚚褰掑磼閻愭彃鍋嶉梺鎯х箰濠€杈ㄧ▔瀹ュ鐓熼柡鍌氱仢閹垿鏌￠崪浣稿缂佺粯鐩獮瀣倷閹绘帗鐦撻梻浣告啞閸╁啴宕戦幘缁樷拻濞达綀妫勯崥褰掓煕鐎ｎ偆鈽夐摶鐐翠繆閵堝懏鍣圭紒鐘靛枛閺屟嗙疀閿濆懍绨电紓浣筋嚙濡繈寮诲☉妯兼殕闁逞屽墴瀹曟垿鎮欑€靛壊娴?
// DefaultOpenAIFastPolicySettings returns default OpenAI fast/flex policy settings.
func DefaultOpenAIFastPolicySettings() *OpenAIFastPolicySettings {
	return &OpenAIFastPolicySettings{
		Rules: []OpenAIFastPolicyRule{},
	}
}

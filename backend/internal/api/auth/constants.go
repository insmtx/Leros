package auth

const (
	// ProviderGitHub 表示 GitHub 平台。
	ProviderGitHub = "github"

	// SubjectTypeUser identifies a user-owned execution subject.
	SubjectTypeUser = "user"

	// ScopeTypeEvent identifies an event-scoped execution request.
	ScopeTypeEvent = "event"

	// AccountOwnerTypeUser 表示账户属于具体用户。
	AccountOwnerTypeUser = "user"

	// AccountOwnerTypeSystem 表示账户属于系统或组织级执行身份。
	AccountOwnerTypeSystem = "system"

	// AccountTypeUserOAuth 表示通过 OAuth 授权得到的用户账户。
	AccountTypeUserOAuth = "user_oauth"

	// AccountTypeAppInstallation 表示第三方平台应用安装身份。
	AccountTypeAppInstallation = "app_installation"

	// GrantTypeOAuth2 表示 OAuth2 访问令牌。
	GrantTypeOAuth2 = "oauth2"

	// AccountStatusActive 表示账户当前可用。
	AccountStatusActive = "active"

	// AccountStatusDisabled 表示账户已禁用。
	AccountStatusDisabled = "disabled"
)

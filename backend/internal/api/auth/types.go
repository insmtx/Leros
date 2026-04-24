package auth

import (
	"time"
)

// AuthorizedAccount 表示使用者授权后可被系统复用的第三方账户。
type AuthorizedAccount struct {
	ID                string            `json:"id"`
	UserID            string            `json:"user_id"`
	Provider          string            `json:"provider"`
	OwnerType         string            `json:"owner_type"`
	AccountType       string            `json:"account_type"`
	ExternalAccountID string            `json:"external_account_id"`
	DisplayName       string            `json:"display_name"`
	Scopes            []string          `json:"scopes"`
	Status            string            `json:"status"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// AccountCredential 表示账户当前可用的授权材料。
type AccountCredential struct {
	AccountID    string            `json:"account_id"`
	GrantType    string            `json:"grant_type"`
	AccessToken  string            `json:"access_token,omitempty"`
	RefreshToken string            `json:"refresh_token,omitempty"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// UserProviderBinding 表示某用户在某 provider 下的默认账户绑定。
type UserProviderBinding struct {
	UserID    string `json:"user_id"`
	Provider  string `json:"provider"`
	AccountID string `json:"account_id"`
	IsDefault bool   `json:"is_default"`
	Priority  int    `json:"priority"`
}

// OAuthState 表示一次未完成的 OAuth 授权会话。
type OAuthState struct {
	State       string    `json:"state"`
	UserID      string    `json:"user_id"`
	Provider    string    `json:"provider"`
	RedirectURI string    `json:"redirect_uri,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// StartAuthorizationRequest 表示发起授权请求。
type StartAuthorizationRequest struct {
	UserID      string
	Provider    string
	RedirectURI string
}

// AuthorizationCallbackRequest 表示 OAuth 回调请求。
type AuthorizationCallbackRequest struct {
	Provider string
	State    string
	Code     string
}

// CompleteAuthorizationRequest 表示 provider 完成授权所需的上下文。
type CompleteAuthorizationRequest struct {
	State *OAuthState
	Code  string
}

// AuthorizationResult 表示 provider 完成授权后的结果。
type AuthorizationResult struct {
	Account    *AuthorizedAccount
	Credential *AccountCredential
}

// AuthSelector carries the minimal execution identity hints needed for runtime auth resolution.
type AuthSelector struct {
	Provider          string            `json:"provider,omitempty"`
	ExplicitProfileID string            `json:"explicit_profile_id,omitempty"`
	SubjectType       string            `json:"subject_type,omitempty"`
	SubjectID         string            `json:"subject_id,omitempty"`
	ScopeType         string            `json:"scope_type,omitempty"`
	ScopeID           string            `json:"scope_id,omitempty"`
	ExternalRefs      map[string]string `json:"external_refs,omitempty"`
}

// ResolveAccountRequest 表示一次运行时账户解析请求。
type ResolveAccountRequest struct {
	Selector *AuthSelector

	// Legacy compatibility fields. New call sites should prefer Selector.
	UserID    string
	Provider  string
	AccountID string
}

// ResolvedAccount 表示解析完成的账户与凭证结果。
type ResolvedAccount struct {
	Account    *AuthorizedAccount
	Credential *AccountCredential
	ResolvedBy string
}

// ResolveAuthorizationRequest describes a provider-agnostic runtime authorization lookup.
type ResolveAuthorizationRequest struct {
	Selector *AuthSelector

	// Legacy compatibility fields. New call sites should prefer Selector.
	UserID    string
	Provider  string
	AccountID string
}

// ResolvedAuthorization is the provider-agnostic output of runtime authorization resolution.
type ResolvedAuthorization struct {
	Provider   string
	ProfileID  string
	ResolvedBy string
	Account    *AuthorizedAccount
	Credential *AccountCredential
	Labels     map[string]string
	Resources  map[string]interface{}
}

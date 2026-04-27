package contract

import "time"

// AuthorizedAccount 表示授权后可被系统复用的第三方账户
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

// AccountCredential 表示账户当前可用的授权材料
type AccountCredential struct {
	AccountID    string            `json:"account_id"`
	GrantType    string            `json:"grant_type"`
	AccessToken  string            `json:"access_token,omitempty"`
	RefreshToken string            `json:"refresh_token,omitempty"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// StartAuthorizationRequest 表示发起授权请求
type StartAuthorizationRequest struct {
	UserID      string `json:"user_id"`
	Provider    string `json:"provider"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

// AuthorizationCallbackRequest 表示 OAuth 回调请求
type AuthorizationCallbackRequest struct {
	Provider string `json:"provider"`
	State    string `json:"state"`
	Code     string `json:"code"`
}

// ResolveAuthorizationRequest 表示运行时授权解析请求
type ResolveAuthorizationRequest struct {
	UserID    string `json:"user_id"`
	Provider  string `json:"provider"`
	AccountID string `json:"account_id,omitempty"`
}

// AuthorizationResult 表示完成授权后的结果
type AuthorizationResult struct {
	Account    *AuthorizedAccount `json:"account"`
	Credential *AccountCredential `json:"credential"`
}

// ResolvedAuthorization 表示解析完成的授权结果
type ResolvedAuthorization struct {
	Account    *AuthorizedAccount `json:"account"`
	Credential *AccountCredential `json:"credential"`
}

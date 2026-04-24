package auth

import (
	"context"
)

// ProviderAuthResolver resolves one provider-specific authorization path.
type ProviderAuthResolver interface {
	ResolveAuthorization(ctx context.Context, req *ResolveAuthorizationRequest) (*ResolvedAuthorization, bool, error)
}

// AuthorizationProvider 定义 provider 授权接入所需接口。
type AuthorizationProvider interface {
	ProviderCode() string
	BuildAuthorizationURL(req *StartAuthorizationRequest, state *OAuthState) (string, error)
	CompleteAuthorization(req *CompleteAuthorizationRequest) (*AuthorizationResult, error)
}

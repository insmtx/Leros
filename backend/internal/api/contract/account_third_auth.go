package contract

import "context"

// ThirdAuthService 定义第三方平台授权服务接口
type ThirdAuthService interface {
	// OAuth 流程启动 - 返回授权 URL
	StartAuthorization(ctx context.Context, req *StartAuthorizationRequest) (string, error)

	// OAuth 回调处理 - 处理 provider 回调，保存授权账户
	HandleAuthorizationCallback(ctx context.Context, req *AuthorizationCallbackRequest) (*AuthorizationResult, error)

	// 运行时授权解析 - 根据条件解析可用的授权账户
	ResolveAuthorization(ctx context.Context, req *ResolveAuthorizationRequest) (*ResolvedAuthorization, error)
}

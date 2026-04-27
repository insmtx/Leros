package auth

import "context"

// Caller 定义了一个执行身份，包含用户 ID、租户 ID 和角色信息。
type Caller struct {
	Uin   uint
	OrgID uint
}

// Trace 定义了一个跟踪信息结构体，用于在请求链路中传递跟踪标识符，帮助进行分布式追踪和日志关联。
type Trace struct {
	// RequestID 是一个全局唯一的标识符，用于标识一次请求，可以用于日志关联和调试。
	RequestID string
	// TraceID 是一个全局唯一的标识符，用于跟踪请求链路中的调用关系。
	TraceID string
	// SpanID 是可选的，如果请求链路中包含多个调用，可以用来标识具体的调用跨度。
	SpanID []string
}

type IdentityContext struct {
	Caller *Caller
	Trace  *Trace
}

// WithContext 携带 Caller 和 Trace 信息的上下文对象。
func WithContext(ctx context.Context, caller *Caller, trace *Trace) context.Context {
	ctx = context.WithValue(ctx, "caller", caller)
	ctx = context.WithValue(ctx, "trace", trace)
	return ctx
}

// FromContext 从上下文中提取 Caller 和 Trace 信息。
func FromContext(ctx context.Context) (*Caller, *Trace) {
	caller, _ := ctx.Value("caller").(*Caller)
	trace, _ := ctx.Value("trace").(*Trace)
	return caller, trace
}

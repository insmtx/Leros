package lifecycle

import (
	"context"
	"fmt"

	"github.com/insmtx/SingerOS/backend/internal/agent"
)

// RuntimeRouter 在注册 runtime 时统一套用 Agent 生命周期。
type RuntimeRouter struct {
	router           *agent.RuntimeRouter
	builder          *ContextBuilder
	toolAvailability ToolAvailability
}

// NewRuntimeRouter 创建带生命周期的 runtime 路由器。
func NewRuntimeRouter(defaultKind string, builder *ContextBuilder, toolAvailability ...ToolAvailability) *RuntimeRouter {
	var availability ToolAvailability
	if len(toolAvailability) > 0 {
		availability = toolAvailability[0]
	}
	return &RuntimeRouter{
		router:           agent.NewRuntimeRouter(defaultKind),
		builder:          builder,
		toolAvailability: availability,
	}
}

// Register 注册原始 runtime，并在内部包装统一生命周期。
func (r *RuntimeRouter) Register(kind string, runner agent.Runner) error {
	if r == nil || r.router == nil {
		return fmt.Errorf("lifecycle runtime router is nil")
	}
	if runner == nil {
		return fmt.Errorf("runtime %q runner is nil", kind)
	}
	if _, ok := runner.(*Runner); ok {
		return r.router.Register(kind, runner)
	}
	return r.router.Register(kind, newRunner(runner, r.builder, r.toolAvailability))
}

// SetDefault 更新默认 runtime。
func (r *RuntimeRouter) SetDefault(kind string) {
	if r == nil || r.router == nil {
		return
	}
	r.router.SetDefault(kind)
}

// Run 根据请求选择 runtime，并执行完整生命周期。
func (r *RuntimeRouter) Run(ctx context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	if r == nil || r.router == nil {
		return nil, fmt.Errorf("lifecycle runtime router is nil")
	}
	return r.router.Run(ctx, req)
}

var _ agent.Runner = (*RuntimeRouter)(nil)

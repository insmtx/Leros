package lifecycle

import (
	"context"
	"fmt"

	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/ygpkg/yg-go/logs"
)

// Runner 在具体运行时外层统一执行 Agent 生命周期。
type Runner struct {
	delegate         agent.Runner
	builder          *ContextBuilder
	toolAvailability ToolAvailability
}

// ToolAvailability resolves registered tool names for lifecycle hooks.
type ToolAvailability interface {
	AvailableToolNames(names []string) []string
}

func newRunner(delegate agent.Runner, builder *ContextBuilder, toolAvailability ToolAvailability) *Runner {
	return &Runner{
		delegate:         delegate,
		builder:          builder,
		toolAvailability: toolAvailability,
	}
}

// Run 构建统一上下文、执行具体运行时，并在结束后触发自我学习检查。
func (r *Runner) Run(ctx context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	if r == nil || r.delegate == nil {
		return nil, fmt.Errorf("lifecycle delegate runner is required")
	}
	if r.builder == nil {
		return nil, fmt.Errorf("lifecycle context builder is required")
	}

	prepared, err := r.builder.Prepare(ctx, req)
	if err != nil {
		return nil, err
	}

	recorder := &traceRecorder{}
	prepared.EventSink = wrapSink(prepared.EventSink, recorder)

	result, runErr := r.delegate.Run(ctx, prepared)
	if err := r.AfterRunLearning(ctx, prepared, result, recorder.trace()); err != nil {
		logs.WarnContextf(ctx, "SingerOS lifecycle learning check failed: %v", err)
	}
	return result, runErr
}

var _ agent.Runner = (*Runner)(nil)

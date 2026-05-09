package agent

import (
	"context"
	"fmt"
	"strings"
)

const (
	// RuntimeKindSingerOS is the built-in SingerOS agent runtime.
	RuntimeKindSingerOS = "singeros"
)

// RuntimeRouter 根据请求选择具体 runtime。
type RuntimeRouter struct {
	defaultKind string
	runners     map[string]Runner
}

// NewRuntimeRouter 创建带默认 runtime 的路由器。
func NewRuntimeRouter(defaultKind string) *RuntimeRouter {
	return &RuntimeRouter{
		defaultKind: normalizeRuntimeKind(defaultKind),
		runners:     make(map[string]Runner),
	}
}

// Register 注册或替换一个 runtime runner。
func (r *RuntimeRouter) Register(kind string, runner Runner) error {
	if r == nil {
		return fmt.Errorf("runtime router is nil")
	}
	kind = normalizeRuntimeKind(kind)
	if kind == "" {
		return fmt.Errorf("runtime kind is required")
	}
	if runner == nil {
		return fmt.Errorf("runtime %q runner is nil", kind)
	}
	if r.runners == nil {
		r.runners = make(map[string]Runner)
	}
	r.runners[kind] = runner
	return nil
}

// SetDefault 更新默认 runtime。
func (r *RuntimeRouter) SetDefault(kind string) {
	if r == nil {
		return
	}
	r.defaultKind = normalizeRuntimeKind(kind)
}

// Run 使用请求指定的 runtime 执行任务。
func (r *RuntimeRouter) Run(ctx context.Context, req *RequestContext) (*RunResult, error) {
	if r == nil {
		return nil, fmt.Errorf("runtime router is nil")
	}

	kind := r.defaultKind
	if req != nil && strings.TrimSpace(req.Runtime.Kind) != "" {
		kind = normalizeRuntimeKind(req.Runtime.Kind)
	}
	if kind == "" {
		return nil, fmt.Errorf("runtime kind is required")
	}

	runner := r.runners[kind]
	if runner == nil {
		return nil, fmt.Errorf("runtime %q is not available", kind)
	}
	return runner.Run(ctx, req)
}

func normalizeRuntimeKind(kind string) string {
	return strings.ToLower(strings.TrimSpace(kind))
}

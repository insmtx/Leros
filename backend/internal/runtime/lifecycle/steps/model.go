package steps

import (
	"context"
	"fmt"

	"github.com/insmtx/Leros/backend/internal/agent"
)

// ModelResolver 解析单次运行的具体模型配置。
type ModelResolver interface {
	ResolveModel(ctx context.Context, req *agent.RequestContext) (*agent.ModelOptions, error)
}

type ModelStep struct {
	Resolver ModelResolver
}

func (ModelStep) Name() string {
	return "model"
}

func (s ModelStep) Run(ctx context.Context, state *State) error {
	return EnsureModelConfig(ctx, state.Request, s.Resolver)
}

// EnsureModelConfig 在需要时将解析后的模型配置应用到 req。
func EnsureModelConfig(ctx context.Context, req *agent.RequestContext, resolver ModelResolver) error {
	if req == nil {
		return fmt.Errorf("request context is required")
	}
	if req.Model.Provider != "" && req.Model.Model != "" && req.Model.APIKey != "" {
		return nil
	}
	if resolver == nil {
		return fmt.Errorf("llm model config is required")
	}
	resolved, err := resolver.ResolveModel(ctx, req)
	if err != nil {
		return err
	}
	if resolved == nil {
		return nil
	}
	req.Model = *resolved
	return nil
}

package steps

import (
	"context"
	"fmt"

	"github.com/insmtx/Leros/backend/internal/agent"
)

type ExecuteStep struct {
	Delegate agent.Runner
}

func (ExecuteStep) Name() string {
	return "execute"
}

func (s ExecuteStep) Run(ctx context.Context, state *State) error {
	if s.Delegate == nil {
		return fmt.Errorf("delegate runner is required")
	}
	result, err := s.Delegate.Run(ctx, state.Request)
	state.Result = result
	state.Err = err
	return nil
}

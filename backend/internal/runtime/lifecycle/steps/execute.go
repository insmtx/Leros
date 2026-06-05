package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/Leros/backend/internal/agent"
)

// DefaultExecuteTimeout is the hard timeout for the runtime execution step.
const DefaultExecuteTimeout = 2 * time.Hour

var executeTimeout = DefaultExecuteTimeout

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
	executeCtx, cancel := context.WithTimeout(ctx, executeTimeout)
	defer cancel()

	result, err := s.Delegate.Run(executeCtx, state.Request)
	state.Result = result
	state.Err = err
	return nil
}

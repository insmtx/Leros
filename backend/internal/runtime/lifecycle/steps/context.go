package steps

import (
	"context"
	"errors"
	"fmt"

	"github.com/insmtx/Leros/backend/internal/agent"
	lifecyclecontext "github.com/insmtx/Leros/backend/internal/runtime/lifecycle/context"
	"github.com/ygpkg/yg-go/logs"
)

type ContextBuilder interface {
	Prepare(ctx context.Context, req *agent.RequestContext) (*agent.RequestContext, error)
}

type ContextStep struct {
	Builder ContextBuilder
}

func (ContextStep) Name() string {
	return "context"
}

func (s ContextStep) Run(ctx context.Context, state *State) error {
	if s.Builder == nil {
		return fmt.Errorf("context builder is required")
	}
	prepared, err := s.Builder.Prepare(ctx, state.Request)
	if err != nil {
		if errors.Is(err, lifecyclecontext.ErrNoPendingSessionMessages) {
			state.Skipped = true
			logs.InfoContextf(ctx, "Agent lifecycle run skipped: run_id=%s trace_id=%s reason=no_pending_messages",
				requestRunID(state.Request), requestTraceID(state.Request))
			return nil
		}
		return err
	}
	prepared.EventSink = state.Journal
	state.Request = prepared
	return nil
}

package steps

import (
	"context"

	lifecyclecontext "github.com/insmtx/Leros/backend/internal/runtime/lifecycle/context"
	"github.com/ygpkg/yg-go/logs"
)

type SessionCompleteStep struct {
	Provider lifecyclecontext.SessionMessageProvider
}

func (SessionCompleteStep) Name() string {
	return "session_complete"
}

func (s SessionCompleteStep) Run(ctx context.Context, state *State) error {
	if s.Provider == nil || state == nil || state.Request == nil || state.Skipped {
		return nil
	}
	if err := s.Provider.CompleteClaimed(ctx, state.Request); err != nil {
		logs.WarnContextf(ctx, "Agent lifecycle complete claimed messages failed: run_id=%s trace_id=%s error=%v",
			state.Request.RunID, state.Request.TraceID, err)
	}
	return nil
}

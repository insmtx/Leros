package steps

import (
	"context"

	lifecyclejournal "github.com/insmtx/Leros/backend/internal/runtime/lifecycle/journal"
)

type PersistStep struct{}

func (PersistStep) Name() string {
	return "persist"
}

func (PersistStep) Run(ctx context.Context, state *State) error {
	if state.Err != nil {
		result, err := lifecyclejournal.EmitFailed(ctx, state.Journal, state.Request, lifecyclejournal.RunPhaseRuntime, state.Err, metadataFromResult(state.Result))
		state.Result = result
		state.Err = err
		return nil
	}
	if err := lifecyclejournal.EmitSucceeded(ctx, state.Journal, state.Request, state.Result); err != nil {
		state.Err = err
		return err
	}
	return nil
}

package steps

import (
	"context"

	"github.com/insmtx/Leros/backend/internal/runtime/events"
	lifecyclejournal "github.com/insmtx/Leros/backend/internal/runtime/lifecycle/journal"
)

type StartEventStep struct{}

func (StartEventStep) Name() string {
	return "start_event"
}

func (StartEventStep) Run(ctx context.Context, state *State) error {
	return lifecyclejournal.AppendLifecycleEvent(ctx, state.Journal, state.Request, events.EventStarted, "")
}

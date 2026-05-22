package steps

import (
	"context"

	lifecyclejournal "github.com/insmtx/Leros/backend/internal/runtime/lifecycle/journal"
)

type JournalStep struct{}

func (JournalStep) Name() string {
	return "journal"
}

func (JournalStep) Run(_ context.Context, state *State) error {
	if state == nil {
		return nil
	}
	state.Journal = lifecyclejournal.NewRunJournal(state.Request, nil)
	if state.Request != nil {
		state.Journal = lifecyclejournal.NewRunJournal(state.Request, state.Request.EventSink)
		state.Request.EventSink = state.Journal
	}
	return nil
}

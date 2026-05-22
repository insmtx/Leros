package lifecycle

import lifecyclejournal "github.com/insmtx/Leros/backend/internal/runtime/lifecycle/journal"

const (
	RunPhasePrepare = lifecyclejournal.RunPhasePrepare
	RunPhaseModel   = lifecyclejournal.RunPhaseModel
	RunPhaseRuntime = lifecyclejournal.RunPhaseRuntime
	RunPhasePanic   = lifecyclejournal.RunPhasePanic
)

func phaseForError(state *RunState, err error) lifecyclejournal.RunPhase {
	if state == nil {
		return lifecyclejournal.RunPhasePrepare
	}
	if state.Request == nil || state.Request.SystemPrompt == "" {
		return lifecyclejournal.RunPhasePrepare
	}
	if state.Result != nil || state.Err != nil {
		return lifecyclejournal.RunPhaseRuntime
	}
	return lifecyclejournal.RunPhaseModel
}

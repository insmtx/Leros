package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/Leros/backend/internal/agent"
)

type NormalizeStep struct{}

func (NormalizeStep) Name() string {
	return "normalize"
}

func (NormalizeStep) Run(_ context.Context, state *State) error {
	if state == nil {
		return nil
	}
	NormalizeRequest(state.OriginalRequest)
	state.Request = state.OriginalRequest
	return nil
}

func NormalizeRequest(req *agent.RequestContext) {
	if req == nil {
		return
	}
	if req.RunID == "" {
		req.RunID = fmt.Sprintf("run_%d", time.Now().UTC().UnixNano())
	}
	if req.Input.Type == "" {
		req.Input.Type = agent.InputTypeMessage
	}
}

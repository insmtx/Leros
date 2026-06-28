// Package codex provides the Codex CLI Runtime type.
package codex

import (
	"context"
	"fmt"

	"github.com/insmtx/Leros/backend/agent"
	"github.com/insmtx/Leros/backend/agent/runtime/externalcli"
)

const (
	// Kind is the canonical runtime kind for Codex CLI.
	Kind = "codex"
)

// Runtime executes requests through Codex CLI.
type Runtime struct {
	driver *externalcli.Driver
}

// New creates a Codex Runtime from shared CLI infrastructure.
func New(driver *externalcli.Driver) *Runtime {
	return &Runtime{driver: driver}
}

func (r *Runtime) Name() string {
	return Kind
}

func (r *Runtime) Execute(
	ctx context.Context,
	request agent.ExecutionRequest,
	observer agent.Observer,
) (agent.ExecutionResult, error) {
	if r == nil || r.driver == nil {
		return agent.ExecutionResult{}, fmt.Errorf("codex runtime is not initialized")
	}
	return r.driver.Execute(ctx, request, observer)
}

var _ agent.Runtime = (*Runtime)(nil)

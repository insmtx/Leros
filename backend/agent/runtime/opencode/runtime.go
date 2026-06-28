// Package opencode provides the OpenCode Runtime type.
package opencode

import (
	"context"
	"fmt"

	"github.com/insmtx/Leros/backend/agent"
	"github.com/insmtx/Leros/backend/agent/runtime/externalcli"
)

const (
	// Kind is the canonical runtime kind for OpenCode.
	Kind = "opencode"
)

// Runtime executes requests through OpenCode.
type Runtime struct {
	driver *externalcli.Driver
}

// New creates an OpenCode Runtime from shared CLI infrastructure.
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
		return agent.ExecutionResult{}, fmt.Errorf("opencode runtime is not initialized")
	}
	return r.driver.Execute(ctx, request, observer)
}

var _ agent.Runtime = (*Runtime)(nil)

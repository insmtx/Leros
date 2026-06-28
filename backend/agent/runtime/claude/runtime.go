// Package claude provides the Claude Code Runtime.
package claude

import (
	"context"
	"fmt"

	"github.com/insmtx/Leros/backend/agent"
	"github.com/insmtx/Leros/backend/agent/runtime/externalcli"
)

const (
	// Kind is the canonical runtime kind for Claude Code.
	Kind = "claude"
)

// Runtime executes requests through Claude Code.
type Runtime struct {
	driver *externalcli.Driver
}

// New creates a Claude Runtime from shared CLI infrastructure.
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
		return agent.ExecutionResult{}, fmt.Errorf("claude runtime is not initialized")
	}
	return r.driver.Execute(ctx, request, observer)
}

var _ agent.Runtime = (*Runtime)(nil)

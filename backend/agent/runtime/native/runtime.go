// Package native provides the built-in Leros (Eino) Runtime.
package native

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/Leros/backend/agent"
	"github.com/insmtx/Leros/backend/agent/runtime/events"
	"github.com/insmtx/Leros/backend/agent/runtime/externalcli"
	engines "github.com/insmtx/Leros/backend/agent/runtime/provider"
)

const (
	// Kind is the canonical runtime kind for the built-in Leros engine.
	Kind = "leros"
)

// Runtime executes requests directly through the in-process Eino runner.
type Runtime struct {
	runner *Runner
}

// New creates the native Runtime.
func New() (*Runtime, error) {
	runner, err := NewRunner(context.Background())
	if err != nil {
		return nil, fmt.Errorf("create native runner: %w", err)
	}
	return &Runtime{runner: runner}, nil
}

func (r *Runtime) Name() string {
	return Kind
}

func (r *Runtime) Execute(
	ctx context.Context,
	request agent.ExecutionRequest,
	observer agent.Observer,
) (agent.ExecutionResult, error) {
	if r == nil || r.runner == nil {
		return agent.ExecutionResult{}, fmt.Errorf("native runtime is not initialized")
	}
	if strings.TrimSpace(request.ExecutionID) == "" {
		return agent.ExecutionResult{}, fmt.Errorf("execution id is required")
	}
	eventChannel, err := r.runner.Run(ctx, engines.RunRequest{
		ExecutionID:  request.ExecutionID,
		SessionID:    request.SessionKey,
		WorkDir:      request.Filesystem.WorkDir,
		TaskDir:      request.Filesystem.TaskDir,
		SystemPrompt: request.SystemPrompt,
		Prompt:       request.Prompt,
		Messages:     append([]agent.Message(nil), request.Messages...),
		Tools:        append([]agent.Tool(nil), request.Tools...),
		AllowedTools: append([]string(nil), request.Policy.AllowedTools...),
		TraceID:      request.TraceID,
		SessionKey:   request.SessionKey,
		Model: engines.ModelConfig{
			Provider: request.Model.Provider,
			Model:    request.Model.Model,
			APIKey:   request.Model.APIKey,
			BaseURL:  request.Model.BaseURL,
		},
	})
	if err != nil {
		return agent.ExecutionResult{}, err
	}
	var sink events.Sink = events.NewNoopSink()
	if observer != nil {
		sink = observer
	}
	consumed, err := externalcli.ConsumeEvents(
		ctx,
		sink,
		&engines.RunHandle{Events: eventChannel},
		request.ExecutionID,
		request.TraceID,
		nil,
		nil,
	)
	if err != nil {
		return agent.ExecutionResult{}, err
	}
	return agent.ExecutionResult{
		Message: strings.TrimSpace(consumed.Message),
		Usage:   consumed.Usage,
	}, nil
}

var _ agent.Runtime = (*Runtime)(nil)

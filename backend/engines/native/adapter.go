// Package native implements the built-in Eino-backed Leros engine.
package native

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/insmtx/Leros/backend/engines"
	"github.com/insmtx/Leros/backend/internal/agent"
	"github.com/insmtx/Leros/backend/internal/runtime/deps"
	"github.com/insmtx/Leros/backend/internal/runtime/events"
	"github.com/insmtx/Leros/backend/pkg/leros"
)

// EngineName is the registry name for the native engine.
const EngineName = agent.RuntimeKindLeros // "leros"

// Adapter implements engines.Engine for the in-process Eino runtime.
type Adapter struct {
	runner *Runner
	env    *deps.Container

	mu       sync.RWMutex
	skillDir string
}

// NewAdapter creates a native engine adapter.
func NewAdapter(env *deps.Container) (*Adapter, error) {
	runner, err := NewRunner(context.Background(), env)
	if err != nil {
		return nil, fmt.Errorf("create native runner: %w", err)
	}

	skillDir, err := leros.SkillsDir()
	if err != nil {
		skillDir = "" // best-effort
	}

	return &Adapter{
		runner:   runner,
		env:      env,
		skillDir: skillDir,
	}, nil
}

// Prepare satisfies engines.Engine.
func (a *Adapter) Prepare(_ context.Context, _ engines.PrepareRequest) error {
	return nil
}

// RegisterMCP satisfies engines.Engine.
func (a *Adapter) RegisterMCP(_ context.Context, _ engines.MCPServerConfig) error {
	return nil
}

// GetSkillDir satisfies engines.Engine.
func (a *Adapter) GetSkillDir() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.skillDir
}

// Run satisfies engines.Engine by executing the Eino flow and publishing
// events through a channel so the externalcli bridge can consume them.
func (a *Adapter) Run(ctx context.Context, req engines.RunRequest) (*engines.RunHandle, error) {
	if a == nil || a.runner == nil {
		return nil, fmt.Errorf("native engine is not initialized")
	}

	eventsCh := make(chan events.Event, 256)

	go func() {
		defer close(eventsCh)
		a.execute(ctx, req, eventsCh)
	}()

	return &engines.RunHandle{
		Process:   &noopProcess{pid: os.Getpid()},
		Events:    eventsCh,
		Responder: &noopResponder{},
	}, nil
}

// execute runs the Eino flow and feeds events into the channel.
func (a *Adapter) execute(ctx context.Context, req engines.RunRequest, eventsCh chan<- events.Event) {
	// Emit run.started.
	sendEvent(eventsCh, events.EventStarted, req.ExecutionID)

	// Build a minimal agent.RequestContext from the run request.
	agentReq := a.buildAgentRequest(req)

	// Wire a channel-backed EventSink so stream deltas and tool events
	// flow through the channel consumed by externalcli.consumeEvents.
	agentReq.EventSink = events.SinkFunc(func(ctx context.Context, event *events.Event) error {
		if event != nil {
			select {
			case eventsCh <- *event:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	result, err := a.runner.Run(ctx, agentReq)
	if err != nil {
		sendEvent(eventsCh, events.EventFailed, fmt.Sprintf("run_id=%s error=%s", req.ExecutionID, err.Error()))
		return
	}

	// Emit message.result with final message and usage.
	payload, _ := json.Marshal(events.MessageResultPayload{
		Message: result.Message,
		Usage:   result.Usage,
	})
	eventsCh <- events.Event{
		Type:    events.EventResult,
		Payload: events.RawPayload(payload),
		Content: result.Message,
	}

	// Emit run.completed.
	sendEvent(eventsCh, events.EventCompleted, req.ExecutionID)
}

// buildAgentRequest constructs a minimal agent.RequestContext from an
// engines.RunRequest so the existing Eino runner can execute it.
func (a *Adapter) buildAgentRequest(req engines.RunRequest) *agent.RequestContext {
	return &agent.RequestContext{
		RunID:   req.ExecutionID,
		TraceID: req.SessionID,
		Conversation: agent.ConversationContext{
			ID: req.SessionID,
		},
		Input: agent.InputContext{
			Type: agent.InputTypeMessage,
			Messages: []agent.InputMessage{
				{Role: "user", Content: req.Prompt},
			},
		},
		Runtime: agent.RuntimeOptions{
			Kind:    EngineName,
			WorkDir: req.WorkDir,
		},
		Model: agent.ModelOptions{
			Provider: req.Model.Provider,
			Model:    req.Model.Model,
			APIKey:   req.Model.APIKey,
			BaseURL:  req.Model.BaseURL,
		},
		SystemPrompt: req.SystemPrompt,
	}
}

func sendEvent(eventsCh chan<- events.Event, eventType events.EventType, content string) {
	select {
	case eventsCh <- events.Event{Type: eventType, Content: content}:
	default:
	}
}

// noopProcess satisfies engines.Process for the in-process engine.
type noopProcess struct {
	pid int
}

func (p *noopProcess) PID() int    { return p.pid }
func (p *noopProcess) Stop() error { return nil }

// noopResponder satisfies engines.ApprovalResponder.
type noopResponder struct{}

func (r *noopResponder) WriteDecision(string, string) error { return nil }

var _ engines.Engine = (*Adapter)(nil)

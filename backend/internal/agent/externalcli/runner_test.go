package externalcli

import (
	"context"
	"testing"

	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
)

func TestRunnerAdaptsEngineResult(t *testing.T) {
	SetDefaultProviderSessionStore(NewInMemoryProviderSessionStore())
	engine := &fakeEngine{result: "done"}
	runner, err := NewRunner("fake", engine, nil)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	result, err := runner.Run(context.Background(), &agent.RequestContext{
		RunID: "run_cli",
		Input: agent.InputContext{
			Type: agent.InputTypeMessage,
			Text: "hello",
		},
		Runtime: agent.RuntimeOptions{WorkDir: "/tmp"},
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Status != agent.RunStatusCompleted {
		t.Fatalf("expected completed, got %s", result.Status)
	}
	if result.Message != "done" {
		t.Fatalf("expected extracted result, got %q", result.Message)
	}
	if engine.runReq.WorkDir != "/tmp" {
		t.Fatalf("expected work dir /tmp, got %q", engine.runReq.WorkDir)
	}
	if engine.runReq.Prompt == "" {
		t.Fatal("expected prompt to be built")
	}
}

func TestRunnerStoresProviderSessionAndResumes(t *testing.T) {
	store := NewInMemoryProviderSessionStore()
	SetDefaultProviderSessionStore(store)
	engine := &fakeEngine{
		result:            "done",
		providerSessionID: "provider-session-1",
	}
	runner, err := NewRunner("codex", engine, nil)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	req := &agent.RequestContext{
		RunID: "run_first",
		Conversation: agent.ConversationContext{
			ID: "internal-session-1",
		},
		Assistant: agent.AssistantContext{
			ID: "assistant-1",
		},
		Input: agent.InputContext{
			Type: agent.InputTypeMessage,
			Text: "hello",
		},
		Runtime: agent.RuntimeOptions{WorkDir: "/tmp"},
	}

	if _, err := runner.Run(context.Background(), req); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if engine.runReq.Resume {
		t.Fatal("first run should not resume")
	}
	if engine.runReq.SessionID != "" {
		t.Fatalf("first codex run should not preallocate provider session, got %q", engine.runReq.SessionID)
	}

	req.RunID = "run_second"
	if _, err := runner.Run(context.Background(), req); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if !engine.runReq.Resume {
		t.Fatal("second run should resume")
	}
	if engine.runReq.SessionID != "provider-session-1" {
		t.Fatalf("expected provider session id, got %q", engine.runReq.SessionID)
	}
}

func TestRunnerPreallocatesClaudeProviderSession(t *testing.T) {
	SetDefaultProviderSessionStore(NewInMemoryProviderSessionStore())
	engine := &fakeEngine{result: "done"}
	runner, err := NewRunner(engines.EngineClaude, engine, nil)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	_, err = runner.Run(context.Background(), &agent.RequestContext{
		RunID: "run_claude",
		Conversation: agent.ConversationContext{
			ID: "internal-session-claude",
		},
		Input: agent.InputContext{
			Type: agent.InputTypeMessage,
			Text: "hello",
		},
		Runtime: agent.RuntimeOptions{WorkDir: "/tmp"},
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if engine.runReq.Resume {
		t.Fatal("first claude run should not resume")
	}
	if engine.runReq.SessionID == "" {
		t.Fatal("expected preallocated claude provider session id")
	}
}

type fakeEngine struct {
	runReq            engines.RunRequest
	result            string
	providerSessionID string
}

func (e *fakeEngine) Prepare(_ context.Context, _ engines.PrepareRequest) error {
	return nil
}

func (e *fakeEngine) RegisterMCP(_ context.Context, _ engines.MCPServerConfig) error {
	return nil
}

func (e *fakeEngine) Run(_ context.Context, req engines.RunRequest) (*engines.RunHandle, error) {
	e.runReq = req
	events := make(chan engines.Event, 4)
	events <- engines.Event{Type: engines.EventStarted}
	if e.providerSessionID != "" {
		events <- engines.Event{Type: engines.EventProviderSessionStarted, Content: e.providerSessionID}
	}
	events <- engines.Event{Type: engines.EventResult, Content: e.result}
	events <- engines.Event{Type: engines.EventDone}
	close(events)
	return &engines.RunHandle{
		Events: events,
	}, nil
}

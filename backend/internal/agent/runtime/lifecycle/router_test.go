package lifecycle

import (
	"context"
	"strings"
	"testing"

	"github.com/insmtx/Leros/backend/internal/agent"
)

func TestRuntimeRouterWrapsRegisteredRuntime(t *testing.T) {
	capture := &captureRuntime{}
	router := NewRuntimeRouter(agent.RuntimeKindLeros, NewContextBuilder(ContextBuilder{
		BaseSystemPrompt: "base prompt",
	}))
	if err := router.Register(agent.RuntimeKindLeros, capture); err != nil {
		t.Fatalf("register runner: %v", err)
	}

	result, err := router.Run(context.Background(), &agent.RequestContext{
		Input: agent.InputContext{
			Type: agent.InputTypeMessage,
			Text: "hello",
		},
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result == nil || result.Message != "ok" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if capture.run == nil {
		t.Fatalf("expected lifecycle request")
	}
	if !strings.Contains(capture.run.SystemPrompt, "base prompt") {
		t.Fatalf("expected lifecycle system prompt, got %q", capture.run.SystemPrompt)
	}
	if capture.run.Input.Text != "hello" {
		t.Fatalf("expected request input, got %q", capture.run.Input.Text)
	}
}

type captureRuntime struct {
	run *agent.RequestContext
}

func (r *captureRuntime) Run(_ context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	r.run = req
	return &agent.RunResult{
		Status:  agent.RunStatusCompleted,
		Message: "ok",
	}, nil
}

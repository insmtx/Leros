package lifecycle

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/insmtx/Leros/backend/internal/agent"
	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
)

func TestRunnerEmitsSuccessResultAndUsageThroughSink(t *testing.T) {
	sink := &recordingSink{}
	runner := newRunner(successRuntime{}, NewContextBuilder(ContextBuilder{
		BaseSystemPrompt: "base",
	}), nil)

	result, err := runner.Run(context.Background(), lifecycleTestRequest(sink))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result == nil || result.Status != agent.RunStatusCompleted {
		t.Fatalf("expected completed result, got %+v", result)
	}

	got := sink.Events()
	if len(got) != 4 {
		t.Fatalf("expected started, usage, result, completed events, got %d", len(got))
	}
	expectedTypes := []events.EventType{
		events.EventStarted,
		events.EventUsage,
		events.EventResult,
		events.EventCompleted,
	}
	for i, expected := range expectedTypes {
		if got[i].Type != expected {
			t.Fatalf("event %d: expected %s, got %s", i, expected, got[i].Type)
		}
		if i > 0 && got[i].Seq != got[i-1].Seq+1 {
			t.Fatalf("event %d: expected contiguous seq, got prev=%d current=%d", i, got[i-1].Seq, got[i].Seq)
		}
	}
	if !strings.Contains(got[1].Content, `"total_tokens":3`) {
		t.Fatalf("expected usage payload, got %q", got[1].Content)
	}
	if got[2].Content != "final answer" {
		t.Fatalf("expected final result content, got %q", got[2].Content)
	}
	if got[3].Content != "final answer" {
		t.Fatalf("expected completed content, got %q", got[3].Content)
	}
}

func TestRunnerEmitsFailureThroughSink(t *testing.T) {
	sink := &recordingSink{}
	runner := newRunner(&errorRuntime{err: errors.New("runtime unavailable")}, NewContextBuilder(ContextBuilder{
		BaseSystemPrompt: "base",
	}), nil)

	result, err := runner.Run(context.Background(), lifecycleTestRequest(sink))
	if err == nil {
		t.Fatal("expected run error")
	}
	if result == nil {
		t.Fatal("expected failed result")
	}
	if result.Status != agent.RunStatusFailed {
		t.Fatalf("expected failed status, got %s", result.Status)
	}
	if result.Error != "runtime unavailable" {
		t.Fatalf("expected error message, got %q", result.Error)
	}
	if phase := result.Metadata["phase"]; phase != string(RunPhaseRuntime) {
		t.Fatalf("expected runtime phase, got %v", phase)
	}

	got := sink.Events()
	if len(got) != 2 {
		t.Fatalf("expected started and terminal events, got %d", len(got))
	}
	if got[0].Type != events.EventStarted {
		t.Fatalf("expected started event, got %s", got[0].Type)
	}
	if got[1].Type != events.EventFailed {
		t.Fatalf("expected failed event, got %s", got[1].Type)
	}
	if got[1].Seq != got[0].Seq+1 {
		t.Fatalf("expected failure event seq to continue, got started=%d failed=%d", got[0].Seq, got[1].Seq)
	}
	if got[1].Content != "runtime unavailable" {
		t.Fatalf("expected event error content, got %q", got[1].Content)
	}
}

func TestRunnerEmitsCancelledThroughSink(t *testing.T) {
	sink := &recordingSink{}
	runner := newRunner(&errorRuntime{err: context.DeadlineExceeded}, NewContextBuilder(ContextBuilder{
		BaseSystemPrompt: "base",
	}), nil)

	result, err := runner.Run(context.Background(), lifecycleTestRequest(sink))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error, got %v", err)
	}
	if result == nil || result.Status != agent.RunStatusCancelled {
		t.Fatalf("expected cancelled result, got %+v", result)
	}
	got := sink.Events()
	if len(got) != 2 {
		t.Fatalf("expected started and terminal events, got %d", len(got))
	}
	if got[0].Type != events.EventStarted {
		t.Fatalf("expected started event, got %s", got[0].Type)
	}
	if got[1].Type != events.EventCancelled {
		t.Fatalf("expected cancelled event, got %s", got[1].Type)
	}
}

func TestRunnerRecoversPanicThroughSink(t *testing.T) {
	sink := &recordingSink{}
	runner := newRunner(panicRuntime{}, NewContextBuilder(ContextBuilder{
		BaseSystemPrompt: "base",
	}), nil)

	result, err := runner.Run(context.Background(), lifecycleTestRequest(sink))
	if err == nil {
		t.Fatal("expected panic error")
	}
	if result == nil || result.Status != agent.RunStatusFailed {
		t.Fatalf("expected failed result, got %+v", result)
	}
	if phase := result.Metadata["phase"]; phase != string(RunPhasePanic) {
		t.Fatalf("expected panic phase, got %v", phase)
	}
	got := sink.Events()
	if len(got) != 2 {
		t.Fatalf("expected started and terminal events, got %d", len(got))
	}
	if got[0].Type != events.EventStarted {
		t.Fatalf("expected started event, got %s", got[0].Type)
	}
	if got[1].Type != events.EventFailed {
		t.Fatalf("expected failed event, got %s", got[1].Type)
	}
}

func lifecycleTestRequest(sink events.Sink) *agent.RequestContext {
	return &agent.RequestContext{
		RunID: "run_lifecycle_error",
		Input: agent.InputContext{
			Type: agent.InputTypeMessage,
			Text: "hello",
		},
		Model: agent.ModelOptions{
			Provider: "test",
			Model:    "test-model",
			APIKey:   "test-key",
		},
		EventSink: sink,
	}
}

type errorRuntime struct {
	err error
}

func (r *errorRuntime) Run(context.Context, *agent.RequestContext) (*agent.RunResult, error) {
	return &agent.RunResult{
		Status: agent.RunStatusFailed,
		Metadata: map[string]any{
			"runtime": "test",
		},
	}, r.err
}

type successRuntime struct{}

func (successRuntime) Run(_ context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	return &agent.RunResult{
		RunID:   req.RunID,
		TraceID: req.TraceID,
		Status:  agent.RunStatusCompleted,
		Message: "final answer",
		Usage: &events.UsagePayload{
			InputTokens:  1,
			OutputTokens: 2,
			TotalTokens:  3,
		},
	}, nil
}

type panicRuntime struct{}

func (panicRuntime) Run(context.Context, *agent.RequestContext) (*agent.RunResult, error) {
	panic("runtime exploded")
}

type recordingSink struct {
	events []*events.Event
}

func (s *recordingSink) Emit(_ context.Context, event *events.Event) error {
	if event == nil {
		return nil
	}
	copied := *event
	s.events = append(s.events, &copied)
	return nil
}

func (s *recordingSink) Events() []*events.Event {
	return append([]*events.Event{}, s.events...)
}

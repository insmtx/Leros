package agent

import (
	"context"
	"errors"
	"testing"
)

type runtimeStub struct {
	name   string
	result ExecutionResult
	err    error
}

func (r runtimeStub) Name() string { return r.name }

func (r runtimeStub) Execute(context.Context, ExecutionRequest, Observer) (ExecutionResult, error) {
	return r.result, r.err
}

type observerRecorder struct {
	events []*Event
	errAt  EventType
}

func (o *observerRecorder) Emit(_ context.Context, event *Event) error {
	if event != nil {
		o.events = append(o.events, event)
		if event.Type == o.errAt {
			return errors.New("observer failed")
		}
	}
	return nil
}

func TestExecutorUsesDefaultRuntimeAndEmitsLifecycle(t *testing.T) {
	registry := NewRegistry()
	registry.Register("native", runtimeStub{name: "native", result: ExecutionResult{Message: "done"}})
	registry.SetDefault("native")
	observer := &observerRecorder{}

	result, err := NewExecutor(registry).Execute(context.Background(), ExecutionRequest{
		ExecutionID: "run-1",
		TraceID:     "trace-1",
	}, observer)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Message != "done" {
		t.Fatalf("result message = %q, want done", result.Message)
	}
	if len(observer.events) != 2 ||
		observer.events[0].Type != "execution.started" ||
		observer.events[1].Type != "execution.completed" {
		t.Fatalf("lifecycle events = %#v", observer.events)
	}
}

func TestExecutorEmitsCancelled(t *testing.T) {
	registry := NewRegistry()
	registry.Register("native", runtimeStub{name: "native", err: context.Canceled})
	registry.SetDefault("native")
	observer := &observerRecorder{}

	_, err := NewExecutor(registry).Execute(context.Background(), ExecutionRequest{ExecutionID: "run-1"}, observer)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute() error = %v, want context.Canceled", err)
	}
	if len(observer.events) != 2 || observer.events[1].Type != "execution.cancelled" {
		t.Fatalf("lifecycle events = %#v", observer.events)
	}
}

func TestExecutorStopsOnObserverError(t *testing.T) {
	registry := NewRegistry()
	registry.Register("native", runtimeStub{name: "native", result: ExecutionResult{}})
	registry.SetDefault("native")
	observer := &observerRecorder{errAt: "execution.started"}

	_, err := NewExecutor(registry).Execute(context.Background(), ExecutionRequest{ExecutionID: "run-1"}, observer)
	if err == nil {
		t.Fatal("Execute() error = nil, want observer error")
	}
}

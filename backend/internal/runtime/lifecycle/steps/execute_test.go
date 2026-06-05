package steps

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/insmtx/Leros/backend/internal/agent"
)

func TestExecuteStepAppliesRuntimeTimeout(t *testing.T) {
	restore := setExecuteTimeoutForTest(10 * time.Millisecond)
	defer restore()

	state := &State{Request: &agent.RequestContext{}}
	err := ExecuteStep{Delegate: blockingRunner{}}.Run(context.Background(), state)
	if err != nil {
		t.Fatalf("execute step should store delegate error in state, got %v", err)
	}
	if state.Err == nil || !errors.Is(state.Err, context.DeadlineExceeded) {
		t.Fatalf("expected state err deadline exceeded, got %v", state.Err)
	}
	if state.Result == nil || state.Result.Status != agent.RunStatusCancelled {
		t.Fatalf("expected cancelled result, got %#v", state.Result)
	}
}

func TestExecuteStepPreservesCallerCancellation(t *testing.T) {
	restore := setExecuteTimeoutForTest(time.Hour)
	defer restore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	state := &State{Request: &agent.RequestContext{}}
	err := ExecuteStep{Delegate: blockingRunner{}}.Run(ctx, state)
	if err != nil {
		t.Fatalf("execute step should store delegate error in state, got %v", err)
	}
	if state.Err == nil || !errors.Is(state.Err, context.Canceled) {
		t.Fatalf("expected state err context canceled, got %v", state.Err)
	}
	if errors.Is(state.Err, context.DeadlineExceeded) {
		t.Fatalf("did not expect deadline exceeded for caller cancellation")
	}
}

func setExecuteTimeoutForTest(timeout time.Duration) func() {
	previous := executeTimeout
	executeTimeout = timeout
	return func() {
		executeTimeout = previous
	}
}

type blockingRunner struct{}

func (blockingRunner) Run(ctx context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	<-ctx.Done()
	runID := ""
	if req != nil {
		runID = req.RunID
	}
	return &agent.RunResult{
		RunID:  runID,
		Status: agent.RunStatusCancelled,
	}, ctx.Err()
}

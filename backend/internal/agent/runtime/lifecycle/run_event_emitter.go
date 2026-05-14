package lifecycle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/insmtx/Leros/backend/internal/agent"
	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
	"github.com/ygpkg/yg-go/logs"
)

// RunPhase 表示 Agent 运行错误的生命周期阶段。
type RunPhase string

const (
	RunPhasePrepare RunPhase = "prepare" // 准备阶段：初始化、参数校验等
	RunPhaseModel   RunPhase = "model"   // 模型阶段：调用 LLM 生成结果
	RunPhaseRuntime RunPhase = "runtime" // 运行时阶段：执行技能、处理结果
	RunPhasePanic   RunPhase = "panic"   // 异常阶段：程序发生 panic
)

type sequenceFunc func() int64

// RunEventEmitter emits lifecycle-owned run events through the request sink.
type RunEventEmitter struct {
	nextSeq sequenceFunc
}

// NewRunEventEmitter creates the shared lifecycle event emitter.
func NewRunEventEmitter() *RunEventEmitter {
	return &RunEventEmitter{}
}

func (e *RunEventEmitter) setSequence(nextSeq sequenceFunc) {
	if e == nil {
		return
	}
	e.nextSeq = nextSeq
}

// EmitStarted emits the lifecycle-owned run start event.
func (e *RunEventEmitter) EmitStarted(ctx context.Context, req *agent.RequestContext) error {
	if req != nil {
		ensureErrorRunDefaults(req)
	}
	return e.emit(ctx, req, events.EventStarted, "")
}

// EmitSucceeded emits the lifecycle-owned final result events for a completed run.
func (e *RunEventEmitter) EmitSucceeded(ctx context.Context, req *agent.RequestContext, result *agent.RunResult) error {
	if req == nil || result == nil || result.Status != agent.RunStatusCompleted {
		return nil
	}
	if result.Usage != nil {
		if err := e.emit(ctx, req, events.EventUsage, eventContentJSON(result.Usage)); err != nil {
			return err
		}
	}
	if result.Message != "" {
		if err := e.emit(ctx, req, events.EventResult, result.Message); err != nil {
			return err
		}
	}
	return e.emit(ctx, req, events.EventCompleted, result.Message)
}

// EmitFailed converts err into a final RunResult and emits one terminal failure event.
func (e *RunEventEmitter) EmitFailed(ctx context.Context, req *agent.RequestContext, startedAt time.Time, phase RunPhase, err error, metadata map[string]any) (*agent.RunResult, error) {
	if err == nil {
		return nil, nil
	}

	if req != nil {
		ensureErrorRunDefaults(req)
	}

	status, eventType := failureStatus(err)
	message := err.Error()
	logs.WarnContextf(ctx, "Agent run failed: run_id=%s trace_id=%s task_id=%s runtime=%s phase=%s status=%s error=%v",
		requestRunID(req),
		requestTraceID(req),
		requestTaskID(req),
		requestRuntimeKind(req),
		phase,
		status,
		err,
	)

	emitErr := e.emit(ctx, req, eventType, message)
	if emitErr != nil {
		logs.WarnContextf(ctx, "Agent run failure event emit failed: run_id=%s trace_id=%s phase=%s error=%v",
			requestRunID(req), requestTraceID(req), phase, emitErr)
	}

	result := &agent.RunResult{
		RunID:       requestRunID(req),
		TraceID:     requestTraceID(req),
		Status:      status,
		Error:       message,
		StartedAt:   startedAt,
		CompletedAt: time.Now().UTC(),
		Metadata:    metadataWithLifecyclePhase(metadata, phase),
	}
	return result, err
}

// EmitPanic converts a recovered panic value into the standard runtime failure path.
func (e *RunEventEmitter) EmitPanic(ctx context.Context, req *agent.RequestContext, startedAt time.Time, recovered any) (*agent.RunResult, error) {
	if recovered == nil {
		return nil, nil
	}
	err, ok := recovered.(error)
	if !ok {
		err = fmt.Errorf("%v", recovered)
	}
	return e.EmitFailed(ctx, req, startedAt, RunPhasePanic, fmt.Errorf("agent runtime panic: %w", err), nil)
}

func (e *RunEventEmitter) emit(ctx context.Context, req *agent.RequestContext, eventType events.EventType, message string) error {
	if req == nil || req.EventSink == nil {
		return nil
	}
	seq := int64(1)
	if e != nil && e.nextSeq != nil {
		seq = e.nextSeq()
	}
	return req.EventSink.Emit(ctx, &events.Event{
		ID:        fmt.Sprintf("%s:%d", req.RunID, seq),
		RunID:     req.RunID,
		TraceID:   req.TraceID,
		Seq:       seq,
		Type:      eventType,
		CreatedAt: time.Now().UTC(),
		Content:   message,
	})
}

func failureStatus(err error) (agent.RunStatus, events.EventType) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return agent.RunStatusCancelled, events.EventCancelled
	}
	return agent.RunStatusFailed, events.EventFailed
}

func ensureErrorRunDefaults(req *agent.RequestContext) {
	if req.RunID == "" {
		req.RunID = fmt.Sprintf("run_%d", time.Now().UTC().UnixNano())
	}
}

func metadataWithLifecyclePhase(metadata map[string]any, phase RunPhase) map[string]any {
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["phase"] = string(phase)
	return metadata
}

func eventContentJSON(value interface{}) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(encoded)
}

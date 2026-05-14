// Package externalcli adapts external agent CLIs to the Leros agent.Runner boundary.
package externalcli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/Leros/backend/engines"
	"github.com/insmtx/Leros/backend/internal/agent"
	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
	"github.com/ygpkg/yg-go/logs"
)

// Runner 通过外部 Agent CLI 引擎执行 Leros 请求。
type Runner struct {
	name         string
	engine       engines.Engine
	sessionStore ProviderSessionStore
}

// NewRunner 创建基于外部 CLI 引擎的 Leros runner。
func NewRunner(name string, engine engines.Engine) (*Runner, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("runtime name is required")
	}
	if engine == nil {
		return nil, fmt.Errorf("runtime %q engine is nil", name)
	}
	return &Runner{
		name:         name,
		engine:       engine,
		sessionStore: DefaultProviderSessionStore(),
	}, nil
}

// SetSessionStore replaces the provider session store used for external CLI resumes.
func (r *Runner) SetSessionStore(store ProviderSessionStore) {
	if r == nil || store == nil {
		return
	}
	r.sessionStore = store
}

// Run 直接通过外部 CLI 执行标准化请求；统一生命周期入口应优先使用 lifecycle.Runner。
func (r *Runner) Run(ctx context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	startedAt := time.Now().UTC()
	if r == nil || r.engine == nil {
		return nil, fmt.Errorf("external CLI runtime is not initialized")
	}
	if req == nil {
		return nil, fmt.Errorf("request context is required")
	}
	ensureRunDefaults(req)

	emitter := events.NewEmitter(req.RunID, req.TraceID, sinkForRequest(req))

	workDir := strings.TrimSpace(req.Runtime.WorkDir)
	if workDir == "" {
		workDir = "."
	}
	if err := r.engine.Prepare(ctx, engines.PrepareRequest{WorkDir: workDir}); err != nil {
		return r.failedResult(req, startedAt, err, failureMetadata(workDir)), err
	}

	sessionPlan := r.resolveProviderSession(ctx, req, workDir)
	handle, err := r.engine.Run(ctx, engines.RunRequest{
		ExecutionID: req.RunID,
		SessionID:   sessionPlan.ProviderSessionID,
		Resume:      sessionPlan.Resume,
		WorkDir:     workDir,
		Prompt:      buildPrompt(req),
		Model:       modelForRequest(req),
	})
	if err != nil {
		return r.failedResult(req, startedAt, err, failureMetadata(workDir)), err
	}

	if handle != nil && handle.Process != nil {
		logs.InfoContextf(ctx, "External runtime %s started with pid %d", r.name, handle.Process.PID())
	}

	consumeResult, err := consumeEvents(ctx, emitter, handle)
	if err != nil {
		r.markProviderSessionFailed(ctx, sessionPlan, err)
		return r.failedResult(req, startedAt, err, failureMetadata(workDir)), err
	}
	r.persistProviderSession(ctx, sessionPlan, consumeResult.ProviderSessionID)

	result := &agent.RunResult{
		RunID:       req.RunID,
		TraceID:     req.TraceID,
		Status:      agent.RunStatusCompleted,
		Message:     strings.TrimSpace(consumeResult.Message),
		StartedAt:   startedAt,
		CompletedAt: time.Now().UTC(),
		Metadata: map[string]any{
			"runtime":     r.name,
			"work_dir":    workDir,
			"resume":      sessionPlan.Resume,
			"session_id":  sessionPlan.InternalSessionID,
			"provider_id": firstNonEmptyString(sessionPlan.ProviderSessionID, consumeResult.ProviderSessionID),
		},
	}
	return result, nil
}

type consumeResult struct {
	Message           string
	ProviderSessionID string
}

func consumeEvents(ctx context.Context, emitter *events.Emitter, handle *engines.RunHandle) (consumeResult, error) {
	if handle == nil || handle.Events == nil {
		return consumeResult{}, nil
	}
	var result strings.Builder
	resultSeen := false
	consumed := consumeResult{}
	for event := range handle.Events {
		switch event.Type {
		case events.EventStarted:
			continue
		case engines.EventProviderSessionStarted:
			if strings.TrimSpace(event.Content) != "" {
				consumed.ProviderSessionID = strings.TrimSpace(event.Content)
			}
		case events.EventResult:
			if strings.TrimSpace(event.Content) != "" {
				result.Reset()
				result.WriteString(event.Content)
				resultSeen = true
			}
		case events.EventCompleted:
			consumed.Message = result.String()
			return consumed, nil
		case events.EventFailed:
			if strings.TrimSpace(event.Content) == "" {
				consumed.Message = result.String()
				return consumed, fmt.Errorf("external runtime failed")
			}
			consumed.Message = result.String()
			return consumed, fmt.Errorf("%s", event.Content)
		case events.EventMessageDelta:
			if strings.TrimSpace(event.Content) != "" {
				_ = emitter.Emit(ctx, &events.Event{
					Type:    events.EventMessageDelta,
					Content: event.Content,
				})
				if !resultSeen {
					result.WriteString(event.Content)
				}
			}
		case events.EventToolCallStarted, events.EventToolCallCompleted, events.EventToolCallFailed:
			_ = emitter.Emit(ctx, &events.Event{
				Type:    event.Type,
				Content: event.Content,
			})
		default:
			if strings.TrimSpace(event.Content) != "" {
				if !resultSeen {
					result.WriteString(event.Content)
				}
			}
		}
	}
	consumed.Message = result.String()
	return consumed, nil
}

type providerSessionPlan struct {
	InternalSessionID string
	ProviderSessionID string
	Resume            bool
	Key               ProviderSessionKey
}

func (r *Runner) resolveProviderSession(ctx context.Context, req *agent.RequestContext, workDir string) providerSessionPlan {
	internalSessionID := internalSessionIDFromRequest(req)
	plan := providerSessionPlan{
		InternalSessionID: internalSessionID,
		Key: ProviderSessionKey{
			InternalSessionID: internalSessionID,
			Provider:          r.name,
			WorkDir:           workDir,
			AssistantID:       req.Assistant.ID,
		},
	}
	if internalSessionID == "" || r.sessionStore == nil {
		return plan
	}
	binding, err := r.sessionStore.Get(ctx, plan.Key)
	if err != nil {
		logs.WarnContextf(ctx, "Resolve provider session failed: provider=%s session=%s error=%v", r.name, internalSessionID, err)
		return plan
	}
	if binding != nil && strings.TrimSpace(binding.ProviderSessionID) != "" && binding.Status != providerSessionStatusFailed {
		plan.ProviderSessionID = strings.TrimSpace(binding.ProviderSessionID)
		plan.Resume = true
		return plan
	}
	return plan
}

func (r *Runner) persistProviderSession(ctx context.Context, plan providerSessionPlan, observedProviderSessionID string) {
	if r.sessionStore == nil || plan.InternalSessionID == "" {
		return
	}
	providerSessionID := firstNonEmptyString(observedProviderSessionID, plan.ProviderSessionID)
	if providerSessionID == "" {
		return
	}
	if plan.Resume && providerSessionID == plan.ProviderSessionID {
		return
	}
	if err := r.sessionStore.Upsert(ctx, &ProviderSessionBinding{
		InternalSessionID: plan.InternalSessionID,
		Provider:          plan.Key.Provider,
		ProviderSessionID: providerSessionID,
		WorkDir:           plan.Key.WorkDir,
		AssistantID:       plan.Key.AssistantID,
		Status:            providerSessionStatusActive,
	}); err != nil {
		logs.WarnContextf(ctx, "Store provider session failed: provider=%s session=%s provider_session=%s error=%v", plan.Key.Provider, plan.InternalSessionID, providerSessionID, err)
	}
}

func (r *Runner) markProviderSessionFailed(ctx context.Context, plan providerSessionPlan, runErr error) {
	if r.sessionStore == nil || plan.InternalSessionID == "" || plan.ProviderSessionID == "" || runErr == nil {
		return
	}
	if err := r.sessionStore.MarkFailed(ctx, plan.Key, runErr.Error()); err != nil {
		logs.WarnContextf(ctx, "Mark provider session failed: provider=%s session=%s error=%v", plan.Key.Provider, plan.InternalSessionID, err)
	}
}

func internalSessionIDFromRequest(req *agent.RequestContext) string {
	if req == nil {
		return ""
	}
	if strings.TrimSpace(req.Conversation.ID) != "" {
		return strings.TrimSpace(req.Conversation.ID)
	}
	if value := metadataString(req.Metadata, "session_id"); value != "" {
		return value
	}
	return metadataString(req.Metadata, "sessionId")
}

func metadataString(metadata map[string]any, key string) string {
	if len(metadata) == 0 {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (r *Runner) failedResult(req *agent.RequestContext, startedAt time.Time, err error, metadata map[string]any) *agent.RunResult {
	status := agent.RunStatusFailed
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		status = agent.RunStatusCancelled
	}
	message := ""
	if err != nil {
		message = err.Error()
	}
	return &agent.RunResult{
		RunID:       req.RunID,
		TraceID:     req.TraceID,
		Status:      status,
		Error:       message,
		StartedAt:   startedAt,
		CompletedAt: time.Now().UTC(),
		Metadata:    metadataWithRuntime(metadata, r.name),
	}
}

func failureMetadata(workDir string) map[string]any {
	metadata := map[string]any{}
	if strings.TrimSpace(workDir) != "" {
		metadata["work_dir"] = workDir
	}
	return metadata
}

func metadataWithRuntime(metadata map[string]any, runtimeName string) map[string]any {
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["runtime"] = runtimeName
	return metadata
}

func ensureRunDefaults(req *agent.RequestContext) {
	if req.RunID == "" {
		req.RunID = fmt.Sprintf("run_%d", time.Now().UTC().UnixNano())
	}
	if req.Input.Type == "" {
		req.Input.Type = agent.InputTypeMessage
	}
}

func sinkForRequest(req *agent.RequestContext) events.Sink {
	if req == nil || req.EventSink == nil {
		return events.NewNoopSink()
	}
	return req.EventSink
}

func modelForRequest(req *agent.RequestContext) engines.ModelConfig {
	if req == nil {
		return engines.ModelConfig{}
	}
	model := engines.ModelConfig{
		Provider: req.Model.Provider,
		Model:    req.Model.Model,
		APIKey:   req.Model.APIKey,
		BaseURL:  req.Model.BaseURL,
	}
	return model
}

var _ agent.Runner = (*Runner)(nil)

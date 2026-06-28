// Package externalcli adapts external agent CLI providers to the agent.Runtime contract.
package externalcli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/insmtx/Leros/backend/agent"
	"github.com/insmtx/Leros/backend/agent/runtime/events"
	engines "github.com/insmtx/Leros/backend/agent/runtime/provider"
	runtimetodo "github.com/insmtx/Leros/backend/agent/runtime/todo"
	"github.com/ygpkg/yg-go/logs"
)

// Driver contains the shared process, provider-session, and event parsing machinery
// used by concrete CLI Runtime implementations.
type Driver struct {
	name            string
	engine          engines.Engine
	sessionStore    ProviderSessionStore
	approvalHandler engines.ApprovalHandler
	questionHandler engines.QuestionHandler
	mcpServers      []engines.MCPServerConfig
}

// NewDriver creates shared infrastructure for one concrete CLI Runtime.
func NewDriver(name string, engine engines.Engine) (*Driver, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("runtime name is required")
	}
	if engine == nil {
		return nil, fmt.Errorf("runtime %q engine is nil", name)
	}
	return &Driver{
		name:         name,
		engine:       engine,
		sessionStore: NewInMemoryProviderSessionStore(),
	}, nil
}

// SetSessionStore 替换用于外部 CLI 恢复的提供者会话存储。
func (r *Driver) SetSessionStore(store ProviderSessionStore) {
	if r == nil || store == nil {
		return
	}
	r.sessionStore = store
}

// SetApprovalHandler 设置审批处理器，用于 on-request 和 auto 模式。
func (r *Driver) SetApprovalHandler(handler engines.ApprovalHandler) {
	if r == nil {
		return
	}
	r.approvalHandler = handler
}

// SetQuestionHandler 设置问题处理器，用于引擎向用户提问时。
func (r *Driver) SetQuestionHandler(handler engines.QuestionHandler) {
	if r == nil {
		return
	}
	r.questionHandler = handler
}

// SetMCPServers 设置 MCP 服务配置，用于后续 Run() 时传入引擎。
func (r *Driver) SetMCPServers(cfgs []engines.MCPServerConfig) {
	if r == nil {
		return
	}
	r.mcpServers = cfgs
}

// Execute runs one request for the concrete Runtime that owns this Driver.
func (r *Driver) Execute(
	ctx context.Context,
	request agent.ExecutionRequest,
	observer agent.Observer,
) (agent.ExecutionResult, error) {
	if r == nil || r.engine == nil {
		return agent.ExecutionResult{}, fmt.Errorf("external CLI runtime is not initialized")
	}
	if strings.TrimSpace(request.ExecutionID) == "" {
		return agent.ExecutionResult{}, fmt.Errorf("execution id is required")
	}
	var eventSink events.Sink = events.NewNoopSink()
	if observer != nil {
		eventSink = observer
	}
	workDir := strings.TrimSpace(request.Filesystem.WorkDir)
	if err := r.engine.Prepare(ctx, engines.PrepareRequest{WorkDir: workDir}); err != nil {
		return agent.ExecutionResult{}, err
	}

	sessionPlan := r.resolveProviderSession(ctx, request)
	handle, err := r.engine.Run(ctx, engines.RunRequest{
		ExecutionID:  request.ExecutionID,
		SessionID:    sessionPlan.ProviderSessionID,
		Resume:       sessionPlan.Resume,
		WorkDir:      workDir,
		TaskDir:      request.Filesystem.TaskDir,
		SystemPrompt: strings.TrimSpace(request.SystemPrompt),
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
		ExtraEnv:        nil,
		PermissionMode:  engines.PermissionMode(request.Policy.PermissionMode),
		ApprovalHandler: r.approvalHandler,
		MCPServers:      r.mcpServers,
	})
	if err != nil {
		return agent.ExecutionResult{}, err
	}

	if handle != nil && handle.Process != nil {
		logs.InfoContextf(ctx, "External runtime %s started with pid %d", r.name, handle.Process.PID())
	}

	consumeResult, err := ConsumeEvents(
		ctx,
		eventSink,
		handle,
		request.ExecutionID,
		request.TraceID,
		r.approvalHandler,
		r.questionHandler,
	)
	if err != nil {
		r.markProviderSessionFailed(ctx, sessionPlan, err)
		return agent.ExecutionResult{}, err
	}
	r.persistProviderSession(ctx, sessionPlan, consumeResult.ProviderSessionID)

	return agent.ExecutionResult{
		Message:                strings.TrimSpace(consumeResult.Message),
		Usage:                  consumeResult.Usage,
		ProviderConversationID: firstNonEmptyString(sessionPlan.ProviderSessionID, consumeResult.ProviderSessionID),
	}, nil
}

// ConsumeResult is the provider-independent result extracted from an event stream.
type ConsumeResult struct {
	Message           string
	ProviderSessionID string
	Usage             *agent.Usage
}

// ConsumeEvents parses provider activity events and forwards normalized activity to sink.
func ConsumeEvents(
	ctx context.Context,
	sink events.Sink,
	handle *engines.RunHandle,
	runID string,
	traceID string,
	approvalHandler engines.ApprovalHandler,
	questionHandler engines.QuestionHandler,
) (ConsumeResult, error) {
	if handle == nil || handle.Events == nil {
		return ConsumeResult{}, nil
	}
	if sink == nil {
		sink = events.NewNoopSink()
	}
	var result strings.Builder
	resultSeen := false
	consumed := ConsumeResult{}
	messageIDs := events.NewMessageIDMapper()
	todoSink := events.SinkFunc(func(emitCtx context.Context, event *agent.Event) error {
		if event == nil {
			return nil
		}
		if event.RunID == "" {
			event.RunID = runID
		}
		if event.TraceID == "" {
			event.TraceID = traceID
		}
		return sink.Emit(emitCtx, event)
	})
	todoTracker := runtimetodo.NewTracker(runtimetodo.Options{RunID: runID, Sink: todoSink})
	for event := range handle.Events {
		// Fill RunID/TraceID that the old Journal would have provided.
		if event.RunID == "" {
			event.RunID = runID
		}
		if event.TraceID == "" {
			event.TraceID = traceID
		}
		switch event.Type {
		case events.EventStarted:
			continue
		case events.EventProviderSessionStarted:
			if strings.TrimSpace(event.Content) != "" {
				consumed.ProviderSessionID = strings.TrimSpace(event.Content)
			}
		case events.EventResult:
			if resultPayload, err := events.DecodePayload[events.MessageResultPayload](&event); err == nil {
				if strings.TrimSpace(resultPayload.Message) != "" {
					result.Reset()
					result.WriteString(resultPayload.Message)
					resultSeen = true
				}
				if resultPayload.Usage != nil {
					consumed.Usage = resultPayload.Usage
				}
			} else if strings.TrimSpace(event.Content) != "" {
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
				_ = sink.Emit(ctx, normalizeRuntimeEvent(event, messageIDs))
				if !resultSeen {
					result.WriteString(event.Content)
				}
			}
		case events.EventReasoningDelta:
			_ = sink.Emit(ctx, normalizeRuntimeEvent(event, messageIDs))
		case events.EventToolCallStarted, events.EventToolCallCompleted, events.EventToolCallFailed:
			_ = sink.Emit(ctx, normalizeRuntimeEvent(event, messageIDs))
		case events.EventTodoSnapshot:
			if items, err := events.DecodePayload[[]events.RuntimeTodoItem](&event); err == nil {
				_ = todoTracker.Snapshot(ctx, items)
			}
		case events.EventTodoUpdated:
			if items, err := events.DecodePayload[[]events.RuntimeTodoItem](&event); err == nil {
				_ = todoTracker.Update(ctx, items, true)
			}
		case events.EventApprovalRequested:
			_ = sink.Emit(ctx, normalizeRuntimeEvent(event, messageIDs))
			if handle.Responder == nil {
				logs.WarnContextf(ctx, "approval request dropped: no Responder (PermissionMode may need to be on-request/auto)")
			}
			if approvalHandler != nil && handle.Responder != nil {
				req, decErr := events.DecodePayload[events.ApprovalRequestPayload](&event)
				if decErr != nil {
					logs.WarnContextf(ctx, "decode approval request: %v", decErr)
					continue
				}
				decision, decErr := approvalHandler.RequestApproval(ctx, &agent.ApprovalRequest{
					RequestID:   req.RequestID,
					ToolCallID:  req.ToolCallID,
					ToolName:    req.ToolName,
					Arguments:   json.RawMessage(req.Arguments),
					Description: req.Description,
					Runtime:     metadataString(req.Metadata, "engine"),
				})
				if decErr != nil {
					logs.WarnContextf(ctx, "approval handler error: %v", decErr)
					continue
				}
				if wErr := handle.Responder.WriteDecision(req.RequestID, decision.Action); wErr != nil {
					logs.WarnContextf(ctx, "write approval decision to stdin: %v", wErr)
				}
				_ = sink.Emit(ctx, normalizeRuntimeEvent(*events.NewApprovalResolved(events.ApprovalDecisionPayload{
					RequestID: req.RequestID,
					Action:    decision.Action,
					Reason:    decision.Reason,
				}), messageIDs))
			}
		case events.EventApprovalResolved:
			_ = sink.Emit(ctx, normalizeRuntimeEvent(event, messageIDs))
		case events.EventQuestionAsked:
			_ = sink.Emit(ctx, normalizeRuntimeEvent(event, messageIDs))
			if handle.Questions == nil {
				logs.WarnContextf(ctx, "question request dropped: no QuestionResponder")
			}
			if questionHandler != nil && handle.Questions != nil {
				req, decErr := events.DecodePayload[events.QuestionRequestPayload](&event)
				if decErr != nil {
					logs.WarnContextf(ctx, "decode question request: %v", decErr)
					continue
				}
				// 构建 engine 层的 QuestionRequest
				qItems := make([]agent.QuestionItem, 0, len(req.Questions))
				for _, q := range req.Questions {
					opts := make([]agent.QuestionOption, 0, len(q.Options))
					for _, o := range q.Options {
						opts = append(opts, agent.QuestionOption{
							Label:       o.Label,
							Description: o.Description,
						})
					}
					qItems = append(qItems, agent.QuestionItem{
						Question:    q.Question,
						Header:      q.Header,
						Options:     opts,
						MultiSelect: q.MultiSelect,
						Custom:      q.Custom,
					})
				}
				answer, decErr := questionHandler.RequestAnswer(ctx, &agent.QuestionRequest{
					RequestID:   req.RequestID,
					SessionKey:  req.SessionID,
					Questions:   qItems,
					ToolCallID:  req.ToolCallID,
					Description: firstQuestionText(req.Questions),
					Runtime:     metadataString(req.Metadata, "engine"),
				})
				if decErr != nil {
					logs.WarnContextf(ctx, "question handler error: %v", decErr)
					continue
				}
				if wErr := handle.Questions.WriteAnswer(req.RequestID, answer.Answers); wErr != nil {
					logs.WarnContextf(ctx, "write question answer: %v", wErr)
				}
				_ = sink.Emit(ctx, normalizeRuntimeEvent(*events.NewQuestionAnswered(events.QuestionAnswerPayload{
					RequestID: req.RequestID,
					Answers:   answer.Answers,
				}), messageIDs))
			}
		case events.EventQuestionAnswered:
			_ = sink.Emit(ctx, normalizeRuntimeEvent(event, messageIDs))
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

func normalizeRuntimeEvent(event agent.Event, messageIDs *events.MessageIDMapper) *agent.Event {
	switch event.Type {
	case events.EventMessageDelta:
		if len(event.Payload) > 0 {
			payload, err := events.DecodePayload[events.MessageDeltaPayload](&event)
			if err == nil && strings.TrimSpace(payload.MessageID) != "" {
				return events.NewMessageDelta(messageIDs.ForProvider(payload.MessageID), payload.Content)
			}
			if err == nil {
				return events.NewMessageDelta(messageIDs.CurrentOrNew(), payload.Content)
			}
			return &event
		}
		return events.NewMessageDelta(messageIDs.CurrentOrNew(), event.Content)
	case events.EventReasoningDelta:
		if len(event.Payload) > 0 {
			payload, err := events.DecodePayload[events.MessageDeltaPayload](&event)
			if err == nil && strings.TrimSpace(payload.MessageID) != "" {
				return events.NewReasoningDelta(messageIDs.ForProvider(payload.MessageID), payload.Content)
			}
			if err == nil {
				return events.NewReasoningDelta(messageIDs.CurrentOrNew(), payload.Content)
			}
			return &event
		}
		return events.NewReasoningDelta(messageIDs.CurrentOrNew(), event.Content)
	case events.EventToolCallStarted:
		payload, err := events.DecodePayload[events.ToolCallPayload](&event)
		if err != nil {
			return &event
		}
		return events.NewToolCallStarted(firstNonEmptyString(payload.ToolCallID, legacyToolCallID(event)), payload.Name, payload.Arguments)
	case events.EventToolCallCompleted:
		payload, err := events.DecodePayload[events.ToolCallResultPayload](&event)
		if err != nil {
			return &event
		}
		return events.NewToolCallCompleted(payload.ToolCallID, payload.Name, payload.Result, payload.ElapsedMS)
	case events.EventToolCallFailed:
		payload, err := events.DecodePayload[events.ToolCallResultPayload](&event)
		if err != nil {
			return &event
		}
		return events.NewToolCallFailed(payload.ToolCallID, payload.Name, payload.Error, payload.ElapsedMS)
	default:
		return &event
	}
}

func legacyToolCallID(event agent.Event) string {
	var payload struct {
		CallID     string `json:"call_id"`
		ToolCallID string `json:"tool_call_id"`
	}
	if len(event.Payload) > 0 && json.Unmarshal(event.Payload, &payload) == nil {
		return firstNonEmptyString(payload.ToolCallID, payload.CallID)
	}
	if strings.TrimSpace(event.Content) != "" && json.Unmarshal([]byte(event.Content), &payload) == nil {
		return firstNonEmptyString(payload.ToolCallID, payload.CallID)
	}
	return ""
}

type providerSessionPlan struct {
	InternalSessionID string
	ProviderSessionID string
	Resume            bool
	Key               ProviderSessionKey
}

func (r *Driver) resolveProviderSession(ctx context.Context, request agent.ExecutionRequest) providerSessionPlan {
	internalSessionID := strings.TrimSpace(request.SessionKey)
	plan := providerSessionPlan{
		InternalSessionID: internalSessionID,
		Key: ProviderSessionKey{
			InternalSessionID: internalSessionID,
			Provider:          r.name,
			WorkDir:           request.Filesystem.WorkDir,
			AssistantID:       request.InstanceKey,
		},
	}
	// Native 引擎不使用外部 CLI 会话，直接使用 Leros 内部 session ID。
	if r.name == engines.EngineNative {
		plan.ProviderSessionID = internalSessionID
		return plan
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

func (r *Driver) persistProviderSession(ctx context.Context, plan providerSessionPlan, observedProviderSessionID string) {
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

func (r *Driver) markProviderSessionFailed(ctx context.Context, plan providerSessionPlan, runErr error) {
	if r.sessionStore == nil || plan.InternalSessionID == "" || plan.ProviderSessionID == "" || runErr == nil {
		return
	}
	if err := r.sessionStore.MarkFailed(ctx, plan.Key, runErr.Error()); err != nil {
		logs.WarnContextf(ctx, "Mark provider session failed: provider=%s session=%s error=%v", plan.Key.Provider, plan.InternalSessionID, err)
	}
}

func metadataString(meta map[string]string, key string) string {
	if meta == nil {
		return ""
	}
	return meta[key]
}

func firstQuestionText(questions []events.QuestionItem) string {
	if len(questions) == 0 {
		return ""
	}
	if questions[0].Header != "" {
		return questions[0].Header
	}
	return questions[0].Question
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

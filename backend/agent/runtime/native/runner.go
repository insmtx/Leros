// Package native implements the built-in Eino-backed Leros runtime.
package native

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/insmtx/Leros/backend/agent"
	"github.com/insmtx/Leros/backend/agent/runtime/events"
	engines "github.com/insmtx/Leros/backend/agent/runtime/provider"
	runtimetodo "github.com/insmtx/Leros/backend/agent/runtime/todo"
	pkgeino "github.com/insmtx/Leros/backend/pkg/eino"
	"github.com/insmtx/Leros/backend/prompts"
	"github.com/ygpkg/yg-go/logs"
)

// Runner 是 Leros 内置 Eino 运行时入口。
type Runner struct{}

// NewRunner 创建基于 Eino Flow 的 Leros 内置 Agent。
func NewRunner(context.Context) (*Runner, error) {
	return &Runner{}, nil
}

// Run 执行标准化请求，通过统一的 agent.Event channel 支持流式输出。
func (r *Runner) Run(ctx context.Context, req engines.RunRequest) (<-chan agent.Event, error) {
	if r == nil {
		return nil, fmt.Errorf("leros runner is not initialized")
	}

	eventsCh := make(chan agent.Event, 256)

	go func() {
		defer close(eventsCh)
		r.executeStreaming(ctx, req, eventsCh)
	}()

	return eventsCh, nil
}

// executeStreaming 在 goroutine 中执行推理，将全部事件写入 channel。
func (r *Runner) executeStreaming(ctx context.Context, req engines.RunRequest, eventsCh chan<- agent.Event) {
	channelSink := engineSinkFunc(func(ctx context.Context, event *agent.Event) error {
		if event != nil {
			select {
			case eventsCh <- *event:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	sendEngineEvent(eventsCh, engines.EngineEventStarted, req.ExecutionID)

	message, usage, err := r.runWithState(ctx, req, channelSink)
	if err != nil {
		sendEngineEvent(eventsCh, engines.EngineEventFailed, fmt.Sprintf("run_id=%s error=%s", req.ExecutionID, err.Error()))
		return
	}

	payload, _ := json.Marshal(engineMessageResultPayload{
		Message: message,
		Usage:   usage,
	})
	eventsCh <- agent.Event{
		Type:    engines.EngineEventResult,
		Payload: payload,
		Content: message,
	}

	sendEngineEvent(eventsCh, engines.EngineEventCompleted, req.ExecutionID)
}

// engineMessageResultPayload mirrors the message result payload for engine events.
type engineMessageResultPayload struct {
	Message string       `json:"message,omitempty"`
	Usage   *agent.Usage `json:"usage,omitempty"`
}

// engineSinkFunc is a function adapter for engine event sinks within native runner internals.
type engineSinkFunc func(ctx context.Context, event *agent.Event) error

func (f engineSinkFunc) Emit(ctx context.Context, event *agent.Event) error {
	return f(ctx, event)
}

// engineSink is the sink interface used internally by native runner.
type engineSink interface {
	Emit(ctx context.Context, event *agent.Event) error
}

// streamSink adapts engineSink to the Eino streaming sink protocol.
type streamSink struct {
	sink engineSink
}

func (s streamSink) EmitMessageDelta(ctx context.Context, messageID string, content string) error {
	if s.sink == nil {
		return nil
	}
	return s.sink.Emit(ctx, newEngineMessageDelta(messageID, content))
}

func (s streamSink) EmitReasoningDelta(ctx context.Context, messageID string, content string) error {
	if s.sink == nil {
		return nil
	}
	return s.sink.Emit(ctx, newEngineReasoningDelta(messageID, content))
}

func sendEngineEvent(eventsCh chan<- agent.Event, eventType agent.EventType, content string) {
	select {
	case eventsCh <- agent.Event{Type: eventType, Content: content}:
	default:
	}
}

// Engine-level event constructors (replace events.New* calls).
func newEngineMessageDelta(messageID, content string) *agent.Event {
	payload, _ := json.Marshal(engineMessageDeltaPayload{
		MessageID: messageID,
		Role:      "assistant",
		Content:   content,
	})
	return &agent.Event{
		Type:    engines.EngineEventMessageDelta,
		Payload: payload,
		Content: content,
	}
}

func newEngineReasoningDelta(messageID, content string) *agent.Event {
	payload, _ := json.Marshal(engineMessageDeltaPayload{
		MessageID: messageID,
		Role:      "assistant",
		Content:   content,
	})
	return &agent.Event{
		Type:    engines.EngineEventReasoningDelta,
		Payload: payload,
		Content: content,
	}
}

type engineMessageDeltaPayload struct {
	MessageID string `json:"message_id,omitempty"`
	Role      string `json:"role,omitempty"`
	Content   string `json:"content"`
}

func (r *Runner) runWithState(ctx context.Context, req engines.RunRequest, sink engineSink) (string, *agent.Usage, error) {
	chatModel, err := pkgeino.NewChatModel(ctx, &pkgeino.ChatModelConfig{
		Provider: req.Model.Provider,
		APIKey:   req.Model.APIKey,
		Model:    req.Model.Model,
		BaseURL:  req.Model.BaseURL,
	})
	if err != nil {
		return "", nil, err
	}

	systemPrompt := r.buildSystemPrompt(req)

	binding := r.buildToolBinding(req, sink)
	toolSpecs, toolInvoker, err := buildRuntimeTools(binding, sink)
	if err != nil {
		return "", nil, fmt.Errorf("build eino tools: %w", err)
	}
	einoBaseTools := buildEinoTools(toolSpecs, toolInvoker)

	historyMessages := buildHistoryMessages(req.Messages, 20)

	flow, err := pkgeino.NewFlow(ctx, &pkgeino.FlowConfig{
		Model:        chatModel,
		Tools:        einoBaseTools,
		SystemPrompt: systemPrompt,
		MaxStep:      90,
		Messages:     historyMessages,
	})
	if err != nil {
		return "", nil, err
	}

	var message interface {
		String() string
	}
	var resultMessage string
	var usage *agent.Usage
	if sink != nil {
		streamedMessage, streamedUsage, streamErr := flow.StreamWithUsage(ctx, req.Prompt, streamSink{sink: sink})
		err = streamErr
		if streamedMessage != nil {
			message = streamedMessage
			resultMessage = strings.TrimSpace(streamedMessage.Content)
			usage = engineUsagePayload(streamedUsage)
		}
	} else {
		generatedMessage, generatedUsage, generateErr := flow.GenerateWithUsage(ctx, req.Prompt)
		err = generateErr
		if generatedMessage != nil {
			message = generatedMessage
			resultMessage = strings.TrimSpace(generatedMessage.Content)
			usage = engineUsagePayload(generatedUsage)
		}
	}
	if err != nil {
		return "", nil, err
	}
	if resultMessage == "" && message != nil {
		resultMessage = formatLLMResultForLog(message)
	}

	logs.InfoContextf(ctx, "Leros runtime final LLM result: run_id=%s result=%s",
		req.ExecutionID, formatLLMResultForLog(message))

	return resultMessage, usage, nil
}

// buildHistoryMessages converts prepared execution messages into Eino ADK history.
func buildHistoryMessages(messages []agent.Message, maxMessages int) []adk.Message {
	if len(messages) == 0 {
		return nil
	}

	einoMessages := make([]pkgeino.Message, 0, len(messages))
	for _, msg := range messages {
		if strings.TrimSpace(msg.Content) == "" {
			continue
		}
		einoMessages = append(einoMessages, pkgeino.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if maxMessages > 0 && len(einoMessages) > maxMessages {
		einoMessages = einoMessages[len(einoMessages)-maxMessages:]
	}

	return pkgeino.BuildMessages(einoMessages)
}

func (r *Runner) buildToolBinding(req engines.RunRequest, sink engineSink) toolBinding {
	// Create an events.Sink adapter from engineSink for the todo tracker.
	eventsSink := engineSinkToEventsSink(sink)
	return toolBinding{
		Tools:        append([]agent.Tool(nil), req.Tools...),
		AllowedTools: append([]string(nil), req.AllowedTools...),
		EngineSink:   sink,
		TodoReporter: runtimetodo.NewTracker(runtimetodo.Options{
			RunID: req.ExecutionID,
			Sink:  eventsSink,
		}),
	}
}

// engineSinkToEventsSink wraps an engineSink as an events.Sink for the todo tracker.
func engineSinkToEventsSink(sink engineSink) events.Sink {
	if sink == nil {
		return events.NewNoopSink()
	}
	return events.SinkFunc(func(ctx context.Context, evt *agent.Event) error {
		return sink.Emit(ctx, &agent.Event{
			Type:    evt.Type,
			Content: evt.Content,
			Payload: json.RawMessage(evt.Payload),
		})
	})
}

func (r *Runner) buildSystemPrompt(req engines.RunRequest) string {
	prompt := req.SystemPrompt
	if hint := strings.TrimSpace(prompts.Get(prompts.KeyAgentNativeSkillUsageHint)); hint != "" {
		prompt += "\n\n" + hint
	}
	return prompt
}

func formatLLMResultForLog(message interface{ String() string }) string {
	if message == nil {
		return "<nil>"
	}

	formatted := strings.TrimSpace(message.String())
	if formatted == "" {
		return "<empty>"
	}
	if len(formatted) > 2000 {
		return formatted[:2000] + "...(truncated)"
	}
	return formatted
}

func engineUsagePayload(usage *pkgeino.Usage) *agent.Usage {
	if usage == nil {
		return nil
	}
	return &agent.Usage{
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		TotalTokens:  usage.TotalTokens,
	}
}

func buildEinoTools(specs []pkgeino.ToolSpec, invoker pkgeino.ToolInvoker) []einotool.BaseTool {
	if len(specs) == 0 {
		return nil
	}
	result := make([]einotool.BaseTool, 0, len(specs))
	for _, spec := range specs {
		result = append(result, pkgeino.NewTool(spec, invoker))
	}
	return result
}

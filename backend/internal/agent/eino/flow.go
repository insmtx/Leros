package eino

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/adk"
	einomodel "github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	einoschema "github.com/cloudwego/eino/schema"
	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
	"github.com/ygpkg/yg-go/logs"
)

type Flow struct {
	agent        adk.Agent
	runner       *adk.Runner
	streamRunner *adk.Runner
}

type FlowConfig struct {
	Model        einomodel.ToolCallingChatModel
	Tools        []einotool.BaseTool
	Emitter      *events.Emitter
	SystemPrompt string
	MaxStep      int
}

func NewFlow(ctx context.Context, cfg *FlowConfig) (*Flow, error) {
	if cfg == nil {
		return nil, fmt.Errorf("flow config is required")
	}
	if cfg.Model == nil {
		return nil, fmt.Errorf("tool-calling model is required")
	}

	maxStep := cfg.MaxStep
	if maxStep <= 0 {
		maxStep = 20
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "LerosAgent",
		Description: "Leros runtime agent",
		Model:       cfg.Model,
		Instruction: cfg.SystemPrompt,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: cfg.Tools,
				UnknownToolsHandler: func(ctx context.Context, toolName, toolInput string) (string, error) {
					logs.WarnContextf(ctx, "[WARN] unknown tool call: %s with input: %s", toolName, toolInput)
					return fmt.Sprintf(`Tool "%s" does not exist. Please use a valid tool and retry.`, toolName), nil
				},
			},
		},
		MaxIterations: maxStep,
		ModelRetryConfig: &adk.ModelRetryConfig{
			MaxRetries: 5,
			IsRetryAble: func(_ context.Context, err error) bool {
				return strings.Contains(err.Error(), "429") ||
					strings.Contains(err.Error(), "Too Many Requests") ||
					strings.Contains(err.Error(), "qpm limit")
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create eino agent: %w", err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})
	streamRunner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent, EnableStreaming: true})

	return &Flow{agent: agent, runner: runner, streamRunner: streamRunner}, nil
}

func (f *Flow) Generate(ctx context.Context, userInput string) (*einoschema.Message, error) {
	if f == nil || f.runner == nil {
		return nil, fmt.Errorf("flow is not initialized")
	}
	if strings.TrimSpace(userInput) == "" {
		return nil, fmt.Errorf("user input is required")
	}

	iter := f.runner.Query(ctx, userInput)
	var result *einoschema.Message

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return nil, event.Err
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				return nil, err
			}
			if msg != nil {
				result = msg
			}
		}
	}

	if result == nil {
		return nil, fmt.Errorf("agent returned no message")
	}
	return result, nil
}

func (f *Flow) Stream(ctx context.Context, userInput string, emitter *events.Emitter) (*einoschema.Message, error) {
	if f == nil || f.streamRunner == nil {
		return nil, fmt.Errorf("flow is not initialized")
	}
	if strings.TrimSpace(userInput) == "" {
		return nil, fmt.Errorf("user input is required")
	}

	iter := f.streamRunner.Query(ctx, userInput)

	var lastMsg *einoschema.Message
	var currentMessageID string
	messageIDs := events.NewMessageIDMapper()

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return nil, event.Err
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			mv := event.Output.MessageOutput

			if mv.Role == einoschema.Tool {
				continue
			}

			currentMessageID := messageIDs.StartNew()

			if mv.IsStreaming && mv.MessageStream != nil {
				streams := mv.MessageStream.Copy(2)
				emitStream := streams[0]
				concatStream := streams[1]
				emitStream.SetAutomaticClose()
				for {
					chunk, err := emitStream.Recv()
					if errors.Is(err, io.EOF) {
						break
					}
					if err != nil {
						return nil, fmt.Errorf("read stream chunk: %w", err)
					}
					if chunk.Content != "" {
						_ = emitter.Emit(ctx, events.NewMessageDelta(currentMessageID, chunk.Content))
					}
					if chunk.ReasoningContent != "" {
						_ = emitter.Emit(ctx, events.NewReasoningDelta(currentMessageID, chunk.ReasoningContent))
					}
				}
				lastMsg, _ = einoschema.ConcatMessageStream(concatStream)
			} else {
				msg, err := mv.GetMessage()
				if err != nil {
					return nil, err
				}
				if msg != nil {
					lastMsg = msg
					_ = emitter.Emit(ctx, events.NewMessageDelta(currentMessageID, msg.Content))
				}
			}

			logs.InfoContextf(ctx, "agent msg:msgID=%s, content=%s, reasoning=%s",
				currentMessageID, lastMsg.Content, lastMsg.ReasoningContent)
		}
	}

	if lastMsg == nil {
		return nil, fmt.Errorf("agent stream returned no messages")
	}

	lastMsg.Extra = map[string]any{"message_id": currentMessageID}
	return lastMsg, nil
}

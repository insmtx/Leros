// Package taskconsumer consumes worker task messages and executes agent runs.
package taskconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/insmtx/SingerOS/backend/internal/agent"
	agentevents "github.com/insmtx/SingerOS/backend/internal/agent/events"
	eventbus "github.com/insmtx/SingerOS/backend/internal/infra/mq"
	"github.com/insmtx/SingerOS/backend/pkg/dm"
	"github.com/ygpkg/yg-go/logs"
)

// Config controls one standalone worker task consumer.
type Config struct {
	OrgID    string
	WorkerID string
}

// Consumer subscribes to one worker task topic and dispatches tasks to an agent runtime.
type Consumer struct {
	cfg        Config
	subscriber eventbus.Subscriber
	publisher  eventbus.RealtimePublisher
	runner     agent.Runner
}

// New creates a worker task consumer.
func New(cfg Config, subscriber eventbus.Subscriber, publisher eventbus.RealtimePublisher, runner agent.Runner) (*Consumer, error) {
	cfg.OrgID = strings.TrimSpace(cfg.OrgID)
	cfg.WorkerID = strings.TrimSpace(cfg.WorkerID)
	if cfg.OrgID == "" {
		return nil, fmt.Errorf("worker org_id is required")
	}
	if cfg.WorkerID == "" {
		return nil, fmt.Errorf("worker worker_id is required")
	}
	if subscriber == nil {
		return nil, fmt.Errorf("subscriber is required")
	}
	if runner == nil {
		return nil, fmt.Errorf("agent runner is required")
	}
	return &Consumer{
		cfg:        cfg,
		subscriber: subscriber,
		publisher:  publisher,
		runner:     runner,
	}, nil
}

// TaskTopic returns the NATS subject consumed by this worker.
func (c *Consumer) TaskTopic() string {
	return dm.Topic().Org(c.cfg.OrgID).Worker(c.cfg.WorkerID).Task().Build()
}

// Start subscribes to the worker task topic.
func (c *Consumer) Start(ctx context.Context) error {
	topic := c.TaskTopic()
	logs.InfoContextf(ctx, "Starting worker task subscription: %s", topic)
	return c.subscriber.Subscribe(ctx, topic, func(event any) {
		if err := c.handleEvent(ctx, event); err != nil {
			logs.ErrorContextf(ctx, "Failed to handle worker task: %v", err)
		}
	})
}

func (c *Consumer) handleEvent(ctx context.Context, event any) error {
	msg, err := decodeWorkerTask(event)
	if err != nil {
		return err
	}
	if err := c.validateRoute(msg); err != nil {
		return err
	}
	if msg.Body.TaskType != dm.TaskTypeAgentRun {
		return fmt.Errorf("unsupported worker task type %q", msg.Body.TaskType)
	}

	req := RequestFromWorkerTask(msg)
	req.EventSink = NewMQStreamSink(c.publisher, msg)

	result, err := c.runner.Run(ctx, req)
	if err != nil {
		return err
	}
	if result != nil {
		logs.InfoContextf(ctx, "Worker task completed: task_id=%s run_id=%s status=%s", req.TaskID, result.RunID, result.Status)
	}
	return nil
}

func decodeWorkerTask(event any) (dm.WorkerTaskMessage, error) {
	var msg dm.WorkerTaskMessage
	body, err := json.Marshal(event)
	if err != nil {
		return msg, fmt.Errorf("marshal worker task: %w", err)
	}
	if err := json.Unmarshal(body, &msg); err != nil {
		return msg, fmt.Errorf("unmarshal worker task: %w", err)
	}
	if msg.Type != "" && msg.Type != dm.MessageTypeWorkerTask {
		return msg, fmt.Errorf("unexpected worker task message type %q", msg.Type)
	}
	return msg, nil
}

func (c *Consumer) validateRoute(msg dm.WorkerTaskMessage) error {
	if strings.TrimSpace(msg.Route.OrgID) != "" && msg.Route.OrgID != c.cfg.OrgID {
		return fmt.Errorf("task org_id %q does not match worker org_id %q", msg.Route.OrgID, c.cfg.OrgID)
	}
	if strings.TrimSpace(msg.Route.WorkerID) != "" && msg.Route.WorkerID != c.cfg.WorkerID {
		return fmt.Errorf("task worker_id %q does not match worker_id %q", msg.Route.WorkerID, c.cfg.WorkerID)
	}
	return nil
}

// RequestFromWorkerTask converts the domain message protocol into the agent runtime boundary.
func RequestFromWorkerTask(msg dm.WorkerTaskMessage) *agent.RequestContext {
	return &agent.RequestContext{
		RunID:   firstNonEmpty(msg.Trace.RunID, msg.Trace.TaskID, msg.ID),
		TraceID: msg.Trace.TraceID,
		TaskID:  msg.Trace.TaskID,
		Assistant: agent.AssistantContext{
			ID:     msg.Body.Execution.AssistantID,
			Skills: append([]string(nil), msg.Body.Execution.Skills...),
			Tools:  append([]string(nil), msg.Body.Execution.Tools...),
		},
		Actor: agent.ActorContext{
			UserID:      msg.Body.Actor.UserID,
			DisplayName: msg.Body.Actor.DisplayName,
			Channel:     msg.Body.Actor.Channel,
			ExternalID:  msg.Body.Actor.ExternalID,
			AccountID:   msg.Body.Actor.AccountID,
		},
		Conversation: agent.ConversationContext{
			ID: msg.Route.SessionID,
		},
		Input: agent.InputContext{
			Type:        agent.InputType(msg.Body.Input.Type),
			Text:        msg.Body.Input.Text,
			Messages:    inputMessagesFromTask(msg.Body.Input.Messages),
			Attachments: attachmentsFromTask(msg.Body.Input.Attachments),
		},
		Runtime: agent.RuntimeOptions{
			Kind:    msg.Body.Runtime.Kind,
			WorkDir: msg.Body.Runtime.WorkDir,
			MaxStep: msg.Body.Runtime.MaxStep,
		},
		Capability: agent.CapabilityContext{
			AllowedTools: append([]string(nil), msg.Body.Execution.Tools...),
		},
		Policy: agent.PolicyContext{
			RequireApproval: msg.Body.Policy.RequireApproval,
		},
		Metadata: map[string]any{
			"message_id": msg.ID,
			"org_id":     msg.Route.OrgID,
			"worker_id":  msg.Route.WorkerID,
			"session_id": msg.Route.SessionID,
			"agent_id":   msg.Body.Execution.AgentID,
			"metadata":   msg.Metadata,
		},
	}
}

func inputMessagesFromTask(messages []dm.ChatMessage) []agent.InputMessage {
	if len(messages) == 0 {
		return nil
	}
	result := make([]agent.InputMessage, 0, len(messages))
	for _, message := range messages {
		result = append(result, agent.InputMessage{
			Role:    string(message.Role),
			Content: message.Content,
		})
	}
	return result
}

func attachmentsFromTask(attachments []dm.Attachment) []agent.Attachment {
	if len(attachments) == 0 {
		return nil
	}
	result := make([]agent.Attachment, 0, len(attachments))
	for _, attachment := range attachments {
		result = append(result, agent.Attachment{
			ID:       attachment.ID,
			Name:     attachment.Name,
			MimeType: attachment.MimeType,
			URL:      attachment.URL,
		})
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

var _ agentevents.EventSink = (*MQStreamSink)(nil)

package taskconsumer

import (
	"context"
	"fmt"
	"time"

	agentevents "github.com/insmtx/SingerOS/backend/internal/agent/events"
	eventbus "github.com/insmtx/SingerOS/backend/internal/infra/mq"
	"github.com/insmtx/SingerOS/backend/pkg/dm"
	"github.com/ygpkg/yg-go/logs"
)

// MQStreamSink publishes agent runtime events as realtime messages.
type MQStreamSink struct {
	publisher eventbus.RealtimePublisher
	task      dm.WorkerTaskMessage
}

// NewMQStreamSink creates a stream sink for one worker task.
func NewMQStreamSink(publisher eventbus.RealtimePublisher, task dm.WorkerTaskMessage) *MQStreamSink {
	return &MQStreamSink{
		publisher: publisher,
		task:      task,
	}
}

// Emit publishes one runtime event to the task's session stream subject.
func (s *MQStreamSink) Emit(ctx context.Context, event *agentevents.RunEvent) error {
	if s == nil || s.publisher == nil || event == nil {
		return nil
	}

	topic := s.streamTopic()
	msg := dm.MessageStreamMessage{
		ID:        firstNonEmpty(event.ID, fmt.Sprintf("%s:%d", event.RunID, event.Seq)),
		Type:      dm.MessageTypeStream,
		CreatedAt: time.Now().UTC(),
		Trace: dm.TraceContext{
			TraceID:   firstNonEmpty(event.TraceID, s.task.Trace.TraceID),
			RequestID: s.task.Trace.RequestID,
			TaskID:    s.task.Trace.TaskID,
			RunID:     firstNonEmpty(event.RunID, s.task.Trace.RunID),
			ParentID:  s.task.Trace.ParentID,
		},
		Route: s.task.Route,
		Body: dm.StreamBody{
			Seq:   event.Seq,
			Event: streamEventType(event.Type),
			Payload: dm.StreamPayload{
				Role:    dm.MessageRoleAssistant,
				Content: event.Content,
			},
		},
	}
	if msg.Body.Event == dm.StreamEventRunFailed {
		msg.Body.Error = &dm.StreamError{Message: event.Content}
	}

	if err := s.publisher.PublishRealtime(ctx, topic, msg); err != nil {
		logs.WarnContextf(ctx, "Failed to publish worker stream event to %s: %v", topic, err)
	}
	return nil
}

func (s *MQStreamSink) streamTopic() string {
	if s.task.Route.SessionID != "" {
		return dm.Topic().Org(s.task.Route.OrgID).Session(s.task.Route.SessionID).Message().Stream().Build()
	}
	return dm.Topic().Org(s.task.Route.OrgID).Worker(s.task.Route.WorkerID).Stream().Build()
}

func streamEventType(eventType agentevents.RunEventType) dm.StreamEventType {
	switch eventType {
	case agentevents.RunEventStarted:
		return dm.StreamEventRunStarted
	case agentevents.RunEventCompleted:
		return dm.StreamEventRunCompleted
	case agentevents.RunEventFailed, agentevents.RunEventCancelled:
		return dm.StreamEventRunFailed
	case agentevents.RunEventMessageDelta, agentevents.RunEventReasoningDelta:
		return dm.StreamEventMessageDelta
	case agentevents.RunEventResult:
		return dm.StreamEventMessageCompleted
	case agentevents.RunEventToolCallStarted:
		return dm.StreamEventToolCallStarted
	case agentevents.RunEventToolCallArguments:
		return dm.StreamEventToolCallDelta
	case agentevents.RunEventToolCallOutput, agentevents.RunEventToolCallCompleted:
		return dm.StreamEventToolCallFinished
	case agentevents.RunEventToolCallFailed:
		return dm.StreamEventToolCallFinished
	default:
		return dm.StreamEventMessageDelta
	}
}

package taskconsumer

import (
	"context"
	"fmt"
	"time"

	agentevents "github.com/insmtx/Leros/backend/internal/agent/events"
	"github.com/insmtx/Leros/backend/internal/agent/eventtypes"
	eventbus "github.com/insmtx/Leros/backend/internal/infra/mq"
	"github.com/insmtx/Leros/backend/pkg/dm"
	"github.com/ygpkg/yg-go/logs"
)

// MQStreamSink publishes agent runtime events as realtime messages.
type MQStreamSink struct {
	publisher eventbus.RealtimePublisher
	task      eventtypes.WorkerTaskMessage
}

// NewMQStreamSink creates a stream sink for one worker task.
func NewMQStreamSink(publisher eventbus.RealtimePublisher, task eventtypes.WorkerTaskMessage) *MQStreamSink {
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
	msg := eventtypes.MessageStreamMessage{
		ID:        firstNonEmpty(event.ID, fmt.Sprintf("%s:%d", event.RunID, event.Seq)),
		Type:      eventtypes.MessageTypeStream,
		CreatedAt: time.Now().UTC(),
		Trace: eventtypes.TraceContext{
			TraceID:   firstNonEmpty(event.TraceID, s.task.Trace.TraceID),
			RequestID: s.task.Trace.RequestID,
			TaskID:    s.task.Trace.TaskID,
			RunID:     firstNonEmpty(event.RunID, s.task.Trace.RunID),
			ParentID:  s.task.Trace.ParentID,
		},
		Route: s.task.Route,
		Body: eventtypes.StreamBody{
			Seq:   event.Seq,
			Event: streamEventType(event.Type),
			Payload: eventtypes.StreamPayload{
				Role:    eventtypes.MessageRoleAssistant,
				Content: event.Content,
			},
		},
	}
	if msg.Body.Event == eventtypes.StreamEventRunFailed {
		msg.Body.Error = &eventtypes.StreamError{Message: event.Content}
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

func streamEventType(eventType agentevents.RunEventType) eventtypes.StreamEventType {
	switch eventType {
	case agentevents.RunEventStarted:
		return eventtypes.StreamEventRunStarted
	case agentevents.RunEventCompleted:
		return eventtypes.StreamEventRunCompleted
	case agentevents.RunEventFailed, agentevents.RunEventCancelled:
		return eventtypes.StreamEventRunFailed
	case agentevents.RunEventMessageDelta, agentevents.RunEventReasoningDelta:
		return eventtypes.StreamEventMessageDelta
	case agentevents.RunEventResult:
		return eventtypes.StreamEventMessageCompleted
	case agentevents.RunEventToolCallStarted:
		return eventtypes.StreamEventToolCallStarted
	case agentevents.RunEventToolCallArguments:
		return eventtypes.StreamEventToolCallDelta
	case agentevents.RunEventToolCallOutput, agentevents.RunEventToolCallCompleted:
		return eventtypes.StreamEventToolCallFinished
	case agentevents.RunEventToolCallFailed:
		return eventtypes.StreamEventToolCallFinished
	default:
		return eventtypes.StreamEventMessageDelta
	}
}

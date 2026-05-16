package taskconsumer

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
	eventbus "github.com/insmtx/Leros/backend/internal/infra/mq"
	"github.com/insmtx/Leros/backend/pkg/dm"
	"github.com/ygpkg/yg-go/logs"
)

// ResultPublisher publishes worker run stream events and persisted result events.
type ResultPublisher interface {
	eventbus.Publisher
	eventbus.RealtimePublisher
}

// MQStreamSink publishes agent runtime events as realtime messages.
type MQStreamSink struct {
	publisher ResultPublisher
	task      events.WorkerTaskMessage
}

// NewMQStreamSink creates a stream sink for one worker task.
func NewMQStreamSink(publisher ResultPublisher, task events.WorkerTaskMessage) *MQStreamSink {
	return &MQStreamSink{
		publisher: publisher,
		task:      task,
	}
}

// Emit publishes one runtime event to the task's session stream subject.
func (s *MQStreamSink) Emit(ctx context.Context, event *events.Event) error {
	if s == nil || s.publisher == nil || event == nil {
		return nil
	}

	topic := s.streamTopic()
	msg := events.MessageStreamMessage{
		ID:        firstNonEmpty(event.ID, fmt.Sprintf("%s:%d", event.RunID, event.Seq)),
		Type:      events.MessageTypeStream,
		CreatedAt: time.Now().UTC(),
		Trace: events.TraceContext{
			TraceID:   firstNonEmpty(event.TraceID, s.task.Trace.TraceID),
			RequestID: s.task.Trace.RequestID,
			TaskID:    s.task.Trace.TaskID,
			RunID:     firstNonEmpty(event.RunID, s.task.Trace.RunID),
			ParentID:  s.task.Trace.ParentID,
		},
		Route: s.task.Route,
		Body: events.StreamBody{
			Seq:   event.Seq,
			Event: streamEventType(event.Type),
			Payload: events.StreamPayload{
				Role:    events.MessageRoleAssistant,
				Content: event.Content,
			},
		},
	}
	if msg.Body.Event == events.StreamEventRunFailed {
		msg.Body.Error = &events.StreamError{Message: event.Content}
	}

	if err := s.publisher.PublishRealtime(ctx, topic, msg); err != nil {
		logs.WarnContextf(ctx, "Failed to publish worker stream event to %s: %v", topic, err)
	}
	if msg.Body.Event == events.StreamEventRunCompleted || msg.Body.Event == events.StreamEventRunFailed {
		s.emitCompleted(ctx, msg)
	}
	return nil
}

func (s *MQStreamSink) streamTopic() string {
	if s.task.Route.SessionID != "" {
		t, _ := dm.SessionResultStreamSubject(s.task.Route.OrgID, s.task.Route.SessionID)
		return t
	}
	t, err := dm.WorkerTaskSubject(s.task.Route.OrgID, s.task.Route.WorkerID)
	if err != nil {
		logs.Errorf("Failed to get worker task topic for stream sink: %v", err)
		return ""
	}
	return t
}

func (s *MQStreamSink) emitCompleted(ctx context.Context, msg events.MessageStreamMessage) {
	if s.task.Route.SessionID == "" {
		return
	}
	topic, err := dm.SessionCompletedSubject(s.task.Route.OrgID, s.task.Route.SessionID)
	if err != nil {
		logs.WarnContextf(ctx, "Failed to get session completed topic for stream sink: %v", err)
		return
	}
	if err := s.publisher.Publish(ctx, topic, msg); err != nil {
		logs.WarnContextf(ctx, "Failed to publish worker completed event to %s: %v", topic, err)
		return
	}
	if err := s.publisher.FlushRealtime(ctx); err != nil {
		logs.WarnContextf(ctx, "Failed to flush worker completed event to %s: %v", topic, err)
	}
}

func streamEventType(eventType events.EventType) events.StreamEventType {
	switch eventType {
	case events.EventStarted:
		return events.StreamEventRunStarted
	case events.EventCompleted:
		return events.StreamEventRunCompleted
	case events.EventFailed, events.EventCancelled:
		return events.StreamEventRunFailed
	case events.EventMessageDelta, events.EventReasoningDelta:
		return events.StreamEventMessageDelta
	case events.EventResult:
		return events.StreamEventMessageCompleted
	case events.EventToolCallStarted:
		return events.StreamEventToolCallStarted
	case events.EventToolCallArguments:
		return events.StreamEventToolCallDelta
	case events.EventToolCallOutput, events.EventToolCallCompleted:
		return events.StreamEventToolCallFinished
	case events.EventToolCallFailed:
		return events.StreamEventToolCallFinished
	default:
		return events.StreamEventMessageDelta
	}
}

package taskconsumer

import (
	"context"
	"testing"

	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
	"github.com/insmtx/Leros/backend/pkg/dm"
)

func TestMQStreamSinkPublishesCompletedEventToSessionCompletedTopic(t *testing.T) {
	tests := []struct {
		name       string
		eventType  events.EventType
		wantStream events.StreamEventType
	}{
		{
			name:       "run completed",
			eventType:  events.EventCompleted,
			wantStream: events.StreamEventRunCompleted,
		},
		{
			name:       "run failed",
			eventType:  events.EventFailed,
			wantStream: events.StreamEventRunFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orgID := uint(1)
			sessionID := "session_test"
			task := events.WorkerTaskMessage{
				Trace: events.TraceContext{
					TraceID:   "trace_test",
					RequestID: "request_test",
					TaskID:    "task_test",
					RunID:     "run_test",
				},
				Route: events.RouteContext{
					OrgID:     orgID,
					SessionID: sessionID,
					WorkerID:  2,
				},
			}
			publisher := &recordingRealtimePublisher{}
			sink := NewMQStreamSink(publisher, task)

			err := sink.Emit(context.Background(), &events.Event{
				ID:      "event_test",
				Type:    tt.eventType,
				RunID:   "run_test",
				TraceID: "trace_test",
				Seq:     7,
				Content: "done",
			})
			if err != nil {
				t.Fatalf("Emit() error = %v", err)
			}

			streamTopic, _ := dm.SessionResultStreamSubject(orgID, sessionID)
			completedTopic, _ := dm.SessionCompletedSubject(orgID, sessionID)
			if len(publisher.realtimeCalls) != 1 {
				t.Fatalf("expected one stream realtime publish, got %d", len(publisher.realtimeCalls))
			}
			if len(publisher.persistentCalls) != 1 {
				t.Fatalf("expected one completed persistent publish, got %d", len(publisher.persistentCalls))
			}
			if publisher.flushes != 1 {
				t.Fatalf("expected one realtime flush, got %d", publisher.flushes)
			}
			if publisher.realtimeCalls[0].topic != streamTopic {
				t.Fatalf("expected realtime publish to stream topic %q, got %q", streamTopic, publisher.realtimeCalls[0].topic)
			}
			if publisher.persistentCalls[0].topic != completedTopic {
				t.Fatalf("expected persistent publish to completed topic %q, got %q", completedTopic, publisher.persistentCalls[0].topic)
			}
			completedMsg, ok := publisher.persistentCalls[0].event.(events.MessageStreamMessage)
			if !ok {
				t.Fatalf("expected completed publish event type %T, got %T", completedMsg, publisher.persistentCalls[0].event)
			}
			if completedMsg.Body.Event != tt.wantStream {
				t.Fatalf("expected completed event %q, got %q", tt.wantStream, completedMsg.Body.Event)
			}
			if completedMsg.Trace.TaskID != task.Trace.TaskID || completedMsg.Trace.RunID != task.Trace.RunID {
				t.Fatalf("completed trace mismatch: got task_id=%q run_id=%q", completedMsg.Trace.TaskID, completedMsg.Trace.RunID)
			}
		})
	}
}

type recordingRealtimePublisher struct {
	realtimeCalls   []publishCall
	persistentCalls []publishCall
	flushes         int
}

type publishCall struct {
	topic string
	event any
}

func (p *recordingRealtimePublisher) Publish(_ context.Context, topic string, event any) error {
	p.persistentCalls = append(p.persistentCalls, publishCall{
		topic: topic,
		event: event,
	})
	return nil
}

func (p *recordingRealtimePublisher) PublishRealtime(_ context.Context, topic string, event any) error {
	p.realtimeCalls = append(p.realtimeCalls, publishCall{
		topic: topic,
		event: event,
	})
	return nil
}

func (p *recordingRealtimePublisher) FlushRealtime(context.Context) error {
	p.flushes++
	return nil
}

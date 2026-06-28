package eventpub

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/nats-io/nats.go"

	"github.com/insmtx/Leros/backend/agent"
	assistantdomain "github.com/insmtx/Leros/backend/internal/assistant/domain"
	"github.com/insmtx/Leros/backend/pkg/messaging"
)

type publisherRecorder struct {
	contextErr error
	topic      string
	event      any
}

func TestNATSEventSinkRoutesBothLanesAndPreservesRawToolPayload(t *testing.T) {
	publisher := &publisherRecorder{}
	sink := NewNATSEventSink(publisher, RunEventContext{
		OrgID:     1,
		WorkerID:  2,
		SessionID: "session-1",
		RequestID: "request-1",
		TaskID:    "task-1",
	})
	if err := sink.Emit(context.Background(), &agent.Event{
		RunID:   "run-1",
		TraceID: "trace-1",
		Seq:     1,
		Type:    "tool_call.started",
		Payload: json.RawMessage(`{"tool_call_id":"tool-1","name":"read","arguments":{"path":"README.md"}}`),
	}); err != nil {
		t.Fatalf("Emit(tool) error = %v", err)
	}
	if !strings.Contains(publisher.topic, ".run.stream") {
		t.Fatalf("tool topic = %q", publisher.topic)
	}
	streamEvent, ok := publisher.event.(messaging.RunEvent)
	if !ok || streamEvent.Body.Payload.ToolCall == nil {
		t.Fatalf("stream event = %#v", publisher.event)
	}
	encoded, err := json.Marshal(streamEvent)
	if err != nil {
		t.Fatalf("marshal stream event: %v", err)
	}
	if !strings.Contains(string(encoded), `"arguments":{"path":"README.md"}`) {
		t.Fatalf("raw arguments changed JSON shape: %s", encoded)
	}

	if err := sink.Emit(context.Background(), &agent.Event{
		RunID:   "run-1",
		TraceID: "trace-1",
		Seq:     2,
		Type:    "run.failed",
		Content: "provider failed",
		Payload: json.RawMessage(`{
			"status":"failed",
			"error":"provider failed",
			"usage":{"total_tokens":9},
			"artifacts":[{"artifact_id":"artifact-1"}],
			"events":[{"seq":1,"type":"tool_call.started","payload":{"tool_call_id":"tool-1"}}]
		}`),
	}); err != nil {
		t.Fatalf("Emit(failed) error = %v", err)
	}
	if !strings.Contains(publisher.topic, ".run.state") {
		t.Fatalf("terminal topic = %q", publisher.topic)
	}
	stateEvent, ok := publisher.event.(messaging.RunEvent)
	if !ok || stateEvent.Body.RunCompleted == nil || stateEvent.Body.Error == nil {
		t.Fatalf("state event = %#v", publisher.event)
	}
	if stateEvent.Body.RunCompleted.Usage == nil ||
		stateEvent.Body.RunCompleted.Usage.TotalTokens != 9 ||
		len(stateEvent.Body.RunCompleted.Artifacts) != 1 ||
		len(stateEvent.Body.RunCompleted.Events) != 1 {
		t.Fatalf("terminal archive = %#v", stateEvent.Body.RunCompleted)
	}
}

func (p *publisherRecorder) Publish(ctx context.Context, topic string, event any) error {
	p.contextErr = ctx.Err()
	p.topic = topic
	p.event = event
	return nil
}

func (*publisherRecorder) Request(context.Context, string, any) (*nats.Msg, error) {
	return nil, nil
}

func TestNATSEventSinkMapsTerminalPayloadAndDetachedContext(t *testing.T) {
	publisher := &publisherRecorder{}
	sink := NewNATSEventSink(publisher, RunEventContext{
		OrgID:             1,
		WorkerID:          2,
		SessionID:         "session-1",
		RequestID:         "request-1",
		TaskID:            "task-1",
		ReplyToMessageIDs: []string{"1", "1", "2"},
	})
	payload, err := json.Marshal(assistantdomain.TerminalPayload{
		Status:    string(assistantdomain.RunStatusCancelled),
		Message:   "已取消",
		Error:     "provider stopped: context canceled",
		Usage:     &agent.Usage{TotalTokens: 7},
		Artifacts: []assistantdomain.ArtifactRecord{{ArtifactID: "artifact-1", Title: "report"}},
		Events: []assistantdomain.TerminalEventRecord{{
			Seq:       2,
			LastSeq:   4,
			Type:      "message.delta",
			Timestamp: 123,
			Payload:   json.RawMessage(`{"message_id":"m1"}`),
		}},
	})
	if err != nil {
		t.Fatalf("marshal terminal payload: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sink.Emit(ctx, &agent.Event{
		RunID:   "run-1",
		Seq:     5,
		Type:    "run.cancelled",
		Content: "已取消",
		Payload: json.RawMessage(payload),
	}); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}
	if publisher.contextErr != nil {
		t.Fatalf("publish context error = %v, want detached context", publisher.contextErr)
	}
	event, ok := publisher.event.(messaging.RunEvent)
	if !ok {
		t.Fatalf("published event type = %T", publisher.event)
	}
	if event.Body.Error == nil || event.Body.Error.Message != "provider stopped: context canceled" {
		t.Fatalf("terminal error = %#v", event.Body.Error)
	}
	if event.Body.RunCompleted == nil ||
		len(event.Body.RunCompleted.Artifacts) != 1 ||
		len(event.Body.RunCompleted.Events) != 1 {
		t.Fatalf("run completed payload = %#v", event.Body.RunCompleted)
	}
	if got := event.Body.RunCompleted.Events[0]; got.Seq != 2 || got.LastSeq != 4 || len(got.Payload) == 0 {
		t.Fatalf("archived event = %#v", got)
	}
	if got := event.Body.ReplyToMessageIDs; len(got) != 2 || got[0] != "1" || got[1] != "2" {
		t.Fatalf("reply IDs = %v", got)
	}
	if messaging.ClassifyRunEvent(event.Body.Event) != messaging.RunEventLaneState {
		t.Fatalf("terminal event lane = %s", messaging.ClassifyRunEvent(event.Body.Event))
	}
}

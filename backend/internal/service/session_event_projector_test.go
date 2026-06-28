package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/insmtx/Leros/backend/agent"
	"github.com/insmtx/Leros/backend/agent/runtime/events"
	"github.com/insmtx/Leros/backend/internal/api/dto"
	assistantdomain "github.com/insmtx/Leros/backend/internal/assistant/domain"
	"github.com/insmtx/Leros/backend/pkg/messaging"
	"github.com/insmtx/Leros/backend/types"
)

func TestProjectRunEventPreservesTerminalArchiveForAllStatuses(t *testing.T) {
	tests := []struct {
		name      string
		eventType messaging.RunEventType
		public    agent.EventType
		status    string
		errorText string
	}{
		{name: "completed", eventType: messaging.RunEventRunCompleted, public: events.EventCompleted, status: "completed"},
		{name: "failed", eventType: messaging.RunEventRunFailed, public: events.EventFailed, status: "failed", errorText: "provider failed"},
		{name: "cancelled", eventType: messaging.RunEventRunCancelled, public: events.EventCancelled, status: "cancelled", errorText: "context canceled"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runEvent := messaging.RunEvent{
				CreatedAt: time.UnixMilli(1234),
				Trace:     messaging.TraceContext{RunID: "run-1"},
				Route:     messaging.RouteContext{SessionID: "session-1"},
				Body: messaging.RunEventBody{
					Seq:   9,
					Event: test.eventType,
					RunCompleted: &messaging.RunCompletedPayload{
						Status: test.status,
						Result: messaging.RunResultPayload{Message: "result"},
						Artifacts: []messaging.ArtifactPayload{{
							ArtifactID: "artifact-1",
							Title:      "report",
						}},
						Usage: &messaging.UsagePayload{TotalTokens: 11},
						Events: []messaging.RunEventRecord{{
							Seq:       3,
							LastSeq:   5,
							Type:      "message.delta",
							Timestamp: 100,
							Payload:   json.RawMessage(`{"message_id":"m1"}`),
						}},
						Metadata: &messaging.RunMetadataPayload{Runtime: "codex"},
					},
				},
			}
			if test.errorText != "" {
				runEvent.Body.Error = &messaging.RunEventError{Message: test.errorText}
			}
			projected, ok := ProjectRunEvent(runEvent)
			if !ok || projected.Type != test.public {
				t.Fatalf("ProjectRunEvent() = %#v, %v", projected, ok)
			}
			payload, ok := projected.Payload.(dto.RunTerminalPayload)
			if !ok {
				t.Fatalf("payload type = %T", projected.Payload)
			}
			if payload.Status != test.status ||
				payload.Result.Message != "result" ||
				payload.Error != test.errorText ||
				payload.Usage == nil ||
				payload.Usage.TotalTokens != 11 ||
				len(payload.Artifacts) != 1 ||
				len(payload.Events) != 1 ||
				payload.Metadata == nil ||
				payload.Metadata.Runtime != "codex" {
				t.Fatalf("terminal payload = %#v", payload)
			}
		})
	}
}

func TestProjectRunEventRecordKeepsCancelledTypeAndTerminalPayload(t *testing.T) {
	raw, err := json.Marshal(assistantdomain.TerminalPayload{
		Status:  "cancelled",
		Message: "已取消",
		Error:   "context canceled",
		Usage:   &agent.Usage{TotalTokens: 5},
		Artifacts: []assistantdomain.ArtifactRecord{{
			ArtifactID: "artifact-1",
		}},
		Events: []assistantdomain.TerminalEventRecord{{
			Seq:     1,
			Type:    "message.delta",
			Payload: json.RawMessage(`{"content":"partial"}`),
		}},
	})
	if err != nil {
		t.Fatalf("marshal terminal payload: %v", err)
	}
	projected, ok := ProjectRunEventRecord("session-1", types.MessageChunk{
		Seq:       7,
		Type:      string(events.EventCancelled),
		Timestamp: 123,
		Payload:   raw,
	})
	if !ok || projected.Type != string(events.EventCancelled) {
		t.Fatalf("ProjectRunEventRecord() = %#v, %v", projected, ok)
	}
	payload, ok := projected.Payload.(dto.RunTerminalPayload)
	if !ok || payload.Status != "cancelled" || payload.Result.Message != "已取消" ||
		payload.Error != "context canceled" || payload.Usage == nil ||
		payload.Usage.TotalTokens != 5 || len(payload.Artifacts) != 1 || len(payload.Events) != 1 {
		t.Fatalf("terminal payload = %#v", projected.Payload)
	}
}

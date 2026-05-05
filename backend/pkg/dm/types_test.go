package dm

import (
	"encoding/json"
	"testing"
	"time"
)

func TestWorkerTaskMessageJSONShape(t *testing.T) {
	message := WorkerTaskMessage{
		ID:        "msg_1",
		Type:      MessageTypeWorkerTask,
		CreatedAt: time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC),
		Trace: TraceContext{
			TraceID:   "trace_1",
			RequestID: "req_1",
			TaskID:    "task_1",
			RunID:     "run_1",
		},
		Route: RouteContext{
			OrgID:     "1001",
			SessionID: "sess_1",
			WorkerID:  "worker_1",
		},
		Body: WorkerTaskBody{
			TaskType: TaskTypeAgentRun,
			Execution: ExecutionTarget{
				AssistantID: "assistant_1",
				AgentID:     "agent_1",
			},
			Input: TaskInput{
				Type: InputTypeMessage,
				Text: "hello",
			},
		},
	}

	body, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}

	if _, ok := got["task_id"]; ok {
		t.Fatalf("task_id should live under trace, got top-level field in %s", body)
	}
	if _, ok := got["org_id"]; ok {
		t.Fatalf("org_id should live under route, got top-level field in %s", body)
	}
	if got["type"] != string(MessageTypeWorkerTask) {
		t.Fatalf("unexpected message type: %#v", got["type"])
	}
	if _, ok := got["trace"].(map[string]any); !ok {
		t.Fatalf("expected trace object in %s", body)
	}
	if _, ok := got["route"].(map[string]any); !ok {
		t.Fatalf("expected route object in %s", body)
	}
	if _, ok := got["body"].(map[string]any); !ok {
		t.Fatalf("expected body object in %s", body)
	}
	bodyObject := got["body"].(map[string]any)
	if _, ok := bodyObject["target"]; ok {
		t.Fatalf("target should be named execution in %s", body)
	}
	if _, ok := bodyObject["execution"].(map[string]any); !ok {
		t.Fatalf("expected execution object in %s", body)
	}
}

func TestMessageStreamMessageJSONShape(t *testing.T) {
	message := MessageStreamMessage{
		ID:   "evt_1",
		Type: MessageTypeStream,
		Trace: TraceContext{
			TraceID: "trace_1",
			RunID:   "run_1",
		},
		Route: RouteContext{
			OrgID:     "1001",
			SessionID: "sess_1",
			WorkerID:  "worker_1",
		},
		Body: StreamBody{
			Seq:   1,
			Event: StreamEventMessageDelta,
			Payload: StreamPayload{
				Role:    MessageRoleAssistant,
				Content: "hello",
			},
		},
	}

	body, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}

	var got struct {
		Type MessageType `json:"type"`
		Body struct {
			Seq   int64           `json:"seq"`
			Event StreamEventType `json:"event"`
		} `json:"body"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}
	if got.Type != MessageTypeStream {
		t.Fatalf("got type %q, want %q", got.Type, MessageTypeStream)
	}
	if got.Body.Event != StreamEventMessageDelta {
		t.Fatalf("got event %q, want %q", got.Body.Event, StreamEventMessageDelta)
	}
	if got.Body.Seq != 1 {
		t.Fatalf("got seq %d, want 1", got.Body.Seq)
	}
}

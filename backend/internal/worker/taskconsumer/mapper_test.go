package taskconsumer

import (
	"testing"
	"time"

	"github.com/insmtx/Leros/backend/internal/worker/protocol"
)

func TestRequestFromWorkerTaskMapsWorkspaceContext(t *testing.T) {
	msg := protocol.WorkerTaskMessage{
		ID:        "msg_1",
		Type:      protocol.MessageTypeWorkerTask,
		CreatedAt: time.Now().UTC(),
		Trace: protocol.TraceContext{
			TraceID:   "trace_1",
			RequestID: "req_1",
			TaskID:    "task_1",
			RunID:     "run_1",
		},
		Route: protocol.RouteContext{
			OrgID:     42,
			SessionID: "sess_1",
			WorkerID:  7,
		},
		Body: protocol.WorkerTaskBody{
			TaskType: protocol.TaskTypeAgentRun,
			Execution: protocol.ExecutionTarget{
				AssistantID: "assistant_1",
			},
			Workspace: protocol.WorkspaceOptions{
				ProjectID: "project_1",
			},
			Input: protocol.TaskInput{
				Type: protocol.InputTypeMessage,
				Text: "hello",
			},
		},
		Metadata: map[string]any{
			"source": "test",
		},
	}

	req := RequestFromWorkerTask(msg)

	if req.Conversation.ID != "sess_1" {
		t.Fatalf("conversation id = %q, want sess_1", req.Conversation.ID)
	}
	if req.Workspace.OrgID != 42 {
		t.Fatalf("workspace org id = %d, want 42", req.Workspace.OrgID)
	}
	if req.Workspace.ProjectID != "project_1" {
		t.Fatalf("workspace project id = %q, want project_1", req.Workspace.ProjectID)
	}
	if req.Workspace.TaskID != "task_1" {
		t.Fatalf("workspace task id = %q, want task_1", req.Workspace.TaskID)
	}
	if req.Workspace.RequestID != "req_1" {
		t.Fatalf("workspace request id = %q, want req_1", req.Workspace.RequestID)
	}

	for _, key := range []string{"org_id", "worker_id", "session_id", "task_id", "request_id", "agent_id", "project_id"} {
		if _, ok := req.Metadata[key]; ok {
			t.Fatalf("metadata should not carry %q: %#v", key, req.Metadata)
		}
	}
}

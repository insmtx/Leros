package taskconsumer

import (
	"context"
	"testing"

	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/pkg/dm"
)

func TestTaskTopic(t *testing.T) {
	consumer, err := New(Config{OrgID: "1001", WorkerID: "worker_1"}, &noopSubscriber{}, nil, &noopRunner{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got, want := consumer.TaskTopic(), "org.1001.worker.worker_1.task"; got != want {
		t.Fatalf("TaskTopic() = %q, want %q", got, want)
	}
}

func TestRequestFromWorkerTask(t *testing.T) {
	msg := dm.WorkerTaskMessage{
		ID:   "msg_1",
		Type: dm.MessageTypeWorkerTask,
		Trace: dm.TraceContext{
			TraceID: "trace_1",
			TaskID:  "task_1",
			RunID:   "run_1",
		},
		Route: dm.RouteContext{
			OrgID:     "1001",
			SessionID: "session_1",
			WorkerID:  "worker_1",
		},
		Body: dm.WorkerTaskBody{
			TaskType: dm.TaskTypeAgentRun,
			Actor: dm.ActorContext{
				UserID:      "user_1",
				DisplayName: "Caleb",
				Channel:     "web",
			},
			Execution: dm.ExecutionTarget{
				AssistantID: "assistant_1",
				AgentID:     "agent_1",
				Skills:      []string{"code-review"},
				Tools:       []string{"skill.use"},
			},
			Input: dm.TaskInput{
				Type: dm.InputTypeTaskInstruction,
				Text: "review this change",
				Messages: []dm.ChatMessage{
					{Role: dm.MessageRoleUser, Content: "hello"},
				},
			},
			Runtime: dm.RuntimeOptions{
				Kind:    "claude",
				WorkDir: "/tmp/repo",
				MaxStep: 10,
			},
			Policy: dm.TaskPolicy{RequireApproval: true},
		},
	}

	req := RequestFromWorkerTask(msg)
	if req.RunID != "run_1" || req.TraceID != "trace_1" || req.TaskID != "task_1" {
		t.Fatalf("unexpected trace mapping: %+v", req)
	}
	if req.Assistant.ID != "assistant_1" || req.Assistant.Skills[0] != "code-review" {
		t.Fatalf("unexpected assistant mapping: %+v", req.Assistant)
	}
	if req.Actor.UserID != "user_1" || req.Actor.Channel != "web" {
		t.Fatalf("unexpected actor mapping: %+v", req.Actor)
	}
	if req.Conversation.ID != "session_1" {
		t.Fatalf("Conversation.ID = %q, want session_1", req.Conversation.ID)
	}
	if req.Input.Type != agent.InputTypeTaskInstruction || req.Input.Text != "review this change" {
		t.Fatalf("unexpected input mapping: %+v", req.Input)
	}
	if req.Runtime.Kind != "claude" || req.Runtime.WorkDir != "/tmp/repo" || req.Runtime.MaxStep != 10 {
		t.Fatalf("unexpected runtime mapping: %+v", req.Runtime)
	}
	if !req.Policy.RequireApproval {
		t.Fatalf("RequireApproval = false, want true")
	}
	if req.Metadata["agent_id"] != "agent_1" {
		t.Fatalf("metadata agent_id = %v, want agent_1", req.Metadata["agent_id"])
	}
}

type noopSubscriber struct{}

func (noopSubscriber) Subscribe(_ context.Context, _ string, _ func(any)) error {
	return nil
}

type noopRunner struct{}

func (noopRunner) Run(_ context.Context, _ *agent.RequestContext) (*agent.RunResult, error) {
	return &agent.RunResult{Status: agent.RunStatusCompleted}, nil
}

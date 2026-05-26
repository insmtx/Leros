package taskconsumer

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/insmtx/Leros/backend/internal/agent"
	"github.com/insmtx/Leros/backend/internal/worker/protocol"
	"github.com/insmtx/Leros/backend/pkg/leros"
)

func TestPrepareWorkspaceUsesProjectRepoWorkDir(t *testing.T) {
	workspaceRoot := t.TempDir()
	t.Setenv(leros.EnvWorkspaceRoot, workspaceRoot)

	req := &agent.RequestContext{}
	msg := protocol.WorkerTaskMessage{
		ID: "msg_1",
		Trace: protocol.TraceContext{
			RequestID: "req_1",
			TaskID:    "task_1",
		},
		Route: protocol.RouteContext{OrgID: 42},
		Body: protocol.WorkerTaskBody{
			Workspace: protocol.WorkspaceOptions{ProjectID: "project_1"},
		},
		CreatedAt: time.Now().UTC(),
	}

	plan, err := (&Consumer{}).prepareWorkspace(context.Background(), msg, req)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}
	if plan == nil {
		t.Fatal("expected project workspace plan")
	}
	expected := filepath.Join(workspaceRoot, "projects", "42", "project_1", "repo")
	if req.Runtime.WorkDir != expected {
		t.Fatalf("runtime work dir = %q, want %q", req.Runtime.WorkDir, expected)
	}
}

func TestPrepareWorkspaceFallsBackToWorkspaceTemp(t *testing.T) {
	workspaceRoot := t.TempDir()
	t.Setenv(leros.EnvWorkspaceRoot, workspaceRoot)

	req := &agent.RequestContext{}
	msg := protocol.WorkerTaskMessage{
		ID:        "msg_1",
		Route:     protocol.RouteContext{OrgID: 42},
		CreatedAt: time.Now().UTC(),
	}

	plan, err := (&Consumer{}).prepareWorkspace(context.Background(), msg, req)
	if err != nil {
		t.Fatalf("prepare workspace: %v", err)
	}
	if plan != nil {
		t.Fatalf("expected no project workspace plan, got %#v", plan)
	}
	expected := filepath.Join(workspaceRoot, "temp")
	if req.Runtime.WorkDir != expected {
		t.Fatalf("runtime work dir = %q, want %q", req.Runtime.WorkDir, expected)
	}
}

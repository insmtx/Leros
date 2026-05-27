package externalcli

import (
	"strings"
	"testing"

	"github.com/insmtx/Leros/backend/internal/agent"
	"github.com/insmtx/Leros/backend/pkg/leros"
)

func TestBuildPromptDoesNotLeakWorkspacePaths(t *testing.T) {
	workspaceRoot := t.TempDir()
	t.Setenv(leros.EnvWorkspaceRoot, workspaceRoot)

	prompt := buildPrompt(&agent.RequestContext{
		TaskID: "task_1",
		Workspace: agent.WorkspaceContext{
			OrgID:     42,
			ProjectID: "project_1",
			TaskID:    "task_1",
			RequestID: "req_1",
		},
		Input: agent.InputContext{
			Type: agent.InputTypeMessage,
			Text: "生成一个文件",
		},
	})

	if strings.Contains(prompt, workspaceRoot) {
		t.Fatalf("prompt leaked workspace root %q: %s", workspaceRoot, prompt)
	}
	if strings.Contains(prompt, "LEROS_ARTIFACT_FILE") {
		t.Fatalf("prompt should not contain LEROS_ARTIFACT_FILE, got %s", prompt)
	}
	if strings.Contains(prompt, "artifacts.jsonl") {
		t.Fatalf("prompt should not expose manifest path, got %s", prompt)
	}
	if !strings.Contains(prompt, "artifact_declare") {
		t.Fatalf("prompt should instruct artifact declaration, got %s", prompt)
	}
	if !strings.Contains(prompt, "完整路径") {
		t.Fatalf("prompt should require complete artifact paths, got %s", prompt)
	}
}

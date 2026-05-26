package externalcli

import (
	"strings"
	"testing"

	"github.com/insmtx/Leros/backend/internal/agent"
	"github.com/insmtx/Leros/backend/pkg/leros"
)

func TestBuildPromptIncludesArtifactContractWithoutAbsoluteWorkspacePaths(t *testing.T) {
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

	expected := []string{
		"## File Output Contract",
		"LEROS_ARTIFACT_FILE",
		"manifest 使用 JSON Lines",
		"最终交付文件不要写入临时目录、日志目录或缓存目录",
		"不要在最终回复中生成下载链接",
	}
	for _, text := range expected {
		if !strings.Contains(prompt, text) {
			t.Fatalf("expected prompt to contain %q, got %s", text, prompt)
		}
	}
	if strings.Contains(prompt, workspaceRoot) {
		t.Fatalf("prompt leaked workspace root %q: %s", workspaceRoot, prompt)
	}
	if strings.Contains(prompt, "artifacts.jsonl") {
		t.Fatalf("prompt should not expose manifest path, got %s", prompt)
	}
}

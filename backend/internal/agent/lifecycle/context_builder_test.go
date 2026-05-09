package lifecycle

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/insmtx/SingerOS/backend/internal/agent"
	skillcatalog "github.com/insmtx/SingerOS/backend/internal/skill/catalog"
)

type mockRuntimeProvider struct {
	skillsProvider skillcatalog.CatalogProvider
}

func (m *mockRuntimeProvider) SkillsProvider() skillcatalog.CatalogProvider {
	return m.skillsProvider
}

func TestContextBuilderBuildSystemPromptIncludesSkillsAndSession(t *testing.T) {
	catalog, err := skillcatalog.NewCatalog(fstest.MapFS{
		"code-review/SKILL.md": {
			Data: []byte(`---
name: code-review
description: Review code.
metadata:
  singeros:
    always: true
---
Always inspect diffs first.`),
		},
	})
	if err != nil {
		t.Fatalf("new skills catalog: %v", err)
	}

	builder := NewContextBuilder(ContextBuilder{
		BaseSystemPrompt: "Base runtime prompt.",
		Runtime: &mockRuntimeProvider{
			skillsProvider: skillcatalog.NewStaticCatalogProvider(catalog),
		},
	})
	prompt, err := builder.BuildSystemPrompt(context.Background(), &agent.RequestContext{
		Assistant: agent.AssistantContext{SystemPrompt: "Assistant-specific prompt."},
		Conversation: agent.ConversationContext{
			Messages: []agent.InputMessage{
				{Role: "user", Content: "请记住这个项目使用 Go。"},
			},
		},
	})
	if err != nil {
		t.Fatalf("build system prompt: %v", err)
	}

	for _, expected := range []string{
		"Base runtime prompt.",
		"Assistant-specific prompt.",
		"Available skills:",
		"## Skill: code-review",
		"<session-summary>",
		"请记住这个项目使用 Go。",
		"自我学习规则",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("expected prompt to contain %q, got %s", expected, prompt)
		}
	}
}

func TestShouldRunLearningCheck(t *testing.T) {
	result := &agent.RunResult{Status: agent.RunStatusCompleted}

	if !shouldRunLearningCheck(&agent.RequestContext{
		Input: agent.InputContext{Type: agent.InputTypeMessage, Text: "以后都按这个规范处理"},
	}, result, &RunTrace{}) {
		t.Fatalf("expected learning check for explicit user learning cue")
	}

	if !shouldRunLearningCheck(&agent.RequestContext{
		Input: agent.InputContext{Type: agent.InputTypeMessage, Text: "处理这个复杂任务"},
	}, result, &RunTrace{ToolCalls: 5}) {
		t.Fatalf("expected learning check after complex tool run")
	}

	if shouldRunLearningCheck(&agent.RequestContext{
		Input: agent.InputContext{Type: agent.InputTypeMessage, Text: "处理这个复杂任务"},
	}, result, &RunTrace{ToolCalls: 5, ToolNames: []string{toolNameMemory}}) {
		t.Fatalf("did not expect learning check after memory was already called")
	}
}

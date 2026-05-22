package lifecyclecontext

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/insmtx/Leros/backend/internal/agent"
	skillcatalog "github.com/insmtx/Leros/backend/internal/skill/catalog"
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
  leros:
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
				{Role: "user", Content: "remember this project uses Go"},
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
		"remember this project uses Go",
		"Self-learning rules",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("expected prompt to contain %q, got %s", expected, prompt)
		}
	}
}

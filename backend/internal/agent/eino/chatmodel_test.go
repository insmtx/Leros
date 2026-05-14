package eino

import (
	"context"
	"strings"
	"testing"

	"github.com/insmtx/Leros/backend/config"
	"github.com/insmtx/Leros/backend/types"
)

func TestNewChatModelRequiresModel(t *testing.T) {
	t.Parallel()

	_, err := NewChatModel(context.Background(), &config.LLMConfig{
		Provider: string(types.LLMProviderOpenAI),
		APIKey:   "sk-test",
	})
	if err == nil || !strings.Contains(err.Error(), "llm model is required") {
		t.Fatalf("expected model required error, got %v", err)
	}
}

func TestNewChatModelRejectsUnknownProvider(t *testing.T) {
	t.Parallel()

	_, err := NewChatModel(context.Background(), &config.LLMConfig{
		Provider: "unknown",
		APIKey:   "sk-test",
		Model:    "test-model",
	})
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("expected unsupported provider error, got %v", err)
	}
}

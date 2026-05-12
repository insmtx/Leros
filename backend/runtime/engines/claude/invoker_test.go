package claude

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/insmtx/Leros/backend/runtime/engines"
	"github.com/insmtx/Leros/backend/runtime/events"
)

func TestAdapterAskCurrentTime(t *testing.T) {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		t.Skip("claude CLI not found in PATH")
	}
	apiKey := firstNonEmptyEnv("LEROS_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("set LEROS_LLM_API_KEY to run the real claude adapter test")
	}

	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	adapter := NewAdapter(claudePath, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	handle, err := adapter.Run(ctx, engines.RunRequest{
		WorkDir: workDir,
		Prompt:  "请查询当前系统时间，并用一句中文回答。不要修改任何文件。",
		Model: engines.ModelConfig{
			Provider: "anthropic",
			APIKey:   apiKey,
			Model:    firstNonEmptyEnv("LEROS_LLM_MODEL"),
			BaseURL:  firstNonEmptyEnv("LEROS_LLM_BASE_URL"),
		},
		Timeout: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("run claude adapter: %v", err)
	}

	var finalEvent events.Event
	var result string
	for event := range handle.Events {
		t.Logf("received event: type=%s, content=%s", event.Type, event.Content)
		if event.Type == events.EventResult {
			result = strings.TrimSpace(event.Content)
		}
		finalEvent = event
	}
	if finalEvent.Type == events.EventFailed {
		t.Fatalf("claude execution failed: %s", finalEvent.Content)
	}
	if finalEvent.Type != events.EventCompleted {
		t.Fatalf("unexpected final event: %#v", finalEvent)
	}

	if result == "" {
		t.Fatal("expected non-empty claude result")
	}
	t.Logf("claude current time result: %s", result)
}

func TestParseClaudeLineEmitsResultEvent(t *testing.T) {
	state := &claudeStreamState{}
	event := parseClaudeLine(`{"type":"result","result":"final","is_error":false}`, state)
	if event.Type != events.EventResult || event.Content != "final" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if state.result != "final" || state.isError {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestParseClaudeLineTracksAssistantFallback(t *testing.T) {
	state := &claudeStreamState{}
	event := parseClaudeLine(`{"type":"assistant","message":{"content":[{"type":"text","text":"answer"}]}}`, state)
	if event.Type != events.EventMessageDelta || event.Content != "answer" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if state.lastAssistantText != "answer" {
		t.Fatalf("got %q, want answer", state.lastAssistantText)
	}
}

func TestClaudeFailureContentPrefersClaudeResult(t *testing.T) {
	err := errors.New("exit status 1")
	state := &claudeStreamState{result: "authentication failed"}

	content := claudeFailureContent(err, state, "stderr detail")
	if content != "authentication failed (exit status 1)" {
		t.Fatalf("got %q", content)
	}
}

func TestClaudeFailureContentFallsBackToStderr(t *testing.T) {
	err := errors.New("exit status 1")

	content := claudeFailureContent(err, &claudeStreamState{}, "stderr detail")
	if content != "stderr detail (exit status 1)" {
		t.Fatalf("got %q", content)
	}
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

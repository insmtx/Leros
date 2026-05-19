package claude

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insmtx/Leros/backend/engines"
	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
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

func TestParseClaudeLineAttachesUsageToResultEvent(t *testing.T) {
	state := &claudeStreamState{}
	parsed := parseClaudeLineEvents(`{"type":"result","result":"final","is_error":false,"usage":{"input_tokens":10,"cache_creation_input_tokens":2,"cache_read_input_tokens":30,"output_tokens":4}}`, state)
	if len(parsed) != 1 {
		t.Fatalf("expected result event, got %#v", parsed)
	}
	if parsed[0].Type != events.EventResult {
		t.Fatalf("expected event result, got %#v", parsed[0])
	}
	result, err := events.DecodePayload[events.MessageResultPayload](&parsed[0])
	if err != nil {
		t.Fatalf("decode result payload: %v", err)
	}
	if result.Message != "final" {
		t.Fatalf("unexpected result message: %q", result.Message)
	}
	if result.Usage == nil || result.Usage.InputTokens != 42 || result.Usage.OutputTokens != 4 || result.Usage.TotalTokens != 46 {
		t.Fatalf("unexpected usage payload: %#v", result.Usage)
	}
}

func TestBuildArgsAppendsSystemPrompt(t *testing.T) {
	args := buildArgs(engines.RunRequest{
		SystemPrompt: "system only",
		Prompt:       "user only",
	})

	value, ok := argValue(args, "--append-system-prompt")
	if !ok {
		t.Fatalf("expected --append-system-prompt in args: %#v", args)
	}
	if value != "system only" {
		t.Fatalf("expected system prompt arg value, got %q", value)
	}
	if args[len(args)-1] != "--print" {
		t.Fatalf("expected claude to run in print mode, got %#v", args)
	}
	if containsArg(args, "user only") {
		t.Fatalf("expected user prompt to be passed via stdin, got %#v", args)
	}
}

func TestBuildArgsSkipsEmptySystemPrompt(t *testing.T) {
	args := buildArgs(engines.RunRequest{
		SystemPrompt: "   ",
		Prompt:       "user only",
	})

	if _, ok := argValue(args, "--append-system-prompt"); ok {
		t.Fatalf("expected empty system prompt to be skipped: %#v", args)
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

func TestParseClaudeLineMapsMessageIDToUUID(t *testing.T) {
	state := &claudeStreamState{}
	parsed := parseClaudeLineEvents(`{"type":"assistant","message":{"id":"msg_provider_1","content":[{"type":"text","text":"answer"},{"type":"thinking","thinking":"reason"}]}}`, state)
	if len(parsed) != 2 {
		t.Fatalf("expected two events, got %+v", parsed)
	}
	messagePayload, err := events.DecodePayload[events.MessageDeltaPayload](&parsed[0])
	if err != nil {
		t.Fatalf("decode message payload: %v", err)
	}
	reasoningPayload, err := events.DecodePayload[events.MessageDeltaPayload](&parsed[1])
	if err != nil {
		t.Fatalf("decode reasoning payload: %v", err)
	}
	if _, err := uuid.Parse(messagePayload.MessageID); err != nil {
		t.Fatalf("message id should be uuid, got %q: %v", messagePayload.MessageID, err)
	}
	if reasoningPayload.MessageID != messagePayload.MessageID {
		t.Fatalf("expected text and reasoning to share message id, got %q and %q", messagePayload.MessageID, reasoningPayload.MessageID)
	}
}

func TestParseClaudeLineEmitsToolCallStarted(t *testing.T) {
	state := &claudeStreamState{}
	event := parseClaudeLine(`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"call_123","name":"Bash","input":{"command":"date","description":"查询当前系统时间"}}]}}`, state)
	if event.Type != events.EventToolCallStarted {
		t.Fatalf("unexpected event type: %#v", event)
	}

	content := decodeEventContent(t, event.Content)
	if content["tool_call_id"] != "call_123" || content["name"] != "Bash" {
		t.Fatalf("unexpected tool call content: %#v", content)
	}
	args, ok := content["arguments"].(map[string]any)
	if !ok || args["command"] != "date" {
		t.Fatalf("unexpected tool call arguments: %#v", content["arguments"])
	}
}

func TestParseClaudeLineEmitsToolCallCompleted(t *testing.T) {
	state := &claudeStreamState{toolNames: map[string]string{"call_123": "Bash"}}
	event := parseClaudeLine(`{"type":"user","message":{"role":"user","content":[{"tool_use_id":"call_123","type":"tool_result","content":"Thu May 14 14:19:24 CST 2026","is_error":false}]}}`, state)
	if event.Type != events.EventToolCallCompleted {
		t.Fatalf("unexpected event type: %#v", event)
	}

	content := decodeEventContent(t, event.Content)
	if content["tool_call_id"] != "call_123" || content["name"] != "Bash" || content["result"] != "Thu May 14 14:19:24 CST 2026" || content["is_error"] != false {
		t.Fatalf("unexpected tool result content: %#v", content)
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

func argValue(args []string, name string) (string, bool) {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == name {
			return args[i+1], true
		}
	}
	return "", false
}

func containsArg(args []string, value string) bool {
	for _, arg := range args {
		if arg == value {
			return true
		}
	}
	return false
}

func decodeEventContent(t *testing.T, content string) map[string]any {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		t.Fatalf("decode event content: %v", err)
	}
	return decoded
}

package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/insmtx/Leros/backend/engines"
	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
	"github.com/ygpkg/yg-go/logs"
)

// Invoker 启动 Codex CLI 进程。
type Invoker struct {
	binary  string   // codex 可执行文件路径
	baseEnv []string // 基础环境变量
}

// NewInvoker 创建 Codex CLI 调用器。
func NewInvoker(binary string, extraEnv map[string]string) *Invoker {
	return &Invoker{
		binary:  binary,
		baseEnv: engines.BuildBaseEnv(extraEnv),
	}
}

type codexEvent struct {
	Type     string     `json:"type"`
	ThreadID string     `json:"thread_id,omitempty"`
	Item     *codexItem `json:"item,omitempty"`
}

type codexItem struct {
	Type        string          `json:"type"`
	Text        json.RawMessage `json:"text,omitempty"`
	Command     string          `json:"command,omitempty"`
	CommandLine string          `json:"command_line,omitempty"`
	Name        string          `json:"name,omitempty"`
	Output      string          `json:"output,omitempty"`
}

// Run 启动 Codex CLI 进程并将 stdout/stderr 直接转换为引擎事件。
func (inv *Invoker) Run(ctx context.Context, req engines.RunRequest) (engines.Process, <-chan events.Event, error) {
	threadID, resume := resolveThread(req.SessionID, req.Resume)
	args := buildArgs(threadID, resume, req)

	execCtx := ctx
	cancel := func() {}
	if req.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, req.Timeout)
	}

	cmd := exec.CommandContext(execCtx, inv.binary, args...)
	cmd.Dir = req.WorkDir
	cmd.Env = engines.BuildRunEnv(inv.baseEnv, req.ExtraEnv, req.Model)
	if !resume {
		cmd.Stdin = strings.NewReader(req.Prompt)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("open codex stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("open codex stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("start codex: %w", err)
	}

	evtChan := make(chan events.Event, 16)
	proc := engines.NewCmdProcess(cmd)
	evtChan <- events.Event{Type: events.EventStarted}

	go func() {
		defer close(evtChan)
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			scanStdout(ctx, stdout, evtChan)
		}()
		go func() {
			defer wg.Done()
			scanPlainOutput(ctx, stderr, evtChan, events.EventMessageDelta)
		}()

		err := cmd.Wait()
		wg.Wait()
		if err != nil {
			evtChan <- events.Event{Type: events.EventFailed, Content: err.Error()}
			return
		}
		evtChan <- events.Event{Type: events.EventCompleted}
	}()

	return proc, evtChan, nil
}

func scanStdout(ctx context.Context, r interface{ Read([]byte) (int, error) }, evtChan chan<- events.Event) {
	engines.ScanJSONLines(r, func(line string) bool {
		event := parseCodexLine(line)
		if event.Type == "" {
			return true
		}
		return sendEvent(ctx, evtChan, event)
	})
}

func parseCodexLine(line string) events.Event {
	logs.Infof("Parse Codex line: %s", line)
	line = strings.TrimSpace(line)
	if line == "" {
		return events.Event{}
	}
	var event codexEvent
	if json.Unmarshal([]byte(line), &event) != nil {
		return events.Event{Type: events.EventMessageDelta, Content: line}
	}
	if event.Type == "thread.started" && event.ThreadID != "" {
		return events.Event{Type: engines.EventProviderSessionStarted, Content: event.ThreadID}
	}
	if event.Item == nil {
		return events.Event{}
	}

	item := event.Item
	switch item.Type {
	case "agent_message":
		text := decodeCodexText(item.Text)
		if text == "" {
			return events.Event{}
		}
		eventType := events.EventMessageDelta
		if event.Type == "item.completed" {
			eventType = events.EventResult
		}
		return events.Event{Type: eventType, Content: text}
	case "command_execution", "tool_call", "shell_command":
		command := firstNonEmptyString(item.Command, item.CommandLine, item.Name)
		if command != "" {
			return events.Event{Type: events.EventMessageDelta, Content: "$ " + command}
		}
	case "command_output", "tool_output", "shell_output":
		output := firstNonEmptyString(item.Output, decodeCodexText(item.Text))
		if output != "" {
			return events.Event{Type: events.EventMessageDelta, Content: truncateOutput(output, 300)}
		}
	}
	return events.Event{}
}

func decodeCodexText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var parts []any
	if json.Unmarshal(raw, &parts) == nil {
		var b strings.Builder
		for _, part := range parts {
			if value, ok := part.(string); ok {
				b.WriteString(value)
			}
		}
		return b.String()
	}
	return ""
}

func scanPlainOutput(ctx context.Context, r interface{ Read([]byte) (int, error) }, evtChan chan<- events.Event, eventType events.EventType) {
	engines.ScanJSONLines(r, func(line string) bool {
		line = strings.TrimSpace(line)
		if line == "" {
			return true
		}
		return sendEvent(ctx, evtChan, events.Event{Type: eventType, Content: line})
	})
}

func sendEvent(ctx context.Context, evtChan chan<- events.Event, event events.Event) bool {
	select {
	case <-ctx.Done():
		return false
	case evtChan <- event:
		return true
	}
}

func truncateOutput(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "..."
}

func buildArgs(threadID string, resume bool, req engines.RunRequest) []string {
	args := []string{"exec"}
	args = append(args, singerProviderConfigArgs(req)...)
	if req.Model.Model != "" {
		args = append(args, "--model", req.Model.Model)
	}
	if resume && threadID != "" {
		args = append(args, "resume", threadID, "--json", "--skip-git-repo-check", "--dangerously-bypass-approvals-and-sandbox")
		if req.Prompt != "" {
			args = append(args, req.Prompt)
		}
		return args
	}
	return append(args, "-", "--json", "--skip-git-repo-check", "--dangerously-bypass-approvals-and-sandbox")
}

func singerProviderConfigArgs(req engines.RunRequest) []string {
	baseURL := firstNonEmptyString(
		req.Model.BaseURL,
		envValue(req.ExtraEnv, "OPENAI_API_BASE"),
		envValue(req.ExtraEnv, "OPENAI_BASE_URL"),
		os.Getenv("OPENAI_API_BASE"),
		os.Getenv("OPENAI_BASE_URL"),
	)
	return []string{
		"-c", `model_provider="singer"`,
		"-c", `model_providers.singer.name="singer"`,
		"-c", fmt.Sprintf(`model_providers.singer.base_url=%q`, baseURL),
		"-c", `model_providers.singer.env_key="OPENAI_API_KEY"`,
		"-c", `model_providers.singer.wire_api="chat"`,
	}
}

func envValue(entries []string, key string) string {
	prefix := key + "="
	for _, entry := range entries {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix)
		}
	}
	return ""
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func resolveThread(sessionID string, resume bool) (string, bool) {
	if !resume {
		return "", false
	}
	threadID := strings.TrimSpace(sessionID)
	return threadID, threadID != ""
}

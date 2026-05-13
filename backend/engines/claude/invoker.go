package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/insmtx/Leros/backend/engines"
	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
	"github.com/ygpkg/yg-go/logs"
)

// Invoker 启动 Claude Code 进程。
type Invoker struct {
	binary  string
	baseEnv []string
}

// NewInvoker 创建 Claude Code 调用器。
func NewInvoker(binary string, extraEnv map[string]string) *Invoker {
	return &Invoker{
		binary:  binary,
		baseEnv: engines.BuildBaseEnv(extraEnv),
	}
}

type streamEvent struct {
	Type    string         `json:"type"`
	Message *streamMessage `json:"message,omitempty"`
	Result  string         `json:"result,omitempty"`
	IsError bool           `json:"is_error,omitempty"`
}

type streamMessage struct {
	Content []streamContent `json:"content"`
}

type streamContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Name string `json:"name,omitempty"`
}

// Run 启动 Claude Code 进程并将 stdout/stderr 直接转换为引擎事件。
func (inv *Invoker) Run(ctx context.Context, req engines.RunRequest) (engines.Process, <-chan events.Event, error) {
	args := buildArgs(req)

	execCtx := ctx
	cancel := func() {}
	if req.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, req.Timeout)
	}

	cmd := exec.CommandContext(execCtx, inv.binary, args...)
	cmd.Dir = req.WorkDir
	cmd.Stdin = strings.NewReader(req.Prompt)
	cmd.Env = engines.BuildRunEnv(inv.baseEnv, req.ExtraEnv, req.Model)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("open claude stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("open claude stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("start claude: %w", err)
	}

	evtChan := make(chan events.Event, 16)
	proc := engines.NewCmdProcess(cmd)
	evtChan <- events.Event{Type: events.EventStarted}

	go func() {
		defer close(evtChan)
		defer cancel()

		parseState := &claudeStreamState{}
		var stderrText string
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			scanClaudeStdout(ctx, stdout, evtChan, parseState)
		}()
		go func() {
			defer wg.Done()
			stderrText = scanPlainOutput(ctx, stderr, evtChan, events.EventMessageDelta)
		}()

		err := cmd.Wait()
		wg.Wait()
		if err != nil {
			evtChan <- events.Event{Type: events.EventFailed, Content: claudeFailureContent(err, parseState, stderrText)}
			return
		}
		if parseState.isError {
			if parseState.result == "" {
				parseState.result = "claude execution failed"
			}
			evtChan <- events.Event{Type: events.EventFailed, Content: parseState.result}
			return
		}
		if parseState.result == "" && parseState.lastAssistantText != "" {
			if !sendEvent(ctx, evtChan, events.Event{Type: events.EventResult, Content: parseState.lastAssistantText}) {
				return
			}
		}
		evtChan <- events.Event{Type: events.EventCompleted}
	}()

	return proc, evtChan, nil
}

type claudeStreamState struct {
	result            string
	isError           bool
	lastAssistantText string
}

func scanClaudeStdout(ctx context.Context, r interface{ Read([]byte) (int, error) }, evtChan chan<- events.Event, state *claudeStreamState) {
	engines.ScanJSONLines(r, func(line string) bool {
		event := parseClaudeLine(line, state)
		if event.Type == "" {
			return true
		}
		return sendEvent(ctx, evtChan, event)
	})
}

func parseClaudeLine(line string, state *claudeStreamState) events.Event {
	logs.Infof("Parse Claude line: %s", line)
	line = strings.TrimSpace(line)
	if line == "" {
		return events.Event{}
	}
	var event streamEvent
	if json.Unmarshal([]byte(line), &event) != nil {
		return events.Event{Type: events.EventMessageDelta, Content: line}
	}
	switch event.Type {
	case "assistant":
		if event.Message == nil {
			return events.Event{}
		}
		var b strings.Builder
		for _, block := range event.Message.Content {
			switch block.Type {
			case "text":
				if block.Text != "" {
					state.lastAssistantText = block.Text
					b.WriteString(block.Text)
				}
			case "tool_use":
				if block.Name != "" {
					b.WriteString("[调用工具: ")
					b.WriteString(block.Name)
					b.WriteString("]")
				}
			}
		}
		if b.Len() == 0 {
			return events.Event{}
		}
		return events.Event{Type: events.EventMessageDelta, Content: b.String()}
	case "result":
		state.result = event.Result
		state.isError = event.IsError
		if event.IsError || event.Result == "" {
			return events.Event{}
		}
		return events.Event{Type: events.EventResult, Content: event.Result}
	}
	return events.Event{}
}

func scanPlainOutput(ctx context.Context, r interface{ Read([]byte) (int, error) }, evtChan chan<- events.Event, eventType events.EventType) string {
	var output strings.Builder
	engines.ScanJSONLines(r, func(line string) bool {
		line = strings.TrimSpace(line)
		if line == "" {
			return true
		}
		if output.Len() > 0 {
			output.WriteString("\n")
		}
		output.WriteString(line)
		return sendEvent(ctx, evtChan, events.Event{Type: eventType, Content: line})
	})
	return output.String()
}

func sendEvent(ctx context.Context, evtChan chan<- events.Event, event events.Event) bool {
	select {
	case <-ctx.Done():
		return false
	case evtChan <- event:
		return true
	}
}

func buildArgs(req engines.RunRequest) []string {
	args := []string{
		"--dangerously-skip-permissions",
		"--verbose",
		"--output-format", "stream-json",
	}
	if req.Model.Model != "" {
		args = append(args, "--model", req.Model.Model)
	}
	if req.SessionID != "" {
		if req.Resume {
			args = append(args, "--resume", req.SessionID)
		} else {
			args = append(args, "--session-id", req.SessionID)
		}
	}
	return append(args, "--print")
}

func claudeFailureContent(err error, state *claudeStreamState, stderrText string) string {
	detail := ""
	if state != nil {
		detail = strings.TrimSpace(state.result)
	}
	if detail == "" {
		detail = strings.TrimSpace(stderrText)
	}
	if err == nil {
		return detail
	}
	if detail == "" {
		return err.Error()
	}
	return fmt.Sprintf("%s (%v)", detail, err)
}

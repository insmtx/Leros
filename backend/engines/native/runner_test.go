package native

import (
	"context"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"testing"
	"time"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/insmtx/Leros/backend/engines"
	"github.com/insmtx/Leros/backend/internal/runtime/deps"
	"github.com/insmtx/Leros/backend/internal/runtime/events"
	runtimetodo "github.com/insmtx/Leros/backend/internal/runtime/todo"
	skillcatalog "github.com/insmtx/Leros/backend/internal/skill/catalog"
	"github.com/insmtx/Leros/backend/pkg/leros"
	"github.com/insmtx/Leros/backend/tools"
	memorytools "github.com/insmtx/Leros/backend/tools/memory"
	nodetools "github.com/insmtx/Leros/backend/tools/node"
	skillmanagetools "github.com/insmtx/Leros/backend/tools/skill_manage"
	skillusetools "github.com/insmtx/Leros/backend/tools/skill_use"
	todotools "github.com/insmtx/Leros/backend/tools/todo"
	"github.com/ygpkg/yg-go/logs"
	"go.uber.org/zap/zapcore"
)

func TestRunnerBuildToolBindingMergesDefaultTools(t *testing.T) {
	registry := tools.NewRegistry()
	if err := memorytools.Register(registry); err != nil {
		t.Fatalf("register memory tools: %v", err)
	}
	if err := skillusetools.Register(registry); err != nil {
		t.Fatalf("register skill use tools: %v", err)
	}
	if err := skillmanagetools.Register(registry); err != nil {
		t.Fatalf("register skill manage tools: %v", err)
	}
	if err := todotools.Register(registry); err != nil {
		t.Fatalf("register todo tools: %v", err)
	}
	if err := nodetools.Register(registry); err != nil {
		t.Fatalf("register node tools: %v", err)
	}

	runner := &Runner{
		toolAdapter: newToolAdapter(registry),
	}
	binding := runner.buildToolBinding(engines.RunRequest{
		ExecutionID: "run_tools",
		Prompt:      "hello",
	}, events.NewNoopSink())

	expected := []string{
		memorytools.ToolNameMemory,
		skillusetools.ToolNameSkillUse,
		skillmanagetools.ToolNameSkillManage,
		todotools.ToolNameTodo,
		nodetools.ToolNameNodeShell,
		nodetools.ToolNameNodeFileRead,
		nodetools.ToolNameNodeFileWrite,
	}
	if got := strings.Join(binding.AllowedTools, ","); got != strings.Join(expected, ",") {
		t.Fatalf("unexpected allowed tools:\nwant: %v\n got: %v", expected, binding.AllowedTools)
	}
}

func TestToolInvokerInjectsTodoReporter(t *testing.T) {
	registry := tools.NewRegistry()
	if err := todotools.Register(registry); err != nil {
		t.Fatalf("register todo tool: %v", err)
	}

	var emitted []events.Event
	reporter := runtimetodo.NewTracker(runtimetodo.Options{
		RunID: "run_adapter",
		Sink: events.SinkFunc(func(_ context.Context, event *events.Event) error {
			emitted = append(emitted, *event)
			return nil
		}),
	})

	adapter := newToolAdapter(registry)
	specs, invoker, err := adapter.EinoTools(toolBinding{
		TodoReporter: reporter,
		AllowedTools: []string{todotools.ToolNameTodo},
	}, events.SinkFunc(func(_ context.Context, event *events.Event) error {
		emitted = append(emitted, *event)
		return nil
	}))
	if err != nil {
		t.Fatalf("build tools: %v", err)
	}
	einoTools := buildEinoTools(specs, invoker)
	if len(einoTools) != 1 {
		t.Fatalf("expected one tool, got %d", len(einoTools))
	}

	runnable, ok := einoTools[0].(interface {
		InvokableRun(context.Context, string, ...einotool.Option) (string, error)
	})
	if !ok {
		t.Fatalf("expected invokable tool, got %T", einoTools[0])
	}

	output, err := runnable.InvokableRun(context.Background(), `{"todos":[{"content":"Plan","status":"pending"}]}`)
	if err != nil {
		t.Fatalf("run tool: %v", err)
	}
	if output == "" {
		t.Fatalf("expected tool output")
	}
	if len(emitted) != 1 || emitted[0].Type != events.EventTodoSnapshot {
		t.Fatalf("expected todo snapshot, got %#v", emitted)
	}
}

func TestToolInvokerEmitsToolEventsForNonTodoTool(t *testing.T) {
	registry := tools.NewRegistry()
	if err := registry.Register(&mockTool{
		BaseTool: tools.NewBaseTool(
			"regular_tool",
			"Regular test tool",
			tools.Schema{Type: "object"},
		),
	}); err != nil {
		t.Fatalf("register mock tool: %v", err)
	}

	var emitted []events.Event
	adapter := newToolAdapter(registry)
	specs, invoker, err := adapter.EinoTools(toolBinding{
		AllowedTools: []string{"regular_tool"},
	}, events.SinkFunc(func(_ context.Context, event *events.Event) error {
		emitted = append(emitted, *event)
		return nil
	}))
	if err != nil {
		t.Fatalf("build tools: %v", err)
	}
	einoTools := buildEinoTools(specs, invoker)
	runnable, ok := einoTools[0].(interface {
		InvokableRun(context.Context, string, ...einotool.Option) (string, error)
	})
	if !ok {
		t.Fatalf("expected invokable tool, got %T", einoTools[0])
	}

	if _, err := runnable.InvokableRun(context.Background(), `{}`); err != nil {
		t.Fatalf("run tool: %v", err)
	}
	if len(emitted) != 2 ||
		emitted[0].Type != events.EventToolCallStarted ||
		emitted[1].Type != events.EventToolCallCompleted {
		t.Fatalf("expected regular tool call events, got %#v", emitted)
	}
}

func TestAgentRunRealModel(t *testing.T) {
	logs.SetLevel(zapcore.DebugLevel)

	apiKey := firstNonEmptyEnv("LEROS_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("set LEROS_LLM_API_KEY to run the real model agent test")
	}

	ctx, cancel := realModelTestContext(t)
	defer cancel()

	runtimeDeps, err := deps.New(ctx, deps.Options{})
	if err != nil {
		t.Fatalf("new runtime env: %v", err)
	}

	agt, err := NewRunner(ctx, runtimeDeps)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}

	eventsCh, err := agt.Run(ctx, engines.RunRequest{
		ExecutionID: "run_real_model_message",
		Prompt:      "Reply with exactly this text: Leros agent runtime ok",
		Model:       realModelOptions(apiKey),
	})
	if err != nil {
		t.Fatalf("run agent: %v", err)
	}
	message := collectResult(t, eventsCh)
	if strings.TrimSpace(message) == "" {
		t.Fatalf("expected non-empty model response")
	}
	if !strings.Contains(message, "Leros agent runtime ok") {
		t.Fatalf("unexpected model response: %s", message)
	}
}

func TestAgentRunNodeTool(t *testing.T) {
	logs.SetLevel(zapcore.DebugLevel)

	apiKey := firstNonEmptyEnv("LEROS_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("set LEROS_LLM_API_KEY to run the real model agent tool-call test")
	}

	ctx, cancel := realModelTestContext(t)
	defer cancel()
	registry := tools.NewRegistry()
	if err := nodetools.Register(registry); err != nil {
		t.Fatalf("register node tools: %v", err)
	}

	runtimeDeps, err := deps.New(ctx, deps.Options{})
	if err != nil {
		t.Fatalf("new runtime env: %v", err)
	}

	agt, err := NewRunner(ctx, runtimeDeps)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}

	eventsCh, err := agt.Run(ctx, engines.RunRequest{
		ExecutionID: "run_real_model_node_shell_time",
		SystemPrompt: strings.Join([]string{
			"You must use tools to complete the user task; do not answer without tool usage.",
			"node_shell executes commands in the current worker environment.",
		}, "\n"),
		Prompt: "Use a tool to query the current system time.",
		Model:  realModelOptions(apiKey),
	})
	if err != nil {
		t.Fatalf("run agent: %v", err)
	}
	message := collectResult(t, eventsCh)
	if strings.TrimSpace(message) == "" {
		t.Fatalf("expected non-empty model response")
	}

}

func TestAgentRunWeatherSkillQuery(t *testing.T) {
	logs.SetLevel(zapcore.DebugLevel)

	apiKey := firstNonEmptyEnv("LEROS_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("set LEROS_LLM_API_KEY to run the real model agent weather skill test")
	}

	ctx, cancel := realModelTestContext(t)
	defer cancel()
	skillDir := newBundledRuntimeSkillsCatalog(t)
	// Verify the weather skill is available via dynamic scan
	if _, err := skillcatalog.Get("weather"); err != nil {
		t.Fatalf("weather skill must be available in %s: %v", skillDir, err)
	}

	registry := tools.NewRegistry()
	if err := skillusetools.Register(registry); err != nil {
		t.Fatalf("register skill tools: %v", err)
	}
	if err := nodetools.Register(registry); err != nil {
		t.Fatalf("register node tools: %v", err)
	}

	runtimeDeps, err := deps.New(ctx, deps.Options{})
	if err != nil {
		t.Fatalf("new runtime env: %v", err)
	}

	agt, err := NewRunner(ctx, runtimeDeps)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}

	eventsCh, err := agt.Run(ctx, engines.RunRequest{
		ExecutionID: "run_real_model_weather_skill_shanghai",
		SystemPrompt: strings.Join([]string{
			"You must use tools to complete the user task; do not answer without tool usage.",
			"node_shell executes commands in the current worker environment.",
		}, "\n"),
		Prompt: "Use the weather skill to query the weather in Shanghai.",
		Model:  realModelOptions(apiKey),
	})
	if err != nil {
		t.Fatalf("run weather skill agent: %v", err)
	}
	message := collectResult(t, eventsCh)
	if strings.TrimSpace(message) == "" {
		t.Fatalf("expected non-empty model response")
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

func realModelOptions(apiKey string) engines.ModelConfig {
	return engines.ModelConfig{
		Provider: "openai",
		APIKey:   apiKey,
		Model:    firstNonEmptyEnv("LEROS_LLM_MODEL"),
		BaseURL:  firstNonEmptyEnv("LEROS_LLM_BASE_URL"),
	}
}

// collectResult reads events from the channel and returns the final message.
func collectResult(t *testing.T, eventsCh <-chan events.Event) string {
	t.Helper()
	var message string
	for event := range eventsCh {
		switch event.Type {
		case events.EventResult:
			message = event.Content
		case events.EventFailed:
			t.Fatalf("run failed: %s", event.Content)
		}
	}
	return message
}

func newBundledRuntimeSkillsCatalog(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := goruntime.Caller(0)
	if !ok {
		t.Fatalf("resolve current test file")
	}

	skillsDir := filepath.Join(filepath.Dir(currentFile), "..", "skills")
	// The skills dir is <workspace>/.leros/skills, so the workspace root is the parent of .leros
	workspaceRoot := filepath.Dir(filepath.Dir(skillsDir))
	t.Setenv(leros.EnvWorkspaceRoot, workspaceRoot)
	return skillsDir
}

func realModelTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	timeoutValue := strings.TrimSpace(os.Getenv("LEROS_TEST_TIMEOUT"))
	if timeoutValue == "" {
		timeoutValue = "3m"
	}
	if timeoutValue == "0" || strings.EqualFold(timeoutValue, "none") {
		return context.Background(), func() {}
	}

	timeout, err := time.ParseDuration(timeoutValue)
	if err != nil {
		t.Fatalf("parse LEROS_TEST_TIMEOUT: %v", err)
	}
	return context.WithTimeout(context.Background(), timeout)
}

type recordingEventSink struct {
	mu     sync.Mutex
	events []*events.Event
}

func (s *recordingEventSink) Emit(ctx context.Context, event *events.Event) error {
	if event == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	copied := *event
	logs.DebugContextf(ctx, "recordingEventSink event: type=%s run_id=%s seq=%d content=%s",
		copied.Type, copied.RunID, copied.Seq, copied.Content)
	s.events = append(s.events, &copied)
	return nil
}

type mockTool struct {
	tools.BaseTool
}

func (m *mockTool) Validate(input map[string]interface{}) error {
	return nil
}

func (m *mockTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	return tools.JSONString(map[string]interface{}{
		"tool": m.Name(),
	})
}

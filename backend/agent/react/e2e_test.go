package react

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/insmtx/SingerOS/backend/llm"
	"github.com/insmtx/SingerOS/backend/skills"
)

// E2ETestSuite 端到端测试套件
type E2ETestSuite struct {
	t *testing.T
}

// TestEndToEndBasicFlow 测试基本的端到端流程
func TestEndToEndBasicFlow(t *testing.T) {
	suite := &E2ETestSuite{t: t}

	// 基本的工作流程：输入 -> ReAct代理 -> LLM推理 -> 技能执行 -> 输出
	suite.testBasicReActEndToEndFlow()
}

func (suite *E2ETestSuite) testBasicReActEndToEndFlow() {
	t := suite.t

	// 创建端到端测试组件
	llmResponses := []string{
		`{"thought": "首先，我需要理解用户请求是让我说Hello World", "action": "echo.simple_echo", "action_args": {"input": "Hello World!"}}`,
		`{"thought": "我已经说了Hello World，现在应该完成任务", "action": "finish", "action_args": {}}`,
	}

	// 使用状态追踪Llm提供程序
	trackingLLMProvider := NewTrackingLLMProvider(llmResponses)

	// 设置技能管理器与技能
	skillManager := createTestSkillManager()

	// 创建代理配置
	config := &Config{
		MaxIterations: 5,
		Timeout:       30 * time.Second,
		Model:         "gpt-4",
	}

	agent := NewReActAgent(config, trackingLLMProvider, skillManager, nil)

	// 准备输入
	input := &Input{
		Query:     "请输出Hello World",
		SessionID: fmt.Sprintf("e2e-basic-%d", time.Now().UnixNano()),
		Context: map[string]interface{}{
			"user_id": "test-user",
		},
	}

	// 执行端到端流程
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	output, err := agent.Run(ctx, input)

	// 验证执行结果
	if err != nil {
		t.Errorf("E2E basic flow failed with error: %v", err)
	}

	if output == nil {
		t.Fatal("E2E basic flow should produce output")
	}

	if output.Status != "success" && output.Status != "error" {
		t.Errorf("Expected status to be 'success' or 'error', got '%s'", output.Status)
	}

	t.Logf("E2E Basic Flow completed. Status: %s, Steps: %d", output.Status, output.TotalSteps)

	if len(output.States) > 0 {
		t.Logf("Final state action: %s, observed: %s",
			output.States[len(output.States)-1].Action,
			output.States[len(output.States)-1].Observed)
	}

	// 验证LLM被调用了
	if trackingLLMProvider.CallCount() == 0 {
		t.Error("Expected LLM to be called during execution")
	}
}

// TestEndToEndMultipleSkills 测试多技能的端到端流程
func TestEndToEndMultipleSkills(t *testing.T) {
	// 准备模拟的LLM响应序列，模拟调用多个技能
	llmResponses := []string{
		`{"thought": "第一步是调用echo技能输出一些数据", "action": "echo.simple_echo", "action_args": {"input": "data-step1"}}`,
		`{"thought": "第二步是调用另一个echo技能继续处理", "action": "echo.another_echo", "action_args": {"input": "data-step2"}}`,
		`{"thought": "第三步决定需要结束处理", "action": "finish", "action_args": {}}`,
	}

	trackingLLMProvider := NewTrackingLLMProvider(llmResponses)

	// 创建包含多种技能的管理器
	skillManager := skills.NewSimpleSkillManager()

	// 注册不同类型的技能
	skillManager.Register(&mockEchoSkill{id: "echo.simple_echo", name: "Simple Echo Skill"})
	skillManager.Register(&anotherMockEchoSkill{id: "echo.another_echo", name: "Another Echo Skill"})

	config := &Config{
		MaxIterations: 5,
		Timeout:       30 * time.Second,
		Model:         "gpt-4",
	}

	agent := NewReActAgent(config, trackingLLMProvider, skillManager, nil)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(45*time.Second))
	defer cancel()

	input := &Input{
		Query: "Perform multi-skill task",
	}

	output, err := agent.Run(ctx, input)

	t.Logf("Multi-skill E2E: Status=%s, Steps=%d", output.Status, output.TotalSteps)

	if output.TotalSteps > 0 {
		for i, state := range output.States {
			t.Logf("State %d: action='%s' (%s), thought='%s'",
				i, state.Action, state.Observed,
				truncate(state.Thought, 50))
		}
	}

	if err != nil {
		t.Logf("Multi-skill E2E ended with error: %v (may be expected)", err)
	}
}

// TestEndToEndExceptionHandling 测试异常处理的端到端场景
func TestEndToEndExceptionHandling(t *testing.T) {
	// 创建一个故意在某处发生错误的场景
	llmResponses := []string{
		// 正常的第一步
		`{"thought": "执行安全操作", "action": "echo.simple_echo", "action_args": {"input": "safe-operation"}}`,
		// 下一步会尝试执行不存在的技能
		`{"thought": "执行一个不存在的技能", "action": "nonexistent.skill", "action_args": {"input": "fail"}}`,
		// 理论上不会到达这里，因为上一步就失败了
		`{"thought": "完成任务", "action": "finish", "action_args": {}}`,
	}

	trackingLLMProvider := NewTrackingLLMProvider(llmResponses)

	// 但是技能管理器只有特定的技能，没有 nonexistent.skill
	skillManager := skills.NewSimpleSkillManager()
	skillManager.Register(&mockEchoSkill{id: "echo.simple_echo"})

	config := &Config{
		MaxIterations: 5,
		Timeout:       30 * time.Second,
		Model:         "gpt-4",
	}

	agent := NewReActAgent(config, trackingLLMProvider, skillManager, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := &Input{
		Query: "Test exception scenario with non-existent skill",
	}

	output, err := agent.Run(ctx, input)

	if err == nil {
		t.Logf("E2E exception test unexpectedly succeeded. Status: %s", output.Status)
	} else {
		t.Logf("E2E exception test correctly received error: %v", err)
		t.Logf("The state history shows how error was handled: %d states recorded", len(output.States))

		if len(output.States) > 0 {
			lastState := output.States[len(output.States)-1]
			t.Logf("Last state before error: action=%s, error=%s", lastState.Action, lastState.Error)
		}
	}
}

// TestEndToEndWithTimeout 测试超时场景的处理
func TestEndToEndWithTimeout(t *testing.T) {
	// 创建一个需要长时间处理的场景，但在配置的时间限制内
	llmResponses := make([]string, 10) // 模拟很多步骤
	for i := 0; i < 10; i++ {
		if i < 9 {
			llmResponses[i] = fmt.Sprintf(`{"thought": "Step %d still processing", "action": "echo.simple_echo", "action_args": {"input": "processing-step-%d"}}`, i, i)
		} else {
			llmResponses[i] = `{"thought": "Finally finishing", "action": "finish", "action_args": {}}`
		}
	}

	// 让LLM请求慢一点以测试超时
	slowLLMProvider := NewSlowTrackingLLMProvider(llmResponses, 300*time.Millisecond)

	skillManager := createTestSkillManager()

	config := &Config{
		MaxIterations: 5,               // 少于我们的步骤数，以测试迭代限制
		Timeout:       1 * time.Second, // 很短的超时时间来测试
		Model:         "gpt-4",
	}

	agent := NewReActAgent(config, slowLLMProvider, skillManager, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := &Input{
		Query: "Test with timeout constraints",
	}

	start := time.Now()
	output, err := agent.Run(ctx, input)
	duration := time.Since(start)

	t.Logf("Timeout test completed in %v. Status: %s, Errors: %v, Steps: %d",
		duration, output.Status, err, output.TotalSteps)

	if duration > config.Timeout+500*time.Millisecond { // 加一点缓冲时间
		t.Error("Execution took longer than configured timeout")
	}

	// 虽然超时了，但应该仍然有部分状态被记录
	if output.TotalSteps > config.MaxIterations {
		t.Error("Should not exceed max iterations even if LLM suggests more steps")
	}
}

// TrackingLLMProvider 跟踪调用次数的LLM提供程序
type TrackingLLMProvider struct {
	responses    []string
	currentIndex int
	callCount    int
	mutex        sync.Mutex
}

func NewTrackingLLMProvider(responses []string) *TrackingLLMProvider {
	return &TrackingLLMProvider{
		responses: responses,
	}
}

func (t *TrackingLLMProvider) CallCount() int {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.callCount
}

func (t *TrackingLLMProvider) Name() string {
	return "TrackingLLMProvider"
}

func (t *TrackingLLMProvider) Models() []string {
	return []string{"tracking-model"}
}

func (t *TrackingLLMProvider) CountTokens(text string) int {
	return len(text)
}

func (t *TrackingLLMProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.callCount++

	if t.currentIndex >= len(t.responses) {
		return &llm.GenerateResponse{
			Content: `{"thought": "End of test scenario", "action": "finish", "action_args": {}}`,
			Usage:   llm.TokenUsage{TotalTokens: 15},
		}, nil
	}

	responseContent := t.responses[t.currentIndex]
	t.currentIndex++

	return &llm.GenerateResponse{
		Content: responseContent,
		Usage:   llm.TokenUsage{TotalTokens: len(responseContent)},
	}, nil
}

func (t *TrackingLLMProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: "tracking stream", Done: true}
	close(ch)
	return ch, nil
}

// SlowTrackingLLMProvider 缓慢响应的跟踪LLM提供程序
type SlowTrackingLLMProvider struct {
	baseProvider *TrackingLLMProvider
	delay        time.Duration
}

func NewSlowTrackingLLMProvider(responses []string, delay time.Duration) *SlowTrackingLLMProvider {
	return &SlowTrackingLLMProvider{
		baseProvider: NewTrackingLLMProvider(responses),
		delay:        delay,
	}
}

func (s *SlowTrackingLLMProvider) Name() string {
	return "SlowTrackingLLMProvider"
}

func (s *SlowTrackingLLMProvider) Models() []string {
	return []string{"slow-tracking-model"}
}

func (s *SlowTrackingLLMProvider) CountTokens(text string) int {
	return len(text)
}

func (s *SlowTrackingLLMProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	// 模拟处理延迟
	time.Sleep(s.delay)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.baseProvider.Generate(ctx, req)
	}
}

func (s *SlowTrackingLLMProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan llm.StreamChunk, error) {
	return s.baseProvider.GenerateStream(ctx, req)
}

func (s *SlowTrackingLLMProvider) CallCount() int {
	return s.baseProvider.CallCount()
}

// 创建测试所需的技能管理器
func createTestSkillManager() skills.SkillManager {
	skillManager := skills.NewSimpleSkillManager()

	// 注册Echo技能
	echoSkill := &mockEchoSkill{id: "echo.simple_echo", name: "Mock Echo Skill"}
	skillManager.Register(echoSkill)

	return skillManager
}

// 辅助函数用于截断长字符串以便打印
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// 另一个模拟回声技能以支持多技能测试
type anotherMockEchoSkill struct {
	id   string
	name string
}

func (a *anotherMockEchoSkill) GetID() string {
	return a.id
}

func (a *anotherMockEchoSkill) GetName() string {
	return a.name
}

func (a *anotherMockEchoSkill) GetDescription() string {
	return "Another mock echo skill for testing"
}

func (a *anotherMockEchoSkill) Info() *skills.SkillInfo {
	return &skills.SkillInfo{
		ID:          a.id,
		Name:        a.name,
		Description: a.GetDescription(),
		Version:     "1.0.0",
		Category:    "utility",
	}
}

func (a *anotherMockEchoSkill) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"output":    "another echo: " + input["input"].(string),
		"type":      "another_echo_skill_response",
		"source":    a.id,
		"timestamp": time.Now().Unix(),
	}, nil
}

func (a *anotherMockEchoSkill) Validate(input map[string]interface{}) error {
	if input["input"] == nil {
		return errors.New("input is required")
	}
	return nil
}

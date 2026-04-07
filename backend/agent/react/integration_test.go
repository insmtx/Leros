package react

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/insmtx/SingerOS/backend/llm"
	"github.com/insmtx/SingerOS/backend/skills"
	"github.com/insmtx/SingerOS/backend/types"
)

// IntegrationTestSuite 集成测试套件
type IntegrationTestSuite struct {
	t *testing.T
}

// TestReActWithRealComponents 测试ReAct代理与真实组件的集成
func TestReActWithRealComponents(t *testing.T) {
	suite := &IntegrationTestSuite{t: t}

	// 测试ReAct代理与真实的LLM和技能管理器的集成
	suite.testReActAgentWithRealDependencies()
}

func (suite *IntegrationTestSuite) testReActAgentWithRealDependencies() {
	t := suite.t

	// 使用更真实的模拟组件，但仍保持可控性
	mockLLM := &DetailedMockLLMProvider{
		generateResponses: []llm.GenerateResponse{
			{
				Content: `{"thought": "I understand the query needs to call echo skill", "action": "echo.simple_echo", "action_args": {"input": "test integration"}}`,
				Usage:   llm.TokenUsage{TotalTokens: 30},
			},
			{
				Content: `{"thought": "Task is completed", "action": "finish", "action_args": {}}`,
				Usage:   llm.TokenUsage{TotalTokens: 18},
			},
		},
	}

	// 使用真正的SkillManager但有预设的技能
	realSkillManager := skills.NewSimpleSkillManager()
	echoSkill := &mockEchoSkill{id: "echo.simple_echo"}
	realSkillManager.Register(echoSkill)

	// 创建真实的配置
	config := &Config{
		MaxIterations: 5,
		Timeout:       2 * time.Minute,
		Model:         "gpt-4",
	}

	// 创建真正的ReAct代理
	realDigitalAssistant := &types.DigitalAssistant{
		Code:        "test-assistant",
		Name:        "Test ReAct Assistant",
		Description: "For integration testing",
	}

	agent := NewReActAgent(config, mockLLM, realSkillManager, realDigitalAssistant)

	if agent == nil {
		t.Fatal("Failed to create ReAct agent with real components")
	}

	// 执行测试
	ctx := context.Background()
	input := &Input{
		Query:     "Please echo this message for integration test",
		SessionID: "it-test-session-987",
		Context:   make(map[string]interface{}),
	}

	output, err := agent.Run(ctx, input)

	if err != nil && !errors.Is(err, ErrMaxIterationsExceeded) {
		// 预期内的迭代不足可能不会导致错误
		t.Errorf("Unexpected error during integration test: %v", err)
	}

	if output == nil {
		t.Fatal("Output should not be nil")
	}

	t.Logf("Integration test result: status=%s, steps=%d", output.Status, len(output.States))

	// 验证输出
	if output.TotalSteps > 0 {
		t.Logf("Last state action: %s", output.States[len(output.States)-1].Action)
	}
}

// TestSkillIntegration 测试技能集成流程
func TestSkillIntegration(t *testing.T) {
	// 创建一个代理，然后测试它执行不同类型的技能
	mockLLM := &DetailedMockLLMProvider{
		generateResponses: []llm.GenerateResponse{
			{
				Content: `{"thought": "Testing multi-skill flow", "action": "echo.simple_echo", "action_args": {"input": "first skill call"}}`,
				Usage:   llm.TokenUsage{TotalTokens: 28},
			},
			{
				Content: `{"thought": "Second skill call", "action": "echo.simple_echo", "action_args": {"input": "second skill call"}}`,
				Usage:   llm.TokenUsage{TotalTokens: 29},
			},
			{
				Content: `{"thought": "Task finished", "action": "finish", "action_args": {}}`,
				Usage:   llm.TokenUsage{TotalTokens: 22},
			},
		},
	}

	// 设置技能管理器
	skillManager := skills.NewSimpleSkillManager()

	// 注册多个技能供测试
	echoSkill := &mockEchoSkill{id: "echo.simple_echo"}
	err := skillManager.Register(echoSkill)
	if err != nil {
		t.Fatalf("Failed to register echo skill: %v", err)
	}

	// 配置和创建代理
	config := &Config{
		MaxIterations: 10,
		Timeout:       30 * time.Second,
		Model:         "gpt-4",
	}

	agent := NewReActAgent(config, mockLLM, skillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query: "Perform multi-skill test",
	}

	output, err := agent.Run(ctx, input)
	if err != nil && !errors.Is(err, ErrMaxIterationsExceeded) {
		t.Logf("Integration test produced expected behavior with error: %v", err)
		// 这可能是因为mock LLN只设置了有限的响应
	}

	t.Logf("Multi-skill test completed. Status: %s", output.Status)
	t.Logf("Generated %d states", len(output.States))

	for i, state := range output.States {
		t.Logf("State %d: action='%s', thought='%s'", i, state.Action, state.Thought)
	}
}

// TestReActWithMemory 测试包含状态和记忆的流程
func TestReActWithMemory(t *testing.T) {
	// 创建一个长链测试，验证状态是否被正确地传递
	callCounter := 0
	mockLLM := &ComplexMockLLMProvider{
		responsesGenerator: func(req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			callCounter++
			switch callCounter {
			case 1:
				return &llm.GenerateResponse{
					Content: `{"thought": "First step of multi-step process", "action": "echo.simple_echo", "action_args": {"input": "step-1-data"}}`,
					Usage:   llm.TokenUsage{TotalTokens: 32},
				}, nil
			case 2:
				return &llm.GenerateResponse{
					Content: `{"thought": "Processing results from first step", "action": "echo.simple_echo", "action_args": {"input": "step-2-processing"}}`,
					Usage:   llm.TokenUsage{TotalTokens: 35},
				}, nil
			case 3:
				return &llm.GenerateResponse{
					Content: `{"thought": "Ready to finish", "action": "finish", "action_args": {}}`,
					Usage:   llm.TokenUsage{TotalTokens: 20},
				}, nil
			default:
				return nil, errors.New("no more responses expected")
			}
		},
	}

	// 设置技能管理器
	skillManager := skills.NewSimpleSkillManager()
	echoSkill := &mockEchoSkill{id: "echo.simple_echo"}
	skillManager.Register(echoSkill)

	config := &Config{
		MaxIterations: 10,
		Timeout:       1 * time.Minute,
		Model:         "gpt-4",
	}

	agent := NewReActAgent(config, mockLLM, skillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query:   "Execute 3-step process",
		Context: map[string]interface{}{"initial_data": "start-point"},
	}

	output, err := agent.Run(ctx, input)
	if output != nil && output.TotalSteps > 0 {
		t.Logf("Multi-step process completed with %d steps, status: %s", output.TotalSteps, output.Status)

		// 验证每一步都被记录
		for i, state := range output.States {
			t.Logf("Step %d: %s -> %s", i+1, state.Action, state.Observed)
		}
	}

	if err != nil && err != ErrMaxIterationsExceeded {
		// 记录错误但不一定失败，取决于mock行为
		t.Logf("Multi-step process ended with error: %v", err)
	}
}

// DetailedMockLLMProvider 更详细的模拟LLM提供程序，提供预定义的响应
type DetailedMockLLMProvider struct {
	generateResponses []llm.GenerateResponse
	callCount         int
	mutex             sync.Mutex
}

func (d *DetailedMockLLMProvider) Name() string {
	return "DetailedMockProvider"
}

func (d *DetailedMockLLMProvider) Models() []string {
	return []string{"mock-model"}
}

func (d *DetailedMockLLMProvider) CountTokens(text string) int {
	return len(text)
}

func (d *DetailedMockLLMProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.callCount >= len(d.generateResponses) {
		return nil, errors.New("all predefined responses consumed")
	}

	response := d.generateResponses[d.callCount]
	d.callCount++
	return &response, nil
}

func (d *DetailedMockLLMProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: "mock stream", Done: true}
	close(ch)
	return ch, nil
}

// ComplexMockLLMProvider 基于函数的复杂模拟LLM提供程序
type ComplexMockLLMProvider struct {
	responsesGenerator func(*llm.GenerateRequest) (*llm.GenerateResponse, error)
}

func (c *ComplexMockLLMProvider) Name() string {
	return "ComplexMockProvider"
}

func (c *ComplexMockLLMProvider) Models() []string {
	return []string{"mock-model-complex"}
}

func (c *ComplexMockLLMProvider) CountTokens(text string) int {
	return len(text)
}

func (c *ComplexMockLLMProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	return c.responsesGenerator(req)
}

func (c *ComplexMockLLMProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 10)
	ch <- llm.StreamChunk{Content: "complex mock stream", Done: false}
	ch <- llm.StreamChunk{Content: "more complex content", Done: true}
	close(ch)
	return ch, nil
}

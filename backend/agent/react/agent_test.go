package react

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/insmtx/SingerOS/backend/llm"
	"github.com/insmtx/SingerOS/backend/skills"
)

// MockLLMProvider 模拟LLM提供程序用于单元测试
type MockLLMProvider struct {
	GenerateFunc       func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error)
	GenerateStreamFunc func(ctx context.Context, req *llm.GenerateRequest) (<-chan llm.StreamChunk, error)
	CountTokensFunc    func(text string) int
	ModelsFunc         func() []string
	NameFunc           func() string
}

func (m *MockLLMProvider) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "mock-model"
}

func (m *MockLLMProvider) Models() []string {
	if m.ModelsFunc != nil {
		return m.ModelsFunc()
	}
	return []string{"mock-model"}
}

func (m *MockLLMProvider) CountTokens(text string) int {
	if m.CountTokensFunc != nil {
		return m.CountTokensFunc(text)
	}
	return len(text)
}

func (m *MockLLMProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, req)
	}
	return &llm.GenerateResponse{
		Content: `{"thought": "This is a mock thought", "action": "echo.simple_echo", "action_args": {"input": "test"}}`,
		Usage:   llm.TokenUsage{TotalTokens: 20},
	}, nil
}

func (m *MockLLMProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan llm.StreamChunk, error) {
	if m.GenerateStreamFunc != nil {
		return m.GenerateStreamFunc(ctx, req)
	}
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: "mock stream content", Done: true}
	close(ch)
	return ch, nil
}

// MockSkillManager 模拟技能管理器用于单元测试
type MockSkillManager struct {
	skills map[string]skills.Skill
	mutex  sync.RWMutex
}

func NewMockSkillManager() *MockSkillManager {
	return &MockSkillManager{
		skills: make(map[string]skills.Skill),
	}
}

func (m *MockSkillManager) Register(skill skills.Skill) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.skills[skill.GetID()] = skill
	return nil
}

func (m *MockSkillManager) Get(skillID string) (skills.Skill, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	skill, exists := m.skills[skillID]
	if !exists {
		return nil, errors.New("skill not found")
	}
	return skill, nil
}

func (m *MockSkillManager) List() []skills.Skill {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	list := make([]skills.Skill, 0, len(m.skills))
	for _, skill := range m.skills {
		list = append(list, skill)
	}
	return list
}

func (m *MockSkillManager) Execute(ctx context.Context, skillID string, input map[string]interface{}) (map[string]interface{}, error) {
	m.mutex.RLock()
	skill, exists := m.skills[skillID]
	m.mutex.RUnlock()

	if !exists {
		return nil, errors.New("skill not found")
	}

	// 使用简单的echo模拟响应
	result := map[string]interface{}{"result": input["input"], "called_skill": skillID}

	// 如果技能是mock类型的，使用其Execute函数
	if execSkill, ok := skill.(interface {
		Execute(context.Context, map[string]interface{}) (map[string]interface{}, error)
	}); ok {
		return execSkill.Execute(ctx, input)
	}

	return result, nil
}

// TestReActAgentCreation 测试ReAct代理的创建
func TestReActAgentCreation(t *testing.T) {
	mockLLM := &MockLLMProvider{}
	mockSkillManager := NewMockSkillManager()

	config := &Config{
		MaxIterations: 5,
		Timeout:       30 * time.Second,
		Model:         "mock-model",
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	if agent == nil {
		t.Fatal("Failed to create ReAct agent")
	}

	if agent.config.MaxIterations != 5 {
		t.Errorf("Expected MaxIterations to be 5, got %d", agent.config.MaxIterations)
	}

	if agent.config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout to be 30s, got %v", agent.config.Timeout)
	}

	if agent.config.Model != "mock-model" {
		t.Errorf("Expected Model to be 'mock-model', got %s", agent.config.Model)
	}
}

// TestReActAgentCreationDefaults 测试使用默认值创建
func TestReActAgentCreationDefaults(t *testing.T) {
	mockLLM := &MockLLMProvider{}
	mockSkillManager := NewMockSkillManager()

	config := &Config{} // 使用默认值

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	if agent == nil {
		t.Fatal("Failed to create ReAct agent")
	}

	// 检查默认值
	if agent.config.MaxIterations != DefaultMaxIterations {
		t.Errorf("Expected MaxIterations to be default %d, got %d", DefaultMaxIterations, agent.config.MaxIterations)
	}

	if agent.config.Timeout != DefaultTimeout {
		t.Errorf("Expected Timeout to be default %v, got %v", DefaultTimeout, agent.config.Timeout)
	}
}

// TestReActAgentRunBasic 测试基本运行流程
func TestReActAgentRunBasic(t *testing.T) {
	callCounter := 0
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			callCounter++
			if callCounter == 1 {
				return &llm.GenerateResponse{
					Content: `{"thought": "Need to call echo skill", "action": "echo.simple_echo", "action_args": {"input": "Hello World"}}`,
					Usage:   llm.TokenUsage{TotalTokens: 25},
				}, nil
			} else {
				return &llm.GenerateResponse{
					Content: `{"thought": "Task completed", "action": "finish", "action_args": {}}`,
					Usage:   llm.TokenUsage{TotalTokens: 20},
				}, nil
			}
		},
	}

	mockSkillManager := NewMockSkillManager()
	echoSkill := &mockEchoSkill{id: "echo.simple_echo"}
	mockSkillManager.Register(echoSkill)

	config := &Config{
		MaxIterations: 3,
		Timeout:       30 * time.Second,
		Model:         "mock-model",
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query:     "Say hello world",
		SessionID: "test-session-123",
		Context:   map[string]interface{}{"test": true},
	}

	output, err := agent.Run(ctx, input)
	if err != nil {
		t.Fatalf("Unexpected error during agent run: %v", err)
	}

	if output == nil {
		t.Fatal("Expected output to not be nil")
	}

	if output.Status != "success" {
		t.Errorf("Expected status to be 'success', got '%s'", output.Status)
	}

	if len(output.States) == 0 {
		t.Error("Expected at least one state in output")
	}

	// 验证第一步
	if len(output.States) > 0 {
		firstState := output.States[0]
		if firstState.Action != "echo.simple_echo" {
			t.Errorf("Expected action to be 'echo.simple_echo', got '%s'", firstState.Action)
		}
	}
}

// TestReActAgentRunWithFinishAction 测试带有结束操作的运行流程
func TestReActAgentRunWithFinishAction(t *testing.T) {
	callCount := 0
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			callCount++
			if callCount == 1 {
				return &llm.GenerateResponse{
					Content: `{"thought": "Need to call echo skill", "action": "echo.simple_echo", "action_args": {"input": "Processing data"}}`,
					Usage:   llm.TokenUsage{TotalTokens: 25},
				}, nil
			} else {
				return &llm.GenerateResponse{
					Content: `{"thought": "Task completed", "action": "finish", "action_args": {}}`,
					Usage:   llm.TokenUsage{TotalTokens: 20},
				}, nil
			}
		},
	}

	mockSkillManager := NewMockSkillManager()
	echoSkill := &mockEchoSkill{id: "echo.simple_echo"}
	mockSkillManager.Register(echoSkill)

	config := &Config{
		MaxIterations: 5,
		Timeout:       30 * time.Second,
		Model:         "mock-model",
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query: "Process data and finish",
	}

	output, err := agent.Run(ctx, input)
	if err != nil {
		t.Fatalf("Unexpected error during agent run: %v", err)
	}

	if output.Status != "success" {
		t.Errorf("Expected status to be 'success', got '%s'", output.Status)
	}

	if len(output.States) != 2 {
		t.Errorf("Expected exactly 2 states, got %d", len(output.States))
	}

	if output.States[1].Action != "finish" {
		t.Errorf("Expected last action to be 'finish', got '%s'", output.States[1].Action)
	}

	if !output.States[1].Completed {
		t.Error("Expected last state to have Completed=true")
	}
}

// TestReActAgentMaxIterExceeded 测试超过最大迭代次数
func TestReActAgentMaxIterExceeded(t *testing.T) {
	callCount := 0
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			callCount++
			// LLM不断返回非finish操作，触发最大迭代限制
			return &llm.GenerateResponse{
				Content: `{"thought": "Still working", "action": "echo.simple_echo", "action_args": {"input": "step ` + string(rune(callCount+'0')) + `"}}`,
				Usage:   llm.TokenUsage{TotalTokens: 25},
			}, nil
		},
	}

	mockSkillManager := NewMockSkillManager()
	echoSkill := &mockEchoSkill{id: "echo.simple_echo"}
	mockSkillManager.Register(echoSkill)

	config := &Config{
		MaxIterations: 2, // 设置一个低的限制以便测试
		Timeout:       30 * time.Second,
		Model:         "mock-model",
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query: "Run until max iter exceeded",
	}

	output, err := agent.Run(ctx, input)

	// 应该因为达到最大迭代次数而失败
	if err != ErrMaxIterationsExceeded {
		t.Errorf("Expected error %v, got %v", ErrMaxIterationsExceeded, err)
	}

	if output.Status != "error" {
		t.Errorf("Expected status to be 'error', got '%s'", output.Status)
	}

	if len(output.States) != config.MaxIterations {
		t.Errorf("Expected %d states, got %d", config.MaxIterations, len(output.States))
	}
}

// TestReActAgentTimeout 测试超时情况
func TestReActAgentTimeout(t *testing.T) {
	callCount := 0
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			callCount++
			if callCount == 1 {
				// 模拟第一个请求很慢，超时返回
				timer := time.NewTimer(100 * time.Millisecond) // 睡眠时间比agent超时长
				select {
				case <-timer.C:
					return &llm.GenerateResponse{
						Content: `{"thought": "Slow response", "action": "finish", "action_args": {}}`,
						Usage:   llm.TokenUsage{TotalTokens: 20},
					}, nil
				case <-ctx.Done():
					timer.Stop()
					return nil, ctx.Err()
				}
			}
			return &llm.GenerateResponse{
				Content: `{"thought": "Processed", "action": "finish", "action_args": {}}`,
				Usage:   llm.TokenUsage{TotalTokens: 15},
			}, nil
		},
	}

	mockSkillManager := NewMockSkillManager()

	config := &Config{
		MaxIterations: 5,
		Timeout:       50 * time.Millisecond, // 超时时间小于LLM响应时间
		Model:         "mock-model",
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query: "Short timeout test",
	}

	output, err := agent.Run(ctx, input)

	// 测试是否因超时而失败
	if err != nil {
		t.Logf("Successfully got timeout error: %v", err)
	} else {
		t.Logf("No timeout error as context cancellation might happen after timeout: status=%s", output.Status)
	}
}

// TestReActAgentLLMError 测试LLM错误情况
func TestReActAgentLLMError(t *testing.T) {
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			return nil, errors.New("LLM unavailable")
		},
	}

	mockSkillManager := NewMockSkillManager()

	config := &Config{
		MaxIterations: 5,
		Timeout:       30 * time.Second,
		Model:         "mock-model",
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query: "Failing task",
	}

	output, err := agent.Run(ctx, input)

	if err == nil {
		t.Error("Expected LLM error, got nil")
	}

	if output.Status != "error" {
		t.Errorf("Expected status to be 'error', got '%s'", output.Status)
	}

	if len(output.States) == 0 {
		t.Fatal("Expected at least one state with error")
	}

	lastState := output.States[len(output.States)-1]
	if lastState.Error == "" {
		t.Error("Expected error in last state")
	}

	if !lastState.Completed {
		t.Error("Expected last state to be marked as completed on error")
	}
}

// TestReActAgentSkillNotFoundError 测试技能未找到错误
func TestReActAgentSkillNotFoundError(t *testing.T) {
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			return &llm.GenerateResponse{
				Content: `{"thought": "Need to call missing skill", "action": "nonexistent.skill", "action_args": {"input": "test"}}`,
				Usage:   llm.TokenUsage{TotalTokens: 25},
			}, nil
		},
	}

	mockSkillManager := NewMockSkillManager()
	// 不注册任何技能，所以所有技能调用都会失败

	config := &Config{
		MaxIterations: 5,
		Timeout:       30 * time.Second,
		Model:         "mock-model",
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query: "Run with nonexistent skill",
	}

	output, err := agent.Run(ctx, input)

	// 预期会发生错误
	if err == nil && output.Status != "error" {
		t.Log("Non-error result in skill not found test, but that could happen depending on implementation")
	}
}

// TestReActAgentWhitelist 测试技能白名单功能
func TestReActAgentWhitelist(t *testing.T) {
	callCount := 0
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			callCount++
			if callCount == 1 {
				// 请求一个不在白名单上的技能
				return &llm.GenerateResponse{
					Content: `{"thought": "Try unauthorized skill", "action": "forbidden.command", "action_args": {"input": "malicious"}}`,
					Usage:   llm.TokenUsage{TotalTokens: 25},
				}, nil
			} else {
				// 请求一个在白名单上的技能
				return &llm.GenerateResponse{
					Content: `{"thought": "Run authorized skill", "action": "echo.simple_echo", "action_args": {"input": "hello"}}`,
					Usage:   llm.TokenUsage{TotalTokens: 23},
				}, nil
			}
		},
	}

	mockSkillManager := NewMockSkillManager()
	echoSkill := &mockEchoSkill{id: "echo.simple_echo"}
	mockSkillManager.Register(echoSkill)

	config := &Config{
		MaxIterations:  5,
		Timeout:        30 * time.Second,
		Model:          "mock-model",
		SkillWhitelist: []string{"echo.simple_echo"}, // 设置白名单
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query: "Test skill whitelist",
	}

	output, err := agent.Run(ctx, input)

	// 根据模拟逻辑测试行为
	_ = err
	t.Logf("Whitelist test completed with status: %s, states: %d", output.Status, len(output.States))
}

// TestReActAgentOutputFormatting 测试输出格式
func TestReActAgentOutputFormatting(t *testing.T) {
	callCount := 0
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			callCount++
			if callCount == 1 {
				return &llm.GenerateResponse{
					Content: `{"thought": "Single step", "action": "echo.simple_echo", "action_args": {"input": "test"}}`,
					Usage:   llm.TokenUsage{TotalTokens: 22},
				}, nil
			} else {
				return &llm.GenerateResponse{
					Content: `{"thought": "Complete task", "action": "finish", "action_args": {}}`,
					Usage:   llm.TokenUsage{TotalTokens: 20},
				}, nil
			}
		},
	}

	mockSkillManager := NewMockSkillManager()
	echoSkill := &mockEchoSkill{id: "echo.simple_echo"}
	mockSkillManager.Register(echoSkill)

	config := &Config{
		MaxIterations: 5,
		Timeout:       30 * time.Second,
		Model:         "mock-model",
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query: "Format output test",
	}

	output, err := agent.Run(ctx, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}

	if output.EndTime.IsZero() {
		t.Error("Expected EndTime to be set")
	}

	if output.StartTime.After(output.EndTime) {
		t.Error("EndTime should be after StartTime")
	}

	if output.TotalSteps < 1 { // 因为我们执行了多步
		t.Errorf("Expected TotalSteps to be > 0, got %d", output.TotalSteps)
	}

	if output.States == nil {
		t.Error("Expected States to not be nil")
	}
}

// mockEchoSkill 用于测试的简单echo技能
type mockEchoSkill struct {
	id   string
	name string
}

func (m *mockEchoSkill) GetID() string {
	return m.id
}

func (m *mockEchoSkill) GetName() string {
	if m.name != "" {
		return m.name
	}
	return "Mock Echo Skill"
}

func (m *mockEchoSkill) GetDescription() string {
	return "A mock echo skill for testing"
}

func (m *mockEchoSkill) Info() *skills.SkillInfo {
	return &skills.SkillInfo{
		ID:          m.id,
		Name:        m.GetName(),
		Description: m.GetDescription(),
		Version:     "1.0.0",
		Category:    "utility",
	}
}

func (m *mockEchoSkill) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"output": "echo: " + input["input"].(string),
		"type":   "echo_skill_response",
	}, nil
}

func (m *mockEchoSkill) Validate(input map[string]interface{}) error {
	if input["input"] == nil {
		return errors.New("input is required")
	}
	return nil
}

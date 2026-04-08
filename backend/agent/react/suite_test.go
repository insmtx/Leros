package react

import (
	"context"
	"testing"
	"time"

	"github.com/insmtx/SingerOS/backend/llm"
)

// TestSuite 运行所有ReAct代理测试套件
//
// 该测试套件覆盖以下方面：
// 1. ReAct状态管理
// 2. 推理循环
// 3. 技能集成
// 4. Agent执行流程
// 5. 异常处理
func TestSuite(t *testing.T) {
	// 设置所有测试所需的内容
	setupTestEnvironment(t)

	// 运行所有测试
	t.Run("TestReActAgentCreation", TestReActAgentCreation)
	t.Run("TestReActAgentCreationDefaults", TestReActAgentCreationDefaults)
	t.Run("TestReActAgentRunBasic", TestReActAgentRunBasic)
	t.Run("TestReActAgentRunWithFinishAction", TestReActAgentRunWithFinishAction)
	t.Run("TestReActAgentMaxIterExceeded", TestReActAgentMaxIterExceeded)
	t.Run("TestReActAgentTimeout", TestReActAgentTimeout)
	t.Run("TestReActAgentLLMError", TestReActAgentLLMError)
	t.Run("TestReActAgentSkillNotFoundError", TestReActAgentSkillNotFoundError)
	t.Run("TestReActAgentWhitelist", TestReActAgentWhitelist)
	t.Run("TestReActAgentOutputFormatting", TestReActAgentOutputFormatting)

	// 状态管理测试
	t.Run("TestReActStateManagement", TestReActStateManagement)
	t.Run("TestReActLongRunningState", TestReActLongRunningState)
	t.Run("TestReActStateErrorHandling", TestReActStateErrorHandling)
	t.Run("TestReActStateConcurrentAccess", TestReActStateConcurrentAccess)

	// 集成和端到端测试
	t.Run("TestReActWithRealComponents", TestReActWithRealComponents)
	t.Run("TestSkillIntegration", TestSkillIntegration)
	t.Run("TestReActWithMemory", TestReActWithMemory)
	t.Run("TestEndToEndBasicFlow", TestEndToEndBasicFlow)
	t.Run("TestEndToEndMultipleSkills", TestEndToEndMultipleSkills)
	t.Run("TestEndToEndExceptionHandling", TestEndToEndExceptionHandling)
	t.Run("TestEndToEndWithTimeout", TestEndToEndWithTimeout)

	teardownTestEnvironment(t)
}

// setupTestEnvironment 为测试设置环境
func setupTestEnvironment(t *testing.T) {
	// 确保必要的配置或测试数据已准备
	t.Log("Setting up test environment for ReAct Agent suite...")
}

// teardownTestEnvironment 清理测试后环境
func teardownTestEnvironment(t *testing.T) {
	// 清理测试过程中产生的临时数据
	t.Log("Cleaning up test environment for ReAct Agent suite...")
}

// BenchmarkReActAgentRun 基准测试基本运行性能
func BenchmarkReActAgentRun(b *testing.B) {
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			// 模拟返回基本响应
			return &llm.GenerateResponse{
				Content: `{"thought": "Benchmark test", "action": "finish", "action_args": {}}`,
				Usage:   llm.TokenUsage{TotalTokens: 15},
			}, nil
		},
	}

	mockSkillManager := NewMockSkillManager()

	config := &Config{
		MaxIterations: 3,
		Timeout:       30 * time.Second,
		Model:         "gpt-4",
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query: "Benchmark test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = agent.Run(ctx, input)
	}
}

// 注意：上面基准测试有一个小问题，context.Context不能这样使用
// 实际应用中应该使用正确的类型，下面是修复版本

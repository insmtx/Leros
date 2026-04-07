package react

import (
	"context"
	"testing"
	"time"

	"github.com/insmtx/SingerOS/backend/llm"
)

// BenchmarkReActAgentBasicRun 基准测试基本运行性能
func BenchmarkReActAgentBasicRun(b *testing.B) {
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			return &llm.GenerateResponse{
				Content: `{"thought": "Benchmark basic test thought", "action": "finish", "action_args": {}}`,
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
		Query: "Benchmark basic run",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = agent.Run(ctx, input)
	}
}

// BenchmarkReActAgentMultiStep 基准测试多步骤运行性能
func BenchmarkReActAgentMultiStep(b *testing.B) {
	callCount := 0
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			callCount++
			if callCount == 1 {
				return &llm.GenerateResponse{
					Content: `{"thought": "First step", "action": "echo.simple_echo", "action_args": {"input": "multi-step-test"}}`,
					Usage:   llm.TokenUsage{TotalTokens: 30},
				}, nil
			} else {
				return &llm.GenerateResponse{
					Content: `{"thought": "Finishing multi-step", "action": "finish", "action_args": {}}`,
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
		Timeout:       60 * time.Second,
		Model:         "gpt-4",
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query: "Benchmark multi-step run",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		callCount = 0 // 重置计数器
		_, _ = agent.Run(ctx, input)
	}
}

// BenchmarkReActAgentConcurrent 基准测试并发性能
func BenchmarkReActAgentConcurrent(b *testing.B) {
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			return &llm.GenerateResponse{
				Content: `{"thought": "Concurrent benchmark thought", "action": "finish", "action_args": {}}`,
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

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			input := &Input{
				Query: "Concurrent benchmark",
			}
			_, _ = agent.Run(ctx, input)
		}
	})
}

// BenchmarkProgrammingAgent 基准测试程序员助手性能
func BenchmarkProgrammingAgent(b *testing.B) {
	llmResponses := []string{
		`{"thought": "分析编程任务，首先查看需求", "action": "code.review", "action_args": {"code": "function hello() {\n  console.log('Hello World');\n}", "language": "JavaScript"}}`,
		`{"thought": "完成代码审查，准备输出结果", "action": "finish", "action_args": {}}`,
	}

	responseIndex := 0
	mockLLM := &MockLLMProvider{
		GenerateFunc: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
			idx := responseIndex % len(llmResponses)
			responseIndex++
			return &llm.GenerateResponse{
				Content: llmResponses[idx],
				Usage:   llm.TokenUsage{TotalTokens: 45},
			}, nil
		},
	}

	mockSkillManager := NewMockSkillManager()
	codeReviewSkill := NewCodeReviewSkill()
	mockSkillManager.Register(codeReviewSkill)

	config := &Config{
		MaxIterations:  10,
		Timeout:        60 * time.Second,
		Model:          "gpt-4",
		ThroughtPrompt: getSimpleThoughtPrompt(),
	}

	agent := NewReActAgent(config, mockLLM, mockSkillManager, nil)

	ctx := context.Background()
	input := &Input{
		Query: "Review the provided JavaScript function and suggest improvements.",
		Context: map[string]interface{}{
			"task_type": "programming",
			"language":  "javascript",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		responseIndex = 0 // 重置响应索引
		_, _ = agent.Run(ctx, input)
	}
}

// getSimpleThoughtPrompt 获取简化的提示词，用于基准测试
func getSimpleThoughtPrompt() string {
	return `你是一个智能助手，执行ReAct（推理+行动）循环以完成任务。
你的回答必须严格遵循以下JSON格式：
{
  "thought": "你对当前情况的思考",
  "action": "要执行的动作名称或 'finish' 表示完成",
  "action_args": {
    // 执行动作所需的参数
  }
}
`
}

package react

import (
	"context"
	"time"

	"github.com/insmtx/SingerOS/backend/llm"
	"github.com/insmtx/SingerOS/backend/skills"
	"github.com/insmtx/SingerOS/backend/types"
)

const (
	// DefaultProgrammerMaxIterations 默认程序员助手最大迭代次数
	DefaultProgrammerMaxIterations = 15
	// DefaultProgrammerTimeout 默认程序员助手超时时间
	DefaultProgrammerTimeout = 3 * time.Minute
	// ProgrammerAssistantName 程序员助手名称
	ProgrammerAssistantName = "CodeHelperAssistant"
)

// ProgrammingAgentConfig 程序员助手特殊配置
type ProgrammingAgentConfig struct {
	BaseConfig         *Config                 // 基础ReAct配置
	LLM                llm.Provider            // LLM提供程序
	SkillManager       skills.SkillManager     // 技能管理器
	DigitalAssistant   *types.DigitalAssistant // 数字助理实例
	ProgrammingContext map[string]interface{}  // 编程上下文，如编程语言偏好、框架选择等
}

// NewProgrammingAgent 创建一个程序员助手
func NewProgrammingAgent(config *ProgrammingAgentConfig) *ReActAgent {
	if config.BaseConfig == nil {
		// 设置默认配置
		config.BaseConfig = &Config{
			MaxIterations:  DefaultProgrammerMaxIterations,
			Timeout:        DefaultProgrammerTimeout,
			Model:          "gpt-4",
			ThroughtPrompt: getProgrammerThoughtPrompt(config.ProgrammingContext),
		}
	} else {
		// 如果某些值未设置，则使用默认值
		if config.BaseConfig.MaxIterations == 0 {
			config.BaseConfig.MaxIterations = DefaultProgrammerMaxIterations
		}
		if config.BaseConfig.Timeout == 0 {
			config.BaseConfig.Timeout = DefaultProgrammerTimeout
		}
		if config.BaseConfig.ThroughtPrompt == "" {
			config.BaseConfig.ThroughtPrompt = getProgrammerThoughtPrompt(config.ProgrammingContext)
		}
	}

	return NewReActAgent(
		config.BaseConfig,
		config.LLM,
		config.SkillManager,
		config.DigitalAssistant,
	)
}

// getProgrammerThoughtPrompt 获取程序员助手的思考提示词
func getProgrammerThoughtPrompt(progCtx map[string]interface{}) string {
	language := "Go"
	framework := "Gin"
	additionalContext := ""

	if progCtx != nil {
		if lang, ok := progCtx["language"].(string); ok && lang != "" {
			language = lang
		}
		if fw, ok := progCtx["framework"].(string); ok && fw != "" {
			framework = fw
		}
		if ctx, ok := progCtx["additional_context"].(string); ok {
			additionalContext = ctx
		}
	}

	return `你是一个高级编程助手，精通` + language + `和` + framework + `框架。` +
		additionalContext +
		`
你的目标是作为智能编程助手，帮助程序员解决问题。
你必须按照 ReAct 模型，即 Reasoning + Acting (推理 + 行动)，在每个步骤中：
1. 思考当前任务需要什么
2. 采取行动调用适当技能
3. 观察行动的结果
4. 依此循环直到任务完成

你的回答必须严格遵循以下JSON格式：
{
  "thought": "你对当前情况的思考，包括分析、计划、和下一部打算做什么",
  "action": "要执行的动作名称或 'finish' 表示完成",
  "action_args": {
    // 执行动作所需的参数
  }
}

重要限制和提示：
- 优先使用可用的技能完成任务
- 除非绝对必要，否则不要构造假的技能ID
- 智能选择技能以完成用户的请求
- 任务完成后始终使用 'finish' 作为 'action'

如果不确定技能ID，请调用 'skills.list_available' 查询可用技能。

可用编程相关技能：
- 'code.review': 代码审查，接收 code 参数
- 'code.generate': 代码生成，接收 requirement 参数  
- 'github.pr_analysis': GitHub PR分析，接收 repo, pr_number 参数
- 'github.issue_analysis': Issue分析，接收 repo, issue_number 参数
- 'git.diff_parse': 解析diff，接收 diff_contents 参数
- 'doc.search': 文档搜索，接收 query 参数
- 'skills.list_available': 获取可用技能列表`
}

// RunProgrammingTask 运行程序员助手任务
func (r *ReActAgent) RunProgrammingTask(ctx context.Context, input *Input) (*Output, error) {
	// 验证输入是否适合编程任务
	if input == nil {
		return nil, ErrInvalidInput
	}

	// 设置特定于编程任务的上下文
	if input.Context == nil {
		input.Context = make(map[string]interface{})
	}
	input.Context["task_type"] = "programming"
	input.Context["agent_type"] = "programmer-assistant"

	// 添加编程专用上下文信息
	input.Context["programming_skills_available"] = true
	input.Context["programming_language"] = "Go"

	return r.Run(ctx, input)
}

// PredefinedProgrammingSkills 返回预定义的编程技能列表
func PredefinedProgrammingSkills() []string {
	return []string{
		"code.review",
		"code.generate",
		"github.pr_analysis",
		"github.issue_analysis",
		"git.diff_parse",
		"doc.search",
		"repo.search",
		"code.explain",
		"bug.identify",
		"test.generate",
	}
}

// IsProgrammingRelated 检查输入是否与编程相关
func IsProgrammingRelated(query string) bool {
	programmingKeywords := []string{
		"code", "program", "function", "method", "class", "struct",
		"variable", "algorithm", "logic", "bug", "error", "fix",
		"review", "optimize", "refactor", "implement", "develop",
		"programming", "software", "application", "script", "module",
		"library", "framework", "api", "endpoint", "route",
		"go", "golang", "javascript", "java", "python", "c++", "rust",
		"pr", "pull request", "commit", "branch", "merge", "git",
		"test", "unittest", "integration test", "build", "deploy",
		"architecture", "design pattern", "database", "sql", "query",
	}

	for _, keyword := range programmingKeywords {
		if isSubstring(query, keyword) {
			return true
		}
	}

	return false
}

var (
	// ErrInvalidInput 输入无效错误
	ErrInvalidInput = &ReActError{"invalid input for programming agent"}
)

// ReActError ReAct相关错误类型
type ReActError struct {
	msg string
}

func (e *ReActError) Error() string {
	return e.msg
}

// isSubstring 检查target是否包含in字符串
func isSubstring(target, in string) bool {
	if len(in) == 0 {
		return true
	}
	if len(target) < len(in) {
		return false
	}
	for i := 0; i <= len(target)-len(in); i++ {
		match := true
		for j := 0; j < len(in); j++ {
			if target[i+j] != in[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

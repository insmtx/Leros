package react

import (
	"context"
	"time"

	"github.com/insmtx/SingerOS/backend/llm"
	"github.com/insmtx/SingerOS/backend/skills"
	"github.com/insmtx/SingerOS/backend/types"
)

// State 定义ReAct代理的状态
type State struct {
	Thought    string                 `json:"thought"`         // 思考内容
	Action     string                 `json:"action"`          // 行动名称
	ActionArgs map[string]interface{} `json:"action_args"`     // 行动参数
	Observed   string                 `json:"observed"`        // 观察结果
	Completed  bool                   `json:"completed"`       // 是否完成
	Error      string                 `json:"error,omitempty"` // 错误信息
	Iteration  int                    `json:"iteration"`       // 当前迭代次数
	CreatedAt  time.Time              `json:"created_at"`      // 创建时间
}

// Input 定义ReAct代理的输入
type Input struct {
	Query       string                 `json:"query"`                  // 用户查询
	Context     map[string]interface{} `json:"context,omitempty"`      // 上下文信息
	SessionID   string                 `json:"session_id"`             // 会话ID
	UserID      string                 `json:"user_id"`                // 用户ID
	ExtraParams map[string]interface{} `json:"extra_params,omitempty"` // 附加参数
}

// Output 定义ReAct代理的输出
type Output struct {
	Result     string    `json:"result"`          // 最终结果
	Status     string    `json:"status"`          // 运行状态 (success,error,timeout)
	States     []State   `json:"states"`          // 整个执行过程的所有状态
	Error      string    `json:"error,omitempty"` // 错误信息
	StartTime  time.Time `json:"start_time"`      // 开始时间
	EndTime    time.Time `json:"end_time"`        // 结束时间
	TotalSteps int       `json:"total_steps"`     // 总步数
}

// Config 定义ReAct代理的配置
type Config struct {
	MaxIterations  int           `json:"max_iterations"`  // 最大迭代次数
	Timeout        time.Duration `json:"timeout"`         // 执行超时时间
	ThroughtPrompt string        `json:"thought_prompt"`  // 思考提示词模板
	Model          string        `json:"model"`           // 使用的LLM模型
	SkillWhitelist []string      `json:"skill_whitelist"` // 允许的技能白名单
}

// ReActAgent ReAct代理实现
type ReActAgent struct {
	config           *Config
	llmProvider      llm.Provider
	skillManager     skills.SkillManager
	digitalAssistant *types.DigitalAssistant
}

// ReActAgentStateMachine 表示ReAct代理的状态机
type ReActAgentStateMachine struct {
	agent  *ReActAgent
	states []State
	ctx    context.Context
	cancel context.CancelFunc
	input  *Input
	output *Output
}

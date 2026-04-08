package react

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/SingerOS/backend/llm"
	"github.com/insmtx/SingerOS/backend/skills"
	"github.com/insmtx/SingerOS/backend/types"
)

const (
	// MaxStateHistorySize 设置最大历史状态数量，防内存溢出
	MaxStateHistorySize = 50
	// DefaultMaxIterations 默认最大迭代次数
	DefaultMaxIterations = 10
	// DefaultTimeout 默认超时时间
	DefaultTimeout = 5 * time.Minute
)

var (
	ErrMaxIterationsExceeded = errors.New("max iterations exceeded")
	ErrTimeout               = errors.New("execution timed out")
	ErrAgentStopped          = errors.New("agent was stopped")
)

// NewReActAgent 创建一个新的ReAct代理实例
func NewReActAgent(
	config *Config,
	llmProvider llm.Provider,
	skillManager skills.SkillManager,
	digitalAssistant *types.DigitalAssistant,
) *ReActAgent {
	if config.MaxIterations == 0 {
		config.MaxIterations = DefaultMaxIterations
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}
	if config.ThroughtPrompt == "" {
		config.ThroughtPrompt = defaultThoughtPrompt
	}

	agent := &ReActAgent{
		config:           config,
		llmProvider:      llmProvider,
		skillManager:     skillManager,
		digitalAssistant: digitalAssistant,
	}
	return agent
}

// defaultThoughtPrompt 默认的思考提示词模板
var defaultThoughtPrompt = `
根据当前任务和上下文，执行ReAct（推理+行动）循环。
你的回答必须严格遵循以下JSON格式：
{
  "thought": "你对该步骤的思考过程",
  "action": "要执行的动作名称或 'finish' 表示完成",
  "action_args": {
    // 执行动作所需的参数
  }
}

如果任务已完成，将'action'设置为'finish'。
`

// Run 运行ReAct代理，处理输入并返回输出
func (r *ReActAgent) Run(ctx context.Context, input *Input) (*Output, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	stateMachine := &ReActAgentStateMachine{
		agent:  r,
		states: make([]State, 0),
		ctx:    ctx,
		cancel: cancel,
		input:  input,
		output: &Output{
			StartTime: time.Now(),
			States:    make([]State, 0),
		},
	}

	return stateMachine.Run()
}

// Run 运行状态机直到完成
func (sm *ReActAgentStateMachine) Run() (*Output, error) {
	currentInput := map[string]interface{}{
		"query":      sm.input.Query,
		"session_id": sm.input.SessionID,
		"context":    sm.input.Context,
		"timestamp":  time.Now().Unix(),
	}

	for i := 0; i < sm.agent.config.MaxIterations; i++ {
		select {
		case <-sm.ctx.Done():
			return sm.createOutput("error", "context cancelled or timeout"), fmt.Errorf("context cancelled: %w", sm.ctx.Err())
		default:
		}

		// 更新迭代次数
		currentState := State{
			Iteration: i,
			CreatedAt: time.Now(),
		}

		// 使用LLM生成思考和行动
		thought, action, actionArgs, err := sm.agent.generateAction(sm.ctx, currentInput, sm.states)
		if err != nil {
			currentState.Error = err.Error()
			currentState.Completed = true
			sm.states = append(sm.states, currentState)
			return sm.createOutput("error", err.Error()), err
		}

		currentState.Thought = thought
		currentState.Action = action
		currentState.ActionArgs = actionArgs

		// 检查是否完成
		if action == "finish" {
			currentState.Observed = "Task completed by finishing action"
			currentState.Completed = true
			sm.states = append(sm.states, currentState)
			return sm.createOutput("success", "task completed successfully"), nil
		}

		// 执行技能
		result, err := sm.executeAction(currentState.Action, currentState.ActionArgs)
		if err != nil {
			currentState.Observed = fmt.Sprintf("Action error: %v", err)
			currentState.Error = err.Error()
		} else {
			currentState.Observed = fmt.Sprintf("Action successful: %+v", result)

			// 更新上下文供下一步使用
			for k, v := range result {
				currentInput[k] = v
			}
		}

		// 添加当前状态到历史
		sm.states = append(sm.states, currentState)

		// 检查是否超出大小限制 - 使用滑动窗口保留最新的状态
		if len(sm.states) > MaxStateHistorySize {
			// 保留最新的MaxStateHistorySize个状态
			sm.states = sm.states[len(sm.states)-MaxStateHistorySize:]
		}
	}

	// 达到最大迭代次数
	return sm.createOutput("error", "max iterations exceeded"), ErrMaxIterationsExceeded
}

// generateAction 使用LLM生成思考和行动
func (r *ReActAgent) generateAction(ctx context.Context, contextInput map[string]interface{}, previousStates []State) (thought string, action string, actionArgs map[string]interface{}, err error) {
	// 构建消息
	messages := []llm.Message{
		{
			Role:    "system",
			Content: r.config.ThroughtPrompt,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(`
Current input: %+v
Previous states: %s

Please respond with a JSON object like:
{
  "thought": "your thought process...",
  "action": "action name or 'finish'",
  "action_args": {...}
}
`,
				contextInput,
				r.formatPreviousStates(previousStates)),
		},
	}

	req := &llm.GenerateRequest{
		Messages:    messages,
		Model:       r.config.Model,
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	resp, err := r.llmProvider.Generate(ctx, req)
	if err != nil {
		return "", "", nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// 解析响应
	var actionResult struct {
		Thought    string                 `json:"thought"`
		Action     string                 `json:"action"`
		ActionArgs map[string]interface{} `json:"action_args"`
	}

	// 尝试解析JSON响应
	content := strings.TrimSpace(resp.Content)
	if json.Valid([]byte(content)) {
		err = json.Unmarshal([]byte(content), &actionResult)
		if err != nil {
			// 如果直接解析失败，尝试提取其中的JSON块
			found, thoughtExtract, actionExtract, argsExtract := extractJSONFromText(content)
			if !found {
				return "", "", nil, fmt.Errorf("failed to parse LLM response as JSON: %w. Response: %s", err, content)
			}
			return thoughtExtract, actionExtract, argsExtract, nil
		}
	} else {
		// 如果不是有效JSON，则尝试查找可能的JSON块
		found, thoughtExtract, actionExtract, argsExtract := extractJSONFromText(content)
		if !found {
			return "", "", nil, fmt.Errorf("failed to extract action from LLM response: %s", content)
		}
		return thoughtExtract, actionExtract, argsExtract, nil
	}

	return actionResult.Thought, actionResult.Action, actionResult.ActionArgs, nil
}

// formatPreviousStates 格式化先前的状态作为上下文
func (r *ReActAgent) formatPreviousStates(states []State) string {
	if states == nil || len(states) == 0 {
		return "[]"
	}

	// 为了防止上下文过长，只取最近有限数量的状态
	startIdx := 0
	if len(states) > 5 {
		startIdx = len(states) - 5
	}

	result := make([]State, 0, len(states)-startIdx)
	for i := startIdx; i < len(states); i++ {
		result = append(result, states[i])
	}

	b, _ := json.Marshal(result)
	return string(b)
}

// executeAction 执行指定的动作/技能
func (sm *ReActAgentStateMachine) executeAction(action string, args map[string]interface{}) (map[string]interface{}, error) {
	// 检查是否在技能白名单中
	if len(sm.agent.config.SkillWhitelist) > 0 {
		found := false
		for _, allowedSkill := range sm.agent.config.SkillWhitelist {
			if action == allowedSkill {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("action %s not in whitelist", action)
		}
	}

	// 尝试从技能管理器执行技能
	result, err := sm.agent.skillManager.Execute(sm.ctx, action, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute skill %s: %w", action, err)
	}

	return result, nil
}

// createOutput 创建输出对象
func (sm *ReActAgentStateMachine) createOutput(status, message string) *Output {
	output := sm.output
	output.Status = status
	output.States = sm.states
	output.EndTime = time.Now()
	output.TotalSteps = len(sm.states)

	if lastState := sm.getLastState(); lastState != nil && lastState.Observed != "" {
		output.Result = lastState.Observed
	} else {
		output.Result = message
	}

	return output
}

// getLastState 获取最后的状态
func (sm *ReActAgentStateMachine) getLastState() *State {
	if len(sm.states) == 0 {
		return nil
	}
	return &sm.states[len(sm.states)-1]
}

// extractJSONFromText 从文本中提取JSON块
func extractJSONFromText(text string) (bool, string, string, map[string]interface{}) {
	// 尝试找到JSON块
	if idx := strings.Index(text, "{"); idx != -1 {
		candidates := []string{text[idx:]}
		// 检查文本多个位置以寻找可能的JSON结构
		for i, ch := range text {
			if ch == '{' && i > 0 {
				candidates = append(candidates, text[i:])
			}
		}

		for _, candidate := range candidates {
			end := strings.LastIndex(candidate, "}")
			if end != -1 {
				jsonStr := candidate[:end+1]

				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
					// 确保必需字段存在
					thought, exists1 := parsed["thought"].(string)
					action, exists2 := parsed["action"].(string)

					args, ok := parsed["action_args"].(map[string]interface{})
					if !ok {
						// 尝试其他可能的格式
						argsIntf, exists := parsed["action_args"]
						if exists && argsIntf != nil {
							if argsBytes, err := json.Marshal(argsIntf); err == nil {
								args = make(map[string]interface{})
								json.Unmarshal(argsBytes, &args)
							} else {
								args = make(map[string]interface{})
							}
						} else {
							args = make(map[string]interface{})
						}
					}

					if exists1 && exists2 && thought != "" && action != "" {
						return true, thought, action, args
					}
				}
			}
		}
	}

	return false, "", "", nil
}

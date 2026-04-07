package react

import (
	"context"

	"github.com/insmtx/SingerOS/backend/interaction"
	"github.com/insmtx/SingerOS/backend/llm"
	skills "github.com/insmtx/SingerOS/backend/skills"
)

// AgentOrchestrator 封装了与Orchestrator的集成逻辑
type AgentOrchestrator struct {
	reactAgent       *ReActAgent
	programmingAgent *ReActAgent // 程序员助手代理
	skillManager     skills.SkillManager
}

// NewAgentOrchestrator 创建一个新的与Orchestrator集成的实例
func NewAgentOrchestrator(llmProvider llm.Provider, skillManager skills.SkillManager) *AgentOrchestrator {
	// 创建基础ReAct代理
	reactConfig := &Config{
		MaxIterations: DefaultMaxIterations,
		Timeout:       DefaultTimeout,
		Model:         "gpt-4",
	}

	// 创建程序员助手代理
	progConfig := &ProgrammingAgentConfig{
		BaseConfig: &Config{
			MaxIterations:  DefaultProgrammerMaxIterations,
			Timeout:        DefaultProgrammerTimeout,
			Model:          "gpt-4",
			ThroughtPrompt: getDefaultProgrammingPrompt(),
		},
		LLM:              llmProvider,
		SkillManager:     skillManager,
		DigitalAssistant: nil, // 在实际环境中应该设置合适的DigitalAssistant
		ProgrammingContext: map[string]interface{}{
			"language":           "Go",
			"framework":          "SingerOS",
			"additional_context": "帮助分析SingerOS代码库的PR和issues",
		},
	}

	return &AgentOrchestrator{
		reactAgent:       NewReActAgent(reactConfig, llmProvider, skillManager, nil),
		programmingAgent: NewProgrammingAgent(progConfig),
		skillManager:     skillManager,
	}
}

// HandleEvent 将事件输入转换为适合ReAct代理的输入并执行
func (ao *AgentOrchestrator) HandleEvent(ctx context.Context, event *interaction.Event) error {
	// 首先检查事件是否与编程相关，如果是，使用程序员助手
	query := buildQueryFromEvent(event)

	input := &Input{
		Query:       query,
		SessionID:   event.EventID,
		Context:     buildContextFromEvent(event),
		UserID:      event.Actor,
		ExtraParams: map[string]interface{}{"event_type": event.EventType},
	}

	// 智能路由：根据事件内容决定使用哪个代理
	if IsProgrammingRelated(query) {
		_, err := ao.programmingAgent.RunProgrammingTask(ctx, input)
		return err
	} else {
		_, err := ao.reactAgent.Run(ctx, input)
		return err
	}
}

// HandleEventAdvanced 使用更复杂的逻辑处理事件，可能涉及多轮代理交互
func (ao *AgentOrchestrator) HandleEventAdvanced(ctx context.Context, event *interaction.Event) error {
	query := buildQueryFromEvent(event)

	// 构建高级输入
	input := &Input{
		Query:       query,
		SessionID:   event.EventID + "-advanced",
		Context:     buildAdvancedContextFromEvent(event),
		UserID:      event.Actor,
		ExtraParams: map[string]interface{}{"event_type": event.EventType, "advanced_processing": true},
	}

	// 尝试根据事件类型和内容路由到适当的处理流程
	switch event.EventType {
	case "github.issue_comment", "github.pr_comment":
		// 如果是注释并且看起来像是代码相关问题，使用程序员助手
		if IsProgrammingRelated(query) {
			_, err := ao.programmingAgent.RunProgrammingTask(ctx, input)
			return err
		}
	case "github.pull_request.opened", "github.pull_request.synchronize":
		// 对PR事件使用程序员助手进行分析
		_, err := ao.programmingAgent.RunProgrammingTask(ctx, input)
		return err
	default:
		// 使用基本的ReAct代理
		_, err := ao.reactAgent.Run(ctx, input)
		return err
	}

	// 如果上面没有特别处理，根据内容判断
	if IsProgrammingRelated(query) {
		_, err := ao.programmingAgent.RunProgrammingTask(ctx, input)
		return err
	} else {
		_, err := ao.reactAgent.Run(ctx, input)
		return err
	}
}

// buildQueryFromEvent 从事件构建查询字符串
func buildQueryFromEvent(event *interaction.Event) string {
	queryParts := []string{}

	// 添加事件基本信息
	if event.EventType != "" {
		queryParts = append(queryParts, "Event Type: "+event.EventType)
	}
	if event.Actor != "" {
		queryParts = append(queryParts, "Actor: "+event.Actor)
	}
	if event.Repository != "" {
		queryParts = append(queryParts, "Repository: "+event.Repository)
	}

	// 从 Payload 中提取关键信息
	if event.Payload != nil {
		payload, ok := event.Payload.(map[string]interface{})
		if ok {
			// 尝试提取标题（取决于事件类型）
			if title, exists := payload["title"]; exists && title != nil {
				if titleStr, ok := title.(string); ok && titleStr != "" {
					queryParts = append(queryParts, titleStr)
				}
			}

			// 尝试提取描述
			if desc, exists := payload["description"]; exists && desc != nil {
				if descStr, ok := desc.(string); ok && descStr != "" {
					queryParts = append(queryParts, descStr)
				}
			}

			// 尝试提取评论内容
			if comment, exists := payload["comment"]; exists && comment != nil {
				if commentStr, ok := comment.(string); ok && commentStr != "" {
					queryParts = append(queryParts, commentStr)
				}
			}

			// 从PR或其他对象中提取标题
			if pullRequest, exists := payload["pull_request"]; exists && pullRequest != nil {
				if prObj, ok := pullRequest.(map[string]interface{}); ok {
					if prTitle, exists := prObj["title"]; exists && prTitle != nil {
						if prTitleStr, ok := prTitle.(string); ok && prTitleStr != "" {
							queryParts = append(queryParts, prTitleStr)
						}
					}
				}
			}
		}
	}

	// 组合所有部分
	query := ""
	for i, part := range queryParts {
		if i > 0 {
			query += ". "
		}
		query += part
	}

	return query
}

// buildContextFromEvent 从事件构建上下文
func buildContextFromEvent(event *interaction.Event) map[string]interface{} {
	context := map[string]interface{}{}

	// 添加一些标准上下文信息
	if event.EventType != "" {
		context["event_type"] = event.EventType
	}
	if event.Actor != "" {
		context["actor"] = event.Actor
	}
	if event.Repository != "" {
		context["repository"] = event.Repository
	}
	if event.EventType == "github.pull_request" {
		context["event_subtype"] = "github.pull_request"
	}

	return context
}

// buildAdvancedContextFromEvent 构建更高级的上下文信息
func buildAdvancedContextFromEvent(event *interaction.Event) map[string]interface{} {
	context := buildContextFromEvent(event)

	// 添加事件载荷原始信息（如果安全的话）
	payload, ok := event.Payload.(map[string]interface{})
	if ok {
		// 添加PR或Issue编号
		if number, exists := payload["number"]; exists {
			context["item_number"] = number
		}

		// 添加原始标题
		if title, exists := payload["title"]; exists {
			context["original_title"] = title
		}

		// 添加创建时间信息
		if created, exists := payload["created_at"]; exists {
			context["created_at"] = created
		}

		// 如果有diff或变更信息，也可以加入（但在这里我们需要小心数据量）
		context["has_diff"] = payload["diff"] != nil
		// 不直接添加diff内容以免过大
	}

	return context
}

// GetAgentStats 获取代理统计信息
func (ao *AgentOrchestrator) GetAgentStats() map[string]interface{} {
	stats := map[string]interface{}{}

	if ao.programmingAgent != nil {
		stats["programming_agent"] = map[string]interface{}{
			"type": "ReActAgent",
			"name": ProgrammerAssistantName,
		}
	}

	if ao.reactAgent != nil {
		stats["base_react_agent"] = map[string]interface{}{
			"type": "ReActAgent",
		}
	}

	stats["registered_skills_count"] = countRegisteredSkills(ao.skillManager)

	return stats
}

// countRegisteredSkills 计算已注册技能的数量
func countRegisteredSkills(sm skills.SkillManager) int {
	skillsList := sm.List()
	return len(skillsList)
}

// getDefaultProgrammingPrompt 获取默认的编程助手提示词
func getDefaultProgrammingPrompt() string {
	return `你是一个高级编程助手，专门帮助开发人员分析代码、PR、以及技术问题。
你应该使用ReAct方法，即Reasoning + Acting (推理 + 行动)，在每次响应中：

1. 首先思考你当前面对的情况和用户请求
2. 然后决定是否需要执行特定工具
3. 如果需要，就调用合适的技能

你的响应必须严格遵循以下JSON格式：
{
  "thought": "你内心的想法和计划",
  "action": "要调用的技能或 'finish' 表示完成",
  "action_args": {
    "param1": "value1"
  }
}

可用技能：
- 'code.review': 代码审查
- 'github.pr_analysis': 分析GitHub PR
- 'skills.list_available': 列出可用技能`
}

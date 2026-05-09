package lifecycle

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/SingerOS/backend/internal/agent"
	agentevents "github.com/insmtx/SingerOS/backend/internal/agent/events"
)

const (
	toolNameMemory      = "memory"
	toolNameSkillManage = "skill_manage"
)

const learningCheckPrompt = `[系统学习检查]
这次任务已经完成。请判断是否有长期价值信息需要保存：
- 用户偏好、稳定事实、项目约定、工具坑点：调用 memory
- 可复用的多步骤流程，或使用某个技能时发现缺失步骤：调用 skill_manage
- 临时进度、一次性结果、普通日志：不要保存
如果没有值得保存的内容，直接回复“无需保存”。`

const memoryFlushPrompt = `[系统记忆保存]
当前会话即将被压缩或重置。请只保存未来仍有价值的信息：
- 用户偏好和纠正
- 环境事实和项目约定
- 工具坑点和稳定工作流
不要保存当前任务进度、临时日志或一次性结果。
如需保存请调用 memory；否则回复“无需保存”。`

// AfterRunLearning 在主任务结束后触发一轮轻量自我学习检查。
func (r *Runner) AfterRunLearning(ctx context.Context, req *agent.RequestContext, result *agent.RunResult, trace *RunTrace) error {
	if r == nil || r.builder == nil || r.delegate == nil || req == nil || result == nil || !shouldRunLearningCheck(req, result, trace) {
		return nil
	}

	allowedTools := availableToolNames(r.toolAvailability, []string{toolNameMemory, toolNameSkillManage})
	if len(allowedTools) == 0 {
		return nil
	}

	learningReq := cloneRequest(req)
	learningReq.Input = agent.InputContext{
		Type: agent.InputTypeTaskInstruction,
		Text: buildLearningPrompt(req, result, trace),
	}
	learningReq.Capability.AllowedTools = allowedTools
	learningReq.Runtime.MaxStep = 3
	learningReq.EventSink = agentevents.NewNoopSink()

	next, err := r.builder.Prepare(ctx, learningReq)
	if err != nil {
		return err
	}
	next.EventSink = agentevents.NewNoopSink()
	_, err = r.delegate.Run(ctx, next)
	return err
}

// BeforeCompact 在未来会话压缩前触发长期记忆保存。
func (r *Runner) BeforeCompact(ctx context.Context, req *agent.RequestContext) error {
	return r.runMemoryFlush(ctx, req, "compact")
}

// BeforeReset 在未来会话重置前触发长期记忆保存。
func (r *Runner) BeforeReset(ctx context.Context, req *agent.RequestContext) error {
	return r.runMemoryFlush(ctx, req, "reset")
}

func (r *Runner) runMemoryFlush(ctx context.Context, req *agent.RequestContext, reason string) error {
	if r == nil || r.builder == nil || r.delegate == nil || req == nil {
		return nil
	}
	allowedTools := availableToolNames(r.toolAvailability, []string{toolNameMemory})
	if len(allowedTools) == 0 {
		return nil
	}

	flushReq := cloneRequest(req)
	flushReq.Input = agent.InputContext{
		Type: agent.InputTypeTaskInstruction,
		Text: memoryFlushPrompt + "\n\n触发原因：" + strings.TrimSpace(reason),
	}
	flushReq.Capability.AllowedTools = allowedTools
	flushReq.Runtime.MaxStep = 2
	flushReq.EventSink = agentevents.NewNoopSink()

	prepared, err := r.builder.Prepare(ctx, flushReq)
	if err != nil {
		return err
	}
	prepared.EventSink = agentevents.NewNoopSink()
	_, err = r.delegate.Run(ctx, prepared)
	return err
}

func availableToolNames(availability ToolAvailability, names []string) []string {
	if availability == nil {
		return nil
	}
	return availability.AvailableToolNames(names)
}

func shouldRunLearningCheck(req *agent.RequestContext, result *agent.RunResult, trace *RunTrace) bool {
	if req == nil || result == nil || result.Status != agent.RunStatusCompleted {
		return false
	}
	if trace == nil {
		trace = &RunTrace{}
	}
	if alreadyCalledLearningTool(trace.ToolNames) {
		return false
	}
	if containsLearningCue(buildUserInput(req)) {
		return true
	}
	if trace.ToolFailures > 0 {
		return true
	}
	if trace.ToolCalls >= 5 {
		return true
	}
	if trace.UsedSkillTool && trace.ToolCalls >= 3 {
		return true
	}
	return false
}

func buildLearningPrompt(req *agent.RequestContext, result *agent.RunResult, trace *RunTrace) string {
	if trace == nil {
		trace = &RunTrace{}
	}
	var builder strings.Builder
	builder.WriteString(learningCheckPrompt)
	builder.WriteString("\n\n运行摘要：")
	if req != nil {
		if req.Input.Type != "" {
			builder.WriteString("\n- input_type: ")
			builder.WriteString(string(req.Input.Type))
		}
		if req.Actor.UserID != "" {
			builder.WriteString("\n- actor: ")
			builder.WriteString(req.Actor.UserID)
		}
	}
	builder.WriteString(fmt.Sprintf("\n- status: %s", result.Status))
	builder.WriteString(fmt.Sprintf("\n- tool_calls: %d", trace.ToolCalls))
	builder.WriteString(fmt.Sprintf("\n- tool_failures: %d", trace.ToolFailures))
	if len(trace.ToolNames) > 0 {
		builder.WriteString("\n- tools: ")
		builder.WriteString(strings.Join(uniqueStrings(trace.ToolNames), ", "))
	}
	if strings.TrimSpace(result.Message) != "" {
		builder.WriteString("\n- final_answer: ")
		builder.WriteString(truncateForPrompt(result.Message, 1200))
	}
	return builder.String()
}

func learningGuidance() string {
	return `## 自我学习规则
- 当用户明确纠正你、表达偏好或要求“记住”时，优先使用 memory 保存长期事实。
- 当你完成复杂任务、修复流程缺口或发现可复用步骤时，可以使用 skill_manage 更新流程性记忆。
- 不要保存临时任务进度、普通日志、一次性结果或容易重新发现的信息。`
}

func alreadyCalledLearningTool(names []string) bool {
	for _, name := range names {
		switch name {
		case toolNameMemory, toolNameSkillManage:
			return true
		}
	}
	return false
}

func containsLearningCue(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return false
	}
	cues := []string{
		"记住", "记一下", "以后", "下次", "不要再", "别再", "偏好", "习惯", "规范",
		"remember", "next time", "preference", "don't do that again", "do not do that again",
	}
	for _, cue := range cues {
		if strings.Contains(text, cue) {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

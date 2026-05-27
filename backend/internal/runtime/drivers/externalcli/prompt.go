package externalcli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/insmtx/Leros/backend/internal/agent"
)

func buildPrompt(req *agent.RequestContext) string {
	if req == nil {
		return ""
	}

	sections := []string{"# Runtime Context"}

	// TODO 角色定义上下文
	// if req.Assistant.ID != "" || req.Assistant.Name != "" || req.Assistant.Role != "" || req.Assistant.SystemPrompt != "" {
	// 	sections = append(sections, formatJSONSection("Assistant", req.Assistant))
	// }
	// if req.Actor.UserID != "" || req.Actor.Channel != "" || req.Actor.ExternalID != "" || req.Actor.AccountID != "" {
	// 	sections = append(sections, formatJSONSection("Actor", req.Actor))
	// }
	if req.Conversation.ID != "" || len(req.Conversation.Messages) > 0 {
		sections = append(sections, formatJSONSection("Conversation Context", req.Conversation))
	}
	sections = append(sections, formatCurrentUserTaskSection(req.Input))

	sections = append(sections, `## Output Contract
- 使用中文输出最终结果。
- 不要编造未实际执行的命令、文件、链接、ID 或状态。
- 如果需要执行真实环境操作，请使用 runtime 已配置的工具或 MCP 能力。
- 如果本轮创建了需要用户查看、下载或复用的文件，请在文件创建完成后调用 artifact_declare 声明最终产物，path 使用该文件的完整路径。
- 最终回复只说明完成结果和文件名，不要输出本地绝对路径或下载链接。`)

	return strings.Join(sections, "\n\n")
}

func formatCurrentUserTaskSection(input agent.InputContext) string {
	return fmt.Sprintf("## Current User Task\n\n%s", currentUserTaskText(input))
}

func currentUserTaskText(input agent.InputContext) string {
	if text := strings.TrimSpace(input.Text); text != "" {
		return text
	}
	if len(input.Messages) > 0 {
		lines := make([]string, 0, len(input.Messages))
		for _, message := range input.Messages {
			content := strings.TrimSpace(message.Content)
			if content == "" {
				continue
			}
			if role := strings.TrimSpace(message.Role); role != "" {
				lines = append(lines, fmt.Sprintf("%s: %s", role, content))
				continue
			}
			lines = append(lines, content)
		}
		return strings.Join(lines, "\n")
	}
	return string(input.Type)
}

func formatJSONSection(title string, value any) string {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("## %s\n%v", title, value)
	}
	return fmt.Sprintf("## %s\n```json\n%s\n```", title, string(encoded))
}

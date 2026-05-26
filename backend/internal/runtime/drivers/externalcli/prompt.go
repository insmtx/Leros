package externalcli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/insmtx/Leros/backend/internal/agent"
	agentworkspace "github.com/insmtx/Leros/backend/internal/workspace"
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
	if outputContract := formatWorkspaceOutputContract(req); outputContract != "" {
		sections = append(sections, outputContract)
	}
	// if req.Policy.RequireApproval {
	// 	sections = append(sections, formatJSONSection("Policy", req.Policy))
	// }

	sections = append(sections, `## Output Contract
- 使用中文输出最终结果。
- 不要编造未实际执行的命令、文件、链接、ID 或状态。
- 如果需要执行真实环境操作，请使用 runtime 已配置的工具或 MCP 能力。`)

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

func formatWorkspaceOutputContract(req *agent.RequestContext) string {
	_, ok, err := agentworkspace.FromAgentRequest(req)
	if err != nil || !ok {
		return ""
	}
	return `## File Output Contract

- 你在受控项目工作区内执行任务。只能在当前项目仓库内读取和写入文件，不要访问仓库外路径。
- 不要在回复中暴露本地绝对路径、容器路径或 sandbox 路径。
- 如果本轮生成了需要交付给用户下载、查看或复用的文件，必须将其声明为最终产物。
- 最终产物必须写入 LEROS_ARTIFACT_FILE 指向的 manifest 文件。
- manifest 使用 JSON Lines，每行一个对象：
  {"path":"相对项目仓库的文件路径","title":"展示名称","description":"简短说明","mime_type":"可选 MIME 类型","artifact_type":"file","is_final":true}
- path 必须是相对项目仓库的路径；如果当前在子目录，请包含子目录前缀。
- 禁止声明绝对路径、..、不存在文件、目录、临时文件、日志文件和缓存文件。
- 最终交付文件不要写入临时目录、日志目录或缓存目录；这些目录只用于过程文件。
- 只有 is_final=true 的声明会展示给用户。
- 不要在最终回复中生成下载链接，系统会根据 artifact 声明自动提供下载入口。
- 最终回复只需要说明完成结果；如有文件，可提到文件名，但不要输出真实路径。`
}

func formatJSONSection(title string, value any) string {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("## %s\n%v", title, value)
	}
	return fmt.Sprintf("## %s\n```json\n%s\n```", title, string(encoded))
}

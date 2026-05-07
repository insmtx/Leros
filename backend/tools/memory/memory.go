// Package memory exposes SingerOS built-in memory as a runtime tool.
package memory

import (
	"context"
	"fmt"
	"strings"

	localmemory "github.com/insmtx/SingerOS/backend/internal/memory/local"
	"github.com/insmtx/SingerOS/backend/tools"
)

const (
	// ToolNameMemory is the stable runtime tool name for built-in memory.
	ToolNameMemory = "memory"
)

// Tool lets the agent manage built-in USER.md and MEMORY.md files.
type Tool struct {
	tools.BaseTool
	store *localmemory.Store
}

// NewTool creates the built-in memory tool with the default local store.
func NewTool() *Tool {
	store, _ := localmemory.NewStore(localmemory.Options{})
	return NewToolWithStore(store)
}

// NewToolWithStore creates the built-in memory tool with an explicit store.
func NewToolWithStore(store *localmemory.Store) *Tool {
	return &Tool{
		BaseTool: tools.NewBaseTool(
			ToolNameMemory,
			"管理内置长期记忆。用于保存会跨会话复用的用户偏好、稳定事实、环境信息和经验。"+
				"支持 action=add/replace/remove，target 只能是 user 或 memory。"+
				"user 用于用户偏好、身份和沟通风格；memory 用于 worker 学到的环境事实、项目约定、工具坑点和长期经验。"+
				"不要保存临时任务进度、一次性 TODO、原始日志或大段数据。",
			tools.Schema{
				Type:     "object",
				Required: []string{"action", "target"},
				Properties: map[string]*tools.Property{
					"action": {
						Type:        "string",
						Description: "操作类型：add 新增记忆，replace 替换已有记忆，remove 删除已有记忆。",
						Enum:        []string{"add", "replace", "remove"},
					},
					"target": {
						Type:        "string",
						Description: "记忆目标：user 表示用户画像；memory 表示 worker/assistant 的长期事实和经验。",
						Enum:        []string{"user", "memory"},
					},
					"content": {
						Type:        "string",
						Description: "add/replace 使用的新记忆内容。应简洁、稳定、有长期价值。",
					},
					"old_text": {
						Type:        "string",
						Description: "replace/remove 用于定位已有条目的唯一短文本片段。",
					},
				},
			},
		),
		store: store,
	}
}

// Validate checks memory tool input before execution.
func (t *Tool) Validate(input map[string]interface{}) error {
	action := strings.TrimSpace(stringValue(input, "action"))
	target := strings.TrimSpace(stringValue(input, "target"))
	if action == "" {
		return fmt.Errorf("action is required")
	}
	if target == "" {
		return fmt.Errorf("target is required")
	}
	if target != localmemory.TargetUser && target != localmemory.TargetMemory {
		return fmt.Errorf("invalid target %q: use user or memory", target)
	}

	switch action {
	case "add":
		if strings.TrimSpace(stringValue(input, "content")) == "" {
			return fmt.Errorf("content is required for add")
		}
	case "replace":
		if strings.TrimSpace(stringValue(input, "old_text")) == "" {
			return fmt.Errorf("old_text is required for replace")
		}
		if strings.TrimSpace(stringValue(input, "content")) == "" {
			return fmt.Errorf("content is required for replace")
		}
	case "remove":
		if strings.TrimSpace(stringValue(input, "old_text")) == "" {
			return fmt.Errorf("old_text is required for remove")
		}
	default:
		return fmt.Errorf("unknown action %q: use add, replace, or remove", action)
	}

	return nil
}

// Execute performs the memory operation.
func (t *Tool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	if t == nil || t.store == nil {
		return "", fmt.Errorf("memory store is not initialized")
	}
	if err := t.Validate(input); err != nil {
		return "", err
	}

	action := strings.TrimSpace(stringValue(input, "action"))
	target := strings.TrimSpace(stringValue(input, "target"))
	content := stringValue(input, "content")
	oldText := stringValue(input, "old_text")

	var result *localmemory.Result
	var err error
	switch action {
	case "add":
		result, err = t.store.Add(ctx, target, content)
	case "replace":
		result, err = t.store.Replace(ctx, target, oldText, content)
	case "remove":
		result, err = t.store.Remove(ctx, target, oldText)
	default:
		return "", fmt.Errorf("unknown action %q", action)
	}
	if err != nil {
		return "", err
	}
	return tools.JSONString(result)
}

func stringValue(input map[string]interface{}, key string) string {
	if input == nil {
		return ""
	}
	value, ok := input[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprintf("%v", typed)
	}
}

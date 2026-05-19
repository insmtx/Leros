// Package skilluse provides the runtime tool for loading Leros skill documents.
package skilluse

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	skillcatalog "github.com/insmtx/Leros/backend/internal/skill/catalog"
	"github.com/insmtx/Leros/backend/tools"
)

const (
	// ToolNameSkillUse is the runtime tool used to discover and load skill documents.
	ToolNameSkillUse = "skill_use"
)

const (
	actionList     = "list"
	actionGet      = "get"
	actionReadFile = "read_file"
)

const (
	defaultSkillFileListLimit = 10
	maxSkillFileReadBytes     = 128 * 1024
)

// SkillUseTool lets an agent query and load skills from the runtime skill catalog.
type SkillUseTool struct {
	tools.BaseTool
	provider skillcatalog.CatalogProvider
}

// NewSkillUseTool creates a catalog-backed skill use tool.
func NewSkillUseTool(catalog skillcatalog.SkillCatalog) *SkillUseTool {
	return NewSkillUseToolWithProvider(skillcatalog.NewStaticCatalogProvider(catalog))
}

// NewSkillUseToolWithProvider creates a provider-backed skill use tool.
func NewSkillUseToolWithProvider(provider skillcatalog.CatalogProvider) *SkillUseTool {
	return &SkillUseTool{
		BaseTool: tools.NewBaseTool(
			ToolNameSkillUse,
			strings.Join([]string{
				"管理和使用技能（Skill）。",
				"支持 list 列出所有可用技能，get 获取指定技能完整内容和可注入上下文，read_file 读取技能目录下的附加文件。",
				"当任务需要查看、选择或加载技能说明时调用此工具。",
			}, ""),
			tools.Schema{
				Type:     "object",
				Required: []string{"action"},
				Properties: map[string]*tools.Property{
					"action": {
						Type:        "string",
						Enum:        []string{actionList, actionGet, actionReadFile},
						Description: "操作类型：list 列出技能，get 获取技能正文，read_file 读取技能目录下的文件",
					},
					"name": {
						Type:        "string",
						Description: "技能名称，get 和 read_file 时必填",
					},
					"path": {
						Type:        "string",
						Description: "技能目录内的相对文件路径，read_file 时必填",
					},
				},
			},
		),
		provider: provider,
	}
}

// Validate checks skill use tool input.
func (t *SkillUseTool) Validate(input map[string]interface{}) error {
	if input == nil {
		return fmt.Errorf("input is required")
	}

	action := stringValue(input, "action")
	switch action {
	case actionList:
		return nil
	case actionGet:
		if stringValue(input, "name") == "" {
			return fmt.Errorf("name is required")
		}
		return nil
	case actionReadFile:
		if stringValue(input, "name") == "" {
			return fmt.Errorf("name is required")
		}
		if stringValue(input, "path") == "" {
			return fmt.Errorf("path is required")
		}
		return nil
	case "":
		return fmt.Errorf("action is required")
	default:
		return fmt.Errorf("unsupported action %q", action)
	}
}

// Execute performs the requested skill catalog action.
func (t *SkillUseTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	if err := t.Validate(input); err != nil {
		return "", err
	}
	catalog, err := t.currentCatalog()
	if err != nil {
		return "", err
	}
	if catalog == nil {
		return "", fmt.Errorf("skill catalog is required")
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	switch stringValue(input, "action") {
	case actionList:
		return tools.JSONString(listSkills(catalog))
	case actionGet:
		return tools.JSONString(getSkill(catalog, stringValue(input, "name")))
	case actionReadFile:
		return tools.JSONString(readSkillFile(catalog, stringValue(input, "name"), stringValue(input, "path")))
	default:
		return "", fmt.Errorf("unsupported action %q", stringValue(input, "action"))
	}
}

func (t *SkillUseTool) currentCatalog() (skillcatalog.SkillCatalog, error) {
	if t == nil || t.provider == nil {
		return nil, fmt.Errorf("skill catalog is required")
	}
	return t.provider.Current(), nil
}

func listSkills(catalog skillcatalog.SkillCatalog) map[string]interface{} {
	summaries := catalog.List()
	skills := make([]map[string]interface{}, 0, len(summaries))
	for _, summary := range summaries {
		skills = append(skills, summaryMap(summary))
	}

	return map[string]interface{}{
		"success": true,
		"count":   len(skills),
		"skills":  skills,
	}
}

func getSkill(catalog skillcatalog.SkillCatalog, name string) map[string]interface{} {
	entry, err := findSkill(catalog, name)
	if err != nil {
		return skillNotFound(name, catalog.List())
	}

	files, err := catalog.ListFiles(entry.Manifest.Name, defaultSkillFileListLimit)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	skillDir := displaySkillDir(entry)
	return map[string]interface{}{
		"success":          true,
		"name":             entry.Manifest.Name,
		"description":      entry.Manifest.Description,
		"version":          entry.Manifest.Version,
		"category":         entry.Manifest.Metadata.Leros.Category,
		"tags":             entry.Manifest.Metadata.Leros.Tags,
		"related_skills":   []string{},
		"content":          entry.Body,
		"path":             entry.Path,
		"skill_dir":        skillDir,
		"linked_files":     optionalFiles(files),
		"usage_hint":       usageHint(files),
		"always":           entry.Manifest.Metadata.Leros.Always,
		"requires_tools":   entry.Manifest.Metadata.Leros.RequiresTools,
		"scope":            "catalog",
		"skill_type":       "file",
		"enabled":          true,
		"file_list_limit":  defaultSkillFileListLimit,
		"setup_needed":     false,
		"readiness_status": "available",
	}
}

func readSkillFile(catalog skillcatalog.SkillCatalog, name string, relativePath string) map[string]interface{} {
	entry, err := findSkill(catalog, name)
	if err != nil {
		return skillNotFound(name, catalog.List())
	}

	content, err := catalog.ReadFile(entry.Manifest.Name, relativePath)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	displayContent, truncated := truncateFileContent(content, maxSkillFileReadBytes)
	return map[string]interface{}{
		"success":         true,
		"name":            entry.Manifest.Name,
		"path":            relativePath,
		"content":         displayContent,
		"size":            len(content),
		"truncated":       truncated,
		"max_read_bytes":  maxSkillFileReadBytes,
		"original_length": len(content),
	}
}

func summaryMap(summary skillcatalog.Summary) map[string]interface{} {
	return map[string]interface{}{
		"name":           summary.Name,
		"description":    summary.Description,
		"version":        summary.Version,
		"category":       summary.Category,
		"tags":           summary.Tags,
		"always":         summary.Always,
		"requires_tools": summary.RequiresTools,
		"scope":          "catalog",
		"skill_type":     "file",
		"enabled":        true,
	}
}

func optionalFiles(files []string) interface{} {
	if len(files) == 0 {
		return nil
	}
	return files
}

func usageHint(files []string) interface{} {
	if len(files) == 0 {
		return nil
	}
	return "To view linked files, call skill_use with action=read_file and path set to a linked file path."
}

func findSkill(catalog skillcatalog.SkillCatalog, name string) (*skillcatalog.Entry, error) {
	entry, err := catalog.Get(name)
	if err == nil {
		return entry, nil
	}

	for _, summary := range catalog.List() {
		if !strings.EqualFold(summary.Name, name) {
			continue
		}
		return catalog.Get(summary.Name)
	}

	return nil, err
}

func displaySkillDir(entry *skillcatalog.Entry) string {
	if entry == nil {
		return ""
	}
	if entry.AbsoluteDir != "" {
		return entry.AbsoluteDir
	}
	return entry.Dir
}

func truncateFileContent(content []byte, maxBytes int) (string, bool) {
	if maxBytes <= 0 || len(content) <= maxBytes {
		return string(content), false
	}

	truncated := content[:maxBytes]
	for len(truncated) > 0 && !utf8.Valid(truncated) {
		truncated = truncated[:len(truncated)-1]
	}

	return string(truncated), true
}

func skillNotFound(name string, summaries []skillcatalog.Summary) map[string]interface{} {
	available := make([]string, 0, len(summaries))
	for _, summary := range summaries {
		available = append(available, summary.Name)
	}

	return map[string]interface{}{
		"success":   false,
		"error":     fmt.Sprintf("skill %q not found", name),
		"available": available,
	}
}

func stringValue(input map[string]interface{}, key string) string {
	value, _ := input[key].(string)
	return strings.TrimSpace(value)
}

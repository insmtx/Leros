package catalog

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest 描述文件型 Skill 的元数据区域。
type Manifest struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Version     string           `yaml:"version,omitempty"`
	Metadata    ManifestMetadata `yaml:"metadata,omitempty"`
}

// Normalize 在解析 Skill 文档后补齐派生默认值。
func (m *Manifest) Normalize(defaultName string) {
	if m.Name == "" {
		m.Name = defaultName
	}

	if m.Description == "" {
		m.Description = m.Name
	}
}

// ManifestMetadata 存储 Leros 专用的元数据扩展。
type ManifestMetadata struct {
	Leros LerosMetadata `yaml:"leros,omitempty"`
}

// LerosMetadata 存储运行时使用的第一组 Skill 路由提示。
type LerosMetadata struct {
	Category      string   `yaml:"category,omitempty"`
	Tags          []string `yaml:"tags,omitempty"`
	Always        bool     `yaml:"always,omitempty"`
	RequiresTools []string `yaml:"requires_tools,omitempty"`
}

// Entry 表示一个已发现并解析出元数据和正文的 Skill 文档。
type Entry struct {
	Manifest Manifest
	Body     string
	Dir      string
	Path     string
}

// Summary 是注入运行时提示词的紧凑视图。
type Summary struct {
	Name          string
	Description   string
	Version       string
	Category      string
	Tags          []string
	Always        bool
	RequiresTools []string
}

// Summary 返回适合提示词使用的 Skill 条目摘要。
func (e *Entry) Summary() Summary {
	return Summary{
		Name:          e.Manifest.Name,
		Description:   e.Manifest.Description,
		Version:       e.Manifest.Version,
		Category:      e.Manifest.Metadata.Leros.Category,
		Tags:          e.Manifest.Metadata.Leros.Tags,
		Always:        e.Manifest.Metadata.Leros.Always,
		RequiresTools: e.Manifest.Metadata.Leros.RequiresTools,
	}
}

// ParseDocument 解析带可选 YAML frontmatter 的 SKILL.md 文档。
func ParseDocument(raw []byte) (*Manifest, string, error) {
	manifest := &Manifest{}
	content := strings.TrimSpace(string(raw))
	if content == "" {
		return manifest, "", nil
	}

	if !strings.HasPrefix(content, "---") {
		return manifest, content, nil
	}

	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return nil, "", fmt.Errorf("invalid frontmatter header")
	}

	endIndex := -1
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			endIndex = idx
			break
		}
	}
	if endIndex == -1 {
		return nil, "", fmt.Errorf("frontmatter closing delimiter not found")
	}

	var yamlBuffer bytes.Buffer
	for idx := 1; idx < endIndex; idx++ {
		yamlBuffer.WriteString(lines[idx])
		yamlBuffer.WriteByte('\n')
	}
	if err := yaml.Unmarshal(yamlBuffer.Bytes(), manifest); err != nil {
		return nil, "", fmt.Errorf("unmarshal frontmatter: %w", err)
	}

	body := strings.Join(lines[endIndex+1:], "\n")
	return manifest, strings.TrimSpace(body), nil
}

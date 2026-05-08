package catalog

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/insmtx/SingerOS/backend/pkg/singeros"
)

const skillFileName = "SKILL.md"

var defaultSkillDirs = []string{
	"./backend/skills",
	"/app/backend/skills",
}

// Catalog 存储已发现的文件型 Skill，用于运行时提示词组装。
type Catalog struct {
	fs      fs.FS
	entryFS map[string]fs.FS
	entries map[string]*Entry
}

// LoadDefaultCatalog 从默认 SingerOS Skill 目录加载 Skill。
func LoadDefaultCatalog() (*Catalog, string, error) {
	candidates := defaultSkillDirCandidates()

	merged := NewEmptyCatalog()
	loadedDirs := make([]string, 0, len(candidates))
	var lastErr error
	for _, dir := range candidates {
		if _, err := os.Stat(dir); err != nil {
			lastErr = err
			continue
		}

		catalog, err := NewCatalog(os.DirFS(dir))
		if err != nil {
			lastErr = err
			continue
		}

		merged.merge(catalog)
		loadedDirs = append(loadedDirs, dir)
	}

	if len(loadedDirs) > 0 {
		return merged, strings.Join(loadedDirs, ","), nil
	}

	if lastErr != nil {
		return nil, "", fmt.Errorf("load skills from default directories: %w", lastErr)
	}
	return nil, "", fmt.Errorf("load skills from default directories: no candidates configured")
}

func defaultSkillDirCandidates() []string {
	candidates := append([]string{}, defaultSkillDirs...)
	if userDir, err := defaultSingerOSSkillsDir(); err == nil {
		candidates = append([]string{userDir}, candidates...)
	}
	return candidates
}

func defaultSingerOSSkillsDir() (string, error) {
	skillsDir, err := singeros.SkillsDir()
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(skillsDir), nil
}

// NewEmptyCatalog 创建一个不加载任何 Skill 的空 Catalog。
func NewEmptyCatalog() *Catalog {
	return &Catalog{
		entryFS: make(map[string]fs.FS),
		entries: make(map[string]*Entry),
	}
}

// NewCatalog 扫描直接子目录中的 SKILL.md 文件并创建 Catalog。
func NewCatalog(skillFS fs.FS) (*Catalog, error) {
	entries := make(map[string]*Entry)
	entryFS := make(map[string]fs.FS)

	rootEntries, err := fs.ReadDir(skillFS, ".")
	if err != nil {
		return nil, fmt.Errorf("read skill root directory: %w", err)
	}
	for _, rootEntry := range rootEntries {
		if !rootEntry.IsDir() {
			continue
		}

		dir := rootEntry.Name()
		filePath := path.Join(dir, skillFileName)
		raw, err := fs.ReadFile(skillFS, filePath)
		if err != nil {
			continue
		}

		manifest, body, err := ParseDocument(raw)
		if err != nil {
			return nil, fmt.Errorf("parse skill file %s: %w", filePath, err)
		}

		manifest.Normalize(path.Base(dir))

		entry := &Entry{
			Manifest: *manifest,
			Body:     body,
			Dir:      dir,
			Path:     filePath,
		}
		if _, exists := entries[entry.Manifest.Name]; exists {
			return nil, fmt.Errorf("duplicate skill name %q", entry.Manifest.Name)
		}
		entries[entry.Manifest.Name] = entry
		entryFS[entry.Manifest.Name] = skillFS
	}

	return &Catalog{
		fs:      skillFS,
		entryFS: entryFS,
		entries: entries,
	}, nil
}

func (c *Catalog) merge(other *Catalog) {
	if c == nil || other == nil {
		return
	}
	if c.entries == nil {
		c.entries = make(map[string]*Entry)
	}
	if c.entryFS == nil {
		c.entryFS = make(map[string]fs.FS)
	}
	for name, entry := range other.entries {
		if _, exists := c.entries[name]; exists {
			continue
		}
		c.entries[name] = entry
		if sourceFS := other.entryFS[name]; sourceFS != nil {
			c.entryFS[name] = sourceFS
		} else {
			c.entryFS[name] = other.fs
		}
	}
}

// List 返回按名称排序的 Skill 摘要。
func (c *Catalog) List() []Summary {
	if c == nil {
		return nil
	}

	summaries := make([]Summary, 0, len(c.entries))
	for _, entry := range c.entries {
		summaries = append(summaries, entry.Summary())
	}

	slices.SortFunc(summaries, func(left, right Summary) int {
		return strings.Compare(left.Name, right.Name)
	})

	return summaries
}

// Get 按名称返回完整的 Skill 条目。
func (c *Catalog) Get(name string) (*Entry, error) {
	if c == nil {
		return nil, fmt.Errorf("catalog is nil")
	}

	entry, ok := c.entries[name]
	if !ok {
		return nil, fmt.Errorf("skill %q not found", name)
	}

	return entry, nil
}

// ReadFile 读取 Skill 目录下的附加文件。
func (c *Catalog) ReadFile(name string, relativePath string) ([]byte, error) {
	entry, err := c.Get(name)
	if err != nil {
		return nil, err
	}

	cleanPath := path.Clean(relativePath)
	if cleanPath == "." || strings.HasPrefix(cleanPath, "../") || path.IsAbs(cleanPath) {
		return nil, fmt.Errorf("invalid skill file path %q", relativePath)
	}

	fullPath := cleanPath
	if entry.Dir != "" {
		fullPath = path.Join(entry.Dir, cleanPath)
	}

	skillFS := c.fsForSkill(name)
	content, err := fs.ReadFile(skillFS, fullPath)
	if err != nil {
		return nil, fmt.Errorf("read skill file %s: %w", fullPath, err)
	}

	return content, nil
}

// ListFiles 返回 Skill 目录下除 SKILL.md 以外的附加文件。
func (c *Catalog) ListFiles(name string, limit int) ([]string, error) {
	entry, err := c.Get(name)
	if err != nil {
		return nil, err
	}

	root := entry.Dir
	if root == "" {
		root = "."
	}

	files := make([]string, 0)
	skillFS := c.fsForSkill(name)
	err = fs.WalkDir(skillFS, root, func(filePath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if path.Base(filePath) == skillFileName {
			return nil
		}

		relativePath := filePath
		if entry.Dir != "" {
			relativePath = strings.TrimPrefix(filePath, entry.Dir+"/")
		}
		files = append(files, relativePath)
		if limit > 0 && len(files) >= limit {
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list skill files %s: %w", name, err)
	}

	slices.Sort(files)
	return files, nil
}

func (c *Catalog) fsForSkill(name string) fs.FS {
	if c != nil && c.entryFS != nil {
		if skillFS := c.entryFS[name]; skillFS != nil {
			return skillFS
		}
	}
	if c != nil && c.fs != nil {
		return c.fs
	}
	return os.DirFS(".")
}

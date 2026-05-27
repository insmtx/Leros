package catalog

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/insmtx/Leros/backend/pkg/leros"
)

const skillFileName = "SKILL.md"

// Catalog stores discovered file-based skills for runtime prompt assembly.
type Catalog struct {
	rootDir string
	fs      fs.FS
	entryFS map[string]fs.FS
	entries map[string]*Entry
}

// LoadDefaultCatalog loads skills from the configured default directories.
func LoadDefaultCatalog() (*Catalog, string, error) {
	dir, err := defaultLerosSkillsDir()
	if err != nil {
		return nil, "", fmt.Errorf("resolve default skill directory: %w", err)
	}

	if err := os.MkdirAll(filepath.FromSlash(dir), 0o755); err != nil {
		return nil, "", fmt.Errorf("create default skill directory %s: %w", dir, err)
	}

	catalog, err := NewCatalogFromDir(dir)
	if err != nil {
		return nil, "", err
	}

	return catalog, catalog.rootDir, nil
}

func defaultLerosSkillsDir() (string, error) {
	skillsDir, err := leros.SkillsDir()
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(skillsDir), nil
}

// NewEmptyCatalog creates a catalog without loading any skills.
func NewEmptyCatalog() *Catalog {
	return &Catalog{
		entryFS: make(map[string]fs.FS),
		entries: make(map[string]*Entry),
	}
}

// NewCatalog scans direct child directories in the provided filesystem for SKILL.md files.
func NewCatalog(skillFS fs.FS) (*Catalog, error) {
	return newCatalog(skillFS, "")
}

// NewCatalogFromDir scans a filesystem directory and preserves its absolute path for display metadata.
func NewCatalogFromDir(rootDir string) (*Catalog, error) {
	absolute, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolve skill root directory %s: %w", rootDir, err)
	}
	absolute = filepath.Clean(absolute)
	return newCatalog(os.DirFS(absolute), filepath.ToSlash(absolute))
}

func newCatalog(skillFS fs.FS, rootDir string) (*Catalog, error) {
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
			Manifest:    *manifest,
			Body:        body,
			Dir:         dir,
			Path:        filePath,
			AbsoluteDir: absoluteSkillDir(rootDir, dir),
		}
		if _, exists := entries[entry.Manifest.Name]; exists {
			return nil, fmt.Errorf("duplicate skill name %q", entry.Manifest.Name)
		}
		entries[entry.Manifest.Name] = entry
		entryFS[entry.Manifest.Name] = skillFS
	}

	return &Catalog{
		rootDir: rootDir,
		fs:      skillFS,
		entryFS: entryFS,
		entries: entries,
	}, nil
}

func absoluteSkillDir(rootDir string, dir string) string {
	if rootDir == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Join(filepath.FromSlash(rootDir), filepath.FromSlash(dir)))
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

// List returns skill summaries sorted by name.
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

// Get returns a full skill entry by name.
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

// ReadFile reads a supporting file from a skill directory.
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

// ListFiles returns supporting files in a skill directory, excluding SKILL.md.
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

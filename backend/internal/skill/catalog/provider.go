package catalog

import (
	"context"
	"fmt"
	"sync"
)

// SkillCatalog 是运行时读取 Skill 所需的只读 Catalog 接口。
type SkillCatalog interface {
	List() []Summary
	Get(name string) (*Entry, error)
	ReadFile(name string, relativePath string) ([]byte, error)
	ListFiles(name string, limit int) ([]string, error)
}

// CatalogProvider 为运行时读取返回最新的 Skill Catalog 快照。
type CatalogProvider interface {
	Current() SkillCatalog
}

// CatalogReloader 从后端来源刷新 Catalog 快照。
type CatalogReloader interface {
	Reload(ctx context.Context) error
}

// FileCatalogProvider 从默认文件来源加载 SingerOS Skill。
type FileCatalogProvider struct {
	mu      sync.RWMutex
	catalog SkillCatalog
	dirs    string
}

// NewFileCatalogProvider 创建 Provider 并加载初始 Catalog 快照。
func NewFileCatalogProvider(ctx context.Context) (*FileCatalogProvider, error) {
	provider := &FileCatalogProvider{}
	if err := provider.Reload(ctx); err != nil {
		return nil, err
	}
	return provider, nil
}

// NewStaticCatalogProvider 将固定 Catalog 包装成 Provider，主要用于测试和显式运行时装配。
func NewStaticCatalogProvider(catalog SkillCatalog) CatalogProvider {
	if catalog == nil {
		catalog = NewEmptyCatalog()
	}
	return staticCatalogProvider{catalog: catalog}
}

type staticCatalogProvider struct {
	catalog SkillCatalog
}

// Current 返回固定的 Catalog 快照。
func (p staticCatalogProvider) Current() SkillCatalog {
	if p.catalog == nil {
		return NewEmptyCatalog()
	}
	return p.catalog
}

// Current 返回最近一次加载的 Catalog 快照。
func (p *FileCatalogProvider) Current() SkillCatalog {
	if p == nil {
		return NewEmptyCatalog()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.catalog == nil {
		return NewEmptyCatalog()
	}
	return p.catalog
}

// LoadedDirs 返回最近一次 reload 加载过的目录，多个目录用逗号分隔。
func (p *FileCatalogProvider) LoadedDirs() string {
	if p == nil {
		return ""
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.dirs
}

// Reload 从默认 SingerOS Skill 目录重新加载 Catalog。
func (p *FileCatalogProvider) Reload(ctx context.Context) error {
	if p == nil {
		return fmt.Errorf("catalog provider is nil")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	catalog, dirs, err := LoadDefaultCatalog()
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.catalog = catalog
	p.dirs = dirs
	return nil
}

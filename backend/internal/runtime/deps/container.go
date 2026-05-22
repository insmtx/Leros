package deps

import (
	"context"
	"fmt"
	"strings"

	skillcatalog "github.com/insmtx/Leros/backend/internal/skill/catalog"
	skillmanage "github.com/insmtx/Leros/backend/internal/skill/manage"
	skillstore "github.com/insmtx/Leros/backend/internal/skill/store"
	"github.com/insmtx/Leros/backend/tools"
	memorytools "github.com/insmtx/Leros/backend/tools/memory"
	nodetools "github.com/insmtx/Leros/backend/tools/node"
	skillmanagetools "github.com/insmtx/Leros/backend/tools/skill_manage"
	skillusetools "github.com/insmtx/Leros/backend/tools/skill_use"
	todotools "github.com/insmtx/Leros/backend/tools/todo"
	"github.com/ygpkg/yg-go/logs"
)

// Options 控制运行时依赖组装。
type Options struct {
	ToolsEnabled bool
}

// Container 持有生命周期和具体运行时共享的运行时依赖。
type Container struct {
	skillsProvider skillcatalog.CatalogProvider
	toolRegistry   *tools.Registry
}

// New 为单个工作进程构建共享的运行时依赖容器。
func New(ctx context.Context, opts Options) (*Container, error) {
	catalogProvider, err := skillcatalog.NewFileCatalogProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("load skills: %w", err)
	}

	logs.Infof("Loaded %d skills from %s for runtime", len(catalogProvider.Current().List()), catalogProvider.LoadedDirs())

	registry := tools.NewRegistry()
	if opts.ToolsEnabled {
		if err := registerTools(registry, catalogProvider); err != nil {
			return nil, err
		}
	}
	logs.Infof("Loaded %d tools for runtime", len(registry.List()))

	return &Container{
		skillsProvider: catalogProvider,
		toolRegistry:   registry,
	}, nil
}

// SkillsProvider 返回可重载的技能目录提供者。
func (c *Container) SkillsProvider() skillcatalog.CatalogProvider {
	if c == nil || c.skillsProvider == nil {
		return skillcatalog.NewStaticCatalogProvider(skillcatalog.NewEmptyCatalog())
	}
	return c.skillsProvider
}

// ToolRegistry 返回运行时工具注册表。
func (c *Container) ToolRegistry() *tools.Registry {
	if c == nil || c.toolRegistry == nil {
		return tools.NewRegistry()
	}
	return c.toolRegistry
}

// AvailableToolNames 从请求列表中返回已注册的工具名称。
func (c *Container) AvailableToolNames(names []string) []string {
	if c == nil || c.toolRegistry == nil || len(names) == 0 {
		return nil
	}
	result := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		if _, err := c.toolRegistry.Get(name); err == nil {
			result = append(result, name)
			seen[name] = struct{}{}
		}
	}
	return result
}

func registerTools(registry *tools.Registry, catalogProvider *skillcatalog.FileCatalogProvider) error {
	if err := skillusetools.RegisterWithProvider(registry, catalogProvider); err != nil {
		return fmt.Errorf("register skill use tool: %w", err)
	}
	store, err := skillstore.NewSkillStore("")
	if err != nil {
		return fmt.Errorf("new skill store: %w", err)
	}
	manager, err := skillmanage.NewManager(store, skillmanage.NewPostProcessor(store.RootDir(), catalogProvider))
	if err != nil {
		return fmt.Errorf("new skill manager: %w", err)
	}
	if err := skillmanagetools.RegisterWithManager(registry, manager); err != nil {
		return fmt.Errorf("register skill manage tool: %w", err)
	}
	if err := memorytools.Register(registry); err != nil {
		return fmt.Errorf("register memory tool: %w", err)
	}
	if err := todotools.Register(registry); err != nil {
		return fmt.Errorf("register todo tool: %w", err)
	}
	if err := nodetools.Register(registry); err != nil {
		return fmt.Errorf("register node tools: %w", err)
	}
	return nil
}

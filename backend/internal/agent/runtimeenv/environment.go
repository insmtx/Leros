package runtimeenv

import (
	"context"
	"fmt"
	"strings"

	skillcatalog "github.com/insmtx/SingerOS/backend/internal/skill/catalog"
	skillruntime "github.com/insmtx/SingerOS/backend/internal/skill/runtime"
	skillstore "github.com/insmtx/SingerOS/backend/internal/skill/store"
	"github.com/insmtx/SingerOS/backend/tools"
	memorytools "github.com/insmtx/SingerOS/backend/tools/memory"
	skillmanagetools "github.com/insmtx/SingerOS/backend/tools/skill_manage"
	skillusetools "github.com/insmtx/SingerOS/backend/tools/skill_use"
	"github.com/ygpkg/yg-go/logs"
)

// Options controls runtime environment assembly.
type Options struct {
	ToolsEnabled bool
}

// Environment owns runtime dependencies shared by lifecycle and concrete runtimes.
type Environment struct {
	skillsProvider skillcatalog.CatalogProvider
	toolRegistry   *tools.Registry
}

// New builds the shared runtime environment for one worker process.
func New(ctx context.Context, opts Options) (*Environment, error) {
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

	return &Environment{
		skillsProvider: catalogProvider,
		toolRegistry:   registry,
	}, nil
}

// SkillsProvider returns the reloadable skill catalog provider.
func (e *Environment) SkillsProvider() skillcatalog.CatalogProvider {
	if e == nil || e.skillsProvider == nil {
		return skillcatalog.NewStaticCatalogProvider(skillcatalog.NewEmptyCatalog())
	}
	return e.skillsProvider
}

// ToolRegistry returns the runtime tool registry.
func (e *Environment) ToolRegistry() *tools.Registry {
	if e == nil || e.toolRegistry == nil {
		return tools.NewRegistry()
	}
	return e.toolRegistry
}

// AvailableToolNames returns registered tool names from the requested list.
func (e *Environment) AvailableToolNames(names []string) []string {
	if e == nil || e.toolRegistry == nil || len(names) == 0 {
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
		if _, err := e.toolRegistry.Get(name); err == nil {
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
	manager, err := skillruntime.NewManager(store, skillruntime.NewPostProcessor(store.RootDir(), catalogProvider))
	if err != nil {
		return fmt.Errorf("new skill manager: %w", err)
	}
	if err := skillmanagetools.RegisterWithManager(registry, manager); err != nil {
		return fmt.Errorf("register skill manage tool: %w", err)
	}
	if err := memorytools.Register(registry); err != nil {
		return fmt.Errorf("register memory tool: %w", err)
	}
	return nil
}

package deps

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/insmtx/Leros/backend/engines"
	skillstore "github.com/insmtx/Leros/backend/internal/skill/store"
	"github.com/insmtx/Leros/backend/tools"
	memorytools "github.com/insmtx/Leros/backend/tools/memory"
	nodetools "github.com/insmtx/Leros/backend/tools/node"
	skillmanagetools "github.com/insmtx/Leros/backend/tools/skill_manage"
	skillusetools "github.com/insmtx/Leros/backend/tools/skill_use"
	todotools "github.com/insmtx/Leros/backend/tools/todo"
	"github.com/ygpkg/yg-go/logs"
)

type Options struct {
	CLISkillDirs []string
}

type Container struct {
	toolRegistry *tools.Registry
}

var (
	defaultContainer     *Container
	defaultContainerInit sync.Mutex
)

func Default(ctx context.Context) (*Container, error) {
	defaultContainerInit.Lock()
	defer defaultContainerInit.Unlock()

	if defaultContainer != nil {
		return defaultContainer, nil
	}

	c, err := New(ctx, Options{})
	if err != nil {
		return nil, err
	}
	defaultContainer = c
	return defaultContainer, nil
}

func ResetDefaultForTest() {
	defaultContainerInit.Lock()
	defer defaultContainerInit.Unlock()
	defaultContainer = nil
}

func New(ctx context.Context, opts Options) (*Container, error) {
	registry := tools.NewRegistry()
	if err := registerTools(registry, opts.CLISkillDirs); err != nil {
		return nil, err
	}
	logs.Infof("Loaded %d tools for runtime", len(registry.List()))

	return &Container{
		toolRegistry: registry,
	}, nil
}

func (c *Container) ToolRegistry() *tools.Registry {
	if c == nil || c.toolRegistry == nil {
		return tools.NewRegistry()
	}
	return c.toolRegistry
}

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

func registerTools(registry *tools.Registry, cliSkillDirs []string) error {
	if err := skillusetools.Register(registry); err != nil {
		return fmt.Errorf("register skill use tool: %w", err)
	}
	// skill_manage mutation 后处理：create 时创建外部 CLI symlink，delete 时清理。
	// NewTool() 内部创建的 SkillStore 会读取此回调。
	skillmanagetools.OnMutation = func(ctx context.Context, kind skillstore.MutationKind, name, action string) {
		if len(cliSkillDirs) > 0 {
			switch kind {
			case skillstore.MutationCreate:
				_ = engines.EnsureExternalSkillLink(name, cliSkillDirs)
			case skillstore.MutationDelete:
				_ = engines.RemoveExternalSkillLink(name, cliSkillDirs)
			}
		}
	}

	if err := skillmanagetools.Register(registry); err != nil {
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

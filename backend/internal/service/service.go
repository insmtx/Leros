package service

import (
	"context"
	"fmt"

	auth "github.com/insmtx/SingerOS/backend/auth"
	githubprovider "github.com/insmtx/SingerOS/backend/pkg/providers/github"
	"github.com/insmtx/SingerOS/backend/config"
	githubauth "github.com/insmtx/SingerOS/backend/pkg/providers/github"
	agentruntime "github.com/insmtx/SingerOS/backend/internal/agent"
	bundledskills "github.com/insmtx/SingerOS/backend/skills/bundled"
	skillcatalog "github.com/insmtx/SingerOS/backend/skills/catalog"
	"github.com/insmtx/SingerOS/backend/toolruntime"
	"github.com/insmtx/SingerOS/backend/tools"
	githubtools "github.com/insmtx/SingerOS/backend/tools/github"
)

// NewAuthService creates a new auth service with configured providers.
func NewAuthService(cfg *config.Config) *auth.Service {
	store := auth.NewInMemoryStore()
	resolver := auth.NewAccountResolver(store)
	svc := auth.NewService(store, resolver)

	if cfg != nil && cfg.Github != nil {
		svc.RegisterProvider(githubauth.NewOAuthProvider(*cfg.Github))
	}

	return svc
}

// NewToolRegistry creates a new tool registry with all registered tools.
func NewToolRegistry(cfg *config.Config, authService *auth.Service) (*tools.Registry, *toolruntime.Runtime, error) {
	registry := tools.NewRegistry()

	var githubClientFactory *githubprovider.ClientFactory
	if cfg != nil && cfg.Github != nil {
		githubClientFactory = githubprovider.NewClientFactory(*cfg.Github, authService)

		registerTool := func(tool tools.Tool, name string) error {
			if err := registry.Register(tool); err != nil {
				return fmt.Errorf("register %s: %w", name, err)
			}
			return nil
		}

		if err := registerTool(githubtools.NewAccountInfoTool(nil), "account_info"); err != nil {
			return nil, nil, err
		}
		if err := registerTool(githubtools.NewPullRequestMetadataTool(nil), "pr_metadata"); err != nil {
			return nil, nil, err
		}
		if err := registerTool(githubtools.NewPullRequestFilesTool(nil), "pr_files"); err != nil {
			return nil, nil, err
		}
		if err := registerTool(githubtools.NewRepositoryFileTool(nil), "repo_file"); err != nil {
			return nil, nil, err
		}
		if err := registerTool(githubtools.NewCompareCommitsTool(nil), "compare_commits"); err != nil {
			return nil, nil, err
		}
		if err := registerTool(githubtools.NewPullRequestReviewPublishTool(nil), "pr_review_publish"); err != nil {
			return nil, nil, err
		}
	}

	return registry, toolruntime.New(registry, githubClientFactory), nil
}

// NewRuntimeConfig creates a new runtime configuration.
func NewRuntimeConfig(cfg *config.Config, authService *auth.Service) (agentruntime.Config, error) {
	catalog, err := skillcatalog.New(bundledskills.FS)
	if err != nil {
		return agentruntime.Config{}, fmt.Errorf("load bundled skills: %w", err)
	}

	toolRegistry, toolRuntime, err := NewToolRegistry(cfg, authService)
	if err != nil {
		return agentruntime.Config{}, err
	}

	return agentruntime.Config{
		SkillsCatalog: catalog,
		ToolRegistry:  toolRegistry,
		ToolRuntime:   toolRuntime,
	}, nil
}

// NewEinoRunner creates a new Eino runtime runner.
func NewEinoRunner(ctx context.Context, cfg *config.Config, runtimeConfig agentruntime.Config) (agentruntime.Runtime, error) {
	if cfg == nil || cfg.LLM == nil || cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("llm config is required")
	}

	switch cfg.LLM.Provider {
	case "", "openai":
		return agentruntime.NewEinoRunner(ctx, cfg.LLM, runtimeConfig)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.LLM.Provider)
	}
}

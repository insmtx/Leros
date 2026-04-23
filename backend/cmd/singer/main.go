package main

import (
	"os"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/database"
	agentruntime "github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/internal/eventengine"
	"github.com/insmtx/SingerOS/backend/internal/execution"
	"github.com/insmtx/SingerOS/backend/internal/infra/mq/rabbitmq"
	"github.com/insmtx/SingerOS/backend/internal/service/middleware"
	trace "github.com/insmtx/SingerOS/backend/internal/service/middleware"
	agentruntime "github.com/insmtx/SingerOS/backend/runtime"
	githubprovider "github.com/insmtx/SingerOS/backend/pkg/providers/github"
	bundledskills "github.com/insmtx/SingerOS/backend/skills/bundled"
	skillcatalog "github.com/insmtx/SingerOS/backend/skills/catalog"
	"github.com/insmtx/SingerOS/backend/toolruntime"
	"github.com/insmtx/SingerOS/backend/tools"
	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/logs"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "singer",
	Short: "Backend service for the SingerOS Backend",
	Long:  `This is the backend service for the SingerOS Backend, responsible for handling API requests and business logic.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Configuration file path")
}

func loadConfig() (*config.Config, error) {
<<<<<<< HEAD
	var cfg config.Config

	if configPath != "" {
		// Load config from specified path
		err := ygconfig.LoadYamlLocalFile(configPath, &cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %v", configPath, err)
		}
	} else {
		// Try to load from default locations
		pathsToTry := []string{"./config.yaml", "/app/config.yaml"}

		err := fmt.Errorf("config file not found in any location")
		for _, path := range pathsToTry {
			if err = ygconfig.LoadYamlLocalFile(path, &cfg); err == nil {
				logs.Infof("Loaded config from: %s", path)
				break
			}
		}

		if err != nil {
			logs.Warnf("Could not load config from any path (%v), will proceed without config", err)
			return &config.Config{}, nil
		}
	}

	logs.Info("Configuration loaded successfully")
	return &cfg, nil
}

func buildRuntimeConfig() (agentruntime.Config, error) {
	catalog, skillDir, err := skilltools.LoadDefaultCatalog()
	if err != nil {
		return agentruntime.Config{}, fmt.Errorf("load skills: %w", err)
	}

	logs.Infof("Loaded %d skills from %s for runtime", len(catalog.List()), skillDir)

	toolRegistry, err := buildTooling(catalog)
	if err != nil {
		return agentruntime.Config{}, err
	}

	return agentruntime.Config{
		SkillsCatalog: catalog,
		ToolRegistry:  toolRegistry,
	}, nil
}

func buildAuthService(cfg *config.Config) *auth.Service {
	accountStore := auth.NewInMemoryStore()
	accountResolver := auth.NewAccountResolver(accountStore)
	authService := auth.NewService(accountStore, accountResolver)

	if cfg != nil && cfg.Github != nil {
		authService.RegisterProvider(github.NewOAuthProvider(*cfg.Github))
	}

	return authService
}

func buildTooling(catalog *skilltools.Catalog) (*tools.Registry, error) {
	registry := tools.NewRegistry()

<<<<<<< HEAD
	if err := skilltools.Register(registry, catalog); err != nil {
		return nil, fmt.Errorf("register skill use tool: %w", err)
=======
	var githubFactory *githubprovider.ClientFactory
	if cfg != nil && cfg.Github != nil {
		githubFactory = github.NewClientFactory(*cfg.Github, authService)
		if err := registry.Register(githubtools.NewAccountInfoTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github account info tool: %w", err)
		}
		if err := registry.Register(githubtools.NewPullRequestMetadataTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github pr metadata tool: %w", err)
		}
		if err := registry.Register(githubtools.NewPullRequestFilesTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github pr files tool: %w", err)
		}
		if err := registry.Register(githubtools.NewRepositoryFileTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github repository file tool: %w", err)
		}
		if err := registry.Register(githubtools.NewCompareCommitsTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github compare commits tool: %w", err)
		}
		if err := registry.Register(githubtools.NewPullRequestReviewPublishTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github pr review publish tool: %w", err)
		}
>>>>>>> a23a3b2 (refactor: 合并 providers 模块到 pkg/providers)
	}

	logs.Infof("Loaded %d tools for runtime", len(registry.List()))

	return registry, nil
}

func buildRuntimeRunner(ctx context.Context, cfg *config.Config, runtimeConfig agentruntime.Config) (agentruntime.Runtime, error) {
	if cfg == nil || cfg.LLM == nil || cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("llm config is required")
	}

	switch cfg.LLM.Provider {
	case "", "openai":
		logs.Info("Using SingerOS agent runtime")
		return agentruntime.NewAgent(ctx, cfg.LLM, runtimeConfig)
	default:
		return nil, fmt.Errorf("unsupported Eino chat model provider: %s", cfg.LLM.Provider)
	}
=======
	return config.Load(configPath)
>>>>>>> a93d3ab (refactor(cmd): 拆分 singer 命令为 server 和 worker 子命令)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logs.Errorf("Error executing command: %v", err)
		os.Exit(1)
	}
}

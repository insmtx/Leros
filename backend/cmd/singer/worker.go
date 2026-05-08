package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/internal/agent/externalcli"
	skillcatalog "github.com/insmtx/SingerOS/backend/internal/skill/catalog"
	skillruntime "github.com/insmtx/SingerOS/backend/internal/skill/runtime"
	skillstore "github.com/insmtx/SingerOS/backend/internal/skill/store"
	"github.com/insmtx/SingerOS/backend/internal/worker/client"
	singerMCP "github.com/insmtx/SingerOS/backend/mcp"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
	"github.com/insmtx/SingerOS/backend/runtime/engines/builtin"
	"github.com/insmtx/SingerOS/backend/tools"
	memorytools "github.com/insmtx/SingerOS/backend/tools/memory"
	skillmanagetools "github.com/insmtx/SingerOS/backend/tools/skill_manage"
	skillusetools "github.com/insmtx/SingerOS/backend/tools/skill_use"
	"github.com/spf13/cobra"
	ygconfig "github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
)

var (
	workerConfigPath     string
	workerServerAddr     string
	workerDefaultRuntime string
	workerListenAddr     string
	workerWorkerID       string
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the SingerOS background worker",
	Long:  `Start the background worker service for processing asynchronous tasks and events.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		worker, err := createWorker(ctx)
		if err != nil {
			logs.Fatalf("Failed to create worker: %v", err)
			return
		}

		if err := worker.Start(ctx); err != nil {
			logs.Fatalf("Failed to start worker: %v", err)
		}
	},
}

func init() {
	workerCmd.Flags().StringVar(&workerConfigPath, "config", "", "Configuration file path")
	workerCmd.Flags().StringVar(&workerServerAddr, "server-addr", "127.0.0.1:8080", "Server address for WebSocket connection")
	workerCmd.Flags().StringVar(&workerListenAddr, "listen-addr", ":8081", "Worker MCP server listen address for runtime bootstrap")
	workerCmd.Flags().StringVar(&workerWorkerID, "worker-id", "", "Worker ID for configuration retrieval")
	workerCmd.Flags().StringVar(&workerDefaultRuntime, "default-runtime", "", "Default agent runtime kind, for example singeros, claude, or codex")
	rootCmd.AddCommand(workerCmd)
}

func createWorker(ctx context.Context) (*client.WorkerClient, error) {
	_, err := loadWorkerConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return client.NewWorker(ctx, &client.WorkerConfig{
		ServerAddr:   workerServerAddr,
		WorkerID:     workerWorkerID,
		SkillsDir:    "",
		ToolsEnabled: true,
	})
}

func loadWorkerConfig() (*config.WorkerConfig, error) {
	cfg := &config.WorkerConfig{}
	if workerConfigPath != "" {
		err := ygconfig.LoadYamlLocalFile(workerConfigPath, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", workerConfigPath, err)
		}
	}
	if strings.TrimSpace(workerWorkerID) != "" {
		cfg.WorkerID = workerWorkerID
		logs.Infof("Using worker ID from flag: %s", workerWorkerID)
	}
	if strings.TrimSpace(workerServerAddr) != "" {
		cfg.ServerAddr = workerServerAddr
		logs.Infof("Using server address from flag: %s", workerServerAddr)
	}

	return cfg, nil
}

func startWorkerMCPServer(addr string) (*http.Server, error) {
	if strings.TrimSpace(addr) == "" {
		addr = ":8081"
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", addr, err)
	}

	r := gin.New()
	v1 := r.Group("/v1")
	singerMCP.RegisterRoutes(v1, singerMCP.NewServer())

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		logs.Infof("Worker MCP server listening on %s", listener.Addr().String())
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			logs.Errorf("Worker MCP server stopped unexpectedly: %v", err)
		}
	}()

	return server, nil
}

func defaultCLIBootstrapOptions(addr string) builtin.BootstrapOptions {
	return builtin.BootstrapOptions{
		MCP: engines.MCPServerConfig{
			Name:        "singeros",
			URL:         mcpURLFromAddr(addr),
			BearerToken: singerMCP.DefaultAuthToken(),
		},
	}
}

func mcpURLFromAddr(addr string) string {
	host := "localhost"
	port := "8081"

	if strings.TrimSpace(addr) != "" {
		if splitHost, splitPort, err := net.SplitHostPort(addr); err == nil {
			if splitHost != "" && splitHost != "0.0.0.0" && splitHost != "::" && splitHost != "[::]" {
				host = splitHost
			}
			if splitPort != "" {
				port = splitPort
			}
		} else if strings.HasPrefix(addr, ":") {
			port = strings.TrimPrefix(addr, ":")
		} else {
			host = addr
		}
	}

	return fmt.Sprintf("http://%s:%s/v1/mcp", host, port)
}

func buildRuntimeConfig() (agent.Config, error) {
	catalogProvider, err := skillcatalog.NewFileCatalogProvider(context.Background())
	if err != nil {
		return agent.Config{}, fmt.Errorf("load skills: %w", err)
	}

	logs.Infof("Loaded %d skills from %s for runtime", len(catalogProvider.Current().List()), catalogProvider.LoadedDirs())

	toolRegistry, err := buildTooling(catalogProvider)
	if err != nil {
		return agent.Config{}, err
	}

	return agent.Config{
		SkillsCatalogProvider: catalogProvider,
		ToolRegistry:          toolRegistry,
	}, nil
}

func buildTooling(catalogProvider *skillcatalog.FileCatalogProvider) (*tools.Registry, error) {
	registry := tools.NewRegistry()

	if err := skillusetools.RegisterWithProvider(registry, catalogProvider); err != nil {
		return nil, fmt.Errorf("register skill use tool: %w", err)
	}
	store, err := skillstore.NewSkillStore("")
	if err != nil {
		return nil, fmt.Errorf("new skill store: %w", err)
	}
	manager, err := skillruntime.NewManager(store, skillruntime.NewPostProcessor(store.RootDir(), catalogProvider))
	if err != nil {
		return nil, fmt.Errorf("new skill manager: %w", err)
	}
	if err := skillmanagetools.RegisterWithManager(registry, manager); err != nil {
		return nil, fmt.Errorf("register skill manage tool: %w", err)
	}
	if err := memorytools.Register(registry); err != nil {
		return nil, fmt.Errorf("register memory tool: %w", err)
	}

	logs.Infof("Loaded %d tools for runtime", len(registry.List()))

	return registry, nil
}

func buildRuntimeRunner(ctx context.Context, cfg *config.WorkerConfig, runtimeConfig agent.Config, defaultRuntime string) (agent.Runner, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	router := agent.NewRuntimeRouter(agent.RuntimeKindSingerOS)
	registered := 0
	registeredKinds := make(map[string]struct{})

	if cfg.LLM != nil && cfg.LLM.APIKey != "" {
		switch cfg.LLM.Provider {
		case "", "openai":
			logs.Info("Registering SingerOS agent runtime")
			singerRunner, err := agent.NewAgent(ctx, cfg.LLM, runtimeConfig)
			if err != nil {
				return nil, err
			}
			if err := router.Register(agent.RuntimeKindSingerOS, singerRunner); err != nil {
				return nil, err
			}
			registered++
			registeredKinds[agent.RuntimeKindSingerOS] = struct{}{}
		default:
			logs.Warnf("Skipping SingerOS agent runtime for unsupported Eino chat model provider: %s", cfg.LLM.Provider)
		}
	}

	cliRegistry, err := builtin.NewRegistryFromConfig(cfg.CLI)
	if err != nil {
		return nil, fmt.Errorf("create CLI engine registry: %w", err)
	}
	cliNames := cliRegistry.Names()
	for _, name := range cliNames {
		engine, ok := cliRegistry.Get(name)
		if !ok {
			continue
		}
		runner, err := externalcli.NewRunner(name, engine, cfg.LLM)
		if err != nil {
			return nil, err
		}
		if err := router.Register(name, runner); err != nil {
			return nil, err
		}
		registered++
		registeredKinds[strings.ToLower(strings.TrimSpace(name))] = struct{}{}
		logs.Infof("Registering external agent CLI runtime: %s", name)
	}

	if registered == 0 {
		return nil, fmt.Errorf("no agent runtime is available")
	}

	selectedDefault := selectDefaultRuntime(defaultRuntime, cfg, cliNames)
	if selectedDefault == "" {
		selectedDefault = agent.RuntimeKindSingerOS
	}
	normalizedDefault := strings.ToLower(strings.TrimSpace(selectedDefault))
	if _, ok := registeredKinds[normalizedDefault]; !ok {
		return nil, fmt.Errorf("default agent runtime %q is not available", selectedDefault)
	}
	router.SetDefault(selectedDefault)
	return router, nil
}

func selectDefaultRuntime(defaultRuntime string, cfg *config.WorkerConfig, cliNames []string) string {
	if strings.TrimSpace(defaultRuntime) != "" {
		return defaultRuntime
	}
	if cfg != nil && cfg.CLI != nil && strings.TrimSpace(cfg.CLI.Default) != "" {
		return cfg.CLI.Default
	}
	if cfg != nil && cfg.LLM != nil && cfg.LLM.APIKey != "" {
		return agent.RuntimeKindSingerOS
	}
	if len(cliNames) > 0 {
		return cliNames[0]
	}
	return ""
}

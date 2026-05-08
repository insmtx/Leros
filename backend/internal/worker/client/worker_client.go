package client

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/agent"
	agentruntime "github.com/insmtx/SingerOS/backend/internal/agent/runtime"
	"github.com/ygpkg/yg-go/logs"
)

type WorkerClient struct {
	runtime   agent.AgentRuntime
	config    *WorkerConfig
	workerID  string
	startedAt time.Time
	status    string
	wsClient  *WSClient
}

type WorkerConfig struct {
	Runtime      agent.AgentRuntime
	LLMConfig    *config.LLMConfig
	SkillsDir    string
	ToolsEnabled bool
	ServerAddr   string
	WorkerID     string
}

func NewWorker(ctx context.Context, cfg *WorkerConfig) (*WorkerClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("worker config is required")
	}

	workerID := fmt.Sprintf("worker_%d", time.Now().UnixNano())

	w := &WorkerClient{
		config:    cfg,
		workerID:  workerID,
		startedAt: time.Now(),
		status:    "initialized",
	}

	if cfg.ServerAddr != "" {
		w.wsClient = NewWSClient(cfg.ServerAddr, workerID,
			WithWorkerID(workerID),
			WithOnConfigReady(func(assistantConfig map[string]interface{}) {
				w.handleAssistantConfig(ctx, assistantConfig)
			}),
		)
	}

	return w, nil
}

func (w *WorkerClient) handleAssistantConfig(ctx context.Context, assistantConfig map[string]interface{}) {
	logs.Info("Processing assistant configuration from server")

	llmConfigRaw, ok := assistantConfig["llm_config"]
	if !ok {
		logs.Warn("llm_config not found in assistant config, using default")
		return
	}

	llmConfigMap, ok := llmConfigRaw.(map[string]interface{})
	if !ok {
		logs.Warn("llm_config is not a valid object")
		return
	}

	llmConfig := &config.LLMConfig{}
	if provider, ok := llmConfigMap["type"].(string); ok {
		llmConfig.Provider = provider
	}
	if apiKey, ok := llmConfigMap["api_key"].(string); ok {
		llmConfig.APIKey = apiKey
	}
	if model, ok := llmConfigMap["model"].(string); ok {
		llmConfig.Model = model
	}
	if baseURL, ok := llmConfigMap["base_url"].(string); ok {
		llmConfig.BaseURL = baseURL
	}

	if llmConfig.Provider == "" || llmConfig.APIKey == "" {
		logs.Warn("incomplete llm_config, skipping runtime initialization")
		return
	}

	runtime, err := buildDefaultRuntime(ctx, &WorkerConfig{
		LLMConfig:    llmConfig,
		SkillsDir:    w.config.SkillsDir,
		ToolsEnabled: w.config.ToolsEnabled,
	})
	if err != nil {
		logs.Errorf("Failed to build runtime: %v", err)
		return
	}

	w.runtime = runtime
	w.status = "ready"
	logs.Infof("Worker %s initialized with assistant config", w.workerID)
}
func buildDefaultRuntime(ctx context.Context, cfg *WorkerConfig) (agent.AgentRuntime, error) {
	if cfg.LLMConfig == nil {
		return nil, fmt.Errorf("llm config is required")
	}

	runtimeService, err := agentruntime.NewService(ctx, agentruntime.Options{
		LLMConfig:    cfg.LLMConfig,
		ToolsEnabled: cfg.ToolsEnabled,
	})
	if err != nil {
		return nil, fmt.Errorf("create runtime service: %w", err)
	}
	return runtimeService, nil
}

func (w *WorkerClient) Run(ctx context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	if w == nil || w.runtime == nil {
		return nil, fmt.Errorf("worker runtime is not initialized")
	}

	w.status = "processing"
	result, err := w.runtime.Run(ctx, req)
	if err != nil {
		w.status = "error"
		return nil, err
	}

	w.status = "idle"
	return result, nil
}

func (w *WorkerClient) Start(ctx context.Context) error {
	w.status = "running"
	logs.Infof("Worker %s started", w.workerID)

	if w.wsClient != nil {
		if err := w.wsClient.Connect(ctx); err != nil {
			logs.Warnf("Failed to connect to server WebSocket: %v", err)
		} else {
			logs.Info("Connected to server via WebSocket")
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logs.Info("Worker context cancelled")
		return w.Shutdown(ctx)
	case sig := <-sigChan:
		logs.Infof("Received signal %v, shutting down", sig)
		return w.Shutdown(ctx)
	}
}

func (w *WorkerClient) Shutdown(ctx context.Context) error {
	logs.Info("Worker shutting down...")
	w.status = "stopping"

	if w.wsClient != nil {
		w.wsClient.Close()
	}

	return nil
}

func (w *WorkerClient) GetWorkerID() string {
	return w.workerID
}

func (w *WorkerClient) GetStartedAt() time.Time {
	return w.startedAt
}

func (w *WorkerClient) GetStatus() string {
	return w.status
}

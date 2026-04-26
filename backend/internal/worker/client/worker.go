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
	"github.com/insmtx/SingerOS/backend/tools"
	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
	"github.com/ygpkg/yg-go/logs"
)

type Worker struct {
	agent      *agent.Agent
	config     *WorkerConfig
	workerID   string
	startedAt  time.Time
	status     string
	wsClient   *WSClient
}

type WorkerConfig struct {
	LLMConfig    *config.LLMConfig
	SkillsDir    string
	ToolsEnabled bool
	ServerAddr   string
}

func NewWorker(ctx context.Context, cfg *WorkerConfig) (*Worker, error) {
	if cfg == nil {
		return nil, fmt.Errorf("worker config is required")
	}

	if cfg.LLMConfig == nil {
		return nil, fmt.Errorf("LLM config is required")
	}

	catalog, err := loadSkillsCatalog(cfg.SkillsDir)
	if err != nil {
		return nil, fmt.Errorf("load skills catalog: %w", err)
	}

	toolRegistry := tools.NewRegistry()

	if cfg.ToolsEnabled {
		if err := skilltools.Register(toolRegistry, catalog); err != nil {
			return nil, fmt.Errorf("register tools: %w", err)
		}
	}

	runtimeConfig := agent.Config{
		SkillsCatalog: catalog,
		ToolRegistry:  toolRegistry,
	}

	agentInstance, err := agent.NewAgent(ctx, cfg.LLMConfig, runtimeConfig)
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	workerID := fmt.Sprintf("worker_%d", time.Now().UnixNano())

	w := &Worker{
		agent:     agentInstance,
		config:    cfg,
		workerID:  workerID,
		startedAt: time.Now(),
		status:    "initialized",
	}

	if cfg.ServerAddr != "" {
		w.wsClient = NewWSClient(cfg.ServerAddr, workerID)
	}

	return w, nil
}

func (w *Worker) Run(ctx context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	if w == nil || w.agent == nil {
		return nil, fmt.Errorf("worker agent is not initialized")
	}
	
	w.status = "processing"
	result, err := w.agent.Run(ctx, req)
	if err != nil {
		w.status = "error"
		return nil, err
	}
	
	w.status = "idle"
	return result, nil
}

func (w *Worker) Start(ctx context.Context) error {
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

func (w *Worker) Shutdown(ctx context.Context) error {
	logs.Info("Worker shutting down...")
	w.status = "stopping"

	if w.wsClient != nil {
		w.wsClient.Close()
	}

	return nil
}

func (w *Worker) GetWorkerID() string {
	return w.workerID
}

func (w *Worker) GetStartedAt() time.Time {
	return w.startedAt
}

func (w *Worker) GetStatus() string {
	return w.status
}

func loadSkillsCatalog(skillsDir string) (*skilltools.Catalog, error) {
	if skillsDir == "" {
		return skilltools.NewEmptyCatalog(), nil
	}

	catalog, _, err := skilltools.LoadDefaultCatalog()
	if err != nil {
		return nil, err
	}
	return catalog, nil
}

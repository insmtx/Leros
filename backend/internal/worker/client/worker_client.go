package client

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/insmtx/Leros/backend/config"
	"github.com/insmtx/Leros/backend/internal/agent"
	"github.com/ygpkg/yg-go/logs"
)

type WorkerClient struct {
	runtime   agent.Runner
	config    *WorkerConfig
	daID      uint
	workerID  uint
	startedAt time.Time
	status    string
	wsClient  *WSClient
}

type WorkerConfig struct {
	Runtime            agent.Runner
	LLMConfig          *config.LLMConfig
	ServerAddr         string
	DigitalAssistantID uint
	WorkerID           uint
}

func NewWorker(ctx context.Context, cfg *WorkerConfig) (*WorkerClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("worker config is required")
	}

	workerID := cfg.WorkerID
	if workerID == 0 {
		workerID = cfg.DigitalAssistantID
	}

	w := &WorkerClient{
		config:    cfg,
		daID:      cfg.DigitalAssistantID,
		workerID:  workerID,
		startedAt: time.Now(),
		status:    "initialized",
	}

	// TODO: worker与server交互时重新实现 config 接收
	if cfg.ServerAddr != "" {
		w.wsClient = NewWSClient(cfg.ServerAddr, cfg.DigitalAssistantID,
			WithWorkerID(cfg.DigitalAssistantID),
		)
	}

	return w, nil
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

func (w *WorkerClient) GetWorkerID() uint {
	return w.workerID
}

func (w *WorkerClient) GetStartedAt() time.Time {
	return w.startedAt
}

func (w *WorkerClient) GetStatus() string {
	return w.status
}

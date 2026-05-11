package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/Leros/backend/config"
	agentruntime "github.com/insmtx/Leros/backend/internal/agent/runtime"
	"github.com/insmtx/Leros/backend/internal/infra/mq"
	"github.com/insmtx/Leros/backend/internal/worker/taskconsumer"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

var claudeWorkerCmd = &cobra.Command{
	Use:     "agent-worker",
	Aliases: []string{"claude-worker"},
	Short:   "Start a standalone task worker backed by available agent runtimes",
	Long:    `Start a standalone Leros worker that subscribes to org.{org_id}.worker.{worker_id}.task and executes agent.run tasks through the configured default agent runtime.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTaskWorker(workerDefaultRuntime)
	},
}

func init() {
	rootCmd.AddCommand(claudeWorkerCmd)
}

func runTaskWorker(defaultRuntime string) {
	cfg, err := loadWorkerConfig()
	if err != nil {
		logs.Fatalf("Failed to load config: %v", err)
		return
	}
	if err := validateTaskWorkerConfig(cfg); err != nil {
		logs.Fatalf("Invalid worker config: %v", err)
		return
	}

	mcpServer, err := startWorkerMCPServer(workerListenAddr)
	if err != nil {
		logs.Fatalf("Failed to start worker MCP server: %v", err)
		return
	}

	natsURL := "nats://nats:4222"
	if cfg.NATS != nil && strings.TrimSpace(cfg.NATS.URL) != "" {
		natsURL = cfg.NATS.URL
	}

	bus, err := mq.NewPublisher(natsURL)
	if err != nil {
		logs.Fatalf("Failed to create NATS client: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	runtimeService, err := agentruntime.NewService(ctx, agentruntime.Options{
		LLMConfig:      cfg.LLM,
		CLIConfig:      cfg.CLI,
		ToolsEnabled:   true,
		DefaultRuntime: defaultRuntime,
	})
	if err != nil {
		cancel()
		_ = bus.Close()
		logs.Fatalf("Failed to create agent runtime service: %v", err)
		return
	}

	consumer, err := taskconsumer.New(taskconsumer.Config{
		OrgID:    cfg.OrgID,
		WorkerID: cfg.WorkerID,
	}, bus, bus, runtimeService)
	if err != nil {
		cancel()
		_ = bus.Close()
		logs.Fatalf("Failed to create worker task consumer: %v", err)
		return
	}
	if err := consumer.Start(ctx); err != nil {
		cancel()
		_ = bus.Close()
		logs.Fatalf("Failed to start worker task consumer: %v", err)
		return
	}

	lifecycle.Std().AddCloseFunc(func() error {
		cancel()
		return nil
	})
	lifecycle.Std().AddCloseFunc(func() error {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		return mcpServer.Shutdown(shutdownCtx)
	})
	lifecycle.Std().AddCloseFunc(bus.Close)

	logs.Infof("Agent worker started: org_id=%s worker_id=%s topic=%s", cfg.OrgID, cfg.WorkerID, consumer.TaskTopic())
	lifecycle.Std().WaitExit()
	logs.Info("Agent worker exited")
}

func validateTaskWorkerConfig(cfg *config.WorkerConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	if strings.TrimSpace(cfg.AssistantCode) == "" {
		return fmt.Errorf("worker.assistant_code is required")
	}
	return nil
}

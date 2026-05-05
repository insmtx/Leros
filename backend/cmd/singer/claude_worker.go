package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/internal/agent/externalcli"
	"github.com/insmtx/SingerOS/backend/internal/infra/mq"
	"github.com/insmtx/SingerOS/backend/internal/worker/taskconsumer"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
	"github.com/insmtx/SingerOS/backend/runtime/engines/builtin"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

var (
	taskWorkerConfigPath string
	taskWorkerServerAddr string
)

var claudeWorkerCmd = &cobra.Command{
	Use:   "claude-worker",
	Short: "Start a standalone task worker backed by Claude Code",
	Long:  `Start a standalone SingerOS worker that subscribes to org.{org_id}.worker.{worker_id}.task and executes agent.run tasks through Claude Code.`,
	Run: func(cmd *cobra.Command, args []string) {
		mcpServer, err := startWorkerMCPServer(taskWorkerServerAddr)
		if err != nil {
			logs.Fatalf("Failed to start worker MCP server: %v", err)
			return
		}

		cfg, err := loadWorkerConfig(taskWorkerConfigPath, taskWorkerServerAddr)
		if err != nil {
			logs.Fatalf("Failed to load config: %v", err)
			return
		}
		if err := validateTaskWorkerConfig(cfg); err != nil {
			logs.Fatalf("Invalid worker config: %v", err)
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

		runner, err := buildClaudeCodeRunner(cfg)
		if err != nil {
			_ = bus.Close()
			logs.Fatalf("Failed to create Claude Code runtime: %v", err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		consumer, err := taskconsumer.New(taskconsumer.Config{
			OrgID:    cfg.Worker.OrgID,
			WorkerID: cfg.Worker.WorkerID,
		}, bus, bus, runner)
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

		logs.Infof("Claude worker started: org_id=%s worker_id=%s topic=%s", cfg.Worker.OrgID, cfg.Worker.WorkerID, consumer.TaskTopic())
		lifecycle.Std().WaitExit()
		logs.Info("Claude worker exited")
	},
}

func init() {
	claudeWorkerCmd.Flags().StringVar(&taskWorkerConfigPath, "config", "", "Configuration file path")
	claudeWorkerCmd.Flags().StringVar(&taskWorkerServerAddr, "server-addr", ":8081", "Worker MCP server listen address for runtime bootstrap")
	rootCmd.AddCommand(claudeWorkerCmd)
}

func validateTaskWorkerConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	if cfg.Worker == nil {
		return fmt.Errorf("worker config is required")
	}
	if strings.TrimSpace(cfg.Worker.OrgID) == "" {
		return fmt.Errorf("worker.org_id is required")
	}
	if strings.TrimSpace(cfg.Worker.WorkerID) == "" {
		return fmt.Errorf("worker.worker_id is required")
	}
	return nil
}

func buildClaudeCodeRunner(cfg *config.Config) (agent.Runner, error) {
	cliRegistry, err := builtin.NewRegistryFromConfig(cfg.CLI)
	if err != nil {
		return nil, fmt.Errorf("create CLI engine registry: %w", err)
	}
	claudeEngine, ok := cliRegistry.Get(engines.EngineClaude)
	if !ok {
		return nil, fmt.Errorf("Claude Code CLI is not available; install or expose the claude binary in PATH")
	}

	claudeRunner, err := externalcli.NewRunner(engines.EngineClaude, claudeEngine, cfg.LLM)
	if err != nil {
		return nil, err
	}

	router := agent.NewRuntimeRouter(engines.EngineClaude)
	if err := router.Register(engines.EngineClaude, claudeRunner); err != nil {
		return nil, err
	}
	router.SetDefault(engines.EngineClaude)
	return router, nil
}

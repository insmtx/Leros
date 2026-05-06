package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/infra/mq"
	"github.com/insmtx/SingerOS/backend/internal/worker/taskconsumer"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

var (
	taskWorkerConfigPath     string
	taskWorkerServerAddr     string
	taskWorkerDefaultRuntime string
)

var claudeWorkerCmd = &cobra.Command{
	Use:     "agent-worker",
	Aliases: []string{"claude-worker"},
	Short:   "Start a standalone task worker backed by available agent runtimes",
	Long:    `Start a standalone SingerOS worker that subscribes to org.{org_id}.worker.{worker_id}.task and executes agent.run tasks through the configured default agent runtime.`,
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

		runtimeConfig, err := buildRuntimeConfig()
		if err != nil {
			_ = bus.Close()
			logs.Fatalf("Failed to build runtime config: %v", err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		runner, err := buildRuntimeRunner(ctx, cfg, runtimeConfig, taskWorkerDefaultRuntime)
		if err != nil {
			cancel()
			_ = bus.Close()
			logs.Fatalf("Failed to create agent runtime: %v", err)
			return
		}

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

		logs.Infof("Agent worker started: org_id=%s worker_id=%s topic=%s", cfg.Worker.OrgID, cfg.Worker.WorkerID, consumer.TaskTopic())
		lifecycle.Std().WaitExit()
		logs.Info("Agent worker exited")
	},
}

func init() {
	claudeWorkerCmd.Flags().StringVar(&taskWorkerConfigPath, "config", "", "Configuration file path")
	claudeWorkerCmd.Flags().StringVar(&taskWorkerServerAddr, "server-addr", ":8081", "Worker MCP server listen address for runtime bootstrap")
	claudeWorkerCmd.Flags().StringVar(&taskWorkerDefaultRuntime, "default-runtime", engines.EngineClaude, "Default agent runtime kind, for example singeros, claude, or codex")
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

package main

import (
	"context"
	"fmt"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/worker/client"
	"github.com/spf13/cobra"
	ygconfig "github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
)

var (
	workerConfigPath    string
	workerServerAddr    string
	workerAssistantCode string
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
	workerCmd.Flags().StringVar(&workerServerAddr, "server-addr", ":8080", "Server address for WebSocket connection")
	workerCmd.Flags().StringVar(&workerAssistantCode, "assistant-code", "", "Assistant code for configuration retrieval")
	rootCmd.AddCommand(workerCmd)
}

func createWorker(ctx context.Context) (*client.Worker, error) {
	_, err := loadWorkerConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return client.NewWorker(ctx, &client.WorkerConfig{
		ServerAddr:    workerServerAddr,
		AssistantCode: workerAssistantCode,
		SkillsDir:     "",
		ToolsEnabled:  true,
	})
}

func loadWorkerConfig() (*config.Config, error) {
	if workerConfigPath != "" {
		cfg := &config.Config{}
		err := ygconfig.LoadYamlLocalFile(workerConfigPath, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", workerConfigPath, err)
		}
		return cfg, nil
	}

	return &config.Config{}, nil
}

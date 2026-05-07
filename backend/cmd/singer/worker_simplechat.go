package main

import (
	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/spf13/cobra"
)

var simpleChatDefaultRuntime string

var simpleChatCmd = &cobra.Command{
	Use:   "simplechat",
	Short: "Start a standalone task worker backed by the built-in SingerOS runtime",
	Long:  `Start a standalone SingerOS worker that subscribes to org.{org_id}.worker.{worker_id}.task and executes agent.run tasks through the built-in SingerOS agent runtime.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTaskWorker(simpleChatDefaultRuntime)
	},
}

func init() {
	simpleChatCmd.Flags().StringVar(&workerConfigPath, "config", "", "Configuration file path")
	simpleChatCmd.Flags().StringVar(&workerServerAddr, "server-addr", "127.0.0.1:8080", "Server address for WebSocket connection")
	simpleChatCmd.Flags().StringVar(&workerListenAddr, "listen-addr", ":8081", "Worker MCP server listen address for runtime bootstrap")
	simpleChatCmd.Flags().StringVar(&workerWorkerID, "worker-id", "", "Worker ID for configuration retrieval")
	simpleChatCmd.Flags().StringVar(&simpleChatDefaultRuntime, "default-runtime", agent.RuntimeKindSingerOS, "Default agent runtime kind, for example singeros, claude, or codex")

	workerCmd.AddCommand(simpleChatCmd)
}

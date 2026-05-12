package main

import (
	"github.com/spf13/cobra"
)

var claudeWorkerCmd = &cobra.Command{
	Use:   "claude-worker",
	Short: "Start a standalone task worker backed by available agent runtimes",
	Long:  `Start a standalone Leros worker that subscribes to org.{org_id}.worker.{worker_id}.task and executes agent.run tasks through the configured default agent runtime.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTaskWorker(workerDefaultRuntime)
	},
}

func init() {
	workerCmd.AddCommand(claudeWorkerCmd)
}

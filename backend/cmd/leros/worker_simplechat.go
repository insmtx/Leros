package main

import (
	"github.com/spf13/cobra"
)

var simpleChatDefaultRuntime string

var simpleChatCmd = &cobra.Command{
	Use:   "simplechat",
	Short: "Start a standalone task worker backed by the built-in Leros runtime",
	Long:  `Start a standalone Leros worker that subscribes to org.{org_id}.worker.{worker_id}.task and executes agent.run tasks through the built-in Leros agent runtime.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTaskWorker(simpleChatDefaultRuntime)
	},
}

func init() {
	workerCmd.AddCommand(simpleChatCmd)
}

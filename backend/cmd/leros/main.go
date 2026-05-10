package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/logs"
)

var rootCmd = &cobra.Command{
	Use:   "leros",
	Short: "Backend service for the Leros Backend",
	Long:  `This is the backend service for the Leros Backend, responsible for handling API requests and business logic.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logs.Errorf("Error executing command: %v", err)
		os.Exit(1)
	}
}

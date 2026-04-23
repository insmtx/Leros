package main

import (
	"os"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/logs"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "singer",
	Short: "Backend service for the SingerOS Backend",
	Long:  `This is the backend service for the SingerOS Backend, responsible for handling API requests and business logic.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Configuration file path")
}

func loadConfig() (*config.Config, error) {
	return config.Load(configPath)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logs.Errorf("Error executing command: %v", err)
		os.Exit(1)
	}
}

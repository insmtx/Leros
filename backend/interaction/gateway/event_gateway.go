package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/interaction"
	"github.com/insmtx/SingerOS/backend/interaction/connectors/github"
	"github.com/insmtx/SingerOS/backend/interaction/eventbus"
	"github.com/ygpkg/yg-go/logs"
)

func SetupRouter(r gin.IRouter, cfg config.Config, publisher eventbus.Publisher) {
	registry := interaction.NewRegistry()

	// Check if GitHub configuration is provided and enabled
	if cfg.Github != nil {
		logs.Info("Setting up GitHub connector")
		githubConnector := github.NewConnector(*cfg.Github, publisher)
		registry.Register(githubConnector)
		logs.Info("GitHub connector registered successfully")
	} else {
		logs.Debug("No GitHub configuration provided, skipping GitHub connector setup")
	}

	registry.RegisterRoutes(r)
	logs.Info("Event gateway routes registered successfully")
}

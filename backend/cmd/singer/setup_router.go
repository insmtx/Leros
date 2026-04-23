package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/SingerOS/backend/auth"
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/connectors"
	"github.com/insmtx/SingerOS/backend/internal/connectors/client"
	githubconn "github.com/insmtx/SingerOS/backend/internal/connectors/github"
	"github.com/insmtx/SingerOS/backend/internal/connectors/gitlab"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, event interface{}) error
	Close() error
}

func SetupRouter(r gin.IRouter, cfg config.Config, publisher EventPublisher, db *gorm.DB, authService *auth.Service) {
	registry := connectors.NewRegistry()

	if cfg.Github != nil {
		logs.Info("Setting up GitHub connector")
		githubConnector := githubconn.NewConnector(*cfg.Github, publisher, db, authService)
		registry.Register(githubConnector)
		logs.Info("GitHub connector registered successfully")
	}

	if cfg.Gitlab != nil {
		logs.Info("Setting up GitLab connector")
		gitlabConnector := gitlab.NewConnector(*cfg.Gitlab, publisher)
		registry.Register(gitlabConnector)
		logs.Info("GitLab connector registered successfully")
	}

	clientConnector := client.NewConnector(publisher)
	registry.Register(clientConnector)
	logs.Info("Client WebSocket connector registered successfully")

	registry.RegisterRoutes(r)
	logs.Info("Routes registered successfully")
}

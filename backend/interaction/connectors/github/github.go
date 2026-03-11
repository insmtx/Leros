package github

import (
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v78/github"
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/interaction/connectors"
	"github.com/insmtx/SingerOS/backend/interaction/eventbus"
	"github.com/ygpkg/yg-go/logs"
)

var _ connectors.Connector = (*GitHubConnector)(nil)

type GitHubConnector struct {
	config config.GithubAppConfig

	client *github.Client

	publisher eventbus.Publisher
}

func (GitHubConnector) ChannelCode() string {
	return "github"
}

func (c *GitHubConnector) RegisterRoutes(r gin.IRouter) {
	r.POST("/github/webhook", c.HandleWebhook)
}

func NewConnector(cfg config.GithubAppConfig, publisher eventbus.Publisher) *GitHubConnector {
	logs.Infof("Creating new GitHub connector for app ID: %d", cfg.AppID)

	// Initialize GitHub client if credentials are provided
	var githubClient *github.Client

	if cfg.AppID != 0 && cfg.PrivateKey != "" {
		// In a real implementation, we would use github.NewTokenClient or initialize an authenticated client
		// For now we just note it in logs
		logs.Debugf("GitHub connector initialized with app ID: %d", cfg.AppID)
	} else {
		logs.Warnf("GitHub connector initialized without authentication - limited functionality")
	}

	return &GitHubConnector{
		config:    cfg,
		client:    githubClient, // Would be properly initialized in a complete impl
		publisher: publisher,
	}
}

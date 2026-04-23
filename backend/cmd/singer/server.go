package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/SingerOS/backend/database"
	"github.com/insmtx/SingerOS/backend/internal/eventengine"
	"github.com/insmtx/SingerOS/backend/internal/execution"
	"github.com/insmtx/SingerOS/backend/internal/infra/mq/rabbitmq"
	"github.com/insmtx/SingerOS/backend/internal/service"
	ygconfig "github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/apis/runtime/middleware"
	"gorm.io/gorm"
)

var serverAddr string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the SingerOS control plane server",
	Long:  `Start the HTTP server, connectors, and event engine for handling external events.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := loadConfig()
		if err != nil {
			logs.Fatalf("Failed to load config: %v", err)
			return
		}

		rmqUrl := "amqp://singer_user:singer_password@rabbitmq:5672/"
		if cfg.RabbitMQ != nil && cfg.RabbitMQ.URL != "" {
			rmqUrl = cfg.RabbitMQ.URL
		}

		rmqCfg := ygconfig.RabbitMQConfig{URL: rmqUrl}
		publisher, _, err := rabbitmq.NewPublisher(rmqCfg)
		if err != nil {
			logs.Fatalf("Failed to create event publisher: %v", err)
			return
		}

		if cfg.LLM == nil || cfg.LLM.APIKey == "" {
			logs.Fatalf("LLM configuration is required for Eino runtime")
			return
		}

		authService := service.NewAuthService(cfg)

		executionEngine := execution.NewExecutionEngine()
		eventEngine := eventengine.NewEventEngine(publisher, executionEngine)

		var db *gorm.DB
		if cfg.Database != nil && cfg.Database.URL != "" {
			db, err = database.InitDB(*cfg.Database)
			if err != nil {
				logs.Fatalf("Failed to initialize database: %v", err)
				return
			}
			logs.Info("Database initialized successfully")
		} else {
			logs.Warn("No database configuration provided.")
			logs.Warn("  To enable database, add database.url to your config file.")
		}

		r := gin.New()
		{
			r.Use(middleware.CORS())
			r.Use(middleware.CustomerHeader())
			r.Use(middleware.Logger(".Ping", "metrics"))
			r.Use(middleware.Recovery())
		}

		service.SetupRouter(r, *cfg, publisher, db, authService)

		srv := &http.Server{
			Addr:    serverAddr,
			Handler: r,
		}

		logs.Info("Starting SingerOS backend service...")
		logs.Infof("Listening on %s", serverAddr)

		ctx := context.Background()
		if err := eventEngine.Start(ctx); err != nil {
			logs.Errorf("Failed to start Event Engine: %v", err)
		} else {
			logs.Info("Event Engine started successfully")
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logs.Fatalf("Failed to start server: %v", err)
			}
		}()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logs.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logs.Errorf("Server forced to shutdown: %v", err)
		}

		if publisher != nil {
			publisher.Close()
		}

		logs.Info("Server exited")
	},
}

func init() {
	serverCmd.Flags().StringVar(&serverAddr, "addr", ":8080", "HTTP server address")
	rootCmd.AddCommand(serverCmd)
}

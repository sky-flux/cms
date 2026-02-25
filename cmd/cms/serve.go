package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sky-flux/cms/internal/config"
	"github.com/sky-flux/cms/internal/cron"
	"github.com/sky-flux/cms/internal/database"
	"github.com/sky-flux/cms/internal/router"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动 HTTP 服务",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "", "服务端口 (覆盖 SERVER_PORT)")
	serveCmd.Flags().String("mode", "", "运行模式: debug/release (覆盖 SERVER_MODE)")
	_ = viper.BindPFlag("SERVER_PORT", serveCmd.Flags().Lookup("port"))
	_ = viper.BindPFlag("SERVER_MODE", serveCmd.Flags().Lookup("mode"))
}

func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	initLogger(cfg)

	slog.Info("starting server", "port", cfg.Server.Port, "mode", cfg.Server.Mode)

	// Connect to PostgreSQL
	db, err := database.NewPostgres(cfg)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer db.Close()
	slog.Info("postgres connected")

	// Connect to Redis
	rdb, err := database.NewRedis(cfg)
	if err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}
	defer rdb.Close()
	slog.Info("redis connected")

	// Connect to Meilisearch (graceful degradation)
	meili, err := database.NewMeilisearch(cfg)
	if err != nil {
		slog.Warn("meilisearch not available, search features disabled", "error", err)
	} else {
		slog.Info("meilisearch connected")
	}

	// Connect to RustFS (graceful degradation)
	var s3Client *s3.Client
	s3Client, err = database.NewRustFS(cfg)
	if err != nil {
		slog.Warn("rustfs not available, media features disabled", "error", err)
	} else {
		slog.Info("rustfs connected")
	}

	// Setup Gin engine
	if cfg.Server.Mode != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	router.Setup(engine, db, rdb, meili, s3Client, cfg)

	// HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server listen error", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("server started", "addr", srv.Addr)

	// Start cron scheduler
	cronScheduler := cron.NewScheduler(cron.Deps{
		Sites:     cron.NewBunSiteLister(db),
		Schema:    cron.NewBunSchemaExecutor(db),
		Publisher: cron.NewBunScheduledPublisher(db),
		Cleaner:   cron.NewBunTokenCleaner(db),
		Purger:    cron.NewBunSoftDeletePurger(db),
	})
	cronScheduler.Start()

	// Wait for interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("shutting down server...")
	cronScheduler.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("server stopped")
	return nil
}

func initLogger(cfg *config.Config) {
	var level slog.Level
	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	if cfg.Server.Mode == "debug" {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      level,
			TimeFormat: "15:04:05",
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	slog.SetDefault(slog.New(handler))
}

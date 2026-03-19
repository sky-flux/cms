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

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"

	"github.com/sky-flux/cms/internal/config"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the CMS HTTP server",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringP("port", "p", "", "HTTP listen port (overrides SERVER_PORT env)")
	serveCmd.Flags().String("mode", "", "Server mode: debug|release (overrides SERVER_MODE env)")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, _ []string) error {
	cfgFilePath, _ := cmd.Root().PersistentFlags().GetString("config")
	cfg, err := config.Load(cfgFilePath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// CLI flag overrides
	if p, _ := cmd.Flags().GetString("port"); p != "" {
		cfg.Server.Port = p
	}
	if m, _ := cmd.Flags().GetString("mode"); m != "" {
		cfg.Server.Mode = m
	}

	handler := newServer()

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting", "addr", srv.Addr, "mode", cfg.Server.Mode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// newServer constructs the Chi router and Huma API instance.
// It is a separate function (not a method) so tests can call it
// without starting a real listener.
func newServer() http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)

	// Huma API on /api/v1
	api := humachi.New(r, huma.DefaultConfig("Sky Flux CMS API", "1.0.0"))

	// Health check — registered directly on Huma so it appears in OpenAPI spec
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
		Tags:        []string{"system"},
	}, func(ctx context.Context, _ *struct{}) (*struct {
		Body struct {
			Status string `json:"status"`
		}
	}, error) {
		resp := &struct {
			Body struct {
				Status string `json:"status"`
			}
		}{}
		resp.Body.Status = "ok"
		return resp, nil
	})

	return r
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

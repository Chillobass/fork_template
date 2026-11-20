package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Chillobass/fork_template/internal/config"
	"github.com/Chillobass/fork_template/internal/container"
	"github.com/Chillobass/fork_template/internal/infra/telemetry"
)

const serviceName = "trip-service"

func main() {
	cfg, err := config.Load()
	if err != nil {
		telemetry.NewLogger(serviceName, slog.LevelInfo).Error("configuration failed", "error", err)
		os.Exit(1)
	}
	logger := telemetry.NewLogger(serviceName, cfg.LogLevel)
	if err := run(cfg, logger); err != nil {
		logger.Error("service stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := container.New(ctx, cfg, serviceName, logger)
	if err != nil {
		return err
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server started", "addr", cfg.HTTP.Addr)
		serverErr <- app.Server.ListenAndServe()
	}()

	select {
	case err = <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			err = fmt.Errorf("http server: %w", err)
		} else {
			err = nil
		}
	case <-ctx.Done():
	}

	stop()
	logger.Info("shutdown started", "timeout", cfg.ShutdownTimeout.String())
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	err = errors.Join(err, app.Shutdown(shutdownCtx, logger))
	if err == nil {
		logger.Info("service stopped")
	}
	return err
}

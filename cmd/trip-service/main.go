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
	"github.com/Chillobass/fork_template/internal/httpapi"
	"github.com/Chillobass/fork_template/internal/postgres"
	"github.com/Chillobass/fork_template/internal/repository"
	"github.com/Chillobass/fork_template/internal/service"
)

const serviceName = "trip-service"

func main() {
	if err := run(); err != nil {
		slog.Error("service stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})).
		With("service", serviceName)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}

	txManager := postgres.NewTxManager(pool, cfg.Database.QueryTimeout)
	tripRepo := repository.NewTripRepository(pool, cfg.Database.QueryTimeout)
	historyRepo := repository.NewStatusHistoryRepository(pool, cfg.Database.QueryTimeout)
	tripService := service.NewTripService(txManager, tripRepo, historyRepo)
	handler := httpapi.NewHandler(tripService, pool, cfg.Database.QueryTimeout, logger)

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           httpapi.NewRouter(handler, logger),
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server started", "addr", cfg.HTTP.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		pool.Close()
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	stop()
	logger.Info("shutdown started", "timeout", cfg.ShutdownTimeout.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown timed out, closing connections forcibly", "error", err)
		_ = server.Close()
	}

	pool.Close()
	logger.Info("service stopped")

	return nil
}

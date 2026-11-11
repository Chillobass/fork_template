package container

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Chillobass/fork_template/internal/config"
	transport "github.com/Chillobass/fork_template/internal/http"
	"github.com/Chillobass/fork_template/internal/http/health"
	"github.com/Chillobass/fork_template/internal/http/middleware"
	httpcreate "github.com/Chillobass/fork_template/internal/http/trip/create"
	httpfinish "github.com/Chillobass/fork_template/internal/http/trip/finish"
	httpget "github.com/Chillobass/fork_template/internal/http/trip/get"
	"github.com/Chillobass/fork_template/internal/infra/postgres"
	"github.com/Chillobass/fork_template/internal/infra/postgres/transaction"
	"github.com/Chillobass/fork_template/internal/infra/telemetry"
	repository "github.com/Chillobass/fork_template/internal/repository/trip"
	"github.com/Chillobass/fork_template/internal/usecase/trip/create"
	"github.com/Chillobass/fork_template/internal/usecase/trip/finish"
	"github.com/Chillobass/fork_template/internal/usecase/trip/get"
)

type App struct {
	Server *http.Server
	pool   *pgxpool.Pool
	otel   telemetry.Providers
}

func New(ctx context.Context, cfg config.Config, service string, logger *slog.Logger) (_ *App, err error) {
	providers := telemetry.New(ctx, service, logger)
	defer func() {
		if err != nil {
			cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
			defer cancel()
			if shutdownErr := providers.Shutdown(cleanup); shutdownErr != nil {
				logger.WarnContext(cleanup, "telemetry shutdown failed", "error", shutdownErr)
			}
		}
	}()

	meter := providers.Metrics.Meter(service)
	tracer, err := postgres.NewTracer(providers.Traces, meter)
	if err != nil {
		return nil, err
	}
	httpMetrics, err := middleware.New(logger, providers.Traces, meter)
	if err != nil {
		return nil, err
	}
	pool, err := postgres.NewPool(ctx, cfg.Database, tracer)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	tx := transaction.NewTxManager(pool, cfg.Database.QueryTimeout)
	trips := repository.New(tx, cfg.Database.QueryTimeout)
	events, err := telemetry.NewTrips(meter, trips, logger)
	if err != nil {
		pool.Close()
		return nil, err
	}
	handlers := &transport.Server{
		CreateHandler: httpcreate.New(create.New(tx, trips, trips, time.Now), logger, events.Created),
		GetHandler:    httpget.New(get.New(trips)),
		FinishHandler: httpfinish.New(finish.New(tx, trips, trips, time.Now), logger, events.Completed),
		HealthHandler: health.New(pool, cfg.Database.QueryTimeout, logger),
	}
	return &App{
		Server: &http.Server{
			Addr:              cfg.HTTP.Addr,
			Handler:           transport.NewRouter(handlers, httpMetrics),
			ReadTimeout:       cfg.HTTP.ReadTimeout,
			ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
			WriteTimeout:      cfg.HTTP.WriteTimeout,
			IdleTimeout:       cfg.HTTP.IdleTimeout,
			ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
		},
		pool: pool,
		otel: providers,
	}, nil
}

func (a *App) Shutdown(ctx context.Context, logger *slog.Logger) error {
	err := a.Server.Shutdown(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "graceful HTTP shutdown failed, closing connections", "error", err)
		err = errors.Join(err, a.Server.Close())
	}
	if telemetryErr := a.otel.Shutdown(ctx); telemetryErr != nil {
		logger.WarnContext(ctx, "telemetry shutdown failed", "error", telemetryErr)
	}
	closed := make(chan struct{})
	go func() {
		a.pool.Close()
		close(closed)
	}()
	select {
	case <-closed:
		return err
	case <-ctx.Done():
		return errors.Join(err, fmt.Errorf("close database pool: %w", ctx.Err()))
	}
}

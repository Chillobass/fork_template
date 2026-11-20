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
	"github.com/Chillobass/fork_template/internal/gateway/push"
	transport "github.com/Chillobass/fork_template/internal/http"
	"github.com/Chillobass/fork_template/internal/http/health"
	"github.com/Chillobass/fork_template/internal/http/middleware"
	httpposadd "github.com/Chillobass/fork_template/internal/http/position/add"
	httpposlist "github.com/Chillobass/fork_template/internal/http/position/list"
	httpcreate "github.com/Chillobass/fork_template/internal/http/trip/create"
	httpfinish "github.com/Chillobass/fork_template/internal/http/trip/finish"
	httpget "github.com/Chillobass/fork_template/internal/http/trip/get"
	"github.com/Chillobass/fork_template/internal/infra/postgres"
	"github.com/Chillobass/fork_template/internal/infra/postgres/transaction"
	"github.com/Chillobass/fork_template/internal/infra/telemetry"
	repository "github.com/Chillobass/fork_template/internal/repository/trip"
	posadd "github.com/Chillobass/fork_template/internal/usecase/position/add"
	poslist "github.com/Chillobass/fork_template/internal/usecase/position/list"
	"github.com/Chillobass/fork_template/internal/usecase/trip/create"
	"github.com/Chillobass/fork_template/internal/usecase/trip/finish"
	"github.com/Chillobass/fork_template/internal/usecase/trip/get"
	"github.com/Chillobass/fork_template/internal/worker/positions"
)

type App struct {
	Server *http.Server
	poller *positions.Poller
	push   *push.Client
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
	defer func() {
		if err != nil {
			pool.Close()
		}
	}()

	tx := transaction.NewTxManager(pool, cfg.Database.QueryTimeout)
	trips := repository.New(tx, cfg.Database.QueryTimeout)

	events, err := telemetry.NewTrips(meter, trips, logger)
	if err != nil {
		return nil, err
	}
	posMetrics, err := telemetry.NewPositions(meter)
	if err != nil {
		return nil, err
	}

	pushClient := push.New(cfg.Push.HTTPURL, cfg.Push.RequestTimeout, providers.Traces)
	poller, err := positions.New(
		positions.Options{
			CheckInterval: cfg.Positions.CheckInterval,
			StaleAfter:    cfg.Positions.StaleAfter,
			Workers:       cfg.Positions.Workers,
			QueueSize:     cfg.Positions.QueueSize,
			ScanBatchSize: cfg.Positions.ScanBatchSize,
			PushTimeout:   cfg.Push.RequestTimeout,
		},
		trips,
		pushClient,
		logger,
		providers.Traces,
		meter,
		positions.NewRealTicker,
		time.Now,
	)
	if err != nil {
		pushClient.Close()
		return nil, err
	}

	poller.Start(context.WithoutCancel(ctx))

	handlers := &transport.Server{
		CreateHandler:       httpcreate.New(create.New(tx, trips, trips, time.Now), logger, events.Created),
		GetHandler:          httpget.New(get.New(trips)),
		FinishHandler:       httpfinish.New(finish.New(tx, trips, trips, time.Now), logger, events.Completed),
		AddPositionHandler:  httpposadd.New(posadd.New(tx, trips, trips, time.Now), poller, posMetrics.Saved),
		ListPositionHandler: httpposlist.New(poslist.New(trips, trips)),
		HealthHandler:       health.New(pool, cfg.Database.QueryTimeout, logger),
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
		poller: poller,
		push:   pushClient,
		pool:   pool,
		otel:   providers,
	}, nil
}

func (a *App) Shutdown(ctx context.Context, logger *slog.Logger) error {
	err := a.Server.Shutdown(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "graceful HTTP shutdown failed, closing connections", "error", err)
		err = errors.Join(err, a.Server.Close())
	}

	if pollerErr := a.poller.Shutdown(ctx); pollerErr != nil {
		logger.ErrorContext(ctx, "positions poller shutdown failed", "error", pollerErr)
		err = errors.Join(err, pollerErr)
	}

	if telemetryErr := a.otel.Shutdown(ctx); telemetryErr != nil {
		logger.WarnContext(ctx, "telemetry shutdown failed", "error", telemetryErr)
	}

	a.push.Close()

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

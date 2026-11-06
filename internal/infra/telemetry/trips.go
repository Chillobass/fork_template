package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/metric"
)

type activeTripCounter interface {
	CountActive(context.Context) (int64, error)
}

type Trips struct {
	created   metric.Int64Counter
	completed metric.Int64Counter
}

func NewTrips(meter metric.Meter, repo activeTripCounter, logger *slog.Logger) (Trips, error) {
	created, err := meter.Int64Counter("tripgo.trips.created", metric.WithUnit("{trip}"))
	if err != nil {
		return Trips{}, fmt.Errorf("register created trips: %w", err)
	}
	completed, err := meter.Int64Counter("tripgo.trips.completed", metric.WithUnit("{trip}"))
	if err != nil {
		return Trips{}, fmt.Errorf("register completed trips: %w", err)
	}
	_, err = meter.Int64ObservableGauge("tripgo.trips.active", metric.WithUnit("{trip}"),
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			count, err := repo.CountActive(ctx)
			if err != nil {
				logger.WarnContext(ctx, "could not measure active trips", "error", err)
				return nil
			}
			observer.Observe(count)
			return nil
		}))
	if err != nil {
		return Trips{}, fmt.Errorf("register active trips: %w", err)
	}
	return Trips{created: created, completed: completed}, nil
}

func (t Trips) Created(ctx context.Context) {
	t.created.Add(ctx, 1)
}

func (t Trips) Completed(ctx context.Context) {
	t.completed.Add(ctx, 1)
}

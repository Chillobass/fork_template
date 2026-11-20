package positions

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	resultQueued    = "queued"
	resultDuplicate = "duplicate"
	resultDropped   = "dropped"
	resultSuccess   = "success"
	resultError     = "error"
	resultCancelled = "cancelled"
)

type metrics struct {
	staleChecks metric.Int64Counter
	rowsRead    metric.Int64Histogram
	requests    metric.Int64Counter
	duration    metric.Float64Histogram
}

func newMetrics(meter metric.Meter, queueSize, activeWorkers func() int64) (*metrics, error) {
	staleChecks, err := meter.Int64Counter("tripgo.stale_checks", metric.WithUnit("{check}"))
	if err != nil {
		return nil, fmt.Errorf("register stale checks counter: %w", err)
	}
	rowsRead, err := meter.Int64Histogram(
		"tripgo.stale_checks.rows_read",
		metric.WithUnit("{row}"),
		metric.WithExplicitBucketBoundaries(0, 1, 10, 50, 100, 250, 500, 1000, 5000),
	)
	if err != nil {
		return nil, fmt.Errorf("register stale checks rows_read histogram: %w", err)
	}
	requests, err := meter.Int64Counter("tripgo.position_requests", metric.WithUnit("{request}"))
	if err != nil {
		return nil, fmt.Errorf("register position requests counter: %w", err)
	}
	duration, err := meter.Float64Histogram(
		"tripgo.position_requests.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 1, 2.5, 5),
	)
	if err != nil {
		return nil, fmt.Errorf("register position requests duration histogram: %w", err)
	}

	_, err = meter.Int64ObservableGauge(
		"tripgo.position_requests.queue.size",
		metric.WithUnit("{request}"),
		metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			observer.Observe(queueSize())
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("register queue size gauge: %w", err)
	}

	_, err = meter.Int64ObservableGauge(
		"tripgo.position_requests.workers.active",
		metric.WithUnit("{worker}"),
		metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			observer.Observe(activeWorkers())
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("register active workers gauge: %w", err)
	}

	return &metrics{
		staleChecks: staleChecks,
		rowsRead:    rowsRead,
		requests:    requests,
		duration:    duration,
	}, nil
}

func (m *metrics) recordStaleCheck(ctx context.Context, result string) {
	m.staleChecks.Add(ctx, 1, metric.WithAttributeSet(attribute.NewSet(attribute.String("result", result))))
}

func (m *metrics) recordRowsRead(ctx context.Context, rows int64) {
	m.rowsRead.Record(ctx, rows)
}

func (m *metrics) recordRequest(ctx context.Context, result string, dur time.Duration) {
	attrs := metric.WithAttributeSet(attribute.NewSet(attribute.String("result", result)))
	m.requests.Add(ctx, 1, attrs)
	m.duration.Record(ctx, dur.Seconds(), attrs)
}

func (m *metrics) recordCancelled(ctx context.Context) {
	m.requests.Add(ctx, 1, metric.WithAttributeSet(attribute.NewSet(attribute.String("result", resultCancelled))))
}

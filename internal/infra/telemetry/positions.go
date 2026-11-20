package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

type Positions struct {
	saved metric.Int64Counter
}

func NewPositions(meter metric.Meter) (Positions, error) {
	saved, err := meter.Int64Counter("tripgo.positions.saved", metric.WithUnit("{position}"))
	if err != nil {
		return Positions{}, fmt.Errorf("register saved positions: %w", err)
	}
	return Positions{saved: saved}, nil
}

func (p Positions) Saved(ctx context.Context) {
	p.saved.Add(ctx, 1)
}

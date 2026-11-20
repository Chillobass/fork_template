//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package add

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type txManager interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}

type tripRepository interface {
	GetForUpdate(ctx context.Context, id uuid.UUID) (trip.Trip, error)
	UpdateLastPositionAt(ctx context.Context, tripID uuid.UUID, recordedAt, updatedAt time.Time) error
}

type positionRepository interface {
	AddPosition(ctx context.Context, tripID uuid.UUID, data trip.PositionData) (trip.Position, error)
}

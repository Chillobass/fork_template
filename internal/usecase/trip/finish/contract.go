//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package finish

import (
	"context"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/model"
)

type txManager interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}

type historyRepository interface {
	Add(ctx context.Context, change model.StatusChange) error
}

type tripRepository interface {
	GetForUpdate(ctx context.Context, id uuid.UUID) (trip.Trip, error)
	Update(ctx context.Context, trip trip.Trip) error
}

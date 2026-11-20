//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package add

import (
	"context"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type usecase interface {
	Add(ctx context.Context, tripID uuid.UUID, input trip.PositionData) (trip.Position, error)
}

type requestCanceller interface {
	Cancel(ctx context.Context, tripID uuid.UUID) bool
}

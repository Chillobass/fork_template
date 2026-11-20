//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package positions

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/model"
)

type staleTripsRepository interface {
	ListStale(ctx context.Context, staleBefore time.Time, after *model.StaleCursor, limit int) ([]model.StaleTrip, error)
}

type pushGateway interface {
	RequestPosition(ctx context.Context, tripID, driverID uuid.UUID) error
}

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

package add

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type Usecase struct {
	tx        txManager
	trips     tripRepository
	positions positionRepository
	now       func() time.Time
}

func New(tx txManager, trips tripRepository, positions positionRepository, now func() time.Time) *Usecase {
	return &Usecase{tx: tx, trips: trips, positions: positions, now: now}
}

func (u *Usecase) Add(ctx context.Context, tripID uuid.UUID, input trip.PositionData) (trip.Position, error) {
	now := u.now().UTC().Truncate(time.Microsecond)
	input.RecordedAt = input.RecordedAt.UTC().Truncate(time.Microsecond)
	if err := input.Validate(now); err != nil {
		return trip.Position{}, fmt.Errorf("validate position: %w", err)
	}

	var saved trip.Position
	err := u.tx.Do(ctx, func(ctx context.Context) error {
		current, err := u.trips.GetForUpdate(ctx, tripID)
		if err != nil {
			return err
		}
		if err := current.CanAddPosition(); err != nil {
			return err
		}

		position, err := u.positions.AddPosition(ctx, tripID, input)
		if err != nil {
			return err
		}
		if err := u.trips.UpdateLastPositionAt(ctx, tripID, input.RecordedAt, now); err != nil {
			return err
		}

		saved = position
		return nil
	})
	if err != nil {
		return trip.Position{}, fmt.Errorf("add position: %w", err)
	}

	return saved, nil
}

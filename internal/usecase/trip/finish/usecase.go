package finish

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/model"
)

type Usecase struct {
	tx      txManager
	trips   tripRepository
	history historyRepository
	now     func() time.Time
}

func New(tx txManager, trips tripRepository, history historyRepository, now func() time.Time) *Usecase {
	return &Usecase{tx: tx, trips: trips, history: history, now: now}
}

func (u *Usecase) Finish(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	var finished trip.Trip
	err := u.tx.Do(ctx, func(ctx context.Context) error {
		current, err := u.trips.GetForUpdate(ctx, id)
		if err != nil {
			return err
		}

		from := current.Status
		now := u.now().UTC().Truncate(time.Microsecond)
		if err := current.Finish(now); err != nil {
			return err
		}
		if err := u.trips.Update(ctx, current); err != nil {
			return err
		}
		if err := u.history.Add(ctx, model.StatusChange{
			TripID: current.ID, From: &from, To: current.Status, Reason: "trip finished", ChangedAt: now,
		}); err != nil {
			return err
		}
		finished = current
		return nil
	})
	if err != nil {
		return trip.Trip{}, fmt.Errorf("finish trip: %w", err)
	}
	return finished, nil
}

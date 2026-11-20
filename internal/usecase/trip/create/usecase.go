package create

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

func (u *Usecase) Create(ctx context.Context, input trip.Data) (trip.Trip, error) {
	// Match postgresql timestamp precision in the first response and in get
	now := u.now().UTC().Truncate(time.Microsecond)
	created, err := trip.New(uuid.New(), input, now)
	if err != nil {
		return trip.Trip{}, fmt.Errorf("validate trip: %w", err)
	}

	err = u.tx.Do(ctx, func(ctx context.Context) error {
		if err := u.trips.Create(ctx, created); err != nil {
			return err
		}
		return u.history.Add(ctx, model.StatusChange{
			TripID: created.ID, To: created.Status, Reason: "trip created", ChangedAt: now,
		})
	})
	if err != nil {
		return trip.Trip{}, fmt.Errorf("create trip: %w", err)
	}
	return created, nil
}

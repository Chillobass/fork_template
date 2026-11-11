package get

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type Usecase struct {
	trips tripRepository
}

func New(trips tripRepository) *Usecase {
	return &Usecase{trips: trips}
}

func (u *Usecase) Get(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	found, err := u.trips.GetByID(ctx, id)
	if err != nil {
		return trip.Trip{}, fmt.Errorf("get trip: %w", err)
	}
	return found, nil
}

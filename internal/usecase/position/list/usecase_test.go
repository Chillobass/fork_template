package list_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/usecase/position/list"
)

func TestList(t *testing.T) {
	tripID := uuid.New()
	sample := []trip.Position{
		{ID: 1, TripID: tripID, PositionData: trip.PositionData{Coordinates: trip.Point{Latitude: 59.9, Longitude: 30.3}, RecordedAt: time.Now().UTC()}},
	}

	for _, tc := range []struct {
		name      string
		tripErr   error
		positions []trip.Position
		listErr   error
		wantErr   error
	}{
		{name: "found", positions: sample},
		{name: "empty", positions: []trip.Position{}},
		{name: "trip not found", tripErr: trip.ErrTripNotFound, wantErr: trip.ErrTripNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			trips := NewMocktripRepository(ctrl)
			positions := NewMockpositionRepository(ctrl)
			ctx := t.Context()

			trips.EXPECT().GetByID(ctx, tripID).Return(trip.Trip{ID: tripID}, tc.tripErr)
			if tc.tripErr == nil {
				positions.EXPECT().ListPositions(ctx, tripID).Return(tc.positions, tc.listErr)
			}

			got, err := list.New(trips, positions).List(ctx, tripID)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && len(got) != len(tc.positions) {
				t.Fatalf("positions len = %d, want %d", len(got), len(tc.positions))
			}
		})
	}
}

package add_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/usecase/position/add"
)

func TestAdd(t *testing.T) {
	dbErr := errors.New("db error")
	now := time.Date(2026, 9, 1, 12, 0, 0, 123456789, time.FixedZone("UTC+3", 10800))
	at := now.UTC().Truncate(time.Microsecond)
	recorded := now.Add(-5 * time.Second)
	recordedAt := recorded.UTC().Truncate(time.Microsecond)

	for _, tc := range []struct {
		name      string
		status    trip.Status
		fetchErr  error
		insertErr error
		updateErr error
		wantErr   error
	}{
		{name: "success", status: trip.StatusActive},
		{name: "trip not found", fetchErr: trip.ErrTripNotFound, wantErr: trip.ErrTripNotFound},
		{name: "trip completed", status: trip.StatusCompleted, wantErr: trip.ErrTripCompleted},
		{name: "insert error", status: trip.StatusActive, insertErr: dbErr, wantErr: dbErr},
		{name: "update error", status: trip.StatusActive, updateErr: dbErr, wantErr: dbErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMocktxManager(ctrl)
			trips := NewMocktripRepository(ctrl)
			positions := NewMockpositionRepository(ctrl)

			tripID := uuid.New()
			input := trip.PositionData{
				Coordinates: trip.Point{Latitude: 59.935, Longitude: 30.335},
				RecordedAt:  recorded,
			}
			normalized := trip.PositionData{
				Coordinates: input.Coordinates,
				RecordedAt:  recordedAt,
			}
			wantPos := trip.Position{ID: 42, TripID: tripID, PositionData: normalized}

			ctx := t.Context()
			type transactionKey struct{}
			txCtx := context.WithValue(ctx, transactionKey{}, true)
			tx.EXPECT().Do(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				return fn(txCtx)
			})
			fetch := trips.EXPECT().GetForUpdate(txCtx, tripID).Return(trip.Trip{ID: tripID, Status: tc.status}, tc.fetchErr)

			if tc.fetchErr == nil && tc.status == trip.StatusActive {
				insert := positions.EXPECT().AddPosition(txCtx, tripID, normalized).Return(wantPos, tc.insertErr)
				gomock.InOrder(fetch, insert)
				if tc.insertErr == nil {
					update := trips.EXPECT().UpdateLastPositionAt(txCtx, tripID, recordedAt, at).Return(tc.updateErr)
					gomock.InOrder(insert, update)
				}
			}

			got, err := add.New(tx, trips, positions, func() time.Time { return now }).Add(ctx, tripID, input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && got != wantPos {
				t.Fatalf("position = %+v, want %+v", got, wantPos)
			}
		})
	}
}

func TestAddRejectsInvalidBeforeTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	uc := add.New(NewMocktxManager(ctrl), NewMocktripRepository(ctrl), NewMockpositionRepository(ctrl), func() time.Time { return now })

	var validation *trip.ValidationError
	_, err := uc.Add(t.Context(), uuid.New(), trip.PositionData{
		Coordinates: trip.Point{Latitude: 59.935, Longitude: 30.335},
		RecordedAt:  now.Add(2 * time.Minute),
	})
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

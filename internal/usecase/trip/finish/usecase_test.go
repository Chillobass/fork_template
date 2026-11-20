package finish_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/model"
	"github.com/Chillobass/fork_template/internal/usecase/trip/finish"
)

func TestFinish(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   trip.Status
		fetchErr error
		wantErr  error
	}{
		{name: "active", status: trip.StatusActive},
		{name: "completed", status: trip.StatusCompleted, wantErr: trip.ErrTripCompleted},
		{name: "missing", fetchErr: trip.ErrTripNotFound, wantErr: trip.ErrTripNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMocktxManager(ctrl)
			repo := NewMocktripRepository(ctrl)
			history := NewMockhistoryRepository(ctrl)
			now := time.Date(2026, 10, 3, 8, 0, 0, 123456789, time.UTC)
			at := now.Truncate(time.Microsecond)
			original := trip.Trip{ID: uuid.New(), Status: tc.status, StartedAt: at.Add(-time.Hour)}
			if original.Status == trip.StatusCompleted {
				original.FinishedAt = new(at.Add(-time.Minute))
			}

			ctx := t.Context()
			type transactionKey struct{}
			txCtx := context.WithValue(ctx, transactionKey{}, true)
			tx.EXPECT().Do(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				return fn(txCtx)
			})
			fetch := repo.EXPECT().GetForUpdate(txCtx, original.ID).Return(original, tc.fetchErr)

			var want trip.Trip
			if tc.wantErr == nil {
				want = original
				want.Status = trip.StatusCompleted
				want.FinishedAt = &at
				gomock.InOrder(
					fetch,
					repo.EXPECT().Update(txCtx, want),
					history.EXPECT().Add(txCtx, model.StatusChange{
						TripID: original.ID, From: new(trip.StatusActive), To: trip.StatusCompleted,
						ChangedAt: at, Reason: "trip finished",
					}),
				)
			}

			got, err := finish.New(tx, repo, history, func() time.Time { return now }).Finish(ctx, original.ID)
			if !errors.Is(err, tc.wantErr) || got.ID != want.ID || got.Status != want.Status || got.StartedAt != want.StartedAt {
				t.Fatalf("finish = (%+v, %v), want (%+v, %v)", got, err, want, tc.wantErr)
			}
			if want.FinishedAt != nil && (got.FinishedAt == nil || !got.FinishedAt.Equal(*want.FinishedAt)) {
				t.Fatalf("finished_at = %v, want %v", got.FinishedAt, want.FinishedAt)
			}
		})
	}
}

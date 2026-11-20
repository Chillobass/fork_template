package create_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/model"
	"github.com/Chillobass/fork_template/internal/usecase/trip/create"
)

func TestCreate(t *testing.T) {
	dbFailure := errors.New("database unavailable")
	for _, tc := range []struct {
		name       string
		insertErr  error
		historyErr error
	}{
		{name: "success"},
		{name: "busy driver", insertErr: trip.ErrDriverBusy},
		{name: "history error", historyErr: dbFailure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMocktxManager(ctrl)
			repo := NewMocktripRepository(ctrl)
			history := NewMockhistoryRepository(ctrl)
			now := time.Date(2026, 10, 2, 8, 0, 0, 123456789, time.FixedZone("UTC+3", 10800))
			at := now.UTC().Truncate(time.Microsecond)
			input := trip.Data{UserID: uuid.New(), DriverID: uuid.New(), Price: 10}
			var inserted trip.Trip

			ctx := t.Context()
			type transactionKey struct{}
			txCtx := context.WithValue(ctx, transactionKey{}, true)
			tx.EXPECT().Do(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				return fn(txCtx)
			})
			repo.EXPECT().Create(txCtx, gomock.Any()).DoAndReturn(func(_ context.Context, value trip.Trip) error {
				if value.ID == uuid.Nil || value.Data != input || value.Status != trip.StatusActive ||
					value.StartedAt != at || value.FinishedAt != nil {
					t.Fatalf("inserted trip = %+v", value)
				}
				inserted = value
				return tc.insertErr
			})
			if tc.insertErr == nil {
				history.EXPECT().Add(txCtx, gomock.Cond(func(change model.StatusChange) bool {
					return change.TripID == inserted.ID && change.From == nil && change.To == trip.StatusActive &&
						change.ChangedAt.Equal(at) && change.Reason == "trip created"
				})).Return(tc.historyErr)
			}

			uc := create.New(tx, repo, history, func() time.Time { return now })
			got, err := uc.Create(ctx, input)
			wantErr := tc.insertErr
			if wantErr == nil {
				wantErr = tc.historyErr
			}
			if wantErr != nil {
				if !errors.Is(err, wantErr) || got != (trip.Trip{}) {
					t.Fatalf("create = (%+v, %v), want zero trip and %v", got, err, wantErr)
				}
				return
			}
			if err != nil || got != inserted {
				t.Fatalf("create = (%+v, %v), want %+v", got, err, inserted)
			}
		})
	}
}

func TestCreateRejectsInvalidDataBeforeTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	uc := create.New(NewMocktxManager(ctrl), NewMocktripRepository(ctrl), NewMockhistoryRepository(ctrl), time.Now)
	var validation *trip.ValidationError
	if _, err := uc.Create(t.Context(), trip.Data{}); !errors.As(err, &validation) {
		t.Fatalf("create invalid trip: %v", err)
	}
}

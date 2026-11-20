package get_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/usecase/trip/get"
)

func TestGet(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "found"},
		{name: "missing", err: trip.ErrTripNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMocktripRepository(gomock.NewController(t))
			id := uuid.New()
			var want trip.Trip
			if tc.err == nil {
				want = trip.Trip{ID: id, Status: trip.StatusActive}
			}
			ctx := t.Context()
			repo.EXPECT().GetByID(ctx, id).Return(want, tc.err)

			got, err := get.New(repo).Get(ctx, id)
			if !errors.Is(err, tc.err) || got != want {
				t.Fatalf("get = (%+v, %v), want (%+v, %v)", got, err, want, tc.err)
			}
		})
	}
}

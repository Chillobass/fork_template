package finish_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/trip/finish"
)

func TestFinishTrip(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "completed", status: http.StatusOK},
		{name: "already completed", err: trip.ErrTripCompleted, status: http.StatusConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewMockusecase(gomock.NewController(t))
			id := uuid.New()
			at := time.Now().UTC()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+id.String()+"/finish", nil)
			uc.EXPECT().Finish(req.Context(), id).Return(trip.Trip{ID: id, Status: trip.StatusCompleted, FinishedAt: &at}, tc.err)
			rec := httptest.NewRecorder()
			count := 0
			finish.New(uc, slog.New(slog.DiscardHandler), func(_ context.Context) { count++ }).FinishTrip(rec, req, id)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.status, rec.Body)
			}
			if tc.err == nil {
				var got api.Trip
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				finished, err := got.FinishedAt.Get()
				if err != nil || got.Id != id || got.Status != api.Completed || !finished.Equal(at) || count != 1 {
					t.Fatalf("unexpected trip: %+v, count %d", got, count)
				}
				return
			}
			var problem api.Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
				t.Fatal(err)
			}
			if problem.Code != "trip_completed" || count != 0 {
				t.Fatalf("unexpected problem: %s, count %d", rec.Body, count)
			}
		})
	}
}

package get_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/trip/get"
)

func TestGetTrip(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "found", status: http.StatusOK},
		{name: "missing", err: trip.ErrTripNotFound, status: http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewMockusecase(gomock.NewController(t))
			id := uuid.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+id.String(), nil)
			uc.EXPECT().Get(req.Context(), id).Return(trip.Trip{ID: id, Status: trip.StatusActive}, tc.err)
			rec := httptest.NewRecorder()
			get.New(uc).GetTrip(rec, req, id)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.status, rec.Body)
			}
			if tc.err == nil {
				var got api.Trip
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.Id != id || got.Status != api.Active || !got.LastPositionAt.IsNull() {
					t.Fatalf("unexpected trip: %+v", got)
				}
				return
			}
			var problem api.Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
				t.Fatal(err)
			}
			if problem.Code != "trip_not_found" {
				t.Fatalf("unexpected problem: %s", rec.Body)
			}
		})
	}
}

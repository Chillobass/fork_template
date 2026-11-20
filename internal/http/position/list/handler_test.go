package list_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/position/list"
)

func TestListTripPositions(t *testing.T) {
	tripID := uuid.New()
	at := time.Date(2026, 9, 1, 10, 5, 0, 0, time.UTC)

	for _, tc := range []struct {
		name   string
		items  []trip.Position
		err    error
		status int
	}{
		{
			name: "non-empty",
			items: []trip.Position{
				{ID: 1, TripID: tripID, PositionData: trip.PositionData{Coordinates: trip.Point{Latitude: 59.9, Longitude: 30.3}, RecordedAt: at}},
			},
			status: http.StatusOK,
		},
		{name: "empty array not null", items: nil, status: http.StatusOK},
		{name: "missing trip", err: trip.ErrTripNotFound, status: http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewMockusecase(gomock.NewController(t))
			req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+tripID.String()+"/positions", nil)
			uc.EXPECT().List(req.Context(), tripID).Return(tc.items, tc.err)

			rec := httptest.NewRecorder()
			list.New(uc).ListTripPositions(rec, req, tripID)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.status, rec.Body)
			}
			if tc.err != nil {
				var problem api.Problem
				if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
					t.Fatal(err)
				}
				if problem.Code != "trip_not_found" {
					t.Fatalf("unexpected problem: %s", rec.Body)
				}
				return
			}
			if tc.items == nil && !strings.Contains(rec.Body.String(), `"positions":[]`) {
				t.Fatalf("expected empty JSON array, got %s", rec.Body)
			}
			var got api.TripPositionsResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if len(got.Positions) != len(tc.items) {
				t.Fatalf("positions len = %d, want %d", len(got.Positions), len(tc.items))
			}
		})
	}
}

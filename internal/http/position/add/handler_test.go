package add_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/position/add"
)

const validPositionRequest = `{
	"coordinates": {"latitude": 0, "longitude": 30.335},
	"recorded_at": "2026-09-01T10:05:00Z"
}`

func TestCreateTripPosition(t *testing.T) {
	const (
		private        = "SQL failure"
		invalidRequest = "invalid_request"
	)

	for _, tc := range []struct {
		name       string
		body       string
		expectCall bool
		err        error
		status     int
		code       string
	}{
		{name: "created", body: validPositionRequest, expectCall: true, status: http.StatusCreated},
		{name: "not found", body: validPositionRequest, expectCall: true, err: trip.ErrTripNotFound, status: http.StatusNotFound, code: "trip_not_found"},
		{name: "completed", body: validPositionRequest, expectCall: true, err: trip.ErrTripCompleted, status: http.StatusConflict, code: "trip_completed"},
		{name: "internal error", body: validPositionRequest, expectCall: true, err: errors.New(private), status: http.StatusInternalServerError, code: "internal_error"},
		{name: "invalid JSON", body: "{", status: http.StatusBadRequest, code: invalidRequest},
		{name: "missing latitude", body: `{"coordinates":{"longitude":30},"recorded_at":"2026-09-01T10:05:00Z"}`, status: http.StatusBadRequest, code: invalidRequest},
		{name: "missing recorded_at", body: `{"coordinates":{"latitude":59,"longitude":30}}`, status: http.StatusBadRequest, code: invalidRequest},
		{name: "bad timestamp", body: `{"coordinates":{"latitude":59,"longitude":30},"recorded_at":"not-a-date"}`, status: http.StatusBadRequest, code: invalidRequest},
		{name: "unknown field", body: `{"coordinates":{"latitude":59,"longitude":30},"recorded_at":"2026-09-01T10:05:00Z","extra":1}`, status: http.StatusBadRequest, code: invalidRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := NewMockusecase(ctrl)
			canceller := NewMockrequestCanceller(ctrl)
			tripID := uuid.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+tripID.String()+"/positions", strings.NewReader(tc.body))

			recordedAt := time.Date(2026, 9, 1, 10, 5, 0, 0, time.UTC)
			data := trip.PositionData{
				Coordinates: trip.Point{Latitude: 0, Longitude: 30.335},
				RecordedAt:  recordedAt,
			}
			saved := trip.Position{ID: 42, TripID: tripID, PositionData: data}

			if tc.expectCall {
				uc.EXPECT().Add(req.Context(), tripID, data).Return(saved, tc.err)
				if tc.err == nil {
					canceller.EXPECT().Cancel(req.Context(), tripID).Return(true)
				}
			}

			rec := httptest.NewRecorder()
			count := 0
			add.New(uc, canceller, func(_ context.Context) { count++ }).CreateTripPosition(rec, req, tripID)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.status, rec.Body)
			}
			if tc.status == http.StatusCreated {
				var got api.TripPosition
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.Id != 42 || got.TripId != tripID || !got.RecordedAt.Equal(recordedAt) || count != 1 {
					t.Fatalf("unexpected response: %+v, count %d", got, count)
				}
				return
			}

			var problem api.Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
				t.Fatal(err)
			}
			if problem.Code != tc.code || strings.Contains(rec.Body.String(), private) || count != 0 {
				t.Fatalf("unexpected error response: %s, count %d", rec.Body, count)
			}
		})
	}
}

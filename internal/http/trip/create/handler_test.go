package create_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/trip/create"
)

const validRequest = `{
	"user_id": "589b9e2a-52f2-45d0-a04c-f2823c48b9a7",
	"driver_id": "c0d1fc78-11b7-4cf8-a43e-f079eec2c5a5",
	"start_point": {"latitude": 0, "longitude": 3},
	"end_point": {"latitude": 4, "longitude": 5},
	"price": 0
}`

func TestCreateTrip(t *testing.T) {
	const private = "SQL: could not connect"
	const invalidRequest = "invalid_request"
	for _, tc := range []struct {
		name       string
		body       string
		expectCall bool
		err        error
		status     int
		code       string
	}{
		{name: "created", body: validRequest, expectCall: true, status: http.StatusCreated},
		{name: "busy driver", body: validRequest, expectCall: true, err: trip.ErrDriverBusy, status: http.StatusConflict, code: "driver_busy"},
		{name: "internal error", body: validRequest, expectCall: true, err: errors.New(private), status: http.StatusInternalServerError, code: "internal_error"},
		{name: "invalid JSON", body: "{", status: http.StatusBadRequest, code: invalidRequest},
		{name: "missing price", body: strings.Replace(validRequest, ",\n\t\"price\": 0", "", 1), status: http.StatusBadRequest, code: invalidRequest},
		{name: "null latitude", body: strings.Replace(validRequest, "\"latitude\": 0", "\"latitude\": null", 1), status: http.StatusBadRequest, code: invalidRequest},
		{name: "unknown field", body: strings.Replace(validRequest, "\"price\": 0", "\"price\": 0, \"other\": 1", 1), status: http.StatusBadRequest, code: invalidRequest},
		{name: "multiple objects", body: validRequest + "{}", status: http.StatusBadRequest, code: invalidRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewMockusecase(gomock.NewController(t))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(tc.body))
			data := trip.Data{
				UserID:     uuid.MustParse("589b9e2a-52f2-45d0-a04c-f2823c48b9a7"),
				DriverID:   uuid.MustParse("c0d1fc78-11b7-4cf8-a43e-f079eec2c5a5"),
				StartPoint: trip.Point{Longitude: 3},
				EndPoint:   trip.Point{Latitude: 4, Longitude: 5},
			}
			created := trip.Trip{ID: uuid.New(), Data: data, Status: trip.StatusActive, StartedAt: time.Now().UTC()}
			if tc.expectCall {
				uc.EXPECT().Create(req.Context(), data).Return(created, tc.err)
			}
			rec := httptest.NewRecorder()
			count := 0
			create.New(uc, slog.New(slog.DiscardHandler), func(_ context.Context) { count++ }).CreateTrip(rec, req, api.CreateTripParams{})

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.status, rec.Body)
			}
			if tc.status == http.StatusCreated {
				var got api.Trip
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.Id != created.ID || got.Status != api.Active || !got.FinishedAt.IsNull() ||
					count != 1 || rec.Header().Get("Location") != "/api/v1/trips/"+created.ID.String() {
					t.Fatalf("unexpected create response: %+v, count %d, headers %v", got, count, rec.Header())
				}
				return
			}
			var got api.Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Code != tc.code || strings.Contains(rec.Body.String(), private) || count != 0 ||
				rec.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("unexpected error response: %s, count %d, headers %v", rec.Body, count, rec.Header())
			}
		})
	}
}

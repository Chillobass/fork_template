package trip_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

func TestValidate(t *testing.T) {
	valid := trip.Data{
		UserID: uuid.New(), DriverID: uuid.New(),
		StartPoint: trip.Point{Latitude: -90, Longitude: 180},
		EndPoint:   trip.Point{Latitude: 90, Longitude: -180},
	}
	cases := []struct {
		name   string
		change func(*trip.Data)
		field  string
	}{
		{name: "boundary values and zero price"},
		{name: "missing user", change: func(d *trip.Data) { d.UserID = uuid.Nil }, field: "user_id"},
		{name: "missing driver", change: func(d *trip.Data) { d.DriverID = uuid.Nil }, field: "driver_id"},
		{name: "start latitude", change: func(d *trip.Data) { d.StartPoint.Latitude = -91 }, field: "start_point.latitude"},
		{name: "end longitude", change: func(d *trip.Data) { d.EndPoint.Longitude = 181 }, field: "end_point.longitude"},
		{name: "NaN longitude", change: func(d *trip.Data) { d.StartPoint.Longitude = math.NaN() }, field: "start_point.longitude"},
		{name: "negative price", change: func(d *trip.Data) { d.Price = -1 }, field: "price"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := valid
			if tc.change != nil {
				tc.change(&data)
			}
			err := data.Validate()
			if tc.field == "" {
				if err != nil {
					t.Fatalf("validate: %v", err)
				}
				return
			}
			var validation *trip.ValidationError
			if !errors.As(err, &validation) || validation.Field != tc.field || validation.Detail == "" {
				t.Fatalf("expected validation error on %s, got %v", tc.field, err)
			}
		})
	}
}

func TestFinish(t *testing.T) {
	at := time.Date(2026, 9, 1, 11, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name       string
		status     trip.Status
		finishedAt *time.Time
		wantErr    error
	}{
		{name: "active", status: trip.StatusActive},
		{name: "completed", status: trip.StatusCompleted, finishedAt: new(at.Add(-time.Hour)), wantErr: trip.ErrTripCompleted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value := trip.Trip{Status: tc.status, FinishedAt: tc.finishedAt}
			if err := value.Finish(at); !errors.Is(err, tc.wantErr) {
				t.Fatalf("finish error = %v, want %v", err, tc.wantErr)
			}
			want := at
			if tc.finishedAt != nil {
				want = *tc.finishedAt
			}
			if value.Status != trip.StatusCompleted || value.FinishedAt == nil || !value.FinishedAt.Equal(want) {
				t.Fatalf("unexpected trip after finish: %+v", value)
			}
		})
	}
}

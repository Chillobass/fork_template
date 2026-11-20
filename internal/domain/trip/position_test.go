package trip_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

func TestPositionValidate(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	valid := trip.PositionData{
		Coordinates: trip.Point{Latitude: 59.935, Longitude: 30.335},
		RecordedAt:  now,
	}

	cases := []struct {
		name   string
		change func(*trip.PositionData)
		field  string
	}{
		{name: "valid current time"},
		{name: "valid boundary 1m ahead", change: func(d *trip.PositionData) { d.RecordedAt = now.Add(time.Minute) }},
		{name: "valid past time", change: func(d *trip.PositionData) { d.RecordedAt = now.Add(-time.Hour) }},
		{name: "latitude out of range", change: func(d *trip.PositionData) { d.Coordinates.Latitude = 91 }, field: "coordinates.latitude"},
		{name: "longitude NaN", change: func(d *trip.PositionData) { d.Coordinates.Longitude = math.NaN() }, field: "coordinates.longitude"},
		{name: "zero recorded_at", change: func(d *trip.PositionData) { d.RecordedAt = time.Time{} }, field: "recorded_at"},
		{name: "too far in future", change: func(d *trip.PositionData) { d.RecordedAt = now.Add(time.Minute + time.Second) }, field: "recorded_at"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := valid
			if tc.change != nil {
				tc.change(&data)
			}
			err := data.Validate(now)
			if tc.field == "" {
				if err != nil {
					t.Fatalf("validate: %v", err)
				}
				return
			}
			var validation *trip.ValidationError
			if !errors.As(err, &validation) || validation.Field != tc.field {
				t.Fatalf("expected validation error on %s, got %v", tc.field, err)
			}
		})
	}
}

func TestTripCanAddPosition(t *testing.T) {
	if err := (trip.Trip{Status: trip.StatusActive}).CanAddPosition(); err != nil {
		t.Fatalf("active trip: %v", err)
	}
	if err := (trip.Trip{Status: trip.StatusCompleted}).CanAddPosition(); !errors.Is(err, trip.ErrTripCompleted) {
		t.Fatalf("completed trip error = %v, want %v", err, trip.ErrTripCompleted)
	}
}

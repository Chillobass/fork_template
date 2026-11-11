package trip

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
)

type Point struct {
	Latitude  float64
	Longitude float64
}

type Data struct {
	UserID     uuid.UUID
	DriverID   uuid.UUID
	StartPoint Point
	EndPoint   Point
	Price      int64
}

type Trip struct {
	Data
	ID         uuid.UUID
	Status     Status
	StartedAt  time.Time
	FinishedAt *time.Time
}

func New(id uuid.UUID, data Data, at time.Time) (Trip, error) {
	if err := data.Validate(); err != nil {
		return Trip{}, err
	}
	return Trip{Data: data, ID: id, Status: StatusActive, StartedAt: at}, nil
}

func (t *Trip) Finish(at time.Time) error {
	if t.Status != StatusActive {
		return ErrTripCompleted
	}
	t.Status = StatusCompleted
	t.FinishedAt = &at
	return nil
}

func (d Data) Validate() error {
	switch {
	case d.UserID == uuid.Nil:
		return &ValidationError{Field: "user_id", Detail: "must be a non-empty UUID"}
	case d.DriverID == uuid.Nil:
		return &ValidationError{Field: "driver_id", Detail: "must be a non-empty UUID"}
	}
	if err := d.StartPoint.validate("start_point"); err != nil {
		return err
	}
	if err := d.EndPoint.validate("end_point"); err != nil {
		return err
	}
	if d.Price < 0 {
		return &ValidationError{Field: "price", Detail: "must be greater than or equal to 0"}
	}
	return nil
}

func (p Point) validate(field string) error {
	switch {
	case math.IsNaN(p.Latitude) || math.Abs(p.Latitude) > 90:
		return &ValidationError{Field: field + ".latitude", Detail: "must be between -90 and 90"}
	case math.IsNaN(p.Longitude) || math.Abs(p.Longitude) > 180:
		return &ValidationError{Field: field + ".longitude", Detail: "must be between -180 and 180"}
	default:
		return nil
	}
}

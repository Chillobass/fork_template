package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TripStatus string

const (
	TripStatusActive    TripStatus = "active"
	TripStatusCompleted TripStatus = "completed"
)

type Point struct {
	Latitude  float64
	Longitude float64
}

type Trip struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	DriverID   uuid.UUID
	StartPoint Point
	EndPoint   Point
	Price      int64
	Status     TripStatus
	StartedAt  time.Time
	FinishedAt *time.Time
}

type StatusChange struct {
	TripID    uuid.UUID
	From      *TripStatus
	To        TripStatus
	Reason    string
	ChangedAt time.Time
}

var (
	ErrTripNotFound  = errors.New("trip not found")
	ErrTripCompleted = errors.New("trip is already completed")
	ErrDriverBusy    = errors.New("driver already has an active trip")
)

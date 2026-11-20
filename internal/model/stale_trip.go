package model

import (
	"time"

	"github.com/google/uuid"
)

type StaleTrip struct {
	TripID         uuid.UUID
	DriverID       uuid.UUID
	LastActivityAt time.Time
}

type StaleCursor struct {
	LastActivityAt time.Time
	TripID         uuid.UUID
}

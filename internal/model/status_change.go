package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type StatusChange struct {
	TripID    uuid.UUID
	From      *trip.Status
	To        trip.Status
	Reason    string
	ChangedAt time.Time
}

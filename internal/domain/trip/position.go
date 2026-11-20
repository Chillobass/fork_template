package trip

import (
	"time"

	"github.com/google/uuid"
)

const MaxRecordedAtFutureSkew = time.Minute

type PositionData struct {
	Coordinates Point
	RecordedAt  time.Time
}

type Position struct {
	ID     int64
	TripID uuid.UUID
	PositionData
}

func (d PositionData) Validate(now time.Time) error {
	if err := d.Coordinates.validate("coordinates"); err != nil {
		return err
	}
	if d.RecordedAt.IsZero() {
		return &ValidationError{Field: "recorded_at", Detail: "must be a valid RFC 3339 timestamp"}
	}
	if d.RecordedAt.After(now.Add(MaxRecordedAtFutureSkew)) {
		return &ValidationError{Field: "recorded_at", Detail: "must not be more than 1 minute in the future"}
	}
	return nil
}

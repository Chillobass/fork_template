package trip

import (
	"time"

	sq "github.com/Masterminds/squirrel"
)

const (
	tripsTable                 = "trips"
	uniqueViolationCode        = "23505"
	driverActiveTripConstraint = "trips_driver_active_uniq"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var tripColumns = []string{
	"id",
	"user_id",
	"driver_id",
	"start_latitude",
	"start_longitude",
	"end_latitude",
	"end_longitude",
	"price",
	"status",
	"started_at",
	"finished_at",
}

type TripRepository struct {
	db           executor
	queryTimeout time.Duration
}

func New(db executor, queryTimeout time.Duration) *TripRepository {
	return &TripRepository{db: db, queryTimeout: queryTimeout}
}

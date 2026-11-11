package trip

import (
	"github.com/jackc/pgx/v5"

	domain "github.com/Chillobass/fork_template/internal/domain/trip"
)

func scanTrip(row pgx.Row) (domain.Trip, error) {
	var (
		trip   domain.Trip
		status string
	)

	err := row.Scan(
		&trip.ID,
		&trip.UserID,
		&trip.DriverID,
		&trip.StartPoint.Latitude,
		&trip.StartPoint.Longitude,
		&trip.EndPoint.Latitude,
		&trip.EndPoint.Longitude,
		&trip.Price,
		&status,
		&trip.StartedAt,
		&trip.FinishedAt,
	)
	if err != nil {
		return domain.Trip{}, err
	}

	trip.Status = domain.Status(status)

	return trip, nil
}

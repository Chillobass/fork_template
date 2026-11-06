package trip

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"

	domain "github.com/Chillobass/fork_template/internal/domain/trip"
)

func (r *TripRepository) Create(ctx context.Context, trip domain.Trip) error {
	query, args, err := psql.
		Insert(tripsTable).
		Columns(tripColumns...).
		Values(
			trip.ID,
			trip.UserID,
			trip.DriverID,
			trip.StartPoint.Latitude,
			trip.StartPoint.Longitude,
			trip.EndPoint.Latitude,
			trip.EndPoint.Longitude,
			trip.Price,
			string(trip.Status),
			trip.StartedAt,
			trip.FinishedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert trip query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	_, err = r.db.Exec(ctx, query, args...)
	if isUniqueViolation(err, driverActiveTripConstraint) {
		return fmt.Errorf("insert active trip: %w: %w", domain.ErrDriverBusy, err)
	}
	if err != nil {
		return fmt.Errorf("insert trip %s: %w", trip.ID, err)
	}

	return nil
}

func (r *TripRepository) Update(ctx context.Context, trip domain.Trip) error {
	query, args, err := psql.
		Update(tripsTable).
		Set("status", string(trip.Status)).
		Set("finished_at", trip.FinishedAt).
		Set("updated_at", trip.FinishedAt).
		Where(sq.Eq{"id": trip.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update trip query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("update trip %s: %w", trip.ID, err)
	}
	return nil
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == uniqueViolationCode &&
		pgErr.ConstraintName == constraint
}

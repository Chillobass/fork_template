package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Chillobass/fork_template/internal/domain"
	"github.com/Chillobass/fork_template/internal/postgres"
)

const (
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
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewTripRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *TripRepository {
	return &TripRepository{pool: pool, queryTimeout: queryTimeout}
}

func (r *TripRepository) Create(ctx context.Context, trip domain.Trip) error {
	query, args, err := psql.
		Insert("trips").
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

	_, err = postgres.QuerierFromContext(ctx, r.pool).Exec(ctx, query, args...)
	if isUniqueViolation(err, driverActiveTripConstraint) {
		return domain.ErrDriverBusy
	}
	if err != nil {
		return fmt.Errorf("insert trip %s: %w", trip.ID, err)
	}

	return nil
}

func (r *TripRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	query, args, err := psql.
		Select(tripColumns...).
		From("trips").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return domain.Trip{}, fmt.Errorf("build select trip query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	trip, err := scanTrip(postgres.QuerierFromContext(ctx, r.pool).QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Trip{}, domain.ErrTripNotFound
	}
	if err != nil {
		return domain.Trip{}, fmt.Errorf("select trip %s: %w", id, err)
	}

	return trip, nil
}

func (r *TripRepository) Finish(ctx context.Context, id uuid.UUID, finishedAt time.Time) (domain.Trip, error) {
	query, args, err := psql.
		Update("trips").
		Set("status", string(domain.TripStatusCompleted)).
		Set("finished_at", finishedAt).
		Set("updated_at", finishedAt).
		Where(sq.Eq{"id": id, "status": string(domain.TripStatusActive)}).
		Suffix("RETURNING " + strings.Join(tripColumns, ", ")).
		ToSql()
	if err != nil {
		return domain.Trip{}, fmt.Errorf("build finish trip query: %w", err)
	}

	updateCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	trip, err := scanTrip(postgres.QuerierFromContext(updateCtx, r.pool).QueryRow(updateCtx, query, args...))
	if err == nil {
		return trip, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Trip{}, fmt.Errorf("finish trip %s: %w", id, err)
	}

	// UPDATE ничего не обновил: поездки либо нет (GetByID вернёт ErrTripNotFound), либо она уже завершена
	if _, err := r.GetByID(ctx, id); err != nil {
		return domain.Trip{}, err
	}

	return domain.Trip{}, domain.ErrTripCompleted
}

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

	trip.Status = domain.TripStatus(status)

	return trip, nil
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == uniqueViolationCode &&
		pgErr.ConstraintName == constraint
}

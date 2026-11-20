package trip

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domain "github.com/Chillobass/fork_template/internal/domain/trip"
)

func (r *TripRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	return r.fetch(ctx, id, false)
}

// GetForUpdate holds the row lock until the calling transaction ends.
func (r *TripRepository) GetForUpdate(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	return r.fetch(ctx, id, true)
}

func (r *TripRepository) fetch(ctx context.Context, id uuid.UUID, lock bool) (domain.Trip, error) {
	statement := psql.Select(tripColumns...).From(tripsTable).Where(sq.Eq{"id": id})
	if lock {
		statement = statement.Suffix("FOR UPDATE")
	}
	query, args, err := statement.ToSql()
	if err != nil {
		return domain.Trip{}, fmt.Errorf("build select trip query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	trip, err := scanTrip(r.db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Trip{}, domain.ErrTripNotFound
	}
	if err != nil {
		return domain.Trip{}, fmt.Errorf("select trip %s: %w", id, err)
	}

	return trip, nil
}

func (r *TripRepository) CountActive(ctx context.Context) (int64, error) {
	const query = "SELECT count(*) FROM trips WHERE status = $1"

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	var count int64
	if err := r.db.QueryRow(ctx, query, domain.StatusActive).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active trips: %w", err)
	}
	return count, nil
}

package trip

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domain "github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/model"
)

func (r *TripRepository) AddPosition(ctx context.Context, tripID uuid.UUID, data domain.PositionData) (domain.Position, error) {
	query, args, err := psql.
		Insert(tripPositionsTable).
		Columns("trip_id", "latitude", "longitude", "recorded_at").
		Values(tripID, data.Coordinates.Latitude, data.Coordinates.Longitude, data.RecordedAt).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return domain.Position{}, fmt.Errorf("build insert position query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var id int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return domain.Position{}, fmt.Errorf("insert position for trip %s: %w", tripID, err)
	}

	return domain.Position{
		ID:           id,
		TripID:       tripID,
		PositionData: data,
	}, nil
}

func (r *TripRepository) ListPositions(ctx context.Context, tripID uuid.UUID) ([]domain.Position, error) {
	query, args, err := psql.
		Select("id", "trip_id", "latitude", "longitude", "recorded_at").
		From(tripPositionsTable).
		Where(sq.Eq{"trip_id": tripID}).
		OrderBy("recorded_at ASC", "id ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list positions query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list positions for trip %s: %w", tripID, err)
	}
	defer rows.Close()

	positions := make([]domain.Position, 0)
	for rows.Next() {
		var p domain.Position
		if err := rows.Scan(
			&p.ID,
			&p.TripID,
			&p.Coordinates.Latitude,
			&p.Coordinates.Longitude,
			&p.RecordedAt,
		); err != nil {
			return nil, fmt.Errorf("scan position for trip %s: %w", tripID, err)
		}
		positions = append(positions, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate positions for trip %s: %w", tripID, err)
	}

	return positions, nil
}

const (
	listStaleFirstBatchQuery = `
		SELECT id, driver_id, COALESCE(last_position_at, started_at) AS last_activity_at
		FROM trips
		WHERE status = 'active'
		  AND COALESCE(last_position_at, started_at) < $1
		ORDER BY COALESCE(last_position_at, started_at) ASC, id ASC
		LIMIT $2`

	listStaleNextBatchQuery = `
		SELECT id, driver_id, COALESCE(last_position_at, started_at) AS last_activity_at
		FROM trips
		WHERE status = 'active'
		  AND COALESCE(last_position_at, started_at) < $1
		  AND (COALESCE(last_position_at, started_at), id) > ($2, $3)
		ORDER BY COALESCE(last_position_at, started_at) ASC, id ASC
		LIMIT $4`
)

func (r *TripRepository) ListStale(ctx context.Context, staleBefore time.Time, after *model.StaleCursor, limit int) ([]model.StaleTrip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var (
		rows pgx.Rows
		err  error
	)
	if after == nil {
		rows, err = r.db.Query(ctx, listStaleFirstBatchQuery, staleBefore, limit)
	} else {
		rows, err = r.db.Query(ctx, listStaleNextBatchQuery, staleBefore, after.LastActivityAt, after.TripID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("list stale trips: %w", err)
	}
	defer rows.Close()

	result := make([]model.StaleTrip, 0, limit)
	for rows.Next() {
		var item model.StaleTrip
		if err := rows.Scan(&item.TripID, &item.DriverID, &item.LastActivityAt); err != nil {
			return nil, fmt.Errorf("scan stale trip: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stale trips: %w", err)
	}

	return result, nil
}

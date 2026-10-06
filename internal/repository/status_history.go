package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Chillobass/fork_template/internal/domain"
	"github.com/Chillobass/fork_template/internal/postgres"
)

type StatusHistoryRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewStatusHistoryRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *StatusHistoryRepository {
	return &StatusHistoryRepository{pool: pool, queryTimeout: queryTimeout}
}

func (r *StatusHistoryRepository) Add(ctx context.Context, change domain.StatusChange) error {
	var from *string
	if change.From != nil {
		s := string(*change.From)
		from = &s
	}

	query, args, err := psql.
		Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(change.TripID, from, string(change.To), change.Reason, change.ChangedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert status history query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	if _, err := postgres.QuerierFromContext(ctx, r.pool).Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("insert status history for trip %s: %w", change.TripID, err)
	}

	return nil
}

//go:build integration

package trip_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/infra/postgres/transaction"
	"github.com/Chillobass/fork_template/internal/model"
	repository "github.com/Chillobass/fork_template/internal/repository/trip"
	"github.com/Chillobass/fork_template/internal/usecase/trip/create"
	"github.com/Chillobass/fork_template/internal/usecase/trip/finish"
)

const queryTimeout = 5 * time.Second

func setup(t *testing.T) (*pgxpool.Pool, *transaction.TxManager, *repository.TripRepository, trip.Data) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required. Run tripgoctl environment start and make migrate.")
	}

	ctx, cancel := context.WithTimeout(t.Context(), queryTimeout)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	input := trip.Data{UserID: uuid.New(), DriverID: uuid.New(), Price: 120}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), queryTimeout)
		defer cancel()
		// Deleting the trip removes its history through the migration's foreign key.
		if _, err := pool.Exec(ctx, "DELETE FROM trips WHERE driver_id = $1", input.DriverID); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	})

	tx := transaction.NewTxManager(pool, queryTimeout)
	return pool, tx, repository.New(tx, queryTimeout), input
}

func TestTripLifecycle(t *testing.T) {
	pool, tx, repo, input := setup(t)
	ctx, cancel := context.WithTimeout(t.Context(), 4*queryTimeout)
	defer cancel()
	creator := create.New(tx, repo, repo, time.Now)
	created, err := creator.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := creator.Create(ctx, input); !errors.Is(err, trip.ErrDriverBusy) {
		t.Fatalf("duplicate active trip: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil || got.Data != created.Data || got.Status != trip.StatusActive ||
		!got.StartedAt.Equal(created.StartedAt) || got.FinishedAt != nil {
		t.Fatalf("stored trip = (%+v, %v), created %+v", got, err, created)
	}

	finisher := finish.New(tx, repo, repo, time.Now)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	start := make(chan struct{})
	for range 2 {
		wg.Go(func() {
			<-start
			_, err := finisher.Finish(ctx, created.ID)
			results <- err
		})
	}
	close(start)
	wg.Wait()
	close(results)
	var completed, conflicts int
	for err := range results {
		switch {
		case err == nil:
			completed++
		case errors.Is(err, trip.ErrTripCompleted):
			conflicts++
		default:
			t.Fatalf("concurrent finish: %v", err)
		}
	}
	if completed != 1 || conflicts != 1 {
		t.Fatalf("finish outcomes: %d successes, %d conflicts", completed, conflicts)
	}

	got, err = repo.GetByID(ctx, created.ID)
	if err != nil || got.Status != trip.StatusCompleted || got.FinishedAt == nil {
		t.Fatalf("persisted completion = (%+v, %v)", got, err)
	}
	var transitions int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM trip_status_history WHERE trip_id = $1", created.ID).Scan(&transitions); err != nil {
		t.Fatal(err)
	}
	if transitions != 2 {
		t.Fatalf("expected creation and one completion, got %d transitions", transitions)
	}
	if _, err := creator.Create(ctx, input); err != nil {
		t.Fatalf("driver still busy after completion: %v", err)
	}
}

func TestRollbackInserts(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status trip.Status
	}{
		{name: "second insert fails", status: "invalid-status"},
		{name: "both inserts succeed before rollback", status: trip.StatusActive},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool, tx, repo, input := setup(t)
			ctx, cancel := context.WithTimeout(t.Context(), 4*queryTimeout)
			defer cancel()
			at := time.Now().UTC()
			value, err := trip.New(uuid.New(), input, at)
			if err != nil {
				t.Fatal(err)
			}
			forced := errors.New("force rollback after both inserts")
			err = tx.Do(ctx, func(ctx context.Context) error {
				if err := repo.Create(ctx, value); err != nil {
					return err
				}
				if err := repo.Add(ctx, model.StatusChange{TripID: value.ID, To: tc.status, ChangedAt: at}); err != nil {
					return err
				}
				return forced
			})
			if err == nil || errors.Is(err, forced) != (tc.status == trip.StatusActive) {
				t.Fatalf("unexpected transaction error: %v", err)
			}
			if _, err := repo.GetByID(ctx, value.ID); !errors.Is(err, trip.ErrTripNotFound) {
				t.Fatalf("trip insert was not rolled back: %v", err)
			}
			var history int
			if err := pool.QueryRow(ctx, "SELECT count(*) FROM trip_status_history WHERE trip_id = $1", value.ID).Scan(&history); err != nil {
				t.Fatal(err)
			}
			if history != 0 {
				t.Fatalf("history insert was not rolled back: %d rows", history)
			}
		})
	}
}

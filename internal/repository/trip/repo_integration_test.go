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
	posadd "github.com/Chillobass/fork_template/internal/usecase/position/add"
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

func TestPositionsAtomicUpdateAndOutOfOrder(t *testing.T) {
	pool, tx, repo, input := setup(t)
	ctx, cancel := context.WithTimeout(t.Context(), 4*queryTimeout)
	defer cancel()

	base := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	created, err := create.New(tx, repo, repo, func() time.Time { return base }).Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}

	adder := posadd.New(tx, repo, repo, func() time.Time { return base.Add(10 * time.Minute) })

	tNew := base.Add(5 * time.Minute)
	tOld := base.Add(2 * time.Minute)
	p1, err := adder.Add(ctx, created.ID, trip.PositionData{
		Coordinates: trip.Point{Latitude: 59.935, Longitude: 30.335},
		RecordedAt:  tNew,
	})
	if err != nil {
		t.Fatal(err)
	}
	p2, err := adder.Add(ctx, created.ID, trip.PositionData{
		Coordinates: trip.Point{Latitude: 59.931, Longitude: 30.331},
		RecordedAt:  tOld,
	})
	if err != nil {
		t.Fatal(err)
	}

	gotTrip, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotTrip.LastPositionAt == nil || !gotTrip.LastPositionAt.Equal(tNew) {
		t.Fatalf("last_position_at = %v, want %v (older position must not regress it)", gotTrip.LastPositionAt, tNew)
	}

	positions, err := repo.ListPositions(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(positions) != 2 || positions[0].ID != p2.ID || positions[1].ID != p1.ID {
		t.Fatalf("positions must be sorted by recorded_at ASC: %+v", positions)
	}

	tRolledBack := base.Add(8 * time.Minute)
	forced := errors.New("force rollback")
	err = tx.Do(ctx, func(ctx context.Context) error {
		if _, err := repo.AddPosition(ctx, created.ID, trip.PositionData{
			Coordinates: trip.Point{Latitude: 59.94, Longitude: 30.34},
			RecordedAt:  tRolledBack,
		}); err != nil {
			return err
		}
		if err := repo.UpdateLastPositionAt(ctx, created.ID, tRolledBack, base.Add(9*time.Minute)); err != nil {
			return err
		}
		return forced
	})
	if !errors.Is(err, forced) {
		t.Fatalf("expected forced error, got %v", err)
	}

	gotTrip, err = repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotTrip.LastPositionAt == nil || !gotTrip.LastPositionAt.Equal(tNew) {
		t.Fatalf("last_position_at changed after rollback: %v", gotTrip.LastPositionAt)
	}
	var posCount int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM trip_positions WHERE trip_id = $1", created.ID).Scan(&posCount); err != nil {
		t.Fatal(err)
	}
	if posCount != 2 {
		t.Fatalf("expected 2 positions after rollback, got %d", posCount)
	}
}

func TestListStaleTripsAndKeysetPagination(t *testing.T) {
	pool, _, repo, _ := setup(t)
	ctx, cancel := context.WithTimeout(t.Context(), 4*queryTimeout)
	defer cancel()

	base := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	cutoff := base.Add(-time.Minute)

	var createdIDs []uuid.UUID
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.WithoutCancel(t.Context()), queryTimeout)
		defer cleanupCancel()
		if _, err := pool.Exec(cleanupCtx, "DELETE FROM trips WHERE id = ANY($1)", createdIDs); err != nil {
			t.Errorf("cleanup stale test trips: %v", err)
		}
	})

	makeTrip := func(startedAt time.Time, status trip.Status, lastPos *time.Time) trip.Trip {
		t.Helper()
		val := trip.Trip{
			ID: uuid.New(),
			Data: trip.Data{
				UserID:   uuid.New(),
				DriverID: uuid.New(),
				Price:    100,
			},
			Status:         status,
			StartedAt:      startedAt,
			LastPositionAt: lastPos,
		}
		if status == trip.StatusCompleted {
			f := startedAt.Add(time.Minute)
			val.FinishedAt = &f
		}
		if err := repo.Create(ctx, val); err != nil {
			t.Fatal(err)
		}
		createdIDs = append(createdIDs, val.ID)
		return val
	}

	staleNoPos := makeTrip(base.Add(-10*time.Minute), trip.StatusActive, nil)
	sameTS := base.Add(-5 * time.Minute)
	staleTie1 := makeTrip(base.Add(-20*time.Minute), trip.StatusActive, &sameTS)
	staleTie2 := makeTrip(base.Add(-15*time.Minute), trip.StatusActive, &sameTS)
	freshPos := base.Add(-10 * time.Second)
	_ = makeTrip(base.Add(-20*time.Minute), trip.StatusActive, &freshPos)
	_ = makeTrip(base.Add(-30*time.Minute), trip.StatusCompleted, nil)

	var (
		collected []model.StaleTrip
		cursor    *model.StaleCursor
	)
	for {
		batch, err := repo.ListStale(ctx, cutoff, cursor, 2)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range batch {
			if item.TripID == staleNoPos.ID || item.TripID == staleTie1.ID || item.TripID == staleTie2.ID {
				collected = append(collected, item)
			}
		}
		if len(batch) < 2 {
			break
		}
		last := batch[len(batch)-1]
		cursor = &model.StaleCursor{LastActivityAt: last.LastActivityAt, TripID: last.TripID}
	}

	if len(collected) != 3 {
		t.Fatalf("collected %d stale trips, want 3: %+v", len(collected), collected)
	}
	if collected[0].TripID != staleNoPos.ID {
		t.Fatalf("first stale trip = %s, want oldest %s", collected[0].TripID, staleNoPos.ID)
	}
	seen := map[uuid.UUID]bool{}
	for _, item := range collected {
		if seen[item.TripID] {
			t.Fatalf("duplicate trip %s across batches", item.TripID)
		}
		seen[item.TripID] = true
	}
	if !seen[staleTie1.ID] || !seen[staleTie2.ID] {
		t.Fatalf("missing tied trips in collected: %+v", collected)
	}
}

package positions_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/goleak"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/model"
	"github.com/Chillobass/fork_template/internal/worker/positions"
)

const (
	metricStaleChecks      = "tripgo.stale_checks"
	metricPositionRequests = "tripgo.position_requests"
	metricRowsRead         = "tripgo.stale_checks.rows_read"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type fakeTicker struct {
	ch chan time.Time
}

func newFakeTicker() *fakeTicker {
	return &fakeTicker{ch: make(chan time.Time)}
}

func (f *fakeTicker) C() <-chan time.Time { return f.ch }
func (f *fakeTicker) Stop()               {}

func (f *fakeTicker) Tick(t *testing.T, at time.Time) {
	t.Helper()
	select {
	case f.ch <- at:
	case <-time.After(3 * time.Second):
		t.Fatal("ticker send timed out")
	}
}

type testEnv struct {
	repo   *MockstaleTripsRepository
	push   *MockpushGateway
	ticker *fakeTicker
	spans  *tracetest.SpanRecorder
	reader *sdkmetric.ManualReader
	poller *positions.Poller
	now    time.Time
}

func newTestEnv(t *testing.T, opts positions.Options) *testEnv {
	t.Helper()
	ctrl := gomock.NewController(t)
	repo := NewMockstaleTripsRepository(ctrl)
	push := NewMockpushGateway(ctrl)
	ft := newFakeTicker()
	spans := tracetest.NewSpanRecorder()
	reader := sdkmetric.NewManualReader()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

	if opts.CheckInterval == 0 {
		opts.CheckInterval = 30 * time.Second
	}
	if opts.StaleAfter == 0 {
		opts.StaleAfter = time.Minute
	}
	if opts.ScanBatchSize == 0 {
		opts.ScanBatchSize = 100
	}
	if opts.PushTimeout == 0 {
		opts.PushTimeout = time.Second
	}

	poller, err := positions.New(
		opts,
		repo,
		push,
		slog.New(slog.DiscardHandler),
		tp,
		mp.Meter("test"),
		func(time.Duration) positions.Ticker { return ft },
		func() time.Time { return now },
	)
	if err != nil {
		t.Fatal(err)
	}

	return &testEnv{
		repo:   repo,
		push:   push,
		ticker: ft,
		spans:  spans,
		reader: reader,
		poller: poller,
		now:    now,
	}
}

func counterValue(t *testing.T, reader *sdkmetric.ManualReader, name, result string) int64 {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(t.Context(), &rm); err != nil {
		t.Fatal(err)
	}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("metric %s is not Sum[int64]", name)
			}
			for _, dp := range sum.DataPoints {
				if result == "" {
					return dp.Value
				}
				v, ok := dp.Attributes.Value("result")
				if ok && v.AsString() == result {
					return dp.Value
				}
			}
		}
	}
	return 0
}

func histogramSum(t *testing.T, reader *sdkmetric.ManualReader, name string) int64 {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(t.Context(), &rm); err != nil {
		t.Fatal(err)
	}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			hist, ok := m.Data.(metricdata.Histogram[int64])
			if !ok {
				t.Fatalf("metric %s is not Histogram[int64]", name)
			}
			if len(hist.DataPoints) > 0 {
				return hist.DataPoints[0].Sum
			}
		}
	}
	return 0
}

func TestDuplicateNotAddedTwice(t *testing.T) {
	env := newTestEnv(t, positions.Options{Workers: 1, QueueSize: 10})
	tripItem := model.StaleTrip{
		TripID:         uuid.New(),
		DriverID:       uuid.New(),
		LastActivityAt: env.now.Add(-5 * time.Minute),
	}

	enteredPush := make(chan struct{})
	releasePush := make(chan struct{})
	thirdScanEntered := make(chan struct{})

	gomock.InOrder(
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).Return([]model.StaleTrip{tripItem}, nil),
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).Return([]model.StaleTrip{tripItem}, nil),
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).DoAndReturn(
			func(context.Context, time.Time, *model.StaleCursor, int) ([]model.StaleTrip, error) {
				close(thirdScanEntered)
				return nil, nil
			},
		),
	)
	env.push.EXPECT().RequestPosition(gomock.Any(), tripItem.TripID, tripItem.DriverID).Times(1).DoAndReturn(
		func(context.Context, uuid.UUID, uuid.UUID) error {
			close(enteredPush)
			<-releasePush
			return nil
		},
	)

	env.poller.Start(t.Context())
	env.ticker.Tick(t, env.now)
	<-enteredPush

	env.ticker.Tick(t, env.now)
	env.ticker.Tick(t, env.now)
	<-thirdScanEntered
	close(releasePush)

	if err := env.poller.Shutdown(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := counterValue(t, env.reader, metricStaleChecks, "queued"); got != 1 {
		t.Fatalf("queued = %d, want 1", got)
	}
	if got := counterValue(t, env.reader, metricStaleChecks, "duplicate"); got != 1 {
		t.Fatalf("duplicate = %d, want 1", got)
	}
}

func TestCancelPreventsPush(t *testing.T) {
	env := newTestEnv(t, positions.Options{Workers: 1, QueueSize: 10})
	blocker := model.StaleTrip{TripID: uuid.New(), DriverID: uuid.New(), LastActivityAt: env.now.Add(-10 * time.Minute)}
	cancelled := model.StaleTrip{TripID: uuid.New(), DriverID: uuid.New(), LastActivityAt: env.now.Add(-5 * time.Minute)}

	blockerStarted := make(chan struct{})
	releaseBlocker := make(chan struct{})
	scanFinished := make(chan struct{})

	env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).Return([]model.StaleTrip{blocker, cancelled}, nil)
	env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).DoAndReturn(
		func(context.Context, time.Time, *model.StaleCursor, int) ([]model.StaleTrip, error) {
			close(scanFinished)
			return nil, nil
		},
	)
	env.push.EXPECT().RequestPosition(gomock.Any(), blocker.TripID, blocker.DriverID).DoAndReturn(
		func(context.Context, uuid.UUID, uuid.UUID) error {
			close(blockerStarted)
			<-releaseBlocker
			return nil
		},
	)

	env.poller.Start(t.Context())
	env.ticker.Tick(t, env.now)
	<-blockerStarted
	env.ticker.Tick(t, env.now)
	<-scanFinished

	if !env.poller.Cancel(t.Context(), cancelled.TripID) {
		t.Fatal("expected queued task to be cancelled")
	}
	close(releaseBlocker)

	if err := env.poller.Shutdown(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := counterValue(t, env.reader, metricPositionRequests, "cancelled"); got != 1 {
		t.Fatalf("cancelled = %d, want 1", got)
	}
	if got := counterValue(t, env.reader, metricPositionRequests, "success"); got != 1 {
		t.Fatalf("success = %d, want 1", got)
	}
}

func TestQueueOverflowDropsWithoutBlockingScanner(t *testing.T) {
	env := newTestEnv(t, positions.Options{Workers: 1, QueueSize: 2, ScanBatchSize: 10})
	items := make([]model.StaleTrip, 5)
	for i := range items {
		items[i] = model.StaleTrip{
			TripID:         uuid.New(),
			DriverID:       uuid.New(),
			LastActivityAt: env.now.Add(-time.Duration(10-i) * time.Minute),
		}
	}

	firstBlocked := make(chan struct{})
	release := make(chan struct{})
	secondTickDone := make(chan struct{})

	gomock.InOrder(
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 10).Return(items, nil),
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 10).DoAndReturn(
			func(context.Context, time.Time, *model.StaleCursor, int) ([]model.StaleTrip, error) {
				close(secondTickDone)
				return nil, nil
			},
		),
	)

	var calls atomic.Int32
	env.push.EXPECT().RequestPosition(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(context.Context, uuid.UUID, uuid.UUID) error {
			if calls.Add(1) == 1 {
				close(firstBlocked)
				<-release
			}
			return nil
		},
	)

	env.poller.Start(t.Context())
	env.ticker.Tick(t, env.now)
	<-firstBlocked

	env.ticker.Tick(t, env.now)
	<-secondTickDone
	close(release)

	if err := env.poller.Shutdown(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := counterValue(t, env.reader, metricStaleChecks, "dropped"); got == 0 {
		t.Fatalf("expected dropped > 0, got %d", got)
	}
}

func TestPoolConcurrencyAndWorkerSurvivesError(t *testing.T) {
	const workers = 3
	const totalTasks = 6

	env := newTestEnv(t, positions.Options{Workers: workers, QueueSize: 20})
	items := make([]model.StaleTrip, totalTasks)
	for i := range items {
		items[i] = model.StaleTrip{
			TripID:         uuid.New(),
			DriverID:       uuid.New(),
			LastActivityAt: env.now.Add(-time.Duration(totalTasks-i+2) * time.Minute),
		}
	}

	var (
		inFlight  atomic.Int32
		maxActive atomic.Int32
		processed atomic.Int32
	)
	allWorkersBusy := make(chan struct{})
	release := make(chan struct{})
	var closeBusyOnce sync.Once

	env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).Return(items, nil)
	env.push.EXPECT().RequestPosition(gomock.Any(), gomock.Any(), gomock.Any()).Times(totalTasks).DoAndReturn(
		func(context.Context, uuid.UUID, uuid.UUID) error {
			cur := inFlight.Add(1)
			for {
				prev := maxActive.Load()
				if cur <= prev || maxActive.CompareAndSwap(prev, cur) {
					break
				}
			}
			if cur == workers {
				closeBusyOnce.Do(func() { close(allWorkersBusy) })
			}
			<-release
			inFlight.Add(-1)
			if processed.Add(1) == 1 {
				return errors.New("push temporary error")
			}
			return nil
		},
	)

	env.poller.Start(t.Context())
	env.ticker.Tick(t, env.now)
	<-allWorkersBusy
	close(release)

	if err := env.poller.Shutdown(t.Context()); err != nil {
		t.Fatal(err)
	}

	if maxActive.Load() != workers {
		t.Fatalf("peak concurrency = %d, want %d", maxActive.Load(), workers)
	}
	if got := counterValue(t, env.reader, metricPositionRequests, "error"); got != 1 {
		t.Fatalf("error count = %d, want 1", got)
	}
	if got := counterValue(t, env.reader, metricPositionRequests, "success"); got != totalTasks-1 {
		t.Fatalf("success count = %d, want %d", got, totalTasks-1)
	}

	var scanSpanID string
	var linkedRequests int

	for _, sp := range env.spans.Ended() {
		switch sp.Name() {
		case "positions.scan":
			scanSpanID = sp.SpanContext().SpanID().String()
		case "positions.request":
			if sp.Parent().IsValid() {
				t.Fatalf("positions.request must not have a parent span, got %v", sp.Parent())
			}
			if len(sp.Links()) == 1 && sp.Links()[0].SpanContext.SpanID().String() != "" {
				linkedRequests++
			}
		}
	}
	if scanSpanID == "" || linkedRequests != totalTasks {
		t.Fatalf("scanSpanID=%q linkedRequests=%d want %d", scanSpanID, linkedRequests, totalTasks)
	}
}

func TestContextCancelStopsAllGoroutines(t *testing.T) {
	env := newTestEnv(t, positions.Options{Workers: 2, QueueSize: 10})
	ctx, cancel := context.WithCancel(t.Context())

	env.poller.Start(ctx)
	cancel()
	env.poller.Wait()
}

func TestBatchScanReadsAllBatches(t *testing.T) {
	env := newTestEnv(t, positions.Options{Workers: 2, QueueSize: 300, ScanBatchSize: 100})
	all := make([]model.StaleTrip, 250)
	for i := range all {
		ts := env.now.Add(-time.Duration(300-i/10) * time.Minute)
		all[i] = model.StaleTrip{TripID: uuid.New(), DriverID: uuid.New(), LastActivityAt: ts}
	}

	c1 := &model.StaleCursor{LastActivityAt: all[99].LastActivityAt, TripID: all[99].TripID}
	c2 := &model.StaleCursor{LastActivityAt: all[199].LastActivityAt, TripID: all[199].TripID}

	done := make(chan struct{})
	var count atomic.Int32

	gomock.InOrder(
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).Return(all[:100], nil),
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), c1, 100).Return(all[100:200], nil),
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), c2, 100).Return(all[200:], nil),
	)
	env.push.EXPECT().RequestPosition(gomock.Any(), gomock.Any(), gomock.Any()).Times(250).DoAndReturn(
		func(context.Context, uuid.UUID, uuid.UUID) error {
			if count.Add(1) == 250 {
				close(done)
			}
			return nil
		},
	)

	env.poller.Start(t.Context())
	env.ticker.Tick(t, env.now)
	<-done

	if err := env.poller.Shutdown(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := histogramSum(t, env.reader, metricRowsRead); got != 250 {
		t.Fatalf("rows_read = %d, want 250", got)
	}
}

func TestBatchScanStopsOnFirstDrop(t *testing.T) {
	env := newTestEnv(t, positions.Options{Workers: 1, QueueSize: 50, ScanBatchSize: 100})
	batch1 := make([]model.StaleTrip, 100)
	for i := range batch1 {
		batch1[i] = model.StaleTrip{
			TripID:         uuid.New(),
			DriverID:       uuid.New(),
			LastActivityAt: env.now.Add(-time.Duration(200-i) * time.Minute),
		}
	}

	firstEntered := make(chan struct{})
	release := make(chan struct{})
	secondTickStarted := make(chan struct{})

	gomock.InOrder(
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).Return(batch1, nil),
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).DoAndReturn(
			func(context.Context, time.Time, *model.StaleCursor, int) ([]model.StaleTrip, error) {
				close(secondTickStarted)
				return nil, nil
			},
		),
	)

	var once sync.Once
	env.push.EXPECT().RequestPosition(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(context.Context, uuid.UUID, uuid.UUID) error {
			once.Do(func() {
				close(firstEntered)
				<-release
			})
			return nil
		},
	)

	env.poller.Start(t.Context())
	env.ticker.Tick(t, env.now)
	<-firstEntered

	env.ticker.Tick(t, env.now)
	<-secondTickStarted
	close(release)

	if err := env.poller.Shutdown(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := counterValue(t, env.reader, metricStaleChecks, "dropped"); got != 1 {
		t.Fatalf("dropped = %d, want 1", got)
	}
}

func TestPriorityOrderInWorkers(t *testing.T) {
	env := newTestEnv(t, positions.Options{Workers: 1, QueueSize: 10})
	blocker := model.StaleTrip{TripID: uuid.New(), DriverID: uuid.New(), LastActivityAt: env.now.Add(-60 * time.Minute)}
	tOldest := model.StaleTrip{TripID: uuid.New(), DriverID: uuid.New(), LastActivityAt: env.now.Add(-30 * time.Minute)}
	tMiddle := model.StaleTrip{TripID: uuid.New(), DriverID: uuid.New(), LastActivityAt: env.now.Add(-20 * time.Minute)}
	tNewest := model.StaleTrip{TripID: uuid.New(), DriverID: uuid.New(), LastActivityAt: env.now.Add(-10 * time.Minute)}

	blockerIn := make(chan struct{})
	release := make(chan struct{})
	scan2Done := make(chan struct{})

	gomock.InOrder(
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).Return([]model.StaleTrip{blocker}, nil),
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).Return([]model.StaleTrip{tMiddle, tNewest, tOldest}, nil),
		env.repo.EXPECT().ListStale(gomock.Any(), env.now.Add(-time.Minute), nil, 100).DoAndReturn(
			func(context.Context, time.Time, *model.StaleCursor, int) ([]model.StaleTrip, error) {
				close(scan2Done)
				return nil, nil
			},
		),
	)

	var (
		mu    sync.Mutex
		order []uuid.UUID
	)
	env.push.EXPECT().RequestPosition(gomock.Any(), gomock.Any(), gomock.Any()).Times(4).DoAndReturn(
		func(_ context.Context, tripID, _ uuid.UUID) error {
			if tripID == blocker.TripID {
				close(blockerIn)
				<-release
				return nil
			}
			mu.Lock()
			order = append(order, tripID)
			mu.Unlock()
			return nil
		},
	)

	env.poller.Start(t.Context())
	env.ticker.Tick(t, env.now)
	<-blockerIn

	env.ticker.Tick(t, env.now)
	env.ticker.Tick(t, env.now)
	<-scan2Done
	close(release)

	if err := env.poller.Shutdown(t.Context()); err != nil {
		t.Fatal(err)
	}

	want := []uuid.UUID{tOldest.TripID, tMiddle.TripID, tNewest.TripID}
	if len(order) != len(want) {
		t.Fatalf("order len = %d, want %d", len(order), len(want))
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order[%d] = %s, want %s", i, order[i], want[i])
		}
	}
}

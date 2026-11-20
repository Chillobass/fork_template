package positions

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/Chillobass/fork_template/internal/model"
)

type Options struct {
	CheckInterval time.Duration
	StaleAfter    time.Duration
	Workers       int
	QueueSize     int
	ScanBatchSize int
	PushTimeout   time.Duration
}

type Poller struct {
	opts      Options
	repo      staleTripsRepository
	push      pushGateway
	logger    *slog.Logger
	tracer    trace.Tracer
	metrics   *metrics
	newTicker func(time.Duration) Ticker
	now       func() time.Time

	queue         *priorityQueue
	activeWorkers atomic.Int64

	startOnce    sync.Once
	shutdownOnce sync.Once
	scanCancel   context.CancelFunc
	workerCancel context.CancelFunc
	scannerDone  chan struct{}
	workersDone  chan struct{}
}

func New(
	opts Options,
	repo staleTripsRepository,
	push pushGateway,
	logger *slog.Logger,
	tracerProvider trace.TracerProvider,
	meter metric.Meter,
	newTicker func(time.Duration) Ticker,
	now func() time.Time,
) (*Poller, error) {
	p := &Poller{
		opts:        opts,
		repo:        repo,
		push:        push,
		logger:      logger,
		tracer:      tracerProvider.Tracer("worker/positions"),
		newTicker:   newTicker,
		now:         now,
		queue:       newPriorityQueue(opts.QueueSize),
		scannerDone: make(chan struct{}),
		workersDone: make(chan struct{}),
	}

	m, err := newMetrics(meter, p.queue.Len, p.activeWorkers.Load)
	if err != nil {
		return nil, err
	}
	p.metrics = m

	return p, nil
}

func (p *Poller) Start(ctx context.Context) {
	p.startOnce.Do(func() {
		scanCtx, scanCancel := context.WithCancel(ctx)
		workerCtx, workerCancel := context.WithCancel(context.WithoutCancel(ctx))
		p.scanCancel = scanCancel
		p.workerCancel = workerCancel

		p.logger.InfoContext(ctx, "positions poller started",
			"workers", p.opts.Workers,
			"queue_size", p.opts.QueueSize,
			"batch_size", p.opts.ScanBatchSize,
			"check_interval", p.opts.CheckInterval.String(),
			"stale_after", p.opts.StaleAfter.String(),
		)

		var workersWG sync.WaitGroup
		for range p.opts.Workers {
			workersWG.Go(func() {
				p.runWorker(workerCtx)
			})
		}

		go func() {
			workersWG.Wait()
			workerCancel()
			close(p.workersDone)
		}()

		go func() {
			defer func() {
				p.queue.Close()
				close(p.scannerDone)
			}()
			p.runScanner(scanCtx)
		}()
	})
}

func (p *Poller) Shutdown(ctx context.Context) error {
	var err error
	p.shutdownOnce.Do(func() {
		if p.scanCancel == nil {
			return
		}
		p.logger.InfoContext(ctx, "positions poller stopping")
		p.scanCancel()

		select {
		case <-p.scannerDone:
		case <-ctx.Done():
			p.queue.Close()
			p.workerCancel()
			<-p.scannerDone
			<-p.workersDone
			err = fmt.Errorf("stop positions scanner: %w", ctx.Err())
			return
		}

		select {
		case <-p.workersDone:
			p.logger.InfoContext(ctx, "positions poller stopped")
		case <-ctx.Done():
			p.logger.WarnContext(ctx, "positions poller drain timed out, cancelling workers")
			p.workerCancel()
			<-p.workersDone
			err = fmt.Errorf("drain positions workers: %w", ctx.Err())
		}
	})
	return err
}

func (p *Poller) Wait() {
	if p.scanCancel == nil {
		return
	}
	<-p.scannerDone
	<-p.workersDone
}

func (p *Poller) Cancel(ctx context.Context, tripID uuid.UUID) bool {
	if !p.queue.Cancel(tripID) {
		return false
	}
	p.metrics.recordCancelled(ctx)
	return true
}

func (p *Poller) runScanner(ctx context.Context) {
	ticker := p.newTicker(p.opts.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C():
			if ctx.Err() != nil {
				return
			}
			p.scanOnce(ctx)
		}
	}
}

func (p *Poller) scanOnce(ctx context.Context) {
	scanCtx, span := p.tracer.Start(ctx, "positions.scan", trace.WithNewRoot())
	defer span.End()

	staleBefore := p.now().UTC().Truncate(time.Microsecond).Add(-p.opts.StaleAfter)
	var (
		cursor    *model.StaleCursor
		totalRead int64
	)

	defer func() {
		p.metrics.recordRowsRead(scanCtx, totalRead)
		span.SetAttributes(attribute.Int64("positions.scan.rows_read", totalRead))
	}()

	for {
		if scanCtx.Err() != nil {
			return
		}

		batch, err := p.repo.ListStale(scanCtx, staleBefore, cursor, p.opts.ScanBatchSize)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "")
			p.logger.WarnContext(scanCtx, "stale trips scan failed", "error", err)
			return
		}

		totalRead += int64(len(batch))
		if len(batch) == 0 {
			return
		}

		if stop := p.enqueueBatch(scanCtx, span.SpanContext(), batch); stop {
			return
		}
		if len(batch) < p.opts.ScanBatchSize {
			return
		}

		last := batch[len(batch)-1]
		cursor = &model.StaleCursor{
			LastActivityAt: last.LastActivityAt,
			TripID:         last.TripID,
		}
	}
}

func (p *Poller) enqueueBatch(ctx context.Context, scanSpanCtx trace.SpanContext, batch []model.StaleTrip) bool {
	for _, item := range batch {
		res := p.queue.Enqueue(task{
			TripID:         item.TripID,
			DriverID:       item.DriverID,
			LastActivityAt: item.LastActivityAt,
			SpanContext:    scanSpanCtx,
		})

		switch res.outcome {
		case outcomeQueued:
			p.metrics.recordStaleCheck(ctx, resultQueued)
		case outcomeDuplicate:
			p.metrics.recordStaleCheck(ctx, resultDuplicate)
		case outcomeEvicted:
			p.metrics.recordStaleCheck(ctx, resultQueued)
			p.metrics.recordStaleCheck(ctx, resultDropped)
			p.logger.WarnContext(ctx, "position request evicted from full queue",
				"evicted_trip_id", res.evictedTripID,
				"queued_trip_id", item.TripID,
			)
		case outcomeDropped:
			p.metrics.recordStaleCheck(ctx, resultDropped)
			p.logger.WarnContext(ctx, "position request queue full, task dropped", "trip_id", item.TripID)
			return true
		case outcomeClosed:
			return true
		}
	}
	return false
}

func (p *Poller) runWorker(ctx context.Context) {
	for {
		item, ok := p.queue.Pop()
		if !ok {
			return
		}
		if ctx.Err() != nil {
			p.queue.Done(item.TripID)
			return
		}
		p.processTask(ctx, item)
	}
}

func (p *Poller) processTask(ctx context.Context, item task) {
	p.activeWorkers.Add(1)
	defer func() {
		p.activeWorkers.Add(-1)
		p.queue.Done(item.TripID)
	}()

	opts := []trace.SpanStartOption{
		trace.WithNewRoot(),
		trace.WithAttributes(
			attribute.String("trip.id", item.TripID.String()),
			attribute.String("driver.id", item.DriverID.String()),
		),
	}
	if item.SpanContext.IsValid() {
		opts = append(opts, trace.WithLinks(trace.Link{SpanContext: item.SpanContext}))
	}

	taskCtx, span := p.tracer.Start(ctx, "positions.request", opts...)
	defer span.End()

	reqCtx, cancel := context.WithTimeout(taskCtx, p.opts.PushTimeout)
	defer cancel()

	started := time.Now()
	err := p.push.RequestPosition(reqCtx, item.TripID, item.DriverID)
	elapsed := time.Since(started)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "")
		p.metrics.recordRequest(taskCtx, resultError, elapsed)
		p.logger.WarnContext(taskCtx, "push service request failed",
			"trip_id", item.TripID,
			"driver_id", item.DriverID,
			"error", err,
		)
		return
	}

	p.metrics.recordRequest(taskCtx, resultSuccess, elapsed)
}

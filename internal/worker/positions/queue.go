package positions

import (
	"bytes"
	"container/heap"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type task struct {
	TripID         uuid.UUID
	DriverID       uuid.UUID
	LastActivityAt time.Time
	SpanContext    trace.SpanContext
}

func (t task) isOlderThan(other task) bool {
	if !t.LastActivityAt.Equal(other.LastActivityAt) {
		return t.LastActivityAt.Before(other.LastActivityAt)
	}
	return bytes.Compare(t.TripID[:], other.TripID[:]) < 0
}

type enqueueOutcome int

const (
	outcomeQueued enqueueOutcome = iota
	outcomeDuplicate
	outcomeDropped
	outcomeEvicted
	outcomeClosed
)

type enqueueResult struct {
	outcome       enqueueOutcome
	evictedTripID uuid.UUID
}

type queueItem struct {
	task   task
	minIdx int
	maxIdx int
}

type minTaskHeap []*queueItem

func (h minTaskHeap) Len() int           { return len(h) }
func (h minTaskHeap) Less(i, j int) bool { return h[i].task.isOlderThan(h[j].task) }
func (h minTaskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].minIdx = i
	h[j].minIdx = j
}

func (h *minTaskHeap) Push(x any) {
	item := x.(*queueItem)
	item.minIdx = len(*h)
	*h = append(*h, item)
}

func (h *minTaskHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.minIdx = -1
	*h = old[:n-1]
	return item
}

type maxTaskHeap []*queueItem

func (h maxTaskHeap) Len() int           { return len(h) }
func (h maxTaskHeap) Less(i, j int) bool { return h[j].task.isOlderThan(h[i].task) }
func (h maxTaskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].maxIdx = i
	h[j].maxIdx = j
}

func (h *maxTaskHeap) Push(x any) {
	item := x.(*queueItem)
	item.maxIdx = len(*h)
	*h = append(*h, item)
}

func (h *maxTaskHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.maxIdx = -1
	*h = old[:n-1]
	return item
}

type priorityQueue struct {
	mu       sync.Mutex
	cond     *sync.Cond
	capacity int
	closed   bool

	min      minTaskHeap
	max      maxTaskHeap
	pending  map[uuid.UUID]*queueItem
	inFlight map[uuid.UUID]struct{}
}

func newPriorityQueue(capacity int) *priorityQueue {
	q := &priorityQueue{
		capacity: capacity,
		min:      make(minTaskHeap, 0, capacity),
		max:      make(maxTaskHeap, 0, capacity),
		pending:  make(map[uuid.UUID]*queueItem, capacity),
		inFlight: make(map[uuid.UUID]struct{}),
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *priorityQueue) Enqueue(t task) enqueueResult {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return enqueueResult{outcome: outcomeClosed}
	}
	if _, ok := q.pending[t.TripID]; ok {
		return enqueueResult{outcome: outcomeDuplicate}
	}
	if _, ok := q.inFlight[t.TripID]; ok {
		return enqueueResult{outcome: outcomeDuplicate}
	}

	if len(q.pending) < q.capacity {
		q.pushLocked(t)
		q.cond.Signal()
		return enqueueResult{outcome: outcomeQueued}
	}

	newest := q.max[0]
	if !t.isOlderThan(newest.task) {
		return enqueueResult{outcome: outcomeDropped}
	}

	evictedID := newest.task.TripID
	q.removeLocked(newest)
	q.pushLocked(t)
	q.cond.Signal()
	return enqueueResult{outcome: outcomeEvicted, evictedTripID: evictedID}
}

func (q *priorityQueue) Pop() (task, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.pending) == 0 && !q.closed {
		q.cond.Wait()
	}
	if len(q.pending) == 0 {
		return task{}, false
	}

	oldest := q.min[0]
	q.removeLocked(oldest)
	q.inFlight[oldest.task.TripID] = struct{}{}
	return oldest.task, true
}

func (q *priorityQueue) Done(tripID uuid.UUID) {
	q.mu.Lock()
	delete(q.inFlight, tripID)
	q.mu.Unlock()
}

func (q *priorityQueue) Cancel(tripID uuid.UUID) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, ok := q.pending[tripID]
	if !ok {
		return false
	}
	q.removeLocked(item)
	return true
}

func (q *priorityQueue) Close() {
	q.mu.Lock()
	if !q.closed {
		q.closed = true
		q.cond.Broadcast()
	}
	q.mu.Unlock()
}

func (q *priorityQueue) Len() int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return int64(len(q.pending))
}

func (q *priorityQueue) pushLocked(t task) {
	item := &queueItem{task: t}
	heap.Push(&q.min, item)
	heap.Push(&q.max, item)
	q.pending[t.TripID] = item
}

func (q *priorityQueue) removeLocked(item *queueItem) {
	heap.Remove(&q.min, item.minIdx)
	heap.Remove(&q.max, item.maxIdx)
	delete(q.pending, item.task.TripID)
}

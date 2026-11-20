package positions

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPriorityQueueOrderAndEviction(t *testing.T) {
	base := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	oldest := task{TripID: uuid.New(), LastActivityAt: base.Add(-30 * time.Minute)}
	middle := task{TripID: uuid.New(), LastActivityAt: base.Add(-20 * time.Minute)}
	newest := task{TripID: uuid.New(), LastActivityAt: base.Add(-10 * time.Minute)}

	q := newPriorityQueue(3)
	if res := q.Enqueue(middle); res.outcome != outcomeQueued {
		t.Fatalf("enqueue middle: %+v", res)
	}
	if res := q.Enqueue(newest); res.outcome != outcomeQueued {
		t.Fatalf("enqueue newest: %+v", res)
	}
	if res := q.Enqueue(oldest); res.outcome != outcomeQueued {
		t.Fatalf("enqueue oldest: %+v", res)
	}

	if res := q.Enqueue(middle); res.outcome != outcomeDuplicate {
		t.Fatalf("duplicate pending: %+v", res)
	}

	fresher := task{TripID: uuid.New(), LastActivityAt: base.Add(-5 * time.Minute)}
	if res := q.Enqueue(fresher); res.outcome != outcomeDropped {
		t.Fatalf("fresher on full queue: %+v", res)
	}

	veryOld := task{TripID: uuid.New(), LastActivityAt: base.Add(-40 * time.Minute)}
	if res := q.Enqueue(veryOld); res.outcome != outcomeEvicted || res.evictedTripID != newest.TripID {
		t.Fatalf("eviction result = %+v, want evicted %s", res, newest.TripID)
	}

	if !q.Cancel(middle.TripID) {
		t.Fatal("expected middle to be cancelled")
	}
	if q.Cancel(middle.TripID) {
		t.Fatal("second cancel must return false")
	}

	got1, ok := q.Pop()
	if !ok || got1.TripID != veryOld.TripID {
		t.Fatalf("pop 1 = (%+v, %v), want %s", got1, ok, veryOld.TripID)
	}
	if res := q.Enqueue(veryOld); res.outcome != outcomeDuplicate {
		t.Fatalf("duplicate in-flight: %+v", res)
	}
	q.Done(got1.TripID)

	got2, ok := q.Pop()
	if !ok || got2.TripID != oldest.TripID {
		t.Fatalf("pop 2 = (%+v, %v), want %s", got2, ok, oldest.TripID)
	}
	q.Done(got2.TripID)

	q.Close()
	if _, ok := q.Pop(); ok {
		t.Fatal("expected closed empty queue to return false")
	}
}

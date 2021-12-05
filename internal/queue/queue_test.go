package queue

import (
	"context"
	"strconv"
	"testing"
	"time"
)

func TestSimpleEnqueueDequeue(t *testing.T) {
	q := New(context.Background())

	for i := 0; i < 3; i++ {
		q.Enqueue(&Job{ID: strconv.Itoa(i)})
	}

	for i := 2; i >= 0; i-- {
		j, err := q.Dequeue()
		if err != nil {
			t.Errorf("error on dequeue: %v", err)
		}
		if j.ID != strconv.Itoa(i) {
			t.Errorf("dequeue expected: %d got %s", i, j.ID)
		}
		if err := q.Conclude(j.ID); err != nil {
			t.Errorf("could not conclude job: %v", err)
		}
	}

	if _, err := q.Dequeue(); err != QueueEmpty {
		t.Errorf("(%v).Dequeue err is %v, expected %v", q, err, QueueEmpty)
	}
}

func TestCannotConcludedUnknownJob(t *testing.T) {
	q := New(context.Background())

	if err := q.Conclude("unknown"); err != JobNotFound {
		t.Errorf("expected: %v, got: %v", JobNotFound, err)
	}
}

func TestAbandonedJobReturnsToQueue(t *testing.T) {
	q := New(context.Background())

	q.Enqueue(&Job{ID: "test"})
	if j, err := q.Dequeue(); err != nil || j.ID != "test" {
		t.Errorf("did not dequeue test job")
	}

	if _, err := q.Dequeue(); err != QueueEmpty {
		t.Errorf("expected the queue to be empty")
	}

	time_Now = func() time.Time { return time.Now().Add(10 * time.Minute) }
	q.checkExpiredJobs()

	if j, err := q.Dequeue(); err != nil || j.ID != "test" {
		t.Errorf("did not dequeue test job")
	}
}

func TestConcludeJob(t *testing.T) {
	q := New(context.Background())

	q.Enqueue(&Job{ID: "test"})
	if j, err := q.Dequeue(); err != nil || j.ID != "test" {
		t.Errorf("did not dequeue test job")
	}

	if err := q.Conclude("test"); err != nil {
		t.Errorf("error concluding job: %v", err)
	}

	time_Now = func() time.Time { return time.Now().Add(10 * time.Minute) }
	q.checkExpiredJobs()

	if _, err := q.Dequeue(); err != QueueEmpty {
		t.Errorf("expected the queue to be empty")
	}
}

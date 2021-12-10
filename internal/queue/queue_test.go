package queue

import (
	"context"
	"testing"
	"time"
)

func TestSimpleEnqueueDequeue(t *testing.T) {
	q := New(context.Background())
	var jobs []*Job

	for i := 0; i < 10; i++ {
		j := &Job{Type: JobType_TimeCritical}
		q.Enqueue(j)
		jobs = append(jobs, j)
	}

	for _, j := range jobs {
		d, err := q.Dequeue()
		if err != nil {
			t.Errorf("error on dequeue: %v", err)
		}
		if d.ID != j.ID {
			t.Errorf("dequeue expected: %s got %s", j.ID, d.ID)
		}
		if _, err := q.Conclude(d.ID); err != nil {
			t.Errorf("could not conclude job: %v", err)
		}
	}

	if _, err := q.Dequeue(); err != QueueEmpty {
		t.Errorf("expected the queue to be empty")
	}
}

func TestPriorityWithTwoJobs(t *testing.T) {
	q := New(context.Background())

	j1 := &Job{Type: JobType_TimeCritical, Priority: 1}
	q.Enqueue(j1)

	j2 := &Job{Type: JobType_TimeCritical, Priority: 2}
	q.Enqueue(j2)

	d1, err := q.Dequeue()
	if err != nil {
		t.Errorf("error on dequeue: %v", err)
	}
	if d1.ID != j2.ID {
		t.Errorf("Dequeue expected %v got %v", j2.ID, d1.ID)
	}

	d2, err := q.Dequeue()
	if err != nil {
		t.Errorf("error on dequeue: %v", err)
	}
	if d2.ID != j1.ID {
		t.Errorf("Dequeue expected %v got %v", j1.ID, d2.ID)
	}
}

func TestCannotConcludedUnknownJob(t *testing.T) {
	q := New(context.Background())

	if _, err := q.Conclude("unknown"); err != JobNotFound {
		t.Errorf("expected: %v, got: %v", JobNotFound, err)
	}
}

func TestAbandonedJobReturnsToQueue(t *testing.T) {
	q := New(context.Background())

	j := &Job{Type: JobType_TimeCritical}
	q.Enqueue(j)
	if d, err := q.Dequeue(); err != nil || d.ID != j.ID {
		t.Errorf("did not dequeue test job")
	}

	if _, err := q.Dequeue(); err != QueueEmpty {
		t.Errorf("expected the queue to be empty")
	}

	time_Now = func() time.Time { return time.Now().Add(10 * time.Minute) }
	q.checkExpiredJobs()

	if d, err := q.Dequeue(); err != nil || d.ID != j.ID {
		t.Errorf("did not dequeue test job")
	}
}

func TestConcludeJob(t *testing.T) {
	q := New(context.Background())

	j := &Job{Type: JobType_TimeCritical}
	q.Enqueue(j)
	if d, err := q.Dequeue(); err != nil || d.ID != j.ID {
		t.Errorf("did not dequeue test job")
	}

	if _, err := q.Conclude(j.ID); err != nil {
		t.Errorf("error concluding job: %v", err)
	}

	time_Now = func() time.Time { return time.Now().Add(10 * time.Minute) }
	q.checkExpiredJobs()

	if _, err := q.Dequeue(); err != QueueEmpty {
		t.Errorf("expected the queue to be empty")
	}
}

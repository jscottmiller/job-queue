package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var time_Now = time.Now

type Queue struct {
	sync.Mutex
	byID     map[string]*Job
	pending  []*Job
	head     int
	entries  int
	dequeued []dequeuedJob
}

type Job struct {
	ID     string
	Status JobStatus
	Type   JobType
}

type JobType string

const (
	JobType_TimeCritical    JobType = "TIME_CRITICAL"
	JobType_NotTimeCritical         = "NOT_TIME_CRITICAL"
)

type JobStatus string

const (
	JobStatus_Queued     JobStatus = "QUEUED"
	JobStatus_InProgress           = "IN_PROGRESS"
	JobStatus_Concluded            = "CONCLUDED"
)

type dequeuedJob struct {
	job  *Job
	time time.Time
}

var QueueEmpty = fmt.Errorf("the queue is empty")
var JobNotFound = fmt.Errorf("the job was not found")
var JobNotInProgress = fmt.Errorf("the job must be in progress to be concluded")

func New(ctx context.Context) *Queue {
	q := &Queue{
		byID: make(map[string]*Job),
	}

	go func() {
		t := time.NewTicker(1 * time.Minute)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				q.checkExpiredJobs()

			case <-ctx.Done():
				return
			}

		}
	}()

	return q
}

func (q *Queue) GetJob(id string) (j *Job) {
	q.Lock()
	defer q.Unlock()

	j, ok := q.byID[id]
	if !ok {
		return nil
	}
	return j
}

func (q *Queue) Enqueue(j *Job) {
	q.Lock()
	defer q.Unlock()

	j.ID = uuid.New().String()
	j.Status = JobStatus_Queued

	q.byID[j.ID] = j

	q.pending = append(q.pending, j)
	q.entries++
}

func (q *Queue) Dequeue() (*Job, error) {
	q.Lock()
	defer q.Unlock()

	if q.entries == 0 {
		return nil, QueueEmpty
	}

	first := q.pending[q.head]
	q.pending[q.head] = nil
	first.Status = JobStatus_InProgress
	q.head++
	q.entries--

	// Compact the queue if it becomes too sparse
	// TODO: Fill constant could become a configuration parameter to the queue
	if float64(q.entries)/float64(len(q.pending)) < 0.50 {
		extent := len(q.pending) - q.head
		copy(q.pending, q.pending[q.head:])
		q.pending = q.pending[:extent]
		q.head = 0
	}

	q.dequeued = append(q.dequeued, dequeuedJob{job: first, time: time_Now()})

	return first, nil
}

func (q *Queue) Conclude(id string) (*Job, error) {
	q.Lock()
	defer q.Unlock()

	j, ok := q.byID[id]
	if !ok {
		return nil, JobNotFound
	}

	if j.Status != JobStatus_InProgress {
		return nil, JobNotInProgress
	}

	j.Status = JobStatus_Concluded
	delete(q.byID, id)

	for i, deq := range q.dequeued {
		if j == deq.job {
			q.dequeued[i].job = nil
			break
		}
	}

	return j, nil
}

func (q *Queue) checkExpiredJobs() {
	q.Lock()
	defer q.Unlock()

	n := time_Now()

	// Find any dequeued jobs that have been inactive for more than 5 minutes and
	// return them to the queue
	for i, dequeuedJob := range q.dequeued {
		if dequeuedJob.job == nil {
			continue
		}
		if n.Sub(dequeuedJob.time) < 5*time.Minute {
			break
		}
		j := dequeuedJob.job
		j.Status = JobStatus_Queued
		q.pending = append(q.pending, j)
		q.entries++
		q.dequeued[i].job = nil
	}

	// Compact the dequeued jobs slice
	var insert int
	for current := 0; current < len(q.dequeued); current++ {
		if q.dequeued[current].job != nil {
			q.dequeued[insert] = q.dequeued[current]
			insert++
		}
	}
	q.dequeued = q.dequeued[:insert]
}

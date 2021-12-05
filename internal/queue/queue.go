package queue

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var time_Now = time.Now

type Queue struct {
	sync.Mutex
	byID     map[string]*Job
	pending  []*Job
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
	j, ok := q.byID[id]
	if !ok {
		return nil
	}
	return j
}

func (q *Queue) Enqueue(j *Job) {
	q.Lock()
	defer q.Unlock()

	q.byID[j.ID] = j
	q.pending = append(q.pending, j)
	j.Status = JobStatus_Queued
}

func (q *Queue) Dequeue() (*Job, error) {
	q.Lock()
	defer q.Unlock()

	if len(q.pending) == 0 {
		return nil, QueueEmpty
	}

	var last *Job
	last, q.pending = q.pending[len(q.pending)-1], q.pending[:len(q.pending)-1]
	q.dequeued = append(q.dequeued, dequeuedJob{job: last, time: time_Now()})
	last.Status = JobStatus_InProgress

	return last, nil
}

func (q *Queue) Conclude(id string) error {
	q.Lock()
	defer q.Unlock()

	j, ok := q.byID[id]
	if !ok {
		return JobNotFound
	}

	if j.Status != JobStatus_InProgress {
		return JobNotInProgress
	}

	delete(q.byID, id)

	for i, deq := range q.dequeued {
		if j == deq.job {
			q.dequeued[i].job = nil
			break
		}
	}

	return nil
}

func (q *Queue) checkExpiredJobs() {
	q.Lock()
	defer q.Unlock()

	n := time_Now()

	for _, dequeuedJob := range q.dequeued {
		if dequeuedJob.job == nil {
			continue
		}
		if n.Sub(dequeuedJob.time) < 5*time.Minute {
			break
		}
		j := dequeuedJob.job
		q.pending = append(q.pending, j)
		j.Status = JobStatus_Queued
		dequeuedJob.job = nil
	}

	var insert int
	for current := 0; current < len(q.dequeued); current++ {
		if q.dequeued[current].job != nil {
			q.dequeued[insert] = q.dequeued[current]
			insert++
		}
	}
	q.dequeued = q.dequeued[:insert]
}

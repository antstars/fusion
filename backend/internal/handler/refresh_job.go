package handler

import (
	"fmt"
	"sync"
	"time"
)

const maxRefreshJobs = 100

type refreshJobStatus string
type refreshJobScope string

const (
	refreshJobStatusRunning   refreshJobStatus = "running"
	refreshJobStatusCompleted refreshJobStatus = "completed"
	refreshJobStatusFailed    refreshJobStatus = "failed"

	refreshJobScopeAll  refreshJobScope = "all"
	refreshJobScopeFeed refreshJobScope = "feed"
)

type refreshJob struct {
	ID         string           `json:"id"`
	Scope      refreshJobScope  `json:"scope"`
	FeedID     *int64           `json:"feed_id,omitempty"`
	Status     refreshJobStatus `json:"status"`
	StartedAt  int64            `json:"started_at"`
	FinishedAt int64            `json:"finished_at,omitempty"`
	Error      string           `json:"error,omitempty"`
}

type refreshJobStore struct {
	mu          sync.Mutex
	nextID      int64
	jobs        map[string]*refreshJob
	order       []string
	runningAll  string
	runningFeed map[int64]string
}

func newRefreshJobStore() *refreshJobStore {
	return &refreshJobStore{
		jobs:        make(map[string]*refreshJob),
		runningFeed: make(map[int64]string),
	}
}

func (s *refreshJobStore) startAll() (*refreshJob, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.runningAll != "" {
		if job, ok := s.jobs[s.runningAll]; ok && job.Status == refreshJobStatusRunning {
			return cloneRefreshJob(job), false
		}
		s.runningAll = ""
	}

	job := s.newLocked(refreshJobScopeAll, nil)
	s.runningAll = job.ID
	return cloneRefreshJob(job), true
}

func (s *refreshJobStore) startFeed(feedID int64) (*refreshJob, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if jobID := s.runningFeed[feedID]; jobID != "" {
		if job, ok := s.jobs[jobID]; ok && job.Status == refreshJobStatusRunning {
			return cloneRefreshJob(job), false
		}
		delete(s.runningFeed, feedID)
	}

	job := s.newLocked(refreshJobScopeFeed, &feedID)
	s.runningFeed[feedID] = job.ID
	return cloneRefreshJob(job), true
}

func (s *refreshJobStore) get(id string) (*refreshJob, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[id]
	if !ok {
		return nil, false
	}
	return cloneRefreshJob(job), true
}

func (s *refreshJobStore) finish(id string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[id]
	if !ok {
		return
	}

	job.FinishedAt = time.Now().Unix()
	if err != nil {
		job.Status = refreshJobStatusFailed
		job.Error = err.Error()
	} else {
		job.Status = refreshJobStatusCompleted
	}

	if job.Scope == refreshJobScopeAll && s.runningAll == id {
		s.runningAll = ""
	}
	if job.Scope == refreshJobScopeFeed && job.FeedID != nil {
		if s.runningFeed[*job.FeedID] == id {
			delete(s.runningFeed, *job.FeedID)
		}
	}
}

func (s *refreshJobStore) newLocked(scope refreshJobScope, feedID *int64) *refreshJob {
	s.nextID++
	id := fmt.Sprintf("%d", s.nextID)
	job := &refreshJob{
		ID:        id,
		Scope:     scope,
		FeedID:    feedID,
		Status:    refreshJobStatusRunning,
		StartedAt: time.Now().Unix(),
	}
	s.jobs[id] = job
	s.order = append(s.order, id)
	s.pruneLocked()
	return job
}

func (s *refreshJobStore) pruneLocked() {
	attempts := len(s.order)
	for len(s.order) > maxRefreshJobs && attempts > 0 {
		attempts--
		id := s.order[0]
		s.order = s.order[1:]

		job, ok := s.jobs[id]
		if !ok || job.Status == refreshJobStatusRunning {
			s.order = append(s.order, id)
			continue
		}

		delete(s.jobs, id)
	}
}

func cloneRefreshJob(job *refreshJob) *refreshJob {
	if job == nil {
		return nil
	}

	clone := *job
	if job.FeedID != nil {
		feedID := *job.FeedID
		clone.FeedID = &feedID
	}
	return &clone
}

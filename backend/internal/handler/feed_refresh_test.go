package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/cache"
	"github.com/0x2E/fusion/internal/config"
	"github.com/gin-gonic/gin"
)

type refreshJobTestPuller struct {
	allStarted  chan struct{}
	allRelease  chan struct{}
	feedStarted chan int64
	feedRelease chan struct{}
	allErr      error
	feedErr     error
}

func newRefreshJobTestPuller() *refreshJobTestPuller {
	return &refreshJobTestPuller{
		allStarted:  make(chan struct{}, 1),
		allRelease:  make(chan struct{}),
		feedStarted: make(chan int64, 1),
		feedRelease: make(chan struct{}),
	}
}

func (p *refreshJobTestPuller) RefreshAll(context.Context) (int, error) {
	p.allStarted <- struct{}{}
	<-p.allRelease
	return 1, p.allErr
}

func (p *refreshJobTestPuller) RefreshFeed(_ context.Context, feedID int64) error {
	p.feedStarted <- feedID
	<-p.feedRelease
	return p.feedErr
}

func TestRefreshAllJobLifecycle(t *testing.T) {
	puller := newRefreshJobTestPuller()
	h := &Handler{
		config:      &config.Config{},
		cache:       cache.NoopCache{},
		puller:      puller,
		refreshJobs: newRefreshJobStore(),
	}
	r := newTestRouter()
	r.POST("/api/feeds/refresh", h.refreshAllFeeds)
	r.GET("/api/feeds/refresh-jobs/:id", h.getRefreshJob)

	first := performRequest(r, http.MethodPost, "/api/feeds/refresh", nil, nil)
	if first.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", first.Code)
	}
	firstJob := decodeRefreshJobResponse(t, first.Body.Bytes())
	if firstJob.Status != refreshJobStatusRunning {
		t.Fatalf("expected running job, got %#v", firstJob)
	}

	select {
	case <-puller.allStarted:
	case <-time.After(time.Second):
		t.Fatal("refresh all did not start")
	}

	second := performRequest(r, http.MethodPost, "/api/feeds/refresh", nil, nil)
	secondJob := decodeRefreshJobResponse(t, second.Body.Bytes())
	if secondJob.ID != firstJob.ID {
		t.Fatalf("expected duplicate refresh-all to reuse job %q, got %q", firstJob.ID, secondJob.ID)
	}

	close(puller.allRelease)
	waitForRefreshJobStatus(t, r, firstJob.ID, refreshJobStatusCompleted)
}

func TestRefreshFeedJobStoreLifecycle(t *testing.T) {
	puller := newRefreshJobTestPuller()
	h := &Handler{
		config:      &config.Config{},
		cache:       cache.NoopCache{},
		puller:      puller,
		refreshJobs: newRefreshJobStore(),
	}
	r := newTestRouter()
	r.POST("/api/test-refresh-feed", func(c *gin.Context) {
		job, shouldStart := h.refreshJobs.startFeed(42)
		c.JSON(http.StatusAccepted, gin.H{"data": job})
		if !shouldStart {
			return
		}
		go func() {
			err := h.puller.RefreshFeed(context.Background(), 42)
			h.refreshJobs.finish(job.ID, err)
		}()
	})
	r.GET("/api/feeds/refresh-jobs/:id", h.getRefreshJob)

	first := performRequest(r, http.MethodPost, "/api/test-refresh-feed", nil, nil)
	firstJob := decodeRefreshJobResponse(t, first.Body.Bytes())
	if firstJob.FeedID == nil || *firstJob.FeedID != 42 {
		t.Fatalf("expected feed_id 42, got %#v", firstJob)
	}

	select {
	case feedID := <-puller.feedStarted:
		if feedID != 42 {
			t.Fatalf("expected feed 42, got %d", feedID)
		}
	case <-time.After(time.Second):
		t.Fatal("refresh feed did not start")
	}

	second := performRequest(r, http.MethodPost, "/api/test-refresh-feed", nil, nil)
	secondJob := decodeRefreshJobResponse(t, second.Body.Bytes())
	if secondJob.ID != firstJob.ID {
		t.Fatalf("expected duplicate feed refresh to reuse job %q, got %q", firstJob.ID, secondJob.ID)
	}

	close(puller.feedRelease)
	waitForRefreshJobStatus(t, r, firstJob.ID, refreshJobStatusCompleted)
}

func TestRefreshJobFailedStatus(t *testing.T) {
	jobStore := newRefreshJobStore()
	job, _ := jobStore.startAll()
	jobStore.finish(job.ID, errors.New("boom"))

	h := &Handler{refreshJobs: jobStore}
	r := newTestRouter()
	r.GET("/api/feeds/refresh-jobs/:id", h.getRefreshJob)

	w := performRequest(r, http.MethodGet, "/api/feeds/refresh-jobs/"+job.ID, nil, nil)
	got := decodeRefreshJobResponse(t, w.Body.Bytes())
	if got.Status != refreshJobStatusFailed || got.Error != "boom" {
		t.Fatalf("expected failed boom, got %#v", got)
	}
}

func decodeRefreshJobResponse(t *testing.T, body []byte) refreshJob {
	t.Helper()

	var payload struct {
		Data refreshJob `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode refresh job: %v\n%s", err, string(body))
	}
	return payload.Data
}

func waitForRefreshJobStatus(t *testing.T, r http.Handler, id string, status refreshJobStatus) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		w := performRequest(r, http.MethodGet, "/api/feeds/refresh-jobs/"+id, nil, nil)
		got := decodeRefreshJobResponse(t, w.Body.Bytes())
		if got.Status == status {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("refresh job %s did not reach status %s", id, status)
}

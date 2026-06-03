package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const refreshEventName = "refresh-completed"

type refreshEventScope string

const (
	refreshEventScopeAll  refreshEventScope = "all"
	refreshEventScopeFeed refreshEventScope = "feed"
)

type refreshEvent struct {
	Type       string            `json:"type"`
	Scope      refreshEventScope `json:"scope"`
	FeedID     *int64            `json:"feed_id"`
	FinishedAt int64             `json:"finished_at"`
}

type refreshEventBroker struct {
	mu          sync.Mutex
	nextID      int64
	subscribers map[int64]chan refreshEvent
}

func newRefreshEventBroker() *refreshEventBroker {
	return &refreshEventBroker{subscribers: make(map[int64]chan refreshEvent)}
}

func (b *refreshEventBroker) subscribe() (int64, <-chan refreshEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextID++
	id := b.nextID
	ch := make(chan refreshEvent, 1)
	b.subscribers[id] = ch
	return id, ch
}

func (b *refreshEventBroker) unsubscribe(id int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.subscribers, id)
}

func (b *refreshEventBroker) publish(event refreshEvent) {
	b.mu.Lock()
	channels := make([]chan refreshEvent, 0, len(b.subscribers))
	for _, ch := range b.subscribers {
		channels = append(channels, ch)
	}
	b.mu.Unlock()

	for _, ch := range channels {
		select {
		case ch <- event:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- event:
			default:
			}
		}
	}
}

func (h *Handler) streamRefreshEvents(c *gin.Context) {
	if h.refreshEvents == nil {
		internalError(c, errors.New("refresh event broker is not initialized"), "refresh events unavailable")
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		internalError(c, errors.New("response writer does not support flushing"), "streaming unsupported")
		return
	}

	id, events := h.refreshEvents.subscribe()
	defer h.refreshEvents.unsubscribe(id)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	if _, err := io.WriteString(c.Writer, ": connected\n\n"); err != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event := <-events:
			if err := writeRefreshSSE(c.Writer, event); err != nil {
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			if _, err := io.WriteString(c.Writer, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeRefreshSSE(w io.Writer, event refreshEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", refreshEventName, payload)
	return err
}

func (h *Handler) publishAllRefreshCompleted() {
	h.publishRefreshCompleted(refreshEventScopeAll, nil)
}

func (h *Handler) publishFeedRefreshCompleted(feedID int64) {
	h.publishRefreshCompleted(refreshEventScopeFeed, &feedID)
}

func (h *Handler) publishRefreshCompleted(scope refreshEventScope, feedID *int64) {
	if h.refreshEvents == nil {
		return
	}

	h.refreshEvents.publish(refreshEvent{
		Type:       "refresh_completed",
		Scope:      scope,
		FeedID:     feedID,
		FinishedAt: time.Now().Unix(),
	})
}

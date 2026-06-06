package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0x2E/fusion/internal/config"
	"github.com/gin-gonic/gin"
)

func TestBatchCreateFeedsValidation(t *testing.T) {
	feeds := make([]gin.H, maxBatchCreateFeeds+1)
	for i := range feeds {
		feeds[i] = gin.H{
			"group_id": 1,
			"name":     "Feed",
			"link":     "https://example.com/feed.xml",
		}
	}

	tests := []struct {
		name string
		body any
	}{
		{
			name: "rejects empty feeds",
			body: gin.H{"feeds": []gin.H{}},
		},
		{
			name: "rejects too many feeds",
			body: gin.H{"feeds": feeds},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{config: &config.Config{}}
			r := newTestRouter()
			r.POST("/api/feeds/batch", h.batchCreateFeeds)

			w := performRequest(
				r,
				http.MethodPost,
				"/api/feeds/batch",
				mustJSONBody(t, tt.body),
				map[string]string{"Content-Type": "application/json"},
			)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", w.Code)
			}
		})
	}
}

func TestValidateFeedRejectsPrivateURLByDefault(t *testing.T) {
	h := &Handler{config: &config.Config{}}
	r := newTestRouter()
	r.POST("/api/feeds/validate", h.validateFeed)

	w := performRequest(
		r,
		http.MethodPost,
		"/api/feeds/validate",
		mustJSONBody(t, gin.H{"url": "http://127.0.0.1/feed.xml"}),
		map[string]string{"Content-Type": "application/json"},
	)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestValidateFeedDiscoversAlternateFeedWithHardenedClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><link rel="alternate" type="application/rss+xml" title="RSS" href="/rss.xml"></head><body></body></html>`))
	}))
	t.Cleanup(server.Close)

	h := &Handler{config: &config.Config{AllowPrivateFeeds: true}}
	r := newTestRouter()
	r.POST("/api/feeds/validate", h.validateFeed)

	w := performRequest(
		r,
		http.MethodPost,
		"/api/feeds/validate",
		mustJSONBody(t, gin.H{"url": server.URL}),
		map[string]string{"Content-Type": "application/json"},
	)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Data validateFeedResponse `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data.Feeds) != 1 {
		t.Fatalf("expected one discovered feed, got %d", len(body.Data.Feeds))
	}
	if got, want := body.Data.Feeds[0].Link, server.URL+"/rss.xml"; got != want {
		t.Fatalf("feed link = %q, want %q", got, want)
	}
}

package handler

import (
	"net/http"
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

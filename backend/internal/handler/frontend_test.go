package handler

import (
	"io/fs"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestSetupRouterServesEmbeddedFrontend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{}
	r := h.SetupRouter()

	t.Run("serves index for root", func(t *testing.T) {
		w := performRequest(r, http.MethodGet, "/", nil, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if contentType := w.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
			t.Fatalf("expected text/html content type, got %q", contentType)
		}
		if csp := w.Header().Get("Content-Security-Policy"); csp == "" {
			t.Fatal("expected Content-Security-Policy header")
		}
	})

	t.Run("serves index for client-side route", func(t *testing.T) {
		w := performRequest(r, http.MethodGet, "/feeds", nil, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if contentType := w.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
			t.Fatalf("expected text/html content type, got %q", contentType)
		}
	})

	t.Run("returns 404 for unknown api path", func(t *testing.T) {
		w := performRequest(r, http.MethodGet, "/api/not-found", nil, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("returns 404 for missing asset", func(t *testing.T) {
		w := performRequest(r, http.MethodGet, "/assets/does-not-exist-0db74db9a5.js", nil, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}

func TestFrontendCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	setupFrontendRoutesWithFS(r, fstest.MapFS{
		"index.html": {
			Data: []byte("<!doctype html><title>Fusion</title>"),
			Mode: fs.ModePerm,
		},
		"assets/index-abc123.js": {
			Data: []byte("console.log('fusion')"),
			Mode: fs.ModePerm,
		},
		"sw.js": {
			Data: []byte("self.addEventListener('fetch', () => {})"),
			Mode: fs.ModePerm,
		},
		"icon-32.png": {
			Data: []byte("png"),
			Mode: fs.ModePerm,
		},
	})

	tests := []struct {
		name         string
		path         string
		cacheControl string
	}{
		{
			name:         "immutable hashed asset",
			path:         "/assets/index-abc123.js",
			cacheControl: "public, max-age=31536000, immutable",
		},
		{
			name:         "spa route html",
			path:         "/unread?article=36166",
			cacheControl: "no-cache",
		},
		{
			name:         "service worker",
			path:         "/sw.js",
			cacheControl: "no-cache",
		},
		{
			name:         "public icon",
			path:         "/icon-32.png",
			cacheControl: "public, max-age=86400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(r, http.MethodGet, tt.path, nil, nil)
			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", w.Code)
			}
			if got := w.Header().Get("Cache-Control"); got != tt.cacheControl {
				t.Fatalf("expected Cache-Control %q, got %q", tt.cacheControl, got)
			}
		})
	}
}

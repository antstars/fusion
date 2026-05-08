package pull

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestRefreshFeedPreservesValidatorsWhen304OmitHeaders(t *testing.T) {
	st, err := newTestStore(t)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer st.Close()

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Header().Set("ETag", `"etag-v1"`)
			w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
			w.Header().Set("Cache-Control", "max-age=86400")
			w.Header().Set("Expires", "Tue, 03 Jan 2006 15:04:05 GMT")
			_, _ = fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Demo</title><link>https://example.com</link>
<item><guid>g1</guid><title>Item</title><link>https://example.com/1</link></item>
</channel></rss>`)
			return
		}

		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	feed, err := st.CreateFeed(1, "Feed A", server.URL, "", "")
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}

	p := New(st, &config.Config{
		PullInterval:      1800,
		PullTimeout:       5,
		PullConcurrency:   1,
		PullMaxBackoff:    604800,
		AllowPrivateFeeds: true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := p.RefreshFeed(ctx, feed.ID); err != nil {
		t.Fatalf("first refresh: %v", err)
	}

	if err := p.RefreshFeed(ctx, feed.ID); err != nil {
		t.Fatalf("second refresh: %v", err)
	}

	updatedFeed, err := st.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("get feed: %v", err)
	}

	if updatedFeed.FetchState.ETag != `"etag-v1"` {
		t.Fatalf("etag = %q, want %q", updatedFeed.FetchState.ETag, `"etag-v1"`)
	}
	if updatedFeed.FetchState.LastModified != "Mon, 02 Jan 2006 15:04:05 GMT" {
		t.Fatalf("last_modified = %q, want %q", updatedFeed.FetchState.LastModified, "Mon, 02 Jan 2006 15:04:05 GMT")
	}
	if updatedFeed.FetchState.CacheControl != "max-age=86400" {
		t.Fatalf("cache_control = %q, want %q", updatedFeed.FetchState.CacheControl, "max-age=86400")
	}

	expires, err := http.ParseTime("Tue, 03 Jan 2006 15:04:05 GMT")
	if err != nil {
		t.Fatalf("parse expires: %v", err)
	}
	if updatedFeed.FetchState.ExpiresAt != expires.Unix() {
		t.Fatalf("expires_at = %d, want %d", updatedFeed.FetchState.ExpiresAt, expires.Unix())
	}
}

func TestRefreshAllWaitsForRunningJobs(t *testing.T) {
	st, err := newTestStore(t)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer st.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Demo</title><link>https://example.com</link>
<item><guid>%s</guid><title>Item</title><link>https://example.com%s</link></item>
</channel></rss>`, r.URL.Path, r.URL.Path)
	}))
	defer server.Close()

	if _, err := st.CreateFeed(1, "Feed A", server.URL+"/a", "", ""); err != nil {
		t.Fatalf("create feed A: %v", err)
	}
	if _, err := st.CreateFeed(1, "Feed B", server.URL+"/b", "", ""); err != nil {
		t.Fatalf("create feed B: %v", err)
	}

	p := New(st, &config.Config{
		PullInterval:      1800,
		PullTimeout:       5,
		PullConcurrency:   1,
		PullMaxBackoff:    604800,
		AllowPrivateFeeds: true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := p.RefreshAll(ctx)
	if err != nil {
		t.Fatalf("refresh all: %v", err)
	}
	if count != 2 {
		t.Fatalf("refresh count = %d, want 2", count)
	}

	feeds, err := st.ListFeeds()
	if err != nil {
		t.Fatalf("list feeds: %v", err)
	}
	for _, feed := range feeds {
		if feed.FetchState.LastSuccessAt <= 0 {
			t.Fatalf("feed %d last_success_at = %d, want > 0", feed.ID, feed.FetchState.LastSuccessAt)
		}
	}
}

func newTestStore(t *testing.T) (*store.Store, error) {
	t.Helper()

	databaseURL := os.Getenv("FUSION_TEST_POSTGRES_URL")
	if databaseURL == "" {
		t.Skip("FUSION_TEST_POSTGRES_URL is not set")
	}
	st, err := store.New(databaseURL)
	if err != nil {
		return nil, err
	}
	resetTestDB(t, databaseURL)
	return st, nil
}

func resetTestDB(t *testing.T, databaseURL string) {
	t.Helper()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open postgres test database: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		TRUNCATE TABLE bookmarks, items, feed_fetch_state, feeds, groups RESTART IDENTITY CASCADE;
		INSERT INTO groups (id, name) VALUES (1, 'Default');
		SELECT setval(pg_get_serial_sequence('groups', 'id'), 1);
	`); err != nil {
		t.Fatalf("reset postgres test database: %v", err)
	}
}

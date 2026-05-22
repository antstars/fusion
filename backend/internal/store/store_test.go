package store

import (
	"os"
	"testing"

	"github.com/0x2E/fusion/internal/model"
)

func closeStore(t *testing.T, store *Store) {
	t.Helper()
	if err := store.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func setupTestDB(t *testing.T) (*Store, string) {
	t.Helper()

	databaseURL := os.Getenv("FUSION_TEST_POSTGRES_URL")
	if databaseURL == "" {
		t.Skip("FUSION_TEST_POSTGRES_URL is not set")
	}

	store, err := New(databaseURL)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	resetTestDB(t, store)

	return store, databaseURL
}

func resetTestDB(t *testing.T, store *Store) {
	t.Helper()

	if _, err := store.db.Exec(`
		TRUNCATE TABLE read_later_items, bookmarks, items, feed_fetch_state, feeds, groups, app_settings RESTART IDENTITY CASCADE;
		INSERT INTO groups (id, name) VALUES (1, 'Default');
		SELECT setval(pg_get_serial_sequence('groups', 'id'), 1);
		INSERT INTO app_settings (key, value) VALUES ('max_articles', '0'), ('retention_days', '30');
	`); err != nil {
		_ = store.Close()
		t.Fatalf("reset test database: %v", err)
	}
}

func mustCreateGroup(t *testing.T, store *Store, name string) *model.Group {
	t.Helper()

	group, err := store.CreateGroup(name)
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	return group
}

func mustCreateFeed(t *testing.T, store *Store, groupID int64, name, link, siteURL, proxy string) *model.Feed {
	t.Helper()

	feed, err := store.CreateFeed(groupID, name, link, siteURL, proxy)
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	return feed
}

func mustCreateItem(t *testing.T, store *Store, feedID int64, guid, title, link, content string, pubDate int64) *model.Item {
	t.Helper()

	item, err := store.CreateItem(feedID, guid, title, link, content, pubDate)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	return item
}

func mustCreateBookmark(t *testing.T, store *Store, itemID *int64, link, title, content string, pubDate int64, feedName string) *model.Bookmark {
	t.Helper()

	bookmark, err := store.CreateBookmark(itemID, link, title, content, pubDate, feedName)
	if err != nil {
		t.Fatalf("CreateBookmark() failed: %v", err)
	}

	return bookmark
}

func mustCreateReadLaterItem(t *testing.T, store *Store, itemID *int64, link, title, content string, pubDate int64, feedName string) *model.ReadLaterItem {
	t.Helper()

	item, err := store.CreateReadLaterItem(itemID, link, title, content, pubDate, feedName)
	if err != nil {
		t.Fatalf("CreateReadLaterItem() failed: %v", err)
	}

	return item
}

func TestNew(t *testing.T) {
	databaseURL := os.Getenv("FUSION_TEST_POSTGRES_URL")
	if databaseURL == "" {
		t.Skip("FUSION_TEST_POSTGRES_URL is not set")
	}

	store, err := New(databaseURL)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer closeStore(t, store)

	// Verify database connection is alive
	if err := store.db.Ping(); err != nil {
		t.Errorf("database ping failed: %v", err)
	}
}

func TestClose(t *testing.T) {
	store, _ := setupTestDB(t)

	if err := store.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Verify connection is closed by attempting to ping
	if err := store.db.Ping(); err == nil {
		t.Error("expected ping to fail after Close(), but it succeeded")
	}
}

func TestNewInvalidPath(t *testing.T) {
	_, err := New("postgres://invalid:invalid@127.0.0.1:1/invalid?sslmode=disable")
	if err == nil {
		t.Error("expected New() to fail with invalid database URL, but it succeeded")
	}
}

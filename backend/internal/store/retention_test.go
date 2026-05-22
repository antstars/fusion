package store

import (
	"database/sql"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/model"
)

func mustSetItemTimes(t *testing.T, store *Store, itemID, pubDate, createdAt int64) {
	t.Helper()

	if _, err := store.db.Exec(
		`UPDATE items SET pub_date = :pub_date, created_at = :created_at WHERE id = :id`,
		sql.Named("pub_date", pubDate),
		sql.Named("created_at", createdAt),
		sql.Named("id", itemID),
	); err != nil {
		t.Fatalf("set item times: %v", err)
	}
}

func TestRetentionSettingsDefaultsAndUpdate(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	settings, err := store.GetRetentionSettings()
	if err != nil {
		t.Fatalf("GetRetentionSettings() failed: %v", err)
	}
	if settings.MaxArticles != 0 || settings.RetentionDays != 30 {
		t.Fatalf("unexpected defaults: %+v", settings)
	}

	updated, err := store.UpdateRetentionSettings(model.RetentionSettings{MaxArticles: 25, RetentionDays: 90})
	if err != nil {
		t.Fatalf("UpdateRetentionSettings() failed: %v", err)
	}
	if updated.MaxArticles != 25 || updated.RetentionDays != 90 {
		t.Fatalf("unexpected updated settings: %+v", updated)
	}
}

func TestCleanupItemsByRetentionAge(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Tech")
	feed := mustCreateFeed(t, store, group.ID, "Feed", "https://example.com/rss.xml", "", "")
	now := time.Unix(1_700_000_000, 0)

	old := mustCreateItem(t, store, feed.ID, "old", "Old", "https://example.com/old", "", now.AddDate(0, 0, -31).Unix())
	recent := mustCreateItem(t, store, feed.ID, "recent", "Recent", "https://example.com/recent", "", now.AddDate(0, 0, -10).Unix())
	mustSetItemTimes(t, store, old.ID, old.PubDate, old.PubDate)
	mustSetItemTimes(t, store, recent.ID, recent.PubDate, recent.PubDate)

	deleted, err := store.CleanupItemsByRetention(model.RetentionSettings{RetentionDays: 30}, now)
	if err != nil {
		t.Fatalf("CleanupItemsByRetention() failed: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}

	if _, err := store.GetItem(old.ID); err == nil {
		t.Fatal("expected old item to be deleted")
	}
	if _, err := store.GetItem(recent.ID); err != nil {
		t.Fatalf("expected recent item to remain: %v", err)
	}
}

func TestCleanupItemsByRetentionMaxArticles(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Tech")
	feed := mustCreateFeed(t, store, group.ID, "Feed", "https://example.com/rss.xml", "", "")
	now := time.Unix(1_700_000_000, 0)

	oldest := mustCreateItem(t, store, feed.ID, "oldest", "Oldest", "https://example.com/oldest", "", now.Add(-3*time.Hour).Unix())
	middle := mustCreateItem(t, store, feed.ID, "middle", "Middle", "https://example.com/middle", "", now.Add(-2*time.Hour).Unix())
	newest := mustCreateItem(t, store, feed.ID, "newest", "Newest", "https://example.com/newest", "", now.Add(-time.Hour).Unix())
	for _, item := range []*model.Item{oldest, middle, newest} {
		mustSetItemTimes(t, store, item.ID, item.PubDate, item.PubDate)
	}

	deleted, err := store.CleanupItemsByRetention(model.RetentionSettings{MaxArticles: 2}, now)
	if err != nil {
		t.Fatalf("CleanupItemsByRetention() failed: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}

	if _, err := store.GetItem(oldest.ID); err == nil {
		t.Fatal("expected oldest item to be deleted")
	}
	if _, err := store.GetItem(middle.ID); err != nil {
		t.Fatalf("expected middle item to remain: %v", err)
	}
	if _, err := store.GetItem(newest.ID); err != nil {
		t.Fatalf("expected newest item to remain: %v", err)
	}
}

func TestCleanupItemsByRetentionCombinedAndSavedSnapshots(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Tech")
	feed := mustCreateFeed(t, store, group.ID, "Feed", "https://example.com/rss.xml", "", "")
	now := time.Unix(1_700_000_000, 0)

	old := mustCreateItem(t, store, feed.ID, "old", "Old", "https://example.com/old", "old content", now.AddDate(0, 0, -31).Unix())
	kept := mustCreateItem(t, store, feed.ID, "kept", "Kept", "https://example.com/kept", "", now.AddDate(0, 0, -1).Unix())
	mustSetItemTimes(t, store, old.ID, old.PubDate, old.PubDate)
	mustSetItemTimes(t, store, kept.ID, kept.PubDate, kept.PubDate)

	bookmark := mustCreateBookmark(t, store, &old.ID, old.Link, old.Title, old.Content, old.PubDate, feed.Name)
	readLater := mustCreateReadLaterItem(t, store, &old.ID, old.Link, old.Title, old.Content, old.PubDate, feed.Name)

	deleted, err := store.CleanupItemsByRetention(model.RetentionSettings{MaxArticles: 1, RetentionDays: 30}, now)
	if err != nil {
		t.Fatalf("CleanupItemsByRetention() failed: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}

	updatedBookmark, err := store.GetBookmark(bookmark.ID)
	if err != nil {
		t.Fatalf("GetBookmark() failed: %v", err)
	}
	if updatedBookmark.ItemID != nil || updatedBookmark.Content != old.Content {
		t.Fatalf("expected bookmark snapshot to remain with nil item_id, got %+v", updatedBookmark)
	}

	updatedReadLater, err := store.GetReadLaterItem(readLater.ID)
	if err != nil {
		t.Fatalf("GetReadLaterItem() failed: %v", err)
	}
	if updatedReadLater.ItemID != nil || updatedReadLater.Content != old.Content {
		t.Fatalf("expected read-later snapshot to remain with nil item_id, got %+v", updatedReadLater)
	}
}

func TestCleanupItemsByRetentionUnlimited(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Tech")
	feed := mustCreateFeed(t, store, group.ID, "Feed", "https://example.com/rss.xml", "", "")
	mustCreateItem(t, store, feed.ID, "item", "Item", "https://example.com/item", "", 1)

	deleted, err := store.CleanupItemsByRetention(model.RetentionSettings{}, time.Unix(1_700_000_000, 0))
	if err != nil {
		t.Fatalf("CleanupItemsByRetention() failed: %v", err)
	}
	if deleted != 0 {
		t.Fatalf("deleted = %d, want 0", deleted)
	}

	total, err := store.CountItems(ListItemsParams{})
	if err != nil {
		t.Fatalf("CountItems() failed: %v", err)
	}
	if total != 1 {
		t.Fatalf("total = %d, want 1", total)
	}
}

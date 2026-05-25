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

func TestCleanupItemsByRetentionSkipsSavedItems(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Tech")
	feed := mustCreateFeed(t, store, group.ID, "Feed", "https://example.com/rss.xml", "", "")
	now := time.Unix(1_700_000_000, 0)

	oldBookmarked := mustCreateItem(t, store, feed.ID, "old-bookmarked", "Old Bookmarked", "https://example.com/old-bookmarked", "bookmarked content", now.AddDate(0, 0, -33).Unix())
	oldReadLater := mustCreateItem(t, store, feed.ID, "old-read-later", "Old Read Later", "https://example.com/old-read-later", "read later content", now.AddDate(0, 0, -32).Unix())
	oldUnsaved := mustCreateItem(t, store, feed.ID, "old-unsaved", "Old Unsaved", "https://example.com/old-unsaved", "", now.AddDate(0, 0, -31).Unix())
	recent := mustCreateItem(t, store, feed.ID, "recent", "Recent", "https://example.com/recent", "", now.AddDate(0, 0, -1).Unix())
	for _, item := range []*model.Item{oldBookmarked, oldReadLater, oldUnsaved, recent} {
		mustSetItemTimes(t, store, item.ID, item.PubDate, item.PubDate)
	}

	bookmark := mustCreateBookmark(t, store, &oldBookmarked.ID, oldBookmarked.Link, oldBookmarked.Title, oldBookmarked.Content, oldBookmarked.PubDate, feed.Name)
	readLater := mustCreateReadLaterItem(t, store, &oldReadLater.ID, oldReadLater.Link, oldReadLater.Title, oldReadLater.Content, oldReadLater.PubDate, feed.Name)

	deleted, err := store.CleanupItemsByRetention(model.RetentionSettings{MaxArticles: 1, RetentionDays: 30}, now)
	if err != nil {
		t.Fatalf("CleanupItemsByRetention() failed: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}

	if _, err := store.GetItem(oldBookmarked.ID); err != nil {
		t.Fatalf("expected bookmarked item to remain: %v", err)
	}
	if _, err := store.GetItem(oldReadLater.ID); err != nil {
		t.Fatalf("expected read-later item to remain: %v", err)
	}
	if _, err := store.GetItem(oldUnsaved.ID); err == nil {
		t.Fatal("expected old unsaved item to be deleted")
	}
	if _, err := store.GetItem(recent.ID); err != nil {
		t.Fatalf("expected recent item to remain: %v", err)
	}

	updatedBookmark, err := store.GetBookmark(bookmark.ID)
	if err != nil {
		t.Fatalf("GetBookmark() failed: %v", err)
	}
	if updatedBookmark.ItemID == nil || *updatedBookmark.ItemID != oldBookmarked.ID {
		t.Fatalf("expected bookmark item_id to remain linked to %d, got %+v", oldBookmarked.ID, updatedBookmark)
	}

	updatedReadLater, err := store.GetReadLaterItem(readLater.ID)
	if err != nil {
		t.Fatalf("GetReadLaterItem() failed: %v", err)
	}
	if updatedReadLater.ItemID == nil || *updatedReadLater.ItemID != oldReadLater.ID {
		t.Fatalf("expected read-later item_id to remain linked to %d, got %+v", oldReadLater.ID, updatedReadLater)
	}
}

func TestCleanupItemsByRetentionMaxArticlesSkipsSavedItems(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Tech")
	feed := mustCreateFeed(t, store, group.ID, "Feed", "https://example.com/rss.xml", "", "")
	now := time.Unix(1_700_000_000, 0)

	oldestProtected := mustCreateItem(t, store, feed.ID, "oldest-protected", "Oldest Protected", "https://example.com/oldest-protected", "", now.Add(-5*time.Hour).Unix())
	oldestRegular := mustCreateItem(t, store, feed.ID, "oldest-regular", "Oldest Regular", "https://example.com/oldest-regular", "", now.Add(-4*time.Hour).Unix())
	middleProtected := mustCreateItem(t, store, feed.ID, "middle-protected", "Middle Protected", "https://example.com/middle-protected", "", now.Add(-3*time.Hour).Unix())
	middleRegular := mustCreateItem(t, store, feed.ID, "middle-regular", "Middle Regular", "https://example.com/middle-regular", "", now.Add(-2*time.Hour).Unix())
	newestRegular := mustCreateItem(t, store, feed.ID, "newest-regular", "Newest Regular", "https://example.com/newest-regular", "", now.Add(-time.Hour).Unix())
	for _, item := range []*model.Item{oldestProtected, oldestRegular, middleProtected, middleRegular, newestRegular} {
		mustSetItemTimes(t, store, item.ID, item.PubDate, item.PubDate)
	}

	mustCreateBookmark(t, store, &oldestProtected.ID, oldestProtected.Link, oldestProtected.Title, oldestProtected.Content, oldestProtected.PubDate, feed.Name)
	mustCreateReadLaterItem(t, store, &middleProtected.ID, middleProtected.Link, middleProtected.Title, middleProtected.Content, middleProtected.PubDate, feed.Name)

	deleted, err := store.CleanupItemsByRetention(model.RetentionSettings{MaxArticles: 2}, now)
	if err != nil {
		t.Fatalf("CleanupItemsByRetention() failed: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}

	if _, err := store.GetItem(oldestProtected.ID); err != nil {
		t.Fatalf("expected oldest protected item to remain: %v", err)
	}
	if _, err := store.GetItem(oldestRegular.ID); err == nil {
		t.Fatal("expected oldest regular item to be deleted")
	}
	if _, err := store.GetItem(middleProtected.ID); err != nil {
		t.Fatalf("expected middle protected item to remain: %v", err)
	}
	if _, err := store.GetItem(middleRegular.ID); err != nil {
		t.Fatalf("expected middle regular item to remain: %v", err)
	}
	if _, err := store.GetItem(newestRegular.ID); err != nil {
		t.Fatalf("expected newest regular item to remain: %v", err)
	}
}

func TestCleanupItemsByRetentionMaxArticlesDeletesNothingWhenOverflowIsSaved(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Tech")
	feed := mustCreateFeed(t, store, group.ID, "Feed", "https://example.com/rss.xml", "", "")
	now := time.Unix(1_700_000_000, 0)

	oldBookmarked := mustCreateItem(t, store, feed.ID, "old-bookmarked", "Old Bookmarked", "https://example.com/old-bookmarked", "", now.Add(-3*time.Hour).Unix())
	oldReadLater := mustCreateItem(t, store, feed.ID, "old-read-later", "Old Read Later", "https://example.com/old-read-later", "", now.Add(-2*time.Hour).Unix())
	newestRegular := mustCreateItem(t, store, feed.ID, "newest-regular", "Newest Regular", "https://example.com/newest-regular", "", now.Add(-time.Hour).Unix())
	for _, item := range []*model.Item{oldBookmarked, oldReadLater, newestRegular} {
		mustSetItemTimes(t, store, item.ID, item.PubDate, item.PubDate)
	}

	mustCreateBookmark(t, store, &oldBookmarked.ID, oldBookmarked.Link, oldBookmarked.Title, oldBookmarked.Content, oldBookmarked.PubDate, feed.Name)
	mustCreateReadLaterItem(t, store, &oldReadLater.ID, oldReadLater.Link, oldReadLater.Title, oldReadLater.Content, oldReadLater.PubDate, feed.Name)

	deleted, err := store.CleanupItemsByRetention(model.RetentionSettings{MaxArticles: 1}, now)
	if err != nil {
		t.Fatalf("CleanupItemsByRetention() failed: %v", err)
	}
	if deleted != 0 {
		t.Fatalf("deleted = %d, want 0", deleted)
	}

	for _, item := range []*model.Item{oldBookmarked, oldReadLater, newestRegular} {
		if _, err := store.GetItem(item.ID); err != nil {
			t.Fatalf("expected item %d to remain: %v", item.ID, err)
		}
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

package store

import (
	"database/sql"
	"errors"
	"testing"
)

func TestListReadLaterItems(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	items, err := store.ListReadLaterItems(10, 0)
	if err != nil {
		t.Fatalf("ListReadLaterItems() failed: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 read later items, got %d", len(items))
	}

	pubDate := int64(123)
	i1 := mustCreateReadLaterItem(t, store, nil, "https://example.com/1", "Item 1", "Content 1", pubDate, "Feed 1")
	i2 := mustCreateReadLaterItem(t, store, nil, "https://example.com/2", "Item 2", "Content 2", pubDate, "Feed 2")
	i3 := mustCreateReadLaterItem(t, store, nil, "https://example.com/3", "Item 3", "Content 3", pubDate, "Feed 3")

	for _, tc := range []struct {
		id        int64
		createdAt int64
	}{{i1.ID, 100}, {i2.ID, 200}, {i3.ID, 300}} {
		if _, err := store.db.Exec(
			`UPDATE read_later_items SET created_at = :created_at WHERE id = :id`,
			sql.Named("created_at", tc.createdAt),
			sql.Named("id", tc.id),
		); err != nil {
			t.Fatalf("failed to set created_at: %v", err)
		}
	}

	t.Run("list all items ordered by created_at DESC", func(t *testing.T) {
		items, err := store.ListReadLaterItems(10, 0)
		if err != nil {
			t.Fatalf("ListReadLaterItems() failed: %v", err)
		}
		if len(items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(items))
		}
		if items[0].ID != i3.ID || items[1].ID != i2.ID || items[2].ID != i1.ID {
			t.Error("read later items not ordered by created_at DESC")
		}
	})

	t.Run("pagination with limit", func(t *testing.T) {
		items, err := store.ListReadLaterItems(2, 0)
		if err != nil {
			t.Fatalf("ListReadLaterItems() failed: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items with limit=2, got %d", len(items))
		}
	})

	t.Run("pagination with offset", func(t *testing.T) {
		items, err := store.ListReadLaterItems(10, 2)
		if err != nil {
			t.Fatalf("ListReadLaterItems() failed: %v", err)
		}
		if len(items) != 1 {
			t.Errorf("expected 1 item with offset=2, got %d", len(items))
		}
		if items[0].ID != i1.ID {
			t.Error("incorrect read later item returned with offset")
		}
	})
}

func TestListReadLaterItemsIncludesUnreadState(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Unread Read Later Group")
	feed := mustCreateFeed(t, store, group.ID, "Unread Read Later Feed", "https://example.com/feed.xml", "https://example.com", "")
	unreadItem := mustCreateItem(t, store, feed.ID, "read-later-unread", "Unread Item", "https://example.com/unread", "Content", 123)
	readItem := mustCreateItem(t, store, feed.ID, "read-later-read", "Read Item", "https://example.com/read", "Content", 124)
	if err := store.UpdateItemUnread(readItem.ID, false); err != nil {
		t.Fatalf("UpdateItemUnread() failed: %v", err)
	}

	mustCreateReadLaterItem(t, store, &unreadItem.ID, unreadItem.Link, unreadItem.Title, unreadItem.Content, unreadItem.PubDate, feed.Name)
	mustCreateReadLaterItem(t, store, &readItem.ID, readItem.Link, readItem.Title, readItem.Content, readItem.PubDate, feed.Name)
	mustCreateReadLaterItem(t, store, nil, "https://example.com/orphan", "Orphan", "Content", 125, feed.Name)

	items, err := store.ListReadLaterItems(10, 0)
	if err != nil {
		t.Fatalf("ListReadLaterItems() failed: %v", err)
	}

	unreadByTitle := map[string]bool{}
	for _, item := range items {
		unreadByTitle[item.Title] = item.Unread
	}

	if !unreadByTitle["Unread Item"] {
		t.Fatal("expected linked unread read-later item to report unread=true")
	}
	if unreadByTitle["Read Item"] {
		t.Fatal("expected linked read read-later item to report unread=false")
	}
	if unreadByTitle["Orphan"] {
		t.Fatal("expected orphan read-later item to report unread=false")
	}
}

func TestGetReadLaterItem(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	created := mustCreateReadLaterItem(t, store, nil, "https://example.com/test", "Test Item", "Content", 123, "Test Feed")

	item, err := store.GetReadLaterItem(created.ID)
	if err != nil {
		t.Fatalf("GetReadLaterItem() failed: %v", err)
	}
	if item.ID != created.ID || item.Title != created.Title {
		t.Error("retrieved read later item doesn't match created item")
	}

	_, err = store.GetReadLaterItem(99999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for non-existent read later item, got %v", err)
	}
}

func TestCreateReadLaterItem(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	t.Run("create with item_id", func(t *testing.T) {
		group := mustCreateGroup(t, store, "Test Group")
		feed := mustCreateFeed(t, store, group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
		item := mustCreateItem(t, store, feed.ID, "guid-1", "Test Item", "https://example.com/item", "Content", 123)

		readLater := mustCreateReadLaterItem(t, store, &item.ID, item.Link, item.Title, item.Content, item.PubDate, "Test Feed")

		if readLater.ItemID == nil || *readLater.ItemID != item.ID {
			t.Error("expected item_id to be set")
		}
		if readLater.Link != item.Link || readLater.Title != item.Title {
			t.Error("read later fields don't match input")
		}
		if readLater.ID == 0 || readLater.CreatedAt == 0 {
			t.Error("expected auto-populated fields to be set")
		}
	})

	t.Run("unique constraint on link", func(t *testing.T) {
		link := "https://example.com/duplicate"
		mustCreateReadLaterItem(t, store, nil, link, "Item 1", "Content", 123, "Feed")

		_, err := store.CreateReadLaterItem(nil, link, "Item 2", "Content", 123, "Feed")
		if err == nil {
			t.Error("expected error when creating duplicate read later link, got nil")
		}
	})
}

func TestDeleteReadLaterItem(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	item := mustCreateReadLaterItem(t, store, nil, "https://example.com/test", "Test Item", "Content", 123, "Test Feed")

	if err := store.DeleteReadLaterItem(item.ID); err != nil {
		t.Fatalf("DeleteReadLaterItem() failed: %v", err)
	}

	_, err := store.GetReadLaterItem(item.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestDeleteFeedPreservesReadLaterSnapshot(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Test Group")
	feed := mustCreateFeed(t, store, group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
	item := mustCreateItem(t, store, feed.ID, "guid-1", "Test Item", "https://example.com/item", "Content", 123)
	readLater := mustCreateReadLaterItem(t, store, &item.ID, item.Link, item.Title, item.Content, item.PubDate, feed.Name)

	if err := store.DeleteFeed(feed.ID); err != nil {
		t.Fatalf("DeleteFeed() failed: %v", err)
	}

	saved, err := store.GetReadLaterItem(readLater.ID)
	if err != nil {
		t.Fatalf("GetReadLaterItem() failed: %v", err)
	}
	if saved.ItemID != nil {
		t.Fatal("expected read later item_id to be NULL after feed deletion")
	}
	if saved.Link != item.Link || saved.Title != item.Title || saved.Content != item.Content {
		t.Fatal("expected read later snapshot fields to be preserved")
	}
}

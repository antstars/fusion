package store

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestPostgresStoreSmoke(t *testing.T) {
	databaseURL := os.Getenv("FUSION_TEST_POSTGRES_URL")
	if databaseURL == "" {
		t.Skip("FUSION_TEST_POSTGRES_URL is not set")
	}

	st, err := New(databaseURL)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer st.Close()

	suffix := time.Now().UnixNano()
	group, err := st.CreateGroup(fmt.Sprintf("pg-smoke-%d", suffix))
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	feed, err := st.CreateFeed(group.ID, "Postgres Smoke", fmt.Sprintf("https://example.com/%d.xml", suffix), "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	if _, err := st.CreateItem(feed.ID, fmt.Sprintf("guid-%d", suffix), "Postgres Item", "https://example.com/item", "content", suffix); err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	items, err := st.ListItems(ListItemsParams{FeedID: &feed.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListItems() failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

package store

import "testing"

func TestMigrate(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	tables := []string{"groups", "feeds", "feed_fetch_state", "items", "bookmarks", "app_settings", "schema_migrations"}
	for _, table := range tables {
		var exists bool
		err := store.db.QueryRow(`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)
		`, table).Scan(&exists)
		if err != nil {
			t.Errorf("failed to check table %s: %v", table, err)
			continue
		}
		if !exists {
			t.Errorf("expected table %s to exist", table)
		}
	}
}

func TestMigrateIdempotent(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	var initialCount int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&initialCount); err != nil {
		t.Fatalf("failed to query initial migration count: %v", err)
	}

	if err := store.migrate(); err != nil {
		t.Fatalf("second migrate() call failed: %v", err)
	}

	var finalCount int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&finalCount); err != nil {
		t.Fatalf("failed to query final migration count: %v", err)
	}
	if finalCount != initialCount {
		t.Errorf("migrations were re-applied: initial=%d, final=%d", initialCount, finalCount)
	}
}

package store

import (
	"fmt"
	"log/slog"
	"time"
)

func (s *Store) migrate() error {
	startedAt := time.Now()
	slog.Info("postgres database migration started")

	schema := `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	applied_at BIGINT NOT NULL DEFAULT (CAST(EXTRACT(EPOCH FROM NOW()) AS BIGINT))
);

CREATE TABLE IF NOT EXISTS groups (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	created_at BIGINT NOT NULL DEFAULT (CAST(EXTRACT(EPOCH FROM NOW()) AS BIGINT)),
	updated_at BIGINT NOT NULL DEFAULT (CAST(EXTRACT(EPOCH FROM NOW()) AS BIGINT))
);
INSERT INTO groups (id, name)
VALUES (1, 'Default')
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('groups', 'id'), GREATEST((SELECT MAX(id) FROM groups), 1));

CREATE TABLE IF NOT EXISTS feeds (
	id BIGSERIAL PRIMARY KEY,
	group_id BIGINT NOT NULL DEFAULT 1 REFERENCES groups(id) ON UPDATE CASCADE ON DELETE RESTRICT,
	name TEXT NOT NULL,
	link TEXT NOT NULL UNIQUE,
	site_url TEXT DEFAULT '',
	suspended INTEGER DEFAULT 0,
	proxy TEXT DEFAULT '',
	created_at BIGINT NOT NULL DEFAULT (CAST(EXTRACT(EPOCH FROM NOW()) AS BIGINT)),
	updated_at BIGINT NOT NULL DEFAULT (CAST(EXTRACT(EPOCH FROM NOW()) AS BIGINT))
);
CREATE INDEX IF NOT EXISTS idx_feeds_group_id ON feeds(group_id);

CREATE TABLE IF NOT EXISTS items (
	id BIGSERIAL PRIMARY KEY,
	feed_id BIGINT NOT NULL REFERENCES feeds(id) ON UPDATE CASCADE ON DELETE CASCADE,
	guid TEXT NOT NULL,
	title TEXT DEFAULT '',
	link TEXT DEFAULT '',
	content TEXT DEFAULT '',
	pub_date BIGINT DEFAULT 0,
	unread INTEGER DEFAULT 1,
	created_at BIGINT NOT NULL DEFAULT (CAST(EXTRACT(EPOCH FROM NOW()) AS BIGINT))
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_items_feed_guid ON items(feed_id, guid);
CREATE INDEX IF NOT EXISTS idx_items_unread ON items(unread) WHERE unread = 1;
CREATE INDEX IF NOT EXISTS idx_items_pub_date ON items(pub_date DESC);
CREATE INDEX IF NOT EXISTS idx_items_feed_unread ON items(feed_id, unread);

CREATE TABLE IF NOT EXISTS bookmarks (
	id BIGSERIAL PRIMARY KEY,
	item_id BIGINT REFERENCES items(id) ON UPDATE CASCADE ON DELETE SET NULL,
	link TEXT NOT NULL UNIQUE,
	title TEXT DEFAULT '',
	content TEXT DEFAULT '',
	pub_date BIGINT DEFAULT 0,
	feed_name TEXT DEFAULT '',
	created_at BIGINT NOT NULL DEFAULT (CAST(EXTRACT(EPOCH FROM NOW()) AS BIGINT))
);
CREATE INDEX IF NOT EXISTS idx_bookmarks_created_at ON bookmarks(created_at DESC);

CREATE TABLE IF NOT EXISTS feed_fetch_state (
	feed_id BIGINT PRIMARY KEY REFERENCES feeds(id) ON UPDATE CASCADE ON DELETE CASCADE,
	etag TEXT NOT NULL DEFAULT '',
	last_modified TEXT NOT NULL DEFAULT '',
	cache_control TEXT NOT NULL DEFAULT '',
	expires_at BIGINT NOT NULL DEFAULT 0,
	last_checked_at BIGINT NOT NULL DEFAULT 0,
	next_check_at BIGINT NOT NULL DEFAULT 0,
	last_http_status INTEGER NOT NULL DEFAULT 0,
	retry_after_until BIGINT NOT NULL DEFAULT 0,
	last_success_at BIGINT NOT NULL DEFAULT 0,
	last_error_at BIGINT NOT NULL DEFAULT 0,
	last_error TEXT NOT NULL DEFAULT '',
	consecutive_failures BIGINT NOT NULL DEFAULT 0,
	updated_at BIGINT NOT NULL DEFAULT (CAST(EXTRACT(EPOCH FROM NOW()) AS BIGINT))
);
CREATE INDEX IF NOT EXISTS idx_feed_fetch_state_next_check_at ON feed_fetch_state(next_check_at);

INSERT INTO schema_migrations (version) VALUES (1), (2)
ON CONFLICT (version) DO NOTHING;
`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("initialize postgres schema: %w", err)
	}

	slog.Info("postgres database migration finished", "duration", time.Since(startedAt))
	return nil
}

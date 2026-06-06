package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/0x2E/fusion/internal/model"
)

func (s *Store) ListReadLaterItems(limit, offset int) ([]*model.ReadLaterItem, error) {
	query := `
		SELECT read_later_items.id, read_later_items.item_id, read_later_items.link,
		       read_later_items.title, read_later_items.content, read_later_items.pub_date,
		       read_later_items.feed_name, read_later_items.created_at,
		       COALESCE(items.unread, 0) AS unread
		FROM read_later_items
		LEFT JOIN items ON items.id = read_later_items.item_id
		ORDER BY read_later_items.created_at DESC, read_later_items.id DESC
	`
	args := []any{}

	if limit > 0 {
		query += ` LIMIT :limit`
		args = append(args, sql.Named("limit", limit))
	}
	if offset > 0 {
		query += ` OFFSET :offset`
		args = append(args, sql.Named("offset", offset))
	}

	rows, err := s.query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*model.ReadLaterItem{}
	for rows.Next() {
		item := &model.ReadLaterItem{}
		var unread int
		if err := rows.Scan(&item.ID, &item.ItemID, &item.Link, &item.Title, &item.Content, &item.PubDate, &item.FeedName, &item.CreatedAt, &unread); err != nil {
			return nil, err
		}
		item.Unread = intToBool(unread)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetReadLaterItem(id int64) (*model.ReadLaterItem, error) {
	item := &model.ReadLaterItem{}
	var unread int
	err := s.queryRow(`
		SELECT read_later_items.id, read_later_items.item_id, read_later_items.link,
		       read_later_items.title, read_later_items.content, read_later_items.pub_date,
		       read_later_items.feed_name, read_later_items.created_at,
		       COALESCE(items.unread, 0) AS unread
		FROM read_later_items
		LEFT JOIN items ON items.id = read_later_items.item_id
		WHERE read_later_items.id = :id
	`, sql.Named("id", id)).Scan(&item.ID, &item.ItemID, &item.Link, &item.Title, &item.Content, &item.PubDate, &item.FeedName, &item.CreatedAt, &unread)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: read later item", ErrNotFound)
		}
		return nil, fmt.Errorf("get read later item: %w", err)
	}
	item.Unread = intToBool(unread)
	return item, nil
}

// CreateReadLaterItem saves a snapshot so the item can remain available even
// if its source feed is later deleted.
func (s *Store) CreateReadLaterItem(itemID *int64, link, title, content string, pubDate int64, feedName string) (*model.ReadLaterItem, error) {
	id, err := s.insertAndReturnID(s.db, `
		INSERT INTO read_later_items (item_id, link, title, content, pub_date, feed_name)
		VALUES (:item_id, :link, :title, :content, :pub_date, :feed_name)
	`, sql.Named("item_id", itemID), sql.Named("link", link), sql.Named("title", title),
		sql.Named("content", content), sql.Named("pub_date", pubDate), sql.Named("feed_name", feedName))
	if err != nil {
		return nil, err
	}

	return s.GetReadLaterItem(id)
}

func (s *Store) DeleteReadLaterItem(id int64) error {
	result, err := s.exec(`DELETE FROM read_later_items WHERE id = :id`, sql.Named("id", id))
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: read later item", ErrNotFound)
	}
	return nil
}

func (s *Store) DeleteReadLaterItemByItemID(itemID int64) error {
	result, err := s.exec(`DELETE FROM read_later_items WHERE item_id = :item_id`, sql.Named("item_id", itemID))
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: read later item", ErrNotFound)
	}
	return nil
}

func (s *Store) CountReadLaterItems() (int, error) {
	var count int
	err := s.queryRow(`SELECT COUNT(*) FROM read_later_items`).Scan(&count)
	return count, err
}

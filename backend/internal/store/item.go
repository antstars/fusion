package store

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/0x2E/fusion/internal/model"
)

// ListItemsParams specifies filtering and pagination for item queries.
//
// Pointer fields (FeedID, GroupID, Unread) are optional filters - nil means "no filter".
// OrderBy accepts "pub_date" (default) or "created_at".
// Limit/Offset = 0 means no limit/offset.
type ListItemsParams struct {
	FeedID  *int64
	GroupID *int64
	Unread  *bool
	Limit   int
	Offset  int
	OrderBy string // "pub_date" or "created_at"
}

func (s *Store) ListItems(params ListItemsParams) ([]*model.Item, error) {
	query := `
		SELECT items.id, items.feed_id, items.guid, items.title, items.link, items.content,
		       items.pub_date, items.unread, items.created_at,
		       EXISTS(SELECT 1 FROM bookmarks WHERE bookmarks.item_id = items.id) AS bookmarked,
		       EXISTS(SELECT 1 FROM read_later_items WHERE read_later_items.item_id = items.id) AS read_later
		FROM items
	`
	args := []any{}

	// Join feeds table if filtering by GroupID
	if params.GroupID != nil {
		query += ` INNER JOIN feeds ON items.feed_id = feeds.id`
	}

	query += ` WHERE 1=1`

	if params.FeedID != nil {
		query += ` AND items.feed_id = :feed_id`
		args = append(args, sql.Named("feed_id", *params.FeedID))
	}
	if params.GroupID != nil {
		query += ` AND feeds.group_id = :group_id`
		args = append(args, sql.Named("group_id", *params.GroupID))
	}
	if params.Unread != nil {
		query += ` AND items.unread = :unread`
		args = append(args, sql.Named("unread", boolToInt(*params.Unread)))
	}

	// ORDER BY cannot use named parameters, validated via allowlist instead
	orderBy := "items.pub_date DESC, items.id DESC"
	if params.OrderBy == "created_at" {
		orderBy = "items.created_at DESC, items.id DESC"
	}
	query += ` ORDER BY ` + orderBy

	if params.Limit > 0 {
		query += ` LIMIT :limit`
		args = append(args, sql.Named("limit", params.Limit))
	}
	if params.Offset > 0 {
		query += ` OFFSET :offset`
		args = append(args, sql.Named("offset", params.Offset))
	}

	rows, err := s.query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*model.Item{}
	for rows.Next() {
		i := &model.Item{}
		var unread int
		if err := rows.Scan(
			&i.ID,
			&i.FeedID,
			&i.GUID,
			&i.Title,
			&i.Link,
			&i.Content,
			&i.PubDate,
			&unread,
			&i.CreatedAt,
			&i.Bookmarked,
			&i.ReadLater,
		); err != nil {
			return nil, err
		}
		i.Unread = intToBool(unread)
		items = append(items, i)
	}
	return items, rows.Err()
}

func (s *Store) GetItem(id int64) (*model.Item, error) {
	i := &model.Item{}
	var unread int
	err := s.queryRow(`
		SELECT id, feed_id, guid, title, link, content, pub_date, unread, created_at,
		       EXISTS(SELECT 1 FROM bookmarks WHERE bookmarks.item_id = items.id) AS bookmarked,
		       EXISTS(SELECT 1 FROM read_later_items WHERE read_later_items.item_id = items.id) AS read_later
		FROM items
		WHERE id = :id
	`, sql.Named("id", id)).Scan(
		&i.ID,
		&i.FeedID,
		&i.GUID,
		&i.Title,
		&i.Link,
		&i.Content,
		&i.PubDate,
		&unread,
		&i.CreatedAt,
		&i.Bookmarked,
		&i.ReadLater,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: item", ErrNotFound)
		}
		return nil, fmt.Errorf("get item: %w", err)
	}

	i.Unread = intToBool(unread)
	return i, nil
}

func (s *Store) CreateItem(feedID int64, guid, title, link, content string, pubDate int64) (*model.Item, error) {
	id, err := s.insertAndReturnID(s.db, `
		INSERT INTO items (feed_id, guid, title, link, content, pub_date)
		VALUES (:feed_id, :guid, :title, :link, :content, :pub_date)
	`, sql.Named("feed_id", feedID), sql.Named("guid", guid), sql.Named("title", title),
		sql.Named("link", link), sql.Named("content", content), sql.Named("pub_date", pubDate))
	if err != nil {
		return nil, err
	}

	return s.GetItem(id)
}

type BatchCreateItemInput struct {
	GUID    string
	Title   string
	Link    string
	Content string
	PubDate int64
}

// BatchCreateItemsIgnore inserts items in one transaction and ignores duplicates by (feed_id, guid).
// Returns the number of newly inserted rows.
func (s *Store) BatchCreateItemsIgnore(feedID int64, inputs []BatchCreateItemInput) (int, error) {
	if len(inputs) == 0 {
		return 0, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	created := 0
	for _, input := range inputs {
		result, err := s.execWith(tx, `
			INSERT INTO items (feed_id, guid, title, link, content, pub_date)
			VALUES (:feed_id, :guid, :title, :link, :content, :pub_date)
			ON CONFLICT(feed_id, guid) DO NOTHING
		`,
			sql.Named("feed_id", feedID),
			sql.Named("guid", input.GUID),
			sql.Named("title", input.Title),
			sql.Named("link", input.Link),
			sql.Named("content", input.Content),
			sql.Named("pub_date", input.PubDate),
		)
		if err != nil {
			return 0, err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		if affected > 0 {
			created++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return created, nil
}

func (s *Store) UpdateItemUnread(id int64, unread bool) error {
	result, err := s.exec(`UPDATE items SET unread = :unread WHERE id = :id`,
		sql.Named("unread", boolToInt(unread)), sql.Named("id", id))
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: item", ErrNotFound)
	}
	return nil
}

// BatchUpdateItemsUnread marks multiple items as read/unread.
// IDs are chunked to keep SQL statements bounded and avoid oversized IN clauses.
func (s *Store) BatchUpdateItemsUnread(ids []int64, unread bool) error {
	if len(ids) == 0 {
		return nil
	}

	const chunkSize = 500
	for start := 0; start < len(ids); start += chunkSize {
		end := min(start+chunkSize, len(ids))

		if err := s.batchUpdateItemsUnreadChunk(ids[start:end], unread); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) batchUpdateItemsUnreadChunk(ids []int64, unread bool) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, sql.Named("unread", boolToInt(unread)))
	for i, id := range ids {
		paramName := fmt.Sprintf("id%d", i)
		placeholders[i] = ":" + paramName
		args = append(args, sql.Named(paramName, id))
	}

	query := fmt.Sprintf(`UPDATE items SET unread = :unread WHERE id IN (%s)`, strings.Join(placeholders, ","))
	_, err := s.exec(query, args...)
	return err
}

// MarkAllAsRead marks items as read. If feedID is nil, marks ALL items across all feeds.
// If feedID is non-nil, only marks items from that specific feed.
func (s *Store) MarkAllAsRead(feedID *int64) error {
	if feedID != nil {
		_, err := s.exec(`UPDATE items SET unread = 0 WHERE feed_id = :feed_id`, sql.Named("feed_id", *feedID))
		return err
	}
	_, err := s.exec(`UPDATE items SET unread = 0`)
	return err
}

func (s *Store) MarkGroupAsRead(groupID int64) error {
	_, err := s.exec(`
		UPDATE items
		SET unread = 0
		WHERE feed_id IN (
			SELECT id
			FROM feeds
			WHERE group_id = :group_id
		)
	`, sql.Named("group_id", groupID))
	return err
}

func (s *Store) MarkFeedAsReadBefore(feedID, before int64) error {
	_, err := s.exec(`
		UPDATE items
		SET unread = 0
		WHERE feed_id = :feed_id
		  AND (CASE WHEN pub_date > 0 THEN pub_date ELSE created_at END) <= :before
	`, sql.Named("feed_id", feedID), sql.Named("before", before))
	return err
}

func (s *Store) MarkGroupAsReadBefore(groupID, before int64) error {
	_, err := s.exec(`
		UPDATE items
		SET unread = 0
		WHERE feed_id IN (
			SELECT id
			FROM feeds
			WHERE group_id = :group_id
		)
		  AND (CASE WHEN pub_date > 0 THEN pub_date ELSE created_at END) <= :before
	`, sql.Named("group_id", groupID), sql.Named("before", before))
	return err
}

func (s *Store) MarkAllAsReadBefore(before int64) error {
	_, err := s.exec(`
		UPDATE items
		SET unread = 0
		WHERE (CASE WHEN pub_date > 0 THEN pub_date ELSE created_at END) <= :before
	`, sql.Named("before", before))
	return err
}

func (s *Store) ListUnreadItemIDs() ([]int64, error) {
	rows, err := s.query(`
		SELECT id
		FROM items
		WHERE unread = 1
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

type ListFeverItemsParams struct {
	WithIDs []int64
	SinceID *int64
	MaxID   *int64
	Limit   int
	SortAsc bool
}

func (s *Store) ListFeverItems(params ListFeverItemsParams) ([]*model.Item, error) {
	query := `
		SELECT id, feed_id, guid, title, link, content, pub_date, unread, created_at
		FROM items
		WHERE 1=1
	`
	args := []any{}

	if len(params.WithIDs) > 0 {
		placeholders := make([]string, len(params.WithIDs))
		for i, id := range params.WithIDs {
			name := fmt.Sprintf("with_id_%d", i)
			placeholders[i] = ":" + name
			args = append(args, sql.Named(name, id))
		}
		query += fmt.Sprintf(" AND id IN (%s)", strings.Join(placeholders, ","))
	}

	if params.SinceID != nil {
		query += ` AND id > :since_id`
		args = append(args, sql.Named("since_id", *params.SinceID))
	}

	if params.MaxID != nil {
		query += ` AND id <= :max_id`
		args = append(args, sql.Named("max_id", *params.MaxID))
	}

	orderBy := "DESC"
	if params.SortAsc {
		orderBy = "ASC"
	}
	query += ` ORDER BY id ` + orderBy

	if params.Limit > 0 {
		query += ` LIMIT :limit`
		args = append(args, sql.Named("limit", params.Limit))
	}

	rows, err := s.query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*model.Item{}
	for rows.Next() {
		i := &model.Item{}
		var unread int
		if err := rows.Scan(&i.ID, &i.FeedID, &i.GUID, &i.Title, &i.Link, &i.Content, &i.PubDate, &unread, &i.CreatedAt); err != nil {
			return nil, err
		}
		i.Unread = intToBool(unread)
		items = append(items, i)
	}

	return items, rows.Err()
}

func (s *Store) ItemExists(feedID int64, guid string) (bool, error) {
	var exists bool
	err := s.queryRow(`SELECT EXISTS(SELECT 1 FROM items WHERE feed_id = :feed_id AND guid = :guid)`,
		sql.Named("feed_id", feedID), sql.Named("guid", guid)).Scan(&exists)
	return exists, err
}

type SearchItemResult struct {
	ID      int64  `json:"id"`
	FeedID  int64  `json:"feed_id"`
	Title   string `json:"title"`
	PubDate int64  `json:"pub_date"`
}

func (s *Store) SearchItems(query string, limit int) ([]*SearchItemResult, error) {
	tsQuery := buildPrefixTSQuery(query)
	if tsQuery == "" {
		return s.searchItemsLike(query, limit)
	}

	return s.searchItemsFTS(tsQuery, limit)
}

var searchTokenPattern = regexp.MustCompile(`[\p{L}\p{N}_]+`)

func buildPrefixTSQuery(query string) string {
	tokens := searchTokenPattern.FindAllString(strings.ToLower(query), -1)
	if len(tokens) == 0 {
		return ""
	}

	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		parts = append(parts, token+":*")
	}
	return strings.Join(parts, " & ")
}

func (s *Store) searchItemsFTS(query string, limit int) ([]*SearchItemResult, error) {
	rows, err := s.query(`
		WITH search_query AS (
			SELECT to_tsquery('simple', :query) AS value
		)
		SELECT items.id, items.feed_id, items.title, items.pub_date
		FROM items, search_query
		WHERE to_tsvector('simple', COALESCE(items.title, '') || ' ' || COALESCE(items.content, '')) @@ search_query.value
		ORDER BY items.pub_date DESC, items.id DESC
		LIMIT :limit
	`, sql.Named("query", query), sql.Named("limit", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*SearchItemResult{}
	for rows.Next() {
		i := &SearchItemResult{}
		if err := rows.Scan(&i.ID, &i.FeedID, &i.Title, &i.PubDate); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (s *Store) searchItemsLike(query string, limit int) ([]*SearchItemResult, error) {
	rows, err := s.query(`
		SELECT id, feed_id, title, pub_date
		FROM items
		WHERE title LIKE :query OR content LIKE :query
		ORDER BY pub_date DESC, id DESC
		LIMIT :limit
	`, sql.Named("query", "%"+query+"%"), sql.Named("limit", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*SearchItemResult{}
	for rows.Next() {
		i := &SearchItemResult{}
		if err := rows.Scan(&i.ID, &i.FeedID, &i.Title, &i.PubDate); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// CountItems returns the total count of items matching the filter criteria.
func (s *Store) CountItems(params ListItemsParams) (int, error) {
	query := `SELECT COUNT(*) FROM items`
	args := []any{}

	if params.GroupID != nil {
		query += ` INNER JOIN feeds ON items.feed_id = feeds.id`
	}

	query += ` WHERE 1=1`

	if params.FeedID != nil {
		query += ` AND items.feed_id = :feed_id`
		args = append(args, sql.Named("feed_id", *params.FeedID))
	}
	if params.GroupID != nil {
		query += ` AND feeds.group_id = :group_id`
		args = append(args, sql.Named("group_id", *params.GroupID))
	}
	if params.Unread != nil {
		query += ` AND items.unread = :unread`
		args = append(args, sql.Named("unread", boolToInt(*params.Unread)))
	}

	var count int
	err := s.queryRow(query, args...).Scan(&count)
	return count, err
}

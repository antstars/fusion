package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/0x2E/fusion/internal/model"
)

func (s *Store) CleanupItemsByRetention(settings model.RetentionSettings, now time.Time) (int, error) {
	deleted := 0

	if settings.RetentionDays > 0 {
		cutoff := now.AddDate(0, 0, -settings.RetentionDays).Unix()
		affected, err := s.deleteItemsWhere(`
			(CASE WHEN pub_date > 0 THEN pub_date ELSE created_at END) < :cutoff
		`, sql.Named("cutoff", cutoff))
		if err != nil {
			return deleted, err
		}
		deleted += affected
	}

	if settings.MaxArticles > 0 {
		affected, err := s.deleteItemsExceedingLimit(settings.MaxArticles)
		if err != nil {
			return deleted, err
		}
		deleted += affected
	}

	return deleted, nil
}

func (s *Store) CleanupItemsByCurrentRetention(now time.Time) (int, error) {
	settings, err := s.GetRetentionSettings()
	if err != nil {
		return 0, err
	}

	return s.CleanupItemsByRetention(*settings, now)
}

func (s *Store) deleteItemsWhere(where string, args ...any) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if _, err := s.execWith(tx, `
		UPDATE bookmarks
		SET item_id = NULL
		WHERE item_id IN (SELECT id FROM items WHERE `+where+`)
	`, args...); err != nil {
		return 0, err
	}

	if _, err := s.execWith(tx, `
		UPDATE read_later_items
		SET item_id = NULL
		WHERE item_id IN (SELECT id FROM items WHERE `+where+`)
	`, args...); err != nil {
		return 0, err
	}

	result, err := s.execWith(tx, `DELETE FROM items WHERE `+where, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int(affected), nil
}

func (s *Store) deleteItemsExceedingLimit(maxArticles int) (int, error) {
	if maxArticles <= 0 {
		return 0, nil
	}

	ids, err := s.listItemIDsExceedingLimit(maxArticles)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		name := fmt.Sprintf("id%d", i)
		placeholders[i] = ":" + name
		args = append(args, sql.Named(name, id))
	}

	return s.deleteItemsWhere("id IN ("+strings.Join(placeholders, ",")+")", args...)
}

func (s *Store) listItemIDsExceedingLimit(maxArticles int) ([]int64, error) {
	rows, err := s.query(`
		SELECT id
		FROM items
		ORDER BY (CASE WHEN pub_date > 0 THEN pub_date ELSE created_at END) DESC, id DESC
		OFFSET :max_articles
	`, sql.Named("max_articles", maxArticles))
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

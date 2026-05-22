package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/0x2E/fusion/internal/model"
)

const (
	settingMaxArticlesKey   = "max_articles"
	settingRetentionDaysKey = "retention_days"
	defaultMaxArticles      = 0
	defaultRetentionDays    = 30
)

func (s *Store) GetRetentionSettings() (*model.RetentionSettings, error) {
	values, err := s.getAppSettings(settingMaxArticlesKey, settingRetentionDaysKey)
	if err != nil {
		return nil, err
	}

	maxArticles, err := parseIntSetting(values[settingMaxArticlesKey], defaultMaxArticles)
	if err != nil {
		return nil, fmt.Errorf("parse max articles setting: %w", err)
	}

	retentionDays, err := parseIntSetting(values[settingRetentionDaysKey], defaultRetentionDays)
	if err != nil {
		return nil, fmt.Errorf("parse retention days setting: %w", err)
	}

	return &model.RetentionSettings{
		MaxArticles:   maxArticles,
		RetentionDays: retentionDays,
	}, nil
}

func (s *Store) UpdateRetentionSettings(settings model.RetentionSettings) (*model.RetentionSettings, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for key, value := range map[string]int{
		settingMaxArticlesKey:   settings.MaxArticles,
		settingRetentionDaysKey: settings.RetentionDays,
	} {
		if _, err := s.execWith(tx, `
			INSERT INTO app_settings (key, value)
			VALUES (:key, :value)
			ON CONFLICT(key) DO UPDATE SET
				value = excluded.value,
				updated_at = CAST(EXTRACT(EPOCH FROM NOW()) AS BIGINT)
		`, sql.Named("key", key), sql.Named("value", strconv.Itoa(value))); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetRetentionSettings()
}

func (s *Store) getAppSettings(keys ...string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		var value string
		err := s.queryRow(`
			SELECT value
			FROM app_settings
			WHERE key = :key
		`, sql.Named("key", key)).Scan(&value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, err
		}
		values[key] = value
	}

	return values, nil
}

func parseIntSetting(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}

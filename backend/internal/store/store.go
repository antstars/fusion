// Package store provides the PostgreSQL data access layer for Fusion RSS reader.
//
// All timestamps are stored as Unix epoch seconds (BIGINT).
// Boolean fields are stored as INTEGER (0/1) and converted to/from Go bool.
// Named SQL parameters (:param_name) are used throughout for safety.
package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/0x2E/fusion/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	db                       *sql.DB
	itemsSearchVectorIndexed bool
}

func New(databaseURL string) (*Store, error) {
	return NewWithPool(databaseURL, config.DatabasePoolConfig{
		MaxOpenConns:           20,
		MaxIdleConns:           10,
		ConnMaxLifetimeMinutes: 30,
		ConnMaxIdleTimeMinutes: 10,
	})
}

func NewWithPool(databaseURL string, pool config.DatabasePoolConfig) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(pool.MaxOpenConns)
	db.SetMaxIdleConns(pool.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(pool.ConnMaxLifetimeMinutes) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(pool.ConnMaxIdleTimeMinutes) * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	if err := s.detectCapabilities(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) detectCapabilities() error {
	return s.queryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = 'items'
			  AND column_name = 'search_vector'
		)
	`).Scan(&s.itemsSearchVectorIndexed)
}

func (s *Store) Close() error {
	return s.db.Close()
}

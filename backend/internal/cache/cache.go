package cache

import (
	"context"
	"errors"
	"time"
)

var ErrMiss = errors.New("cache miss")

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	DeletePrefix(ctx context.Context, prefix string) error
	Close() error
}

type NoopCache struct{}

func (NoopCache) Get(context.Context, string) ([]byte, error) {
	return nil, ErrMiss
}

func (NoopCache) Set(context.Context, string, []byte, time.Duration) error {
	return nil
}

func (NoopCache) DeletePrefix(context.Context, string) error {
	return nil
}

func (NoopCache) Close() error {
	return nil
}

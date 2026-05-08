package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/0x2E/fusion/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client    *redis.Client
	scanCount int64
}

func NewRedis(redisURL string) (*RedisCache, error) {
	options, err := redis.ParseURL(strings.TrimSpace(redisURL))
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	return newRedisClient(options, 500)
}

func NewRedisWithConfig(cfg config.RedisConfig) (*RedisCache, error) {
	if strings.TrimSpace(cfg.URL) != "" {
		options, err := redis.ParseURL(strings.TrimSpace(cfg.URL))
		if err != nil {
			return nil, fmt.Errorf("parse redis url: %w", err)
		}
		return newRedisClient(options, cfg.ScanCount)
	}

	return newRedisClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolTimeout:  cfg.PoolTimeout,
	}, cfg.ScanCount)
}

func newRedisClient(options *redis.Options, scanCount int64) (*RedisCache, error) {
	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisCache{client: client, scanCount: scanCount}, nil
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrMiss
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) DeletePrefix(ctx context.Context, prefix string) error {
	iter := c.client.Scan(ctx, 0, prefix+"*", c.scanCount).Iterator()
	keys := make([]string, 0, 100)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 100 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

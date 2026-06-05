package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/cache"
	"github.com/gin-gonic/gin"
)

type fakeCache struct {
	mu         sync.Mutex
	values     map[string][]byte
	getErr     error
	setErr     error
	deleteErr  error
	incrErr    error
	deletes    int
	increments int
}

func newFakeCache() *fakeCache {
	return &fakeCache{values: map[string][]byte{}}
}

func (c *fakeCache) Get(_ context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.getErr != nil {
		return nil, c.getErr
	}
	value, ok := c.values[key]
	if !ok {
		return nil, cache.ErrMiss
	}
	return append([]byte(nil), value...), nil
}

func (c *fakeCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.setErr != nil {
		return c.setErr
	}
	c.values[key] = append([]byte(nil), value...)
	return nil
}

func (c *fakeCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.deleteErr != nil {
		return c.deleteErr
	}
	delete(c.values, key)
	return nil
}

func (c *fakeCache) Increment(_ context.Context, key string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.increments++
	if c.incrErr != nil {
		return 0, c.incrErr
	}
	value := int64(1)
	if raw, ok := c.values[key]; ok {
		parsed, err := strconv.ParseInt(string(raw), 10, 64)
		if err != nil {
			return 0, err
		}
		value = parsed + 1
	}
	c.values[key] = []byte(strconv.FormatInt(value, 10))
	return value, nil
}

func (c *fakeCache) DeletePrefix(_ context.Context, prefix string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deletes++
	if c.deleteErr != nil {
		return c.deleteErr
	}
	for key := range c.values {
		if strings.HasPrefix(key, prefix) {
			delete(c.values, key)
		}
	}
	return nil
}

func (c *fakeCache) Close() error {
	return nil
}

func TestCacheMiddlewareCachesGetResponses(t *testing.T) {
	responseCache := newFakeCache()
	h := &Handler{cache: responseCache, cacheTTL: time.Minute}
	r := gin.New()
	r.Use(h.cacheMiddleware())

	calls := 0
	r.GET("/api/groups", func(c *gin.Context) {
		calls++
		c.JSON(http.StatusOK, gin.H{"calls": calls})
	})

	first := performRequest(r, http.MethodGet, "/api/groups", nil, nil, &http.Cookie{Name: "session", Value: "s1"})
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d", first.Code)
	}

	second := performRequest(r, http.MethodGet, "/api/groups", nil, nil, &http.Cookie{Name: "session", Value: "s1"})
	if second.Code != http.StatusOK {
		t.Fatalf("second status = %d", second.Code)
	}
	if calls != 1 {
		t.Fatalf("expected cached second request, handler calls = %d", calls)
	}
}

func TestCacheMiddlewareInvalidatesAfterMutation(t *testing.T) {
	responseCache := newFakeCache()
	h := &Handler{cache: responseCache, cacheTTL: time.Minute}
	r := gin.New()
	r.Use(h.cacheMiddleware())

	r.GET("/api/groups", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.POST("/api/groups", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"created": true})
	})

	performRequest(r, http.MethodGet, "/api/groups", nil, nil, &http.Cookie{Name: "session", Value: "s1"})
	performRequest(r, http.MethodPost, "/api/groups", strings.NewReader(`{"name":"Work"}`), map[string]string{"Content-Type": "application/json"})

	if responseCache.increments != 1 {
		t.Fatalf("expected one cache version increment, got %d", responseCache.increments)
	}
	if responseCache.deletes != 0 {
		t.Fatalf("expected no prefix deletes, got %d", responseCache.deletes)
	}
	if _, ok := responseCache.values[readCacheVersionKey]; !ok {
		t.Fatal("expected cache version key to be set")
	}
}

func TestCacheMiddlewareFallsBackWhenCacheFails(t *testing.T) {
	responseCache := newFakeCache()
	responseCache.getErr = errors.New("boom")
	responseCache.setErr = errors.New("boom")
	h := &Handler{cache: responseCache, cacheTTL: time.Minute}
	r := gin.New()
	r.Use(h.cacheMiddleware())

	r.GET("/api/groups", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	res := performRequest(r, http.MethodGet, "/api/groups", nil, nil, &http.Cookie{Name: "session", Value: "s1"})
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestRefreshEventsPathIsNotCacheable(t *testing.T) {
	h := &Handler{}

	if h.isCacheablePath("/api/feeds/refresh-events") {
		t.Fatal("expected refresh events path to bypass cache")
	}
}

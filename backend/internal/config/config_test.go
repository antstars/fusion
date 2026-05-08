package config

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

const testDatabaseURL = "postgres://fusion:fusion@localhost:5432/fusion?sslmode=disable"

func TestLoadParsesCORSAndPrivateFeedSettings(t *testing.T) {
	t.Setenv("FUSION_PASSWORD", "secret")
	t.Setenv("FUSION_DATABASE_URL", testDatabaseURL)
	t.Setenv("FUSION_CORS_ALLOWED_ORIGINS", " https://app.example.com , , https://admin.example.com/ ")
	t.Setenv("FUSION_TRUSTED_PROXIES", " 10.0.0.1 , 192.168.1.0/24 ")
	t.Setenv("FUSION_ALLOW_PRIVATE_FEEDS", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected 2 allowed origins, got %d", len(cfg.CORSAllowedOrigins))
	}
	if cfg.CORSAllowedOrigins[0] != "https://app.example.com" {
		t.Fatalf("unexpected first origin: %q", cfg.CORSAllowedOrigins[0])
	}
	if cfg.CORSAllowedOrigins[1] != "https://admin.example.com/" {
		t.Fatalf("unexpected second origin: %q", cfg.CORSAllowedOrigins[1])
	}
	if !cfg.AllowPrivateFeeds {
		t.Fatal("expected AllowPrivateFeeds to be true")
	}
	if len(cfg.TrustedProxies) != 2 {
		t.Fatalf("expected 2 trusted proxies, got %d", len(cfg.TrustedProxies))
	}
	if cfg.TrustedProxies[0] != "10.0.0.1" {
		t.Fatalf("unexpected first trusted proxy: %q", cfg.TrustedProxies[0])
	}
	if cfg.TrustedProxies[1] != "192.168.1.0/24" {
		t.Fatalf("unexpected second trusted proxy: %q", cfg.TrustedProxies[1])
	}
}

func TestLoadUsesDefaultPullMaxBackoff(t *testing.T) {
	t.Setenv("FUSION_PASSWORD", "secret")
	t.Setenv("FUSION_DATABASE_URL", testDatabaseURL)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.PullMaxBackoff != 172800 {
		t.Fatalf("expected default PullMaxBackoff to be 172800, got %d", cfg.PullMaxBackoff)
	}
}

func TestLoadFeverUsername(t *testing.T) {
	t.Run("uses default username", func(t *testing.T) {
		t.Setenv("FUSION_PASSWORD", "secret")
		t.Setenv("FUSION_DATABASE_URL", testDatabaseURL)

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.FeverUsername != "fusion" {
			t.Fatalf("expected default FeverUsername to be %q, got %q", "fusion", cfg.FeverUsername)
		}
	})

	t.Run("uses explicit username", func(t *testing.T) {
		t.Setenv("FUSION_PASSWORD", "secret")
		t.Setenv("FUSION_DATABASE_URL", testDatabaseURL)
		t.Setenv("FUSION_FEVER_USERNAME", " reader ")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.FeverUsername != "reader" {
			t.Fatalf("expected FeverUsername to be %q, got %q", "reader", cfg.FeverUsername)
		}
	})
}

func TestLoadParsesKubernetesStyleFusionPort(t *testing.T) {
	t.Setenv("FUSION_PASSWORD", "secret")
	t.Setenv("FUSION_DATABASE_URL", testDatabaseURL)
	t.Setenv("FUSION_PORT", "tcp://10.43.157.55:8080")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Port != 8080 {
		t.Fatalf("expected Port to be 8080, got %d", cfg.Port)
	}
}

func TestLoadRejectsInvalidFusionPort(t *testing.T) {
	t.Setenv("FUSION_PASSWORD", "secret")
	t.Setenv("FUSION_DATABASE_URL", testDatabaseURL)
	t.Setenv("FUSION_PORT", "tcp://10.43.157.55")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load() to fail for invalid FUSION_PORT")
	}
	if !strings.Contains(err.Error(), "invalid FUSION_PORT") {
		t.Fatalf("expected error to mention invalid FUSION_PORT, got %v", err)
	}
}

func TestLoadParsesPostgresAndRedisSettings(t *testing.T) {
	t.Setenv("FUSION_PASSWORD", "secret")
	t.Setenv("FUSION_DATABASE_HOST", "192.168.2.6")
	t.Setenv("FUSION_DATABASE_PORT", "5433")
	t.Setenv("FUSION_DATABASE_USER", "postgres")
	t.Setenv("FUSION_DATABASE_PASSWORD", "secret!")
	t.Setenv("FUSION_DATABASE_NAME", "fusion")
	t.Setenv("FUSION_DATABASE_SSLMODE", "disable")
	t.Setenv("FUSION_DATABASE_MAX_OPEN_CONNS", "64")
	t.Setenv("FUSION_DATABASE_MAX_IDLE_CONNS", "32")
	t.Setenv("FUSION_DATABASE_CONN_MAX_LIFETIME_MINUTES", "30")
	t.Setenv("FUSION_DATABASE_CONN_MAX_IDLE_TIME_MINUTES", "10")
	t.Setenv("FUSION_REDIS_ENABLED", "true")
	t.Setenv("FUSION_REDIS_ADDR", "192.168.2.6:6379")
	t.Setenv("FUSION_REDIS_PASSWORD", "redis-secret")
	t.Setenv("FUSION_REDIS_DB", "15")
	t.Setenv("FUSION_CACHE_TTL_SECONDS", "600")
	t.Setenv("FUSION_REDIS_POOL_SIZE", "80")
	t.Setenv("FUSION_REDIS_MIN_IDLE_CONNS", "16")
	t.Setenv("FUSION_REDIS_DIAL_TIMEOUT_SECONDS", "2")
	t.Setenv("FUSION_REDIS_READ_TIMEOUT_SECONDS", "3")
	t.Setenv("FUSION_REDIS_WRITE_TIMEOUT_SECONDS", "4")
	t.Setenv("FUSION_REDIS_POOL_TIMEOUT_SECONDS", "5")
	t.Setenv("FUSION_REDIS_SCAN_COUNT", "500")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	parsed, err := url.Parse(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("parse DatabaseURL: %v", err)
	}
	if parsed.Host != "192.168.2.6:5433" || parsed.User.Username() != "postgres" || parsed.Path != "/fusion" {
		t.Fatalf("unexpected DatabaseURL: %q", cfg.DatabaseURL)
	}
	if password, _ := parsed.User.Password(); password != "secret!" {
		t.Fatalf("unexpected database password: %q", password)
	}
	if parsed.Query().Get("sslmode") != "disable" {
		t.Fatalf("unexpected sslmode: %q", parsed.Query().Get("sslmode"))
	}
	if cfg.DatabasePool.MaxOpenConns != 64 || cfg.DatabasePool.MaxIdleConns != 32 {
		t.Fatalf("unexpected database pool config: %+v", cfg.DatabasePool)
	}
	if !cfg.RedisEnabled {
		t.Fatal("expected RedisEnabled to be true")
	}
	if cfg.Redis.Addr != "192.168.2.6:6379" {
		t.Fatalf("unexpected Redis addr: %q", cfg.Redis.Addr)
	}
	if cfg.Redis.Password != "redis-secret" || cfg.Redis.DB != 15 {
		t.Fatalf("unexpected Redis auth/db config: %+v", cfg.Redis)
	}
	if cfg.Redis.CacheTTLSeconds != 600 {
		t.Fatalf("expected CacheTTLSeconds 600, got %d", cfg.Redis.CacheTTLSeconds)
	}
	if cfg.Redis.PoolSize != 80 || cfg.Redis.MinIdleConns != 16 {
		t.Fatalf("unexpected Redis pool config: %+v", cfg.Redis)
	}
	if cfg.Redis.DialTimeout != 2*time.Second || cfg.Redis.ReadTimeout != 3*time.Second || cfg.Redis.WriteTimeout != 4*time.Second || cfg.Redis.PoolTimeout != 5*time.Second {
		t.Fatalf("unexpected Redis timeouts: %+v", cfg.Redis)
	}
	if cfg.Redis.ScanCount != 500 {
		t.Fatalf("expected ScanCount 500, got %d", cfg.Redis.ScanCount)
	}
}

func TestLoadURLOverridesStructuredSettings(t *testing.T) {
	t.Setenv("FUSION_PASSWORD", "secret")
	t.Setenv("FUSION_DATABASE_HOST", "ignored")
	t.Setenv("FUSION_DATABASE_USER", "ignored")
	t.Setenv("FUSION_DATABASE_NAME", "ignored")
	t.Setenv("FUSION_DATABASE_URL", testDatabaseURL)
	t.Setenv("FUSION_REDIS_ENABLED", "false")
	t.Setenv("FUSION_REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("FUSION_REDIS_SCAN_COUNT", "700")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.DatabaseURL != testDatabaseURL {
		t.Fatalf("expected database URL override, got %q", cfg.DatabaseURL)
	}
	if !cfg.RedisEnabled {
		t.Fatal("expected Redis URL override to enable Redis")
	}
	if cfg.Redis.URL != "redis://localhost:6379/0" {
		t.Fatalf("unexpected Redis URL override: %q", cfg.Redis.URL)
	}
	if cfg.Redis.ScanCount != 700 {
		t.Fatalf("expected Redis scan count to remain configurable, got %d", cfg.Redis.ScanCount)
	}
}

func TestLoadDisablesRedisByDefault(t *testing.T) {
	t.Setenv("FUSION_PASSWORD", "secret")
	t.Setenv("FUSION_DATABASE_URL", testDatabaseURL)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.RedisEnabled {
		t.Fatal("expected Redis to be disabled by default")
	}
}

func TestLoadRejectsInvalidDatabaseSettings(t *testing.T) {
	t.Run("missing postgres url", func(t *testing.T) {
		t.Setenv("FUSION_PASSWORD", "secret")

		_, err := Load()
		if err == nil || !strings.Contains(err.Error(), "FUSION_DATABASE_HOST is required") {
			t.Fatalf("expected missing database host error, got %v", err)
		}
	})

	t.Run("invalid redis url", func(t *testing.T) {
		t.Setenv("FUSION_PASSWORD", "secret")
		t.Setenv("FUSION_DATABASE_URL", testDatabaseURL)
		t.Setenv("FUSION_REDIS_URL", "http://localhost:6379")

		_, err := Load()
		if err == nil || !strings.Contains(err.Error(), "invalid FUSION_REDIS_URL") {
			t.Fatalf("expected invalid redis url error, got %v", err)
		}
	})

	t.Run("invalid numeric setting", func(t *testing.T) {
		t.Setenv("FUSION_PASSWORD", "secret")
		t.Setenv("FUSION_DATABASE_URL", testDatabaseURL)
		t.Setenv("FUSION_REDIS_POOL_SIZE", "zero")

		_, err := Load()
		if err == nil || !strings.Contains(err.Error(), "invalid FUSION_REDIS_POOL_SIZE") {
			t.Fatalf("expected invalid redis pool size error, got %v", err)
		}
	})
}

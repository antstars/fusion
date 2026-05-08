package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL   string
	DatabasePool  DatabasePoolConfig
	Password      string // Plaintext password from env
	Port          int
	FeverUsername string // Username used to derive Fever API key.

	CORSAllowedOrigins []string // Allowed Origins for CORS. Empty means allow all.
	TrustedProxies     []string // Trusted reverse proxies for client IP resolution. Empty disables proxy trust.
	AllowPrivateFeeds  bool     // Allow pulling private/localhost feed URLs.

	PullInterval    int // Pull interval in seconds (default: 1800 = 30 min)
	PullTimeout     int // Request timeout in seconds (default: 30)
	PullConcurrency int // Max concurrent pulls (default: 10)
	PullMaxBackoff  int // Global max scheduling delay in seconds (default: 172800 = 48 hours)

	LoginRateLimit int // Max failed login attempts per window (default: 10)
	LoginWindow    int // Login rate limit window in seconds (default: 60)
	LoginBlock     int // Login block duration in seconds (default: 300)

	LogLevel  string // Log level: DEBUG, INFO, WARN, ERROR (default: INFO)
	LogFormat string // Log format: text, json, auto (default: auto)

	RedisEnabled bool
	Redis        RedisConfig

	// OIDC Configuration (optional, enabled when OIDCIssuer is set)
	OIDCIssuer       string // OIDC provider URL
	OIDCClientID     string // OAuth2 client ID
	OIDCClientSecret string // OAuth2 client secret
	OIDCRedirectURI  string // Callback URL (required when OIDC is enabled)
	OIDCAllowedUser  string // Optional: restrict to specific user identity (email or sub)
}

type DatabasePoolConfig struct {
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeMinutes int
	ConnMaxIdleTimeMinutes int
}

type RedisConfig struct {
	URL             string
	Addr            string
	Password        string
	DB              int
	CacheTTLSeconds int
	PoolSize        int
	MinIdleConns    int
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
	ScanCount       int64
}

func Load() (*Config, error) {
	// Backward compatible env vars:
	// - PASSWORD (legacy) -> FUSION_PASSWORD
	// - PORT (legacy) -> FUSION_PORT
	databaseURL, err := loadDatabaseURL()
	if err != nil {
		return nil, err
	}
	databasePool, err := loadDatabasePool()
	if err != nil {
		return nil, err
	}

	password := os.Getenv("FUSION_PASSWORD")
	if password == "" {
		password = os.Getenv("PASSWORD")
	}

	allowEmptyPassword, err := getEnvBool("FUSION_ALLOW_EMPTY_PASSWORD", false)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(password) == "" && !allowEmptyPassword {
		return nil, fmt.Errorf("FUSION_PASSWORD is required (or set FUSION_ALLOW_EMPTY_PASSWORD=true)")
	}

	port := os.Getenv("FUSION_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}
	parsedPort, err := parsePort(port)
	if err != nil {
		return nil, fmt.Errorf("invalid FUSION_PORT: %w", err)
	}
	if parsedPort <= 0 || parsedPort > 65535 {
		return nil, fmt.Errorf("invalid FUSION_PORT: must be in range 1-65535")
	}

	pullInterval, err := getEnvInt("FUSION_PULL_INTERVAL", 1800, 1)
	if err != nil {
		return nil, err
	}
	pullTimeout, err := getEnvInt("FUSION_PULL_TIMEOUT", 30, 1)
	if err != nil {
		return nil, err
	}
	pullConcurrency, err := getEnvInt("FUSION_PULL_CONCURRENCY", 10, 1)
	if err != nil {
		return nil, err
	}
	pullMaxBackoff, err := getEnvInt("FUSION_PULL_MAX_BACKOFF", 172800, 1)
	if err != nil {
		return nil, err
	}

	loginRateLimit, err := getEnvInt("FUSION_LOGIN_RATE_LIMIT", 10, 1)
	if err != nil {
		return nil, err
	}
	loginWindow, err := getEnvInt("FUSION_LOGIN_WINDOW", 60, 1)
	if err != nil {
		return nil, err
	}
	loginBlock, err := getEnvInt("FUSION_LOGIN_BLOCK", 300, 1)
	if err != nil {
		return nil, err
	}

	corsAllowedOrigins := parseCSVEnv(os.Getenv("FUSION_CORS_ALLOWED_ORIGINS"))
	trustedProxies := parseCSVEnv(os.Getenv("FUSION_TRUSTED_PROXIES"))

	allowPrivateFeeds, err := getEnvBool("FUSION_ALLOW_PRIVATE_FEEDS", false)
	if err != nil {
		return nil, err
	}

	logLevel := os.Getenv("FUSION_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}

	logFormat := os.Getenv("FUSION_LOG_FORMAT")
	if logFormat == "" {
		logFormat = "auto"
	}

	redisEnabled, err := getEnvBool("FUSION_REDIS_ENABLED", false)
	if err != nil {
		return nil, err
	}
	redisConfig, err := loadRedisConfig()
	if err != nil {
		return nil, err
	}
	if redisConfig.URL != "" {
		redisEnabled = true
	}

	return &Config{
		DatabaseURL:        databaseURL,
		DatabasePool:       databasePool,
		Password:           password,
		Port:               parsedPort,
		FeverUsername:      getEnvString("FUSION_FEVER_USERNAME", "fusion"),
		CORSAllowedOrigins: corsAllowedOrigins,
		TrustedProxies:     trustedProxies,
		AllowPrivateFeeds:  allowPrivateFeeds,
		PullInterval:       pullInterval,
		PullTimeout:        pullTimeout,
		PullConcurrency:    pullConcurrency,
		PullMaxBackoff:     pullMaxBackoff,
		LoginRateLimit:     loginRateLimit,
		LoginWindow:        loginWindow,
		LoginBlock:         loginBlock,
		LogLevel:           logLevel,
		LogFormat:          logFormat,
		RedisEnabled:       redisEnabled,
		Redis:              redisConfig,

		OIDCIssuer:       os.Getenv("FUSION_OIDC_ISSUER"),
		OIDCClientID:     os.Getenv("FUSION_OIDC_CLIENT_ID"),
		OIDCClientSecret: os.Getenv("FUSION_OIDC_CLIENT_SECRET"),
		OIDCRedirectURI:  os.Getenv("FUSION_OIDC_REDIRECT_URI"),
		OIDCAllowedUser:  os.Getenv("FUSION_OIDC_ALLOWED_USER"),
	}, nil
}

func getEnvString(key, defaultVal string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}

	return val
}

func loadDatabaseURL() (string, error) {
	databaseURL := strings.TrimSpace(os.Getenv("FUSION_DATABASE_URL"))
	if databaseURL == "" {
		host := getEnvString("FUSION_DATABASE_HOST", "")
		if host == "" {
			return "", fmt.Errorf("FUSION_DATABASE_HOST is required (or set FUSION_DATABASE_URL)")
		}
		port, err := getEnvInt("FUSION_DATABASE_PORT", 5432, 1)
		if err != nil {
			return "", err
		}
		user := getEnvString("FUSION_DATABASE_USER", "")
		if user == "" {
			return "", fmt.Errorf("FUSION_DATABASE_USER is required (or set FUSION_DATABASE_URL)")
		}
		dbName := getEnvString("FUSION_DATABASE_NAME", "")
		if dbName == "" {
			return "", fmt.Errorf("FUSION_DATABASE_NAME is required (or set FUSION_DATABASE_URL)")
		}
		sslMode := getEnvString("FUSION_DATABASE_SSLMODE", "disable")
		if err := validatePostgresSSLMode(sslMode); err != nil {
			return "", err
		}

		u := &url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(user, os.Getenv("FUSION_DATABASE_PASSWORD")),
			Host:   fmt.Sprintf("%s:%d", host, port),
			Path:   dbName,
		}
		q := u.Query()
		q.Set("sslmode", sslMode)
		u.RawQuery = q.Encode()
		databaseURL = u.String()
	}

	if err := validatePostgresURL(databaseURL); err != nil {
		return "", err
	}
	return databaseURL, nil
}

func loadDatabasePool() (DatabasePoolConfig, error) {
	maxOpenConns, err := getEnvInt("FUSION_DATABASE_MAX_OPEN_CONNS", 64, 1)
	if err != nil {
		return DatabasePoolConfig{}, err
	}
	maxIdleConns, err := getEnvInt("FUSION_DATABASE_MAX_IDLE_CONNS", 32, 0)
	if err != nil {
		return DatabasePoolConfig{}, err
	}
	connMaxLifetime, err := getEnvInt("FUSION_DATABASE_CONN_MAX_LIFETIME_MINUTES", 30, 0)
	if err != nil {
		return DatabasePoolConfig{}, err
	}
	connMaxIdleTime, err := getEnvInt("FUSION_DATABASE_CONN_MAX_IDLE_TIME_MINUTES", 10, 0)
	if err != nil {
		return DatabasePoolConfig{}, err
	}

	return DatabasePoolConfig{
		MaxOpenConns:           maxOpenConns,
		MaxIdleConns:           maxIdleConns,
		ConnMaxLifetimeMinutes: connMaxLifetime,
		ConnMaxIdleTimeMinutes: connMaxIdleTime,
	}, nil
}

func loadRedisConfig() (RedisConfig, error) {
	redisURL := strings.TrimSpace(os.Getenv("FUSION_REDIS_URL"))
	if redisURL != "" {
		if err := validateRedisURL(redisURL); err != nil {
			return RedisConfig{}, err
		}
	}

	db, err := getEnvInt("FUSION_REDIS_DB", 0, 0)
	if err != nil {
		return RedisConfig{}, err
	}
	cacheTTLSeconds, err := getEnvInt("FUSION_CACHE_TTL_SECONDS", 120, 0)
	if err != nil {
		return RedisConfig{}, err
	}
	poolSize, err := getEnvInt("FUSION_REDIS_POOL_SIZE", 80, 1)
	if err != nil {
		return RedisConfig{}, err
	}
	minIdleConns, err := getEnvInt("FUSION_REDIS_MIN_IDLE_CONNS", 16, 0)
	if err != nil {
		return RedisConfig{}, err
	}
	dialTimeout, err := getEnvDurationSeconds("FUSION_REDIS_DIAL_TIMEOUT_SECONDS", 2, 0)
	if err != nil {
		return RedisConfig{}, err
	}
	readTimeout, err := getEnvDurationSeconds("FUSION_REDIS_READ_TIMEOUT_SECONDS", 2, 0)
	if err != nil {
		return RedisConfig{}, err
	}
	writeTimeout, err := getEnvDurationSeconds("FUSION_REDIS_WRITE_TIMEOUT_SECONDS", 2, 0)
	if err != nil {
		return RedisConfig{}, err
	}
	poolTimeout, err := getEnvDurationSeconds("FUSION_REDIS_POOL_TIMEOUT_SECONDS", 4, 0)
	if err != nil {
		return RedisConfig{}, err
	}
	scanCount, err := getEnvInt("FUSION_REDIS_SCAN_COUNT", 500, 1)
	if err != nil {
		return RedisConfig{}, err
	}

	return RedisConfig{
		URL:             redisURL,
		Addr:            getEnvString("FUSION_REDIS_ADDR", "127.0.0.1:6379"),
		Password:        os.Getenv("FUSION_REDIS_PASSWORD"),
		DB:              db,
		CacheTTLSeconds: cacheTTLSeconds,
		PoolSize:        poolSize,
		MinIdleConns:    minIdleConns,
		DialTimeout:     dialTimeout,
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		PoolTimeout:     poolTimeout,
		ScanCount:       int64(scanCount),
	}, nil
}

func validatePostgresURL(databaseURL string) error {
	parsedURL, err := url.Parse(databaseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid FUSION_DATABASE_URL")
	}
	if parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql" {
		return fmt.Errorf("invalid FUSION_DATABASE_URL: scheme must be postgres or postgresql")
	}
	return nil
}

func validatePostgresSSLMode(sslMode string) error {
	switch sslMode {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
		return nil
	default:
		return fmt.Errorf("invalid FUSION_DATABASE_SSLMODE")
	}
}

func validateRedisURL(redisURL string) error {
	parsedURL, err := url.Parse(redisURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid FUSION_REDIS_URL")
	}
	if parsedURL.Scheme != "redis" && parsedURL.Scheme != "rediss" {
		return fmt.Errorf("invalid FUSION_REDIS_URL: scheme must be redis or rediss")
	}
	return nil
}

func getEnvInt(key string, defaultVal, minVal int) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal, nil
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	if parsed < minVal {
		return 0, fmt.Errorf("invalid %s: must be >= %d", key, minVal)
	}
	return parsed, nil
}

func getEnvDurationSeconds(key string, defaultVal, minVal int) (time.Duration, error) {
	seconds, err := getEnvInt(key, defaultVal, minVal)
	if err != nil {
		return 0, err
	}
	return time.Duration(seconds) * time.Second, nil
}

func getEnvBool(key string, defaultVal bool) (bool, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal, nil
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return false, fmt.Errorf("invalid %s: %w", key, err)
	}
	return parsed, nil
}

// parsePort accepts plain numeric ports and Kubernetes service-link URL values
// such as tcp://10.43.157.55:8080.
func parsePort(val string) (int, error) {
	trimmed := strings.TrimSpace(val)
	parsed, err := strconv.Atoi(trimmed)
	if err == nil {
		return parsed, nil
	}

	if !strings.Contains(trimmed, "://") {
		return 0, err
	}

	parsedURL, err := url.Parse(trimmed)
	if err != nil {
		return 0, err
	}

	port := parsedURL.Port()
	if port == "" {
		return 0, fmt.Errorf("missing port")
	}

	return strconv.Atoi(port)
}

func parseCSVEnv(val string) []string {
	if strings.TrimSpace(val) == "" {
		return nil
	}

	parts := strings.Split(val, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		values = append(values, part)
	}

	if len(values) == 0 {
		return nil
	}

	return values
}

package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0x2E/fusion/internal/auth"
	"github.com/0x2E/fusion/internal/cache"
	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	store                *store.Store
	config               *config.Config
	cache                cache.Cache
	cacheTTL             time.Duration
	passwordHash         string // bcrypt hash computed at startup
	passwordLoginEnabled bool
	feverAPIKey          string // md5(username:password) used by Fever API
	allowAnonAPI         bool   // true when both password and OIDC auth are disabled
	puller               interface {
		RefreshFeed(ctx context.Context, feedID int64) error
		RefreshAll(ctx context.Context) (int, error)
	}
	sessions  map[string]int64        // sessionID -> unix expiry seconds
	mu        sync.RWMutex            // protects sessions state
	oidcAuth  *auth.OIDCAuthenticator // nil when OIDC is disabled
	limiter   *loginLimiter
	lastSweep int64

	refreshAllMu      sync.Mutex
	refreshAllRunning bool
}

func New(store *store.Store, config *config.Config, puller interface {
	RefreshFeed(ctx context.Context, feedID int64) error
	RefreshAll(ctx context.Context) (int, error)
}) (*Handler, error) {
	return NewWithCache(store, config, puller, cache.NoopCache{})
}

func NewWithCache(store *store.Store, config *config.Config, puller interface {
	RefreshFeed(ctx context.Context, feedID int64) error
	RefreshAll(ctx context.Context) (int, error)
}, responseCache cache.Cache) (*Handler, error) {
	password := config.Password
	passwordLoginEnabled := strings.TrimSpace(password) != ""
	passwordHash := ""
	feverAPIKey := ""
	if passwordLoginEnabled {
		var err error
		passwordHash, err = auth.HashPassword(password)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		feverAPIKey = deriveFeverAPIKey(config.FeverUsername, password)
	}
	cacheTTLSeconds := config.Redis.CacheTTLSeconds
	if cacheTTLSeconds == 0 {
		cacheTTLSeconds = 120
	}

	h := &Handler{
		store:                store,
		config:               config,
		cache:                responseCache,
		cacheTTL:             time.Duration(cacheTTLSeconds) * time.Second,
		passwordHash:         passwordHash,
		passwordLoginEnabled: passwordLoginEnabled,
		feverAPIKey:          feverAPIKey,
		allowAnonAPI:         strings.TrimSpace(config.Password) == "" && strings.TrimSpace(config.OIDCIssuer) == "",
		puller:               puller,
		sessions:             make(map[string]int64),
		limiter:              newLoginLimiter(config.LoginRateLimit, config.LoginWindow, config.LoginBlock),
	}

	if h.allowAnonAPI {
		slog.Warn("authentication is disabled because both password and OIDC are empty")
	}

	if config.OIDCIssuer != "" {
		if strings.TrimSpace(config.OIDCRedirectURI) == "" {
			return nil, fmt.Errorf("FUSION_OIDC_REDIRECT_URI is required when OIDC is enabled")
		}

		oidcAuth, err := auth.NewOIDC(
			context.Background(),
			config.OIDCIssuer,
			config.OIDCClientID,
			config.OIDCClientSecret,
			config.OIDCRedirectURI,
		)
		if err != nil {
			return nil, fmt.Errorf("initialize OIDC: %w", err)
		}
		if config.OIDCAllowedUser != "" {
			oidcAuth.SetAllowedUser(config.OIDCAllowedUser)
		}
		h.oidcAuth = oidcAuth
		slog.Info("OIDC authentication enabled", "issuer", config.OIDCIssuer)
	}

	return h, nil
}

func (h *Handler) SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(requestLogMiddleware(), recoveryMiddleware())

	if err := h.configureTrustedProxies(r); err != nil {
		slog.Warn("failed to configure trusted proxies", "error", err)
	}

	r.Use(h.corsMiddleware())
	r.POST("/fever", h.fever)
	r.POST("/fever/", h.fever)
	r.POST("/fever.php", h.fever)

	api := r.Group("/api")
	{
		api.POST("/sessions", h.login)
		api.DELETE("/sessions", h.logout)

		// OIDC routes (public, no auth middleware)
		api.GET("/oidc/enabled", h.oidcEnabled)
		if h.oidcAuth != nil {
			api.GET("/oidc/login", h.oidcLogin)
			api.GET("/oidc/callback", h.oidcCallback)
			// Compatibility route for deployments that configured redirect_uri without /api.
			r.GET("/oidc/callback", h.oidcCallback)
		}

		auth := api.Group("")
		auth.Use(h.authMiddleware())
		auth.Use(h.cacheMiddleware())
		{
			auth.GET("/groups", h.listGroups)
			auth.POST("/groups", h.createGroup)
			auth.GET("/groups/:id", h.getGroup)
			auth.PATCH("/groups/:id", h.updateGroup)
			auth.DELETE("/groups/:id", h.deleteGroup)

			auth.GET("/feeds", h.listFeeds)
			auth.POST("/feeds", h.createFeed)
			auth.POST("/feeds/batch", h.batchCreateFeeds)
			auth.POST("/feeds/refresh", h.refreshAllFeeds)
			auth.GET("/feeds/:id", h.getFeed)
			auth.PATCH("/feeds/:id", h.updateFeed)
			auth.DELETE("/feeds/:id", h.deleteFeed)
			auth.POST("/feeds/validate", h.validateFeed)
			auth.POST("/feeds/:id/refresh", h.refreshFeed)

			auth.GET("/items", h.listItems)
			auth.GET("/items/:id", h.getItem)
			auth.PATCH("/items/-/read", h.markItemsRead)
			auth.PATCH("/items/-/unread", h.markItemsUnread)

			auth.GET("/search", h.search)

			auth.GET("/bookmarks", h.listBookmarks)
			auth.POST("/bookmarks", h.createBookmark)
			auth.GET("/bookmarks/:id", h.getBookmark)
			auth.DELETE("/bookmarks/:id", h.deleteBookmark)

			auth.GET("/read-later", h.listReadLaterItems)
			auth.POST("/read-later", h.createReadLaterItem)
			auth.GET("/read-later/:id", h.getReadLaterItem)
			auth.DELETE("/read-later/:id", h.deleteReadLaterItem)
		}
	}

	if err := h.setupFrontendRoutes(r); err != nil {
		slog.Warn("failed to configure frontend routes", "error", err)
	}

	return r
}

type cacheResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w cacheResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (h *Handler) cacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.cacheTTL <= 0 || c.Request.Method != http.MethodGet || !h.isCacheablePath(c.Request.URL.Path) {
			c.Next()
			if c.Request.Method != http.MethodGet && c.Writer.Status() < 400 {
				h.invalidateReadCache(c.Request.Context())
			}
			return
		}

		key := h.cacheKey(c)
		if cached, err := h.cache.Get(c.Request.Context(), key); err == nil {
			c.Data(http.StatusOK, "application/json; charset=utf-8", cached)
			c.Abort()
			return
		} else if err != nil && !errors.Is(err, cache.ErrMiss) {
			slog.Warn("read cache get failed", "error", err)
		}

		writer := &cacheResponseWriter{ResponseWriter: c.Writer, body: bytes.NewBuffer(nil)}
		c.Writer = writer
		c.Next()

		if c.Writer.Status() == http.StatusOK && strings.HasPrefix(c.Writer.Header().Get("Content-Type"), "application/json") {
			if err := h.cache.Set(c.Request.Context(), key, writer.body.Bytes(), h.cacheTTL); err != nil {
				slog.Warn("read cache set failed", "error", err)
			}
		}
	}
}

func (h *Handler) isCacheablePath(path string) bool {
	for _, prefix := range []string{"/api/groups", "/api/feeds", "/api/items", "/api/bookmarks", "/api/read-later", "/api/search"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

func (h *Handler) cacheKey(c *gin.Context) string {
	sessionID, _ := c.Cookie("session")
	return "fusion:api:" + sessionID + ":" + c.Request.URL.RequestURI()
}

func (h *Handler) invalidateReadCache(ctx context.Context) {
	if err := h.cache.DeletePrefix(ctx, "fusion:api:"); err != nil {
		slog.Warn("read cache invalidation failed", "error", err)
	}
}

func (h *Handler) configureTrustedProxies(r *gin.Engine) error {
	if h.config == nil || len(h.config.TrustedProxies) == 0 {
		return r.SetTrustedProxies(nil)
	}

	return r.SetTrustedProxies(h.config.TrustedProxies)
}

func (h *Handler) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.Request.Header.Get("Origin"))
		if origin != "" {
			if !h.isOriginAllowed(origin) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (h *Handler) isOriginAllowed(origin string) bool {
	if h.config == nil {
		return true
	}

	if len(h.config.CORSAllowedOrigins) == 0 {
		return true
	}

	normalizedOrigin := normalizeOrigin(origin)
	for _, allowed := range h.config.CORSAllowedOrigins {
		normalizedAllowed := normalizeOrigin(allowed)
		if normalizedAllowed == "*" || normalizedAllowed == normalizedOrigin {
			return true
		}
	}

	return false
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	origin = strings.TrimSuffix(origin, "/")
	return strings.ToLower(origin)
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.allowAnonAPI {
			c.Next()
			return
		}

		sessionID, err := c.Cookie("session")
		if err != nil {
			unauthorizedError(c)
			c.Abort()
			return
		}

		if !h.isSessionValid(sessionID) {
			unauthorizedError(c)
			c.Abort()
			return
		}

		c.Next()
	}
}

func dataResponse(c *gin.Context, data any) {
	c.JSON(200, gin.H{"data": data})
}

func listResponse(c *gin.Context, data any, total int) {
	c.JSON(200, gin.H{"data": data, "total": total})
}

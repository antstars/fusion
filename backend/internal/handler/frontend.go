package handler

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/0x2E/fusion/internal/web"
	"github.com/gin-gonic/gin"
)

func (h *Handler) setupFrontendRoutes(r *gin.Engine) error {
	frontendFS, _, err := web.FrontendFS()
	if err != nil {
		return err
	}

	setupFrontendRoutesWithFS(r, frontendFS)

	return nil
}

func setupFrontendRoutesWithFS(r *gin.Engine, frontendFS fs.FS) {
	fileServer := http.FileServer(http.FS(frontendFS))
	r.GET("/assets/*filepath", func(c *gin.Context) {
		serveFrontendRoute(c, frontendFS, fileServer)
	})
	r.HEAD("/assets/*filepath", func(c *gin.Context) {
		serveFrontendRoute(c, frontendFS, fileServer)
	})
	r.NoRoute(func(c *gin.Context) {
		serveFrontendRoute(c, frontendFS, fileServer)
	})
}

func serveFrontendRoute(c *gin.Context, frontendFS fs.FS, fileServer http.Handler) {
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		c.Status(http.StatusNotFound)
		return
	}

	setFrontendSecurityHeaders(c)

	cleanedPath := path.Clean(c.Request.URL.Path)
	if cleanedPath == "." {
		cleanedPath = "/"
	}

	if isAPIPath(cleanedPath) {
		c.Status(http.StatusNotFound)
		return
	}

	if cleanedPath == "/" {
		serveFrontendIndex(c, fileServer)
		return
	}

	assetPath := strings.TrimPrefix(cleanedPath, "/")
	if assetPath == "" {
		assetPath = "index.html"
	}

	if frontendFileExists(frontendFS, assetPath) {
		serveFrontendRequestPath(c, fileServer, "/"+assetPath)
		return
	}

	if looksLikeAssetPath(assetPath) {
		c.Status(http.StatusNotFound)
		return
	}

	serveFrontendIndex(c, fileServer)
}

func serveFrontendIndex(c *gin.Context, fileServer http.Handler) {
	setFrontendCacheHeaders(c, "index.html")
	serveFrontendRequestPath(c, fileServer, "/")
}

func setFrontendSecurityHeaders(c *gin.Context) {
	c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' http: https: data:; font-src 'self' data:; connect-src 'self' http: https:; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("Referrer-Policy", "no-referrer")
}

func serveFrontendRequestPath(c *gin.Context, fileServer http.Handler, requestPath string) {
	setFrontendCacheHeaders(c, strings.TrimPrefix(requestPath, "/"))

	originalPath := c.Request.URL.Path
	c.Request.URL.Path = requestPath
	fileServer.ServeHTTP(c.Writer, c.Request)
	c.Request.URL.Path = originalPath
}

func setFrontendCacheHeaders(c *gin.Context, filePath string) {
	cleanedPath := path.Clean("/" + filePath)

	switch {
	case strings.HasPrefix(cleanedPath, "/assets/"):
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	case cleanedPath == "/index.html" || cleanedPath == "/manifest.json" || cleanedPath == "/sw.js":
		c.Header("Cache-Control", "no-cache")
	case isFrontendIconPath(cleanedPath):
		c.Header("Cache-Control", "public, max-age=86400")
	default:
		c.Header("Cache-Control", "no-cache")
	}
}

func isFrontendIconPath(filePath string) bool {
	base := path.Base(filePath)
	return base == "favicon.ico" ||
		base == "apple-touch-icon.png" ||
		(strings.HasPrefix(base, "icon-") && strings.HasSuffix(base, ".png"))
}

func frontendFileExists(frontendFS fs.FS, filePath string) bool {
	info, err := fs.Stat(frontendFS, filePath)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func isAPIPath(requestPath string) bool {
	cleanedPath := path.Clean(requestPath)
	if cleanedPath == "." {
		cleanedPath = "/"
	}

	return cleanedPath == "/api" || strings.HasPrefix(cleanedPath, "/api/")
}

func looksLikeAssetPath(filePath string) bool {
	base := path.Base(filePath)
	return strings.Contains(base, ".")
}

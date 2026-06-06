package handler

import (
	"bytes"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestLogMiddlewareDoesNotLogRawQuery(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{})))
	t.Cleanup(func() {
		slog.SetDefault(previous)
	})

	r := newTestRouter()
	r.Use(requestLogMiddleware())
	r.GET("/api/oidc/callback", func(c *gin.Context) {
		c.Status(http.StatusBadRequest)
	})

	performRequest(r, http.MethodGet, "/api/oidc/callback?code=secret-code&state=secret-state", nil, nil)

	output := logs.String()
	if strings.Contains(output, "secret-code") || strings.Contains(output, "secret-state") {
		t.Fatalf("log output contains sensitive query values: %s", output)
	}
	if !strings.Contains(output, "/api/oidc/callback") {
		t.Fatalf("log output should include request path, got: %s", output)
	}
}

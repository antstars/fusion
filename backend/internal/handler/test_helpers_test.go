package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func requirePostgresTestURL(t *testing.T) string {
	t.Helper()

	databaseURL := os.Getenv("FUSION_TEST_POSTGRES_URL")
	if databaseURL == "" {
		t.Skip("FUSION_TEST_POSTGRES_URL is not set")
	}
	return databaseURL
}

func resetPostgresTestDB(t *testing.T, databaseURL string) {
	t.Helper()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open postgres test database: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		TRUNCATE TABLE read_later_items, bookmarks, items, feed_fetch_state, feeds, groups, app_settings RESTART IDENTITY CASCADE;
		INSERT INTO groups (id, name) VALUES (1, 'Default');
		SELECT setval(pg_get_serial_sequence('groups', 'id'), 1);
		INSERT INTO app_settings (key, value) VALUES ('max_articles', '0'), ('retention_days', '30');
	`); err != nil {
		t.Fatalf("reset postgres test database: %v", err)
	}
}

func mustJSONBody(t *testing.T, payload any) *bytes.Reader {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	return bytes.NewReader(body)
}

func performRequest(
	r http.Handler,
	method, target string,
	body io.Reader,
	headers map[string]string,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, body)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

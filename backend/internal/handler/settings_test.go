package handler

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

func newRetentionSettingsTestHandler(t *testing.T) (*Handler, *store.Store) {
	t.Helper()

	databaseURL := requirePostgresTestURL(t)
	st, err := store.New(databaseURL)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	resetPostgresTestDB(t, databaseURL)

	h, err := New(st, &config.Config{
		Password:       "secret",
		PullTimeout:    30,
		LoginRateLimit: 10,
		LoginWindow:    60,
		LoginBlock:     300,
	}, noopPuller{})
	if err != nil {
		_ = st.Close()
		t.Fatalf("new handler: %v", err)
	}

	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})

	return h, st
}

func TestRetentionSettingsHandlers(t *testing.T) {
	h, _ := newRetentionSettingsTestHandler(t)

	r := newTestRouter()
	r.GET("/api/settings/retention", h.getRetentionSettings)
	r.PATCH("/api/settings/retention", h.updateRetentionSettings)

	get := performRequest(r, http.MethodGet, "/api/settings/retention", nil, nil)
	if get.Code != http.StatusOK {
		t.Fatalf("expected GET status 200, got %d", get.Code)
	}

	var getPayload struct {
		Data retentionSettingsResponse `json:"data"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &getPayload); err != nil {
		t.Fatalf("unmarshal GET response: %v", err)
	}
	if getPayload.Data.MaxArticles != 0 || getPayload.Data.RetentionDays != 30 {
		t.Fatalf("unexpected default settings: %+v", getPayload.Data)
	}

	patch := performRequest(
		r,
		http.MethodPatch,
		"/api/settings/retention",
		mustJSONBody(t, gin.H{"max_articles": 10, "retention_days": 90}),
		map[string]string{"Content-Type": "application/json"},
	)
	if patch.Code != http.StatusOK {
		t.Fatalf("expected PATCH status 200, got %d", patch.Code)
	}

	var patchPayload struct {
		Data retentionSettingsResponse `json:"data"`
	}
	if err := json.Unmarshal(patch.Body.Bytes(), &patchPayload); err != nil {
		t.Fatalf("unmarshal PATCH response: %v", err)
	}
	if patchPayload.Data.MaxArticles != 10 || patchPayload.Data.RetentionDays != 90 {
		t.Fatalf("unexpected patched settings: %+v", patchPayload.Data)
	}
}

func TestRetentionSettingsRouteRequiresAuth(t *testing.T) {
	h, _ := newRetentionSettingsTestHandler(t)
	r := h.SetupRouter()

	w := performRequest(r, http.MethodGet, "/api/settings/retention", nil, nil)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUpdateRetentionSettingsValidation(t *testing.T) {
	h, _ := newRetentionSettingsTestHandler(t)

	tests := []struct {
		name string
		body gin.H
	}{
		{name: "rejects negative max articles", body: gin.H{"max_articles": -1, "retention_days": 30}},
		{name: "rejects unsupported retention days", body: gin.H{"max_articles": 0, "retention_days": 7}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter()
			r.PATCH("/api/settings/retention", h.updateRetentionSettings)

			w := performRequest(
				r,
				http.MethodPatch,
				"/api/settings/retention",
				mustJSONBody(t, tt.body),
				map[string]string{"Content-Type": "application/json"},
			)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", w.Code)
			}
		})
	}
}

func TestUpdateRetentionSettingsRunsCleanup(t *testing.T) {
	h, st := newRetentionSettingsTestHandler(t)

	group, err := st.CreateGroup("Tech")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}
	feed, err := st.CreateFeed(group.ID, "Feed", "https://example.com/rss.xml", "", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}
	old := time.Now().AddDate(0, 0, -31).Unix()
	item, err := st.CreateItem(feed.ID, "old", "Old", "https://example.com/old", "", old)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	r := newTestRouter()
	r.PATCH("/api/settings/retention", h.updateRetentionSettings)
	w := performRequest(
		r,
		http.MethodPatch,
		"/api/settings/retention",
		mustJSONBody(t, gin.H{"max_articles": 0, "retention_days": 30}),
		map[string]string{"Content-Type": "application/json"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if _, err := st.GetItem(item.ID); err == nil {
		t.Fatal("expected old item to be deleted")
	}
}

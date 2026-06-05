package handler

import (
	"net/http"
	"testing"

	"github.com/0x2E/fusion/internal/model"
	"github.com/gin-gonic/gin"
)

func TestMarkItemsBatchValidation(t *testing.T) {
	ids := make([]int64, maxBatchUpdateIDs+1)
	for i := range ids {
		ids[i] = int64(i + 1)
	}

	tests := []struct {
		name    string
		path    string
		handler gin.HandlerFunc
		body    any
	}{
		{name: "read rejects too many ids", path: "/api/items/-/read", handler: (&Handler{}).markItemsRead, body: gin.H{"ids": ids}},
		{name: "unread rejects empty ids", path: "/api/items/-/unread", handler: (&Handler{}).markItemsUnread, body: gin.H{"ids": []int64{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter()
			r.PATCH(tt.path, tt.handler)

			w := performRequest(
				r,
				http.MethodPatch,
				tt.path,
				mustJSONBody(t, tt.body),
				map[string]string{"Content-Type": "application/json"},
			)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", w.Code)
			}
		})
	}
}

func TestListItemsRejectsInvalidOrderBy(t *testing.T) {
	r := newTestRouter()
	r.GET("/api/items", (&Handler{}).listItems)

	w := performRequest(r, http.MethodGet, "/api/items?order_by=pub_date:desc", nil, nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestListItemsRejectsOffset(t *testing.T) {
	r := newTestRouter()
	r.GET("/api/items", (&Handler{}).listItems)

	w := performRequest(r, http.MethodGet, "/api/items?offset=10", nil, nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestItemCursorRoundTrip(t *testing.T) {
	item := &model.Item{ID: 42, PubDate: 100, CreatedAt: 200}

	pubDateCursor := encodeItemCursor(item, "pub_date")
	decoded, err := decodeItemCursor(pubDateCursor, "pub_date")
	if err != nil {
		t.Fatalf("decode pub_date cursor failed: %v", err)
	}
	if decoded.Value != item.PubDate || decoded.ID != item.ID {
		t.Fatalf("decoded pub_date cursor = %+v, want value=%d id=%d", decoded, item.PubDate, item.ID)
	}

	createdAtCursor := encodeItemCursor(item, "created_at")
	decoded, err = decodeItemCursor(createdAtCursor, "created_at")
	if err != nil {
		t.Fatalf("decode created_at cursor failed: %v", err)
	}
	if decoded.Value != item.CreatedAt || decoded.ID != item.ID {
		t.Fatalf("decoded created_at cursor = %+v, want value=%d id=%d", decoded, item.CreatedAt, item.ID)
	}

	if _, err := decodeItemCursor(createdAtCursor, "pub_date"); err == nil {
		t.Fatal("expected cursor with mismatched order_by to be rejected")
	}
}

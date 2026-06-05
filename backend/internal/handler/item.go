package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/0x2E/fusion/internal/model"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

const maxListLimit = 100
const maxBatchUpdateIDs = 1000

type itemCursorPayload struct {
	OrderBy string `json:"order_by"`
	Value   int64  `json:"value"`
	ID      int64  `json:"id"`
}

type itemListResponse struct {
	Data       []*model.Item `json:"data"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

type markItemsReadRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

func (h *Handler) listItems(c *gin.Context) {
	params := store.ListItemsParams{}

	if feedID := c.Query("feed_id"); feedID != "" {
		id, err := strconv.ParseInt(feedID, 10, 64)
		if err != nil {
			badRequestError(c, "invalid feed_id")
			return
		}
		params.FeedID = &id
	}

	if groupID := c.Query("group_id"); groupID != "" {
		id, err := strconv.ParseInt(groupID, 10, 64)
		if err != nil {
			badRequestError(c, "invalid group_id")
			return
		}
		params.GroupID = &id
	}

	if unread := c.Query("unread"); unread != "" {
		val, err := strconv.ParseBool(unread)
		if err != nil {
			badRequestError(c, "invalid unread")
			return
		}
		params.Unread = &val
	}

	if c.Query("offset") != "" {
		badRequestError(c, "offset is not supported")
		return
	}

	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		val, err := strconv.Atoi(limitParam)
		if err != nil || val <= 0 {
			badRequestError(c, "invalid limit")
			return
		}
		if val > maxListLimit {
			val = maxListLimit
		}
		limit = val
	}
	params.Limit = limit + 1

	orderBy := c.Query("order_by")
	if orderBy == "" {
		orderBy = "pub_date"
	} else {
		if orderBy != "pub_date" && orderBy != "created_at" {
			badRequestError(c, "invalid order_by")
			return
		}
	}
	params.OrderBy = orderBy

	if cursor := c.Query("cursor"); cursor != "" {
		decodedCursor, err := decodeItemCursor(cursor, params.OrderBy)
		if err != nil {
			badRequestError(c, "invalid cursor")
			return
		}
		params.Cursor = decodedCursor
	}

	items, err := h.store.ListItems(params)
	if err != nil {
		internalError(c, err, "list items")
		return
	}

	nextCursor := ""
	if len(items) > limit {
		items = items[:limit]
		nextCursor = encodeItemCursor(items[len(items)-1], params.OrderBy)
	}

	c.JSON(http.StatusOK, itemListResponse{
		Data:       items,
		NextCursor: nextCursor,
	})
}

func decodeItemCursor(rawCursor, orderBy string) (*store.ListItemsCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(rawCursor)
	if err != nil {
		return nil, err
	}

	var payload itemCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, err
	}
	if payload.OrderBy != orderBy || payload.ID <= 0 {
		return nil, errors.New("cursor does not match request")
	}

	return &store.ListItemsCursor{Value: payload.Value, ID: payload.ID}, nil
}

func encodeItemCursor(item *model.Item, orderBy string) string {
	value := item.PubDate
	if orderBy == "created_at" {
		value = item.CreatedAt
	}
	payload := itemCursorPayload{
		OrderBy: orderBy,
		Value:   value,
		ID:      item.ID,
	}
	encoded, _ := json.Marshal(payload)
	return base64.RawURLEncoding.EncodeToString(encoded)
}

func (h *Handler) getItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	item, err := h.store.GetItem(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "item")
			return
		}
		internalError(c, err, "get item")
		return
	}

	dataResponse(c, item)
}

func (h *Handler) markItemsRead(c *gin.Context) {
	var req markItemsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}
	if len(req.IDs) == 0 || len(req.IDs) > maxBatchUpdateIDs {
		badRequestError(c, "invalid ids")
		return
	}

	if err := h.store.BatchUpdateItemsUnread(req.IDs, false); err != nil {
		internalError(c, err, "mark items as read")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) markItemsUnread(c *gin.Context) {
	var req markItemsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}
	if len(req.IDs) == 0 || len(req.IDs) > maxBatchUpdateIDs {
		badRequestError(c, "invalid ids")
		return
	}

	if err := h.store.BatchUpdateItemsUnread(req.IDs, true); err != nil {
		internalError(c, err, "mark items as unread")
		return
	}

	c.Status(http.StatusNoContent)
}

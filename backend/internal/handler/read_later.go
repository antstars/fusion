package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

type createReadLaterItemRequest struct {
	ItemID   *int64 `json:"item_id"`
	Link     string `json:"link"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	PubDate  int64  `json:"pub_date"`
	FeedName string `json:"feed_name"`
}

func (h *Handler) listReadLaterItems(c *gin.Context) {
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		val, err := strconv.Atoi(limitStr)
		if err != nil || val <= 0 {
			badRequestError(c, "invalid limit")
			return
		}
		if val > maxListLimit {
			val = maxListLimit
		}
		limit = val
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		val, err := strconv.Atoi(offsetStr)
		if err != nil || val < 0 {
			badRequestError(c, "invalid offset")
			return
		}
		offset = val
	}

	items, err := h.store.ListReadLaterItems(limit, offset)
	if err != nil {
		internalError(c, err, "list read later items")
		return
	}

	total, err := h.store.CountReadLaterItems()
	if err != nil {
		internalError(c, err, "count read later items")
		return
	}

	listResponse(c, items, total)
}

func (h *Handler) getReadLaterItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	item, err := h.store.GetReadLaterItem(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "read later item")
			return
		}
		internalError(c, err, "get read later item")
		return
	}

	dataResponse(c, item)
}

func (h *Handler) createReadLaterItem(c *gin.Context) {
	var req createReadLaterItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	var link, title, content, feedName string
	var pubDate int64

	if req.ItemID != nil {
		item, err := h.store.GetItem(*req.ItemID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				notFoundError(c, "item")
				return
			}
			internalError(c, err, "get item for read later")
			return
		}

		feed, err := h.store.GetFeed(item.FeedID)
		if err != nil {
			internalError(c, err, "get feed for read later")
			return
		}

		link = item.Link
		title = item.Title
		content = item.Content
		pubDate = item.PubDate
		feedName = feed.Name
	} else {
		if req.Link == "" || req.Title == "" || req.Content == "" || req.FeedName == "" {
			badRequestError(c, "missing required fields")
			return
		}
		link = req.Link
		title = req.Title
		content = req.Content
		pubDate = req.PubDate
		feedName = req.FeedName
	}

	item, err := h.store.CreateReadLaterItem(req.ItemID, link, title, content, pubDate, feedName)
	if err != nil {
		internalError(c, err, "create read later item")
		return
	}

	dataResponse(c, item)
}

func (h *Handler) deleteReadLaterItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	if err := h.store.DeleteReadLaterItem(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "read later item")
			return
		}
		internalError(c, err, "delete read later item")
		return
	}

	c.Status(http.StatusNoContent)
}

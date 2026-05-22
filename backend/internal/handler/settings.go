package handler

import (
	"net/http"
	"time"

	"github.com/0x2E/fusion/internal/model"
	"github.com/gin-gonic/gin"
)

type updateRetentionSettingsRequest struct {
	MaxArticles   int `json:"max_articles"`
	RetentionDays int `json:"retention_days"`
}

type retentionSettingsResponse struct {
	MaxArticles   int `json:"max_articles"`
	RetentionDays int `json:"retention_days"`
	Deleted       int `json:"deleted,omitempty"`
}

func (h *Handler) getRetentionSettings(c *gin.Context) {
	settings, err := h.store.GetRetentionSettings()
	if err != nil {
		internalError(c, err, "get retention settings")
		return
	}

	dataResponse(c, retentionSettingsResponse{
		MaxArticles:   settings.MaxArticles,
		RetentionDays: settings.RetentionDays,
	})
}

func (h *Handler) updateRetentionSettings(c *gin.Context) {
	var req updateRetentionSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	if req.MaxArticles < 0 || !isValidRetentionDays(req.RetentionDays) {
		badRequestError(c, "invalid retention settings")
		return
	}

	settings, err := h.store.UpdateRetentionSettings(model.RetentionSettings{
		MaxArticles:   req.MaxArticles,
		RetentionDays: req.RetentionDays,
	})
	if err != nil {
		internalError(c, err, "update retention settings")
		return
	}

	deleted, err := h.store.CleanupItemsByRetention(*settings, time.Now())
	if err != nil {
		internalError(c, err, "cleanup items by retention")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": retentionSettingsResponse{
			MaxArticles:   settings.MaxArticles,
			RetentionDays: settings.RetentionDays,
			Deleted:       deleted,
		},
	})
}

func isValidRetentionDays(days int) bool {
	switch days {
	case 0, 30, 90, 365:
		return true
	default:
		return false
	}
}

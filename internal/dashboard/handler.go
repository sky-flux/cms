package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetStats returns aggregated dashboard statistics for the current site.
func (h *Handler) GetStats(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	if siteSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "site_slug is required"})
		return
	}
	schemaName := "site_" + siteSlug

	stats, err := h.svc.GetStats(c.Request.Context(), schemaName)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, stats)
}

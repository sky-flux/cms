package post

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// ListTranslations returns all translations for a post.
func (h *Handler) ListTranslations(c *gin.Context) {
	ts, err := h.svc.ListTranslations(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]TranslationListItem, len(ts))
	for i := range ts {
		items[i] = ToTranslationListItem(&ts[i])
	}
	response.Success(c, items)
}

// GetTranslation returns a specific locale translation.
func (h *Handler) GetTranslation(c *gin.Context) {
	t, err := h.svc.GetTranslation(c.Request.Context(), c.Param("id"), c.Param("locale"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToTranslationResp(t))
}

// UpsertTranslation creates or updates a translation.
func (h *Handler) UpsertTranslation(c *gin.Context) {
	var req UpsertTranslationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	t, err := h.svc.UpsertTranslation(c.Request.Context(), c.Param("id"), c.Param("locale"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToTranslationResp(t))
}

// DeleteTranslation removes a specific locale translation.
func (h *Handler) DeleteTranslation(c *gin.Context) {
	if err := h.svc.DeleteTranslation(c.Request.Context(), c.Param("id"), c.Param("locale")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "translation deleted"})
}

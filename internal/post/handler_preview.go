package post

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// CreatePreviewToken generates a new preview token for a post.
func (h *Handler) CreatePreviewToken(c *gin.Context) {
	userID := c.GetString("user_id")
	resp, err := h.svc.CreatePreviewToken(c.Request.Context(), c.Param("id"), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, resp)
}

// ListPreviewTokens returns active preview tokens for a post.
func (h *Handler) ListPreviewTokens(c *gin.Context) {
	tokens, err := h.svc.ListPreviewTokens(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, tokens)
}

// RevokeAllPreviewTokens deletes all preview tokens for a post.
func (h *Handler) RevokeAllPreviewTokens(c *gin.Context) {
	count, err := h.svc.RevokeAllPreviewTokens(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"revoked": count})
}

// RevokePreviewToken deletes a single preview token.
func (h *Handler) RevokePreviewToken(c *gin.Context) {
	if err := h.svc.RevokePreviewToken(c.Request.Context(), c.Param("token_id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "preview token revoked"})
}

package post

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Publish transitions a post to published status.
func (h *Handler) Publish(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	post, err := h.svc.Publish(c.Request.Context(), siteSlug, c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToPostResp(post))
}

// Unpublish transitions a published post to archived.
func (h *Handler) Unpublish(c *gin.Context) {
	post, err := h.svc.Unpublish(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToPostResp(post))
}

// RevertToDraft transitions a post back to draft.
func (h *Handler) RevertToDraft(c *gin.Context) {
	post, err := h.svc.RevertToDraft(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToPostResp(post))
}

// Restore restores a soft-deleted post.
func (h *Handler) Restore(c *gin.Context) {
	post, err := h.svc.RestorePost(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToPostResp(post))
}

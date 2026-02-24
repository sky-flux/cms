package post

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// ListRevisions returns the revision history for a post.
func (h *Handler) ListRevisions(c *gin.Context) {
	revs, err := h.svc.ListRevisions(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]RevisionResp, len(revs))
	for i := range revs {
		items[i] = ToRevisionResp(&revs[i])
	}
	response.Success(c, items)
}

// Rollback restores a post to a specific revision's content.
func (h *Handler) Rollback(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	userID := c.GetString("user_id")

	post, err := h.svc.Rollback(c.Request.Context(), siteSlug, userID, c.Param("id"), c.Param("rev_id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToPostResp(post))
}

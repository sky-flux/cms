package post

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// ListPosts returns a paginated list of posts.
func (h *Handler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	f := ListFilter{
		Page:           page,
		PerPage:        perPage,
		Status:         c.Query("status"),
		Query:          c.Query("q"),
		CategoryID:     c.Query("category_id"),
		TagID:          c.Query("tag_id"),
		AuthorID:       c.Query("author_id"),
		Sort:           c.Query("sort"),
		IncludeDeleted: c.Query("include_deleted") == "true",
	}

	posts, total, err := h.svc.ListPosts(c.Request.Context(), f)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]PostListItem, len(posts))
	for i := range posts {
		items[i] = ToPostListItem(&posts[i])
	}
	response.Paginated(c, items, total, page, perPage)
}

// CreatePost creates a new post.
func (h *Handler) CreatePost(c *gin.Context) {
	var req CreatePostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	siteSlug := c.GetString("site_slug")
	userID := c.GetString("user_id")

	post, err := h.svc.CreatePost(c.Request.Context(), siteSlug, userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, ToPostResp(post))
}

// GetPost returns a single post by ID.
func (h *Handler) GetPost(c *gin.Context) {
	post, err := h.svc.GetPost(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToPostResp(post))
}

// UpdatePost updates a post with optimistic locking.
func (h *Handler) UpdatePost(c *gin.Context) {
	var req UpdatePostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	siteSlug := c.GetString("site_slug")
	userID := c.GetString("user_id")

	post, err := h.svc.UpdatePost(c.Request.Context(), siteSlug, userID, c.Param("id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToPostResp(post))
}

// DeletePost soft-deletes a post.
func (h *Handler) DeletePost(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	if err := h.svc.DeletePost(c.Request.Context(), siteSlug, c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "post deleted"})
}

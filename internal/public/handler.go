package public

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler handles HTTP requests for the public headless API.
type Handler struct {
	svc *Service
}

// NewHandler creates a new public API handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListPosts handles GET /posts — paginated list of published posts.
func (h *Handler) ListPosts(c *gin.Context) {
	siteSlug := c.GetString("site_slug")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filter := PostListFilter{
		Page:     page,
		PerPage:  perPage,
		Category: c.Query("category"),
		Tag:      c.Query("tag"),
		Locale:   c.Query("locale"),
		Sort:     c.Query("sort"),
	}

	items, total, err := h.svc.ListPosts(c.Request.Context(), siteSlug, filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, items, total, page, perPage)
}

// GetPost handles GET /posts/:slug — single published post.
func (h *Handler) GetPost(c *gin.Context) {
	siteSlug := c.GetString("site_slug")

	detail, err := h.svc.GetPost(c.Request.Context(), siteSlug, c.Param("slug"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, detail)
}

// ListCategories handles GET /categories — category tree with post counts.
func (h *Handler) ListCategories(c *gin.Context) {
	siteSlug := c.GetString("site_slug")

	nodes, err := h.svc.ListCategories(c.Request.Context(), siteSlug)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nodes)
}

// ListTags handles GET /tags — tags with post counts.
func (h *Handler) ListTags(c *gin.Context) {
	siteSlug := c.GetString("site_slug")

	sortBy := c.DefaultQuery("sort", "name:asc")

	items, err := h.svc.ListTags(c.Request.Context(), siteSlug, sortBy)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

// Search handles GET /search — full-text search via Meilisearch.
func (h *Handler) Search(c *gin.Context) {
	siteSlug := c.GetString("site_slug")

	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	items, total, err := h.svc.Search(c.Request.Context(), siteSlug, query, page, perPage)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, items, total, page, perPage)
}

// ListComments handles GET /posts/:slug/comments — paginated comment tree.
func (h *Handler) ListComments(c *gin.Context) {
	postSlug := c.Param("slug")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	result, err := h.svc.ListComments(c.Request.Context(), postSlug, page, perPage)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, gin.H{
		"comment_count": result.CommentCount,
		"comments":      result.Comments,
	}, result.Total, result.Page, result.PerPage)
}

// CreateComment handles POST /posts/:slug/comments — submit a public comment.
func (h *Handler) CreateComment(c *gin.Context) {
	postSlug := c.Param("slug")

	var req CreateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request body", err))
		return
	}

	// Extract optional authenticated user context (may be empty for guests).
	userID := ""
	if v, exists := c.Get("user_id"); exists {
		userID, _ = v.(string)
	}
	userName := ""
	if v, exists := c.Get("user_name"); exists {
		userName, _ = v.(string)
	}
	userEmail := ""
	if v, exists := c.Get("user_email"); exists {
		userEmail, _ = v.(string)
	}

	result, err := h.svc.CreateComment(
		c.Request.Context(),
		postSlug,
		&req,
		userID,
		userName,
		userEmail,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

// GetMenu handles GET /menus — public menu by location or slug.
func (h *Handler) GetMenu(c *gin.Context) {
	location := c.Query("location")
	slug := c.Query("slug")

	menu, err := h.svc.GetMenu(c.Request.Context(), location, slug)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, menu)
}

// Preview handles GET /preview/:token — consume a preview token and return post.
func (h *Handler) Preview(c *gin.Context) {
	rawToken := c.Param("token")

	result, err := h.svc.Preview(c.Request.Context(), rawToken)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

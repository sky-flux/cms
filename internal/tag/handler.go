package tag

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler handles HTTP requests for tags.
type Handler struct {
	svc *Service
}

// NewHandler creates a new tag handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /tags — returns a paginated list of tags.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	filter := ListFilter{
		Page:    page,
		PerPage: perPage,
		Query:   c.Query("q"),
		Sort:    c.Query("sort"),
	}

	tags, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, tags, total, page, perPage)
}

// Get handles GET /tags/:id — returns a single tag.
func (h *Handler) Get(c *gin.Context) {
	tag, err := h.svc.GetTag(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, tag)
}

// Suggest handles GET /tags/suggest — autocomplete search via Meilisearch.
func (h *Handler) Suggest(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	query := c.Query("q")

	tags, err := h.svc.Suggest(c.Request.Context(), siteSlug, query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, tags)
}

// Create handles POST /tags — creates a new tag.
func (h *Handler) Create(c *gin.Context) {
	var req CreateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	siteSlug := c.GetString("site_slug")
	tag, err := h.svc.CreateTag(c.Request.Context(), siteSlug, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, tag)
}

// Update handles PUT /tags/:id — updates an existing tag.
func (h *Handler) Update(c *gin.Context) {
	var req UpdateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	siteSlug := c.GetString("site_slug")
	tag, err := h.svc.UpdateTag(c.Request.Context(), siteSlug, c.Param("id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, tag)
}

// Delete handles DELETE /tags/:id — deletes a tag.
func (h *Handler) Delete(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	if err := h.svc.DeleteTag(c.Request.Context(), siteSlug, c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "tag deleted"})
}

package redirect

import (
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler handles HTTP requests for redirects.
type Handler struct {
	svc *Service
}

// NewHandler creates a new redirect handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /redirects — paginated list of redirects.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	statusCode, _ := strconv.Atoi(c.Query("status_code"))

	filter := ListFilter{
		Page:       page,
		PerPage:    perPage,
		Query:      c.Query("q"),
		StatusCode: statusCode,
		Status:     c.Query("status"),
		Sort:       c.Query("sort"),
	}

	results, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, results, total, page, perPage)
}

// Create handles POST /redirects — creates a new redirect.
func (h *Handler) Create(c *gin.Context) {
	var req CreateRedirectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	userID := c.GetString("user_id")
	siteSlug := c.GetString("site_slug")

	rd, err := h.svc.Create(c.Request.Context(), siteSlug, &req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, rd)
}

// Update handles PUT /redirects/:id — updates a redirect.
func (h *Handler) Update(c *gin.Context) {
	var req UpdateRedirectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	siteSlug := c.GetString("site_slug")
	rd, err := h.svc.Update(c.Request.Context(), siteSlug, c.Param("id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, rd)
}

// Delete handles DELETE /redirects/:id — deletes a redirect.
func (h *Handler) Delete(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	if err := h.svc.Delete(c.Request.Context(), siteSlug, c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "redirect deleted"})
}

// BatchDelete handles DELETE /redirects/batch — bulk delete redirects.
func (h *Handler) BatchDelete(c *gin.Context) {
	var req BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	siteSlug := c.GetString("site_slug")
	count, err := h.svc.BatchDelete(c.Request.Context(), siteSlug, req.IDs)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"deleted_count": count})
}

// Import handles POST /redirects/import — CSV import.
func (h *Handler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, apperror.Validation("file is required", err))
		return
	}

	// 1MB limit.
	if file.Size > 1<<20 {
		response.Error(c, apperror.Validation("file too large (max 1MB)", nil))
		return
	}

	f, err := file.Open()
	if err != nil {
		response.Error(c, apperror.Validation("cannot open file", err))
		return
	}
	defer f.Close()

	userID := c.GetString("user_id")
	siteSlug := c.GetString("site_slug")

	result, err := h.svc.Import(c.Request.Context(), siteSlug, f, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// Export handles GET /redirects/export — CSV export.
func (h *Handler) Export(c *gin.Context) {
	redirects, err := h.svc.Export(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="redirects.csv"`)

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"source_path", "target_url", "status_code"})
	for _, rd := range redirects {
		_ = w.Write([]string{rd.SourcePath, rd.TargetURL, fmt.Sprintf("%d", rd.StatusCode)})
	}
	w.Flush()
}

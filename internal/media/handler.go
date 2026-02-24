package media

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler exposes HTTP endpoints for the media module.
type Handler struct {
	svc *Service
}

// NewHandler creates a new media handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /media — returns a paginated list of media files.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filter := ListFilter{
		Page:    page,
		PerPage: perPage,
		Query:   c.Query("q"),
	}

	if mt := c.Query("media_type"); mt != "" {
		v, err := strconv.Atoi(mt)
		if err == nil {
			filter.MediaType = &v
		}
	}

	items, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, items, total, page, perPage)
}

// Get handles GET /media/:id — returns a single media file.
func (h *Handler) Get(c *gin.Context) {
	resp, err := h.svc.GetMedia(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// Upload handles POST /media — uploads a new media file.
func (h *Handler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, apperror.Validation("file is required", err))
		return
	}
	defer file.Close()

	uploaderID, _ := c.Get("user_id")
	siteSlug, _ := c.Get("site_slug")

	altText := c.PostForm("alt_text")

	resp, err := h.svc.Upload(
		c.Request.Context(),
		siteSlug.(string),
		uploaderID.(string),
		file,
		header.Filename,
		header.Header.Get("Content-Type"),
		header.Size,
		altText,
	)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, resp)
}

// Update handles PUT /media/:id — updates media metadata (alt_text).
func (h *Handler) Update(c *gin.Context) {
	var req UpdateMediaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	resp, err := h.svc.UpdateMedia(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// Delete handles DELETE /media/:id — soft-deletes a media file.
func (h *Handler) Delete(c *gin.Context) {
	force := c.Query("force") == "true"
	if err := h.svc.DeleteMedia(c.Request.Context(), c.Param("id"), force); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "media file deleted"})
}

// BatchDelete handles DELETE /media/batch — soft-deletes multiple media files.
func (h *Handler) BatchDelete(c *gin.Context) {
	var req BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	force := c.Query("force") == "true"
	resp, err := h.svc.BatchDeleteMedia(c.Request.Context(), req.IDs, force)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

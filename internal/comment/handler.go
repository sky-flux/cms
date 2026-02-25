package comment

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler handles HTTP requests for comments.
type Handler struct {
	svc *Service
}

// NewHandler creates a new comment handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /comments — paginated list.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	filter := ListFilter{
		Page:    page,
		PerPage: perPage,
		PostID:  c.Query("post_id"),
		Status:  c.Query("status"),
		Query:   c.Query("q"),
	}

	results, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, results, total, page, perPage)
}

// Get handles GET /comments/:id — comment detail with replies.
func (h *Handler) Get(c *gin.Context) {
	result, err := h.svc.GetComment(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// UpdateStatus handles PUT /comments/:id/status.
func (h *Handler) UpdateStatus(c *gin.Context) {
	var req UpdateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), c.Param("id"), &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "status updated"})
}

// TogglePin handles PUT /comments/:id/pin.
func (h *Handler) TogglePin(c *gin.Context) {
	var req TogglePinReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.TogglePin(c.Request.Context(), c.Param("id"), &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "pin toggled"})
}

// Reply handles POST /comments/:id/reply — admin reply.
func (h *Handler) Reply(c *gin.Context) {
	var req ReplyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	userEmail := c.GetString("user_email")

	result, err := h.svc.Reply(c.Request.Context(), c.Param("id"), &req, userID, userName, userEmail)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

// BatchStatus handles PUT /comments/batch-status.
func (h *Handler) BatchStatus(c *gin.Context) {
	var req BatchStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	affected, err := h.svc.BatchUpdateStatus(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"updated_count": affected})
}

// Delete handles DELETE /comments/:id — hard delete.
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.DeleteComment(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "comment deleted"})
}

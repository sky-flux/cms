package menu

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler handles HTTP requests for menus.
type Handler struct {
	svc *Service
}

// NewHandler creates a new menu handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListMenus handles GET /menus.
func (h *Handler) ListMenus(c *gin.Context) {
	location := c.Query("location")
	menus, err := h.svc.ListMenus(c.Request.Context(), location)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, menus)
}

// GetMenu handles GET /menus/:id.
func (h *Handler) GetMenu(c *gin.Context) {
	result, err := h.svc.GetMenu(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// CreateMenu handles POST /menus.
func (h *Handler) CreateMenu(c *gin.Context) {
	var req CreateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	result, err := h.svc.CreateMenu(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

// UpdateMenu handles PUT /menus/:id.
func (h *Handler) UpdateMenu(c *gin.Context) {
	var req UpdateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	result, err := h.svc.UpdateMenu(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// DeleteMenu handles DELETE /menus/:id.
func (h *Handler) DeleteMenu(c *gin.Context) {
	if err := h.svc.DeleteMenu(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "menu deleted"})
}

// AddItem handles POST /menus/:id/items.
func (h *Handler) AddItem(c *gin.Context) {
	var req CreateMenuItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	result, err := h.svc.AddItem(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

// UpdateItem handles PUT /menus/:id/items/:item_id.
func (h *Handler) UpdateItem(c *gin.Context) {
	var req UpdateMenuItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	result, err := h.svc.UpdateItem(c.Request.Context(), c.Param("id"), c.Param("item_id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// DeleteItem handles DELETE /menus/:id/items/:item_id.
func (h *Handler) DeleteItem(c *gin.Context) {
	if err := h.svc.DeleteItem(c.Request.Context(), c.Param("id"), c.Param("item_id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "menu item deleted"})
}

// ReorderItems handles PUT /menus/:id/items/reorder.
func (h *Handler) ReorderItems(c *gin.Context) {
	var req ReorderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.ReorderItems(c.Request.Context(), c.Param("id"), &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "items reordered"})
}

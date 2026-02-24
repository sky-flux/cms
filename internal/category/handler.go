package category

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler exposes category endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new category handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List returns all categories as a tree.
func (h *Handler) List(c *gin.Context) {
	tree, err := h.svc.ListTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, tree)
}

// Get returns a single category by ID.
func (h *Handler) Get(c *gin.Context) {
	cat, err := h.svc.GetCategory(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, cat)
}

// Create creates a new category.
func (h *Handler) Create(c *gin.Context) {
	var req CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	cat, err := h.svc.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, cat)
}

// Update updates an existing category.
func (h *Handler) Update(c *gin.Context) {
	var req UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	cat, err := h.svc.UpdateCategory(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, cat)
}

// Delete removes a category.
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.DeleteCategory(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "category deleted"})
}

// Reorder batch-updates category sort orders.
func (h *Handler) Reorder(c *gin.Context) {
	var req ReorderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.Reorder(c.Request.Context(), req.Orders); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "categories reordered"})
}

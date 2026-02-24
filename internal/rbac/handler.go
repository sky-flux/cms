package rbac

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler provides HTTP endpoints for RBAC management.
type Handler struct {
	svc          *Service
	roleRepo     RoleRepository
	apiRepo      APIRepository
	roleAPIRepo  RoleAPIRepository
	menuRepo     MenuRepository
	templateRepo TemplateRepository
	userRoleRepo UserRoleRepository
}

// NewHandler creates an RBAC handler with all required dependencies.
func NewHandler(
	svc *Service,
	roleRepo RoleRepository,
	apiRepo APIRepository,
	roleAPIRepo RoleAPIRepository,
	menuRepo MenuRepository,
	templateRepo TemplateRepository,
	userRoleRepo UserRoleRepository,
) *Handler {
	return &Handler{
		svc:          svc,
		roleRepo:     roleRepo,
		apiRepo:      apiRepo,
		roleAPIRepo:  roleAPIRepo,
		menuRepo:     menuRepo,
		templateRepo: templateRepo,
		userRoleRepo: userRoleRepo,
	}
}

// --- Role CRUD ---

func (h *Handler) GetRole(c *gin.Context) {
	role, err := h.roleRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, role)
}

func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.roleRepo.List(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, roles)
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	role := &model.Role{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	}

	if err := h.roleRepo.Create(c.Request.Context(), role); err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, role)
}

func (h *Handler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	role, err := h.roleRepo.GetByID(ctx, id)
	if err != nil {
		response.Error(c, err)
		return
	}

	if role.BuiltIn == model.ToggleYes && role.Slug == "super" {
		response.Error(c, apperror.Forbidden("super role cannot be modified", nil))
		return
	}

	var req UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Status != nil {
		role.Status = *req.Status
	}

	if err := h.roleRepo.Update(ctx, role); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, role)
}

func (h *Handler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	role, err := h.roleRepo.GetByID(ctx, id)
	if err != nil {
		response.Error(c, err)
		return
	}

	if role.BuiltIn == model.ToggleYes {
		response.Error(c, apperror.Forbidden("built-in roles cannot be deleted", nil))
		return
	}

	if err := h.roleRepo.Delete(ctx, id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// --- Role-API Permissions ---

func (h *Handler) GetRoleAPIs(c *gin.Context) {
	apis, err := h.roleAPIRepo.GetAPIsByRoleID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, apis)
}

func (h *Handler) SetRoleAPIs(c *gin.Context) {
	roleID := c.Param("id")
	ctx := c.Request.Context()

	var req SetRoleAPIsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.roleAPIRepo.SetRoleAPIs(ctx, roleID, req.APIIDs); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.InvalidateRoleCache(ctx, roleID); err != nil {
		slog.Error("invalidate role cache", "error", err, "role_id", roleID)
	}

	response.NoContent(c)
}

// --- Role-Menu Permissions ---

func (h *Handler) GetRoleMenus(c *gin.Context) {
	menus, err := h.menuRepo.GetMenusByRoleID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, menus)
}

func (h *Handler) SetRoleMenus(c *gin.Context) {
	roleID := c.Param("id")
	ctx := c.Request.Context()

	var req SetRoleMenusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.menuRepo.SetRoleMenus(ctx, roleID, req.MenuIDs); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.InvalidateRoleCache(ctx, roleID); err != nil {
		slog.Error("invalidate role cache", "error", err, "role_id", roleID)
	}

	response.NoContent(c)
}

// --- Menu CRUD ---

func (h *Handler) ListMenus(c *gin.Context) {
	menus, err := h.menuRepo.ListTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, menus)
}

func (h *Handler) CreateMenu(c *gin.Context) {
	var req CreateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	menu := &model.AdminMenu{
		ParentID:  req.ParentID,
		Name:      req.Name,
		Icon:      req.Icon,
		Path:      req.Path,
		SortOrder: req.SortOrder,
	}

	if err := h.menuRepo.Create(c.Request.Context(), menu); err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, menu)
}

func (h *Handler) UpdateMenu(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req UpdateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	menu := &model.AdminMenu{ID: id}
	if req.Name != "" {
		menu.Name = req.Name
	}
	if req.Icon != nil {
		menu.Icon = *req.Icon
	}
	if req.Path != nil {
		menu.Path = *req.Path
	}
	if req.SortOrder != nil {
		menu.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		menu.Status = *req.Status
	}
	if req.ParentID != nil {
		menu.ParentID = req.ParentID
	}

	if err := h.menuRepo.Update(ctx, menu); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, menu)
}

func (h *Handler) DeleteMenu(c *gin.Context) {
	if err := h.menuRepo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// --- Templates ---

func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.templateRepo.List(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, templates)
}

func (h *Handler) GetTemplate(c *gin.Context) {
	tmpl, err := h.templateRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, tmpl)
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	tmpl := &model.RoleTemplate{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.templateRepo.Create(c.Request.Context(), tmpl); err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, tmpl)
}

func (h *Handler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	tmpl, err := h.templateRepo.GetByID(ctx, id)
	if err != nil {
		response.Error(c, err)
		return
	}

	var req UpdateTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if req.Name != "" {
		tmpl.Name = req.Name
	}
	if req.Description != "" {
		tmpl.Description = req.Description
	}

	if err := h.templateRepo.Update(ctx, tmpl); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, tmpl)
}

func (h *Handler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	tmpl, err := h.templateRepo.GetByID(ctx, id)
	if err != nil {
		response.Error(c, err)
		return
	}

	if tmpl.BuiltIn == model.ToggleYes {
		response.Error(c, apperror.Forbidden("built-in templates cannot be deleted", nil))
		return
	}

	if err := h.templateRepo.Delete(ctx, id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// --- Apply Template to Role ---

func (h *Handler) ApplyTemplate(c *gin.Context) {
	roleID := c.Param("id")
	ctx := c.Request.Context()

	role, err := h.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		response.Error(c, err)
		return
	}

	if role.BuiltIn == model.ToggleYes && role.Slug == "super" {
		response.Error(c, apperror.Forbidden("cannot modify super role permissions", nil))
		return
	}

	var req ApplyTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if _, err := h.templateRepo.GetByID(ctx, req.TemplateID); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.roleAPIRepo.CloneFromTemplate(ctx, roleID, req.TemplateID); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.InvalidateRoleCache(ctx, roleID); err != nil {
		slog.Error("invalidate role cache after template apply", "error", err, "role_id", roleID)
	}

	response.NoContent(c)
}

// --- User Roles ---

func (h *Handler) GetUserRoles(c *gin.Context) {
	roles, err := h.userRoleRepo.GetRolesByUserID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, roles)
}

func (h *Handler) SetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	ctx := c.Request.Context()

	var req SetUserRolesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.userRoleRepo.SetUserRoles(ctx, userID, req.RoleIDs); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.InvalidateUserCache(ctx, userID); err != nil {
		slog.Error("invalidate user cache", "error", err, "user_id", userID)
	}

	response.NoContent(c)
}

// --- Current User Menu ---

func (h *Handler) GetMyMenus(c *gin.Context) {
	userID := c.GetString("user_id")
	menus, err := h.svc.GetUserMenuTree(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, menus)
}

// --- API Registry ---

func (h *Handler) ListAPIs(c *gin.Context) {
	apis, err := h.apiRepo.List(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, apis)
}

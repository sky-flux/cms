package site

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListSites(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	f := ListFilter{
		Page:    page,
		PerPage: perPage,
		Query:   c.Query("q"),
	}
	if v := c.Query("is_active"); v != "" {
		active := v == "true"
		f.IsActive = &active
	}

	sites, total, err := h.svc.ListSites(c.Request.Context(), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, ToSiteRespList(sites), total, page, perPage)
}

func (h *Handler) CreateSite(c *gin.Context) {
	var req CreateSiteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	site, err := h.svc.CreateSite(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	resp := ToSiteResp(site)
	response.Created(c, resp)
}

func (h *Handler) GetSite(c *gin.Context) {
	site, err := h.svc.GetSite(c.Request.Context(), c.Param("slug"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToSiteResp(site))
}

func (h *Handler) UpdateSite(c *gin.Context) {
	var req UpdateSiteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	site, err := h.svc.UpdateSite(c.Request.Context(), c.Param("slug"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToSiteResp(site))
}

func (h *Handler) DeleteSite(c *gin.Context) {
	var req DeleteSiteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.DeleteSite(c.Request.Context(), c.Param("slug"), req.ConfirmSlug); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "site deleted"})
}

func (h *Handler) ListSiteUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	f := UserFilter{
		Page:    page,
		PerPage: perPage,
		Role:    c.Query("role"),
		Query:   c.Query("q"),
	}

	users, total, err := h.svc.ListSiteUsers(c.Request.Context(), c.Param("slug"), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, ToSiteUserRespList(users), total, page, perPage)
}

func (h *Handler) AssignSiteRole(c *gin.Context) {
	var req AssignSiteRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	slug := c.Param("slug")
	userID := c.Param("user_id")

	if err := h.svc.AssignSiteRole(c.Request.Context(), slug, userID, req.Role); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"user_id": userID, "site_slug": slug, "role": req.Role})
}

func (h *Handler) RemoveSiteRole(c *gin.Context) {
	slug := c.Param("slug")
	userID := c.Param("user_id")

	if err := h.svc.RemoveSiteRole(c.Request.Context(), slug, userID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "user role removed"})
}

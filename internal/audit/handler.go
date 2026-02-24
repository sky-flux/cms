package audit

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

type Handler struct {
	repo AuditRepository
}

func NewHandler(repo AuditRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	f := ListFilter{
		Page:         page,
		PerPage:      perPage,
		ActorID:      c.Query("actor_id"),
		ResourceType: c.Query("resource_type"),
	}

	if v := c.Query("action"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 11 {
			response.Error(c, apperror.Validation("invalid action value", nil))
			return
		}
		action := model.LogAction(n)
		f.Action = &action
	}

	if v := c.Query("start_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, apperror.Validation("invalid start_date format, use ISO 8601", nil))
			return
		}
		f.StartDate = &t
	}

	if v := c.Query("end_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, apperror.Validation("invalid end_date format, use ISO 8601", nil))
			return
		}
		f.EndDate = &t
	}

	items, total, err := h.repo.List(c.Request.Context(), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, ToAuditRespList(items), total, page, perPage)
}

package system

import (
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

func (h *Handler) ListSettings(c *gin.Context) {
	configs, err := h.svc.ListConfigs(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToConfigRespList(configs))
}

func (h *Handler) UpdateSetting(c *gin.Context) {
	var req UpdateConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	key := c.Param("key")
	userID, _ := c.Get("user_id")

	cfg, err := h.svc.UpdateConfig(c.Request.Context(), key, &req, userID.(string))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToConfigResp(cfg))
}

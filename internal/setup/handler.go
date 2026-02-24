package setup

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

func (h *Handler) Check(c *gin.Context) {
	installed := h.svc.Check(c.Request.Context())
	response.Success(c, CheckResp{Installed: installed})
}

func (h *Handler) Initialize(c *gin.Context) {
	var req InitializeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}
	resp, err := h.svc.Initialize(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, resp)
}

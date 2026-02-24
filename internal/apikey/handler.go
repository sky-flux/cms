package apikey

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

func (h *Handler) ListAPIKeys(c *gin.Context) {
	keys, err := h.svc.ListAPIKeys(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToAPIKeyRespList(keys))
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	ownerID, _ := c.Get("user_id")
	key, plainKey, err := h.svc.CreateAPIKey(c.Request.Context(), ownerID.(string), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	resp := CreateAPIKeyResp{
		APIKeyResp: ToAPIKeyResp(key),
		PlainKey:   plainKey,
	}
	response.Created(c, resp)
}

func (h *Handler) RevokeAPIKey(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.RevokeAPIKey(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "api key revoked"})
}

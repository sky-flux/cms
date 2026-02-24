package apikey

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

type CreateAPIKeyReq struct {
	Name      string     `json:"name" binding:"required,max=100"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// --- Response DTOs ---

type APIKeyResp struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	KeyPrefix  string             `json:"key_prefix"`
	Status     model.APIKeyStatus `json:"status"`
	LastUsedAt *time.Time         `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time         `json:"expires_at,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
	RevokedAt  *time.Time         `json:"revoked_at,omitempty"`
}

type CreateAPIKeyResp struct {
	APIKeyResp
	PlainKey string `json:"plain_key"`
}

func ToAPIKeyResp(k *model.APIKey) APIKeyResp {
	return APIKeyResp{
		ID:         k.ID,
		Name:       k.Name,
		KeyPrefix:  k.KeyPrefix,
		Status:     k.Status,
		LastUsedAt: k.LastUsedAt,
		ExpiresAt:  k.ExpiresAt,
		CreatedAt:  k.CreatedAt,
		RevokedAt:  k.RevokedAt,
	}
}

func ToAPIKeyRespList(keys []model.APIKey) []APIKeyResp {
	out := make([]APIKeyResp, len(keys))
	for i := range keys {
		out[i] = ToAPIKeyResp(&keys[i])
	}
	return out
}

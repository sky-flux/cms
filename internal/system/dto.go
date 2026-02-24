package system

import (
	"encoding/json"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

type UpdateConfigReq struct {
	Value json.RawMessage `json:"value" binding:"required"`
}

// --- Response DTOs ---

type ConfigResp struct {
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	Description string          `json:"description,omitempty"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func ToConfigResp(c *model.SiteConfig) ConfigResp {
	return ConfigResp{
		Key:         c.Key,
		Value:       c.Value,
		Description: c.Description,
		UpdatedAt:   c.UpdatedAt,
	}
}

func ToConfigRespList(configs []model.SiteConfig) []ConfigResp {
	out := make([]ConfigResp, len(configs))
	for i := range configs {
		out[i] = ToConfigResp(&configs[i])
	}
	return out
}

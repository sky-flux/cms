package model

import (
	"context"
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type SiteConfig struct {
	bun.BaseModel `bun:"table:sfc_site_configs,alias:scfg"`

	Key         string          `bun:"key,pk" json:"key"`
	Value       json.RawMessage `bun:"value,type:jsonb" json:"value"`
	Description string          `bun:"description" json:"description,omitempty"`
	UpdatedBy   *string         `bun:"updated_by,type:uuid" json:"updated_by,omitempty"`
	UpdatedAt   time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (sc *SiteConfig) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetUpdatedAt(&sc.UpdatedAt, query)
	return nil
}

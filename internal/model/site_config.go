package model

import (
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

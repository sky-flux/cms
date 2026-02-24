package model

import (
	"context"
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type Config struct {
	bun.BaseModel `bun:"table:sfc_configs,alias:sc"`

	Key         string          `bun:"key,pk" json:"key"`
	Value       json.RawMessage `bun:"value,type:jsonb" json:"value"`
	Description string          `bun:"description" json:"description,omitempty"`
	UpdatedBy   *string         `bun:"updated_by,type:uuid" json:"updated_by,omitempty"`
	UpdatedAt   time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (c *Config) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetUpdatedAt(&c.UpdatedAt, query)
	return nil
}

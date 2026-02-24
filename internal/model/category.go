package model

import (
	"context"
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type Category struct {
	bun.BaseModel `bun:"table:sfc_site_categories,alias:c"`

	ID          string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	ParentID    *string         `bun:"parent_id,type:uuid" json:"parent_id,omitempty"`
	Name        string          `bun:"name,notnull" json:"name"`
	Slug        string          `bun:"slug,notnull" json:"slug"`
	Path        string          `bun:"path,notnull,default:'/'" json:"path"`
	Description string          `bun:"description" json:"description,omitempty"`
	SortOrder   int             `bun:"sort_order,notnull,default:0" json:"sort_order"`
	Meta        json.RawMessage `bun:"meta,type:jsonb,default:'{}'" json:"meta,omitempty"`
	CreatedAt   time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	Children []*Category `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (c *Category) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&c.CreatedAt, &c.UpdatedAt, query)
	return nil
}

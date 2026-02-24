package model

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type PostType struct {
	bun.BaseModel `bun:"table:sfc_site_post_types,alias:pty"`

	ID          string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Name        string          `bun:"name,notnull,unique" json:"name"`
	Slug        string          `bun:"slug,notnull,unique" json:"slug"`
	Description string          `bun:"description" json:"description,omitempty"`
	Fields      json.RawMessage `bun:"fields,type:jsonb,default:'[]'" json:"fields"`
	BuiltIn     bool            `bun:"built_in,notnull,default:false" json:"built_in"`
	CreatedAt   time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

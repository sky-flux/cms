package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Role struct {
	bun.BaseModel `bun:"table:sfc_roles,alias:r"`

	ID          string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Name        string    `bun:"name,notnull,unique" json:"name"`
	Slug        string    `bun:"slug,notnull,unique" json:"slug"`
	Description string    `bun:"description" json:"description,omitempty"`
	BuiltIn     Toggle     `bun:"built_in,notnull,type:smallint,default:1" json:"built_in"`
	Status      RoleStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (r *Role) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&r.CreatedAt, &r.UpdatedAt, query)
	return nil
}

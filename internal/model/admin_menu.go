package model

import (
	"time"

	"github.com/uptrace/bun"
)

// AdminMenu represents a backend management sidebar menu item (public.sfc_menus).
// Distinct from sfc_site_menus which is the frontend navigation menu in site schemas.
type AdminMenu struct {
	bun.BaseModel `bun:"table:sfc_menus,alias:m"`

	ID        string       `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	ParentID  *string      `bun:"parent_id,type:uuid" json:"parent_id,omitempty"`
	Name      string       `bun:"name,notnull" json:"name"`
	Icon      string       `bun:"icon" json:"icon,omitempty"`
	Path      string       `bun:"path" json:"path,omitempty"`
	SortOrder int          `bun:"sort_order,notnull,default:0" json:"sort_order"`
	Status    bool         `bun:"status,notnull,default:true" json:"status"`
	CreatedAt time.Time    `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time    `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	Children  []*AdminMenu `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
}

type RoleMenu struct {
	bun.BaseModel `bun:"table:sfc_role_menus"`

	RoleID string `bun:"role_id,pk,type:uuid" json:"role_id"`
	MenuID string `bun:"menu_id,pk,type:uuid" json:"menu_id"`
}

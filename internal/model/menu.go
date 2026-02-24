package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// SiteMenu maps to sfc_site_menus (frontend navigation menu).
// Distinct from AdminMenu (public.sfc_menus) which is the backend sidebar menu.
type SiteMenu struct {
	bun.BaseModel `bun:"table:sfc_site_menus,alias:sm"`

	ID        string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Name      string    `bun:"name,notnull" json:"name"`
	Slug      string    `bun:"slug,notnull,unique" json:"slug"`
	Location  string    `bun:"location" json:"location,omitempty"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	Items []*SiteMenuItem `bun:"rel:has-many,join:id=menu_id" json:"items,omitempty"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (sm *SiteMenu) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&sm.CreatedAt, &sm.UpdatedAt, query)
	return nil
}

// SiteMenuItem maps to sfc_site_menu_items.
type SiteMenuItem struct {
	bun.BaseModel `bun:"table:sfc_site_menu_items,alias:mi"`

	ID          string       `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	MenuID      string       `bun:"menu_id,notnull,type:uuid" json:"menu_id"`
	ParentID    *string      `bun:"parent_id,type:uuid" json:"parent_id,omitempty"`
	Label       string       `bun:"label,notnull" json:"label"`
	URL         string       `bun:"url" json:"url,omitempty"`
	Target      string       `bun:"target,notnull,default:'_self'" json:"target"`
	Type        MenuItemType `bun:"type,notnull,type:smallint,default:1" json:"type"`
	ReferenceID *string      `bun:"reference_id,type:uuid" json:"reference_id,omitempty"`
	SortOrder   int          `bun:"sort_order,notnull,default:0" json:"sort_order"`
	IsActive    bool         `bun:"is_active,notnull,default:true" json:"is_active"`
	CreatedAt   time.Time    `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time    `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	Children []*SiteMenuItem `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (mi *SiteMenuItem) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&mi.CreatedAt, &mi.UpdatedAt, query)
	return nil
}

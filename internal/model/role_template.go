package model

import (
	"time"

	"github.com/uptrace/bun"
)

type RoleTemplate struct {
	bun.BaseModel `bun:"table:sfc_role_templates,alias:rtpl"`

	ID          string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Name        string    `bun:"name,notnull,unique" json:"name"`
	Description string    `bun:"description" json:"description,omitempty"`
	BuiltIn     bool      `bun:"built_in,notnull,default:false" json:"built_in"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type RoleTemplateAPI struct {
	bun.BaseModel `bun:"table:sfc_role_template_apis"`

	TemplateID string `bun:"template_id,pk,type:uuid" json:"template_id"`
	APIID      string `bun:"api_id,pk,type:uuid" json:"api_id"`
}

type RoleTemplateMenu struct {
	bun.BaseModel `bun:"table:sfc_role_template_menus"`

	TemplateID string `bun:"template_id,pk,type:uuid" json:"template_id"`
	MenuID     string `bun:"menu_id,pk,type:uuid" json:"menu_id"`
}

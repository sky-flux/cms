package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type APIEndpoint struct {
	bun.BaseModel `bun:"table:sfc_apis,alias:api"`

	ID          string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Method      string    `bun:"method,notnull" json:"method"`
	Path        string    `bun:"path,notnull" json:"path"`
	Name        string    `bun:"name,notnull" json:"name"`
	Description string    `bun:"description" json:"description,omitempty"`
	Group       string    `bun:"group,notnull" json:"group"`
	Status      APIStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (ae *APIEndpoint) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&ae.CreatedAt, &ae.UpdatedAt, query)
	return nil
}

type RoleAPI struct {
	bun.BaseModel `bun:"table:sfc_role_apis"`

	RoleID string `bun:"role_id,pk,type:uuid" json:"role_id"`
	APIID  string `bun:"api_id,pk,type:uuid" json:"api_id"`
}

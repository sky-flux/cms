package model

import (
	"time"

	"github.com/uptrace/bun"
)

type UserRole struct {
	bun.BaseModel `bun:"table:sfc_user_roles"`

	UserID    string    `bun:"user_id,pk,type:uuid" json:"user_id"`
	RoleID    string    `bun:"role_id,pk,type:uuid" json:"role_id"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`

	// Relations
	Role *Role `bun:"rel:belongs-to,join:role_id=id" json:"role,omitempty"`
	User *User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}

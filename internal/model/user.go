package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:sfc_users,alias:u"`

	ID           string     `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Email        string     `bun:"email,notnull,unique" json:"email"`
	PasswordHash string     `bun:"password_hash,notnull" json:"-"`
	DisplayName  string     `bun:"display_name,notnull" json:"display_name"`
	AvatarURL    string     `bun:"avatar_url" json:"avatar_url,omitempty"`
	IsActive     bool       `bun:"is_active,notnull,default:true" json:"is_active"`
	LastLoginAt  *time.Time `bun:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt    *time.Time `bun:"deleted_at,soft_delete,nullzero" json:"-"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (u *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&u.CreatedAt, &u.UpdatedAt, query)
	NormalizeEmail(&u.Email)
	return nil
}

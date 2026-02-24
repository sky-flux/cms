package model

import (
	"time"

	"github.com/uptrace/bun"
)

type PasswordResetToken struct {
	bun.BaseModel `bun:"table:sfc_password_reset_tokens,alias:prt"`

	ID        string     `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	UserID    string     `bun:"user_id,notnull,type:uuid" json:"user_id"`
	TokenHash string     `bun:"token_hash,notnull,unique" json:"-"`
	ExpiresAt time.Time  `bun:"expires_at,notnull" json:"expires_at"`
	UsedAt    *time.Time `bun:"used_at" json:"used_at,omitempty"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

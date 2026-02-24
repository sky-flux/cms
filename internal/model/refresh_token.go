package model

import (
	"time"

	"github.com/uptrace/bun"
)

type RefreshToken struct {
	bun.BaseModel `bun:"table:sfc_refresh_tokens,alias:rt"`

	ID        string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	UserID    string    `bun:"user_id,notnull,type:uuid" json:"user_id"`
	TokenHash string    `bun:"token_hash,notnull,unique" json:"-"`
	ExpiresAt time.Time `bun:"expires_at,notnull" json:"expires_at"`
	Revoked   bool      `bun:"revoked,notnull,default:false" json:"revoked"`
	IPAddress string    `bun:"ip_address,type:inet" json:"ip_address,omitempty"`
	UserAgent string    `bun:"user_agent" json:"user_agent,omitempty"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

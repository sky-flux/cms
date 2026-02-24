package model

import (
	"time"

	"github.com/uptrace/bun"
)

type APIKey struct {
	bun.BaseModel `bun:"table:sfc_site_api_keys,alias:ak"`

	ID         string     `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	OwnerID    string     `bun:"owner_id,notnull,type:uuid" json:"owner_id"`
	Name       string     `bun:"name,notnull" json:"name"`
	KeyHash    string     `bun:"key_hash,notnull,unique" json:"-"`
	KeyPrefix  string     `bun:"key_prefix,notnull" json:"key_prefix"`
	IsActive   bool       `bun:"is_active,notnull,default:true" json:"is_active"`
	LastUsedAt *time.Time `bun:"last_used_at" json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `bun:"expires_at" json:"expires_at,omitempty"`
	RateLimit  int        `bun:"rate_limit,notnull,default:100" json:"rate_limit"`
	CreatedAt  time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	RevokedAt  *time.Time `bun:"revoked_at" json:"revoked_at,omitempty"`
}

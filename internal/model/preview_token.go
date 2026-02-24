package model

import (
	"time"

	"github.com/uptrace/bun"
)

type PreviewToken struct {
	bun.BaseModel `bun:"table:sfc_site_preview_tokens,alias:pvt"`

	ID        string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	PostID    string    `bun:"post_id,notnull,type:uuid" json:"post_id"`
	TokenHash string    `bun:"token_hash,notnull,unique" json:"-"`
	ExpiresAt time.Time `bun:"expires_at,notnull" json:"expires_at"`
	CreatedBy *string   `bun:"created_by,type:uuid" json:"created_by,omitempty"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

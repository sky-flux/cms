package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Redirect struct {
	bun.BaseModel `bun:"table:sfc_site_redirects,alias:rd"`

	ID         string     `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	SourcePath string     `bun:"source_path,notnull,unique" json:"source_path"`
	TargetURL  string     `bun:"target_url,notnull" json:"target_url"`
	StatusCode int        `bun:"status_code,notnull,default:301" json:"status_code"`
	IsActive   bool       `bun:"is_active,notnull,default:true" json:"is_active"`
	HitCount   int64      `bun:"hit_count,notnull,default:0" json:"hit_count"`
	LastHitAt  *time.Time `bun:"last_hit_at" json:"last_hit_at,omitempty"`
	CreatedBy  *string    `bun:"created_by,type:uuid" json:"created_by,omitempty"`
	CreatedAt  time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

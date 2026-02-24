package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Tag struct {
	bun.BaseModel `bun:"table:sfc_site_tags,alias:t"`

	ID        string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Name      string    `bun:"name,notnull,unique" json:"name"`
	Slug      string    `bun:"slug,notnull,unique" json:"slug"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

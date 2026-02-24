package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Site struct {
	bun.BaseModel `bun:"table:sfc_sites,alias:s"`

	ID            string     `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Name          string     `bun:"name,notnull" json:"name"`
	Slug          string     `bun:"slug,notnull,unique" json:"slug"`
	Domain        string     `bun:"domain,unique,nullzero" json:"domain,omitempty"`
	Description   string     `bun:"description" json:"description,omitempty"`
	LogoURL       string     `bun:"logo_url" json:"logo_url,omitempty"`
	DefaultLocale string     `bun:"default_locale,notnull,default:'zh-CN'" json:"default_locale"`
	Timezone      string     `bun:"timezone,notnull,default:'Asia/Shanghai'" json:"timezone"`
	Status        SiteStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
	Settings      string     `bun:"settings,type:jsonb,default:'{}'" json:"settings"`
	CreatedAt     time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt     *time.Time `bun:"deleted_at,soft_delete,nullzero" json:"-"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (s *Site) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&s.CreatedAt, &s.UpdatedAt, query)
	return nil
}

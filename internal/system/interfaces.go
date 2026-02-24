package system

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// ConfigRepository handles sfc_site_configs table CRUD.
type ConfigRepository interface {
	List(ctx context.Context) ([]model.SiteConfig, error)
	GetByKey(ctx context.Context, key string) (*model.SiteConfig, error)
	Upsert(ctx context.Context, cfg *model.SiteConfig) error
}

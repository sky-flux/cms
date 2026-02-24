package system

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

type ConfigRepo struct {
	db *bun.DB
}

func NewConfigRepo(db *bun.DB) *ConfigRepo {
	return &ConfigRepo{db: db}
}

func (r *ConfigRepo) List(ctx context.Context) ([]model.SiteConfig, error) {
	var configs []model.SiteConfig
	err := r.db.NewSelect().Model(&configs).OrderExpr("key ASC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("config list: %w", err)
	}
	return configs, nil
}

func (r *ConfigRepo) GetByKey(ctx context.Context, key string) (*model.SiteConfig, error) {
	cfg := new(model.SiteConfig)
	err := r.db.NewSelect().Model(cfg).Where("key = ?", key).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("config not found", err)
		}
		return nil, fmt.Errorf("config get by key: %w", err)
	}
	return cfg, nil
}

func (r *ConfigRepo) Upsert(ctx context.Context, cfg *model.SiteConfig) error {
	_, err := r.db.NewInsert().Model(cfg).
		On("CONFLICT (key) DO UPDATE").
		Set("value = EXCLUDED.value").
		Set("description = EXCLUDED.description").
		Set("updated_by = EXCLUDED.updated_by").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("config upsert: %w", err)
	}
	return nil
}

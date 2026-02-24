package setup

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

type configRepo struct {
	db *bun.DB
}

func NewConfigRepo(db *bun.DB) ConfigRepository {
	return &configRepo{db: db}
}

func (r *configRepo) GetValue(ctx context.Context, key string) (json.RawMessage, error) {
	var cfg model.Config
	err := r.db.NewSelect().Model(&cfg).Where("key = ?", key).Scan(ctx)
	if err != nil {
		return nil, apperror.NotFound("config key not found", err)
	}
	return cfg.Value, nil
}

func (r *configRepo) SetValue(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal config value: %w", err)
	}
	cfg := &model.Config{
		Key:   key,
		Value: data,
	}
	_, err = r.db.NewInsert().Model(cfg).
		On("CONFLICT (key) DO UPDATE").
		Set("value = EXCLUDED.value").
		Set("updated_at = NOW()").
		Exec(ctx)
	return err
}

type siteRepo struct {
	db *bun.DB
}

func NewSiteRepo(db *bun.DB) SiteRepository {
	return &siteRepo{db: db}
}

func (r *siteRepo) Create(ctx context.Context, site *model.Site) error {
	_, err := r.db.NewInsert().Model(site).Exec(ctx)
	return err
}

type userRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

type userRoleRepo struct {
	db *bun.DB
}

func NewUserRoleRepo(db *bun.DB) UserRoleRepository {
	return &userRoleRepo{db: db}
}

func (r *userRoleRepo) AssignRole(ctx context.Context, userID, roleSlug string) error {
	var role model.Role
	err := r.db.NewSelect().Model(&role).Where("slug = ?", roleSlug).Scan(ctx)
	if err != nil {
		return apperror.NotFound("role not found: "+roleSlug, err)
	}
	ur := &model.UserRole{UserID: userID, RoleID: role.ID}
	_, err = r.db.NewInsert().Model(ur).
		On("CONFLICT DO NOTHING").
		Exec(ctx)
	return err
}

type roleRepo struct {
	db *bun.DB
}

func NewRoleRepo(db *bun.DB) RoleRepository {
	return &roleRepo{db: db}
}

func (r *roleRepo) GetBySlug(ctx context.Context, slug string) (*model.Role, error) {
	var role model.Role
	err := r.db.NewSelect().Model(&role).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		return nil, apperror.NotFound("role not found", err)
	}
	return &role, nil
}

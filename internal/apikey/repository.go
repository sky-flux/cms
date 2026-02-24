package apikey

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

type Repo struct {
	db *bun.DB
}

func NewRepo(db *bun.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) List(ctx context.Context) ([]model.APIKey, error) {
	var keys []model.APIKey
	err := r.db.NewSelect().
		Model(&keys).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("apikey list: %w", err)
	}
	return keys, nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (*model.APIKey, error) {
	key := new(model.APIKey)
	err := r.db.NewSelect().Model(key).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("api key not found", err)
		}
		return nil, fmt.Errorf("apikey get by id: %w", err)
	}
	return key, nil
}

func (r *Repo) Create(ctx context.Context, key *model.APIKey) error {
	_, err := r.db.NewInsert().Model(key).Exec(ctx)
	if err != nil {
		return fmt.Errorf("apikey create: %w", err)
	}
	return nil
}

func (r *Repo) Revoke(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*model.APIKey)(nil)).
		Set("status = ?", model.APIKeyStatusRevoked).
		Set("revoked_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("apikey revoke: %w", err)
	}
	return nil
}

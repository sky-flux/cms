package posttype

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func (r *Repo) List(ctx context.Context) ([]model.PostType, error) {
	var pts []model.PostType
	err := r.db.NewSelect().Model(&pts).
		OrderExpr("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("posttype list: %w", err)
	}
	return pts, nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (*model.PostType, error) {
	pt := new(model.PostType)
	err := r.db.NewSelect().Model(pt).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("post type not found", err)
		}
		return nil, fmt.Errorf("posttype get by id: %w", err)
	}
	return pt, nil
}

func (r *Repo) GetBySlug(ctx context.Context, slug string) (*model.PostType, error) {
	pt := new(model.PostType)
	err := r.db.NewSelect().Model(pt).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("post type not found", err)
		}
		return nil, fmt.Errorf("posttype get by slug: %w", err)
	}
	return pt, nil
}

func (r *Repo) Create(ctx context.Context, pt *model.PostType) error {
	_, err := r.db.NewInsert().Model(pt).Exec(ctx)
	if err != nil {
		return fmt.Errorf("posttype create: %w", err)
	}
	return nil
}

func (r *Repo) Update(ctx context.Context, pt *model.PostType) error {
	_, err := r.db.NewUpdate().Model(pt).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("posttype update: %w", err)
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.PostType)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("posttype delete: %w", err)
	}
	return nil
}

package post

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

type PreviewRepo struct {
	db *bun.DB
}

func NewPreviewRepo(db *bun.DB) *PreviewRepo {
	return &PreviewRepo{db: db}
}

func (r *PreviewRepo) List(ctx context.Context, postID string) ([]model.PreviewToken, error) {
	var tokens []model.PreviewToken
	err := r.db.NewSelect().
		Model(&tokens).
		Where("post_id = ?", postID).
		Where("expires_at > ?", time.Now()).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("preview list: %w", err)
	}
	return tokens, nil
}

func (r *PreviewRepo) Create(ctx context.Context, token *model.PreviewToken) error {
	_, err := r.db.NewInsert().Model(token).Exec(ctx)
	if err != nil {
		return fmt.Errorf("preview create: %w", err)
	}
	return nil
}

func (r *PreviewRepo) CountActive(ctx context.Context, postID string) (int, error) {
	count, err := r.db.NewSelect().
		Model((*model.PreviewToken)(nil)).
		Where("post_id = ?", postID).
		Where("expires_at > ?", time.Now()).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("preview count active: %w", err)
	}
	return count, nil
}

func (r *PreviewRepo) DeleteAll(ctx context.Context, postID string) (int64, error) {
	res, err := r.db.NewDelete().
		Model((*model.PreviewToken)(nil)).
		Where("post_id = ?", postID).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("preview delete all: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *PreviewRepo) DeleteByID(ctx context.Context, id string) error {
	res, err := r.db.NewDelete().
		Model((*model.PreviewToken)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("preview delete by id: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.NotFound("preview token not found", nil)
	}
	return nil
}

func (r *PreviewRepo) GetByHash(ctx context.Context, hash string) (*model.PreviewToken, error) {
	token := new(model.PreviewToken)
	err := r.db.NewSelect().
		Model(token).
		Where("token_hash = ?", hash).
		Where("expires_at > ?", time.Now()).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("preview token not found or expired", err)
		}
		return nil, fmt.Errorf("preview get by hash: %w", err)
	}
	return token, nil
}

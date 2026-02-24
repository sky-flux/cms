package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

type TranslationRepo struct {
	db *bun.DB
}

func NewTranslationRepo(db *bun.DB) *TranslationRepo {
	return &TranslationRepo{db: db}
}

func (r *TranslationRepo) List(ctx context.Context, postID string) ([]model.PostTranslation, error) {
	var ts []model.PostTranslation
	err := r.db.NewSelect().
		Model(&ts).
		Where("post_id = ?", postID).
		OrderExpr("locale ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("translation list: %w", err)
	}
	return ts, nil
}

func (r *TranslationRepo) Get(ctx context.Context, postID, locale string) (*model.PostTranslation, error) {
	t := new(model.PostTranslation)
	err := r.db.NewSelect().
		Model(t).
		Where("post_id = ?", postID).
		Where("locale = ?", locale).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("translation not found", err)
		}
		return nil, fmt.Errorf("translation get: %w", err)
	}
	return t, nil
}

func (r *TranslationRepo) Upsert(ctx context.Context, t *model.PostTranslation) error {
	_, err := r.db.NewInsert().
		Model(t).
		On("CONFLICT (post_id, locale) DO UPDATE").
		Set("title = EXCLUDED.title").
		Set("excerpt = EXCLUDED.excerpt").
		Set("content = EXCLUDED.content").
		Set("content_json = EXCLUDED.content_json").
		Set("meta_title = EXCLUDED.meta_title").
		Set("meta_description = EXCLUDED.meta_description").
		Set("og_image_url = EXCLUDED.og_image_url").
		Set("updated_at = NOW()").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("translation upsert: %w", err)
	}
	return nil
}

func (r *TranslationRepo) Delete(ctx context.Context, postID, locale string) error {
	res, err := r.db.NewDelete().
		Model((*model.PostTranslation)(nil)).
		Where("post_id = ?", postID).
		Where("locale = ?", locale).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("translation delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.NotFound("translation not found", nil)
	}
	return nil
}

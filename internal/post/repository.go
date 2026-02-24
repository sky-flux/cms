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

type PostRepo struct {
	db *bun.DB
}

func NewPostRepo(db *bun.DB) *PostRepo {
	return &PostRepo{db: db}
}

func (r *PostRepo) List(ctx context.Context, f ListFilter) ([]model.Post, int64, error) {
	var posts []model.Post
	q := r.db.NewSelect().
		Model(&posts).
		Relation("Author")

	if f.Status != "" {
		if s, ok := parseStatus(f.Status); ok {
			q = q.Where("p.status = ?", s)
		}
	}

	if f.AuthorID != "" {
		q = q.Where("p.author_id = ?", f.AuthorID)
	}

	if f.CategoryID != "" {
		q = q.Where("p.id IN (SELECT post_id FROM sfc_site_post_category_map WHERE category_id = ?)", f.CategoryID)
	}

	if f.TagID != "" {
		q = q.Where("p.id IN (SELECT post_id FROM sfc_site_post_tag_map WHERE tag_id = ?)", f.TagID)
	}

	if !f.IncludeDeleted {
		q = q.Where("p.deleted_at IS NULL")
	}

	// Sort
	switch f.Sort {
	case "published_at:desc":
		q = q.OrderExpr("p.published_at DESC NULLS LAST")
	case "title:asc":
		q = q.OrderExpr("p.title ASC")
	default:
		q = q.OrderExpr("p.created_at DESC")
	}

	count, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("post list count: %w", err)
	}

	perPage := f.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	page := f.Page
	if page <= 0 {
		page = 1
	}

	err = q.Limit(perPage).Offset((page - 1) * perPage).Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("post list: %w", err)
	}

	return posts, int64(count), nil
}

func (r *PostRepo) GetByID(ctx context.Context, id string) (*model.Post, error) {
	post := new(model.Post)
	err := r.db.NewSelect().
		Model(post).
		Relation("Author").
		Where("p.id = ?", id).
		Where("p.deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("post not found", err)
		}
		return nil, fmt.Errorf("post get by id: %w", err)
	}
	return post, nil
}

func (r *PostRepo) GetByIDUnscoped(ctx context.Context, id string) (*model.Post, error) {
	post := new(model.Post)
	err := r.db.NewSelect().
		Model(post).
		Relation("Author").
		Where("p.id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("post not found", err)
		}
		return nil, fmt.Errorf("post get by id unscoped: %w", err)
	}
	return post, nil
}

func (r *PostRepo) Create(ctx context.Context, post *model.Post) error {
	_, err := r.db.NewInsert().Model(post).Exec(ctx)
	if err != nil {
		return fmt.Errorf("post create: %w", err)
	}
	return nil
}

func (r *PostRepo) Update(ctx context.Context, post *model.Post, expectedVersion int) error {
	res, err := r.db.NewUpdate().
		Model(post).
		WherePK().
		Where("version = ?", expectedVersion).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("post update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.VersionConflict("post has been modified by another user", nil)
	}
	return nil
}

func (r *PostRepo) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.NewDelete().
		Model((*model.Post)(nil)).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("post soft delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.NotFound("post not found", nil)
	}
	return nil
}

func (r *PostRepo) Restore(ctx context.Context, id string) error {
	res, err := r.db.NewUpdate().
		Model((*model.Post)(nil)).
		Set("deleted_at = NULL").
		Set("status = ?", model.PostStatusDraft).
		Where("id = ?", id).
		Where("deleted_at IS NOT NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("post restore: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.NotFound("post not found in trash", nil)
	}
	return nil
}

func (r *PostRepo) SlugExists(ctx context.Context, slug, excludeID string) (bool, error) {
	q := r.db.NewSelect().
		Model((*model.Post)(nil)).
		Where("slug = ?", slug).
		Where("deleted_at IS NULL")
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("post slug exists: %w", err)
	}
	return exists, nil
}

func (r *PostRepo) UpdateStatus(ctx context.Context, id string, status model.PostStatus) error {
	q := r.db.NewUpdate().
		Model((*model.Post)(nil)).
		Set("status = ?", status).
		Where("id = ?", id).
		Where("deleted_at IS NULL")

	if status == model.PostStatusPublished {
		q = q.Set("published_at = NOW()")
	}

	res, err := q.Exec(ctx)
	if err != nil {
		return fmt.Errorf("post update status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.NotFound("post not found", nil)
	}
	return nil
}

func (r *PostRepo) SyncCategories(ctx context.Context, postID string, categoryIDs []string, primaryID string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete existing
		_, err := tx.NewDelete().
			Model((*model.PostCategoryMap)(nil)).
			Where("post_id = ?", postID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("delete post categories: %w", err)
		}

		if len(categoryIDs) == 0 {
			return nil
		}

		// Insert new — use model.Toggle for Primary field
		maps := make([]model.PostCategoryMap, len(categoryIDs))
		for i, cid := range categoryIDs {
			primary := model.ToggleNo
			if cid == primaryID {
				primary = model.ToggleYes
			}
			maps[i] = model.PostCategoryMap{
				PostID:     postID,
				CategoryID: cid,
				Primary:    primary,
			}
		}
		_, err = tx.NewInsert().Model(&maps).Exec(ctx)
		if err != nil {
			return fmt.Errorf("insert post categories: %w", err)
		}
		return nil
	})
}

func (r *PostRepo) SyncTags(ctx context.Context, postID string, tagIDs []string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewDelete().
			Model((*model.PostTagMap)(nil)).
			Where("post_id = ?", postID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("delete post tags: %w", err)
		}

		if len(tagIDs) == 0 {
			return nil
		}

		maps := make([]model.PostTagMap, len(tagIDs))
		for i, tid := range tagIDs {
			maps[i] = model.PostTagMap{PostID: postID, TagID: tid}
		}
		_, err = tx.NewInsert().Model(&maps).Exec(ctx)
		if err != nil {
			return fmt.Errorf("insert post tags: %w", err)
		}
		return nil
	})
}

func (r *PostRepo) LoadRelations(ctx context.Context, post *model.Post) error {
	return nil
}

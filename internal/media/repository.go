package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// Repo implements MediaRepository using uptrace/bun.
type Repo struct {
	db *bun.DB
}

// NewRepo creates a new media repository.
func NewRepo(db *bun.DB) *Repo {
	return &Repo{db: db}
}

// List returns a paginated, optionally filtered list of media files.
func (r *Repo) List(ctx context.Context, filter ListFilter) ([]model.MediaFile, int64, error) {
	var files []model.MediaFile

	q := r.db.NewSelect().Model(&files)

	if filter.Query != "" {
		q = q.Where("original_name ILIKE ?", "%"+filter.Query+"%")
	}
	if filter.MediaType != nil {
		q = q.Where("media_type = ?", *filter.MediaType)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("media list count: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage
	err = q.OrderExpr("created_at DESC").
		Limit(filter.PerPage).
		Offset(offset).
		Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("media list: %w", err)
	}

	return files, int64(total), nil
}

// GetByID retrieves a single media file by its ID.
func (r *Repo) GetByID(ctx context.Context, id string) (*model.MediaFile, error) {
	mf := new(model.MediaFile)
	err := r.db.NewSelect().Model(mf).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("media file not found", err)
		}
		return nil, fmt.Errorf("media get by id: %w", err)
	}
	return mf, nil
}

// Create inserts a new media file record.
func (r *Repo) Create(ctx context.Context, mf *model.MediaFile) error {
	_, err := r.db.NewInsert().Model(mf).Exec(ctx)
	if err != nil {
		return fmt.Errorf("media create: %w", err)
	}
	return nil
}

// Update modifies an existing media file record.
func (r *Repo) Update(ctx context.Context, mf *model.MediaFile) error {
	_, err := r.db.NewUpdate().Model(mf).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("media update: %w", err)
	}
	return nil
}

// SoftDelete performs a soft delete on a single media file.
func (r *Repo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.MediaFile)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("media soft delete: %w", err)
	}
	return nil
}

// BatchSoftDelete performs a soft delete on multiple media files and returns the count.
func (r *Repo) BatchSoftDelete(ctx context.Context, ids []string) (int64, error) {
	res, err := r.db.NewDelete().Model((*model.MediaFile)(nil)).
		Where("id IN (?)", bun.In(ids)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("media batch soft delete: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// GetReferencingPosts returns posts that reference a media file via cover_image_id.
func (r *Repo) GetReferencingPosts(ctx context.Context, mediaID string) ([]PostRef, error) {
	var refs []PostRef
	err := r.db.NewSelect().
		TableExpr("sfc_site_posts").
		ColumnExpr("id, title").
		Where("cover_image_id = ?", mediaID).
		Where("deleted_at IS NULL").
		Scan(ctx, &refs)
	if err != nil {
		return nil, fmt.Errorf("media get referencing posts: %w", err)
	}
	return refs, nil
}

// GetBatchReferencingPosts returns a map of mediaID -> reference count for posts.
func (r *Repo) GetBatchReferencingPosts(ctx context.Context, mediaIDs []string) (map[string]int64, error) {
	type row struct {
		CoverImageID string `bun:"cover_image_id"`
		Count        int64  `bun:"count"`
	}
	var rows []row

	err := r.db.NewSelect().
		TableExpr("sfc_site_posts").
		ColumnExpr("cover_image_id").
		ColumnExpr("COUNT(*) AS count").
		Where("cover_image_id IN (?)", bun.In(mediaIDs)).
		Where("deleted_at IS NULL").
		GroupExpr("cover_image_id").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("media get batch referencing posts: %w", err)
	}

	result := make(map[string]int64, len(rows))
	for _, r := range rows {
		result[r.CoverImageID] = r.Count
	}
	return result, nil
}

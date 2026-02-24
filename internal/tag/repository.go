package tag

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// Repo implements TagRepository using uptrace/bun.
type Repo struct {
	db *bun.DB
}

// NewRepo creates a new tag repository.
func NewRepo(db *bun.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) List(ctx context.Context, filter ListFilter) ([]model.Tag, int64, error) {
	var tags []model.Tag

	q := r.db.NewSelect().Model(&tags)

	if filter.Query != "" {
		q = q.Where("name ILIKE ?", "%"+filter.Query+"%")
	}

	switch filter.Sort {
	case "name_desc":
		q = q.OrderExpr("name DESC")
	case "created_asc":
		q = q.OrderExpr("created_at ASC")
	case "created_desc":
		q = q.OrderExpr("created_at DESC")
	default: // "name_asc" or empty
		q = q.OrderExpr("name ASC")
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("tag list count: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage
	err = q.Limit(filter.PerPage).Offset(offset).Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("tag list: %w", err)
	}

	return tags, int64(total), nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (*model.Tag, error) {
	tag := new(model.Tag)
	err := r.db.NewSelect().Model(tag).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("tag not found", err)
		}
		return nil, fmt.Errorf("tag get by id: %w", err)
	}
	return tag, nil
}

func (r *Repo) Create(ctx context.Context, tag *model.Tag) error {
	_, err := r.db.NewInsert().Model(tag).Exec(ctx)
	if err != nil {
		return fmt.Errorf("tag create: %w", err)
	}
	return nil
}

func (r *Repo) Update(ctx context.Context, tag *model.Tag) error {
	_, err := r.db.NewUpdate().Model(tag).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("tag update: %w", err)
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.Tag)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("tag delete: %w", err)
	}
	return nil
}

func (r *Repo) SlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	q := r.db.NewSelect().Model((*model.Tag)(nil)).Where("slug = ?", slug)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("tag slug exists: %w", err)
	}
	return exists, nil
}

func (r *Repo) NameExists(ctx context.Context, name string, excludeID string) (bool, error) {
	q := r.db.NewSelect().Model((*model.Tag)(nil)).Where("name = ?", name)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("tag name exists: %w", err)
	}
	return exists, nil
}

func (r *Repo) CountPosts(ctx context.Context, tagID string) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*model.PostTagMap)(nil)).
		Where("tag_id = ?", tagID).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("tag count posts: %w", err)
	}
	return int64(count), nil
}

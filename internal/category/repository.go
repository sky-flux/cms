package category

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// Repo implements CategoryRepository using uptrace/bun.
type Repo struct {
	db *bun.DB
}

// NewRepo creates a new category repository.
func NewRepo(db *bun.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) List(ctx context.Context) ([]model.Category, error) {
	var cats []model.Category
	err := r.db.NewSelect().
		Model(&cats).
		OrderExpr("sort_order ASC, created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("category list: %w", err)
	}
	return cats, nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (*model.Category, error) {
	cat := new(model.Category)
	err := r.db.NewSelect().Model(cat).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("category not found", err)
		}
		return nil, fmt.Errorf("category get by id: %w", err)
	}
	return cat, nil
}

func (r *Repo) GetChildren(ctx context.Context, parentID string) ([]model.Category, error) {
	var children []model.Category
	err := r.db.NewSelect().
		Model(&children).
		Where("parent_id = ?", parentID).
		OrderExpr("sort_order ASC, created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("category get children: %w", err)
	}
	return children, nil
}

func (r *Repo) Create(ctx context.Context, cat *model.Category) error {
	_, err := r.db.NewInsert().Model(cat).Exec(ctx)
	if err != nil {
		return fmt.Errorf("category create: %w", err)
	}
	return nil
}

func (r *Repo) Update(ctx context.Context, cat *model.Category) error {
	_, err := r.db.NewUpdate().Model(cat).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("category update: %w", err)
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*model.Category)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("category delete: %w", err)
	}
	return nil
}

func (r *Repo) SlugExistsUnderParent(ctx context.Context, slug string, parentID *string, excludeID string) (bool, error) {
	q := r.db.NewSelect().
		Model((*model.Category)(nil)).
		Where("slug = ?", slug)

	if parentID != nil {
		q = q.Where("parent_id = ?", *parentID)
	} else {
		q = q.Where("parent_id IS NULL")
	}

	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}

	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("category slug exists: %w", err)
	}
	return exists, nil
}

func (r *Repo) UpdatePathPrefix(ctx context.Context, oldPrefix, newPrefix string) (int64, error) {
	res, err := r.db.NewUpdate().
		Model((*model.Category)(nil)).
		Set("path = REPLACE(path, ?, ?)", oldPrefix, newPrefix).
		Where("path LIKE ?", oldPrefix+"%").
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("category update path prefix: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *Repo) BatchUpdateSortOrder(ctx context.Context, orders []SortOrderItem) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, o := range orders {
			_, err := tx.NewUpdate().
				Model((*model.Category)(nil)).
				Set("sort_order = ?", o.SortOrder).
				Where("id = ?", o.ID).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("category batch update sort order id=%s: %w", o.ID, err)
			}
		}
		return nil
	})
}

func (r *Repo) CountPosts(ctx context.Context, categoryID string) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*model.PostCategoryMap)(nil)).
		Where("category_id = ?", categoryID).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("category count posts: %w", err)
	}
	return int64(count), nil
}

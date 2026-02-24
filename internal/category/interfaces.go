package category

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// CategoryRepository handles sfc_site_categories table CRUD.
type CategoryRepository interface {
	List(ctx context.Context) ([]model.Category, error)
	GetByID(ctx context.Context, id string) (*model.Category, error)
	GetChildren(ctx context.Context, parentID string) ([]model.Category, error)
	Create(ctx context.Context, cat *model.Category) error
	Update(ctx context.Context, cat *model.Category) error
	Delete(ctx context.Context, id string) error
	SlugExistsUnderParent(ctx context.Context, slug string, parentID *string, excludeID string) (bool, error)
	UpdatePathPrefix(ctx context.Context, oldPrefix, newPrefix string) (int64, error)
	BatchUpdateSortOrder(ctx context.Context, orders []SortOrderItem) error
	CountPosts(ctx context.Context, categoryID string) (int64, error)
}

// SortOrderItem represents a single category sort-order update.
type SortOrderItem struct {
	ID        string `json:"id" binding:"required"`
	SortOrder int    `json:"sort_order"`
}

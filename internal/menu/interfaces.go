package menu

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// MenuRepository handles sfc_site_menus table operations.
type MenuRepository interface {
	List(ctx context.Context, location string) ([]model.SiteMenu, error)
	GetByID(ctx context.Context, id string) (*model.SiteMenu, error)
	Create(ctx context.Context, menu *model.SiteMenu) error
	Update(ctx context.Context, menu *model.SiteMenu) error
	Delete(ctx context.Context, id string) error
	SlugExists(ctx context.Context, slug string, excludeID string) (bool, error)
	CountItems(ctx context.Context, menuID string) (int64, error)
}

// MenuItemRepository handles sfc_site_menu_items table operations.
type MenuItemRepository interface {
	ListByMenuID(ctx context.Context, menuID string) ([]*model.SiteMenuItem, error)
	GetByID(ctx context.Context, id string) (*model.SiteMenuItem, error)
	Create(ctx context.Context, item *model.SiteMenuItem) error
	Update(ctx context.Context, item *model.SiteMenuItem) error
	Delete(ctx context.Context, id string) error
	BelongsToMenu(ctx context.Context, id string, menuID string) (bool, error)
	BatchUpdateOrder(ctx context.Context, items []ReorderItem) error
	GetDepth(ctx context.Context, itemID string) (int, error)
}

// ReorderItem represents a single item's new position in a reorder operation.
type ReorderItem struct {
	ID        string  `json:"id" binding:"required"`
	ParentID  *string `json:"parent_id"`
	SortOrder int     `json:"sort_order"`
}

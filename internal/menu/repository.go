package menu

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// MenuRepo implements MenuRepository.
type MenuRepo struct {
	db *bun.DB
}

// NewMenuRepo creates a new menu repository.
func NewMenuRepo(db *bun.DB) *MenuRepo {
	return &MenuRepo{db: db}
}

func (r *MenuRepo) List(ctx context.Context, location string) ([]model.SiteMenu, error) {
	var menus []model.SiteMenu
	q := r.db.NewSelect().Model(&menus).OrderExpr("created_at ASC")
	if location != "" {
		q = q.Where("location = ?", location)
	}
	if err := q.Scan(ctx); err != nil {
		return nil, fmt.Errorf("menu list: %w", err)
	}
	return menus, nil
}

func (r *MenuRepo) GetByID(ctx context.Context, id string) (*model.SiteMenu, error) {
	menu := new(model.SiteMenu)
	err := r.db.NewSelect().Model(menu).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("menu not found", err)
		}
		return nil, fmt.Errorf("menu get by id: %w", err)
	}
	return menu, nil
}

func (r *MenuRepo) Create(ctx context.Context, menu *model.SiteMenu) error {
	_, err := r.db.NewInsert().Model(menu).Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu create: %w", err)
	}
	return nil
}

func (r *MenuRepo) Update(ctx context.Context, menu *model.SiteMenu) error {
	_, err := r.db.NewUpdate().Model(menu).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu update: %w", err)
	}
	return nil
}

func (r *MenuRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.SiteMenu)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu delete: %w", err)
	}
	return nil
}

func (r *MenuRepo) SlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	q := r.db.NewSelect().Model((*model.SiteMenu)(nil)).Where("slug = ?", slug)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("menu slug exists: %w", err)
	}
	return exists, nil
}

func (r *MenuRepo) CountItems(ctx context.Context, menuID string) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*model.SiteMenuItem)(nil)).
		Where("menu_id = ?", menuID).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("menu count items: %w", err)
	}
	return int64(count), nil
}

// ItemRepo implements MenuItemRepository.
type ItemRepo struct {
	db *bun.DB
}

// NewItemRepo creates a new menu item repository.
func NewItemRepo(db *bun.DB) *ItemRepo {
	return &ItemRepo{db: db}
}

func (r *ItemRepo) ListByMenuID(ctx context.Context, menuID string) ([]*model.SiteMenuItem, error) {
	var items []*model.SiteMenuItem
	err := r.db.NewSelect().
		Model(&items).
		Where("menu_id = ?", menuID).
		OrderExpr("COALESCE(parent_id, id), sort_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("menu item list: %w", err)
	}
	return items, nil
}

func (r *ItemRepo) GetByID(ctx context.Context, id string) (*model.SiteMenuItem, error) {
	item := new(model.SiteMenuItem)
	err := r.db.NewSelect().Model(item).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("menu item not found", err)
		}
		return nil, fmt.Errorf("menu item get by id: %w", err)
	}
	return item, nil
}

func (r *ItemRepo) Create(ctx context.Context, item *model.SiteMenuItem) error {
	_, err := r.db.NewInsert().Model(item).Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu item create: %w", err)
	}
	return nil
}

func (r *ItemRepo) Update(ctx context.Context, item *model.SiteMenuItem) error {
	_, err := r.db.NewUpdate().Model(item).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu item update: %w", err)
	}
	return nil
}

func (r *ItemRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.SiteMenuItem)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu item delete: %w", err)
	}
	return nil
}

func (r *ItemRepo) BelongsToMenu(ctx context.Context, id string, menuID string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*model.SiteMenuItem)(nil)).
		Where("id = ?", id).
		Where("menu_id = ?", menuID).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("menu item belongs to menu: %w", err)
	}
	return exists, nil
}

func (r *ItemRepo) BatchUpdateOrder(ctx context.Context, items []ReorderItem) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, item := range items {
			q := tx.NewUpdate().
				Model((*model.SiteMenuItem)(nil)).
				Set("sort_order = ?", item.SortOrder).
				Set("parent_id = ?", item.ParentID).
				Where("id = ?", item.ID)
			if _, err := q.Exec(ctx); err != nil {
				return fmt.Errorf("reorder item %s: %w", item.ID, err)
			}
		}
		return nil
	})
}

func (r *ItemRepo) GetDepth(ctx context.Context, itemID string) (int, error) {
	depth := 0
	currentID := itemID
	for depth < 5 {
		item := new(model.SiteMenuItem)
		err := r.db.NewSelect().Model(item).Column("parent_id").Where("id = ?", currentID).Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}
			return 0, fmt.Errorf("menu item depth: %w", err)
		}
		if item.ParentID == nil {
			break
		}
		depth++
		currentID = *item.ParentID
	}
	return depth, nil
}

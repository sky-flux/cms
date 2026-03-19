package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/site/domain"
)

type menuRow struct {
	bun.BaseModel `bun:"table:sfc_menus,alias:mn"`

	ID        string    `bun:"id,pk,type:uuid,default:uuidv7()"`
	Name      string    `bun:"name,notnull"`
	Slug      string    `bun:"slug,notnull,unique"`
	CreatedAt time.Time `bun:"created_at,nullzero,default:now()"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,default:now()"`
}

type menuItemRow struct {
	bun.BaseModel `bun:"table:sfc_menu_items,alias:mi"`

	ID        string    `bun:"id,pk,type:uuid,default:uuidv7()"`
	MenuID    string    `bun:"menu_id,type:uuid,notnull"`
	ParentID  string    `bun:"parent_id,type:uuid"`
	Label     string    `bun:"label,notnull"`
	URL       string    `bun:"url,notnull"`
	Order     int       `bun:"sort_order,notnull,default:0"`
	CreatedAt time.Time `bun:"created_at,nullzero,default:now()"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,default:now()"`
}

// BunMenuRepo implements domain.MenuRepository using uptrace/bun.
type BunMenuRepo struct {
	db *bun.DB
}

func NewBunMenuRepo(db *bun.DB) *BunMenuRepo {
	return &BunMenuRepo{db: db}
}

func (r *BunMenuRepo) Save(ctx context.Context, m *domain.Menu) error {
	row := &menuRow{ID: m.ID, Name: m.Name, Slug: m.Slug}
	if m.ID == "" {
		if _, err := r.db.NewInsert().Model(row).Exec(ctx); err != nil {
			return fmt.Errorf("insert menu: %w", err)
		}
		m.ID = row.ID
		return nil
	}
	if _, err := r.db.NewUpdate().Model(row).WherePK().Exec(ctx); err != nil {
		return fmt.Errorf("update menu: %w", err)
	}
	return nil
}

func (r *BunMenuRepo) FindByID(ctx context.Context, id string) (*domain.Menu, error) {
	row := &menuRow{}
	err := r.db.NewSelect().Model(row).Where("mn.id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find menu: %w", err)
	}
	return &domain.Menu{ID: row.ID, Name: row.Name, Slug: row.Slug, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}

func (r *BunMenuRepo) List(ctx context.Context) ([]*domain.Menu, error) {
	var rows []menuRow
	if err := r.db.NewSelect().Model(&rows).OrderExpr("created_at ASC").Scan(ctx); err != nil {
		return nil, fmt.Errorf("list menus: %w", err)
	}
	menus := make([]*domain.Menu, len(rows))
	for i, row := range rows {
		menus[i] = &domain.Menu{ID: row.ID, Name: row.Name, Slug: row.Slug, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
	}
	return menus, nil
}

func (r *BunMenuRepo) Delete(ctx context.Context, id string) error {
	row := &menuRow{ID: id}
	if _, err := r.db.NewDelete().Model(row).WherePK().Exec(ctx); err != nil {
		return fmt.Errorf("delete menu: %w", err)
	}
	return nil
}

func (r *BunMenuRepo) SaveItem(ctx context.Context, item *domain.MenuItem) error {
	row := &menuItemRow{
		ID: item.ID, MenuID: item.MenuID, ParentID: item.ParentID,
		Label: item.Label, URL: item.URL, Order: item.Order,
	}
	if item.ID == "" {
		if _, err := r.db.NewInsert().Model(row).Exec(ctx); err != nil {
			return fmt.Errorf("insert menu item: %w", err)
		}
		item.ID = row.ID
		return nil
	}
	if _, err := r.db.NewUpdate().Model(row).WherePK().Exec(ctx); err != nil {
		return fmt.Errorf("update menu item: %w", err)
	}
	return nil
}

func (r *BunMenuRepo) ListItems(ctx context.Context, menuID string) ([]*domain.MenuItem, error) {
	var rows []menuItemRow
	if err := r.db.NewSelect().Model(&rows).
		Where("mi.menu_id = ?", menuID).
		OrderExpr("mi.sort_order ASC").
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("list menu items: %w", err)
	}
	items := make([]*domain.MenuItem, len(rows))
	for i, row := range rows {
		items[i] = &domain.MenuItem{
			ID: row.ID, MenuID: row.MenuID, ParentID: row.ParentID,
			Label: row.Label, URL: row.URL, Order: row.Order,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
	}
	return items, nil
}

func (r *BunMenuRepo) DeleteItem(ctx context.Context, itemID string) error {
	row := &menuItemRow{ID: itemID}
	if _, err := r.db.NewDelete().Model(row).WherePK().Exec(ctx); err != nil {
		return fmt.Errorf("delete menu item: %w", err)
	}
	return nil
}

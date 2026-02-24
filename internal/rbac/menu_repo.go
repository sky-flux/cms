package rbac

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
	"github.com/uptrace/bun"
)

// MenuRepo handles persistence for admin menus and role-menu associations.
type MenuRepo struct {
	db bun.IDB
}

// NewMenuRepo creates a MenuRepo backed by the given bun database handle.
func NewMenuRepo(db bun.IDB) *MenuRepo {
	return &MenuRepo{db: db}
}

// ListTree returns top-level menus with their children eager-loaded.
func (r *MenuRepo) ListTree(ctx context.Context) ([]model.AdminMenu, error) {
	var menus []model.AdminMenu
	err := r.db.NewSelect().
		Model(&menus).
		Where("parent_id IS NULL").
		Where("status = true").
		Relation("Children", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("status = true").OrderExpr("sort_order ASC")
		}).
		OrderExpr("sort_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return menus, nil
}

// Create inserts a new admin menu.
func (r *MenuRepo) Create(ctx context.Context, menu *model.AdminMenu) error {
	_, err := r.db.NewInsert().Model(menu).Exec(ctx)
	return err
}

// Update modifies an existing admin menu by primary key.
func (r *MenuRepo) Update(ctx context.Context, menu *model.AdminMenu) error {
	_, err := r.db.NewUpdate().Model(menu).WherePK().Exec(ctx)
	return err
}

// Delete removes an admin menu by ID.
func (r *MenuRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*model.AdminMenu)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// GetMenusByRoleID returns all menus assigned to a given role.
func (r *MenuRepo) GetMenusByRoleID(ctx context.Context, roleID string) ([]model.AdminMenu, error) {
	var menus []model.AdminMenu
	err := r.db.NewSelect().
		Model(&menus).
		Join("JOIN sfc_role_menus AS rm ON rm.menu_id = m.id").
		Where("rm.role_id = ?", roleID).
		OrderExpr("m.sort_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return menus, nil
}

// SetRoleMenus replaces all menu associations for a role in a single transaction.
func (r *MenuRepo) SetRoleMenus(ctx context.Context, roleID string, menuIDs []string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete existing associations.
		_, err := tx.NewDelete().
			Model((*model.RoleMenu)(nil)).
			Where("role_id = ?", roleID).
			Exec(ctx)
		if err != nil {
			return err
		}

		if len(menuIDs) == 0 {
			return nil
		}

		// Bulk insert new associations.
		roleMenus := make([]model.RoleMenu, len(menuIDs))
		for i, mid := range menuIDs {
			roleMenus[i] = model.RoleMenu{RoleID: roleID, MenuID: mid}
		}
		_, err = tx.NewInsert().Model(&roleMenus).Exec(ctx)
		return err
	})
}

// GetMenusByUserID returns the union of menus from all roles assigned to a user.
func (r *MenuRepo) GetMenusByUserID(ctx context.Context, userID string) ([]model.AdminMenu, error) {
	var menus []model.AdminMenu
	err := r.db.NewSelect().
		DistinctOn("m.id").
		Model(&menus).
		Join("JOIN sfc_role_menus AS rm ON rm.menu_id = m.id").
		Join("JOIN sfc_user_roles AS ur ON ur.role_id = rm.role_id").
		Where("ur.user_id = ?", userID).
		Where("m.status = true").
		OrderExpr("m.id, m.sort_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return menus, nil
}

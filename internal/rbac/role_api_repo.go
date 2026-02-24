package rbac

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/uptrace/bun"
)

// RoleAPIRepo handles persistence for the sfc_role_apis join table.
type RoleAPIRepo struct {
	db *bun.DB
}

// NewRoleAPIRepo creates a RoleAPIRepo backed by the given bun.DB.
func NewRoleAPIRepo(db *bun.DB) *RoleAPIRepo {
	return &RoleAPIRepo{db: db}
}

// GetAPIsByRoleID returns all API endpoints associated with a role.
func (r *RoleAPIRepo) GetAPIsByRoleID(ctx context.Context, roleID string) ([]model.APIEndpoint, error) {
	var apis []model.APIEndpoint
	err := r.db.NewSelect().
		TableExpr("sfc_apis AS api").
		ColumnExpr("api.*").
		Join("JOIN sfc_role_apis AS ra ON ra.api_id = api.id").
		Where("ra.role_id = ?", roleID).
		OrderExpr(`api."group" ASC, api.method ASC, api.path ASC`).
		Scan(ctx, &apis)
	if err != nil {
		return nil, fmt.Errorf("role api get by role id: %w", err)
	}
	return apis, nil
}

// SetRoleAPIs replaces all API permissions for a role with the given API IDs.
// Runs in a transaction: delete existing, then bulk insert new mappings.
func (r *RoleAPIRepo) SetRoleAPIs(ctx context.Context, roleID string, apiIDs []string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewDelete().
			Model((*model.RoleAPI)(nil)).
			Where("role_id = ?", roleID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("role api delete existing: %w", err)
		}

		if len(apiIDs) == 0 {
			return nil
		}

		mappings := make([]model.RoleAPI, len(apiIDs))
		for i, apiID := range apiIDs {
			mappings[i] = model.RoleAPI{
				RoleID: roleID,
				APIID:  apiID,
			}
		}

		_, err = tx.NewInsert().Model(&mappings).Exec(ctx)
		if err != nil {
			return fmt.Errorf("role api bulk insert: %w", err)
		}
		return nil
	})
}

// GetRoleIDsByMethodPath returns all role IDs that have permission for the
// given HTTP method and path combination. Only active APIs are considered.
func (r *RoleAPIRepo) GetRoleIDsByMethodPath(ctx context.Context, method, path string) ([]string, error) {
	var roleIDs []string
	err := r.db.NewSelect().
		TableExpr("sfc_role_apis AS ra").
		ColumnExpr("ra.role_id").
		Join("JOIN sfc_apis AS a ON a.id = ra.api_id").
		Where("a.method = ?", method).
		Where("a.path = ?", path).
		Where("a.status = true").
		Scan(ctx, &roleIDs)
	if err != nil {
		return nil, fmt.Errorf("role api get role ids by method path: %w", err)
	}
	return roleIDs, nil
}

// CloneFromTemplate replaces a role's API and menu permissions with those
// defined in the given template. Runs in a single transaction.
func (r *RoleAPIRepo) CloneFromTemplate(ctx context.Context, roleID, templateID string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Clear existing API permissions.
		_, err := tx.NewDelete().
			Model((*model.RoleAPI)(nil)).
			Where("role_id = ?", roleID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("clone template clear apis: %w", err)
		}

		// Copy API permissions from template.
		_, err = tx.ExecContext(ctx,
			"INSERT INTO sfc_role_apis (role_id, api_id) SELECT ?, api_id FROM sfc_role_template_apis WHERE template_id = ?",
			roleID, templateID,
		)
		if err != nil {
			return fmt.Errorf("clone template copy apis: %w", err)
		}

		// Clear existing menu permissions.
		_, err = tx.NewDelete().
			Model((*model.RoleMenu)(nil)).
			Where("role_id = ?", roleID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("clone template clear menus: %w", err)
		}

		// Copy menu permissions from template.
		_, err = tx.ExecContext(ctx,
			"INSERT INTO sfc_role_menus (role_id, menu_id) SELECT ?, menu_id FROM sfc_role_template_menus WHERE template_id = ?",
			roleID, templateID,
		)
		if err != nil {
			return fmt.Errorf("clone template copy menus: %w", err)
		}

		return nil
	})
}

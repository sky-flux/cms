package rbac

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/uptrace/bun"
)

// UserRoleRepo handles persistence for the sfc_user_roles table.
type UserRoleRepo struct {
	db *bun.DB
}

// NewUserRoleRepo creates a UserRoleRepo backed by the given bun.DB.
func NewUserRoleRepo(db *bun.DB) *UserRoleRepo {
	return &UserRoleRepo{db: db}
}

// GetRolesByUserID returns all roles assigned to a user via JOIN.
func (r *UserRoleRepo) GetRolesByUserID(ctx context.Context, userID string) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.NewSelect().
		Model(&roles).
		Join("JOIN sfc_user_roles AS ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("user roles get by user id: %w", err)
	}
	return roles, nil
}

// GetRoleSlugs returns only the slug column for all roles assigned to a user.
func (r *UserRoleRepo) GetRoleSlugs(ctx context.Context, userID string) ([]string, error) {
	var slugs []string
	err := r.db.NewSelect().
		TableExpr("sfc_user_roles AS ur").
		Join("JOIN sfc_roles AS r ON r.id = ur.role_id").
		ColumnExpr("r.slug").
		Where("ur.user_id = ?", userID).
		Scan(ctx, &slugs)
	if err != nil {
		return nil, fmt.Errorf("user role slugs: %w", err)
	}
	return slugs, nil
}

// SetUserRoles replaces all roles for a user in a single transaction.
func (r *UserRoleRepo) SetUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete existing role assignments.
		_, err := tx.NewDelete().
			Model((*model.UserRole)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("set user roles delete: %w", err)
		}

		// Insert new role assignments.
		if len(roleIDs) == 0 {
			return nil
		}

		userRoles := make([]model.UserRole, len(roleIDs))
		for i, roleID := range roleIDs {
			userRoles[i] = model.UserRole{
				UserID: userID,
				RoleID: roleID,
			}
		}

		_, err = tx.NewInsert().Model(&userRoles).Exec(ctx)
		if err != nil {
			return fmt.Errorf("set user roles insert: %w", err)
		}

		return nil
	})
}

// HasRole checks whether a user has a specific role (by slug).
func (r *UserRoleRepo) HasRole(ctx context.Context, userID, roleSlug string) (bool, error) {
	exists, err := r.db.NewSelect().
		TableExpr("sfc_user_roles AS ur").
		Join("JOIN sfc_roles AS r ON r.id = ur.role_id").
		Where("ur.user_id = ?", userID).
		Where("r.slug = ?", roleSlug).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("user has role: %w", err)
	}
	return exists, nil
}

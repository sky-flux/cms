package rbac

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// RoleRepo handles persistence for the sfc_roles table.
type RoleRepo struct {
	db *bun.DB
}

// NewRoleRepo creates a RoleRepo backed by the given bun.DB.
func NewRoleRepo(db *bun.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

// List returns all roles ordered by built_in DESC, created_at ASC.
func (r *RoleRepo) List(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.NewSelect().
		Model(&roles).
		OrderExpr("built_in DESC, created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("role list: %w", err)
	}
	return roles, nil
}

// GetByID returns a role by its primary key.
func (r *RoleRepo) GetByID(ctx context.Context, id string) (*model.Role, error) {
	role := new(model.Role)
	err := r.db.NewSelect().
		Model(role).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("role not found", err)
		}
		return nil, fmt.Errorf("role get by id: %w", err)
	}
	return role, nil
}

// GetBySlug returns a role by its unique slug.
func (r *RoleRepo) GetBySlug(ctx context.Context, slug string) (*model.Role, error) {
	role := new(model.Role)
	err := r.db.NewSelect().
		Model(role).
		Where("slug = ?", slug).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("role not found", err)
		}
		return nil, fmt.Errorf("role get by slug: %w", err)
	}
	return role, nil
}

// Create inserts a new role.
func (r *RoleRepo) Create(ctx context.Context, role *model.Role) error {
	_, err := r.db.NewInsert().Model(role).Exec(ctx)
	if err != nil {
		return fmt.Errorf("role create: %w", err)
	}
	return nil
}

// Update saves changes to an existing role (matched by PK).
func (r *RoleRepo) Update(ctx context.Context, role *model.Role) error {
	_, err := r.db.NewUpdate().Model(role).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("role update: %w", err)
	}
	return nil
}

// Delete removes a custom (non-built-in) role by ID.
// Built-in roles are protected and cannot be deleted.
func (r *RoleRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*model.Role)(nil)).
		Where("id = ?", id).
		Where("built_in = false").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("role delete: %w", err)
	}
	return nil
}

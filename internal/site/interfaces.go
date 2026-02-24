package site

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// SiteRepository handles sfc_sites table CRUD.
type SiteRepository interface {
	List(ctx context.Context, filter ListFilter) ([]model.Site, int64, error)
	GetBySlug(ctx context.Context, slug string) (*model.Site, error)
	Create(ctx context.Context, site *model.Site) error
	Update(ctx context.Context, site *model.Site) error
	Delete(ctx context.Context, id string) error
	CountActive(ctx context.Context) (int64, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	DomainExists(ctx context.Context, domain, excludeID string) (bool, error)
}

// UserRoleRepository manages user-role assignments (sfc_user_roles).
type UserRoleRepository interface {
	ListUsersWithRoles(ctx context.Context, filter UserFilter) ([]UserWithRole, int64, error)
	AssignRole(ctx context.Context, userID, roleID string) error
	RemoveRole(ctx context.Context, userID string) error
	UserExists(ctx context.Context, userID string) (bool, error)
}

// RoleResolver looks up roles by slug.
type RoleResolver interface {
	GetBySlug(ctx context.Context, slug string) (*model.Role, error)
}

// RBACInvalidator abstracts RBAC cache invalidation.
type RBACInvalidator interface {
	InvalidateUserCache(ctx context.Context, userID string) error
}

// SchemaManager abstracts site schema lifecycle.
type SchemaManager interface {
	Create(ctx context.Context, slug string) error
	Drop(ctx context.Context, slug string) error
}

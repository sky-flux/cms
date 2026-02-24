package rbac

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// UserRoleRepository defines the contract for user-role data access.
type UserRoleRepository interface {
	GetRolesByUserID(ctx context.Context, userID string) ([]model.Role, error)
	GetRoleSlugs(ctx context.Context, userID string) ([]string, error)
	SetUserRoles(ctx context.Context, userID string, roleIDs []string) error
	HasRole(ctx context.Context, userID, roleSlug string) (bool, error)
}

// RoleAPIRepository defines the contract for role-API permission data access.
type RoleAPIRepository interface {
	GetAPIsByRoleID(ctx context.Context, roleID string) ([]model.APIEndpoint, error)
	SetRoleAPIs(ctx context.Context, roleID string, apiIDs []string) error
	GetRoleIDsByMethodPath(ctx context.Context, method, path string) ([]string, error)
	CloneFromTemplate(ctx context.Context, roleID, templateID string) error
}

// MenuRepository defines the contract for admin menu data access.
type MenuRepository interface {
	ListTree(ctx context.Context) ([]model.AdminMenu, error)
	Create(ctx context.Context, menu *model.AdminMenu) error
	Update(ctx context.Context, menu *model.AdminMenu) error
	Delete(ctx context.Context, id string) error
	GetMenusByRoleID(ctx context.Context, roleID string) ([]model.AdminMenu, error)
	SetRoleMenus(ctx context.Context, roleID string, menuIDs []string) error
	GetMenusByUserID(ctx context.Context, userID string) ([]model.AdminMenu, error)
}

// RoleRepository defines the contract for role data access.
type RoleRepository interface {
	List(ctx context.Context) ([]model.Role, error)
	GetByID(ctx context.Context, id string) (*model.Role, error)
	GetBySlug(ctx context.Context, slug string) (*model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id string) error
}

// APIRepository defines the contract for API endpoint data access.
type APIRepository interface {
	UpsertBatch(ctx context.Context, endpoints []model.APIEndpoint) error
	DisableStale(ctx context.Context, activeKeys []string) error
	List(ctx context.Context) ([]model.APIEndpoint, error)
	ListByGroup(ctx context.Context, group string) ([]model.APIEndpoint, error)
	GetByMethodPath(ctx context.Context, method, path string) (*model.APIEndpoint, error)
}

// TemplateRepository defines the contract for role template data access.
type TemplateRepository interface {
	List(ctx context.Context) ([]model.RoleTemplate, error)
	GetByID(ctx context.Context, id string) (*model.RoleTemplate, error)
	Create(ctx context.Context, tmpl *model.RoleTemplate) error
	Update(ctx context.Context, tmpl *model.RoleTemplate) error
	Delete(ctx context.Context, id string) error
	GetTemplateAPIs(ctx context.Context, templateID string) ([]model.APIEndpoint, error)
	SetTemplateAPIs(ctx context.Context, templateID string, apiIDs []string) error
	GetTemplateMenus(ctx context.Context, templateID string) ([]model.AdminMenu, error)
	SetTemplateMenus(ctx context.Context, templateID string, menuIDs []string) error
}

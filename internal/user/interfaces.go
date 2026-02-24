package user

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/mail"
)

// Ensure interfaces are used (suppress lint).
var (
	_ audit.Logger = (*audit.Service)(nil)
	_ mail.Sender  = (*mail.NoopSender)(nil)
)

// UserRepository handles sfc_users table CRUD.
type UserRepository interface {
	List(ctx context.Context, filter ListFilter) ([]model.User, int64, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	SoftDelete(ctx context.Context, id string) error
}

// RoleRepository looks up roles by slug.
type RoleRepository interface {
	GetBySlug(ctx context.Context, slug string) (*model.Role, error)
}

// UserRoleRepository manages user-role assignments (sfc_user_roles).
type UserRoleRepository interface {
	Assign(ctx context.Context, userID, roleID string) error
	GetRoleSlug(ctx context.Context, userID string) (string, error)
	CountActiveByRoleSlug(ctx context.Context, roleSlug string) (int64, error)
}

// TokenRevoker revokes all refresh tokens for a user.
type TokenRevoker interface {
	RevokeAllForUser(ctx context.Context, userID string) error
}

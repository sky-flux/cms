package setup

import (
	"context"
	"encoding/json"

	"github.com/sky-flux/cms/internal/model"
)

type ConfigRepository interface {
	GetValue(ctx context.Context, key string) (json.RawMessage, error)
	SetValue(ctx context.Context, key string, value interface{}) error
}

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
}

type SiteRepository interface {
	Create(ctx context.Context, site *model.Site) error
}

type UserRoleRepository interface {
	AssignRole(ctx context.Context, userID, roleSlug string) error
}

type RoleRepository interface {
	GetBySlug(ctx context.Context, slug string) (*model.Role, error)
}

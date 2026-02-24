package auth

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	UpdateLastLogin(ctx context.Context, id string) error
}

type TokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
	CreatePasswordResetToken(ctx context.Context, token *model.PasswordResetToken) error
	GetPasswordResetTokenByHash(ctx context.Context, hash string) (*model.PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, id string) error
}

type TOTPRepository interface {
	GetByUserID(ctx context.Context, userID string) (*model.UserTOTP, error)
	Upsert(ctx context.Context, totp *model.UserTOTP) error
	Enable(ctx context.Context, id string) error
	Delete(ctx context.Context, userID string) error
	UpdateBackupCodes(ctx context.Context, id string, codes []string) error
}

type RoleLoader interface {
	GetUserRoles(ctx context.Context, userID string) ([]model.Role, error)
}

type SiteLoader interface {
	GetUserSites(ctx context.Context, userID string) ([]model.Site, error)
}

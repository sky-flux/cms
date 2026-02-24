package auth

import (
	"context"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// --- User Repository ---

type authUserRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) UserRepository {
	return &authUserRepo{db: db}
}

func (r *authUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.NewSelect().Model(&user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, apperror.NotFound("user not found", err)
	}
	return &user, nil
}

func (r *authUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.NewSelect().Model(&user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, apperror.NotFound("user not found", err)
	}
	return &user, nil
}

func (r *authUserRepo) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.db.NewUpdate().Model((*model.User)(nil)).
		Set("password_hash = ?", passwordHash).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *authUserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.User)(nil)).
		Set("last_login_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// --- Token Repository ---

type authTokenRepo struct {
	db *bun.DB
}

func NewTokenRepo(db *bun.DB) TokenRepository {
	return &authTokenRepo{db: db}
}

func (r *authTokenRepo) CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	_, err := r.db.NewInsert().Model(token).Exec(ctx)
	return err
}

func (r *authTokenRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.NewSelect().Model(&token).
		Where("token_hash = ?", hash).
		Where("revoked = false").
		Where("expires_at > NOW()").
		Scan(ctx)
	if err != nil {
		return nil, apperror.NotFound("refresh token not found or expired", err)
	}
	return &token, nil
}

func (r *authTokenRepo) RevokeRefreshToken(ctx context.Context, id string) error {
	_, err := r.db.NewUpdate().Model((*model.RefreshToken)(nil)).
		Set("revoked = true").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *authTokenRepo) RevokeAllUserTokens(ctx context.Context, userID string) error {
	_, err := r.db.NewUpdate().Model((*model.RefreshToken)(nil)).
		Set("revoked = true").
		Where("user_id = ?", userID).
		Where("revoked = false").
		Exec(ctx)
	return err
}

func (r *authTokenRepo) CreatePasswordResetToken(ctx context.Context, token *model.PasswordResetToken) error {
	_, err := r.db.NewInsert().Model(token).Exec(ctx)
	return err
}

func (r *authTokenRepo) GetPasswordResetTokenByHash(ctx context.Context, hash string) (*model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	err := r.db.NewSelect().Model(&token).
		Where("token_hash = ?", hash).
		Where("used_at IS NULL").
		Where("expires_at > NOW()").
		Scan(ctx)
	if err != nil {
		return nil, apperror.NotFound("password reset token not found or expired", err)
	}
	return &token, nil
}

func (r *authTokenRepo) MarkPasswordResetTokenUsed(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.PasswordResetToken)(nil)).
		Set("used_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// --- TOTP Repository ---

type authTOTPRepo struct {
	db *bun.DB
}

func NewTOTPRepo(db *bun.DB) TOTPRepository {
	return &authTOTPRepo{db: db}
}

func (r *authTOTPRepo) GetByUserID(ctx context.Context, userID string) (*model.UserTOTP, error) {
	var totp model.UserTOTP
	err := r.db.NewSelect().Model(&totp).Where("user_id = ?", userID).Scan(ctx)
	if err != nil {
		return nil, apperror.NotFound("TOTP record not found", err)
	}
	return &totp, nil
}

func (r *authTOTPRepo) Upsert(ctx context.Context, totp *model.UserTOTP) error {
	_, err := r.db.NewInsert().Model(totp).
		On("CONFLICT (user_id) DO UPDATE").
		Set("secret_encrypted = EXCLUDED.secret_encrypted").
		Set("backup_codes_hash = EXCLUDED.backup_codes_hash").
		Set("is_enabled = EXCLUDED.is_enabled").
		Set("verified_at = EXCLUDED.verified_at").
		Set("updated_at = NOW()").
		Exec(ctx)
	return err
}

func (r *authTOTPRepo) Enable(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.UserTOTP)(nil)).
		Set("is_enabled = true").
		Set("verified_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *authTOTPRepo) Delete(ctx context.Context, userID string) error {
	_, err := r.db.NewDelete().Model((*model.UserTOTP)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)
	return err
}

func (r *authTOTPRepo) UpdateBackupCodes(ctx context.Context, id string, codes []string) error {
	_, err := r.db.NewUpdate().Model((*model.UserTOTP)(nil)).
		Set("backup_codes_hash = ?", pgdialect.Array(codes)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// --- Role Loader ---

type authRoleLoader struct {
	db *bun.DB
}

func NewRoleLoader(db *bun.DB) RoleLoader {
	return &authRoleLoader{db: db}
}

func (r *authRoleLoader) GetUserRoles(ctx context.Context, userID string) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.NewSelect().Model(&roles).
		Join("JOIN sfc_user_roles AS ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// --- Site Loader ---

type authSiteLoader struct {
	db *bun.DB
}

func NewSiteLoader(db *bun.DB) SiteLoader {
	return &authSiteLoader{db: db}
}

func (r *authSiteLoader) GetUserSites(ctx context.Context, userID string) ([]model.Site, error) {
	// For now, return all active sites (super admin sees all)
	// TODO: filter by user role assignments when site-level RBAC is implemented
	var sites []model.Site
	err := r.db.NewSelect().Model(&sites).
		Where("is_active = true").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return sites, nil
}

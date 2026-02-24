package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// --- UserRepo ---

type UserRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) List(ctx context.Context, f ListFilter) ([]model.User, int64, error) {
	var users []model.User
	q := r.db.NewSelect().Model(&users).Where("deleted_at IS NULL")

	if f.Query != "" {
		q = q.Where("(display_name ILIKE ? OR email ILIKE ?)", "%"+f.Query+"%", "%"+f.Query+"%")
	}
	if f.Role != "" {
		q = q.Where("id IN (SELECT ur.user_id FROM sfc_user_roles ur JOIN sfc_roles r ON r.id = ur.role_id WHERE r.slug = ?)", f.Role)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("user list count: %w", err)
	}

	offset := (f.Page - 1) * f.PerPage
	if offset < 0 {
		offset = 0
	}

	err = q.OrderExpr("created_at DESC").
		Limit(f.PerPage).
		Offset(offset).
		Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("user list: %w", err)
	}
	return users, int64(total), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Where("deleted_at IS NULL").Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("user not found", err)
		}
		return nil, fmt.Errorf("user get by id: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().Model(user).Where("email = ?", email).Where("deleted_at IS NULL").Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("user not found", err)
		}
		return nil, fmt.Errorf("user get by email: %w", err)
	}
	return user, nil
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return fmt.Errorf("user create: %w", err)
	}
	return nil
}

func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	_, err := r.db.NewUpdate().Model(user).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("user update: %w", err)
	}
	return nil
}

func (r *UserRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.User)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("user soft delete: %w", err)
	}
	return nil
}

// --- RoleRepo ---

type RoleRepo struct {
	db *bun.DB
}

func NewRoleRepo(db *bun.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

func (r *RoleRepo) GetBySlug(ctx context.Context, slug string) (*model.Role, error) {
	role := new(model.Role)
	err := r.db.NewSelect().Model(role).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("role not found", err)
		}
		return nil, fmt.Errorf("role get by slug: %w", err)
	}
	return role, nil
}

// --- UserRoleRepo ---

type UserRoleRepo struct {
	db *bun.DB
}

func NewUserRoleRepo(db *bun.DB) *UserRoleRepo {
	return &UserRoleRepo{db: db}
}

func (r *UserRoleRepo) Assign(ctx context.Context, userID, roleID string) error {
	ur := &model.UserRole{UserID: userID, RoleID: roleID}
	_, err := r.db.NewInsert().Model(ur).
		On("CONFLICT (user_id) DO UPDATE SET role_id = EXCLUDED.role_id").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("user role assign: %w", err)
	}
	return nil
}

func (r *UserRoleRepo) GetRoleSlug(ctx context.Context, userID string) (string, error) {
	var slug string
	err := r.db.NewSelect().
		TableExpr("sfc_user_roles AS ur").
		ColumnExpr("r.slug").
		Join("JOIN sfc_roles AS r ON r.id = ur.role_id").
		Where("ur.user_id = ?", userID).
		Scan(ctx, &slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("user role get slug: %w", err)
	}
	return slug, nil
}

func (r *UserRoleRepo) CountActiveByRoleSlug(ctx context.Context, roleSlug string) (int64, error) {
	count, err := r.db.NewSelect().
		TableExpr("sfc_user_roles AS ur").
		Join("JOIN sfc_roles AS r ON r.id = ur.role_id").
		Join("JOIN sfc_users AS u ON u.id = ur.user_id").
		Where("r.slug = ?", roleSlug).
		Where("u.status = ?", model.UserStatusActive).
		Where("u.deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count active by role slug: %w", err)
	}
	return int64(count), nil
}

// --- TokenRevokerImpl ---

type TokenRevokerImpl struct {
	db *bun.DB
}

func NewTokenRevoker(db *bun.DB) *TokenRevokerImpl {
	return &TokenRevokerImpl{db: db}
}

func (r *TokenRevokerImpl) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.db.NewUpdate().Model((*model.RefreshToken)(nil)).
		Set("revoked = ?", model.ToggleYes).
		Where("user_id = ?", userID).
		Where("revoked = ?", model.ToggleNo).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("revoke all tokens for user: %w", err)
	}
	return nil
}

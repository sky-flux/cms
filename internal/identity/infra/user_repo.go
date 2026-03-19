package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/sky-flux/cms/internal/identity/domain"
)

// bunUser is the bun ORM model. Kept private to infra; mapped to/from domain.User.
type bunUser struct {
	bun.BaseModel `bun:"table:sfc_users,alias:u"`

	ID           string            `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Email        string            `bun:"email,notnull,unique"`
	PasswordHash string            `bun:"password_hash,notnull"`
	DisplayName  string            `bun:"display_name,notnull"`
	AvatarURL    string            `bun:"avatar_url,notnull,default:''"`
	Status       domain.UserStatus `bun:"status,notnull,type:smallint,default:1"`
	LastLoginAt  *time.Time        `bun:"last_login_at"`
	CreatedAt    time.Time         `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt    time.Time         `bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt    *time.Time        `bun:"deleted_at,soft_delete,nullzero"`
}

// BunUserRepo implements domain.UserRepository using uptrace/bun.
type BunUserRepo struct {
	db *bun.DB
}

func NewBunUserRepo(db *bun.DB) *BunUserRepo {
	return &BunUserRepo{db: db}
}

func (r *BunUserRepo) Save(ctx context.Context, u *domain.User) error {
	row := domainToRow(u)
	_, err := r.db.NewInsert().Model(row).Exec(ctx)
	if err != nil {
		return fmt.Errorf("user_repo.Save: %w", err)
	}
	u.ID = row.ID // propagate DB-assigned ID back to domain entity
	return nil
}

func (r *BunUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := new(bunUser)
	err := r.db.NewSelect().Model(row).Where("email = ?", email).WhereAllWithDeleted().Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("user_repo.FindByEmail: %w", err)
	}
	return rowToDomain(row), nil
}

func (r *BunUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := new(bunUser)
	err := r.db.NewSelect().Model(row).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("user_repo.FindByID: %w", err)
	}
	return rowToDomain(row), nil
}

func (r *BunUserRepo) UpdatePassword(ctx context.Context, id, hash string) error {
	_, err := r.db.NewUpdate().TableExpr("sfc_users").
		Set("password_hash = ?", hash).
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *BunUserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	_, err := r.db.NewUpdate().TableExpr("sfc_users").
		Set("last_login_at = NOW()").
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// --- mapping helpers ---

func domainToRow(u *domain.User) *bunUser {
	return &bunUser{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		DisplayName:  u.DisplayName,
		AvatarURL:    u.AvatarURL,
		Status:       u.Status,
		LastLoginAt:  u.LastLoginAt,
	}
}

func rowToDomain(r *bunUser) *domain.User {
	return &domain.User{
		ID:           r.ID,
		Email:        r.Email,
		PasswordHash: r.PasswordHash,
		DisplayName:  r.DisplayName,
		AvatarURL:    r.AvatarURL,
		Status:       r.Status,
		LastLoginAt:  r.LastLoginAt,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

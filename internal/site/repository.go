package site

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/schema"
	"github.com/uptrace/bun"
)

// --- SiteRepo ---

type SiteRepo struct {
	db *bun.DB
}

func NewSiteRepo(db *bun.DB) *SiteRepo {
	return &SiteRepo{db: db}
}

func (r *SiteRepo) List(ctx context.Context, f ListFilter) ([]model.Site, int64, error) {
	var sites []model.Site
	q := r.db.NewSelect().Model(&sites)

	if f.Query != "" {
		q = q.Where("(name ILIKE ? OR slug ILIKE ?)", "%"+f.Query+"%", "%"+f.Query+"%")
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("site list count: %w", err)
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
		return nil, 0, fmt.Errorf("site list: %w", err)
	}
	return sites, int64(total), nil
}

func (r *SiteRepo) GetBySlug(ctx context.Context, slug string) (*model.Site, error) {
	site := new(model.Site)
	err := r.db.NewSelect().Model(site).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("site not found", err)
		}
		return nil, fmt.Errorf("site get by slug: %w", err)
	}
	return site, nil
}

func (r *SiteRepo) Create(ctx context.Context, site *model.Site) error {
	_, err := r.db.NewInsert().Model(site).Exec(ctx)
	if err != nil {
		return fmt.Errorf("site create: %w", err)
	}
	return nil
}

func (r *SiteRepo) Update(ctx context.Context, site *model.Site) error {
	_, err := r.db.NewUpdate().Model(site).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("site update: %w", err)
	}
	return nil
}

func (r *SiteRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.Site)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("site delete: %w", err)
	}
	return nil
}

func (r *SiteRepo) CountActive(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().Model((*model.Site)(nil)).Where("status = ?", model.SiteStatusActive).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("site count active: %w", err)
	}
	return int64(count), nil
}

func (r *SiteRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	exists, err := r.db.NewSelect().Model((*model.Site)(nil)).Where("slug = ?", slug).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("site slug exists: %w", err)
	}
	return exists, nil
}

func (r *SiteRepo) DomainExists(ctx context.Context, domain, excludeID string) (bool, error) {
	q := r.db.NewSelect().Model((*model.Site)(nil)).Where("domain = ?", domain)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("site domain exists: %w", err)
	}
	return exists, nil
}

// --- UserRoleRepo (for site user management) ---

type UserRoleRepo struct {
	db *bun.DB
}

func NewUserRoleRepo(db *bun.DB) *UserRoleRepo {
	return &UserRoleRepo{db: db}
}

func (r *UserRoleRepo) ListUsersWithRoles(ctx context.Context, f UserFilter) ([]UserWithRole, int64, error) {
	type row struct {
		model.User
		RoleSlug   string       `bun:"role_slug"`
		AssignedAt sql.NullTime `bun:"assigned_at"`
	}

	var rows []row
	q := r.db.NewSelect().
		TableExpr("sfc_users AS u").
		ColumnExpr("u.*").
		ColumnExpr("r.slug AS role_slug").
		ColumnExpr("ur.created_at AS assigned_at").
		Join("JOIN sfc_user_roles AS ur ON ur.user_id = u.id").
		Join("JOIN sfc_roles AS r ON r.id = ur.role_id")

	if f.Role != "" {
		q = q.Where("r.slug = ?", f.Role)
	}
	if f.Query != "" {
		q = q.Where("(u.display_name ILIKE ? OR u.email ILIKE ?)", "%"+f.Query+"%", "%"+f.Query+"%")
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("site users count: %w", err)
	}

	offset := (f.Page - 1) * f.PerPage
	if offset < 0 {
		offset = 0
	}

	err = q.OrderExpr("ur.created_at DESC").
		Limit(f.PerPage).
		Offset(offset).
		Scan(ctx, &rows)
	if err != nil {
		return nil, 0, fmt.Errorf("site users list: %w", err)
	}

	result := make([]UserWithRole, len(rows))
	for i, row := range rows {
		result[i] = UserWithRole{
			User:     row.User,
			RoleSlug: row.RoleSlug,
		}
		if row.AssignedAt.Valid {
			result[i].CreatedAt = row.AssignedAt.Time
		}
	}
	return result, int64(total), nil
}

func (r *UserRoleRepo) AssignRole(ctx context.Context, userID, roleID string) error {
	ur := &model.UserRole{UserID: userID, RoleID: roleID}
	_, err := r.db.NewInsert().Model(ur).
		On("CONFLICT (user_id, role_id) DO NOTHING").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("assign role: %w", err)
	}
	return nil
}

func (r *UserRoleRepo) RemoveRole(ctx context.Context, userID string) error {
	_, err := r.db.NewDelete().Model((*model.UserRole)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("remove role: %w", err)
	}
	return nil
}

func (r *UserRoleRepo) UserExists(ctx context.Context, userID string) (bool, error) {
	exists, err := r.db.NewSelect().Model((*model.User)(nil)).Where("id = ?", userID).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("user exists: %w", err)
	}
	return exists, nil
}

// --- RoleResolverImpl ---

type RoleResolverImpl struct {
	db *bun.DB
}

func NewRoleResolver(db *bun.DB) *RoleResolverImpl {
	return &RoleResolverImpl{db: db}
}

func (r *RoleResolverImpl) GetBySlug(ctx context.Context, slug string) (*model.Role, error) {
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

// --- SchemaManagerImpl ---

type SchemaManagerImpl struct {
	db *bun.DB
}

func NewSchemaManager(db *bun.DB) *SchemaManagerImpl {
	return &SchemaManagerImpl{db: db}
}

func (m *SchemaManagerImpl) Create(ctx context.Context, slug string) error {
	return schema.CreateSiteSchema(ctx, m.db, slug)
}

func (m *SchemaManagerImpl) Drop(ctx context.Context, slug string) error {
	return schema.DropSiteSchema(ctx, m.db, slug)
}

# Batch 2+3 Implementation Plan: Sites Management + RBAC Completion

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement 8 Sites management endpoints and complete the RBAC module (3 missing handler methods + router registration + API Registry).

**Architecture:** Handler → Service → Repository (三层). Sites module is built from scratch; RBAC is mostly filling gaps and wiring routes.

**Tech Stack:** Go / Gin / uptrace/bun / PostgreSQL / Redis / testify / miniredis

**Design doc:** `docs/plans/2026-02-24-global-endpoints-batch2-3-design.md`

**Reference docs:** `docs/api.md` (sections 4.2 + RBAC), `docs/database.md` (sfc_sites, sfc_user_roles, sfc_roles)

---

### Task 1: Site module interfaces and DTOs

**Files:**
- Create: `internal/site/interfaces.go`
- Create: `internal/site/dto.go`

**Step 1:** Write `internal/site/interfaces.go` — repository and dependency interfaces.

```go
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
```

**Step 2:** Write `internal/site/dto.go` — request/response DTOs and filter types.

```go
package site

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Filters ---

type ListFilter struct {
	Page     int
	PerPage  int
	Query    string
	IsActive *bool
}

type UserFilter struct {
	Page    int
	PerPage int
	Role    string
	Query   string
}

// --- Request DTOs ---

type CreateSiteReq struct {
	Name          string `json:"name" binding:"required,max=200"`
	Slug          string `json:"slug" binding:"required,min=3,max=50"`
	Domain        string `json:"domain"`
	Description   string `json:"description"`
	DefaultLocale string `json:"default_locale"`
	Timezone      string `json:"timezone"`
}

type UpdateSiteReq struct {
	Name          *string `json:"name" binding:"omitempty,max=200"`
	Domain        *string `json:"domain"`
	Description   *string `json:"description"`
	LogoURL       *string `json:"logo_url" binding:"omitempty,url"`
	DefaultLocale *string `json:"default_locale"`
	Timezone      *string `json:"timezone"`
	IsActive      *bool   `json:"is_active"`
	Settings      *string `json:"settings"`
}

type DeleteSiteReq struct {
	ConfirmSlug string `json:"confirm_slug" binding:"required"`
}

type AssignSiteRoleReq struct {
	Role string `json:"role" binding:"required"`
}

// --- Response DTOs ---

type SiteResp struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Domain        string    `json:"domain,omitempty"`
	Description   string    `json:"description,omitempty"`
	LogoURL       string    `json:"logo_url,omitempty"`
	DefaultLocale string    `json:"default_locale"`
	Timezone      string    `json:"timezone"`
	IsActive      bool      `json:"is_active"`
	Settings      string    `json:"settings,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func ToSiteResp(s *model.Site) SiteResp {
	return SiteResp{
		ID:            s.ID,
		Name:          s.Name,
		Slug:          s.Slug,
		Domain:        s.Domain,
		Description:   s.Description,
		LogoURL:       s.LogoURL,
		DefaultLocale: s.DefaultLocale,
		Timezone:      s.Timezone,
		IsActive:      s.IsActive,
		Settings:      s.Settings,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

func ToSiteRespList(sites []model.Site) []SiteResp {
	out := make([]SiteResp, len(sites))
	for i := range sites {
		out[i] = ToSiteResp(&sites[i])
	}
	return out
}

// UserWithRole is a joined view of user + their role.
type UserWithRole struct {
	User      model.User `json:"user"`
	RoleSlug  string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
}

type SiteUserResp struct {
	User      UserBriefResp `json:"user"`
	Role      string        `json:"role"`
	CreatedAt time.Time     `json:"created_at"`
}

type UserBriefResp struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	IsActive    bool   `json:"is_active"`
}

func ToSiteUserRespList(items []UserWithRole) []SiteUserResp {
	out := make([]SiteUserResp, len(items))
	for i, item := range items {
		out[i] = SiteUserResp{
			User: UserBriefResp{
				ID:          item.User.ID,
				Email:       item.User.Email,
				DisplayName: item.User.DisplayName,
				AvatarURL:   item.User.AvatarURL,
				IsActive:    item.User.IsActive,
			},
			Role:      item.RoleSlug,
			CreatedAt: item.CreatedAt,
		}
	}
	return out
}
```

**Step 3:** Verify compilation.

Run: `go build ./internal/site/...`
Expected: success (no errors)

**Step 4:** Commit.

```bash
git add internal/site/interfaces.go internal/site/dto.go
git commit -m "feat(site): add interfaces and DTOs for sites management"
```

---

### Task 2: Site repository implementation

**Files:**
- Create: `internal/site/repository.go` (replace stub)

**Step 1:** Write `internal/site/repository.go` — bun-backed implementations for SiteRepository, UserRoleRepository, RoleResolver, and SchemaManager.

```go
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
	if f.IsActive != nil {
		q = q.Where("is_active = ?", *f.IsActive)
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
	count, err := r.db.NewSelect().Model((*model.Site)(nil)).Where("is_active = true").Count(ctx)
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
		RoleSlug  string    `bun:"role_slug"`
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
```

**Step 2:** Verify compilation.

Run: `go build ./internal/site/...`
Expected: success

**Step 3:** Commit.

```bash
git add internal/site/repository.go
git commit -m "feat(site): implement repository layer for sites and user-role management"
```

---

### Task 3: Site service implementation

**Files:**
- Create: `internal/site/service.go` (replace stub)

**Step 1:** Write `internal/site/service.go` — business logic for all 8 endpoints.

```go
package site

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9_]{3,50}$`)

type Service struct {
	siteRepo     SiteRepository
	userRoleRepo UserRoleRepository
	roleResolver RoleResolver
	rbac         RBACInvalidator
	schemaMgr    SchemaManager
}

func NewService(
	siteRepo SiteRepository,
	userRoleRepo UserRoleRepository,
	roleResolver RoleResolver,
	rbac RBACInvalidator,
	schemaMgr SchemaManager,
) *Service {
	return &Service{
		siteRepo:     siteRepo,
		userRoleRepo: userRoleRepo,
		roleResolver: roleResolver,
		rbac:         rbac,
		schemaMgr:    schemaMgr,
	}
}

func (s *Service) ListSites(ctx context.Context, f ListFilter) ([]model.Site, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 20
	}
	return s.siteRepo.List(ctx, f)
}

func (s *Service) GetSite(ctx context.Context, slug string) (*model.Site, error) {
	return s.siteRepo.GetBySlug(ctx, slug)
}

func (s *Service) CreateSite(ctx context.Context, req *CreateSiteReq) (*model.Site, error) {
	if !slugRegex.MatchString(req.Slug) {
		return nil, apperror.Validation("invalid slug: must match ^[a-z0-9_]{3,50}$", nil)
	}

	exists, err := s.siteRepo.SlugExists(ctx, req.Slug)
	if err != nil {
		return nil, fmt.Errorf("create site check slug: %w", err)
	}
	if exists {
		return nil, apperror.Conflict("site slug already exists", nil)
	}

	if req.Domain != "" {
		domainExists, err := s.siteRepo.DomainExists(ctx, req.Domain, "")
		if err != nil {
			return nil, fmt.Errorf("create site check domain: %w", err)
		}
		if domainExists {
			return nil, apperror.Conflict("domain already exists", nil)
		}
	}

	locale := req.DefaultLocale
	if locale == "" {
		locale = "zh-CN"
	}
	tz := req.Timezone
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	site := &model.Site{
		Name:          req.Name,
		Slug:          req.Slug,
		Domain:        req.Domain,
		Description:   req.Description,
		DefaultLocale: locale,
		Timezone:      tz,
		IsActive:      true,
	}

	if err := s.siteRepo.Create(ctx, site); err != nil {
		return nil, fmt.Errorf("create site insert: %w", err)
	}

	if err := s.schemaMgr.Create(ctx, req.Slug); err != nil {
		slog.Error("create site schema failed", "error", err, "slug", req.Slug)
		return nil, apperror.Internal("site schema creation failed", err)
	}

	return site, nil
}

func (s *Service) UpdateSite(ctx context.Context, slug string, req *UpdateSiteReq) (*model.Site, error) {
	site, err := s.siteRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if req.Domain != nil && *req.Domain != site.Domain {
		if *req.Domain != "" {
			domainExists, err := s.siteRepo.DomainExists(ctx, *req.Domain, site.ID)
			if err != nil {
				return nil, fmt.Errorf("update site check domain: %w", err)
			}
			if domainExists {
				return nil, apperror.Conflict("domain already exists", nil)
			}
		}
		site.Domain = *req.Domain
	}

	if req.Name != nil {
		site.Name = *req.Name
	}
	if req.Description != nil {
		site.Description = *req.Description
	}
	if req.LogoURL != nil {
		site.LogoURL = *req.LogoURL
	}
	if req.DefaultLocale != nil {
		site.DefaultLocale = *req.DefaultLocale
	}
	if req.Timezone != nil {
		site.Timezone = *req.Timezone
	}
	if req.IsActive != nil {
		site.IsActive = *req.IsActive
	}
	if req.Settings != nil {
		site.Settings = *req.Settings
	}

	if err := s.siteRepo.Update(ctx, site); err != nil {
		return nil, fmt.Errorf("update site: %w", err)
	}
	return site, nil
}

func (s *Service) DeleteSite(ctx context.Context, slug string, confirmSlug string) error {
	if slug != confirmSlug {
		return apperror.Validation("confirm_slug does not match", nil)
	}

	site, err := s.siteRepo.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}

	count, err := s.siteRepo.CountActive(ctx)
	if err != nil {
		return fmt.Errorf("delete site count: %w", err)
	}
	if count <= 1 {
		return &apperror.AppError{Code: 422, Message: "cannot delete the last site"}
	}

	if err := s.siteRepo.Delete(ctx, site.ID); err != nil {
		return fmt.Errorf("delete site: %w", err)
	}

	if err := s.schemaMgr.Drop(ctx, slug); err != nil {
		slog.Error("drop site schema failed", "error", err, "slug", slug)
	}

	return nil
}

func (s *Service) ListSiteUsers(ctx context.Context, slug string, f UserFilter) ([]UserWithRole, int64, error) {
	if _, err := s.siteRepo.GetBySlug(ctx, slug); err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 20
	}
	return s.userRoleRepo.ListUsersWithRoles(ctx, f)
}

func (s *Service) AssignSiteRole(ctx context.Context, slug, userID string, roleSlug string) error {
	if _, err := s.siteRepo.GetBySlug(ctx, slug); err != nil {
		return err
	}

	exists, err := s.userRoleRepo.UserExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("assign role check user: %w", err)
	}
	if !exists {
		return apperror.NotFound("user not found", nil)
	}

	role, err := s.roleResolver.GetBySlug(ctx, roleSlug)
	if err != nil {
		return err
	}

	if err := s.userRoleRepo.AssignRole(ctx, userID, role.ID); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}

	if err := s.rbac.InvalidateUserCache(ctx, userID); err != nil {
		slog.Error("invalidate user cache after role assign", "error", err, "user_id", userID)
	}

	return nil
}

func (s *Service) RemoveSiteRole(ctx context.Context, slug, userID string) error {
	if _, err := s.siteRepo.GetBySlug(ctx, slug); err != nil {
		return err
	}

	if err := s.userRoleRepo.RemoveRole(ctx, userID); err != nil {
		return fmt.Errorf("remove role: %w", err)
	}

	if err := s.rbac.InvalidateUserCache(ctx, userID); err != nil {
		slog.Error("invalidate user cache after role remove", "error", err, "user_id", userID)
	}

	return nil
}
```

**Step 2:** Verify compilation.

Run: `go build ./internal/site/...`
Expected: success

**Step 3:** Commit.

```bash
git add internal/site/service.go
git commit -m "feat(site): implement service layer with schema lifecycle and cache invalidation"
```

---

### Task 4: Site handler implementation

**Files:**
- Create: `internal/site/handler.go` (replace stub)

**Step 1:** Write `internal/site/handler.go` — 8 HTTP handler methods.

```go
package site

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListSites(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	f := ListFilter{
		Page:    page,
		PerPage: perPage,
		Query:   c.Query("q"),
	}
	if v := c.Query("is_active"); v != "" {
		active := v == "true"
		f.IsActive = &active
	}

	sites, total, err := h.svc.ListSites(c.Request.Context(), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, ToSiteRespList(sites), total, page, perPage)
}

func (h *Handler) CreateSite(c *gin.Context) {
	var req CreateSiteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	site, err := h.svc.CreateSite(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	resp := ToSiteResp(site)
	response.Created(c, resp)
}

func (h *Handler) GetSite(c *gin.Context) {
	site, err := h.svc.GetSite(c.Request.Context(), c.Param("slug"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToSiteResp(site))
}

func (h *Handler) UpdateSite(c *gin.Context) {
	var req UpdateSiteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	site, err := h.svc.UpdateSite(c.Request.Context(), c.Param("slug"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ToSiteResp(site))
}

func (h *Handler) DeleteSite(c *gin.Context) {
	var req DeleteSiteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.DeleteSite(c.Request.Context(), c.Param("slug"), req.ConfirmSlug); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "site deleted"})
}

func (h *Handler) ListSiteUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	f := UserFilter{
		Page:    page,
		PerPage: perPage,
		Role:    c.Query("role"),
		Query:   c.Query("q"),
	}

	users, total, err := h.svc.ListSiteUsers(c.Request.Context(), c.Param("slug"), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, ToSiteUserRespList(users), total, page, perPage)
}

func (h *Handler) AssignSiteRole(c *gin.Context) {
	var req AssignSiteRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	slug := c.Param("slug")
	userID := c.Param("user_id")

	if err := h.svc.AssignSiteRole(c.Request.Context(), slug, userID, req.Role); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"user_id": userID, "site_slug": slug, "role": req.Role})
}

func (h *Handler) RemoveSiteRole(c *gin.Context) {
	slug := c.Param("slug")
	userID := c.Param("user_id")

	if err := h.svc.RemoveSiteRole(c.Request.Context(), slug, userID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "user role removed"})
}
```

**Step 2:** Verify compilation.

Run: `go build ./internal/site/...`
Expected: success

**Step 3:** Commit.

```bash
git add internal/site/handler.go
git commit -m "feat(site): implement HTTP handler with 8 site management endpoints"
```

---

### Task 5: Site service tests

**Files:**
- Create: `internal/site/service_test.go` (replace stub)

**Step 1:** Write `internal/site/service_test.go` with mock repos and comprehensive test cases.

The test file should include:
- Mock implementations for all 5 interfaces (SiteRepository, UserRoleRepository, RoleResolver, RBACInvalidator, SchemaManager)
- Test helper `newTestService(t)` that returns service + mocks
- Tests for all 8 service methods:
  - `TestService_ListSites_Success` — returns paginated list
  - `TestService_ListSites_DefaultsPagination` — page < 1 defaults to 1
  - `TestService_GetSite_Success` — returns site by slug
  - `TestService_GetSite_NotFound` — returns 404
  - `TestService_CreateSite_Success` — creates with schema
  - `TestService_CreateSite_InvalidSlug` — rejects bad slug format
  - `TestService_CreateSite_SlugConflict` — returns 409
  - `TestService_CreateSite_DomainConflict` — returns 409
  - `TestService_CreateSite_DefaultLocale` — empty locale defaults to zh-CN
  - `TestService_UpdateSite_Success` — partial update
  - `TestService_UpdateSite_DomainConflict` — new domain conflicts
  - `TestService_UpdateSite_NotFound` — site not found
  - `TestService_DeleteSite_Success` — deletes + drops schema
  - `TestService_DeleteSite_ConfirmMismatch` — confirm_slug mismatch
  - `TestService_DeleteSite_LastSite` — cannot delete last site
  - `TestService_ListSiteUsers_SiteNotFound` — validates site existence
  - `TestService_AssignSiteRole_Success` — assigns + invalidates cache
  - `TestService_AssignSiteRole_UserNotFound` — user doesn't exist
  - `TestService_AssignSiteRole_RoleNotFound` — role doesn't exist
  - `TestService_RemoveSiteRole_Success` — removes + invalidates cache

Use the same mock pattern as `internal/auth/service_test.go` — struct-based mocks with configurable return values, not testify/mock.

**Step 2:** Run tests.

Run: `go test ./internal/site/... -v -count=1`
Expected: all tests PASS

**Step 3:** Commit.

```bash
git add internal/site/service_test.go
git commit -m "test(site): add comprehensive service unit tests with 20 test cases"
```

---

### Task 6: Site handler tests

**Files:**
- Create: `internal/site/handler_test.go` (replace stub)

**Step 1:** Write `internal/site/handler_test.go` — HTTP-level tests through real gin router.

The test file should include:
- `setupTestRouter(t)` helper that creates handler with mock repos and returns `*gin.Engine`
- Tests for all 8 handler methods:
  - `TestHandler_ListSites_Success` — 200 with paginated response
  - `TestHandler_CreateSite_Success` — 201 with site data
  - `TestHandler_CreateSite_ValidationError` — 422 for missing name
  - `TestHandler_GetSite_Success` — 200 with site data
  - `TestHandler_GetSite_NotFound` — 404
  - `TestHandler_UpdateSite_Success` — 200 with updated data
  - `TestHandler_DeleteSite_Success` — 200 with message
  - `TestHandler_DeleteSite_MissingConfirm` — 422
  - `TestHandler_ListSiteUsers_Success` — 200 with paginated users
  - `TestHandler_AssignSiteRole_Success` — 200
  - `TestHandler_AssignSiteRole_InvalidBody` — 422
  - `TestHandler_RemoveSiteRole_Success` — 200

Reuse mock types from service_test.go (same `site_test` package).

**Step 2:** Run tests.

Run: `go test ./internal/site/... -v -count=1`
Expected: all tests PASS (service + handler)

**Step 3:** Commit.

```bash
git add internal/site/handler_test.go
git commit -m "test(site): add HTTP handler tests for all 8 endpoints"
```

---

### Task 7: RBAC handler completion — add GetRole, ListMenus, fix ApplyTemplate

**Files:**
- Modify: `internal/rbac/handler.go` — add `GetRole`, `ListMenus`; replace `ApplyTemplate` stub

**Step 1:** Add `GetRole` method after `ListRoles` (around line 53):

```go
func (h *Handler) GetRole(c *gin.Context) {
	role, err := h.roleRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, role)
}
```

**Step 2:** Add `ListMenus` method after `SetRoleMenus` (around line 199):

```go
func (h *Handler) ListMenus(c *gin.Context) {
	menus, err := h.menuRepo.ListTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, menus)
}
```

**Step 3:** Add `GetTemplate` method after `ListTemplates` (around line 210):

```go
func (h *Handler) GetTemplate(c *gin.Context) {
	tmpl, err := h.templateRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, tmpl)
}
```

**Step 4:** Replace the `ApplyTemplate` TODO stub (around line 286-303) with:

```go
func (h *Handler) ApplyTemplate(c *gin.Context) {
	roleID := c.Param("id")
	ctx := c.Request.Context()

	role, err := h.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		response.Error(c, err)
		return
	}

	if role.BuiltIn && role.Slug == "super" {
		response.Error(c, apperror.Forbidden("cannot modify super role permissions", nil))
		return
	}

	var req ApplyTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if _, err := h.templateRepo.GetByID(ctx, req.TemplateID); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.roleAPIRepo.CloneFromTemplate(ctx, roleID, req.TemplateID); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.InvalidateRoleCache(ctx, roleID); err != nil {
		slog.Error("invalidate role cache after template apply", "error", err, "role_id", roleID)
	}

	response.NoContent(c)
}
```

**Step 5:** Add `CreateMenu`, `UpdateMenu`, `DeleteMenu` handler methods:

```go
func (h *Handler) CreateMenu(c *gin.Context) {
	var req CreateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	menu := &model.AdminMenu{
		ParentID:  req.ParentID,
		Name:      req.Name,
		Icon:      req.Icon,
		Path:      req.Path,
		SortOrder: req.SortOrder,
	}

	if err := h.menuRepo.Create(c.Request.Context(), menu); err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, menu)
}

func (h *Handler) UpdateMenu(c *gin.Context) {
	// Fetch the menu by ID to get current values, then apply partial update.
	// MenuRepo.Update expects the full model, so we build it from the request.
	var req UpdateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	menu := &model.AdminMenu{ID: c.Param("id")}
	if req.Name != "" {
		menu.Name = req.Name
	}
	if req.Icon != nil {
		menu.Icon = *req.Icon
	}
	if req.Path != nil {
		menu.Path = *req.Path
	}
	if req.SortOrder != nil {
		menu.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		menu.Status = *req.Status
	}
	if req.ParentID != nil {
		menu.ParentID = req.ParentID
	}

	if err := h.menuRepo.Update(c.Request.Context(), menu); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, menu)
}

func (h *Handler) DeleteMenu(c *gin.Context) {
	if err := h.menuRepo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
```

**Step 6:** Add missing DTO types in `internal/rbac/dto.go`:

```go
// --- Menu ---

type CreateMenuReq struct {
	ParentID  *string `json:"parent_id" binding:"omitempty,uuid"`
	Name      string  `json:"name" binding:"required,max=50"`
	Icon      string  `json:"icon"`
	Path      string  `json:"path"`
	SortOrder int     `json:"sort_order"`
}

type UpdateMenuReq struct {
	ParentID  *string `json:"parent_id"`
	Name      string  `json:"name" binding:"omitempty,max=50"`
	Icon      *string `json:"icon"`
	Path      *string `json:"path"`
	SortOrder *int    `json:"sort_order"`
	Status    *bool   `json:"status"`
}
```

**Step 7:** Verify compilation.

Run: `go build ./internal/rbac/...`
Expected: success

**Step 8:** Commit.

```bash
git add internal/rbac/handler.go internal/rbac/dto.go
git commit -m "feat(rbac): add GetRole, GetTemplate, ListMenus, CreateMenu, UpdateMenu, DeleteMenu handlers and implement ApplyTemplate"
```

---

### Task 8: RBAC handler tests for new methods

**Files:**
- Modify: `internal/rbac/handler_test.go` — add tests for new handler methods

**Step 1:** Add tests for the new methods at the end of `handler_test.go`:

- `TestHandler_GetRole_Success` — returns role by ID
- `TestHandler_GetRole_NotFound` — returns 404
- `TestHandler_GetTemplate_Success` — returns template by ID
- `TestHandler_GetTemplate_NotFound` — returns 404
- `TestHandler_ApplyTemplate_Success` — returns 204
- `TestHandler_ApplyTemplate_SuperProtected` — returns 403 for super role
- `TestHandler_ApplyTemplate_RoleNotFound` — returns 404
- `TestHandler_ApplyTemplate_TemplateNotFound` — returns 404
- `TestHandler_ListMenus_Success` — returns menu tree
- `TestHandler_CreateMenu_Success` — returns 201
- `TestHandler_DeleteMenu_Success` — returns 204

**Step 2:** Run tests.

Run: `go test ./internal/rbac/... -v -count=1`
Expected: all tests PASS (existing + new)

**Step 3:** Commit.

```bash
git add internal/rbac/handler_test.go
git commit -m "test(rbac): add tests for GetRole, GetTemplate, ApplyTemplate, ListMenus, CreateMenu, DeleteMenu"
```

---

### Task 9: API Registry metadata map

**Files:**
- Create: `internal/router/api_meta.go`

**Step 1:** Write `internal/router/api_meta.go` — defines metadata for all RBAC-protected endpoints.

```go
package router

import "github.com/sky-flux/cms/internal/rbac"

// BuildAPIMetaMap returns metadata for all RBAC-protected endpoints.
// The key format is "METHOD:/api/v1/path" matching Gin's FullPath().
// Endpoints NOT in this map are public or JWT-only (no RBAC check).
func BuildAPIMetaMap() map[string]rbac.APIMeta {
	return map[string]rbac.APIMeta{
		// Auth admin
		"DELETE:/api/v1/auth/2fa/users/:user_id": {Name: "Force disable 2FA", Description: "Force disable another user's 2FA", Group: "auth"},

		// Sites management
		"GET:/api/v1/sites":                              {Name: "List sites", Description: "List all sites", Group: "sites"},
		"POST:/api/v1/sites":                             {Name: "Create site", Description: "Create a new site with schema", Group: "sites"},
		"GET:/api/v1/sites/:slug":                        {Name: "Get site", Description: "Get site details by slug", Group: "sites"},
		"PUT:/api/v1/sites/:slug":                        {Name: "Update site", Description: "Update site information", Group: "sites"},
		"DELETE:/api/v1/sites/:slug":                     {Name: "Delete site", Description: "Delete site and drop schema", Group: "sites"},
		"GET:/api/v1/sites/:slug/users":                  {Name: "List site users", Description: "List users with roles", Group: "sites"},
		"PUT:/api/v1/sites/:slug/users/:user_id/role":    {Name: "Assign site role", Description: "Assign role to user", Group: "sites"},
		"DELETE:/api/v1/sites/:slug/users/:user_id/role": {Name: "Remove site role", Description: "Remove user role assignment", Group: "sites"},

		// RBAC roles
		"GET:/api/v1/rbac/roles":                     {Name: "List roles", Description: "List all roles", Group: "rbac"},
		"POST:/api/v1/rbac/roles":                    {Name: "Create role", Description: "Create custom role", Group: "rbac"},
		"GET:/api/v1/rbac/roles/:id":                 {Name: "Get role", Description: "Get role details", Group: "rbac"},
		"PUT:/api/v1/rbac/roles/:id":                 {Name: "Update role", Description: "Update role", Group: "rbac"},
		"DELETE:/api/v1/rbac/roles/:id":              {Name: "Delete role", Description: "Delete custom role", Group: "rbac"},
		"GET:/api/v1/rbac/roles/:id/apis":            {Name: "Get role APIs", Description: "List API permissions for role", Group: "rbac"},
		"PUT:/api/v1/rbac/roles/:id/apis":            {Name: "Set role APIs", Description: "Set API permissions for role", Group: "rbac"},
		"GET:/api/v1/rbac/roles/:id/menus":           {Name: "Get role menus", Description: "List menu visibility for role", Group: "rbac"},
		"PUT:/api/v1/rbac/roles/:id/menus":           {Name: "Set role menus", Description: "Set menu visibility for role", Group: "rbac"},
		"POST:/api/v1/rbac/roles/:id/apply-template": {Name: "Apply template", Description: "Apply permission template to role", Group: "rbac"},

		// RBAC user roles
		"GET:/api/v1/rbac/users/:id/roles":  {Name: "Get user roles", Description: "List roles for user", Group: "rbac"},
		"POST:/api/v1/rbac/users/:id/roles": {Name: "Set user roles", Description: "Set roles for user", Group: "rbac"},

		// RBAC menus
		"GET:/api/v1/rbac/menus":       {Name: "List menus", Description: "List admin menu tree", Group: "rbac"},
		"POST:/api/v1/rbac/menus":      {Name: "Create menu", Description: "Create admin menu item", Group: "rbac"},
		"PUT:/api/v1/rbac/menus/:id":   {Name: "Update menu", Description: "Update admin menu item", Group: "rbac"},
		"DELETE:/api/v1/rbac/menus/:id": {Name: "Delete menu", Description: "Delete admin menu item", Group: "rbac"},

		// RBAC APIs
		"GET:/api/v1/rbac/apis": {Name: "List APIs", Description: "List registered API endpoints", Group: "rbac"},

		// RBAC templates
		"GET:/api/v1/rbac/templates":       {Name: "List templates", Description: "List permission templates", Group: "rbac"},
		"POST:/api/v1/rbac/templates":      {Name: "Create template", Description: "Create permission template", Group: "rbac"},
		"GET:/api/v1/rbac/templates/:id":   {Name: "Get template", Description: "Get template details", Group: "rbac"},
		"PUT:/api/v1/rbac/templates/:id":   {Name: "Update template", Description: "Update permission template", Group: "rbac"},
		"DELETE:/api/v1/rbac/templates/:id": {Name: "Delete template", Description: "Delete permission template", Group: "rbac"},
	}
}
```

**Step 2:** Verify compilation.

Run: `go build ./internal/router/...`
Expected: success

**Step 3:** Commit.

```bash
git add internal/router/api_meta.go
git commit -m "feat(router): add API metadata map for RBAC-protected endpoints"
```

---

### Task 10: Router registration — wire sites + RBAC route groups

**Files:**
- Modify: `internal/router/router.go` — add sites + RBAC route groups with DI

**Step 1:** Update `router.go` to:
1. Import `internal/site` package
2. After auth module DI, add site module DI (repos → service → handler)
3. After RBAC service creation, add RBAC handler creation (repos exist, just need to wire NewHandler)
4. Add sites route group (JWT + RBAC)
5. Add RBAC route group (JWT + RBAC)
6. Add RBAC /me/menus route (JWT only)

Key wiring in `Setup()`:

```go
// ── Site module (repos → service → handler) ─────────────────
siteRepo := site.NewSiteRepo(db)
siteUserRoleRepo := site.NewUserRoleRepo(db)
siteRoleResolver := site.NewRoleResolver(db)
siteSchemaMgr := site.NewSchemaManager(db)
siteSvc := site.NewService(siteRepo, siteUserRoleRepo, siteRoleResolver, rbacSvc, siteSchemaMgr)
siteHandler := site.NewHandler(siteSvc)

// ── RBAC handler (repos already created above) ──────────────
rbacRoleRepo := rbac.NewRoleRepo(db)
rbacAPIRepo := rbac.NewAPIRepo(db)
rbacTemplateRepo := rbac.NewTemplateRepo(db)
rbacHandler := rbac.NewHandler(rbacSvc, rbacRoleRepo, rbacAPIRepo, rbacRoleAPIRepo, rbacMenuRepo, rbacTemplateRepo, rbacUserRoleRepo)
```

Route groups to add after existing auth routes:

```go
// Sites management (JWT + RBAC)
sites := v1.Group("/sites")
sites.Use(middleware.Auth(jwtMgr))
sites.Use(middleware.RBAC(rbacSvc))
sites.GET("", siteHandler.ListSites)
sites.POST("", siteHandler.CreateSite)
sites.GET("/:slug", siteHandler.GetSite)
sites.PUT("/:slug", siteHandler.UpdateSite)
sites.DELETE("/:slug", siteHandler.DeleteSite)
sites.GET("/:slug/users", siteHandler.ListSiteUsers)
sites.PUT("/:slug/users/:user_id/role", siteHandler.AssignSiteRole)
sites.DELETE("/:slug/users/:user_id/role", siteHandler.RemoveSiteRole)

// RBAC management (JWT + RBAC)
rbacGroup := v1.Group("/rbac")
rbacGroup.Use(middleware.Auth(jwtMgr))
rbacGroup.Use(middleware.RBAC(rbacSvc))
rbacGroup.GET("/roles", rbacHandler.ListRoles)
rbacGroup.POST("/roles", rbacHandler.CreateRole)
rbacGroup.GET("/roles/:id", rbacHandler.GetRole)
rbacGroup.PUT("/roles/:id", rbacHandler.UpdateRole)
rbacGroup.DELETE("/roles/:id", rbacHandler.DeleteRole)
rbacGroup.GET("/roles/:id/apis", rbacHandler.GetRoleAPIs)
rbacGroup.PUT("/roles/:id/apis", rbacHandler.SetRoleAPIs)
rbacGroup.GET("/roles/:id/menus", rbacHandler.GetRoleMenus)
rbacGroup.PUT("/roles/:id/menus", rbacHandler.SetRoleMenus)
rbacGroup.POST("/roles/:id/apply-template", rbacHandler.ApplyTemplate)
rbacGroup.GET("/users/:id/roles", rbacHandler.GetUserRoles)
rbacGroup.POST("/users/:id/roles", rbacHandler.SetUserRoles)
rbacGroup.GET("/menus", rbacHandler.ListMenus)
rbacGroup.POST("/menus", rbacHandler.CreateMenu)
rbacGroup.PUT("/menus/:id", rbacHandler.UpdateMenu)
rbacGroup.DELETE("/menus/:id", rbacHandler.DeleteMenu)
rbacGroup.GET("/apis", rbacHandler.ListAPIs)
rbacGroup.GET("/templates", rbacHandler.ListTemplates)
rbacGroup.POST("/templates", rbacHandler.CreateTemplate)
rbacGroup.GET("/templates/:id", rbacHandler.GetTemplate)
rbacGroup.PUT("/templates/:id", rbacHandler.UpdateTemplate)
rbacGroup.DELETE("/templates/:id", rbacHandler.DeleteTemplate)

// My menus (JWT only — every authenticated user can see their own menus)
rbacMe := v1.Group("/rbac")
rbacMe.Use(middleware.Auth(jwtMgr))
rbacMe.GET("/me/menus", rbacHandler.GetMyMenus)
```

**Step 2:** Verify compilation.

Run: `go build ./internal/router/... && go build ./cmd/cms/...`
Expected: success

**Step 3:** Commit.

```bash
git add internal/router/router.go
git commit -m "feat(router): register sites and RBAC route groups with full DI wiring"
```

---

### Task 11: API Registry sync at startup

**Files:**
- Modify: `cmd/cms/serve.go` — add SyncRoutes call after router.Setup
- Modify: `internal/router/router.go` — export engine reference or return API repo

**Step 1:** The simplest approach: call SyncRoutes inside `router.Setup()` since it has access to `engine`, `db`, and can create the registry.

Add at the end of `router.Setup()` (after all routes registered):

```go
// ── API Registry — sync routes to sfc_apis ──────────────────
registry := rbac.NewRegistry(rbacAPIRepo)
metaMap := BuildAPIMetaMap()
go func() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := registry.SyncRoutes(ctx, engine, metaMap); err != nil {
		slog.Error("api registry sync failed", "error", err)
	}
}()
```

Note: `context` and `time` imports are already present in router.go. `slog` needs to be added to imports.

**Step 2:** Verify compilation.

Run: `go build ./cmd/cms/...`
Expected: success

**Step 3:** Commit.

```bash
git add internal/router/router.go
git commit -m "feat(router): add API Registry sync at startup for RBAC dynamic permissions"
```

---

### Task 12: Full test suite verification

**Step 1:** Run all tests (excluding schema tests that need Docker).

Run: `go test $(go list ./... | grep -v internal/schema) -count=1`
Expected: all PASS

**Step 2:** Run vet.

Run: `go vet ./...`
Expected: clean

**Step 3:** Verify full build.

Run: `go build ./cmd/cms/...`
Expected: success

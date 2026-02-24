# Global Endpoints — Batch 2+3 Design: Sites + RBAC Completion

**Date**: 2026-02-24
**Scope**: Sites (8 new endpoints) + RBAC completion (23 endpoints, mostly wiring)
**Architecture**: Handler → Service → Repository (三层, same as Batch 1)

---

## 1. Overview

Batch 2 implements Sites management — CRUD for multi-site instances plus user-role assignment per site. Batch 3 completes the RBAC module by filling in missing handler methods, registering all routes, and wiring the API Registry.

### What Already Exists

**RBAC module** (`internal/rbac/`) is ~90% implemented:
- 6 repositories: RoleRepo, APIRepo, RoleAPIRepo, UserRoleRepo, MenuRepo, TemplateRepo (all with tests)
- Service: CheckPermission, GetUserMenuTree, InvalidateUserCache/RoleCache
- Handler: 20 of 23 methods implemented
- DTOs and interfaces: complete
- API Registry: SyncRoutes implemented

**Sites module** (`internal/site/`) is stub-only — all 4 files are `// TODO` placeholders.

### Dependencies on Existing Code
- Models: Site, Role, UserRole, AdminMenu, APIEndpoint, RoleTemplate (all exist)
- Schema: CreateSiteSchema, DropSiteSchema (exist in `internal/schema/`)
- RBAC Service: InvalidateUserCache (exists)
- Middleware: Auth (JWT), RBAC (both exist)
- Utilities: apperror, response, pkg/crypto (all exist)

---

## 2. Batch 2 — Sites Management (8 endpoints)

### New Files

| File | Purpose |
|------|---------|
| `internal/site/interfaces.go` | Repository + dependency interfaces |
| `internal/site/dto.go` | Request/Response DTOs with validation tags |
| `internal/site/repository.go` | Site CRUD + user-role queries via bun |
| `internal/site/service.go` | Business logic: schema lifecycle, cache invalidation, validation |
| `internal/site/handler.go` | HTTP handlers: 8 endpoint methods |
| `internal/site/service_test.go` | Service unit tests (mock repos) |
| `internal/site/handler_test.go` | Handler HTTP tests |

### Endpoint Mapping

```
GET    /api/v1/sites                              → site.Handler.ListSites         [JWT+RBAC]
POST   /api/v1/sites                              → site.Handler.CreateSite        [JWT+RBAC]
GET    /api/v1/sites/:slug                        → site.Handler.GetSite           [JWT+RBAC]
PUT    /api/v1/sites/:slug                        → site.Handler.UpdateSite        [JWT+RBAC]
DELETE /api/v1/sites/:slug                        → site.Handler.DeleteSite        [JWT+RBAC]
GET    /api/v1/sites/:slug/users                  → site.Handler.ListSiteUsers     [JWT+RBAC]
PUT    /api/v1/sites/:slug/users/:user_id/role    → site.Handler.AssignSiteRole    [JWT+RBAC]
DELETE /api/v1/sites/:slug/users/:user_id/role    → site.Handler.RemoveSiteRole    [JWT+RBAC]
```

All 8 endpoints require JWT + RBAC middleware. Per the permission matrix, only Super role has access.

### Interface Definitions

```go
// SiteRepository handles sfc_sites table CRUD.
type SiteRepository interface {
    List(ctx context.Context, filter ListFilter) ([]model.Site, int, error)
    GetBySlug(ctx context.Context, slug string) (*model.Site, error)
    Create(ctx context.Context, site *model.Site) error
    Update(ctx context.Context, site *model.Site) error
    Delete(ctx context.Context, id string) error
    CountActive(ctx context.Context) (int, error)
    SlugExists(ctx context.Context, slug string) (bool, error)
    DomainExists(ctx context.Context, domain string, excludeID string) (bool, error)
}

// SiteUserRepository handles user-role queries scoped to a site context.
type SiteUserRepository interface {
    ListUsersWithRoles(ctx context.Context, siteSlug string, filter UserFilter) ([]UserWithRole, int, error)
    AssignRole(ctx context.Context, userID, roleID string) error
    RemoveRole(ctx context.Context, userID, roleID string) error
    GetUserRole(ctx context.Context, userID string) (*model.UserRole, error)
}

// RBACInvalidator abstracts RBAC cache invalidation (implemented by rbac.Service).
type RBACInvalidator interface {
    InvalidateUserCache(ctx context.Context, userID string) error
}

// SchemaManager abstracts site schema lifecycle.
type SchemaManager interface {
    CreateSiteSchema(ctx context.Context, db bun.IDB, slug string) error
    DropSiteSchema(ctx context.Context, db bun.IDB, slug string) error
}
```

### Key Business Flows

#### Create Site Flow
1. Validate request body (name, slug format `^[a-z0-9_]{3,50}$`, optional domain)
2. Check slug uniqueness → 409 SITE_SLUG_EXISTS
3. Check domain uniqueness (if provided) → 409 DOMAIN_EXISTS
4. Transaction: INSERT site → CreateSiteSchema(slug) → seed default sfc_site_configs
5. COMMIT → return 201 with site data

#### Delete Site Flow
1. Validate confirm_slug matches :slug param → 400
2. Check site exists → 404
3. Count active sites → if only 1 → 422 "cannot delete last site"
4. Transaction: soft-delete site record → DropSiteSchema(slug)
5. Flush Redis site cache keys

#### Assign/Remove Role Flow
1. Validate site exists (by slug) → 404
2. Validate role exists (by slug from request body) → 404
3. Upsert/Delete user_role record
4. Call rbac.Service.InvalidateUserCache(userID) → flush L1 cache

### DTOs

```go
// Request DTOs
type CreateSiteReq struct {
    Name          string `json:"name" binding:"required,max=200"`
    Slug          string `json:"slug" binding:"required,min=3,max=50"`
    Domain        string `json:"domain" binding:"omitempty,fqdn"`
    Description   string `json:"description"`
    DefaultLocale string `json:"default_locale"`
    Timezone      string `json:"timezone"`
}

type UpdateSiteReq struct {
    Name          *string `json:"name" binding:"omitempty,max=200"`
    Domain        *string `json:"domain" binding:"omitempty"`
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

// Response DTOs
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

type SiteUserResp struct {
    User      UserBriefResp `json:"user"`
    Role      string        `json:"role"`
    CreatedAt time.Time     `json:"created_at"`
}

type ListFilter struct {
    Page     int
    PerPage  int
    Query    string // search name/slug
    IsActive *bool
}

type UserFilter struct {
    Page    int
    PerPage int
    Role    string
    Query   string // search name/email
}
```

---

## 3. Batch 3 — RBAC Completion

### Missing Handler Methods (3 methods to implement)

| Method | Status | What to do |
|--------|--------|------------|
| `GetRole` | **Missing** | Add `h.roleRepo.GetByID()` → response.Success |
| `ApplyTemplate` | **Stub (TODO)** | Parse req → verify role + template exist → CloneFromTemplate → InvalidateRoleCache |
| `ListMenus` | **Missing** | Add `h.menuRepo.ListTree()` → response.Success |

All other 20 handler methods are already implemented and tested.

### Router Registration

Current `router.go` registers: setup (2) + auth (17) + health (1) = 20 routes.

Add two new route groups:

```go
// Sites management (JWT + RBAC)
sites := v1.Group("/sites")
sites.Use(middleware.Auth(jwtMgr))
sites.Use(middleware.RBAC(rbacSvc))
{
    sites.GET("",                              siteHandler.ListSites)
    sites.POST("",                             siteHandler.CreateSite)
    sites.GET("/:slug",                        siteHandler.GetSite)
    sites.PUT("/:slug",                        siteHandler.UpdateSite)
    sites.DELETE("/:slug",                     siteHandler.DeleteSite)
    sites.GET("/:slug/users",                  siteHandler.ListSiteUsers)
    sites.PUT("/:slug/users/:user_id/role",    siteHandler.AssignSiteRole)
    sites.DELETE("/:slug/users/:user_id/role", siteHandler.RemoveSiteRole)
}

// RBAC management (JWT + RBAC, except /me/menus which is JWT-only)
rbacGroup := v1.Group("/rbac")
rbacGroup.Use(middleware.Auth(jwtMgr))
rbacGroup.Use(middleware.RBAC(rbacSvc))
{
    rbacGroup.GET("/roles",                     rbacHandler.ListRoles)
    rbacGroup.POST("/roles",                    rbacHandler.CreateRole)
    rbacGroup.GET("/roles/:id",                 rbacHandler.GetRole)
    rbacGroup.PUT("/roles/:id",                 rbacHandler.UpdateRole)
    rbacGroup.DELETE("/roles/:id",              rbacHandler.DeleteRole)
    rbacGroup.GET("/roles/:id/apis",            rbacHandler.GetRoleAPIs)
    rbacGroup.PUT("/roles/:id/apis",            rbacHandler.SetRoleAPIs)
    rbacGroup.GET("/roles/:id/menus",           rbacHandler.GetRoleMenus)
    rbacGroup.PUT("/roles/:id/menus",           rbacHandler.SetRoleMenus)
    rbacGroup.POST("/roles/:id/apply-template", rbacHandler.ApplyTemplate)
    rbacGroup.GET("/users/:id/roles",           rbacHandler.GetUserRoles)
    rbacGroup.POST("/users/:id/roles",          rbacHandler.SetUserRoles)
    rbacGroup.GET("/menus",                     rbacHandler.ListMenus)
    rbacGroup.POST("/menus",                    rbacHandler.CreateMenu)
    rbacGroup.PUT("/menus/:id",                 rbacHandler.UpdateMenu)
    rbacGroup.DELETE("/menus/:id",              rbacHandler.DeleteMenu)
    rbacGroup.GET("/apis",                      rbacHandler.ListAPIs)
    rbacGroup.GET("/templates",                 rbacHandler.ListTemplates)
    rbacGroup.POST("/templates",                rbacHandler.CreateTemplate)
    rbacGroup.GET("/templates/:id",             rbacHandler.GetTemplate)
    rbacGroup.PUT("/templates/:id",             rbacHandler.UpdateTemplate)
    rbacGroup.DELETE("/templates/:id",          rbacHandler.DeleteTemplate)
}

// My menus (JWT only, no RBAC — every authenticated user can see their own menus)
rbacMe := v1.Group("/rbac")
rbacMe.Use(middleware.Auth(jwtMgr))
rbacMe.GET("/me/menus", rbacHandler.GetMyMenus)
```

### API Registry MetaMap

Define a `BuildAPIMetaMap()` function in `internal/router/api_meta.go` that returns `map[string]rbac.APIMeta` with metadata for all RBAC-protected endpoints.

Call `registry.SyncRoutes(ctx, engine, metaMap)` in `serve.go` after `router.Setup()` completes.

### RBAC Handler Missing Methods Detail

#### GetRole
```go
func (h *Handler) GetRole(c *gin.Context) {
    role, err := h.roleRepo.GetByID(c.Request.Context(), c.Param("id"))
    if err != nil { response.Error(c, err); return }
    response.Success(c, role)
}
```

#### ApplyTemplate (replace TODO stub)
```go
func (h *Handler) ApplyTemplate(c *gin.Context) {
    roleID := c.Param("id")
    ctx := c.Request.Context()

    // Verify role exists and is not super
    role, err := h.roleRepo.GetByID(ctx, roleID)
    if err != nil { response.Error(c, err); return }
    if role.BuiltIn && role.Slug == "super" {
        response.Error(c, apperror.Forbidden("cannot modify super role permissions", nil))
        return
    }

    var req ApplyTemplateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, apperror.Validation("invalid request", err))
        return
    }

    // Verify template exists
    if _, err := h.templateRepo.GetByID(ctx, req.TemplateID); err != nil {
        response.Error(c, err)
        return
    }

    // Clone template permissions to role
    if err := h.roleAPIRepo.CloneFromTemplate(ctx, roleID, req.TemplateID); err != nil {
        response.Error(c, err)
        return
    }

    // Invalidate role cache
    if err := h.svc.InvalidateRoleCache(ctx, roleID); err != nil {
        slog.Error("invalidate role cache after template apply", "error", err, "role_id", roleID)
    }

    response.NoContent(c)
}
```

#### ListMenus
```go
func (h *Handler) ListMenus(c *gin.Context) {
    menus, err := h.menuRepo.ListTree(c.Request.Context())
    if err != nil { response.Error(c, err); return }
    response.Success(c, menus)
}
```

---

## 4. Testing Strategy

| Package | Test Type | Key Scenarios |
|---------|-----------|---------------|
| site/service | Unit (mock repo) | create site + schema, delete with confirm, last site guard, slug/domain uniqueness |
| site/handler | HTTP (mock svc) | all 8 endpoints happy path + validation errors + 404/409/422 |
| rbac/handler | HTTP (augment existing) | GetRole, ApplyTemplate (super protection, template clone), ListMenus |
| router | Smoke | verify all new routes respond with correct middleware chain |

---

## 5. Implementation Order

```
Phase 1: Sites Module (Batch 2)
  1. site/interfaces.go + dto.go
  2. site/repository.go
  3. site/service.go
  4. site/handler.go
  5. site/service_test.go + handler_test.go

Phase 2: RBAC Completion (Batch 3)
  6. rbac/handler.go — add GetRole, ApplyTemplate, ListMenus
  7. rbac/handler_test.go — augment with new tests

Phase 3: Router + API Registry
  8. router/api_meta.go — BuildAPIMetaMap()
  9. router/router.go — register sites + rbac route groups
  10. cmd/cms/serve.go — call registry.SyncRoutes() at startup
  11. Full test suite verification (go test ./... + go vet)
```

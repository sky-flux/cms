# 动态 RBAC 权限系统实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将固定 ENUM 角色替换为表驱动的动态 RBAC 系统，支持 API 路由级权限控制、后台菜单可见性管理、权限模板。

**Architecture:** 9 张新 RBAC 表（全部在 public schema），两级 Redis 缓存（L1 用户角色 / L2 角色权限），应用启动时自动从 Gin 路由表同步 API 端点到 `sfc_apis`，RBAC 中间件基于 method+path 动态匹配权限。

**Tech Stack:** Go 1.25+ / uptrace/bun / Gin / Redis / PostgreSQL 18 / log/slog / testify

**设计文档:** `docs/plans/2026-02-24-dynamic-rbac-design.md`

---

## Phase 1: Models & Migrations

### Task 1: Create RBAC Model Structs

**Files:**
- Create: `internal/model/role.go`
- Create: `internal/model/api_endpoint.go`
- Create: `internal/model/admin_menu.go`
- Create: `internal/model/role_template.go`
- Create: `internal/model/user_role.go`

**Step 1: Create Role model**

```go
// internal/model/role.go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Role struct {
	bun.BaseModel `bun:"table:sfc_roles,alias:r"`

	ID          string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Name        string    `bun:"name,notnull,unique" json:"name"`
	Slug        string    `bun:"slug,notnull,unique" json:"slug"`
	Description string    `bun:"description" json:"description,omitempty"`
	BuiltIn     bool      `bun:"built_in,notnull,default:false" json:"built_in"`
	Status      bool      `bun:"status,notnull,default:true" json:"status"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
```

**Step 2: Create APIEndpoint model**

```go
// internal/model/api_endpoint.go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type APIEndpoint struct {
	bun.BaseModel `bun:"table:sfc_apis,alias:api"`

	ID          string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Method      string    `bun:"method,notnull" json:"method"`
	Path        string    `bun:"path,notnull" json:"path"`
	Name        string    `bun:"name,notnull" json:"name"`
	Description string    `bun:"description" json:"description,omitempty"`
	Group       string    `bun:"group,notnull" json:"group"`
	Status      bool      `bun:"status,notnull,default:true" json:"status"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type RoleAPI struct {
	bun.BaseModel `bun:"table:sfc_role_apis"`

	RoleID string `bun:"role_id,pk,type:uuid" json:"role_id"`
	APIID  string `bun:"api_id,pk,type:uuid" json:"api_id"`
}
```

**Step 3: Create AdminMenu model**

```go
// internal/model/admin_menu.go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type AdminMenu struct {
	bun.BaseModel `bun:"table:sfc_menus,alias:m"`

	ID        string       `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	ParentID  *string      `bun:"parent_id,type:uuid" json:"parent_id,omitempty"`
	Name      string       `bun:"name,notnull" json:"name"`
	Icon      string       `bun:"icon" json:"icon,omitempty"`
	Path      string       `bun:"path" json:"path,omitempty"`
	SortOrder int          `bun:"sort_order,notnull,default:0" json:"sort_order"`
	Status    bool         `bun:"status,notnull,default:true" json:"status"`
	CreatedAt time.Time    `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time    `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	Children  []*AdminMenu `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
}

type RoleMenu struct {
	bun.BaseModel `bun:"table:sfc_role_menus"`

	RoleID string `bun:"role_id,pk,type:uuid" json:"role_id"`
	MenuID string `bun:"menu_id,pk,type:uuid" json:"menu_id"`
}
```

**Step 4: Create RoleTemplate model**

```go
// internal/model/role_template.go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type RoleTemplate struct {
	bun.BaseModel `bun:"table:sfc_role_templates,alias:rt"`

	ID          string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Name        string    `bun:"name,notnull,unique" json:"name"`
	Description string    `bun:"description" json:"description,omitempty"`
	BuiltIn     bool      `bun:"built_in,notnull,default:false" json:"built_in"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type RoleTemplateAPI struct {
	bun.BaseModel `bun:"table:sfc_role_template_apis"`

	TemplateID string `bun:"template_id,pk,type:uuid" json:"template_id"`
	APIID      string `bun:"api_id,pk,type:uuid" json:"api_id"`
}

type RoleTemplateMenu struct {
	bun.BaseModel `bun:"table:sfc_role_template_menus"`

	TemplateID string `bun:"template_id,pk,type:uuid" json:"template_id"`
	MenuID     string `bun:"menu_id,pk,type:uuid" json:"menu_id"`
}
```

**Step 5: Create UserRole model (replaces sfc_site_user_roles)**

```go
// internal/model/user_role.go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type UserRole struct {
	bun.BaseModel `bun:"table:sfc_user_roles"`

	UserID    string    `bun:"user_id,pk,type:uuid" json:"user_id"`
	RoleID    string    `bun:"role_id,pk,type:uuid" json:"role_id"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`

	// Relations
	Role *Role `bun:"rel:belongs-to,join:role_id=id" json:"role,omitempty"`
	User *User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}
```

**Step 6: Commit**

```bash
git add internal/model/role.go internal/model/api_endpoint.go internal/model/admin_menu.go internal/model/role_template.go internal/model/user_role.go
git commit -m "feat(model): add RBAC model structs for dynamic permission system"
```

---

### Task 2: Modify Migration 1 — Remove user_role ENUM

**Files:**
- Modify: `migrations/20260224000001_create_enums_and_functions.go`

**Step 1: Read current migration 1 to understand the full ENUM list**

**Step 2: Remove user_role ENUM from up function**

Remove the line:
```sql
CREATE TYPE user_role AS ENUM ('superadmin', 'admin', 'editor', 'viewer');
```

Remove from down function:
```sql
DROP TYPE IF EXISTS user_role;
```

**Step 3: Run migration test (dry run)**

```bash
go build ./cmd/cms
```
Expected: Compiles without errors.

**Step 4: Commit**

```bash
git add migrations/20260224000001_create_enums_and_functions.go
git commit -m "refactor(migration): remove user_role ENUM, replaced by sfc_roles table"
```

---

### Task 3: Modify Migration 2 — Replace sfc_site_user_roles with RBAC Tables

**Files:**
- Modify: `migrations/20260224000002_create_public_schema.go`

**Step 1: Read current migration 2**

**Step 2: Remove sfc_site_user_roles DDL from up function**

Remove all DDL related to `sfc_site_user_roles` (CREATE TABLE, INDEX, TRIGGER).

**Step 3: Add RBAC tables DDL to up function**

Add the following tables in order (respecting FK dependencies):

1. `sfc_roles` — no FK dependencies
2. `sfc_user_roles` — FK to sfc_users, sfc_roles
3. `sfc_apis` — no FK dependencies
4. `sfc_role_apis` — FK to sfc_roles, sfc_apis
5. `sfc_menus` — self-referencing FK
6. `sfc_role_menus` — FK to sfc_roles, sfc_menus
7. `sfc_role_templates` — no FK dependencies
8. `sfc_role_template_apis` — FK to sfc_role_templates, sfc_apis
9. `sfc_role_template_menus` — FK to sfc_role_templates, sfc_menus

Use the DDL from `docs/plans/2026-02-24-dynamic-rbac-design.md` Section 2.2.

**Step 4: Update down function**

Add DROP TABLE statements in reverse order:
```sql
DROP TABLE IF EXISTS public.sfc_role_template_menus;
DROP TABLE IF EXISTS public.sfc_role_template_apis;
DROP TABLE IF EXISTS public.sfc_role_templates;
DROP TABLE IF EXISTS public.sfc_role_menus;
DROP TABLE IF EXISTS public.sfc_menus;
DROP TABLE IF EXISTS public.sfc_role_apis;
DROP TABLE IF EXISTS public.sfc_apis;
DROP TABLE IF EXISTS public.sfc_user_roles;
DROP TABLE IF EXISTS public.sfc_roles;
```

Remove the old:
```sql
DROP TABLE IF EXISTS public.sfc_site_user_roles;
```

**Step 5: Build to verify**

```bash
go build ./cmd/cms
```
Expected: Compiles without errors.

**Step 6: Commit**

```bash
git add migrations/20260224000002_create_public_schema.go
git commit -m "feat(migration): replace sfc_site_user_roles with 9 RBAC tables"
```

---

### Task 4: Add Seed Migration for Built-in Roles & Templates

**Files:**
- Create: `migrations/20260224000004_seed_rbac_builtins.go`

**Step 1: Create seed migration file**

```go
package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Seed 4 built-in roles
		_, err := db.ExecContext(ctx, `
			INSERT INTO public.sfc_roles (name, slug, description, built_in, status) VALUES
			('超级管理员', 'super', '拥有所有权限，不可修改/删除', true, true),
			('管理员', 'admin', '站点管理，不可删除', true, true),
			('编辑', 'editor', '内容创建与编辑，不可删除', true, true),
			('查看者', 'viewer', '只读访问，不可删除', true, true)
			ON CONFLICT (slug) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("seed built-in roles: %w", err)
		}

		// Seed 4 built-in role templates
		_, err = db.ExecContext(ctx, `
			INSERT INTO public.sfc_role_templates (name, description, built_in) VALUES
			('超级管理员模板', '预置超级管理员权限集', true),
			('管理员模板', '预置管理员权限集', true),
			('编辑模板', '预置编辑权限集', true),
			('查看者模板', '预置查看者权限集', true)
			ON CONFLICT (name) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("seed built-in templates: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
			DELETE FROM public.sfc_role_templates WHERE built_in = true;
			DELETE FROM public.sfc_roles WHERE built_in = true;
		`)
		if err != nil {
			return fmt.Errorf("rollback built-in seeds: %w", err)
		}
		return nil
	})
}
```

**Step 2: Build to verify**

```bash
go build ./cmd/cms
```

**Step 3: Commit**

```bash
git add migrations/20260224000004_seed_rbac_builtins.go
git commit -m "feat(migration): seed built-in roles and templates"
```

---

### Task 5: Update Schema Template — Remove site_user_roles References

**Files:**
- Modify: `internal/schema/template.go`

**Step 1: Read current template.go**

**Step 2: Check if sfc_site_user_roles is referenced in the site template DDL**

Since `sfc_site_user_roles` was in public schema (not site schema), the template likely doesn't reference it. Verify and remove any references if found.

**Step 3: Commit (if changes needed)**

```bash
git add internal/schema/template.go
git commit -m "refactor(schema): remove site_user_roles references from site template"
```

---

## Phase 2: Repository Layer

### Task 6: Role Repository

**Files:**
- Create: `internal/rbac/role_repo.go`
- Create: `internal/rbac/role_repo_test.go`

**Step 1: Write repository interface and implementation**

```go
// internal/rbac/role_repo.go
package rbac

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

type RoleRepo struct {
	db *bun.DB
}

func NewRoleRepo(db *bun.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

func (r *RoleRepo) List(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.NewSelect().Model(&roles).
		OrderExpr("built_in DESC, created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	return roles, nil
}

func (r *RoleRepo) GetByID(ctx context.Context, id string) (*model.Role, error) {
	role := new(model.Role)
	err := r.db.NewSelect().Model(role).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, apperror.NotFound("role not found", err)
	}
	return role, nil
}

func (r *RoleRepo) GetBySlug(ctx context.Context, slug string) (*model.Role, error) {
	role := new(model.Role)
	err := r.db.NewSelect().Model(role).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		return nil, apperror.NotFound("role not found", err)
	}
	return role, nil
}

func (r *RoleRepo) Create(ctx context.Context, role *model.Role) error {
	_, err := r.db.NewInsert().Model(role).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create role: %w", err)
	}
	return nil
}

func (r *RoleRepo) Update(ctx context.Context, role *model.Role) error {
	_, err := r.db.NewUpdate().Model(role).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

func (r *RoleRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.Role)(nil)).Where("id = ? AND built_in = false", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}
```

**Step 2: Write failing tests**

```go
// internal/rbac/role_repo_test.go
package rbac_test

// Tests use testcontainers-go for real PostgreSQL.
// Test: Create → GetByID → GetBySlug → List → Update → Delete
// Test: Delete built-in role returns error / no rows affected
// Test: Create duplicate slug returns conflict error
```

**Step 3: Run tests to verify they fail**

```bash
go test ./internal/rbac/ -run TestRoleRepo -v
```

**Step 4: Fix any compilation issues, run tests to pass**

**Step 5: Commit**

```bash
git add internal/rbac/
git commit -m "feat(rbac): add role repository with CRUD operations"
```

---

### Task 7: API Endpoint Repository

**Files:**
- Create: `internal/rbac/api_repo.go`
- Create: `internal/rbac/api_repo_test.go`

**Step 1: Write repository**

```go
// internal/rbac/api_repo.go
package rbac

// Methods:
// - UpsertBatch(ctx, []model.APIEndpoint) error      — 批量 upsert（自动注册用）
// - DisableStale(ctx, activeMethodPaths []string) error — 标记不在当前路由表中的为 status=false
// - List(ctx) ([]model.APIEndpoint, error)             — 列出所有 API
// - ListByGroup(ctx, group string) ([]model.APIEndpoint, error)
// - GetByMethodPath(ctx, method, path string) (*model.APIEndpoint, error)
```

**Step 2: Write tests**

**Step 3: Run tests, make them pass**

**Step 4: Commit**

```bash
git add internal/rbac/api_repo.go internal/rbac/api_repo_test.go
git commit -m "feat(rbac): add API endpoint repository with upsert batch"
```

---

### Task 8: Role-API Permission Repository

**Files:**
- Create: `internal/rbac/role_api_repo.go`
- Create: `internal/rbac/role_api_repo_test.go`

**Step 1: Write repository**

```go
// internal/rbac/role_api_repo.go
package rbac

// Methods:
// - GetAPIsByRoleID(ctx, roleID string) ([]model.APIEndpoint, error)
// - SetRoleAPIs(ctx, roleID string, apiIDs []string) error  — 全量替换
// - GetRoleIDsByMethodPath(ctx, method, path string) ([]string, error)  — 哪些角色能访问此 API
// - CloneFromTemplate(ctx, roleID, templateID string) error  — 从模板复制
```

**Step 2: Write tests**

**Step 3: Run tests, make them pass**

**Step 4: Commit**

```bash
git add internal/rbac/role_api_repo.go internal/rbac/role_api_repo_test.go
git commit -m "feat(rbac): add role-API permission repository"
```

---

### Task 9: Admin Menu Repository

**Files:**
- Create: `internal/rbac/menu_repo.go`
- Create: `internal/rbac/menu_repo_test.go`

**Step 1: Write repository**

```go
// internal/rbac/menu_repo.go
package rbac

// Methods:
// - ListTree(ctx) ([]model.AdminMenu, error)           — 树形结构
// - Create(ctx, menu *model.AdminMenu) error
// - Update(ctx, menu *model.AdminMenu) error
// - Delete(ctx, id string) error
// - GetMenusByRoleID(ctx, roleID string) ([]model.AdminMenu, error)
// - SetRoleMenus(ctx, roleID string, menuIDs []string) error
// - GetMenusByUserID(ctx, userID string) ([]model.AdminMenu, error)  — 合并用户所有角色的菜单
```

**Step 2: Write tests**

**Step 3: Run tests, make them pass**

**Step 4: Commit**

```bash
git add internal/rbac/menu_repo.go internal/rbac/menu_repo_test.go
git commit -m "feat(rbac): add admin menu repository with tree query"
```

---

### Task 10: Template Repository

**Files:**
- Create: `internal/rbac/template_repo.go`
- Create: `internal/rbac/template_repo_test.go`

**Step 1: Write repository**

```go
// internal/rbac/template_repo.go
package rbac

// Methods:
// - List(ctx) ([]model.RoleTemplate, error)
// - GetByID(ctx, id string) (*model.RoleTemplate, error)
// - Create(ctx, tmpl *model.RoleTemplate) error
// - Update(ctx, tmpl *model.RoleTemplate) error
// - Delete(ctx, id string) error  — built_in 不可删
// - GetTemplateAPIs(ctx, templateID string) ([]model.APIEndpoint, error)
// - SetTemplateAPIs(ctx, templateID string, apiIDs []string) error
// - GetTemplateMenus(ctx, templateID string) ([]model.AdminMenu, error)
// - SetTemplateMenus(ctx, templateID string, menuIDs []string) error
```

**Step 2: Write tests**

**Step 3: Run tests, make them pass**

**Step 4: Commit**

```bash
git add internal/rbac/template_repo.go internal/rbac/template_repo_test.go
git commit -m "feat(rbac): add role template repository"
```

---

### Task 11: User-Role Repository

**Files:**
- Create: `internal/rbac/user_role_repo.go`
- Create: `internal/rbac/user_role_repo_test.go`

**Step 1: Write repository**

```go
// internal/rbac/user_role_repo.go
package rbac

// Methods:
// - GetRolesByUserID(ctx, userID string) ([]model.Role, error)
// - GetRoleSlugs(ctx, userID string) ([]string, error)     — 只取 slug 列表，性能优化
// - SetUserRoles(ctx, userID string, roleIDs []string) error  — 全量替换
// - HasRole(ctx, userID, roleSlug string) (bool, error)
```

**Step 2: Write tests**

**Step 3: Run tests, make them pass**

**Step 4: Commit**

```bash
git add internal/rbac/user_role_repo.go internal/rbac/user_role_repo_test.go
git commit -m "feat(rbac): add user-role repository"
```

---

## Phase 3: Service & Caching

### Task 12: RBAC Service with Two-Level Redis Cache

**Files:**
- Create: `internal/rbac/service.go`
- Create: `internal/rbac/service_test.go`

**Step 1: Write service with caching logic**

```go
// internal/rbac/service.go
package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/model"
)

const (
	userRolesTTL = 300 * time.Second  // L1: 5 minutes
	roleAPIsTTL  = 600 * time.Second  // L2: 10 minutes
	userMenusTTL = 300 * time.Second
)

type Service struct {
	roleRepo     *RoleRepo
	apiRepo      *APIRepo
	roleAPIRepo  *RoleAPIRepo
	menuRepo     *MenuRepo
	templateRepo *TemplateRepo
	userRoleRepo *UserRoleRepo
	rdb          *redis.Client
}

// CheckPermission verifies if user can access method+path.
// Flow:
//   1. Get user's role slugs (L1 cache or DB)
//   2. If "super" in slugs → return true (short circuit)
//   3. For each role, get API set (L2 cache or DB)
//   4. Union all API sets, check if method:path matches
func (s *Service) CheckPermission(ctx context.Context, userID, method, path string) (bool, error) {
	// Implementation
	return false, nil
}

// GetUserMenuTree returns merged menu tree for user's roles.
func (s *Service) GetUserMenuTree(ctx context.Context, userID string) ([]model.AdminMenu, error) {
	return nil, nil
}

// InvalidateUserCache clears L1 cache for a user (on role change).
func (s *Service) InvalidateUserCache(ctx context.Context, userID string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("user:%s:role_ids", userID)).Err()
}

// InvalidateRoleCache clears L2 cache for a role (on permission change).
func (s *Service) InvalidateRoleCache(ctx context.Context, roleID string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("role:%s:api_set", roleID)).Err()
}
```

Key cache patterns:
- L1 key: `user:{userID}:role_ids` → JSON array of role slugs, TTL 300s
- L2 key: `role:{roleID}:api_set` → JSON array of `"METHOD:/path"` strings, TTL 600s
- Menu key: `user:{userID}:menu_ids` → JSON array of menu IDs, TTL 300s

**Step 2: Write tests using miniredis for Redis mocking**

Test cases:
- CheckPermission with super role → always true
- CheckPermission with cached roles → no DB hit
- CheckPermission with uncached roles → DB query + cache set
- InvalidateUserCache → L1 key deleted
- InvalidateRoleCache → L2 key deleted
- GetUserMenuTree → merged tree from multiple roles

**Step 3: Run tests, make them pass**

**Step 4: Commit**

```bash
git add internal/rbac/service.go internal/rbac/service_test.go
git commit -m "feat(rbac): add RBAC service with two-level Redis cache"
```

---

### Task 13: API Auto-Registration Service

**Files:**
- Create: `internal/rbac/api_registry.go`
- Create: `internal/rbac/api_registry_test.go`

**Step 1: Define route metadata struct**

```go
// internal/rbac/api_registry.go
package rbac

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
)

// APIMeta holds metadata for route registration.
// Attached to Gin routes via route group conventions.
type APIMeta struct {
	Name        string
	Description string
	Group       string
}

// Registry scans Gin routes and syncs to sfc_apis.
type Registry struct {
	apiRepo *APIRepo
}

func NewRegistry(apiRepo *APIRepo) *Registry {
	return &Registry{apiRepo: apiRepo}
}

// SyncRoutes reads all registered Gin routes and upserts to sfc_apis.
// Called once at application startup after all routes are registered.
func (r *Registry) SyncRoutes(ctx context.Context, engine *gin.Engine, metaMap map[string]APIMeta) error {
	routes := engine.Routes()

	var endpoints []model.APIEndpoint
	var activeKeys []string

	for _, route := range routes {
		key := route.Method + ":" + route.Path
		meta, ok := metaMap[key]
		if !ok {
			// Skip routes without metadata (public, health, etc.)
			continue
		}

		endpoints = append(endpoints, model.APIEndpoint{
			Method:      route.Method,
			Path:        route.Path,
			Name:        meta.Name,
			Description: meta.Description,
			Group:       meta.Group,
			Status:      true,
		})
		activeKeys = append(activeKeys, key)
	}

	if err := r.apiRepo.UpsertBatch(ctx, endpoints); err != nil {
		return fmt.Errorf("upsert api endpoints: %w", err)
	}

	if err := r.apiRepo.DisableStale(ctx, activeKeys); err != nil {
		return fmt.Errorf("disable stale endpoints: %w", err)
	}

	slog.Info("api registry synced", "total", len(endpoints))
	return nil
}
```

**Step 2: Write tests**

Test cases:
- SyncRoutes with new routes → inserted
- SyncRoutes with existing routes → updated
- SyncRoutes removes stale routes → status=false
- Routes without metadata are skipped

**Step 3: Run tests, make them pass**

**Step 4: Commit**

```bash
git add internal/rbac/api_registry.go internal/rbac/api_registry_test.go
git commit -m "feat(rbac): add API auto-registration from Gin routes"
```

---

## Phase 4: Middleware

### Task 14: RBAC Middleware

**Files:**
- Modify: `internal/middleware/rbac.go` (currently a stub)
- Create: `internal/middleware/rbac_test.go`

**Step 1: Write failing test**

```go
// internal/middleware/rbac_test.go
package middleware_test

// Test cases:
// - Request with super role → 200
// - Request with matching permission → 200
// - Request without matching permission → 403
// - Request without auth (no user_id in context) → 401
// - Disabled role → 403
```

**Step 2: Implement RBAC middleware**

```go
// internal/middleware/rbac.go
package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/rbac"
)

// RBAC returns middleware that checks API-level permissions.
// It reads user_id from Gin context (set by JWT middleware),
// then delegates to rbac.Service.CheckPermission().
func RBAC(svc *rbac.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
			})
			return
		}

		method := c.Request.Method
		path := c.FullPath() // Gin route pattern, e.g. /api/v1/posts/:id

		allowed, err := svc.CheckPermission(c.Request.Context(), userID, method, path)
		if err != nil {
			slog.Error("rbac check failed",
				"error", err,
				"user_id", userID,
				"method", method,
				"path", path,
			)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "permission check failed",
			})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "forbidden",
			})
			return
		}

		c.Next()
	}
}
```

**Step 3: Run tests, make them pass**

```bash
go test ./internal/middleware/ -run TestRBAC -v
```

**Step 4: Commit**

```bash
git add internal/middleware/rbac.go internal/middleware/rbac_test.go
git commit -m "feat(middleware): implement dynamic RBAC permission checking"
```

---

## Phase 5: Handlers & Routes

### Task 15: RBAC DTOs

**Files:**
- Create: `internal/rbac/dto.go`

**Step 1: Define request/response DTOs**

```go
// internal/rbac/dto.go
package rbac

// --- Role ---
type CreateRoleReq struct {
	Name        string `json:"name" binding:"required,max=50"`
	Slug        string `json:"slug" binding:"required,max=50,lowercase"`
	Description string `json:"description"`
}

type UpdateRoleReq struct {
	Name        string `json:"name" binding:"max=50"`
	Description string `json:"description"`
	Status      *bool  `json:"status"`
}

// --- Role-API Permission ---
type SetRoleAPIsReq struct {
	APIIDs []string `json:"api_ids" binding:"required"`
}

// --- Role-Menu Permission ---
type SetRoleMenusReq struct {
	MenuIDs []string `json:"menu_ids" binding:"required"`
}

// --- User-Role Assignment ---
type SetUserRolesReq struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}

// --- Template ---
type CreateTemplateReq struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
}

type UpdateTemplateReq struct {
	Name        string `json:"name" binding:"max=100"`
	Description string `json:"description"`
}

type SetTemplateAPIsReq struct {
	APIIDs []string `json:"api_ids" binding:"required"`
}

type SetTemplateMenusReq struct {
	MenuIDs []string `json:"menu_ids" binding:"required"`
}

// --- Apply Template ---
type ApplyTemplateReq struct {
	TemplateID string `json:"template_id" binding:"required,uuid"`
}
```

**Step 2: Commit**

```bash
git add internal/rbac/dto.go
git commit -m "feat(rbac): add request/response DTOs"
```

---

### Task 16: RBAC Handlers

**Files:**
- Create: `internal/rbac/handler.go`
- Create: `internal/rbac/handler_test.go`

**Step 1: Write handler struct with all endpoint methods**

```go
// internal/rbac/handler.go
package rbac

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// --- Role CRUD ---
func (h *Handler) ListRoles(c *gin.Context)  { /* ... */ }
func (h *Handler) CreateRole(c *gin.Context) { /* ... */ }
func (h *Handler) UpdateRole(c *gin.Context) { /* ... */ }
func (h *Handler) DeleteRole(c *gin.Context) { /* ... */ }

// --- Role-API Permissions ---
func (h *Handler) GetRoleAPIs(c *gin.Context) { /* ... */ }
func (h *Handler) SetRoleAPIs(c *gin.Context) { /* ... */ }

// --- Role-Menu Permissions ---
func (h *Handler) GetRoleMenus(c *gin.Context) { /* ... */ }
func (h *Handler) SetRoleMenus(c *gin.Context) { /* ... */ }

// --- Templates ---
func (h *Handler) ListTemplates(c *gin.Context)  { /* ... */ }
func (h *Handler) CreateTemplate(c *gin.Context) { /* ... */ }
func (h *Handler) UpdateTemplate(c *gin.Context) { /* ... */ }
func (h *Handler) DeleteTemplate(c *gin.Context) { /* ... */ }
func (h *Handler) ApplyTemplate(c *gin.Context)  { /* ... */ }

// --- User Roles ---
func (h *Handler) GetUserRoles(c *gin.Context) { /* ... */ }
func (h *Handler) SetUserRoles(c *gin.Context) { /* ... */ }

// --- Current User Menu ---
func (h *Handler) GetMyMenus(c *gin.Context) { /* ... */ }

// --- API Registry ---
func (h *Handler) ListAPIs(c *gin.Context) { /* ... */ }
```

Handler patterns follow existing codebase:
- Use `response.Success()`, `response.Created()`, `response.Error()`
- Use `binding:"required"` for validation
- Use `c.Param("id")` for path params
- Call `svc.InvalidateUserCache()` / `svc.InvalidateRoleCache()` after mutations

**Step 2: Write handler tests using httptest**

**Step 3: Run tests, make them pass**

**Step 4: Commit**

```bash
git add internal/rbac/handler.go internal/rbac/handler_test.go
git commit -m "feat(rbac): add HTTP handlers for role management"
```

---

### Task 17: Register RBAC Routes with Metadata

**Files:**
- Modify: `internal/router/router.go`

**Step 1: Read current router.go**

**Step 2: Add RBAC route group with metadata map**

```go
// In Setup() function, after existing middleware registration:

// Build RBAC service dependencies
rbacRepo := rbac.NewRoleRepo(db)
// ... other repos ...
rbacSvc := rbac.NewService(/* repos + redis */)
rbacHandler := rbac.NewHandler(rbacSvc)

v1 := engine.Group("/api/v1")

// RBAC management routes (protected by Auth + RBAC middleware)
roles := v1.Group("/roles")
{
    roles.GET("", rbacHandler.ListRoles)
    roles.POST("", rbacHandler.CreateRole)
    roles.PUT("/:id", rbacHandler.UpdateRole)
    roles.DELETE("/:id", rbacHandler.DeleteRole)
    roles.GET("/:id/apis", rbacHandler.GetRoleAPIs)
    roles.PUT("/:id/apis", rbacHandler.SetRoleAPIs)
    roles.GET("/:id/menus", rbacHandler.GetRoleMenus)
    roles.PUT("/:id/menus", rbacHandler.SetRoleMenus)
    roles.POST("/:id/apply-template", rbacHandler.ApplyTemplate)
}

templates := v1.Group("/role-templates")
{
    templates.GET("", rbacHandler.ListTemplates)
    templates.POST("", rbacHandler.CreateTemplate)
    templates.PUT("/:id", rbacHandler.UpdateTemplate)
    templates.DELETE("/:id", rbacHandler.DeleteTemplate)
}

users := v1.Group("/users")
{
    users.GET("/:id/roles", rbacHandler.GetUserRoles)
    users.PUT("/:id/roles", rbacHandler.SetUserRoles)
}

v1.GET("/user/menus", rbacHandler.GetMyMenus)
v1.GET("/apis", rbacHandler.ListAPIs)
```

**Step 3: Define API metadata map for auto-registration**

```go
// internal/router/api_meta.go
package router

import "github.com/sky-flux/cms/internal/rbac"

// APIMetaMap maps "METHOD:/path" to metadata for sfc_apis auto-registration.
var APIMetaMap = map[string]rbac.APIMeta{
    "GET:/api/v1/roles":                    {Name: "角色列表", Group: "角色管理"},
    "POST:/api/v1/roles":                   {Name: "创建角色", Group: "角色管理"},
    "PUT:/api/v1/roles/:id":                {Name: "更新角色", Group: "角色管理"},
    "DELETE:/api/v1/roles/:id":             {Name: "删除角色", Group: "角色管理"},
    "GET:/api/v1/roles/:id/apis":           {Name: "获取角色API权限", Group: "角色管理"},
    "PUT:/api/v1/roles/:id/apis":           {Name: "设置角色API权限", Group: "角色管理"},
    "GET:/api/v1/roles/:id/menus":          {Name: "获取角色菜单权限", Group: "角色管理"},
    "PUT:/api/v1/roles/:id/menus":          {Name: "设置角色菜单权限", Group: "角色管理"},
    "POST:/api/v1/roles/:id/apply-template": {Name: "应用模板到角色", Group: "角色管理"},
    "GET:/api/v1/role-templates":            {Name: "模板列表", Group: "模板管理"},
    "POST:/api/v1/role-templates":           {Name: "创建模板", Group: "模板管理"},
    "PUT:/api/v1/role-templates/:id":        {Name: "更新模板", Group: "模板管理"},
    "DELETE:/api/v1/role-templates/:id":     {Name: "删除模板", Group: "模板管理"},
    "GET:/api/v1/users/:id/roles":           {Name: "获取用户角色", Group: "用户管理"},
    "PUT:/api/v1/users/:id/roles":           {Name: "设置用户角色", Group: "用户管理"},
    "GET:/api/v1/user/menus":                {Name: "我的菜单", Group: "个人"},
    "GET:/api/v1/apis":                      {Name: "API列表", Group: "角色管理"},
}
```

**Step 4: Call Registry.SyncRoutes in serve.go startup**

In `cmd/cms/serve.go`, after `router.Setup()` and before `engine.Run()`:

```go
// Auto-register API endpoints to sfc_apis
registry := rbac.NewRegistry(apiRepo)
if err := registry.SyncRoutes(ctx, engine, router.APIMetaMap); err != nil {
    slog.Error("api registry sync failed", "error", err)
}
```

**Step 5: Build and verify**

```bash
go build ./cmd/cms
```

**Step 6: Commit**

```bash
git add internal/router/router.go internal/router/api_meta.go cmd/cms/serve.go
git commit -m "feat(router): register RBAC routes and API auto-registration"
```

---

## Phase 6: Documentation Updates

### Task 18: Update database.md

**Files:**
- Modify: `docs/database.md`

**Changes:**
1. ER diagram: Remove `sfc_site_user_roles`, add 9 RBAC tables with relationships
2. Section 2A DDL: Remove `user_role` ENUM, remove `sfc_site_user_roles` DDL, add 9 RBAC table DDLs
3. Redis key space: Replace `site:{slug}:role:{user_id}` with `user:{user_id}:role_ids` + `role:{role_id}:api_set` + `user:{user_id}:menu_ids`
4. Migration file listing: Update to reflect new/modified migrations

**Commit:**

```bash
git add docs/database.md
git commit -m "docs(database): update schema for dynamic RBAC tables"
```

---

### Task 19: Update prd.md, api.md, security.md, story.md, architecture.md

**Files:**
- Modify: `docs/prd.md` — Permission matrix: "4 fixed roles" → "4 built-in + custom roles", add role management feature description
- Modify: `docs/api.md` — Add RBAC management API group (17 endpoints), update middleware from `RequireRole()` to dynamic matching
- Modify: `docs/security.md` — RBAC architecture from ENUM to dynamic tables, two-level cache strategy
- Modify: `docs/story.md` — Add user stories: role CRUD, permission assignment, template management
- Modify: `docs/architecture.md` — Middleware chain reorder (RBAC before SiteResolver)

**Commit per file or batch:**

```bash
git add docs/prd.md docs/api.md docs/security.md docs/story.md docs/architecture.md
git commit -m "docs: update all design docs for dynamic RBAC system"
```

---

### Task 20: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

**Changes:**
1. Key decisions table: Add RBAC row (`动态 RBAC | sfc_roles + sfc_apis 表驱动 | ENUM 固定角色`)
2. Multi-site section: Remove `sfc_site_user_roles` reference, add `sfc_user_roles` (global)
3. Middleware chain: Update to show RBAC before SiteResolver
4. Directory structure: Add `internal/rbac/` entry

**Commit:**

```bash
git add CLAUDE.md
git commit -m "docs(claude): update project instructions for dynamic RBAC"
```

---

## Dependency Graph

```
Task 1 (Models) ──→ Task 2,3,4 (Migrations) ──→ Task 5 (Schema Template)
                                                       │
Task 1 ──→ Task 6-11 (Repositories) ──→ Task 12 (Service) ──→ Task 14 (Middleware)
                                         │                      │
                                         └→ Task 13 (Registry)  │
                                                                 │
           Task 15 (DTOs) ──→ Task 16 (Handlers) ──→ Task 17 (Router) ──→ Task 18-20 (Docs)
```

## Verification Checklist

- [ ] `go build ./cmd/cms` — compiles without errors
- [ ] `go test ./internal/rbac/... -v` — all repository/service/handler tests pass
- [ ] `go test ./internal/middleware/... -v` — RBAC middleware tests pass
- [ ] `go run ./cmd/cms migrate up` — migrations execute successfully (with Docker PG)
- [ ] `go run ./cmd/cms migrate down` — rollback works cleanly
- [ ] `go run ./cmd/cms serve` — starts without errors, API registry logs endpoint count
- [ ] Manual test: super role user can access all endpoints
- [ ] Manual test: role without permission gets 403
- [ ] All 6 docs updated and internally consistent

# Batch 4 Implementation Plan — Infrastructure + Simple CRUD

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build site-scoped routing infrastructure, audit/mail services, and 5 business modules (15 endpoints) — the foundation for all future site-scoped work.

**Architecture:** Boolean→Smallint migration first, then infrastructure (audit/mail/middleware), then CRUD modules. Users module on global route group; Settings/API Keys/Post Types/Audit Logs on site-scoped route group (SiteResolver → Schema → Auth → RBAC).

**Tech Stack:** Go / Gin / uptrace/bun / PostgreSQL 18 / Redis / Resend SDK v3

**Design Doc:** `docs/plans/2026-02-24-site-scoped-batch4-design.md`

**Existing Patterns:** Follow `internal/site/` (handler/service/repo/dto/interfaces) and `internal/auth/` for conventions.

---

## Task 1: Add New Enum Types to enums.go

**Files:**
- Modify: `internal/model/enums.go`

**Step 1: Add Toggle and Status enums**

Append to `internal/model/enums.go` after existing enums:

```go
// Toggle is a generic binary enum for fields like built_in, revoked, enabled, primary, pinned.
type Toggle int8

const (
	ToggleNo  Toggle = iota + 1 // 1
	ToggleYes                    // 2
)

// UserStatus maps to sfc_users.status (SMALLINT)
type UserStatus int8

const (
	UserStatusActive   UserStatus = iota + 1 // 1
	UserStatusDisabled                        // 2
)

// SiteStatus maps to sfc_sites.status (SMALLINT)
type SiteStatus int8

const (
	SiteStatusActive   SiteStatus = iota + 1 // 1
	SiteStatusDisabled                        // 2
)

// RoleStatus maps to sfc_roles.status (SMALLINT)
type RoleStatus int8

const (
	RoleStatusActive   RoleStatus = iota + 1 // 1
	RoleStatusDisabled                        // 2
)

// APIStatus maps to sfc_apis.status (SMALLINT)
type APIStatus int8

const (
	APIStatusActive   APIStatus = iota + 1 // 1
	APIStatusDisabled                       // 2
)

// MenuStatus maps to sfc_menus.status (SMALLINT) — admin menus
type MenuStatus int8

const (
	MenuStatusActive MenuStatus = iota + 1 // 1
	MenuStatusHidden                        // 2
)

// APIKeyStatus maps to sfc_site_api_keys.status (SMALLINT)
type APIKeyStatus int8

const (
	APIKeyStatusActive  APIKeyStatus = iota + 1 // 1
	APIKeyStatusRevoked                          // 2
)

// RedirectStatus maps to sfc_site_redirects.status (SMALLINT)
type RedirectStatus int8

const (
	RedirectStatusActive   RedirectStatus = iota + 1 // 1
	RedirectStatusDisabled                            // 2
)

// MenuItemStatus maps to sfc_site_menu_items.status (SMALLINT)
type MenuItemStatus int8

const (
	MenuItemStatusActive MenuItemStatus = iota + 1 // 1
	MenuItemStatusHidden                            // 2
)
```

**Step 2: Verify compilation**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/model/enums.go
git commit -m "feat(model): add Toggle and Status enum types for boolean-to-smallint migration"
```

---

## Task 2: Update All Models — Boolean to Smallint

**Files:**
- Modify: `internal/model/user.go`
- Modify: `internal/model/site.go`
- Modify: `internal/model/role.go`
- Modify: `internal/model/api_endpoint.go`
- Modify: `internal/model/admin_menu.go`
- Modify: `internal/model/api_key.go`
- Modify: `internal/model/redirect.go`
- Modify: `internal/model/menu.go` (MenuItem)
- Modify: `internal/model/post_type.go`
- Modify: `internal/model/role_template.go`
- Modify: `internal/model/refresh_token.go`
- Modify: `internal/model/user_totp.go`
- Modify: `internal/model/post.go` (PostCategoryMap)
- Modify: `internal/model/comment.go`

**Step 1: Update each model field**

For each file, replace the boolean field with the appropriate enum type:

**user.go** — `IsActive bool` → `Status UserStatus`:
```go
Status UserStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
```
Remove old `IsActive` field.

**site.go** — `IsActive bool` → `Status SiteStatus`:
```go
Status SiteStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
```

**role.go** — two fields:
```go
BuiltIn Toggle     `bun:"built_in,notnull,type:smallint,default:1" json:"built_in"`
Status  RoleStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
```
Note: `built_in` default is `1` (ToggleNo) because most roles are not built-in.

**api_endpoint.go**:
```go
Status APIStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
```

**admin_menu.go**:
```go
Status MenuStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
```

**api_key.go** — `IsActive bool` → `Status APIKeyStatus`:
```go
Status APIKeyStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
```

**redirect.go** — `IsActive bool` → `Status RedirectStatus`:
```go
Status RedirectStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
```

**menu.go** (MenuItem struct) — `IsActive bool` → `Status MenuItemStatus`:
```go
Status MenuItemStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
```

**post_type.go**:
```go
BuiltIn Toggle `bun:"built_in,notnull,type:smallint,default:1" json:"built_in"`
```

**role_template.go**:
```go
BuiltIn Toggle `bun:"built_in,notnull,type:smallint,default:1" json:"built_in"`
```

**refresh_token.go** — `Revoked bool` → `Revoked Toggle`:
```go
Revoked Toggle `bun:"revoked,notnull,type:smallint,default:1" json:"revoked"`
```
Default 1 = ToggleNo (not revoked).

**user_totp.go** — `IsEnabled bool` → `Enabled Toggle`:
```go
Enabled Toggle `bun:"enabled,notnull,type:smallint,default:1" json:"enabled"`
```
Default 1 = ToggleNo (not enabled).

**post.go** (PostCategoryMap) — `IsPrimary bool` → `Primary Toggle`:
```go
Primary Toggle `bun:"primary,notnull,type:smallint,default:1" json:"primary"`
```

**comment.go** — `IsPinned bool` → `Pinned Toggle`:
```go
Pinned Toggle `bun:"pinned,notnull,type:smallint,default:1" json:"pinned"`
```

**Step 2: Verify compilation**

Run: `go build ./internal/model/...`
Expected: FAIL — dependent packages reference old field names. That's expected, we fix them in Tasks 3-5.

**Step 3: Commit**

```bash
git add internal/model/
git commit -m "refactor(model): convert all boolean fields to smallint enums"
```

---

## Task 3: Migration 5 — Boolean to Smallint in Database

**Files:**
- Create: `migrations/20260224000005_boolean_to_smallint.go`
- Modify: `internal/schema/template.go`

**Step 1: Create migration**

```go
package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// ── public schema ──────────────────────────────────────

		// sfc_users: is_active BOOLEAN → status SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_users
				ADD COLUMN status SMALLINT NOT NULL DEFAULT 1;
			UPDATE public.sfc_users SET status = CASE WHEN is_active THEN 1 ELSE 2 END;
			ALTER TABLE public.sfc_users DROP COLUMN is_active;
		`); err != nil {
			return fmt.Errorf("migrate sfc_users: %w", err)
		}

		// sfc_sites: is_active BOOLEAN → status SMALLINT
		if _, err := db.ExecContext(ctx, `
			DROP INDEX IF EXISTS public.idx_sfc_sites_active;
			ALTER TABLE public.sfc_sites
				ADD COLUMN status SMALLINT NOT NULL DEFAULT 1;
			UPDATE public.sfc_sites SET status = CASE WHEN is_active THEN 1 ELSE 2 END;
			ALTER TABLE public.sfc_sites DROP COLUMN is_active;
			CREATE INDEX idx_sfc_sites_status ON public.sfc_sites(status) WHERE status = 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_sites: %w", err)
		}

		// sfc_roles: built_in BOOLEAN → SMALLINT, status BOOLEAN → SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_roles
				ALTER COLUMN built_in TYPE SMALLINT USING CASE WHEN built_in THEN 2 ELSE 1 END,
				ALTER COLUMN built_in SET DEFAULT 1,
				ALTER COLUMN status TYPE SMALLINT USING CASE WHEN status THEN 1 ELSE 2 END,
				ALTER COLUMN status SET DEFAULT 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_roles: %w", err)
		}

		// sfc_apis: status BOOLEAN → SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_apis
				ALTER COLUMN status TYPE SMALLINT USING CASE WHEN status THEN 1 ELSE 2 END,
				ALTER COLUMN status SET DEFAULT 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_apis: %w", err)
		}

		// sfc_menus (admin): status BOOLEAN → SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_menus
				ALTER COLUMN status TYPE SMALLINT USING CASE WHEN status THEN 1 ELSE 2 END,
				ALTER COLUMN status SET DEFAULT 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_menus: %w", err)
		}

		// sfc_role_templates: built_in BOOLEAN → SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_role_templates
				ALTER COLUMN built_in TYPE SMALLINT USING CASE WHEN built_in THEN 2 ELSE 1 END,
				ALTER COLUMN built_in SET DEFAULT 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_role_templates: %w", err)
		}

		// sfc_refresh_tokens: revoked BOOLEAN → SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_refresh_tokens
				ALTER COLUMN revoked TYPE SMALLINT USING CASE WHEN revoked THEN 2 ELSE 1 END,
				ALTER COLUMN revoked SET DEFAULT 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_refresh_tokens: %w", err)
		}

		// sfc_user_totp: is_enabled BOOLEAN → enabled SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_user_totp
				ADD COLUMN enabled SMALLINT NOT NULL DEFAULT 1;
			UPDATE public.sfc_user_totp SET enabled = CASE WHEN is_enabled THEN 2 ELSE 1 END;
			ALTER TABLE public.sfc_user_totp DROP COLUMN is_enabled;
		`); err != nil {
			return fmt.Errorf("migrate sfc_user_totp: %w", err)
		}

		// ── site schemas ──────────────────────────────────────
		// Migrate all existing site schemas (for already-created sites)
		rows, err := db.QueryContext(ctx, `SELECT slug FROM public.sfc_sites`)
		if err != nil {
			return fmt.Errorf("list sites: %w", err)
		}
		defer rows.Close()

		var slugs []string
		for rows.Next() {
			var slug string
			if err := rows.Scan(&slug); err != nil {
				return fmt.Errorf("scan slug: %w", err)
			}
			slugs = append(slugs, slug)
		}

		for _, slug := range slugs {
			schema := "site_" + slug
			if _, err := db.ExecContext(ctx, fmt.Sprintf(`
				-- post_types: built_in BOOLEAN → SMALLINT
				ALTER TABLE %[1]s.sfc_site_post_types
					ALTER COLUMN built_in TYPE SMALLINT USING CASE WHEN built_in THEN 2 ELSE 1 END,
					ALTER COLUMN built_in SET DEFAULT 1;

				-- post_category_map: is_primary BOOLEAN → primary SMALLINT
				ALTER TABLE %[1]s.sfc_site_post_category_map
					ADD COLUMN "primary" SMALLINT NOT NULL DEFAULT 1;
				UPDATE %[1]s.sfc_site_post_category_map SET "primary" = CASE WHEN is_primary THEN 2 ELSE 1 END;
				ALTER TABLE %[1]s.sfc_site_post_category_map DROP COLUMN is_primary;

				-- comments: is_pinned BOOLEAN → pinned SMALLINT
				ALTER TABLE %[1]s.sfc_site_comments
					ADD COLUMN pinned SMALLINT NOT NULL DEFAULT 1;
				UPDATE %[1]s.sfc_site_comments SET pinned = CASE WHEN is_pinned THEN 2 ELSE 1 END;
				ALTER TABLE %[1]s.sfc_site_comments DROP COLUMN is_pinned;

				-- menu_items: is_active BOOLEAN → status SMALLINT
				ALTER TABLE %[1]s.sfc_site_menu_items
					ADD COLUMN status SMALLINT NOT NULL DEFAULT 1;
				UPDATE %[1]s.sfc_site_menu_items SET status = CASE WHEN is_active THEN 1 ELSE 2 END;
				ALTER TABLE %[1]s.sfc_site_menu_items DROP COLUMN is_active;

				-- redirects: is_active BOOLEAN → status SMALLINT
				DROP INDEX IF EXISTS %[1]s.idx_sfc_site_redirects_source;
				ALTER TABLE %[1]s.sfc_site_redirects
					ADD COLUMN status SMALLINT NOT NULL DEFAULT 1;
				UPDATE %[1]s.sfc_site_redirects SET status = CASE WHEN is_active THEN 1 ELSE 2 END;
				ALTER TABLE %[1]s.sfc_site_redirects DROP COLUMN is_active;
				CREATE INDEX idx_sfc_site_redirects_source ON %[1]s.sfc_site_redirects(source_path)
					WHERE status = 1;

				-- api_keys: is_active BOOLEAN → status SMALLINT
				ALTER TABLE %[1]s.sfc_site_api_keys
					ADD COLUMN status SMALLINT NOT NULL DEFAULT 1;
				UPDATE %[1]s.sfc_site_api_keys SET status = CASE WHEN is_active THEN 1 ELSE 2 END;
				ALTER TABLE %[1]s.sfc_site_api_keys DROP COLUMN is_active;
			`, schema)); err != nil {
				return fmt.Errorf("migrate site schema %s: %w", schema, err)
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		// Down migration: revert to boolean (simplified — production should never need this)
		return fmt.Errorf("down migration not supported for boolean-to-smallint conversion")
	})
}
```

**Step 2: Update site schema DDL template**

Modify `internal/schema/template.go` — replace all BOOLEAN fields with SMALLINT:

- `sfc_site_post_types.built_in`: `BOOLEAN NOT NULL DEFAULT FALSE` → `SMALLINT NOT NULL DEFAULT 1`
- `sfc_site_post_category_map.is_primary`: `is_primary BOOLEAN NOT NULL DEFAULT FALSE` → `"primary" SMALLINT NOT NULL DEFAULT 1`
- `sfc_site_comments.is_pinned`: `is_pinned BOOLEAN NOT NULL DEFAULT FALSE` → `pinned SMALLINT NOT NULL DEFAULT 1`
- `sfc_site_menu_items.is_active`: `is_active BOOLEAN NOT NULL DEFAULT TRUE` → `status SMALLINT NOT NULL DEFAULT 1`
- `sfc_site_redirects.is_active`: `is_active BOOLEAN NOT NULL DEFAULT TRUE` → `status SMALLINT NOT NULL DEFAULT 1`
- `sfc_site_api_keys.is_active`: `is_active BOOLEAN NOT NULL DEFAULT TRUE` → `status SMALLINT NOT NULL DEFAULT 1`
- Index on `sfc_site_redirects`: `WHERE is_active = TRUE` → `WHERE status = 1`

**Step 3: Verify compilation**

Run: `go build ./migrations/... && go build ./internal/schema/...`

**Step 4: Commit**

```bash
git add migrations/20260224000005_boolean_to_smallint.go internal/schema/template.go
git commit -m "feat(migration): convert all boolean fields to smallint enums in public and site schemas"
```

---

## Task 4: Adapt Auth Module for New Field Names

**Files:**
- Modify: `internal/auth/service.go`
- Modify: `internal/auth/repository.go`
- Modify: `internal/auth/dto.go`
- Modify: `internal/auth/service_test.go`
- Modify: `internal/auth/handler_test.go`

**Step 1: Update references**

**service.go:**
- Line 72: `!user.IsActive` → `user.Status != model.UserStatusActive`
- Line 96: `totp.IsEnabled` → `totp.Enabled == model.ToggleYes`
- Line 197: `IsActive: user.IsActive` → `Status: user.Status`
- Line 283: `existing.IsEnabled` → `existing.Enabled == model.ToggleYes`
- Line 307: `IsEnabled: false` → `Enabled: model.ToggleNo`
- Line 326: `totp.IsEnabled` → `totp.Enabled == model.ToggleYes`
- Line 453: `Enabled: totp.IsEnabled` → `Enabled: totp.Enabled == model.ToggleYes`

**repository.go:**
- Line 77: `revoked = false` → `revoked = 1` (ToggleNo)
- Line 88: `revoked = true` → `revoked = 2` (ToggleYes)
- Line 96: `revoked = true` → `revoked = 2`
- Line 98: `revoked = false` → `revoked = 1`
- Line 154: `is_enabled = EXCLUDED.is_enabled` → `enabled = EXCLUDED.enabled`
- Line 164: `is_enabled = true` → `enabled = 2` (ToggleYes)
- Line 223: `is_active = true` → `status = 1` (UserStatusActive)

**dto.go:**
- Line 45: `IsActive bool` → `Status model.UserStatus`

**Step 2: Update test files**

- All `IsActive: true` → `Status: model.UserStatusActive`
- All `IsActive: false` → `Status: model.UserStatusDisabled`
- All `IsEnabled: true` → `Enabled: model.ToggleYes`
- All `IsEnabled: false` → `Enabled: model.ToggleNo`
- `user.IsActive = false` → `user.Status = model.UserStatusDisabled`

**Step 3: Run tests**

Run: `go test ./internal/auth/... -v -count=1`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add internal/auth/
git commit -m "refactor(auth): adapt to boolean-to-smallint enum migration"
```

---

## Task 5: Adapt Site Module for New Field Names

**Files:**
- Modify: `internal/site/dto.go`
- Modify: `internal/site/repository.go`
- Modify: `internal/site/service.go`
- Modify: `internal/site/handler.go`
- Modify: `internal/site/service_test.go`

**Step 1: Update references**

**dto.go:**
- `ListFilter.IsActive *bool` → `Status *model.SiteStatus`
- `UpdateSiteReq.IsActive *bool` → `Status *model.SiteStatus`
- `SiteResp.IsActive bool` → `Status model.SiteStatus`
- `ToSiteResp`: `IsActive: s.IsActive` → `Status: s.Status`
- `UserBriefResp.IsActive bool` → `Status model.UserStatus`
- `ToSiteUserRespList`: `IsActive: item.User.IsActive` → `Status: item.User.Status`

**repository.go:**
- `is_active = ?` → `status = ?`
- `is_active = true` → `status = 1`

**service.go:**
- `IsActive: true` → `Status: model.SiteStatusActive`
- `req.IsActive` → `req.Status`
- `site.IsActive = *req.IsActive` → `site.Status = *req.Status`

**handler.go:**
- Replace `is_active` query param parsing with `status` int parsing

**service_test.go:**
- `IsActive: true` → `Status: model.SiteStatusActive`

**Step 2: Run tests**

Run: `go test ./internal/site/... -v -count=1`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/site/
git commit -m "refactor(site): adapt to boolean-to-smallint enum migration"
```

---

## Task 6: Adapt RBAC Module for New Field Names

**Files:**
- Modify: `internal/rbac/role_repo.go`
- Modify: `internal/rbac/template_repo.go`
- Modify: `internal/rbac/handler.go`
- Modify: `internal/rbac/handler_test.go`
- Modify: `internal/rbac/role_repo_test.go`
- Modify: `internal/rbac/template_repo_test.go`

**Step 1: Update references**

**role_repo.go:**
- `built_in DESC` → stays (column name unchanged, just type changed)
- `built_in = false` → `built_in = 1` (ToggleNo)

**template_repo.go:**
- `built_in DESC` → stays
- `built_in = false` → `built_in = 1`

**handler.go:**
- `role.BuiltIn` (bool check) → `role.BuiltIn == model.ToggleYes`
- `tmpl.BuiltIn` → `tmpl.BuiltIn == model.ToggleYes`

**handler_test.go:**
- `BuiltIn: true` → `BuiltIn: model.ToggleYes`
- `BuiltIn: false` → `BuiltIn: model.ToggleNo`

**Step 2: Run tests**

Run: `go test ./internal/rbac/... -v -count=1`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/rbac/
git commit -m "refactor(rbac): adapt to boolean-to-smallint enum migration"
```

---

## Task 7: Verify Full Build After Migration

**Step 1: Build entire project**

Run: `go build ./...`
Expected: SUCCESS — no compilation errors remaining.

**Step 2: Run all existing tests**

Run: `go test ./internal/model/... ./internal/pkg/... ./internal/auth/... ./internal/site/... ./internal/rbac/... -v -count=1`
Expected: ALL PASS

**Step 3: Commit any remaining fixes (if needed)**

---

## Task 8: Add Resend Config

**Files:**
- Modify: `internal/config/config.go`

**Step 1: Add ResendConfig struct and fields**

Add to Config struct:
```go
Resend ResendConfig
```

Add struct:
```go
type ResendConfig struct {
	APIKey    string
	FromName  string
	FromEmail string
}
```

Add defaults in `Load()`:
```go
viper.SetDefault("RESEND_FROM_NAME", "Sky Flux CMS")
viper.SetDefault("RESEND_FROM_EMAIL", "noreply@example.com")
```

Add to cfg construction:
```go
Resend: ResendConfig{
	APIKey:    viper.GetString("RESEND_API_KEY"),
	FromName:  viper.GetString("RESEND_FROM_NAME"),
	FromEmail: viper.GetString("RESEND_FROM_EMAIL"),
},
```

**Step 2: Verify**

Run: `go build ./internal/config/...`

**Step 3: Commit**

```bash
git add internal/config/config.go
git commit -m "feat(config): add Resend email service configuration"
```

---

## Task 9: Implement AuditContext Middleware

**Files:**
- Create: `internal/middleware/audit_context.go`
- Create: `internal/middleware/audit_context_test.go`

**Step 1: Write test**

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuditContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	var capturedIP, capturedUA string
	r.Use(AuditContext())
	r.GET("/test", func(c *gin.Context) {
		capturedIP, _ = c.Get("audit_ip").(string) // wrong — c.Get returns (any, bool)
		// Actually fix:
		v, _ := c.Get("audit_ip")
		capturedIP = v.(string)
		v2, _ := c.Get("audit_ua")
		capturedUA = v2.(string)
		c.Status(200)
	})

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("User-Agent", "TestAgent/1.0")
	c.Request.RemoteAddr = "192.168.1.1:1234"
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, "192.168.1.1", capturedIP)
	assert.Equal(t, "TestAgent/1.0", capturedUA)
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/middleware/ -run TestAuditContext -v`
Expected: FAIL — AuditContext not defined

**Step 3: Implement**

```go
package middleware

import "github.com/gin-gonic/gin"

// AuditContext extracts client IP and User-Agent from the request
// and stores them in the gin context for the audit service to read.
func AuditContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("audit_ip", c.ClientIP())
		c.Set("audit_ua", c.GetHeader("User-Agent"))
		c.Next()
	}
}
```

**Step 4: Run test**

Run: `go test ./internal/middleware/ -run TestAuditContext -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/middleware/audit_context.go internal/middleware/audit_context_test.go
git commit -m "feat(middleware): add AuditContext middleware for IP and User-Agent extraction"
```

---

## Task 10: Implement SiteResolver Middleware

**Files:**
- Modify: `internal/middleware/site_resolver.go`
- Modify: `internal/middleware/site_resolver_test.go`

**Step 1: Check existing test file for patterns**

Read `internal/middleware/site_resolver_test.go` to understand existing test structure (it may have tests for the TODO stub).

**Step 2: Implement SiteResolver**

SiteResolver reads site slug from `X-Site-Slug` header or resolves from Host header via domain lookup. Stores `site_id` and `site_slug` in gin context.

```go
package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SiteLookup is the interface the middleware needs to resolve sites.
type SiteLookup interface {
	GetIDBySlug(ctx context.Context, slug string) (string, error)
	GetSlugByDomain(ctx context.Context, domain string) (string, string, error) // slug, id, error
}

func SiteResolver(lookup SiteLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.GetHeader("X-Site-Slug")

		if slug == "" {
			// Try domain-based resolution
			host := c.Request.Host
			if host != "" {
				s, id, err := lookup.GetSlugByDomain(c.Request.Context(), host)
				if err == nil && s != "" {
					c.Set("site_slug", s)
					c.Set("site_id", id)
					c.Next()
					return
				}
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "missing X-Site-Slug header or valid domain",
			})
			return
		}

		id, err := lookup.GetIDBySlug(c.Request.Context(), slug)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "site not found",
			})
			return
		}

		c.Set("site_slug", slug)
		c.Set("site_id", id)
		c.Next()
	}
}
```

**Step 3: Write/update tests, run**

Run: `go test ./internal/middleware/ -run TestSiteResolver -v`

**Step 4: Commit**

```bash
git add internal/middleware/site_resolver.go internal/middleware/site_resolver_test.go
git commit -m "feat(middleware): implement SiteResolver with X-Site-Slug header and domain lookup"
```

---

## Task 11: Implement Schema Middleware

**Files:**
- Modify: `internal/middleware/schema.go`
- Modify: `internal/middleware/schema_test.go`

**Step 1: Implement**

SchemaMiddleware reads `site_slug` from context (set by SiteResolver) and executes `SET search_path TO 'site_{slug}', 'public'`.

```go
package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/schema"
	"github.com/uptrace/bun"
)

func Schema(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug, exists := c.Get("site_slug")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "site context not set",
			})
			return
		}

		slugStr := slug.(string)
		if !schema.ValidateSlug(slugStr) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid site slug",
			})
			return
		}

		schemaName := "site_" + slugStr
		_, err := db.ExecContext(c.Request.Context(), fmt.Sprintf(
			"SET search_path TO %s, public", bun.Ident(schemaName),
		))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to set schema context",
			})
			return
		}

		c.Next()
	}
}
```

**Step 2: Write tests, run**

Run: `go test ./internal/middleware/ -run TestSchema -v`

**Step 3: Commit**

```bash
git add internal/middleware/schema.go internal/middleware/schema_test.go
git commit -m "feat(middleware): implement Schema middleware with search_path isolation"
```

---

## Task 12: Implement AuditService (pkg/audit)

**Files:**
- Create: `internal/pkg/audit/audit.go`
- Create: `internal/pkg/audit/audit_test.go`

**Step 1: Write test**

Test that `Log()` inserts a record and auto-extracts actor info from context.

**Step 2: Implement**

```go
package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/sky-flux/cms/internal/model"
)

type Logger interface {
	Log(ctx context.Context, entry Entry) error
}

type Entry struct {
	Action           model.LogAction
	ResourceType     string
	ResourceID       string
	ResourceSnapshot any
}

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

// ctxKey types for context value extraction
type ctxKey string

const (
	KeyActorID    ctxKey = "user_id"
	KeyActorEmail ctxKey = "user_email"
	KeyAuditIP    ctxKey = "audit_ip"
	KeyAuditUA    ctxKey = "audit_ua"
)

func (s *Service) Log(ctx context.Context, entry Entry) error {
	var snapshot json.RawMessage
	if entry.ResourceSnapshot != nil {
		data, err := json.Marshal(entry.ResourceSnapshot)
		if err != nil {
			return fmt.Errorf("marshal snapshot: %w", err)
		}
		snapshot = data
	}

	record := &model.Audit{
		Action:           entry.Action,
		ResourceType:     entry.ResourceType,
		ResourceID:       entry.ResourceID,
		ResourceSnapshot: snapshot,
	}

	// Extract actor info from gin context values (set by Auth + AuditContext middleware)
	// These values are propagated via gin's c.Request.Context() or c.Set()
	// We need to handle both gin context and standard context
	if v := ctxValue(ctx, "user_id"); v != "" {
		record.ActorID = &v
	}
	if v := ctxValue(ctx, "user_email"); v != "" {
		record.ActorEmail = v
	}
	if v := ctxValue(ctx, "audit_ip"); v != "" {
		record.IPAddress = v
	}
	if v := ctxValue(ctx, "audit_ua"); v != "" {
		record.UserAgent = v
	}

	_, err := s.db.NewInsert().Model(record).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

// ctxValue extracts a string value from context.
// Gin stores values via c.Set() which are accessible via c.Request.Context().Value().
func ctxValue(ctx context.Context, key string) string {
	if v := ctx.Value(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
```

> **Note**: Gin's `c.Set()` stores values in gin.Context, which are also accessible through `c.Request.Context()` if the handler passes `c.Request.Context()` to service methods. The service layer must receive the gin-enhanced context. Verify this works; if gin context values don't propagate, the handler layer should explicitly set them via `context.WithValue`.

**Step 3: Run tests**

Run: `go test ./internal/pkg/audit/... -v`

**Step 4: Commit**

```bash
git add internal/pkg/audit/
git commit -m "feat(pkg/audit): implement AuditService with context-based actor extraction"
```

---

## Task 13: Implement Mail Service (pkg/mail)

**Files:**
- Create: `internal/pkg/mail/mail.go`
- Create: `internal/pkg/mail/templates.go`
- Create: `internal/pkg/mail/mail_test.go`

**Step 1: Implement Sender interface and ResendSender**

```go
package mail

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v3"
)

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type Message struct {
	To      string
	Subject string
	HTML    string
}

type ResendSender struct {
	client    *resend.Client
	fromName  string
	fromEmail string
}

func NewResendSender(apiKey, fromName, fromEmail string) *ResendSender {
	return &ResendSender{
		client:    resend.NewClient(apiKey),
		fromName:  fromName,
		fromEmail: fromEmail,
	}
}

func (s *ResendSender) Send(ctx context.Context, msg Message) error {
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{msg.To},
		Subject: msg.Subject,
		Html:    msg.HTML,
	}
	_, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend send: %w", err)
	}
	return nil
}

// NoopSender is a no-op implementation for testing.
type NoopSender struct{}

func (n *NoopSender) Send(_ context.Context, _ Message) error { return nil }
```

**Step 2: Implement templates**

```go
package mail

import (
	"bytes"
	"html/template"
)

var welcomeTpl = template.Must(template.New("welcome").Parse(`
<!DOCTYPE html>
<html><body>
<h2>Welcome to {{.SiteName}}</h2>
<p>Your account has been created. Here are your login credentials:</p>
<p><strong>Email:</strong> {{.Email}}</p>
<p><strong>Temporary Password:</strong> {{.Password}}</p>
<p>Please change your password after first login.</p>
</body></html>
`))

var disabledTpl = template.Must(template.New("disabled").Parse(`
<!DOCTYPE html>
<html><body>
<h2>Account Disabled</h2>
<p>Your account ({{.Email}}) on {{.SiteName}} has been disabled by an administrator.</p>
<p>If you believe this is a mistake, please contact support.</p>
</body></html>
`))

func RenderWelcome(siteName, email, password string) (string, error) {
	var buf bytes.Buffer
	err := welcomeTpl.Execute(&buf, map[string]string{
		"SiteName": siteName,
		"Email":    email,
		"Password": password,
	})
	return buf.String(), err
}

func RenderDisabled(siteName, email string) (string, error) {
	var buf bytes.Buffer
	err := disabledTpl.Execute(&buf, map[string]string{
		"SiteName": siteName,
		"Email":    email,
	})
	return buf.String(), err
}
```

**Step 3: Write test for template rendering**

**Step 4: Install resend-go dependency**

Run: `go get github.com/resend/resend-go/v3`

**Step 5: Run tests**

Run: `go test ./internal/pkg/mail/... -v`

**Step 6: Commit**

```bash
git add internal/pkg/mail/ go.mod go.sum
git commit -m "feat(pkg/mail): implement Resend mail service with welcome and disabled templates"
```

---

## Task 14: Users Module

**Files:**
- Create: `internal/user/interfaces.go`
- Create: `internal/user/dto.go`
- Create: `internal/user/repository.go`
- Create: `internal/user/service.go`
- Create: `internal/user/handler.go`
- Create: `internal/user/service_test.go`
- Create: `internal/user/handler_test.go`

Follow the pattern from `internal/site/` (interfaces → dto → repo → service → handler).

**interfaces.go:**
```go
package user

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/mail"
)

type UserRepository interface {
	List(ctx context.Context, filter ListFilter) ([]model.User, int64, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	SoftDelete(ctx context.Context, id string) error
	CountByStatus(ctx context.Context, status model.UserStatus) (int64, error)
}

type RoleRepository interface {
	GetBySlug(ctx context.Context, slug string) (*model.Role, error)
}

type UserRoleRepository interface {
	Assign(ctx context.Context, userID, roleID string) error
	GetRoleSlug(ctx context.Context, userID string) (string, error)
}

type TokenRevoker interface {
	RevokeAllForUser(ctx context.Context, userID string) error
}

type JWTBlacklister interface {
	BlacklistUser(ctx context.Context, userID string) error
}
```

**dto.go:**
```go
package user

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

type ListFilter struct {
	Page    int
	PerPage int
	Role    string
	Query   string
}

type CreateUserReq struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,max=100"`
	Role        string `json:"role" binding:"required"`
}

type UpdateUserReq struct {
	DisplayName *string            `json:"display_name" binding:"omitempty,max=100"`
	Role        *string            `json:"role"`
	Status      *model.UserStatus  `json:"status"`
}

type UserResp struct {
	ID          string           `json:"id"`
	Email       string           `json:"email"`
	DisplayName string           `json:"display_name"`
	AvatarURL   string           `json:"avatar_url,omitempty"`
	Role        string           `json:"role"`
	Status      model.UserStatus `json:"status"`
	LastLoginAt *time.Time       `json:"last_login_at,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
```

**Key business logic in service.go:**
- `CreateUser`: generate temp password (crypto.GenerateRandomToken) → bcrypt hash → insert user → assign role → async welcome email
- `UpdateUser`: if status changes to Disabled → async disabled email + revoke refresh tokens
- `DeleteUser`: self-delete check → last-super check → soft delete → revoke tokens → blacklist JWT
- `ListUsers`: paginated with role JOIN and q search

**Key test scenarios:**
- Create user with invalid role → error
- Create user with duplicate email → conflict
- Delete self → forbidden
- Delete last super → forbidden
- Disable user → tokens revoked + email sent
- List with pagination and role filter

**Step: Write tests → Implement → Verify → Commit**

```bash
git add internal/user/
git commit -m "feat(user): implement user management module with 5 endpoints"
```

---

## Task 15: Settings Module

**Files:**
- Replace: `internal/system/handler.go` (currently 3-line stub)
- Create: `internal/system/service.go`
- Create: `internal/system/repository.go`
- Create: `internal/system/dto.go`
- Create: `internal/system/interfaces.go`
- Create: `internal/system/service_test.go`

**2 endpoints:** GET /settings (Admin+), PUT /settings/:key (Super)

**interfaces.go:**
```go
type ConfigRepository interface {
	List(ctx context.Context) ([]model.SiteConfig, error)
	GetByKey(ctx context.Context, key string) (*model.SiteConfig, error)
	Upsert(ctx context.Context, cfg *model.SiteConfig) error
}
```

**Key logic:**
- GET: simple list from sfc_site_configs
- PUT: fetch old value → update → audit log with old+new snapshot

**Step: Write tests → Implement → Verify → Commit**

```bash
git add internal/system/
git commit -m "feat(system): implement settings module with 2 endpoints"
```

---

## Task 16: API Keys Module

**Files:**
- Replace: `internal/apikey/handler.go` (currently 3-line stub)
- Create: `internal/apikey/service.go`
- Create: `internal/apikey/repository.go`
- Create: `internal/apikey/dto.go`
- Create: `internal/apikey/interfaces.go`
- Create: `internal/apikey/service_test.go`

**3 endpoints:** GET /api-keys, POST /api-keys, DELETE /api-keys/:id

**Key logic:**
- Create: generate `cms_live_` + 32-byte hex → SHA-256 hash → store hash + prefix → return plaintext once
- Delete: set `revoked_at` + `status = APIKeyStatusRevoked`, not physical delete
- List: return all (active + revoked), sorted by created_at DESC

**Step: Write tests → Implement → Verify → Commit**

```bash
git add internal/apikey/
git commit -m "feat(apikey): implement API key management with 3 endpoints"
```

---

## Task 17: Post Types Module

**Files:**
- Create: `internal/posttype/handler.go`
- Create: `internal/posttype/service.go`
- Create: `internal/posttype/repository.go`
- Create: `internal/posttype/dto.go`
- Create: `internal/posttype/interfaces.go`
- Create: `internal/posttype/service_test.go`

**4 endpoints:** GET /post-types, POST /post-types, PUT /post-types/:id, DELETE /post-types/:id

**Key logic:**
- Fields JSON schema validation in service layer
- `built_in = ToggleYes` → cannot delete, cannot modify slug
- List includes `field_count` (len of fields array) and optionally `post_count`

**Step: Write tests → Implement → Verify → Commit**

```bash
git add internal/posttype/
git commit -m "feat(posttype): implement post types module with 4 endpoints"
```

---

## Task 18: Audit Logs Module

**Files:**
- Replace: `internal/audit/handler.go` (if exists as stub, otherwise create)
- Create: `internal/audit/repository.go`
- Create: `internal/audit/dto.go`
- Create: `internal/audit/interfaces.go`
- Create: `internal/audit/handler_test.go`

**1 endpoint:** GET /audit-logs (Super)

**No service layer** — handler calls repo directly (read-only query).

**Key logic:**
- Pagination + filters (actor_id, action, resource_type, start_date, end_date)
- JOIN `public.sfc_users` for actor display_name (cross-schema query)

**Step: Write tests → Implement → Verify → Commit**

```bash
git add internal/audit/
git commit -m "feat(audit): implement audit logs query endpoint"
```

---

## Task 19: Router — Wire Site-Scoped Routes + API Registry

**Files:**
- Modify: `internal/router/router.go`
- Modify: `internal/router/api_meta.go`

**Step 1: Add imports and DI for new modules**

Add to `Setup()`:
- `audit.NewService(db)` for audit logger
- `mail.NewResendSender(cfg.Resend.APIKey, ...)` for mailer
- Users module wiring (repo → service → handler)
- Settings module wiring
- API Keys module wiring
- Post Types module wiring
- Audit Logs module wiring

**Step 2: Create SiteLookup adapter for SiteResolver middleware**

Create a small adapter struct that implements `middleware.SiteLookup` by delegating to `site.SiteRepo`.

**Step 3: Register site-scoped route group**

```go
// Site-scoped routes
siteScoped := v1.Group("")
siteScoped.Use(middleware.SiteResolver(siteLookup))
siteScoped.Use(middleware.Schema(db))
siteScoped.Use(middleware.AuditContext())
siteScoped.Use(middleware.Auth(jwtMgr))
siteScoped.Use(middleware.RBAC(rbacSvc))

// Settings
siteScoped.GET("/settings", systemHandler.List)
siteScoped.PUT("/settings/:key", systemHandler.Update)

// API Keys
siteScoped.GET("/api-keys", apikeyHandler.List)
siteScoped.POST("/api-keys", apikeyHandler.Create)
siteScoped.DELETE("/api-keys/:id", apikeyHandler.Delete)

// Post Types
siteScoped.GET("/post-types", posttypeHandler.List)
siteScoped.POST("/post-types", posttypeHandler.Create)
siteScoped.PUT("/post-types/:id", posttypeHandler.Update)
siteScoped.DELETE("/post-types/:id", posttypeHandler.Delete)

// Audit Logs
siteScoped.GET("/audit-logs", auditHandler.List)
```

**Step 4: Register Users on global route group**

```go
// Users management (global, JWT + RBAC)
users := v1.Group("/users")
users.Use(middleware.Auth(jwtMgr))
users.Use(middleware.AuditContext())
users.Use(middleware.RBAC(rbacSvc))
users.GET("", userHandler.List)
users.POST("", userHandler.Create)
users.GET("/:id", userHandler.Get)
users.PUT("/:id", userHandler.Update)
users.DELETE("/:id", userHandler.Delete)
```

**Step 5: Update API Registry metadata**

Add 15 new entries to `BuildAPIMetaMap()` in `api_meta.go`.

**Step 6: Verify build**

Run: `go build ./...`

**Step 7: Commit**

```bash
git add internal/router/
git commit -m "feat(router): register site-scoped routes and users management with API Registry"
```

---

## Task 20: Final Verification

**Step 1: Run all tests**

Run: `go test ./... -count=1 -timeout 60s`
Expected: ALL PASS (skip testcontainers tests if Docker not running)

**Step 2: Verify go vet**

Run: `go vet ./...`
Expected: No issues

**Step 3: Count endpoints**

Previous: 51 global endpoints
New: 5 (users) + 10 (site-scoped) = 15
Total: **66 endpoints**

**Step 4: Final commit (if any fixes needed)**

---

## Summary

| Task | Description | Commit |
|------|-------------|--------|
| 1 | Enum types | `feat(model): add Toggle and Status enum types` |
| 2 | Model updates | `refactor(model): convert all boolean fields to smallint enums` |
| 3 | Migration 5 | `feat(migration): convert all boolean fields to smallint` |
| 4 | Auth adaptation | `refactor(auth): adapt to boolean-to-smallint` |
| 5 | Site adaptation | `refactor(site): adapt to boolean-to-smallint` |
| 6 | RBAC adaptation | `refactor(rbac): adapt to boolean-to-smallint` |
| 7 | Full build verify | — |
| 8 | Resend config | `feat(config): add Resend email service configuration` |
| 9 | AuditContext MW | `feat(middleware): add AuditContext middleware` |
| 10 | SiteResolver MW | `feat(middleware): implement SiteResolver` |
| 11 | Schema MW | `feat(middleware): implement Schema middleware` |
| 12 | AuditService | `feat(pkg/audit): implement AuditService` |
| 13 | MailService | `feat(pkg/mail): implement Resend mail service` |
| 14 | Users module | `feat(user): implement user management (5 endpoints)` |
| 15 | Settings module | `feat(system): implement settings (2 endpoints)` |
| 16 | API Keys module | `feat(apikey): implement API keys (3 endpoints)` |
| 17 | Post Types module | `feat(posttype): implement post types (4 endpoints)` |
| 18 | Audit Logs module | `feat(audit): implement audit logs query (1 endpoint)` |
| 19 | Router wiring | `feat(router): register site-scoped routes + users + API Registry` |
| 20 | Final verification | — |

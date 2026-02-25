# Batch 7: Comments + Menus + Redirects — Design Document

> **Goal:** Implement 23 site-scoped management endpoints for comment moderation, navigation menu management, and URL redirect management.
>
> **Dependencies:** Batch 4 (site-scoped infra) + Batch 6 (Posts — comments reference post_id FK)
>
> **Deferred to Batch 8:** RedirectMiddleware, public endpoints, rate limiting, Redis caching for public API

## Architecture Overview

All three modules follow the established handler → service → repository pattern. They mount on the existing `siteScoped` route group with the middleware chain: `SiteResolver → Schema → AuditContext → Auth → RBAC`.

**Key decisions:**
- **Tree building:** Flat SQL query + Go-level tree assembly (not recursive CTE)
- **Model migration:** migration_6 adds missing fields (`description` for menus, `icon`/`css_class` for menu items)
- **Redirect cache:** CRUD writes invalidate Redis `site:{slug}:redirects:map` key (prep for Batch 8)
- **Comment notifications:** Admin reply triggers Resend email to original comment author

## Module 1: Comments (7 endpoints)

### Endpoints

| # | Method | Path | Description | Role |
|---|--------|------|-------------|------|
| 1 | GET | `/api/v1/site/comments` | List comments (paginated, filterable) | Editor+ |
| 2 | GET | `/api/v1/site/comments/:id` | Comment detail with replies | Editor+ |
| 3 | PUT | `/api/v1/site/comments/:id/status` | Update comment status | Editor+ |
| 4 | PUT | `/api/v1/site/comments/:id/pin` | Toggle pin status | Editor+ |
| 5 | POST | `/api/v1/site/comments/:id/reply` | Admin reply to comment | Editor+ |
| 6 | PUT | `/api/v1/site/comments/batch-status` | Batch status update (max 100) | Admin+ |
| 7 | DELETE | `/api/v1/site/comments/:id` | Hard delete (CASCADE) | Admin+ |

### File Structure

```
internal/comment/
├── dto.go          # Request/Response DTOs
├── handler.go      # Handler with 7 methods
├── service.go      # Service with business logic
├── repository.go   # CommentRepo
├── interfaces.go   # Interface definitions
├── handler_test.go
└── service_test.go
```

### Business Rules

- **List:** JOIN `sfc_site_posts` for post title/slug. DTO computes `gravatar_url` from `author_email` (MD5 hash). Filter by `post_id`, `status`, search `q` (content or author_name).
- **Detail:** Returns comment + flat query of direct children (`parent_id = :id`).
- **Pin:** Only top-level comments (`parent_id IS NULL`) can be pinned. Max 3 pinned per post. Enforced at service layer by counting existing pinned comments for the same `post_id`.
- **Reply:** `user_id`, `author_name`, `author_email` filled from JWT claims. Status auto-set to `approved`. Nesting depth limited to 3 levels (service checks parent chain depth). After insert, async Resend email notification to original comment's `author_email` via `pkg/mail`.
- **Batch status:** Max 100 IDs, transactional bulk UPDATE.
- **Delete:** Hard delete, FK `ON DELETE CASCADE` removes child replies automatically.
- **Audit:** Status changes, replies, and deletes are logged via `pkg/audit`.

### DTO Conversions

| Model Field | DTO Field | Conversion |
|-------------|-----------|------------|
| `Status CommentStatus` (smallint 1-4) | `status string` | "pending"/"approved"/"spam"/"trash" |
| `Pinned Toggle` (smallint 1/2) | `is_pinned bool` | ToggleYes → true |
| `AuthorEmail string` | `gravatar_url string` | MD5 hash → gravatar URL |
| `AuthorIP`, `UserAgent` | included in management detail | hidden from public API (Batch 8) |

### Dependencies

- `pkg/audit.Service` — audit logging
- `pkg/mail.Sender` — Resend notification on admin reply
- `sfc_site_posts` — JOIN for post title/slug (FK exists in DDL)

## Module 2: Menus — Site Navigation (9 endpoints)

### Endpoints

| # | Method | Path | Description | Role |
|---|--------|------|-------------|------|
| 1 | GET | `/api/v1/site/menus` | List menus (optional location filter) | Admin+ |
| 2 | POST | `/api/v1/site/menus` | Create menu | Admin+ |
| 3 | GET | `/api/v1/site/menus/:id` | Menu detail with nested item tree | Admin+ |
| 4 | PUT | `/api/v1/site/menus/:id` | Update menu metadata | Admin+ |
| 5 | DELETE | `/api/v1/site/menus/:id` | Delete menu (CASCADE) | Admin+ |
| 6 | POST | `/api/v1/site/menus/:id/items` | Add menu item | Admin+ |
| 7 | PUT | `/api/v1/site/menus/:id/items/:item_id` | Update menu item | Admin+ |
| 8 | DELETE | `/api/v1/site/menus/:id/items/:item_id` | Delete menu item (CASCADE) | Admin+ |
| 9 | PUT | `/api/v1/site/menus/:id/items/reorder` | Batch reorder items | Admin+ |

### File Structure

```
internal/menu/
├── dto.go          # Request/Response DTOs
├── handler.go      # Handler with 9 methods
├── service.go      # Service with tree building + validation
├── repository.go   # MenuRepo + MenuItemRepo
├── interfaces.go   # Interface definitions
├── tree.go         # BuildMenuTree() — flat-to-nested assembly
├── handler_test.go
└── service_test.go
```

### Business Rules

- **List:** Optional `location` filter. Returns `item_count` (SELECT COUNT per menu).
- **Detail:** Fetch menu + flat query all items for menu_id. Go-level tree assembly via `BuildMenuTree()` (sort by parent_id → sort_order).
- **Create menu:** `slug` unique within schema. Regex: `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`.
- **Add item:** Type-based validation:
  - `type=custom` → `url` required
  - `type=post/category/tag/page` → `reference_id` required (existence check deferred to Batch 8 public API)
  - `parent_id` must belong to same menu, max 3-level nesting
  - `target` must be `_self` or `_blank`
- **Reorder:** Transaction bulk UPDATE `parent_id` + `sort_order`. Validates: all item IDs belong to same menu, nesting depth ≤ 3.
- **Audit:** All write operations logged via `pkg/audit`.

### Migration 6: Add Missing Fields

```sql
-- Applied to all existing site schemas + schema template
ALTER TABLE {schema}.sfc_site_menus ADD COLUMN description TEXT;
ALTER TABLE {schema}.sfc_site_menu_items ADD COLUMN icon VARCHAR(50);
ALTER TABLE {schema}.sfc_site_menu_items ADD COLUMN css_class VARCHAR(100);
```

Model updates:
- `SiteMenu`: add `Description string \`bun:"description" json:"description,omitempty"\``
- `SiteMenuItem`: add `Icon string \`bun:"icon" json:"icon,omitempty"\`` and `CSSClass string \`bun:"css_class" json:"css_class,omitempty"\``

### Tree Assembly Algorithm

```go
func BuildMenuTree(items []*SiteMenuItem) []*SiteMenuItem {
    byID := make(map[string]*SiteMenuItem, len(items))
    for _, item := range items {
        byID[item.ID] = item
    }
    var roots []*SiteMenuItem
    for _, item := range items {
        if item.ParentID == nil {
            roots = append(roots, item)
        } else if parent, ok := byID[*item.ParentID]; ok {
            parent.Children = append(parent.Children, item)
        }
    }
    // Sort each level by sort_order (already sorted by SQL ORDER BY)
    return roots
}
```

## Module 3: Redirects (7 endpoints)

### Endpoints

| # | Method | Path | Description | Role |
|---|--------|------|-------------|------|
| 1 | GET | `/api/v1/site/redirects` | List redirects (paginated, filterable) | Admin+ |
| 2 | POST | `/api/v1/site/redirects` | Create redirect | Admin+ |
| 3 | PUT | `/api/v1/site/redirects/:id` | Update redirect | Admin+ |
| 4 | DELETE | `/api/v1/site/redirects/:id` | Delete redirect | Admin+ |
| 5 | DELETE | `/api/v1/site/redirects/batch` | Batch delete (max 100) | Admin+ |
| 6 | POST | `/api/v1/site/redirects/import` | CSV import | Admin+ |
| 7 | GET | `/api/v1/site/redirects/export` | CSV export | Admin+ |

### File Structure

```
internal/redirect/
├── dto.go          # Request/Response DTOs
├── handler.go      # Handler with 7 methods
├── service.go      # Service with CSV logic + cache invalidation
├── repository.go   # RedirectRepo
├── interfaces.go   # Interface definitions
├── handler_test.go
└── service_test.go
```

### Business Rules

- **List:** Filter by `q` (source_path/target_url), `status_code` (301/302), `status` (active/disabled). Sort options: `created_at:desc`, `hit_count:desc`, `last_hit_at:desc`, `source_path:asc`.
- **Create/Update:**
  - `source_path` must start with `/`, max 500 chars, trailing slash auto-stripped, no `?` allowed
  - `source_path` unique within schema
  - `status_code` must be 301 or 302 (default 301)
  - `created_by` filled from JWT claims user_id
  - After write: Redis DEL `site:{slug}:redirects:map`
- **Batch delete:** Max 100 IDs, transactional. Non-existent IDs silently skipped. Returns `deleted_count`.
- **CSV import:**
  - `multipart/form-data`, max 1MB file, max 1000 data rows
  - Go `encoding/csv` reader with header validation (`source_path,target_url,status_code`)
  - `status_code` column optional (defaults to 301)
  - Duplicate `source_path` entries skipped, reported in `errors` array
  - Transactional bulk INSERT
- **CSV export:**
  - `Content-Type: text/csv`, `Content-Disposition: attachment; filename="redirects.csv"`
  - Stream directly to response via `csv.NewWriter(c.Writer)`
  - Columns: `source_path,target_url,status_code,status,hit_count,created_at`
- **Audit:** Create, update, delete, batch delete, and import logged via `pkg/audit`.

### DTO Conversions

| Model Field | DTO Field | Conversion |
|-------------|-----------|------------|
| `Status RedirectStatus` (smallint 1/2) | `is_active bool` | Active(1) → true |
| `CreatedBy *string` (UUID) | `created_by {id, display_name}` | JOIN `sfc_users` |

## Cross-Module: Router Integration

### Route Registration

All 23 endpoints mount on the existing `siteScoped` route group in `internal/router/router.go`.

**Route conflict prevention:** Static paths (`/comments/batch-status`, `/redirects/batch`, `/redirects/import`, `/redirects/export`) must be registered BEFORE parameterized paths (`/comments/:id`, `/redirects/:id`) to prevent Gin from capturing "batch-status" as an `:id` parameter.

### API Registry

23 new RBAC metadata entries in `BuildAPIMetaMap()`. Total: 46 (existing) + 23 = **69 RBAC-protected endpoints**.

### DI Wiring

```
comment: commentRepo → commentSvc(repo, auditSvc, mailer) → commentHandler(svc)
menu:    menuRepo + menuItemRepo → menuSvc(menuRepo, itemRepo, auditSvc) → menuHandler(svc)
redirect: redirectRepo → redirectSvc(repo, auditSvc, rdb) → redirectHandler(svc)
```

## Summary

| Module | Endpoints | New Files | Migration | External Integration |
|--------|-----------|-----------|-----------|---------------------|
| Comments | 7 | 7 | — | Resend (reply notification) |
| Menus | 9 | 8 (incl. tree.go) | migration_6 (3 columns) | — |
| Redirects | 7 | 7 | — | Redis (cache invalidation) |
| **Total** | **23** | **22** | **1 migration** | Resend + Redis |

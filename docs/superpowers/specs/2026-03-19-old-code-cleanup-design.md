# Old Code Cleanup — Design Spec

**Date:** 2026-03-19
**Status:** Draft
**Goal:** Remove all legacy Gin-based modules that have been superseded by DDD bounded contexts (Chi + Huma v2).

---

## 1. Background

The project has migrated from a flat Gin-based architecture to DDD bounded contexts using Chi v5 + Huma v2. Old modules remain in `internal/` alongside their replacements, creating confusion and preventing `go mod tidy` from removing the `gin-gonic/gin` dependency.

The old Gin router (`internal/router/`) is no longer wired into `cmd/cms/serve.go` — the new Chi-based `newServer()` function is the sole entry point.

---

## 2. Package Audit

### 2.1 DELETE — Fully Replaced by DDD Code

These packages are 100% superseded. No new DDD code imports them. They are only referenced by each other and by the old `internal/router/router.go`.

| Old Package | Replaced By | File Count |
|-------------|-------------|------------|
| `internal/auth/` | `internal/identity/` | 7 |
| `internal/user/` | `internal/identity/` | 7 |
| `internal/post/` | `internal/content/` | 20 |
| `internal/category/` | `internal/content/` | 7 |
| `internal/tag/` | `internal/content/` (Tag aggregate TBD, but old code unused) | 7 |
| `internal/comment/` | `internal/content/` (Comment aggregate TBD, but old code unused) | 7 |
| `internal/setup/` | `internal/platform/` | 6 |
| `internal/audit/` | `internal/platform/` | 8 |
| `internal/feed/` | `internal/delivery/feed/` | 7 |
| `internal/public/` | `internal/delivery/public/` | 7 |
| `internal/redirect/` | `internal/site/` | 7 |
| `internal/menu/` | `internal/site/` | 8 |
| `internal/router/` | Chi routing in `cmd/cms/serve.go` | 3 |
| `internal/schema/` | Removed (v2 single-site, no dynamic schema creation) | 5 |
| `internal/preview/` | `internal/content/` or `internal/delivery/public/` (preview token flow) | 5 |
| `internal/posttype/` | Removed (post types handled in content BC) | 6 |
| `internal/system/` | `internal/platform/` (system config) | 7 |
| `internal/apikey/` | `internal/site/` or `internal/platform/` (API key management) | 6 |
| `internal/rbac/` | `internal/identity/` (RBAC is part of identity BC) | 20 |
| `internal/dashboard/` | New DDD dashboard not yet built, but old Gin code is dead | 9 |

**Total: 20 packages, ~163 files to delete.**

### 2.2 DELETE — Old Shared Infrastructure (Gin-dependent)

| Package | Reason |
|---------|--------|
| `internal/middleware/` | All 12 files are Gin middleware. Replaced by `internal/platform/middleware/` (Chi). |
| `internal/pkg/response/` | Gin-specific response helper. New code uses Huma's native response handling. |
| `internal/testutil/` | Gin test helpers (`gin.TestMode`, `*gin.Engine`). New tests use `net/http/httptest` + Chi. |

### 2.3 DELETE — Old Model Package

| Package | Reason |
|---------|--------|
| `internal/model/` | 26 bun model files. Only imported by old packages + `internal/pkg/audit/` + `internal/cron/`. New DDD code uses domain entities in each BC's `domain/` layer. |

### 2.4 MERGE — Partially Replaced, Needs Selective Cleanup

| Package | Old Files (DELETE) | New DDD Files (KEEP) |
|---------|-------------------|---------------------|
| `internal/media/` | `handler.go`, `handler_test.go`, `service.go`, `service_test.go`, `interfaces.go`, `repository.go`, `dto.go` (7 old Gin files) | `domain/`, `app/`, `infra/`, `delivery/` (DDD layers) |
| `internal/site/` | `handler.go`, `handler_test.go`, `service.go`, `service_test.go`, `interfaces.go`, `repository.go`, `dto.go` (7 old Gin files) | `domain/`, `app/`, `infra/`, `delivery/` (DDD layers) |

### 2.5 MERGE — `internal/pkg/audit/` (depends on `internal/model`)

`internal/pkg/audit/audit.go` imports `internal/model` for `model.LogAction`. After `internal/model/` is deleted, this package must be updated to use the platform domain's audit types or define its own constants.

**Note:** `pkg/audit` is imported by 33 old-code files but **zero** new DDD files. It can be deleted entirely if the new `internal/platform/` provides its own audit service.

### 2.6 MERGE — `internal/cron/` (depends on `internal/model`)

`internal/cron/repository.go` imports `internal/model`. The cron scheduler itself is framework-agnostic (uses `time.Ticker`), but its repository queries use old bun models. Decision: **DELETE** — the cron tasks (scheduled publish, token cleanup, soft-delete purge) will need to be reimplemented to use DDD domain entities.

### 2.7 KEEP — No Changes Needed

| Package | Reason |
|---------|--------|
| `internal/config/` | Already migrated to koanf. No gin dependency. |
| `internal/database/` | DB/Redis connection setup. Framework-agnostic. |
| `internal/pkg/apperror/` | Pure Go error types. Used by new DDD code. |
| `internal/pkg/crypto/` | Password hashing, token generation. Framework-agnostic. |
| `internal/pkg/jwt/` | JWT sign/verify. Framework-agnostic. |
| `internal/pkg/mail/` | Resend email sender. Framework-agnostic. |
| `internal/pkg/cache/` | Redis typed operations. Framework-agnostic. |
| `internal/pkg/search/` | Meilisearch wrapper. Framework-agnostic. |
| `internal/pkg/storage/` | S3/RustFS client. Framework-agnostic. |
| `internal/pkg/imaging/` | Image processing. Framework-agnostic. |
| `internal/shared/` | DDD scaffold test + doc.go. Keep. |
| `internal/identity/` | New DDD bounded context. Keep. |
| `internal/content/` | New DDD bounded context. Keep. |
| `internal/platform/` | New DDD bounded context. Keep. |
| `internal/delivery/` | New DDD bounded context. Keep. |

### 2.8 DELETE — Old Frontend

| Directory | Reason |
|-----------|--------|
| `web/` | Old Astro 5 SSR frontend. Replaced by `console/` (React SPA) + `web/` (Go Templ, TBD). The current `web/` contains Astro components, not Go Templ. |

> **Note:** `web/` deletion is out of scope for this cleanup. It will be handled separately when the console SPA and Go Templ web are built.

---

## 3. Dependency Graph (What Blocks What)

```
internal/router/ ──imports──> ALL old packages + middleware + model
ALL old packages ──imports──> internal/model/
internal/middleware/ ──imports──> internal/schema/ + internal/model/
internal/setup/ ──imports──> internal/schema/
internal/site/ (old files) ──imports──> internal/schema/ + internal/model/
internal/pkg/audit/ ──imports──> internal/model/
internal/pkg/response/ ──imports──> gin
internal/testutil/ ──imports──> gin
internal/cron/ ──imports──> internal/model/
```

Deletion order must respect this: delete `router/` first (top consumer), then leaf packages, then `model/` and `schema/` last.

---

## 4. Post-Cleanup Expected State

After cleanup, `internal/` should contain only:

```
internal/
├── identity/          # DDD BC: User, Auth, RBAC
├── content/           # DDD BC: Post, Category, Tag, Comment
├── media/             # DDD BC: MediaFile (only domain/app/infra/delivery)
├── site/              # DDD BC: Site, Menu, Redirect (only domain/app/infra/delivery)
├── platform/          # DDD BC: Install, Audit, SystemConfig
│   └── middleware/    # Chi middleware (install guard, etc.)
├── delivery/          # Public API, Feed, Web
├── shared/            # Cross-BC utilities
├── config/            # koanf config loader
├── database/          # DB + Redis connection
└── pkg/               # Shared utilities
    ├── apperror/
    ├── crypto/
    ├── jwt/
    ├── mail/
    ├── cache/
    ├── search/
    ├── storage/
    └── imaging/
```

Removed from `go.mod` after `go mod tidy`:
- `github.com/gin-gonic/gin`
- `github.com/gin-contrib/sse`
- Any other gin-only transitive dependencies

---

## 5. Risk Assessment

| Risk | Mitigation |
|------|------------|
| `go build` breaks after deletion | Task 4 verifies compilation before committing |
| `go test` failures in remaining packages | Task 4 runs `go test ./... -short` |
| `pkg/audit` or `cron` silently imported by new code | Grep verified: zero imports from new DDD code |
| Missing functionality after deletion | All deleted code is legacy; new DDD code is the source of truth |

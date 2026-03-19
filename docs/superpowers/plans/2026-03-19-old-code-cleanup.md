# Old Code Cleanup — Execution Plan

**Date:** 2026-03-19
**Spec:** `docs/superpowers/specs/2026-03-19-old-code-cleanup-design.md`
**Commit convention:** gitmoji, no Co-Authored-By

---

## Task 1: Delete the Old Gin Router (Top-Level Consumer)

**Why first:** `internal/router/` imports every old package. Deleting it first breaks the import chain.

```bash
rm -rf internal/router/
```

**Files deleted (3):**
- `internal/router/router.go`
- `internal/router/router_test.go`
- `internal/router/api_meta.go`

**Commit:** `🔥 remove old Gin router`

---

## Task 2: Delete All Fully Replaced Old Packages

Delete 18 old Gin-based module packages in one batch:

```bash
rm -rf internal/auth/
rm -rf internal/user/
rm -rf internal/post/
rm -rf internal/category/
rm -rf internal/tag/
rm -rf internal/comment/
rm -rf internal/setup/
rm -rf internal/audit/
rm -rf internal/feed/
rm -rf internal/public/
rm -rf internal/redirect/
rm -rf internal/menu/
rm -rf internal/preview/
rm -rf internal/posttype/
rm -rf internal/system/
rm -rf internal/apikey/
rm -rf internal/rbac/
rm -rf internal/dashboard/
```

**Files deleted (~146):**
- `auth/`: dto.go, handler.go, handler_test.go, interfaces.go, repository.go, service.go, service_test.go
- `user/`: dto.go, handler.go, handler_test.go, interfaces.go, repository.go, service.go, service_test.go
- `post/`: dto.go, handler_crud.go, handler_preview.go, handler_revision.go, handler_status.go, handler_test.go, handler_translation.go, handler.go, interfaces.go, repository_preview.go, repository_revision.go, repository_translation.go, repository.go, service_preview_test.go, service_preview.go, service_revision_test.go, service_revision.go, service_status.go, service_test.go, service_translation_test.go, service_translation.go, service.go
- `category/`: dto.go, handler.go, handler_test.go, interfaces.go, repository.go, service.go, service_test.go
- `tag/`: dto.go, handler.go, handler_test.go, interfaces.go, repository.go, service.go, service_test.go
- `comment/`: dto.go, handler.go, handler_test.go, interfaces.go, repository.go, service.go, service_test.go
- `setup/`: dto.go, handler.go, handler_test.go, interfaces.go, repository.go, service.go
- `audit/`: dto.go, handler.go, handler_test.go, interfaces.go, middleware.go, repository.go, service.go, service_test.go
- `feed/`: handler.go, handler_test.go, interfaces.go, repository.go, service.go, service_test.go, types.go
- `public/`: dto.go, handler.go, handler_test.go, interfaces.go, repository.go, service.go, service_test.go
- `redirect/`: dto.go, handler.go, handler_test.go, interfaces.go, repository.go, service.go, service_test.go
- `menu/`: dto.go, handler.go, handler_test.go, interfaces.go, repository.go, service.go, service_test.go, tree.go
- `preview/`: dto.go, handler.go, handler_test.go, repository.go, service.go
- `posttype/`: dto.go, handler.go, interfaces.go, repository.go, service.go, service_test.go
- `system/`: dto.go, handler.go, handler_test.go, interfaces.go, repository.go, service.go, service_test.go
- `apikey/`: dto.go, handler.go, interfaces.go, repository.go, service.go, service_test.go
- `rbac/`: api_registry.go, api_registry_test.go, api_repo.go, api_repo_test.go, dto.go, handler.go, handler_test.go, interfaces.go, menu_repo.go, menu_repo_test.go, role_api_repo.go, role_api_repo_test.go, role_repo.go, role_repo_test.go, service.go, service_test.go, template_repo.go, template_repo_test.go, user_role_repo.go, user_role_repo_test.go
- `dashboard/`: dto.go, dto_test.go, handler.go, handler_test.go, interfaces.go, repository.go, repository_test.go, service.go, service_test.go

**Commit:** `🔥 remove 18 old Gin-based module packages`

---

## Task 3: Delete Old Shared Infrastructure

```bash
rm -rf internal/middleware/
rm -rf internal/schema/
rm -rf internal/testutil/
rm -rf internal/pkg/response/
rm -rf internal/pkg/audit/
rm -rf internal/cron/
rm -rf internal/model/
```

**Files deleted:**
- `middleware/` (24 files): api_key.go, api_key_test.go, audit_context.go, audit_context_test.go, auth.go, auth_test.go, cors.go, cors_test.go, installation_guard.go, installation_guard_test.go, logger.go, logger_test.go, rate_limit.go, rate_limit_test.go, rbac.go, rbac_test.go, recovery.go, recovery_test.go, request_id.go, request_id_test.go, schema.go, schema_test.go, site_resolver.go, site_resolver_test.go
- `schema/` (5 files): migrate.go, schema_test.go, template.go, validate.go, validate_test.go
- `testutil/` (2 files): containers.go, httptest.go
- `pkg/response/` (2 files): response.go, response_test.go
- `pkg/audit/` (2 files): audit.go, audit_test.go
- `cron/` (7 files): cron_test.go, interfaces.go, repository.go, scheduler.go, scheduler_test.go, tasks.go, tasks_test.go
- `model/` (26 files): admin_menu.go, api_endpoint.go, api_key.go, audit.go, category.go, comment.go, config.go, enums.go, hooks.go, hooks_test.go, media.go, menu.go, password_reset_token.go, post_type.go, post.go, preview_token.go, redirect.go, refresh_token.go, role_template.go, role.go, site_config.go, site.go, tag.go, user_role.go, user_totp.go, user.go

**Commit:** `🔥 remove old middleware, model, schema, cron, and shared Gin utilities`

---

## Task 4: Clean Old Code from Mixed Packages (media, site)

Remove old Gin handler/service/repo files from packages that have BOTH old and new DDD code.

### `internal/media/` — delete 7 old files:

```bash
rm internal/media/handler.go
rm internal/media/handler_test.go
rm internal/media/service.go
rm internal/media/service_test.go
rm internal/media/interfaces.go
rm internal/media/repository.go
rm internal/media/dto.go
```

### `internal/site/` — delete 7 old files:

```bash
rm internal/site/handler.go
rm internal/site/handler_test.go
rm internal/site/service.go
rm internal/site/service_test.go
rm internal/site/interfaces.go
rm internal/site/repository.go
rm internal/site/dto.go
```

**Commit:** `🔥 remove old Gin files from media and site packages`

---

## Task 5: Run `go mod tidy` to Remove Gin Dependency

```bash
go mod tidy
```

Expected removals from `go.mod`:
- `github.com/gin-gonic/gin v1.11.0`
- `github.com/gin-contrib/sse` (indirect)
- Other gin-only transitive dependencies

**Commit:** `🔧 go mod tidy after old code removal`

---

## Task 6: Verify Build and Tests

```bash
go build ./cmd/cms
go vet ./...
go test ./... -short -count=1
```

All three must pass with zero errors. If any fail, fix before committing.

**No separate commit — verification only.**

---

## Summary

| Category | Packages | Approx Files |
|----------|----------|-------------|
| DELETE (fully replaced modules) | 18 packages | ~146 |
| DELETE (old router) | 1 package | 3 |
| DELETE (old shared infra) | 7 packages | 68 |
| MERGE (clean old from mixed) | 2 packages | 14 |
| **Total removed** | **28 packages** | **~231 files** |
| KEEP (DDD + infra) | 14 packages | unchanged |

### Commit Sequence

1. `🔥 remove old Gin router`
2. `🔥 remove 18 old Gin-based module packages`
3. `🔥 remove old middleware, model, schema, cron, and shared Gin utilities`
4. `🔥 remove old Gin files from media and site packages`
5. `🔧 go mod tidy after old code removal`

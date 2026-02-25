# E2E Testing Design — Playwright Full-Stack (v2 Rewrite)

**Date**: 2026-02-26
**Status**: Approved

## Overview

Full-stack E2E tests using Playwright against real Go backend + PostgreSQL + Redis + Meilisearch + RustFS. Covers all critical user journeys including Batch 12 system management pages. Tests all 4 roles (Super/Admin/Editor/Viewer).

## Architecture

```
docker-compose up (PG+Redis+Meilisearch+RustFS)
  → go run ./cmd/cms serve (:8080)
    → bun dev (Astro :4321, proxy /api → :8080)
      → Playwright browser tests (:4321)
```

## 4 Roles

| Role | Slug | Scope |
|------|------|-------|
| Super | `super` | All operations + multi-site management |
| Admin | `admin` | Site management (users/roles/settings/api-keys/comments/menus/redirects), NOT sites |
| Editor | `editor` | Content management (posts/categories/tags/media) |
| Viewer | `viewer` | Read-only access |

## File Structure

```
web/e2e/
├── fixtures/
│   ├── test-image.png
│   └── test-redirects.csv
├── helpers/
│   ├── auth.ts          # loginViaUI + loginViaAPI
│   ├── api.ts           # API seeding (setup, users, posts, sites, roles)
│   └── constants.ts     # 4 test users + test site + API_BASE
├── setup.spec.ts        # Installation wizard (5 cases)
├── auth.spec.ts         # Authentication flows (8 cases)
├── content.spec.ts      # Posts + Categories + Tags (10 cases, Editor)
├── media.spec.ts        # Media management (5 cases, Editor)
├── system.spec.ts       # System management (12 cases, Admin) — Batch 12 pages use test.fixme()
├── rbac.spec.ts         # 4-role permission matrix (8 cases)
└── multisite.spec.ts    # Multi-site isolation (5 cases, Super)
```

## Selector Strategy

Based on actual component code:
- Form fields: `#id` selectors (`#email`, `#password`, `#admin_display_name`, etc.)
- Navigation: `aria-label` (`User menu`, `Toggle sidebar`)
- Tables: `role` + text (`getByRole('heading', { name: /posts/i })`)
- Toasts: `[data-sonner-toast]`
- Dialogs: `getByRole('alertdialog')`

## Execution Order

Playwright `projects` with `dependencies`:
```
setup → auth → [content, media, system, rbac, multisite]
```

## Test Scenarios (53 total)

### setup.spec.ts (5)
1. Setup check returns not installed
2. Navigate to /setup shows wizard
3. Complete 3-step wizard (admin account → site config → review & install)
4. Setup check returns installed
5. Repeat install attempt rejected

### auth.spec.ts (8)
1. Login success → /dashboard
2. Wrong password → error toast
3. Unauthenticated /dashboard → redirect /login
4. User menu visible after login
5. Logout → session invalidated
6. Forgot password → /forgot-password/sent
7. Invalid email format → validation error
8. Password too short → validation error

### content.spec.ts (10)
1. Navigate to posts list
2. Create draft post (title + BlockNote editor)
3. Post appears in list
4. Publish post
5. Unpublish post
6. Soft delete post
7. Navigate to categories
8. Create category
9. Navigate to tags
10. Create tag

### media.spec.ts (5)
1. Navigate to media library
2. Upload image (react-dropzone)
3. Media detail dialog shows info
4. Search media by filename
5. Delete media

### system.spec.ts (12) — Batch 12, test.fixme()
1. Navigate to users list
2. Create user with role assignment
3. Navigate to roles list
4. View role permissions
5. Navigate to settings
6. Update site setting
7. Navigate to API keys
8. Create API key
9. Navigate to audit logs
10. Navigate to comments
11. Navigate to menus
12. Navigate to redirects

### rbac.spec.ts (8)
1. Super sees all navigation items (users/roles/sites/settings)
2. Admin sees management nav but NOT sites
3. Editor sees content nav only (posts/categories/tags/media)
4. Viewer sees content nav (read-only, no create buttons)
5. Editor cannot access /dashboard/users → forbidden
6. Viewer cannot see "New Post" button
7. API 403: Editor → GET /api/v1/sites
8. API 403: Viewer → POST /api/v1/posts

### multisite.spec.ts (5)
1. Create two sites via API
2. Create posts in different sites
3. Site A Public API only returns A posts
4. Site B Public API only returns B posts
5. Both sites visible in sites list

## Data Management

1. globalSetup not used — setup.spec.ts handles initialization
2. rbac.spec.ts beforeAll seeds 3 additional users (admin/editor/viewer) via API
3. Each spec creates its own data, no cross-file dependencies beyond setup
4. storageState not used — each test logs in fresh (simpler, more reliable)

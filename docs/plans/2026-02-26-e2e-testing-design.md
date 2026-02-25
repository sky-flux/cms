# E2E Testing Design — Playwright Full-Stack

**Date**: 2026-02-26
**Status**: Approved

## Overview

Full-stack E2E tests using Playwright against real Go backend + PostgreSQL + Redis + Meilisearch + RustFS. Covers all 6 critical user journeys defined in testing.md.

## Architecture

```
docker-compose up (PG+Redis+Meilisearch+RustFS)
  → go run ./cmd/cms serve (:8080)
    → bun dev (Astro :4321, proxy /api → :8080)
      → Playwright browser tests (:4321)
```

- **globalSetup**: Health-check all services, reset DB, run migrations, call setup/initialize API
- **No manual cleanup**: globalSetup resets DB on each full run
- **storageState**: Reuse auth sessions across tests within a file

## File Structure

```
web/
├── e2e/
│   ├── fixtures/
│   │   ├── test-image.png
│   │   └── test-redirects.csv
│   ├── helpers/
│   │   ├── auth.ts          # login/logout helpers (UI + API cookie injection)
│   │   ├── api.ts           # Direct API calls for data seeding
│   │   └── constants.ts     # Shared test data (users, site, API base URL)
│   ├── setup.spec.ts        # Installation wizard
│   ├── auth.spec.ts         # Authentication flows
│   ├── posts.spec.ts        # Post lifecycle
│   ├── media.spec.ts        # Media management
│   ├── rbac.spec.ts         # Role-based access control
│   └── multisite.spec.ts    # Multi-site isolation
├── playwright.config.ts
└── package.json
```

Execution order controlled via Playwright `projects` with `dependencies` (setup runs first, all others depend on it).

## Test Scenarios

### setup.spec.ts — Installation Wizard (5 cases)

- `GET /api/v1/setup/check` returns `installed: false`
- Navigate to /setup → complete 3-step wizard (site info, admin account, confirm)
- After install: redirected to /login, setup/check returns `installed: true`
- Repeat install attempt → rejected (409)
- Concurrent install protection (optional, hard to test in browser)

### auth.spec.ts — Authentication (7 cases)

- Login success → redirect to /dashboard
- Login failure → error message shown
- Unauthenticated access to /dashboard → redirect to /login
- 2FA setup → enable → login requires TOTP → verify → dashboard
- Change password → logout → login with new password
- Forgot password → email sent (verify API response)
- Invalid reset token → error shown
- Session expiry → auto-refresh (verify no interruption)
- Logout → token invalidated → protected routes inaccessible

### posts.spec.ts — Post Lifecycle (6 cases)

- Create draft post with title/content
- Edit post → revision created
- Publish post → Public API returns it
- Schedule post (set future publish date)
- Unpublish → Public API no longer returns it
- Soft delete → not in list
- Restore from trash → visible again
- Concurrent edit conflict → 409 (using API helper to simulate)

### media.spec.ts — Media Management (4 cases)

- Upload image → appears in media library with thumbnails
- Upload invalid file type → error shown
- Create post referencing image → delete image → reference protection error
- Batch delete unreferenced media → success
- Media search by filename

### rbac.spec.ts — Role-Based Access (5 cases)

- Viewer: no "New Post" button, no edit actions
- Viewer: direct navigation to /dashboard/posts/new → 403/redirect
- Editor: can create/edit posts, cannot access /dashboard/users
- Admin: can access settings, cannot access user management
- Super: full access to all pages
- API-level 403: Editor calls Super-only endpoint via fetch

### multisite.spec.ts — Multi-Site Isolation (5 cases)

- Create site_a and site_b via UI
- Create post in site_a with X-Site-Slug header
- Verify site_b Public API does not return site_a's post
- Create same-slug content in both sites → no conflict
- Delete site → schema dropped, data gone

## Data Management

1. **globalSetup** resets DB: truncate all tables or drop/recreate schemas + re-migrate
2. **setup.spec.ts** calls setup/initialize → creates super admin
3. **rbac.spec.ts** uses API helpers to create test users (editor, viewer)
4. **Each spec file** creates its own test data via API, no cross-file dependencies beyond initial setup
5. **storageState files** saved to `e2e/.auth/` for session reuse

## Configuration

### playwright.config.ts

- `testDir: './e2e'`
- `timeout: 30_000` (60s for setup spec)
- `retries: 1` (CI: 2)
- `fullyParallel: false` + `projects` with `dependencies` for ordered execution
- `use.baseURL: 'http://localhost:4321'`
- `use.trace: 'retain-on-failure'`
- `use.screenshot: 'only-on-failure'`
- `webServer`: optional, can auto-start `bun dev` in web/

### Dependencies

- `@playwright/test` (devDependency in web/package.json)
- No MSW needed — real backend handles all API calls

## Estimated Output

32 test cases across 6 spec files, covering all P1 E2E scenarios from testing.md.

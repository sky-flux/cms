# k6 Performance Testing Design

**Date**: 2026-02-26
**Status**: Approved

## Overview

Comprehensive k6 performance test suite for Sky Flux CMS, covering Public API, Auth flow, Content Management, and Feed/Sitemap endpoints with staged load ramping up to 1000 VUs.

## Directory Structure

```
web/performance/
├── lib/
│   ├── config.ts          # Environment vars, thresholds, stage templates
│   ├── auth.ts            # Login helper, auth headers
│   └── checks.ts          # Shared check functions
├── scenarios/
│   ├── public-api.ts      # 7 Public API endpoints (70% traffic)
│   ├── auth-flow.ts       # Login/refresh/logout cycle (10% traffic)
│   ├── content-mgmt.ts    # Admin CRUD operations (15% traffic)
│   └── feed-sitemap.ts    # RSS/Atom/Sitemap (5% traffic)
├── smoke.ts               # Smoke test: 5 VUs x 30s
└── full-load.ts           # Full load: k6 scenarios API, staged ramp to 1000 VUs
```

## Performance Thresholds

| Metric | Target |
|--------|--------|
| P99 response time | < 200ms |
| Error rate | < 1% |
| Max concurrency | 1000 VUs |

## Load Stages (full-load.ts)

| Phase | Duration | VUs | Purpose |
|-------|----------|-----|---------|
| Ramp-up 1 | 30s | 0 → 50 | Warm-up |
| Steady 1 | 1min | 50 | Baseline |
| Ramp-up 2 | 1min | 50 → 500 | Scale up |
| Steady 2 | 2min | 500 | Medium load |
| Ramp-up 3 | 1min | 500 → 1000 | Peak ramp |
| Steady 3 | 2min | 1000 | Peak load |
| Ramp-down | 30s | 1000 → 0 | Cool-down |

## Scenario Details

### 1. public-api.ts (70% VUs)

Simulates frontend users and crawlers accessing public endpoints:
- `GET /api/public/v1/posts` (paginated list)
- `GET /api/public/v1/posts/:slug` (detail)
- `GET /api/public/v1/categories` (tree)
- `GET /api/public/v1/tags` (list)
- `GET /api/public/v1/search?q=keyword`
- `GET /api/public/v1/posts/:slug/comments`
- `GET /api/public/v1/menus?location=header`

Auth: X-API-Key header

### 2. auth-flow.ts (10% VUs)

Simulates user login/refresh/logout cycle:
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/refresh`
- `PUT /api/v1/auth/password`
- `POST /api/v1/auth/logout`

Auth: JWT Bearer token

### 3. content-mgmt.ts (15% VUs)

Simulates Editor/Admin content operations (read-heavy):
- Posts: list → create draft → update → publish → revisions
- Categories: tree → create
- Tags: list → create → suggest
- Media: list → upload (small file)

Auth: JWT Bearer token (editor role)

### 4. feed-sitemap.ts (5% VUs)

Simulates search engine/RSS reader crawling:
- `GET /sitemap.xml`
- `GET /sitemap-posts.xml`
- `GET /feed/rss.xml`
- `GET /feed/atom.xml`

Auth: None

### 5. smoke.ts

5 VUs x 30s, 1-2 VUs per scenario. Validates all endpoints are reachable and respond correctly. Suitable for CI integration.

### 6. full-load.ts

Uses k6 `scenarios` API to run all 4 scenarios in parallel with realistic traffic distribution.

## Shared Libraries

### lib/config.ts
- Environment: `BASE_URL`, `API_KEY`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `SITE_SLUG`
- Default thresholds object
- Stage template functions: `smokeStages()`, `loadStages()`

### lib/auth.ts
- `login(email, password)` → `{ accessToken, refreshCookie }`
- `authHeaders(token)` → headers object
- Executed once in k6 `setup()` phase

### lib/checks.ts
- `checkResponse(res, name, maxLatency?)` — status 2xx + optional latency check
- `checkJSON(res, name)` — JSON parse + success field validation

## Run Commands

```bash
# Smoke test (CI, quick)
k6 run web/performance/smoke.ts

# Single scenario debug
k6 run web/performance/scenarios/public-api.ts

# Full staged load test
k6 run web/performance/full-load.ts

# With env vars
k6 run -e BASE_URL=http://localhost:8080 -e API_KEY=cms_live_xxx web/performance/full-load.ts

# JSON output
k6 run --out json=results.json web/performance/full-load.ts
```

## Makefile Integration

```makefile
test-perf-smoke:
	k6 run web/performance/smoke.ts

test-perf:
	k6 run web/performance/full-load.ts
```

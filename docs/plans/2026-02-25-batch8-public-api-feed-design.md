# Batch 8 Design: Public Headless API + RSS Feed & Sitemap

**Date**: 2026-02-25
**Status**: Approved
**Scope**: 15 new endpoints (9 Public API + 6 Feed/Sitemap) + 2 new middleware

---

## 1. Overview

Complete the remaining backend endpoints before frontend development:

- **Public Headless API** (`/api/public/v1/*`) — 9 endpoints for frontend consumption
- **RSS Feed & Sitemap** (`/feed/*`, `/sitemap*`) — 6 endpoints for SEO
- **New middleware**: API Key authentication + IP rate limiting

**Total after completion**: 138 endpoints (123 existing + 15 new)

---

## 2. Middleware

### 2.1 API Key Middleware (`middleware/api_key.go`)

**Flow**:
1. Read key from `X-API-Key` header
2. SHA-256 hash the key
3. Query `sfc_site_api_keys.key_hash` (via `apikey.Repo.GetByHash`)
4. Validate: `status = active`, `expires_at` not expired
5. Async update `last_used_at`
6. Store API key info in Gin context

**Errors**:
- 401 `API_KEY_MISSING` — no header
- 401 `API_KEY_INVALID` — hash not found or revoked
- 401 `API_KEY_EXPIRED` — expired

**Dependencies**:
- New method: `apikey.Repo.GetByHash(ctx, hash) (*model.APIKey, error)`

### 2.2 Rate Limit Middleware (`middleware/rate_limit.go`)

**Flow**:
1. Redis `SET site:{slug}:ratelimit:comment:{ip} 1 EX 30 NX`
2. If key exists (NX fails) → 429
3. Only applied to `POST /api/public/v1/posts/:slug/comments`

**Errors**:
- 429 `COMMENT_RATE_LIMITED`

---

## 3. Middleware Chains

| Route Group | Middleware Chain |
|-------------|----------------|
| `/api/public/v1/*` | InstallationGuard → SiteResolver → Schema → APIKey |
| `POST .../comments` | + RateLimit(30s) |
| `/api/public/v1/preview/:token` | InstallationGuard → SiteResolver → Schema (NO APIKey) |
| `/feed/*` + `/sitemap*` | InstallationGuard → SiteResolver → Schema (NO auth) |

---

## 4. Public Module (`internal/public/`)

### 4.1 File Structure

```
internal/public/
├── handler.go          — 9 HTTP handlers
├── service.go          — Business logic (query + Redis cache)
├── dto.go              — Public response DTOs (sanitized: no email/IP/UA)
├── interfaces.go       — Dependency interface definitions
├── handler_test.go     — Handler unit tests
└── service_test.go     — Service unit tests
```

### 4.2 Endpoints

| # | Method | Path | Cache | Description |
|---|--------|------|-------|-------------|
| 1 | GET | /posts | 60s | Published post list (paginated, locale, fields filter) |
| 2 | GET | /posts/:slug | 300s | Post detail (with SEO, categories/tags) |
| 3 | GET | /categories | 300s | Category tree (with post_count) |
| 4 | GET | /tags | 300s | Tag list (with post_count) |
| 5 | GET | /search | none | Meilisearch full-text search |
| 6 | GET | /posts/:slug/comments | 60s | Nested comment tree (approved only, sanitized) |
| 7 | POST | /posts/:slug/comments | none | Submit comment (honeypot + IP rate limit) |
| 8 | GET | /menus | 300s | Menu tree by location/slug |
| 9 | GET | /preview/:token | none | Draft preview (token consumption) |

### 4.3 Reuse Strategy

Service layer injects existing repo interfaces — NO new repos:

| Dependency | Source |
|-----------|--------|
| Post queries | `post.PostRepo` |
| Category tree | `category.Repo` |
| Tag list | `tag.Repo` |
| Comment queries + create | `comment.Repo` |
| Menu tree | `menu.Repo` |
| Preview token | `post.PreviewRepo` |
| Full-text search | `pkg/search.Client` (Meilisearch) |
| Caching | `pkg/cache.Client` (Redis) |

### 4.4 Key Behaviors

**Posts list** (`GET /posts`):
- Only `status = published`, `deleted_at IS NULL`
- Support `locale`, `fields` (sparse fieldset), `category`, `tag` query params
- Pagination: page/per_page (default 20, max 100)
- Redis cache key: `public:posts:{siteSlug}:{hash(queryParams)}`

**Post detail** (`GET /posts/:slug`):
- Lookup by slug (not ID)
- Include: author, cover_image, categories, tags, SEO fields, extra_fields
- Increment `view_count` asynchronously

**Categories** (`GET /categories`):
- Full tree with `post_count` (count of published posts per category)
- Flat query → tree builder (reuse category tree builder pattern)

**Tags** (`GET /tags`):
- Flat list with `post_count`
- Optional `sort=post_count:desc` for tag cloud

**Search** (`GET /search`):
- Proxy to Meilisearch index `posts-{siteSlug}`
- Only published posts in results
- Return relevance score

**Comments** (`GET /posts/:slug/comments`):
- Only `status = approved`
- Nested tree (max 3 levels), pinned first
- Sanitized: NO `author_email`, `author_ip`, `user_agent`
- Paginate top-level only, inline all replies

**Submit comment** (`POST /posts/:slug/comments`):
- Honeypot field: non-empty → auto-mark as spam (still 201 response)
- Guest: requires `author_name`, `author_email`
- Authenticated (optional JWT): auto-fill from token
- New comments default to `status = pending`
- Validate `parent_id` references approved comment in same post
- Max 3-level nesting

**Menus** (`GET /menus`):
- Query by `location` or `slug` (one required)
- Only `is_active = true` items
- Filter out `is_broken = true` items
- Nested tree response

**Preview** (`GET /preview/:token`):
- SHA-256 hash token → query `sfc_site_preview_tokens`
- Check `expires_at`
- Return full post content (any status)
- No caching — always fresh
- 404 if not found, 410 if expired

---

## 5. Feed Module (`internal/feed/`)

### 5.1 File Structure

```
internal/feed/
├── handler.go          — 6 XML handlers
├── service.go          — RSS/Atom/Sitemap generation
├── types.go            — encoding/xml struct definitions
├── interfaces.go       — Dependency interface definitions
├── handler_test.go     — Handler unit tests
└── service_test.go     — Service unit tests
```

### 5.2 Endpoints

| # | Path | Content-Type | Cache |
|---|------|-------------|-------|
| 1 | /feed/rss.xml | application/rss+xml; charset=utf-8 | 3600s |
| 2 | /feed/atom.xml | application/atom+xml; charset=utf-8 | 3600s |
| 3 | /sitemap.xml | application/xml; charset=utf-8 | 3600s |
| 4 | /sitemap-posts.xml | application/xml; charset=utf-8 | 3600s |
| 5 | /sitemap-categories.xml | application/xml; charset=utf-8 | 3600s |
| 6 | /sitemap-tags.xml | application/xml; charset=utf-8 | 3600s |

### 5.3 XML Generation

Use Go `encoding/xml` with struct tags. Define XML structs in `types.go`:

- `RSSFeed` / `RSSChannel` / `RSSItem` — RSS 2.0
- `AtomFeed` / `AtomEntry` — Atom 1.0
- `SitemapIndex` / `Sitemap` — Sitemap index
- `URLSet` / `URL` — Sitemap URLs

### 5.4 Key Behaviors

**RSS Feed** (`/feed/rss.xml`):
- Latest 20 published posts (configurable via `limit` param, max 50)
- Optional `category`/`tag` slug filter
- Include `content:encoded` with full HTML
- Include `dc:creator`, `pubDate`, `category` elements
- `atom:link` self-reference

**Atom Feed** (`/feed/atom.xml`):
- Same query as RSS, different XML format
- Atom 1.0 namespace

**Sitemap Index** (`/sitemap.xml`):
- References: sitemap-posts.xml, sitemap-categories.xml, sitemap-tags.xml
- `lastmod` from most recent content in each type

**Sitemap Posts** (`/sitemap-posts.xml`):
- All published posts
- Priority rules: 7d→0.9/daily, 30d→0.8/weekly, 90d→0.7/weekly, >90d→0.5/monthly, page type→0.6/monthly
- Auto-pagination at 50,000 URLs

**Sitemap Categories** (`/sitemap-categories.xml`):
- All categories
- Root: priority 0.6, children: priority 0.5
- `lastmod` from latest post in category

**Sitemap Tags** (`/sitemap-tags.xml`):
- Tags with at least 1 published post
- Fixed priority 0.4
- `lastmod` from latest post with tag

### 5.5 HTTP Response Headers

All feed/sitemap endpoints:
- `Cache-Control: public, max-age=3600`
- `ETag`: content hash
- `Last-Modified`: based on latest content change

### 5.6 Dependencies

| Dependency | Source |
|-----------|--------|
| Post queries | `post.PostRepo` (or new read-only interface) |
| Category list | `category.Repo` |
| Tag list | `tag.Repo` |
| Site config | `system.Repo` (for site title, URL, description) |
| Caching | `pkg/cache.Client` (Redis) |

---

## 6. Router Integration

```go
// Feed & Sitemap (no auth, site via Host header)
feeds := r.Group("")
feeds.Use(installGuard, siteResolver, schemaMW)
{
    feeds.GET("/feed/rss.xml",           feedHandler.RSSFeed)
    feeds.GET("/feed/atom.xml",          feedHandler.AtomFeed)
    feeds.GET("/sitemap.xml",            feedHandler.SitemapIndex)
    feeds.GET("/sitemap-posts.xml",      feedHandler.SitemapPosts)
    feeds.GET("/sitemap-categories.xml", feedHandler.SitemapCategories)
    feeds.GET("/sitemap-tags.xml",       feedHandler.SitemapTags)
}

// Public API (API Key auth, site via Host header)
public := r.Group("/api/public/v1")
public.Use(installGuard, siteResolver, schemaMW, apiKeyMW)
{
    public.GET("/posts",          publicHandler.ListPosts)
    public.GET("/posts/:slug",    publicHandler.GetPost)
    public.GET("/categories",     publicHandler.ListCategories)
    public.GET("/tags",           publicHandler.ListTags)
    public.GET("/search",         publicHandler.Search)
    public.GET("/posts/:slug/comments",  publicHandler.ListComments)
    public.POST("/posts/:slug/comments", rateLimitMW, publicHandler.CreateComment)
    public.GET("/menus",                 publicHandler.GetMenu)
}

// Preview (no API Key, token-based auth)
preview := r.Group("/api/public/v1")
preview.Use(installGuard, siteResolver, schemaMW)
{
    preview.GET("/preview/:token", publicHandler.Preview)
}
```

---

## 7. Testing Strategy

| Layer | Method | Coverage |
|-------|--------|----------|
| API Key middleware | Unit test (mock repo) | Valid/invalid/expired/revoked key |
| Rate limit middleware | Unit test (mock redis) | Allow/block/edge cases |
| Public handler | Unit test (mock service) | Request parsing, response format, error codes |
| Public service | Unit test (mock repos/cache/search) | Business logic, cache hit/miss, sanitization |
| Feed handler | Unit test (mock service) | Content-Type, headers, XML response |
| Feed service | Unit test (mock repos/cache) | RSS/Atom/Sitemap XML validation |

**Test count estimate**: ~80-100 tests total

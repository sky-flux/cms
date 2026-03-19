# Router Integration Spec — Chi + Huma Route Wiring

**Date:** 2026-03-19
**Status:** Draft
**Depends on:** All BC delivery layers (identity, content, media, site, delivery/feed, delivery/public, delivery/web, platform)

---

## 1. Goal

Wire all new DDD handler packages into `cmd/cms/serve.go`, replacing the old Gin-based `internal/router/router.go`. The result is a single `newServer(cfg)` function that returns a fully-configured `http.Handler` with all route groups, middleware, and DI.

---

## 2. Route Structure

```
Chi Router
├── Global Middleware: RealIP → RequestID → Recoverer → CORS → InstallGuard → RateLimit
│
├── GET  /health                          → inline Huma handler (system health)
│
├── /setup                                → InstallHandler (plain Chi, no auth)
│   ├── GET  /setup                       → SetupPage
│   ├── POST /setup/test-db               → TestDB
│   ├── POST /setup/migrate               → Migrate
│   └── POST /setup/create-admin          → CreateAdmin
│
├── /console/*                            → embed.FS SPA (index.html fallback)
│   └── Static files from ConsoleFS       → io/fs.Sub("console/dist")
│   └── Fallback: any non-file path       → serve index.html (SPA routing)
│
├── /api/v1/
│   ├── /public/                          → Huma API (no JWT required)
│   │   ├── Middleware: OptionalAPIKey
│   │   ├── GET  /posts                   → public.ListPosts
│   │   ├── GET  /posts/{slug}            → public.GetPost
│   │   ├── GET  /categories              → public.ListCategories
│   │   ├── GET  /tags                    → public.ListTags
│   │   └── GET  /search                  → public.Search
│   │
│   ├── /admin/
│   │   ├── /auth/                        → Huma API (no JWT — login/refresh/logout)
│   │   │   └── POST /login               → identity.Login
│   │   │   (future: refresh, logout, forgot-password, 2fa/*)
│   │   │
│   │   └── (JWT + RBAC protected group)
│   │       ├── Middleware: JWTAuth → RBAC
│   │       ├── POST /posts               → content.CreatePost
│   │       ├── POST /posts/{post_id}/publish → content.PublishPost
│   │       ├── POST /categories          → content.CreateCategory
│   │       ├── POST /media               → media.Upload
│   │       ├── GET  /media               → media.List
│   │       ├── DELETE /media/{id}        → media.Delete
│   │       ├── GET  /settings            → site.GetSettings
│   │       ├── PUT  /settings            → site.UpdateSettings
│   │       ├── POST /menus               → site.CreateMenu
│   │       ├── POST /menus/{id}/items    → site.AddMenuItem
│   │       ├── POST /redirects           → site.CreateRedirect
│   │       └── GET  /audit               → platform.ListAudit
│   │
│   └── (future: /admin/users, /admin/roles, /admin/sites, etc.)
│
├── /feed/                                → Feed handler (plain Chi, no auth)
│   ├── Middleware: (none — public XML)
│   ├── GET  /feed/rss                    → feed.RSS
│   ├── GET  /feed/atom                   → feed.Atom
│   ├── GET  /sitemap.xml                 → feed.SitemapIndex
│   ├── GET  /sitemap-posts.xml           → feed.SitemapPosts
│   ├── GET  /sitemap-categories.xml      → feed.SitemapCategories
│   └── GET  /sitemap-tags.xml            → feed.SitemapTags
│
└── /*                                    → WebHandler (Templ SSR, plain Chi)
    ├── GET  /                            → web.Home
    ├── GET  /posts/{slug}                → web.PostDetail
    ├── GET  /posts/partial               → web.PostsPartial
    ├── GET  /categories/{slug}           → web.CategoryArchive
    ├── GET  /categories/{slug}/partial   → web.CategoryPartial
    ├── GET  /tags/{slug}                 → web.TagArchive
    ├── GET  /tags/{slug}/partial         → web.TagPartial
    ├── GET  /search                      → web.Search
    ├── POST /posts/{slug}/comments       → web.SubmitComment
    └── GET  /{slug}                      → web.CustomPage (catch-all, MUST be last)
```

---

## 3. Huma API Instances

There are two Huma API instances sharing the same Chi router, differentiated by path prefix:

| Instance | Prefix | Purpose |
|----------|--------|---------|
| `adminAPI` | `/api/v1/admin` | Admin endpoints (identity, content, media, site, audit) |
| `publicAPI` | `/api/v1/public` | Public headless API (posts, categories, tags, search) |

Feed and Web routes use plain Chi handlers (no Huma) because they return XML/HTML, not JSON.

Install routes use plain Chi handlers for simplicity (JSON but simple enough without Huma validation).

---

## 4. Middleware Ordering

### Global (applied to all routes)

```go
r.Use(chimiddleware.RealIP)
r.Use(chimiddleware.RequestID)
r.Use(chimiddleware.Recoverer)
r.Use(CORSMiddleware(cfg.Server.FrontendURL))
r.Use(InstallGuard(installChecker))    // passthrough: /setup/*, /console/*, /health
r.Use(RateLimitMiddleware(rdb))        // global rate limit, fail-open
```

### Per-Group

| Group | Middleware Stack |
|-------|-----------------|
| `/setup` | (none — exempt via InstallGuard passthrough) |
| `/console/*` | (none — static file serving) |
| `/api/v1/public/*` | OptionalAPIKey |
| `/api/v1/admin/auth/*` | (none — login/refresh need no JWT) |
| `/api/v1/admin/*` (protected) | JWTAuth → RBAC |
| `/feed/*`, `/*` (web) | (none — public HTML/XML) |

### Middleware Sources

| Middleware | Current Location | Action |
|------------|-----------------|--------|
| InstallGuard | `internal/platform/middleware/install_guard.go` | **Use as-is** (already Chi-native) |
| OptionalAPIKey | `internal/delivery/public/middleware.go` | **Use as-is** (already Chi-native) |
| JWTAuth | `internal/middleware/auth.go` | **Needs Chi rewrite** (currently Gin) |
| RBAC | `internal/middleware/rbac.go` | **Needs Chi rewrite** (currently Gin) |
| RateLimit | `internal/middleware/rate_limit.go` | **Needs Chi rewrite** (currently Gin) |
| CORS | `internal/middleware/cors.go` | **Needs Chi rewrite** (currently Gin) |

The Chi-rewritten middleware should live in `internal/shared/middleware/` to avoid import cycles and clearly separate from the old Gin middleware.

---

## 5. DI Wiring (`cmd/cms/wire.go`)

A dedicated `wire.go` file constructs all dependencies in a top-down fashion. No DI framework — manual constructor injection.

### Dependency Graph

```
Config
  ├── DB connection (*bun.DB)
  ├── Redis client (*redis.Client)
  ├── Meilisearch client
  └── S3 client (RustFS)

Infrastructure Clients (shared)
  ├── JWTManager (jwt.NewManager)
  └── Mailer (resend or noop)

BC Use Cases (app layer)
  ├── identity:  LoginUseCase(UserRepo, JWTManager)
  ├── content:   CreatePost(PostRepo), PublishPost(PostRepo), CreateCategory(CategoryRepo)
  ├── media:     Upload(MediaRepo, S3, Imaging), List(MediaRepo), Delete(MediaRepo, S3)
  ├── site:      GetSite(SiteRepo), UpdateSite(SiteRepo), CreateMenu(MenuRepo), ...
  └── platform:  InstallUseCase(DB), AuditQuery(AuditRepo)

Delivery Handlers
  ├── identity.Handler(LoginUseCase)
  ├── content.Handler(CreatePost, PublishPost, CreateCategory)
  ├── media.Handler(Upload, List, Delete)
  ├── site.Handler(GetSite, UpdateSite, CreateMenu, AddMenuItem, CreateRedirect)
  ├── platform.InstallHandler(InstallUseCase)
  ├── platform.AuditHandler(AuditQuery)
  ├── public.Handler(PostQuery, CategoryQuery, TagQuery, SearchQuery)
  ├── feed.Handler(FeedPostQuery, FeedCategoryQuery, FeedTagQuery, FeedSiteQuery)
  └── web.WebHandler(PostQuery, CategoryQuery, TagQuery, CommentWriter, SiteConfigLoader)
```

### Wire Function Signature

```go
// wire.go
type Handlers struct {
    Install      *platformdelivery.InstallHandler
    Identity     identitydelivery.LoginExecutor
    Content      struct {
        CreatePost     contentdelivery.CreatePostExecutor
        PublishPost    contentdelivery.PublishPostExecutor
        CreateCategory contentdelivery.CreateCategoryExecutor
    }
    Media        struct {
        Upload UploadExecutor
        List   ListExecutor
        Delete DeleteExecutor
    }
    Site         struct {
        GetSite        sitedelivery.GetSiteExecutor
        UpdateSite     sitedelivery.UpdateSiteExecutor
        CreateMenu     sitedelivery.CreateMenuExecutor
        AddMenuItem    sitedelivery.AddMenuItemExecutor
        CreateRedirect sitedelivery.CreateRedirectExecutor
    }
    Audit        *platformdelivery.AuditHandler
    PublicAPI    *public.Handler
    Feed         *feed.Handler
    Web          *web.WebHandler

    // Middleware dependencies
    InstallChecker platformmw.InstallChecker
    APIKeyValidator public.APIKeyValidator
}

func wireHandlers(cfg *config.Config, db *bun.DB, rdb *redis.Client, ...) (*Handlers, error)
```

---

## 6. Console SPA Serving (`/console/*`)

The console React SPA is embedded via `go:embed all:console/dist` in the root `embed.go`.

### Production Mode

```go
consoleSub, _ := fs.Sub(rootfs.ConsoleFS, "console/dist")
fileServer := http.FileServer(http.FS(consoleSub))

r.Route("/console", func(r chi.Router) {
    r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
        // Try to serve the static file first.
        // If not found, serve index.html for SPA client-side routing.
        path := strings.TrimPrefix(r.URL.Path, "/console")
        if path == "" || path == "/" {
            path = "/index.html"
        }
        f, err := consoleSub.Open(strings.TrimPrefix(path, "/"))
        if err != nil {
            // File not found → serve index.html (SPA fallback)
            r.URL.Path = "/index.html"
        } else {
            f.Close()
        }
        fileServer.ServeHTTP(w, r)
    })
})
```

### Development Mode (`--dev` flag)

Proxy `/console/*` to Vite dev server at `localhost:3000`:

```go
if cfg.Server.Mode == "debug" {
    // httputil.ReverseProxy to localhost:3000
}
```

---

## 7. Web Static Assets

Static CSS and JS for the public Templ site are embedded via `go:embed all:web/static`:

```go
webStaticSub, _ := fs.Sub(rootfs.WebStaticFS, "web/static")
r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(webStaticSub))))
```

---

## 8. Old Router Cleanup

After the new `serve.go` is fully wired and tested:

1. **Delete** `internal/router/router.go` — all Gin-based route registration code
2. **Delete** `internal/router/registry.go` (API meta map for old RBAC sync)
3. **Delete** `internal/middleware/*.go` (old Gin middleware) — replaced by `internal/shared/middleware/`
4. **Keep** `internal/middleware/` directory temporarily if any non-Gin utilities exist there
5. Remove all `github.com/gin-gonic/gin` imports from `go.mod` once no code references it

---

## 9. Health Check

The health check remains a Huma endpoint at `/health` on the admin API instance (or a standalone Chi handler). It checks DB, Redis, Meilisearch, and RustFS connectivity.

---

## 10. Future Extensibility

The current new handlers cover the v1 proof-of-concept endpoints. As more use cases are added to each BC (e.g., `ListPosts`, `UpdatePost`, `DeletePost` in content), they follow the same pattern:

1. Add use case in `{bc}/app/`
2. Add executor interface + Huma registration in `{bc}/delivery/handler.go`
3. Add executor to `wire.go` construction
4. No changes to `serve.go` — `RegisterRoutes()` in each handler package handles route registration

This keeps `serve.go` stable as the codebase grows.

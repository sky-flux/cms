# Web Public Site Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the public-facing CMS website using Go Templ (SSR) + HTMX (dynamic interactions) + Tailwind CSS V4, embedded into the Go binary.

**Architecture:** web/templates/*.templ → go generate → Go code → go:embed into binary. Plain Chi handlers return HTML, no Huma/JSON. HTMX handles search, pagination, comment submission.

**Tech Stack:** Go Templ, HTMX 2.x, Tailwind CSS V4 CLI, Chi v5, testify + httptest

---

## Context

Per the redesign spec (`docs/superpowers/specs/2026-03-19-project-redesign-design.md` §5.6), the public site is a first-class deliverable embedded in the single Go binary. Pages:

| Page | URL | Rendering |
|------|-----|-----------|
| Homepage | `/` | Templ SSR — latest posts list |
| Post detail | `/posts/:slug` | Templ SSR + HTMX comment form |
| Category archive | `/categories/:slug` | Templ SSR + HTMX pagination |
| Tag archive | `/tags/:slug` | Templ SSR + HTMX pagination |
| Search | `/search?q=` | HTMX partial refresh |
| Custom page | `/:slug` | Templ SSR (Page-type post) |

Chi routes are registered **after** `/api` and `/console` routes. The catch-all `/*` serves Templ SSR pages.

Asset pipeline:
- Source: `web/styles/input.css` (`@import "tailwindcss"`)
- Output: `web/static/app.css` (go:embed target)
- HTMX: `web/static/htmx.min.js` (local copy, not CDN)

---

## Task 1: Setup Templ + Tailwind

Install tooling, create directory scaffolding, update build commands.

- [ ] **1.1** Install the Templ CLI:
  ```bash
  go install github.com/a-h/templ/cmd/templ@latest
  ```
  Verify with `templ version`.

- [ ] **1.2** Add `github.com/a-h/templ` runtime to `go.mod`:
  ```bash
  go get github.com/a-h/templ
  ```
  The runtime package (`github.com/a-h/templ`) is required at compile time; the CLI is only needed at generate time.

- [ ] **1.3** Create directory structure:
  ```bash
  mkdir -p web/templates web/styles web/static
  touch web/static/.gitkeep
  touch web/templates/.gitkeep
  ```

- [ ] **1.4** Create `web/styles/input.css`:
  ```css
  @import "tailwindcss";

  /* Sky Flux CMS — public site base styles */
  @theme {
    --font-sans: "Inter", ui-sans-serif, system-ui, sans-serif;
    --color-brand-500: #6366f1;
    --color-brand-600: #4f46e5;
  }
  ```

- [ ] **1.5** Install Tailwind CSS V4 CLI (standalone binary — no Node.js required):
  ```bash
  # macOS arm64
  curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
  chmod +x tailwindcss-macos-arm64
  mv tailwindcss-macos-arm64 /usr/local/bin/tailwindcss
  ```
  Verify with `tailwindcss --version`.

- [ ] **1.6** Download HTMX 2.x to `web/static/`:
  ```bash
  curl -sL https://unpkg.com/htmx.org@2/dist/htmx.min.js -o web/static/htmx.min.js
  ```

- [ ] **1.7** Update `Makefile` — add new targets and update existing ones:

  Add these targets after the existing `build` target:

  ```makefile
  # ──────────────────────────────────────
  # Web 公共站点 (Templ + HTMX + Tailwind)
  # ──────────────────────────────────────

  templ-generate:
  	templ generate ./web/templates/

  templ-watch:
  	templ generate --watch ./web/templates/

  css-build:
  	tailwindcss -i web/styles/input.css -o web/static/app.css --minify

  css-watch:
  	tailwindcss -i web/styles/input.css -o web/static/app.css --watch
  ```

  Update the `build` target to include Templ + CSS steps:
  ```makefile
  build:
  	templ generate ./web/templates/
  	tailwindcss -i web/styles/input.css -o web/static/app.css --minify
  	go build -ldflags="-w -s" -o ./tmp/cms ./cmd/cms
  ```

  Update `dev` to use a `Procfile` with overmind (parallel processes):
  ```makefile
  dev:
  	overmind start
  ```

  Create `Procfile` at repo root:
  ```
  api:     air -c .air.toml
  templ:   templ generate --watch ./web/templates/
  css:     tailwindcss -i web/styles/input.css -o web/static/app.css --watch
  console: cd console && bun run dev
  ```

- [ ] **1.8** Add `//go:generate` directive in `web/templates/` (optional convenience):
  ```go
  // web/templates/generate.go
  package templates

  //go:generate templ generate
  ```

- [ ] **1.9** Update `embed.go` at project root to include `web/static`:
  ```go
  // embed.go
  package cms

  import "embed"

  //go:embed all:console/dist
  var ConsoleFS embed.FS

  //go:embed all:web/static
  var WebStaticFS embed.FS
  ```
  Ensure `web/static/.gitkeep` is committed so `go:embed` never fails on a fresh clone.

- [ ] **1.10** Verify the go generate pipeline works end-to-end:
  ```bash
  # Write a trivial hello.templ, run generate, confirm _templ.go appears
  templ generate ./web/templates/
  go build ./...
  ```

---

## Task 2: Base Layout Template

`web/templates/layout.templ` is the HTML shell shared by all pages. It injects `app.css`, `htmx.min.js`, the `<nav>`, and `<footer>`.

- [ ] **2.1** Create `web/templates/layout.templ`:

  ```go
  package templates

  // NavItem represents a navigation menu entry.
  type NavItem struct {
      Label string
      URL   string
  }

  // SiteConfig holds site-wide settings for the layout.
  type SiteConfig struct {
      Name        string
      Description string
      NavItems    []NavItem
  }

  // Layout wraps all public pages in the HTML shell.
  templ Layout(cfg SiteConfig, title string) {
      <!DOCTYPE html>
      <html lang="en" class="scroll-smooth">
          <head>
              <meta charset="UTF-8"/>
              <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
              <title>{ title } — { cfg.Name }</title>
              <meta name="description" content={ cfg.Description }/>
              <link rel="stylesheet" href="/static/app.css"/>
              <!-- HTMX 2.x: hx-boost on <body> upgrades all same-origin links to AJAX -->
              <script src="/static/htmx.min.js" defer></script>
          </head>
          <body class="min-h-screen bg-white text-gray-900 antialiased" hx-boost="true">
              @Nav(cfg)
              <main id="main-content" class="mx-auto max-w-5xl px-4 py-8">
                  { children... }
              </main>
              @Footer(cfg)
          </body>
      </html>
  }

  // Nav renders the top navigation bar.
  templ Nav(cfg SiteConfig) {
      <header class="border-b border-gray-200 bg-white">
          <div class="mx-auto flex max-w-5xl items-center justify-between px-4 py-4">
              <a href="/" class="text-xl font-bold text-brand-600">{ cfg.Name }</a>
              <nav class="flex items-center gap-6 text-sm">
                  for _, item := range cfg.NavItems {
                      <a href={ templ.SafeURL(item.URL) } class="text-gray-600 hover:text-brand-600 transition-colors">
                          { item.Label }
                      </a>
                  }
                  <!-- Search input: HTMX triggers GET /search?q=... on input -->
                  <form action="/search" method="get" role="search">
                      <input
                          type="search"
                          name="q"
                          placeholder="Search…"
                          class="rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                          hx-get="/search"
                          hx-trigger="input changed delay:300ms"
                          hx-target="#search-results"
                          hx-push-url="true"
                      />
                  </form>
              </nav>
          </div>
      </header>
  }

  // Footer renders the site footer.
  templ Footer(cfg SiteConfig) {
      <footer class="mt-16 border-t border-gray-200 py-8 text-center text-sm text-gray-500">
          <p>© { cfg.Name }. Powered by <a href="https://github.com/sky-flux/cms" class="underline hover:text-brand-600">Sky Flux CMS</a>.</p>
      </footer>
  }
  ```

- [ ] **2.2** Create `web/templates/layout_test.go`:

  ```go
  package templates_test

  import (
      "bytes"
      "context"
      "strings"
      "testing"

      "github.com/sky-flux/cms/web/templates"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestLayout_RendersValidHTML(t *testing.T) {
      cfg := templates.SiteConfig{
          Name:        "Test Site",
          Description: "A test site",
          NavItems: []templates.NavItem{
              {Label: "Blog", URL: "/"},
              {Label: "About", URL: "/about"},
          },
      }

      var buf bytes.Buffer
      err := templates.Layout(cfg, "Home").Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, "<!DOCTYPE html>")
      assert.Contains(t, html, "<title>Home — Test Site</title>")
      assert.Contains(t, html, `src="/static/htmx.min.js"`)
      assert.Contains(t, html, `href="/static/app.css"`)
      assert.Contains(t, html, "hx-boost=\"true\"")
      assert.Contains(t, html, "Test Site")
      assert.Contains(t, html, "/about")
  }

  func TestNav_RendersNavItems(t *testing.T) {
      cfg := templates.SiteConfig{
          Name: "MySite",
          NavItems: []templates.NavItem{
              {Label: "Posts", URL: "/posts"},
              {Label: "Tags", URL: "/tags"},
          },
      }

      var buf bytes.Buffer
      err := templates.Nav(cfg).Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, "Posts")
      assert.Contains(t, html, "/posts")
      assert.Contains(t, html, "Tags")
      // Search input with HTMX attributes
      assert.Contains(t, html, `hx-get="/search"`)
      assert.Contains(t, html, `hx-trigger="input changed delay:300ms"`)
  }

  func TestFooter_RendersCopyright(t *testing.T) {
      cfg := templates.SiteConfig{Name: "MySite"}

      var buf bytes.Buffer
      err := templates.Footer(cfg).Render(context.Background(), &buf)
      require.NoError(t, err)

      assert.True(t, strings.Contains(buf.String(), "MySite"))
  }
  ```

- [ ] **2.3** Run `templ generate ./web/templates/` and `go test ./web/templates/... -v -count=1` — confirm RED (no handler yet), then GREEN after generate.

---

## Task 3: Homepage Template

Shows the latest published posts. No HTMX on initial load (full SSR). HTMX is used only for pagination (loaded later via `posts_partial.templ`).

- [ ] **3.1** Create `web/templates/home.templ`:

  ```go
  package templates

  import "time"

  // PostSummary is the view model for a post card on list pages.
  type PostSummary struct {
      Slug        string
      Title       string
      Excerpt     string
      CoverURL    string
      PublishedAt time.Time
      AuthorName  string
      CategorySlug string
      CategoryName string
      Tags        []TagRef
  }

  // TagRef is a lightweight tag reference used in post cards.
  type TagRef struct {
      Slug string
      Name string
  }

  // HomePage renders the site homepage with the latest posts.
  templ HomePage(cfg SiteConfig, posts []PostSummary) {
      @Layout(cfg, "Home") {
          <section class="mb-10">
              <h1 class="text-3xl font-bold tracking-tight text-gray-900">Latest Posts</h1>
              <p class="mt-2 text-gray-500">{ cfg.Description }</p>
          </section>

          <!-- Posts grid; initially server-rendered, HTMX pagination appends more below -->
          <div id="posts-list" class="grid gap-8 sm:grid-cols-2">
              for _, p := range posts {
                  @PostCard(p)
              }
          </div>

          <!-- HTMX infinite scroll sentinel:
               When this element enters the viewport, GET /posts/partial?page=2
               appends results into #posts-list.
               hx-swap="beforeend" + hx-target="#posts-list" means new cards are appended.
               hx-swap-oob on the sentinel itself updates the next-page URL. -->
          if len(posts) >= 10 {
              <div
                  id="load-more-sentinel"
                  class="mt-10 flex justify-center"
                  hx-get="/posts/partial?page=2"
                  hx-trigger="intersect once"
                  hx-target="#posts-list"
                  hx-swap="beforeend"
                  hx-indicator="#loading-spinner"
              >
                  <span class="text-sm text-gray-400">Loading more…</span>
              </div>
          }
          <div id="loading-spinner" class="htmx-indicator mt-4 flex justify-center">
              <span class="text-sm text-gray-400">Loading…</span>
          </div>
      }
  }

  // PostCard renders a single post card for use in list views.
  templ PostCard(p PostSummary) {
      <article class="rounded-xl border border-gray-200 bg-white shadow-sm hover:shadow-md transition-shadow">
          if p.CoverURL != "" {
              <img
                  src={ p.CoverURL }
                  alt={ p.Title }
                  class="h-48 w-full rounded-t-xl object-cover"
                  loading="lazy"
              />
          }
          <div class="p-5">
              if p.CategoryName != "" {
                  <a
                      href={ templ.SafeURL("/categories/" + p.CategorySlug) }
                      class="text-xs font-semibold uppercase tracking-wider text-brand-600"
                  >{ p.CategoryName }</a>
              }
              <h2 class="mt-2 text-xl font-semibold leading-snug text-gray-900">
                  <a href={ templ.SafeURL("/posts/" + p.Slug) } class="hover:text-brand-600">
                      { p.Title }
                  </a>
              </h2>
              if p.Excerpt != "" {
                  <p class="mt-2 text-sm text-gray-600 line-clamp-3">{ p.Excerpt }</p>
              }
              <div class="mt-4 flex items-center justify-between text-xs text-gray-400">
                  <span>{ p.AuthorName }</span>
                  <time datetime={ p.PublishedAt.Format("2006-01-02") }>
                      { p.PublishedAt.Format("Jan 2, 2006") }
                  </time>
              </div>
          </div>
      </article>
  }
  ```

- [ ] **3.2** Create `web/templates/home_test.go`:

  ```go
  package templates_test

  import (
      "bytes"
      "context"
      "testing"
      "time"

      "github.com/sky-flux/cms/web/templates"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestHomePage_RendersPostCards(t *testing.T) {
      cfg := templates.SiteConfig{Name: "My Blog", Description: "A cool blog"}
      posts := []templates.PostSummary{
          {
              Slug:         "hello-world",
              Title:        "Hello World",
              Excerpt:      "First post ever.",
              PublishedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
              AuthorName:   "Alice",
              CategorySlug: "news",
              CategoryName: "News",
          },
          {
              Slug:        "second-post",
              Title:       "Second Post",
              PublishedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
              AuthorName:  "Bob",
          },
      }

      var buf bytes.Buffer
      err := templates.HomePage(cfg, posts).Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, "Hello World")
      assert.Contains(t, html, "/posts/hello-world")
      assert.Contains(t, html, "Second Post")
      assert.Contains(t, html, "First post ever.")
      assert.Contains(t, html, "/categories/news")
      assert.Contains(t, html, "News")
      // No load-more sentinel for fewer than 10 posts
      assert.NotContains(t, html, "load-more-sentinel")
  }

  func TestHomePage_ShowsLoadMoreWhenFull(t *testing.T) {
      cfg := templates.SiteConfig{Name: "Blog"}
      posts := make([]templates.PostSummary, 10)
      for i := range posts {
          posts[i] = templates.PostSummary{
              Slug:        "post",
              Title:       "Post",
              PublishedAt: time.Now(),
          }
      }

      var buf bytes.Buffer
      err := templates.HomePage(cfg, posts).Render(context.Background(), &buf)
      require.NoError(t, err)

      assert.Contains(t, buf.String(), "load-more-sentinel")
      assert.Contains(t, buf.String(), `hx-get="/posts/partial?page=2"`)
  }

  func TestPostCard_RendersCoverImage(t *testing.T) {
      p := templates.PostSummary{
          Slug:        "with-cover",
          Title:       "With Cover",
          CoverURL:    "https://example.com/cover.jpg",
          PublishedAt: time.Now(),
      }

      var buf bytes.Buffer
      err := templates.PostCard(p).Render(context.Background(), &buf)
      require.NoError(t, err)

      assert.Contains(t, buf.String(), "https://example.com/cover.jpg")
  }

  func TestPostCard_OmitsCoverWhenEmpty(t *testing.T) {
      p := templates.PostSummary{Slug: "no-cover", Title: "No Cover", PublishedAt: time.Now()}

      var buf bytes.Buffer
      err := templates.PostCard(p).Render(context.Background(), &buf)
      require.NoError(t, err)

      assert.NotContains(t, buf.String(), "<img")
  }
  ```

---

## Task 4: Post Detail Template

Full article view with rendered HTML body, author bio, and an HTMX-powered comment form. Comments are submitted via `hx-post` without a full page reload.

- [ ] **4.1** Create `web/templates/post.templ`:

  ```go
  package templates

  import "time"

  // PostDetail is the full view model for a post detail page.
  type PostDetail struct {
      Slug        string
      Title       string
      BodyHTML    string // Pre-rendered HTML from BlockNote/Markdown
      CoverURL    string
      PublishedAt time.Time
      UpdatedAt   time.Time
      AuthorName  string
      AuthorBio   string
      CategorySlug string
      CategoryName string
      Tags        []TagRef
      Comments    []Comment
      AllowComments bool
  }

  // Comment is a single comment view model.
  type Comment struct {
      ID          string
      AuthorName  string
      AuthorEmail string // MD5 hash for Gravatar only
      Body        string
      CreatedAt   time.Time
      IsApproved  bool
      Children    []Comment
  }

  // PostPage renders a full article detail page.
  templ PostPage(cfg SiteConfig, post PostDetail) {
      @Layout(cfg, post.Title) {
          <article class="mx-auto max-w-2xl">
              <!-- Breadcrumb -->
              <nav class="mb-6 text-sm text-gray-400" aria-label="Breadcrumb">
                  <a href="/" class="hover:text-brand-600">Home</a>
                  <span class="mx-2">/</span>
                  if post.CategoryName != "" {
                      <a href={ templ.SafeURL("/categories/" + post.CategorySlug) } class="hover:text-brand-600">
                          { post.CategoryName }
                      </a>
                      <span class="mx-2">/</span>
                  }
                  <span class="text-gray-700">{ post.Title }</span>
              </nav>

              <!-- Header -->
              <header class="mb-8">
                  <h1 class="text-4xl font-bold leading-tight tracking-tight text-gray-900">
                      { post.Title }
                  </h1>
                  <div class="mt-4 flex flex-wrap items-center gap-4 text-sm text-gray-500">
                      <span>{ post.AuthorName }</span>
                      <time datetime={ post.PublishedAt.Format("2006-01-02") }>
                          { post.PublishedAt.Format("January 2, 2006") }
                      </time>
                      for _, tag := range post.Tags {
                          <a
                              href={ templ.SafeURL("/tags/" + tag.Slug) }
                              class="rounded-full bg-gray-100 px-3 py-0.5 text-xs text-gray-600 hover:bg-brand-100 hover:text-brand-600"
                          >{ tag.Name }</a>
                      }
                  </div>
              </header>

              if post.CoverURL != "" {
                  <img
                      src={ post.CoverURL }
                      alt={ post.Title }
                      class="mb-8 w-full rounded-xl object-cover shadow"
                      loading="eager"
                  />
              }

              <!-- Body: pre-rendered HTML, scoped with prose typography -->
              <div class="prose prose-gray max-w-none">
                  @templ.Raw(post.BodyHTML)
              </div>

              <!-- Author bio -->
              if post.AuthorBio != "" {
                  <aside class="mt-12 rounded-xl border border-gray-200 bg-gray-50 p-6">
                      <p class="font-semibold text-gray-900">About { post.AuthorName }</p>
                      <p class="mt-1 text-sm text-gray-600">{ post.AuthorBio }</p>
                  </aside>
              }
          </article>

          <!-- Comments section -->
          if post.AllowComments {
              <section id="comments" class="mx-auto mt-16 max-w-2xl">
                  <h2 class="mb-6 text-2xl font-bold text-gray-900">
                      Comments ({ len(post.Comments) })
                  </h2>

                  <!-- Approved comments list; HTMX swaps this after successful submission -->
                  <div id="comments-list">
                      for _, c := range post.Comments {
                          @CommentItem(c, 0)
                      }
                  </div>

                  <!-- Comment form:
                       hx-post submits to /posts/{slug}/comments
                       On 201 Created the server returns an OOB swap of #comments-list
                       and a success message in #comment-form-status.
                       hx-swap="outerHTML" replaces the entire form with a thank-you message. -->
                  @CommentForm(post.Slug)
              </section>
          }
      }
  }

  // CommentItem renders a single comment, supporting up to 3 levels of nesting.
  templ CommentItem(c Comment, depth int) {
      <div class={ "border-l-2 border-gray-200 pl-4", templ.KV("ml-8", depth > 0), templ.KV("ml-16", depth > 1) }>
          <div class="flex items-start gap-3 py-4">
              <div class="flex-1">
                  <div class="flex items-center gap-2">
                      <span class="font-semibold text-sm text-gray-900">{ c.AuthorName }</span>
                      <time class="text-xs text-gray-400" datetime={ c.CreatedAt.Format("2006-01-02") }>
                          { c.CreatedAt.Format("Jan 2, 2006") }
                      </time>
                  </div>
                  <p class="mt-1 text-sm text-gray-700">{ c.Body }</p>
              </div>
          </div>
          if depth < 2 {
              for _, child := range c.Children {
                  @CommentItem(child, depth+1)
              }
          }
      </div>
  }

  // CommentForm renders the new comment submission form.
  // HTMX interaction: hx-post="/posts/{slug}/comments" sends form data.
  // Server responds with HX-Reswap + OOB swap to update #comments-list.
  templ CommentForm(slug string) {
      <form
          id="comment-form"
          class="mt-8 rounded-xl border border-gray-200 bg-gray-50 p-6"
          hx-post={ "/posts/" + slug + "/comments" }
          hx-target="#comment-form-status"
          hx-swap="innerHTML"
          hx-on::after-request="if(event.detail.successful) this.reset()"
      >
          <h3 class="mb-4 text-lg font-semibold text-gray-900">Leave a Comment</h3>
          <div id="comment-form-status" class="mb-4 text-sm" aria-live="polite"></div>

          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                  <label for="comment-name" class="block text-sm font-medium text-gray-700">Name *</label>
                  <input
                      id="comment-name"
                      name="author_name"
                      type="text"
                      required
                      class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                  />
              </div>
              <div>
                  <label for="comment-email" class="block text-sm font-medium text-gray-700">Email *</label>
                  <input
                      id="comment-email"
                      name="author_email"
                      type="email"
                      required
                      class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                  />
              </div>
          </div>

          <div class="mt-4">
              <label for="comment-body" class="block text-sm font-medium text-gray-700">Comment *</label>
              <textarea
                  id="comment-body"
                  name="body"
                  rows="5"
                  required
                  class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
              ></textarea>
          </div>

          <button
              type="submit"
              class="mt-4 rounded-md bg-brand-600 px-5 py-2 text-sm font-semibold text-white hover:bg-brand-500 focus:outline-none focus:ring-2 focus:ring-brand-500"
              hx-disabled-elt="this"
          >
              Submit Comment
          </button>
      </form>
  }
  ```

- [ ] **4.2** Create `web/templates/post_test.go`:

  ```go
  package templates_test

  import (
      "bytes"
      "context"
      "testing"
      "time"

      "github.com/sky-flux/cms/web/templates"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestPostPage_RendersTitle(t *testing.T) {
      cfg := templates.SiteConfig{Name: "My Blog"}
      post := templates.PostDetail{
          Slug:          "test-post",
          Title:         "Test Post",
          BodyHTML:      "<p>Hello world</p>",
          PublishedAt:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
          AuthorName:    "Alice",
          AllowComments: true,
      }

      var buf bytes.Buffer
      err := templates.PostPage(cfg, post).Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, "Test Post")
      assert.Contains(t, html, "<p>Hello world</p>")
      assert.Contains(t, html, "Alice")
      assert.Contains(t, html, "March 1, 2026")
  }

  func TestPostPage_ShowsCommentFormWhenEnabled(t *testing.T) {
      cfg := templates.SiteConfig{Name: "Blog"}
      post := templates.PostDetail{
          Slug:          "commentable",
          Title:         "Commentable Post",
          PublishedAt:   time.Now(),
          AllowComments: true,
      }

      var buf bytes.Buffer
      err := templates.PostPage(cfg, post).Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, `hx-post="/posts/commentable/comments"`)
      assert.Contains(t, html, "Leave a Comment")
  }

  func TestPostPage_HidesCommentFormWhenDisabled(t *testing.T) {
      cfg := templates.SiteConfig{Name: "Blog"}
      post := templates.PostDetail{
          Slug:          "no-comments",
          Title:         "No Comments Post",
          PublishedAt:   time.Now(),
          AllowComments: false,
      }

      var buf bytes.Buffer
      err := templates.PostPage(cfg, post).Render(context.Background(), &buf)
      require.NoError(t, err)

      assert.NotContains(t, buf.String(), "Leave a Comment")
  }

  func TestPostPage_RendersNestedComments(t *testing.T) {
      cfg := templates.SiteConfig{Name: "Blog"}
      post := templates.PostDetail{
          Slug:        "nested",
          Title:       "Post with Comments",
          PublishedAt: time.Now(),
          AllowComments: true,
          Comments: []templates.Comment{
              {
                  ID:         "1",
                  AuthorName: "Alice",
                  Body:       "Top level comment",
                  CreatedAt:  time.Now(),
                  Children: []templates.Comment{
                      {
                          ID:         "2",
                          AuthorName: "Bob",
                          Body:       "Nested reply",
                          CreatedAt:  time.Now(),
                      },
                  },
              },
          },
      }

      var buf bytes.Buffer
      err := templates.PostPage(cfg, post).Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, "Top level comment")
      assert.Contains(t, html, "Nested reply")
      assert.Contains(t, html, "Alice")
      assert.Contains(t, html, "Bob")
  }

  func TestPostPage_RendersBreadcrumbWithCategory(t *testing.T) {
      cfg := templates.SiteConfig{Name: "Blog"}
      post := templates.PostDetail{
          Slug:         "cat-post",
          Title:        "Post in Category",
          PublishedAt:  time.Now(),
          CategorySlug: "tech",
          CategoryName: "Technology",
      }

      var buf bytes.Buffer
      err := templates.PostPage(cfg, post).Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, "/categories/tech")
      assert.Contains(t, html, "Technology")
  }

  func TestCommentForm_HasHTMXAttributes(t *testing.T) {
      var buf bytes.Buffer
      err := templates.CommentForm("my-post").Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, `hx-post="/posts/my-post/comments"`)
      assert.Contains(t, html, `hx-target="#comment-form-status"`)
      assert.Contains(t, html, `hx-disabled-elt="this"`)
  }
  ```

---

## Task 5: Archive Templates + Posts Partial

Category and tag archive pages share the same layout. A separate `posts_partial.templ` returns only the post card grid — used by HTMX for infinite scroll / "Load More".

- [ ] **5.1** Create `web/templates/category.templ`:

  ```go
  package templates

  // CategoryArchivePage renders posts filtered by category.
  templ CategoryArchivePage(cfg SiteConfig, categoryName string, categorySlug string, posts []PostSummary, page int) {
      @Layout(cfg, "Category: "+categoryName) {
          <section class="mb-10">
              <p class="text-sm text-gray-400 uppercase tracking-wider">Category</p>
              <h1 class="mt-1 text-3xl font-bold tracking-tight text-gray-900">{ categoryName }</h1>
          </section>

          <div id="posts-list" class="grid gap-8 sm:grid-cols-2">
              for _, p := range posts {
                  @PostCard(p)
              }
          </div>

          <!-- HTMX "Load More" button (alternative to intersection observer):
               Clicking appends the next page of posts into #posts-list.
               The server returns only the PostsPartial fragment (no full HTML shell).
               hx-swap="beforeend" appends cards; OOB swap replaces this button. -->
          if len(posts) >= 10 {
              <div class="mt-10 flex justify-center">
                  <button
                      id="load-more-btn"
                      class="rounded-md border border-gray-300 px-6 py-2 text-sm text-gray-600 hover:border-brand-600 hover:text-brand-600 transition-colors"
                      hx-get={ "/categories/" + categorySlug + "/partial?page=" + itoa(page+1) }
                      hx-target="#posts-list"
                      hx-swap="beforeend"
                      hx-indicator="#loading-spinner"
                  >
                      Load More
                  </button>
              </div>
          }
          <div id="loading-spinner" class="htmx-indicator mt-4 flex justify-center">
              <span class="text-sm text-gray-400">Loading…</span>
          </div>
      }
  }
  ```

- [ ] **5.2** Create `web/templates/tag.templ`:

  ```go
  package templates

  // TagArchivePage renders posts filtered by tag.
  templ TagArchivePage(cfg SiteConfig, tagName string, tagSlug string, posts []PostSummary, page int) {
      @Layout(cfg, "Tag: "+tagName) {
          <section class="mb-10">
              <p class="text-sm text-gray-400 uppercase tracking-wider">Tag</p>
              <h1 class="mt-1 text-3xl font-bold tracking-tight text-gray-900">{ tagName }</h1>
          </section>

          <div id="posts-list" class="grid gap-8 sm:grid-cols-2">
              for _, p := range posts {
                  @PostCard(p)
              }
          </div>

          if len(posts) >= 10 {
              <div class="mt-10 flex justify-center">
                  <button
                      id="load-more-btn"
                      class="rounded-md border border-gray-300 px-6 py-2 text-sm text-gray-600 hover:border-brand-600 hover:text-brand-600 transition-colors"
                      hx-get={ "/tags/" + tagSlug + "/partial?page=" + itoa(page+1) }
                      hx-target="#posts-list"
                      hx-swap="beforeend"
                  >
                      Load More
                  </button>
              </div>
          }
      }
  }
  ```

- [ ] **5.3** Create `web/templates/posts_partial.templ` — returns ONLY post cards (no HTML shell). The server must detect the `HX-Request: true` header and return this fragment instead of the full page.

  ```go
  package templates

  // PostsPartial returns just the post card grid rows for HTMX infinite scroll.
  // The handler checks r.Header.Get("HX-Request") == "true" before using this.
  // The response also includes an OOB swap to update/remove the load-more button.
  templ PostsPartial(posts []PostSummary, nextPage int, loadMoreURL string) {
      <!-- Post cards appended into #posts-list by HTMX -->
      for _, p := range posts {
          @PostCard(p)
      }

      <!-- OOB swap: replaces #load-more-btn with a new page or removes it if no more -->
      if len(posts) >= 10 && loadMoreURL != "" {
          <button
              id="load-more-btn"
              hx-swap-oob="true"
              class="rounded-md border border-gray-300 px-6 py-2 text-sm text-gray-600 hover:border-brand-600 hover:text-brand-600 transition-colors"
              hx-get={ loadMoreURL }
              hx-target="#posts-list"
              hx-swap="beforeend"
          >
              Load More
          </button>
      } else {
          <!-- Remove the load-more button: OOB swap with empty content -->
          <span id="load-more-btn" hx-swap-oob="true"></span>
      }
  }
  ```

- [ ] **5.4** Add a helper function `itoa` in a shared Go file (not a `.templ` file, since Templ does not generate arbitrary Go helpers):

  ```go
  // web/templates/helpers.go
  package templates

  import "strconv"

  // itoa converts an int to string for use in Templ expressions.
  func itoa(n int) string {
      return strconv.Itoa(n)
  }
  ```

- [ ] **5.5** Create `web/templates/archive_test.go`:

  ```go
  package templates_test

  import (
      "bytes"
      "context"
      "testing"
      "time"

      "github.com/sky-flux/cms/web/templates"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestCategoryArchivePage_RendersHeading(t *testing.T) {
      cfg := templates.SiteConfig{Name: "Blog"}
      posts := []templates.PostSummary{
          {Slug: "p1", Title: "Post 1", PublishedAt: time.Now()},
      }

      var buf bytes.Buffer
      err := templates.CategoryArchivePage(cfg, "Technology", "technology", posts, 1).Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, "Technology")
      assert.Contains(t, html, "Post 1")
      assert.NotContains(t, html, "load-more-btn") // fewer than 10 posts
  }

  func TestCategoryArchivePage_ShowsLoadMoreWhenFull(t *testing.T) {
      cfg := templates.SiteConfig{Name: "Blog"}
      posts := make([]templates.PostSummary, 10)
      for i := range posts {
          posts[i] = templates.PostSummary{Slug: "p", Title: "P", PublishedAt: time.Now()}
      }

      var buf bytes.Buffer
      err := templates.CategoryArchivePage(cfg, "Tech", "tech", posts, 1).Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, "load-more-btn")
      assert.Contains(t, html, `/categories/tech/partial?page=2`)
  }

  func TestTagArchivePage_RendersHeading(t *testing.T) {
      cfg := templates.SiteConfig{Name: "Blog"}
      posts := []templates.PostSummary{
          {Slug: "p1", Title: "Tagged Post", PublishedAt: time.Now()},
      }

      var buf bytes.Buffer
      err := templates.TagArchivePage(cfg, "Go", "go", posts, 1).Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, "Go")
      assert.Contains(t, html, "Tagged Post")
  }

  func TestPostsPartial_AppendsCardsAndOOBButton(t *testing.T) {
      posts := []templates.PostSummary{
          {Slug: "p1", Title: "Post 1", PublishedAt: time.Now()},
      }

      var buf bytes.Buffer
      err := templates.PostsPartial(posts, 3, "/categories/tech/partial?page=3").Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, "Post 1")
      // With fewer than 10 posts, the OOB removes the button (empty span)
      assert.Contains(t, html, `hx-swap-oob="true"`)
  }
  ```

---

## Task 6: Search Template

The search page uses HTMX for live results. The initial `GET /search?q=` renders the full page with results inline. Subsequent keystrokes in the Nav search bar trigger `hx-get="/search"` which returns only the `SearchResults` partial (the handler detects `HX-Request: true`).

- [ ] **6.1** Create `web/templates/search.templ`:

  ```go
  package templates

  // SearchPage renders the full search results page (initial GET).
  templ SearchPage(cfg SiteConfig, query string, results []PostSummary) {
      @Layout(cfg, "Search: "+query) {
          <section class="mb-8">
              <h1 class="text-2xl font-bold text-gray-900">
                  if query != "" {
                      Search results for "{ query }"
                  } else {
                      Search
                  }
              </h1>

              <!-- Standalone search bar on the results page with live HTMX updates.
                   hx-get="/search" — sends q param to the server.
                   hx-target="#search-results" — replaces just the results list.
                   hx-push-url="true" — updates the URL bar so results are bookmarkable.
                   hx-trigger="input changed delay:300ms, search" covers both keystrokes and clear. -->
              <form class="mt-4" action="/search" method="get" role="search">
                  <input
                      type="search"
                      name="q"
                      value={ query }
                      placeholder="Type to search…"
                      autofocus
                      class="w-full rounded-xl border border-gray-300 px-4 py-3 text-lg focus:outline-none focus:ring-2 focus:ring-brand-500"
                      hx-get="/search"
                      hx-trigger="input changed delay:300ms, search"
                      hx-target="#search-results"
                      hx-push-url="true"
                      hx-include="this"
                  />
              </form>
          </section>

          <!-- Results container; HTMX replaces innerHTML -->
          <div id="search-results">
              @SearchResults(query, results)
          </div>
      }
  }

  // SearchResults is the HTMX partial fragment returned for live search updates.
  // The handler returns ONLY this component when HX-Request header is present.
  templ SearchResults(query string, results []PostSummary) {
      if len(results) == 0 {
          <p class="text-gray-500 text-sm">
              if query != "" {
                  No results for "{ query }". Try different keywords.
              } else {
                  Start typing to search posts…
              }
          </p>
      } else {
          <p class="mb-4 text-sm text-gray-400">{ itoa(len(results)) } result(s) found</p>
          <div class="grid gap-6 sm:grid-cols-2">
              for _, p := range results {
                  @PostCard(p)
              }
          </div>
      }
  }
  ```

- [ ] **6.2** Create `web/templates/search_test.go`:

  ```go
  package templates_test

  import (
      "bytes"
      "context"
      "testing"
      "time"

      "github.com/sky-flux/cms/web/templates"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestSearchPage_RendersInputWithQuery(t *testing.T) {
      cfg := templates.SiteConfig{Name: "Blog"}
      results := []templates.PostSummary{
          {Slug: "found-post", Title: "Found Post", PublishedAt: time.Now()},
      }

      var buf bytes.Buffer
      err := templates.SearchPage(cfg, "golang", results).Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, `value="golang"`)
      assert.Contains(t, html, `hx-get="/search"`)
      assert.Contains(t, html, `hx-push-url="true"`)
      assert.Contains(t, html, "Found Post")
      assert.Contains(t, html, `Search results for "golang"`)
  }

  func TestSearchResults_ShowsEmptyState(t *testing.T) {
      var buf bytes.Buffer
      err := templates.SearchResults("noresult", nil).Render(context.Background(), &buf)
      require.NoError(t, err)

      assert.Contains(t, buf.String(), `No results for "noresult"`)
  }

  func TestSearchResults_ShowsCountAndCards(t *testing.T) {
      results := []templates.PostSummary{
          {Slug: "r1", Title: "Result 1", PublishedAt: time.Now()},
          {Slug: "r2", Title: "Result 2", PublishedAt: time.Now()},
      }

      var buf bytes.Buffer
      err := templates.SearchResults("go", results).Render(context.Background(), &buf)
      require.NoError(t, err)

      html := buf.String()
      assert.Contains(t, html, "2 result(s) found")
      assert.Contains(t, html, "Result 1")
      assert.Contains(t, html, "Result 2")
  }

  func TestSearchResults_EmptyQueryPrompt(t *testing.T) {
      var buf bytes.Buffer
      err := templates.SearchResults("", nil).Render(context.Background(), &buf)
      require.NoError(t, err)

      assert.Contains(t, buf.String(), "Start typing to search posts")
  }
  ```

---

## Task 7: Web Chi Handlers

`internal/delivery/web/handler.go` contains plain `http.HandlerFunc`-compatible methods. It depends on query interfaces (not concrete bun repos), so it can be tested with mock implementations using `httptest`.

- [ ] **7.1** Define query interfaces in `internal/delivery/web/queries.go`:

  ```go
  package web

  import (
      "context"

      "github.com/sky-flux/cms/web/templates"
  )

  // PostQuery fetches posts for web rendering. Implemented by infra layer.
  type PostQuery interface {
      // ListLatest returns the most recent published posts (page starts at 1).
      ListLatest(ctx context.Context, page, pageSize int) ([]templates.PostSummary, error)
      // GetBySlug returns a single published post by slug.
      GetBySlug(ctx context.Context, slug string) (*templates.PostDetail, error)
      // ListByCategory returns published posts for a category (page starts at 1).
      ListByCategory(ctx context.Context, categorySlug string, page, pageSize int) ([]templates.PostSummary, error)
      // ListByTag returns published posts for a tag (page starts at 1).
      ListByTag(ctx context.Context, tagSlug string, page, pageSize int) ([]templates.PostSummary, error)
      // Search returns posts matching query string via Meilisearch.
      Search(ctx context.Context, query string, limit int) ([]templates.PostSummary, error)
  }

  // CategoryQuery fetches category metadata for archive pages.
  type CategoryQuery interface {
      GetBySlug(ctx context.Context, slug string) (name string, err error)
  }

  // TagQuery fetches tag metadata for archive pages.
  type TagQuery interface {
      GetBySlug(ctx context.Context, slug string) (name string, err error)
  }

  // CommentWriter submits a new comment for moderation.
  type CommentWriter interface {
      Submit(ctx context.Context, postSlug, authorName, authorEmail, body string) error
  }

  // SiteConfigLoader loads site-wide settings (name, description, nav items).
  type SiteConfigLoader interface {
      Load(ctx context.Context) (templates.SiteConfig, error)
  }
  ```

- [ ] **7.2** Create `internal/delivery/web/handler.go`:

  ```go
  package web

  import (
      "log/slog"
      "net/http"
      "strconv"

      "github.com/go-chi/chi/v5"
      "github.com/sky-flux/cms/web/templates"
  )

  const defaultPageSize = 10

  // WebHandler handles HTTP requests for public Templ SSR pages.
  type WebHandler struct {
      postQuery    PostQuery
      categoryQuery CategoryQuery
      tagQuery     TagQuery
      commentWriter CommentWriter
      siteConfig   SiteConfigLoader
      log          *slog.Logger
  }

  // NewWebHandler constructs a WebHandler with all required dependencies.
  func NewWebHandler(
      postQuery PostQuery,
      categoryQuery CategoryQuery,
      tagQuery TagQuery,
      commentWriter CommentWriter,
      siteConfig SiteConfigLoader,
      log *slog.Logger,
  ) *WebHandler {
      return &WebHandler{
          postQuery:    postQuery,
          categoryQuery: categoryQuery,
          tagQuery:     tagQuery,
          commentWriter: commentWriter,
          siteConfig:   siteConfig,
          log:          log,
      }
  }

  // Home renders the homepage with the latest posts.
  // GET /
  func (h *WebHandler) Home(w http.ResponseWriter, r *http.Request) {
      cfg, err := h.siteConfig.Load(r.Context())
      if err != nil {
          h.serverError(w, r, err)
          return
      }
      posts, err := h.postQuery.ListLatest(r.Context(), 1, defaultPageSize)
      if err != nil {
          h.serverError(w, r, err)
          return
      }
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      templates.HomePage(cfg, posts).Render(r.Context(), w)
  }

  // PostDetail renders a single post page.
  // GET /posts/:slug
  func (h *WebHandler) PostDetail(w http.ResponseWriter, r *http.Request) {
      slug := chi.URLParam(r, "slug")
      cfg, err := h.siteConfig.Load(r.Context())
      if err != nil {
          h.serverError(w, r, err)
          return
      }
      post, err := h.postQuery.GetBySlug(r.Context(), slug)
      if err != nil {
          h.notFound(w, r)
          return
      }
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      templates.PostPage(cfg, *post).Render(r.Context(), w)
  }

  // PostsPartial returns only the post card fragment for HTMX infinite scroll.
  // GET /posts/partial?page=N
  func (h *WebHandler) PostsPartial(w http.ResponseWriter, r *http.Request) {
      page := parsePageParam(r)
      posts, err := h.postQuery.ListLatest(r.Context(), page, defaultPageSize)
      if err != nil {
          h.serverError(w, r, err)
          return
      }
      nextURL := ""
      if len(posts) >= defaultPageSize {
          nextURL = "/posts/partial?page=" + strconv.Itoa(page+1)
      }
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      templates.PostsPartial(posts, page+1, nextURL).Render(r.Context(), w)
  }

  // CategoryArchive renders the category archive page.
  // GET /categories/:slug
  func (h *WebHandler) CategoryArchive(w http.ResponseWriter, r *http.Request) {
      slug := chi.URLParam(r, "slug")
      page := parsePageParam(r)
      cfg, err := h.siteConfig.Load(r.Context())
      if err != nil {
          h.serverError(w, r, err)
          return
      }
      name, err := h.categoryQuery.GetBySlug(r.Context(), slug)
      if err != nil {
          h.notFound(w, r)
          return
      }
      posts, err := h.postQuery.ListByCategory(r.Context(), slug, page, defaultPageSize)
      if err != nil {
          h.serverError(w, r, err)
          return
      }
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      templates.CategoryArchivePage(cfg, name, slug, posts, page).Render(r.Context(), w)
  }

  // CategoryPartial returns HTMX post card fragment for category pagination.
  // GET /categories/:slug/partial?page=N
  func (h *WebHandler) CategoryPartial(w http.ResponseWriter, r *http.Request) {
      slug := chi.URLParam(r, "slug")
      page := parsePageParam(r)
      posts, err := h.postQuery.ListByCategory(r.Context(), slug, page, defaultPageSize)
      if err != nil {
          h.serverError(w, r, err)
          return
      }
      nextURL := ""
      if len(posts) >= defaultPageSize {
          nextURL = "/categories/" + slug + "/partial?page=" + strconv.Itoa(page+1)
      }
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      templates.PostsPartial(posts, page+1, nextURL).Render(r.Context(), w)
  }

  // TagArchive renders the tag archive page.
  // GET /tags/:slug
  func (h *WebHandler) TagArchive(w http.ResponseWriter, r *http.Request) {
      slug := chi.URLParam(r, "slug")
      page := parsePageParam(r)
      cfg, err := h.siteConfig.Load(r.Context())
      if err != nil {
          h.serverError(w, r, err)
          return
      }
      name, err := h.tagQuery.GetBySlug(r.Context(), slug)
      if err != nil {
          h.notFound(w, r)
          return
      }
      posts, err := h.postQuery.ListByTag(r.Context(), slug, page, defaultPageSize)
      if err != nil {
          h.serverError(w, r, err)
          return
      }
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      templates.TagArchivePage(cfg, name, slug, posts, page).Render(r.Context(), w)
  }

  // TagPartial returns HTMX post card fragment for tag pagination.
  // GET /tags/:slug/partial?page=N
  func (h *WebHandler) TagPartial(w http.ResponseWriter, r *http.Request) {
      slug := chi.URLParam(r, "slug")
      page := parsePageParam(r)
      posts, err := h.postQuery.ListByTag(r.Context(), slug, page, defaultPageSize)
      if err != nil {
          h.serverError(w, r, err)
          return
      }
      nextURL := ""
      if len(posts) >= defaultPageSize {
          nextURL = "/tags/" + slug + "/partial?page=" + strconv.Itoa(page+1)
      }
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      templates.PostsPartial(posts, page+1, nextURL).Render(r.Context(), w)
  }

  // Search renders the search results page or partial.
  // GET /search?q=
  // When HX-Request header is present, returns SearchResults partial only.
  func (h *WebHandler) Search(w http.ResponseWriter, r *http.Request) {
      query := r.URL.Query().Get("q")
      cfg, _ := h.siteConfig.Load(r.Context())

      var results []templates.PostSummary
      if query != "" {
          var err error
          results, err = h.postQuery.Search(r.Context(), query, 20)
          if err != nil {
              h.log.ErrorContext(r.Context(), "search failed", "query", query, "err", err)
              results = nil
          }
      }

      w.Header().Set("Content-Type", "text/html; charset=utf-8")

      // HTMX partial: return only the results fragment
      if r.Header.Get("HX-Request") == "true" {
          templates.SearchResults(query, results).Render(r.Context(), w)
          return
      }

      templates.SearchPage(cfg, query, results).Render(r.Context(), w)
  }

  // SubmitComment handles HTMX comment form submission.
  // POST /posts/:slug/comments
  // Returns an HTML fragment (success or error message) for #comment-form-status.
  func (h *WebHandler) SubmitComment(w http.ResponseWriter, r *http.Request) {
      slug := chi.URLParam(r, "slug")
      if err := r.ParseForm(); err != nil {
          w.WriteHeader(http.StatusBadRequest)
          w.Write([]byte(`<span class="text-red-600">Invalid form submission.</span>`))
          return
      }
      name := r.FormValue("author_name")
      email := r.FormValue("author_email")
      body := r.FormValue("body")

      if name == "" || email == "" || body == "" {
          w.WriteHeader(http.StatusUnprocessableEntity)
          w.Write([]byte(`<span class="text-red-600">All fields are required.</span>`))
          return
      }

      if err := h.commentWriter.Submit(r.Context(), slug, name, email, body); err != nil {
          h.log.ErrorContext(r.Context(), "comment submit failed", "slug", slug, "err", err)
          w.WriteHeader(http.StatusInternalServerError)
          w.Write([]byte(`<span class="text-red-600">Failed to submit comment. Please try again.</span>`))
          return
      }

      w.WriteHeader(http.StatusCreated)
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      w.Write([]byte(`<span class="text-green-600">Thank you! Your comment is awaiting moderation.</span>`))
  }

  // CustomPage handles /:slug for Page-type posts (e.g. /about).
  // GET /:slug
  func (h *WebHandler) CustomPage(w http.ResponseWriter, r *http.Request) {
      slug := chi.URLParam(r, "slug")
      cfg, err := h.siteConfig.Load(r.Context())
      if err != nil {
          h.serverError(w, r, err)
          return
      }
      post, err := h.postQuery.GetBySlug(r.Context(), slug)
      if err != nil {
          h.notFound(w, r)
          return
      }
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      templates.PostPage(cfg, *post).Render(r.Context(), w)
  }

  // serverError writes a plain 500 response. In production, render a Templ 500 page instead.
  func (h *WebHandler) serverError(w http.ResponseWriter, r *http.Request, err error) {
      h.log.ErrorContext(r.Context(), "internal server error", "err", err)
      http.Error(w, "Internal Server Error", http.StatusInternalServerError)
  }

  // notFound writes a plain 404 response. In production, render a Templ 404 page instead.
  func (h *WebHandler) notFound(w http.ResponseWriter, r *http.Request) {
      http.Error(w, "Not Found", http.StatusNotFound)
  }

  // parsePageParam extracts the ?page= query param, defaulting to 1.
  func parsePageParam(r *http.Request) int {
      p, err := strconv.Atoi(r.URL.Query().Get("page"))
      if err != nil || p < 1 {
          return 1
      }
      return p
  }
  ```

- [ ] **7.3** Create `internal/delivery/web/handler_test.go`:

  ```go
  package web_test

  import (
      "context"
      "log/slog"
      "net/http"
      "net/http/httptest"
      "strings"
      "testing"
      "time"

      "github.com/go-chi/chi/v5"
      "github.com/sky-flux/cms/internal/delivery/web"
      "github.com/sky-flux/cms/web/templates"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  // ─── Mock implementations ───────────────────────────────────────────────────

  type mockPostQuery struct {
      listLatest    func(ctx context.Context, page, size int) ([]templates.PostSummary, error)
      getBySlug     func(ctx context.Context, slug string) (*templates.PostDetail, error)
      listByCategory func(ctx context.Context, slug string, page, size int) ([]templates.PostSummary, error)
      listByTag     func(ctx context.Context, slug string, page, size int) ([]templates.PostSummary, error)
      search        func(ctx context.Context, query string, limit int) ([]templates.PostSummary, error)
  }

  func (m *mockPostQuery) ListLatest(ctx context.Context, page, size int) ([]templates.PostSummary, error) {
      if m.listLatest != nil {
          return m.listLatest(ctx, page, size)
      }
      return nil, nil
  }
  func (m *mockPostQuery) GetBySlug(ctx context.Context, slug string) (*templates.PostDetail, error) {
      if m.getBySlug != nil {
          return m.getBySlug(ctx, slug)
      }
      return nil, nil
  }
  func (m *mockPostQuery) ListByCategory(ctx context.Context, slug string, page, size int) ([]templates.PostSummary, error) {
      if m.listByCategory != nil {
          return m.listByCategory(ctx, slug, page, size)
      }
      return nil, nil
  }
  func (m *mockPostQuery) ListByTag(ctx context.Context, slug string, page, size int) ([]templates.PostSummary, error) {
      if m.listByTag != nil {
          return m.listByTag(ctx, slug, page, size)
      }
      return nil, nil
  }
  func (m *mockPostQuery) Search(ctx context.Context, query string, limit int) ([]templates.PostSummary, error) {
      if m.search != nil {
          return m.search(ctx, query, limit)
      }
      return nil, nil
  }

  type mockCategoryQuery struct {
      getBySlug func(ctx context.Context, slug string) (string, error)
  }

  func (m *mockCategoryQuery) GetBySlug(ctx context.Context, slug string) (string, error) {
      if m.getBySlug != nil {
          return m.getBySlug(ctx, slug)
      }
      return slug, nil
  }

  type mockTagQuery struct {
      getBySlug func(ctx context.Context, slug string) (string, error)
  }

  func (m *mockTagQuery) GetBySlug(ctx context.Context, slug string) (string, error) {
      if m.getBySlug != nil {
          return m.getBySlug(ctx, slug)
      }
      return slug, nil
  }

  type mockCommentWriter struct {
      submit func(ctx context.Context, postSlug, name, email, body string) error
  }

  func (m *mockCommentWriter) Submit(ctx context.Context, postSlug, name, email, body string) error {
      if m.submit != nil {
          return m.submit(ctx, postSlug, name, email, body)
      }
      return nil
  }

  type mockSiteConfig struct {
      load func(ctx context.Context) (templates.SiteConfig, error)
  }

  func (m *mockSiteConfig) Load(ctx context.Context) (templates.SiteConfig, error) {
      if m.load != nil {
          return m.load(ctx)
      }
      return templates.SiteConfig{Name: "Test Site"}, nil
  }

  // ─── Helper ──────────────────────────────────────────────────────────────────

  func newHandler(pq web.PostQuery, cq web.CategoryQuery, tq web.TagQuery, cw web.CommentWriter, sc web.SiteConfigLoader) *web.WebHandler {
      return web.NewWebHandler(pq, cq, tq, cw, sc, slog.Default())
  }

  // ─── Tests ───────────────────────────────────────────────────────────────────

  func TestHome_Returns200WithPosts(t *testing.T) {
      pq := &mockPostQuery{
          listLatest: func(_ context.Context, page, size int) ([]templates.PostSummary, error) {
              return []templates.PostSummary{
                  {Slug: "hello", Title: "Hello World", PublishedAt: time.Now()},
              }, nil
          },
      }
      h := newHandler(pq, &mockCategoryQuery{}, &mockTagQuery{}, &mockCommentWriter{}, &mockSiteConfig{})

      req := httptest.NewRequest(http.MethodGet, "/", nil)
      w := httptest.NewRecorder()
      h.Home(w, req)

      res := w.Result()
      require.Equal(t, http.StatusOK, res.StatusCode)
      assert.Contains(t, w.Body.String(), "Hello World")
  }

  func TestPostDetail_Returns200ForValidSlug(t *testing.T) {
      pq := &mockPostQuery{
          getBySlug: func(_ context.Context, slug string) (*templates.PostDetail, error) {
              return &templates.PostDetail{
                  Slug:        slug,
                  Title:       "Test Post",
                  BodyHTML:    "<p>Content</p>",
                  PublishedAt: time.Now(),
              }, nil
          },
      }
      h := newHandler(pq, &mockCategoryQuery{}, &mockTagQuery{}, &mockCommentWriter{}, &mockSiteConfig{})

      r := chi.NewRouter()
      r.Get("/posts/{slug}", h.PostDetail)
      req := httptest.NewRequest(http.MethodGet, "/posts/test-post", nil)
      w := httptest.NewRecorder()
      r.ServeHTTP(w, req)

      require.Equal(t, http.StatusOK, w.Result().StatusCode)
      assert.Contains(t, w.Body.String(), "Test Post")
      assert.Contains(t, w.Body.String(), "<p>Content</p>")
  }

  func TestPostDetail_Returns404ForMissingPost(t *testing.T) {
      pq := &mockPostQuery{
          getBySlug: func(_ context.Context, slug string) (*templates.PostDetail, error) {
              return nil, fmt.Errorf("not found")
          },
      }
      h := newHandler(pq, &mockCategoryQuery{}, &mockTagQuery{}, &mockCommentWriter{}, &mockSiteConfig{})

      r := chi.NewRouter()
      r.Get("/posts/{slug}", h.PostDetail)
      req := httptest.NewRequest(http.MethodGet, "/posts/missing", nil)
      w := httptest.NewRecorder()
      r.ServeHTTP(w, req)

      assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
  }

  func TestSearch_FullPageWithoutHTMXHeader(t *testing.T) {
      pq := &mockPostQuery{
          search: func(_ context.Context, query string, limit int) ([]templates.PostSummary, error) {
              return []templates.PostSummary{
                  {Slug: "r1", Title: "Result 1", PublishedAt: time.Now()},
              }, nil
          },
      }
      h := newHandler(pq, &mockCategoryQuery{}, &mockTagQuery{}, &mockCommentWriter{}, &mockSiteConfig{})

      req := httptest.NewRequest(http.MethodGet, "/search?q=golang", nil)
      w := httptest.NewRecorder()
      h.Search(w, req)

      body := w.Body.String()
      assert.Equal(t, http.StatusOK, w.Result().StatusCode)
      assert.Contains(t, body, "<!DOCTYPE html>") // full page
      assert.Contains(t, body, "Result 1")
  }

  func TestSearch_PartialWithHTMXHeader(t *testing.T) {
      pq := &mockPostQuery{
          search: func(_ context.Context, query string, limit int) ([]templates.PostSummary, error) {
              return []templates.PostSummary{
                  {Slug: "r1", Title: "HTMX Result", PublishedAt: time.Now()},
              }, nil
          },
      }
      h := newHandler(pq, &mockCategoryQuery{}, &mockTagQuery{}, &mockCommentWriter{}, &mockSiteConfig{})

      req := httptest.NewRequest(http.MethodGet, "/search?q=htmx", nil)
      req.Header.Set("HX-Request", "true")
      w := httptest.NewRecorder()
      h.Search(w, req)

      body := w.Body.String()
      assert.Equal(t, http.StatusOK, w.Result().StatusCode)
      assert.NotContains(t, body, "<!DOCTYPE html>") // partial only
      assert.Contains(t, body, "HTMX Result")
  }

  func TestSubmitComment_Returns201OnSuccess(t *testing.T) {
      cw := &mockCommentWriter{
          submit: func(_ context.Context, slug, name, email, body string) error {
              return nil
          },
      }
      h := newHandler(&mockPostQuery{}, &mockCategoryQuery{}, &mockTagQuery{}, cw, &mockSiteConfig{})

      r := chi.NewRouter()
      r.Post("/posts/{slug}/comments", h.SubmitComment)

      form := strings.NewReader("author_name=Alice&author_email=alice@example.com&body=Great+post!")
      req := httptest.NewRequest(http.MethodPost, "/posts/hello/comments", form)
      req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
      w := httptest.NewRecorder()
      r.ServeHTTP(w, req)

      assert.Equal(t, http.StatusCreated, w.Result().StatusCode)
      assert.Contains(t, w.Body.String(), "awaiting moderation")
  }

  func TestSubmitComment_Returns422WhenFieldsMissing(t *testing.T) {
      h := newHandler(&mockPostQuery{}, &mockCategoryQuery{}, &mockTagQuery{}, &mockCommentWriter{}, &mockSiteConfig{})

      r := chi.NewRouter()
      r.Post("/posts/{slug}/comments", h.SubmitComment)

      form := strings.NewReader("author_name=Alice") // missing email and body
      req := httptest.NewRequest(http.MethodPost, "/posts/hello/comments", form)
      req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
      w := httptest.NewRecorder()
      r.ServeHTTP(w, req)

      assert.Equal(t, http.StatusUnprocessableEntity, w.Result().StatusCode)
  }

  func TestCategoryArchive_Returns200(t *testing.T) {
      cq := &mockCategoryQuery{
          getBySlug: func(_ context.Context, slug string) (string, error) {
              return "Technology", nil
          },
      }
      pq := &mockPostQuery{
          listByCategory: func(_ context.Context, slug string, page, size int) ([]templates.PostSummary, error) {
              return []templates.PostSummary{
                  {Slug: "p1", Title: "Tech Post", PublishedAt: time.Now()},
              }, nil
          },
      }
      h := newHandler(pq, cq, &mockTagQuery{}, &mockCommentWriter{}, &mockSiteConfig{})

      r := chi.NewRouter()
      r.Get("/categories/{slug}", h.CategoryArchive)
      req := httptest.NewRequest(http.MethodGet, "/categories/technology", nil)
      w := httptest.NewRecorder()
      r.ServeHTTP(w, req)

      require.Equal(t, http.StatusOK, w.Result().StatusCode)
      body := w.Body.String()
      assert.Contains(t, body, "Technology")
      assert.Contains(t, body, "Tech Post")
  }

  func TestTagArchive_Returns200(t *testing.T) {
      tq := &mockTagQuery{
          getBySlug: func(_ context.Context, slug string) (string, error) {
              return "Go", nil
          },
      }
      pq := &mockPostQuery{
          listByTag: func(_ context.Context, slug string, page, size int) ([]templates.PostSummary, error) {
              return []templates.PostSummary{
                  {Slug: "p1", Title: "Go Post", PublishedAt: time.Now()},
              }, nil
          },
      }
      h := newHandler(pq, &mockCategoryQuery{}, tq, &mockCommentWriter{}, &mockSiteConfig{})

      r := chi.NewRouter()
      r.Get("/tags/{slug}", h.TagArchive)
      req := httptest.NewRequest(http.MethodGet, "/tags/go", nil)
      w := httptest.NewRecorder()
      r.ServeHTTP(w, req)

      require.Equal(t, http.StatusOK, w.Result().StatusCode)
      assert.Contains(t, w.Body.String(), "Go Post")
  }
  ```

  > **Note:** `handler_test.go` needs `"fmt"` imported for `fmt.Errorf`. Add it to the import block of the test file.

- [ ] **7.4** Run tests (RED phase — compilation will fail until `templ generate` is run):
  ```bash
  templ generate ./web/templates/
  go test ./internal/delivery/web/... -v -count=1
  ```

---

## Task 8: Wire Web Routes into Chi Router

Web routes are registered **last** in the router, after `/api` and `/console`, so they never shadow API endpoints.

- [ ] **8.1** Locate the main Chi router file. Based on the project structure it is at `internal/delivery/` or similar. Add a `RegisterWebRoutes` function in a new file `internal/delivery/web/routes.go`:

  ```go
  package web

  import "github.com/go-chi/chi/v5"

  // RegisterRoutes mounts all public website routes onto the given router.
  // IMPORTANT: Call this AFTER mounting /api and /console routes.
  func (h *WebHandler) RegisterRoutes(r chi.Router) {
      // Static assets (app.css, htmx.min.js) served by the embed handler upstream.
      // These routes handle the HTML pages only.

      r.Get("/", h.Home)

      // Posts
      r.Get("/posts/{slug}", h.PostDetail)
      r.Get("/posts/partial", h.PostsPartial) // HTMX infinite scroll fragment

      // Category archive + HTMX pagination fragment
      r.Get("/categories/{slug}", h.CategoryArchive)
      r.Get("/categories/{slug}/partial", h.CategoryPartial)

      // Tag archive + HTMX pagination fragment
      r.Get("/tags/{slug}", h.TagArchive)
      r.Get("/tags/{slug}/partial", h.TagPartial)

      // Search (full page + HTMX partial via HX-Request header detection)
      r.Get("/search", h.Search)

      // Comment submission (HTMX form POST)
      r.Post("/posts/{slug}/comments", h.SubmitComment)

      // Custom pages / catch-all (must be LAST)
      r.Get("/{slug}", h.CustomPage)
  }
  ```

- [ ] **8.2** In the main router assembly (e.g., `internal/delivery/router.go` or `cmd/cms/serve.go`), import the web package and wire it up:

  ```go
  // After mounting /api and /console:

  // Static files from go:embed
  staticFS, _ := fs.Sub(rootfs.WebStaticFS, "web/static")
  r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

  // Public website (Templ SSR) — must come after /api and /console
  webHandler := web.NewWebHandler(
      postQueryImpl,      // *infra.BunPostQuery
      categoryQueryImpl,  // *infra.BunCategoryQuery
      tagQueryImpl,       // *infra.BunTagQuery
      commentWriterImpl,  // *infra.BunCommentWriter
      siteConfigImpl,     // *infra.BunSiteConfigLoader
      logger,
  )
  webHandler.RegisterRoutes(r)
  ```

- [ ] **8.3** Verify route order with a route list print (optional, useful during dev):
  ```bash
  go run ./cmd/cms serve --list-routes 2>&1 | grep -E "^(GET|POST)"
  ```
  Confirm `/api/v1/*` appears before `/posts/*` and `/*`.

- [ ] **8.4** Run the full test suite:
  ```bash
  go test ./... -count=1 -short
  ```

---

## Task 9: Tailwind + HTMX Assets Verification

Ensure the full asset pipeline produces correct output and `go:embed` picks up the files at build time.

- [ ] **9.1** Run the Tailwind build and confirm `web/static/app.css` is generated:
  ```bash
  tailwindcss -i web/styles/input.css -o web/static/app.css --minify
  ls -lh web/static/app.css
  ```
  Expected: non-empty CSS file containing Tailwind utilities.

- [ ] **9.2** Confirm HTMX is present:
  ```bash
  ls -lh web/static/htmx.min.js
  head -c 100 web/static/htmx.min.js  # should start with "/*! htmx.org..."
  ```

- [ ] **9.3** Verify `go:embed` compiles successfully with both files present:
  ```bash
  go build ./...
  ```
  If this fails with "pattern web/static: no matching files", ensure `.gitkeep` is committed and actual files exist.

- [ ] **9.4** Write an embed verification test in `embed_test.go` at project root:

  ```go
  package cms_test

  import (
      "io/fs"
      "testing"

      rootfs "github.com/sky-flux/cms"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestWebStaticFS_ContainsRequiredFiles(t *testing.T) {
      sub, err := fs.Sub(rootfs.WebStaticFS, "web/static")
      require.NoError(t, err)

      files := []string{"app.css", "htmx.min.js"}
      for _, f := range files {
          _, err := sub.Open(f)
          assert.NoError(t, err, "expected %s in WebStaticFS", f)
      }
  }
  ```

- [ ] **9.5** Confirm Tailwind scans `.templ` files for utility classes. Tailwind V4 CLI auto-discovers content files via the `@import "tailwindcss"` directive; no explicit `content` config is needed. Verify by adding a class in a `.templ` file and checking it appears in the output CSS:
  ```bash
  # Add a unique class e.g. `text-brand-600` in a .templ file, rebuild:
  tailwindcss -i web/styles/input.css -o web/static/app.css
  grep "brand-600" web/static/app.css
  ```
  > **Note:** If Tailwind V4's auto-scanner doesn't pick up `.templ` files, add an explicit source hint in `input.css`:
  > ```css
  > @source "../templates/**/*.templ";
  > ```

- [ ] **9.6** Test the static file server in isolation:

  ```go
  // internal/delivery/web/static_test.go
  package web_test

  import (
      "io/fs"
      "net/http"
      "net/http/httptest"
      "testing"

      rootfs "github.com/sky-flux/cms"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestStaticFileServer_ServesCSS(t *testing.T) {
      sub, err := fs.Sub(rootfs.WebStaticFS, "web/static")
      require.NoError(t, err)

      r := http.NewServeMux()
      r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))

      req := httptest.NewRequest(http.MethodGet, "/static/app.css", nil)
      w := httptest.NewRecorder()
      r.ServeHTTP(w, req)

      assert.Equal(t, http.StatusOK, w.Result().StatusCode)
      assert.Contains(t, w.Header().Get("Content-Type"), "text/css")
  }
  ```

---

## Acceptance Criteria

All of the following must pass before this plan is considered complete:

- [ ] `templ generate ./web/templates/` runs without errors and produces `*_templ.go` files
- [ ] `go test ./web/templates/... -v -count=1` — all Templ unit tests GREEN
- [ ] `go test ./internal/delivery/web/... -v -count=1` — all handler tests GREEN
- [ ] `go build ./...` — full binary compiles with embedded assets
- [ ] `ls web/static/app.css web/static/htmx.min.js` — both files exist
- [ ] Manual smoke test: `go run ./cmd/cms serve` → `curl http://localhost:8080/` returns HTTP 200 with HTML

---

## HTMX Interaction Summary

| Interaction | Trigger | HTMX Attributes | Server Response |
|-------------|---------|-----------------|-----------------|
| Nav search | 300ms debounce on input | `hx-get="/search"`, `hx-trigger="input changed delay:300ms"`, `hx-target="#search-results"` | `SearchResults` partial HTML |
| Homepage infinite scroll | Intersection observer on sentinel | `hx-get="/posts/partial?page=N"`, `hx-trigger="intersect once"`, `hx-swap="beforeend"` | `PostsPartial` cards + OOB button swap |
| Category/Tag "Load More" | Button click | `hx-get="/categories/:slug/partial?page=N"`, `hx-swap="beforeend"` | `PostsPartial` cards + OOB button swap |
| Comment submission | Form submit | `hx-post="/posts/:slug/comments"`, `hx-target="#comment-form-status"` | HTML `<span>` success/error message |
| `hx-boost` | All links/forms in `<body>` | `hx-boost="true"` on `<body>` | Full page (Templ renders it, HTMX swaps `<body>`) |

---

## Key Implementation Notes

1. **Templ safety:** Always use `templ.SafeURL()` when constructing `href` attributes from data. Use `templ.Raw()` only for `BodyHTML` which comes from the CMS editor (already sanitized server-side).

2. **HTMX partial detection:** The `h.Search` handler checks `r.Header.Get("HX-Request") == "true"` to decide between full page and partial. The same pattern applies anywhere the handler can serve both.

3. **OOB swaps:** `PostsPartial` uses `hx-swap-oob="true"` on the load-more button to replace it with the next-page URL. When there are no more results, it swaps in an empty `<span>` to remove the button.

4. **Comment form reset:** The form uses `hx-on::after-request="if(event.detail.successful) this.reset()"` to clear fields on success without a full page reload.

5. **`hx-disabled-elt`:** The submit button uses `hx-disabled-elt="this"` to prevent double-submission while the HTMX request is in flight.

6. **Content-Type header:** Every handler must set `w.Header().Set("Content-Type", "text/html; charset=utf-8")` before calling `.Render()`. Templ does not set this automatically.

7. **Page size constant:** `defaultPageSize = 10` drives the "show load-more" decision. Adjust in `handler.go` if needed.

8. **`itoa` helper:** Templ expressions cannot call `strconv.Itoa` directly. The `web/templates/helpers.go` file exports `itoa(n int) string` for use in `.templ` files.

9. **Tailwind V4 source scanning:** If custom `.templ` class utilities are not appearing in the output CSS, add `@source "../templates/**/*.templ";` to `web/styles/input.css`.

10. **embed.go root package:** The file `embed.go` uses `package cms` (root package). The `cmd/cms/main.go` imports it as `rootfs "github.com/sky-flux/cms"` to access `rootfs.WebStaticFS`.

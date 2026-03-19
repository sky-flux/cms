# Delivery BC Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Delivery bounded context for RSS/Atom feeds, XML sitemaps, and public-facing JSON API.

**Architecture:** Feed handlers use plain Chi (return XML/text). Public API uses Huma (return JSON). Both use query interfaces injected from content/site BCs. API Key auth is optional middleware.

**Tech Stack:** Go 1.25+, Chi v5, Huma v2, encoding/xml, testify + httptest

---

## Prerequisites

Before starting, verify the new DDD directory skeleton exists (or create it):

```bash
mkdir -p internal/delivery/feed
mkdir -p internal/delivery/public
```

Key references from existing codebase:
- `internal/feed/types.go` — existing RSS/Atom/Sitemap XML struct definitions (reuse field names and xml tags)
- `internal/feed/interfaces.go` — `FeedPostReader`, `FeedCategoryReader`, `FeedTagReader`, `SiteConfigReader` (port to delivery BC)
- `internal/public/interfaces.go` — `PostReader`, `CategoryReader`, `TagReader`, `CommentReader`, `MenuReader` (port to delivery BC)
- `internal/identity/delivery/handler.go` — Huma handler pattern: `huma.Register`, `huma.Operation`, `huma.NewError`, typed request/response structs
- `internal/identity/domain/user.go` — domain entity pattern (pure Go, no framework deps)

Spec references: `docs/superpowers/specs/2026-03-19-project-redesign-design.md`
- Feed endpoints live under `/api/v1/public/feed/*` (RSS, Atom, Sitemap index + sub-sitemaps)
- Public API endpoints live under `/api/v1/public/` (posts, categories, tags, search)
- API Key auth is **optional** middleware (`OptionalAPIKey`) — requests without a key still succeed
- v1 single schema (`public`), table prefix `sfc_` (no `site_` infix)
- Chi v5 + Huma v2, **not Gin**
- koanf config (not Viper)

---

## TASK 1 — Feed domain types

**Files:**
- `internal/delivery/feed/feed.go`
- `internal/delivery/feed/feed_test.go`

**TDD cycle:**

- [ ] **RED** — Write `feed_test.go` verifying XML marshaling round-trips:
  ```go
  package feed_test

  import (
      "encoding/xml"
      "strings"
      "testing"

      "github.com/sky-flux/cms/internal/delivery/feed"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestRSSItem_MarshalXML(t *testing.T) {
      item := feed.RSSItem{
          Title:   "Hello World",
          Link:    "https://example.com/posts/hello",
          PubDate: "Mon, 01 Jan 2026 00:00:00 +0000",
          GUID: feed.GUID{
              IsPermaLink: "true",
              Value:       "https://example.com/posts/hello",
          },
      }
      data, err := xml.Marshal(item)
      require.NoError(t, err)
      assert.True(t, strings.Contains(string(data), "Hello World"))
      assert.True(t, strings.Contains(string(data), "https://example.com/posts/hello"))
  }

  func TestAtomEntry_MarshalXML(t *testing.T) {
      entry := feed.AtomEntry{
          Title: "Test Post",
          ID:    "urn:uuid:abc-123",
          Link: feed.AtomFeedLink{
              Href: "https://example.com/posts/test",
              Rel:  "alternate",
          },
          Updated: "2026-01-01T00:00:00Z",
      }
      data, err := xml.Marshal(entry)
      require.NoError(t, err)
      assert.True(t, strings.Contains(string(data), "Test Post"))
  }

  func TestSitemapURL_MarshalXML(t *testing.T) {
      u := feed.SitemapURL{
          Loc:        "https://example.com/posts/hello",
          Lastmod:    "2026-01-01",
          Changefreq: "weekly",
          Priority:   "0.8",
      }
      data, err := xml.Marshal(u)
      require.NoError(t, err)
      assert.True(t, strings.Contains(string(data), "https://example.com/posts/hello"))
      assert.True(t, strings.Contains(string(data), "weekly"))
  }
  ```

- [ ] **Verify RED** — `go test ./internal/delivery/feed/... -v -count=1` — must fail: "cannot find package"

- [ ] **GREEN** — Create `internal/delivery/feed/feed.go`:
  ```go
  package feed

  import "encoding/xml"

  // --- RSS 2.0 ---

  // RSSFeed is the root RSS 2.0 document.
  type RSSFeed struct {
      XMLName xml.Name   `xml:"rss"`
      Version string     `xml:"version,attr"`
      Atom    string     `xml:"xmlns:atom,attr"`
      DC      string     `xml:"xmlns:dc,attr"`
      Content string     `xml:"xmlns:content,attr"`
      Channel RSSChannel `xml:"channel"`
  }

  // RSSChannel holds feed metadata and items.
  type RSSChannel struct {
      Title         string    `xml:"title"`
      Link          string    `xml:"link"`
      Description   string    `xml:"description"`
      Language      string    `xml:"language"`
      LastBuildDate string    `xml:"lastBuildDate"`
      Generator     string    `xml:"generator"`
      AtomLink      AtomLink  `xml:"atom:link"`
      Items         []RSSItem `xml:"item"`
  }

  // AtomLink is the atom:link self-referencing element in RSS.
  type AtomLink struct {
      Href string `xml:"href,attr"`
      Rel  string `xml:"rel,attr"`
      Type string `xml:"type,attr"`
  }

  // RSSItem represents a single post entry in RSS.
  type RSSItem struct {
      Title          string   `xml:"title"`
      Link           string   `xml:"link"`
      GUID           GUID     `xml:"guid"`
      Description    string   `xml:"description"`
      ContentEncoded string   `xml:"content:encoded"`
      Creator        string   `xml:"dc:creator"`
      PubDate        string   `xml:"pubDate"`
      Categories     []string `xml:"category"`
  }

  // GUID is the globally unique identifier for an RSS item.
  type GUID struct {
      IsPermaLink string `xml:"isPermaLink,attr"`
      Value       string `xml:",chardata"`
  }

  // --- Atom 1.0 ---

  // AtomFeed is the root Atom 1.0 document.
  type AtomFeed struct {
      XMLName   xml.Name       `xml:"feed"`
      XMLNS     string         `xml:"xmlns,attr"`
      Title     string         `xml:"title"`
      Link      []AtomFeedLink `xml:"link"`
      Updated   string         `xml:"updated"`
      ID        string         `xml:"id"`
      Author    *AtomAuthor    `xml:"author,omitempty"`
      Generator string         `xml:"generator"`
      Entries   []AtomEntry    `xml:"entry"`
  }

  // AtomFeedLink is a link element within an Atom feed or entry.
  type AtomFeedLink struct {
      Href string `xml:"href,attr"`
      Rel  string `xml:"rel,attr,omitempty"`
      Type string `xml:"type,attr,omitempty"`
  }

  // AtomAuthor holds the author's name for Atom entries.
  type AtomAuthor struct {
      Name string `xml:"name"`
  }

  // AtomEntry represents a single post entry in Atom.
  type AtomEntry struct {
      Title     string       `xml:"title"`
      Link      AtomFeedLink `xml:"link"`
      ID        string       `xml:"id"`
      Updated   string       `xml:"updated"`
      Published string       `xml:"published"`
      Author    *AtomAuthor  `xml:"author,omitempty"`
      Summary   string       `xml:"summary,omitempty"`
      Content   *AtomContent `xml:"content,omitempty"`
  }

  // AtomContent wraps the full post content in CDATA.
  type AtomContent struct {
      Type  string `xml:"type,attr"`
      Value string `xml:",cdata"`
  }

  // --- Sitemap ---

  // SitemapIndex is the root sitemap index document.
  type SitemapIndex struct {
      XMLName  xml.Name       `xml:"sitemapindex"`
      XMLNS    string         `xml:"xmlns,attr"`
      Sitemaps []SitemapEntry `xml:"sitemap"`
  }

  // SitemapEntry is a single sitemap reference in an index.
  type SitemapEntry struct {
      Loc     string `xml:"loc"`
      Lastmod string `xml:"lastmod,omitempty"`
  }

  // URLSet is the root sitemap URL set document.
  type URLSet struct {
      XMLName xml.Name     `xml:"urlset"`
      XMLNS   string       `xml:"xmlns,attr"`
      URLs    []SitemapURL `xml:"url"`
  }

  // SitemapURL is a single URL entry in a sitemap.
  type SitemapURL struct {
      Loc        string `xml:"loc"`
      Lastmod    string `xml:"lastmod,omitempty"`
      Changefreq string `xml:"changefreq,omitempty"`
      Priority   string `xml:"priority,omitempty"`
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/delivery/feed/... -v -count=1` — all pass

- [ ] **REFACTOR** — Add package-level doc comment `// Package feed provides XML feed and sitemap types for the Delivery BC.`

- [ ] **Commit:** `git commit -m "✨ feat(delivery): add feed domain types (RSS/Atom/Sitemap)"`

---

## TASK 2 — Feed query interfaces

**Files:**
- `internal/delivery/feed/queries.go`
- `internal/delivery/feed/queries_test.go`

**TDD cycle:**

- [ ] **RED** — Write `queries_test.go` verifying interface satisfaction at compile time:
  ```go
  package feed_test

  import (
      "context"
      "testing"
      "time"

      "github.com/sky-flux/cms/internal/delivery/feed"
  )

  // Compile-time interface checks.
  var _ feed.FeedPostQuery = (*mockFeedPostQuery)(nil)
  var _ feed.FeedSiteQuery = (*mockFeedSiteQuery)(nil)

  type mockFeedPostQuery struct {
      listPublishedFn      func(ctx context.Context, limit int) ([]feed.FeedPost, error)
      latestPublishedAtFn  func(ctx context.Context) (*time.Time, error)
  }

  func (m *mockFeedPostQuery) ListPublished(ctx context.Context, limit int) ([]feed.FeedPost, error) {
      return m.listPublishedFn(ctx, limit)
  }
  func (m *mockFeedPostQuery) LatestPublishedAt(ctx context.Context) (*time.Time, error) {
      return m.latestPublishedAtFn(ctx)
  }

  type mockFeedSiteQuery struct {
      getTitleFn       func(ctx context.Context) string
      getURLFn         func(ctx context.Context) string
      getDescriptionFn func(ctx context.Context) string
      getLanguageFn    func(ctx context.Context) string
  }

  func (m *mockFeedSiteQuery) GetSiteTitle(ctx context.Context) string       { return m.getTitleFn(ctx) }
  func (m *mockFeedSiteQuery) GetSiteURL(ctx context.Context) string         { return m.getURLFn(ctx) }
  func (m *mockFeedSiteQuery) GetSiteDescription(ctx context.Context) string { return m.getDescriptionFn(ctx) }
  func (m *mockFeedSiteQuery) GetSiteLanguage(ctx context.Context) string    { return m.getLanguageFn(ctx) }

  func TestFeedQueryInterfaces_Compile(t *testing.T) {
      t.Log("FeedPostQuery and FeedSiteQuery interfaces satisfied")
  }
  ```

- [ ] **Verify RED** — fails: `feed.FeedPostQuery` undefined, `feed.FeedPost` undefined

- [ ] **GREEN** — Create `internal/delivery/feed/queries.go`:
  ```go
  package feed

  import (
      "context"
      "time"
  )

  // FeedPost is a minimal projection of a post for feed generation.
  // Avoids a direct dependency on the content BC's domain types.
  type FeedPost struct {
      ID          string
      Title       string
      Slug        string
      Excerpt     string
      Content     string
      AuthorName  string
      PublishedAt time.Time
      UpdatedAt   time.Time
      Categories  []string // category names
  }

  // FeedCategoryItem is a minimal projection of a category for sitemap generation.
  type FeedCategoryItem struct {
      Slug        string
      LastPostAt  *time.Time
  }

  // FeedTagItem is a minimal projection of a tag for sitemap generation.
  type FeedTagItem struct {
      Slug       string
      LastPostAt *time.Time
  }

  // FeedPostQuery is the read port the feed handler uses to fetch posts.
  // Implemented by the content BC's infra layer.
  type FeedPostQuery interface {
      // ListPublished returns the most recent published posts up to limit.
      ListPublished(ctx context.Context, limit int) ([]FeedPost, error)
      // LatestPublishedAt returns the time of the most recently published post.
      LatestPublishedAt(ctx context.Context) (*time.Time, error)
  }

  // FeedCategoryQuery is the read port for category sitemap data.
  type FeedCategoryQuery interface {
      // ListAll returns all categories that have at least one published post.
      ListAll(ctx context.Context) ([]FeedCategoryItem, error)
  }

  // FeedTagQuery is the read port for tag sitemap data.
  type FeedTagQuery interface {
      // ListWithPosts returns tags that have at least one published post.
      ListWithPosts(ctx context.Context) ([]FeedTagItem, error)
  }

  // FeedSiteQuery is the read port for site metadata used in feed headers.
  // Implemented by the platform BC's infra layer (reads sfc_configs).
  type FeedSiteQuery interface {
      GetSiteTitle(ctx context.Context) string
      GetSiteURL(ctx context.Context) string
      GetSiteDescription(ctx context.Context) string
      GetSiteLanguage(ctx context.Context) string
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/delivery/feed/... -v -count=1` — all pass

- [ ] **Commit:** `git commit -m "✨ feat(delivery): add feed query interfaces"`

---

## TASK 3 — Feed Chi handler (XML responses)

**Files:**
- `internal/delivery/feed/handler.go`
- `internal/delivery/feed/handler_test.go`

**Note:** Feed handlers use **plain Chi** (`http.ResponseWriter`, `*http.Request`), NOT Huma, because they return XML/text content rather than JSON. This matches the spec's guidance for public-facing feed endpoints.

**TDD cycle:**

- [ ] **RED** — Write `handler_test.go`:
  ```go
  package feed_test

  import (
      "context"
      "net/http"
      "net/http/httptest"
      "strings"
      "testing"
      "time"

      "github.com/go-chi/chi/v5"
      "github.com/sky-flux/cms/internal/delivery/feed"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  // --- stubs ---

  type stubFeedPostQuery struct {
      posts []feed.FeedPost
      latestAt *time.Time
  }

  func (s *stubFeedPostQuery) ListPublished(_ context.Context, _ int) ([]feed.FeedPost, error) {
      return s.posts, nil
  }
  func (s *stubFeedPostQuery) LatestPublishedAt(_ context.Context) (*time.Time, error) {
      return s.latestAt, nil
  }

  type stubFeedCategoryQuery struct{}
  func (s *stubFeedCategoryQuery) ListAll(_ context.Context) ([]feed.FeedCategoryItem, error) {
      return []feed.FeedCategoryItem{{Slug: "news"}}, nil
  }

  type stubFeedTagQuery struct{}
  func (s *stubFeedTagQuery) ListWithPosts(_ context.Context) ([]feed.FeedTagItem, error) {
      return []feed.FeedTagItem{{Slug: "go"}}, nil
  }

  type stubFeedSiteQuery struct{}
  func (s *stubFeedSiteQuery) GetSiteTitle(_ context.Context) string       { return "Test Site" }
  func (s *stubFeedSiteQuery) GetSiteURL(_ context.Context) string         { return "https://example.com" }
  func (s *stubFeedSiteQuery) GetSiteDescription(_ context.Context) string { return "A test site" }
  func (s *stubFeedSiteQuery) GetSiteLanguage(_ context.Context) string    { return "en" }

  func newTestRouter() *chi.Mux {
      now := time.Now()
      posts := []feed.FeedPost{
          {
              ID:          "1",
              Title:       "Hello World",
              Slug:        "hello-world",
              AuthorName:  "Alice",
              PublishedAt: now,
              UpdatedAt:   now,
          },
      }
      h := feed.NewHandler(
          &stubFeedPostQuery{posts: posts, latestAt: &now},
          &stubFeedCategoryQuery{},
          &stubFeedTagQuery{},
          &stubFeedSiteQuery{},
      )
      r := chi.NewRouter()
      feed.RegisterRoutes(r, h)
      return r
  }

  func TestRSSFeed_Returns200AndXML(t *testing.T) {
      r := newTestRouter()
      req := httptest.NewRequest(http.MethodGet, "/feed/rss", nil)
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
      assert.Contains(t, rec.Header().Get("Content-Type"), "application/rss+xml")
      assert.True(t, strings.Contains(rec.Body.String(), "Hello World"))
      assert.True(t, strings.Contains(rec.Body.String(), "Test Site"))
  }

  func TestAtomFeed_Returns200AndXML(t *testing.T) {
      r := newTestRouter()
      req := httptest.NewRequest(http.MethodGet, "/feed/atom", nil)
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
      assert.Contains(t, rec.Header().Get("Content-Type"), "application/atom+xml")
      assert.True(t, strings.Contains(rec.Body.String(), "Hello World"))
  }

  func TestSitemapIndex_Returns200AndXML(t *testing.T) {
      r := newTestRouter()
      req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
      assert.Contains(t, rec.Header().Get("Content-Type"), "application/xml")
      assert.True(t, strings.Contains(rec.Body.String(), "sitemap"))
  }

  func TestSitemapPosts_Returns200(t *testing.T) {
      r := newTestRouter()
      req := httptest.NewRequest(http.MethodGet, "/sitemap-posts.xml", nil)
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)
      require.Equal(t, http.StatusOK, rec.Code)
  }

  func TestFeed_LimitCap(t *testing.T) {
      // limit query param above 50 should be silently capped
      r := newTestRouter()
      req := httptest.NewRequest(http.MethodGet, "/feed/rss?limit=9999", nil)
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)
      require.Equal(t, http.StatusOK, rec.Code)
  }
  ```

- [ ] **Verify RED** — fails: `feed.NewHandler` undefined, `feed.RegisterRoutes` undefined

- [ ] **GREEN** — Create `internal/delivery/feed/handler.go`:
  ```go
  package feed

  import (
      "crypto/md5"
      "encoding/xml"
      "fmt"
      "net/http"
      "strconv"
      "time"

      "github.com/go-chi/chi/v5"
  )

  const (
      xmlHeaderRSS  = `<?xml version="1.0" encoding="UTF-8"?>`
      sitemapNS     = "http://www.sitemaps.org/schemas/sitemap/0.9"
      atomNS        = "http://www.w3.org/2005/Atom"
      feedGenerator = "Sky Flux CMS"
      maxFeedLimit  = 50
      feedCacheSecs = 3600
  )

  // Handler serves XML feed and sitemap endpoints via plain Chi.
  type Handler struct {
      posts      FeedPostQuery
      categories FeedCategoryQuery
      tags       FeedTagQuery
      site       FeedSiteQuery
  }

  // NewHandler creates a feed handler with the required query ports.
  func NewHandler(
      posts FeedPostQuery,
      categories FeedCategoryQuery,
      tags FeedTagQuery,
      site FeedSiteQuery,
  ) *Handler {
      return &Handler{posts: posts, categories: categories, tags: tags, site: site}
  }

  // RegisterRoutes wires feed endpoints onto a Chi router.
  func RegisterRoutes(r chi.Router, h *Handler) {
      r.Get("/feed/rss", h.RSS)
      r.Get("/feed/atom", h.Atom)
      r.Get("/sitemap.xml", h.SitemapIndex)
      r.Get("/sitemap-posts.xml", h.SitemapPosts)
      r.Get("/sitemap-categories.xml", h.SitemapCategories)
      r.Get("/sitemap-tags.xml", h.SitemapTags)
  }

  // RSS serves GET /feed/rss — RSS 2.0 XML.
  func (h *Handler) RSS(w http.ResponseWriter, r *http.Request) {
      limit := parseLimit(r, maxFeedLimit)
      ctx := r.Context()

      posts, err := h.posts.ListPublished(ctx, limit)
      if err != nil {
          http.Error(w, "feed generation error", http.StatusInternalServerError)
          return
      }
      siteURL := h.site.GetSiteURL(ctx)
      feedURL := siteURL + "/feed/rss"

      ch := RSSChannel{
          Title:         h.site.GetSiteTitle(ctx),
          Link:          siteURL,
          Description:   h.site.GetSiteDescription(ctx),
          Language:      h.site.GetSiteLanguage(ctx),
          LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
          Generator:     feedGenerator,
          AtomLink:      AtomLink{Href: feedURL, Rel: "self", Type: "application/rss+xml"},
      }
      for _, p := range posts {
          ch.Items = append(ch.Items, RSSItem{
              Title:   p.Title,
              Link:    siteURL + "/posts/" + p.Slug,
              GUID:    GUID{IsPermaLink: "true", Value: siteURL + "/posts/" + p.Slug},
              Description: p.Excerpt,
              ContentEncoded: p.Content,
              Creator: p.AuthorName,
              PubDate: p.PublishedAt.UTC().Format(time.RFC1123Z),
              Categories: p.Categories,
          })
      }
      doc := RSSFeed{
          Version: "2.0",
          Atom:    "http://www.w3.org/2005/Atom",
          DC:      "http://purl.org/dc/elements/1.1/",
          Content: "http://purl.org/rss/1.0/modules/content/",
          Channel: ch,
      }
      writeXMLDoc(w, "application/rss+xml; charset=utf-8", doc)
  }

  // Atom serves GET /feed/atom — Atom 1.0 XML.
  func (h *Handler) Atom(w http.ResponseWriter, r *http.Request) {
      limit := parseLimit(r, maxFeedLimit)
      ctx := r.Context()

      posts, err := h.posts.ListPublished(ctx, limit)
      if err != nil {
          http.Error(w, "feed generation error", http.StatusInternalServerError)
          return
      }
      siteURL := h.site.GetSiteURL(ctx)
      feedURL := siteURL + "/feed/atom"

      doc := AtomFeed{
          XMLNS:     atomNS,
          Title:     h.site.GetSiteTitle(ctx),
          ID:        siteURL,
          Updated:   time.Now().UTC().Format(time.RFC3339),
          Generator: feedGenerator,
          Link: []AtomFeedLink{
              {Href: siteURL, Rel: "alternate", Type: "text/html"},
              {Href: feedURL, Rel: "self", Type: "application/atom+xml"},
          },
      }
      for _, p := range posts {
          postURL := siteURL + "/posts/" + p.Slug
          entry := AtomEntry{
              Title:     p.Title,
              ID:        postURL,
              Updated:   p.UpdatedAt.UTC().Format(time.RFC3339),
              Published: p.PublishedAt.UTC().Format(time.RFC3339),
              Link:      AtomFeedLink{Href: postURL, Rel: "alternate", Type: "text/html"},
              Summary:   p.Excerpt,
          }
          if p.Content != "" {
              entry.Content = &AtomContent{Type: "html", Value: p.Content}
          }
          if p.AuthorName != "" {
              entry.Author = &AtomAuthor{Name: p.AuthorName}
          }
          doc.Entries = append(doc.Entries, entry)
      }
      writeXMLDoc(w, "application/atom+xml; charset=utf-8", doc)
  }

  // SitemapIndex serves GET /sitemap.xml — index referencing sub-sitemaps.
  func (h *Handler) SitemapIndex(w http.ResponseWriter, r *http.Request) {
      ctx := r.Context()
      siteURL := h.site.GetSiteURL(ctx)
      now := time.Now().UTC().Format("2006-01-02")

      doc := SitemapIndex{
          XMLNS: sitemapNS,
          Sitemaps: []SitemapEntry{
              {Loc: siteURL + "/sitemap-posts.xml", Lastmod: now},
              {Loc: siteURL + "/sitemap-categories.xml", Lastmod: now},
              {Loc: siteURL + "/sitemap-tags.xml", Lastmod: now},
          },
      }
      writeXMLDoc(w, "application/xml; charset=utf-8", doc)
  }

  // SitemapPosts serves GET /sitemap-posts.xml.
  func (h *Handler) SitemapPosts(w http.ResponseWriter, r *http.Request) {
      ctx := r.Context()
      posts, err := h.posts.ListPublished(ctx, 1000)
      if err != nil {
          http.Error(w, "sitemap generation error", http.StatusInternalServerError)
          return
      }
      siteURL := h.site.GetSiteURL(ctx)
      doc := URLSet{XMLNS: sitemapNS}
      for _, p := range posts {
          doc.URLs = append(doc.URLs, SitemapURL{
              Loc:        siteURL + "/posts/" + p.Slug,
              Lastmod:    p.UpdatedAt.UTC().Format("2006-01-02"),
              Changefreq: "weekly",
              Priority:   "0.8",
          })
      }
      writeXMLDoc(w, "application/xml; charset=utf-8", doc)
  }

  // SitemapCategories serves GET /sitemap-categories.xml.
  func (h *Handler) SitemapCategories(w http.ResponseWriter, r *http.Request) {
      ctx := r.Context()
      cats, err := h.categories.ListAll(ctx)
      if err != nil {
          http.Error(w, "sitemap generation error", http.StatusInternalServerError)
          return
      }
      siteURL := h.site.GetSiteURL(ctx)
      doc := URLSet{XMLNS: sitemapNS}
      for _, c := range cats {
          u := SitemapURL{
              Loc:        siteURL + "/categories/" + c.Slug,
              Changefreq: "weekly",
              Priority:   "0.6",
          }
          if c.LastPostAt != nil {
              u.Lastmod = c.LastPostAt.UTC().Format("2006-01-02")
          }
          doc.URLs = append(doc.URLs, u)
      }
      writeXMLDoc(w, "application/xml; charset=utf-8", doc)
  }

  // SitemapTags serves GET /sitemap-tags.xml.
  func (h *Handler) SitemapTags(w http.ResponseWriter, r *http.Request) {
      ctx := r.Context()
      tags, err := h.tags.ListWithPosts(ctx)
      if err != nil {
          http.Error(w, "sitemap generation error", http.StatusInternalServerError)
          return
      }
      siteURL := h.site.GetSiteURL(ctx)
      doc := URLSet{XMLNS: sitemapNS}
      for _, tag := range tags {
          u := SitemapURL{
              Loc:        siteURL + "/tags/" + tag.Slug,
              Changefreq: "weekly",
              Priority:   "0.5",
          }
          if tag.LastPostAt != nil {
              u.Lastmod = tag.LastPostAt.UTC().Format("2006-01-02")
          }
          doc.URLs = append(doc.URLs, u)
      }
      writeXMLDoc(w, "application/xml; charset=utf-8", doc)
  }

  // writeXMLDoc marshals v to XML and writes it with ETag + Cache-Control headers.
  func writeXMLDoc(w http.ResponseWriter, contentType string, v any) {
      data, err := xml.MarshalIndent(v, "", "  ")
      if err != nil {
          http.Error(w, "xml marshal error", http.StatusInternalServerError)
          return
      }
      out := []byte(xmlHeaderRSS + "\n")
      out = append(out, data...)
      etag := fmt.Sprintf(`"%x"`, md5.Sum(out))
      w.Header().Set("Content-Type", contentType)
      w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", feedCacheSecs))
      w.Header().Set("ETag", etag)
      w.WriteHeader(http.StatusOK)
      w.Write(out)
  }

  // parseLimit reads ?limit= from the request, caps at maxLimit, defaults to 20.
  func parseLimit(r *http.Request, maxLimit int) int {
      limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
      if limit <= 0 {
          limit = 20
      }
      if limit > maxLimit {
          limit = maxLimit
      }
      return limit
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/delivery/feed/... -v -count=1` — all pass

- [ ] **REFACTOR** — Extract `buildRSSItem` and `buildAtomEntry` helpers to reduce handler method length

- [ ] **Commit:** `git commit -m "✨ feat(delivery): add feed Chi handler (RSS/Atom/Sitemap)"`

---

## TASK 4 — Public API query interfaces

**Files:**
- `internal/delivery/public/queries.go`
- `internal/delivery/public/queries_test.go`

**TDD cycle:**

- [ ] **RED** — Write `queries_test.go` verifying interface satisfaction at compile time:
  ```go
  package public_test

  import (
      "context"
      "testing"

      "github.com/sky-flux/cms/internal/delivery/public"
  )

  var _ public.PublicPostQuery = (*mockPublicPostQuery)(nil)
  var _ public.PublicCategoryQuery = (*mockPublicCategoryQuery)(nil)
  var _ public.PublicTagQuery = (*mockPublicTagQuery)(nil)
  var _ public.PublicSearchQuery = (*mockPublicSearchQuery)(nil)

  type mockPublicPostQuery struct{}

  func (m *mockPublicPostQuery) ListPublished(ctx context.Context, f public.PostFilter) ([]public.PublicPost, int64, error) {
      return nil, 0, nil
  }
  func (m *mockPublicPostQuery) GetBySlug(ctx context.Context, slug string) (*public.PublicPost, error) {
      return nil, nil
  }
  func (m *mockPublicPostQuery) IncrementViewCount(ctx context.Context, id string) error { return nil }

  type mockPublicCategoryQuery struct{}

  func (m *mockPublicCategoryQuery) ListWithPostCounts(ctx context.Context) ([]public.PublicCategory, error) {
      return nil, nil
  }

  type mockPublicTagQuery struct{}

  func (m *mockPublicTagQuery) ListWithPostCounts(ctx context.Context, sort string) ([]public.PublicTag, error) {
      return nil, nil
  }

  type mockPublicSearchQuery struct{}

  func (m *mockPublicSearchQuery) Search(ctx context.Context, q string, page, perPage int) ([]public.SearchResult, int64, error) {
      return nil, 0, nil
  }

  func TestPublicQueryInterfaces_Compile(t *testing.T) {
      t.Log("all public query interfaces satisfied")
  }
  ```

- [ ] **Verify RED** — fails: package `public` not found

- [ ] **GREEN** — Create `internal/delivery/public/queries.go`:
  ```go
  package public

  import (
      "context"
      "time"
  )

  // PostFilter controls pagination and filtering for the public post list.
  type PostFilter struct {
      Page     int
      PerPage  int
      Category string // category slug
      Tag      string // tag slug
      Sort     string // e.g. "published_at:desc"
  }

  // PublicPost is the read model projected for the public API.
  type PublicPost struct {
      ID          string    `json:"id"`
      Title       string    `json:"title"`
      Slug        string    `json:"slug"`
      Excerpt     string    `json:"excerpt"`
      Content     string    `json:"content,omitempty"`
      AuthorName  string    `json:"author_name"`
      CoverURL    string    `json:"cover_url,omitempty"`
      PublishedAt time.Time `json:"published_at"`
      UpdatedAt   time.Time `json:"updated_at"`
      ViewCount   int64     `json:"view_count"`
      Tags        []string  `json:"tags,omitempty"`
      Categories  []string  `json:"categories,omitempty"`
  }

  // PublicCategory is a flat category node with post count.
  type PublicCategory struct {
      ID        string `json:"id"`
      Name      string `json:"name"`
      Slug      string `json:"slug"`
      ParentID  string `json:"parent_id,omitempty"`
      PostCount int64  `json:"post_count"`
  }

  // PublicTag is a tag with post count.
  type PublicTag struct {
      ID        string `json:"id"`
      Name      string `json:"name"`
      Slug      string `json:"slug"`
      PostCount int64  `json:"post_count"`
  }

  // SearchResult is a single item returned by Meilisearch.
  type SearchResult struct {
      ID      string `json:"id"`
      Title   string `json:"title"`
      Slug    string `json:"slug"`
      Excerpt string `json:"excerpt"`
      Type    string `json:"type"` // "post"
  }

  // PublicPostQuery is the read port for published posts.
  type PublicPostQuery interface {
      // ListPublished returns paginated published posts matching the filter.
      ListPublished(ctx context.Context, f PostFilter) ([]PublicPost, int64, error)
      // GetBySlug returns a single published post by URL slug, or nil + ErrNotFound.
      GetBySlug(ctx context.Context, slug string) (*PublicPost, error)
      // IncrementViewCount records a page view (fire-and-forget; errors are ignored).
      IncrementViewCount(ctx context.Context, id string) error
  }

  // PublicCategoryQuery is the read port for public category listing.
  type PublicCategoryQuery interface {
      // ListWithPostCounts returns all categories with their published post counts.
      ListWithPostCounts(ctx context.Context) ([]PublicCategory, error)
  }

  // PublicTagQuery is the read port for public tag listing.
  type PublicTagQuery interface {
      // ListWithPostCounts returns all tags with their published post counts.
      // sort is a field:direction string, e.g. "name:asc" or "post_count:desc".
      ListWithPostCounts(ctx context.Context, sort string) ([]PublicTag, error)
  }

  // PublicSearchQuery is the read port for full-text search via Meilisearch.
  type PublicSearchQuery interface {
      // Search queries the Meilisearch index and returns matching posts.
      Search(ctx context.Context, q string, page, perPage int) ([]SearchResult, int64, error)
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/delivery/public/... -v -count=1` — all pass

- [ ] **Commit:** `git commit -m "✨ feat(delivery): add public API query interfaces"`

---

## TASK 5 — Public API Huma handler

**Files:**
- `internal/delivery/public/handler.go`
- `internal/delivery/public/handler_test.go`

**TDD cycle:**

- [ ] **RED** — Write `handler_test.go`:
  ```go
  package public_test

  import (
      "context"
      "encoding/json"
      "net/http"
      "net/http/httptest"
      "testing"
      "time"

      "github.com/danielgtaylor/huma/v2"
      "github.com/danielgtaylor/huma/v2/humatest"
      "github.com/sky-flux/cms/internal/delivery/public"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  // --- stubs ---

  type stubPostQuery struct {
      posts []public.PublicPost
      total int64
      post  *public.PublicPost
  }

  func (s *stubPostQuery) ListPublished(_ context.Context, _ public.PostFilter) ([]public.PublicPost, int64, error) {
      return s.posts, s.total, nil
  }
  func (s *stubPostQuery) GetBySlug(_ context.Context, _ string) (*public.PublicPost, error) {
      return s.post, nil
  }
  func (s *stubPostQuery) IncrementViewCount(_ context.Context, _ string) error { return nil }

  type stubCategoryQuery struct{}

  func (s *stubCategoryQuery) ListWithPostCounts(_ context.Context) ([]public.PublicCategory, error) {
      return []public.PublicCategory{
          {ID: "cat-1", Name: "News", Slug: "news", PostCount: 3},
      }, nil
  }

  type stubTagQuery struct{}

  func (s *stubTagQuery) ListWithPostCounts(_ context.Context, _ string) ([]public.PublicTag, error) {
      return []public.PublicTag{
          {ID: "tag-1", Name: "Go", Slug: "go", PostCount: 5},
      }, nil
  }

  type stubSearchQuery struct{}

  func (s *stubSearchQuery) Search(_ context.Context, q string, _, _ int) ([]public.SearchResult, int64, error) {
      if q == "" {
          return nil, 0, nil
      }
      return []public.SearchResult{{ID: "1", Title: "Hello", Slug: "hello", Type: "post"}}, 1, nil
  }

  func newTestAPI(t *testing.T) (huma.API, *public.Handler) {
      t.Helper()
      _, api := humatest.New(t, huma.DefaultConfig("Test API", "1.0.0"))
      now := time.Now()
      posts := []public.PublicPost{
          {ID: "1", Title: "Hello World", Slug: "hello-world", PublishedAt: now},
      }
      h := public.NewHandler(
          &stubPostQuery{posts: posts, total: 1, post: &posts[0]},
          &stubCategoryQuery{},
          &stubTagQuery{},
          &stubSearchQuery{},
      )
      public.RegisterRoutes(api, h)
      return api, h
  }

  func TestListPosts_Returns200(t *testing.T) {
      api, _ := newTestAPI(t)
      req := httptest.NewRequest(http.MethodGet, "/api/v1/public/posts", nil)
      rec := httptest.NewRecorder()
      api.Adapter().ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
      var body map[string]any
      require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
      items, ok := body["items"].([]any)
      require.True(t, ok)
      assert.Len(t, items, 1)
  }

  func TestGetPost_Returns200(t *testing.T) {
      api, _ := newTestAPI(t)
      req := httptest.NewRequest(http.MethodGet, "/api/v1/public/posts/hello-world", nil)
      rec := httptest.NewRecorder()
      api.Adapter().ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
  }

  func TestListCategories_Returns200(t *testing.T) {
      api, _ := newTestAPI(t)
      req := httptest.NewRequest(http.MethodGet, "/api/v1/public/categories", nil)
      rec := httptest.NewRecorder()
      api.Adapter().ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
      var body map[string]any
      require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
      items := body["items"].([]any)
      assert.Len(t, items, 1)
  }

  func TestListTags_Returns200(t *testing.T) {
      api, _ := newTestAPI(t)
      req := httptest.NewRequest(http.MethodGet, "/api/v1/public/tags", nil)
      rec := httptest.NewRecorder()
      api.Adapter().ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
  }

  func TestSearch_WithQuery_Returns200(t *testing.T) {
      api, _ := newTestAPI(t)
      req := httptest.NewRequest(http.MethodGet, "/api/v1/public/search?q=hello", nil)
      rec := httptest.NewRecorder()
      api.Adapter().ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
      var body map[string]any
      require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
      items := body["items"].([]any)
      assert.Len(t, items, 1)
  }

  func TestSearch_EmptyQuery_ReturnsEmpty(t *testing.T) {
      api, _ := newTestAPI(t)
      req := httptest.NewRequest(http.MethodGet, "/api/v1/public/search?q=", nil)
      rec := httptest.NewRecorder()
      api.Adapter().ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
  }
  ```

- [ ] **Verify RED** — fails: `public.NewHandler` undefined, `public.RegisterRoutes` undefined

- [ ] **GREEN** — Create `internal/delivery/public/handler.go`:
  ```go
  package public

  import (
      "context"
      "net/http"

      "github.com/danielgtaylor/huma/v2"
  )

  // Handler holds all public API delivery dependencies.
  type Handler struct {
      posts      PublicPostQuery
      categories PublicCategoryQuery
      tags       PublicTagQuery
      search     PublicSearchQuery
  }

  // NewHandler creates a public API handler.
  func NewHandler(
      posts PublicPostQuery,
      categories PublicCategoryQuery,
      tags PublicTagQuery,
      search PublicSearchQuery,
  ) *Handler {
      return &Handler{posts: posts, categories: categories, tags: tags, search: search}
  }

  // RegisterRoutes wires all public endpoints onto a Huma API.
  func RegisterRoutes(api huma.API, h *Handler) {
      huma.Register(api, huma.Operation{
          OperationID: "public-list-posts",
          Method:      http.MethodGet,
          Path:        "/api/v1/public/posts",
          Summary:     "List published posts",
          Tags:        []string{"Public"},
      }, h.ListPosts)

      huma.Register(api, huma.Operation{
          OperationID: "public-get-post",
          Method:      http.MethodGet,
          Path:        "/api/v1/public/posts/{slug}",
          Summary:     "Get a published post by slug",
          Tags:        []string{"Public"},
      }, h.GetPost)

      huma.Register(api, huma.Operation{
          OperationID: "public-list-categories",
          Method:      http.MethodGet,
          Path:        "/api/v1/public/categories",
          Summary:     "List categories with post counts",
          Tags:        []string{"Public"},
      }, h.ListCategories)

      huma.Register(api, huma.Operation{
          OperationID: "public-list-tags",
          Method:      http.MethodGet,
          Path:        "/api/v1/public/tags",
          Summary:     "List tags with post counts",
          Tags:        []string{"Public"},
      }, h.ListTags)

      huma.Register(api, huma.Operation{
          OperationID: "public-search",
          Method:      http.MethodGet,
          Path:        "/api/v1/public/search",
          Summary:     "Full-text search via Meilisearch",
          Tags:        []string{"Public"},
      }, h.Search)
  }

  // --- Request / Response types ---

  type ListPostsInput struct {
      Page     int    `query:"page" default:"1" minimum:"1"`
      PerPage  int    `query:"per_page" default:"20" minimum:"1" maximum:"100"`
      Category string `query:"category"`
      Tag      string `query:"tag"`
      Sort     string `query:"sort" default:"published_at:desc"`
  }

  type PostListOutput struct {
      Body struct {
          Items []PublicPost `json:"items"`
          Total int64        `json:"total"`
          Page  int          `json:"page"`
      }
  }

  type GetPostInput struct {
      Slug string `path:"slug"`
  }

  type PostOutput struct {
      Body PublicPost
  }

  type ListCategoriesOutput struct {
      Body struct {
          Items []PublicCategory `json:"items"`
      }
  }

  type ListTagsInput struct {
      Sort string `query:"sort" default:"name:asc"`
  }

  type ListTagsOutput struct {
      Body struct {
          Items []PublicTag `json:"items"`
      }
  }

  type SearchInput struct {
      Q       string `query:"q"`
      Page    int    `query:"page" default:"1" minimum:"1"`
      PerPage int    `query:"per_page" default:"20" minimum:"1" maximum:"100"`
  }

  type SearchOutput struct {
      Body struct {
          Items []SearchResult `json:"items"`
          Total int64          `json:"total"`
      }
  }

  // --- Handlers ---

  // ListPosts handles GET /api/v1/public/posts.
  func (h *Handler) ListPosts(ctx context.Context, in *ListPostsInput) (*PostListOutput, error) {
      items, total, err := h.posts.ListPublished(ctx, PostFilter{
          Page:     in.Page,
          PerPage:  in.PerPage,
          Category: in.Category,
          Tag:      in.Tag,
          Sort:     in.Sort,
      })
      if err != nil {
          return nil, huma.NewError(http.StatusInternalServerError, "failed to list posts")
      }
      out := &PostListOutput{}
      out.Body.Items = items
      out.Body.Total = total
      out.Body.Page = in.Page
      return out, nil
  }

  // GetPost handles GET /api/v1/public/posts/{slug}.
  func (h *Handler) GetPost(ctx context.Context, in *GetPostInput) (*PostOutput, error) {
      post, err := h.posts.GetBySlug(ctx, in.Slug)
      if err != nil {
          return nil, huma.NewError(http.StatusNotFound, "post not found")
      }
      // Fire-and-forget view count; ignore errors.
      go h.posts.IncrementViewCount(context.Background(), post.ID)
      return &PostOutput{Body: *post}, nil
  }

  // ListCategories handles GET /api/v1/public/categories.
  func (h *Handler) ListCategories(ctx context.Context, _ *struct{}) (*ListCategoriesOutput, error) {
      cats, err := h.categories.ListWithPostCounts(ctx)
      if err != nil {
          return nil, huma.NewError(http.StatusInternalServerError, "failed to list categories")
      }
      out := &ListCategoriesOutput{}
      out.Body.Items = cats
      return out, nil
  }

  // ListTags handles GET /api/v1/public/tags.
  func (h *Handler) ListTags(ctx context.Context, in *ListTagsInput) (*ListTagsOutput, error) {
      tags, err := h.tags.ListWithPostCounts(ctx, in.Sort)
      if err != nil {
          return nil, huma.NewError(http.StatusInternalServerError, "failed to list tags")
      }
      out := &ListTagsOutput{}
      out.Body.Items = tags
      return out, nil
  }

  // Search handles GET /api/v1/public/search.
  func (h *Handler) Search(ctx context.Context, in *SearchInput) (*SearchOutput, error) {
      if in.Q == "" {
          return &SearchOutput{}, nil
      }
      results, total, err := h.search.Search(ctx, in.Q, in.Page, in.PerPage)
      if err != nil {
          return nil, huma.NewError(http.StatusInternalServerError, "search failed")
      }
      out := &SearchOutput{}
      out.Body.Items = results
      out.Body.Total = total
      return out, nil
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/delivery/public/... -v -count=1` — all pass

- [ ] **REFACTOR** — Extract `mapNotFound` error helper; add `ErrNotFound` sentinel to queries package

- [ ] **Commit:** `git commit -m "✨ feat(delivery): add public API Huma handler"`

---

## TASK 6 — API Key auth middleware

**Files:**
- `internal/delivery/public/middleware.go`
- `internal/delivery/public/middleware_test.go`

**TDD cycle:**

- [ ] **RED** — Write `middleware_test.go`:
  ```go
  package public_test

  import (
      "context"
      "net/http"
      "net/http/httptest"
      "testing"

      "github.com/go-chi/chi/v5"
      "github.com/sky-flux/cms/internal/delivery/public"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  type stubAPIKeyValidator struct {
      valid bool
  }

  func (s *stubAPIKeyValidator) ValidateAPIKey(ctx context.Context, rawKey string) (bool, error) {
      return s.valid, nil
  }

  func okHandler(w http.ResponseWriter, r *http.Request) {
      w.WriteHeader(http.StatusOK)
      w.Write([]byte("ok"))
  }

  func TestOptionalAPIKey_NoKey_Passthrough(t *testing.T) {
      // Without a key header the request must still reach the handler (optional auth).
      mw := public.OptionalAPIKey(&stubAPIKeyValidator{valid: false})
      r := chi.NewRouter()
      r.With(mw).Get("/test", okHandler)

      req := httptest.NewRequest(http.MethodGet, "/test", nil)
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
  }

  func TestOptionalAPIKey_InvalidKey_Returns401(t *testing.T) {
      // An explicit but invalid key must be rejected.
      mw := public.OptionalAPIKey(&stubAPIKeyValidator{valid: false})
      r := chi.NewRouter()
      r.With(mw).Get("/test", okHandler)

      req := httptest.NewRequest(http.MethodGet, "/test", nil)
      req.Header.Set("X-API-Key", "bad-key")
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)

      assert.Equal(t, http.StatusUnauthorized, rec.Code)
  }

  func TestOptionalAPIKey_ValidKey_Passthrough(t *testing.T) {
      mw := public.OptionalAPIKey(&stubAPIKeyValidator{valid: true})
      r := chi.NewRouter()
      r.With(mw).Get("/test", okHandler)

      req := httptest.NewRequest(http.MethodGet, "/test", nil)
      req.Header.Set("X-API-Key", "valid-key-123")
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
  }
  ```

- [ ] **Verify RED** — fails: `public.OptionalAPIKey` undefined

- [ ] **GREEN** — Create `internal/delivery/public/middleware.go`:
  ```go
  package public

  import (
      "context"
      "encoding/json"
      "net/http"
  )

  // APIKeyValidator is the port used by OptionalAPIKey middleware.
  // Implemented by the site BC's infra layer (lookup by SHA-256 hash in sfc_api_keys).
  type APIKeyValidator interface {
      // ValidateAPIKey checks whether rawKey is a valid, active API key.
      // Returns (false, nil) for an invalid key; (false, err) for internal errors.
      ValidateAPIKey(ctx context.Context, rawKey string) (bool, error)
  }

  // OptionalAPIKey returns a Chi middleware that:
  //   - Allows requests without an X-API-Key header (public access).
  //   - Rejects requests that supply an X-API-Key header but fail validation (401).
  //   - Passes requests with a valid key, setting "api_key_valid" in context.
  func OptionalAPIKey(v APIKeyValidator) func(http.Handler) http.Handler {
      return func(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              key := r.Header.Get("X-API-Key")
              if key == "" {
                  // No key supplied — allow anonymous access.
                  next.ServeHTTP(w, r)
                  return
              }
              ok, err := v.ValidateAPIKey(r.Context(), key)
              if err != nil {
                  // Internal error — fail-open (log and continue).
                  next.ServeHTTP(w, r)
                  return
              }
              if !ok {
                  w.Header().Set("Content-Type", "application/json")
                  w.WriteHeader(http.StatusUnauthorized)
                  json.NewEncoder(w).Encode(map[string]string{
                      "title":  "Unauthorized",
                      "detail": "invalid or inactive API key",
                  })
                  return
              }
              ctx := context.WithValue(r.Context(), ctxKeyAPIKeyValid{}, true)
              next.ServeHTTP(w, r.WithContext(ctx))
          })
      }
  }

  type ctxKeyAPIKeyValid struct{}

  // IsAPIKeyAuthenticated reports whether the current request has a validated API key.
  func IsAPIKeyAuthenticated(ctx context.Context) bool {
      v, _ := ctx.Value(ctxKeyAPIKeyValid{}).(bool)
      return v
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/delivery/public/... -v -count=1` — all pass

- [ ] **REFACTOR** — Extract `writeJSON` helper to avoid repeating `json.NewEncoder(w).Encode`

- [ ] **Commit:** `git commit -m "✨ feat(delivery): add OptionalAPIKey middleware"`

---

## Final verification

- [ ] `go test ./internal/delivery/... -v -count=1` — all tasks green
- [ ] `go vet ./internal/delivery/...` — zero warnings
- [ ] `go build ./...` — clean compile

- [ ] **Final commit:** `git commit -m "✅ test(delivery): all delivery BC tests green"`

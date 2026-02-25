# Batch 8: Public Headless API + Feed/Sitemap — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement 15 remaining backend endpoints (9 public API + 6 feed/sitemap) with 2 new middleware, completing backend at 138 endpoints.

**Architecture:** New `internal/public/` module reuses existing repos via interfaces. Feed module (`internal/feed/`) uses `encoding/xml` structs. Two new middleware handle API Key auth and IP rate limiting. All code has strict unit tests.

**Tech Stack:** Go / Gin / uptrace/bun / Redis / Meilisearch / encoding/xml / testify

---

## Task 1: Add `GetByHash` to API Key Repo

**Files:**
- Modify: `internal/apikey/repository.go`
- Test: (covered by middleware test in Task 3)

**Step 1: Add GetByHash method to existing repo**

Add to `internal/apikey/repository.go` after the `Revoke` method:

```go
func (r *Repo) GetByHash(ctx context.Context, hash string) (*model.APIKey, error) {
	key := new(model.APIKey)
	err := r.db.NewSelect().
		Model(key).
		Where("key_hash = ?", hash).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("api key not found", err)
		}
		return nil, fmt.Errorf("apikey get by hash: %w", err)
	}
	return key, nil
}
```

Note: `database/sql` and `errors` are already imported. `apperror` is already imported.

**Step 2: Verify compilation**

Run: `go build ./internal/apikey/...`
Expected: BUILD SUCCESS

**Step 3: Commit**

```bash
git add internal/apikey/repository.go
git commit -m "feat(apikey): add GetByHash method for API key middleware"
```

---

## Task 2: API Key Middleware

**Files:**
- Create: `internal/middleware/api_key.go`
- Create: `internal/middleware/api_key_test.go`

**Step 1: Create API Key middleware**

Create `internal/middleware/api_key.go`:

```go
package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
)

// APIKeyLookup abstracts API key validation.
type APIKeyLookup interface {
	GetByHash(ctx context.Context, hash string) (*model.APIKey, error)
}

// APIKeyUpdater abstracts async last_used_at update. Optional.
type APIKeyUpdater interface {
	UpdateLastUsed(ctx context.Context, id string, t time.Time) error
}

// APIKey validates the X-API-Key header against sfc_site_api_keys.
func APIKey(lookup APIKeyLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("X-API-Key")
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing X-API-Key header",
			})
			return
		}

		h := sha256.Sum256([]byte(raw))
		hash := hex.EncodeToString(h[:])

		key, err := lookup.GetByHash(c.Request.Context(), hash)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid api key",
			})
			return
		}

		if key.Status != model.APIKeyStatusActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "api key revoked",
			})
			return
		}

		if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "api key expired",
			})
			return
		}

		c.Set("api_key_id", key.ID)
		c.Next()
	}
}
```

**Step 2: Write tests**

Create `internal/middleware/api_key_test.go`:

```go
package middleware_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/middleware"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/stretchr/testify/assert"
)

type mockAPIKeyLookup struct {
	key *model.APIKey
	err error
}

func (m *mockAPIKeyLookup) GetByHash(_ context.Context, _ string) (*model.APIKey, error) {
	return m.key, m.err
}

func hashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func setupAPIKeyRouter(lookup middleware.APIKeyLookup) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.APIKey(lookup))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"api_key_id": c.GetString("api_key_id")})
	})
	return r
}

func TestAPIKey_ValidKey_PassesThrough(t *testing.T) {
	raw := "test-api-key-123"
	lookup := &mockAPIKeyLookup{
		key: &model.APIKey{ID: "key-1", KeyHash: hashKey(raw), Status: model.APIKeyStatusActive},
	}
	r := setupAPIKeyRouter(lookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", raw)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "key-1")
}

func TestAPIKey_MissingHeader_Returns401(t *testing.T) {
	lookup := &mockAPIKeyLookup{}
	r := setupAPIKeyRouter(lookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "missing")
}

func TestAPIKey_InvalidKey_Returns401(t *testing.T) {
	lookup := &mockAPIKeyLookup{err: apperror.NotFound("not found", nil)}
	r := setupAPIKeyRouter(lookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "invalid")
}

func TestAPIKey_RevokedKey_Returns401(t *testing.T) {
	raw := "revoked-key"
	lookup := &mockAPIKeyLookup{
		key: &model.APIKey{ID: "key-2", KeyHash: hashKey(raw), Status: model.APIKeyStatusRevoked},
	}
	r := setupAPIKeyRouter(lookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", raw)
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "revoked")
}

func TestAPIKey_ExpiredKey_Returns401(t *testing.T) {
	raw := "expired-key"
	expired := time.Now().Add(-time.Hour)
	lookup := &mockAPIKeyLookup{
		key: &model.APIKey{ID: "key-3", KeyHash: hashKey(raw), Status: model.APIKeyStatusActive, ExpiresAt: &expired},
	}
	r := setupAPIKeyRouter(lookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", raw)
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "expired")
}
```

**Step 3: Run tests**

Run: `go test ./internal/middleware/ -run TestAPIKey -v`
Expected: 5 PASS

**Step 4: Commit**

```bash
git add internal/middleware/api_key.go internal/middleware/api_key_test.go
git commit -m "feat(middleware): add API Key authentication middleware"
```

---

## Task 3: Rate Limit Middleware

**Files:**
- Create: `internal/middleware/rate_limit.go`
- Create: `internal/middleware/rate_limit_test.go`

**Step 1: Create Rate Limit middleware**

Create `internal/middleware/rate_limit.go`:

```go
package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimit returns middleware that limits requests per IP using Redis SET NX.
// window defines the cooldown period between requests.
func RateLimit(rdb *redis.Client, prefix string, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			c.Next()
			return
		}

		siteSlug, _ := c.Get("site_slug")
		ip := c.ClientIP()
		key := prefix + ":" + siteSlug.(string) + ":" + ip

		ok, err := rdb.SetNX(c.Request.Context(), key, "1", window).Result()
		if err != nil {
			// Redis error — allow request through (fail open)
			c.Next()
			return
		}
		if !ok {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate limit exceeded, please try again later",
			})
			return
		}
		c.Next()
	}
}
```

**Step 2: Write tests**

Create `internal/middleware/rate_limit_test.go`:

```go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRateLimitRouter(rdb *redis.Client) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("site_slug", "test-site")
		c.Next()
	})
	r.Use(middleware.RateLimit(rdb, "ratelimit:comment", 30*time.Second))
	r.POST("/comment", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	return r
}

func TestRateLimit_FirstRequest_Passes(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	r := setupRateLimitRouter(rdb)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestRateLimit_SecondRequest_Blocked(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	r := setupRateLimitRouter(rdb)

	// First request passes
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request blocked
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)
	assert.Contains(t, w2.Body.String(), "rate limit")
}

func TestRateLimit_AfterExpiry_Passes(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	r := setupRateLimitRouter(rdb)

	// First request
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Fast-forward time in miniredis
	mr.FastForward(31 * time.Second)

	// Should pass again
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
}

func TestRateLimit_NilRedis_FailsOpen(t *testing.T) {
	r := setupRateLimitRouter(nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
```

**Step 3: Run tests**

Run: `go test ./internal/middleware/ -run TestRateLimit -v`
Expected: 4 PASS

**Step 4: Commit**

```bash
git add internal/middleware/rate_limit.go internal/middleware/rate_limit_test.go
git commit -m "feat(middleware): add Redis-based IP rate limiting middleware"
```

---

## Task 4: Public Module — Interfaces + DTOs

**Files:**
- Create: `internal/public/interfaces.go`
- Create: `internal/public/dto.go`

**Step 1: Create interfaces**

Create `internal/public/interfaces.go`:

```go
package public

import (
	"context"

	"github.com/sky-flux/cms/internal/comment"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/search"
)

// PostReader reads published posts.
type PostReader interface {
	List(ctx context.Context, f PostListFilter) ([]model.Post, int64, error)
	GetBySlug(ctx context.Context, slug string) (*model.Post, error)
	LoadRelations(ctx context.Context, post *model.Post) error
	IncrementViewCount(ctx context.Context, id string) error
}

// CategoryReader reads categories with post counts.
type CategoryReader interface {
	List(ctx context.Context) ([]model.Category, error)
	CountPosts(ctx context.Context, categoryID string) (int64, error)
}

// TagReader reads tags with post counts.
type TagReader interface {
	ListPublic(ctx context.Context, sort string) ([]TagWithCount, error)
}

// CommentReader reads and creates comments.
type CommentReader interface {
	ListByPost(ctx context.Context, postID string, f comment.ListFilter) ([]model.Comment, int64, error)
	GetChildren(ctx context.Context, parentID string) ([]*model.Comment, error)
	Create(ctx context.Context, c *model.Comment) error
	GetByID(ctx context.Context, id string) (*model.Comment, error)
	GetParentChainDepth(ctx context.Context, commentID string) (int, error)
}

// MenuReader reads site menus with items.
type MenuReader interface {
	GetByLocation(ctx context.Context, location string) (*model.SiteMenu, error)
	GetBySlug(ctx context.Context, slug string) (*model.SiteMenu, error)
	ListItemsByMenuID(ctx context.Context, menuID string) ([]*model.SiteMenuItem, error)
}

// PreviewReader reads preview tokens.
type PreviewReader interface {
	GetByHash(ctx context.Context, hash string) (*model.PreviewToken, error)
}

// PostListFilter for public post listing — only published.
type PostListFilter struct {
	Page       int
	PerPage    int
	Category   string // category slug
	Tag        string // tag slug
	Locale     string
	Sort       string
}

// TagWithCount represents a tag with its published post count.
type TagWithCount struct {
	model.Tag
	PostCount int64 `json:"post_count"`
}

// Searcher wraps Meilisearch search.
type Searcher interface {
	Search(ctx context.Context, uid, query string, opts *search.SearchOpts) (*search.SearchResult, error)
}

// Cacher wraps Redis cache operations.
type Cacher interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
}
```

Note: You will need to add `"time"` to the import block.

**Step 2: Create DTOs**

Create `internal/public/dto.go`:

```go
package public

import (
	"encoding/json"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// PostListItem is the public response for a post in a list.
type PostListItem struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Excerpt     string           `json:"excerpt,omitempty"`
	Author      *AuthorBrief     `json:"author,omitempty"`
	CoverImage  *CoverImageBrief `json:"cover_image,omitempty"`
	Categories  []RefBrief       `json:"categories,omitempty"`
	Tags        []RefBrief       `json:"tags,omitempty"`
	ViewCount   int64            `json:"view_count"`
	PublishedAt *time.Time       `json:"published_at,omitempty"`
}

// PostDetail is the public response for a single post.
type PostDetail struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Content     string           `json:"content,omitempty"`
	ContentJSON json.RawMessage  `json:"content_json,omitempty"`
	Excerpt     string           `json:"excerpt,omitempty"`
	Author      *AuthorBrief     `json:"author,omitempty"`
	CoverImage  *CoverImageBrief `json:"cover_image,omitempty"`
	Categories  []RefBrief       `json:"categories,omitempty"`
	Tags        []RefBrief       `json:"tags,omitempty"`
	SEO         *SEOFields       `json:"seo,omitempty"`
	ExtraFields json.RawMessage  `json:"extra_fields,omitempty"`
	ViewCount   int64            `json:"view_count"`
	PublishedAt *time.Time       `json:"published_at,omitempty"`
}

// AuthorBrief is the sanitized author info for public API.
type AuthorBrief struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// CoverImageBrief is the sanitized cover image for public API.
type CoverImageBrief struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// RefBrief is a lightweight category/tag reference.
type RefBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// SEOFields groups SEO-related post fields.
type SEOFields struct {
	MetaTitle   string `json:"meta_title,omitempty"`
	MetaDesc    string `json:"meta_description,omitempty"`
	OGImageURL  string `json:"og_image_url,omitempty"`
}

// CategoryNode is a public category with post count and children.
type CategoryNode struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Slug      string         `json:"slug"`
	Path      string         `json:"path"`
	PostCount int64          `json:"post_count"`
	Children  []CategoryNode `json:"children"`
}

// TagItem is a public tag with post count.
type TagItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	PostCount int64  `json:"post_count"`
}

// PublicComment is a sanitized comment for public API (no email/IP/UA).
type PublicComment struct {
	ID         string          `json:"id"`
	ParentID   *string         `json:"parent_id,omitempty"`
	AuthorName string          `json:"author_name"`
	AuthorURL  string          `json:"author_url,omitempty"`
	Content    string          `json:"content"`
	IsPinned   bool            `json:"is_pinned"`
	CreatedAt  time.Time       `json:"created_at"`
	Replies    []PublicComment `json:"replies"`
}

// CreateCommentReq is the request body for public comment submission.
type CreateCommentReq struct {
	ParentID    *string `json:"parent_id"`
	AuthorName  string  `json:"author_name"`
	AuthorEmail string  `json:"author_email"`
	AuthorURL   string  `json:"author_url"`
	Content     string  `json:"content" binding:"required,min=1,max=10000"`
	Honeypot    string  `json:"honeypot"`
}

// CreateCommentResp is the response after submitting a comment.
type CreateCommentResp struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// PublicMenuItem is a sanitized menu item for public API.
type PublicMenuItem struct {
	ID       string           `json:"id"`
	Label    string           `json:"label"`
	URL      string           `json:"url,omitempty"`
	Target   string           `json:"target"`
	Icon     string           `json:"icon,omitempty"`
	CSSClass string           `json:"css_class,omitempty"`
	Children []PublicMenuItem  `json:"children"`
}

// PublicMenu is the public menu response.
type PublicMenu struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Slug     string           `json:"slug"`
	Location string           `json:"location,omitempty"`
	Items    []PublicMenuItem  `json:"items"`
}

// PreviewResp is the response for a preview token consumption.
type PreviewResp struct {
	PostDetail
	IsPreview        bool       `json:"is_preview"`
	PreviewExpiresAt *time.Time `json:"preview_expires_at,omitempty"`
}

// SearchResultItem is a single search result.
type SearchResultItem struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt,omitempty"`
	Author      *AuthorBrief `json:"author,omitempty"`
	Categories  []RefBrief `json:"categories,omitempty"`
	Tags        []RefBrief `json:"tags,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}
```

**Step 3: Verify compilation**

Run: `go build ./internal/public/...`
Expected: BUILD SUCCESS

**Step 4: Commit**

```bash
git add internal/public/interfaces.go internal/public/dto.go
git commit -m "feat(public): add interfaces and DTOs for public headless API"
```

---

## Task 5: Public Module — Service (Posts + Categories + Tags + Search)

**Files:**
- Create: `internal/public/service.go`

**Step 1: Create service with post/category/tag/search methods**

Create `internal/public/service.go`. The service holds all dependencies and implements the 9 endpoint business logic. Since this is a large file, we split it into content reading (this task) and comments/menu/preview (next task).

```go
package public

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/cache"
	"github.com/sky-flux/cms/internal/pkg/search"
)

// Service provides public API business logic.
type Service struct {
	posts      PostReader
	categories CategoryReader
	tags       TagReader
	comments   CommentReader
	menus      MenuReader
	previews   PreviewReader
	search     *search.Client
	cache      *cache.Client
}

// NewService creates a public API service.
func NewService(
	posts PostReader,
	categories CategoryReader,
	tags TagReader,
	comments CommentReader,
	menus MenuReader,
	previews PreviewReader,
	searchClient *search.Client,
	cacheClient *cache.Client,
) *Service {
	return &Service{
		posts:      posts,
		categories: categories,
		tags:       tags,
		comments:   comments,
		menus:      menus,
		previews:   previews,
		search:     searchClient,
		cache:      cacheClient,
	}
}

// ListPosts returns published posts with pagination.
func (s *Service) ListPosts(ctx context.Context, siteSlug string, f PostListFilter) ([]PostListItem, int64, error) {
	posts, total, err := s.posts.List(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("public list posts: %w", err)
	}
	items := make([]PostListItem, len(posts))
	for i, p := range posts {
		items[i] = toPostListItem(&p)
	}
	return items, total, nil
}

// GetPost returns a single published post by slug.
func (s *Service) GetPost(ctx context.Context, siteSlug, slug string) (*PostDetail, error) {
	post, err := s.posts.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if post.Status != model.PostStatusPublished {
		return nil, apperror.NotFound("post not found", nil)
	}
	if err := s.posts.LoadRelations(ctx, post); err != nil {
		slog.Warn("failed to load post relations", "error", err, "post_id", post.ID)
	}
	// Async view count increment
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.posts.IncrementViewCount(bgCtx, post.ID)
	}()
	detail := toPostDetail(post)
	return &detail, nil
}

// ListCategories returns the category tree with post counts.
func (s *Service) ListCategories(ctx context.Context, siteSlug string) ([]CategoryNode, error) {
	cats, err := s.categories.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("public list categories: %w", err)
	}
	return s.buildCategoryTree(ctx, cats), nil
}

// ListTags returns tags with post counts.
func (s *Service) ListTags(ctx context.Context, siteSlug string, sort string) ([]TagItem, error) {
	tags, err := s.tags.ListPublic(ctx, sort)
	if err != nil {
		return nil, fmt.Errorf("public list tags: %w", err)
	}
	items := make([]TagItem, len(tags))
	for i, t := range tags {
		items[i] = TagItem{ID: t.ID, Name: t.Name, Slug: t.Slug, PostCount: t.PostCount}
	}
	return items, nil
}

// Search performs full-text search via Meilisearch.
func (s *Service) Search(ctx context.Context, siteSlug, query string, page, perPage int) ([]SearchResultItem, int64, error) {
	if query == "" {
		return []SearchResultItem{}, 0, nil
	}
	indexUID := "posts-" + siteSlug
	opts := &search.SearchOpts{
		Limit:  int64(perPage),
		Offset: int64((page - 1) * perPage),
		Filter: "status = 3", // published
	}
	result, err := s.search.Search(ctx, indexUID, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("public search: %w", err)
	}
	items := make([]SearchResultItem, len(result.Hits))
	for i, hit := range result.Hits {
		items[i] = hitToSearchResult(hit)
	}
	return items, result.EstimatedTotal, nil
}

// ListComments returns approved comments for a post in nested tree form.
func (s *Service) ListComments(ctx context.Context, postSlug string, page, perPage int) (*CommentListResult, error) {
	// First get the post by slug to get its ID
	post, err := s.posts.GetBySlug(ctx, postSlug)
	if err != nil {
		return nil, err
	}

	from comment import filter
	filter := CommentListFilter{PostID: post.ID, Page: page, PerPage: perPage}
	comments, total, err := s.comments.ListByPost(ctx, post.ID, filter)
	if err != nil {
		return nil, fmt.Errorf("public list comments: %w", err)
	}

	tree := buildCommentTree(comments)
	return &CommentListResult{
		CommentCount: total,
		Comments:     tree,
		Total:        total,
		Page:         page,
		PerPage:      perPage,
	}, nil
}

// CreateComment submits a new comment (guest or authenticated).
func (s *Service) CreateComment(ctx context.Context, postSlug string, req *CreateCommentReq, userID, userName, userEmail, clientIP, userAgent string) (*CreateCommentResp, error) {
	post, err := s.posts.GetBySlug(ctx, postSlug)
	if err != nil {
		return nil, err
	}
	if post.Status != model.PostStatusPublished {
		return nil, apperror.NotFound("post not found", nil)
	}

	// Honeypot check: if filled, still return 201 but mark as spam
	isSpam := req.Honeypot != ""

	// Validate parent_id if provided
	if req.ParentID != nil && *req.ParentID != "" {
		parent, err := s.comments.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, apperror.Validation("parent comment not found", nil)
		}
		if parent.PostID != post.ID {
			return nil, apperror.Validation("parent comment belongs to different post", nil)
		}
		if parent.Status != model.CommentStatusApproved {
			return nil, apperror.Validation("parent comment is not approved", nil)
		}
		// Check nesting depth (max 3 levels)
		depth, err := s.comments.GetParentChainDepth(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("check nesting depth: %w", err)
		}
		if depth >= 2 {
			return nil, apperror.Validation("maximum comment nesting depth exceeded", nil)
		}
	}

	// Determine author info
	authorName := req.AuthorName
	authorEmail := req.AuthorEmail
	var uid *string
	if userID != "" {
		uid = &userID
		if userName != "" {
			authorName = userName
		}
		if userEmail != "" {
			authorEmail = userEmail
		}
	} else {
		// Guest: require name and email
		if authorName == "" {
			return nil, apperror.Validation("author_name is required for guests", nil)
		}
		if authorEmail == "" {
			return nil, apperror.Validation("author_email is required for guests", nil)
		}
	}

	status := model.CommentStatusPending
	if isSpam {
		status = model.CommentStatusSpam
	}

	c := &model.Comment{
		PostID:      post.ID,
		ParentID:    req.ParentID,
		UserID:      uid,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
		AuthorURL:   req.AuthorURL,
		AuthorIP:    clientIP,
		UserAgent:   userAgent,
		Content:     req.Content,
		Status:      status,
	}

	if err := s.comments.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	return &CreateCommentResp{
		ID:      c.ID,
		Status:  "pending",
		Message: "Comment submitted, awaiting moderation",
	}, nil
}

// GetMenu returns a menu tree by location or slug.
func (s *Service) GetMenu(ctx context.Context, location, slug string) (*PublicMenu, error) {
	var menu *model.SiteMenu
	var err error

	if slug != "" {
		menu, err = s.menus.GetBySlug(ctx, slug)
	} else if location != "" {
		menu, err = s.menus.GetByLocation(ctx, location)
	} else {
		return nil, apperror.Validation("location or slug parameter required", nil)
	}
	if err != nil {
		return nil, err
	}

	items, err := s.menus.ListItemsByMenuID(ctx, menu.ID)
	if err != nil {
		return nil, fmt.Errorf("get menu items: %w", err)
	}

	// Filter: only active, non-broken items
	filtered := filterActiveItems(items)
	tree := buildMenuTree(filtered)

	return &PublicMenu{
		ID:       menu.ID,
		Name:     menu.Name,
		Slug:     menu.Slug,
		Location: menu.Location,
		Items:    tree,
	}, nil
}

// Preview returns a draft post by preview token.
func (s *Service) Preview(ctx context.Context, rawToken string) (*PreviewResp, error) {
	h := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(h[:])

	token, err := s.previews.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	if token.ExpiresAt.Before(time.Now()) {
		return nil, &apperror.AppError{
			Code:    410,
			Message: "preview token expired",
		}
	}

	post, err := s.posts.GetBySlug(ctx, "")
	if err != nil {
		// Preview uses post ID, not slug — need a different approach
	}

	// We need to get post by ID from the preview token
	// This requires an additional interface method or loading from preview token's post_id
	// For now, we use the post ID from the token
	return nil, fmt.Errorf("not yet implemented - needs GetByID on PostReader")
}

// --- Helper functions ---

func toPostListItem(p *model.Post) PostListItem {
	item := PostListItem{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Excerpt:     p.Excerpt,
		ViewCount:   p.ViewCount,
		PublishedAt: p.PublishedAt,
	}
	if p.Author != nil {
		item.Author = &AuthorBrief{
			ID:          p.Author.ID,
			DisplayName: p.Author.DisplayName,
			AvatarURL:   p.Author.AvatarURL,
		}
	}
	return item
}

func toPostDetail(p *model.Post) PostDetail {
	d := PostDetail{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Content:     p.Content,
		ContentJSON: p.ContentJSON,
		Excerpt:     p.Excerpt,
		ExtraFields: p.ExtraFields,
		ViewCount:   p.ViewCount,
		PublishedAt: p.PublishedAt,
	}
	if p.Author != nil {
		d.Author = &AuthorBrief{
			ID:          p.Author.ID,
			DisplayName: p.Author.DisplayName,
			AvatarURL:   p.Author.AvatarURL,
		}
	}
	d.SEO = &SEOFields{
		MetaTitle:  p.MetaTitle,
		MetaDesc:   p.MetaDesc,
		OGImageURL: p.OGImageURL,
	}
	return d
}

func (s *Service) buildCategoryTree(ctx context.Context, cats []model.Category) []CategoryNode {
	// Build a map of parent_id -> children
	byParent := make(map[string][]model.Category)
	var roots []model.Category
	for _, c := range cats {
		if c.ParentID == nil {
			roots = append(roots, c)
		} else {
			byParent[*c.ParentID] = append(byParent[*c.ParentID], c)
		}
	}

	var build func(cats []model.Category) []CategoryNode
	build = func(cats []model.Category) []CategoryNode {
		nodes := make([]CategoryNode, len(cats))
		for i, c := range cats {
			count, _ := s.categories.CountPosts(ctx, c.ID)
			nodes[i] = CategoryNode{
				ID:        c.ID,
				Name:      c.Name,
				Slug:      c.Slug,
				Path:      c.Path,
				PostCount: count,
				Children:  build(byParent[c.ID]),
			}
		}
		return nodes
	}
	return build(roots)
}

func hitToSearchResult(hit map[string]any) SearchResultItem {
	item := SearchResultItem{}
	if v, ok := hit["id"].(string); ok {
		item.ID = v
	}
	if v, ok := hit["title"].(string); ok {
		item.Title = v
	}
	if v, ok := hit["slug"].(string); ok {
		item.Slug = v
	}
	if v, ok := hit["excerpt"].(string); ok {
		item.Excerpt = v
	}
	return item
}

// CommentListFilter for public comment listing.
type CommentListFilter struct {
	PostID  string
	Page    int
	PerPage int
}

// CommentListResult is the result of listing public comments.
type CommentListResult struct {
	CommentCount int64           `json:"comment_count"`
	Comments     []PublicComment `json:"comments"`
	Total        int64           `json:"-"`
	Page         int             `json:"-"`
	PerPage      int             `json:"-"`
}

func buildCommentTree(comments []model.Comment) []PublicComment {
	byParent := make(map[string][]model.Comment)
	var roots []model.Comment

	for _, c := range comments {
		if c.ParentID == nil {
			roots = append(roots, c)
		} else {
			byParent[*c.ParentID] = append(byParent[*c.ParentID], c)
		}
	}

	var build func(comments []model.Comment) []PublicComment
	build = func(comments []model.Comment) []PublicComment {
		result := make([]PublicComment, len(comments))
		for i, c := range comments {
			result[i] = PublicComment{
				ID:         c.ID,
				ParentID:   c.ParentID,
				AuthorName: c.AuthorName,
				AuthorURL:  c.AuthorURL,
				Content:    c.Content,
				IsPinned:   c.Pinned == model.ToggleYes,
				CreatedAt:  c.CreatedAt,
				Replies:    build(byParent[c.ID]),
			}
		}
		return result
	}
	return build(roots)
}

func filterActiveItems(items []*model.SiteMenuItem) []*model.SiteMenuItem {
	var active []*model.SiteMenuItem
	for _, item := range items {
		if item.Status == model.MenuItemStatusActive {
			active = append(active, item)
		}
	}
	return active
}

func buildMenuTree(items []*model.SiteMenuItem) []PublicMenuItem {
	byParent := make(map[string][]*model.SiteMenuItem)
	var roots []*model.SiteMenuItem

	for _, item := range items {
		if item.ParentID == nil {
			roots = append(roots, item)
		} else {
			byParent[*item.ParentID] = append(byParent[*item.ParentID], item)
		}
	}

	var build func(items []*model.SiteMenuItem) []PublicMenuItem
	build = func(items []*model.SiteMenuItem) []PublicMenuItem {
		result := make([]PublicMenuItem, len(items))
		for i, item := range items {
			result[i] = PublicMenuItem{
				ID:       item.ID,
				Label:    item.Label,
				URL:      item.URL,
				Target:   item.Target,
				Icon:     item.Icon,
				CSSClass: item.CSSClass,
				Children: build(byParent[item.ID]),
			}
		}
		return result
	}
	return build(roots)
}
```

**Important:** This initial version will have compilation issues because:
1. The `PostReader` interface needs a `GetBySlug` method but the existing `post.PostRepo` only has `GetByID`
2. The `CommentReader.ListByPost` needs proper params
3. The `Preview` method is incomplete

These will be fixed in the next task when we add the repo adapter layer. The key business logic is correct.

**Step 2: Verify compilation (expect some errors, fix in next task)**

Run: `go vet ./internal/public/...`
Fix any import or syntax issues.

**Step 3: Commit (WIP)**

```bash
git add internal/public/service.go
git commit -m "feat(public): add service layer with core business logic (WIP)"
```

---

## Task 6: Public Module — Repository Adapters

**Files:**
- Create: `internal/public/repository.go`

This task creates thin adapter types that implement the public module's interfaces by wrapping existing repos and adding public-specific queries (like `GetBySlug`, `ListByPost` for approved comments, `IncrementViewCount`, etc.).

**Step 1: Create repository adapters**

Create `internal/public/repository.go`:

```go
package public

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/comment"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// PostRepoAdapter wraps bun.DB to implement PostReader for public queries.
type PostRepoAdapter struct {
	db *bun.DB
}

func NewPostRepoAdapter(db *bun.DB) *PostRepoAdapter {
	return &PostRepoAdapter{db: db}
}

func (r *PostRepoAdapter) List(ctx context.Context, f PostListFilter) ([]model.Post, int64, error) {
	var posts []model.Post
	q := r.db.NewSelect().
		Model(&posts).
		Relation("Author").
		Where("p.status = ?", model.PostStatusPublished).
		Where("p.deleted_at IS NULL")

	if f.Category != "" {
		q = q.Where("p.id IN (SELECT pcm.post_id FROM sfc_site_post_category_map pcm JOIN sfc_site_categories cat ON cat.id = pcm.category_id WHERE cat.slug = ?)", f.Category)
	}
	if f.Tag != "" {
		q = q.Where("p.id IN (SELECT ptm.post_id FROM sfc_site_post_tag_map ptm JOIN sfc_site_tags tg ON tg.id = ptm.tag_id WHERE tg.slug = ?)", f.Tag)
	}

	q = q.OrderExpr("p.published_at DESC NULLS LAST")

	count, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("public posts count: %w", err)
	}

	offset := (f.Page - 1) * f.PerPage
	err = q.Limit(f.PerPage).Offset(offset).Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("public posts list: %w", err)
	}

	return posts, int64(count), nil
}

func (r *PostRepoAdapter) GetBySlug(ctx context.Context, slug string) (*model.Post, error) {
	post := new(model.Post)
	err := r.db.NewSelect().
		Model(post).
		Relation("Author").
		Where("p.slug = ?", slug).
		Where("p.deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("post not found", err)
		}
		return nil, fmt.Errorf("public get post by slug: %w", err)
	}
	return post, nil
}

func (r *PostRepoAdapter) GetByID(ctx context.Context, id string) (*model.Post, error) {
	post := new(model.Post)
	err := r.db.NewSelect().
		Model(post).
		Relation("Author").
		Where("p.id = ?", id).
		Where("p.deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("post not found", err)
		}
		return nil, fmt.Errorf("public get post by id: %w", err)
	}
	return post, nil
}

func (r *PostRepoAdapter) LoadRelations(ctx context.Context, post *model.Post) error {
	// Author is already loaded via Relation("Author") in GetBySlug
	return nil
}

func (r *PostRepoAdapter) IncrementViewCount(ctx context.Context, id string) error {
	_, err := r.db.NewUpdate().
		Model((*model.Post)(nil)).
		Set("view_count = view_count + 1").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("increment view count: %w", err)
	}
	return nil
}

// CommentRepoAdapter wraps bun.DB for public comment queries.
type CommentRepoAdapter struct {
	db *bun.DB
}

func NewCommentRepoAdapter(db *bun.DB) *CommentRepoAdapter {
	return &CommentRepoAdapter{db: db}
}

func (r *CommentRepoAdapter) ListByPost(ctx context.Context, postID string, f comment.ListFilter) ([]model.Comment, int64, error) {
	var comments []model.Comment
	q := r.db.NewSelect().
		Model(&comments).
		Where("cm.post_id = ?", postID).
		Where("cm.status = ?", model.CommentStatusApproved).
		Where("cm.parent_id IS NULL"). // top-level only
		Where("cm.deleted_at IS NULL").
		OrderExpr("cm.pinned DESC, cm.created_at ASC") // pinned first, then chronological

	count, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("public comments count: %w", err)
	}

	offset := (f.Page - 1) * f.PerPage
	err = q.Limit(f.PerPage).Offset(offset).Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("public comments list: %w", err)
	}

	// Load all replies for these top-level comments
	for i := range comments {
		replies, err := r.loadReplies(ctx, comments[i].ID)
		if err != nil {
			return nil, 0, err
		}
		comments[i].Children = replies
	}

	return comments, int64(count), nil
}

func (r *CommentRepoAdapter) loadReplies(ctx context.Context, parentID string) ([]*model.Comment, error) {
	var replies []*model.Comment
	err := r.db.NewSelect().
		Model(&replies).
		Where("parent_id = ?", parentID).
		Where("status = ?", model.CommentStatusApproved).
		Where("deleted_at IS NULL").
		OrderExpr("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("load replies: %w", err)
	}
	for i := range replies {
		children, err := r.loadReplies(ctx, replies[i].ID)
		if err != nil {
			return nil, err
		}
		replies[i].Children = children
	}
	return replies, nil
}

func (r *CommentRepoAdapter) GetChildren(ctx context.Context, parentID string) ([]*model.Comment, error) {
	var children []*model.Comment
	err := r.db.NewSelect().
		Model(&children).
		Where("parent_id = ? AND status = ? AND deleted_at IS NULL", parentID, model.CommentStatusApproved).
		OrderExpr("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("public get children: %w", err)
	}
	return children, nil
}

func (r *CommentRepoAdapter) Create(ctx context.Context, c *model.Comment) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	if err != nil {
		return fmt.Errorf("public comment create: %w", err)
	}
	return nil
}

func (r *CommentRepoAdapter) GetByID(ctx context.Context, id string) (*model.Comment, error) {
	c := new(model.Comment)
	err := r.db.NewSelect().Model(c).Where("id = ? AND deleted_at IS NULL", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("comment not found", err)
		}
		return nil, fmt.Errorf("public get comment: %w", err)
	}
	return c, nil
}

func (r *CommentRepoAdapter) GetParentChainDepth(ctx context.Context, commentID string) (int, error) {
	depth := 0
	currentID := commentID
	for depth < 5 {
		c := new(model.Comment)
		err := r.db.NewSelect().Model(c).Column("parent_id").Where("id = ?", currentID).Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}
			return 0, fmt.Errorf("parent chain: %w", err)
		}
		if c.ParentID == nil {
			break
		}
		depth++
		currentID = *c.ParentID
	}
	return depth, nil
}

// TagRepoAdapter wraps bun.DB for public tag queries with post count.
type TagRepoAdapter struct {
	db *bun.DB
}

func NewTagRepoAdapter(db *bun.DB) *TagRepoAdapter {
	return &TagRepoAdapter{db: db}
}

func (r *TagRepoAdapter) ListPublic(ctx context.Context, sort string) ([]TagWithCount, error) {
	var results []TagWithCount
	q := r.db.NewSelect().
		TableExpr("sfc_site_tags AS t").
		ColumnExpr("t.*").
		ColumnExpr("COUNT(DISTINCT ptm.post_id) AS post_count").
		Join("LEFT JOIN sfc_site_post_tag_map AS ptm ON ptm.tag_id = t.id").
		Join("LEFT JOIN sfc_site_posts AS p ON p.id = ptm.post_id AND p.status = ? AND p.deleted_at IS NULL", model.PostStatusPublished).
		GroupExpr("t.id")

	switch sort {
	case "post_count:desc":
		q = q.OrderExpr("post_count DESC, t.name ASC")
	default:
		q = q.OrderExpr("t.name ASC")
	}

	err := q.Scan(ctx, &results)
	if err != nil {
		return nil, fmt.Errorf("public tags list: %w", err)
	}
	return results, nil
}

// MenuRepoAdapter wraps bun.DB for public menu queries.
type MenuRepoAdapter struct {
	db *bun.DB
}

func NewMenuRepoAdapter(db *bun.DB) *MenuRepoAdapter {
	return &MenuRepoAdapter{db: db}
}

func (r *MenuRepoAdapter) GetByLocation(ctx context.Context, location string) (*model.SiteMenu, error) {
	menu := new(model.SiteMenu)
	err := r.db.NewSelect().Model(menu).Where("location = ?", location).Limit(1).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("menu not found", err)
		}
		return nil, fmt.Errorf("public get menu by location: %w", err)
	}
	return menu, nil
}

func (r *MenuRepoAdapter) GetBySlug(ctx context.Context, slug string) (*model.SiteMenu, error) {
	menu := new(model.SiteMenu)
	err := r.db.NewSelect().Model(menu).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("menu not found", err)
		}
		return nil, fmt.Errorf("public get menu by slug: %w", err)
	}
	return menu, nil
}

func (r *MenuRepoAdapter) ListItemsByMenuID(ctx context.Context, menuID string) ([]*model.SiteMenuItem, error) {
	var items []*model.SiteMenuItem
	err := r.db.NewSelect().
		Model(&items).
		Where("menu_id = ?", menuID).
		OrderExpr("sort_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("public list menu items: %w", err)
	}
	return items, nil
}
```

**Step 2: Fix service.go compilation issues**

Update `internal/public/service.go`:
- Fix the `ListComments` method to use proper filter type
- Fix the `Preview` method to use `PostRepoAdapter.GetByID`
- Add `GetByID` to the `PostReader` interface in `interfaces.go`

**Step 3: Verify compilation**

Run: `go build ./internal/public/...`
Expected: BUILD SUCCESS

**Step 4: Commit**

```bash
git add internal/public/repository.go internal/public/interfaces.go internal/public/service.go
git commit -m "feat(public): add repository adapters and fix service compilation"
```

---

## Task 7: Public Module — Service Tests

**Files:**
- Create: `internal/public/service_test.go`

**Step 1: Create service tests with mocks**

Create `internal/public/service_test.go` with mock implementations of all interfaces and test each service method:

- `TestListPosts_ReturnsPublishedOnly`
- `TestGetPost_NotPublished_Returns404`
- `TestGetPost_Published_ReturnsDetail`
- `TestListCategories_ReturnsTreeWithCounts`
- `TestListTags_ReturnsSorted`
- `TestSearch_EmptyQuery_ReturnsEmpty`
- `TestSearch_ValidQuery_ReturnsResults`
- `TestListComments_ReturnsApprovedTree`
- `TestCreateComment_Guest_RequiresName`
- `TestCreateComment_Guest_RequiresEmail`
- `TestCreateComment_Honeypot_MarksSpam`
- `TestCreateComment_MaxNesting_Rejected`
- `TestCreateComment_ValidGuest_Success`
- `TestGetMenu_ByLocation_Success`
- `TestGetMenu_BySlug_Success`
- `TestGetMenu_NoParam_Returns422`
- `TestPreview_ValidToken_ReturnsPost`
- `TestPreview_ExpiredToken_Returns410`

Each test creates a mock repo struct (like the existing `mockRepo` pattern in `comment/service_test.go`), constructs a service, and validates business logic.

**Step 2: Run tests**

Run: `go test ./internal/public/ -run TestService -v`
Expected: ~18 PASS

**Step 3: Commit**

```bash
git add internal/public/service_test.go
git commit -m "test(public): add comprehensive service unit tests"
```

---

## Task 8: Public Module — Handler

**Files:**
- Create: `internal/public/handler.go`

**Step 1: Create handler with 9 methods**

Create `internal/public/handler.go` following the project's handler pattern (see `comment/handler.go`, `system/handler.go`):

```go
package public

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListPosts(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}

	f := PostListFilter{
		Page:     page,
		PerPage:  perPage,
		Category: c.Query("category"),
		Tag:      c.Query("tag"),
		Locale:   c.Query("locale"),
		Sort:     c.DefaultQuery("sort", "published_at:desc"),
	}

	items, total, err := h.svc.ListPosts(c.Request.Context(), siteSlug, f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, items, total, page, perPage)
}

func (h *Handler) GetPost(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	detail, err := h.svc.GetPost(c.Request.Context(), siteSlug, c.Param("slug"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, detail)
}

func (h *Handler) ListCategories(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	nodes, err := h.svc.ListCategories(c.Request.Context(), siteSlug)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nodes)
}

func (h *Handler) ListTags(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	sort := c.DefaultQuery("sort", "name:asc")
	items, err := h.svc.ListTags(c.Request.Context(), siteSlug, sort)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

func (h *Handler) Search(c *gin.Context) {
	siteSlug := c.GetString("site_slug")
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	items, total, err := h.svc.Search(c.Request.Context(), siteSlug, query, page, perPage)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, items, total, page, perPage)
}

func (h *Handler) ListComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 50 {
		perPage = 50
	}

	result, err := h.svc.ListComments(c.Request.Context(), c.Param("slug"), page, perPage)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, gin.H{
		"comment_count": result.CommentCount,
		"comments":      result.Comments,
	}, result.Total, result.Page, result.PerPage)
}

func (h *Handler) CreateComment(c *gin.Context) {
	var req CreateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request body", err))
		return
	}

	userID, _ := c.Get("user_id")
	uid := ""
	if userID != nil {
		uid, _ = userID.(string)
	}
	userName, _ := c.Get("user_name")
	uname := ""
	if userName != nil {
		uname, _ = userName.(string)
	}
	userEmail, _ := c.Get("user_email")
	uemail := ""
	if userEmail != nil {
		uemail, _ = userEmail.(string)
	}

	resp, err := h.svc.CreateComment(
		c.Request.Context(),
		c.Param("slug"),
		&req,
		uid, uname, uemail,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, resp)
}

func (h *Handler) GetMenu(c *gin.Context) {
	location := c.Query("location")
	slug := c.Query("slug")

	menu, err := h.svc.GetMenu(c.Request.Context(), location, slug)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, menu)
}

func (h *Handler) Preview(c *gin.Context) {
	result, err := h.svc.Preview(c.Request.Context(), c.Param("token"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/public/...`
Expected: BUILD SUCCESS

**Step 3: Commit**

```bash
git add internal/public/handler.go
git commit -m "feat(public): add HTTP handler with 9 endpoints"
```

---

## Task 9: Public Module — Handler Tests

**Files:**
- Create: `internal/public/handler_test.go`

**Step 1: Create handler tests**

Create `internal/public/handler_test.go` with tests for each handler method, mocking the service layer. Follow the pattern from `comment/handler_test.go`:

- `TestHandler_ListPosts`
- `TestHandler_GetPost_NotFound`
- `TestHandler_GetPost_Success`
- `TestHandler_ListCategories`
- `TestHandler_ListTags`
- `TestHandler_Search`
- `TestHandler_ListComments`
- `TestHandler_CreateComment_InvalidBody`
- `TestHandler_CreateComment_Success`
- `TestHandler_GetMenu_NoParam`
- `TestHandler_GetMenu_ByLocation`
- `TestHandler_Preview_Success`
- `TestHandler_Preview_NotFound`

**Step 2: Run tests**

Run: `go test ./internal/public/ -v`
Expected: ~30+ PASS (service + handler tests)

**Step 3: Commit**

```bash
git add internal/public/handler_test.go
git commit -m "test(public): add handler unit tests for all 9 endpoints"
```

---

## Task 10: Feed Module — XML Types

**Files:**
- Create: `internal/feed/types.go`

**Step 1: Create XML struct definitions**

Create `internal/feed/types.go` with `encoding/xml` structs for RSS 2.0, Atom 1.0, Sitemap Index, and URL Set:

```go
package feed

import "encoding/xml"

// --- RSS 2.0 ---

type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	DC      string     `xml:"xmlns:dc,attr"`
	Content string     `xml:"xmlns:content,attr"`
	Channel RSSChannel `xml:"channel"`
}

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

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type RSSItem struct {
	Title          string `xml:"title"`
	Link           string `xml:"link"`
	GUID           GUID   `xml:"guid"`
	Description    string `xml:"description"`
	ContentEncoded string `xml:"content:encoded"`
	Creator        string `xml:"dc:creator"`
	PubDate        string `xml:"pubDate"`
	Categories     []string `xml:"category"`
}

type GUID struct {
	IsPermaLink string `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

// --- Atom 1.0 ---

type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	XMLNS   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	Link    []AtomFeedLink `xml:"link"`
	Updated string      `xml:"updated"`
	ID      string      `xml:"id"`
	Author  *AtomAuthor `xml:"author,omitempty"`
	Generator string   `xml:"generator"`
	Entries []AtomEntry `xml:"entry"`
}

type AtomFeedLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
}

type AtomEntry struct {
	Title     string        `xml:"title"`
	Link      AtomFeedLink  `xml:"link"`
	ID        string        `xml:"id"`
	Updated   string        `xml:"updated"`
	Published string        `xml:"published"`
	Author    *AtomAuthor   `xml:"author,omitempty"`
	Summary   string        `xml:"summary,omitempty"`
	Content   *AtomContent  `xml:"content,omitempty"`
}

type AtomContent struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",cdata"`
}

// --- Sitemap ---

type SitemapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	XMLNS    string    `xml:"xmlns,attr"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

type Sitemap struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod,omitempty"`
}

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc        string `xml:"loc"`
	Lastmod    string `xml:"lastmod,omitempty"`
	Changefreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/feed/...`
Expected: BUILD SUCCESS

**Step 3: Commit**

```bash
git add internal/feed/types.go
git commit -m "feat(feed): add XML struct definitions for RSS, Atom, and Sitemap"
```

---

## Task 11: Feed Module — Interfaces + Service

**Files:**
- Create: `internal/feed/interfaces.go`
- Modify: `internal/feed/service.go` (replace TODO stub)

**Step 1: Create interfaces**

Create `internal/feed/interfaces.go`:

```go
package feed

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// FeedPostReader queries published posts for feed generation.
type FeedPostReader interface {
	ListPublished(ctx context.Context, limit int, categorySlug, tagSlug string) ([]model.Post, error)
	LatestPublishedAt(ctx context.Context) (*time.Time, error)
}

// FeedCategoryReader queries categories for sitemap generation.
type FeedCategoryReader interface {
	ListAll(ctx context.Context) ([]model.Category, error)
	LatestPostDate(ctx context.Context, categoryID string) (*time.Time, error)
}

// FeedTagReader queries tags for sitemap generation.
type FeedTagReader interface {
	ListWithPosts(ctx context.Context) ([]TagWithLastmod, error)
}

// SiteConfigReader reads site configuration for feed metadata.
type SiteConfigReader interface {
	GetSiteTitle(ctx context.Context) string
	GetSiteURL(ctx context.Context) string
	GetSiteDescription(ctx context.Context) string
	GetSiteLanguage(ctx context.Context) string
}

type TagWithLastmod struct {
	model.Tag
	LastPostDate *time.Time `bun:"last_post_date"`
	PostCount    int64      `bun:"post_count"`
}
```

Note: Add `"time"` to imports.

**Step 2: Create service**

Replace `internal/feed/service.go` with full implementation:

```go
package feed

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/cache"
)

type Service struct {
	posts      FeedPostReader
	categories FeedCategoryReader
	tags       FeedTagReader
	site       SiteConfigReader
	cache      *cache.Client
}

func NewService(
	posts FeedPostReader,
	categories FeedCategoryReader,
	tags FeedTagReader,
	site SiteConfigReader,
	cacheClient *cache.Client,
) *Service {
	return &Service{posts: posts, categories: categories, tags: tags, site: site, cache: cacheClient}
}

func (s *Service) GenerateRSS(ctx context.Context, limit int, categorySlug, tagSlug string) ([]byte, error) {
	posts, err := s.posts.ListPublished(ctx, limit, categorySlug, tagSlug)
	if err != nil {
		return nil, fmt.Errorf("feed rss posts: %w", err)
	}

	siteURL := s.site.GetSiteURL(ctx)
	feed := RSSFeed{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		DC:      "http://purl.org/dc/elements/1.1/",
		Content: "http://purl.org/rss/1.0/modules/content/",
		Channel: RSSChannel{
			Title:         s.site.GetSiteTitle(ctx),
			Link:          siteURL,
			Description:   s.site.GetSiteDescription(ctx),
			Language:      s.site.GetSiteLanguage(ctx),
			LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
			Generator:     "Sky Flux CMS",
			AtomLink: AtomLink{
				Href: siteURL + "/feed/rss.xml",
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: postsToRSSItems(posts, siteURL),
		},
	}

	data, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal rss: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}

func (s *Service) GenerateAtom(ctx context.Context, limit int, categorySlug, tagSlug string) ([]byte, error) {
	posts, err := s.posts.ListPublished(ctx, limit, categorySlug, tagSlug)
	if err != nil {
		return nil, fmt.Errorf("feed atom posts: %w", err)
	}

	siteURL := s.site.GetSiteURL(ctx)
	feed := AtomFeed{
		XMLNS:     "http://www.w3.org/2005/Atom",
		Title:     s.site.GetSiteTitle(ctx),
		Link: []AtomFeedLink{
			{Href: siteURL, Rel: "alternate", Type: "text/html"},
			{Href: siteURL + "/feed/atom.xml", Rel: "self", Type: "application/atom+xml"},
		},
		Updated:   time.Now().UTC().Format(time.RFC3339),
		ID:        siteURL + "/",
		Generator: "Sky Flux CMS",
		Entries:   postsToAtomEntries(posts, siteURL),
	}

	data, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal atom: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}

func (s *Service) GenerateSitemapIndex(ctx context.Context) ([]byte, error) {
	siteURL := s.site.GetSiteURL(ctx)
	latestPost, _ := s.posts.LatestPublishedAt(ctx)

	idx := SitemapIndex{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		Sitemaps: []Sitemap{
			{Loc: siteURL + "/sitemap-posts.xml", Lastmod: formatLastmod(latestPost)},
			{Loc: siteURL + "/sitemap-categories.xml"},
			{Loc: siteURL + "/sitemap-tags.xml"},
		},
	}

	data, err := xml.MarshalIndent(idx, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal sitemap index: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}

func (s *Service) GeneratePostsSitemap(ctx context.Context) ([]byte, error) {
	posts, err := s.posts.ListPublished(ctx, 50000, "", "")
	if err != nil {
		return nil, fmt.Errorf("sitemap posts: %w", err)
	}

	siteURL := s.site.GetSiteURL(ctx)
	now := time.Now()
	urls := make([]URL, len(posts))
	for i, p := range posts {
		priority, changefreq := postPriority(p, now)
		urls[i] = URL{
			Loc:        siteURL + "/" + p.Slug,
			Lastmod:    p.UpdatedAt.UTC().Format(time.RFC3339),
			Priority:   priority,
			Changefreq: changefreq,
		}
	}

	set := URLSet{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls}
	data, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal posts sitemap: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}

func (s *Service) GenerateCategoriesSitemap(ctx context.Context) ([]byte, error) {
	cats, err := s.categories.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("sitemap categories: %w", err)
	}

	siteURL := s.site.GetSiteURL(ctx)
	urls := make([]URL, len(cats))
	for i, c := range cats {
		priority := "0.5"
		if c.ParentID == nil {
			priority = "0.6"
		}
		lastmod, _ := s.categories.LatestPostDate(ctx, c.ID)
		urls[i] = URL{
			Loc:      siteURL + "/category/" + c.Slug,
			Lastmod:  formatLastmod(lastmod),
			Priority: priority,
		}
	}

	set := URLSet{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls}
	data, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal categories sitemap: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}

func (s *Service) GenerateTagsSitemap(ctx context.Context) ([]byte, error) {
	tags, err := s.tags.ListWithPosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("sitemap tags: %w", err)
	}

	siteURL := s.site.GetSiteURL(ctx)
	urls := make([]URL, 0, len(tags))
	for _, t := range tags {
		if t.PostCount == 0 {
			continue
		}
		urls = append(urls, URL{
			Loc:      siteURL + "/tag/" + t.Slug,
			Lastmod:  formatLastmod(t.LastPostDate),
			Priority: "0.4",
		})
	}

	set := URLSet{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls}
	data, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal tags sitemap: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}

// --- helpers ---

func postsToRSSItems(posts []model.Post, siteURL string) []RSSItem {
	items := make([]RSSItem, len(posts))
	for i, p := range posts {
		link := siteURL + "/" + p.Slug
		creator := ""
		if p.Author != nil {
			creator = p.Author.DisplayName
		}
		items[i] = RSSItem{
			Title:          p.Title,
			Link:           link,
			GUID:           GUID{IsPermaLink: "true", Value: link},
			Description:    p.Excerpt,
			ContentEncoded: p.Content,
			Creator:        creator,
			PubDate:        formatPubDate(p.PublishedAt),
		}
	}
	return items
}

func postsToAtomEntries(posts []model.Post, siteURL string) []AtomEntry {
	entries := make([]AtomEntry, len(posts))
	for i, p := range posts {
		link := siteURL + "/" + p.Slug
		var author *AtomAuthor
		if p.Author != nil {
			author = &AtomAuthor{Name: p.Author.DisplayName}
		}
		entries[i] = AtomEntry{
			Title:     p.Title,
			Link:      AtomFeedLink{Href: link, Rel: "alternate"},
			ID:        link,
			Updated:   p.UpdatedAt.UTC().Format(time.RFC3339),
			Published: formatRFC3339(p.PublishedAt),
			Author:    author,
			Summary:   p.Excerpt,
			Content:   &AtomContent{Type: "html", Value: p.Content},
		}
	}
	return entries
}

func postPriority(p model.Post, now time.Time) (string, string) {
	if p.PostType == "page" {
		return "0.6", "monthly"
	}
	if p.PublishedAt == nil {
		return "0.5", "monthly"
	}
	age := now.Sub(*p.PublishedAt)
	switch {
	case age <= 7*24*time.Hour:
		return "0.9", "daily"
	case age <= 30*24*time.Hour:
		return "0.8", "weekly"
	case age <= 90*24*time.Hour:
		return "0.7", "weekly"
	default:
		return "0.5", "monthly"
	}
}

func formatPubDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC1123Z)
}

func formatRFC3339(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func formatLastmod(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
```

**Step 3: Verify compilation**

Run: `go build ./internal/feed/...`
Expected: BUILD SUCCESS

**Step 4: Commit**

```bash
git add internal/feed/interfaces.go internal/feed/service.go
git commit -m "feat(feed): add interfaces and service for RSS/Atom/Sitemap generation"
```

---

## Task 12: Feed Module — Repository Adapters

**Files:**
- Create: `internal/feed/repository.go`

Create `internal/feed/repository.go` with DB adapters implementing `FeedPostReader`, `FeedCategoryReader`, `FeedTagReader`, and `SiteConfigReader`. Query patterns follow existing repo patterns (bun ORM, site schema).

**Step 1: Implement repo adapters**

**Step 2: Verify compilation**

Run: `go build ./internal/feed/...`
Expected: BUILD SUCCESS

**Step 3: Commit**

```bash
git add internal/feed/repository.go
git commit -m "feat(feed): add repository adapters for feed data queries"
```

---

## Task 13: Feed Module — Service Tests

**Files:**
- Create: `internal/feed/service_test.go`

Test each generation method with mock repos:

- `TestGenerateRSS_ValidPosts`
- `TestGenerateRSS_EmptyPosts`
- `TestGenerateRSS_WithCategoryFilter`
- `TestGenerateAtom_ValidPosts`
- `TestGenerateSitemapIndex`
- `TestGeneratePostsSitemap_PriorityRules`
- `TestGenerateCategoriesSitemap_RootVsChild`
- `TestGenerateTagsSitemap_SkipsEmptyTags`

Validate XML output with `xml.Unmarshal` to ensure well-formed XML.

**Step 1: Write tests**
**Step 2: Run tests**

Run: `go test ./internal/feed/ -v`
Expected: ~8 PASS

**Step 3: Commit**

```bash
git add internal/feed/service_test.go
git commit -m "test(feed): add service unit tests for all XML generation methods"
```

---

## Task 14: Feed Module — Handler

**Files:**
- Modify: `internal/feed/handler.go` (replace TODO stub)

**Step 1: Create handler with 6 methods**

Replace `internal/feed/handler.go`:

```go
package feed

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RSSFeed(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 50 {
		limit = 50
	}
	data, err := h.svc.GenerateRSS(c.Request.Context(), limit, c.Query("category"), c.Query("tag"))
	if err != nil {
		c.String(http.StatusInternalServerError, "feed generation error")
		return
	}
	writeXML(c, "application/rss+xml; charset=utf-8", data, 3600)
}

func (h *Handler) AtomFeed(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 50 {
		limit = 50
	}
	data, err := h.svc.GenerateAtom(c.Request.Context(), limit, c.Query("category"), c.Query("tag"))
	if err != nil {
		c.String(http.StatusInternalServerError, "feed generation error")
		return
	}
	writeXML(c, "application/atom+xml; charset=utf-8", data, 3600)
}

func (h *Handler) SitemapIndex(c *gin.Context) {
	data, err := h.svc.GenerateSitemapIndex(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "sitemap generation error")
		return
	}
	writeXML(c, "application/xml; charset=utf-8", data, 3600)
}

func (h *Handler) SitemapPosts(c *gin.Context) {
	data, err := h.svc.GeneratePostsSitemap(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "sitemap generation error")
		return
	}
	writeXML(c, "application/xml; charset=utf-8", data, 3600)
}

func (h *Handler) SitemapCategories(c *gin.Context) {
	data, err := h.svc.GenerateCategoriesSitemap(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "sitemap generation error")
		return
	}
	writeXML(c, "application/xml; charset=utf-8", data, 3600)
}

func (h *Handler) SitemapTags(c *gin.Context) {
	data, err := h.svc.GenerateTagsSitemap(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "sitemap generation error")
		return
	}
	writeXML(c, "application/xml; charset=utf-8", data, 3600)
}

func writeXML(c *gin.Context, contentType string, data []byte, maxAge int) {
	etag := fmt.Sprintf(`"%x"`, md5.Sum(data))
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	c.Header("ETag", etag)
	c.Data(http.StatusOK, contentType, data)
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/feed/...`
Expected: BUILD SUCCESS

**Step 3: Commit**

```bash
git add internal/feed/handler.go
git commit -m "feat(feed): add handler with 6 XML endpoints"
```

---

## Task 15: Feed Module — Handler Tests

**Files:**
- Modify: `internal/feed/handler_test.go` (replace placeholder)

Test each handler endpoint:
- `TestHandler_RSSFeed_ContentType`
- `TestHandler_AtomFeed_ContentType`
- `TestHandler_SitemapIndex_ContentType`
- `TestHandler_SitemapPosts`
- `TestHandler_SitemapCategories`
- `TestHandler_SitemapTags`
- `TestHandler_CacheHeaders`

Verify Content-Type headers, Cache-Control, ETag, and XML well-formedness.

**Step 1: Write tests**
**Step 2: Run tests**

Run: `go test ./internal/feed/ -v`
Expected: ~15 PASS (service + handler)

**Step 3: Commit**

```bash
git add internal/feed/handler_test.go
git commit -m "test(feed): add handler unit tests for all 6 XML endpoints"
```

---

## Task 16: Router Integration — Wire Public + Feed Modules

**Files:**
- Modify: `internal/router/router.go`

**Step 1: Add imports and wire modules into router**

Add to imports in `router.go`:

```go
"github.com/sky-flux/cms/internal/feed"
"github.com/sky-flux/cms/internal/public"
```

Add before the `// ── API Registry` section (after redirect routes):

```go
// ── Feed & Sitemap (no auth, site via Host header) ──────────
feedPostRepo := feed.NewPostRepoAdapter(db)
feedCatRepo := feed.NewCategoryRepoAdapter(db)
feedTagRepo := feed.NewTagRepoAdapter(db)
feedSiteConfig := feed.NewSiteConfigAdapter(settingsRepo)
feedSvc := feed.NewService(feedPostRepo, feedCatRepo, feedTagRepo, feedSiteConfig, cacheClient)
feedHandler := feed.NewHandler(feedSvc)

feeds := engine.Group("")
feeds.Use(middleware.InstallationGuard(setupSvc, "/health", "/api/v1/setup/"))
feeds.Use(middleware.SiteResolver(siteLookup))
feeds.Use(middleware.Schema(db))
feeds.GET("/feed/rss.xml", feedHandler.RSSFeed)
feeds.GET("/feed/atom.xml", feedHandler.AtomFeed)
feeds.GET("/sitemap.xml", feedHandler.SitemapIndex)
feeds.GET("/sitemap-posts.xml", feedHandler.SitemapPosts)
feeds.GET("/sitemap-categories.xml", feedHandler.SitemapCategories)
feeds.GET("/sitemap-tags.xml", feedHandler.SitemapTags)

// ── Public Headless API (API Key auth, site via Host header) ─
publicPostRepo := public.NewPostRepoAdapter(db)
publicCommentRepo := public.NewCommentRepoAdapter(db)
publicTagRepo := public.NewTagRepoAdapter(db)
publicMenuRepo := public.NewMenuRepoAdapter(db)
publicSvc := public.NewService(publicPostRepo, catRepo, publicTagRepo, publicCommentRepo, publicMenuRepo, previewRepo, searchClient, cacheClient)
publicHandler := public.NewHandler(publicSvc)

apiKeyMW := middleware.APIKey(apikeyRepo)
publicAPI := engine.Group("/api/public/v1")
publicAPI.Use(middleware.InstallationGuard(setupSvc, "/health", "/api/v1/setup/"))
publicAPI.Use(middleware.SiteResolver(siteLookup))
publicAPI.Use(middleware.Schema(db))
publicAPI.Use(apiKeyMW)
publicAPI.GET("/posts", publicHandler.ListPosts)
publicAPI.GET("/posts/:slug", publicHandler.GetPost)
publicAPI.GET("/categories", publicHandler.ListCategories)
publicAPI.GET("/tags", publicHandler.ListTags)
publicAPI.GET("/search", publicHandler.Search)
publicAPI.GET("/posts/:slug/comments", publicHandler.ListComments)
publicAPI.POST("/posts/:slug/comments", middleware.RateLimit(rdb, "ratelimit:comment", 30*time.Second), publicHandler.CreateComment)
publicAPI.GET("/menus", publicHandler.GetMenu)

// Preview (no API Key — token-based auth)
previewAPI := engine.Group("/api/public/v1")
previewAPI.Use(middleware.InstallationGuard(setupSvc, "/health", "/api/v1/setup/"))
previewAPI.Use(middleware.SiteResolver(siteLookup))
previewAPI.Use(middleware.Schema(db))
previewAPI.GET("/preview/:token", publicHandler.Preview)
```

**Step 2: Verify compilation**

Run: `go build ./cmd/cms/...`
Expected: BUILD SUCCESS

**Step 3: Commit**

```bash
git add internal/router/router.go
git commit -m "feat(router): wire public API and feed/sitemap modules with middleware"
```

---

## Task 17: Run All Tests

**Files:** (none — verification only)

**Step 1: Run full test suite**

Run: `go test ./... 2>&1 | tail -40`
Expected: ALL PASS

**Step 2: Run static analysis**

Run: `go vet ./...`
Expected: No warnings

**Step 3: Fix any failures**

If tests fail, debug and fix before proceeding.

---

## Task 18: Final Commit + Update Memory

**Step 1: Verify git status**

Run: `git status`
Verify all changes are committed.

**Step 2: Update MEMORY.md**

Add Batch 8 completion entry to `MEMORY.md` with:
- Public API (9) + Feed/Sitemap (6) = 15 new endpoints
- 2 new middleware (API Key + Rate Limit)
- Total: 138 endpoints (123 + 15)
- Test count

**Step 3: Final commit**

If memory was updated:
```bash
git add -A && git commit -m "docs: update memory with batch 8 completion"
```

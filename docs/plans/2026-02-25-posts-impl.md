# Batch 6: Posts Module Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the 19-endpoint Posts module with CRUD, status transitions, revisions, translations, and preview tokens.

**Architecture:** Single `internal/post/` package with file-level separation by sub-feature. Follows existing patterns from `internal/category/`, `internal/tag/`, `internal/media/`. All repositories use uptrace/bun, services depend on interfaces for testability, handlers delegate to service layer.

**Tech Stack:** Go / Gin / uptrace/bun / PostgreSQL / Meilisearch (`pkg/search`) / Redis (`pkg/cache`) / `pkg/crypto` (tokens) / `pkg/audit` (logging)

**Reference files for patterns:**
- Handler pattern: `internal/category/handler.go`
- Service pattern: `internal/category/service.go`
- Repository pattern: `internal/category/repository.go`
- Interface pattern: `internal/category/interfaces.go`
- DTO pattern: `internal/category/dto.go`
- Test pattern: `internal/tag/service_test.go` (mock + testEnv struct)
- Router registration: `internal/router/router.go:236-311`
- API meta: `internal/router/api_meta.go`
- Apperror: `internal/pkg/apperror/errors.go`
- Audit: `internal/pkg/audit/audit.go`
- Crypto: `internal/pkg/crypto/token.go`
- Search: `internal/pkg/search/client.go`

**Context values from middleware (set by auth/site middleware):**
- `c.GetString("user_id")` — JWT subject (user UUID)
- `c.GetString("site_slug")` — site slug from SiteResolver
- `c.GetString("site_id")` — site ID from SiteResolver

---

### Task 1: Add ErrVersionConflict to apperror

**Files:**
- Modify: `internal/pkg/apperror/errors.go`

**Step 1: Add the sentinel error and constructor**

In `internal/pkg/apperror/errors.go`, add to the sentinel vars block:

```go
ErrVersionConflict = errors.New("version conflict")
```

Add to the `HTTPStatusCode` switch (before `default`):

```go
case errors.Is(err, ErrVersionConflict):
    return http.StatusConflict
```

Add constructor:

```go
func VersionConflict(msg string, err error) *AppError {
    return &AppError{Code: http.StatusConflict, Message: msg, Err: errors.Join(ErrVersionConflict, err)}
}
```

**Step 2: Run existing tests**

Run: `go test ./internal/pkg/apperror/...`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/pkg/apperror/errors.go
git commit -m "feat(apperror): add ErrVersionConflict sentinel for optimistic locking"
```

---

### Task 2: Interfaces + DTOs + Slug utility

**Files:**
- Create: `internal/post/interfaces.go`
- Create: `internal/post/dto.go`
- Create: `internal/post/slug.go`
- Create: `internal/post/slug_test.go`

**Step 1: Create interfaces.go**

```go
package post

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/search"
)

// PostRepository handles sfc_site_posts table.
type PostRepository interface {
	List(ctx context.Context, f ListFilter) ([]model.Post, int64, error)
	GetByID(ctx context.Context, id string) (*model.Post, error)
	GetByIDUnscoped(ctx context.Context, id string) (*model.Post, error) // includes soft-deleted
	Create(ctx context.Context, post *model.Post) error
	Update(ctx context.Context, post *model.Post, expectedVersion int) error
	SoftDelete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	SlugExists(ctx context.Context, slug, excludeID string) (bool, error)
	UpdateStatus(ctx context.Context, id string, status model.PostStatus) error
	SyncCategories(ctx context.Context, postID string, categoryIDs []string, primaryID string) error
	SyncTags(ctx context.Context, postID string, tagIDs []string) error
	LoadRelations(ctx context.Context, post *model.Post) error
}

// RevisionRepository handles sfc_site_post_revisions.
type RevisionRepository interface {
	List(ctx context.Context, postID string) ([]model.PostRevision, error)
	GetByID(ctx context.Context, id string) (*model.PostRevision, error)
	Create(ctx context.Context, rev *model.PostRevision) error
}

// TranslationRepository handles sfc_site_post_translations.
type TranslationRepository interface {
	List(ctx context.Context, postID string) ([]model.PostTranslation, error)
	Get(ctx context.Context, postID, locale string) (*model.PostTranslation, error)
	Upsert(ctx context.Context, t *model.PostTranslation) error
	Delete(ctx context.Context, postID, locale string) error
}

// PreviewTokenRepository handles sfc_site_preview_tokens.
type PreviewTokenRepository interface {
	List(ctx context.Context, postID string) ([]model.PreviewToken, error)
	Create(ctx context.Context, token *model.PreviewToken) error
	CountActive(ctx context.Context, postID string) (int, error)
	DeleteAll(ctx context.Context, postID string) (int64, error)
	DeleteByID(ctx context.Context, id string) error
	GetByHash(ctx context.Context, hash string) (*model.PreviewToken, error)
}

// ListFilter holds query parameters for listing posts.
type ListFilter struct {
	Page           int
	PerPage        int
	Status         string
	Query          string
	CategoryID     string
	TagID          string
	AuthorID       string
	Sort           string
	IncludeDeleted bool
}

// These interfaces represent external dependencies injected into the service.
// They are defined here as narrow interfaces following the Go convention.
var (
	_ audit.Logger   = (*audit.Service)(nil)    // compile-time check
	_ *search.Client = (*search.Client)(nil)
)
```

**Step 2: Create dto.go**

```go
package post

import (
	"encoding/json"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

type CreatePostReq struct {
	Title             string          `json:"title" binding:"required,max=500"`
	Slug              string          `json:"slug" binding:"omitempty,max=200"`
	Content           string          `json:"content"`
	ContentJSON       json.RawMessage `json:"content_json"`
	Excerpt           string          `json:"excerpt"`
	Status            string          `json:"status" binding:"omitempty,oneof=draft published scheduled"`
	ScheduledAt       *time.Time      `json:"scheduled_at"`
	CoverImageID      *string         `json:"cover_image_id"`
	CategoryIDs       []string        `json:"category_ids"`
	PrimaryCategoryID string          `json:"primary_category_id"`
	TagIDs            []string        `json:"tag_ids"`
	MetaTitle         string          `json:"meta_title" binding:"max=200"`
	MetaDescription   string          `json:"meta_description" binding:"max=500"`
	OGImageURL        string          `json:"og_image_url"`
	ExtraFields       json.RawMessage `json:"extra_fields"`
}

type UpdatePostReq struct {
	Title             *string         `json:"title" binding:"omitempty,max=500"`
	Slug              *string         `json:"slug" binding:"omitempty,max=200"`
	Content           *string         `json:"content"`
	ContentJSON       json.RawMessage `json:"content_json"`
	Excerpt           *string         `json:"excerpt"`
	Status            *string         `json:"status" binding:"omitempty,oneof=draft published scheduled archived"`
	ScheduledAt       *time.Time      `json:"scheduled_at"`
	CoverImageID      *string         `json:"cover_image_id"`
	CategoryIDs       []string        `json:"category_ids"`
	PrimaryCategoryID *string         `json:"primary_category_id"`
	TagIDs            []string        `json:"tag_ids"`
	MetaTitle         *string         `json:"meta_title" binding:"omitempty,max=200"`
	MetaDescription   *string         `json:"meta_description" binding:"omitempty,max=500"`
	OGImageURL        *string         `json:"og_image_url"`
	ExtraFields       json.RawMessage `json:"extra_fields"`
	Version           int             `json:"version" binding:"required,min=1"`
}

type UpsertTranslationReq struct {
	Title           string          `json:"title" binding:"max=500"`
	Excerpt         string          `json:"excerpt"`
	Content         string          `json:"content"`
	ContentJSON     json.RawMessage `json:"content_json"`
	MetaTitle       string          `json:"meta_title" binding:"max=200"`
	MetaDescription string          `json:"meta_description" binding:"max=500"`
	OGImageURL      string          `json:"og_image_url"`
}

// --- Response DTOs ---

type PostResp struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Status      string           `json:"status"`
	Excerpt     string           `json:"excerpt,omitempty"`
	Content     string           `json:"content,omitempty"`
	ContentJSON json.RawMessage  `json:"content_json,omitempty"`
	Author      *AuthorResp      `json:"author,omitempty"`
	CoverImage  *CoverImageResp  `json:"cover_image,omitempty"`
	Categories  []CategoryBrief  `json:"categories,omitempty"`
	Tags        []TagBrief       `json:"tags,omitempty"`
	SEO         *SEOResp         `json:"seo,omitempty"`
	ExtraFields json.RawMessage  `json:"extra_fields,omitempty"`
	ViewCount   int64            `json:"view_count"`
	Version     int              `json:"version"`
	ScheduledAt *time.Time       `json:"scheduled_at,omitempty"`
	PublishedAt *time.Time       `json:"published_at,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type PostListItem struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Slug        string          `json:"slug"`
	Status      string          `json:"status"`
	Author      *AuthorResp     `json:"author,omitempty"`
	CoverImage  *CoverImageResp `json:"cover_image,omitempty"`
	Categories  []CategoryBrief `json:"categories,omitempty"`
	Tags        []TagBrief      `json:"tags,omitempty"`
	ViewCount   int64           `json:"view_count"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type AuthorResp struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type CoverImageResp struct {
	ID            string          `json:"id"`
	URL           string          `json:"url"`
	WebpURL       string          `json:"webp_url,omitempty"`
	ThumbnailURLs json.RawMessage `json:"thumbnail_urls,omitempty"`
}

type CategoryBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type TagBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type SEOResp struct {
	MetaTitle       string `json:"meta_title,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	OGImageURL      string `json:"og_image_url,omitempty"`
}

type RevisionResp struct {
	ID          string      `json:"id"`
	Version     int         `json:"version"`
	Editor      *AuthorResp `json:"editor,omitempty"`
	DiffSummary string      `json:"diff_summary,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

type TranslationResp struct {
	Locale          string          `json:"locale"`
	Title           string          `json:"title,omitempty"`
	Excerpt         string          `json:"excerpt,omitempty"`
	Content         string          `json:"content,omitempty"`
	ContentJSON     json.RawMessage `json:"content_json,omitempty"`
	MetaTitle       string          `json:"meta_title,omitempty"`
	MetaDescription string          `json:"meta_description,omitempty"`
	OGImageURL      string          `json:"og_image_url,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type TranslationListItem struct {
	Locale    string    `json:"locale"`
	Title     string    `json:"title,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PreviewTokenResp struct {
	PreviewURL string    `json:"preview_url,omitempty"`
	Token      string    `json:"token,omitempty"`
	ID         string    `json:"id,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	ActiveCount int      `json:"active_count,omitempty"`
}

// --- Status helpers ---

var statusMap = map[model.PostStatus]string{
	model.PostStatusDraft:     "draft",
	model.PostStatusScheduled: "scheduled",
	model.PostStatusPublished: "published",
	model.PostStatusArchived:  "archived",
}

var statusReverseMap = map[string]model.PostStatus{
	"draft":     model.PostStatusDraft,
	"scheduled": model.PostStatusScheduled,
	"published": model.PostStatusPublished,
	"archived":  model.PostStatusArchived,
}

func statusString(s model.PostStatus) string {
	if v, ok := statusMap[s]; ok {
		return v
	}
	return "unknown"
}

func parseStatus(s string) (model.PostStatus, bool) {
	v, ok := statusReverseMap[s]
	return v, ok
}

// --- Converters ---

func ToPostResp(p *model.Post) PostResp {
	resp := PostResp{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Status:      statusString(p.Status),
		Excerpt:     p.Excerpt,
		Content:     p.Content,
		ContentJSON: p.ContentJSON,
		ExtraFields: p.ExtraFields,
		ViewCount:   p.ViewCount,
		Version:     p.Version,
		ScheduledAt: p.ScheduledAt,
		PublishedAt:  p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if p.MetaTitle != "" || p.MetaDesc != "" || p.OGImageURL != "" {
		resp.SEO = &SEOResp{
			MetaTitle:       p.MetaTitle,
			MetaDescription: p.MetaDesc,
			OGImageURL:      p.OGImageURL,
		}
	}

	if p.Author != nil {
		resp.Author = &AuthorResp{
			ID:          p.Author.ID,
			DisplayName: p.Author.DisplayName,
			AvatarURL:   p.Author.AvatarURL,
		}
	}

	return resp
}

func ToPostListItem(p *model.Post) PostListItem {
	item := PostListItem{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Status:      statusString(p.Status),
		ViewCount:   p.ViewCount,
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if p.Author != nil {
		item.Author = &AuthorResp{
			ID:          p.Author.ID,
			DisplayName: p.Author.DisplayName,
			AvatarURL:   p.Author.AvatarURL,
		}
	}

	return item
}

func ToRevisionResp(r *model.PostRevision) RevisionResp {
	return RevisionResp{
		ID:          r.ID,
		Version:     r.Version,
		DiffSummary: r.DiffSummary,
		CreatedAt:   r.CreatedAt,
	}
}

func ToTranslationResp(t *model.PostTranslation) TranslationResp {
	return TranslationResp{
		Locale:          t.Locale,
		Title:           t.Title,
		Excerpt:         t.Excerpt,
		Content:         t.Content,
		ContentJSON:     t.ContentJSON,
		MetaTitle:       t.MetaTitle,
		MetaDescription: t.MetaDesc,
		OGImageURL:      t.OGImageURL,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

func ToTranslationListItem(t *model.PostTranslation) TranslationListItem {
	return TranslationListItem{
		Locale:    t.Locale,
		Title:     t.Title,
		UpdatedAt: t.UpdatedAt,
	}
}
```

**Step 3: Create slug.go**

```go
package post

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var nonAlphanumDash = regexp.MustCompile(`[^a-z0-9-]+`)
var multiDash = regexp.MustCompile(`-{2,}`)

// GenerateSlug creates a URL-friendly slug from an English title.
func GenerateSlug(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = nonAlphanumDash.ReplaceAllString(s, "-")
	s = multiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 200 {
		s = s[:200]
		s = strings.TrimRight(s, "-")
	}
	return s
}

// UniqueSlug generates a slug and resolves collisions by appending -2, -3, etc.
func UniqueSlug(ctx context.Context, title string, excludeID string, existsFn func(ctx context.Context, slug, excludeID string) (bool, error)) (string, error) {
	base := GenerateSlug(title)
	if base == "" {
		base = "untitled"
	}

	slug := base
	for i := 2; i <= 100; i++ {
		exists, err := existsFn(ctx, slug, excludeID)
		if err != nil {
			return "", fmt.Errorf("check slug uniqueness: %w", err)
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
	return "", fmt.Errorf("could not generate unique slug after 100 attempts")
}
```

**Step 4: Create slug_test.go — write failing tests first**

```go
package post

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSlug_BasicEnglish(t *testing.T) {
	assert.Equal(t, "my-first-post", GenerateSlug("My First Post"))
}

func TestGenerateSlug_SpecialChars(t *testing.T) {
	assert.Equal(t, "hello-world-2024", GenerateSlug("Hello, World! 2024"))
}

func TestGenerateSlug_LeadingTrailingSpaces(t *testing.T) {
	assert.Equal(t, "trimmed", GenerateSlug("  trimmed  "))
}

func TestGenerateSlug_MultipleDashes(t *testing.T) {
	assert.Equal(t, "a-b-c", GenerateSlug("a---b---c"))
}

func TestGenerateSlug_MaxLength(t *testing.T) {
	long := strings.Repeat("a", 250)
	result := GenerateSlug(long)
	assert.LessOrEqual(t, len(result), 200)
}

func TestGenerateSlug_EmptyTitle(t *testing.T) {
	assert.Equal(t, "", GenerateSlug(""))
}

func TestUniqueSlug_NoCollision(t *testing.T) {
	fn := func(_ context.Context, _, _ string) (bool, error) {
		return false, nil
	}
	slug, err := UniqueSlug(context.Background(), "Test Post", "", fn)
	require.NoError(t, err)
	assert.Equal(t, "test-post", slug)
}

func TestUniqueSlug_WithCollision(t *testing.T) {
	call := 0
	fn := func(_ context.Context, slug, _ string) (bool, error) {
		call++
		if slug == "test-post" {
			return true, nil // first one exists
		}
		return false, nil
	}
	slug, err := UniqueSlug(context.Background(), "Test Post", "", fn)
	require.NoError(t, err)
	assert.Equal(t, "test-post-2", slug)
}

func TestUniqueSlug_EmptyTitle(t *testing.T) {
	fn := func(_ context.Context, _, _ string) (bool, error) {
		return false, nil
	}
	slug, err := UniqueSlug(context.Background(), "", "", fn)
	require.NoError(t, err)
	assert.Equal(t, "untitled", slug)
}
```

> Note: Add `"strings"` import to slug_test.go for TestGenerateSlug_MaxLength.

**Step 5: Run tests**

Run: `go test ./internal/post/... -v -run TestGenerateSlug`
Expected: PASS

Run: `go test ./internal/post/... -v -run TestUniqueSlug`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/post/interfaces.go internal/post/dto.go internal/post/slug.go internal/post/slug_test.go
git commit -m "feat(post): add interfaces, DTOs, and slug generation utility"
```

---

### Task 3: Repository implementations

**Files:**
- Create: `internal/post/repository.go`
- Create: `internal/post/repository_revision.go`
- Create: `internal/post/repository_translation.go`
- Create: `internal/post/repository_preview.go`

**Step 1: Create repository.go**

```go
package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

type PostRepo struct {
	db *bun.DB
}

func NewPostRepo(db *bun.DB) *PostRepo {
	return &PostRepo{db: db}
}

func (r *PostRepo) List(ctx context.Context, f ListFilter) ([]model.Post, int64, error) {
	var posts []model.Post
	q := r.db.NewSelect().
		Model(&posts).
		Relation("Author")

	if f.Status != "" {
		if s, ok := parseStatus(f.Status); ok {
			q = q.Where("p.status = ?", s)
		}
	}

	if f.AuthorID != "" {
		q = q.Where("p.author_id = ?", f.AuthorID)
	}

	if f.CategoryID != "" {
		q = q.Where("p.id IN (SELECT post_id FROM sfc_site_post_category_map WHERE category_id = ?)", f.CategoryID)
	}

	if f.TagID != "" {
		q = q.Where("p.id IN (SELECT post_id FROM sfc_site_post_tag_map WHERE tag_id = ?)", f.TagID)
	}

	if !f.IncludeDeleted {
		q = q.Where("p.deleted_at IS NULL")
	}

	// Sort
	switch f.Sort {
	case "published_at:desc":
		q = q.OrderExpr("p.published_at DESC NULLS LAST")
	case "title:asc":
		q = q.OrderExpr("p.title ASC")
	default:
		q = q.OrderExpr("p.created_at DESC")
	}

	count, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("post list count: %w", err)
	}

	perPage := f.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	page := f.Page
	if page <= 0 {
		page = 1
	}

	err = q.Limit(perPage).Offset((page - 1) * perPage).Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("post list: %w", err)
	}

	return posts, int64(count), nil
}

func (r *PostRepo) GetByID(ctx context.Context, id string) (*model.Post, error) {
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
		return nil, fmt.Errorf("post get by id: %w", err)
	}
	return post, nil
}

func (r *PostRepo) GetByIDUnscoped(ctx context.Context, id string) (*model.Post, error) {
	post := new(model.Post)
	err := r.db.NewSelect().
		Model(post).
		Relation("Author").
		Where("p.id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("post not found", err)
		}
		return nil, fmt.Errorf("post get by id unscoped: %w", err)
	}
	return post, nil
}

func (r *PostRepo) Create(ctx context.Context, post *model.Post) error {
	_, err := r.db.NewInsert().Model(post).Exec(ctx)
	if err != nil {
		return fmt.Errorf("post create: %w", err)
	}
	return nil
}

func (r *PostRepo) Update(ctx context.Context, post *model.Post, expectedVersion int) error {
	res, err := r.db.NewUpdate().
		Model(post).
		WherePK().
		Where("version = ?", expectedVersion).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("post update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.VersionConflict("post has been modified by another user", nil)
	}
	return nil
}

func (r *PostRepo) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.NewDelete().
		Model((*model.Post)(nil)).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("post soft delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.NotFound("post not found", nil)
	}
	return nil
}

func (r *PostRepo) Restore(ctx context.Context, id string) error {
	res, err := r.db.NewUpdate().
		Model((*model.Post)(nil)).
		Set("deleted_at = NULL").
		Set("status = ?", model.PostStatusDraft).
		Where("id = ?", id).
		Where("deleted_at IS NOT NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("post restore: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.NotFound("post not found in trash", nil)
	}
	return nil
}

func (r *PostRepo) SlugExists(ctx context.Context, slug, excludeID string) (bool, error) {
	q := r.db.NewSelect().
		Model((*model.Post)(nil)).
		Where("slug = ?", slug).
		Where("deleted_at IS NULL")
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("post slug exists: %w", err)
	}
	return exists, nil
}

func (r *PostRepo) UpdateStatus(ctx context.Context, id string, status model.PostStatus) error {
	q := r.db.NewUpdate().
		Model((*model.Post)(nil)).
		Set("status = ?", status).
		Where("id = ?", id).
		Where("deleted_at IS NULL")

	if status == model.PostStatusPublished {
		q = q.Set("published_at = NOW()")
	}

	res, err := q.Exec(ctx)
	if err != nil {
		return fmt.Errorf("post update status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.NotFound("post not found", nil)
	}
	return nil
}

func (r *PostRepo) SyncCategories(ctx context.Context, postID string, categoryIDs []string, primaryID string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete existing
		_, err := tx.NewDelete().
			Model((*model.PostCategoryMap)(nil)).
			Where("post_id = ?", postID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("delete post categories: %w", err)
		}

		if len(categoryIDs) == 0 {
			return nil
		}

		// Insert new
		maps := make([]model.PostCategoryMap, len(categoryIDs))
		for i, cid := range categoryIDs {
			maps[i] = model.PostCategoryMap{
				PostID:     postID,
				CategoryID: cid,
				IsPrimary:  cid == primaryID,
			}
		}
		_, err = tx.NewInsert().Model(&maps).Exec(ctx)
		if err != nil {
			return fmt.Errorf("insert post categories: %w", err)
		}
		return nil
	})
}

func (r *PostRepo) SyncTags(ctx context.Context, postID string, tagIDs []string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewDelete().
			Model((*model.PostTagMap)(nil)).
			Where("post_id = ?", postID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("delete post tags: %w", err)
		}

		if len(tagIDs) == 0 {
			return nil
		}

		maps := make([]model.PostTagMap, len(tagIDs))
		for i, tid := range tagIDs {
			maps[i] = model.PostTagMap{PostID: postID, TagID: tid}
		}
		_, err = tx.NewInsert().Model(&maps).Exec(ctx)
		if err != nil {
			return fmt.Errorf("insert post tags: %w", err)
		}
		return nil
	})
}

func (r *PostRepo) LoadRelations(ctx context.Context, post *model.Post) error {
	// This is a placeholder for eager-loading categories, tags, cover_image
	// when the repository methods don't use Relation().
	// Individual handlers call this after GetByID to populate response fields.
	return nil
}
```

**Step 2: Create repository_revision.go**

```go
package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

type RevisionRepo struct {
	db *bun.DB
}

func NewRevisionRepo(db *bun.DB) *RevisionRepo {
	return &RevisionRepo{db: db}
}

func (r *RevisionRepo) List(ctx context.Context, postID string) ([]model.PostRevision, error) {
	var revs []model.PostRevision
	err := r.db.NewSelect().
		Model(&revs).
		Where("post_id = ?", postID).
		OrderExpr("version DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("revision list: %w", err)
	}
	return revs, nil
}

func (r *RevisionRepo) GetByID(ctx context.Context, id string) (*model.PostRevision, error) {
	rev := new(model.PostRevision)
	err := r.db.NewSelect().Model(rev).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("revision not found", err)
		}
		return nil, fmt.Errorf("revision get by id: %w", err)
	}
	return rev, nil
}

func (r *RevisionRepo) Create(ctx context.Context, rev *model.PostRevision) error {
	_, err := r.db.NewInsert().Model(rev).Exec(ctx)
	if err != nil {
		return fmt.Errorf("revision create: %w", err)
	}
	return nil
}
```

**Step 3: Create repository_translation.go**

```go
package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

type TranslationRepo struct {
	db *bun.DB
}

func NewTranslationRepo(db *bun.DB) *TranslationRepo {
	return &TranslationRepo{db: db}
}

func (r *TranslationRepo) List(ctx context.Context, postID string) ([]model.PostTranslation, error) {
	var ts []model.PostTranslation
	err := r.db.NewSelect().
		Model(&ts).
		Where("post_id = ?", postID).
		OrderExpr("locale ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("translation list: %w", err)
	}
	return ts, nil
}

func (r *TranslationRepo) Get(ctx context.Context, postID, locale string) (*model.PostTranslation, error) {
	t := new(model.PostTranslation)
	err := r.db.NewSelect().
		Model(t).
		Where("post_id = ?", postID).
		Where("locale = ?", locale).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("translation not found", err)
		}
		return nil, fmt.Errorf("translation get: %w", err)
	}
	return t, nil
}

func (r *TranslationRepo) Upsert(ctx context.Context, t *model.PostTranslation) error {
	_, err := r.db.NewInsert().
		Model(t).
		On("CONFLICT (post_id, locale) DO UPDATE").
		Set("title = EXCLUDED.title").
		Set("excerpt = EXCLUDED.excerpt").
		Set("content = EXCLUDED.content").
		Set("content_json = EXCLUDED.content_json").
		Set("meta_title = EXCLUDED.meta_title").
		Set("meta_description = EXCLUDED.meta_description").
		Set("og_image_url = EXCLUDED.og_image_url").
		Set("updated_at = NOW()").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("translation upsert: %w", err)
	}
	return nil
}

func (r *TranslationRepo) Delete(ctx context.Context, postID, locale string) error {
	res, err := r.db.NewDelete().
		Model((*model.PostTranslation)(nil)).
		Where("post_id = ?", postID).
		Where("locale = ?", locale).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("translation delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.NotFound("translation not found", nil)
	}
	return nil
}
```

**Step 4: Create repository_preview.go**

```go
package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

type PreviewRepo struct {
	db *bun.DB
}

func NewPreviewRepo(db *bun.DB) *PreviewRepo {
	return &PreviewRepo{db: db}
}

func (r *PreviewRepo) List(ctx context.Context, postID string) ([]model.PreviewToken, error) {
	var tokens []model.PreviewToken
	err := r.db.NewSelect().
		Model(&tokens).
		Where("post_id = ?", postID).
		Where("expires_at > ?", time.Now()).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("preview list: %w", err)
	}
	return tokens, nil
}

func (r *PreviewRepo) Create(ctx context.Context, token *model.PreviewToken) error {
	_, err := r.db.NewInsert().Model(token).Exec(ctx)
	if err != nil {
		return fmt.Errorf("preview create: %w", err)
	}
	return nil
}

func (r *PreviewRepo) CountActive(ctx context.Context, postID string) (int, error) {
	count, err := r.db.NewSelect().
		Model((*model.PreviewToken)(nil)).
		Where("post_id = ?", postID).
		Where("expires_at > ?", time.Now()).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("preview count active: %w", err)
	}
	return count, nil
}

func (r *PreviewRepo) DeleteAll(ctx context.Context, postID string) (int64, error) {
	res, err := r.db.NewDelete().
		Model((*model.PreviewToken)(nil)).
		Where("post_id = ?", postID).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("preview delete all: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *PreviewRepo) DeleteByID(ctx context.Context, id string) error {
	res, err := r.db.NewDelete().
		Model((*model.PreviewToken)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("preview delete by id: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperror.NotFound("preview token not found", nil)
	}
	return nil
}

func (r *PreviewRepo) GetByHash(ctx context.Context, hash string) (*model.PreviewToken, error) {
	token := new(model.PreviewToken)
	err := r.db.NewSelect().
		Model(token).
		Where("token_hash = ?", hash).
		Where("expires_at > ?", time.Now()).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("preview token not found or expired", err)
		}
		return nil, fmt.Errorf("preview get by hash: %w", err)
	}
	return token, nil
}
```

**Step 5: Verify compilation**

Run: `go build ./internal/post/...`
Expected: No errors

**Step 6: Commit**

```bash
git add internal/post/repository.go internal/post/repository_revision.go internal/post/repository_translation.go internal/post/repository_preview.go
git commit -m "feat(post): add repository implementations for posts, revisions, translations, preview tokens"
```

---

### Task 4: Service — CRUD + status transitions

**Files:**
- Create: `internal/post/service.go`
- Create: `internal/post/service_status.go`
- Create: `internal/post/service_test.go`

**Step 1: Create service.go**

```go
package post

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/search"
)

const postSearchIndex = "posts-%s" // posts-{siteSlug}

// Service implements post business logic.
type Service struct {
	posts    PostRepository
	revs     RevisionRepository
	trans    TranslationRepository
	preview  PreviewTokenRepository
	search   *search.Client
	auditLog audit.Logger
}

// NewService creates a new post service.
func NewService(
	posts PostRepository,
	revs RevisionRepository,
	trans TranslationRepository,
	preview PreviewTokenRepository,
	s *search.Client,
	auditLog audit.Logger,
) *Service {
	return &Service{
		posts:    posts,
		revs:     revs,
		trans:    trans,
		preview:  preview,
		search:   s,
		auditLog: auditLog,
	}
}

// ListPosts returns a paginated list of posts.
func (s *Service) ListPosts(ctx context.Context, f ListFilter) ([]model.Post, int64, error) {
	if f.Page <= 0 {
		return nil, 0, apperror.Validation("page must be positive", nil)
	}
	if f.PerPage <= 0 {
		return nil, 0, apperror.Validation("per_page must be positive", nil)
	}
	return s.posts.List(ctx, f)
}

// CreatePost creates a new post with optional category/tag associations.
func (s *Service) CreatePost(ctx context.Context, siteSlug, authorID string, req *CreatePostReq) (*model.Post, error) {
	// Generate or validate slug.
	slug := req.Slug
	if slug == "" {
		var err error
		slug, err = UniqueSlug(ctx, req.Title, "", s.posts.SlugExists)
		if err != nil {
			return nil, err
		}
	} else {
		exists, err := s.posts.SlugExists(ctx, slug, "")
		if err != nil {
			return nil, fmt.Errorf("check slug: %w", err)
		}
		if exists {
			return nil, apperror.Conflict("slug already exists", nil)
		}
	}

	// Determine status.
	status := model.PostStatusDraft
	if req.Status != "" {
		var ok bool
		status, ok = parseStatus(req.Status)
		if !ok {
			return nil, apperror.Validation("invalid status", nil)
		}
	}

	// Validate scheduled_at.
	if status == model.PostStatusScheduled {
		if req.ScheduledAt == nil {
			return nil, apperror.Validation("scheduled_at is required for scheduled status", nil)
		}
		if req.ScheduledAt.Before(time.Now()) {
			return nil, apperror.Validation("scheduled_at must be in the future", nil)
		}
	}

	post := &model.Post{
		AuthorID:     authorID,
		Title:        req.Title,
		Slug:         slug,
		Content:      req.Content,
		ContentJSON:  req.ContentJSON,
		Excerpt:      req.Excerpt,
		Status:       status,
		ScheduledAt:  req.ScheduledAt,
		CoverImageID: req.CoverImageID,
		MetaTitle:    req.MetaTitle,
		MetaDesc:     req.MetaDescription,
		OGImageURL:   req.OGImageURL,
		ExtraFields:  req.ExtraFields,
	}

	if status == model.PostStatusPublished {
		now := time.Now()
		post.PublishedAt = &now
	}

	if err := s.posts.Create(ctx, post); err != nil {
		return nil, err
	}

	// Sync categories and tags.
	if len(req.CategoryIDs) > 0 {
		if err := s.posts.SyncCategories(ctx, post.ID, req.CategoryIDs, req.PrimaryCategoryID); err != nil {
			return nil, fmt.Errorf("sync categories: %w", err)
		}
	}
	if len(req.TagIDs) > 0 {
		if err := s.posts.SyncTags(ctx, post.ID, req.TagIDs); err != nil {
			return nil, fmt.Errorf("sync tags: %w", err)
		}
	}

	// Create initial revision (version 1).
	rev := &model.PostRevision{
		PostID:      post.ID,
		EditorID:    authorID,
		Version:     1,
		Title:       post.Title,
		Content:     post.Content,
		ContentJSON: post.ContentJSON,
		DiffSummary: "Initial version",
	}
	if err := s.revs.Create(ctx, rev); err != nil {
		slog.Error("create initial revision failed", "error", err, "post_id", post.ID)
	}

	// Async: push to Meilisearch.
	go s.indexPost(context.Background(), siteSlug, post)

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionCreate,
		ResourceType: "post",
		ResourceID:   post.ID,
	}); err != nil {
		slog.Error("audit log post create failed", "error", err)
	}

	return post, nil
}

// GetPost returns a single post by ID.
func (s *Service) GetPost(ctx context.Context, id string) (*model.Post, error) {
	return s.posts.GetByID(ctx, id)
}

// UpdatePost updates a post with optimistic locking and creates a revision.
func (s *Service) UpdatePost(ctx context.Context, siteSlug, editorID, id string, req *UpdatePostReq) (*model.Post, error) {
	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldTitle := post.Title
	oldContent := post.Content
	oldExcerpt := post.Excerpt

	// Apply updates.
	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.Content = *req.Content
	}
	if req.ContentJSON != nil {
		post.ContentJSON = req.ContentJSON
	}
	if req.Excerpt != nil {
		post.Excerpt = *req.Excerpt
	}
	if req.MetaTitle != nil {
		post.MetaTitle = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		post.MetaDesc = *req.MetaDescription
	}
	if req.OGImageURL != nil {
		post.OGImageURL = *req.OGImageURL
	}
	if req.ExtraFields != nil {
		post.ExtraFields = req.ExtraFields
	}
	if req.CoverImageID != nil {
		post.CoverImageID = req.CoverImageID
	}
	if req.ScheduledAt != nil {
		post.ScheduledAt = req.ScheduledAt
	}

	// Handle slug change.
	if req.Slug != nil && *req.Slug != post.Slug {
		exists, err := s.posts.SlugExists(ctx, *req.Slug, id)
		if err != nil {
			return nil, fmt.Errorf("check slug: %w", err)
		}
		if exists {
			return nil, apperror.Conflict("slug already exists", nil)
		}
		post.Slug = *req.Slug
	}

	// Handle status transition.
	if req.Status != nil {
		newStatus, ok := parseStatus(*req.Status)
		if !ok {
			return nil, apperror.Validation("invalid status", nil)
		}
		if newStatus != post.Status {
			if err := validateTransition(post.Status, newStatus); err != nil {
				return nil, err
			}
			post.Status = newStatus
			if newStatus == model.PostStatusPublished && post.PublishedAt == nil {
				now := time.Now()
				post.PublishedAt = &now
			}
		}
	}

	// Optimistic lock update.
	expectedVersion := req.Version
	if err := s.posts.Update(ctx, post, expectedVersion); err != nil {
		return nil, err
	}

	// Sync categories/tags if provided.
	if req.CategoryIDs != nil {
		primaryID := ""
		if req.PrimaryCategoryID != nil {
			primaryID = *req.PrimaryCategoryID
		}
		if err := s.posts.SyncCategories(ctx, post.ID, req.CategoryIDs, primaryID); err != nil {
			return nil, fmt.Errorf("sync categories: %w", err)
		}
	}
	if req.TagIDs != nil {
		if err := s.posts.SyncTags(ctx, post.ID, req.TagIDs); err != nil {
			return nil, fmt.Errorf("sync tags: %w", err)
		}
	}

	// Create revision.
	diffSummary := buildDiffSummary(oldTitle, post.Title, oldContent, post.Content, oldExcerpt, post.Excerpt)
	rev := &model.PostRevision{
		PostID:      post.ID,
		EditorID:    editorID,
		Version:     post.Version, // version was incremented by BeforeAppendModel
		Title:       post.Title,
		Content:     post.Content,
		ContentJSON: post.ContentJSON,
		DiffSummary: diffSummary,
	}
	if err := s.revs.Create(ctx, rev); err != nil {
		slog.Error("create revision failed", "error", err, "post_id", post.ID)
	}

	// Async: update search index.
	go s.indexPost(context.Background(), siteSlug, post)

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "post",
		ResourceID:   post.ID,
	}); err != nil {
		slog.Error("audit log post update failed", "error", err)
	}

	return post, nil
}

// DeletePost soft-deletes a post.
func (s *Service) DeletePost(ctx context.Context, siteSlug, id string) error {
	if err := s.posts.SoftDelete(ctx, id); err != nil {
		return err
	}

	// Remove from search index.
	go func() {
		idx := fmt.Sprintf(postSearchIndex, siteSlug)
		if err := s.search.DeleteDocuments(context.Background(), idx, []string{id}); err != nil {
			slog.Error("remove post from search failed", "error", err, "post_id", id)
		}
	}()

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "post",
		ResourceID:   id,
	}); err != nil {
		slog.Error("audit log post delete failed", "error", err)
	}

	return nil
}

// indexPost pushes a post to Meilisearch.
func (s *Service) indexPost(ctx context.Context, siteSlug string, post *model.Post) {
	idx := fmt.Sprintf(postSearchIndex, siteSlug)
	doc := map[string]any{
		"id":      post.ID,
		"title":   post.Title,
		"excerpt": post.Excerpt,
		"slug":    post.Slug,
		"status":  statusString(post.Status),
	}
	if err := s.search.UpsertDocuments(ctx, idx, []map[string]any{doc}); err != nil {
		slog.Error("index post failed", "error", err, "post_id", post.ID)
	}
}

// buildDiffSummary generates a field-level change summary.
func buildDiffSummary(oldTitle, newTitle, oldContent, newContent, oldExcerpt, newExcerpt string) string {
	var changed []string
	if oldTitle != newTitle {
		changed = append(changed, "title")
	}
	if oldContent != newContent {
		changed = append(changed, "content")
	}
	if oldExcerpt != newExcerpt {
		changed = append(changed, "excerpt")
	}
	if len(changed) == 0 {
		return "No content changes"
	}
	return "Updated " + strings.Join(changed, ", ")
}
```

**Step 2: Create service_status.go**

```go
package post

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
)

// Allowed state transitions.
var allowedTransitions = map[model.PostStatus][]model.PostStatus{
	model.PostStatusDraft:     {model.PostStatusPublished, model.PostStatusScheduled},
	model.PostStatusScheduled: {model.PostStatusDraft, model.PostStatusPublished},
	model.PostStatusPublished: {model.PostStatusDraft, model.PostStatusArchived},
	model.PostStatusArchived:  {model.PostStatusPublished, model.PostStatusDraft},
}

// validateTransition checks if a status transition is allowed.
func validateTransition(from, to model.PostStatus) error {
	allowed, ok := allowedTransitions[from]
	if !ok {
		return apperror.Validation("invalid current status", nil)
	}
	for _, a := range allowed {
		if a == to {
			return nil
		}
	}

	// Build allowed list for error detail.
	var names []string
	for _, a := range allowed {
		names = append(names, statusString(a))
	}
	msg := fmt.Sprintf("cannot transition from %s to %s; allowed: %v", statusString(from), statusString(to), names)
	return apperror.Validation(msg, nil)
}

// Publish transitions a post to published status.
func (s *Service) Publish(ctx context.Context, siteSlug, id string) (*model.Post, error) {
	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(post.Status, model.PostStatusPublished); err != nil {
		return nil, err
	}

	if err := s.posts.UpdateStatus(ctx, id, model.PostStatusPublished); err != nil {
		return nil, err
	}

	// Async: update search index.
	post.Status = model.PostStatusPublished
	go s.indexPost(context.Background(), siteSlug, post)

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionPublish,
		ResourceType: "post",
		ResourceID:   id,
	}); err != nil {
		slog.Error("audit log post publish failed", "error", err)
	}

	// Return updated post.
	return s.posts.GetByID(ctx, id)
}

// Unpublish transitions a published post to archived.
func (s *Service) Unpublish(ctx context.Context, id string) (*model.Post, error) {
	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(post.Status, model.PostStatusArchived); err != nil {
		return nil, err
	}

	if err := s.posts.UpdateStatus(ctx, id, model.PostStatusArchived); err != nil {
		return nil, err
	}

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionUnpublish,
		ResourceType: "post",
		ResourceID:   id,
	}); err != nil {
		slog.Error("audit log post unpublish failed", "error", err)
	}

	return s.posts.GetByID(ctx, id)
}

// RevertToDraft transitions a post back to draft.
func (s *Service) RevertToDraft(ctx context.Context, id string) (*model.Post, error) {
	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(post.Status, model.PostStatusDraft); err != nil {
		return nil, err
	}

	if err := s.posts.UpdateStatus(ctx, id, model.PostStatusDraft); err != nil {
		return nil, err
	}

	return s.posts.GetByID(ctx, id)
}

// RestorePost restores a soft-deleted post to draft status.
func (s *Service) RestorePost(ctx context.Context, id string) (*model.Post, error) {
	if err := s.posts.Restore(ctx, id); err != nil {
		return nil, err
	}

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionRestore,
		ResourceType: "post",
		ResourceID:   id,
	}); err != nil {
		slog.Error("audit log post restore failed", "error", err)
	}

	return s.posts.GetByID(ctx, id)
}
```

**Step 3: Create service_test.go — write failing tests**

Create comprehensive service tests using the mock pattern from `internal/tag/service_test.go`. Include tests for:
- `TestCreatePost_Success` — happy path
- `TestCreatePost_DuplicateSlug` — conflict
- `TestCreatePost_ScheduledWithoutDate` — validation
- `TestUpdatePost_Success` — happy path with version check
- `TestUpdatePost_VersionConflict` — 409
- `TestDeletePost_Success` — soft delete
- `TestPublish_FromDraft` — valid transition
- `TestPublish_FromArchived` — valid transition
- `TestPublish_FromPublished` — invalid (already published)
- `TestUnpublish_FromPublished` — valid
- `TestUnpublish_FromDraft` — invalid transition
- `TestRevertToDraft_FromScheduled` — valid
- `TestRestorePost_Success` — restore from trash
- `TestValidateTransition_AllPaths` — table-driven test for all valid/invalid transitions
- `TestBuildDiffSummary` — diff summary generation

The test file follows the exact pattern of `internal/tag/service_test.go`:
1. Define `mockPostRepo`, `mockRevisionRepo` structs implementing the interfaces
2. Define `testEnv` struct with svc + all mocks
3. Each test configures mocks → calls service → asserts result

**Step 4: Run tests**

Run: `go test ./internal/post/... -v -run TestCreatePost`
Expected: PASS

Run: `go test ./internal/post/... -v -run TestValidateTransition`
Expected: PASS

Run: `go test ./internal/post/... -v`
Expected: All tests PASS

**Step 5: Commit**

```bash
git add internal/post/service.go internal/post/service_status.go internal/post/service_test.go
git commit -m "feat(post): implement CRUD and status transition service with tests"
```

---

### Task 5: Service — revisions

**Files:**
- Create: `internal/post/service_revision.go`
- Create: `internal/post/service_revision_test.go`

**Step 1: Create service_revision.go**

```go
package post

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
)

// ListRevisions returns the revision history for a post.
func (s *Service) ListRevisions(ctx context.Context, postID string) ([]model.PostRevision, error) {
	// Verify post exists.
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}
	return s.revs.List(ctx, postID)
}

// Rollback restores a post to a specific revision's content.
// Creates a new revision (version continues incrementing).
func (s *Service) Rollback(ctx context.Context, siteSlug, editorID, postID, revisionID string) (*model.Post, error) {
	// Get the target revision.
	rev, err := s.revs.GetByID(ctx, revisionID)
	if err != nil {
		return nil, err
	}

	// Verify the revision belongs to this post.
	if rev.PostID != postID {
		return nil, fmt.Errorf("revision does not belong to this post")
	}

	// Get current post.
	post, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Apply revision content to current post.
	post.Title = rev.Title
	post.Content = rev.Content
	post.ContentJSON = rev.ContentJSON

	// Save with current version (optimistic lock).
	currentVersion := post.Version
	if err := s.posts.Update(ctx, post, currentVersion); err != nil {
		return nil, err
	}

	// Create a new revision for the rollback.
	newRev := &model.PostRevision{
		PostID:      postID,
		EditorID:    editorID,
		Version:     post.Version, // incremented by Update
		Title:       post.Title,
		Content:     post.Content,
		ContentJSON: post.ContentJSON,
		DiffSummary: fmt.Sprintf("Rolled back to version %d", rev.Version),
	}
	if err := s.revs.Create(ctx, newRev); err != nil {
		slog.Error("create rollback revision failed", "error", err, "post_id", postID)
	}

	// Async: update search index.
	go s.indexPost(context.Background(), siteSlug, post)

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "post",
		ResourceID:   postID,
	}); err != nil {
		slog.Error("audit log post rollback failed", "error", err)
	}

	return post, nil
}
```

**Step 2: Create service_revision_test.go**

Tests:
- `TestListRevisions_Success` — happy path
- `TestListRevisions_PostNotFound` — 404
- `TestRollback_Success` — rollback creates new version
- `TestRollback_RevisionNotFound` — 404
- `TestRollback_WrongPost` — revision belongs to different post

**Step 3: Run tests**

Run: `go test ./internal/post/... -v -run TestListRevisions`
Expected: PASS

Run: `go test ./internal/post/... -v -run TestRollback`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/post/service_revision.go internal/post/service_revision_test.go
git commit -m "feat(post): implement revision management service with tests"
```

---

### Task 6: Service — translations

**Files:**
- Create: `internal/post/service_translation.go`
- Create: `internal/post/service_translation_test.go`

**Step 1: Create service_translation.go**

```go
package post

import (
	"context"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
)

// ListTranslations returns all translations for a post.
func (s *Service) ListTranslations(ctx context.Context, postID string) ([]model.PostTranslation, error) {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}
	return s.trans.List(ctx, postID)
}

// GetTranslation returns a specific locale translation.
func (s *Service) GetTranslation(ctx context.Context, postID, locale string) (*model.PostTranslation, error) {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}
	return s.trans.Get(ctx, postID, locale)
}

// UpsertTranslation creates or updates a translation for a post.
func (s *Service) UpsertTranslation(ctx context.Context, postID, locale string, req *UpsertTranslationReq) (*model.PostTranslation, error) {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}

	t := &model.PostTranslation{
		PostID:      postID,
		Locale:      locale,
		Title:       req.Title,
		Excerpt:     req.Excerpt,
		Content:     req.Content,
		ContentJSON: req.ContentJSON,
		MetaTitle:   req.MetaTitle,
		MetaDesc:    req.MetaDescription,
		OGImageURL:  req.OGImageURL,
	}

	if err := s.trans.Upsert(ctx, t); err != nil {
		return nil, err
	}

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "post_translation",
		ResourceID:   postID,
	}); err != nil {
		slog.Error("audit log translation upsert failed", "error", err)
	}

	// Return the saved translation.
	return s.trans.Get(ctx, postID, locale)
}

// DeleteTranslation removes a specific locale translation.
func (s *Service) DeleteTranslation(ctx context.Context, postID, locale string) error {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return err
	}

	if err := s.trans.Delete(ctx, postID, locale); err != nil {
		return err
	}

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "post_translation",
		ResourceID:   postID,
	}); err != nil {
		slog.Error("audit log translation delete failed", "error", err)
	}

	return nil
}
```

**Step 2: Create service_translation_test.go**

Tests:
- `TestListTranslations_Success`
- `TestGetTranslation_Success`
- `TestGetTranslation_NotFound`
- `TestUpsertTranslation_Create`
- `TestUpsertTranslation_PostNotFound`
- `TestDeleteTranslation_Success`
- `TestDeleteTranslation_NotFound`

**Step 3: Run tests and commit**

Run: `go test ./internal/post/... -v -run TestTranslation`
Expected: PASS

```bash
git add internal/post/service_translation.go internal/post/service_translation_test.go
git commit -m "feat(post): implement translation management service with tests"
```

---

### Task 7: Service — preview tokens

**Files:**
- Create: `internal/post/service_preview.go`
- Create: `internal/post/service_preview_test.go`

**Step 1: Create service_preview.go**

```go
package post

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/crypto"
)

const (
	previewTokenTTL   = 24 * time.Hour
	previewMaxPerPost = 5
	previewTokenBytes = 32
)

// CreatePreviewToken generates a new preview token for a post.
func (s *Service) CreatePreviewToken(ctx context.Context, postID, creatorID string) (*PreviewTokenResp, error) {
	// Verify post exists and is not deleted.
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}

	// Check active token count.
	count, err := s.preview.CountActive(ctx, postID)
	if err != nil {
		return nil, err
	}
	if count >= previewMaxPerPost {
		return nil, apperror.Validation("preview token limit reached (max 5 active tokens per post)", nil)
	}

	// Generate token.
	raw, hash, err := crypto.GenerateToken(previewTokenBytes)
	if err != nil {
		return nil, fmt.Errorf("generate preview token: %w", err)
	}

	tokenStr := "sky_preview_" + base64.RawURLEncoding.EncodeToString([]byte(raw))

	token := &model.PreviewToken{
		PostID:    postID,
		TokenHash: crypto.HashToken(tokenStr),
		ExpiresAt: time.Now().Add(previewTokenTTL),
		CreatedBy: &creatorID,
	}

	// We don't actually use `hash` since we hash the full tokenStr.
	_ = hash

	if err := s.preview.Create(ctx, token); err != nil {
		return nil, err
	}

	return &PreviewTokenResp{
		Token:       tokenStr,
		ID:          token.ID,
		ExpiresAt:   token.ExpiresAt,
		CreatedAt:   token.CreatedAt,
		ActiveCount: count + 1,
	}, nil
}

// ListPreviewTokens returns active (non-expired) tokens for a post.
func (s *Service) ListPreviewTokens(ctx context.Context, postID string) ([]model.PreviewToken, error) {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}
	return s.preview.List(ctx, postID)
}

// RevokeAllPreviewTokens deletes all preview tokens for a post.
func (s *Service) RevokeAllPreviewTokens(ctx context.Context, postID string) (int64, error) {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return 0, err
	}

	count, err := s.preview.DeleteAll(ctx, postID)
	if err != nil {
		return 0, err
	}

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "preview_token",
		ResourceID:   postID,
	}); err != nil {
		slog.Error("audit log revoke preview tokens failed", "error", err)
	}

	return count, nil
}

// RevokePreviewToken deletes a single preview token.
func (s *Service) RevokePreviewToken(ctx context.Context, tokenID string) error {
	return s.preview.DeleteByID(ctx, tokenID)
}

// GetPostByPreviewToken looks up a post by its preview token hash.
func (s *Service) GetPostByPreviewToken(ctx context.Context, rawToken string) (*model.Post, error) {
	hash := crypto.HashToken(rawToken)
	token, err := s.preview.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	// Get the post (including soft-deleted check is done by GetByIDUnscoped
	// since previews work for any state).
	post, err := s.posts.GetByIDUnscoped(ctx, token.PostID)
	if err != nil {
		return nil, err
	}

	// If the post is hard-deleted (shouldn't happen with soft deletes).
	if post.DeletedAt != nil {
		return nil, apperror.NotFound("post not found", nil)
	}

	return post, nil
}
```

**Step 2: Create service_preview_test.go**

Tests:
- `TestCreatePreviewToken_Success`
- `TestCreatePreviewToken_PostNotFound`
- `TestCreatePreviewToken_LimitReached`
- `TestListPreviewTokens_Success`
- `TestRevokeAllPreviewTokens_Success`
- `TestRevokePreviewToken_Success`
- `TestRevokePreviewToken_NotFound`
- `TestGetPostByPreviewToken_Success`
- `TestGetPostByPreviewToken_Expired`

**Step 3: Run tests and commit**

Run: `go test ./internal/post/... -v -run TestPreview`
Expected: PASS

```bash
git add internal/post/service_preview.go internal/post/service_preview_test.go
git commit -m "feat(post): implement preview token service with tests"
```

---

### Task 8: Handlers — all 19 endpoints

**Files:**
- Create: `internal/post/handler.go`
- Create: `internal/post/handler_crud.go`
- Create: `internal/post/handler_status.go`
- Create: `internal/post/handler_revision.go`
- Create: `internal/post/handler_translation.go`
- Create: `internal/post/handler_preview.go`

**Step 1: Create handler.go**

```go
package post

// Handler exposes post endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new post handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}
```

**Step 2: Create handler_crud.go**

5 handlers: `ListPosts`, `CreatePost`, `GetPost`, `UpdatePost`, `DeletePost`.

Each follows the pattern from `internal/category/handler.go`:
- Parse query params / bind JSON
- Call service method with `c.Request.Context()`
- Get `user_id` via `c.GetString("user_id")`
- Get `site_slug` via `c.GetString("site_slug")`
- Use `response.Paginated()`, `response.Created()`, `response.Success()`, `response.Error()`

**Step 3: Create handler_status.go**

4 handlers: `Publish`, `Unpublish`, `RevertToDraft`, `Restore`.

Each calls the corresponding service method and returns `response.Success()` with the updated post.

**Step 4: Create handler_revision.go**

2 handlers: `ListRevisions`, `Rollback`.

**Step 5: Create handler_translation.go**

4 handlers: `ListTranslations`, `GetTranslation`, `UpsertTranslation`, `DeleteTranslation`.

**Step 6: Create handler_preview.go**

4 handlers: `CreatePreviewToken`, `ListPreviewTokens`, `RevokeAllPreviewTokens`, `RevokePreviewToken`.

**Step 7: Verify compilation**

Run: `go build ./internal/post/...`
Expected: No errors

**Step 8: Commit**

```bash
git add internal/post/handler.go internal/post/handler_crud.go internal/post/handler_status.go internal/post/handler_revision.go internal/post/handler_translation.go internal/post/handler_preview.go
git commit -m "feat(post): add handler layer for all 19 endpoints"
```

---

### Task 9: Handler tests

**Files:**
- Create: `internal/post/handler_test.go`

**Step 1: Create handler_test.go**

Follow the pattern from `internal/category/handler_test.go` or `internal/tag/handler_test.go`. Use `httptest.NewRecorder` + `gin.CreateTestContext` for request simulation.

Test at minimum:
- `TestListPosts_Success` — 200 with pagination
- `TestCreatePost_Success` — 201
- `TestCreatePost_InvalidJSON` — 422
- `TestGetPost_Success` — 200
- `TestGetPost_NotFound` — 404
- `TestUpdatePost_Success` — 200
- `TestUpdatePost_VersionConflict` — 409
- `TestDeletePost_Success` — 200
- `TestPublish_Success` — 200
- `TestPublish_InvalidTransition` — 422
- `TestListRevisions_Success` — 200
- `TestRollback_Success` — 200
- `TestListTranslations_Success` — 200
- `TestUpsertTranslation_Success` — 200
- `TestDeleteTranslation_Success` — 200
- `TestCreatePreviewToken_Success` — 201
- `TestListPreviewTokens_Success` — 200
- `TestRevokeAllPreviewTokens_Success` — 200
- `TestRevokePreviewToken_Success` — 200

**Step 2: Run all tests**

Run: `go test ./internal/post/... -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add internal/post/handler_test.go
git commit -m "test(post): add handler layer tests for all 19 endpoints"
```

---

### Task 10: Router registration + API meta

**Files:**
- Modify: `internal/router/router.go`
- Modify: `internal/router/api_meta.go`

**Step 1: Add post module to router.go**

In `internal/router/router.go`, add import:
```go
"github.com/sky-flux/cms/internal/post"
```

After the Media section (line ~311), add:

```go
	// Posts
	postRepo := post.NewPostRepo(db)
	revRepo := post.NewRevisionRepo(db)
	transRepo := post.NewTranslationRepo(db)
	previewRepo := post.NewPreviewRepo(db)
	postSvc := post.NewService(postRepo, revRepo, transRepo, previewRepo, searchClient, auditSvc)
	postHandler := post.NewHandler(postSvc)

	// Post CRUD
	siteScoped.GET("/posts", postHandler.ListPosts)
	siteScoped.POST("/posts", postHandler.CreatePost)
	siteScoped.GET("/posts/:id", postHandler.GetPost)
	siteScoped.PUT("/posts/:id", postHandler.UpdatePost)
	siteScoped.DELETE("/posts/:id", postHandler.DeletePost)

	// Post status transitions
	siteScoped.POST("/posts/:id/publish", postHandler.Publish)
	siteScoped.POST("/posts/:id/unpublish", postHandler.Unpublish)
	siteScoped.POST("/posts/:id/revert-to-draft", postHandler.RevertToDraft)
	siteScoped.POST("/posts/:id/restore", postHandler.Restore)

	// Post revisions
	siteScoped.GET("/posts/:id/revisions", postHandler.ListRevisions)
	siteScoped.POST("/posts/:id/revisions/:rev_id/rollback", postHandler.Rollback)

	// Post translations
	siteScoped.GET("/posts/:id/translations", postHandler.ListTranslations)
	siteScoped.GET("/posts/:id/translations/:locale", postHandler.GetTranslation)
	siteScoped.PUT("/posts/:id/translations/:locale", postHandler.UpsertTranslation)
	siteScoped.DELETE("/posts/:id/translations/:locale", postHandler.DeleteTranslation)

	// Preview tokens
	siteScoped.POST("/posts/:id/preview", postHandler.CreatePreviewToken)
	siteScoped.GET("/posts/:id/preview", postHandler.ListPreviewTokens)
	siteScoped.DELETE("/posts/:id/preview", postHandler.RevokeAllPreviewTokens)
	siteScoped.DELETE("/posts/:id/preview/:token_id", postHandler.RevokePreviewToken)
```

**Step 2: Add entries to api_meta.go**

In `internal/router/api_meta.go`, add 19 entries after the Media section:

```go
		// Site-scoped: Posts CRUD
		"GET:/api/v1/site/posts":        {Name: "List posts", Description: "List posts with filters", Group: "posts"},
		"POST:/api/v1/site/posts":       {Name: "Create post", Description: "Create a new post", Group: "posts"},
		"GET:/api/v1/site/posts/:id":    {Name: "Get post", Description: "Get post details", Group: "posts"},
		"PUT:/api/v1/site/posts/:id":    {Name: "Update post", Description: "Update post with optimistic locking", Group: "posts"},
		"DELETE:/api/v1/site/posts/:id": {Name: "Delete post", Description: "Soft delete a post", Group: "posts"},

		// Site-scoped: Posts status
		"POST:/api/v1/site/posts/:id/publish":         {Name: "Publish post", Description: "Publish a post", Group: "posts"},
		"POST:/api/v1/site/posts/:id/unpublish":       {Name: "Unpublish post", Description: "Archive a published post", Group: "posts"},
		"POST:/api/v1/site/posts/:id/revert-to-draft": {Name: "Revert to draft", Description: "Revert post to draft", Group: "posts"},
		"POST:/api/v1/site/posts/:id/restore":         {Name: "Restore post", Description: "Restore from trash", Group: "posts"},

		// Site-scoped: Posts revisions
		"GET:/api/v1/site/posts/:id/revisions":                       {Name: "List revisions", Description: "List post revision history", Group: "posts"},
		"POST:/api/v1/site/posts/:id/revisions/:rev_id/rollback":     {Name: "Rollback revision", Description: "Rollback to a specific version", Group: "posts"},

		// Site-scoped: Posts translations
		"GET:/api/v1/site/posts/:id/translations":         {Name: "List translations", Description: "List post translations", Group: "posts"},
		"GET:/api/v1/site/posts/:id/translations/:locale": {Name: "Get translation", Description: "Get translation by locale", Group: "posts"},
		"PUT:/api/v1/site/posts/:id/translations/:locale": {Name: "Upsert translation", Description: "Create or update translation", Group: "posts"},
		"DELETE:/api/v1/site/posts/:id/translations/:locale": {Name: "Delete translation", Description: "Delete translation by locale", Group: "posts"},

		// Site-scoped: Preview tokens
		"POST:/api/v1/site/posts/:id/preview":           {Name: "Create preview", Description: "Generate preview token", Group: "posts"},
		"GET:/api/v1/site/posts/:id/preview":             {Name: "List previews", Description: "List active preview tokens", Group: "posts"},
		"DELETE:/api/v1/site/posts/:id/preview":          {Name: "Revoke all previews", Description: "Revoke all preview tokens", Group: "posts"},
		"DELETE:/api/v1/site/posts/:id/preview/:token_id": {Name: "Revoke preview", Description: "Revoke single preview token", Group: "posts"},
```

**Step 3: Verify compilation**

Run: `go build ./...`
Expected: No errors

**Step 4: Run full test suite**

Run: `go test ./... 2>&1 | tail -20`
Expected: All packages PASS

Run: `go vet ./...`
Expected: No warnings

**Step 5: Commit**

```bash
git add internal/router/router.go internal/router/api_meta.go
git commit -m "feat(router): register 19 post endpoints with API registry metadata"
```

---

### Task 11: Add DisplayName and AvatarURL to User model (if missing)

**Files:**
- Check: `internal/model/user.go`

**Step 1: Verify User model has DisplayName and AvatarURL fields**

The `PostResp.Author` uses `DisplayName` and `AvatarURL`. Check `internal/model/user.go` has these fields. If missing, add them.

**Step 2: Commit if changes made**

```bash
git add internal/model/user.go
git commit -m "fix(model): ensure User model has display_name and avatar_url fields"
```

---

### Task 12: Final verification + integration commit

**Step 1: Run go vet**

Run: `go vet ./...`
Expected: No warnings

**Step 2: Run all tests**

Run: `go test ./... -count=1 2>&1 | tail -30`
Expected: All packages PASS

**Step 3: Count total routes**

Expected total: 81 (existing) + 19 (posts) = **100 routes**

**Step 4: Count API meta entries**

Expected total: 64 (existing) + 19 (posts) = **83 RBAC-protected entries**

---

## Summary

| Task | Description | Files | Tests |
|------|-------------|-------|-------|
| 1 | ErrVersionConflict sentinel | 1 modified | existing pass |
| 2 | Interfaces + DTOs + Slug | 4 created | ~10 slug tests |
| 3 | Repository implementations | 4 created | compilation only |
| 4 | Service CRUD + status | 3 created | ~15 tests |
| 5 | Service revisions | 2 created | ~5 tests |
| 6 | Service translations | 2 created | ~7 tests |
| 7 | Service preview tokens | 2 created | ~9 tests |
| 8 | Handler layer (all 19) | 6 created | compilation only |
| 9 | Handler tests | 1 created | ~19 tests |
| 10 | Router + API meta | 2 modified | full suite pass |
| 11 | User model check | 0-1 modified | — |
| 12 | Final verification | — | full suite pass |

**Total new files**: ~22
**Total estimated tests**: ~65
**Total new endpoints**: 19

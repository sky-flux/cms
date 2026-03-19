package web_test

import (
	"context"
	"fmt"
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
	listLatest     func(ctx context.Context, page, size int) ([]templates.PostSummary, error)
	getBySlug      func(ctx context.Context, slug string) (*templates.PostDetail, error)
	listByCategory func(ctx context.Context, slug string, page, size int) ([]templates.PostSummary, error)
	listByTag      func(ctx context.Context, slug string, page, size int) ([]templates.PostSummary, error)
	search         func(ctx context.Context, query string, limit int) ([]templates.PostSummary, error)
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
	assert.Contains(t, body, "<!doctype html>") // full page (templ generates lowercase)
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
	assert.NotContains(t, body, "<!doctype html>") // partial only
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

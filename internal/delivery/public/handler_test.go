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

package delivery_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-flux/cms/internal/content/app"
	"github.com/sky-flux/cms/internal/content/delivery"
	"github.com/sky-flux/cms/internal/content/domain"
)

// --- stub use cases ---

type stubCreatePost struct {
	out *domain.Post
	err error
}

func (s *stubCreatePost) Execute(ctx context.Context, in app.CreatePostInput) (*domain.Post, error) {
	return s.out, s.err
}

type stubPublishPost struct {
	out *domain.Post
	err error
}

func (s *stubPublishPost) Execute(ctx context.Context, in app.PublishPostInput) (*domain.Post, error) {
	return s.out, s.err
}

type stubCreateCategory struct {
	out *domain.Category
	err error
}

func (s *stubCreateCategory) Execute(ctx context.Context, in app.CreateCategoryInput) (*domain.Category, error) {
	return s.out, s.err
}

func newContentAPI(t *testing.T, cp delivery.CreatePostExecutor, pp delivery.PublishPostExecutor, cc delivery.CreateCategoryExecutor) huma.API {
	t.Helper()
	_, api := humatest.New(t, huma.DefaultConfig("CMS API", "1.0.0"))
	delivery.RegisterRoutes(api, cp, pp, cc)
	return api
}

func TestCreatePostHandler_Success(t *testing.T) {
	created := &domain.Post{ID: "post-1", Title: "Hello", Slug: "hello", Status: domain.PostStatusDraft, Version: 1}
	api := newContentAPI(t,
		&stubCreatePost{out: created},
		&stubPublishPost{},
		&stubCreateCategory{},
	)

	body := `{"title":"Hello","slug":"hello","author_id":"author-1"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/posts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	assert.Equal(t, "post-1", out["id"])
}

func TestCreatePostHandler_ValidationError(t *testing.T) {
	api := newContentAPI(t,
		&stubCreatePost{err: domain.ErrEmptyTitle},
		&stubPublishPost{},
		&stubCreateCategory{},
	)

	body := `{"title":"","slug":"slug","author_id":"a"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/posts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnprocessableEntity, resp.Code)
}

func TestPublishPostHandler_Success(t *testing.T) {
	published := &domain.Post{ID: "post-1", Status: domain.PostStatusPublished, Version: 2}
	api := newContentAPI(t,
		&stubCreatePost{},
		&stubPublishPost{out: published},
		&stubCreateCategory{},
	)

	body := `{"expected_version":1}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/posts/post-1/publish", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestPublishPostHandler_VersionConflict(t *testing.T) {
	api := newContentAPI(t,
		&stubCreatePost{},
		&stubPublishPost{err: domain.ErrVersionConflict},
		&stubCreateCategory{},
	)

	body := `{"expected_version":1}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/posts/post-1/publish", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusConflict, resp.Code)
}

func TestCreateCategoryHandler_Success(t *testing.T) {
	cat := &domain.Category{ID: "cat-1", Name: "Tech", Slug: "tech"}
	api := newContentAPI(t,
		&stubCreatePost{},
		&stubPublishPost{},
		&stubCreateCategory{out: cat},
	)

	body := `{"name":"Tech","slug":"tech"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/categories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
}

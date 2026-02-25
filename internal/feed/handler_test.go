package feed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFeedRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/feed/rss.xml", h.RSSFeed)
	r.GET("/feed/atom.xml", h.AtomFeed)
	r.GET("/sitemap.xml", h.SitemapIndex)
	r.GET("/sitemap-posts.xml", h.SitemapPosts)
	r.GET("/sitemap-categories.xml", h.SitemapCategories)
	r.GET("/sitemap-tags.xml", h.SitemapTags)
	return r
}

func newTestHandler() *Handler {
	now := time.Now().UTC()
	pub := now.Add(-2 * 24 * time.Hour)
	posts := []model.Post{
		{
			ID:          "p1",
			Title:       "Test Post",
			Slug:        "test-post",
			Excerpt:     "A test post",
			Content:     "<p>Hello</p>",
			Status:      model.PostStatusPublished,
			PostType:    "article",
			PublishedAt: &pub,
			UpdatedAt:   pub,
			Author:      &model.User{DisplayName: "Author"},
		},
	}

	parentID := "c1"
	cats := []model.Category{
		{ID: "c1", Name: "Tech", Slug: "tech", ParentID: nil},
		{ID: "c2", Name: "Go", Slug: "go", ParentID: &parentID},
	}

	lastPost := now
	tags := []TagWithLastmod{
		{
			Tag:          model.Tag{ID: "t1", Name: "Golang", Slug: "golang"},
			LastPostDate: &lastPost,
			PostCount:    3,
		},
	}

	svc := NewService(
		&mockPostReader{posts: posts},
		&mockCategoryReader{cats: cats, latestPostDate: &now},
		&mockTagReader{tags: tags},
		&mockSiteConfig{},
	)
	return NewHandler(svc)
}

// mockPostReader, mockCategoryReader, mockTagReader, mockSiteConfig
// are defined in service_test.go (same package).

func TestHandler_RSSFeed_ContentType(t *testing.T) {
	h := newTestHandler()
	r := setupFeedRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/feed/rss.xml", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "rss+xml")
}

func TestHandler_AtomFeed_ContentType(t *testing.T) {
	h := newTestHandler()
	r := setupFeedRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/feed/atom.xml", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "atom+xml")
}

func TestHandler_SitemapIndex_ContentType(t *testing.T) {
	h := newTestHandler()
	r := setupFeedRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/sitemap.xml", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/xml")
}

func TestHandler_SitemapPosts(t *testing.T) {
	h := newTestHandler()
	r := setupFeedRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/sitemap-posts.xml", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Body.Bytes())
}

func TestHandler_SitemapCategories(t *testing.T) {
	h := newTestHandler()
	r := setupFeedRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/sitemap-categories.xml", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Body.Bytes())
}

func TestHandler_SitemapTags(t *testing.T) {
	h := newTestHandler()
	r := setupFeedRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/sitemap-tags.xml", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Body.Bytes())
}

func TestHandler_CacheHeaders(t *testing.T) {
	h := newTestHandler()
	r := setupFeedRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/feed/rss.xml", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Cache-Control"), "max-age=3600")
	assert.NotEmpty(t, w.Header().Get("ETag"), "ETag header should be present")
}

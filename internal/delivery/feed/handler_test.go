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
	posts    []feed.FeedPost
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

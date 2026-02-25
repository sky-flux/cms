package feed

import (
	"context"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mocks ---

type mockPostReader struct {
	posts  []model.Post
	latest *time.Time
	err    error
}

func (m *mockPostReader) ListPublished(_ context.Context, _ int, _, _ string) ([]model.Post, error) {
	return m.posts, m.err
}

func (m *mockPostReader) LatestPublishedAt(_ context.Context) (*time.Time, error) {
	return m.latest, m.err
}

type mockCategoryReader struct {
	cats           []model.Category
	latestPostDate *time.Time
	err            error
}

func (m *mockCategoryReader) ListAll(_ context.Context) ([]model.Category, error) {
	return m.cats, m.err
}

func (m *mockCategoryReader) LatestPostDate(_ context.Context, _ string) (*time.Time, error) {
	return m.latestPostDate, m.err
}

type mockTagReader struct {
	tags []TagWithLastmod
	err  error
}

func (m *mockTagReader) ListWithPosts(_ context.Context) ([]TagWithLastmod, error) {
	return m.tags, m.err
}

type mockSiteConfig struct{}

func (m *mockSiteConfig) GetSiteTitle(_ context.Context) string       { return "Test Blog" }
func (m *mockSiteConfig) GetSiteURL(_ context.Context) string         { return "https://example.com" }
func (m *mockSiteConfig) GetSiteDescription(_ context.Context) string { return "A test blog" }
func (m *mockSiteConfig) GetSiteLanguage(_ context.Context) string    { return "en" }

// --- helpers ---

func newTestService(posts *mockPostReader, cats *mockCategoryReader, tags *mockTagReader) *Service {
	return NewService(posts, cats, tags, &mockSiteConfig{})
}

func timePtr(t time.Time) *time.Time { return &t }

func makePosts(n int) []model.Post {
	now := time.Now().UTC()
	posts := make([]model.Post, n)
	for i := range n {
		pub := now.Add(-time.Duration(i) * 24 * time.Hour)
		posts[i] = model.Post{
			ID:          "p" + string(rune('1'+i)),
			Title:       "Post " + string(rune('1'+i)),
			Slug:        "post-" + string(rune('1'+i)),
			Excerpt:     "Excerpt for post",
			Content:     "<p>Hello</p>",
			Status:      model.PostStatusPublished,
			PostType:    "article",
			PublishedAt: &pub,
			UpdatedAt:   pub,
			Author:      &model.User{DisplayName: "Author"},
		}
	}
	return posts
}

// --- RSS tests ---

func TestGenerateRSS_ValidPosts(t *testing.T) {
	posts := makePosts(2)
	svc := newTestService(&mockPostReader{posts: posts}, nil, nil)

	data, err := svc.GenerateRSS(context.Background(), 20, "", "")
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var feed RSSFeed
	err = xml.Unmarshal(data, &feed)
	require.NoError(t, err)

	assert.Equal(t, "Test Blog", feed.Channel.Title)
	assert.Equal(t, "A test blog", feed.Channel.Description)
	assert.Equal(t, "en", feed.Channel.Language)
	assert.Len(t, feed.Channel.Items, 2)

	// Verify site link in raw XML (atom:link namespace causes unmarshal issues for <link>)
	assert.True(t, strings.Contains(string(data), "<link>https://example.com</link>"))
	assert.Equal(t, "Post 1", feed.Channel.Items[0].Title)
	assert.Equal(t, "Post 2", feed.Channel.Items[1].Title)
	assert.NotEmpty(t, feed.Channel.LastBuildDate)

	// Verify pubDate is RFC1123Z formatted
	for _, item := range feed.Channel.Items {
		assert.NotEmpty(t, item.PubDate)
		_, parseErr := time.Parse(time.RFC1123Z, item.PubDate)
		assert.NoError(t, parseErr, "pubDate should be RFC1123Z format")
	}
}

func TestGenerateRSS_EmptyPosts(t *testing.T) {
	svc := newTestService(&mockPostReader{posts: []model.Post{}}, nil, nil)

	data, err := svc.GenerateRSS(context.Background(), 20, "", "")
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var feed RSSFeed
	err = xml.Unmarshal(data, &feed)
	require.NoError(t, err)

	assert.Equal(t, "Test Blog", feed.Channel.Title)
	assert.Empty(t, feed.Channel.Items)
	assert.Empty(t, feed.Channel.LastBuildDate)
}

// --- Atom tests ---

func TestGenerateAtom_ValidPosts(t *testing.T) {
	posts := makePosts(2)
	svc := newTestService(&mockPostReader{posts: posts}, nil, nil)

	data, err := svc.GenerateAtom(context.Background(), 20, "", "")
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var feed AtomFeed
	err = xml.Unmarshal(data, &feed)
	require.NoError(t, err)

	assert.Equal(t, "Test Blog", feed.Title)
	assert.Equal(t, "https://example.com/", feed.ID)
	assert.Len(t, feed.Entries, 2)
	assert.Equal(t, "Post 1", feed.Entries[0].Title)
	assert.Equal(t, "Post 2", feed.Entries[1].Title)
	assert.NotEmpty(t, feed.Updated)
}

// --- Sitemap Index tests ---

func TestGenerateSitemapIndex(t *testing.T) {
	svc := newTestService(nil, nil, nil)

	data, err := svc.GenerateSitemapIndex(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var idx SitemapIndex
	err = xml.Unmarshal(data, &idx)
	require.NoError(t, err)

	assert.Len(t, idx.Sitemaps, 3)
	assert.Equal(t, "https://example.com/sitemap-posts.xml", idx.Sitemaps[0].Loc)
	assert.Equal(t, "https://example.com/sitemap-categories.xml", idx.Sitemaps[1].Loc)
	assert.Equal(t, "https://example.com/sitemap-tags.xml", idx.Sitemaps[2].Loc)

	for _, s := range idx.Sitemaps {
		assert.NotEmpty(t, s.Lastmod)
	}
}

// --- Posts Sitemap tests ---

func TestGeneratePostsSitemap_PriorityRules(t *testing.T) {
	now := time.Now().UTC()
	recentPub := now.Add(-24 * time.Hour) // 1 day old -> priority 0.9
	oldPub := now.Add(-100 * 24 * time.Hour) // 100 days old -> priority 0.5

	posts := []model.Post{
		{
			ID:          "p1",
			Title:       "Recent Post",
			Slug:        "recent-post",
			PostType:    "article",
			Status:      model.PostStatusPublished,
			PublishedAt: &recentPub,
			UpdatedAt:   recentPub,
		},
		{
			ID:          "p2",
			Title:       "Old Post",
			Slug:        "old-post",
			PostType:    "article",
			Status:      model.PostStatusPublished,
			PublishedAt: &oldPub,
			UpdatedAt:   oldPub,
		},
	}

	svc := newTestService(&mockPostReader{posts: posts}, nil, nil)

	data, err := svc.GeneratePostsSitemap(context.Background())
	require.NoError(t, err)

	var urlSet URLSet
	err = xml.Unmarshal(data, &urlSet)
	require.NoError(t, err)

	require.Len(t, urlSet.URLs, 2)
	assert.Equal(t, "0.9", urlSet.URLs[0].Priority, "recent post should have priority 0.9")
	assert.Equal(t, "daily", urlSet.URLs[0].Changefreq)
	assert.Equal(t, "0.5", urlSet.URLs[1].Priority, "old post should have priority 0.5")
	assert.Equal(t, "monthly", urlSet.URLs[1].Changefreq)
}

// --- Categories Sitemap tests ---

func TestGenerateCategoriesSitemap_RootVsChild(t *testing.T) {
	parentID := "c1"
	cats := []model.Category{
		{
			ID:       "c1",
			Name:     "Tech",
			Slug:     "tech",
			ParentID: nil, // root category
		},
		{
			ID:       "c2",
			Name:     "Go",
			Slug:     "go",
			ParentID: &parentID, // child category
		},
	}

	latestDate := time.Now().UTC()
	svc := newTestService(nil, &mockCategoryReader{cats: cats, latestPostDate: &latestDate}, nil)

	data, err := svc.GenerateCategoriesSitemap(context.Background())
	require.NoError(t, err)

	var urlSet URLSet
	err = xml.Unmarshal(data, &urlSet)
	require.NoError(t, err)

	require.Len(t, urlSet.URLs, 2)
	assert.Equal(t, "https://example.com/category/tech", urlSet.URLs[0].Loc)
	assert.Equal(t, "0.6", urlSet.URLs[0].Priority, "root category should have priority 0.6")
	assert.Equal(t, "https://example.com/category/go", urlSet.URLs[1].Loc)
	assert.Equal(t, "0.5", urlSet.URLs[1].Priority, "child category should have priority 0.5")
}

// --- Tags Sitemap tests ---

func TestGenerateTagsSitemap_SkipsEmptyTags(t *testing.T) {
	lastPost := time.Now().UTC()
	tags := []TagWithLastmod{
		{
			Tag:          model.Tag{ID: "t1", Name: "Golang", Slug: "golang"},
			LastPostDate: &lastPost,
			PostCount:    5,
		},
		// The second tag with 0 posts should NOT be returned by ListWithPosts
		// because the repo filters to tags that have posts.
		// But we test the service processes whatever the repo returns.
	}

	svc := newTestService(nil, nil, &mockTagReader{tags: tags})

	data, err := svc.GenerateTagsSitemap(context.Background())
	require.NoError(t, err)

	var urlSet URLSet
	err = xml.Unmarshal(data, &urlSet)
	require.NoError(t, err)

	require.Len(t, urlSet.URLs, 1)
	assert.Equal(t, "https://example.com/tag/golang", urlSet.URLs[0].Loc)
	assert.Equal(t, "0.4", urlSet.URLs[0].Priority)
	assert.NotEmpty(t, urlSet.URLs[0].Lastmod)
}

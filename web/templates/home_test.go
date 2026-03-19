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

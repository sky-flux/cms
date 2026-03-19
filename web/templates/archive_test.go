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

func TestCategoryArchivePage_RendersHeading(t *testing.T) {
	cfg := templates.SiteConfig{Name: "Blog"}
	posts := []templates.PostSummary{
		{Slug: "p1", Title: "Post 1", PublishedAt: time.Now()},
	}

	var buf bytes.Buffer
	err := templates.CategoryArchivePage(cfg, "Technology", "technology", posts, 1).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "Technology")
	assert.Contains(t, html, "Post 1")
	assert.NotContains(t, html, "load-more-btn") // fewer than 10 posts
}

func TestCategoryArchivePage_ShowsLoadMoreWhenFull(t *testing.T) {
	cfg := templates.SiteConfig{Name: "Blog"}
	posts := make([]templates.PostSummary, 10)
	for i := range posts {
		posts[i] = templates.PostSummary{Slug: "p", Title: "P", PublishedAt: time.Now()}
	}

	var buf bytes.Buffer
	err := templates.CategoryArchivePage(cfg, "Tech", "tech", posts, 1).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "load-more-btn")
	assert.Contains(t, html, `/categories/tech/partial?page=2`)
}

func TestTagArchivePage_RendersHeading(t *testing.T) {
	cfg := templates.SiteConfig{Name: "Blog"}
	posts := []templates.PostSummary{
		{Slug: "p1", Title: "Tagged Post", PublishedAt: time.Now()},
	}

	var buf bytes.Buffer
	err := templates.TagArchivePage(cfg, "Go", "go", posts, 1).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "Go")
	assert.Contains(t, html, "Tagged Post")
}

func TestPostsPartial_AppendsCardsAndOOBButton(t *testing.T) {
	posts := []templates.PostSummary{
		{Slug: "p1", Title: "Post 1", PublishedAt: time.Now()},
	}

	var buf bytes.Buffer
	err := templates.PostsPartial(posts, 3, "/categories/tech/partial?page=3").Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "Post 1")
	// With fewer than 10 posts, the OOB removes the button (empty span)
	assert.Contains(t, html, `hx-swap-oob="true"`)
}

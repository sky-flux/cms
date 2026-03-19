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

func TestSearchPage_RendersInputWithQuery(t *testing.T) {
	cfg := templates.SiteConfig{Name: "Blog"}
	results := []templates.PostSummary{
		{Slug: "found-post", Title: "Found Post", PublishedAt: time.Now()},
	}

	var buf bytes.Buffer
	err := templates.SearchPage(cfg, "golang", results).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, `value="golang"`)
	assert.Contains(t, html, `hx-get="/search"`)
	assert.Contains(t, html, `hx-push-url="true"`)
	assert.Contains(t, html, "Found Post")
	assert.Contains(t, html, `Search results for "golang"`)
}

func TestSearchResults_ShowsEmptyState(t *testing.T) {
	var buf bytes.Buffer
	err := templates.SearchResults("noresult", nil).Render(context.Background(), &buf)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), `No results for "noresult"`)
}

func TestSearchResults_ShowsCountAndCards(t *testing.T) {
	results := []templates.PostSummary{
		{Slug: "r1", Title: "Result 1", PublishedAt: time.Now()},
		{Slug: "r2", Title: "Result 2", PublishedAt: time.Now()},
	}

	var buf bytes.Buffer
	err := templates.SearchResults("go", results).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "2 result(s) found")
	assert.Contains(t, html, "Result 1")
	assert.Contains(t, html, "Result 2")
}

func TestSearchResults_EmptyQueryPrompt(t *testing.T) {
	var buf bytes.Buffer
	err := templates.SearchResults("", nil).Render(context.Background(), &buf)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Start typing to search posts")
}

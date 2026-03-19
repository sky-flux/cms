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

func TestPostPage_RendersTitle(t *testing.T) {
	cfg := templates.SiteConfig{Name: "My Blog"}
	post := templates.PostDetail{
		Slug:          "test-post",
		Title:         "Test Post",
		BodyHTML:      "<p>Hello world</p>",
		PublishedAt:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		AuthorName:    "Alice",
		AllowComments: true,
	}

	var buf bytes.Buffer
	err := templates.PostPage(cfg, post).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "Test Post")
	assert.Contains(t, html, "<p>Hello world</p>")
	assert.Contains(t, html, "Alice")
	assert.Contains(t, html, "March 1, 2026")
}

func TestPostPage_ShowsCommentFormWhenEnabled(t *testing.T) {
	cfg := templates.SiteConfig{Name: "Blog"}
	post := templates.PostDetail{
		Slug:          "commentable",
		Title:         "Commentable Post",
		PublishedAt:   time.Now(),
		AllowComments: true,
	}

	var buf bytes.Buffer
	err := templates.PostPage(cfg, post).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, `hx-post="/posts/commentable/comments"`)
	assert.Contains(t, html, "Leave a Comment")
}

func TestPostPage_HidesCommentFormWhenDisabled(t *testing.T) {
	cfg := templates.SiteConfig{Name: "Blog"}
	post := templates.PostDetail{
		Slug:          "no-comments",
		Title:         "No Comments Post",
		PublishedAt:   time.Now(),
		AllowComments: false,
	}

	var buf bytes.Buffer
	err := templates.PostPage(cfg, post).Render(context.Background(), &buf)
	require.NoError(t, err)

	assert.NotContains(t, buf.String(), "Leave a Comment")
}

func TestPostPage_RendersNestedComments(t *testing.T) {
	cfg := templates.SiteConfig{Name: "Blog"}
	post := templates.PostDetail{
		Slug:          "nested",
		Title:         "Post with Comments",
		PublishedAt:   time.Now(),
		AllowComments: true,
		Comments: []templates.Comment{
			{
				ID:         "1",
				AuthorName: "Alice",
				Body:       "Top level comment",
				CreatedAt:  time.Now(),
				Children: []templates.Comment{
					{
						ID:         "2",
						AuthorName: "Bob",
						Body:       "Nested reply",
						CreatedAt:  time.Now(),
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := templates.PostPage(cfg, post).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "Top level comment")
	assert.Contains(t, html, "Nested reply")
	assert.Contains(t, html, "Alice")
	assert.Contains(t, html, "Bob")
}

func TestPostPage_RendersBreadcrumbWithCategory(t *testing.T) {
	cfg := templates.SiteConfig{Name: "Blog"}
	post := templates.PostDetail{
		Slug:         "cat-post",
		Title:        "Post in Category",
		PublishedAt:  time.Now(),
		CategorySlug: "tech",
		CategoryName: "Technology",
	}

	var buf bytes.Buffer
	err := templates.PostPage(cfg, post).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "/categories/tech")
	assert.Contains(t, html, "Technology")
}

func TestCommentForm_HasHTMXAttributes(t *testing.T) {
	var buf bytes.Buffer
	err := templates.CommentForm("my-post").Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, `hx-post="/posts/my-post/comments"`)
	assert.Contains(t, html, `hx-target="#comment-form-status"`)
	assert.Contains(t, html, `hx-disabled-elt="this"`)
}

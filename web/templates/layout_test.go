package templates_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/sky-flux/cms/web/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLayout_RendersValidHTML(t *testing.T) {
	cfg := templates.SiteConfig{
		Name:        "Test Site",
		Description: "A test site",
		NavItems: []templates.NavItem{
			{Label: "Blog", URL: "/"},
			{Label: "About", URL: "/about"},
		},
	}

	var buf bytes.Buffer
	err := templates.Layout(cfg, "Home").Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	// Templ generates lowercase doctype
	assert.Contains(t, html, "<!doctype html>")
	assert.Contains(t, html, "<title>Home — Test Site</title>")
	assert.Contains(t, html, `src="/static/htmx.min.js"`)
	assert.Contains(t, html, `href="/static/app.css"`)
	assert.Contains(t, html, "hx-boost=\"true\"")
	assert.Contains(t, html, "Test Site")
	assert.Contains(t, html, "/about")
}

func TestNav_RendersNavItems(t *testing.T) {
	cfg := templates.SiteConfig{
		Name: "MySite",
		NavItems: []templates.NavItem{
			{Label: "Posts", URL: "/posts"},
			{Label: "Tags", URL: "/tags"},
		},
	}

	var buf bytes.Buffer
	err := templates.Nav(cfg).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "Posts")
	assert.Contains(t, html, "/posts")
	assert.Contains(t, html, "Tags")
	// Search input with HTMX attributes
	assert.Contains(t, html, `hx-get="/search"`)
	assert.Contains(t, html, `hx-trigger="input changed delay:300ms"`)
}

func TestFooter_RendersCopyright(t *testing.T) {
	cfg := templates.SiteConfig{Name: "MySite"}

	var buf bytes.Buffer
	err := templates.Footer(cfg).Render(context.Background(), &buf)
	require.NoError(t, err)

	assert.True(t, strings.Contains(buf.String(), "MySite"))
}

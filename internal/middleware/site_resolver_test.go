package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSiteLookup implements SiteLookup for testing.
type mockSiteLookup struct {
	slugs   map[string]string // slug -> id
	domains map[string]struct {
		slug string
		id   string
	}
}

func (m *mockSiteLookup) GetIDBySlug(_ context.Context, slug string) (string, error) {
	id, ok := m.slugs[slug]
	if !ok {
		return "", errors.New("not found")
	}
	return id, nil
}

func (m *mockSiteLookup) GetSlugByDomain(_ context.Context, domain string) (string, string, error) {
	entry, ok := m.domains[domain]
	if !ok {
		return "", "", errors.New("not found")
	}
	return entry.slug, entry.id, nil
}

func newMockLookup() *mockSiteLookup {
	return &mockSiteLookup{
		slugs: map[string]string{
			"my_blog": "site-uuid-001",
		},
		domains: map[string]struct {
			slug string
			id   string
		}{
			"blog.example.com": {slug: "my_blog", id: "site-uuid-001"},
		},
	}
}

func setupRouter(lookup SiteLookup) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SiteResolver(lookup))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"site_slug": c.GetString("site_slug"),
			"site_id":   c.GetString("site_id"),
		})
	})
	return r
}

func TestSiteResolver_ValidSlugHeader(t *testing.T) {
	r := setupRouter(newMockLookup())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Site-Slug", "my_blog")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "my_blog", body["site_slug"])
	assert.Equal(t, "site-uuid-001", body["site_id"])
}

func TestSiteResolver_InvalidSlugHeader(t *testing.T) {
	r := setupRouter(newMockLookup())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Site-Slug", "nonexistent")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, false, body["success"])
	assert.Equal(t, "site not found", body["error"])
}

func TestSiteResolver_MissingHeaderNoDomain(t *testing.T) {
	r := setupRouter(newMockLookup())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No X-Site-Slug header, Host will be empty or unresolvable
	req.Host = ""
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, false, body["success"])
	assert.Contains(t, body["error"], "missing X-Site-Slug header")
}

func TestSiteResolver_DomainResolution(t *testing.T) {
	r := setupRouter(newMockLookup())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No X-Site-Slug header, but Host maps to a known domain
	req.Host = "blog.example.com"
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "my_blog", body["site_slug"])
	assert.Equal(t, "site-uuid-001", body["site_id"])
}

func TestSiteResolver_UnknownDomainNoHeader(t *testing.T) {
	r := setupRouter(newMockLookup())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "unknown.example.com"
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

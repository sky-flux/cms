package middleware_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func newSchemaTestRouter(db *bun.DB, slugKey string, slugValue any) (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	if slugKey != "" {
		r.Use(func(c *gin.Context) {
			c.Set(slugKey, slugValue)
			c.Next()
		})
	}

	r.Use(middleware.Schema(db))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	return r, w
}

// makeDummyBunDB creates a bun.DB backed by a connector to a non-existent PG
// server. Good for testing error paths without a real database.
func makeDummyBunDB() *bun.DB {
	connector := pgdriver.NewConnector(
		pgdriver.WithDSN("postgres://invalid:invalid@localhost:0/invalid?sslmode=disable"),
	)
	sqldb := sql.OpenDB(connector)
	return bun.NewDB(sqldb, pgdialect.New())
}

func TestSchema_MissingSiteSlug_Returns500(t *testing.T) {
	db := makeDummyBunDB()
	r, w := newSchemaTestRouter(db, "", nil)
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "site context not set")
}

func TestSchema_EmptySlug_Returns500(t *testing.T) {
	db := makeDummyBunDB()
	r, w := newSchemaTestRouter(db, "site_slug", "")
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "site context not set")
}

func TestSchema_NonStringSlug_Returns500(t *testing.T) {
	db := makeDummyBunDB()
	r, w := newSchemaTestRouter(db, "site_slug", 12345)
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "site context not set")
}

func TestSchema_InvalidSlugFormat_Returns400(t *testing.T) {
	db := makeDummyBunDB()

	cases := []struct {
		name string
		slug string
	}{
		{"too short", "ab"},
		{"uppercase", "MyBlog"},
		{"special chars", "my-blog!"},
		{"spaces", "my blog"},
		{"sql injection attempt", "'; DROP TABLE users; --"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, w := newSchemaTestRouter(db, "site_slug", tc.slug)
			req, _ := http.NewRequest("GET", "/test", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "invalid site slug format")
		})
	}
}

func TestSchema_ValidSlug_DBFailure_Returns500(t *testing.T) {
	db := makeDummyBunDB()
	r, w := newSchemaTestRouter(db, "site_slug", "my_blog")
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to set schema context")
}

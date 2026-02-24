package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/middleware"
	"github.com/stretchr/testify/assert"
)

type mockRBACChecker struct {
	allowed bool
	err     error
}

func (m *mockRBACChecker) CheckPermission(_ context.Context, _, _, _ string) (bool, error) {
	return m.allowed, m.err
}

func setupRBACRouter(checker middleware.PermissionChecker) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RBAC(checker))
	r.GET("/api/v1/posts", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	return r
}

func TestRBAC_NoUserID_Returns401(t *testing.T) {
	r := setupRBACRouter(&mockRBACChecker{allowed: true})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/posts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRBAC_Allowed_Returns200(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/posts", nil)

	// Simulate auth middleware setting user_id
	gin.SetMode(gin.TestMode)
	r2 := gin.New()
	r2.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	r2.Use(middleware.RBAC(&mockRBACChecker{allowed: true}))
	r2.GET("/api/v1/posts", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	r2.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRBAC_Denied_Returns403(t *testing.T) {
	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-2")
		c.Next()
	})
	r.Use(middleware.RBAC(&mockRBACChecker{allowed: false}))
	r.GET("/api/v1/posts", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/posts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRBAC_CheckError_Returns500(t *testing.T) {
	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-3")
		c.Next()
	})
	r.Use(middleware.RBAC(&mockRBACChecker{err: assert.AnError}))
	r.GET("/api/v1/posts", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/posts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

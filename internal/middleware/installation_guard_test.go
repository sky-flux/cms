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

type mockInstallChecker struct {
	installed bool
}

func (m *mockInstallChecker) IsInstalled(ctx context.Context) bool {
	return m.installed
}

func (m *mockInstallChecker) MarkInstalled() {}

func setupGuardRouter(checker middleware.InstallChecker, exemptPaths ...string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.InstallationGuard(checker, exemptPaths...))
	r.GET("/api/v1/posts", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})
	r.POST("/api/v1/setup/check", func(c *gin.Context) {
		c.JSON(200, gin.H{"installed": false})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return r
}

func TestInstallGuard_Installed_PassThrough(t *testing.T) {
	r := setupGuardRouter(&mockInstallChecker{installed: true})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/posts", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestInstallGuard_NotInstalled_Returns503(t *testing.T) {
	r := setupGuardRouter(&mockInstallChecker{installed: false})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/posts", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 503, w.Code)
}

func TestInstallGuard_NotInstalled_SetupExempt(t *testing.T) {
	r := setupGuardRouter(&mockInstallChecker{installed: false}, "/api/v1/setup/")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/setup/check", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestInstallGuard_NotInstalled_HealthExempt(t *testing.T) {
	r := setupGuardRouter(&mockInstallChecker{installed: false}, "/health")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

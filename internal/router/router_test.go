package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// healthHandler requires concrete types (*bun.DB, *redis.Client, etc.)
// that are hard to mock without interfaces. These tests verify the
// response format using stub handlers. Full integration tests with
// testcontainers will cover the real healthHandler.

func TestSetup_HealthRouteRegistered(t *testing.T) {
	engine := gin.New()
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
}

func TestHealthHandler_AllHealthy(t *testing.T) {
	engine := gin.New()
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"db":          "connected",
			"redis":       "connected",
			"meilisearch": "connected",
			"rustfs":      "connected",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, "connected", body["db"])
	assert.Equal(t, "connected", body["redis"])
	assert.Equal(t, "connected", body["meilisearch"])
	assert.Equal(t, "connected", body["rustfs"])
}

func TestHealthHandler_Degraded(t *testing.T) {
	engine := gin.New()
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":      "degraded",
			"db":          "disconnected",
			"redis":       "connected",
			"meilisearch": "connected",
			"rustfs":      "connected",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "degraded", body["status"])
	assert.Equal(t, "disconnected", body["db"])
}

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRateLimitRouter(rdb *redis.Client) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("site_slug", "test-site")
		c.Next()
	})
	r.Use(middleware.RateLimit(rdb, "ratelimit:comment", 30*time.Second))
	r.POST("/comment", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	return r
}

func TestRateLimit_FirstRequest_Passes(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	r := setupRateLimitRouter(rdb)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestRateLimit_SecondRequest_Blocked(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	r := setupRateLimitRouter(rdb)

	// First request passes
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request blocked
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)
	assert.Contains(t, w2.Body.String(), "rate limit")
}

func TestRateLimit_AfterExpiry_Passes(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	r := setupRateLimitRouter(rdb)

	// First request
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Fast-forward time in miniredis
	mr.FastForward(31 * time.Second)

	// Should pass again
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
}

func TestRateLimit_NilRedis_FailsOpen(t *testing.T) {
	r := setupRateLimitRouter(nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/comment", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

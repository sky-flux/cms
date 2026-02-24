package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/middleware"
	"github.com/sky-flux/cms/internal/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const authTestSecret = "test-secret-key-at-least-32-bytes!"

func newAuthTestSetup(t *testing.T) (*jwt.Manager, *gin.Engine) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	mgr := jwt.NewManager(authTestSecret, 15*time.Minute, 5*time.Minute, rdb)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Auth(mgr))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"user_id": c.GetString("user_id")})
	})
	return mgr, r
}

func TestAuth_ValidToken_SetsUserID(t *testing.T) {
	mgr, r := newAuthTestSetup(t)
	token, _ := mgr.SignAccessToken("user-42")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "user-42")
}

func TestAuth_MissingHeader_Returns401(t *testing.T) {
	_, r := newAuthTestSetup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestAuth_InvalidToken_Returns401(t *testing.T) {
	_, r := newAuthTestSetup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestAuth_BlacklistedToken_Returns401(t *testing.T) {
	mgr, r := newAuthTestSetup(t)
	token, _ := mgr.SignAccessToken("user-99")
	claims, _ := mgr.Verify(token)
	mgr.Blacklist(context.Background(), claims.JTI, time.Hour)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestAuth_TempToken_SetsPurpose(t *testing.T) {
	mgr, _ := newAuthTestSetup(t)
	token, _ := mgr.SignTempToken("user-55", "2fa_verification")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Auth(mgr))
	r.POST("/2fa", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id": c.GetString("user_id"),
			"purpose": c.GetString("token_purpose"),
		})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/2fa", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "2fa_verification")
}

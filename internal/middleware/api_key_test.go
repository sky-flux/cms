package middleware_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/middleware"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/stretchr/testify/assert"
)

type mockAPIKeyLookup struct {
	key *model.APIKey
	err error
}

func (m *mockAPIKeyLookup) GetByHash(_ context.Context, _ string) (*model.APIKey, error) {
	return m.key, m.err
}

func hashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func setupAPIKeyRouter(lookup middleware.APIKeyLookup) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.APIKey(lookup))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"api_key_id": c.GetString("api_key_id")})
	})
	return r
}

func TestAPIKey_ValidKey_PassesThrough(t *testing.T) {
	raw := "test-api-key-123"
	lookup := &mockAPIKeyLookup{
		key: &model.APIKey{ID: "key-1", KeyHash: hashKey(raw), Status: model.APIKeyStatusActive},
	}
	r := setupAPIKeyRouter(lookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", raw)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "key-1")
}

func TestAPIKey_MissingHeader_Returns401(t *testing.T) {
	lookup := &mockAPIKeyLookup{}
	r := setupAPIKeyRouter(lookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "missing")
}

func TestAPIKey_InvalidKey_Returns401(t *testing.T) {
	lookup := &mockAPIKeyLookup{err: apperror.NotFound("not found", nil)}
	r := setupAPIKeyRouter(lookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "invalid")
}

func TestAPIKey_RevokedKey_Returns401(t *testing.T) {
	raw := "revoked-key"
	lookup := &mockAPIKeyLookup{
		key: &model.APIKey{ID: "key-2", KeyHash: hashKey(raw), Status: model.APIKeyStatusRevoked},
	}
	r := setupAPIKeyRouter(lookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", raw)
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "revoked")
}

func TestAPIKey_ExpiredKey_Returns401(t *testing.T) {
	raw := "expired-key"
	expired := time.Now().Add(-time.Hour)
	lookup := &mockAPIKeyLookup{
		key: &model.APIKey{ID: "key-3", KeyHash: hashKey(raw), Status: model.APIKeyStatusActive, ExpiresAt: &expired},
	}
	r := setupAPIKeyRouter(lookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", raw)
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "expired")
}

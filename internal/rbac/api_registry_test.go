package rbac_test

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock API Repo for Registry tests ---

type mockAPIRepo struct {
	upserted    []model.APIEndpoint
	disabledKey []string
}

func (m *mockAPIRepo) UpsertBatch(_ context.Context, endpoints []model.APIEndpoint) error {
	m.upserted = append(m.upserted, endpoints...)
	return nil
}
func (m *mockAPIRepo) DisableStale(_ context.Context, activeKeys []string) error {
	m.disabledKey = activeKeys
	return nil
}
func (m *mockAPIRepo) List(_ context.Context) ([]model.APIEndpoint, error)                    { return nil, nil }
func (m *mockAPIRepo) ListByGroup(_ context.Context, _ string) ([]model.APIEndpoint, error)   { return nil, nil }
func (m *mockAPIRepo) GetByMethodPath(_ context.Context, _, _ string) (*model.APIEndpoint, error) {
	return nil, nil
}

func TestRegistry_SyncRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// Register test routes
	engine.GET("/api/v1/posts", func(c *gin.Context) {})
	engine.POST("/api/v1/posts", func(c *gin.Context) {})
	engine.GET("/health", func(c *gin.Context) {}) // no metadata

	metaMap := map[string]rbac.APIMeta{
		"GET:/api/v1/posts":  {Name: "文章列表", Group: "内容管理"},
		"POST:/api/v1/posts": {Name: "创建文章", Group: "内容管理"},
		// /health intentionally omitted — should be skipped
	}

	repo := &mockAPIRepo{}
	registry := rbac.NewRegistry(repo)

	err := registry.SyncRoutes(context.Background(), engine, metaMap)
	require.NoError(t, err)

	// Should upsert 2 endpoints (not health)
	assert.Len(t, repo.upserted, 2)
	assert.Equal(t, "GET", repo.upserted[0].Method)
	assert.Equal(t, "/api/v1/posts", repo.upserted[0].Path)
	assert.Equal(t, "文章列表", repo.upserted[0].Name)

	// DisableStale should receive 2 active keys
	assert.Len(t, repo.disabledKey, 2)
	assert.Contains(t, repo.disabledKey, "GET:/api/v1/posts")
	assert.Contains(t, repo.disabledKey, "POST:/api/v1/posts")
}

func TestRegistry_SyncRoutes_NoMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/health", func(c *gin.Context) {})

	repo := &mockAPIRepo{}
	registry := rbac.NewRegistry(repo)

	err := registry.SyncRoutes(context.Background(), engine, map[string]rbac.APIMeta{})
	require.NoError(t, err)

	assert.Empty(t, repo.upserted, "routes without metadata should not be registered")
}

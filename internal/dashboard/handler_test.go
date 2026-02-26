package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubStats struct {
	stats *DashboardStats
	err   error
}

func (s *stubStats) GetStats(_ context.Context, _ string) (*DashboardStats, error) {
	return s.stats, s.err
}

func setupHandlerTest(reader StatsReader) (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	svc := NewService(reader)
	h := NewHandler(svc)
	r := gin.New()
	return r, h
}

func TestHandler_GetStats_Success(t *testing.T) {
	expected := &DashboardStats{
		Posts:    PostStats{Total: 100, Published: 80, Draft: 15, Scheduled: 5},
		Users:    UserStats{Total: 10, Active: 9, Inactive: 1},
		Comments: CommentStats{Total: 50, Pending: 5, Approved: 40, Spam: 5},
		Media:    MediaStats{Total: 200, StorageUsed: 104857600},
	}

	r, h := setupHandlerTest(&stubStats{stats: expected})
	r.GET("/stats", func(c *gin.Context) {
		c.Set("site_slug", "test-site")
		h.GetStats(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Success bool            `json:"success"`
		Data    DashboardStats  `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, int64(100), resp.Data.Posts.Total)
	assert.Equal(t, int64(80), resp.Data.Posts.Published)
	assert.Equal(t, int64(104857600), resp.Data.Media.StorageUsed)
}

func TestHandler_GetStats_NoSiteSlug(t *testing.T) {
	r, h := setupHandlerTest(&stubStats{stats: &DashboardStats{}})
	r.GET("/stats", h.GetStats) // no site_slug set

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	r.ServeHTTP(w, req)

	// Should return 400 because site_slug is required for site-scoped routes
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetStats_RepoError(t *testing.T) {
	r, h := setupHandlerTest(&stubStats{err: assert.AnError})
	r.GET("/stats", func(c *gin.Context) {
		c.Set("site_slug", "test-site")
		h.GetStats(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

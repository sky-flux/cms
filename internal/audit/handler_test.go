package audit_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/audit"
	"github.com/sky-flux/cms/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: AuditRepository
// ---------------------------------------------------------------------------

type mockAuditRepo struct {
	items []audit.AuditWithActor
	total int64
	err   error
}

func (m *mockAuditRepo) List(_ context.Context, _ audit.ListFilter) ([]audit.AuditWithActor, int64, error) {
	return m.items, m.total, m.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func setupRouter(repo audit.AuditRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := audit.NewHandler(repo)
	r := gin.New()
	r.GET("/audit-logs", h.ListAuditLogs)
	return r
}

func doGet(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func testAuditItem() audit.AuditWithActor {
	actorID := "actor-1"
	return audit.AuditWithActor{
		Audit: model.Audit{
			ID:           "audit-1",
			ActorID:      &actorID,
			ActorEmail:   "admin@example.com",
			Action:       model.LogActionCreate,
			ResourceType: "post",
			ResourceID:   "post-1",
			IPAddress:    "127.0.0.1",
			CreatedAt:    time.Now(),
		},
		ActorDisplayName: "Admin User",
	}
}

// ---------------------------------------------------------------------------
// Tests: Default pagination
// ---------------------------------------------------------------------------

func TestHandler_ListAuditLogs_DefaultPagination(t *testing.T) {
	repo := &mockAuditRepo{
		items: []audit.AuditWithActor{testAuditItem()},
		total: 1,
	}
	r := setupRouter(repo)

	w := doGet(r, "/audit-logs")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))

	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(1), meta["page"])
	assert.Equal(t, float64(20), meta["per_page"])
	assert.Equal(t, float64(1), meta["total"])
}

// ---------------------------------------------------------------------------
// Tests: With filters
// ---------------------------------------------------------------------------

func TestHandler_ListAuditLogs_WithFilters(t *testing.T) {
	repo := &mockAuditRepo{
		items: []audit.AuditWithActor{testAuditItem()},
		total: 1,
	}
	r := setupRouter(repo)

	w := doGet(r, "/audit-logs?page=2&per_page=5&actor_id=actor-1&action=1&resource_type=post&start_date=2026-01-01T00:00:00Z&end_date=2026-12-31T23:59:59Z")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))

	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(2), meta["page"])
	assert.Equal(t, float64(5), meta["per_page"])
}

// ---------------------------------------------------------------------------
// Tests: Invalid action returns 422
// ---------------------------------------------------------------------------

func TestHandler_ListAuditLogs_InvalidAction(t *testing.T) {
	repo := &mockAuditRepo{}
	r := setupRouter(repo)

	w := doGet(r, "/audit-logs?action=99")
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_ListAuditLogs_ActionNotNumber(t *testing.T) {
	repo := &mockAuditRepo{}
	r := setupRouter(repo)

	w := doGet(r, "/audit-logs?action=abc")
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Invalid date returns 422
// ---------------------------------------------------------------------------

func TestHandler_ListAuditLogs_InvalidStartDate(t *testing.T) {
	repo := &mockAuditRepo{}
	r := setupRouter(repo)

	w := doGet(r, "/audit-logs?start_date=not-a-date")
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_ListAuditLogs_InvalidEndDate(t *testing.T) {
	repo := &mockAuditRepo{}
	r := setupRouter(repo)

	w := doGet(r, "/audit-logs?end_date=not-a-date")
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PerPage cap at 100
// ---------------------------------------------------------------------------

func TestHandler_ListAuditLogs_PerPageCapped(t *testing.T) {
	repo := &mockAuditRepo{
		items: nil,
		total: 0,
	}
	r := setupRouter(repo)

	w := doGet(r, "/audit-logs?per_page=999")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(100), meta["per_page"])
}

// ---------------------------------------------------------------------------
// Tests: Repo error
// ---------------------------------------------------------------------------

func TestHandler_ListAuditLogs_RepoError(t *testing.T) {
	repo := &mockAuditRepo{
		err: errors.New("db connection lost"),
	}
	r := setupRouter(repo)

	w := doGet(r, "/audit-logs")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Empty results
// ---------------------------------------------------------------------------

func TestHandler_ListAuditLogs_EmptyResults(t *testing.T) {
	repo := &mockAuditRepo{
		items: []audit.AuditWithActor{},
		total: 0,
	}
	r := setupRouter(repo)

	w := doGet(r, "/audit-logs")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 0)
}

// ---------------------------------------------------------------------------
// Tests: Response structure
// ---------------------------------------------------------------------------

func TestHandler_ListAuditLogs_ResponseStructure(t *testing.T) {
	repo := &mockAuditRepo{
		items: []audit.AuditWithActor{testAuditItem()},
		total: 1,
	}
	r := setupRouter(repo)

	w := doGet(r, "/audit-logs")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	data := resp["data"].([]interface{})
	require.Len(t, data, 1)

	item := data[0].(map[string]interface{})
	assert.Equal(t, "audit-1", item["id"])
	assert.Equal(t, "actor-1", item["actor_id"])
	assert.Equal(t, "admin@example.com", item["actor_email"])
	assert.Equal(t, "Admin User", item["actor_display_name"])
	assert.Equal(t, float64(1), item["action"])
	assert.Equal(t, "post", item["resource_type"])
	assert.Equal(t, "post-1", item["resource_id"])
	assert.Equal(t, "127.0.0.1", item["ip_address"])
	assert.NotEmpty(t, item["created_at"])
}

package delivery_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/sky-flux/cms/internal/platform/delivery"
	"github.com/sky-flux/cms/internal/platform/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAuditLister struct {
	entries []domain.AuditEntry
	total   int64
}

func (s *stubAuditLister) List(_ context.Context, _ domain.AuditFilter) ([]domain.AuditEntry, int64, error) {
	return s.entries, s.total, nil
}

func newAuditAPI(t *testing.T, lister delivery.AuditLister) huma.API {
	t.Helper()
	_, api := humatest.New(t, huma.DefaultConfig("Test API", "1.0.0"))
	h := delivery.NewAuditHandler(lister)
	delivery.RegisterAuditRoutes(api, h)
	return api
}

func TestListAudit_Returns200WithItems(t *testing.T) {
	now := time.Now()
	lister := &stubAuditLister{
		entries: []domain.AuditEntry{
			{ID: "entry-1", UserID: "user-1", Action: domain.AuditActionCreate, Resource: "post", CreatedAt: now},
		},
		total: 1,
	}
	api := newAuditAPI(t, lister)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit", nil)
	rec := httptest.NewRecorder()
	api.Adapter().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	items, ok := body["items"].([]any)
	require.True(t, ok)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(1), int64(body["total"].(float64)))
}

func TestListAudit_EmptyResult_Returns200(t *testing.T) {
	api := newAuditAPI(t, &stubAuditLister{entries: nil, total: 0})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit", nil)
	rec := httptest.NewRecorder()
	api.Adapter().ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestListAudit_WithFilters_Returns200(t *testing.T) {
	api := newAuditAPI(t, &stubAuditLister{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit?page=2&per_page=10&resource=post", nil)
	rec := httptest.NewRecorder()
	api.Adapter().ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

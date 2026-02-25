package redirect_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/cache"
	"github.com/sky-flux/cms/internal/redirect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Handler test helpers
// ---------------------------------------------------------------------------

func setupHandlerRouter() (*gin.Engine, *mockRepo) {
	gin.SetMode(gin.TestMode)
	repo := &mockRepo{}
	a := &audit.NoopLogger{}
	cc := cache.NewClient(nil)
	svc := redirect.NewService(repo, a, cc)
	h := redirect.NewHandler(svc)

	r := gin.New()
	g := r.Group("/api/v1/redirects")

	// Static paths BEFORE :id
	g.DELETE("/batch", h.BatchDelete)
	g.POST("/import", h.Import)
	g.GET("/export", h.Export)

	g.GET("", h.List)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)

	return r, repo
}

func doRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// Tests: List
// ---------------------------------------------------------------------------

func TestHandler_List(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.listRedirects = []model.Redirect{
		{
			ID:         "rd-1",
			SourcePath: "/old",
			TargetURL:  "https://new.com",
			StatusCode: 301,
			Status:     model.RedirectStatusActive,
		},
	}
	repo.listTotal = 1

	w := doRequest(r, http.MethodGet, "/api/v1/redirects?page=1&per_page=20", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestHandler_Create(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.sourcePathExists = false

	w := doRequest(r, http.MethodPost, "/api/v1/redirects", map[string]interface{}{
		"source_path": "/old-page",
		"target_url":  "https://example.com/new-page",
		"status_code": 301,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_Create_ValidationError(t *testing.T) {
	r, _ := setupHandlerRouter()

	// Missing required fields.
	w := doRequest(r, http.MethodPost, "/api/v1/redirects", map[string]interface{}{})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestHandler_Update(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.getByID = &model.Redirect{
		ID:         "rd-1",
		SourcePath: "/old",
		TargetURL:  "https://old.com",
		StatusCode: 301,
		Status:     model.RedirectStatusActive,
	}

	newURL := "https://new.com"
	w := doRequest(r, http.MethodPut, "/api/v1/redirects/rd-1", map[string]interface{}{
		"target_url": newURL,
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Delete
// ---------------------------------------------------------------------------

func TestHandler_Delete(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.getByID = &model.Redirect{
		ID:         "rd-1",
		SourcePath: "/old",
		TargetURL:  "https://old.com",
		StatusCode: 301,
		Status:     model.RedirectStatusActive,
	}

	w := doRequest(r, http.MethodDelete, "/api/v1/redirects/rd-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Delete_NotFound(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.getByIDErr = apperror.NotFound("redirect not found", nil)

	w := doRequest(r, http.MethodDelete, "/api/v1/redirects/nonexistent", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: BatchDelete
// ---------------------------------------------------------------------------

func TestHandler_BatchDelete(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.batchDeleteCount = 2

	w := doRequest(r, http.MethodDelete, "/api/v1/redirects/batch", map[string]interface{}{
		"ids": []string{"rd-1", "rd-2"},
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_BatchDelete_ValidationError(t *testing.T) {
	r, _ := setupHandlerRouter()

	// Empty ids.
	w := doRequest(r, http.MethodDelete, "/api/v1/redirects/batch", map[string]interface{}{
		"ids": []string{},
	})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Import
// ---------------------------------------------------------------------------

func TestHandler_Import(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.sourcePathExists = false

	// Build multipart form with CSV file.
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "redirects.csv")
	require.NoError(t, err)
	_, _ = fw.Write([]byte("source_path,target_url,status_code\n/old,https://new.com,301\n"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/redirects/import", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

// ---------------------------------------------------------------------------
// Tests: Export
// ---------------------------------------------------------------------------

func TestHandler_Export(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.listAllRedirects = []model.Redirect{
		{ID: "rd-1", SourcePath: "/old", TargetURL: "https://new.com", StatusCode: 301},
		{ID: "rd-2", SourcePath: "/foo", TargetURL: "https://bar.com", StatusCode: 302},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/redirects/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
	assert.Contains(t, w.Header().Get("Content-Disposition"), "redirects.csv")

	// Parse CSV response.
	reader := csv.NewReader(w.Body)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	assert.Len(t, records, 3) // header + 2 rows
	assert.Equal(t, "source_path", records[0][0])
	assert.Equal(t, "/old", records[1][0])
	assert.Equal(t, "/foo", records[2][0])
}


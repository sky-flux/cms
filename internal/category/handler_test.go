package category_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/category"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHandlerRouter() (*gin.Engine, *mockCategoryRepo) {
	gin.SetMode(gin.TestMode)
	repo := newMockRepo()
	svc := category.NewService(repo, cache.NewClient(nil), audit.NewNoopLogger())
	h := category.NewHandler(svc)

	r := gin.New()
	g := r.Group("/api/v1/categories")
	g.GET("", h.List)
	g.POST("", h.Create)
	g.PUT("/reorder", h.Reorder)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)

	return r, repo
}

func doHandlerRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
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

func TestHandler_List_Success(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.listCats = []model.Category{
		{ID: "cat-1", Name: "Tech", Slug: "tech", Path: "/tech/"},
	}

	w := doHandlerRequest(r, http.MethodGet, "/api/v1/categories", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

// ---------------------------------------------------------------------------
// Tests: Get
// ---------------------------------------------------------------------------

func TestHandler_Get_Success(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.getByIDMap["cat-1"] = &model.Category{
		ID: "cat-1", Name: "Tech", Slug: "tech", Path: "/tech/",
	}

	w := doHandlerRequest(r, http.MethodGet, "/api/v1/categories/cat-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_Get_NotFound(t *testing.T) {
	r, _ := setupHandlerRouter()

	w := doHandlerRequest(r, http.MethodGet, "/api/v1/categories/nonexistent", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestHandler_Create_Success(t *testing.T) {
	r, _ := setupHandlerRouter()

	w := doHandlerRequest(r, http.MethodPost, "/api/v1/categories", map[string]string{
		"name": "Tech", "slug": "tech",
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_Create_ValidationError(t *testing.T) {
	r, _ := setupHandlerRouter()

	// Missing required fields.
	w := doHandlerRequest(r, http.MethodPost, "/api/v1/categories", map[string]string{})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestHandler_Update_Success(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.getByIDMap["cat-1"] = &model.Category{
		ID: "cat-1", Name: "Old", Slug: "old", Path: "/old/",
	}

	w := doHandlerRequest(r, http.MethodPut, "/api/v1/categories/cat-1", map[string]string{
		"name": "New",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Delete
// ---------------------------------------------------------------------------

func TestHandler_Delete_Success(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.getByIDMap["cat-1"] = &model.Category{
		ID: "cat-1", Name: "Leaf", Slug: "leaf", Path: "/leaf/",
	}

	w := doHandlerRequest(r, http.MethodDelete, "/api/v1/categories/cat-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Delete_Conflict(t *testing.T) {
	r, repo := setupHandlerRouter()
	repo.getByIDMap["cat-1"] = &model.Category{
		ID: "cat-1", Name: "Parent", Slug: "parent", Path: "/parent/",
	}
	repo.childrenMap["cat-1"] = []model.Category{
		{ID: "cat-2", Name: "Child", Slug: "child", Path: "/parent/child/"},
	}

	w := doHandlerRequest(r, http.MethodDelete, "/api/v1/categories/cat-1", nil)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Reorder
// ---------------------------------------------------------------------------

func TestHandler_Reorder_Success(t *testing.T) {
	r, _ := setupHandlerRouter()

	w := doHandlerRequest(r, http.MethodPut, "/api/v1/categories/reorder", map[string]interface{}{
		"orders": []map[string]interface{}{
			{"id": "cat-1", "sort_order": 2},
			{"id": "cat-2", "sort_order": 1},
		},
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Reorder_ValidationError(t *testing.T) {
	r, _ := setupHandlerRouter()

	// Empty orders.
	w := doHandlerRequest(r, http.MethodPut, "/api/v1/categories/reorder", map[string]interface{}{
		"orders": []map[string]interface{}{},
	})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

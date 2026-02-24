package tag_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/tag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, *testEnv) {
	gin.SetMode(gin.TestMode)
	env := newTestEnv()
	h := tag.NewHandler(env.svc)

	r := gin.New()

	// Inject site_slug into context.
	r.Use(func(c *gin.Context) {
		c.Set("site_slug", "test-site")
		c.Next()
	})

	g := r.Group("/api/v1/tags")
	g.GET("", h.List)
	g.GET("/suggest", h.Suggest)
	g.GET("/:id", h.Get)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)

	return r, env
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

func TestHandler_List_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.listTags = []model.Tag{*testTag()}
	env.repo.listTotal = 1
	env.repo.postCount = 2

	w := doRequest(r, http.MethodGet, "/api/v1/tags?page=1&per_page=10", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["meta"])

	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(1), meta["total"])
	assert.Equal(t, float64(1), meta["page"])
	assert.Equal(t, float64(10), meta["per_page"])
}

// ---------------------------------------------------------------------------
// Tests: Get
// ---------------------------------------------------------------------------

func TestHandler_Get_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.getByID = testTag()
	env.repo.postCount = 3

	w := doRequest(r, http.MethodGet, "/api/v1/tags/tag-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Go", data["name"])
}

func TestHandler_Get_NotFound(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.getByIDErr = apperror.NotFound("tag not found", nil)

	w := doRequest(r, http.MethodGet, "/api/v1/tags/nope", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Suggest
// ---------------------------------------------------------------------------

func TestHandler_Suggest_Success(t *testing.T) {
	r, _ := setupTestRouter()

	w := doRequest(r, http.MethodGet, "/api/v1/tags/suggest?q=go", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestHandler_Create_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.nameExists = false
	env.repo.slugExists = false

	w := doRequest(r, http.MethodPost, "/api/v1/tags", map[string]string{
		"name": "Rust",
		"slug": "rust",
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_Create_ValidationError(t *testing.T) {
	r, _ := setupTestRouter()

	// Missing required "name" and "slug" fields.
	w := doRequest(r, http.MethodPost, "/api/v1/tags", map[string]string{})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_Create_MissingSlug(t *testing.T) {
	r, _ := setupTestRouter()

	w := doRequest(r, http.MethodPost, "/api/v1/tags", map[string]string{
		"name": "Rust",
	})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestHandler_Update_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.getByID = testTag()
	env.repo.nameExists = false

	w := doRequest(r, http.MethodPut, "/api/v1/tags/tag-1", map[string]string{
		"name": "GoLang",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Delete
// ---------------------------------------------------------------------------

func TestHandler_Delete_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.getByID = testTag()

	w := doRequest(r, http.MethodDelete, "/api/v1/tags/tag-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_Delete_NotFound(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.getByIDErr = apperror.NotFound("tag not found", nil)

	w := doRequest(r, http.MethodDelete, "/api/v1/tags/nonexistent", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

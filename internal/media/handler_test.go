package media_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/media"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, *testEnv) {
	gin.SetMode(gin.TestMode)
	env := newTestEnv()
	h := media.NewHandler(env.svc)

	r := gin.New()

	// Inject context values.
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Set("site_slug", "my-site")
		c.Next()
	})

	g := r.Group("/api/v1/media")
	g.GET("", h.List)
	g.POST("", h.Upload)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	g.DELETE("/batch", h.BatchDelete)

	return r, env
}

func doJSONRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
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
	env.repo.listFiles = []model.MediaFile{*testMediaFile()}
	env.repo.listTotal = 1

	w := doJSONRequest(r, http.MethodGet, "/api/v1/media?page=1&per_page=10", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["meta"])
}

// ---------------------------------------------------------------------------
// Tests: Get
// ---------------------------------------------------------------------------

func TestHandler_Get_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.getByID = testMediaFile()

	w := doJSONRequest(r, http.MethodGet, "/api/v1/media/mf-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_Get_NotFound(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.getByIDErr = apperror.NotFound("media file not found", nil)

	w := doJSONRequest(r, http.MethodGet, "/api/v1/media/nonexistent", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Upload
// ---------------------------------------------------------------------------

func TestHandler_Upload_Success(t *testing.T) {
	r, env := setupTestRouter()
	_ = env

	// Build multipart form.
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="photo.jpg"`)
	partHeader.Set("Content-Type", "image/jpeg")
	fw, err := mw.CreatePart(partHeader)
	require.NoError(t, err)
	_, err = io.WriteString(fw, "fake-image-data")
	require.NoError(t, err)

	err = mw.WriteField("alt_text", "A nice photo")
	require.NoError(t, err)

	err = mw.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/media", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_Upload_NoFile(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/media", nil)
	req.Header.Set("Content-Type", "multipart/form-data")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestHandler_Update_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.getByID = testMediaFile()

	w := doJSONRequest(r, http.MethodPut, "/api/v1/media/mf-1", map[string]string{
		"alt_text": "Updated alt text",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Delete
// ---------------------------------------------------------------------------

func TestHandler_Delete_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.getByID = testMediaFile()
	env.repo.referencingPosts = nil

	w := doJSONRequest(r, http.MethodDelete, "/api/v1/media/mf-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Delete_Conflict(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.getByID = testMediaFile()
	env.repo.referencingPosts = []media.PostRef{
		{ID: "post-1", Title: "My Post"},
	}

	w := doJSONRequest(r, http.MethodDelete, "/api/v1/media/mf-1", nil)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: BatchDelete
// ---------------------------------------------------------------------------

func TestHandler_BatchDelete_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.repo.batchRefCounts = map[string]int64{}
	env.repo.batchDeleteCount = 2

	w := doJSONRequest(r, http.MethodDelete, "/api/v1/media/batch", media.BatchDeleteReq{
		IDs: []string{"mf-1", "mf-2"},
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_BatchDelete_ValidationError(t *testing.T) {
	r, _ := setupTestRouter()

	// Empty IDs should fail validation.
	w := doJSONRequest(r, http.MethodDelete, "/api/v1/media/batch", map[string]interface{}{
		"ids": []string{},
	})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

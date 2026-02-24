package post_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/post"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupHandlerRouter creates a test gin engine with all 19 post routes.
// It injects site_slug and user_id via middleware, simulating real middleware.
func setupHandlerRouter(env *testEnv) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := post.NewHandler(env.svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("site_slug", "test-site")
		c.Set("user_id", "user-1")
		c.Next()
	})

	g := r.Group("/api/v1/site")

	// CRUD
	g.GET("/posts", h.ListPosts)
	g.POST("/posts", h.CreatePost)
	g.GET("/posts/:id", h.GetPost)
	g.PUT("/posts/:id", h.UpdatePost)
	g.DELETE("/posts/:id", h.DeletePost)

	// Status transitions
	g.POST("/posts/:id/publish", h.Publish)
	g.POST("/posts/:id/unpublish", h.Unpublish)
	g.POST("/posts/:id/revert-to-draft", h.RevertToDraft)
	g.POST("/posts/:id/restore", h.Restore)

	// Revisions
	g.GET("/posts/:id/revisions", h.ListRevisions)
	g.POST("/posts/:id/revisions/:rev_id/rollback", h.Rollback)

	// Translations
	g.GET("/posts/:id/translations", h.ListTranslations)
	g.GET("/posts/:id/translations/:locale", h.GetTranslation)
	g.PUT("/posts/:id/translations/:locale", h.UpsertTranslation)
	g.DELETE("/posts/:id/translations/:locale", h.DeleteTranslation)

	// Preview tokens
	g.POST("/posts/:id/preview", h.CreatePreviewToken)
	g.GET("/posts/:id/preview", h.ListPreviewTokens)
	g.DELETE("/posts/:id/preview", h.RevokeAllPreviewTokens)
	g.DELETE("/posts/:id/preview/:token_id", h.RevokePreviewToken)

	return r
}

func doHandlerReq(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
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

func parseResp(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

// ---------------------------------------------------------------------------
// Tests: CRUD
// ---------------------------------------------------------------------------

func TestHandler_ListPosts_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.listPosts = []model.Post{
		{ID: "p1", Title: "Post 1", Slug: "post-1", Status: model.PostStatusDraft, Version: 1},
	}
	env.posts.listTotal = 1

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodGet, "/api/v1/site/posts?page=1&per_page=20", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["meta"])
}

func TestHandler_CreatePost_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.slugExists = false

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodPost, "/api/v1/site/posts", map[string]string{
		"title":   "New Post",
		"content": "Hello world",
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

func TestHandler_CreatePost_InvalidJSON(t *testing.T) {
	env := newTestEnv()
	r := setupHandlerRouter(env)

	// Missing required title field.
	w := doHandlerReq(r, http.MethodPost, "/api/v1/site/posts", map[string]string{
		"content": "body only",
	})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_GetPost_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodGet, "/api/v1/site/posts/post-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

func TestHandler_GetPost_NotFound(t *testing.T) {
	env := newTestEnv()
	env.posts.getByIDErr = apperror.NotFound("post not found", nil)

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodGet, "/api/v1/site/posts/nonexistent", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_UpdatePost_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodPut, "/api/v1/site/posts/post-1", map[string]interface{}{
		"title":   "Updated",
		"version": 1,
	})
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

func TestHandler_UpdatePost_VersionConflict(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.posts.updateErr = apperror.VersionConflict("post has been modified", nil)

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodPut, "/api/v1/site/posts/post-1", map[string]interface{}{
		"title":   "Updated",
		"version": 1,
	})
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandler_DeletePost_Success(t *testing.T) {
	env := newTestEnv()
	r := setupHandlerRouter(env)

	w := doHandlerReq(r, http.MethodDelete, "/api/v1/site/posts/post-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

// ---------------------------------------------------------------------------
// Tests: Status transitions
// ---------------------------------------------------------------------------

func TestHandler_Publish_Success(t *testing.T) {
	env := newTestEnv()
	p := testPost()
	p.Status = model.PostStatusDraft
	env.posts.getByID = p

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodPost, "/api/v1/site/posts/post-1/publish", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

func TestHandler_Publish_InvalidTransition(t *testing.T) {
	env := newTestEnv()
	p := testPost()
	p.Status = model.PostStatusPublished
	env.posts.getByID = p

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodPost, "/api/v1/site/posts/post-1/publish", nil)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Revisions
// ---------------------------------------------------------------------------

func TestHandler_ListRevisions_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.revs.listRevs = []model.PostRevision{
		{ID: "rev-1", PostID: "post-1", Version: 1, DiffSummary: "Initial version", CreatedAt: time.Now()},
	}

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodGet, "/api/v1/site/posts/post-1/revisions", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

func TestHandler_Rollback_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.revs.getByID = &model.PostRevision{
		ID:      "rev-1",
		PostID:  "post-1",
		Version: 1,
		Title:   "Old Title",
		Content: "Old Content",
	}

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodPost, "/api/v1/site/posts/post-1/revisions/rev-1/rollback", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

// ---------------------------------------------------------------------------
// Tests: Translations
// ---------------------------------------------------------------------------

func TestHandler_ListTranslations_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.trans.listTrans = []model.PostTranslation{
		{PostID: "post-1", Locale: "en", Title: "English Title"},
	}

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodGet, "/api/v1/site/posts/post-1/translations", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

func TestHandler_UpsertTranslation_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.trans.getTrans = &model.PostTranslation{
		PostID: "post-1", Locale: "en", Title: "English Title",
	}

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodPut, "/api/v1/site/posts/post-1/translations/en", map[string]string{
		"title":   "English Title",
		"content": "English content",
	})
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

func TestHandler_DeleteTranslation_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodDelete, "/api/v1/site/posts/post-1/translations/en", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

// ---------------------------------------------------------------------------
// Tests: Preview tokens
// ---------------------------------------------------------------------------

func TestHandler_CreatePreviewToken_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.preview.countActive = 0

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodPost, "/api/v1/site/posts/post-1/preview", nil)
	assert.Equal(t, http.StatusCreated, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

func TestHandler_ListPreviewTokens_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.preview.listTokens = []model.PreviewToken{
		{ID: "tok-1", PostID: "post-1", ExpiresAt: time.Now().Add(time.Hour)},
	}

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodGet, "/api/v1/site/posts/post-1/preview", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

func TestHandler_RevokeAllPreviewTokens_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.preview.delAllCount = 3

	r := setupHandlerRouter(env)
	w := doHandlerReq(r, http.MethodDelete, "/api/v1/site/posts/post-1/preview", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

func TestHandler_RevokePreviewToken_Success(t *testing.T) {
	env := newTestEnv()
	r := setupHandlerRouter(env)

	w := doHandlerReq(r, http.MethodDelete, "/api/v1/site/posts/post-1/preview/tok-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp["success"].(bool))
}

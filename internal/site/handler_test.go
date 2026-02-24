package site_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/site"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, *testEnv) {
	gin.SetMode(gin.TestMode)
	env := newTestEnv()
	h := site.NewHandler(env.svc)

	r := gin.New()
	g := r.Group("/api/v1/sites")
	g.GET("", h.ListSites)
	g.POST("", h.CreateSite)
	g.GET("/:slug", h.GetSite)
	g.PUT("/:slug", h.UpdateSite)
	g.DELETE("/:slug", h.DeleteSite)
	g.GET("/:slug/users", h.ListSiteUsers)
	g.PUT("/:slug/users/:user_id/role", h.AssignSiteRole)
	g.DELETE("/:slug/users/:user_id/role", h.RemoveSiteRole)

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
// Tests: ListSites
// ---------------------------------------------------------------------------

func TestHandler_ListSites_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.siteRepo.listSites = []model.Site{*testSite()}
	env.siteRepo.listTotal = 1

	w := doRequest(r, http.MethodGet, "/api/v1/sites?page=1&per_page=10", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["meta"])
}

// ---------------------------------------------------------------------------
// Tests: CreateSite
// ---------------------------------------------------------------------------

func TestHandler_CreateSite_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.siteRepo.slugExists = false

	w := doRequest(r, http.MethodPost, "/api/v1/sites", map[string]string{
		"name": "New Site", "slug": "new_site",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_CreateSite_ValidationError(t *testing.T) {
	r, _ := setupTestRouter()

	w := doRequest(r, http.MethodPost, "/api/v1/sites", map[string]string{
		"slug": "new_site", // missing name
	})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GetSite
// ---------------------------------------------------------------------------

func TestHandler_GetSite_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.siteRepo.getBySlug = testSite()

	w := doRequest(r, http.MethodGet, "/api/v1/sites/test_site", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_GetSite_NotFound(t *testing.T) {
	r, env := setupTestRouter()
	env.siteRepo.getBySlugErr = apperror.NotFound("site not found", nil)

	w := doRequest(r, http.MethodGet, "/api/v1/sites/nope", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: UpdateSite
// ---------------------------------------------------------------------------

func TestHandler_UpdateSite_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.siteRepo.getBySlug = testSite()

	w := doRequest(r, http.MethodPut, "/api/v1/sites/test_site", map[string]string{
		"name": "Updated",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: DeleteSite
// ---------------------------------------------------------------------------

func TestHandler_DeleteSite_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.siteRepo.getBySlug = testSite()
	env.siteRepo.countActive = 2

	w := doRequest(r, http.MethodDelete, "/api/v1/sites/test_site", map[string]string{
		"confirm_slug": "test_site",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_DeleteSite_MissingConfirm(t *testing.T) {
	r, _ := setupTestRouter()

	w := doRequest(r, http.MethodDelete, "/api/v1/sites/test_site", map[string]string{})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: ListSiteUsers
// ---------------------------------------------------------------------------

func TestHandler_ListSiteUsers_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.siteRepo.getBySlug = testSite()
	env.urRepo.listUsers = []site.UserWithRole{
		{User: model.User{ID: "u1", Email: "a@b.com"}, RoleSlug: "admin"},
	}
	env.urRepo.listTotal = 1

	w := doRequest(r, http.MethodGet, "/api/v1/sites/test_site/users", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: AssignSiteRole
// ---------------------------------------------------------------------------

func TestHandler_AssignSiteRole_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.siteRepo.getBySlug = testSite()
	env.urRepo.userExists = true
	env.roleRes.role = &model.Role{ID: "role-1", Slug: "editor"}

	w := doRequest(r, http.MethodPut, "/api/v1/sites/test_site/users/user-1/role", map[string]string{
		"role": "editor",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_AssignSiteRole_InvalidBody(t *testing.T) {
	r, _ := setupTestRouter()

	w := doRequest(r, http.MethodPut, "/api/v1/sites/test_site/users/user-1/role", map[string]string{})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: RemoveSiteRole
// ---------------------------------------------------------------------------

func TestHandler_RemoveSiteRole_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.siteRepo.getBySlug = testSite()

	w := doRequest(r, http.MethodDelete, "/api/v1/sites/test_site/users/user-1/role", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

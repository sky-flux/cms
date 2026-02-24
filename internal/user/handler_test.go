package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/mail"
	"github.com/sky-flux/cms/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, *testEnv) {
	gin.SetMode(gin.TestMode)
	env := newTestEnv()
	h := user.NewHandler(env.svc)

	r := gin.New()

	// Inject user_id into context for delete self-check.
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "caller-id")
		c.Next()
	})

	g := r.Group("/api/v1/users")
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
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
	env.userRepo.listUsers = []model.User{*testUser()}
	env.userRepo.listTotal = 1
	env.urRepo.roleSlug = "admin"

	w := doRequest(r, http.MethodGet, "/api/v1/users?page=1&per_page=10", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["meta"])
}

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestHandler_Create_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.roleRepo.role = &model.Role{ID: "role-1", Slug: "editor"}
	env.userRepo.getByEmailErr = apperror.NotFound("not found", nil)

	w := doRequest(r, http.MethodPost, "/api/v1/users", map[string]string{
		"email": "new@example.com", "password": "password123",
		"display_name": "New User", "role": "editor",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_Create_ValidationError(t *testing.T) {
	r, _ := setupTestRouter()

	w := doRequest(r, http.MethodPost, "/api/v1/users", map[string]string{
		"email": "bad", // invalid email, missing required fields
	})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Get
// ---------------------------------------------------------------------------

func TestHandler_Get_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.userRepo.getByID = testUser()
	env.urRepo.roleSlug = "admin"

	w := doRequest(r, http.MethodGet, "/api/v1/users/user-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Get_NotFound(t *testing.T) {
	r, env := setupTestRouter()
	env.userRepo.getByIDErr = apperror.NotFound("user not found", nil)

	w := doRequest(r, http.MethodGet, "/api/v1/users/nope", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestHandler_Update_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.userRepo.getByID = testUser()
	env.urRepo.roleSlug = "admin"

	w := doRequest(r, http.MethodPut, "/api/v1/users/user-1", map[string]string{
		"display_name": "Updated",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Delete
// ---------------------------------------------------------------------------

func TestHandler_Delete_Success(t *testing.T) {
	r, env := setupTestRouter()
	env.userRepo.getByID = testUser()
	env.urRepo.roleSlug = "editor"

	w := doRequest(r, http.MethodDelete, "/api/v1/users/user-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Delete_SelfDelete(t *testing.T) {
	r, env := setupTestRouter()
	env.userRepo.getByID = testUser()

	// The middleware sets user_id to "caller-id", so deleting "caller-id" is self-delete.
	w := doRequest(r, http.MethodDelete, "/api/v1/users/caller-id", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// Need a separate router setup without the NoopSender to verify mailer is passed correctly.
func newTestEnvWithMailer() *testEnv {
	ur := &mockUserRepo{}
	rr := &mockRoleRepo{}
	urr := &mockUserRoleRepo{}
	tr := &mockTokenRevoker{}
	al := &mockAuditLogger{}
	ml := &mail.NoopSender{}
	return &testEnv{
		svc:          user.NewService(ur, rr, urr, tr, al, ml, "TestSite"),
		userRepo:     ur,
		roleRepo:     rr,
		urRepo:       urr,
		tokenRevoker: tr,
		auditLog:     al,
		mailer:       ml,
	}
}

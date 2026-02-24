package rbac_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Handler mock repos (prefixed to avoid conflict with service_test.go mocks) ---

type handlerMockRoleRepo struct {
	roles     []model.Role
	listErr   error
	byID      *model.Role
	byIDErr   error
	createErr error
	updateErr error
	deleteErr error
}

func (m *handlerMockRoleRepo) List(_ context.Context) ([]model.Role, error) {
	return m.roles, m.listErr
}
func (m *handlerMockRoleRepo) GetByID(_ context.Context, _ string) (*model.Role, error) {
	return m.byID, m.byIDErr
}
func (m *handlerMockRoleRepo) GetBySlug(_ context.Context, _ string) (*model.Role, error) {
	return nil, nil
}
func (m *handlerMockRoleRepo) Create(_ context.Context, _ *model.Role) error { return m.createErr }
func (m *handlerMockRoleRepo) Update(_ context.Context, _ *model.Role) error { return m.updateErr }
func (m *handlerMockRoleRepo) Delete(_ context.Context, _ string) error      { return m.deleteErr }

type handlerMockAPIRepo struct {
	apis []model.APIEndpoint
	err  error
}

func (m *handlerMockAPIRepo) UpsertBatch(_ context.Context, _ []model.APIEndpoint) error {
	return nil
}
func (m *handlerMockAPIRepo) DisableStale(_ context.Context, _ []string) error { return nil }
func (m *handlerMockAPIRepo) List(_ context.Context) ([]model.APIEndpoint, error) {
	return m.apis, m.err
}
func (m *handlerMockAPIRepo) ListByGroup(_ context.Context, _ string) ([]model.APIEndpoint, error) {
	return nil, nil
}
func (m *handlerMockAPIRepo) GetByMethodPath(_ context.Context, _, _ string) (*model.APIEndpoint, error) {
	return nil, nil
}

type handlerMockRoleAPIRepo struct {
	apis     []model.APIEndpoint
	err      error
	setErr   error
	cloneErr error
}

func (m *handlerMockRoleAPIRepo) GetAPIsByRoleID(_ context.Context, _ string) ([]model.APIEndpoint, error) {
	return m.apis, m.err
}
func (m *handlerMockRoleAPIRepo) SetRoleAPIs(_ context.Context, _ string, _ []string) error {
	return m.setErr
}
func (m *handlerMockRoleAPIRepo) GetRoleIDsByMethodPath(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}
func (m *handlerMockRoleAPIRepo) CloneFromTemplate(_ context.Context, _, _ string) error {
	return m.cloneErr
}

type handlerMockMenuRepo struct {
	menus      []model.AdminMenu
	err        error
	setErr     error
	listTree   []model.AdminMenu
	listTreeErr error
	createErr  error
	updateErr  error
	deleteErr  error
}

func (m *handlerMockMenuRepo) ListTree(_ context.Context) ([]model.AdminMenu, error) {
	return m.listTree, m.listTreeErr
}
func (m *handlerMockMenuRepo) Create(_ context.Context, _ *model.AdminMenu) error { return m.createErr }
func (m *handlerMockMenuRepo) Update(_ context.Context, _ *model.AdminMenu) error { return m.updateErr }
func (m *handlerMockMenuRepo) Delete(_ context.Context, _ string) error           { return m.deleteErr }
func (m *handlerMockMenuRepo) GetMenusByRoleID(_ context.Context, _ string) ([]model.AdminMenu, error) {
	return m.menus, m.err
}
func (m *handlerMockMenuRepo) SetRoleMenus(_ context.Context, _ string, _ []string) error {
	return m.setErr
}
func (m *handlerMockMenuRepo) GetMenusByUserID(_ context.Context, _ string) ([]model.AdminMenu, error) {
	return m.menus, nil
}

type handlerMockTemplateRepo struct {
	templates []model.RoleTemplate
	byID      *model.RoleTemplate
	byIDErr   error
	createErr error
	deleteErr error
}

func (m *handlerMockTemplateRepo) List(_ context.Context) ([]model.RoleTemplate, error) {
	return m.templates, nil
}
func (m *handlerMockTemplateRepo) GetByID(_ context.Context, _ string) (*model.RoleTemplate, error) {
	return m.byID, m.byIDErr
}
func (m *handlerMockTemplateRepo) Create(_ context.Context, _ *model.RoleTemplate) error {
	return m.createErr
}
func (m *handlerMockTemplateRepo) Update(_ context.Context, _ *model.RoleTemplate) error { return nil }
func (m *handlerMockTemplateRepo) Delete(_ context.Context, _ string) error {
	return m.deleteErr
}
func (m *handlerMockTemplateRepo) GetTemplateAPIs(_ context.Context, _ string) ([]model.APIEndpoint, error) {
	return nil, nil
}
func (m *handlerMockTemplateRepo) SetTemplateAPIs(_ context.Context, _ string, _ []string) error {
	return nil
}
func (m *handlerMockTemplateRepo) GetTemplateMenus(_ context.Context, _ string) ([]model.AdminMenu, error) {
	return nil, nil
}
func (m *handlerMockTemplateRepo) SetTemplateMenus(_ context.Context, _ string, _ []string) error {
	return nil
}

// --- Helper ---

func setupHandlerTest(t *testing.T, roleRepo *handlerMockRoleRepo, apiRepo *handlerMockAPIRepo, roleAPIRepo *handlerMockRoleAPIRepo, menuRepo *handlerMockMenuRepo, templateRepo *handlerMockTemplateRepo) *rbac.Handler {
	t.Helper()
	if roleRepo == nil {
		roleRepo = &handlerMockRoleRepo{}
	}
	if apiRepo == nil {
		apiRepo = &handlerMockAPIRepo{}
	}
	if roleAPIRepo == nil {
		roleAPIRepo = &handlerMockRoleAPIRepo{}
	}
	if menuRepo == nil {
		menuRepo = &handlerMockMenuRepo{}
	}
	if templateRepo == nil {
		templateRepo = &handlerMockTemplateRepo{}
	}

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	userRoleRepo := &mockUserRoleRepo{slugs: []string{"super"}, roles: []model.Role{{ID: "r1", Slug: "super"}}}
	svc := rbac.NewService(userRoleRepo, roleAPIRepo, menuRepo, rdb)

	return rbac.NewHandler(svc, roleRepo, apiRepo, roleAPIRepo, menuRepo, templateRepo, userRoleRepo)
}

func doJSON(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

// --- Tests ---

func TestHandler_ListRoles_Success(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{roles: []model.Role{{ID: "1", Name: "Admin", Slug: "admin"}}}
	h := setupHandlerTest(t, roleRepo, nil, nil, nil, nil)

	r := gin.New()
	r.GET("/roles", h.ListRoles)
	w := doJSON(r, "GET", "/roles", "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_ListRoles_Error(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{listErr: apperror.Internal("db error", nil)}
	h := setupHandlerTest(t, roleRepo, nil, nil, nil, nil)

	r := gin.New()
	r.GET("/roles", h.ListRoles)
	w := doJSON(r, "GET", "/roles", "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_CreateRole_Success(t *testing.T) {
	h := setupHandlerTest(t, nil, nil, nil, nil, nil)

	r := gin.New()
	r.POST("/roles", h.CreateRole)
	w := doJSON(r, "POST", "/roles", `{"name":"Editor","slug":"editor"}`)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_CreateRole_InvalidJSON(t *testing.T) {
	h := setupHandlerTest(t, nil, nil, nil, nil, nil)

	r := gin.New()
	r.POST("/roles", h.CreateRole)
	w := doJSON(r, "POST", "/roles", `{"invalid":}`)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_UpdateRole_SuperProtected(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{byID: &model.Role{ID: "1", Slug: "super", BuiltIn: model.ToggleYes}}
	h := setupHandlerTest(t, roleRepo, nil, nil, nil, nil)

	r := gin.New()
	r.PUT("/roles/:id", h.UpdateRole)
	w := doJSON(r, "PUT", "/roles/1", `{"name":"New Name"}`)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_UpdateRole_NotFound(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{byIDErr: apperror.NotFound("role not found", nil)}
	h := setupHandlerTest(t, roleRepo, nil, nil, nil, nil)

	r := gin.New()
	r.PUT("/roles/:id", h.UpdateRole)
	w := doJSON(r, "PUT", "/roles/999", `{"name":"New Name"}`)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_DeleteRole_Success(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{byID: &model.Role{ID: "1", Slug: "custom", BuiltIn: model.ToggleNo}}
	h := setupHandlerTest(t, roleRepo, nil, nil, nil, nil)

	r := gin.New()
	r.DELETE("/roles/:id", h.DeleteRole)
	w := doJSON(r, "DELETE", "/roles/1", "")

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandler_DeleteRole_BuiltInProtected(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{byID: &model.Role{ID: "1", Slug: "admin", BuiltIn: model.ToggleYes}}
	h := setupHandlerTest(t, roleRepo, nil, nil, nil, nil)

	r := gin.New()
	r.DELETE("/roles/:id", h.DeleteRole)
	w := doJSON(r, "DELETE", "/roles/1", "")

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_SetRoleAPIs_Success(t *testing.T) {
	h := setupHandlerTest(t, nil, nil, &handlerMockRoleAPIRepo{}, nil, nil)

	r := gin.New()
	r.PUT("/roles/:id/apis", h.SetRoleAPIs)
	w := doJSON(r, "PUT", "/roles/1/apis", `{"api_ids":["a1","a2"]}`)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandler_SetRoleAPIs_InvalidJSON(t *testing.T) {
	h := setupHandlerTest(t, nil, nil, nil, nil, nil)

	r := gin.New()
	r.PUT("/roles/:id/apis", h.SetRoleAPIs)
	w := doJSON(r, "PUT", "/roles/1/apis", `not json`)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_DeleteTemplate_BuiltIn(t *testing.T) {
	templateRepo := &handlerMockTemplateRepo{byID: &model.RoleTemplate{ID: "1", BuiltIn: model.ToggleYes}}
	h := setupHandlerTest(t, nil, nil, nil, nil, templateRepo)

	r := gin.New()
	r.DELETE("/templates/:id", h.DeleteTemplate)
	w := doJSON(r, "DELETE", "/templates/1", "")

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_GetMyMenus_Success(t *testing.T) {
	menuRepo := &handlerMockMenuRepo{menus: []model.AdminMenu{{ID: "m1", Name: "Dashboard"}}}
	h := setupHandlerTest(t, nil, nil, nil, menuRepo, nil)

	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("user_id", "user-1"); c.Next() })
	r.GET("/me/menus", h.GetMyMenus)
	w := doJSON(r, "GET", "/me/menus", "")

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_ListAPIs_Success(t *testing.T) {
	apiRepo := &handlerMockAPIRepo{apis: []model.APIEndpoint{{Method: "GET", Path: "/api/v1/posts"}}}
	h := setupHandlerTest(t, nil, apiRepo, nil, nil, nil)

	r := gin.New()
	r.GET("/apis", h.ListAPIs)
	w := doJSON(r, "GET", "/apis", "")

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- New method tests ---

func TestHandler_GetRole_Success(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{byID: &model.Role{ID: "1", Name: "Admin", Slug: "admin"}}
	h := setupHandlerTest(t, roleRepo, nil, nil, nil, nil)

	r := gin.New()
	r.GET("/roles/:id", h.GetRole)
	w := doJSON(r, "GET", "/roles/1", "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_GetRole_NotFound(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{byIDErr: apperror.NotFound("role not found", nil)}
	h := setupHandlerTest(t, roleRepo, nil, nil, nil, nil)

	r := gin.New()
	r.GET("/roles/:id", h.GetRole)
	w := doJSON(r, "GET", "/roles/999", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_GetTemplate_Success(t *testing.T) {
	tmplRepo := &handlerMockTemplateRepo{byID: &model.RoleTemplate{ID: "t1", Name: "Editor Template"}}
	h := setupHandlerTest(t, nil, nil, nil, nil, tmplRepo)

	r := gin.New()
	r.GET("/templates/:id", h.GetTemplate)
	w := doJSON(r, "GET", "/templates/t1", "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_GetTemplate_NotFound(t *testing.T) {
	tmplRepo := &handlerMockTemplateRepo{byIDErr: apperror.NotFound("template not found", nil)}
	h := setupHandlerTest(t, nil, nil, nil, nil, tmplRepo)

	r := gin.New()
	r.GET("/templates/:id", h.GetTemplate)
	w := doJSON(r, "GET", "/templates/999", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_ApplyTemplate_Success(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{byID: &model.Role{ID: "r1", Slug: "editor", BuiltIn: model.ToggleNo}}
	tmplRepo := &handlerMockTemplateRepo{byID: &model.RoleTemplate{ID: "t1"}}
	h := setupHandlerTest(t, roleRepo, nil, &handlerMockRoleAPIRepo{}, nil, tmplRepo)

	r := gin.New()
	r.POST("/roles/:id/apply-template", h.ApplyTemplate)
	w := doJSON(r, "POST", "/roles/r1/apply-template", `{"template_id":"00000000-0000-0000-0000-000000000001"}`)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandler_ApplyTemplate_SuperProtected(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{byID: &model.Role{ID: "r1", Slug: "super", BuiltIn: model.ToggleYes}}
	h := setupHandlerTest(t, roleRepo, nil, nil, nil, nil)

	r := gin.New()
	r.POST("/roles/:id/apply-template", h.ApplyTemplate)
	w := doJSON(r, "POST", "/roles/r1/apply-template", `{"template_id":"00000000-0000-0000-0000-000000000001"}`)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_ApplyTemplate_RoleNotFound(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{byIDErr: apperror.NotFound("role not found", nil)}
	h := setupHandlerTest(t, roleRepo, nil, nil, nil, nil)

	r := gin.New()
	r.POST("/roles/:id/apply-template", h.ApplyTemplate)
	w := doJSON(r, "POST", "/roles/999/apply-template", `{"template_id":"00000000-0000-0000-0000-000000000001"}`)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_ApplyTemplate_TemplateNotFound(t *testing.T) {
	roleRepo := &handlerMockRoleRepo{byID: &model.Role{ID: "r1", Slug: "editor"}}
	tmplRepo := &handlerMockTemplateRepo{byIDErr: apperror.NotFound("template not found", nil)}
	h := setupHandlerTest(t, roleRepo, nil, nil, nil, tmplRepo)

	r := gin.New()
	r.POST("/roles/:id/apply-template", h.ApplyTemplate)
	w := doJSON(r, "POST", "/roles/r1/apply-template", `{"template_id":"00000000-0000-0000-0000-000000000001"}`)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_ListMenus_Success(t *testing.T) {
	menuRepo := &handlerMockMenuRepo{listTree: []model.AdminMenu{{ID: "m1", Name: "Dashboard"}}}
	h := setupHandlerTest(t, nil, nil, nil, menuRepo, nil)

	r := gin.New()
	r.GET("/menus", h.ListMenus)
	w := doJSON(r, "GET", "/menus", "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_CreateMenu_Success(t *testing.T) {
	h := setupHandlerTest(t, nil, nil, nil, &handlerMockMenuRepo{}, nil)

	r := gin.New()
	r.POST("/menus", h.CreateMenu)
	w := doJSON(r, "POST", "/menus", `{"name":"Reports","path":"/reports"}`)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_DeleteMenu_Success(t *testing.T) {
	h := setupHandlerTest(t, nil, nil, nil, &handlerMockMenuRepo{}, nil)

	r := gin.New()
	r.DELETE("/menus/:id", h.DeleteMenu)
	w := doJSON(r, "DELETE", "/menus/m1", "")

	assert.Equal(t, http.StatusNoContent, w.Code)
}

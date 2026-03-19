package delivery_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/sky-flux/cms/internal/site/app"
	"github.com/sky-flux/cms/internal/site/delivery"
	"github.com/sky-flux/cms/internal/site/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- hand-written mocks ----

type mockGetSite struct {
	site *domain.Site
	err  error
}

func (m *mockGetSite) Execute(ctx context.Context) (*domain.Site, error) { return m.site, m.err }

type mockUpdateSite struct{ err error }

func (m *mockUpdateSite) Execute(ctx context.Context, in app.UpdateSiteConfigInput) error {
	return m.err
}

type mockCreateMenu struct {
	menu *domain.Menu
	err  error
}

func (m *mockCreateMenu) Execute(ctx context.Context, in app.CreateMenuInput) (*domain.Menu, error) {
	return m.menu, m.err
}

type mockAddMenuItem struct {
	item *domain.MenuItem
	err  error
}

func (m *mockAddMenuItem) Execute(ctx context.Context, in app.AddMenuItemInput) (*domain.MenuItem, error) {
	return m.item, m.err
}

type mockCreateRedirect struct {
	redirect *domain.Redirect
	err      error
}

func (m *mockCreateRedirect) Execute(ctx context.Context, in app.CreateRedirectInput) (*domain.Redirect, error) {
	return m.redirect, m.err
}

// ---- helper ----

func newTestAPI(t *testing.T,
	getSite delivery.GetSiteExecutor,
	updateSite delivery.UpdateSiteExecutor,
	createMenu delivery.CreateMenuExecutor,
	addItem delivery.AddMenuItemExecutor,
	createRedirect delivery.CreateRedirectExecutor,
) huma.API {
	t.Helper()
	_, api := humatest.New(t, huma.DefaultConfig("Test API", "0.0.1"))
	delivery.RegisterRoutes(api, getSite, updateSite, createMenu, addItem, createRedirect)
	return api
}

// ---- tests ----

func TestGetSettings_Success(t *testing.T) {
	site := &domain.Site{ID: 1, Name: "My Blog", Language: "en", Timezone: "UTC"}
	api := newTestAPI(t,
		&mockGetSite{site: site},
		&mockUpdateSite{},
		&mockCreateMenu{},
		&mockAddMenuItem{},
		&mockCreateRedirect{},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	resp := httptest.NewRecorder()
	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "My Blog", result["name"])
}

func TestUpdateSettings_Success(t *testing.T) {
	api := newTestAPI(t,
		&mockGetSite{site: &domain.Site{Name: "Blog", Language: "en", Timezone: "UTC"}},
		&mockUpdateSite{},
		&mockCreateMenu{},
		&mockAddMenuItem{},
		&mockCreateRedirect{},
	)

	body := `{"name":"Updated","language":"zh-CN","timezone":"Asia/Shanghai"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestCreateMenu_Success(t *testing.T) {
	menu := &domain.Menu{ID: "m1", Name: "Main Nav", Slug: "main-nav"}
	api := newTestAPI(t,
		&mockGetSite{site: &domain.Site{}},
		&mockUpdateSite{},
		&mockCreateMenu{menu: menu},
		&mockAddMenuItem{},
		&mockCreateRedirect{},
	)

	body := `{"name":"Main Nav","slug":"main-nav"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/menus", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "m1", result["id"])
}

func TestCreateRedirect_Success(t *testing.T) {
	redirect := &domain.Redirect{ID: "r1", FromPath: "/old", ToPath: "/new", StatusCode: 301}
	api := newTestAPI(t,
		&mockGetSite{site: &domain.Site{}},
		&mockUpdateSite{},
		&mockCreateMenu{},
		&mockAddMenuItem{},
		&mockCreateRedirect{redirect: redirect},
	)

	body := `{"from_path":"/old","to_path":"/new","status_code":301}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/redirects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
}

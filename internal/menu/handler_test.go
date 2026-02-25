package menu_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/menu"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Handler test helpers
// ---------------------------------------------------------------------------

func setupRouter() (*gin.Engine, *mockMenuRepo, *mockItemRepo) {
	gin.SetMode(gin.TestMode)
	mr := &mockMenuRepo{}
	ir := &mockItemRepo{}
	a := audit.NewNoopLogger()
	svc := menu.NewService(mr, ir, a)
	h := menu.NewHandler(svc)

	r := gin.New()
	g := r.Group("/api/v1/menus")
	g.GET("", h.ListMenus)
	g.POST("", h.CreateMenu)
	g.GET("/:id", h.GetMenu)
	g.PUT("/:id", h.UpdateMenu)
	g.DELETE("/:id", h.DeleteMenu)
	g.POST("/:id/items", h.AddItem)
	g.PUT("/:id/items/reorder", h.ReorderItems) // BEFORE :item_id
	g.PUT("/:id/items/:item_id", h.UpdateItem)
	g.DELETE("/:id/items/:item_id", h.DeleteItem)

	return r, mr, ir
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

func defaultMenu() *model.SiteMenu {
	return &model.SiteMenu{
		ID:        "menu-1",
		Name:      "Main",
		Slug:      "main",
		Location:  "header",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func defaultItem() *model.SiteMenuItem {
	return &model.SiteMenuItem{
		ID:        "item-1",
		MenuID:    "menu-1",
		Label:     "Home",
		URL:       "https://example.com",
		Target:    "_self",
		Type:      model.MenuItemTypeCustom,
		SortOrder: 0,
		Status:    model.MenuItemStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Tests: ListMenus
// ---------------------------------------------------------------------------

func TestHandler_ListMenus(t *testing.T) {
	r, mr, _ := setupRouter()
	mr.listMenus = []model.SiteMenu{*defaultMenu()}
	mr.itemCount = 2

	w := doRequest(r, http.MethodGet, "/api/v1/menus", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_ListMenus_WithLocation(t *testing.T) {
	r, mr, _ := setupRouter()
	mr.listMenus = []model.SiteMenu{*defaultMenu()}

	w := doRequest(r, http.MethodGet, "/api/v1/menus?location=header", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GetMenu
// ---------------------------------------------------------------------------

func TestHandler_GetMenu(t *testing.T) {
	r, mr, ir := setupRouter()
	mr.getByID = defaultMenu()
	ir.listItems = []*model.SiteMenuItem{defaultItem()}

	w := doRequest(r, http.MethodGet, "/api/v1/menus/menu-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_GetMenu_NotFound(t *testing.T) {
	r, mr, _ := setupRouter()
	mr.getByIDErr = apperror.NotFound("menu not found", nil)

	w := doRequest(r, http.MethodGet, "/api/v1/menus/nonexistent", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: CreateMenu
// ---------------------------------------------------------------------------

func TestHandler_CreateMenu(t *testing.T) {
	r, _, _ := setupRouter()

	w := doRequest(r, http.MethodPost, "/api/v1/menus", map[string]string{
		"name": "Footer", "slug": "footer",
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestHandler_CreateMenu_ValidationError(t *testing.T) {
	r, _, _ := setupRouter()

	// Missing required fields.
	w := doRequest(r, http.MethodPost, "/api/v1/menus", map[string]string{})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_CreateMenu_DuplicateSlug(t *testing.T) {
	r, mr, _ := setupRouter()
	mr.slugExists = true

	w := doRequest(r, http.MethodPost, "/api/v1/menus", map[string]string{
		"name": "Footer", "slug": "footer",
	})
	assert.Equal(t, http.StatusConflict, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: UpdateMenu
// ---------------------------------------------------------------------------

func TestHandler_UpdateMenu(t *testing.T) {
	r, mr, _ := setupRouter()
	mr.getByID = defaultMenu()

	w := doRequest(r, http.MethodPut, "/api/v1/menus/menu-1", map[string]string{
		"name": "Updated Nav",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_UpdateMenu_NotFound(t *testing.T) {
	r, mr, _ := setupRouter()
	mr.getByIDErr = apperror.NotFound("menu not found", nil)

	w := doRequest(r, http.MethodPut, "/api/v1/menus/nonexistent", map[string]string{
		"name": "Nope",
	})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: DeleteMenu
// ---------------------------------------------------------------------------

func TestHandler_DeleteMenu(t *testing.T) {
	r, mr, _ := setupRouter()
	mr.getByID = defaultMenu()

	w := doRequest(r, http.MethodDelete, "/api/v1/menus/menu-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_DeleteMenu_NotFound(t *testing.T) {
	r, mr, _ := setupRouter()
	mr.getByIDErr = apperror.NotFound("menu not found", nil)

	w := doRequest(r, http.MethodDelete, "/api/v1/menus/nonexistent", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: AddItem
// ---------------------------------------------------------------------------

func TestHandler_AddItem(t *testing.T) {
	r, mr, _ := setupRouter()
	mr.getByID = defaultMenu()

	w := doRequest(r, http.MethodPost, "/api/v1/menus/menu-1/items", map[string]interface{}{
		"label": "Home",
		"url":   "https://example.com",
		"type":  "custom",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_AddItem_ValidationError(t *testing.T) {
	r, _, _ := setupRouter()

	// Missing required fields.
	w := doRequest(r, http.MethodPost, "/api/v1/menus/menu-1/items", map[string]string{})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: UpdateItem
// ---------------------------------------------------------------------------

func TestHandler_UpdateItem(t *testing.T) {
	r, _, ir := setupRouter()
	ir.belongsToMenu = true
	ir.getByID = defaultItem()

	w := doRequest(r, http.MethodPut, "/api/v1/menus/menu-1/items/item-1", map[string]string{
		"label": "Updated Home",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_UpdateItem_NotInMenu(t *testing.T) {
	r, _, ir := setupRouter()
	ir.belongsToMenu = false

	w := doRequest(r, http.MethodPut, "/api/v1/menus/menu-1/items/item-other", map[string]string{
		"label": "Nope",
	})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: DeleteItem
// ---------------------------------------------------------------------------

func TestHandler_DeleteItem(t *testing.T) {
	r, _, ir := setupRouter()
	ir.belongsToMenu = true

	w := doRequest(r, http.MethodDelete, "/api/v1/menus/menu-1/items/item-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_DeleteItem_NotInMenu(t *testing.T) {
	r, _, ir := setupRouter()
	ir.belongsToMenu = false

	w := doRequest(r, http.MethodDelete, "/api/v1/menus/menu-1/items/item-other", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: ReorderItems
// ---------------------------------------------------------------------------

func TestHandler_ReorderItems(t *testing.T) {
	r, _, ir := setupRouter()
	ir.belongsToMenu = true

	w := doRequest(r, http.MethodPut, "/api/v1/menus/menu-1/items/reorder", map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": "item-1", "sort_order": 1},
			{"id": "item-2", "sort_order": 0},
		},
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_ReorderItems_ValidationError(t *testing.T) {
	r, _, _ := setupRouter()

	// Empty items.
	w := doRequest(r, http.MethodPut, "/api/v1/menus/menu-1/items/reorder", map[string]interface{}{
		"items": []map[string]interface{}{},
	})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_ReorderItems_InvalidItem(t *testing.T) {
	r, _, ir := setupRouter()
	ir.belongsToMenu = false

	w := doRequest(r, http.MethodPut, "/api/v1/menus/menu-1/items/reorder", map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": "foreign-item", "sort_order": 0},
		},
	})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

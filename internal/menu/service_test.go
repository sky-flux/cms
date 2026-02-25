package menu_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/menu"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: MenuRepository
// ---------------------------------------------------------------------------

type mockMenuRepo struct {
	listMenus     []model.SiteMenu
	listErr       error
	getByID       *model.SiteMenu
	getByIDErr    error
	createErr     error
	updateErr     error
	deleteErr     error
	slugExists    bool
	slugExistsErr error
	itemCount     int64
	itemCountErr  error
}

func (m *mockMenuRepo) List(_ context.Context, _ string) ([]model.SiteMenu, error) {
	return m.listMenus, m.listErr
}
func (m *mockMenuRepo) GetByID(_ context.Context, _ string) (*model.SiteMenu, error) {
	return m.getByID, m.getByIDErr
}
func (m *mockMenuRepo) Create(_ context.Context, sm *model.SiteMenu) error {
	if m.createErr == nil {
		sm.ID = "new-menu-id"
		sm.CreatedAt = time.Now()
		sm.UpdatedAt = time.Now()
	}
	return m.createErr
}
func (m *mockMenuRepo) Update(_ context.Context, _ *model.SiteMenu) error { return m.updateErr }
func (m *mockMenuRepo) Delete(_ context.Context, _ string) error           { return m.deleteErr }
func (m *mockMenuRepo) SlugExists(_ context.Context, _ string, _ string) (bool, error) {
	return m.slugExists, m.slugExistsErr
}
func (m *mockMenuRepo) CountItems(_ context.Context, _ string) (int64, error) {
	return m.itemCount, m.itemCountErr
}

// ---------------------------------------------------------------------------
// Mock: MenuItemRepository
// ---------------------------------------------------------------------------

type mockItemRepo struct {
	listItems       []*model.SiteMenuItem
	listErr         error
	getByID         *model.SiteMenuItem
	getByIDErr      error
	createErr       error
	updateErr       error
	deleteErr       error
	belongsToMenu   bool
	belongsErr      error
	batchUpdateErr  error
	depth           int
	depthErr        error
}

func (m *mockItemRepo) ListByMenuID(_ context.Context, _ string) ([]*model.SiteMenuItem, error) {
	return m.listItems, m.listErr
}
func (m *mockItemRepo) GetByID(_ context.Context, _ string) (*model.SiteMenuItem, error) {
	return m.getByID, m.getByIDErr
}
func (m *mockItemRepo) Create(_ context.Context, item *model.SiteMenuItem) error {
	if m.createErr == nil {
		item.ID = "new-item-id"
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()
	}
	return m.createErr
}
func (m *mockItemRepo) Update(_ context.Context, _ *model.SiteMenuItem) error { return m.updateErr }
func (m *mockItemRepo) Delete(_ context.Context, _ string) error              { return m.deleteErr }
func (m *mockItemRepo) BelongsToMenu(_ context.Context, _ string, _ string) (bool, error) {
	return m.belongsToMenu, m.belongsErr
}
func (m *mockItemRepo) BatchUpdateOrder(_ context.Context, _ []menu.ReorderItem) error {
	return m.batchUpdateErr
}
func (m *mockItemRepo) GetDepth(_ context.Context, _ string) (int, error) {
	return m.depth, m.depthErr
}

// ---------------------------------------------------------------------------
// Mock: AuditLogger
// ---------------------------------------------------------------------------

type mockAudit struct {
	lastEntry *audit.Entry
	err       error
}

func (m *mockAudit) Log(_ context.Context, entry audit.Entry) error {
	m.lastEntry = &entry
	return m.err
}

// ---------------------------------------------------------------------------
// Test environment
// ---------------------------------------------------------------------------

type testEnv struct {
	svc      *menu.Service
	menuRepo *mockMenuRepo
	itemRepo *mockItemRepo
	audit    *mockAudit
}

func newTestEnv() *testEnv {
	mr := &mockMenuRepo{}
	ir := &mockItemRepo{}
	a := &mockAudit{}
	return &testEnv{
		svc:      menu.NewService(mr, ir, a),
		menuRepo: mr,
		itemRepo: ir,
		audit:    a,
	}
}

func testMenu() *model.SiteMenu {
	return &model.SiteMenu{
		ID:          "menu-1",
		Name:        "Main Nav",
		Slug:        "main-nav",
		Location:    "header",
		Description: "Primary navigation",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func testMenuItem() *model.SiteMenuItem {
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

func TestListMenus_Success(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.listMenus = []model.SiteMenu{*testMenu()}
	env.menuRepo.itemCount = 3

	menus, err := env.svc.ListMenus(context.Background(), "")
	require.NoError(t, err)
	assert.Len(t, menus, 1)
	assert.Equal(t, "Main Nav", menus[0].Name)
	assert.Equal(t, int64(3), menus[0].ItemCount)
}

// ---------------------------------------------------------------------------
// Tests: CreateMenu
// ---------------------------------------------------------------------------

func TestCreateMenu_Success(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.slugExists = false

	resp, err := env.svc.CreateMenu(context.Background(), &menu.CreateMenuReq{
		Name: "Footer",
		Slug: "footer",
	})
	require.NoError(t, err)
	assert.Equal(t, "footer", resp.Slug)
	assert.Equal(t, "Footer", resp.Name)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionCreate, env.audit.lastEntry.Action)
}

func TestCreateMenu_DuplicateSlug(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.slugExists = true

	_, err := env.svc.CreateMenu(context.Background(), &menu.CreateMenuReq{
		Name: "Footer",
		Slug: "footer",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestCreateMenu_InvalidSlug(t *testing.T) {
	env := newTestEnv()

	_, err := env.svc.CreateMenu(context.Background(), &menu.CreateMenuReq{
		Name: "Bad",
		Slug: "INVALID SLUG!",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

// ---------------------------------------------------------------------------
// Tests: GetMenu
// ---------------------------------------------------------------------------

func TestGetMenu_WithTree(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.getByID = testMenu()

	parentID := "item-1"
	env.itemRepo.listItems = []*model.SiteMenuItem{
		{
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
		},
		{
			ID:        "item-2",
			MenuID:    "menu-1",
			ParentID:  &parentID,
			Label:     "Sub Item",
			URL:       "https://example.com/sub",
			Target:    "_self",
			Type:      model.MenuItemTypeCustom,
			SortOrder: 0,
			Status:    model.MenuItemStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	detail, err := env.svc.GetMenu(context.Background(), "menu-1")
	require.NoError(t, err)
	assert.Equal(t, "Main Nav", detail.Name)
	require.Len(t, detail.Items, 1)
	assert.Equal(t, "Home", detail.Items[0].Label)
	require.Len(t, detail.Items[0].Children, 1)
	assert.Equal(t, "Sub Item", detail.Items[0].Children[0].Label)
}

func TestGetMenu_NotFound(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.getByIDErr = apperror.NotFound("menu not found", nil)

	_, err := env.svc.GetMenu(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: AddItem
// ---------------------------------------------------------------------------

func TestAddItem_Success(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.getByID = testMenu()

	resp, err := env.svc.AddItem(context.Background(), "menu-1", &menu.CreateMenuItemReq{
		Label: "Home",
		URL:   "https://example.com",
		Type:  "custom",
	})
	require.NoError(t, err)
	assert.Equal(t, "Home", resp.Label)
	assert.Equal(t, "_self", resp.Target)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionCreate, env.audit.lastEntry.Action)
}

func TestAddItem_CustomRequiresURL(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.getByID = testMenu()

	_, err := env.svc.AddItem(context.Background(), "menu-1", &menu.CreateMenuItemReq{
		Label: "No URL",
		Type:  "custom",
		URL:   "",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestAddItem_NonCustomRequiresReferenceID(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.getByID = testMenu()

	_, err := env.svc.AddItem(context.Background(), "menu-1", &menu.CreateMenuItemReq{
		Label: "Post Link",
		Type:  "post",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestAddItem_MaxDepth(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.getByID = testMenu()
	env.itemRepo.belongsToMenu = true
	env.itemRepo.depth = 2 // already at depth 2, adding child would exceed 3 levels

	parentID := "item-deep"
	_, err := env.svc.AddItem(context.Background(), "menu-1", &menu.CreateMenuItemReq{
		ParentID: &parentID,
		Label:    "Too Deep",
		URL:      "https://example.com",
		Type:     "custom",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestAddItem_MenuNotFound(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.getByIDErr = apperror.NotFound("menu not found", nil)

	_, err := env.svc.AddItem(context.Background(), "nonexistent", &menu.CreateMenuItemReq{
		Label: "Home",
		URL:   "https://example.com",
		Type:  "custom",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: UpdateItem
// ---------------------------------------------------------------------------

func TestUpdateItem_Success(t *testing.T) {
	env := newTestEnv()
	env.itemRepo.belongsToMenu = true
	env.itemRepo.getByID = testMenuItem()

	newLabel := "Updated Home"
	resp, err := env.svc.UpdateItem(context.Background(), "menu-1", "item-1", &menu.UpdateMenuItemReq{
		Label: &newLabel,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Home", resp.Label)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionUpdate, env.audit.lastEntry.Action)
}

func TestUpdateItem_NotInMenu(t *testing.T) {
	env := newTestEnv()
	env.itemRepo.belongsToMenu = false

	newLabel := "Nope"
	_, err := env.svc.UpdateItem(context.Background(), "menu-1", "item-other", &menu.UpdateMenuItemReq{
		Label: &newLabel,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: DeleteItem
// ---------------------------------------------------------------------------

func TestDeleteItem_Success(t *testing.T) {
	env := newTestEnv()
	env.itemRepo.belongsToMenu = true

	err := env.svc.DeleteItem(context.Background(), "menu-1", "item-1")
	require.NoError(t, err)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionDelete, env.audit.lastEntry.Action)
}

func TestDeleteItem_NotInMenu(t *testing.T) {
	env := newTestEnv()
	env.itemRepo.belongsToMenu = false

	err := env.svc.DeleteItem(context.Background(), "menu-1", "item-other")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: ReorderItems
// ---------------------------------------------------------------------------

func TestReorderItems_Success(t *testing.T) {
	env := newTestEnv()
	env.itemRepo.belongsToMenu = true

	err := env.svc.ReorderItems(context.Background(), "menu-1", &menu.ReorderReq{
		Items: []menu.ReorderItem{
			{ID: "item-1", SortOrder: 1},
			{ID: "item-2", SortOrder: 0},
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, env.audit.lastEntry)
}

func TestReorderItems_InvalidItem(t *testing.T) {
	env := newTestEnv()
	env.itemRepo.belongsToMenu = false

	err := env.svc.ReorderItems(context.Background(), "menu-1", &menu.ReorderReq{
		Items: []menu.ReorderItem{
			{ID: "item-foreign", SortOrder: 0},
		},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

// ---------------------------------------------------------------------------
// Tests: DeleteMenu
// ---------------------------------------------------------------------------

func TestDeleteMenu_Success(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.getByID = testMenu()

	err := env.svc.DeleteMenu(context.Background(), "menu-1")
	require.NoError(t, err)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionDelete, env.audit.lastEntry.Action)
}

func TestDeleteMenu_NotFound(t *testing.T) {
	env := newTestEnv()
	env.menuRepo.getByIDErr = apperror.NotFound("menu not found", nil)

	err := env.svc.DeleteMenu(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

package domain_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/site/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMenu_ValidInput(t *testing.T) {
	m, err := domain.NewMenu("Main Nav", "main-nav")
	require.NoError(t, err)
	assert.Equal(t, "Main Nav", m.Name)
	assert.Equal(t, "main-nav", m.Slug)
}

func TestNewMenu_EmptyName(t *testing.T) {
	_, err := domain.NewMenu("", "main-nav")
	assert.ErrorIs(t, err, domain.ErrEmptyMenuName)
}

func TestNewMenu_EmptySlug(t *testing.T) {
	_, err := domain.NewMenu("Main Nav", "")
	assert.ErrorIs(t, err, domain.ErrEmptyMenuSlug)
}

func TestNewMenuItem_ValidInput(t *testing.T) {
	item, err := domain.NewMenuItem("menu-1", "Home", "https://example.com", 1, "")
	require.NoError(t, err)
	assert.Equal(t, "Home", item.Label)
	assert.Equal(t, "https://example.com", item.URL)
	assert.Equal(t, 1, item.Order)
	assert.Equal(t, "", item.ParentID)
}

func TestNewMenuItem_EmptyLabel(t *testing.T) {
	_, err := domain.NewMenuItem("menu-1", "", "https://example.com", 1, "")
	assert.ErrorIs(t, err, domain.ErrEmptyMenuItemLabel)
}

func TestNewMenuItem_EmptyURL(t *testing.T) {
	_, err := domain.NewMenuItem("menu-1", "Home", "", 1, "")
	assert.ErrorIs(t, err, domain.ErrEmptyMenuItemURL)
}

func TestMenuItem_IsTopLevel(t *testing.T) {
	item, _ := domain.NewMenuItem("m1", "Home", "/", 1, "")
	assert.True(t, item.IsTopLevel())

	child, _ := domain.NewMenuItem("m1", "Sub", "/sub", 2, "parent-id")
	assert.False(t, child.IsTopLevel())
}

// ---- compile-check for MenuRepository ----

var _ domain.MenuRepository = (*mockMenuRepo)(nil)

type mockMenuRepo struct {
	saveFn       func(ctx context.Context, m *domain.Menu) error
	findByIDFn   func(ctx context.Context, id string) (*domain.Menu, error)
	listFn       func(ctx context.Context) ([]*domain.Menu, error)
	deleteFn     func(ctx context.Context, id string) error
	saveItemFn   func(ctx context.Context, item *domain.MenuItem) error
	listItemsFn  func(ctx context.Context, menuID string) ([]*domain.MenuItem, error)
	deleteItemFn func(ctx context.Context, itemID string) error
}

func (m *mockMenuRepo) Save(ctx context.Context, menu *domain.Menu) error {
	return m.saveFn(ctx, menu)
}
func (m *mockMenuRepo) FindByID(ctx context.Context, id string) (*domain.Menu, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockMenuRepo) List(ctx context.Context) ([]*domain.Menu, error) {
	return m.listFn(ctx)
}
func (m *mockMenuRepo) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}
func (m *mockMenuRepo) SaveItem(ctx context.Context, item *domain.MenuItem) error {
	return m.saveItemFn(ctx, item)
}
func (m *mockMenuRepo) ListItems(ctx context.Context, menuID string) ([]*domain.MenuItem, error) {
	return m.listItemsFn(ctx, menuID)
}
func (m *mockMenuRepo) DeleteItem(ctx context.Context, itemID string) error {
	return m.deleteItemFn(ctx, itemID)
}

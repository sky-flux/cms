package app_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/site/app"
	"github.com/sky-flux/cms/internal/site/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- hand-written mocks ----

type mockSiteRepo struct {
	site   *domain.Site
	upsErr error
}

func (m *mockSiteRepo) GetSite(ctx context.Context) (*domain.Site, error) {
	return m.site, nil
}
func (m *mockSiteRepo) Upsert(ctx context.Context, s *domain.Site) error {
	m.site = s
	return m.upsErr
}

type mockMenuRepo struct {
	saved *domain.Menu
	items []*domain.MenuItem
}

func (m *mockMenuRepo) Save(ctx context.Context, menu *domain.Menu) error {
	m.saved = menu
	return nil
}
func (m *mockMenuRepo) FindByID(ctx context.Context, id string) (*domain.Menu, error) {
	if m.saved != nil && m.saved.ID == id {
		return m.saved, nil
	}
	return nil, nil
}
func (m *mockMenuRepo) List(ctx context.Context) ([]*domain.Menu, error) { return nil, nil }
func (m *mockMenuRepo) Delete(ctx context.Context, id string) error       { return nil }
func (m *mockMenuRepo) SaveItem(ctx context.Context, item *domain.MenuItem) error {
	m.items = append(m.items, item)
	return nil
}
func (m *mockMenuRepo) ListItems(ctx context.Context, menuID string) ([]*domain.MenuItem, error) {
	return m.items, nil
}
func (m *mockMenuRepo) DeleteItem(ctx context.Context, itemID string) error { return nil }

type mockRedirectRepo struct {
	saved *domain.Redirect
}

func (m *mockRedirectRepo) Save(ctx context.Context, r *domain.Redirect) error {
	m.saved = r
	return nil
}
func (m *mockRedirectRepo) FindByPath(ctx context.Context, fromPath string) (*domain.Redirect, error) {
	return nil, nil
}
func (m *mockRedirectRepo) List(ctx context.Context, offset, limit int) ([]*domain.Redirect, int, error) {
	return nil, 0, nil
}
func (m *mockRedirectRepo) Delete(ctx context.Context, id string) error { return nil }

// ---- GetSiteConfig ----

func TestGetSiteConfig_ReturnsSite(t *testing.T) {
	site := &domain.Site{Name: "My Blog", Language: "en", Timezone: "UTC"}
	repo := &mockSiteRepo{site: site}
	uc := app.NewGetSiteConfigUseCase(repo)

	result, err := uc.Execute(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "My Blog", result.Name)
}

func TestGetSiteConfig_NilSiteReturnsDefault(t *testing.T) {
	repo := &mockSiteRepo{site: nil}
	uc := app.NewGetSiteConfigUseCase(repo)

	result, err := uc.Execute(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "My Site", result.Name) // default name
}

// ---- UpdateSiteConfig ----

func TestUpdateSiteConfig_Success(t *testing.T) {
	repo := &mockSiteRepo{site: &domain.Site{Name: "Old", Language: "en", Timezone: "UTC"}}
	uc := app.NewUpdateSiteConfigUseCase(repo)

	err := uc.Execute(context.Background(), app.UpdateSiteConfigInput{
		Name:     "New Name",
		Language: "zh-CN",
		Timezone: "Asia/Shanghai",
	})
	require.NoError(t, err)
	assert.Equal(t, "New Name", repo.site.Name)
}

func TestUpdateSiteConfig_EmptyNameRejected(t *testing.T) {
	repo := &mockSiteRepo{site: &domain.Site{Name: "Old", Language: "en", Timezone: "UTC"}}
	uc := app.NewUpdateSiteConfigUseCase(repo)

	err := uc.Execute(context.Background(), app.UpdateSiteConfigInput{Name: ""})
	assert.ErrorIs(t, err, domain.ErrEmptySiteName)
}

// ---- CreateMenu ----

func TestCreateMenu_Success(t *testing.T) {
	repo := &mockMenuRepo{}
	uc := app.NewCreateMenuUseCase(repo)

	menu, err := uc.Execute(context.Background(), app.CreateMenuInput{
		Name: "Main Nav",
		Slug: "main-nav",
	})
	require.NoError(t, err)
	assert.Equal(t, "Main Nav", menu.Name)
	assert.NotNil(t, repo.saved)
}

func TestCreateMenu_EmptyNameRejected(t *testing.T) {
	repo := &mockMenuRepo{}
	uc := app.NewCreateMenuUseCase(repo)

	_, err := uc.Execute(context.Background(), app.CreateMenuInput{Name: "", Slug: "slug"})
	assert.ErrorIs(t, err, domain.ErrEmptyMenuName)
}

// ---- AddMenuItem ----

func TestAddMenuItem_Success(t *testing.T) {
	repo := &mockMenuRepo{}
	uc := app.NewAddMenuItemUseCase(repo)

	item, err := uc.Execute(context.Background(), app.AddMenuItemInput{
		MenuID: "menu-1",
		Label:  "Home",
		URL:    "/",
		Order:  1,
	})
	require.NoError(t, err)
	assert.Equal(t, "Home", item.Label)
	assert.Len(t, repo.items, 1)
}

func TestAddMenuItem_EmptyLabelRejected(t *testing.T) {
	repo := &mockMenuRepo{}
	uc := app.NewAddMenuItemUseCase(repo)

	_, err := uc.Execute(context.Background(), app.AddMenuItemInput{
		MenuID: "m1",
		Label:  "",
		URL:    "/",
	})
	assert.ErrorIs(t, err, domain.ErrEmptyMenuItemLabel)
}

// ---- CreateRedirect ----

func TestCreateRedirect_Success(t *testing.T) {
	repo := &mockRedirectRepo{}
	uc := app.NewCreateRedirectUseCase(repo)

	r, err := uc.Execute(context.Background(), app.CreateRedirectInput{
		FromPath:   "/old",
		ToPath:     "/new",
		StatusCode: 301,
	})
	require.NoError(t, err)
	assert.Equal(t, "/old", r.FromPath)
	assert.Equal(t, 301, r.StatusCode)
	assert.NotNil(t, repo.saved)
}

func TestCreateRedirect_InvalidPath(t *testing.T) {
	repo := &mockRedirectRepo{}
	uc := app.NewCreateRedirectUseCase(repo)

	_, err := uc.Execute(context.Background(), app.CreateRedirectInput{
		FromPath:   "no-slash",
		ToPath:     "/new",
		StatusCode: 301,
	})
	assert.ErrorIs(t, err, domain.ErrInvalidFromPath)
}

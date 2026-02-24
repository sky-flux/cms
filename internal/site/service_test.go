package site_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/site"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: SiteRepository
// ---------------------------------------------------------------------------

type mockSiteRepo struct {
	listSites    []model.Site
	listTotal    int64
	listErr      error
	getBySlug    *model.Site
	getBySlugErr error
	createErr    error
	updateErr    error
	deleteErr    error
	countActive  int64
	countErr     error
	slugExists   bool
	slugExistsErr error
	domainExists  bool
	domainExistsErr error
}

func (m *mockSiteRepo) List(_ context.Context, _ site.ListFilter) ([]model.Site, int64, error) {
	return m.listSites, m.listTotal, m.listErr
}
func (m *mockSiteRepo) GetBySlug(_ context.Context, _ string) (*model.Site, error) {
	return m.getBySlug, m.getBySlugErr
}
func (m *mockSiteRepo) Create(_ context.Context, s *model.Site) error {
	if m.createErr == nil {
		s.ID = "new-site-id"
		s.CreatedAt = time.Now()
		s.UpdatedAt = time.Now()
	}
	return m.createErr
}
func (m *mockSiteRepo) Update(_ context.Context, _ *model.Site) error { return m.updateErr }
func (m *mockSiteRepo) Delete(_ context.Context, _ string) error      { return m.deleteErr }
func (m *mockSiteRepo) CountActive(_ context.Context) (int64, error)  { return m.countActive, m.countErr }
func (m *mockSiteRepo) SlugExists(_ context.Context, _ string) (bool, error) {
	return m.slugExists, m.slugExistsErr
}
func (m *mockSiteRepo) DomainExists(_ context.Context, _, _ string) (bool, error) {
	return m.domainExists, m.domainExistsErr
}

// ---------------------------------------------------------------------------
// Mock: UserRoleRepository
// ---------------------------------------------------------------------------

type mockUserRoleRepo struct {
	listUsers    []site.UserWithRole
	listTotal    int64
	listErr      error
	assignErr    error
	removeErr    error
	userExists   bool
	userExistsErr error
}

func (m *mockUserRoleRepo) ListUsersWithRoles(_ context.Context, _ site.UserFilter) ([]site.UserWithRole, int64, error) {
	return m.listUsers, m.listTotal, m.listErr
}
func (m *mockUserRoleRepo) AssignRole(_ context.Context, _, _ string) error { return m.assignErr }
func (m *mockUserRoleRepo) RemoveRole(_ context.Context, _ string) error    { return m.removeErr }
func (m *mockUserRoleRepo) UserExists(_ context.Context, _ string) (bool, error) {
	return m.userExists, m.userExistsErr
}

// ---------------------------------------------------------------------------
// Mock: RoleResolver
// ---------------------------------------------------------------------------

type mockRoleResolver struct {
	role *model.Role
	err  error
}

func (m *mockRoleResolver) GetBySlug(_ context.Context, _ string) (*model.Role, error) {
	return m.role, m.err
}

// ---------------------------------------------------------------------------
// Mock: RBACInvalidator
// ---------------------------------------------------------------------------

type mockRBACInvalidator struct {
	err error
}

func (m *mockRBACInvalidator) InvalidateUserCache(_ context.Context, _ string) error {
	return m.err
}

// ---------------------------------------------------------------------------
// Mock: SchemaManager
// ---------------------------------------------------------------------------

type mockSchemaManager struct {
	createErr error
	dropErr   error
}

func (m *mockSchemaManager) Create(_ context.Context, _ string) error { return m.createErr }
func (m *mockSchemaManager) Drop(_ context.Context, _ string) error   { return m.dropErr }

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

type testEnv struct {
	svc         *site.Service
	siteRepo    *mockSiteRepo
	urRepo      *mockUserRoleRepo
	roleRes     *mockRoleResolver
	rbac        *mockRBACInvalidator
	schemaMgr   *mockSchemaManager
}

func newTestEnv() *testEnv {
	sr := &mockSiteRepo{}
	ur := &mockUserRoleRepo{}
	rr := &mockRoleResolver{}
	ri := &mockRBACInvalidator{}
	sm := &mockSchemaManager{}
	return &testEnv{
		svc:       site.NewService(sr, ur, rr, ri, sm),
		siteRepo:  sr,
		urRepo:    ur,
		roleRes:   rr,
		rbac:      ri,
		schemaMgr: sm,
	}
}

func testSite() *model.Site {
	return &model.Site{
		ID:            "site-1",
		Name:          "Test Site",
		Slug:          "test_site",
		Domain:        "test.example.com",
		DefaultLocale: "zh-CN",
		Timezone:      "Asia/Shanghai",
		Status:        model.SiteStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Tests: ListSites
// ---------------------------------------------------------------------------

func TestService_ListSites_Success(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.listSites = []model.Site{*testSite()}
	env.siteRepo.listTotal = 1

	sites, total, err := env.svc.ListSites(context.Background(), site.ListFilter{Page: 1, PerPage: 10})
	require.NoError(t, err)
	assert.Len(t, sites, 1)
	assert.Equal(t, int64(1), total)
}

func TestService_ListSites_DefaultsPagination(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.listSites = nil
	env.siteRepo.listTotal = 0

	// Page=0 and PerPage=0 should be normalized
	_, _, err := env.svc.ListSites(context.Background(), site.ListFilter{Page: 0, PerPage: 0})
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Tests: GetSite
// ---------------------------------------------------------------------------

func TestService_GetSite_Success(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlug = testSite()

	s, err := env.svc.GetSite(context.Background(), "test_site")
	require.NoError(t, err)
	assert.Equal(t, "test_site", s.Slug)
}

func TestService_GetSite_NotFound(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlugErr = apperror.NotFound("site not found", nil)

	_, err := env.svc.GetSite(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: CreateSite
// ---------------------------------------------------------------------------

func TestService_CreateSite_Success(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.slugExists = false

	s, err := env.svc.CreateSite(context.Background(), &site.CreateSiteReq{
		Name: "My Blog",
		Slug: "my_blog",
	})
	require.NoError(t, err)
	assert.Equal(t, "my_blog", s.Slug)
	assert.Equal(t, "zh-CN", s.DefaultLocale)
	assert.Equal(t, "Asia/Shanghai", s.Timezone)
}

func TestService_CreateSite_InvalidSlug(t *testing.T) {
	env := newTestEnv()

	_, err := env.svc.CreateSite(context.Background(), &site.CreateSiteReq{
		Name: "Bad", Slug: "NO-CAPS!",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestService_CreateSite_SlugConflict(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.slugExists = true

	_, err := env.svc.CreateSite(context.Background(), &site.CreateSiteReq{
		Name: "Dup", Slug: "existing_slug",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestService_CreateSite_DomainConflict(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.slugExists = false
	env.siteRepo.domainExists = true

	_, err := env.svc.CreateSite(context.Background(), &site.CreateSiteReq{
		Name: "Site", Slug: "new_site", Domain: "taken.com",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestService_CreateSite_DefaultLocale(t *testing.T) {
	env := newTestEnv()

	s, err := env.svc.CreateSite(context.Background(), &site.CreateSiteReq{
		Name:          "With Locale",
		Slug:          "with_locale",
		DefaultLocale: "en",
		Timezone:      "UTC",
	})
	require.NoError(t, err)
	assert.Equal(t, "en", s.DefaultLocale)
	assert.Equal(t, "UTC", s.Timezone)
}

func TestService_CreateSite_SchemaFail(t *testing.T) {
	env := newTestEnv()
	env.schemaMgr.createErr = errors.New("pg error")

	_, err := env.svc.CreateSite(context.Background(), &site.CreateSiteReq{
		Name: "Fail", Slug: "fail_site",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrInternal))
}

// ---------------------------------------------------------------------------
// Tests: UpdateSite
// ---------------------------------------------------------------------------

func TestService_UpdateSite_Success(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlug = testSite()

	newName := "Updated Name"
	s, err := env.svc.UpdateSite(context.Background(), "test_site", &site.UpdateSiteReq{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", s.Name)
}

func TestService_UpdateSite_DomainConflict(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlug = testSite()
	env.siteRepo.domainExists = true

	newDomain := "taken.com"
	_, err := env.svc.UpdateSite(context.Background(), "test_site", &site.UpdateSiteReq{
		Domain: &newDomain,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestService_UpdateSite_NotFound(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlugErr = apperror.NotFound("site not found", nil)

	newName := "X"
	_, err := env.svc.UpdateSite(context.Background(), "nope", &site.UpdateSiteReq{Name: &newName})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: DeleteSite
// ---------------------------------------------------------------------------

func TestService_DeleteSite_Success(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlug = testSite()
	env.siteRepo.countActive = 2

	err := env.svc.DeleteSite(context.Background(), "test_site", "test_site")
	require.NoError(t, err)
}

func TestService_DeleteSite_ConfirmMismatch(t *testing.T) {
	env := newTestEnv()

	err := env.svc.DeleteSite(context.Background(), "test_site", "wrong_slug")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestService_DeleteSite_LastSite(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlug = testSite()
	env.siteRepo.countActive = 1

	err := env.svc.DeleteSite(context.Background(), "test_site", "test_site")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete the last site")
}

// ---------------------------------------------------------------------------
// Tests: ListSiteUsers
// ---------------------------------------------------------------------------

func TestService_ListSiteUsers_SiteNotFound(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlugErr = apperror.NotFound("site not found", nil)

	_, _, err := env.svc.ListSiteUsers(context.Background(), "nope", site.UserFilter{Page: 1, PerPage: 10})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

func TestService_ListSiteUsers_Success(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlug = testSite()
	env.urRepo.listUsers = []site.UserWithRole{
		{User: model.User{ID: "u1", Email: "a@b.com"}, RoleSlug: "admin"},
	}
	env.urRepo.listTotal = 1

	users, total, err := env.svc.ListSiteUsers(context.Background(), "test_site", site.UserFilter{Page: 1, PerPage: 10})
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, int64(1), total)
}

// ---------------------------------------------------------------------------
// Tests: AssignSiteRole
// ---------------------------------------------------------------------------

func TestService_AssignSiteRole_Success(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlug = testSite()
	env.urRepo.userExists = true
	env.roleRes.role = &model.Role{ID: "role-1", Slug: "editor"}

	err := env.svc.AssignSiteRole(context.Background(), "test_site", "user-1", "editor")
	require.NoError(t, err)
}

func TestService_AssignSiteRole_UserNotFound(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlug = testSite()
	env.urRepo.userExists = false

	err := env.svc.AssignSiteRole(context.Background(), "test_site", "user-1", "editor")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

func TestService_AssignSiteRole_RoleNotFound(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlug = testSite()
	env.urRepo.userExists = true
	env.roleRes.err = apperror.NotFound("role not found", nil)

	err := env.svc.AssignSiteRole(context.Background(), "test_site", "user-1", "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: RemoveSiteRole
// ---------------------------------------------------------------------------

func TestService_RemoveSiteRole_Success(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlug = testSite()

	err := env.svc.RemoveSiteRole(context.Background(), "test_site", "user-1")
	require.NoError(t, err)
}

func TestService_RemoveSiteRole_SiteNotFound(t *testing.T) {
	env := newTestEnv()
	env.siteRepo.getBySlugErr = apperror.NotFound("site not found", nil)

	err := env.svc.RemoveSiteRole(context.Background(), "nope", "user-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

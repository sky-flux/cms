package rbac_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock implementations ---

type mockUserRoleRepo struct {
	slugs    []string
	roles    []model.Role
	slugsErr error
}

func (m *mockUserRoleRepo) GetRolesByUserID(_ context.Context, _ string) ([]model.Role, error) {
	return m.roles, nil
}
func (m *mockUserRoleRepo) GetRoleSlugs(_ context.Context, _ string) ([]string, error) {
	return m.slugs, m.slugsErr
}
func (m *mockUserRoleRepo) SetUserRoles(_ context.Context, _ string, _ []string) error { return nil }
func (m *mockUserRoleRepo) HasRole(_ context.Context, _, _ string) (bool, error)       { return false, nil }

type mockRoleAPIRepo struct {
	apisByRole map[string][]model.APIEndpoint
}

func (m *mockRoleAPIRepo) GetAPIsByRoleID(_ context.Context, roleID string) ([]model.APIEndpoint, error) {
	return m.apisByRole[roleID], nil
}
func (m *mockRoleAPIRepo) SetRoleAPIs(_ context.Context, _ string, _ []string) error { return nil }
func (m *mockRoleAPIRepo) GetRoleIDsByMethodPath(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockRoleAPIRepo) CloneFromTemplate(_ context.Context, _, _ string) error { return nil }

type mockMenuRepo struct {
	menus []model.AdminMenu
}

func (m *mockMenuRepo) ListTree(_ context.Context) ([]model.AdminMenu, error)                   { return nil, nil }
func (m *mockMenuRepo) Create(_ context.Context, _ *model.AdminMenu) error                      { return nil }
func (m *mockMenuRepo) Update(_ context.Context, _ *model.AdminMenu) error                      { return nil }
func (m *mockMenuRepo) Delete(_ context.Context, _ string) error                                 { return nil }
func (m *mockMenuRepo) GetMenusByRoleID(_ context.Context, _ string) ([]model.AdminMenu, error)  { return nil, nil }
func (m *mockMenuRepo) SetRoleMenus(_ context.Context, _ string, _ []string) error               { return nil }
func (m *mockMenuRepo) GetMenusByUserID(_ context.Context, _ string) ([]model.AdminMenu, error) {
	return m.menus, nil
}

// --- Helper ---

func setupTestService(t *testing.T, userRoleRepo rbac.UserRoleRepository, roleAPIRepo rbac.RoleAPIRepository, menuRepo rbac.MenuRepository) (*rbac.Service, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	svc := rbac.NewService(userRoleRepo, roleAPIRepo, menuRepo, rdb)
	return svc, mr
}

// --- Tests ---

func TestService_CheckPermission_SuperRoleBypass(t *testing.T) {
	userRoleRepo := &mockUserRoleRepo{
		slugs: []string{"super"},
		roles: []model.Role{{ID: "role-1", Slug: "super"}},
	}
	roleAPIRepo := &mockRoleAPIRepo{apisByRole: map[string][]model.APIEndpoint{}}

	svc, _ := setupTestService(t, userRoleRepo, roleAPIRepo, &mockMenuRepo{})

	allowed, err := svc.CheckPermission(context.Background(), "user-1", "DELETE", "/api/v1/dangerous")
	require.NoError(t, err)
	assert.True(t, allowed, "super role should bypass all permission checks")
}

func TestService_CheckPermission_MatchingPermission(t *testing.T) {
	userRoleRepo := &mockUserRoleRepo{
		slugs: []string{"editor"},
		roles: []model.Role{{ID: "role-editor", Slug: "editor"}},
	}
	roleAPIRepo := &mockRoleAPIRepo{
		apisByRole: map[string][]model.APIEndpoint{
			"role-editor": {
				{Method: "GET", Path: "/api/v1/posts"},
				{Method: "POST", Path: "/api/v1/posts"},
			},
		},
	}

	svc, _ := setupTestService(t, userRoleRepo, roleAPIRepo, &mockMenuRepo{})

	allowed, err := svc.CheckPermission(context.Background(), "user-2", "GET", "/api/v1/posts")
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestService_CheckPermission_NoPermission(t *testing.T) {
	userRoleRepo := &mockUserRoleRepo{
		slugs: []string{"viewer"},
		roles: []model.Role{{ID: "role-viewer", Slug: "viewer"}},
	}
	roleAPIRepo := &mockRoleAPIRepo{
		apisByRole: map[string][]model.APIEndpoint{
			"role-viewer": {
				{Method: "GET", Path: "/api/v1/posts"},
			},
		},
	}

	svc, _ := setupTestService(t, userRoleRepo, roleAPIRepo, &mockMenuRepo{})

	allowed, err := svc.CheckPermission(context.Background(), "user-3", "DELETE", "/api/v1/posts/:id")
	require.NoError(t, err)
	assert.False(t, allowed, "viewer should not have DELETE permission")
}

func TestService_CheckPermission_CachesUserRoles(t *testing.T) {
	callCount := 0
	userRoleRepo := &mockUserRoleRepo{
		slugs: []string{"editor"},
		roles: []model.Role{{ID: "role-editor", Slug: "editor"}},
	}
	// Wrap to count calls
	countingRepo := &countingUserRoleRepo{inner: userRoleRepo, callCount: &callCount}

	roleAPIRepo := &mockRoleAPIRepo{
		apisByRole: map[string][]model.APIEndpoint{
			"role-editor": {{Method: "GET", Path: "/api/v1/posts"}},
		},
	}

	svc, _ := setupTestService(t, countingRepo, roleAPIRepo, &mockMenuRepo{})
	ctx := context.Background()

	// First call: hits DB
	_, err := svc.CheckPermission(ctx, "user-4", "GET", "/api/v1/posts")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount, "first call should hit DB")

	// Second call: should use cache
	_, err = svc.CheckPermission(ctx, "user-4", "GET", "/api/v1/posts")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount, "second call should use cache, not hit DB again")
}

func TestService_InvalidateUserCache(t *testing.T) {
	userRoleRepo := &mockUserRoleRepo{slugs: []string{"editor"}, roles: []model.Role{{ID: "r1", Slug: "editor"}}}
	roleAPIRepo := &mockRoleAPIRepo{apisByRole: map[string][]model.APIEndpoint{"r1": {{Method: "GET", Path: "/api/v1/posts"}}}}

	svc, mr := setupTestService(t, userRoleRepo, roleAPIRepo, &mockMenuRepo{})
	ctx := context.Background()

	// Populate cache
	_, _ = svc.CheckPermission(ctx, "user-5", "GET", "/api/v1/posts")
	assert.True(t, mr.Exists("user:user-5:roles"), "cache should be populated")

	// Invalidate
	err := svc.InvalidateUserCache(ctx, "user-5")
	require.NoError(t, err)
	assert.False(t, mr.Exists("user:user-5:roles"), "cache should be cleared")
}

func TestService_InvalidateRoleCache(t *testing.T) {
	userRoleRepo := &mockUserRoleRepo{slugs: []string{"editor"}, roles: []model.Role{{ID: "r1", Slug: "editor"}}}
	roleAPIRepo := &mockRoleAPIRepo{apisByRole: map[string][]model.APIEndpoint{"r1": {{Method: "GET", Path: "/api/v1/posts"}}}}

	svc, mr := setupTestService(t, userRoleRepo, roleAPIRepo, &mockMenuRepo{})
	ctx := context.Background()

	// Populate cache by triggering CheckPermission
	_, _ = svc.CheckPermission(ctx, "user-6", "GET", "/api/v1/posts")
	assert.True(t, mr.Exists("role:r1:api_set"), "role API cache should be populated")

	// Invalidate
	err := svc.InvalidateRoleCache(ctx, "r1")
	require.NoError(t, err)
	assert.False(t, mr.Exists("role:r1:api_set"), "role API cache should be cleared")
}

func TestService_GetUserMenuTree(t *testing.T) {
	menuRepo := &mockMenuRepo{
		menus: []model.AdminMenu{
			{ID: "m1", Name: "Dashboard", Path: "/dashboard", SortOrder: 0},
			{ID: "m2", Name: "Posts", Path: "/posts", SortOrder: 1},
		},
	}
	svc, _ := setupTestService(t, &mockUserRoleRepo{slugs: []string{"editor"}}, &mockRoleAPIRepo{}, menuRepo)

	menus, err := svc.GetUserMenuTree(context.Background(), "user-7")
	require.NoError(t, err)
	assert.Len(t, menus, 2)
	assert.Equal(t, "Dashboard", menus[0].Name)
}

// --- Counting wrapper for cache test ---

type countingUserRoleRepo struct {
	inner     *mockUserRoleRepo
	callCount *int
}

func (c *countingUserRoleRepo) GetRolesByUserID(ctx context.Context, userID string) ([]model.Role, error) {
	*c.callCount++
	return c.inner.GetRolesByUserID(ctx, userID)
}
func (c *countingUserRoleRepo) GetRoleSlugs(ctx context.Context, userID string) ([]string, error) {
	*c.callCount++
	return c.inner.GetRoleSlugs(ctx, userID)
}
func (c *countingUserRoleRepo) SetUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	return c.inner.SetUserRoles(ctx, userID, roleIDs)
}
func (c *countingUserRoleRepo) HasRole(ctx context.Context, userID, slug string) (bool, error) {
	return c.inner.HasRole(ctx, userID, slug)
}

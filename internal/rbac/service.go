package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/model"
)

const (
	userRolesTTL = 300 * time.Second // L1: 5 minutes
	roleAPIsTTL  = 600 * time.Second // L2: 10 minutes
	userMenusTTL = 300 * time.Second
)

// cachedUserRoles stores both slugs and role IDs for a user.
type cachedUserRoles struct {
	Slugs   []string `json:"slugs"`
	RoleIDs []string `json:"role_ids"`
}

// Service provides RBAC permission checking with two-level Redis caching.
type Service struct {
	userRoleRepo UserRoleRepository
	roleAPIRepo  RoleAPIRepository
	menuRepo     MenuRepository
	rdb          *redis.Client
}

// NewService creates an RBAC service.
func NewService(
	userRoleRepo UserRoleRepository,
	roleAPIRepo RoleAPIRepository,
	menuRepo MenuRepository,
	rdb *redis.Client,
) *Service {
	return &Service{
		userRoleRepo: userRoleRepo,
		roleAPIRepo:  roleAPIRepo,
		menuRepo:     menuRepo,
		rdb:          rdb,
	}
}

// CheckPermission verifies if user can access method+path.
//
// Flow:
//  1. Get user's roles (L1 cache or DB)
//  2. If "super" in slugs → return true (short circuit)
//  3. For each role, get API set (L2 cache or DB)
//  4. Union all API sets, check if method:path matches
func (s *Service) CheckPermission(ctx context.Context, userID, method, path string) (bool, error) {
	// Step 1: Get user roles (L1 cache)
	cached, err := s.getUserRoles(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("check permission get user roles: %w", err)
	}

	// Step 2: Super role bypass
	if slices.Contains(cached.Slugs, "super") {
		return true, nil
	}

	// Step 3+4: Check each role's API set
	target := method + ":" + path
	for _, roleID := range cached.RoleIDs {
		apiSet, err := s.getRoleAPISet(ctx, roleID)
		if err != nil {
			slog.Error("check permission get role api set", "error", err, "role_id", roleID)
			continue
		}
		if slices.Contains(apiSet, target) {
			return true, nil
		}
	}

	return false, nil
}

// GetUserMenuTree returns merged menu tree for user's roles.
func (s *Service) GetUserMenuTree(ctx context.Context, userID string) ([]model.AdminMenu, error) {
	// Check menu cache
	cacheKey := fmt.Sprintf("user:%s:menu_tree", userID)
	cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var menus []model.AdminMenu
		if json.Unmarshal(cached, &menus) == nil {
			return menus, nil
		}
	}

	// Cache miss — query DB
	menus, err := s.menuRepo.GetMenusByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user menu tree: %w", err)
	}

	// Set cache
	data, err := json.Marshal(menus)
	if err == nil {
		s.rdb.Set(ctx, cacheKey, data, userMenusTTL)
	}

	return menus, nil
}

// InvalidateUserCache clears L1 cache for a user (on role change).
func (s *Service) InvalidateUserCache(ctx context.Context, userID string) error {
	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("user:%s:roles", userID))
	pipe.Del(ctx, fmt.Sprintf("user:%s:menu_tree", userID))
	_, err := pipe.Exec(ctx)
	return err
}

// InvalidateRoleCache clears L2 cache for a role (on permission change).
func (s *Service) InvalidateRoleCache(ctx context.Context, roleID string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("role:%s:api_set", roleID)).Err()
}

// getUserRoles retrieves user roles from L1 cache or DB.
func (s *Service) getUserRoles(ctx context.Context, userID string) (*cachedUserRoles, error) {
	cacheKey := fmt.Sprintf("user:%s:roles", userID)

	// Try L1 cache
	cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var result cachedUserRoles
		if json.Unmarshal(cached, &result) == nil {
			return &result, nil
		}
	}

	// Cache miss — query DB
	roles, err := s.userRoleRepo.GetRolesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user roles from db: %w", err)
	}

	result := &cachedUserRoles{
		Slugs:   make([]string, len(roles)),
		RoleIDs: make([]string, len(roles)),
	}
	for i, r := range roles {
		result.Slugs[i] = r.Slug
		result.RoleIDs[i] = r.ID
	}

	// Set L1 cache
	data, err := json.Marshal(result)
	if err == nil {
		s.rdb.Set(ctx, cacheKey, data, userRolesTTL)
	}

	return result, nil
}

// getRoleAPISet retrieves a role's API permission set from L2 cache or DB.
// Returns a set of "METHOD:/path" strings.
func (s *Service) getRoleAPISet(ctx context.Context, roleID string) ([]string, error) {
	cacheKey := fmt.Sprintf("role:%s:api_set", roleID)

	// Try L2 cache
	cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var apiSet []string
		if json.Unmarshal(cached, &apiSet) == nil {
			return apiSet, nil
		}
	}

	// Cache miss — query DB
	apis, err := s.roleAPIRepo.GetAPIsByRoleID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("get role api set from db: %w", err)
	}

	apiSet := make([]string, len(apis))
	for i, api := range apis {
		apiSet[i] = api.Method + ":" + api.Path
	}

	// Set L2 cache
	data, err := json.Marshal(apiSet)
	if err == nil {
		s.rdb.Set(ctx, cacheKey, data, roleAPIsTTL)
	}

	return apiSet, nil
}

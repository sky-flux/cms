package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/mail"
	"github.com/sky-flux/cms/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: UserRepository
// ---------------------------------------------------------------------------

type mockUserRepo struct {
	listUsers  []model.User
	listTotal  int64
	listErr    error
	getByID    *model.User
	getByIDErr error
	getByEmail    *model.User
	getByEmailErr error
	createErr  error
	updateErr  error
	deleteErr  error
}

func (m *mockUserRepo) List(_ context.Context, _ user.ListFilter) ([]model.User, int64, error) {
	return m.listUsers, m.listTotal, m.listErr
}
func (m *mockUserRepo) GetByID(_ context.Context, _ string) (*model.User, error) {
	return m.getByID, m.getByIDErr
}
func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*model.User, error) {
	return m.getByEmail, m.getByEmailErr
}
func (m *mockUserRepo) Create(_ context.Context, u *model.User) error {
	if m.createErr == nil {
		u.ID = "new-user-id"
		u.CreatedAt = time.Now()
		u.UpdatedAt = time.Now()
	}
	return m.createErr
}
func (m *mockUserRepo) Update(_ context.Context, _ *model.User) error { return m.updateErr }
func (m *mockUserRepo) SoftDelete(_ context.Context, _ string) error  { return m.deleteErr }

// ---------------------------------------------------------------------------
// Mock: RoleRepository
// ---------------------------------------------------------------------------

type mockRoleRepo struct {
	role *model.Role
	err  error
}

func (m *mockRoleRepo) GetBySlug(_ context.Context, _ string) (*model.Role, error) {
	return m.role, m.err
}

// ---------------------------------------------------------------------------
// Mock: UserRoleRepository
// ---------------------------------------------------------------------------

type mockUserRoleRepo struct {
	assignErr        error
	roleSlug         string
	roleSlugErr      error
	countActive      int64
	countActiveErr   error
}

func (m *mockUserRoleRepo) Assign(_ context.Context, _, _ string) error { return m.assignErr }
func (m *mockUserRoleRepo) GetRoleSlug(_ context.Context, _ string) (string, error) {
	return m.roleSlug, m.roleSlugErr
}
func (m *mockUserRoleRepo) CountActiveByRoleSlug(_ context.Context, _ string) (int64, error) {
	return m.countActive, m.countActiveErr
}

// ---------------------------------------------------------------------------
// Mock: TokenRevoker
// ---------------------------------------------------------------------------

type mockTokenRevoker struct {
	err error
}

func (m *mockTokenRevoker) RevokeAllForUser(_ context.Context, _ string) error { return m.err }

// ---------------------------------------------------------------------------
// Mock: AuditLogger
// ---------------------------------------------------------------------------

type mockAuditLogger struct {
	err error
}

func (m *mockAuditLogger) Log(_ context.Context, _ audit.Entry) error { return m.err }

// ---------------------------------------------------------------------------
// Test environment
// ---------------------------------------------------------------------------

type testEnv struct {
	svc          *user.Service
	userRepo     *mockUserRepo
	roleRepo     *mockRoleRepo
	urRepo       *mockUserRoleRepo
	tokenRevoker *mockTokenRevoker
	auditLog     *mockAuditLogger
	mailer       *mail.NoopSender
}

func newTestEnv() *testEnv {
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

func testUser() *model.User {
	return &model.User{
		ID:           "user-1",
		Email:        "test@example.com",
		PasswordHash: "$2a$12$hashedpassword",
		DisplayName:  "Test User",
		Status:       model.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Tests: ListUsers
// ---------------------------------------------------------------------------

func TestService_ListUsers_Success(t *testing.T) {
	env := newTestEnv()
	env.userRepo.listUsers = []model.User{*testUser()}
	env.userRepo.listTotal = 1
	env.urRepo.roleSlug = "admin"

	users, total, err := env.svc.ListUsers(context.Background(), user.ListFilter{Page: 1, PerPage: 10})
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "admin", users[0].Role)
}

func TestService_ListUsers_DefaultsPagination(t *testing.T) {
	env := newTestEnv()
	env.userRepo.listUsers = nil
	env.userRepo.listTotal = 0

	_, _, err := env.svc.ListUsers(context.Background(), user.ListFilter{Page: 0, PerPage: 0})
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Tests: GetUser
// ---------------------------------------------------------------------------

func TestService_GetUser_Success(t *testing.T) {
	env := newTestEnv()
	env.userRepo.getByID = testUser()
	env.urRepo.roleSlug = "editor"

	u, err := env.svc.GetUser(context.Background(), "user-1")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", u.Email)
	assert.Equal(t, "editor", u.Role)
}

func TestService_GetUser_NotFound(t *testing.T) {
	env := newTestEnv()
	env.userRepo.getByIDErr = apperror.NotFound("user not found", nil)

	_, err := env.svc.GetUser(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: CreateUser
// ---------------------------------------------------------------------------

func TestService_CreateUser_Success(t *testing.T) {
	env := newTestEnv()
	env.roleRepo.role = &model.Role{ID: "role-1", Slug: "editor"}
	env.userRepo.getByEmailErr = apperror.NotFound("not found", nil)

	resp, err := env.svc.CreateUser(context.Background(), &user.CreateUserReq{
		Email:       "new@example.com",
		Password:    "password123",
		DisplayName: "New User",
		Role:        "editor",
	})
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", resp.Email)
	assert.Equal(t, "editor", resp.Role)
}

func TestService_CreateUser_InvalidRole(t *testing.T) {
	env := newTestEnv()
	env.roleRepo.err = apperror.NotFound("role not found", nil)

	_, err := env.svc.CreateUser(context.Background(), &user.CreateUserReq{
		Email:       "new@example.com",
		Password:    "password123",
		DisplayName: "New User",
		Role:        "nonexistent",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

func TestService_CreateUser_DuplicateEmail(t *testing.T) {
	env := newTestEnv()
	env.roleRepo.role = &model.Role{ID: "role-1", Slug: "editor"}
	env.userRepo.getByEmail = testUser()

	_, err := env.svc.CreateUser(context.Background(), &user.CreateUserReq{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Dup User",
		Role:        "editor",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

// ---------------------------------------------------------------------------
// Tests: UpdateUser
// ---------------------------------------------------------------------------

func TestService_UpdateUser_Success(t *testing.T) {
	env := newTestEnv()
	env.userRepo.getByID = testUser()
	env.urRepo.roleSlug = "admin"

	newName := "Updated Name"
	resp, err := env.svc.UpdateUser(context.Background(), "user-1", &user.UpdateUserReq{
		DisplayName: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", resp.DisplayName)
}

func TestService_UpdateUser_NotFound(t *testing.T) {
	env := newTestEnv()
	env.userRepo.getByIDErr = apperror.NotFound("user not found", nil)

	newName := "X"
	_, err := env.svc.UpdateUser(context.Background(), "nope", &user.UpdateUserReq{DisplayName: &newName})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

func TestService_UpdateUser_ChangeRole(t *testing.T) {
	env := newTestEnv()
	env.userRepo.getByID = testUser()
	env.roleRepo.role = &model.Role{ID: "role-2", Slug: "admin"}

	newRole := "admin"
	resp, err := env.svc.UpdateUser(context.Background(), "user-1", &user.UpdateUserReq{
		Role: &newRole,
	})
	require.NoError(t, err)
	assert.Equal(t, "admin", resp.Role)
}

func TestService_UpdateUser_DisableUser(t *testing.T) {
	env := newTestEnv()
	env.userRepo.getByID = testUser()
	env.urRepo.roleSlug = "editor"

	disabled := model.UserStatusDisabled
	resp, err := env.svc.UpdateUser(context.Background(), "user-1", &user.UpdateUserReq{
		Status: &disabled,
	})
	require.NoError(t, err)
	assert.Equal(t, model.UserStatusDisabled, resp.Status)
}

// ---------------------------------------------------------------------------
// Tests: DeleteUser
// ---------------------------------------------------------------------------

func TestService_DeleteUser_Success(t *testing.T) {
	env := newTestEnv()
	env.userRepo.getByID = testUser()
	env.urRepo.roleSlug = "editor"

	ctx := context.WithValue(context.Background(), "user_id", "other-user")
	err := env.svc.DeleteUser(ctx, "user-1")
	require.NoError(t, err)
}

func TestService_DeleteUser_SelfDelete(t *testing.T) {
	env := newTestEnv()

	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	err := env.svc.DeleteUser(ctx, "user-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrForbidden))
}

func TestService_DeleteUser_LastSuperAdmin(t *testing.T) {
	env := newTestEnv()
	env.userRepo.getByID = testUser()
	env.urRepo.roleSlug = "super"
	env.urRepo.countActive = 1

	ctx := context.WithValue(context.Background(), "user_id", "other-user")
	err := env.svc.DeleteUser(ctx, "user-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrForbidden))
	assert.Contains(t, err.Error(), "last super admin")
}

func TestService_DeleteUser_SuperAdminWithOthers(t *testing.T) {
	env := newTestEnv()
	env.userRepo.getByID = testUser()
	env.urRepo.roleSlug = "super"
	env.urRepo.countActive = 3

	ctx := context.WithValue(context.Background(), "user_id", "other-user")
	err := env.svc.DeleteUser(ctx, "user-1")
	require.NoError(t, err)
}

func TestService_DeleteUser_NotFound(t *testing.T) {
	env := newTestEnv()
	env.userRepo.getByIDErr = apperror.NotFound("user not found", nil)

	ctx := context.WithValue(context.Background(), "user_id", "other-user")
	err := env.svc.DeleteUser(ctx, "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/auth"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/crypto"
	"github.com/sky-flux/cms/internal/pkg/jwt"
	"github.com/sky-flux/cms/internal/pkg/mail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test constants
// ---------------------------------------------------------------------------

const (
	testJWTSecret = "test-secret-key-at-least-32-bytes!"
	// 64 hex characters = 32 bytes for AES-256 encryption key
	testTOTPEncKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	testUserID     = "00000000-0000-0000-0000-000000000001"
	testUserEmail  = "alice@example.com"
	testPassword   = "Str0ngP@ss!"
	testIP         = "127.0.0.1"
	testUserAgent  = "TestAgent/1.0"
)

// ---------------------------------------------------------------------------
// Mock: UserRepository
// ---------------------------------------------------------------------------

type mockUserRepo struct {
	byEmail       *model.User
	byEmailErr    error
	byID          *model.User
	byIDErr       error
	updatePwdErr  error
	lastLoginCall bool
}

func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*model.User, error) {
	return m.byEmail, m.byEmailErr
}
func (m *mockUserRepo) GetByID(_ context.Context, _ string) (*model.User, error) {
	return m.byID, m.byIDErr
}
func (m *mockUserRepo) UpdatePassword(_ context.Context, _, _ string) error {
	return m.updatePwdErr
}
func (m *mockUserRepo) UpdateLastLogin(_ context.Context, _ string) error {
	m.lastLoginCall = true
	return nil
}

// ---------------------------------------------------------------------------
// Mock: TokenRepository
// ---------------------------------------------------------------------------

type mockTokenRepo struct {
	createRTErr       error
	getRTByHash       *model.RefreshToken
	getRTByHashErr    error
	revokeRTErr       error
	revokeAllErr      error
	createPRTErr      error
	getPRTByHash      *model.PasswordResetToken
	getPRTByHashErr   error
	markPRTUsedErr    error
	revokedTokenIDs   []string
	revokedAllUserIDs []string
}

func (m *mockTokenRepo) CreateRefreshToken(_ context.Context, _ *model.RefreshToken) error {
	return m.createRTErr
}
func (m *mockTokenRepo) GetRefreshTokenByHash(_ context.Context, _ string) (*model.RefreshToken, error) {
	return m.getRTByHash, m.getRTByHashErr
}
func (m *mockTokenRepo) RevokeRefreshToken(_ context.Context, id string) error {
	m.revokedTokenIDs = append(m.revokedTokenIDs, id)
	return m.revokeRTErr
}
func (m *mockTokenRepo) RevokeAllUserTokens(_ context.Context, userID string) error {
	m.revokedAllUserIDs = append(m.revokedAllUserIDs, userID)
	return m.revokeAllErr
}
func (m *mockTokenRepo) CreatePasswordResetToken(_ context.Context, _ *model.PasswordResetToken) error {
	return m.createPRTErr
}
func (m *mockTokenRepo) GetPasswordResetTokenByHash(_ context.Context, _ string) (*model.PasswordResetToken, error) {
	return m.getPRTByHash, m.getPRTByHashErr
}
func (m *mockTokenRepo) MarkPasswordResetTokenUsed(_ context.Context, _ string) error {
	return m.markPRTUsedErr
}

// ---------------------------------------------------------------------------
// Mock: TOTPRepository
// ---------------------------------------------------------------------------

type mockTOTPRepo struct {
	getByUserID        *model.UserTOTP
	getByUserIDErr     error
	upsertErr          error
	enableErr          error
	deleteErr          error
	updateBackupErr    error
	lastBackupCodes    []string
	lastEnabledID      string
	lastDeletedUserID  string
	lastUpsertedRecord *model.UserTOTP
}

func (m *mockTOTPRepo) GetByUserID(_ context.Context, _ string) (*model.UserTOTP, error) {
	return m.getByUserID, m.getByUserIDErr
}
func (m *mockTOTPRepo) Upsert(_ context.Context, t *model.UserTOTP) error {
	m.lastUpsertedRecord = t
	return m.upsertErr
}
func (m *mockTOTPRepo) Enable(_ context.Context, id string) error {
	m.lastEnabledID = id
	return m.enableErr
}
func (m *mockTOTPRepo) Delete(_ context.Context, userID string) error {
	m.lastDeletedUserID = userID
	return m.deleteErr
}
func (m *mockTOTPRepo) UpdateBackupCodes(_ context.Context, _ string, codes []string) error {
	m.lastBackupCodes = codes
	return m.updateBackupErr
}

// ---------------------------------------------------------------------------
// Mock: RoleLoader
// ---------------------------------------------------------------------------

type mockRoleLoader struct {
	roles []model.Role
	err   error
}

func (m *mockRoleLoader) GetUserRoles(_ context.Context, _ string) ([]model.Role, error) {
	return m.roles, m.err
}

// ---------------------------------------------------------------------------
// Mock: SiteLoader
// ---------------------------------------------------------------------------

type mockSiteLoader struct {
	sites []model.Site
	err   error
}

func (m *mockSiteLoader) GetUserSites(_ context.Context, _ string) ([]model.Site, error) {
	return m.sites, m.err
}

// ---------------------------------------------------------------------------
// Mock: mail.Sender
// ---------------------------------------------------------------------------

type mockMailer struct {
	sent    []mail.Message
	sendErr error
}

func (m *mockMailer) Send(_ context.Context, msg mail.Message) error {
	m.sent = append(m.sent, msg)
	return m.sendErr
}

// ---------------------------------------------------------------------------
// Test harness
// ---------------------------------------------------------------------------

type testHarness struct {
	svc        *auth.Service
	userRepo   *mockUserRepo
	tokenRepo  *mockTokenRepo
	totpRepo   *mockTOTPRepo
	roleLoader *mockRoleLoader
	siteLoader *mockSiteLoader
	mailer     *mockMailer
	jwtMgr     *jwt.Manager
	rdb        *redis.Client
	mr         *miniredis.Miniredis
}

func newHarness(t *testing.T) *testHarness {
	return newHarnessWithMailer(t, nil)
}

func newHarnessWithMailer(t *testing.T, mailer *mockMailer) *testHarness {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	jwtMgr := jwt.NewManager(testJWTSecret, 15*time.Minute, 5*time.Minute, rdb)

	userRepo := &mockUserRepo{}
	tokenRepo := &mockTokenRepo{}
	totpRepo := &mockTOTPRepo{}
	roleLoader := &mockRoleLoader{}
	siteLoader := &mockSiteLoader{}

	cfg := auth.ServiceConfig{
		TOTPEncryptionKey: testTOTPEncKey,
		AccessExpiry:      15 * time.Minute,
		RefreshExpiry:     7 * 24 * time.Hour,
	}
	if mailer != nil {
		cfg.Mailer = mailer
		cfg.SiteName = "Test Site"
		cfg.FrontendURL = "http://localhost:3000"
	}

	svc := auth.NewService(
		userRepo,
		tokenRepo,
		totpRepo,
		roleLoader,
		siteLoader,
		jwtMgr,
		rdb,
		cfg,
	)

	return &testHarness{
		svc:        svc,
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		totpRepo:   totpRepo,
		roleLoader: roleLoader,
		siteLoader: siteLoader,
		mailer:     mailer,
		jwtMgr:     jwtMgr,
		rdb:        rdb,
		mr:         mr,
	}
}

// makeUser creates a model.User with a real bcrypt hash for testPassword.
func makeUser(t *testing.T) *model.User {
	t.Helper()
	hash, err := crypto.HashPassword(testPassword)
	require.NoError(t, err)
	now := time.Now()
	return &model.User{
		ID:           testUserID,
		Email:        testUserEmail,
		PasswordHash: hash,
		DisplayName:  "Alice",
		Status:       model.UserStatusActive,
		LastLoginAt:  &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// =========================================================================
// Login tests
// =========================================================================

func TestLogin_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byEmail = user
	h.totpRepo.getByUserIDErr = apperror.NotFound("not found", nil)

	loginResp, twoFAResp, err := h.svc.Login(context.Background(),
		&auth.LoginReq{Email: testUserEmail, Password: testPassword},
		testIP, testUserAgent,
	)

	require.NoError(t, err)
	assert.Nil(t, twoFAResp)
	require.NotNil(t, loginResp)
	assert.Equal(t, testUserID, loginResp.User.ID)
	assert.Equal(t, testUserEmail, loginResp.User.Email)
	assert.Equal(t, "Alice", loginResp.User.DisplayName)
	assert.Equal(t, "Bearer", loginResp.TokenType)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, loginResp.RawRefreshToken)
	assert.Equal(t, 900, loginResp.ExpiresIn) // 15 min
	assert.True(t, h.userRepo.lastLoginCall)
}

func TestLogin_WrongPassword(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byEmail = user

	_, _, err := h.svc.Login(context.Background(),
		&auth.LoginReq{Email: testUserEmail, Password: "wrongpassword"},
		testIP, testUserAgent,
	)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "invalid email or password")
}

func TestLogin_UserNotFound(t *testing.T) {
	h := newHarness(t)
	h.userRepo.byEmailErr = apperror.NotFound("user not found", nil)

	_, _, err := h.svc.Login(context.Background(),
		&auth.LoginReq{Email: "nobody@example.com", Password: testPassword},
		testIP, testUserAgent,
	)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "invalid email or password")
}

func TestLogin_InactiveUser(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	user.Status = model.UserStatusDisabled
	h.userRepo.byEmail = user

	_, _, err := h.svc.Login(context.Background(),
		&auth.LoginReq{Email: testUserEmail, Password: testPassword},
		testIP, testUserAgent,
	)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "account is disabled")
}

func TestLogin_LockedOut(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byEmail = user

	// Simulate 5 failed attempts already stored in Redis
	ctx := context.Background()
	h.rdb.Set(ctx, "login_fail:"+testUserEmail, "5", 15*time.Minute)

	_, _, err := h.svc.Login(ctx,
		&auth.LoginReq{Email: testUserEmail, Password: testPassword},
		testIP, testUserAgent,
	)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, 429, appErr.Code)
	assert.Contains(t, appErr.Message, "temporarily locked")
}

func TestLogin_2FARequired(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byEmail = user
	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:        "totp-1",
		UserID:    testUserID,
		Enabled: model.ToggleYes,
	}

	loginResp, twoFAResp, err := h.svc.Login(context.Background(),
		&auth.LoginReq{Email: testUserEmail, Password: testPassword},
		testIP, testUserAgent,
	)

	require.NoError(t, err)
	assert.Nil(t, loginResp)
	require.NotNil(t, twoFAResp)
	assert.NotEmpty(t, twoFAResp.TempToken)
	assert.Equal(t, "Bearer", twoFAResp.TokenType)
	assert.Equal(t, 300, twoFAResp.ExpiresIn)
	assert.Equal(t, "totp", twoFAResp.Requires)
}

func TestLogin_WrongPassword_IncrementsLockout(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byEmail = user

	ctx := context.Background()
	_, _, _ = h.svc.Login(ctx,
		&auth.LoginReq{Email: testUserEmail, Password: "wrong1"},
		testIP, testUserAgent,
	)

	val, err := h.rdb.Get(ctx, "login_fail:"+testUserEmail).Int()
	require.NoError(t, err)
	assert.Equal(t, 1, val)
}

func TestLogin_Success_ResetsLockoutCounter(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byEmail = user
	h.totpRepo.getByUserIDErr = apperror.NotFound("not found", nil)

	ctx := context.Background()
	// Set 3 prior failed attempts
	h.rdb.Set(ctx, "login_fail:"+testUserEmail, "3", 15*time.Minute)

	_, _, err := h.svc.Login(ctx,
		&auth.LoginReq{Email: testUserEmail, Password: testPassword},
		testIP, testUserAgent,
	)

	require.NoError(t, err)
	// The lockout key should be deleted
	exists := h.rdb.Exists(ctx, "login_fail:"+testUserEmail).Val()
	assert.Equal(t, int64(0), exists)
}

// =========================================================================
// Refresh tests
// =========================================================================

func TestRefresh_Success(t *testing.T) {
	h := newHarness(t)
	raw, hash, err := crypto.GenerateToken(32)
	require.NoError(t, err)

	h.tokenRepo.getRTByHash = &model.RefreshToken{
		ID:        "rt-1",
		UserID:    testUserID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	resp, newRaw, err := h.svc.Refresh(context.Background(), raw, testIP, testUserAgent)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, 900, resp.ExpiresIn)
	assert.NotEmpty(t, newRaw)
	assert.NotEqual(t, raw, newRaw) // New token should differ
	// Verify old token was revoked
	assert.Contains(t, h.tokenRepo.revokedTokenIDs, "rt-1")
}

func TestRefresh_InvalidToken(t *testing.T) {
	h := newHarness(t)
	h.tokenRepo.getRTByHashErr = apperror.NotFound("not found", nil)

	resp, newRaw, err := h.svc.Refresh(context.Background(), "bad-token", testIP, testUserAgent)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, newRaw)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "invalid or expired refresh token")
}

func TestRefresh_ExpiredToken(t *testing.T) {
	h := newHarness(t)
	// Repository returns not found for expired tokens
	h.tokenRepo.getRTByHashErr = apperror.NotFound("expired", nil)

	_, _, err := h.svc.Refresh(context.Background(), "expired-token", testIP, testUserAgent)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "invalid or expired refresh token")
}

// =========================================================================
// Logout tests
// =========================================================================

func TestLogout_Success_BlacklistsJTI(t *testing.T) {
	h := newHarness(t)

	// Create a valid access token to get a JTI
	token, err := h.jwtMgr.SignAccessToken(testUserID)
	require.NoError(t, err)
	claims, err := h.jwtMgr.Verify(token)
	require.NoError(t, err)

	// Create a refresh token for revocation
	raw, hash, err := crypto.GenerateToken(32)
	require.NoError(t, err)
	h.tokenRepo.getRTByHash = &model.RefreshToken{
		ID:        "rt-logout",
		UserID:    testUserID,
		TokenHash: hash,
	}

	err = h.svc.Logout(context.Background(), claims.JTI, raw)
	require.NoError(t, err)

	// Verify JTI is blacklisted
	blacklisted, err := h.jwtMgr.IsBlacklisted(context.Background(), claims.JTI)
	require.NoError(t, err)
	assert.True(t, blacklisted)

	// Verify refresh token was revoked
	assert.Contains(t, h.tokenRepo.revokedTokenIDs, "rt-logout")
}

func TestLogout_EmptyRefreshToken(t *testing.T) {
	h := newHarness(t)

	token, err := h.jwtMgr.SignAccessToken(testUserID)
	require.NoError(t, err)
	claims, err := h.jwtMgr.Verify(token)
	require.NoError(t, err)

	err = h.svc.Logout(context.Background(), claims.JTI, "")
	require.NoError(t, err)

	// JTI should still be blacklisted
	blacklisted, err := h.jwtMgr.IsBlacklisted(context.Background(), claims.JTI)
	require.NoError(t, err)
	assert.True(t, blacklisted)

	// No refresh token revocation should have happened
	assert.Empty(t, h.tokenRepo.revokedTokenIDs)
}

// =========================================================================
// Me tests
// =========================================================================

func TestMe_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user
	h.roleLoader.roles = []model.Role{
		{ID: "role-1", Name: "Admin", Slug: "admin"},
	}
	h.siteLoader.sites = []model.Site{
		{ID: "site-1", Name: "My Blog", Slug: "my-blog"},
	}

	resp, err := h.svc.Me(context.Background(), testUserID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, testUserID, resp.ID)
	assert.Equal(t, testUserEmail, resp.Email)
	assert.Equal(t, "Alice", resp.DisplayName)
	assert.Equal(t, model.UserStatusActive, resp.Status)
	require.Len(t, resp.Roles, 1)
	assert.Equal(t, "admin", resp.Roles[0].Slug)
	require.Len(t, resp.Sites, 1)
	assert.Equal(t, "my-blog", resp.Sites[0].Slug)
}

func TestMe_UserNotFound(t *testing.T) {
	h := newHarness(t)
	h.userRepo.byIDErr = apperror.NotFound("user not found", nil)

	resp, err := h.svc.Me(context.Background(), testUserID)

	require.Error(t, err)
	assert.Nil(t, resp)
}

// =========================================================================
// ChangePassword tests
// =========================================================================

func TestChangePassword_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user

	err := h.svc.ChangePassword(context.Background(), testUserID,
		&auth.ChangePasswordReq{
			CurrentPassword: testPassword,
			NewPassword:     "NewStr0ng!Pass",
		},
	)

	require.NoError(t, err)
	// Verify all tokens were revoked after password change
	assert.Contains(t, h.tokenRepo.revokedAllUserIDs, testUserID)
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user

	err := h.svc.ChangePassword(context.Background(), testUserID,
		&auth.ChangePasswordReq{
			CurrentPassword: "wrongcurrent",
			NewPassword:     "NewStr0ng!Pass",
		},
	)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "current password is incorrect")
}

func TestChangePassword_UserNotFound(t *testing.T) {
	h := newHarness(t)
	h.userRepo.byIDErr = apperror.NotFound("not found", nil)

	err := h.svc.ChangePassword(context.Background(), testUserID,
		&auth.ChangePasswordReq{
			CurrentPassword: testPassword,
			NewPassword:     "NewStr0ng!Pass",
		},
	)

	require.Error(t, err)
}

// =========================================================================
// ForgotPassword tests
// =========================================================================

func TestForgotPassword_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byEmail = user

	err := h.svc.ForgotPassword(context.Background(),
		&auth.ForgotPasswordReq{Email: testUserEmail},
	)

	// Always returns nil for security
	assert.NoError(t, err)
}

func TestForgotPassword_UnknownEmail_StillReturnsNil(t *testing.T) {
	h := newHarness(t)
	h.userRepo.byEmailErr = apperror.NotFound("not found", nil)

	err := h.svc.ForgotPassword(context.Background(),
		&auth.ForgotPasswordReq{Email: "unknown@example.com"},
	)

	// Must return nil even when email not found
	assert.NoError(t, err)
}

// =========================================================================
// ResetPassword tests
// =========================================================================

func TestResetPassword_Success(t *testing.T) {
	h := newHarness(t)

	raw, hash, err := crypto.GenerateToken(32)
	require.NoError(t, err)

	h.tokenRepo.getPRTByHash = &model.PasswordResetToken{
		ID:        "prt-1",
		UserID:    testUserID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	err = h.svc.ResetPassword(context.Background(),
		&auth.ResetPasswordReq{
			Token:       raw,
			NewPassword: "NewStr0ng!Pass",
		},
	)

	require.NoError(t, err)
	assert.Contains(t, h.tokenRepo.revokedAllUserIDs, testUserID)
}

func TestResetPassword_InvalidToken(t *testing.T) {
	h := newHarness(t)
	h.tokenRepo.getPRTByHashErr = apperror.NotFound("not found", nil)

	err := h.svc.ResetPassword(context.Background(),
		&auth.ResetPasswordReq{
			Token:       "invalid-token",
			NewPassword: "NewStr0ng!Pass",
		},
	)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "invalid or expired reset token")
}

// =========================================================================
// Setup2FA tests
// =========================================================================

func TestSetup2FA_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user
	// No existing TOTP record
	h.totpRepo.getByUserIDErr = apperror.NotFound("not found", nil)

	resp, err := h.svc.Setup2FA(context.Background(), testUserID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Secret)
	assert.NotEmpty(t, resp.QRCodeURI)
	assert.Len(t, resp.BackupCodes, 10)
	// Verify the TOTP record was upserted with Enabled=ToggleNo
	require.NotNil(t, h.totpRepo.lastUpsertedRecord)
	assert.Equal(t, model.ToggleNo, h.totpRepo.lastUpsertedRecord.Enabled)
	assert.Equal(t, testUserID, h.totpRepo.lastUpsertedRecord.UserID)
}

func TestSetup2FA_AlreadyEnabled(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user
	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:        "totp-1",
		UserID:    testUserID,
		Enabled: model.ToggleYes,
	}

	resp, err := h.svc.Setup2FA(context.Background(), testUserID)

	require.Error(t, err)
	assert.Nil(t, resp)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "2FA is already enabled")
}

func TestSetup2FA_NotYetEnabled_AllowsReSetup(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user
	// TOTP exists but is NOT enabled — user can re-setup
	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:        "totp-1",
		UserID:    testUserID,
		Enabled: model.ToggleNo,
	}

	resp, err := h.svc.Setup2FA(context.Background(), testUserID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Secret)
}

// =========================================================================
// Verify2FA tests
// =========================================================================

func TestVerify2FA_Success(t *testing.T) {
	h := newHarness(t)

	// Generate a real TOTP key, encrypt it, and store it
	key, err := crypto.GenerateTOTPKey(testUserEmail, "Sky Flux CMS")
	require.NoError(t, err)
	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), testTOTPEncKey)
	require.NoError(t, err)

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleNo, // not yet enabled
	}

	// Generate a valid TOTP code from the secret
	code, err := totp.GenerateCode(key.Secret(), time.Now())
	require.NoError(t, err)

	err = h.svc.Verify2FA(context.Background(), testUserID, &auth.Verify2FAReq{Code: code})

	require.NoError(t, err)
	assert.Equal(t, "totp-1", h.totpRepo.lastEnabledID)
}

func TestVerify2FA_InvalidCode(t *testing.T) {
	h := newHarness(t)

	key, err := crypto.GenerateTOTPKey(testUserEmail, "Sky Flux CMS")
	require.NoError(t, err)
	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), testTOTPEncKey)
	require.NoError(t, err)

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleNo,
	}

	err = h.svc.Verify2FA(context.Background(), testUserID, &auth.Verify2FAReq{Code: "000000"})

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "invalid TOTP code")
}

func TestVerify2FA_AlreadyEnabled(t *testing.T) {
	h := newHarness(t)

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:        "totp-1",
		UserID:    testUserID,
		Enabled: model.ToggleYes,
	}

	err := h.svc.Verify2FA(context.Background(), testUserID, &auth.Verify2FAReq{Code: "123456"})

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "already verified")
}

func TestVerify2FA_NoSetup(t *testing.T) {
	h := newHarness(t)
	h.totpRepo.getByUserIDErr = apperror.NotFound("not found", nil)

	err := h.svc.Verify2FA(context.Background(), testUserID, &auth.Verify2FAReq{Code: "123456"})

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "setup not found")
}

// =========================================================================
// Validate2FA tests (login 2FA challenge)
// =========================================================================

func TestValidate2FA_SuccessWithTOTPCode(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user

	key, err := crypto.GenerateTOTPKey(testUserEmail, "Sky Flux CMS")
	require.NoError(t, err)
	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), testTOTPEncKey)
	require.NoError(t, err)

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleYes,
	}

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	require.NoError(t, err)

	resp, err := h.svc.Validate2FA(context.Background(), testUserID,
		&auth.Validate2FAReq{Code: code}, testIP, testUserAgent,
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, testUserID, resp.User.ID)
}

func TestValidate2FA_SuccessWithBackupCode(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user

	key, err := crypto.GenerateTOTPKey(testUserEmail, "Sky Flux CMS")
	require.NoError(t, err)
	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), testTOTPEncKey)
	require.NoError(t, err)

	// Create a known backup code and hash it
	backupCode := "ABCD-1234"
	backupHash := crypto.HashToken(backupCode)
	otherHash := crypto.HashToken("WXYZ-5678")

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleYes,
		BackupCodesHash: []string{otherHash, backupHash},
	}

	resp, err := h.svc.Validate2FA(context.Background(), testUserID,
		&auth.Validate2FAReq{Code: backupCode}, testIP, testUserAgent,
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, testUserID, resp.User.ID)
	// Verify the used backup code was removed (only otherHash remains)
	require.Len(t, h.totpRepo.lastBackupCodes, 1)
	assert.Equal(t, otherHash, h.totpRepo.lastBackupCodes[0])
}

func TestValidate2FA_InvalidCode(t *testing.T) {
	h := newHarness(t)

	key, err := crypto.GenerateTOTPKey(testUserEmail, "Sky Flux CMS")
	require.NoError(t, err)
	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), testTOTPEncKey)
	require.NoError(t, err)

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleYes,
		BackupCodesHash: []string{crypto.HashToken("REAL-CODE")},
	}

	resp, err := h.svc.Validate2FA(context.Background(), testUserID,
		&auth.Validate2FAReq{Code: "WRONG-CODE"}, testIP, testUserAgent,
	)

	require.Error(t, err)
	assert.Nil(t, resp)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "invalid 2FA code")
}

func TestValidate2FA_TOTPNotConfigured(t *testing.T) {
	h := newHarness(t)
	h.totpRepo.getByUserIDErr = apperror.NotFound("not configured", nil)

	resp, err := h.svc.Validate2FA(context.Background(), testUserID,
		&auth.Validate2FAReq{Code: "123456"}, testIP, testUserAgent,
	)

	require.Error(t, err)
	assert.Nil(t, resp)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "2FA not configured")
}

func TestValidate2FA_ReplayDetection(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user

	key, err := crypto.GenerateTOTPKey(testUserEmail, "Sky Flux CMS")
	require.NoError(t, err)
	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), testTOTPEncKey)
	require.NoError(t, err)

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleYes,
	}

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	require.NoError(t, err)

	// First use succeeds
	resp, err := h.svc.Validate2FA(context.Background(), testUserID,
		&auth.Validate2FAReq{Code: code}, testIP, testUserAgent,
	)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Second use of same code fails (replay)
	resp, err = h.svc.Validate2FA(context.Background(), testUserID,
		&auth.Validate2FAReq{Code: code}, testIP, testUserAgent,
	)
	require.Error(t, err)
	assert.Nil(t, resp)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "already used")
}

func TestValidate2FA_RateLimited(t *testing.T) {
	h := newHarness(t)

	key, err := crypto.GenerateTOTPKey(testUserEmail, "Sky Flux CMS")
	require.NoError(t, err)
	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), testTOTPEncKey)
	require.NoError(t, err)

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleYes,
		BackupCodesHash: []string{},
	}

	ctx := context.Background()
	// Simulate 5 failed attempts already in Redis
	h.rdb.Set(ctx, "2fa:attempts:"+testUserID, "5", 5*time.Minute)

	resp, err := h.svc.Validate2FA(ctx, testUserID,
		&auth.Validate2FAReq{Code: "000000"}, testIP, testUserAgent,
	)

	require.Error(t, err)
	assert.Nil(t, resp)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, 429, appErr.Code)
	assert.Contains(t, appErr.Message, "too many 2FA attempts")
}

func TestValidate2FA_FailureIncrementsAttemptCounter(t *testing.T) {
	h := newHarness(t)

	key, err := crypto.GenerateTOTPKey(testUserEmail, "Sky Flux CMS")
	require.NoError(t, err)
	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), testTOTPEncKey)
	require.NoError(t, err)

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleYes,
		BackupCodesHash: []string{},
	}

	ctx := context.Background()
	_, _ = h.svc.Validate2FA(ctx, testUserID,
		&auth.Validate2FAReq{Code: "WRONG-CODE"}, testIP, testUserAgent,
	)

	val, err := h.rdb.Get(ctx, "2fa:attempts:"+testUserID).Int()
	require.NoError(t, err)
	assert.Equal(t, 1, val)
}

func TestVerify2FA_ReplayDetection(t *testing.T) {
	h := newHarness(t)

	key, err := crypto.GenerateTOTPKey(testUserEmail, "Sky Flux CMS")
	require.NoError(t, err)
	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), testTOTPEncKey)
	require.NoError(t, err)

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleNo,
	}

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	require.NoError(t, err)

	// First use succeeds
	err = h.svc.Verify2FA(context.Background(), testUserID, &auth.Verify2FAReq{Code: code})
	require.NoError(t, err)

	// Reset mock so second call doesn't hit "already enabled"
	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleNo,
	}

	// Second use of same code fails (replay)
	err = h.svc.Verify2FA(context.Background(), testUserID, &auth.Verify2FAReq{Code: code})
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "already used")
}

// =========================================================================
// Disable2FA tests
// =========================================================================

func TestDisable2FA_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user

	key, err := crypto.GenerateTOTPKey(testUserEmail, "Sky Flux CMS")
	require.NoError(t, err)
	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), testTOTPEncKey)
	require.NoError(t, err)

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleYes,
	}

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	require.NoError(t, err)

	err = h.svc.Disable2FA(context.Background(), testUserID,
		&auth.Disable2FAReq{Password: testPassword, Code: code},
	)

	require.NoError(t, err)
	assert.Equal(t, testUserID, h.totpRepo.lastDeletedUserID)
	assert.Contains(t, h.tokenRepo.revokedAllUserIDs, testUserID)
}

func TestDisable2FA_WrongPassword(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user

	err := h.svc.Disable2FA(context.Background(), testUserID,
		&auth.Disable2FAReq{Password: "wrongpassword", Code: "123456"},
	)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "invalid password")
}

func TestDisable2FA_InvalidTOTPCode(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user

	key, err := crypto.GenerateTOTPKey(testUserEmail, "Sky Flux CMS")
	require.NoError(t, err)
	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), testTOTPEncKey)
	require.NoError(t, err)

	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:              "totp-1",
		UserID:          testUserID,
		SecretEncrypted: encrypted,
		Enabled:         model.ToggleYes,
	}

	err = h.svc.Disable2FA(context.Background(), testUserID,
		&auth.Disable2FAReq{Password: testPassword, Code: "000000"},
	)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "invalid TOTP code")
}

func TestDisable2FA_NotEnabled(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user
	h.totpRepo.getByUserIDErr = apperror.NotFound("not found", nil)

	err := h.svc.Disable2FA(context.Background(), testUserID,
		&auth.Disable2FAReq{Password: testPassword, Code: "123456"},
	)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "2FA is not enabled")
}

// =========================================================================
// RegenerateBackupCodes tests
// =========================================================================

func TestRegenerateBackupCodes_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user
	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:        "totp-1",
		UserID:    testUserID,
		Enabled: model.ToggleYes,
	}

	resp, err := h.svc.RegenerateBackupCodes(context.Background(), testUserID,
		&auth.RegenerateBackupCodesReq{Password: testPassword},
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.BackupCodes, 10)
	// Verify backup codes were updated
	require.NotNil(t, h.totpRepo.lastBackupCodes)
	assert.Len(t, h.totpRepo.lastBackupCodes, 10)
}

func TestRegenerateBackupCodes_WrongPassword(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user

	resp, err := h.svc.RegenerateBackupCodes(context.Background(), testUserID,
		&auth.RegenerateBackupCodesReq{Password: "wrongpassword"},
	)

	require.Error(t, err)
	assert.Nil(t, resp)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "invalid password")
}

func TestRegenerateBackupCodes_TOTPNotEnabled(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user
	h.totpRepo.getByUserIDErr = apperror.NotFound("not found", nil)

	resp, err := h.svc.RegenerateBackupCodes(context.Background(), testUserID,
		&auth.RegenerateBackupCodesReq{Password: testPassword},
	)

	require.Error(t, err)
	assert.Nil(t, resp)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Contains(t, appErr.Message, "2FA is not enabled")
}

// =========================================================================
// Get2FAStatus tests
// =========================================================================

func TestGet2FAStatus_Enabled(t *testing.T) {
	h := newHarness(t)
	verifiedAt := time.Now().Add(-24 * time.Hour)
	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:         "totp-1",
		UserID:     testUserID,
		Enabled:    model.ToggleYes,
		VerifiedAt: &verifiedAt,
	}

	resp, err := h.svc.Get2FAStatus(context.Background(), testUserID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Enabled)
	assert.NotNil(t, resp.VerifiedAt)
}

func TestGet2FAStatus_NotSetup(t *testing.T) {
	h := newHarness(t)
	h.totpRepo.getByUserIDErr = apperror.NotFound("not found", nil)

	resp, err := h.svc.Get2FAStatus(context.Background(), testUserID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Enabled)
	assert.Nil(t, resp.VerifiedAt)
}

func TestGet2FAStatus_SetupButNotEnabled(t *testing.T) {
	h := newHarness(t)
	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:        "totp-1",
		UserID:    testUserID,
		Enabled: model.ToggleNo,
	}

	resp, err := h.svc.Get2FAStatus(context.Background(), testUserID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Enabled)
}

func TestGet2FAStatus_NonAppError_ReturnsError(t *testing.T) {
	h := newHarness(t)
	// Return a plain error (not *apperror.AppError)
	h.totpRepo.getByUserIDErr = errors.New("database connection error")

	resp, err := h.svc.Get2FAStatus(context.Background(), testUserID)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "database connection error")
}

// =========================================================================
// ForceDisable2FA tests
// =========================================================================

func TestForceDisable2FA_Success(t *testing.T) {
	h := newHarness(t)

	err := h.svc.ForceDisable2FA(context.Background(), testUserID, "admin-1", "user requested reset")

	require.NoError(t, err)
	assert.Equal(t, testUserID, h.totpRepo.lastDeletedUserID)
	assert.Contains(t, h.tokenRepo.revokedAllUserIDs, testUserID)
}

func TestForceDisable2FA_DeleteFails(t *testing.T) {
	h := newHarness(t)
	h.totpRepo.deleteErr = errors.New("db error")

	err := h.svc.ForceDisable2FA(context.Background(), testUserID, "admin-1", "reset")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// =========================================================================
// ForgotPassword email notification tests
// =========================================================================

func TestForgotPassword_SendsResetEmail(t *testing.T) {
	mailer := &mockMailer{}
	h := newHarnessWithMailer(t, mailer)
	user := makeUser(t)
	h.userRepo.byEmail = user

	err := h.svc.ForgotPassword(context.Background(),
		&auth.ForgotPasswordReq{Email: testUserEmail},
	)

	require.NoError(t, err)
	// Email is sent async — wait briefly for goroutine
	time.Sleep(100 * time.Millisecond)

	require.Len(t, mailer.sent, 1)
	assert.Equal(t, testUserEmail, mailer.sent[0].To)
	assert.Contains(t, mailer.sent[0].Subject, "Password Reset")
	assert.Contains(t, mailer.sent[0].HTML, "http://localhost:3000")
}

func TestForgotPassword_NoMailerSkipsEmail(t *testing.T) {
	h := newHarness(t) // no mailer set
	user := makeUser(t)
	h.userRepo.byEmail = user

	err := h.svc.ForgotPassword(context.Background(),
		&auth.ForgotPasswordReq{Email: testUserEmail},
	)

	// Should still succeed, just not send email
	require.NoError(t, err)
}

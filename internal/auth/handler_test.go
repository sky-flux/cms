package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/auth"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Response parsing helpers
// ---------------------------------------------------------------------------

type jsonResp struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

func parseResp(t *testing.T, w *httptest.ResponseRecorder) jsonResp {
	t.Helper()
	var resp jsonResp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "failed to parse response body: %s", w.Body.String())
	return resp
}

func jsonBody(t *testing.T, v interface{}) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

// ---------------------------------------------------------------------------
// Router setup helper
// ---------------------------------------------------------------------------

func setupHandlerRouter(t *testing.T, h *testHarness) (*gin.Engine, *auth.Handler) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler := auth.NewHandler(h.svc, 7*24*time.Hour)
	r := gin.New()

	// Public auth routes
	authPublic := r.Group("/api/v1/auth")
	authPublic.POST("/login", handler.Login)
	authPublic.POST("/refresh", handler.Refresh)
	authPublic.POST("/forgot-password", handler.ForgotPassword)
	authPublic.POST("/reset-password", handler.ResetPassword)

	// 2FA validate uses temp token — inject user_id to simulate purpose middleware
	authPublic.POST("/2fa/validate", func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	}, handler.Validate2FA)

	// Protected routes — simulate auth middleware by setting context values
	authProtected := r.Group("/api/v1/auth")
	authProtected.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Set("token_jti", "test-jti")
		c.Next()
	})
	authProtected.POST("/logout", handler.Logout)
	authProtected.GET("/me", handler.Me)
	authProtected.PUT("/password", handler.ChangePassword)
	authProtected.POST("/2fa/setup", handler.Setup2FA)
	authProtected.POST("/2fa/verify", handler.Verify2FA)
	authProtected.POST("/2fa/disable", handler.Disable2FA)
	authProtected.POST("/2fa/backup-codes", handler.RegenerateBackupCodes)
	authProtected.GET("/2fa/status", handler.Get2FAStatus)

	// Admin route
	authProtected.DELETE("/2fa/users/:user_id", handler.ForceDisable2FA)

	return r, handler
}

// =========================================================================
// Login handler tests
// =========================================================================

func TestHandlerLogin_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byEmail = user
	h.totpRepo.getByUserIDErr = apperror.NotFound("not found", nil)

	r, _ := setupHandlerRouter(t, h)

	body := jsonBody(t, map[string]string{
		"email":    testUserEmail,
		"password": testPassword,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp.Success)
	assert.Empty(t, resp.Error)

	// Parse data payload
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.NotEmpty(t, data["access_token"])
	assert.Equal(t, "Bearer", data["token_type"])
	assert.NotNil(t, data["user"])

	userMap := data["user"].(map[string]interface{})
	assert.Equal(t, testUserID, userMap["id"])
	assert.Equal(t, testUserEmail, userMap["email"])
	assert.Equal(t, "Alice", userMap["display_name"])

	// Check Set-Cookie header for refresh_token
	cookies := w.Result().Cookies()
	var refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}
	require.NotNil(t, refreshCookie, "refresh_token cookie should be set")
	assert.NotEmpty(t, refreshCookie.Value)
	assert.True(t, refreshCookie.HttpOnly)
	assert.True(t, refreshCookie.Secure)
}

func TestHandlerLogin_InvalidBody(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	tests := []struct {
		name string
		body interface{}
	}{
		{
			name: "missing all fields",
			body: map[string]string{},
		},
		{
			name: "missing password",
			body: map[string]string{"email": testUserEmail},
		},
		{
			name: "invalid email format",
			body: map[string]string{"email": "not-an-email", "password": testPassword},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/auth/login", jsonBody(t, tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			resp := parseResp(t, w)
			assert.False(t, resp.Success)
			assert.NotEmpty(t, resp.Error)
		})
	}
}

func TestHandlerLogin_WrongPassword(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byEmail = user

	r, _ := setupHandlerRouter(t, h)

	body := jsonBody(t, map[string]string{
		"email":    testUserEmail,
		"password": "wrongpassword",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResp(t, w)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "invalid email or password")
}

func TestHandlerLogin_2FARequired(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byEmail = user
	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:        "totp-1",
		UserID:    testUserID,
		Enabled: model.ToggleYes,
	}

	r, _ := setupHandlerRouter(t, h)

	body := jsonBody(t, map[string]string{
		"email":    testUserEmail,
		"password": testPassword,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp.Success)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.NotEmpty(t, data["temp_token"])
	assert.Equal(t, "Bearer", data["token_type"])
	assert.Equal(t, float64(300), data["expires_in"])
	assert.Equal(t, "totp", data["requires"])

	// Should NOT set refresh_token cookie for 2FA challenge
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		assert.NotEqual(t, "refresh_token", c.Name, "refresh_token cookie should not be set for 2FA challenge")
	}
}

// =========================================================================
// Refresh handler tests
// =========================================================================

func TestHandlerRefresh_Success(t *testing.T) {
	h := newHarness(t)
	raw, hash, err := crypto.GenerateToken(32)
	require.NoError(t, err)

	h.tokenRepo.getRTByHash = &model.RefreshToken{
		ID:        "rt-1",
		UserID:    testUserID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	r, _ := setupHandlerRouter(t, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: raw})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp.Success)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.NotEmpty(t, data["access_token"])
	assert.Equal(t, "Bearer", data["token_type"])

	// Check new refresh_token cookie is set
	cookies := w.Result().Cookies()
	var refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}
	require.NotNil(t, refreshCookie, "new refresh_token cookie should be set")
	assert.NotEmpty(t, refreshCookie.Value)
}

func TestHandlerRefresh_NoCookie(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
	// No cookie set
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResp(t, w)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "missing refresh token")
}

func TestHandlerRefresh_InvalidToken(t *testing.T) {
	h := newHarness(t)
	h.tokenRepo.getRTByHashErr = apperror.NotFound("not found", nil)

	r, _ := setupHandlerRouter(t, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "bad-token"})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResp(t, w)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "invalid or expired refresh token")
}

// =========================================================================
// Logout handler tests
// =========================================================================

func TestHandlerLogout_Success(t *testing.T) {
	h := newHarness(t)

	// Set up a valid refresh token so the service can find and revoke it
	raw, hash, err := crypto.GenerateToken(32)
	require.NoError(t, err)
	h.tokenRepo.getRTByHash = &model.RefreshToken{
		ID:        "rt-logout",
		UserID:    testUserID,
		TokenHash: hash,
	}

	r, _ := setupHandlerRouter(t, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: raw})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify Set-Cookie clears the refresh_token
	cookies := w.Result().Cookies()
	var refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}
	require.NotNil(t, refreshCookie, "refresh_token cookie should be present to clear it")
	assert.True(t, refreshCookie.MaxAge < 0, "MaxAge should be negative to clear cookie")
}

func TestHandlerLogout_NoCookie(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
	// No refresh_token cookie — logout should still succeed
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// =========================================================================
// Me handler tests
// =========================================================================

func TestHandlerMe_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user
	h.roleLoader.roles = []model.Role{
		{ID: "role-1", Name: "Admin", Slug: "admin"},
	}
	h.siteLoader.sites = []model.Site{
		{ID: "site-1", Name: "My Blog", Slug: "my-blog"},
	}

	r, _ := setupHandlerRouter(t, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp.Success)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, testUserID, data["id"])
	assert.Equal(t, testUserEmail, data["email"])
	assert.Equal(t, "Alice", data["display_name"])
	assert.Equal(t, float64(model.UserStatusActive), data["status"])

	roles := data["roles"].([]interface{})
	require.Len(t, roles, 1)
	roleMap := roles[0].(map[string]interface{})
	assert.Equal(t, "admin", roleMap["slug"])

	sites := data["sites"].([]interface{})
	require.Len(t, sites, 1)
	siteMap := sites[0].(map[string]interface{})
	assert.Equal(t, "my-blog", siteMap["slug"])
}

func TestHandlerMe_UserNotFound(t *testing.T) {
	h := newHarness(t)
	h.userRepo.byIDErr = apperror.NotFound("user not found", nil)

	r, _ := setupHandlerRouter(t, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResp(t, w)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "user not found")
}

// =========================================================================
// ChangePassword handler tests
// =========================================================================

func TestHandlerChangePassword_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user

	r, _ := setupHandlerRouter(t, h)

	body := jsonBody(t, map[string]string{
		"current_password": testPassword,
		"new_password":     "NewStr0ng!Pass",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/auth/password", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandlerChangePassword_InvalidBody(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	tests := []struct {
		name string
		body interface{}
	}{
		{
			name: "missing all fields",
			body: map[string]string{},
		},
		{
			name: "missing new_password",
			body: map[string]string{"current_password": testPassword},
		},
		{
			name: "new_password too short",
			body: map[string]string{"current_password": testPassword, "new_password": "short"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/api/v1/auth/password", jsonBody(t, tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			resp := parseResp(t, w)
			assert.False(t, resp.Success)
			assert.NotEmpty(t, resp.Error)
		})
	}
}

func TestHandlerChangePassword_WrongCurrentPassword(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user

	r, _ := setupHandlerRouter(t, h)

	body := jsonBody(t, map[string]string{
		"current_password": "wrongcurrent",
		"new_password":     "NewStr0ng!Pass",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/auth/password", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResp(t, w)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "current password is incorrect")
}

// =========================================================================
// ForgotPassword handler tests
// =========================================================================

func TestHandlerForgotPassword_AlwaysOK(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		setupRepo func(h *testHarness)
	}{
		{
			name:  "existing email",
			email: testUserEmail,
			setupRepo: func(h *testHarness) {
				h.userRepo.byEmail = makeUser(nil) // nil t is fine for makeUser helper scenario below
			},
		},
		{
			name:  "non-existing email",
			email: "unknown@example.com",
			setupRepo: func(h *testHarness) {
				h.userRepo.byEmailErr = apperror.NotFound("not found", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			if tt.name == "existing email" {
				user := makeUser(t)
				h.userRepo.byEmail = user
			} else {
				h.userRepo.byEmailErr = apperror.NotFound("not found", nil)
			}

			r, _ := setupHandlerRouter(t, h)

			body := jsonBody(t, map[string]string{"email": tt.email})
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/auth/forgot-password", body)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			resp := parseResp(t, w)
			assert.True(t, resp.Success)
		})
	}
}

func TestHandlerForgotPassword_InvalidBody(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	body := jsonBody(t, map[string]string{"email": "not-an-email"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/forgot-password", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	resp := parseResp(t, w)
	assert.False(t, resp.Success)
}

// =========================================================================
// ResetPassword handler tests
// =========================================================================

func TestHandlerResetPassword_Success(t *testing.T) {
	h := newHarness(t)

	raw, hash, err := crypto.GenerateToken(32)
	require.NoError(t, err)

	h.tokenRepo.getPRTByHash = &model.PasswordResetToken{
		ID:        "prt-1",
		UserID:    testUserID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	r, _ := setupHandlerRouter(t, h)

	body := jsonBody(t, map[string]string{
		"token":        raw,
		"new_password": "NewStr0ng!Pass",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/reset-password", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResp(t, w)
	assert.True(t, resp.Success)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Contains(t, data["message"], "password has been reset")
}

func TestHandlerResetPassword_InvalidToken(t *testing.T) {
	h := newHarness(t)
	h.tokenRepo.getPRTByHashErr = apperror.NotFound("not found", nil)

	r, _ := setupHandlerRouter(t, h)

	body := jsonBody(t, map[string]string{
		"token":        "invalid-token",
		"new_password": "NewStr0ng!Pass",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/reset-password", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResp(t, w)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "invalid or expired reset token")
}

func TestHandlerResetPassword_InvalidBody(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	tests := []struct {
		name string
		body interface{}
	}{
		{
			name: "missing token",
			body: map[string]string{"new_password": "NewStr0ng!Pass"},
		},
		{
			name: "missing new_password",
			body: map[string]string{"token": "some-token"},
		},
		{
			name: "new_password too short",
			body: map[string]string{"token": "some-token", "new_password": "short"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/auth/reset-password", jsonBody(t, tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			resp := parseResp(t, w)
			assert.False(t, resp.Success)
		})
	}
}

// =========================================================================
// Setup2FA handler tests
// =========================================================================

func TestHandlerSetup2FA_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user
	h.totpRepo.getByUserIDErr = apperror.NotFound("not found", nil)

	r, _ := setupHandlerRouter(t, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/2fa/setup", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp.Success)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.NotEmpty(t, data["secret"])
	assert.NotEmpty(t, data["qr_code_uri"])

	backupCodes, ok := data["backup_codes"].([]interface{})
	require.True(t, ok, "backup_codes should be an array")
	assert.Len(t, backupCodes, 10)
}

func TestHandlerSetup2FA_AlreadyEnabled(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user
	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:        "totp-1",
		UserID:    testUserID,
		Enabled: model.ToggleYes,
	}

	r, _ := setupHandlerRouter(t, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/2fa/setup", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	resp := parseResp(t, w)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "2FA is already enabled")
}

// =========================================================================
// Get2FAStatus handler tests
// =========================================================================

func TestHandlerGet2FAStatus_NotSetup(t *testing.T) {
	h := newHarness(t)
	h.totpRepo.getByUserIDErr = apperror.NotFound("not found", nil)

	r, _ := setupHandlerRouter(t, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/auth/2fa/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp.Success)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, false, data["enabled"])
}

func TestHandlerGet2FAStatus_Enabled(t *testing.T) {
	h := newHarness(t)
	verifiedAt := time.Now().Add(-24 * time.Hour)
	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:         "totp-1",
		UserID:     testUserID,
		Enabled:    model.ToggleYes,
		VerifiedAt: &verifiedAt,
	}

	r, _ := setupHandlerRouter(t, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/auth/2fa/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp.Success)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, true, data["enabled"])
	assert.NotNil(t, data["verified_at"])
}

// =========================================================================
// ForceDisable2FA handler tests
// =========================================================================

func TestHandlerForceDisable2FA_Success(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	targetUserID := "00000000-0000-0000-0000-000000000099"
	body := jsonBody(t, map[string]string{
		"reason": "user requested reset via support",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/auth/2fa/users/"+targetUserID, body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, targetUserID, h.totpRepo.lastDeletedUserID)
}

func TestHandlerForceDisable2FA_MissingReason(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	targetUserID := "00000000-0000-0000-0000-000000000099"
	body := jsonBody(t, map[string]string{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/auth/2fa/users/"+targetUserID, body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	resp := parseResp(t, w)
	assert.False(t, resp.Success)
}

// =========================================================================
// Verify2FA handler tests
// =========================================================================

func TestHandlerVerify2FA_InvalidBody(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	tests := []struct {
		name string
		body interface{}
	}{
		{
			name: "missing code",
			body: map[string]string{},
		},
		{
			name: "code wrong length",
			body: map[string]string{"code": "12345"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/auth/2fa/verify", jsonBody(t, tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			resp := parseResp(t, w)
			assert.False(t, resp.Success)
		})
	}
}

// =========================================================================
// Validate2FA handler tests
// =========================================================================

func TestHandlerValidate2FA_InvalidBody(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	body := jsonBody(t, map[string]string{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/2fa/validate", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	resp := parseResp(t, w)
	assert.False(t, resp.Success)
}

// =========================================================================
// Disable2FA handler tests
// =========================================================================

func TestHandlerDisable2FA_InvalidBody(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	tests := []struct {
		name string
		body interface{}
	}{
		{
			name: "missing all fields",
			body: map[string]string{},
		},
		{
			name: "missing code",
			body: map[string]string{"password": testPassword},
		},
		{
			name: "code wrong length",
			body: map[string]string{"password": testPassword, "code": "12345"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/auth/2fa/disable", jsonBody(t, tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			resp := parseResp(t, w)
			assert.False(t, resp.Success)
		})
	}
}

// =========================================================================
// RegenerateBackupCodes handler tests
// =========================================================================

func TestHandlerRegenerateBackupCodes_InvalidBody(t *testing.T) {
	h := newHarness(t)
	r, _ := setupHandlerRouter(t, h)

	body := jsonBody(t, map[string]string{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/2fa/backup-codes", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	resp := parseResp(t, w)
	assert.False(t, resp.Success)
}

func TestHandlerRegenerateBackupCodes_Success(t *testing.T) {
	h := newHarness(t)
	user := makeUser(t)
	h.userRepo.byID = user
	h.totpRepo.getByUserID = &model.UserTOTP{
		ID:        "totp-1",
		UserID:    testUserID,
		Enabled: model.ToggleYes,
	}

	r, _ := setupHandlerRouter(t, h)

	body := jsonBody(t, map[string]string{"password": testPassword})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/2fa/backup-codes", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResp(t, w)
	assert.True(t, resp.Success)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	backupCodes, ok := data["backup_codes"].([]interface{})
	require.True(t, ok, "backup_codes should be an array")
	assert.Len(t, backupCodes, 10)
}

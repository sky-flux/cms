package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler provides HTTP endpoints for authentication.
type Handler struct {
	svc           *Service
	refreshExpiry time.Duration
}

// NewHandler creates an auth handler.
func NewHandler(svc *Service, refreshExpiry time.Duration) *Handler {
	return &Handler{svc: svc, refreshExpiry: refreshExpiry}
}

// setRefreshTokenCookie writes or clears the refresh_token httpOnly cookie.
func setRefreshTokenCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("refresh_token", token, maxAge, "/api/v1/auth", "", true, true)
}

// Login authenticates a user and returns tokens or a 2FA challenge.
func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	loginResp, login2FAResp, err := h.svc.Login(
		c.Request.Context(), &req, c.ClientIP(), c.GetHeader("User-Agent"),
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	if loginResp != nil {
		setRefreshTokenCookie(c, loginResp.RawRefreshToken, int(h.refreshExpiry.Seconds()))
		response.Success(c, loginResp)
		return
	}

	response.Success(c, login2FAResp)
}

// Refresh issues a new access token using the refresh_token cookie.
func (h *Handler) Refresh(c *gin.Context) {
	rawRefresh, err := c.Cookie("refresh_token")
	if err != nil {
		response.Error(c, apperror.Unauthorized("missing refresh token", nil))
		return
	}

	resp, newRawRefresh, err := h.svc.Refresh(
		c.Request.Context(), rawRefresh, c.ClientIP(), c.GetHeader("User-Agent"),
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	setRefreshTokenCookie(c, newRawRefresh, int(h.refreshExpiry.Seconds()))
	response.Success(c, resp)
}

// Logout blacklists the access token and revokes the refresh token.
func (h *Handler) Logout(c *gin.Context) {
	jti := c.GetString("token_jti")
	rawRefresh, _ := c.Cookie("refresh_token")

	if err := h.svc.Logout(c.Request.Context(), jti, rawRefresh); err != nil {
		response.Error(c, err)
		return
	}

	setRefreshTokenCookie(c, "", -1)
	response.NoContent(c)
}

// Me returns the current user's profile.
func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("user_id")

	resp, err := h.svc.Me(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// ChangePassword updates the current user's password.
func (h *Handler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// ForgotPassword sends a password reset email. Always returns 200 for security.
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	_ = h.svc.ForgotPassword(c.Request.Context(), &req)
	response.Success(c, gin.H{"message": "if the email exists, a reset link has been sent"})
}

// ResetPassword validates a reset token and updates the password.
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "password has been reset successfully"})
}

// Setup2FA generates a new TOTP secret and backup codes.
func (h *Handler) Setup2FA(c *gin.Context) {
	userID := c.GetString("user_id")

	resp, err := h.svc.Setup2FA(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, resp)
}

// Verify2FA confirms the user's TOTP setup, enabling 2FA.
func (h *Handler) Verify2FA(c *gin.Context) {
	userID := c.GetString("user_id")

	var req Verify2FAReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.Verify2FA(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "2FA has been enabled successfully"})
}

// Validate2FA validates a TOTP code during the login 2FA challenge.
func (h *Handler) Validate2FA(c *gin.Context) {
	userID := c.GetString("user_id")

	var req Validate2FAReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	resp, err := h.svc.Validate2FA(
		c.Request.Context(), userID, &req, c.ClientIP(), c.GetHeader("User-Agent"),
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	setRefreshTokenCookie(c, resp.RawRefreshToken, int(h.refreshExpiry.Seconds()))
	response.Success(c, resp)
}

// Disable2FA disables 2FA after verifying password and TOTP code.
func (h *Handler) Disable2FA(c *gin.Context) {
	userID := c.GetString("user_id")

	var req Disable2FAReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.Disable2FA(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// RegenerateBackupCodes generates new backup codes after password verification.
func (h *Handler) RegenerateBackupCodes(c *gin.Context) {
	userID := c.GetString("user_id")

	var req RegenerateBackupCodesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	resp, err := h.svc.RegenerateBackupCodes(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// Get2FAStatus returns the current 2FA status for the authenticated user.
func (h *Handler) Get2FAStatus(c *gin.Context) {
	userID := c.GetString("user_id")

	resp, err := h.svc.Get2FAStatus(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// ForceDisable2FA allows an admin to disable 2FA for any user.
func (h *Handler) ForceDisable2FA(c *gin.Context) {
	adminID := c.GetString("user_id")
	targetUserID := c.Param("user_id")

	var req ForceDisable2FAReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.ForceDisable2FA(c.Request.Context(), targetUserID, adminID, req.Reason); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

package auth

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Login ---
type LoginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	User            LoginUserResp `json:"user"`
	AccessToken     string        `json:"access_token"`
	TokenType       string        `json:"token_type"`
	ExpiresIn       int           `json:"expires_in"`
	RawRefreshToken string        `json:"-"` // internal use only, set as cookie by handler
}

type Login2FAResp struct {
	TempToken string `json:"temp_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
	Requires  string `json:"requires"`
}

type LoginUserResp struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// --- Refresh ---
type RefreshResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// --- Me ---
type MeResp struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	AvatarURL   string     `json:"avatar_url,omitempty"`
	Status      model.UserStatus `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	Roles       []RoleResp `json:"roles"`
	Sites       []SiteResp `json:"sites,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type RoleResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type SiteResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// --- Password ---
type ChangePasswordReq struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type ForgotPasswordReq struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordReq struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// --- 2FA ---
type Setup2FAResp struct {
	Secret      string   `json:"secret"`
	QRCodeURI   string   `json:"qr_code_uri"`
	BackupCodes []string `json:"backup_codes"`
}

type Verify2FAReq struct {
	Code string `json:"code" binding:"required,len=6"`
}

type Validate2FAReq struct {
	Code string `json:"code" binding:"required"`
}

type Disable2FAReq struct {
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required,len=6"`
}

type RegenerateBackupCodesReq struct {
	Password string `json:"password" binding:"required"`
}

type RegenerateBackupCodesResp struct {
	BackupCodes []string `json:"backup_codes"`
}

type Get2FAStatusResp struct {
	Enabled    bool       `json:"enabled"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}

type ForceDisable2FAReq struct {
	Reason string `json:"reason" binding:"required"`
}

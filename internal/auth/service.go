package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/crypto"
	"github.com/sky-flux/cms/internal/pkg/jwt"
)

const (
	maxLoginAttempts = 5
	lockoutWindow    = 15 * time.Minute
	resetTokenExpiry = 30 * time.Minute
)

type Service struct {
	userRepo      UserRepository
	tokenRepo     TokenRepository
	totpRepo      TOTPRepository
	roleLoader    RoleLoader
	siteLoader    SiteLoader
	jwtMgr        *jwt.Manager
	rdb           *redis.Client
	totpKey       string // AES-256 encryption key hex
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

type ServiceConfig struct {
	TOTPEncryptionKey string
	AccessExpiry      time.Duration
	RefreshExpiry     time.Duration
}

func NewService(
	userRepo UserRepository,
	tokenRepo TokenRepository,
	totpRepo TOTPRepository,
	roleLoader RoleLoader,
	siteLoader SiteLoader,
	jwtMgr *jwt.Manager,
	rdb *redis.Client,
	cfg ServiceConfig,
) *Service {
	return &Service{
		userRepo:      userRepo,
		tokenRepo:     tokenRepo,
		totpRepo:      totpRepo,
		roleLoader:    roleLoader,
		siteLoader:    siteLoader,
		jwtMgr:        jwtMgr,
		rdb:           rdb,
		totpKey:       cfg.TOTPEncryptionKey,
		accessExpiry:  cfg.AccessExpiry,
		refreshExpiry: cfg.RefreshExpiry,
	}
}

// Login authenticates a user and returns tokens or a 2FA challenge.
func (s *Service) Login(ctx context.Context, req *LoginReq, ip, userAgent string) (*LoginResp, *Login2FAResp, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, apperror.Unauthorized("invalid email or password", nil)
	}
	if user.Status != model.UserStatusActive {
		return nil, nil, apperror.Unauthorized("account is disabled", nil)
	}

	// Check lockout
	lockoutKey := fmt.Sprintf("login_fail:%s", req.Email)
	attempts, _ := s.rdb.Get(ctx, lockoutKey).Int()
	if attempts >= maxLoginAttempts {
		return nil, nil, apperror.Unauthorized("account temporarily locked due to too many failed attempts", nil)
	}

	// Verify password
	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		s.rdb.Incr(ctx, lockoutKey)
		s.rdb.Expire(ctx, lockoutKey, lockoutWindow)
		return nil, nil, apperror.Unauthorized("invalid email or password", nil)
	}

	// Reset lockout counter on success
	s.rdb.Del(ctx, lockoutKey)
	s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Check 2FA
	totp, err := s.totpRepo.GetByUserID(ctx, user.ID)
	if err == nil && totp.Enabled == model.ToggleYes {
		// Issue temp token for 2FA
		tempToken, err := s.jwtMgr.SignTempToken(user.ID, "2fa_verification")
		if err != nil {
			return nil, nil, apperror.Internal("sign temp token failed", err)
		}
		return nil, &Login2FAResp{
			TempToken: tempToken,
			TokenType: "Bearer",
			ExpiresIn: 300, // 5 minutes
			Requires:  "totp",
		}, nil
	}

	// No 2FA — issue full tokens
	loginResp, err := s.issueTokens(ctx, user, ip, userAgent)
	if err != nil {
		return nil, nil, err
	}
	return loginResp, nil, nil
}

// Refresh issues a new access token using a refresh token.
func (s *Service) Refresh(ctx context.Context, rawRefreshToken, ip, userAgent string) (*RefreshResp, string, error) {
	hash := crypto.HashToken(rawRefreshToken)
	rt, err := s.tokenRepo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return nil, "", apperror.Unauthorized("invalid or expired refresh token", nil)
	}

	// Revoke old
	s.tokenRepo.RevokeRefreshToken(ctx, rt.ID)

	// Issue new access token
	accessToken, err := s.jwtMgr.SignAccessToken(rt.UserID)
	if err != nil {
		return nil, "", apperror.Internal("sign access token failed", err)
	}

	// Issue new refresh token
	newRaw, newHash, err := crypto.GenerateToken(32)
	if err != nil {
		return nil, "", apperror.Internal("generate refresh token failed", err)
	}

	newRT := &model.RefreshToken{
		UserID:    rt.UserID,
		TokenHash: newHash,
		ExpiresAt: time.Now().Add(s.refreshExpiry),
		IPAddress: ip,
		UserAgent: userAgent,
	}
	if err := s.tokenRepo.CreateRefreshToken(ctx, newRT); err != nil {
		return nil, "", apperror.Internal("store refresh token failed", err)
	}

	return &RefreshResp{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.accessExpiry.Seconds()),
	}, newRaw, nil
}

// Logout blacklists the access token and revokes the refresh token.
func (s *Service) Logout(ctx context.Context, jti string, rawRefreshToken string) error {
	s.jwtMgr.Blacklist(ctx, jti, s.accessExpiry)
	if rawRefreshToken != "" {
		hash := crypto.HashToken(rawRefreshToken)
		rt, err := s.tokenRepo.GetRefreshTokenByHash(ctx, hash)
		if err == nil {
			s.tokenRepo.RevokeRefreshToken(ctx, rt.ID)
		}
	}
	return nil
}

// Me returns the current user's profile with roles and sites.
func (s *Service) Me(ctx context.Context, userID string) (*MeResp, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles, _ := s.roleLoader.GetUserRoles(ctx, userID)
	sites, _ := s.siteLoader.GetUserSites(ctx, userID)

	roleResps := make([]RoleResp, len(roles))
	for i, r := range roles {
		roleResps[i] = RoleResp{ID: r.ID, Name: r.Name, Slug: r.Slug}
	}

	siteResps := make([]SiteResp, len(sites))
	for i, st := range sites {
		siteResps[i] = SiteResp{ID: st.ID, Name: st.Name, Slug: st.Slug}
	}

	return &MeResp{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
		Roles:       roleResps,
		Sites:       siteResps,
		CreatedAt:   user.CreatedAt,
	}, nil
}

// ChangePassword verifies the current password and updates to a new one.
func (s *Service) ChangePassword(ctx context.Context, userID string, req *ChangePasswordReq) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !crypto.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		return apperror.Unauthorized("current password is incorrect", nil)
	}
	newHash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return apperror.Internal("hash password failed", err)
	}
	if err := s.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		return err
	}
	s.tokenRepo.RevokeAllUserTokens(ctx, userID)
	return nil
}

// ForgotPassword generates a password reset token. Always returns nil for security.
func (s *Service) ForgotPassword(ctx context.Context, req *ForgotPasswordReq) error {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil // Don't reveal whether email exists
	}

	raw, hash, err := crypto.GenerateToken(32)
	if err != nil {
		slog.Error("generate reset token failed", "error", err)
		return nil
	}

	prt := &model.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(resetTokenExpiry),
	}
	if err := s.tokenRepo.CreatePasswordResetToken(ctx, prt); err != nil {
		slog.Error("store reset token failed", "error", err)
		return nil
	}

	// TODO: Send email via Resend
	slog.Info("password reset token generated", "user_id", user.ID, "token", raw)
	return nil
}

// ResetPassword validates a reset token and updates the password.
func (s *Service) ResetPassword(ctx context.Context, req *ResetPasswordReq) error {
	hash := crypto.HashToken(req.Token)
	prt, err := s.tokenRepo.GetPasswordResetTokenByHash(ctx, hash)
	if err != nil {
		return apperror.Unauthorized("invalid or expired reset token", nil)
	}

	newHash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return apperror.Internal("hash password failed", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, prt.UserID, newHash); err != nil {
		return err
	}
	s.tokenRepo.MarkPasswordResetTokenUsed(ctx, prt.ID)
	s.tokenRepo.RevokeAllUserTokens(ctx, prt.UserID)
	return nil
}

// Setup2FA generates a new TOTP secret and backup codes.
func (s *Service) Setup2FA(ctx context.Context, userID string) (*Setup2FAResp, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if already enabled
	existing, err := s.totpRepo.GetByUserID(ctx, userID)
	if err == nil && existing.Enabled == model.ToggleYes {
		return nil, apperror.Conflict("2FA is already enabled", nil)
	}

	key, err := crypto.GenerateTOTPKey(user.Email, "Sky Flux CMS")
	if err != nil {
		return nil, apperror.Internal("generate TOTP key failed", err)
	}

	encrypted, err := crypto.EncryptTOTPSecret(key.Secret(), s.totpKey)
	if err != nil {
		return nil, apperror.Internal("encrypt TOTP secret failed", err)
	}

	backupCodes := crypto.GenerateBackupCodes(10)
	hashedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashedCodes[i] = crypto.HashToken(code)
	}

	totp := &model.UserTOTP{
		UserID:          userID,
		SecretEncrypted: encrypted,
		BackupCodesHash: hashedCodes,
		Enabled:         model.ToggleNo,
	}
	if err := s.totpRepo.Upsert(ctx, totp); err != nil {
		return nil, apperror.Internal("store TOTP record failed", err)
	}

	return &Setup2FAResp{
		Secret:      key.Secret(),
		QRCodeURI:   key.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// Verify2FA confirms that the user's TOTP setup works, enabling 2FA.
func (s *Service) Verify2FA(ctx context.Context, userID string, req *Verify2FAReq) error {
	totp, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return apperror.NotFound("2FA setup not found — call setup first", nil)
	}
	if totp.Enabled == model.ToggleYes {
		return apperror.Conflict("2FA is already verified and enabled", nil)
	}

	secret, err := crypto.DecryptTOTPSecret(totp.SecretEncrypted, s.totpKey)
	if err != nil {
		return apperror.Internal("decrypt TOTP secret failed", err)
	}

	if !crypto.ValidateTOTPCode(secret, req.Code) {
		return apperror.Unauthorized("invalid TOTP code", nil)
	}

	return s.totpRepo.Enable(ctx, totp.ID)
}

// Validate2FA validates a TOTP code during login (2FA challenge).
func (s *Service) Validate2FA(ctx context.Context, userID string, req *Validate2FAReq, ip, userAgent string) (*LoginResp, error) {
	totp, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.NotFound("2FA not configured", nil)
	}

	secret, err := crypto.DecryptTOTPSecret(totp.SecretEncrypted, s.totpKey)
	if err != nil {
		return nil, apperror.Internal("decrypt TOTP secret failed", err)
	}

	// Try TOTP code first
	if !crypto.ValidateTOTPCode(secret, req.Code) {
		// Try backup codes
		codeHash := crypto.HashToken(req.Code)
		matched := false
		remaining := make([]string, 0, len(totp.BackupCodesHash))
		for _, h := range totp.BackupCodesHash {
			if h == codeHash && !matched {
				matched = true
				continue // Remove used backup code
			}
			remaining = append(remaining, h)
		}
		if !matched {
			return nil, apperror.Unauthorized("invalid 2FA code", nil)
		}
		// Update remaining backup codes
		s.totpRepo.UpdateBackupCodes(ctx, totp.ID, remaining)
	}

	// Issue full tokens
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, user, ip, userAgent)
}

// Disable2FA disables 2FA after verifying password and TOTP code.
func (s *Service) Disable2FA(ctx context.Context, userID string, req *Disable2FAReq) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		return apperror.Unauthorized("invalid password", nil)
	}

	totp, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return apperror.NotFound("2FA is not enabled", nil)
	}

	secret, err := crypto.DecryptTOTPSecret(totp.SecretEncrypted, s.totpKey)
	if err != nil {
		return apperror.Internal("decrypt TOTP secret failed", err)
	}

	if !crypto.ValidateTOTPCode(secret, req.Code) {
		return apperror.Unauthorized("invalid TOTP code", nil)
	}

	if err := s.totpRepo.Delete(ctx, userID); err != nil {
		return err
	}
	s.tokenRepo.RevokeAllUserTokens(ctx, userID)
	return nil
}

// RegenerateBackupCodes generates new backup codes after password verification.
func (s *Service) RegenerateBackupCodes(ctx context.Context, userID string, req *RegenerateBackupCodesReq) (*RegenerateBackupCodesResp, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		return nil, apperror.Unauthorized("invalid password", nil)
	}

	totp, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.NotFound("2FA is not enabled", nil)
	}

	codes := crypto.GenerateBackupCodes(10)
	hashed := make([]string, len(codes))
	for i, c := range codes {
		hashed[i] = crypto.HashToken(c)
	}

	if err := s.totpRepo.UpdateBackupCodes(ctx, totp.ID, hashed); err != nil {
		return nil, err
	}

	return &RegenerateBackupCodesResp{BackupCodes: codes}, nil
}

// Get2FAStatus returns the current 2FA status for a user.
func (s *Service) Get2FAStatus(ctx context.Context, userID string) (*Get2FAStatusResp, error) {
	totp, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		// Not found means 2FA is not set up
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return &Get2FAStatusResp{Enabled: false}, nil
		}
		return nil, err
	}
	return &Get2FAStatusResp{
		Enabled:    totp.Enabled == model.ToggleYes,
		VerifiedAt: totp.VerifiedAt,
	}, nil
}

// ForceDisable2FA allows an admin to disable 2FA for any user.
func (s *Service) ForceDisable2FA(ctx context.Context, targetUserID, adminUserID, reason string) error {
	if err := s.totpRepo.Delete(ctx, targetUserID); err != nil {
		return err
	}
	s.tokenRepo.RevokeAllUserTokens(ctx, targetUserID)
	slog.Info("2FA force-disabled by admin",
		"target_user_id", targetUserID,
		"admin_user_id", adminUserID,
		"reason", reason,
	)
	return nil
}

// issueTokens creates access + refresh tokens for a successful login.
func (s *Service) issueTokens(ctx context.Context, user *model.User, ip, userAgent string) (*LoginResp, error) {
	accessToken, err := s.jwtMgr.SignAccessToken(user.ID)
	if err != nil {
		return nil, apperror.Internal("sign access token failed", err)
	}

	rawRefresh, hashRefresh, err := crypto.GenerateToken(32)
	if err != nil {
		return nil, apperror.Internal("generate refresh token failed", err)
	}

	rt := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashRefresh,
		ExpiresAt: time.Now().Add(s.refreshExpiry),
		IPAddress: ip,
		UserAgent: userAgent,
	}
	if err := s.tokenRepo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, apperror.Internal("store refresh token failed", err)
	}

	return &LoginResp{
		User: LoginUserResp{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
		AccessToken:     accessToken,
		TokenType:       "Bearer",
		ExpiresIn:       int(s.accessExpiry.Seconds()),
		RawRefreshToken: rawRefresh,
	}, nil
}

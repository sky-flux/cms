package app

import (
	"context"
	"errors"

	"github.com/sky-flux/cms/internal/identity/domain"
)

const maxLoginAttempts = 5

const lockKeyPrefix = "login_fail:"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account temporarily locked")
	ErrAccountDisabled    = errors.New("account is disabled")
)

// LoginInput carries the raw credentials from the delivery layer.
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput is returned on successful authentication.
type LoginOutput struct {
	UserID      string
	AccessToken string
	Requires2FA bool
	TempToken   string // set when Requires2FA is true
}

// PasswordChecker abstracts bcrypt so domain stays framework-free.
type PasswordChecker interface {
	Check(plain, hash string) bool
}

// TokenIssuer abstracts JWT signing.
type TokenIssuer interface {
	IssueAccessToken(userID string) (string, error)
}

// LockoutChecker abstracts Redis-based brute-force protection.
type LockoutChecker interface {
	Attempts(ctx context.Context, key string) (int, error)
	Increment(ctx context.Context, key string) error
	Reset(ctx context.Context, key string) error
}

// LoginUseCase orchestrates credential validation and token issuance.
type LoginUseCase struct {
	users   domain.UserRepository
	pw      PasswordChecker
	tokens  TokenIssuer
	lockout LockoutChecker
}

func NewLoginUseCase(
	users domain.UserRepository,
	pw PasswordChecker,
	tokens TokenIssuer,
	lockout LockoutChecker,
) *LoginUseCase {
	return &LoginUseCase{users: users, pw: pw, tokens: tokens, lockout: lockout}
}

func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	user, err := uc.users.FindByEmail(ctx, in.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive() {
		return nil, ErrAccountDisabled
	}

	lockKey := lockKeyPrefix + in.Email
	attempts, _ := uc.lockout.Attempts(ctx, lockKey)
	if attempts >= maxLoginAttempts {
		return nil, ErrAccountLocked
	}

	if !uc.pw.Check(in.Password, user.PasswordHash) {
		_ = uc.lockout.Increment(ctx, lockKey)
		return nil, ErrInvalidCredentials
	}

	_ = uc.lockout.Reset(ctx, lockKey)
	_ = uc.users.UpdateLastLogin(ctx, user.ID)

	// 2FA required — issue temp token for the challenge step.
	if user.TOTPEnabled {
		temp, err := uc.tokens.IssueAccessToken(user.ID) // reuse; delivery layer marks purpose
		if err != nil {
			return nil, err
		}
		return &LoginOutput{UserID: user.ID, Requires2FA: true, TempToken: temp}, nil
	}

	access, err := uc.tokens.IssueAccessToken(user.ID)
	if err != nil {
		return nil, err
	}
	return &LoginOutput{UserID: user.ID, AccessToken: access}, nil
}

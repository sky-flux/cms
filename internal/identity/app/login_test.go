package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sky-flux/cms/internal/identity/app"
	"github.com/sky-flux/cms/internal/identity/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- hand-written mocks ---

type mockUserRepo struct {
	findByEmailFn     func(ctx context.Context, email string) (*domain.User, error)
	updateLastLoginFn func(ctx context.Context, id string) error
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.findByEmailFn(ctx, email)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Save(ctx context.Context, u *domain.User) error { return nil }
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id, hash string) error {
	return nil
}
func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	if m.updateLastLoginFn != nil {
		return m.updateLastLoginFn(ctx, id)
	}
	return nil
}

type mockPasswordChecker struct {
	result bool
}

func (m *mockPasswordChecker) Check(plain, hash string) bool { return m.result }

type mockTokenIssuer struct {
	token string
	err   error
}

func (m *mockTokenIssuer) IssueAccessToken(userID string) (string, error) {
	return m.token, m.err
}

type mockLockout struct {
	attempts int
	locked   bool
}

func (m *mockLockout) Attempts(ctx context.Context, key string) (int, error) {
	return m.attempts, nil
}
func (m *mockLockout) Increment(ctx context.Context, key string) error { return nil }
func (m *mockLockout) Reset(ctx context.Context, key string) error     { return nil }

// --- tests ---

func activeUser() *domain.User {
	u, _ := domain.NewUser("alice@example.com", "Alice", "$2a$12$hashed")
	u.ID = "user-uuid-1"
	return u
}

func TestLoginUseCase_Success(t *testing.T) {
	uc := app.NewLoginUseCase(
		&mockUserRepo{
			findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return activeUser(), nil
			},
		},
		&mockPasswordChecker{result: true},
		&mockTokenIssuer{token: "access-jwt"},
		&mockLockout{attempts: 0},
	)

	result, err := uc.Execute(context.Background(), app.LoginInput{
		Email:    "alice@example.com",
		Password: "correct-password",
	})
	require.NoError(t, err)
	assert.Equal(t, "access-jwt", result.AccessToken)
	assert.Equal(t, "user-uuid-1", result.UserID)
}

func TestLoginUseCase_UserNotFound(t *testing.T) {
	uc := app.NewLoginUseCase(
		&mockUserRepo{
			findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return nil, errors.New("not found")
			},
		},
		&mockPasswordChecker{result: false},
		&mockTokenIssuer{},
		&mockLockout{},
	)

	_, err := uc.Execute(context.Background(), app.LoginInput{
		Email:    "nobody@example.com",
		Password: "pw",
	})
	assert.ErrorIs(t, err, app.ErrInvalidCredentials)
}

func TestLoginUseCase_WrongPassword(t *testing.T) {
	uc := app.NewLoginUseCase(
		&mockUserRepo{
			findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return activeUser(), nil
			},
		},
		&mockPasswordChecker{result: false},
		&mockTokenIssuer{},
		&mockLockout{attempts: 0},
	)

	_, err := uc.Execute(context.Background(), app.LoginInput{
		Email:    "alice@example.com",
		Password: "wrong",
	})
	assert.ErrorIs(t, err, app.ErrInvalidCredentials)
}

func TestLoginUseCase_AccountLocked(t *testing.T) {
	uc := app.NewLoginUseCase(
		&mockUserRepo{
			findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return activeUser(), nil
			},
		},
		&mockPasswordChecker{result: true},
		&mockTokenIssuer{},
		&mockLockout{attempts: 5}, // maxLoginAttempts = 5
	)

	_, err := uc.Execute(context.Background(), app.LoginInput{
		Email:    "alice@example.com",
		Password: "correct",
	})
	assert.ErrorIs(t, err, app.ErrAccountLocked)
}

func TestLoginUseCase_DisabledAccount(t *testing.T) {
	disabledUser := activeUser()
	disabledUser.Disable()

	uc := app.NewLoginUseCase(
		&mockUserRepo{
			findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return disabledUser, nil
			},
		},
		&mockPasswordChecker{result: true},
		&mockTokenIssuer{},
		&mockLockout{attempts: 0},
	)

	_, err := uc.Execute(context.Background(), app.LoginInput{
		Email:    "alice@example.com",
		Password: "correct",
	})
	assert.ErrorIs(t, err, app.ErrAccountDisabled)
}

func TestLoginUseCase_Requires2FA(t *testing.T) {
	user := activeUser()
	user.TOTPEnabled = true

	uc := app.NewLoginUseCase(
		&mockUserRepo{
			findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return user, nil
			},
		},
		&mockPasswordChecker{result: true},
		&mockTokenIssuer{token: "temp-jwt"},
		&mockLockout{attempts: 0},
	)

	result, err := uc.Execute(context.Background(), app.LoginInput{
		Email:    "alice@example.com",
		Password: "correct",
	})
	require.NoError(t, err)
	assert.True(t, result.Requires2FA)
	assert.Equal(t, "temp-jwt", result.TempToken)
}

package domain_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/identity/domain"
)

// Compile-time interface satisfaction check.
var _ domain.UserRepository = (*mockUserRepo)(nil)

type mockUserRepo struct {
	findByEmailFn     func(ctx context.Context, email string) (*domain.User, error)
	findByIDFn        func(ctx context.Context, id string) (*domain.User, error)
	saveFn            func(ctx context.Context, u *domain.User) error
	updatePasswordFn  func(ctx context.Context, id, hash string) error
	updateLastLoginFn func(ctx context.Context, id string) error
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.findByEmailFn(ctx, email)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockUserRepo) Save(ctx context.Context, u *domain.User) error {
	return m.saveFn(ctx, u)
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id, hash string) error {
	return m.updatePasswordFn(ctx, id, hash)
}
func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	return m.updateLastLoginFn(ctx, id)
}

func TestUserRepository_Interface(t *testing.T) {
	// Satisfied if it compiles.
	t.Log("UserRepository interface satisfied by mockUserRepo")
}

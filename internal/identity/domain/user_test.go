package domain_test

import (
	"testing"

	"github.com/sky-flux/cms/internal/identity/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_ValidInput(t *testing.T) {
	u, err := domain.NewUser("alice@example.com", "Alice", "hashedpw")
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", u.Email)
	assert.Equal(t, "Alice", u.DisplayName)
	assert.Equal(t, domain.UserStatusActive, u.Status)
	assert.Empty(t, u.ID) // set by DB default; domain sets to ""
}

func TestNewUser_InvalidEmail(t *testing.T) {
	_, err := domain.NewUser("not-an-email", "Alice", "hashedpw")
	assert.ErrorIs(t, err, domain.ErrInvalidEmail)
}

func TestNewUser_EmptyDisplayName(t *testing.T) {
	_, err := domain.NewUser("alice@example.com", "", "hashedpw")
	assert.ErrorIs(t, err, domain.ErrEmptyDisplayName)
}

func TestUser_Disable(t *testing.T) {
	u, _ := domain.NewUser("alice@example.com", "Alice", "pw")
	u.Disable()
	assert.Equal(t, domain.UserStatusDisabled, u.Status)
}

func TestUser_Enable(t *testing.T) {
	u, _ := domain.NewUser("alice@example.com", "Alice", "pw")
	u.Disable()
	u.Enable()
	assert.Equal(t, domain.UserStatusActive, u.Status)
}

func TestUser_IsActive(t *testing.T) {
	u, _ := domain.NewUser("alice@example.com", "Alice", "pw")
	assert.True(t, u.IsActive())
	u.Disable()
	assert.False(t, u.IsActive())
}

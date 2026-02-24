package crypto_test

import (
	"testing"

	"github.com/sky-flux/cms/internal/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_ReturnsHash(t *testing.T) {
	hash, err := crypto.HashPassword("SecureP@ss1")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "SecureP@ss1", hash)
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	hash, _ := crypto.HashPassword("SecureP@ss1")
	assert.True(t, crypto.CheckPassword("SecureP@ss1", hash))
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	hash, _ := crypto.HashPassword("SecureP@ss1")
	assert.False(t, crypto.CheckPassword("WrongPass", hash))
}

func TestHashPassword_DifferentHashesForSameInput(t *testing.T) {
	h1, _ := crypto.HashPassword("same")
	h2, _ := crypto.HashPassword("same")
	assert.NotEqual(t, h1, h2, "bcrypt should produce different hashes due to salt")
}

package crypto_test

import (
	"testing"

	"github.com/sky-flux/cms/internal/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken_ReturnsTokenAndHash(t *testing.T) {
	raw, hash, err := crypto.GenerateToken(32)
	require.NoError(t, err)
	assert.Len(t, raw, 64, "32 bytes = 64 hex chars")
	assert.Len(t, hash, 64, "SHA-256 = 64 hex chars")
	assert.NotEqual(t, raw, hash)
}

func TestGenerateToken_Uniqueness(t *testing.T) {
	r1, _, _ := crypto.GenerateToken(32)
	r2, _, _ := crypto.GenerateToken(32)
	assert.NotEqual(t, r1, r2)
}

func TestHashToken_Deterministic(t *testing.T) {
	h1 := crypto.HashToken("abc123")
	h2 := crypto.HashToken("abc123")
	assert.Equal(t, h1, h2)
}

func TestHashToken_DifferentInputDifferentOutput(t *testing.T) {
	assert.NotEqual(t, crypto.HashToken("a"), crypto.HashToken("b"))
}

func TestGenerateBackupCodes_Returns10Codes(t *testing.T) {
	codes := crypto.GenerateBackupCodes(10)
	assert.Len(t, codes, 10)
	for _, code := range codes {
		assert.Regexp(t, `^[A-Z0-9]{4}-[A-Z0-9]{4}$`, code)
	}
}

func TestGenerateBackupCodes_Unique(t *testing.T) {
	codes := crypto.GenerateBackupCodes(10)
	seen := make(map[string]bool)
	for _, c := range codes {
		assert.False(t, seen[c], "duplicate backup code")
		seen[c] = true
	}
}

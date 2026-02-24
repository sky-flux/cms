package crypto_test

import (
	"testing"

	"github.com/pquerna/otp/totp"
	"github.com/sky-flux/cms/internal/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testEncKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	plaintext := "JBSWY3DPEHPK3PXP"
	encrypted, err := crypto.EncryptTOTPSecret(plaintext, testEncKey)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)
	decrypted, err := crypto.DecryptTOTPSecret(encrypted, testEncKey)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncrypt_DifferentCiphertexts(t *testing.T) {
	e1, _ := crypto.EncryptTOTPSecret("secret", testEncKey)
	e2, _ := crypto.EncryptTOTPSecret("secret", testEncKey)
	assert.NotEqual(t, e1, e2, "AES-GCM should produce different ciphertexts due to random nonce")
}

func TestDecrypt_WrongKey_Fails(t *testing.T) {
	encrypted, _ := crypto.EncryptTOTPSecret("secret", testEncKey)
	wrongKey := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	_, err := crypto.DecryptTOTPSecret(encrypted, wrongKey)
	assert.Error(t, err)
}

func TestGenerateTOTPKey_Valid(t *testing.T) {
	key, err := crypto.GenerateTOTPKey("admin@example.com", "Sky Flux CMS")
	require.NoError(t, err)
	assert.NotEmpty(t, key.Secret())
	assert.Contains(t, key.URL(), "otpauth://totp/")
}

func TestValidateTOTPCode_ValidCode(t *testing.T) {
	key, _ := crypto.GenerateTOTPKey("admin@example.com", "Sky Flux CMS")
	code, _ := totp.GenerateCode(key.Secret(), crypto.TOTPNow())
	assert.True(t, crypto.ValidateTOTPCode(key.Secret(), code))
}

func TestValidateTOTPCode_InvalidCode(t *testing.T) {
	key, _ := crypto.GenerateTOTPKey("admin@example.com", "Sky Flux CMS")
	assert.False(t, crypto.ValidateTOTPCode(key.Secret(), "000000"))
}

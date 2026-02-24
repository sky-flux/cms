# Global Endpoints Batch 1: Setup + Auth — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement 19 global endpoints (Setup 2 + Auth 17) with full test coverage, building from infrastructure up.

**Architecture:** Handler → Service → Repository three-layer pattern. Infrastructure packages (jwt, crypto) provide shared utilities. Middleware (auth, installation_guard) integrates into Gin chain. Router is the single assembly point.

**Tech Stack:** Go 1.25+ / Gin v1.11 / uptrace/bun / golang-jwt/jwt/v5 / golang.org/x/crypto (bcrypt) / pquerna/otp (TOTP) / Redis / testify

**Design Doc:** `docs/plans/2026-02-24-global-endpoints-batch1-design.md`

**Reference Files:**
- Models: `internal/model/user.go`, `user_totp.go`, `refresh_token.go`, `password_reset_token.go`, `config.go`, `site.go`, `user_role.go`
- Patterns: `internal/rbac/handler.go` (handler pattern), `internal/rbac/service.go` (service + Redis caching), `internal/rbac/interfaces.go` (interface pattern)
- Existing: `internal/pkg/apperror/errors.go`, `internal/pkg/response/response.go`, `internal/middleware/rbac.go` (middleware pattern)
- Config: `internal/config/config.go` (JWTConfig, TOTPConfig fields)
- Schema: `internal/schema/migrate.go` (CreateSiteSchema signature)

---

## Task 1: Add TOTP dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add pquerna/otp**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go get github.com/pquerna/otp@latest
```

**Step 2: Verify**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && grep "pquerna/otp" go.mod
```

Expected: Line showing `github.com/pquerna/otp vX.X.X`

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add go.mod go.sum && git commit -m "deps: add pquerna/otp for TOTP 2FA support"
```

---

## Task 2: Crypto package — password hashing

**Files:**
- Create: `internal/pkg/crypto/password.go`
- Create: `internal/pkg/crypto/password_test.go`

**Step 1: Write tests**

```go
// internal/pkg/crypto/password_test.go
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
```

**Step 2: Run tests — verify they fail**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/pkg/crypto/ -run TestHashPassword -v
```

Expected: FAIL — package does not exist

**Step 3: Implement**

```go
// internal/pkg/crypto/password.go
package crypto

import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

// HashPassword returns a bcrypt hash of the password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
```

**Step 4: Run tests — verify they pass**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/pkg/crypto/ -run TestPassword -v
```

Expected: PASS (all 4 tests)

**Step 5: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/pkg/crypto/password.go internal/pkg/crypto/password_test.go && git commit -m "feat(crypto): add bcrypt password hashing with cost=12"
```

---

## Task 3: Crypto package — secure token generation

**Files:**
- Create: `internal/pkg/crypto/token.go`
- Create: `internal/pkg/crypto/token_test.go`

**Step 1: Write tests**

```go
// internal/pkg/crypto/token_test.go
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
```

**Step 2: Run tests — verify fail**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/pkg/crypto/ -run "TestGenerateToken|TestHashToken|TestGenerateBackupCodes" -v
```

**Step 3: Implement**

```go
// internal/pkg/crypto/token.go
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// GenerateToken creates a cryptographically random token and its SHA-256 hash.
// Returns (rawHex, hashHex, error). Store hashHex in DB, send rawHex to client.
func GenerateToken(byteLen int) (string, string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate random bytes: %w", err)
	}
	raw := hex.EncodeToString(b)
	return raw, HashToken(raw), nil
}

// HashToken returns the SHA-256 hex digest of a token string.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// GenerateBackupCodes generates n backup codes in XXXX-XXXX format.
// Characters: A-Z0-9 (no ambiguous chars like 0/O, 1/I/L).
func GenerateBackupCodes(n int) []string {
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789" // exclude 0,O,1,I,L
	codes := make([]string, 0, n)
	seen := make(map[string]bool)

	for len(codes) < n {
		code := randomCode(charset, 4) + "-" + randomCode(charset, 4)
		if !seen[code] {
			seen[code] = true
			codes = append(codes, code)
		}
	}
	return codes
}

func randomCode(charset string, length int) string {
	b := make([]byte, length)
	rand.Read(b)
	var sb strings.Builder
	for _, v := range b {
		sb.WriteByte(charset[int(v)%len(charset)])
	}
	return sb.String()
}
```

**Step 4: Run tests — verify pass**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/pkg/crypto/ -run "TestGenerateToken|TestHashToken|TestGenerateBackupCodes" -v
```

**Step 5: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/pkg/crypto/token.go internal/pkg/crypto/token_test.go && git commit -m "feat(crypto): add secure token generation and backup code generator"
```

---

## Task 4: Crypto package — TOTP encryption

**Files:**
- Create: `internal/pkg/crypto/totp.go`
- Create: `internal/pkg/crypto/totp_test.go`

**Step 1: Write tests**

```go
// internal/pkg/crypto/totp_test.go
package crypto_test

import (
	"testing"

	"github.com/pquerna/otp/totp"
	"github.com/sky-flux/cms/internal/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 32-byte hex key for AES-256 tests
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
```

**Step 2: Run tests — verify fail**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/pkg/crypto/ -run "TestEncrypt|TestDecrypt|TestGenerateTOTP|TestValidateTOTP" -v
```

**Step 3: Implement**

```go
// internal/pkg/crypto/totp.go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// EncryptTOTPSecret encrypts a TOTP secret using AES-256-GCM.
// encKeyHex is a 64-char hex string (32 bytes).
func EncryptTOTPSecret(secret, encKeyHex string) (string, error) {
	key, err := hex.DecodeString(encKeyHex)
	if err != nil {
		return "", fmt.Errorf("decode encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nonce, nonce, []byte(secret), nil)
	return hex.EncodeToString(ciphertext), nil
}

// DecryptTOTPSecret decrypts an AES-256-GCM encrypted TOTP secret.
func DecryptTOTPSecret(encryptedHex, encKeyHex string) (string, error) {
	key, err := hex.DecodeString(encKeyHex)
	if err != nil {
		return "", fmt.Errorf("decode encryption key: %w", err)
	}

	data, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// GenerateTOTPKey creates a new TOTP key for the given account and issuer.
func GenerateTOTPKey(account, issuer string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: account,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
}

// ValidateTOTPCode validates a TOTP code against a secret with ±1 window.
func ValidateTOTPCode(secret, code string) bool {
	valid, _ := totp.ValidateCustom(code, secret, TOTPNow(), totp.ValidateOpts{
		Period:    30,
		Skew:     1,
		Digits:   otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	return valid
}

// TOTPNow returns the current time. Extracted for testability.
func TOTPNow() time.Time {
	return time.Now()
}
```

**Step 4: Run tests — verify pass**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/pkg/crypto/ -v
```

Expected: ALL tests in crypto package pass

**Step 5: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/pkg/crypto/totp.go internal/pkg/crypto/totp_test.go && git commit -m "feat(crypto): add AES-256-GCM TOTP encryption and TOTP validation"
```

---

## Task 5: JWT package

**Files:**
- Create: `internal/pkg/jwt/jwt.go`
- Create: `internal/pkg/jwt/jwt_test.go`

**Step 1: Write tests**

```go
// internal/pkg/jwt/jwt_test.go
package jwt_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-at-least-32-bytes!"

func newTestManager(t *testing.T) (*jwt.Manager, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	mgr := jwt.NewManager(testSecret, 15*time.Minute, 5*time.Minute, rdb)
	return mgr, mr
}

func TestSignAndVerify_AccessToken(t *testing.T) {
	mgr, _ := newTestManager(t)
	token, err := mgr.SignAccessToken("user-123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := mgr.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.Subject)
	assert.Empty(t, claims.Purpose)
}

func TestSignAndVerify_TempToken(t *testing.T) {
	mgr, _ := newTestManager(t)
	token, err := mgr.SignTempToken("user-456", "2fa_verification")
	require.NoError(t, err)

	claims, err := mgr.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, "user-456", claims.Subject)
	assert.Equal(t, "2fa_verification", claims.Purpose)
}

func TestVerify_ExpiredToken(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:0"}) // won't be used
	mgr := jwt.NewManager(testSecret, -1*time.Second, -1*time.Second, rdb)
	token, _ := mgr.SignAccessToken("user-789")
	_, err := mgr.Verify(token)
	assert.Error(t, err)
}

func TestVerify_InvalidSignature(t *testing.T) {
	mgr, _ := newTestManager(t)
	_, err := mgr.Verify("eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ0ZXN0In0.invalid")
	assert.Error(t, err)
}

func TestBlacklist_BlocksToken(t *testing.T) {
	mgr, _ := newTestManager(t)
	token, _ := mgr.SignAccessToken("user-111")
	claims, _ := mgr.Verify(token)

	err := mgr.Blacklist(context.Background(), claims.JTI, 15*time.Minute)
	require.NoError(t, err)

	ok, err := mgr.IsBlacklisted(context.Background(), claims.JTI)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestBlacklist_NonBlacklistedToken(t *testing.T) {
	mgr, _ := newTestManager(t)
	ok, err := mgr.IsBlacklisted(context.Background(), "non-existent-jti")
	require.NoError(t, err)
	assert.False(t, ok)
}
```

**Step 2: Run tests — verify fail**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/pkg/jwt/ -v
```

**Step 3: Implement**

```go
// internal/pkg/jwt/jwt.go
package jwt

import (
	"context"
	"fmt"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Claims represents JWT token claims used throughout the application.
type Claims struct {
	Subject string // user ID
	JTI     string // token ID
	Purpose string // empty for access token, "2fa_verification" for temp token
}

type registeredClaims struct {
	jwtgo.RegisteredClaims
	Purpose string `json:"purpose,omitempty"`
}

// Manager handles JWT signing, verification, and blacklisting.
type Manager struct {
	secret      []byte
	accessTTL   time.Duration
	tempTTL     time.Duration
	rdb         *redis.Client
}

// NewManager creates a JWT manager.
func NewManager(secret string, accessTTL, tempTTL time.Duration, rdb *redis.Client) *Manager {
	return &Manager{
		secret:    []byte(secret),
		accessTTL: accessTTL,
		tempTTL:   tempTTL,
		rdb:       rdb,
	}
}

// SignAccessToken issues a standard access token for the given user ID.
func (m *Manager) SignAccessToken(userID string) (string, error) {
	return m.sign(userID, "", m.accessTTL)
}

// SignTempToken issues a temporary token with a specific purpose (e.g. "2fa_verification").
func (m *Manager) SignTempToken(userID, purpose string) (string, error) {
	return m.sign(userID, purpose, m.tempTTL)
}

func (m *Manager) sign(userID, purpose string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := registeredClaims{
		RegisteredClaims: jwtgo.RegisteredClaims{
			Subject:   userID,
			ID:        uuid.NewString(),
			IssuedAt:  jwtgo.NewNumericDate(now),
			ExpiresAt: jwtgo.NewNumericDate(now.Add(ttl)),
		},
		Purpose: purpose,
	}

	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Verify parses and validates a JWT token, returning its claims.
func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwtgo.ParseWithClaims(tokenStr, &registeredClaims{}, func(t *jwtgo.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtgo.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	rc, ok := token.Claims.(*registeredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &Claims{
		Subject: rc.Subject,
		JTI:     rc.ID,
		Purpose: rc.Purpose,
	}, nil
}

// Blacklist adds a JTI to the Redis blacklist with the given TTL.
func (m *Manager) Blacklist(ctx context.Context, jti string, ttl time.Duration) error {
	return m.rdb.Set(ctx, "auth:blacklist:"+jti, "1", ttl).Err()
}

// IsBlacklisted checks if a JTI is in the Redis blacklist.
func (m *Manager) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := m.rdb.Exists(ctx, "auth:blacklist:"+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
```

**Step 4: Run tests — verify pass**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/pkg/jwt/ -v
```

**Step 5: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/pkg/jwt/ && git commit -m "feat(jwt): add JWT signing, verification, and Redis blacklist"
```

---

## Task 6: Auth middleware

**Files:**
- Rewrite: `internal/middleware/auth.go`
- Create: `internal/middleware/auth_test.go`

**Step 1: Write tests**

```go
// internal/middleware/auth_test.go
package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/middleware"
	"github.com/sky-flux/cms/internal/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const authTestSecret = "test-secret-key-at-least-32-bytes!"

func newAuthTestSetup(t *testing.T) (*jwt.Manager, *gin.Engine) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	mgr := jwt.NewManager(authTestSecret, 15*time.Minute, 5*time.Minute, rdb)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Auth(mgr))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"user_id": c.GetString("user_id")})
	})
	return mgr, r
}

func TestAuth_ValidToken_SetsUserID(t *testing.T) {
	mgr, r := newAuthTestSetup(t)
	token, _ := mgr.SignAccessToken("user-42")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "user-42")
}

func TestAuth_MissingHeader_Returns401(t *testing.T) {
	_, r := newAuthTestSetup(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestAuth_InvalidToken_Returns401(t *testing.T) {
	_, r := newAuthTestSetup(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestAuth_BlacklistedToken_Returns401(t *testing.T) {
	mgr, r := newAuthTestSetup(t)
	token, _ := mgr.SignAccessToken("user-99")
	claims, _ := mgr.Verify(token)
	mgr.Blacklist(context.Background(), claims.JTI, time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestAuth_TempToken_SetsPurpose(t *testing.T) {
	mgr, _ := newAuthTestSetup(t)
	token, _ := mgr.SignTempToken("user-55", "2fa_verification")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Auth(mgr))
	r.POST("/2fa", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id": c.GetString("user_id"),
			"purpose": c.GetString("token_purpose"),
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/2fa", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "2fa_verification")
}
```

**Step 2: Run tests — verify fail**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/middleware/ -run TestAuth -v
```

**Step 3: Implement**

```go
// internal/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/jwt"
)

// Auth returns middleware that validates JWT from the Authorization header.
// Sets "user_id" and optionally "token_purpose" and "token_jti" in Gin context.
func Auth(jwtMgr *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing or invalid authorization header",
			})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtMgr.Verify(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid or expired token",
			})
			return
		}

		// Check blacklist
		blacklisted, err := jwtMgr.IsBlacklisted(c.Request.Context(), claims.JTI)
		if err != nil || blacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "token has been revoked",
			})
			return
		}

		c.Set("user_id", claims.Subject)
		c.Set("token_jti", claims.JTI)
		if claims.Purpose != "" {
			c.Set("token_purpose", claims.Purpose)
		}
		c.Next()
	}
}
```

**Step 4: Run tests — verify pass**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/middleware/ -run TestAuth -v
```

**Step 5: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/middleware/auth.go internal/middleware/auth_test.go && git commit -m "feat(middleware): add JWT auth middleware with blacklist support"
```

---

## Task 7: Installation guard middleware

**Files:**
- Rewrite: `internal/middleware/installation_guard.go`
- Create: `internal/middleware/installation_guard_test.go` (replace stub)

Note: The guard uses an `InstallChecker` interface so it can be tested without real DB/Redis.

**Step 1: Write tests**

```go
// internal/middleware/installation_guard_test.go
package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/middleware"
	"github.com/stretchr/testify/assert"
)

type mockInstallChecker struct {
	installed bool
}

func (m *mockInstallChecker) IsInstalled(ctx context.Context) bool {
	return m.installed
}

func (m *mockInstallChecker) MarkInstalled() {}

func setupGuardRouter(checker middleware.InstallChecker, exemptPaths ...string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.InstallationGuard(checker, exemptPaths...))
	r.GET("/api/v1/posts", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})
	r.POST("/api/v1/setup/check", func(c *gin.Context) {
		c.JSON(200, gin.H{"installed": false})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return r
}

func TestInstallGuard_Installed_PassThrough(t *testing.T) {
	r := setupGuardRouter(&mockInstallChecker{installed: true})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/posts", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestInstallGuard_NotInstalled_Returns503(t *testing.T) {
	r := setupGuardRouter(&mockInstallChecker{installed: false})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/posts", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 503, w.Code)
}

func TestInstallGuard_NotInstalled_SetupExempt(t *testing.T) {
	r := setupGuardRouter(&mockInstallChecker{installed: false}, "/api/v1/setup/")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/setup/check", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestInstallGuard_NotInstalled_HealthExempt(t *testing.T) {
	r := setupGuardRouter(&mockInstallChecker{installed: false}, "/health")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
```

**Step 2: Run tests — verify fail**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/middleware/ -run TestInstallGuard -v
```

**Step 3: Implement**

```go
// internal/middleware/installation_guard.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// InstallChecker abstracts the installation status check.
type InstallChecker interface {
	IsInstalled(ctx context.Context) bool
	MarkInstalled()
}

// InstallationGuard returns middleware that blocks requests when CMS is not installed.
// exemptPrefixes are path prefixes that bypass the guard (e.g. "/api/v1/setup/", "/health").
func InstallationGuard(checker InstallChecker, exemptPrefixes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if path is exempt
		for _, prefix := range exemptPrefixes {
			if strings.HasPrefix(c.Request.URL.Path, prefix) {
				c.Next()
				return
			}
		}

		if checker.IsInstalled(c.Request.Context()) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   gin.H{
				"code":    "NOT_INSTALLED",
				"message": "CMS is not installed. Please complete setup first.",
			},
		})
	}
}
```

**Step 4: Run tests — verify pass**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/middleware/ -run TestInstallGuard -v
```

**Step 5: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/middleware/installation_guard.go internal/middleware/installation_guard_test.go && git commit -m "feat(middleware): add installation guard with triple-check and path exemptions"
```

---

## Task 8: Setup module — DTO, interfaces, repository

**Files:**
- Rewrite: `internal/setup/dto.go`
- Rewrite: `internal/setup/repository.go`
- Create: `internal/setup/interfaces.go`

**Step 1: Implement DTOs**

```go
// internal/setup/dto.go
package setup

// InitializeReq is the request body for POST /api/v1/setup/initialize.
type InitializeReq struct {
	SiteName         string `json:"site_name" binding:"required,max=200"`
	SiteSlug         string `json:"site_slug" binding:"required,min=3,max=50"`
	SiteURL          string `json:"site_url" binding:"required,url"`
	AdminEmail       string `json:"admin_email" binding:"required,email"`
	AdminPassword    string `json:"admin_password" binding:"required,min=8"`
	AdminDisplayName string `json:"admin_display_name" binding:"required,max=100"`
	Locale           string `json:"locale" binding:"omitempty,max=10"`
}

// CheckResp is the response for POST /api/v1/setup/check.
type CheckResp struct {
	Installed bool `json:"installed"`
}

// InitializeResp is the response for POST /api/v1/setup/initialize.
type InitializeResp struct {
	User        UserResp `json:"user"`
	Site        SiteResp `json:"site"`
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"`
}

type UserResp struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type SiteResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
```

**Step 2: Implement interfaces**

```go
// internal/setup/interfaces.go
package setup

import (
	"context"
	"encoding/json"

	"github.com/sky-flux/cms/internal/model"
)

// ConfigRepository provides access to sfc_configs table.
type ConfigRepository interface {
	GetValue(ctx context.Context, key string) (json.RawMessage, error)
	SetValue(ctx context.Context, key string, value interface{}) error
}

// UserRepository provides access to sfc_users table for setup.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
}

// SiteRepository provides access to sfc_sites table for setup.
type SiteRepository interface {
	Create(ctx context.Context, site *model.Site) error
}

// UserRoleRepository provides access to sfc_user_roles table for setup.
type UserRoleRepository interface {
	AssignRole(ctx context.Context, userID, roleSlug string) error
}

// RoleRepository provides access to sfc_roles table for setup.
type RoleRepository interface {
	GetBySlug(ctx context.Context, slug string) (*model.Role, error)
}
```

**Step 3: Implement repository**

```go
// internal/setup/repository.go
package setup

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// configRepo implements ConfigRepository.
type configRepo struct {
	db *bun.DB
}

func NewConfigRepo(db *bun.DB) ConfigRepository {
	return &configRepo{db: db}
}

func (r *configRepo) GetValue(ctx context.Context, key string) (json.RawMessage, error) {
	var cfg model.Config
	err := r.db.NewSelect().Model(&cfg).Where("key = ?", key).Scan(ctx)
	if err != nil {
		return nil, apperror.NotFound("config key not found", err)
	}
	return cfg.Value, nil
}

func (r *configRepo) SetValue(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal config value: %w", err)
	}

	cfg := &model.Config{
		Key:   key,
		Value: data,
	}

	_, err = r.db.NewInsert().Model(cfg).
		On("CONFLICT (key) DO UPDATE").
		Set("value = EXCLUDED.value").
		Set("updated_at = NOW()").
		Exec(ctx)
	return err
}

// siteRepo implements SiteRepository.
type siteRepo struct {
	db *bun.DB
}

func NewSiteRepo(db *bun.DB) SiteRepository {
	return &siteRepo{db: db}
}

func (r *siteRepo) Create(ctx context.Context, site *model.Site) error {
	_, err := r.db.NewInsert().Model(site).Exec(ctx)
	return err
}

// userRepo implements UserRepository.
type userRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

// userRoleRepo implements UserRoleRepository.
type userRoleRepo struct {
	db *bun.DB
}

func NewUserRoleRepo(db *bun.DB) UserRoleRepository {
	return &userRoleRepo{db: db}
}

func (r *userRoleRepo) AssignRole(ctx context.Context, userID, roleSlug string) error {
	var role model.Role
	err := r.db.NewSelect().Model(&role).Where("slug = ?", roleSlug).Scan(ctx)
	if err != nil {
		return apperror.NotFound("role not found: "+roleSlug, err)
	}

	ur := &model.UserRole{UserID: userID, RoleID: role.ID}
	_, err = r.db.NewInsert().Model(ur).
		On("CONFLICT DO NOTHING").
		Exec(ctx)
	return err
}

// roleRepo implements RoleRepository.
type roleRepo struct {
	db *bun.DB
}

func NewRoleRepo(db *bun.DB) RoleRepository {
	return &roleRepo{db: db}
}

func (r *roleRepo) GetBySlug(ctx context.Context, slug string) (*model.Role, error) {
	var role model.Role
	err := r.db.NewSelect().Model(&role).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		return nil, apperror.NotFound("role not found", err)
	}
	return &role, nil
}
```

**Step 4: Run compilation check**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go build ./internal/setup/...
```

**Step 5: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/setup/dto.go internal/setup/interfaces.go internal/setup/repository.go && git commit -m "feat(setup): add DTOs, interfaces, and repository implementations"
```

---

## Task 9: Setup module — service and install checker

**Files:**
- Rewrite: `internal/setup/service.go`

This service implements both the business logic AND the `middleware.InstallChecker` interface for the installation guard.

**Step 1: Implement service**

```go
// internal/setup/service.go
package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/crypto"
	"github.com/sky-flux/cms/internal/pkg/jwt"
	"github.com/sky-flux/cms/internal/schema"
	"github.com/uptrace/bun"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9_]{3,50}$`)

// Service handles installation check and system initialization.
// It implements middleware.InstallChecker.
type Service struct {
	db           *bun.DB
	rdb          *redis.Client
	jwtMgr       *jwt.Manager
	configRepo   ConfigRepository
	userRepo     UserRepository
	siteRepo     SiteRepository
	userRoleRepo UserRoleRepository
	accessExpiry time.Duration
	installed    atomic.Int32
}

func NewService(
	db *bun.DB,
	rdb *redis.Client,
	jwtMgr *jwt.Manager,
	configRepo ConfigRepository,
	userRepo UserRepository,
	siteRepo SiteRepository,
	userRoleRepo UserRoleRepository,
	accessExpiry time.Duration,
) *Service {
	return &Service{
		db:           db,
		rdb:          rdb,
		jwtMgr:       jwtMgr,
		configRepo:   configRepo,
		userRepo:     userRepo,
		siteRepo:     siteRepo,
		userRoleRepo: userRoleRepo,
		accessExpiry: accessExpiry,
	}
}

// IsInstalled implements middleware.InstallChecker.
// Triple-check: atomic → Redis → DB.
func (s *Service) IsInstalled(ctx context.Context) bool {
	// L1: atomic variable
	if s.installed.Load() == 1 {
		return true
	}

	// L2: Redis
	val, err := s.rdb.Get(ctx, "system:installed").Result()
	if err == nil && val == "true" {
		s.installed.Store(1)
		return true
	}

	// L3: Database
	raw, err := s.configRepo.GetValue(ctx, "system.installed")
	if err != nil {
		return false
	}

	var installed bool
	if json.Unmarshal(raw, &installed) == nil && installed {
		s.rdb.Set(ctx, "system:installed", "true", 0)
		s.installed.Store(1)
		return true
	}

	return false
}

// MarkInstalled sets the in-memory flag. Called after successful initialization.
func (s *Service) MarkInstalled() {
	s.installed.Store(1)
}

// Check returns the installation status.
func (s *Service) Check(ctx context.Context) bool {
	return s.IsInstalled(ctx)
}

// Initialize performs system initialization in a single transaction.
func (s *Service) Initialize(ctx context.Context, req *InitializeReq) (*InitializeResp, error) {
	if s.IsInstalled(ctx) {
		return nil, apperror.Conflict("system already installed", nil)
	}

	if !slugRegex.MatchString(req.SiteSlug) {
		return nil, apperror.Validation("invalid site slug: must match ^[a-z0-9_]{3,50}$", nil)
	}

	// Hash password
	passwordHash, err := crypto.HashPassword(req.AdminPassword)
	if err != nil {
		return nil, apperror.Internal("hash password failed", err)
	}

	locale := req.Locale
	if locale == "" {
		locale = "zh-CN"
	}

	var user model.User
	var site model.Site

	// Run initialization in a transaction
	err = s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Advisory lock to prevent concurrent initialization
		var locked bool
		if err := tx.QueryRowContext(ctx, "SELECT pg_try_advisory_xact_lock(1)").Scan(&locked); err != nil {
			return fmt.Errorf("advisory lock: %w", err)
		}
		if !locked {
			return apperror.Conflict("initialization already in progress", nil)
		}

		// Double-check not installed within the lock
		var cfg model.Config
		err := tx.NewSelect().Model(&cfg).Where("key = ?", "system.installed").Scan(ctx)
		if err == nil {
			var installed bool
			if json.Unmarshal(cfg.Value, &installed) == nil && installed {
				return apperror.Conflict("system already installed", nil)
			}
		}

		// Create admin user
		user = model.User{
			Email:        req.AdminEmail,
			PasswordHash: passwordHash,
			DisplayName:  req.AdminDisplayName,
			IsActive:     true,
		}
		if _, err := tx.NewInsert().Model(&user).Exec(ctx); err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}

		// Create site
		site = model.Site{
			Name:          req.SiteName,
			Slug:          req.SiteSlug,
			Domain:        req.SiteURL,
			DefaultLocale: locale,
			IsActive:      true,
		}
		if _, err := tx.NewInsert().Model(&site).Exec(ctx); err != nil {
			return fmt.Errorf("create site: %w", err)
		}

		// Assign super role
		var role model.Role
		if err := tx.NewSelect().Model(&role).Where("slug = ?", "super").Scan(ctx); err != nil {
			return fmt.Errorf("find super role: %w", err)
		}

		ur := &model.UserRole{UserID: user.ID, RoleID: role.ID}
		if _, err := tx.NewInsert().Model(ur).Exec(ctx); err != nil {
			return fmt.Errorf("assign super role: %w", err)
		}

		// Mark as installed
		installedValue, _ := json.Marshal(true)
		_, err = tx.NewUpdate().Model(&model.Config{}).
			Set("value = ?", json.RawMessage(installedValue)).
			Where("key = ?", "system.installed").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("set installed flag: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Create site schema (outside transaction — schema.CreateSiteSchema manages its own tx)
	if err := schema.CreateSiteSchema(ctx, s.db, req.SiteSlug); err != nil {
		slog.Error("create site schema failed after install", "error", err, "slug", req.SiteSlug)
		return nil, apperror.Internal("site schema creation failed", err)
	}

	// Update caches
	s.rdb.Set(ctx, "system:installed", "true", 0)
	s.installed.Store(1)

	// Issue JWT
	token, err := s.jwtMgr.SignAccessToken(user.ID)
	if err != nil {
		return nil, apperror.Internal("sign token failed", err)
	}

	return &InitializeResp{
		User: UserResp{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
		Site: SiteResp{
			ID:   site.ID,
			Name: site.Name,
			Slug: site.Slug,
		},
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.accessExpiry.Seconds()),
	}, nil
}
```

**Step 2: Run compilation check**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go build ./internal/setup/...
```

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/setup/service.go && git commit -m "feat(setup): add service with triple-check install guard and transactional init"
```

---

## Task 10: Setup module — handler + tests

**Files:**
- Rewrite: `internal/setup/handler.go`
- Rewrite: `internal/setup/handler_test.go`

**Step 1: Implement handler**

```go
// internal/setup/handler.go
package setup

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler provides HTTP endpoints for system setup.
type Handler struct {
	svc *Service
}

// NewHandler creates a setup handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Check handles POST /api/v1/setup/check.
func (h *Handler) Check(c *gin.Context) {
	installed := h.svc.Check(c.Request.Context())
	response.Success(c, CheckResp{Installed: installed})
}

// Initialize handles POST /api/v1/setup/initialize.
func (h *Handler) Initialize(c *gin.Context) {
	var req InitializeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	resp, err := h.svc.Initialize(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, resp)
}
```

**Step 2: Write handler tests** (mock-based, no real DB)

```go
// internal/setup/handler_test.go
package setup_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/setup"
	"github.com/stretchr/testify/assert"
)

// Note: Handler tests are integration-level since Service requires DB.
// These tests verify request parsing and response formatting.
// Full integration tests will be in Task 15.

func TestInitializeReq_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body map[string]string
		code int
	}{
		{
			name: "missing required fields",
			body: map[string]string{},
			code: 422,
		},
		{
			name: "invalid email",
			body: map[string]string{
				"site_name": "Blog", "site_slug": "blog", "site_url": "https://blog.com",
				"admin_email": "not-an-email", "admin_password": "Pass123!", "admin_display_name": "Admin",
			},
			code: 422,
		},
		{
			name: "password too short",
			body: map[string]string{
				"site_name": "Blog", "site_slug": "blog", "site_url": "https://blog.com",
				"admin_email": "a@b.com", "admin_password": "short", "admin_display_name": "Admin",
			},
			code: 422,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			// Use a dummy handler that only tests DTO binding
			r.POST("/api/v1/setup/initialize", func(c *gin.Context) {
				var req setup.InitializeReq
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(422, gin.H{"success": false, "error": err.Error()})
					return
				}
				c.JSON(200, gin.H{"success": true})
			})

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/setup/initialize", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.code, w.Code)
		})
	}
}
```

**Step 3: Run tests**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/setup/ -v
```

**Step 4: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/setup/handler.go internal/setup/handler_test.go && git commit -m "feat(setup): add handler with Check and Initialize endpoints"
```

---

## Task 11: Auth module — DTOs and interfaces

**Files:**
- Rewrite: `internal/auth/dto.go`
- Create: `internal/auth/interfaces.go`

**Step 1: Implement DTOs**

See design doc section 5 for full DTO list. Create comprehensive request/response types for all 17 endpoints:
- LoginReq, LoginResp, Login2FAResp
- RefreshResp
- MeResp (with roles + sites)
- ChangePasswordReq
- ForgotPasswordReq, ResetPasswordReq
- Setup2FAResp, Verify2FAReq, Validate2FAReq, Disable2FAReq
- RegenerateBackupCodesReq, Get2FAStatusResp
- ForceDisable2FAReq

**Step 2: Implement interfaces**

See design doc section 5 for full interface definitions:
- UserRepository (GetByEmail, GetByID, UpdatePassword, UpdateLastLogin, Create, ListForAdmin)
- TokenRepository (CreateRefreshToken, GetRefreshTokenByHash, RevokeRefreshToken, RevokeAllUserTokens, CreatePasswordResetToken, GetPasswordResetTokenByHash, MarkPasswordResetTokenUsed)
- TOTPRepository (GetByUserID, Upsert, Enable, Delete, UpdateBackupCodes)

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/auth/dto.go internal/auth/interfaces.go && git commit -m "feat(auth): add DTOs and repository interfaces for all 17 endpoints"
```

---

## Task 12: Auth module — repository implementation

**Files:**
- Rewrite: `internal/auth/repository.go`

Implement all three repository interfaces using uptrace/bun:
- `authUserRepo` implements `UserRepository`
- `authTokenRepo` implements `TokenRepository`
- `authTOTPRepo` implements `TOTPRepository`

Follow the patterns in `internal/setup/repository.go` and `internal/rbac/` repos. Use `apperror.NotFound()` for missing records.

**Step 1: Implement repository**

**Step 2: Compilation check**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go build ./internal/auth/...
```

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/auth/repository.go && git commit -m "feat(auth): add repository implementations for user, token, and TOTP"
```

---

## Task 13: Auth module — service (core auth)

**Files:**
- Rewrite: `internal/auth/service.go`

Implement the auth service with all business logic:
- Login flow (lockout check, bcrypt, 2FA detection, token issuance)
- Refresh flow (validate refresh token, issue new access token)
- Logout (blacklist access token, revoke refresh token)
- Me (get user profile with roles and sites)
- ChangePassword (verify current, hash new, revoke all tokens)
- ForgotPassword (generate reset token, **stub** email sending for now)
- ResetPassword (verify token, update password, revoke tokens)
- 2FA: Setup, Verify, Validate, Disable, RegenerateBackupCodes, GetStatus, ForceDisable

Dependencies: jwt.Manager, Redis client, all 3 repositories, config (TOTP encryption key, cookie settings).

**Step 1: Implement service**

**Step 2: Compilation check**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go build ./internal/auth/...
```

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/auth/service.go && git commit -m "feat(auth): add service with login, 2FA, password reset, and token management"
```

---

## Task 14: Auth module — service tests

**Files:**
- Create: `internal/auth/service_test.go`

Write comprehensive unit tests with mock repositories using testify/mock or manual mocks (follow the pattern in `internal/middleware/rbac_test.go` with simple struct mocks).

Key test scenarios:
- Login: success, wrong password, locked out, inactive user, 2FA required
- Refresh: success, revoked token, expired token
- Logout: success (blacklists JTI)
- ChangePassword: success, wrong current password
- 2FA Setup: success, already enabled
- 2FA Validate: success, invalid code, backup code, replay detection
- ForgotPassword: success, unknown email (still returns OK)
- ResetPassword: success, expired token, used token

**Step 1: Write tests**

**Step 2: Run tests**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/auth/ -v -count=1
```

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/auth/service_test.go && git commit -m "test(auth): add comprehensive service unit tests with mock repos"
```

---

## Task 15: Auth module — handler

**Files:**
- Rewrite: `internal/auth/handler.go`

Implement handler with 17 methods. Follow the pattern in `internal/rbac/handler.go`:
- Parse request → call service → write response
- Use `response.Success()`, `response.Created()`, `response.Error()`
- Set refresh token cookie in Login and Validate2FA handlers

Handler methods:
1. Login, Refresh, Logout, Me, ChangePassword
2. ForgotPassword, ResetPassword
3. Setup2FA, Verify2FA, Validate2FA, Disable2FA, RegenerateBackupCodes, Get2FAStatus, ForceDisable2FA

**Special: Refresh token cookie helper**

```go
func setRefreshTokenCookie(c *gin.Context, token string, maxAge int) {
	c.SetCookie("refresh_token", token, maxAge, "/api/v1/auth", "", true, true)
	c.SetSameSite(http.SameSiteLaxMode)
}
```

**Step 1: Implement handler**

**Step 2: Compilation check**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go build ./internal/auth/...
```

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/auth/handler.go && git commit -m "feat(auth): add HTTP handlers for all 17 auth endpoints"
```

---

## Task 16: Auth module — handler tests

**Files:**
- Rewrite: `internal/auth/handler_test.go`

Write HTTP tests for all handler methods. Test request parsing, response format, and HTTP status codes. Use mock service (interface-based).

**Step 1: Write tests**

**Step 2: Run tests**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./internal/auth/ -v -count=1
```

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/auth/handler_test.go && git commit -m "test(auth): add HTTP handler tests for all 17 auth endpoints"
```

---

## Task 17: Router integration

**Files:**
- Modify: `internal/router/router.go`

Update the router `Setup()` function to:
1. Accept `*config.Config` parameter for JWT config
2. Create jwt.Manager
3. Create setup repos and service (which implements InstallChecker)
4. Create auth repos and service
5. Add InstallationGuard middleware (exempt: `/health`, `/api/v1/setup/`)
6. Register setup routes (no auth): `POST /api/v1/setup/check`, `POST /api/v1/setup/initialize`
7. Register auth public routes (no auth): login, refresh, forgot-password, reset-password
8. Register auth protected routes (JWT): logout, me, password, 2fa/*
9. Register auth admin route (JWT + RBAC): `DELETE /api/v1/auth/2fa/users/:user_id`

**Step 1: Update router**

**Step 2: Compilation check**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go build ./...
```

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add internal/router/router.go && git commit -m "feat(router): register setup and auth endpoints with middleware chain"
```

---

## Task 18: Full test suite verification

**Step 1: Run all tests**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go test ./... -v -count=1
```

Verify ALL tests pass — crypto, jwt, middleware, setup, auth.

**Step 2: Run vet and compilation**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go vet ./...
```

**Step 3: Final commit if any fixes needed**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add -A && git commit -m "fix: resolve test and compilation issues from batch 1 integration"
```

---

## Summary

| Task | Component | Files | Tests |
|------|-----------|-------|-------|
| 1 | TOTP dependency | go.mod | — |
| 2 | crypto/password | 2 | 4 tests |
| 3 | crypto/token | 2 | 6 tests |
| 4 | crypto/totp | 2 | 6 tests |
| 5 | pkg/jwt | 2 | 6 tests |
| 6 | middleware/auth | 2 | 5 tests |
| 7 | middleware/installation_guard | 2 | 4 tests |
| 8 | setup/dto+interfaces+repo | 3 | — |
| 9 | setup/service | 1 | — |
| 10 | setup/handler+tests | 2 | 3 tests |
| 11 | auth/dto+interfaces | 2 | — |
| 12 | auth/repository | 1 | — |
| 13 | auth/service | 1 | — |
| 14 | auth/service tests | 1 | ~15 tests |
| 15 | auth/handler | 1 | — |
| 16 | auth/handler tests | 1 | ~17 tests |
| 17 | router integration | 1 (modify) | — |
| 18 | full verification | — | all |

**Total: ~25 new/modified files, ~66 tests, 18 commits**

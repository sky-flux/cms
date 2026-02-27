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
	mgr := jwt.NewManager(testSecret, 15*time.Minute, 5*time.Minute, 7*24*time.Hour, rdb)
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
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:0"})
	mgr := jwt.NewManager(testSecret, -1*time.Second, -1*time.Second, -1*time.Second, rdb)
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

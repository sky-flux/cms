package apikey_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/apikey"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: APIKeyRepository
// ---------------------------------------------------------------------------

type mockRepo struct {
	listKeys  []model.APIKey
	listErr   error
	getKey    *model.APIKey
	getErr    error
	createErr error
	revokeErr error

	createdKey *model.APIKey
}

func (m *mockRepo) List(_ context.Context) ([]model.APIKey, error) {
	return m.listKeys, m.listErr
}

func (m *mockRepo) GetByID(_ context.Context, _ string) (*model.APIKey, error) {
	return m.getKey, m.getErr
}

func (m *mockRepo) Create(_ context.Context, key *model.APIKey) error {
	if m.createErr == nil {
		key.ID = "ak-new-id"
		key.CreatedAt = time.Now()
		m.createdKey = key
	}
	return m.createErr
}

func (m *mockRepo) Revoke(_ context.Context, _ string) error {
	return m.revokeErr
}

// ---------------------------------------------------------------------------
// Mock: audit.Logger
// ---------------------------------------------------------------------------

type mockAuditor struct {
	logged []audit.Entry
	err    error
}

func (m *mockAuditor) Log(_ context.Context, entry audit.Entry) error {
	m.logged = append(m.logged, entry)
	return m.err
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

type testEnv struct {
	svc     *apikey.Service
	repo    *mockRepo
	auditor *mockAuditor
}

func newTestEnv() *testEnv {
	r := &mockRepo{}
	a := &mockAuditor{}
	return &testEnv{
		svc:     apikey.NewService(r, a),
		repo:    r,
		auditor: a,
	}
}

// ---------------------------------------------------------------------------
// Tests: ListAPIKeys
// ---------------------------------------------------------------------------

func TestService_ListAPIKeys_Success(t *testing.T) {
	env := newTestEnv()
	now := time.Now()
	env.repo.listKeys = []model.APIKey{
		{ID: "ak-1", Name: "Key 1", Status: model.APIKeyStatusActive, CreatedAt: now},
		{ID: "ak-2", Name: "Key 2", Status: model.APIKeyStatusRevoked, CreatedAt: now.Add(-time.Hour)},
	}

	keys, err := env.svc.ListAPIKeys(context.Background())
	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestService_ListAPIKeys_Empty(t *testing.T) {
	env := newTestEnv()
	env.repo.listKeys = []model.APIKey{}

	keys, err := env.svc.ListAPIKeys(context.Background())
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestService_ListAPIKeys_RepoError(t *testing.T) {
	env := newTestEnv()
	env.repo.listErr = errors.New("db error")

	_, err := env.svc.ListAPIKeys(context.Background())
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: CreateAPIKey
// ---------------------------------------------------------------------------

func TestService_CreateAPIKey_Success(t *testing.T) {
	env := newTestEnv()

	key, plainKey, err := env.svc.CreateAPIKey(context.Background(), "owner-1", &apikey.CreateAPIKeyReq{
		Name: "My API Key",
	})
	require.NoError(t, err)
	assert.Equal(t, "ak-new-id", key.ID)
	assert.Equal(t, "My API Key", key.Name)
	assert.Equal(t, "owner-1", key.OwnerID)
	assert.Equal(t, model.APIKeyStatusActive, key.Status)

	// PlainKey should start with cms_live_ prefix
	assert.True(t, strings.HasPrefix(plainKey, "cms_live_"))
	// PlainKey should be cms_live_ + 64 hex chars (32 bytes)
	assert.Len(t, plainKey, len("cms_live_")+64)

	// KeyPrefix should start with cms_live_ followed by first 8 chars of hex
	assert.True(t, strings.HasPrefix(key.KeyPrefix, "cms_live_"))
	assert.Len(t, key.KeyPrefix, len("cms_live_")+8)

	// Audit log should be recorded
	require.Len(t, env.auditor.logged, 1)
	assert.Equal(t, model.LogActionCreate, env.auditor.logged[0].Action)
	assert.Equal(t, "api_key", env.auditor.logged[0].ResourceType)
}

func TestService_CreateAPIKey_WithExpiry(t *testing.T) {
	env := newTestEnv()
	expires := time.Now().Add(30 * 24 * time.Hour)

	key, _, err := env.svc.CreateAPIKey(context.Background(), "owner-1", &apikey.CreateAPIKeyReq{
		Name:      "Expiring Key",
		ExpiresAt: &expires,
	})
	require.NoError(t, err)
	assert.NotNil(t, key.ExpiresAt)
}

func TestService_CreateAPIKey_RepoError(t *testing.T) {
	env := newTestEnv()
	env.repo.createErr = errors.New("db error")

	_, _, err := env.svc.CreateAPIKey(context.Background(), "owner-1", &apikey.CreateAPIKeyReq{
		Name: "Fail Key",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create apikey insert")
}

func TestService_CreateAPIKey_AuditFailDoesNotBlock(t *testing.T) {
	env := newTestEnv()
	env.auditor.err = errors.New("audit fail")

	key, _, err := env.svc.CreateAPIKey(context.Background(), "owner-1", &apikey.CreateAPIKeyReq{
		Name: "Key With Audit Fail",
	})
	require.NoError(t, err)
	assert.Equal(t, "ak-new-id", key.ID)
}

// ---------------------------------------------------------------------------
// Tests: RevokeAPIKey
// ---------------------------------------------------------------------------

func TestService_RevokeAPIKey_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.getKey = &model.APIKey{
		ID:     "ak-1",
		Name:   "Key 1",
		Status: model.APIKeyStatusActive,
	}

	err := env.svc.RevokeAPIKey(context.Background(), "ak-1")
	require.NoError(t, err)

	// Audit log
	require.Len(t, env.auditor.logged, 1)
	assert.Equal(t, model.LogActionDelete, env.auditor.logged[0].Action)
}

func TestService_RevokeAPIKey_NotFound(t *testing.T) {
	env := newTestEnv()
	env.repo.getErr = apperror.NotFound("api key not found", nil)

	err := env.svc.RevokeAPIKey(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

func TestService_RevokeAPIKey_AlreadyRevoked(t *testing.T) {
	env := newTestEnv()
	revoked := time.Now()
	env.repo.getKey = &model.APIKey{
		ID:        "ak-1",
		Status:    model.APIKeyStatusRevoked,
		RevokedAt: &revoked,
	}

	err := env.svc.RevokeAPIKey(context.Background(), "ak-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestService_RevokeAPIKey_RepoError(t *testing.T) {
	env := newTestEnv()
	env.repo.getKey = &model.APIKey{
		ID:     "ak-1",
		Status: model.APIKeyStatusActive,
	}
	env.repo.revokeErr = errors.New("db error")

	err := env.svc.RevokeAPIKey(context.Background(), "ak-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "revoke apikey")
}

func TestService_RevokeAPIKey_AuditFailDoesNotBlock(t *testing.T) {
	env := newTestEnv()
	env.repo.getKey = &model.APIKey{
		ID:     "ak-1",
		Status: model.APIKeyStatusActive,
	}
	env.auditor.err = errors.New("audit fail")

	err := env.svc.RevokeAPIKey(context.Background(), "ak-1")
	require.NoError(t, err)
}

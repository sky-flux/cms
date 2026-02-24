package system_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: ConfigRepository
// ---------------------------------------------------------------------------

type mockConfigRepo struct {
	configs   []model.SiteConfig
	listErr   error
	getConfig *model.SiteConfig
	getErr    error
	upsertErr error
}

func (m *mockConfigRepo) List(_ context.Context) ([]model.SiteConfig, error) {
	return m.configs, m.listErr
}

func (m *mockConfigRepo) GetByKey(_ context.Context, _ string) (*model.SiteConfig, error) {
	return m.getConfig, m.getErr
}

func (m *mockConfigRepo) Upsert(_ context.Context, _ *model.SiteConfig) error {
	return m.upsertErr
}

// ---------------------------------------------------------------------------
// Mock: audit.Logger
// ---------------------------------------------------------------------------

type mockAuditor struct {
	lastEntry *audit.Entry
	err       error
}

func (m *mockAuditor) Log(_ context.Context, entry audit.Entry) error {
	m.lastEntry = &entry
	return m.err
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

type testEnv struct {
	svc     *system.Service
	repo    *mockConfigRepo
	auditor *mockAuditor
}

func newTestEnv() *testEnv {
	repo := &mockConfigRepo{}
	aud := &mockAuditor{}
	return &testEnv{
		svc:     system.NewService(repo, aud),
		repo:    repo,
		auditor: aud,
	}
}

func testConfig() *model.SiteConfig {
	return &model.SiteConfig{
		Key:         "site_title",
		Value:       json.RawMessage(`"My Blog"`),
		Description: "Site title",
		UpdatedAt:   time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Tests: ListConfigs
// ---------------------------------------------------------------------------

func TestService_ListConfigs_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.configs = []model.SiteConfig{*testConfig()}

	configs, err := env.svc.ListConfigs(context.Background())
	require.NoError(t, err)
	assert.Len(t, configs, 1)
	assert.Equal(t, "site_title", configs[0].Key)
}

func TestService_ListConfigs_Empty(t *testing.T) {
	env := newTestEnv()
	env.repo.configs = []model.SiteConfig{}

	configs, err := env.svc.ListConfigs(context.Background())
	require.NoError(t, err)
	assert.Empty(t, configs)
}

func TestService_ListConfigs_RepoError(t *testing.T) {
	env := newTestEnv()
	env.repo.listErr = errors.New("db error")

	_, err := env.svc.ListConfigs(context.Background())
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: UpdateConfig
// ---------------------------------------------------------------------------

func TestService_UpdateConfig_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.getConfig = testConfig()

	req := &system.UpdateConfigReq{
		Value: json.RawMessage(`"New Title"`),
	}

	cfg, err := env.svc.UpdateConfig(context.Background(), "site_title", req, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "site_title", cfg.Key)
	assert.JSONEq(t, `"New Title"`, string(cfg.Value))

	// Verify audit was called
	require.NotNil(t, env.auditor.lastEntry)
	assert.Equal(t, model.LogActionSettingsChange, env.auditor.lastEntry.Action)
	assert.Equal(t, "site_config", env.auditor.lastEntry.ResourceType)
	assert.Equal(t, "site_title", env.auditor.lastEntry.ResourceID)
}

func TestService_UpdateConfig_AuditSnapshot(t *testing.T) {
	env := newTestEnv()
	env.repo.getConfig = testConfig()

	req := &system.UpdateConfigReq{
		Value: json.RawMessage(`"Updated"`),
	}

	_, err := env.svc.UpdateConfig(context.Background(), "site_title", req, "user-1")
	require.NoError(t, err)

	snapshot, ok := env.auditor.lastEntry.ResourceSnapshot.(map[string]json.RawMessage)
	require.True(t, ok)
	assert.JSONEq(t, `"My Blog"`, string(snapshot["old_value"]))
	assert.JSONEq(t, `"Updated"`, string(snapshot["new_value"]))
}

func TestService_UpdateConfig_NotFound(t *testing.T) {
	env := newTestEnv()
	env.repo.getErr = apperror.NotFound("config not found", nil)

	req := &system.UpdateConfigReq{
		Value: json.RawMessage(`"val"`),
	}

	_, err := env.svc.UpdateConfig(context.Background(), "nonexistent", req, "user-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

func TestService_UpdateConfig_UpsertError(t *testing.T) {
	env := newTestEnv()
	env.repo.getConfig = testConfig()
	env.repo.upsertErr = errors.New("db error")

	req := &system.UpdateConfigReq{
		Value: json.RawMessage(`"val"`),
	}

	_, err := env.svc.UpdateConfig(context.Background(), "site_title", req, "user-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update config")
}

func TestService_UpdateConfig_AuditError_NonFatal(t *testing.T) {
	env := newTestEnv()
	env.repo.getConfig = testConfig()
	env.auditor.err = errors.New("audit failed")

	req := &system.UpdateConfigReq{
		Value: json.RawMessage(`"val"`),
	}

	// Should still succeed even if audit fails
	cfg, err := env.svc.UpdateConfig(context.Background(), "site_title", req, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "site_title", cfg.Key)
}

func TestService_UpdateConfig_SetsUpdatedBy(t *testing.T) {
	env := newTestEnv()
	cfg := testConfig()
	env.repo.getConfig = cfg

	req := &system.UpdateConfigReq{
		Value: json.RawMessage(`"val"`),
	}

	result, err := env.svc.UpdateConfig(context.Background(), "site_title", req, "user-42")
	require.NoError(t, err)
	require.NotNil(t, result.UpdatedBy)
	assert.Equal(t, "user-42", *result.UpdatedBy)
}

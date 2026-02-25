package cron

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockSiteLister struct {
	slugs []string
	err   error
}

func (m *mockSiteLister) ListActiveSlugs(_ context.Context) ([]string, error) {
	return m.slugs, m.err
}

type mockSchemaExecutor struct {
	setCalls   []string
	resetCalls int
	setErr     error
	resetErr   error
}

func (m *mockSchemaExecutor) SetSearchPath(_ context.Context, slug string) error {
	m.setCalls = append(m.setCalls, slug)
	return m.setErr
}

func (m *mockSchemaExecutor) ResetSearchPath(_ context.Context) error {
	m.resetCalls++
	return m.resetErr
}

type mockScheduledPublisher struct {
	count int64
	err   error
}

func (m *mockScheduledPublisher) PublishDue(_ context.Context) (int64, error) {
	return m.count, m.err
}

type mockTokenCleaner struct {
	count int64
	err   error
}

func (m *mockTokenCleaner) DeleteExpired(_ context.Context) (int64, error) {
	return m.count, m.err
}

type mockSoftDeletePurger struct {
	count int64
	err   error
}

func (m *mockSoftDeletePurger) PurgeOlderThan(_ context.Context, _ int) (int64, error) {
	return m.count, m.err
}

// --- RunScheduledPublishing tests ---

func TestRunScheduledPublishing_PublishesAcrossSites(t *testing.T) {
	sites := &mockSiteLister{slugs: []string{"blog", "shop"}}
	schema := &mockSchemaExecutor{}
	pub := &mockScheduledPublisher{count: 3}

	total, err := RunScheduledPublishing(context.Background(), sites, schema, pub)

	require.NoError(t, err)
	assert.Equal(t, int64(6), total) // 3 per site × 2 sites
	assert.Equal(t, []string{"blog", "shop"}, schema.setCalls)
	assert.Equal(t, 2, schema.resetCalls)
}

func TestRunScheduledPublishing_NoSites(t *testing.T) {
	sites := &mockSiteLister{slugs: []string{}}
	schema := &mockSchemaExecutor{}
	pub := &mockScheduledPublisher{count: 0}

	total, err := RunScheduledPublishing(context.Background(), sites, schema, pub)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

func TestRunScheduledPublishing_SiteListError(t *testing.T) {
	sites := &mockSiteLister{err: errors.New("db down")}
	schema := &mockSchemaExecutor{}
	pub := &mockScheduledPublisher{}

	_, err := RunScheduledPublishing(context.Background(), sites, schema, pub)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list active sites")
}

func TestRunScheduledPublishing_ContinuesOnPublishError(t *testing.T) {
	sites := &mockSiteLister{slugs: []string{"blog", "shop"}}
	schema := &mockSchemaExecutor{}
	pub := &mockScheduledPublisher{err: errors.New("publish failed")}

	total, err := RunScheduledPublishing(context.Background(), sites, schema, pub)

	// Should not return error — logs it and continues to next site.
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, 2, schema.resetCalls) // still resets for each site
}

// --- RunTokenCleanup tests ---

func TestRunTokenCleanup_CleansAcrossSites(t *testing.T) {
	sites := &mockSiteLister{slugs: []string{"blog", "shop"}}
	schema := &mockSchemaExecutor{}
	cleaner := &mockTokenCleaner{count: 5}

	total, err := RunTokenCleanup(context.Background(), sites, schema, cleaner)

	require.NoError(t, err)
	assert.Equal(t, int64(10), total) // 5 per site × 2
}

func TestRunTokenCleanup_ContinuesOnError(t *testing.T) {
	sites := &mockSiteLister{slugs: []string{"blog"}}
	schema := &mockSchemaExecutor{}
	cleaner := &mockTokenCleaner{err: errors.New("clean failed")}

	total, err := RunTokenCleanup(context.Background(), sites, schema, cleaner)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

// --- RunSoftDeletePurge tests ---

func TestRunSoftDeletePurge_PurgesAcrossSites(t *testing.T) {
	sites := &mockSiteLister{slugs: []string{"blog"}}
	schema := &mockSchemaExecutor{}
	purger := &mockSoftDeletePurger{count: 10}

	total, err := RunSoftDeletePurge(context.Background(), sites, schema, purger, 30)

	require.NoError(t, err)
	assert.Equal(t, int64(10), total)
}

func TestRunSoftDeletePurge_DefaultRetention(t *testing.T) {
	sites := &mockSiteLister{slugs: []string{"blog"}}
	schema := &mockSchemaExecutor{}
	purger := &mockSoftDeletePurger{count: 2}

	// retentionDays <= 0 should use default 30
	total, err := RunSoftDeletePurge(context.Background(), sites, schema, purger, 0)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
}

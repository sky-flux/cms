package dashboard

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStatsReader struct {
	stats *DashboardStats
	err   error
}

func (m *mockStatsReader) GetStats(_ context.Context, _ string) (*DashboardStats, error) {
	return m.stats, m.err
}

func TestService_GetStats_DelegatesToRepo(t *testing.T) {
	expected := &DashboardStats{
		Posts: PostStats{Total: 42, Published: 30, Draft: 10, Scheduled: 2},
	}
	mock := &mockStatsReader{stats: expected}
	svc := NewService(mock)

	result, err := svc.GetStats(context.Background(), "site_test")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestService_GetStats_PropagatesError(t *testing.T) {
	mock := &mockStatsReader{err: assert.AnError}
	svc := NewService(mock)

	result, err := svc.GetStats(context.Background(), "site_test")
	assert.Error(t, err)
	assert.Nil(t, result)
}

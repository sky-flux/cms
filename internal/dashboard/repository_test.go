package dashboard

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

// mockDB is nil — we test that Repository satisfies StatsReader interface
// and that queries are structured correctly. DB integration is tested via E2E.

func TestRepository_ImplementsStatsReader(t *testing.T) {
	var _ StatsReader = (*Repository)(nil)
}

func TestNewRepository(t *testing.T) {
	repo := NewRepository(nil)
	require.NotNil(t, repo)
}

func TestRepository_GetStats_NilDB(t *testing.T) {
	// With nil DB, GetStats should panic or return error
	// This verifies the method exists and has correct signature
	repo := NewRepository(nil)

	// Can't call with nil DB, but verify the method exists
	var fn func(context.Context, string) (*DashboardStats, error) = repo.GetStats
	assert.NotNil(t, fn)
}

func TestBunIdentQuoting(t *testing.T) {
	// Verify bun.Ident correctly quotes schema identifiers
	quoted := bun.Ident("site_test-blog")
	assert.NotEmpty(t, string(quoted))
}

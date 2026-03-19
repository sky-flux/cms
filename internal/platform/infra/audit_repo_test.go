package infra_test

import (
	"context"
	"os"
	"testing"

	"github.com/sky-flux/cms/internal/platform/domain"
	"github.com/sky-flux/cms/internal/platform/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func openTestDB(t *testing.T) *bun.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	// Use testcontainers-go or a pre-existing test DB.
	// For now we connect to the env-provided DSN directly.
	db, err := infra.OpenBunDB(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestBunAuditRepo_Save_And_List(t *testing.T) {
	db := openTestDB(t)
	repo := infra.NewBunAuditRepository(db)
	ctx := context.Background()

	entry, err := domain.NewAuditEntry("user-1", domain.AuditActionCreate, "post", "post-1", "127.0.0.1", "Go-Test")
	require.NoError(t, err)

	err = repo.Save(ctx, entry)
	require.NoError(t, err)

	items, total, err := repo.List(ctx, domain.AuditFilter{Page: 1, PerPage: 10, UserID: "user-1"})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.NotEmpty(t, items)
}

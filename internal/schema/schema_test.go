package schema_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/schema"
	"github.com/sky-flux/cms/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func setupDB(t *testing.T) *bun.DB {
	t.Helper()
	pg := testutil.SetupPostgres(t)
	return pg.DB
}

func tableExists(t *testing.T, db *bun.DB, schemaName, tableName string) bool {
	t.Helper()
	var exists bool
	err := db.NewSelect().
		ColumnExpr("EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = ? AND table_name = ?)", schemaName, tableName).
		Scan(context.Background(), &exists)
	require.NoError(t, err)
	return exists
}

func schemaExists(t *testing.T, db *bun.DB, schemaName string) bool {
	t.Helper()
	var exists bool
	err := db.NewSelect().
		ColumnExpr("EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = ?)", schemaName).
		Scan(context.Background(), &exists)
	require.NoError(t, err)
	return exists
}

func TestCreateSiteSchema_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}
	db := setupDB(t)
	ctx := context.Background()

	err := schema.CreateSiteSchema(ctx, db, "test_blog")
	require.NoError(t, err)

	assert.True(t, schemaExists(t, db, "site_test_blog"))

	expectedTables := []string{
		"sfc_site_posts",
		"sfc_site_categories",
		"sfc_site_tags",
		"sfc_site_media_files",
		"sfc_site_comments",
		"sfc_site_menus",
		"sfc_site_redirects",
		"sfc_site_preview_tokens",
		"sfc_site_api_keys",
		"sfc_site_audits",
		"sfc_site_configs",
	}
	for _, table := range expectedTables {
		assert.True(t, tableExists(t, db, "site_test_blog", table), "table %s should exist", table)
	}

	_ = schema.DropSiteSchema(ctx, db, "test_blog")
}

func TestCreateSiteSchema_InvalidSlug(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}
	db := setupDB(t)

	err := schema.CreateSiteSchema(context.Background(), db, "INVALID")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid site slug")
}

func TestCreateSiteSchema_Idempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}
	db := setupDB(t)
	ctx := context.Background()

	require.NoError(t, schema.CreateSiteSchema(ctx, db, "idem_test"))
	require.NoError(t, schema.CreateSiteSchema(ctx, db, "idem_test"))

	_ = schema.DropSiteSchema(ctx, db, "idem_test")
}

func TestDropSiteSchema_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}
	db := setupDB(t)
	ctx := context.Background()

	require.NoError(t, schema.CreateSiteSchema(ctx, db, "drop_test"))
	assert.True(t, schemaExists(t, db, "site_drop_test"))

	require.NoError(t, schema.DropSiteSchema(ctx, db, "drop_test"))
	assert.False(t, schemaExists(t, db, "site_drop_test"))
}

func TestDropSiteSchema_NonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}
	db := setupDB(t)

	err := schema.DropSiteSchema(context.Background(), db, "nonexistent_xyz")
	require.NoError(t, err)
}

func TestDropSiteSchema_InvalidSlug(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}
	db := setupDB(t)

	err := schema.DropSiteSchema(context.Background(), db, "BAD")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid site slug")
}

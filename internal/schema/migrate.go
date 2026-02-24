package schema

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

// CreateSiteSchema creates a new site schema with all content tables.
// It validates the slug, creates the schema, then executes the template DDL
// with {schema} replaced by the actual schema name (site_{slug}).
func CreateSiteSchema(ctx context.Context, db *bun.DB, slug string) error {
	if !ValidateSlug(slug) {
		return fmt.Errorf("invalid site slug: %q", slug)
	}

	schemaName := "site_" + slug

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create the schema
	if _, err := tx.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", bun.Ident(schemaName))); err != nil {
		return fmt.Errorf("create schema %s: %w", schemaName, err)
	}

	// Execute the template DDL with schema name substituted
	ddl := strings.ReplaceAll(siteTemplateDDL, "{schema}", schemaName)
	if _, err := tx.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("create tables in %s: %w", schemaName, err)
	}

	// Create initial audit log partitions
	if err := createAuditPartitions(ctx, tx, schemaName); err != nil {
		return fmt.Errorf("create audit partitions in %s: %w", schemaName, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit schema creation: %w", err)
	}

	return nil
}

// DropSiteSchema drops a site schema and all its tables.
func DropSiteSchema(ctx context.Context, db *bun.DB, slug string) error {
	if !ValidateSlug(slug) {
		return fmt.Errorf("invalid site slug: %q", slug)
	}

	schemaName := "site_" + slug
	_, err := db.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", bun.Ident(schemaName)))
	if err != nil {
		return fmt.Errorf("drop schema %s: %w", schemaName, err)
	}
	return nil
}

// createAuditPartitions creates monthly sfc_site_audits partitions for the
// current month and the next 2 months.
func createAuditPartitions(ctx context.Context, tx bun.Tx, schemaName string) error {
	now := time.Now()
	for i := 0; i < 3; i++ {
		t := now.AddDate(0, i, 0)
		start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)

		partName := fmt.Sprintf("%s.sfc_site_audits_%s", schemaName, start.Format("2006_01"))
		sql := fmt.Sprintf(
			"CREATE TABLE IF NOT EXISTS %s PARTITION OF %s.sfc_site_audits FOR VALUES FROM ('%s') TO ('%s')",
			partName,
			schemaName,
			start.Format("2006-01-02"),
			end.Format("2006-01-02"),
		)

		if _, err := tx.ExecContext(ctx, sql); err != nil {
			return fmt.Errorf("create partition %s: %w", partName, err)
		}
	}

	// Create audit indexes (these are on the parent, PG propagates to partitions)
	indexes := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_sfc_site_audits_actor ON %s.sfc_site_audits(actor_id, created_at DESC)", schemaName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_sfc_site_audits_resource ON %s.sfc_site_audits(resource_type, resource_id)", schemaName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_sfc_site_audits_time ON %s.sfc_site_audits(created_at DESC)", schemaName),
	}
	for _, idx := range indexes {
		if _, err := tx.ExecContext(ctx, idx); err != nil {
			return fmt.Errorf("create audit index: %w", err)
		}
	}

	return nil
}

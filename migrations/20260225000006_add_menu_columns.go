package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Get all site schemas
		var schemas []string
		err := db.NewSelect().
			TableExpr("information_schema.schemata").
			ColumnExpr("schema_name").
			Where("schema_name LIKE 'site_%'").
			Scan(ctx, &schemas)
		if err != nil {
			return fmt.Errorf("list site schemas: %w", err)
		}

		for _, schema := range schemas {
			if _, err := db.ExecContext(ctx, fmt.Sprintf(`
				ALTER TABLE %q.sfc_site_menus ADD COLUMN IF NOT EXISTS description TEXT;
				ALTER TABLE %q.sfc_site_menu_items ADD COLUMN IF NOT EXISTS icon VARCHAR(50);
				ALTER TABLE %q.sfc_site_menu_items ADD COLUMN IF NOT EXISTS css_class VARCHAR(100);
			`, schema, schema, schema)); err != nil {
				return fmt.Errorf("migrate %s menus: %w", schema, err)
			}
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		var schemas []string
		err := db.NewSelect().
			TableExpr("information_schema.schemata").
			ColumnExpr("schema_name").
			Where("schema_name LIKE 'site_%'").
			Scan(ctx, &schemas)
		if err != nil {
			return fmt.Errorf("list site schemas: %w", err)
		}

		for _, schema := range schemas {
			if _, err := db.ExecContext(ctx, fmt.Sprintf(`
				ALTER TABLE %q.sfc_site_menus DROP COLUMN IF EXISTS description;
				ALTER TABLE %q.sfc_site_menu_items DROP COLUMN IF EXISTS icon;
				ALTER TABLE %q.sfc_site_menu_items DROP COLUMN IF EXISTS css_class;
			`, schema, schema, schema)); err != nil {
				return fmt.Errorf("rollback %s menus: %w", schema, err)
			}
		}
		return nil
	})
}

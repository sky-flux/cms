package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// ── public schema ──────────────────────────────────────

		// sfc_users: is_active BOOLEAN → status SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_users
				ADD COLUMN status SMALLINT NOT NULL DEFAULT 1;
			UPDATE public.sfc_users SET status = CASE WHEN is_active THEN 1 ELSE 2 END;
			ALTER TABLE public.sfc_users DROP COLUMN is_active;
		`); err != nil {
			return fmt.Errorf("migrate sfc_users: %w", err)
		}

		// sfc_sites: is_active BOOLEAN → status SMALLINT
		if _, err := db.ExecContext(ctx, `
			DROP INDEX IF EXISTS public.idx_sfc_sites_active;
			ALTER TABLE public.sfc_sites
				ADD COLUMN status SMALLINT NOT NULL DEFAULT 1;
			UPDATE public.sfc_sites SET status = CASE WHEN is_active THEN 1 ELSE 2 END;
			ALTER TABLE public.sfc_sites DROP COLUMN is_active;
			CREATE INDEX idx_sfc_sites_status ON public.sfc_sites(status) WHERE status = 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_sites: %w", err)
		}

		// sfc_roles: built_in BOOLEAN → SMALLINT, status BOOLEAN → SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_roles
				ALTER COLUMN built_in TYPE SMALLINT USING CASE WHEN built_in THEN 2 ELSE 1 END,
				ALTER COLUMN built_in SET DEFAULT 1,
				ALTER COLUMN status TYPE SMALLINT USING CASE WHEN status THEN 1 ELSE 2 END,
				ALTER COLUMN status SET DEFAULT 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_roles: %w", err)
		}

		// sfc_apis: status BOOLEAN → SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_apis
				ALTER COLUMN status TYPE SMALLINT USING CASE WHEN status THEN 1 ELSE 2 END,
				ALTER COLUMN status SET DEFAULT 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_apis: %w", err)
		}

		// sfc_menus (admin): status BOOLEAN → SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_menus
				ALTER COLUMN status TYPE SMALLINT USING CASE WHEN status THEN 1 ELSE 2 END,
				ALTER COLUMN status SET DEFAULT 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_menus: %w", err)
		}

		// sfc_role_templates: built_in BOOLEAN → SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_role_templates
				ALTER COLUMN built_in TYPE SMALLINT USING CASE WHEN built_in THEN 2 ELSE 1 END,
				ALTER COLUMN built_in SET DEFAULT 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_role_templates: %w", err)
		}

		// sfc_refresh_tokens: revoked BOOLEAN → SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_refresh_tokens
				ALTER COLUMN revoked TYPE SMALLINT USING CASE WHEN revoked THEN 2 ELSE 1 END,
				ALTER COLUMN revoked SET DEFAULT 1;
		`); err != nil {
			return fmt.Errorf("migrate sfc_refresh_tokens: %w", err)
		}

		// sfc_user_totp: is_enabled BOOLEAN → enabled SMALLINT
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE public.sfc_user_totp
				ADD COLUMN enabled SMALLINT NOT NULL DEFAULT 1;
			UPDATE public.sfc_user_totp SET enabled = CASE WHEN is_enabled THEN 2 ELSE 1 END;
			ALTER TABLE public.sfc_user_totp DROP COLUMN is_enabled;
		`); err != nil {
			return fmt.Errorf("migrate sfc_user_totp: %w", err)
		}

		// ── site schemas ──────────────────────────────────────
		// Migrate all existing site schemas (for already-created sites)
		rows, err := db.QueryContext(ctx, `SELECT slug FROM public.sfc_sites`)
		if err != nil {
			return fmt.Errorf("list sites: %w", err)
		}
		defer rows.Close()

		var slugs []string
		for rows.Next() {
			var slug string
			if err := rows.Scan(&slug); err != nil {
				return fmt.Errorf("scan slug: %w", err)
			}
			slugs = append(slugs, slug)
		}

		for _, slug := range slugs {
			schema := "site_" + slug
			if _, err := db.ExecContext(ctx, fmt.Sprintf(`
				-- post_types: built_in BOOLEAN → SMALLINT
				ALTER TABLE %[1]s.sfc_site_post_types
					ALTER COLUMN built_in TYPE SMALLINT USING CASE WHEN built_in THEN 2 ELSE 1 END,
					ALTER COLUMN built_in SET DEFAULT 1;

				-- post_category_map: is_primary BOOLEAN → primary SMALLINT
				ALTER TABLE %[1]s.sfc_site_post_category_map
					ADD COLUMN "primary" SMALLINT NOT NULL DEFAULT 1;
				UPDATE %[1]s.sfc_site_post_category_map SET "primary" = CASE WHEN is_primary THEN 2 ELSE 1 END;
				ALTER TABLE %[1]s.sfc_site_post_category_map DROP COLUMN is_primary;

				-- comments: is_pinned BOOLEAN → pinned SMALLINT
				ALTER TABLE %[1]s.sfc_site_comments
					ADD COLUMN pinned SMALLINT NOT NULL DEFAULT 1;
				UPDATE %[1]s.sfc_site_comments SET pinned = CASE WHEN is_pinned THEN 2 ELSE 1 END;
				ALTER TABLE %[1]s.sfc_site_comments DROP COLUMN is_pinned;

				-- menu_items: is_active BOOLEAN → status SMALLINT
				ALTER TABLE %[1]s.sfc_site_menu_items
					ADD COLUMN status SMALLINT NOT NULL DEFAULT 1;
				UPDATE %[1]s.sfc_site_menu_items SET status = CASE WHEN is_active THEN 1 ELSE 2 END;
				ALTER TABLE %[1]s.sfc_site_menu_items DROP COLUMN is_active;

				-- redirects: is_active BOOLEAN → status SMALLINT
				DROP INDEX IF EXISTS %[1]s.idx_sfc_site_redirects_source;
				ALTER TABLE %[1]s.sfc_site_redirects
					ADD COLUMN status SMALLINT NOT NULL DEFAULT 1;
				UPDATE %[1]s.sfc_site_redirects SET status = CASE WHEN is_active THEN 1 ELSE 2 END;
				ALTER TABLE %[1]s.sfc_site_redirects DROP COLUMN is_active;
				CREATE INDEX idx_sfc_site_redirects_source ON %[1]s.sfc_site_redirects(source_path)
					WHERE status = 1;

				-- api_keys: is_active BOOLEAN → status SMALLINT
				ALTER TABLE %[1]s.sfc_site_api_keys
					ADD COLUMN status SMALLINT NOT NULL DEFAULT 1;
				UPDATE %[1]s.sfc_site_api_keys SET status = CASE WHEN is_active THEN 1 ELSE 2 END;
				ALTER TABLE %[1]s.sfc_site_api_keys DROP COLUMN is_active;
			`, schema)); err != nil {
				return fmt.Errorf("migrate site schema %s: %w", schema, err)
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		return fmt.Errorf("down migration not supported for boolean-to-smallint conversion")
	})
}

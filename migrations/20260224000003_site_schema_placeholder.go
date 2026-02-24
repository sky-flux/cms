package migrations

import (
	"context"
	"log/slog"

	"github.com/uptrace/bun"
)

// Site schemas (site_{slug}) are NOT created by standard migrations.
// They are created dynamically when a new site is registered, via the
// internal/schema package (schema.CreateSiteSchema).
//
// This migration is a no-op placeholder to maintain sequential ordering.

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		slog.Info("site schemas are created dynamically via internal/schema package — skipping")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		slog.Info("site schemas are managed via internal/schema package — skipping")
		return nil
	})
}

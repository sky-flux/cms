package migrations

import "github.com/uptrace/bun/migrate"

// Migrations is the global migration registry.
// Each migration file registers itself via init() + MustRegister().
var Migrations = migrate.NewMigrations()

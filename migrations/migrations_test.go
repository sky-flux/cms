package migrations

import (
	"testing"
)

// TestMigrationsCount verifies we have the expected number of registered
// migrations. The site_schema_placeholder (old migration 3) must be deleted.
// After Phase 0 additions: 1 (core) + 2 (rbac) + 3 (seed) + 4 (bool→smallint)
// + 5 (menu columns) + 6 (content tables) + 7 (content indexes) = 7 migrations.
func TestMigrationsCount(t *testing.T) {
	got := len(Migrations.Sorted())
	want := 7
	if got != want {
		t.Errorf("expected %d migrations, got %d — did you forget to delete the placeholder or add new files?", want, got)
	}
}

package infra_test

import (
	"testing"

	"github.com/sky-flux/cms/internal/media/domain"
	"github.com/sky-flux/cms/internal/media/infra"
)

// Compile-time interface check.
var _ domain.MediaFileRepository = (*infra.BunMediaRepo)(nil)

func TestBunMediaRepo_SaveAndFindByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	// TODO: wire testcontainers-go PostgreSQL 18 + run migrations
	// For now, just check struct satisfies interface via compile check above.
	t.Log("integration test placeholder — run without -short when Docker is available")
}

func TestBunMediaRepo_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	t.Log("integration test placeholder")
}

func TestBunMediaRepo_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	t.Log("integration test placeholder")
}

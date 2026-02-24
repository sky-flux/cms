package rbac_test

import "testing"

func TestRoleRepo_CRUD(t *testing.T) {
	t.Skip("requires testcontainers-go PostgreSQL")
}

func TestRoleRepo_DeleteBuiltIn(t *testing.T) {
	t.Skip("requires testcontainers-go PostgreSQL")
}

func TestRoleRepo_DuplicateSlug(t *testing.T) {
	t.Skip("requires testcontainers-go PostgreSQL")
}

package infra_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/sky-flux/cms/internal/identity/domain"
	"github.com/sky-flux/cms/internal/identity/infra"
)

func setupTestDB(t *testing.T) *bun.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:18-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "cms_test",
			"POSTGRES_USER":     "cms",
			"POSTGRES_PASSWORD": "secret",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	connStr := "postgres://cms:secret@" + host + ":" + port.Port() + "/cms_test?sslmode=disable"

	sqldb, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(false)))

	// Create the sfc_users table for tests.
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS sfc_users (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email         TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			display_name  TEXT NOT NULL,
			avatar_url    TEXT NOT NULL DEFAULT '',
			status        SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 2),
			last_login_at TIMESTAMPTZ,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at    TIMESTAMPTZ
		)
	`)
	require.NoError(t, err)

	return db
}

func TestBunUserRepo_SaveAndFindByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := infra.NewBunUserRepo(db)
	ctx := context.Background()

	u, err := domain.NewUser("bob@example.com", "Bob", "$2a$12$hash")
	require.NoError(t, err)

	err = repo.Save(ctx, u)
	require.NoError(t, err)
	assert.NotEmpty(t, u.ID) // DB-assigned UUID

	found, err := repo.FindByEmail(ctx, "bob@example.com")
	require.NoError(t, err)
	assert.Equal(t, "bob@example.com", found.Email)
	assert.Equal(t, "Bob", found.DisplayName)
	assert.Equal(t, domain.UserStatusActive, found.Status)
}

func TestBunUserRepo_FindByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := infra.NewBunUserRepo(db)

	_, err := repo.FindByEmail(context.Background(), "nobody@example.com")
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestBunUserRepo_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := infra.NewBunUserRepo(db)
	ctx := context.Background()

	u, _ := domain.NewUser("carol@example.com", "Carol", "hash")
	require.NoError(t, repo.Save(ctx, u))

	found, err := repo.FindByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, u.ID, found.ID)
}

func TestBunUserRepo_UpdatePassword(t *testing.T) {
	db := setupTestDB(t)
	repo := infra.NewBunUserRepo(db)
	ctx := context.Background()

	u, _ := domain.NewUser("dave@example.com", "Dave", "old-hash")
	require.NoError(t, repo.Save(ctx, u))

	err := repo.UpdatePassword(ctx, u.ID, "new-hash")
	require.NoError(t, err)

	found, _ := repo.FindByID(ctx, u.ID)
	assert.Equal(t, "new-hash", found.PasswordHash)
}

func TestBunUserRepo_UpdateLastLogin(t *testing.T) {
	db := setupTestDB(t)
	repo := infra.NewBunUserRepo(db)
	ctx := context.Background()

	u, _ := domain.NewUser("eve@example.com", "Eve", "hash")
	require.NoError(t, repo.Save(ctx, u))

	err := repo.UpdateLastLogin(ctx, u.ID)
	require.NoError(t, err)

	found, _ := repo.FindByID(ctx, u.ID)
	assert.NotNil(t, found.LastLoginAt)
}

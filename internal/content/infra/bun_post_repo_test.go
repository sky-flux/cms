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

	"github.com/sky-flux/cms/internal/content/domain"
	"github.com/sky-flux/cms/internal/content/infra"
)

func setupPostDB(t *testing.T) *bun.DB {
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
	pgC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgC.Terminate(ctx) })

	host, err := pgC.Host(ctx)
	require.NoError(t, err)
	port, err := pgC.MappedPort(ctx, "5432")
	require.NoError(t, err)

	connStr := "postgres://cms:secret@" + host + ":" + port.Port() + "/cms_test?sslmode=disable"
	sqldb, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	db := bun.NewDB(sqldb, pgdialect.New())

	// Minimal schema for posts.
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS sfc_posts (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			author_id   UUID NOT NULL,
			title       TEXT NOT NULL,
			slug        TEXT NOT NULL UNIQUE,
			excerpt     TEXT NOT NULL DEFAULT '',
			content     TEXT NOT NULL DEFAULT '',
			status      SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4),
			version     INT NOT NULL DEFAULT 1,
			published_at TIMESTAMPTZ,
			scheduled_at TIMESTAMPTZ,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at  TIMESTAMPTZ
		)
	`)
	require.NoError(t, err)
	return db
}

func TestBunPostRepo_SaveAndFindByID(t *testing.T) {
	db := setupPostDB(t)
	repo := infra.NewBunPostRepo(db)
	ctx := context.Background()

	post, _ := domain.NewPost("First Post", "first-post", "author-uuid")
	require.NoError(t, repo.Save(ctx, post))
	assert.NotEmpty(t, post.ID)

	found, err := repo.FindByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, "first-post", found.Slug)
	assert.Equal(t, domain.PostStatusDraft, found.Status)
}

func TestBunPostRepo_FindBySlug_NotFound(t *testing.T) {
	db := setupPostDB(t)
	repo := infra.NewBunPostRepo(db)

	_, err := repo.FindBySlug(context.Background(), "no-such-slug")
	assert.ErrorIs(t, err, domain.ErrPostNotFound)
}

func TestBunPostRepo_SlugExists(t *testing.T) {
	db := setupPostDB(t)
	repo := infra.NewBunPostRepo(db)
	ctx := context.Background()

	p, _ := domain.NewPost("Slug Test", "existing-slug", "author-uuid")
	require.NoError(t, repo.Save(ctx, p))

	exists, err := repo.SlugExists(ctx, "existing-slug", "")
	require.NoError(t, err)
	assert.True(t, exists)

	// Exclude the post itself (update scenario).
	exists, err = repo.SlugExists(ctx, "existing-slug", p.ID)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestBunPostRepo_Update_OptimisticLock(t *testing.T) {
	db := setupPostDB(t)
	repo := infra.NewBunPostRepo(db)
	ctx := context.Background()

	p, _ := domain.NewPost("Lock Test", "lock-slug", "author-uuid")
	require.NoError(t, repo.Save(ctx, p))

	// Correct version — should succeed.
	p.Title = "Updated Title"
	err := repo.Update(ctx, p, 1)
	require.NoError(t, err)

	// Stale version — should fail.
	p.Title = "Stale Update"
	err = repo.Update(ctx, p, 1) // version is now 2
	assert.ErrorIs(t, err, domain.ErrVersionConflict)
}

func TestBunPostRepo_SoftDelete(t *testing.T) {
	db := setupPostDB(t)
	repo := infra.NewBunPostRepo(db)
	ctx := context.Background()

	p, _ := domain.NewPost("Delete Me", "delete-me", "author-uuid")
	require.NoError(t, repo.Save(ctx, p))

	err := repo.SoftDelete(ctx, p.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, p.ID)
	assert.ErrorIs(t, err, domain.ErrPostNotFound)
}

func TestBunPostRepo_List_Pagination(t *testing.T) {
	db := setupPostDB(t)
	repo := infra.NewBunPostRepo(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		p, _ := domain.NewPost(
			"Post "+string(rune('A'+i)),
			"post-"+string(rune('a'+i)),
			"author-uuid",
		)
		require.NoError(t, repo.Save(ctx, p))
	}

	posts, total, err := repo.List(ctx, domain.PostFilter{Page: 1, PerPage: 3})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, posts, 3)
}

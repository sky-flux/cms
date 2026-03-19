package app_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/content/app"
	"github.com/sky-flux/cms/internal/content/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock repo for app tests ---
type mockPostRepo struct {
	saveFn       func(ctx context.Context, p *domain.Post) error
	slugExistsFn func(ctx context.Context, slug, excludeID string) (bool, error)
	findByIDFn   func(ctx context.Context, id string) (*domain.Post, error)
	updateFn     func(ctx context.Context, p *domain.Post, expectedVersion int) error
}

func (m *mockPostRepo) Save(ctx context.Context, p *domain.Post) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, p)
	}
	return nil
}
func (m *mockPostRepo) FindByID(ctx context.Context, id string) (*domain.Post, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrPostNotFound
}
func (m *mockPostRepo) FindBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	return nil, domain.ErrPostNotFound
}
func (m *mockPostRepo) Update(ctx context.Context, p *domain.Post, ev int) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, p, ev)
	}
	return nil
}
func (m *mockPostRepo) SoftDelete(ctx context.Context, id string) error { return nil }
func (m *mockPostRepo) List(ctx context.Context, f domain.PostFilter) ([]*domain.Post, int64, error) {
	return nil, 0, nil
}
func (m *mockPostRepo) SlugExists(ctx context.Context, slug, excludeID string) (bool, error) {
	if m.slugExistsFn != nil {
		return m.slugExistsFn(ctx, slug, excludeID)
	}
	return false, nil
}

// --- CreatePost tests ---

func TestCreatePostUseCase_Success(t *testing.T) {
	uc := app.NewCreatePostUseCase(&mockPostRepo{})

	out, err := uc.Execute(context.Background(), app.CreatePostInput{
		Title:    "Hello World",
		Slug:     "hello-world",
		AuthorID: "author-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "hello-world", out.Slug)
	assert.Equal(t, domain.PostStatusDraft, out.Status)
}

func TestCreatePostUseCase_SlugConflict(t *testing.T) {
	uc := app.NewCreatePostUseCase(&mockPostRepo{
		slugExistsFn: func(_ context.Context, _, _ string) (bool, error) {
			return true, nil
		},
	})

	_, err := uc.Execute(context.Background(), app.CreatePostInput{
		Title:    "Hello",
		Slug:     "hello-world",
		AuthorID: "author-1",
	})
	assert.ErrorIs(t, err, domain.ErrSlugConflict)
}

func TestCreatePostUseCase_EmptyTitle(t *testing.T) {
	uc := app.NewCreatePostUseCase(&mockPostRepo{})

	_, err := uc.Execute(context.Background(), app.CreatePostInput{
		Title:    "",
		Slug:     "slug",
		AuthorID: "author-1",
	})
	assert.ErrorIs(t, err, domain.ErrEmptyTitle)
}

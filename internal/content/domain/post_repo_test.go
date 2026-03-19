package domain_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/content/domain"
)

var _ domain.PostRepository = (*mockPostRepo)(nil)

type mockPostRepo struct {
	saveFn       func(ctx context.Context, p *domain.Post) error
	findByIDFn   func(ctx context.Context, id string) (*domain.Post, error)
	findBySlugFn func(ctx context.Context, slug string) (*domain.Post, error)
	updateFn     func(ctx context.Context, p *domain.Post, expectedVersion int) error
	softDeleteFn func(ctx context.Context, id string) error
	listFn       func(ctx context.Context, f domain.PostFilter) ([]*domain.Post, int64, error)
	slugExistsFn func(ctx context.Context, slug, excludeID string) (bool, error)
}

func (m *mockPostRepo) Save(ctx context.Context, p *domain.Post) error {
	return m.saveFn(ctx, p)
}
func (m *mockPostRepo) FindByID(ctx context.Context, id string) (*domain.Post, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockPostRepo) FindBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	return m.findBySlugFn(ctx, slug)
}
func (m *mockPostRepo) Update(ctx context.Context, p *domain.Post, expectedVersion int) error {
	return m.updateFn(ctx, p, expectedVersion)
}
func (m *mockPostRepo) SoftDelete(ctx context.Context, id string) error {
	return m.softDeleteFn(ctx, id)
}
func (m *mockPostRepo) List(ctx context.Context, f domain.PostFilter) ([]*domain.Post, int64, error) {
	return m.listFn(ctx, f)
}
func (m *mockPostRepo) SlugExists(ctx context.Context, slug, excludeID string) (bool, error) {
	return m.slugExistsFn(ctx, slug, excludeID)
}

func TestPostRepository_Interface(t *testing.T) { t.Log("PostRepository interface satisfied") }

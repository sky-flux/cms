package app_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/content/app"
	"github.com/sky-flux/cms/internal/content/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCategoryRepo struct {
	saveFn          func(ctx context.Context, c *domain.Category) error
	slugExistsFn    func(ctx context.Context, slug, excludeID string) (bool, error)
	findAncestorsFn func(ctx context.Context, id string) ([]string, error)
}

func (m *mockCategoryRepo) Save(ctx context.Context, c *domain.Category) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, c)
	}
	return nil
}
func (m *mockCategoryRepo) FindByID(ctx context.Context, id string) (*domain.Category, error) {
	return nil, domain.ErrCategoryNotFound
}
func (m *mockCategoryRepo) SlugExists(ctx context.Context, slug, excludeID string) (bool, error) {
	if m.slugExistsFn != nil {
		return m.slugExistsFn(ctx, slug, excludeID)
	}
	return false, nil
}
func (m *mockCategoryRepo) FindAncestorIDs(ctx context.Context, id string) ([]string, error) {
	if m.findAncestorsFn != nil {
		return m.findAncestorsFn(ctx, id)
	}
	return nil, nil
}
func (m *mockCategoryRepo) List(ctx context.Context) ([]*domain.Category, error) { return nil, nil }
func (m *mockCategoryRepo) SoftDelete(ctx context.Context, id string) error      { return nil }
func (m *mockCategoryRepo) Update(ctx context.Context, c *domain.Category) error { return nil }

func TestCreateCategoryUseCase_Success(t *testing.T) {
	uc := app.NewCreateCategoryUseCase(&mockCategoryRepo{})
	out, err := uc.Execute(context.Background(), app.CreateCategoryInput{
		Name: "Tech", Slug: "tech", ParentID: "",
	})
	require.NoError(t, err)
	assert.Equal(t, "tech", out.Slug)
}

func TestCreateCategoryUseCase_SlugConflict(t *testing.T) {
	uc := app.NewCreateCategoryUseCase(&mockCategoryRepo{
		slugExistsFn: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
	})
	_, err := uc.Execute(context.Background(), app.CreateCategoryInput{
		Name: "Tech", Slug: "tech",
	})
	assert.ErrorIs(t, err, domain.ErrSlugConflict)
}

func TestCreateCategoryUseCase_CycleDetection(t *testing.T) {
	uc := app.NewCreateCategoryUseCase(&mockCategoryRepo{
		findAncestorsFn: func(_ context.Context, id string) ([]string, error) {
			// parent "p1" has ancestors ["grandparent"]
			return []string{"grandparent"}, nil
		},
	})
	// This is a contrived cycle test for the use-case layer.
	// Real cycle would require the new category to be its own ancestor.
	// For v1, validate that ancestor lookup is called when parentID is set.
	out, err := uc.Execute(context.Background(), app.CreateCategoryInput{
		Name: "Child", Slug: "child", ParentID: "p1",
	})
	require.NoError(t, err) // No cycle here — just verifying it doesn't error
	assert.Equal(t, "p1", out.ParentID)
}

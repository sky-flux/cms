package public_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/delivery/public"
)

var _ public.PublicPostQuery = (*mockPublicPostQuery)(nil)
var _ public.PublicCategoryQuery = (*mockPublicCategoryQuery)(nil)
var _ public.PublicTagQuery = (*mockPublicTagQuery)(nil)
var _ public.PublicSearchQuery = (*mockPublicSearchQuery)(nil)

type mockPublicPostQuery struct{}

func (m *mockPublicPostQuery) ListPublished(ctx context.Context, f public.PostFilter) ([]public.PublicPost, int64, error) {
	return nil, 0, nil
}
func (m *mockPublicPostQuery) GetBySlug(ctx context.Context, slug string) (*public.PublicPost, error) {
	return nil, nil
}
func (m *mockPublicPostQuery) IncrementViewCount(ctx context.Context, id string) error { return nil }

type mockPublicCategoryQuery struct{}

func (m *mockPublicCategoryQuery) ListWithPostCounts(ctx context.Context) ([]public.PublicCategory, error) {
	return nil, nil
}

type mockPublicTagQuery struct{}

func (m *mockPublicTagQuery) ListWithPostCounts(ctx context.Context, sort string) ([]public.PublicTag, error) {
	return nil, nil
}

type mockPublicSearchQuery struct{}

func (m *mockPublicSearchQuery) Search(ctx context.Context, q string, page, perPage int) ([]public.SearchResult, int64, error) {
	return nil, 0, nil
}

func TestPublicQueryInterfaces_Compile(t *testing.T) {
	t.Log("all public query interfaces satisfied")
}

package feed_test

import (
	"context"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/delivery/feed"
)

// Compile-time interface checks.
var _ feed.FeedPostQuery = (*mockFeedPostQuery)(nil)
var _ feed.FeedSiteQuery = (*mockFeedSiteQuery)(nil)

type mockFeedPostQuery struct {
	listPublishedFn     func(ctx context.Context, limit int) ([]feed.FeedPost, error)
	latestPublishedAtFn func(ctx context.Context) (*time.Time, error)
}

func (m *mockFeedPostQuery) ListPublished(ctx context.Context, limit int) ([]feed.FeedPost, error) {
	return m.listPublishedFn(ctx, limit)
}
func (m *mockFeedPostQuery) LatestPublishedAt(ctx context.Context) (*time.Time, error) {
	return m.latestPublishedAtFn(ctx)
}

type mockFeedSiteQuery struct {
	getTitleFn       func(ctx context.Context) string
	getURLFn         func(ctx context.Context) string
	getDescriptionFn func(ctx context.Context) string
	getLanguageFn    func(ctx context.Context) string
}

func (m *mockFeedSiteQuery) GetSiteTitle(ctx context.Context) string {
	return m.getTitleFn(ctx)
}
func (m *mockFeedSiteQuery) GetSiteURL(ctx context.Context) string { return m.getURLFn(ctx) }
func (m *mockFeedSiteQuery) GetSiteDescription(ctx context.Context) string {
	return m.getDescriptionFn(ctx)
}
func (m *mockFeedSiteQuery) GetSiteLanguage(ctx context.Context) string {
	return m.getLanguageFn(ctx)
}

func TestFeedQueryInterfaces_Compile(t *testing.T) {
	t.Log("FeedPostQuery and FeedSiteQuery interfaces satisfied")
}

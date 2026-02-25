package feed

import (
	"context"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// FeedPostReader queries published posts for feed generation.
type FeedPostReader interface {
	ListPublished(ctx context.Context, limit int, categorySlug, tagSlug string) ([]model.Post, error)
	LatestPublishedAt(ctx context.Context) (*time.Time, error)
}

// FeedCategoryReader queries categories for sitemap generation.
type FeedCategoryReader interface {
	ListAll(ctx context.Context) ([]model.Category, error)
	LatestPostDate(ctx context.Context, categoryID string) (*time.Time, error)
}

// FeedTagReader queries tags for sitemap generation.
type FeedTagReader interface {
	ListWithPosts(ctx context.Context) ([]TagWithLastmod, error)
}

// SiteConfigReader reads site configuration for feed metadata.
type SiteConfigReader interface {
	GetSiteTitle(ctx context.Context) string
	GetSiteURL(ctx context.Context) string
	GetSiteDescription(ctx context.Context) string
	GetSiteLanguage(ctx context.Context) string
}

// TagWithLastmod is a tag with its latest post date and count.
type TagWithLastmod struct {
	model.Tag
	LastPostDate *time.Time `bun:"last_post_date"`
	PostCount    int64      `bun:"post_count"`
}

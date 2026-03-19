package feed

import (
	"context"
	"time"
)

// FeedPost is a minimal projection of a post for feed generation.
// Avoids a direct dependency on the content BC's domain types.
type FeedPost struct {
	ID          string
	Title       string
	Slug        string
	Excerpt     string
	Content     string
	AuthorName  string
	PublishedAt time.Time
	UpdatedAt   time.Time
	Categories  []string // category names
}

// FeedCategoryItem is a minimal projection of a category for sitemap generation.
type FeedCategoryItem struct {
	Slug       string
	LastPostAt *time.Time
}

// FeedTagItem is a minimal projection of a tag for sitemap generation.
type FeedTagItem struct {
	Slug       string
	LastPostAt *time.Time
}

// FeedPostQuery is the read port the feed handler uses to fetch posts.
// Implemented by the content BC's infra layer.
type FeedPostQuery interface {
	// ListPublished returns the most recent published posts up to limit.
	ListPublished(ctx context.Context, limit int) ([]FeedPost, error)
	// LatestPublishedAt returns the time of the most recently published post.
	LatestPublishedAt(ctx context.Context) (*time.Time, error)
}

// FeedCategoryQuery is the read port for category sitemap data.
type FeedCategoryQuery interface {
	// ListAll returns all categories that have at least one published post.
	ListAll(ctx context.Context) ([]FeedCategoryItem, error)
}

// FeedTagQuery is the read port for tag sitemap data.
type FeedTagQuery interface {
	// ListWithPosts returns tags that have at least one published post.
	ListWithPosts(ctx context.Context) ([]FeedTagItem, error)
}

// FeedSiteQuery is the read port for site metadata used in feed headers.
// Implemented by the platform BC's infra layer (reads sfc_configs).
type FeedSiteQuery interface {
	GetSiteTitle(ctx context.Context) string
	GetSiteURL(ctx context.Context) string
	GetSiteDescription(ctx context.Context) string
	GetSiteLanguage(ctx context.Context) string
}

// Package public provides the public-facing REST API delivery layer.
package public

import (
	"context"
	"time"
)

// PostFilter controls pagination and filtering for the public post list.
type PostFilter struct {
	Page     int
	PerPage  int
	Category string // category slug
	Tag      string // tag slug
	Sort     string // e.g. "published_at:desc"
}

// PublicPost is the read model projected for the public API.
type PublicPost struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Excerpt     string    `json:"excerpt"`
	Content     string    `json:"content,omitempty"`
	AuthorName  string    `json:"author_name"`
	CoverURL    string    `json:"cover_url,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ViewCount   int64     `json:"view_count"`
	Tags        []string  `json:"tags,omitempty"`
	Categories  []string  `json:"categories,omitempty"`
}

// PublicCategory is a flat category node with post count.
type PublicCategory struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	ParentID  string `json:"parent_id,omitempty"`
	PostCount int64  `json:"post_count"`
}

// PublicTag is a tag with post count.
type PublicTag struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	PostCount int64  `json:"post_count"`
}

// SearchResult is a single item returned by Meilisearch.
type SearchResult struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Slug    string `json:"slug"`
	Excerpt string `json:"excerpt"`
	Type    string `json:"type"` // "post"
}

// PublicPostQuery is the read port for published posts.
type PublicPostQuery interface {
	// ListPublished returns paginated published posts matching the filter.
	ListPublished(ctx context.Context, f PostFilter) ([]PublicPost, int64, error)
	// GetBySlug returns a single published post by URL slug, or nil + ErrNotFound.
	GetBySlug(ctx context.Context, slug string) (*PublicPost, error)
	// IncrementViewCount records a page view (fire-and-forget; errors are ignored).
	IncrementViewCount(ctx context.Context, id string) error
}

// PublicCategoryQuery is the read port for public category listing.
type PublicCategoryQuery interface {
	// ListWithPostCounts returns all categories with their published post counts.
	ListWithPostCounts(ctx context.Context) ([]PublicCategory, error)
}

// PublicTagQuery is the read port for public tag listing.
type PublicTagQuery interface {
	// ListWithPostCounts returns all tags with their published post counts.
	// sort is a field:direction string, e.g. "name:asc" or "post_count:desc".
	ListWithPostCounts(ctx context.Context, sort string) ([]PublicTag, error)
}

// PublicSearchQuery is the read port for full-text search via Meilisearch.
type PublicSearchQuery interface {
	// Search queries the Meilisearch index and returns matching posts.
	Search(ctx context.Context, q string, page, perPage int) ([]SearchResult, int64, error)
}

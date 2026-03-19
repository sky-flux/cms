package web

import (
	"context"

	"github.com/sky-flux/cms/web/templates"
)

// PostQuery fetches posts for web rendering. Implemented by infra layer.
type PostQuery interface {
	// ListLatest returns the most recent published posts (page starts at 1).
	ListLatest(ctx context.Context, page, pageSize int) ([]templates.PostSummary, error)
	// GetBySlug returns a single published post by slug.
	GetBySlug(ctx context.Context, slug string) (*templates.PostDetail, error)
	// ListByCategory returns published posts for a category (page starts at 1).
	ListByCategory(ctx context.Context, categorySlug string, page, pageSize int) ([]templates.PostSummary, error)
	// ListByTag returns published posts for a tag (page starts at 1).
	ListByTag(ctx context.Context, tagSlug string, page, pageSize int) ([]templates.PostSummary, error)
	// Search returns posts matching query string via Meilisearch.
	Search(ctx context.Context, query string, limit int) ([]templates.PostSummary, error)
}

// CategoryQuery fetches category metadata for archive pages.
type CategoryQuery interface {
	GetBySlug(ctx context.Context, slug string) (name string, err error)
}

// TagQuery fetches tag metadata for archive pages.
type TagQuery interface {
	GetBySlug(ctx context.Context, slug string) (name string, err error)
}

// CommentWriter submits a new comment for moderation.
type CommentWriter interface {
	Submit(ctx context.Context, postSlug, authorName, authorEmail, body string) error
}

// SiteConfigLoader loads site-wide settings (name, description, nav items).
type SiteConfigLoader interface {
	Load(ctx context.Context) (templates.SiteConfig, error)
}

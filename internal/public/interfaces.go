package public

import (
	"context"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// PostReader reads published posts.
type PostReader interface {
	List(ctx context.Context, f PostListFilter) ([]model.Post, int64, error)
	GetBySlug(ctx context.Context, slug string) (*model.Post, error)
	GetByID(ctx context.Context, id string) (*model.Post, error)
	LoadRelations(ctx context.Context, post *model.Post) error
	IncrementViewCount(ctx context.Context, id string) error
}

// CategoryReader reads categories with post counts.
type CategoryReader interface {
	List(ctx context.Context) ([]model.Category, error)
	CountPosts(ctx context.Context, categoryID string) (int64, error)
}

// TagReader reads tags with post counts.
type TagReader interface {
	ListPublic(ctx context.Context, sort string) ([]TagWithCount, error)
}

// CommentReader reads and creates comments.
type CommentReader interface {
	ListByPost(ctx context.Context, postID string, page, perPage int) ([]model.Comment, int64, error)
	Create(ctx context.Context, c *model.Comment) error
	GetByID(ctx context.Context, id string) (*model.Comment, error)
	GetParentChainDepth(ctx context.Context, commentID string) (int, error)
}

// MenuReader reads site menus with items.
type MenuReader interface {
	GetByLocation(ctx context.Context, location string) (*model.SiteMenu, error)
	GetBySlug(ctx context.Context, slug string) (*model.SiteMenu, error)
	ListItemsByMenuID(ctx context.Context, menuID string) ([]*model.SiteMenuItem, error)
}

// PreviewReader reads preview tokens.
type PreviewReader interface {
	GetByHash(ctx context.Context, hash string) (*model.PreviewToken, error)
}

// Searcher wraps Meilisearch search operations.
type Searcher interface {
	Search(ctx context.Context, uid, query string, opts interface{}) (interface{}, error)
}

// Cacher wraps Redis cache operations.
type Cacher interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
}

// PostListFilter for public post listing — only published.
type PostListFilter struct {
	Page     int
	PerPage  int
	Category string // category slug
	Tag      string // tag slug
	Locale   string
	Sort     string
}

// TagWithCount represents a tag with its published post count.
type TagWithCount struct {
	model.Tag
	PostCount int64 `json:"post_count" bun:"post_count"`
}

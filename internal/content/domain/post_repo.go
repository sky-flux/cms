package domain

import (
	"context"
	"errors"
)

var (
	ErrPostNotFound    = errors.New("post not found")
	ErrSlugConflict    = errors.New("slug already exists")
	ErrVersionConflict = errors.New("post was modified by another request")
)

// PostFilter carries list query parameters.
type PostFilter struct {
	Status   *PostStatus
	AuthorID string
	Page     int
	PerPage  int
}

// PostRepository is the persistence port for the Post aggregate.
type PostRepository interface {
	Save(ctx context.Context, p *Post) error
	FindByID(ctx context.Context, id string) (*Post, error)
	FindBySlug(ctx context.Context, slug string) (*Post, error)
	Update(ctx context.Context, p *Post, expectedVersion int) error
	SoftDelete(ctx context.Context, id string) error
	List(ctx context.Context, f PostFilter) ([]*Post, int64, error)
	SlugExists(ctx context.Context, slug, excludeID string) (bool, error)
}

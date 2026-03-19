package app

import (
	"context"
	"errors"

	"github.com/sky-flux/cms/internal/content/domain"
)

// ErrInvalidFilter is returned when list parameters are invalid.
var ErrInvalidFilter = errors.New("invalid list filter: page must be >= 1")

// ListPostsUseCase lists posts with pagination and optional filters.
type ListPostsUseCase struct {
	posts domain.PostRepository
}

func NewListPostsUseCase(posts domain.PostRepository) *ListPostsUseCase {
	return &ListPostsUseCase{posts: posts}
}

func (uc *ListPostsUseCase) Execute(ctx context.Context, f domain.PostFilter) ([]*domain.Post, int64, error) {
	if f.Page < 1 {
		return nil, 0, ErrInvalidFilter
	}
	if f.PerPage < 1 {
		f.PerPage = 20
	}
	return uc.posts.List(ctx, f)
}

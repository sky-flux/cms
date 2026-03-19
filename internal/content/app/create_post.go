package app

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/content/domain"
)

// CreatePostInput carries validated input from the delivery layer.
type CreatePostInput struct {
	Title    string
	Slug     string
	AuthorID string
	Content  string
	Excerpt  string
}

// CreatePostUseCase orchestrates post creation.
type CreatePostUseCase struct {
	posts domain.PostRepository
}

func NewCreatePostUseCase(posts domain.PostRepository) *CreatePostUseCase {
	return &CreatePostUseCase{posts: posts}
}

func (uc *CreatePostUseCase) Execute(ctx context.Context, in CreatePostInput) (*domain.Post, error) {
	// Check slug uniqueness before constructing entity.
	exists, err := uc.posts.SlugExists(ctx, in.Slug, "")
	if err != nil {
		return nil, fmt.Errorf("check slug: %w", err)
	}
	if exists {
		return nil, domain.ErrSlugConflict
	}

	post, err := domain.NewPost(in.Title, in.Slug, in.AuthorID)
	if err != nil {
		return nil, err
	}
	post.Content = in.Content
	post.Excerpt = in.Excerpt

	if err := uc.posts.Save(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

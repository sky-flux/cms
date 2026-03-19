package app

import (
	"context"

	"github.com/sky-flux/cms/internal/content/domain"
)

// PublishPostInput carries the post ID and optimistic lock version.
type PublishPostInput struct {
	PostID          string
	ExpectedVersion int
}

// PublishPostUseCase transitions a post to published state.
type PublishPostUseCase struct {
	posts domain.PostRepository
}

func NewPublishPostUseCase(posts domain.PostRepository) *PublishPostUseCase {
	return &PublishPostUseCase{posts: posts}
}

func (uc *PublishPostUseCase) Execute(ctx context.Context, in PublishPostInput) (*domain.Post, error) {
	post, err := uc.posts.FindByID(ctx, in.PostID)
	if err != nil {
		return nil, err
	}

	if err := post.Publish(); err != nil {
		return nil, err
	}

	if err := uc.posts.Update(ctx, post, in.ExpectedVersion); err != nil {
		return nil, err
	}
	return post, nil
}

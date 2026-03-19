package app_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/content/app"
	"github.com/sky-flux/cms/internal/content/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func draftPost(id string) *domain.Post {
	p, _ := domain.NewPost("Test Post", "test-post", "author-1")
	p.ID = id
	return p
}

func TestPublishPostUseCase_Success(t *testing.T) {
	post := draftPost("post-1")
	repo := &mockPostRepo{
		findByIDFn: func(_ context.Context, id string) (*domain.Post, error) {
			return post, nil
		},
		updateFn: func(_ context.Context, p *domain.Post, _ int) error {
			return nil
		},
	}
	uc := app.NewPublishPostUseCase(repo)

	out, err := uc.Execute(context.Background(), app.PublishPostInput{
		PostID:          "post-1",
		ExpectedVersion: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.PostStatusPublished, out.Status)
	assert.NotNil(t, out.PublishedAt)
}

func TestPublishPostUseCase_PostNotFound(t *testing.T) {
	uc := app.NewPublishPostUseCase(&mockPostRepo{})

	_, err := uc.Execute(context.Background(), app.PublishPostInput{PostID: "ghost"})
	assert.ErrorIs(t, err, domain.ErrPostNotFound)
}

func TestPublishPostUseCase_AlreadyPublished(t *testing.T) {
	post := draftPost("post-1")
	_ = post.Publish() // already published

	repo := &mockPostRepo{
		findByIDFn: func(_ context.Context, _ string) (*domain.Post, error) {
			return post, nil
		},
	}
	uc := app.NewPublishPostUseCase(repo)

	_, err := uc.Execute(context.Background(), app.PublishPostInput{PostID: "post-1"})
	assert.ErrorIs(t, err, domain.ErrInvalidTransition)
}

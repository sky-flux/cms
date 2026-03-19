package domain_test

import (
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/content/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPost_ValidInput(t *testing.T) {
	p, err := domain.NewPost("Hello World", "hello-world", "author-id")
	require.NoError(t, err)
	assert.Equal(t, "Hello World", p.Title)
	assert.Equal(t, "hello-world", p.Slug)
	assert.Equal(t, domain.PostStatusDraft, p.Status)
	assert.Equal(t, 1, p.Version)
}

func TestNewPost_EmptyTitle(t *testing.T) {
	_, err := domain.NewPost("", "slug", "author-id")
	assert.ErrorIs(t, err, domain.ErrEmptyTitle)
}

func TestNewPost_EmptySlug(t *testing.T) {
	_, err := domain.NewPost("Title", "", "author-id")
	assert.ErrorIs(t, err, domain.ErrEmptySlug)
}

func TestPost_Publish_FromDraft(t *testing.T) {
	p, _ := domain.NewPost("Title", "slug", "author-id")
	err := p.Publish()
	require.NoError(t, err)
	assert.Equal(t, domain.PostStatusPublished, p.Status)
	assert.NotNil(t, p.PublishedAt)
}

func TestPost_Publish_AlreadyPublished(t *testing.T) {
	p, _ := domain.NewPost("Title", "slug", "author-id")
	_ = p.Publish()
	err := p.Publish()
	assert.ErrorIs(t, err, domain.ErrInvalidTransition)
}

func TestPost_Archive_FromPublished(t *testing.T) {
	p, _ := domain.NewPost("Title", "slug", "author-id")
	_ = p.Publish()
	err := p.Archive()
	require.NoError(t, err)
	assert.Equal(t, domain.PostStatusArchived, p.Status)
}

func TestPost_Archive_FromDraft(t *testing.T) {
	p, _ := domain.NewPost("Title", "slug", "author-id")
	err := p.Archive()
	// Draft → Archived is not a valid transition.
	assert.ErrorIs(t, err, domain.ErrInvalidTransition)
}

func TestPost_Unpublish_FromPublished(t *testing.T) {
	p, _ := domain.NewPost("Title", "slug", "author-id")
	_ = p.Publish()
	err := p.Unpublish()
	require.NoError(t, err)
	assert.Equal(t, domain.PostStatusDraft, p.Status)
}

func TestPost_Schedule_WithFutureTime(t *testing.T) {
	p, _ := domain.NewPost("Title", "slug", "author-id")
	future := time.Now().Add(24 * time.Hour)
	err := p.Schedule(future)
	require.NoError(t, err)
	assert.Equal(t, domain.PostStatusScheduled, p.Status)
	require.NotNil(t, p.ScheduledAt)
	assert.True(t, p.ScheduledAt.Equal(future))
}

func TestPost_Schedule_WithPastTime(t *testing.T) {
	p, _ := domain.NewPost("Title", "slug", "author-id")
	past := time.Now().Add(-1 * time.Hour)
	err := p.Schedule(past)
	assert.ErrorIs(t, err, domain.ErrScheduledAtInPast)
}

func TestPost_IncrementVersion(t *testing.T) {
	p, _ := domain.NewPost("Title", "slug", "author-id")
	assert.Equal(t, 1, p.Version)
	p.IncrementVersion()
	assert.Equal(t, 2, p.Version)
}

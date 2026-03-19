package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyTitle        = errors.New("post title must not be empty")
	ErrEmptySlug         = errors.New("post slug must not be empty")
	ErrInvalidTransition = errors.New("invalid post status transition")
	ErrScheduledAtInPast = errors.New("scheduled_at must be in the future")
)

// PostStatus mirrors model.PostStatus. Domain layer owns this type.
type PostStatus int8

const (
	PostStatusDraft     PostStatus = 1
	PostStatusScheduled PostStatus = 2
	PostStatusPublished PostStatus = 3
	PostStatusArchived  PostStatus = 4
)

// Post is the aggregate root for the Content BC.
type Post struct {
	ID          string
	AuthorID    string
	Title       string
	Slug        string
	Excerpt     string
	Content     string
	Status      PostStatus
	Version     int
	PublishedAt *time.Time
	ScheduledAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Relations (optional, loaded by repo)
	CategoryIDs []string
	TagIDs      []string
}

// NewPost validates inputs and returns a draft Post ready for persistence.
func NewPost(title, slug, authorID string) (*Post, error) {
	if strings.TrimSpace(title) == "" {
		return nil, ErrEmptyTitle
	}
	if strings.TrimSpace(slug) == "" {
		return nil, ErrEmptySlug
	}
	return &Post{
		Title:    title,
		Slug:     slug,
		AuthorID: authorID,
		Status:   PostStatusDraft,
		Version:  1,
	}, nil
}

// Publish transitions draft or scheduled → published.
func (p *Post) Publish() error {
	if p.Status != PostStatusDraft && p.Status != PostStatusScheduled {
		return ErrInvalidTransition
	}
	now := time.Now()
	p.Status = PostStatusPublished
	p.PublishedAt = &now
	return nil
}

// Unpublish transitions published → draft.
func (p *Post) Unpublish() error {
	if p.Status != PostStatusPublished {
		return ErrInvalidTransition
	}
	p.Status = PostStatusDraft
	return nil
}

// Archive transitions published → archived.
func (p *Post) Archive() error {
	if p.Status != PostStatusPublished {
		return ErrInvalidTransition
	}
	p.Status = PostStatusArchived
	return nil
}

// Schedule transitions draft → scheduled with a future timestamp.
func (p *Post) Schedule(at time.Time) error {
	if !at.After(time.Now()) {
		return ErrScheduledAtInPast
	}
	p.Status = PostStatusScheduled
	p.ScheduledAt = &at
	return nil
}

// IncrementVersion bumps the optimistic lock counter on update.
func (p *Post) IncrementVersion() { p.Version++ }

// IsPublished returns true if the post is in published state.
func (p *Post) IsPublished() bool { return p.Status == PostStatusPublished }

// IsDraft returns true if the post is in draft state.
func (p *Post) IsDraft() bool { return p.Status == PostStatusDraft }

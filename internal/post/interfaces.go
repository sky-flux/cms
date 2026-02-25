package post

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/search"
)

// PostRepository handles sfc_site_posts table.
type PostRepository interface {
	List(ctx context.Context, f ListFilter) ([]model.Post, int64, error)
	GetByID(ctx context.Context, id string) (*model.Post, error)
	GetByIDUnscoped(ctx context.Context, id string) (*model.Post, error) // includes soft-deleted
	Create(ctx context.Context, post *model.Post) error
	Update(ctx context.Context, post *model.Post, expectedVersion int) error
	SoftDelete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	SlugExists(ctx context.Context, slug, excludeID string) (bool, error)
	UpdateStatus(ctx context.Context, id string, status model.PostStatus) error
	SyncCategories(ctx context.Context, postID string, categoryIDs []string, primaryID string) error
	SyncTags(ctx context.Context, postID string, tagIDs []string) error
	LoadRelations(ctx context.Context, post *model.Post) error
}

// RevisionRepository handles sfc_site_post_revisions.
type RevisionRepository interface {
	List(ctx context.Context, postID string) ([]model.PostRevision, error)
	GetByID(ctx context.Context, id string) (*model.PostRevision, error)
	Create(ctx context.Context, rev *model.PostRevision) error
}

// TranslationRepository handles sfc_site_post_translations.
type TranslationRepository interface {
	List(ctx context.Context, postID string) ([]model.PostTranslation, error)
	Get(ctx context.Context, postID, locale string) (*model.PostTranslation, error)
	Upsert(ctx context.Context, t *model.PostTranslation) error
	Delete(ctx context.Context, postID, locale string) error
}

// PreviewTokenRepository handles sfc_site_preview_tokens.
type PreviewTokenRepository interface {
	List(ctx context.Context, postID string) ([]model.PreviewToken, error)
	Create(ctx context.Context, token *model.PreviewToken) error
	CountActive(ctx context.Context, postID string) (int, error)
	DeleteAll(ctx context.Context, postID string) (int64, error)
	DeleteByID(ctx context.Context, id string) error
	GetByHash(ctx context.Context, hash string) (*model.PreviewToken, error)
}

// ListFilter holds query parameters for listing posts.
type ListFilter struct {
	Page           int
	PerPage        int
	Status         string
	Query          string
	CategoryID     string
	TagID          string
	AuthorID       string
	Sort           string
	IncludeDeleted bool
}

// These interfaces represent external dependencies injected into the service.
// They are defined here as narrow interfaces following the Go convention.
var (
	_ audit.Logger   = (*audit.Service)(nil)    // compile-time check
	_ *search.Client = (*search.Client)(nil)
)

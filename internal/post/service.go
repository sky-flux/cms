package post

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/search"
)

const postSearchIndex = "posts-%s" // posts-{siteSlug}

// Service implements post business logic.
type Service struct {
	posts    PostRepository
	revs     RevisionRepository
	trans    TranslationRepository
	preview  PreviewTokenRepository
	search   *search.Client
	auditLog audit.Logger
}

// NewService creates a new post service.
func NewService(
	posts PostRepository,
	revs RevisionRepository,
	trans TranslationRepository,
	preview PreviewTokenRepository,
	s *search.Client,
	auditLog audit.Logger,
) *Service {
	return &Service{
		posts:    posts,
		revs:     revs,
		trans:    trans,
		preview:  preview,
		search:   s,
		auditLog: auditLog,
	}
}

// ListPosts returns a paginated list of posts.
func (s *Service) ListPosts(ctx context.Context, f ListFilter) ([]model.Post, int64, error) {
	if f.Page <= 0 {
		return nil, 0, apperror.Validation("page must be positive", nil)
	}
	if f.PerPage <= 0 {
		return nil, 0, apperror.Validation("per_page must be positive", nil)
	}
	return s.posts.List(ctx, f)
}

// CreatePost creates a new post with optional category/tag associations.
func (s *Service) CreatePost(ctx context.Context, siteSlug, authorID string, req *CreatePostReq) (*model.Post, error) {
	// Generate or validate slug.
	slug := req.Slug
	if slug == "" {
		var err error
		slug, err = UniqueSlug(ctx, req.Title, "", s.posts.SlugExists)
		if err != nil {
			return nil, err
		}
	} else {
		exists, err := s.posts.SlugExists(ctx, slug, "")
		if err != nil {
			return nil, fmt.Errorf("check slug: %w", err)
		}
		if exists {
			return nil, apperror.Conflict("slug already exists", nil)
		}
	}

	// Determine status.
	status := model.PostStatusDraft
	if req.Status != "" {
		var ok bool
		status, ok = parseStatus(req.Status)
		if !ok {
			return nil, apperror.Validation("invalid status", nil)
		}
	}

	// Validate scheduled_at.
	if status == model.PostStatusScheduled {
		if req.ScheduledAt == nil {
			return nil, apperror.Validation("scheduled_at is required for scheduled status", nil)
		}
		if req.ScheduledAt.Before(time.Now()) {
			return nil, apperror.Validation("scheduled_at must be in the future", nil)
		}
	}

	post := &model.Post{
		AuthorID:     authorID,
		Title:        req.Title,
		Slug:         slug,
		Content:      req.Content,
		ContentJSON:  req.ContentJSON,
		Excerpt:      req.Excerpt,
		Status:       status,
		ScheduledAt:  req.ScheduledAt,
		CoverImageID: req.CoverImageID,
		MetaTitle:    req.MetaTitle,
		MetaDesc:     req.MetaDescription,
		OGImageURL:   req.OGImageURL,
		ExtraFields:  req.ExtraFields,
	}

	if status == model.PostStatusPublished {
		now := time.Now()
		post.PublishedAt = &now
	}

	if err := s.posts.Create(ctx, post); err != nil {
		return nil, err
	}

	// Sync categories and tags.
	if len(req.CategoryIDs) > 0 {
		if err := s.posts.SyncCategories(ctx, post.ID, req.CategoryIDs, req.PrimaryCategoryID); err != nil {
			return nil, fmt.Errorf("sync categories: %w", err)
		}
	}
	if len(req.TagIDs) > 0 {
		if err := s.posts.SyncTags(ctx, post.ID, req.TagIDs); err != nil {
			return nil, fmt.Errorf("sync tags: %w", err)
		}
	}

	// Create initial revision (version 1).
	rev := &model.PostRevision{
		PostID:      post.ID,
		EditorID:    authorID,
		Version:     1,
		Title:       post.Title,
		Content:     post.Content,
		ContentJSON: post.ContentJSON,
		DiffSummary: "Initial version",
	}
	if err := s.revs.Create(ctx, rev); err != nil {
		slog.Error("create initial revision failed", "error", err, "post_id", post.ID)
	}

	// Async: push to Meilisearch.
	go s.indexPost(context.Background(), siteSlug, post)

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionCreate,
		ResourceType: "post",
		ResourceID:   post.ID,
	}); err != nil {
		slog.Error("audit log post create failed", "error", err)
	}

	return post, nil
}

// GetPost returns a single post by ID.
func (s *Service) GetPost(ctx context.Context, id string) (*model.Post, error) {
	return s.posts.GetByID(ctx, id)
}

// UpdatePost updates a post with optimistic locking and creates a revision.
func (s *Service) UpdatePost(ctx context.Context, siteSlug, editorID, id string, req *UpdatePostReq) (*model.Post, error) {
	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldTitle := post.Title
	oldContent := post.Content
	oldExcerpt := post.Excerpt

	// Apply updates.
	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.Content = *req.Content
	}
	if req.ContentJSON != nil {
		post.ContentJSON = req.ContentJSON
	}
	if req.Excerpt != nil {
		post.Excerpt = *req.Excerpt
	}
	if req.MetaTitle != nil {
		post.MetaTitle = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		post.MetaDesc = *req.MetaDescription
	}
	if req.OGImageURL != nil {
		post.OGImageURL = *req.OGImageURL
	}
	if req.ExtraFields != nil {
		post.ExtraFields = req.ExtraFields
	}
	if req.CoverImageID != nil {
		post.CoverImageID = req.CoverImageID
	}
	if req.ScheduledAt != nil {
		post.ScheduledAt = req.ScheduledAt
	}

	// Handle slug change.
	if req.Slug != nil && *req.Slug != post.Slug {
		exists, err := s.posts.SlugExists(ctx, *req.Slug, id)
		if err != nil {
			return nil, fmt.Errorf("check slug: %w", err)
		}
		if exists {
			return nil, apperror.Conflict("slug already exists", nil)
		}
		post.Slug = *req.Slug
	}

	// Handle status transition.
	if req.Status != nil {
		newStatus, ok := parseStatus(*req.Status)
		if !ok {
			return nil, apperror.Validation("invalid status", nil)
		}
		if newStatus != post.Status {
			if err := validateTransition(post.Status, newStatus); err != nil {
				return nil, err
			}
			post.Status = newStatus
			if newStatus == model.PostStatusPublished && post.PublishedAt == nil {
				now := time.Now()
				post.PublishedAt = &now
			}
		}
	}

	// Optimistic lock update.
	expectedVersion := req.Version
	if err := s.posts.Update(ctx, post, expectedVersion); err != nil {
		return nil, err
	}

	// Sync categories/tags if provided.
	if req.CategoryIDs != nil {
		primaryID := ""
		if req.PrimaryCategoryID != nil {
			primaryID = *req.PrimaryCategoryID
		}
		if err := s.posts.SyncCategories(ctx, post.ID, req.CategoryIDs, primaryID); err != nil {
			return nil, fmt.Errorf("sync categories: %w", err)
		}
	}
	if req.TagIDs != nil {
		if err := s.posts.SyncTags(ctx, post.ID, req.TagIDs); err != nil {
			return nil, fmt.Errorf("sync tags: %w", err)
		}
	}

	// Create revision.
	diffSummary := buildDiffSummary(oldTitle, post.Title, oldContent, post.Content, oldExcerpt, post.Excerpt)
	rev := &model.PostRevision{
		PostID:      post.ID,
		EditorID:    editorID,
		Version:     post.Version, // version was incremented by BeforeAppendModel
		Title:       post.Title,
		Content:     post.Content,
		ContentJSON: post.ContentJSON,
		DiffSummary: diffSummary,
	}
	if err := s.revs.Create(ctx, rev); err != nil {
		slog.Error("create revision failed", "error", err, "post_id", post.ID)
	}

	// Async: update search index.
	go s.indexPost(context.Background(), siteSlug, post)

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "post",
		ResourceID:   post.ID,
	}); err != nil {
		slog.Error("audit log post update failed", "error", err)
	}

	return post, nil
}

// DeletePost soft-deletes a post.
func (s *Service) DeletePost(ctx context.Context, siteSlug, id string) error {
	if err := s.posts.SoftDelete(ctx, id); err != nil {
		return err
	}

	// Remove from search index.
	go func() {
		idx := fmt.Sprintf(postSearchIndex, siteSlug)
		if err := s.search.DeleteDocuments(context.Background(), idx, []string{id}); err != nil {
			slog.Error("remove post from search failed", "error", err, "post_id", id)
		}
	}()

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "post",
		ResourceID:   id,
	}); err != nil {
		slog.Error("audit log post delete failed", "error", err)
	}

	return nil
}

// indexPost pushes a post to Meilisearch.
func (s *Service) indexPost(ctx context.Context, siteSlug string, post *model.Post) {
	idx := fmt.Sprintf(postSearchIndex, siteSlug)
	doc := map[string]any{
		"id":      post.ID,
		"title":   post.Title,
		"excerpt": post.Excerpt,
		"slug":    post.Slug,
		"status":  statusString(post.Status),
	}
	if err := s.search.UpsertDocuments(ctx, idx, []map[string]any{doc}); err != nil {
		slog.Error("index post failed", "error", err, "post_id", post.ID)
	}
}

// buildDiffSummary generates a field-level change summary.
func buildDiffSummary(oldTitle, newTitle, oldContent, newContent, oldExcerpt, newExcerpt string) string {
	var changed []string
	if oldTitle != newTitle {
		changed = append(changed, "title")
	}
	if oldContent != newContent {
		changed = append(changed, "content")
	}
	if oldExcerpt != newExcerpt {
		changed = append(changed, "excerpt")
	}
	if len(changed) == 0 {
		return "No content changes"
	}
	return "Updated " + strings.Join(changed, ", ")
}

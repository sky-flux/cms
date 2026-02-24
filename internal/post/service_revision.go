package post

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
)

// ListRevisions returns the revision history for a post.
func (s *Service) ListRevisions(ctx context.Context, postID string) ([]model.PostRevision, error) {
	// Verify post exists.
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}
	return s.revs.List(ctx, postID)
}

// Rollback restores a post to a specific revision's content.
// Creates a new revision (version continues incrementing).
func (s *Service) Rollback(ctx context.Context, siteSlug, editorID, postID, revisionID string) (*model.Post, error) {
	// Get the target revision.
	rev, err := s.revs.GetByID(ctx, revisionID)
	if err != nil {
		return nil, err
	}

	// Verify the revision belongs to this post.
	if rev.PostID != postID {
		return nil, fmt.Errorf("revision does not belong to this post")
	}

	// Get current post.
	post, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Apply revision content to current post.
	post.Title = rev.Title
	post.Content = rev.Content
	post.ContentJSON = rev.ContentJSON

	// Save with current version (optimistic lock).
	currentVersion := post.Version
	if err := s.posts.Update(ctx, post, currentVersion); err != nil {
		return nil, err
	}

	// Create a new revision for the rollback.
	newRev := &model.PostRevision{
		PostID:      postID,
		EditorID:    editorID,
		Version:     post.Version, // incremented by Update
		Title:       post.Title,
		Content:     post.Content,
		ContentJSON: post.ContentJSON,
		DiffSummary: fmt.Sprintf("Rolled back to version %d", rev.Version),
	}
	if err := s.revs.Create(ctx, newRev); err != nil {
		slog.Error("create rollback revision failed", "error", err, "post_id", postID)
	}

	// Async: update search index.
	go s.indexPost(context.Background(), siteSlug, post)

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "post",
		ResourceID:   postID,
	}); err != nil {
		slog.Error("audit log post rollback failed", "error", err)
	}

	return post, nil
}

package post

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
)

// Allowed state transitions.
var allowedTransitions = map[model.PostStatus][]model.PostStatus{
	model.PostStatusDraft:     {model.PostStatusPublished, model.PostStatusScheduled},
	model.PostStatusScheduled: {model.PostStatusDraft, model.PostStatusPublished},
	model.PostStatusPublished: {model.PostStatusDraft, model.PostStatusArchived},
	model.PostStatusArchived:  {model.PostStatusPublished, model.PostStatusDraft},
}

// validateTransition checks if a status transition is allowed.
func validateTransition(from, to model.PostStatus) error {
	allowed, ok := allowedTransitions[from]
	if !ok {
		return apperror.Validation("invalid current status", nil)
	}
	for _, a := range allowed {
		if a == to {
			return nil
		}
	}

	// Build allowed list for error detail.
	var names []string
	for _, a := range allowed {
		names = append(names, statusString(a))
	}
	msg := fmt.Sprintf("cannot transition from %s to %s; allowed: %v", statusString(from), statusString(to), names)
	return apperror.Validation(msg, nil)
}

// Publish transitions a post to published status.
func (s *Service) Publish(ctx context.Context, siteSlug, id string) (*model.Post, error) {
	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(post.Status, model.PostStatusPublished); err != nil {
		return nil, err
	}

	if err := s.posts.UpdateStatus(ctx, id, model.PostStatusPublished); err != nil {
		return nil, err
	}

	// Async: update search index.
	post.Status = model.PostStatusPublished
	go s.indexPost(context.Background(), siteSlug, post)

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionPublish,
		ResourceType: "post",
		ResourceID:   id,
	}); err != nil {
		slog.Error("audit log post publish failed", "error", err)
	}

	// Return updated post.
	return s.posts.GetByID(ctx, id)
}

// Unpublish transitions a published post to archived.
func (s *Service) Unpublish(ctx context.Context, id string) (*model.Post, error) {
	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(post.Status, model.PostStatusArchived); err != nil {
		return nil, err
	}

	if err := s.posts.UpdateStatus(ctx, id, model.PostStatusArchived); err != nil {
		return nil, err
	}

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionUnpublish,
		ResourceType: "post",
		ResourceID:   id,
	}); err != nil {
		slog.Error("audit log post unpublish failed", "error", err)
	}

	return s.posts.GetByID(ctx, id)
}

// RevertToDraft transitions a post back to draft.
func (s *Service) RevertToDraft(ctx context.Context, id string) (*model.Post, error) {
	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(post.Status, model.PostStatusDraft); err != nil {
		return nil, err
	}

	if err := s.posts.UpdateStatus(ctx, id, model.PostStatusDraft); err != nil {
		return nil, err
	}

	return s.posts.GetByID(ctx, id)
}

// RestorePost restores a soft-deleted post to draft status.
func (s *Service) RestorePost(ctx context.Context, id string) (*model.Post, error) {
	if err := s.posts.Restore(ctx, id); err != nil {
		return nil, err
	}

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionRestore,
		ResourceType: "post",
		ResourceID:   id,
	}); err != nil {
		slog.Error("audit log post restore failed", "error", err)
	}

	return s.posts.GetByID(ctx, id)
}

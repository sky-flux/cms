package post

import (
	"context"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
)

// ListTranslations returns all translations for a post.
func (s *Service) ListTranslations(ctx context.Context, postID string) ([]model.PostTranslation, error) {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}
	return s.trans.List(ctx, postID)
}

// GetTranslation returns a specific locale translation.
func (s *Service) GetTranslation(ctx context.Context, postID, locale string) (*model.PostTranslation, error) {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}
	return s.trans.Get(ctx, postID, locale)
}

// UpsertTranslation creates or updates a translation for a post.
func (s *Service) UpsertTranslation(ctx context.Context, postID, locale string, req *UpsertTranslationReq) (*model.PostTranslation, error) {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}

	t := &model.PostTranslation{
		PostID:      postID,
		Locale:      locale,
		Title:       req.Title,
		Excerpt:     req.Excerpt,
		Content:     req.Content,
		ContentJSON: req.ContentJSON,
		MetaTitle:   req.MetaTitle,
		MetaDesc:    req.MetaDescription,
		OGImageURL:  req.OGImageURL,
	}

	if err := s.trans.Upsert(ctx, t); err != nil {
		return nil, err
	}

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "post_translation",
		ResourceID:   postID,
	}); err != nil {
		slog.Error("audit log translation upsert failed", "error", err)
	}

	// Return the saved translation.
	return s.trans.Get(ctx, postID, locale)
}

// DeleteTranslation removes a specific locale translation.
func (s *Service) DeleteTranslation(ctx context.Context, postID, locale string) error {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return err
	}

	if err := s.trans.Delete(ctx, postID, locale); err != nil {
		return err
	}

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "post_translation",
		ResourceID:   postID,
	}); err != nil {
		slog.Error("audit log translation delete failed", "error", err)
	}

	return nil
}

package post

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/crypto"
)

const (
	previewTokenTTL   = 24 * time.Hour
	previewMaxPerPost = 5
	previewTokenBytes = 32
)

// CreatePreviewToken generates a new preview token for a post.
func (s *Service) CreatePreviewToken(ctx context.Context, postID, creatorID string) (*PreviewTokenResp, error) {
	// Verify post exists and is not deleted.
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}

	// Check active token count.
	count, err := s.preview.CountActive(ctx, postID)
	if err != nil {
		return nil, err
	}
	if count >= previewMaxPerPost {
		return nil, apperror.Validation("preview token limit reached (max 5 active tokens per post)", nil)
	}

	// Generate token.
	raw, hash, err := crypto.GenerateToken(previewTokenBytes)
	if err != nil {
		return nil, fmt.Errorf("generate preview token: %w", err)
	}

	tokenStr := "sky_preview_" + base64.RawURLEncoding.EncodeToString([]byte(raw))

	token := &model.PreviewToken{
		PostID:    postID,
		TokenHash: crypto.HashToken(tokenStr),
		ExpiresAt: time.Now().Add(previewTokenTTL),
		CreatedBy: &creatorID,
	}

	// We don't actually use `hash` since we hash the full tokenStr.
	_ = hash

	if err := s.preview.Create(ctx, token); err != nil {
		return nil, err
	}

	return &PreviewTokenResp{
		Token:       tokenStr,
		ID:          token.ID,
		ExpiresAt:   token.ExpiresAt,
		CreatedAt:   token.CreatedAt,
		ActiveCount: count + 1,
	}, nil
}

// ListPreviewTokens returns active (non-expired) tokens for a post.
func (s *Service) ListPreviewTokens(ctx context.Context, postID string) ([]model.PreviewToken, error) {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return nil, err
	}
	return s.preview.List(ctx, postID)
}

// RevokeAllPreviewTokens deletes all preview tokens for a post.
func (s *Service) RevokeAllPreviewTokens(ctx context.Context, postID string) (int64, error) {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return 0, err
	}

	count, err := s.preview.DeleteAll(ctx, postID)
	if err != nil {
		return 0, err
	}

	// Audit.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "preview_token",
		ResourceID:   postID,
	}); err != nil {
		slog.Error("audit log revoke preview tokens failed", "error", err)
	}

	return count, nil
}

// RevokePreviewToken deletes a single preview token.
func (s *Service) RevokePreviewToken(ctx context.Context, tokenID string) error {
	return s.preview.DeleteByID(ctx, tokenID)
}

// GetPostByPreviewToken looks up a post by its preview token hash.
func (s *Service) GetPostByPreviewToken(ctx context.Context, rawToken string) (*model.Post, error) {
	hash := crypto.HashToken(rawToken)
	token, err := s.preview.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	// Get the post (including soft-deleted check is done by GetByIDUnscoped
	// since previews work for any state).
	p, err := s.posts.GetByIDUnscoped(ctx, token.PostID)
	if err != nil {
		return nil, err
	}

	// If the post is hard-deleted (shouldn't happen with soft deletes).
	if p.DeletedAt != nil {
		return nil, apperror.NotFound("post not found", nil)
	}

	return p, nil
}

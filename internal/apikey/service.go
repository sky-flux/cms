package apikey

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/crypto"
)

const keyPrefix = "cms_live_"

type Service struct {
	repo    APIKeyRepository
	auditor audit.Logger
}

func NewService(repo APIKeyRepository, auditor audit.Logger) *Service {
	return &Service{repo: repo, auditor: auditor}
}

func (s *Service) ListAPIKeys(ctx context.Context) ([]model.APIKey, error) {
	return s.repo.List(ctx)
}

func (s *Service) CreateAPIKey(ctx context.Context, ownerID string, req *CreateAPIKeyReq) (*model.APIKey, string, error) {
	raw, hash, err := crypto.GenerateToken(32)
	if err != nil {
		return nil, "", fmt.Errorf("create apikey generate token: %w", err)
	}

	plainKey := keyPrefix + raw

	key := &model.APIKey{
		OwnerID:   ownerID,
		Name:      req.Name,
		KeyHash:   hash,
		KeyPrefix: keyPrefix + raw[:8],
		Status:    model.APIKeyStatusActive,
		ExpiresAt: req.ExpiresAt,
	}

	if err := s.repo.Create(ctx, key); err != nil {
		return nil, "", fmt.Errorf("create apikey insert: %w", err)
	}

	if err := s.auditor.Log(ctx, audit.Entry{
		Action:       model.LogActionCreate,
		ResourceType: "api_key",
		ResourceID:   key.ID,
	}); err != nil {
		slog.Error("audit log api key create", "error", err)
	}

	return key, plainKey, nil
}

func (s *Service) RevokeAPIKey(ctx context.Context, id string) error {
	key, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if key.Status == model.APIKeyStatusRevoked {
		return apperror.Validation("api key is already revoked", nil)
	}

	if err := s.repo.Revoke(ctx, id); err != nil {
		return fmt.Errorf("revoke apikey: %w", err)
	}

	if err := s.auditor.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "api_key",
		ResourceID:   id,
	}); err != nil {
		slog.Error("audit log api key revoke", "error", err)
	}

	return nil
}

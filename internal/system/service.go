package system

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
)

type Service struct {
	repo    ConfigRepository
	auditor audit.Logger
}

func NewService(repo ConfigRepository, auditor audit.Logger) *Service {
	return &Service{repo: repo, auditor: auditor}
}

func (s *Service) ListConfigs(ctx context.Context) ([]model.SiteConfig, error) {
	return s.repo.List(ctx)
}

func (s *Service) UpdateConfig(ctx context.Context, key string, req *UpdateConfigReq, userID string) (*model.SiteConfig, error) {
	cfg, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	oldValue := make(json.RawMessage, len(cfg.Value))
	copy(oldValue, cfg.Value)

	cfg.Value = req.Value
	cfg.UpdatedBy = &userID

	if err := s.repo.Upsert(ctx, cfg); err != nil {
		return nil, fmt.Errorf("update config: %w", err)
	}

	snapshot := map[string]json.RawMessage{
		"old_value": oldValue,
		"new_value": req.Value,
	}
	if err := s.auditor.Log(ctx, audit.Entry{
		Action:           model.LogActionSettingsChange,
		ResourceType:     "site_config",
		ResourceID:       key,
		ResourceSnapshot: snapshot,
	}); err != nil {
		slog.Error("audit config update failed", "error", err, "key", key)
	}

	return cfg, nil
}

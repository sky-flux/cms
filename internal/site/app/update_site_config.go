package app

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/site/domain"
)

// UpdateSiteConfigInput carries validated fields from the delivery layer.
type UpdateSiteConfigInput struct {
	Name        string
	Language    string
	Timezone    string
	Description string
	BaseURL     string
	LogoURL     string
}

// UpdateSiteConfigUseCase applies new configuration and upserts the record.
type UpdateSiteConfigUseCase struct {
	sites domain.SiteRepository
}

func NewUpdateSiteConfigUseCase(sites domain.SiteRepository) *UpdateSiteConfigUseCase {
	return &UpdateSiteConfigUseCase{sites: sites}
}

func (uc *UpdateSiteConfigUseCase) Execute(ctx context.Context, in UpdateSiteConfigInput) error {
	// Validate via domain constructor to reuse domain rules.
	if _, err := domain.NewSite(in.Name, in.Language, in.Timezone); err != nil {
		return err
	}

	site, err := uc.sites.GetSite(ctx)
	if err != nil {
		return fmt.Errorf("get site: %w", err)
	}
	if site == nil {
		site = &domain.Site{}
	}
	site.Update(in.Name, in.Language, in.Timezone, in.Description, in.BaseURL)
	site.LogoURL = in.LogoURL

	if err := uc.sites.Upsert(ctx, site); err != nil {
		return fmt.Errorf("upsert site: %w", err)
	}
	return nil
}

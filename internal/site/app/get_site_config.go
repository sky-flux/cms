package app

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/site/domain"
)

// GetSiteConfigUseCase returns the single site configuration record.
// v1: if no record exists yet (fresh install), returns a default Site.
type GetSiteConfigUseCase struct {
	sites domain.SiteRepository
}

func NewGetSiteConfigUseCase(sites domain.SiteRepository) *GetSiteConfigUseCase {
	return &GetSiteConfigUseCase{sites: sites}
}

func (uc *GetSiteConfigUseCase) Execute(ctx context.Context) (*domain.Site, error) {
	site, err := uc.sites.GetSite(ctx)
	if err != nil {
		return nil, fmt.Errorf("get site config: %w", err)
	}
	if site == nil {
		// Return sensible defaults before first save.
		return &domain.Site{Name: "My Site", Language: "en", Timezone: "UTC"}, nil
	}
	return site, nil
}

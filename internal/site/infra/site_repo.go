package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/sky-flux/cms/internal/site/domain"
)

type siteRow struct {
	bun.BaseModel `bun:"table:sfc_sites,alias:s"`

	ID          int       `bun:"id,pk,autoincrement"`
	Name        string    `bun:"name,notnull"`
	Slug        string    `bun:"slug,notnull"`
	Description string    `bun:"description"`
	Language    string    `bun:"language,notnull,default:'en'"`
	Timezone    string    `bun:"timezone,notnull,default:'UTC'"`
	BaseURL     string    `bun:"base_url"`
	LogoURL     string    `bun:"logo_url"`
	CreatedAt   time.Time `bun:"created_at,nullzero,default:now()"`
	UpdatedAt   time.Time `bun:"updated_at,nullzero,default:now()"`
}

// BunSiteRepo implements domain.SiteRepository using uptrace/bun.
type BunSiteRepo struct {
	db *bun.DB
}

func NewBunSiteRepo(db *bun.DB) *BunSiteRepo {
	return &BunSiteRepo{db: db}
}

func (r *BunSiteRepo) GetSite(ctx context.Context) (*domain.Site, error) {
	row := &siteRow{}
	err := r.db.NewSelect().Model(row).Limit(1).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // no site configured yet
	}
	if err != nil {
		return nil, fmt.Errorf("get site: %w", err)
	}
	return siteRowToDomain(row), nil
}

// Upsert inserts or updates the single site record (v1: always ID=1).
func (r *BunSiteRepo) Upsert(ctx context.Context, s *domain.Site) error {
	row := siteRowFromDomain(s)
	if row.ID == 0 {
		row.ID = 1 // v1 single site always has ID=1
	}
	_, err := r.db.NewInsert().Model(row).
		On("CONFLICT (id) DO UPDATE").
		Set("name = EXCLUDED.name").
		Set("slug = EXCLUDED.slug").
		Set("description = EXCLUDED.description").
		Set("language = EXCLUDED.language").
		Set("timezone = EXCLUDED.timezone").
		Set("base_url = EXCLUDED.base_url").
		Set("logo_url = EXCLUDED.logo_url").
		Set("updated_at = now()").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("upsert site: %w", err)
	}
	s.ID = 1
	return nil
}

func siteRowFromDomain(s *domain.Site) *siteRow {
	return &siteRow{
		ID:          s.ID,
		Name:        s.Name,
		Slug:        s.Slug,
		Description: s.Description,
		Language:    s.Language,
		Timezone:    s.Timezone,
		BaseURL:     s.BaseURL,
		LogoURL:     s.LogoURL,
	}
}

func siteRowToDomain(row *siteRow) *domain.Site {
	return &domain.Site{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: row.Description,
		Language:    row.Language,
		Timezone:    row.Timezone,
		BaseURL:     row.BaseURL,
		LogoURL:     row.LogoURL,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

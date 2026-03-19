package domain_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/site/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSite_ValidInput(t *testing.T) {
	s, err := domain.NewSite("My Blog", "en", "UTC")
	require.NoError(t, err)
	assert.Equal(t, "My Blog", s.Name)
	assert.Equal(t, "en", s.Language)
	assert.Equal(t, "UTC", s.Timezone)
}

func TestNewSite_EmptyName(t *testing.T) {
	_, err := domain.NewSite("", "en", "UTC")
	assert.ErrorIs(t, err, domain.ErrEmptySiteName)
}

func TestNewSite_EmptyLanguage(t *testing.T) {
	_, err := domain.NewSite("Blog", "", "UTC")
	assert.ErrorIs(t, err, domain.ErrEmptyLanguage)
}

func TestSite_Update(t *testing.T) {
	s, _ := domain.NewSite("My Blog", "en", "UTC")
	s.Update("Updated Blog", "zh-CN", "Asia/Shanghai", "A description", "https://example.com")
	assert.Equal(t, "Updated Blog", s.Name)
	assert.Equal(t, "zh-CN", s.Language)
	assert.Equal(t, "Asia/Shanghai", s.Timezone)
	assert.Equal(t, "A description", s.Description)
	assert.Equal(t, "https://example.com", s.BaseURL)
}

// ---- compile-check for SiteRepository ----

var _ domain.SiteRepository = (*mockSiteRepo)(nil)

type mockSiteRepo struct {
	getSiteFn func(ctx context.Context) (*domain.Site, error)
	upsertFn  func(ctx context.Context, s *domain.Site) error
}

func (m *mockSiteRepo) GetSite(ctx context.Context) (*domain.Site, error) {
	return m.getSiteFn(ctx)
}
func (m *mockSiteRepo) Upsert(ctx context.Context, s *domain.Site) error {
	return m.upsertFn(ctx, s)
}

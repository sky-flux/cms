package domain_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/site/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedirect_Valid301(t *testing.T) {
	r, err := domain.NewRedirect("/old-path", "/new-path", 301)
	require.NoError(t, err)
	assert.Equal(t, "/old-path", r.FromPath)
	assert.Equal(t, "/new-path", r.ToPath)
	assert.Equal(t, 301, r.StatusCode)
}

func TestNewRedirect_Valid302(t *testing.T) {
	r, err := domain.NewRedirect("/a", "/b", 302)
	require.NoError(t, err)
	assert.Equal(t, 302, r.StatusCode)
}

func TestNewRedirect_InvalidStatusCode(t *testing.T) {
	_, err := domain.NewRedirect("/a", "/b", 200)
	assert.ErrorIs(t, err, domain.ErrInvalidStatusCode)
}

func TestNewRedirect_FromPathMissingSlash(t *testing.T) {
	_, err := domain.NewRedirect("no-slash", "/b", 301)
	assert.ErrorIs(t, err, domain.ErrInvalidFromPath)
}

func TestNewRedirect_FromPathContainsQuery(t *testing.T) {
	_, err := domain.NewRedirect("/path?q=1", "/b", 301)
	assert.ErrorIs(t, err, domain.ErrInvalidFromPath)
}

func TestNewRedirect_EmptyToPath(t *testing.T) {
	_, err := domain.NewRedirect("/from", "", 301)
	assert.ErrorIs(t, err, domain.ErrEmptyToPath)
}

func TestRedirect_IsPermanent(t *testing.T) {
	r301, _ := domain.NewRedirect("/a", "/b", 301)
	assert.True(t, r301.IsPermanent())

	r302, _ := domain.NewRedirect("/a", "/b", 302)
	assert.False(t, r302.IsPermanent())
}

// ---- compile-check for RedirectRepository ----

var _ domain.RedirectRepository = (*mockRedirectRepo)(nil)

type mockRedirectRepo struct {
	saveFn       func(ctx context.Context, r *domain.Redirect) error
	findByPathFn func(ctx context.Context, fromPath string) (*domain.Redirect, error)
	listFn       func(ctx context.Context, offset, limit int) ([]*domain.Redirect, int, error)
	deleteFn     func(ctx context.Context, id string) error
}

func (m *mockRedirectRepo) Save(ctx context.Context, r *domain.Redirect) error {
	return m.saveFn(ctx, r)
}
func (m *mockRedirectRepo) FindByPath(ctx context.Context, fromPath string) (*domain.Redirect, error) {
	return m.findByPathFn(ctx, fromPath)
}
func (m *mockRedirectRepo) List(ctx context.Context, offset, limit int) ([]*domain.Redirect, int, error) {
	return m.listFn(ctx, offset, limit)
}
func (m *mockRedirectRepo) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

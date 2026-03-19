package domain_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/media/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMediaFile_ValidImage(t *testing.T) {
	f, err := domain.NewMediaFile("photo.jpg", "image/jpeg", 1024*1024, "media/2026/03/abc.jpg", "http://cdn/media/2026/03/abc.jpg")
	require.NoError(t, err)
	assert.Equal(t, "photo.jpg", f.Filename)
	assert.Equal(t, "image/jpeg", f.MimeType)
	assert.Equal(t, int64(1024*1024), f.Size)
	assert.Equal(t, "media/2026/03/abc.jpg", f.StorageKey)
	assert.Equal(t, "http://cdn/media/2026/03/abc.jpg", f.URL)
}

func TestNewMediaFile_EmptyFilename(t *testing.T) {
	_, err := domain.NewMediaFile("", "image/jpeg", 100, "key", "url")
	assert.ErrorIs(t, err, domain.ErrEmptyFilename)
}

func TestNewMediaFile_UnsupportedMIME(t *testing.T) {
	_, err := domain.NewMediaFile("doc.pdf", "application/pdf", 100, "key", "url")
	assert.ErrorIs(t, err, domain.ErrUnsupportedMIMEType)
}

func TestNewMediaFile_TooLarge(t *testing.T) {
	_, err := domain.NewMediaFile("big.jpg", "image/jpeg", 21*1024*1024, "key", "url")
	assert.ErrorIs(t, err, domain.ErrFileTooLarge)
}

func TestMediaFile_IsImage(t *testing.T) {
	f, _ := domain.NewMediaFile("a.png", "image/png", 100, "k", "u")
	assert.True(t, f.IsImage())
}

func TestMediaFile_SetDimensions(t *testing.T) {
	f, _ := domain.NewMediaFile("a.jpg", "image/jpeg", 100, "k", "u")
	f.SetDimensions(1920, 1080)
	assert.Equal(t, 1920, f.Width)
	assert.Equal(t, 1080, f.Height)
}

func TestMediaFile_SetThumbnails(t *testing.T) {
	f, _ := domain.NewMediaFile("a.jpg", "image/jpeg", 100, "k", "u")
	f.SetThumbnails("media/2026/03/abc_sm.jpg", "http://cdn/abc_sm.jpg",
		"media/2026/03/abc_md.jpg", "http://cdn/abc_md.jpg")
	assert.Equal(t, "media/2026/03/abc_sm.jpg", f.ThumbSmKey)
	assert.Equal(t, "http://cdn/abc_sm.jpg", f.ThumbSmURL)
	assert.Equal(t, "media/2026/03/abc_md.jpg", f.ThumbMdKey)
	assert.Equal(t, "http://cdn/abc_md.jpg", f.ThumbMdURL)
}

func TestMediaFile_Validate_EmptyStorageKey(t *testing.T) {
	f := &domain.MediaFile{Filename: "a.jpg", MimeType: "image/jpeg", Size: 100}
	assert.ErrorIs(t, f.Validate(), domain.ErrEmptyStorageKey)
}

// Compile-time interface checks (hand-written mocks used by app layer tests).
var _ domain.MediaFileRepository = (*mockMediaRepo)(nil)
var _ domain.StoragePort = (*mockStorage)(nil)

type mockMediaRepo struct {
	saveFn     func(ctx context.Context, f *domain.MediaFile) error
	findByIDFn func(ctx context.Context, id string) (*domain.MediaFile, error)
	listFn     func(ctx context.Context, offset, limit int) ([]*domain.MediaFile, int, error)
	deleteFn   func(ctx context.Context, id string) error
}

func (m *mockMediaRepo) Save(ctx context.Context, f *domain.MediaFile) error {
	return m.saveFn(ctx, f)
}
func (m *mockMediaRepo) FindByID(ctx context.Context, id string) (*domain.MediaFile, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockMediaRepo) List(ctx context.Context, offset, limit int) ([]*domain.MediaFile, int, error) {
	return m.listFn(ctx, offset, limit)
}
func (m *mockMediaRepo) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

type mockStorage struct {
	uploadFn func(ctx context.Context, key string, data []byte, contentType string) error
	deleteFn func(ctx context.Context, key string) error
	urlFn    func(key string) string
}

func (m *mockStorage) Upload(ctx context.Context, key string, data []byte, contentType string) error {
	return m.uploadFn(ctx, key, data, contentType)
}
func (m *mockStorage) Delete(ctx context.Context, key string) error {
	return m.deleteFn(ctx, key)
}
func (m *mockStorage) URL(key string) string { return m.urlFn(key) }

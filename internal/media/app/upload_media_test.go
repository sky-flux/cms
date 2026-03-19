package app_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/media/app"
	"github.com/sky-flux/cms/internal/media/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- hand-written mocks ----

type mockRepo struct {
	saved *domain.MediaFile
	err   error
}

func (m *mockRepo) Save(ctx context.Context, f *domain.MediaFile) error {
	m.saved = f
	return m.err
}
func (m *mockRepo) FindByID(ctx context.Context, id string) (*domain.MediaFile, error) {
	return nil, nil
}
func (m *mockRepo) List(ctx context.Context, offset, limit int) ([]*domain.MediaFile, int, error) {
	return nil, 0, nil
}
func (m *mockRepo) Delete(ctx context.Context, id string) error { return nil }

type mockStorage struct {
	uploadedKey string
	err         error
}

func (m *mockStorage) Upload(ctx context.Context, key string, data []byte, contentType string) error {
	m.uploadedKey = key
	return m.err
}
func (m *mockStorage) Delete(ctx context.Context, key string) error { return nil }
func (m *mockStorage) URL(key string) string                        { return "http://cdn/" + key }

// ---- tests ----

func TestUploadMedia_Success(t *testing.T) {
	repo := &mockRepo{}
	stor := &mockStorage{}
	uc := app.NewUploadMediaUseCase(repo, stor)

	data := bytes.Repeat([]byte("x"), 512)
	in := app.UploadMediaInput{
		Filename:   "test.jpg",
		MimeType:   "image/jpeg",
		Data:       data,
		UploaderID: "user-1",
	}

	out, err := uc.Execute(context.Background(), in)
	require.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, "test.jpg", out.Filename)
	assert.NotEmpty(t, stor.uploadedKey)
	assert.NotNil(t, repo.saved)
}

func TestUploadMedia_UnsupportedMIME(t *testing.T) {
	repo := &mockRepo{}
	stor := &mockStorage{}
	uc := app.NewUploadMediaUseCase(repo, stor)

	in := app.UploadMediaInput{
		Filename: "doc.pdf",
		MimeType: "application/pdf",
		Data:     []byte("data"),
	}
	_, err := uc.Execute(context.Background(), in)
	assert.ErrorIs(t, err, domain.ErrUnsupportedMIMEType)
}

func TestUploadMedia_TooLarge(t *testing.T) {
	repo := &mockRepo{}
	stor := &mockStorage{}
	uc := app.NewUploadMediaUseCase(repo, stor)

	in := app.UploadMediaInput{
		Filename: "big.jpg",
		MimeType: "image/jpeg",
		Data:     bytes.Repeat([]byte("x"), 21*1024*1024),
	}
	_, err := uc.Execute(context.Background(), in)
	assert.ErrorIs(t, err, domain.ErrFileTooLarge)
}

func TestUploadMedia_StorageError(t *testing.T) {
	repo := &mockRepo{}
	stor := &mockStorage{err: assert.AnError}
	uc := app.NewUploadMediaUseCase(repo, stor)

	in := app.UploadMediaInput{
		Filename: "a.jpg",
		MimeType: "image/jpeg",
		Data:     []byte("data"),
	}
	_, err := uc.Execute(context.Background(), in)
	require.Error(t, err)
}

package delivery_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/sky-flux/cms/internal/media/app"
	"github.com/sky-flux/cms/internal/media/delivery"
	"github.com/sky-flux/cms/internal/media/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- hand-written mocks ----

type mockUploader struct {
	out *domain.MediaFile
	err error
}

func (m *mockUploader) Execute(ctx context.Context, in app.UploadMediaInput) (*domain.MediaFile, error) {
	return m.out, m.err
}

type mockLister struct {
	files []*domain.MediaFile
	total int
	err   error
}

func (m *mockLister) Execute(ctx context.Context, offset, limit int) ([]*domain.MediaFile, int, error) {
	return m.files, m.total, m.err
}

type mockDeleter struct{ err error }

func (m *mockDeleter) Execute(ctx context.Context, id string) error { return m.err }

// ---- helpers ----

func newTestAPI(t *testing.T, upload delivery.UploadExecutor, list delivery.ListExecutor, del delivery.DeleteExecutor) huma.API {
	t.Helper()
	_, api := humatest.New(t, huma.DefaultConfig("Test API", "0.0.1"))
	delivery.RegisterRoutes(api, upload, list, del)
	return api
}

// ---- tests ----

func TestUploadMedia_Success(t *testing.T) {
	uploaded := &domain.MediaFile{
		ID:       "file-1",
		Filename: "test.jpg",
		MimeType: "image/jpeg",
		Size:     512,
		URL:      "http://cdn/media/2026/03/abc.jpg",
	}
	api := newTestAPI(t,
		&mockUploader{out: uploaded},
		&mockLister{},
		&mockDeleter{},
	)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, _ := w.CreateFormFile("file", "test.jpg")
	part.Write([]byte("fake-jpeg-data"))
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/media", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()
	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "file-1", result["id"])
}

func TestListMedia_Success(t *testing.T) {
	files := []*domain.MediaFile{
		{ID: "f1", Filename: "a.jpg", URL: "http://cdn/a.jpg"},
	}
	api := newTestAPI(t,
		&mockUploader{},
		&mockLister{files: files, total: 1},
		&mockDeleter{},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/media", nil)
	resp := httptest.NewRecorder()
	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestDeleteMedia_Success(t *testing.T) {
	api := newTestAPI(t,
		&mockUploader{},
		&mockLister{},
		&mockDeleter{},
	)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/media/file-1", nil)
	resp := httptest.NewRecorder()
	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNoContent, resp.Code)
}

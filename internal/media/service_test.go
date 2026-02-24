package media_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/media"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: MediaRepository
// ---------------------------------------------------------------------------

type mockRepo struct {
	listFiles  []model.MediaFile
	listTotal  int64
	listErr    error
	getByID    *model.MediaFile
	getByIDErr error
	createErr  error
	updateErr  error
	deleteErr  error

	batchDeleteCount int64
	batchDeleteErr   error

	referencingPosts    []media.PostRef
	referencingPostsErr error

	batchRefCounts    map[string]int64
	batchRefCountsErr error

	createdFile *model.MediaFile
}

func (m *mockRepo) List(_ context.Context, _ media.ListFilter) ([]model.MediaFile, int64, error) {
	return m.listFiles, m.listTotal, m.listErr
}

func (m *mockRepo) GetByID(_ context.Context, _ string) (*model.MediaFile, error) {
	return m.getByID, m.getByIDErr
}

func (m *mockRepo) Create(_ context.Context, mf *model.MediaFile) error {
	if m.createErr == nil {
		mf.ID = "mf-new-id"
		mf.CreatedAt = time.Now()
		mf.UpdatedAt = time.Now()
		m.createdFile = mf
	}
	return m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ *model.MediaFile) error {
	return m.updateErr
}

func (m *mockRepo) SoftDelete(_ context.Context, _ string) error {
	return m.deleteErr
}

func (m *mockRepo) BatchSoftDelete(_ context.Context, _ []string) (int64, error) {
	return m.batchDeleteCount, m.batchDeleteErr
}

func (m *mockRepo) GetReferencingPosts(_ context.Context, _ string) ([]media.PostRef, error) {
	return m.referencingPosts, m.referencingPostsErr
}

func (m *mockRepo) GetBatchReferencingPosts(_ context.Context, _ []string) (map[string]int64, error) {
	return m.batchRefCounts, m.batchRefCountsErr
}

// ---------------------------------------------------------------------------
// Mock: StorageUploader
// ---------------------------------------------------------------------------

type mockStorage struct {
	uploadErr      error
	uploadBytesErr error
	deleteErr      error
	batchDeleteErr error
	publicURLBase  string

	uploadedKeys []string
}

func (m *mockStorage) Upload(_ context.Context, key string, _ io.Reader, _ string, _ int64) error {
	m.uploadedKeys = append(m.uploadedKeys, key)
	return m.uploadErr
}

func (m *mockStorage) UploadBytes(_ context.Context, key string, _ []byte, _ string) error {
	m.uploadedKeys = append(m.uploadedKeys, key)
	return m.uploadBytesErr
}

func (m *mockStorage) Delete(_ context.Context, _ string) error {
	return m.deleteErr
}

func (m *mockStorage) BatchDelete(_ context.Context, _ []string) error {
	return m.batchDeleteErr
}

func (m *mockStorage) PublicURL(key string) string {
	return m.publicURLBase + "/" + key
}

// ---------------------------------------------------------------------------
// Mock: ImageProcessor
// ---------------------------------------------------------------------------

type mockImaging struct {
	width     int
	height    int
	dimErr    error
	thumbData []byte
	thumbErr  error
}

func (m *mockImaging) ExtractDimensions(_ io.Reader) (int, int, error) {
	return m.width, m.height, m.dimErr
}

func (m *mockImaging) Thumbnail(_ io.Reader, _, _ int, _ string) ([]byte, error) {
	return m.thumbData, m.thumbErr
}

// ---------------------------------------------------------------------------
// Mock: audit.Logger
// ---------------------------------------------------------------------------

type mockAudit struct {
	logged []audit.Entry
	err    error
}

func (m *mockAudit) Log(_ context.Context, entry audit.Entry) error {
	m.logged = append(m.logged, entry)
	return m.err
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

type testEnv struct {
	svc     *media.Service
	repo    *mockRepo
	storage *mockStorage
	imaging *mockImaging
	audit   *mockAudit
}

func newTestEnv() *testEnv {
	r := &mockRepo{}
	st := &mockStorage{publicURLBase: "https://cdn.example.com"}
	img := &mockImaging{
		width:     800,
		height:    600,
		thumbData: []byte("thumb-data"),
	}
	a := &mockAudit{}
	return &testEnv{
		svc:     media.NewService(r, st, img, a),
		repo:    r,
		storage: st,
		imaging: img,
		audit:   a,
	}
}

func testMediaFile() *model.MediaFile {
	return &model.MediaFile{
		ID:           "mf-1",
		UploaderID:   "user-1",
		FileName:     "abc123.jpg",
		OriginalName: "photo.jpg",
		MimeType:     "image/jpeg",
		MediaType:    model.MediaTypeImage,
		FileSize:     1024,
		StoragePath:  "media/2026/02/abc123.jpg",
		PublicURL:    "https://cdn.example.com/media/2026/02/abc123.jpg",
		AltText:      "A photo",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Tests: Upload
// ---------------------------------------------------------------------------

func TestService_Upload_ImageSuccess(t *testing.T) {
	env := newTestEnv()

	resp, err := env.svc.Upload(
		context.Background(),
		"my-site",
		"user-1",
		strings.NewReader("fake-image-data"),
		"photo.jpg",
		"image/jpeg",
		15,
		"A nice photo",
	)
	require.NoError(t, err)
	assert.Equal(t, "mf-new-id", resp.ID)
	assert.Equal(t, "user-1", resp.UploaderID)
	assert.Equal(t, "photo.jpg", resp.OriginalName)
	assert.Equal(t, int(model.MediaTypeImage), resp.MediaType)
	assert.Equal(t, "A nice photo", resp.AltText)
	assert.NotNil(t, resp.Width)
	assert.NotNil(t, resp.Height)
	assert.Equal(t, 800, *resp.Width)
	assert.Equal(t, 600, *resp.Height)

	// Should have uploaded: sm thumb + md thumb + original = 3 keys.
	assert.Len(t, env.storage.uploadedKeys, 3)

	// Thumbnails should be in the response.
	var thumbs map[string]string
	require.NoError(t, json.Unmarshal(resp.ThumbnailURLs, &thumbs))
	assert.Contains(t, thumbs, "sm")
	assert.Contains(t, thumbs, "md")

	// Audit log.
	require.Len(t, env.audit.logged, 1)
	assert.Equal(t, model.LogActionCreate, env.audit.logged[0].Action)
	assert.Equal(t, "media", env.audit.logged[0].ResourceType)
}

func TestService_Upload_NonImageSuccess(t *testing.T) {
	env := newTestEnv()

	resp, err := env.svc.Upload(
		context.Background(),
		"my-site",
		"user-1",
		strings.NewReader("fake-video-data"),
		"video.mp4",
		"video/mp4",
		2048,
		"",
	)
	require.NoError(t, err)
	assert.Equal(t, int(model.MediaTypeVideo), resp.MediaType)
	assert.Nil(t, resp.Width)
	assert.Nil(t, resp.Height)

	// Only original file uploaded, no thumbnails.
	assert.Len(t, env.storage.uploadedKeys, 1)
}

func TestService_Upload_InvalidMIME(t *testing.T) {
	env := newTestEnv()

	_, err := env.svc.Upload(
		context.Background(),
		"my-site",
		"user-1",
		strings.NewReader("data"),
		"malware.exe",
		"application/x-msdownload",
		100,
		"",
	)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestService_Upload_SVGSkipsImaging(t *testing.T) {
	env := newTestEnv()

	resp, err := env.svc.Upload(
		context.Background(),
		"my-site",
		"user-1",
		strings.NewReader("<svg></svg>"),
		"icon.svg",
		"image/svg+xml",
		11,
		"",
	)
	require.NoError(t, err)
	assert.Equal(t, int(model.MediaTypeImage), resp.MediaType)
	assert.Nil(t, resp.Width)  // SVG skips dimension extraction
	assert.Nil(t, resp.Height) // SVG skips dimension extraction

	// Only original file uploaded.
	assert.Len(t, env.storage.uploadedKeys, 1)
}

// ---------------------------------------------------------------------------
// Tests: Delete
// ---------------------------------------------------------------------------

func TestService_DeleteMedia_NoRefs(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testMediaFile()
	env.repo.referencingPosts = nil

	err := env.svc.DeleteMedia(context.Background(), "mf-1", false)
	require.NoError(t, err)

	require.Len(t, env.audit.logged, 1)
	assert.Equal(t, model.LogActionDelete, env.audit.logged[0].Action)
}

func TestService_DeleteMedia_HasRefs_NoForce(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testMediaFile()
	env.repo.referencingPosts = []media.PostRef{
		{ID: "post-1", Title: "My Post"},
	}

	err := env.svc.DeleteMedia(context.Background(), "mf-1", false)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestService_DeleteMedia_HasRefs_Force(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testMediaFile()
	env.repo.referencingPosts = []media.PostRef{
		{ID: "post-1", Title: "My Post"},
	}

	err := env.svc.DeleteMedia(context.Background(), "mf-1", true)
	require.NoError(t, err)

	require.Len(t, env.audit.logged, 1)
	assert.Equal(t, model.LogActionDelete, env.audit.logged[0].Action)
}

func TestService_DeleteMedia_NotFound(t *testing.T) {
	env := newTestEnv()
	env.repo.getByIDErr = apperror.NotFound("media file not found", nil)

	err := env.svc.DeleteMedia(context.Background(), "nonexistent", false)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: BatchDelete
// ---------------------------------------------------------------------------

func TestService_BatchDelete_PartialSkip(t *testing.T) {
	env := newTestEnv()
	env.repo.batchRefCounts = map[string]int64{
		"mf-2": 3, // referenced
	}
	env.repo.batchDeleteCount = 2

	resp, err := env.svc.BatchDeleteMedia(context.Background(), []string{"mf-1", "mf-2", "mf-3"}, false)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.DeletedCount)
	require.Len(t, resp.Skipped, 1)
	assert.Equal(t, "mf-2", resp.Skipped[0].ID)
	assert.Contains(t, resp.Skipped[0].Reason, "3 post(s)")
}

func TestService_BatchDelete_AllDeleted(t *testing.T) {
	env := newTestEnv()
	env.repo.batchRefCounts = map[string]int64{}
	env.repo.batchDeleteCount = 2

	resp, err := env.svc.BatchDeleteMedia(context.Background(), []string{"mf-1", "mf-2"}, false)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.DeletedCount)
	assert.Empty(t, resp.Skipped)
}

func TestService_BatchDelete_ForceDeleteAll(t *testing.T) {
	env := newTestEnv()
	env.repo.batchRefCounts = map[string]int64{
		"mf-1": 5,
	}
	env.repo.batchDeleteCount = 2

	resp, err := env.svc.BatchDeleteMedia(context.Background(), []string{"mf-1", "mf-2"}, true)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.DeletedCount)
	assert.Empty(t, resp.Skipped) // force=true skips nothing
}

// ---------------------------------------------------------------------------
// Tests: List
// ---------------------------------------------------------------------------

func TestService_List_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.listFiles = []model.MediaFile{*testMediaFile()}
	env.repo.listTotal = 1

	items, total, err := env.svc.List(context.Background(), media.ListFilter{
		Page: 1, PerPage: 20,
	})
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(1), total)
}

func TestService_List_Empty(t *testing.T) {
	env := newTestEnv()
	env.repo.listFiles = nil
	env.repo.listTotal = 0

	items, total, err := env.svc.List(context.Background(), media.ListFilter{
		Page: 1, PerPage: 20,
	})
	require.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, int64(0), total)
}

// ---------------------------------------------------------------------------
// Tests: UpdateMedia
// ---------------------------------------------------------------------------

func TestService_UpdateMedia_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testMediaFile()

	newAlt := "Updated alt text"
	resp, err := env.svc.UpdateMedia(context.Background(), "mf-1", &media.UpdateMediaReq{
		AltText: &newAlt,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated alt text", resp.AltText)

	require.Len(t, env.audit.logged, 1)
	assert.Equal(t, model.LogActionUpdate, env.audit.logged[0].Action)
}

func TestService_UpdateMedia_NotFound(t *testing.T) {
	env := newTestEnv()
	env.repo.getByIDErr = apperror.NotFound("media file not found", nil)

	newAlt := "X"
	_, err := env.svc.UpdateMedia(context.Background(), "nonexistent", &media.UpdateMediaReq{
		AltText: &newAlt,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

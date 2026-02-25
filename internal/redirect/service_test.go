package redirect_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/cache"
	"github.com/sky-flux/cms/internal/redirect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: RedirectRepository
// ---------------------------------------------------------------------------

type mockRepo struct {
	listRedirects []model.Redirect
	listTotal     int64
	listErr       error

	getByID    *model.Redirect
	getByIDErr error

	createErr error
	updateErr error
	deleteErr error

	batchDeleteCount int64
	batchDeleteErr   error

	sourcePathExists    bool
	sourcePathExistsErr error

	bulkInsertCount int64
	bulkInsertErr   error
	bulkInserted    []*model.Redirect

	listAllRedirects []model.Redirect
	listAllErr       error
}

func (m *mockRepo) List(_ context.Context, _ redirect.ListFilter) ([]model.Redirect, int64, error) {
	return m.listRedirects, m.listTotal, m.listErr
}

func (m *mockRepo) GetByID(_ context.Context, _ string) (*model.Redirect, error) {
	return m.getByID, m.getByIDErr
}

func (m *mockRepo) Create(_ context.Context, rd *model.Redirect) error {
	if m.createErr == nil {
		rd.ID = "new-redirect-id"
		rd.CreatedAt = time.Now()
		rd.UpdatedAt = time.Now()
	}
	return m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ *model.Redirect) error {
	return m.updateErr
}

func (m *mockRepo) Delete(_ context.Context, _ string) error {
	return m.deleteErr
}

func (m *mockRepo) BatchDelete(_ context.Context, _ []string) (int64, error) {
	return m.batchDeleteCount, m.batchDeleteErr
}

func (m *mockRepo) SourcePathExists(_ context.Context, _ string, _ string) (bool, error) {
	return m.sourcePathExists, m.sourcePathExistsErr
}

func (m *mockRepo) BulkInsert(_ context.Context, redirects []*model.Redirect) (int64, error) {
	m.bulkInserted = redirects
	if m.bulkInsertErr != nil {
		return 0, m.bulkInsertErr
	}
	if m.bulkInsertCount > 0 {
		return m.bulkInsertCount, nil
	}
	return int64(len(redirects)), nil
}

func (m *mockRepo) ListAll(_ context.Context) ([]model.Redirect, error) {
	return m.listAllRedirects, m.listAllErr
}

// ---------------------------------------------------------------------------
// Mock: AuditLogger
// ---------------------------------------------------------------------------

type mockAudit struct {
	lastEntry *audit.Entry
}

func (m *mockAudit) Log(_ context.Context, entry audit.Entry) error {
	m.lastEntry = &entry
	return nil
}

// ---------------------------------------------------------------------------
// Test environment
// ---------------------------------------------------------------------------

type testEnv struct {
	svc   *redirect.Service
	repo  *mockRepo
	audit *mockAudit
}

func newTestEnv() *testEnv {
	r := &mockRepo{}
	a := &mockAudit{}
	cc := cache.NewClient(nil)
	return &testEnv{
		svc:   redirect.NewService(r, a, cc),
		repo:  r,
		audit: a,
	}
}

func testRedirect() *model.Redirect {
	return &model.Redirect{
		ID:         "rd-1",
		SourcePath: "/old-page",
		TargetURL:  "https://example.com/new-page",
		StatusCode: 301,
		Status:     model.RedirectStatusActive,
		HitCount:   10,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestCreate_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.sourcePathExists = false

	resp, err := env.svc.Create(context.Background(), "test-site", &redirect.CreateRedirectReq{
		SourcePath: "/old-page",
		TargetURL:  "https://example.com/new-page",
	}, "user-1")

	require.NoError(t, err)
	assert.Equal(t, "/old-page", resp.SourcePath)
	assert.Equal(t, "https://example.com/new-page", resp.TargetURL)
	assert.Equal(t, 301, resp.StatusCode)
	assert.True(t, resp.IsActive)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionCreate, env.audit.lastEntry.Action)
	assert.Equal(t, "redirect", env.audit.lastEntry.ResourceType)
}

func TestCreate_InvalidPath(t *testing.T) {
	env := newTestEnv()

	_, err := env.svc.Create(context.Background(), "test-site", &redirect.CreateRedirectReq{
		SourcePath: "no-leading-slash",
		TargetURL:  "https://example.com",
	}, "user-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must start with /")
}

func TestCreate_PathWithQueryString(t *testing.T) {
	env := newTestEnv()

	_, err := env.svc.Create(context.Background(), "test-site", &redirect.CreateRedirectReq{
		SourcePath: "/old?foo=bar",
		TargetURL:  "https://example.com",
	}, "user-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "query string")
}

func TestCreate_DuplicatePath(t *testing.T) {
	env := newTestEnv()
	env.repo.sourcePathExists = true

	_, err := env.svc.Create(context.Background(), "test-site", &redirect.CreateRedirectReq{
		SourcePath: "/old-page",
		TargetURL:  "https://example.com",
	}, "user-1")

	require.Error(t, err)
	assert.True(t, apperror.HTTPStatusCode(err) == 409)
}

func TestCreate_StripsTrailingSlash(t *testing.T) {
	env := newTestEnv()
	env.repo.sourcePathExists = false

	resp, err := env.svc.Create(context.Background(), "test-site", &redirect.CreateRedirectReq{
		SourcePath: "/old-page/",
		TargetURL:  "https://example.com",
	}, "user-1")

	require.NoError(t, err)
	assert.Equal(t, "/old-page", resp.SourcePath)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestUpdate_ToggleStatus(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testRedirect()

	isActive := false
	resp, err := env.svc.Update(context.Background(), "test-site", "rd-1", &redirect.UpdateRedirectReq{
		IsActive: &isActive,
	})

	require.NoError(t, err)
	assert.False(t, resp.IsActive)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionUpdate, env.audit.lastEntry.Action)
}

func TestUpdate_SourcePathConflict(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testRedirect()
	env.repo.sourcePathExists = true

	newPath := "/taken-path"
	_, err := env.svc.Update(context.Background(), "test-site", "rd-1", &redirect.UpdateRedirectReq{
		SourcePath: &newPath,
	})

	require.Error(t, err)
	assert.True(t, apperror.HTTPStatusCode(err) == 409)
}

// ---------------------------------------------------------------------------
// Tests: Delete
// ---------------------------------------------------------------------------

func TestDelete_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testRedirect()

	err := env.svc.Delete(context.Background(), "test-site", "rd-1")
	require.NoError(t, err)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionDelete, env.audit.lastEntry.Action)
}

func TestDelete_NotFound(t *testing.T) {
	env := newTestEnv()
	env.repo.getByIDErr = apperror.NotFound("redirect not found", nil)

	err := env.svc.Delete(context.Background(), "test-site", "nonexistent")
	require.Error(t, err)
	assert.True(t, apperror.HTTPStatusCode(err) == 404)
}

// ---------------------------------------------------------------------------
// Tests: BatchDelete
// ---------------------------------------------------------------------------

func TestBatchDelete_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.batchDeleteCount = 3

	count, err := env.svc.BatchDelete(context.Background(), "test-site", []string{"rd-1", "rd-2", "rd-3"})
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.NotNil(t, env.audit.lastEntry)
}

// ---------------------------------------------------------------------------
// Tests: Import
// ---------------------------------------------------------------------------

func TestImport_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.sourcePathExists = false

	csvData := "source_path,target_url,status_code\n/old,https://new.com,302\n/foo,https://bar.com,301\n"
	result, err := env.svc.Import(context.Background(), "test-site", strings.NewReader(csvData), "user-1")

	require.NoError(t, err)
	assert.Equal(t, 2, result.Imported)
	assert.Equal(t, 0, result.Skipped)
	assert.Empty(t, result.Errors)
}

func TestImport_SkipDuplicates(t *testing.T) {
	env := newTestEnv()
	env.repo.sourcePathExists = true

	csvData := "source_path,target_url\n/old,https://new.com\n/foo,https://bar.com\n"
	result, err := env.svc.Import(context.Background(), "test-site", strings.NewReader(csvData), "user-1")

	require.NoError(t, err)
	assert.Equal(t, 0, result.Imported)
	assert.Equal(t, 2, result.Skipped)
}

func TestImport_InvalidHeader(t *testing.T) {
	env := newTestEnv()

	csvData := "wrong_column,target_url\n/old,https://new.com\n"
	_, err := env.svc.Import(context.Background(), "test-site", strings.NewReader(csvData), "user-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CSV header")
}

func TestImport_InvalidPath(t *testing.T) {
	env := newTestEnv()

	csvData := "source_path,target_url\nno-slash,https://new.com\n"
	result, err := env.svc.Import(context.Background(), "test-site", strings.NewReader(csvData), "user-1")

	require.NoError(t, err)
	assert.Equal(t, 0, result.Imported)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "must start with /")
}

// ---------------------------------------------------------------------------
// Tests: Export
// ---------------------------------------------------------------------------

func TestExport_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.listAllRedirects = []model.Redirect{
		{ID: "rd-1", SourcePath: "/old", TargetURL: "https://new.com", StatusCode: 301},
		{ID: "rd-2", SourcePath: "/foo", TargetURL: "https://bar.com", StatusCode: 302},
	}

	redirects, err := env.svc.Export(context.Background())
	require.NoError(t, err)
	assert.Len(t, redirects, 2)
}

// ---------------------------------------------------------------------------
// Tests: List
// ---------------------------------------------------------------------------

func TestList_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.listRedirects = []model.Redirect{*testRedirect()}
	env.repo.listTotal = 1

	results, total, err := env.svc.List(context.Background(), redirect.ListFilter{Page: 1, PerPage: 20})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, int64(1), total)
	assert.True(t, results[0].IsActive)
}

func TestList_DefaultsPagination(t *testing.T) {
	env := newTestEnv()
	env.repo.listRedirects = nil
	env.repo.listTotal = 0

	results, total, err := env.svc.List(context.Background(), redirect.ListFilter{Page: 0, PerPage: 0})
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, int64(0), total)
}

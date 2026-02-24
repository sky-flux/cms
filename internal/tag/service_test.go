package tag_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/cache"
	"github.com/sky-flux/cms/internal/pkg/search"
	"github.com/sky-flux/cms/internal/tag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: TagRepository
// ---------------------------------------------------------------------------

type mockRepo struct {
	listTags   []model.Tag
	listTotal  int64
	listErr    error
	getByID    *model.Tag
	getByIDErr error
	createErr  error
	updateErr  error
	deleteErr  error
	slugExists    bool
	slugExistsErr error
	nameExists    bool
	nameExistsErr error
	postCount     int64
	postCountErr  error
}

func (m *mockRepo) List(_ context.Context, _ tag.ListFilter) ([]model.Tag, int64, error) {
	return m.listTags, m.listTotal, m.listErr
}
func (m *mockRepo) GetByID(_ context.Context, _ string) (*model.Tag, error) {
	return m.getByID, m.getByIDErr
}
func (m *mockRepo) Create(_ context.Context, t *model.Tag) error {
	if m.createErr == nil {
		t.ID = "new-tag-id"
		t.CreatedAt = time.Now()
	}
	return m.createErr
}
func (m *mockRepo) Update(_ context.Context, _ *model.Tag) error { return m.updateErr }
func (m *mockRepo) Delete(_ context.Context, _ string) error     { return m.deleteErr }
func (m *mockRepo) SlugExists(_ context.Context, _ string, _ string) (bool, error) {
	return m.slugExists, m.slugExistsErr
}
func (m *mockRepo) NameExists(_ context.Context, _ string, _ string) (bool, error) {
	return m.nameExists, m.nameExistsErr
}
func (m *mockRepo) CountPosts(_ context.Context, _ string) (int64, error) {
	return m.postCount, m.postCountErr
}

// ---------------------------------------------------------------------------
// Mock: AuditLogger
// ---------------------------------------------------------------------------

type mockAudit struct {
	lastEntry *audit.Entry
	err       error
}

func (m *mockAudit) Log(_ context.Context, entry audit.Entry) error {
	m.lastEntry = &entry
	return m.err
}

// ---------------------------------------------------------------------------
// Test environment
// ---------------------------------------------------------------------------

type testEnv struct {
	svc   *tag.Service
	repo  *mockRepo
	audit *mockAudit
}

func newTestEnv() *testEnv {
	r := &mockRepo{}
	a := &mockAudit{}
	sc := search.NewClient(nil)
	cc := cache.NewClient(nil)
	return &testEnv{
		svc:   tag.NewService(r, sc, cc, a),
		repo:  r,
		audit: a,
	}
}

func testTag() *model.Tag {
	return &model.Tag{
		ID:        "tag-1",
		Name:      "Go",
		Slug:      "go",
		CreatedAt: time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Tests: CreateTag
// ---------------------------------------------------------------------------

func TestCreateTag_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.nameExists = false
	env.repo.slugExists = false

	resp, err := env.svc.CreateTag(context.Background(), "test-site", &tag.CreateTagReq{
		Name: "Rust",
		Slug: "rust",
	})
	require.NoError(t, err)
	assert.Equal(t, "rust", resp.Slug)
	assert.Equal(t, "Rust", resp.Name)
	assert.Equal(t, int64(0), resp.PostCount)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionCreate, env.audit.lastEntry.Action)
	assert.Equal(t, "tag", env.audit.lastEntry.ResourceType)
}

func TestCreateTag_DuplicateName(t *testing.T) {
	env := newTestEnv()
	env.repo.nameExists = true

	_, err := env.svc.CreateTag(context.Background(), "test-site", &tag.CreateTagReq{
		Name: "Go",
		Slug: "go-lang",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestCreateTag_DuplicateSlug(t *testing.T) {
	env := newTestEnv()
	env.repo.nameExists = false
	env.repo.slugExists = true

	_, err := env.svc.CreateTag(context.Background(), "test-site", &tag.CreateTagReq{
		Name: "Golang",
		Slug: "go",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

// ---------------------------------------------------------------------------
// Tests: UpdateTag
// ---------------------------------------------------------------------------

func TestUpdateTag_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testTag()
	env.repo.nameExists = false

	newName := "GoLang"
	resp, err := env.svc.UpdateTag(context.Background(), "test-site", "tag-1", &tag.UpdateTagReq{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "GoLang", resp.Name)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionUpdate, env.audit.lastEntry.Action)
}

func TestUpdateTag_NameConflict(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testTag()
	env.repo.nameExists = true

	newName := "Rust"
	_, err := env.svc.UpdateTag(context.Background(), "test-site", "tag-1", &tag.UpdateTagReq{
		Name: &newName,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestUpdateTag_SlugConflict(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testTag()
	env.repo.slugExists = true

	newSlug := "rust"
	_, err := env.svc.UpdateTag(context.Background(), "test-site", "tag-1", &tag.UpdateTagReq{
		Slug: &newSlug,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

// ---------------------------------------------------------------------------
// Tests: DeleteTag
// ---------------------------------------------------------------------------

func TestDeleteTag_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testTag()

	err := env.svc.DeleteTag(context.Background(), "test-site", "tag-1")
	require.NoError(t, err)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionDelete, env.audit.lastEntry.Action)
}

func TestDeleteTag_NotFound(t *testing.T) {
	env := newTestEnv()
	env.repo.getByIDErr = apperror.NotFound("tag not found", nil)

	err := env.svc.DeleteTag(context.Background(), "test-site", "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: ListTags
// ---------------------------------------------------------------------------

func TestListTags_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.listTags = []model.Tag{*testTag()}
	env.repo.listTotal = 1
	env.repo.postCount = 5

	tags, total, err := env.svc.List(context.Background(), tag.ListFilter{Page: 1, PerPage: 10})
	require.NoError(t, err)
	assert.Len(t, tags, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(5), tags[0].PostCount)
}

func TestListTags_DefaultsPagination(t *testing.T) {
	env := newTestEnv()
	env.repo.listTags = nil
	env.repo.listTotal = 0

	tags, total, err := env.svc.List(context.Background(), tag.ListFilter{Page: 0, PerPage: 0})
	require.NoError(t, err)
	assert.Empty(t, tags)
	assert.Equal(t, int64(0), total)
}

// ---------------------------------------------------------------------------
// Tests: GetTag
// ---------------------------------------------------------------------------

func TestGetTag_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testTag()
	env.repo.postCount = 3

	resp, err := env.svc.GetTag(context.Background(), "tag-1")
	require.NoError(t, err)
	assert.Equal(t, "Go", resp.Name)
	assert.Equal(t, int64(3), resp.PostCount)
}

func TestGetTag_NotFound(t *testing.T) {
	env := newTestEnv()
	env.repo.getByIDErr = apperror.NotFound("tag not found", nil)

	_, err := env.svc.GetTag(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: Suggest
// ---------------------------------------------------------------------------

func TestSuggest_NilSearch(t *testing.T) {
	env := newTestEnv()

	// search.NewClient(nil) returns a client with nil ms, Search returns empty result.
	tags, err := env.svc.Suggest(context.Background(), "test-site", "go")
	require.NoError(t, err)
	assert.Empty(t, tags)
}

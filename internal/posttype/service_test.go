package posttype_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/posttype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: PostTypeRepository
// ---------------------------------------------------------------------------

type mockRepo struct {
	listPTs    []model.PostType
	listErr    error
	getByID    *model.PostType
	getByIDErr error
	getBySlug    *model.PostType
	getBySlugErr error
	createErr  error
	updateErr  error
	deleteErr  error
}

func (m *mockRepo) List(_ context.Context) ([]model.PostType, error) {
	return m.listPTs, m.listErr
}
func (m *mockRepo) GetByID(_ context.Context, _ string) (*model.PostType, error) {
	return m.getByID, m.getByIDErr
}
func (m *mockRepo) GetBySlug(_ context.Context, _ string) (*model.PostType, error) {
	return m.getBySlug, m.getBySlugErr
}
func (m *mockRepo) Create(_ context.Context, pt *model.PostType) error {
	if m.createErr == nil {
		pt.ID = "new-pt-id"
		pt.CreatedAt = time.Now()
		pt.UpdatedAt = time.Now()
	}
	return m.createErr
}
func (m *mockRepo) Update(_ context.Context, _ *model.PostType) error { return m.updateErr }
func (m *mockRepo) Delete(_ context.Context, _ string) error          { return m.deleteErr }

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
// Helper
// ---------------------------------------------------------------------------

type testEnv struct {
	svc   *posttype.Service
	repo  *mockRepo
	audit *mockAudit
}

func newTestEnv() *testEnv {
	r := &mockRepo{}
	a := &mockAudit{}
	return &testEnv{
		svc:   posttype.NewService(r, a),
		repo:  r,
		audit: a,
	}
}

func testPostType() *model.PostType {
	return &model.PostType{
		ID:          "pt-1",
		Name:        "Article",
		Slug:        "article",
		Description: "Standard article type",
		Fields:      json.RawMessage(`[{"name":"body","type":"richtext"}]`),
		BuiltIn:     model.ToggleNo,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func builtInPostType() *model.PostType {
	pt := testPostType()
	pt.ID = "pt-builtin"
	pt.Slug = "post"
	pt.Name = "Post"
	pt.BuiltIn = model.ToggleYes
	return pt
}

// ---------------------------------------------------------------------------
// Tests: List
// ---------------------------------------------------------------------------

func TestService_List_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.listPTs = []model.PostType{*testPostType()}

	pts, err := env.svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, pts, 1)
}

func TestService_List_Empty(t *testing.T) {
	env := newTestEnv()
	env.repo.listPTs = nil

	pts, err := env.svc.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, pts)
}

func TestService_List_RepoError(t *testing.T) {
	env := newTestEnv()
	env.repo.listErr = errors.New("db error")

	_, err := env.svc.List(context.Background())
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestService_Create_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.getBySlugErr = apperror.NotFound("not found", nil)

	pt, err := env.svc.Create(context.Background(), &posttype.CreatePostTypeReq{
		Name:   "Gallery",
		Slug:   "gallery",
		Fields: json.RawMessage(`[{"name":"images","type":"gallery"}]`),
	})
	require.NoError(t, err)
	assert.Equal(t, "gallery", pt.Slug)
	assert.Equal(t, model.ToggleNo, pt.BuiltIn)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionCreate, env.audit.lastEntry.Action)
}

func TestService_Create_DefaultEmptyFields(t *testing.T) {
	env := newTestEnv()
	env.repo.getBySlugErr = apperror.NotFound("not found", nil)

	pt, err := env.svc.Create(context.Background(), &posttype.CreatePostTypeReq{
		Name: "Empty",
		Slug: "empty",
	})
	require.NoError(t, err)
	assert.Equal(t, json.RawMessage("[]"), pt.Fields)
}

func TestService_Create_InvalidSlug(t *testing.T) {
	env := newTestEnv()

	_, err := env.svc.Create(context.Background(), &posttype.CreatePostTypeReq{
		Name: "Bad", Slug: "NO CAPS!",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestService_Create_InvalidFields(t *testing.T) {
	env := newTestEnv()

	_, err := env.svc.Create(context.Background(), &posttype.CreatePostTypeReq{
		Name:   "Bad",
		Slug:   "bad_fields",
		Fields: json.RawMessage(`{"not":"an_array"}`),
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestService_Create_SlugConflict(t *testing.T) {
	env := newTestEnv()
	env.repo.getBySlug = testPostType()

	_, err := env.svc.Create(context.Background(), &posttype.CreatePostTypeReq{
		Name: "Dup", Slug: "article",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestService_Create_RepoError(t *testing.T) {
	env := newTestEnv()
	env.repo.getBySlugErr = apperror.NotFound("not found", nil)
	env.repo.createErr = errors.New("db error")

	_, err := env.svc.Create(context.Background(), &posttype.CreatePostTypeReq{
		Name: "Fail", Slug: "fail",
	})
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestService_Update_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testPostType()

	newName := "Updated Article"
	pt, err := env.svc.Update(context.Background(), "pt-1", &posttype.UpdatePostTypeReq{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Article", pt.Name)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionUpdate, env.audit.lastEntry.Action)
}

func TestService_Update_ChangeSlug(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testPostType()
	env.repo.getBySlugErr = apperror.NotFound("not found", nil)

	newSlug := "new-slug"
	pt, err := env.svc.Update(context.Background(), "pt-1", &posttype.UpdatePostTypeReq{
		Slug: &newSlug,
	})
	require.NoError(t, err)
	assert.Equal(t, "new-slug", pt.Slug)
}

func TestService_Update_SlugConflict(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testPostType()
	env.repo.getBySlug = &model.PostType{ID: "other-pt", Slug: "taken"}

	newSlug := "taken"
	_, err := env.svc.Update(context.Background(), "pt-1", &posttype.UpdatePostTypeReq{
		Slug: &newSlug,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestService_Update_BuiltInCannotChangeSlug(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = builtInPostType()

	newSlug := "new-slug"
	_, err := env.svc.Update(context.Background(), "pt-builtin", &posttype.UpdatePostTypeReq{
		Slug: &newSlug,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestService_Update_BuiltInCanChangeName(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = builtInPostType()

	newName := "Blog Post"
	pt, err := env.svc.Update(context.Background(), "pt-builtin", &posttype.UpdatePostTypeReq{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "Blog Post", pt.Name)
}

func TestService_Update_NotFound(t *testing.T) {
	env := newTestEnv()
	env.repo.getByIDErr = apperror.NotFound("post type not found", nil)

	newName := "X"
	_, err := env.svc.Update(context.Background(), "nonexistent", &posttype.UpdatePostTypeReq{
		Name: &newName,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

func TestService_Update_InvalidFields(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testPostType()

	badFields := json.RawMessage(`"not-an-array"`)
	_, err := env.svc.Update(context.Background(), "pt-1", &posttype.UpdatePostTypeReq{
		Fields: &badFields,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestService_Update_InvalidSlugFormat(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testPostType()

	badSlug := "INVALID SLUG!"
	_, err := env.svc.Update(context.Background(), "pt-1", &posttype.UpdatePostTypeReq{
		Slug: &badSlug,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

// ---------------------------------------------------------------------------
// Tests: Delete
// ---------------------------------------------------------------------------

func TestService_Delete_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testPostType()

	err := env.svc.Delete(context.Background(), "pt-1")
	require.NoError(t, err)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionDelete, env.audit.lastEntry.Action)
}

func TestService_Delete_BuiltIn(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = builtInPostType()

	err := env.svc.Delete(context.Background(), "pt-builtin")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestService_Delete_NotFound(t *testing.T) {
	env := newTestEnv()
	env.repo.getByIDErr = apperror.NotFound("post type not found", nil)

	err := env.svc.Delete(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

func TestService_Delete_RepoError(t *testing.T) {
	env := newTestEnv()
	env.repo.getByID = testPostType()
	env.repo.deleteErr = errors.New("db error")

	err := env.svc.Delete(context.Background(), "pt-1")
	require.Error(t, err)
}

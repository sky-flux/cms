package post_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/search"
	"github.com/sky-flux/cms/internal/post"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: PostRepository
// ---------------------------------------------------------------------------

type mockPostRepo struct {
	listPosts   []model.Post
	listTotal   int64
	listErr     error
	getByID     *model.Post
	getByIDErr  error
	createErr   error
	updateErr   error
	softDelErr  error
	restoreErr  error
	slugExists  bool
	slugExErr   error
	statusErr   error
	syncCatErr  error
	syncTagErr  error
	loadRelErr  error

	// capture calls
	lastUpdateVersion int
	lastStatus        model.PostStatus
}

func (m *mockPostRepo) List(_ context.Context, _ post.ListFilter) ([]model.Post, int64, error) {
	return m.listPosts, m.listTotal, m.listErr
}
func (m *mockPostRepo) GetByID(_ context.Context, _ string) (*model.Post, error) {
	return m.getByID, m.getByIDErr
}
func (m *mockPostRepo) GetByIDUnscoped(_ context.Context, _ string) (*model.Post, error) {
	return m.getByID, m.getByIDErr
}
func (m *mockPostRepo) Create(_ context.Context, p *model.Post) error {
	if m.createErr == nil {
		p.ID = "post-new-id"
		p.Version = 1
		p.CreatedAt = time.Now()
	}
	return m.createErr
}
func (m *mockPostRepo) Update(_ context.Context, _ *model.Post, ver int) error {
	m.lastUpdateVersion = ver
	return m.updateErr
}
func (m *mockPostRepo) SoftDelete(_ context.Context, _ string) error { return m.softDelErr }
func (m *mockPostRepo) Restore(_ context.Context, _ string) error    { return m.restoreErr }
func (m *mockPostRepo) SlugExists(_ context.Context, _ string, _ string) (bool, error) {
	return m.slugExists, m.slugExErr
}
func (m *mockPostRepo) UpdateStatus(_ context.Context, _ string, s model.PostStatus) error {
	m.lastStatus = s
	return m.statusErr
}
func (m *mockPostRepo) SyncCategories(_ context.Context, _ string, _ []string, _ string) error {
	return m.syncCatErr
}
func (m *mockPostRepo) SyncTags(_ context.Context, _ string, _ []string) error {
	return m.syncTagErr
}
func (m *mockPostRepo) LoadRelations(_ context.Context, _ *model.Post) error {
	return m.loadRelErr
}

// ---------------------------------------------------------------------------
// Mock: RevisionRepository
// ---------------------------------------------------------------------------

type mockRevisionRepo struct {
	listRevs []model.PostRevision
	listErr  error
	getByID  *model.PostRevision
	getErr   error
	createErr error
}

func (m *mockRevisionRepo) List(_ context.Context, _ string) ([]model.PostRevision, error) {
	return m.listRevs, m.listErr
}
func (m *mockRevisionRepo) GetByID(_ context.Context, _ string) (*model.PostRevision, error) {
	return m.getByID, m.getErr
}
func (m *mockRevisionRepo) Create(_ context.Context, r *model.PostRevision) error {
	if m.createErr == nil {
		r.ID = "rev-new-id"
		r.CreatedAt = time.Now()
	}
	return m.createErr
}

// ---------------------------------------------------------------------------
// Mock: TranslationRepository
// ---------------------------------------------------------------------------

type mockTransRepo struct {
	listTrans []model.PostTranslation
	listErr   error
	getTrans  *model.PostTranslation
	getErr    error
	upsertErr error
	deleteErr error
}

func (m *mockTransRepo) List(_ context.Context, _ string) ([]model.PostTranslation, error) {
	return m.listTrans, m.listErr
}
func (m *mockTransRepo) Get(_ context.Context, _, _ string) (*model.PostTranslation, error) {
	return m.getTrans, m.getErr
}
func (m *mockTransRepo) Upsert(_ context.Context, _ *model.PostTranslation) error {
	return m.upsertErr
}
func (m *mockTransRepo) Delete(_ context.Context, _, _ string) error { return m.deleteErr }

// ---------------------------------------------------------------------------
// Mock: PreviewTokenRepository
// ---------------------------------------------------------------------------

type mockPreviewRepo struct {
	listTokens  []model.PreviewToken
	listErr     error
	createErr   error
	countActive int
	countErr    error
	delAllCount int64
	delAllErr   error
	delByIDErr  error
	getByHash   *model.PreviewToken
	getHashErr  error
}

func (m *mockPreviewRepo) List(_ context.Context, _ string) ([]model.PreviewToken, error) {
	return m.listTokens, m.listErr
}
func (m *mockPreviewRepo) Create(_ context.Context, t *model.PreviewToken) error {
	if m.createErr == nil {
		t.ID = "token-new-id"
		t.CreatedAt = time.Now()
	}
	return m.createErr
}
func (m *mockPreviewRepo) CountActive(_ context.Context, _ string) (int, error) {
	return m.countActive, m.countErr
}
func (m *mockPreviewRepo) DeleteAll(_ context.Context, _ string) (int64, error) {
	return m.delAllCount, m.delAllErr
}
func (m *mockPreviewRepo) DeleteByID(_ context.Context, _ string) error { return m.delByIDErr }
func (m *mockPreviewRepo) GetByHash(_ context.Context, _ string) (*model.PreviewToken, error) {
	return m.getByHash, m.getHashErr
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
	svc     *post.Service
	posts   *mockPostRepo
	revs    *mockRevisionRepo
	trans   *mockTransRepo
	preview *mockPreviewRepo
	audit   *mockAudit
}

func newTestEnv() *testEnv {
	p := &mockPostRepo{}
	r := &mockRevisionRepo{}
	tr := &mockTransRepo{}
	pr := &mockPreviewRepo{}
	a := &mockAudit{}
	sc := search.NewClient(nil)
	return &testEnv{
		svc:     post.NewService(p, r, tr, pr, sc, a),
		posts:   p,
		revs:    r,
		trans:   tr,
		preview: pr,
		audit:   a,
	}
}

func testPost() *model.Post {
	return &model.Post{
		ID:       "post-1",
		AuthorID: "user-1",
		Title:    "Test Post",
		Slug:     "test-post",
		Status:   model.PostStatusDraft,
		Version:  1,
	}
}

// ---------------------------------------------------------------------------
// Tests: CreatePost
// ---------------------------------------------------------------------------

func TestCreatePost_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.slugExists = false

	p, err := env.svc.CreatePost(context.Background(), "site1", "user-1", &post.CreatePostReq{
		Title:   "My Post",
		Content: "body",
	})
	require.NoError(t, err)
	assert.Equal(t, "post-new-id", p.ID)
	assert.Equal(t, "My Post", p.Title)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionCreate, env.audit.lastEntry.Action)
}

func TestCreatePost_DuplicateSlug(t *testing.T) {
	env := newTestEnv()
	env.posts.slugExists = true

	_, err := env.svc.CreatePost(context.Background(), "site1", "user-1", &post.CreatePostReq{
		Title: "My Post",
		Slug:  "existing-slug",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestCreatePost_ScheduledWithoutDate(t *testing.T) {
	env := newTestEnv()
	env.posts.slugExists = false

	_, err := env.svc.CreatePost(context.Background(), "site1", "user-1", &post.CreatePostReq{
		Title:  "My Post",
		Status: "scheduled",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

// ---------------------------------------------------------------------------
// Tests: UpdatePost
// ---------------------------------------------------------------------------

func TestUpdatePost_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()

	newTitle := "Updated Title"
	p, err := env.svc.UpdatePost(context.Background(), "site1", "user-1", "post-1", &post.UpdatePostReq{
		Title:   &newTitle,
		Version: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", p.Title)
	assert.Equal(t, 1, env.posts.lastUpdateVersion)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionUpdate, env.audit.lastEntry.Action)
}

func TestUpdatePost_VersionConflict(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.posts.updateErr = apperror.VersionConflict("post has been modified", nil)

	newTitle := "Updated"
	_, err := env.svc.UpdatePost(context.Background(), "site1", "user-1", "post-1", &post.UpdatePostReq{
		Title:   &newTitle,
		Version: 1,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrVersionConflict))
}

// ---------------------------------------------------------------------------
// Tests: DeletePost
// ---------------------------------------------------------------------------

func TestDeletePost_Success(t *testing.T) {
	env := newTestEnv()

	err := env.svc.DeletePost(context.Background(), "site1", "post-1")
	require.NoError(t, err)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionDelete, env.audit.lastEntry.Action)
}

// ---------------------------------------------------------------------------
// Tests: Publish / Unpublish / RevertToDraft
// ---------------------------------------------------------------------------

func TestPublish_FromDraft(t *testing.T) {
	env := newTestEnv()
	p := testPost()
	p.Status = model.PostStatusDraft
	env.posts.getByID = p

	result, err := env.svc.Publish(context.Background(), "site1", "post-1")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.PostStatusPublished, env.posts.lastStatus)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionPublish, env.audit.lastEntry.Action)
}

func TestPublish_FromArchived(t *testing.T) {
	env := newTestEnv()
	p := testPost()
	p.Status = model.PostStatusArchived
	env.posts.getByID = p

	result, err := env.svc.Publish(context.Background(), "site1", "post-1")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.PostStatusPublished, env.posts.lastStatus)
}

func TestPublish_FromPublished(t *testing.T) {
	env := newTestEnv()
	p := testPost()
	p.Status = model.PostStatusPublished
	env.posts.getByID = p

	_, err := env.svc.Publish(context.Background(), "site1", "post-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestUnpublish_FromPublished(t *testing.T) {
	env := newTestEnv()
	p := testPost()
	p.Status = model.PostStatusPublished
	env.posts.getByID = p

	result, err := env.svc.Unpublish(context.Background(), "post-1")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.PostStatusArchived, env.posts.lastStatus)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionUnpublish, env.audit.lastEntry.Action)
}

func TestUnpublish_FromDraft(t *testing.T) {
	env := newTestEnv()
	p := testPost()
	p.Status = model.PostStatusDraft
	env.posts.getByID = p

	_, err := env.svc.Unpublish(context.Background(), "post-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

func TestRevertToDraft_FromScheduled(t *testing.T) {
	env := newTestEnv()
	p := testPost()
	p.Status = model.PostStatusScheduled
	env.posts.getByID = p

	result, err := env.svc.RevertToDraft(context.Background(), "post-1")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.PostStatusDraft, env.posts.lastStatus)
}

// ---------------------------------------------------------------------------
// Tests: RestorePost
// ---------------------------------------------------------------------------

func TestRestorePost_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()

	result, err := env.svc.RestorePost(context.Background(), "post-1")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionRestore, env.audit.lastEntry.Action)
}

// ---------------------------------------------------------------------------
// Tests: validateTransition (table-driven)
// ---------------------------------------------------------------------------

func TestValidateTransition_AllPaths(t *testing.T) {
	tests := []struct {
		name    string
		from    model.PostStatus
		to      model.PostStatus
		wantErr bool
	}{
		// Draft transitions
		{"draft→published", model.PostStatusDraft, model.PostStatusPublished, false},
		{"draft→scheduled", model.PostStatusDraft, model.PostStatusScheduled, false},
		{"draft→archived", model.PostStatusDraft, model.PostStatusArchived, true},
		// Scheduled transitions
		{"scheduled→draft", model.PostStatusScheduled, model.PostStatusDraft, false},
		{"scheduled→published", model.PostStatusScheduled, model.PostStatusPublished, false},
		{"scheduled→archived", model.PostStatusScheduled, model.PostStatusArchived, true},
		// Published transitions
		{"published→draft", model.PostStatusPublished, model.PostStatusDraft, false},
		{"published→archived", model.PostStatusPublished, model.PostStatusArchived, false},
		{"published→scheduled", model.PostStatusPublished, model.PostStatusScheduled, true},
		// Archived transitions
		{"archived→published", model.PostStatusArchived, model.PostStatusPublished, false},
		{"archived→draft", model.PostStatusArchived, model.PostStatusDraft, false},
		{"archived→scheduled", model.PostStatusArchived, model.PostStatusScheduled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We use the exported Publish/Unpublish/RevertToDraft instead of
			// directly calling validateTransition (unexported). This tests the
			// same behavior via public API.
			// However, for thorough coverage, we test the state machine logic
			// by using UpdatePost with status field.
			env := newTestEnv()
			p := testPost()
			p.Status = tt.from
			env.posts.getByID = p

			newStatus := ""
			switch tt.to {
			case model.PostStatusDraft:
				newStatus = "draft"
			case model.PostStatusScheduled:
				newStatus = "scheduled"
			case model.PostStatusPublished:
				newStatus = "published"
			case model.PostStatusArchived:
				newStatus = "archived"
			}

			_, err := env.svc.UpdatePost(context.Background(), "site1", "user-1", "post-1", &post.UpdatePostReq{
				Status:  &newStatus,
				Version: 1,
			})
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, apperror.ErrValidation))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: buildDiffSummary (via UpdatePost)
// ---------------------------------------------------------------------------

func TestBuildDiffSummary_TitleChange(t *testing.T) {
	env := newTestEnv()
	p := testPost()
	p.Title = "Old Title"
	p.Content = "same"
	env.posts.getByID = p

	newTitle := "New Title"
	_, err := env.svc.UpdatePost(context.Background(), "site1", "user-1", "post-1", &post.UpdatePostReq{
		Title:   &newTitle,
		Version: 1,
	})
	require.NoError(t, err)
	// Revision was created — check that mock captured the call.
	// Since we can't directly inspect buildDiffSummary, we verify no error on the path.
}

func TestBuildDiffSummary_NoChanges(t *testing.T) {
	env := newTestEnv()
	p := testPost()
	env.posts.getByID = p

	// Update with no actual changes.
	_, err := env.svc.UpdatePost(context.Background(), "site1", "user-1", "post-1", &post.UpdatePostReq{
		Version: 1,
	})
	require.NoError(t, err)
}

package comment

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/mail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockRepo struct {
	comments        map[string]*model.Comment
	listRows        []CommentRow
	listTotal       int64
	listErr         error
	getByIDErr      error
	children        []*model.Comment
	childrenErr     error
	updateStatusErr error
	updatePinnedErr error
	createErr       error
	batchAffected   int64
	batchErr        error
	deleteErr       error
	pinnedCount     int64
	pinnedCountErr  error
	parentDepth     int
	parentDepthErr  error
	lastCreated     *model.Comment
}

func (m *mockRepo) List(_ context.Context, _ ListFilter) ([]CommentRow, int64, error) {
	return m.listRows, m.listTotal, m.listErr
}
func (m *mockRepo) GetByID(_ context.Context, id string) (*model.Comment, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if c, ok := m.comments[id]; ok {
		return c, nil
	}
	return &model.Comment{ID: id, PostID: "post-1"}, nil
}
func (m *mockRepo) GetChildren(_ context.Context, _ string) ([]*model.Comment, error) {
	return m.children, m.childrenErr
}
func (m *mockRepo) UpdateStatus(_ context.Context, _ string, _ model.CommentStatus) error {
	return m.updateStatusErr
}
func (m *mockRepo) UpdatePinned(_ context.Context, _ string, _ model.Toggle) error {
	return m.updatePinnedErr
}
func (m *mockRepo) Create(_ context.Context, c *model.Comment) error {
	m.lastCreated = c
	return m.createErr
}
func (m *mockRepo) BatchUpdateStatus(_ context.Context, _ []string, _ model.CommentStatus) (int64, error) {
	return m.batchAffected, m.batchErr
}
func (m *mockRepo) Delete(_ context.Context, _ string) error {
	return m.deleteErr
}
func (m *mockRepo) CountPinnedByPost(_ context.Context, _ string) (int64, error) {
	return m.pinnedCount, m.pinnedCountErr
}
func (m *mockRepo) GetParentChainDepth(_ context.Context, _ string) (int, error) {
	return m.parentDepth, m.parentDepthErr
}

type mockAudit struct{ lastEntry audit.Entry }

func (m *mockAudit) Log(_ context.Context, e audit.Entry) error {
	m.lastEntry = e
	return nil
}

type testEnv struct {
	svc   *Service
	repo  *mockRepo
	audit *mockAudit
}

func newTestEnv() *testEnv {
	repo := &mockRepo{comments: make(map[string]*model.Comment)}
	a := &mockAudit{}
	mailer := &mail.NoopSender{}
	svc := NewService(repo, a, mailer)
	return &testEnv{svc: svc, repo: repo, audit: a}
}

// --- Tests ---

func TestList_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.listRows = []CommentRow{
		{Comment: model.Comment{ID: "c1", PostID: "p1", AuthorEmail: "a@b.com", Status: model.CommentStatusApproved}},
	}
	env.repo.listTotal = 1

	results, total, err := env.svc.List(context.Background(), ListFilter{Page: 1, PerPage: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "approved", results[0].Status)
	assert.Contains(t, results[0].GravatarURL, "gravatar.com")
}

func TestGetComment_WithChildren(t *testing.T) {
	env := newTestEnv()
	env.repo.comments["c1"] = &model.Comment{
		ID: "c1", PostID: "p1", Content: "hello",
		AuthorEmail: "a@b.com", Status: model.CommentStatusApproved,
	}
	env.repo.children = []*model.Comment{
		{ID: "c2", ParentID: strPtr("c1"), Content: "reply"},
	}

	resp, err := env.svc.GetComment(context.Background(), "c1")
	require.NoError(t, err)
	assert.Equal(t, "c1", resp.ID)
	assert.Len(t, resp.Children, 1)
}

func TestTogglePin_OnlyTopLevel(t *testing.T) {
	env := newTestEnv()
	parentID := "parent-1"
	env.repo.comments["c1"] = &model.Comment{ID: "c1", ParentID: &parentID}

	err := env.svc.TogglePin(context.Background(), "c1", &TogglePinReq{Pinned: true})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "top-level")
}

func TestTogglePin_MaxThree(t *testing.T) {
	env := newTestEnv()
	env.repo.comments["c1"] = &model.Comment{ID: "c1", PostID: "p1", Pinned: model.ToggleNo}
	env.repo.pinnedCount = 3

	err := env.svc.TogglePin(context.Background(), "c1", &TogglePinReq{Pinned: true})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum 3")
}

func TestReply_MaxDepth(t *testing.T) {
	env := newTestEnv()
	env.repo.comments["c1"] = &model.Comment{ID: "c1", PostID: "p1"}
	env.repo.parentDepth = 2

	_, err := env.svc.Reply(context.Background(), "c1", &ReplyReq{Content: "hi"}, "u1", "Admin", "admin@test.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nesting depth")
}

func TestReply_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.comments["c1"] = &model.Comment{ID: "c1", PostID: "p1", AuthorEmail: "guest@test.com"}
	env.repo.parentDepth = 0

	resp, err := env.svc.Reply(context.Background(), "c1", &ReplyReq{Content: "thanks"}, "u1", "Admin", "admin@test.com")
	require.NoError(t, err)
	assert.Equal(t, "approved", resp.Status)
	assert.Equal(t, "Admin", resp.AuthorName)
	assert.Equal(t, model.LogActionCreate, env.audit.lastEntry.Action)
}

func TestBatchUpdateStatus_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.batchAffected = 5

	count, err := env.svc.BatchUpdateStatus(context.Background(), &BatchStatusReq{
		IDs:    []string{"c1", "c2", "c3", "c4", "c5"},
		Status: "approved",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestDeleteComment_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.comments["c1"] = &model.Comment{ID: "c1"}

	err := env.svc.DeleteComment(context.Background(), "c1")
	require.NoError(t, err)
	assert.Equal(t, model.LogActionDelete, env.audit.lastEntry.Action)
}

func strPtr(s string) *string { return &s }

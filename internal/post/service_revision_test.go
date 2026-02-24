package post_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Tests: ListRevisions
// ---------------------------------------------------------------------------

func TestListRevisions_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.revs.listRevs = []model.PostRevision{
		{ID: "rev-1", PostID: "post-1", Version: 1, DiffSummary: "Initial version"},
		{ID: "rev-2", PostID: "post-1", Version: 2, DiffSummary: "Updated title"},
	}

	revs, err := env.svc.ListRevisions(context.Background(), "post-1")
	require.NoError(t, err)
	assert.Len(t, revs, 2)
	assert.Equal(t, "rev-1", revs[0].ID)
}

func TestListRevisions_PostNotFound(t *testing.T) {
	env := newTestEnv()
	env.posts.getByIDErr = apperror.NotFound("post not found", nil)

	_, err := env.svc.ListRevisions(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: Rollback
// ---------------------------------------------------------------------------

func TestRollback_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.revs.getByID = &model.PostRevision{
		ID:       "rev-1",
		PostID:   "post-1",
		Version:  1,
		Title:    "Old Title",
		Content:  "Old Content",
	}

	result, err := env.svc.Rollback(context.Background(), "site1", "user-1", "post-1", "rev-1")
	require.NoError(t, err)
	assert.Equal(t, "Old Title", result.Title)
	assert.Equal(t, "Old Content", result.Content)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionUpdate, env.audit.lastEntry.Action)
}

func TestRollback_RevisionNotFound(t *testing.T) {
	env := newTestEnv()
	env.revs.getErr = apperror.NotFound("revision not found", nil)

	_, err := env.svc.Rollback(context.Background(), "site1", "user-1", "post-1", "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

func TestRollback_WrongPost(t *testing.T) {
	env := newTestEnv()
	env.revs.getByID = &model.PostRevision{
		ID:     "rev-1",
		PostID: "other-post", // belongs to a different post
	}

	_, err := env.svc.Rollback(context.Background(), "site1", "user-1", "post-1", "rev-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not belong")
}

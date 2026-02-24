package post_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Tests: CreatePreviewToken
// ---------------------------------------------------------------------------

func TestCreatePreviewToken_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.preview.countActive = 0

	resp, err := env.svc.CreatePreviewToken(context.Background(), "post-1", "user-1")
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Contains(t, resp.Token, "sky_preview_")
	assert.Equal(t, "token-new-id", resp.ID)
	assert.Equal(t, 1, resp.ActiveCount)
	assert.False(t, resp.ExpiresAt.IsZero())
}

func TestCreatePreviewToken_PostNotFound(t *testing.T) {
	env := newTestEnv()
	env.posts.getByIDErr = apperror.NotFound("post not found", nil)

	_, err := env.svc.CreatePreviewToken(context.Background(), "nonexistent", "user-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

func TestCreatePreviewToken_LimitReached(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.preview.countActive = 5

	_, err := env.svc.CreatePreviewToken(context.Background(), "post-1", "user-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
	assert.Contains(t, err.Error(), "limit reached")
}

// ---------------------------------------------------------------------------
// Tests: ListPreviewTokens
// ---------------------------------------------------------------------------

func TestListPreviewTokens_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.preview.listTokens = []model.PreviewToken{
		{ID: "t1", PostID: "post-1", ExpiresAt: time.Now().Add(time.Hour)},
		{ID: "t2", PostID: "post-1", ExpiresAt: time.Now().Add(2 * time.Hour)},
	}

	tokens, err := env.svc.ListPreviewTokens(context.Background(), "post-1")
	require.NoError(t, err)
	assert.Len(t, tokens, 2)
}

// ---------------------------------------------------------------------------
// Tests: RevokeAllPreviewTokens
// ---------------------------------------------------------------------------

func TestRevokeAllPreviewTokens_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.preview.delAllCount = 3

	count, err := env.svc.RevokeAllPreviewTokens(context.Background(), "post-1")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionDelete, env.audit.lastEntry.Action)
	assert.Equal(t, "preview_token", env.audit.lastEntry.ResourceType)
}

// ---------------------------------------------------------------------------
// Tests: RevokePreviewToken
// ---------------------------------------------------------------------------

func TestRevokePreviewToken_Success(t *testing.T) {
	env := newTestEnv()

	err := env.svc.RevokePreviewToken(context.Background(), "token-1")
	require.NoError(t, err)
}

func TestRevokePreviewToken_NotFound(t *testing.T) {
	env := newTestEnv()
	env.preview.delByIDErr = apperror.NotFound("preview token not found", nil)

	err := env.svc.RevokePreviewToken(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: GetPostByPreviewToken
// ---------------------------------------------------------------------------

func TestGetPostByPreviewToken_Success(t *testing.T) {
	env := newTestEnv()
	env.preview.getByHash = &model.PreviewToken{
		ID:     "t1",
		PostID: "post-1",
	}
	env.posts.getByID = testPost()

	p, err := env.svc.GetPostByPreviewToken(context.Background(), "sky_preview_abc123")
	require.NoError(t, err)
	assert.Equal(t, "post-1", p.ID)
}

func TestGetPostByPreviewToken_Expired(t *testing.T) {
	env := newTestEnv()
	env.preview.getHashErr = apperror.NotFound("preview token not found or expired", nil)

	_, err := env.svc.GetPostByPreviewToken(context.Background(), "sky_preview_expired")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

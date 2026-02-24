package post_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/post"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Tests: ListTranslations
// ---------------------------------------------------------------------------

func TestListTranslations_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.trans.listTrans = []model.PostTranslation{
		{PostID: "post-1", Locale: "en", Title: "English Title"},
		{PostID: "post-1", Locale: "zh-CN", Title: "中文标题"},
	}

	ts, err := env.svc.ListTranslations(context.Background(), "post-1")
	require.NoError(t, err)
	assert.Len(t, ts, 2)
	assert.Equal(t, "en", ts[0].Locale)
	assert.Equal(t, "zh-CN", ts[1].Locale)
}

// ---------------------------------------------------------------------------
// Tests: GetTranslation
// ---------------------------------------------------------------------------

func TestGetTranslation_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.trans.getTrans = &model.PostTranslation{
		PostID: "post-1",
		Locale: "en",
		Title:  "English Title",
	}

	tr, err := env.svc.GetTranslation(context.Background(), "post-1", "en")
	require.NoError(t, err)
	assert.Equal(t, "English Title", tr.Title)
	assert.Equal(t, "en", tr.Locale)
}

func TestGetTranslation_NotFound(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.trans.getErr = apperror.NotFound("translation not found", nil)

	_, err := env.svc.GetTranslation(context.Background(), "post-1", "fr")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: UpsertTranslation
// ---------------------------------------------------------------------------

func TestUpsertTranslation_Create(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.trans.getTrans = &model.PostTranslation{
		PostID:    "post-1",
		Locale:    "en",
		Title:     "English Title",
		Content:   "English body",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := env.svc.UpsertTranslation(context.Background(), "post-1", "en", &post.UpsertTranslationReq{
		Title:   "English Title",
		Content: "English body",
	})
	require.NoError(t, err)
	assert.Equal(t, "en", result.Locale)
	assert.Equal(t, "English Title", result.Title)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionUpdate, env.audit.lastEntry.Action)
	assert.Equal(t, "post_translation", env.audit.lastEntry.ResourceType)
}

func TestUpsertTranslation_PostNotFound(t *testing.T) {
	env := newTestEnv()
	env.posts.getByIDErr = apperror.NotFound("post not found", nil)

	_, err := env.svc.UpsertTranslation(context.Background(), "nonexistent", "en", &post.UpsertTranslationReq{
		Title: "test",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Tests: DeleteTranslation
// ---------------------------------------------------------------------------

func TestDeleteTranslation_Success(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()

	err := env.svc.DeleteTranslation(context.Background(), "post-1", "en")
	require.NoError(t, err)
	assert.NotNil(t, env.audit.lastEntry)
	assert.Equal(t, model.LogActionDelete, env.audit.lastEntry.Action)
	assert.Equal(t, "post_translation", env.audit.lastEntry.ResourceType)
}

func TestDeleteTranslation_NotFound(t *testing.T) {
	env := newTestEnv()
	env.posts.getByID = testPost()
	env.trans.deleteErr = apperror.NotFound("translation not found", nil)

	err := env.svc.DeleteTranslation(context.Background(), "post-1", "fr")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

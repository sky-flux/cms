package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c := gin.CreateTestContextOnly(w, gin.New())
	return c, w
}

func TestSuccess(t *testing.T) {
	c, w := newTestContext()
	Success(c, map[string]string{"name": "test"})

	assert.Equal(t, http.StatusOK, w.Code)

	var resp body
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
}

func TestCreated(t *testing.T) {
	c, w := newTestContext()
	Created(c, map[string]string{"id": "123"})

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp body
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
}

func TestNoContent(t *testing.T) {
	r := gin.New()
	r.DELETE("/test", func(c *gin.Context) {
		NoContent(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestError_AppError(t *testing.T) {
	c, w := newTestContext()
	appErr := apperror.NotFound("user not found", nil)
	Error(c, appErr)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp body
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	assert.Equal(t, "user not found", resp.Error)
}

func TestError_SentinelError(t *testing.T) {
	c, w := newTestContext()
	Error(c, apperror.ErrForbidden)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp body
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
}

func TestError_GenericError(t *testing.T) {
	c, w := newTestContext()
	Error(c, errors.New("something broke"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPaginated(t *testing.T) {
	c, w := newTestContext()
	items := []string{"a", "b", "c"}
	Paginated(c, items, 42, 2, 10)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp paginatedBody
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, int64(42), resp.Meta.Total)
	assert.Equal(t, 2, resp.Meta.Page)
	assert.Equal(t, 10, resp.Meta.PerPage)
}

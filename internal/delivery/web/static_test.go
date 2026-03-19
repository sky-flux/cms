package web_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	rootfs "github.com/sky-flux/cms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticFileServer_ServesCSS(t *testing.T) {
	sub, err := fs.Sub(rootfs.WebStaticFS, "web/static")
	require.NoError(t, err)

	r := http.NewServeMux()
	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))

	req := httptest.NewRequest(http.MethodGet, "/static/app.css", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/css")
}

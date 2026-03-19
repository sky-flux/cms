package cms_test

import (
	"io/fs"
	"testing"

	cms "github.com/sky-flux/cms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsoleFSReadable(t *testing.T) {
	// console/dist/.gitkeep must exist for go:embed to compile.
	// This test verifies the embedded FS is accessible at runtime.
	_, err := fs.Stat(cms.ConsoleFS, "console/dist/.gitkeep")
	require.NoError(t, err, "console/dist/.gitkeep must be embedded — run: mkdir -p console/dist && touch console/dist/.gitkeep")
}

func TestWebStaticFSReadable(t *testing.T) {
	_, err := fs.Stat(cms.WebStaticFS, "web/static/htmx.min.js")
	require.NoError(t, err, "web/static/htmx.min.js must be embedded")
	assert.NoError(t, err)
}

func TestWebStaticFS_ContainsRequiredFiles(t *testing.T) {
	sub, err := fs.Sub(cms.WebStaticFS, "web/static")
	require.NoError(t, err)

	files := []string{"app.css", "htmx.min.js"}
	for _, f := range files {
		_, err := sub.Open(f)
		assert.NoError(t, err, "expected %s in WebStaticFS", f)
	}
}

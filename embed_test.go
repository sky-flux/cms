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
	_, err := fs.Stat(cms.WebStaticFS, "web/static/.gitkeep")
	require.NoError(t, err, "web/static/.gitkeep must be embedded — run: mkdir -p web/static && touch web/static/.gitkeep")
	assert.NoError(t, err)
}

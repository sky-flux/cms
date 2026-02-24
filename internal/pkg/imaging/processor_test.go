package imaging_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/sky-flux/cms/internal/pkg/imaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testImage(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := range w {
		for y := range h {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestExtractDimensions(t *testing.T) {
	p := imaging.NewProcessor()
	data := testImage(800, 600)

	w, h, err := p.ExtractDimensions(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, 800, w)
	assert.Equal(t, 600, h)
}

func TestThumbnail_Crop(t *testing.T) {
	p := imaging.NewProcessor()
	data := testImage(800, 600)

	thumb, err := p.Thumbnail(bytes.NewReader(data), 150, 150, "crop")
	require.NoError(t, err)
	assert.NotEmpty(t, thumb)

	cfg, _, err := image.DecodeConfig(bytes.NewReader(thumb))
	require.NoError(t, err)
	assert.Equal(t, 150, cfg.Width)
	assert.Equal(t, 150, cfg.Height)
}

func TestThumbnail_Fit(t *testing.T) {
	p := imaging.NewProcessor()
	data := testImage(800, 600)

	thumb, err := p.Thumbnail(bytes.NewReader(data), 400, 400, "fit")
	require.NoError(t, err)
	assert.NotEmpty(t, thumb)

	cfg, _, err := image.DecodeConfig(bytes.NewReader(thumb))
	require.NoError(t, err)
	assert.LessOrEqual(t, cfg.Width, 400)
	assert.LessOrEqual(t, cfg.Height, 400)
}

package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	_ "golang.org/x/image/webp"

	"golang.org/x/image/draw"
)

// Processor handles image operations.
type Processor struct{}

// NewProcessor creates a new image processor.
func NewProcessor() *Processor {
	return &Processor{}
}

// ExtractDimensions returns the width and height of an image.
func (p *Processor) ExtractDimensions(src io.Reader) (width, height int, err error) {
	cfg, _, err := image.DecodeConfig(src)
	if err != nil {
		return 0, 0, fmt.Errorf("decode image config: %w", err)
	}
	return cfg.Width, cfg.Height, nil
}

// Thumbnail generates a thumbnail of the given dimensions.
// mode: "crop" for center-crop to exact size, "fit" for fit-within bounds.
func (p *Processor) Thumbnail(src io.Reader, width, height int, mode string) ([]byte, error) {
	img, format, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	var dst *image.RGBA
	srcBounds := img.Bounds()

	if mode == "crop" {
		dst = p.cropCenter(img, srcBounds, width, height)
	} else {
		dst = p.fitWithin(img, srcBounds, width, height)
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 80})
	case "png":
		err = png.Encode(&buf, dst)
	case "gif":
		err = gif.Encode(&buf, dst, nil)
	default:
		err = png.Encode(&buf, dst)
	}
	if err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}
	return buf.Bytes(), nil
}

// cropCenter crops from the center to exact dimensions.
func (p *Processor) cropCenter(img image.Image, bounds image.Rectangle, w, h int) *image.RGBA {
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	// Determine crop region.
	srcRatio := float64(srcW) / float64(srcH)
	dstRatio := float64(w) / float64(h)

	var cropRect image.Rectangle
	if srcRatio > dstRatio {
		cropH := srcH
		cropW := int(float64(cropH) * dstRatio)
		x0 := (srcW - cropW) / 2
		cropRect = image.Rect(x0, 0, x0+cropW, cropH)
	} else {
		cropW := srcW
		cropH := int(float64(cropW) / dstRatio)
		y0 := (srcH - cropH) / 2
		cropRect = image.Rect(0, y0, cropW, y0+cropH)
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, cropRect, draw.Over, nil)
	return dst
}

// fitWithin scales to fit within bounds, maintaining aspect ratio.
func (p *Processor) fitWithin(img image.Image, bounds image.Rectangle, maxW, maxH int) *image.RGBA {
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	ratio := min(float64(maxW)/float64(srcW), float64(maxH)/float64(srcH))
	if ratio >= 1 {
		ratio = 1
	}

	dstW := int(float64(srcW) * ratio)
	dstH := int(float64(srcH) * ratio)

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst
}

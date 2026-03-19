package domain

import (
	"errors"
	"strings"
	"time"
)

// Sentinel errors — no framework deps.
var (
	ErrEmptyFilename       = errors.New("filename must not be empty")
	ErrUnsupportedMIMEType = errors.New("unsupported MIME type")
	ErrFileTooLarge        = errors.New("file exceeds maximum size of 20 MB")
	ErrEmptyStorageKey     = errors.New("storage key must not be empty")
)

const MaxFileSize = 20 * 1024 * 1024 // 20 MB

var allowedMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// MediaFile is the aggregate root for the Media BC.
// ID is empty on construction; the DB sets it via uuidv7().
type MediaFile struct {
	ID         string
	Filename   string
	MimeType   string
	Size       int64
	StorageKey string
	URL        string
	Width      int
	Height     int
	ThumbSmKey string
	ThumbSmURL string
	ThumbMdKey string
	ThumbMdURL string
	UploaderID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewMediaFile validates inputs and constructs a MediaFile ready for persistence.
func NewMediaFile(filename, mimeType string, size int64, storageKey, url string) (*MediaFile, error) {
	if strings.TrimSpace(filename) == "" {
		return nil, ErrEmptyFilename
	}
	if !allowedMIMETypes[mimeType] {
		return nil, ErrUnsupportedMIMEType
	}
	if size > MaxFileSize {
		return nil, ErrFileTooLarge
	}
	return &MediaFile{
		Filename:   filename,
		MimeType:   mimeType,
		Size:       size,
		StorageKey: storageKey,
		URL:        url,
	}, nil
}

// Validate checks invariants on an already-constructed MediaFile.
func (f *MediaFile) Validate() error {
	if strings.TrimSpace(f.StorageKey) == "" {
		return ErrEmptyStorageKey
	}
	return nil
}

// IsImage returns true when the file is a raster image.
func (f *MediaFile) IsImage() bool {
	return strings.HasPrefix(f.MimeType, "image/")
}

// SetDimensions records image width and height after extraction.
func (f *MediaFile) SetDimensions(width, height int) {
	f.Width = width
	f.Height = height
}

// SetThumbnails records storage keys and public URLs for both thumbnail sizes.
func (f *MediaFile) SetThumbnails(smKey, smURL, mdKey, mdURL string) {
	f.ThumbSmKey = smKey
	f.ThumbSmURL = smURL
	f.ThumbMdKey = mdKey
	f.ThumbMdURL = mdURL
}

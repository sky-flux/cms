package media

import (
	"context"
	"io"

	"github.com/sky-flux/cms/internal/model"
)

// MediaRepository handles sfc_site_media_files table CRUD.
type MediaRepository interface {
	List(ctx context.Context, filter ListFilter) ([]model.MediaFile, int64, error)
	GetByID(ctx context.Context, id string) (*model.MediaFile, error)
	Create(ctx context.Context, mf *model.MediaFile) error
	Update(ctx context.Context, mf *model.MediaFile) error
	SoftDelete(ctx context.Context, id string) error
	BatchSoftDelete(ctx context.Context, ids []string) (int64, error)
	GetReferencingPosts(ctx context.Context, mediaID string) ([]PostRef, error)
	GetBatchReferencingPosts(ctx context.Context, mediaIDs []string) (map[string]int64, error)
}

// StorageUploader abstracts S3-compatible storage operations.
type StorageUploader interface {
	Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) error
	UploadBytes(ctx context.Context, key string, data []byte, contentType string) error
	Delete(ctx context.Context, key string) error
	BatchDelete(ctx context.Context, keys []string) error
	PublicURL(key string) string
}

// ImageProcessor abstracts image dimension extraction and thumbnail generation.
type ImageProcessor interface {
	ExtractDimensions(src io.Reader) (width, height int, err error)
	Thumbnail(src io.Reader, width, height int, mode string) ([]byte, error)
}

// PostRef is a lightweight reference to a post that uses a media file.
type PostRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ListFilter holds query parameters for listing media files.
type ListFilter struct {
	Page      int
	PerPage   int
	Query     string
	MediaType *int // filter by media_type enum value
}

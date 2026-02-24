package media

import (
	"encoding/json"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

// UpdateMediaReq is the request body for updating a media file.
type UpdateMediaReq struct {
	AltText *string `json:"alt_text"`
}

// BatchDeleteReq is the request body for batch deleting media files.
type BatchDeleteReq struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}

// --- Response DTOs ---

// BatchDeleteResp is the response for a batch delete operation.
type BatchDeleteResp struct {
	DeletedCount int            `json:"deleted_count"`
	Skipped      []SkippedMedia `json:"skipped,omitempty"`
}

// SkippedMedia describes a media file that was skipped during batch delete.
type SkippedMedia struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

// MediaResp is the JSON response representation of a media file.
type MediaResp struct {
	ID             string          `json:"id"`
	UploaderID     string          `json:"uploader_id"`
	FileName       string          `json:"file_name"`
	OriginalName   string          `json:"original_name"`
	MimeType       string          `json:"mime_type"`
	MediaType      int             `json:"media_type"`
	FileSize       int64           `json:"file_size"`
	Width          *int            `json:"width,omitempty"`
	Height         *int            `json:"height,omitempty"`
	StoragePath    string          `json:"storage_path"`
	PublicURL      string          `json:"public_url"`
	WebpURL        string          `json:"webp_url,omitempty"`
	ThumbnailURLs  json.RawMessage `json:"thumbnail_urls,omitempty"`
	ReferenceCount int             `json:"reference_count"`
	AltText        string          `json:"alt_text,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// ToMediaResp converts a model.MediaFile to MediaResp.
func ToMediaResp(mf *model.MediaFile) MediaResp {
	return MediaResp{
		ID:             mf.ID,
		UploaderID:     mf.UploaderID,
		FileName:       mf.FileName,
		OriginalName:   mf.OriginalName,
		MimeType:       mf.MimeType,
		MediaType:      int(mf.MediaType),
		FileSize:       mf.FileSize,
		Width:          mf.Width,
		Height:         mf.Height,
		StoragePath:    mf.StoragePath,
		PublicURL:      mf.PublicURL,
		WebpURL:        mf.WebpURL,
		ThumbnailURLs:  mf.ThumbnailURLs,
		ReferenceCount: mf.ReferenceCount,
		AltText:        mf.AltText,
		CreatedAt:      mf.CreatedAt,
		UpdatedAt:      mf.UpdatedAt,
	}
}

// ToMediaRespList converts a slice of model.MediaFile to a slice of MediaResp.
func ToMediaRespList(files []model.MediaFile) []MediaResp {
	out := make([]MediaResp, len(files))
	for i := range files {
		out[i] = ToMediaResp(&files[i])
	}
	return out
}

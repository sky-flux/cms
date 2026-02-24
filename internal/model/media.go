package model

import (
	"context"
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type MediaFile struct {
	bun.BaseModel `bun:"table:sfc_site_media_files,alias:mf"`

	ID             string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	UploaderID     string          `bun:"uploader_id,notnull,type:uuid" json:"uploader_id"`
	FileName       string          `bun:"file_name,notnull" json:"file_name"`
	OriginalName   string          `bun:"original_name,notnull" json:"original_name"`
	MimeType       string          `bun:"mime_type,notnull" json:"mime_type"`
	MediaType      MediaType       `bun:"media_type,notnull,type:smallint,default:5" json:"media_type"`
	FileSize       int64           `bun:"file_size,notnull" json:"file_size"`
	Width          *int            `bun:"width" json:"width,omitempty"`
	Height         *int            `bun:"height" json:"height,omitempty"`
	StoragePath    string          `bun:"storage_path,notnull" json:"storage_path"`
	PublicURL      string          `bun:"public_url,notnull" json:"public_url"`
	WebpURL        string          `bun:"webp_url" json:"webp_url,omitempty"`
	ThumbnailURLs  json.RawMessage `bun:"thumbnail_urls,type:jsonb,default:'{}'" json:"thumbnail_urls,omitempty"`
	ReferenceCount int             `bun:"reference_count,notnull,default:0" json:"reference_count"`
	AltText        string          `bun:"alt_text" json:"alt_text,omitempty"`
	Metadata       json.RawMessage `bun:"metadata,type:jsonb,default:'{}'" json:"metadata,omitempty"`
	CreatedAt      time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt      *time.Time      `bun:"deleted_at,soft_delete,nullzero" json:"-"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (mf *MediaFile) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&mf.CreatedAt, &mf.UpdatedAt, query)
	return nil
}

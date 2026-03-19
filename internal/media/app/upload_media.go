package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sky-flux/cms/internal/media/domain"
)

// UploadMediaInput carries validated input from the delivery layer.
type UploadMediaInput struct {
	Filename   string
	MimeType   string
	Data       []byte
	UploaderID string
}

// UploadMediaUseCase orchestrates file validation, storage upload, and persistence.
type UploadMediaUseCase struct {
	repo    domain.MediaFileRepository
	storage domain.StoragePort
}

func NewUploadMediaUseCase(repo domain.MediaFileRepository, storage domain.StoragePort) *UploadMediaUseCase {
	return &UploadMediaUseCase{repo: repo, storage: storage}
}

func (uc *UploadMediaUseCase) Execute(ctx context.Context, in UploadMediaInput) (*domain.MediaFile, error) {
	// Validate type and size via domain constructor.
	mf, err := domain.NewMediaFile(in.Filename, in.MimeType, int64(len(in.Data)), "", "")
	if err != nil {
		return nil, err
	}
	mf.UploaderID = in.UploaderID

	// Generate deterministic object key: media/{year}/{month}/{uuid}.{ext}
	key := buildStorageKey(in.Filename, time.Now().UTC())

	// Upload to RustFS via StoragePort.
	if err := uc.storage.Upload(ctx, key, in.Data, in.MimeType); err != nil {
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	mf.StorageKey = key
	mf.URL = uc.storage.URL(key)

	// Persist record.
	if err := uc.repo.Save(ctx, mf); err != nil {
		return nil, fmt.Errorf("save media file: %w", err)
	}
	return mf, nil
}

// buildStorageKey produces the object key for the given filename and upload time.
func buildStorageKey(filename string, t time.Time) string {
	ext := strings.ToLower(filepath.Ext(filename))
	return fmt.Sprintf("media/%d/%02d/%d%s", t.Year(), t.Month(), t.UnixNano(), ext)
}

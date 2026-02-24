package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
)

// allowedMIMEs is the whitelist of accepted MIME types for upload.
var allowedMIMEs = map[string]bool{
	"image/jpeg":       true,
	"image/png":        true,
	"image/gif":        true,
	"image/webp":       true,
	"image/svg+xml":    true,
	"video/mp4":        true,
	"video/webm":       true,
	"application/pdf":  true,
	"audio/mpeg":       true,
	"audio/ogg":        true,
}

// Service implements media business logic.
type Service struct {
	repo    MediaRepository
	storage StorageUploader
	imaging ImageProcessor
	audit   audit.Logger
}

// NewService creates a new media service.
func NewService(repo MediaRepository, storage StorageUploader, imaging ImageProcessor, auditLogger audit.Logger) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
		imaging: imaging,
		audit:   auditLogger,
	}
}

// List returns a paginated list of media files.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]MediaResp, int64, error) {
	files, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list media: %w", err)
	}
	return ToMediaRespList(files), total, nil
}

// GetMedia returns a single media file by ID.
func (s *Service) GetMedia(ctx context.Context, id string) (*MediaResp, error) {
	mf, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := ToMediaResp(mf)
	return &resp, nil
}

// Upload handles the complete upload flow: validate, store, create DB record.
func (s *Service) Upload(ctx context.Context, siteSlug, uploaderID string, file io.Reader, fileName, mimeType string, fileSize int64, altText string) (*MediaResp, error) {
	// 1. Validate MIME type.
	if !allowedMIMEs[mimeType] {
		return nil, apperror.Validation("unsupported file type: "+mimeType, nil)
	}

	// 2. Determine MediaType from MIME prefix.
	mediaType := classifyMIME(mimeType)

	// 3. Generate storage key.
	now := time.Now()
	ext := filepath.Ext(fileName)
	uid := uuid.New().String()
	storageKey := fmt.Sprintf("media/%s/%s/%s%s", now.Format("2006"), now.Format("01"), uid, ext)

	// 4. Read the file into a buffer so we can process and upload it.
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return nil, fmt.Errorf("read upload data: %w", err)
	}
	fileData := buf.Bytes()

	// 5. Image processing (skip SVG).
	var width, height *int
	thumbnailURLs := json.RawMessage("{}")

	isImage := strings.HasPrefix(mimeType, "image/") && mimeType != "image/svg+xml"
	if isImage && s.imaging != nil {
		// Extract dimensions.
		w, h, err := s.imaging.ExtractDimensions(bytes.NewReader(fileData))
		if err != nil {
			slog.Warn("extract dimensions failed", "error", err, "file", fileName)
		} else {
			width = &w
			height = &h
		}

		// Generate thumbnails.
		thumbs := map[string]string{}

		smData, err := s.imaging.Thumbnail(bytes.NewReader(fileData), 150, 150, "crop")
		if err == nil {
			smKey := fmt.Sprintf("media/%s/%s/%s_sm%s", now.Format("2006"), now.Format("01"), uid, ext)
			if uploadErr := s.storage.UploadBytes(ctx, smKey, smData, mimeType); uploadErr != nil {
				slog.Warn("upload sm thumbnail failed", "error", uploadErr)
			} else {
				thumbs["sm"] = s.storage.PublicURL(smKey)
			}
		}

		mdData, err := s.imaging.Thumbnail(bytes.NewReader(fileData), 400, 400, "fit")
		if err == nil {
			mdKey := fmt.Sprintf("media/%s/%s/%s_md%s", now.Format("2006"), now.Format("01"), uid, ext)
			if uploadErr := s.storage.UploadBytes(ctx, mdKey, mdData, mimeType); uploadErr != nil {
				slog.Warn("upload md thumbnail failed", "error", uploadErr)
			} else {
				thumbs["md"] = s.storage.PublicURL(mdKey)
			}
		}

		if len(thumbs) > 0 {
			data, _ := json.Marshal(thumbs)
			thumbnailURLs = data
		}
	}

	// 6. Upload original file.
	if err := s.storage.Upload(ctx, storageKey, bytes.NewReader(fileData), mimeType, fileSize); err != nil {
		return nil, fmt.Errorf("upload file to storage: %w", err)
	}

	publicURL := s.storage.PublicURL(storageKey)

	// 7. Build DB record.
	mf := &model.MediaFile{
		UploaderID:     uploaderID,
		FileName:       uid + ext,
		OriginalName:   fileName,
		MimeType:       mimeType,
		MediaType:      mediaType,
		FileSize:       fileSize,
		Width:          width,
		Height:         height,
		StoragePath:    storageKey,
		PublicURL:      publicURL,
		ThumbnailURLs:  thumbnailURLs,
		ReferenceCount: 0,
		AltText:        altText,
	}

	if err := s.repo.Create(ctx, mf); err != nil {
		return nil, fmt.Errorf("create media record: %w", err)
	}

	// 8. Audit log.
	if err := s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionCreate,
		ResourceType: "media",
		ResourceID:   mf.ID,
	}); err != nil {
		slog.Error("audit log media upload failed", "error", err)
	}

	resp := ToMediaResp(mf)
	return &resp, nil
}

// UpdateMedia updates the alt_text of a media file.
func (s *Service) UpdateMedia(ctx context.Context, id string, req *UpdateMediaReq) (*MediaResp, error) {
	mf, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.AltText != nil {
		mf.AltText = *req.AltText
	}

	if err := s.repo.Update(ctx, mf); err != nil {
		return nil, fmt.Errorf("update media: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "media",
		ResourceID:   mf.ID,
	}); err != nil {
		slog.Error("audit log media update failed", "error", err)
	}

	resp := ToMediaResp(mf)
	return &resp, nil
}

// DeleteMedia soft-deletes a media file. If force is false and the file is
// referenced by posts, it returns a Conflict error with the referencing posts.
func (s *Service) DeleteMedia(ctx context.Context, id string, force bool) error {
	mf, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	refs, err := s.repo.GetReferencingPosts(ctx, id)
	if err != nil {
		return fmt.Errorf("check media references: %w", err)
	}

	if len(refs) > 0 && !force {
		refData, _ := json.Marshal(refs)
		return apperror.Conflict(
			fmt.Sprintf("media file is referenced by %d post(s)", len(refs)),
			fmt.Errorf("referencing_posts: %s", string(refData)),
		)
	}

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("delete media: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:           model.LogActionDelete,
		ResourceType:     "media",
		ResourceID:       mf.ID,
		ResourceSnapshot: mf,
	}); err != nil {
		slog.Error("audit log media delete failed", "error", err)
	}

	return nil
}

// BatchDeleteMedia soft-deletes multiple media files, skipping those with
// references when force is false.
func (s *Service) BatchDeleteMedia(ctx context.Context, ids []string, force bool) (*BatchDeleteResp, error) {
	refCounts, err := s.repo.GetBatchReferencingPosts(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("check batch media references: %w", err)
	}

	var toDelete []string
	var skipped []SkippedMedia

	for _, id := range ids {
		count, hasRefs := refCounts[id]
		if hasRefs && count > 0 && !force {
			skipped = append(skipped, SkippedMedia{
				ID:     id,
				Reason: fmt.Sprintf("referenced by %d post(s)", count),
			})
		} else {
			toDelete = append(toDelete, id)
		}
	}

	var deletedCount int64
	if len(toDelete) > 0 {
		n, err := s.repo.BatchSoftDelete(ctx, toDelete)
		if err != nil {
			return nil, fmt.Errorf("batch delete media: %w", err)
		}
		deletedCount = n

		if err := s.audit.Log(ctx, audit.Entry{
			Action:       model.LogActionDelete,
			ResourceType: "media",
			ResourceID:   strings.Join(toDelete, ","),
		}); err != nil {
			slog.Error("audit log media batch delete failed", "error", err)
		}
	}

	return &BatchDeleteResp{
		DeletedCount: int(deletedCount),
		Skipped:      skipped,
	}, nil
}

// classifyMIME maps a MIME type to a MediaType enum value.
func classifyMIME(mime string) model.MediaType {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return model.MediaTypeImage
	case strings.HasPrefix(mime, "video/"):
		return model.MediaTypeVideo
	case strings.HasPrefix(mime, "audio/"):
		return model.MediaTypeAudio
	case mime == "application/pdf":
		return model.MediaTypeDocument
	default:
		return model.MediaTypeOther
	}
}

package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/sky-flux/cms/internal/media/domain"
	"github.com/sky-flux/cms/internal/pkg/apperror"
)

// mediaFileRow is the bun model for sfc_media_files.
type mediaFileRow struct {
	bun.BaseModel `bun:"table:sfc_media_files,alias:mf"`

	ID         string     `bun:"id,pk,type:uuid,default:uuidv7()"`
	Filename   string     `bun:"filename,notnull"`
	MimeType   string     `bun:"mime_type,notnull"`
	Size       int64      `bun:"size,notnull"`
	StorageKey string     `bun:"storage_key,notnull"`
	URL        string     `bun:"url,notnull"`
	Width      int        `bun:"width"`
	Height     int        `bun:"height"`
	ThumbSmKey string     `bun:"thumb_sm_key"`
	ThumbSmURL string     `bun:"thumb_sm_url"`
	ThumbMdKey string     `bun:"thumb_md_key"`
	ThumbMdURL string     `bun:"thumb_md_url"`
	UploaderID string     `bun:"uploader_id,type:uuid"`
	CreatedAt  time.Time  `bun:"created_at,nullzero,default:now()"`
	UpdatedAt  time.Time  `bun:"updated_at,nullzero,default:now()"`
	DeletedAt  *time.Time `bun:"deleted_at,soft_delete,nullzero"`
}

// BunMediaRepo implements domain.MediaFileRepository using uptrace/bun.
type BunMediaRepo struct {
	db *bun.DB
}

func NewBunMediaRepo(db *bun.DB) *BunMediaRepo {
	return &BunMediaRepo{db: db}
}

func (r *BunMediaRepo) Save(ctx context.Context, f *domain.MediaFile) error {
	row := toRow(f)
	if f.ID == "" {
		if _, err := r.db.NewInsert().Model(row).Exec(ctx); err != nil {
			return fmt.Errorf("insert media file: %w", err)
		}
		f.ID = row.ID
		f.CreatedAt = row.CreatedAt
		f.UpdatedAt = row.UpdatedAt
		return nil
	}
	if _, err := r.db.NewUpdate().Model(row).WherePK().Exec(ctx); err != nil {
		return fmt.Errorf("update media file: %w", err)
	}
	return nil
}

func (r *BunMediaRepo) FindByID(ctx context.Context, id string) (*domain.MediaFile, error) {
	row := &mediaFileRow{}
	err := r.db.NewSelect().Model(row).Where("mf.id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find media file: %w", err)
	}
	return toDomain(row), nil
}

func (r *BunMediaRepo) List(ctx context.Context, offset, limit int) ([]*domain.MediaFile, int, error) {
	var rows []mediaFileRow
	count, err := r.db.NewSelect().Model(&rows).
		OrderExpr("created_at DESC").
		Offset(offset).Limit(limit).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list media files: %w", err)
	}
	files := make([]*domain.MediaFile, len(rows))
	for i := range rows {
		files[i] = toDomain(&rows[i])
	}
	return files, count, nil
}

func (r *BunMediaRepo) Delete(ctx context.Context, id string) error {
	row := &mediaFileRow{ID: id}
	if _, err := r.db.NewDelete().Model(row).WherePK().Exec(ctx); err != nil {
		return fmt.Errorf("delete media file: %w", err)
	}
	return nil
}

// toRow maps domain entity → bun model row.
func toRow(f *domain.MediaFile) *mediaFileRow {
	return &mediaFileRow{
		ID:         f.ID,
		Filename:   f.Filename,
		MimeType:   f.MimeType,
		Size:       f.Size,
		StorageKey: f.StorageKey,
		URL:        f.URL,
		Width:      f.Width,
		Height:     f.Height,
		ThumbSmKey: f.ThumbSmKey,
		ThumbSmURL: f.ThumbSmURL,
		ThumbMdKey: f.ThumbMdKey,
		ThumbMdURL: f.ThumbMdURL,
		UploaderID: f.UploaderID,
	}
}

// toDomain maps bun model row → domain entity.
func toDomain(row *mediaFileRow) *domain.MediaFile {
	return &domain.MediaFile{
		ID:         row.ID,
		Filename:   row.Filename,
		MimeType:   row.MimeType,
		Size:       row.Size,
		StorageKey: row.StorageKey,
		URL:        row.URL,
		Width:      row.Width,
		Height:     row.Height,
		ThumbSmKey: row.ThumbSmKey,
		ThumbSmURL: row.ThumbSmURL,
		ThumbMdKey: row.ThumbMdKey,
		ThumbMdURL: row.ThumbMdURL,
		UploaderID: row.UploaderID,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

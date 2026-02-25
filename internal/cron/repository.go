package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/uptrace/bun"
)

// --- SiteLister implementation ---

// BunSiteLister queries active site slugs from sfc_sites.
type BunSiteLister struct {
	db *bun.DB
}

func NewBunSiteLister(db *bun.DB) *BunSiteLister {
	return &BunSiteLister{db: db}
}

func (r *BunSiteLister) ListActiveSlugs(ctx context.Context) ([]string, error) {
	var slugs []string
	err := r.db.NewSelect().
		Model((*model.Site)(nil)).
		Column("slug").
		Where("status = ?", model.SiteStatusActive).
		Where("deleted_at IS NULL").
		Scan(ctx, &slugs)
	if err != nil {
		return nil, fmt.Errorf("list active site slugs: %w", err)
	}
	return slugs, nil
}

// --- SchemaExecutor implementation ---

// BunSchemaExecutor sets/resets search_path using bun.DB.
type BunSchemaExecutor struct {
	db *bun.DB
}

func NewBunSchemaExecutor(db *bun.DB) *BunSchemaExecutor {
	return &BunSchemaExecutor{db: db}
}

func (e *BunSchemaExecutor) SetSearchPath(ctx context.Context, slug string) error {
	_, err := e.db.ExecContext(ctx, fmt.Sprintf("SET search_path TO 'site_%s', 'public'", slug))
	return err
}

func (e *BunSchemaExecutor) ResetSearchPath(ctx context.Context) error {
	_, err := e.db.ExecContext(ctx, "SET search_path TO 'public'")
	return err
}

// --- ScheduledPublisher implementation ---

// BunScheduledPublisher publishes posts whose scheduled_at has arrived.
type BunScheduledPublisher struct {
	db *bun.DB
}

func NewBunScheduledPublisher(db *bun.DB) *BunScheduledPublisher {
	return &BunScheduledPublisher{db: db}
}

func (r *BunScheduledPublisher) PublishDue(ctx context.Context) (int64, error) {
	now := time.Now()
	res, err := r.db.NewUpdate().
		Model((*model.Post)(nil)).
		Set("status = ?", model.PostStatusPublished).
		Set("published_at = ?", now).
		Set("updated_at = ?", now).
		Where("status = ?", model.PostStatusScheduled).
		Where("scheduled_at <= ?", now).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("publish scheduled posts: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// --- TokenCleaner implementation ---

// BunTokenCleaner deletes expired preview tokens.
type BunTokenCleaner struct {
	db *bun.DB
}

func NewBunTokenCleaner(db *bun.DB) *BunTokenCleaner {
	return &BunTokenCleaner{db: db}
}

func (r *BunTokenCleaner) DeleteExpired(ctx context.Context) (int64, error) {
	res, err := r.db.NewDelete().
		Model((*model.PreviewToken)(nil)).
		Where("expires_at < ?", time.Now()).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("delete expired preview tokens: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// --- SoftDeletePurger implementation ---

// BunSoftDeletePurger permanently removes old soft-deleted records.
type BunSoftDeletePurger struct {
	db *bun.DB
}

func NewBunSoftDeletePurger(db *bun.DB) *BunSoftDeletePurger {
	return &BunSoftDeletePurger{db: db}
}

func (r *BunSoftDeletePurger) PurgeOlderThan(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	var total int64

	// Purge posts (force delete bypasses soft_delete hook)
	res, err := r.db.NewDelete().
		Model((*model.Post)(nil)).
		ForceDelete().
		Where("deleted_at IS NOT NULL").
		Where("deleted_at < ?", cutoff).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("purge posts: %w", err)
	}
	n, _ := res.RowsAffected()
	total += n

	// Purge media files
	res, err = r.db.NewDelete().
		Model((*model.MediaFile)(nil)).
		ForceDelete().
		Where("deleted_at IS NOT NULL").
		Where("deleted_at < ?", cutoff).
		Exec(ctx)
	if err != nil {
		return total, fmt.Errorf("purge media: %w", err)
	}
	n, _ = res.RowsAffected()
	total += n

	// Purge comments
	res, err = r.db.NewDelete().
		Model((*model.Comment)(nil)).
		ForceDelete().
		Where("deleted_at IS NOT NULL").
		Where("deleted_at < ?", cutoff).
		Exec(ctx)
	if err != nil {
		return total, fmt.Errorf("purge comments: %w", err)
	}
	n, _ = res.RowsAffected()
	total += n

	return total, nil
}

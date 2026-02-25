package cron

import (
	"context"
)

// SiteLister retrieves all active site slugs for cross-schema iteration.
type SiteLister interface {
	ListActiveSlugs(ctx context.Context) ([]string, error)
}

// ScheduledPublisher publishes posts whose scheduled time has arrived.
type ScheduledPublisher interface {
	// PublishDue finds posts with status=scheduled and scheduled_at <= now,
	// updates them to published status, and returns the count of published posts.
	PublishDue(ctx context.Context) (int64, error)
}

// TokenCleaner removes expired preview tokens.
type TokenCleaner interface {
	// DeleteExpired removes all preview tokens with expires_at < now,
	// returning the count of deleted tokens.
	DeleteExpired(ctx context.Context) (int64, error)
}

// SoftDeletePurger permanently removes soft-deleted records older than the retention period.
type SoftDeletePurger interface {
	// PurgeOlderThan permanently deletes soft-deleted records where
	// deleted_at < (now - retentionDays). Returns total count of purged records.
	PurgeOlderThan(ctx context.Context, retentionDays int) (int64, error)
}

// SchemaExecutor sets the search_path for a given site slug.
type SchemaExecutor interface {
	SetSearchPath(ctx context.Context, slug string) error
	ResetSearchPath(ctx context.Context) error
}

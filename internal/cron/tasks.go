package cron

import (
	"context"
	"fmt"
	"log/slog"
)

const defaultRetentionDays = 30

// RunScheduledPublishing iterates all active sites and publishes posts
// whose scheduled_at time has arrived.
func RunScheduledPublishing(ctx context.Context, sites SiteLister, schema SchemaExecutor, pub ScheduledPublisher) (int64, error) {
	slugs, err := sites.ListActiveSlugs(ctx)
	if err != nil {
		return 0, fmt.Errorf("list active sites: %w", err)
	}

	var total int64
	for _, slug := range slugs {
		if err := schema.SetSearchPath(ctx, slug); err != nil {
			slog.Error("cron: set search_path failed", "site", slug, "error", err)
			continue
		}

		n, err := pub.PublishDue(ctx)
		if err != nil {
			slog.Error("cron: publish scheduled posts failed", "site", slug, "error", err)
			schema.ResetSearchPath(ctx) //nolint:errcheck
			continue
		}

		total += n
		if n > 0 {
			slog.Info("cron: published scheduled posts", "site", slug, "count", n)
		}

		schema.ResetSearchPath(ctx) //nolint:errcheck
	}
	return total, nil
}

// RunTokenCleanup iterates all active sites and removes expired preview tokens.
func RunTokenCleanup(ctx context.Context, sites SiteLister, schema SchemaExecutor, cleaner TokenCleaner) (int64, error) {
	slugs, err := sites.ListActiveSlugs(ctx)
	if err != nil {
		return 0, fmt.Errorf("list active sites: %w", err)
	}

	var total int64
	for _, slug := range slugs {
		if err := schema.SetSearchPath(ctx, slug); err != nil {
			slog.Error("cron: set search_path failed", "site", slug, "error", err)
			continue
		}

		n, err := cleaner.DeleteExpired(ctx)
		if err != nil {
			slog.Error("cron: clean expired tokens failed", "site", slug, "error", err)
			schema.ResetSearchPath(ctx) //nolint:errcheck
			continue
		}

		total += n
		if n > 0 {
			slog.Info("cron: cleaned expired preview tokens", "site", slug, "count", n)
		}

		schema.ResetSearchPath(ctx) //nolint:errcheck
	}
	return total, nil
}

// RunSoftDeletePurge iterates all active sites and permanently removes
// soft-deleted records older than the retention period.
func RunSoftDeletePurge(ctx context.Context, sites SiteLister, schema SchemaExecutor, purger SoftDeletePurger, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = defaultRetentionDays
	}

	slugs, err := sites.ListActiveSlugs(ctx)
	if err != nil {
		return 0, fmt.Errorf("list active sites: %w", err)
	}

	var total int64
	for _, slug := range slugs {
		if err := schema.SetSearchPath(ctx, slug); err != nil {
			slog.Error("cron: set search_path failed", "site", slug, "error", err)
			continue
		}

		n, err := purger.PurgeOlderThan(ctx, retentionDays)
		if err != nil {
			slog.Error("cron: purge soft-deleted records failed", "site", slug, "error", err)
			schema.ResetSearchPath(ctx) //nolint:errcheck
			continue
		}

		total += n
		if n > 0 {
			slog.Info("cron: purged soft-deleted records", "site", slug, "count", n)
		}

		schema.ResetSearchPath(ctx) //nolint:errcheck
	}
	return total, nil
}

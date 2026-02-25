package cron

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Deps holds the dependencies required by the cron scheduler.
type Deps struct {
	Sites     SiteLister
	Schema    SchemaExecutor
	Publisher ScheduledPublisher
	Cleaner   TokenCleaner
	Purger    SoftDeletePurger
}

// Option configures the Scheduler.
type Option func(*Scheduler)

// Scheduler runs periodic cron tasks across all site schemas.
type Scheduler struct {
	deps Deps

	publishInterval      time.Duration
	tokenCleanupInterval time.Duration
	purgeInterval        time.Duration
	retentionDays        int

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewScheduler creates a Scheduler with the given dependencies and options.
func NewScheduler(deps Deps, opts ...Option) *Scheduler {
	s := &Scheduler{
		deps:                 deps,
		publishInterval:      1 * time.Minute,
		tokenCleanupInterval: 1 * time.Hour,
		purgeInterval:        24 * time.Hour,
		retentionDays:        30,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithPublishInterval(d time.Duration) Option      { return func(s *Scheduler) { s.publishInterval = d } }
func WithTokenCleanupInterval(d time.Duration) Option  { return func(s *Scheduler) { s.tokenCleanupInterval = d } }
func WithPurgeInterval(d time.Duration) Option         { return func(s *Scheduler) { s.purgeInterval = d } }
func WithRetentionDays(days int) Option                { return func(s *Scheduler) { s.retentionDays = days } }

// Start launches the cron goroutines. Call Stop to shut them down.
func (s *Scheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.wg.Add(3)
	go s.loop(ctx, "scheduled-publish", s.publishInterval, func(ctx context.Context) {
		RunScheduledPublishing(ctx, s.deps.Sites, s.deps.Schema, s.deps.Publisher) //nolint:errcheck
	})
	go s.loop(ctx, "token-cleanup", s.tokenCleanupInterval, func(ctx context.Context) {
		RunTokenCleanup(ctx, s.deps.Sites, s.deps.Schema, s.deps.Cleaner) //nolint:errcheck
	})
	go s.loop(ctx, "soft-delete-purge", s.purgeInterval, func(ctx context.Context) {
		RunSoftDeletePurge(ctx, s.deps.Sites, s.deps.Schema, s.deps.Purger, s.retentionDays) //nolint:errcheck
	})

	slog.Info("cron: scheduler started",
		"publish_interval", s.publishInterval,
		"token_cleanup_interval", s.tokenCleanupInterval,
		"purge_interval", s.purgeInterval,
	)
}

// Stop gracefully shuts down all cron goroutines.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	slog.Info("cron: scheduler stopped")
}

func (s *Scheduler) loop(ctx context.Context, name string, interval time.Duration, fn func(context.Context)) {
	defer s.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn(ctx)
		}
	}
}

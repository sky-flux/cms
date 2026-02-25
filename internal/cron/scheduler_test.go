package cron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewScheduler_DefaultIntervals(t *testing.T) {
	s := NewScheduler(Deps{})

	assert.Equal(t, 1*time.Minute, s.publishInterval)
	assert.Equal(t, 1*time.Hour, s.tokenCleanupInterval)
	assert.Equal(t, 24*time.Hour, s.purgeInterval)
	assert.Equal(t, 30, s.retentionDays)
}

func TestNewScheduler_CustomIntervals(t *testing.T) {
	s := NewScheduler(Deps{},
		WithPublishInterval(30*time.Second),
		WithTokenCleanupInterval(10*time.Minute),
		WithPurgeInterval(12*time.Hour),
		WithRetentionDays(7),
	)

	assert.Equal(t, 30*time.Second, s.publishInterval)
	assert.Equal(t, 10*time.Minute, s.tokenCleanupInterval)
	assert.Equal(t, 12*time.Hour, s.purgeInterval)
	assert.Equal(t, 7, s.retentionDays)
}

func TestScheduler_StartStop(t *testing.T) {
	s := NewScheduler(Deps{
		Sites:     &mockSiteLister{slugs: []string{}},
		Schema:    &mockSchemaExecutor{},
		Publisher: &mockScheduledPublisher{},
		Cleaner:   &mockTokenCleaner{},
		Purger:    &mockSoftDeletePurger{},
	},
		WithPublishInterval(10*time.Millisecond),
		WithTokenCleanupInterval(10*time.Millisecond),
		WithPurgeInterval(10*time.Millisecond),
	)

	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()

	// Should not panic or hang. If Stop blocks, the test will time out.
}

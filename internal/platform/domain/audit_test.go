package domain_test

import (
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/platform/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditEntry_Valid(t *testing.T) {
	entry, err := domain.NewAuditEntry(
		"user-123",
		domain.AuditActionCreate,
		"post",
		"post-456",
		"127.0.0.1",
		"Mozilla/5.0",
	)
	require.NoError(t, err)
	assert.Equal(t, "user-123", entry.UserID)
	assert.Equal(t, domain.AuditActionCreate, entry.Action)
	assert.Equal(t, "post", entry.Resource)
	assert.Equal(t, "post-456", entry.ResourceID)
	assert.Equal(t, "127.0.0.1", entry.IP)
	assert.Equal(t, "Mozilla/5.0", entry.UserAgent)
	assert.False(t, entry.CreatedAt.IsZero())
}

func TestNewAuditEntry_EmptyUserID_ReturnsError(t *testing.T) {
	_, err := domain.NewAuditEntry("", domain.AuditActionCreate, "post", "post-456", "", "")
	assert.ErrorIs(t, err, domain.ErrEmptyUserID)
}

func TestNewAuditEntry_EmptyResource_ReturnsError(t *testing.T) {
	_, err := domain.NewAuditEntry("user-123", domain.AuditActionCreate, "", "post-456", "", "")
	assert.ErrorIs(t, err, domain.ErrEmptyResource)
}

func TestAuditEntry_CreatedAt_IsRecent(t *testing.T) {
	before := time.Now()
	entry, _ := domain.NewAuditEntry("u1", domain.AuditActionDelete, "media", "m1", "", "")
	after := time.Now()
	assert.True(t, entry.CreatedAt.After(before) || entry.CreatedAt.Equal(before))
	assert.True(t, entry.CreatedAt.Before(after) || entry.CreatedAt.Equal(after))
}

package audit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sky-flux/cms/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time check: Service must satisfy the Logger interface.
var _ Logger = (*Service)(nil)

func TestCtxValue(t *testing.T) {
	t.Run("returns string value", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", "abc-123")
		assert.Equal(t, "abc-123", ctxValue(ctx, "user_id"))
	})

	t.Run("returns empty for missing key", func(t *testing.T) {
		ctx := context.Background()
		assert.Equal(t, "", ctxValue(ctx, "user_id"))
	})

	t.Run("returns empty for non-string value", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", 42)
		assert.Equal(t, "", ctxValue(ctx, "user_id"))
	})
}

func TestEntrySnapshotMarshal(t *testing.T) {
	t.Run("nil snapshot produces nil json", func(t *testing.T) {
		entry := Entry{
			Action:       model.LogActionCreate,
			ResourceType: "post",
			ResourceID:   "post-1",
		}

		// Simulate the marshaling logic from Log()
		var snapshot json.RawMessage
		if entry.ResourceSnapshot != nil {
			data, err := json.Marshal(entry.ResourceSnapshot)
			require.NoError(t, err)
			snapshot = data
		}
		assert.Nil(t, snapshot)
	})

	t.Run("struct snapshot marshals to JSON", func(t *testing.T) {
		type postSnap struct {
			Title  string `json:"title"`
			Status string `json:"status"`
		}
		entry := Entry{
			Action:       model.LogActionUpdate,
			ResourceType: "post",
			ResourceID:   "post-2",
			ResourceSnapshot: postSnap{
				Title:  "Hello World",
				Status: "published",
			},
		}

		data, err := json.Marshal(entry.ResourceSnapshot)
		require.NoError(t, err)

		var parsed map[string]string
		require.NoError(t, json.Unmarshal(data, &parsed))
		assert.Equal(t, "Hello World", parsed["title"])
		assert.Equal(t, "published", parsed["status"])
	})

	t.Run("map snapshot marshals to JSON", func(t *testing.T) {
		entry := Entry{
			Action:       model.LogActionDelete,
			ResourceType: "media",
			ResourceID:   "media-1",
			ResourceSnapshot: map[string]any{
				"filename": "photo.jpg",
				"size":     1024,
			},
		}

		data, err := json.Marshal(entry.ResourceSnapshot)
		require.NoError(t, err)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal(data, &parsed))
		assert.Equal(t, "photo.jpg", parsed["filename"])
		assert.InDelta(t, 1024, parsed["size"], 0)
	})
}

func TestBuildAuditRecord(t *testing.T) {
	// Test that context values are correctly extracted and applied to the record.
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_id", "uid-abc")
	ctx = context.WithValue(ctx, "user_email", "admin@example.com")
	ctx = context.WithValue(ctx, "audit_ip", "192.168.1.1")
	ctx = context.WithValue(ctx, "audit_ua", "Mozilla/5.0")

	entry := Entry{
		Action:       model.LogActionPublish,
		ResourceType: "post",
		ResourceID:   "post-99",
		ResourceSnapshot: map[string]string{
			"title": "Published Post",
		},
	}

	// Replicate the record-building logic from Log() without the DB call.
	var snapshot json.RawMessage
	if entry.ResourceSnapshot != nil {
		data, err := json.Marshal(entry.ResourceSnapshot)
		require.NoError(t, err)
		snapshot = data
	}

	record := &model.Audit{
		Action:           entry.Action,
		ResourceType:     entry.ResourceType,
		ResourceID:       entry.ResourceID,
		ResourceSnapshot: snapshot,
	}

	if v := ctxValue(ctx, "user_id"); v != "" {
		record.ActorID = &v
	}
	if v := ctxValue(ctx, "user_email"); v != "" {
		record.ActorEmail = v
	}
	if v := ctxValue(ctx, "audit_ip"); v != "" {
		record.IPAddress = v
	}
	if v := ctxValue(ctx, "audit_ua"); v != "" {
		record.UserAgent = v
	}

	assert.Equal(t, model.LogActionPublish, record.Action)
	assert.Equal(t, "post", record.ResourceType)
	assert.Equal(t, "post-99", record.ResourceID)
	require.NotNil(t, record.ActorID)
	assert.Equal(t, "uid-abc", *record.ActorID)
	assert.Equal(t, "admin@example.com", record.ActorEmail)
	assert.Equal(t, "192.168.1.1", record.IPAddress)
	assert.Equal(t, "Mozilla/5.0", record.UserAgent)
	assert.Contains(t, string(record.ResourceSnapshot), "Published Post")
}

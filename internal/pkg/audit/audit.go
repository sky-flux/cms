package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/uptrace/bun"
)

// Logger is the interface for audit logging. Other packages depend on this
// interface rather than the concrete Service, making it easy to mock in tests.
type Logger interface {
	Log(ctx context.Context, entry Entry) error
}

// Entry holds the data for a single audit log record.
type Entry struct {
	Action           model.LogAction
	ResourceType     string
	ResourceID       string
	ResourceSnapshot any
}

// Service implements Logger by inserting records into sfc_site_audits.
// The table lives in the site_{slug} schema; search_path is set by
// SchemaMiddleware before the request reaches the handler.
type Service struct {
	db *bun.DB
}

// NewService creates a new audit Service.
func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

// Log persists an audit entry. Actor metadata (user_id, email, IP, UA) is
// extracted from context values set by the auth and audit middleware.
func (s *Service) Log(ctx context.Context, entry Entry) error {
	var snapshot json.RawMessage
	if entry.ResourceSnapshot != nil {
		data, err := json.Marshal(entry.ResourceSnapshot)
		if err != nil {
			return fmt.Errorf("marshal snapshot: %w", err)
		}
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

	_, err := s.db.NewInsert().Model(record).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

// NoopLogger is a no-op implementation of Logger for use in tests.
type NoopLogger struct{}

func NewNoopLogger() *NoopLogger { return &NoopLogger{} }

func (n *NoopLogger) Log(_ context.Context, _ Entry) error { return nil }

// ctxValue extracts a string value from context. Returns "" if the key is
// missing or the value is not a string.
func ctxValue(ctx context.Context, key string) string {
	if v := ctx.Value(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

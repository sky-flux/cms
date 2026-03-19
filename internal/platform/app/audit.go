package app

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/platform/domain"
)

// LogAuditInput is the input DTO for the LogAuditUseCase.
type LogAuditInput struct {
	UserID     string
	Action     domain.AuditAction
	Resource   string
	ResourceID string
	IP         string
	UserAgent  string
}

// LogAuditUseCase creates and persists a single audit log entry.
type LogAuditUseCase struct {
	repo domain.AuditRepository
}

// NewLogAuditUseCase creates a LogAuditUseCase.
func NewLogAuditUseCase(repo domain.AuditRepository) *LogAuditUseCase {
	return &LogAuditUseCase{repo: repo}
}

// Execute validates inputs via the domain, then persists the entry.
// It is safe to call in a goroutine for fire-and-forget audit logging.
func (uc *LogAuditUseCase) Execute(ctx context.Context, in LogAuditInput) error {
	entry, err := domain.NewAuditEntry(
		in.UserID,
		in.Action,
		in.Resource,
		in.ResourceID,
		in.IP,
		in.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("invalid audit input: %w", err)
	}
	if err := uc.repo.Save(ctx, entry); err != nil {
		return fmt.Errorf("save audit entry: %w", err)
	}
	return nil
}

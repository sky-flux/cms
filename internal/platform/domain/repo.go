package domain

import (
	"context"
	"time"
)

// AuditFilter controls pagination and filtering for AuditRepository.List.
type AuditFilter struct {
	Page      int
	PerPage   int
	UserID    string       // filter by actor (empty = all users)
	Resource  string       // filter by resource type (empty = all)
	Action    *AuditAction // nil = all actions
	StartDate *time.Time
	EndDate   *time.Time
}

// AuditRepository is the persistence port for audit log entries.
// The domain defines the interface; infra/ provides the bun implementation.
type AuditRepository interface {
	// Save persists a new audit entry. The DB sets ID via uuidv7().
	Save(ctx context.Context, e *AuditEntry) error
	// List returns paginated audit entries matching the filter.
	List(ctx context.Context, f AuditFilter) ([]AuditEntry, int64, error)
}

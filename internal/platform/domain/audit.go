package domain

import (
	"errors"
	"time"
)

// Sentinel errors for AuditEntry validation.
var (
	ErrEmptyUserID   = errors.New("audit entry: userID must not be empty")
	ErrEmptyResource = errors.New("audit entry: resource must not be empty")
)

// AuditAction is a typed constant for the kind of operation logged.
type AuditAction int8

const (
	AuditActionCreate   AuditAction = 1
	AuditActionUpdate   AuditAction = 2
	AuditActionDelete   AuditAction = 3
	AuditActionPublish  AuditAction = 4
	AuditActionArchive  AuditAction = 5
	AuditActionLogin    AuditAction = 6
	AuditActionLogout   AuditAction = 7
	AuditActionRestore  AuditAction = 8
	AuditActionApprove  AuditAction = 9
	AuditActionReject   AuditAction = 10
	AuditActionGenerate AuditAction = 11
)

// AuditEntry is the aggregate for a single immutable audit log record.
// Once created it is never mutated; all fields are set at construction time.
type AuditEntry struct {
	ID         string      // set by the DB via uuidv7()
	UserID     string
	Action     AuditAction
	Resource   string // e.g. "post", "user", "media"
	ResourceID string // UUID of the affected resource
	IP         string
	UserAgent  string
	CreatedAt  time.Time
}

// NewAuditEntry validates inputs and constructs an AuditEntry ready for persistence.
// CreatedAt is set to time.Now() in UTC; ID is left empty for the DB to assign.
func NewAuditEntry(
	userID string,
	action AuditAction,
	resource, resourceID string,
	ip, userAgent string,
) (*AuditEntry, error) {
	if userID == "" {
		return nil, ErrEmptyUserID
	}
	if resource == "" {
		return nil, ErrEmptyResource
	}
	return &AuditEntry{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IP:         ip,
		UserAgent:  userAgent,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

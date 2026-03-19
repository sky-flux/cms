// Package infra contains the infrastructure layer for the Platform BC:
// bun ORM implementations of domain repository interfaces.
package infra

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sky-flux/cms/internal/platform/domain"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	_ "github.com/uptrace/bun/driver/pgdriver"
)

// auditRecord is the bun ORM model for sfc_audits.
type auditRecord struct {
	bun.BaseModel `bun:"table:sfc_audits,alias:a"`

	ID         string             `bun:"id,pk,type:uuid,default:uuidv7()"`
	UserID     string             `bun:"user_id,type:uuid,notnull"`
	Action     domain.AuditAction `bun:"action,notnull"`
	Resource   string             `bun:"resource,notnull"`
	ResourceID string             `bun:"resource_id,type:uuid"`
	IP         string             `bun:"ip"`
	UserAgent  string             `bun:"user_agent"`
	CreatedAt  time.Time          `bun:"created_at,notnull,default:now()"`
}

// BunAuditRepository implements domain.AuditRepository using uptrace/bun.
type BunAuditRepository struct {
	db *bun.DB
}

// NewBunAuditRepository creates a new BunAuditRepository.
func NewBunAuditRepository(db *bun.DB) *BunAuditRepository {
	return &BunAuditRepository{db: db}
}

// OpenBunDB opens a bun.DB connection to the given PostgreSQL DSN.
func OpenBunDB(dsn string) (*bun.DB, error) {
	sqldb, err := sql.Open("pg", dsn)
	if err != nil {
		return nil, fmt.Errorf("open pg driver: %w", err)
	}
	db := bun.NewDB(sqldb, pgdialect.New(), bun.WithDiscardUnknownColumns())
	return db, nil
}

// Save inserts a new audit entry into sfc_audits.
func (r *BunAuditRepository) Save(ctx context.Context, e *domain.AuditEntry) error {
	rec := &auditRecord{
		UserID:     e.UserID,
		Action:     e.Action,
		Resource:   e.Resource,
		ResourceID: e.ResourceID,
		IP:         e.IP,
		UserAgent:  e.UserAgent,
	}
	_, err := r.db.NewInsert().Model(rec).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert audit record: %w", err)
	}
	e.ID = rec.ID
	return nil
}

// List returns paginated audit entries filtered by AuditFilter.
func (r *BunAuditRepository) List(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 {
		f.PerPage = 20
	}
	if f.PerPage > 100 {
		f.PerPage = 100
	}

	var recs []auditRecord
	q := r.db.NewSelect().Model(&recs).OrderExpr("a.created_at DESC").
		Limit(f.PerPage).Offset((f.Page - 1) * f.PerPage)

	if f.UserID != "" {
		q = q.Where("a.user_id = ?", f.UserID)
	}
	if f.Resource != "" {
		q = q.Where("a.resource = ?", f.Resource)
	}
	if f.Action != nil {
		q = q.Where("a.action = ?", *f.Action)
	}
	if f.StartDate != nil {
		q = q.Where("a.created_at >= ?", f.StartDate)
	}
	if f.EndDate != nil {
		q = q.Where("a.created_at <= ?", f.EndDate)
	}

	total, err := q.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit records: %w", err)
	}

	entries := make([]domain.AuditEntry, len(recs))
	for i, rec := range recs {
		entries[i] = domain.AuditEntry{
			ID:         rec.ID,
			UserID:     rec.UserID,
			Action:     rec.Action,
			Resource:   rec.Resource,
			ResourceID: rec.ResourceID,
			IP:         rec.IP,
			UserAgent:  rec.UserAgent,
			CreatedAt:  rec.CreatedAt,
		}
	}
	return entries, int64(total), nil
}

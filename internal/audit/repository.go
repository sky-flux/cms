package audit

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

type AuditRepo struct {
	db *bun.DB
}

func NewAuditRepo(db *bun.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) List(ctx context.Context, f ListFilter) ([]AuditWithActor, int64, error) {
	var items []AuditWithActor
	q := r.db.NewSelect().
		TableExpr("sfc_site_audits AS a").
		ColumnExpr("a.*").
		ColumnExpr("u.display_name AS actor_display_name").
		Join("LEFT JOIN public.sfc_users AS u ON a.actor_id = u.id")

	if f.ActorID != "" {
		q = q.Where("a.actor_id = ?", f.ActorID)
	}
	if f.Action != nil {
		q = q.Where("a.action = ?", *f.Action)
	}
	if f.ResourceType != "" {
		q = q.Where("a.resource_type = ?", f.ResourceType)
	}
	if f.StartDate != nil {
		q = q.Where("a.created_at >= ?", *f.StartDate)
	}
	if f.EndDate != nil {
		q = q.Where("a.created_at <= ?", *f.EndDate)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("audit list count: %w", err)
	}

	offset := (f.Page - 1) * f.PerPage
	if offset < 0 {
		offset = 0
	}

	err = q.OrderExpr("a.created_at DESC").
		Limit(f.PerPage).
		Offset(offset).
		Scan(ctx, &items)
	if err != nil {
		return nil, 0, fmt.Errorf("audit list: %w", err)
	}
	return items, int64(total), nil
}

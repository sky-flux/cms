package audit

import (
	"context"
)

// AuditRepository handles read-only queries on sfc_site_audits.
type AuditRepository interface {
	List(ctx context.Context, f ListFilter) ([]AuditWithActor, int64, error)
}

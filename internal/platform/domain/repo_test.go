package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/platform/domain"
)

// Compile-time interface satisfaction check.
var _ domain.AuditRepository = (*mockAuditRepo)(nil)

type mockAuditRepo struct {
	saveFn func(ctx context.Context, e *domain.AuditEntry) error
	listFn func(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, int64, error)
}

func (m *mockAuditRepo) Save(ctx context.Context, e *domain.AuditEntry) error {
	return m.saveFn(ctx, e)
}
func (m *mockAuditRepo) List(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, int64, error) {
	return m.listFn(ctx, f)
}

func TestAuditRepository_Interface(t *testing.T) {
	t.Log("AuditRepository interface satisfied by mockAuditRepo")
}

func TestAuditFilter_ZeroValue_IsValid(t *testing.T) {
	// A zero-value AuditFilter must be usable (no required fields).
	var f domain.AuditFilter
	if f.Page == 0 {
		f.Page = 1
	}
	if f.PerPage == 0 {
		f.PerPage = 20
	}
	_ = f
}

func TestAuditFilter_WithDateRange(t *testing.T) {
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	f := domain.AuditFilter{
		StartDate: &start,
		EndDate:   &end,
	}
	_ = f
}

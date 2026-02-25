package redirect

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// RedirectRepository handles sfc_site_redirects table operations.
type RedirectRepository interface {
	List(ctx context.Context, filter ListFilter) ([]model.Redirect, int64, error)
	GetByID(ctx context.Context, id string) (*model.Redirect, error)
	Create(ctx context.Context, redirect *model.Redirect) error
	Update(ctx context.Context, redirect *model.Redirect) error
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) (int64, error)
	SourcePathExists(ctx context.Context, path string, excludeID string) (bool, error)
	BulkInsert(ctx context.Context, redirects []*model.Redirect) (int64, error)
	ListAll(ctx context.Context) ([]model.Redirect, error)
}

// ListFilter holds pagination and filtering options for redirect listing.
type ListFilter struct {
	Page       int
	PerPage    int
	Query      string // search source_path or target_url
	StatusCode int    // 301 or 302
	Status     string // "active" or "disabled"
	Sort       string // "created_at:desc", "hit_count:desc", "last_hit_at:desc", "source_path:asc"
}

package tag

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// TagRepository handles sfc_site_tags table CRUD.
type TagRepository interface {
	List(ctx context.Context, filter ListFilter) ([]model.Tag, int64, error)
	GetByID(ctx context.Context, id string) (*model.Tag, error)
	Create(ctx context.Context, tag *model.Tag) error
	Update(ctx context.Context, tag *model.Tag) error
	Delete(ctx context.Context, id string) error
	SlugExists(ctx context.Context, slug string, excludeID string) (bool, error)
	NameExists(ctx context.Context, name string, excludeID string) (bool, error)
	CountPosts(ctx context.Context, tagID string) (int64, error)
}

// ListFilter holds pagination and filtering options for tag listing.
type ListFilter struct {
	Page    int
	PerPage int
	Query   string
	Sort    string // "name_asc", "name_desc", "created_asc", "created_desc"
}

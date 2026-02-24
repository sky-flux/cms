package posttype

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// PostTypeRepository handles sfc_site_post_types table CRUD.
type PostTypeRepository interface {
	List(ctx context.Context) ([]model.PostType, error)
	GetByID(ctx context.Context, id string) (*model.PostType, error)
	GetBySlug(ctx context.Context, slug string) (*model.PostType, error)
	Create(ctx context.Context, pt *model.PostType) error
	Update(ctx context.Context, pt *model.PostType) error
	Delete(ctx context.Context, id string) error
}

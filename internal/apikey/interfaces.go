package apikey

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// APIKeyRepository handles sfc_site_api_keys table CRUD.
type APIKeyRepository interface {
	List(ctx context.Context) ([]model.APIKey, error)
	GetByID(ctx context.Context, id string) (*model.APIKey, error)
	Create(ctx context.Context, key *model.APIKey) error
	Revoke(ctx context.Context, id string) error
}

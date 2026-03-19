package domain

import "context"

// CategoryRepository is the persistence port for the Category aggregate.
type CategoryRepository interface {
	Save(ctx context.Context, c *Category) error
	FindByID(ctx context.Context, id string) (*Category, error)
	SlugExists(ctx context.Context, slug, excludeID string) (bool, error)
	FindAncestorIDs(ctx context.Context, id string) ([]string, error)
	List(ctx context.Context) ([]*Category, error)
	SoftDelete(ctx context.Context, id string) error
	Update(ctx context.Context, c *Category) error
}

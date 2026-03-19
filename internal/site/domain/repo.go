package domain

import "context"

// SiteRepository is the persistence port for site configuration.
// v1: single site record, so no ID parameter — Get returns the sole record, Upsert creates or updates it.
type SiteRepository interface {
	GetSite(ctx context.Context) (*Site, error)
	Upsert(ctx context.Context, s *Site) error
}

// MenuRepository is the persistence port for navigation menus and their items.
type MenuRepository interface {
	Save(ctx context.Context, m *Menu) error
	FindByID(ctx context.Context, id string) (*Menu, error)
	List(ctx context.Context) ([]*Menu, error)
	Delete(ctx context.Context, id string) error

	SaveItem(ctx context.Context, item *MenuItem) error
	ListItems(ctx context.Context, menuID string) ([]*MenuItem, error)
	DeleteItem(ctx context.Context, itemID string) error
}

// RedirectRepository is the persistence port for URL redirects.
type RedirectRepository interface {
	Save(ctx context.Context, r *Redirect) error
	FindByPath(ctx context.Context, fromPath string) (*Redirect, error)
	List(ctx context.Context, offset, limit int) ([]*Redirect, int, error)
	Delete(ctx context.Context, id string) error
}

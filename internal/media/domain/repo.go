package domain

import "context"

// MediaFileRepository is the persistence port for the Media BC.
type MediaFileRepository interface {
	Save(ctx context.Context, f *MediaFile) error
	FindByID(ctx context.Context, id string) (*MediaFile, error)
	List(ctx context.Context, offset, limit int) ([]*MediaFile, int, error)
	Delete(ctx context.Context, id string) error
}

// StoragePort is the object-storage port for the Media BC.
// Implementations: infra/rustfs_adapter.go (wraps pkg/storage.Client).
type StoragePort interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) error
	Delete(ctx context.Context, key string) error
	URL(key string) string
}

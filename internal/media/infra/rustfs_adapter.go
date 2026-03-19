package infra

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/pkg/storage"
)

// RustFSAdapter adapts pkg/storage.Client to domain.StoragePort.
// This is a thin wrapper — all S3 logic stays in pkg/storage.
type RustFSAdapter struct {
	client *storage.Client
	cdnURL string
}

// NewRustFSAdapter creates an adapter. client may be nil (storage unavailable).
func NewRustFSAdapter(client *storage.Client, cdnURL string) *RustFSAdapter {
	return &RustFSAdapter{client: client, cdnURL: cdnURL}
}

func (a *RustFSAdapter) Upload(ctx context.Context, key string, data []byte, contentType string) error {
	if a.client == nil || !a.client.Available() {
		return fmt.Errorf("storage not available")
	}
	return a.client.UploadBytes(ctx, key, data, contentType)
}

func (a *RustFSAdapter) Delete(ctx context.Context, key string) error {
	if a.client == nil || !a.client.Available() {
		return nil // no-op: mirrors pkg/storage.Client.Delete(nil) behavior
	}
	return a.client.Delete(ctx, key)
}

func (a *RustFSAdapter) URL(key string) string {
	return a.cdnURL + "/" + key
}

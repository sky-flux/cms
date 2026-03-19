package infra_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/media/domain"
	"github.com/sky-flux/cms/internal/media/infra"
	"github.com/stretchr/testify/assert"
)

// Compile-time: RustFSAdapter must satisfy StoragePort.
var _ domain.StoragePort = (*infra.RustFSAdapter)(nil)

func TestRustFSAdapter_URL(t *testing.T) {
	adapter := infra.NewRustFSAdapter(nil, "http://cdn.example.com")
	assert.Equal(t, "http://cdn.example.com/media/2026/03/abc.jpg", adapter.URL("media/2026/03/abc.jpg"))
}

func TestRustFSAdapter_Upload_NilClient(t *testing.T) {
	// When pkg/storage.Client is nil (unavailable), Upload returns an error.
	adapter := infra.NewRustFSAdapter(nil, "http://cdn")
	err := adapter.Upload(context.Background(), "key", []byte("data"), "image/jpeg")
	assert.Error(t, err)
}

func TestRustFSAdapter_Delete_NilClient(t *testing.T) {
	// Nil client: Delete should be a no-op (not error) — mirrors pkg/storage.Client.Delete behavior.
	adapter := infra.NewRustFSAdapter(nil, "http://cdn")
	err := adapter.Delete(context.Background(), "key")
	assert.NoError(t, err)
}

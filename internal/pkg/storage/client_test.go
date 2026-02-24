package storage_test

import (
	"testing"

	"github.com/sky-flux/cms/internal/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewClient_NilS3(t *testing.T) {
	c := storage.NewClient(nil, "test", "http://localhost:9000/test")
	assert.False(t, c.Available())
}

func TestPublicURL(t *testing.T) {
	c := storage.NewClient(nil, "cms-media", "http://localhost:9000/cms-media")
	url := c.PublicURL("media/2026/02/abc.jpg")
	assert.Equal(t, "http://localhost:9000/cms-media/media/2026/02/abc.jpg", url)
}

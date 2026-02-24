package search_test

import (
	"testing"

	"github.com/sky-flux/cms/internal/pkg/search"
	"github.com/stretchr/testify/assert"
)

func TestNewClient_NilMS_Available(t *testing.T) {
	c := search.NewClient(nil)
	assert.False(t, c.Available())
}

func TestNewClient_NilMS_GracefulDegradation(t *testing.T) {
	c := search.NewClient(nil)

	// All operations should be no-ops when ms is nil.
	err := c.EnsureIndex(t.Context(), "test-index", nil)
	assert.NoError(t, err)

	err = c.UpsertDocuments(t.Context(), "test-index", []map[string]any{{"id": "1"}})
	assert.NoError(t, err)

	err = c.DeleteDocuments(t.Context(), "test-index", []string{"1"})
	assert.NoError(t, err)

	result, err := c.Search(t.Context(), "test-index", "query", nil)
	assert.NoError(t, err)
	assert.Empty(t, result.Hits)
}

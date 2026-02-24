package post

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSlug_BasicEnglish(t *testing.T) {
	assert.Equal(t, "my-first-post", GenerateSlug("My First Post"))
}

func TestGenerateSlug_SpecialChars(t *testing.T) {
	assert.Equal(t, "hello-world-2024", GenerateSlug("Hello, World! 2024"))
}

func TestGenerateSlug_LeadingTrailingSpaces(t *testing.T) {
	assert.Equal(t, "trimmed", GenerateSlug("  trimmed  "))
}

func TestGenerateSlug_MultipleDashes(t *testing.T) {
	assert.Equal(t, "a-b-c", GenerateSlug("a---b---c"))
}

func TestGenerateSlug_MaxLength(t *testing.T) {
	long := strings.Repeat("a", 250)
	result := GenerateSlug(long)
	assert.LessOrEqual(t, len(result), 200)
}

func TestGenerateSlug_EmptyTitle(t *testing.T) {
	assert.Equal(t, "", GenerateSlug(""))
}

func TestUniqueSlug_NoCollision(t *testing.T) {
	fn := func(_ context.Context, _, _ string) (bool, error) {
		return false, nil
	}
	slug, err := UniqueSlug(context.Background(), "Test Post", "", fn)
	require.NoError(t, err)
	assert.Equal(t, "test-post", slug)
}

func TestUniqueSlug_WithCollision(t *testing.T) {
	call := 0
	fn := func(_ context.Context, slug, _ string) (bool, error) {
		call++
		if slug == "test-post" {
			return true, nil // first one exists
		}
		return false, nil
	}
	slug, err := UniqueSlug(context.Background(), "Test Post", "", fn)
	require.NoError(t, err)
	assert.Equal(t, "test-post-2", slug)
}

func TestUniqueSlug_EmptyTitle(t *testing.T) {
	fn := func(_ context.Context, _, _ string) (bool, error) {
		return false, nil
	}
	slug, err := UniqueSlug(context.Background(), "", "", fn)
	require.NoError(t, err)
	assert.Equal(t, "untitled", slug)
}

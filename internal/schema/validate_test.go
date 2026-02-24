package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSlug_Valid(t *testing.T) {
	tests := []string{
		"blog",
		"my_site_01",
		"abc",
		"a_b_c_d_e",
		"site123",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 50 chars, max length
	}
	for _, slug := range tests {
		t.Run(slug, func(t *testing.T) {
			assert.True(t, ValidateSlug(slug), "expected valid: %q", slug)
		})
	}
}

func TestValidateSlug_Invalid(t *testing.T) {
	tests := []struct {
		name string
		slug string
	}{
		{"too short (2 chars)", "ab"},
		{"too short (1 char)", "a"},
		{"empty", ""},
		{"uppercase", "Blog"},
		{"hyphen", "my-site"},
		{"space", "my site"},
		{"special chars", "site@123"},
		{"dot", "my.site"},
		{"too long (51 chars)", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, ValidateSlug(tt.slug), "expected invalid: %q", tt.slug)
		})
	}
}

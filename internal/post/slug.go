package post

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var nonAlphanumDash = regexp.MustCompile(`[^a-z0-9-]+`)
var multiDash = regexp.MustCompile(`-{2,}`)

// GenerateSlug creates a URL-friendly slug from an English title.
func GenerateSlug(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = nonAlphanumDash.ReplaceAllString(s, "-")
	s = multiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 200 {
		s = s[:200]
		s = strings.TrimRight(s, "-")
	}
	return s
}

// UniqueSlug generates a slug and resolves collisions by appending -2, -3, etc.
func UniqueSlug(ctx context.Context, title string, excludeID string, existsFn func(ctx context.Context, slug, excludeID string) (bool, error)) (string, error) {
	base := GenerateSlug(title)
	if base == "" {
		base = "untitled"
	}

	slug := base
	for i := 2; i <= 100; i++ {
		exists, err := existsFn(ctx, slug, excludeID)
		if err != nil {
			return "", fmt.Errorf("check slug uniqueness: %w", err)
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
	return "", fmt.Errorf("could not generate unique slug after 100 attempts")
}

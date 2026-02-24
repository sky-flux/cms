package model

import (
	"strings"
	"time"

	"github.com/uptrace/bun"
)

// SetTimestamps sets created_at and updated_at on INSERT, only updated_at on UPDATE.
func SetTimestamps(createdAt *time.Time, updatedAt *time.Time, query bun.Query) {
	now := time.Now()
	switch query.(type) {
	case *bun.InsertQuery:
		*createdAt = now
		*updatedAt = now
	case *bun.UpdateQuery:
		*updatedAt = now
	}
}

// SetUpdatedAt sets updated_at on UPDATE only. Use for models without created_at (e.g. Config).
func SetUpdatedAt(updatedAt *time.Time, query bun.Query) {
	if _, ok := query.(*bun.UpdateQuery); ok {
		*updatedAt = time.Now()
	}
}

// NormalizeEmail lowercases and trims whitespace from an email address.
func NormalizeEmail(email *string) {
	*email = strings.ToLower(strings.TrimSpace(*email))
}

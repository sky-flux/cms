# Bun ORM Model Hooks Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement `BeforeAppendModelHook` on 17 models to manage timestamps, email normalization, and optimistic locking at the application layer, replacing database triggers.

**Architecture:** A shared `hooks.go` file provides `SetTimestamps`, `SetUpdatedAt`, and `NormalizeEmail` helper functions. Each model implements the `bun.BeforeAppendModelHook` interface, calling the shared helpers plus any model-specific logic (email normalization for User/Comment, version increment for Post).

**Tech Stack:** Go 1.25+ / uptrace/bun ORM / testify

**Design doc:** `docs/plans/2026-02-24-bun-hooks-design.md`

---

### Task 1: Create hooks.go with shared helper functions

**Files:**
- Create: `internal/model/hooks.go`
- Test: `internal/model/hooks_test.go`

**Step 1: Write failing tests for SetTimestamps**

```go
// internal/model/hooks_test.go
package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

func TestSetTimestamps_Insert(t *testing.T) {
	var createdAt, updatedAt time.Time
	q := (*bun.InsertQuery)(nil)

	SetTimestamps(&createdAt, &updatedAt, q)

	assert.False(t, createdAt.IsZero(), "createdAt should be set on INSERT")
	assert.False(t, updatedAt.IsZero(), "updatedAt should be set on INSERT")
	assert.Equal(t, createdAt, updatedAt, "both timestamps should match on INSERT")
}

func TestSetTimestamps_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	createdAt := original
	updatedAt := original
	q := (*bun.UpdateQuery)(nil)

	SetTimestamps(&createdAt, &updatedAt, q)

	assert.Equal(t, original, createdAt, "createdAt should NOT change on UPDATE")
	assert.NotEqual(t, original, updatedAt, "updatedAt should change on UPDATE")
}

func TestSetTimestamps_Select(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	createdAt := original
	updatedAt := original
	q := (*bun.SelectQuery)(nil)

	SetTimestamps(&createdAt, &updatedAt, q)

	assert.Equal(t, original, createdAt, "createdAt should NOT change on SELECT")
	assert.Equal(t, original, updatedAt, "updatedAt should NOT change on SELECT")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/model/ -run TestSetTimestamps -v`
Expected: FAIL — `SetTimestamps` undefined

**Step 3: Write failing tests for SetUpdatedAt**

Add to `internal/model/hooks_test.go`:

```go
func TestSetUpdatedAt_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := original
	q := (*bun.UpdateQuery)(nil)

	SetUpdatedAt(&updatedAt, q)

	assert.NotEqual(t, original, updatedAt, "updatedAt should change on UPDATE")
}

func TestSetUpdatedAt_Insert(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := original
	q := (*bun.InsertQuery)(nil)

	SetUpdatedAt(&updatedAt, q)

	assert.Equal(t, original, updatedAt, "updatedAt should NOT change on INSERT for SetUpdatedAt")
}
```

**Step 4: Write failing tests for NormalizeEmail**

Add to `internal/model/hooks_test.go`:

```go
func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  Alice@Example.COM  ", "alice@example.com"},
		{"user@test.com", "user@test.com"},
		{" BOB@MAIL.COM", "bob@mail.com"},
		{"", ""},
	}
	for _, tt := range tests {
		email := tt.input
		NormalizeEmail(&email)
		assert.Equal(t, tt.expected, email)
	}
}
```

**Step 5: Implement hooks.go**

```go
// internal/model/hooks.go
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
```

**Step 6: Run all tests to verify they pass**

Run: `go test ./internal/model/ -run "TestSetTimestamps|TestSetUpdatedAt|TestNormalizeEmail" -v`
Expected: ALL PASS

**Step 7: Commit**

```bash
git add internal/model/hooks.go internal/model/hooks_test.go
git commit -m "feat(model): add shared hook helpers for timestamps and email normalization"
```

---

### Task 2: Add BeforeAppendModel to User model

**Files:**
- Modify: `internal/model/user.go:1-22`
- Test: `internal/model/hooks_test.go` (append)

**Step 1: Write failing test**

Append to `internal/model/hooks_test.go`:

```go
func TestUser_BeforeAppendModel_Insert(t *testing.T) {
	u := &User{Email: "  Alice@Example.COM  "}
	q := (*bun.InsertQuery)(nil)

	err := u.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.False(t, u.CreatedAt.IsZero())
	assert.False(t, u.UpdatedAt.IsZero())
	assert.Equal(t, "alice@example.com", u.Email)
}

func TestUser_BeforeAppendModel_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	u := &User{
		Email:     " BOB@Test.com ",
		CreatedAt: original,
		UpdatedAt: original,
	}
	q := (*bun.UpdateQuery)(nil)

	err := u.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, original, u.CreatedAt, "createdAt should not change on UPDATE")
	assert.NotEqual(t, original, u.UpdatedAt)
	assert.Equal(t, "bob@test.com", u.Email)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/model/ -run TestUser_BeforeAppendModel -v`
Expected: FAIL — `BeforeAppendModel` method not found

**Step 3: Implement hook on User**

Add to end of `internal/model/user.go` (after the struct, before closing):

```go
import "context"
// (add "context" to the existing import block)

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (u *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&u.CreatedAt, &u.UpdatedAt, query)
	NormalizeEmail(&u.Email)
	return nil
}
```

Note: Add `"context"` to the existing import block at the top of the file.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/model/ -run TestUser_BeforeAppendModel -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/model/user.go internal/model/hooks_test.go
git commit -m "feat(model): add BeforeAppendModel hook to User"
```

---

### Task 3: Add BeforeAppendModel to Post and PostTranslation

**Files:**
- Modify: `internal/model/post.go:1-87`
- Test: `internal/model/hooks_test.go` (append)

**Step 1: Write failing tests**

Append to `internal/model/hooks_test.go`:

```go
func TestPost_BeforeAppendModel_Insert(t *testing.T) {
	p := &Post{Version: 1}
	q := (*bun.InsertQuery)(nil)

	err := p.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.False(t, p.CreatedAt.IsZero())
	assert.False(t, p.UpdatedAt.IsZero())
	assert.Equal(t, 1, p.Version, "Version should NOT increment on INSERT")
}

func TestPost_BeforeAppendModel_Update_IncrVersion(t *testing.T) {
	p := &Post{Version: 3}
	q := (*bun.UpdateQuery)(nil)

	err := p.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, 4, p.Version, "Version should increment to 4 on UPDATE")
	assert.False(t, p.UpdatedAt.IsZero())
}

func TestPostTranslation_BeforeAppendModel_Insert(t *testing.T) {
	pt := &PostTranslation{}
	q := (*bun.InsertQuery)(nil)

	err := pt.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.False(t, pt.CreatedAt.IsZero())
	assert.False(t, pt.UpdatedAt.IsZero())
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/model/ -run "TestPost_BeforeAppendModel|TestPostTranslation_BeforeAppendModel" -v`
Expected: FAIL

**Step 3: Implement hooks on Post and PostTranslation**

Add `"context"` to the import block in `internal/model/post.go`, then add after the `PostTagMap` struct:

```go
// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (p *Post) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&p.CreatedAt, &p.UpdatedAt, query)
	if _, ok := query.(*bun.UpdateQuery); ok {
		p.Version++
	}
	return nil
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (pt *PostTranslation) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&pt.CreatedAt, &pt.UpdatedAt, query)
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/model/ -run "TestPost_BeforeAppendModel|TestPostTranslation_BeforeAppendModel" -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/model/post.go internal/model/hooks_test.go
git commit -m "feat(model): add BeforeAppendModel hooks to Post and PostTranslation"
```

---

### Task 4: Add BeforeAppendModel to Comment

**Files:**
- Modify: `internal/model/comment.go:1-29`
- Test: `internal/model/hooks_test.go` (append)

**Step 1: Write failing tests**

Append to `internal/model/hooks_test.go`:

```go
func TestComment_BeforeAppendModel_Insert_Anonymous(t *testing.T) {
	c := &Comment{
		AuthorEmail: " Guest@Example.COM ",
		Content:     "hello",
	}
	q := (*bun.InsertQuery)(nil)

	err := c.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, "guest@example.com", c.AuthorEmail)
	assert.False(t, c.CreatedAt.IsZero())
}

func TestComment_BeforeAppendModel_Insert_Authenticated(t *testing.T) {
	uid := "user-123"
	c := &Comment{
		UserID:      &uid,
		AuthorEmail: " UPPER@test.COM ",
	}
	q := (*bun.InsertQuery)(nil)

	err := c.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, " UPPER@test.COM ", c.AuthorEmail, "should NOT normalize when user_id is set")
}

func TestComment_BeforeAppendModel_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	c := &Comment{CreatedAt: original, UpdatedAt: original}
	q := (*bun.UpdateQuery)(nil)

	err := c.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, original, c.CreatedAt)
	assert.NotEqual(t, original, c.UpdatedAt)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/model/ -run TestComment_BeforeAppendModel -v`
Expected: FAIL

**Step 3: Implement hook on Comment**

Add `"context"` to the import block in `internal/model/comment.go`, then add after the struct:

```go
// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (c *Comment) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&c.CreatedAt, &c.UpdatedAt, query)
	if _, ok := query.(*bun.InsertQuery); ok && c.UserID == nil && c.AuthorEmail != "" {
		NormalizeEmail(&c.AuthorEmail)
	}
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/model/ -run TestComment_BeforeAppendModel -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/model/comment.go internal/model/hooks_test.go
git commit -m "feat(model): add BeforeAppendModel hook to Comment"
```

---

### Task 5: Add BeforeAppendModel to Config and SiteConfig

**Files:**
- Modify: `internal/model/config.go:1-18`
- Modify: `internal/model/site_config.go:1-18`
- Test: `internal/model/hooks_test.go` (append)

**Step 1: Write failing tests**

Append to `internal/model/hooks_test.go`:

```go
func TestConfig_BeforeAppendModel_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	c := &Config{UpdatedAt: original}
	q := (*bun.UpdateQuery)(nil)

	err := c.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.NotEqual(t, original, c.UpdatedAt)
}

func TestConfig_BeforeAppendModel_Insert(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	c := &Config{UpdatedAt: original}
	q := (*bun.InsertQuery)(nil)

	err := c.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, original, c.UpdatedAt, "SetUpdatedAt should not change on INSERT")
}

func TestSiteConfig_BeforeAppendModel_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	sc := &SiteConfig{UpdatedAt: original}
	q := (*bun.UpdateQuery)(nil)

	err := sc.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.NotEqual(t, original, sc.UpdatedAt)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/model/ -run "TestConfig_BeforeAppendModel|TestSiteConfig_BeforeAppendModel" -v`
Expected: FAIL

**Step 3: Implement hooks on Config and SiteConfig**

Add `"context"` to the import block in `internal/model/config.go`:

```go
// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (c *Config) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetUpdatedAt(&c.UpdatedAt, query)
	return nil
}
```

Add `"context"` to the import block in `internal/model/site_config.go`:

```go
// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (sc *SiteConfig) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetUpdatedAt(&sc.UpdatedAt, query)
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/model/ -run "TestConfig_BeforeAppendModel|TestSiteConfig_BeforeAppendModel" -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/model/config.go internal/model/site_config.go internal/model/hooks_test.go
git commit -m "feat(model): add BeforeAppendModel hooks to Config and SiteConfig"
```

---

### Task 6: Add BeforeAppendModel to remaining models (batch)

These 10 models all have identical hook logic (SetTimestamps only, no extra behavior).

**Files:**
- Modify: `internal/model/site.go:1-25`
- Modify: `internal/model/media.go:1-32`
- Modify: `internal/model/category.go:1-25`
- Modify: `internal/model/role.go:1-20`
- Modify: `internal/model/api_endpoint.go:1-28`
- Modify: `internal/model/admin_menu.go:1-31`
- Modify: `internal/model/role_template.go:1-32`
- Modify: `internal/model/user_totp.go:1-20`
- Modify: `internal/model/menu.go:1-42`
- Modify: `internal/model/redirect.go:1-22`
- Test: `internal/model/hooks_test.go` (append)

**Step 1: Write a table-driven test covering all 10 + SiteMenu + SiteMenuItem (12 models)**

Append to `internal/model/hooks_test.go`:

```go
func TestTimestampOnly_BeforeAppendModel(t *testing.T) {
	type hookable interface {
		BeforeAppendModel(ctx context.Context, query bun.Query) error
	}
	type hasTimestamps interface {
		hookable
		GetCreatedAt() time.Time
		GetUpdatedAt() time.Time
	}

	// We test via direct struct access instead of interface to avoid adding getters.
	// Each sub-test verifies INSERT sets both timestamps and UPDATE only sets updated_at.

	t.Run("Site", func(t *testing.T) {
		m := &Site{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
		assert.False(t, m.UpdatedAt.IsZero())

		orig := m.CreatedAt
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.UpdateQuery)(nil)))
		assert.Equal(t, orig, m.CreatedAt)
	})

	t.Run("MediaFile", func(t *testing.T) {
		m := &MediaFile{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("Category", func(t *testing.T) {
		m := &Category{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("Role", func(t *testing.T) {
		m := &Role{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("APIEndpoint", func(t *testing.T) {
		m := &APIEndpoint{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("AdminMenu", func(t *testing.T) {
		m := &AdminMenu{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("RoleTemplate", func(t *testing.T) {
		m := &RoleTemplate{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("UserTOTP", func(t *testing.T) {
		m := &UserTOTP{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("SiteMenu", func(t *testing.T) {
		m := &SiteMenu{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("SiteMenuItem", func(t *testing.T) {
		m := &SiteMenuItem{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("Redirect", func(t *testing.T) {
		m := &Redirect{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/model/ -run TestTimestampOnly_BeforeAppendModel -v`
Expected: FAIL — `BeforeAppendModel` not found on these types

**Step 3: Implement hooks on all 12 models**

For each file, add `"context"` to the import block and append the hook method after the struct.

The hook method is identical for all 12:

```go
// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (x *TYPE) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&x.CreatedAt, &x.UpdatedAt, query)
	return nil
}
```

Replace `TYPE` and receiver name per model:

| File | Receiver | Type |
|------|----------|------|
| `site.go` | `s` | `Site` |
| `media.go` | `mf` | `MediaFile` |
| `category.go` | `c` | `Category` |
| `role.go` | `r` | `Role` |
| `api_endpoint.go` | `ae` | `APIEndpoint` |
| `admin_menu.go` | `m` | `AdminMenu` |
| `role_template.go` | `rt` | `RoleTemplate` |
| `user_totp.go` | `ut` | `UserTOTP` |
| `menu.go` (`SiteMenu`) | `sm` | `SiteMenu` |
| `menu.go` (`SiteMenuItem`) | `mi` | `SiteMenuItem` |
| `redirect.go` | `rd` | `Redirect` |

**Step 4: Run all tests to verify they pass**

Run: `go test ./internal/model/ -run TestTimestampOnly_BeforeAppendModel -v`
Expected: ALL PASS (11 sub-tests)

**Step 5: Commit**

```bash
git add internal/model/site.go internal/model/media.go internal/model/category.go \
       internal/model/role.go internal/model/api_endpoint.go internal/model/admin_menu.go \
       internal/model/role_template.go internal/model/user_totp.go internal/model/menu.go \
       internal/model/redirect.go internal/model/hooks_test.go
git commit -m "feat(model): add BeforeAppendModel hooks to remaining 11 timestamp models"
```

---

### Task 7: Run full test suite and verify compilation

**Step 1: Run full model test suite**

Run: `go test ./internal/model/ -v -count=1`
Expected: ALL PASS

**Step 2: Verify the entire project compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Run go vet**

Run: `go vet ./internal/model/...`
Expected: No issues

**Step 4: Commit (if any fixes were needed)**

Only if fixes were required — otherwise skip.

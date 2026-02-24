# Model Layer Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement all 32 bun ORM model structs in `internal/model/`, converting PostgreSQL ENUMs to SMALLINT.

**Architecture:** Centralized enums in `enums.go`, one model per file (related join tables colocated). Models are pure structs with bun tags — no business logic. DDL in `database.md` updated to reflect ENUM→SMALLINT migration.

**Tech Stack:** Go 1.25+ / uptrace/bun ORM / PostgreSQL 18 / `json.RawMessage` for JSONB

---

### Task 1: Create enum definitions

**Files:**
- Create: `internal/model/enums.go`

**Step 1: Create enums.go with all 5 enum types**

```go
package model

// PostStatus maps to sfc_site_posts.status (SMALLINT)
// DDL: CHECK (status BETWEEN 1 AND 4)
type PostStatus int8

const (
	PostStatusDraft     PostStatus = iota + 1 // 1
	PostStatusScheduled                       // 2
	PostStatusPublished                       // 3
	PostStatusArchived                        // 4
)

// MediaType maps to sfc_site_media_files.media_type (SMALLINT)
// DDL: CHECK (media_type BETWEEN 1 AND 5)
type MediaType int8

const (
	MediaTypeImage    MediaType = iota + 1 // 1
	MediaTypeVideo                         // 2
	MediaTypeAudio                         // 3
	MediaTypeDocument                      // 4
	MediaTypeOther                         // 5
)

// CommentStatus maps to sfc_site_comments.status (SMALLINT)
// DDL: CHECK (status BETWEEN 1 AND 4)
type CommentStatus int8

const (
	CommentStatusPending  CommentStatus = iota + 1 // 1
	CommentStatusApproved                          // 2
	CommentStatusSpam                              // 3
	CommentStatusTrash                             // 4
)

// MenuItemType maps to sfc_site_menu_items.type (SMALLINT)
// DDL: CHECK (type BETWEEN 1 AND 5)
type MenuItemType int8

const (
	MenuItemTypeCustom   MenuItemType = iota + 1 // 1
	MenuItemTypePost                              // 2
	MenuItemTypeCategory                          // 3
	MenuItemTypeTag                               // 4
	MenuItemTypePage                              // 5
)

// LogAction maps to sfc_site_audits.action (SMALLINT)
// DDL: CHECK (action BETWEEN 1 AND 11)
type LogAction int8

const (
	LogActionCreate         LogAction = iota + 1 // 1
	LogActionUpdate                              // 2
	LogActionDelete                              // 3
	LogActionRestore                             // 4
	LogActionLogin                               // 5
	LogActionLogout                              // 6
	LogActionPublish                             // 7
	LogActionUnpublish                           // 8
	LogActionArchive                             // 9
	LogActionPasswordChange                      // 10
	LogActionSettingsChange                      // 11
)
```

**Step 2: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS (no output)

**Step 3: Commit**

```bash
git add internal/model/enums.go
git commit -m "feat(model): add enum type definitions for SMALLINT mapping"
```

---

### Task 2: Fix existing Config model

**Files:**
- Modify: `internal/model/config.go`

**Step 1: Update Config struct**

Replace entire file content with:

```go
package model

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type Config struct {
	bun.BaseModel `bun:"table:sfc_configs,alias:sc"`

	Key         string          `bun:"key,pk" json:"key"`
	Value       json.RawMessage `bun:"value,type:jsonb" json:"value"`
	Description string          `bun:"description" json:"description,omitempty"`
	UpdatedBy   *string         `bun:"updated_by,type:uuid" json:"updated_by,omitempty"`
	UpdatedAt   time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
```

Changes from original:
- `Value`: `string` → `json.RawMessage` (per JSONB decision)
- `UpdatedBy`: `string` → `*string` (nullable FK per DDL)
- Added `Description string` field (missing per DDL)
- Removed `nullzero` from UpdatedBy (use pointer nil instead)

**Step 2: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/model/config.go
git commit -m "fix(model): add Description field and fix Value/UpdatedBy types in Config"
```

---

### Task 3: Create public schema auth models

**Files:**
- Create: `internal/model/refresh_token.go`
- Create: `internal/model/user_totp.go`
- Create: `internal/model/password_reset_token.go`

**Step 1: Create refresh_token.go**

```go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type RefreshToken struct {
	bun.BaseModel `bun:"table:sfc_refresh_tokens,alias:rt"`

	ID        string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	UserID    string    `bun:"user_id,notnull,type:uuid" json:"user_id"`
	TokenHash string    `bun:"token_hash,notnull,unique" json:"-"`
	ExpiresAt time.Time `bun:"expires_at,notnull" json:"expires_at"`
	Revoked   bool      `bun:"revoked,notnull,default:false" json:"revoked"`
	IPAddress string    `bun:"ip_address,type:inet" json:"ip_address,omitempty"`
	UserAgent string    `bun:"user_agent" json:"user_agent,omitempty"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}
```

**Step 2: Create user_totp.go**

```go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type UserTOTP struct {
	bun.BaseModel `bun:"table:sfc_user_totp,alias:totp"`

	ID              string     `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	UserID          string     `bun:"user_id,notnull,unique,type:uuid" json:"user_id"`
	SecretEncrypted string     `bun:"secret_encrypted,notnull" json:"-"`
	BackupCodesHash []string   `bun:"backup_codes_hash,type:text[],array" json:"-"`
	IsEnabled       bool       `bun:"is_enabled,notnull,default:false" json:"is_enabled"`
	VerifiedAt      *time.Time `bun:"verified_at" json:"verified_at,omitempty"`
	CreatedAt       time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
```

**Step 3: Create password_reset_token.go**

```go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type PasswordResetToken struct {
	bun.BaseModel `bun:"table:sfc_password_reset_tokens,alias:prt"`

	ID        string     `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	UserID    string     `bun:"user_id,notnull,type:uuid" json:"user_id"`
	TokenHash string     `bun:"token_hash,notnull,unique" json:"-"`
	ExpiresAt time.Time  `bun:"expires_at,notnull" json:"expires_at"`
	UsedAt    *time.Time `bun:"used_at" json:"used_at,omitempty"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}
```

**Step 4: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 5: Commit**

```bash
git add internal/model/refresh_token.go internal/model/user_totp.go internal/model/password_reset_token.go
git commit -m "feat(model): add RefreshToken, UserTOTP, PasswordResetToken models"
```

---

### Task 4: Implement Post and related models

**Files:**
- Modify: `internal/model/post.go` (currently empty, only `package model`)

**Step 1: Write Post and related structs**

Replace entire `post.go` with:

```go
package model

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

// Post maps to sfc_site_posts in site_{slug} schema.
type Post struct {
	bun.BaseModel `bun:"table:sfc_site_posts,alias:p"`

	ID             string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	AuthorID       string          `bun:"author_id,notnull,type:uuid" json:"author_id"`
	CoverImageID   *string         `bun:"cover_image_id,type:uuid" json:"cover_image_id,omitempty"`
	PostType       string          `bun:"post_type,notnull,default:'article'" json:"post_type"`
	Status         PostStatus      `bun:"status,notnull,type:smallint,default:1" json:"status"`
	Title          string          `bun:"title,notnull" json:"title"`
	Slug           string          `bun:"slug,notnull" json:"slug"`
	Excerpt        string          `bun:"excerpt" json:"excerpt,omitempty"`
	Content        string          `bun:"content" json:"content,omitempty"`
	ContentJSON    json.RawMessage `bun:"content_json,type:jsonb" json:"content_json,omitempty"`
	MetaTitle      string          `bun:"meta_title" json:"meta_title,omitempty"`
	MetaDesc       string          `bun:"meta_description" json:"meta_description,omitempty"`
	OGImageURL     string          `bun:"og_image_url" json:"og_image_url,omitempty"`
	ExtraFields    json.RawMessage `bun:"extra_fields,type:jsonb,default:'{}'" json:"extra_fields,omitempty"`
	ViewCount      int64           `bun:"view_count,notnull,default:0" json:"view_count"`
	Version        int             `bun:"version,notnull,default:1" json:"version"`
	PublishedAt    *time.Time      `bun:"published_at" json:"published_at,omitempty"`
	ScheduledAt    *time.Time      `bun:"scheduled_at" json:"scheduled_at,omitempty"`
	CreatedAt      time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt      *time.Time      `bun:"deleted_at,soft_delete,nullzero" json:"-"`

	Author *User `bun:"rel:belongs-to,join:author_id=id" json:"author,omitempty"`
}

// PostTranslation maps to sfc_site_post_translations.
type PostTranslation struct {
	bun.BaseModel `bun:"table:sfc_site_post_translations,alias:pt"`

	ID          string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	PostID      string          `bun:"post_id,notnull,type:uuid" json:"post_id"`
	Locale      string          `bun:"locale,notnull" json:"locale"`
	Title       string          `bun:"title" json:"title,omitempty"`
	Excerpt     string          `bun:"excerpt" json:"excerpt,omitempty"`
	Content     string          `bun:"content" json:"content,omitempty"`
	ContentJSON json.RawMessage `bun:"content_json,type:jsonb" json:"content_json,omitempty"`
	MetaTitle   string          `bun:"meta_title" json:"meta_title,omitempty"`
	MetaDesc    string          `bun:"meta_description" json:"meta_description,omitempty"`
	OGImageURL  string          `bun:"og_image_url" json:"og_image_url,omitempty"`
	CreatedAt   time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// PostRevision maps to sfc_site_post_revisions.
type PostRevision struct {
	bun.BaseModel `bun:"table:sfc_site_post_revisions,alias:pr"`

	ID          string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	PostID      string          `bun:"post_id,notnull,type:uuid" json:"post_id"`
	EditorID    string          `bun:"editor_id,notnull,type:uuid" json:"editor_id"`
	Version     int             `bun:"version,notnull" json:"version"`
	Title       string          `bun:"title" json:"title,omitempty"`
	Content     string          `bun:"content" json:"content,omitempty"`
	ContentJSON json.RawMessage `bun:"content_json,type:jsonb" json:"content_json,omitempty"`
	DiffSummary string          `bun:"diff_summary" json:"diff_summary,omitempty"`
	CreatedAt   time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

// PostCategoryMap maps to sfc_site_post_category_map (many-to-many).
type PostCategoryMap struct {
	bun.BaseModel `bun:"table:sfc_site_post_category_map"`

	PostID     string `bun:"post_id,pk,type:uuid" json:"post_id"`
	CategoryID string `bun:"category_id,pk,type:uuid" json:"category_id"`
	IsPrimary  bool   `bun:"is_primary,notnull,default:false" json:"is_primary"`
}

// PostTagMap maps to sfc_site_post_tag_map (many-to-many).
type PostTagMap struct {
	bun.BaseModel `bun:"table:sfc_site_post_tag_map"`

	PostID string `bun:"post_id,pk,type:uuid" json:"post_id"`
	TagID  string `bun:"tag_id,pk,type:uuid" json:"tag_id"`
}
```

**Step 2: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/model/post.go
git commit -m "feat(model): implement Post, PostTranslation, PostRevision, PostCategoryMap, PostTagMap"
```

---

### Task 5: Implement Category and Tag models

**Files:**
- Modify: `internal/model/category.go` (currently empty)
- Modify: `internal/model/tag.go` (currently empty)

**Step 1: Write category.go**

```go
package model

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type Category struct {
	bun.BaseModel `bun:"table:sfc_site_categories,alias:c"`

	ID          string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	ParentID    *string         `bun:"parent_id,type:uuid" json:"parent_id,omitempty"`
	Name        string          `bun:"name,notnull" json:"name"`
	Slug        string          `bun:"slug,notnull" json:"slug"`
	Path        string          `bun:"path,notnull,default:'/'" json:"path"`
	Description string          `bun:"description" json:"description,omitempty"`
	SortOrder   int             `bun:"sort_order,notnull,default:0" json:"sort_order"`
	Meta        json.RawMessage `bun:"meta,type:jsonb,default:'{}'" json:"meta,omitempty"`
	CreatedAt   time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	Children []*Category `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
}
```

**Step 2: Write tag.go**

```go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Tag struct {
	bun.BaseModel `bun:"table:sfc_site_tags,alias:t"`

	ID        string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Name      string    `bun:"name,notnull,unique" json:"name"`
	Slug      string    `bun:"slug,notnull,unique" json:"slug"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}
```

**Step 3: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add internal/model/category.go internal/model/tag.go
git commit -m "feat(model): implement Category and Tag models"
```

---

### Task 6: Implement MediaFile model

**Files:**
- Modify: `internal/model/media.go` (currently empty)

**Step 1: Write media.go**

```go
package model

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type MediaFile struct {
	bun.BaseModel `bun:"table:sfc_site_media_files,alias:mf"`

	ID             string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	UploaderID     string          `bun:"uploader_id,notnull,type:uuid" json:"uploader_id"`
	FileName       string          `bun:"file_name,notnull" json:"file_name"`
	OriginalName   string          `bun:"original_name,notnull" json:"original_name"`
	MimeType       string          `bun:"mime_type,notnull" json:"mime_type"`
	MediaType      MediaType       `bun:"media_type,notnull,type:smallint,default:5" json:"media_type"`
	FileSize       int64           `bun:"file_size,notnull" json:"file_size"`
	Width          *int            `bun:"width" json:"width,omitempty"`
	Height         *int            `bun:"height" json:"height,omitempty"`
	StoragePath    string          `bun:"storage_path,notnull" json:"storage_path"`
	PublicURL      string          `bun:"public_url,notnull" json:"public_url"`
	WebpURL        string          `bun:"webp_url" json:"webp_url,omitempty"`
	ThumbnailURLs  json.RawMessage `bun:"thumbnail_urls,type:jsonb,default:'{}'" json:"thumbnail_urls,omitempty"`
	ReferenceCount int             `bun:"reference_count,notnull,default:0" json:"reference_count"`
	AltText        string          `bun:"alt_text" json:"alt_text,omitempty"`
	Metadata       json.RawMessage `bun:"metadata,type:jsonb,default:'{}'" json:"metadata,omitempty"`
	CreatedAt      time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt      *time.Time      `bun:"deleted_at,soft_delete,nullzero" json:"-"`
}
```

**Step 2: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/model/media.go
git commit -m "feat(model): implement MediaFile model"
```

---

### Task 7: Implement Comment model

**Files:**
- Modify: `internal/model/comment.go` (currently empty)

**Step 1: Write comment.go**

```go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Comment struct {
	bun.BaseModel `bun:"table:sfc_site_comments,alias:cm"`

	ID          string        `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	PostID      string        `bun:"post_id,notnull,type:uuid" json:"post_id"`
	ParentID    *string       `bun:"parent_id,type:uuid" json:"parent_id,omitempty"`
	UserID      *string       `bun:"user_id,type:uuid" json:"user_id,omitempty"`
	AuthorName  string        `bun:"author_name" json:"author_name,omitempty"`
	AuthorEmail string        `bun:"author_email" json:"author_email,omitempty"`
	AuthorURL   string        `bun:"author_url" json:"author_url,omitempty"`
	AuthorIP    string        `bun:"author_ip,type:inet" json:"-"`
	UserAgent   string        `bun:"user_agent" json:"-"`
	Content     string        `bun:"content,notnull" json:"content"`
	Status      CommentStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
	IsPinned    bool          `bun:"is_pinned,notnull,default:false" json:"is_pinned"`
	CreatedAt   time.Time     `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time     `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt   *time.Time    `bun:"deleted_at,soft_delete,nullzero" json:"-"`

	Children []*Comment `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
}
```

**Step 2: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/model/comment.go
git commit -m "feat(model): implement Comment model"
```

---

### Task 8: Implement SiteMenu and SiteMenuItem models

**Files:**
- Modify: `internal/model/menu.go` (currently empty)

**Step 1: Write menu.go**

```go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

// SiteMenu maps to sfc_site_menus (frontend navigation menu).
// Distinct from AdminMenu (public.sfc_menus) which is the backend sidebar menu.
type SiteMenu struct {
	bun.BaseModel `bun:"table:sfc_site_menus,alias:sm"`

	ID        string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Name      string    `bun:"name,notnull" json:"name"`
	Slug      string    `bun:"slug,notnull,unique" json:"slug"`
	Location  string    `bun:"location" json:"location,omitempty"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	Items []*SiteMenuItem `bun:"rel:has-many,join:id=menu_id" json:"items,omitempty"`
}

// SiteMenuItem maps to sfc_site_menu_items.
type SiteMenuItem struct {
	bun.BaseModel `bun:"table:sfc_site_menu_items,alias:mi"`

	ID          string       `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	MenuID      string       `bun:"menu_id,notnull,type:uuid" json:"menu_id"`
	ParentID    *string      `bun:"parent_id,type:uuid" json:"parent_id,omitempty"`
	Label       string       `bun:"label,notnull" json:"label"`
	URL         string       `bun:"url" json:"url,omitempty"`
	Target      string       `bun:"target,notnull,default:'_self'" json:"target"`
	Type        MenuItemType `bun:"type,notnull,type:smallint,default:1" json:"type"`
	ReferenceID *string      `bun:"reference_id,type:uuid" json:"reference_id,omitempty"`
	SortOrder   int          `bun:"sort_order,notnull,default:0" json:"sort_order"`
	IsActive    bool         `bun:"is_active,notnull,default:true" json:"is_active"`
	CreatedAt   time.Time    `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time    `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	Children []*SiteMenuItem `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
}
```

**Step 2: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/model/menu.go
git commit -m "feat(model): implement SiteMenu and SiteMenuItem models"
```

---

### Task 9: Implement Redirect model

**Files:**
- Modify: `internal/model/redirect.go` (currently empty)

**Step 1: Write redirect.go**

```go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Redirect struct {
	bun.BaseModel `bun:"table:sfc_site_redirects,alias:rd"`

	ID         string     `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	SourcePath string     `bun:"source_path,notnull,unique" json:"source_path"`
	TargetURL  string     `bun:"target_url,notnull" json:"target_url"`
	StatusCode int        `bun:"status_code,notnull,default:301" json:"status_code"`
	IsActive   bool       `bun:"is_active,notnull,default:true" json:"is_active"`
	HitCount   int64      `bun:"hit_count,notnull,default:0" json:"hit_count"`
	LastHitAt  *time.Time `bun:"last_hit_at" json:"last_hit_at,omitempty"`
	CreatedBy  *string    `bun:"created_by,type:uuid" json:"created_by,omitempty"`
	CreatedAt  time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
```

**Step 2: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/model/redirect.go
git commit -m "feat(model): implement Redirect model"
```

---

### Task 10: Implement PreviewToken model

**Files:**
- Modify: `internal/model/preview_token.go` (currently empty)

**Step 1: Write preview_token.go**

```go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type PreviewToken struct {
	bun.BaseModel `bun:"table:sfc_site_preview_tokens,alias:pvt"`

	ID        string    `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	PostID    string    `bun:"post_id,notnull,type:uuid" json:"post_id"`
	TokenHash string    `bun:"token_hash,notnull,unique" json:"-"`
	ExpiresAt time.Time `bun:"expires_at,notnull" json:"expires_at"`
	CreatedBy *string   `bun:"created_by,type:uuid" json:"created_by,omitempty"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}
```

**Step 2: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/model/preview_token.go
git commit -m "feat(model): implement PreviewToken model"
```

---

### Task 11: Implement APIKey model

**Files:**
- Modify: `internal/model/api_key.go` (currently empty)

**Step 1: Write api_key.go**

```go
package model

import (
	"time"

	"github.com/uptrace/bun"
)

type APIKey struct {
	bun.BaseModel `bun:"table:sfc_site_api_keys,alias:ak"`

	ID         string     `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	OwnerID    string     `bun:"owner_id,notnull,type:uuid" json:"owner_id"`
	Name       string     `bun:"name,notnull" json:"name"`
	KeyHash    string     `bun:"key_hash,notnull,unique" json:"-"`
	KeyPrefix  string     `bun:"key_prefix,notnull" json:"key_prefix"`
	IsActive   bool       `bun:"is_active,notnull,default:true" json:"is_active"`
	LastUsedAt *time.Time `bun:"last_used_at" json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `bun:"expires_at" json:"expires_at,omitempty"`
	RateLimit  int        `bun:"rate_limit,notnull,default:100" json:"rate_limit"`
	CreatedAt  time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	RevokedAt  *time.Time `bun:"revoked_at" json:"revoked_at,omitempty"`
}
```

**Step 2: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/model/api_key.go
git commit -m "feat(model): implement APIKey model"
```

---

### Task 12: Implement Audit model (rename audit_log.go → audit.go)

**Files:**
- Delete: `internal/model/audit_log.go`
- Create: `internal/model/audit.go`

**Step 1: Delete old file and create audit.go**

```bash
rm internal/model/audit_log.go
```

Then create `internal/model/audit.go`:

```go
package model

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type Audit struct {
	bun.BaseModel `bun:"table:sfc_site_audits,alias:a"`

	ID               string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	ActorID          *string         `bun:"actor_id,type:uuid" json:"actor_id,omitempty"`
	ActorEmail       string          `bun:"actor_email" json:"actor_email,omitempty"`
	Action           LogAction       `bun:"action,notnull,type:smallint" json:"action"`
	ResourceType     string          `bun:"resource_type,notnull" json:"resource_type"`
	ResourceID       string          `bun:"resource_id" json:"resource_id,omitempty"`
	ResourceSnapshot json.RawMessage `bun:"resource_snapshot,type:jsonb" json:"resource_snapshot,omitempty"`
	IPAddress        string          `bun:"ip_address,type:inet" json:"ip_address,omitempty"`
	UserAgent        string          `bun:"user_agent" json:"user_agent,omitempty"`
	CreatedAt        time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}
```

**Step 2: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/model/audit.go
git rm internal/model/audit_log.go
git commit -m "feat(model): implement Audit model (rename audit_log.go → audit.go)"
```

---

### Task 13: Create PostType and SiteConfig models

**Files:**
- Create: `internal/model/post_type.go`
- Create: `internal/model/site_config.go`

**Step 1: Create post_type.go**

```go
package model

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type PostType struct {
	bun.BaseModel `bun:"table:sfc_site_post_types,alias:pty"`

	ID          string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	Name        string          `bun:"name,notnull,unique" json:"name"`
	Slug        string          `bun:"slug,notnull,unique" json:"slug"`
	Description string          `bun:"description" json:"description,omitempty"`
	Fields      json.RawMessage `bun:"fields,type:jsonb,default:'[]'" json:"fields"`
	BuiltIn     bool            `bun:"built_in,notnull,default:false" json:"built_in"`
	CreatedAt   time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
```

**Step 2: Create site_config.go**

```go
package model

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type SiteConfig struct {
	bun.BaseModel `bun:"table:sfc_site_configs,alias:scfg"`

	Key         string          `bun:"key,pk" json:"key"`
	Value       json.RawMessage `bun:"value,type:jsonb" json:"value"`
	Description string          `bun:"description" json:"description,omitempty"`
	UpdatedBy   *string         `bun:"updated_by,type:uuid" json:"updated_by,omitempty"`
	UpdatedAt   time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
```

**Step 3: Verify build**

Run: `go build ./internal/model/...`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add internal/model/post_type.go internal/model/site_config.go
git commit -m "feat(model): implement PostType and SiteConfig models"
```

---

### Task 14: Update database.md — Remove ENUM types, add SMALLINT + CHECK

**Files:**
- Modify: `docs/database.md`

**Step 1: Remove ENUM type definitions**

In `docs/database.md`, find the enum creation block (Section 2A, around lines 420-429) and replace:

```sql
-- Old:
CREATE TYPE post_status    AS ENUM ('draft', 'scheduled', 'published', 'archived');
CREATE TYPE media_type     AS ENUM ('image', 'video', 'audio', 'document', 'other');
CREATE TYPE comment_status AS ENUM ('pending', 'approved', 'spam', 'trash');
CREATE TYPE menu_item_type AS ENUM ('custom', 'post', 'category', 'tag', 'page');
CREATE TYPE log_action     AS ENUM (
    'create', 'update', 'delete', 'restore',
    'login', 'logout', 'publish', 'unpublish',
    'archive', 'password_change', 'settings_change'
);
```

With a comment block:

```sql
-- ============================================
-- 枚举值映射（SMALLINT 常量定义，参见 internal/model/enums.go）
-- ============================================
-- post_status:    1=draft, 2=scheduled, 3=published, 4=archived
-- media_type:     1=image, 2=video, 3=audio, 4=document, 5=other
-- comment_status: 1=pending, 2=approved, 3=spam, 4=trash
-- menu_item_type: 1=custom, 2=post, 3=category, 4=tag, 5=page
-- log_action:     1=create, 2=update, 3=delete, 4=restore,
--                 5=login, 6=logout, 7=publish, 8=unpublish,
--                 9=archive, 10=password_change, 11=settings_change
```

**Step 2: Update column types in DDL**

Replace these column definitions throughout the site schema DDL:

| Location | Old | New |
|----------|-----|-----|
| sfc_site_posts.status | `status post_status NOT NULL DEFAULT 'draft'` | `status SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4)` |
| sfc_site_media_files.media_type | `media_type media_type NOT NULL DEFAULT 'other'` | `media_type SMALLINT NOT NULL DEFAULT 5 CHECK (media_type BETWEEN 1 AND 5)` |
| sfc_site_comments.status | `status comment_status NOT NULL DEFAULT 'pending'` | `status SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4)` |
| sfc_site_menu_items.type | `type menu_item_type NOT NULL DEFAULT 'custom'` | `type SMALLINT NOT NULL DEFAULT 1 CHECK (type BETWEEN 1 AND 5)` |
| sfc_site_audits.action | `action log_action NOT NULL` | `action SMALLINT NOT NULL CHECK (action BETWEEN 1 AND 11)` |

**Step 3: Update ER diagram column types**

In the mermaid ER diagram (Section 1), change `enum status` → `smallint status` for all affected columns.

**Step 4: Verify no markdown rendering issues**

Open `docs/database.md` and visually check the mermaid diagram renders correctly (no broken syntax).

**Step 5: Commit**

```bash
git add docs/database.md
git commit -m "refactor(db): replace PostgreSQL ENUMs with SMALLINT + CHECK constraints"
```

---

### Task 15: Final verification

**Step 1: Full build check**

Run: `go build ./...`
Expected: SUCCESS (no errors)

**Step 2: Verify all model files exist**

Run: `ls -la internal/model/*.go | wc -l`
Expected: 23 files (18 existing + 6 new - 1 deleted + renamed)

Complete file list:
```
internal/model/
├── admin_menu.go          (existing - AdminMenu, RoleMenu)
├── api_endpoint.go        (existing - APIEndpoint, RoleAPI)
├── api_key.go             (completed - APIKey)
├── audit.go               (new, renamed from audit_log.go - Audit)
├── category.go            (completed - Category)
├── comment.go             (completed - Comment)
├── config.go              (fixed - Config)
├── enums.go               (new - all enum types)
├── media.go               (completed - MediaFile)
├── menu.go                (completed - SiteMenu, SiteMenuItem)
├── password_reset_token.go (new - PasswordResetToken)
├── post.go                (completed - Post, PostTranslation, PostRevision, PostCategoryMap, PostTagMap)
├── post_type.go           (new - PostType)
├── preview_token.go       (completed - PreviewToken)
├── redirect.go            (completed - Redirect)
├── refresh_token.go       (new - RefreshToken)
├── role.go                (existing - Role)
├── role_template.go       (existing - RoleTemplate, RoleTemplateAPI, RoleTemplateMenu)
├── site.go                (existing - Site)
├── site_config.go         (new - SiteConfig)
├── tag.go                 (completed - Tag)
├── user.go                (existing - User)
└── user_role.go           (existing - UserRole)
    user_totp.go           (new - UserTOTP)
```

**Step 3: Commit design doc update**

```bash
git add docs/plans/2026-02-24-model-layer-design.md
git commit -m "docs: update model layer design with Audit naming fix"
```

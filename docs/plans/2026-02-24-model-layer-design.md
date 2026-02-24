# Model Layer Implementation Design

## Goal

Implement all bun ORM model structs in `internal/model/` to map the 32 database tables defined in `docs/database.md`, converting PostgreSQL ENUMs to SMALLINT with Go iota constants.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| PostgreSQL ENUM | **SMALLINT + CHECK** | Avoids `ALTER TYPE ADD VALUE` transaction limitation; Go iota mapping clean |
| JSONB fields | **json.RawMessage** | Deferred parsing; preserves structure; parse in Service layer when needed |
| Relation annotations | **Minimal** | Only self-referential hierarchies + core belongs-to; rest via Service layer JOIN |
| Enum location | **Centralized enums.go** | Avoids cross-file dependency; single source of truth |
| iota start | **iota + 1** | Zero value = unset/invalid, explicit detection of missing data |

## Scope

### Enum Migration (DDL Change)

Remove 5 `CREATE TYPE ... AS ENUM` and replace column types:

| PostgreSQL ENUM | Go Type | SMALLINT Range | CHECK |
|-----------------|---------|----------------|-------|
| `post_status` | `PostStatus int8` | 1-4 | `status BETWEEN 1 AND 4` |
| `media_type` | `MediaType int8` | 1-5 | `media_type BETWEEN 1 AND 5` |
| `comment_status` | `CommentStatus int8` | 1-4 | `status BETWEEN 1 AND 4` |
| `menu_item_type` | `MenuItemType int8` | 1-5 | `type BETWEEN 1 AND 5` |
| `log_action` | `LogAction int8` | 1-11 | `action BETWEEN 1 AND 11` |

### Files to Create (6 new files)

| File | Models |
|------|--------|
| `enums.go` | PostStatus, MediaType, CommentStatus, MenuItemType, LogAction |
| `refresh_token.go` | RefreshToken |
| `user_totp.go` | UserTOTP |
| `password_reset_token.go` | PasswordResetToken |
| `post_type.go` | PostType |
| `site_config.go` | SiteConfig |

### Files to Complete (10 empty files)

| File | Models |
|------|--------|
| `post.go` | Post, PostTranslation, PostRevision, PostCategoryMap, PostTagMap |
| `category.go` | Category |
| `tag.go` | Tag |
| `media.go` | MediaFile |
| `comment.go` | Comment |
| `menu.go` | SiteMenu, SiteMenuItem |
| `redirect.go` | Redirect |
| `preview_token.go` | PreviewToken |
| `api_key.go` | APIKey |
| `audit.go` | Audit (rename from audit_log.go) |

### Files to Fix (1 existing model)

| File | Fix |
|------|-----|
| `config.go` | Add missing `Description` field per DDL |

### DDL to Update

| File | Changes |
|------|---------|
| `docs/database.md` | Remove 5 ENUM types; change column types to SMALLINT; add CHECK constraints |

## Model Patterns

All models follow these conventions (derived from existing User/Site/Role models):

```go
// Standard field patterns:
ID        string     `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
UpdatedAt time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
DeletedAt *time.Time `bun:"deleted_at,soft_delete,nullzero" json:"-"`

// Nullable FK:
ParentID *string `bun:"parent_id,type:uuid" json:"parent_id,omitempty"`

// Enum field:
Status PostStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`

// JSONB field:
Settings json.RawMessage `bun:"settings,type:jsonb,default:'{}'" json:"settings,omitempty"`

// PostgreSQL array:
BackupCodesHash []string `bun:"backup_codes_hash,type:text[],array" json:"-"`

// Sensitive field (hidden from JSON):
TokenHash string `bun:"token_hash,notnull,unique" json:"-"`

// Self-referential hierarchy:
Children []*Category `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
```

### Relation Strategy

Only add `bun:"rel:..."` tags for:
1. **Self-referential trees**: Category.Children, Comment.Children, AdminMenu.Children, SiteMenuItem.Children
2. **Parent-to-items**: SiteMenu.Items (menu always loaded with items)
3. **Core belongs-to**: Post.Author, UserRole.Role, UserRole.User

All other joins handled at Service/Repository layer.

## Enum Definitions

```go
// PostStatus: draft=1, scheduled=2, published=3, archived=4
type PostStatus int8
const (
    PostStatusDraft     PostStatus = iota + 1
    PostStatusScheduled
    PostStatusPublished
    PostStatusArchived
)

// MediaType: image=1, video=2, audio=3, document=4, other=5
type MediaType int8
const (
    MediaTypeImage    MediaType = iota + 1
    MediaTypeVideo
    MediaTypeAudio
    MediaTypeDocument
    MediaTypeOther
)

// CommentStatus: pending=1, approved=2, spam=3, trash=4
type CommentStatus int8
const (
    CommentStatusPending  CommentStatus = iota + 1
    CommentStatusApproved
    CommentStatusSpam
    CommentStatusTrash
)

// MenuItemType: custom=1, post=2, category=3, tag=4, page=5
type MenuItemType int8
const (
    MenuItemTypeCustom   MenuItemType = iota + 1
    MenuItemTypePost
    MenuItemTypeCategory
    MenuItemTypeTag
    MenuItemTypePage
)

// LogAction: create=1 .. settings_change=11
type LogAction int8
const (
    LogActionCreate         LogAction = iota + 1
    LogActionUpdate
    LogActionDelete
    LogActionRestore
    LogActionLogin
    LogActionLogout
    LogActionPublish
    LogActionUnpublish
    LogActionArchive
    LogActionPasswordChange
    LogActionSettingsChange
)
```

## DDL Change Summary

### Remove (from Section 2A header)

```sql
-- DELETE these 5 lines:
CREATE TYPE post_status    AS ENUM ('draft', 'scheduled', 'published', 'archived');
CREATE TYPE media_type     AS ENUM ('image', 'video', 'audio', 'document', 'other');
CREATE TYPE comment_status AS ENUM ('pending', 'approved', 'spam', 'trash');
CREATE TYPE menu_item_type AS ENUM ('custom', 'post', 'category', 'tag', 'page');
CREATE TYPE log_action     AS ENUM (...);
```

### Replace Column Types

| Table | Column | Old | New |
|-------|--------|-----|-----|
| sfc_site_posts | status | `post_status NOT NULL DEFAULT 'draft'` | `SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4)` |
| sfc_site_media_files | media_type | `media_type NOT NULL DEFAULT 'other'` | `SMALLINT NOT NULL DEFAULT 5 CHECK (media_type BETWEEN 1 AND 5)` |
| sfc_site_comments | status | `comment_status NOT NULL DEFAULT 'pending'` | `SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4)` |
| sfc_site_menu_items | type | `menu_item_type NOT NULL DEFAULT 'custom'` | `SMALLINT NOT NULL DEFAULT 1 CHECK (type BETWEEN 1 AND 5)` |
| sfc_site_audits | action | `log_action NOT NULL` | `SMALLINT NOT NULL CHECK (action BETWEEN 1 AND 11)` |

### ER Diagram Updates

Change enum column types in mermaid ER diagram from `enum` to `smallint`.

## Not In Scope

- Migration files (separate task)
- Repository layer
- Service layer
- Validation logic (belongs in DTO/handler layer)

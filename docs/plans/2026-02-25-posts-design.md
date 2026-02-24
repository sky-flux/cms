# Batch 6: Posts Module Design

**Date**: 2026-02-25
**Scope**: 19 site-scoped endpoints (CRUD + status transitions + revisions + translations + preview tokens)
**Prerequisites**: Batch 4 (infrastructure) + Batch 5 (categories, tags, media) completed

---

## 1. Package Structure

Single `internal/post/` package with file-level separation by sub-feature:

```
internal/post/
├── handler.go                  # Handler struct + NewHandler
├── handler_crud.go             # ListPosts, CreatePost, GetPost, UpdatePost, DeletePost (5)
├── handler_status.go           # Publish, Unpublish, RevertToDraft, Restore (4)
├── handler_revision.go         # ListRevisions, Rollback (2)
├── handler_translation.go      # ListTranslations, GetTranslation, UpsertTranslation, DeleteTranslation (4)
├── handler_preview.go          # CreatePreviewToken, ListPreviewTokens, RevokeAll, RevokeOne (4)
├── service.go                  # Service struct + NewService + core CRUD logic
├── service_status.go           # Status transition logic + validateTransition()
├── service_revision.go         # Revision management logic
├── service_translation.go      # Translation CRUD logic
├── service_preview.go          # Preview token generation/revocation logic
├── service_test.go             # CRUD + status transition tests
├── service_revision_test.go    # Revision tests
├── service_translation_test.go # Translation tests
├── service_preview_test.go     # Preview token tests
├── handler_test.go             # Handler layer tests (all endpoints)
├── repository.go               # PostRepository interface + implementation
├── repository_revision.go      # RevisionRepository interface + implementation
├── repository_translation.go   # TranslationRepository interface + implementation
├── repository_preview.go       # PreviewTokenRepository interface + implementation
├── interfaces.go               # All interface definitions (incl. external deps)
├── dto.go                      # Request/response DTOs + conversion functions
└── slug.go                     # Slug generation + collision handling
```

**Rationale**: All 5 sub-features (CRUD, status, revisions, translations, preview) operate on the Post entity. Splitting into sub-packages would create circular dependencies (preview needs post service, translation needs post service). Consistent with existing patterns (`auth/` 14 methods, `rbac/` 23 endpoints — both single packages).

## 2. Endpoints (19 total)

### CRUD (5)
| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | /api/v1/posts | Viewer+ | List posts (all statuses) |
| POST | /api/v1/posts | Editor+ | Create post |
| GET | /api/v1/posts/:id | Viewer+ | Get post detail |
| PUT | /api/v1/posts/:id | Editor+ | Update post (optimistic lock) |
| DELETE | /api/v1/posts/:id | Editor+ | Soft delete post |

### Status Transitions (4)
| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | /api/v1/posts/:id/publish | Editor+ | Publish (draft/scheduled/archived → published) |
| POST | /api/v1/posts/:id/unpublish | Editor+ | Archive (published → archived) |
| POST | /api/v1/posts/:id/revert-to-draft | Editor+ | Revert to draft |
| POST | /api/v1/posts/:id/restore | Editor+ | Restore from trash |

### Revisions (2)
| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | /api/v1/posts/:id/revisions | Viewer+ | List revision history |
| POST | /api/v1/posts/:id/revisions/:rev_id/rollback | Editor+ | Rollback to specific version |

### Translations (4)
| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | /api/v1/posts/:id/translations | Viewer+ | List all translations |
| GET | /api/v1/posts/:id/translations/:locale | Viewer+ | Get translation by locale |
| PUT | /api/v1/posts/:id/translations/:locale | Editor+ | Create/update translation |
| DELETE | /api/v1/posts/:id/translations/:locale | Editor+ | Delete translation |

### Preview Tokens (4)
| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | /api/v1/posts/:id/preview | Editor+ | Generate preview token |
| GET | /api/v1/posts/:id/preview | Editor+ | List active tokens |
| DELETE | /api/v1/posts/:id/preview | Editor+ | Revoke all tokens |
| DELETE | /api/v1/posts/:id/preview/:token_id | Editor+ | Revoke single token |

## 3. Interface Design

```go
// PostRepository handles sfc_site_posts.
type PostRepository interface {
    List(ctx context.Context, f ListFilter) ([]model.Post, int64, error)
    GetByID(ctx context.Context, id string) (*model.Post, error)
    Create(ctx context.Context, post *model.Post) error
    Update(ctx context.Context, post *model.Post, version int) error // optimistic lock
    SoftDelete(ctx context.Context, id string) error
    Restore(ctx context.Context, id string) error
    SlugExists(ctx context.Context, slug, excludeID string) (bool, error)
    UpdateStatus(ctx context.Context, id string, status model.PostStatus, version int) error
    SyncCategories(ctx context.Context, postID string, categoryIDs []string, primaryID string) error
    SyncTags(ctx context.Context, postID string, tagIDs []string) error
}

// RevisionRepository handles sfc_site_post_revisions.
type RevisionRepository interface {
    List(ctx context.Context, postID string) ([]model.PostRevision, error)
    GetByID(ctx context.Context, id string) (*model.PostRevision, error)
    Create(ctx context.Context, rev *model.PostRevision) error
}

// TranslationRepository handles sfc_site_post_translations.
type TranslationRepository interface {
    List(ctx context.Context, postID string) ([]model.PostTranslation, error)
    Get(ctx context.Context, postID, locale string) (*model.PostTranslation, error)
    Upsert(ctx context.Context, t *model.PostTranslation) error
    Delete(ctx context.Context, postID, locale string) error
}

// PreviewTokenRepository handles sfc_site_preview_tokens.
type PreviewTokenRepository interface {
    List(ctx context.Context, postID string) ([]model.PreviewToken, error)
    Create(ctx context.Context, token *model.PreviewToken) error
    CountActive(ctx context.Context, postID string) (int, error)
    DeleteAll(ctx context.Context, postID string) (int64, error)
    DeleteByID(ctx context.Context, id string) error
    GetByHash(ctx context.Context, hash string) (*model.PreviewToken, error)
}
```

## 4. Key Design Decisions

### 4.1 State Machine

```
Valid transitions:
  draft     → published, scheduled
  scheduled → draft, published
  published → draft, archived
  archived  → published, draft

Any status → soft_delete (via DELETE endpoint, sets deleted_at)
soft_delete → draft (via restore endpoint, clears deleted_at)
```

`validateTransition(current, target PostStatus) error` enforces the above. Invalid transitions return 422.

### 4.2 Optimistic Locking

- `sfc_site_posts.version` is the authoritative version counter
- `Update()` uses `WHERE id = ? AND version = ?`; zero affected rows → `apperror.ErrVersionConflict` (409)
- Client must submit current `version` in PUT body
- Version auto-increments via `BeforeAppendModel` hook (already implemented in model)

### 4.3 Revision Strategy

- Every successful `UpdatePost` creates a `PostRevision` record within the same transaction
- `diff_summary` auto-generated server-side: compare old vs new field values, produce "Updated title, content" style summary
- Rollback = read target revision → create new revision (version continues incrementing) → update post body
- Version numbers never reused

### 4.4 Slug Generation

- English titles only: lowercase + replace spaces/special chars with `-` + trim
- Max 200 characters
- Collision handling: query `SlugExists()`, append `-2`, `-3`... on conflict
- No pinyin library needed

### 4.5 Preview Token Security

- Token format: `sky_preview_{base64url_random_32bytes}`
- Database stores SHA-256 hash only; raw token returned once at creation
- Max 5 active (non-expired) tokens per post
- Token TTL: 24 hours
- Rate limit: 10 creations per user per hour
- Public preview endpoint: hash incoming token → query by hash → validate expiry

### 4.6 Many-to-Many Sync

- `SyncCategories`: delete existing mappings → insert new ones (within transaction)
- `SyncTags`: same delete-then-insert pattern
- `primary_category_id` tracked via `is_primary` flag in `sfc_site_post_category_map`

### 4.7 Meilisearch Integration

- After create/update/publish: async push post data to `posts-{siteSlug}` index (reuse `pkg/search`)
- After soft delete: remove from index
- Searchable fields: title, content (stripped HTML), excerpt, tags

## 5. External Dependencies

| Dependency | Usage | Package |
|------------|-------|---------|
| pkg/search | Meilisearch index sync | Post create/update/delete |
| pkg/cache | Redis operations | View count buffering (future) |
| pkg/crypto | Token generation, SHA-256 | Preview tokens |
| pkg/audit | Audit logging | Via AuditContext middleware |
| pkg/apperror | Sentinel errors | ErrNotFound, ErrVersionConflict |

## 6. Database Tables Involved

All in `site_{slug}` schema:
- `sfc_site_posts` — main posts table (6 indexes)
- `sfc_site_post_translations` — multilingual content
- `sfc_site_post_revisions` — version history (append-only)
- `sfc_site_post_category_map` — many-to-many with categories
- `sfc_site_post_tag_map` — many-to-many with tags
- `sfc_site_preview_tokens` — preview access tokens

Cross-schema reference: `author_id → public.sfc_users(id)`

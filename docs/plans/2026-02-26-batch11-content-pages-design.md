# Batch 11: Content Management Pages Design

> Status: Approved
> Date: 2026-02-26
> Scope: Posts (list/new/edit/revisions), Categories, Tags, Media Library

## Overview

Batch 11 implements all content management frontend pages for Sky Flux CMS.
Backend API is 100% complete (37 content endpoints across 4 modules). This batch
creates the React Islands that consume those endpoints using DashboardLayout.

## Pages & Routes

| Page | Route | Backend Endpoint | Auth |
|------|-------|-----------------|------|
| Posts List | `/dashboard/posts` | `GET /api/v1/site/posts` | JWT |
| New Post | `/dashboard/posts/new` | `POST /api/v1/site/posts` | JWT |
| Edit Post | `/dashboard/posts/[id]/edit` | `GET/PUT /api/v1/site/posts/:id` | JWT |
| Revisions | `/dashboard/posts/[id]/revisions` | `GET /api/v1/site/posts/:id/revisions` | JWT |
| Categories | `/dashboard/categories` | `GET/POST/PUT/DELETE /api/v1/site/categories` | JWT |
| Tags | `/dashboard/tags` | `GET/POST/PUT/DELETE /api/v1/site/tags` | JWT |
| Media Library | `/dashboard/media` | `GET/POST/PUT/DELETE /api/v1/site/media` | JWT |

## Layout

All pages use **DashboardLayout** — sidebar + header + content area.

```
┌─────────────────────────────────────────────────────┐
│ [Sidebar]  │  [Header: breadcrumb + user menu]      │
│            │                                         │
│  Dashboard │  ┌─────────────────────────────────┐   │
│  Posts  ←  │  │  Page title + action buttons     │   │
│  Categories│  │                                   │   │
│  Tags      │  │  <page content>                   │   │
│  Media     │  │                                   │   │
│  ...       │  └─────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

Page content area uses `p-6` padding, consistent with existing dashboard/index.astro.

## File Structure

```
src/
├── pages/dashboard/
│   ├── posts/
│   │   ├── index.astro                  # Posts list page
│   │   ├── new.astro                    # New post page
│   │   └── [id]/
│   │       ├── edit.astro               # Edit post page
│   │       └── revisions.astro          # Revision history page
│   ├── categories/
│   │   └── index.astro                  # Categories tree page
│   ├── tags/
│   │   └── index.astro                  # Tags list page
│   └── media/
│       └── index.astro                  # Media library page
├── components/content/
│   ├── PostsTable.tsx                   # Posts list with filters + bulk actions
│   ├── PostEditor.tsx                   # BlockNote editor + metadata panel
│   ├── PostStatusActions.tsx            # Publish/unpublish/schedule actions
│   ├── RevisionHistory.tsx              # Revision list + rollback
│   ├── CategoryTree.tsx                 # Tree view with drag-and-drop reorder
│   ├── CategoryForm.tsx                 # Create/edit category dialog
│   ├── TagsTable.tsx                    # Tags list with inline edit
│   ├── TagForm.tsx                      # Create/edit tag dialog
│   ├── MediaLibrary.tsx                 # Grid/list view with upload zone
│   ├── MediaUploader.tsx                # Drag-and-drop upload component
│   ├── MediaDetailDialog.tsx            # File detail + metadata edit
│   └── __tests__/
│       ├── PostsTable.test.tsx
│       ├── PostEditor.test.tsx
│       ├── PostStatusActions.test.tsx
│       ├── RevisionHistory.test.tsx
│       ├── CategoryTree.test.tsx
│       ├── CategoryForm.test.tsx
│       ├── TagsTable.test.tsx
│       ├── TagForm.test.tsx
│       ├── MediaLibrary.test.tsx
│       ├── MediaUploader.test.tsx
│       └── MediaDetailDialog.test.tsx
├── components/shared/
│   ├── DataTable.tsx                    # Reusable table with pagination
│   ├── ConfirmDialog.tsx                # Reusable delete confirmation
│   ├── StatusBadge.tsx                  # Post status badge (draft/published/...)
│   └── __tests__/
│       ├── DataTable.test.tsx
│       ├── ConfirmDialog.test.tsx
│       └── StatusBadge.test.tsx
├── hooks/
│   ├── use-pagination.ts               # Page/per_page state management
│   ├── use-debounce.ts                 # Search input debounce (300ms)
│   └── __tests__/
│       ├── use-pagination.test.ts
│       └── use-debounce.test.ts
└── lib/
    └── content-api.ts                   # Content API call wrappers
```

## User Flows

### Posts List

```
/dashboard/posts → GET /site/posts (paginated)
  ├─ Filter by: status dropdown, search input (debounced), category, tag
  ├─ Sort: created_at / published_at / title
  ├─ Bulk actions: delete selected, publish selected
  ├─ Click row → /dashboard/posts/:id/edit
  └─ "New Post" button → /dashboard/posts/new
```

### Post Editor (New + Edit)

```
/dashboard/posts/new → empty BlockNote editor
/dashboard/posts/:id/edit → GET /site/posts/:id → populate editor

  ┌──────────────────────────────┬──────────────┐
  │  [Title input]               │  Status      │
  │                              │  [Publish]   │
  │  [BlockNote editor]          │  Categories  │
  │  (content_json)              │  Tags        │
  │                              │  Cover Image │
  │                              │  Excerpt     │
  │                              │  SEO Panel   │
  │                              │  Slug        │
  └──────────────────────────────┴──────────────┘

  ├─ Auto-save: 30s debounce → PUT /site/posts/:id (draft only)
  ├─ Manual save: Ctrl+S / Save button
  ├─ Publish: POST /site/posts/:id/publish
  ├─ Schedule: status=scheduled + scheduled_at
  ├─ Version conflict (409): toast + "Refresh" button
  └─ Unsaved changes: beforeunload warning
```

### Category Management

```
/dashboard/categories → GET /site/categories (tree)
  ├─ Tree display with expand/collapse
  ├─ Drag-and-drop reorder → PUT /site/categories/reorder
  ├─ Click "Add" → dialog with CategoryForm
  ├─ Click "Edit" → dialog with CategoryForm (pre-filled)
  ├─ Click "Delete" → ConfirmDialog
  │   ├─ Leaf: DELETE /site/categories/:id → success
  │   └─ Has children: 409 error → show message
  └─ Post count badge per category
```

### Tags Management

```
/dashboard/tags → GET /site/tags (paginated table)
  ├─ Search: debounced input → GET /site/tags?q=xxx
  ├─ Sort: post_count / name / created_at
  ├─ Click "Add" → dialog with TagForm
  ├─ Click "Edit" → dialog with TagForm (pre-filled)
  └─ Click "Delete" → ConfirmDialog → DELETE /site/tags/:id
```

### Media Library

```
/dashboard/media → GET /site/media (grid/list toggle)
  ├─ Upload zone: react-dropzone (drag + click)
  │   └─ POST /site/media (multipart/form-data)
  ├─ Filter: media_type (image/video/document)
  ├─ Search: debounced input → GET /site/media?q=xxx
  ├─ Click thumbnail → MediaDetailDialog
  │   ├─ Preview, metadata display
  │   ├─ Edit alt_text/title → PUT /site/media/:id
  │   └─ Delete → ConfirmDialog
  │       ├─ No refs: DELETE /site/media/:id → success
  │       └─ Has refs: 409 → show referencing posts, offer force delete
  ├─ Select multiple → batch delete toolbar
  │   └─ DELETE /site/media/batch
  └─ View toggle: grid (thumbnail cards) / list (table rows)
```

## Component Design

### PostsTable.tsx

- TanStack Table with server-side pagination
- Column: checkbox, title, status (StatusBadge), author, categories, published_at, actions
- Filter bar: status Select, search Input (debounced 300ms), category ComboBox
- Bulk action toolbar: appears when checkboxes selected
- Empty state: illustration + "Create your first post" CTA
- Loading: skeleton rows

### PostEditor.tsx

- Two-column layout: editor (left ~70%) + metadata panel (right ~30%)
- Title: large Input at top, auto-generates slug
- Content: BlockNote editor with `content_json` (JSON) as source of truth
- Metadata panel (right sidebar, scrollable):
  - Status section: current status badge + action buttons (publish/unpublish/schedule)
  - Categories: multi-select with search (from GET /categories tree, flattened)
  - Tags: multi-select with autocomplete (GET /tags/suggest)
  - Cover image: click to select from MediaLibrary or upload
  - Excerpt: textarea
  - SEO: collapsible panel (meta_title, meta_description, og_image_url)
  - Slug: auto-generated, editable
- Auto-save: `useEditorStore.saveDraft()` locally + PUT API every 30s for drafts
- Version: track `version` field, submit with PUT for optimistic locking
- Unsaved indicator: dot in browser tab title

### PostStatusActions.tsx

- Contextual buttons based on current status:
  - draft → "Publish" (primary), "Schedule" (secondary)
  - scheduled → "Publish Now", "Revert to Draft"
  - published → "Unpublish", "Revert to Draft"
  - archived → "Republish", "Revert to Draft"
- Schedule: DateTimePicker for `scheduled_at` (future only)
- Each action calls dedicated endpoint (publish/unpublish/revert-to-draft)

### RevisionHistory.tsx

- Timeline list of revisions from GET /posts/:id/revisions
- Each entry: version number, editor name, diff_summary, timestamp
- "Rollback" button per revision → ConfirmDialog → POST rollback
- Current version highlighted

### CategoryTree.tsx

- Recursive tree component rendering nested categories
- @dnd-kit sortable for drag-and-drop reorder within same level
- Expand/collapse toggle per node with children
- Inline action buttons: edit, add child, delete
- Post count badge (muted text)
- Add root category button at top

### CategoryForm.tsx

- Dialog form (shadcn Dialog)
- Fields: name (Input), slug (Input, auto-generated), parent (Select from tree), description (Textarea), sort_order (Input number)
- Validation: name required, slug regex `^[a-z0-9-]+$`
- Mode: create (POST) / edit (PUT) via prop

### TagsTable.tsx

- DataTable with columns: name, slug, post_count, created_at, actions
- Search bar with debounce
- Sort by post_count (default desc) or name
- Inline actions: edit (opens TagForm dialog), delete (ConfirmDialog)

### TagForm.tsx

- Dialog form
- Fields: name (Input), slug (Input, auto-generated from name)
- Validation: name required
- Mode: create (POST) / edit (PUT)

### MediaLibrary.tsx

- View toggle: grid (default, 4-col thumbnail cards) / list (table rows)
- MediaUploader at top (collapsible)
- Filter bar: media_type Select, search Input
- Grid view: thumbnail + filename + size overlay
- List view: DataTable with thumbnail, filename, type, size, date columns
- Multi-select: checkbox on each item, toolbar appears with batch delete
- Click item → MediaDetailDialog

### MediaUploader.tsx

- react-dropzone zone: dashed border, icon, "Drop files here or click to upload"
- Accepted types: image/*, video/*, application/pdf, .doc, .docx
- Max file size: 50MB (configurable)
- Progress bar per file during upload
- POST /site/media (multipart/form-data) — needs FormData bypass in api-client
- Multiple file upload support
- Upload complete → auto-refresh media list

### MediaDetailDialog.tsx

- Dialog with image preview (or file icon for non-images)
- Info: filename, type, size, dimensions, upload date, reference count
- Editable fields: alt_text (Input), title (Input) → PUT /site/media/:id
- Referencing posts list (if any) with links
- Delete button → ConfirmDialog (with force option if referenced)

## Shared Components

### DataTable.tsx

- Wrapper around @tanstack/react-table
- Server-side pagination via `use-pagination` hook
- Configurable columns via column definitions
- Loading state: skeleton rows
- Empty state: customizable message + optional CTA
- Checkbox column for multi-select (optional)
- Sort indicator headers

### ConfirmDialog.tsx

- AlertDialog from shadcn/ui
- Props: title, description, onConfirm, variant (danger/warning)
- Danger variant: red destructive button
- Loading state on confirm button

### StatusBadge.tsx

- Badge component with color per status:
  - draft → gray
  - scheduled → blue
  - published → green
  - archived → yellow
  - deleted → red

## API Client Extension

`content-api.ts` wraps all content management endpoints:

```typescript
// Posts
postsApi.list(params: PostListParams) → PaginatedResponse<PostSummary>
postsApi.get(id: string) → Post
postsApi.create(data: CreatePostDTO) → Post
postsApi.update(id: string, data: UpdatePostDTO) → Post
postsApi.delete(id: string) → void
postsApi.publish(id: string) → Post
postsApi.unpublish(id: string) → Post
postsApi.revertToDraft(id: string) → Post
postsApi.restore(id: string) → Post
postsApi.getRevisions(id: string) → Revision[]
postsApi.rollback(id: string, revId: string) → Post

// Categories
categoriesApi.tree() → CategoryNode[]
categoriesApi.get(id: string) → Category
categoriesApi.create(data: CreateCategoryDTO) → Category
categoriesApi.update(id: string, data: UpdateCategoryDTO) → Category
categoriesApi.delete(id: string) → void
categoriesApi.reorder(orders: ReorderItem[]) → void

// Tags
tagsApi.list(params: TagListParams) → PaginatedResponse<Tag>
tagsApi.get(id: string) → Tag
tagsApi.create(data: CreateTagDTO) → Tag
tagsApi.update(id: string, data: UpdateTagDTO) → Tag
tagsApi.delete(id: string) → void
tagsApi.suggest(q: string) → Tag[]

// Media
mediaApi.list(params: MediaListParams) → PaginatedResponse<MediaFile>
mediaApi.get(id: string) → MediaFileDetail
mediaApi.upload(file: File, altText?: string) → MediaFile
mediaApi.updateMeta(id: string, data: UpdateMediaDTO) → MediaFile
mediaApi.delete(id: string, force?: boolean) → void
mediaApi.batchDelete(ids: string[], force?: boolean) → BatchDeleteResult
```

Note: `mediaApi.upload` uses FormData instead of JSON — requires extending api-client.ts
with a `requestFormData` helper that skips JSON Content-Type header.

## New Dependencies

| Package | Purpose |
|---------|---------|
| `@blocknote/core` | Block editor engine |
| `@blocknote/react` | React bindings for BlockNote |
| `@blocknote/shadcn` | shadcn/ui theme for BlockNote |
| `@dnd-kit/core` | Drag-and-drop primitives |
| `@dnd-kit/sortable` | Sortable preset |
| `@dnd-kit/utilities` | CSS transform utilities |
| `react-dropzone` | File drag-and-drop upload |
| `@tanstack/react-table` | Headless table with pagination/sorting |

## i18n

Extend zh-CN.json and en.json with content-specific keys:

```json
{
  "content": {
    "posts": "Posts",
    "newPost": "New Post",
    "editPost": "Edit Post",
    "postTitle": "Title",
    "postTitlePlaceholder": "Enter post title...",
    "postContent": "Content",
    "postExcerpt": "Excerpt",
    "postSlug": "Slug",
    "postStatus": "Status",
    "postCoverImage": "Cover Image",
    "postCategories": "Categories",
    "postTags": "Tags",
    "postSeo": "SEO Settings",
    "postMetaTitle": "Meta Title",
    "postMetaDescription": "Meta Description",
    "postOgImage": "OG Image URL",
    "postScheduledAt": "Schedule Time",
    "statusDraft": "Draft",
    "statusPublished": "Published",
    "statusScheduled": "Scheduled",
    "statusArchived": "Archived",
    "publish": "Publish",
    "unpublish": "Unpublish",
    "schedule": "Schedule",
    "revertToDraft": "Revert to Draft",
    "restore": "Restore",
    "autoSaved": "Auto-saved",
    "unsavedChanges": "Unsaved changes",
    "versionConflict": "This post was modified by another user. Please refresh.",
    "revisions": "Revisions",
    "revisionVersion": "Version {{version}}",
    "rollback": "Rollback to this version",
    "rollbackConfirm": "Are you sure you want to rollback to version {{version}}?",
    "categories": "Categories",
    "categoryName": "Name",
    "categorySlug": "Slug",
    "categoryDescription": "Description",
    "categoryParent": "Parent Category",
    "categoryNone": "None (Root)",
    "addCategory": "Add Category",
    "addSubcategory": "Add Subcategory",
    "editCategory": "Edit Category",
    "deleteCategory": "Delete Category",
    "deleteCategoryConfirm": "Delete category \"{{name}}\"? Posts in this category will be unlinked.",
    "categoryHasChildren": "Cannot delete: this category has subcategories. Remove them first.",
    "tags": "Tags",
    "tagName": "Name",
    "tagSlug": "Slug",
    "addTag": "Add Tag",
    "editTag": "Edit Tag",
    "deleteTag": "Delete Tag",
    "deleteTagConfirm": "Delete tag \"{{name}}\"?",
    "postCount": "{{count}} posts",
    "media": "Media Library",
    "uploadMedia": "Upload Files",
    "dropzone": "Drop files here or click to upload",
    "uploading": "Uploading...",
    "uploadProgress": "{{percent}}%",
    "mediaDetail": "File Details",
    "mediaAltText": "Alt Text",
    "mediaTitle": "Title",
    "mediaFileName": "File Name",
    "mediaFileSize": "File Size",
    "mediaDimensions": "Dimensions",
    "mediaType": "Type",
    "mediaReferences": "Referenced by",
    "deleteMedia": "Delete File",
    "deleteMediaConfirm": "Delete \"{{name}}\"?",
    "deleteMediaReferenced": "This file is referenced by {{count}} posts. Force delete?",
    "batchDelete": "Delete Selected",
    "batchDeleteConfirm": "Delete {{count}} selected files?",
    "gridView": "Grid View",
    "listView": "List View",
    "noPostsFound": "No posts found",
    "createFirstPost": "Create your first post",
    "noCategoriesFound": "No categories yet",
    "noTagsFound": "No tags yet",
    "noMediaFound": "No media files yet",
    "searchPlaceholder": "Search...",
    "filterByStatus": "Filter by status",
    "filterByType": "Filter by type",
    "sortBy": "Sort by",
    "confirmDelete": "Confirm Delete",
    "deleteWarning": "This action cannot be undone.",
    "selected": "{{count}} selected"
  }
}
```

## Testing Strategy (TDD)

Each React component: test file written alongside implementation.

- PostsTable: list rendering, filter/sort, pagination, bulk select, empty state (~15 tests)
- PostEditor: title input, editor mount, save, auto-save trigger, version conflict (~15 tests)
- PostStatusActions: status transitions, schedule picker, confirm dialogs (~10 tests)
- RevisionHistory: list rendering, rollback confirm, current version highlight (~8 tests)
- CategoryTree: tree rendering, expand/collapse, reorder, CRUD actions (~12 tests)
- CategoryForm: validation, create/edit modes, parent select (~8 tests)
- TagsTable: list, search, sort, CRUD actions (~10 tests)
- TagForm: validation, create/edit modes (~6 tests)
- MediaLibrary: grid/list toggle, filter, multi-select, upload trigger (~12 tests)
- MediaUploader: dropzone, file validation, progress, completion (~10 tests)
- MediaDetailDialog: display, edit metadata, delete with refs (~8 tests)
- Shared (DataTable, ConfirmDialog, StatusBadge): rendering, interactions (~10 tests)
- Hooks (use-pagination, use-debounce): state updates, timing (~6 tests)

Total estimate: ~120-130 tests

## Agent Teams Division

| Agent | Scope | Deliverables |
|-------|-------|-------------|
| Agent 1 (infra) | Shared components, hooks, content-api.ts, i18n keys, dependency install | Foundation layer |
| Agent 2 (posts-list) | PostsTable + PostStatusActions + posts list page + tests | Posts list |
| Agent 3 (post-editor) | PostEditor + RevisionHistory + new/edit/revisions pages + tests | Post editor |
| Agent 4 (taxonomy-media) | CategoryTree + CategoryForm + TagsTable + TagForm + MediaLibrary + MediaUploader + MediaDetailDialog + pages + tests | Categories + Tags + Media |

Dependencies: Agents 2, 3, 4 depend on Agent 1 completing shared infra first.
Agent 1 delivers: DataTable, ConfirmDialog, StatusBadge, use-pagination, use-debounce, content-api.ts, i18n keys, all new npm packages installed.

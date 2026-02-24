# Site Schema Prefix Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `sfc_site_` prefix to all 17 site schema tables, indexes, triggers, and constraints for naming consistency with public schema's `sfc_` prefix.

**Architecture:** Direct in-place rename of DDL template and documentation. No migration needed (pre-production, no data). Two Go files + 10 docs + CLAUDE.md.

**Tech Stack:** Go (string template), PostgreSQL DDL, Markdown documentation

---

### Task 1: Update `internal/schema/template.go` — DDL Template

**Files:**
- Modify: `internal/schema/template.go`

This is the core change. All 17 `CREATE TABLE`, indexes, triggers, constraints, and FK references need renaming.

**Step 1: Apply table name replacements**

Replace every table name in the DDL template string. Apply these substitutions in order (longer names first to avoid partial matches):

```
post_category_map → sfc_site_post_category_map
post_translations → sfc_site_post_translations
post_tag_map      → sfc_site_post_tag_map
post_revisions    → sfc_site_post_revisions
preview_tokens    → sfc_site_preview_tokens
media_files       → sfc_site_media_files
menu_items        → sfc_site_menu_items
post_types        → sfc_site_post_types
audit_logs        → sfc_site_audits
system_configs    → sfc_site_configs
categories        → sfc_site_categories
comments          → sfc_site_comments
redirects         → sfc_site_redirects
menus             → sfc_site_menus
posts             → sfc_site_posts
tags              → sfc_site_tags
api_keys          → sfc_site_api_keys
```

Special cases (different target name, not just adding prefix):
- `audit_logs` → `sfc_site_audits` (not `sfc_site_audit_logs`)
- `system_configs` → `sfc_site_configs` (not `sfc_site_system_configs`)

**Step 2: Apply index name replacements**

All index names follow the pattern `idx_{table}_{col}` → `idx_sfc_site_{table}_{col}`:

```
idx_post_types_slug        → idx_sfc_site_post_types_slug
idx_posts_slug             → idx_sfc_site_posts_slug
idx_posts_author           → idx_sfc_site_posts_author
idx_posts_status           → idx_sfc_site_posts_status
idx_posts_published        → idx_sfc_site_posts_published
idx_posts_extra            → idx_sfc_site_posts_extra
idx_posts_scheduled        → idx_sfc_site_posts_scheduled
idx_pt_post_locale         → idx_sfc_site_pt_post_locale
idx_revisions_post         → idx_sfc_site_revisions_post
idx_revisions_editor       → idx_sfc_site_revisions_editor
idx_categories_parent      → idx_sfc_site_categories_parent
idx_categories_path        → idx_sfc_site_categories_path
idx_categories_slug        → idx_sfc_site_categories_slug
idx_tags_name_trgm         → idx_sfc_site_tags_name_trgm
idx_pcm_category           → idx_sfc_site_pcm_category
idx_ptm_tag                → idx_sfc_site_ptm_tag
idx_media_uploader         → idx_sfc_site_media_uploader
idx_media_type             → idx_sfc_site_media_type
idx_media_name_trgm        → idx_sfc_site_media_name_trgm
idx_comments_post_status   → idx_sfc_site_comments_post_status
idx_comments_parent        → idx_sfc_site_comments_parent
idx_comments_moderation    → idx_sfc_site_comments_moderation
idx_comments_email         → idx_sfc_site_comments_email
idx_comments_user          → idx_sfc_site_comments_user
idx_menu_items_menu        → idx_sfc_site_menu_items_menu
idx_menu_items_parent      → idx_sfc_site_menu_items_parent
idx_menu_items_reference   → idx_sfc_site_menu_items_reference
idx_redirects_source       → idx_sfc_site_redirects_source
idx_redirects_created      → idx_sfc_site_redirects_created
idx_preview_tokens_post    → idx_sfc_site_preview_tokens_post
idx_preview_tokens_hash    → idx_sfc_site_preview_tokens_hash
idx_preview_tokens_expires → idx_sfc_site_preview_tokens_expires
idx_apikeys_owner          → idx_sfc_site_apikeys_owner
idx_apikeys_hash           → idx_sfc_site_apikeys_hash
```

**Step 3: Apply trigger name replacements**

```
trg_post_types_updated_at         → trg_sfc_site_post_types_updated_at
trg_posts_updated_at              → trg_sfc_site_posts_updated_at
trg_post_translations_updated_at  → trg_sfc_site_post_translations_updated_at
trg_categories_updated_at         → trg_sfc_site_categories_updated_at
trg_media_updated_at              → trg_sfc_site_media_updated_at
trg_comments_updated_at           → trg_sfc_site_comments_updated_at
trg_menus_updated_at              → trg_sfc_site_menus_updated_at
trg_menu_items_updated_at         → trg_sfc_site_menu_items_updated_at
trg_redirects_updated_at          → trg_sfc_site_redirects_updated_at
```

**Step 4: Apply constraint and FK name replacements**

```
fk_posts_cover_image                 → fk_sfc_site_posts_cover_image
chk_comment_content_length           → chk_sfc_site_comment_content_length
chk_comment_guest                    → chk_sfc_site_comment_guest
chk_redirect_status_code             → chk_sfc_site_redirect_status_code
```

**Step 5: Update system_configs INSERT statement**

The initial data `INSERT INTO {schema}.system_configs` → `INSERT INTO {schema}.sfc_site_configs`.

**Step 6: Commit**

```bash
git add internal/schema/template.go
git commit -m "refactor: add sfc_site_ prefix to all site schema tables in DDL template"
```

---

### Task 2: Update `internal/schema/migrate.go` — Audit Partition Function

**Files:**
- Modify: `internal/schema/migrate.go`

**Step 1: Rename audit_logs references to sfc_site_audits**

Line 65 comment: `audit_logs` → `sfc_site_audits`
Line 74: `audit_logs_%s` → `sfc_site_audits_%s`
Line 76: `%s.audit_logs` → `%s.sfc_site_audits`
Line 88 comment: update reference
Line 90: `idx_audit_actor` → `idx_sfc_site_audits_actor`, `%s.audit_logs` → `%s.sfc_site_audits`
Line 91: `idx_audit_resource` → `idx_sfc_site_audits_resource`, `%s.audit_logs` → `%s.sfc_site_audits`
Line 92: `idx_audit_time` → `idx_sfc_site_audits_time`, `%s.audit_logs` → `%s.sfc_site_audits`

**Step 2: Commit**

```bash
git add internal/schema/migrate.go
git commit -m "refactor: rename audit_logs to sfc_site_audits in partition function"
```

---

### Task 3: Verify Go Build

**Step 1: Run go build**

```bash
go build ./...
```

Expected: clean build, no errors.

**Step 2: Run go vet**

```bash
go vet ./...
```

Expected: no issues.

---

### Task 4: Update `docs/database.md` — ER Diagram + DDL

**Files:**
- Modify: `docs/database.md`

This file has ~157 occurrences. It contains the ER diagram and complete DDL.

**Step 1: Update ER diagram entity names**

In the `erDiagram` mermaid block, replace all site schema table names with their `sfc_site_` prefixed versions. Keep public schema names (`sfc_users`, `sfc_sites`, etc.) unchanged.

Special cases in ER diagram:
- `system_configs_site` → `sfc_site_configs`
- `audit_logs` → `sfc_site_audits`

**Step 2: Update ER diagram relationship lines**

All relationship lines referencing site schema tables need updating.

**Step 3: Update Site Schema DDL section**

The DDL section under "站点 Schema DDL" should match the updated `template.go` exactly. Replace all table names, index names, trigger names, and constraint names.

**Step 4: Update any prose references**

Scan for table names in descriptive text and update them.

**Step 5: Commit**

```bash
git add docs/database.md
git commit -m "docs: update database.md with sfc_site_ prefixed table names"
```

---

### Task 5: Update Remaining Documentation Files

**Files:**
- Modify: `docs/api.md` (~207 occurrences)
- Modify: `docs/architecture.md` (~69 occurrences)
- Modify: `docs/standard.md` (~66 occurrences)
- Modify: `docs/testing.md` (~57 occurrences)
- Modify: `docs/story.md` (~53 occurrences)
- Modify: `docs/security.md` (~21 occurrences)
- Modify: `docs/deployment.md` (~9 occurrences)
- Modify: `docs/prd.md` (~4 occurrences)
- Modify: `docs/setup.md` (~1 occurrence)

**Important:** Only rename when used as a table/entity name (in code blocks, SQL, table definitions, bun model tags). Do NOT rename when used as generic English words in natural language prose (e.g., "用户可以发布 posts" — though most Chinese docs use Chinese terms for these).

Apply the same mapping as Task 1. For each file:

**Step 1: Read the file**
**Step 2: Apply context-sensitive replacements**
**Step 3: Verify no broken references**

Process files from most to fewest occurrences for maximum impact first.

**Step 4: Commit all docs together**

```bash
git add docs/api.md docs/architecture.md docs/standard.md docs/testing.md docs/story.md docs/security.md docs/deployment.md docs/prd.md docs/setup.md
git commit -m "docs: update all design docs with sfc_site_ prefixed table names"
```

---

### Task 6: Update `CLAUDE.md`

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Update multi-site architecture section**

Replace the `site_{slug}` schema table list:

```
- `site_{slug}` schema: sfc_site_posts, sfc_site_categories, sfc_site_tags, sfc_site_media_files, sfc_site_comments, sfc_site_menus, sfc_site_redirects, sfc_site_preview_tokens, sfc_site_api_keys, sfc_site_audits, sfc_site_configs
```

**Step 2: Update bun usage examples if they reference site tables**

Check the uptrace/bun usage section for any site table references.

**Step 3: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with sfc_site_ prefixed site schema tables"
```

---

### Task 7: Final Verification

**Step 1: Grep for orphaned old table names**

```bash
grep -rn '\baudit_logs\b' internal/ docs/ CLAUDE.md --include='*.go' --include='*.md' | grep -v 'sfc_site_audits' | grep -v 'plans/'
grep -rn '\bsystem_configs\b' internal/ docs/ CLAUDE.md --include='*.go' --include='*.md' | grep -v 'sfc_site_configs' | grep -v 'sfc_system_configs' | grep -v 'sfc_configs' | grep -v 'plans/'
```

Expected: no results (all occurrences should be prefixed now, except in the design plan docs).

**Step 2: Verify Go build still passes**

```bash
go build ./...
```

Expected: clean build.

**Step 3: Update memory files**

Update `MEMORY.md` to reflect the new naming convention.

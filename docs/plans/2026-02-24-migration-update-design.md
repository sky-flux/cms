# Migration & DDL 更新设计

> **日期**: 2026-02-24
> **状态**: 已批准
> **范围**: Migration 重写 + Site Schema Template 修复 + 触发器移除 + 文档同步

## 背景

Model 层完成 ENUM→SMALLINT 重构后（commit `0b551a7`），发现 SQL DDL 层（migrations + template.go）存在多处不一致：

1. Migration 1 仍创建 5 个 PostgreSQL ENUM 类型（model 已改用 SMALLINT iota）
2. Migration 2 缺少 `sfc_password_reset_tokens` 表
3. Migration 2 中 `sfc_sites` 缺少 `deleted_at` 列
4. `template.go` 中 5 处引用 ENUM 类型（应为 SMALLINT + CHECK）
5. `template.go` 中索引条件使用 ENUM 字符串（如 `WHERE status = 'published'`）
6. `template.go` 中 2 处使用 `pg_trgm` 扩展索引（违反零扩展原则）
7. `update_updated_at()` 触发器函数及所有 BEFORE UPDATE 触发器是冗余的（updated_at 由 Go 应用层管理）

## 决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 修改策略 | 全部重写 | 项目未上线，无需增量迁移 |
| pg_trgm | 移除 | 统一使用 Meilisearch，PG 保持零扩展 |
| sfc_sites.deleted_at | DDL 添加 | 软删除更安全，防止误删站点后无法恢复 |
| Migration 分组 | 4 个文件 | Core Tables / RBAC / Placeholder / Seed |
| 触发器 | 全部移除 | updated_at 由 bun ORM 应用层统一管理，触发器是冗余开销 |

## Migration 文件结构

### 000001_create_core_tables.go

用户、站点、认证相关表（6 张）。**无触发器函数、无 ENUM 类型**。

| 表名 | 变更说明 |
|------|---------|
| sfc_users | 无变更 |
| sfc_sites | **新增 `deleted_at TIMESTAMPTZ` 列** |
| sfc_refresh_tokens | 无变更 |
| **sfc_password_reset_tokens** | **新增表**（id, user_id, token_hash, expires_at, used_at, created_at） |
| sfc_user_totp | 无变更 |
| sfc_configs | 无变更（含 `system.installed` seed） |

`sfc_password_reset_tokens` 索引：
- `idx_sfc_prt_user` ON (user_id)
- `idx_sfc_prt_token` ON (token_hash)
- `idx_sfc_prt_expires` ON (expires_at) WHERE used_at IS NULL

### 000002_create_rbac_tables.go

RBAC 体系 9 张表（无触发器）：
- sfc_roles, sfc_user_roles, sfc_apis, sfc_role_apis
- sfc_menus（后台管理菜单）, sfc_role_menus
- sfc_role_templates, sfc_role_template_apis, sfc_role_template_menus

无结构变更，仅从原 migration 2 拆出 + 删除所有 CREATE TRIGGER 语句。

### 000003_site_schema_placeholder.go

No-op 占位符，与现有 migration 3 相同。

### 000004_seed_rbac_builtins.go

Seed 4 内置角色 + 4 权限模板，与现有 migration 4 相同。

## template.go 改造

### ENUM → SMALLINT + CHECK（5 处）

| 表 | 列 | 旧 DDL | 新 DDL |
|----|-----|--------|--------|
| sfc_site_posts | status | `post_status NOT NULL DEFAULT 'draft'` | `SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4)` |
| sfc_site_media_files | media_type | `media_type NOT NULL DEFAULT 'other'` | `SMALLINT NOT NULL DEFAULT 5 CHECK (media_type BETWEEN 1 AND 5)` |
| sfc_site_comments | status | `comment_status NOT NULL DEFAULT 'pending'` | `SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4)` |
| sfc_site_menu_items | type | `menu_item_type NOT NULL DEFAULT 'custom'` | `SMALLINT NOT NULL DEFAULT 1 CHECK (type BETWEEN 1 AND 5)` |
| sfc_site_audits | action | `log_action NOT NULL` | `SMALLINT NOT NULL CHECK (action BETWEEN 1 AND 11)` |

### 索引条件更新

```sql
-- sfc_site_posts
WHERE status = 'published' → WHERE status = 3
WHERE status = 'scheduled' → WHERE status = 2

-- sfc_site_comments
WHERE status = 'pending'   → WHERE status = 1
```

### 移除 pg_trgm 索引（2 处）

```sql
-- 删除:
CREATE INDEX idx_sfc_site_tags_name_trgm ON {schema}.sfc_site_tags USING gin(name gin_trgm_ops);
CREATE INDEX idx_sfc_site_media_name_trgm ON {schema}.sfc_site_media_files USING gin(file_name gin_trgm_ops);
```

### 移除所有 BEFORE UPDATE 触发器（9 处）

删除 template.go 中全部 `CREATE TRIGGER trg_xxx_updated_at` 语句（9 处）。

## 其他文件修改

### docs/database.md

- `sfc_sites` 表定义添加 `deleted_at TIMESTAMPTZ` 列
- 更新迁移文件列表（4 个文件新命名）

### internal/model/role_template.go

- alias `rt` → `rtpl`（避免与 RefreshToken 的 `rt` 冲突）

## 不修改的文件

- `internal/schema/migrate.go` — 逻辑正确
- `internal/schema/validate.go` — 无关
- `internal/model/enums.go` — 已是 SMALLINT
- 其他 model 文件 — 已与设计对齐

## 修改文件清单

| 文件 | 动作 |
|------|------|
| `migrations/20260224000001_create_enums_and_functions.go` | 删除 |
| `migrations/20260224000002_create_public_schema.go` | 删除 |
| `migrations/20260224000003_create_site_template.go` | 删除 |
| `migrations/20260224000004_seed_rbac_builtins.go` | 删除 |
| `migrations/20260224000001_create_core_tables.go` | 新建 |
| `migrations/20260224000002_create_rbac_tables.go` | 新建 |
| `migrations/20260224000003_site_schema_placeholder.go` | 新建 |
| `migrations/20260224000004_seed_rbac_builtins.go` | 新建 |
| `internal/schema/template.go` | 修改 |
| `docs/database.md` | 修改 |
| `internal/model/role_template.go` | 修改 |

# Site Schema 表名前缀统一设计

**日期**: 2026-02-24
**状态**: 已批准

## 动机

统一全项目表命名前缀策略，消除 public schema (`sfc_` 前缀) 与 site schema (无前缀) 之间的命名不对称。加前缀后便于在 SQL 日志、`pg_stat_*` 视图、监控面板中快速识别 CMS 自有表。

## 命名规则

| 对象类型 | 模式 | 示例 |
|---------|------|------|
| 表名 | `sfc_site_{entity}` | `sfc_site_posts` |
| 索引 | `idx_sfc_site_{entity}_{col}` | `idx_sfc_site_posts_slug` |
| 触发器 | `trg_sfc_site_{entity}_{purpose}` | `trg_sfc_site_posts_updated_at` |
| 约束 | `chk_sfc_site_{entity}_{rule}` | `chk_sfc_site_comments_content_length` |
| 外键 | `fk_sfc_site_{entity}_{ref}` | `fk_sfc_site_posts_cover_image` |

## 完整表名映射 (17 表)

| # | 原名 | 新名 |
|---|------|------|
| 1 | post_types | sfc_site_post_types |
| 2 | posts | sfc_site_posts |
| 3 | post_translations | sfc_site_post_translations |
| 4 | post_revisions | sfc_site_post_revisions |
| 5 | categories | sfc_site_categories |
| 6 | tags | sfc_site_tags |
| 7 | post_category_map | sfc_site_post_category_map |
| 8 | post_tag_map | sfc_site_post_tag_map |
| 9 | media_files | sfc_site_media_files |
| 10 | comments | sfc_site_comments |
| 11 | menus | sfc_site_menus |
| 12 | menu_items | sfc_site_menu_items |
| 13 | redirects | sfc_site_redirects |
| 14 | preview_tokens | sfc_site_preview_tokens |
| 15 | api_keys | sfc_site_api_keys |
| 16 | audit_logs | sfc_site_audits |
| 17 | system_configs | sfc_site_configs |

> 注意 #16 和 #17 同时做了简化重命名。

## 受影响文件

| 文件 | 变更内容 |
|------|---------|
| `internal/schema/template.go` | 所有 CREATE TABLE / INDEX / TRIGGER / CONSTRAINT 名称 |
| `docs/database.md` | ER 图表名 + 完整 DDL |
| `CLAUDE.md` | 多站点架构描述中的 site schema 表名列表 |
| 其他 docs (`api.md`, `architecture.md`, `story.md` 等) | 扫描并替换所有 site schema 表名引用 |

## 不受影响

- `migrations/20260224000002_create_public_schema.go` — public schema 表不变
- `migrations/20260224000001_create_enums_and_functions.go` — 枚举/函数不变
- Go 模块目录名 — `internal/post/`, `internal/category/` 等保持不变
- `internal/model/` — 未来 bun model 的 struct tag 将直接使用新名称

## 实施策略

项目处于脚手架阶段（无生产数据），直接原地修改，无需 ALTER TABLE RENAME 迁移脚本。

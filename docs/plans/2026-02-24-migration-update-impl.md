# Migration & DDL 更新实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 重写全部 migration 文件 + 修复 site schema template DDL，使 SQL DDL 与 model 层的 SMALLINT+CHECK 策略完全对齐，移除所有冗余触发器。

**Architecture:** 删除现有 4 个 migration 文件，重写为 4 个（core tables / RBAC tables / placeholder / seed）。同步修改 `template.go`：5 处 ENUM→SMALLINT、移除 2 处 pg_trgm 索引、移除 9 处触发器。更新 database.md 和 model alias。

**Tech Stack:** Go / uptrace/bun migrations / PostgreSQL 18 / raw SQL DDL

**Design doc:** `docs/plans/2026-02-24-migration-update-design.md`

---

### Task 1: 删除旧 migration 文件

**Files:**
- Delete: `migrations/20260224000001_create_enums_and_functions.go`
- Delete: `migrations/20260224000002_create_public_schema.go`
- Delete: `migrations/20260224000003_create_site_template.go`
- Delete: `migrations/20260224000004_seed_rbac_builtins.go`

**Step 1: 删除 4 个旧文件**

```bash
rm migrations/20260224000001_create_enums_and_functions.go \
   migrations/20260224000002_create_public_schema.go \
   migrations/20260224000003_create_site_template.go \
   migrations/20260224000004_seed_rbac_builtins.go
```

**Step 2: 验证 migrations/ 目录只剩 main.go**

```bash
ls migrations/
```

Expected: 只有 `main.go`

---

### Task 2: 创建 migration 1 — Core Tables

**Files:**
- Create: `migrations/20260224000001_create_core_tables.go`

**Step 1: 创建文件**

包含 6 张表：sfc_users, sfc_sites（含 deleted_at）, sfc_refresh_tokens, sfc_password_reset_tokens（新增）, sfc_user_totp, sfc_configs。无触发器、无 ENUM。

```go
package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
-- 1. 用户表
CREATE TABLE public.sfc_users (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name  VARCHAR(100) NOT NULL,
    avatar_url    TEXT,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_sfc_users_email ON public.sfc_users(email) WHERE deleted_at IS NULL;

-- 2. 站点注册表
CREATE TABLE public.sfc_sites (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    name            VARCHAR(200) NOT NULL,
    slug            VARCHAR(50)  NOT NULL UNIQUE,
    domain          VARCHAR(255) UNIQUE,
    description     TEXT,
    logo_url        TEXT,
    default_locale  VARCHAR(10)  NOT NULL DEFAULT 'zh-CN',
    timezone        VARCHAR(50)  NOT NULL DEFAULT 'Asia/Shanghai',
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    settings        JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_sfc_sites_slug CHECK (slug ~ '^[a-z0-9_]{3,50}$')
);

CREATE INDEX idx_sfc_sites_domain ON public.sfc_sites(domain) WHERE domain IS NOT NULL;
CREATE INDEX idx_sfc_sites_active ON public.sfc_sites(is_active) WHERE is_active = TRUE;

-- 3. 刷新令牌表
CREATE TABLE public.sfc_refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id     UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN NOT NULL DEFAULT FALSE,
    ip_address  INET,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_rt_user_id ON public.sfc_refresh_tokens(user_id);
CREATE INDEX idx_sfc_rt_token   ON public.sfc_refresh_tokens(token_hash);

-- 4. 密码重置令牌表
CREATE TABLE public.sfc_password_reset_tokens (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id     UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_prt_user    ON public.sfc_password_reset_tokens(user_id);
CREATE INDEX idx_sfc_prt_token   ON public.sfc_password_reset_tokens(token_hash);
CREATE INDEX idx_sfc_prt_expires ON public.sfc_password_reset_tokens(expires_at)
    WHERE used_at IS NULL;

-- 5. 用户 TOTP 双因素认证表
CREATE TABLE public.sfc_user_totp (
    id                UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id           UUID NOT NULL UNIQUE REFERENCES public.sfc_users(id) ON DELETE CASCADE,
    secret_encrypted  TEXT NOT NULL,
    backup_codes_hash TEXT[],
    is_enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. 系统配置表（全局级别）
CREATE TABLE public.sfc_configs (
    key         VARCHAR(100) PRIMARY KEY,
    value       JSONB NOT NULL,
    description TEXT,
    updated_by  UUID REFERENCES public.sfc_users(id),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO public.sfc_configs (key, value, description) VALUES
('system.installed', 'false', '系统是否已通过安装向导初始化')
ON CONFLICT (key) DO NOTHING;
		`)
		if err != nil {
			return fmt.Errorf("create core tables: %w", err)
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
DROP TABLE IF EXISTS public.sfc_configs CASCADE;
DROP TABLE IF EXISTS public.sfc_user_totp CASCADE;
DROP TABLE IF EXISTS public.sfc_password_reset_tokens CASCADE;
DROP TABLE IF EXISTS public.sfc_refresh_tokens CASCADE;
DROP TABLE IF EXISTS public.sfc_sites CASCADE;
DROP TABLE IF EXISTS public.sfc_users CASCADE;
		`)
		if err != nil {
			return fmt.Errorf("drop core tables: %w", err)
		}
		return nil
	})
}
```

**Step 2: 验证编译通过**

```bash
go build ./migrations/...
```

Expected: 无错误

---

### Task 3: 创建 migration 2 — RBAC Tables

**Files:**
- Create: `migrations/20260224000002_create_rbac_tables.go`

**Step 1: 创建文件**

包含 9 张 RBAC 表。无触发器。

```go
package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
-- 1. 角色定义表
CREATE TABLE public.sfc_roles (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(50)  NOT NULL UNIQUE,
    slug        VARCHAR(50)  NOT NULL UNIQUE,
    description TEXT,
    built_in    BOOLEAN      NOT NULL DEFAULT FALSE,
    status      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 2. 用户-角色分配表
CREATE TABLE public.sfc_user_roles (
    user_id    UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES public.sfc_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_sfc_user_roles_role ON public.sfc_user_roles(role_id);

-- 3. API 端点注册表
CREATE TABLE public.sfc_apis (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    method      VARCHAR(10)  NOT NULL,
    path        VARCHAR(500) NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    "group"     VARCHAR(50)  NOT NULL,
    status      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(method, path)
);

CREATE INDEX idx_sfc_apis_group ON public.sfc_apis("group");

-- 4. 角色-API 权限映射表
CREATE TABLE public.sfc_role_apis (
    role_id UUID NOT NULL REFERENCES public.sfc_roles(id) ON DELETE CASCADE,
    api_id  UUID NOT NULL REFERENCES public.sfc_apis(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, api_id)
);

CREATE INDEX idx_sfc_role_apis_api ON public.sfc_role_apis(api_id);

-- 5. 后台管理菜单表
CREATE TABLE public.sfc_menus (
    id         UUID PRIMARY KEY DEFAULT uuidv7(),
    parent_id  UUID REFERENCES public.sfc_menus(id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    icon       VARCHAR(50),
    path       VARCHAR(200),
    sort_order INT     NOT NULL DEFAULT 0,
    status     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_menus_parent ON public.sfc_menus(parent_id);

-- 6. 角色-菜单可见性映射表
CREATE TABLE public.sfc_role_menus (
    role_id UUID NOT NULL REFERENCES public.sfc_roles(id) ON DELETE CASCADE,
    menu_id UUID NOT NULL REFERENCES public.sfc_menus(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, menu_id)
);

CREATE INDEX idx_sfc_role_menus_menu ON public.sfc_role_menus(menu_id);

-- 7. 权限模板定义表
CREATE TABLE public.sfc_role_templates (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    built_in    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 8. 模板-API 映射表
CREATE TABLE public.sfc_role_template_apis (
    template_id UUID NOT NULL REFERENCES public.sfc_role_templates(id) ON DELETE CASCADE,
    api_id      UUID NOT NULL REFERENCES public.sfc_apis(id) ON DELETE CASCADE,
    PRIMARY KEY (template_id, api_id)
);

-- 9. 模板-菜单映射表
CREATE TABLE public.sfc_role_template_menus (
    template_id UUID NOT NULL REFERENCES public.sfc_role_templates(id) ON DELETE CASCADE,
    menu_id     UUID NOT NULL REFERENCES public.sfc_menus(id) ON DELETE CASCADE,
    PRIMARY KEY (template_id, menu_id)
);
		`)
		if err != nil {
			return fmt.Errorf("create RBAC tables: %w", err)
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
DROP TABLE IF EXISTS public.sfc_role_template_menus CASCADE;
DROP TABLE IF EXISTS public.sfc_role_template_apis CASCADE;
DROP TABLE IF EXISTS public.sfc_role_templates CASCADE;
DROP TABLE IF EXISTS public.sfc_role_menus CASCADE;
DROP TABLE IF EXISTS public.sfc_menus CASCADE;
DROP TABLE IF EXISTS public.sfc_role_apis CASCADE;
DROP TABLE IF EXISTS public.sfc_apis CASCADE;
DROP TABLE IF EXISTS public.sfc_user_roles CASCADE;
DROP TABLE IF EXISTS public.sfc_roles CASCADE;
		`)
		if err != nil {
			return fmt.Errorf("drop RBAC tables: %w", err)
		}
		return nil
	})
}
```

**Step 2: 验证编译通过**

```bash
go build ./migrations/...
```

Expected: 无错误

---

### Task 4: 创建 migration 3 + 4 — Placeholder + Seed

**Files:**
- Create: `migrations/20260224000003_site_schema_placeholder.go`
- Create: `migrations/20260224000004_seed_rbac_builtins.go`

**Step 1: 创建占位符文件**

```go
package migrations

import (
	"context"
	"log/slog"

	"github.com/uptrace/bun"
)

// Site schemas (site_{slug}) are NOT created by standard migrations.
// They are created dynamically when a new site is registered, via the
// internal/schema package (schema.CreateSiteSchema).
//
// This migration is a no-op placeholder to maintain sequential ordering.

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		slog.Info("site schemas are created dynamically via internal/schema package — skipping")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		slog.Info("site schemas are managed via internal/schema package — skipping")
		return nil
	})
}
```

**Step 2: 创建 seed 文件**

```go
package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO public.sfc_roles (name, slug, description, built_in, status) VALUES
			('超级管理员', 'super', '拥有所有权限，不可修改/删除', true, true),
			('管理员', 'admin', '站点管理，不可删除', true, true),
			('编辑', 'editor', '内容创建与编辑，不可删除', true, true),
			('查看者', 'viewer', '只读访问，不可删除', true, true)
			ON CONFLICT (slug) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("seed built-in roles: %w", err)
		}

		_, err = db.ExecContext(ctx, `
			INSERT INTO public.sfc_role_templates (name, description, built_in) VALUES
			('超级管理员模板', '预置超级管理员权限集', true),
			('管理员模板', '预置管理员权限集', true),
			('编辑模板', '预置编辑权限集', true),
			('查看者模板', '预置查看者权限集', true)
			ON CONFLICT (name) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("seed built-in templates: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
			DELETE FROM public.sfc_role_templates WHERE built_in = true;
			DELETE FROM public.sfc_roles WHERE built_in = true;
		`)
		if err != nil {
			return fmt.Errorf("rollback built-in seeds: %w", err)
		}
		return nil
	})
}
```

**Step 3: 验证编译 + 提交所有 migration**

```bash
go build ./...
```

Expected: 无错误

```bash
git add migrations/
git commit -m "refactor(migrations): rewrite all migrations — remove ENUMs and triggers

- Split into 4 files: core tables / RBAC / placeholder / seed
- Remove all PostgreSQL ENUM types (model uses SMALLINT+CHECK)
- Remove update_updated_at() function and all BEFORE UPDATE triggers
- Add missing sfc_password_reset_tokens table with 3 indexes
- Add deleted_at column to sfc_sites for soft delete support"
```

---

### Task 5: 修复 template.go — ENUM → SMALLINT + 移除触发器 + 移除 pg_trgm

**Files:**
- Modify: `internal/schema/template.go`

这是最复杂的修改，共 16 处变更。按顺序执行：

**Step 1: 替换 sfc_site_posts.status**

```
旧: status          post_status NOT NULL DEFAULT 'draft',
新: status          SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4),
```

**Step 2: 替换 sfc_site_posts 索引条件**

```
旧: WHERE status = 'published' AND deleted_at IS NULL;
新: WHERE status = 3 AND deleted_at IS NULL;

旧: WHERE status = 'scheduled';
新: WHERE status = 2;
```

**Step 3: 删除 sfc_site_post_types 的触发器（2 行）**

```sql
-- 删除:
CREATE TRIGGER trg_sfc_site_post_types_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_post_types FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Step 4: 删除 sfc_site_posts 的触发器（2 行）**

```sql
-- 删除:
CREATE TRIGGER trg_sfc_site_posts_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_posts FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Step 5: 删除 sfc_site_post_translations 的触发器（2 行）**

```sql
-- 删除:
CREATE TRIGGER trg_sfc_site_post_translations_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_post_translations FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Step 6: 删除 sfc_site_categories 的触发器（2 行）**

```sql
-- 删除:
CREATE TRIGGER trg_sfc_site_categories_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_categories FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Step 7: 删除 sfc_site_tags 的 pg_trgm 索引（1 行）**

```sql
-- 删除:
CREATE INDEX idx_sfc_site_tags_name_trgm ON {schema}.sfc_site_tags USING gin(name gin_trgm_ops);
```

**Step 8: 替换 sfc_site_media_files.media_type**

```
旧: media_type      media_type NOT NULL DEFAULT 'other',
新: media_type      SMALLINT NOT NULL DEFAULT 5 CHECK (media_type BETWEEN 1 AND 5),
```

**Step 9: 删除 sfc_site_media_files 的 pg_trgm 索引（1 行）**

```sql
-- 删除:
CREATE INDEX idx_sfc_site_media_name_trgm ON {schema}.sfc_site_media_files USING gin(file_name gin_trgm_ops);
```

**Step 10: 删除 sfc_site_media_files 的触发器（2 行）**

```sql
-- 删除:
CREATE TRIGGER trg_sfc_site_media_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_media_files FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Step 11: 替换 sfc_site_comments.status**

```
旧: status        comment_status NOT NULL DEFAULT 'pending',
新: status        SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4),
```

**Step 12: 替换 sfc_site_comments 索引条件**

```
旧: WHERE status = 'pending' AND deleted_at IS NULL;
新: WHERE status = 1 AND deleted_at IS NULL;
```

**Step 13: 删除 sfc_site_comments 的触发器（2 行）**

```sql
-- 删除:
CREATE TRIGGER trg_sfc_site_comments_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_comments FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Step 14: 删除 sfc_site_menus 的触发器（2 行）**

```sql
-- 删除:
CREATE TRIGGER trg_sfc_site_menus_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_menus FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Step 15: 替换 sfc_site_menu_items.type**

```
旧: type          menu_item_type NOT NULL DEFAULT 'custom',
新: type          SMALLINT NOT NULL DEFAULT 1 CHECK (type BETWEEN 1 AND 5),
```

**Step 16: 删除 sfc_site_menu_items 的触发器（2 行）**

```sql
-- 删除:
CREATE TRIGGER trg_sfc_site_menu_items_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_menu_items FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Step 17: 删除 sfc_site_redirects 的触发器（2 行）**

```sql
-- 删除:
CREATE TRIGGER trg_sfc_site_redirects_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_redirects FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Step 18: 替换 sfc_site_audits.action**

```
旧: action            log_action NOT NULL,
新: action            SMALLINT NOT NULL CHECK (action BETWEEN 1 AND 11),
```

**Step 19: 编译验证**

```bash
go build ./...
```

Expected: 无错误

**Step 20: 提交**

```bash
git add internal/schema/template.go
git commit -m "refactor(schema): replace ENUMs with SMALLINT+CHECK, remove triggers and pg_trgm

- Replace 5 ENUM type references with SMALLINT + CHECK constraints
- Update index WHERE conditions from ENUM strings to integer values
- Remove all 9 BEFORE UPDATE triggers (updated_at managed by Go app layer)
- Remove 2 pg_trgm GIN indexes (search handled by Meilisearch)"
```

---

### Task 6: 更新 database.md

**Files:**
- Modify: `docs/database.md`

**Step 1: 给 sfc_sites ER 图添加 deleted_at**

在 `docs/database.md` 的 sfc_sites ER 定义中（约第 29-42 行），在 `timestamptz updated_at` 后添加:

```
        timestamptz deleted_at
```

**Step 2: 给 sfc_sites DDL 添加 deleted_at**

在 `docs/database.md` 的 sfc_sites CREATE TABLE 中（约第 490 行 `updated_at` 后），添加:

```sql
    deleted_at      TIMESTAMPTZ,
```

**Step 3: 更新迁移文件列表**

在 `docs/database.md` 的迁移文件结构部分（约第 1370-1377 行），替换为:

```
migrations/
├── 20260224000001_create_core_tables.go        -- sfc_users, sfc_sites, sfc_refresh_tokens, sfc_password_reset_tokens, sfc_user_totp, sfc_configs
├── 20260224000002_create_rbac_tables.go        -- sfc_roles + sfc_user_roles + sfc_apis + sfc_role_apis + sfc_menus + sfc_role_menus + sfc_role_templates + sfc_role_template_apis + sfc_role_template_menus
├── 20260224000003_site_schema_placeholder.go   -- 占位符（站点 schema 由 internal/schema 动态创建）
├── 20260224000004_seed_rbac_builtins.go        -- Seed 4 内置角色 + 4 内置权限模板
└── ...
```

**Step 4: 提交**

```bash
git add docs/database.md
git commit -m "docs(database): add deleted_at to sfc_sites, update migration listing"
```

---

### Task 7: 修复 RoleTemplate alias 冲突

**Files:**
- Modify: `internal/model/role_template.go:9`

**Step 1: 修改 alias**

```
旧: bun.BaseModel `bun:"table:sfc_role_templates,alias:rt"`
新: bun.BaseModel `bun:"table:sfc_role_templates,alias:rtpl"`
```

**Step 2: 编译验证**

```bash
go build ./...
```

Expected: 无错误

**Step 3: 提交**

```bash
git add internal/model/role_template.go
git commit -m "fix(model): rename RoleTemplate alias from rt to rtpl to avoid RefreshToken conflict"
```

---

### Task 8: 最终验证

**Step 1: 完整编译**

```bash
go build ./...
```

Expected: 无错误

**Step 2: 检查 commit 历史**

```bash
git log --oneline -5
```

Expected: 4 个新 commit（migration 重写 / template 修复 / database.md / alias 修复）

**Step 3: 验证无遗漏的 ENUM 引用**

```bash
grep -rn "post_status\|comment_status\|menu_item_type\|log_action" migrations/ internal/schema/template.go
```

Expected: 无匹配

注意: `media_type` 会在 template.go 中作为列名出现，这是正常的（列名不是 ENUM 引用）。只要不是 `media_type media_type NOT NULL` 的类型引用即可。

**Step 4: 验证无遗漏的触发器引用**

```bash
grep -rn "CREATE TRIGGER\|update_updated_at" migrations/ internal/schema/template.go
```

Expected: 无匹配

**Step 5: 验证无遗漏的 pg_trgm 引用**

```bash
grep -rn "trgm" migrations/ internal/schema/template.go
```

Expected: 无匹配

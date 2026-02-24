# CMS 内容管理系统 — 数据库设计

**数据库**：PostgreSQL 18
**命名规范**：snake_case，时间字段统一 `timestamptz`，软删除使用 `deleted_at`
**多站点架构**：Schema Isolation — `public` schema 存放全局表（用户、站点注册、认证），`site_{slug}` schema 存放各站点内容表。每个请求通过 `SET search_path TO 'site_{slug}', 'public'` 自动路由到正确的站点 schema，无需 `site_id` 列。

> 所有主键 `id` 使用 `uuidv7()`（PostgreSQL 18 原生支持），生成基于时间戳的 UUIDv7，兼具全局唯一性和时间有序性。相比 UUIDv4（`gen_random_uuid()`），UUIDv7 的 B-tree 索引写入始终追加到末尾，避免随机页分裂，显著提升写入性能和索引效率。无需任何额外扩展。

---

## 1. 数据库整体 ER 关系

```mermaid
erDiagram
    %% ============ PUBLIC SCHEMA（全局表） ============
    sfc_users {
        uuid    id            PK "uuidv7()"
        varchar email         UK
        varchar password_hash
        varchar display_name
        text    avatar_url
        boolean is_active
        timestamptz last_login_at
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    sfc_sites {
        uuid    id              PK "uuidv7()"
        varchar name
        varchar slug            UK
        varchar domain          UK "nullable"
        text    description
        text    logo_url
        varchar default_locale
        varchar timezone
        boolean is_active
        jsonb   settings
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    sfc_roles {
        uuid    id          PK "uuidv7()"
        varchar name        UK
        varchar slug        UK
        text    description
        boolean built_in
        boolean status
        timestamptz created_at
        timestamptz updated_at
    }

    sfc_user_roles {
        uuid    user_id     FK
        uuid    role_id     FK
        timestamptz created_at
    }

    sfc_apis {
        uuid    id          PK "uuidv7()"
        varchar method
        varchar path
        varchar name
        text    description
        varchar group
        boolean status
        timestamptz created_at
        timestamptz updated_at
    }

    sfc_role_apis {
        uuid    role_id     FK
        uuid    api_id      FK
    }

    sfc_menus {
        uuid    id          PK "uuidv7() 后台管理菜单"
        uuid    parent_id   FK "self-ref"
        varchar name
        varchar icon
        varchar path
        int     sort_order
        boolean status
        timestamptz created_at
        timestamptz updated_at
    }

    sfc_role_menus {
        uuid    role_id     FK
        uuid    menu_id     FK
    }

    sfc_role_templates {
        uuid    id          PK "uuidv7()"
        varchar name        UK
        text    description
        boolean built_in
        timestamptz created_at
        timestamptz updated_at
    }

    sfc_role_template_apis {
        uuid    template_id FK
        uuid    api_id      FK
    }

    sfc_role_template_menus {
        uuid    template_id FK
        uuid    menu_id     FK
    }

    sfc_refresh_tokens {
        uuid    id          PK "uuidv7()"
        uuid    user_id     FK
        varchar token_hash  UK
        timestamptz expires_at
        boolean revoked
        inet    ip_address
        text    user_agent
        timestamptz created_at
    }

    sfc_user_totp {
        uuid    id               PK "uuidv7()"
        uuid    user_id          FK UK
        text    secret_encrypted
        text[]  backup_codes_hash
        boolean is_enabled
        timestamptz verified_at
        timestamptz created_at
        timestamptz updated_at
    }

    sfc_password_reset_tokens {
        uuid    id          PK "uuidv7()"
        uuid    user_id     FK
        varchar token_hash  UK
        timestamptz expires_at
        timestamptz used_at
        timestamptz created_at
    }

    sfc_configs {
        varchar key   PK
        jsonb   value
        uuid    updated_by FK
        timestamptz updated_at
    }

    %% ============ SITE SCHEMA（站点内容表，每站点一套） ============
    sfc_site_posts {
        uuid    id             PK "uuidv7()"
        uuid    author_id      FK "public.sfc_users"
        uuid    cover_image_id FK "sfc_site_media_files"
        varchar post_type
        smallint status
        varchar title
        varchar slug           UK
        text    excerpt
        text    content
        jsonb   content_json
        varchar meta_title
        varchar meta_description
        text    og_image_url
        jsonb   extra_fields
        bigint  view_count
        int     version

        timestamptz published_at
        timestamptz scheduled_at
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    sfc_site_post_translations {
        uuid    id      PK "uuidv7()"
        uuid    post_id FK
        varchar locale
        varchar title
        text    excerpt
        text    content
        jsonb   content_json
        varchar meta_title
        varchar meta_description
        varchar og_image_url
        timestamptz created_at
        timestamptz updated_at
    }

    sfc_site_post_revisions {
        uuid    id          PK "uuidv7()"
        uuid    post_id     FK
        uuid    editor_id   FK "public.sfc_users"
        int     version
        varchar title
        text    content
        jsonb   content_json
        text    diff_summary
        timestamptz created_at
    }

    sfc_site_categories {
        uuid    id        PK "uuidv7()"
        uuid    parent_id FK "self-ref"
        varchar name
        varchar slug
        text    path
        text    description
        int     sort_order
        jsonb   meta
        timestamptz created_at
        timestamptz updated_at
    }

    sfc_site_tags {
        uuid    id         PK "uuidv7()"
        varchar name       UK
        varchar slug       UK
        timestamptz created_at
    }

    sfc_site_media_files {
        uuid    id              PK "uuidv7()"
        uuid    uploader_id     FK "public.sfc_users"
        varchar file_name
        varchar original_name
        varchar mime_type
        smallint media_type
        bigint  file_size
        int     width
        int     height
        text    storage_path
        text    public_url
        text    webp_url
        jsonb   thumbnail_urls
        int     reference_count
        text    alt_text
        jsonb   metadata
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    sfc_site_comments {
        uuid    id           PK "uuidv7()"
        uuid    post_id      FK
        uuid    parent_id    FK "self-ref, max 3 levels"
        uuid    user_id      FK "public.sfc_users, nullable"
        varchar author_name
        varchar author_email
        varchar author_url
        inet    author_ip
        text    user_agent
        text    content
        smallint status
        boolean is_pinned
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    sfc_site_menus {
        uuid    id        PK "uuidv7()"
        varchar name
        varchar slug      UK
        varchar location
        timestamptz created_at
        timestamptz updated_at
    }

    sfc_site_menu_items {
        uuid    id           PK "uuidv7()"
        uuid    menu_id      FK
        uuid    parent_id    FK "self-ref"
        varchar label
        text    url
        varchar target
        smallint type
        uuid    reference_id
        int     sort_order
        boolean is_active
        timestamptz created_at
        timestamptz updated_at
    }

    sfc_site_redirects {
        uuid    id           PK "uuidv7()"
        varchar source_path  UK
        text    target_url
        int     status_code
        boolean is_active
        bigint  hit_count
        timestamptz last_hit_at
        uuid    created_by   FK "public.sfc_users"
        timestamptz created_at
        timestamptz updated_at
    }

    sfc_site_preview_tokens {
        uuid    id           PK "uuidv7()"
        uuid    post_id      FK
        varchar token_hash   UK
        timestamptz expires_at
        uuid    created_by   FK "public.sfc_users"
        timestamptz created_at
    }

    sfc_site_api_keys {
        uuid    id          PK "uuidv7()"
        uuid    owner_id    FK "public.sfc_users"
        varchar name
        varchar key_hash    UK
        varchar key_prefix
        boolean is_active
        timestamptz last_used_at
        timestamptz expires_at
        int     rate_limit
        timestamptz created_at
        timestamptz revoked_at
    }

    sfc_site_audits {
        uuid    id            PK "uuidv7()"
        uuid    actor_id      FK "public.sfc_users"
        varchar actor_email
        smallint action
        varchar resource_type
        text    resource_id
        jsonb   resource_snapshot
        inet    ip_address
        text    user_agent
        timestamptz created_at
    }

    sfc_site_post_category_map {
        uuid    post_id     FK
        uuid    category_id FK
        boolean is_primary
    }

    sfc_site_post_tag_map {
        uuid post_id FK
        uuid tag_id  FK
    }

    sfc_site_post_types {
        uuid    id          PK "uuidv7()"
        varchar name        UK
        varchar slug        UK
        text    description
        jsonb   fields
        boolean built_in
        timestamptz created_at
        timestamptz updated_at
    }

    sfc_site_configs {
        varchar key   PK
        jsonb   value
        uuid    updated_by FK
        timestamptz updated_at
    }

    %% ============ RELATIONSHIPS ============
    %% --- RBAC 关系 ---
    sfc_users                  ||--o{ sfc_user_roles             : "assigned to"
    sfc_roles                  ||--o{ sfc_user_roles             : "has users"
    sfc_roles                  ||--o{ sfc_role_apis              : "has api perms"
    sfc_apis                   ||--o{ sfc_role_apis              : "granted to roles"
    sfc_roles                  ||--o{ sfc_role_menus             : "has menu perms"
    sfc_menus            ||--o{ sfc_role_menus             : "visible to roles"
    sfc_menus            ||--o{ sfc_menus            : "parent_child"
    sfc_role_templates         ||--o{ sfc_role_template_apis     : "has api set"
    sfc_apis                   ||--o{ sfc_role_template_apis     : "in template"
    sfc_role_templates         ||--o{ sfc_role_template_menus    : "has menu set"
    sfc_menus            ||--o{ sfc_role_template_menus    : "in template"
    %% --- 认证关系 ---
    sfc_users                  ||--o{ sfc_refresh_tokens         : "has"
    sfc_users                  ||--o{ sfc_password_reset_tokens  : "resets"
    sfc_users                  ||--o| sfc_user_totp              : "has 2FA"
    sfc_users                  ||--o{ sfc_site_posts             : "authors"
    sfc_users                  ||--o{ sfc_site_post_revisions    : "edits"
    sfc_users                  ||--o{ sfc_site_media_files       : "uploads"
    sfc_users                  ||--o{ sfc_site_api_keys          : "owns"
    sfc_users                  ||--o{ sfc_site_audits            : "performs"
    sfc_users                  ||--o{ sfc_site_comments          : "writes"
    sfc_users                  ||--o| sfc_configs                : "updates"
    sfc_site_posts             ||--o{ sfc_site_post_translations : "has"
    sfc_site_posts             ||--o{ sfc_site_post_revisions    : "versions"
    sfc_site_posts             ||--o{ sfc_site_comments          : "has"
    sfc_site_posts             ||--o{ sfc_site_preview_tokens    : "has"
    sfc_site_posts             }o--o{ sfc_site_categories        : "sfc_site_post_category_map"
    sfc_site_posts             }o--o{ sfc_site_tags              : "sfc_site_post_tag_map"
    sfc_site_posts             }o--o| sfc_site_media_files       : "cover_image"
    sfc_site_categories        ||--o{ sfc_site_categories        : "parent_child"
    sfc_site_comments          ||--o{ sfc_site_comments          : "replies"
    sfc_site_menus             ||--o{ sfc_site_menu_items        : "contains"
    sfc_site_menu_items        ||--o{ sfc_site_menu_items        : "parent_child"
```

---

## 2. 完整 DDL

### 2A. 全局 Schema（public）DDL

```sql
-- ============================================
-- 扩展（安装在 public schema）
-- ============================================
-- PostgreSQL 18 原生内置 uuidv7()，无需额外扩展
-- 全文搜索由 Meilisearch 独立服务承担，不在 PostgreSQL 层实现

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

-- ============================================
-- 公共触发器函数（所有 schema 共享）
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 1. 用户表（全局，所有站点共享）
-- ============================================
-- 注意：角色通过 sfc_user_roles + sfc_roles 动态 RBAC 系统分配
CREATE TABLE public.sfc_users (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,          -- bcrypt, cost=12
    display_name  VARCHAR(100) NOT NULL,
    avatar_url    TEXT,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_sfc_users_email ON public.sfc_users(email) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_sfc_users_updated_at
    BEFORE UPDATE ON public.sfc_users FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 用户软删除联动策略说明
-- ============================================
-- 当用户被软删除（deleted_at 置为时间戳）时，应用层需处理以下关联数据：
--
-- 1. sfc_refresh_tokens：立即吊销该用户所有 Refresh Token（UPDATE SET revoked = true）
-- 2. sfc_site_api_keys（跨所有站点 schema）：立即停用该用户拥有的所有 API Key（UPDATE SET is_active = false, revoked_at = NOW()）
-- 3. sfc_site_posts（跨所有站点 schema）：保留文章不变（author_id 仍指向该用户），文章可由 Admin 重新分配作者
-- 4. sfc_site_media_files（跨所有站点 schema）：保留媒体文件不变（uploader_id 仍指向该用户），文件仍可被引用
-- 5. sfc_site_audits：保留审计日志不变（actor_email 冗余字段确保日志可读）
--
-- 注意：以上联动操作在 Service 层（user_service.go）的 SoftDelete 方法中实现，
-- 不使用数据库级联删除（ON DELETE CASCADE），因为软删除不触发 CASCADE。

-- ============================================
-- 2. 站点注册表（全局，每个站点对应一个 site_{slug} schema）
-- ============================================
CREATE TABLE public.sfc_sites (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    name            VARCHAR(200) NOT NULL,
    slug            VARCHAR(50)  NOT NULL UNIQUE,
    domain          VARCHAR(255) UNIQUE,              -- 可空；自定义域名映射
    description     TEXT,
    logo_url        TEXT,
    default_locale  VARCHAR(10)  NOT NULL DEFAULT 'zh-CN',
    timezone        VARCHAR(50)  NOT NULL DEFAULT 'Asia/Shanghai',
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    settings        JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    -- slug 必须匹配 schema 命名规则
    CONSTRAINT chk_sfc_sites_slug CHECK (slug ~ '^[a-z0-9_]{3,50}$')
);

CREATE INDEX idx_sfc_sites_domain ON public.sfc_sites(domain) WHERE domain IS NOT NULL;
CREATE INDEX idx_sfc_sites_active ON public.sfc_sites(is_active) WHERE is_active = TRUE;

CREATE TRIGGER trg_sfc_sites_updated_at
    BEFORE UPDATE ON public.sfc_sites FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 3. 角色定义表（全局，动态 RBAC）
-- ============================================
-- 系统内置 4 个角色（super/admin/editor/viewer），built_in=true 不可删除
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

CREATE TRIGGER trg_sfc_roles_updated_at
    BEFORE UPDATE ON public.sfc_roles FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 4. 用户-角色分配表（全局，多对多）
-- ============================================
-- 用户可分配多个角色，角色为全局性质（非 per-site）
CREATE TABLE public.sfc_user_roles (
    user_id    UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES public.sfc_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_sfc_user_roles_role ON public.sfc_user_roles(role_id);

-- ============================================
-- 5. API 端点注册表（全局，由启动时自动注册）
-- ============================================
-- 存储所有受 RBAC 保护的 API 端点，启动时通过 ApiRegistry 自动同步
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

CREATE TRIGGER trg_sfc_apis_updated_at
    BEFORE UPDATE ON public.sfc_apis FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 6. 角色-API 权限映射表
-- ============================================
-- 控制角色可以访问的 API 端点
CREATE TABLE public.sfc_role_apis (
    role_id UUID NOT NULL REFERENCES public.sfc_roles(id) ON DELETE CASCADE,
    api_id  UUID NOT NULL REFERENCES public.sfc_apis(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, api_id)
);

CREATE INDEX idx_sfc_role_apis_api ON public.sfc_role_apis(api_id);

-- ============================================
-- 7. 后台管理菜单表（全局，控制管理后台菜单可见性）
-- ============================================
-- 注意：此 sfc_menus 为管理后台菜单（RBAC），非站点 schema 的 sfc_site_menus（前台导航菜单）
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

CREATE TRIGGER trg_sfc_menus_updated_at
    BEFORE UPDATE ON public.sfc_menus FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 8. 角色-菜单可见性映射表
-- ============================================
-- 控制角色在管理后台可以看到的菜单项
CREATE TABLE public.sfc_role_menus (
    role_id UUID NOT NULL REFERENCES public.sfc_roles(id) ON DELETE CASCADE,
    menu_id UUID NOT NULL REFERENCES public.sfc_menus(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, menu_id)
);

CREATE INDEX idx_sfc_role_menus_menu ON public.sfc_role_menus(menu_id);

-- ============================================
-- 9. 权限模板定义表（全局，快速创建角色的权限预设）
-- ============================================
-- 内置 4 个模板（超级管理员/管理员/编辑/查看者模板），创建自定义角色时可选择模板快速初始化权限
CREATE TABLE public.sfc_role_templates (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    built_in    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_sfc_role_templates_updated_at
    BEFORE UPDATE ON public.sfc_role_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 10. 模板-API 映射表
-- ============================================
CREATE TABLE public.sfc_role_template_apis (
    template_id UUID NOT NULL REFERENCES public.sfc_role_templates(id) ON DELETE CASCADE,
    api_id      UUID NOT NULL REFERENCES public.sfc_apis(id) ON DELETE CASCADE,
    PRIMARY KEY (template_id, api_id)
);

-- ============================================
-- 11. 模板-菜单映射表
-- ============================================
CREATE TABLE public.sfc_role_template_menus (
    template_id UUID NOT NULL REFERENCES public.sfc_role_templates(id) ON DELETE CASCADE,
    menu_id     UUID NOT NULL REFERENCES public.sfc_menus(id) ON DELETE CASCADE,
    PRIMARY KEY (template_id, menu_id)
);

-- ============================================
-- 12. 刷新令牌表（全局，持久化 Refresh Token）
-- ============================================
CREATE TABLE public.sfc_refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id     UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,     -- SHA-256 hash
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN NOT NULL DEFAULT FALSE,
    ip_address  INET,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_rt_user_id ON public.sfc_refresh_tokens(user_id);
CREATE INDEX idx_sfc_rt_token   ON public.sfc_refresh_tokens(token_hash);

-- ============================================
-- 13. 用户 TOTP 双因素认证表（全局，用户级别）
-- ============================================
-- 2FA 是用户级别的安全功能，启用后保护该用户在所有站点的登录
CREATE TABLE public.sfc_user_totp (
    id                UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id           UUID NOT NULL UNIQUE REFERENCES public.sfc_users(id) ON DELETE CASCADE,
    secret_encrypted  TEXT NOT NULL,              -- AES-256-GCM 加密的 TOTP 密钥
    backup_codes_hash TEXT[],                     -- bcrypt 哈希的备用码
    is_enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at       TIMESTAMPTZ,                -- 首次验证/激活 2FA 的时间
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_sfc_user_totp_updated_at
    BEFORE UPDATE ON public.sfc_user_totp FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 14. 密码重置令牌表（全局）
-- ============================================
-- 用于"忘记密码"流程，存储 SHA-256 哈希令牌，有效期 30 分钟，单次使用
CREATE TABLE public.sfc_password_reset_tokens (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id     UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,     -- SHA-256(raw_token)
    expires_at  TIMESTAMPTZ NOT NULL,             -- NOW() + 30min
    used_at     TIMESTAMPTZ,                      -- NULL 表示未使用，非 NULL 表示已使用
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_prt_user    ON public.sfc_password_reset_tokens(user_id);
CREATE INDEX idx_sfc_prt_token   ON public.sfc_password_reset_tokens(token_hash);
CREATE INDEX idx_sfc_prt_expires ON public.sfc_password_reset_tokens(expires_at)
    WHERE used_at IS NULL;

-- ============================================
-- 15. 系统配置表（全局级别）
-- ============================================
-- 全局系统配置（如安装标志），与各站点 schema 中的 sfc_site_configs 表分开
CREATE TABLE public.sfc_configs (
    key         VARCHAR(100) PRIMARY KEY,
    value       JSONB NOT NULL,
    description TEXT,
    updated_by  UUID REFERENCES public.sfc_users(id),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 初始配置：安装标志（由安装向导设置为 true）
INSERT INTO public.sfc_configs (key, value, description) VALUES
('system.installed', 'false', '系统是否已通过安装向导初始化')
ON CONFLICT (key) DO NOTHING;
```

---

### 2B. 站点 Schema 模板（site_{slug}）DDL

> 以下为每个站点 schema 的完整 DDL 模板。创建新站点时，系统执行 `CREATE SCHEMA site_{slug}` 后在该 schema 内创建以下所有表。DDL 中使用 `{schema}` 作为实际 schema 名称的占位符（如 `site_blog`）。
>
> **Schema 隔离**：所有内容表无需 `site_id` 列，schema 本身提供站点隔离。引用用户表的外键一律使用 `REFERENCES public.sfc_users(id)` 跨 schema 引用。

```sql
-- ============================================
-- 创建站点 Schema
-- ============================================
CREATE SCHEMA IF NOT EXISTS {schema};

-- 后续 DDL 在 SET search_path TO '{schema}', 'public' 环境下执行，
-- 使 REFERENCES public.sfc_users(id) 等跨 schema 引用正确解析

-- ============================================
-- 1. 内容类型表
-- ============================================
-- V1.0 仅使用 sfc_site_posts.post_type VARCHAR 字段支持基础类型（article/page），V1.1 引入 sfc_site_post_types 表实现完整的自定义内容类型管理
CREATE TABLE {schema}.sfc_site_post_types (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(100) NOT NULL UNIQUE,          -- 显示名称
    slug        VARCHAR(100) NOT NULL UNIQUE,          -- URL 安全标识符
    description TEXT,
    fields      JSONB NOT NULL DEFAULT '[]',           -- JSON Schema 定义自定义字段
    built_in    BOOLEAN NOT NULL DEFAULT FALSE,        -- 是否为系统内置类型（article/page）
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_site_post_types_slug ON {schema}.sfc_site_post_types(slug);

CREATE TRIGGER trg_sfc_site_post_types_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_post_types FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 2. 文章主表
-- ============================================
-- 无 site_id 列 — schema 隔离提供站点范围
CREATE TABLE {schema}.sfc_site_posts (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    author_id       UUID NOT NULL REFERENCES public.sfc_users(id),
    cover_image_id  UUID,                                      -- FK 在 sfc_site_media_files 创建后添加
    post_type       VARCHAR(50) NOT NULL DEFAULT 'article',    -- article/page/product
    status          SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4),

    -- 内容字段（默认语言，zh-CN）
    title           VARCHAR(500) NOT NULL,
    slug            VARCHAR(600) NOT NULL,
    excerpt         TEXT,
    content         TEXT,                           -- HTML
    content_json    JSONB,                          -- BlockNote / TipTap JSON

    -- SEO
    meta_title       VARCHAR(200),
    meta_description VARCHAR(500),
    og_image_url     TEXT,

    -- 自定义字段
    extra_fields    JSONB DEFAULT '{}',

    -- 统计
    view_count      BIGINT NOT NULL DEFAULT 0,
    version         INT    NOT NULL DEFAULT 1,     -- 乐观锁版本号，每次更新 +1

    -- 时间控制
    published_at    TIMESTAMPTZ,
    scheduled_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- sfc_site_posts.slug 在站点 schema 内唯一
CREATE UNIQUE INDEX idx_sfc_site_posts_slug      ON {schema}.sfc_site_posts(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_site_posts_author           ON {schema}.sfc_site_posts(author_id);
CREATE INDEX idx_sfc_site_posts_status           ON {schema}.sfc_site_posts(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_site_posts_published        ON {schema}.sfc_site_posts(published_at DESC)
    WHERE status = 'published' AND deleted_at IS NULL;

CREATE INDEX idx_sfc_site_posts_extra            ON {schema}.sfc_site_posts USING gin(extra_fields);
CREATE INDEX idx_sfc_site_posts_scheduled        ON {schema}.sfc_site_posts(scheduled_at) WHERE status = 'scheduled';

CREATE TRIGGER trg_sfc_site_posts_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_posts FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- view_count 使用 UPDATE sfc_site_posts SET view_count = view_count + 1 保证原子性，高并发场景通过 Redis INCR 缓冲后批量写入

-- **乐观锁**：`sfc_site_posts.version` 为权威版本号，用于并发编辑冲突检测。`sfc_site_post_revisions.version` 为历史修订版本号，两者独立维护。更新文章时 `sfc_site_posts.version` 自增，同时创建新的 revision 记录。

-- ============================================
-- 3. 文章多语言表
-- ============================================
CREATE TABLE {schema}.sfc_site_post_translations (
    id               UUID PRIMARY KEY DEFAULT uuidv7(),
    post_id          UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    locale           VARCHAR(10) NOT NULL,              -- zh-CN / en / ja
    title            VARCHAR(500),
    excerpt          TEXT,
    content          TEXT,
    content_json     JSONB,
    meta_title       VARCHAR(200),
    meta_description VARCHAR(500),
    og_image_url     VARCHAR(500),                     -- 翻译可覆盖主文章的 OG 图片，为 NULL 时继承主文章
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(post_id, locale)
);

CREATE INDEX idx_sfc_site_pt_post_locale ON {schema}.sfc_site_post_translations(post_id, locale);

CREATE TRIGGER trg_sfc_site_post_translations_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_post_translations FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 4. 文章修订历史
-- ============================================
CREATE TABLE {schema}.sfc_site_post_revisions (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    post_id      UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    editor_id    UUID NOT NULL REFERENCES public.sfc_users(id),
    version      INT NOT NULL,
    title        VARCHAR(500),
    content      TEXT,
    content_json JSONB,
    diff_summary TEXT,                              -- 本次变更摘要
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_site_revisions_post   ON {schema}.sfc_site_post_revisions(post_id, version DESC);
CREATE INDEX idx_sfc_site_revisions_editor ON {schema}.sfc_site_post_revisions(editor_id);

-- ============================================
-- 5. 分类表（Materialized Path）
-- ============================================
CREATE TABLE {schema}.sfc_site_categories (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    parent_id   UUID REFERENCES {schema}.sfc_site_categories(id) ON DELETE RESTRICT,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(200) NOT NULL,
    path        TEXT NOT NULL DEFAULT '/',         -- e.g. /tech/backend/
    description TEXT,
    sort_order  INT NOT NULL DEFAULT 0,
    meta        JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(parent_id, slug)                        -- slug 同级唯一（同一 parent_id 下）
);

CREATE INDEX idx_sfc_site_categories_parent ON {schema}.sfc_site_categories(parent_id);
CREATE INDEX idx_sfc_site_categories_path   ON {schema}.sfc_site_categories(path);
CREATE INDEX idx_sfc_site_categories_slug   ON {schema}.sfc_site_categories(parent_id, slug);

CREATE TRIGGER trg_sfc_site_categories_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_categories FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 6. 标签表
-- ============================================
CREATE TABLE {schema}.sfc_site_tags (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    slug        VARCHAR(200) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_site_tags_name_trgm ON {schema}.sfc_site_tags USING gin(name gin_trgm_ops);

-- ============================================
-- 7. 文章-分类 多对多
-- ============================================
CREATE TABLE {schema}.sfc_site_post_category_map (
    post_id     UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES {schema}.sfc_site_categories(id) ON DELETE CASCADE,
    is_primary  BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (post_id, category_id)
);

CREATE INDEX idx_sfc_site_pcm_category ON {schema}.sfc_site_post_category_map(category_id);

-- ============================================
-- 8. 文章-标签 多对多
-- ============================================
CREATE TABLE {schema}.sfc_site_post_tag_map (
    post_id UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES {schema}.sfc_site_tags(id)  ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

CREATE INDEX idx_sfc_site_ptm_tag ON {schema}.sfc_site_post_tag_map(tag_id);

-- ============================================
-- 9. 媒体文件表
-- ============================================
CREATE TABLE {schema}.sfc_site_media_files (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    uploader_id     UUID NOT NULL REFERENCES public.sfc_users(id),
    file_name       VARCHAR(500) NOT NULL,
    original_name   VARCHAR(500) NOT NULL,
    mime_type       VARCHAR(100) NOT NULL,
    media_type      SMALLINT NOT NULL DEFAULT 5 CHECK (media_type BETWEEN 1 AND 5),
    file_size       BIGINT NOT NULL,              -- bytes
    width           INT,                          -- 图片宽度 px
    height          INT,                          -- 图片高度 px
    storage_path    TEXT NOT NULL,                -- 相对存储路径
    public_url      TEXT NOT NULL,                -- 访问 URL
    webp_url        TEXT,                         -- WebP 版本 URL
    thumbnail_urls  JSONB DEFAULT '{}',           -- {"sm": "url", "md": "url"}
    reference_count INT NOT NULL DEFAULT 0,       -- 引用计数（见下方软删除注意事项）
    alt_text        TEXT,
    metadata        JSONB DEFAULT '{}',           -- EXIF / 视频时长等
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- 添加 sfc_site_posts.cover_image_id 外键（两张表都已创建）
ALTER TABLE {schema}.sfc_site_posts
    ADD CONSTRAINT fk_sfc_site_posts_cover_image
    FOREIGN KEY (cover_image_id) REFERENCES {schema}.sfc_site_media_files(id);

CREATE INDEX idx_sfc_site_media_uploader  ON {schema}.sfc_site_media_files(uploader_id);
CREATE INDEX idx_sfc_site_media_type      ON {schema}.sfc_site_media_files(media_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_site_media_name_trgm ON {schema}.sfc_site_media_files USING gin(file_name gin_trgm_ops);

CREATE TRIGGER trg_sfc_site_media_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_media_files FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- sfc_site_media_files.reference_count 与软删除注意事项
-- ============================================
-- 当文章被软删除（sfc_site_posts.deleted_at 置为时间戳）时，关联媒体的 reference_count 不应递减，
-- 因为软删除的文章可以被恢复（restore），恢复后仍需引用原媒体文件。
-- 仅在文章被永久删除（物理删除，即从数据库中真正 DELETE）时，才递减关联媒体的 reference_count。
-- 该逻辑在 Service 层（post_service.go / media_service.go）中实现。

-- ============================================
-- 10. 评论表
-- ============================================
-- 支持游客和登录用户评论。游客评论需提供 author_name 和 author_email（通过 CHECK 约束强制）。
-- 自引用 parent_id 支持嵌套回复（应用层限制最大 3 级深度）。
CREATE TABLE {schema}.sfc_site_comments (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    post_id       UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    parent_id     UUID REFERENCES {schema}.sfc_site_comments(id) ON DELETE CASCADE,
    user_id       UUID REFERENCES public.sfc_users(id) ON DELETE SET NULL,  -- 非空表示登录用户评论
    author_name   VARCHAR(100),
    author_email  VARCHAR(255),
    author_url    VARCHAR(500),
    author_ip     INET,
    user_agent    TEXT,
    content       TEXT NOT NULL,
    status        SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4),
    is_pinned     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,

    -- 内容长度约束
    CONSTRAINT chk_sfc_site_comment_content_length CHECK (length(content) BETWEEN 1 AND 10000),
    -- 游客评论必须提供 author_name 和 author_email
    CONSTRAINT chk_sfc_site_comment_guest CHECK (
        user_id IS NOT NULL OR (author_name IS NOT NULL AND author_email IS NOT NULL)
    )
);

CREATE INDEX idx_sfc_site_comments_post_status ON {schema}.sfc_site_comments(post_id, status, created_at)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_site_comments_parent      ON {schema}.sfc_site_comments(parent_id)
    WHERE parent_id IS NOT NULL;
CREATE INDEX idx_sfc_site_comments_moderation  ON {schema}.sfc_site_comments(status, created_at DESC)
    WHERE status = 'pending' AND deleted_at IS NULL;
CREATE INDEX idx_sfc_site_comments_email       ON {schema}.sfc_site_comments(author_email);
CREATE INDEX idx_sfc_site_comments_user        ON {schema}.sfc_site_comments(user_id)
    WHERE user_id IS NOT NULL AND deleted_at IS NULL;

CREATE TRIGGER trg_sfc_site_comments_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_comments FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- **最大嵌套深度**（3 级）在应用层强制，不在数据库约束中实现，以避免递归 CHECK 复杂性。

-- ============================================
-- 11. 导航菜单表
-- ============================================
CREATE TABLE {schema}.sfc_site_menus (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    location    VARCHAR(50),                       -- e.g. 'header', 'footer', 'sidebar'
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_sfc_site_menus_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_menus FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 12. 菜单项表
-- ============================================
-- 层级菜单项，支持排序
CREATE TABLE {schema}.sfc_site_menu_items (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    menu_id       UUID NOT NULL REFERENCES {schema}.sfc_site_menus(id) ON DELETE CASCADE,
    parent_id     UUID REFERENCES {schema}.sfc_site_menu_items(id) ON DELETE CASCADE,
    label         VARCHAR(200) NOT NULL,
    url           TEXT,
    target        VARCHAR(10) NOT NULL DEFAULT '_self',
    type          SMALLINT NOT NULL DEFAULT 1 CHECK (type BETWEEN 1 AND 5),
    reference_id  UUID,                            -- 指向 post/category/tag 的 FK（按 type 解析）
    sort_order    INT NOT NULL DEFAULT 0,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_site_menu_items_menu      ON {schema}.sfc_site_menu_items(menu_id, sort_order);
CREATE INDEX idx_sfc_site_menu_items_parent    ON {schema}.sfc_site_menu_items(parent_id)
    WHERE parent_id IS NOT NULL;
CREATE INDEX idx_sfc_site_menu_items_reference ON {schema}.sfc_site_menu_items(type, reference_id)
    WHERE reference_id IS NOT NULL;

CREATE TRIGGER trg_sfc_site_menu_items_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_menu_items FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 13. URL 重定向表
-- ============================================
CREATE TABLE {schema}.sfc_site_redirects (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    source_path   VARCHAR(500) NOT NULL UNIQUE,
    target_url    TEXT NOT NULL,
    status_code   INT NOT NULL DEFAULT 301,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    hit_count     BIGINT NOT NULL DEFAULT 0,
    last_hit_at   TIMESTAMPTZ,
    created_by    UUID REFERENCES public.sfc_users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_sfc_site_redirect_status_code CHECK (status_code IN (301, 302))
);

CREATE INDEX idx_sfc_site_redirects_source ON {schema}.sfc_site_redirects(source_path)
    WHERE is_active = TRUE;
CREATE INDEX idx_sfc_site_redirects_created ON {schema}.sfc_site_redirects(created_at DESC);

CREATE TRIGGER trg_sfc_site_redirects_updated_at
    BEFORE UPDATE ON {schema}.sfc_site_redirects FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 14. 草稿预览令牌表
-- ============================================
-- 时间限制的令牌，用于预览未发布文章
CREATE TABLE {schema}.sfc_site_preview_tokens (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    post_id     UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,      -- SHA-256(raw_token)
    expires_at  TIMESTAMPTZ NOT NULL,
    created_by  UUID REFERENCES public.sfc_users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_site_preview_tokens_post    ON {schema}.sfc_site_preview_tokens(post_id);
CREATE INDEX idx_sfc_site_preview_tokens_hash    ON {schema}.sfc_site_preview_tokens(token_hash)
    WHERE expires_at > NOW();
CREATE INDEX idx_sfc_site_preview_tokens_expires ON {schema}.sfc_site_preview_tokens(expires_at);

-- ============================================
-- 15. API Key 表
-- ============================================
CREATE TABLE {schema}.sfc_site_api_keys (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    owner_id     UUID NOT NULL REFERENCES public.sfc_users(id),
    name         VARCHAR(100) NOT NULL,
    key_hash     VARCHAR(255) NOT NULL UNIQUE,     -- SHA-256(raw_key)
    key_prefix   VARCHAR(20) NOT NULL,             -- 展示前缀 cms_live_xxxx
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    rate_limit   INT NOT NULL DEFAULT 100,         -- req/min
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX idx_sfc_site_apikeys_owner ON {schema}.sfc_site_api_keys(owner_id);
CREATE INDEX idx_sfc_site_apikeys_hash  ON {schema}.sfc_site_api_keys(key_hash);

-- ============================================
-- 16. 操作审计日志（按月分区）
-- ============================================
CREATE TABLE {schema}.sfc_site_audits (
    id                UUID NOT NULL DEFAULT uuidv7(),
    actor_id          UUID REFERENCES public.sfc_users(id),
    actor_email       VARCHAR(255),                    -- 冗余，防止用户被删后日志丢失
    action            SMALLINT NOT NULL CHECK (action BETWEEN 1 AND 11),
    resource_type     VARCHAR(50) NOT NULL,             -- post/user/media/category/comment/menu...
    resource_id       TEXT,
    resource_snapshot JSONB,                            -- 操作时的资源快照
    ip_address        INET,
    user_agent        TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- 初始分区（日期按需调整；自动化任务会创建未来分区）
CREATE TABLE {schema}.sfc_site_audits_2026_02 PARTITION OF {schema}.sfc_site_audits
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE {schema}.sfc_site_audits_2026_03 PARTITION OF {schema}.sfc_site_audits
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE {schema}.sfc_site_audits_2026_04 PARTITION OF {schema}.sfc_site_audits
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

-- 分区自动管理（参考 deployment.md 中的定时任务设计）：
-- - CreateFutureAuditPartitions：每月 1 号自动创建未来 2 个月的分区（遍历所有站点 schema）
-- - DropExpiredAuditPartitions：自动清理超过保留期的历史分区（默认保留 12 个月，遍历所有站点 schema）

CREATE INDEX idx_sfc_site_audits_actor    ON {schema}.sfc_site_audits(actor_id, created_at DESC);
CREATE INDEX idx_sfc_site_audits_resource ON {schema}.sfc_site_audits(resource_type, resource_id);
CREATE INDEX idx_sfc_site_audits_time     ON {schema}.sfc_site_audits(created_at DESC);

-- ============================================
-- 17. 系统配置表（站点级别）
-- ============================================
-- 站点级别的配置，与 public.sfc_configs（全局级别）分开
CREATE TABLE {schema}.sfc_site_configs (
    key         VARCHAR(100) PRIMARY KEY,
    value       JSONB NOT NULL,
    description TEXT,
    updated_by  UUID REFERENCES public.sfc_users(id),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 站点默认配置
INSERT INTO {schema}.sfc_site_configs (key, value, description) VALUES
('media.max_size',     '104857600',  '最大文件大小（bytes）'),
('media.storage',      '"rustfs"',   '存储驱动：rustfs（S3 兼容）'),
('content.trash_days', '30',         '回收站保留天数')
ON CONFLICT (key) DO NOTHING;

-- ============================================
-- 18. 触发器汇总（所有站点表的 updated_at 触发器）
-- ============================================
-- 以下触发器已在各表创建语句中定义，此处汇总列出：
-- trg_sfc_site_post_types_updated_at          ON {schema}.sfc_site_post_types
-- trg_sfc_site_posts_updated_at               ON {schema}.sfc_site_posts
-- trg_sfc_site_post_translations_updated_at   ON {schema}.sfc_site_post_translations
-- trg_sfc_site_media_updated_at               ON {schema}.sfc_site_media_files
-- trg_sfc_site_categories_updated_at          ON {schema}.sfc_site_categories
-- trg_sfc_site_comments_updated_at            ON {schema}.sfc_site_comments
-- trg_sfc_site_menus_updated_at               ON {schema}.sfc_site_menus
-- trg_sfc_site_menu_items_updated_at          ON {schema}.sfc_site_menu_items
-- trg_sfc_site_redirects_updated_at           ON {schema}.sfc_site_redirects
--
-- 以下表无 updated_at 触发器：
-- sfc_site_tags（无 updated_at 列）、sfc_site_post_revisions（仅追加）、sfc_site_post_category_map / sfc_site_post_tag_map（关联表）、
-- sfc_site_preview_tokens（仅追加）、sfc_site_api_keys（无 updated_at）、sfc_site_audits（仅追加）

-- ============================================
-- 19. 初始化内置内容类型
-- ============================================
INSERT INTO {schema}.sfc_site_post_types (name, slug, built_in) VALUES
('文章', 'article', TRUE),
('页面', 'page', TRUE)
ON CONFLICT (slug) DO NOTHING;
```

---

### 2C. 设计决策说明

```sql
-- ============================================
-- post_count 设计决策
-- ============================================
-- **设计决策**：post_count 采用实时 COUNT 查询而非反范式字段。
-- 理由：
--   (1) 软删除导致反范式计数不准确；
--   (2) 分类/标签数量有限，JOIN COUNT 性能可接受；
--   (3) 热点数据由 Redis 缓存 60s。
-- 因此从 sfc_site_categories 和 sfc_site_tags 表中移除 post_count 列，改为查询时动态计算。
--
-- 分类文章数查询示例：
--   SELECT c.*, (
--     SELECT COUNT(*) FROM sfc_site_post_category_map pcm
--     JOIN sfc_site_posts p ON p.id = pcm.post_id
--     WHERE pcm.category_id = c.id AND p.deleted_at IS NULL
--   ) AS post_count
--   FROM sfc_site_categories c;
--
-- 标签文章数查询示例：
--   SELECT t.*, (
--     SELECT COUNT(*) FROM sfc_site_post_tag_map ptm
--     JOIN sfc_site_posts p ON p.id = ptm.post_id
--     WHERE ptm.tag_id = t.id AND p.deleted_at IS NULL
--   ) AS post_count
--   FROM sfc_site_tags t;
```

---

## 2.1 初始数据（Seed）

以下为**开发环境**的初始 seed 数据，仅供参考。生产环境通过 Web 安装向导（`/setup`）完成初始化。

> **生产环境请勿使用此 SQL**，应通过安装向导（`/setup`）完成初始化。

```sql
-- ============================================
-- 1. 创建默认站点
-- ============================================
INSERT INTO public.sfc_sites (name, slug, default_locale, timezone) VALUES
('默认站点', 'default_site', 'zh-CN', 'Asia/Shanghai')
ON CONFLICT (slug) DO NOTHING;

-- ============================================
-- 2. Seed 内置角色（由迁移 20260224000004 自动执行）
-- ============================================
INSERT INTO public.sfc_roles (name, slug, description, built_in, status) VALUES
('超级管理员', 'super', '拥有所有权限，不可修改/删除', true, true),
('管理员', 'admin', '站点管理，不可删除', true, true),
('编辑', 'editor', '内容创建与编辑，不可删除', true, true),
('查看者', 'viewer', '只读访问，不可删除', true, true)
ON CONFLICT (slug) DO NOTHING;

-- ============================================
-- 3. Seed 内置权限模板（由迁移 20260224000004 自动执行）
-- ============================================
INSERT INTO public.sfc_role_templates (name, description, built_in) VALUES
('超级管理员模板', '预置超级管理员权限集', true),
('管理员模板', '预置管理员权限集', true),
('编辑模板', '预置编辑权限集', true),
('查看者模板', '预置查看者权限集', true)
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- 4. 创建初始管理员（密码：Admin@123456，bcrypt cost=12）
-- ============================================
-- 仅限开发/测试环境使用，生产环境应通过安装向导或 CLI 命令创建
-- 注意：角色通过 sfc_user_roles + sfc_roles 动态 RBAC 系统分配
INSERT INTO sfc_users (email, password_hash, display_name) VALUES
('admin@example.com', '$2a$12$...placeholder...', 'Admin')
ON CONFLICT (email) DO NOTHING;

-- ============================================
-- 5. 分配超级管理员角色给初始管理员
-- ============================================
INSERT INTO public.sfc_user_roles (user_id, role_id)
SELECT u.id, r.id
FROM public.sfc_users u, public.sfc_roles r
WHERE u.email = 'admin@example.com' AND r.slug = 'super'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- ============================================
-- 6. 设置安装标志
-- ============================================
INSERT INTO public.sfc_configs (key, value, description) VALUES
('system.installed', 'true', '系统是否已通过安装向导初始化')
ON CONFLICT (key) DO UPDATE SET value = 'true';
```

---

## 3. Redis 键空间设计

```
# ============================================
# 全局键（无站点前缀）
# ============================================
auth:blacklist:{jti}                     TTL=token剩余时间   -- 登出黑名单（jti = JWT Token ID）
login_fail:{email}                       TTL=900s            -- 登录失败计数
password_reset:{email}                   TTL=300s            -- 密码重置限流（1 次/5 分钟/邮箱）
system:installed                         TTL=indefinite      -- 安装标志

# 2FA 相关（全局，用户级别，无站点前缀）
2fa:rate:{user_id}                       TTL=300s            -- 2FA 验证码尝试限流（5 次/5 分钟）
2fa:setup_rate:{user_id}                 TTL=3600s           -- 2FA 设置限流（3 次/小时）
2fa:used:{user_id}:{code}               TTL=90s             -- TOTP 重放防护

# ============================================
# 站点解析缓存
# ============================================
site:domain:{domain}                     TTL=600s            -- 域名 → 站点 slug 映射
site:slug:{slug}                         TTL=600s            -- slug → 完整站点 JSON

# ============================================
# RBAC 权限缓存（两级缓存：L1 Redis + L2 Redis）
# ============================================
# L1: 用户角色缓存（角色变更时失效）
user:{user_id}:roles                     TTL=300s            -- 用户角色 slugs + role_ids（JSON）
user:{user_id}:menu_tree                 TTL=300s            -- 用户可见管理后台菜单树（JSON）
# L2: 角色权限缓存（权限变更时失效）
role:{role_id}:api_set                   TTL=600s            -- 角色可访问 API 集合 ["METHOD:/path", ...]

# ============================================
# 站点内容缓存（所有键以 site:{slug}: 为前缀）
# ============================================
site:{slug}:cache:post:list:{hash}       TTL=60s             -- 文章列表（查询参数哈希）
site:{slug}:cache:post:detail:{post_id}  TTL=300s            -- 文章详情
site:{slug}:cache:category:tree          TTL=600s            -- 分类树
site:{slug}:cache:tag:popular            TTL=300s            -- 热门标签
site:{slug}:cache:comments:post:{post_id}:{hash}  TTL=60s   -- 评论列表（公开 API）
site:{slug}:cache:comments:count:{post_id}         TTL=60s   -- 评论数
site:{slug}:cache:menu:loc:{location}    TTL=300s            -- 按位置获取菜单（公开 API）
site:{slug}:cache:menu:slug:{menu_slug}  TTL=300s            -- 按 slug 获取菜单
site:{slug}:cache:menu:detail:{menu_id}  TTL=300s            -- 菜单详情（管理 API）
site:{slug}:cache:feed:rss:{hash}        TTL=3600s           -- RSS Feed
site:{slug}:cache:feed:atom:{hash}       TTL=3600s           -- Atom Feed
site:{slug}:cache:sitemap:index          TTL=3600s           -- Sitemap 索引
site:{slug}:cache:sitemap:posts          TTL=3600s           -- 文章 Sitemap
site:{slug}:cache:sitemap:categories     TTL=3600s           -- 分类 Sitemap
site:{slug}:cache:sitemap:tags           TTL=3600s           -- 标签 Sitemap

# ============================================
# 站点重定向缓存
# ============================================
site:{slug}:redirects:map                TTL=600s            -- 所有活跃重定向的哈希表（source → target）
site:{slug}:redirects:lock               TTL=5s              -- 重建锁（防缓存击穿）
site:{slug}:redirect:hits:{redirect_id}  TTL=indefinite      -- 命中计数缓冲（定时刷入 DB）

# ============================================
# 站点限流
# ============================================
site:{slug}:ratelimit:api:{key_hash}:{minute}     TTL=60s    -- API 滑动窗口计数
site:{slug}:ratelimit:comment:{ip}                 TTL=30s    -- 评论限流
site:{slug}:ratelimit:preview_gen:{user_id}        TTL=3600s  -- 预览令牌生成限流

# ============================================
# 站点配置缓存
# ============================================
site:{slug}:config:system                TTL=indefinite       -- 站点 sfc_site_configs（手动失效）

# ============================================
# Singleflight 锁（防缓存击穿）
# ============================================
site:{slug}:lock:post:{post_id}          TTL=5s              -- 重建锁
site:{slug}:lock:category:tree           TTL=5s              -- 重建锁

# ============================================
# 评论去重
# ============================================
site:{slug}:comment:dedup:{sha256}       TTL=3600s           -- 重复评论检测
```

---

## 4. 迁移管理

### 4.1 迁移工具

使用 **uptrace/bun** 内置迁移功能管理数据库版本（与 ORM 统一工具链，无需额外依赖）。

### 4.2 Schema 感知的迁移策略

多站点架构下，迁移分为三类：

| 类型 | 作用域 | 示例 |
|------|--------|------|
| 全局迁移 | `public` schema | 添加 `sfc_users` 列、创建新全局表 |
| 站点迁移 | 所有 `site_{slug}` schemas | 添加 `sfc_site_posts` 列、创建新站点表 |
| 混合迁移 | 两者 | 添加全局表 + 从站点表引用 |

### 4.3 迁移文件结构

```
migrations/
├── 20260224000001_create_core_tables.go        -- sfc_users, sfc_sites, sfc_refresh_tokens, sfc_password_reset_tokens, sfc_user_totp, sfc_configs
├── 20260224000002_create_rbac_tables.go        -- sfc_roles + sfc_user_roles + sfc_apis + sfc_role_apis + sfc_menus + sfc_role_menus + sfc_role_templates + sfc_role_template_apis + sfc_role_template_menus
├── 20260224000003_site_schema_placeholder.go   -- 占位符（站点 schema 由 internal/schema 动态创建）
├── 20260224000004_seed_rbac_builtins.go        -- Seed 4 内置角色 + 4 内置权限模板
└── ...
```

### 4.4 ForEachSiteSchema 辅助函数

当迁移需要应用到所有已有站点 schema 时，使用 `ForEachSiteSchema` 辅助函数：

```go
// ForEachSiteSchema 在事务中对每个站点 schema 执行迁移函数。
func ForEachSiteSchema(ctx context.Context, db *bun.DB, fn func(tx bun.Tx, schema string) error) error {
    // 1. 查询所有站点 slug
    var slugs []string
    err := db.NewSelect().
        TableExpr("public.sfc_sites").
        Column("slug").
        Scan(ctx, &slugs)
    if err != nil {
        return err
    }

    // 2. 对每个 schema 执行迁移
    for _, slug := range slugs {
        schema := "site_" + slug
        err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
            // 设置当前事务的 search_path
            _, err := tx.ExecContext(ctx,
                fmt.Sprintf("SET LOCAL search_path TO '%s', 'public'", schema))
            if err != nil {
                return err
            }
            return fn(tx, schema)
        })
        if err != nil {
            return fmt.Errorf("migration failed for schema %s: %w", schema, err)
        }
    }
    return nil
}
```

### 4.5 站点模板 DDL

完整的站点 DDL（本文档 Section 2B）维护为一个 Go 函数 `CreateSiteSchema(ctx, tx, slug)`，在以下场景调用：
- 安装向导创建第一个站点
- 通过 API 创建新站点

当新迁移向站点模板添加表/列时：
1. 更新 `CreateSiteSchema()` 以包含变更（用于新站点）
2. 编写迁移使用 `ForEachSiteSchema()` 将变更应用到已有站点

### 4.6 迁移命令

```bash
# 执行迁移
go run ./cmd/cms migrate up

# 回滚一步
go run ./cmd/cms migrate down

# 查看迁移状态
go run ./cmd/cms migrate status
```

> bun migrations 使用 Go 代码定义迁移（非纯 SQL 文件），支持 `db.NewCreateTable()` 等类型安全的 DDL 构建器，也可在迁移函数中直接执行原始 SQL。详见 [bun migrations 文档](https://bun.uptrace.dev/guide/migrations.html)。

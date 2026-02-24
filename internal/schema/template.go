package schema

// siteTemplateDDL contains the DDL for creating all tables within a site schema.
// The placeholder {schema} is replaced with the actual schema name (e.g. site_blog)
// before execution by CreateSiteSchema.
const siteTemplateDDL = `
-- 1. 内容类型表
CREATE TABLE {schema}.sfc_site_post_types (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    fields      JSONB NOT NULL DEFAULT '[]',
    built_in    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_site_post_types_slug ON {schema}.sfc_site_post_types(slug);

-- 2. 文章主表
CREATE TABLE {schema}.sfc_site_posts (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    author_id       UUID NOT NULL REFERENCES public.sfc_users(id),
    cover_image_id  UUID,
    post_type       VARCHAR(50) NOT NULL DEFAULT 'article',
    status          SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4),
    title           VARCHAR(500) NOT NULL,
    slug            VARCHAR(600) NOT NULL,
    excerpt         TEXT,
    content         TEXT,
    content_json    JSONB,
    meta_title       VARCHAR(200),
    meta_description VARCHAR(500),
    og_image_url     TEXT,
    extra_fields    JSONB DEFAULT '{}',
    view_count      BIGINT NOT NULL DEFAULT 0,
    version         INT    NOT NULL DEFAULT 1,
    published_at    TIMESTAMPTZ,
    scheduled_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_sfc_site_posts_slug      ON {schema}.sfc_site_posts(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_site_posts_author           ON {schema}.sfc_site_posts(author_id);
CREATE INDEX idx_sfc_site_posts_status           ON {schema}.sfc_site_posts(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_site_posts_published        ON {schema}.sfc_site_posts(published_at DESC)
    WHERE status = 3 AND deleted_at IS NULL;
CREATE INDEX idx_sfc_site_posts_extra            ON {schema}.sfc_site_posts USING gin(extra_fields);
CREATE INDEX idx_sfc_site_posts_scheduled        ON {schema}.sfc_site_posts(scheduled_at) WHERE status = 2;

-- 3. 文章多语言表
CREATE TABLE {schema}.sfc_site_post_translations (
    id               UUID PRIMARY KEY DEFAULT uuidv7(),
    post_id          UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    locale           VARCHAR(10) NOT NULL,
    title            VARCHAR(500),
    excerpt          TEXT,
    content          TEXT,
    content_json     JSONB,
    meta_title       VARCHAR(200),
    meta_description VARCHAR(500),
    og_image_url     VARCHAR(500),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(post_id, locale)
);

CREATE INDEX idx_sfc_site_pt_post_locale ON {schema}.sfc_site_post_translations(post_id, locale);

-- 4. 文章修订历史
CREATE TABLE {schema}.sfc_site_post_revisions (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    post_id      UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    editor_id    UUID NOT NULL REFERENCES public.sfc_users(id),
    version      INT NOT NULL,
    title        VARCHAR(500),
    content      TEXT,
    content_json JSONB,
    diff_summary TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_site_revisions_post   ON {schema}.sfc_site_post_revisions(post_id, version DESC);
CREATE INDEX idx_sfc_site_revisions_editor ON {schema}.sfc_site_post_revisions(editor_id);

-- 5. 分类表（Materialized Path）
CREATE TABLE {schema}.sfc_site_categories (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    parent_id   UUID REFERENCES {schema}.sfc_site_categories(id) ON DELETE RESTRICT,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(200) NOT NULL,
    path        TEXT NOT NULL DEFAULT '/',
    description TEXT,
    sort_order  INT NOT NULL DEFAULT 0,
    meta        JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(parent_id, slug)
);

CREATE INDEX idx_sfc_site_categories_parent ON {schema}.sfc_site_categories(parent_id);
CREATE INDEX idx_sfc_site_categories_path   ON {schema}.sfc_site_categories(path);
CREATE INDEX idx_sfc_site_categories_slug   ON {schema}.sfc_site_categories(parent_id, slug);

-- 6. 标签表
CREATE TABLE {schema}.sfc_site_tags (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    slug        VARCHAR(200) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 7. 文章-分类 多对多
CREATE TABLE {schema}.sfc_site_post_category_map (
    post_id     UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES {schema}.sfc_site_categories(id) ON DELETE CASCADE,
    is_primary  BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (post_id, category_id)
);

CREATE INDEX idx_sfc_site_pcm_category ON {schema}.sfc_site_post_category_map(category_id);

-- 8. 文章-标签 多对多
CREATE TABLE {schema}.sfc_site_post_tag_map (
    post_id UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES {schema}.sfc_site_tags(id)  ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

CREATE INDEX idx_sfc_site_ptm_tag ON {schema}.sfc_site_post_tag_map(tag_id);

-- 9. 媒体文件表
CREATE TABLE {schema}.sfc_site_media_files (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    uploader_id     UUID NOT NULL REFERENCES public.sfc_users(id),
    file_name       VARCHAR(500) NOT NULL,
    original_name   VARCHAR(500) NOT NULL,
    mime_type       VARCHAR(100) NOT NULL,
    media_type      SMALLINT NOT NULL DEFAULT 5 CHECK (media_type BETWEEN 1 AND 5),
    file_size       BIGINT NOT NULL,
    width           INT,
    height          INT,
    storage_path    TEXT NOT NULL,
    public_url      TEXT NOT NULL,
    webp_url        TEXT,
    thumbnail_urls  JSONB DEFAULT '{}',
    reference_count INT NOT NULL DEFAULT 0,
    alt_text        TEXT,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

ALTER TABLE {schema}.sfc_site_posts
    ADD CONSTRAINT fk_sfc_site_posts_cover_image
    FOREIGN KEY (cover_image_id) REFERENCES {schema}.sfc_site_media_files(id);

CREATE INDEX idx_sfc_site_media_uploader  ON {schema}.sfc_site_media_files(uploader_id);
CREATE INDEX idx_sfc_site_media_type      ON {schema}.sfc_site_media_files(media_type) WHERE deleted_at IS NULL;

-- 10. 评论表
CREATE TABLE {schema}.sfc_site_comments (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    post_id       UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    parent_id     UUID REFERENCES {schema}.sfc_site_comments(id) ON DELETE CASCADE,
    user_id       UUID REFERENCES public.sfc_users(id) ON DELETE SET NULL,
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
    CONSTRAINT chk_sfc_site_comment_content_length CHECK (length(content) BETWEEN 1 AND 10000),
    CONSTRAINT chk_sfc_site_comment_guest CHECK (
        user_id IS NOT NULL OR (author_name IS NOT NULL AND author_email IS NOT NULL)
    )
);

CREATE INDEX idx_sfc_site_comments_post_status ON {schema}.sfc_site_comments(post_id, status, created_at)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_site_comments_parent      ON {schema}.sfc_site_comments(parent_id)
    WHERE parent_id IS NOT NULL;
CREATE INDEX idx_sfc_site_comments_moderation  ON {schema}.sfc_site_comments(status, created_at DESC)
    WHERE status = 1 AND deleted_at IS NULL;
CREATE INDEX idx_sfc_site_comments_email       ON {schema}.sfc_site_comments(author_email);
CREATE INDEX idx_sfc_site_comments_user        ON {schema}.sfc_site_comments(user_id)
    WHERE user_id IS NOT NULL AND deleted_at IS NULL;

-- 11. 导航菜单表
CREATE TABLE {schema}.sfc_site_menus (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    location    VARCHAR(50),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 12. 菜单项表
CREATE TABLE {schema}.sfc_site_menu_items (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    menu_id       UUID NOT NULL REFERENCES {schema}.sfc_site_menus(id) ON DELETE CASCADE,
    parent_id     UUID REFERENCES {schema}.sfc_site_menu_items(id) ON DELETE CASCADE,
    label         VARCHAR(200) NOT NULL,
    url           TEXT,
    target        VARCHAR(10) NOT NULL DEFAULT '_self',
    type          SMALLINT NOT NULL DEFAULT 1 CHECK (type BETWEEN 1 AND 5),
    reference_id  UUID,
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

-- 13. URL 重定向表
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

-- 14. 草稿预览令牌表
CREATE TABLE {schema}.sfc_site_preview_tokens (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    post_id     UUID NOT NULL REFERENCES {schema}.sfc_site_posts(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_by  UUID REFERENCES public.sfc_users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sfc_site_preview_tokens_post    ON {schema}.sfc_site_preview_tokens(post_id);
CREATE INDEX idx_sfc_site_preview_tokens_hash    ON {schema}.sfc_site_preview_tokens(token_hash)
    WHERE expires_at > NOW();
CREATE INDEX idx_sfc_site_preview_tokens_expires ON {schema}.sfc_site_preview_tokens(expires_at);

-- 15. API Key 表
CREATE TABLE {schema}.sfc_site_api_keys (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    owner_id     UUID NOT NULL REFERENCES public.sfc_users(id),
    name         VARCHAR(100) NOT NULL,
    key_hash     VARCHAR(255) NOT NULL UNIQUE,
    key_prefix   VARCHAR(20) NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    rate_limit   INT NOT NULL DEFAULT 100,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX idx_sfc_site_apikeys_owner ON {schema}.sfc_site_api_keys(owner_id);
CREATE INDEX idx_sfc_site_apikeys_hash  ON {schema}.sfc_site_api_keys(key_hash);

-- 16. 操作审计日志（按月分区）
CREATE TABLE {schema}.sfc_site_audits (
    id                UUID NOT NULL DEFAULT uuidv7(),
    actor_id          UUID REFERENCES public.sfc_users(id),
    actor_email       VARCHAR(255),
    action            SMALLINT NOT NULL CHECK (action BETWEEN 1 AND 11),
    resource_type     VARCHAR(50) NOT NULL,
    resource_id       TEXT,
    resource_snapshot JSONB,
    ip_address        INET,
    user_agent        TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- 17. 系统配置表（站点级别）
CREATE TABLE {schema}.sfc_site_configs (
    key         VARCHAR(100) PRIMARY KEY,
    value       JSONB NOT NULL,
    description TEXT,
    updated_by  UUID REFERENCES public.sfc_users(id),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO {schema}.sfc_site_configs (key, value, description) VALUES
('media.max_size',     '104857600',  '最大文件大小（bytes）'),
('media.storage',      '"rustfs"',   '存储驱动：rustfs（S3 兼容）'),
('content.trash_days', '30',         '回收站保留天数')
ON CONFLICT (key) DO NOTHING;
`

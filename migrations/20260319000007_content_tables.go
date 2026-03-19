package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
-- v1 content tables: all in public schema, sfc_ prefix, no site_id.
-- v2 multi-site will move these to site_{slug} schemas.

-- Categories (tree via materialized path)
CREATE TABLE public.sfc_categories (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(200) NOT NULL,
    slug        VARCHAR(200) NOT NULL UNIQUE,
    description TEXT,
    parent_id   UUID REFERENCES public.sfc_categories(id) ON DELETE SET NULL,
    path        TEXT NOT NULL DEFAULT '',   -- materialized path: /uuid/uuid/
    depth       SMALLINT NOT NULL DEFAULT 0,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Tags (flat)
CREATE TABLE public.sfc_tags (
    id         UUID PRIMARY KEY DEFAULT uuidv7(),
    name       VARCHAR(100) NOT NULL,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Posts (state machine: draft → published → archived)
CREATE TYPE post_status AS ENUM ('draft', 'published', 'archived', 'scheduled');
CREATE TYPE post_type   AS ENUM ('article', 'page');

CREATE TABLE public.sfc_posts (
    id               UUID PRIMARY KEY DEFAULT uuidv7(),
    title            VARCHAR(500) NOT NULL,
    slug             VARCHAR(500) NOT NULL UNIQUE,
    excerpt          TEXT,
    content          TEXT,
    content_json     JSONB,
    cover_image_url  TEXT,
    status           post_status NOT NULL DEFAULT 'draft',
    type             post_type NOT NULL DEFAULT 'article',
    author_id        UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE RESTRICT,
    category_id      UUID REFERENCES public.sfc_categories(id) ON DELETE SET NULL,
    published_at     TIMESTAMPTZ,
    scheduled_at     TIMESTAMPTZ,
    view_count       INTEGER NOT NULL DEFAULT 0,
    comment_count    INTEGER NOT NULL DEFAULT 0,
    is_featured      BOOLEAN NOT NULL DEFAULT FALSE,
    allow_comments   BOOLEAN NOT NULL DEFAULT TRUE,
    meta_title       VARCHAR(500),
    meta_description TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

-- Post ↔ Tag many-to-many
CREATE TABLE public.sfc_post_tags (
    post_id UUID NOT NULL REFERENCES public.sfc_posts(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES public.sfc_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

-- Post revisions (immutable history)
CREATE TABLE public.sfc_post_revisions (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    post_id      UUID NOT NULL REFERENCES public.sfc_posts(id) ON DELETE CASCADE,
    title        VARCHAR(500) NOT NULL,
    content      TEXT,
    content_json JSONB,
    author_id    UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE RESTRICT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Media files
CREATE TYPE media_status AS ENUM ('active', 'deleted');

CREATE TABLE public.sfc_media_files (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    filename        VARCHAR(500) NOT NULL,
    original_name   VARCHAR(500) NOT NULL,
    mime_type       VARCHAR(100) NOT NULL,
    size_bytes      BIGINT NOT NULL,
    storage_key     TEXT NOT NULL UNIQUE,  -- RustFS object key
    url             TEXT NOT NULL,
    width           INTEGER,
    height          INTEGER,
    thumb_sm_url    TEXT,                  -- 150×150 crop
    thumb_md_url    TEXT,                  -- 400×400 fit
    alt_text        TEXT,
    caption         TEXT,
    uploader_id     UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE RESTRICT,
    status          media_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Comments (3-level nesting max, moderated)
CREATE TYPE comment_status AS ENUM ('pending', 'approved', 'spam', 'rejected');

CREATE TABLE public.sfc_comments (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    post_id      UUID NOT NULL REFERENCES public.sfc_posts(id) ON DELETE CASCADE,
    parent_id    UUID REFERENCES public.sfc_comments(id) ON DELETE CASCADE,
    depth        SMALLINT NOT NULL DEFAULT 0 CHECK (depth <= 2),
    author_name  VARCHAR(100) NOT NULL,
    author_email VARCHAR(255) NOT NULL,
    author_url   TEXT,
    author_ip    INET,
    content      TEXT NOT NULL,
    status       comment_status NOT NULL DEFAULT 'pending',
    is_pinned    BOOLEAN NOT NULL DEFAULT FALSE,
    is_admin     BOOLEAN NOT NULL DEFAULT FALSE,  -- TRUE = admin reply
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Navigation menus
CREATE TABLE public.sfc_menus (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(200) NOT NULL,
    slug        VARCHAR(200) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE public.sfc_menu_items (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    menu_id     UUID NOT NULL REFERENCES public.sfc_menus(id) ON DELETE CASCADE,
    parent_id   UUID REFERENCES public.sfc_menu_items(id) ON DELETE CASCADE,
    depth       SMALLINT NOT NULL DEFAULT 0 CHECK (depth <= 2),
    label       VARCHAR(200) NOT NULL,
    url         TEXT NOT NULL,
    target      VARCHAR(20) NOT NULL DEFAULT '_self',
    icon        VARCHAR(50),
    css_class   VARCHAR(100),
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- URL redirects
CREATE TYPE redirect_type AS ENUM ('301', '302');

CREATE TABLE public.sfc_redirects (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    from_path   TEXT NOT NULL UNIQUE,
    to_path     TEXT NOT NULL,
    type        redirect_type NOT NULL DEFAULT '301',
    hit_count   INTEGER NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit log (immutable)
CREATE TABLE public.sfc_audits (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id     UUID REFERENCES public.sfc_users(id) ON DELETE SET NULL,
    action      VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id   UUID,
    meta        JSONB NOT NULL DEFAULT '{}',
    ip_address  INET,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
        `)
		if err != nil {
			return fmt.Errorf("create content tables: %w", err)
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
DROP TABLE IF EXISTS public.sfc_audits CASCADE;
DROP TABLE IF EXISTS public.sfc_redirects CASCADE;
DROP TABLE IF EXISTS public.sfc_menu_items CASCADE;
DROP TABLE IF EXISTS public.sfc_menus CASCADE;
DROP TABLE IF EXISTS public.sfc_comments CASCADE;
DROP TABLE IF EXISTS public.sfc_media_files CASCADE;
DROP TABLE IF EXISTS public.sfc_post_revisions CASCADE;
DROP TABLE IF EXISTS public.sfc_post_tags CASCADE;
DROP TABLE IF EXISTS public.sfc_posts CASCADE;
DROP TABLE IF EXISTS public.sfc_tags CASCADE;
DROP TABLE IF EXISTS public.sfc_categories CASCADE;
DROP TYPE IF EXISTS redirect_type CASCADE;
DROP TYPE IF EXISTS comment_status CASCADE;
DROP TYPE IF EXISTS media_status CASCADE;
DROP TYPE IF EXISTS post_type CASCADE;
DROP TYPE IF EXISTS post_status CASCADE;
        `)
		if err != nil {
			return fmt.Errorf("drop content tables: %w", err)
		}
		return nil
	})
}

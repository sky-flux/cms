package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
-- sfc_posts indexes
CREATE INDEX idx_sfc_posts_status     ON public.sfc_posts(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_posts_author     ON public.sfc_posts(author_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_posts_category   ON public.sfc_posts(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_posts_published  ON public.sfc_posts(published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;
CREATE INDEX idx_sfc_posts_scheduled  ON public.sfc_posts(scheduled_at) WHERE status = 'scheduled' AND deleted_at IS NULL;
CREATE INDEX idx_sfc_posts_featured   ON public.sfc_posts(is_featured) WHERE is_featured = TRUE AND deleted_at IS NULL;

-- sfc_categories indexes
CREATE INDEX idx_sfc_cats_parent      ON public.sfc_categories(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_cats_path        ON public.sfc_categories(path) WHERE deleted_at IS NULL;

-- sfc_tags indexes
CREATE INDEX idx_sfc_tags_slug        ON public.sfc_tags(slug) WHERE deleted_at IS NULL;

-- sfc_post_tags indexes
CREATE INDEX idx_sfc_pt_tag_id        ON public.sfc_post_tags(tag_id);

-- sfc_post_revisions indexes
CREATE INDEX idx_sfc_revisions_post   ON public.sfc_post_revisions(post_id, created_at DESC);

-- sfc_media_files indexes
CREATE INDEX idx_sfc_media_uploader   ON public.sfc_media_files(uploader_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_media_mime       ON public.sfc_media_files(mime_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_sfc_media_status     ON public.sfc_media_files(status) WHERE deleted_at IS NULL;

-- sfc_comments indexes
CREATE INDEX idx_sfc_comments_post    ON public.sfc_comments(post_id, status);
CREATE INDEX idx_sfc_comments_parent  ON public.sfc_comments(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_sfc_comments_status  ON public.sfc_comments(status);
CREATE INDEX idx_sfc_comments_email   ON public.sfc_comments(author_email);

-- sfc_menu_items indexes
CREATE INDEX idx_sfc_mi_menu         ON public.sfc_menu_items(menu_id, sort_order);
CREATE INDEX idx_sfc_mi_parent       ON public.sfc_menu_items(parent_id) WHERE parent_id IS NOT NULL;

-- sfc_redirects indexes
CREATE INDEX idx_sfc_redirects_from  ON public.sfc_redirects(from_path) WHERE is_active = TRUE;

-- sfc_audits indexes
CREATE INDEX idx_sfc_audits_user     ON public.sfc_audits(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_sfc_audits_entity   ON public.sfc_audits(entity_type, entity_id) WHERE entity_id IS NOT NULL;
CREATE INDEX idx_sfc_audits_created  ON public.sfc_audits(created_at DESC);

-- updated_at auto-update trigger for content tables
-- (update_updated_at() function created in migration 1)
CREATE TRIGGER trg_sfc_categories_updated_at
    BEFORE UPDATE ON public.sfc_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_sfc_tags_updated_at
    BEFORE UPDATE ON public.sfc_tags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_sfc_posts_updated_at
    BEFORE UPDATE ON public.sfc_posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_sfc_media_files_updated_at
    BEFORE UPDATE ON public.sfc_media_files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_sfc_comments_updated_at
    BEFORE UPDATE ON public.sfc_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_sfc_menus_updated_at
    BEFORE UPDATE ON public.sfc_menus
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_sfc_menu_items_updated_at
    BEFORE UPDATE ON public.sfc_menu_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_sfc_redirects_updated_at
    BEFORE UPDATE ON public.sfc_redirects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
        `)
		if err != nil {
			return fmt.Errorf("create content indexes and triggers: %w", err)
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
-- Triggers drop with their tables in migration 7 rollback.
-- Explicit index drops here in case partial rollback is needed.
DROP INDEX IF EXISTS public.idx_sfc_redirects_from;
DROP INDEX IF EXISTS public.idx_sfc_mi_parent;
DROP INDEX IF EXISTS public.idx_sfc_mi_menu;
DROP INDEX IF EXISTS public.idx_sfc_audits_created;
DROP INDEX IF EXISTS public.idx_sfc_audits_entity;
DROP INDEX IF EXISTS public.idx_sfc_audits_user;
DROP INDEX IF EXISTS public.idx_sfc_comments_email;
DROP INDEX IF EXISTS public.idx_sfc_comments_status;
DROP INDEX IF EXISTS public.idx_sfc_comments_parent;
DROP INDEX IF EXISTS public.idx_sfc_comments_post;
DROP INDEX IF EXISTS public.idx_sfc_media_status;
DROP INDEX IF EXISTS public.idx_sfc_media_mime;
DROP INDEX IF EXISTS public.idx_sfc_media_uploader;
DROP INDEX IF EXISTS public.idx_sfc_revisions_post;
DROP INDEX IF EXISTS public.idx_sfc_pt_tag_id;
DROP INDEX IF EXISTS public.idx_sfc_tags_slug;
DROP INDEX IF EXISTS public.idx_sfc_cats_path;
DROP INDEX IF EXISTS public.idx_sfc_cats_parent;
DROP INDEX IF EXISTS public.idx_sfc_posts_featured;
DROP INDEX IF EXISTS public.idx_sfc_posts_scheduled;
DROP INDEX IF EXISTS public.idx_sfc_posts_published;
DROP INDEX IF EXISTS public.idx_sfc_posts_category;
DROP INDEX IF EXISTS public.idx_sfc_posts_author;
DROP INDEX IF EXISTS public.idx_sfc_posts_status;
        `)
		if err != nil {
			return fmt.Errorf("drop content indexes: %w", err)
		}
		return nil
	})
}

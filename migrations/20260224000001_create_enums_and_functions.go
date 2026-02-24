package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
-- 枚举类型（在 public schema 中创建，所有 site schema 通过 search_path 共享）
CREATE TYPE post_status    AS ENUM ('draft', 'scheduled', 'published', 'archived');
CREATE TYPE media_type     AS ENUM ('image', 'video', 'audio', 'document', 'other');
CREATE TYPE comment_status AS ENUM ('pending', 'approved', 'spam', 'trash');
CREATE TYPE menu_item_type AS ENUM ('custom', 'post', 'category', 'tag', 'page');
CREATE TYPE log_action     AS ENUM (
    'create', 'update', 'delete', 'restore',
    'login', 'logout', 'publish', 'unpublish',
    'archive', 'password_change', 'settings_change'
);

-- 公共触发器函数（所有 schema 共享）
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
		`)
		if err != nil {
			return fmt.Errorf("create enums and functions: %w", err)
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
DROP FUNCTION IF EXISTS update_updated_at() CASCADE;
DROP TYPE IF EXISTS log_action CASCADE;
DROP TYPE IF EXISTS menu_item_type CASCADE;
DROP TYPE IF EXISTS comment_status CASCADE;
DROP TYPE IF EXISTS media_type CASCADE;
DROP TYPE IF EXISTS post_status CASCADE;
		`)
		if err != nil {
			return fmt.Errorf("drop enums and functions: %w", err)
		}
		return nil
	})
}

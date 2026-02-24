package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Seed 4 built-in roles
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

		// Seed 4 built-in role templates
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

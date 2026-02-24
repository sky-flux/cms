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

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

CREATE TRIGGER trg_sfc_users_updated_at
    BEFORE UPDATE ON public.sfc_users FOR EACH ROW EXECUTE FUNCTION update_updated_at();

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

    CONSTRAINT chk_sfc_sites_slug CHECK (slug ~ '^[a-z0-9_]{3,50}$')
);

CREATE INDEX idx_sfc_sites_domain ON public.sfc_sites(domain) WHERE domain IS NOT NULL;
CREATE INDEX idx_sfc_sites_active ON public.sfc_sites(is_active) WHERE is_active = TRUE;

CREATE TRIGGER trg_sfc_sites_updated_at
    BEFORE UPDATE ON public.sfc_sites FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- 3. 角色定义表
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

-- 4. 用户-角色分配表
CREATE TABLE public.sfc_user_roles (
    user_id    UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES public.sfc_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_sfc_user_roles_role ON public.sfc_user_roles(role_id);

-- 5. API 端点注册表
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

-- 6. 角色-API 权限映射表
CREATE TABLE public.sfc_role_apis (
    role_id UUID NOT NULL REFERENCES public.sfc_roles(id) ON DELETE CASCADE,
    api_id  UUID NOT NULL REFERENCES public.sfc_apis(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, api_id)
);

CREATE INDEX idx_sfc_role_apis_api ON public.sfc_role_apis(api_id);

-- 7. 后台管理菜单表
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

-- 8. 角色-菜单可见性映射表
CREATE TABLE public.sfc_role_menus (
    role_id UUID NOT NULL REFERENCES public.sfc_roles(id) ON DELETE CASCADE,
    menu_id UUID NOT NULL REFERENCES public.sfc_menus(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, menu_id)
);

CREATE INDEX idx_sfc_role_menus_menu ON public.sfc_role_menus(menu_id);

-- 9. 权限模板定义表
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

-- 10. 模板-API 映射表
CREATE TABLE public.sfc_role_template_apis (
    template_id UUID NOT NULL REFERENCES public.sfc_role_templates(id) ON DELETE CASCADE,
    api_id      UUID NOT NULL REFERENCES public.sfc_apis(id) ON DELETE CASCADE,
    PRIMARY KEY (template_id, api_id)
);

-- 11. 模板-菜单映射表
CREATE TABLE public.sfc_role_template_menus (
    template_id UUID NOT NULL REFERENCES public.sfc_role_templates(id) ON DELETE CASCADE,
    menu_id     UUID NOT NULL REFERENCES public.sfc_menus(id) ON DELETE CASCADE,
    PRIMARY KEY (template_id, menu_id)
);

-- 12. 刷新令牌表
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

CREATE TRIGGER trg_sfc_user_totp_updated_at
    BEFORE UPDATE ON public.sfc_user_totp FOR EACH ROW EXECUTE FUNCTION update_updated_at();

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
			return fmt.Errorf("create public schema: %w", err)
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
DROP TABLE IF EXISTS public.sfc_configs CASCADE;
DROP TABLE IF EXISTS public.sfc_user_totp CASCADE;
DROP TABLE IF EXISTS public.sfc_refresh_tokens CASCADE;
DROP TABLE IF EXISTS public.sfc_role_template_menus CASCADE;
DROP TABLE IF EXISTS public.sfc_role_template_apis CASCADE;
DROP TABLE IF EXISTS public.sfc_role_templates CASCADE;
DROP TABLE IF EXISTS public.sfc_role_menus CASCADE;
DROP TABLE IF EXISTS public.sfc_menus CASCADE;
DROP TABLE IF EXISTS public.sfc_role_apis CASCADE;
DROP TABLE IF EXISTS public.sfc_apis CASCADE;
DROP TABLE IF EXISTS public.sfc_user_roles CASCADE;
DROP TABLE IF EXISTS public.sfc_roles CASCADE;
DROP TABLE IF EXISTS public.sfc_sites CASCADE;
DROP TABLE IF EXISTS public.sfc_users CASCADE;
		`)
		if err != nil {
			return fmt.Errorf("drop public schema: %w", err)
		}
		return nil
	})
}

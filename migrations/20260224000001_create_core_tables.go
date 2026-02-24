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

# Sky Flux CMS — Claude Code 项目指令

> **项目阶段: 基础脚手架已完成 → 功能开发中**
> 核心骨架已就绪：CLI (Cobra)、配置 (Viper)、数据库连接、bun 迁移框架、中间件链、路由。
> 编码前务必先阅读 `docs/` 下的相关设计文档。

## 技术栈

```
后端: Go 1.25+ / Gin v1.11+ / uptrace/bun (ORM) / PostgreSQL 18 / Redis 8 / Meilisearch / RustFS
前端: Astro 5 SSR + React 19 + shadcn/ui + TanStack Query v5 + Zustand + Tailwind V4
认证: JWT HS256 (15min) + Refresh Token (7d httpOnly Cookie) + TOTP 2FA
日志: log/slog (Go 标准库)
测试: testify + testcontainers-go + miniredis / Vitest + RTL + Playwright / k6
```

## 开发环境要求

```
Go 1.25+    | Docker 27+         | PostgreSQL 18
Bun 1.2+    | Docker Compose 2+  | Redis 8 / Meilisearch / RustFS
```

> 详细搭建步骤见 `docs/setup.md`

## 关键决策（务必遵守）

| 决策 | 选择 | 禁止 |
|------|------|------|
| ORM | **uptrace/bun** (链式查询) | sqlx, gorm, ent |
| 日志 | **log/slog** | zap, logrus, zerolog |
| UUID | **PG18 原生 `uuidv7()`** | UUIDv4, gen_random_uuid() |
| 迁移 | **bun 内置 migrations** (Go code) | golang-migrate, goose |
| 多站点 | **Schema 隔离** (`site_{slug}` schema) | shared table + site_id |
| 前端包管理 | **bun** | npm, pnpm, yarn |
| 搜索引擎 | **Meilisearch** (独立全文搜索) | PostgreSQL FTS, Elasticsearch |
| 对象存储 | **RustFS** (S3 兼容, AWS SDK v2) | 本地文件系统, MinIO |
| CLI 框架 | **Cobra** + Viper | 裸 flag 包, urfave/cli |

## 多站点架构

- `public` schema: sfc_users, sfc_sites, sfc_site_user_roles, sfc_refresh_tokens, sfc_user_totp, sfc_configs（`sfc_` 前缀标识 Sky Flux CMS 专属表）
- `site_{slug}` schema: sfc_site_posts, sfc_site_categories, sfc_site_tags, sfc_site_media_files, sfc_site_comments, sfc_site_menus, sfc_site_redirects, sfc_site_preview_tokens, sfc_site_api_keys, sfc_site_audits, sfc_site_configs
- 中间件链: InstallationGuard → SiteResolver → SchemaMiddleware (`SET search_path TO 'site_{slug}', 'public'`) → Auth → RBAC
- 内容表无 `site_id` 列 — Schema 本身即隔离边界

## 代码规范摘要

### Go

- 包名小写单词，接口用 -er 后缀，缩写全大写 (`userID` 非 `userId`)
- 错误跨层传递用 `fmt.Errorf("context: %w", err)`
- 哨兵错误定义在 `internal/pkg/apperror/`
- Repository 层返回 `apperror.ErrNotFound`（非原始 sql.ErrNoRows）
- 密码 bcrypt cost=12, API Key SHA-256, TOTP 密钥 AES-256-GCM 加密
- Schema slug 校验: `^[a-z0-9_]{3,50}$`，永不将原始用户输入拼入 search_path

### uptrace/bun 用法

```go
// 查询
db.NewSelect().Model(&post).Where("id = ?", id).Scan(ctx)
db.NewSelect().Model(&posts).Where("status = ?", status).Limit(20).Scan(ctx)

// 写入
db.NewInsert().Model(&post).Exec(ctx)
db.NewUpdate().Model(&post).WherePK().Exec(ctx)
db.NewDelete().Model(&post).WherePK().Exec(ctx)
```

- 占位符用 `?`（bun 自动转换为 PostgreSQL 的 `$1, $2...`）
- 软删除: model 嵌入 `bun.BaseModel` + `DeletedAt *time.Time \`bun:"deleted_at,soft_delete,nullzero"\``
- 迁移: `migrations/main.go` 注册表 + `init()` + `MustRegister(up, down)` 模式
- 事务: `db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error { ... })`
- 生产配置: `bun.WithDiscardUnknownColumns()` + 连接池超时 (`ConnMaxLifetime`, `ConnMaxIdleTime`)

### 前端

- Astro Islands 架构: 仅交互组件加载 React Runtime
- 状态管理: Zustand (全局) + TanStack Query (服务端状态)
- UI 组件: shadcn/ui (Radix UI + Tailwind CSS V4)
- 错误提示: Sonner Toast
- 界面语言: react-i18next (zh-CN / en)
- 代码质量: Biome (lint + format, 替代 ESLint + Prettier)

## 目录结构

```
sky-flux-cms/                          # Go 单体多模块分层架构
├── cmd/
│   └── cms/                           # Cobra CLI 单一二进制
│       ├── main.go                    # 入口
│       ├── root.go                    # rootCmd + --config flag
│       ├── serve.go                   # cms serve (HTTP 服务)
│       ├── migrate.go                 # cms migrate up|down|status
│       └── version.go                 # cms version
├── internal/
│   ├── auth/                          # 认证模块（handler/service/repo/dto 自包含）
│   ├── user/                          # 用户模块
│   ├── post/                          # 文章模块
│   ├── category/ tag/ media/          # 分类 / 标签 / 媒体
│   ├── comment/ menu/ redirect/       # 评论 / 菜单 / 重定向
│   ├── preview/ site/ setup/          # 预览 / 站点 / 安装向导
│   ├── feed/ apikey/ audit/ system/   # 订阅 / API Key / 审计 / 系统
│   ├── model/                         # 共享数据模型
│   ├── middleware/                    # 共享 Gin 中间件
│   ├── config/                        # Viper 配置加载
│   ├── database/                      # DB + Redis + RustFS 连接
│   ├── schema/                        # 站点 Schema 管理
│   ├── router/                        # 路由注册（唯一组装点）
│   ├── cron/                          # 定时任务
│   └── pkg/                           # 共享工具包 (apperror/jwt/crypto/slug)
├── migrations/                        # bun Go code migrations
├── web/                               # Astro 管理后台
│   ├── src/
│   │   ├── components/               # React 组件 (shadcn/ui)
│   │   ├── pages/                    # Astro 页面 (SSR)
│   │   ├── lib/                      # API 客户端、工具函数
│   │   ├── hooks/                    # React Hooks
│   │   ├── stores/                   # Zustand 状态
│   │   └── i18n/                     # 国际化资源
│   └── astro.config.mjs
├── docs/                              # 设计文档 (10 份)
├── go.mod
├── Makefile
└── docker-compose.yml
```

## 常用命令

```bash
make setup          # 一键初始化 (.env + PG/Redis + 依赖安装)
make dev            # 启动开发环境 (后端热重载 + 前端)
make test           # 运行全部测试
make migrate-up     # 执行数据库迁移 (public + 所有 site schemas)
make migrate-down   # 回滚最近一次迁移
make lint           # 代码检查 (golangci-lint + Biome)

# Cobra CLI 直接调用
go run ./cmd/cms serve --port 9090     # 指定端口启动
go run ./cmd/cms migrate init          # 创建迁移元数据表
go run ./cmd/cms migrate up            # 执行迁移
go run ./cmd/cms version               # 版本信息
go run ./cmd/cms --help                # 查看所有命令
```

## 设计文档

所有设计文档位于 `docs/`，编码前应先阅读相关文档。

**编码工作流**: 实现某个功能前，按此顺序阅读：
1. `story.md` — 找到对应用户故事和验收标准
2. `api.md` — 确认 API 端点设计
3. `database.md` — 确认数据模型和索引
4. `standard.md` — 遵循编码规范

| 文档 | 内容 |
|------|------|
| prd.md | 产品需求、功能范围、权限矩阵 |
| architecture.md | 系统架构、目录结构、技术决策 |
| api.md | 完整 API 设计 (OpenAPI 3.1) |
| database.md | 数据库 Schema、ER 图、索引策略 |
| story.md | 用户故事与验收标准 |
| testing.md | 测试策略、用例模板 |
| deployment.md | 部署指南、环境变量清单 |
| standard.md | 编码规范、日志规范 |
| security.md | 安全策略、威胁模型 |
| setup.md | 开发环境搭建指南 |

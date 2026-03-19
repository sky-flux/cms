# Sky Flux CMS — Claude Code 项目指令

> **项目阶段: 架构重设计 → DDD 重构中**
> 技术栈已从 Gin+Viper+Astro 迁移至 Chi v5+Huma v2+koanf+Templ+React SPA。
> 编码前务必先阅读 `docs/superpowers/specs/` 下的最新设计文档。

## Commit 规范

使用 [gitmoji](https://gitmoji.dev/) 规范，**禁止包含 Co-Authored-By 或任何 AI 标识**：

```
<gitmoji> <简短描述>（英文，72字符以内）
```

| Emoji | 用途 |
|-------|------|
| ✨ | 新功能 |
| 🐛 | 修复 bug |
| ⚡ | 性能优化 |
| ♻️ | 重构 |
| ✅ | 添加/更新测试 |
| 📝 | 文档 |
| 🔧 | 配置文件 |
| 🚚 | 移动/重命名文件 |
| 🔥 | 删除代码/文件 |
| ⬆️ | 升级依赖 |
| 💥 | Breaking change |
| 🎉 | 初始提交 |

完整列表见 https://gitmoji.dev/

## 技术栈

```
后端 API:  Go 1.25+ / Chi v5 / Huma v2 / uptrace/bun (ORM) / PostgreSQL 18 / Redis 8 / Meilisearch / RustFS / Resend
配置管理:  koanf v2（多源合并：CLI flag > ENV > .env > 默认值）
管理后台:  React 19 SPA + TanStack Router v1 + TanStack Query v5 + TanStack Form + shadcn/ui + Paraglide (i18n) + Tailwind V4
公共站点:  Go Templ + HTMX 2.x + Tailwind CSS V4
认证:      JWT HS256 (15min) + Refresh Token (7d httpOnly Cookie) + TOTP 2FA
日志:      log/slog (Go 标准库)
测试:      testify + testcontainers-go + miniredis / Vitest + RTL + Playwright
```

## 开发环境要求

```
Go 1.25+    | Docker 27+ (colima)  | PostgreSQL 18
Bun 1.2+    | Docker Compose 2+    | Redis 8 / Meilisearch / RustFS
overmind    | templ CLI            | tailwindcss CLI
```

## 关键决策（务必遵守）

| 决策 | 选择 | 禁止 |
|------|------|------|
| HTTP 路由 | **Chi v5** | Gin, Echo, Fiber |
| API 框架 | **Huma v2**（OpenAPI 3.1 自动生成，RFC 9457 错误） | 手写 OpenAPI |
| ORM | **uptrace/bun**（链式查询） | sqlx, gorm, ent |
| 配置 | **koanf v2**（多源合并） | Viper, envconfig |
| 日志 | **log/slog** | zap, logrus, zerolog |
| UUID | **PG18 原生 `uuidv7()`** | UUIDv4, gen_random_uuid() |
| 迁移 | **bun 内置 migrations**（Go code） | golang-migrate, goose |
| 站点架构 | **v1 单站点**（所有表在 public schema） | 多站点 schema 隔离（v2 延后） |
| 前端包管理 | **bun** | npm, pnpm, yarn |
| 管理后台 | **React 19 SPA**（TanStack Router，纯客户端渲染） | Astro SSR, Next.js, TanStack Start SSR |
| 公共站点 | **Go Templ + HTMX**（SSR，go:embed） | React SSR, Astro |
| i18n | **Paraglide**（编译时类型安全） | react-i18next, i18next |
| 状态管理 | **TanStack Query v5 + TanStack Store** | Zustand, Redux |
| 表单 | **TanStack Form + Zod** | react-hook-form |
| 搜索引擎 | **Meilisearch**（独立全文搜索） | PostgreSQL FTS, Elasticsearch |
| 对象存储 | **RustFS**（S3 兼容，AWS SDK v2） | 本地文件系统, MinIO |
| CLI 框架 | **Cobra**（serve / install / migrate 子命令） | 裸 flag 包, urfave/cli |
| 邮件服务 | **Resend**（HTTP API，Go SDK v3） | SMTP, SendGrid, SES |
| 代码质量 | **Biome**（lint + format，替代 ESLint + Prettier） | ESLint, Prettier |
| 模板引擎 | **Templ**（类型安全，编译时检查） | html/template, pongo2 |

## 架构概览

### 三个子项目（Monorepo，单二进制分发）

| 项目 | 技术栈 | 职责 |
|------|--------|------|
| `cmd/cms/` + `internal/` | Go + Chi v5 + Huma v2 | REST API 后端 + Web SSR |
| `console/` | React 19 SPA + TanStack Router | 管理后台（构建产物 go:embed 嵌入二进制） |
| `web/` | Go Templ + HTMX 2.x | 公共站点（模板编译进二进制） |

### DDD 界限上下文（Bounded Contexts）

| BC | 聚合根 | 职责 |
|----|--------|------|
| **identity** | User, Role | 用户、认证（JWT+TOTP）、RBAC 权限 |
| **content** | Post, Category, Tag, Comment | 文章生命周期、分类树、评论审核 |
| **media** | MediaFile | 上传、缩略图、RustFS 存储 |
| **site** | Site, Menu, Redirect | 站点配置、导航菜单、URL 重定向 |
| **delivery** | - | Public API、RSS/Atom、Sitemap |
| **platform** | AuditLog, SystemConfig | 安装向导、审计日志、系统配置 |

### 各 BC 内部分层

```
internal/{bc}/
├── domain/          # 实体、值对象、领域事件、Repository 接口（无框架依赖）
├── app/             # 应用服务 / 用例（编排 domain，依赖 domain 接口）
├── infra/           # 具体实现（bun repo、Redis、Meilisearch、RustFS）
└── delivery/        # Huma handler + DTO（HTTP 边界）
```

依赖方向: `delivery → app → domain ← infra`

### v1 数据库 Schema 策略

v1 **单站点**，所有表在 PostgreSQL 默认 `public` schema，无多站点 Schema 隔离：

- 全局表: `sfc_` 前缀（`sfc_users`、`sfc_roles` 等）
- 内容表: `sfc_` 前缀（`sfc_posts`、`sfc_categories` 等）——v1 无 `site_` 中缀
- v2 多站点时内容表迁移到 `site_{slug}` schema 并加 `sfc_site_` 前缀

### Chi 路由分发

```
GET  /setup              → 安装向导（未安装时 InstallGuard 拦截）
GET  /console/*          → React SPA（go:embed，SPA fallback index.html）

/api/v1/
├── /public/*            → 公共 API（Rate Limit + API Key 可选）
│   ├── /posts           → 文章列表/详情
│   ├── /search          → Meilisearch 全文搜索
│   ├── /comments        → 评论提交
│   └── /feed/*          → RSS/Atom/Sitemap
└── /admin/              → 管理 API（JWT + RBAC）
    ├── /auth/*          → 登录/刷新/登出（无需 JWT）
    └── 所有管理端点     → 需要 JWT + RBAC

GET  /*                  → Templ + HTMX 公共站点（SSR）
```

### 单二进制分发（go:embed）

```go
// embed.go（项目根目录）
package cms

import "embed"

//go:embed all:console/dist
var ConsoleFS embed.FS

//go:embed all:web/static
var WebStaticFS embed.FS
```

开发模式: Go 代理 `/console/*` 到 Vite `:3000`（通过 `--dev` flag 切换）。

## 代码规范摘要

### Go

- 包名小写单词，接口用 -er 后缀，缩写全大写 (`userID` 非 `userId`)
- 错误跨层传递用 `fmt.Errorf("context: %w", err)`
- 哨兵错误定义在 `internal/shared/apperror/`
- Repository 层返回 `apperror.ErrNotFound`（非原始 sql.ErrNoRows）
- 密码 bcrypt cost=12, API Key SHA-256, TOTP 密钥 AES-256-GCM 加密
- Huma handler 使用 RFC 9457 Problem Details 错误格式
- Web handler（Templ 页面）使用纯 Chi handler（非 Huma），返回 HTML

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
- 生产配置: `bun.WithDiscardUnknownColumns()` + 连接池超时

### console/ 前端

- **纯 SPA**: React 19 + TanStack Router v1（文件路由，beforeLoad auth guard）
- **服务端状态**: TanStack Query v5
- **表单**: TanStack Form + @tanstack/zod-form-adapter + Zod v4
- **UI 组件**: shadcn/ui（Radix UI + Tailwind CSS V4）
- **i18n**: Paraglide（inlang，编译时类型安全）
- **构建**: Vite（纯静态 SPA 产物 → go:embed）
- **代码质量**: Biome（lint + format）
- **Feature-based 架构**: `src/features/{auth,posts,categories,tags,media,users,roles,...}/`

### web/ 公共站点

- **模板**: Go Templ（`.templ` 文件，类型安全）
- **动态交互**: HTMX 2.x（局部刷新：评论、搜索、分页）
- **样式**: Tailwind CSS V4 CLI（构建时生成 → go:embed）
- **v1 硬编码内置主题**，不可自定义（v2 开放 themes/ 目录）

## 测试驱动开发 (TDD)

**铁律: 没有失败的测试，就不写生产代码。**

### Red-Green-Refactor 循环

1. **RED** — 写一个最小的失败测试，描述期望行为
2. **验证 RED** — 运行测试，确认因功能缺失而失败（非语法错误）
3. **GREEN** — 写最少的代码让测试通过，不多不少
4. **验证 GREEN** — 运行测试，确认全部通过
5. **REFACTOR** — 在保持绿灯的前提下清理代码

### DDD 测试分层

| 层 | 类型 | 工具 | 速度 |
|----|------|------|------|
| domain/ | 纯单元测试 | testify | 毫秒 |
| app/ | 单元测试（手写 mock repo） | testify | 毫秒 |
| infra/ | 集成测试 | testcontainers-go（真实 PG + Redis） | 秒 |
| delivery/ | HTTP 测试 | httptest + Huma test helper | 毫秒 |
| console/ | 组件测试 | Vitest + RTL | 秒 |
| e2e/ | 全栈测试 | Playwright | 分钟 |

### TDD 执行顺序（每个用例）

```
1. domain 层测试（纯逻辑）→ 实现实体/值对象
2. app 层测试（mock repo）→ 实现用例
3. infra 层集成测试 → 实现 bun repository
4. delivery 层测试（httptest）→ 实现 Huma handler
```

### Mock 策略

v1 使用**手写 mock 结构体**（非 mockery），保持零依赖：

```go
type mockUserRepo struct {
    findByEmailFn func(ctx context.Context, email string) (*domain.User, error)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    return m.findByEmailFn(ctx, email)
}
```

### 禁止事项

- 先写实现再补测试（测试立即通过 = 无法证明它测对了东西）
- 跳过验证失败步骤（没看到红灯 = 不知道测试是否有效）
- 保留「未经测试验证」的代码作为参考（删掉，从测试重新开始）

### 测试命令

```bash
# Go — 单包测试
go test ./internal/identity/... -v -count=1
# Go — 单个测试函数
go test ./internal/identity/app/... -run TestLoginUseCase -v
# Go — 全部（跳过需 Docker 的集成测试）
go test ./... -short -count=1

# console — 单文件
cd console && bun run vitest run src/features/auth/__tests__/LoginForm.test.tsx
# console — watch 模式
cd console && bun run vitest src/features/auth/
# console — 全部
cd console && bun run test
```

## 目录结构

```
sky-flux-cms/                          # Monorepo，单一 go.mod
├── cmd/cms/                           # Cobra CLI 单一二进制
│   ├── main.go                        # 入口
│   ├── root.go                        # rootCmd + --config flag
│   ├── serve.go                       # cms serve（HTTP 服务）
│   ├── install.go                     # cms install（CLI 安装模式）
│   └── migrate.go                     # cms migrate up|down|status
├── internal/                          # DDD 领域代码
│   ├── identity/                      # 用户、认证、RBAC
│   │   ├── domain/                    # 实体 + Repository 接口
│   │   ├── app/                       # 用例（login, register, rbac...）
│   │   ├── infra/                     # bun repo 实现
│   │   └── delivery/                  # Huma handler + DTO
│   ├── content/                       # 文章、分类、标签、评论
│   ├── media/                         # 媒体上传、缩略图
│   ├── site/                          # 站点配置、菜单、重定向
│   ├── delivery/                      # Public API、RSS/Atom、Sitemap
│   ├── platform/                      # 安装向导、审计、系统配置
│   └── shared/                        # apperror、middleware、event bus
├── config/                            # koanf 配置加载
├── migrations/                        # bun Go code migrations
├── embed.go                           # go:embed console/dist + web/static
├── console/                           # React 19 SPA（管理后台）
│   ├── src/
│   │   ├── features/                  # Feature-based 架构
│   │   │   ├── auth/
│   │   │   ├── posts/
│   │   │   ├── categories/
│   │   │   ├── tags/
│   │   │   ├── media/
│   │   │   ├── users/
│   │   │   ├── roles/
│   │   │   ├── sites/
│   │   │   ├── comments/
│   │   │   ├── menus/
│   │   │   ├── redirects/
│   │   │   ├── settings/
│   │   │   ├── api-keys/
│   │   │   └── shared/
│   │   ├── routes/                    # TanStack Router 文件路由
│   │   ├── components/                # 共享 UI 组件 (shadcn/ui)
│   │   └── paraglide/                 # i18n 编译产物
│   └── dist/                          # 构建产物（go:embed 目标）
├── web/                               # Templ + HTMX 公共站点
│   ├── templates/                     # .templ 文件
│   ├── styles/                        # Tailwind 源文件
│   └── static/                        # 构建产物（go:embed 目标）
├── docs/                              # 设计文档（旧版，仍有参考价值）
├── docs/superpowers/                  # 最新设计 spec 和 plan
│   ├── specs/                         # 设计规范
│   └── plans/                         # 实施计划
├── Procfile                           # overmind 多进程声明
├── go.mod
├── Makefile
└── docker-compose.yml
```

## 常用命令

```bash
# 开发环境
make setup               # 一键初始化 (.env + PG/Redis + 依赖安装)
make dev                  # overmind 启动所有进程（api + templ + css + console）
make build                # 构建单二进制（console build + templ generate + tailwind + go build）

# Go 后端
go run ./cmd/cms serve    # 启动 API + Web 服务（:8080）
go run ./cmd/cms install  # CLI 交互式安装
go run ./cmd/cms migrate up    # 执行迁移
templ generate            # 编译 .templ 文件
templ generate --watch    # 监听模式

# console 前端
cd console && bun run dev     # Vite dev server（:3000）
cd console && bun run build   # 构建 SPA 产物到 dist/
cd console && bun run test    # Vitest 全部测试
cd console && bun run check   # Biome lint + format

# Tailwind（web 公共站点）
tailwindcss -i web/styles/input.css -o web/static/app.css --watch

# 测试
make test                 # 运行全部测试（Go + console）
make lint                 # 代码检查（golangci-lint + Biome）

# overmind 进程管理
overmind start            # 启动 Procfile 中所有进程
overmind connect api      # 连接到 api 进程日志
```

## 设计文档

**最新文档**位于 `docs/superpowers/`，编码前应先阅读相关 spec。

旧文档位于 `docs/`，部分内容仍有参考价值，但技术栈和架构已过时。

**编码工作流**: 实现某个功能前，按此顺序执行：
1. `docs/superpowers/specs/` — 找到对应设计规范
2. `docs/superpowers/plans/` — 确认实施计划
3. **TDD 循环** — domain 测试 → app 测试 → infra 测试 → delivery 测试

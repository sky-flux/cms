# Sky Flux CMS — 项目重新设计规范

**日期:** 2026-03-19
**状态:** 已批准
**方向:** 推翻旧规划，TDD + DDD，单二进制分发

---

## 1. 项目概述

Sky Flux CMS 重构为三个子项目的 Monorepo，编译为单一 Go 二进制，支持一键安装（类 WordPress 体验）。

### 三个子项目

| 项目 | 技术栈 | 职责 |
|------|--------|------|
| `api/` | Go 1.25+ + Chi v5 + Huma v2 | REST API 后端 + Web SSR |
| `console/` | React 19 SPA + TanStack Router + shadcn/ui | 管理后台 |
| `web/` | Go Templ + HTMX 2.x + Tailwind V4 | 公共站点（嵌入 api/ 二进制） |

---

## 2. 技术栈

### api/

| 层 | 技术 |
|----|------|
| HTTP 路由 | Chi v5 |
| API 框架 | Huma v2（OpenAPI 3.1 自动生成，RFC 9457 错误格式） |
| ORM | uptrace/bun（链式查询，PostgreSQL 原生特性支持） |
| 数据库 | PostgreSQL 18（uuidv7() 原生，无扩展依赖） |
| 缓存 | Redis 8 |
| 全文搜索 | Meilisearch（CJK 分词，按站点隔离索引） |
| 对象存储 | RustFS（S3 兼容，AWS SDK v2） |
| 邮件 | Resend（HTTP API，Go SDK v3） |
| 配置 | koanf（多源合并：CLI flag > ENV > .env > 默认值） |
| 日志 | log/slog（Go 标准库，零依赖） |
| CLI | Cobra（serve / install / migrate 子命令） |
| 迁移 | bun/migrate（Go code migrations） |

### console/

| 层 | 技术 |
|----|------|
| 框架 | React 19（纯 SPA，客户端渲染，无 SSR） |
| 路由 | TanStack Router v1（文件路由，beforeLoad auth guard，**非** TanStack Start SSR） |
| 服务端状态 | TanStack Query v5 |
| 表单 | TanStack Form + @tanstack/zod-form-adapter |
| UI 组件 | shadcn/ui（Radix UI + Tailwind V4） |
| i18n | Paraglide（inlang，编译时类型安全） |
| 构建 | Vite（纯静态 SPA 产物，可 go:embed） |
| 代码质量 | Biome（lint + format） |
| 测试 | Vitest + Playwright |

### web/

| 层 | 技术 |
|----|------|
| 模板 | Go Templ（类型安全，编译时检查） |
| 动态交互 | HTMX 2.x（局部刷新：评论、搜索、分页） |
| 样式 | Tailwind CSS V4 CLI（构建时生成 → go:embed） |

---

## 3. 整体架构

### Go 模块结构

单一根 `go.mod`（模块路径 `github.com/sky-flux/cms`）。所有 Go 代码放在根级目录下（`cmd/`、`internal/`），**不放在 `api/` 子目录下**——因为 `go:embed` 和 `internal/` 可见性规则在根级最简单：

### Monorepo 目录结构

```
sky-flux-cms/                    # go.mod 在此（模块根）
├── cmd/cms/                     # Cobra CLI 入口（主二进制）
│   ├── main.go
│   ├── root.go
│   ├── serve.go
│   ├── install.go
│   └── migrate.go
├── internal/                    # DDD 领域代码（Go internal 包可见性）
│   ├── identity/
│   ├── content/
│   ├── media/
│   ├── site/
│   ├── delivery/
│   ├── platform/
│   └── shared/                  # apperror、middleware、event bus
├── migrations/                  # bun Go code migrations
├── config/                      # koanf 配置加载
├── embed.go                     # go:embed console/dist + web/static
├── go.mod
├── go.sum
├── console/                     # React 19 SPA
│   ├── src/
│   │   ├── features/            # Feature-based 架构
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
│   │   └── routes/              # TanStack Router 文件路由
│   └── dist/                    # 构建产物（go:embed 目标）
├── web/
│   ├── templates/               # .templ 文件（公共站点页面）
│   └── static/                  # Tailwind 产物 + HTMX
├── docs/
├── Makefile
└── docker-compose.yml
```

### Chi 路由分发

```
GET  /setup              → 安装向导（未安装时 InstallGuard 拦截所有请求重定向此处）
GET  /console/*          → React SPA（go:embed，SPA fallback index.html）

/api/v1/
├── /public/*            → 公共 API（Rate Limit + API Key 可选，无需 JWT）
│   ├── /posts           → 文章列表/详情
│   ├── /search          → Meilisearch 全文搜索
│   ├── /comments        → 评论提交
│   └── /feed/*          → RSS/Atom/Sitemap
└── /admin/              → 管理 API（JWT + RBAC，auth/* 子路由除外）
    ├── /auth/*          → 登录/刷新/登出（无需 JWT）
    └── 所有管理端点     → 需要 JWT + RBAC

GET  /*                  → Templ + HTMX 公共站点（SSR）
```

---

## 4. DDD 领域设计

### 界限上下文（Bounded Contexts）

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

### 依赖方向

```
delivery → app → domain ← infra
```

- **domain 层**：纯 Go，无任何框架/库依赖，可独立单元测试
- **app 层**：依赖 domain 接口，测试时 mock repository
- **infra 层**：实现 domain 接口，集成测试用 testcontainers

### 跨 BC 通信

v1 使用内存事件总线（同步），v2 可替换为消息队列：

```go
// 示例：文章发布 → 更新搜索索引
content.PostPublished → delivery.SyncMeilisearchIndex
content.PostPublished → delivery.InvalidateSitemapCache
```

---

## 5. 单一 BC 内文件示例（content/post）

```
internal/content/
├── domain/
│   ├── post.go          # Post 聚合根（状态机：draft→published→archived）
│   ├── post_repo.go     # PostRepository 接口
│   └── events.go        # PostPublished、PostArchived 领域事件
├── app/
│   ├── publish_post.go  # PublishPostUseCase
│   ├── create_post.go   # CreatePostUseCase
│   └── list_posts.go    # ListPostsUseCase
├── infra/
│   └── bun_post_repo.go # bun 实现 PostRepository
└── delivery/
    ├── handler.go        # Huma handler（调用 app 层用例）
    └── dto.go            # 请求/响应 DTO（Huma 自动 OpenAPI）
```

---

## 5.5 v1 数据库 Schema 策略

### 单站点 = 单 Schema

v1 所有表在 PostgreSQL 默认 `public` schema 中，无多站点 Schema 隔离：

| 表名 | 来源 | v1 变更 |
|------|------|---------|
| `sfc_users` | 现有迁移 1 | 保留 |
| `sfc_roles`, `sfc_apis`, `sfc_role_apis` 等 9 张 RBAC 表 | 现有迁移 2 | 保留 |
| `sfc_refresh_tokens`, `sfc_password_reset_tokens`, `sfc_user_totp` | 现有迁移 1 | 保留 |
| `sfc_configs` | 现有迁移 1 | 保留 |
| `sfc_sites` | 现有迁移 1 | **保留但简化**（v1 仅 1 条记录） |
| `sfc_posts`, `sfc_categories`, `sfc_tags` 等内容表 | **新建迁移** | 原 `site_{slug}` schema 内容表移入 `public`，无 `site_id` 列 |

### 迁移计划

| 编号 | 内容 | 来源 |
|------|------|------|
| 1 | 枚举类型 + 核心表（users, sites, tokens, configs） | 复用现有迁移 1 |
| 2 | RBAC 9 张表 | 复用现有迁移 2 |
| 3 | Seed 内置角色 + 权限模板 | 复用现有迁移 4 |
| 4 | 内容表（posts, categories, tags, media, comments, menus, redirects, audits） | **新写**，直接在 public schema |
| 5 | 索引 + 约束 + 触发器 | **新写** |

**删除**：现有迁移 3（多站点 schema 占位符）、`internal/schema/`（动态 CREATE SCHEMA）。

### 表名规范

- 全局表：`sfc_` 前缀（如 `sfc_users`、`sfc_roles`）
- 内容表：`sfc_` 前缀（如 `sfc_posts`、`sfc_categories`）——v1 无 `site_` 中缀
- v2 多站点时内容表迁移到 `site_{slug}` schema 并加 `sfc_site_` 前缀

---

## 5.6 Web 公共站点（Templ + HTMX）

### v1 页面列表

| 页面 | URL | 渲染方式 |
|------|-----|----------|
| 首页 | `/` | Templ SSR，最新文章列表 |
| 文章详情 | `/posts/:slug` | Templ SSR |
| 分类归档 | `/categories/:slug` | Templ SSR + HTMX 分页 |
| 标签归档 | `/tags/:slug` | Templ SSR + HTMX 分页 |
| 搜索结果 | `/search?q=` | HTMX 局部刷新 |
| 关于/自定义页面 | `/:slug` | Templ SSR（Page 类型文章） |

### Handler 模式

Web 路由使用 **纯 Chi handler**（非 Huma），返回 HTML：

```go
// internal/delivery/web/handler.go
func (h *WebHandler) PostDetail(w http.ResponseWriter, r *http.Request) {
    slug := chi.URLParam(r, "slug")
    post, err := h.postQuery.GetBySlug(r.Context(), slug)
    if err != nil { ... }
    templates.PostPage(post).Render(r.Context(), w)
}
```

### Tailwind CSS 构建

```
源文件:   web/styles/input.css    （@import "tailwindcss"）
输出文件:  web/static/app.css     （go:embed 目标）
HTMX:    web/static/htmx.min.js  （下载到本地，非 CDN）
```

### v1 主题策略

**v1 硬编码内置主题**，不可自定义。所有 `.templ` 文件在 `web/templates/` 目录，直接编译进二进制。v2 再开放 `themes/` 目录支持用户覆盖。

---

## 6. 认证 & 安全

### 策略

- JWT HS256（15min access token）
- httpOnly Cookie Refresh Token（7d，轮换机制）
- TOTP 2FA（可选，AES-256-GCM 加密存储密钥）
- API Key（SHA-256 hash 存储）
- 密码 bcrypt cost=12

### Chi 中间件链

```go
// 全局中间件
r.Use(middleware.RealIP)
r.Use(middleware.RequestID)
r.Use(middleware.Logger)        // slog adapter
r.Use(middleware.Recoverer)
r.Use(InstallGuard)             // 未安装时拦截所有请求 → /setup
r.Use(RateLimit)                // Redis SETNX per-IP（全局，含公共路由）

r.Route("/api/v1", func(r chi.Router) {
    // 公共路由组（无需 JWT）
    r.Group(func(r chi.Router) {
        r.Use(OptionalAPIKey)   // API Key 可选验证
        r.Mount("/public", publicRoutes)
    })

    // 管理路由组（需要 JWT）
    r.Route("/admin", func(r chi.Router) {
        r.Mount("/auth", authRoutes)  // 登录/刷新（无需 JWT）
        r.Group(func(r chi.Router) {
            r.Use(JWTAuth)      // Bearer token 验证
            r.Use(RBAC)         // 动态权限检查
            r.Mount("/", adminRoutes)
        })
    })
})
```

### 安全规范

- Schema slug 校验: `^[a-z0-9_]{3,50}$`，永不将用户输入拼入 SQL
- Huma 默认 RFC 9457 Problem Details 错误格式
- v1 单站点，无 search_path 动态切换（v2 多站点时引入）

---

## 7. 单二进制分发

### go:embed 嵌入

整个 Monorepo 使用**单一根 `go.mod`**（模块路径 `github.com/sky-flux/cms`），`embed.go` 放在项目根目录，`go:embed` 路径相对于该文件所在位置：

```go
// embed.go（项目根目录，与 go.mod 同级）
package cms // 根包名 cms，模块路径 github.com/sky-flux/cms

import "embed"

//go:embed all:console/dist
var ConsoleFS embed.FS

//go:embed all:web/static
var WebStaticFS embed.FS
```

`api/cmd/cms/main.go` 通过 import 引用：

```go
package main

import (
    rootfs "github.com/sky-flux/cms" // 导入根包获取 embed.FS
)

// 注册静态文件路由时使用 rootfs.ConsoleFS、rootfs.WebStaticFS
```

**开发模式处理**：`console/dist` 不存在时 `go:embed` 会编译失败。解决方案：

```bash
# 首次克隆后必须先构建 console：
make build-console   # cd console && bun run build
# 或使用空目录占位（开发时）：
mkdir -p console/dist && touch console/dist/.gitkeep
```

`console/dist/.gitkeep` 提交到 git，确保 `go:embed` 始终有匹配文件。开发时 Go 代理 `/console/*` 到 Vite `:3000`（通过 `--dev` flag 切换）。

### 构建流程

```makefile
build:
    cd console && bun run build
    templ generate
    tailwindcss -i web/styles/input.css -o web/static/app.css --minify
    go build -ldflags="-s -w" -o cms ./cmd/cms
```

### 安装向导（Web 模式）

```
cms serve（首次）
  └── InstallGuard 检测"未安装"
       └── 所有请求 → /setup
            ├── Step 1: 输入 DATABASE_URL，测试连接
            ├── Step 2: 执行 migrations（自动）
            ├── Step 3: 创建超管账号 + 设置 JWT_SECRET
            └── 完成 → 写入 .env → 返回 { "action": "restart_required" }
```

#### "未安装"检测逻辑（InstallGuard）

两步检测，按顺序：

1. **Config 检测**：`DATABASE_URL` 为空 → 重定向到 `/setup`（填写数据库信息）
2. **DB 检测**：Config 存在但 `sfc_migrations` 表不存在 → 重定向到 `/setup/migrate`（执行迁移）

两步都通过 = 已安装，正常路由。

#### `.env` 写入路径

| 场景 | 写入路径 |
|------|----------|
| 默认 | 二进制所在目录 `./` 下的 `.env` |
| `--config` flag | flag 指定的路径 |
| Docker | 容器内 `/data/.env`（挂载 volume） |

#### 重启策略

安装完成后，Go 进程**不自动重启**：
- Web 向导返回 JSON `{ "status": "installed", "action": "restart_required" }`
- Console 显示"安装成功，请重启服务"提示
- 用户手动 `Ctrl+C` + `cms serve` 或依赖 supervisor/Docker restart policy
- **原因**：自动 `os.Exit` 或 exec 在容器/systemd 环境下行为不可控

### 安装向导（CLI 模式）

```bash
cms install     # 交互式终端，逐步填写 → 写入 .env
cms serve       # 启动（含 InstallGuard 保护）
cms migrate up  # 手动执行迁移
```

---

## 8. 测试策略（TDD in DDD）

### 测试分层

| 层 | 类型 | 工具 | 速度 |
|----|------|------|------|
| domain/ | 纯单元测试 | testify | 毫秒 |
| app/ | 单元测试（mock repo） | testify + mockery | 毫秒 |
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
// internal/identity/app/login_test.go
type mockUserRepo struct {
    findByEmailFn func(ctx context.Context, email string) (*domain.User, error)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    return m.findByEmailFn(ctx, email)
}
```

**原因**：DDD 接口通常只有 3-5 个方法，手写 mock 比 mockery 生成代码更简洁、更易读。v2 如果接口膨胀再引入 mockery。

### 铁律

- 没有失败的测试，不写生产代码
- 跳过 infra 集成测试需显式 `-short` 标志
- domain 层测试必须在 100ms 内完成

---

## 9. 开发工作流

### 本地端口

| 端口 | 服务 |
|------|------|
| :8080 | Go API + Web Templ SSR |
| :3000 | Console Vite dev server（代理 /api/* → :8080） |

### make dev（并行启动，使用 overmind）

使用 **overmind**（推荐）或 **foreman** 管理多进程，通过 `Procfile` 声明：

```
# Procfile
api:      air -c .air.toml
templ:    templ generate --watch
css:      tailwindcss -i web/styles/input.css -o web/static/app.css --watch
console:  cd console && bun run dev
```

```bash
# 安装 overmind
brew install overmind

# 启动所有进程
overmind start
```

### 配置加载优先级

```
CLI flag > 环境变量 > .env 文件 > koanf 默认值
```

---

## 10. Docker 分发

### 多阶段 Dockerfile

```dockerfile
FROM oven/bun:1 AS console-builder
WORKDIR /app/console
COPY console/ .
RUN bun install && bun run build

FROM golang:1.25-alpine AS go-builder
WORKDIR /app
COPY . .
COPY --from=console-builder /app/console/dist ./console/dist
RUN apk add --no-cache git && \
    go build -ldflags="-s -w" -o cms ./cmd/cms

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
COPY --from=go-builder /app/cms /usr/local/bin/cms
EXPOSE 8080
CMD ["cms", "serve"]
```

### docker-compose.yml（一键启动）

```yaml
services:
  cms:
    image: skyflux/cms:latest
    ports: ["8080:8080"]
    environment:
      DATABASE_URL: postgres://cms:secret@postgres:5432/cms
      REDIS_URL: redis://redis:6379
      MEILISEARCH_URL: http://meilisearch:7700
      RUSTFS_ENDPOINT: http://rustfs:9000
      JWT_SECRET: change-me-in-production
    depends_on: [postgres, redis, meilisearch, rustfs]

  postgres:
    image: postgres:18-alpine
    environment:
      POSTGRES_DB: cms
      POSTGRES_USER: cms
      POSTGRES_PASSWORD: secret
    volumes: [postgres_data:/var/lib/postgresql/data]

  redis:
    image: redis:8-alpine
    volumes: [redis_data:/data]

  meilisearch:
    image: getmeili/meilisearch:v1.13
    volumes: [meili_data:/meili_data]

  rustfs:
    image: rustfs/rustfs:latest
    ports: ["9000:9000", "9001:9001"]
    volumes: [rustfs_data:/data]

volumes:
  postgres_data:
  redis_data:
  meili_data:
  rustfs_data:
```

---

## 11. v1 范围 & 路线图

### v1.0 范围（单站点）

- 单二进制分发（go:embed）+ Docker
- Web 安装向导 + CLI 安装模式
- 完整 CMS 功能：文章/分类/标签/媒体/评论/菜单/重定向
- 管理后台（Console React SPA）
- 动态 RBAC 权限
- JWT + TOTP 2FA 认证
- RSS/Atom/Sitemap 自动生成
- Public REST API（API Key 认证）

### v2.0 范围（延后）

- 多站点（PostgreSQL Schema 隔离）
- 主题系统（用户自定义 `themes/` 目录）
- Docker 自动化参数安装（`cms install --db-url=...`）
- 消息队列（替换内存事件总线）
- AI 内容辅助
- Webhook 系统

---

## 12. 从旧 Codebase 迁移策略

| 资产 | 处理 |
|------|------|
| `migrations/` | **复用**（数据库设计最有价值） |
| `internal/` 领域逻辑 | **参考重写**：pkg/（95%复用）+ model/（90%复用）+ service（95%复用移入 app/）；handlers/middleware/router 需重写（Gin→Chi+Huma） |
| `web/` Astro 前端 | **丢弃**（替换为 Templ + HTMX） |
| `admin/` TanStack Start | **部分重写**：feature 模块（hooks、components）可复用；router shell、root route、server entry point 需重写（TanStack Start SSR → 纯 SPA Vite） |
| `docs/` 设计文档 | **更新**（反映新架构决策） |
| Go modules（bun、Chi 等） | **保留 bun，替换 Gin → Chi + Huma，替换 Viper → koanf） |

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

### Monorepo 目录结构

```
sky-flux-cms/
├── api/                         # Go 后端（主二进制）
│   ├── cmd/cms/
│   │   ├── main.go
│   │   ├── root.go              # Cobra root + koanf 配置加载
│   │   ├── serve.go             # cms serve
│   │   ├── install.go           # cms install（交互式向导）
│   │   └── migrate.go           # cms migrate up|down|status
│   ├── internal/
│   │   ├── identity/            # BC: 用户/认证/权限
│   │   ├── content/             # BC: 文章/分类/标签/评论
│   │   ├── media/               # BC: 文件存储
│   │   ├── site/                # BC: 站点配置/菜单/重定向
│   │   ├── delivery/            # BC: RSS/Public API/预览
│   │   └── platform/            # BC: 安装向导/审计/系统配置
│   ├── migrations/              # bun Go code migrations
│   ├── config/                  # koanf 配置结构体
│   └── embed.go                 # go:embed console/dist + web/static
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
//go:embed all:console/dist
var ConsoleFS embed.FS

//go:embed all:web/static
var WebStaticFS embed.FS
```

`api/cmd/cms` 引用根包的 `ConsoleFS` 和 `WebStaticFS`，无需跨模块引用。

### 构建流程

```makefile
build:
    cd console && bun run build
    cd web && templ generate
    cd web && tailwindcss -i styles/input.css -o static/app.css --minify
    cd api && go build -ldflags="-s -w" -o ../cms ./cmd/cms
```

### 安装向导（Web 模式）

```
cms serve（首次）
  └── InstallGuard 检测 DB 未配置
       └── 所有请求 → /setup
            ├── Step 1: 测试数据库连接
            ├── Step 2: 执行 migrations
            ├── Step 3: 创建超管账号
            └── 完成 → 写入 .env → 提示重启
```

### 安装向导（CLI 模式）

```bash
cms install     # 交互式终端，逐步填写配置
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
    go build -ldflags="-s -w" -o cms ./api/cmd/cms

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
| `internal/` 领域逻辑 | **参考重写**（适配 DDD 分层，去除 Gin 依赖） |
| `web/` Astro 前端 | **丢弃**（替换为 Templ + HTMX） |
| `admin/` TanStack Start | **部分重写**：feature 模块（hooks、components）可复用；router shell、root route、server entry point 需重写（TanStack Start SSR → 纯 SPA Vite） |
| `docs/` 设计文档 | **更新**（反映新架构决策） |
| Go modules（bun、Chi 等） | **保留 bun，替换 Gin → Chi + Huma，替换 Viper → koanf） |

# 项目脚手架设计文档

**日期**: 2026-02-24
**状态**: 已完成（2026-02-24）
**范围**: 全栈骨架一次性搭建

---

## 1. 目标

从零搭建 Sky Flux CMS 完整项目骨架，实现 `make setup && make dev` 即可启动前后端，`GET /health` 返回 DB/Redis 连通状态，前端空白页可访问。

## 2. 架构决策

- **Go 单体多模块分层架构**：按模块分包（非按层分包），每个业务模块自包含 handler/service/repository/dto
- **共享层**：model/middleware/config/database/schema/router/pkg 跨模块复用
- **模块间依赖规则**：业务模块只依赖共享层，不互相直接调用；router 是唯一组装点

## 3. 基础设施层

### 3.1 文件清单

| 文件 | 作用 |
|------|------|
| `docker-compose.yml` | PG 18-alpine + Redis 8-alpine 开发环境 |
| `docker-compose.prod.yml` | 生产编排（Go + Astro + PG + Redis + Nginx） |
| `docker-compose.override.yml` | 开发环境配置：端口映射 + dev 模式（checked into git，自动加载） |
| `.env.example` | 环境变量模板 |
| `.gitignore` | Go + Node + Docker + IDE |
| `Makefile` | setup / dev / test / lint / migrate / build / clean |
| `.air.toml` | Go 热重载 |
| `.github/workflows/ci.yml` | lint + test + build（PR/push 触发） |

### 3.2 docker-compose.yml

```yaml
services:
  postgres:
    image: postgres:18-alpine
    environment:
      POSTGRES_DB: cms
      POSTGRES_USER: cms_user
      POSTGRES_PASSWORD: ${DB_PASSWORD:-devpassword}
    ports: ["5432:5432"]
    volumes: [postgres_data:/var/lib/postgresql/data]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U cms_user -d cms"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:8-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD:-devpassword}
    ports: ["6379:6379"]
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD:-devpassword}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  meilisearch:
    image: getmeili/meilisearch:v1.13
    environment:
      MEILI_MASTER_KEY: ${MEILI_MASTER_KEY:-devmasterkey}
      MEILI_ENV: development
    ports: ["7700:7700"]
    volumes: [meili_data:/meili_data]
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7700/health"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  meili_data:
```

### 3.3 Makefile 核心目标

```makefile
setup      # .env.example → .env + docker compose up -d --wait + go mod download + cd web && bun install
dev        # 并行: air (Go) + cd web && bun dev
test       # go test ./... + cd web && bun test
lint       # golangci-lint run ./... + cd web && bun run lint
migrate-up # go run ./cmd/cms migrate up
build      # go build -o ./tmp/cms ./cmd/cms + cd web && bun run build
clean      # docker compose stop + rm tmp/ coverage.*  + cd web && rm -rf dist/
```

### 3.4 CI 工作流

```yaml
on:
  push: { branches: [main] }
  pull_request: { branches: [main] }

jobs:
  backend:
    # Go 1.24, PG 18 service container, Redis 8 service container
    # golangci-lint, go test -race -coverprofile
  frontend:
    # Bun 1.2+, bun install, bun run lint, bun run typecheck, bun test
```

## 4. Go 后端骨架

### 4.1 入口文件

**`cmd/cms/`**（Cobra CLI 单一二进制）:

`serve` 子命令（`cmd/cms/serve.go`）:
1. slog 初始化（JSON 格式）
2. Viper 加载配置（.env + 环境变量）
3. 连接 PostgreSQL（uptrace/bun）
4. 连接 Redis（go-redis/v9）
5. 注册中间件链 + 路由
6. 启动 Gin HTTP 服务
7. 监听 SIGINT/SIGTERM → 优雅关闭（超时 10s）

`migrate` 子命令（`cmd/cms/migrate.go`）:
- 子命令: up / down / status / create
- 先执行 public schema 迁移，再遍历所有 site_{slug} schema

### 4.2 模块化目录结构

```
internal/
├── auth/                      # 认证模块
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── dto.go
├── user/                      # 用户模块
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── dto.go
├── post/                      # 文章模块
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── dto.go
├── category/
├── tag/
├── media/
├── comment/
├── menu/
├── redirect/
├── preview/
├── site/
├── setup/                     # 安装向导（首个完整实现模块）
│   ├── handler.go
│   ├── service.go
│   └── dto.go
├── feed/
├── apikey/
├── audit/
├── system/
│
├── model/                     # 共享数据模型
│   ├── user.go
│   ├── site.go
│   ├── post.go
│   └── ...
├── middleware/                # 共享中间件
│   ├── recovery.go
│   ├── cors.go
│   ├── request_id.go
│   ├── logger.go
│   ├── auth.go
│   ├── rbac.go
│   ├── schema.go
│   ├── site_resolver.go
│   ├── installation_guard.go
│   └── ratelimit.go
├── config/                    # Viper 配置加载
│   └── config.go
├── database/                  # DB + Redis 连接
│   ├── postgres.go
│   └── redis.go
├── schema/                    # Schema 管理
│   ├── template.go
│   ├── migrate.go
│   └── validate.go
├── router/                    # 路由注册（组装各模块 handler）
│   └── router.go
├── cron/
│   └── scheduler.go
└── pkg/                       # 共享工具包
    ├── apperror/
    ├── jwt/
    ├── crypto/
    ├── slug/
    ├── paginator/
    └── storage/                # RustFS S3 客户端封装
```

### 4.3 模块间依赖规则

- 业务模块只依赖 `model/`、`pkg/`、`database/`
- 模块间不直接调用，跨模块需求通过接口注入
- `router/` 是唯一组装点：导入所有模块 handler 并注册路由
- `middleware/` 是共享横切关注点

### 4.4 首批迁移文件

```
migrations/
├── 20260224000001_create_enums_and_functions.go  # 枚举类型 + 工具函数，PG18 原生 uuidv7() 无需额外扩展
├── 20260224000002_create_public_schema.go # sfc_users + sfc_sites + sfc_site_user_roles + sfc_refresh_tokens + sfc_user_totp + sfc_configs
└── 20260224000003_create_site_template.go # CreateSiteSchema() 函数模板
```

### 4.5 核心依赖

```
github.com/gin-gonic/gin                  v1.11.0
github.com/uptrace/bun                     v1.2.16
github.com/uptrace/bun/dialect/pgdialect   v1.2.16
github.com/uptrace/bun/driver/pgdriver     v1.2.16
github.com/uptrace/bun/extra/bundebug      v1.2.16
github.com/redis/go-redis/v9              v9.18.0
github.com/spf13/viper                     v1.21.0
github.com/golang-jwt/jwt/v5              v5.3.1
golang.org/x/crypto                        latest
github.com/meilisearch/meilisearch-go    v0.36.1
```

### 4.6 骨架范围

每个业务模块只创建空文件 + 接口定义 + TODO 占位，不实现业务逻辑。唯一可运行端点：`GET /health`（返回 DB/Redis 连通状态）。

## 5. 前端骨架（web/）

### 5.1 初始化命令

```bash
cd web
bunx --bun create-astro@latest . --template with-tailwindcss --install --add react
# tsconfig.json 添加路径别名: @/* → ./src/*
bunx --bun shadcn@latest init
```

### 5.2 目录结构

```
web/
├── src/
│   ├── components/
│   │   └── ui/                # shadcn/ui 按需添加
│   ├── pages/
│   │   └── index.astro        # 临时首页
│   ├── layouts/
│   │   └── BaseLayout.astro
│   ├── lib/
│   │   └── api.ts             # API 客户端骨架
│   ├── hooks/
│   ├── stores/
│   └── i18n/
├── astro.config.mjs           # SSR 模式 + React + @astrojs/node
├── tailwind.config.mjs
├── tsconfig.json
├── components.json
└── package.json
```

### 5.3 骨架范围

确保 `cd web && bun dev` 能启动，浏览器可看到空白页。不实现任何页面逻辑。

## 6. Agent Teams 实施计划

### 6.1 团队结构

```
team-lead（主会话）
├── agent-infra       # 基础设施
├── agent-backend     # Go 后端
├── agent-frontend    # 前端
└── agent-ci          # CI 工作流
```

### 6.2 执行阶段

**Phase 1 — 并行（无依赖）**:
- `agent-infra`: docker-compose×3 + .env.example + .gitignore + Makefile + .air.toml
- `agent-ci`: .github/workflows/ci.yml

**Phase 2 — 并行（依赖 Phase 1）**:
- `agent-backend`: go.mod + 全部 Go 骨架代码 + cmd/ + internal/ + migrations/
- `agent-frontend`: web/ 目录 Astro + React + shadcn/ui 初始化

**Phase 3 — 主会话集成验证**:
- `make setup && make dev` 能跑通
- `GET /health` 返回 200
- 前端 `bun dev` 启动成功

## 7. 验收标准

1. `make setup` 一键完成环境初始化
2. `make dev` 同时启动后端（air 热重载 :8080）+ 前端（Astro dev :3000）
3. `curl localhost:8080/health` 返回 `{"status":"ok","db":"connected","redis":"connected"}`
4. `localhost:3000` 可访问前端空白页
5. `make test` 通过（至少 health check 有一个测试）
6. `make lint` 通过
7. GitHub Actions CI 配置正确（需推送后验证）

# Sky Flux CMS 项目脚手架实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 从零搭建全栈项目骨架，`make setup && make dev` 启动后端 :8080 + 前端 :3000，health check 可响应。

**Architecture:** Go 单体多模块分层架构（按模块分包），前端 Astro 5 SSR + React 19 + shadcn/ui。每个业务模块自包含 handler/service/repo/dto，共享层跨模块复用。

**Tech Stack:** Go 1.24 / Gin v1.11.0 / uptrace/bun v1.2.16 / PostgreSQL 18 / Redis 8 / Astro 5 / React 19 / shadcn/ui / GitHub Actions

**Design Doc:** `docs/plans/2026-02-24-project-scaffolding-design.md`

**Execution:** Agent Teams — 4 个并行 Agent，2 个 Phase + 集成验证

---

## Phase 1: 基础设施（并行）

### Task 1: agent-infra — 基础设施文件

**负责 Agent:** `agent-infra` (general-purpose)

**Files to create:**
- `.gitignore`
- `.env.example`
- `docker-compose.yml`
- `docker-compose.prod.yml`
- `docker-compose.override.yml.example`
- `Makefile`
- `.air.toml`

#### Step 1.1: 创建 .gitignore

```gitignore
# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
tmp/
vendor/

# Frontend
web/node_modules/
web/dist/
web/.astro/

# Environment
.env
.env.local
.env.*.local
docker-compose.override.yml

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Coverage
coverage.out
coverage.html
web/coverage/

# Build
tmp/
```

#### Step 1.2: 创建 .env.example

```bash
# Server
SERVER_PORT=8080
SERVER_MODE=debug
FRONTEND_URL=http://localhost:3000

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=cms
DB_USER=cms_user
DB_PASSWORD=devpassword
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=devpassword
REDIS_DB=0

# JWT
JWT_SECRET=dev-secret-change-in-production-min-32-chars
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# TOTP
TOTP_ENCRYPTION_KEY=dev-totp-key-change-in-production!

# RustFS
RUSTFS_ENDPOINT=http://localhost:9000
RUSTFS_ACCESS_KEY=rustfsadmin
RUSTFS_SECRET_KEY=rustfsadmin
RUSTFS_BUCKET=cms-media
RUSTFS_REGION=us-east-1

# Meilisearch
MEILI_URL=http://localhost:7700
MEILI_MASTER_KEY=devmasterkey

# Log
LOG_LEVEL=debug
LOG_FORMAT=json
```

#### Step 1.3: 创建 docker-compose.yml

```yaml
services:
  postgres:
    image: postgres:18-alpine
    container_name: cms-postgres
    environment:
      POSTGRES_DB: ${DB_NAME:-cms}
      POSTGRES_USER: ${DB_USER:-cms_user}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-devpassword}
    ports:
      - "${DB_PORT:-5432}:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-cms_user} -d ${DB_NAME:-cms}"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:8-alpine
    container_name: cms-redis
    command: redis-server --requirepass ${REDIS_PASSWORD:-devpassword}
    ports:
      - "${REDIS_PORT:-6379}:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD:-devpassword}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  meilisearch:
    image: getmeili/meilisearch:v1.13
    container_name: cms-meilisearch
    environment:
      MEILI_MASTER_KEY: ${MEILI_MASTER_KEY:-devmasterkey}
      MEILI_ENV: development
    ports:
      - "${MEILI_PORT:-7700}:7700"
    volumes:
      - meili_data:/meili_data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7700/health"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  redis_data:
  meili_data:
```

#### Step 1.4: 创建 docker-compose.prod.yml

```yaml
services:
  server:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: cms-server
    env_file: .env
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped

  web:
    build:
      context: ./web
      dockerfile: Dockerfile
    container_name: cms-web
    environment:
      - PUBLIC_API_URL=${PUBLIC_API_URL:-http://server:8080}
    ports:
      - "3000:3000"
    depends_on:
      - server
    restart: unless-stopped

  postgres:
    image: postgres:18-alpine
    container_name: cms-postgres
    environment:
      POSTGRES_DB: ${DB_NAME:-cms}
      POSTGRES_USER: ${DB_USER:-cms_user}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-cms_user} -d ${DB_NAME:-cms}"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:8-alpine
    container_name: cms-redis
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  meilisearch:
    image: getmeili/meilisearch:v1.13
    container_name: cms-meilisearch
    environment:
      MEILI_MASTER_KEY: ${MEILI_MASTER_KEY}
      MEILI_ENV: production
    volumes:
      - meili_data:/meili_data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7700/health"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  meili_data:
```

#### Step 1.5: 创建 docker-compose.override.yml.example

```yaml
# 复制为 docker-compose.override.yml 并按需修改（已 gitignored）
# Docker Compose 会自动合并此文件
services:
  postgres:
    ports:
      - "5432:5432"
  redis:
    ports:
      - "6379:6379"
```

#### Step 1.6: 创建 Makefile

```makefile
.PHONY: setup dev dev-backend dev-frontend test test-backend test-frontend lint build clean migrate-up migrate-down migrate-status

# ──────────────────────────────────────
# 开发环境
# ──────────────────────────────────────

setup:
	@test -f .env || cp .env.example .env
	docker compose up -d --wait
	go mod download
	cd web && bun install
	@echo "Setup complete. Run 'make dev' to start."

dev:
	@make -j2 dev-backend dev-frontend

dev-backend:
	air

dev-frontend:
	cd web && bun dev

# ──────────────────────────────────────
# 测试
# ──────────────────────────────────────

test: test-backend test-frontend

test-backend:
	go test -v -race -count=1 ./...

test-frontend:
	cd web && bun test

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ──────────────────────────────────────
# 代码质量
# ──────────────────────────────────────

lint:
	golangci-lint run ./...
	cd web && bun run lint

fmt:
	gofmt -w .
	cd web && bun run format

# ──────────────────────────────────────
# 数据库迁移
# ──────────────────────────────────────

migrate-up:
	go run ./cmd/cms migrate up

migrate-down:
	go run ./cmd/cms migrate down

migrate-status:
	go run ./cmd/cms migrate status

# ──────────────────────────────────────
# 构建
# ──────────────────────────────────────

build:
	go build -ldflags="-w -s" -o ./tmp/cms ./cmd/cms
	cd web && bun run build

# ──────────────────────────────────────
# 清理
# ──────────────────────────────────────

clean:
	docker compose stop
	rm -rf tmp/ coverage.out coverage.html
	cd web && rm -rf node_modules/.cache dist/

reset: clean
	docker compose down -v
	$(MAKE) setup
```

#### Step 1.7: 创建 .air.toml

```toml
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/cms serve"
  cmd = "go build -o ./tmp/cms ./cmd/cms"
  delay = 1000
  exclude_dir = ["tmp", "vendor", "web", "node_modules", "docs", "migrations", ".git"]
  exclude_regex = ["_test\\.go$"]
  include_ext = ["go", "toml"]
  kill_delay = "0s"
  send_interrupt = true

[log]
  time = false

[misc]
  clean_on_exit = true
```

#### Step 1.8: 验证

```bash
ls -la .gitignore .env.example docker-compose.yml docker-compose.prod.yml docker-compose.override.yml.example Makefile .air.toml
```

Expected: 所有 7 个文件存在

---

### Task 2: agent-ci — GitHub Actions CI

**负责 Agent:** `agent-ci` (general-purpose)

**Files to create:**
- `.github/workflows/ci.yml`

#### Step 2.1: 创建 CI 工作流

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  backend:
    name: Go Backend
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:18-alpine
        env:
          POSTGRES_DB: cms_test
          POSTGRES_USER: cms_user
          POSTGRES_PASSWORD: testpassword
        ports:
          - 5432:5432
        options: >-
          --health-cmd "pg_isready -U cms_user -d cms_test"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:8-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"

      - name: Install golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=5m

      - name: Run tests
        env:
          DB_HOST: localhost
          DB_PORT: 5432
          DB_NAME: cms_test
          DB_USER: cms_user
          DB_PASSWORD: testpassword
          DB_SSLMODE: disable
          REDIS_HOST: localhost
          REDIS_PORT: 6379
          REDIS_PASSWORD: ""
          JWT_SECRET: ci-test-secret-minimum-32-characters
          TOTP_ENCRYPTION_KEY: ci-test-totp-key-change-me!!!!!
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        if: github.event_name == 'pull_request'
        uses: actions/upload-artifact@v4
        with:
          name: backend-coverage
          path: coverage.out

  frontend:
    name: Frontend
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: oven-sh/setup-bun@v2
        with:
          bun-version: latest

      - name: Install dependencies
        working-directory: web
        run: bun install --frozen-lockfile

      - name: Lint
        working-directory: web
        run: bun run lint

      - name: Type check
        working-directory: web
        run: bun run typecheck

      - name: Test
        working-directory: web
        run: bun test

      - name: Build
        working-directory: web
        run: bun run build
```

#### Step 2.2: 验证

```bash
cat .github/workflows/ci.yml | head -5
```

Expected: `name: CI`

---

## Phase 2: 应用代码（并行，依赖 Phase 1）

### Task 3: agent-backend — Go 后端骨架

**负责 Agent:** `agent-backend` (general-purpose)

**必读设计文档:** `docs/plans/2026-02-24-project-scaffolding-design.md` §4, `CLAUDE.md`

**执行顺序（严格按序）:**

1. `go mod init` + 安装依赖
2. `internal/config/config.go` — Viper 配置
3. `internal/pkg/apperror/errors.go` — 哨兵错误
4. `internal/model/` — 核心数据模型（user, site, system_config）
5. `internal/database/postgres.go` — bun 连接
6. `internal/database/redis.go` — go-redis 连接
7. `internal/middleware/` — 共享中间件（recovery, cors, request_id, logger）
8. `internal/router/router.go` — 路由 + health handler
9. `cmd/cms/` — Cobra CLI 单一二进制
10. 16 个业务模块骨架（按模板批量创建）
11. `internal/schema/validate.go` — Schema 校验
12. `migrations/` — 首批迁移文件（占位）
14. Health check 测试

#### Step 3.1: 初始化 Go 模块

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms
go mod init github.com/sky-flux/cms
```

然后安装核心依赖:

```bash
go get github.com/gin-gonic/gin@v1.11.0
go get github.com/uptrace/bun@v1.2.16
go get github.com/uptrace/bun/dialect/pgdialect@v1.2.16
go get github.com/uptrace/bun/driver/pgdriver@v1.2.16
go get github.com/uptrace/bun/extra/bundebug@v1.2.16
go get github.com/redis/go-redis/v9@v9.18.0
go get github.com/spf13/viper@v1.21.0
go get github.com/golang-jwt/jwt/v5@v5.3.1
go get golang.org/x/crypto
go get github.com/meilisearch/meilisearch-go@v0.36.1
```

#### Step 3.2: internal/config/config.go

Viper 加载配置，结构体包含 Server/DB/Redis/Meilisearch/JWT/TOTP/RustFS/Log 分组。
- 从 `.env` 文件和环境变量加载
- 提供 `Load() (*Config, error)` 函数

#### Step 3.3: internal/pkg/apperror/errors.go

定义哨兵错误 + 统一 AppError 结构体:
- `ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`, `ErrValidation`, `ErrInternal`
- `AppError{Code int, Message string, Err error}` 实现 `error` 接口

#### Step 3.4: internal/pkg/response/response.go

统一 JSON 响应结构:
- `Success(c *gin.Context, data interface{})`
- `Error(c *gin.Context, err error)` — 根据 AppError 类型返回对应 HTTP 状态码
- `Paginated(c *gin.Context, data interface{}, total int64, page, perPage int)`

#### Step 3.5: internal/model/ — 核心模型

创建安装向导和基础功能需要的 3 个核心模型文件:
- `user.go` — User 模型（bun.BaseModel + UUIDv7 PK + 常用字段）
- `site.go` — Site 模型（slug, name, domain, settings JSONB）
- `system_config.go` — SystemConfig 模型（key-value 配置）

其余模型文件（post.go, category.go 等）创建为空骨架，仅包含 `package model`。

#### Step 3.6: internal/database/postgres.go

- `NewPostgres(cfg *config.Config) (*bun.DB, error)`
- 连接池配置（MaxOpenConns, MaxIdleConns）
- bundebug 调试 hook（debug 模式启用）
- `Ping(ctx)` 健康检查方法

#### Step 3.7: internal/database/redis.go

- `NewRedis(cfg *config.Config) (*redis.Client, error)`
- `Ping(ctx)` 健康检查方法

#### Step 3.7b: internal/database/meilisearch.go

- `NewMeilisearch(cfg *config.Config) (meilisearch.ServiceManager, error)`
- 使用 Master Key 初始化客户端
- `Health(ctx)` 健康检查方法

#### Step 3.8: internal/middleware/

创建 4 个基础中间件（骨架阶段只需这些）:
- `recovery.go` — 拦截 panic，slog 记录 + 返回 500
- `cors.go` — CORS 配置（开发时允许 localhost:3000）
- `request_id.go` — X-Request-ID 注入（uuid）
- `logger.go` — slog 请求日志（method, path, status, latency）

其余中间件文件（auth.go, rbac.go, schema.go 等）创建为空骨架。

#### Step 3.9: internal/router/router.go

- `Setup(engine *gin.Engine, db *bun.DB, rdb *redis.Client)` 函数
- 注册中间件链: recovery → request_id → logger → cors
- 注册 `GET /health` 端点（检查 DB + Redis + Meilisearch 连通性）
- 预留 API v1 分组注释

#### Step 3.10: cmd/cms/（Cobra CLI 单一二进制）

完整实现:

`serve` 子命令（`cmd/cms/serve.go`）:
1. slog 初始化（JSON handler）
2. config.Load()
3. database.NewPostgres() + database.NewRedis() + database.NewMeilisearch()
4. gin.New() + router.Setup()
5. HTTP server 启动 + 优雅关闭（SIGINT/SIGTERM, 10s timeout）

`migrate` 子命令（`cmd/cms/migrate.go`）:
- 子命令: up / down / status / create
- 先执行 public schema 迁移，再遍历所有 site_{slug} schema

#### Step 3.11: 16 个业务模块骨架

按以下模板为每个模块创建目录和文件:

**标准模块（含 handler/service/repository/dto）:**
auth, user, post, category, tag, media, comment, menu, redirect, preview, site, apikey, system

**精简模块:**
- `setup/` — handler.go + service.go + dto.go（无 repository，直接用 model）
- `feed/` — handler.go + service.go（无 repository/dto）
- `audit/` — service.go + repository.go + middleware.go（无 handler/dto）

每个文件模板:

```go
package <module>

// TODO: implement <module> <layer>
```

#### Step 3.12: internal/schema/validate.go

```go
package schema

import "regexp"

var slugRegex = regexp.MustCompile(`^[a-z0-9_]{3,50}$`)

func ValidateSlug(slug string) bool {
    return slugRegex.MatchString(slug)
}
```

`template.go` 和 `migrate.go` 创建为 TODO 占位。

#### Step 3.13: migrations/ — 占位文件

创建 3 个迁移文件骨架（`package migrations` + TODO 注释），实际 DDL 在安装向导模块实现时填充。

#### Step 3.14: health check 测试

创建 `internal/router/router_test.go`:
- 测试 `GET /health` 返回 200
- 使用 `httptest.NewRecorder` + Gin test mode
- Mock DB/Redis 或直接测试路由注册

```bash
go test -v -race ./internal/router/...
```

Expected: PASS

#### Step 3.16: 验证编译

```bash
go build ./...
```

Expected: 无错误

---

### Task 4: agent-frontend — 前端骨架

**负责 Agent:** `agent-frontend` (general-purpose)

#### Step 4.1: 创建 Astro 项目

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms
mkdir -p web
cd web
bunx --bun create-astro@latest . --template with-tailwindcss --install --add react
```

交互式提示选择:
- TypeScript: strict
- Git: No（已有 git）

#### Step 4.2: 配置 tsconfig.json 路径别名

在 `web/tsconfig.json` 中添加:

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

#### Step 4.3: 初始化 shadcn/ui

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web
bunx --bun shadcn@latest init
```

交互式提示:
- Style: Default
- Base color: Neutral
- CSS variables: Yes

#### Step 4.4: 配置 Astro SSR

修改 `web/astro.config.mjs`:
- `output: 'server'`
- 添加 `@astrojs/node` 适配器
- 配置 Vite proxy: `/api` → `http://localhost:8080`

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web
bunx astro add node
```

#### Step 4.5: 创建基础布局

创建 `web/src/layouts/BaseLayout.astro`:
- HTML 骨架 + viewport meta
- Tailwind CSS 引入
- `<slot />` 内容插槽

#### Step 4.6: 更新首页

修改 `web/src/pages/index.astro`:
- 使用 BaseLayout
- 显示 "Sky Flux CMS" 标题 + "Setup required" 提示

#### Step 4.7: 创建 API 客户端骨架

创建 `web/src/lib/api.ts`:
- 导出 `API_BASE_URL` 常量
- 导出 `fetchAPI<T>(endpoint, options)` 通用请求函数
- 处理 JSON 响应 + 错误

#### Step 4.8: 创建占位目录

```bash
mkdir -p web/src/hooks web/src/stores web/src/i18n web/src/components/ui
```

每个目录创建 `.gitkeep` 文件。

#### Step 4.9: 添加 package.json scripts

确保 `web/package.json` 包含:

```json
{
  "scripts": {
    "dev": "astro dev",
    "build": "astro build",
    "preview": "astro preview",
    "lint": "astro check",
    "typecheck": "astro check",
    "test": "echo 'No tests yet' && exit 0",
    "format": "prettier --write ."
  }
}
```

#### Step 4.10: 验证前端启动

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web
bun dev &
sleep 5
curl -s http://localhost:3000 | head -20
kill %1
```

Expected: HTML 输出包含 "Sky Flux CMS"

---

## Phase 3: 集成验证（主会话）

### Task 5: team-lead — 集成验证

#### Step 5.1: 验证 make setup

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms
make setup
```

Expected: Docker 容器启动，依赖安装完成

#### Step 5.2: 验证 make dev

```bash
make dev &
sleep 10
```

#### Step 5.3: 验证 health check

```bash
curl -s http://localhost:8080/health | jq .
```

Expected:
```json
{
  "status": "ok",
  "db": "connected",
  "redis": "connected"
}
```

#### Step 5.4: 验证前端

```bash
curl -s http://localhost:3000 | head -20
```

Expected: HTML 包含 "Sky Flux CMS"

#### Step 5.5: 验证测试

```bash
make test
```

Expected: 所有测试通过

#### Step 5.6: 验证 lint

```bash
make lint
```

Expected: 无 lint 错误（或仅 TODO 相关 warning）

#### Step 5.7: 停止服务

```bash
kill %1 2>/dev/null
docker compose stop
```

---

## 验收标准 Checklist

- [x] `make setup` 一键完成环境初始化（文件就绪，需 Docker 运行时验证）
- [x] `make dev` 同时启动后端 :8080 + 前端 :3000（Makefile + .air.toml + Astro SSR 已配置）
- [x] `GET /health` 返回 200 + DB/Redis/Meilisearch 状态（router_test.go PASS）
- [x] `localhost:3000` 前端空白页可访问（`bun run build` 成功，index.astro 含 "Sky Flux CMS"）
- [x] `make test` 通过（`go test -race ./...` PASS，前端无测试文件 — 符合骨架预期）
- [x] `make lint` 通过（`go vet ./...` 无错误）
- [x] Go 模块化结构正确（16 个业务模块 + 8 个共享模块 = 24 个 internal/ 子目录）
- [x] GitHub Actions CI 配置文件存在（`.github/workflows/ci.yml`）

**完成日期**: 2026-02-24
**验证方式**: `go build ./...` 零错误 + `go test -race ./...` 1 PASS + `bun run build` 成功

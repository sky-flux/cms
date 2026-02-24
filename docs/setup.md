# 开发环境搭建指南

> 从零开始搭建 Sky Flux CMS 本地开发环境，目标：**5 分钟内跑通前后端**。

---

## 前置要求

| 工具 | 版本 | 安装 |
|------|------|------|
| Go | 1.25+ | https://go.dev/dl/ |
| bun | 1.2+ | `curl -fsSL https://bun.sh/install \| bash` |
| Docker | 25+ | https://docs.docker.com/get-docker/ |
| Docker Compose | 2.24+ | Docker Desktop 自带 |
| Make | 3.81+ | macOS 自带；Linux: `sudo apt install make` |

> **包管理器说明**：前端统一使用 **bun**，禁止使用 npm / pnpm / yarn。

---

## 快速开始

```bash
# 1. 克隆项目
git clone https://github.com/sky-flux/cms.git
cd cms

# 2. 一键初始化（生成 .env + 启动 PG/Redis + 安装依赖）
make setup

# 3. 启动全部服务（后端 + 前端）
make dev

# 4. 打开浏览器
#    安装向导：http://localhost:3000/setup
#    完成后进入管理后台：http://localhost:3000/admin
```

首次运行会进入 **Web 安装向导**，按提示填写站点名称、管理员邮箱和密码即可。

---

## Docker Compose 服务

开发环境需 PostgreSQL、Redis、Meilisearch 和 RustFS 四个基础设施服务，Go 后端和 Astro 前端在宿主机直接运行（支持热重载）。

### docker-compose.yml

```yaml
# docker-compose.yml — 开发环境基础设施
services:
  postgres:
    image: postgres:18-alpine
    container_name: cms-postgres
    restart: unless-stopped
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: cms_user
      POSTGRES_PASSWORD: ${DB_PASSWORD:-changeme}
      POSTGRES_DB: cms
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U cms_user -d cms"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:8-alpine
    container_name: cms-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    command: redis-server --requirepass ${REDIS_PASSWORD:-changeme} --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - redisdata:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD:-changeme}", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  meilisearch:
    image: getmeili/meilisearch:v1.13
    container_name: cms-meilisearch
    restart: unless-stopped
    ports:
      - "7700:7700"
    environment:
      MEILI_MASTER_KEY: ${MEILI_MASTER_KEY:-devmasterkey}
      MEILI_ENV: development
    volumes:
      - meilidata:/meili_data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7700/health"]
      interval: 5s
      timeout: 3s
      retries: 5

  rustfs:
    image: rustfs/rustfs:latest
    container_name: cms-rustfs
    restart: unless-stopped
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      RUSTFS_ACCESS_KEY: ${RUSTFS_ACCESS_KEY:-rustfsadmin}
      RUSTFS_SECRET_KEY: ${RUSTFS_SECRET_KEY:-rustfsadmin}
    command: /data
    volumes:
      - rustfsdata:/data
    healthcheck:
      test: ["CMD-SHELL", "curl -sf http://localhost:9000/ || exit 1"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  pgdata:
  redisdata:
  meilidata:
  rustfsdata:
```

### 常用命令

```bash
# 启动基础设施
docker compose up -d

# 查看服务状态（确认 healthy）
docker compose ps

# 查看日志
docker compose logs -f postgres
docker compose logs -f redis

# 停止服务（保留数据）
docker compose stop

# 停止并删除数据卷（完全重置）
docker compose down -v
```

---

## Makefile 命令参考

### 初始化 & 环境

```makefile
# 一键初始化开发环境：生成 .env、启动 Docker、安装依赖
setup:
	@test -f .env || cp .env.example .env
	docker compose up -d --wait
	go mod download
	cd web && bun install
	@echo "Setup complete. Run 'make dev' to start."
```

### 开发服务器

```makefile
# 并行启动前后端
dev:
	@make -j2 dev-backend dev-frontend

# 后端开发服务器（带热重载）
dev-backend:
	air

# 前端开发服务器
dev-frontend:
	cd web && bun dev
```

### 数据库迁移

```makefile
# 执行所有待迁移
migrate-up:
	go run ./cmd/cms migrate up

# 回滚最近一次迁移
migrate-down:
	go run ./cmd/cms migrate down

# 查看迁移状态
migrate-status:
	go run ./cmd/cms migrate status
```

### 测试

```makefile
# 运行全部后端测试
test:
	go test ./... -v -count=1

# 运行后端测试（带覆盖率）
test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告：coverage.html"

# 运行前端测试
test-frontend:
	cd web && bun test

# 运行 E2E 测试（Playwright）
test-e2e:
	cd web && bun run test:e2e

# 运行全部测试
test-all: test test-frontend
```

### 代码质量

```makefile
# 后端 lint
lint:
	golangci-lint run ./...

# 前端 lint + type check
lint-frontend:
	cd web && bun run lint
	cd web && bun run typecheck

# 格式化代码
fmt:
	gofmt -w .
	cd web && bun run format
```

### 清理

```makefile
# 停止 Docker + 清理构建产物
clean:
	docker compose stop
	rm -rf tmp/ coverage.out coverage.html
	cd web && rm -rf node_modules/.cache dist/

# 完全重置（删除数据卷 + 重新初始化）
reset: clean
	docker compose down -v
	$(MAKE) setup
```

---

## 项目目录结构

```
sky-flux-cms/
├── cmd/
│   └── cms/                    # Cobra CLI 单一二进制（serve / migrate 子命令）
├── internal/
│   ├── auth/                   # 认证模块（handler/service/repo/dto 自包含）
│   ├── user/ post/ ...         # 各业务模块（同上模式）
│   ├── model/                  # 共享数据模型
│   ├── middleware/             # 共享中间件（auth, schema, rate-limit...）
│   ├── config/                 # Viper 配置加载
│   ├── database/               # DB + Redis 连接
│   ├── router/                 # 路由注册（唯一组装点）
│   ├── schema/                 # 站点 Schema 管理
│   └── pkg/                    # 共享工具包 (apperror/jwt/crypto/slug)
├── migrations/                 # bun 迁移文件（Go 代码）
├── web/                        # Astro 5 SSR + React 19 前端
│   ├── src/
│   │   ├── pages/              # Astro 页面（含 /setup 安装向导）
│   │   ├── components/         # React 组件
│   │   ├── layouts/            # 页面布局
│   │   ├── stores/             # Zustand 状态管理
│   │   └── lib/                # 工具函数 + API 客户端
│   ├── astro.config.mjs
│   ├── bun.lock
│   └── package.json
├── docs/                       # 设计文档
├── .air.toml                   # air 热重载配置
├── go.mod
├── go.sum
├── Makefile
├── docker-compose.yml          # 开发环境基础设施
├── docker-compose.prod.yml     # 生产环境编排
├── .env.example                # 环境变量模板
└── .gitignore
```

---

## 环境变量速查

开发环境必填项（`.env`）：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_PASSWORD` | `changeme` | PostgreSQL 密码 |
| `DATABASE_URL` | `postgres://cms_user:changeme@localhost:5432/cms?sslmode=disable` | 数据库连接串 |
| `REDIS_URL` | `redis://:changeme@localhost:6379/0` | Redis 连接串 |
| `JWT_SECRET` | — | JWT 签名密钥（`make setup` 自动生成） |
| `TOTP_ENCRYPTION_KEY` | — | 2FA 加密密钥（`make setup` 自动生成） |
| `APP_PORT` | `8080` | 后端端口 |
| `APP_ENV` | `development` | 运行环境 |
| `PUBLIC_API_URL` | `http://localhost:8080` | 前端访问的后端 API 地址 |

完整环境变量清单见 [deployment.md §3](deployment.md#3-环境变量清单)。

---

## 开发工作流

### 日常开发

```bash
# 每天开始
make dev                    # 启动前后端（Docker 已在后台运行）

# 写代码...（air 自动重载后端，bun dev 自动重载前端）

# 运行测试
make test                   # 后端测试
make test-frontend          # 前端测试

# 提交前
make lint                   # 代码检查
make fmt                    # 格式化
```

### 数据库变更

```bash
# 1. 手动创建新迁移 Go 文件（bun 迁移是 Go 代码，不是 SQL 文件）
#    在 migrations/ 目录下创建 YYYYMMDDHHMMSS_description.go
#    使用 init() + Migrations.MustRegister(up, down) 模式

# 2. 执行迁移
make migrate-up

# 3. 验证
make migrate-status
```

### 多站点开发

系统采用 PostgreSQL Schema 隔离，每个站点一个独立 schema（`site_{slug}`）：

```bash
# 通过安装向导创建第一个站点
open http://localhost:3000/setup

# 通过 API 创建更多站点（需 Super 角色 Token）
curl -X POST http://localhost:8080/api/v1/sites \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "技术博客", "slug": "tech_blog"}'

# 切换站点上下文（API 请求头）
curl http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Site-Slug: tech_blog"
```

### 热重载配置

后端使用 [air](https://github.com/air-verse/air) 实现热重载：

```toml
# .air.toml
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/cms serve"
  cmd = "go build -o ./tmp/cms ./cmd/cms"
  delay = 1000
  exclude_dir = ["tmp", "vendor", "node_modules"]
  exclude_regex = ["_test.go"]
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"

[log]
  time = false

[misc]
  clean_on_exit = true
```

---

## 常见问题

### PostgreSQL 连接失败

```bash
# 检查容器状态
docker compose ps

# 手动测试连接
docker compose exec postgres psql -U cms_user -d cms -c "SELECT 1"

# 查看 PG 日志
docker compose logs postgres
```

### 端口被占用

```bash
# 检查端口占用
lsof -i :5432    # PostgreSQL
lsof -i :6379    # Redis
lsof -i :8080    # 后端
lsof -i :3000    # 前端

# 修改端口：编辑 .env 中的 APP_PORT，docker-compose.yml 中的 ports 映射
```

### Meilisearch 搜索服务

如果 Meilisearch 无法连接：

```bash
# 检查服务状态
docker compose ps meilisearch
curl http://localhost:7700/health

# 查看日志
docker compose logs meilisearch

# 重建服务
docker compose up -d meilisearch --force-recreate
```

### 完全重置开发环境

```bash
make reset    # 删除所有数据卷 + 重新初始化
```

---

## 深入阅读

| 文档 | 内容 |
|------|------|
| [deployment.md](deployment.md) | 生产部署、Nginx 配置、SSL、备份策略 |
| [database.md](database.md) | 完整数据库设计、Schema 隔离架构、DDL |
| [api.md](api.md) | 完整 API 规格（OpenAPI 3.1） |
| [architecture.md](architecture.md) | 系统架构、中间件链、目录结构 |
| [standard.md](standard.md) | 编码规范、Schema 隔离模式、测试模式 |
| [security.md](security.md) | 安全策略、认证流程、2FA |
| [testing.md](testing.md) | 测试策略、用例清单 |

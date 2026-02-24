<p align="center">
  <h1 align="center">Sky Flux CMS</h1>
  <p align="center">面向中小型团队及独立开发者的现代化、高性能 Headless 内容管理系统。</p>
</p>

<p align="center">
  <a href="./README.md">English</a> | <a href="./README.zh-TW.md">繁體中文</a> | <a href="./README.de.md">Deutsch</a>
</p>

## 核心特性

- **多站点** — PostgreSQL Schema 隔离（`site_{slug}`），完全数据分离
- **Headless API** — RESTful API-First 设计，任意前端消费
- **动态 RBAC** — 内置 4 个角色（super/admin/editor/viewer），支持自定义角色与权限
- **丰富内容** — 文章、分类、标签、媒体管理，草稿/发布/定时发布工作流
- **现代管理后台** — Astro 5 SSR + React 19 + shadcn/ui
- **全文搜索** — Meilisearch，内置 CJK 分词
- **Web 安装向导** — 浏览器中即可完成首次配置
- **双因素认证** — TOTP 2FA 支持
- **评论系统** — 内置评论与审核功能
- **SEO 友好** — RSS/Atom 订阅、Sitemap、URL 重定向、草稿预览

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.25+、Gin、uptrace/bun ORM |
| 数据库 | PostgreSQL 18、Redis 8 |
| 搜索 | Meilisearch |
| 对象存储 | RustFS（S3 兼容） |
| 邮件 | Resend |
| 前端 | Astro 5 SSR、React 19、shadcn/ui、TanStack Query v5、Zustand |
| 认证 | JWT + Refresh Token + TOTP 2FA |

## 快速开始

### 环境要求

- Go 1.25+
- Docker 27+ & Docker Compose 2+
- [Bun](https://bun.sh) 1.2+

### 安装

```bash
git clone https://github.com/sky-flux/cms.git
cd cms

# 初始化环境、启动服务、安装依赖
make setup

# 启动开发环境（后端热重载 + 前端）
make dev
```

管理后台访问 `http://localhost:4321`，API 访问 `http://localhost:8080`。

首次访问时，Web 安装向导将引导你完成数据库配置、站点设置和管理员账号创建。

### 常用命令

```bash
make dev              # 启动开发环境（后端热重载 + 前端）
make test             # 运行全部测试
make lint             # 代码检查（golangci-lint + Biome）
make migrate-up       # 执行数据库迁移
make migrate-down     # 回滚最近一次迁移
make build            # 构建生产二进制 + 前端
```

## 项目结构

```
sky-flux-cms/
├── cmd/cms/            # Cobra CLI（serve/migrate/version）
├── internal/           # 业务模块（auth、post、media、rbac 等）
│   ├── config/         # Viper 配置加载
│   ├── database/       # DB + Redis + RustFS 连接
│   ├── middleware/      # Gin 中间件链
│   ├── model/          # 共享数据模型（bun ORM）
│   ├── router/         # 路由注册
│   └── pkg/            # 共享工具包（apperror/jwt/crypto）
├── migrations/         # bun Go 代码迁移
├── web/                # Astro 5 管理后台
├── docs/               # 设计文档
├── docker-compose.yml  # PostgreSQL、Redis、Meilisearch、RustFS
└── Makefile
```

## 设计文档

所有设计文档位于 `docs/` 目录：

| 文档 | 内容 |
|------|------|
| prd.md | 产品需求与功能范围 |
| architecture.md | 系统架构与技术决策 |
| api.md | API 设计（OpenAPI 3.1） |
| database.md | 数据库 Schema 与索引策略 |
| story.md | 用户故事与验收标准 |
| security.md | 安全策略与威胁模型 |

## 开源协议

[MIT](./LICENSE)

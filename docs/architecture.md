# CMS 内容管理系统 — 系统架构文档

**版本**：v2.0
**日期**：2026-02-24
**状态**：草稿

---

## 1. 整体架构图

```mermaid
flowchart TB
    subgraph Clients["Client Layer"]
        AD["🖥️ Admin Dashboard\nAstro 5 + React 19\nshadcn/ui + TanStack Query"]
        CF["🌐 Consumer Frontend\n任意前端框架\n(Headless API 消费者)"]
    end

    subgraph Gateway["API Gateway"]
        NG["Nginx\nHTTPS 终结 | 负载均衡 | 静态文件"]
    end

    subgraph API["Gin API Server  (Go 1.25+)"]
        direction TB
        subgraph MW["Middleware Chain"]
            direction LR
            M0["Installation\nGuard"]
            M1["Site\nResolver"]
            M2["Schema\nMiddleware"]
            M3["JWT Auth"]
            M4["RBAC"]
            M5["Rate Limit"]
            M6["CORS"]
            M7["Audit Log"]
            M8["Recovery"]
            M0 --> M1 --> M2 --> M3 --> M4
        end
        subgraph Modules["业务模块"]
            direction LR
            AM["Auth Module\nJWT / Dynamic RBAC / 2FA"]
            CM["Content Module\nPost / Category / Tag"]
            MM["Media Module\nUpload + Process"]
            UM["User Module\nCRUD / Role"]
            PA["Public API\nHeadless"]
            SM["System Module\nConfig / Audit"]
            SITE["Site Module\nMulti-Site CRUD"]
            COMMENT["Comment Module\nModeration / Spam"]
            MENU["Menu Module\nNavigation"]
            REDIR["Redirect Module\nURL Redirect"]
            PREVIEW["Preview Module\nDraft Preview"]
            FEED["Feed Module\nRSS / Atom / Sitemap"]
            SETUP["Setup Module\nInstallation Wizard"]
        end
    end

    subgraph Storage["Storage Layer"]
        direction TB
        subgraph PGSchemas["PostgreSQL 18 (Schema Isolation)"]
            PGPublic[("public schema\nusers / sites / RBAC 9表 /\nrefresh_tokens /\nuser_totp /\nconfigs")]
            PGSite1[("site_{slug} schema\nsfc_site_posts / sfc_site_categories /\nsfc_site_tags / sfc_site_media_files /\nsfc_site_comments / sfc_site_menus /\nsfc_site_redirects / sfc_site_preview_tokens /\nsfc_site_api_keys / sfc_site_audits /\nsfc_site_configs")]
            PGSiteN[("site_{slug_n} ...\n(identical structure)")]
        end
        RD[("Redis 8\nSession 黑名单\n内容缓存\n限流计数\n分布式锁\n2FA 防重放")]
        FS["RustFS\nS3 Compatible"]
    end

    AD & CF --> NG
    NG --> API
    MW --> Modules
    API --> PGSchemas
    API --> RD
    MM --> FS
```

### 1.1 Multi-Site Schema Isolation

```mermaid
flowchart LR
    subgraph Request["Incoming Request"]
        REQ["Host: blog.example.com\nGET /api/v1/posts"]
    end

    subgraph Middleware["Middleware Chain"]
        IG["InstallationGuard\n已安装? → 继续"]
        SR["SiteResolver\nHost → slug=blog"]
        SC["SchemaMiddleware\nSET search_path TO\n'site_blog', 'public'"]
        AUTH["AuthMiddleware\nJWT → user_id\nRole from sfc_user_roles (global)"]
        RBAC["RBACMiddleware\nCheck sfc_role_apis\n(two-level cache)"]
    end

    subgraph Handler["Route Handler"]
        H["PostHandler.List()\nSELECT * FROM sfc_site_posts\n(resolves to site_blog.sfc_site_posts)"]
    end

    REQ --> IG --> SR --> SC --> AUTH --> RBAC --> H

    style SC fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px
```

**Schema Layout**:

```
PostgreSQL Database
├── public (global schema)
│   ├── sfc_users              -- all user accounts
│   ├── sfc_sites              -- site registry (slug → schema mapping)
│   ├── sfc_roles              -- 角色定义（name, slug, built_in）
│   ├── sfc_user_roles         -- 全局用户-角色分配（user_id, role_id）
│   ├── sfc_apis               -- API 端点注册（method, path, group，启动时自动同步）
│   ├── sfc_role_apis          -- 角色-API 权限映射
│   ├── sfc_menus              -- 后台管理菜单（区别于 site schema 的 sfc_site_menus 导航菜单）
│   ├── sfc_role_menus         -- 角色-菜单可见性映射
│   ├── sfc_role_templates     -- 权限模板定义
│   ├── sfc_role_template_apis -- 模板-API 映射
│   ├── sfc_role_template_menus-- 模板-菜单映射
│   ├── sfc_refresh_tokens     -- JWT refresh tokens
│   ├── sfc_user_totp          -- 2FA TOTP secrets (AES-256-GCM encrypted)
│   └── sfc_configs            -- global flags (e.g., system.installed)
│
├── site_blog (site schema for slug="blog")
│   ├── sfc_site_post_types / sfc_site_posts / sfc_site_post_translations / sfc_site_post_revisions
│   ├── sfc_site_categories / sfc_site_tags / sfc_site_post_category_map / sfc_site_post_tag_map
│   ├── sfc_site_media_files
│   ├── sfc_site_comments      -- threaded comments with moderation
│   ├── sfc_site_menus / sfc_site_menu_items -- navigation menus
│   ├── sfc_site_redirects     -- URL redirect rules (301/302)
│   ├── sfc_site_preview_tokens -- draft preview tokens (SHA-256 hashed)
│   ├── sfc_site_api_keys
│   ├── sfc_site_audits        -- partitioned by month
│   └── sfc_site_configs       -- site-level settings
│
├── site_docs (site schema for slug="docs")
│   └── ... (identical table structure)
│
└── site_{slug} ...
```

---

## 2. 项目目录结构

```
sky-flux-cms/
├── cmd/
│   └── cms/                            # Cobra CLI 单一二进制（serve/migrate 子命令）
│
├── internal/
│   │── ── ── 业务模块（按模块分包，每个模块自包含 handler/service/repo） ── ── ──
│   │
│   ├── auth/                           # 认证模块（登录/JWT/刷新Token/2FA）
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── user/                           # 用户模块（CRUD / 角色管理）
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── post/                           # 文章模块（CRUD / 修订 / 翻译）
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── category/                       # 分类模块
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── tag/                            # 标签模块
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── media/                          # 媒体模块（上传 / 处理 / 存储）
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── comment/                        # 评论模块（审核 / 垃圾检测 / Gravatar）
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── menu/                           # 导航菜单模块（CRUD / 排序）
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── redirect/                       # URL 重定向模块（301/302 / 导入导出）
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── preview/                        # 草稿预览模块（Token 生成/验证）
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── site/                           # 站点管理模块（多站点 CRUD + Schema 创建）
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── setup/                          # 安装向导模块（check / initialize）
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── dto.go
│   ├── feed/                           # RSS / Atom / Sitemap 模块
│   │   ├── handler.go
│   │   └── service.go
│   ├── apikey/                         # API Key 模块
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── audit/                          # 审计日志模块
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── middleware.go
│   ├── system/                         # 系统配置模块
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── dto.go
│   ├── rbac/                            # 动态 RBAC 模块（角色 / API 权限 / 菜单 / 模板）
│   │   ├── handler.go                   # RBAC 管理 API handlers
│   │   ├── service.go                   # 两级 Redis 缓存 service (L1 用户角色 / L2 角色-API)
│   │   ├── api_registry.go             # 启动时路由自动发现，同步到 sfc_apis
│   │   ├── interfaces.go              # Repository 接口定义
│   │   ├── dto.go                      # 请求/响应 DTO
│   │   ├── role_repo.go               # sfc_roles CRUD
│   │   ├── user_role_repo.go          # sfc_user_roles 用户-角色分配
│   │   ├── api_repo.go                # sfc_apis + sfc_role_apis CRUD
│   │   ├── role_api_repo.go           # 角色-API 映射查询
│   │   ├── menu_repo.go              # sfc_menus + sfc_role_menus CRUD
│   │   └── template_repo.go          # sfc_role_templates CRUD
│   │
│   │── ── ── 共享层（跨模块复用） ── ── ──
│   │
│   ├── model/                          # 共享数据模型
│   │   ├── user.go
│   │   ├── post.go
│   │   ├── category.go
│   │   ├── tag.go
│   │   ├── media.go
│   │   ├── apikey.go
│   │   ├── audit.go
│   │   ├── refresh_token.go
│   │   ├── post_translation.go
│   │   ├── post_type.go
│   │   ├── system_config.go
│   │   ├── site.go                     # 站点模型（slug, domain, settings）
│   │   ├── comment.go                  # 评论模型（含 guest/authenticated 双模式）
│   │   ├── menu.go                     # 菜单 + 菜单项模型
│   │   ├── redirect.go                 # URL 重定向模型
│   │   ├── preview_token.go            # 预览令牌模型
│   │   └── user_totp.go               # 2FA TOTP 模型（加密密钥 + 备用码）
│   ├── middleware/                     # 共享 Gin 中间件
│   │   ├── installation_guard.go      # 安装状态守卫（未安装 → 重定向 /setup）
│   │   ├── site_resolver.go           # 站点解析（Host / X-Site-Slug → site record）
│   │   ├── schema.go                  # Schema 切换（SET search_path TO site_{slug}, public）
│   │   ├── auth.go                    # JWT 验证
│   │   ├── rbac.go                    # 动态 RBAC 权限控制（调用 rbac.Service.CheckPermission）
│   │   ├── ratelimit.go              # Redis 滑动窗口限流
│   │   ├── cors.go
│   │   ├── request_id.go             # 请求 ID 注入
│   │   ├── logger.go                 # slog 请求日志
│   │   └── recovery.go
│   ├── config/                        # 配置加载（Viper）
│   │   └── config.go
│   ├── database/                      # DB + Redis + Meili + RustFS 连接
│   │   ├── postgres.go
│   │   ├── redis.go
│   │   ├── meilisearch.go            # Meilisearch 客户端初始化
│   │   └── rustfs.go                 # RustFS S3 兼容客户端（AWS SDK v2）
│   ├── schema/                        # Per-site Schema 管理
│   │   ├── template.go                # CreateSiteSchema(ctx, tx, slug) — 完整建表模板
│   │   ├── migrate.go                 # ForEachSiteSchema() — 批量迁移辅助
│   │   └── validate.go               # Schema 名称校验 (^[a-z0-9_]{3,50}$)
│   ├── router/                        # 路由注册（组装各模块 handler）
│   │   └── router.go
│   ├── cron/                          # 定时任务调度（robfig/cron）
│   │   └── scheduler.go
│   ├── testutil/                      # 测试工具包
│   │   ├── containers.go             # testcontainers-go PostgreSQL 辅助
│   │   └── httptest.go               # Gin HTTP 测试辅助
│   └── pkg/                           # 共享工具包
│       ├── apperror/                  # 全局错误定义
│       ├── response/                  # 统一响应格式（Success / Error / Paginated）
│       ├── jwt/                       # JWT 签发/验证（待实现）
│       ├── crypto/                    # 密码哈希 / TOTP 加密 / 预览 Token 生成（待实现）
│       ├── slug/                      # Slug 生成（待实现）
│       ├── paginator/                 # 分页（待实现）
│       └── storage/                   # RustFS S3 客户端封装（待实现）
│
├── migrations/                        # bun Go code migrations
│   ├── main.go                        # 迁移注册表（Migrations 全局变量）
│   ├── 20260224000001_create_core_tables.go    # 核心表（users, sites, tokens, totp, configs）
│   ├── 20260224000002_create_rbac_tables.go    # RBAC 9 张表
│   ├── 20260224000003_site_schema_placeholder.go  # 占位符（站点 schema 动态创建）
│   └── 20260224000004_seed_rbac_builtins.go    # 内置角色 + 权限模板种子数据
│
├── web/                               # Astro 管理后台
│   ├── src/
│   │   ├── components/
│   │   │   ├── ui/                    # shadcn/ui 组件（按需添加）
│   │   │   ├── layout/
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   ├── Header.tsx
│   │   │   │   └── MainLayout.tsx
│   │   │   ├── auth/
│   │   │   │   ├── LoginForm.tsx
│   │   │   │   └── TwoFactorForm.tsx  # 2FA TOTP 输入表单
│   │   │   ├── posts/
│   │   │   │   ├── PostList.tsx
│   │   │   │   ├── PostEditor.tsx
│   │   │   │   ├── PostSidebar.tsx    # SEO / 分类 / 标签
│   │   │   │   └── RevisionHistory.tsx
│   │   │   ├── media/
│   │   │   │   ├── MediaLibrary.tsx
│   │   │   │   └── MediaUploader.tsx
│   │   │   ├── categories/
│   │   │   │   └── CategoryTree.tsx
│   │   │   ├── comments/
│   │   │   │   ├── CommentList.tsx    # 评论审核列表
│   │   │   │   └── CommentModeration.tsx
│   │   │   ├── menus/
│   │   │   │   ├── MenuList.tsx
│   │   │   │   └── MenuEditor.tsx     # 菜单项拖拽排序编辑器
│   │   │   ├── redirects/
│   │   │   │   ├── RedirectList.tsx
│   │   │   │   └── RedirectForm.tsx
│   │   │   ├── settings/
│   │   │   │   └── TwoFactorSetup.tsx # 2FA 设置向导（QR 码 + 备用码）
│   │   │   └── dashboard/
│   │   │       ├── StatsCard.tsx
│   │   │       └── RecentPosts.tsx
│   │   ├── pages/                     # Astro 页面（SSR）
│   │   │   ├── index.astro            # 重定向到 /dashboard
│   │   │   ├── login.astro
│   │   │   ├── setup/
│   │   │   │   └── index.astro        # 安装向导页面
│   │   │   ├── dashboard/
│   │   │   │   └── index.astro
│   │   │   ├── posts/
│   │   │   │   ├── index.astro        # 文章列表
│   │   │   │   ├── new.astro          # 新建文章
│   │   │   │   └── [id].astro         # 编辑文章
│   │   │   ├── media/
│   │   │   │   └── index.astro
│   │   │   ├── categories/
│   │   │   │   └── index.astro
│   │   │   ├── tags/
│   │   │   │   └── index.astro
│   │   │   ├── comments/
│   │   │   │   └── index.astro        # 评论审核页
│   │   │   ├── menus/
│   │   │   │   └── index.astro        # 菜单管理页
│   │   │   ├── redirects/
│   │   │   │   └── index.astro        # 重定向管理页
│   │   │   ├── users/
│   │   │   │   └── index.astro
│   │   │   └── settings/
│   │   │       └── index.astro        # 含 2FA 设置区域
│   │   ├── lib/
│   │   │   ├── api.ts                 # API 客户端封装
│   │   │   ├── auth.ts                # Token 管理（含 2FA temp token 处理）
│   │   │   └── utils.ts
│   │   ├── hooks/                     # React 自定义 Hooks
│   │   │   ├── useAuth.ts
│   │   │   ├── usePosts.ts
│   │   │   ├── useMedia.ts
│   │   │   ├── useComments.ts
│   │   │   ├── useMenus.ts
│   │   │   └── useRedirects.ts
│   │   ├── stores/                    # Zustand 状态
│   │   │   ├── authStore.ts
│   │   │   ├── uiStore.ts
│   │   │   └── editorStore.ts         # 编辑器状态（自动保存、草稿）
│   │   └── middleware.ts              # Astro SSR 路由守卫
│   ├── astro.config.mjs
│   ├── tsconfig.json
│   ├── components.json                # shadcn/ui 配置
│   ├── package.json
│   └── Dockerfile
│
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── docker-compose.prod.yml
└── README.md
```

> **模块化架构说明**：Go 单体多模块分层架构 — 每个业务模块（auth/post/comment 等）自包含 handler/service/repository/dto，模块间不直接调用。共享层（model/middleware/config/database/schema/pkg）跨模块复用，router 是唯一组装点。

> **Post Types 版本说明**：V1.0 仅支持基础 `post_type` 字段（`article`/`page`）；V1.1 启用完整 Post Types CRUD（见 [api.md §14](./api.md)）。

---

## 3. 关键技术决策

### 3.1 为什么选择 Gin（而非 Echo/Fiber）
- 生态最成熟，中间件资源最丰富
- Gin 1.11 支持 HTTP/3
- 零分配路由，性能卓越
- 社区活跃，文档完善

### 3.2 为什么选择 Astro（而非 Next.js）
- Islands 架构：HTML 静态 + React 孤岛，管理后台无需重型 SPA
- SSR 模式支持服务端路由守卫（Token 验证）
- 按需加载 React，首屏更快
- shadcn/ui 官方支持 Astro 安装方式

### 3.3 Multi-Site: PostgreSQL Schema Isolation

每个站点使用独立的 PostgreSQL Schema（`site_{slug}`）进行数据隔离，全局实体（用户、认证、站点注册）存放在 `public` Schema。

| 决策 | 理由 |
|------|------|
| Schema-per-site（非 `site_id` 共享表） | 更强的隔离性，无跨站数据泄露风险，查询无需 `WHERE site_id = ?`，`DROP SCHEMA` 即可清理 |
| 用户全局化（`public` schema） | 单账号登录多站点 — 适合代理/自由职业者工作流 |
| 全局角色（`sfc_user_roles`） | 用户拥有全局角色，通过 `sfc_role_apis` 控制 API 访问权限 |
| 每请求 `search_path` | 查询自动解析到正确的站点 Schema；跨 Schema FK 到 `public.sfc_users(id)` 原生支持 |
| 内容表无 `site_id` 列 | Schema 本身提供隔离，无需复合索引 |
| JWT Claims 不携带 `role` | 角色从 `sfc_user_roles` 每请求解析，两级缓存（L1 用户角色 300s / L2 角色-API 600s） |

### 3.4 2FA: TOTP (RFC 6238) + AES-256-GCM

| 参数 | 值 |
|------|-----|
| 算法 | SHA-1（RFC 6238 标准，所有认证器 App 支持） |
| 位数 | 6 位 |
| 周期 | 30 秒 |
| 窗口 | +/-1（90 秒有效窗口） |
| 密钥加密 | AES-256-GCM，密钥从 `TOTP_ENCRYPTION_KEY` 环境变量加载 |
| 备用码 | 10 个，bcrypt cost=12 哈希存储，一次性使用 |

### 3.5 数据访问策略
```go
// 使用 uptrace/bun 轻量级 ORM
// 选择理由：
// 1. 链式查询 API，接近 Laravel Eloquent 的开发体验
// 2. 自动结构体映射，类型安全
// 3. 内置软删除、关联加载、迁移工具
// 4. 性能接近 database/sql，支持原生 SQL 回退

type PostRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*model.Post, error)
    FindAll(ctx context.Context, filter PostFilter) ([]*model.Post, int, error)
    Create(ctx context.Context, post *model.Post) error
    Update(ctx context.Context, post *model.Post) error
    SoftDelete(ctx context.Context, id uuid.UUID) error
}
```

### 3.6 缓存策略（Cache-Aside Pattern）

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant Gin as Gin Server
    participant Redis
    participant PG as PostgreSQL

    rect rgb(240, 248, 255)
        Note over Client,PG: 读取流程 (Cache-Aside Read)
        Client->>Gin: GET /api/public/v1/posts/:slug
        Gin->>Redis: GET site:{slug}:cache:post:detail:{id}
        alt 缓存命中 (Hit)
            Redis-->>Gin: JSON data (TTL 300s)
            Gin-->>Client: 200 响应 (< 20ms)
        else 缓存未命中 (Miss)
            Redis-->>Gin: nil
            Gin->>Gin: singleflight.Do(key, fn)<br/>防止缓存击穿
            Gin->>PG: SELECT * FROM sfc_site_posts WHERE slug=?<br/>(resolves to site_{slug}.sfc_site_posts via search_path)
            PG-->>Gin: post record
            Gin->>Redis: SET site:{slug}:cache:post:detail:{id} EX 300
            Gin-->>Client: 200 响应 (< 100ms)
        end
    end

    rect rgb(240, 255, 240)
        Note over Client,PG: 写入流程 (Cache Invalidation)
        Client->>Gin: PUT /api/v1/posts/:id
        Gin->>PG: UPDATE sfc_site_posts SET ...<br/>(resolves to site_{slug}.sfc_site_posts via search_path)
        PG-->>Gin: 成功
        Gin->>Redis: DEL site:{slug}:cache:post:detail:{id}
        Gin->>Redis: DEL site:{slug}:cache:post:list:* (pattern 批量失效)
        Gin->>Redis: DEL site:{slug}:cache:feed:* + site:{slug}:cache:sitemap:*
        Gin-->>Client: 200 响应
    end
```

#### Redis 不可用时的降级策略

缓存层设计遵循**优雅降级**原则——Redis 故障不应导致服务不可用：

| 场景 | 行为 | 说明 |
|------|------|------|
| Redis 连接失败 | 请求直接穿透到 PostgreSQL | 跳过缓存读写，确保服务可用 |
| Redis 读取超时 | 回退到数据库查询 | 设置短超时（如 100ms），避免阻塞 |
| Redis 写入失败 | 忽略错误，继续响应 | 下次请求会重新填充缓存 |

**关键原则**：
- **缓存故障记录为 Warning 级别日志**，不向客户端返回错误。缓存是加速层而非必要依赖
- **考虑引入 Circuit Breaker（熔断器）模式**：当 Redis 连续失败超过阈值时，暂时跳过所有缓存操作，定期探测恢复后自动恢复缓存访问，避免大量超时请求拖慢整体响应
- **Singleflight 不依赖 Redis**：`singleflight.Do()` 是进程内去重机制，即使 Redis 完全不可用，仍可防止同一进程内对相同 key 的并发数据库查询（防止缓存击穿时的 thundering herd 问题）

```go
// internal/service/cache_service.go — Redis 降级示例
func (s *CacheService) GetPostBySlug(ctx context.Context, slug string) (*model.Post, error) {
    siteSlug := ctx.Value("site_slug").(string)
    cacheKey := fmt.Sprintf("site:%s:cache:post:detail:%s", siteSlug, slug)

    // 尝试从 Redis 读取
    cached, err := s.redis.Get(ctx, cacheKey).Bytes()
    if err == nil {
        var post model.Post
        if json.Unmarshal(cached, &post) == nil {
            return &post, nil
        }
    }
    if err != nil && err != redis.Nil {
        // Redis 故障：记录 Warning，不返回错误
        slog.Warn("redis cache read failed, falling back to database",
            slog.String("key", cacheKey),
            slog.String("error", err.Error()))
    }

    // 缓存未命中或 Redis 不可用：查询数据库（singleflight 去重）
    v, err, _ := s.sfGroup.Do(cacheKey, func() (any, error) {
        return s.postRepo.FindBySlug(ctx, slug)
    })
    if err != nil {
        return nil, err
    }
    post := v.(*model.Post)

    // 尝试回填缓存（失败不影响响应）
    if data, err := json.Marshal(post); err == nil {
        if err := s.redis.Set(ctx, cacheKey, data, 300*time.Second).Err(); err != nil {
            slog.Warn("redis cache write failed", slog.String("key", cacheKey), slog.String("error", err.Error()))
        }
    }

    return post, nil
}
```

### 3.7 配置管理

使用 [viper](https://github.com/spf13/viper) 加载应用配置，支持多来源合并。

**配置优先级**（从高到低）：

| 优先级 | 来源 | 说明 |
|--------|------|------|
| 1 | CLI flag | `--port`, `--mode` 等命令行参数（通过 `viper.BindPFlag` 绑定） |
| 2 | 环境变量 | 生产环境通过 Docker / K8s 注入 |
| 3 | `.env` 文件 | 项目根目录 `.env` 文件（开发环境） |
| 4 | 默认值 | 代码中 `viper.SetDefault()` 定义 |

**配置文件格式**：`.env`（Key-Value），通过 Viper 加载：

```bash
# .env 结构示例（完整清单见 .env.example）
SERVER_PORT=8080
SERVER_MODE=debug
FRONTEND_URL=http://localhost:3000

DB_HOST=localhost
DB_PORT=5432
DB_NAME=cms
DB_USER=cms_user
DB_PASSWORD=devpassword
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=30m

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=devpassword
REDIS_DB=0

JWT_SECRET=your-secret-key-min-32-chars
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

RUSTFS_ENDPOINT=http://localhost:9000
RUSTFS_ACCESS_KEY=rustfsadmin
RUSTFS_SECRET_KEY=rustfsadmin
RUSTFS_BUCKET=cms-media

MEILI_URL=http://localhost:7700
MEILI_MASTER_KEY=devmasterkey

LOG_LEVEL=debug
LOG_FORMAT=json
```

> 完整环境变量清单请参阅 [deployment.md — 环境变量](./deployment.md)。

### 3.8 日志方案

选用 Go 1.21+ 内置的 `log/slog` 作为日志库，零外部依赖。

| 环境 | 格式 | Handler | 说明 |
|------|------|---------|------|
| 开发 | 文本（彩色） | `slog.NewTextHandler` | 终端可读性好 |
| 生产 | JSON | `slog.NewJSONHandler` | 便于 ELK / Loki 采集 |

```go
// internal/config/logger.go
func InitLogger(env string) *slog.Logger {
    var handler slog.Handler
    if env == "production" {
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })
    } else {
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    }
    return slog.New(handler)
}
```

> 日志级别规范请参阅 [standard.md — 日志规范](./standard.md)。

---

## 4. 请求数据流

### 4.1 Site-Scoped 请求流程

```
Request (Host: blog.example.com, Authorization: Bearer <jwt>)
  │
  ├─ 1. InstallationGuardMiddleware
  │     └─ Check: system.installed = true?
  │     └─ Not installed → redirect /setup (pages) or 503 (API)
  │     └─ Exception: /api/v1/setup/* and /setup always pass through
  │
  ├─ 2. SiteResolverMiddleware
  │     └─ Resolution order:
  │         1. X-Site-Slug header (admin API)
  │         2. Host header → domain lookup (public API)
  │         3. Fallback: single-site default
  │     └─ Cache: Redis site:slug:{slug} TTL=600s
  │     └─ Inject: c.Set("site", record), c.Set("site_slug", slug), c.Set("site_id", id)
  │
  ├─ 3. SchemaMiddleware
  │     └─ SET search_path TO 'site_{slug}', 'public'
  │     └─ Validates slug: ^[a-z0-9_]{3,50}$ (against sites table, not raw input)
  │     └─ All subsequent DB queries auto-resolve to site_{slug} schema
  │
  ├─ 4. AuthMiddleware (protected routes only)
  │     └─ Verify JWT (HS256), check blacklist in Redis
  │     └─ Inject: c.Set("user_id"), c.Set("user")
  │
  ├─ 5. RBACMiddleware (protected routes only)
  │     └─ 调用 rbac.Service.CheckPermission(ctx, userID, method, path)
  │     └─ L1 缓存: Redis user:{user_id}:roles TTL=300s
  │     └─ Fallback: SELECT r.slug, r.id FROM sfc_user_roles JOIN sfc_roles ON ...
  │     └─ 若 slugs 包含 "super" → 直接放行
  │     └─ L2 缓存: Redis role:{role_id}:api_set TTL=600s
  │     └─ Fallback: SELECT method, path FROM sfc_role_apis JOIN sfc_apis ON ...
  │     └─ 命中 method:path → 放行，否则 403
  │
  └─ 6. Route Handler
        └─ All DB queries scoped to site_{slug} schema via search_path
        └─ Cross-schema access to public.sfc_users via FK
```

### 4.2 Global 请求流程（Auth / 2FA / User Management）

```
Request (POST /api/v1/auth/login)
  │
  ├─ 1. InstallationGuardMiddleware (same as above)
  │
  ├─ 2. No SiteResolver / No SchemaMiddleware
  │     └─ Auth endpoints operate on public schema only
  │
  ├─ 3. Handler queries public.sfc_users, public.sfc_user_totp directly
  │     └─ 2FA flow: password → temp token (5min) → TOTP validation → full JWT
  │
  └─ 4. JWT issued without role claim
        └─ Role resolved per-request from sfc_user_roles via RBAC middleware (two-level cache)
```

---

## 5. 路由注册

```go
func SetupRouter(r *gin.Engine, deps *Dependencies) {
    // =============================================
    // Global Middleware (all routes)
    // =============================================
    r.Use(RecoveryMiddleware())
    r.Use(CORSMiddleware(deps.Config.CORS.AllowedOrigins))

    // =============================================
    // Setup / Installation (no auth, no site scope)
    // =============================================
    setup := r.Group("/api/v1/setup")
    setup.Use(RateLimitMiddleware(10, time.Minute)) // 10 req/min per IP
    {
        setup.POST("/check",      handler.SetupCheck)
        setup.POST("/initialize", handler.SetupInitialize)
    }

    // =============================================
    // Auth routes (global, no site scope)
    // =============================================
    r.POST("/api/v1/auth/login",          handler.Login)          // Modified for 2FA flow
    r.POST("/api/v1/auth/refresh",        handler.RefreshToken)
    r.POST("/api/v1/auth/2fa/validate",   handler.Validate2FA)    // Temp token auth

    // =============================================
    // Auth management (JWT required, global)
    // =============================================
    auth := r.Group("/api/v1/auth")
    auth.Use(InstallationGuardMiddleware(), JWTMiddleware())
    {
        auth.POST("/logout",            handler.Logout)
        auth.GET("/me",                 handler.GetCurrentUser)
        auth.PUT("/password",           handler.ChangePassword)

        // 2FA management (global, user-level)
        auth.GET("/2fa/status",         handler.Get2FAStatus)
        auth.POST("/2fa/setup",         handler.Setup2FA)
        auth.POST("/2fa/verify",        handler.Verify2FASetup)
        auth.POST("/2fa/disable",       handler.Disable2FA)
        auth.POST("/2fa/backup-codes",  handler.RegenerateBackupCodes)
    }

    // =============================================
    // Site management (global, RBAC 中间件自动匹配权限)
    // 权限通过 sfc_role_apis 动态配置，无需硬编码 RequireRole
    // =============================================
    sites := r.Group("/api/v1/sites")
    sites.Use(InstallationGuardMiddleware(), JWTMiddleware(), RBACMiddleware(deps.RBACService))
    {
        sites.GET("",          handler.ListSites)
        sites.POST("",         handler.CreateSite)
        sites.GET("/:slug",    handler.GetSite)
        sites.PUT("/:slug",    handler.UpdateSite)
        sites.DELETE("/:slug", handler.DeleteSite)
    }

    // Force disable 2FA for a user (global, under auth group, RBAC 控制)
    auth.DELETE("/2fa/users/:user_id", handler.ForceDisable2FA)

    // =============================================
    // Site-scoped Admin API (JWT + Schema + RBAC)
    // 所有端点权限通过 sfc_role_apis 动态配置，RBAC 中间件自动匹配 method+path
    // =============================================
    api := r.Group("/api/v1")
    api.Use(
        InstallationGuardMiddleware(),
        SiteResolverMiddleware(deps.SiteRepo),
        SchemaMiddleware(deps.DB),
        JWTMiddleware(),
        RBACMiddleware(deps.RBACService),
    )
    {
        // Posts
        posts := api.Group("/posts")
        {
            posts.GET("",              handler.ListPosts)
            posts.GET("/:id",          handler.GetPost)
            posts.POST("",             handler.CreatePost)
            posts.PUT("/:id",          handler.UpdatePost)
            posts.DELETE("/:id",       handler.DeletePost)
            posts.POST("/:id/preview",         handler.CreatePreviewToken)
            posts.GET("/:id/preview",          handler.ListPreviewTokens)
            posts.DELETE("/:id/preview",       handler.RevokeAllPreviewTokens)
            posts.DELETE("/:id/preview/:token_id", handler.RevokeSinglePreviewToken)
        }

        // Categories, Tags, Media (existing)
        // ...

        // Comments
        comments := api.Group("/comments")
        {
            comments.GET("",                handler.ListComments)
            comments.GET("/:id",            handler.GetComment)
            comments.PUT("/:id/status",     handler.UpdateCommentStatus)
            comments.PUT("/:id/pin",        handler.ToggleCommentPin)
            comments.POST("/:id/reply",     handler.AdminReplyComment)
            comments.PUT("/batch-status",   handler.BatchUpdateCommentStatus)
            comments.DELETE("/:id",         handler.DeleteComment)
        }

        // Menus
        menus := api.Group("/menus")
        {
            menus.GET("",                       handler.ListMenus)
            menus.POST("",                      handler.CreateMenu)
            menus.GET("/:id",                   handler.GetMenu)
            menus.PUT("/:id",                   handler.UpdateMenu)
            menus.DELETE("/:id",                handler.DeleteMenu)
            menus.POST("/:id/items",            handler.CreateMenuItem)
            menus.PUT("/:id/items/:item_id",    handler.UpdateMenuItem)
            menus.DELETE("/:id/items/:item_id", handler.DeleteMenuItem)
            menus.PUT("/:id/items/reorder",     handler.ReorderMenuItems)
        }

        // Redirects
        redirects := api.Group("/redirects")
        {
            redirects.GET("",              handler.ListRedirects)
            redirects.POST("",             handler.CreateRedirect)
            redirects.PUT("/:id",          handler.UpdateRedirect)
            redirects.DELETE("/:id",       handler.DeleteRedirect)
            redirects.DELETE("/batch",      handler.BulkDeleteRedirects)
            redirects.POST("/import",      handler.ImportRedirects)
            redirects.GET("/export",       handler.ExportRedirects)
        }
    }

    // =============================================
    // Site-scoped Public API (API Key auth, Host-based site resolution)
    // =============================================
    public := r.Group("/api/public/v1")
    public.Use(
        InstallationGuardMiddleware(),
        SiteResolverMiddleware(deps.SiteRepo),
        SchemaMiddleware(deps.DB),
        APIKeyMiddleware(),
        RateLimitMiddleware(100, time.Minute),
    )
    {
        public.GET("/posts",                   handler.PublicListPosts)
        public.GET("/posts/:slug",             handler.PublicGetPost)
        public.GET("/posts/:slug/comments",    handler.PublicListComments)
        public.POST("/posts/:slug/comments",   handler.PublicCreateComment)
        public.GET("/menus",                   handler.PublicGetMenu)
        public.GET("/preview/:token",          handler.PublicPreview)
    }

    // =============================================
    // Feed & Sitemap (unauthenticated, Host-based site resolution)
    // =============================================
    feeds := r.Group("")
    feeds.Use(
        InstallationGuardMiddleware(),
        SiteResolverMiddleware(deps.SiteRepo),
        SchemaMiddleware(deps.DB),
    )
    {
        feeds.GET("/feed/rss.xml",           handler.RSSFeed)
        feeds.GET("/feed/atom.xml",          handler.AtomFeed)
        feeds.GET("/sitemap.xml",            handler.SitemapIndex)
        feeds.GET("/sitemap-posts.xml",      handler.SitemapPosts)
        feeds.GET("/sitemap-categories.xml", handler.SitemapCategories)
        feeds.GET("/sitemap-tags.xml",       handler.SitemapTags)
    }
}
```

---

## 6. Docker Compose 配置

### 服务依赖关系

```mermaid
flowchart LR
    PG[("🐘 postgres:18\n:5432")]
    RD[("🔴 redis:8\n:6379")]
    BE["⚙️ backend\n:8080"]
    FE["🌐 frontend\n:3000"]
    NG["🔀 nginx\n:80 / :443"]

    PG -->|healthy| BE
    RD -->|healthy| BE
    BE --> FE
    BE --> NG
    FE --> NG
```

```yaml
# docker-compose.yml
# Docker Compose V2 不再需要 version 字段，已移除

services:
  postgres:
    image: postgres:18-alpine
    environment:
      POSTGRES_DB: cms
      POSTGRES_USER: cms_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U cms_user -d cms"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:8-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD} --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD-SHELL", "redis-cli -a \"${REDIS_PASSWORD}\" ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  backend:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      DATABASE_URL: postgres://cms_user:${DB_PASSWORD}@postgres:5432/cms?sslmode=disable
      REDIS_URL: redis://:${REDIS_PASSWORD}@redis:6379/0
      JWT_SECRET: ${JWT_SECRET}
      TOTP_ENCRYPTION_KEY: ${TOTP_ENCRYPTION_KEY}
      APP_PORT: 8080
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      # 媒体文件存储在 RustFS 对象存储中，无需本地 volume

  frontend:
    build:
      context: ./web
      dockerfile: Dockerfile
    environment:
      PUBLIC_API_URL: http://backend:8080
    ports:
      - "3000:3000"
    depends_on:
      - backend

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
    depends_on:
      - backend
      - frontend

volumes:
  postgres_data:
  redis_data:
  sfc_site_media_files:
```

---

## 6.1 Makefile 构建任务

项目根目录提供 `Makefile` 用于常用开发命令的快捷执行：

```makefile
.PHONY: dev test build migrate-up migrate-down lint

dev:          ## Start development environment
	docker compose up -d
	air  # Go hot-reload

test:         ## Run all tests
	go test ./... -v -count=1

build:        ## Build production images
	docker compose -f docker-compose.prod.yml build

migrate-up:   ## Run database migrations (public + all site schemas)
	go run ./cmd/cms migrate up

migrate-down: ## Rollback last migration
	go run ./cmd/cms migrate down

lint:         ## Run linters
	golangci-lint run ./...
	cd web && bun run lint
```

使用方式：

```bash
# 查看所有可用目标
make help

# 启动开发环境
make dev

# 运行测试
make test

# 执行数据库迁移（public schema + 所有 site_{slug} schema）
make migrate-up

# 代码检查
make lint
```

> 完整 Makefile 位于项目根目录，上述为核心目标摘要。

---

## 7. 安全设计

| 威胁 | 防御措施 |
|------|----------|
| SQL 注入 | 参数化查询（uptrace/bun `?` 占位符，自动转换为 `$1`），无字符串拼接 |
| Schema 注入 | 站点 slug 严格校验 `^[a-z0-9_]{3,50}$`，仅从 `sites` 表解析，不使用原始用户输入 |
| XSS | 富文本 HTML 白名单过滤（bluemonday），评论内容剥离所有 HTML |
| CSRF | SameSite Cookie + Origin 校验 |
| 暴力破解 | Redis 登录失败计数 + 账号锁定；2FA TOTP 5 次/5 分钟限制 |
| API 滥用 | 滑动窗口限流（Redis + Lua Script） |
| 敏感数据 | 密码 bcrypt cost=12，API Key SHA-256，TOTP 密钥 AES-256-GCM 加密 |
| Token 泄露 | Access Token 短有效期（15min），登出即失效（黑名单） |
| 文件上传 | MIME 类型白名单、文件大小限制、随机文件名 |
| 跨站数据泄露 | Schema Isolation — 每站独立 Schema，`search_path` 每请求设定 |
| 评论垃圾 | Honeypot + IP 限流 + 关键词黑名单 + 重复检测 |
| 预览 Token | 256 bits 熵，SHA-256 哈希存储，24h 过期，每篇最多 5 个 |
| 安装安全 | `pg_advisory_xact_lock` 防并发，安装后 setup 端点永久禁用 |

---

## 8. 性能优化策略

```
数据库层：
├── 连接池：pgx 连接池，max_open=25，max_idle=10
├── 查询优化：EXPLAIN ANALYZE 分析慢查询
├── 全文检索：Meilisearch（独立搜索引擎，CJK 原生支持）
├── 分区表：sfc_site_audits 按月分区（每站点 Schema 独立分区）
└── Schema 隔离：无 site_id 过滤开销，索引更紧凑

缓存层：
├── L1：应用内 sync.Map（热点配置，如系统设置、安装状态）
├── L2：Redis（内容列表 60s，详情 300s，菜单 300s，重定向 600s，Feed 3600s）
└── CDN：媒体文件通过 CDN 分发

API 层：
├── 响应压缩：Gzip（Gin 中间件）
├── 并发控制：Goroutine Pool + Context 超时
├── 请求聚合：批量接口减少往返次数
└── 预取关联：文章列表预加载分类/标签（JOIN，避免 N+1）
```

---

## 9. Astro 前端架构

```typescript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import tailwind from '@astrojs/tailwind';
import node from '@astrojs/node';

export default defineConfig({
  output: 'server',              // SSR 模式
  adapter: node({ mode: 'standalone' }),
  integrations: [
    react(),                     // React Islands
    tailwind({ applyBaseStyles: false }),
  ],
  vite: {
    define: {
      'import.meta.env.PUBLIC_API_URL': JSON.stringify(process.env.PUBLIC_API_URL)
    }
  }
});
```

```typescript
// src/middleware.ts — 路由守卫
import { defineMiddleware } from 'astro:middleware';

const PUBLIC_ROUTES = ['/login', '/setup'];

export const onRequest = defineMiddleware(async (context, next) => {
  const isPublicRoute = PUBLIC_ROUTES.some(r => context.url.pathname.startsWith(r));

  if (!isPublicRoute) {
    const token = context.cookies.get('access_token')?.value;
    if (!token) {
      return context.redirect('/login');
    }
    // 验证 token 有效性（可调用后端 /api/v1/auth/me）
  }

  return next();
});
```

```tsx
// 典型 React Island 组件：文章列表
// src/components/posts/PostList.tsx
import { useQuery } from '@tanstack/react-query';
import { DataTable } from '@/components/ui/data-table';
import { fetchPosts } from '@/lib/api';

export default function PostList() {
  const { data, isLoading } = useQuery({
    queryKey: ['posts'],
    queryFn: () => fetchPosts({ page: 1, per_page: 20 }),
    staleTime: 30_000,
  });

  if (isLoading) return <PostListSkeleton />;

  return (
    <DataTable
      columns={postColumns}
      data={data?.data ?? []}
      pagination={data?.pagination}
    />
  );
}
```

### 9.1 前端错误处理策略

```
错误处理分层：
├── 渲染层：React Error Boundary
│   └── 全局 ErrorBoundary 包裹应用根组件
│       → 捕获 React 渲染异常，显示友好错误页面（含重试按钮）
│
├── 网络层：API 客户端响应拦截器
│   └── axios / ky 实例统一配置
│       → 401 Unauthorized：自动清除 Token 并跳转 /login
│       → 403 Forbidden：提示权限不足
│       → 500+ Server Error：通过 Sonner Toast 弹出错误提示
│
└── 数据层：TanStack Query 全局 onError
    └── QueryClient 默认 onError 回调
        → 统一处理查询/变更失败，避免各组件重复错误处理逻辑
```

```tsx
// src/lib/query-client.ts — 全局错误处理配置示例
import { QueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
    },
    mutations: {
      onError: (error) => {
        toast.error(error.message || '操作失败，请稍后重试');
      },
    },
  },
});
```

### 9.2 前端国际化方案

管理后台界面语言与内容多语言是两个独立层面：

| 维度 | 方案 | 说明 |
|------|------|------|
| **界面语言**（UI labels） | `react-i18next` | 管理后台按钮、菜单、提示文案的多语言 |
| **内容多语言**（CMS 数据） | `sfc_site_post_translations` 表 | 文章/页面内容的多语言存储，与界面语言无关 |

**支持语言**：中文（zh-CN）、英文（en），默认 zh-CN。

```
web/src/
├── i18n/
│   ├── config.ts          # i18next 初始化配置
│   ├── zh-CN.json         # 中文翻译资源
│   └── en.json            # 英文翻译资源
```

- **界面语言切换**：用户在后台设置中选择偏好语言，存储在 `localStorage`，`react-i18next` 根据偏好切换 UI 文案
- **内容多语言**：文章编辑页通过语言选项卡切换不同翻译版本，数据存储在 `sfc_site_post_translations` 表中（详见 [database.md](./database.md)）

### 9.3 2FA 登录流程前端集成

```
Login Flow:
1. 用户输入 email + password → POST /api/v1/auth/login
2. 如果响应包含 requires_2fa: true：
   a. 显示 TOTP 输入表单
   b. 将 temp_token 存储在内存（非 localStorage）
   c. 用户输入 6 位 TOTP 码 → POST /api/v1/auth/2fa/validate (Bearer temp_token)
   d. 成功 → 获取 access_token，正常登录流程
   e. temp_token 过期（5min）→ 重定向登录页，提示"2FA 验证超时"
3. 如果无 2FA → 正常登录（与之前相同）
```

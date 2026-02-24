# Batch 4 设计文档 — 基础设施 + 简单 CRUD

**日期**: 2026-02-24
**范围**: 15 endpoints + 2 基础设施组件 + Boolean→Smallint 迁移
**前置**: Batch 1 (Setup+Auth), Batch 2+3 (Sites+RBAC) 已完成

---

## 1. 总体架构

| 组件 | 类型 | 端点数 | 路由组 |
|------|------|--------|--------|
| Boolean→Smallint 迁移 | 迁移 | — | — |
| AuditService (pkg) | 基础设施 | — | — |
| Mail Service (pkg) | 基础设施 | — | — |
| 站点级路由组 | Router 扩展 | — | SiteResolver → Schema → Auth → RBAC |
| Users | 全局路由组 | 5 | JWT + RBAC (Super) |
| Settings | 站点级路由组 | 2 | Admin+ / Super |
| API Keys | 站点级路由组 | 3 | Admin+ |
| Post Types | 站点级路由组 | 4 | Viewer+ / Admin+ |
| Audit Logs | 站点级路由组 | 1 | Super |

---

## 2. Boolean → Smallint 统一迁移

全部 15 个 boolean 字段统一为 `smallint` 整数枚举，提升一致性和可扩展性。

### 枚举定义 (enums.go 新增)

```go
// 通用二元枚举 — built_in, revoked, enabled, primary, pinned 共用
type Toggle int8
const (
    ToggleNo  Toggle = iota + 1 // 1
    ToggleYes                    // 2
)

type UserStatus     int8 // 1=Active, 2=Disabled
type SiteStatus     int8 // 1=Active, 2=Disabled
type RoleStatus     int8 // 1=Active, 2=Disabled
type APIStatus      int8 // 1=Active, 2=Disabled
type MenuStatus     int8 // 1=Active, 2=Hidden
type APIKeyStatus   int8 // 1=Active, 2=Revoked
type RedirectStatus int8 // 1=Active, 2=Disabled
type MenuItemStatus int8 // 1=Active, 2=Hidden
```

### 字段映射

| Model | 旧字段 | 新字段 | 枚举类型 | 值 |
|-------|--------|--------|----------|-----|
| User | `is_active bool` | `status smallint` | `UserStatus` | 1=Active, 2=Disabled |
| Site | `is_active bool` | `status smallint` | `SiteStatus` | 1=Active, 2=Disabled |
| Role | `status bool` | `status smallint` | `RoleStatus` | 1=Active, 2=Disabled |
| Role | `built_in bool` | `built_in smallint` | `Toggle` | 1=No, 2=Yes |
| APIEndpoint | `status bool` | `status smallint` | `APIStatus` | 1=Active, 2=Disabled |
| AdminMenu | `status bool` | `status smallint` | `MenuStatus` | 1=Active, 2=Hidden |
| APIKey | `is_active bool` | `status smallint` | `APIKeyStatus` | 1=Active, 2=Revoked |
| Redirect | `is_active bool` | `status smallint` | `RedirectStatus` | 1=Active, 2=Disabled |
| MenuItem | `is_active bool` | `status smallint` | `MenuItemStatus` | 1=Active, 2=Hidden |
| PostType | `built_in bool` | `built_in smallint` | `Toggle` | 1=No, 2=Yes |
| RoleTemplate | `built_in bool` | `built_in smallint` | `Toggle` | 1=No, 2=Yes |
| RefreshToken | `revoked bool` | `revoked smallint` | `Toggle` | 1=No, 2=Yes |
| TOTP | `is_enabled bool` | `enabled smallint` | `Toggle` | 1=No, 2=Yes |
| PostCategoryMap | `is_primary bool` | `primary smallint` | `Toggle` | 1=No, 2=Yes |
| Comment | `is_pinned bool` | `pinned smallint` | `Toggle` | 1=No, 2=Yes |

### 迁移策略

新增迁移脚本（migration 5），对 public schema 和 site schema DDL 模板同时修改：

- `ALTER COLUMN` + `TYPE smallint USING CASE WHEN old_value THEN 2 ELSE 1 END`（built_in 等 true=Yes=2）
- `ALTER COLUMN` + `TYPE smallint USING CASE WHEN old_value THEN 1 ELSE 2 END`（is_active 等 true=Active=1）
- `RENAME COLUMN` 重命名字段
- `ALTER COLUMN SET DEFAULT` 设置新默认值

### Auth 模块联动

Auth service 中所有 `user.IsActive` 判断改为 `user.Status == model.UserStatusActive`。

---

## 3. AuditService

**位置**: `internal/pkg/audit/`

### 接口

```go
type Logger interface {
    Log(ctx context.Context, entry Entry) error
}

type Entry struct {
    Action           model.LogAction
    ResourceType     string
    ResourceID       string
    ResourceSnapshot any  // json.Marshal → jsonb
}
```

### 设计决策

- **Service 层显式调用**（非中间件自动捕获）：审计需要业务语义（resource_type, resource_id, snapshot），中间件层无法获取
- **ctx 自动提取**: ActorID、ActorEmail、IPAddress、UserAgent 从 context 自动提取，调用方只需提供业务字段
- **同步写入**: 审计日志是合规要求，丢失不可接受，单条 INSERT 开销极低
- **AuditContext 中间件**: 新增轻量中间件（~15 行），将 `c.ClientIP()` 和 `c.GetHeader("User-Agent")` 写入 ctx

---

## 4. Mail Service

**位置**: `internal/pkg/mail/`

### 接口

```go
type Sender interface {
    Send(ctx context.Context, msg Message) error
}

type Message struct {
    To      string
    Subject string
    HTML    string
}
```

### 实现

- `ResendSender`: 封装 `resend-go/v3` SDK
- 配置从 `*config.Config` 读取: `RESEND_API_KEY`, `RESEND_FROM_NAME`, `RESEND_FROM_EMAIL`
- 模板: 内嵌 Go `html/template`
  - `welcome.html` — 新用户欢迎邮件（含临时密码）
  - `account_disabled.html` — 账号禁用通知

### 调用方式

异步 goroutine，失败仅日志记录，不阻塞主流程：
```go
go func() {
    if err := s.mailer.Send(bgCtx, mail.Message{...}); err != nil {
        slog.Error("failed to send email", "email", user.Email, "error", err)
    }
}()
```

---

## 5. Router 扩展 — 站点级路由组

**变更文件**: `internal/router/router.go`

### 新增站点级路由组

```go
siteScoped := v1.Group("")
siteScoped.Use(middleware.SiteResolver(siteRepo))
siteScoped.Use(middleware.Schema(db))
siteScoped.Use(middleware.AuditContext())
siteScoped.Use(middleware.Auth(jwtMgr))
siteScoped.Use(middleware.RBAC(rbacSvc))
```

中间件链: `SiteResolver → SchemaMiddleware → AuditContext → Auth → RBAC → Handler`

### SiteResolver 依赖

复用现有 `site.SiteRepo.GetBySlug`，抽取接口给 middleware 层使用。

---

## 6. Users 模块 (5 endpoints)

**位置**: `internal/user/` — handler.go / service.go / repository.go / dto.go / interfaces.go

**路由组**: 全局（JWT + RBAC, Super），和 sites/rbac 并列

### 端点

```
GET    /api/v1/users        — 列表（分页 + q 搜索 + role 筛选）
POST   /api/v1/users        — 创建（临时密码 + 欢迎邮件）
GET    /api/v1/users/:id    — 详情
PUT    /api/v1/users/:id    — 更新（display_name / role / status）
DELETE /api/v1/users/:id    — 软删除
```

### 与 Auth 模块职责分界

| 操作 | Auth | User |
|------|------|------|
| 登录/登出/密码/2FA | ✓ | |
| 创建/删除用户 | | ✓ |
| 修改他人角色/状态 | | ✓ |
| 查询用户列表 | | ✓ |

### 关键业务逻辑

- **创建用户**: 生成随机临时密码 → bcrypt hash → INSERT sfc_users → 分配角色 sfc_user_roles → 异步欢迎邮件
- **禁用用户**: status 从 1→2 时，异步禁用通知邮件 + 吊销所有 refresh tokens
- **删除用户**: 软删除 + 吊销 refresh tokens + JWT blacklist
- **不可自删**: actor.ID == target.ID → 403
- **不可删最后 super**: 至少保留一个 Super 用户
- **复用**: `pkg/crypto` (密码生成), `pkg/jwt` (blacklist)

---

## 7. Settings 模块 (2 endpoints)

**位置**: `internal/system/` — handler.go / service.go / repository.go / dto.go / interfaces.go

**路由组**: 站点级

### 端点

```
GET /api/v1/settings      — 列表（Admin+）
PUT /api/v1/settings/:key — 更新（Super）
```

**数据表**: `sfc_site_configs`（key-value，value 为 jsonb）

**审计**: 更新时记录 `LogActionSettingsChange`，snapshot 包含旧值和新值。

---

## 8. API Keys 模块 (3 endpoints)

**位置**: `internal/apikey/` — handler.go / service.go / repository.go / dto.go / interfaces.go

**路由组**: 站点级（Admin+）

### 端点

```
GET    /api/v1/api-keys      — 列表
POST   /api/v1/api-keys      — 创建（返回明文 key，仅此一次）
DELETE /api/v1/api-keys/:id  — 吊销（设置 revoked_at + status=2）
```

### Key 生成策略

- 格式: `cms_live_` + 32 字节 crypto/rand hex ≈ 40 字符
- 存储: SHA-256 hash → `key_hash`，前缀 `cms_live_a1b2` → `key_prefix`
- 创建响应返回一次明文，之后只能看到 prefix
- 复用 `pkg/crypto` 的 hash 函数

---

## 9. Post Types 模块 (4 endpoints)

**位置**: `internal/posttype/`（Go 包名小写无下划线）

**路由组**: 站点级

### 端点

```
GET    /api/v1/post-types      — 列表（Viewer+）
POST   /api/v1/post-types      — 创建（Admin+）
PUT    /api/v1/post-types/:id  — 更新（Admin+）
DELETE /api/v1/post-types/:id  — 删除（Admin+）
```

### Fields Schema 校验

`fields` 是 jsonb 数组，Service 层校验每个元素：
```json
{ "name": "string", "label": "string", "type": "string|number|boolean|date|rich_text", "required": bool, "default_value": any }
```

**保护**: `built_in = ToggleYes` 的类型不可删除、不可修改 slug。

---

## 10. Audit Logs 模块 (1 endpoint)

**位置**: `internal/audit/` — handler.go / repository.go / dto.go / interfaces.go

> 无 service 层 — 只有只读查询，handler 直接调用 repo。

**路由组**: 站点级（Super）

### 端点

```
GET /api/v1/audit-logs — 查询（支持 actor_id, action, resource_type, start_date, end_date 过滤 + 分页）
```

### 跨 Schema 查询

actor 字段需 JOIN public.sfc_users 获取 display_name：
```go
db.NewSelect().
    TableExpr("sfc_site_audits AS a").
    ColumnExpr("a.*").
    ColumnExpr("u.display_name AS actor_display_name").
    Join("LEFT JOIN public.sfc_users AS u ON u.id = a.actor_id")
```

---

## 11. 依赖关系

```
Boolean→Smallint 迁移 (migration 5)
        │
        ├── Model 更新 + Auth 适配
        │
        ▼
AuditService + MailService (pkg)
        │
        ▼
Router 站点级路由组
        │
        ├── Users (全局路由组, 复用 mail + audit)
        ├── Settings (站点级, 复用 audit)
        ├── API Keys (站点级, 复用 crypto)
        ├── Post Types (站点级)
        └── Audit Logs (站点级, 只读)
```

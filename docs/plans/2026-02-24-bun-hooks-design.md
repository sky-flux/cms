# Bun ORM Model Hooks 设计

> **日期**: 2026-02-24
> **状态**: 已批准
> **范围**: BeforeAppendModel hook 统一管理时间戳 + 输入标准化 + 乐观锁

## 背景

迁移重写设计（`migration-update-design.md`）已决定移除所有 `update_updated_at()` 数据库触发器。
本设计定义 bun ORM 应用层 hook 作为触发器的替代方案，同时扩展到 email 标准化和乐观锁等场景。

## 决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 实现方案 | 通用辅助函数 + 每个 Model 各自实现 BeforeAppendModel | DRY 时间戳逻辑，同时保留 model 特定扩展能力 |
| Slug 生成 | Service 层 | 涉及唯一性检查、CJK transliteration 等需要 DB 访问的复杂逻辑 |
| 内容清理 | Service 层 | HTML/XSS 清理逻辑复杂，允许部分 Markdown |
| 软删除过滤 | bun 内置 `soft_delete` tag | 已在所有软删除 model 上配置，无需额外 hook |
| DB DEFAULT | 保留 `DEFAULT NOW()` | Hook 为主、DB DEFAULT 兜底，防御性编程 |

## 架构

### 新增文件

```
internal/model/hooks.go    # 通用辅助函数（SetTimestamps, SetUpdatedAt, NormalizeEmail）
```

### 辅助函数

```go
// SetTimestamps — INSERT 时设置 created_at + updated_at，UPDATE 时仅设置 updated_at
func SetTimestamps(createdAt *time.Time, updatedAt *time.Time, query bun.Query)

// SetUpdatedAt — 仅处理 updated_at（用于没有 created_at 的 model，如 Config）
func SetUpdatedAt(updatedAt *time.Time, query bun.Query)

// NormalizeEmail — 统一邮箱格式：lowercase + trim
func NormalizeEmail(email *string)
```

### Hook 边界原则

- **Model hook 适合**: 无副作用的纯值转换（时间戳、格式标准化、计数器自增）
- **Service 层适合**: 需要 DB 查询或外部依赖的逻辑（slug 唯一性、CJK transliteration、内容清理）

## Model Hook 映射

### 需要实现 BeforeAppendModel 的 Model（17 个）

| # | Model | 文件 | 时间戳函数 | 额外逻辑 |
|---|-------|------|-----------|---------|
| 1 | User | user.go | SetTimestamps | NormalizeEmail |
| 2 | Site | site.go | SetTimestamps | — |
| 3 | Post | post.go | SetTimestamps | Version++ on UPDATE |
| 4 | PostTranslation | post.go | SetTimestamps | — |
| 5 | Comment | comment.go | SetTimestamps | NormalizeEmail (匿名评论) |
| 6 | MediaFile | media.go | SetTimestamps | — |
| 7 | Category | category.go | SetTimestamps | — |
| 8 | Config | config.go | SetUpdatedAt | — |
| 9 | SiteConfig | site_config.go | SetUpdatedAt | — |
| 10 | Role | role.go | SetTimestamps | — |
| 11 | APIEndpoint | api_endpoint.go | SetTimestamps | — |
| 12 | AdminMenu | admin_menu.go | SetTimestamps | — |
| 13 | RoleTemplate | role_template.go | SetTimestamps | — |
| 14 | UserTOTP | user_totp.go | SetTimestamps | — |
| 15 | SiteMenu | menu.go | SetTimestamps | — |
| 16 | SiteMenuItem | menu.go | SetTimestamps | — |
| 17 | Redirect | redirect.go | SetTimestamps | — |

### 不需要 hook 的 Model（13 个）

| Model | 原因 |
|-------|------|
| Audit | 只读追加，仅 created_at (DB DEFAULT) |
| PostRevision | 只读追加，仅 created_at |
| RefreshToken | 仅 created_at |
| PreviewToken | 仅 created_at |
| Tag | 仅 created_at |
| PasswordResetToken | 仅 created_at |
| UserRole | 仅 created_at（联合主键表） |
| RoleAPI | 无时间戳 |
| RoleMenu | 无时间戳 |
| PostCategoryMap | 无时间戳 |
| PostTagMap | 无时间戳 |
| RoleTemplateAPI | 无时间戳 |
| RoleTemplateMenu | 无时间戳 |

## 实现示例

### User（时间戳 + Email 标准化）

```go
func (u *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
    SetTimestamps(&u.CreatedAt, &u.UpdatedAt, query)
    NormalizeEmail(&u.Email)
    return nil
}
```

### Post（时间戳 + 乐观锁）

```go
func (p *Post) BeforeAppendModel(ctx context.Context, query bun.Query) error {
    SetTimestamps(&p.CreatedAt, &p.UpdatedAt, query)
    if _, ok := query.(*bun.UpdateQuery); ok {
        p.Version++
    }
    return nil
}
```

### Comment（时间戳 + 匿名评论 Email 标准化）

```go
func (c *Comment) BeforeAppendModel(ctx context.Context, query bun.Query) error {
    SetTimestamps(&c.CreatedAt, &c.UpdatedAt, query)
    if _, ok := query.(*bun.InsertQuery); ok && c.UserID == nil && c.AuthorEmail != "" {
        NormalizeEmail(&c.AuthorEmail)
    }
    return nil
}
```

### Config（仅 updated_at）

```go
func (c *Config) BeforeAppendModel(ctx context.Context, query bun.Query) error {
    SetUpdatedAt(&c.UpdatedAt, query)
    return nil
}
```

## 迁移文件关联变更

本设计与 `migration-update-design.md` 的触发器移除决策一致：

- 删除 `update_updated_at()` PL/pgSQL 函数
- 删除 public schema 所有 `CREATE TRIGGER trg_xxx_updated_at` 语句
- 删除 `template.go` 中所有 `CREATE TRIGGER` 语句（9 处）
- 保留 `DEFAULT NOW()` 作为 DB 兜底

## 修改文件清单

| 文件 | 动作 |
|------|------|
| `internal/model/hooks.go` | 新建 — 通用辅助函数 |
| `internal/model/user.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/site.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/post.go` | 修改 — 添加 BeforeAppendModel (Post + PostTranslation) |
| `internal/model/comment.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/media.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/category.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/config.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/site_config.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/role.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/api_endpoint.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/admin_menu.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/role_template.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/user_totp.go` | 修改 — 添加 BeforeAppendModel |
| `internal/model/menu.go` | 修改 — 添加 BeforeAppendModel (SiteMenu + SiteMenuItem) |
| `internal/model/redirect.go` | 修改 — 添加 BeforeAppendModel |

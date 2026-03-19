# Console 路由 + 布局 + Auth Guard 设计规范

**日期:** 2026-03-20
**状态:** 已批准
**范围:** Sub-project A — 路由基础设施（不含 6 个缺失 feature 模块的实现）

---

## 1. 概述

为 console/ 管理后台建立完整的路由架构、布局组件和认证守卫。现有 9 个 feature 模块（auth/posts/categories/tags/media/users/roles/sites/shared）已有 hooks + components，但无路由层和布局层。

---

## 2. API 路由认证分层

API 按前端消费者分组（与项目结构 `web/` + `console/` 对齐）：

```
/api/v1/
│
├── /auth/                    → 认证端点（独立组）
│   ├── POST /login           ○ 无需 JWT
│   ├── POST /login/totp      ○ 无需 JWT（用 2FA session token）
│   ├── POST /refresh         ○ 无需 JWT（用 httpOnly refresh cookie）
│   ├── POST /forgot-password ○ 无需 JWT
│   ├── POST /reset-password  ○ 无需 JWT（用 reset token）
│   ├── GET  /me              ● 需要 JWT
│   ├── POST /logout          ● 需要 JWT
│   └── POST /totp/setup      ● 需要 JWT
│
├── /web/                     → 公共站点 API（给 Templ+HTMX 消费，无需 JWT）
│   ├── GET /posts            ○ 可选 API Key 限流
│   ├── GET /posts/:slug      ○
│   ├── GET /categories       ○
│   ├── GET /tags             ○
│   ├── GET /search           ○
│   └── POST /comments        ○
│
└── /console/                 → 管理后台 API（给 Console SPA 消费，全部 JWT+RBAC）
    ├── /posts/*              ●
    ├── /categories/*         ●
    ├── /tags/*               ●
    ├── /media/*              ●
    ├── /comments/*           ●
    ├── /menus/*              ●
    ├── /redirects/*          ●
    ├── /users/*              ●
    ├── /roles/*              ●
    ├── /settings             ●
    ├── /api-keys/*           ●
    └── /audit                ●
```

### Chi 中间件映射

```go
r.Route("/api/v1", func(r chi.Router) {
    r.Route("/auth", func(r chi.Router) {
        r.Post("/login", ...)
        r.Post("/refresh", ...)
        r.Post("/forgot-password", ...)
        r.Post("/reset-password", ...)
        r.Group(func(r chi.Router) {
            r.Use(JWTAuth)
            r.Get("/me", ...)
            r.Post("/logout", ...)
            r.Post("/totp/setup", ...)
        })
    })

    r.Route("/web", func(r chi.Router) {
        r.Use(OptionalAPIKey)
        r.Use(RateLimit)
        // ... public endpoints
    })

    r.Route("/console", func(r chi.Router) {
        r.Use(JWTAuth)
        r.Use(RBAC)
        // ... admin endpoints
    })
})
```

---

## 3. TanStack Router 路由文件结构

使用嵌套路由组：`_auth` 组用 AuthLayout，`_dashboard` 组用 DashboardLayout。

```
console/src/routes/
├── __root.tsx                    # 根（QueryProvider + devtools）
├── index.tsx                     # / → 重定向到 /dashboard
├── _auth.tsx                     # AuthLayout（无侧边栏，居中卡片）
├── _auth/
│   ├── login.tsx                 # /login
│   ├── forgot-password.tsx       # /forgot-password
│   └── reset-password.tsx        # /reset-password
├── _dashboard.tsx                # DashboardLayout（sidebar + header + auth guard）
└── _dashboard/
    ├── index.tsx                 # /dashboard（仪表盘首页）
    ├── posts/
    │   ├── index.tsx             # /posts（列表）
    │   ├── $postId.edit.tsx      # /posts/:postId/edit
    │   └── new.tsx               # /posts/new
    ├── categories.tsx            # /categories
    ├── tags.tsx                  # /tags
    ├── media.tsx                 # /media
    ├── comments.tsx              # /comments
    ├── menus.tsx                 # /menus
    ├── redirects.tsx             # /redirects
    ├── users.tsx                 # /users
    ├── roles.tsx                 # /roles
    ├── settings.tsx              # /settings
    ├── api-keys.tsx              # /api-keys
    └── audit.tsx                 # /audit
```

---

## 4. Auth Guard

### `_dashboard.tsx` beforeLoad

```tsx
beforeLoad: async ({ context }) => {
  try {
    const user = await context.queryClient.ensureQueryData({
      queryKey: ['auth', 'me'],
      queryFn: () => apiClient.get('/api/v1/auth/me'),
      staleTime: 5 * 60 * 1000,
    })
    return { user }
  } catch {
    throw redirect({ to: '/login' })
  }
}
```

- `ensureQueryData` 首次调 API 验证 token，之后走 TanStack Query 缓存（5min staleTime）
- 401 → 重定向到 `/login`
- 返回 `{ user }` 注入路由 context，子路由通过 `useRouteContext()` 获取当前用户和权限

### Token 自动刷新

`api-client.ts` 拦截 401 响应：

```tsx
// 伪代码
if (response.status === 401 && !isRefreshing) {
  isRefreshing = true
  refreshPromise = apiClient.post('/api/v1/auth/refresh')
  const newToken = await refreshPromise
  isRefreshing = false
  // 用新 token 重发原请求
  return retry(originalRequest, newToken)
}
// 并发请求共享同一个 refreshPromise
if (response.status === 401 && isRefreshing) {
  await refreshPromise
  return retry(originalRequest)
}
// refresh 也失败 → 清除状态 → /login
```

---

## 5. 布局组件

### DashboardLayout

```
┌─────────────────────────────────────────────┐
│ ┌──────┐ ┌──────────────────────────────┐   │
│ │ Logo │ │ Header (面包屑 + 主题 + 用户) │   │
│ ├──────┤ ├──────────────────────────────┤   │
│ │      │ │                              │   │
│ │ Nav  │ │  <Outlet />                  │   │
│ │ 分组 │ │  (页面内容)                   │   │
│ │      │ │                              │   │
│ ├──────┤ │                              │   │
│ │ User │ │                              │   │
│ └──────┘ └──────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

- **Sidebar**: 固定 240px，shadcn `ScrollArea`
- **Header**: 64px，面包屑 + 主题切换 + 用户菜单
- **响应式**: `lg`（1024px）断点，小屏 sidebar → shadcn `Sheet` 侧滑抽屉

### Sidebar 导航分组

```
📊 Dashboard

── 内容管理 ──
📝 文章
📂 分类
🏷️ 标签
🖼️ 媒体
💬 评论

── 站点 ──
🔗 导航菜单
↪️ 重定向
⚙️ 站点设置

── 系统 ──
👥 用户
🛡️ 角色
🔑 API Keys
📋 审计日志

── 底部 ──
👤 用户名 + 登出
```

菜单项根据用户权限过滤（无权限的直接隐藏，不是 disabled）。

### AuthLayout

```
┌─────────────────────────────────┐
│              Logo               │
│        ┌──────────────┐         │
│        │  Login Form  │         │
│        │              │         │
│        └──────────────┘         │
│      (居中 Card，max-w-400px)    │
└─────────────────────────────────┘
```

---

## 6. RBAC 前端控制

### 数据流

```
/api/v1/auth/me → { user, permissions: ["posts.create", "posts.delete", "users.manage", ...] }
                      ↓
              路由 context.user
                      ↓
         ┌────────────┴────────────┐
         │                         │
    Sidebar 过滤              usePermission hook
    (隐藏无权限菜单)           (组件级控制)
```

### usePermission Hook

```tsx
function usePermission(permission: string): boolean {
  const { user } = useRouteContext({ from: '/_dashboard' })
  return user.permissions.includes(permission)
}

// 使用
const canCreatePost = usePermission('posts.create')
{canCreatePost && <Button>新建文章</Button>}
```

### 页面级权限守卫

```tsx
// _dashboard/users.tsx
export const Route = createFileRoute('/_dashboard/users')({
  beforeLoad: ({ context }) => {
    if (!context.user.permissions.includes('users.manage')) {
      throw redirect({ to: '/dashboard/403' })
    }
  },
  component: UsersPage,
})
```

---

## 7. 面包屑

从 TanStack Router `useMatches()` 自动生成。每个路由文件定义 `staticData`：

```tsx
// _dashboard/posts/index.tsx
export const Route = createFileRoute('/_dashboard/posts/')({
  staticData: { title: '文章' },
  component: PostsPage,
})
```

Header 组件遍历 `useMatches()` 提取 `staticData.title` 生成：

```
Dashboard > 文章 > 编辑
```

---

## 8. 暗色主题

- 三种模式：Light / Dark / System（跟随 OS）
- 切换按钮在 Header 右侧（太阳/月亮图标）
- `class` 策略：`<html class="dark">`，Tailwind V4 原生支持
- 偏好存 `localStorage("theme")`
- 使用已有的 `features/shared/components/ThemeProvider.tsx`

---

## 9. URL 状态管理

列表页的筛选、分页、排序持久化到 URL search params：

```tsx
// _dashboard/posts/index.tsx
const postsSearchSchema = z.object({
  page: z.number().default(1),
  perPage: z.number().default(20),
  status: z.enum(['all', 'draft', 'published', 'archived']).default('all'),
  search: z.string().optional(),
  sortBy: z.string().default('created_at'),
  sortOrder: z.enum(['asc', 'desc']).default('desc'),
})

export const Route = createFileRoute('/_dashboard/posts/')({
  validateSearch: postsSearchSchema,
  component: PostsPage,
})

// URL: /posts?page=2&status=published&search=hello
```

好处：可分享链接、浏览器前进后退保留状态。

---

## 10. 错误处理

### 错误页面

| 场景 | 处理 |
|------|------|
| 404 路由不存在 | `__root.tsx` 的 `notFoundComponent` → "页面不存在" + 返回 Dashboard |
| 403 权限不足 | `beforeLoad` throw redirect 到 403 页 → "权限不足，请联系管理员" |
| API 500 | Sonner `toast.error()` 报错，保持当前页面 |
| React crash | `ErrorBoundary` → "出错了" + 刷新按钮 |

### Toast 通知

- 使用 **Sonner**（shadcn 已集成）
- `toast.success("文章已发布")`、`toast.error("保存失败")`
- 位置：右下角，自动消失 4 秒

---

## 11. Loading 状态

### Suspense 骨架屏

`_dashboard.tsx` 用 `Suspense` 包裹 `<Outlet />`：

```tsx
<Suspense fallback={<PageSkeleton />}>
  <Outlet />
</Suspense>
```

### 骨架屏类型

| 页面类型 | Skeleton |
|----------|----------|
| 列表页 | 表头 + 5 行灰色占位条 |
| 编辑页 | 表单标签 + 输入框占位 |
| Dashboard | 4 个统计卡片占位 + 图表占位 |

---

## 12. 空状态

每个列表页有专属空状态：

```tsx
// 示例
<EmptyState
  icon={<FileText className="h-12 w-12" />}
  title="还没有文章"
  description="创建你的第一篇文章开始写作"
  action={<Button onClick={...}>新建文章</Button>}
/>
```

---

## 13. 移动端适配

| 断点 | Sidebar 行为 |
|------|-------------|
| `≥ 1024px`（lg） | 固定 sidebar 240px |
| `< 1024px` | sidebar 隐藏，Header 显示汉堡按钮 → 点击打开 shadcn `Sheet` 侧滑 |

---

## 14. 实现范围（Sub-project A）

本次只实现基础设施，不实现缺失的 feature 模块：

### 创建（新文件）

- `_auth.tsx` + `_auth/login.tsx` + `_auth/forgot-password.tsx` + `_auth/reset-password.tsx`
- `_dashboard.tsx`（DashboardLayout + Auth Guard + Suspense）
- `_dashboard/index.tsx`（Dashboard 首页占位）
- `_dashboard/posts/index.tsx`、`_dashboard/categories.tsx` 等（接入已有 feature 组件）
- `components/layouts/DashboardLayout.tsx`（Sidebar + Header）
- `components/layouts/AuthLayout.tsx`（居中卡片）
- `components/layouts/Sidebar.tsx`（导航分组 + RBAC 过滤）
- `components/layouts/Header.tsx`（面包屑 + 主题 + 用户菜单）
- `components/shared/PageSkeleton.tsx`（通用骨架屏）
- `components/shared/EmptyState.tsx`（通用空状态）
- `hooks/usePermission.ts`（RBAC hook）

### 修改

- `api-client.ts` — 添加 401 拦截 + token 自动刷新
- `__root.tsx` — 添加 `notFoundComponent`

### 不做（Sub-project B）

- comments/menus/redirects/settings/api-keys/audit/dashboard 的 feature 模块实现
- 这些路由页面先用占位组件 `<EmptyState title="即将推出" />`

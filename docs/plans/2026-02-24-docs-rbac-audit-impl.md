# 文档 RBAC 审计实现计划

> **For Claude:** 使用 Agent Teams 并行执行，每个 Agent 负责指定文档的更新。

**Goal:** 更新所有设计文档，使其反映动态 RBAC 系统的代码现实（9 张新表、全局角色、API-level 权限、两级缓存）。

**Architecture:** 5 个并行 Agent 各自负责 1-3 份文档，共享统一的变更规则清单。主 Agent 负责 CLAUDE.md 更新和最终交叉验证。

**Tech Stack:** Markdown 文档编辑、Mermaid 图表更新

---

## 全局变更规则（所有 Agent 必须遵守）

### 术语替换表

| 旧术语 | 新术语 | 说明 |
|--------|--------|------|
| `sfc_site_user_roles` | `sfc_user_roles` | 全局用户-角色分配表（无 site_id） |
| `user_role` ENUM | `sfc_roles` 表 | 动态角色定义 |
| `superadmin` (角色 slug) | `super` | 内置超级管理员角色 slug |
| `SuperAdmin` (UI/文档显示名) | `Super` 或 `超级管理员` | 保持可读性时可用显示名 |
| `RequireRole("superadmin")` | RBAC 中间件 + `sfc_role_apis` 动态匹配 | API-level 权限 |
| `per-site 角色` / `角色 per-site` | `全局角色` | 角色不再绑定站点 |
| `site:{slug}:role:{user_id}` | `rbac:user:{user_id}:slugs` (L2) + local cache (L1) | 两级缓存 |
| `四级角色控制` | `动态 RBAC 权限控制` | 不再是固定四级 |
| `SELECT role FROM sfc_site_user_roles WHERE site_id = $1 AND user_id = $2` | 通过 `sfc_user_roles` JOIN `sfc_roles` 查询 | 无 site_id |

### 新增 public schema 表（需在 database.md 添加完整 DDL）

1. `sfc_roles` (id, name, slug, description, built_in, status, created_at, updated_at)
2. `sfc_user_roles` (user_id, role_id, created_at) — PK(user_id, role_id)
3. `sfc_apis` (id, method, path, name, description, group, status, created_at, updated_at) — UNIQUE(method, path)
4. `sfc_role_apis` (role_id, api_id) — PK(role_id, api_id)
5. `sfc_menus` (id, parent_id, name, icon, path, sort_order, status, created_at, updated_at) — 后台管理菜单
6. `sfc_role_menus` (role_id, menu_id) — PK(role_id, menu_id)
7. `sfc_role_templates` (id, name, description, built_in, created_at, updated_at)
8. `sfc_role_template_apis` (template_id, api_id) — PK(template_id, api_id)
9. `sfc_role_template_menus` (template_id, menu_id) — PK(template_id, menu_id)

### 内置角色 seed 数据

```sql
INSERT INTO sfc_roles (name, slug, description, built_in, status) VALUES
('超级管理员', 'super', '拥有所有权限，不可修改/删除', true, true),
('管理员', 'admin', '站点管理，不可删除', true, true),
('编辑', 'editor', '内容创建与编辑，不可删除', true, true),
('查看者', 'viewer', '只读访问，不可删除', true, true);
```

### 新增代码模块

```
internal/rbac/
├── handler.go           # RBAC 管理 API handlers
├── service.go           # 两级 Redis 缓存 service
├── service_test.go      # Service 单元测试
├── api_registry.go      # 路由自动发现
├── api_registry_test.go # Registry 测试
├── interfaces.go        # Repository 接口定义
├── dto.go               # 请求/响应 DTO
├── role_repo.go         # 角色 CRUD
├── user_role_repo.go    # 用户-角色分配
├── api_repo.go          # API + RoleAPI CRUD
├── role_api_repo.go     # 角色-API 映射
├── menu_repo.go         # 菜单 + RoleMenu CRUD
├── template_repo.go     # 模板 CRUD
└── *_test.go            # 各 repo 测试桩
```

### RBAC 中间件工作流程

```
请求 → 提取 JWT user_id → 查询用户角色 slugs（两级缓存）
→ 若包含 "super" 则直接放行
→ 否则查询 sfc_role_apis 检查当前 method+path 是否被用户任一角色覆盖
→ 命中则放行，否则 403
```

### 迁移文件实际名称

```
migrations/
├── 20260224000001_create_enums_and_functions.go  # 枚举 + trigger 函数（无 user_role ENUM）
├── 20260224000002_create_public_schema.go        # 全部 public 表（含 9 张 RBAC 表）
├── 20260224000003_create_site_template.go        # 站点 schema DDL 模板
├── 20260224000004_seed_rbac_builtins.go          # Seed 4 内置角色 + 4 内置模板
└── main.go
```

---

## Task 1: 更新 database.md（Agent: agent-database）

**Files:** `docs/database.md`

**变更清单:**

1. **ER 图 (§1)**: 移除 `sfc_site_user_roles` 实体及其与 `sfc_sites`/`sfc_users` 的关系线。新增 9 个 RBAC 实体：`sfc_roles`, `sfc_user_roles`, `sfc_apis`, `sfc_role_apis`, `sfc_menus`(后台菜单), `sfc_role_menus`, `sfc_role_templates`, `sfc_role_template_apis`, `sfc_role_template_menus`。新增关系线。

2. **§2A DDL — 移除旧内容**:
   - 移除 `CREATE TYPE user_role AS ENUM (...)`
   - 移除 `sfc_site_user_roles` 的 CREATE TABLE / INDEX / TRIGGER（约 L427-L439）
   - 更新 L363 注释（不再引用 `sfc_site_user_roles`）

3. **§2A DDL — 新增 RBAC 表**:
   在 `sfc_sites` 之后、`sfc_refresh_tokens` 之前，添加 9 张 RBAC 表的完整 DDL（参照实际迁移文件 `20260224000002_create_public_schema.go`）

4. **§2.1 Seed 数据**:
   - 移除旧的 `sfc_site_user_roles` INSERT
   - 新增 `sfc_user_roles` seed（通过 role_id 引用 sfc_roles）
   - 更新 `superadmin` → `super` 的角色引用

5. **§3 Redis 键空间**:
   - 移除 `site:{slug}:role:{user_id}` 键
   - 新增 RBAC 缓存键：
     - `rbac:user:{user_id}:slugs` TTL=300s（L2 Redis 缓存）
     - `rbac:role:{role_slug}:apis` TTL=300s（角色 API 权限缓存）
     - `rbac:user:{user_id}:menus` TTL=300s（用户菜单缓存）
   - 说明 L1 本地缓存（sync.Map TTL=60s）

6. **§4 迁移管理**:
   - §4.3 迁移文件结构：更新为实际文件名
   - 移除 `sfc_site_user_roles` 引用
   - 新增 migration 4 (`seed_rbac_builtins`)

**验证:** 搜索文件确认不再包含 `sfc_site_user_roles`、`user_role.*ENUM`、`superadmin`（角色 slug 上下文）

---

## Task 2: 更新 architecture.md（Agent: agent-architecture）

**Files:** `docs/architecture.md`

**变更清单:**

1. **§1 架构图 Mermaid**:
   - `PGPublic` 节点：`sfc_site_user_roles` → `sfc_roles / sfc_user_roles / sfc_apis / sfc_role_apis / sfc_menus / sfc_role_menus / sfc_role_templates ...`
   - `AM["Auth Module"]` 描述更新：`RBAC` → `Dynamic RBAC`

2. **§1.1 Multi-Site Schema Isolation Mermaid**:
   - `AUTH` 节点：`Role from sfc_site_user_roles` → `Role from sfc_user_roles (global)`
   - 新增 RBAC 中间件节点

3. **Schema Layout (L98-L126)**:
   - `sfc_site_user_roles` → 替换为 `sfc_roles` + `sfc_user_roles` + 其他 7 张 RBAC 表
   - 添加注释说明 `sfc_menus` 是后台管理菜单（区别于 site schema 的 `sfc_site_menus`）

4. **§2 目录结构**:
   - 新增 `internal/rbac/` 模块（handler/service/api_registry/repo/interfaces/dto）
   - `internal/middleware/rbac.go` 描述更新

5. **§3.3 Multi-Site 决策表**:
   - 移除 `Per-site 角色（sfc_site_user_roles）` 行
   - 改为 `全局角色（sfc_user_roles）| 用户拥有全局角色，通过 sfc_role_apis 控制 API 访问权限`
   - 移除 `JWT Claims 不携带 role | 角色从 sfc_site_user_roles 每请求解析`
   - 改为 `JWT Claims 不携带 role | 角色从 sfc_user_roles 每请求解析（两级缓存：local 60s + Redis 300s）`

6. **§4.1 Site-Scoped 请求流程**:
   - 步骤 4 AuthMiddleware：`SELECT role FROM sfc_site_user_roles WHERE site_id AND user_id` → `从 sfc_user_roles JOIN sfc_roles 查询用户角色 slugs`
   - 新增步骤 5 RBAC Middleware：`根据 method+path 查 sfc_role_apis，super 角色直接放行`
   - Cache key 更新

7. **§5 路由注册代码**:
   - 所有 `RequireRole("superadmin")` → 描述为 "RBAC 中间件自动匹配（super 角色可访问）"
   - 更新路由注册代码注释，说明权限通过 `sfc_role_apis` 动态配置而非硬编码

**验证:** 搜索确认不再包含 `sfc_site_user_roles`、`RequireRole`、`superadmin`

---

## Task 3: 更新 security.md（Agent: agent-security）

**Files:** `docs/security.md`

**变更清单:**

1. **§1 安全架构概览 Mermaid**:
   - `RBAC["RBAC 权限\n四级角色控制\n(per-site)"]` → `RBAC["RBAC 权限\n动态角色控制\n(全局)"]`

2. **§2.1 隔离机制表**:
   - `SuperAdmin 跨站操作` → `Super 角色全局操作`

3. **§3.2 Token 生命周期** — 无需大改，仅确认 JWT 不含 role

4. **§3.5 注释** (L165):
   - `角色是 per-site 的，存储在 public.sfc_site_user_roles 表中` → `角色是全局的，存储在 public.sfc_user_roles 表中`

5. **§3.7 认证流程 Mermaid 图**:
   - L400: `SELECT role FROM public.sfc_site_user_roles` → `从 sfc_user_roles JOIN sfc_roles 查询角色 slugs`
   - L396-L407: 角色缓存键更新
   - L744: `创建 sfc_site_user_roles 条目` → `创建 sfc_user_roles 条目`

6. **§4.6 SuperAdmin 强制禁用**:
   - `SuperAdmin` → `Super`（角色 slug 上下文）
   - 保持文档可读性的地方可用"超级管理员"

7. **§5 RBAC 权限控制 — 全面重写**:
   - §5.1: 从 "Per-Site 角色模型" 改为 "全局角色模型 + 动态 API 权限"
     - 描述新的角色系统：`sfc_roles` 表定义角色，`sfc_user_roles` 分配，`sfc_role_apis` 控制 API 访问
     - 内置角色表（super/admin/editor/viewer）+ 自定义角色支持
   - §5.2: 角色解析策略完全重写
     - 两级缓存（L1 local sync.Map TTL=60s + L2 Redis TTL=300s）
     - 查询 `sfc_user_roles` 而非 `sfc_site_user_roles`
     - RBAC 中间件通过 method+path 查 `sfc_role_apis` 判断权限
   - §5.3: 权限矩阵
     - 说明权限现在通过 `sfc_role_apis` 动态配置
     - 内置角色的默认权限通过 `sfc_role_templates` seed
     - 保留示例矩阵但说明可通过管理界面调整
   - §5.4: 角色缓存安全更新
     - 缓存键更新
     - 两级缓存失效策略

8. **§8 安装安全** (L744):
   - `sfc_site_user_roles` → `sfc_user_roles`

9. **§16 安全检查清单**:
   - 更新 RBAC 相关检查项
   - 新增 API Registry 和 Role Template 检查项

**验证:** 搜索确认不含旧 RBAC 引用

---

## Task 4: 更新 api.md + prd.md（Agent: agent-api-prd）

**Files:** `docs/api.md`, `docs/prd.md`

### api.md 变更清单:

1. **§0 中间件链**:
   - L28: `从 sfc_site_user_roles 加载角色` → `从 sfc_user_roles 加载角色（两级缓存）`

2. **路由总表**:
   - 所有 `SuperAdmin` 权限标注 → `Super`
   - 新增 RBAC 管理 API 端点组：
     - `GET/POST /api/v1/rbac/roles` — 角色列表/创建
     - `GET/PUT/DELETE /api/v1/rbac/roles/:id` — 角色 CRUD
     - `GET/PUT /api/v1/rbac/roles/:id/apis` — 角色 API 权限
     - `GET/PUT /api/v1/rbac/roles/:id/menus` — 角色菜单可见性
     - `GET/POST /api/v1/rbac/users/:id/roles` — 用户角色分配
     - `GET/POST/PUT/DELETE /api/v1/rbac/menus/*` — 后台菜单管理
     - `GET /api/v1/rbac/apis` — API 端点列表
     - `GET/POST/PUT/DELETE /api/v1/rbac/templates/*` — 权限模板
     - `POST /api/v1/rbac/roles/:id/apply-template` — 应用模板

3. **§2 权限矩阵说明** (L246-268):
   - 完全重写：说明权限现在通过 `sfc_role_apis` 动态控制
   - 移除 `sfc_site_user_roles` 引用
   - 移除 `RequireRole` 引用

4. **§3 Auth API**:
   - L432: JWT Claims 注释更新
   - L482: `"role": "superadmin"` → 移除 role 字段或改为 roles 数组
   - L496: sites 数组说明更新

5. **§4.1 安装向导**:
   - L662: 执行流程中 `sfc_site_user_roles(superadmin)` → `sfc_user_roles(super)`

6. **§4.2 站点管理 API**:
   - 所有 `SuperAdmin` → `Super`
   - 用户角色分配端点（L875-L904）需适配新的 RBAC 模型
   - L884: role 校验从 ENUM 改为 sfc_roles 表查询
   - L749: 创建站点流程中角色分配更新

7. **§16 路由注册代码** (L3905+):
   - 所有 `RequireRole(...)` → 说明为 RBAC 中间件自动处理
   - L4166/4172/4174: 角色解析代码更新

### prd.md 变更清单:

1. **§3.1.1**: `SuperAdmin / Admin / Editor / Viewer` → 说明为动态角色系统，内置 4 个角色
2. **§3.1.11**: `SuperAdmin` → `Super` (角色 slug)
3. **§3.1.12 权限矩阵**: `SuperAdmin` 列标题 → `Super`

**验证:** 搜索确认不含旧引用

---

## Task 5: 更新 story.md + standard.md + testing.md（Agent: agent-misc）

**Files:** `docs/story.md`, `docs/standard.md`, `docs/testing.md`

### story.md 变更清单:

1. L42: 角色列表 `SuperAdmin / Admin / Editor / Viewer` → `Super / Admin / Editor / Viewer（内置角色，可通过 RBAC 管理扩展自定义角色）`
2. L61: `Admin 访问 SuperAdmin 专属功能` → `Admin 访问 Super 专属功能`
3. L74: 权限矩阵 `SuperAdmin` 列 → `Super`
4. L425/441/484: `As a SuperAdmin` → `As a Super`（或保持"超级管理员"）
5. L495: `per-site 分配` → `全局分配`
6. L495: `sfc_site_user_roles` → `sfc_user_roles`
7. L498: JWT claims 说明 + `sfc_site_user_roles` → `sfc_user_roles`
8. L524: 安装流程 `superadmin 角色` → `super 角色`
9. L731/741: SuperAdmin 相关引用更新

### standard.md 变更清单:

1. L278: `c.GetString("role")` 注释 `per-site 角色` → `全局角色`
2. L480: `sfc_site_user_roles` → `sfc_user_roles`
3. L497: SQL 查询更新
4. L590/595/600: `RequireRole(...)` → 说明为 RBAC 中间件

### testing.md 变更清单:

1. L299/302/311: RequireRole 测试用例 → RBAC 中间件测试
2. L605/617-626: `superadmin` 角色引用 → `super`
3. L1072/1073/1081: SuperAdmin 引用更新
4. L1139: `sfc_site_user_roles` → `sfc_user_roles`
5. L1150: 初始化测试 `superadmin` → `super`
6. L1214: SuperAdmin 强制禁用测试更新

**验证:** 搜索确认不含旧引用

---

## Task 6: 更新 CLAUDE.md + setup.md + deployment.md（主 Agent）

**Files:** `CLAUDE.md`, `docs/setup.md`, `docs/deployment.md`

### CLAUDE.md 变更:

1. 多站点架构部分：`sfc_site_user_roles` → `sfc_user_roles`（全局）+ 9 张 RBAC 表
2. 中间件链：`Auth → RBAC` 描述更新
3. public schema 表列表更新

### setup.md 变更:

1. L336: `SuperAdmin Token` → `Super Token`

### deployment.md 变更:

1. L188: `sfc_site_user_roles` → RBAC 表
2. L190: `superadmin 角色` → `super 角色`
3. L953: 迁移文件名更新

---

## Task 7: 交叉验证（主 Agent）

**步骤:**

1. 全文搜索所有 docs/ 文件，确认无残留旧引用：
   - `sfc_site_user_roles`
   - `user_role.*ENUM`
   - `superadmin`（角色 slug 上下文，非用户故事中的显示名）
   - `RequireRole`（硬编码权限调用）
   - `per-site.*角色` / `角色.*per-site`

2. 检查文档间一致性：
   - database.md 的 DDL 与 architecture.md 的 Schema Layout 一致
   - security.md 的权限矩阵与 api.md 的路由权限标注一致
   - story.md 的角色引用与 prd.md 的权限矩阵一致

3. 验证新增内容：
   - 9 张 RBAC 表在 database.md 有 DDL
   - `internal/rbac/` 在 architecture.md 目录结构中出现
   - RBAC 管理 API 在 api.md 路由总表中出现

---

## Task 8: 提交

```bash
git add docs/
git commit -m "docs: update all design docs for dynamic RBAC system

- Replace sfc_site_user_roles with 9 RBAC tables across all docs
- Update role model from per-site ENUM to global dynamic roles
- Rename superadmin role slug to super
- Add internal/rbac/ module to architecture docs
- Update Redis cache keys for two-level RBAC caching
- Add RBAC management API endpoints to api.md
- Update migration file names to match actual files"
```

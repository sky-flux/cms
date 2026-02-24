# 文档审计设计：RBAC 系统重构后文档一致性更新

**日期**：2026-02-24
**状态**：已批准

---

## 背景

Dynamic RBAC 实现完成后（9 张新表 + internal/rbac/ 模块），所有 10 份设计文档仍描述旧的 `sfc_site_user_roles` + `user_role` ENUM 模型。需要全面更新文档以反映代码现实。

## 核心变更清单

### RBAC 架构变更

| 维度 | 旧 | 新 |
|------|---|---|
| 角色存储 | `sfc_site_user_roles` (site_id, user_id, role ENUM) | `sfc_roles` + `sfc_user_roles` (user_id, role_id) |
| 角色类型 | `user_role` ENUM (superadmin/admin/editor/viewer) | 动态 `sfc_roles` 表，built-in slugs: super/admin/editor/viewer |
| 角色范围 | Per-site | 全局 |
| SuperAdmin slug | `superadmin` | `super` |
| 权限控制 | 硬编码 `RequireRole()` | API-level: `sfc_role_apis` 映射 |
| 后台菜单 | 无独立表 | `sfc_menus` + `sfc_role_menus` (角色可见性) |
| 权限模板 | 无 | `sfc_role_templates` + `sfc_role_template_apis` + `sfc_role_template_menus` |
| 缓存策略 | Redis `site:{slug}:role:{user_id}` TTL=300s | 两级缓存 (local sync.Map TTL=60s + Redis TTL=300s) |
| API 注册 | 无 | `sfc_apis` + 路由自动发现 (APIRegistry) |

### 新增 public schema 表（共 9 张）

1. `sfc_roles` — 角色定义
2. `sfc_user_roles` — 用户-角色分配
3. `sfc_apis` — API 端点注册
4. `sfc_role_apis` — 角色-API 权限映射
5. `sfc_menus` — 后台管理菜单
6. `sfc_role_menus` — 角色-菜单可见性
7. `sfc_role_templates` — 权限模板
8. `sfc_role_template_apis` — 模板-API 映射
9. `sfc_role_template_menus` — 模板-菜单映射

### 移除

- `user_role` ENUM 类型
- `sfc_site_user_roles` 表
- 所有 per-site 角色相关描述

### 新增代码模块

- `internal/rbac/` — handler, service (两级缓存), api_registry (路由自动发现), 6 个 repository, interfaces, dto
- `internal/middleware/rbac.go` — 基于 API-level 权限的中间件

### 迁移文件修正

- 文档中的 `20260201000001_create_extensions.go` → 实际 `20260224000001_create_enums_and_functions.go`
- 新增迁移 4: `20260224000004_seed_rbac_builtins.go`

## 审计方案：按文档分片并行

### Agent 分工

| Agent | 文档 | 审计重点 |
|-------|------|---------|
| agent-database | database.md | 移除旧 RBAC 表/ENUM，新增 9 张 RBAC 表 DDL + ER 图，Redis 键空间 |
| agent-architecture | architecture.md | 目录结构(+rbac/)，中间件链，路由注册，Schema Layout，角色解析 |
| agent-security | security.md | §5 RBAC 重写，权限矩阵，认证流程图，缓存安全 |
| agent-api-prd | api.md + prd.md | RBAC 管理 API，权限标注，角色描述 |
| agent-misc | story.md + standard.md + testing.md | 角色引用，编码规范，测试策略 |
| 主 Agent | CLAUDE.md + setup.md + deployment.md + 交叉验证 | 项目指令，环境配置，最终一致性 |

### 修改原则

1. **代码为准**：文档必须反映已实现的代码现实
2. **一致性优先**：所有文档使用相同的术语和表名
3. **最小修改**：仅修改与 RBAC 变更相关的内容，不做无关重构
4. **保留格式**：保持各文档原有的格式风格和章节结构

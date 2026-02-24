# Go 单元测试全覆盖设计

**日期**: 2026-02-24
**状态**: 已批准
**范围**: internal/ 下所有 Go 包的单元测试 + testcontainers 集成测试

---

## 1. 当前状态

### 已有测试 (4 包, 885 行)

| 包 | 测试文件 | 状态 |
|----|---------|------|
| `internal/model` | hooks_test.go (295 行) | EXCELLENT — 全覆盖 |
| `internal/middleware` | rbac_test.go (99 行) | GOOD — 仅 RBAC 中间件 |
| `internal/router` | router_test.go (51 行) | STUB — 硬编码假 handler |
| `internal/rbac` | 7 个 test 文件 (440 行) | PARTIAL — service 好, repo 全 t.Skip |

### 待测试 — 有实际逻辑 (5 包, ~41 导出函数)

| 包 | 函数数 | 依赖 |
|----|--------|------|
| `pkg/apperror` | 7 函数 + 8 哨兵错误 | 无 |
| `pkg/response` | 5 函数 | Gin httptest |
| `config` | Load + DSN + Addr | Viper |
| `schema` | 3 函数 | testcontainers PG |
| `database` | 4 连接工厂 | testcontainers PG/Redis/Meili/RustFS |

### 待测试 — 已有测试需补全

| 包 | 缺口 |
|----|------|
| `middleware` | 7/8 未测 (CORS/Logger/Recovery/RequestID 已实现; Auth/InstallGuard/Schema/SiteResolver 是 stub) |
| `rbac` handler | 17 HTTP handler 完全未测 |
| `rbac` repo | 33 函数全部 t.Skip |
| `rbac` service | 缺 4 个边界场景 |
| `router` | healthHandler 需重写, Setup 未测 |

### Stub 包 (16 个) — 无逻辑

apikey, audit, auth, category, comment, cron, feed, media, menu, post, preview, redirect, setup, site, system, tag, user

---

## 2. 方案: 自底向上 + 共享 testcontainers 基础设施

### Phase 1: 测试基础设施 — `internal/testutil/`

#### containers.go

```go
type SharedContainers struct {
    PG    *bun.DB
    Redis *redis.Client
}

func SetupContainers(m *testing.M) *SharedContainers
func (sc *SharedContainers) Teardown()
```

- 每个需 DB 的包在 TestMain 中调用一次
- PG 容器 `postgres:18-alpine`, 启动后执行 migrations/
- Redis 用 miniredis (无需 Docker, 与 RBAC service 测试一致)
- Meilisearch/RustFS 仅在 database 包测试中按需启动

#### httptest.go

```go
func NewTestRouter() *gin.Engine
func DoRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder
func DoRequestWithAuth(router *gin.Engine, method, path string, body interface{}, userID string) *httptest.ResponseRecorder
```

### Phase 2: 纯逻辑层 (~20 测试函数)

#### pkg/apperror/errors_test.go

| 测试函数 | 覆盖 |
|----------|------|
| TestAppError_Error | 有/无内部 err |
| TestAppError_Unwrap | errors.Is / errors.As 链式匹配 |
| TestHTTPStatusCode | 8 哨兵 → HTTP 状态码 + 未知 → 500 |
| TestConstructors | 6 构造函数: Code + Message + errors.Is 追溯 |

关键: `errors.Join(ErrNotFound, originalErr)` 后, errors.Is 对两个错误都为 true。

#### pkg/response/response_test.go

| 测试函数 | 覆盖 |
|----------|------|
| TestSuccess | 200 + `{"success":true,"data":...}` |
| TestCreated | 201 |
| TestNoContent | 204 + 空 body |
| TestError_AppError | appErr.Code + appErr.Message |
| TestError_SentinelError | HTTPStatusCode 映射 |
| TestError_GenericError | 500 |
| TestPaginated | 200 + meta{total,page,per_page} |

#### config/config_test.go

| 测试函数 | 覆盖 |
|----------|------|
| TestLoad_Defaults | 无 .env 时默认值 |
| TestLoad_FromEnvFile | 临时 .env 读取 |
| TestLoad_EnvVarOverride | 环境变量覆盖 |
| TestLoad_InvalidDuration | 错误 duration → error |
| TestDBConfig_DSN | DSN 格式 |
| TestRedisConfig_Addr | Addr 格式 |

注意: 每测试用 t.Setenv() + viper.Reset() 隔离全局状态。

### Phase 3: 基础设施层 (~15 测试函数)

#### database/*_test.go

| 测试函数 | 覆盖 |
|----------|------|
| TestNewPostgres_Success | testcontainer PG18 → Ping OK |
| TestNewPostgres_InvalidDSN | 错误连接 → error |
| TestNewPostgres_PoolConfig | MaxOpenConns 等配置应用 |
| TestNewRedis_Success | miniredis → Ping OK |
| TestNewRedis_InvalidAddr | 错误地址 → error |
| TestNewMeilisearch_Success | testcontainer Meili → Health OK |
| TestNewMeilisearch_InvalidURL | 错误 URL → error |
| TestNewRustFS_Success | testcontainer RustFS → bucket 创建 |
| TestNewRustFS_InvalidEndpoint | 错误 endpoint → error |

#### schema/schema_test.go

| 测试函数 | 覆盖 |
|----------|------|
| TestValidateSlug_Valid | "blog", "my_site_01", "abc" (3 字符边界) |
| TestValidateSlug_Invalid | 大写/太短/含连字符/空/51字符 |
| TestCreateSiteSchema_Success | 17 张表 + 3 分区 + 索引 |
| TestCreateSiteSchema_InvalidSlug | 非法 slug → error |
| TestCreateSiteSchema_Idempotent | 重复创建 → IF NOT EXISTS |
| TestDropSiteSchema_Success | 创建后删除 |
| TestDropSiteSchema_NonExistent | IF EXISTS |
| TestDropSiteSchema_InvalidSlug | error |
| TestCreateAuditPartitions | 3 月分区名 + 索引 |

前置: TestMain 执行 public schema migrations (site 表有 FK 到 sfc_users)。

### Phase 4: 中间件层 (~12 测试函数)

#### cors_test.go

| 测试函数 | 覆盖 |
|----------|------|
| TestCORS_AllowedOrigin | 匹配 → 5 个 CORS 头 |
| TestCORS_DisallowedOrigin | 不匹配 → 无 CORS 头 |
| TestCORS_MultipleOrigins | 逗号分隔多 origin |
| TestCORS_Preflight | OPTIONS → 204 |
| TestCORS_EmptyOrigin | 无 Origin → 放行 |

#### request_id_test.go

| 测试函数 | 覆盖 |
|----------|------|
| TestRequestID_Generated | 无 header → 生成 UUID |
| TestRequestID_Preserved | 有 header → 保留 |

#### logger_test.go

| 测试函数 | 覆盖 |
|----------|------|
| TestLogger_LogsRequest | 自定义 slog.Handler 捕获 → 验证字段 |

#### recovery_test.go

| 测试函数 | 覆盖 |
|----------|------|
| TestRecovery_PanicCaught | panic → 500 + JSON |
| TestRecovery_NoPanic | 正常 → 不触发 |

#### Stub 中间件占位

auth_test.go / installation_guard_test.go / schema_test.go / site_resolver_test.go — package 声明 + 注释, 无 Test 函数。

### Phase 5: RBAC 全面覆盖 (~55 测试函数)

#### 5.1 handler_test.go (新文件, ~18 测试)

Mock 6 个 Repository 接口 + Service, 测试 HTTP 行为:

| 测试函数 | 覆盖 |
|----------|------|
| TestListRoles_Success | GET → 200 |
| TestListRoles_Error | repo error → 500 |
| TestCreateRole_Success | POST → 201 |
| TestCreateRole_InvalidJSON | 格式错 → 422 |
| TestUpdateRole_Success | PUT → 200 |
| TestUpdateRole_SuperProtected | super → 403 |
| TestUpdateRole_NotFound | 不存在 → 404 |
| TestDeleteRole_Success | 非内置 → 204 |
| TestDeleteRole_BuiltInProtected | 内置 → 403 |
| TestSetRoleAPIs_Success | PUT → 204 + cache 失效 |
| TestSetRoleAPIs_InvalidJSON | 格式错 → 422 |
| TestSetRoleMenus_Success | PUT → 204 + cache 失效 |
| TestCreateTemplate_Success | POST → 201 |
| TestDeleteTemplate_BuiltIn | 内置 → 403 |
| TestGetUserRoles_Success | GET → 200 |
| TestSetUserRoles_Success | PUT → 204 + user cache 失效 |
| TestGetMyMenus_Success | context user_id → 菜单树 |
| TestListAPIs_Success | GET → 200 |

#### 5.2 Repo 测试 — 替换 t.Skip (~33 测试)

6 个 repo 共享 TestMain 的 testcontainers PG:
- api_repo: UpsertBatch / DisableStale / List / ListByGroup / GetByMethodPath
- role_repo: List / GetByID / GetBySlug / Create / Update / Delete (built_in 保护)
- role_api_repo: GetAPIsByRoleID / SetRoleAPIs / GetRoleIDsByMethodPath / CloneFromTemplate
- menu_repo: ListTree / Create / Update / Delete / GetMenusByRoleID / SetRoleMenus / GetMenusByUserID
- template_repo: List / GetByID / Create / Update / Delete / Get/SetTemplateAPIs / Get/SetTemplateMenus
- user_role_repo: GetRolesByUserID / GetRoleSlugs / SetUserRoles / HasRole

#### 5.3 Service 补充 (4 个新测试)

| 测试函数 | 覆盖 |
|----------|------|
| TestCheckPermission_GetRoleSlugsError | getUserRoles 失败 → error |
| TestCheckPermission_GetRoleAPISetError | API 查询失败 → error |
| TestCheckPermission_MultipleRoles | 多角色权限合并 |
| TestCheckPermission_EmptyRoles | 无角色 → denied |

### Phase 6: Router + Stub 占位 (~6 测试 + 16 文件)

#### router_test.go 重写

| 测试函数 | 覆盖 |
|----------|------|
| TestHealthHandler_AllHealthy | 200 + ok |
| TestHealthHandler_DBDown | 503 + degraded |
| TestHealthHandler_RedisDown | 503 + degraded |
| TestHealthHandler_MeiliDown | 503 + degraded |
| TestHealthHandler_RustFSNil | 503 + degraded |
| TestSetup_MiddlewareRegistered | /health 路由存在 |

healthHandler 未导出, 测试文件需 `package router`。
依赖 mock: sqlmock / miniredis / mock ServiceManager / nil s3Client。

#### 16 个 Stub 包占位文件

每包 1 个文件, package 声明 + 注释, 无 Test 函数:

```
internal/{apikey,audit,auth,category,comment,cron,feed,media,
menu,post,preview,redirect,setup,site,system,tag,user}/handler_test.go
```

---

## 3. 总量估算

| Phase | 新增测试函数 | 新增文件 |
|-------|-------------|---------|
| 1. testutil 基础设施 | 0 | 2 (containers.go, httptest.go) |
| 2. 纯逻辑层 | ~20 | 3 |
| 3. 基础设施层 | ~15 | 5 |
| 4. 中间件层 | ~12 | 4 + 4 stub |
| 5. RBAC 全覆盖 | ~55 | 1 新 + 6 重写 |
| 6. Router + Stub | ~6 | 1 重写 + 16 stub |
| **合计** | **~108** | **~42** |

## 4. 测试工具链

| 工具 | 用途 |
|------|------|
| `testify/assert` + `testify/require` | 断言 |
| `net/http/httptest` | HTTP handler 测试 |
| `testcontainers-go` | PG18 / Meilisearch / RustFS 容器 |
| `miniredis` | Redis 内存模拟 |
| 手写 mock struct | Repository 接口 mock (与现有 service_test.go 一致) |
| `gin.CreateTestContextOnly` | 中间件隔离测试 |
| `t.Setenv()` + `viper.Reset()` | 配置测试隔离 |

## 5. 约束

- testcontainers 需要 Docker (项目已有 colima 环境)
- RBAC repo 测试需先执行 migrations (public schema + seed 数据)
- Viper 全局状态需在测试间 Reset
- CI 中 testcontainers 需 Docker-in-Docker 或 service containers

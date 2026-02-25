# CMS 内容管理系统 — 测试策略与用例文档

**版本**：v2.0
**日期**：2026-02-24
**状态**：草稿

---

## 1. 测试策略概述

### 1.1 测试理念

采用分层全覆盖策略，确保每一层代码都有对应的测试保障：

| 测试层级 | 覆盖目标 | 关注点 |
|----------|----------|--------|
| **单元测试** | 覆盖所有代码（每个函数、每个分支） | 单个函数/方法的逻辑正确性 |
| **集成测试** | 覆盖所有关键业务路径 | 多组件交互的正确性（Handler↔Service↔Repository↔DB） |
| **E2E 测试** | 覆盖所有用户关键操作路径 | 从 UI 到数据库的完整链路正确性 |
| **性能测试** | 验证非功能需求指标 | P99 响应时间、并发承载能力 |

### 1.2 覆盖率目标

| 层级 | 覆盖率目标 | 说明 |
|------|-----------|------|
| Go Service 层 | ≥ 80%（逐步提升至 90%） | 核心业务逻辑全覆盖 |
| Go Handler/Middleware | ≥ 70% | 关键路径覆盖 |
| Go 工具包（pkg） | ≥ 90% | 通用组件高覆盖 |
| 前端 React 组件 | ≥ 70% | 分支覆盖 |
| 前端工具函数 | ≥ 90% | 纯函数全覆盖 |

> **说明**：CI 流水线中后端整体覆盖率阈值设为 80%，低于此阈值将阻断合并。各层级目标与 standard.md 的要求保持一致。

### 1.3 测试工具链

```
后端（Go）
├── 测试框架：go test（标准库）
├── 断言库：testify/assert + testify/require
├── Mock：testify/mock（接口 Mock）
├── HTTP 测试：net/http/httptest
├── 数据库测试：testcontainers-go（真实 PostgreSQL 18 容器）
├── Redis 测试：miniredis（内存 Redis 模拟）
└── 性能测试：k6（HTTP 压测）

前端（TypeScript）
├── 单元/组件测试：Vitest + React Testing Library
├── E2E 测试：Playwright
├── Mock：MSW（Mock Service Worker）
└── 覆盖率：Vitest coverage（istanbul）
```

### 1.4 测试架构分层

```mermaid
flowchart TB
    subgraph E2E["E2E 测试（Playwright）"]
        E1["用户登录→创建文章→发布→公开API查询"]
        E2["媒体上传→文章引用→删除保护"]
        E3["角色权限全链路验证"]
        E4["安装向导→多站点→跨站隔离"]
    end

    subgraph Integration["集成测试"]
        I1["Handler + Service + Repository + DB"]
        I2["认证链路：登录→JWT→刷新→登出→黑名单"]
        I3["缓存链路：写入→缓存失效→重建"]
        I4["文章生命周期：草稿→发布→下架→删除→恢复"]
        I5["Schema 隔离：多站点数据互不可见"]
        I6["评论/菜单/RSS/重定向/预览/2FA"]
    end

    subgraph Unit["单元测试"]
        U1["Handler 层：HTTP 请求/响应"]
        U2["Service 层：业务逻辑"]
        U3["Repository 层：SQL 查询"]
        U4["Middleware 层：JWT/RBAC/Schema/限流"]
        U5["工具包：slug/crypto/paginator/totp"]
    end

    E2E --> Integration --> Unit
```

---

## 2. Go 后端单元测试

### 2.1 测试目录结构

```
sky-flux-cms/
├── internal/
│   ├── handler/
│   │   ├── auth.go
│   │   └── auth_test.go          # Handler 单测
│   ├── service/
│   │   ├── auth_service.go
│   │   └── auth_service_test.go  # Service 单测
│   ├── repository/
│   │   ├── user_repo.go
│   │   └── user_repo_test.go     # Repository 集成测试（testcontainers）
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── schema.go
│   │   └── auth_test.go          # Middleware 单测
│   └── pkg/
│       ├── jwt/
│       │   └── jwt_test.go
│       ├── crypto/
│       │   └── crypto_test.go
│       ├── slug/
│       │   └── slug_test.go
│       └── totp/
│           └── totp_test.go
└── tests/
    └── integration/              # 跨层集成测试
        ├── auth_flow_test.go
        ├── post_lifecycle_test.go
        ├── cache_invalidation_test.go
        ├── schema_isolation_test.go
        ├── comment_flow_test.go
        ├── menu_flow_test.go
        ├── redirect_flow_test.go
        ├── preview_token_test.go
        └── totp_flow_test.go
```

### 2.2 Handler 层测试规范

Handler 测试使用 `httptest` 构建请求，Mock Service 层接口：

```go
func TestLoginHandler_Success(t *testing.T) {
    // Arrange: Mock AuthService
    mockService := new(mocks.AuthService)
    mockService.On("Login", mock.Anything, "admin@example.com", "SecurePass123").
        Return(&dto.LoginResponse{
            AccessToken: "mock-token",
            ExpiresIn:   900,
            User:        mockUser,
        }, nil)

    handler := NewAuthHandler(mockService)

    // Act: 构建 HTTP 请求
    body := `{"email":"admin@example.com","password":"SecurePass123"}`
    req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router := gin.New()
    router.POST("/api/v1/auth/login", handler.Login)
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)

    var resp dto.Response
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.True(t, resp.Success)

    mockService.AssertExpectations(t)
}
```

### 2.3 Service 层测试规范

Service 测试 Mock Repository 接口，专注业务逻辑验证：

```go
func TestPostService_Publish(t *testing.T) {
    mockRepo := new(mocks.PostRepository)
    mockCache := new(mocks.CacheService)
    service := NewPostService(mockRepo, mockCache)

    t.Run("发布草稿文章成功", func(t *testing.T) {
        post := &model.Post{ID: testUUID, Status: model.StatusDraft}
        mockRepo.On("FindByID", mock.Anything, testUUID).Return(post, nil)
        mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
        mockCache.On("InvalidatePost", mock.Anything, testUUID).Return(nil)

        err := service.Publish(context.Background(), testUUID)

        require.NoError(t, err)
        assert.Equal(t, model.StatusPublished, post.Status)
        assert.NotNil(t, post.PublishedAt)
    })

    t.Run("已发布文章重复发布返回错误", func(t *testing.T) {
        post := &model.Post{ID: testUUID, Status: model.StatusPublished}
        mockRepo.On("FindByID", mock.Anything, testUUID).Return(post, nil)

        err := service.Publish(context.Background(), testUUID)

        assert.ErrorIs(t, err, ErrAlreadyPublished)
    })
}
```

### 2.4 Repository 层测试规范

Repository 测试使用 `testcontainers-go` 启动真实 PostgreSQL 实例，验证 SQL 正确性：

```go
func TestPostRepository_FindAll(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t) // testcontainers-go 启动 PG 18
    defer db.Close()

    repo := NewPostRepository(db)

    // 准备测试数据
    seedPosts(t, db, 25)

    t.Run("分页查询第一页", func(t *testing.T) {
        filter := PostFilter{Page: 1, PerPage: 10, Status: "published"}
        posts, total, err := repo.FindAll(ctx, filter)

        require.NoError(t, err)
        assert.Len(t, posts, 10)
        assert.Equal(t, 25, total)
    })

    t.Run("全文检索", func(t *testing.T) {
        filter := PostFilter{Query: "性能优化"}
        posts, _, err := repo.FindAll(ctx, filter)

        require.NoError(t, err)
        assert.NotEmpty(t, posts)
    })

    t.Run("软删除文章不出现在默认查询", func(t *testing.T) {
        // soft delete one post
        repo.SoftDelete(ctx, testPostID)

        filter := PostFilter{IncludeDeleted: false}
        posts, _, _ := repo.FindAll(ctx, filter)

        for _, p := range posts {
            assert.Nil(t, p.DeletedAt)
        }
    })
}
```

### 2.5 Middleware 测试规范

```go
func TestJWTMiddleware(t *testing.T) {
    t.Run("有效 Token 放行", func(t *testing.T) {
        token := generateTestJWT(t, "user-id", "editor")

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest("GET", "/", nil)
        c.Request.Header.Set("Authorization", "Bearer "+token)

        JWTMiddleware()(c)

        assert.False(t, c.IsAborted())
        assert.Equal(t, "user-id", c.GetString("user_id"))
    })

    t.Run("无 Token 返回 401", func(t *testing.T) {
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest("GET", "/", nil)

        JWTMiddleware()(c)

        assert.True(t, c.IsAborted())
        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })

    t.Run("黑名单 Token 返回 401", func(t *testing.T) {
        token := generateTestJWT(t, "user-id", "editor")
        // 将 token 加入 Redis 黑名单
        addToBlacklist(t, token)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest("GET", "/", nil)
        c.Request.Header.Set("Authorization", "Bearer "+token)

        JWTMiddleware()(c)

        assert.True(t, c.IsAborted())
        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })
}

func TestRBACMiddleware(t *testing.T) {
    // RBAC 中间件通过 sfc_role_apis 动态匹配 API 权限，
    // 不再使用硬编码的 RequireRole(...) 检查
    tests := []struct {
        name       string
        userRole   string
        method     string
        path       string
        shouldPass bool
    }{
        {"super 访问用户管理接口", "super", "GET", "/api/v1/users", true},
        {"admin 访问用户管理接口", "admin", "GET", "/api/v1/users", false},
        {"editor 访问文章接口", "editor", "POST", "/api/v1/posts", true},
        {"viewer 访问文章接口", "viewer", "POST", "/api/v1/posts", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := httptest.NewRecorder()
            c, _ := gin.CreateTestContext(w)
            c.Set("user_role", tt.userRole)
            c.Request = httptest.NewRequest(tt.method, tt.path, nil)

            // RBAC 中间件根据 sfc_role_apis 表动态匹配
            RBACMiddleware()(c)

            assert.Equal(t, !tt.shouldPass, c.IsAborted())
        })
    }
}
```

### 2.6 工具包测试

```go
// slug 生成测试
func TestSlugGenerate(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"Go 性能优化实践", "go-xing-neng-you-hua-shi-jian"},
        {"Hello World", "hello-world"},
        {"  空格  测试  ", "kong-ge-ce-shi"},
        {"Special!@#Characters", "specialcharacters"},
        {"", ""},
    }
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := slug.Generate(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}

// 密码哈希测试
func TestPasswordHash(t *testing.T) {
    password := "SecurePass123"

    hash, err := crypto.HashPassword(password)
    require.NoError(t, err)
    assert.NotEqual(t, password, hash)

    assert.True(t, crypto.CheckPassword(password, hash))
    assert.False(t, crypto.CheckPassword("WrongPass", hash))
}
```

---

## 3. Go 后端集成测试

### 3.1 集成测试环境

集成测试使用 `testcontainers-go` 启动完整的依赖服务。测试环境在 PostgreSQL 容器中创建 `public` schema 和测试用的 `site_{slug}` schema，确保 schema 隔离逻辑与生产一致。

```go
// tests/integration/setup.go
type TestEnv struct {
    DB    *bun.DB
    Redis *redis.Client
    Router *gin.Engine
}

func SetupTestEnv(t *testing.T) *TestEnv {
    t.Helper()

    // 启动 PostgreSQL 容器
    pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:18-alpine",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_DB":       "cms_test",
                "POSTGRES_USER":     "test",
                "POSTGRES_PASSWORD": "test",
            },
            WaitingFor: wait.ForListeningPort("5432/tcp"),
        },
        Started: true,
    })
    require.NoError(t, err)
    t.Cleanup(func() { pgContainer.Terminate(ctx) })

    // 启动 Redis 容器
    redisContainer := setupRedisContainer(t)

    // 执行数据库迁移（public schema + test site schema）
    runMigrations(t, db)

    // 创建默认测试站点 schema
    CreateTestSite(t, db, "test_site")

    // 构建完整的路由（Handler + Service + Repository）
    router := setupRouter(db, redisClient)

    return &TestEnv{DB: db, Redis: redisClient, Router: router}
}

// CreateTestSite 创建测试用站点 schema，包含所有站点级别表
func CreateTestSite(t *testing.T, db *bun.DB, slug string) {
    t.Helper()
    ctx := context.Background()
    // 在 public.sfc_sites 表中注册站点
    _, err := db.ExecContext(ctx,
        "INSERT INTO public.sfc_sites (name, slug, is_active) VALUES ($1, $2, TRUE)",
        "Test Site "+slug, slug)
    require.NoError(t, err)
    // 创建 site schema 和所有站点级别表
    err = schema.CreateSiteSchema(ctx, db, slug)
    require.NoError(t, err)
}

// DropTestSite 清理测试站点 schema
func DropTestSite(t *testing.T, db *bun.DB, slug string) {
    t.Helper()
    ctx := context.Background()
    _, err := db.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS site_%s CASCADE", slug))
    require.NoError(t, err)
    _, err = db.ExecContext(ctx, "DELETE FROM public.sfc_sites WHERE slug = $1", slug)
    require.NoError(t, err)
}
```

### 3.2 认证链路集成测试

```go
func TestAuthFlow_Integration(t *testing.T) {
    env := SetupTestEnv(t)
    seedUser(t, env.DB, "admin@example.com", "SecurePass123", "admin")

    var accessToken, refreshCookie string

    t.Run("1. 登录成功获取 Token", func(t *testing.T) {
        body := `{"email":"admin@example.com","password":"SecurePass123"}`
        w := performRequest(env.Router, "POST", "/api/v1/auth/login", body, "")

        assert.Equal(t, 200, w.Code)

        var resp map[string]interface{}
        json.Unmarshal(w.Body.Bytes(), &resp)
        accessToken = resp["data"].(map[string]interface{})["access_token"].(string)

        // 验证 Refresh Token Cookie
        cookies := w.Result().Cookies()
        refreshCookie = findCookie(cookies, "refresh_token")
        assert.NotEmpty(t, refreshCookie)
    })

    t.Run("2. 携带 Token 访问受保护接口", func(t *testing.T) {
        w := performRequest(env.Router, "GET", "/api/v1/auth/me", "", accessToken)

        assert.Equal(t, 200, w.Code)
    })

    t.Run("3. 刷新 Token", func(t *testing.T) {
        w := performRequestWithCookie(env.Router, "POST", "/api/v1/auth/refresh",
            "", "refresh_token="+refreshCookie)

        assert.Equal(t, 200, w.Code)

        var resp map[string]interface{}
        json.Unmarshal(w.Body.Bytes(), &resp)
        newToken := resp["data"].(map[string]interface{})["access_token"].(string)
        assert.NotEqual(t, accessToken, newToken)
        accessToken = newToken
    })

    t.Run("4. 登出后 Token 失效", func(t *testing.T) {
        w := performRequest(env.Router, "POST", "/api/v1/auth/logout", "", accessToken)
        assert.Equal(t, 200, w.Code)

        // 登出后再访问应返回 401
        w = performRequest(env.Router, "GET", "/api/v1/auth/me", "", accessToken)
        assert.Equal(t, 401, w.Code)
    })
}
```

### 3.3 文章生命周期集成测试

```go
func TestPostLifecycle_Integration(t *testing.T) {
    env := SetupTestEnv(t)
    token := loginAsEditor(t, env)
    var postID string

    t.Run("1. 创建草稿文章", func(t *testing.T) {
        body := `{
            "title": "测试文章",
            "content": "<p>正文内容</p>",
            "status": "draft",
            "category_ids": [],
            "tag_ids": []
        }`
        w := performRequest(env.Router, "POST", "/api/v1/posts", body, token)

        assert.Equal(t, 201, w.Code)
        postID = extractID(w)
    })

    t.Run("2. 更新文章生成修订版本", func(t *testing.T) {
        body := `{"title": "测试文章（更新）", "content": "<p>更新内容</p>"}`
        w := performRequest(env.Router, "PUT", "/api/v1/posts/"+postID, body, token)

        assert.Equal(t, 200, w.Code)

        // 验证修订版本已创建
        w = performRequest(env.Router, "GET", "/api/v1/posts/"+postID+"/revisions", "", token)
        var resp map[string]interface{}
        json.Unmarshal(w.Body.Bytes(), &resp)
        revisions := resp["data"].([]interface{})
        assert.GreaterOrEqual(t, len(revisions), 1)
    })

    t.Run("3. 发布文章", func(t *testing.T) {
        w := performRequest(env.Router, "POST", "/api/v1/posts/"+postID+"/publish", "", token)
        assert.Equal(t, 200, w.Code)

        // 验证公开 API 可访问
        w = performRequestWithAPIKey(env.Router, "GET", "/api/public/v1/posts", testAPIKey)
        assert.Equal(t, 200, w.Code)
    })

    t.Run("4. 下架文章", func(t *testing.T) {
        w := performRequest(env.Router, "POST", "/api/v1/posts/"+postID+"/unpublish", "", token)
        assert.Equal(t, 200, w.Code)
    })

    t.Run("5. 软删除文章", func(t *testing.T) {
        w := performRequest(env.Router, "DELETE", "/api/v1/posts/"+postID, "", token)
        assert.Equal(t, 200, w.Code)

        // 默认查询不可见
        w = performRequest(env.Router, "GET", "/api/v1/posts/"+postID, "", token)
        assert.Equal(t, 404, w.Code)
    })

    t.Run("6. 恢复文章", func(t *testing.T) {
        w := performRequest(env.Router, "POST", "/api/v1/posts/"+postID+"/restore", "", token)
        assert.Equal(t, 200, w.Code)

        w = performRequest(env.Router, "GET", "/api/v1/posts/"+postID, "", token)
        assert.Equal(t, 200, w.Code)
    })
}
```

### 3.4 缓存失效集成测试

```go
func TestCacheInvalidation_Integration(t *testing.T) {
    env := SetupTestEnv(t)
    token := loginAsEditor(t, env)
    apiKey := createTestAPIKey(t, env)

    // 创建并发布一篇文章
    postID := createAndPublishPost(t, env, token)

    t.Run("1. 首次查询缓存 Miss，写入缓存", func(t *testing.T) {
        w := performRequestWithAPIKey(env.Router, "GET",
            "/api/public/v1/posts/test-slug", apiKey)
        assert.Equal(t, 200, w.Code)

        // 验证 Redis 中已有缓存
        cached, err := env.Redis.Get(ctx, "cache:post:detail:"+postID).Result()
        require.NoError(t, err)
        assert.NotEmpty(t, cached)
    })

    t.Run("2. 更新文章后缓存自动失效", func(t *testing.T) {
        body := `{"title": "更新标题"}`
        performRequest(env.Router, "PUT", "/api/v1/posts/"+postID, body, token)

        // 验证缓存已被删除
        _, err := env.Redis.Get(ctx, "cache:post:detail:"+postID).Result()
        assert.ErrorIs(t, err, redis.Nil)
    })

    t.Run("3. 再次查询重建缓存", func(t *testing.T) {
        w := performRequestWithAPIKey(env.Router, "GET",
            "/api/public/v1/posts/test-slug", apiKey)
        assert.Equal(t, 200, w.Code)

        // 验证新缓存内容包含更新后的标题
        cached, _ := env.Redis.Get(ctx, "cache:post:detail:"+postID).Result()
        assert.Contains(t, cached, "更新标题")
    })
}
```

### 3.5 权限控制集成测试

```go
func TestRBAC_Integration(t *testing.T) {
    env := SetupTestEnv(t)

    tokens := map[string]string{
        "super":  loginAs(t, env, "super"),
        "admin":  loginAs(t, env, "admin"),
        "editor": loginAs(t, env, "editor"),
        "viewer": loginAs(t, env, "viewer"),
    }

    tests := []struct {
        method   string
        path     string
        roles    map[string]int // role -> expected status code
    }{
        {"POST", "/api/v1/posts", map[string]int{
            "super": 201, "admin": 201, "editor": 201, "viewer": 403,
        }},
        {"POST", "/api/v1/users", map[string]int{
            "super": 201, "admin": 403, "editor": 403, "viewer": 403,
        }},
        {"POST", "/api/v1/categories", map[string]int{
            "super": 201, "admin": 201, "editor": 403, "viewer": 403,
        }},
        {"GET", "/api/v1/posts", map[string]int{
            "super": 200, "admin": 200, "editor": 200, "viewer": 200,
        }},
    }

    for _, tt := range tests {
        for role, expectedCode := range tt.roles {
            t.Run(fmt.Sprintf("%s %s as %s", tt.method, tt.path, role), func(t *testing.T) {
                body := getTestBody(tt.method, tt.path)
                w := performRequest(env.Router, tt.method, tt.path, body, tokens[role])
                assert.Equal(t, expectedCode, w.Code)
            })
        }
    }
}
```

---

## 4. 前端测试

### 4.1 组件单元测试（Vitest + React Testing Library）

```typescript
// src/components/auth/__tests__/LoginForm.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { LoginForm } from '../LoginForm';

describe('LoginForm', () => {
  it('应渲染邮箱和密码输入框', () => {
    render(<LoginForm />);
    expect(screen.getByLabelText(/邮箱/)).toBeInTheDocument();
    expect(screen.getByLabelText(/密码/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /登录/ })).toBeInTheDocument();
  });

  it('邮箱格式错误时显示校验提示', async () => {
    render(<LoginForm />);
    const emailInput = screen.getByLabelText(/邮箱/);

    fireEvent.change(emailInput, { target: { value: 'invalid-email' } });
    fireEvent.blur(emailInput);

    await waitFor(() => {
      expect(screen.getByText(/请输入有效的邮箱地址/)).toBeInTheDocument();
    });
  });

  it('提交表单时调用 onSubmit 回调', async () => {
    const onSubmit = vi.fn();
    render(<LoginForm onSubmit={onSubmit} />);

    fireEvent.change(screen.getByLabelText(/邮箱/), {
      target: { value: 'admin@example.com' },
    });
    fireEvent.change(screen.getByLabelText(/密码/), {
      target: { value: 'SecurePass123' },
    });
    fireEvent.click(screen.getByRole('button', { name: /登录/ }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        email: 'admin@example.com',
        password: 'SecurePass123',
      });
    });
  });
});
```

### 4.2 Hook 测试

```typescript
// src/hooks/__tests__/usePosts.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { usePosts } from '../usePosts';
import { QueryClientWrapper } from '../../test-utils';

describe('usePosts', () => {
  it('应获取文章列表', async () => {
    const { result } = renderHook(() => usePosts({ page: 1 }), {
      wrapper: QueryClientWrapper,
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.data).toHaveLength(20);
    expect(result.current.data?.pagination.total).toBeGreaterThan(0);
  });

  it('筛选条件变化时重新请求', async () => {
    const { result, rerender } = renderHook(
      ({ status }) => usePosts({ page: 1, status }),
      { wrapper: QueryClientWrapper, initialProps: { status: 'draft' } }
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const draftCount = result.current.data?.pagination.total;

    rerender({ status: 'published' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.pagination.total).not.toBe(draftCount);
  });
});
```

### 4.3 API Mock 配置（MSW）

```typescript
// src/mocks/handlers.ts
import { http, HttpResponse } from 'msw';

export const handlers = [
  http.post('/api/v1/auth/login', async ({ request }) => {
    const body = await request.json();
    if (body.email === 'admin@example.com' && body.password === 'SecurePass123') {
      return HttpResponse.json({
        success: true,
        data: {
          access_token: 'mock-jwt-token',
          expires_in: 900,
          user: { id: 'uuid', email: 'admin@example.com', role: 'admin' },
        },
      });
    }
    return HttpResponse.json(
      { success: false, error: { code: 'UNAUTHORIZED', message: '邮箱或密码错误' } },
      { status: 401 }
    );
  }),

  http.get('/api/v1/posts', () => {
    return HttpResponse.json({
      success: true,
      data: mockPosts,
      pagination: { page: 1, per_page: 20, total: 42, total_pages: 3 },
    });
  }),
];
```

---

## 5. E2E 测试（Playwright）

### 5.1 测试配置

```typescript
// playwright.config.ts
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  retries: 1,
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  webServer: {
    command: 'docker compose up -d && bun dev',
    url: 'http://localhost:3000',
    timeout: 60_000,
    reuseExistingServer: true,
  },
});
```

### 5.2 E2E 测试用例

```typescript
// e2e/auth.spec.ts
import { test, expect } from '@playwright/test';

test.describe('认证流程', () => {
  test('成功登录并跳转到 Dashboard', async ({ page }) => {
    await page.goto('/login');
    await page.fill('[name="email"]', 'admin@example.com');
    await page.fill('[name="password"]', 'SecurePass123');
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('text=Dashboard')).toBeVisible();
  });

  test('错误密码显示错误提示', async ({ page }) => {
    await page.goto('/login');
    await page.fill('[name="email"]', 'admin@example.com');
    await page.fill('[name="password"]', 'WrongPassword');
    await page.click('button[type="submit"]');

    await expect(page.locator('text=邮箱或密码错误')).toBeVisible();
  });

  test('未登录访问受保护页面重定向到登录', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL('/login');
  });
});

// e2e/post-management.spec.ts
test.describe('文章管理', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsEditor(page);
  });

  test('创建文章→保存草稿→发布→公开API可访问', async ({ page, request }) => {
    // 1. 创建文章
    await page.goto('/posts/new');
    await page.fill('[name="title"]', 'E2E 测试文章');
    await page.locator('.editor').fill('这是测试正文内容');

    // 2. 保存草稿
    await page.click('button:has-text("保存草稿")');
    await expect(page.locator('text=保存成功')).toBeVisible();

    // 3. 发布
    await page.click('button:has-text("发布")');
    await expect(page.locator('text=发布成功')).toBeVisible();

    // 4. 验证公开 API 可访问
    const resp = await request.get('/api/public/v1/posts/e2e-ce-shi-wen-zhang', {
      headers: { 'X-API-Key': testAPIKey },
    });
    expect(resp.ok()).toBeTruthy();
    const data = await resp.json();
    expect(data.data.title).toBe('E2E 测试文章');
  });

  test('媒体上传→文章引用→删除保护', async ({ page }) => {
    // 1. 上传图片
    await page.goto('/media');
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles('e2e/fixtures/test-image.png');
    await expect(page.locator('text=上传成功')).toBeVisible();

    // 2. 创建文章引用该图片
    await page.goto('/posts/new');
    await page.fill('[name="title"]', '含图片文章');
    // 设置封面图...
    await page.click('button:has-text("发布")');

    // 3. 尝试删除被引用的图片
    await page.goto('/media');
    await page.click('[data-testid="delete-media"]');
    await expect(page.locator('text=该文件被 1 篇文章引用')).toBeVisible();
  });
});

// e2e/rbac.spec.ts
test.describe('角色权限', () => {
  test('Viewer 角色看不到文章编辑菜单', async ({ page }) => {
    await loginAsViewer(page);
    await page.goto('/dashboard');

    // 侧栏不应包含"新建文章"入口
    await expect(page.locator('a:has-text("新建文章")')).not.toBeVisible();
  });

  test('Editor 角色不能访问用户管理页', async ({ page }) => {
    await loginAsEditor(page);
    await page.goto('/users');

    await expect(page.locator('text=无权限')).toBeVisible();
  });
});
```

### 5.3 E2E 测试用例矩阵

| 用例编号 | 测试场景 | 验证内容 | 优先级 |
|----------|----------|----------|--------|
| I18N-E2E-001 | 多语言内容完整工作流 | 创建中文文章 → 添加英文翻译 → 发布 → Public API 按 locale 查询 → 验证两种语言内容正确 | P1 |
| I18N-E2E-002 | 界面语言切换 | 用户切换 UI 语言 zh-CN ↔ en-US → 验证所有界面文本正确切换 | P2 |
| I18N-E2E-003 | 默认语言翻译不可删除 | 尝试删除 default_locale 的翻译 → 返回 400 错误 | P1 |
| E2E-WORKFLOW-001 | 文章完整生命周期 | 注册 → 登录 → 创建文章 → 添加分类/标签 → 上传封面图 → 发布 → Public API 查看 → 下架 → 软删除 → 恢复 → 永久删除 | P1 |
| E2E-WORKFLOW-002 | 多用户协作流程 | Admin 创建用户 → Editor 登录 → 创建草稿 → Admin 审阅 → 发布 | P2 |
| E2E-WORKFLOW-003 | 权限升级攻击 | Editor 尝试直接调用 Admin API → 全部返回 403 | P1 |

---

## 6. 性能测试

### 6.1 k6 压测方案

```javascript
// web/performance/api-load.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 100 },   // 30秒内爬升到 100 并发
    { duration: '2m',  target: 500 },   // 2分钟保持 500 并发
    { duration: '1m',  target: 1000 },  // 1分钟爬升到 1000 并发
    { duration: '2m',  target: 1000 },  // 2分钟保持峰值
    { duration: '30s', target: 0 },     // 30秒收尾
  ],
  thresholds: {
    http_req_duration: ['p(99)<200'],    // P99 < 200ms
    http_req_failed: ['rate<0.01'],      // 错误率 < 1%
  },
};

const API_KEY = __ENV.API_KEY || 'cms_live_test_key';
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // 公开 API 文章列表（带缓存）
  const listRes = http.get(`${BASE_URL}/api/public/v1/posts?page=1&per_page=20`, {
    headers: { 'X-API-Key': API_KEY },
  });
  check(listRes, {
    '列表状态码 200': (r) => r.status === 200,
    '列表响应 < 200ms': (r) => r.timings.duration < 200,
  });

  sleep(0.5);

  // 文章详情（带缓存）
  const detailRes = http.get(`${BASE_URL}/api/public/v1/posts/test-slug`, {
    headers: { 'X-API-Key': API_KEY },
  });
  check(detailRes, {
    '详情状态码 200': (r) => r.status === 200,
    '详情响应 < 200ms': (r) => r.timings.duration < 200,
  });

  sleep(0.5);
}
```

### 6.2 性能指标验证

| 指标 | 目标 | 验证方法 |
|------|------|----------|
| API P99 响应时间 | < 200ms | k6 `http_req_duration` p(99) |
| 缓存命中时响应 | < 20ms | k6 自定义 check |
| 1000 并发错误率 | < 1% | k6 `http_req_failed` |
| 管理后台 FCP | < 1.5s | Lighthouse CI |
| 数据库连接池使用率 | < 80% | pgBouncer / 监控 |

### 6.3 缓存性能验证用例

| 用例编号 | 测试场景 | 验证内容 | 优先级 |
|----------|----------|----------|--------|
| PERF-CACHE-001 | 列表缓存 TTL 验证 | GET /public/posts → 首次查询 DB → 第二次命中 Redis → 60 秒后缓存过期 → 再次查询 DB | P2 |
| PERF-CACHE-002 | 缓存命中响应时间 | 缓存命中时 P99 < 20ms | P1 |
| PERF-CACHE-003 | 写操作缓存失效 | PUT /posts/:id → 立即 GET → 应返回最新数据（缓存已主动失效） | P1 |

---

## 7. 测试用例矩阵

### 7.1 认证模块

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| AUTH-001 | 正确邮箱密码登录 | 单元+集成 | 返回 200，包含 access_token |
| AUTH-002 | 错误密码登录 | 单元+集成 | 返回 401，错误计数 +1 |
| AUTH-003 | 连续 5 次失败后锁定 | 集成 | 返回 429，锁定 15 分钟 |
| AUTH-004 | 锁定期间正确密码也拒绝 | 集成 | 返回 429 |
| AUTH-005 | Token 过期后自动刷新 | 集成+E2E | 无感刷新，请求不中断 |
| AUTH-006 | Refresh Token 过期 | 集成 | 返回 401，需重新登录 |
| AUTH-007 | 登出后 Token 失效 | 集成 | 访问受保护接口返回 401 |
| AUTH-008 | 修改密码后 Token 全部失效 | 集成 | 旧 Token 返回 401 |
| AUTH-009 | 不存在的邮箱登录 | 单元 | 返回 401（不暴露邮箱是否存在） |
| AUTH-010 | SQL 注入尝试 | 集成 | 返回 400/401，数据库无影响 |
| AUTH-011 | 修改密码 — 成功 | 集成 | PUT /api/v1/users/me/password { old_password, new_password } → 200，旧密码验证通过，新密码保存，所有 Refresh Token 失效 |
| AUTH-012 | 修改密码 — 旧密码错误 | 单元+集成 | PUT /api/v1/users/me/password { wrong_old_password } → 400 INVALID_PASSWORD |
| AUTH-013 | 忘记密码 — 请求重置 | 集成 | POST /api/v1/auth/forgot-password { email } → 200（无论邮箱是否存在），已注册邮箱收到重置邮件 |
| AUTH-014 | 忘记密码 — 执行重置 | 集成 | POST /api/v1/auth/reset-password { token, new_password } → 200，密码已更新，所有 Refresh Token 失效 |
| AUTH-015 | 重置令牌过期 | 集成 | 使用 30 分钟前生成的 reset token → 400 TOKEN_EXPIRED |
| AUTH-016 | 重置令牌重复使用 | 集成 | 使用已消费的 reset token → 400 TOKEN_INVALID |

### 7.2 文章管理模块

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| POST-001 | 创建草稿文章 | 单元+集成 | 返回 201，状态 draft |
| POST-002 | 自动生成 Slug | 单元 | 中文标题正确转为 pinyin slug |
| POST-003 | Slug 重复处理 | 集成 | 自动追加后缀（-1, -2） |
| POST-004 | 发布文章 | 集成 | 状态变 published，published_at 设置 |
| POST-005 | 重复发布已发布文章 | 单元 | 返回业务错误 |
| POST-006 | 计划发布 | 集成 | 状态 scheduled，到点后自动发布 |
| POST-007 | 更新文章生成修订 | 集成 | 修订版本号 +1，内容记录 |
| POST-008 | 回滚到历史版本 | 集成 | 文章内容恢复，新增回滚修订记录 |
| POST-009 | 软删除文章 | 集成 | deleted_at 设置，默认查询不可见 |
| POST-010 | 恢复已删除文章 | 集成 | deleted_at 清空，恢复为 draft 状态 |
| POST-011 | 全文检索（中文） | 集成 | Meilisearch 返回相关结果 |
| POST-012 | 分页查询边界 | 单元+集成 | 第一页/最后一页/超出范围正确处理 |
| POST-013 | 联合筛选（状态+分类+标签） | 集成 | 返回满足所有条件的文章 |
| POST-014 | XSS 内容过滤 | 单元 | `<script>` 标签被过滤 |
| POST-015 | 创建文章时携带 extra_fields | 集成 | POST /posts { extra_fields: { color: "red" } } → 201，响应含 extra_fields |
| POST-016 | 更新文章 extra_fields（部分更新） | 集成 | PUT /posts/:id { extra_fields: { color: "blue" } } → 合并更新，不覆盖其他字段 |
| POST-017 | Public API 返回 extra_fields | 集成 | GET /public/posts/:slug → 响应含完整 extra_fields |
| POST-018 | 自动保存草稿 | E2E | 编辑文章后 30 秒无操作 → 自动触发 PUT /posts/:id → 验证内容已保存 |
| PAGE-001 | 分页参数 page=0 | 单元+集成 | GET /posts?page=0 → 返回 400 VALIDATION_ERROR |
| PAGE-002 | per_page 超过上限 | 单元+集成 | GET /posts?per_page=999 → 自动截断为 100，正常返回 |
| PAGE-003 | page 超过总页数 | 集成 | GET /posts?page=999 → 返回空 data[]，pagination.total 正确 |

### 7.3 分类与标签模块

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| CAT-001 | 创建顶级分类 | 单元+集成 | path 为 /slug/ |
| CAT-002 | 创建子分类 | 集成 | path 为 /parent-slug/child-slug/ |
| CAT-003 | 删除有子分类的分类 | 集成 | 返回 409，提示需先处理子分类 |
| CAT-004 | 拖拽排序 | 集成 | sort_order 批量更新 |
| CAT-005 | post_count 自动更新 | 集成 | 文章关联/取消关联时计数变化 |
| TAG-001 | 创建标签 | 单元+集成 | 自动生成 slug |
| TAG-002 | 标签自动补全 | 集成 | 模糊匹配返回建议 |
| TAG-003 | post_count 触发器 | 集成 | 关联/取消关联时计数正确更新 |

### 7.4 媒体管理模块

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| MEDIA-001 | 上传图片（jpg/png/webp） | 集成 | 返回 201，生成缩略图和 WebP |
| MEDIA-002 | 上传超大文件（>100MB） | 集成 | 返回 413 |
| MEDIA-003 | 上传非白名单类型 | 集成 | 返回 415 |
| MEDIA-004 | MIME 类型欺骗（伪装扩展名） | 集成 | 返回 415（检查实际 MIME） |
| MEDIA-005 | 删除无引用的媒体 | 集成 | 删除成功，文件从存储中移除 |
| MEDIA-006 | 删除有引用的媒体 | 集成 | 返回 409，包含引用文章列表 |
| MEDIA-007 | 媒体库搜索 | 集成 | 按文件名模糊搜索 |

### 7.5 Headless API 模块

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| API-001 | 无 API Key 访问 | 集成 | 返回 401 |
| API-002 | 无效 API Key | 集成 | 返回 401 |
| API-003 | 已吊销 API Key | 集成 | 返回 401 |
| API-004 | 有效 Key 获取文章列表 | 集成 | 返回 200，仅包含 published 状态 |
| API-005 | 超过限流阈值 | 集成 | 返回 429，包含 Retry-After |
| API-006 | Cache-Control 头验证 | 集成 | 包含正确的 max-age |
| API-007 | 全文搜索公开 API | 集成 | 返回相关结果 |

### 7.6 系统设置模块

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| SYS-001 | Admin 获取系统配置列表 | 集成 | 返回 200，包含所有配置项 |
| SYS-002 | 超级管理员更新配置项 | 集成 | 返回 200，值已更新，Redis 缓存同步刷新 |
| SYS-003 | Admin 尝试更新配置（非超级管理员） | 集成 | 返回 403 Forbidden |
| SYS-004 | Editor 访问系统配置 | 集成 | 返回 403 Forbidden |
| SYS-005 | 更新不存在的配置项 | 集成 | 返回 404 Not Found |

### 7.7 审计日志模块

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| AUDIT-001 | 超级管理员查询审计日志 | 集成 | 返回 200，支持分页 |
| AUDIT-002 | 按操作人筛选审计日志 | 集成 | 仅返回指定 actor_id 的日志 |
| AUDIT-003 | 按操作类型筛选（create/update/delete） | 集成 | 仅返回指定 action 的日志 |
| AUDIT-004 | 按时间范围筛选 | 集成 | 仅返回 start_date 和 end_date 之间的日志 |
| AUDIT-005 | Admin 尝试查询审计日志 | 集成 | 返回 403 Forbidden |
| AUDIT-006 | 写操作自动记录审计日志 | 集成 | 创建/更新/删除文章后 sfc_site_audits 表有对应记录 |

### 7.8 多语言模块

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| I18N-001 | 为文章创建英文翻译版本 | 集成 | 返回 200，sfc_site_post_translations 表有对应记录 |
| I18N-002 | 获取文章所有语言版本列表 | 集成 | 返回 200，包含所有已创建的 locale 列表 |
| I18N-003 | 获取指定语言版本内容 | 集成 | 返回 200，标题和正文为对应语言内容 |
| I18N-004 | 更新已有语言版本 | 集成 | 返回 200，内容已更新，updated_at 刷新 |
| I18N-005 | 删除语言版本 | 集成 | 返回 200，sfc_site_post_translations 记录已删除 |
| I18N-006 | 公开 API 按 locale 参数返回对应语言 | 集成 | 传入 locale=en 返回英文版本，缺省返回默认语言 |

### 7.9 安全测试用例

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| SEC-001 | SVG XSS 注入 | 集成 | 上传含 `<script>` 的 SVG 文件 → 应拒绝或清洗危险标签 |
| SEC-002 | 文件名路径遍历 | 集成 | 上传文件名为 `../../../etc/passwd` → 应清洗为安全文件名 |
| SEC-003 | MIME 类型伪造 | 集成 | 将 .exe 重命名为 .jpg 上传 → 应通过 magic bytes 检测真实类型并拒绝 |
| SEC-004 | XSS 内容过滤 | 单元+集成 | 文章内容含 `<img onerror="alert(1)">` → bluemonday 应移除 onerror 属性 |
| SEC-005 | SQL 注入 | 集成 | 搜索参数含 `'; DROP TABLE posts; --` → 参数化查询应安全处理 |
| SEC-006 | JWT 篡改 | 单元+集成 | 修改 JWT payload 后发送请求 → 签名验证失败返回 401 |
| SEC-007 | Refresh Token 重放攻击 | 集成 | 使用已 revoked 的 Refresh Token → 返回 401 并撤销该用户所有 Token |

### 7.10 Redis 故障降级测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| REDIS-001 | Redis 不可用时缓存降级 | 集成 | 停止 Redis → GET /posts → 应直接查询数据库返回正确结果（性能下降但功能正常） |
| REDIS-002 | Redis 不可用时 Rate Limit 降级 | 集成 | 停止 Redis → 发送请求 → 应放行（fail-open）并记录告警日志 |
| REDIS-003 | Redis 不可用时 Token 黑名单 | 集成 | 停止 Redis → 使用已登出的 Token → 应拒绝（fail-closed，安全优先） |

### 7.11 数据库事务回滚测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| TX-001 | 文章创建事务回滚 | 集成 | 创建文章时 revision 插入失败 → 文章记录也应回滚，数据库无脏数据 |
| TX-002 | 分类更新事务回滚 | 集成 | 更新分类 path 时子分类更新失败 → 父分类 path 也应回滚 |

### 7.12 并发操作测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| CONC-001 | 并发编辑冲突检测 | 集成 | 两个请求同时 PUT /posts/:id（相同 version）→ 一个成功，一个返回 409 VERSION_CONFLICT |
| CONC-002 | 并发文件上传 | 集成 | 同时上传相同文件名 → 两个文件均成功，生成不同存储路径 |
| CONC-003 | 并发 Token 刷新 | 集成 | 同时发送两个 refresh 请求 → 第一个成功，第二个因 Token 已 rotated 返回 401 |

### 7.13 多站点（Multi-Site）测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| SITE-001 | 创建新站点 | 集成 | POST /api/v1/sites → schema `site_{slug}` 创建成功，包含所有站点级别表（sfc_site_posts, sfc_site_categories, sfc_site_tags, sfc_site_menus, sfc_site_comments, sfc_site_redirects, sfc_site_preview_tokens, sfc_site_api_keys, sfc_site_audits, sfc_site_configs, sfc_site_post_types） |
| SITE-002 | 删除站点 | 集成 | DELETE /api/v1/sites/:slug { confirm_slug } → schema 被 DROP CASCADE，sfc_user_roles 清理，public.sfc_sites 记录删除 |
| SITE-003 | Schema 隔离 — 数据不可见 | 集成 | 在 site_a 中创建文章，设置 search_path 为 site_b → 查询 sfc_site_posts 返回空结果，site_a 的数据在 site_b 中不可见 |
| SITE-004 | 全局角色分配 | 集成 | 用户通过 `sfc_user_roles` 分配全局角色 → RBAC 中间件根据 `sfc_role_apis` 动态匹配 API 权限，角色权限全局生效 |
| SITE-005 | 域名映射 | 集成 | Host header 为 blog.example.com → SiteResolverMiddleware 正确解析到对应站点，SchemaMiddleware 设置正确的 search_path |
| SITE-006 | 无效站点 slug | 集成 | 请求不存在的站点 slug → 返回 404 NOT_FOUND |

### 7.14 安装向导（Installation Wizard）测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| SETUP-001 | 全新安装检查 | 集成 | POST /api/v1/setup/check → 返回 `{ "installed": false }` |
| SETUP-002 | 执行初始化 | 集成 | POST /api/v1/setup/initialize { site_name, site_slug, admin_email, admin_password, ... } → 创建 public schema 表 + 第一个站点 schema + admin 用户 + sfc_user_roles(super) + 返回 JWT access_token |
| SETUP-003 | 安装后端点禁用 | 集成 | 安装完成后，POST /api/v1/setup/initialize → 返回 409 ALREADY_INSTALLED；POST /api/v1/setup/check → 返回 `{ "installed": true }` |
| SETUP-004 | 并发安装竞争 | 集成 | 两个并发请求同时调用 /api/v1/setup/initialize → 只有一个成功（PostgreSQL advisory lock），另一个返回 409 ALREADY_INSTALLED |

### 7.15 评论（Comments）测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| COMMENT-001 | 游客提交评论 | 集成 | POST /api/public/v1/posts/:slug/comments { author_name, author_email, content } → 返回 201，status 为 pending |
| COMMENT-002 | 已登录用户提交评论 | 集成 | 携带 JWT 提交评论 → 返回 201，响应包含 user_id，author_name 和 author_email 从用户记录填充 |
| COMMENT-003 | 管理员审核评论 | 集成 | PUT /api/v1/comments/:id/status { status: "approved" } → 返回 200；PUT { status: "spam" } → 返回 200 |
| COMMENT-004 | 嵌套回复（最大 3 级） | 集成 | 创建 level 0 → reply level 1 → reply level 2 成功；尝试回复 level 2 评论 → 返回 422 MAX_NESTING_DEPTH |
| COMMENT-005 | 频率限制 | 集成 | 同一 IP 30 秒内提交第 2 条评论 → 返回 429 RATE_LIMITED |
| COMMENT-006 | 蜜罐字段反垃圾 | 集成 | 提交 honeypot 字段非空 → 返回 201 但评论自动标记为 spam（静默拒绝） |
| COMMENT-007 | XSS 内容清洗 | 单元+集成 | 评论内容含 `<script>alert(1)</script>` → 存储的 content 中 HTML 标签已被剥离 |

### 7.16 菜单（Menu）测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| MENU-001 | 创建菜单及菜单项 | 集成 | POST /api/v1/menus → 创建成功；POST /api/v1/menus/:id/items → 菜单项按 sort_order 正确排序 |
| MENU-002 | 重新排序菜单项 | 集成 | PUT /api/v1/menus/:id/items/reorder → 批量更新 sort_order 和 parent_id，GET 返回新的排序 |
| MENU-003 | 嵌套菜单项（最大 3 级） | 集成 | 创建 3 级嵌套成功；尝试创建第 4 级 → 返回 422 UNPROCESSABLE |
| MENU-004 | 引用解析 — 文章链接 | 集成 | 菜单项 type=post, reference_id=post_uuid → GET 响应中 url 自动解析为该文章的 slug URL |
| MENU-005 | 已删除文章的引用 | 集成 | 文章被软删除 → GET 菜单时该菜单项 url 为 null，is_broken 为 true |

### 7.17 RSS/Sitemap 测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| FEED-001 | RSS feed 内容 | 集成 | GET /feed/rss.xml → 返回 application/rss+xml，包含最新 20 篇 published 文章，Content 包含标题、链接、pubDate |
| FEED-002 | Atom feed 有效 XML | 集成 | GET /feed/atom.xml → 返回 application/atom+xml，XML 可正确解析，包含 entry 元素 |
| FEED-003 | Sitemap 包含所有已发布文章 | 集成 | GET /sitemap-posts.xml → 每个 published 文章有对应 `<url>` 元素，包含 `<lastmod>` |
| FEED-004 | 发布新文章后缓存失效 | 集成 | 发布文章 → 立即 GET /feed/rss.xml → 新文章出现在 feed 中（缓存已被清除重建） |

### 7.18 草稿预览（Preview）测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| PREVIEW-001 | 生成预览令牌 | 集成 | POST /api/v1/posts/:id/preview → 返回 201，包含 preview_url 和 token；GET /api/public/v1/preview/:token → 返回 200，包含草稿文章完整内容 |
| PREVIEW-002 | 过期令牌 | 集成 | 使用已过期的预览令牌访问 → 返回 410 GONE |
| PREVIEW-003 | 每篇文章最多 5 个活跃令牌 | 集成 | 已有 5 个活跃令牌 → 尝试创建第 6 个 → 返回 422 UNPROCESSABLE |
| PREVIEW-004 | 撤销所有令牌 | 集成 | DELETE /api/v1/posts/:id/preview → 返回 200；之前生成的所有预览链接返回 404 |

### 7.19 URL 重定向（Redirect）测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| REDIRECT-001 | 创建 301 重定向 | 集成 | POST /api/v1/redirects { source_path: "/old", target_url: "/new", status_code: 301 } → 201；GET /old → 返回 301 Location: /new |
| REDIRECT-002 | 文章 slug 变更自动重定向 | 集成 | PUT /api/v1/posts/:id { slug: "new-slug" } → 自动在 redirects 表创建 301 记录 /old-slug → /new-slug |
| REDIRECT-003 | 重定向链折叠 | 集成 | 已有 A→B，更新 B 的 slug 为 C → 结果为 A→C + B→C（链被折叠，不存在 A→B→C） |
| REDIRECT-004 | 命中计数 | 集成 | GET /old（触发重定向）→ hit_count 通过 Redis INCR 缓冲 → cron flush 后数据库 hit_count 递增 |

### 7.20 双因素认证（2FA）测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| 2FA-001 | 初始化 2FA | 集成 | POST /api/v1/auth/2fa/setup → 返回 200，包含 secret（Base32）、qr_code_uri、qr_code_svg、10 个 backup_codes |
| 2FA-002 | 验证有效 TOTP 启用 2FA | 集成 | POST /api/v1/auth/2fa/verify { code: valid_totp } → 返回 200，user_totp.is_enabled = true，verified_at 已设置 |
| 2FA-003 | 启用 2FA 后登录需两步 | 集成 | POST /auth/login → 返回 { requires_2fa: true, temp_token }；POST /auth/2fa/validate { code } with temp_token → 返回完整 access_token + refresh_token |
| 2FA-004 | 无效 TOTP 码被拒绝 | 集成 | POST /api/v1/auth/2fa/validate { code: "000000" } → 返回 400 INVALID_CODE |
| 2FA-005 | 验证频率限制 | 集成 | 5 分钟内第 6 次验证尝试 → 返回 429 RATE_LIMITED |
| 2FA-006 | 备用码一次性使用 | 集成 | 使用有效 backup_code 登录成功 → 同一 backup_code 再次使用 → 返回 400 INVALID_CODE（已消费） |
| 2FA-007 | 禁用 2FA 需要密码和 TOTP | 集成 | POST /api/v1/auth/2fa/disable { password, code } → 两者都正确时返回 200，user_totp 记录删除 |
| 2FA-008 | 超级管理员强制禁用 | 集成 | DELETE /api/v1/auth/2fa/users/:user_id { reason } → 返回 200，目标用户 2FA 被禁用，所有 refresh_token 被撤销，audit_log 记录原因 |

### 7.21 Schema 隔离 E2E 测试

| 用例编号 | 测试场景 | 测试类型 | 预期结果 |
|----------|----------|----------|----------|
| SCHEMA-E2E-001 | 跨站点完整工作流 | E2E | 创建 site_a 和 site_b → 在 site_a 创建文章 → 在 site_b 创建文章 → 验证 site_a 的 Public API 仅返回 site_a 的文章，site_b 同理 |
| SCHEMA-E2E-002 | 多角色跨站点用户 | E2E | 用户 U 在 site_a 为 Admin，在 site_b 为 Editor → 以 U 登录 site_a 可管理用户 → 以 U 登录 site_b 不能管理用户（403） |
| SCHEMA-E2E-003 | 迁移应用到所有 schema | 集成 | 新增站点级别迁移 → `go run ./cmd/cms migrate up` → 验证所有已存在的 site schema 均已应用变更 |

---

## 8. CI/CD 测试集成

### 8.1 GitHub Actions 配置

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  backend-test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:18-alpine
        env:
          POSTGRES_DB: cms_test
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports:
          - 5432:5432
        options: >-
          --health-cmd="pg_isready -U test"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=5
      redis:
        image: redis:8-alpine
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run unit tests
        working-directory: .
        run: go test -v -race -coverprofile=coverage.out ./internal/...

      - name: Run integration tests
        working-directory: .
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/cms_test?sslmode=disable
          REDIS_URL: redis://localhost:6379/0
          TOTP_ENCRYPTION_KEY: ${{ secrets.TEST_TOTP_ENCRYPTION_KEY }}
        run: go test -v -race -tags=integration ./tests/...

      - name: Check coverage
        working-directory: .
        run: |
          go tool cover -func=coverage.out
          # 确保覆盖率 ≥ 80%（逐步提高到 90%）
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage $COVERAGE% is below threshold 80%"
            exit 1
          fi

  frontend-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
        with:
          bun-version: latest

      - name: Install dependencies
        working-directory: ./web
        run: bun install --frozen-lockfile

      - name: Run unit & component tests
        working-directory: ./web
        run: bunx vitest run --coverage

      - name: Install Playwright
        working-directory: ./web
        run: bunx playwright install --with-deps

      - name: Run E2E tests
        working-directory: ./web
        run: bunx playwright test

  performance-test:
    runs-on: ubuntu-latest
    needs: [backend-test, frontend-test]
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: grafana/k6-action@v0.3.1
        with:
          filename: web/performance/api-load.js
          flags: --out json=results.json
        env:
          BASE_URL: http://localhost:8080
          API_KEY: ${{ secrets.TEST_API_KEY }}
```

### 8.2 测试运行命令汇总

```bash
# 后端单元测试
go test -v -race ./internal/...

# 后端集成测试（需要 Docker）
go test -v -race -tags=integration ./tests/integration/...

# 后端覆盖率报告
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# 前端单元测试
cd web && bunx vitest run

# 前端覆盖率
cd web && bunx vitest run --coverage

# E2E 测试
cd web && bunx playwright test

# E2E 测试（带 UI）
cd web && bunx playwright test --ui

# 性能测试
k6 run web/performance/api-load.js

# 全量测试
make test-all
```

---

## 9. Makefile 测试命令

```makefile
.PHONY: test test-unit test-integration test-e2e test-perf test-all

test-unit:
	go test -v -race -coverprofile=coverage.out ./internal/...
	cd web && bunx vitest run --coverage

test-integration:
	go test -v -race -tags=integration ./tests/integration/...

test-e2e:
	cd web && bunx playwright test

test-perf:
	k6 run web/performance/api-load.js

test-all: test-unit test-integration test-e2e

test-coverage:
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"
```

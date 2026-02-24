# Batch 5: 内容分类体系 + 媒体管理 — 详细设计

**日期**: 2026-02-24
**范围**: Categories (6) + Tags (6) + Media (6) = 18 个站点级端点
**前置条件**: Batch 4 已完成（63 路由, SiteResolver + Schema + AuditContext 中间件就绪）

---

## 1. 设计决策摘要

| 决策 | 选择 | 备选方案 |
|------|------|----------|
| 图片处理 | 同步处理 (上传时 WebP + 缩略图) | 异步 goroutine / V1.0 不处理 |
| Tags suggest | Meilisearch | pg_trgm / 先 pg 后迁移 |
| Categories path 维护 | Service 层级联更新 | DB 触发器 / 不存储 path |
| Meilisearch 集成 | 共享基础设施 (pkg/search) | 模块内部直接集成 |
| 实施顺序 | 基础设施优先 + 模块串行 | 模块并行 / 混合方案 |
| StorageClient 位置 | pkg/storage | internal/database/ |
| Redis 缓存位置 | pkg/cache (重构) | internal/database/redis.go |

---

## 2. 共享基础设施

### 2.1 pkg/search — Meilisearch 客户端包装

```
pkg/search/
├── client.go      # Meilisearch 客户端初始化 + 连接管理
├── indexer.go     # 通用索引同步接口 + 实现
└── search.go      # 通用搜索查询接口
```

**核心接口**:

```go
// Client 管理 Meilisearch 连接
type Client struct {
    ms meilisearch.ServiceManager
}

func NewClient(host, apiKey string) (*Client, error)

// Indexer 定义索引同步操作
type Indexer interface {
    UpsertDocuments(ctx context.Context, indexUID string, docs any) error
    DeleteDocuments(ctx context.Context, indexUID string, ids []string) error
    EnsureIndex(ctx context.Context, indexUID string, settings *IndexSettings) error
}

// Searcher 定义搜索查询
type Searcher interface {
    Search(ctx context.Context, indexUID string, query string, opts *SearchOpts) (*SearchResult, error)
}

// IndexSettings 配置可搜索属性
type IndexSettings struct {
    SearchableAttributes []string
    DisplayedAttributes  []string
    FilterableAttributes []string
    SortableAttributes   []string
}

// SearchOpts 搜索参数
type SearchOpts struct {
    Limit  int
    Offset int
    Filter string
}

// SearchResult 搜索结果
type SearchResult struct {
    Hits             []map[string]any
    EstimatedTotal   int64
    ProcessingTimeMs int64
}
```

**索引命名规则**: `tags-{siteSlug}`, `posts-{siteSlug}` (后续 Batch 6)

**连接初始化**: `database/meilisearch.go` 负责从 config 创建连接, serve.go 中调用。

### 2.2 pkg/imaging — 图片处理

```
pkg/imaging/
├── processor.go   # WebP 转换 + 缩略图生成
└── metadata.go    # 图片尺寸提取
```

**核心接口**:

```go
type Processor interface {
    ToWebP(src io.Reader, quality int) ([]byte, error)
    Thumbnail(src io.Reader, width, height int) ([]byte, error)
    ExtractDimensions(src io.Reader) (width, height int, err error)
}
```

**缩略图规格**:
- `sm`: 150×150 (裁剪居中)
- `md`: 400×400 (等比缩放)

**技术选型**: Go 标准库 `image` + `golang.org/x/image/webp` 编码, 无 CGO 依赖。

### 2.3 pkg/storage — RustFS/S3 客户端

```
pkg/storage/
├── client.go    # StorageClient 接口 + S3 实现
└── config.go    # 连接配置
```

**核心接口**:

```go
type StorageClient interface {
    Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) error
    Delete(ctx context.Context, key string) error
    BatchDelete(ctx context.Context, keys []string) error
    PublicURL(key string) string
}
```

**S3 实现**: 使用 AWS SDK v2 (`aws-sdk-go-v2/service/s3`), PathStyle = true。
**Bucket**: `cms-media`, serve.go 启动时检查/创建。
**对象 Key 格式**: `media/{year}/{month}/{uuid}.{ext}`

### 2.4 pkg/cache — Redis 缓存 (重构)

从 `internal/database/redis.go` 迁移到 `pkg/cache`:

```
pkg/cache/
├── client.go    # Redis 客户端包装
└── cache.go     # 通用缓存操作 (Get/Set/Del with TTL)
```

用于 Categories 和 Tags 的 `post_count` 缓存 (TTL 60s)。

### 2.5 database/meilisearch.go — 连接工厂

在 `internal/database/` 中新增:

```go
func NewMeilisearch(cfg *config.Config) (*search.Client, error)
```

Config 新增字段: `MEILISEARCH_HOST`, `MEILISEARCH_API_KEY`。

---

## 3. Categories 模块 (6 endpoints)

### 3.1 文件结构

```
internal/category/
├── interfaces.go
├── dto.go
├── repository.go
├── service.go
├── handler.go
├── handler_test.go
└── service_test.go
```

### 3.2 端点清单

| 方法 | 路径 | 权限 | Handler 方法 |
|------|------|------|-------------|
| GET | /api/v1/categories | Viewer+ | List |
| GET | /api/v1/categories/:id | Viewer+ | Get |
| POST | /api/v1/categories | Admin+ | Create |
| PUT | /api/v1/categories/:id | Admin+ | Update |
| DELETE | /api/v1/categories/:id | Admin+ | Delete |
| PUT | /api/v1/categories/reorder | Admin+ | Reorder |

### 3.3 Repository 接口

```go
type CategoryRepository interface {
    List(ctx context.Context) ([]model.Category, error)
    GetByID(ctx context.Context, id string) (*model.Category, error)
    GetChildren(ctx context.Context, parentID string) ([]model.Category, error)
    Create(ctx context.Context, cat *model.Category) error
    Update(ctx context.Context, cat *model.Category) error
    Delete(ctx context.Context, id string) error
    SlugExistsUnderParent(ctx context.Context, slug string, parentID *string, excludeID string) (bool, error)
    UpdatePathPrefix(ctx context.Context, oldPrefix, newPrefix string) (int64, error)
    BatchUpdateSortOrder(ctx context.Context, orders []SortOrderItem) error
}
```

### 3.4 树形结构构建

- `List()` 全量查询所有分类 (通常 <200 条)
- Service 层在内存按 `parent_id` 组装树
- `post_count` 通过 COUNT 查询 `sfc_site_post_category_map`, Redis 缓存 60s

### 3.5 Path 级联更新

当 `slug` 或 `parent_id` 变更时:
1. 遍历 parent 链计算新 path: `/{parent_slug}/.../slug/`
2. `UPDATE sfc_site_categories SET path = REPLACE(path, old_prefix, new_prefix) WHERE path LIKE old_prefix || '%'`
3. **循环引用检测**: 更新前验证新 parent 不是自身或自身的后代

### 3.6 Delete 规则

- 仅允许删除叶子分类 (无子分类), 否则 409 Conflict
- 删除时清除 `sfc_site_post_category_map` 关联 (CASCADE 已配置)
- `post.primary_category_id` 不在本模块处理

### 3.7 Reorder

- 接收 `[{id, sort_order}]` 数组
- 事务中批量 UPDATE sort_order
- 允许部分更新 (不要求包含全部分类)

---

## 4. Tags 模块 (6 endpoints)

### 4.1 文件结构

```
internal/tag/
├── interfaces.go
├── dto.go
├── repository.go
├── service.go
├── handler.go
├── handler_test.go
└── service_test.go
```

### 4.2 端点清单

| 方法 | 路径 | 权限 | Handler 方法 |
|------|------|------|-------------|
| GET | /api/v1/tags | Viewer+ | List |
| GET | /api/v1/tags/:id | Viewer+ | Get |
| GET | /api/v1/tags/suggest | Viewer+ | Suggest |
| POST | /api/v1/tags | Editor+ | Create |
| PUT | /api/v1/tags/:id | Editor+ | Update |
| DELETE | /api/v1/tags/:id | Editor+ | Delete |

### 4.3 Repository 接口

```go
type TagRepository interface {
    List(ctx context.Context, filter ListFilter) ([]model.Tag, int64, error)
    GetByID(ctx context.Context, id string) (*model.Tag, error)
    Create(ctx context.Context, tag *model.Tag) error
    Update(ctx context.Context, tag *model.Tag) error
    Delete(ctx context.Context, id string) error
    SlugExists(ctx context.Context, slug string, excludeID string) (bool, error)
    NameExists(ctx context.Context, name string, excludeID string) (bool, error)
}
```

### 4.4 Meilisearch 同步策略

**索引文档**:
```json
{
  "id": "uuid",
  "name": "Go",
  "slug": "go"
}
```

**同步时机**:
- Create/Update/Delete → DB 同步写入后, goroutine 异步推送 Meilisearch
- Meilisearch 不可用时 slog 记录错误, 不阻塞 CRUD

**EnsureIndex**: 首次访问或站点创建时调用, searchableAttributes = ["name", "slug"]

### 4.5 Suggest 端点

`GET /api/v1/tags/suggest?q=Go`
- 调用 `pkg/search.Searcher.Search("tags-{siteSlug}", q, &SearchOpts{Limit: 10})`
- Meilisearch 内置 typo tolerance + prefix search
- 返回最多 10 个匹配结果

### 4.6 post_count

与 Categories 一致: COUNT 查询 `sfc_site_post_tag_map`, Redis 缓存 60s。

### 4.7 Delete

- DB CASCADE 自动清理 `sfc_site_post_tag_map`
- 同步从 Meilisearch 删除文档
- 无引用检查

---

## 5. Media 模块 (6 endpoints)

### 5.1 文件结构

```
internal/media/
├── interfaces.go
├── dto.go
├── repository.go
├── service.go
├── handler.go
├── handler_test.go
└── service_test.go
```

### 5.2 端点清单

| 方法 | 路径 | 权限 | Handler 方法 |
|------|------|------|-------------|
| POST | /api/v1/media | Editor+ | Upload |
| GET | /api/v1/media | Viewer+ | List |
| GET | /api/v1/media/:id | Viewer+ | Get |
| PUT | /api/v1/media/:id | Editor+ | Update |
| DELETE | /api/v1/media/:id | Editor+ | Delete |
| DELETE | /api/v1/media/batch | Admin+ | BatchDelete |

### 5.3 Repository 接口

```go
type MediaRepository interface {
    List(ctx context.Context, filter ListFilter) ([]model.MediaFile, int64, error)
    GetByID(ctx context.Context, id string) (*model.MediaFile, error)
    Create(ctx context.Context, mf *model.MediaFile) error
    Update(ctx context.Context, mf *model.MediaFile) error
    SoftDelete(ctx context.Context, id string) error
    BatchSoftDelete(ctx context.Context, ids []string) (int64, error)
    GetReferencingPosts(ctx context.Context, mediaID string) ([]PostRef, error)
    GetBatchReferencingPosts(ctx context.Context, mediaIDs []string) (map[string]int, error)
}

// PostRef 引用文章摘要 (用于 409 响应)
type PostRef struct {
    ID    string `json:"id"`
    Title string `json:"title"`
}
```

### 5.4 Upload 流程

```
POST /api/v1/media (multipart/form-data)
  1. Handler: 解析 multipart file
     - 校验文件大小 ≤ 50MB
     - 校验 MIME 白名单
  2. Service:
     a. 生成 storage key: media/{year}/{month}/{uuid}.{ext}
     b. 判断 MediaType (image/video/audio/document/other)
     c. 如果 MediaType == image:
        - pkg/imaging.ExtractDimensions → width, height
        - pkg/imaging.ToWebP(quality=80) → webp_data → Upload to RustFS
        - pkg/imaging.Thumbnail(150, 150) → Upload to RustFS (thumbs/sm_{key})
        - pkg/imaging.Thumbnail(400, 400) → Upload to RustFS (thumbs/md_{key})
     d. Upload 原文件到 RustFS
     e. 创建 DB 记录 (含所有 URL)
  3. 返回 MediaFile 响应
```

**MIME 白名单**:
- 图片: `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `image/svg+xml`
- 视频: `video/mp4`, `video/webm`
- 音频: `audio/mpeg`, `audio/ogg`, `audio/wav`
- 文档: `application/pdf`, `application/msword`, `application/vnd.openxmlformats-officedocument.*`

### 5.5 Delete 逻辑

**单个删除 `DELETE /api/v1/media/:id`**:
1. 查询 reference_count
2. 如果 > 0 且无 `?force=true` → 409 + 引用文章列表
3. `force=true` 需要 Admin+ 权限
4. 软删除 DB 记录 (设 deleted_at)
5. 不立即删除 RustFS 文件 (物理清理由后续 cron 处理)

**批量删除 `DELETE /api/v1/media/batch`**:
1. 校验 ids 长度 ≤ 100
2. 无 force → 逐个检查引用, 跳过有引用的 (返回 skipped 列表)
3. `force=true` → Admin+, 全部软删除
4. 返回 `{deleted_count, skipped}`

### 5.6 Update 端点

仅更新元数据 (alt_text 等), 不支持替换文件。

### 5.7 RustFS 文件物理清理

不在 Batch 5 范围, 属于后续 cron 模块。

---

## 6. Router 注册

```go
// 站点级路由组内新增 (在 siteGroup 下)

// Categories
categories := siteGroup.Group("/categories")
categories.GET("", categoryHandler.List)
categories.PUT("/reorder", categoryHandler.Reorder)  // 固定路径在 /:id 之前
categories.GET("/:id", categoryHandler.Get)
categories.POST("", categoryHandler.Create)
categories.PUT("/:id", categoryHandler.Update)
categories.DELETE("/:id", categoryHandler.Delete)

// Tags
tags := siteGroup.Group("/tags")
tags.GET("", tagHandler.List)
tags.GET("/suggest", tagHandler.Suggest)  // 固定路径在 /:id 之前
tags.GET("/:id", tagHandler.Get)
tags.POST("", tagHandler.Create)
tags.PUT("/:id", tagHandler.Update)
tags.DELETE("/:id", tagHandler.Delete)

// Media
media := siteGroup.Group("/media")
media.GET("", mediaHandler.List)
media.DELETE("/batch", mediaHandler.BatchDelete)  // 固定路径在 /:id 之前
media.POST("", mediaHandler.Upload)
media.GET("/:id", mediaHandler.Get)
media.PUT("/:id", mediaHandler.Update)
media.DELETE("/:id", mediaHandler.Delete)
```

**路由顺序**: Gin 要求固定路径 (`/reorder`, `/suggest`, `/batch`) 注册在参数路径 (`/:id`) 之前。

**API Registry**: `BuildAPIMetaMap()` 新增 18 条 RBAC 元数据, 总计 46 → 64 条。

---

## 7. DI 依赖注入 (serve.go)

```
Config
  → database.NewMeilisearch()  → search.Client (Indexer + Searcher)
  → storage.NewS3Client()      → storage.StorageClient
  → imaging.NewProcessor()     → imaging.Processor
  → cache.NewClient()          → cache.Client (重构自 database/redis.go)

search.Client     → tag.NewService(tagRepo, searchClient, cacheClient)
storage.Client    → media.NewService(mediaRepo, storageClient, processor, cacheClient)
imaging.Processor → media.NewService(...)
cache.Client      → category.NewService(catRepo, cacheClient)
```

---

## 8. 测试策略

| 层 | 测试方式 | 覆盖范围 |
|---|---|---|
| Service | 单元测试 + mock repo/search/storage/cache | 业务逻辑, 树构建, path, 引用检查 |
| Handler | HTTP 测试 + mock service | 请求解析, 参数校验, 响应格式 |
| Repository | 可选集成测试 (testcontainers) | SQL 正确性 |

**重点测试场景**:
- Categories: 树构建算法, 循环引用检测, path 级联更新, reorder 事务
- Tags: CRUD + Meilisearch 异步回调 (mock Indexer), suggest 查询
- Media: 上传流程 (mock StorageClient + Processor), 引用计数逻辑, force 删除, batch 删除 skipped 返回

---

## 9. 实施顺序

```
Step 1: 共享基础设施
  → pkg/search  (Meilisearch 客户端)
  → pkg/imaging (图片处理)
  → pkg/storage (RustFS S3 客户端)
  → pkg/cache   (Redis 缓存, 重构自 database/redis.go)
  → database/meilisearch.go (连接工厂)

Step 2: Categories (6 endpoints)
  → interfaces → dto → repository → service → handler → tests

Step 3: Tags (6 endpoints)
  → interfaces → dto → repository → service → handler → tests

Step 4: Media (6 endpoints)
  → interfaces → dto → repository → service → handler → tests

Step 5: Router 注册 + API Registry 更新

Step 6: 全量测试 + go vet
```

**总计**: 4 个新 pkg/ 包 + 3 个业务模块 + router 更新 = 18 个新端点, 路由总数 63 → 81

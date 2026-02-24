# Batch 5: Categories + Tags + Media — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement 18 site-scoped endpoints (Categories 6 + Tags 6 + Media 6) with shared infrastructure packages for Meilisearch, image processing, S3 storage, and Redis caching.

**Architecture:** Follows existing Batch 4 patterns — each module has interfaces.go, dto.go, repository.go, service.go, handler.go with _test files. Four new `internal/pkg/` packages provide reusable infrastructure. Site-scoped routes mount under the existing `v1.Group("/site")` route group with SiteResolver + Schema + AuditContext + Auth + RBAC middleware chain.

**Tech Stack:** Go 1.25+, Gin, uptrace/bun, meilisearch-go v0.36.1, AWS SDK v2 (S3), go-redis/v9, golang.org/x/image (WebP)

**Key reference files:**
- Design: `docs/plans/2026-02-24-batch5-content-taxonomy-media-design.md`
- API spec: `docs/api.md` sections 7 (Categories), 8 (Tags), 9 (Media)
- DB schema: `docs/database.md` tables 5-9
- Existing patterns: `internal/user/` (handler/service/repo/dto/interfaces), `internal/apikey/`
- Router: `internal/router/router.go` — site-scoped group at line 231
- API Registry: `internal/router/api_meta.go`
- Models: `internal/model/category.go`, `internal/model/tag.go`, `internal/model/media.go`
- Response helpers: `internal/pkg/response/response.go`
- Error types: `internal/pkg/apperror/errors.go`

**Route prefix note:** All site-scoped routes use `/api/v1/site/` prefix (e.g., `GET:/api/v1/site/categories`), matching existing Batch 4 pattern (`/api/v1/site/settings`, `/api/v1/site/api-keys`).

---

## Task 1: Add golang.org/x/image dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add the dependency**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && go get golang.org/x/image@latest
```

**Step 2: Verify**

```bash
grep "golang.org/x/image" go.mod
```

Expected: `golang.org/x/image v0.x.x`

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add golang.org/x/image for WebP encoding"
```

---

## Task 2: pkg/search — Meilisearch client wrapper

**Files:**
- Create: `internal/pkg/search/client.go`
- Create: `internal/pkg/search/client_test.go`

**Step 1: Write the search client package**

`internal/pkg/search/client.go`:
```go
package search

import (
	"context"
	"fmt"

	"github.com/meilisearch/meilisearch-go"
)

// Client wraps meilisearch.ServiceManager with simplified interfaces.
type Client struct {
	ms meilisearch.ServiceManager
}

// NewClient creates a Meilisearch client wrapper.
// Pass nil for ms to create a no-op client (graceful degradation).
func NewClient(ms meilisearch.ServiceManager) *Client {
	return &Client{ms: ms}
}

// Available returns true if the Meilisearch connection is live.
func (c *Client) Available() bool {
	return c.ms != nil && c.ms.IsHealthy()
}

// IndexSettings configures a Meilisearch index.
type IndexSettings struct {
	SearchableAttributes []string
	DisplayedAttributes  []string
	FilterableAttributes []string
	SortableAttributes   []string
}

// EnsureIndex creates an index if it doesn't exist and applies settings.
func (c *Client) EnsureIndex(ctx context.Context, uid string, settings *IndexSettings) error {
	if c.ms == nil {
		return nil
	}

	_, err := c.ms.CreateIndex(&meilisearch.IndexConfig{Uid: uid, PrimaryKey: "id"})
	if err != nil {
		// Index may already exist — not an error.
	}

	if settings != nil {
		s := &meilisearch.Settings{}
		if len(settings.SearchableAttributes) > 0 {
			s.SearchableAttributes = settings.SearchableAttributes
		}
		if len(settings.DisplayedAttributes) > 0 {
			s.DisplayedAttributes = settings.DisplayedAttributes
		}
		if len(settings.FilterableAttributes) > 0 {
			s.FilterableAttributes = settings.FilterableAttributes
		}
		if len(settings.SortableAttributes) > 0 {
			s.SortableAttributes = settings.SortableAttributes
		}
		if _, err := c.ms.Index(uid).UpdateSettings(s); err != nil {
			return fmt.Errorf("update index settings %s: %w", uid, err)
		}
	}
	return nil
}

// UpsertDocuments adds or updates documents in an index.
func (c *Client) UpsertDocuments(ctx context.Context, uid string, docs any) error {
	if c.ms == nil {
		return nil
	}
	_, err := c.ms.Index(uid).AddDocuments(docs, "id")
	if err != nil {
		return fmt.Errorf("upsert documents %s: %w", uid, err)
	}
	return nil
}

// DeleteDocuments removes documents by ID from an index.
func (c *Client) DeleteDocuments(ctx context.Context, uid string, ids []string) error {
	if c.ms == nil {
		return nil
	}
	_, err := c.ms.Index(uid).DeleteDocuments(ids)
	if err != nil {
		return fmt.Errorf("delete documents %s: %w", uid, err)
	}
	return nil
}

// SearchOpts configures a search query.
type SearchOpts struct {
	Limit  int64
	Offset int64
	Filter string
}

// SearchResult contains search results.
type SearchResult struct {
	Hits           []map[string]any
	EstimatedTotal int64
}

// Search queries an index.
func (c *Client) Search(ctx context.Context, uid, query string, opts *SearchOpts) (*SearchResult, error) {
	if c.ms == nil {
		return &SearchResult{}, nil
	}

	req := &meilisearch.SearchRequest{Query: query}
	if opts != nil {
		if opts.Limit > 0 {
			req.Limit = opts.Limit
		}
		if opts.Offset > 0 {
			req.Offset = opts.Offset
		}
		if opts.Filter != "" {
			req.Filter = opts.Filter
		}
	}

	resp, err := c.ms.Index(uid).Search(query, req)
	if err != nil {
		return nil, fmt.Errorf("search %s: %w", uid, err)
	}

	hits := make([]map[string]any, len(resp.Hits))
	for i, h := range resp.Hits {
		if m, ok := h.(map[string]any); ok {
			hits[i] = m
		}
	}

	return &SearchResult{
		Hits:           hits,
		EstimatedTotal: resp.EstimatedTotalHits,
	}, nil
}

// DeleteIndex removes an entire index (used when deleting a site).
func (c *Client) DeleteIndex(ctx context.Context, uid string) error {
	if c.ms == nil {
		return nil
	}
	_, err := c.ms.DeleteIndex(uid)
	if err != nil {
		return fmt.Errorf("delete index %s: %w", uid, err)
	}
	return nil
}
```

**Step 2: Write a basic test**

`internal/pkg/search/client_test.go`:
```go
package search_test

import (
	"testing"

	"github.com/sky-flux/cms/internal/pkg/search"
	"github.com/stretchr/testify/assert"
)

func TestNewClient_NilMS_Available(t *testing.T) {
	c := search.NewClient(nil)
	assert.False(t, c.Available())
}

func TestNewClient_NilMS_GracefulDegradation(t *testing.T) {
	c := search.NewClient(nil)

	// All operations should be no-ops when ms is nil.
	err := c.EnsureIndex(t.Context(), "test-index", nil)
	assert.NoError(t, err)

	err = c.UpsertDocuments(t.Context(), "test-index", []map[string]any{{"id": "1"}})
	assert.NoError(t, err)

	err = c.DeleteDocuments(t.Context(), "test-index", []string{"1"})
	assert.NoError(t, err)

	result, err := c.Search(t.Context(), "test-index", "query", nil)
	assert.NoError(t, err)
	assert.Empty(t, result.Hits)
}
```

**Step 3: Run tests**

```bash
go test ./internal/pkg/search/... -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/pkg/search/
git commit -m "feat(pkg): add search package wrapping Meilisearch client"
```

---

## Task 3: pkg/imaging — Image processing

**Files:**
- Create: `internal/pkg/imaging/processor.go`
- Create: `internal/pkg/imaging/processor_test.go`

**Step 1: Write the imaging package**

`internal/pkg/imaging/processor.go`:
```go
package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

// Processor handles image operations.
type Processor struct{}

// NewProcessor creates a new image processor.
func NewProcessor() *Processor {
	return &Processor{}
}

// ExtractDimensions returns the width and height of an image.
func (p *Processor) ExtractDimensions(src io.Reader) (width, height int, err error) {
	cfg, _, err := image.DecodeConfig(src)
	if err != nil {
		return 0, 0, fmt.Errorf("decode image config: %w", err)
	}
	return cfg.Width, cfg.Height, nil
}

// ToWebP converts an image to WebP format.
// quality: 1-100 (higher = better quality, larger file).
func (p *Processor) ToWebP(src io.Reader, quality int) ([]byte, error) {
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	var buf bytes.Buffer
	opts := &webp.Options{Quality: float32(quality)}
	if err := webp.Encode(&buf, img, opts); err != nil {
		return nil, fmt.Errorf("encode webp: %w", err)
	}
	return buf.Bytes(), nil
}

// Thumbnail generates a thumbnail of the given dimensions.
// mode: "crop" for center-crop to exact size, "fit" for fit-within bounds.
func (p *Processor) Thumbnail(src io.Reader, width, height int, mode string) ([]byte, error) {
	img, format, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	var dst *image.RGBA
	srcBounds := img.Bounds()

	if mode == "crop" {
		dst = p.cropCenter(img, srcBounds, width, height)
	} else {
		dst = p.fitWithin(img, srcBounds, width, height)
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 80})
	case "png":
		err = png.Encode(&buf, dst)
	case "gif":
		err = gif.Encode(&buf, dst, nil)
	default:
		err = png.Encode(&buf, dst)
	}
	if err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}
	return buf.Bytes(), nil
}

// cropCenter crops from the center to exact dimensions.
func (p *Processor) cropCenter(img image.Image, bounds image.Rectangle, w, h int) *image.RGBA {
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	// Determine crop region.
	srcRatio := float64(srcW) / float64(srcH)
	dstRatio := float64(w) / float64(h)

	var cropRect image.Rectangle
	if srcRatio > dstRatio {
		cropH := srcH
		cropW := int(float64(cropH) * dstRatio)
		x0 := (srcW - cropW) / 2
		cropRect = image.Rect(x0, 0, x0+cropW, cropH)
	} else {
		cropW := srcW
		cropH := int(float64(cropW) / dstRatio)
		y0 := (srcH - cropH) / 2
		cropRect = image.Rect(0, y0, cropW, y0+cropH)
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, cropRect, draw.Over, nil)
	return dst
}

// fitWithin scales to fit within bounds, maintaining aspect ratio.
func (p *Processor) fitWithin(img image.Image, bounds image.Rectangle, maxW, maxH int) *image.RGBA {
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	ratio := min(float64(maxW)/float64(srcW), float64(maxH)/float64(srcH))
	if ratio >= 1 {
		ratio = 1
	}

	dstW := int(float64(srcW) * ratio)
	dstH := int(float64(srcH) * ratio)

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst
}
```

**Step 2: Write tests**

`internal/pkg/imaging/processor_test.go`:
```go
package imaging_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/sky-flux/cms/internal/pkg/imaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testImage(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := range w {
		for y := range h {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestExtractDimensions(t *testing.T) {
	p := imaging.NewProcessor()
	data := testImage(800, 600)

	w, h, err := p.ExtractDimensions(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, 800, w)
	assert.Equal(t, 600, h)
}

func TestThumbnail_Crop(t *testing.T) {
	p := imaging.NewProcessor()
	data := testImage(800, 600)

	thumb, err := p.Thumbnail(bytes.NewReader(data), 150, 150, "crop")
	require.NoError(t, err)
	assert.NotEmpty(t, thumb)

	cfg, _, err := image.DecodeConfig(bytes.NewReader(thumb))
	require.NoError(t, err)
	assert.Equal(t, 150, cfg.Width)
	assert.Equal(t, 150, cfg.Height)
}

func TestThumbnail_Fit(t *testing.T) {
	p := imaging.NewProcessor()
	data := testImage(800, 600)

	thumb, err := p.Thumbnail(bytes.NewReader(data), 400, 400, "fit")
	require.NoError(t, err)
	assert.NotEmpty(t, thumb)

	cfg, _, err := image.DecodeConfig(bytes.NewReader(thumb))
	require.NoError(t, err)
	assert.LessOrEqual(t, cfg.Width, 400)
	assert.LessOrEqual(t, cfg.Height, 400)
}
```

**Step 3: Run tests**

```bash
go test ./internal/pkg/imaging/... -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/pkg/imaging/
git commit -m "feat(pkg): add imaging package for WebP conversion and thumbnails"
```

---

## Task 4: pkg/storage — S3/RustFS client wrapper

**Files:**
- Create: `internal/pkg/storage/client.go`
- Create: `internal/pkg/storage/client_test.go`

**Step 1: Write the storage package**

`internal/pkg/storage/client.go`:
```go
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Client wraps an S3-compatible client for object storage operations.
type Client struct {
	s3     *s3.Client
	bucket string
	cdnURL string // public URL prefix (e.g., "https://cdn.example.com")
}

// NewClient creates a storage client.
// cdnURL is the public base URL for accessing files (e.g., "http://localhost:9000/cms-media").
func NewClient(s3Client *s3.Client, bucket, cdnURL string) *Client {
	return &Client{s3: s3Client, bucket: bucket, cdnURL: cdnURL}
}

// Available returns true if the S3 client is configured.
func (c *Client) Available() bool {
	return c.s3 != nil
}

// Upload stores an object in S3.
func (c *Client) Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) error {
	if c.s3 == nil {
		return fmt.Errorf("storage client not available")
	}

	input := &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		Body:          data,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}

	if _, err := c.s3.PutObject(ctx, input); err != nil {
		return fmt.Errorf("upload %s: %w", key, err)
	}
	return nil
}

// UploadBytes is a convenience wrapper for uploading byte slices.
func (c *Client) UploadBytes(ctx context.Context, key string, data []byte, contentType string) error {
	return c.Upload(ctx, key, bytes.NewReader(data), contentType, int64(len(data)))
}

// Delete removes a single object.
func (c *Client) Delete(ctx context.Context, key string) error {
	if c.s3 == nil {
		return nil
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	if _, err := c.s3.DeleteObject(ctx, input); err != nil {
		return fmt.Errorf("delete %s: %w", key, err)
	}
	return nil
}

// BatchDelete removes multiple objects.
func (c *Client) BatchDelete(ctx context.Context, keys []string) error {
	if c.s3 == nil || len(keys) == 0 {
		return nil
	}

	objects := make([]types.ObjectIdentifier, len(keys))
	for i, k := range keys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(k)}
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(c.bucket),
		Delete: &types.Delete{Objects: objects, Quiet: aws.Bool(true)},
	}

	if _, err := c.s3.DeleteObjects(ctx, input); err != nil {
		return fmt.Errorf("batch delete: %w", err)
	}
	return nil
}

// PublicURL returns the public URL for an object key.
func (c *Client) PublicURL(key string) string {
	return c.cdnURL + "/" + key
}

// EnsureBucket creates the bucket if it doesn't exist.
func (c *Client) EnsureBucket(ctx context.Context) error {
	if c.s3 == nil {
		return nil
	}

	_, err := c.s3.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(c.bucket)})
	if err == nil {
		return nil // bucket exists
	}

	_, err = c.s3.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(c.bucket)})
	if err != nil {
		return fmt.Errorf("create bucket %s: %w", c.bucket, err)
	}
	return nil
}
```

**Step 2: Write test**

`internal/pkg/storage/client_test.go`:
```go
package storage_test

import (
	"testing"

	"github.com/sky-flux/cms/internal/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewClient_NilS3(t *testing.T) {
	c := storage.NewClient(nil, "test", "http://localhost:9000/test")
	assert.False(t, c.Available())
}

func TestPublicURL(t *testing.T) {
	c := storage.NewClient(nil, "cms-media", "http://localhost:9000/cms-media")
	url := c.PublicURL("media/2026/02/abc.jpg")
	assert.Equal(t, "http://localhost:9000/cms-media/media/2026/02/abc.jpg", url)
}
```

**Step 3: Run tests**

```bash
go test ./internal/pkg/storage/... -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/pkg/storage/
git commit -m "feat(pkg): add storage package wrapping S3-compatible client"
```

---

## Task 5: pkg/cache — Redis cache (refactor from database/redis.go)

**Files:**
- Create: `internal/pkg/cache/client.go`
- Create: `internal/pkg/cache/client_test.go`
- Modify: `internal/database/redis.go` — keep as-is (connection factory), cache ops go to pkg

**Design note:** `internal/database/redis.go` stays as the connection factory (returns `*redis.Client`). `pkg/cache` provides typed cache operations on top of `*redis.Client`. No migration needed — this is purely additive.

**Step 1: Write the cache package**

`internal/pkg/cache/client.go`:
```go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps Redis with typed cache operations.
type Client struct {
	rdb *redis.Client
}

// NewClient creates a cache client. Pass nil for rdb to disable caching.
func NewClient(rdb *redis.Client) *Client {
	return &Client{rdb: rdb}
}

// Get retrieves a cached value and unmarshals it into dest.
// Returns false if key doesn't exist or cache is unavailable.
func (c *Client) Get(ctx context.Context, key string, dest any) (bool, error) {
	if c.rdb == nil {
		return false, nil
	}

	data, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cache get %s: %w", key, err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, fmt.Errorf("cache unmarshal %s: %w", key, err)
	}
	return true, nil
}

// Set stores a value in cache with the given TTL.
func (c *Client) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}

	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("cache marshal %s: %w", key, err)
	}

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set %s: %w", key, err)
	}
	return nil
}

// Del removes a key from cache.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	if c.rdb == nil || len(keys) == 0 {
		return nil
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("cache del: %w", err)
	}
	return nil
}

// GetOrSet retrieves from cache; on miss, calls fn to produce the value, caches it, and returns it.
func (c *Client) GetOrSet(ctx context.Context, key string, dest any, ttl time.Duration, fn func() (any, error)) error {
	found, err := c.Get(ctx, key, dest)
	if err != nil {
		// Cache read error — fall through to fn.
	}
	if found {
		return nil
	}

	val, err := fn()
	if err != nil {
		return err
	}

	// Marshal val into dest.
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("cache getorset marshal: %w", err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("cache getorset unmarshal: %w", err)
	}

	_ = c.Set(ctx, key, val, ttl)
	return nil
}
```

**Step 2: Write test with miniredis**

`internal/pkg/cache/client_test.go`:
```go
package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCache(t *testing.T) *cache.Client {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return cache.NewClient(rdb)
}

func TestCache_SetAndGet(t *testing.T) {
	c := setupCache(t)
	ctx := context.Background()

	err := c.Set(ctx, "test:key", map[string]int{"count": 42}, time.Minute)
	require.NoError(t, err)

	var result map[string]int
	found, err := c.Get(ctx, "test:key", &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, 42, result["count"])
}

func TestCache_GetMiss(t *testing.T) {
	c := setupCache(t)
	ctx := context.Background()

	var result string
	found, err := c.Get(ctx, "nonexistent", &result)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestCache_NilClient(t *testing.T) {
	c := cache.NewClient(nil)
	ctx := context.Background()

	err := c.Set(ctx, "k", "v", time.Minute)
	assert.NoError(t, err)

	var v string
	found, err := c.Get(ctx, "k", &v)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestCache_Del(t *testing.T) {
	c := setupCache(t)
	ctx := context.Background()

	_ = c.Set(ctx, "del:key", "value", time.Minute)
	err := c.Del(ctx, "del:key")
	require.NoError(t, err)

	var v string
	found, _ := c.Get(ctx, "del:key", &v)
	assert.False(t, found)
}
```

**Step 3: Run tests**

```bash
go test ./internal/pkg/cache/... -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/pkg/cache/
git commit -m "feat(pkg): add cache package with typed Redis operations"
```

---

## Task 6: Category — interfaces + dto

**Files:**
- Modify: `internal/category/dto.go` (replace stub)
- Create: `internal/category/interfaces.go`

**Step 1: Write interfaces**

`internal/category/interfaces.go`:
```go
package category

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// CategoryRepository handles sfc_site_categories table operations.
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
	CountPosts(ctx context.Context, categoryID string) (int64, error)
}

// SortOrderItem represents a sort order update for a single category.
type SortOrderItem struct {
	ID        string `json:"id" binding:"required"`
	SortOrder int    `json:"sort_order"`
}
```

**Step 2: Write DTOs**

Replace `internal/category/dto.go`:
```go
package category

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

type CreateCategoryReq struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Slug        string  `json:"slug" binding:"required,max=200"`
	ParentID    *string `json:"parent_id"`
	Description string  `json:"description"`
	SortOrder   int     `json:"sort_order"`
}

type UpdateCategoryReq struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Slug        *string `json:"slug" binding:"omitempty,max=200"`
	ParentID    *string `json:"parent_id"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}

type ReorderReq struct {
	Orders []SortOrderItem `json:"orders" binding:"required,min=1"`
}

// --- Response DTOs ---

type CategoryResp struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Path        string          `json:"path"`
	ParentID    *string         `json:"parent_id,omitempty"`
	Description string          `json:"description,omitempty"`
	PostCount   int64           `json:"post_count"`
	SortOrder   int             `json:"sort_order"`
	Children    []*CategoryResp `json:"children,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func ToCategoryResp(c *model.Category, postCount int64) CategoryResp {
	return CategoryResp{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Path:        c.Path,
		ParentID:    c.ParentID,
		Description: c.Description,
		PostCount:   postCount,
		SortOrder:   c.SortOrder,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
```

**Step 3: Verify compilation**

```bash
go build ./internal/category/...
```

Expected: no errors

**Step 4: Commit**

```bash
git add internal/category/
git commit -m "feat(category): add interfaces and DTOs"
```

---

## Task 7: Category — repository

**Files:**
- Modify: `internal/category/repository.go` (replace stub)

**Step 1: Write repository**

Replace `internal/category/repository.go`:
```go
package category

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

type Repo struct {
	db *bun.DB
}

func NewRepo(db *bun.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) List(ctx context.Context) ([]model.Category, error) {
	var cats []model.Category
	err := r.db.NewSelect().Model(&cats).OrderExpr("sort_order ASC, created_at ASC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("category list: %w", err)
	}
	return cats, nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (*model.Category, error) {
	cat := new(model.Category)
	err := r.db.NewSelect().Model(cat).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("category not found", err)
		}
		return nil, fmt.Errorf("category get by id: %w", err)
	}
	return cat, nil
}

func (r *Repo) GetChildren(ctx context.Context, parentID string) ([]model.Category, error) {
	var cats []model.Category
	err := r.db.NewSelect().Model(&cats).Where("parent_id = ?", parentID).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("category get children: %w", err)
	}
	return cats, nil
}

func (r *Repo) Create(ctx context.Context, cat *model.Category) error {
	_, err := r.db.NewInsert().Model(cat).Exec(ctx)
	if err != nil {
		return fmt.Errorf("category create: %w", err)
	}
	return nil
}

func (r *Repo) Update(ctx context.Context, cat *model.Category) error {
	_, err := r.db.NewUpdate().Model(cat).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("category update: %w", err)
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.Category)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("category delete: %w", err)
	}
	return nil
}

func (r *Repo) SlugExistsUnderParent(ctx context.Context, slug string, parentID *string, excludeID string) (bool, error) {
	q := r.db.NewSelect().Model((*model.Category)(nil)).Where("slug = ?", slug)
	if parentID != nil {
		q = q.Where("parent_id = ?", *parentID)
	} else {
		q = q.Where("parent_id IS NULL")
	}
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("category slug exists: %w", err)
	}
	return exists, nil
}

func (r *Repo) UpdatePathPrefix(ctx context.Context, oldPrefix, newPrefix string) (int64, error) {
	res, err := r.db.NewUpdate().
		Model((*model.Category)(nil)).
		Set("path = REPLACE(path, ?, ?)", oldPrefix, newPrefix).
		Where("path LIKE ?", oldPrefix+"%").
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("category update path prefix: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *Repo) BatchUpdateSortOrder(ctx context.Context, orders []SortOrderItem) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, o := range orders {
			_, err := tx.NewUpdate().Model((*model.Category)(nil)).
				Set("sort_order = ?", o.SortOrder).
				Where("id = ?", o.ID).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("category batch sort order %s: %w", o.ID, err)
			}
		}
		return nil
	})
}

func (r *Repo) CountPosts(ctx context.Context, categoryID string) (int64, error) {
	count, err := r.db.NewSelect().
		TableExpr("sfc_site_post_category_map").
		Where("category_id = ?", categoryID).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("category count posts: %w", err)
	}
	return int64(count), nil
}
```

**Step 2: Verify compilation**

```bash
go build ./internal/category/...
```

**Step 3: Commit**

```bash
git add internal/category/repository.go
git commit -m "feat(category): implement repository with bun ORM"
```

---

## Task 8: Category — service + service_test

**Files:**
- Modify: `internal/category/service.go` (replace stub)
- Create: `internal/category/service_test.go`

**Step 1: Write service**

Replace `internal/category/service.go`:
```go
package category

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/cache"
)

const postCountCacheTTL = 60 * time.Second

type Service struct {
	repo    CategoryRepository
	cache   *cache.Client
	auditor audit.Logger
}

func NewService(repo CategoryRepository, cache *cache.Client, auditor audit.Logger) *Service {
	return &Service{repo: repo, cache: cache, auditor: auditor}
}

// ListTree returns the full category tree with post counts.
func (s *Service) ListTree(ctx context.Context) ([]*CategoryResp, error) {
	cats, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildTree(ctx, cats), nil
}

// GetCategory returns a single category with children and post count.
func (s *Service) GetCategory(ctx context.Context, id string) (*CategoryResp, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	count := s.getPostCount(ctx, id)
	resp := ToCategoryResp(cat, count)

	children, err := s.repo.GetChildren(ctx, id)
	if err == nil {
		for _, ch := range children {
			chCount := s.getPostCount(ctx, ch.ID)
			chResp := ToCategoryResp(&ch, chCount)
			resp.Children = append(resp.Children, &chResp)
		}
	}
	return &resp, nil
}

// CreateCategory creates a new category and computes its path.
func (s *Service) CreateCategory(ctx context.Context, req *CreateCategoryReq) (*CategoryResp, error) {
	exists, err := s.repo.SlugExistsUnderParent(ctx, req.Slug, req.ParentID, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.Conflict("slug already exists under this parent", nil)
	}

	path, err := s.computePath(ctx, req.ParentID, req.Slug)
	if err != nil {
		return nil, err
	}

	cat := &model.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		ParentID:    req.ParentID,
		Path:        path,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}

	if err := s.repo.Create(ctx, cat); err != nil {
		return nil, err
	}

	if err := s.auditor.Log(ctx, audit.Entry{
		Action: model.LogActionCreate, ResourceType: "category", ResourceID: cat.ID,
	}); err != nil {
		slog.Error("audit log category create", "error", err)
	}

	resp := ToCategoryResp(cat, 0)
	return &resp, nil
}

// UpdateCategory updates a category and cascades path changes.
func (s *Service) UpdateCategory(ctx context.Context, id string, req *UpdateCategoryReq) (*CategoryResp, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	slugChanged := false
	parentChanged := false
	oldPath := cat.Path

	if req.Name != nil {
		cat.Name = *req.Name
	}
	if req.Slug != nil && *req.Slug != cat.Slug {
		exists, err := s.repo.SlugExistsUnderParent(ctx, *req.Slug, cat.ParentID, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.Conflict("slug already exists under this parent", nil)
		}
		cat.Slug = *req.Slug
		slugChanged = true
	}
	if req.ParentID != nil {
		newParent := req.ParentID
		if newParent != nil && *newParent == "" {
			newParent = nil
		}
		if err := s.detectCycle(ctx, id, newParent); err != nil {
			return nil, err
		}
		cat.ParentID = newParent
		parentChanged = true
	}
	if req.Description != nil {
		cat.Description = *req.Description
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	}

	if slugChanged || parentChanged {
		newPath, err := s.computePath(ctx, cat.ParentID, cat.Slug)
		if err != nil {
			return nil, err
		}
		cat.Path = newPath

		if oldPath != newPath {
			if _, err := s.repo.UpdatePathPrefix(ctx, oldPath, newPath); err != nil {
				return nil, fmt.Errorf("cascade path update: %w", err)
			}
		}
	}

	if err := s.repo.Update(ctx, cat); err != nil {
		return nil, err
	}

	if err := s.auditor.Log(ctx, audit.Entry{
		Action: model.LogActionUpdate, ResourceType: "category", ResourceID: id,
	}); err != nil {
		slog.Error("audit log category update", "error", err)
	}

	count := s.getPostCount(ctx, id)
	resp := ToCategoryResp(cat, count)
	return &resp, nil
}

// DeleteCategory deletes a leaf category.
func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	children, err := s.repo.GetChildren(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return apperror.Conflict("category has children, delete them first", nil)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.auditor.Log(ctx, audit.Entry{
		Action: model.LogActionDelete, ResourceType: "category", ResourceID: id,
	}); err != nil {
		slog.Error("audit log category delete", "error", err)
	}

	return nil
}

// Reorder batch-updates sort_order for categories.
func (s *Service) Reorder(ctx context.Context, req *ReorderReq) error {
	return s.repo.BatchUpdateSortOrder(ctx, req.Orders)
}

// --- Helpers ---

func (s *Service) computePath(ctx context.Context, parentID *string, slug string) (string, error) {
	if parentID == nil {
		return "/" + slug + "/", nil
	}
	parent, err := s.repo.GetByID(ctx, *parentID)
	if err != nil {
		return "", err
	}
	return parent.Path + slug + "/", nil
}

func (s *Service) detectCycle(ctx context.Context, id string, newParentID *string) error {
	if newParentID == nil {
		return nil
	}
	if *newParentID == id {
		return apperror.Validation("category cannot be its own parent", nil)
	}

	current := *newParentID
	visited := map[string]bool{id: true}
	for current != "" {
		if visited[current] {
			return apperror.Validation("circular reference detected", nil)
		}
		visited[current] = true
		parent, err := s.repo.GetByID(ctx, current)
		if err != nil {
			return err
		}
		if parent.ParentID == nil {
			break
		}
		current = *parent.ParentID
	}
	return nil
}

func (s *Service) buildTree(ctx context.Context, cats []model.Category) []*CategoryResp {
	nodeMap := make(map[string]*CategoryResp, len(cats))
	for i := range cats {
		count := s.getPostCount(ctx, cats[i].ID)
		resp := ToCategoryResp(&cats[i], count)
		nodeMap[cats[i].ID] = &resp
	}

	var roots []*CategoryResp
	for i := range cats {
		node := nodeMap[cats[i].ID]
		if cats[i].ParentID != nil {
			if parent, ok := nodeMap[*cats[i].ParentID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}
	return roots
}

func (s *Service) getPostCount(ctx context.Context, categoryID string) int64 {
	cacheKey := "cat:postcount:" + categoryID
	var count int64

	found, _ := s.cache.Get(ctx, cacheKey, &count)
	if found {
		return count
	}

	count, err := s.repo.CountPosts(ctx, categoryID)
	if err != nil {
		return 0
	}

	_ = s.cache.Set(ctx, cacheKey, count, postCountCacheTTL)
	return count
}
```

**Step 2: Write service tests**

`internal/category/service_test.go`:
```go
package category_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/category"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Repository ---

type mockCatRepo struct {
	categories []model.Category
	getByID    *model.Category
	getByIDErr error
	children   []model.Category
	childrenErr error
	slugExists bool
	slugExistsErr error
	createErr  error
	updateErr  error
	deleteErr  error
	pathUpdated int64
	pathUpdateErr error
	postCount  int64
	postCountErr error
}

func (m *mockCatRepo) List(_ context.Context) ([]model.Category, error) {
	return m.categories, nil
}
func (m *mockCatRepo) GetByID(_ context.Context, _ string) (*model.Category, error) {
	return m.getByID, m.getByIDErr
}
func (m *mockCatRepo) GetChildren(_ context.Context, _ string) ([]model.Category, error) {
	return m.children, m.childrenErr
}
func (m *mockCatRepo) Create(_ context.Context, c *model.Category) error {
	if m.createErr == nil {
		c.ID = "new-cat-id"
	}
	return m.createErr
}
func (m *mockCatRepo) Update(_ context.Context, _ *model.Category) error { return m.updateErr }
func (m *mockCatRepo) Delete(_ context.Context, _ string) error          { return m.deleteErr }
func (m *mockCatRepo) SlugExistsUnderParent(_ context.Context, _ string, _ *string, _ string) (bool, error) {
	return m.slugExists, m.slugExistsErr
}
func (m *mockCatRepo) UpdatePathPrefix(_ context.Context, _, _ string) (int64, error) {
	return m.pathUpdated, m.pathUpdateErr
}
func (m *mockCatRepo) BatchUpdateSortOrder(_ context.Context, _ []category.SortOrderItem) error {
	return nil
}
func (m *mockCatRepo) CountPosts(_ context.Context, _ string) (int64, error) {
	return m.postCount, m.postCountErr
}

// --- Tests ---

func newTestService(repo category.CategoryRepository) *category.Service {
	return category.NewService(repo, cache.NewClient(nil), audit.NewNoopLogger())
}

func TestCreateCategory_Success(t *testing.T) {
	repo := &mockCatRepo{}
	svc := newTestService(repo)

	resp, err := svc.CreateCategory(context.Background(), &category.CreateCategoryReq{
		Name: "Tech", Slug: "tech",
	})
	require.NoError(t, err)
	assert.Equal(t, "new-cat-id", resp.ID)
	assert.Equal(t, "/tech/", resp.Path)
}

func TestCreateCategory_DuplicateSlug(t *testing.T) {
	repo := &mockCatRepo{slugExists: true}
	svc := newTestService(repo)

	_, err := svc.CreateCategory(context.Background(), &category.CreateCategoryReq{
		Name: "Tech", Slug: "tech",
	})
	require.Error(t, err)
	assert.True(t, apperror.IsConflict(err))
}

func TestCreateCategory_WithParent(t *testing.T) {
	parentID := "parent-id"
	repo := &mockCatRepo{
		getByID: &model.Category{ID: parentID, Path: "/tech/"},
	}
	svc := newTestService(repo)

	resp, err := svc.CreateCategory(context.Background(), &category.CreateCategoryReq{
		Name: "Backend", Slug: "backend", ParentID: &parentID,
	})
	require.NoError(t, err)
	assert.Equal(t, "/tech/backend/", resp.Path)
}

func TestDeleteCategory_HasChildren(t *testing.T) {
	repo := &mockCatRepo{
		children: []model.Category{{ID: "child-1"}},
	}
	svc := newTestService(repo)

	err := svc.DeleteCategory(context.Background(), "cat-id")
	require.Error(t, err)
	assert.True(t, apperror.IsConflict(err))
}

func TestDeleteCategory_Leaf(t *testing.T) {
	repo := &mockCatRepo{}
	svc := newTestService(repo)

	err := svc.DeleteCategory(context.Background(), "cat-id")
	require.NoError(t, err)
}

func TestBuildTree(t *testing.T) {
	parentID := "p1"
	repo := &mockCatRepo{
		categories: []model.Category{
			{ID: "p1", Name: "Tech", Slug: "tech", Path: "/tech/"},
			{ID: "c1", Name: "Backend", Slug: "backend", Path: "/tech/backend/", ParentID: &parentID},
		},
	}
	svc := newTestService(repo)

	tree, err := svc.ListTree(context.Background())
	require.NoError(t, err)
	require.Len(t, tree, 1)
	assert.Equal(t, "Tech", tree[0].Name)
	require.Len(t, tree[0].Children, 1)
	assert.Equal(t, "Backend", tree[0].Children[0].Name)
}
```

**Step 3: Run tests**

```bash
go test ./internal/category/... -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/category/service.go internal/category/service_test.go
git commit -m "feat(category): implement service with tree building and path cascade"
```

---

## Task 9: Category — handler + handler_test

**Files:**
- Modify: `internal/category/handler.go` (replace stub)
- Modify: `internal/category/handler_test.go` (replace stub)

**Step 1: Write handler**

Replace `internal/category/handler.go`:
```go
package category

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *gin.Context) {
	tree, err := h.svc.ListTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, tree)
}

func (h *Handler) Get(c *gin.Context) {
	cat, err := h.svc.GetCategory(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, cat)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}
	cat, err := h.svc.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, cat)
}

func (h *Handler) Update(c *gin.Context) {
	var req UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}
	cat, err := h.svc.UpdateCategory(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, cat)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.DeleteCategory(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "category deleted"})
}

func (h *Handler) Reorder(c *gin.Context) {
	var req ReorderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}
	if err := h.svc.Reorder(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "categories reordered"})
}
```

**Step 2: Write handler tests**

Replace `internal/category/handler_test.go` — follow the existing pattern from `internal/user/handler_test.go` with HTTP test recorder. (The implementing agent should write 6-8 handler tests covering List, Get, Create success, Create validation error, Update, Delete success, Delete conflict, Reorder.)

**Step 3: Run tests**

```bash
go test ./internal/category/... -v -count=1
```

**Step 4: Commit**

```bash
git add internal/category/handler.go internal/category/handler_test.go
git commit -m "feat(category): implement handler with 6 endpoints"
```

---

## Task 10: Tag — interfaces + dto + repository

**Files:**
- Create: `internal/tag/interfaces.go`
- Create: `internal/tag/dto.go`
- Create: `internal/tag/repository.go`

Follows same patterns as Category. Repository uses standard bun CRUD.

Tag model has no `updated_at` — only `created_at`. Tag `name` and `slug` are UNIQUE.

**Key differences from Category:**
- No tree structure (flat list with pagination)
- ListFilter has `Page`, `PerPage`, `Query`, `Sort`
- `NameExists` + `SlugExists` checks (both UNIQUE)
- `CountPosts` queries `sfc_site_post_tag_map`

**Step 1-3: Write interfaces, dto, repository**

(Implementing agent: follow Category patterns. TagRepository: List with pagination, GetByID, Create, Update, Delete, SlugExists, NameExists, CountPosts.)

**Step 4: Verify compilation**

```bash
go build ./internal/tag/...
```

**Step 5: Commit**

```bash
git add internal/tag/
git commit -m "feat(tag): add interfaces, DTOs, and repository"
```

---

## Task 11: Tag — service + service_test

**Files:**
- Create: `internal/tag/service.go`
- Create: `internal/tag/service_test.go`

**Key logic:**
- CRUD operations with audit logging
- Meilisearch async sync: after Create/Update/Delete, spawn goroutine to push to `tags-{siteSlug}` index
- `Suggest(ctx, siteSlug, query)` calls `search.Client.Search("tags-"+siteSlug, query, &SearchOpts{Limit: 10})`
- `post_count` via Redis cache (same pattern as Category)
- Site slug extracted from `ctx` via `middleware.GetSiteSlug(c)` or passed as parameter

**Meilisearch sync pattern:**
```go
func (s *Service) syncToSearch(siteSlug string, tag *model.Tag) {
    go func() {
        doc := map[string]any{"id": tag.ID, "name": tag.Name, "slug": tag.Slug}
        if err := s.search.UpsertDocuments(context.Background(), "tags-"+siteSlug, []map[string]any{doc}); err != nil {
            slog.Error("meilisearch tag sync failed", "error", err, "tag_id", tag.ID)
        }
    }()
}
```

**Step 1-2: Write service and tests**

(Implementing agent: 8-10 tests covering CRUD + Suggest + duplicate checks.)

**Step 3: Run tests**

```bash
go test ./internal/tag/... -v
```

**Step 4: Commit**

```bash
git add internal/tag/service.go internal/tag/service_test.go
git commit -m "feat(tag): implement service with Meilisearch sync"
```

---

## Task 12: Tag — handler + handler_test

**Files:**
- Create: `internal/tag/handler.go`
- Create: `internal/tag/handler_test.go`

6 handler methods: List, Get, Suggest, Create, Update, Delete.

**Suggest handler** extracts site slug from `c.GetString("site_slug")` (set by SiteResolver middleware) and query param `q`.

**Step 1-2: Write handler and tests**

**Step 3: Run tests and commit**

```bash
go test ./internal/tag/... -v
git add internal/tag/handler.go internal/tag/handler_test.go
git commit -m "feat(tag): implement handler with 6 endpoints including suggest"
```

---

## Task 13: Media — interfaces + dto

**Files:**
- Create: `internal/media/interfaces.go`
- Modify: `internal/media/dto.go` (replace stub)

**Key interfaces:**
- `MediaRepository` — List, GetByID, Create, Update, SoftDelete, BatchSoftDelete, GetReferencingPosts, GetBatchReferencingPosts
- `StorageUploader` — wraps `pkg/storage.Client` methods needed by media service
- `ImageProcessor` — wraps `pkg/imaging.Processor` methods

```go
// PostRef is used in 409 conflict responses.
type PostRef struct {
    ID    string `json:"id"`
    Title string `json:"title"`
}
```

**DTOs:**
- `ListFilter` with Page, PerPage, Query, MediaType
- `UpdateMediaReq` with AltText *string
- `MediaResp` converting model.MediaFile to JSON-safe response
- `BatchDeleteReq` with IDs []string
- `BatchDeleteResp` with DeletedCount + Skipped list

**Step 1-2: Write and verify compilation**

**Step 3: Commit**

```bash
git add internal/media/interfaces.go internal/media/dto.go
git commit -m "feat(media): add interfaces and DTOs"
```

---

## Task 14: Media — repository

**Files:**
- Modify: `internal/media/repository.go` (replace stub)

Standard bun CRUD with soft delete. Key queries:
- `List` with pagination + optional media_type filter + `WHERE deleted_at IS NULL`
- `SoftDelete` uses bun's built-in soft delete (model has `DeletedAt *time.Time` with `soft_delete` tag)
- `BatchSoftDelete` uses `WHERE id IN (?)` with `bun.In(ids)`
- `GetReferencingPosts` joins through `sfc_site_posts` where `cover_image_id = ?` (simple for now)
- `GetBatchReferencingPosts` returns `map[mediaID]->refCount`

**Step 1: Write repository**

**Step 2: Verify compilation and commit**

```bash
go build ./internal/media/...
git add internal/media/repository.go
git commit -m "feat(media): implement repository with soft delete and reference queries"
```

---

## Task 15: Media — service + service_test

**Files:**
- Modify: `internal/media/service.go` (replace stub)
- Create: `internal/media/service_test.go`

**Upload flow (the most complex method):**
```
Upload(ctx, siteSlug, file, altText) -> (*MediaResp, error)
1. Validate MIME type against whitelist
2. Determine MediaType from MIME
3. Generate storage key: "media/{year}/{month}/{uuid}.{ext}"
4. If image:
   a. Read file into memory (for multiple passes)
   b. ExtractDimensions
   c. ToWebP -> upload webp to "{key}.webp"
   d. Thumbnail(150,150,"crop") -> upload to "thumbs/sm_{basename}"
   e. Thumbnail(400,400,"fit") -> upload to "thumbs/md_{basename}"
5. Upload original file
6. Create DB record with all URLs
7. Return response
```

**Delete flow:**
```
Delete(ctx, id, force) -> error
1. Get media by ID
2. If reference_count > 0 && !force -> return Conflict with referencing posts
3. If force -> (caller must be Admin+, checked at handler level)
4. Soft delete DB record
5. Audit log
```

**Service dependencies:** `MediaRepository`, `storage.Client`, `imaging.Processor`, `cache.Client`, `audit.Logger`

**Tests (mock all dependencies):**
- Upload success (image with thumbnails)
- Upload success (non-image, skips processing)
- Upload invalid MIME -> validation error
- Delete success (no references)
- Delete blocked (has references, no force)
- Delete force (has references, force=true)
- BatchDelete with partial skip
- List with pagination
- Update metadata

**Step 1-2: Write service and tests**

**Step 3: Run tests and commit**

```bash
go test ./internal/media/... -v
git add internal/media/service.go internal/media/service_test.go
git commit -m "feat(media): implement service with upload, image processing, and reference checks"
```

---

## Task 16: Media — handler + handler_test

**Files:**
- Modify: `internal/media/handler.go` (replace stub)
- Modify: `internal/media/handler_test.go` (replace stub)

**Upload handler** must parse multipart form:
```go
func (h *Handler) Upload(c *gin.Context) {
    file, header, err := c.Request.FormFile("file")
    // ... validate, pass to service
}
```

**Delete handler** checks `?force=true` query param.

**BatchDelete handler** accepts JSON body `{"ids": [...]}` + `?force=true`.

**Step 1-2: Write handler and tests**

**Step 3: Run tests and commit**

```bash
go test ./internal/media/... -v
git add internal/media/handler.go internal/media/handler_test.go
git commit -m "feat(media): implement handler with multipart upload and batch delete"
```

---

## Task 17: Router + API Registry

**Files:**
- Modify: `internal/router/router.go` — add DI + route registration for categories, tags, media
- Modify: `internal/router/api_meta.go` — add 18 new entries

**Step 1: Update router.go**

In `router.Setup()`, after the existing audit-logs route (line ~265), add:

```go
// Categories
catRepo := category.NewRepo(db)
catSvc := category.NewService(catRepo, cacheClient, auditSvc)
catHandler := category.NewHandler(catSvc)
siteScoped.GET("/categories", catHandler.List)
siteScoped.PUT("/categories/reorder", catHandler.Reorder)
siteScoped.GET("/categories/:id", catHandler.Get)
siteScoped.POST("/categories", catHandler.Create)
siteScoped.PUT("/categories/:id", catHandler.Update)
siteScoped.DELETE("/categories/:id", catHandler.Delete)

// Tags
tagRepo := tag.NewRepo(db)
tagSvc := tag.NewService(tagRepo, searchClient, cacheClient, auditSvc)
tagHandler := tag.NewHandler(tagSvc)
siteScoped.GET("/tags", tagHandler.List)
siteScoped.GET("/tags/suggest", tagHandler.Suggest)
siteScoped.GET("/tags/:id", tagHandler.Get)
siteScoped.POST("/tags", tagHandler.Create)
siteScoped.PUT("/tags/:id", tagHandler.Update)
siteScoped.DELETE("/tags/:id", tagHandler.Delete)

// Media
mediaRepo := media.NewRepo(db)
mediaSvc := media.NewService(mediaRepo, storageClient, imgProcessor, cacheClient, auditSvc)
mediaHandler := media.NewHandler(mediaSvc)
siteScoped.GET("/media", mediaHandler.List)
siteScoped.DELETE("/media/batch", mediaHandler.BatchDelete)
siteScoped.POST("/media", mediaHandler.Upload)
siteScoped.GET("/media/:id", mediaHandler.Get)
siteScoped.PUT("/media/:id", mediaHandler.Update)
siteScoped.DELETE("/media/:id", mediaHandler.Delete)
```

Also add necessary imports and DI setup for:
- `searchClient := search.NewClient(meili)` — wraps the existing `meili` parameter
- `storageClient := storage.NewClient(s3Client, cfg.RustFS.Bucket, cfg.RustFS.Endpoint+"/"+cfg.RustFS.Bucket)`
- `imgProcessor := imaging.NewProcessor()`
- `cacheClient := cache.NewClient(rdb)`

**Step 2: Update api_meta.go**

Add 18 entries to `BuildAPIMetaMap()`:

```go
// Site-scoped: Categories
"GET:/api/v1/site/categories":          {Name: "List categories", Description: "List category tree", Group: "categories"},
"PUT:/api/v1/site/categories/reorder":  {Name: "Reorder categories", Description: "Batch update sort order", Group: "categories"},
"GET:/api/v1/site/categories/:id":      {Name: "Get category", Description: "Get category details", Group: "categories"},
"POST:/api/v1/site/categories":         {Name: "Create category", Description: "Create a category", Group: "categories"},
"PUT:/api/v1/site/categories/:id":      {Name: "Update category", Description: "Update a category", Group: "categories"},
"DELETE:/api/v1/site/categories/:id":   {Name: "Delete category", Description: "Delete a leaf category", Group: "categories"},

// Site-scoped: Tags
"GET:/api/v1/site/tags":          {Name: "List tags", Description: "List tags with pagination", Group: "tags"},
"GET:/api/v1/site/tags/suggest":  {Name: "Suggest tags", Description: "Tag autocomplete via Meilisearch", Group: "tags"},
"GET:/api/v1/site/tags/:id":      {Name: "Get tag", Description: "Get tag details", Group: "tags"},
"POST:/api/v1/site/tags":         {Name: "Create tag", Description: "Create a tag", Group: "tags"},
"PUT:/api/v1/site/tags/:id":      {Name: "Update tag", Description: "Update a tag", Group: "tags"},
"DELETE:/api/v1/site/tags/:id":   {Name: "Delete tag", Description: "Delete a tag", Group: "tags"},

// Site-scoped: Media
"GET:/api/v1/site/media":          {Name: "List media", Description: "List media files", Group: "media"},
"DELETE:/api/v1/site/media/batch": {Name: "Batch delete media", Description: "Batch delete media files", Group: "media"},
"POST:/api/v1/site/media":         {Name: "Upload media", Description: "Upload a media file", Group: "media"},
"GET:/api/v1/site/media/:id":      {Name: "Get media", Description: "Get media file details", Group: "media"},
"PUT:/api/v1/site/media/:id":      {Name: "Update media", Description: "Update media metadata", Group: "media"},
"DELETE:/api/v1/site/media/:id":   {Name: "Delete media", Description: "Soft delete media file", Group: "media"},
```

**Step 3: Verify compilation**

```bash
go build ./cmd/cms/...
```

**Step 4: Commit**

```bash
git add internal/router/ internal/pkg/search/ internal/pkg/imaging/ internal/pkg/storage/ internal/pkg/cache/
git commit -m "feat(router): register categories + tags + media routes with DI and API Registry"
```

---

## Task 18: Add audit.NoopLogger for tests

**Files:**
- Modify: `internal/pkg/audit/service.go` — add `NoopLogger` if not already present

Check if `audit.NewNoopLogger()` exists. If not, add:
```go
type NoopLogger struct{}
func NewNoopLogger() *NoopLogger { return &NoopLogger{} }
func (n *NoopLogger) Log(_ context.Context, _ Entry) error { return nil }
```

This is needed by service tests.

**Step 1: Check and add if needed**

```bash
grep -r "NoopLogger" internal/pkg/audit/
```

**Step 2: Commit if changed**

```bash
git add internal/pkg/audit/
git commit -m "feat(audit): add NoopLogger for test use"
```

---

## Task 19: Full verification

**Step 1: Run all tests**

```bash
go test ./... -count=1 2>&1 | tail -40
```

Expected: All packages PASS (skip testcontainers tests if Docker not running)

**Step 2: Run go vet**

```bash
go vet ./...
```

Expected: No warnings

**Step 3: Verify route count**

```bash
go run ./cmd/cms serve --help
```

Verify the app compiles and starts (if Docker services available).

**Step 4: Final commit (if any fixups needed)**

```bash
git add -A && git commit -m "fix: address issues found during full verification"
```

---

## Summary

| Task | Description | New Files | Key Dependencies |
|------|-------------|-----------|-----------------|
| 1 | Add x/image dep | go.mod | — |
| 2 | pkg/search | 2 | meilisearch-go |
| 3 | pkg/imaging | 2 | x/image |
| 4 | pkg/storage | 2 | aws-sdk-go-v2 |
| 5 | pkg/cache | 2 | go-redis |
| 6 | Category interfaces+dto | 2 | model |
| 7 | Category repository | 1 | bun |
| 8 | Category service+tests | 2 | cache, audit |
| 9 | Category handler+tests | 2 | gin, response |
| 10 | Tag interfaces+dto+repo | 3 | bun, model |
| 11 | Tag service+tests | 2 | search, cache, audit |
| 12 | Tag handler+tests | 2 | gin |
| 13 | Media interfaces+dto | 2 | model |
| 14 | Media repository | 1 | bun |
| 15 | Media service+tests | 2 | storage, imaging, cache, audit |
| 16 | Media handler+tests | 2 | gin, multipart |
| 17 | Router + API Registry | 2 modified | all modules |
| 18 | Audit NoopLogger | 1 | — |
| 19 | Full verification | — | — |

**Total new routes: 18** (6 categories + 6 tags + 6 media)
**Total route count: 63 → 81**
**Total RBAC entries: 46 → 64**

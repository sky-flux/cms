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

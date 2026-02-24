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

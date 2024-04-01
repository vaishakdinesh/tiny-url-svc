package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/vaishakdinesh/tiny-url-svc/types"
)

const cacheExpire = time.Hour * 24 * 1 // 1 day

type cacheService struct {
	c *redis.Client
}

// NewCacheService return a new instance of a cache service
func NewCacheService(c *redis.Client) types.CacheService {
	return &cacheService{c: c}
}

// Cache stores a key:value pair in the cache
func (c *cacheService) Cache(ctx context.Context, key string, val any) error {
	statusCMD := c.c.Set(ctx, key, val, cacheExpire)
	return statusCMD.Err()
}

// Delete removes a key:value pair from the cache
func (c *cacheService) Delete(ctx context.Context, key string) error {
	deleted := c.c.Del(ctx, key)
	return deleted.Err()
}

// GetCachedValue returns the cached value for a key
func (c *cacheService) GetCachedValue(ctx context.Context, key string) (string, error) {
	cached := c.c.Get(ctx, key)
	if cached.Err() != nil {
		return "", cached.Err()
	}
	return cached.Val(), nil
}

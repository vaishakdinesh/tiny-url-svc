package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/vaishakdinesh/tiny-url-svc/types"
)

const cacheExpire = time.Hour * 24 * 365

type cacheService struct {
	c *redis.Client
}

func NewCacheService(c *redis.Client) types.CacheService {
	return &cacheService{c: c}
}

func (c *cacheService) Cache(ctx context.Context, key string, val any) error {
	statusCMD := c.c.Set(ctx, key, val, cacheExpire)
	return statusCMD.Err()
}

func (c *cacheService) Delete(ctx context.Context, key string) error {
	deleted := c.c.Del(ctx, key)
	return deleted.Err()
}

func (c *cacheService) GetCachedValue(ctx context.Context, key string) (string, error) {
	cached := c.c.Get(ctx, key)
	if cached.Err() != nil {
		return "", cached.Err()
	}
	return cached.Val(), nil
}

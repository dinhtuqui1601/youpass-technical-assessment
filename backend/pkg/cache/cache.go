package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is a thin Redis-backed cache used for short-lived lookups such as
// share-link code mappings. Cached values are never used as the final authority
// for authorization decisions.
type Cache struct {
	rdb *redis.Client
}

// New builds a Cache from a go-redis client.
func New(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

// Get returns the value for key, or an empty string when the key is missing.
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	v, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return v, nil
}

// Set stores value under key with the given TTL.
func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Del removes one or more keys.
func (c *Cache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

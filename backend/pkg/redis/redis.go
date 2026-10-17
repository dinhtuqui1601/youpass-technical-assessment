package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

// Client is an alias for the go-redis client used across the backend.
type Client = goredis.Client

// NewClient builds a Redis client from the given connection settings.
func NewClient(addr, password string, db int) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// Ping verifies connectivity with Redis.
func Ping(ctx context.Context, client *goredis.Client) error {
	return client.Ping(ctx).Err()
}

package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/youpass/backend/pkg/response"
)

// RateLimit implements a Redis fixed-window rate limiter keyed by client IP.
// It fails open when Redis is unavailable so the API remains usable.
func RateLimit(rdb *redis.Client, keyPrefix string, limit int, window time.Duration, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := keyPrefix + ":" + c.ClientIP()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			if logger != nil {
				logger.Warn("rate limit skipped due to redis error", "error", err)
			}
			c.Next()
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}
		if count > int64(limit) {
			response.Fail(c, 429, response.CodeRateLimited, "Too many requests, please try again later")
			return
		}
		c.Next()
	}
}

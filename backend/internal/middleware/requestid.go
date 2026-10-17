package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/youpass/backend/pkg/response"
)

// RequestID assigns a unique request id to every request and exposes it both in
// the X-Request-ID header and in the unified response envelope.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(string(response.RequestIDKey), rid)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}

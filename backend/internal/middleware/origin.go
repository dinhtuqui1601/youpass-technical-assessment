package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/youpass/backend/pkg/response"
)

// OriginCheck validates the Origin header against the configured allowlist for
// cookie-bearing endpoints. Requests without an Origin header (same-origin
// navigations, curl, ...) are allowed through; browser CORS is still enforced
// by the CORS middleware.
func OriginCheck(allowedOrigins string) gin.HandlerFunc {
	allowed := splitOrigins(allowedOrigins)
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && !isAllowed(allowed, origin) {
			response.Fail(c, 403, response.CodeForbidden, "Origin is not allowed")
			return
		}
		c.Next()
	}
}

func splitOrigins(allowedOrigins string) []string {
	parts := strings.Split(allowedOrigins, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/youpass/backend/internal/auth"
	"github.com/youpass/backend/pkg/response"
)

// Auth validates the Bearer access token and injects the user identity into the
// request context. It distinguishes expired tokens so the frontend knows to
// trigger a refresh.
func Auth(tm *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Fail(c, 401, response.CodeMissingAccessToken, "Missing access token")
			return
		}

		raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := tm.Parse(raw)
		if err != nil {
			if errors.Is(err, auth.ErrTokenExpired) {
				response.Fail(c, 401, response.CodeAccessTokenExpired, "Access token has expired")
			} else {
				response.Fail(c, 401, response.CodeInvalidAccessToken, "Invalid access token")
			}
			return
		}

		c.Set(string(response.UserIDKey), claims.Subject)
		c.Set(string(response.RoleKey), claims.Role)
		c.Next()
	}
}

package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ContextKey is the type used for gin context keys set by middleware.
type ContextKey string

const (
	// RequestIDKey stores the request id in the gin context.
	RequestIDKey ContextKey = "requestId"
	// UserIDKey stores the authenticated user id in the gin context.
	UserIDKey ContextKey = "userID"
	// RoleKey stores the authenticated user role in the gin context.
	RoleKey ContextKey = "role"
)

// APIError is a machine-readable error returned in the response envelope.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error makes *APIError implement the standard error interface so services can
// return it directly.
func (e *APIError) Error() string {
	return e.Code + ": " + e.Message
}

// Envelope is the unified response format for every JSON endpoint.
type Envelope struct {
	Success   bool      `json:"success"`
	Data      any       `json:"data"`
	Error     *APIError `json:"error"`
	RequestID string    `json:"requestId"`
}

// Error codes used across the API. The frontend relies on these to implement
// seamless authentication (only ACCESS_TOKEN_EXPIRED triggers a refresh).
const (
	CodeValidationError     = "VALIDATION_ERROR"
	CodeInternalError       = "INTERNAL_ERROR"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeEmailExists         = "EMAIL_ALREADY_EXISTS"
	CodeInvalidCredentials  = "INVALID_CREDENTIALS"
	CodeAccessTokenExpired  = "ACCESS_TOKEN_EXPIRED"
	CodeInvalidAccessToken  = "INVALID_ACCESS_TOKEN"
	CodeMissingAccessToken  = "MISSING_ACCESS_TOKEN"
	CodeRefreshTokenExpired = "REFRESH_TOKEN_EXPIRED"
	CodeInvalidRefreshToken = "INVALID_REFRESH_TOKEN"
	CodeMissingRefreshToken = "MISSING_REFRESH_TOKEN"
	CodeSessionRevoked      = "SESSION_REVOKED"
	CodeForbidden           = "FORBIDDEN"
	CodeRateLimited         = "RATE_LIMITED"
	CodeShareLinkNotFound   = "SHARE_LINK_NOT_FOUND"
	CodeShareLinkDisabled   = "SHARE_LINK_DISABLED"
	CodeShareLinkDeleted    = "SHARE_LINK_DELETED"
	CodeShareLinkExpired    = "SHARE_LINK_EXPIRED"
	CodeSubmissionNotFound  = "SUBMISSION_NOT_FOUND"
	CodeQueueUnavailable    = "QUEUE_UNAVAILABLE"
)

// NewError builds an *APIError with the given code and message.
func NewError(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

// StatusForCode maps an API error code to its HTTP status.
func StatusForCode(code string) int {
	switch code {
	case CodeAccessTokenExpired, CodeInvalidAccessToken, CodeMissingAccessToken,
		CodeRefreshTokenExpired, CodeInvalidRefreshToken, CodeMissingRefreshToken,
		CodeInvalidCredentials, CodeSessionRevoked:
		return http.StatusUnauthorized
	case CodeForbidden, CodeShareLinkDisabled, CodeShareLinkDeleted, CodeShareLinkExpired:
		return http.StatusForbidden
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeQueueUnavailable:
		return http.StatusServiceUnavailable
	case CodeNotFound, CodeShareLinkNotFound, CodeSubmissionNotFound:
		return http.StatusNotFound
	case CodeEmailExists, CodeConflict:
		return http.StatusConflict
	case CodeValidationError:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func requestID(c *gin.Context) string {
	if v, ok := c.Get(string(RequestIDKey)); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// JSON writes a successful response using the unified envelope.
func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, Envelope{Success: true, Data: data, RequestID: requestID(c)})
}

// Fail writes an error response using the unified envelope and aborts the chain.
func Fail(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, Envelope{
		Success:   false,
		Error:     NewError(code, message),
		RequestID: requestID(c),
	})
}

// FailErr writes an error response derived from an error. Recognizes *APIError
// to keep the machine-readable code; anything else becomes a 500.
func FailErr(c *gin.Context, err error) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		Fail(c, StatusForCode(apiErr.Code), apiErr.Code, apiErr.Message)
		return
	}
	Fail(c, http.StatusInternalServerError, CodeInternalError, "Internal server error")
}

// UserID returns the authenticated user id set by the Auth middleware.
func UserID(c *gin.Context) string {
	v, _ := c.Get(string(UserIDKey))
	s, _ := v.(string)
	return s
}

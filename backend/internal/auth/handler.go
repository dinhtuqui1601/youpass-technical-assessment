package auth

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/youpass/backend/pkg/response"
)

const refreshCookieName = "refresh_token"

// CookieConfig controls the refresh-token cookie attributes.
type CookieConfig struct {
	Secure   bool
	SameSite http.SameSite
	Domain   string
}

// Handler exposes the authentication HTTP endpoints.
type Handler struct {
	svc        *Service
	cookie     CookieConfig
	refreshTTL time.Duration
	logger     *slog.Logger
}

// NewHandler builds an auth Handler.
func NewHandler(svc *Service, cookie CookieConfig, refreshTTL time.Duration, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, cookie: cookie, refreshTTL: refreshTTL, logger: logger}
}

// Register godoc
// @Summary      Register a new account
// @Description  Creates a learner account (demo endpoint).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "Registration payload"
// @Success      201      {object}  response.Envelope{data=UserDTO}
// @Failure      400      {object}  response.Envelope
// @Failure      409      {object}  response.Envelope
// @Router       /api/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeValidationError, err.Error())
		return
	}

	u, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, toUserDTO(u))
}

// Login godoc
// @Summary      Login
// @Description  Verifies credentials and issues an access token. The refresh token is set as an HttpOnly cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Login payload"
// @Success      200      {object}  response.Envelope{data=LoginResponse}
// @Failure      400      {object}  response.Envelope
// @Failure      401      {object}  response.Envelope
// @Failure      429      {object}  response.Envelope
// @Router       /api/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeValidationError, err.Error())
		return
	}

	metadata := map[string]string{
		"ip":        c.ClientIP(),
		"userAgent": c.Request.UserAgent(),
	}
	resp, refreshToken, err := h.svc.Login(c.Request.Context(), req, metadata)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	h.setRefreshCookie(c, refreshToken)
	response.JSON(c, http.StatusOK, resp)
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Rotates the refresh token from the HttpOnly cookie and returns a new access token.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  response.Envelope{data=LoginResponse}
// @Failure      401  {object}  response.Envelope
// @Failure      429  {object}  response.Envelope
// @Router       /api/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	token, err := c.Cookie(refreshCookieName)
	if err != nil || token == "" {
		response.Fail(c, http.StatusUnauthorized, response.CodeMissingRefreshToken, "Missing refresh token")
		return
	}

	resp, newToken, err := h.svc.Refresh(c.Request.Context(), token)
	if err != nil {
		h.clearRefreshCookie(c)
		response.FailErr(c, err)
		return
	}

	h.setRefreshCookie(c, newToken)
	response.JSON(c, http.StatusOK, resp)
}

// Logout godoc
// @Summary      Logout
// @Description  Revokes the current session and clears the refresh-token cookie.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  response.Envelope
// @Router       /api/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	if token, err := c.Cookie(refreshCookieName); err == nil && token != "" {
		if err := h.svc.Logout(c.Request.Context(), token); err != nil && h.logger != nil {
			h.logger.Warn("logout failed", "error", err)
		}
	}
	h.clearRefreshCookie(c)
	response.JSON(c, http.StatusOK, gin.H{"loggedOut": true})
}

// Me godoc
// @Summary      Get current user
// @Description  Returns the profile of the authenticated user.
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.Envelope{data=UserDTO}
// @Failure      401  {object}  response.Envelope
// @Router       /api/auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(response.UserID(c))
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeInvalidAccessToken, "Invalid token subject")
		return
	}

	u, err := h.svc.GetUserByID(c.Request.Context(), id)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toUserDTO(u))
}

func (h *Handler) setRefreshCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/api/auth",
		Domain:   h.cookie.Domain,
		MaxAge:   int(h.refreshTTL.Seconds()),
		Expires:  time.Now().Add(h.refreshTTL),
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: h.cookie.SameSite,
	})
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/auth",
		Domain:   h.cookie.Domain,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: h.cookie.SameSite,
	})
}
